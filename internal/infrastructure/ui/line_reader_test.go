//go:build test

package ui

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScannerLineReader_PromptWritesAndReads(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	r := newScannerLineReader(strings.NewReader("hello\n"), &out)
	line, err := r.Prompt("> ")
	assert.NoError(t, err)
	assert.Equal(t, "hello", line)
	assert.Equal(t, "> ", out.String())
}

func TestScannerLineReader_EOFReturnsErrEOF(t *testing.T) {
	t.Parallel()
	r := newScannerLineReader(strings.NewReader(""), &bytes.Buffer{})
	_, err := r.Prompt("> ")
	assert.True(t, errors.Is(err, io.EOF), "want io.EOF, got %v", err)
}

func TestScannerLineReader_NoOpsAreSafe(t *testing.T) {
	t.Parallel()
	r := newScannerLineReader(strings.NewReader(""), &bytes.Buffer{})
	// All these are documented no-ops; the test just exercises them so
	// future regressions to "panic on no-op" are caught.
	r.AppendHistory("anything")
	r.SetCompleter(func(string) []string { return nil })
	assert.NoError(t, r.Close())
}

func TestInstallCompleter_FirstToken(t *testing.T) {
	t.Parallel()

	var captured func(string) []string
	r := &captureCompleterReader{
		LineReader: newScannerLineReader(&bytes.Buffer{}, &bytes.Buffer{}),
		set:        func(fn func(string) []string) { captured = fn },
	}

	lister := &fakeLister{
		topLevel: []string{"switch", "games"},
		args:     map[string][]string{"switch": {"blackjack", "baccarat"}},
	}
	installCompleter(r, lister)

	cases := []struct {
		name   string
		prefix string
		want   []string
	}{
		{
			name:   "empty prefix returns common commands plus lister top-level",
			prefix: "",
			want:   []string{"q", "quit", "exit", "r", "reset", "help", "?", "switch", "games"},
		},
		{
			name:   "first-token prefix matches common command",
			prefix: "q",
			want:   []string{"q", "quit"},
		},
		{
			name:   "first-token prefix matches lister command",
			prefix: "swi",
			want:   []string{"switch"},
		},
		{
			name:   "bare game name is NOT a first-token candidate",
			prefix: "bla",
			want:   nil,
		},
		{
			name:   "non-matching prefix returns nothing",
			prefix: "zzz",
			want:   nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := captured(tc.prefix)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestInstallCompleter_SecondTokenSwitchArgument(t *testing.T) {
	t.Parallel()

	var captured func(string) []string
	r := &captureCompleterReader{
		LineReader: newScannerLineReader(&bytes.Buffer{}, &bytes.Buffer{}),
		set:        func(fn func(string) []string) { captured = fn },
	}

	lister := &fakeLister{
		topLevel: []string{"switch", "games"},
		args:     map[string][]string{"switch": {"blackjack", "baccarat", "poker"}},
	}
	installCompleter(r, lister)

	cases := []struct {
		name   string
		prefix string
		want   []string
	}{
		{
			name:   "switch + space returns all game names",
			prefix: "switch ",
			want:   []string{"blackjack", "baccarat", "poker"},
		},
		{
			name:   "switch + partial game name filters by prefix",
			prefix: "switch b",
			want:   []string{"blackjack", "baccarat"},
		},
		{
			name:   "switch + unknown prefix returns nothing",
			prefix: "switch zzz",
			want:   nil,
		},
		{
			name:   "command without arg-completion returns nothing on second token",
			prefix: "games ",
			want:   nil,
		},
		{
			name:   "third token has no completion contract",
			prefix: "switch blackjack ",
			want:   nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := captured(tc.prefix)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestInstallCompleter_NilListerSkipsAllListerCandidates(t *testing.T) {
	t.Parallel()
	var captured func(string) []string
	r := &captureCompleterReader{
		LineReader: newScannerLineReader(&bytes.Buffer{}, &bytes.Buffer{}),
		set:        func(fn func(string) []string) { captured = fn },
	}
	installCompleter(r, nil)
	// First-token completion still works for the common-command set.
	assert.Equal(t, []string{"help"}, captured("h"))
	// Second-token completion is a no-op without a lister.
	assert.Nil(t, captured("switch "))
}

// TestInstallCompleter_DoesNotMutatePackageSlice guards against accidentally
// using `commonCompletionCommands` as the seed for the candidate slice. If
// `append` ever resizes via the package-level backing array, the next call
// would see corrupted output. We invoke the completer enough times that any
// such corruption would surface as a length or content drift.
func TestInstallCompleter_DoesNotMutatePackageSlice(t *testing.T) {
	t.Parallel()
	var captured func(string) []string
	r := &captureCompleterReader{
		LineReader: newScannerLineReader(&bytes.Buffer{}, &bytes.Buffer{}),
		set:        func(fn func(string) []string) { captured = fn },
	}
	original := append([]string(nil), commonCompletionCommands...)
	lister := &fakeLister{topLevel: []string{"switch", "games", "extra"}}
	installCompleter(r, lister)
	for range 5 {
		_ = captured("")
	}
	assert.Equal(t, original, commonCompletionCommands,
		"installCompleter must not mutate the package-level commonCompletionCommands slice")
}

func TestGameManager_CompletionCandidates(t *testing.T) {
	t.Parallel()
	mgr := NewGameManager("blackjack")
	got := mgr.CompletionCandidates()
	// Only manager-level commands are first-token candidates. Bare game
	// names are intentionally excluded — they're reachable via
	// ArgumentCandidates("switch").
	assert.Equal(t, []string{"switch", "games"}, got)
}

func TestGameManager_ArgumentCandidates(t *testing.T) {
	t.Parallel()
	mgr := NewGameManager("blackjack")

	switchArgs := mgr.ArgumentCandidates("switch")
	assert.Contains(t, switchArgs, "blackjack")
	assert.Contains(t, switchArgs, "poker")
	// Aliases get included (e.g. "gin" → "ginrummy").
	assert.Contains(t, switchArgs, "gin", "aliases should be tab-completable as switch arguments")

	// Non-arg-completing commands return nil.
	assert.Nil(t, mgr.ArgumentCandidates("games"))
	assert.Nil(t, mgr.ArgumentCandidates("help"))
	assert.Nil(t, mgr.ArgumentCandidates("nonexistent"))
}

// fakeLister stubs commandLister with fixed top-level and per-arg sets.
type fakeLister struct {
	topLevel []string
	args     map[string][]string
}

func (f *fakeLister) CompletionCandidates() []string { return f.topLevel }
func (f *fakeLister) ArgumentCandidates(cmd string) []string {
	if f.args == nil {
		return nil
	}
	return f.args[cmd]
}

// captureCompleterReader wraps a LineReader to intercept SetCompleter so the
// installed function can be invoked from a test. Everything else delegates
// to the embedded reader.
type captureCompleterReader struct {
	LineReader
	set func(fn func(string) []string)
}

func (c *captureCompleterReader) SetCompleter(fn func(string) []string) {
	c.set(fn)
}
