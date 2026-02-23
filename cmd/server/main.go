package main

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/web"
)

func main() {
	w := web.NewTrumpCardsWeb()
	w.Exec()
}
