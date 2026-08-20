package ui

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// signalExitCode prints the goodbye banner, runs cleanup, and returns the
// POSIX exit code (128 + signal number) for sig.
//
// os.Exit does NOT run deferred functions, so a signal handler that simply
// calls os.Exit would skip the deferred reader.Close() in the CUI loops —
// losing the command history (never written to ~/.trumpcards_history) and,
// worse, leaving the terminal stuck in liner's raw mode (requiring `reset` /
// `stty sane` to recover). To avoid that, the signal watcher runs `cleanup`
// (which performs exactly what the deferred teardown would: history save +
// terminal restore) before exiting. Extracted as a pure function so the
// exit-code mapping and the cleanup invocation are unit-testable without
// raising a real signal or calling os.Exit. See issue #2096.
func signalExitCode(sig os.Signal, cleanup func()) int {
	fmt.Println("\n" + i18n.T("bye"))
	if cleanup != nil {
		cleanup()
	}
	// POSIX convention: exit code = 128 + signal number (SIGINT=130, SIGTERM=143).
	exitCode := 128
	if sigNum, ok := sig.(syscall.Signal); ok {
		exitCode += int(sigNum)
	}
	return exitCode
}

// runSignalWatcher installs a SIGINT/SIGTERM handler scoped to a single CUI
// loop and returns a stop function the loop must defer. On a signal it runs
// cleanup (see signalExitCode) and exits; on normal loop return the deferred
// stop() tears the watcher down without touching cleanup, so the loop's own
// deferred teardown runs exactly once. cleanup may be nil for loops with no
// resources to release (e.g. the plain-stdin Doubt loop).
//
// This replaces the previous package-level os.Exit handler, which bypassed
// deferred terminal restore and history persistence (issue #2096) — mirroring
// the local-signal approach already used by RunRealtimeCuiLoop.
func runSignalWatcher(cleanup func()) (stop func()) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	done := make(chan struct{})
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		select {
		case sig := <-sigCh:
			signal.Stop(sigCh)
			os.Exit(signalExitCode(sig, cleanup))
		case <-done:
			signal.Stop(sigCh)
		}
	}()
	return func() {
		close(done)
		<-stopped
	}
}

// CuiExecer CUIコントローラの共通インタフェース
type CuiExecer interface {
	Exec(command string) string
}

// commandLister exposes the candidate sets the readline tab-completer can
// offer. Implemented today by *GameManager; the interface lives here in the
// infra layer so future controllers can opt in without dragging the ui
// package into adapter/controller dependencies.
type commandLister interface {
	// CompletionCandidates returns commands that are valid as the first
	// token on a line (i.e., callable directly). Bare game names are NOT
	// valid first tokens — they only work as `switch <name>` — so
	// implementations must omit them here and surface them via
	// ArgumentCandidates("switch") instead.
	CompletionCandidates() []string

	// ArgumentCandidates returns valid completions for a token typed after
	// cmd (e.g. game names for cmd="switch"). Return nil for commands
	// without argument completion.
	ArgumentCandidates(cmd string) []string
}

// commonCompletionCommands is the always-available first-token command set.
// Mirrors cui_controller_helper.go's commonCommands; duplicated here because
// that list lives in the adapter layer and we want infra to stay independent.
var commonCompletionCommands = []string{
	"q", "quit", "exit",
	"r", "reset",
	"help", "?",
}

// installCompleter wires a tab-completion function into r. Two completion
// modes:
//
//   - First token: union of commonCompletionCommands and the lister's
//     standalone-runnable commands. Bare game names are intentionally
//     excluded — completing `bla<Tab>` to `blackjack` would mislead the
//     user, since `blackjack<Enter>` isn't a valid command (it has to be
//     `switch blackjack`).
//   - Second token after a recognised command (e.g. `switch <Tab>`):
//     forwarded to lister.ArgumentCandidates(cmd). Returns nil for
//     commands without argument completion (e.g. `b 100` — bet amounts
//     are game-state-dependent and can't be enumerated).
func installCompleter(r LineReader, lister commandLister) {
	r.SetCompleter(func(prefix string) []string {
		fields := strings.Fields(prefix)
		endsWithSpace := strings.HasSuffix(prefix, " ")

		// Past the second token: argful commands (e.g. "b 100") have no
		// general completion contract — game-state-dependent valid args
		// aren't enumerable here.
		if len(fields) > 2 || (len(fields) == 2 && endsWithSpace) {
			return nil
		}

		// On the second token (cursor either after "cmd " or in the middle
		// of "cmd arg"): consult the lister for argument completion.
		if len(fields) == 2 || (len(fields) == 1 && endsWithSpace) {
			if lister == nil {
				return nil
			}
			cmd := fields[0]
			argToken := ""
			if len(fields) == 2 {
				argToken = fields[1]
			}
			return filterByPrefix(lister.ArgumentCandidates(cmd), argToken)
		}

		// First token (or empty input): union of common + lister candidates.
		// Start with a fresh slice so append never mutates the package-level
		// commonCompletionCommands (its capacity is implementation-defined).
		candidates := append([]string(nil), commonCompletionCommands...)
		if lister != nil {
			candidates = append(candidates, lister.CompletionCandidates()...)
		}
		token := ""
		if len(fields) == 1 {
			token = fields[0]
		}
		return filterByPrefix(candidates, token)
	})
}

// filterByPrefix returns the entries of candidates that start with token,
// dedup-ed in original order. Returns nil (not an empty slice) when nothing
// matches so callers can use the standard `if got != nil` idiom.
func filterByPrefix(candidates []string, token string) []string {
	if len(candidates) == 0 {
		return nil
	}
	var out []string
	seen := make(map[string]bool, len(candidates))
	for _, c := range candidates {
		if seen[c] {
			continue
		}
		seen[c] = true
		if strings.HasPrefix(c, token) {
			out = append(out, c)
		}
	}
	return out
}

// RunInteractiveCuiLoop runs an interactive multi-game CUI loop with game
// switching support. Returns 0 on normal exit (EOF, "quit", "exit", or
// QuitSentinel from a game) and 1 when stdin reads fail with a non-EOF I/O
// error (issue #1839). The manager handles help/? commands internally;
// other commands are delegated to the current game.
func RunInteractiveCuiLoop(manager *GameManager) int {
	initMsg := manager.InitCurrentGame()
	if initMsg != "" {
		// through printResult, not fmt.Println: the first deal is the one place
		// a raw marker would reach the screen unnoticed.
		printResult(initMsg)
	}
	fmt.Println(i18n.T("typeHelp"))
	reader := newDefaultLineReader()
	defer func() { _ = reader.Close() }()
	// On SIGINT/SIGTERM, close the reader (history save + terminal restore)
	// before exiting since os.Exit skips the defer above (issue #2096).
	defer runSignalWatcher(func() { _ = reader.Close() })()
	// *GameManager statically satisfies commandLister, so no type assertion
	// is needed here. RunCuiLoop below uses one because its parameter is the
	// CuiExecer interface and we don't know the concrete type.
	installCompleter(reader, manager)
	for {
		input, exit, ioErr := readInput(reader, manager.CurrentGame())
		if ioErr != nil {
			return 1
		}
		if exit {
			fmt.Println(i18n.T("bye"))
			return 0
		}
		reader.AppendHistory(input)
		res := manager.Exec(input)
		res, ioErr = handlePromptLoop(reader, manager, res, manager.CurrentGame(), os.Stdout)
		if ioErr != nil {
			return 1
		}
		if res == i18n.QuitSentinel {
			fmt.Println(i18n.T("bye"))
			return 0
		}
		printResult(res)
	}
}

// printResult writes res to stdout, or to stderr when marked as an error.
func printResult(res string) {
	if body, isErr := i18n.StripErrorPrefix(res); isErr {
		fmt.Fprintln(os.Stderr, color.RedStderr(body))
		return
	}
	// A board carrying a refusal line stays on stdout -- it is a board. The
	// marker exists so callers can tell, not to change where it goes.
	body, _ := i18n.StripErrorLines(res)
	fmt.Println(body)
}

// RunCuiLoop runs a single-game CUI loop. gameName is shown in the prompt
// (e.g. "[blackjack] > ") so single-game mode matches the interactive-mode
// prompt — this gives scrollback context and lets users tell ターミナル/タブ
// apart when running multiple games in parallel. Pass "" to keep the legacy
// bare "> " prompt. helpLines is shown when the user types "help" / "?".
//
// Returns 0 on normal exit (EOF, "quit", QuitSentinel) and 1 when stdin
// reads fail with a non-EOF I/O error (issue #1839). See issue #1605.
func RunCuiLoop(gameName string, controller CuiExecer, helpLines []string) int {
	printResult(controller.Exec("r"))
	fmt.Println(i18n.T("typeHelp"))
	reader := newDefaultLineReader()
	defer func() { _ = reader.Close() }()
	// On SIGINT/SIGTERM, close the reader (history save + terminal restore)
	// before exiting since os.Exit skips the defer above (issue #2096).
	defer runSignalWatcher(func() { _ = reader.Close() })()
	if lister, ok := controller.(commandLister); ok {
		installCompleter(reader, lister)
	} else {
		installCompleter(reader, nil)
	}
	for {
		input, exit, ioErr := readInput(reader, gameName)
		if ioErr != nil {
			return 1
		}
		if exit {
			fmt.Println(i18n.T("bye"))
			return 0
		}
		reader.AppendHistory(input)
		trimmed := strings.TrimSpace(input)
		if trimmed == "help" || trimmed == "?" {
			for _, line := range helpLines {
				fmt.Println(line)
			}
			continue
		}
		res := controller.Exec(input)
		res, ioErr = handlePromptLoop(reader, controller, res, gameName, os.Stdout)
		if ioErr != nil {
			return 1
		}
		if res == i18n.QuitSentinel {
			fmt.Println(i18n.T("bye"))
			return 0
		}
		printResult(res)
	}
}

// handlePromptLoop handles interactive prompting when a controller returns a
// prompt request. It loops until the controller returns a non-prompt result
// (allowing chained wizard-style prompts). Returns (result, ioErr); ioErr is
// non-nil only when an inner stdin read fails with a non-EOF error so the
// top-level runner can exit non-zero (issue #1839). EOF stays a clean quit
// (QuitSentinel + nil error), matching the bye-on-Ctrl+D contract.
func handlePromptLoop(reader LineReader, execer CuiExecer, result, gameName string, w io.Writer) (string, error) {
	for cuiutil.IsPromptRequest(result) {
		promptMsg, tmpl := cuiutil.ParsePromptRequest(result)
		if tmpl == "" {
			// Malformed prompt; treat as regular message.
			return promptMsg, nil
		}
		_, _ = fmt.Fprintln(w, promptMsg)
		input, exit, ioErr := readInput(reader, gameName)
		if ioErr != nil {
			return i18n.QuitSentinel, ioErr
		}
		if exit {
			return i18n.QuitSentinel, nil
		}
		input = strings.TrimSpace(input)
		if input == "" {
			return i18n.T("cancelled"), nil
		}
		fullCmd := cuiutil.FillTemplate(tmpl, input)
		result = execer.Exec(fullCmd)
	}
	return result, nil
}
