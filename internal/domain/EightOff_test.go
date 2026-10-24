//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestEightOff() *EightOff {
	return NewEightOff(NewTrumpCards(0))
}

func setupPlayingEightOff() *EightOff {
	e := newTestEightOff()
	e.Reset()
	return e
}

func clearTableauEO(e *EightOff) {
	for i := 0; i < EightOffTableauCnt; i++ {
		e.tableau[i] = nil
	}
}

func clearFreeCellsEO(e *EightOff) {
	for i := 0; i < EightOffCellCnt; i++ {
		e.freeCells[i] = nil
	}
}

func TestEightOffLegalTargetsFrom(t *testing.T) {
	e := newTestEightOff()
	e.tableau[0] = []*Card{NewCard(CardDesignSpade, 13, false), NewCard(CardDesignSpade, 12, false)}
	e.tableau[1] = []*Card{NewCard(CardDesignSpade, 14, false)}
	tab, foundation := e.LegalTargetsFrom(0, 0)
	assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 7}, tab)
	assert.Empty(t, foundation)
	tab, foundation = e.LegalTargetsFrom(0, 1)
	assert.Empty(t, tab)
	assert.Empty(t, foundation)
	badTab, badFoundation := e.LegalTargetsFrom(-1, 0)
	assert.Empty(t, badTab)
	assert.Empty(t, badFoundation)
}

func TestEightOffFreeCellLegalTargets(t *testing.T) {
	e := newTestEightOff()
	e.freeCells[0] = NewCard(CardDesignSpade, 12, false)
	e.tableau[1] = []*Card{NewCard(CardDesignSpade, 13, false)}
	tab, foundation := e.FreeCellLegalTargets(0)
	assert.Equal(t, []int{1}, tab)
	assert.Empty(t, foundation)
	badTab, badFoundation := e.FreeCellLegalTargets(1)
	assert.Empty(t, badTab)
	assert.Empty(t, badFoundation)
}

func TestEightOffLegalTargetRules(t *testing.T) {
	t.Run("tableau placement and supermove capacity", func(t *testing.T) {
		e := newTestEightOff()
		clearTableauEO(e)
		clearFreeCellsEO(e)
		e.tableau[0] = []*Card{NewCard(CardDesignHeart, 12, false), NewCard(CardDesignHeart, 11, false)}
		e.tableau[1] = []*Card{NewCard(CardDesignHeart, 13, false)}
		for col := 2; col < EightOffTableauCnt; col++ {
			e.tableau[col] = []*Card{NewCard(CardDesignSpade, col, false)}
		}
		for cell := range EightOffCellCnt {
			e.freeCells[cell] = NewCard(CardDesignSpade, cell+1, false)
		}
		targets, _ := e.LegalTargetsFrom(0, 0)
		assert.Empty(t, targets, "a two-card stack exceeds capacity with full cells and no empty columns")
		e.freeCells[0] = nil
		targets, _ = e.LegalTargetsFrom(0, 0)
		assert.Equal(t, []int{1}, targets, "matching suit and descending rank are legal at capacity")
		clearTableauEO(e)
		clearFreeCellsEO(e)
		e.tableau[2] = []*Card{NewCard(CardDesignSpade, 13, false)}
		targets, _ = e.LegalTargetsFrom(2, 0)
		assert.Equal(t, []int{0, 1, 3, 4, 5, 6, 7}, targets, "only Kings can move to empty columns")
	})

	t.Run("foundation eligibility is top card only and suit-specific", func(t *testing.T) {
		e := newTestEightOff()
		clearTableauEO(e)
		clearFreeCellsEO(e)
		e.tableau[0] = []*Card{NewCard(CardDesignHeart, 2, false), NewCard(CardDesignHeart, 1, false)}
		_, foundation := e.LegalTargetsFrom(0, 0)
		assert.Empty(t, foundation, "a buried Ace cannot be swept")
		_, foundation = e.LegalTargetsFrom(0, 1)
		assert.Equal(t, []int{2}, foundation, "an Ace can start its empty suit pile")
		e.foundation[2] = []*Card{NewCard(CardDesignHeart, 1, false)}
		e.tableau[0][1] = NewCard(CardDesignHeart, 3, false)
		_, foundation = e.LegalTargetsFrom(0, 1)
		assert.Empty(t, foundation, "foundation cannot skip a rank")
		e.tableau[0][1] = NewCard(CardDesignSpade, 2, false)
		_, foundation = e.LegalTargetsFrom(0, 1)
		assert.Empty(t, foundation, "foundation suit must match")
	})

	t.Run("free cell placement and foundation", func(t *testing.T) {
		e := newTestEightOff()
		clearTableauEO(e)
		clearFreeCellsEO(e)
		e.tableau[0] = []*Card{NewCard(CardDesignHeart, 12, false)}
		e.freeCells[0] = NewCard(CardDesignHeart, 11, false)
		targets, foundation := e.FreeCellLegalTargets(0)
		assert.Equal(t, []int{0}, targets)
		assert.Empty(t, foundation)
		e.freeCells[0] = NewCard(CardDesignSpade, 13, false)
		targets, _ = e.FreeCellLegalTargets(0)
		assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 7}, targets)
		e.freeCells[0] = NewCard(CardDesignClover, 1, false)
		_, foundation = e.FreeCellLegalTargets(0)
		assert.Equal(t, []int{1}, foundation)
	})
}

// --- Reset tests ---

func TestEightOffReset(t *testing.T) {
	e := newTestEightOff()
	e.Reset()

	assert.Equal(t, EightOffPhasePlaying, e.GetPhase())
	assert.Equal(t, 0, e.GetMoveCount())

	// 各列に6枚ずつ = 48枚
	tableau := e.GetTableau()
	for i := 0; i < EightOffTableauCnt; i++ {
		assert.Equal(t, EightOffTableauColCards, len(tableau[i]))
	}

	// フリーセル0〜3にカードが入り、4〜7は空
	cells := e.GetFreeCells()
	for i := 0; i < 4; i++ {
		assert.NotNil(t, cells[i])
	}
	for i := 4; i < EightOffCellCnt; i++ {
		assert.Nil(t, cells[i])
	}

	// ファンデーションは全て空
	foundation := e.GetFoundation()
	for i := 0; i < EightOffFoundationCnt; i++ {
		assert.Empty(t, foundation[i])
	}

	// 合計52枚 (48 + 4)
	totalCards := 0
	for i := 0; i < EightOffTableauCnt; i++ {
		totalCards += len(tableau[i])
	}
	for i := 0; i < EightOffCellCnt; i++ {
		if cells[i] != nil {
			totalCards++
		}
	}
	assert.Equal(t, 52, totalCards)
}

func TestEightOffResetClearsHistory(t *testing.T) {
	e := setupPlayingEightOff()
	assert.Nil(t, e.GetActionLog())
	assert.False(t, e.CanUndo())
}

// --- MoveTableauToTableau tests (same-suit descending) ---

func TestEightOffMoveTableauToTableauSameSuit(t *testing.T) {
	e := setupPlayingEightOff()
	clearTableauEO(e)
	clearFreeCellsEO(e)

	// 同スート降順: ♠5 の上に ♠4
	e.tableau[0] = []*Card{makeCard(CardDesignSpade, 5)}
	e.tableau[1] = []*Card{makeCard(CardDesignSpade, 4)}

	err := e.MoveTableauToTableau(1, 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(e.tableau[0]))
	assert.Equal(t, 0, len(e.tableau[1]))
	assert.Equal(t, 1, e.GetMoveCount())
}

func TestEightOffMoveTableauToTableauRejectsDifferentSuit(t *testing.T) {
	e := setupPlayingEightOff()
	clearTableauEO(e)
	clearFreeCellsEO(e)

	// FreeCellでは合法な赤黒交互も、Eight Offでは異スートなので不可
	e.tableau[0] = []*Card{makeCard(CardDesignSpade, 13)}
	e.tableau[1] = []*Card{makeCard(CardDesignHeart, 12)}

	err := e.MoveTableauToTableau(1, 0, 0)
	assert.Error(t, err)
}

func TestEightOffMoveTableauToTableauSupermove(t *testing.T) {
	e := setupPlayingEightOff()
	clearTableauEO(e)
	clearFreeCellsEO(e)

	// ♠K の上に ♠Q→♠J の同スート連鎖を移動
	e.tableau[0] = []*Card{makeCard(CardDesignSpade, 13)}
	e.tableau[1] = []*Card{
		makeCard(CardDesignSpade, 12),
		makeCard(CardDesignSpade, 11),
	}

	err := e.MoveTableauToTableau(1, 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(e.tableau[0]))
}

func TestEightOffMoveTableauToTableauTopCardShortcut(t *testing.T) {
	e := setupPlayingEightOff()
	clearTableauEO(e)
	clearFreeCellsEO(e)

	e.tableau[0] = []*Card{makeCard(CardDesignSpade, 5)}
	e.tableau[1] = []*Card{makeCard(CardDesignSpade, 4)}

	err := e.MoveTableauToTableau(1, -1, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(e.tableau[0]))
}

func TestEightOffMoveTableauToTableauTopCardEmpty(t *testing.T) {
	e := setupPlayingEightOff()
	clearTableauEO(e)
	clearFreeCellsEO(e)
	err := e.MoveTableauToTableau(0, -1, 1)
	assert.Error(t, err)
}

func TestEightOffMoveTableauToTableauKingToEmpty(t *testing.T) {
	e := setupPlayingEightOff()
	clearTableauEO(e)
	clearFreeCellsEO(e)

	e.tableau[0] = []*Card{makeCard(CardDesignSpade, 13)}

	err := e.MoveTableauToTableau(0, 0, 1)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(e.tableau[0]))
	assert.Equal(t, 1, len(e.tableau[1]))
}

func TestEightOffMoveTableauToTableauNonKingToEmptyFails(t *testing.T) {
	e := setupPlayingEightOff()
	clearTableauEO(e)
	clearFreeCellsEO(e)
	// Eight Off固有: 空列には King 以外置けない
	e.tableau[0] = []*Card{makeCard(CardDesignSpade, 5)}

	err := e.MoveTableauToTableau(0, 0, 1)
	assert.Error(t, err)
}

func TestEightOffMoveTableauToTableauErrors(t *testing.T) {
	t.Run("not playing", func(t *testing.T) {
		e := setupPlayingEightOff()
		e.SetPhase(EightOffPhaseGameOver)
		err := e.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})
	t.Run("invalid from col negative", func(t *testing.T) {
		e := setupPlayingEightOff()
		err := e.MoveTableauToTableau(-1, 0, 1)
		assert.Error(t, err)
	})
	t.Run("invalid from col too large", func(t *testing.T) {
		e := setupPlayingEightOff()
		err := e.MoveTableauToTableau(EightOffTableauCnt, 0, 1)
		assert.Error(t, err)
	})
	t.Run("invalid to col negative", func(t *testing.T) {
		e := setupPlayingEightOff()
		err := e.MoveTableauToTableau(0, 0, -1)
		assert.Error(t, err)
	})
	t.Run("invalid to col too large", func(t *testing.T) {
		e := setupPlayingEightOff()
		err := e.MoveTableauToTableau(0, 0, EightOffTableauCnt)
		assert.Error(t, err)
	})
	t.Run("same column", func(t *testing.T) {
		e := setupPlayingEightOff()
		err := e.MoveTableauToTableau(0, 0, 0)
		assert.Error(t, err)
	})
	t.Run("invalid card index negative", func(t *testing.T) {
		e := setupPlayingEightOff()
		err := e.MoveTableauToTableau(0, -2, 1)
		assert.Error(t, err)
	})
	t.Run("invalid card index too large", func(t *testing.T) {
		e := setupPlayingEightOff()
		clearTableauEO(e)
		clearFreeCellsEO(e)
		e.tableau[0] = []*Card{makeCard(CardDesignSpade, 1)}
		err := e.MoveTableauToTableau(0, 5, 1)
		assert.Error(t, err)
	})
	t.Run("invalid sequence", func(t *testing.T) {
		e := setupPlayingEightOff()
		clearTableauEO(e)
		clearFreeCellsEO(e)
		// 異スートの連続は不正
		e.tableau[0] = []*Card{
			makeCard(CardDesignSpade, 10),
			makeCard(CardDesignHeart, 9),
		}
		e.tableau[1] = []*Card{makeCard(CardDesignSpade, 11)}
		err := e.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})
	t.Run("too many cards", func(t *testing.T) {
		e := setupPlayingEightOff()
		clearTableauEO(e)
		// すべてのフリーセルを埋める
		for i := 0; i < EightOffCellCnt; i++ {
			e.freeCells[i] = makeCard(CardDesignSpade, 1)
		}
		// すべての他列も埋める
		for i := 2; i < EightOffTableauCnt; i++ {
			e.tableau[i] = []*Card{makeCard(CardDesignSpade, 1)}
		}
		e.tableau[0] = []*Card{
			makeCard(CardDesignHeart, 5),
			makeCard(CardDesignHeart, 4),
			makeCard(CardDesignHeart, 3),
		}
		e.tableau[1] = []*Card{makeCard(CardDesignHeart, 6)}
		// max = (1+0) * 2^0 = 1, trying to move 3 cards
		err := e.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})
	t.Run("cannot place on tableau (different suit)", func(t *testing.T) {
		e := setupPlayingEightOff()
		clearTableauEO(e)
		clearFreeCellsEO(e)
		e.tableau[0] = []*Card{makeCard(CardDesignHeart, 5)}
		e.tableau[1] = []*Card{makeCard(CardDesignSpade, 6)}
		err := e.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})
	t.Run("non-king to empty fails", func(t *testing.T) {
		e := setupPlayingEightOff()
		clearTableauEO(e)
		clearFreeCellsEO(e)
		e.tableau[0] = []*Card{makeCard(CardDesignSpade, 5)}
		err := e.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})
}

// --- MoveTableauToFoundation tests ---

func TestEightOffMoveTableauToFoundation(t *testing.T) {
	e := setupPlayingEightOff()
	clearTableauEO(e)
	clearFreeCellsEO(e)

	e.tableau[0] = []*Card{makeCard(CardDesignSpade, 1)}
	err := e.MoveTableauToFoundation(0)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(e.tableau[0]))
	assert.Equal(t, 1, len(e.foundation[0]))
	log := e.GetActionLog()
	assert.Len(t, log, 1)
	assert.Equal(t, "eightoff.log.moveTableauToFoundation", log[0].DetailCode)
	assert.Equal(t, map[string]string{"col": "0"}, log[0].DetailParams)
}

func TestEightOffMoveTableauToFoundationSequence(t *testing.T) {
	e := setupPlayingEightOff()
	clearTableauEO(e)
	clearFreeCellsEO(e)

	e.foundation[0] = []*Card{makeCard(CardDesignSpade, 1)}
	e.tableau[0] = []*Card{makeCard(CardDesignSpade, 2)}
	err := e.MoveTableauToFoundation(0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(e.foundation[0]))
}

func TestEightOffMoveTableauToFoundationErrors(t *testing.T) {
	t.Run("not playing", func(t *testing.T) {
		e := setupPlayingEightOff()
		e.SetPhase(EightOffPhaseGameOver)
		err := e.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})
	t.Run("invalid column negative", func(t *testing.T) {
		e := setupPlayingEightOff()
		err := e.MoveTableauToFoundation(-1)
		assert.Error(t, err)
	})
	t.Run("invalid column too large", func(t *testing.T) {
		e := setupPlayingEightOff()
		err := e.MoveTableauToFoundation(EightOffTableauCnt)
		assert.Error(t, err)
	})
	t.Run("empty column", func(t *testing.T) {
		e := setupPlayingEightOff()
		clearTableauEO(e)
		clearFreeCellsEO(e)
		err := e.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})
	t.Run("cannot place on foundation", func(t *testing.T) {
		e := setupPlayingEightOff()
		clearTableauEO(e)
		clearFreeCellsEO(e)
		// 2 on empty foundation
		e.tableau[0] = []*Card{makeCard(CardDesignSpade, 2)}
		err := e.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})
}

// --- MoveTableauToFreeCell tests ---

func TestEightOffMoveTableauToFreeCell(t *testing.T) {
	e := setupPlayingEightOff()
	clearTableauEO(e)
	clearFreeCellsEO(e)

	e.tableau[0] = []*Card{makeCard(CardDesignSpade, 5)}
	err := e.MoveTableauToFreeCell(0, 4)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(e.tableau[0]))
	assert.NotNil(t, e.freeCells[4])
}

func TestEightOffMoveTableauToFreeCellErrors(t *testing.T) {
	t.Run("not playing", func(t *testing.T) {
		e := setupPlayingEightOff()
		e.SetPhase(EightOffPhaseGameOver)
		err := e.MoveTableauToFreeCell(0, 0)
		assert.Error(t, err)
	})
	t.Run("invalid col", func(t *testing.T) {
		e := setupPlayingEightOff()
		err := e.MoveTableauToFreeCell(-1, 0)
		assert.Error(t, err)
		err = e.MoveTableauToFreeCell(EightOffTableauCnt, 0)
		assert.Error(t, err)
	})
	t.Run("invalid cell", func(t *testing.T) {
		e := setupPlayingEightOff()
		err := e.MoveTableauToFreeCell(0, -1)
		assert.Error(t, err)
		err = e.MoveTableauToFreeCell(0, EightOffCellCnt)
		assert.Error(t, err)
	})
	t.Run("empty column", func(t *testing.T) {
		e := setupPlayingEightOff()
		clearTableauEO(e)
		err := e.MoveTableauToFreeCell(0, 4)
		assert.Error(t, err)
	})
	t.Run("cell occupied", func(t *testing.T) {
		e := setupPlayingEightOff()
		clearTableauEO(e)
		clearFreeCellsEO(e)
		e.tableau[0] = []*Card{makeCard(CardDesignSpade, 1)}
		e.freeCells[4] = makeCard(CardDesignHeart, 2)
		err := e.MoveTableauToFreeCell(0, 4)
		assert.Error(t, err)
	})
}

// --- MoveFreeCellToTableau tests ---

func TestEightOffMoveFreeCellToTableau(t *testing.T) {
	e := setupPlayingEightOff()
	clearTableauEO(e)
	clearFreeCellsEO(e)

	e.tableau[0] = []*Card{makeCard(CardDesignSpade, 5)}
	e.freeCells[0] = makeCard(CardDesignSpade, 4)

	err := e.MoveFreeCellToTableau(0, 0)
	assert.NoError(t, err)
	assert.Nil(t, e.freeCells[0])
	assert.Equal(t, 2, len(e.tableau[0]))
}

func TestEightOffMoveFreeCellToTableauKingToEmpty(t *testing.T) {
	e := setupPlayingEightOff()
	clearTableauEO(e)
	clearFreeCellsEO(e)

	e.freeCells[0] = makeCard(CardDesignSpade, 13)
	err := e.MoveFreeCellToTableau(0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(e.tableau[0]))
}

func TestEightOffMoveFreeCellToTableauNonKingEmptyFails(t *testing.T) {
	e := setupPlayingEightOff()
	clearTableauEO(e)
	clearFreeCellsEO(e)

	e.freeCells[0] = makeCard(CardDesignSpade, 5)
	err := e.MoveFreeCellToTableau(0, 0)
	assert.Error(t, err)
}

func TestEightOffMoveFreeCellToTableauErrors(t *testing.T) {
	t.Run("not playing", func(t *testing.T) {
		e := setupPlayingEightOff()
		e.SetPhase(EightOffPhaseGameOver)
		err := e.MoveFreeCellToTableau(0, 0)
		assert.Error(t, err)
	})
	t.Run("invalid cell", func(t *testing.T) {
		e := setupPlayingEightOff()
		err := e.MoveFreeCellToTableau(-1, 0)
		assert.Error(t, err)
		err = e.MoveFreeCellToTableau(EightOffCellCnt, 0)
		assert.Error(t, err)
	})
	t.Run("invalid col", func(t *testing.T) {
		e := setupPlayingEightOff()
		err := e.MoveFreeCellToTableau(0, -1)
		assert.Error(t, err)
		err = e.MoveFreeCellToTableau(0, EightOffTableauCnt)
		assert.Error(t, err)
	})
	t.Run("empty cell", func(t *testing.T) {
		e := setupPlayingEightOff()
		clearFreeCellsEO(e)
		err := e.MoveFreeCellToTableau(0, 0)
		assert.Error(t, err)
	})
}

// --- MoveFreeCellToFoundation tests ---

func TestEightOffMoveFreeCellToFoundation(t *testing.T) {
	e := setupPlayingEightOff()
	clearTableauEO(e)
	clearFreeCellsEO(e)

	e.freeCells[0] = makeCard(CardDesignSpade, 1)
	err := e.MoveFreeCellToFoundation(0)
	assert.NoError(t, err)
	assert.Nil(t, e.freeCells[0])
	assert.Equal(t, 1, len(e.foundation[0]))
}

func TestEightOffMoveFreeCellToFoundationErrors(t *testing.T) {
	t.Run("not playing", func(t *testing.T) {
		e := setupPlayingEightOff()
		e.SetPhase(EightOffPhaseGameOver)
		err := e.MoveFreeCellToFoundation(0)
		assert.Error(t, err)
	})
	t.Run("invalid cell", func(t *testing.T) {
		e := setupPlayingEightOff()
		err := e.MoveFreeCellToFoundation(-1)
		assert.Error(t, err)
		err = e.MoveFreeCellToFoundation(EightOffCellCnt)
		assert.Error(t, err)
	})
	t.Run("empty cell", func(t *testing.T) {
		e := setupPlayingEightOff()
		clearFreeCellsEO(e)
		err := e.MoveFreeCellToFoundation(0)
		assert.Error(t, err)
	})
	t.Run("cannot place on foundation", func(t *testing.T) {
		e := setupPlayingEightOff()
		clearFreeCellsEO(e)
		e.freeCells[0] = makeCard(CardDesignSpade, 5)
		err := e.MoveFreeCellToFoundation(0)
		assert.Error(t, err)
	})
}

// --- Game state ---

func TestEightOffGiveUp(t *testing.T) {
	e := setupPlayingEightOff()
	e.GiveUp()
	assert.Equal(t, EightOffPhaseGameOver, e.GetPhase())
	assert.True(t, e.GetGameEndFlag())

	// after game over GiveUp is no-op
	e.GiveUp()
	assert.Equal(t, EightOffPhaseGameOver, e.GetPhase())
}

func TestEightOffCheckGameClear(t *testing.T) {
	e := setupPlayingEightOff()
	clearTableauEO(e)
	clearFreeCellsEO(e)

	// Fill all foundations to K
	for d := 0; d < EightOffFoundationCnt; d++ {
		for v := 1; v <= CardValueMax; v++ {
			e.foundation[d] = append(e.foundation[d], makeCard(d+1, v))
		}
	}
	// 1 card on tableau that will go to foundation
	e.foundation[0] = e.foundation[0][:CardValueMax-1]
	e.tableau[0] = []*Card{makeCard(CardDesignSpade, 13)}

	err := e.MoveTableauToFoundation(0)
	assert.NoError(t, err)
	assert.Equal(t, EightOffPhaseGameClear, e.GetPhase())
}

// --- Hint tests ---

func TestEightOffHintTableauToFoundation(t *testing.T) {
	e := setupPlayingEightOff()
	clearTableauEO(e)
	clearFreeCellsEO(e)

	e.tableau[0] = []*Card{makeCard(CardDesignSpade, 1)}
	hint := e.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "tableau", hint.FromZone)
	assert.Equal(t, "foundation", hint.ToZone)
}

func TestEightOffHintFreeCellToFoundation(t *testing.T) {
	e := setupPlayingEightOff()
	clearTableauEO(e)
	clearFreeCellsEO(e)

	e.freeCells[0] = makeCard(CardDesignSpade, 1)
	hint := e.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "freecell", hint.FromZone)
	assert.Equal(t, "foundation", hint.ToZone)
}

func TestEightOffHintTableauToTableau(t *testing.T) {
	e := setupPlayingEightOff()
	clearTableauEO(e)
	clearFreeCellsEO(e)

	e.tableau[0] = []*Card{makeCard(CardDesignSpade, 5)}
	e.tableau[1] = []*Card{makeCard(CardDesignSpade, 4)}
	hint := e.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "tableau", hint.FromZone)
	assert.Equal(t, "tableau", hint.ToZone)
}

func TestEightOffHintFreeCellToTableau(t *testing.T) {
	e := setupPlayingEightOff()
	clearTableauEO(e)
	clearFreeCellsEO(e)

	e.tableau[0] = []*Card{makeCard(CardDesignSpade, 5)}
	e.freeCells[0] = makeCard(CardDesignSpade, 4)
	hint := e.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "freecell", hint.FromZone)
	assert.Equal(t, "tableau", hint.ToZone)
}

func TestEightOffHintTableauToFreeCell(t *testing.T) {
	e := setupPlayingEightOff()
	clearTableauEO(e)
	clearFreeCellsEO(e)

	// 単一カードしかなく、ファンデーション可は不可、Tableau→Tableau可も不可、
	// FreeCell→他は空、結果としてTableau→FreeCellがヒント
	e.tableau[0] = []*Card{makeCard(CardDesignSpade, 5)}
	hint := e.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "tableau", hint.FromZone)
	assert.Equal(t, "freecell", hint.ToZone)
}

func TestEightOffHintNoneWhenNotPlaying(t *testing.T) {
	e := setupPlayingEightOff()
	e.SetPhase(EightOffPhaseGameOver)
	assert.Nil(t, e.GetHint())
}

// --- AutoComplete ---

func TestEightOffAutoComplete(t *testing.T) {
	e := setupPlayingEightOff()
	clearTableauEO(e)
	clearFreeCellsEO(e)

	// Place an Ace on each foundation slot via tableau / freecell
	e.tableau[0] = []*Card{makeCard(CardDesignSpade, 1)}
	e.tableau[1] = []*Card{makeCard(CardDesignClover, 1)}
	e.freeCells[0] = makeCard(CardDesignHeart, 1)
	e.freeCells[5] = makeCard(CardDesignDiamond, 1)

	err := e.AutoComplete()
	assert.NoError(t, err)
	for i := 0; i < EightOffFoundationCnt; i++ {
		assert.Equal(t, 1, len(e.foundation[i]))
	}
}

func TestEightOffAutoCompleteNotPlaying(t *testing.T) {
	e := setupPlayingEightOff()
	e.SetPhase(EightOffPhaseGameOver)
	err := e.AutoComplete()
	assert.Error(t, err)
}

// --- Undo ---

func TestEightOffUndo(t *testing.T) {
	e := setupPlayingEightOff()
	clearTableauEO(e)
	clearFreeCellsEO(e)

	e.tableau[0] = []*Card{makeCard(CardDesignSpade, 5)}
	require := assert.New(t)
	require.NoError(e.MoveTableauToFreeCell(0, 4))
	require.True(e.CanUndo())
	require.NoError(e.Undo())
	require.False(e.CanUndo())
	require.Equal(0, e.GetMoveCount())
}

func TestEightOffUndoNoHistory(t *testing.T) {
	e := setupPlayingEightOff()
	e.history = nil
	err := e.Undo()
	assert.Error(t, err)
}

func TestEightOffUndoNotPlaying(t *testing.T) {
	e := setupPlayingEightOff()
	e.SetPhase(EightOffPhaseGameOver)
	err := e.Undo()
	assert.Error(t, err)
}

func TestEightOffUndoN(t *testing.T) {
	e := setupPlayingEightOff()
	clearTableauEO(e)
	clearFreeCellsEO(e)

	e.tableau[0] = []*Card{makeCard(CardDesignSpade, 5)}
	e.tableau[1] = []*Card{makeCard(CardDesignSpade, 4)}
	require := assert.New(t)
	require.NoError(e.MoveTableauToFreeCell(0, 4))
	require.NoError(e.MoveTableauToFreeCell(1, 5))
	require.NoError(e.UndoN(2))
	require.False(e.CanUndo())
}

func TestEightOffUndoNFailure(t *testing.T) {
	e := setupPlayingEightOff()
	// no history -- should fail on first step
	err := e.UndoN(1)
	assert.Error(t, err)
}

// --- UndoToEscape ---

func TestEightOffUndoToEscapeNotStalemate(t *testing.T) {
	e := setupPlayingEightOff()
	e.SetIsStalemate(false)
	assert.Equal(t, 0, e.UndoToEscape())
}

func TestEightOffUndoToEscapeAllStalemate(t *testing.T) {
	e := setupPlayingEightOff()
	e.SetIsStalemate(true)
	e.history = []*eightOffSnapshot{
		{isStalemate: true},
		{isStalemate: true},
	}
	assert.Equal(t, -1, e.UndoToEscape())
}

func TestEightOffUndoToEscapeFindsExit(t *testing.T) {
	e := setupPlayingEightOff()
	e.SetIsStalemate(true)
	e.history = []*eightOffSnapshot{
		{isStalemate: false},
		{isStalemate: true},
		{isStalemate: true},
	}
	// 3 entries, oldest is non-stalemate at index 0, so we undo 3 times
	assert.Equal(t, 3, e.UndoToEscape())
}

// --- Setters ---

func TestEightOffSetters(t *testing.T) {
	e := newTestEightOff()
	var tableau [EightOffTableauCnt][]*Card
	tableau[0] = []*Card{makeCard(CardDesignSpade, 1)}
	e.SetTableau(tableau)
	assert.Equal(t, 1, len(e.GetTableau()[0]))

	var cells [EightOffCellCnt]*Card
	cells[7] = makeCard(CardDesignHeart, 1)
	e.SetFreeCells(cells)
	assert.NotNil(t, e.GetFreeCells()[7])

	var fnd [EightOffFoundationCnt][]*Card
	fnd[2] = []*Card{makeCard(CardDesignHeart, 1)}
	e.SetFoundation(fnd)
	assert.Equal(t, 1, len(e.GetFoundation()[2]))
}

func TestEightOffHintSkipsWholeColumnToEmpty(t *testing.T) {
	for _, useful := range []bool{true, false} {
		e := setupPlayingEightOff()
		e.Reset()
		var tableau [EightOffTableauCnt][]*Card
		tableau[0] = []*Card{makeCard(CardDesignSpade, 13)}
		if useful {
			tableau[2] = []*Card{makeCard(CardDesignClover, 3), makeCard(CardDesignHeart, 13), makeCard(CardDesignHeart, 12)}
		}
		e.SetTableau(tableau)
		var cells [EightOffCellCnt]*Card
		for i := range cells {
			cells[i] = makeCard(CardDesignSpade, 9)
		}
		e.SetFreeCells(cells)
		var foundation [EightOffFoundationCnt][]*Card
		for i := range foundation {
			foundation[i] = []*Card{makeCard(i+1, 13)}
		}
		e.SetFoundation(foundation)
		hint := e.GetHint()
		if !useful {
			assert.Nil(t, hint)
			continue
		}
		if assert.NotNil(t, hint) {
			assert.Equal(t, 2, hint.FromCol)
			assert.Equal(t, 1, hint.CardIndex)
			assert.Equal(t, 1, hint.ToCol)
		}
	}
}
