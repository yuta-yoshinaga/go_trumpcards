package main

import (
	"log/slog"
	"os"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/web"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
	w := web.NewTrumpCardsWeb()
	w.Exec()
}
