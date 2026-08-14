//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupCribbageSquaresCuiMock() *interfaces.MockCribbageSquaresGame {
	pg := new(interfaces.MockCribbageSquaresGame)
	pg.On("GetPhase").Return(domain.CribbageSquaresPhasePlaying).Maybe()
	pg.On("GetPlacedCount").Return(0).Maybe()
	pg.On("GetCurrentCard").Return(domain.NewCard(domain.CardDesignSpade, 5, false)).Maybe()
	var board [domain.CribbageSquaresGridSize][domain.CribbageSquaresGridSize]*domain.Card
	pg.On("GetBoard").Return(board).Maybe()
	for i := 0; i < domain.CribbageSquaresGridSize; i++ {
		pg.On("RowScore", i).Return(0).Maybe()
		pg.On("ColScore", i).Return(0).Maybe()
	}
	pg.On("TotalScore").Return(0).Maybe()
	pg.On("GetStarter").Return((*domain.Card)(nil)).Maybe()
	pg.On("IsWin").Return(false).Maybe()
	for i := range domain.CribbageSquaresGridSize {
		pg.On("RowDetail", i).Return(domain.CribbageScoreDetail{}).Maybe()
		pg.On("ColDetail", i).Return(domain.CribbageScoreDetail{}).Maybe()
	}
	return pg
}

func TestCribbageSquaresCuiPresenter_Output_Playing(t *testing.T) {
	pg := setupCribbageSquaresCuiMock()
	p := &CribbageSquaresCuiPresenter{}
	out := p.Output(pg, nil)
	assert.Contains(t, out, "Cribbage Squares")
	assert.Contains(t, out, "手持ちカード:")
	assert.Contains(t, out, "配置済み: 0/16", "the grid is 4x4, not 5x5")
	assert.Contains(t, out, "カードを置く: p")
	// The starter is hidden during play, and saying so beats an absent line
	// that reads as a rendering bug.
	assert.Contains(t, out, "伏せたまま")
}

func TestCribbageSquaresCuiPresenter_Output_Complete(t *testing.T) {
	pg := new(interfaces.MockCribbageSquaresGame)
	pg.On("GetPhase").Return(domain.CribbageSquaresPhaseComplete).Maybe()
	pg.On("GetPlacedCount").Return(domain.CribbageSquaresTotalCells).Maybe()
	pg.On("GetCurrentCard").Return((*domain.Card)(nil)).Maybe()
	var board [domain.CribbageSquaresGridSize][domain.CribbageSquaresGridSize]*domain.Card
	for r := 0; r < domain.CribbageSquaresGridSize; r++ {
		for c := 0; c < domain.CribbageSquaresGridSize; c++ {
			board[r][c] = domain.NewCard(domain.CardDesignSpade, 2, false)
		}
	}
	pg.On("GetBoard").Return(board).Maybe()
	for i := 0; i < domain.CribbageSquaresGridSize; i++ {
		pg.On("RowScore", i).Return(2).Maybe()
		pg.On("ColScore", i).Return(2).Maybe()
	}
	pg.On("TotalScore").Return(20).Maybe()
	pg.On("GetStarter").Return(domain.NewCard(domain.CardDesignHeart, 7, true)).Maybe()
	pg.On("IsWin").Return(false).Maybe()
	for i := range domain.CribbageSquaresGridSize {
		pg.On("RowDetail", i).Return(domain.CribbageScoreDetail{Pairs: 2, Total: 2}).Maybe()
		pg.On("ColDetail", i).Return(domain.CribbageScoreDetail{Pairs: 2, Total: 2}).Maybe()
	}

	p := &CribbageSquaresCuiPresenter{}
	out := p.Output(pg, nil)
	assert.Contains(t, out, "ゲーム終了")
	assert.Contains(t, out, "20")
	assert.Contains(t, out, "61", "the target is shown next to the score")
	assert.Contains(t, out, "スターター: ", "the starter is revealed once the board is full")
	// The per-hand breakdown: a bare total says nothing about what worked.
	assert.Contains(t, out, "行0:")
	assert.Contains(t, out, "列3:")
	assert.Contains(t, out, "ペア2")
	assert.NotContains(t, out, "カードを置く: p")
}

// A hand that scored nothing must say so rather than print an empty bracket.
func TestCribbageSquaresCuiPresenter_DetailLine_Zero(t *testing.T) {
	line := cribbageSquaresDetailLine("行0", domain.CribbageScoreDetail{})
	assert.Contains(t, line, "得点なし")
	assert.NotContains(t, line, "ペア")
}

// Only the components that actually scored are listed.
func TestCribbageSquaresCuiPresenter_DetailLine_DropsZeroParts(t *testing.T) {
	line := cribbageSquaresDetailLine("行1", domain.CribbageScoreDetail{Fifteens: 4, Nobs: 1, Total: 5})
	assert.Contains(t, line, "15が4")
	assert.Contains(t, line, "ノブズ1")
	assert.NotContains(t, line, "ペア")
	assert.NotContains(t, line, "ラン")
	assert.NotContains(t, line, "フラッシュ")
}

// The verdict has to differ, or a 61-point board and a 5-point board read the same.
func TestCribbageSquaresCuiPresenter_Output_Win(t *testing.T) {
	pg := new(interfaces.MockCribbageSquaresGame)
	pg.On("GetPhase").Return(domain.CribbageSquaresPhaseComplete).Maybe()
	pg.On("GetPlacedCount").Return(domain.CribbageSquaresTotalCells).Maybe()
	pg.On("GetCurrentCard").Return((*domain.Card)(nil)).Maybe()
	var board [domain.CribbageSquaresGridSize][domain.CribbageSquaresGridSize]*domain.Card
	pg.On("GetBoard").Return(board).Maybe()
	for i := range domain.CribbageSquaresGridSize {
		pg.On("RowScore", i).Return(8).Maybe()
		pg.On("ColScore", i).Return(8).Maybe()
		pg.On("RowDetail", i).Return(domain.CribbageScoreDetail{Fifteens: 8, Total: 8}).Maybe()
		pg.On("ColDetail", i).Return(domain.CribbageScoreDetail{Fifteens: 8, Total: 8}).Maybe()
	}
	pg.On("TotalScore").Return(64).Maybe()
	pg.On("GetStarter").Return(domain.NewCard(domain.CardDesignHeart, 7, true)).Maybe()
	pg.On("IsWin").Return(true).Maybe()

	out := (&CribbageSquaresCuiPresenter{}).Output(pg, nil)
	assert.Contains(t, out, "64")
	assert.Contains(t, out, "61")
}

func TestCribbageSquaresCuiPresenter_Hint_Synergy(t *testing.T) {
	pg := new(interfaces.MockCribbageSquaresGame)
	pg.On("GetHint").Return(&domain.CribbageSquaresHint{Row: 1, Col: 2, Score: 6, Synergy: true})
	pg.On("GetCurrentCard").Return(domain.NewCard(domain.CardDesignSpade, 5, false))
	p := &CribbageSquaresCuiPresenter{}
	out := p.HintOutput(pg)
	assert.Contains(t, out, "ヒント")
	assert.Contains(t, out, "(1,2)")
	assert.Contains(t, out, "15")
}

func TestCribbageSquaresCuiPresenter_Hint_NoSynergy(t *testing.T) {
	pg := new(interfaces.MockCribbageSquaresGame)
	pg.On("GetHint").Return(&domain.CribbageSquaresHint{Row: 0, Col: 0, Score: 0, Synergy: false})
	pg.On("GetCurrentCard").Return(domain.NewCard(domain.CardDesignHeart, 9, false))
	p := &CribbageSquaresCuiPresenter{}
	out := p.HintOutput(pg)
	assert.Contains(t, out, "(0,0)")
	assert.Contains(t, out, "噛み合うカードがありません")
}

func TestCribbageSquaresCuiPresenter_Hint_None(t *testing.T) {
	pg := new(interfaces.MockCribbageSquaresGame)
	pg.On("GetHint").Return((*domain.CribbageSquaresHint)(nil))
	p := &CribbageSquaresCuiPresenter{}
	out := p.HintOutput(pg)
	assert.Contains(t, out, "ヒントはありません")
}

func TestCribbageSquaresCuiPresenter_Hint_NoCurrentCard(t *testing.T) {
	pg := new(interfaces.MockCribbageSquaresGame)
	pg.On("GetHint").Return(&domain.CribbageSquaresHint{Row: 3, Col: 3, Score: 2, Synergy: true})
	pg.On("GetCurrentCard").Return((*domain.Card)(nil))
	p := &CribbageSquaresCuiPresenter{}
	out := p.HintOutput(pg)
	assert.Contains(t, out, "(3,3)")
}

func TestCribbageSquaresCuiPresenter_ActionLog_Playing(t *testing.T) {
	pg := new(interfaces.MockCribbageSquaresGame)
	pg.On("GetPhase").Return(domain.CribbageSquaresPhasePlaying)
	p := &CribbageSquaresCuiPresenter{}
	assert.Contains(t, p.ActionLogOutput(pg), "棋譜はありません")
}

func TestCribbageSquaresCuiPresenter_ActionLog_Complete(t *testing.T) {
	pg := new(interfaces.MockCribbageSquaresGame)
	pg.On("GetPhase").Return(domain.CribbageSquaresPhaseComplete)
	pg.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, ActionType: "place", Detail: "test"},
	})
	p := &CribbageSquaresCuiPresenter{}
	out := p.ActionLogOutput(pg)
	assert.Contains(t, out, "place")
}
