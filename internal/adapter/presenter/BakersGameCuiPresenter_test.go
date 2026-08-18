//go:build test

package presenter

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestBakersGameCuiPresenterOutputPlaying(t *testing.T) {
	p := new(BakersGameCuiPresenter)
	f := domain.NewDefaultBakersGame()
	f.Reset()
	f.SetPhase(domain.FreeCellPhasePlaying)

	result := p.Output(f, nil)

	assert.Contains(t, result, "Baker")
	assert.Contains(t, result, "FreeCells:")
	assert.Contains(t, result, "Foundation:")
	assert.Contains(t, result, "手数:")
}

func TestBakersGameCuiPresenterOutputGameClear(t *testing.T) {
	p := new(BakersGameCuiPresenter)
	f := domain.NewDefaultBakersGame()
	f.Reset()
	f.SetPhase(domain.FreeCellPhaseGameClear)

	assert.Contains(t, p.Output(f, nil), "ゲームクリア")
}

func TestBakersGameCuiPresenterOutputGameOver(t *testing.T) {
	p := new(BakersGameCuiPresenter)
	f := domain.NewDefaultBakersGame()
	f.Reset()
	f.SetPhase(domain.FreeCellPhaseGameOver)

	assert.Contains(t, p.Output(f, nil), "ゲームオーバー")
}

func TestBakersGameCuiPresenterOutputStalemate(t *testing.T) {
	p := new(BakersGameCuiPresenter)
	f := domain.NewDefaultBakersGame()
	f.Reset()
	f.SetPhase(domain.FreeCellPhasePlaying)
	f.SetIsStalemate(true)

	assert.Contains(t, p.Output(f, nil), "手詰まりです")
}

func TestBakersGameCuiPresenterOutputError(t *testing.T) {
	p := new(BakersGameCuiPresenter)
	f := domain.NewDefaultBakersGame()
	f.Reset()

	assert.Contains(t, p.Output(f, errors.New("test error")), "test error")
}

func TestBakersGameCuiPresenterOutputEmptyTableau(t *testing.T) {
	p := new(BakersGameCuiPresenter)
	f := domain.NewDefaultBakersGame()
	f.Reset()
	f.SetPhase(domain.FreeCellPhasePlaying)

	var emptyTableau [domain.FreeCellTableauCnt][]*domain.Card
	f.SetTableau(emptyTableau)

	assert.Contains(t, p.Output(f, nil), "[空]")
}

func TestBakersGameCuiPresenterHintToFoundation(t *testing.T) {
	p := new(BakersGameCuiPresenter)
	f := domain.NewDefaultBakersGame()
	f.Reset()
	f.SetPhase(domain.FreeCellPhasePlaying)

	var tableau [domain.FreeCellTableauCnt][]*domain.Card
	tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
	f.SetTableau(tableau)

	assert.Contains(t, p.HintOutput(f), "タブロー列")
}

func TestBakersGameCuiPresenterHintFromFreeCell(t *testing.T) {
	p := new(BakersGameCuiPresenter)
	f := domain.NewDefaultBakersGame()
	f.Reset()
	f.SetPhase(domain.FreeCellPhasePlaying)

	var emptyTableau [domain.FreeCellTableauCnt][]*domain.Card
	f.SetTableau(emptyTableau)
	var cells [domain.FreeCellCellCnt]*domain.Card
	cells[0] = domain.NewCard(domain.CardDesignSpade, 1, false)
	f.SetFreeCells(cells)

	assert.Contains(t, p.HintOutput(f), "フリーセル")
}

func TestBakersGameCuiPresenterHintNil(t *testing.T) {
	p := new(BakersGameCuiPresenter)
	f := domain.NewDefaultBakersGame()
	f.Reset()
	f.SetPhase(domain.FreeCellPhaseGameOver)

	assert.Contains(t, p.HintOutput(f), "ヒントはありません")
}

func TestBakersGameCuiPresenterActionLog(t *testing.T) {
	p := new(BakersGameCuiPresenter)
	f := domain.NewDefaultBakersGame()
	f.Reset()
	f.SetPhase(domain.FreeCellPhasePlaying)
	assert.Contains(t, p.ActionLogOutput(f), "棋譜はありません")

	var tableau [domain.FreeCellTableauCnt][]*domain.Card
	tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
	f.SetTableau(tableau)
	_ = f.MoveTableauToFoundation(0)
	f.SetPhase(domain.FreeCellPhaseGameOver)
	assert.Contains(t, p.ActionLogOutput(f), "棋譜")
}

// #5636: FreeCell エンジンを共有する姉妹ゲーム (FreeCell / EightOff / Penguin /
// SeahavenTowers) はどれも一括移動の上限を CUI に出しているのに、Baker's Game
// だけ無かった。Web は常時バッジで出している。
func TestBakersGameCuiPresenterShowsTheSupermoveLimit(t *testing.T) {
	p := new(BakersGameCuiPresenter)
	f := domain.NewDefaultBakersGame()
	f.Reset()
	f.SetPhase(domain.FreeCellPhasePlaying)

	out := p.Output(f, nil)

	// **ドメインが計算した値を出す。**presenter で数え直すと、空き列を経由地に
	// 使えないぶんの差が抜ける。
	assert.Contains(t, out, i18n.Tf("bakersgame.supermoveLine",
		"limit", strconv.Itoa(f.GetMaxMovableCards()),
		"cells", strconv.Itoa(freeCellEmptyCells(f)),
		"cols", strconv.Itoa(freeCellEmptyColumns(f))))
}

// 空き列があるときだけ、低い方の上限も添える (その列自身を経由地に使えないため)。
func TestBakersGameCuiPresenterShowsTheEmptyColumnLimitOnlyWhenThereIsOne(t *testing.T) {
	p := new(BakersGameCuiPresenter)
	f := domain.NewDefaultBakersGame()
	f.Reset()
	f.SetPhase(domain.FreeCellPhasePlaying)
	require.Zero(t, f.GetMaxMovableCardsToEmptyColumn(), "前提: 配り直後は空き列が無い")

	// 枚数を伏せた前置き部分だけで見る (ロケール文字列を testJP に写さないため)。
	toEmptyPrefix, _, ok := strings.Cut(i18n.Tf("bakersgame.supermoveToEmpty", "limit", "\x00"), "\x00")
	require.True(t, ok)
	assert.NotContains(t, p.Output(f, nil), toEmptyPrefix)

	// 1 列空けると出る。
	tableau := f.GetTableau()
	tableau[0] = nil
	f.SetTableau(tableau)
	require.Positive(t, f.GetMaxMovableCardsToEmptyColumn())

	assert.Contains(t, p.Output(f, nil), i18n.Tf("bakersgame.supermoveToEmpty",
		"limit", strconv.Itoa(f.GetMaxMovableCardsToEmptyColumn())))
}
