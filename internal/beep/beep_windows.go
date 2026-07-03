//go:build windows

package beep

import "syscall"

var messageBeep = syscall.NewLazyDLL("user32.dll").NewProc("MessageBeep")

func Start() { messageBeep.Call(0xFFFFFFFF) }
func Stop()  { messageBeep.Call(0xFFFFFFFF) }
func Error() { messageBeep.Call(0x00000010) }
