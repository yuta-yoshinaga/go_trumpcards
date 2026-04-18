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
		{"-help present (Go flag treats as --help)", []string{"-help"}, true},
		{"--h present (Go flag treats as -h)", []string{"--h"}, true},
		{"help mixed with other args", []string{"--lang", "en", "--help"}, true},
		{"short flag among multiple", []string{"foo", "-h", "bar"}, true},
		{"not a help flag", []string{"--helpful", "-help-me"}, false},
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

func TestRunHelpCommandForBuiltinSubcommand(t *testing.T) {
	// "web" is a builtin subcommand (not a game). runHelpCommand should fall
	// through the game-registry lookup and emit the entry from builtinSubcommandHelp.
	var stdout, stderr bytes.Buffer
	helpText := buildHelpText()
	code := runHelpCommand([]string{"web"}, helpText, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("runHelpCommand(web) exit = %d, want 0", code)
	}
	if stderr.Len() != 0 {
		t.Errorf("expected no stderr output; got %q", stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "trumpcards web") {
		t.Errorf("expected web help text; got: %q", out)
	}
}

func TestRunHelpCommandExtraArgs(t *testing.T) {
	// Extra positional args after the game name should trigger a warning
	// on stderr but still return the game's help on stdout with exit 0.
	var stdout, stderr bytes.Buffer
	helpText := buildHelpText()
	code := runHelpCommand([]string{"blackjack", "extra"}, helpText, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("runHelpCommand(blackjack extra) exit = %d, want 0", code)
	}
	if stdout.Len() == 0 {
		t.Errorf("expected blackjack help on stdout despite extra args")
	}
	if stderr.Len() == 0 {
		t.Errorf("expected extra-args warning on stderr; got empty")
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

func TestPrintGamesLongModeAlwaysIncludesAliases(t *testing.T) {
	// Pick a canonical name with known aliases, e.g. ginrummy -> gin.
	var withAlias string
	for alias, canonical := range ui.GameAliases {
		if alias != "" && canonical != "" {
			withAlias = canonical
			break
		}
	}
	if withAlias == "" {
		t.Skip("no aliases registered; cannot verify long-mode alias inlining")
	}
	var buf bytes.Buffer
	// aliases=false — long mode must STILL show aliases inline.
	printGames(false, false, &buf)
	out := buf.String()
	if !strings.Contains(out, "[aliases:") {
		t.Errorf("long mode output should contain inline '[aliases:' for games with aliases; got:\n%s", out)
	}
}

func TestPrintGamesShortModeRespectsAliasesFlag(t *testing.T) {
	var withAlias string
	var aliasSample string
	for alias, canonical := range ui.GameAliases {
		if alias != "" && canonical != "" {
			withAlias = canonical
			aliasSample = alias
			break
		}
	}
	if withAlias == "" {
		t.Skip("no aliases registered")
	}

	var without, with bytes.Buffer
	printGames(true, false, &without)
	printGames(true, true, &with)

	// Without --aliases, alias lines should not appear.
	if strings.Contains(without.String(), "\n"+aliasSample+"\n") || strings.HasPrefix(without.String(), aliasSample+"\n") {
		t.Errorf("short mode without --aliases should not list alias %q; got:\n%s", aliasSample, without.String())
	}
	// With --aliases, alias lines must appear.
	if !strings.Contains(with.String(), aliasSample) {
		t.Errorf("short mode with --aliases should list alias %q; got:\n%s", aliasSample, with.String())
	}
}

func TestCliAliasesWithoutShortKeyRemoved(t *testing.T) {
	// The `cliAliasesWithoutShort` warning contradicted the long-mode behavior
	// and was removed; ensure it hasn't crept back into either locale.
	if got := i18n.T("cliAliasesWithoutShort"); got != "cliAliasesWithoutShort" && got != "" {
		t.Errorf("i18n key 'cliAliasesWithoutShort' should be removed but still resolves to: %q", got)
	}
}
