//go:build test

package presenter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func newNertzPlayerWithCards(name string, isCpu bool, deckIdx int) *domain.NertzPlayer {
	p := domain.NewNertzPlayer(name, isCpu, deckIdx)
	p.PushNertz(domain.NewCard(domain.CardDesignSpade, 1, false))
	p.PushTableau(0, &domain.NertzTableauCard{Card: domain.NewCard(domain.CardDesignHeart, 5, false), FaceUp: true})
	p.PushWaste(domain.NewCard(domain.CardDesignDiamond, 7, false))
	p.PushStock(domain.NewCard(domain.CardDesignClover, 9, false))
	return p
}

func newNertzFoundationWithAce() *domain.NertzFoundation {
	f := domain.NewNertzFoundation()
	_ = f.Push(domain.NewCard(domain.CardDesignSpade, 1, false), 0)
	return f
}

func setupNertzWebMockDefaults(g *interfaces.MockNertzGame) {
	cfg := domain.DefaultNertzConfig()
	g.On("GetPhase").Return(domain.NertzPhasePlaying).Maybe()
	g.On("GetRoundNo").Return(1).Maybe()
	g.On("GetWinnerIdx").Return(-1).Maybe()
	g.On("GetMatchWinner").Return(-1).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("GetConfig").Return(cfg).Maybe()
	human := newNertzPlayerWithCards("You", false, 0)
	cpu := newNertzPlayerWithCards("CPU1", true, 1)
	g.On("GetPlayers").Return([]*domain.NertzPlayer{human, cpu}).Maybe()
	g.On("GetFoundations").Return([]*domain.NertzFoundation{
		newNertzFoundationWithAce(),
		domain.NewNertzFoundation(),
	}).Maybe()
}

func decodeNertzWebOutput(t *testing.T, raw string) *controller.NertzWebOutput {
	t.Helper()
	var out controller.NertzWebOutput
	require.NoError(t, json.Unmarshal([]byte(raw), &out))
	return &out
}

// setupNertzOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**ように
// なった (#4483) ので GetHint を呼べるようにする。共有ヘルパーに置くと、先に
// 登録されたこの期待が HintOutput テストの「ヒントあり」を食う。
func setupNertzOutputMock(g *interfaces.MockNertzGame) {
	setupNertzWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestNertzWebPresenter_Output(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		g := new(interfaces.MockNertzGame)
		setupNertzOutputMock(g)
		raw := new(NertzWebPresenter).Output(g, nil)
		out := decodeNertzWebOutput(t, raw)
		assert.Equal(t, "nertz.playing", out.MessageCode)
		assert.Len(t, out.Players, 2)
		assert.Len(t, out.Players[0].Tableau, domain.NertzTableauCnt)
		require.NotNil(t, out.Players[0].NertzTop)
		require.NotNil(t, out.Players[0].WasteTop)
		assert.Equal(t, 1, out.Players[0].StockSize)
		assert.Len(t, out.Foundations, 2)
		require.NotNil(t, out.Foundations[0].Top)
		// Issue #1532: locked-in shape — RoundNumber populated, IsHuman is the
		// inverse of the domain's IsCpu (so true for the human player at idx 0
		// and false for CPU at idx 1, matching Slapjack / majority of games).
		assert.Equal(t, 1, out.RoundNumber, "RoundNumber must be populated from GetRoundNo")
		assert.True(t, out.Players[0].IsHuman, "human player (idx 0) must report IsHuman=true")
		assert.False(t, out.Players[1].IsHuman, "CPU player (idx 1) must report IsHuman=false")
	})

	// TestNertzWebPresenter_JSONShape pins down the JSON-tag rename to lock
	// in the public API contract. See issue #1532.
	t.Run("JSON shape uses isHuman / roundNumber", func(t *testing.T) {
		g := new(interfaces.MockNertzGame)
		setupNertzOutputMock(g)
		raw := new(NertzWebPresenter).Output(g, nil)
		// New canonical names must appear.
		assert.Contains(t, raw, `"roundNumber":`, "JSON output must use roundNumber")
		assert.Contains(t, raw, `"isHuman":`, "JSON output must use isHuman")
		// Old legacy names must NOT appear (this is a breaking rename, not a dual emit).
		assert.NotContains(t, raw, `"roundNo":`, "JSON output must not include legacy roundNo")
		assert.NotContains(t, raw, `"isCpu":`, "JSON output must not include legacy isCpu")
	})

	t.Run("with error", func(t *testing.T) {
		g := new(interfaces.MockNertzGame)
		setupNertzOutputMock(g)
		raw := new(NertzWebPresenter).Output(g, assert.AnError)
		out := decodeNertzWebOutput(t, raw)
		assert.Equal(t, assert.AnError.Error(), out.Message)
	})

	t.Run("round end", func(t *testing.T) {
		g := new(interfaces.MockNertzGame)
		g.On("GetPhase").Return(domain.NertzPhaseRoundEnd).Maybe()
		g.On("GetRoundNo").Return(1).Maybe()
		g.On("GetWinnerIdx").Return(0).Maybe()
		g.On("GetMatchWinner").Return(-1).Maybe()
		g.On("GetMoveCount").Return(13).Maybe()
		g.On("CanUndo").Return(false).Maybe()
		g.On("GetConfig").Return(domain.DefaultNertzConfig()).Maybe()
		g.On("GetPlayers").Return([]*domain.NertzPlayer{}).Maybe()
		g.On("GetFoundations").Return([]*domain.NertzFoundation{}).Maybe()
		raw := new(NertzWebPresenter).Output(g, nil)
		out := decodeNertzWebOutput(t, raw)
		assert.Equal(t, "nertz.roundEnd", out.MessageCode)
		assert.Equal(t, "0", out.MessageParams["winner"])
	})

	t.Run("human win", func(t *testing.T) {
		g := new(interfaces.MockNertzGame)
		g.On("GetPhase").Return(domain.NertzPhaseGameEnd).Maybe()
		g.On("GetRoundNo").Return(3).Maybe()
		g.On("GetWinnerIdx").Return(0).Maybe()
		g.On("GetMatchWinner").Return(0).Maybe()
		g.On("GetMoveCount").Return(50).Maybe()
		g.On("CanUndo").Return(false).Maybe()
		g.On("GetConfig").Return(domain.DefaultNertzConfig()).Maybe()
		g.On("GetPlayers").Return([]*domain.NertzPlayer{}).Maybe()
		g.On("GetFoundations").Return([]*domain.NertzFoundation{}).Maybe()
		raw := new(NertzWebPresenter).Output(g, nil)
		out := decodeNertzWebOutput(t, raw)
		assert.Equal(t, "nertz.win", out.MessageCode)
	})

	t.Run("cpu win", func(t *testing.T) {
		g := new(interfaces.MockNertzGame)
		g.On("GetPhase").Return(domain.NertzPhaseGameEnd).Maybe()
		g.On("GetRoundNo").Return(3).Maybe()
		g.On("GetWinnerIdx").Return(1).Maybe()
		g.On("GetMatchWinner").Return(1).Maybe()
		g.On("GetMoveCount").Return(50).Maybe()
		g.On("CanUndo").Return(false).Maybe()
		g.On("GetConfig").Return(domain.DefaultNertzConfig()).Maybe()
		g.On("GetPlayers").Return([]*domain.NertzPlayer{}).Maybe()
		g.On("GetFoundations").Return([]*domain.NertzFoundation{}).Maybe()
		raw := new(NertzWebPresenter).Output(g, nil)
		out := decodeNertzWebOutput(t, raw)
		assert.Equal(t, "nertz.lose", out.MessageCode)
	})

	t.Run("nil player and nil foundation", func(t *testing.T) {
		g := new(interfaces.MockNertzGame)
		g.On("GetPhase").Return(domain.NertzPhasePlaying).Maybe()
		g.On("GetRoundNo").Return(1).Maybe()
		g.On("GetWinnerIdx").Return(-1).Maybe()
		g.On("GetMatchWinner").Return(-1).Maybe()
		g.On("GetMoveCount").Return(0).Maybe()
		g.On("CanUndo").Return(false).Maybe()
		g.On("GetConfig").Return(domain.DefaultNertzConfig()).Maybe()
		g.On("GetPlayers").Return([]*domain.NertzPlayer{nil}).Maybe()
		g.On("GetFoundations").Return([]*domain.NertzFoundation{nil}).Maybe()
		g.On("GetHint").Return(nil).Maybe()
		raw := new(NertzWebPresenter).Output(g, nil)
		out := decodeNertzWebOutput(t, raw)
		assert.Empty(t, out.Players)
		require.Len(t, out.Foundations, 1)
		assert.Equal(t, -1, out.Foundations[0].Suit)
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestNertzWebPresenterOutputCarriesTheHint(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		ntg := new(interfaces.MockNertzGame)
		setupNertzWebMockDefaults(ntg)
		ntg.On("GetHint").Return(&domain.NertzHint{FromZone: "waste", FromCol: -1, CardIndex: -1, ToZone: "foundation", ToCol: 2}).Maybe()

		result := new(NertzWebPresenter).Output(ntg, nil)
		assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	})

	t.Run("not once the game has ended", func(t *testing.T) {
		ntg := new(interfaces.MockNertzGame)
		setupNertzWebMockDefaults(ntg)
		ntg.ExpectedCalls = filterCalls(ntg.ExpectedCalls, "GetPhase")
		ntg.On("GetPhase").Return(domain.NertzPhaseGameEnd)
		ntg.On("GetHint").Return(&domain.NertzHint{FromZone: "waste", FromCol: -1, CardIndex: -1, ToZone: "foundation", ToCol: 2}).Maybe()

		result := new(NertzWebPresenter).Output(ntg, nil)
		assert.NotContains(t, result, `"hint"`)
	})
}

func TestNertzWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		g := new(interfaces.MockNertzGame)
		setupNertzWebMockDefaults(g)
		g.On("GetHint").Return(&domain.NertzHint{
			FromZone: "nertz", FromCol: -1, CardIndex: -1,
			ToZone: "foundation", ToCol: 2,
		})
		raw := new(NertzWebPresenter).HintOutput(g)
		out := decodeNertzWebOutput(t, raw)
		assert.Equal(t, "nertz.hintAvailable", out.MessageCode)
		require.NotNil(t, out.Hint)
		assert.Equal(t, "nertz", out.Hint.FromZone)
		assert.Equal(t, 2, out.Hint.ToCol)
	})

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockNertzGame)
		setupNertzWebMockDefaults(g)
		g.On("GetHint").Return((*domain.NertzHint)(nil))
		raw := new(NertzWebPresenter).HintOutput(g)
		out := decodeNertzWebOutput(t, raw)
		assert.Equal(t, "nertz.noHint", out.MessageCode)
		assert.Nil(t, out.Hint)
	})
}

func TestNertzWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing returns empty log", func(t *testing.T) {
		g := new(interfaces.MockNertzGame)
		g.On("GetPhase").Return(domain.NertzPhasePlaying)
		g.On("GetGameEndFlag").Return(false)
		assert.NotEmpty(t, new(NertzWebPresenter).ActionLogOutput(g))
	})
	t.Run("after round end returns full log", func(t *testing.T) {
		g := new(interfaces.MockNertzGame)
		g.On("GetPhase").Return(domain.NertzPhaseRoundEnd)
		g.On("GetGameEndFlag").Return(true)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{{TurnNumber: 1, ActionType: "moveNF"}})
		assert.NotEmpty(t, new(NertzWebPresenter).ActionLogOutput(g))
	})
}
