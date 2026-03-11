package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/ui"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/web"
)

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, `USAGE:
  go_trumpcards [game]
  go_trumpcards --help

GAMES:
  blackjack    BlackJack (ブラックジャック)
  poker        5-card Draw Poker (ポーカー)
  oldmaid      Old Maid (ババ抜き)
  daifugo      Daifugo / Great Fool (大富豪)
  sevens       Sevens (7並べ)
  doubt        Doubt (ダウト)
  holdem       Texas Hold'em (テキサスホールデム)
  hearts       Hearts (ハーツ)
  memory       Memory / Concentration (神経衰弱)
  klondike     Klondike Solitaire (ソリティア)
  baccarat     Baccarat (バカラ)
  web          Start REST API + web GUI server

  (no argument) Interactive mode with game switching

OPTIONS:
  -h, --help   Show this help message
`)
	}
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
			flag.Usage()
			os.Exit(1)
		}
		// No argument: start interactive multi-game mode (defaults to blackjack).
		manager := ui.NewGameManager("blackjack")
		ui.RunInteractiveCuiLoop(manager)
	}
}
