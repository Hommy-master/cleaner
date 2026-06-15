//go:build windows

package util

import (
	"os"
	"syscall"
	"unsafe"
)

const (
	enableExtendedFlags = 0x0080
	enableQuickEditMode = 0x0040
)

// disableQuickEdit turns off console QuickEdit mode so selecting text in cmd.exe
// does not freeze the process while it writes to stdout.
func disableQuickEdit() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getConsoleMode := kernel32.NewProc("GetConsoleMode")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")

	handle := syscall.Handle(os.Stdin.Fd())
	var mode uint32
	r, _, _ := getConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode)))
	if r == 0 {
		return
	}

	mode &^= enableQuickEditMode
	mode |= enableExtendedFlags
	setConsoleMode.Call(uintptr(handle), uintptr(mode))
}
