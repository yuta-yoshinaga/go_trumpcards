//go:build test

package presenter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestFreeCellCuiPresenterOutputPlaying(t *testing.T) {
	p := new(FreeCellCuiPresenter)
	f := domain.NewFreeCell(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.FreeCellPhasePlaying)

	result := p.Output(f, nil)

	assert.Contains(t, result, "FreeCell")
	assert.Contains(t, result, "FreeCells:")
	assert.Contains(t, result, "Foundation:")
	assert.Contains(t, result, "手数:")
}

func TestFreeCellCuiPresenterOutputGameClear(t *testing.T) {
	p := new(FreeCellCuiPresenter)
	f := domain.NewFreeCell(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.FreeCellPhaseGameClear)

	result := p.Output(f, nil)

	assert.Contains(t, result, "ゲームクリア")
}

func TestFreeCellCuiPresenterOutputGameOver(t *testing.T) {
	p := new(FreeCellCuiPresenter)
	f := domain.NewFreeCell(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.FreeCellPhaseGameOver)

	result := p.Output(f, nil)

	assert.Contains(t, result, "ゲームオーバー")
}

func TestFreeCellCuiPresenterOutputStalemate(t *testing.T) {
	p := new(FreeCellCuiPresenter)
	f := domain.NewFreeCell(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.FreeCellPhasePlaying)
	f.SetIsStalemate(true)

	result := p.Output(f, nil)

	assert.Contains(t, result, "手詰まりです")
}

func TestFreeCellCuiPresenterOutputError(t *testing.T) {
	p := new(FreeCellCuiPresenter)
	f := domain.NewFreeCell(domain.NewTrumpCards(0))
	f.Reset()

	result := p.Output(f, errors.New("test error"))

	assert.Contains(t, result, "test error")
}

func TestFreeCellCuiPresenterOutputFreeCellsOccupied(t *testing.T) {
	p := new(FreeCellCuiPresenter)
	f := domain.NewFreeCell(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.FreeCellPhasePlaying)

	var cells [domain.FreeCellCellCnt]*domain.Card
	cells[0] = domain.NewCard(domain.CardDesignSpade, 5, false)
	f.SetFreeCells(cells)

	result := p.Output(f, nil)

	assert.Contains(t, result, "FreeCells:")
	assert.Contains(t, result, "SPADE 5")
}

func TestFreeCellCuiPresenterOutputFoundationWithCards(t *testing.T) {
	p := new(FreeCellCuiPresenter)
	f := domain.NewFreeCell(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.FreeCellPhasePlaying)

	var foundation [domain.FreeCellFoundationCnt][]*domain.Card
	foundation[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
	f.SetFoundation(foundation)

	result := p.Output(f, nil)

	assert.Contains(t, result, "Foundation:")
	assert.Contains(t, result, "SPADE 1")
}

func TestFreeCellCuiPresenterOutputEmptyTableau(t *testing.T) {
	p := new(FreeCellCuiPresenter)
	f := domain.NewFreeCell(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.FreeCellPhasePlaying)

	var emptyTableau [domain.FreeCellTableauCnt][]*domain.Card
	f.SetTableau(emptyTableau)

	result := p.Output(f, nil)

	assert.Contains(t, result, "[空]")
}

func TestFreeCellCuiPresenterHintTableau(t *testing.T) {
	p := new(FreeCellCuiPresenter)
	f := domain.NewFreeCell(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.FreeCellPhasePlaying)

	// Place an Ace on tableau so hint suggests moving to foundation
	var tableau [domain.FreeCellTableauCnt][]*domain.Card
	tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
	f.SetTableau(tableau)

	result := p.HintOutput(f)

	assert.Contains(t, result, "タブロー列")
}

func TestFreeCellCuiPresenterHintFreeCell(t *testing.T) {
	p := new(FreeCellCuiPresenter)
	f := domain.NewFreeCell(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.FreeCellPhasePlaying)

	// Place an Ace in a free cell so hint suggests moving to foundation
	var emptyTableau [domain.FreeCellTableauCnt][]*domain.Card
	f.SetTableau(emptyTableau)
	var cells [domain.FreeCellCellCnt]*domain.Card
	cells[0] = domain.NewCard(domain.CardDesignSpade, 1, false)
	f.SetFreeCells(cells)

	result := p.HintOutput(f)

	assert.Contains(t, result, "フリーセル")
}

func TestFreeCellCuiPresenterHintToTableau(t *testing.T) {
	p := new(FreeCellCuiPresenter)
	f := domain.NewFreeCell(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.FreeCellPhasePlaying)

	// Place a card in free cell and a compatible card on tableau so hint suggests freecell -> tableau
	var tableau [domain.FreeCellTableauCnt][]*domain.Card
	tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 6, false)}
	f.SetTableau(tableau)
	var cells [domain.FreeCellCellCnt]*domain.Card
	cells[0] = domain.NewCard(domain.CardDesignHeart, 5, false)
	f.SetFreeCells(cells)

	result := p.HintOutput(f)

	// Should mention tableau as destination
	assert.Contains(t, result, "タブロー列")
}

func TestFreeCellCuiPresenterHintToFreeCell(t *testing.T) {
	p := new(FreeCellCuiPresenter)
	f := domain.NewFreeCell(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.FreeCellPhasePlaying)

	// Set up tableau with cards that can't go to foundation or other tableau,
	// with all columns filled so there are no empty columns to use as fallback.
	// Only the freecell is a valid destination.
	var tableau [domain.FreeCellTableauCnt][]*domain.Card
	tableau[0] = []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 10, false),
		domain.NewCard(domain.CardDesignSpade, 9, false),
	}
	for i := 1; i < domain.FreeCellTableauCnt; i++ {
		tableau[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 2, false)}
	}
	f.SetTableau(tableau)
	var cells [domain.FreeCellCellCnt]*domain.Card
	f.SetFreeCells(cells)

	result := p.HintOutput(f)

	assert.Contains(t, result, "フリーセル")
}

func TestFreeCellCuiPresenterHintNil(t *testing.T) {
	p := new(FreeCellCuiPresenter)
	f := domain.NewFreeCell(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.FreeCellPhaseGameOver)

	result := p.HintOutput(f)

	assert.Contains(t, result, "ヒントはありません")
}

func TestFreeCellCuiPresenterActionLogPlaying(t *testing.T) {
	p := new(FreeCellCuiPresenter)
	f := domain.NewFreeCell(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.FreeCellPhasePlaying)

	result := p.ActionLogOutput(f)

	assert.Contains(t, result, "棋譜はありません")
}

func TestFreeCellCuiPresenterActionLogGameOver(t *testing.T) {
	p := new(FreeCellCuiPresenter)
	f := domain.NewFreeCell(domain.NewTrumpCards(0))
	f.Reset()

	// Make a move to generate action log
	var tableau [domain.FreeCellTableauCnt][]*domain.Card
	tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
	f.SetTableau(tableau)
	f.SetPhase(domain.FreeCellPhasePlaying)
	_ = f.MoveTableauToFoundation(0)
	f.SetPhase(domain.FreeCellPhaseGameOver)

	result := p.ActionLogOutput(f)

	assert.Contains(t, result, "棋譜")
}

// **何枚まとめて動かせるかが CUI に出ていなかった (#4777)。**姉妹ゲームの
// Seahaven Towers は supermoveLine を出している。
func TestFreeCellCuiPresenterOutput_SupermoveLine(t *testing.T) {
	p := new(FreeCellCuiPresenter)
	card := func(v int) *domain.Card { return domain.NewCard(domain.CardDesignSpade, v, false) }
	board := func(filledCells, filledCols int) *domain.FreeCell {
		f := domain.NewFreeCell(domain.NewTrumpCards(0))
		f.Reset()
		f.SetPhase(domain.FreeCellPhasePlaying)
		var cells [domain.FreeCellCellCnt]*domain.Card
		for i := 0; i < filledCells && i < domain.FreeCellCellCnt; i++ {
			cells[i] = card(i + 2)
		}
		f.SetFreeCells(cells)
		var tableau [domain.FreeCellTableauCnt][]*domain.Card
		for i := 0; i < domain.FreeCellTableauCnt; i++ {
			if i < filledCols {
				tableau[i] = []*domain.Card{card(5)}
			}
		}
		f.SetTableau(tableau)
		return f
	}

	t.Run("names the limit and what it is made of", func(t *testing.T) {
		out := p.Output(board(0, domain.FreeCellTableauCnt), nil)
		assert.Contains(t, out, "最大5枚")
		assert.Contains(t, out, "空きセル4")
		assert.Contains(t, out, "空き列0")
	})

	// **空き列があるときだけ、そこへ置く上限も出す。**同じ数だと嘘になる。
	t.Run("adds the lower empty-column limit when one exists", func(t *testing.T) {
		out := p.Output(board(0, domain.FreeCellTableauCnt-1), nil)
		assert.Contains(t, out, "最大10枚")
		assert.Contains(t, out, "空き列へは5枚")
	})

	t.Run("omits the empty-column limit when no column is empty", func(t *testing.T) {
		out := p.Output(board(0, domain.FreeCellTableauCnt), nil)
		assert.NotContains(t, out, "空き列へは")
	})

	t.Run("shows a limit of one when the board is packed", func(t *testing.T) {
		out := p.Output(board(domain.FreeCellCellCnt, domain.FreeCellTableauCnt), nil)
		assert.Contains(t, out, "最大1枚")
	})

	t.Run("shows nothing once the game is cleared", func(t *testing.T) {
		f := board(0, domain.FreeCellTableauCnt)
		f.SetPhase(domain.FreeCellPhaseGameClear)
		assert.NotContains(t, p.Output(f, nil), "一括移動")
	})
}
