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

	// Set up tableau with cards that can't go to foundation or other tableau, only freecell
	var tableau [domain.FreeCellTableauCnt][]*domain.Card
	tableau[0] = []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 10, false),
		domain.NewCard(domain.CardDesignSpade, 9, false),
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
