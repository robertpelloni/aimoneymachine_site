package orchestrator

import (
	"os"
	"path/filepath"
)

// WriteFileAtomic writes data to path via a same-directory temp file plus
// rename, so a kill or crash mid-write can never leave a truncated state
// file behind. The daemon is restarted on every deploy, and its JSON state
// (tasks.json, chains.json, ledger.json, memory.json, STATUS.json) is
// rewritten constantly — a plain os.WriteFile truncates on open, so one
// unlucky SIGKILL corrupted the file the next boot had to parse.
// Rename within one directory is atomic on POSIX and Windows.
func WriteFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".aimm-tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	// Best-effort cleanup; after a successful rename this is a no-op.
	defer os.Remove(tmpName)

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}
