//go:build test

package ui

import (
	"bufio"
	"bytes"
	"os"
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
			wantPrompt: "[poker] > Goodbye.\n",
			wantText:   "",
			wantExit:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout to verify prompt
			origStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			scanner := bufio.NewScanner(strings.NewReader(tt.input))
			text, exit := readInput(scanner, tt.gameName)

			_ = w.Close()
			os.Stdout = origStdout

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			_ = r.Close()

			assert.Equal(t, tt.wantPrompt, buf.String())
			assert.Equal(t, tt.wantText, text)
			assert.Equal(t, tt.wantExit, exit)
		})
	}
}
