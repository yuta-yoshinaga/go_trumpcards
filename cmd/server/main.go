package main

import (
	"fmt"
	"os"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/web"
)

func main() {
	infrastructure.InitLogger()
	w := web.NewTrumpCardsWeb()
	if err := w.Exec(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
