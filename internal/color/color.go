package color

import "sync/atomic"

// Independent flags for stdout and stderr color output. This lets the CLI
// enable color on stderr while stdout is being piped (e.g. `trumpcards <game> | tee log.txt`)
// and conversely disable stderr color when it is redirected to a file.
var (
	noColorStdout atomic.Bool
	noColorStderr atomic.Bool
)

// SetNoColor disables color output for both stdout and stderr when v is true,
// and enables it for both when v is false. Kept for backward compatibility with
// callers that do not need per-stream control.
func SetNoColor(v bool) {
	noColorStdout.Store(v)
	noColorStderr.Store(v)
}

// SetStdoutColor enables (true) or disables (false) color output for stdout.
func SetStdoutColor(enabled bool) { noColorStdout.Store(!enabled) }

// SetStderrColor enables (true) or disables (false) color output for stderr.
func SetStderrColor(enabled bool) { noColorStderr.Store(!enabled) }

// NoColor reports whether stdout color output is disabled.
// Kept as the reader half of SetNoColor, which CUI presenter tests use to save
// and restore color state; prefer NoColorStdout / NoColorStderr in new code.
func NoColor() bool { return noColorStdout.Load() }

// NoColorStdout reports whether color output is disabled for stdout.
func NoColorStdout() bool { return noColorStdout.Load() }

// NoColorStderr reports whether color output is disabled for stderr.
func NoColorStderr() bool { return noColorStderr.Load() }

const reset = "\033[0m"

func wrap(code, s string, off bool) string {
	if s == "" || off {
		return s
	}
	return code + s + reset
}

// Red wraps s with red ANSI color code (stdout-targeted).
func Red(s string) string { return wrap("\033[31m", s, noColorStdout.Load()) }

// Green wraps s with green ANSI color code (stdout-targeted).
func Green(s string) string { return wrap("\033[32m", s, noColorStdout.Load()) }

// Yellow wraps s with yellow ANSI color code (stdout-targeted).
func Yellow(s string) string { return wrap("\033[33m", s, noColorStdout.Load()) }

// Bold wraps s with bold ANSI code (stdout-targeted).
func Bold(s string) string { return wrap("\033[1m", s, noColorStdout.Load()) }

// BoldYellow wraps s with bold yellow ANSI code (stdout-targeted).
func BoldYellow(s string) string { return wrap("\033[1;33m", s, noColorStdout.Load()) }
