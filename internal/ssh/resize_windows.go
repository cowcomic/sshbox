package ssh

import (
	"golang.org/x/crypto/ssh"
)

// Windows doesn't support SIGWINCH; terminal resize is handled by the console host.
func startResizeWatcher(fd int, session *ssh.Session) {}

func stopResizeWatcher() {}
