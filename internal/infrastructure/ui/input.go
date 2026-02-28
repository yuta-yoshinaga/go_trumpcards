package ui

import (
	"bufio"
	"fmt"
	"os"
)

// readInput reads a line from the scanner, handling EOF and errors.
// It returns the input text and a boolean indicating if the program should exit.
func readInput(scanner *bufio.Scanner) (text string, exit bool) {
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "入力の読み取り中にエラーが発生しました: %v\n", err)
		}
		return "", true
	}
	return scanner.Text(), false
}
