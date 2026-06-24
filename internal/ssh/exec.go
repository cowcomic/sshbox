package ssh

import (
	"fmt"

	"golang.org/x/crypto/ssh"
)

// Exec runs a command on the remote host and returns its combined output.
func Exec(host string, port int, user, password, command string) ([]byte, error) {
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
		return nil, fmt.Errorf("无法连接到 %s: %w", addr, err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("创建会话失败: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		// Still return output even on error (e.g. non-zero exit code)
		return output, fmt.Errorf("命令执行失败: %w", err)
	}
	return output, nil
}
