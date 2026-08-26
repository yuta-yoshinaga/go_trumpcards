//go:build test

package ui

import (
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// wrongPhaseNeedle is the sentence domain.ErrWrongPhase carries. Phase
// rejections are the one refusal shape that can be recognised exactly instead
// of by tone: every one of them comes from this single sentinel, so this guard
// does not depend on a wording heuristic the way
// TestCuiHelpExamplesAreNotQuietlyRefused has to.
var wrongPhaseNeedle = domain.ErrWrongPhase.Error()

// runExamples deals a fresh game and executes its examples in order, reporting
// which of them the game turned down as being for another phase.
func runExamples(entry GameRegistryEntry, cmds []string) []bool {
	g := entry.NewCui()
	ctrl := g.Controller()
	ctrl.Exec("r")
	refused := make([]bool, len(cmds))
	for i, cmd := range cmds {
		refused[i] = strings.Contains(ctrl.Exec(cmd), wrongPhaseNeedle)
	}
	return refused
}

// exampleCommands pulls a game's examples out of its rendered help.
func exampleCommands(entry GameRegistryEntry) []string {
	_, examples := helpSections(entry.NewCui().HelpLines())
	cmds := make([]string, 0, len(examples))
	for _, line := range examples {
		if cmd := helpExampleCommand(line); cmd != "" {
			cmds = append(cmds, cmd)
		}
	}
	return cmds
}

// TestCuiHelpExamplesAreReachable fails on an example that can never be typed:
// one the game answers "wrong game phase" to on every deal.
//
// This is the refusal the two older guards are blind to.
// TestCuiHelpExamplesExecute asks i18n.StripErrorPrefix, and a phase rejection
// is rendered as a red line *inside* the board rather than as the whole reply,
// so it never looks like an error there. TestCuiHelpExamplesAreNotQuietlyRefused
// matches on wording, and `wrong game phase` contains none of the words in
// exampleRefusalRe -- so 24 games shipped examples that a player typing them
// straight after `r` could not use, euchre's `p 0` among them: it needs the
// pickup and discard phases dealt with first, and the help never said so.
//
// Why "on every deal" and not "on this deal": which phase a game is in right
// after `r` is deal-dependent, because the CPUs act first. Euchre lands in
// Pickup about 40% of the time and in Discard the rest, so its `o` example is
// legal in some deals and not in others -- and an assertion of zero rejections
// would fail at that rate rather than catch a defect. What is never legitimate
// is an example no deal can reach.
//
// The retry pass runs only against the examples the first pass caught, so the
// 30 extra deals cost a handful of games rather than a second full sweep.
func TestCuiHelpExamplesAreReachable(t *testing.T) {
	registry := GameRegistry()
	if len(registry) < 300 {
		t.Fatalf("only %d games in the registry -- the walk broke", len(registry))
	}
	if wrongPhaseNeedle == "" {
		t.Fatal("domain.ErrWrongPhase carries no message; this guard would match everything")
	}

	// Negative control. Euchre's `p 0` is refused whenever the deal leaves the
	// human in the pickup phase, which is the exact shape this guard hunts. If
	// the needle stops matching, every game passes and the guard proves nothing.
	var euchre *GameRegistryEntry
	for i := range registry {
		if registry[i].Name == "euchre" {
			euchre = &registry[i]
			break
		}
	}
	if euchre == nil {
		t.Fatal("euchre is gone from the registry -- pick another negative control")
	}
	sawRejection := false
	for d := 0; d < 40 && !sawRejection; d++ {
		sawRejection = runExamples(*euchre, []string{"p 0"})[0]
	}
	if !sawRejection {
		t.Fatal("euchre never answered a bare `p 0` with a phase rejection in 40 deals -- " +
			"the needle no longer matches, so this guard cannot fail")
	}

	// Pass 1: one deal per game, to find the suspects cheaply.
	type suspect struct {
		entry GameRegistryEntry
		cmds  []string
		idx   int
	}
	var suspects []suspect
	games, executed := 0, 0
	for _, entry := range registry {
		cmds := exampleCommands(entry)
		if len(cmds) == 0 {
			continue
		}
		games++
		executed += len(cmds)
		for i, refused := range runExamples(entry, cmds) {
			if refused {
				suspects = append(suspects, suspect{entry, cmds, i})
			}
		}
	}
	if games == 0 || executed == 0 {
		t.Fatal("no examples were executed -- helpSections or helpExampleCommandRe stopped matching")
	}

	// Pass 2: only the suspects, over enough fresh deals that a branch the CPUs
	// sometimes take away is separated from one that does not exist. An example
	// legal 30% of the time survives all 30 deals with probability 0.7^30, which
	// is about 2 in 100,000.
	const retries = 30
	var dead []string
	for _, s := range suspects {
		reached := false
		for d := 0; d < retries && !reached; d++ {
			reached = !runExamples(s.entry, s.cmds)[s.idx]
		}
		if !reached {
			dead = append(dead, s.entry.Name+": `"+s.cmds[s.idx]+"` is for a phase the game never reaches from a fresh deal"+
				" (example "+strconv.Itoa(s.idx+1)+" of "+strconv.Itoa(len(s.cmds))+")")
		}
	}

	sort.Strings(dead)
	if len(dead) > 0 {
		t.Errorf("examples no player can type (%d of %d executed across %d games).\n"+
			"An earlier phase is missing from the example list: add the bid, exchange or\n"+
			"discard step in front of it so the sequence reaches the phase it documents.\n  %s",
			len(dead), executed, games, strings.Join(dead, "\n  "))
	}
}
