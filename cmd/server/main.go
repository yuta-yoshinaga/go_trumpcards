package main

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/web"
)

func main() {
	infrastructure.InitLogger()
	w := web.NewTrumpCardsWeb()
	w.Exec()
}
