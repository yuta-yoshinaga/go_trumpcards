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
//   - any other I/O error: ("", true, err); inputReadError already written to stderr
//
// Side effects are deliberately scoped to error reporting: the "bye" goodbye
// line is printed by the top-level loop on EOF, not here, so a Ctrl+D inside
// a wizard-style prompt only prints "bye" once rather than twice. EOF and
// "exit" remain clean shutdowns; only the non-EOF path carries an ioErr so
// callers can map it to a non-zero process exit code. Matches the Updater's
// stdin contract (Updater.go: scan errors propagate as errors).
//
// The prompt is constructed here rather than in cui_runner.go so the same
// "[gameName] > " format is used by the top-level loop and the wizard-style
// prompt loop.
func readInput(r LineReader, gameName string) (text string, exit bool, ioErr error) {
	prompt := "> "
	if gameName != "" {
		prompt = "[" + gameName + "] > "
	}
	line, err := r.Prompt(prompt)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return "", true, nil
		}
		fmt.Fprintln(os.Stderr, i18n.Tf("inputReadError", "error", err.Error()))
		return "", true, err
	}
	return line, false, nil
}
