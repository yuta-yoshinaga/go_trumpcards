package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
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
	fmt.Println(`Type "help" or "?" for commands.`)
	scanner := bufio.NewScanner(os.Stdin)
	for {
		input, exit := readInput(scanner)
		if exit {
			break
		}
		res := manager.Exec(input)
		fmt.Println(res)
		if res == "bye." {
			break
		}
	}
}

// RunCuiLoop 標準CUIゲームループを実行する
// helpLines は "help" / "?" コマンドが入力されたときのみ表示される
func RunCuiLoop(controller CuiExecer, helpLines []string) {
	fmt.Println(controller.Exec("r"))
	fmt.Println(`Type "help" or "?" for commands.`)
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
		fmt.Println(res)
		if res == "bye." {
			break
		}
	}
}
