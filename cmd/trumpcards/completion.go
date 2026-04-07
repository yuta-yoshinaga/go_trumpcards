package main

import (
	"fmt"
	"io"
	"os"
)

// completionSubcommands is the list of all subcommands for shell completion.
var completionSubcommands = []string{
	"baccarat", "blackjack", "bridge", "canasta", "clocksolitaire", "completion",
	"crazyeights", "cribbage", "daifugo", "deuceswild", "doubt", "euchre",
	"freecell", "games", "ginrummy", "gofish", "golf", "hearts", "holdem",
	"indianpoker", "jokerpoker", "klondike", "memory", "napoleon", "ohhell",
	"oldmaid", "omaha", "pigtail", "pineapple", "pinochle", "poker", "pyramid",
	"sevencardstud", "sevens", "shortdeck", "spades", "speed", "spider",
	"threecard", "tripeaks", "update", "videopoker", "web",
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
	script := `_trumpcards() {
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
        --port|-p)
            return
            ;;
    esac

    if [[ "$cur" == -* ]]; then
        COMPREPLY=( $(compgen -W "--help --lang --no-color --version -h -V" -- "$cur") )
        return
    fi

    local commands="baccarat blackjack bridge canasta clocksolitaire completion crazyeights cribbage daifugo deuceswild doubt euchre freecell games ginrummy gofish golf hearts holdem indianpoker jokerpoker klondike memory napoleon ohhell oldmaid omaha pigtail pineapple pinochle poker pyramid sevencardstud sevens shortdeck spades speed spider threecard tripeaks update videopoker web"
    COMPREPLY=( $(compgen -W "$commands" -- "$cur") )
}
complete -F _trumpcards trumpcards
`
	_, err := fmt.Fprint(w, script)
	return err
}

func writeZshCompletion(w io.Writer) error {
	script := `#compdef trumpcards

_trumpcards() {
    local -a commands
    commands=(
        'baccarat:Baccarat'
        'blackjack:BlackJack'
        'bridge:Contract Bridge'
        'canasta:Canasta'
        'clocksolitaire:Clock Solitaire'
        'completion:Generate shell completion script'
        'crazyeights:Crazy Eights'
        'cribbage:Cribbage'
        'daifugo:Daifugo'
        'deuceswild:Deuces Wild'
        'doubt:Doubt'
        'euchre:Euchre'
        'freecell:FreeCell'
        'games:List available games'
        'ginrummy:Gin Rummy'
        'gofish:Go Fish'
        'golf:Golf Solitaire'
        'hearts:Hearts'
        'holdem:Texas Hold'\''em'
        'indianpoker:Indian Poker'
        'jokerpoker:Joker Poker'
        'klondike:Klondike Solitaire'
        'memory:Memory'
        'napoleon:Napoleon'
        'ohhell:Oh Hell'
        'oldmaid:Old Maid'
        'omaha:Omaha Hold'\''em'
        'pigtail:Pig'\''s Tail'
        'pineapple:Pineapple Poker'
        'pinochle:Pinochle'
        'poker:5-card Draw Poker'
        'pyramid:Pyramid'
        'sevencardstud:Seven Card Stud'
        'sevens:Sevens'
        'shortdeck:Short Deck Hold'\''em'
        'spades:Spades'
        'speed:Speed'
        'spider:Spider Solitaire'
        'threecard:Three Card Poker'
        'tripeaks:TriPeaks'
        'update:Self-update to the latest version'
        'videopoker:Video Poker'
        'web:Start REST API + web GUI server'
    )

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
                web)
                    _arguments \
                        '(-p --port)'{-p,--port}'[Port number]:port:'
                    ;;
            esac
            ;;
    esac
}

_trumpcards "$@"
`
	_, err := fmt.Fprint(w, script)
	return err
}

func writeFishCompletion(w io.Writer) error {
	script := `# Fish completion for trumpcards
complete -c trumpcards -f

# Global options
complete -c trumpcards -l help -s h -d 'Show help message'
complete -c trumpcards -l version -s V -d 'Show version information'
complete -c trumpcards -l lang -x -a 'ja en' -d 'Language'
complete -c trumpcards -l no-color -d 'Disable color output'

# Subcommands
complete -c trumpcards -n __fish_use_subcommand -a baccarat -d 'Baccarat'
complete -c trumpcards -n __fish_use_subcommand -a blackjack -d 'BlackJack'
complete -c trumpcards -n __fish_use_subcommand -a bridge -d 'Contract Bridge'
complete -c trumpcards -n __fish_use_subcommand -a canasta -d 'Canasta'
complete -c trumpcards -n __fish_use_subcommand -a clocksolitaire -d 'Clock Solitaire'
complete -c trumpcards -n __fish_use_subcommand -a completion -d 'Generate shell completion script'
complete -c trumpcards -n __fish_use_subcommand -a crazyeights -d 'Crazy Eights'
complete -c trumpcards -n __fish_use_subcommand -a cribbage -d 'Cribbage'
complete -c trumpcards -n __fish_use_subcommand -a daifugo -d 'Daifugo'
complete -c trumpcards -n __fish_use_subcommand -a deuceswild -d 'Deuces Wild'
complete -c trumpcards -n __fish_use_subcommand -a doubt -d 'Doubt'
complete -c trumpcards -n __fish_use_subcommand -a euchre -d 'Euchre'
complete -c trumpcards -n __fish_use_subcommand -a freecell -d 'FreeCell'
complete -c trumpcards -n __fish_use_subcommand -a games -d 'List available games'
complete -c trumpcards -n __fish_use_subcommand -a ginrummy -d 'Gin Rummy'
complete -c trumpcards -n __fish_use_subcommand -a gofish -d 'Go Fish'
complete -c trumpcards -n __fish_use_subcommand -a golf -d 'Golf Solitaire'
complete -c trumpcards -n __fish_use_subcommand -a hearts -d 'Hearts'
complete -c trumpcards -n __fish_use_subcommand -a holdem -d 'Texas Hold'\''em'
complete -c trumpcards -n __fish_use_subcommand -a indianpoker -d 'Indian Poker'
complete -c trumpcards -n __fish_use_subcommand -a jokerpoker -d 'Joker Poker'
complete -c trumpcards -n __fish_use_subcommand -a klondike -d 'Klondike Solitaire'
complete -c trumpcards -n __fish_use_subcommand -a memory -d 'Memory'
complete -c trumpcards -n __fish_use_subcommand -a napoleon -d 'Napoleon'
complete -c trumpcards -n __fish_use_subcommand -a ohhell -d 'Oh Hell'
complete -c trumpcards -n __fish_use_subcommand -a oldmaid -d 'Old Maid'
complete -c trumpcards -n __fish_use_subcommand -a omaha -d 'Omaha Hold'\''em'
complete -c trumpcards -n __fish_use_subcommand -a pigtail -d 'Pig'\''s Tail'
complete -c trumpcards -n __fish_use_subcommand -a pineapple -d 'Pineapple Poker'
complete -c trumpcards -n __fish_use_subcommand -a pinochle -d 'Pinochle'
complete -c trumpcards -n __fish_use_subcommand -a poker -d '5-card Draw Poker'
complete -c trumpcards -n __fish_use_subcommand -a pyramid -d 'Pyramid'
complete -c trumpcards -n __fish_use_subcommand -a sevencardstud -d 'Seven Card Stud'
complete -c trumpcards -n __fish_use_subcommand -a sevens -d 'Sevens'
complete -c trumpcards -n __fish_use_subcommand -a shortdeck -d 'Short Deck Hold'\''em'
complete -c trumpcards -n __fish_use_subcommand -a spades -d 'Spades'
complete -c trumpcards -n __fish_use_subcommand -a speed -d 'Speed'
complete -c trumpcards -n __fish_use_subcommand -a spider -d 'Spider Solitaire'
complete -c trumpcards -n __fish_use_subcommand -a threecard -d 'Three Card Poker'
complete -c trumpcards -n __fish_use_subcommand -a tripeaks -d 'TriPeaks'
complete -c trumpcards -n __fish_use_subcommand -a update -d 'Self-update to the latest version'
complete -c trumpcards -n __fish_use_subcommand -a videopoker -d 'Video Poker'
complete -c trumpcards -n __fish_use_subcommand -a web -d 'Start REST API + web GUI server'

# completion subcommand
complete -c trumpcards -n '__fish_seen_subcommand_from completion' -a 'bash zsh fish'

# web subcommand
complete -c trumpcards -n '__fish_seen_subcommand_from web' -l port -s p -d 'Port number' -x
`
	_, err := fmt.Fprint(w, script)
	return err
}
