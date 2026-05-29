//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestAcesUp creates an AcesUp with a fresh deck for testing.
func newTestAcesUp() *AcesUp {
	return NewAcesUp(NewTrumpCards(0))
}

// setupAcesUpPlaying returns a playing game with explicit columns and empty stock.
func setupAcesUpPlaying(t *testing.T, columns [AcesUpColCnt][]*Card) *AcesUp {
	t.Helper()
	a := newTestAcesUp()
	a.phase = AcesUpPhasePlaying
	a.columns = columns
	a.stock = nil
	a.discard = nil
	return a
}

func auCard(design, value int) *Card { return NewCard(design, value, true) }

func TestAcesUp_NewAcesUp(t *testing.T) {
	a := NewAcesUp(NewTrumpCards(0))
	assert.NotNil(t, a)
	assert.Equal(t, AcesUpPhase(0), a.GetPhase())
}

func TestAcesUp_NewDefaultAcesUp(t *testing.T) {
	a := NewDefaultAcesUp()
	assert.NotNil(t, a)
}

func TestAcesUp_Reset(t *testing.T) {
	a := newTestAcesUp()
	a.Reset()

	assert.Equal(t, AcesUpPhasePlaying, a.GetPhase())
	assert.Equal(t, 0, a.GetMoveCount())
	assert.False(t, a.IsStalemate())

	// 4 columns, 1 card each
	cols := a.GetColumns()
	total := 0
	for c := range AcesUpColCnt {
		assert.Len(t, cols[c], 1, "column %d should have 1 card", c)
		total += len(cols[c])
	}
	assert.Equal(t, AcesUpColCnt, total)
	// 52 - 4 = 48 in stock
	assert.Equal(t, 48, a.GetStockCount())
	assert.Equal(t, 0, a.GetDiscardCount())
}

func TestAcesUp_acesUpRank(t *testing.T) {
	assert.Equal(t, acesUpAceRank, acesUpRank(auCard(CardDesignSpade, 1)))
	assert.Equal(t, 13, acesUpRank(auCard(CardDesignSpade, 13)))
	assert.Equal(t, 7, acesUpRank(auCard(CardDesignSpade, 7)))
	// Ace outranks King
	assert.Greater(t, acesUpRank(auCard(CardDesignSpade, 1)), acesUpRank(auCard(CardDesignSpade, 13)))
}

func TestAcesUp_Draw(t *testing.T) {
	a := setupAcesUpPlaying(t, [AcesUpColCnt][]*Card{
		{auCard(CardDesignSpade, 3)},
		{auCard(CardDesignHeart, 4)},
		{auCard(CardDesignClover, 5)},
		{auCard(CardDesignDiamond, 6)},
	})
	a.stock = []*Card{
		auCard(CardDesignSpade, 7), auCard(CardDesignHeart, 8),
		auCard(CardDesignClover, 9), auCard(CardDesignDiamond, 10),
	}

	require.NoError(t, a.Draw())
	assert.Equal(t, 0, a.GetStockCount())
	assert.Equal(t, 1, a.GetMoveCount())
	cols := a.GetColumns()
	for c := range AcesUpColCnt {
		assert.Len(t, cols[c], 2)
	}
	assert.Len(t, a.GetActionLog(), 1)
}

func TestAcesUp_Draw_EmptyStock(t *testing.T) {
	a := setupAcesUpPlaying(t, [AcesUpColCnt][]*Card{{auCard(CardDesignSpade, 3)}})
	assert.EqualError(t, a.Draw(), "no cards in stock")
}

func TestAcesUp_Draw_NotPlaying(t *testing.T) {
	a := newTestAcesUp()
	a.SetPhase(AcesUpPhaseGameOver)
	assert.EqualError(t, a.Draw(), "game is not in playing phase")
}

func TestAcesUp_Draw_PartialStock(t *testing.T) {
	a := setupAcesUpPlaying(t, [AcesUpColCnt][]*Card{
		{auCard(CardDesignSpade, 3)}, {auCard(CardDesignHeart, 4)},
		{auCard(CardDesignClover, 5)}, {auCard(CardDesignDiamond, 6)},
	})
	a.stock = []*Card{auCard(CardDesignSpade, 7), auCard(CardDesignHeart, 8)}

	require.NoError(t, a.Draw())
	assert.Equal(t, 0, a.GetStockCount())
	cols := a.GetColumns()
	assert.Len(t, cols[0], 2)
	assert.Len(t, cols[1], 2)
	assert.Len(t, cols[2], 1)
	assert.Len(t, cols[3], 1)
}

func TestAcesUp_Remove_Success(t *testing.T) {
	a := setupAcesUpPlaying(t, [AcesUpColCnt][]*Card{
		{auCard(CardDesignSpade, 5)},
		{auCard(CardDesignSpade, 9)},
		{auCard(CardDesignHeart, 4)},
		{auCard(CardDesignDiamond, 6)},
	})

	require.NoError(t, a.Remove(0))
	assert.Equal(t, 0, len(a.GetColumns()[0]))
	assert.Equal(t, 1, a.GetDiscardCount())
	assert.Equal(t, 1, a.GetMoveCount())
}

func TestAcesUp_Remove_AceNotRemovable(t *testing.T) {
	a := setupAcesUpPlaying(t, [AcesUpColCnt][]*Card{
		{auCard(CardDesignSpade, 1)},
		{auCard(CardDesignSpade, 13)},
		{auCard(CardDesignHeart, 4)},
		{auCard(CardDesignDiamond, 6)},
	})
	// Ace is highest; K cannot remove A
	assert.EqualError(t, a.Remove(0), "card is not removable")
	// but the King can be removed by the Ace
	require.NoError(t, a.Remove(1))
}

func TestAcesUp_Remove_NotRemovable(t *testing.T) {
	a := setupAcesUpPlaying(t, [AcesUpColCnt][]*Card{
		{auCard(CardDesignSpade, 9)},
		{auCard(CardDesignHeart, 4)},
		{auCard(CardDesignClover, 5)},
		{auCard(CardDesignDiamond, 6)},
	})
	// no other spade higher than 9
	assert.EqualError(t, a.Remove(0), "card is not removable")
}

func TestAcesUp_Remove_EmptyColumn(t *testing.T) {
	a := setupAcesUpPlaying(t, [AcesUpColCnt][]*Card{
		{}, {auCard(CardDesignHeart, 4)},
	})
	assert.EqualError(t, a.Remove(0), "no card in column")
}

func TestAcesUp_Remove_InvalidCol(t *testing.T) {
	a := setupAcesUpPlaying(t, [AcesUpColCnt][]*Card{{auCard(CardDesignSpade, 5)}})
	assert.EqualError(t, a.Remove(-1), "invalid column")
	assert.EqualError(t, a.Remove(AcesUpColCnt), "invalid column")
}

func TestAcesUp_Remove_NotPlaying(t *testing.T) {
	a := setupAcesUpPlaying(t, [AcesUpColCnt][]*Card{{auCard(CardDesignSpade, 5)}})
	a.phase = AcesUpPhaseGameOver
	assert.EqualError(t, a.Remove(0), "game is not in playing phase")
}

func TestAcesUp_Move_Success(t *testing.T) {
	a := setupAcesUpPlaying(t, [AcesUpColCnt][]*Card{
		{auCard(CardDesignSpade, 5), auCard(CardDesignHeart, 7)},
		{},
		{auCard(CardDesignClover, 5)},
		{auCard(CardDesignDiamond, 6)},
	})

	require.NoError(t, a.Move(0))
	assert.Equal(t, 1, len(a.GetColumns()[0]))
	assert.Equal(t, 1, len(a.GetColumns()[1]))
	assert.Equal(t, 7, a.GetColumns()[1][0].GetValue())
	assert.Equal(t, 1, a.GetMoveCount())
}

func TestAcesUp_Move_NoEmptyColumn(t *testing.T) {
	a := setupAcesUpPlaying(t, [AcesUpColCnt][]*Card{
		{auCard(CardDesignSpade, 5)},
		{auCard(CardDesignHeart, 7)},
		{auCard(CardDesignClover, 5)},
		{auCard(CardDesignDiamond, 6)},
	})
	assert.EqualError(t, a.Move(0), "no empty column")
}

func TestAcesUp_Move_EmptyColumn(t *testing.T) {
	a := setupAcesUpPlaying(t, [AcesUpColCnt][]*Card{{}, {}})
	assert.EqualError(t, a.Move(0), "no card in column")
}

func TestAcesUp_Move_InvalidCol(t *testing.T) {
	a := setupAcesUpPlaying(t, [AcesUpColCnt][]*Card{{auCard(CardDesignSpade, 5)}, {}})
	assert.EqualError(t, a.Move(-1), "invalid column")
	assert.EqualError(t, a.Move(AcesUpColCnt), "invalid column")
}

func TestAcesUp_Move_NotPlaying(t *testing.T) {
	a := setupAcesUpPlaying(t, [AcesUpColCnt][]*Card{{auCard(CardDesignSpade, 5)}, {}})
	a.phase = AcesUpPhaseGameClear
	assert.EqualError(t, a.Move(0), "game is not in playing phase")
}

func TestAcesUp_GiveUp(t *testing.T) {
	a := newTestAcesUp()
	a.Reset()
	a.GiveUp()
	assert.Equal(t, AcesUpPhaseGameOver, a.GetPhase())
}

func TestAcesUp_GiveUp_NotPlaying(t *testing.T) {
	a := newTestAcesUp()
	a.SetPhase(AcesUpPhaseGameClear)
	a.GiveUp()
	assert.Equal(t, AcesUpPhaseGameClear, a.GetPhase())
}

func TestAcesUp_GetHint_Remove(t *testing.T) {
	a := setupAcesUpPlaying(t, [AcesUpColCnt][]*Card{
		{auCard(CardDesignSpade, 5)},
		{auCard(CardDesignSpade, 9)},
		{auCard(CardDesignHeart, 4)},
		{auCard(CardDesignDiamond, 6)},
	})
	hint := a.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "remove", hint.Type)
	assert.Equal(t, 0, hint.Col)
}

func TestAcesUp_GetHint_Draw(t *testing.T) {
	a := setupAcesUpPlaying(t, [AcesUpColCnt][]*Card{
		{auCard(CardDesignSpade, 9)},
		{auCard(CardDesignHeart, 4)},
		{auCard(CardDesignClover, 5)},
		{auCard(CardDesignDiamond, 6)},
	})
	a.stock = []*Card{auCard(CardDesignSpade, 2)}
	hint := a.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "draw", hint.Type)
	assert.Equal(t, -1, hint.Col)
}

func TestAcesUp_GetHint_Move(t *testing.T) {
	a := setupAcesUpPlaying(t, [AcesUpColCnt][]*Card{
		{auCard(CardDesignHeart, 3)},
		{},
		{auCard(CardDesignSpade, 7), auCard(CardDesignDiamond, 2)},
		{auCard(CardDesignSpade, 5)},
	})
	hint := a.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "move", hint.Type)
	assert.Equal(t, 2, hint.Col)
}

func TestAcesUp_GetHint_None(t *testing.T) {
	a := setupAcesUpPlaying(t, [AcesUpColCnt][]*Card{
		{auCard(CardDesignSpade, 9)},
		{auCard(CardDesignHeart, 4)},
		{auCard(CardDesignClover, 5)},
		{auCard(CardDesignDiamond, 6)},
	})
	assert.Nil(t, a.GetHint())
}

func TestAcesUp_GetHint_NotPlaying(t *testing.T) {
	a := newTestAcesUp()
	a.SetPhase(AcesUpPhaseGameOver)
	assert.Nil(t, a.GetHint())
}

func TestAcesUp_Stalemate(t *testing.T) {
	a := setupAcesUpPlaying(t, [AcesUpColCnt][]*Card{
		{auCard(CardDesignSpade, 9)},
		{auCard(CardDesignHeart, 4)},
		{auCard(CardDesignClover, 5)},
		{auCard(CardDesignDiamond, 6)},
	})
	a.checkStalemate()
	assert.True(t, a.IsStalemate())
}

func TestAcesUp_Stalemate_NotPlayingSkips(t *testing.T) {
	a := newTestAcesUp()
	a.SetPhase(AcesUpPhaseGameClear)
	a.checkStalemate()
	assert.False(t, a.IsStalemate())
}

func TestAcesUp_GameClear_ViaMove(t *testing.T) {
	a := setupAcesUpPlaying(t, [AcesUpColCnt][]*Card{
		{auCard(CardDesignSpade, 1), auCard(CardDesignHeart, 1)},
		{auCard(CardDesignDiamond, 1)},
		{auCard(CardDesignClover, 1)},
		{},
	})
	require.NoError(t, a.Move(0))
	assert.Equal(t, AcesUpPhaseGameClear, a.GetPhase())
}

func TestAcesUp_isWon(t *testing.T) {
	a := setupAcesUpPlaying(t, [AcesUpColCnt][]*Card{
		{auCard(CardDesignSpade, 1)},
		{auCard(CardDesignHeart, 1)},
		{auCard(CardDesignDiamond, 1)},
		{auCard(CardDesignClover, 1)},
	})
	assert.True(t, a.isWon())

	// non-ace breaks win
	a.columns[0] = []*Card{auCard(CardDesignSpade, 2)}
	assert.False(t, a.isWon())

	// more than one card breaks win
	a.columns[0] = []*Card{auCard(CardDesignSpade, 1), auCard(CardDesignSpade, 2)}
	assert.False(t, a.isWon())

	// non-empty stock breaks win
	a.columns[0] = []*Card{auCard(CardDesignSpade, 1)}
	a.stock = []*Card{auCard(CardDesignSpade, 2)}
	assert.False(t, a.isWon())
}

func TestAcesUp_CanRemove_CanMove(t *testing.T) {
	a := setupAcesUpPlaying(t, [AcesUpColCnt][]*Card{
		{auCard(CardDesignSpade, 5)},
		{auCard(CardDesignSpade, 9)},
		{},
		{auCard(CardDesignDiamond, 6)},
	})
	assert.True(t, a.CanRemove(0))
	assert.False(t, a.CanRemove(1))
	assert.False(t, a.CanRemove(2)) // empty
	assert.False(t, a.CanRemove(-1))
	assert.True(t, a.CanMove(0))
	assert.False(t, a.CanMove(2)) // empty source
	assert.False(t, a.CanMove(-1))
}

func TestAcesUp_CanMove_NoEmpty(t *testing.T) {
	a := setupAcesUpPlaying(t, [AcesUpColCnt][]*Card{
		{auCard(CardDesignSpade, 5)},
		{auCard(CardDesignSpade, 9)},
		{auCard(CardDesignClover, 5)},
		{auCard(CardDesignDiamond, 6)},
	})
	assert.False(t, a.CanMove(0))
}

func TestAcesUp_Undo_Remove(t *testing.T) {
	a := setupAcesUpPlaying(t, [AcesUpColCnt][]*Card{
		{auCard(CardDesignSpade, 5)},
		{auCard(CardDesignSpade, 9)},
		{auCard(CardDesignHeart, 4)},
		{auCard(CardDesignDiamond, 6)},
	})
	require.NoError(t, a.Remove(0))
	assert.Equal(t, 0, len(a.GetColumns()[0]))

	require.NoError(t, a.Undo())
	assert.Equal(t, 1, len(a.GetColumns()[0]))
	assert.Equal(t, 0, a.GetMoveCount())
	assert.Equal(t, 0, a.GetDiscardCount())
}

func TestAcesUp_Undo_Draw(t *testing.T) {
	a := setupAcesUpPlaying(t, [AcesUpColCnt][]*Card{
		{auCard(CardDesignSpade, 3)}, {auCard(CardDesignHeart, 4)},
		{auCard(CardDesignClover, 5)}, {auCard(CardDesignDiamond, 6)},
	})
	a.stock = []*Card{
		auCard(CardDesignSpade, 7), auCard(CardDesignHeart, 8),
		auCard(CardDesignClover, 9), auCard(CardDesignDiamond, 10),
	}
	require.NoError(t, a.Draw())
	require.NoError(t, a.Undo())
	assert.Equal(t, 4, a.GetStockCount())
	assert.Equal(t, 0, a.GetMoveCount())
}

func TestAcesUp_Undo_NoHistory(t *testing.T) {
	a := newTestAcesUp()
	a.Reset()
	assert.EqualError(t, a.Undo(), "cannot undo: no history")
}

func TestAcesUp_Undo_NotPlaying(t *testing.T) {
	a := newTestAcesUp()
	a.SetPhase(AcesUpPhaseGameOver)
	assert.EqualError(t, a.Undo(), "cannot undo: game is not in playing phase")
}

func TestAcesUp_CanUndo(t *testing.T) {
	a := setupAcesUpPlaying(t, [AcesUpColCnt][]*Card{
		{auCard(CardDesignSpade, 5)}, {auCard(CardDesignSpade, 9)},
		{auCard(CardDesignHeart, 4)}, {auCard(CardDesignDiamond, 6)},
	})
	assert.False(t, a.CanUndo())
	require.NoError(t, a.Remove(0))
	assert.True(t, a.CanUndo())
	a.SetPhase(AcesUpPhaseGameOver)
	assert.False(t, a.CanUndo())
}

func TestAcesUp_UndoToEscape_NotInStalemate(t *testing.T) {
	a := newTestAcesUp()
	a.Reset()
	assert.Equal(t, 0, a.UndoToEscape())
}

func TestAcesUp_UndoToEscape_NoHistory(t *testing.T) {
	a := newTestAcesUp()
	a.Reset()
	a.SetIsStalemate(true)
	assert.Equal(t, -1, a.UndoToEscape())
}

func TestAcesUp_UndoToEscape_WithEscape(t *testing.T) {
	a := setupAcesUpPlaying(t, [AcesUpColCnt][]*Card{
		{auCard(CardDesignSpade, 5)}, {auCard(CardDesignSpade, 9)},
		{auCard(CardDesignHeart, 4)}, {auCard(CardDesignDiamond, 6)},
	})
	require.NoError(t, a.Remove(0)) // history[0].isStalemate == false
	a.SetIsStalemate(true)
	assert.Equal(t, 1, a.UndoToEscape())
}

func TestAcesUp_UndoN(t *testing.T) {
	a := setupAcesUpPlaying(t, [AcesUpColCnt][]*Card{
		{auCard(CardDesignSpade, 5), auCard(CardDesignSpade, 2)},
		{auCard(CardDesignSpade, 9)},
		{auCard(CardDesignHeart, 4)},
		{auCard(CardDesignDiamond, 6)},
	})
	require.NoError(t, a.Remove(0)) // remove 2♠ (9♠ higher)
	require.NoError(t, a.Remove(0)) // remove 5♠ (9♠ higher)
	require.NoError(t, a.UndoN(2))
	assert.Equal(t, 2, len(a.GetColumns()[0]))
	assert.Equal(t, 0, a.GetMoveCount())

	assert.NoError(t, a.UndoN(0))
	assert.Error(t, a.UndoN(5))
}

func TestAcesUp_GetActionLog(t *testing.T) {
	a := newTestAcesUp()
	a.Reset()
	assert.Nil(t, a.GetActionLog())
}

func TestAcesUp_Setters(t *testing.T) {
	a := newTestAcesUp()
	a.SetPhase(AcesUpPhaseGameClear)
	assert.Equal(t, AcesUpPhaseGameClear, a.GetPhase())
	a.SetIsStalemate(true)
	assert.True(t, a.IsStalemate())
	a.SetStock([]*Card{auCard(CardDesignSpade, 1)})
	assert.Equal(t, 1, a.GetStockCount())
	var cols [AcesUpColCnt][]*Card
	cols[0] = []*Card{auCard(CardDesignClover, 3)}
	a.SetColumns(cols)
	assert.Equal(t, 1, len(a.GetColumns()[0]))
}

func TestAcesUp_JSON_RoundTrip(t *testing.T) {
	a := newTestAcesUp()
	a.Reset()
	_ = a.Draw()

	data, err := json.Marshal(a)
	require.NoError(t, err)

	var a2 AcesUp
	require.NoError(t, json.Unmarshal(data, &a2))
	assert.Equal(t, a.GetPhase(), a2.GetPhase())
	assert.Equal(t, a.GetMoveCount(), a2.GetMoveCount())
	assert.Equal(t, a.GetStockCount(), a2.GetStockCount())
	assert.Equal(t, a.GetDiscardCount(), a2.GetDiscardCount())
	assert.Equal(t, a.IsStalemate(), a2.IsStalemate())
}

func TestAcesUp_JSON_MaxSliceLen(t *testing.T) {
	j := acesUpJSON{Stock: make([]*Card, acesUpMaxSliceLen+1)}
	data, err := json.Marshal(j)
	require.NoError(t, err)

	var a AcesUp
	err = json.Unmarshal(data, &a)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum allowed size")
}

func TestAcesUp_JSON_MaxColumnLen(t *testing.T) {
	var j acesUpJSON
	j.Columns[0] = make([]*Card, acesUpMaxSliceLen+1)
	data, err := json.Marshal(j)
	require.NoError(t, err)

	var a AcesUp
	err = json.Unmarshal(data, &a)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "column exceeds maximum allowed size")
}

func TestAcesUp_JSON_NilFields(t *testing.T) {
	j := acesUpJSON{Phase: AcesUpPhasePlaying}
	data, err := json.Marshal(j)
	require.NoError(t, err)

	var a AcesUp
	require.NoError(t, json.Unmarshal(data, &a))
	assert.NotNil(t, a.stock)
	assert.NotNil(t, a.discard)
	assert.NotNil(t, a.actionLog)
	assert.NotNil(t, a.trumpCards)
	for c := range AcesUpColCnt {
		assert.NotNil(t, a.columns[c])
	}
}

func TestAcesUp_Snapshot_JSON_RoundTrip(t *testing.T) {
	a := setupAcesUpPlaying(t, [AcesUpColCnt][]*Card{
		{auCard(CardDesignSpade, 5)}, {auCard(CardDesignSpade, 9)},
		{auCard(CardDesignHeart, 4)}, {auCard(CardDesignDiamond, 6)},
	})
	require.NoError(t, a.Remove(0))

	data, err := json.Marshal(a)
	require.NoError(t, err)
	var a2 AcesUp
	require.NoError(t, json.Unmarshal(data, &a2))
	require.NoError(t, a2.Undo())
	assert.Equal(t, 1, len(a2.GetColumns()[0]))
}

func TestAcesUp_Snapshot_JSON_MaxSliceLen(t *testing.T) {
	j := acesUpSnapshotJSON{Stock: make([]*Card, acesUpMaxSliceLen+1)}
	data, err := json.Marshal(j)
	require.NoError(t, err)
	var s acesUpSnapshot
	err = json.Unmarshal(data, &s)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum allowed size")
}

func TestAcesUp_Snapshot_JSON_MaxColumnLen(t *testing.T) {
	var j acesUpSnapshotJSON
	j.Columns[0] = make([]*Card, acesUpMaxSliceLen+1)
	data, err := json.Marshal(j)
	require.NoError(t, err)
	var s acesUpSnapshot
	err = json.Unmarshal(data, &s)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "column exceeds maximum allowed size")
}
