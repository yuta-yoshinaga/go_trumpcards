//go:build test

package ui

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// helpExampleCommandRe pulls the typed command out of a rendered example line.
// The line is "  <command>   <description>" and the two columns are separated by
// a run of two or more spaces, which is what keeps a multi-word command such as
// `pass 0 3 7` intact where strings.Fields would split it into the verb alone.
var helpExampleCommandRe = regexp.MustCompile(`^\s+(\S+(?:\s\S+)*?)\s{2,}\S`)

func helpExampleCommand(line string) string {
	m := helpExampleCommandRe.FindStringSubmatch(line)
	if m == nil {
		return ""
	}
	return m[1]
}

// TestCuiHelpExamplesExecute runs every documented example through the game's
// real CUI controller and fails if the parser rejects it.
//
// TestCuiHelpExamplesUseRealCommands already checks that an example's verb is
// listed in the same game's command table, but the command table is itself
// documentation: when the table and the parser disagree the example can name a
// verb that is documented and still unimplemented, and nothing notices. This
// executes the string instead of comparing it against another string, so the
// parser is the authority. It is also the only check that sees the argument at
// all -- `p 0` and `p abc` are indistinguishable to a verb-only guard.
//
// The examples of one game are executed in order on ONE controller, because
// they are written as a worked sequence: blackjack's `h` is only legal after its
// `b 100` has been accepted.
//
// What this does NOT catch, measured rather than assumed: of the 291 games with
// an argument-taking command, 289 answer a malformed argument with a board
// redraw instead of a marked error (`b abc`, `b -5` and `ra 999999999` are all
// silently swallowed by the poker controllers). So for those games this guard
// still only proves the verb reaches a handler. Tightening it needs the
// controllers to mark those rejections first -- tracked separately.
func TestCuiHelpExamplesExecute(t *testing.T) {
	registry := GameRegistry()
	if len(registry) < 300 {
		t.Fatalf("only %d games in the registry -- the walk broke", len(registry))
	}

	games, executed := 0, 0
	var bad []string
	for _, entry := range registry {
		g := entry.NewCui()
		_, examples := helpSections(g.HelpLines())
		if len(examples) == 0 {
			continue
		}
		games++
		ctrl := g.Controller()
		// Deal first. GameManager.initGame sends "r" before a game accepts any
		// input, and without it every controller sits in an un-dealt state where
		// it answers a board redraw to anything -- which reads as "the example
		// was accepted" and made the first version of this guard vacuous.
		ctrl.Exec("r")
		for _, line := range examples {
			cmd := helpExampleCommand(line)
			if cmd == "" {
				bad = append(bad, entry.Name+": no command column in example line "+strconv.Quote(line))
				continue
			}
			executed++
			if body, isErr := i18n.StripErrorPrefix(ctrl.Exec(cmd)); isErr {
				bad = append(bad, entry.Name+": `"+cmd+"` was rejected -- "+strings.TrimSpace(body))
			}
		}
	}

	if games == 0 {
		t.Fatal("no game rendered an examples section -- either none wires ExampleKeys or helpSections stopped matching")
	}
	if executed == 0 {
		t.Fatalf("%d games have examples but no command was parsed out of them -- helpExampleCommandRe stopped matching", games)
	}
	sort.Strings(bad)
	if len(bad) > 0 {
		t.Errorf("examples the CUI parser rejects (%d of %d executed across %d games):\n  %s",
			len(bad), executed, games, strings.Join(bad, "\n  "))
	}
}

// TestCuiHelpExampleCommandParsing pins the column split, which is the part of
// the guard above that fails silently: a regex that stopped matching would make
// every example unparseable, and "no examples ran" reads exactly like "every
// example passed" unless something asserts the split itself.
func TestCuiHelpExampleCommandParsing(t *testing.T) {
	for _, tc := range []struct{ line, want string }{
		{"  b 100                bet 100 chips and deal", "b 100"},
		{"  pass 0 3 7           pass hand cards 0, 3 and 7", "pass 0 3 7"},
		{"  h                    draw one more card", "h"},
		{"  m t0 f               move a tableau card to a foundation", "m t0 f"},
		{"  b 100", ""},       // no description column -> not an example row
		{"single-column", ""}, // not indented
	} {
		if got := helpExampleCommand(tc.line); got != tc.want {
			t.Errorf("helpExampleCommand(%q) = %q, want %q", tc.line, got, tc.want)
		}
	}
}
