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
