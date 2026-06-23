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

	// Save original terminal state and enter raw mode
	origState, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("设置终端raw模式失败: %w", err)
	}
	defer term.Restore(fd, origState)

	width, height, err := term.GetSize(fd)
	if err != nil {
		width, height = 80, 40
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("xterm-256color", height, width, modes); err != nil {
		return fmt.Errorf("请求PTY失败: %w", err)
	}

	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	if err := session.Shell(); err != nil {
		return fmt.Errorf("启动shell失败: %w", err)
	}

	// Periodically set terminal title to connection name
	// (remote shell's PS1 keeps overwriting it)
	done := make(chan struct{})
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

	// Handle window resize (platform-specific)
	startResizeWatcher(fd, session)

	err = session.Wait()
	close(done)
	stopResizeWatcher()

	// Restore original title
	fmt.Print("\033]0;\007")
	return err
}
