package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/frameworks_drivers/ui"
	"github.com/yuta-yoshinaga/go_trumpcards/frameworks_drivers/web"
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
		// CUI is not implemented yet, but for consistency with others
		fmt.Println("Daifugo CUI is not implemented yet. Please use 'web' mode.")
	case "web":
		web := web.NewTrumpCardsWeb()
		web.Exec()
	default:
		log.Fatal(fmt.Errorf("Error: param not found %s", flag.Arg(0)))
	}
}
