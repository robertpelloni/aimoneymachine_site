//go:build windows

package main

import (
	"os/exec"
	"syscall"
)

const belowNormalPriorityClass = 0x00004000

// configureHiddenProcess hides transient console windows on Windows and optionally
// lowers child CPU scheduling priority to BelowNormal so background daemons/tunnels
// never contend with interactive desktop work.
func configureHiddenProcess(cmd *exec.Cmd, belowNormal bool) {
	attr := &syscall.SysProcAttr{HideWindow: true}
	if belowNormal {
		attr.CreationFlags = belowNormalPriorityClass
	}
	cmd.SysProcAttr = attr
}
