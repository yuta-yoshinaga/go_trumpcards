package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
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

func TestRunUnknownTopLevelFlagIsI18nError(t *testing.T) {
	// Redirect stderr and stdout through os.Pipe so we can capture what
	// run() writes when it encounters an unknown top-level flag.
	// `flag.CommandLine` is global state, so we restore it after each run.
	origArgs := os.Args
	origCmdLine := flag.CommandLine
	origStdout := os.Stdout
	origStderr := os.Stderr
	defer func() {
		os.Args = origArgs
		flag.CommandLine = origCmdLine
		os.Stdout = origStdout
		os.Stderr = origStderr
	}()

	cases := []struct {
		name        string
		args        []string
		wantExit    int
		wantPrefix  string
		wantInclude string
	}{
		{
			name:        "ja locale wraps error in cliFlagError",
			args:        []string{"trumpcards", "--lang", "ja", "--bogus"},
			wantExit:    2,
			wantPrefix:  "エラー: 不明なオプション",
			wantInclude: "-bogus",
		},
		{
			name:        "en locale wraps error in cliFlagError",
			args:        []string{"trumpcards", "--lang", "en", "--bogus"},
			wantExit:    2,
			wantPrefix:  "Error: invalid option",
			wantInclude: "-bogus",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate production: Go's package init gives flag.CommandLine a
			// Usage func that writes the help text. A fresh FlagSet has
			// Usage=nil by default, which would hide a regression where we
			// forget to suppress the internal Usage callback.
			flag.CommandLine = flag.NewFlagSet(tc.args[0], flag.ExitOnError)
			flag.CommandLine.Usage = func() { _, _ = fmt.Fprint(os.Stderr, "USAGE: simulated-prod-help\n") }
			os.Args = tc.args

			rOut, wOut, _ := os.Pipe()
			rErr, wErr, _ := os.Pipe()
			os.Stdout = wOut
			os.Stderr = wErr

			exitCh := make(chan int, 1)
			go func() { exitCh <- run() }()
			exit := <-exitCh

			_ = wOut.Close()
			_ = wErr.Close()
			var outBuf, errBuf bytes.Buffer
			_, _ = outBuf.ReadFrom(rOut)
			_, _ = errBuf.ReadFrom(rErr)

			if exit != tc.wantExit {
				t.Errorf("exit = %d, want %d (stderr=%q)", exit, tc.wantExit, errBuf.String())
			}
			errStr := errBuf.String()
			if !strings.HasPrefix(errStr, tc.wantPrefix) {
				t.Errorf("stderr should start with i18n envelope %q; got prefix %q", tc.wantPrefix, firstLine(errStr))
			}
			if !strings.Contains(errStr, tc.wantInclude) {
				t.Errorf("stderr should include offending flag %q; got: %q", tc.wantInclude, firstLine(errStr))
			}
			if n := strings.Count(errStr, "USAGE:"); n != 1 {
				t.Errorf("help text should be printed exactly once; got %d USAGE: markers", n)
			}
		})
	}
}

func firstLine(s string) string {
	line, _, _ := strings.Cut(s, "\n")
	return line
}

func TestDetectBootstrapLang(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		langEnv string
		want    string
	}{
		{"empty args, empty env -> default", nil, "", "ja"},
		{"LANG=en_US.UTF-8 -> en", nil, "en_US.UTF-8", "en"},
		{"LANG=ja_JP.UTF-8 -> ja", nil, "ja_JP.UTF-8", "ja"},
		{"LANG=fr -> ja (unsupported falls back)", nil, "fr", "ja"},
		{"--lang en overrides LANG", []string{"--lang", "en"}, "ja_JP.UTF-8", "en"},
		{"--lang=ja overrides LANG", []string{"--lang=ja"}, "en_US.UTF-8", "ja"},
		{"-lang en (single dash) also works", []string{"-lang", "en"}, "", "en"},
		{"--lang garbage ignored -> fall back to LANG", []string{"--lang", "klingon"}, "en_US.UTF-8", "en"},
		{"--lang with no value -> ignore", []string{"--lang"}, "en_US.UTF-8", "en"},
		{"--lang appears after other flag", []string{"--bogus", "--lang", "en"}, "", "en"},
		// Gemini review: must mirror flag.Parse semantics.
		{"stops at first positional arg (like flag.Parse)", []string{"blackjack", "--lang", "en"}, "", "ja"},
		{"stops at -- end-of-flags terminator", []string{"--", "--lang", "en"}, "", "ja"},
		{"last --lang wins on repeat", []string{"--lang", "ja", "--lang", "en"}, "", "en"},
		{"last --lang=form wins on repeat", []string{"--lang=en", "--lang=ja"}, "", "ja"},
		{"unsupported last value keeps earlier valid one", []string{"--lang", "en", "--lang", "klingon"}, "", "en"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detectBootstrapLang(tt.args, tt.langEnv); got != tt.want {
				t.Errorf("detectBootstrapLang(%v, %q) = %q, want %q", tt.args, tt.langEnv, got, tt.want)
			}
		})
	}
}

func TestBuildHelpTextMentionsVersionShort(t *testing.T) {
	helpText := buildHelpText()
	if !strings.Contains(helpText, "--version-short") {
		t.Errorf("help text should document --version-short; got:\n%s", helpText)
	}
}
