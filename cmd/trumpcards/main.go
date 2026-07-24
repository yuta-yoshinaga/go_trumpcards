package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"slices"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/term"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games"
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

// supportedLangs is the single source of truth for which i18n locales the CLI
// accepts, shared by bootstrap detection (pre-Parse) and the post-Parse
// validator so the two cannot diverge.
var supportedLangs = map[string]bool{"ja": true, "en": true}

// posixLocalePlaceholders are POSIX placeholder locale prefixes that signal
// "no language preference" rather than a specific unsupported language.
// They are the default in Docker base images (debian-slim, distroless), CI
// runners (GitHub Actions, GitLab CI), and minimal Linux installs, so warning
// about them on every invocation creates noise without information value.
// Treat them as silent fallback to the default locale, matching gettext's
// convention. See issue #1534.
var posixLocalePlaceholders = map[string]bool{"C": true, "POSIX": true}

// isPosixLocalePlaceholder reports whether prefix is a POSIX locale
// placeholder (e.g. the "C" in "C.UTF-8"). Comparison is case-sensitive
// because real-world LANG values use the canonical uppercase forms ("C",
// "POSIX"); a lowercase "c" is unusual enough that it is more likely a typo
// for a real language code than the placeholder, so we let it warn through.
func isPosixLocalePlaceholder(prefix string) bool {
	return posixLocalePlaceholders[prefix]
}

// portInRange reports whether port is a valid TCP port number, with 0
// reserved for "let the OS assign an ephemeral port" (POSIX convention).
func portInRange(port int) bool {
	return port >= 0 && port <= 65535
}

// validHost reports whether s is usable as the host portion of a TCP bind
// address: a literal IP (v4/v6), an empty string, or a syntactically valid
// hostname. An empty host is accepted but treated by the server as "unset" —
// getListenAddr falls back to the 127.0.0.1 default, so `--host ""` is
// equivalent to omitting the flag; pass 0.0.0.0 to expose on all interfaces.
// We deliberately do NOT resolve DNS here — name-resolution failure at bind
// time stays a runtime error (exit 1); only syntactically impossible values
// are rejected as a usage error (exit 2). Whitespace and an embedded ":"
// (which would corrupt net.JoinHostPort, or smuggle in a port) are rejected
// up front. See #2150.
func validHost(s string) bool {
	if s == "" {
		return true
	}
	if strings.TrimSpace(s) != s || strings.ContainsAny(s, " \t") {
		return false
	}
	if net.ParseIP(s) != nil {
		return true
	}
	// Reject anything that would break net.JoinHostPort / contain a port.
	if strings.Contains(s, ":") {
		return false
	}
	return isValidHostname(s)
}

// isValidHostname reports whether s is a syntactically valid RFC 1123 hostname:
// 1-253 characters total, dot-separated labels of 1-63 characters each, where
// every label contains only ASCII letters, digits, or hyphens and does not
// start or end with a hyphen. A single trailing dot (fully-qualified form) is
// tolerated. This is a syntax check only — it does not attempt resolution.
func isValidHostname(s string) bool {
	s = strings.TrimSuffix(s, ".")
	if s == "" || len(s) > 253 {
		return false
	}
	for _, label := range strings.Split(s, ".") {
		if len(label) == 0 || len(label) > 63 {
			return false
		}
		if label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for i := 0; i < len(label); i++ {
			c := label[i]
			switch {
			case c >= 'a' && c <= 'z':
			case c >= 'A' && c <= 'Z':
			case c >= '0' && c <= '9':
			case c == '-':
			default:
				return false
			}
		}
	}
	return true
}

// applyColorMode resolves the tristate `--color=auto|always|never` flag (issue
// #1554), the legacy `--no-color` deprecated alias, and the NO_COLOR env var
// (https://no-color.org/) into per-stream color settings. Precedence:
//
//  1. NO_COLOR=<any non-empty value>     => never (POSIX-spec forces off)
//  2. --color=never  OR  --no-color      => never
//  3. --color=always                     => always
//  4. --color=auto (default) or unset    => per-stream TTY detect
//
// Returns (exitCode, ok). Returns (2, false) on an unrecognized --color value;
// the caller should propagate the exit code. The value comparison is
// case-insensitive and trims surrounding whitespace so users typing
// `--color=ALWAYS` or `--color= auto ` succeed without surprise. The error
// message is rendered via i18n on stderr.
func applyColorMode(mode string, noColorFlag bool, noColorEnv string, stdoutFd, stderrFd uintptr, stderr io.Writer) (int, bool) {
	resolved := strings.ToLower(strings.TrimSpace(mode))
	switch {
	case noColorEnv != "":
		resolved = "never"
	case noColorFlag:
		resolved = "never"
	case resolved == "":
		resolved = "auto"
	}
	switch resolved {
	case "always":
		color.SetStdoutColor(true)
		color.SetStderrColor(true)
	case "never":
		color.SetStdoutColor(false)
		color.SetStderrColor(false)
	case "auto":
		color.SetStdoutColor(term.IsTerminal(int(stdoutFd)))
		color.SetStderrColor(term.IsTerminal(int(stderrFd)))
	default:
		_, _ = fmt.Fprintln(stderr, i18n.Tf("cliInvalidColorMode", "mode", mode))
		return 2, false
	}
	return 0, true
}

// validCategoryNames is the canonical set of `--category` filter values,
// derived from the games registry at package load so it stays in lockstep
// with games.Category.String() automatically. Adding a new Category to the
// games package is enough — no manual sync here. See issue #1535 and
// PR #1538 review feedback.
var validCategoryNames = func() map[string]bool {
	seen := make(map[string]bool)
	for _, g := range games.All() {
		seen[g.Category.String()] = true
	}
	return seen
}()

// validCategory reports whether s names a registered game category. Comparison
// is case-sensitive to match games.Category.String()'s canonical lowercase form.
func validCategory(s string) bool {
	return validCategoryNames[s]
}

// flagSetVisited reports whether any of the named flags was explicitly set
// on fs. This lets the caller distinguish "user did not pass --port" from
// "user passed --port 0" without needing a sentinel value in the integer
// input space — which would misclassify e.g. `--port -1` as "unset".
func flagSetVisited(fs *flag.FlagSet, names ...string) bool {
	want := make(map[string]struct{}, len(names))
	for _, n := range names {
		want[n] = struct{}{}
	}
	seen := false
	fs.Visit(func(fl *flag.Flag) {
		if _, ok := want[fl.Name]; ok {
			seen = true
		}
	})
	return seen
}

func main() {
	os.Exit(run())
}

func run() int {
	helpText := buildHelpText()

	// Route parse errors through i18n. Default is ExitOnError with English
	// output to stderr; switch to ContinueOnError + a discarded sink so we
	// own error rendering. We also suppress FlagSet.Usage — the flag package
	// calls it internally before returning the error, which would print the
	// help text once before our handler prints it again.
	flag.CommandLine.Init(os.Args[0], flag.ContinueOnError)
	flag.CommandLine.SetOutput(io.Discard)
	flag.CommandLine.Usage = func() {}

	lang := flag.String("lang", "", "language (ja or en)")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.BoolVar(showVersion, "V", false, "Show version information (shorthand)")
	showVersionShort := flag.Bool("version-short", false, "Print version number only (machine-readable)")
	noColorFlag := flag.Bool("no-color", false, "Disable color output (deprecated alias for --color=never)")
	colorMode := flag.String("color", "auto", "Color output mode: auto (default; per-stream TTY detect), always, or never")
	quietFlag := flag.Bool("quiet", false, "Suppress non-essential output (warnings, banners). Errors still go to stderr.")
	flag.BoolVar(quietFlag, "q", false, "Suppress non-essential output (shorthand)")
	startGameFlag := flag.String("start", "", "Initial game for interactive mode (no positional arg). Aliases accepted; ignored when a game name is given. See issue #1604.")
	showHelp := flag.Bool("help", false, "Show this help message")
	flag.BoolVar(showHelp, "h", false, "Show this help message (shorthand)")

	// Resolve the locale from LANG env + any --lang in os.Args before Parse
	// so a flag error (e.g. --bogus) is rendered in the user's language,
	// not English.
	i18n.SetLang(detectBootstrapLang(os.Args[1:], os.Getenv("LANG")))

	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		// NB: -h / --help are registered above so flag.Parse handles them
		// itself and returns nil; flag.ErrHelp is therefore unreachable here.
		_, _ = fmt.Fprintln(os.Stderr, i18n.Tf("cliFlagError", "err", err.Error()))
		_, _ = fmt.Fprint(os.Stderr, helpText)
		return 2
	}

	if *showHelp {
		_, _ = fmt.Fprint(os.Stdout, helpText)
		return 0
	}

	// Color control: tristate --color=auto|always|never (issue #1554) plus the
	// legacy --no-color flag (kept as a deprecated alias for --color=never)
	// and the NO_COLOR env var (https://no-color.org/, which the spec defines
	// as "presence => disable" and therefore overrides --color=always).
	// Precedence: NO_COLOR > --color=never (or --no-color) > --color=always >
	// --color=auto (per-stream TTY detect, the historical default).
	if code, ok := applyColorMode(*colorMode, *noColorFlag, os.Getenv("NO_COLOR"), os.Stdout.Fd(), os.Stderr.Fd(), os.Stderr); !ok {
		return code
	}

	if *showVersionShort {
		fmt.Println(version)
		return 0
	}
	if *showVersion {
		fmt.Printf("trumpcards %s (commit: %s, built: %s)\n", version, commit, date)
		return 0
	}

	// Language detection: --lang > LANG env > default "ja".
	// Both --lang and LANG env with unsupported values emit a one-line warning
	// on stderr and fall back to the detected/default locale (suppress with
	// TRUMPCARDS_QUIET=1). Aligning --lang with LANG env behavior matches the
	// "Warning: ... defaulting to ja" message users already see, and prevents
	// typos from tripping `set -e` scripts. See issue #1448.
	detectedLang := "ja"
	// Quiet is the OR of the global --quiet/-q flag and TRUMPCARDS_QUIET env var
	// (issue #1553). The env var was the only knob before; the flag was added
	// so users get the POSIX-conventional `-q` they expect from `grep`/`apt-get`/
	// etc., and so `web --quiet` no longer needs to be the only opt-out path.
	quiet := *quietFlag || os.Getenv("TRUMPCARDS_QUIET") != ""
	var langEnvWarn string // deferred until SetLang is called so i18n resolves
	if envLang := os.Getenv("LANG"); envLang != "" {
		prefix := envLang
		if idx := strings.IndexAny(envLang, "_-."); idx >= 0 {
			prefix = envLang[:idx]
		}
		switch {
		case supportedLangs[prefix]:
			detectedLang = prefix
		case isPosixLocalePlaceholder(prefix):
			// C / POSIX / C.UTF-8 are placeholders, not language preferences;
			// silently fall back to the default. See issue #1534.
		case !quiet:
			langEnvWarn = envLang
		}
	}
	var langFlagWarn string
	if *lang != "" {
		if !supportedLangs[*lang] {
			if !quiet {
				langFlagWarn = *lang
			}
		} else {
			detectedLang = *lang
		}
	}
	i18n.SetLang(detectedLang)
	if langEnvWarn != "" {
		fmt.Fprintln(os.Stderr, i18n.Tf("cliLangEnvFallback", "lang", langEnvWarn))
	}
	if langFlagWarn != "" {
		fmt.Fprintln(os.Stderr, i18n.Tf("cliUnsupportedLang", "lang", langFlagWarn))
		fmt.Fprintln(os.Stderr, i18n.T("cliSupportedLangs"))
	}
	// Build game commands from the registry (single source of truth).
	commands := buildGameCommands()
	commands["games"] = func() int {
		var short, aliases, asJSON bool
		var category string
		// quietSink absorbs `-q`/`--quiet` when placed after the subcommand
		// name so Go's FlagSet does not reject it as an unknown flag. The
		// outer `quiet` was already resolved before subcommand dispatch
		// and is the source of truth for behavior; this binding only
		// exists so subcommand parsing does not exit 2 on a recognized
		// global flag (PR #1582 review).
		var quietSink bool
		fs, code, ok := parseSubFlags("games", func(f *flag.FlagSet) {
			f.BoolVar(&short, "short", false, "Print game names only")
			f.BoolVar(&aliases, "aliases", false, "With --short, also print each alias on its own line (long output always includes aliases inline)")
			f.BoolVar(&asJSON, "json", false, "Emit machine-readable JSON (array of {name, category, description, aliases})")
			f.StringVar(&category, "category", "", "Filter by category: casino|classic|solo")
			f.BoolVar(&quietSink, "quiet", quiet, "Accepted for consistency with the global flag; the global -q already applied")
			f.BoolVar(&quietSink, "q", quiet, "Accepted for consistency with the global flag; the global -q already applied (shorthand)")
		})
		if !ok {
			return code
		}
		if category != "" && !validCategory(category) {
			fmt.Fprintln(os.Stderr, i18n.Tf("cliInvalidCategory", "category", category))
			return 2
		}
		if asJSON {
			// --short / --aliases are silently irrelevant in JSON mode (the
			// schema is fixed). Warn the user instead of dropping the flags
			// without feedback so a typo in scripts surfaces early. The warning
			// goes to stderr so the JSON on stdout stays pure / pipeable. See
			// PR #1538 review feedback.
			if flagSetVisited(fs, "short", "aliases") {
				fmt.Fprintln(os.Stderr, i18n.T("cliGamesJSONIgnoredFlags"))
			}
			if err := printGamesJSON(category, os.Stdout); err != nil {
				fmt.Fprintln(os.Stderr, i18n.Tf("cliGamesJSONError", "err", err.Error()))
				return 1
			}
			return 0
		}
		printGames(short, aliases, category, os.Stdout)
		return 0
	}
	commands["completion"] = func() int {
		var noHint bool
		var quietSink bool // see comment on the games subcommand
		fs, code, ok := parseSubFlags("completion", func(f *flag.FlagSet) {
			f.BoolVar(&noHint, "no-hint", false, "Suppress installation hint comments (also implied when stdout is not a TTY)")
			f.BoolVar(&quietSink, "quiet", quiet, "Accepted for consistency with the global flag; the global -q already applied")
			f.BoolVar(&quietSink, "q", quiet, "Accepted for consistency with the global flag; the global -q already applied (shorthand)")
		})
		if !ok {
			return code
		}
		stdoutIsTTY := term.IsTerminal(int(os.Stdout.Fd()))
		return runCompletion(fs.Args(), stdoutIsTTY, noHint)
	}
	commands["help"] = func() int {
		return runHelpCommand(flag.Args()[1:], helpText, os.Stdout, os.Stderr)
	}
	commands["version"] = func() int {
		var short bool
		_, code, ok := parseSubFlags("version", func(f *flag.FlagSet) {
			f.BoolVar(&short, "short", false, "Print version number only (machine-readable)")
		})
		if !ok {
			return code
		}
		if short {
			fmt.Println(version)
			return 0
		}
		fmt.Printf("trumpcards %s (commit: %s, built: %s)\n", version, commit, date)
		return 0
	}
	commands["update"] = func() int {
		var yes, check bool
		var quietSink bool // see comment on the games subcommand
		_, code, ok := parseSubFlags("update", func(f *flag.FlagSet) {
			f.BoolVar(&yes, "yes", false, "Skip confirmation prompt")
			f.BoolVar(&yes, "y", false, "Skip confirmation prompt (shorthand)")
			f.BoolVar(&check, "check", false, "Check for an update without installing (prints latest tag, status, current to stdout)")
			f.BoolVar(&check, "dry-run", false, "Alias for --check")
			f.BoolVar(&quietSink, "quiet", quiet, "Accepted for consistency with the global flag; the global -q already applied")
			f.BoolVar(&quietSink, "q", quiet, "Accepted for consistency with the global flag; the global -q already applied (shorthand)")
		})
		if !ok {
			return code
		}
		// --check writes machine-readable output to stdout (for CI / scripts)
		// while keeping human-friendly text on stderr. Install mode keeps its
		// legacy "writer=stderr" so `trumpcards update --yes 2>/dev/null` still
		// silences progress without swallowing the exit code. See issue #1484.
		var writer io.Writer = os.Stderr
		if check {
			writer = os.Stdout
		}
		updater := update.NewUpdater(version, os.Stdin, writer, os.Stderr)
		updater.SetAutoConfirm(yes)
		updater.SetReportCancelledAsError(true) // report user cancel as exit 75
		updater.SetCheckOnly(check)
		updater.SetProgressIsTTY(term.IsTerminal(int(os.Stderr.Fd())))
		updater.SetQuiet(quiet) // suppress the "checking latest release..." line under --quiet
		if err := updater.Exec(); err != nil {
			return updateExitCode(err)
		}
		return 0
	}
	commands["web"] = func() int {
		var port int
		var host string
		var openBrowser bool
		webQuiet := quiet // inherit TRUMPCARDS_QUIET env var as default
		fs, code, ok := parseSubFlags("web", func(f *flag.FlagSet) {
			f.IntVar(&port, "port", 0, "Port number for the web server (default: 8080; 0 for OS-assigned ephemeral)")
			f.IntVar(&port, "p", 0, "Port number for the web server (shorthand)")
			f.StringVar(&host, "host", "", "Bind address for the web server (default: 127.0.0.1; use 0.0.0.0 to expose)")
			f.BoolVar(&webQuiet, "quiet", webQuiet, "Suppress human-friendly startup/shutdown messages (structured slog still emitted)")
			f.BoolVar(&webQuiet, "q", webQuiet, "Suppress human-friendly startup/shutdown messages (shorthand)")
			f.BoolVar(&openBrowser, "open", false, "Open the resolved URL in the default browser after the server is ready (silently skipped when stderr is not a TTY)")
			f.BoolVar(&openBrowser, "o", false, "Open the resolved URL in the default browser (shorthand)")
		})
		if !ok {
			return code
		}
		if flagSetVisited(fs, "port", "p") {
			if !portInRange(port) {
				// Usage error (out-of-range value), not a runtime failure.
				// Aligns with the documented EXIT CODES: 2 reads as "the
				// invocation itself is wrong" so systemd / CI can branch on it.
				fmt.Fprintln(os.Stderr, i18n.Tf("cliInvalidPort", "port", strconv.Itoa(port)))
				fmt.Fprintln(os.Stderr, i18n.Tf("cliTryHelp", "cmd", "web"))
				return 2
			}
			_ = os.Setenv("PORT", strconv.Itoa(port))
		}
		if flagSetVisited(fs, "host") {
			if !validHost(host) {
				// Usage error (malformed bind address), not a runtime
				// failure — mirrors --port so systemd / CI can branch on
				// exit 2. DNS-resolvable-but-down names stay exit 1 at bind.
				fmt.Fprintln(os.Stderr, i18n.Tf("cliInvalidHost", "host", host))
				fmt.Fprintln(os.Stderr, i18n.Tf("cliTryHelp", "cmd", "web"))
				return 2
			}
			if host != "" {
				_ = os.Setenv("HOST", host)
			}
		}
		// Default to quiet when stderr is not a TTY (systemd, docker, pipe).
		// See issue #1452: human-friendly text is noise for log shippers, but
		// slog always fires so the lifecycle is still observable.
		if !webQuiet && !term.IsTerminal(int(os.Stderr.Fd())) {
			webQuiet = true
		}
		infrastructure.InitLogger()
		w := web.NewTrumpCardsWeb()
		w.SetQuiet(webQuiet)
		if openBrowser {
			// Skip silently in non-TTY contexts (SSH/Docker/CI) so a script
			// that always passes --open does not hang or spew xdg-open errors.
			// The hint goes to stderr only; never to stdout. See issue #1607.
			stderrIsTTY := term.IsTerminal(int(os.Stderr.Fd()))
			if !stderrIsTTY {
				if !webQuiet {
					fmt.Fprintln(os.Stderr, i18n.T("cliWebOpenSkippedNonTTY"))
				}
			} else {
				w.SetOnReady(func(boundAddr string) {
					url := web.BrowserURLFor(boundAddr)
					if url == "" {
						return
					}
					if err := web.OpenBrowser(url); err != nil && !webQuiet {
						fmt.Fprintln(os.Stderr, i18n.Tf("cliWebOpenFailed", "err", err.Error()))
					}
				})
			}
		}
		// Track which signal (if any) triggers the server's shutdown so we can
		// return 130 (SIGINT) or 143 (SIGTERM) rather than always returning 0.
		// Both this channel and TrumpCardsWeb's internal signal.NotifyContext
		// receive the signal concurrently; TrumpCardsWeb handles graceful
		// shutdown while this channel records which signal was received.
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		if err := w.Exec(); err != nil {
			signal.Stop(sigCh)
			fmt.Fprintln(os.Stderr, i18n.Tf("cliWebStartFailed", "err", err.Error()))
			if errors.Is(err, syscall.EADDRINUSE) {
				fmt.Fprintln(os.Stderr, i18n.T("cliWebPortInUseHint"))
			}
			return 1
		}
		signal.Stop(sigCh)
		select {
		case sig := <-sigCh:
			if sig == syscall.SIGTERM {
				return 143
			}
			return 130
		default:
			return 0
		}
	}

	// Commands that parse their own sub-flags; skip the extra-args warning for these.
	// parseSubFlags-based commands handle extra-args warnings internally.
	subFlagCommands := map[string]bool{"web": true, "completion": true, "games": true, "update": true, "help": true, "version": true}

	arg := strings.ToLower(flag.Arg(0))
	// Resolve game name aliases (e.g., "gin" -> "ginrummy", "7stud" -> "sevencardstud").
	if canonical, ok := ui.GameAliases[arg]; ok {
		arg = canonical
	}
	if handler, ok := commands[arg]; ok {
		if !subFlagCommands[arg] {
			// Apply --lang / --no-color / -q when they land after the game name
			// so `trumpcards <game> --lang en` and `trumpcards <game> -q`
			// match the prepositional form. Done before the help short-circuit
			// so `<game> --lang en --help` renders help in the requested locale.
			// quiet is passed by pointer because a trailing -q must update the
			// caller's value (otherwise the extras warning below would still
			// fire after the user asked for silence).
			extras := applyTrailingGlobalFlags(flag.Args()[1:], &quiet, os.Stderr)
			// `<game> --help` / `<game> -h`: Go's flag package stops parsing at
			// the first non-flag argument, so these trailing flags land in Args().
			// Intercept them and print that game's help instead of launching the
			// game. Subcommands in subFlagCommands are handled by parseSubFlags
			// (which catches flag.ErrHelp).
			if hasHelpFlag(extras) {
				return runHelpCommand([]string{arg}, helpText, os.Stdout, os.Stderr)
			}
			if len(extras) > 0 && !quiet {
				fmt.Fprintln(os.Stderr, i18n.Tf("cliExtraArgsWarning", "args", strings.Join(extras, " ")))
			}
		}
		return handler()
	}

	if arg != "" {
		fmt.Fprintln(os.Stderr, i18n.Tf("cliUnknownGame", "name", arg))
		if suggestion := cuiutil.SuggestCommand(arg, suggestionCandidates(commands), 2); suggestion != "" {
			fmt.Fprintf(os.Stderr, "  %s\n", i18n.Tf("didYouMean", "name", suggestion))
		}
		// A one-line recovery hint instead of re-dumping the full help (which
		// buries the error). Note flag.Usage is a no-op here (SetOutput was
		// pointed at io.Discard above), so the old flag.Usage() call rendered
		// nothing but a stray blank line. Exit 2 (usage error) to match the
		// `--start <unknown>` path (resolveStartGame) and the EXIT CODES table.
		fmt.Fprintln(os.Stderr, i18n.T("cliUnknownGameHint"))
		return 2
	}

	// No argument: start interactive multi-game mode (defaults to blackjack).
	// Startup banner goes to stderr so it matches `trumpcards web` (which also
	// uses stderr for info logs) and leaves stdout free for future
	// machine-readable output. See issue #1451.
	//
	// --start <game> overrides the default starting game (issue #1604). The
	// value goes through alias resolution and validation; an unknown name is
	// a usage error (exit 2) so a typo from a script fails loudly rather
	// than silently defaulting to blackjack. The flag is silently ignored
	// when a positional game arg is given (the early dispatch above already
	// returned for that path), keeping the no-effect-when-game-given
	// behaviour documented in the issue.
	startGame, code, ok := resolveStartGame(*startGameFlag, os.Stderr)
	if !ok {
		return code
	}
	if !quiet {
		fmt.Fprintln(os.Stderr, i18n.Tf("cliStartupBanner", "version", version))
		fmt.Fprintln(os.Stderr, i18n.Tf("cliStartupGame", "game", startGame))
	}
	manager := ui.NewGameManager(startGame)
	return ui.RunInteractiveCuiLoop(manager)
}

// resolveStartGame resolves the --start flag value into a canonical game
// name. Empty string returns the legacy default ("blackjack"). Aliases are
// resolved via ui.GameAliases. An unknown name writes a localized error to
// stderr and returns exit code 2 (usage error) — see the documented EXIT
// CODES table in buildHelpText. Issue #1604.
func resolveStartGame(flagValue string, stderr io.Writer) (string, int, bool) {
	const defaultStart = "blackjack"
	v := strings.ToLower(strings.TrimSpace(flagValue))
	if v == "" {
		return defaultStart, 0, true
	}
	if canonical, ok := ui.GameAliases[v]; ok {
		v = canonical
	}
	if slices.Contains(ui.GameNames(), v) {
		return v, 0, true
	}
	_, _ = fmt.Fprintln(stderr, i18n.Tf("cliUnknownGame", "name", v))
	if suggestion := cuiutil.SuggestCommand(v, helpSuggestionCandidates(), 2); suggestion != "" {
		_, _ = fmt.Fprintf(stderr, "  %s\n", i18n.Tf("didYouMean", "name", suggestion))
	}
	return "", 2, false
}

// builtinSubcommandHelp maps non-game subcommand names to their Usage/Flags/Examples
// help text. Used by both `trumpcards help <cmd>` and `trumpcards <cmd> --help`.
var builtinSubcommandHelp = map[string][]string{
	"web": {
		"USAGE:",
		"  trumpcards web [--port PORT] [--host HOST] [--quiet] [--open]",
		"",
		"FLAGS:",
		"  -p, --port PORT   Port number (default: 8080; 0 = OS-assigned ephemeral; env PORT)",
		"      --host HOST   Bind address (default: 127.0.0.1; use 0.0.0.0 to expose; env HOST)",
		"  -q, --quiet       Suppress human-friendly startup/shutdown messages",
		"                    (implied when stderr is not a TTY; structured slog logs still emitted)",
		"  -o, --open        Open the resolved URL in the default browser after the server is",
		"                    ready. Uses xdg-open / open / cmd-start. Silently skipped when",
		"                    stderr is not a TTY (SSH/Docker/CI). Issue #1607.",
		"",
		"NETWORK EXPOSURE:",
		"  Binding to a non-loopback host (0.0.0.0, ::, a LAN IP, etc.) prints a one-line",
		"  WARNING on stderr after the listening banner so the operator notices that the",
		"  server is reachable from other machines on the LAN/VPN. The warning is suppressed",
		"  by --quiet / TRUMPCARDS_QUIET=1, and slog's structured 'web server listening'",
		"  event always fires regardless. The default 127.0.0.1 bind never warns.",
		"",
		"EXIT CODES:",
		"    0  Server exited cleanly without receiving a signal",
		"    1  Server start failure (includes EADDRINUSE / port already in use —",
		"       systemd / Docker `restart: on-failure` should retry on this code)",
		"    2  Usage error (unknown flag, --port out of range, or invalid --host)",
		"  130  Graceful shutdown triggered by SIGINT (Ctrl+C)  (128 + 2)",
		"  143  Graceful shutdown triggered by SIGTERM  (128 + 15)",
		"",
		"EXAMPLES:",
		"  trumpcards web",
		"  trumpcards web --port 3000",
		"  trumpcards web --port 0            # start on any free port; see startup log for actual port",
		"  trumpcards web --open              # start on default port and open the browser",
		"  trumpcards web --port 0 --open     # ephemeral port + auto-open (no need to read the log)",
		"  trumpcards web --host 0.0.0.0      # exposes on all interfaces; prints exposure warning",
		"  trumpcards web --quiet             # systemd / docker friendly (no exposure warning either)",
		"  HOST=0.0.0.0 PORT=3000 trumpcards web",
	},
	"update": {
		"USAGE:",
		"  trumpcards update [--yes]",
		"  trumpcards update --check",
		"",
		"FLAGS:",
		"  -y, --yes     Skip confirmation prompt (required for non-interactive stdin)",
		"      --check   Check for an update without installing. Writes a tab-separated",
		"                summary '<latest-tag>\\t<status>\\t<current>' (status: latest |",
		"                available | dev) to stdout, and a human-friendly message to",
		"                stderr. Never prompts, never downloads. Safe in CI / cron.",
		"      --dry-run Alias for --check.",
		"",
		"EXIT CODES:",
		"   0  Success (already latest or updated; --check: already latest or dev build)",
		"   2  Usage error (non-interactive without --yes)",
		"   3  Network / release API failure",
		"   4  No binary for this platform",
		"   5  Archive extraction failure",
		"   6  Binary apply failure (permissions, disk, integrity)",
		"  10  --check: a newer version is available (non-error signal for scripts)",
		"  75  User declined the prompt",
		"   1  Unexpected error",
		"",
		"EXAMPLES:",
		"  trumpcards update",
		"  trumpcards update --yes",
		"  trumpcards update --check                   # prints latest tag to stdout; exits 10 when newer",
		"  trumpcards update --check | cut -f1          # latest tag only (e.g. v2.3.1)",
		"  if trumpcards update --check; then           # exit 0 means already latest",
		"    echo 'up to date'; fi",
	},
	"completion": {
		"USAGE:",
		"  trumpcards completion <bash|zsh|fish> [--no-hint]",
		"",
		"FLAGS:",
		"      --no-hint   Suppress installation hint comments in the output",
		"                  (implied when stdout is not a TTY, e.g. shell redirection)",
		"",
		"EXIT CODES:",
		"  0  Script written successfully",
		"  1  Write error (e.g. broken pipe, full disk)",
		"  2  Usage error (missing shell argument or unsupported shell name)",
		"",
		"EXAMPLES:",
		"  source <(trumpcards completion bash)",
		"  trumpcards completion bash > /etc/bash_completion.d/trumpcards",
		"  trumpcards completion zsh > \"${fpath[1]}/_trumpcards\"",
		"  trumpcards completion fish > ~/.config/fish/completions/trumpcards.fish",
	},
	"games": {
		"USAGE:",
		"  trumpcards games [--short] [--aliases] [--json] [--category casino|classic|solo]",
		"",
		"FLAGS:",
		"      --short              Print game names only (for scripting)",
		"      --aliases            With --short, also print each alias on its own line",
		"                           (long output always includes aliases inline)",
		"      --json               Emit machine-readable JSON: an array of",
		"                           {name, category, description, aliases}.",
		"                           aliases is always [] (never null) for stable schema.",
		"                           Combines with --category; --short / --aliases have no",
		"                           effect in JSON mode (the schema is fixed) and emit a",
		"                           one-line warning to stderr if used together.",
		"      --category CAT       Restrict output to one Cloudflare Worker category:",
		"                           casino, classic, or solo. Combinable with --short / --json.",
		"                           Invalid value exits 2.",
		"",
		"EXIT CODES:",
		"  0  List printed successfully",
		"  1  --json: encoding error while writing the JSON array",
		"  2  Usage error (unknown flag, invalid --category value)",
		"",
		"EXAMPLES:",
		"  trumpcards games",
		"  trumpcards games --short",
		"  trumpcards games --short --aliases",
		"  trumpcards games --category casino",
		"  trumpcards games --json | jq -r '.[] | select(.category==\"solo\") | .name'",
		"  trumpcards games --json --category classic",
	},
	"help": {
		"USAGE:",
		"  trumpcards help [game|command]",
		"",
		"EXIT CODES:",
		"  0  Help printed (top-level, a known game, or a known subcommand)",
		"  1  Unknown game / command name (a 'did you mean…' hint is offered when close)",
		"",
		"EXAMPLES:",
		"  trumpcards help",
		"  trumpcards help blackjack",
		"  trumpcards help web",
	},
	"version": {
		"USAGE:",
		"  trumpcards version [--short]",
		"",
		"FLAGS:",
		"      --short   Print the version number only (machine-readable)",
		"",
		"EXIT CODES:",
		"  0  Version printed successfully",
		"  2  Usage error (unknown flag)",
		"",
		"EXAMPLES:",
		"  trumpcards version              # full info: trumpcards <ver> (commit: <sha>, built: <date>)",
		"  trumpcards version --short      # just the version (e.g. 1.2.3)",
		"",
		"NOTES:",
		"  Equivalent to the global --version / -V flag (and --version-short for --short).",
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
	if suggestion := cuiutil.SuggestCommand(target, helpSuggestionCandidates(), 2); suggestion != "" {
		_, _ = fmt.Fprintf(stderr, "  %s\n", i18n.Tf("didYouMean", "name", suggestion))
	}
	return 1
}

// helpSuggestionCandidates returns the deduplicated set of canonical game
// names, builtin subcommand names, and their game aliases. Used by
// `runHelpCommand` so that `trumpcards help <typo>` recovers a likely
// alias (e.g. `gni` -> `gin`) instead of a far-off canonical name. See
// issue #1555.
func helpSuggestionCandidates() []string {
	capacity := len(ui.GameNames()) + len(builtinSubcommandHelp) + len(ui.GameAliases)
	seen := make(map[string]struct{}, capacity)
	out := make([]string, 0, capacity)
	add := func(name string) {
		if _, ok := seen[name]; ok {
			return
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	for _, name := range ui.GameNames() {
		add(name)
	}
	for name := range builtinSubcommandHelp {
		add(name)
	}
	for alias := range ui.GameAliases {
		add(alias)
	}
	return out
}

// suggestionCandidates returns the deduplicated set of registered top-level
// commands (canonical game names plus builtin subcommands such as `web` or
// `update`) and game aliases (e.g. `gin`, `7stud`) for "did you mean"
// suggestions on unknown game names. Aliases are useful targets here because
// users sometimes typo the alias, not the canonical, and a canonical-only
// candidate list returns nonsense (`gni` -> `gofish` instead of `gin`).
// See issue #1555.
func suggestionCandidates(commands map[string]func() int) []string {
	seen := make(map[string]struct{}, len(commands)+len(ui.GameAliases))
	out := make([]string, 0, len(commands)+len(ui.GameAliases))
	add := func(name string) {
		if _, ok := seen[name]; ok {
			return
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	for k := range commands {
		add(k)
	}
	for alias := range ui.GameAliases {
		add(alias)
	}
	return out
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

// updateExitCode maps a non-nil error from update.Updater.Exec() to a CLI
// exit code. The mapping follows POSIX / BSD sysexits.h conventions where
// practical so automation (CI, cron, package scripts) can distinguish
// retryable failures (network) from terminal ones (no asset, apply failed).
// Returned codes:
//
//	2  — usage error: non-interactive stdin without --yes
//	3  — network / release API failure (retry later)
//	4  — no binary for this platform (no retry without a new release)
//	5  — archive extraction failure
//	6  — binary apply failure (permissions, disk, integrity)
//	10 — --check: a newer release is available (non-error signal)
//	75 — user declined the prompt (EX_TEMPFAIL)
//	1  — any other unexpected error
//
// See issues #1449, #1484.
func updateExitCode(err error) int {
	switch {
	case errors.Is(err, update.ErrNonInteractive):
		return 2
	case errors.Is(err, update.ErrNetwork), errors.Is(err, update.ErrAPIStatus):
		return 3
	case errors.Is(err, update.ErrNoAsset):
		return 4
	case errors.Is(err, update.ErrExtract):
		return 5
	case errors.Is(err, update.ErrApply):
		return 6
	case errors.Is(err, update.ErrUpdateAvailable):
		return 10
	case errors.Is(err, update.ErrUserCancelled):
		return 75
	default:
		return 1
	}
}

// trailingFlag classifies one trailing-arg token against a flag named `name`,
// accepting both the single- and double-dash spellings Go's flag package
// treats as equivalent (`-name` and `--name`). It reports:
//
//   - matched:   the token is this flag, in either bare or `=value` form
//   - hasInline: the token carried an inline `=value` (`--name=v` / `-name=v`)
//   - inlineVal: that value (meaningful only when hasInline)
//   - bare:      the token was exactly `--name` / `-name` (a separate value,
//     if the flag takes one, comes from the following arg)
//
// Extracted so the trailing-flag scanner no longer hand-repeats the
// dash/equals quartet for every flag (issue #2100).
func trailingFlag(arg, name string) (inlineVal string, hasInline, bare, matched bool) {
	if arg == "--"+name || arg == "-"+name {
		return "", false, true, true
	}
	if v, ok := strings.CutPrefix(arg, "--"+name+"="); ok {
		return v, true, false, true
	}
	if v, ok := strings.CutPrefix(arg, "-"+name+"="); ok {
		return v, true, false, true
	}
	return "", false, false, false
}

// trailingFlagAny is trailingFlag against several alias spellings of the same
// flag (e.g. the long "quiet" and short "q"), returning the first match.
func trailingFlagAny(arg string, names ...string) (inlineVal string, hasInline, bare, matched bool) {
	for _, n := range names {
		if v, hi, b, ok := trailingFlag(arg, n); ok {
			return v, hi, b, true
		}
	}
	return "", false, false, false
}

// applyTrailingGlobalFlags scans args for global flags (`--lang`,
// `--no-color`, `--color`, `--quiet`/`-q`) that landed after the game
// name because Go's flag package stops parsing at the first positional
// argument. Each recognized flag is applied to the runtime (i18n locale
// / color streams) and stripped from the returned slice so the caller's
// "extra args" warning fires only for genuinely unknown trailing tokens.
// Unsupported `--lang` values fall back to the existing locale and emit
// the usual cliUnsupportedLang warning on stderr (suppressed when quiet).
//
// Color resolution is deferred to the end of the scan and dispatched
// through `applyColorMode` (the SSoT) so the precedence rule matches the
// top-level flag exactly: --no-color (or --color=never) beats
// --color=always regardless of token order, and NO_COLOR env beats
// everything (PR #1583 review). An invalid trailing --color value emits
// the localized warning but does NOT abort the launched session — the
// ambient state is already valid and a late typo shouldn't kill a game
// that's about to run; applyColorMode's exit code is therefore
// intentionally discarded here.
//
// `--quiet`/`-q` writes through quietPtr so trailing position has the same
// effect as the leading position — Go's `flag` package stops parsing at the
// first positional, so the global pass missed any -q after the game name.
// Writing through a pointer keeps --lang/--color/-q symmetric in their
// "trailing flags also apply" contract without forcing a struct return.
//
// quiet is resolved in a pre-pass so warning suppression is order-independent:
// `<game> --lang xyz -q` and `<game> -q --lang xyz` both suppress the
// cliUnsupportedLang warning. Without the pre-pass, the single-pass
// implementation would only suppress when -q appeared first.
func applyTrailingGlobalFlags(args []string, quietPtr *bool, stderr io.Writer) []string {
	quiet := resolveTrailingQuiet(args, *quietPtr)
	*quietPtr = quiet
	rest := make([]string, 0, len(args))
	// Accumulate color flags across the entire scan so precedence matches
	// applyColorMode's documented order rather than depending on which
	// flag the user wrote last. trailingNoColor stays false unless
	// --no-color is present (or --no-color=true); --no-color=false is
	// consumed silently with no effect.
	var trailingNoColor, haveTrailingColor bool
	var trailingColor string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			rest = append(rest, args[i:]...)
			break
		}

		// --lang / -lang [=value | <next arg>]: value flag.
		if v, _, bare, ok := trailingFlag(a, "lang"); ok {
			langVal := v
			if bare && i+1 < len(args) {
				langVal = args[i+1]
				i++
			}
			switch {
			case supportedLangs[langVal]:
				i18n.SetLang(langVal)
			case langVal == "":
				// Bare `--lang` with no value mirrors detectBootstrapLang's tolerant
				// behavior — silently ignore rather than emit an empty-value warning.
			case !quiet:
				_, _ = fmt.Fprintln(stderr, i18n.Tf("cliUnsupportedLang", "lang", langVal))
				_, _ = fmt.Fprintln(stderr, i18n.T("cliSupportedLangs"))
			}
			continue
		}

		// --no-color / -no-color [=bool]: bool flag. An invalid `=value` is
		// left for the caller (not consumed), matching the original.
		if v, hasInline, _, ok := trailingFlag(a, "no-color"); ok {
			if !hasInline {
				trailingNoColor = true
				continue
			}
			if b, err := strconv.ParseBool(v); err == nil {
				if b {
					trailingNoColor = true
				}
				continue
			}
			rest = append(rest, a) // --no-color=<non-bool>: unrecognized, keep it
			continue
		}

		// --color / -color [=value | <next arg>]: value flag; resolution is
		// deferred to applyColorMode after the scan for correct precedence.
		if v, _, bare, ok := trailingFlag(a, "color"); ok {
			colorVal := v
			if bare && i+1 < len(args) {
				colorVal = args[i+1]
				i++
			}
			haveTrailingColor = true
			trailingColor = colorVal
			continue
		}

		// --quiet / -quiet / --q / -q [=bool]: value already folded into
		// `quiet` by resolveTrailingQuiet; here we only consume the token.
		// An invalid `=value` is left for the caller.
		if v, hasInline, _, ok := trailingFlagAny(a, "quiet", "q"); ok {
			if !hasInline {
				continue
			}
			if _, err := strconv.ParseBool(v); err == nil {
				continue
			}
			rest = append(rest, a) // --quiet=<non-bool>: unrecognized, keep it
			continue
		}

		rest = append(rest, a)
	}
	// Single delegated color resolution. Skipping the call when neither
	// flag was seen preserves the "trailing args without color flags
	// don't change anything" contract — the top-level applyColorMode
	// already ran in run() at startup and its result must remain.
	if haveTrailingColor || trailingNoColor {
		errSink := io.Discard
		if !quiet {
			errSink = stderr
		}
		mode := trailingColor
		if !haveTrailingColor {
			mode = "auto" // --no-color alone with no --color value
		}
		_, _ = applyColorMode(mode, trailingNoColor, os.Getenv("NO_COLOR"), os.Stdout.Fd(), os.Stderr.Fd(), errSink)
	}
	return rest
}

// resolveTrailingQuiet pre-scans args for -q / --quiet (and their =BOOL forms)
// and folds them into the starting quiet value. Used so warning suppression
// inside applyTrailingGlobalFlags is order-independent — `--lang xyz -q` must
// suppress just like `-q --lang xyz`. Stops at the first "--" the same way
// the main scan does.
func resolveTrailingQuiet(args []string, start bool) bool {
	quiet := start
	for _, a := range args {
		if a == "--" {
			break
		}
		v, hasInline, _, ok := trailingFlagAny(a, "quiet", "q")
		switch {
		case !ok:
			// not a quiet flag
		case !hasInline:
			quiet = true
		default:
			if b, err := strconv.ParseBool(v); err == nil {
				quiet = b
			}
		}
	}
	return quiet
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

// gameCategoryPreview is the number of representative game names rendered per
// category in the top-level --help GAMES summary. Five names is enough to
// hint at the variety in each category (e.g. casino → blackjack, baccarat,
// poker, omaha, holdem) without pushing COMMANDS / OPTIONS off the screen.
const gameCategoryPreview = 5

// buildHelpText generates the CLI help text with the games section derived
// from the registry.
//
// The GAMES section used to enumerate every one of the 77 games (#1694),
// which pushed COMMANDS / OPTIONS / EXIT CODES off the visible terminal in
// the common 24–40 line case. Now it presents a category-grouped summary
// with a pointer to `trumpcards games` for the full list, mirroring the
// `git --help` / `kubectl --help` / `cargo --help` style.
func buildHelpText() string {
	var sb strings.Builder
	categories := games.AllCategories()
	categoryNames := make([]string, len(categories))
	for i, c := range categories {
		categoryNames[i] = c.String()
	}
	fmt.Fprintf(&sb, `USAGE:
  trumpcards [--lang ja|en] [game]
  trumpcards --help

GAMES:
  %d games across %d categories. Run 'trumpcards games' for the full list,
  or 'trumpcards games --category <%s>' to filter.

`, len(ui.GameRegistry()), len(categories), strings.Join(categoryNames, "|"))
	for _, cat := range categories {
		entries := games.ByCategory(cat)
		preview := make([]string, 0, gameCategoryPreview)
		for i, g := range entries {
			if i >= gameCategoryPreview {
				break
			}
			preview = append(preview, g.Name)
		}
		more := ""
		if len(entries) > gameCategoryPreview {
			more = ", …"
		}
		fmt.Fprintf(&sb, "  %-8s (%2d)  %s%s\n", cat.String(), len(entries), strings.Join(preview, ", "), more)
	}
	sb.WriteString(`
COMMANDS:
  games        List all available games (--short for names only; with --short, --aliases adds alias lines)
  help [game]  Show this help, or a specific game's help text
  completion   Generate shell completion script (bash, zsh, fish)
  update       Self-update to the latest version
  version      Show version information (equivalent to --version; --short for machine-readable)
  web          Start REST API + web GUI server

  (no argument) Interactive mode with game switching

OPTIONS:
  -h, --help        Show this help message
  --start GAME      Initial game for interactive mode (no positional arg).
                    Aliases accepted (e.g. --start gin → ginrummy). Silently
                    ignored when a positional game name is given. An unknown
                    name exits 2 with a Did-you-mean suggestion. Default: blackjack.
  --lang ja|en      Language (default: ja)
  --color MODE      Color output mode: auto (default), always, never
                    auto:    enable when stdout/stderr is a TTY (per stream)
                    always:  force-enable even when piped (e.g. for tee or less -R)
                    never:   force-disable
                    Matches git/ls/grep convention. Use instead of --no-color.
                    Precedence: NO_COLOR env > --color=never (or --no-color)
                    > --color=always > --color=auto (https://no-color.org/).
  --no-color        DEPRECATED alias for --color=never. Will be removed in a
                    future release; prefer --color=never.
  -q, --quiet       Suppress non-essential output (banners, locale fallback warnings,
                    and the network-exposure warning printed by 'web --host 0.0.0.0').
                    Errors still go to stderr. Equivalent to TRUMPCARDS_QUIET=1.
  -V, --version     Show version information
  --version-short   Print version number only (machine-readable)

EXAMPLES:
  trumpcards                     Start interactive mode (switch games with 'switch <game>')
  trumpcards --start poker       Start interactive mode with poker as the initial game
  trumpcards --start gin         Same — aliases are accepted (gin → ginrummy)
  trumpcards blackjack           Play BlackJack
  trumpcards blackjack --help    Show BlackJack's in-game commands
  trumpcards --lang en poker     Play Poker in English
  trumpcards poker --lang en     Same — global flags also accepted after the game name
  trumpcards blackjack --no-color    Play BlackJack with color disabled (legacy)
  trumpcards blackjack --color=never  Same — preferred form
  trumpcards holdem | tee g.log       --color defaults to auto so the pipe disables color
  trumpcards holdem --color=always | tee g.log  Keep color even when piping
  trumpcards games               List all available games
  trumpcards games --short       List game names only (for scripting)
  trumpcards games --short --aliases  List game names including aliases
  trumpcards games --json        Machine-readable list (name, category, description, aliases)
  trumpcards games --category solo  Filter by Cloudflare Worker category (casino|classic|solo)
  trumpcards update              Self-update to the latest version
  trumpcards update --yes        Update without confirmation prompt
  trumpcards update --check      Report whether an update is available (exit 10 if yes)
  trumpcards --version-short     Print just the version number (e.g. 1.2.3)
  NO_COLOR=1 trumpcards hearts   Play Hearts without color output
  trumpcards web                 Start the web GUI server (binds to 127.0.0.1)
  trumpcards web --port 3000     Start the web GUI on port 3000
  trumpcards web --port 0        Start on an OS-assigned ephemeral port (see startup log)
  trumpcards web --host 0.0.0.0  Expose the web GUI on all interfaces
  source <(trumpcards completion bash)   Enable bash completion

ENVIRONMENT VARIABLES:
  NO_COLOR          Disable color output on both stdout and stderr when set
                    (see https://no-color.org/)
                    Example: NO_COLOR=1 trumpcards blackjack
  TRUMPCARDS_QUIET  Suppress non-essential output when set to a non-empty value
                    (equivalent to --quiet/-q). Errors still go to stderr.
                    Example: TRUMPCARDS_QUIET=1 trumpcards update --yes
  HOST              Bind address for the web server (default: 127.0.0.1)
                    Example: HOST=0.0.0.0 trumpcards web
  PORT              Port number for the web server (default: 8080)
                    Example: PORT=3000 trumpcards web

EXIT CODES:
   0  Success (normal exit, EOF, or 'exit' command)
   1  General error (e.g., web server failed to start, interactive input read error)
   2  Usage error (invalid flags, unknown game/command, unknown category, missing required argument)
  10  'update --check': a newer version is available (non-error signal for scripts)
  75  'update': user declined the confirmation prompt
 130  Terminated by SIGINT (Ctrl+C; POSIX 128 + 2)
 143  Terminated by SIGTERM (POSIX 128 + 15)

  See 'trumpcards help update' for update-specific exit codes (3, 4, 5, 6).
`)
	return sb.String()
}

// detectBootstrapLang returns the locale to use before flag parsing runs,
// so any flag-parse error can be emitted in the user's preferred language.
// Resolution mirrors Go's flag.Parse best-effort: an explicit `--lang`
// (or `-lang`, with optional `=value`) in args wins with last-occurrence
// semantics, scanning stops at the first non-flag arg or at `--`, then
// LANG env, then "ja". Unsupported values fall through to the next source.
// This is best-effort; the authoritative locale is still resolved after
// flag.Parse.
func detectBootstrapLang(args []string, langEnv string) string {
	fromEnv := "ja"
	if langEnv != "" {
		prefix := langEnv
		if idx := strings.IndexAny(langEnv, "_-."); idx >= 0 {
			prefix = langEnv[:idx]
		}
		if supportedLangs[prefix] {
			fromEnv = prefix
		}
	}
	fromArgs := ""
	for i := 0; i < len(args); i++ {
		a := args[i]
		// End-of-flags terminator: flag.Parse stops here.
		if a == "--" {
			break
		}
		// First non-flag positional arg (game name): flag.Parse stops here too.
		if !strings.HasPrefix(a, "-") {
			break
		}
		var val string
		switch {
		case a == "--lang" || a == "-lang":
			if i+1 < len(args) {
				val = args[i+1]
				i++ // consume the value so a non-flag value doesn't stop the scan
			}
		case strings.HasPrefix(a, "--lang="):
			val = strings.TrimPrefix(a, "--lang=")
		case strings.HasPrefix(a, "-lang="):
			val = strings.TrimPrefix(a, "-lang=")
		default:
			continue
		}
		if supportedLangs[val] {
			fromArgs = val // last-valid-wins, matching flag.Parse behavior
		}
	}
	if fromArgs != "" {
		return fromArgs
	}
	return fromEnv
}

// printGames writes the game list to w in the format selected by `short`.
// With short=false (long mode), each line shows the canonical name, description,
// and any aliases inline. With short=true, only canonical names are printed,
// one per line — and if aliases is also true, every alias gets its own line.
// The `aliases` flag is a no-op in long mode because aliases are always shown
// inline there. If category is non-empty, output is restricted to games whose
// games.Category matches; the caller is expected to have validated category
// via validCategory before invoking. See issue #1535.
func printGames(short, aliases bool, category string, w io.Writer) {
	var reverseAliases map[string][]string
	if !short || aliases {
		reverseAliases = buildReverseAliases()
	}

	var descs map[string]string
	if !short {
		descs = ui.GameDescriptions()
	}

	// Only build the category index when a filter is active — otherwise the
	// allocation is wasted because the gating predicate below is a no-op.
	var categoryByName map[string]string
	if category != "" {
		categoryByName = gameCategoryByName()
	}
	for _, name := range ui.GameNames() {
		if category != "" && categoryByName[name] != category {
			continue
		}
		if short {
			_, _ = fmt.Fprintln(w, name)
			if aliases {
				for _, alias := range reverseAliases[name] {
					_, _ = fmt.Fprintln(w, alias)
				}
			}
		} else {
			line := fmt.Sprintf("  %-16s %s", name, descs[name])
			if aliasList := reverseAliases[name]; len(aliasList) > 0 {
				line += fmt.Sprintf("  [aliases: %s]", strings.Join(aliasList, ", "))
			}
			_, _ = fmt.Fprintln(w, line)
		}
	}
}

// gameCategoryByName builds Name→Category-string from the games registry.
// The strings come from games.Category.String() so they stay in lockstep
// with the SSoT and validCategory's accepted values.
func gameCategoryByName() map[string]string {
	all := games.All()
	out := make(map[string]string, len(all))
	for _, g := range all {
		out[g.Name] = g.Category.String()
	}
	return out
}

// buildReverseAliases inverts ui.GameAliases (alias → canonical) into
// canonical → sorted aliases. Sorting per-key keeps output deterministic
// despite Go's randomized map iteration. Shared by printGames and
// printGamesJSON so the two render the same alias set in the same order.
func buildReverseAliases() map[string][]string {
	rev := make(map[string][]string)
	for alias, canonical := range ui.GameAliases {
		rev[canonical] = append(rev[canonical], alias)
	}
	for k := range rev {
		sort.Strings(rev[k])
	}
	return rev
}

// printGamesJSON emits a JSON array describing every game (or only games in
// the given category, if non-empty). Each entry is `{name, category,
// description, aliases}`. Aliases is always a non-nil slice so the JSON
// shape is stable: scripts can rely on `.aliases | length` working without a
// null guard. See issue #1535.
func printGamesJSON(category string, w io.Writer) error {
	type entry struct {
		Name        string   `json:"name"`
		Category    string   `json:"category"`
		Description string   `json:"description"`
		Aliases     []string `json:"aliases"`
	}
	reverseAliases := buildReverseAliases()
	all := games.All()
	out := make([]entry, 0, len(all))
	for _, g := range all {
		cat := g.Category.String()
		if category != "" && cat != category {
			continue
		}
		al := reverseAliases[g.Name]
		if al == nil {
			al = []string{} // emit `[]`, never `null` — stable schema for scripts.
		}
		out = append(out, entry{
			Name:        g.Name,
			Category:    cat,
			Description: games.Description(g.Name),
			Aliases:     al,
		})
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

// realtimeGames are the CUI games that run with the realtime runner
// (raw-mode keystrokes + auto-tick goroutine) instead of the standard
// line-based loop. Slapjack and Egyptian Ratscrew rely on a fast tick
// cadence for CPU pending actions; the line-based loop forced the user
// to type "tick" to advance, which broke the "reflexes" gameplay (#1653).
var realtimeGames = map[string]bool{
	"slapjack":         true,
	"egyptianratscrew": true,
}

// buildGameCommands generates command handlers for all games from the registry.
func buildGameCommands() map[string]func() int {
	registry := ui.GameRegistry()
	commands := make(map[string]func() int, len(registry)+4)
	for _, entry := range registry {
		e := entry // capture loop variable
		commands[e.Name] = func() int {
			g := e.NewCui()
			if realtimeGames[e.Name] {
				return ui.RunRealtimeCuiLoop(e.Name, g.Controller(), g.HelpLines())
			}
			return ui.RunCuiLoop(e.Name, g.Controller(), g.HelpLines())
		}
	}
	return commands
}
