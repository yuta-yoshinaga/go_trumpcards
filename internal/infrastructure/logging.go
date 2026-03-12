package infrastructure

import (
	"log/slog"
	"os"
	"strings"
)

// InitLogger sets up the default structured logger on stderr.
// The log level is controlled by the LOG_LEVEL environment variable
// (DEBUG, INFO, WARN, ERROR; default: INFO).
// The log format is controlled by the LOG_FORMAT environment variable
// ("text" for human-readable output; default: JSON).
func InitLogger() {
	level := parseLogLevel(os.Getenv("LOG_LEVEL"))
	opts := &slog.HandlerOptions{Level: level}
	var handler slog.Handler
	if strings.ToLower(os.Getenv("LOG_FORMAT")) == "text" {
		handler = slog.NewTextHandler(os.Stderr, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	}
	slog.SetDefault(slog.New(handler))
}

func parseLogLevel(s string) slog.Level {
	switch strings.ToUpper(s) {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
