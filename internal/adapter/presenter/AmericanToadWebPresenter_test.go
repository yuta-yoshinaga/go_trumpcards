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

func setupAmericanToadWebMockDefaults(g *interfaces.MockAmericanToadGame) {
	g.On("GetPhase").Return(domain.AmericanToadPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	g.On("GetStockCount").Return(75).Maybe()
	g.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 9, true)}).Maybe()
	g.On("GetReserve").Return([]*domain.Card{domain.NewCard(domain.CardDesignClover, 3, true)}).Maybe()
	g.On("GetBaseRank").Return(5).Maybe()
	g.On("GetPassesUsed").Return(0).Maybe()
	g.On("CanRedeal").Return(false).Maybe()

	var tableau [domain.AmericanToadTableauCnt][]*domain.AmericanToadTableauCard
	for i := range domain.AmericanToadTableauCnt {
		tableau[i] = []*domain.AmericanToadTableauCard{
			{Card: domain.NewCard(domain.CardDesignSpade, i+2, false), FaceUp: true},
		}
	}
	g.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.AmericanToadFoundationCnt][]*domain.Card
	g.On("GetFoundation").Return(foundation).Maybe()
}

func parseAmericanToadOutput(t *testing.T, jsonStr string) *controller.AmericanToadWebOutput {
	t.Helper()
	var out controller.AmericanToadWebOutput
	assert.NoError(t, json.Unmarshal([]byte(jsonStr), &out))
	return &out
}

func TestAmericanToadWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		g := new(interfaces.MockAmericanToadGame)
		setupAmericanToadWebMockDefaults(g)

		result := parseAmericanToadOutput(t, new(AmericanToadWebPresenter).Output(g, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 75, result.StockCount)
		assert.Len(t, result.Reserve, 1)
		assert.Len(t, result.Tableau, domain.AmericanToadTableauCnt)
		assert.Len(t, result.Foundation, domain.AmericanToadFoundationCnt)
		assert.Len(t, result.Waste, 1)
		assert.Equal(t, 5, result.BaseRank)
		assert.Equal(t, 0, result.PassesUsed)
		assert.False(t, result.CanRedeal)
		assert.Equal(t, "americantoad.playing", result.MessageCode)
	})

	// There is only one redeal, so the client must be told when it is on offer
	// rather than inferring it from an empty stock and a non-empty waste.
	t.Run("an available redeal is surfaced", func(t *testing.T) {
		g := new(interfaces.MockAmericanToadGame)
		setupAmericanToadWebMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "CanRedeal")
		g.On("CanRedeal").Return(true)

		result := parseAmericanToadOutput(t, new(AmericanToadWebPresenter).Output(g, nil))
		assert.True(t, result.CanRedeal)
		assert.Equal(t, "americantoad.redealAvailable", result.MessageCode)
	})

	t.Run("stalemate outranks the redeal notice", func(t *testing.T) {
		g := new(interfaces.MockAmericanToadGame)
		setupAmericanToadWebMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "CanRedeal")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("CanRedeal").Return(true)
		g.On("IsStalemate").Return(true)

		result := parseAmericanToadOutput(t, new(AmericanToadWebPresenter).Output(g, nil))
		assert.Equal(t, "americantoad.stalemate", result.MessageCode)
	})

	t.Run("error message", func(t *testing.T) {
		g := new(interfaces.MockAmericanToadGame)
		setupAmericanToadWebMockDefaults(g)

		result := parseAmericanToadOutput(t, new(AmericanToadWebPresenter).Output(g, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})

	for _, tc := range []struct {
		name string
		val  domain.AmericanToadPhase
		code string
	}{
		{"game clear", domain.AmericanToadPhaseGameClear, "americantoad.gameClear"},
		{"game over", domain.AmericanToadPhaseGameOver, "americantoad.gameOver"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockAmericanToadGame)
			setupAmericanToadWebMockDefaults(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.On("GetPhase").Return(tc.val)

			result := parseAmericanToadOutput(t, new(AmericanToadWebPresenter).Output(g, nil))
			assert.Equal(t, tc.code, result.MessageCode)
		})
	}
}

func TestAmericanToadWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		g := new(interfaces.MockAmericanToadGame)
		setupAmericanToadWebMockDefaults(g)
		g.On("GetHint").Return(&domain.AmericanToadHint{
			FromZone: "tableau", FromIdx: 2, CardIndex: 1, ToZone: "foundation", ToIdx: 3,
		})

		result := parseAmericanToadOutput(t, new(AmericanToadWebPresenter).HintOutput(g))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "tableau", result.Hint.FromZone)
		assert.Equal(t, 2, result.Hint.FromIdx)
		assert.Equal(t, 1, result.Hint.CardIndex)
		assert.Equal(t, 3, result.Hint.ToIdx)
		assert.Equal(t, "americantoad.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockAmericanToadGame)
		setupAmericanToadWebMockDefaults(g)
		g.On("GetHint").Return((*domain.AmericanToadHint)(nil))

		result := parseAmericanToadOutput(t, new(AmericanToadWebPresenter).HintOutput(g))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "americantoad.noHint", result.MessageCode)
	})
}

func TestAmericanToadWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		g := new(interfaces.MockAmericanToadGame)
		g.On("GetPhase").Return(domain.AmericanToadPhasePlaying)
		g.On("GetGameEndFlag").Return(false)

		assert.Contains(t, new(AmericanToadWebPresenter).ActionLogOutput(g), "[]")
	})

	t.Run("game over returns log", func(t *testing.T) {
		g := new(interfaces.MockAmericanToadGame)
		g.On("GetPhase").Return(domain.AmericanToadPhaseGameOver)
		g.On("GetGameEndFlag").Return(true)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		assert.Contains(t, new(AmericanToadWebPresenter).ActionLogOutput(g), "move")
	})
}
