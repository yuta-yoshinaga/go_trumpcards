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

func setupCongressWebMockDefaults(g *interfaces.MockCongressGame) {
	g.On("GetPhase").Return(domain.CongressPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	g.On("GetStockCount").Return(96).Maybe()
	g.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 9, true)}).Maybe()

	var tableau [domain.CongressTableauCnt][]*domain.Card
	for i := range domain.CongressTableauCnt {
		tableau[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, i+2, true)}
	}
	g.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.CongressFoundationCnt][]*domain.Card
	g.On("GetFoundation").Return(foundation).Maybe()
}

func parseCongressOutput(t *testing.T, jsonStr string) *controller.CongressWebOutput {
	t.Helper()
	var out controller.CongressWebOutput
	assert.NoError(t, json.Unmarshal([]byte(jsonStr), &out))
	return &out
}

// setupCongressOutputMock は Output 用の既定を組む。
//
// **Output() も受動ヒントを埋めるようになった** (#4483) ので、GetHint を
// 呼べるようにしておく必要がある。共有ヘルパー側に置くと、先に登録された
// この期待が HintOutput テストの「ヒントあり」を食ってしまう。
func setupCongressOutputMock(g *interfaces.MockCongressGame) {
	setupCongressWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestCongressWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		g := new(interfaces.MockCongressGame)
		setupCongressOutputMock(g)

		result := parseCongressOutput(t, new(CongressWebPresenter).Output(g, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 96, result.StockCount)
		assert.Len(t, result.Tableau, domain.CongressTableauCnt)
		assert.Len(t, result.Foundation, domain.CongressFoundationCnt)
		assert.Len(t, result.Waste, 1)
		assert.Equal(t, "congress.playing", result.MessageCode)
	})

	t.Run("stalemate", func(t *testing.T) {
		g := new(interfaces.MockCongressGame)
		setupCongressOutputMock(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)

		result := parseCongressOutput(t, new(CongressWebPresenter).Output(g, nil))
		assert.Equal(t, "congress.stalemate", result.MessageCode)
	})

	t.Run("error message", func(t *testing.T) {
		g := new(interfaces.MockCongressGame)
		setupCongressOutputMock(g)

		result := parseCongressOutput(t, new(CongressWebPresenter).Output(g, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})

	for _, tc := range []struct {
		name string
		val  domain.CongressPhase
		code string
	}{
		{"game clear", domain.CongressPhaseGameClear, "congress.gameClear"},
		{"game over", domain.CongressPhaseGameOver, "congress.gameOver"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockCongressGame)
			setupCongressOutputMock(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.On("GetPhase").Return(tc.val)

			result := parseCongressOutput(t, new(CongressWebPresenter).Output(g, nil))
			assert.Equal(t, tc.code, result.MessageCode)
		})
	}
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない。ここが埋まっていないと
// フロントの `state.hint` は常に undefined で、それを読む分岐が全部死ぬ (#4483)。
func TestCongressWebPresenter_OutputCarriesTheHint(t *testing.T) {
	hint := &domain.CongressHint{FromZone: "tableau", FromIdx: 2, ToZone: "foundation", ToIdx: 1}

	t.Run("while the game is playable", func(t *testing.T) {
		g := new(interfaces.MockCongressGame)
		setupCongressWebMockDefaults(g)
		g.On("GetHint").Return(hint).Maybe()

		result := parseCongressOutput(t, new(CongressWebPresenter).Output(g, nil))
		if result.Hint == nil {
			t.Fatal("Output must carry the hint -- the frontend reads state.hint")
		}
		assert.Equal(t, "tableau", result.Hint.FromZone)
		assert.Equal(t, 2, result.Hint.FromIdx)
	})
}

func TestCongressWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		g := new(interfaces.MockCongressGame)
		setupCongressWebMockDefaults(g)
		g.On("GetHint").Return(&domain.CongressHint{
			FromZone: "stock", FromIdx: -1, ToZone: "tableau", ToIdx: 3,
		})

		result := parseCongressOutput(t, new(CongressWebPresenter).HintOutput(g))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "stock", result.Hint.FromZone)
		assert.Equal(t, "tableau", result.Hint.ToZone)
		assert.Equal(t, 3, result.Hint.ToIdx)
		assert.Equal(t, "congress.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockCongressGame)
		setupCongressWebMockDefaults(g)
		g.On("GetHint").Return((*domain.CongressHint)(nil))

		result := parseCongressOutput(t, new(CongressWebPresenter).HintOutput(g))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "congress.noHint", result.MessageCode)
	})
}

func TestCongressWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		g := new(interfaces.MockCongressGame)
		g.On("GetPhase").Return(domain.CongressPhasePlaying)
		g.On("GetGameEndFlag").Return(false)

		assert.Contains(t, new(CongressWebPresenter).ActionLogOutput(g), "[]")
	})

	t.Run("game over returns log", func(t *testing.T) {
		g := new(interfaces.MockCongressGame)
		g.On("GetPhase").Return(domain.CongressPhaseGameOver)
		g.On("GetGameEndFlag").Return(true)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		assert.Contains(t, new(CongressWebPresenter).ActionLogOutput(g), "move")
	})
}
