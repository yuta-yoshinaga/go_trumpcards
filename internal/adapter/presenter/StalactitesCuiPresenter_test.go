//go:build test

package presenter

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestStalactitesCuiPresenterOutputPlaying(t *testing.T) {
	p := new(StalactitesCuiPresenter)
	f := domain.NewStalactites(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.StalactitesPhasePlaying)

	result := p.Output(f, nil)

	assert.Contains(t, result, "Stalactites")
	assert.Contains(t, result, "セル:")
	assert.Contains(t, result, "Foundation:")
	assert.Contains(t, result, "手数:")
}

func TestStalactitesCuiPresenterOutputGameClear(t *testing.T) {
	p := new(StalactitesCuiPresenter)
	f := domain.NewStalactites(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.StalactitesPhaseGameClear)

	result := p.Output(f, nil)

	assert.Contains(t, result, "ゲームクリア")
}

func TestStalactitesCuiPresenterOutputGameOver(t *testing.T) {
	p := new(StalactitesCuiPresenter)
	f := domain.NewStalactites(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.StalactitesPhaseGameOver)

	result := p.Output(f, nil)

	assert.Contains(t, result, "ゲームオーバー")
}

func TestStalactitesCuiPresenterOutputStalemate(t *testing.T) {
	p := new(StalactitesCuiPresenter)
	f := domain.NewStalactites(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.StalactitesPhasePlaying)
	f.SetIsStalemate(true)

	result := p.Output(f, nil)

	assert.Contains(t, result, "手詰まりです")
}

func TestStalactitesCuiPresenterOutputError(t *testing.T) {
	p := new(StalactitesCuiPresenter)
	f := domain.NewStalactites(domain.NewTrumpCards(0))
	f.Reset()

	result := p.Output(f, errors.New("test error"))

	assert.Contains(t, result, "test error")
}

func TestStalactitesCuiPresenterOutputCellsOccupied(t *testing.T) {
	p := new(StalactitesCuiPresenter)
	f := domain.NewStalactites(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.StalactitesPhasePlaying)

	var cells [domain.StalactitesCellCnt]*domain.Card
	cells[0] = domain.NewCard(domain.CardDesignSpade, 5, false)
	f.SetCells(cells)

	result := p.Output(f, nil)

	assert.Contains(t, result, "セル:")
	assert.Contains(t, result, "SPADE 5")
}

func TestStalactitesCuiPresenterOutputFoundationWithCards(t *testing.T) {
	p := new(StalactitesCuiPresenter)
	f := domain.NewStalactites(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.StalactitesPhasePlaying)

	var foundation [domain.StalactitesFoundationCnt][]*domain.Card
	foundation[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, f.GetBaseRank(), false)}
	f.SetFoundation(foundation)

	result := p.Output(f, nil)

	assert.Contains(t, result, "Foundation:")
	// The card was built from the deal's base rank, so the expectation has to
	// be too -- asserting a literal "SPADE 1" passed only when the shuffle made
	// Ace the base rank, i.e. roughly one run in thirteen.
	assert.Contains(t, result, fmt.Sprintf("SPADE %d", f.GetBaseRank()))
}

func TestStalactitesCuiPresenterOutputEmptyTableau(t *testing.T) {
	p := new(StalactitesCuiPresenter)
	f := domain.NewStalactites(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.StalactitesPhasePlaying)

	var emptyTableau [domain.StalactitesTableauCnt][]*domain.Card
	f.SetTableau(emptyTableau)

	result := p.Output(f, nil)

	assert.Contains(t, result, "[空]")
}

func TestStalactitesCuiPresenterHintTableau(t *testing.T) {
	p := new(StalactitesCuiPresenter)
	f := domain.NewStalactites(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.StalactitesPhasePlaying)

	// Place an Ace on tableau so hint suggests moving to foundation
	var tableau [domain.StalactitesTableauCnt][]*domain.Card
	tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, f.GetBaseRank(), false)}
	f.SetTableau(tableau)

	result := p.HintOutput(f)

	assert.Contains(t, result, "タブロー列")
}

func TestStalactitesCuiPresenterHintStalactites(t *testing.T) {
	p := new(StalactitesCuiPresenter)
	f := domain.NewStalactites(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.StalactitesPhasePlaying)

	// Place an Ace in a free cell so hint suggests moving to foundation
	var emptyTableau [domain.StalactitesTableauCnt][]*domain.Card
	f.SetTableau(emptyTableau)
	var cells [domain.StalactitesCellCnt]*domain.Card
	cells[0] = domain.NewCard(domain.CardDesignSpade, f.GetBaseRank(), false)
	f.SetCells(cells)

	result := p.HintOutput(f)

	assert.Contains(t, result, "セル")
}

// stalactitesRanksAvoiding returns an adjacent (low, high) rank pair with
// neither equal to base, so neither card can open a foundation. Stalactites'
// base rank comes from the deal, so any fixture that hardcodes ranks is a coin
// flip on the shuffle.
func stalactitesRanksAvoiding(base int) (int, int) {
	for high := 3; high <= domain.CardValueMax; high++ {
		low := high - 1
		if low != base && high != base {
			return low, high
		}
	}
	return 2, 3
}

func TestStalactitesCuiPresenterHintToTableau(t *testing.T) {
	p := new(StalactitesCuiPresenter)
	f := domain.NewStalactites(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.StalactitesPhasePlaying)

	// A cell card that can only go onto the tableau, so the hint must name a
	// column.
	//
	// **The ranks must avoid the base rank**, which Reset takes from the deal.
	// Hardcoding 5 and 6 meant that whenever the deal made either of those the
	// base rank, the card became foundation-playable and the higher-priority
	// foundation hint won -- about one run in seven. This is the second flake of
	// this exact shape in this file.
	low, high := stalactitesRanksAvoiding(f.GetBaseRank())
	var tableau [domain.StalactitesTableauCnt][]*domain.Card
	tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, high, false)}
	f.SetTableau(tableau)
	var cells [domain.StalactitesCellCnt]*domain.Card
	cells[0] = domain.NewCard(domain.CardDesignHeart, low, false)
	f.SetCells(cells)

	result := p.HintOutput(f)

	// Should mention tableau as destination
	assert.Contains(t, result, "タブロー列")
}

func TestStalactitesCuiPresenterHintToStalactites(t *testing.T) {
	p := new(StalactitesCuiPresenter)
	f := domain.NewStalactites(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.StalactitesPhasePlaying)

	// Set up a board whose only legal destination is a cell: every column filled
	// (no empty column to fall back on) and no card able to reach a foundation.
	//
	// **The ranks must avoid the base rank**, which Reset picks from the deal.
	// Hardcoding 2/9/10 made this pass or fail on the shuffle -- when the base
	// rank happened to be 2, the filler cards became foundation-playable and the
	// hint pointed there instead.
	base := f.GetBaseRank()
	ranks := make([]int, 0, 3)
	for v := 2; len(ranks) < 3 && v <= domain.CardValueMax; v++ {
		if v != base {
			ranks = append(ranks, v)
		}
	}
	var tableau [domain.StalactitesTableauCnt][]*domain.Card
	// Two non-adjacent ranks, so they cannot stack on each other either.
	tableau[0] = []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, ranks[2], false),
		domain.NewCard(domain.CardDesignSpade, ranks[0], false),
	}
	for i := 1; i < domain.StalactitesTableauCnt; i++ {
		tableau[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, ranks[0], false)}
	}
	f.SetTableau(tableau)
	var cells [domain.StalactitesCellCnt]*domain.Card
	f.SetCells(cells)

	result := p.HintOutput(f)

	assert.Contains(t, result, "セル")
}

func TestStalactitesCuiPresenterHintNil(t *testing.T) {
	p := new(StalactitesCuiPresenter)
	f := domain.NewStalactites(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.StalactitesPhaseGameOver)

	result := p.HintOutput(f)

	assert.Contains(t, result, "ヒントはありません")
}

func TestStalactitesCuiPresenterActionLogPlaying(t *testing.T) {
	p := new(StalactitesCuiPresenter)
	f := domain.NewStalactites(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.StalactitesPhasePlaying)

	result := p.ActionLogOutput(f)

	assert.Contains(t, result, "棋譜はありません")
}

func TestStalactitesCuiPresenterActionLogGameOver(t *testing.T) {
	p := new(StalactitesCuiPresenter)
	f := domain.NewStalactites(domain.NewTrumpCards(0))
	f.Reset()

	// Make a move to generate action log
	var tableau [domain.StalactitesTableauCnt][]*domain.Card
	tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, f.GetBaseRank(), false)}
	f.SetTableau(tableau)
	f.SetPhase(domain.StalactitesPhasePlaying)
	_ = f.MoveTableauToFoundation(0)
	f.SetPhase(domain.StalactitesPhaseGameOver)

	result := p.ActionLogOutput(f)

	assert.Contains(t, result, "棋譜")
}

// **何枚まとめて動かせるかが CUI に出ていなかった (#4777)。**姉妹ゲームの
// Seahaven Towers は supermoveLine を出している。
func TestStalactitesCuiPresenterOutput_SupermoveLine(t *testing.T) {
	p := new(StalactitesCuiPresenter)
	card := func(v int) *domain.Card { return domain.NewCard(domain.CardDesignSpade, v, false) }
	board := func(filledCells, filledCols int) *domain.Stalactites {
		f := domain.NewStalactites(domain.NewTrumpCards(0))
		f.Reset()
		f.SetPhase(domain.StalactitesPhasePlaying)
		var cells [domain.StalactitesCellCnt]*domain.Card
		for i := 0; i < filledCells && i < domain.StalactitesCellCnt; i++ {
			cells[i] = card(i + 2)
		}
		f.SetCells(cells)
		var tableau [domain.StalactitesTableauCnt][]*domain.Card
		for i := 0; i < domain.StalactitesTableauCnt; i++ {
			if i < filledCols {
				tableau[i] = []*domain.Card{card(5)}
			}
		}
		f.SetTableau(tableau)
		return f
	}

	t.Run("names the limit and what it is made of", func(t *testing.T) {
		out := p.Output(board(0, domain.StalactitesTableauCnt), nil)
		assert.Contains(t, out, "最大5枚")
		assert.Contains(t, out, "空きセル4")
		assert.Contains(t, out, "空き列0")
	})

	// **空き列があるときだけ、そこへ置く上限も出す。**同じ数だと嘘になる。
	t.Run("adds the lower empty-column limit when one exists", func(t *testing.T) {
		out := p.Output(board(0, domain.StalactitesTableauCnt-1), nil)
		assert.Contains(t, out, "最大10枚")
		assert.Contains(t, out, "空き列へは5枚")
	})

	t.Run("omits the empty-column limit when no column is empty", func(t *testing.T) {
		out := p.Output(board(0, domain.StalactitesTableauCnt), nil)
		assert.NotContains(t, out, "空き列へは")
	})

	t.Run("shows a limit of one when the board is packed", func(t *testing.T) {
		out := p.Output(board(domain.StalactitesCellCnt, domain.StalactitesTableauCnt), nil)
		assert.Contains(t, out, "最大1枚")
	})

	t.Run("shows nothing once the game is cleared", func(t *testing.T) {
		f := board(0, domain.StalactitesTableauCnt)
		f.SetPhase(domain.StalactitesPhaseGameClear)
		assert.NotContains(t, p.Output(f, nil), "一括移動")
	})
}
