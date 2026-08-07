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

func setupPokerSquaresWebMockDefaults(pg *interfaces.MockPokerSquaresGame) {
	pg.On("GetPhase").Return(domain.PokerSquaresPhasePlaying).Maybe()
	pg.On("GetHint").Return((*domain.PokerSquaresHint)(nil)).Maybe()
	pg.On("GetPlacedCount").Return(0).Maybe()
	pg.On("CanUndo").Return(false).Maybe()
	pg.On("GetCurrentCard").Return(domain.NewCard(domain.CardDesignSpade, 5, false)).Maybe()
	var board [domain.PokerSquaresGridSize][domain.PokerSquaresGridSize]*domain.Card
	pg.On("GetBoard").Return(board).Maybe()
	for i := 0; i < domain.PokerSquaresGridSize; i++ {
		pg.On("RowScore", i).Return(0).Maybe()
		pg.On("ColScore", i).Return(0).Maybe()
	}
	pg.On("TotalScore").Return(0).Maybe()
}

func parsePokerSquaresOutput(t *testing.T, s string) *controller.PokerSquaresWebOutput {
	t.Helper()
	var out controller.PokerSquaresWebOutput
	err := json.Unmarshal([]byte(s), &out)
	assert.NoError(t, err)
	return &out
}

func TestPokerSquaresWebPresenter_Output_Playing(t *testing.T) {
	pg := new(interfaces.MockPokerSquaresGame)
	setupPokerSquaresWebMockDefaults(pg)
	p := &PokerSquaresWebPresenter{}
	out := parsePokerSquaresOutput(t, p.Output(pg, nil))

	assert.Equal(t, 0, out.Phase)
	assert.Equal(t, "pokersquares.playing", out.MessageCode)
	assert.Len(t, out.Board, domain.PokerSquaresGridSize)
	assert.Len(t, out.RowScores, domain.PokerSquaresGridSize)
	assert.Len(t, out.ColScores, domain.PokerSquaresGridSize)
	assert.NotNil(t, out.CurrentCard)
}

func TestPokerSquaresWebPresenter_Output_Error(t *testing.T) {
	pg := new(interfaces.MockPokerSquaresGame)
	setupPokerSquaresWebMockDefaults(pg)
	p := &PokerSquaresWebPresenter{}
	out := parsePokerSquaresOutput(t, p.Output(pg, errors.New("boom")))
	assert.Equal(t, "boom", out.Message)
}

func TestPokerSquaresWebPresenter_Output_Complete(t *testing.T) {
	pg := new(interfaces.MockPokerSquaresGame)
	pg.On("GetPhase").Return(domain.PokerSquaresPhaseComplete).Maybe()
	pg.On("GetHint").Return((*domain.PokerSquaresHint)(nil)).Maybe()
	pg.On("GetPlacedCount").Return(25).Maybe()
	pg.On("CanUndo").Return(false).Maybe()
	pg.On("GetCurrentCard").Return((*domain.Card)(nil)).Maybe()
	var board [domain.PokerSquaresGridSize][domain.PokerSquaresGridSize]*domain.Card
	for r := 0; r < domain.PokerSquaresGridSize; r++ {
		for c := 0; c < domain.PokerSquaresGridSize; c++ {
			board[r][c] = domain.NewCard(domain.CardDesignSpade, 2, false)
		}
	}
	pg.On("GetBoard").Return(board).Maybe()
	for i := 0; i < domain.PokerSquaresGridSize; i++ {
		pg.On("RowScore", i).Return(2).Maybe()
		pg.On("ColScore", i).Return(3).Maybe()
	}
	pg.On("TotalScore").Return(25).Maybe()

	p := &PokerSquaresWebPresenter{}
	out := parsePokerSquaresOutput(t, p.Output(pg, nil))
	assert.Equal(t, 1, out.Phase)
	assert.Equal(t, "pokersquares.complete", out.MessageCode)
	assert.Equal(t, "25", out.MessageParams["totalScore"])
}

func TestPokerSquaresWebPresenter_ActionLog_Playing(t *testing.T) {
	pg := new(interfaces.MockPokerSquaresGame)
	pg.On("GetPhase").Return(domain.PokerSquaresPhasePlaying)
	pg.On("GetHint").Return((*domain.PokerSquaresHint)(nil)).Maybe()
	pg.On("GetGameEndFlag").Return(false)
	p := &PokerSquaresWebPresenter{}
	result := p.ActionLogOutput(pg)
	assert.Contains(t, result, "entries")
}

func TestPokerSquaresWebPresenter_ActionLog_Complete(t *testing.T) {
	pg := new(interfaces.MockPokerSquaresGame)
	pg.On("GetPhase").Return(domain.PokerSquaresPhaseComplete)
	pg.On("GetHint").Return((*domain.PokerSquaresHint)(nil)).Maybe()
	pg.On("GetGameEndFlag").Return(true)
	pg.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, ActionType: "place", Detail: "test"},
	})
	p := &PokerSquaresWebPresenter{}
	result := p.ActionLogOutput(pg)
	assert.Contains(t, result, "place")
}
