//go:build test

package presenter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestEightOffCuiPresenterOutputPlaying(t *testing.T) {
	p := new(EightOffCuiPresenter)
	e := domain.NewEightOff(domain.NewTrumpCards(0))
	e.Reset()
	e.SetPhase(domain.EightOffPhasePlaying)

	result := p.Output(e, nil)

	assert.Contains(t, result, "Eight Off")
	assert.Contains(t, result, "FreeCells:")
	assert.Contains(t, result, "Foundation:")
	assert.Contains(t, result, "手数:")
}

func TestEightOffCuiPresenterSupermoveLine(t *testing.T) {
	p := new(EightOffCuiPresenter)
	e := domain.NewEightOff(domain.NewTrumpCards(0))
	e.Reset()
	e.SetPhase(domain.EightOffPhasePlaying)

	// 6 free cells filled (2 empty) and 7 columns occupied (1 empty) →
	// (1 + 2) * 2^1 = 6 cards movable at once.
	var cells [domain.EightOffCellCnt]*domain.Card
	for i := 0; i < 6; i++ {
		cells[i] = domain.NewCard(domain.CardDesignSpade, i+1, false)
	}
	e.SetFreeCells(cells)
	var tableau [domain.EightOffTableauCnt][]*domain.Card
	for col := 0; col < 7; col++ {
		tableau[col] = []*domain.Card{domain.NewCard(domain.CardDesignHeart, col+1, false)}
	}
	e.SetTableau(tableau)

	result := p.Output(e, nil)
	assert.Contains(t, result, "一度に移動可能: 6枚")
}

func TestEightOffCuiPresenterOutputGameClear(t *testing.T) {
	p := new(EightOffCuiPresenter)
	e := domain.NewEightOff(domain.NewTrumpCards(0))
	e.Reset()
	e.SetPhase(domain.EightOffPhaseGameClear)

	result := p.Output(e, nil)

	assert.Contains(t, result, "ゲームクリア")
}

func TestEightOffCuiPresenterOutputGameOver(t *testing.T) {
	p := new(EightOffCuiPresenter)
	e := domain.NewEightOff(domain.NewTrumpCards(0))
	e.Reset()
	e.SetPhase(domain.EightOffPhaseGameOver)

	result := p.Output(e, nil)

	assert.Contains(t, result, "ゲームオーバー")
}

func TestEightOffCuiPresenterOutputStalemate(t *testing.T) {
	p := new(EightOffCuiPresenter)
	e := domain.NewEightOff(domain.NewTrumpCards(0))
	e.Reset()
	e.SetPhase(domain.EightOffPhasePlaying)
	e.SetIsStalemate(true)

	result := p.Output(e, nil)

	assert.Contains(t, result, "手詰まりです")
}

func TestEightOffCuiPresenterOutputError(t *testing.T) {
	p := new(EightOffCuiPresenter)
	e := domain.NewEightOff(domain.NewTrumpCards(0))
	e.Reset()

	result := p.Output(e, errors.New("test error"))

	assert.Contains(t, result, "test error")
}

func TestEightOffCuiPresenterOutputFreeCellsOccupied(t *testing.T) {
	p := new(EightOffCuiPresenter)
	e := domain.NewEightOff(domain.NewTrumpCards(0))
	e.Reset()
	e.SetPhase(domain.EightOffPhasePlaying)

	var cells [domain.EightOffCellCnt]*domain.Card
	cells[7] = domain.NewCard(domain.CardDesignSpade, 5, false)
	e.SetFreeCells(cells)

	result := p.Output(e, nil)

	assert.Contains(t, result, "FreeCells:")
	assert.Contains(t, result, "SPADE 5")
}

func TestEightOffCuiPresenterOutputFoundationWithCards(t *testing.T) {
	p := new(EightOffCuiPresenter)
	e := domain.NewEightOff(domain.NewTrumpCards(0))
	e.Reset()
	e.SetPhase(domain.EightOffPhasePlaying)

	var foundation [domain.EightOffFoundationCnt][]*domain.Card
	foundation[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
	e.SetFoundation(foundation)

	result := p.Output(e, nil)

	assert.Contains(t, result, "Foundation:")
	assert.Contains(t, result, "SPADE 1")
}

func TestEightOffCuiPresenterOutputEmptyTableau(t *testing.T) {
	p := new(EightOffCuiPresenter)
	e := domain.NewEightOff(domain.NewTrumpCards(0))
	e.Reset()
	e.SetPhase(domain.EightOffPhasePlaying)

	var emptyTableau [domain.EightOffTableauCnt][]*domain.Card
	e.SetTableau(emptyTableau)

	result := p.Output(e, nil)

	assert.Contains(t, result, "[空]")
}

func TestEightOffCuiPresenterHintTableau(t *testing.T) {
	p := new(EightOffCuiPresenter)
	e := domain.NewEightOff(domain.NewTrumpCards(0))
	e.Reset()
	e.SetPhase(domain.EightOffPhasePlaying)

	var tableau [domain.EightOffTableauCnt][]*domain.Card
	tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
	e.SetTableau(tableau)

	result := p.HintOutput(e)

	assert.Contains(t, result, "タブロー列")
}

func TestEightOffCuiPresenterHintFreeCell(t *testing.T) {
	p := new(EightOffCuiPresenter)
	e := domain.NewEightOff(domain.NewTrumpCards(0))
	e.Reset()
	e.SetPhase(domain.EightOffPhasePlaying)

	var emptyTableau [domain.EightOffTableauCnt][]*domain.Card
	e.SetTableau(emptyTableau)
	var cells [domain.EightOffCellCnt]*domain.Card
	cells[0] = domain.NewCard(domain.CardDesignSpade, 1, false)
	e.SetFreeCells(cells)

	result := p.HintOutput(e)

	assert.Contains(t, result, "フリーセル")
}

func TestEightOffCuiPresenterHintToTableau(t *testing.T) {
	p := new(EightOffCuiPresenter)
	e := domain.NewEightOff(domain.NewTrumpCards(0))
	e.Reset()
	e.SetPhase(domain.EightOffPhasePlaying)

	// Place a card in free cell and a compatible (same-suit, +1) card on tableau
	var tableau [domain.EightOffTableauCnt][]*domain.Card
	tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 6, false)}
	e.SetTableau(tableau)
	var cells [domain.EightOffCellCnt]*domain.Card
	cells[0] = domain.NewCard(domain.CardDesignSpade, 5, false)
	e.SetFreeCells(cells)

	result := p.HintOutput(e)

	assert.Contains(t, result, "タブロー列")
}

func TestEightOffCuiPresenterHintToFreeCell(t *testing.T) {
	p := new(EightOffCuiPresenter)
	e := domain.NewEightOff(domain.NewTrumpCards(0))
	e.Reset()
	e.SetPhase(domain.EightOffPhasePlaying)

	// Set up tableau with cards that can't go to foundation or other tableau.
	// Empty columns only accept Kings, so 9 can't be placed in empty.
	var tableau [domain.EightOffTableauCnt][]*domain.Card
	tableau[0] = []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 10, false),
		domain.NewCard(domain.CardDesignSpade, 9, false),
	}
	for i := 1; i < domain.EightOffTableauCnt; i++ {
		tableau[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 2, false)}
	}
	e.SetTableau(tableau)
	var cells [domain.EightOffCellCnt]*domain.Card
	e.SetFreeCells(cells)

	result := p.HintOutput(e)

	assert.Contains(t, result, "フリーセル")
}

func TestEightOffCuiPresenterHintNil(t *testing.T) {
	p := new(EightOffCuiPresenter)
	e := domain.NewEightOff(domain.NewTrumpCards(0))
	e.Reset()
	e.SetPhase(domain.EightOffPhaseGameOver)

	result := p.HintOutput(e)

	assert.Contains(t, result, "ヒントはありません")
}

func TestEightOffCuiPresenterActionLogPlaying(t *testing.T) {
	p := new(EightOffCuiPresenter)
	e := domain.NewEightOff(domain.NewTrumpCards(0))
	e.Reset()
	e.SetPhase(domain.EightOffPhasePlaying)

	result := p.ActionLogOutput(e)

	assert.Contains(t, result, "棋譜はありません")
}

func TestEightOffCuiPresenterActionLogGameOver(t *testing.T) {
	p := new(EightOffCuiPresenter)
	e := domain.NewEightOff(domain.NewTrumpCards(0))
	e.Reset()

	var tableau [domain.EightOffTableauCnt][]*domain.Card
	tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
	e.SetTableau(tableau)
	e.SetPhase(domain.EightOffPhasePlaying)
	_ = e.MoveTableauToFoundation(0)
	e.SetPhase(domain.EightOffPhaseGameOver)

	result := p.ActionLogOutput(e)

	assert.Contains(t, result, "棋譜")
}
