package ssh

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// Connect establishes an interactive SSH session to the given host.
func Connect(name string, host string, port int, user, password string) error {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("无法连接到 %s: %w", addr, err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("创建会话失败: %w", err)
	}
	defer session.Close()

	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return fmt.Errorf("标准输入不是终端")
	}

	// 在进入 raw 模式前获取终端尺寸（避免 raw 模式影响控制台 API）
	width, height, ok := getTerminalSize(fd)
	if !ok {
		width, height = 80, 40
	}

	// 保存原始终端状态并进入 raw 模式
	origState, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("设置终端raw模式失败: %w", err)
	}

	// 统一清理：确保终端状态恢复
	done := make(chan struct{})
	cleanup := func() {
		stopResizeWatcher()
		close(done)
		time.Sleep(50 * time.Millisecond) // 等待标题 goroutine 退出
		term.Restore(fd, origState)
		fmt.Print("\033[?1049l") // 切回主屏幕缓冲区
		fmt.Print("\033]0;\007") // 恢复终端标题
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1, // 由远程 PTY 负责回显
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("xterm-256color", height, width, modes); err != nil {
		cleanup()
		return fmt.Errorf("请求PTY失败: %w", err)
	}

	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	if err := session.Shell(); err != nil {
		cleanup()
		return fmt.Errorf("启动shell失败: %w", err)
	}

	// 定期设置终端标题（远程 shell 的 PS1 会不断覆盖）
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				fmt.Printf("\033]0;%s\007", name)
			case <-done:
				return
			}
		}
	}()

	// 处理窗口大小变化（跨平台）
	startResizeWatcher(fd, session)

	err = session.Wait()
	cleanup()
	return err
}
