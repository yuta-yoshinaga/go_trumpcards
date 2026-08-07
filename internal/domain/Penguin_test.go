//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Helpers ---

func newTestPenguin() *Penguin {
	return NewPenguin(NewTrumpCards(0))
}

func setupPlayingPenguin() *Penguin {
	p := newTestPenguin()
	p.Reset()
	return p
}

func clearTableauPG(p *Penguin) {
	for i := 0; i < PenguinTableauCnt; i++ {
		p.tableau[i] = nil
	}
}

func clearFreeCellsPG(p *Penguin) {
	for i := 0; i < PenguinCellCnt; i++ {
		p.freeCells[i] = nil
	}
}

// --- Reset tests ---

func TestPenguinReset(t *testing.T) {
	p := newTestPenguin()
	p.Reset()

	assert.Equal(t, PenguinPhasePlaying, p.GetPhase())
	assert.Equal(t, 0, p.GetMoveCount())
	assert.False(t, p.GetGameEndFlag())

	tableau := p.GetTableau()
	for i := 0; i < PenguinTableauCnt; i++ {
		assert.Equal(t, PenguinTableauColCards, len(tableau[i]), "col %d should have %d cards", i, PenguinTableauColCards)
	}

	cells := p.GetFreeCells()
	occupiedCells := 0
	for i := 0; i < PenguinCellCnt; i++ {
		if cells[i] != nil {
			occupiedCells++
		}
	}
	assert.Equal(t, 3, occupiedCells, "exactly 3 free cells should be occupied after Reset")

	foundation := p.GetFoundation()
	for i := 0; i < PenguinFoundationCnt; i++ {
		assert.Empty(t, foundation[i])
	}

	// Total cards = 49 in tableau + 3 in free cells = 52
	totalCards := 0
	for i := 0; i < PenguinTableauCnt; i++ {
		totalCards += len(tableau[i])
	}
	for i := 0; i < PenguinCellCnt; i++ {
		if cells[i] != nil {
			totalCards++
		}
	}
	assert.Equal(t, 52, totalCards)
}

func TestPenguinResetFreeCellsShareBaseRank(t *testing.T) {
	p := newTestPenguin()
	p.Reset()

	baseRank := p.GetBaseRank()
	assert.True(t, baseRank >= 1 && baseRank <= 13, "baseRank should be 1-13")

	cells := p.GetFreeCells()
	for i := 0; i < PenguinCellCnt; i++ {
		if cells[i] != nil {
			assert.Equal(t, baseRank, cells[i].GetValue(), "free cell card should share baseRank")
		}
	}
}

func TestPenguinResetClearsHistory(t *testing.T) {
	p := newTestPenguin()
	p.Reset()
	assert.Nil(t, p.GetActionLog())
	assert.False(t, p.CanUndo())
}

func TestPenguinResetClearsPhaseAndMoveCount(t *testing.T) {
	p := setupPlayingPenguin()
	p.SetPhase(PenguinPhaseGameOver)
	p.Reset()
	assert.Equal(t, PenguinPhasePlaying, p.GetPhase())
	assert.Equal(t, 0, p.GetMoveCount())
}

// --- MoveTableauToTableau tests ---

func TestPenguinMoveTableauToTableauSameSuitDescending(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	p.SetBaseRank(5)

	// ♠7 on top, move ♠6 onto it
	p.tableau[0] = []*Card{makeCard(CardDesignSpade, 7)}
	p.tableau[1] = []*Card{makeCard(CardDesignSpade, 6)}

	err := p.MoveTableauToTableau(1, -1, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(p.tableau[0]))
	assert.Equal(t, 0, len(p.tableau[1]))
	assert.Equal(t, 1, p.GetMoveCount())
}

func TestPenguinMoveTableauToTableauRejectsDifferentSuit(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	p.SetBaseRank(5)

	p.tableau[0] = []*Card{makeCard(CardDesignSpade, 7)}
	p.tableau[1] = []*Card{makeCard(CardDesignHeart, 6)}

	err := p.MoveTableauToTableau(1, -1, 0)
	assert.Error(t, err)
}

func TestPenguinMoveTableauToTableauRejectsWrongRank(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	p.SetBaseRank(5)

	p.tableau[0] = []*Card{makeCard(CardDesignSpade, 7)}
	p.tableau[1] = []*Card{makeCard(CardDesignSpade, 5)} // not prevRank(7)=6

	err := p.MoveTableauToTableau(1, -1, 0)
	assert.Error(t, err)
}

func TestPenguinMoveTableauToTableauNotPlayingPhase(t *testing.T) {
	p := setupPlayingPenguin()
	p.SetPhase(PenguinPhaseGameOver)
	err := p.MoveTableauToTableau(0, 0, 1)
	assert.Error(t, err)
}

func TestPenguinMoveTableauToTableauInvalidColumns(t *testing.T) {
	p := setupPlayingPenguin()
	t.Run("from col negative", func(t *testing.T) {
		err := p.MoveTableauToTableau(-1, 0, 1)
		assert.Error(t, err)
	})
	t.Run("from col too large", func(t *testing.T) {
		err := p.MoveTableauToTableau(PenguinTableauCnt, 0, 1)
		assert.Error(t, err)
	})
	t.Run("to col negative", func(t *testing.T) {
		err := p.MoveTableauToTableau(0, 0, -1)
		assert.Error(t, err)
	})
	t.Run("to col too large", func(t *testing.T) {
		err := p.MoveTableauToTableau(0, 0, PenguinTableauCnt)
		assert.Error(t, err)
	})
	t.Run("same column", func(t *testing.T) {
		err := p.MoveTableauToTableau(0, 0, 0)
		assert.Error(t, err)
	})
}

func TestPenguinMoveTableauToTableauInvalidCardIndex(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	p.tableau[0] = []*Card{makeCard(CardDesignSpade, 5)}

	t.Run("index too negative", func(t *testing.T) {
		err := p.MoveTableauToTableau(0, -2, 1)
		assert.Error(t, err)
	})
	t.Run("index too large", func(t *testing.T) {
		err := p.MoveTableauToTableau(0, 5, 1)
		assert.Error(t, err)
	})
}

func TestPenguinMoveTableauToTableauEmptyFromColWithNeg1(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	// col 0 is empty; using -1 shortcut should fail (index becomes -1 → error)
	err := p.MoveTableauToTableau(0, -1, 1)
	assert.Error(t, err)
}

func TestPenguinMoveTableauToTableauSupermoveMultipleCards(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	p.SetBaseRank(5)

	// enough empty free cells and columns to allow moving 2 cards
	p.tableau[0] = []*Card{makeCard(CardDesignHeart, 9)}
	p.tableau[1] = []*Card{
		makeCard(CardDesignHeart, 8),
		makeCard(CardDesignHeart, 7),
	}
	// cols 2-6 and cells 0-6 all empty → max = (1+7)*2^5 = way more than 2

	err := p.MoveTableauToTableau(1, 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(p.tableau[0]))
}

func TestPenguinMoveTableauToTableauSupermoveTooManyCardsRejected(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)

	// Fill all free cells
	for i := 0; i < PenguinCellCnt; i++ {
		p.freeCells[i] = makeCard(CardDesignSpade, 1)
	}
	// Fill all other tableau columns
	for i := 2; i < PenguinTableauCnt; i++ {
		p.tableau[i] = []*Card{makeCard(CardDesignSpade, 1)}
	}
	// max = (1+0) * 2^0 = 1; trying to move 3 cards
	p.tableau[0] = []*Card{
		makeCard(CardDesignClover, 8),
		makeCard(CardDesignClover, 7),
		makeCard(CardDesignClover, 6),
	}
	p.tableau[1] = []*Card{makeCard(CardDesignClover, 9)}

	err := p.MoveTableauToTableau(0, 0, 1)
	assert.Error(t, err)
}

func TestPenguinMoveTableauToTableauEmptyColAcceptsPrevBaseRank(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	p.SetBaseRank(5)

	// prevRank(5) = 4
	p.tableau[0] = []*Card{makeCard(CardDesignDiamond, 4)}

	err := p.MoveTableauToTableau(0, -1, 1)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(p.tableau[1]))
}

func TestPenguinMoveTableauToTableauEmptyColRejectsNonPrevBaseRank(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	p.SetBaseRank(5)

	// prevRank(5) = 4; place a 7 which is not prevRank(5)
	p.tableau[0] = []*Card{makeCard(CardDesignSpade, 7)}

	err := p.MoveTableauToTableau(0, -1, 1)
	assert.Error(t, err)
}

func TestPenguinMoveTableauToTableauWraparoundAceOn2(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	p.SetBaseRank(3)

	// prevRank(2) = 1 (A); place ♠A on ♠2
	p.tableau[0] = []*Card{makeCard(CardDesignSpade, 2)}
	p.tableau[1] = []*Card{makeCard(CardDesignSpade, 1)}

	err := p.MoveTableauToTableau(1, -1, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(p.tableau[0]))
}

func TestPenguinMoveTableauToTableauInvalidSequence(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	p.SetBaseRank(5)

	// Two cards that don't form a valid sequence (different suits)
	p.tableau[0] = []*Card{
		makeCard(CardDesignSpade, 8),
		makeCard(CardDesignHeart, 7),
	}
	p.tableau[1] = []*Card{makeCard(CardDesignSpade, 9)}

	err := p.MoveTableauToTableau(0, 0, 1)
	assert.Error(t, err)
}

// --- MoveTableauToFoundation tests ---

func TestPenguinMoveTableauToFoundationBaseRankToEmpty(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	p.SetBaseRank(7)

	// baseRank=7, put ♠7 on empty foundation
	p.tableau[0] = []*Card{makeCard(CardDesignSpade, 7)}
	err := p.MoveTableauToFoundation(0)
	assert.NoError(t, err)
	assert.Empty(t, p.tableau[0])
	assert.Equal(t, 1, len(p.foundation[CardDesignSpade-1]))
}

func TestPenguinMoveTableauToFoundationNextRankOnNonEmpty(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	p.SetBaseRank(7)

	// nextRank(7) = 8
	p.foundation[CardDesignSpade-1] = []*Card{makeCard(CardDesignSpade, 7)}
	p.tableau[0] = []*Card{makeCard(CardDesignSpade, 8)}

	err := p.MoveTableauToFoundation(0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(p.foundation[CardDesignSpade-1]))
}

func TestPenguinMoveTableauToFoundationWrongSuitRejected(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	p.SetBaseRank(7)

	// ♠7 already on spade foundation; try ♥7 on spade foundation — wrong
	p.foundation[CardDesignSpade-1] = []*Card{makeCard(CardDesignSpade, 7)}
	p.tableau[0] = []*Card{makeCard(CardDesignHeart, 8)} // different suit

	err := p.MoveTableauToFoundation(0)
	assert.Error(t, err)
}

func TestPenguinMoveTableauToFoundationWrongRankRejected(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	p.SetBaseRank(7)

	// empty foundation, place rank 6 (not baseRank=7)
	p.tableau[0] = []*Card{makeCard(CardDesignSpade, 6)}
	err := p.MoveTableauToFoundation(0)
	assert.Error(t, err)
}

func TestPenguinMoveTableauToFoundationEmptyColumnRejected(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)

	err := p.MoveTableauToFoundation(0)
	assert.Error(t, err)
}

func TestPenguinMoveTableauToFoundationNotPlayingRejected(t *testing.T) {
	p := setupPlayingPenguin()
	p.SetPhase(PenguinPhaseGameOver)
	err := p.MoveTableauToFoundation(0)
	assert.Error(t, err)
}

func TestPenguinMoveTableauToFoundationInvalidColumn(t *testing.T) {
	p := setupPlayingPenguin()
	t.Run("negative", func(t *testing.T) {
		err := p.MoveTableauToFoundation(-1)
		assert.Error(t, err)
	})
	t.Run("too large", func(t *testing.T) {
		err := p.MoveTableauToFoundation(PenguinTableauCnt)
		assert.Error(t, err)
	})
}

// --- MoveTableauToFreeCell tests ---

func TestPenguinMoveTableauToFreeCell(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)

	p.tableau[0] = []*Card{makeCard(CardDesignSpade, 5)}
	err := p.MoveTableauToFreeCell(0, 3)
	assert.NoError(t, err)
	assert.Empty(t, p.tableau[0])
	assert.NotNil(t, p.freeCells[3])
	assert.Equal(t, 1, p.GetMoveCount())
}

func TestPenguinMoveTableauToFreeCellOccupiedCellRejected(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)

	p.tableau[0] = []*Card{makeCard(CardDesignSpade, 5)}
	p.freeCells[3] = makeCard(CardDesignHeart, 2)

	err := p.MoveTableauToFreeCell(0, 3)
	assert.Error(t, err)
}

func TestPenguinMoveTableauToFreeCellEmptyColumnRejected(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)

	err := p.MoveTableauToFreeCell(0, 3)
	assert.Error(t, err)
}

func TestPenguinMoveTableauToFreeCellInvalidColumnAndCell(t *testing.T) {
	p := setupPlayingPenguin()
	t.Run("invalid col negative", func(t *testing.T) {
		err := p.MoveTableauToFreeCell(-1, 0)
		assert.Error(t, err)
	})
	t.Run("invalid col too large", func(t *testing.T) {
		err := p.MoveTableauToFreeCell(PenguinTableauCnt, 0)
		assert.Error(t, err)
	})
	t.Run("invalid cell negative", func(t *testing.T) {
		err := p.MoveTableauToFreeCell(0, -1)
		assert.Error(t, err)
	})
	t.Run("invalid cell too large", func(t *testing.T) {
		err := p.MoveTableauToFreeCell(0, PenguinCellCnt)
		assert.Error(t, err)
	})
}

func TestPenguinMoveTableauToFreeCellNotPlayingRejected(t *testing.T) {
	p := setupPlayingPenguin()
	p.SetPhase(PenguinPhaseGameOver)
	err := p.MoveTableauToFreeCell(0, 0)
	assert.Error(t, err)
}

// --- MoveFreeCellToTableau tests ---

func TestPenguinMoveFreeCellToTableau(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	p.SetBaseRank(5)

	p.tableau[0] = []*Card{makeCard(CardDesignClover, 9)}
	p.freeCells[0] = makeCard(CardDesignClover, 8)

	err := p.MoveFreeCellToTableau(0, 0)
	assert.NoError(t, err)
	assert.Nil(t, p.freeCells[0])
	assert.Equal(t, 2, len(p.tableau[0]))
	assert.Equal(t, 1, p.GetMoveCount())
}

func TestPenguinMoveFreeCellToTableauEmptyColCorrectRank(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	p.SetBaseRank(6)

	// prevRank(6) = 5; free cell has rank 5 → OK on empty col
	p.freeCells[0] = makeCard(CardDesignDiamond, 5)

	err := p.MoveFreeCellToTableau(0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(p.tableau[0]))
}

func TestPenguinMoveFreeCellToTableauEmptyColWrongRankRejected(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	p.SetBaseRank(6)

	// prevRank(6) = 5; free cell has rank 7 → rejected
	p.freeCells[0] = makeCard(CardDesignDiamond, 7)

	err := p.MoveFreeCellToTableau(0, 0)
	assert.Error(t, err)
}

func TestPenguinMoveFreeCellToTableauEmptyCellRejected(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)

	err := p.MoveFreeCellToTableau(0, 0)
	assert.Error(t, err)
}

func TestPenguinMoveFreeCellToTableauNotPlayingRejected(t *testing.T) {
	p := setupPlayingPenguin()
	p.SetPhase(PenguinPhaseGameOver)
	err := p.MoveFreeCellToTableau(0, 0)
	assert.Error(t, err)
}

func TestPenguinMoveFreeCellToTableauInvalidCellAndCol(t *testing.T) {
	p := setupPlayingPenguin()
	t.Run("cell negative", func(t *testing.T) {
		err := p.MoveFreeCellToTableau(-1, 0)
		assert.Error(t, err)
	})
	t.Run("cell too large", func(t *testing.T) {
		err := p.MoveFreeCellToTableau(PenguinCellCnt, 0)
		assert.Error(t, err)
	})
	t.Run("col negative", func(t *testing.T) {
		err := p.MoveFreeCellToTableau(0, -1)
		assert.Error(t, err)
	})
	t.Run("col too large", func(t *testing.T) {
		err := p.MoveFreeCellToTableau(0, PenguinTableauCnt)
		assert.Error(t, err)
	})
}

// --- MoveFreeCellToFoundation tests ---

func TestPenguinMoveFreeCellToFoundation(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	p.SetBaseRank(3)

	p.freeCells[0] = makeCard(CardDesignHeart, 3)
	err := p.MoveFreeCellToFoundation(0)
	assert.NoError(t, err)
	assert.Nil(t, p.freeCells[0])
	assert.Equal(t, 1, len(p.foundation[CardDesignHeart-1]))
	assert.Equal(t, 1, p.GetMoveCount())
}

func TestPenguinMoveFreeCellToFoundationEmptyCellRejected(t *testing.T) {
	p := setupPlayingPenguin()
	clearFreeCellsPG(p)

	err := p.MoveFreeCellToFoundation(0)
	assert.Error(t, err)
}

func TestPenguinMoveFreeCellToFoundationNotPlayingRejected(t *testing.T) {
	p := setupPlayingPenguin()
	p.SetPhase(PenguinPhaseGameOver)
	err := p.MoveFreeCellToFoundation(0)
	assert.Error(t, err)
}

func TestPenguinMoveFreeCellToFoundationInvalidCell(t *testing.T) {
	p := setupPlayingPenguin()
	t.Run("negative", func(t *testing.T) {
		err := p.MoveFreeCellToFoundation(-1)
		assert.Error(t, err)
	})
	t.Run("too large", func(t *testing.T) {
		err := p.MoveFreeCellToFoundation(PenguinCellCnt)
		assert.Error(t, err)
	})
}

func TestPenguinMoveFreeCellToFoundationWrongRankRejected(t *testing.T) {
	p := setupPlayingPenguin()
	clearFreeCellsPG(p)
	p.SetBaseRank(3)

	// empty foundation, place rank 2 (not baseRank=3)
	p.freeCells[0] = makeCard(CardDesignSpade, 2)
	err := p.MoveFreeCellToFoundation(0)
	assert.Error(t, err)
}

// --- GiveUp tests ---

func TestPenguinGiveUpSetsGameOver(t *testing.T) {
	p := setupPlayingPenguin()
	p.GiveUp()
	assert.Equal(t, PenguinPhaseGameOver, p.GetPhase())
	assert.True(t, p.GetGameEndFlag())
}

func TestPenguinGiveUpNoOpWhenNotPlaying(t *testing.T) {
	p := setupPlayingPenguin()
	p.GiveUp()
	assert.Equal(t, PenguinPhaseGameOver, p.GetPhase())
	// Call again — should stay GameOver (not panic or change phase)
	p.GiveUp()
	assert.Equal(t, PenguinPhaseGameOver, p.GetPhase())
}

func TestPenguinGiveUpNoOpWhenGameClear(t *testing.T) {
	p := setupPlayingPenguin()
	p.SetPhase(PenguinPhaseGameClear)
	p.GiveUp()
	assert.Equal(t, PenguinPhaseGameClear, p.GetPhase())
}

// --- Undo / CanUndo / UndoN / UndoToEscape tests ---

func TestPenguinUndoAfterMove(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	p.SetBaseRank(5)

	p.tableau[0] = []*Card{makeCard(CardDesignSpade, 7)}
	require.NoError(t, p.MoveTableauToFreeCell(0, 3))
	require.True(t, p.CanUndo())
	require.NoError(t, p.Undo())
	assert.False(t, p.CanUndo())
	assert.Equal(t, 0, p.GetMoveCount())
	assert.Equal(t, 1, len(p.tableau[0]))
	assert.Nil(t, p.freeCells[3])
}

func TestPenguinUndoNoHistory(t *testing.T) {
	p := setupPlayingPenguin()
	p.history = nil
	err := p.Undo()
	assert.Error(t, err)
}

func TestPenguinUndoNotPlaying(t *testing.T) {
	p := setupPlayingPenguin()
	p.SetPhase(PenguinPhaseGameOver)
	err := p.Undo()
	assert.Error(t, err)
}

func TestPenguinCanUndoFalseWhenNotPlaying(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	p.tableau[0] = []*Card{makeCard(CardDesignSpade, 7)}
	require.NoError(t, p.MoveTableauToFreeCell(0, 3))
	p.SetPhase(PenguinPhaseGameOver)
	assert.False(t, p.CanUndo())
}

func TestPenguinUndoN(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	p.SetBaseRank(5)

	p.tableau[0] = []*Card{makeCard(CardDesignSpade, 7)}
	p.tableau[1] = []*Card{makeCard(CardDesignHeart, 8)}
	require.NoError(t, p.MoveTableauToFreeCell(0, 3))
	require.NoError(t, p.MoveTableauToFreeCell(1, 4))
	require.NoError(t, p.UndoN(2))
	assert.False(t, p.CanUndo())
	assert.Equal(t, 0, p.GetMoveCount())
}

func TestPenguinUndoNFailure(t *testing.T) {
	p := setupPlayingPenguin()
	err := p.UndoN(3)
	assert.Error(t, err)
}

func TestPenguinUndoToEscapeNotStalemate(t *testing.T) {
	p := setupPlayingPenguin()
	p.SetIsStalemate(false)
	assert.Equal(t, 0, p.UndoToEscape())
}

func TestPenguinUndoToEscapeAllStalemate(t *testing.T) {
	p := setupPlayingPenguin()
	p.SetIsStalemate(true)
	p.history = []*penguinSnapshot{
		{isStalemate: true},
		{isStalemate: true},
	}
	assert.Equal(t, -1, p.UndoToEscape())
}

func TestPenguinUndoToEscapeFindsExit(t *testing.T) {
	p := setupPlayingPenguin()
	p.SetIsStalemate(true)
	p.history = []*penguinSnapshot{
		{isStalemate: false},
		{isStalemate: true},
		{isStalemate: true},
	}
	// oldest non-stalemate at index 0 → need to undo 3 steps
	assert.Equal(t, 3, p.UndoToEscape())
}

// --- AutoComplete tests ---

func TestPenguinAutoComplete(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	p.SetBaseRank(1)

	// Place Aces on tableau/freecell to move to foundations
	p.tableau[0] = []*Card{makeCard(CardDesignSpade, 1)}
	p.tableau[1] = []*Card{makeCard(CardDesignHeart, 1)}
	p.freeCells[0] = makeCard(CardDesignClover, 1)
	p.freeCells[1] = makeCard(CardDesignDiamond, 1)

	err := p.AutoComplete()
	assert.NoError(t, err)
	for i := 0; i < PenguinFoundationCnt; i++ {
		assert.Equal(t, 1, len(p.foundation[i]), "foundation[%d] should have 1 card", i)
	}
}

func TestPenguinAutoCompleteNotPlaying(t *testing.T) {
	p := setupPlayingPenguin()
	p.SetPhase(PenguinPhaseGameOver)
	err := p.AutoComplete()
	assert.Error(t, err)
}

// --- checkGameClear tests ---

func TestPenguinCheckGameClear(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	p.SetBaseRank(1)

	// Fill all foundations to 13 cards
	for d := 0; d < PenguinFoundationCnt; d++ {
		for v := 1; v <= CardValueMax; v++ {
			p.foundation[d] = append(p.foundation[d], makeCard(d+1, v))
		}
	}
	// Remove last card from spade foundation, put it on tableau
	p.foundation[0] = p.foundation[0][:CardValueMax-1]
	p.tableau[0] = []*Card{makeCard(CardDesignSpade, 13)}

	// Place the missing card such that MoveTableauToFoundation triggers GameClear
	// Foundation[0] top is rank 12 (Q); nextRank(12)=13; baseRank=1 so spade goes to fd[0]
	err := p.MoveTableauToFoundation(0)
	assert.NoError(t, err)
	assert.Equal(t, PenguinPhaseGameClear, p.GetPhase())
}

// --- GetHint tests ---

func TestPenguinGetHintTableauToFoundation(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	p.SetBaseRank(5)

	p.tableau[0] = []*Card{makeCard(CardDesignSpade, 5)}
	hint := p.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "tableau", hint.FromZone)
	assert.Equal(t, "foundation", hint.ToZone)
}

func TestPenguinGetHintFreeCellToFoundation(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	p.SetBaseRank(5)

	p.freeCells[0] = makeCard(CardDesignHeart, 5)
	hint := p.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "freecell", hint.FromZone)
	assert.Equal(t, "foundation", hint.ToZone)
}

func TestPenguinGetHintTableauToTableau(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	p.SetBaseRank(5)

	p.tableau[0] = []*Card{makeCard(CardDesignClover, 8)}
	p.tableau[1] = []*Card{makeCard(CardDesignClover, 7)}
	hint := p.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "tableau", hint.FromZone)
	assert.Equal(t, "tableau", hint.ToZone)
}

func TestPenguinGetHintFreeCellToTableau(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	p.SetBaseRank(5)

	p.tableau[0] = []*Card{makeCard(CardDesignDiamond, 8)}
	p.freeCells[0] = makeCard(CardDesignDiamond, 7)
	hint := p.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "freecell", hint.FromZone)
	assert.Equal(t, "tableau", hint.ToZone)
}

func TestPenguinGetHintTableauToFreeCell(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	p.SetBaseRank(5)

	// Only card on tableau: rank 9, can't go to foundation (baseRank=5) or any tableau;
	// no freecell→tableau moves; should suggest tableau→freecell
	p.tableau[0] = []*Card{makeCard(CardDesignSpade, 9)}
	hint := p.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "tableau", hint.FromZone)
	assert.Equal(t, "freecell", hint.ToZone)
}

func TestPenguinGetHintNilWhenNotPlaying(t *testing.T) {
	p := setupPlayingPenguin()
	p.SetPhase(PenguinPhaseGameOver)
	assert.Nil(t, p.GetHint())
}

func TestPenguinGetHintNilWhenNoMoves(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	// Completely empty board; no moves possible
	assert.Nil(t, p.GetHint())
}

// --- Wraparound tests ---

func TestPenguinTableauWraparoundKingOnAce(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	p.SetBaseRank(2)

	// prevRank(1) = 13 (K); place ♠K on ♠A (value=1)
	p.tableau[0] = []*Card{makeCard(CardDesignSpade, 1)}
	p.tableau[1] = []*Card{makeCard(CardDesignSpade, 13)}

	err := p.MoveTableauToTableau(1, -1, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(p.tableau[0]))
}

func TestPenguinFoundationWrapNextRank(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	p.SetBaseRank(5)

	// Foundation has ♠K on top; nextRank(13)=1 (A)
	p.foundation[CardDesignSpade-1] = []*Card{
		makeCard(CardDesignSpade, 5),
		makeCard(CardDesignSpade, 6),
		makeCard(CardDesignSpade, 7),
		makeCard(CardDesignSpade, 8),
		makeCard(CardDesignSpade, 9),
		makeCard(CardDesignSpade, 10),
		makeCard(CardDesignSpade, 11),
		makeCard(CardDesignSpade, 12),
		makeCard(CardDesignSpade, 13),
	}
	p.tableau[0] = []*Card{makeCard(CardDesignSpade, 1)} // A = nextRank(K)

	err := p.MoveTableauToFoundation(0)
	assert.NoError(t, err)
	assert.Equal(t, 10, len(p.foundation[CardDesignSpade-1]))
}

func TestPenguinEmptyColRankWhenBaseIs1(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	p.SetBaseRank(1)

	// prevRank(1) = 13 (K); only K can go on empty col
	p.tableau[0] = []*Card{makeCard(CardDesignHeart, 13)}
	err := p.MoveTableauToTableau(0, -1, 1)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(p.tableau[1]))
}

func TestPenguinEmptyColNonKingRejectedWhenBaseIs1(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	p.SetBaseRank(1)

	// prevRank(1) = 13; Queen (12) cannot go to empty col
	p.tableau[0] = []*Card{makeCard(CardDesignHeart, 12)}
	err := p.MoveTableauToTableau(0, -1, 1)
	assert.Error(t, err)
}

// --- Snapshot / baseRank preservation on Undo ---

func TestPenguinUndoPreservesBaseRank(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	p.SetBaseRank(7)

	p.tableau[0] = []*Card{makeCard(CardDesignSpade, 9)}
	require.NoError(t, p.MoveTableauToFreeCell(0, 0))
	assert.Equal(t, 7, p.GetBaseRank())
	require.NoError(t, p.Undo())
	assert.Equal(t, 7, p.GetBaseRank())
}

// --- JSON marshal/unmarshal tests ---

func TestPenguinJSONRoundTrip(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	p.SetBaseRank(4)

	p.tableau[0] = []*Card{makeCard(CardDesignSpade, 6)}
	require.NoError(t, p.MoveTableauToFreeCell(0, 0))

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var restored Penguin
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, p.GetMoveCount(), restored.GetMoveCount())
	assert.Equal(t, p.GetBaseRank(), restored.GetBaseRank())
	assert.Equal(t, p.GetPhase(), restored.GetPhase())
	assert.True(t, restored.CanUndo())
}

func TestPenguinJSONOversizedHistoryRejected(t *testing.T) {
	bigHistory := make([]map[string]any, penguinMaxSliceLen+1)
	for i := range bigHistory {
		bigHistory[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"hi": bigHistory,
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Penguin
	err = json.Unmarshal(data, &restored)
	assert.Error(t, err)
}

func TestPenguinJSONOversizedTableauColumnRejected(t *testing.T) {
	bigCol := make([]map[string]any, penguinMaxSliceLen+1)
	for i := range bigCol {
		bigCol[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"tb": []any{bigCol, nil, nil, nil, nil, nil, nil},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Penguin
	err = json.Unmarshal(data, &restored)
	assert.Error(t, err)
}

func TestPenguinJSONOversizedFoundationPileRejected(t *testing.T) {
	bigPile := make([]map[string]any, penguinMaxSliceLen+1)
	for i := range bigPile {
		bigPile[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"fd": []any{bigPile, nil, nil, nil},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Penguin
	err = json.Unmarshal(data, &restored)
	assert.Error(t, err)
}

func TestPenguinJSONSnapshotTableauColumnOversizeRejected(t *testing.T) {
	bigCol := make([]map[string]any, penguinMaxSliceLen+1)
	for i := range bigCol {
		bigCol[i] = map[string]any{}
	}
	snapshot := map[string]any{
		"tb": []any{bigCol, nil, nil, nil, nil, nil, nil},
	}
	payload := map[string]any{
		"tc": nil,
		"hi": []any{snapshot},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Penguin
	err = json.Unmarshal(data, &restored)
	assert.Error(t, err)
}

func TestPenguinJSONSnapshotFoundationPileOversizeRejected(t *testing.T) {
	bigPile := make([]map[string]any, penguinMaxSliceLen+1)
	for i := range bigPile {
		bigPile[i] = map[string]any{}
	}
	snapshot := map[string]any{
		"fd": []any{bigPile, nil, nil, nil},
	}
	payload := map[string]any{
		"tc": nil,
		"hi": []any{snapshot},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Penguin
	err = json.Unmarshal(data, &restored)
	assert.Error(t, err)
}

func TestPenguinJSONRestoredUndoWorks(t *testing.T) {
	p := setupPlayingPenguin()
	clearTableauPG(p)
	clearFreeCellsPG(p)
	p.SetBaseRank(4)

	p.tableau[0] = []*Card{makeCard(CardDesignClover, 6)}
	p.tableau[1] = []*Card{makeCard(CardDesignDiamond, 9)}
	require.NoError(t, p.MoveTableauToFreeCell(0, 0))
	require.NoError(t, p.MoveTableauToFreeCell(1, 1))

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var restored Penguin
	require.NoError(t, json.Unmarshal(data, &restored))
	require.True(t, restored.CanUndo())
	require.NoError(t, restored.Undo())
	assert.Equal(t, p.GetMoveCount()-1, restored.GetMoveCount())
}

// --- Setters ---

func TestPenguinSetters(t *testing.T) {
	p := newTestPenguin()

	var tableau [PenguinTableauCnt][]*Card
	tableau[0] = []*Card{makeCard(CardDesignSpade, 1)}
	p.SetTableau(tableau)
	assert.Equal(t, 1, len(p.GetTableau()[0]))

	var cells [PenguinCellCnt]*Card
	cells[6] = makeCard(CardDesignHeart, 1)
	p.SetFreeCells(cells)
	assert.NotNil(t, p.GetFreeCells()[6])

	var fnd [PenguinFoundationCnt][]*Card
	fnd[2] = []*Card{makeCard(CardDesignHeart, 1)}
	p.SetFoundation(fnd)
	assert.Equal(t, 1, len(p.GetFoundation()[2]))

	p.SetBaseRank(9)
	assert.Equal(t, 9, p.GetBaseRank())

	p.SetPhase(PenguinPhaseGameClear)
	assert.Equal(t, PenguinPhaseGameClear, p.GetPhase())

	p.SetIsStalemate(true)
	assert.True(t, p.IsStalemate())
}

// --- NewDefaultPenguin ---

func TestPenguinNewDefault(t *testing.T) {
	p := NewDefaultPenguin()
	assert.NotNil(t, p)
	p.Reset()
	assert.Equal(t, PenguinPhasePlaying, p.GetPhase())
}

// **上限が出ておらず、拒否されたコマンドで初めて気づく形だった (#4802)。**
// 姉妹の Eight Off は supermoveLine を毎ターン出している。
func TestPenguin_GetMaxMovableCards(t *testing.T) {
	card := func(v int) *Card { return NewCard(CardDesignSpade, v, false) }
	board := func(filledCells, filledCols int) *Penguin {
		p := NewDefaultPenguin()
		p.Reset()
		var cells [PenguinCellCnt]*Card
		for i := 0; i < filledCells && i < PenguinCellCnt; i++ {
			cells[i] = card(i + 2)
		}
		p.SetFreeCells(cells)
		var tableau [PenguinTableauCnt][]*Card
		for i := 0; i < PenguinTableauCnt; i++ {
			if i < filledCols {
				tableau[i] = []*Card{card(5)}
			}
		}
		p.SetTableau(tableau)
		return p
	}

	// **(1 + 空きセル) × 2^(空き列)。**セルは足し算、列は掛け算。
	t.Run("counts free cells additively and empty columns as doubling", func(t *testing.T) {
		assert.Equal(t, 1+PenguinCellCnt, board(0, PenguinTableauCnt).GetMaxMovableCards())
		assert.Equal(t, (1+PenguinCellCnt)*2, board(0, PenguinTableauCnt-1).GetMaxMovableCards())
	})

	t.Run("with everything full only a single card moves", func(t *testing.T) {
		assert.Equal(t, 1, board(PenguinCellCnt, PenguinTableauCnt).GetMaxMovableCards())
	})

	// **一般の上限はどの列も除外しない。**特定の列を除外した値を返すと、
	// 空き列が1つ少ない前提の小さすぎる数を出すことになる。
	t.Run("the general limit excludes no column, not even the first", func(t *testing.T) {
		p := NewDefaultPenguin()
		p.Reset()
		p.SetFreeCells([PenguinCellCnt]*Card{})
		var tableau [PenguinTableauCnt][]*Card
		for i := 1; i < PenguinTableauCnt; i++ {
			tableau[i] = []*Card{card(5)}
		}
		p.SetTableau(tableau)
		assert.Equal(t, (1+PenguinCellCnt)*2, p.GetMaxMovableCards())
	})

	// **空き列を移動先にすると上限は下がる。**その列自身を経由地に使えない。
	t.Run("moving onto an empty column halves the limit", func(t *testing.T) {
		p := board(0, PenguinTableauCnt-1)
		assert.Equal(t, (1+PenguinCellCnt)*2, p.GetMaxMovableCards())
		assert.Equal(t, 1+PenguinCellCnt, p.GetMaxMovableCardsToEmptyColumn())
	})

	t.Run("reports zero when there is no empty column to move onto", func(t *testing.T) {
		assert.Equal(t, 0, board(0, PenguinTableauCnt).GetMaxMovableCardsToEmptyColumn())
	})
}
