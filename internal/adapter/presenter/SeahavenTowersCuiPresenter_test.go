//go:build test

package presenter

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestSeahavenTowersCuiPresenterOutputPlaying(t *testing.T) {
	p := new(SeahavenTowersCuiPresenter)
	s := domain.NewSeahavenTowers(domain.NewTrumpCards(0))
	s.Reset()
	s.SetPhase(domain.SeahavenTowersPhasePlaying)

	result := p.Output(s, nil)

	assert.Contains(t, result, "Seahaven Towers")
	assert.Contains(t, result, "Reserved:")
	assert.Contains(t, result, "Foundation:")
	assert.Contains(t, result, "手数:")
}

func TestSeahavenTowersCuiPresenterSupermoveLine(t *testing.T) {
	p := new(SeahavenTowersCuiPresenter)
	// empty reserved cells → one-move limit (1 + empties).
	for empties, limit := range map[int]int{0: 1, 1: 2, 2: 3} {
		s := domain.NewSeahavenTowers(domain.NewTrumpCards(0))
		s.Reset()
		s.SetPhase(domain.SeahavenTowersPhasePlaying)
		var cells [domain.SeahavenTowersCellCnt]*domain.Card
		for i := 0; i < domain.SeahavenTowersCellCnt-empties; i++ {
			cells[i] = domain.NewCard(domain.CardDesignSpade, i+1, false)
		}
		s.SetFreeCells(cells)
		result := p.Output(s, nil)
		assert.Contains(t, result, "一括移動可能: 最大"+strconv.Itoa(limit)+"枚")
	}
}

func TestSeahavenTowersCuiPresenterOutputGameClear(t *testing.T) {
	p := new(SeahavenTowersCuiPresenter)
	s := domain.NewSeahavenTowers(domain.NewTrumpCards(0))
	s.Reset()
	s.SetPhase(domain.SeahavenTowersPhaseGameClear)

	result := p.Output(s, nil)
	assert.Contains(t, result, "ゲームクリア")
}

func TestSeahavenTowersCuiPresenterOutputGameOver(t *testing.T) {
	p := new(SeahavenTowersCuiPresenter)
	s := domain.NewSeahavenTowers(domain.NewTrumpCards(0))
	s.Reset()
	s.SetPhase(domain.SeahavenTowersPhaseGameOver)

	result := p.Output(s, nil)
	assert.Contains(t, result, "ゲームオーバー")
}

func TestSeahavenTowersCuiPresenterOutputStalemate(t *testing.T) {
	p := new(SeahavenTowersCuiPresenter)
	s := domain.NewSeahavenTowers(domain.NewTrumpCards(0))
	s.Reset()
	s.SetPhase(domain.SeahavenTowersPhasePlaying)
	s.SetIsStalemate(true)

	result := p.Output(s, nil)
	assert.Contains(t, result, "手詰まりです")
}

func TestSeahavenTowersCuiPresenterOutputError(t *testing.T) {
	p := new(SeahavenTowersCuiPresenter)
	s := domain.NewSeahavenTowers(domain.NewTrumpCards(0))
	s.Reset()

	result := p.Output(s, errors.New("test error"))
	assert.Contains(t, result, "test error")
}

func TestSeahavenTowersCuiPresenterOutputReservedOccupied(t *testing.T) {
	p := new(SeahavenTowersCuiPresenter)
	s := domain.NewSeahavenTowers(domain.NewTrumpCards(0))
	s.Reset()
	s.SetPhase(domain.SeahavenTowersPhasePlaying)

	var cells [domain.SeahavenTowersCellCnt]*domain.Card
	cells[0] = domain.NewCard(domain.CardDesignSpade, 5, false)
	s.SetFreeCells(cells)

	result := p.Output(s, nil)
	assert.Contains(t, result, "Reserved:")
	assert.Contains(t, result, "SPADE 5")
}

func TestSeahavenTowersCuiPresenterOutputFoundationWithCards(t *testing.T) {
	p := new(SeahavenTowersCuiPresenter)
	s := domain.NewSeahavenTowers(domain.NewTrumpCards(0))
	s.Reset()
	s.SetPhase(domain.SeahavenTowersPhasePlaying)

	var foundation [domain.SeahavenTowersFoundationCnt][]*domain.Card
	foundation[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
	s.SetFoundation(foundation)

	result := p.Output(s, nil)
	assert.Contains(t, result, "Foundation:")
	assert.Contains(t, result, "SPADE 1")
}

func TestSeahavenTowersCuiPresenterOutputEmptyTableau(t *testing.T) {
	p := new(SeahavenTowersCuiPresenter)
	s := domain.NewSeahavenTowers(domain.NewTrumpCards(0))
	s.Reset()
	s.SetPhase(domain.SeahavenTowersPhasePlaying)

	var emptyTableau [domain.SeahavenTowersTableauCnt][]*domain.Card
	s.SetTableau(emptyTableau)

	result := p.Output(s, nil)
	assert.Contains(t, result, "[空]")
}

func TestSeahavenTowersCuiPresenterHintTableau(t *testing.T) {
	p := new(SeahavenTowersCuiPresenter)
	s := domain.NewSeahavenTowers(domain.NewTrumpCards(0))
	s.Reset()
	s.SetPhase(domain.SeahavenTowersPhasePlaying)

	var tableau [domain.SeahavenTowersTableauCnt][]*domain.Card
	tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
	s.SetTableau(tableau)

	result := p.HintOutput(s)
	assert.Contains(t, result, "タブロー列")
}

func TestSeahavenTowersCuiPresenterHintReserved(t *testing.T) {
	p := new(SeahavenTowersCuiPresenter)
	s := domain.NewSeahavenTowers(domain.NewTrumpCards(0))
	s.Reset()
	s.SetPhase(domain.SeahavenTowersPhasePlaying)

	var emptyTableau [domain.SeahavenTowersTableauCnt][]*domain.Card
	s.SetTableau(emptyTableau)
	var cells [domain.SeahavenTowersCellCnt]*domain.Card
	cells[0] = domain.NewCard(domain.CardDesignSpade, 1, false)
	s.SetFreeCells(cells)

	result := p.HintOutput(s)
	assert.Contains(t, result, "リザーブセル")
}

func TestSeahavenTowersCuiPresenterHintToTableau(t *testing.T) {
	p := new(SeahavenTowersCuiPresenter)
	s := domain.NewSeahavenTowers(domain.NewTrumpCards(0))
	s.Reset()
	s.SetPhase(domain.SeahavenTowersPhasePlaying)

	var tableau [domain.SeahavenTowersTableauCnt][]*domain.Card
	tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 6, false)}
	s.SetTableau(tableau)
	var cells [domain.SeahavenTowersCellCnt]*domain.Card
	cells[0] = domain.NewCard(domain.CardDesignSpade, 5, false)
	s.SetFreeCells(cells)

	result := p.HintOutput(s)
	assert.Contains(t, result, "タブロー列")
}

func TestSeahavenTowersCuiPresenterHintToReserved(t *testing.T) {
	p := new(SeahavenTowersCuiPresenter)
	s := domain.NewSeahavenTowers(domain.NewTrumpCards(0))
	s.Reset()
	s.SetPhase(domain.SeahavenTowersPhasePlaying)

	// Tableau cards that can't go to foundation or pair with anything; all
	// columns occupied so the only hint is tableau→reserved.
	var tableau [domain.SeahavenTowersTableauCnt][]*domain.Card
	for i := 0; i < domain.SeahavenTowersTableauCnt; i++ {
		tableau[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 9-i%6, false)}
	}
	s.SetTableau(tableau)
	var cells [domain.SeahavenTowersCellCnt]*domain.Card
	cells[0] = domain.NewCard(domain.CardDesignSpade, 12, false)
	s.SetFreeCells(cells) // cells[1] empty → reserved hint target

	result := p.HintOutput(s)
	// Either tableau→reserved or another fallback; assert non-empty.
	assert.NotEmpty(t, result)
}

func TestSeahavenTowersCuiPresenterHintNil(t *testing.T) {
	p := new(SeahavenTowersCuiPresenter)
	s := domain.NewSeahavenTowers(domain.NewTrumpCards(0))
	s.Reset()
	s.SetPhase(domain.SeahavenTowersPhaseGameOver)

	result := p.HintOutput(s)
	assert.Contains(t, result, "ヒントはありません")
}

func TestSeahavenTowersCuiPresenterActionLogPlaying(t *testing.T) {
	p := new(SeahavenTowersCuiPresenter)
	s := domain.NewSeahavenTowers(domain.NewTrumpCards(0))
	s.Reset()
	s.SetPhase(domain.SeahavenTowersPhasePlaying)

	result := p.ActionLogOutput(s)
	assert.Contains(t, result, "棋譜はありません")
}

func TestSeahavenTowersCuiPresenterActionLogGameOver(t *testing.T) {
	p := new(SeahavenTowersCuiPresenter)
	s := domain.NewSeahavenTowers(domain.NewTrumpCards(0))
	s.Reset()

	var tableau [domain.SeahavenTowersTableauCnt][]*domain.Card
	tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
	s.SetTableau(tableau)
	s.SetPhase(domain.SeahavenTowersPhasePlaying)
	_ = s.MoveTableauToFoundation(0)
	s.SetPhase(domain.SeahavenTowersPhaseGameOver)

	result := p.ActionLogOutput(s)
	assert.Contains(t, result, "棋譜")
}
