//go:build !linux && !windows

package orchestrator

// DiskUsage is unimplemented on this platform; callers omit the fields.
func DiskUsage(string) (float64, float64, bool) { return 0, 0, false }
