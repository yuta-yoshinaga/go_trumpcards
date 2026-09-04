//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- helpers ---

func newTestPyramid() *Pyramid {
	tc := NewTrumpCards(0)
	p := NewPyramid(tc)
	p.Reset()
	return p
}

// --- Reset ---

func TestPyramid_Reset(t *testing.T) {
	p := newTestPyramid()

	assert.Equal(t, PyramidPhasePlaying, p.GetPhase())
	assert.Equal(t, 0, p.GetMoveCount())
	assert.False(t, p.IsStalemate())

	// Pyramid has 28 cards (7 rows)
	totalCards := 0
	for row := range PyramidRowCnt {
		assert.Len(t, p.pyramid[row], row+1)
		for _, pc := range p.pyramid[row] {
			assert.NotNil(t, pc.Card)
			assert.False(t, pc.Removed)
			totalCards++
		}
	}
	assert.Equal(t, PyramidCardCnt, totalCards)

	// Remaining 24 cards in stock
	assert.Equal(t, 52-PyramidCardCnt, p.GetStockCount())
	assert.Nil(t, p.GetWaste())
}

func TestPyramid_Reset_ClearsHistory(t *testing.T) {
	p := newTestPyramid()
	// Do a draw to create history
	_ = p.Draw()
	assert.True(t, p.CanUndo())

	p.Reset()
	assert.False(t, p.CanUndo())
	assert.Equal(t, 0, p.GetMoveCount())
}

// --- Draw ---

func TestPyramid_Draw_Success(t *testing.T) {
	p := newTestPyramid()
	stockBefore := p.GetStockCount()

	err := p.Draw()
	require.NoError(t, err)

	assert.Equal(t, stockBefore-1, p.GetStockCount())
	assert.Len(t, p.GetWaste(), 1)
	assert.Equal(t, 1, p.GetMoveCount())
}

func TestPyramid_Draw_EmptyStock(t *testing.T) {
	p := newTestPyramid()
	p.SetStock(nil)

	err := p.Draw()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no cards in stock")
}

func TestPyramid_Draw_NotPlaying(t *testing.T) {
	p := newTestPyramid()
	p.SetPhase(PyramidPhaseGameClear)

	err := p.Draw()
	assert.Error(t, err)
}

// --- IsExposed ---

func TestPyramid_IsExposed_BottomRow(t *testing.T) {
	p := newTestPyramid()
	// Bottom row cards are always exposed
	for col := range PyramidRowCnt {
		assert.True(t, p.IsExposed(PyramidRowCnt-1, col))
	}
}

func TestPyramid_IsExposed_UpperRow_NotExposed(t *testing.T) {
	p := newTestPyramid()
	// Row 5 cards are not exposed because row 6 children are present
	assert.False(t, p.IsExposed(5, 0))
}

func TestPyramid_IsExposed_UpperRow_BothChildrenRemoved(t *testing.T) {
	p := newTestPyramid()
	// Remove both children of (5,0) which are (6,0) and (6,1)
	p.pyramid[6][0].Removed = true
	p.pyramid[6][1].Removed = true

	assert.True(t, p.IsExposed(5, 0))
}

func TestPyramid_IsExposed_UpperRow_OneChildOnly(t *testing.T) {
	p := newTestPyramid()
	// Remove only left child
	p.pyramid[6][0].Removed = true

	assert.False(t, p.IsExposed(5, 0))
}

func TestPyramid_IsExposed_RemovedCard(t *testing.T) {
	p := newTestPyramid()
	p.pyramid[6][0].Removed = true

	assert.False(t, p.IsExposed(6, 0))
}

// --- RemovePair ---

func TestPyramid_RemovePair_Success(t *testing.T) {
	p := newTestPyramid()
	// Set up two exposed bottom-row cards that sum to 13
	p.pyramid[6][0] = &PyramidCard{Card: NewCard(CardDesignSpade, 6, false), Removed: false}
	p.pyramid[6][1] = &PyramidCard{Card: NewCard(CardDesignHeart, 7, false), Removed: false}

	err := p.RemovePair(6, 0, 6, 1)
	require.NoError(t, err)

	assert.True(t, p.pyramid[6][0].Removed)
	assert.True(t, p.pyramid[6][1].Removed)
	assert.Equal(t, 1, p.GetMoveCount())
}

func TestPyramid_RemovePair_NotSum13(t *testing.T) {
	p := newTestPyramid()
	p.pyramid[6][0] = &PyramidCard{Card: NewCard(CardDesignSpade, 5, false), Removed: false}
	p.pyramid[6][1] = &PyramidCard{Card: NewCard(CardDesignHeart, 7, false), Removed: false}

	err := p.RemovePair(6, 0, 6, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "do not sum to 13")
}

func TestPyramid_RemovePair_NotExposed(t *testing.T) {
	p := newTestPyramid()
	// Row 5 card is not exposed
	err := p.RemovePair(5, 0, 6, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not exposed")
}

func TestPyramid_RemovePair_SameCard(t *testing.T) {
	p := newTestPyramid()
	err := p.RemovePair(6, 0, 6, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "same card")
}

func TestPyramid_RemovePair_InvalidPosition(t *testing.T) {
	p := newTestPyramid()
	err := p.RemovePair(-1, 0, 6, 0)
	assert.Error(t, err)

	err = p.RemovePair(6, 0, 6, 8)
	assert.Error(t, err)
}

func TestPyramid_RemovePair_AlreadyRemoved(t *testing.T) {
	p := newTestPyramid()
	p.pyramid[6][0].Removed = true

	err := p.RemovePair(6, 0, 6, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already removed")
}

func TestPyramid_RemovePair_NotPlaying(t *testing.T) {
	p := newTestPyramid()
	p.SetPhase(PyramidPhaseGameOver)

	err := p.RemovePair(6, 0, 6, 1)
	assert.Error(t, err)
}

// --- RemoveKing ---

func TestPyramid_RemoveKing_Success(t *testing.T) {
	p := newTestPyramid()
	p.pyramid[6][0] = &PyramidCard{Card: NewCard(CardDesignSpade, 13, false), Removed: false}

	err := p.RemoveKing(6, 0)
	require.NoError(t, err)

	assert.True(t, p.pyramid[6][0].Removed)
	assert.Equal(t, 1, p.GetMoveCount())
}

func TestPyramid_RemoveKing_NotKing(t *testing.T) {
	p := newTestPyramid()
	p.pyramid[6][0] = &PyramidCard{Card: NewCard(CardDesignSpade, 5, false), Removed: false}

	err := p.RemoveKing(6, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a King")
}

func TestPyramid_RemoveKing_NotExposed(t *testing.T) {
	p := newTestPyramid()
	// Set row 5 card to King but it's not exposed
	p.pyramid[5][0] = &PyramidCard{Card: NewCard(CardDesignSpade, 13, false), Removed: false}

	err := p.RemoveKing(5, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not exposed")
}

func TestPyramid_RemoveKing_AlreadyRemoved(t *testing.T) {
	p := newTestPyramid()
	p.pyramid[6][0] = &PyramidCard{Card: NewCard(CardDesignSpade, 13, false), Removed: true}

	err := p.RemoveKing(6, 0)
	assert.Error(t, err)
}

func TestPyramid_RemoveKing_InvalidPosition(t *testing.T) {
	p := newTestPyramid()
	err := p.RemoveKing(7, 0)
	assert.Error(t, err)
}

// --- RemoveWithWaste ---

func TestPyramid_RemoveWithWaste_Success(t *testing.T) {
	p := newTestPyramid()
	p.pyramid[6][0] = &PyramidCard{Card: NewCard(CardDesignSpade, 8, false), Removed: false}
	p.SetWaste([]*Card{NewCard(CardDesignHeart, 5, false)})

	err := p.RemoveWithWaste(6, 0)
	require.NoError(t, err)

	assert.True(t, p.pyramid[6][0].Removed)
	assert.Empty(t, p.GetWaste())
	assert.Equal(t, 1, p.GetMoveCount())
}

func TestPyramid_RemoveWithWaste_NotSum13(t *testing.T) {
	p := newTestPyramid()
	p.pyramid[6][0] = &PyramidCard{Card: NewCard(CardDesignSpade, 8, false), Removed: false}
	p.SetWaste([]*Card{NewCard(CardDesignHeart, 3, false)})

	err := p.RemoveWithWaste(6, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "do not sum to 13")
}

func TestPyramid_RemoveWithWaste_EmptyWaste(t *testing.T) {
	p := newTestPyramid()
	err := p.RemoveWithWaste(6, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "waste is empty")
}

func TestPyramid_RemoveWithWaste_NotExposed(t *testing.T) {
	p := newTestPyramid()
	p.SetWaste([]*Card{NewCard(CardDesignHeart, 5, false)})

	err := p.RemoveWithWaste(5, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not exposed")
}

func TestPyramid_RemoveWithWaste_AlreadyRemoved(t *testing.T) {
	p := newTestPyramid()
	p.pyramid[6][0].Removed = true
	p.SetWaste([]*Card{NewCard(CardDesignHeart, 5, false)})

	err := p.RemoveWithWaste(6, 0)
	assert.Error(t, err)
}

// --- RemoveWasteKing ---

func TestPyramid_RemoveWasteKing_Success(t *testing.T) {
	p := newTestPyramid()
	p.SetWaste([]*Card{NewCard(CardDesignSpade, 13, false)})

	err := p.RemoveWasteKing()
	require.NoError(t, err)

	assert.Empty(t, p.GetWaste())
	assert.Equal(t, 1, p.GetMoveCount())
}

func TestPyramid_RemoveWasteKing_NotKing(t *testing.T) {
	p := newTestPyramid()
	p.SetWaste([]*Card{NewCard(CardDesignSpade, 5, false)})

	err := p.RemoveWasteKing()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a King")
}

func TestPyramid_RemoveWasteKing_EmptyWaste(t *testing.T) {
	p := newTestPyramid()
	err := p.RemoveWasteKing()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "waste is empty")
}

func TestPyramid_RemoveWasteKing_NotPlaying(t *testing.T) {
	p := newTestPyramid()
	p.SetPhase(PyramidPhaseGameOver)
	p.SetWaste([]*Card{NewCard(CardDesignSpade, 13, false)})

	err := p.RemoveWasteKing()
	assert.Error(t, err)
}

// --- GiveUp ---

func TestPyramid_GiveUp(t *testing.T) {
	p := newTestPyramid()
	p.GiveUp()

	assert.Equal(t, PyramidPhaseGameOver, p.GetPhase())
}

func TestPyramid_GiveUp_NotPlaying(t *testing.T) {
	p := newTestPyramid()
	p.SetPhase(PyramidPhaseGameClear)
	p.GiveUp()

	// Phase should not change
	assert.Equal(t, PyramidPhaseGameClear, p.GetPhase())
}

// --- Undo ---

func TestPyramid_Undo_Draw(t *testing.T) {
	p := newTestPyramid()
	stockBefore := p.GetStockCount()

	_ = p.Draw()
	assert.Equal(t, stockBefore-1, p.GetStockCount())

	err := p.Undo()
	require.NoError(t, err)
	assert.Equal(t, stockBefore, p.GetStockCount())
	assert.Empty(t, p.GetWaste())
}

func TestPyramid_Undo_RemovePair(t *testing.T) {
	p := newTestPyramid()
	p.pyramid[6][0] = &PyramidCard{Card: NewCard(CardDesignSpade, 6, false), Removed: false}
	p.pyramid[6][1] = &PyramidCard{Card: NewCard(CardDesignHeart, 7, false), Removed: false}

	_ = p.RemovePair(6, 0, 6, 1)
	assert.True(t, p.pyramid[6][0].Removed)

	err := p.Undo()
	require.NoError(t, err)
	assert.False(t, p.pyramid[6][0].Removed)
	assert.False(t, p.pyramid[6][1].Removed)
}

func TestPyramid_Undo_NoHistory(t *testing.T) {
	p := newTestPyramid()
	err := p.Undo()
	assert.Error(t, err)
}

func TestPyramid_Undo_NotPlaying(t *testing.T) {
	p := newTestPyramid()
	p.SetPhase(PyramidPhaseGameOver)
	err := p.Undo()
	assert.Error(t, err)
}

func TestPyramid_CanUndo(t *testing.T) {
	p := newTestPyramid()
	assert.False(t, p.CanUndo())

	_ = p.Draw()
	assert.True(t, p.CanUndo())
}

// --- GameClear ---

func TestPyramid_GameClear(t *testing.T) {
	p := newTestPyramid()
	// Remove all pyramid cards except two that sum to 13
	for row := range PyramidRowCnt {
		for col := range row + 1 {
			p.pyramid[row][col].Removed = true
		}
	}
	// Set last two as non-removed, summing to 13
	p.pyramid[6][0] = &PyramidCard{Card: NewCard(CardDesignSpade, 6, false), Removed: false}
	p.pyramid[6][1] = &PyramidCard{Card: NewCard(CardDesignHeart, 7, false), Removed: false}

	err := p.RemovePair(6, 0, 6, 1)
	require.NoError(t, err)
	assert.Equal(t, PyramidPhaseGameClear, p.GetPhase())
}

// --- Stalemate ---

func TestPyramid_Stalemate_NoMovesNoStock(t *testing.T) {
	p := newTestPyramid()
	p.SetStock(nil)
	// Set bottom row to values that don't pair to 13 and aren't kings
	for col := range PyramidRowCnt {
		p.pyramid[6][col] = &PyramidCard{Card: NewCard(CardDesignSpade, 1, false), Removed: false}
	}
	p.SetWaste(nil)

	// Force stalemate check via a draw attempt (will fail, but let's check manually)
	p.checkStalemate()
	assert.True(t, p.IsStalemate())
}

func TestPyramid_Stalemate_StockRemaining(t *testing.T) {
	p := newTestPyramid()
	// Even if no hint, stock has cards so not stalemate
	p.checkStalemate()
	assert.False(t, p.IsStalemate())
}

// --- GetHint ---

func TestPyramid_GetHint_King(t *testing.T) {
	p := newTestPyramid()
	p.pyramid[6][0] = &PyramidCard{Card: NewCard(CardDesignSpade, 13, false), Removed: false}

	hint := p.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "king", hint.Type)
	assert.Equal(t, 6, hint.Row1)
	assert.Equal(t, 0, hint.Col1)
}

func TestPyramid_GetHint_Pair(t *testing.T) {
	p := newTestPyramid()
	// No kings in bottom row, but two cards summing to 13
	for col := range PyramidRowCnt {
		p.pyramid[6][col] = &PyramidCard{Card: NewCard(CardDesignSpade, 1, false), Removed: false}
	}
	p.pyramid[6][0] = &PyramidCard{Card: NewCard(CardDesignSpade, 5, false), Removed: false}
	p.pyramid[6][1] = &PyramidCard{Card: NewCard(CardDesignHeart, 8, false), Removed: false}

	hint := p.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "pair", hint.Type)
}

func TestPyramid_GetHint_WasteKing(t *testing.T) {
	p := newTestPyramid()
	// No kings or pairs in pyramid
	for col := range PyramidRowCnt {
		p.pyramid[6][col] = &PyramidCard{Card: NewCard(CardDesignSpade, 1, false), Removed: false}
	}
	p.SetWaste([]*Card{NewCard(CardDesignHeart, 13, false)})

	hint := p.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "waste_king", hint.Type)
}

func TestPyramid_GetHint_WastePair(t *testing.T) {
	p := newTestPyramid()
	for col := range PyramidRowCnt {
		p.pyramid[6][col] = &PyramidCard{Card: NewCard(CardDesignSpade, 1, false), Removed: false}
	}
	p.pyramid[6][0] = &PyramidCard{Card: NewCard(CardDesignSpade, 5, false), Removed: false}
	p.SetWaste([]*Card{NewCard(CardDesignHeart, 8, false)})

	hint := p.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "waste_pair", hint.Type)
	assert.Equal(t, 6, hint.Row1)
	assert.Equal(t, 0, hint.Col1)
}

func TestPyramid_GetHint_NoHint(t *testing.T) {
	p := newTestPyramid()
	// All bottom row = 1, no waste, no kings → no hint
	for col := range PyramidRowCnt {
		p.pyramid[6][col] = &PyramidCard{Card: NewCard(CardDesignSpade, 1, false), Removed: false}
	}
	p.SetWaste(nil)

	hint := p.GetHint()
	assert.Nil(t, hint)
}

func TestPyramid_GetHint_NotPlaying(t *testing.T) {
	p := newTestPyramid()
	p.SetPhase(PyramidPhaseGameOver)

	hint := p.GetHint()
	assert.Nil(t, hint)
}

// --- AllRemoved ---

func TestPyramid_AllRemoved_True(t *testing.T) {
	p := newTestPyramid()
	for row := range PyramidRowCnt {
		for col := range row + 1 {
			p.pyramid[row][col].Removed = true
		}
	}
	assert.True(t, p.AllRemoved())
}

func TestPyramid_AllRemoved_False(t *testing.T) {
	p := newTestPyramid()
	assert.False(t, p.AllRemoved())
}

// --- ActionLog ---

func TestPyramid_ActionLog(t *testing.T) {
	p := newTestPyramid()
	assert.Nil(t, p.GetActionLog())

	_ = p.Draw()
	log := p.GetActionLog()
	assert.Len(t, log, 1)
	assert.Equal(t, "draw", log[0].ActionType)
}

// --- validatePyramidPos ---

func TestPyramid_ValidatePyramidPos(t *testing.T) {
	p := newTestPyramid()

	// Valid positions
	assert.NoError(t, p.validatePyramidPos(0, 0))
	assert.NoError(t, p.validatePyramidPos(6, 6))

	// Invalid row
	assert.Error(t, p.validatePyramidPos(-1, 0))
	assert.Error(t, p.validatePyramidPos(7, 0))

	// Invalid col (col > row)
	assert.Error(t, p.validatePyramidPos(0, 1))
	assert.Error(t, p.validatePyramidPos(3, 4))
	assert.Error(t, p.validatePyramidPos(0, -1))
}

// --- Expose after removal ---

func TestPyramid_ExposedAfterChildrenRemoved(t *testing.T) {
	p := newTestPyramid()
	// (5,2) is exposed when both (6,2) and (6,3) are removed
	assert.False(t, p.IsExposed(5, 2))

	p.pyramid[6][2].Removed = true
	assert.False(t, p.IsExposed(5, 2))

	p.pyramid[6][3].Removed = true
	assert.True(t, p.IsExposed(5, 2))
}

// --- RemovePair exposes parent ---

func TestPyramid_RemovePair_ExposesParent(t *testing.T) {
	p := newTestPyramid()
	// Set bottom row cards to pair
	p.pyramid[6][2] = &PyramidCard{Card: NewCard(CardDesignSpade, 6, false), Removed: false}
	p.pyramid[6][3] = &PyramidCard{Card: NewCard(CardDesignHeart, 7, false), Removed: false}

	assert.False(t, p.IsExposed(5, 2))

	err := p.RemovePair(6, 2, 6, 3)
	require.NoError(t, err)

	assert.True(t, p.IsExposed(5, 2))
}

// --- UndoToEscape / UndoN tests ---

func TestPyramid_UndoToEscape_NotInStalemate(t *testing.T) {
	p := newTestPyramid()
	assert.Equal(t, 0, p.UndoToEscape())
}

func TestPyramid_UndoToEscape_StalemateNoHistory(t *testing.T) {
	p := newTestPyramid()
	p.SetIsStalemate(true)
	assert.Equal(t, -1, p.UndoToEscape())
}

func TestPyramid_UndoToEscape_StalemateWithEscape(t *testing.T) {
	p := newTestPyramid()
	_ = p.Draw()
	p.SetIsStalemate(true)
	n := p.UndoToEscape()
	assert.Equal(t, 1, n)
}

func TestPyramid_UndoN_Zero(t *testing.T) {
	p := newTestPyramid()
	err := p.UndoN(0)
	assert.NoError(t, err)
}

func TestPyramid_UndoN_Valid(t *testing.T) {
	p := newTestPyramid()
	_ = p.Draw()
	_ = p.Draw()
	err := p.UndoN(2)
	assert.NoError(t, err)
}

func TestPyramid_UndoN_Excessive(t *testing.T) {
	p := newTestPyramid()
	_ = p.Draw()
	err := p.UndoN(5)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "undo step")
}

// **キングは相方が要らず単独で消せる (#4782)。**Web は常時ハイライトして
// いるのに、CUI は数値を1枚ずつ見て 13 を自分で探すしかなかった。
func TestPyramid_IsRemovableKing(t *testing.T) {
	// 最下段だけを持つ最小のピラミッドを組む (最下段は必ず露出している)。
	newBoard := func(bottom ...*Card) *Pyramid {
		p := NewDefaultPyramid()
		p.Reset()
		var rows [PyramidRowCnt][]*PyramidCard
		for r := range PyramidRowCnt {
			rows[r] = make([]*PyramidCard, r+1)
			for c := range r + 1 {
				card := NewCard(CardDesignSpade, 2, false)
				removed := true
				if r == PyramidRowCnt-1 && c < len(bottom) {
					card = bottom[c]
					removed = false
				}
				rows[r][c] = &PyramidCard{Card: card, Removed: removed}
			}
		}
		p.SetPyramid(rows)
		return p
	}
	last := PyramidRowCnt - 1

	t.Run("an exposed king can be removed on its own", func(t *testing.T) {
		p := newBoard(NewCard(CardDesignSpade, 13, false))
		assert.True(t, p.IsRemovableKing(last, 0))
	})

	// **13 以外には印を付けない。**Q に印が付くと、通らない手を勧めることになる。
	t.Run("a queen is not a removable king", func(t *testing.T) {
		p := newBoard(NewCard(CardDesignSpade, 12, false))
		assert.False(t, p.IsRemovableKing(last, 0))
	})

	// **除去済みの札に印を付けない。**RemoveKing で消した後も印が残ると、
	// 通らない手を勧め続ける。(isExposed は除去済みを弾かないので、
	// Removed の確認はここでしか踏めない。)
	t.Run("a removed king is no longer removable", func(t *testing.T) {
		p := newBoard(NewCard(CardDesignSpade, 13, false))
		require.True(t, p.IsRemovableKing(last, 0))
		require.NoError(t, p.RemoveKing(last, 0))
		assert.False(t, p.IsRemovableKing(last, 0),
			"消えた札に印が残っている")
	})

	// **印が付く条件と RemoveKing が通る条件は同じでなければならない。**
	t.Run("agrees with RemoveKing", func(t *testing.T) {
		p := newBoard(NewCard(CardDesignSpade, 13, false), NewCard(CardDesignHeart, 5, false))
		require.True(t, p.IsRemovableKing(last, 0))
		assert.NoError(t, p.RemoveKing(last, 0))

		q := newBoard(NewCard(CardDesignSpade, 13, false), NewCard(CardDesignHeart, 5, false))
		require.False(t, q.IsRemovableKing(last, 1))
		assert.Error(t, q.RemoveKing(last, 1))
	})

	t.Run("out-of-range coordinates are not removable", func(t *testing.T) {
		p := newBoard(NewCard(CardDesignSpade, 13, false))
		assert.False(t, p.IsRemovableKing(-1, 0))
		assert.False(t, p.IsRemovableKing(last, 99))
	})

	t.Run("the waste top is flagged only when it is a king", func(t *testing.T) {
		p := newBoard(NewCard(CardDesignSpade, 2, false))
		assert.False(t, p.IsWasteKingRemovable(), "空のウェイストに印は付かない")

		p.SetWaste([]*Card{NewCard(CardDesignHeart, 5, false)})
		assert.False(t, p.IsWasteKingRemovable())

		p.SetWaste([]*Card{NewCard(CardDesignHeart, 5, false), NewCard(CardDesignClover, 13, false)})
		assert.True(t, p.IsWasteKingRemovable(), "見るのは一番上の1枚")
	})
}

// The web keeps a persistent record in localStorage; the CUI has no store, so
// these count the current process only. What matters is that Reset does NOT
// clear them -- that is the whole point of the panel.
func TestPyramidSessionStats(t *testing.T) {
	t.Run("counts nothing before a game finishes", func(t *testing.T) {
		p := NewDefaultPyramid()
		p.Reset()
		// Reset also runs at startup; counting there would report a play that
		// nobody played.
		assert.Equal(t, 0, p.GetSessionPlays())
		assert.Equal(t, 0, p.GetSessionFewestMoves(), "0 means no record yet")
	})

	t.Run("counts a giveup as a play but not a win", func(t *testing.T) {
		p := NewDefaultPyramid()
		p.Reset()
		p.GiveUp()
		assert.Equal(t, 1, p.GetSessionPlays())
		assert.Equal(t, 0, p.GetSessionWins())
		assert.Equal(t, 0, p.GetSessionFewestMoves(), "a loss sets no best")
	})

	t.Run("keeps the totals across Reset", func(t *testing.T) {
		p := NewDefaultPyramid()
		p.Reset()
		p.GiveUp()
		p.Reset()
		p.GiveUp()
		assert.Equal(t, 2, p.GetSessionPlays(), "Reset must not wipe the session record")
	})
}
