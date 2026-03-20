package ui

import (
	"bufio"
	"fmt"
	"os"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// readInput はスキャナから1行を読み込み、EOFとエラーを処理します。
// 入力テキストと、プログラムを終了すべきかを示すブール値を返します。
func readInput(scanner *bufio.Scanner) (text string, exit bool) {
	fmt.Fprint(os.Stderr, "> ")
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, i18n.Tf("inputReadError", "error", err.Error()))
		}
		return "", true
	}
	return scanner.Text(), false
}
