package ui

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// setupSignalHandler registers a handler for SIGINT and SIGTERM that prints
// a goodbye message and exits gracefully. Call this at the start of any CUI loop.
func setupSignalHandler() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		signal.Stop(sigCh)
		fmt.Fprintln(os.Stderr, "\n"+i18n.T("bye"))
		os.Exit(0)
	}()
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
		input, exit := readInput(scanner)
		if exit {
			break
		}
		res := manager.Exec(input)
		if res == i18n.QuitSentinel {
			fmt.Println(i18n.T("bye"))
			break
		}
		fmt.Println(res)
	}
}

// RunCuiLoop 標準CUIゲームループを実行する
// helpLines は "help" / "?" コマンドが入力されたときのみ表示される
func RunCuiLoop(controller CuiExecer, helpLines []string) {
	setupSignalHandler()
	fmt.Println(controller.Exec("r"))
	fmt.Println(i18n.T("typeHelp"))
	scanner := bufio.NewScanner(os.Stdin)
	for {
		input, exit := readInput(scanner)
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
		if res == i18n.QuitSentinel {
			fmt.Println(i18n.T("bye"))
			break
		}
		fmt.Println(res)
	}
}
