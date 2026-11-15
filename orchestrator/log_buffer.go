package orchestrator

import (
	"bytes"
	"io"
	"os"
	"sync"
)

// RingLogBuffer keeps the most recent maxLines lines written to process stdout/stderr
// so GET /api/log serves live orchestrator output even when systemd routes stdout to journald
// instead of a flat orchestrator_live.log file.
type RingLogBuffer struct {
	mu       sync.RWMutex
	lines    []string
	partial  string
	maxLines int
}

var GlobalLogBuffer = NewRingLogBuffer(600)

func NewRingLogBuffer(maxLines int) *RingLogBuffer {
	if maxLines <= 0 {
		maxLines = 500
	}
	return &RingLogBuffer{
		lines:    make([]string, 0, maxLines),
		maxLines: maxLines,
	}
}

func (b *RingLogBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	chunk := b.partial + string(p)
	parts := bytes.Split([]byte(chunk), []byte("\n"))
	b.partial = string(parts[len(parts)-1])

	for i := 0; i < len(parts)-1; i++ {
		line := string(bytes.TrimRight(parts[i], "\r"))
		b.lines = append(b.lines, line)
	}
	if len(b.lines) > b.maxLines {
		b.lines = append([]string(nil), b.lines[len(b.lines)-b.maxLines:]...)
	}
	return len(p), nil
}

func (b *RingLogBuffer) String() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if len(b.lines) == 0 && b.partial == "" {
		return ""
	}
	var buf bytes.Buffer
	for _, l := range b.lines {
		buf.WriteString(l)
		buf.WriteByte('\n')
	}
	if b.partial != "" {
		buf.WriteString(b.partial)
	}
	return buf.String()
}

// InstallStdoutTee tees os.Stdout into GlobalLogBuffer in-process.
func InstallStdoutTee() {
	origOut := os.Stdout
	rOut, wOut, err := os.Pipe()
	if err != nil {
		return
	}
	os.Stdout = wOut
	mw := io.MultiWriter(origOut, GlobalLogBuffer)
	go func() {
		_, _ = io.Copy(mw, rOut)
	}()
}
