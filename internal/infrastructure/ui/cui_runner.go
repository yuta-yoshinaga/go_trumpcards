package ui

import (
	"bufio"
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

// RunInteractiveCuiLoop runs an interactive multi-game CUI loop with game switching support.
// The manager handles help/? commands internally; other commands are delegated to the current game.
func RunInteractiveCuiLoop(manager *GameManager) {
	setupSignalHandler()
	initMsg := manager.InitCurrentGame()
	if initMsg != "" {
		fmt.Println(initMsg)
	}
	fmt.Println(i18n.T("typeHelp"))
	scanner := bufio.NewScanner(os.Stdin)
	for {
		input, exit := readInput(scanner, manager.CurrentGame(), os.Stdout)
		if exit {
			break
		}
		res := manager.Exec(input)
		res = handlePromptLoop(scanner, manager, res, manager.CurrentGame(), os.Stdout)
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
	scanner := bufio.NewScanner(os.Stdin)
	for {
		input, exit := readInput(scanner, gameName, os.Stdout)
		if exit {
			break
		}
		trimmed := strings.TrimSpace(input)
		if trimmed == "help" || trimmed == "?" {
			for _, line := range helpLines {
				fmt.Println(line)
			}
			continue
		}
		res := controller.Exec(input)
		res = handlePromptLoop(scanner, controller, res, gameName, os.Stdout)
		if res == i18n.QuitSentinel {
			fmt.Println(i18n.T("bye"))
			break
		}
		printResult(res)
	}
}

// handlePromptLoop handles interactive prompting when a controller returns a prompt request.
// It loops until the controller returns a non-prompt result (allowing chained wizard-style prompts).
func handlePromptLoop(scanner *bufio.Scanner, execer CuiExecer, result, gameName string, w io.Writer) string {
	for cuiutil.IsPromptRequest(result) {
		promptMsg, tmpl := cuiutil.ParsePromptRequest(result)
		if tmpl == "" {
			// Malformed prompt; treat as regular message.
			return promptMsg
		}
		_, _ = fmt.Fprintln(w, promptMsg)
		input, exit := readInput(scanner, gameName, w)
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
