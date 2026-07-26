package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/ui"
)

// supportedCompletionShells is the canonical list of shells supported by `trumpcards completion`.
var supportedCompletionShells = []string{"bash", "zsh", "fish"}

// completionSubcommands returns the sorted list of all subcommands for shell completion,
// derived from the game registry and aliases.
func completionSubcommands() []string {
	names := make([]string, 0, len(ui.GameNames())+len(ui.GameAliases)+6)
	names = append(names, ui.GameNames()...)
	for alias := range ui.GameAliases {
		names = append(names, alias)
	}
	names = append(names, "completion", "games", "help", "update", "version", "web")
	sort.Strings(names)
	return names
}

// completionGameTargets returns the sorted list of every game name + alias,
// without the non-game subcommands. Used to back the value-completion of
// `--start`, which only accepts a game (issue #1604), and the argument
// completion of `help <game>`.
func completionGameTargets() []string {
	names := make([]string, 0, len(ui.GameNames())+len(ui.GameAliases))
	names = append(names, ui.GameNames()...)
	for alias := range ui.GameAliases {
		names = append(names, alias)
	}
	sort.Strings(names)
	return names
}

// runCompletion outputs a shell completion script for the given shell name.
// Returns 0 on success, 2 on usage error (missing arg or unsupported shell —
// matches the documented EXIT CODES section in buildHelpText), 1 on a real
// I/O failure while writing the script. See issue #1603.
func runCompletion(args []string, stdoutIsTTY, noHint bool) int {
	return runCompletionTo(args, os.Stdout, os.Stderr, stdoutIsTTY, noHint)
}

// runCompletionTo is the testable core of runCompletion: it writes the script
// to stdout and any diagnostic messages to stderr.
//
// Install hints are emitted only when stdout is a TTY and --no-hint was not
// passed. This keeps `trumpcards completion bash > file` output pure while
// preserving the onboarding comment for users who run it interactively
// (`source <(…)` is a pipe, so hints stay out of the shell session too).
// See issue #1450.
//
// Usage errors (no argument or unsupported shell) return 2 to match the
// documented EXIT CODES table and the rest of the CLI (`parseSubFlags` also
// returns 2 for unknown flags). A genuine I/O failure while writing the
// script still returns 1 so CI can distinguish "you typed it wrong" from
// "the disk is full". See issue #1603.
func runCompletionTo(args []string, stdout, stderr io.Writer, stdoutIsTTY, noHint bool) int {
	if len(args) != 1 {
		_, _ = fmt.Fprintln(stderr, i18n.T("cliCompletionUsage"))
		_, _ = fmt.Fprintln(stderr, i18n.Tf("cliTryHelp", "cmd", "completion"))
		return 2
	}
	shell := args[0]
	showHint := stdoutIsTTY && !noHint
	var err error
	switch shell {
	case "bash":
		if showHint {
			writeInstallHint(stdout, shell)
		}
		err = writeBashCompletion(stdout)
	case "zsh":
		if showHint {
			writeInstallHint(stdout, shell)
		}
		err = writeZshCompletion(stdout)
	case "fish":
		if showHint {
			writeInstallHint(stdout, shell)
		}
		err = writeFishCompletion(stdout)
	default:
		_, _ = fmt.Fprintln(stderr, i18n.Tf("cliCompletionUnsupportedShell", "shell", shell))
		if suggestion := cuiutil.SuggestCommand(shell, supportedCompletionShells, 2); suggestion != "" {
			_, _ = fmt.Fprintf(stderr, "  %s\n", i18n.Tf("didYouMean", "name", suggestion))
		}
		return 2
	}
	if err != nil {
		_, _ = fmt.Fprintln(stderr, i18n.Tf("cliCompletionWriteError", "err", err.Error()))
		return 1
	}
	return 0
}

// writeInstallHint writes shell-specific installation instructions to w as #-prefixed comments.
func writeInstallHint(w io.Writer, shell string) {
	var hint string
	switch shell {
	case "bash":
		hint = `# To load completions for the current session:
#   source <(trumpcards completion bash)
#
# To persist across sessions, add to your ~/.bashrc:
#   echo 'source <(trumpcards completion bash)' >> ~/.bashrc
`
	case "zsh":
		hint = `# To load completions for the current session:
#   source <(trumpcards completion zsh)
#
# To persist across sessions (add to fpath):
#   trumpcards completion zsh > "${fpath[1]}/_trumpcards"
#   # Then restart your shell or run: compinit
`
	case "fish":
		hint = `# To load completions for the current session:
#   trumpcards completion fish | source
#
# To persist across sessions:
#   trumpcards completion fish > ~/.config/fish/completions/trumpcards.fish
`
	}
	if hint != "" {
		_, _ = fmt.Fprint(w, hint)
	}
}

func writeBashCompletion(w io.Writer) error {
	cmds := strings.Join(completionSubcommands(), " ")
	games := strings.Join(completionGameTargets(), " ")
	cats := strings.Join(categoryDisplayNames(), " ")
	// `commands` (every game + alias + subcommand) and `games` (game targets
	// only, used by --start and `help <target>`) are declared once at the
	// top of the function so each case reuses them, avoiding the previous
	// duplicate `local games="..."` blocks and ensuring `help` stays in
	// lockstep with the canonical subcommand list (review feedback on #1756).
	script := fmt.Sprintf(`_trumpcards() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    local prev="${COMP_WORDS[COMP_CWORD-1]}"
    local commands="%[1]s"
    local games="%[2]s"

    case "$prev" in
        --lang)
            COMPREPLY=( $(compgen -W "ja en" -- "$cur") )
            return
            ;;
        --color)
            COMPREPLY=( $(compgen -W "auto always never" -- "$cur") )
            return
            ;;
        --start)
            COMPREPLY=( $(compgen -W "$games" -- "$cur") )
            return
            ;;
        completion)
            COMPREPLY=( $(compgen -W "bash zsh fish" -- "$cur") )
            return
            ;;
        update)
            COMPREPLY=( $(compgen -W "--yes -y --check --dry-run" -- "$cur") )
            return
            ;;
        --port|-p|--host)
            return
            ;;
        web)
            COMPREPLY=( $(compgen -W "--port -p --host --open -o --quiet -q" -- "$cur") )
            return
            ;;
        --category)
            COMPREPLY=( $(compgen -W "%[3]s" -- "$cur") )
            return
            ;;
        games|--short|--aliases|--json)
            COMPREPLY=( $(compgen -W "--short --aliases --json --category" -- "$cur") )
            return
            ;;
        help)
            COMPREPLY=( $(compgen -W "$commands" -- "$cur") )
            return
            ;;
        version)
            COMPREPLY=( $(compgen -W "--short" -- "$cur") )
            return
            ;;
    esac

    if [[ "$cur" == -* ]]; then
        COMPREPLY=( $(compgen -W "--help --lang --color --no-color --version --version-short --start --quiet -h -V -q" -- "$cur") )
        return
    fi

    COMPREPLY=( $(compgen -W "$commands" -- "$cur") )
}
complete -F _trumpcards trumpcards
`, cmds, games, cats)
	_, err := fmt.Fprint(w, script)
	return err
}

func writeZshCompletion(w io.Writer) error {
	entries := buildCompletionEntries()
	var sb strings.Builder
	for _, e := range entries {
		// Escape single quotes for zsh: ' -> '\''
		desc := strings.ReplaceAll(e.desc, "'", `'\''`)
		fmt.Fprintf(&sb, "        '%s:%s'\n", e.name, desc)
	}
	games := strings.Join(completionGameTargets(), " ")
	cats := strings.Join(categoryDisplayNames(), " ")
	script := fmt.Sprintf(`#compdef trumpcards

_trumpcards() {
    local -a commands
    commands=(
%[1]s    )
    local -a game_targets
    game_targets=(%[2]s)

    _arguments \
        '(-h --help)'{-h,--help}'[Show help message]' \
        '(-V --version)'{-V,--version}'[Show version information]' \
        '--version-short[Print version number only]' \
        '(-q --quiet)'{-q,--quiet}'[Suppress non-essential output]' \
        '--lang[Language]:language:(ja en)' \
        '--color[Color output mode]:mode:(auto always never)' \
        '--no-color[DEPRECATED alias for --color=never]' \
        '--start[Initial game for interactive mode]:game:(${game_targets})' \
        '1:command:->cmds' \
        '*::arg:->args'

    case "$state" in
        cmds)
            _describe -t commands 'trumpcards command' commands
            ;;
        args)
            case "${words[1]}" in
                completion)
                    _values 'shell' bash zsh fish
                    ;;
                update)
                    _arguments \
                        '(-y --yes)'{-y,--yes}'[Skip confirmation prompt]' \
                        '--check[Check for an update without installing]' \
                        '--dry-run[Alias for --check]'
                    ;;
                games)
                    _arguments \
                        '--short[Print game names only]' \
                        '--aliases[Include aliases in output]' \
                        '--json[Emit machine-readable JSON]' \
                        '--category[Filter by category]:category:(%[3]s)'
                    ;;
                web)
                    _arguments \
                        '(-p --port)'{-p,--port}'[Port number]:port:' \
                        '--host[Bind address]:host:' \
                        '(-q --quiet)'{-q,--quiet}'[Suppress startup/shutdown messages]' \
                        '(-o --open)'{-o,--open}'[Open the resolved URL in the default browser]'
                    ;;
                version)
                    _arguments \
                        '--short[Print version number only]'
                    ;;
                help)
                    _values 'target' "${game_targets[@]}" completion games help update version web
                    ;;
            esac
            ;;
    esac
}

_trumpcards "$@"
`, sb.String(), games, cats)
	_, err := fmt.Fprint(w, script)
	return err
}

func writeFishCompletion(w io.Writer) error {
	entries := buildCompletionEntries()
	var sb strings.Builder
	for _, e := range entries {
		// Escape single quotes for fish: ' -> \'
		desc := strings.ReplaceAll(e.desc, "'", `\'`)
		fmt.Fprintf(&sb, "complete -c trumpcards -n __fish_use_subcommand -a %s -d '%s'\n", e.name, desc)
	}
	games := strings.Join(completionGameTargets(), " ")
	cats := strings.Join(categoryDisplayNames(), " ")
	script := fmt.Sprintf(`# Fish completion for trumpcards
complete -c trumpcards -f

# Global options
complete -c trumpcards -l help -s h -d 'Show help message'
complete -c trumpcards -l version -s V -d 'Show version information'
complete -c trumpcards -l version-short -d 'Print version number only'
complete -c trumpcards -l quiet -s q -d 'Suppress non-essential output'
complete -c trumpcards -l lang -x -a 'ja en' -d 'Language'
complete -c trumpcards -l color -x -a 'auto always never' -d 'Color output mode'
complete -c trumpcards -l no-color -d 'DEPRECATED alias for --color=never'
complete -c trumpcards -l start -x -a '%[2]s' -d 'Initial game for interactive mode'

# Subcommands
%[1]s
# completion subcommand
complete -c trumpcards -n '__fish_seen_subcommand_from completion' -a 'bash zsh fish'

# update subcommand
complete -c trumpcards -n '__fish_seen_subcommand_from update' -l yes -s y -d 'Skip confirmation prompt'
complete -c trumpcards -n '__fish_seen_subcommand_from update' -l check -d 'Check for an update without installing'
complete -c trumpcards -n '__fish_seen_subcommand_from update' -l dry-run -d 'Alias for --check'

# games subcommand
complete -c trumpcards -n '__fish_seen_subcommand_from games' -l short -d 'Print game names only'
complete -c trumpcards -n '__fish_seen_subcommand_from games' -l aliases -d 'Include aliases in output'
complete -c trumpcards -n '__fish_seen_subcommand_from games' -l json -d 'Emit machine-readable JSON'
complete -c trumpcards -n '__fish_seen_subcommand_from games' -l category -x -a '%[3]s' -d 'Filter by category'

# web subcommand
complete -c trumpcards -n '__fish_seen_subcommand_from web' -l port -s p -d 'Port number' -x
complete -c trumpcards -n '__fish_seen_subcommand_from web' -l host -d 'Bind address' -x
complete -c trumpcards -n '__fish_seen_subcommand_from web' -l quiet -s q -d 'Suppress startup/shutdown messages'
complete -c trumpcards -n '__fish_seen_subcommand_from web' -l open -s o -d 'Open the resolved URL in the default browser'

# version subcommand
complete -c trumpcards -n '__fish_seen_subcommand_from version' -l short -d 'Print version number only'

# help subcommand (game or subcommand argument)
complete -c trumpcards -n '__fish_seen_subcommand_from help' -a '%[2]s completion games help update version web'
`, sb.String(), games, cats)
	_, err := fmt.Fprint(w, script)
	return err
}

// completionEntry holds a command name and its short description for shell completion.
type completionEntry struct {
	name string
	desc string
}

// buildCompletionEntries builds a sorted list of all commands with descriptions
// for shell completion, derived from the game registry and aliases.
func buildCompletionEntries() []completionEntry {
	descs := ui.GameDescriptions()
	// Strip Japanese text in parentheses for cleaner completion descriptions.
	stripJa := func(s string) string {
		if idx := strings.Index(s, " ("); idx >= 0 {
			return s[:idx]
		}
		return s
	}

	entries := make([]completionEntry, 0, len(ui.GameNames())+len(ui.GameAliases)+6)
	for _, name := range ui.GameNames() {
		entries = append(entries, completionEntry{name, stripJa(descs[name])})
	}
	for alias, canonical := range ui.GameAliases {
		entries = append(entries, completionEntry{alias, stripJa(descs[canonical]) + " (alias)"})
	}
	entries = append(entries,
		completionEntry{"completion", "Generate shell completion script"},
		completionEntry{"games", "List available games"},
		completionEntry{"help", "Show help, optionally for a specific game"},
		completionEntry{"update", "Self-update to the latest version"},
		completionEntry{"version", "Show version information"},
		completionEntry{"web", "Start REST API + web GUI server"},
	)
	sort.Slice(entries, func(i, j int) bool { return entries[i].name < entries[j].name })
	return entries
}
