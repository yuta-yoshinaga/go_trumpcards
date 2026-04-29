package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/ui"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/update"
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

func TestPortInRange(t *testing.T) {
	tests := []struct {
		port int
		want bool
	}{
		{0, true}, // ephemeral (POSIX)
		{1, true},
		{8080, true},
		{65535, true}, // max
		{-1, false},   // must reject (previously collided with the portUnset sentinel)
		{-2, false},
		{65536, false},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("port=%d", tt.port), func(t *testing.T) {
			if got := portInRange(tt.port); got != tt.want {
				t.Errorf("portInRange(%d) = %v, want %v", tt.port, got, tt.want)
			}
		})
	}
}

func TestFlagSetVisitedDistinguishesExplicitFromDefault(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantVisited bool
		wantPortVal int
	}{
		{"no flag -> not visited, default 0", nil, false, 0},
		{"--port 0 -> visited (critical: collides with default value)", []string{"--port", "0"}, true, 0},
		{"--port -1 -> visited (must not be misclassified as unset)", []string{"--port", "-1"}, true, -1},
		{"--port 3000 -> visited", []string{"--port", "3000"}, true, 3000},
		{"-p 8080 shorthand -> visited", []string{"-p", "8080"}, true, 8080},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var port int
			fs := flag.NewFlagSet("web", flag.ContinueOnError)
			fs.SetOutput(io.Discard)
			fs.IntVar(&port, "port", 0, "")
			fs.IntVar(&port, "p", 0, "")
			if err := fs.Parse(tt.args); err != nil {
				t.Fatalf("parse error: %v", err)
			}
			if got := flagSetVisited(fs, "port", "p"); got != tt.wantVisited {
				t.Errorf("flagSetVisited = %v, want %v", got, tt.wantVisited)
			}
			if port != tt.wantPortVal {
				t.Errorf("port = %d, want %d", port, tt.wantPortVal)
			}
		})
	}
}

// TestRunWebRejectsExplicitInvalidPort covers the bug Gemini and Claude
// flagged: under the previous portUnset=-1 sentinel, `--port -1` silently
// fell through to the default 8080. With the fs.Visit approach the explicit
// -1 must be rejected with cliInvalidPort and exit 1 before any bind.
func TestRunWebRejectsExplicitInvalidPort(t *testing.T) {
	origArgs := os.Args
	origCmdLine := flag.CommandLine
	origStdout := os.Stdout
	origStderr := os.Stderr
	origPortEnv, portEnvWasSet := os.LookupEnv("PORT")
	defer func() {
		os.Args = origArgs
		flag.CommandLine = origCmdLine
		os.Stdout = origStdout
		os.Stderr = origStderr
		if portEnvWasSet {
			_ = os.Setenv("PORT", origPortEnv)
		} else {
			_ = os.Unsetenv("PORT")
		}
	}()
	_ = os.Unsetenv("PORT") // baseline

	flag.CommandLine = flag.NewFlagSet("trumpcards", flag.ExitOnError)
	os.Args = []string{"trumpcards", "web", "--port", "-1"}

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

	if exit != 1 {
		t.Errorf("exit = %d, want 1 (stderr=%q)", exit, errBuf.String())
	}
	if !strings.Contains(errBuf.String(), "-1") {
		t.Errorf("stderr should mention the offending port; got: %q", errBuf.String())
	}
	// Guard against regression: the old sentinel behavior would silently
	// pass through and try to bind 8080, leaving PORT untouched.
	if got := os.Getenv("PORT"); got != "" {
		t.Errorf("PORT should not be set on invalid input; got %q", got)
	}
}

// TestRunUnsupportedLangFlagFallsBackAndWarns verifies issue #1448: an
// unsupported --lang value must warn on stderr and fall back (not exit 2).
// This matches the behavior of an unsupported LANG env var and keeps `set -e`
// scripts from tripping on typos.
func TestRunUnsupportedLangFlagFallsBackAndWarns(t *testing.T) {
	origArgs := os.Args
	origCmdLine := flag.CommandLine
	origStdout := os.Stdout
	origStderr := os.Stderr
	origQuiet, quietWasSet := os.LookupEnv("TRUMPCARDS_QUIET")
	origLang, langWasSet := os.LookupEnv("LANG")
	defer func() {
		os.Args = origArgs
		flag.CommandLine = origCmdLine
		os.Stdout = origStdout
		os.Stderr = origStderr
		if quietWasSet {
			_ = os.Setenv("TRUMPCARDS_QUIET", origQuiet)
		} else {
			_ = os.Unsetenv("TRUMPCARDS_QUIET")
		}
		if langWasSet {
			_ = os.Setenv("LANG", origLang)
		} else {
			_ = os.Unsetenv("LANG")
		}
		i18n.SetLang("ja") // restore test default
	}()

	cases := []struct {
		name      string
		quiet     bool
		wantWarn  bool
		wantLang  string // expected post-run i18n lang (falls back to "ja")
		wantExit  int
		wantWords []string
	}{
		{
			name:      "warn and fall back when TRUMPCARDS_QUIET unset",
			quiet:     false,
			wantWarn:  true,
			wantLang:  "ja",
			wantExit:  0,
			wantWords: []string{"fr"}, // offending value must appear
		},
		{
			name:     "silent when TRUMPCARDS_QUIET=1",
			quiet:    true,
			wantWarn: false,
			wantLang: "ja",
			wantExit: 0,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_ = os.Unsetenv("LANG") // isolate from caller env
			if tc.quiet {
				_ = os.Setenv("TRUMPCARDS_QUIET", "1")
			} else {
				_ = os.Unsetenv("TRUMPCARDS_QUIET")
			}
			flag.CommandLine = flag.NewFlagSet("trumpcards", flag.ExitOnError)
			// Use `games --short` which exits 0 and doesn't hang on stdin,
			// but still exercises the --lang validation block.
			os.Args = []string{"trumpcards", "--lang", "fr", "games", "--short"}

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
			hasWarn := strings.Contains(errBuf.String(), "fr")
			if hasWarn != tc.wantWarn {
				t.Errorf("warning on stderr: got=%v, want=%v (stderr=%q)", hasWarn, tc.wantWarn, errBuf.String())
			}
			for _, w := range tc.wantWords {
				if !strings.Contains(errBuf.String(), w) {
					t.Errorf("stderr missing %q: %q", w, errBuf.String())
				}
			}
			// The fallback must produce a usable locale — `games --short`
			// still prints game names to stdout, exit 0.
			if outBuf.Len() == 0 {
				t.Errorf("expected non-empty stdout from `games --short` after fallback")
			}
		})
	}
}

// TestRunGlobalQuietFlagSuppressesWarnings verifies issue #1553: the global
// --quiet/-q flag is OR-combined with TRUMPCARDS_QUIET, so users can suppress
// the locale-fallback warning without exporting an env var first. Both the
// long form (`--quiet`) and the POSIX shorthand (`-q`) must work.
func TestRunGlobalQuietFlagSuppressesWarnings(t *testing.T) {
	origArgs := os.Args
	origCmdLine := flag.CommandLine
	origStdout := os.Stdout
	origStderr := os.Stderr
	origQuiet, quietWasSet := os.LookupEnv("TRUMPCARDS_QUIET")
	origLang, langWasSet := os.LookupEnv("LANG")
	defer func() {
		os.Args = origArgs
		flag.CommandLine = origCmdLine
		os.Stdout = origStdout
		os.Stderr = origStderr
		if quietWasSet {
			_ = os.Setenv("TRUMPCARDS_QUIET", origQuiet)
		} else {
			_ = os.Unsetenv("TRUMPCARDS_QUIET")
		}
		if langWasSet {
			_ = os.Setenv("LANG", origLang)
		} else {
			_ = os.Unsetenv("LANG")
		}
		i18n.SetLang("ja")
	}()

	cases := []struct {
		name string
		args []string
	}{
		{"--quiet long form", []string{"trumpcards", "--quiet", "--lang", "fr", "games", "--short"}},
		{"-q shorthand", []string{"trumpcards", "-q", "--lang", "fr", "games", "--short"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_ = os.Unsetenv("LANG")
			_ = os.Unsetenv("TRUMPCARDS_QUIET")
			flag.CommandLine = flag.NewFlagSet("trumpcards", flag.ExitOnError)
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

			if exit != 0 {
				t.Errorf("exit = %d, want 0 (stderr=%q)", exit, errBuf.String())
			}
			// The locale-fallback warning must be suppressed under quiet.
			if strings.Contains(errBuf.String(), "fr") {
				t.Errorf("expected no 'fr' fallback warning under quiet; got stderr=%q", errBuf.String())
			}
			// `games --short` must still produce its normal stdout output.
			if outBuf.Len() == 0 {
				t.Error("expected `games --short` stdout despite quiet")
			}
		})
	}
}

// TestRunGlobalQuietFlagAcceptedAfterSubcommand verifies the fix for
// PR #1582 review: subcommand FlagSets (games / update / completion) must
// not reject `--quiet`/`-q` placed after the subcommand name. Without
// this, `trumpcards games -q` exited 2 with "flag provided but not
// defined" — surprising for a flag advertised under OPTIONS as global.
//
// We exercise the games subcommand because it is the only one that
// produces both stdout and exit 0 reliably without external dependencies
// (update needs a network endpoint, completion needs a TTY for hints).
// The other two share the same flag-registration pattern.
func TestRunGlobalQuietFlagAcceptedAfterSubcommand(t *testing.T) {
	origArgs := os.Args
	origCmdLine := flag.CommandLine
	origStdout := os.Stdout
	origStderr := os.Stderr
	origQuiet, quietWasSet := os.LookupEnv("TRUMPCARDS_QUIET")
	defer func() {
		os.Args = origArgs
		flag.CommandLine = origCmdLine
		os.Stdout = origStdout
		os.Stderr = origStderr
		if quietWasSet {
			_ = os.Setenv("TRUMPCARDS_QUIET", origQuiet)
		} else {
			_ = os.Unsetenv("TRUMPCARDS_QUIET")
		}
	}()

	cases := []struct {
		name string
		args []string
	}{
		{"games -q --short", []string{"trumpcards", "games", "-q", "--short"}},
		{"games --quiet --short", []string{"trumpcards", "games", "--quiet", "--short"}},
		{"games --short -q (after another flag)", []string{"trumpcards", "games", "--short", "-q"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_ = os.Unsetenv("TRUMPCARDS_QUIET")
			flag.CommandLine = flag.NewFlagSet("trumpcards", flag.ExitOnError)
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

			if exit != 0 {
				t.Errorf("exit = %d, want 0 (stderr=%q)", exit, errBuf.String())
			}
			if strings.Contains(errBuf.String(), "flag provided but not defined") {
				t.Errorf("subcommand FlagSet must accept -q/--quiet; got stderr=%q", errBuf.String())
			}
			if outBuf.Len() == 0 {
				t.Error("expected `games --short` stdout output")
			}
		})
	}
}

// TestRunPosixLocalePlaceholderEnvSuppressesWarn verifies issue #1534: LANG
// values that are POSIX locale placeholders ("C", "POSIX", "C.UTF-8") do NOT
// trigger the cliLangEnvFallback warning, because they are the default in
// Docker base images, CI runners, and minimal Linux installs and signal
// "no language preference" rather than a specific unsupported language.
// Real unsupported languages (e.g. "fr_FR.UTF-8") still warn — the placeholder
// suppression must be a pinpoint exception, not a general silencer.
func TestRunPosixLocalePlaceholderEnvSuppressesWarn(t *testing.T) {
	origArgs := os.Args
	origCmdLine := flag.CommandLine
	origStdout := os.Stdout
	origStderr := os.Stderr
	origQuiet, quietWasSet := os.LookupEnv("TRUMPCARDS_QUIET")
	origLang, langWasSet := os.LookupEnv("LANG")
	defer func() {
		os.Args = origArgs
		flag.CommandLine = origCmdLine
		os.Stdout = origStdout
		os.Stderr = origStderr
		if quietWasSet {
			_ = os.Setenv("TRUMPCARDS_QUIET", origQuiet)
		} else {
			_ = os.Unsetenv("TRUMPCARDS_QUIET")
		}
		if langWasSet {
			_ = os.Setenv("LANG", origLang)
		} else {
			_ = os.Unsetenv("LANG")
		}
		i18n.SetLang("ja")
	}()

	cases := []struct {
		name     string
		lang     string
		wantWarn bool
	}{
		{"LANG=C suppresses warning", "C", false},
		{"LANG=POSIX suppresses warning", "POSIX", false},
		{"LANG=C.UTF-8 suppresses warning (Docker default)", "C.UTF-8", false},
		{"LANG=C.utf8 suppresses warning (no hyphen variant)", "C.utf8", false},
		{"LANG=POSIX.UTF-8 suppresses warning", "POSIX.UTF-8", false},
		{"LANG=fr_FR.UTF-8 still warns (real unsupported language)", "fr_FR.UTF-8", true},
		{"LANG=zz still warns (unknown prefix)", "zz", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_ = os.Unsetenv("TRUMPCARDS_QUIET")
			_ = os.Setenv("LANG", tc.lang)
			flag.CommandLine = flag.NewFlagSet("trumpcards", flag.ExitOnError)
			os.Args = []string{"trumpcards", "games", "--short"}

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

			if exit != 0 {
				t.Fatalf("exit = %d, want 0 (stderr=%q)", exit, errBuf.String())
			}
			// Detect the LANG-fallback warning specifically by looking for the
			// offending value on stderr; other unrelated warnings would not
			// embed `tc.lang`.
			gotWarn := strings.Contains(errBuf.String(), tc.lang)
			if gotWarn != tc.wantWarn {
				t.Errorf("warning emitted: got=%v, want=%v (stderr=%q)", gotWarn, tc.wantWarn, errBuf.String())
			}
			if outBuf.Len() == 0 {
				t.Errorf("expected non-empty stdout from `games --short`; got empty")
			}
		})
	}
}

// TestIsPosixLocalePlaceholder unit-tests the predicate used to gate the
// LANG-fallback warning. Kept separate from the integration test above so a
// regression in the predicate is easy to localize.
func TestIsPosixLocalePlaceholder(t *testing.T) {
	tests := []struct {
		prefix string
		want   bool
	}{
		{"C", true},
		{"POSIX", true},
		// Predicate operates on the prefix (split at . / _ / -) so callers
		// must strip the codeset themselves; the predicate sees only "C".
		{"c", false},     // case-sensitive: real LANG values use uppercase
		{"posix", false}, // ditto
		{"en", false},
		{"ja", false},
		{"fr", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.prefix, func(t *testing.T) {
			if got := isPosixLocalePlaceholder(tt.prefix); got != tt.want {
				t.Errorf("isPosixLocalePlaceholder(%q) = %v, want %v", tt.prefix, got, tt.want)
			}
		})
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
	printGames(false, false, "", &buf)
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
	printGames(true, false, "", &without)
	printGames(true, true, "", &with)

	// Without --aliases, alias lines should not appear.
	if strings.Contains(without.String(), "\n"+aliasSample+"\n") || strings.HasPrefix(without.String(), aliasSample+"\n") {
		t.Errorf("short mode without --aliases should not list alias %q; got:\n%s", aliasSample, without.String())
	}
	// With --aliases, alias lines must appear.
	if !strings.Contains(with.String(), aliasSample) {
		t.Errorf("short mode with --aliases should list alias %q; got:\n%s", aliasSample, with.String())
	}
}

// TestValidCategory pins down the predicate used to gate `--category`. The
// canonical strings must match games.Category.String() (casino / classic /
// solo) — drift between the CLI list and the games-pkg enum would silently
// reject valid categories or accept invalid ones.
func TestValidCategory(t *testing.T) {
	tests := []struct {
		s    string
		want bool
	}{
		{"casino", true},
		{"classic", true},
		{"solo", true},
		{"Casino", false}, // case-sensitive — match games.Category.String() form
		{"", false},
		{"poker", false}, // game name, not a category
		{"slot", false},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			if got := validCategory(tt.s); got != tt.want {
				t.Errorf("validCategory(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

// TestPrintGamesJSONFullEmitsEveryGame verifies that the JSON output of
// `games --json` (no filter) is a JSON array containing one entry per game in
// the registry, each with the expected schema.
func TestPrintGamesJSONFullEmitsEveryGame(t *testing.T) {
	var buf bytes.Buffer
	if err := printGamesJSON("", &buf); err != nil {
		t.Fatalf("printGamesJSON returned error: %v", err)
	}
	type entry struct {
		Name        string   `json:"name"`
		Category    string   `json:"category"`
		Description string   `json:"description"`
		Aliases     []string `json:"aliases"`
	}
	var got []entry
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\nraw=%s", err, buf.String())
	}
	if len(got) != len(ui.GameNames()) {
		t.Errorf("entry count = %d, want %d", len(got), len(ui.GameNames()))
	}
	// Every entry must have a non-empty name and one of the three canonical categories.
	for _, e := range got {
		if e.Name == "" {
			t.Errorf("entry has empty name: %+v", e)
		}
		if !validCategory(e.Category) {
			t.Errorf("entry %q has invalid category %q", e.Name, e.Category)
		}
		if e.Description == "" {
			t.Errorf("entry %q has empty description", e.Name)
		}
		// Aliases must be a non-nil slice so JSON renders [] not null.
		if e.Aliases == nil {
			t.Errorf("entry %q has nil aliases (want []); JSON shape must be stable", e.Name)
		}
	}
}

// TestPrintGamesJSONNullAliasesAvoided locks down the JSON shape requirement
// that aliases is always `[]`, never `null`. Scripts that do
// `.aliases | length` would crash on a null without this guarantee.
func TestPrintGamesJSONNullAliasesAvoided(t *testing.T) {
	var buf bytes.Buffer
	if err := printGamesJSON("", &buf); err != nil {
		t.Fatalf("printGamesJSON err: %v", err)
	}
	if strings.Contains(buf.String(), `"aliases":null`) {
		t.Errorf("JSON output contains \"aliases\":null — must always emit [] for stable schema; got:\n%s", buf.String())
	}
}

// TestPrintGamesJSONByCategory verifies that --category narrows the JSON
// output to the matching subset and that the result is consistent with
// games.ByCategory (the SSoT).
func TestPrintGamesJSONByCategory(t *testing.T) {
	for _, cat := range []string{"casino", "classic", "solo"} {
		t.Run(cat, func(t *testing.T) {
			var buf bytes.Buffer
			if err := printGamesJSON(cat, &buf); err != nil {
				t.Fatalf("printGamesJSON(%q) err: %v", cat, err)
			}
			var got []map[string]any
			if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
				t.Fatalf("invalid JSON: %v\nraw=%s", err, buf.String())
			}
			if len(got) == 0 {
				t.Fatalf("category %q produced 0 entries (every category should have games)", cat)
			}
			for _, e := range got {
				if e["category"] != cat {
					t.Errorf("entry %v has category %q, want %q", e["name"], e["category"], cat)
				}
			}
		})
	}
}

// TestPrintGamesByCategoryFiltersLong verifies the non-JSON long output
// honors --category.
func TestPrintGamesByCategoryFiltersLong(t *testing.T) {
	var buf bytes.Buffer
	printGames(false, false, "casino", &buf)
	out := buf.String()
	if !strings.Contains(out, "blackjack") {
		t.Errorf("expected casino filter to include blackjack; got:\n%s", out)
	}
	// hearts is classic; must be excluded.
	if strings.Contains(out, "hearts ") {
		t.Errorf("expected casino filter to exclude classic-category 'hearts'; got:\n%s", out)
	}
}

// TestPrintGamesByCategoryFiltersShort verifies short output honors --category.
func TestPrintGamesByCategoryFiltersShort(t *testing.T) {
	var buf bytes.Buffer
	printGames(true, false, "solo", &buf)
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) == 0 || lines[0] == "" {
		t.Fatalf("expected non-empty solo output; got:\n%s", buf.String())
	}
	// klondike is solo; must be present. blackjack is casino; must be absent.
	hasKlondike, hasBlackjack := false, false
	for _, l := range lines {
		switch l {
		case "klondike":
			hasKlondike = true
		case "blackjack":
			hasBlackjack = true
		}
	}
	if !hasKlondike {
		t.Errorf("solo filter should include klondike; got:\n%s", buf.String())
	}
	if hasBlackjack {
		t.Errorf("solo filter should exclude blackjack; got:\n%s", buf.String())
	}
}

// TestRunGamesInvalidCategoryExits2 verifies that --category with an invalid
// value rejects with exit 2 (POSIX usage error) and emits an i18n message
// naming the offending value on stderr.
func TestRunGamesInvalidCategoryExits2(t *testing.T) {
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
	flag.CommandLine = flag.NewFlagSet("trumpcards", flag.ExitOnError)
	os.Args = []string{"trumpcards", "games", "--category", "bogus"}

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

	if exit != 2 {
		t.Errorf("exit = %d, want 2 (stderr=%q)", exit, errBuf.String())
	}
	if !strings.Contains(errBuf.String(), "bogus") {
		t.Errorf("stderr must name the offending category 'bogus'; got: %q", errBuf.String())
	}
}

func TestCliAliasesWithoutShortKeyRemoved(t *testing.T) {
	// The `cliAliasesWithoutShort` warning contradicted the long-mode behavior
	// and was removed; ensure it hasn't crept back into either locale.
	if got := i18n.T("cliAliasesWithoutShort"); got != "cliAliasesWithoutShort" && got != "" {
		t.Errorf("i18n key 'cliAliasesWithoutShort' should be removed but still resolves to: %q", got)
	}
}

// TestValidCategoryNamesDerivedFromRegistry verifies the validCategory set is
// in lockstep with the games registry's actual categories. Adding a new
// games.Category in the registry must automatically extend the CLI filter
// without touching this file. See PR #1538 review feedback.
func TestValidCategoryNamesDerivedFromRegistry(t *testing.T) {
	// Walk every registered game's category via the same helper production
	// uses; assert the predicate accepts it.
	registryCategories := make(map[string]bool)
	for name, cat := range gameCategoryByName() {
		registryCategories[cat] = true
		if !validCategory(cat) {
			t.Errorf("registry has category %q (for game %q) but validCategory rejects it", cat, name)
		}
	}
	if len(registryCategories) == 0 {
		t.Fatal("expected at least one category in registry")
	}
	// And no extra entries that don't appear in the registry — that would
	// mean validCategoryNames drifted away from the SSoT.
	for cat := range validCategoryNames {
		if !registryCategories[cat] {
			t.Errorf("validCategory accepts %q but no game in the registry uses it", cat)
		}
	}
}

// TestRunGamesJSONIgnoredFlagsWarning verifies that passing --short or
// --aliases together with --json emits a one-line warning to stderr
// (the JSON schema is fixed; silently dropping flags hides script bugs).
// See PR #1538 review feedback (observation #4).
func TestRunGamesJSONIgnoredFlagsWarning(t *testing.T) {
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
		name     string
		args     []string
		wantWarn bool
	}{
		{"--json --short warns", []string{"trumpcards", "games", "--json", "--short"}, true},
		{"--json --aliases warns", []string{"trumpcards", "games", "--json", "--aliases"}, true},
		{"--json alone is silent", []string{"trumpcards", "games", "--json"}, false},
		{"--json --category solo is silent", []string{"trumpcards", "games", "--json", "--category", "solo"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet("trumpcards", flag.ExitOnError)
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

			if exit != 0 {
				t.Fatalf("exit = %d, want 0 (stderr=%q)", exit, errBuf.String())
			}
			// JSON must always end up on stdout regardless of warnings.
			if outBuf.Len() == 0 || outBuf.String()[0] != '[' {
				t.Errorf("expected a JSON array on stdout; got: %q", outBuf.String())
			}
			gotWarn := strings.Contains(errBuf.String(), "--json")
			if gotWarn != tc.wantWarn {
				t.Errorf("warning emitted: got=%v, want=%v (stderr=%q)", gotWarn, tc.wantWarn, errBuf.String())
			}
		})
	}
}

// TestUpdateExitCodeMapping verifies that every sentinel error from the
// updater maps to a distinct, documented exit code. Callers depend on this
// mapping for retry / fallback logic. See issue #1449.
func TestUpdateExitCodeMapping(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"non-interactive", update.ErrNonInteractive, 2},
		{"network failure", update.ErrNetwork, 3},
		{"api status", update.ErrAPIStatus, 3},
		{"no asset", update.ErrNoAsset, 4},
		{"extract failure", update.ErrExtract, 5},
		{"apply failure", update.ErrApply, 6},
		{"user cancelled", update.ErrUserCancelled, 75},
		{"update available", update.ErrUpdateAvailable, 10},
		{"unknown error", errors.New("surprise"), 1},
		// Wrapping preserves the category.
		{"wrapped network", fmt.Errorf("dial: %w", update.ErrNetwork), 3},
		{"wrapped apply", fmt.Errorf("chmod: %w", update.ErrApply), 6},
		{"wrapped update available", fmt.Errorf("check: %w", update.ErrUpdateAvailable), 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := updateExitCode(tt.err); got != tt.want {
				t.Errorf("updateExitCode(%v) = %d, want %d", tt.err, got, tt.want)
			}
		})
	}
}

// TestApplyTrailingGlobalFlags exercises issue #1509: --lang / --no-color
// must take effect even when they appear after the game name. The helper
// strips recognized flags from args, applies them to the runtime, and leaves
// genuinely unknown trailing tokens for the caller's extra-args warning.
func TestApplyTrailingGlobalFlags(t *testing.T) {
	// Restore globals after each subtest so the suite stays order-independent.
	origLang := i18n.Lang()
	origStdoutNo := color.NoColorStdout()
	origStderrNo := color.NoColorStderr()
	t.Cleanup(func() {
		i18n.SetLang(origLang)
		color.SetStdoutColor(!origStdoutNo)
		color.SetStderrColor(!origStderrNo)
	})

	tests := []struct {
		name        string
		args        []string
		quiet       bool
		wantRest    []string
		wantLang    string
		wantNoColor bool
		wantWarn    string // substring expected on stderr; empty = no output
	}{
		{
			name:     "no flags -> args passthrough",
			args:     []string{"foo", "bar"},
			wantRest: []string{"foo", "bar"},
			wantLang: "ja",
		},
		{
			name:     "--lang en sets locale and is consumed",
			args:     []string{"--lang", "en"},
			wantRest: []string{},
			wantLang: "en",
		},
		{
			name:     "-lang en (single-dash form) accepted",
			args:     []string{"-lang", "en"},
			wantRest: []string{},
			wantLang: "en",
		},
		{
			name:     "--lang=en (equals form) accepted",
			args:     []string{"--lang=en"},
			wantRest: []string{},
			wantLang: "en",
		},
		{
			name:        "--no-color disables both streams",
			args:        []string{"--no-color"},
			wantRest:    []string{},
			wantLang:    "ja",
			wantNoColor: true,
		},
		{
			name:     "unsupported --lang warns and falls back",
			args:     []string{"--lang", "klingon"},
			wantRest: []string{},
			wantLang: "ja",
			wantWarn: "klingon",
		},
		{
			name:     "unsupported --lang stays silent under quiet",
			args:     []string{"--lang", "klingon"},
			quiet:    true,
			wantRest: []string{},
			wantLang: "ja",
		},
		{
			name:        "mixed: known flags consumed, extras returned",
			args:        []string{"foo", "--lang", "en", "--no-color", "bar"},
			wantRest:    []string{"foo", "bar"},
			wantLang:    "en",
			wantNoColor: true,
		},
		{
			name:     "trailing --lang without value is ignored, not consumed",
			args:     []string{"--lang"},
			wantRest: []string{},
			wantLang: "ja",
		},
		{
			name:     "--lang= (equals form, empty value) silently consumed",
			args:     []string{"--lang="},
			wantRest: []string{},
			wantLang: "ja",
		},
		{
			name:        "--lang followed by flag token treats it as lang value",
			args:        []string{"--lang", "--no-color"},
			wantRest:    []string{},
			wantLang:    "ja",  // "--no-color" is not a supported lang, falls back
			wantNoColor: false, // --no-color was consumed as the lang value, not processed
			wantWarn:    "--no-color",
		},
		{
			name:        "--no-color=true disables both streams",
			args:        []string{"--no-color=true"},
			wantRest:    []string{},
			wantLang:    "ja",
			wantNoColor: true,
		},
		{
			name:        "--no-color=false is consumed without disabling color",
			args:        []string{"--no-color=false"},
			wantRest:    []string{},
			wantLang:    "ja",
			wantNoColor: false,
		},
		{
			name:        "-- stops flag scanning; remainder including -- passed through",
			args:        []string{"--lang", "en", "--", "--no-color"},
			wantRest:    []string{"--", "--no-color"},
			wantLang:    "en",
			wantNoColor: false,
		},
		// PR #1582 review: trailing -q/--quiet must be silently consumed
		// (no spurious "extra arguments ignored" warning) since the flag
		// is documented as global and was already evaluated upstream.
		{
			name:     "-q is silently consumed (no extra-args warning)",
			args:     []string{"-q"},
			wantRest: []string{},
			wantLang: "ja",
		},
		{
			name:     "--quiet is silently consumed",
			args:     []string{"--quiet"},
			wantRest: []string{},
			wantLang: "ja",
		},
		{
			name:     "--quiet=true (equals form) is silently consumed",
			args:     []string{"--quiet=true"},
			wantRest: []string{},
			wantLang: "ja",
		},
		{
			name:     "-q mixed with foo positional preserves foo",
			args:     []string{"foo", "-q", "bar"},
			wantRest: []string{"foo", "bar"},
			wantLang: "ja",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to a known baseline before each subtest.
			i18n.SetLang("ja")
			color.SetStdoutColor(true)
			color.SetStderrColor(true)

			var stderr bytes.Buffer
			got := applyTrailingGlobalFlags(tt.args, tt.quiet, &stderr)

			if !slices.Equal(got, tt.wantRest) {
				t.Errorf("rest = %#v, want %#v", got, tt.wantRest)
			}
			if i18n.Lang() != tt.wantLang {
				t.Errorf("lang = %q, want %q", i18n.Lang(), tt.wantLang)
			}
			if color.NoColorStdout() != tt.wantNoColor || color.NoColorStderr() != tt.wantNoColor {
				t.Errorf("no-color stdout/stderr = %v/%v, want both %v",
					color.NoColorStdout(), color.NoColorStderr(), tt.wantNoColor)
			}
			if tt.wantWarn == "" {
				if stderr.Len() != 0 {
					t.Errorf("expected no stderr; got %q", stderr.String())
				}
			} else if !strings.Contains(stderr.String(), tt.wantWarn) {
				t.Errorf("stderr missing %q: %q", tt.wantWarn, stderr.String())
			}
		})
	}
}
