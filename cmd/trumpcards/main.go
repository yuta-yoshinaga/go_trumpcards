package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/ui"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/update"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/web"
)

// Injected via -ldflags at build time (e.g. by GoReleaser).
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	os.Exit(run())
}

func run() int {
	lang := flag.String("lang", "", "language (ja or en)")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.BoolVar(showVersion, "V", false, "Show version information (shorthand)")
	noColorFlag := flag.Bool("no-color", false, "Disable color output")
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, `USAGE:
  trumpcards [--lang ja|en] [game]
  trumpcards --help

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
  update       Self-update to the latest version
  web          Start REST API + web GUI server

  (no argument) Interactive mode with game switching

OPTIONS:
  -h, --help        Show this help message
  --lang ja|en      Language (default: ja)
  --no-color        Disable color output
  -V, --version     Show version information

ENVIRONMENT VARIABLES:
  NO_COLOR          Disable color output when set (see https://no-color.org/)
                    Example: NO_COLOR=1 trumpcards blackjack
  PORT              Port number for the web server (default: 8080)
                    Example: PORT=3000 trumpcards web
`)
	}
	flag.Parse()

	// Color control: NO_COLOR env var (https://no-color.org/) or --no-color flag.
	if os.Getenv("NO_COLOR") != "" || *noColorFlag {
		color.SetNoColor(true)
	}

	if *showVersion {
		fmt.Printf("trumpcards %s (commit: %s, built: %s)\n", version, commit, date)
		return 0
	}

	// Language detection: --lang > LANG env > default "ja"
	detectedLang := "ja"
	if envLang := os.Getenv("LANG"); envLang != "" {
		prefix := envLang
		if idx := strings.IndexAny(envLang, "_-."); idx >= 0 {
			prefix = envLang[:idx]
		}
		if prefix == "en" || prefix == "ja" {
			detectedLang = prefix
		}
	}
	if *lang != "" {
		detectedLang = *lang
	}
	i18n.SetLang(detectedLang)
	if i18n.Lang() != detectedLang && detectedLang != "" {
		fmt.Fprintf(os.Stderr, "Warning: unsupported language %q, defaulting to ja\n", detectedLang)
	}
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
	case "update":
		updater := update.NewUpdater(version, os.Stdin, os.Stdout, os.Stderr)
		if err := updater.Exec(); err != nil {
			return 1
		}
	case "web":
		infrastructure.InitLogger()
		w := web.NewTrumpCardsWeb()
		if err := w.Exec(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
	default:
		if flag.Arg(0) != "" {
			fmt.Fprintf(os.Stderr, "Error: unknown game %q\n\n", flag.Arg(0))
			flag.Usage()
			return 2
		}
		// No argument: start interactive multi-game mode (defaults to blackjack).
		manager := ui.NewGameManager("blackjack")
		ui.RunInteractiveCuiLoop(manager)
	}
	return 0
}
