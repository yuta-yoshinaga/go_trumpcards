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
// an argument-taking command, only 2 answer a malformed argument with a marked
// error. 262 answer with a specific and perfectly good message that never went
// through i18n.MarkError ("Invalid amount: zznotanumber"), and 27 answer with
// nothing but a redrawn board. Unmarked is indistinguishable from accepted
// here, so for those 289 this guard still only proves the verb reaches a
// handler. Marking them is issue #5377.
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

// exampleRefusalRe matches the wording a controller uses when it turns a
// command down. It is a heuristic, and deliberately so: see below.
var exampleRefusalRe = regexp.MustCompile(`(?i)invalid|required|must |cannot|nothing to|no .* to |not your|too few|too many|不正|無効|必要|できません|ありません`)

// exampleErrorLineRe matches cuiErrorBlock's output: the presenters render a
// rejection as a red line inside the board, so it is not a short reply and
// cannot be found by looking at the whole output's length.
var exampleErrorLineRe = regexp.MustCompile("\x1b\\[31m(.*?)\x1b\\[0m")

// exampleRefusalLine reports the line on which the game turned the command
// down, if it did.
//
// Two shapes, because the CUI has two. A game that answers with nothing but a
// message is a short reply; a game that redraws the board puts the message
// inside it as a red line, which an earlier version of this guard missed
// entirely -- it required the whole reply to be under 200 bytes, and a board is
// not. blackhole's `u` answered "blackhole: nothing to undo" in 965 bytes of
// board and sailed straight through.
func exampleRefusalLine(out string) (string, bool) {
	// Red is not only for errors -- hearts and diamonds are red too, so a red
	// run has to read like a refusal before it counts as one.
	for _, m := range exampleErrorLineRe.FindAllStringSubmatch(out, -1) {
		line := strings.TrimSpace(m[1])
		if line != "" && exampleRefusalRe.MatchString(line) {
			return line, true
		}
	}
	if len(out) < 200 && exampleRefusalRe.MatchString(out) {
		return strings.TrimSpace(out), true
	}
	return "", false
}

// TestCuiHelpExamplesAreNotQuietlyRefused catches the refusals
// TestCuiHelpExamplesExecute cannot see.
//
// That guard asks i18n.StripErrorPrefix, so it only sees a refusal that went
// through i18n.MarkError. The games still on the literal-message ParseIntArg
// return an unmarked string, which is byte-for-byte an ordinary reply -- and
// that is exactly how `p` shipped as mississippistud's example while the game
// answered "Multiplier (1, 2 or 3) is required." to it. My own pre-flight probe
// used the same predicate and reported the same clean result, so the blind spot
// was in the measurement, not just the guard.
//
// Matching on the wording is a heuristic and would not survive a rewording, but
// it is checked against every example on every run, and a false positive is a
// sentence to read rather than a broken build. It can be deleted once #5377
// leaves nothing unmarked.
func TestCuiHelpExamplesAreNotQuietlyRefused(t *testing.T) {
	registry := GameRegistry()
	if len(registry) < 300 {
		t.Fatalf("only %d games in the registry -- the walk broke", len(registry))
	}

	// The predicate has to recognise a refusal, or "0 suspect" means nothing.
	if !exampleRefusalRe.MatchString("Multiplier (1, 2 or 3) is required.") {
		t.Fatal("the refusal pattern no longer matches a real refusal message")
	}
	if exampleRefusalRe.MatchString("Round: 1  Trick: 0") {
		t.Fatal("the refusal pattern matches an ordinary board line")
	}

	checked := 0
	var bad []string
	for _, entry := range registry {
		g := entry.NewCui()
		_, examples := helpSections(g.HelpLines())
		if len(examples) == 0 {
			continue
		}
		ctrl := g.Controller()
		ctrl.Exec("r")
		for _, line := range examples {
			cmd := helpExampleCommand(line)
			if cmd == "" {
				continue
			}
			checked++
			out := ctrl.Exec(cmd)
			if _, isErr := i18n.StripErrorPrefix(out); isErr {
				continue // TestCuiHelpExamplesExecute owns the marked ones
			}
			if line, ok := exampleRefusalLine(out); ok {
				bad = append(bad, entry.Name+": `"+cmd+"` -> "+line)
			}
		}
	}
	if checked == 0 {
		t.Fatal("no example command was executed -- the walk broke")
	}
	sort.Strings(bad)
	if len(bad) > 0 {
		t.Errorf("examples the game turns down without marking it (%d of %d):\n  %s",
			len(bad), checked, strings.Join(bad, "\n  "))
	}
}
