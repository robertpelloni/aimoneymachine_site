//go:build windows

package main

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows/registry"
)

const runKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`
const runValueName = "AIMM-Tray"

// autostartSupported reports whether start-at-login can be toggled here.
func autostartSupported() bool { return true }

// autostartEnabled reports whether the Run key already points at us.
func autostartEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	v, _, err := k.GetStringValue(runValueName)
	if err != nil {
		return false
	}
	exe, _ := os.Executable()
	return v == fmt.Sprintf("%q", exe)
}

// setAutostart adds/removes the HKCU Run value for the current executable.
func setAutostart(enabled bool) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	if !enabled {
		_ = k.DeleteValue(runValueName)
		return nil
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	return k.SetStringValue(runValueName, fmt.Sprintf("%q", exe))
}
