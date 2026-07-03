//go:build windows

package hotkey

import (
	"context"
	"fmt"
	"syscall"
	"unsafe"
)

const (
	modAlt     = 0x0001
	modControl = 0x0002
	vkSpace    = 0x20
	wmHotkey   = 0x0312
)

var user32 = syscall.NewLazyDLL("user32.dll")
var registerHotKey = user32.NewProc("RegisterHotKey")
var unregisterHotKey = user32.NewProc("UnregisterHotKey")
var getMessage = user32.NewProc("GetMessageW")

type msg struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
}

func Listen(ctx context.Context, fn func()) error {
	const id = 1
	r, _, err := registerHotKey.Call(0, id, modControl|modAlt, vkSpace)
	if r == 0 {
		return fmt.Errorf("Ctrl+Alt+Space hotkey unavailable: %w", err)
	}
	defer unregisterHotKey.Call(0, id)

	go func() { <-ctx.Done(); unregisterHotKey.Call(0, id) }()

	var m msg
	for {
		r, _, err := getMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if r == 0 || r == ^uintptr(0) {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		if m.Message == wmHotkey && m.WParam == id {
			go fn()
		}
	}
}
