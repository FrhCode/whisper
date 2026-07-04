//go:build windows

package singleinstance

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

var kernel32 = syscall.NewLazyDLL("kernel32.dll")
var createFileW = kernel32.NewProc("CreateFileW")
var closeHandle = kernel32.NewProc("CloseHandle")

const (
	genericRead           = 0x80000000
	genericWrite          = 0x40000000
	openAlways            = 4
	fileAttributeNormal   = 0x80
	errorSharingViolation = 32
	errorLockViolation    = 33
)

var lockHandle uintptr

func Lock(name string) (func(), error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	dir = filepath.Join(dir, name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "instance.lock")
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	h, _, callErr := createFileW.Call(
		uintptr(unsafe.Pointer(p)),
		genericRead|genericWrite,
		0,
		0,
		openAlways,
		fileAttributeNormal,
		0,
	)
	if h == uintptr(syscall.InvalidHandle) {
		if errno, ok := callErr.(syscall.Errno); ok && (errno == errorSharingViolation || errno == errorLockViolation) {
			return nil, fmt.Errorf("Whispr already running")
		}
		return nil, fmt.Errorf("single instance lock failed: %w", callErr)
	}
	lockHandle = h
	return func() { closeHandle.Call(lockHandle) }, nil
}
