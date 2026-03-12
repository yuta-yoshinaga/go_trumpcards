package color

var noColor bool

// SetNoColor sets whether color output is disabled.
// When true, callers should suppress ANSI color escape codes.
func SetNoColor(v bool) {
	noColor = v
}

// NoColor reports whether color output is disabled.
func NoColor() bool {
	return noColor
}
