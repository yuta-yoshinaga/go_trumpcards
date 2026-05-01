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

func TestInstallCompleter_FirstTokenOnly(t *testing.T) {
	t.Parallel()

	var captured func(string) []string
	r := &captureCompleterReader{
		LineReader: newScannerLineReader(&bytes.Buffer{}, &bytes.Buffer{}),
		set:        func(fn func(string) []string) { captured = fn },
	}

	installCompleter(r, &fakeLister{candidates: []string{"blackjack", "baccarat"}})

	cases := []struct {
		name   string
		prefix string
		want   []string
	}{
		{
			name:   "empty prefix returns all (common + game-specific)",
			prefix: "",
			want:   []string{"q", "quit", "exit", "r", "reset", "help", "?", "blackjack", "baccarat"},
		},
		{
			name:   "single token completes by prefix",
			prefix: "b",
			want:   []string{"blackjack", "baccarat"},
		},
		{
			name:   "second token suppresses completion",
			prefix: "switch ",
			want:   nil,
		},
		{
			name:   "common command prefix matches",
			prefix: "q",
			want:   []string{"q", "quit"},
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

func TestInstallCompleter_NilListerSkipsGameCandidates(t *testing.T) {
	t.Parallel()
	var captured func(string) []string
	r := &captureCompleterReader{
		LineReader: newScannerLineReader(&bytes.Buffer{}, &bytes.Buffer{}),
		set:        func(fn func(string) []string) { captured = fn },
	}
	installCompleter(r, nil)
	got := captured("h")
	assert.Equal(t, []string{"help"}, got)
}

func TestGameManager_CompletionCandidates(t *testing.T) {
	t.Parallel()
	mgr := NewGameManager("blackjack")
	got := mgr.CompletionCandidates()
	assert.Contains(t, got, "switch")
	assert.Contains(t, got, "games")
	assert.Contains(t, got, "blackjack")
	assert.Contains(t, got, "poker")
	// Aliases get included (e.g. "gin" → "ginrummy").
	assert.Contains(t, got, "gin", "aliases should be tab-completable")
}

// fakeLister stubs commandLister with a fixed candidate set.
type fakeLister struct {
	candidates []string
}

func (f *fakeLister) CompletionCandidates() []string { return f.candidates }

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
