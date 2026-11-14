//go:build linux

package orchestrator

import "syscall"

// DiskUsage returns used-percent and free GiB for the filesystem holding
// path. Added after the 2026-11-14 full-disk outage: /status (and thus the
// dashboard and tray) should show the disk filling up, not just the
// watchdog log after it is already fatal.
func DiskUsage(path string) (usedPct float64, freeGB float64, ok bool) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return 0, 0, false
	}
	total := float64(st.Blocks) * float64(st.Bsize)
	free := float64(st.Bavail) * float64(st.Bsize)
	if total == 0 {
		return 0, 0, false
	}
	return (1 - free/total) * 100, free / (1 << 30), true
}
