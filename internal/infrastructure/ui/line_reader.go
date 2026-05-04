package ui

import (
	"bufio"
	"fmt"
	"io"
)

// LineReader reads a single line of user input. Implementations may add
// readline conveniences (history, tab completion, Ctrl+R search) on top of
// raw stdin; tests use the trivial bufio.Scanner-backed implementation.
//
// The interface is intentionally narrow: callers neither know nor care
// whether they're talking to a terminal-backed editor or a piped Reader, and
// the test suite can swap implementations without dragging in tty mocks.
type LineReader interface {
	// Prompt displays prompt and reads one line. Returns io.EOF when stdin is
	// exhausted (or the user pressed Ctrl+D), and any other error verbatim.
	// The trailing newline is stripped.
	Prompt(prompt string) (string, error)

	// AppendHistory records line for arrow-key recall. No-op for readers
	// without history support.
	AppendHistory(line string)

	// SetCompleter installs a tab-completion function. fn receives the
	// already-typed prefix and returns matching completions. No-op for
	// readers without completion support.
	SetCompleter(fn func(prefix string) []string)

	// Close releases any underlying resources (terminal raw mode, history
	// file). Always returns nil for readers that hold no resources.
	Close() error
}

// scannerLineReader is the minimal LineReader: writes the prompt to w and
// reads a line via bufio.Scanner. Used in tests and as the WASM fallback —
// no history, no completion, no terminal awareness.
type scannerLineReader struct {
	scanner *bufio.Scanner
	w       io.Writer
}

// newScannerLineReader wraps r/w as a LineReader. Tests use this with
// strings.Reader + bytes.Buffer to drive readInput deterministically.
func newScannerLineReader(r io.Reader, w io.Writer) *scannerLineReader {
	return &scannerLineReader{
		scanner: bufio.NewScanner(r),
		w:       w,
	}
}

// Prompt writes prompt to the underlying writer and reads the next line.
// Returns io.EOF when the scanner has no more input.
func (s *scannerLineReader) Prompt(prompt string) (string, error) {
	_, _ = fmt.Fprint(s.w, prompt)
	if !s.scanner.Scan() {
		if err := s.scanner.Err(); err != nil {
			return "", err
		}
		return "", io.EOF
	}
	return s.scanner.Text(), nil
}

// AppendHistory is a no-op for the scanner reader.
func (*scannerLineReader) AppendHistory(string) {}

// SetCompleter is a no-op for the scanner reader.
func (*scannerLineReader) SetCompleter(func(string) []string) {}

// Close is a no-op for the scanner reader.
func (*scannerLineReader) Close() error { return nil }
