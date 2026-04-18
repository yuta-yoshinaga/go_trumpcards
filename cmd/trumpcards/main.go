package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"syscall"

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
	showVersionShort := flag.Bool("version-short", false, "Print version number only (machine-readable)")
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

	// Color control: NO_COLOR env var (https://no-color.org/) and --no-color flag
	// force color off on both streams. TTY auto-detection is done per stream so that
	// `game | tee log.txt` keeps stderr colors and `game 2> err.log` keeps stdout colors.
	forceOff := os.Getenv("NO_COLOR") != "" || *noColorFlag
	color.SetStdoutColor(!forceOff && term.IsTerminal(int(os.Stdout.Fd())))
	color.SetStderrColor(!forceOff && term.IsTerminal(int(os.Stderr.Fd())))

	if *showVersionShort {
		fmt.Println(version)
		return 0
	}
	if *showVersion {
		fmt.Printf("trumpcards %s (commit: %s, built: %s)\n", version, commit, date)
		return 0
	}

	// Language detection: --lang > LANG env > default "ja".
	// An explicit --lang with an unsupported value is a hard error (exit 2);
	// an unsupported LANG env value emits a one-line warning and falls back to "ja"
	// (suppress with TRUMPCARDS_QUIET=1).
	supportedLangs := map[string]bool{"ja": true, "en": true}
	detectedLang := "ja"
	var langEnvWarn string // deferred until SetLang is called so i18n resolves
	if envLang := os.Getenv("LANG"); envLang != "" {
		prefix := envLang
		if idx := strings.IndexAny(envLang, "_-."); idx >= 0 {
			prefix = envLang[:idx]
		}
		if supportedLangs[prefix] {
			detectedLang = prefix
		} else if os.Getenv("TRUMPCARDS_QUIET") == "" {
			langEnvWarn = envLang
		}
	}
	if *lang != "" {
		if !supportedLangs[*lang] {
			i18n.SetLang(detectedLang)
			fmt.Fprintln(os.Stderr, i18n.Tf("cliUnsupportedLang", "lang", *lang))
			fmt.Fprintln(os.Stderr, i18n.T("cliSupportedLangs"))
			return 2
		}
		detectedLang = *lang
	}
	i18n.SetLang(detectedLang)
	if langEnvWarn != "" {
		fmt.Fprintln(os.Stderr, i18n.Tf("cliLangEnvFallback", "lang", langEnvWarn))
	}
	// Build game commands from the registry (single source of truth).
	commands := buildGameCommands()
	commands["games"] = func() int {
		var short, aliases bool
		_, code, ok := parseSubFlags("games", func(f *flag.FlagSet) {
			f.BoolVar(&short, "short", false, "Print game names only")
			f.BoolVar(&aliases, "aliases", false, "Include aliases in output (with --short)")
		})
		if !ok {
			return code
		}
		if aliases && !short {
			fmt.Fprintln(os.Stderr, i18n.T("cliAliasesWithoutShort"))
		}
		// Build reverse alias map: canonical name -> sorted list of aliases.
		reverseAliases := make(map[string][]string)
		for alias, canonical := range ui.GameAliases {
			reverseAliases[canonical] = append(reverseAliases[canonical], alias)
		}
		for k := range reverseAliases {
			sort.Strings(reverseAliases[k])
		}
		descs := ui.GameDescriptions()
		for _, name := range ui.GameNames() {
			if short {
				fmt.Println(name)
				if aliases {
					for _, alias := range reverseAliases[name] {
						fmt.Println(alias)
					}
				}
			} else {
				line := fmt.Sprintf("  %-16s %s", name, descs[name])
				if aliasList := reverseAliases[name]; len(aliasList) > 0 {
					line += fmt.Sprintf("  [aliases: %s]", strings.Join(aliasList, ", "))
				}
				fmt.Println(line)
			}
		}
		return 0
	}
	commands["completion"] = func() int {
		return runCompletion(flag.Args()[1:])
	}
	commands["help"] = func() int {
		return runHelpCommand(flag.Args()[1:], helpText, os.Stdout, os.Stderr)
	}
	commands["update"] = func() int {
		var yes bool
		_, code, ok := parseSubFlags("update", func(f *flag.FlagSet) {
			f.BoolVar(&yes, "yes", false, "Skip confirmation prompt")
			f.BoolVar(&yes, "y", false, "Skip confirmation prompt (shorthand)")
		})
		if !ok {
			return code
		}
		updater := update.NewUpdater(version, os.Stdin, os.Stderr, os.Stderr)
		updater.SetAutoConfirm(yes)
		updater.SetProgressIsTTY(term.IsTerminal(int(os.Stderr.Fd())))
		if err := updater.Exec(); err != nil {
			return 1
		}
		return 0
	}
	commands["web"] = func() int {
		var port int
		var host string
		_, code, ok := parseSubFlags("web", func(f *flag.FlagSet) {
			f.IntVar(&port, "port", 0, "Port number for the web server (default: 8080)")
			f.IntVar(&port, "p", 0, "Port number for the web server (shorthand)")
			f.StringVar(&host, "host", "", "Bind address for the web server (default: 127.0.0.1; use 0.0.0.0 to expose)")
		})
		if !ok {
			return code
		}
		if port != 0 {
			if port < 1 || port > 65535 {
				fmt.Fprintln(os.Stderr, i18n.Tf("cliInvalidPort", "port", strconv.Itoa(port)))
				return 1
			}
			_ = os.Setenv("PORT", strconv.Itoa(port))
		}
		if host != "" {
			_ = os.Setenv("HOST", host)
		}
		infrastructure.InitLogger()
		w := web.NewTrumpCardsWeb()
		if err := w.Exec(); err != nil {
			fmt.Fprintln(os.Stderr, i18n.Tf("cliWebStartFailed", "err", err.Error()))
			if errors.Is(err, syscall.EADDRINUSE) {
				fmt.Fprintln(os.Stderr, i18n.T("cliWebPortInUseHint"))
			}
			return 1
		}
		return 0
	}

	// Commands that parse their own sub-flags; skip the extra-args warning for these.
	// parseSubFlags-based commands handle extra-args warnings internally.
	subFlagCommands := map[string]bool{"web": true, "completion": true, "games": true, "update": true, "help": true}

	arg := strings.ToLower(flag.Arg(0))
	// Resolve game name aliases (e.g., "gin" -> "ginrummy", "7stud" -> "sevencardstud").
	if canonical, ok := ui.GameAliases[arg]; ok {
		arg = canonical
	}
	if handler, ok := commands[arg]; ok {
		// `<game> --help` / `<game> -h`: Go's flag package stops parsing at the first
		// non-flag argument, so these trailing flags land in Args(). Intercept them
		// and print that game's help instead of launching the game. Subcommands in
		// subFlagCommands are handled by parseSubFlags (which catches flag.ErrHelp).
		if !subFlagCommands[arg] && hasHelpFlag(flag.Args()[1:]) {
			return runHelpCommand([]string{arg}, helpText, os.Stdout, os.Stderr)
		}
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
	startGame := "blackjack"
	fmt.Println(i18n.Tf("cliStartupBanner", "version", version))
	fmt.Println(i18n.Tf("cliStartupGame", "game", startGame))
	manager := ui.NewGameManager(startGame)
	ui.RunInteractiveCuiLoop(manager)
	return 0
}

// builtinSubcommandHelp maps non-game subcommand names to their Usage/Flags/Examples
// help text. Used by both `trumpcards help <cmd>` and `trumpcards <cmd> --help`.
var builtinSubcommandHelp = map[string][]string{
	"web": {
		"USAGE:",
		"  trumpcards web [--port PORT] [--host HOST]",
		"",
		"FLAGS:",
		"  -p, --port PORT   Port number (default: 8080; env PORT)",
		"      --host HOST   Bind address (default: 127.0.0.1; use 0.0.0.0 to expose; env HOST)",
		"",
		"EXAMPLES:",
		"  trumpcards web",
		"  trumpcards web --port 3000",
		"  trumpcards web --host 0.0.0.0",
		"  HOST=0.0.0.0 PORT=3000 trumpcards web",
	},
	"update": {
		"USAGE:",
		"  trumpcards update [--yes]",
		"",
		"FLAGS:",
		"  -y, --yes   Skip confirmation prompt (required for non-interactive stdin)",
		"",
		"EXAMPLES:",
		"  trumpcards update",
		"  trumpcards update --yes",
	},
	"completion": {
		"USAGE:",
		"  trumpcards completion <bash|zsh|fish>",
		"",
		"EXAMPLES:",
		"  source <(trumpcards completion bash)",
		"  trumpcards completion zsh > \"${fpath[1]}/_trumpcards\"",
		"  trumpcards completion fish > ~/.config/fish/completions/trumpcards.fish",
	},
	"games": {
		"USAGE:",
		"  trumpcards games [--short] [--aliases]",
		"",
		"FLAGS:",
		"      --short     Print game names only (for scripting)",
		"      --aliases   Include aliases (requires --short)",
		"",
		"EXAMPLES:",
		"  trumpcards games",
		"  trumpcards games --short",
		"  trumpcards games --short --aliases",
	},
	"help": {
		"USAGE:",
		"  trumpcards help [game|command]",
		"",
		"EXAMPLES:",
		"  trumpcards help",
		"  trumpcards help blackjack",
		"  trumpcards help web",
	},
}

// runHelpCommand implements the `trumpcards help [game]` subcommand.
// With no args, it writes helpText to stdout. With one arg, it writes the
// HelpLines() of the matching game (resolving aliases) to stdout, or an
// "unknown game" error with a Did-you-mean suggestion to stderr. Extra
// positional arguments after the game name are warned about and ignored,
// matching the behavior of other subcommands.
func runHelpCommand(args []string, helpText string, stdout, stderr io.Writer) int {
	if len(args) > 1 {
		_, _ = fmt.Fprintln(stderr, i18n.Tf("cliExtraArgsWarning", "args", strings.Join(args[1:], " ")))
	}
	if len(args) == 0 {
		_, _ = fmt.Fprint(stdout, helpText)
		return 0
	}
	target := strings.ToLower(args[0])
	if canonical, ok := ui.GameAliases[target]; ok {
		target = canonical
	}
	for _, entry := range ui.GameRegistry() {
		if entry.Name == target {
			g := entry.NewCui()
			for _, line := range g.HelpLines() {
				_, _ = fmt.Fprintln(stdout, line)
			}
			return 0
		}
	}
	if lines, ok := builtinSubcommandHelp[target]; ok {
		for _, line := range lines {
			_, _ = fmt.Fprintln(stdout, line)
		}
		return 0
	}
	_, _ = fmt.Fprintln(stderr, i18n.Tf("cliHelpUnknownGame", "name", target))
	if suggestion := cuiutil.SuggestCommand(target, ui.GameNames(), 2); suggestion != "" {
		_, _ = fmt.Fprintf(stderr, "  %s\n", i18n.Tf("didYouMean", "name", suggestion))
	}
	return 1
}

// parseSubFlags creates a FlagSet, applies setup, parses subcommand args, and
// warns about extra positional arguments. Returns (fs, exitCode, ok). If ok is
// false, the caller should return exitCode immediately.
func parseSubFlags(name string, setup func(*flag.FlagSet)) (*flag.FlagSet, int, bool) {
	return parseSubFlagsTo(name, flag.Args()[1:], setup, os.Stdout, os.Stderr)
}

// parseSubFlagsTo is the testable core of parseSubFlags: it parses args with a
// subcommand FlagSet, wires the shared builtin help text as the usage, and
// writes help/diagnostics to the given streams.
func parseSubFlagsTo(name string, args []string, setup func(*flag.FlagSet), stdout, stderr io.Writer) (*flag.FlagSet, int, bool) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard) // suppress Go's raw English error/usage text
	setup(fs)
	printHelp := func() {
		if lines, ok := builtinSubcommandHelp[name]; ok {
			for _, line := range lines {
				_, _ = fmt.Fprintln(stdout, line)
			}
		}
	}
	fs.Usage = printHelp
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printHelp()
			return nil, 0, false
		}
		_, _ = fmt.Fprintln(stderr, i18n.Tf("cliSubcommandFlagError", "cmd", name, "err", err.Error()))
		_, _ = fmt.Fprintln(stderr, i18n.Tf("cliTryHelp", "cmd", name))
		return nil, 2, false
	}
	if fs.NArg() > 0 {
		_, _ = fmt.Fprintln(stderr, i18n.Tf("cliExtraArgsWarning", "args", strings.Join(fs.Args(), " ")))
	}
	return fs, 0, true
}

// hasHelpFlag reports whether args contains a help flag. It accepts all four
// forms Go's flag package treats as equivalent for a help flag registered as
// both "help" and "h": "-h", "--h", "-help", "--help".
func hasHelpFlag(args []string) bool {
	for _, a := range args {
		switch a {
		case "-h", "--h", "-help", "--help":
			return true
		}
	}
	return false
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
  games        List all available games (--short for names only, --aliases to include aliases with --short)
  help [game]  Show this help, or a specific game's help text
  completion   Generate shell completion script (bash, zsh, fish)
  update       Self-update to the latest version
  web          Start REST API + web GUI server

  (no argument) Interactive mode with game switching

OPTIONS:
  -h, --help        Show this help message
  --lang ja|en      Language (default: ja)
  --no-color        Disable color output (stdout and stderr)
                    Auto-detection is per-stream: stdout color is on only
                    when stdout is a TTY; the same applies to stderr.
  -V, --version     Show version information
  --version-short   Print version number only (machine-readable)

EXAMPLES:
  trumpcards                     Start interactive mode (switch games with 'switch <game>')
  trumpcards blackjack           Play BlackJack
  trumpcards blackjack --help    Show BlackJack's in-game commands
  trumpcards --lang en poker     Play Poker in English
  trumpcards games               List all available games
  trumpcards games --short       List game names only (for scripting)
  trumpcards games --short --aliases  List game names including aliases
  trumpcards update              Self-update to the latest version
  trumpcards update --yes        Update without confirmation prompt
  trumpcards --version-short     Print just the version number (e.g. 1.2.3)
  NO_COLOR=1 trumpcards hearts   Play Hearts without color output
  trumpcards web                 Start the web GUI server (binds to 127.0.0.1)
  trumpcards web --port 3000     Start the web GUI on port 3000
  trumpcards web --host 0.0.0.0  Expose the web GUI on all interfaces
  source <(trumpcards completion bash)   Enable bash completion

ENVIRONMENT VARIABLES:
  NO_COLOR          Disable color output on both stdout and stderr when set
                    (see https://no-color.org/)
                    Example: NO_COLOR=1 trumpcards blackjack
  HOST              Bind address for the web server (default: 127.0.0.1)
                    Example: HOST=0.0.0.0 trumpcards web
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
