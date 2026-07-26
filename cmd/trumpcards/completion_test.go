package main

import (
	"bytes"
	"flag"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/ui"
)

// runCLI drives the real run() entry point with the given argv (the program
// name is prepended) and returns what reached stdout, stderr, and the exit
// code. Both pipes are drained concurrently: the completion scripts are several
// KB, which would deadlock a read-after-close approach once the output exceeds
// the pipe buffer.
func runCLI(t *testing.T, args ...string) (stdout, stderr string, exit int) {
	t.Helper()
	origArgs, origCmdLine := os.Args, flag.CommandLine
	origStdout, origStderr := os.Stdout, os.Stderr
	t.Cleanup(func() {
		os.Args, flag.CommandLine = origArgs, origCmdLine
		os.Stdout, os.Stderr = origStdout, origStderr
	})

	flag.CommandLine = flag.NewFlagSet("trumpcards", flag.ExitOnError)
	os.Args = append([]string{"trumpcards"}, args...)

	rOut, wOut, err := os.Pipe()
	require.NoError(t, err)
	rErr, wErr, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout, os.Stderr = wOut, wErr

	outCh, errCh := make(chan string, 1), make(chan string, 1)
	drain := func(r *os.File, ch chan<- string) {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		ch <- buf.String()
	}
	go drain(rOut, outCh)
	go drain(rErr, errCh)

	exit = run()
	require.NoError(t, wOut.Close())
	require.NoError(t, wErr.Close())
	return <-outCh, <-errCh, exit
}

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
	// Issue #1693: flags added since the last completion update must be
	// reachable via Tab completion. Each is asserted in its applicable
	// context (--start completes a game, web has --open / --quiet,
	// update has --check / --dry-run, version subcommand exists).
	assert.Contains(t, script, "--start", "bash completion missing global --start flag")
	assert.Contains(t, script, "--quiet", "bash completion missing global --quiet flag")
	assert.Contains(t, script, "--open", "bash completion missing web --open flag")
	assert.Contains(t, script, "--check", "bash completion missing update --check flag")
	assert.Contains(t, script, "--dry-run", "bash completion missing update --dry-run flag")
	// Issue #4308: --category value completion must list every registered
	// category (including the `extra` bucket the old hardcoded list omitted).
	assert.Contains(t, script, `compgen -W "`+strings.Join(categoryDisplayNames(), " ")+`"`,
		"bash --category completion must list all registry categories")
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
	// Issue #1693: flags added since the last completion update must be
	// reachable via Tab completion in zsh as well.
	assert.Contains(t, script, "--start", "zsh completion missing global --start flag")
	assert.Contains(t, script, "--quiet", "zsh completion missing global --quiet flag")
	assert.Contains(t, script, "--open", "zsh completion missing web --open flag")
	assert.Contains(t, script, "--check", "zsh completion missing update --check flag")
	assert.Contains(t, script, "--dry-run", "zsh completion missing update --dry-run flag")
	// Issue #4308: --category value completion must list every registered category.
	assert.Contains(t, script, `:category:(`+strings.Join(categoryDisplayNames(), " ")+`)`,
		"zsh --category completion must list all registry categories")
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
	// Issue #1693: flags added since the last completion update must be
	// reachable via Tab completion in fish as well.
	assert.Contains(t, script, "-l start", "fish completion missing global --start flag")
	assert.Contains(t, script, "-l quiet", "fish completion missing global --quiet flag")
	assert.Contains(t, script, "-l open", "fish completion missing web --open flag")
	assert.Contains(t, script, "-l check", "fish completion missing update --check flag")
	assert.Contains(t, script, "-l dry-run", "fish completion missing update --dry-run flag")
	// Issue #4308: --category value completion must list every registered category.
	assert.Contains(t, script, `-l category -x -a '`+strings.Join(categoryDisplayNames(), " ")+`'`,
		"fish --category completion must list all registry categories")
}

// TestCompletionGameTargets_ExcludesSubcommands guards the contract that
// completionGameTargets() returns only game names and their aliases — never
// non-game subcommands. This is what makes `--start` and `help <target>`
// safe to back with the helper instead of completionSubcommands(): a future
// refactor that accidentally rolls subcommands into this list would silently
// let `trumpcards --start web` through and pin a regression.
func TestCompletionGameTargets_ExcludesSubcommands(t *testing.T) {
	targets := completionGameTargets()
	set := make(map[string]struct{}, len(targets))
	for _, n := range targets {
		set[n] = struct{}{}
	}
	for _, sub := range []string{"web", "completion", "games", "update", "version", "help"} {
		if _, found := set[sub]; found {
			t.Errorf("completionGameTargets() must not include subcommand %q", sub)
		}
	}
	// Sanity-check that real games still come through.
	if _, ok := set["blackjack"]; !ok {
		t.Errorf("completionGameTargets() should include 'blackjack'")
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

// Usage errors (no arg, extra args, unsupported shell) must exit 2 to match
// the documented EXIT CODES table and the rest of the CLI. Genuine I/O
// failures while writing the script still exit 1. See issue #1603.
func TestRunCompletion_NoArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runCompletionTo(nil, &stdout, &stderr, false, false)
	assert.Equal(t, 2, code)
	assert.Contains(t, stderr.String(), "trumpcards completion")
	// Issue #1838: missing-arg should also surface the same `help <cmd>`
	// hint that parseSubFlagsTo emits for unknown flags, so the UX is
	// symmetric across all subcommand error paths.
	assert.Contains(t, stderr.String(), "help completion")
}

func TestRunCompletion_ExtraArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runCompletionTo([]string{"bash", "extra"}, &stdout, &stderr, false, false)
	assert.Equal(t, 2, code)
}

// TestCompletionShellArgNotReportedAsIgnored guards the documented, canonical
// invocation `trumpcards completion <shell>`: the shell name is a required
// argument that runCompletionTo consumes, so the generic leftover-args warning
// must not claim it was ignored. Users source this from .bashrc, which made the
// bogus warning appear on every shell startup. See issue #4370.
func TestCompletionShellArgNotReportedAsIgnored(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish"} {
		t.Run(shell, func(t *testing.T) {
			stdout, stderr, exit := runCLI(t, "completion", shell)
			assert.Equal(t, 0, exit)
			assert.NotEmpty(t, stdout, "the completion script must still reach stdout")
			assert.Empty(t, stderr, "a valid invocation must not warn on stderr")
		})
	}
}

func TestRunCompletion_UnsupportedShell(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runCompletionTo([]string{"powershell"}, &stdout, &stderr, false, false)
	assert.Equal(t, 2, code)
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
			code := runCompletionTo([]string{tt.input}, &stdout, &stderr, false, false)
			assert.Equal(t, 2, code)
			assert.Contains(t, stderr.String(), tt.want, "expected suggestion %q for input %q", tt.want, tt.input)
		})
	}
}

func TestRunCompletion_UnsupportedShell_NoSuggestion(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runCompletionTo([]string{"powershell"}, &stdout, &stderr, false, false)
	assert.Equal(t, 2, code)
	// "powershell" is too far from any supported shell — no Did-you-mean line.
	assert.NotContains(t, stderr.String(), "Did you mean")
	assert.NotContains(t, stderr.String(), "もしかして")
}

func TestRunCompletion_ValidShells(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish"} {
		t.Run(shell, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			// Non-TTY default: no hint emitted, only the script body.
			code := runCompletionTo([]string{shell}, &stdout, &stderr, false, false)
			assert.Equal(t, 0, code)
			assert.NotEmpty(t, stdout.String())
			assert.Empty(t, stderr.String())
		})
	}
}

// TestRunCompletion_HintEmittedOnlyForTTY verifies issue #1450: the install
// hint pollutes redirected output (`> file`) but is useful on interactive
// terminals. The hint is gated on stdout being a TTY.
func TestRunCompletion_HintEmittedOnlyForTTY(t *testing.T) {
	tests := []struct {
		name        string
		stdoutIsTTY bool
		noHint      bool
		wantHint    bool
	}{
		{"non-TTY, hint not requested -> no hint (redirect path)", false, false, false},
		{"TTY, hint not suppressed -> hint shown (onboarding)", true, false, true},
		{"TTY but --no-hint -> no hint (explicit override)", true, true, false},
		{"non-TTY + --no-hint -> no hint (consistent)", false, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, shell := range []string{"bash", "zsh", "fish"} {
				var stdout, stderr bytes.Buffer
				code := runCompletionTo([]string{shell}, &stdout, &stderr, tt.stdoutIsTTY, tt.noHint)
				assert.Equal(t, 0, code, shell)
				body := stdout.String()
				// The completion script itself must always be present.
				assert.Contains(t, body, "trumpcards", "shell=%s body missing script", shell)
				// The install hint always starts with "# " and references "source" or "fpath".
				hasHint := strings.Contains(body, "# To load completions")
				if hasHint != tt.wantHint {
					t.Errorf("shell=%s hint emitted=%v, want=%v\nbody=%s", shell, hasHint, tt.wantHint, body)
				}
			}
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

func TestRunHelpCommand_ExtraArgs_Warns(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runHelpCommand([]string{"blackjack", "foo", "bar"}, "", &stdout, &stderr)
	assert.Equal(t, 0, code)
	assert.NotEmpty(t, stdout.String(), "should still print game help")
	assert.Contains(t, stderr.String(), "foo bar", "should warn about extra args")
}

func TestRunHelpCommand_BuiltinSubcommand(t *testing.T) {
	for _, name := range []string{"web", "completion", "games", "update", "help"} {
		t.Run(name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := runHelpCommand([]string{name}, "", &stdout, &stderr)
			assert.Equal(t, 0, code)
			// The USAGE line names the command regardless of the active locale.
			assert.Contains(t, stdout.String(), "trumpcards "+name)
			// Should NOT use the misleading "unknown game" wording.
			assert.NotContains(t, stderr.String(), "不明なゲーム")
			assert.NotContains(t, stderr.String(), "Unknown game")
			assert.NotContains(t, stderr.String(), "not a game")
		})
	}
}

func TestParseSubFlagsTo_HelpFlag_WritesHelpToStdoutAndExitsZero(t *testing.T) {
	var stdout, stderr bytes.Buffer
	_, code, ok := parseSubFlagsTo("web", []string{"--help"}, func(fs *flag.FlagSet) {
		fs.String("port", "", "port")
	}, &stdout, &stderr, false)
	assert.False(t, ok)
	assert.Equal(t, 0, code)
	// The USAGE line names the command regardless of the active locale.
	assert.Contains(t, stdout.String(), "trumpcards web")
	assert.Empty(t, stderr.String())
}

func TestParseSubFlagsTo_UnknownFlag_WritesI18nErrorAndExit2(t *testing.T) {
	var stdout, stderr bytes.Buffer
	_, code, ok := parseSubFlagsTo("web", []string{"--nope"}, func(fs *flag.FlagSet) {
		fs.String("port", "", "port")
	}, &stdout, &stderr, false)
	assert.False(t, ok)
	assert.Equal(t, 2, code)
	// Must not leak Go's raw "flag provided but not defined" line to either stream
	// without being wrapped in our i18n envelope.
	assert.NotContains(t, stdout.String(), "flag provided but not defined")
	assert.Contains(t, stderr.String(), "web")
	assert.Contains(t, stderr.String(), "help web")
}

// The leftover-args warning is right for the five subcommands that take no
// positional arguments and wrong for the one that does, so it is gated on
// takesPositional rather than on fs.NArg() alone. See issue #4370.
func TestParseSubFlagsTo_LeftoverArgs_WarnsOnlyWhenNotExpected(t *testing.T) {
	for _, tc := range []struct {
		name            string
		takesPositional bool
		wantWarning     bool
	}{
		{"warns when the subcommand takes no positional args", false, true},
		{"stays quiet when the subcommand consumes them", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			fs, code, ok := parseSubFlagsTo("completion", []string{"bash"}, func(fs *flag.FlagSet) {
				fs.Bool("no-hint", false, "no-hint")
			}, &stdout, &stderr, tc.takesPositional)
			require.True(t, ok)
			assert.Equal(t, 0, code)
			// Either way the argument is handed to the caller, never swallowed.
			assert.Equal(t, []string{"bash"}, fs.Args())
			if tc.wantWarning {
				assert.NotEmpty(t, stderr.String())
			} else {
				assert.Empty(t, stderr.String())
			}
		})
	}
}

func TestBuiltinSubcommandHelp_CoversAllNonGameSubcommands(t *testing.T) {
	for _, cmd := range builtinSubcommandNames {
		lines, ok := subcommandHelp(cmd)
		assert.True(t, ok, "subcommandHelp missing entry for %q", cmd)
		assert.NotEmpty(t, lines, "help lines for %q must not be empty", cmd)
		joined := strings.Join(lines, "\n")
		// The USAGE line names the command regardless of the active locale.
		assert.Contains(t, joined, "trumpcards "+cmd, "%q help must show its usage", cmd)
	}
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
