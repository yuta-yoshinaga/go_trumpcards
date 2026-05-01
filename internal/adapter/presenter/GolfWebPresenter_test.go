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

func setupGolfWebMockDefaults(gg *interfaces.MockGolfGame) {
	gg.On("GetPhase").Return(domain.GolfPhasePlaying).Maybe()
	gg.On("GetMoveCount").Return(0).Maybe()
	gg.On("GetStockCount").Return(16).Maybe()
	gg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	gg.On("CanUndo").Return(false).Maybe()
	gg.On("IsStalemate").Return(false).Maybe()
	gg.On("UndoToEscape").Return(0).Maybe()
	gg.On("AllRemoved").Return(false).Maybe()

	var layout [domain.GolfColCnt][domain.GolfRowCnt]*domain.GolfCard
	for c := range domain.GolfColCnt {
		for r := range domain.GolfRowCnt {
			layout[c][r] = &domain.GolfCard{
				Card:    domain.NewCard(domain.CardDesignSpade, (c*5+r)%13+1, false),
				Removed: false,
			}
		}
	}
	gg.On("GetLayout").Return(layout).Maybe()
	for c := range domain.GolfColCnt {
		for r := range domain.GolfRowCnt {
			if r == domain.GolfRowCnt-1 {
				gg.On("IsExposed", c, r).Return(true).Maybe()
			} else {
				gg.On("IsExposed", c, r).Return(false).Maybe()
			}
		}
	}
}

func parseGolfOutput(t *testing.T, jsonStr string) *controller.GolfWebOutput {
	t.Helper()
	var out controller.GolfWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

func TestGolfWebPresenterOutput_Playing(t *testing.T) {
	gg := new(interfaces.MockGolfGame)
	setupGolfWebMockDefaults(gg)
	p := &GolfWebPresenter{}

	result := p.Output(gg, nil)
	out := parseGolfOutput(t, result)

	assert.Equal(t, 0, out.Phase)
	assert.Equal(t, 16, out.StockCount)
	assert.Equal(t, "golf.playing", out.MessageCode)
	assert.Len(t, out.Layout, domain.GolfColCnt)
}

func TestGolfWebPresenterOutput_Error(t *testing.T) {
	gg := new(interfaces.MockGolfGame)
	setupGolfWebMockDefaults(gg)
	p := &GolfWebPresenter{}

	result := p.Output(gg, errors.New("test error"))
	out := parseGolfOutput(t, result)

	assert.Equal(t, "test error", out.Message)
}

func TestGolfWebPresenterOutput_Stalemate(t *testing.T) {
	gg := new(interfaces.MockGolfGame)
	setupGolfWebMockDefaults(gg)
	gg.ExpectedCalls = nil
	gg.On("GetPhase").Return(domain.GolfPhasePlaying).Maybe()
	gg.On("GetMoveCount").Return(5).Maybe()
	gg.On("GetStockCount").Return(0).Maybe()
	gg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	gg.On("CanUndo").Return(false).Maybe()
	gg.On("IsStalemate").Return(true).Maybe()
	gg.On("UndoToEscape").Return(-1).Maybe()
	var layout [domain.GolfColCnt][domain.GolfRowCnt]*domain.GolfCard
	gg.On("GetLayout").Return(layout).Maybe()
	for c := range domain.GolfColCnt {
		for r := range domain.GolfRowCnt {
			gg.On("IsExposed", c, r).Return(false).Maybe()
		}
	}

	p := &GolfWebPresenter{}
	result := p.Output(gg, nil)
	out := parseGolfOutput(t, result)

	assert.Equal(t, "golf.stalemate", out.MessageCode)
}

func TestGolfWebPresenterOutput_GameClear(t *testing.T) {
	gg := new(interfaces.MockGolfGame)
	setupGolfWebMockDefaults(gg)
	gg.ExpectedCalls = nil
	gg.On("GetPhase").Return(domain.GolfPhaseGameClear).Maybe()
	gg.On("GetMoveCount").Return(10).Maybe()
	gg.On("GetStockCount").Return(0).Maybe()
	gg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	gg.On("CanUndo").Return(false).Maybe()
	gg.On("IsStalemate").Return(false).Maybe()
	gg.On("UndoToEscape").Return(0).Maybe()
	var layout [domain.GolfColCnt][domain.GolfRowCnt]*domain.GolfCard
	gg.On("GetLayout").Return(layout).Maybe()
	for c := range domain.GolfColCnt {
		for r := range domain.GolfRowCnt {
			gg.On("IsExposed", c, r).Return(false).Maybe()
		}
	}

	p := &GolfWebPresenter{}
	result := p.Output(gg, nil)
	out := parseGolfOutput(t, result)

	assert.Equal(t, "golf.gameClear", out.MessageCode)
	assert.Contains(t, out.Message, "10")
}

func TestGolfWebPresenterOutput_GameOver(t *testing.T) {
	gg := new(interfaces.MockGolfGame)
	setupGolfWebMockDefaults(gg)
	gg.ExpectedCalls = nil
	gg.On("GetPhase").Return(domain.GolfPhaseGameOver).Maybe()
	gg.On("GetMoveCount").Return(5).Maybe()
	gg.On("GetStockCount").Return(0).Maybe()
	gg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	gg.On("CanUndo").Return(false).Maybe()
	gg.On("IsStalemate").Return(false).Maybe()
	gg.On("UndoToEscape").Return(0).Maybe()
	var layout [domain.GolfColCnt][domain.GolfRowCnt]*domain.GolfCard
	gg.On("GetLayout").Return(layout).Maybe()
	for c := range domain.GolfColCnt {
		for r := range domain.GolfRowCnt {
			gg.On("IsExposed", c, r).Return(false).Maybe()
		}
	}

	p := &GolfWebPresenter{}
	result := p.Output(gg, nil)
	out := parseGolfOutput(t, result)

	assert.Equal(t, "golf.gameOver", out.MessageCode)
}

func TestGolfWebPresenterHintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		gg := new(interfaces.MockGolfGame)
		gg.On("GetHint").Return(&domain.GolfHint{Type: "remove", Col: 3})
		gg.On("GetPhase").Return(domain.GolfPhasePlaying)
		gg.On("GetMoveCount").Return(0)
		gg.On("GetStockCount").Return(16)
		gg.On("CanUndo").Return(false)
		gg.On("IsStalemate").Return(false)
		gg.On("UndoToEscape").Return(0)

		p := &GolfWebPresenter{}
		result := p.HintOutput(gg)
		out := parseGolfOutput(t, result)

		assert.NotNil(t, out.Hint)
		assert.Equal(t, "remove", out.Hint.Type)
		assert.Equal(t, "golf.hintAvailable", out.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		gg := new(interfaces.MockGolfGame)
		gg.On("GetHint").Return((*domain.GolfHint)(nil))
		gg.On("GetPhase").Return(domain.GolfPhasePlaying)
		gg.On("GetMoveCount").Return(0)
		gg.On("GetStockCount").Return(0)
		gg.On("CanUndo").Return(false)
		gg.On("IsStalemate").Return(false)
		gg.On("UndoToEscape").Return(0)

		p := &GolfWebPresenter{}
		result := p.HintOutput(gg)
		out := parseGolfOutput(t, result)

		assert.Nil(t, out.Hint)
		assert.Equal(t, "golf.noHint", out.MessageCode)
	})
}

func TestGolfWebPresenterActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		gg := new(interfaces.MockGolfGame)
		gg.On("GetPhase").Return(domain.GolfPhasePlaying)

		gg.On("GetGameEndFlag").Return(false)
		p := &GolfWebPresenter{}
		result := p.ActionLogOutput(gg)
		assert.Contains(t, result, "entries")
	})

	t.Run("game over", func(t *testing.T) {
		gg := new(interfaces.MockGolfGame)
		gg.On("GetPhase").Return(domain.GolfPhaseGameOver)
		gg.On("GetGameEndFlag").Return(true)
		gg.On("GetActionLog").Return([]*domain.ActionLogEntry{})

		p := &GolfWebPresenter{}
		result := p.ActionLogOutput(gg)
		assert.Contains(t, result, "entries")
	})
}
