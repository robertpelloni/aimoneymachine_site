//go:build !windows

package main

import "os/exec"

func configureHiddenProcess(cmd *exec.Cmd, belowNormal bool) {
	// No-op on Unix platforms (no console window popup).
}
