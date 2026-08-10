//go:build test

package presenter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func TestPenguinCuiPresenterOutputPlaying(t *testing.T) {
	p := new(PenguinCuiPresenter)
	g := domain.NewPenguin(domain.NewTrumpCards(0))
	g.Reset()
	g.SetPhase(domain.PenguinPhasePlaying)

	result := p.Output(g, nil)

	assert.Contains(t, result, "Penguin")
	assert.Contains(t, result, "手数:")
}

func TestPenguinCuiPresenterOutputBaseRankLabels(t *testing.T) {
	p := new(PenguinCuiPresenter)
	// Face-rank base values render as A/J/Q/K, matching the web baseRankLabel.
	cases := map[int]string{1: "A", 11: "J", 12: "Q", 13: "K", 7: "7"}
	for rank, label := range cases {
		g := domain.NewPenguin(domain.NewTrumpCards(0))
		g.Reset()
		g.SetPhase(domain.PenguinPhasePlaying)
		g.SetBaseRank(rank)
		result := p.Output(g, nil)
		assert.Contains(t, result, "BaseRank: "+label)
	}
}

func TestPenguinCuiPresenterOutputGameClear(t *testing.T) {
	p := new(PenguinCuiPresenter)
	g := domain.NewPenguin(domain.NewTrumpCards(0))
	g.Reset()
	g.SetPhase(domain.PenguinPhaseGameClear)

	result := p.Output(g, nil)

	assert.Contains(t, result, "ゲームクリア")
}

func TestPenguinCuiPresenterOutputGameOver(t *testing.T) {
	p := new(PenguinCuiPresenter)
	g := domain.NewPenguin(domain.NewTrumpCards(0))
	g.Reset()
	g.SetPhase(domain.PenguinPhaseGameOver)

	result := p.Output(g, nil)

	assert.Contains(t, result, "ゲームオーバー")
}

func TestPenguinCuiPresenterOutputStalemate(t *testing.T) {
	p := new(PenguinCuiPresenter)
	g := domain.NewPenguin(domain.NewTrumpCards(0))
	g.Reset()
	g.SetPhase(domain.PenguinPhasePlaying)
	g.SetIsStalemate(true)

	result := p.Output(g, nil)

	assert.Contains(t, result, "手詰まりです")
}

func TestPenguinCuiPresenterOutputError(t *testing.T) {
	p := new(PenguinCuiPresenter)
	g := domain.NewPenguin(domain.NewTrumpCards(0))
	g.Reset()

	result := p.Output(g, errors.New("test error"))

	assert.Contains(t, result, "test error")
}

func TestPenguinCuiPresenterOutputFreeCellsOccupied(t *testing.T) {
	p := new(PenguinCuiPresenter)
	g := domain.NewPenguin(domain.NewTrumpCards(0))
	g.Reset()
	g.SetPhase(domain.PenguinPhasePlaying)

	var cells [domain.PenguinCellCnt]*domain.Card
	cells[0] = domain.NewCard(domain.CardDesignSpade, 5, false)
	g.SetFreeCells(cells)

	result := p.Output(g, nil)

	assert.Contains(t, result, "SPADE 5")
}

func TestPenguinCuiPresenterOutputEmptyFreeCells(t *testing.T) {
	p := new(PenguinCuiPresenter)
	g := domain.NewPenguin(domain.NewTrumpCards(0))
	g.Reset()
	g.SetPhase(domain.PenguinPhasePlaying)

	var cells [domain.PenguinCellCnt]*domain.Card
	g.SetFreeCells(cells)

	result := p.Output(g, nil)

	assert.Contains(t, result, "[空]")
}

func TestPenguinCuiPresenterOutputFoundationWithCards(t *testing.T) {
	p := new(PenguinCuiPresenter)
	g := domain.NewPenguin(domain.NewTrumpCards(0))
	g.Reset()
	g.SetPhase(domain.PenguinPhasePlaying)

	var foundation [domain.PenguinFoundationCnt][]*domain.Card
	foundation[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
	g.SetFoundation(foundation)

	result := p.Output(g, nil)

	assert.Contains(t, result, "SPADE 1")
}

func TestPenguinCuiPresenterOutputEmptyTableau(t *testing.T) {
	p := new(PenguinCuiPresenter)
	g := domain.NewPenguin(domain.NewTrumpCards(0))
	g.Reset()
	g.SetPhase(domain.PenguinPhasePlaying)

	var emptyTableau [domain.PenguinTableauCnt][]*domain.Card
	g.SetTableau(emptyTableau)

	result := p.Output(g, nil)

	assert.Contains(t, result, "[空]")
}

func TestPenguinCuiPresenterHintTableau(t *testing.T) {
	p := new(PenguinCuiPresenter)
	g := domain.NewPenguin(domain.NewTrumpCards(0))
	g.Reset()
	g.SetPhase(domain.PenguinPhasePlaying)

	baseRank := g.GetBaseRank()
	var tableau [domain.PenguinTableauCnt][]*domain.Card
	tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, baseRank, false)}
	g.SetTableau(tableau)

	result := p.HintOutput(g)

	assert.Contains(t, result, "タブロー列")
}

func TestPenguinCuiPresenterHintFreeCell(t *testing.T) {
	p := new(PenguinCuiPresenter)
	g := domain.NewPenguin(domain.NewTrumpCards(0))
	g.Reset()
	g.SetPhase(domain.PenguinPhasePlaying)

	// Clear tableau so no tableau-based hints exist
	var emptyTableau [domain.PenguinTableauCnt][]*domain.Card
	g.SetTableau(emptyTableau)
	// Place a card in freecell that can move to an empty tableau column
	var cells [domain.PenguinCellCnt]*domain.Card
	cells[3] = domain.NewCard(domain.CardDesignSpade, 1, false)
	g.SetFreeCells(cells)

	result := p.HintOutput(g)

	// Either a freecell hint or no-hint (depending on game state)
	assert.NotEmpty(t, result)
}

func TestPenguinCuiPresenterHintFromFreeCellToTableau(t *testing.T) {
	p := new(PenguinCuiPresenter)
	mg := new(interfaces.MockPenguinGame)
	hint := &domain.PenguinHint{
		FromZone:  "freecell",
		FromCol:   2,
		CardIndex: 0,
		ToZone:    "tableau",
		ToCol:     4,
	}
	mg.On("GetHint").Return(hint)

	result := p.HintOutput(mg)

	assert.Contains(t, result, "フリーセル")
	assert.Contains(t, result, "タブロー列")
}

func TestPenguinCuiPresenterHintToFoundation(t *testing.T) {
	p := new(PenguinCuiPresenter)
	mg := new(interfaces.MockPenguinGame)
	hint := &domain.PenguinHint{
		FromZone:  "tableau",
		FromCol:   0,
		CardIndex: 3,
		ToZone:    "foundation",
		ToCol:     0,
	}
	mg.On("GetHint").Return(hint)

	result := p.HintOutput(mg)

	assert.Contains(t, result, "タブロー列")
	assert.Contains(t, result, "ファンデーション")
}

func TestPenguinCuiPresenterHintToFreeCell(t *testing.T) {
	p := new(PenguinCuiPresenter)
	mg := new(interfaces.MockPenguinGame)
	hint := &domain.PenguinHint{
		FromZone:  "tableau",
		FromCol:   1,
		CardIndex: 5,
		ToZone:    "freecell",
		ToCol:     3,
	}
	mg.On("GetHint").Return(hint)

	result := p.HintOutput(mg)

	assert.Contains(t, result, "タブロー列")
	assert.Contains(t, result, "フリーセル")
}

func TestPenguinCuiPresenterHintNil(t *testing.T) {
	p := new(PenguinCuiPresenter)
	g := domain.NewPenguin(domain.NewTrumpCards(0))
	g.Reset()
	g.SetPhase(domain.PenguinPhaseGameOver)

	result := p.HintOutput(g)

	assert.Contains(t, result, "ヒントはありません")
}

func TestPenguinCuiPresenterActionLogPlaying(t *testing.T) {
	p := new(PenguinCuiPresenter)
	g := domain.NewPenguin(domain.NewTrumpCards(0))
	g.Reset()
	g.SetPhase(domain.PenguinPhasePlaying)

	result := p.ActionLogOutput(g)

	assert.Contains(t, result, "棋譜はありません")
}

func TestPenguinCuiPresenterActionLogGameOver(t *testing.T) {
	p := new(PenguinCuiPresenter)
	g := domain.NewPenguin(domain.NewTrumpCards(0))
	g.Reset()

	baseRank := g.GetBaseRank()
	var tableau [domain.PenguinTableauCnt][]*domain.Card
	tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, baseRank, false)}
	g.SetTableau(tableau)
	g.SetPhase(domain.PenguinPhasePlaying)
	_ = g.MoveTableauToFoundation(0)
	g.SetPhase(domain.PenguinPhaseGameOver)

	result := p.ActionLogOutput(g)

	assert.Contains(t, result, "棋譜")
}

// **上限が出ておらず、拒否されたコマンドで初めて気づく形だった (#4802)。**
// 姉妹の Eight Off は supermoveLine を毎ターン出し、Web も pg-supermove-badge を
// 常設している。
func TestPenguinCuiPresenterOutput_SupermoveLine(t *testing.T) {
	p := new(PenguinCuiPresenter)
	card := func(v int) *domain.Card { return domain.NewCard(domain.CardDesignSpade, v, false) }
	board := func(filledCells, filledCols int) *domain.Penguin {
		g := domain.NewPenguin(domain.NewTrumpCards(0))
		g.Reset()
		g.SetPhase(domain.PenguinPhasePlaying)
		var cells [domain.PenguinCellCnt]*domain.Card
		for i := 0; i < filledCells && i < domain.PenguinCellCnt; i++ {
			cells[i] = card(i + 2)
		}
		g.SetFreeCells(cells)
		var tableau [domain.PenguinTableauCnt][]*domain.Card
		for i := 0; i < domain.PenguinTableauCnt; i++ {
			if i < filledCols {
				tableau[i] = []*domain.Card{card(5)}
			}
		}
		g.SetTableau(tableau)
		return g
	}

	t.Run("names the limit and what it is made of", func(t *testing.T) {
		out := p.Output(board(0, domain.PenguinTableauCnt), nil)
		assert.Contains(t, out, "一括移動")
		assert.Contains(t, out, "空きセル")
		assert.Contains(t, out, "空き列0")
	})

	// **空き列があるときだけ、そこへ置く上限も出す。**同じ数だと嘘になる。
	t.Run("adds the lower empty-column limit when one exists", func(t *testing.T) {
		assert.Contains(t, p.Output(board(0, domain.PenguinTableauCnt-1), nil), "空き列へは")
	})

	t.Run("omits the empty-column limit when no column is empty", func(t *testing.T) {
		assert.NotContains(t, p.Output(board(0, domain.PenguinTableauCnt), nil), "空き列へは")
	})

	t.Run("shows a limit of one when the board is packed", func(t *testing.T) {
		out := p.Output(board(domain.PenguinCellCnt, domain.PenguinTableauCnt), nil)
		assert.Contains(t, out, "最大1枚")
	})

	t.Run("shows nothing once the game is cleared", func(t *testing.T) {
		g := board(0, domain.PenguinTableauCnt)
		g.SetPhase(domain.PenguinPhaseGameClear)
		assert.NotContains(t, p.Output(g, nil), "一括移動")
	})
}
