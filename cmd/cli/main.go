package main

import (
	"flag"
	"log/slog"
	"os"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure"
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
	case "doubt":
		doubt := ui.NewDoubtCui()
		doubt.Exec()
	case "holdem":
		holdem := ui.NewHoldemCui()
		holdem.Exec()
	case "hearts":
		hearts := ui.NewHeartsCui()
		hearts.Exec()
	case "memory":
		memory := ui.NewMemoryCui()
		memory.Exec()
	case "klondike":
		klondike := ui.NewKlondikeCui()
		klondike.Exec()
	case "baccarat":
		baccarat := ui.NewBaccaratCui()
		baccarat.Exec()
	case "web":
		infrastructure.InitLogger()
		w := web.NewTrumpCardsWeb()
		w.Exec()
	default:
		if flag.Arg(0) != "" {
			slog.Error("unknown command", "arg", flag.Arg(0))
			os.Exit(1)
		}
		// No argument: start interactive multi-game mode (defaults to blackjack).
		manager := ui.NewGameManager("blackjack")
		ui.RunInteractiveCuiLoop(manager)
	}
}
