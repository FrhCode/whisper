//go:build windows

package overlay

import (
	"context"
	"sync"
	"syscall"
	"unsafe"
)

const (
	wmClose     = 0x0010
	wmDestroy   = 0x0002
	wsExTopmost = 0x00000008
	wsExToolwnd = 0x00000080
	wsExLayered = 0x00080000
	wsPopup     = 0x80000000
	swHide      = 0
	swShow      = 5
	swpNoSize   = 0x0001
	swpNoZOrder = 0x0004
	colorWindow = 5
	waInactive  = 0
	margin      = 32
	width       = 260
	height      = 56
)

type Overlay struct{ ch chan string }

var (
	user32              = syscall.NewLazyDLL("user32.dll")
	kernel32            = syscall.NewLazyDLL("kernel32.dll")
	getModuleHandle     = kernel32.NewProc("GetModuleHandleW")
	registerClassEx     = user32.NewProc("RegisterClassExW")
	createWindowEx      = user32.NewProc("CreateWindowExW")
	defWindowProc       = user32.NewProc("DefWindowProcW")
	destroyWindow       = user32.NewProc("DestroyWindow")
	showWindow          = user32.NewProc("ShowWindow")
	setWindowText       = user32.NewProc("SetWindowTextW")
	setWindowPos        = user32.NewProc("SetWindowPos")
	getSystemMetrics    = user32.NewProc("GetSystemMetrics")
	getMessage          = user32.NewProc("GetMessageW")
	translateMessage    = user32.NewProc("TranslateMessage")
	dispatchMessage     = user32.NewProc("DispatchMessageW")
	postQuitMessage     = user32.NewProc("PostQuitMessage")
	setLayeredWindowAtt = user32.NewProc("SetLayeredWindowAttributes")
	once                sync.Once
	className           = syscall.StringToUTF16Ptr("WhisprOverlay")
)

type wndClassEx struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   uintptr
	Icon       uintptr
	Cursor     uintptr
	Background uintptr
	MenuName   *uint16
	ClassName  *uint16
	IconSm     uintptr
}

type msg struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
}

func New(ctx context.Context) *Overlay {
	o := &Overlay{ch: make(chan string, 8)}
	go o.run(ctx)
	return o
}

func (o *Overlay) Set(s string) {
	select {
	case o.ch <- s:
	default:
	}
}
func (o *Overlay) Hide() { o.Set("") }

func (o *Overlay) run(ctx context.Context) {
	hinst, _, _ := getModuleHandle.Call(0)
	once.Do(func() {
		wc := wndClassEx{Size: uint32(unsafe.Sizeof(wndClassEx{})), WndProc: syscall.NewCallback(wndProc), Instance: hinst, Background: colorWindow + 1, ClassName: className}
		registerClassEx.Call(uintptr(unsafe.Pointer(&wc)))
	})
	hwnd, _, _ := createWindowEx.Call(wsExTopmost|wsExToolwnd|wsExLayered, uintptr(unsafe.Pointer(className)), uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("Whispr"))), wsPopup, 0, 0, width, height, 0, 0, hinst, 0)
	setLayeredWindowAtt.Call(hwnd, 0, 235, 0x2)
	position(hwnd)

	go func() { <-ctx.Done(); destroyWindow.Call(hwnd) }()
	go func() {
		for s := range o.ch {
			if s == "" {
				showWindow.Call(hwnd, swHide)
				continue
			}
			setWindowText.Call(hwnd, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(s))))
			position(hwnd)
			showWindow.Call(hwnd, swShow)
		}
	}()

	var m msg
	for {
		r, _, _ := getMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if r == 0 || r == ^uintptr(0) {
			return
		}
		translateMessage.Call(uintptr(unsafe.Pointer(&m)))
		dispatchMessage.Call(uintptr(unsafe.Pointer(&m)))
	}
}

func position(hwnd uintptr) {
	sw, _, _ := getSystemMetrics.Call(0)
	sh, _, _ := getSystemMetrics.Call(1)
	setWindowPos.Call(hwnd, 0, sw-width-margin, sh-height-margin, 0, 0, swpNoSize|swpNoZOrder)
}

func wndProc(hwnd uintptr, msg uint32, wparam, lparam uintptr) uintptr {
	switch msg {
	case wmClose:
		showWindow.Call(hwnd, swHide)
		return 0
	case wmDestroy:
		postQuitMessage.Call(0)
		return 0
	}
	r, _, _ := defWindowProc.Call(hwnd, uintptr(msg), wparam, lparam)
	return r
}

var _ = waInactive
