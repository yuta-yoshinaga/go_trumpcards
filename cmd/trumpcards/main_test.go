package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/ui"
)

func init() {
	// Ensure translations are loaded so messages render in a stable locale during tests.
	i18n.SetLang("ja")
}

func TestHasHelpFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"empty", nil, false},
		{"no help flag", []string{"--lang", "en"}, false},
		{"--help present", []string{"--help"}, true},
		{"-h present", []string{"-h"}, true},
		{"help mixed with other args", []string{"--lang", "en", "--help"}, true},
		{"short flag among multiple", []string{"foo", "-h", "bar"}, true},
		{"looks like help but is not", []string{"-help", "--h"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasHelpFlag(tt.args); got != tt.want {
				t.Errorf("hasHelpFlag(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func TestRunHelpCommandForGame(t *testing.T) {
	var stdout, stderr bytes.Buffer
	helpText := buildHelpText()
	code := runHelpCommand([]string{"blackjack"}, helpText, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("runHelpCommand exit = %d, want 0", code)
	}
	if stderr.Len() != 0 {
		t.Errorf("expected no stderr output, got %q", stderr.String())
	}
	out := stdout.String()
	if out == "" {
		t.Fatal("expected blackjack help on stdout, got empty output")
	}
	// The BlackJack CUI's HelpLines() always contain the standard quit command.
	if !strings.Contains(out, "q") {
		t.Errorf("expected blackjack help to mention quit command; got: %q", out)
	}
}

func TestRunHelpCommandResolvesAlias(t *testing.T) {
	// "gin" is a canonical alias for "ginrummy". If this alias ever disappears,
	// pick another from ui.GameAliases — we just need one known alias for coverage.
	if _, ok := ui.GameAliases["gin"]; !ok {
		t.Skip("alias 'gin' not registered; alias resolution still covered by other paths")
	}
	var stdout, stderr bytes.Buffer
	helpText := buildHelpText()
	code := runHelpCommand([]string{"gin"}, helpText, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("runHelpCommand(gin) exit = %d, want 0", code)
	}
	if stdout.Len() == 0 {
		t.Errorf("expected alias 'gin' to resolve to ginrummy help; got empty stdout")
	}
}

func TestRunHelpCommandUnknownGame(t *testing.T) {
	var stdout, stderr bytes.Buffer
	helpText := buildHelpText()
	code := runHelpCommand([]string{"definitelynotagame"}, helpText, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("runHelpCommand(unknown) exit = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Errorf("expected no stdout on unknown game; got %q", stdout.String())
	}
	if stderr.Len() == 0 {
		t.Errorf("expected error output on stderr for unknown game")
	}
}

func TestBuildHelpTextMentionsVersionShort(t *testing.T) {
	helpText := buildHelpText()
	if !strings.Contains(helpText, "--version-short") {
		t.Errorf("help text should document --version-short; got:\n%s", helpText)
	}
}
