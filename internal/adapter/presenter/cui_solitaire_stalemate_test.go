//go:build test

package presenter

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// escapeMoves is an arbitrary positive undo count: the guidance only renders when
// UndoToEscape() > 0, so 0 would silently prove nothing.
const escapeMoves = 3

func assertEscapeHint(t *testing.T, result string) {
	t.Helper()
	assert.Contains(t, result, "手詰まりです",
		"stalemate banner missing -- the fixture did not reach the stalemate branch")
	assert.Contains(t, result, "脱出には",
		"stalemate screen does not tell the player how many undos escape the dead end")
	assert.Contains(t, result, strconv.Itoa(escapeMoves),
		"escape guidance does not carry the undo count from UndoToEscape()")
}

// TestSolitaireCuiUndoToEscapeHint covers the solitaires whose web page offers a
// StalemateEscapeButton while their CUI said only "手詰まりです".
func TestSolitaireCuiUndoToEscapeHint(t *testing.T) {
	t.Run("Accordion", func(t *testing.T) {
		m := new(interfaces.MockAccordionGame)
		setupAccordionCuiMockDefaults(m)
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "IsStalemate")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "UndoToEscape")
		m.On("GetPhase").Return(domain.AccordionPhasePlaying)
		m.On("IsStalemate").Return(true)
		m.On("UndoToEscape").Return(escapeMoves)

		assertEscapeHint(t, new(AccordionCuiPresenter).Output(m, nil))
	})
	t.Run("AcesUp", func(t *testing.T) {
		m := new(interfaces.MockAcesUpGame)
		setupAcesUpCuiMockDefaults(m)
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "IsStalemate")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "UndoToEscape")
		m.On("GetPhase").Return(domain.AcesUpPhasePlaying)
		m.On("IsStalemate").Return(true)
		m.On("UndoToEscape").Return(escapeMoves)

		assertEscapeHint(t, new(AcesUpCuiPresenter).Output(m, nil))
	})
	t.Run("AuldLangSyne", func(t *testing.T) {
		m := new(interfaces.MockAuldLangSyneGame)
		setupAuldLangSyneCuiMockDefaults(m)
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "IsStalemate")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "UndoToEscape")
		m.On("GetPhase").Return(domain.AuldLangSynePhasePlaying)
		m.On("IsStalemate").Return(true)
		m.On("UndoToEscape").Return(escapeMoves)

		assertEscapeHint(t, new(AuldLangSyneCuiPresenter).Output(m, nil))
	})
	t.Run("BakersDozen", func(t *testing.T) {
		m := new(interfaces.MockBakersDozenGame)
		setupBakersDozenCuiMockDefaults(m)
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "IsStalemate")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "UndoToEscape")
		m.On("GetPhase").Return(domain.BakersDozenPhasePlaying)
		m.On("IsStalemate").Return(true)
		m.On("UndoToEscape").Return(escapeMoves)

		assertEscapeHint(t, new(BakersDozenCuiPresenter).Output(m, nil))
	})
	t.Run("Calculation", func(t *testing.T) {
		m := new(interfaces.MockCalculationGame)
		setupCalculationCuiMockDefaults(m)
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "IsStalemate")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "UndoToEscape")
		m.On("GetPhase").Return(domain.CalculationPhasePlaying)
		m.On("IsStalemate").Return(true)
		m.On("UndoToEscape").Return(escapeMoves)

		assertEscapeHint(t, new(CalculationCuiPresenter).Output(m, nil))
	})
	t.Run("Crescent", func(t *testing.T) {
		m := new(interfaces.MockCrescentGame)
		setupCrescentCuiMockDefaults(m)
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "IsStalemate")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "UndoToEscape")
		m.On("GetPhase").Return(domain.CrescentPhasePlaying)
		m.On("IsStalemate").Return(true)
		m.On("UndoToEscape").Return(escapeMoves)

		assertEscapeHint(t, new(CrescentCuiPresenter).Output(m, nil))
	})
	t.Run("Cruel", func(t *testing.T) {
		m := new(interfaces.MockCruelGame)
		setupCruelCuiMockDefaults(m)
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "IsStalemate")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "UndoToEscape")
		m.On("GetPhase").Return(domain.CruelPhasePlaying)
		m.On("IsStalemate").Return(true)
		m.On("UndoToEscape").Return(escapeMoves)

		assertEscapeHint(t, new(CruelCuiPresenter).Output(m, nil))
	})
	t.Run("Easthaven", func(t *testing.T) {
		m := new(interfaces.MockEasthavenGame)
		setupEasthavenCuiMockDefaults(m)
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "IsStalemate")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "UndoToEscape")
		m.On("GetPhase").Return(domain.EasthavenPhasePlaying)
		m.On("IsStalemate").Return(true)
		m.On("UndoToEscape").Return(escapeMoves)

		assertEscapeHint(t, new(EasthavenCuiPresenter).Output(m, nil))
	})
	t.Run("FlowerGarden", func(t *testing.T) {
		m := new(interfaces.MockFlowerGardenGame)
		setupFlowerGardenCuiMockDefaults(m)
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "IsStalemate")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "UndoToEscape")
		m.On("GetPhase").Return(domain.FlowerGardenPhasePlaying)
		m.On("IsStalemate").Return(true)
		m.On("UndoToEscape").Return(escapeMoves)

		assertEscapeHint(t, new(FlowerGardenCuiPresenter).Output(m, nil))
	})
	t.Run("FortyAndEight", func(t *testing.T) {
		m := new(interfaces.MockFortyAndEightGame)
		setupFortyAndEightCuiMockDefaults(m)
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "IsStalemate")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "UndoToEscape")
		m.On("GetPhase").Return(domain.FortyAndEightPhasePlaying)
		m.On("IsStalemate").Return(true)
		m.On("UndoToEscape").Return(escapeMoves)

		assertEscapeHint(t, new(FortyAndEightCuiPresenter).Output(m, nil))
	})
	t.Run("FortyThieves", func(t *testing.T) {
		m := new(interfaces.MockFortyThievesGame)
		setupFortyThievesCuiMockDefaults(m)
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "IsStalemate")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "UndoToEscape")
		m.On("GetPhase").Return(domain.FortyThievesPhasePlaying)
		m.On("IsStalemate").Return(true)
		m.On("UndoToEscape").Return(escapeMoves)

		assertEscapeHint(t, new(FortyThievesCuiPresenter).Output(m, nil))
	})
	t.Run("Golf", func(t *testing.T) {
		m := new(interfaces.MockGolfGame)
		setupGolfCuiMockDefaults(m)
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "IsStalemate")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "UndoToEscape")
		m.On("GetPhase").Return(domain.GolfPhasePlaying)
		m.On("IsStalemate").Return(true)
		m.On("UndoToEscape").Return(escapeMoves)

		assertEscapeHint(t, new(GolfCuiPresenter).Output(m, nil))
	})
	t.Run("Klondike", func(t *testing.T) {
		m := new(interfaces.MockKlondikeGame)
		setupKlondikeCuiMockDefaults(m)
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "IsStalemate")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "UndoToEscape")
		m.On("GetPhase").Return(domain.KlondikePhasePlaying)
		m.On("IsStalemate").Return(true)
		m.On("UndoToEscape").Return(escapeMoves)

		assertEscapeHint(t, new(KlondikeCuiPresenter).Output(m, nil))
	})
	t.Run("RussianSolitaire", func(t *testing.T) {
		m := new(interfaces.MockRussianSolitaireGame)
		setupRussianSolitaireCuiMockDefaults(m)
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "IsStalemate")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "UndoToEscape")
		m.On("GetPhase").Return(domain.RussianSolitairePhasePlaying)
		m.On("IsStalemate").Return(true)
		m.On("UndoToEscape").Return(escapeMoves)

		assertEscapeHint(t, new(RussianSolitaireCuiPresenter).Output(m, nil))
	})
	t.Run("Scorpion", func(t *testing.T) {
		m := new(interfaces.MockScorpionGame)
		setupScorpionCuiMockDefaults(m)
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "IsStalemate")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "UndoToEscape")
		m.On("GetPhase").Return(domain.ScorpionPhasePlaying)
		m.On("IsStalemate").Return(true)
		m.On("UndoToEscape").Return(escapeMoves)

		assertEscapeHint(t, new(ScorpionCuiPresenter).Output(m, nil))
	})
	t.Run("SirTommy", func(t *testing.T) {
		m := new(interfaces.MockSirTommyGame)
		setupSirTommyCuiMockDefaults(m)
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "IsStalemate")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "UndoToEscape")
		m.On("GetPhase").Return(domain.SirTommyPhasePlaying)
		m.On("IsStalemate").Return(true)
		m.On("UndoToEscape").Return(escapeMoves)

		assertEscapeHint(t, new(SirTommyCuiPresenter).Output(m, nil))
	})
	t.Run("Spider", func(t *testing.T) {
		m := new(interfaces.MockSpiderGame)
		setupSpiderCuiMockDefaults(m)
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "IsStalemate")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "UndoToEscape")
		m.On("GetPhase").Return(domain.SpiderPhasePlaying)
		m.On("IsStalemate").Return(true)
		m.On("UndoToEscape").Return(escapeMoves)

		assertEscapeHint(t, new(SpiderCuiPresenter).Output(m, nil))
	})
	t.Run("Spiderette", func(t *testing.T) {
		m := new(interfaces.MockSpideretteGame)
		setupSpideretteCuiMockDefaults(m)
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "IsStalemate")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "UndoToEscape")
		m.On("GetPhase").Return(domain.SpiderettePhasePlaying)
		m.On("IsStalemate").Return(true)
		m.On("UndoToEscape").Return(escapeMoves)

		assertEscapeHint(t, new(SpideretteCuiPresenter).Output(m, nil))
	})
	t.Run("StreetsAndAlleys", func(t *testing.T) {
		m := new(interfaces.MockStreetsAndAlleysGame)
		setupStreetsAndAlleysCuiMockDefaults(m)
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "IsStalemate")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "UndoToEscape")
		m.On("GetPhase").Return(domain.StreetsAndAlleysPhasePlaying)
		m.On("IsStalemate").Return(true)
		m.On("UndoToEscape").Return(escapeMoves)

		assertEscapeHint(t, new(StreetsAndAlleysCuiPresenter).Output(m, nil))
	})
	t.Run("TriPeaks", func(t *testing.T) {
		m := new(interfaces.MockTriPeaksGame)
		setupTriPeaksCuiMockDefaults(m)
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "IsStalemate")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "UndoToEscape")
		m.On("GetPhase").Return(domain.TriPeaksPhasePlaying)
		m.On("IsStalemate").Return(true)
		m.On("UndoToEscape").Return(escapeMoves)

		assertEscapeHint(t, new(TriPeaksCuiPresenter).Output(m, nil))
	})
	t.Run("Wasp", func(t *testing.T) {
		m := new(interfaces.MockWaspGame)
		setupWaspCuiMockDefaults(m)
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "IsStalemate")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "UndoToEscape")
		m.On("GetPhase").Return(domain.WaspPhasePlaying)
		m.On("IsStalemate").Return(true)
		m.On("UndoToEscape").Return(escapeMoves)

		assertEscapeHint(t, new(WaspCuiPresenter).Output(m, nil))
	})
	t.Run("Yukon", func(t *testing.T) {
		m := new(interfaces.MockYukonGame)
		setupYukonCuiMockDefaults(m)
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "IsStalemate")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "UndoToEscape")
		m.On("GetPhase").Return(domain.YukonPhasePlaying)
		m.On("IsStalemate").Return(true)
		m.On("UndoToEscape").Return(escapeMoves)

		assertEscapeHint(t, new(YukonCuiPresenter).Output(m, nil))
	})
	// Pyramid and Gaps are mock-based like the 22 above, but their fixtures use a
	// differently named constructor (setupXCuiMock returning the mock) rather than a
	// setupXCuiMockDefaults(mock) helper, so they are wired here by hand.
	t.Run("Pyramid", func(t *testing.T) {
		m := setupPyramidCuiMock()
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "IsStalemate")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "UndoToEscape")
		m.On("GetPhase").Return(domain.PyramidPhasePlaying)
		m.On("IsStalemate").Return(true)
		m.On("UndoToEscape").Return(escapeMoves)

		assertEscapeHint(t, new(PyramidCuiPresenter).Output(m, nil))
	})

	t.Run("Gaps", func(t *testing.T) {
		m := setupGapsCuiMock()
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "IsStalemate")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "UndoToEscape")
		m.On("GetPhase").Return(domain.GapsPhasePlaying)
		m.On("IsStalemate").Return(true)
		m.On("UndoToEscape").Return(escapeMoves)

		assertEscapeHint(t, new(GapsCuiPresenter).Output(m, nil))
	})
}

// TestEverySolitaireStalemateBranchReportsTheEscapeCount is a structural guard.
//
// Four of the wired presenters (BakersGame, EightOff, FreeCell, SeahavenTowers)
// drive their tests from real domain objects rather than mocks, and a
// freshly reset game has no undo history, so UndoToEscape() returns 0 and the
// behavioural test above cannot reach them. This parses every CUI presenter instead.
// (Penguin mixes both styles and is likewise only covered structurally.)
//
// The rule is keyed on capability, not on a hardcoded list of game names: a presenter
// must report UndoToEscape() in its IsStalemate() branch exactly when its own game
// interface actually exposes that method. That way the guard needs no allowlist to
// excuse Agnes / MonteCarlo / Osmosis / Bhabhi (whose interfaces do not expose it, so
// they *cannot* report it), and it still fails the moment one of them gains the method
// without reporting it -- or a brand-new solitaire presenter forgets the guidance.
func TestEverySolitaireStalemateBranchReportsTheEscapeCount(t *testing.T) {
	files, err := filepath.Glob("*CuiPresenter.go")
	assert.NoError(t, err)
	assert.NotEmpty(t, files, "no presenters found -- the glob is wrong, not the code")

	capable, missing, skipped := 0, []string{}, 0
	for _, path := range files {
		srcBytes, err := os.ReadFile(path) //nolint:gosec // test-only, fixed glob
		assert.NoError(t, err)
		src := string(srcBytes)
		if !strings.Contains(src, "IsStalemate()") {
			continue
		}
		if !gameInterfaceExposesUndoToEscape(t, src) {
			skipped++
			continue
		}
		capable++

		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, src, 0)
		assert.NoError(t, err)

		found := false
		ast.Inspect(f, func(n ast.Node) bool {
			ifStmt, ok := n.(*ast.IfStmt)
			if !ok || !strings.Contains(exprText(fset, src, ifStmt.Cond), "IsStalemate()") {
				return true
			}
			if strings.Contains(exprText(fset, src, ifStmt.Body), "UndoToEscape()") {
				found = true
			}
			return true
		})
		if !found {
			missing = append(missing, path)
		}
	}

	// Negative control on the detection itself: if these counters collapse the guard
	// is inspecting nothing and would pass no matter how broken the presenters were.
	assert.Greater(t, capable, 40,
		"far fewer capable presenters than expected -- the detection is broken, not the code")
	assert.Greater(t, skipped, 0,
		"expected some stalemate presenters whose interface cannot report an escape count")
	assert.Empty(t, missing,
		"these presenters branch on IsStalemate() and their interface exposes UndoToEscape(), but they never report it")
}

// gameInterfaceExposesUndoToEscape reports whether the game interface a presenter's
// Output takes declares UndoToEscape, directly or via the embedded SolitaireGame.
func gameInterfaceExposesUndoToEscape(t *testing.T, presenterSrc string) bool {
	t.Helper()
	m := regexp.MustCompile(`Output\(\w+ interfaces\.(\w+)Game`).FindStringSubmatch(presenterSrc)
	if m == nil {
		return false
	}
	ifacePath := filepath.Join("..", "..", "domain", "interfaces", strings.ToLower(m[1])+".go")
	b, err := os.ReadFile(ifacePath) //nolint:gosec // test-only, derived from a matched identifier
	if err != nil {
		return false
	}
	body := string(b)
	return strings.Contains(body, "UndoToEscape() int") || strings.Contains(body, "SolitaireGame")
}

// exprText returns the source text spanned by a node.
func exprText(fset *token.FileSet, src string, n ast.Node) string {
	return src[fset.Position(n.Pos()).Offset:fset.Position(n.End()).Offset]
}
