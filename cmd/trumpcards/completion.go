package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/ui"
)

// completionSubcommands returns the sorted list of all subcommands for shell completion,
// derived from the game registry and aliases.
func completionSubcommands() []string {
	names := make([]string, 0, len(ui.GameNames())+len(ui.GameAliases)+4)
	names = append(names, ui.GameNames()...)
	for alias := range ui.GameAliases {
		names = append(names, alias)
	}
	names = append(names, "completion", "games", "update", "web")
	sort.Strings(names)
	return names
}

// runCompletion outputs a shell completion script for the given shell name.
// Returns 0 on success, 1 on error.
func runCompletion(args []string) int {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "Usage: trumpcards completion <bash|zsh|fish>")
		return 1
	}
	shell := args[0]
	var err error
	switch shell {
	case "bash":
		writeInstallHint(os.Stdout, shell)
		err = writeBashCompletion(os.Stdout)
	case "zsh":
		writeInstallHint(os.Stdout, shell)
		err = writeZshCompletion(os.Stdout)
	case "fish":
		writeInstallHint(os.Stdout, shell)
		err = writeFishCompletion(os.Stdout)
	default:
		fmt.Fprintf(os.Stderr, "Error: unsupported shell %q (supported: bash, zsh, fish)\n", shell)
		return 1
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
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
	script := fmt.Sprintf(`_trumpcards() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    local prev="${COMP_WORDS[COMP_CWORD-1]}"

    case "$prev" in
        --lang)
            COMPREPLY=( $(compgen -W "ja en" -- "$cur") )
            return
            ;;
        completion)
            COMPREPLY=( $(compgen -W "bash zsh fish" -- "$cur") )
            return
            ;;
        update)
            COMPREPLY=( $(compgen -W "--yes -y" -- "$cur") )
            return
            ;;
        --port|-p)
            return
            ;;
    esac

    if [[ "$cur" == -* ]]; then
        COMPREPLY=( $(compgen -W "--help --lang --no-color --version -h -V" -- "$cur") )
        return
    fi

    local commands="%s"
    COMPREPLY=( $(compgen -W "$commands" -- "$cur") )
}
complete -F _trumpcards trumpcards
`, cmds)
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
	script := fmt.Sprintf(`#compdef trumpcards

_trumpcards() {
    local -a commands
    commands=(
%s    )

    _arguments \
        '(-h --help)'{-h,--help}'[Show help message]' \
        '(-V --version)'{-V,--version}'[Show version information]' \
        '--lang[Language]:language:(ja en)' \
        '--no-color[Disable color output]' \
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
                        '(-y --yes)'{-y,--yes}'[Skip confirmation prompt]'
                    ;;
                web)
                    _arguments \
                        '(-p --port)'{-p,--port}'[Port number]:port:'
                    ;;
            esac
            ;;
    esac
}

_trumpcards "$@"
`, sb.String())
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
	script := fmt.Sprintf(`# Fish completion for trumpcards
complete -c trumpcards -f

# Global options
complete -c trumpcards -l help -s h -d 'Show help message'
complete -c trumpcards -l version -s V -d 'Show version information'
complete -c trumpcards -l lang -x -a 'ja en' -d 'Language'
complete -c trumpcards -l no-color -d 'Disable color output'

# Subcommands
%s
# completion subcommand
complete -c trumpcards -n '__fish_seen_subcommand_from completion' -a 'bash zsh fish'

# update subcommand
complete -c trumpcards -n '__fish_seen_subcommand_from update' -l yes -s y -d 'Skip confirmation prompt'

# web subcommand
complete -c trumpcards -n '__fish_seen_subcommand_from web' -l port -s p -d 'Port number' -x
`, sb.String())
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

	entries := make([]completionEntry, 0, len(ui.GameNames())+len(ui.GameAliases)+4)
	for _, name := range ui.GameNames() {
		entries = append(entries, completionEntry{name, stripJa(descs[name])})
	}
	for alias, canonical := range ui.GameAliases {
		entries = append(entries, completionEntry{alias, stripJa(descs[canonical]) + " (alias)"})
	}
	entries = append(entries,
		completionEntry{"completion", "Generate shell completion script"},
		completionEntry{"games", "List available games"},
		completionEntry{"update", "Self-update to the latest version"},
		completionEntry{"web", "Start REST API + web GUI server"},
	)
	sort.Slice(entries, func(i, j int) bool { return entries[i].name < entries[j].name })
	return entries
}
