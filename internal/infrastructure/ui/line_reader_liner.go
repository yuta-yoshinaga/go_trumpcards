//go:build !js || !wasm

package ui

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/peterh/liner"
)

// historyFileName is the basename of the readline history file. Persisted in
// the user's home dir so arrow-key recall survives across CLI sessions.
const historyFileName = ".trumpcards_history"

// linerLineReader wraps a *liner.State so the multi-game CLI gets arrow-key
// history, Tab completion, and in-line editing. peterh/liner is the lightest
// readline-style library on the Go ecosystem (~700 LoC, no cgo) — picked over
// chzyer/readline (low maintenance) and c-bata/go-prompt (whole-screen TUI
// that fights our existing fmt.Println output).
type linerLineReader struct {
	state       *liner.State
	historyPath string // empty when persistence is disabled (no HOME)
}

// newLinerReader returns a liner-backed LineReader. It loads the persisted
// history file (best-effort — missing or unreadable files are silently
// ignored, matching `psql`/`redis-cli` behaviour) and configures Ctrl+C to
// abort the prompt instead of terminating the process so the SIGINT handler
// in cui_runner.go can print the goodbye banner.
func newLinerReader() *linerLineReader {
	state := liner.NewLiner()
	state.SetCtrlCAborts(true)

	r := &linerLineReader{state: state}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		r.historyPath = filepath.Join(home, historyFileName)
		if f, err := os.Open(r.historyPath); err == nil {
			_, _ = state.ReadHistory(f)
			_ = f.Close()
		}
	}
	return r
}

// Prompt reads one line. liner returns liner.ErrPromptAborted on Ctrl+C;
// translate it to io.EOF so the caller treats it the same as Ctrl+D and
// exits cleanly. Other errors are returned as-is.
func (l *linerLineReader) Prompt(prompt string) (string, error) {
	line, err := l.state.Prompt(prompt)
	if err != nil {
		if errors.Is(err, liner.ErrPromptAborted) {
			return "", io.EOF
		}
		return "", err
	}
	return line, nil
}

// AppendHistory records non-empty lines for arrow-key recall. liner
// deduplicates consecutive identical entries automatically.
func (l *linerLineReader) AppendHistory(line string) {
	if line == "" {
		return
	}
	l.state.AppendHistory(line)
}

// SetCompleter installs a tab-completion function. liner calls fn with the
// full line typed so far; we forward verbatim and let the caller decide how
// to interpret it (typically: split on whitespace, complete the first token).
func (l *linerLineReader) SetCompleter(fn func(prefix string) []string) {
	l.state.SetCompleter(fn)
}

// Close persists the history file (best-effort — failures here are silent
// since we're already shutting down) and tears down the terminal raw mode.
func (l *linerLineReader) Close() error {
	if l.historyPath != "" {
		if f, err := os.Create(l.historyPath); err == nil {
			_, _ = l.state.WriteHistory(f)
			_ = f.Close()
		}
	}
	return l.state.Close()
}

// newDefaultLineReader returns a liner-backed reader for native builds.
// cui_runner.go calls this once per CLI session.
func newDefaultLineReader() LineReader {
	return newLinerReader()
}
