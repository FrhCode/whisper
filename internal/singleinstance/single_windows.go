//go:build windows

package singleinstance

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	kernel32     = syscall.NewLazyDLL("kernel32.dll")
	createMutexW = kernel32.NewProc("CreateMutexW")
	closeHandle  = kernel32.NewProc("CloseHandle")
	getLastError = kernel32.NewProc("GetLastError")
	mutexHandle  uintptr
)

const errorAlreadyExists = 183

func Lock(name string) (func(), error) {
	p, err := syscall.UTF16PtrFromString("Global\\" + name)
	if err != nil {
		return nil, err
	}
	h, _, callErr := createMutexW.Call(0, 1, uintptr(unsafe.Pointer(p)))
	if h == 0 {
		return nil, fmt.Errorf("create mutex failed: %w", callErr)
	}
	last, _, _ := getLastError.Call()
	if last == errorAlreadyExists {
		closeHandle.Call(h)
		return nil, fmt.Errorf("Whispr already running")
	}
	mutexHandle = h
	return func() { closeHandle.Call(mutexHandle) }, nil
}
