//go:build test

package games_test

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// The identifier guard in design_doc_identifiers_test.go reads `classDiagram`
// blocks and nothing else. That covers 12 of the 91 mermaid blocks in
// docs/design/backend.md; the other 79 -- every sequence and state diagram --
// were unchecked, and that is exactly where the drift sat.
//
// The 2026-08-15 sweep (#5350) found 18 dead identifiers in this document and
// every one of them was in a sequence or state diagram: Durak `PickUp`, Cruel
// `Redeal`, Calculation `MoveStockToFoundation`, Cassino/Scopa `NewRound` and
// `CpuStep`, PaiGow `evalPaiGowLow`, PigsTail `Draw`. The class diagrams, which
// were guarded, were clean.
//
// The three checks below close that hole. Each one resolves an *owner* before
// it resolves a name, because "the identifier exists somewhere" is not the
// question -- `FortyThievesMoveZone` exists, but not on `useKlondikeGame`, and
// that wrong-game reference is what shipped in the first fix of #5350.
//
// Precision comes from refusing to guess: a message whose target is not a real
// Go type is skipped, and a state diagram whose section heading does not name a
// real Go type is skipped. A guard that reports false positives gets switched
// off, so the checks only fire where the owner is unambiguous.

var (
	// headingRe captures the identifier in `### 3.12 Cruel フェーズ遷移`.
	headingRe = regexp.MustCompile(`^#{2,4}\s+[\d.]+\s+([A-Za-z][A-Za-z0-9]*)`)
	// fenceOpenRe/fenceCloseRe bracket a mermaid block.
	fenceOpenRe  = regexp.MustCompile("^\\s*```mermaid\\s*$")
	fenceCloseRe = regexp.MustCompile("^\\s*```\\s*$")
	// sequenceKindRe / stateKindRe identify the diagram kind.
	sequenceKindRe = regexp.MustCompile(`(?m)^\s*sequenceDiagram`)
	stateKindRe    = regexp.MustCompile(`(?m)^\s*stateDiagram-v2`)
	// participantRe matches `participant Interactor as DurakInteractor` and the
	// bare `participant Durak`. The label is what names a Go type; the alias is
	// only how the arrows refer to it.
	participantRe = regexp.MustCompile(`^\s*participant\s+(\S+)(?:\s+as\s+(.+?))?\s*$`)
	// messageLabelRe splits `Ctrl->>Interactor: TakeCards()` into target and
	// label, for every mermaid arrow spelling.
	messageLabelRe = regexp.MustCompile(`^\s*\S+\s*--?>>?\s*([^:\s]+?)\s*:\s*(.+)$`)
	// stateConstNoteRe matches `note right of Playing : FortyThievesPhasePlaying = 0`.
	// The name must start uppercase: the notes also carry unexported field
	// shorthand such as `gameEndFlag = false`, which names no Go constant.
	stateConstNoteRe = regexp.MustCompile(`^\s*note\b[^:]*:\s*([A-Z][A-Za-z0-9_]*)\s*=\s*\S+\s*$`)
	// stateTransitionLabelRe captures the label of `Playing --> GameEnd : Shift()`.
	stateTransitionLabelRe = regexp.MustCompile(`^\s*\S+\s*-->\s*\S+\s*:\s*(.+)$`)
	// callRe finds every `identifier(` in a label, so a compound label such as
	// `Reset() / NextRound()` is checked in full rather than only its first
	// call. An identifier followed by anything other than `(` is not a call:
	// that is what makes `Move*()` read as "the Move* family" and skip, which
	// is how the diagrams denote a group of sibling methods too long to list.
	//
	// The optional leading qualifier is captured so that a package-level call
	// such as `ui.RunInteractiveCuiLoop(manager)` can be skipped -- it is not a
	// method on the participant, and checking it against one would be wrong.
	callRe = regexp.MustCompile(`(?:([A-Za-z][A-Za-z0-9_]*)\.)?([A-Za-z][A-Za-z0-9_]*)\(`)
	// camelWordRe counts the words in a CamelCase identifier.
	camelWordRe = regexp.MustCompile(`[A-Z][a-z0-9]*`)
)

// docBlock is one mermaid block together with the section heading it sits
// under. The heading is what supplies the owner for a state diagram, which
// unlike a sequence diagram never names the receiver on the transition itself.
type docBlock struct {
	heading string
	body    string
}

// diagramRef is one identifier the document attributes to one owner.
type diagramRef struct {
	heading string
	owner   string
	name    string
}

func (r diagramRef) String() string {
	if r.owner == "" {
		return r.name + "  (" + r.heading + ")"
	}
	return r.owner + "." + r.name + "  (" + r.heading + ")"
}

// readDesignDoc returns the document split into mermaid blocks tagged with
// their heading.
func readDesignDoc(t *testing.T, path string) []docBlock {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot, path))
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	var blocks []docBlock
	heading, inFence := "", false
	var cur []string
	for line := range strings.SplitSeq(string(data), "\n") {
		switch {
		case !inFence && fenceOpenRe.MatchString(line):
			inFence, cur = true, nil
		case inFence && fenceCloseRe.MatchString(line):
			blocks = append(blocks, docBlock{heading: heading, body: strings.Join(cur, "\n")})
			inFence = false
		case inFence:
			cur = append(cur, line)
		default:
			if m := headingRe.FindStringSubmatch(line); m != nil {
				heading = strings.TrimSpace(strings.TrimLeft(line, "# "))
			}
		}
	}
	if len(blocks) < 50 {
		t.Fatalf("only %d mermaid blocks parsed from %s -- the fence scan broke", len(blocks), path)
	}
	return blocks
}

// headingOwner returns the identifier a section heading names, e.g. `Cruel`
// for `### 3.12 Cruel フェーズ遷移`. It returns "" when the heading is prose or
// names something that is not a bare identifier (`Texas Hold'em`).
func headingOwner(heading string) string {
	if m := headingRe.FindStringSubmatch("### " + heading); m != nil {
		return m[1]
	}
	return ""
}

// sequenceCalls returns every `Target: Method(` message whose target resolves,
// through the participant list, to a name the document itself declares.
func sequenceCalls(blocks []docBlock) []diagramRef {
	var refs []diagramRef
	for _, b := range blocks {
		if !sequenceKindRe.MatchString(b.body) {
			continue
		}
		label := map[string]string{}
		for line := range strings.SplitSeq(b.body, "\n") {
			if m := participantRe.FindStringSubmatch(line); m != nil {
				name := m[1]
				if m[2] != "" {
					name = strings.TrimSpace(m[2])
				}
				label[m[1]] = name
			}
		}
		for line := range strings.SplitSeq(b.body, "\n") {
			m := messageLabelRe.FindStringSubmatch(line)
			if m == nil {
				continue
			}
			target := m[1]
			if l, ok := label[target]; ok {
				target = l
			}
			for _, c := range callRe.FindAllStringSubmatch(m[2], -1) {
				if c[1] != "" {
					continue // `pkg.Func()` -- not a method on this participant
				}
				refs = append(refs, diagramRef{heading: b.heading, owner: target, name: c[2]})
			}
		}
	}
	return refs
}

// stateConstantRefs returns the constants declared in state-diagram notes.
//
// A note reads `note right of Playing : FortyThievesPhasePlaying = 0`. Only
// multi-word CamelCase names are treated as constant references: the diagrams
// also carry shorthand notes such as `Phase = 1` and `gameEndFlag = false`
// that name no Go symbol.
func stateConstantRefs(blocks []docBlock) []diagramRef {
	var refs []diagramRef
	for _, b := range blocks {
		if !stateKindRe.MatchString(b.body) {
			continue
		}
		for line := range strings.SplitSeq(b.body, "\n") {
			m := stateConstNoteRe.FindStringSubmatch(line)
			if m == nil {
				continue
			}
			if len(camelWordRe.FindAllString(m[1], -1)) < 2 {
				continue
			}
			refs = append(refs, diagramRef{heading: b.heading, name: m[1]})
		}
	}
	return refs
}

// stateTransitionRefs returns every `Method()` transition label, attributed to
// the game its section heading names.
func stateTransitionRefs(blocks []docBlock) []diagramRef {
	var refs []diagramRef
	for _, b := range blocks {
		if !stateKindRe.MatchString(b.body) {
			continue
		}
		owner := headingOwner(b.heading)
		if owner == "" {
			continue
		}
		for line := range strings.SplitSeq(b.body, "\n") {
			m := stateTransitionLabelRe.FindStringSubmatch(line)
			if m == nil {
				continue
			}
			for _, c := range callRe.FindAllStringSubmatch(m[1], -1) {
				if c[1] != "" {
					continue // `pkg.Func()` -- not a method on this game
				}
				refs = append(refs, diagramRef{heading: b.heading, owner: owner, name: c[2]})
			}
		}
	}
	return refs
}

// TestDesignDocSequenceCallsExistInGo asserts that every method a sequence
// diagram sends to a participant that names a real Go type exists on that type.
//
// This is what would have caught Durak `PickUp()` (renamed to `TakeCards`) and
// PaiGow `evalPaiGowLow` (never existed; the function is `evalPaiGowLowHand`).
// Participants such as `Controller`, `Presenter` and `ユーザー` name no Go type
// and are skipped -- the check fires only where the receiver is unambiguous.
func TestDesignDocSequenceCallsExistInGo(t *testing.T) {
	blocks := readDesignDoc(t, designDocPath)
	surface := parseGoSurface(t, goSourceRoot)

	checked := 0
	var missing []string
	for _, ref := range sequenceCalls(blocks) {
		if !surface.types[ref.owner] {
			continue // a prose participant, not a Go type
		}
		checked++
		if !surface.has(ref.owner, ref.name, map[string]bool{}) {
			missing = append(missing, ref.String())
		}
	}
	if checked < 60 {
		t.Fatalf("only %d sequence calls checked in %s -- participantRe or messageCallRe stopped matching",
			checked, designDocPath)
	}
	sort.Strings(missing)

	if len(missing) > 0 {
		t.Errorf("sequence-diagram calls in %s with no such method on the target type (%d of %d checked):\n  %s\n"+
			"Rename them in the diagram to match the code.",
			designDocPath, len(missing), checked, strings.Join(missing, "\n  "))
	}
}

// TestDesignDocStateConstantsExistInGo asserts that every constant named in a
// state-diagram note exists in Go.
//
// This is what would have caught `SevenCardStudPhaseThirdSt` (the real name
// ends `ThirdStreet`), `FortyThievesPlaying` (`FortyThievesPhasePlaying`) and
// `LetItRidePhaseDecision1` (`LetItRidePhaseFirstDecision`) -- three families
// whose *values* were right, so nothing else would ever have flagged them.
func TestDesignDocStateConstantsExistInGo(t *testing.T) {
	blocks := readDesignDoc(t, designDocPath)
	surface := parseGoSurface(t, goSourceRoot)

	refs := stateConstantRefs(blocks)
	if len(refs) < 120 {
		t.Fatalf("only %d state-diagram constants parsed from %s -- stateConstNoteRe stopped matching",
			len(refs), designDocPath)
	}

	var missing []string
	for _, ref := range refs {
		if !surface.consts[ref.name] {
			missing = append(missing, ref.String())
		}
	}
	sort.Strings(missing)

	if len(missing) > 0 {
		t.Errorf("constants in %s state-diagram notes that do not exist in Go (%d of %d checked):\n  %s\n"+
			"Rename them in the diagram to match the code.",
			designDocPath, len(missing), len(refs), strings.Join(missing, "\n  "))
	}
}

// TestDesignDocStateTransitionsExistInGo asserts that every `Method()`
// transition label names a method on the game its section heading identifies,
// or on that game's interactor.
//
// This is what would have caught Cruel `Redeal()` (the method is `Shift`),
// Cassino/PageOne/SevenBridge/Scopa `NewRound()` (`NextRound`) and Scopa/Barbu
// `CpuStep()` (`CpuPlay`). Note that grepping for those names without a
// receiver finds hundreds of hits and concludes they exist -- `NextRound` alone
// is defined 578 times. The owner is the whole point of the check.
func TestDesignDocStateTransitionsExistInGo(t *testing.T) {
	blocks := readDesignDoc(t, designDocPath)
	surface := parseGoSurface(t, goSourceRoot)

	checked := 0
	var missing []string
	for _, ref := range stateTransitionRefs(blocks) {
		owners := []string{ref.owner, ref.owner + "Interactor"}
		known := false
		for _, o := range owners {
			if surface.types[o] {
				known = true
			}
		}
		if !known {
			continue // heading names no Go type (e.g. a cross-cutting section)
		}
		checked++
		found := false
		for _, o := range owners {
			if surface.types[o] && surface.has(o, ref.name, map[string]bool{}) {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, ref.String())
		}
	}
	if checked < 120 {
		t.Fatalf("only %d state transitions checked in %s -- stateTransitionRe or headingRe stopped matching",
			checked, designDocPath)
	}
	sort.Strings(missing)

	if len(missing) > 0 {
		t.Errorf("state-diagram transitions in %s with no such method on the game or its interactor (%d of %d checked):\n  %s\n"+
			"Rename them in the diagram to match the code.",
			designDocPath, len(missing), checked, strings.Join(missing, "\n  "))
	}
}

// TestDesignDocDiagramParsersCatchBreakage is the negative control for the
// three checks above.
//
// A guard that never fires is indistinguishable from a guard that passes, and
// the count floors only prove that *something* was parsed. These synthetic
// diagrams carry a known-bad identifier each; if a parser silently stops
// matching, the corresponding case here goes green and says so.
func TestDesignDocDiagramParsersCatchBreakage(t *testing.T) {
	const broken = "```mermaid\n" +
		"sequenceDiagram\n" +
		"    participant Ctrl as Controller\n" +
		"    participant Interactor as DurakInteractor\n" +
		"    Ctrl->>Interactor: NoSuchMethod()\n" +
		"```\n"

	blocks := []docBlock{{heading: "3.1 Durak フェーズ遷移", body: strings.TrimSuffix(
		strings.TrimPrefix(broken, "```mermaid\n"), "```\n")}}

	calls := sequenceCalls(blocks)
	if len(calls) != 1 {
		t.Fatalf("sequenceCalls found %d calls in the synthetic diagram, want 1", len(calls))
	}
	if calls[0].owner != "DurakInteractor" || calls[0].name != "NoSuchMethod" {
		t.Errorf("sequenceCalls resolved %v, want DurakInteractor.NoSuchMethod -- alias resolution broke", calls[0])
	}
	surface := parseGoSurface(t, goSourceRoot)
	if surface.has("DurakInteractor", "NoSuchMethod", map[string]bool{}) {
		t.Error("surface claims DurakInteractor.NoSuchMethod exists -- the resolver is too permissive")
	}

	stateBlocks := []docBlock{{
		heading: "3.1 Cruel フェーズ遷移",
		body: "stateDiagram-v2\n" +
			"    Playing --> GameOver : NoSuchAction()\n" +
			"    note right of Playing : NoSuchPhaseConstant = 0\n" +
			"    note right of Playing : Phase = 1\n",
	}}

	trans := stateTransitionRefs(stateBlocks)
	if len(trans) != 1 || trans[0].owner != "Cruel" || trans[0].name != "NoSuchAction" {
		t.Errorf("stateTransitionRefs resolved %v, want one Cruel.NoSuchAction -- heading attribution broke", trans)
	}

	consts := stateConstantRefs(stateBlocks)
	if len(consts) != 1 || consts[0].name != "NoSuchPhaseConstant" {
		t.Errorf("stateConstantRefs resolved %v, want only NoSuchPhaseConstant -- the single-word filter broke", consts)
	}
	if surface.consts["NoSuchPhaseConstant"] {
		t.Error("surface claims NoSuchPhaseConstant exists -- the const collector is too permissive")
	}
}

// TestDesignDocDiagramParsersAcceptCorrectInput is the positive control: the
// parsers must stay quiet on diagrams that match the code.
//
// Without this, tightening a regex until it matches nothing would look like a
// clean bill of health in every other test here.
func TestDesignDocDiagramParsersAcceptCorrectInput(t *testing.T) {
	surface := parseGoSurface(t, goSourceRoot)

	good := []docBlock{{
		heading: "2.14 Durak アタック・ディフェンスフロー",
		body: "sequenceDiagram\n" +
			"    participant Interactor as DurakInteractor\n" +
			"    participant Domain as Durak\n" +
			"    Ctrl->>Interactor: TakeCards()\n" +
			"    Interactor->>Domain: PlayerTakeCards()\n" +
			"    Domain-->>Interactor: nil\n" +
			"    Domain->>Domain: カード出す → テーブルに配置\n",
	}}

	calls := sequenceCalls(good)
	if len(calls) != 2 {
		t.Fatalf("sequenceCalls found %d calls, want 2 -- prose labels and `nil` returns must not be read as calls: %v", len(calls), calls)
	}
	for _, c := range calls {
		if !surface.has(c.owner, c.name, map[string]bool{}) {
			t.Errorf("%v reported missing, but it exists in Go -- the check would fire on a correct diagram", c)
		}
	}
}
