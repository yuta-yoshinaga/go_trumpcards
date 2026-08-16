//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestTriPeaks creates a TriPeaks with a fresh deck for testing.
func newTestTriPeaks() *TriPeaks {
	return NewTriPeaks(NewTrumpCards(0))
}

// setupTriPeaksForRemove sets up a TriPeaks game with a known layout for testing Remove.
// Places a 5♠ at (3,0) exposed, and sets waste top to 4♥.
func setupTriPeaksForRemove(t *testing.T) *TriPeaks {
	t.Helper()
	tp := newTestTriPeaks()
	tp.phase = TriPeaksPhasePlaying

	// Clear layout
	for r := range TriPeaksRowCnt {
		for c := range TriPeaksColCnt {
			tp.layout[r][c] = nil
		}
	}

	// Place card at (3,0) — bottom row, always exposed
	tp.layout[3][0] = &TriPeaksCard{
		Card:    NewCard(CardDesignSpade, 5, true),
		Removed: false,
	}

	// Waste top: 4♥ (adjacent to 5)
	tp.waste = []*Card{NewCard(CardDesignHeart, 4, true)}
	tp.stock = []*Card{NewCard(CardDesignDiamond, 7, true)}
	return tp
}

func TestTriPeaks_NewTriPeaks(t *testing.T) {
	tp := NewTriPeaks(NewTrumpCards(0))
	assert.NotNil(t, tp)
	assert.Equal(t, TriPeaksPhase(0), tp.GetPhase())
}

func TestTriPeaks_Reset(t *testing.T) {
	tp := newTestTriPeaks()
	tp.Reset()

	assert.Equal(t, TriPeaksPhasePlaying, tp.GetPhase())
	assert.Equal(t, 0, tp.GetMoveCount())
	assert.False(t, tp.IsStalemate())

	// Count tableau cards
	tableauCount := 0
	for r := range TriPeaksRowCnt {
		for c := range TriPeaksColCnt {
			if tp.layout[r][c] != nil {
				tableauCount++
			}
		}
	}
	assert.Equal(t, TriPeaksTableauCnt, tableauCount)

	// Stock + waste should account for remaining cards (52 - 28 = 24; 23 stock + 1 waste)
	assert.Equal(t, 23, tp.GetStockCount())
	assert.Len(t, tp.GetWaste(), 1)
}

func TestTriPeaks_Reset_ValidPositions(t *testing.T) {
	tp := newTestTriPeaks()
	tp.Reset()

	// Verify valid positions have cards and invalid positions are nil
	for r := range TriPeaksRowCnt {
		for c := range TriPeaksColCnt {
			if triPeaksValidPos[r][c] {
				assert.NotNil(t, tp.layout[r][c], "expected card at (%d,%d)", r, c)
			} else {
				assert.Nil(t, tp.layout[r][c], "expected nil at (%d,%d)", r, c)
			}
		}
	}
}

func TestTriPeaks_Draw(t *testing.T) {
	tp := newTestTriPeaks()
	tp.Reset()
	initialStock := tp.GetStockCount()
	initialWaste := len(tp.GetWaste())

	err := tp.Draw()
	require.NoError(t, err)
	assert.Equal(t, initialStock-1, tp.GetStockCount())
	assert.Len(t, tp.GetWaste(), initialWaste+1)
	assert.Equal(t, 1, tp.GetMoveCount())
}

func TestTriPeaks_Draw_EmptyStock(t *testing.T) {
	tp := setupTriPeaksForRemove(t)
	tp.stock = nil

	err := tp.Draw()
	assert.EqualError(t, err, "no cards in stock")
}

func TestTriPeaks_Draw_NotPlaying(t *testing.T) {
	tp := newTestTriPeaks()
	tp.SetPhase(TriPeaksPhaseGameOver)

	err := tp.Draw()
	assert.EqualError(t, err, "game is not in playing phase")
}

func TestTriPeaks_Remove_Success(t *testing.T) {
	tp := setupTriPeaksForRemove(t)

	err := tp.Remove(3, 0)
	require.NoError(t, err)
	assert.True(t, tp.layout[3][0].Removed)
	assert.Equal(t, 1, tp.GetMoveCount())
	// Removed card goes to waste
	assert.Equal(t, 5, tp.GetWaste()[len(tp.GetWaste())-1].GetValue())
}

func TestTriPeaks_Remove_KAWrap(t *testing.T) {
	tp := setupTriPeaksForRemove(t)
	// Place K at (3,0), waste top is A
	tp.layout[3][0] = &TriPeaksCard{Card: NewCard(CardDesignSpade, 13, true), Removed: false}
	tp.waste = []*Card{NewCard(CardDesignHeart, 1, true)}

	err := tp.Remove(3, 0)
	require.NoError(t, err)
	assert.True(t, tp.layout[3][0].Removed)
}

func TestTriPeaks_Remove_AKWrap(t *testing.T) {
	tp := setupTriPeaksForRemove(t)
	// Place A at (3,0), waste top is K
	tp.layout[3][0] = &TriPeaksCard{Card: NewCard(CardDesignSpade, 1, true), Removed: false}
	tp.waste = []*Card{NewCard(CardDesignHeart, 13, true)}

	err := tp.Remove(3, 0)
	require.NoError(t, err)
	assert.True(t, tp.layout[3][0].Removed)
}

func TestTriPeaks_Remove_NotAdjacent(t *testing.T) {
	tp := setupTriPeaksForRemove(t)
	// Place 9♠ at (3,0), waste top is 4♥ — not adjacent
	tp.layout[3][0] = &TriPeaksCard{Card: NewCard(CardDesignSpade, 9, true), Removed: false}

	err := tp.Remove(3, 0)
	assert.EqualError(t, err, "card is not adjacent to waste top")
}

func TestTriPeaks_Remove_NotExposed(t *testing.T) {
	tp := setupTriPeaksForRemove(t)
	// Place card at (2,0) with children (3,0) and (3,1) not removed
	tp.layout[2][0] = &TriPeaksCard{Card: NewCard(CardDesignSpade, 3, true), Removed: false}
	tp.layout[3][1] = &TriPeaksCard{Card: NewCard(CardDesignHeart, 2, true), Removed: false}

	err := tp.Remove(2, 0)
	assert.EqualError(t, err, "card is not exposed")
}

func TestTriPeaks_Remove_AlreadyRemoved(t *testing.T) {
	tp := setupTriPeaksForRemove(t)
	tp.layout[3][0].Removed = true

	err := tp.Remove(3, 0)
	assert.EqualError(t, err, "card is already removed")
}

func TestTriPeaks_Remove_EmptyWaste(t *testing.T) {
	tp := setupTriPeaksForRemove(t)
	tp.waste = nil

	err := tp.Remove(3, 0)
	assert.EqualError(t, err, "waste is empty")
}

func TestTriPeaks_Remove_InvalidRow(t *testing.T) {
	tp := setupTriPeaksForRemove(t)

	err := tp.Remove(-1, 0)
	assert.EqualError(t, err, "invalid row")

	err = tp.Remove(4, 0)
	assert.EqualError(t, err, "invalid row")
}

func TestTriPeaks_Remove_InvalidCol(t *testing.T) {
	tp := setupTriPeaksForRemove(t)

	err := tp.Remove(3, -1)
	assert.EqualError(t, err, "invalid column")

	err = tp.Remove(3, 10)
	assert.EqualError(t, err, "invalid column")
}

func TestTriPeaks_Remove_InvalidPosition(t *testing.T) {
	tp := setupTriPeaksForRemove(t)

	// Row 0, col 1 is not a valid position
	err := tp.Remove(0, 1)
	assert.EqualError(t, err, "invalid position")
}

func TestTriPeaks_Remove_NilCard(t *testing.T) {
	tp := setupTriPeaksForRemove(t)

	// Row 3, col 5 is valid but has no card in our test setup
	err := tp.Remove(3, 5)
	assert.EqualError(t, err, "no card at position")
}

func TestTriPeaks_Remove_NotPlaying(t *testing.T) {
	tp := setupTriPeaksForRemove(t)
	tp.phase = TriPeaksPhaseGameOver

	err := tp.Remove(3, 0)
	assert.EqualError(t, err, "game is not in playing phase")
}

func TestTriPeaks_GiveUp(t *testing.T) {
	tp := newTestTriPeaks()
	tp.Reset()
	tp.GiveUp()
	assert.Equal(t, TriPeaksPhaseGameOver, tp.GetPhase())
}

func TestTriPeaks_GiveUp_NotPlaying(t *testing.T) {
	tp := newTestTriPeaks()
	tp.SetPhase(TriPeaksPhaseGameClear)
	tp.GiveUp()
	// Phase should not change
	assert.Equal(t, TriPeaksPhaseGameClear, tp.GetPhase())
}

func TestTriPeaks_GetHint_RemoveAvailable(t *testing.T) {
	tp := setupTriPeaksForRemove(t)

	hint := tp.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "remove", hint.Type)
	assert.Equal(t, 3, hint.Row)
	assert.Equal(t, 0, hint.Col)
}

func TestTriPeaks_GetHint_DrawOnly(t *testing.T) {
	tp := setupTriPeaksForRemove(t)
	// Make card not adjacent to waste
	tp.layout[3][0] = &TriPeaksCard{Card: NewCard(CardDesignSpade, 9, true), Removed: false}

	hint := tp.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "draw", hint.Type)
	assert.Equal(t, -1, hint.Row)
}

func TestTriPeaks_GetHint_NoHint(t *testing.T) {
	tp := setupTriPeaksForRemove(t)
	tp.layout[3][0] = &TriPeaksCard{Card: NewCard(CardDesignSpade, 9, true), Removed: false}
	tp.stock = nil

	hint := tp.GetHint()
	assert.Nil(t, hint)
}

func TestTriPeaks_GetHint_NotPlaying(t *testing.T) {
	tp := newTestTriPeaks()
	tp.SetPhase(TriPeaksPhaseGameOver)

	hint := tp.GetHint()
	assert.Nil(t, hint)
}

func TestTriPeaks_GetHint_EmptyWaste(t *testing.T) {
	tp := setupTriPeaksForRemove(t)
	tp.waste = nil

	hint := tp.GetHint()
	assert.Nil(t, hint)
}

func TestTriPeaks_Undo(t *testing.T) {
	tp := setupTriPeaksForRemove(t)
	// Add a second card so removing one does NOT trigger game clear
	tp.layout[3][1] = &TriPeaksCard{Card: NewCard(CardDesignHeart, 6, true), Removed: false}

	// Remove a card, then undo
	err := tp.Remove(3, 0)
	require.NoError(t, err)
	assert.True(t, tp.layout[3][0].Removed)
	assert.Equal(t, TriPeaksPhasePlaying, tp.GetPhase())

	err = tp.Undo()
	require.NoError(t, err)
	assert.False(t, tp.layout[3][0].Removed)
	assert.Equal(t, 0, tp.GetMoveCount())
}

func TestTriPeaks_Undo_NoHistory(t *testing.T) {
	tp := newTestTriPeaks()
	tp.Reset()

	err := tp.Undo()
	assert.EqualError(t, err, "cannot undo: no history")
}

func TestTriPeaks_Undo_NotPlaying(t *testing.T) {
	tp := newTestTriPeaks()
	tp.SetPhase(TriPeaksPhaseGameOver)

	err := tp.Undo()
	assert.EqualError(t, err, "cannot undo: game is not in playing phase")
}

func TestTriPeaks_CanUndo(t *testing.T) {
	tp := newTestTriPeaks()
	tp.Reset()
	assert.False(t, tp.CanUndo())

	_ = tp.Draw()
	assert.True(t, tp.CanUndo())

	tp.SetPhase(TriPeaksPhaseGameOver)
	assert.False(t, tp.CanUndo())
}

func TestTriPeaks_IsExposed(t *testing.T) {
	tp := setupTriPeaksForRemove(t)

	// Bottom row card is exposed
	assert.True(t, tp.IsExposed(3, 0))

	// Add card above with children not removed
	tp.layout[2][0] = &TriPeaksCard{Card: NewCard(CardDesignHeart, 3, true), Removed: false}
	tp.layout[3][1] = &TriPeaksCard{Card: NewCard(CardDesignClover, 7, true), Removed: false}
	assert.False(t, tp.IsExposed(2, 0))

	// Remove both children
	tp.layout[3][0].Removed = true
	tp.layout[3][1].Removed = true
	assert.True(t, tp.IsExposed(2, 0))
}

func TestTriPeaks_IsExposed_InvalidPos(t *testing.T) {
	tp := newTestTriPeaks()
	tp.Reset()
	// Invalid position
	assert.False(t, tp.IsExposed(0, 1))
}

func TestTriPeaks_IsExposed_RemovedCard(t *testing.T) {
	tp := setupTriPeaksForRemove(t)
	tp.layout[3][0].Removed = true
	assert.False(t, tp.IsExposed(3, 0))
}

func TestTriPeaks_AllRemoved(t *testing.T) {
	tp := setupTriPeaksForRemove(t)
	assert.False(t, tp.AllRemoved())

	// Remove all cards
	tp.layout[3][0].Removed = true
	assert.True(t, tp.AllRemoved())
}

func TestTriPeaks_GameClear(t *testing.T) {
	tp := setupTriPeaksForRemove(t)

	// Remove the only card — should trigger game clear
	err := tp.Remove(3, 0)
	require.NoError(t, err)
	assert.Equal(t, TriPeaksPhaseGameClear, tp.GetPhase())
}

func TestTriPeaks_Stalemate(t *testing.T) {
	tp := setupTriPeaksForRemove(t)
	// Make card not adjacent, empty stock
	tp.layout[3][0] = &TriPeaksCard{Card: NewCard(CardDesignSpade, 9, true), Removed: false}
	tp.stock = nil

	// Draw should fail but stalemate should be detected via checkStalemate
	// Trigger a draw attempt to see error
	err := tp.Draw()
	assert.Error(t, err)

	// Manually trigger stalemate check
	tp.checkStalemate()
	assert.True(t, tp.IsStalemate())
}

func TestTriPeaks_Stalemate_NotPlayingSkips(t *testing.T) {
	tp := newTestTriPeaks()
	tp.SetPhase(TriPeaksPhaseGameClear)
	tp.checkStalemate()
	assert.False(t, tp.IsStalemate())
}

func TestTriPeaks_IsAdjacentRank(t *testing.T) {
	tp := newTestTriPeaks()
	tests := []struct {
		v1, v2 int
		want   bool
	}{
		{1, 2, true},   // A-2
		{2, 1, true},   // 2-A
		{5, 4, true},   // 5-4
		{5, 6, true},   // 5-6
		{13, 1, true},  // K-A wrap
		{1, 13, true},  // A-K wrap
		{13, 12, true}, // K-Q
		{5, 7, false},  // not adjacent
		{5, 5, false},  // same card
		{1, 3, false},  // not adjacent
	}
	for _, tt := range tests {
		c1 := NewCard(CardDesignSpade, tt.v1, true)
		c2 := NewCard(CardDesignHeart, tt.v2, true)
		assert.Equal(t, tt.want, tp.isAdjacentRank(c1, c2),
			"isAdjacentRank(%d, %d) = %v, want %v", tt.v1, tt.v2, !tt.want, tt.want)
	}
}

func TestTriPeaks_JSON_RoundTrip(t *testing.T) {
	tp := newTestTriPeaks()
	tp.Reset()
	_ = tp.Draw()

	data, err := json.Marshal(tp)
	require.NoError(t, err)

	var tp2 TriPeaks
	err = json.Unmarshal(data, &tp2)
	require.NoError(t, err)

	assert.Equal(t, tp.GetPhase(), tp2.GetPhase())
	assert.Equal(t, tp.GetMoveCount(), tp2.GetMoveCount())
	assert.Equal(t, tp.GetStockCount(), tp2.GetStockCount())
	assert.Equal(t, len(tp.GetWaste()), len(tp2.GetWaste()))
	assert.Equal(t, tp.IsStalemate(), tp2.IsStalemate())
}

func TestTriPeaks_JSON_MaxSliceLen(t *testing.T) {
	// Create JSON with oversized stock array
	bigStock := make([]*Card, triPeaksMaxSliceLen+1)
	j := triPeaksJSON{
		Stock: bigStock,
	}
	data, err := json.Marshal(j)
	require.NoError(t, err)

	var tp TriPeaks
	err = json.Unmarshal(data, &tp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum allowed size")
}

func TestTriPeaks_JSON_NilFields(t *testing.T) {
	// Unmarshal with nil stock/waste/actionLog
	j := triPeaksJSON{
		Phase: TriPeaksPhasePlaying,
	}
	data, err := json.Marshal(j)
	require.NoError(t, err)

	var tp TriPeaks
	err = json.Unmarshal(data, &tp)
	require.NoError(t, err)
	assert.NotNil(t, tp.stock)
	assert.NotNil(t, tp.waste)
	assert.NotNil(t, tp.actionLog)
	assert.NotNil(t, tp.trumpCards)
}

func TestTriPeaks_GetLayout(t *testing.T) {
	tp := newTestTriPeaks()
	tp.Reset()
	layout := tp.GetLayout()
	// Peak at (0,0) should have a card
	assert.NotNil(t, layout[0][0])
}

func TestTriPeaks_GetActionLog(t *testing.T) {
	tp := newTestTriPeaks()
	tp.Reset()
	assert.Nil(t, tp.GetActionLog())

	_ = tp.Draw()
	assert.Len(t, tp.GetActionLog(), 1)
}

func TestTriPeaks_SettersForTest(t *testing.T) {
	tp := newTestTriPeaks()
	tp.SetPhase(TriPeaksPhaseGameClear)
	assert.Equal(t, TriPeaksPhaseGameClear, tp.GetPhase())

	tp.SetIsStalemate(true)
	assert.True(t, tp.IsStalemate())

	stock := []*Card{NewCard(CardDesignSpade, 1, true)}
	tp.SetStock(stock)
	assert.Equal(t, 1, tp.GetStockCount())

	waste := []*Card{NewCard(CardDesignHeart, 2, true)}
	tp.SetWaste(waste)
	assert.Len(t, tp.GetWaste(), 1)

	var layout [TriPeaksRowCnt][TriPeaksColCnt]*TriPeaksCard
	layout[0][0] = &TriPeaksCard{Card: NewCard(CardDesignClover, 3, true), Removed: false}
	tp.SetLayout(layout)
	assert.NotNil(t, tp.GetLayout()[0][0])
}

func TestTriPeaks_ExposureChain(t *testing.T) {
	// Test that removing row 3 cards exposes row 2, and so on up the peaks
	tp := newTestTriPeaks()
	tp.Reset()

	// Row 2, col 0 should NOT be exposed (children at row 3: col 0 and col 1)
	assert.False(t, tp.IsExposed(2, 0))

	// Remove both children
	tp.layout[3][0].Removed = true
	tp.layout[3][1].Removed = true
	assert.True(t, tp.IsExposed(2, 0))
}

func TestTriPeaks_TriPeaksChildren(t *testing.T) {
	// Row 3 has no children
	children := triPeaksChildren(3, 0)
	assert.Nil(t, children)

	// Row 0, col 0 has children at (1,0) and (1,1)
	children = triPeaksChildren(0, 0)
	assert.Equal(t, [][2]int{{1, 0}, {1, 1}}, children)

	// Row 0, col 3 has children at (1,3) and (1,4)
	children = triPeaksChildren(0, 3)
	assert.Equal(t, [][2]int{{1, 3}, {1, 4}}, children)
}

func TestTriPeaks_ValidPos(t *testing.T) {
	// Row 0: only cols 0, 3, 6
	assert.True(t, triPeaksValidPos[0][0])
	assert.False(t, triPeaksValidPos[0][1])
	assert.False(t, triPeaksValidPos[0][2])
	assert.True(t, triPeaksValidPos[0][3])
	assert.False(t, triPeaksValidPos[0][4])
	assert.True(t, triPeaksValidPos[0][6])

	// Row 3: all 10 cols
	for c := range TriPeaksColCnt {
		assert.True(t, triPeaksValidPos[3][c])
	}
}

func TestTriPeaks_UndoDraw(t *testing.T) {
	tp := newTestTriPeaks()
	tp.Reset()
	initialStock := tp.GetStockCount()

	err := tp.Draw()
	require.NoError(t, err)

	err = tp.Undo()
	require.NoError(t, err)
	assert.Equal(t, initialStock, tp.GetStockCount())
	assert.Equal(t, 0, tp.GetMoveCount())
}

func TestTriPeaks_StalemateAfterDraw(t *testing.T) {
	tp := setupTriPeaksForRemove(t)
	// Non-adjacent card, stock has 1 card
	tp.layout[3][0] = &TriPeaksCard{Card: NewCard(CardDesignSpade, 9, true), Removed: false}
	tp.stock = []*Card{NewCard(CardDesignDiamond, 11, true)} // J is not adjacent to 9

	// After draw, if still no moves and stock empty, stalemate
	err := tp.Draw()
	require.NoError(t, err)
	// waste top is J(11), card is 9 — not adjacent, and stock is empty
	assert.True(t, tp.IsStalemate())
}

// --- UndoToEscape / UndoN tests ---

func TestTriPeaks_UndoToEscape_NotInStalemate(t *testing.T) {
	tp := newTestTriPeaks()
	tp.Reset()
	assert.Equal(t, 0, tp.UndoToEscape())
}

func TestTriPeaks_UndoToEscape_StalemateNoHistory(t *testing.T) {
	tp := newTestTriPeaks()
	tp.Reset()
	tp.SetIsStalemate(true)
	assert.Equal(t, -1, tp.UndoToEscape())
}

func TestTriPeaks_UndoToEscape_StalemateWithEscape(t *testing.T) {
	tp := newTestTriPeaks()
	tp.Reset()
	_ = tp.Draw()
	tp.SetIsStalemate(true)
	n := tp.UndoToEscape()
	assert.Equal(t, 1, n)
}

func TestTriPeaks_UndoN_Zero(t *testing.T) {
	tp := newTestTriPeaks()
	tp.Reset()
	err := tp.UndoN(0)
	assert.NoError(t, err)
}

func TestTriPeaks_UndoN_Valid(t *testing.T) {
	tp := newTestTriPeaks()
	tp.Reset()
	_ = tp.Draw()
	_ = tp.Draw()
	err := tp.UndoN(2)
	assert.NoError(t, err)
}

func TestTriPeaks_UndoN_Excessive(t *testing.T) {
	tp := newTestTriPeaks()
	tp.Reset()
	_ = tp.Draw()
	err := tp.UndoN(5)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "undo step")
}

// **得点の式はフロント (frontend/src/utils/tripeaksScore.ts) と同じでなければ
// ならない。** これまで計算はフロントにしか無く、CUI からは触れなかった (#5511)。
//
//	n 手目の連続除去 = n × TriPeaksPointsPerChain
//	山を 1 つ出し切るごとに + TriPeaksPeakBonus
//	山札を引く / アンドゥ → 連鎖は 0、得点は据え置き
func TestTriPeaks_ScoreMatchesTheFrontendFormula(t *testing.T) {
	// 盤面を作り込む: 3 列だけ使い、左の山を 1 枚で構成する。
	newBoard := func() *TriPeaks {
		tp := NewDefaultTriPeaks()
		tp.Reset()
		var layout [TriPeaksRowCnt][TriPeaksColCnt]*TriPeaksCard
		// 左の山 (col<3) に 1 枚、中央 (3<=col<6) に 2 枚。
		layout[TriPeaksRowCnt-1][0] = &TriPeaksCard{Card: NewCard(CardDesignSpade, 5, true)}
		layout[TriPeaksRowCnt-1][3] = &TriPeaksCard{Card: NewCard(CardDesignHeart, 6, true)}
		layout[TriPeaksRowCnt-1][4] = &TriPeaksCard{Card: NewCard(CardDesignClover, 7, true)}
		tp.SetLayout(layout)
		tp.SetWaste([]*Card{NewCard(CardDesignDiamond, 4, true)})
		tp.SetStock([]*Card{NewCard(CardDesignSpade, 9, true)})
		tp.SetPhase(TriPeaksPhasePlaying)
		return tp
	}

	t.Run("a fresh board scores nothing", func(t *testing.T) {
		tp := newBoard()
		if tp.GetScore() != 0 || tp.GetCombo() != 0 {
			t.Errorf("fresh board = (%d, %d), want (0, 0)", tp.GetScore(), tp.GetCombo())
		}
	})

	t.Run("the chain multiplies and clearing a peak pays a bonus", func(t *testing.T) {
		tp := newBoard()
		// 1 手目: 5 を除去。左の山 (1 枚だけ) が空になるのでボーナスも付く。
		if err := tp.Remove(TriPeaksRowCnt-1, 0); err != nil {
			t.Fatalf("remove 1: %v", err)
		}
		want := 1*TriPeaksPointsPerChain + TriPeaksPeakBonus
		if tp.GetScore() != want {
			t.Errorf("after 1 removal score = %d, want %d", tp.GetScore(), want)
		}
		if tp.GetCombo() != 1 {
			t.Errorf("chain = %d, want 1", tp.GetCombo())
		}

		// 2 手目: 6 を除去 (waste は 5)。連鎖 2 なので 200 点、山は空かない。
		if err := tp.Remove(TriPeaksRowCnt-1, 3); err != nil {
			t.Fatalf("remove 2: %v", err)
		}
		want += 2 * TriPeaksPointsPerChain
		if tp.GetScore() != want {
			t.Errorf("after 2 removals score = %d, want %d", tp.GetScore(), want)
		}
		if tp.GetCombo() != 2 {
			t.Errorf("chain = %d, want 2", tp.GetCombo())
		}
	})

	t.Run("drawing breaks the chain but keeps the score", func(t *testing.T) {
		tp := newBoard()
		if err := tp.Remove(TriPeaksRowCnt-1, 0); err != nil {
			t.Fatalf("remove: %v", err)
		}
		before := tp.GetScore()
		if err := tp.Draw(); err != nil {
			t.Fatalf("draw: %v", err)
		}
		if tp.GetCombo() != 0 {
			t.Errorf("chain after draw = %d, want 0", tp.GetCombo())
		}
		if tp.GetScore() != before {
			t.Errorf("score after draw = %d, want %d (draws must not cost points)", tp.GetScore(), before)
		}
	})

	// **アンドゥで点は戻らない。** 戻すと、除去→アンドゥ→除去で稼ぎ直せてしまう。
	t.Run("undo breaks the chain but keeps the score", func(t *testing.T) {
		tp := newBoard()
		if err := tp.Remove(TriPeaksRowCnt-1, 0); err != nil {
			t.Fatalf("remove: %v", err)
		}
		before := tp.GetScore()
		if err := tp.Undo(); err != nil {
			t.Fatalf("undo: %v", err)
		}
		if tp.GetCombo() != 0 {
			t.Errorf("chain after undo = %d, want 0", tp.GetCombo())
		}
		if tp.GetScore() != before {
			t.Errorf("score after undo = %d, want %d", tp.GetScore(), before)
		}
	})

	t.Run("reset clears both", func(t *testing.T) {
		tp := newBoard()
		if err := tp.Remove(TriPeaksRowCnt-1, 0); err != nil {
			t.Fatalf("remove: %v", err)
		}
		tp.Reset()
		if tp.GetScore() != 0 || tp.GetCombo() != 0 {
			t.Errorf("after reset = (%d, %d), want (0, 0)", tp.GetScore(), tp.GetCombo())
		}
	})

	// **KV 往復で消えないこと。** 決着の表示は保存のあとのリクエストで読まれる。
	t.Run("the score survives the snapshot", func(t *testing.T) {
		tp := newBoard()
		if err := tp.Remove(TriPeaksRowCnt-1, 0); err != nil {
			t.Fatalf("remove: %v", err)
		}
		data, err := json.Marshal(tp)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		var back TriPeaks
		if err := json.Unmarshal(data, &back); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if back.GetScore() != tp.GetScore() || back.GetCombo() != tp.GetCombo() {
			t.Errorf("restored = (%d, %d), want (%d, %d)",
				back.GetScore(), back.GetCombo(), tp.GetScore(), tp.GetCombo())
		}
	})
}
