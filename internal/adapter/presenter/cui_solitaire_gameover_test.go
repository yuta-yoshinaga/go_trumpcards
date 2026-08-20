//go:build test

package presenter

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// assertGameOverSummary checks that a solitaire CUI game-over screen carries the
// foundation-progress line the web page shows.
//
// It pins the denominator rather than the count: the count depends on each
// game's mock fixture, but the denominator is the game's deck size, so a
// presenter wired to the wrong total constant (52 vs 104) fails here. The
// count/percent arithmetic itself is pinned in cui_solitaire_helper_test.go.
func assertGameOverSummary(t *testing.T, result string, total int) {
	t.Helper()
	assert.Contains(t, result, "まで到達",
		"game over screen is missing the foundation-progress line")
	assert.Contains(t, result, "/"+strconv.Itoa(total)+" 枚",
		"foundation-progress line reports the wrong deck total")
	// The progress line must come after the game-over banner, not replace it.
	assert.Contains(t, result, "ゲームオーバー")
	assert.Less(t, strings.Index(result, "ゲームオーバー"), strings.Index(result, "まで到達"))
}

// TestSolitaireCuiGameOverSummary covers every solitaire whose web page renders a
// gameOverSummary; before this the CUI printed only "ゲームオーバー" and dropped the
// near-miss detail entirely.
func TestSolitaireCuiGameOverSummary(t *testing.T) {
	t.Run("AmericanToad", func(t *testing.T) {
		g := new(interfaces.MockAmericanToadGame)
		setupAmericanToadCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
		g.On("GetPhase").Return(domain.AmericanToadPhaseGameOver)

		result := new(AmericanToadCuiPresenter).Output(g, nil)
		assertGameOverSummary(t, result, domain.AmericanToadTotalCards)
	})
	t.Run("BeleagueredCastle", func(t *testing.T) {
		g := new(interfaces.MockBeleagueredCastleGame)
		setupBeleagueredCastleCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
		g.On("GetPhase").Return(domain.BeleagueredCastlePhaseGameOver)

		result := new(BeleagueredCastleCuiPresenter).Output(g, nil)
		assertGameOverSummary(t, result, domain.BeleagueredCastleFoundationCnt*domain.CardValueMax)
	})
	t.Run("Bisley", func(t *testing.T) {
		g := new(interfaces.MockBisleyGame)
		setupBisleyCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
		g.On("GetPhase").Return(domain.BisleyPhaseGameOver)

		result := new(BisleyCuiPresenter).Output(g, nil)
		assertGameOverSummary(t, result, domain.BisleyFoundationCnt*domain.CardValueMax)
	})
	t.Run("Braid", func(t *testing.T) {
		g := new(interfaces.MockBraidGame)
		setupBraidCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
		g.On("GetPhase").Return(domain.BraidPhaseGameOver)

		result := new(BraidCuiPresenter).Output(g, nil)
		assertGameOverSummary(t, result, domain.BraidTotalCards)
	})
	t.Run("Congress", func(t *testing.T) {
		g := new(interfaces.MockCongressGame)
		setupCongressCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
		g.On("GetPhase").Return(domain.CongressPhaseGameOver)

		result := new(CongressCuiPresenter).Output(g, nil)
		assertGameOverSummary(t, result, domain.CongressTotalCards)
	})
	t.Run("CrazyQuilt", func(t *testing.T) {
		g := new(interfaces.MockCrazyQuiltGame)
		setupCrazyQuiltCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
		g.On("GetPhase").Return(domain.CrazyQuiltPhaseGameOver)

		result := new(CrazyQuiltCuiPresenter).Output(g, nil)
		assertGameOverSummary(t, result, domain.CrazyQuiltTotalCards)
	})
	t.Run("Diplomat", func(t *testing.T) {
		g := new(interfaces.MockDiplomatGame)
		setupDiplomatCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
		g.On("GetPhase").Return(domain.DiplomatPhaseGameOver)

		result := new(DiplomatCuiPresenter).Output(g, nil)
		assertGameOverSummary(t, result, domain.DiplomatTotalCards)
	})
	t.Run("Duchess", func(t *testing.T) {
		g := new(interfaces.MockDuchessGame)
		setupDuchessCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
		g.On("GetPhase").Return(domain.DuchessPhaseGameOver)

		result := new(DuchessCuiPresenter).Output(g, nil)
		assertGameOverSummary(t, result, domain.DuchessFoundationCnt*domain.CardValueMax)
	})
	t.Run("FlowerGarden", func(t *testing.T) {
		g := new(interfaces.MockFlowerGardenGame)
		setupFlowerGardenCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
		g.On("GetPhase").Return(domain.FlowerGardenPhaseGameOver)

		result := new(FlowerGardenCuiPresenter).Output(g, nil)
		assertGameOverSummary(t, result, domain.FlowerGardenFoundationCnt*domain.CardValueMax)
	})
	t.Run("KingAlbert", func(t *testing.T) {
		g := new(interfaces.MockKingAlbertGame)
		setupKingAlbertCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
		g.On("GetPhase").Return(domain.KingAlbertPhaseGameOver)

		result := new(KingAlbertCuiPresenter).Output(g, nil)
		assertGameOverSummary(t, result, domain.KingAlbertFoundationCnt*domain.CardValueMax)
	})
	t.Run("MissMilligan", func(t *testing.T) {
		g := new(interfaces.MockMissMilliganGame)
		setupMissMilliganCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
		g.On("GetPhase").Return(domain.MissMilliganPhaseGameOver)

		result := new(MissMilliganCuiPresenter).Output(g, nil)
		assertGameOverSummary(t, result, domain.MissMilliganFoundationCnt*domain.CardValueMax)
	})
	t.Run("NapoleonsSquare", func(t *testing.T) {
		g := new(interfaces.MockNapoleonsSquareGame)
		setupNapoleonsSquareCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
		g.On("GetPhase").Return(domain.NapoleonsSquarePhaseGameOver)

		result := new(NapoleonsSquareCuiPresenter).Output(g, nil)
		assertGameOverSummary(t, result, domain.NapoleonsSquareFoundationCnt*domain.CardValueMax)
	})
	t.Run("RoyalCotillion", func(t *testing.T) {
		g := new(interfaces.MockRoyalCotillionGame)
		setupRoyalCotillionCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
		g.On("GetPhase").Return(domain.RoyalCotillionPhaseGameOver)

		result := new(RoyalCotillionCuiPresenter).Output(g, nil)
		assertGameOverSummary(t, result, domain.RoyalCotillionTotalCards)
	})
	t.Run("StreetsAndAlleys", func(t *testing.T) {
		g := new(interfaces.MockStreetsAndAlleysGame)
		setupStreetsAndAlleysCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
		g.On("GetPhase").Return(domain.StreetsAndAlleysPhaseGameOver)

		result := new(StreetsAndAlleysCuiPresenter).Output(g, nil)
		assertGameOverSummary(t, result, domain.StreetsAndAlleysFoundationCnt*domain.CardValueMax)
	})
	t.Run("Terrace", func(t *testing.T) {
		g := new(interfaces.MockTerraceGame)
		setupTerraceCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
		g.On("GetPhase").Return(domain.TerracePhaseGameOver)

		result := new(TerraceCuiPresenter).Output(g, nil)
		assertGameOverSummary(t, result, domain.TerraceTotalCards)
	})
	t.Run("Windmill", func(t *testing.T) {
		g := new(interfaces.MockWindmillGame)
		setupWindmillCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
		g.On("GetPhase").Return(domain.WindmillPhaseGameOver)

		result := new(WindmillCuiPresenter).Output(g, nil)
		assertGameOverSummary(t, result, domain.WindmillTotalCards)
	})
}

// TestGrandfathersClockCuiGameOverFaces is separate because that game reports
// completed clock faces instead of a card count, so a percentage is meaningless.
func TestGrandfathersClockCuiGameOverFaces(t *testing.T) {
	g := new(interfaces.MockGrandfathersClockGame)
	setupGrandfathersClockCuiMockDefaults(g)
	g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
	g.On("GetPhase").Return(domain.GrandfathersClockPhaseGameOver)

	result := new(GrandfathersClockCuiPresenter).Output(g, nil)
	assert.Contains(t, result, "個を完成",
		"game over screen is missing the completed-faces line")
	assert.Contains(t, result, "/"+strconv.Itoa(domain.GrandfathersClockFoundationCnt)+" 個")
	assert.Contains(t, result, "ゲームオーバー")
}
