package ui

import (
	"bytes"
	"os"
	"strings"
	"syscall"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// closeSpyReader is a LineReader whose Close records that it ran, so tests can
// assert the signal path performs the history-save / terminal-restore cleanup
// that os.Exit would otherwise skip (issue #2096).
type closeSpyReader struct {
	closed int
}

func (*closeSpyReader) Prompt(string) (string, error)      { return "", nil }
func (*closeSpyReader) AppendHistory(string)               {}
func (*closeSpyReader) SetCompleter(func(string) []string) {}
func (r *closeSpyReader) Close() error                     { r.closed++; return nil }

func TestSignalExitCode_RunsCleanupAndMapsCode(t *testing.T) {
	tests := []struct {
		name string
		sig  syscall.Signal
		want int
	}{
		{name: "SIGINT maps to 130", sig: syscall.SIGINT, want: 130},
		{name: "SIGTERM maps to 143", sig: syscall.SIGTERM, want: 143},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &closeSpyReader{}
			got := signalExitCode(tt.sig, func() { _ = reader.Close() })
			if got != tt.want {
				t.Errorf("signalExitCode(%v) exit code = %d, want %d", tt.sig, got, tt.want)
			}
			// The cleanup (history save + terminal restore via reader.Close)
			// must run on the signal path — this is the regression the issue
			// is about: os.Exit alone would skip it.
			if reader.closed != 1 {
				t.Errorf("cleanup ran %d times, want exactly 1", reader.closed)
			}
		})
	}
}

func TestSignalExitCode_NilCleanupIsSafe(t *testing.T) {
	// The Doubt loop passes a nil cleanup (no reader/raw mode); it must not panic.
	if got := signalExitCode(syscall.SIGINT, nil); got != 130 {
		t.Errorf("signalExitCode with nil cleanup = %d, want 130", got)
	}
}

func TestRunSignalWatcher_NormalExitDoesNotRunCleanup(t *testing.T) {
	reader := &closeSpyReader{}
	stop := runSignalWatcher(func() { _ = reader.Close() })
	// Simulate a normal loop return: stop the watcher without a signal.
	// stop() now blocks until the goroutine exits, so no sleep needed.
	stop()
	if reader.closed != 0 {
		t.Errorf("cleanup ran on normal exit (%d times); it should only run on a signal", reader.closed)
	}
}

// captureStderr runs fn with os.Stderr redirected and returns what it wrote.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	orig := os.Stderr
	os.Stderr = w
	defer func() { os.Stderr = orig }()

	fn()
	_ = w.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("read: %v", err)
	}
	return buf.String()
}

// printResult must actually apply the stderr colour. The colour package's own
// tests prove RedStderr colours; only this proves printResult calls it, which
// is the wiring that was missing entirely (#5194).
func TestPrintResult_ColoursErrorsOnStderr(t *testing.T) {
	origStderr := color.NoColorStderr()
	t.Cleanup(func() { color.SetStderrColor(!origStderr) })

	color.SetStderrColor(true)
	got := captureStderr(t, func() { printResult(i18n.MarkError("bad command")) })

	if !strings.Contains(got, "\033[31m") {
		t.Errorf("stderr output should carry the red escape code, got %q", got)
	}
	if !strings.Contains(got, "bad command") {
		t.Errorf("stderr output should contain the message, got %q", got)
	}
}

// With colour disabled the same path must emit no escape codes at all -- the
// other half of the gate, so neither "always plain" nor "always coloured" can
// pass unnoticed.
func TestPrintResult_PlainErrorsWhenColourDisabled(t *testing.T) {
	origStderr := color.NoColorStderr()
	t.Cleanup(func() { color.SetStderrColor(!origStderr) })

	color.SetStderrColor(false)
	got := captureStderr(t, func() { printResult(i18n.MarkError("bad command")) })

	if strings.Contains(got, "\033[") {
		t.Errorf("stderr output should carry no escape codes, got %q", got)
	}
	if !strings.Contains(got, "bad command") {
		t.Errorf("stderr output should still contain the message, got %q", got)
	}
}
