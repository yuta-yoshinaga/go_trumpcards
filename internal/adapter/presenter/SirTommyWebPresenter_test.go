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

func setupSirTommyWebMockDefaults(g *interfaces.MockSirTommyGame) {
	g.On("GetPhase").Return(domain.SirTommyPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	g.On("GetStockCount").Return(48).Maybe()
	g.On("GetStockTop").Return(domain.NewCard(domain.CardDesignSpade, 5, false)).Maybe()

	var foundations [domain.SirTommyFoundationCnt][]*domain.Card
	for i := range domain.SirTommyFoundationCnt {
		foundations[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, i+1, false)}
	}
	g.On("GetFoundations").Return(foundations).Maybe()

	var wastes [domain.SirTommyWasteCnt][]*domain.Card
	g.On("GetWastes").Return(wastes).Maybe()
}

func parseSirTommyOutput(t *testing.T, jsonStr string) *controller.SirTommyWebOutput {
	t.Helper()
	var out controller.SirTommyWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

func TestSirTommyWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		g := new(interfaces.MockSirTommyGame)
		setupSirTommyWebMockDefaults(g)
		result := parseSirTommyOutput(t, new(SirTommyWebPresenter).Output(g, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, "sirtommy.playing", result.MessageCode)
		assert.NotNil(t, result.StockTop)
		assert.Equal(t, 48, result.StockCount)
	})

	t.Run("stalemate", func(t *testing.T) {
		g := new(interfaces.MockSirTommyGame)
		g.On("GetPhase").Return(domain.SirTommyPhasePlaying).Maybe()
		g.On("GetMoveCount").Return(5).Maybe()
		g.On("CanUndo").Return(true).Maybe()
		g.On("IsStalemate").Return(true).Maybe()
		g.On("UndoToEscape").Return(1).Maybe()
		g.On("GetStockCount").Return(0).Maybe()
		g.On("GetStockTop").Return((*domain.Card)(nil)).Maybe()
		var foundations [domain.SirTommyFoundationCnt][]*domain.Card
		g.On("GetFoundations").Return(foundations).Maybe()
		var wastes [domain.SirTommyWasteCnt][]*domain.Card
		g.On("GetWastes").Return(wastes).Maybe()

		result := parseSirTommyOutput(t, new(SirTommyWebPresenter).Output(g, nil))
		assert.Equal(t, "sirtommy.stalemate", result.MessageCode)
		assert.True(t, result.IsStalemate)
	})

	t.Run("game clear", func(t *testing.T) {
		g := new(interfaces.MockSirTommyGame)
		g.On("GetPhase").Return(domain.SirTommyPhaseGameClear).Maybe()
		g.On("GetMoveCount").Return(100).Maybe()
		g.On("CanUndo").Return(false).Maybe()
		g.On("IsStalemate").Return(false).Maybe()
		g.On("UndoToEscape").Return(0).Maybe()
		g.On("GetStockCount").Return(0).Maybe()
		g.On("GetStockTop").Return((*domain.Card)(nil)).Maybe()
		var foundations [domain.SirTommyFoundationCnt][]*domain.Card
		g.On("GetFoundations").Return(foundations).Maybe()
		var wastes [domain.SirTommyWasteCnt][]*domain.Card
		g.On("GetWastes").Return(wastes).Maybe()

		result := parseSirTommyOutput(t, new(SirTommyWebPresenter).Output(g, nil))
		assert.Equal(t, "sirtommy.gameClear", result.MessageCode)
	})

	t.Run("game over", func(t *testing.T) {
		g := new(interfaces.MockSirTommyGame)
		g.On("GetPhase").Return(domain.SirTommyPhaseGameOver).Maybe()
		g.On("GetMoveCount").Return(10).Maybe()
		g.On("CanUndo").Return(false).Maybe()
		g.On("IsStalemate").Return(false).Maybe()
		g.On("UndoToEscape").Return(0).Maybe()
		g.On("GetStockCount").Return(0).Maybe()
		g.On("GetStockTop").Return((*domain.Card)(nil)).Maybe()
		var foundations [domain.SirTommyFoundationCnt][]*domain.Card
		g.On("GetFoundations").Return(foundations).Maybe()
		var wastes [domain.SirTommyWasteCnt][]*domain.Card
		g.On("GetWastes").Return(wastes).Maybe()

		result := parseSirTommyOutput(t, new(SirTommyWebPresenter).Output(g, nil))
		assert.Equal(t, "sirtommy.gameOver", result.MessageCode)
	})

	t.Run("with error", func(t *testing.T) {
		g := new(interfaces.MockSirTommyGame)
		setupSirTommyWebMockDefaults(g)
		result := parseSirTommyOutput(t, new(SirTommyWebPresenter).Output(g, errors.New("boom")))
		assert.Equal(t, "boom", result.Message)
	})
}

func TestSirTommyWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		g := new(interfaces.MockSirTommyGame)
		setupSirTommyWebMockDefaults(g)
		g.On("GetHint").Return(&domain.SirTommyHint{FromZone: "waste", WasteIdx: 2, FoundationIdx: 1})

		result := parseSirTommyOutput(t, new(SirTommyWebPresenter).HintOutput(g))
		require := assert.NotNil(t, result.Hint)
		if !require {
			t.FailNow()
		}
		assert.Equal(t, "waste", result.Hint.FromZone)
		assert.Equal(t, 2, result.Hint.WasteIdx)
		assert.Equal(t, 1, result.Hint.FoundationIdx)
		assert.Equal(t, "sirtommy.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockSirTommyGame)
		setupSirTommyWebMockDefaults(g)
		g.On("GetHint").Return((*domain.SirTommyHint)(nil))
		result := parseSirTommyOutput(t, new(SirTommyWebPresenter).HintOutput(g))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "sirtommy.noHint", result.MessageCode)
	})
}

func TestSirTommyWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		g := new(interfaces.MockSirTommyGame)
		g.On("GetPhase").Return(domain.SirTommyPhasePlaying)
		g.On("GetGameEndFlag").Return(false)
		result := new(SirTommyWebPresenter).ActionLogOutput(g)
		assert.Contains(t, result, "entries")
	})

	t.Run("game over", func(t *testing.T) {
		g := new(interfaces.MockSirTommyGame)
		g.On("GetPhase").Return(domain.SirTommyPhaseGameOver)
		g.On("GetGameEndFlag").Return(true)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{{TurnNumber: 1, ActionType: "move"}})
		result := new(SirTommyWebPresenter).ActionLogOutput(g)
		assert.Contains(t, result, "entries")
	})
}
