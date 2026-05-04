//go:build test && (!js || !wasm)

package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewLinerReader_ConstructDoesNotPanic verifies the construction path is
// safe in a non-TTY test runner. Driving interactive input through liner
// requires a real terminal and isn't covered here — the wrapper itself is
// thin enough that "constructs cleanly + closes cleanly" is the boundary
// worth asserting.
func TestNewLinerReader_ConstructDoesNotPanic(t *testing.T) {
	r := newLinerReader()
	assert.NotNil(t, r)
	assert.NotNil(t, r.state)
	// AppendHistory and SetCompleter should be safe to call even before any prompt.
	r.AppendHistory("")     // empty path — guarded
	r.AppendHistory("test") // non-empty — forwarded to liner
	r.SetCompleter(func(string) []string { return nil })
	assert.NoError(t, r.Close())
}

// TestNewDefaultLineReader_ReturnsLiner verifies the build-tag wiring picks
// the liner-backed reader on native builds. If a regression accidentally
// stubs out newDefaultLineReader on non-WASM platforms, this fails fast.
func TestNewDefaultLineReader_ReturnsLiner(t *testing.T) {
	r := newDefaultLineReader()
	defer func() { _ = r.Close() }()
	_, ok := r.(*linerLineReader)
	assert.True(t, ok, "newDefaultLineReader on native build should return *linerLineReader, got %T", r)
}
