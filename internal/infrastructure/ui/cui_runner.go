package ui

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

var signalHandlerOnce sync.Once

// setupSignalHandler registers a handler for SIGINT and SIGTERM that prints
// a goodbye message and exits gracefully. Call this at the start of any CUI loop.
// It is safe to call multiple times; only the first call has an effect.
func setupSignalHandler() {
	signalHandlerOnce.Do(func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		go func() {
			sig := <-sigCh
			signal.Stop(sigCh)
			fmt.Println("\n" + i18n.T("bye"))
			// POSIX convention: exit code = 128 + signal number
			exitCode := 128
			if sigNum, ok := sig.(syscall.Signal); ok {
				exitCode += int(sigNum)
			}
			os.Exit(exitCode)
		}()
	})
}

// CuiExecer CUIコントローラの共通インタフェース
type CuiExecer interface {
	Exec(command string) string
}

// commandLister exposes the alias set a controller will accept. It's
// implemented by any type whose value-set is reachable through CuiExecer +
// reflection-free introspection — currently just *GameManager — so the
// completer below can suggest game-specific aliases. Controllers that don't
// implement it get only the common command set.
type commandLister interface {
	CompletionCandidates() []string
}

// commonCompletionCommands is the always-available command set. Mirrors
// cui_controller_helper.go's commonCommands; duplicated here because that
// list lives in the adapter layer and we want infra to stay independent.
var commonCompletionCommands = []string{
	"q", "quit", "exit",
	"r", "reset",
	"help", "?",
}

// installCompleter wires a tab-completion function into r. The completer
// expands the first token only — argful commands (e.g. "b 100") still need
// the user to type the argument by hand, since alias completion can't know
// game-state-dependent valid arguments.
func installCompleter(r LineReader, lister commandLister) {
	r.SetCompleter(func(prefix string) []string {
		// liner passes the full line; we complete the first whitespace-bounded token.
		fields := strings.Fields(prefix)
		// If the cursor is past the first token, no completion (we don't know
		// per-game arg shapes).
		if len(fields) > 1 || (len(fields) == 1 && strings.HasSuffix(prefix, " ")) {
			return nil
		}
		var token string
		if len(fields) == 1 {
			token = fields[0]
		}
		candidates := commonCompletionCommands
		if lister != nil {
			candidates = append(candidates, lister.CompletionCandidates()...)
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
	})
}

// RunInteractiveCuiLoop runs an interactive multi-game CUI loop with game switching support.
// The manager handles help/? commands internally; other commands are delegated to the current game.
func RunInteractiveCuiLoop(manager *GameManager) {
	setupSignalHandler()
	initMsg := manager.InitCurrentGame()
	if initMsg != "" {
		fmt.Println(initMsg)
	}
	fmt.Println(i18n.T("typeHelp"))
	reader := newDefaultLineReader()
	defer func() { _ = reader.Close() }()
	if lister, ok := any(manager).(commandLister); ok {
		installCompleter(reader, lister)
	} else {
		installCompleter(reader, nil)
	}
	for {
		input, exit := readInput(reader, manager.CurrentGame())
		if exit {
			break
		}
		reader.AppendHistory(input)
		res := manager.Exec(input)
		res = handlePromptLoop(reader, manager, res, manager.CurrentGame(), os.Stdout)
		if res == i18n.QuitSentinel {
			fmt.Println(i18n.T("bye"))
			break
		}
		printResult(res)
	}
}

// printResult writes res to stdout, or to stderr when marked as an error.
func printResult(res string) {
	if body, isErr := i18n.StripErrorPrefix(res); isErr {
		fmt.Fprintln(os.Stderr, body)
		return
	}
	fmt.Println(res)
}

// RunCuiLoop runs a single-game CUI loop. gameName is shown in the prompt
// (e.g. "[blackjack] > ") so single-game mode matches the interactive-mode
// prompt — this gives scrollback context and lets users tell ターミナル/タブ
// apart when running multiple games in parallel. Pass "" to keep the legacy
// bare "> " prompt. helpLines is shown when the user types "help" / "?".
// See issue #1605.
func RunCuiLoop(gameName string, controller CuiExecer, helpLines []string) {
	setupSignalHandler()
	fmt.Println(controller.Exec("r"))
	fmt.Println(i18n.T("typeHelp"))
	reader := newDefaultLineReader()
	defer func() { _ = reader.Close() }()
	if lister, ok := controller.(commandLister); ok {
		installCompleter(reader, lister)
	} else {
		installCompleter(reader, nil)
	}
	for {
		input, exit := readInput(reader, gameName)
		if exit {
			break
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
		res = handlePromptLoop(reader, controller, res, gameName, os.Stdout)
		if res == i18n.QuitSentinel {
			fmt.Println(i18n.T("bye"))
			break
		}
		printResult(res)
	}
}

// handlePromptLoop handles interactive prompting when a controller returns a prompt request.
// It loops until the controller returns a non-prompt result (allowing chained wizard-style prompts).
func handlePromptLoop(reader LineReader, execer CuiExecer, result, gameName string, w io.Writer) string {
	for cuiutil.IsPromptRequest(result) {
		promptMsg, tmpl := cuiutil.ParsePromptRequest(result)
		if tmpl == "" {
			// Malformed prompt; treat as regular message.
			return promptMsg
		}
		_, _ = fmt.Fprintln(w, promptMsg)
		input, exit := readInput(reader, gameName)
		if exit {
			return i18n.QuitSentinel
		}
		input = strings.TrimSpace(input)
		if input == "" {
			return i18n.T("cancelled")
		}
		fullCmd := cuiutil.FillTemplate(tmpl, input)
		result = execer.Exec(fullCmd)
	}
	return result
}
