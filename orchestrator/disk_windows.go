//go:build windows

package orchestrator

import "golang.org/x/sys/windows"

// DiskUsage returns used-percent and free GiB for the volume holding path.
func DiskUsage(path string) (usedPct float64, freeGB float64, ok bool) {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, 0, false
	}
	var freeAvail, total, totalFree uint64
	if err := windows.GetDiskFreeSpaceEx(p, &freeAvail, &total, &totalFree); err != nil {
		return 0, 0, false
	}
	if total == 0 {
		return 0, 0, false
	}
	return (1 - float64(freeAvail)/float64(total)) * 100, float64(freeAvail) / (1 << 30), true
}
