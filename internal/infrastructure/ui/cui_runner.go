package ui

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

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
		input, exit := readInput(scanner, "", os.Stdout)
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
