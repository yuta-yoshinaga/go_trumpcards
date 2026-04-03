package ui

import (
	"bufio"
	"fmt"
	"os"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// readInput はスキャナから1行を読み込み、EOFとエラーを処理します。
// 入力テキストと、プログラムを終了すべきかを示すブール値を返します。
// gameName が空でない場合、プロンプトに "[gameName] > " と表示します。
func readInput(scanner *bufio.Scanner, gameName string) (text string, exit bool) {
	if gameName != "" {
		_, _ = fmt.Fprintf(os.Stdout, "[%s] > ", gameName)
	} else {
		_, _ = fmt.Fprint(os.Stdout, "> ")
	}
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, i18n.Tf("inputReadError", "error", err.Error()))
		} else {
			// EOF: pipe input exhausted or Ctrl+D
			fmt.Println(i18n.T("bye"))
		}
		return "", true
	}
	return scanner.Text(), false
}
