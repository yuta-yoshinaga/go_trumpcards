package color

import "sync/atomic"

var noColor atomic.Bool

// SetNoColor sets whether color output is disabled.
// When true, callers should suppress ANSI color escape codes.
func SetNoColor(v bool) {
	noColor.Store(v)
}

// NoColor reports whether color output is disabled.
func NoColor() bool {
	return noColor.Load()
}

const reset = "\033[0m"

func wrap(code, s string) string {
	if noColor.Load() {
		return s
	}
	return code + s + reset
}

// Red wraps s with red ANSI color code.
func Red(s string) string { return wrap("\033[31m", s) }

// Green wraps s with green ANSI color code.
func Green(s string) string { return wrap("\033[32m", s) }

// Yellow wraps s with yellow ANSI color code.
func Yellow(s string) string { return wrap("\033[33m", s) }

// Bold wraps s with bold ANSI code.
func Bold(s string) string { return wrap("\033[1m", s) }

// BoldYellow wraps s with bold yellow ANSI code.
func BoldYellow(s string) string { return wrap("\033[1;33m", s) }
