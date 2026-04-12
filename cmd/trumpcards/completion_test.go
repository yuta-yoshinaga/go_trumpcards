package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/ui"
)

func TestWriteBashCompletion(t *testing.T) {
	var buf bytes.Buffer
	err := writeBashCompletion(&buf)
	require.NoError(t, err)
	script := buf.String()
	assert.Contains(t, script, "complete -F _trumpcards trumpcards")
	assert.Contains(t, script, "--lang")
	assert.Contains(t, script, "--port")
	// Every subcommand in the canonical list must appear in the bash script.
	for _, cmd := range completionSubcommands() {
		assert.Contains(t, script, cmd, "bash completion missing subcommand %q", cmd)
	}
}

func TestWriteZshCompletion(t *testing.T) {
	var buf bytes.Buffer
	err := writeZshCompletion(&buf)
	require.NoError(t, err)
	script := buf.String()
	assert.Contains(t, script, "#compdef trumpcards")
	assert.Contains(t, script, "--port")
	// Every subcommand in the canonical list must appear in the zsh script.
	for _, cmd := range completionSubcommands() {
		assert.Contains(t, script, "'"+cmd+":", "zsh completion missing subcommand %q", cmd)
	}
}

func TestWriteFishCompletion(t *testing.T) {
	var buf bytes.Buffer
	err := writeFishCompletion(&buf)
	require.NoError(t, err)
	script := buf.String()
	assert.Contains(t, script, "complete -c trumpcards")
	assert.Contains(t, script, "__fish_seen_subcommand_from completion")
	assert.Contains(t, script, "-l port")
	// Every subcommand in the canonical list must appear in the fish script.
	for _, cmd := range completionSubcommands() {
		assert.Contains(t, script, "-a "+cmd, "fish completion missing subcommand %q", cmd)
	}
}

func TestWriteInstallHint(t *testing.T) {
	tests := []struct {
		shell    string
		contains []string
	}{
		{"bash", []string{"~/.bashrc", "source <(trumpcards completion bash)"}},
		{"zsh", []string{"fpath", "compinit", "trumpcards completion zsh"}},
		{"fish", []string{"~/.config/fish/completions/trumpcards.fish", "trumpcards completion fish | source"}},
	}
	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			var buf bytes.Buffer
			writeInstallHint(&buf, tt.shell)
			output := buf.String()
			for _, s := range tt.contains {
				assert.Contains(t, output, s, "install hint for %s should contain %q", tt.shell, s)
			}
		})
	}
}

func TestWriteInstallHint_UnsupportedShell(t *testing.T) {
	var buf bytes.Buffer
	writeInstallHint(&buf, "powershell")
	assert.Empty(t, buf.String())
}

func TestRunCompletion_NoArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runCompletionTo(nil, &stdout, &stderr)
	assert.Equal(t, 1, code)
	assert.Contains(t, stderr.String(), "trumpcards completion")
}

func TestRunCompletion_ExtraArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runCompletionTo([]string{"bash", "extra"}, &stdout, &stderr)
	assert.Equal(t, 1, code)
}

func TestRunCompletion_UnsupportedShell(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runCompletionTo([]string{"powershell"}, &stdout, &stderr)
	assert.Equal(t, 1, code)
	assert.Contains(t, stderr.String(), "powershell")
}

func TestRunCompletion_UnsupportedShell_DidYouMean(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"bashh", "bash"},
		{"fsh", "zsh"}, // fsh is closer to zsh than to fish (distance 2 vs 2 — first match wins)
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := runCompletionTo([]string{tt.input}, &stdout, &stderr)
			assert.Equal(t, 1, code)
			assert.Contains(t, stderr.String(), tt.want, "expected suggestion %q for input %q", tt.want, tt.input)
		})
	}
}

func TestRunCompletion_UnsupportedShell_NoSuggestion(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runCompletionTo([]string{"powershell"}, &stdout, &stderr)
	assert.Equal(t, 1, code)
	// "powershell" is too far from any supported shell — no Did-you-mean line.
	assert.NotContains(t, stderr.String(), "Did you mean")
	assert.NotContains(t, stderr.String(), "もしかして")
}

func TestRunCompletion_ValidShells(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish"} {
		t.Run(shell, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := runCompletionTo([]string{shell}, &stdout, &stderr)
			assert.Equal(t, 0, code)
			assert.NotEmpty(t, stdout.String())
			assert.Empty(t, stderr.String())
		})
	}
}

func TestRunHelpCommand_NoArgs_PrintsHelpText(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runHelpCommand(nil, "HELP_TEXT_FIXTURE", &stdout, &stderr)
	assert.Equal(t, 0, code)
	assert.Equal(t, "HELP_TEXT_FIXTURE", stdout.String())
	assert.Empty(t, stderr.String())
}

func TestRunHelpCommand_KnownGame(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runHelpCommand([]string{"blackjack"}, "", &stdout, &stderr)
	assert.Equal(t, 0, code)
	assert.NotEmpty(t, stdout.String(), "should print game help lines")
	assert.Empty(t, stderr.String())
}

func TestRunHelpCommand_AliasResolved(t *testing.T) {
	// "gin" is an alias for "ginrummy"
	var stdout, stderr bytes.Buffer
	code := runHelpCommand([]string{"gin"}, "", &stdout, &stderr)
	assert.Equal(t, 0, code)
	assert.NotEmpty(t, stdout.String())
}

func TestRunHelpCommand_CaseInsensitive(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runHelpCommand([]string{"BlackJack"}, "", &stdout, &stderr)
	assert.Equal(t, 0, code)
	assert.NotEmpty(t, stdout.String())
}

func TestRunHelpCommand_UnknownGame(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runHelpCommand([]string{"nosuchgame"}, "", &stdout, &stderr)
	assert.Equal(t, 1, code)
	assert.Empty(t, stdout.String())
	assert.Contains(t, stderr.String(), "nosuchgame")
}

func TestRunHelpCommand_UnknownGame_DidYouMean(t *testing.T) {
	// "blackjac" → distance 1 from "blackjack"
	var stdout, stderr bytes.Buffer
	code := runHelpCommand([]string{"blackjac"}, "", &stdout, &stderr)
	assert.Equal(t, 1, code)
	assert.Contains(t, stderr.String(), "blackjack")
}

func TestCompletionSubcommands_ContainsAllGames(t *testing.T) {
	// Verify the list is sorted
	for i := 1; i < len(completionSubcommands()); i++ {
		assert.True(t, completionSubcommands()[i-1] < completionSubcommands()[i],
			"completionSubcommands() not sorted: %q >= %q", completionSubcommands()[i-1], completionSubcommands()[i])
	}

	// Verify key games are present
	expected := []string{"blackjack", "poker", "web", "update", "completion", "golf"}
	joined := strings.Join(completionSubcommands(), " ")
	for _, e := range expected {
		assert.Contains(t, joined, e)
	}
}

func TestCompletionSubcommands_SyncWithGameNames(t *testing.T) {
	subSet := make(map[string]bool, len(completionSubcommands()))
	for _, s := range completionSubcommands() {
		subSet[s] = true
	}
	// Every game in ui.GameNames must appear in completionSubcommands().
	for _, name := range ui.GameNames() {
		assert.True(t, subSet[name], "completionSubcommands() missing game %q from ui.GameNames()", name)
	}
	// Every alias in ui.GameAliases must appear in completionSubcommands().
	for alias := range ui.GameAliases {
		assert.True(t, subSet[alias], "completionSubcommands() missing alias %q from ui.GameAliases", alias)
	}
}
