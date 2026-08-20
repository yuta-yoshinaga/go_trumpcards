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

func setupCribbageSquaresWebMockDefaults(pg *interfaces.MockCribbageSquaresGame) {
	pg.On("GetPhase").Return(domain.CribbageSquaresPhasePlaying).Maybe()
	pg.On("GetHint").Return((*domain.CribbageSquaresHint)(nil)).Maybe()
	pg.On("GetPlacedCount").Return(0).Maybe()
	pg.On("CanUndo").Return(false).Maybe()
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
		// **確定ぶんは呼び出し側が先に登録できる。**testify は最初に一致した
		// 登録を返すので、テストが局面を述べたい行はこのヘルパより前に
		// 登録しておけばよい。
		pg.On("RowPartialDetail", i).Return(domain.CribbageScoreDetail{}).Maybe()
		pg.On("ColPartialDetail", i).Return(domain.CribbageScoreDetail{}).Maybe()
	}
}

func parseCribbageSquaresOutput(t *testing.T, s string) *controller.CribbageSquaresWebOutput {
	t.Helper()
	var out controller.CribbageSquaresWebOutput
	err := json.Unmarshal([]byte(s), &out)
	assert.NoError(t, err)
	return &out
}

func TestCribbageSquaresWebPresenter_Output_Playing(t *testing.T) {
	pg := new(interfaces.MockCribbageSquaresGame)
	setupCribbageSquaresWebMockDefaults(pg)
	p := &CribbageSquaresWebPresenter{}
	out := parseCribbageSquaresOutput(t, p.Output(pg, nil))

	assert.Equal(t, 0, out.Phase)
	assert.Equal(t, "cribbagesquares.playing", out.MessageCode)
	assert.Len(t, out.Board, domain.CribbageSquaresGridSize)
	assert.Len(t, out.RowScores, domain.CribbageSquaresGridSize)
	assert.Len(t, out.ColScores, domain.CribbageSquaresGridSize)
	assert.Len(t, out.RowDetails, domain.CribbageSquaresGridSize)
	assert.Len(t, out.ColDetails, domain.CribbageSquaresGridSize)
	assert.NotNil(t, out.CurrentCard)
	// The starter stays face down during play, and the page needs to know the
	// target rather than hardcoding 61 of its own.
	assert.Nil(t, out.Starter)
	assert.Equal(t, domain.CribbageSquaresWinScore, out.WinScore)
	assert.False(t, out.IsWin)
}

func TestCribbageSquaresWebPresenter_Output_Error(t *testing.T) {
	pg := new(interfaces.MockCribbageSquaresGame)
	setupCribbageSquaresWebMockDefaults(pg)
	p := &CribbageSquaresWebPresenter{}
	out := parseCribbageSquaresOutput(t, p.Output(pg, errors.New("boom")))
	assert.Equal(t, "boom", out.Message)
}

func TestCribbageSquaresWebPresenter_Output_Complete(t *testing.T) {
	pg := new(interfaces.MockCribbageSquaresGame)
	pg.On("GetPhase").Return(domain.CribbageSquaresPhaseComplete).Maybe()
	pg.On("GetHint").Return((*domain.CribbageSquaresHint)(nil)).Maybe()
	pg.On("GetPlacedCount").Return(domain.CribbageSquaresTotalCells).Maybe()
	pg.On("CanUndo").Return(false).Maybe()
	pg.On("GetCurrentCard").Return((*domain.Card)(nil)).Maybe()
	var board [domain.CribbageSquaresGridSize][domain.CribbageSquaresGridSize]*domain.Card
	for r := 0; r < domain.CribbageSquaresGridSize; r++ {
		for c := 0; c < domain.CribbageSquaresGridSize; c++ {
			board[r][c] = domain.NewCard(domain.CardDesignSpade, 2, false)
		}
	}
	pg.On("GetBoard").Return(board).Maybe()
	for i := range domain.CribbageSquaresGridSize {
		pg.On("RowScore", i).Return(2).Maybe()
		pg.On("ColScore", i).Return(3).Maybe()
		pg.On("RowDetail", i).Return(domain.CribbageScoreDetail{Pairs: 2, Total: 2}).Maybe()
		pg.On("ColDetail", i).Return(domain.CribbageScoreDetail{Runs: 3, Total: 3}).Maybe()
		pg.On("RowPartialDetail", i).Return(domain.CribbageScoreDetail{Pairs: 2, Total: 2}).Maybe()
		pg.On("ColPartialDetail", i).Return(domain.CribbageScoreDetail{Runs: 3, Total: 3}).Maybe()
	}
	pg.On("TotalScore").Return(20).Maybe()
	pg.On("GetStarter").Return(domain.NewCard(domain.CardDesignHeart, 7, true)).Maybe()
	pg.On("IsWin").Return(false).Maybe()

	p := &CribbageSquaresWebPresenter{}
	out := parseCribbageSquaresOutput(t, p.Output(pg, nil))
	assert.Equal(t, 1, out.Phase)
	// The total is summed from the per-line details, not taken from TotalScore.
	assert.Equal(t, 20, out.TotalScore)
	assert.Equal(t, "cribbagesquares.lose", out.MessageCode)
	assert.Equal(t, "20", out.MessageParams["totalScore"])
	assert.Equal(t, "61", out.MessageParams["winScore"])
	assert.NotNil(t, out.Starter, "the starter is revealed once the board is full")
	assert.Equal(t, 2, out.RowDetails[0].Pairs)
	assert.Equal(t, 3, out.ColDetails[0].Runs)
	assert.False(t, out.IsWin)
}

// Reaching the target has to change the message, or a winning board and a
// losing one are indistinguishable to the page.
func TestCribbageSquaresWebPresenter_Output_Win(t *testing.T) {
	pg := new(interfaces.MockCribbageSquaresGame)
	pg.On("GetPhase").Return(domain.CribbageSquaresPhaseComplete).Maybe()
	pg.On("GetHint").Return((*domain.CribbageSquaresHint)(nil)).Maybe()
	pg.On("GetPlacedCount").Return(domain.CribbageSquaresTotalCells).Maybe()
	pg.On("CanUndo").Return(false).Maybe()
	pg.On("GetCurrentCard").Return((*domain.Card)(nil)).Maybe()
	var board [domain.CribbageSquaresGridSize][domain.CribbageSquaresGridSize]*domain.Card
	pg.On("GetBoard").Return(board).Maybe()
	for i := range domain.CribbageSquaresGridSize {
		pg.On("RowScore", i).Return(8).Maybe()
		pg.On("ColScore", i).Return(8).Maybe()
		pg.On("RowDetail", i).Return(domain.CribbageScoreDetail{Fifteens: 8, Total: 8}).Maybe()
		pg.On("ColDetail", i).Return(domain.CribbageScoreDetail{Fifteens: 8, Total: 8}).Maybe()
		pg.On("RowPartialDetail", i).Return(domain.CribbageScoreDetail{Fifteens: 8, Total: 8}).Maybe()
		pg.On("ColPartialDetail", i).Return(domain.CribbageScoreDetail{Fifteens: 8, Total: 8}).Maybe()
	}
	pg.On("TotalScore").Return(64).Maybe()
	pg.On("GetStarter").Return(domain.NewCard(domain.CardDesignHeart, 7, true)).Maybe()
	pg.On("IsWin").Return(true).Maybe()

	out := parseCribbageSquaresOutput(t, (&CribbageSquaresWebPresenter{}).Output(pg, nil))
	assert.Equal(t, "cribbagesquares.win", out.MessageCode)
	assert.Equal(t, 64, out.TotalScore)
	assert.True(t, out.IsWin)
}

func TestCribbageSquaresWebPresenter_ActionLog_Playing(t *testing.T) {
	pg := new(interfaces.MockCribbageSquaresGame)
	pg.On("GetPhase").Return(domain.CribbageSquaresPhasePlaying)
	pg.On("GetHint").Return((*domain.CribbageSquaresHint)(nil)).Maybe()
	pg.On("GetGameEndFlag").Return(false)
	p := &CribbageSquaresWebPresenter{}
	result := p.ActionLogOutput(pg)
	assert.Contains(t, result, "entries")
}

func TestCribbageSquaresWebPresenter_ActionLog_Complete(t *testing.T) {
	pg := new(interfaces.MockCribbageSquaresGame)
	pg.On("GetPhase").Return(domain.CribbageSquaresPhaseComplete)
	pg.On("GetHint").Return((*domain.CribbageSquaresHint)(nil)).Maybe()
	pg.On("GetGameEndFlag").Return(true)
	pg.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, ActionType: "place", Detail: "test"},
	})
	p := &CribbageSquaresWebPresenter{}
	result := p.ActionLogOutput(pg)
	assert.Contains(t, result, "place")
}

// **シナジー考慮ヒントは CUI しか受け取れていなかった (#4790)。**Web の
// HintOutput は状態をそのまま返すだけだった。
func TestCribbageSquaresWebPresenter_HintOutput(t *testing.T) {
	pr := new(CribbageSquaresWebPresenter)
	decode := func(js string) *controller.CribbageSquaresWebOutput {
		var out controller.CribbageSquaresWebOutput
		assert.NoError(t, json.Unmarshal([]byte(js), &out))
		return &out
	}
	game := func(hint *domain.CribbageSquaresHint) *interfaces.MockCribbageSquaresGame {
		m := new(interfaces.MockCribbageSquaresGame)
		// **先に登録した期待が勝つ。**defaults の GetHint(nil) より前に置く。
		m.On("GetHint").Return(hint)
		setupCribbageSquaresWebMockDefaults(m)
		return m
	}

	t.Run("carries the suggested cell and its synergy", func(t *testing.T) {
		out := decode(pr.HintOutput(game(&domain.CribbageSquaresHint{Row: 2, Col: 3, Score: 12, Synergy: true})))
		if assert.NotNil(t, out.Hint) {
			assert.Equal(t, 2, out.Hint.Row)
			assert.Equal(t, 3, out.Hint.Col)
			assert.Equal(t, 12, out.Hint.Score)
			assert.True(t, out.Hint.Synergy)
		}
		assert.Equal(t, "cribbagesquares.hintAvailable", out.MessageCode)
	})

	// **押したときだけ出す印を付ける。**Output() は常に hint を載せるので、
	// 要求印が無いとページが「押していないのに助言」を出す。
	t.Run("marks the response as a requested hint", func(t *testing.T) {
		plain := decode(pr.Output(game(&domain.CribbageSquaresHint{Row: 0, Col: 0}), nil))
		assert.NotEqual(t, "cribbagesquares.hintAvailable", plain.MessageCode)
	})

	t.Run("reports when there is no hint", func(t *testing.T) {
		out := decode(pr.HintOutput(game(nil)))
		assert.Nil(t, out.Hint)
		assert.Equal(t, "cribbagesquares.noHint", out.MessageCode)
	})

	// **Output() でも hint を載せる。**command:"hint" 専用のレスポンスは
	// ページの state にマージされないので、ここが nil だと state.hint が
	// 永久に undefined になる (#4483)。
	t.Run("Output carries the hint into the page state", func(t *testing.T) {
		out := decode(pr.Output(game(&domain.CribbageSquaresHint{Row: 1, Col: 3, Score: 5, Synergy: true}), nil))
		if assert.NotNil(t, out.Hint) {
			assert.Equal(t, 1, out.Hint.Row)
			assert.Equal(t, 3, out.Hint.Col)
		}
	})
}

// #6088: `rowScores` / `rowDetails` は `RowDetail` 由来で、スターターが出る
// 16 枚目まで**必ず 0**。つまり Web は対局中の内訳を一切出せていなかった。
// CUI は #6083 でスターター抜きの確定ぶんを出している。
func TestCribbageSquaresWebPresenter_CarriesThePartialDetails(t *testing.T) {
	pg := new(interfaces.MockCribbageSquaresGame)
	// 行 0 に 5 と 10 が置かれていて 15 が 1 つ確定している局面。スターターは
	// まだ伏せたままなので RowDetail は 0 のまま。**既定より先に登録する。**
	pg.On("RowPartialDetail", 0).Return(domain.CribbageScoreDetail{Fifteens: 2, Total: 2})
	setupCribbageSquaresWebMockDefaults(pg)

	p := &CribbageSquaresWebPresenter{}
	out := parseCribbageSquaresOutput(t, p.Output(pg, nil))

	// **公式のスコアは 0 のまま。**低い数字を確定点と誤解させない。
	assert.Equal(t, 0, out.RowScores[0], "スターター前の公式スコアは 0 のまま")
	assert.Equal(t, 0, out.RowDetails[0].Total)
	// 確定ぶんは別枠で運ぶ。
	assert.Equal(t, 2, out.RowPartialDetails[0].Total, "15 の 2 点が確定しているのに 0")
	assert.Equal(t, 2, out.RowPartialDetails[0].Fifteens)
	assert.Equal(t, 0, out.RowPartialDetails[1].Total, "何も置いていない行に点が付いている")
	assert.Len(t, out.ColPartialDetails, domain.CribbageSquaresGridSize)
}
