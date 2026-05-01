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

func setupPyramidWebMockDefaults(pg *interfaces.MockPyramidGame) {
	pg.On("GetPhase").Return(domain.PyramidPhasePlaying).Maybe()
	pg.On("GetMoveCount").Return(0).Maybe()
	pg.On("GetStockCount").Return(24).Maybe()
	pg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	pg.On("CanUndo").Return(false).Maybe()
	pg.On("IsStalemate").Return(false).Maybe()
	pg.On("UndoToEscape").Return(0).Maybe()
	pg.On("AllRemoved").Return(false).Maybe()

	var pyramid [domain.PyramidRowCnt][]*domain.PyramidCard
	for row := range domain.PyramidRowCnt {
		pyramid[row] = make([]*domain.PyramidCard, row+1)
		for col := range row + 1 {
			pyramid[row][col] = &domain.PyramidCard{
				Card:    domain.NewCard(domain.CardDesignSpade, (row+col)%13+1, false),
				Removed: false,
			}
		}
	}
	pg.On("GetPyramid").Return(pyramid).Maybe()
	// IsExposed for all cards
	for row := range domain.PyramidRowCnt {
		for col := range row + 1 {
			if row == domain.PyramidRowCnt-1 {
				pg.On("IsExposed", row, col).Return(true).Maybe()
			} else {
				pg.On("IsExposed", row, col).Return(false).Maybe()
			}
		}
	}
}

func parsePyramidOutput(t *testing.T, jsonStr string) *controller.PyramidWebOutput {
	t.Helper()
	var out controller.PyramidWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

func TestPyramidWebPresenterOutput_Playing(t *testing.T) {
	pg := new(interfaces.MockPyramidGame)
	setupPyramidWebMockDefaults(pg)
	p := &PyramidWebPresenter{}

	result := p.Output(pg, nil)
	out := parsePyramidOutput(t, result)

	assert.Equal(t, 0, out.Phase)
	assert.Equal(t, 24, out.StockCount)
	assert.Equal(t, "pyramid.playing", out.MessageCode)
	assert.Len(t, out.Pyramid, domain.PyramidRowCnt)
}

func TestPyramidWebPresenterOutput_Error(t *testing.T) {
	pg := new(interfaces.MockPyramidGame)
	setupPyramidWebMockDefaults(pg)
	p := &PyramidWebPresenter{}

	result := p.Output(pg, errors.New("test error"))
	out := parsePyramidOutput(t, result)

	assert.Equal(t, "test error", out.Message)
}

func TestPyramidWebPresenterOutput_GameClear(t *testing.T) {
	pg := new(interfaces.MockPyramidGame)
	setupPyramidWebMockDefaults(pg)
	pg.ExpectedCalls = nil
	pg.On("GetPhase").Return(domain.PyramidPhaseGameClear).Maybe()
	pg.On("GetMoveCount").Return(10).Maybe()
	pg.On("GetStockCount").Return(0).Maybe()
	pg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	pg.On("CanUndo").Return(false).Maybe()
	pg.On("IsStalemate").Return(false).Maybe()
	pg.On("UndoToEscape").Return(0).Maybe()

	var pyramid [domain.PyramidRowCnt][]*domain.PyramidCard
	for row := range domain.PyramidRowCnt {
		pyramid[row] = make([]*domain.PyramidCard, row+1)
		for col := range row + 1 {
			pyramid[row][col] = &domain.PyramidCard{
				Card:    domain.NewCard(domain.CardDesignSpade, 1, false),
				Removed: true,
			}
		}
	}
	pg.On("GetPyramid").Return(pyramid).Maybe()
	for row := range domain.PyramidRowCnt {
		for col := range row + 1 {
			pg.On("IsExposed", row, col).Return(false).Maybe()
		}
	}

	p := &PyramidWebPresenter{}
	result := p.Output(pg, nil)
	out := parsePyramidOutput(t, result)

	assert.Equal(t, 1, out.Phase)
	assert.Equal(t, "pyramid.gameClear", out.MessageCode)
	assert.Equal(t, "10", out.MessageParams["moveCount"])
}

func TestPyramidWebPresenterOutput_GameOver(t *testing.T) {
	pg := new(interfaces.MockPyramidGame)
	setupPyramidWebMockDefaults(pg)
	pg.ExpectedCalls = nil
	pg.On("GetPhase").Return(domain.PyramidPhaseGameOver).Maybe()
	pg.On("GetMoveCount").Return(5).Maybe()
	pg.On("GetStockCount").Return(0).Maybe()
	pg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	pg.On("CanUndo").Return(false).Maybe()
	pg.On("IsStalemate").Return(false).Maybe()
	pg.On("UndoToEscape").Return(0).Maybe()

	var pyramid [domain.PyramidRowCnt][]*domain.PyramidCard
	for row := range domain.PyramidRowCnt {
		pyramid[row] = make([]*domain.PyramidCard, row+1)
		for col := range row + 1 {
			pyramid[row][col] = &domain.PyramidCard{Card: domain.NewCard(domain.CardDesignSpade, 1, false), Removed: false}
		}
	}
	pg.On("GetPyramid").Return(pyramid).Maybe()
	for row := range domain.PyramidRowCnt {
		for col := range row + 1 {
			pg.On("IsExposed", row, col).Return(row == domain.PyramidRowCnt-1).Maybe()
		}
	}

	p := &PyramidWebPresenter{}
	result := p.Output(pg, nil)
	out := parsePyramidOutput(t, result)

	assert.Equal(t, 2, out.Phase)
	assert.Equal(t, "pyramid.gameOver", out.MessageCode)
}

func TestPyramidWebPresenterOutput_Stalemate(t *testing.T) {
	pg := new(interfaces.MockPyramidGame)
	setupPyramidWebMockDefaults(pg)
	pg.ExpectedCalls = nil
	pg.On("GetPhase").Return(domain.PyramidPhasePlaying).Maybe()
	pg.On("GetMoveCount").Return(5).Maybe()
	pg.On("GetStockCount").Return(0).Maybe()
	pg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	pg.On("CanUndo").Return(false).Maybe()
	pg.On("IsStalemate").Return(true).Maybe()
	pg.On("UndoToEscape").Return(-1).Maybe()

	var pyramid [domain.PyramidRowCnt][]*domain.PyramidCard
	for row := range domain.PyramidRowCnt {
		pyramid[row] = make([]*domain.PyramidCard, row+1)
		for col := range row + 1 {
			pyramid[row][col] = &domain.PyramidCard{Card: domain.NewCard(domain.CardDesignSpade, 1, false), Removed: false}
		}
	}
	pg.On("GetPyramid").Return(pyramid).Maybe()
	for row := range domain.PyramidRowCnt {
		for col := range row + 1 {
			pg.On("IsExposed", row, col).Return(row == domain.PyramidRowCnt-1).Maybe()
		}
	}

	p := &PyramidWebPresenter{}
	result := p.Output(pg, nil)
	out := parsePyramidOutput(t, result)

	assert.Equal(t, "pyramid.stalemate", out.MessageCode)
}

func TestPyramidWebPresenterHintOutput_WithHint(t *testing.T) {
	pg := new(interfaces.MockPyramidGame)
	pg.On("GetHint").Return(&domain.PyramidHint{Type: "king", Row1: 6, Col1: 0, Row2: -1, Col2: -1})
	pg.On("GetPhase").Return(domain.PyramidPhasePlaying)
	pg.On("GetMoveCount").Return(0)
	pg.On("GetStockCount").Return(24)
	pg.On("CanUndo").Return(false)
	pg.On("IsStalemate").Return(false)
	pg.On("UndoToEscape").Return(0)

	p := &PyramidWebPresenter{}
	result := p.HintOutput(pg)
	out := parsePyramidOutput(t, result)

	assert.NotNil(t, out.Hint)
	assert.Equal(t, "king", out.Hint.Type)
	assert.Equal(t, "pyramid.hintAvailable", out.MessageCode)
}

func TestPyramidWebPresenterHintOutput_NoHint(t *testing.T) {
	pg := new(interfaces.MockPyramidGame)
	pg.On("GetHint").Return((*domain.PyramidHint)(nil))
	pg.On("GetPhase").Return(domain.PyramidPhasePlaying)
	pg.On("GetMoveCount").Return(0)
	pg.On("GetStockCount").Return(24)
	pg.On("CanUndo").Return(false)
	pg.On("IsStalemate").Return(false)
	pg.On("UndoToEscape").Return(0)

	p := &PyramidWebPresenter{}
	result := p.HintOutput(pg)
	out := parsePyramidOutput(t, result)

	assert.Nil(t, out.Hint)
	assert.Equal(t, "pyramid.noHint", out.MessageCode)
}

func TestPyramidWebPresenterActionLogOutput_Playing(t *testing.T) {
	pg := new(interfaces.MockPyramidGame)
	pg.On("GetPhase").Return(domain.PyramidPhasePlaying)

	pg.On("GetGameEndFlag").Return(false)
	p := &PyramidWebPresenter{}
	result := p.ActionLogOutput(pg)
	assert.Contains(t, result, "entries")
}

func TestPyramidWebPresenterActionLogOutput_GameOver(t *testing.T) {
	pg := new(interfaces.MockPyramidGame)
	pg.On("GetPhase").Return(domain.PyramidPhaseGameOver)
	pg.On("GetGameEndFlag").Return(true)
	pg.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, ActionType: "draw", Detail: "test"},
	})

	p := &PyramidWebPresenter{}
	result := p.ActionLogOutput(pg)
	assert.Contains(t, result, "draw")
}
