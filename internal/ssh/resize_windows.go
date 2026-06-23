//go:build windows

package ssh

import (
	"syscall"
	"unsafe"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

var (
	kernel32                         = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleScreenBufferInfo   = kernel32.NewProc("GetConsoleScreenBufferInfo")
	procCreateFileA                  = kernel32.NewProc("CreateFileA")
)

// Windows doesn't support SIGWINCH; terminal resize is handled by the console host.
func startResizeWatcher(fd int, session *ssh.Session) {}

func stopResizeWatcher() {}

type coord struct{ X, Y int16 }
type smallRect struct{ Left, Top, Right, Bottom int16 }
type consoleScreenBufferInfo struct {
	Size              coord
	CursorPosition    coord
	Attributes        uint16
	Window            smallRect
	MaximumWindowSize coord
}

// getTerminalSize returns the visible console window size using Windows API.
// os.Stdin.Fd() 的句柄对 GetConsoleScreenBufferInfo 无效，需要用 CONOUT$ 打开控制台输出句柄。
func getTerminalSize(fd int) (int, int, bool) {
	// 打开 CONOUT$ 获取控制台输出句柄
	conout, _, _ := procCreateFileA.Call(
		uintptr(unsafe.Pointer(syscall.StringBytePtr("CONOUT$"))),
		uintptr(syscall.GENERIC_READ),
		0,
		0,
		uintptr(syscall.OPEN_EXISTING),
		0,
		0,
	)
	if conout == 0 || conout == uintptr(syscall.InvalidHandle) {
		// 回退到 term.GetSize
		w, h, err := term.GetSize(fd)
		if err != nil {
			return 0, 0, false
		}
		return w, h, true
	}
	defer syscall.CloseHandle(syscall.Handle(conout))

	var info consoleScreenBufferInfo
	ret, _, _ := procGetConsoleScreenBufferInfo.Call(conout, uintptr(unsafe.Pointer(&info)))
	if ret == 0 {
		w, h, err := term.GetSize(fd)
		if err != nil {
			return 0, 0, false
		}
		return w, h, true
	}

	width := int(info.Window.Right-info.Window.Left) + 1
	height := int(info.Window.Bottom-info.Window.Top) + 1
	return width, height, true
}
