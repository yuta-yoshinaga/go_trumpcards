package ui

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// readInput reads one line from r, prepending the prompt with gameName when
// non-empty. Returns (text, exit, ioErr):
//
//   - normal line: ("...", false, nil)
//   - EOF (pipe drained, Ctrl+D, signalled-quit reader): ("", true, nil)
//   - any other I/O error: ("", true, err)
//
// EOF and "exit" are clean shutdowns; only the non-EOF error path carries an
// ioErr so callers can map it to a non-zero process exit code (issue #1839).
// Matches the Updater's stdin contract (Updater.go: scan errors propagate as
// errors), giving the whole CLI a consistent exit-code-on-stdin-failure rule.
//
// The prompt is constructed here rather than in cui_runner.go so the same
// "[gameName] > " format is used by the top-level loop and the wizard-style
// prompt loop (issue #1605).
func readInput(r LineReader, gameName string) (text string, exit bool, ioErr error) {
	prompt := "> "
	if gameName != "" {
		prompt = "[" + gameName + "] > "
	}
	line, err := r.Prompt(prompt)
	if err != nil {
		if errors.Is(err, io.EOF) {
			fmt.Println(i18n.T("bye"))
			return "", true, nil
		}
		fmt.Fprintln(os.Stderr, i18n.Tf("inputReadError", "error", err.Error()))
		return "", true, err
	}
	return line, false, nil
}
