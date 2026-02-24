package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/ui"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/web"
)

func main() {
	flag.Parse()
	switch strings.ToLower(flag.Arg(0)) {
	case "blackjack":
		cui := ui.NewBlackJackCui()
		cui.Exec()
	case "poker":
		poker := ui.NewPokerCui()
		poker.Exec()
	case "oldmaid":
		oldmaid := ui.NewOldMaidCui()
		oldmaid.Exec()
	case "daifugo":
		daifugo := ui.NewDaifugoCui()
		daifugo.Exec()
	case "sevens":
		sevens := ui.NewSevensCui()
		sevens.Exec()
	case "web":
		w := web.NewTrumpCardsWeb()
		w.Exec()
	default:
		log.Fatal(fmt.Errorf("Error: param not found %s", flag.Arg(0)))
	}
}
