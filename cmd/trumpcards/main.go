package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
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
	helpText := `USAGE:
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
  omaha        Omaha Hold'em (オマハホールデム)
  hearts       Hearts (ハーツ)
  memory       Memory / Concentration (神経衰弱)
  klondike     Klondike Solitaire (ソリティア)
  freecell     FreeCell (フリーセル)
  baccarat     Baccarat (バカラ)
  spades       Spades (スペード)
  crazyeights  Crazy Eights (クレイジーエイト)
  ginrummy     Gin Rummy (ジンラミー)
  update       Self-update to the latest version
  web          Start REST API + web GUI server

  (no argument) Interactive mode with game switching

OPTIONS:
  -h, --help        Show this help message
  --lang ja|en      Language (default: ja)
  --no-color        Disable color output
  -V, --version     Show version information

EXAMPLES:
  trumpcards                     Start interactive mode (switch games with 'switch <game>')
  trumpcards blackjack           Play BlackJack
  trumpcards --lang en poker     Play Poker in English
  trumpcards update              Self-update to the latest version
  NO_COLOR=1 trumpcards hearts   Play Hearts without color output
  trumpcards web                 Start the web GUI server

ENVIRONMENT VARIABLES:
  NO_COLOR          Disable color output when set (see https://no-color.org/)
                    Example: NO_COLOR=1 trumpcards blackjack
  PORT              Port number for the web server (default: 8080)
                    Example: PORT=3000 trumpcards web
`

	lang := flag.String("lang", "", "language (ja or en)")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.BoolVar(showVersion, "V", false, "Show version information (shorthand)")
	noColorFlag := flag.Bool("no-color", false, "Disable color output")
	showHelp := flag.Bool("help", false, "Show this help message")
	flag.BoolVar(showHelp, "h", false, "Show this help message (shorthand)")
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, helpText)
	}
	flag.Parse()

	if *showHelp {
		_, _ = fmt.Fprint(os.Stdout, helpText)
		return 0
	}

	// Color control: NO_COLOR env var (https://no-color.org/), --no-color flag,
	// or non-TTY stdout (pipe/redirect auto-detection).
	if os.Getenv("NO_COLOR") != "" || *noColorFlag || !term.IsTerminal(int(os.Stdout.Fd())) {
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
	commands := map[string]func() int{
		"blackjack":   func() int { ui.NewBlackJackCui().Exec(); return 0 },
		"poker":       func() int { ui.NewPokerCui().Exec(); return 0 },
		"oldmaid":     func() int { ui.NewOldMaidCui().Exec(); return 0 },
		"daifugo":     func() int { ui.NewDaifugoCui().Exec(); return 0 },
		"sevens":      func() int { ui.NewSevensCui().Exec(); return 0 },
		"doubt":       func() int { ui.NewDoubtCui().Exec(); return 0 },
		"holdem":      func() int { ui.NewHoldemCui().Exec(); return 0 },
		"omaha":       func() int { ui.NewOmahaCui().Exec(); return 0 },
		"hearts":      func() int { ui.NewHeartsCui().Exec(); return 0 },
		"memory":      func() int { ui.NewMemoryCui().Exec(); return 0 },
		"klondike":    func() int { ui.NewKlondikeCui().Exec(); return 0 },
		"freecell":    func() int { ui.NewFreeCellCui().Exec(); return 0 },
		"baccarat":    func() int { ui.NewBaccaratCui().Exec(); return 0 },
		"spades":      func() int { ui.NewSpadesCui().Exec(); return 0 },
		"crazyeights": func() int { ui.NewCrazyEightsCui().Exec(); return 0 },
		"ginrummy":    func() int { ui.NewGinRummyCui().Exec(); return 0 },
		"update": func() int {
			updater := update.NewUpdater(version, os.Stdin, os.Stdout, os.Stderr)
			if err := updater.Exec(); err != nil {
				return 1
			}
			return 0
		},
		"web": func() int {
			infrastructure.InitLogger()
			w := web.NewTrumpCardsWeb()
			if err := w.Exec(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				return 1
			}
			return 0
		},
	}

	arg := strings.ToLower(flag.Arg(0))
	if handler, ok := commands[arg]; ok {
		return handler()
	}

	if arg != "" {
		fmt.Fprintf(os.Stderr, "Error: unknown game %q\n", arg)
		if suggestion := cuiutil.SuggestCommand(arg, mapKeys(commands), 2); suggestion != "" {
			fmt.Fprintf(os.Stderr, "\n  Did you mean %q?\n", suggestion)
		}
		fmt.Fprintln(os.Stderr)
		flag.Usage()
		return 2
	}

	// No argument: start interactive multi-game mode (defaults to blackjack).
	manager := ui.NewGameManager("blackjack")
	ui.RunInteractiveCuiLoop(manager)
	return 0
}

func mapKeys(m map[string]func() int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
