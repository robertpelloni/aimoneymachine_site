//go:build windows

package main

import (
	"errors"

	"golang.org/x/sys/windows"
)

// ErrAlreadyRunning is returned when another aimm-tray holds the mutex.
var ErrAlreadyRunning = errors.New("aimm-tray is already running")

// acquireInstance takes the single-instance mutex. The returned release
// func frees it (also freed automatically when the process exits).
func acquireInstance() (func(), error) {
	name, err := windows.UTF16PtrFromString("AIMM-Tray-SingleInstance")
	if err != nil {
		return nil, err
	}
	h, err := windows.CreateMutex(nil, false, name)
	if err != nil {
		if errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
			if h != 0 {
				_ = windows.CloseHandle(h)
			}
			return nil, ErrAlreadyRunning
		}
		return nil, err
	}
	// CreateMutex reports ERROR_ALREADY_EXISTS via GetLastError even on
	// success-with-existing; the syscall wrapper surfaces it as err above,
	// so a nil err here means we created it.
	return func() { _ = windows.CloseHandle(h) }, nil
}
