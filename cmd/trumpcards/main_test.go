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
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games"
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

func TestValidHost(t *testing.T) {
	tests := []struct {
		host string
		want bool
	}{
		{"", true},                    // empty -> server falls back to its 127.0.0.1 default
		{"0.0.0.0", true},             // expose-all literal
		{"127.0.0.1", true},           // IPv4 loopback
		{"::1", true},                 // IPv6 loopback
		{"2001:db8::1", true},         // IPv6 literal
		{"localhost", true},           // hostname
		{"example.com", true},         // dotted hostname
		{"my-host.local", true},       // hyphen inside label
		{"nonexistent.invalid", true}, // syntactically valid; DNS failure stays exit 1 at bind
		{"bad host", false},           // embedded space
		{"a\tb", false},               // embedded tab
		{" leading", false},           // leading whitespace
		{"trailing ", false},          // trailing whitespace
		{"1.2.3.4:5", false},          // smuggled port (would corrupt JoinHostPort)
		{"-bad.com", false},           // label starts with hyphen
		{"bad-.com", false},           // label ends with hyphen
		{"a..b", false},               // empty label
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("host=%q", tt.host), func(t *testing.T) {
			if got := validHost(tt.host); got != tt.want {
				t.Errorf("validHost(%q) = %v, want %v", tt.host, got, tt.want)
			}
		})
	}
}

// TestRunWebRejectsInvalidHost is the --host analogue of
// TestRunWebRejectsExplicitInvalidPort (#2150): a syntactically malformed
// --host must exit 2 (usage error) with cliInvalidHost + the help hint and
// must NOT set the HOST env var, so systemd/CI can distinguish an operator
// typo from a runtime bind failure (exit 1).
func TestRunWebRejectsInvalidHost(t *testing.T) {
	origArgs := os.Args
	origCmdLine := flag.CommandLine
	origStdout := os.Stdout
	origStderr := os.Stderr
	origHostEnv, hostEnvWasSet := os.LookupEnv("HOST")
	defer func() {
		os.Args = origArgs
		flag.CommandLine = origCmdLine
		os.Stdout = origStdout
		os.Stderr = origStderr
		if hostEnvWasSet {
			_ = os.Setenv("HOST", origHostEnv)
		} else {
			_ = os.Unsetenv("HOST")
		}
	}()
	_ = os.Unsetenv("HOST") // baseline

	flag.CommandLine = flag.NewFlagSet("trumpcards", flag.ExitOnError)
	os.Args = []string{"trumpcards", "web", "--host", "bad host"}

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
		t.Errorf("exit = %d, want 2 (usage error; stderr=%q)", exit, errBuf.String())
	}
	if !strings.Contains(errBuf.String(), "bad host") {
		t.Errorf("stderr should mention the offending host; got: %q", errBuf.String())
	}
	if !strings.Contains(errBuf.String(), "help web") {
		t.Errorf("stderr should include cliTryHelp hint; got: %q", errBuf.String())
	}
	if got := os.Getenv("HOST"); got != "" {
		t.Errorf("HOST should not be set on invalid input; got %q", got)
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
// -1 must be rejected with cliInvalidPort and exit 2 (usage error, per the
// documented EXIT CODES in builtinSubcommandHelp["web"]) before any bind.
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

	if exit != 2 {
		t.Errorf("exit = %d, want 2 (usage error; stderr=%q)", exit, errBuf.String())
	}
	if !strings.Contains(errBuf.String(), "-1") {
		t.Errorf("stderr should mention the offending port; got: %q", errBuf.String())
	}
	if !strings.Contains(errBuf.String(), "help web") {
		t.Errorf("stderr should include cliTryHelp hint; got: %q", errBuf.String())
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

// TestRunUnknownGameExitsUsageError verifies issue #4305: an unknown top-level
// game name exits 2 (usage error) — matching the `--start <unknown>` path — and
// prints a Did-you-mean suggestion plus a one-line recovery hint on stderr,
// instead of the old exit-1-with-dead-flag.Usage()-blank-line behavior.
func TestRunUnknownGameExitsUsageError(t *testing.T) {
	origArgs := os.Args
	origCmdLine := flag.CommandLine
	origStdout := os.Stdout
	origStderr := os.Stderr
	defer func() {
		os.Args = origArgs
		flag.CommandLine = origCmdLine
		os.Stdout = origStdout
		os.Stderr = origStderr
		i18n.SetLang("ja")
	}()

	flag.CommandLine = flag.NewFlagSet("trumpcards", flag.ExitOnError)
	os.Args = []string{"trumpcards", "blackjak"} // typo for blackjack

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
		t.Errorf("exit = %d, want 2 (usage error); stderr=%q", exit, errBuf.String())
	}
	if !strings.Contains(errBuf.String(), "blackjack") {
		t.Errorf("expected Did-you-mean to mention blackjack; stderr=%q", errBuf.String())
	}
	// The recovery hint must point at `games` and `--help`.
	if !strings.Contains(errBuf.String(), "games") || !strings.Contains(errBuf.String(), "--help") {
		t.Errorf("expected recovery hint mentioning 'games' and '--help'; stderr=%q", errBuf.String())
	}
	// Unknown game is an error path: nothing goes to stdout.
	if outBuf.Len() != 0 {
		t.Errorf("expected no stdout on unknown game; got %q", outBuf.String())
	}
}

// TestParseSubFlagsToNoHelpDumpOnFlagError verifies issue #4307: an unknown
// subcommand flag must not dump the full help text to stdout before the
// localized error goes to stderr. The FlagSet's Usage callback fires inside
// fs.Parse on a parse error, so it must be a no-op (help is printed explicitly
// only on the flag.ErrHelp path).
func TestParseSubFlagsToNoHelpDumpOnFlagError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	fs, code, ok := parseSubFlagsTo("version", []string{"--bogus"}, func(fs *flag.FlagSet) {
		var short bool
		fs.BoolVar(&short, "short", false, "")
	}, &stdout, &stderr)

	if ok || fs != nil {
		t.Fatalf("expected (nil, code, false) on flag error; got ok=%v fs=%v", ok, fs)
	}
	if code != 2 {
		t.Errorf("exit = %d, want 2 (usage error)", code)
	}
	if stdout.Len() != 0 {
		t.Errorf("expected NO help dump on stdout for a flag error; got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "help version") {
		t.Errorf("stderr should carry the cliTryHelp hint; got %q", stderr.String())
	}
}

// TestParseSubFlagsToPrintsHelpOnceOnHelpFlag verifies that `-h` prints the
// subcommand help to stdout exactly once (not twice, as it did when Usage
// duplicated the explicit ErrHelp-branch print). See issue #4307.
func TestParseSubFlagsToPrintsHelpOnceOnHelpFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	fs, code, ok := parseSubFlagsTo("version", []string{"-h"}, func(fs *flag.FlagSet) {
		var short bool
		fs.BoolVar(&short, "short", false, "")
	}, &stdout, &stderr)

	if ok || fs != nil {
		t.Fatalf("expected (nil, 0, false) on -h; got ok=%v fs=%v", ok, fs)
	}
	if code != 0 {
		t.Errorf("exit = %d, want 0 for -h", code)
	}
	if n := strings.Count(stdout.String(), "USAGE:"); n != 1 {
		t.Errorf("expected help printed exactly once on stdout; USAGE: appeared %d times in %q", n, stdout.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("expected no stderr on -h; got %q", stderr.String())
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

// TestRunGlobalFlagsAcceptedAfterSubcommand verifies issue #4306: global flags
// (--lang / --color / --no-color) placed after a subcommand name are applied and
// stripped before the subcommand FlagSet parses, instead of being rejected with
// "flag provided but not defined" (exit 2). This makes trailing global flags
// uniform across all subcommands, matching the -q handling and the mental model
// the top-level --help advertises (`trumpcards poker --lang en`).
func TestRunGlobalFlagsAcceptedAfterSubcommand(t *testing.T) {
	origArgs := os.Args
	origCmdLine := flag.CommandLine
	origStdout := os.Stdout
	origStderr := os.Stderr
	defer func() {
		os.Args = origArgs
		flag.CommandLine = origCmdLine
		os.Stdout = origStdout
		os.Stderr = origStderr
		i18n.SetLang("ja")
	}()

	cases := []struct {
		name string
		args []string
	}{
		{"games --lang en --short", []string{"trumpcards", "games", "--lang", "en", "--short"}},
		{"games --short --lang en (trailing)", []string{"trumpcards", "games", "--short", "--lang", "en"}},
		{"games --lang=en --short (inline)", []string{"trumpcards", "games", "--lang=en", "--short"}},
		{"games --color never --short", []string{"trumpcards", "games", "--color", "never", "--short"}},
		{"games --no-color --short", []string{"trumpcards", "games", "--no-color", "--short"}},
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
				t.Errorf("exit = %d, want 0 (stderr=%q)", exit, errBuf.String())
			}
			if strings.Contains(errBuf.String(), "flag provided but not defined") {
				t.Errorf("subcommand must accept trailing global flags; got stderr=%q", errBuf.String())
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

// TestGameNamesAllHaveCategory guards the invariant that the category-grouped
// printGamesLong depends on: every ui.GameNames() entry must map to a non-empty
// games-registry category. printGamesLong iterates games.AllCategories() and
// prints each category's bucket, so a name whose category is "" (two-registry
// drift between ui.gameRegistry and games.registry) would land in the never-
// printed "" bucket and silently vanish from even the unfiltered `games`
// listing. This test makes such drift fail loudly instead. See PR #4320 review.
func TestGameNamesAllHaveCategory(t *testing.T) {
	byName := gameCategoryByName()
	for _, name := range ui.GameNames() {
		if byName[name] == "" {
			t.Errorf("game %q has no games-registry category — it would be dropped from `games` output (registry drift)", name)
		}
	}
}

// TestPrintGamesLongListsEveryGame is the direct backstop for the same drift:
// the unfiltered long listing must emit exactly one row per registered game, so
// a silently-dropped game shrinks the count and fails here.
func TestPrintGamesLongListsEveryGame(t *testing.T) {
	var buf bytes.Buffer
	printGames(false, false, "", &buf)
	rows := 0
	for _, line := range strings.Split(buf.String(), "\n") {
		if strings.HasPrefix(line, "  ") { // game rows are indented; headings are not
			rows++
		}
	}
	if want := len(ui.GameNames()); rows != want {
		t.Errorf("long listing emitted %d game rows, want %d", rows, want)
	}
}

// TestPrintGamesLongGroupsByCategory verifies issue #4311: the long-form list
// prints an uppercase "CATEGORY (N):" heading (derived from games.AllCategories,
// the SSoT) before each group.
func TestPrintGamesLongGroupsByCategory(t *testing.T) {
	var buf bytes.Buffer
	printGames(false, false, "", &buf)
	out := buf.String()
	for _, cat := range games.AllCategories() {
		heading := strings.ToUpper(cat.String()) + " ("
		if !strings.Contains(out, heading) {
			t.Errorf("expected category heading starting %q in long output; got:\n%s", heading, out)
		}
	}
}

// TestPrintGamesLongDynamicWidthAlignsDescriptions verifies issue #4311: the
// name column is sized to the longest displayed name so every description
// starts at the same byte offset — including the row for the longest name
// (ultimatetexasholdem), which the old fixed %-16s clipped out of alignment.
func TestPrintGamesLongDynamicWidthAlignsDescriptions(t *testing.T) {
	var buf bytes.Buffer
	printGames(false, false, "", &buf)
	descs := ui.GameDescriptions()

	width := 0
	for _, n := range ui.GameNames() {
		if len(n) > width {
			width = len(n)
		}
	}
	descStart := 2 + width + 1 // "  " + padded name + single separating space

	checked := 0
	for _, line := range strings.Split(buf.String(), "\n") {
		if !strings.HasPrefix(line, "  ") { // skip headings (no indent) and blanks
			continue
		}
		if len(line) < descStart {
			t.Errorf("line shorter than aligned description column %d: %q", descStart, line)
			continue
		}
		name := strings.TrimSpace(line[2 : descStart-1])
		desc := descs[name]
		if desc == "" {
			continue
		}
		if !strings.HasPrefix(line[descStart:], desc) {
			t.Errorf("description column misaligned for %q: expected desc at offset %d; line=%q", name, descStart, line)
		}
		checked++
	}
	if checked < 10 {
		t.Fatalf("expected to verify many game lines; only checked %d", checked)
	}
}

// TestJoinOr covers the English list renderer used for the --category prose.
func TestJoinOr(t *testing.T) {
	cases := []struct {
		in   []string
		want string
	}{
		{nil, ""},
		{[]string{"a"}, "a"},
		{[]string{"a", "b"}, "a or b"},
		{[]string{"a", "b", "c"}, "a, b, or c"},
		{[]string{"casino", "classic", "solo", "extra"}, "casino, classic, solo, or extra"},
	}
	for _, tc := range cases {
		if got := joinOr(tc.in); got != tc.want {
			t.Errorf("joinOr(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestGamesHelpListsAllCategories verifies issue #4308: the --category values
// advertised in the games subcommand help and the top-level help are derived
// from games.AllCategories() (the SSoT), so every registered category —
// including the `extra` bucket the old hardcoded strings omitted — is listed.
func TestGamesHelpListsAllCategories(t *testing.T) {
	for _, c := range games.AllCategories() {
		if !strings.Contains(categoryFilterPipe, c.String()) {
			t.Errorf("categoryFilterPipe %q missing category %q", categoryFilterPipe, c.String())
		}
		if !strings.Contains(categoryFilterProse, c.String()) {
			t.Errorf("categoryFilterProse %q missing category %q", categoryFilterProse, c.String())
		}
	}
	if !strings.Contains(categoryFilterPipe, "extra") {
		t.Errorf("expected 'extra' in categoryFilterPipe; got %q", categoryFilterPipe)
	}

	gamesHelp := strings.Join(builtinSubcommandHelp["games"], "\n")
	if !strings.Contains(gamesHelp, categoryFilterPipe) {
		t.Errorf("games help USAGE should list %q; got:\n%s", categoryFilterPipe, gamesHelp)
	}
	if !strings.Contains(gamesHelp, categoryFilterProse) {
		t.Errorf("games help --category desc should list %q; got:\n%s", categoryFilterProse, gamesHelp)
	}
	if help := buildHelpText(); !strings.Contains(help, categoryFilterPipe) {
		t.Errorf("top-level help should list the dynamic category list %q", categoryFilterPipe)
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
			q := tt.quiet
			got := applyTrailingGlobalFlags(tt.args, &q, &stderr)

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

// TestApplyTrailingGlobalFlags_TrailingQuietPropagates verifies issue #1840:
// a trailing -q / --quiet must update the caller's quiet flag (not silently
// no-op) so the rest of run() — most importantly the cliUnsupportedLang
// warning emitted inside applyTrailingGlobalFlags itself — actually respects
// it, matching --lang / --color which already apply in trailing position.
func TestApplyTrailingGlobalFlags_TrailingQuietPropagates(t *testing.T) {
	origLang := i18n.Lang()
	t.Cleanup(func() { i18n.SetLang(origLang) })

	cases := []struct {
		name     string
		args     []string
		startQ   bool
		wantQ    bool
		wantWarn string // substring expected on stderr; empty = none
		wantRest []string
	}{
		{
			name:     "-q sets quiet from false to true",
			args:     []string{"-q"},
			wantQ:    true,
			wantRest: []string{},
		},
		{
			name:     "--quiet sets quiet from false to true",
			args:     []string{"--quiet"},
			wantQ:    true,
			wantRest: []string{},
		},
		{
			name:     "--quiet=true sets quiet from false to true",
			args:     []string{"--quiet=true"},
			wantQ:    true,
			wantRest: []string{},
		},
		{
			name:     "--quiet=false unsets a pre-existing quiet",
			args:     []string{"--quiet=false"},
			startQ:   true,
			wantQ:    false,
			wantRest: []string{},
		},
		{
			name:     "trailing -q suppresses unsupported-lang warning",
			args:     []string{"-q", "--lang", "klingon"},
			wantQ:    true,
			wantWarn: "", // -q already true by the time the lang warning runs
			wantRest: []string{},
		},
		{
			// Order-independence: -q must suppress even when it appears
			// after the offending --lang. Resolved via pre-pass in
			// resolveTrailingQuiet; without that this case would still warn.
			name:     "reverse order: --lang xyz -q also suppresses",
			args:     []string{"--lang", "klingon", "-q"},
			wantQ:    true,
			wantWarn: "",
			wantRest: []string{},
		},
		{
			name:     "without -q the unsupported-lang warning fires",
			args:     []string{"--lang", "klingon"},
			wantQ:    false,
			wantWarn: "klingon",
			wantRest: []string{},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			i18n.SetLang("ja")
			var stderr bytes.Buffer
			q := tt.startQ
			rest := applyTrailingGlobalFlags(tt.args, &q, &stderr)
			if q != tt.wantQ {
				t.Errorf("quiet = %v, want %v", q, tt.wantQ)
			}
			if !slices.Equal(rest, tt.wantRest) {
				t.Errorf("rest = %#v, want %#v", rest, tt.wantRest)
			}
			if tt.wantWarn == "" {
				if stderr.Len() != 0 {
					t.Errorf("expected no stderr; got %q", stderr.String())
				}
			} else if !strings.Contains(stderr.String(), tt.wantWarn) {
				t.Errorf("stderr missing %q: got %q", tt.wantWarn, stderr.String())
			}
		})
	}
}

// TestApplyColorMode verifies issue #1554: the new tristate --color flag
// resolves to per-stream color settings with the documented precedence
// (NO_COLOR > --color=never / --no-color > --color=always > auto). The
// helper is the SSoT for color resolution; the run() integration relies
// on it for the explicit-flag path, so any drift here would silently
// change CLI behavior.
func TestApplyColorMode(t *testing.T) {
	// stdout-non-TTY/stderr-non-TTY: under `go test` both are pipes.
	const nonTTY = uintptr(0xDEADBEEF) // any value that IsTerminal returns false for
	tests := []struct {
		name        string
		mode        string
		noColorFlag bool
		noColorEnv  string
		wantStdout  bool
		wantStderr  bool
		wantOK      bool
		wantExit    int
		wantErrSub  string
	}{
		{
			name:       "auto + non-TTY -> off",
			mode:       "auto",
			wantStdout: false, wantStderr: false, wantOK: true,
		},
		{
			name:       "always -> force on",
			mode:       "always",
			wantStdout: true, wantStderr: true, wantOK: true,
		},
		{
			name:       "never -> force off",
			mode:       "never",
			wantStdout: false, wantStderr: false, wantOK: true,
		},
		{
			name:       "ALWAYS (case-insensitive) -> on",
			mode:       "ALWAYS",
			wantStdout: true, wantStderr: true, wantOK: true,
		},
		{
			name:       "  auto  (trim whitespace) -> off",
			mode:       "  auto  ",
			wantStdout: false, wantStderr: false, wantOK: true,
		},
		{
			name:        "--no-color overrides --color=always",
			mode:        "always",
			noColorFlag: true,
			wantStdout:  false, wantStderr: false, wantOK: true,
		},
		{
			name:       "NO_COLOR env beats --color=always (POSIX spec)",
			mode:       "always",
			noColorEnv: "1",
			wantStdout: false, wantStderr: false, wantOK: true,
		},
		{
			name:       "empty mode treated as auto",
			mode:       "",
			wantStdout: false, wantStderr: false, wantOK: true,
		},
		{
			name:       "invalid mode -> exit 2 with i18n error",
			mode:       "rainbow",
			wantOK:     false,
			wantExit:   2,
			wantErrSub: "rainbow",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Reset to a known baseline before each subtest.
			color.SetStdoutColor(true)
			color.SetStderrColor(true)
			var stderr bytes.Buffer
			code, ok := applyColorMode(tc.mode, tc.noColorFlag, tc.noColorEnv, nonTTY, nonTTY, &stderr)
			if ok != tc.wantOK {
				t.Errorf("ok=%v, want %v (stderr=%q)", ok, tc.wantOK, stderr.String())
			}
			if code != tc.wantExit {
				t.Errorf("exit=%d, want %d", code, tc.wantExit)
			}
			if tc.wantErrSub != "" && !strings.Contains(stderr.String(), tc.wantErrSub) {
				t.Errorf("stderr %q does not contain %q", stderr.String(), tc.wantErrSub)
			}
			// Only assert color state when the helper succeeded.
			if !ok {
				return
			}
			gotStdoutColor := !color.NoColorStdout()
			gotStderrColor := !color.NoColorStderr()
			if gotStdoutColor != tc.wantStdout {
				t.Errorf("stdout color = %v, want %v", gotStdoutColor, tc.wantStdout)
			}
			if gotStderrColor != tc.wantStderr {
				t.Errorf("stderr color = %v, want %v", gotStderrColor, tc.wantStderr)
			}
		})
	}
}

// TestApplyTrailingColorFlag verifies issue #1554: trailing `--color=...` (after
// the game name) is honored just like `--lang` and `--no-color`. An invalid
// trailing value is a soft warning rather than an exit-2, because the game has
// already been resolved and we don't want a typo to abort a launched session.
func TestApplyTrailingColorFlag(t *testing.T) {
	origNoColor, hadNoColor := os.LookupEnv("NO_COLOR")
	defer func() {
		if hadNoColor {
			_ = os.Setenv("NO_COLOR", origNoColor)
		} else {
			_ = os.Unsetenv("NO_COLOR")
		}
	}()
	_ = os.Unsetenv("NO_COLOR")

	cases := []struct {
		name       string
		args       []string
		quiet      bool
		noColorEnv string // NO_COLOR for this case; "" leaves unset
		wantStdout bool
		wantStderr bool
		wantWarn   string
	}{
		{
			name:       "--color=never trailing flag",
			args:       []string{"--color=never"},
			wantStdout: false, wantStderr: false,
		},
		{
			name:       "--color always trailing (space-separated)",
			args:       []string{"--color", "always"},
			wantStdout: true, wantStderr: true,
		},
		{
			name:       "invalid value warns (loud) but does not abort",
			args:       []string{"--color=rainbow"},
			wantStdout: true, wantStderr: true, // unchanged
			wantWarn: "rainbow",
		},
		{
			name:       "invalid value silenced under quiet",
			args:       []string{"--color=rainbow"},
			quiet:      true,
			wantStdout: true, wantStderr: true,
		},
		// PR #1583 review #3: missing edge cases for trailing --color.
		{
			// Case-insensitive matching must work in trailing position too,
			// matching applyColorMode's `ALWAYS` test. See review issue #3.
			name:       "--color=NEVER (case-insensitive)",
			args:       []string{"--color=NEVER"},
			wantStdout: false, wantStderr: false,
		},
		{
			// NO_COLOR env precedence applies in trailing position too —
			// users running `trumpcards <game> --color=auto` with NO_COLOR
			// set must still get color off (POSIX no-color spec).
			name:       "--color=auto with NO_COLOR set forces off",
			args:       []string{"--color=auto"},
			noColorEnv: "1",
			wantStdout: false, wantStderr: false,
		},
		{
			// NO_COLOR env beats --color=always even in trailing position
			// (matches the top-level applyColorMode precedence; PR #1583
			// review #2 — single SSoT for resolution).
			name:       "--color=always with NO_COLOR set forces off",
			args:       []string{"--color=always"},
			noColorEnv: "1",
			wantStdout: false, wantStderr: false,
		},
		{
			// Within trailing args, --no-color must beat --color=always
			// regardless of token order, matching top-level precedence.
			// Pre-fix this returned color ON because the second flag won.
			// See PR #1583 review #2.
			name:       "--no-color beats --color=always regardless of order",
			args:       []string{"--color=always", "--no-color"},
			wantStdout: false, wantStderr: false,
		},
		{
			// Reverse order still gives the same answer.
			name:       "--no-color beats --color=always (reverse order)",
			args:       []string{"--no-color", "--color=always"},
			wantStdout: false, wantStderr: false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			color.SetStdoutColor(true)
			color.SetStderrColor(true)
			if tt.noColorEnv == "" {
				_ = os.Unsetenv("NO_COLOR")
			} else {
				_ = os.Setenv("NO_COLOR", tt.noColorEnv)
			}
			t.Cleanup(func() { _ = os.Unsetenv("NO_COLOR") })

			var stderr bytes.Buffer
			q := tt.quiet
			rest := applyTrailingGlobalFlags(tt.args, &q, &stderr)
			if len(rest) != 0 {
				t.Errorf("trailing color flag should be consumed; got rest=%v", rest)
			}
			gotStdoutColor := !color.NoColorStdout()
			gotStderrColor := !color.NoColorStderr()
			if gotStdoutColor != tt.wantStdout || gotStderrColor != tt.wantStderr {
				t.Errorf("stdout/stderr color = %v/%v, want %v/%v",
					gotStdoutColor, gotStderrColor, tt.wantStdout, tt.wantStderr)
			}
			if tt.wantWarn == "" {
				if stderr.Len() != 0 {
					t.Errorf("expected no warning; got %q", stderr.String())
				}
			} else if !strings.Contains(stderr.String(), tt.wantWarn) {
				t.Errorf("stderr missing %q: got %q", tt.wantWarn, stderr.String())
			}
		})
	}
}

// TestRunVersionSubcommand verifies issue #1552: `trumpcards version` is
// equivalent to the `--version` flag, and `trumpcards version --short` is
// equivalent to `--version-short`. The subcommand form matters because it
// matches the conventions of `git`, `gh`, `kubectl`, etc.; users who type
// the subcommand by habit must not see an "unknown game" error.
func TestRunVersionSubcommand(t *testing.T) {
	cases := []struct {
		name      string
		args      []string
		wantExit  int
		wantSubst string
	}{
		{
			name:      "long form prints full version line",
			args:      []string{"trumpcards", "version"},
			wantExit:  0,
			wantSubst: "trumpcards ",
		},
		{
			name:      "--short prints just the version token",
			args:      []string{"trumpcards", "version", "--short"},
			wantExit:  0,
			wantSubst: version,
		},
	}
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

			if exit != tc.wantExit {
				t.Errorf("exit = %d, want %d (stderr=%q)", exit, tc.wantExit, errBuf.String())
			}
			if !strings.Contains(outBuf.String(), tc.wantSubst) {
				t.Errorf("stdout missing %q; got %q", tc.wantSubst, outBuf.String())
			}
			if errBuf.Len() != 0 {
				t.Errorf("expected no stderr output; got %q", errBuf.String())
			}
		})
	}
}

// TestSuggestionCandidatesIncludesAliases verifies issue #1555: the
// "did you mean" candidate set must include game aliases so a typo of
// the alias (`gni` -> `gin`) is recovered to the alias the user knows
// about, rather than to a far-off canonical name (`gofish`).
func TestSuggestionCandidatesIncludesAliases(t *testing.T) {
	commands := map[string]func() int{
		"ginrummy":      func() int { return 0 },
		"sevencardstud": func() int { return 0 },
		"gofish":        func() int { return 0 },
	}
	got := suggestionCandidates(commands)
	gotSet := make(map[string]struct{}, len(got))
	for _, n := range got {
		gotSet[n] = struct{}{}
	}
	for _, want := range []string{"ginrummy", "sevencardstud", "gofish"} {
		if _, ok := gotSet[want]; !ok {
			t.Errorf("missing canonical %q from candidates: %v", want, got)
		}
	}
	// At least one alias must appear; pick a known one if it exists.
	if _, ok := ui.GameAliases["gin"]; ok {
		if _, present := gotSet["gin"]; !present {
			t.Errorf("alias 'gin' should be in candidates: %v", got)
		}
	}
	// Spot-check dedup: if an alias collides with a canonical (it shouldn't,
	// but the helper must still be idempotent), the slice must not contain
	// the same string twice. We assert the invariant via len(map)==len(slice).
	if len(gotSet) != len(got) {
		t.Errorf("candidates contain duplicates: %v", got)
	}
}

// TestHelpSuggestionCandidatesIncludesAliases verifies issue #1555: the
// runHelpCommand suggestion path also pulls aliases, so `trumpcards help
// gni` recovers to `gin` (or `ginrummy`) rather than a distant canonical.
func TestHelpSuggestionCandidatesIncludesAliases(t *testing.T) {
	got := helpSuggestionCandidates()
	gotSet := make(map[string]struct{}, len(got))
	for _, n := range got {
		gotSet[n] = struct{}{}
	}
	if _, ok := ui.GameAliases["gin"]; ok {
		if _, present := gotSet["gin"]; !present {
			t.Errorf("alias 'gin' should be in help candidates: %v", got)
		}
	}
	// Builtin subcommands should also be suggestable for `trumpcards help <cmd>` typos.
	for _, want := range []string{"web", "update", "version"} {
		if _, ok := gotSet[want]; !ok {
			t.Errorf("missing builtin %q from help candidates: %v", want, got)
		}
	}
}

// TestBuildHelpTextHasExitCodes verifies issue #1556: the top-level --help
// must document the exit-code policy so CI / cron / scripts can branch on
// it without reading the source.
func TestBuildHelpTextHasExitCodes(t *testing.T) {
	helpText := buildHelpText()
	for _, want := range []string{"EXIT CODES:", " 130 ", " 143 ", " 10 ", " 75 "} {
		if !strings.Contains(helpText, want) {
			t.Errorf("help text missing %q; got:\n%s", want, helpText)
		}
	}
}

// TestBuildHelpTextMentionsVersionSubcommand verifies issue #1552: the
// COMMANDS section advertises the subcommand alongside the existing
// flags so it is discoverable from `trumpcards --help`.
func TestBuildHelpTextMentionsVersionSubcommand(t *testing.T) {
	helpText := buildHelpText()
	if !strings.Contains(helpText, "version") {
		t.Errorf("help text should advertise the 'version' subcommand; got:\n%s", helpText)
	}
}

// TestBuildHelpTextDocumentsColorTristate verifies issue #1554: the
// top-level --help advertises --color=auto|always|never so the new option
// is discoverable without reading the source.
func TestBuildHelpTextDocumentsColorTristate(t *testing.T) {
	helpText := buildHelpText()
	for _, want := range []string{"--color", "auto", "always", "never", "DEPRECATED"} {
		if !strings.Contains(helpText, want) {
			t.Errorf("help text missing %q; got:\n%s", want, helpText)
		}
	}
}

// TestResolveStartGame covers issue #1604: `--start <game>` selects the
// initial game for interactive mode. Empty string falls back to blackjack;
// aliases resolve to canonical names; unknown names exit 2.
func TestResolveStartGame(t *testing.T) {
	tests := []struct {
		name     string
		flag     string
		wantGame string
		wantCode int
		wantOk   bool
	}{
		{"empty -> blackjack default", "", "blackjack", 0, true},
		{"canonical name", "poker", "poker", 0, true},
		{"alias resolves", "gin", "ginrummy", 0, true},
		{"alias 7stud resolves", "7stud", "sevencardstud", 0, true},
		{"case insensitive", "PoKeR", "poker", 0, true},
		{"surrounding whitespace tolerated", "  poker  ", "poker", 0, true},
		{"unknown -> exit 2", "nosuchgame", "", 2, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stderr bytes.Buffer
			got, code, ok := resolveStartGame(tt.flag, &stderr)
			if got != tt.wantGame || code != tt.wantCode || ok != tt.wantOk {
				t.Errorf("resolveStartGame(%q) = (%q, %d, %v), want (%q, %d, %v)",
					tt.flag, got, code, ok, tt.wantGame, tt.wantCode, tt.wantOk)
			}
		})
	}
}

// TestResolveStartGame_UnknownEmitsDidYouMean verifies the typo-recovery path:
// "blackjac" → close enough to "blackjack" that we surface the suggestion.
func TestResolveStartGame_UnknownEmitsDidYouMean(t *testing.T) {
	var stderr bytes.Buffer
	_, code, ok := resolveStartGame("blackjac", &stderr)
	if ok {
		t.Fatalf("expected ok=false for unknown game, got ok=true")
	}
	if code != 2 {
		t.Errorf("expected exit 2 for unknown --start, got %d", code)
	}
	if !strings.Contains(stderr.String(), "blackjack") {
		t.Errorf("expected Did-you-mean to mention blackjack; stderr=%q", stderr.String())
	}
	// The --start path prints the same recovery hint as the positional path.
	if !strings.Contains(stderr.String(), "games") || !strings.Contains(stderr.String(), "--help") {
		t.Errorf("expected recovery hint mentioning 'games' and '--help'; stderr=%q", stderr.String())
	}
}

// TestBuildHelpTextDocumentsStartFlag verifies the --start flag is advertised
// in the top-level help so users discover it without reading the source.
// See issue #1604.
func TestBuildHelpTextDocumentsStartFlag(t *testing.T) {
	helpText := buildHelpText()
	for _, want := range []string{"--start", "Initial game", "--start poker"} {
		if !strings.Contains(helpText, want) {
			t.Errorf("help text missing %q; got:\n%s", want, helpText)
		}
	}
}

// TestBuildHelpTextGamesSummaryStaysCompact verifies issue #1694: the
// top-level --help GAMES section must NOT enumerate every available game
// (which used to push COMMANDS / OPTIONS off a 24-line terminal).
// Instead it shows a category-grouped summary and points at
// `trumpcards games` for the full list.
func TestBuildHelpTextGamesSummaryStaysCompact(t *testing.T) {
	helpText := buildHelpText()

	// The category summary must appear, with a pointer to `games`.
	for _, want := range []string{"GAMES:", "trumpcards games", "casino", "classic", "solo"} {
		if !strings.Contains(helpText, want) {
			t.Errorf("help text missing %q; got:\n%s", want, helpText)
		}
	}

	// The COMMANDS section must reach the top of the screen on a 24-line
	// terminal — this is the regression from #1694. With 6 USAGE lines and
	// a 5-line GAMES summary we expect COMMANDS to land within the first 20
	// lines (the issue had it at line ~84).
	lines := strings.Split(helpText, "\n")
	commandsIdx := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "COMMANDS:") {
			commandsIdx = i
			break
		}
	}
	if commandsIdx < 0 {
		t.Fatalf("COMMANDS: header not found in help text")
	}
	// commandsIdx is 0-based; index 20 = line 21, which would violate the
	// stated "first 20 lines" bound. Use >= so the guard is strict.
	if commandsIdx >= 20 {
		t.Errorf("COMMANDS: appears at line %d; expected <= 20 so it stays visible on a default terminal", commandsIdx+1)
	}
}
