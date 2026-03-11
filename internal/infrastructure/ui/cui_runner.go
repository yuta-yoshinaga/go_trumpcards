package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// CuiExecer CUIコントローラの共通インタフェース
type CuiExecer interface {
	Exec(command string) string
}

// RunInteractiveCuiLoop runs an interactive multi-game CUI loop with game switching support.
// The manager handles help/? commands internally; other commands are delegated to the current game.
func RunInteractiveCuiLoop(manager *GameManager) {
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
