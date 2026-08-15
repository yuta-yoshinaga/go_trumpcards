package infrastructure

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"golang.org/x/term"
)

// InitLogger sets up the default structured logger on stderr.
//
// The log level is controlled by the LOG_LEVEL environment variable
// (DEBUG, INFO, WARN, ERROR; default: INFO).
//
// The log format is controlled by the LOG_FORMAT environment variable
// ("text" or "json"). When it is unset the format follows the stream: text on
// a terminal, JSON everywhere else.
//
// That default matters because `trumpcards web` already prints a
// human-readable banner to stderr, and slog fires the same lifecycle event
// beside it. Pinning JSON regardless of destination made a terminal user read
// the address twice, and a failed bind print the same failure twice, once in
// each format. Keying the default on the stream leaves the machine-facing side
// untouched — systemd / docker / log shippers are never a terminal, so they
// still get the structured events issue #1452 added. See issue #5357.
func InitLogger() {
	isTTY := term.IsTerminal(int(os.Stderr.Fd()))
	slog.SetDefault(slog.New(logHandlerFor(
		os.Stderr,
		os.Getenv("LOG_FORMAT"),
		logLevelFor(os.Getenv("LOG_LEVEL"), isTTY),
		isTTY,
	)))
}

// logHandlerFor picks the handler for one stream.
//
// Split from InitLogger so the decision is testable without a real terminal:
// the TTY check happens at the edge, the same way main.go resolves colour and
// the update progress bar.
func logHandlerFor(w io.Writer, format string, level slog.Level, isTTY bool) slog.Handler {
	opts := &slog.HandlerOptions{Level: level}
	switch strings.ToLower(format) {
	case "text":
		return slog.NewTextHandler(w, opts)
	case "json":
		return slog.NewJSONHandler(w, opts)
	}
	// Unset or unrecognised: follow the stream rather than guessing.
	if isTTY {
		return slog.NewTextHandler(w, opts)
	}
	return slog.NewJSONHandler(w, opts)
}

// logLevelFor resolves the level for one stream.
//
// An explicit LOG_LEVEL always wins. With none set, a terminal defaults to WARN
// and everything else to INFO: on a terminal the INFO lifecycle events restate
// what the human-readable banner already said ("listening at …", "stopped"),
// so emitting both makes the reader check whether two different things
// happened. Warnings and errors still print. A log shipper keeps INFO.
func logLevelFor(env string, isTTY bool) slog.Level {
	if env != "" {
		return parseLogLevel(env)
	}
	if isTTY {
		return slog.LevelWarn
	}
	return slog.LevelInfo
}

func parseLogLevel(s string) slog.Level {
	switch strings.ToUpper(s) {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	case "INFO":
		return slog.LevelInfo
	default: // empty string or unrecognized value
		return slog.LevelInfo
	}
}
