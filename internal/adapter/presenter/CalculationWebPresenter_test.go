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

func setupCalculationWebMockDefaults(g *interfaces.MockCalculationGame) {
	g.On("GetPhase").Return(domain.CalculationPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	g.On("GetStockCount").Return(48).Maybe()
	g.On("GetStockTop").Return(domain.NewCard(domain.CardDesignSpade, 5, false)).Maybe()

	var foundations [domain.CalculationFoundationCnt][]*domain.Card
	for i := range domain.CalculationFoundationCnt {
		foundations[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, i+1, false)}
	}
	g.On("GetFoundations").Return(foundations).Maybe()

	var wastes [domain.CalculationWasteCnt][]*domain.Card
	g.On("GetWastes").Return(wastes).Maybe()
}

func parseCalculationOutput(t *testing.T, jsonStr string) *controller.CalculationWebOutput {
	t.Helper()
	var out controller.CalculationWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

// setupCalculationOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**
// ようになった (#4483) ので GetHint を呼べるようにする。
func setupCalculationOutputMock(g *interfaces.MockCalculationGame) {
	setupCalculationWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestCalculationWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		g := new(interfaces.MockCalculationGame)
		setupCalculationOutputMock(g)
		result := parseCalculationOutput(t, new(CalculationWebPresenter).Output(g, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, "calculation.playing", result.MessageCode)
		assert.NotNil(t, result.StockTop)
		assert.Equal(t, 48, result.StockCount)
	})

	t.Run("stalemate", func(t *testing.T) {
		g := new(interfaces.MockCalculationGame)
		g.On("GetPhase").Return(domain.CalculationPhasePlaying).Maybe()
		g.On("GetMoveCount").Return(5).Maybe()
		g.On("CanUndo").Return(true).Maybe()
		g.On("IsStalemate").Return(true).Maybe()
		g.On("UndoToEscape").Return(1).Maybe()
		g.On("GetStockCount").Return(0).Maybe()
		g.On("GetStockTop").Return((*domain.Card)(nil)).Maybe()
		var foundations [domain.CalculationFoundationCnt][]*domain.Card
		g.On("GetFoundations").Return(foundations).Maybe()
		var wastes [domain.CalculationWasteCnt][]*domain.Card
		g.On("GetWastes").Return(wastes).Maybe()

		result := parseCalculationOutput(t, new(CalculationWebPresenter).Output(g, nil))
		assert.Equal(t, "calculation.stalemate", result.MessageCode)
		assert.True(t, result.IsStalemate)
	})

	t.Run("game clear", func(t *testing.T) {
		g := new(interfaces.MockCalculationGame)
		g.On("GetPhase").Return(domain.CalculationPhaseGameClear).Maybe()
		g.On("GetMoveCount").Return(100).Maybe()
		g.On("CanUndo").Return(false).Maybe()
		g.On("IsStalemate").Return(false).Maybe()
		g.On("UndoToEscape").Return(0).Maybe()
		g.On("GetStockCount").Return(0).Maybe()
		g.On("GetStockTop").Return((*domain.Card)(nil)).Maybe()
		var foundations [domain.CalculationFoundationCnt][]*domain.Card
		g.On("GetFoundations").Return(foundations).Maybe()
		var wastes [domain.CalculationWasteCnt][]*domain.Card
		g.On("GetWastes").Return(wastes).Maybe()

		result := parseCalculationOutput(t, new(CalculationWebPresenter).Output(g, nil))
		assert.Equal(t, "calculation.gameClear", result.MessageCode)
	})

	t.Run("game over", func(t *testing.T) {
		g := new(interfaces.MockCalculationGame)
		g.On("GetPhase").Return(domain.CalculationPhaseGameOver).Maybe()
		g.On("GetMoveCount").Return(10).Maybe()
		g.On("CanUndo").Return(false).Maybe()
		g.On("IsStalemate").Return(false).Maybe()
		g.On("UndoToEscape").Return(0).Maybe()
		g.On("GetStockCount").Return(0).Maybe()
		g.On("GetStockTop").Return((*domain.Card)(nil)).Maybe()
		var foundations [domain.CalculationFoundationCnt][]*domain.Card
		g.On("GetFoundations").Return(foundations).Maybe()
		var wastes [domain.CalculationWasteCnt][]*domain.Card
		g.On("GetWastes").Return(wastes).Maybe()

		result := parseCalculationOutput(t, new(CalculationWebPresenter).Output(g, nil))
		assert.Equal(t, "calculation.gameOver", result.MessageCode)
	})

	t.Run("with error", func(t *testing.T) {
		g := new(interfaces.MockCalculationGame)
		setupCalculationOutputMock(g)
		result := parseCalculationOutput(t, new(CalculationWebPresenter).Output(g, errors.New("boom")))
		assert.Equal(t, "boom", result.Message)
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestCalculationWebPresenter_OutputCarriesTheHint(t *testing.T) {
	hint := &domain.CalculationHint{FromZone: "waste", WasteIdx: 2, FoundationIdx: 1}

	g := new(interfaces.MockCalculationGame)
	setupCalculationWebMockDefaults(g)
	g.On("GetHint").Return(hint).Maybe()

	result := parseCalculationOutput(t, new(CalculationWebPresenter).Output(g, nil))
	if result.Hint == nil {
		t.Fatal("Output must carry the hint -- the frontend reads state.hint")
	}
	assert.Equal(t, 2, result.Hint.WasteIdx)
}

func TestCalculationWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		g := new(interfaces.MockCalculationGame)
		setupCalculationWebMockDefaults(g)
		g.On("GetHint").Return(&domain.CalculationHint{FromZone: "waste", WasteIdx: 2, FoundationIdx: 1})

		result := parseCalculationOutput(t, new(CalculationWebPresenter).HintOutput(g))
		require := assert.NotNil(t, result.Hint)
		if !require {
			t.FailNow()
		}
		assert.Equal(t, "waste", result.Hint.FromZone)
		assert.Equal(t, 2, result.Hint.WasteIdx)
		assert.Equal(t, 1, result.Hint.FoundationIdx)
		assert.Equal(t, "calculation.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockCalculationGame)
		setupCalculationWebMockDefaults(g)
		g.On("GetHint").Return((*domain.CalculationHint)(nil))
		result := parseCalculationOutput(t, new(CalculationWebPresenter).HintOutput(g))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "calculation.noHint", result.MessageCode)
	})
}

func TestCalculationWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		g := new(interfaces.MockCalculationGame)
		g.On("GetPhase").Return(domain.CalculationPhasePlaying)
		g.On("GetGameEndFlag").Return(false)
		result := new(CalculationWebPresenter).ActionLogOutput(g)
		assert.Contains(t, result, "entries")
	})

	t.Run("game over", func(t *testing.T) {
		g := new(interfaces.MockCalculationGame)
		g.On("GetPhase").Return(domain.CalculationPhaseGameOver)
		g.On("GetGameEndFlag").Return(true)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{{TurnNumber: 1, ActionType: "move"}})
		result := new(CalculationWebPresenter).ActionLogOutput(g)
		assert.Contains(t, result, "entries")
	})
}
