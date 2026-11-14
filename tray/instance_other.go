//go:build !windows

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// ErrAlreadyRunning is returned when another aimm-tray holds the lockfile.
var ErrAlreadyRunning = errors.New("aimm-tray is already running")

func lockPath() string { return filepath.Join(os.TempDir(), "aimm-tray.lock") }

// acquireInstance takes an O_EXCL lockfile, taking over a stale one whose
// recorded PID is no longer alive.
func acquireInstance() (func(), error) {
	p := lockPath()
	for attempt := 0; attempt < 2; attempt++ {
		f, err := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err == nil {
			fmt.Fprintf(f, "%d", os.Getpid())
			_ = f.Close()
			return func() { _ = os.Remove(p) }, nil
		}
		if !os.IsExist(err) {
			return nil, err
		}
		// Lock exists: alive owner -> refuse; stale -> remove and retry.
		data, _ := os.ReadFile(p)
		pid, _ := strconv.Atoi(strings.TrimSpace(string(data)))
		if pid > 0 {
			if proc, ferr := os.FindProcess(pid); ferr == nil {
				if proc.Signal(syscall.Signal(0)) == nil {
					return nil, ErrAlreadyRunning
				}
			}
		}
		_ = os.Remove(p)
	}
	return nil, ErrAlreadyRunning
}
