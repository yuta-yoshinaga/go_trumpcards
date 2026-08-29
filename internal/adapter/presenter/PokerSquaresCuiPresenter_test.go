//go:build test

package presenter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupPokerSquaresCuiMock() *interfaces.MockPokerSquaresGame {
	pg := new(interfaces.MockPokerSquaresGame)
	pg.On("GetPhase").Return(domain.PokerSquaresPhasePlaying).Maybe()
	pg.On("CanUndo").Return(true).Maybe()
	pg.On("GetPlacedCount").Return(0).Maybe()
	pg.On("GetCurrentCard").Return(domain.NewCard(domain.CardDesignSpade, 5, false)).Maybe()
	var board [domain.PokerSquaresGridSize][domain.PokerSquaresGridSize]*domain.Card
	pg.On("GetBoard").Return(board).Maybe()
	for i := 0; i < domain.PokerSquaresGridSize; i++ {
		pg.On("RowScore", i).Return(0).Maybe()
		pg.On("ColScore", i).Return(0).Maybe()
		pg.On("PartialRowRank", i).Return(-1).Maybe()
		pg.On("PartialColRank", i).Return(-1).Maybe()
	}
	pg.On("TotalScore").Return(0).Maybe()
	pg.On("CanUndo").Return(true).Maybe()
	return pg
}

func TestPokerSquaresCuiPresenter_Output_Playing(t *testing.T) {
	pg := setupPokerSquaresCuiMock()
	p := &PokerSquaresCuiPresenter{}
	out := p.Output(pg, nil)
	assert.Contains(t, out, "Poker Squares")
	assert.Contains(t, out, "手持ちカード:")
	assert.Contains(t, out, "配置済み: 0/25")
	assert.Contains(t, out, "カードを置く: p")
}

func TestPokerSquaresCuiPresenter_Output_Complete(t *testing.T) {
	pg := new(interfaces.MockPokerSquaresGame)
	pg.On("GetPhase").Return(domain.PokerSquaresPhaseComplete).Maybe()
	pg.On("CanUndo").Return(true).Maybe()
	pg.On("GetPlacedCount").Return(25).Maybe()
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
		pg.On("ColScore", i).Return(2).Maybe()
		pg.On("PartialRowRank", i).Return(-1).Maybe()
		pg.On("PartialColRank", i).Return(-1).Maybe()
	}
	pg.On("TotalScore").Return(20).Maybe()

	p := &PokerSquaresCuiPresenter{}
	out := p.Output(pg, nil)
	assert.Contains(t, out, "ゲーム終了")
	assert.Contains(t, out, "20")
	assert.NotContains(t, out, "カードを置く: p")
	assert.NotContains(t, out, "部分役") // Negative control: no partial hand name on complete board
	assert.NotContains(t, out, "partial")
}

func TestPokerSquaresCuiPresenter_Hint_Synergy(t *testing.T) {
	pg := new(interfaces.MockPokerSquaresGame)
	pg.On("GetHint").Return(&domain.PokerSquaresHint{Row: 1, Col: 2, Score: 6, Synergy: true})
	pg.On("GetCurrentCard").Return(domain.NewCard(domain.CardDesignSpade, 5, false))
	p := &PokerSquaresCuiPresenter{}
	out := p.HintOutput(pg)
	assert.Contains(t, out, "ヒント")
	assert.Contains(t, out, "(1,2)")
	assert.Contains(t, out, "ペア")
}

func TestPokerSquaresCuiPresenter_Hint_NoSynergy(t *testing.T) {
	pg := new(interfaces.MockPokerSquaresGame)
	pg.On("GetHint").Return(&domain.PokerSquaresHint{Row: 0, Col: 0, Score: 0, Synergy: false})
	pg.On("GetCurrentCard").Return(domain.NewCard(domain.CardDesignHeart, 9, false))
	p := &PokerSquaresCuiPresenter{}
	out := p.HintOutput(pg)
	assert.Contains(t, out, "(0,0)")
	assert.Contains(t, out, "相乗効果はありません")
}

func TestPokerSquaresCuiPresenter_Hint_None(t *testing.T) {
	pg := new(interfaces.MockPokerSquaresGame)
	pg.On("GetHint").Return((*domain.PokerSquaresHint)(nil))
	p := &PokerSquaresCuiPresenter{}
	out := p.HintOutput(pg)
	assert.Contains(t, out, "ヒントはありません")
}

func TestPokerSquaresCuiPresenter_Hint_NoCurrentCard(t *testing.T) {
	pg := new(interfaces.MockPokerSquaresGame)
	pg.On("GetHint").Return(&domain.PokerSquaresHint{Row: 3, Col: 4, Score: 2, Synergy: true})
	pg.On("GetCurrentCard").Return((*domain.Card)(nil))
	p := &PokerSquaresCuiPresenter{}
	out := p.HintOutput(pg)
	assert.Contains(t, out, "(3,4)")
}

func TestPokerSquaresCuiPresenter_ActionLog_Playing(t *testing.T) {
	pg := new(interfaces.MockPokerSquaresGame)
	pg.On("GetPhase").Return(domain.PokerSquaresPhasePlaying)
	pg.On("CanUndo").Return(true).Maybe()
	p := &PokerSquaresCuiPresenter{}
	assert.Contains(t, p.ActionLogOutput(pg), "棋譜はありません")
}

func TestPokerSquaresCuiPresenter_ActionLog_Complete(t *testing.T) {
	pg := new(interfaces.MockPokerSquaresGame)
	pg.On("GetPhase").Return(domain.PokerSquaresPhaseComplete)
	pg.On("CanUndo").Return(true).Maybe()
	pg.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, ActionType: "place", Detail: "test"},
	})
	p := &PokerSquaresCuiPresenter{}
	out := p.ActionLogOutput(pg)
	assert.Contains(t, out, "place")
}

// #5538: Web は Undo ボタンの disabled で押せないことが見えるのに、CUI は
// `u` を打ってエラーが返って初めて分かる状態だった。
func TestPokerSquaresCuiPresenter_Output_UndoAvailability(t *testing.T) {
	p := &PokerSquaresCuiPresenter{}

	outWith := func(canUndo bool) string {
		pg := new(interfaces.MockPokerSquaresGame)
		pg.On("CanUndo").Return(canUndo)
		pg.On("GetPhase").Return(domain.PokerSquaresPhasePlaying).Maybe()
		pg.On("GetPlacedCount").Return(0).Maybe()
		pg.On("GetCurrentCard").Return(domain.NewCard(domain.CardDesignSpade, 5, false)).Maybe()
		var board [domain.PokerSquaresGridSize][domain.PokerSquaresGridSize]*domain.Card
		pg.On("GetBoard").Return(board).Maybe()
		for i := 0; i < domain.PokerSquaresGridSize; i++ {
			pg.On("RowScore", i).Return(0).Maybe()
			pg.On("ColScore", i).Return(0).Maybe()
			pg.On("PartialRowRank", i).Return(-1).Maybe()
			pg.On("PartialColRank", i).Return(-1).Maybe()
		}
		pg.On("TotalScore").Return(0).Maybe()
		return p.Output(pg, nil)
	}

	// 戻せないときだけ注記を出す。
	assert.Contains(t, outWith(false), i18n.T("pokersquares.undoUnavailable"))
	// **戻せるときは何も足さない。**毎手「戻せます」と言われても情報にならない。
	assert.NotContains(t, outWith(true), i18n.T("pokersquares.undoUnavailable"))
}

func TestPokerSquaresCuiPresenter_Output_PartialRank(t *testing.T) {
	pg := new(interfaces.MockPokerSquaresGame)
	pg.On("GetPhase").Return(domain.PokerSquaresPhasePlaying).Maybe()
	pg.On("CanUndo").Return(true).Maybe()
	pg.On("GetPlacedCount").Return(3).Maybe()
	pg.On("GetCurrentCard").Return(domain.NewCard(domain.CardDesignSpade, 5, false)).Maybe()
	var board [domain.PokerSquaresGridSize][domain.PokerSquaresGridSize]*domain.Card
	pg.On("GetBoard").Return(board).Maybe()

	for i := 0; i < domain.PokerSquaresGridSize; i++ {
		pg.On("RowScore", i).Return(0).Maybe()
		pg.On("ColScore", i).Return(0).Maybe()
		if i == 1 {
			pg.On("PartialRowRank", i).Return(domain.PokerHandOnePair).Maybe()
		} else {
			pg.On("PartialRowRank", i).Return(-1).Maybe()
		}
		if i == 2 {
			pg.On("PartialColRank", i).Return(domain.PokerHandHighCard).Maybe()
		} else {
			pg.On("PartialColRank", i).Return(-1).Maybe()
		}
	}
	pg.On("TotalScore").Return(0).Maybe()

	p := &PokerSquaresCuiPresenter{}
	out := p.Output(pg, nil)

	// 行スコアの行を切り出して比べる。出力全体に対する Contains では、
	// 部分役が**どの行に付いたか**を何も見ていないので、行 1 と行 2 を
	// 取り違えても通ってしまう。
	rowLines := []string{}
	for _, l := range strings.Split(out, "\n") {
		if strings.Contains(l, i18n.Tf("pokersquares.rowScore", "score", "0")) {
			rowLines = append(rowLines, l)
		}
	}
	require.Len(t, rowLines, domain.PokerSquaresGridSize)

	assert.Contains(t, rowLines[1], "One Pair", "部分役があるのは行1")
	assert.NotContains(t, rowLines[0], "One Pair", "行0は -1 なので何も付かない")
	assert.NotContains(t, rowLines[0], i18n.T("pokersquares.rowPartialHand"))

	// 列は 1 行にまとめて出るので、列2 の区切りだけを取り出す。
	colLine := ""
	for _, l := range strings.Split(out, "\n") {
		if strings.Contains(l, i18n.Tf("pokersquares.colScore", "idx", "0", "score", "0")) {
			colLine = l
			break
		}
	}
	require.NotEmpty(t, colLine)
	assert.Contains(t, colLine, "High Card", "部分役があるのは列2")
	assert.Equal(t, 1, strings.Count(colLine, "High Card"), "付くのは1列だけ")
}
