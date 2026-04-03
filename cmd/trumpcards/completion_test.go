package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteBashCompletion(t *testing.T) {
	var buf bytes.Buffer
	err := writeBashCompletion(&buf)
	require.NoError(t, err)
	script := buf.String()
	assert.Contains(t, script, "complete -F _trumpcards trumpcards")
	assert.Contains(t, script, "blackjack")
	assert.Contains(t, script, "--lang")
	assert.Contains(t, script, "--port")
}

func TestWriteZshCompletion(t *testing.T) {
	var buf bytes.Buffer
	err := writeZshCompletion(&buf)
	require.NoError(t, err)
	script := buf.String()
	assert.Contains(t, script, "#compdef trumpcards")
	assert.Contains(t, script, "'blackjack:BlackJack'")
	assert.Contains(t, script, "'completion:Generate shell completion script'")
	assert.Contains(t, script, "--port")
}

func TestWriteFishCompletion(t *testing.T) {
	var buf bytes.Buffer
	err := writeFishCompletion(&buf)
	require.NoError(t, err)
	script := buf.String()
	assert.Contains(t, script, "complete -c trumpcards")
	assert.Contains(t, script, "blackjack")
	assert.Contains(t, script, "__fish_seen_subcommand_from completion")
	assert.Contains(t, script, "-l port")
}

func TestRunCompletion_NoArgs(t *testing.T) {
	code := runCompletion(nil)
	assert.Equal(t, 1, code)
}

func TestRunCompletion_UnsupportedShell(t *testing.T) {
	code := runCompletion([]string{"powershell"})
	assert.Equal(t, 1, code)
}

func TestRunCompletion_ValidShells(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish"} {
		t.Run(shell, func(t *testing.T) {
			code := runCompletion([]string{shell})
			assert.Equal(t, 0, code)
		})
	}
}

func TestCompletionSubcommands_ContainsAllGames(t *testing.T) {
	// Verify the list is sorted
	for i := 1; i < len(completionSubcommands); i++ {
		assert.True(t, completionSubcommands[i-1] < completionSubcommands[i],
			"completionSubcommands not sorted: %q >= %q", completionSubcommands[i-1], completionSubcommands[i])
	}

	// Verify key games are present
	expected := []string{"blackjack", "poker", "web", "update", "completion", "golf"}
	joined := strings.Join(completionSubcommands, " ")
	for _, e := range expected {
		assert.Contains(t, joined, e)
	}
}
