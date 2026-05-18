//go:build test

package ui

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadInput_WithGameName(t *testing.T) {
	tests := []struct {
		name       string
		gameName   string
		input      string
		wantPrompt string
		wantText   string
		wantExit   bool
	}{
		{
			name:       "game name shown in prompt",
			gameName:   "blackjack",
			input:      "hit\n",
			wantPrompt: "[blackjack] > ",
			wantText:   "hit",
			wantExit:   false,
		},
		{
			name:       "empty game name uses plain prompt",
			gameName:   "",
			input:      "help\n",
			wantPrompt: "> ",
			wantText:   "help",
			wantExit:   false,
		},
		{
			name:       "EOF returns exit true",
			gameName:   "poker",
			input:      "",
			wantPrompt: "[poker] > ",
			wantText:   "",
			wantExit:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var promptBuf bytes.Buffer
			reader := newScannerLineReader(strings.NewReader(tt.input), &promptBuf)
			text, exit, ioErr := readInput(reader, tt.gameName)

			assert.Equal(t, tt.wantPrompt, promptBuf.String())
			assert.Equal(t, tt.wantText, text)
			assert.Equal(t, tt.wantExit, exit)
			// EOF and normal lines must both report a nil ioErr; only
			// non-EOF I/O errors carry one (issue #1839).
			assert.NoError(t, ioErr)
		})
	}
}

// erroringLineReader returns a fixed non-EOF error from Prompt. Used to
// exercise the issue #1839 contract where readInput surfaces real read
// failures as a non-nil ioErr while keeping EOF clean.
type erroringLineReader struct{ err error }

func (e *erroringLineReader) Prompt(_ string) (string, error)      { return "", e.err }
func (e *erroringLineReader) AppendHistory(_ string)               {}
func (e *erroringLineReader) SetCompleter(_ func(string) []string) {}
func (e *erroringLineReader) Close() error                         { return nil }

func TestReadInput_NonEOFError_ReturnsIOErr(t *testing.T) {
	t.Parallel()
	sentinel := errors.New("pty hung up")
	reader := &erroringLineReader{err: sentinel}
	text, exit, ioErr := readInput(reader, "blackjack")
	assert.Empty(t, text)
	assert.True(t, exit, "non-EOF error must still stop the loop")
	assert.ErrorIs(t, ioErr, sentinel, "non-EOF error must propagate as ioErr")
}
