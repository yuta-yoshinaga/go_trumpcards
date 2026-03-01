package ui

import (
	"bufio"
	"fmt"
	"os"
)

// readInput はスキャナから1行を読み込み、EOFとエラーを処理します。
// 入力テキストと、プログラムを終了すべきかを示すブール値を返します。
func readInput(scanner *bufio.Scanner) (text string, exit bool) {
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "入力の読み取り中にエラーが発生しました: %v\n", err)
		}
		return "", true
	}
	return scanner.Text(), false
}
