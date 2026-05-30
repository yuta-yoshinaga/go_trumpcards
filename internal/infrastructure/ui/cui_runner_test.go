package ui

import (
	"syscall"
	"testing"
	"time"
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
	stop()
	// Give the watcher goroutine a moment to observe the done channel.
	time.Sleep(20 * time.Millisecond)
	if reader.closed != 0 {
		t.Errorf("cleanup ran on normal exit (%d times); it should only run on a signal", reader.closed)
	}
}
