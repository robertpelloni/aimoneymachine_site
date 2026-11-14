package main

import (
	"errors"
	"testing"
)

// A second acquire while the first is held must be refused; after release
// it must succeed again.
func TestSingleInstanceGuard(t *testing.T) {
	release, err := acquireInstance()
	if err != nil {
		t.Fatalf("first acquire: %v", err)
	}
	if _, err := acquireInstance(); !errors.Is(err, ErrAlreadyRunning) {
		t.Fatalf("second acquire err = %v, want ErrAlreadyRunning", err)
	}
	release()
	release2, err := acquireInstance()
	if err != nil {
		t.Fatalf("acquire after release: %v", err)
	}
	release2()
}
