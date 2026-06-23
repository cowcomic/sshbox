//go:build !windows

package ssh

import (
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

var resizeCh chan os.Signal

func startResizeWatcher(fd int, session *ssh.Session) {
	resizeCh = make(chan os.Signal, 1)
	signal.Notify(resizeCh, syscall.SIGWINCH)
	go func() {
		for range resizeCh {
			if w, h, err := term.GetSize(fd); err == nil {
				session.WindowChange(h, w)
			}
		}
	}()
}

func stopResizeWatcher() {
	if resizeCh != nil {
		signal.Stop(resizeCh)
	}
}

// getTerminalSize returns terminal dimensions. On Unix, term.GetSize works correctly.
func getTerminalSize(fd int) (int, int, bool) {
	w, h, err := term.GetSize(fd)
	if err != nil {
		return 0, 0, false
	}
	return w, h, true
}
