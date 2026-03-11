package ui

import (
	"bufio"
	"fmt"
	"os"
)

// CuiExecer CUIコントローラの共通インタフェース
type CuiExecer interface {
	Exec(command string) string
}

// RunInteractiveCuiLoop runs an interactive multi-game CUI loop with game switching support.
// The manager provides dynamic help lines and routes switch/games commands internally.
func RunInteractiveCuiLoop(manager *GameManager) {
	initMsg := manager.InitCurrentGame()
	if initMsg != "" {
		fmt.Println(initMsg)
	}
	scanner := bufio.NewScanner(os.Stdin)
	for {
		for _, line := range manager.HelpLines() {
			fmt.Println(line)
		}
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
// helpLines は毎回入力前に表示されるヘルプメッセージ
func RunCuiLoop(controller CuiExecer, helpLines []string) {
	fmt.Println(controller.Exec("r"))
	scanner := bufio.NewScanner(os.Stdin)
	for {
		for _, line := range helpLines {
			fmt.Println(line)
		}
		input, exit := readInput(scanner)
		if exit {
			break
		}
		res := controller.Exec(input)
		fmt.Println(res)
		if res == "bye." {
			break
		}
	}
}
