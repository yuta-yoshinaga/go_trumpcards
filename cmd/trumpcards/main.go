package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
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
	helpText := buildHelpText()

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
		fmt.Fprintln(os.Stderr, i18n.Tf("cliUnsupportedLang", "lang", detectedLang))
	}
	// Build game commands from the registry (single source of truth).
	commands := buildGameCommands()
	commands["games"] = func() int {
		gamesFlags := flag.NewFlagSet("games", flag.ContinueOnError)
		short := gamesFlags.Bool("short", false, "Print game names only")
		if err := gamesFlags.Parse(flag.Args()[1:]); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return 0
			}
			return 1
		}
		if gamesFlags.NArg() > 0 {
			fmt.Fprintln(os.Stderr, i18n.Tf("cliExtraArgsWarning", "args", strings.Join(gamesFlags.Args(), " ")))
		}
		// Build reverse alias map: canonical name -> sorted list of aliases.
		reverseAliases := make(map[string][]string)
		for alias, canonical := range ui.GameAliases {
			reverseAliases[canonical] = append(reverseAliases[canonical], alias)
		}
		for k := range reverseAliases {
			sort.Strings(reverseAliases[k])
		}
		for _, name := range ui.GameNames {
			if *short {
				fmt.Println(name)
			} else {
				line := fmt.Sprintf("  %-16s %s", name, ui.GameDescriptions[name])
				if aliases := reverseAliases[name]; len(aliases) > 0 {
					line += fmt.Sprintf("  [aliases: %s]", strings.Join(aliases, ", "))
				}
				fmt.Println(line)
			}
		}
		return 0
	}
	commands["completion"] = func() int {
		return runCompletion(flag.Args()[1:])
	}
	commands["update"] = func() int {
		updateFlags := flag.NewFlagSet("update", flag.ContinueOnError)
		yes := updateFlags.Bool("yes", false, "Skip confirmation prompt")
		updateFlags.BoolVar(yes, "y", false, "Skip confirmation prompt (shorthand)")
		if err := updateFlags.Parse(flag.Args()[1:]); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return 0
			}
			return 1
		}
		updater := update.NewUpdater(version, os.Stdin, os.Stderr, os.Stderr)
		updater.SetAutoConfirm(*yes)
		if err := updater.Exec(); err != nil {
			return 1
		}
		return 0
	}
	commands["web"] = func() int {
		webFlags := flag.NewFlagSet("web", flag.ContinueOnError)
		port := webFlags.Int("port", 0, "Port number for the web server (default: 8080)")
		webFlags.IntVar(port, "p", 0, "Port number for the web server (shorthand)")
		if err := webFlags.Parse(flag.Args()[1:]); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return 0
			}
			return 1
		}
		if webFlags.NArg() > 0 {
			fmt.Fprintln(os.Stderr, i18n.Tf("cliExtraArgsWarning", "args", strings.Join(webFlags.Args(), " ")))
		}
		if *port != 0 {
			if *port < 1 || *port > 65535 {
				fmt.Fprintln(os.Stderr, i18n.Tf("cliInvalidPort", "port", strconv.Itoa(*port)))
				return 1
			}
			_ = os.Setenv("PORT", strconv.Itoa(*port))
		}
		infrastructure.InitLogger()
		w := web.NewTrumpCardsWeb()
		if err := w.Exec(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		return 0
	}

	// Commands that parse their own sub-flags; skip the extra-args warning for these.
	subFlagCommands := map[string]bool{"web": true, "completion": true, "games": true, "update": true}

	arg := strings.ToLower(flag.Arg(0))
	// Resolve game name aliases (e.g., "gin" -> "ginrummy", "7stud" -> "sevencardstud").
	if canonical, ok := ui.GameAliases[arg]; ok {
		arg = canonical
	}
	if handler, ok := commands[arg]; ok {
		if flag.NArg() > 1 && !subFlagCommands[arg] {
			fmt.Fprintln(os.Stderr, i18n.Tf("cliExtraArgsWarning", "args", strings.Join(flag.Args()[1:], " ")))
		}
		return handler()
	}

	if arg != "" {
		fmt.Fprintln(os.Stderr, i18n.Tf("cliUnknownGame", "name", arg))
		if suggestion := cuiutil.SuggestCommand(arg, mapKeys(commands), 2); suggestion != "" {
			fmt.Fprintf(os.Stderr, "  %s\n", i18n.Tf("didYouMean", "name", suggestion))
		}
		fmt.Fprintln(os.Stderr)
		flag.Usage()
		return 1
	}

	// No argument: start interactive multi-game mode (defaults to blackjack).
	fmt.Println(i18n.Tf("cliStartupBanner", "version", version))
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

// buildHelpText generates the CLI help text with the games section derived from the registry.
func buildHelpText() string {
	var sb strings.Builder
	sb.WriteString(`USAGE:
  trumpcards [--lang ja|en] [game]
  trumpcards --help

GAMES:
`)
	for _, entry := range ui.GameRegistry() {
		fmt.Fprintf(&sb, "  %-16s %s\n", entry.Name, entry.Description)
	}
	sb.WriteString(`
COMMANDS:
  games        List all available games (--short for names only)
  completion   Generate shell completion script (bash, zsh, fish)
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
  trumpcards games               List all available games
  trumpcards games --short       List game names only (for scripting)
  trumpcards update              Self-update to the latest version
  trumpcards update --yes        Update without confirmation prompt
  NO_COLOR=1 trumpcards hearts   Play Hearts without color output
  trumpcards web                 Start the web GUI server
  trumpcards web --port 3000     Start the web GUI on port 3000
  source <(trumpcards completion bash)   Enable bash completion

ENVIRONMENT VARIABLES:
  NO_COLOR          Disable color output when set (see https://no-color.org/)
                    Example: NO_COLOR=1 trumpcards blackjack
  PORT              Port number for the web server (default: 8080)
                    Example: PORT=3000 trumpcards web
`)
	return sb.String()
}

// buildGameCommands generates command handlers for all games from the registry.
func buildGameCommands() map[string]func() int {
	registry := ui.GameRegistry()
	commands := make(map[string]func() int, len(registry)+4)
	for _, entry := range registry {
		e := entry // capture loop variable
		commands[e.Name] = func() int {
			g := e.NewCui()
			ui.RunCuiLoop(g.Controller(), g.HelpLines())
			return 0
		}
	}
	return commands
}
