//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestGolf creates a Golf with a fresh deck for testing.
func newTestGolf() *Golf {
	return NewGolf(NewTrumpCards(0))
}

// setupGolfForRemove sets up a Golf game with a known layout for testing Remove.
// Places a 5♠ at col 0 (row 4, bottom), and sets waste top to 4♥.
func setupGolfForRemove(t *testing.T) *Golf {
	t.Helper()
	g := newTestGolf()
	g.phase = GolfPhasePlaying

	// Clear layout
	for c := range GolfColCnt {
		for r := range GolfRowCnt {
			g.layout[c][r] = nil
		}
	}

	// Place card at col 0, row 4 (bottom of column)
	g.layout[0][4] = &GolfCard{
		Card:    NewCard(CardDesignSpade, 5, true),
		Removed: false,
	}

	// Waste top: 4♥ (adjacent to 5)
	g.waste = []*Card{NewCard(CardDesignHeart, 4, true)}
	g.stock = []*Card{NewCard(CardDesignDiamond, 7, true)}
	return g
}

func TestGolf_NewGolf(t *testing.T) {
	g := NewGolf(NewTrumpCards(0))
	assert.NotNil(t, g)
	assert.Equal(t, GolfPhase(0), g.GetPhase())
}

func TestGolf_Reset(t *testing.T) {
	g := newTestGolf()
	g.Reset()

	assert.Equal(t, GolfPhasePlaying, g.GetPhase())
	assert.Equal(t, 0, g.GetMoveCount())
	assert.False(t, g.IsStalemate())

	// Count tableau cards
	tableauCount := 0
	for c := range GolfColCnt {
		for r := range GolfRowCnt {
			if g.layout[c][r] != nil {
				tableauCount++
			}
		}
	}
	assert.Equal(t, GolfTableauCnt, tableauCount)

	// Stock + waste should account for remaining cards (52 - 35 = 17; 16 stock + 1 waste)
	assert.Equal(t, 16, g.GetStockCount())
	assert.Len(t, g.GetWaste(), 1)
}

func TestGolf_Reset_AllPositionsFilled(t *testing.T) {
	g := newTestGolf()
	g.Reset()

	// All 7×5 positions should have cards
	for c := range GolfColCnt {
		for r := range GolfRowCnt {
			assert.NotNil(t, g.layout[c][r], "expected card at col=%d row=%d", c, r)
		}
	}
}

func TestGolf_Draw(t *testing.T) {
	g := newTestGolf()
	g.Reset()
	initialStock := g.GetStockCount()
	initialWaste := len(g.GetWaste())

	err := g.Draw()
	require.NoError(t, err)
	assert.Equal(t, initialStock-1, g.GetStockCount())
	assert.Len(t, g.GetWaste(), initialWaste+1)
	assert.Equal(t, 1, g.GetMoveCount())
}

func TestGolf_Draw_EmptyStock(t *testing.T) {
	g := setupGolfForRemove(t)
	g.stock = nil

	err := g.Draw()
	assert.EqualError(t, err, "no cards in stock")
}

func TestGolf_Draw_NotPlaying(t *testing.T) {
	g := newTestGolf()
	g.SetPhase(GolfPhaseGameOver)

	err := g.Draw()
	assert.EqualError(t, err, "game is not in playing phase")
}

func TestGolf_Remove_Success(t *testing.T) {
	g := setupGolfForRemove(t)

	err := g.Remove(0)
	require.NoError(t, err)
	assert.True(t, g.layout[0][4].Removed)
	assert.Equal(t, 1, g.GetMoveCount())
	// Removed card goes to waste
	assert.Equal(t, 5, g.GetWaste()[len(g.GetWaste())-1].GetValue())
}

func TestGolf_Remove_KAWrap(t *testing.T) {
	g := setupGolfForRemove(t)
	// Place K at col 0, waste top is A
	g.layout[0][4] = &GolfCard{Card: NewCard(CardDesignSpade, 13, true), Removed: false}
	g.waste = []*Card{NewCard(CardDesignHeart, 1, true)}

	err := g.Remove(0)
	require.NoError(t, err)
	assert.True(t, g.layout[0][4].Removed)
}

func TestGolf_Remove_AKWrap(t *testing.T) {
	g := setupGolfForRemove(t)
	// Place A at col 0, waste top is K
	g.layout[0][4] = &GolfCard{Card: NewCard(CardDesignSpade, 1, true), Removed: false}
	g.waste = []*Card{NewCard(CardDesignHeart, 13, true)}

	err := g.Remove(0)
	require.NoError(t, err)
	assert.True(t, g.layout[0][4].Removed)
}

func TestGolf_Remove_NotAdjacent(t *testing.T) {
	g := setupGolfForRemove(t)
	// Place 9♠ at col 0, waste top is 4♥ — not adjacent
	g.layout[0][4] = &GolfCard{Card: NewCard(CardDesignSpade, 9, true), Removed: false}

	err := g.Remove(0)
	assert.EqualError(t, err, "card is not adjacent to waste top")
}

func TestGolf_Remove_NotExposed(t *testing.T) {
	g := setupGolfForRemove(t)
	// Place card at row 3 (above row 4), row 4 is not removed
	g.layout[0][3] = &GolfCard{Card: NewCard(CardDesignSpade, 3, true), Removed: false}
	// Row 4 card is the exposed one, not row 3
	// Trying to remove from col 0 should remove row 4 card, not row 3
	// This verifies that Remove always picks the exposed (bottom-most) card
	err := g.Remove(0)
	require.NoError(t, err)
	assert.True(t, g.layout[0][4].Removed)
	assert.False(t, g.layout[0][3].Removed)
}

func TestGolf_Remove_EmptyColumn(t *testing.T) {
	g := setupGolfForRemove(t)
	// Remove the only card in col 0
	g.layout[0][4].Removed = true

	err := g.Remove(0)
	assert.EqualError(t, err, "no card in column")
}

func TestGolf_Remove_EmptyWaste(t *testing.T) {
	g := setupGolfForRemove(t)
	g.waste = nil

	err := g.Remove(0)
	assert.EqualError(t, err, "waste is empty")
}

func TestGolf_Remove_InvalidCol(t *testing.T) {
	g := setupGolfForRemove(t)

	err := g.Remove(-1)
	assert.EqualError(t, err, "invalid column")

	err = g.Remove(7)
	assert.EqualError(t, err, "invalid column")
}

func TestGolf_Remove_NotPlaying(t *testing.T) {
	g := setupGolfForRemove(t)
	g.phase = GolfPhaseGameOver

	err := g.Remove(0)
	assert.EqualError(t, err, "game is not in playing phase")
}

func TestGolf_GiveUp(t *testing.T) {
	g := newTestGolf()
	g.Reset()
	g.GiveUp()
	assert.Equal(t, GolfPhaseGameOver, g.GetPhase())
}

func TestGolf_GiveUp_NotPlaying(t *testing.T) {
	g := newTestGolf()
	g.SetPhase(GolfPhaseGameClear)
	g.GiveUp()
	// Phase should not change
	assert.Equal(t, GolfPhaseGameClear, g.GetPhase())
}

func TestGolf_GetHint_RemoveAvailable(t *testing.T) {
	g := setupGolfForRemove(t)

	hint := g.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "remove", hint.Type)
	assert.Equal(t, 0, hint.Col)
}

func TestGolf_GetHint_DrawOnly(t *testing.T) {
	g := setupGolfForRemove(t)
	// Make card not adjacent to waste
	g.layout[0][4] = &GolfCard{Card: NewCard(CardDesignSpade, 9, true), Removed: false}

	hint := g.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "draw", hint.Type)
	assert.Equal(t, -1, hint.Col)
}

func TestGolf_GetHint_NoHint(t *testing.T) {
	g := setupGolfForRemove(t)
	g.layout[0][4] = &GolfCard{Card: NewCard(CardDesignSpade, 9, true), Removed: false}
	g.stock = nil

	hint := g.GetHint()
	assert.Nil(t, hint)
}

func TestGolf_GetHint_NotPlaying(t *testing.T) {
	g := newTestGolf()
	g.SetPhase(GolfPhaseGameOver)

	hint := g.GetHint()
	assert.Nil(t, hint)
}

func TestGolf_GetHint_EmptyWaste(t *testing.T) {
	g := setupGolfForRemove(t)
	g.waste = nil

	hint := g.GetHint()
	assert.Nil(t, hint)
}

func TestGolf_Undo(t *testing.T) {
	g := setupGolfForRemove(t)
	// Add a second card so removing one does NOT trigger game clear
	g.layout[1][4] = &GolfCard{Card: NewCard(CardDesignHeart, 6, true), Removed: false}

	// Remove a card, then undo
	err := g.Remove(0)
	require.NoError(t, err)
	assert.True(t, g.layout[0][4].Removed)
	assert.Equal(t, GolfPhasePlaying, g.GetPhase())

	err = g.Undo()
	require.NoError(t, err)
	assert.False(t, g.layout[0][4].Removed)
	assert.Equal(t, 0, g.GetMoveCount())
}

func TestGolf_Undo_NoHistory(t *testing.T) {
	g := newTestGolf()
	g.Reset()

	err := g.Undo()
	assert.EqualError(t, err, "cannot undo: no history")
}

func TestGolf_Undo_NotPlaying(t *testing.T) {
	g := newTestGolf()
	g.SetPhase(GolfPhaseGameOver)

	err := g.Undo()
	assert.EqualError(t, err, "cannot undo: game is not in playing phase")
}

func TestGolf_CanUndo(t *testing.T) {
	g := newTestGolf()
	g.Reset()
	assert.False(t, g.CanUndo())

	_ = g.Draw()
	assert.True(t, g.CanUndo())

	g.SetPhase(GolfPhaseGameOver)
	assert.False(t, g.CanUndo())
}

func TestGolf_IsExposed(t *testing.T) {
	g := setupGolfForRemove(t)

	// Bottom card is exposed
	assert.True(t, g.IsExposed(0, 4))

	// Add card above — it should NOT be exposed since row 4 is still present
	g.layout[0][3] = &GolfCard{Card: NewCard(CardDesignHeart, 3, true), Removed: false}
	assert.False(t, g.IsExposed(0, 3))

	// Remove bottom card — row 3 becomes exposed
	g.layout[0][4].Removed = true
	assert.True(t, g.IsExposed(0, 3))
}

func TestGolf_IsExposed_InvalidPos(t *testing.T) {
	g := newTestGolf()
	g.Reset()
	assert.False(t, g.IsExposed(-1, 0))
	assert.False(t, g.IsExposed(0, -1))
	assert.False(t, g.IsExposed(7, 0))
	assert.False(t, g.IsExposed(0, 5))
}

func TestGolf_IsExposed_RemovedCard(t *testing.T) {
	g := setupGolfForRemove(t)
	g.layout[0][4].Removed = true
	assert.False(t, g.IsExposed(0, 4))
}

func TestGolf_AllRemoved(t *testing.T) {
	g := setupGolfForRemove(t)
	assert.False(t, g.AllRemoved())

	// Remove all cards
	g.layout[0][4].Removed = true
	assert.True(t, g.AllRemoved())
}

func TestGolf_GameClear(t *testing.T) {
	g := setupGolfForRemove(t)

	// Remove the only card — should trigger game clear
	err := g.Remove(0)
	require.NoError(t, err)
	assert.Equal(t, GolfPhaseGameClear, g.GetPhase())
}

func TestGolf_Stalemate(t *testing.T) {
	g := setupGolfForRemove(t)
	// Make card not adjacent, empty stock
	g.layout[0][4] = &GolfCard{Card: NewCard(CardDesignSpade, 9, true), Removed: false}
	g.stock = nil

	// Manually trigger stalemate check
	g.checkStalemate()
	assert.True(t, g.IsStalemate())
}

func TestGolf_Stalemate_NotPlayingSkips(t *testing.T) {
	g := newTestGolf()
	g.SetPhase(GolfPhaseGameClear)
	g.checkStalemate()
	assert.False(t, g.IsStalemate())
}

func TestGolf_IsAdjacentRank(t *testing.T) {
	g := newTestGolf()
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
		assert.Equal(t, tt.want, g.isAdjacentRank(c1, c2),
			"isAdjacentRank(%d, %d) = %v, want %v", tt.v1, tt.v2, !tt.want, tt.want)
	}
}

func TestGolf_JSON_RoundTrip(t *testing.T) {
	g := newTestGolf()
	g.Reset()
	_ = g.Draw()

	data, err := json.Marshal(g)
	require.NoError(t, err)

	var g2 Golf
	err = json.Unmarshal(data, &g2)
	require.NoError(t, err)

	assert.Equal(t, g.GetPhase(), g2.GetPhase())
	assert.Equal(t, g.GetMoveCount(), g2.GetMoveCount())
	assert.Equal(t, g.GetStockCount(), g2.GetStockCount())
	assert.Equal(t, len(g.GetWaste()), len(g2.GetWaste()))
	assert.Equal(t, g.IsStalemate(), g2.IsStalemate())
}

func TestGolf_JSON_MaxSliceLen(t *testing.T) {
	// Create JSON with oversized stock array
	bigStock := make([]*Card, golfMaxSliceLen+1)
	j := golfJSON{
		Stock: bigStock,
	}
	data, err := json.Marshal(j)
	require.NoError(t, err)

	var g Golf
	err = json.Unmarshal(data, &g)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum allowed size")
}

func TestGolf_JSON_NilFields(t *testing.T) {
	// Unmarshal with nil stock/waste/actionLog
	j := golfJSON{
		Phase: GolfPhasePlaying,
	}
	data, err := json.Marshal(j)
	require.NoError(t, err)

	var g Golf
	err = json.Unmarshal(data, &g)
	require.NoError(t, err)
	assert.NotNil(t, g.stock)
	assert.NotNil(t, g.waste)
	assert.NotNil(t, g.actionLog)
	assert.NotNil(t, g.trumpCards)
}

func TestGolf_GetLayout(t *testing.T) {
	g := newTestGolf()
	g.Reset()
	layout := g.GetLayout()
	// Col 0, row 0 should have a card
	assert.NotNil(t, layout[0][0])
}

func TestGolf_GetActionLog(t *testing.T) {
	g := newTestGolf()
	g.Reset()
	assert.Nil(t, g.GetActionLog())

	_ = g.Draw()
	assert.Len(t, g.GetActionLog(), 1)
}

func TestGolf_SettersForTest(t *testing.T) {
	g := newTestGolf()
	g.SetPhase(GolfPhaseGameClear)
	assert.Equal(t, GolfPhaseGameClear, g.GetPhase())

	g.SetIsStalemate(true)
	assert.True(t, g.IsStalemate())

	stock := []*Card{NewCard(CardDesignSpade, 1, true)}
	g.SetStock(stock)
	assert.Equal(t, 1, g.GetStockCount())

	waste := []*Card{NewCard(CardDesignHeart, 2, true)}
	g.SetWaste(waste)
	assert.Len(t, g.GetWaste(), 1)

	var layout [GolfColCnt][GolfRowCnt]*GolfCard
	layout[0][0] = &GolfCard{Card: NewCard(CardDesignClover, 3, true), Removed: false}
	g.SetLayout(layout)
	assert.NotNil(t, g.GetLayout()[0][0])
}

func TestGolf_ExposureChain(t *testing.T) {
	// Test that removing bottom cards exposes cards above
	g := newTestGolf()
	g.Reset()

	// Row 3 should NOT be exposed (row 4 is below it)
	assert.False(t, g.IsExposed(0, 3))

	// Remove row 4 — row 3 becomes exposed
	g.layout[0][4].Removed = true
	assert.True(t, g.IsExposed(0, 3))

	// Remove row 3 — row 2 becomes exposed
	g.layout[0][3].Removed = true
	assert.True(t, g.IsExposed(0, 2))
}

func TestGolf_UndoDraw(t *testing.T) {
	g := newTestGolf()
	g.Reset()
	initialStock := g.GetStockCount()

	err := g.Draw()
	require.NoError(t, err)

	err = g.Undo()
	require.NoError(t, err)
	assert.Equal(t, initialStock, g.GetStockCount())
	assert.Equal(t, 0, g.GetMoveCount())
}

func TestGolf_StalemateAfterDraw(t *testing.T) {
	g := setupGolfForRemove(t)
	// Non-adjacent card, stock has 1 card
	g.layout[0][4] = &GolfCard{Card: NewCard(CardDesignSpade, 9, true), Removed: false}
	g.stock = []*Card{NewCard(CardDesignDiamond, 11, true)} // J is not adjacent to 9

	// After draw, if still no moves and stock empty, stalemate
	err := g.Draw()
	require.NoError(t, err)
	// waste top is J(11), card is 9 — not adjacent, and stock is empty
	assert.True(t, g.IsStalemate())
}

func TestGolf_FindExposedRow(t *testing.T) {
	g := setupGolfForRemove(t)

	// Col 0 has card at row 4
	assert.Equal(t, 4, g.findExposedRow(0))

	// Empty column
	g.layout[0][4].Removed = true
	assert.Equal(t, -1, g.findExposedRow(0))

	// Multiple cards in column — exposed is bottom-most non-removed
	g.layout[0][4].Removed = false
	g.layout[0][3] = &GolfCard{Card: NewCard(CardDesignHeart, 3, true), Removed: false}
	assert.Equal(t, 4, g.findExposedRow(0))

	// Remove bottom, row 3 becomes exposed
	g.layout[0][4].Removed = true
	assert.Equal(t, 3, g.findExposedRow(0))
}

// --- UndoToEscape / UndoN tests ---

func TestGolf_UndoToEscape_NotInStalemate(t *testing.T) {
	g := newTestGolf()
	g.Reset()
	assert.Equal(t, 0, g.UndoToEscape())
}

func TestGolf_UndoToEscape_StalemateNoHistory(t *testing.T) {
	g := newTestGolf()
	g.Reset()
	g.SetIsStalemate(true)
	assert.Equal(t, -1, g.UndoToEscape())
}

func TestGolf_UndoToEscape_StalemateWithEscape(t *testing.T) {
	g := newTestGolf()
	g.Reset()
	_ = g.Draw()
	g.SetIsStalemate(true)
	n := g.UndoToEscape()
	assert.Equal(t, 1, n)
}

func TestGolf_UndoN_Zero(t *testing.T) {
	g := newTestGolf()
	g.Reset()
	err := g.UndoN(0)
	assert.NoError(t, err)
}

func TestGolf_UndoN_Valid(t *testing.T) {
	g := newTestGolf()
	g.Reset()
	_ = g.Draw()
	_ = g.Draw()
	err := g.UndoN(2)
	assert.NoError(t, err)
}

func TestGolf_UndoN_Excessive(t *testing.T) {
	g := newTestGolf()
	g.Reset()
	_ = g.Draw()
	err := g.UndoN(5)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "undo step")
}
