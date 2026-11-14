//go:build !windows

package main

import "errors"

// Start-at-login toggling is implemented for Windows (HKCU Run key).
// Other platforms report unsupported and the tray disables the item.
func autostartSupported() bool          { return false }
func autostartEnabled() bool            { return false }
func setAutostart(bool) error           { return errors.New("start-at-login not supported on this platform") }
