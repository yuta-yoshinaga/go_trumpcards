//go:build test

package presenter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupTerraceWebMockDefaults(g *interfaces.MockTerraceGame) {
	g.On("GetPhase").Return(domain.TerracePhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	g.On("GetStockCount").Return(84).Maybe()
	g.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 9, true)}).Maybe()
	g.On("GetReserve").Return([]*domain.Card{domain.NewCard(domain.CardDesignClover, 3, true)}).Maybe()
	g.On("GetBaseRank").Return(5).Maybe()
	g.On("IsAwaitingBaseRank").Return(false).Maybe()

	var tableau [domain.TerraceTableauCnt][]*domain.Card
	for i := range domain.TerraceTableauCnt {
		tableau[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, i+2, true)}
	}
	g.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.TerraceFoundationCnt][]*domain.Card
	g.On("GetFoundation").Return(foundation).Maybe()
}

func parseTerraceOutput(t *testing.T, jsonStr string) *controller.TerraceWebOutput {
	t.Helper()
	var out controller.TerraceWebOutput
	assert.NoError(t, json.Unmarshal([]byte(jsonStr), &out))
	return &out
}

// setupTerraceOutputMock は Output 用の既定を組む。
//
// **Output() も受動ヒントを埋めるようになった** (#4483) ので、GetHint を
// 呼べるようにしておく必要がある。共有ヘルパー側に置くと、先に登録された
// この期待が HintOutput テストの「ヒントあり」を食ってしまう。
func setupTerraceOutputMock(g *interfaces.MockTerraceGame) {
	setupTerraceWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestTerraceWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		g := new(interfaces.MockTerraceGame)
		setupTerraceOutputMock(g)

		result := parseTerraceOutput(t, new(TerraceWebPresenter).Output(g, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 84, result.StockCount)
		assert.Len(t, result.Reserve, 1)
		assert.Len(t, result.Tableau, domain.TerraceTableauCnt)
		assert.Len(t, result.Foundation, domain.TerraceFoundationCnt)
		assert.Equal(t, 5, result.BaseRank)
		assert.False(t, result.AwaitingBaseRank)
		assert.Equal(t, "terrace.playing", result.MessageCode)
	})

	// The client must not have to infer "rank 0 means unchosen".
	t.Run("awaiting the base rank has its own message", func(t *testing.T) {
		g := new(interfaces.MockTerraceGame)
		setupTerraceOutputMock(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsAwaitingBaseRank")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetBaseRank")
		g.On("IsAwaitingBaseRank").Return(true)
		g.On("GetBaseRank").Return(0)

		result := parseTerraceOutput(t, new(TerraceWebPresenter).Output(g, nil))
		assert.True(t, result.AwaitingBaseRank)
		assert.Equal(t, 0, result.BaseRank)
		assert.Equal(t, "terrace.chooseBase", result.MessageCode)
	})

	t.Run("awaiting the base rank outranks stalemate", func(t *testing.T) {
		g := new(interfaces.MockTerraceGame)
		setupTerraceOutputMock(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsAwaitingBaseRank")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsAwaitingBaseRank").Return(true)
		g.On("IsStalemate").Return(true)

		result := parseTerraceOutput(t, new(TerraceWebPresenter).Output(g, nil))
		assert.Equal(t, "terrace.chooseBase", result.MessageCode)
	})

	t.Run("stalemate", func(t *testing.T) {
		g := new(interfaces.MockTerraceGame)
		setupTerraceOutputMock(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)

		result := parseTerraceOutput(t, new(TerraceWebPresenter).Output(g, nil))
		assert.Equal(t, "terrace.stalemate", result.MessageCode)
	})

	t.Run("error message", func(t *testing.T) {
		g := new(interfaces.MockTerraceGame)
		setupTerraceOutputMock(g)

		result := parseTerraceOutput(t, new(TerraceWebPresenter).Output(g, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})

	for _, tc := range []struct {
		name string
		val  domain.TerracePhase
		code string
	}{
		{"game clear", domain.TerracePhaseGameClear, "terrace.gameClear"},
		{"game over", domain.TerracePhaseGameOver, "terrace.gameOver"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockTerraceGame)
			setupTerraceOutputMock(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.On("GetPhase").Return(tc.val)

			result := parseTerraceOutput(t, new(TerraceWebPresenter).Output(g, nil))
			assert.Equal(t, tc.code, result.MessageCode)
		})
	}
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない。ここが埋まっていないと
// フロントの `state.hint` は常に undefined で、それを読む分岐が全部死ぬ (#4483)。
func TestTerraceWebPresenter_OutputCarriesTheHint(t *testing.T) {
	hint := &domain.TerraceHint{FromZone: "tableau", FromIdx: 2, ToZone: "foundation", ToIdx: 1}

	t.Run("while the game is playable", func(t *testing.T) {
		g := new(interfaces.MockTerraceGame)
		setupTerraceWebMockDefaults(g)
		g.On("GetHint").Return(hint).Maybe()

		result := parseTerraceOutput(t, new(TerraceWebPresenter).Output(g, nil))
		if result.Hint == nil {
			t.Fatal("Output must carry the hint -- the frontend reads state.hint")
		}
		assert.Equal(t, "tableau", result.Hint.FromZone)
		assert.Equal(t, 2, result.Hint.FromIdx)
	})

	// **開始ランク待ちでは、フロントが別経路で文言を出すので重ねない。**
	t.Run("not while awaiting the base rank", func(t *testing.T) {
		g := new(interfaces.MockTerraceGame)
		setupTerraceWebMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsAwaitingBaseRank")
		g.On("IsAwaitingBaseRank").Return(true).Maybe()
		g.On("GetHint").Return(hint).Maybe()

		result := parseTerraceOutput(t, new(TerraceWebPresenter).Output(g, nil))
		assert.Nil(t, result.Hint)
		g.AssertNotCalled(t, "GetHint")
	})
	// **手詰まりでは、フロントが別経路で文言を出すので重ねない。**
	t.Run("not while stalemate", func(t *testing.T) {
		g := new(interfaces.MockTerraceGame)
		setupTerraceWebMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true).Maybe()
		g.On("GetHint").Return(hint).Maybe()

		result := parseTerraceOutput(t, new(TerraceWebPresenter).Output(g, nil))
		assert.Nil(t, result.Hint)
		g.AssertNotCalled(t, "GetHint")
	})

}

func TestTerraceWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		g := new(interfaces.MockTerraceGame)
		setupTerraceWebMockDefaults(g)
		g.On("GetHint").Return(&domain.TerraceHint{
			FromZone: "reserve", FromIdx: -1, ToZone: "foundation", ToIdx: 2,
		})

		result := parseTerraceOutput(t, new(TerraceWebPresenter).HintOutput(g))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "reserve", result.Hint.FromZone)
		assert.Equal(t, 2, result.Hint.ToIdx)
		assert.Equal(t, "terrace.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockTerraceGame)
		setupTerraceWebMockDefaults(g)
		g.On("GetHint").Return((*domain.TerraceHint)(nil))

		result := parseTerraceOutput(t, new(TerraceWebPresenter).HintOutput(g))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "terrace.noHint", result.MessageCode)
	})
}

func TestTerraceWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		g := new(interfaces.MockTerraceGame)
		g.On("GetPhase").Return(domain.TerracePhasePlaying)
		g.On("GetGameEndFlag").Return(false)

		assert.Contains(t, new(TerraceWebPresenter).ActionLogOutput(g), "[]")
	})

	t.Run("game over returns log", func(t *testing.T) {
		g := new(interfaces.MockTerraceGame)
		g.On("GetPhase").Return(domain.TerracePhaseGameOver)
		g.On("GetGameEndFlag").Return(true)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		assert.Contains(t, new(TerraceWebPresenter).ActionLogOutput(g), "move")
	})
}
