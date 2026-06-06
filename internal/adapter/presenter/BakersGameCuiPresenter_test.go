//go:build test

package presenter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
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
