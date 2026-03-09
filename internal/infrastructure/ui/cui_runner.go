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
