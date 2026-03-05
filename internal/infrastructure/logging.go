package infrastructure

import (
	"log/slog"
	"os"
)

// InitLogger sets up the default structured JSON logger on stderr.
func InitLogger() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
}
