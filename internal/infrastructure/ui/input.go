package ui

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// readInput reads one line from r, prepending the prompt with gameName when
// non-empty. Returns (text, exit) — exit is true on EOF (pipe drained,
// Ctrl+D, or Ctrl+C in liner-backed readers) and on any other read error.
//
// The prompt is constructed here rather than in cui_runner.go so the same
// "[gameName] > " format is used by the top-level loop and the wizard-style
// prompt loop (issue #1605).
func readInput(r LineReader, gameName string) (text string, exit bool) {
	prompt := "> "
	if gameName != "" {
		prompt = "[" + gameName + "] > "
	}
	line, err := r.Prompt(prompt)
	if err != nil {
		if errors.Is(err, io.EOF) {
			fmt.Println(i18n.T("bye"))
		} else {
			fmt.Fprintln(os.Stderr, i18n.Tf("inputReadError", "error", err.Error()))
		}
		return "", true
	}
	return line, false
}
