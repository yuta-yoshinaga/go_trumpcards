//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMonteCarlo_PersistsUndoHistory verifies issue #1860: when a MonteCarlo
// game is round-tripped through JSON (e.g. a Cloudflare KV restore), the
// undo history must survive so the player can still step backward. Prior
// to the fix UnmarshalJSON set history = nil, leaving a restored game
// with `cannot undo: no history`.
func TestMonteCarlo_PersistsUndoHistory(t *testing.T) {
	t.Parallel()

	original := NewDefaultMonteCarlo()
	original.Reset()
	require.False(t, original.CanUndo(), "fresh game has no history")

	// Inject a deterministic adjacent same-rank pair so Remove always succeeds.
	board := original.GetBoard()
	board[0][0] = NewCard(0, 7, true)
	board[0][1] = NewCard(1, 7, true)
	original.SetBoard(board)
	require.NoError(t, original.Remove(0, 0, 0, 1))

	require.True(t, original.CanUndo(), "remove should leave undoable history")
	originalHistoryLen := len(original.history)
	originalRemovedCount := original.GetRemovedCount()

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored MonteCarlo
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, originalRemovedCount, restored.GetRemovedCount(), "removedCount must round-trip")
	assert.Equal(t, originalHistoryLen, len(restored.history), "history length must round-trip")
	assert.True(t, restored.CanUndo(), "restored game must allow Undo")

	require.NoError(t, restored.Undo())
	assert.Equal(t, originalRemovedCount-2, restored.GetRemovedCount(), "Undo rewinds removedCount by 2")
}

// TestMonteCarlo_PersistsHistoryRestoresExactSnapshot ensures Undo on a
// restored game returns the board to the exact state it had before Remove.
func TestMonteCarlo_PersistsHistoryRestoresExactSnapshot(t *testing.T) {
	t.Parallel()

	original := NewDefaultMonteCarlo()
	original.Reset()
	board := original.GetBoard()
	a := NewCard(0, 9, true)
	b := NewCard(1, 9, true)
	board[2][3] = a
	board[3][4] = b
	original.SetBoard(board)
	require.NoError(t, original.Remove(2, 3, 3, 4))

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored MonteCarlo
	require.NoError(t, json.Unmarshal(data, &restored))

	require.NoError(t, restored.Undo())
	assert.Equal(t, a, restored.GetBoard()[2][3], "Undo restores removed card a")
	assert.Equal(t, b, restored.GetBoard()[3][4], "Undo restores removed card b")
	assert.Equal(t, 0, restored.GetRemovedCount(), "Undo restores removedCount")
	assert.False(t, restored.CanUndo(), "no further history after undoing the only snapshot")
}

// TestMonteCarlo_PersistsDealHistoryRestoresDeckDrawCnt verifies that the
// snapshot's deckDrawCnt round-trips so Undo on Deal correctly returns
// stock cards to the deck on a restored game.
func TestMonteCarlo_PersistsDealHistoryRestoresDeckDrawCnt(t *testing.T) {
	t.Parallel()

	original := NewDefaultMonteCarlo()
	original.Reset()
	// Vacate one cell so Deal draws exactly one card from stock.
	board := original.GetBoard()
	board[4][4] = nil
	original.SetBoard(board)
	stockBefore := original.GetStockCount()
	require.NoError(t, original.Deal())
	require.Less(t, original.GetStockCount(), stockBefore, "Deal should consume stock")

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored MonteCarlo
	require.NoError(t, json.Unmarshal(data, &restored))
	require.NoError(t, restored.Undo())
	assert.Equal(t, stockBefore, restored.GetStockCount(), "Undo on restored game returns stock to pre-Deal size")
	assert.Equal(t, 0, restored.GetDealCount(), "Undo resets dealCount")
}

// TestMonteCarlo_HistoryRespectsMaxSliceLen rejects oversized history payloads.
func TestMonteCarlo_HistoryRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigHistory := make([]map[string]any, monteCarloMaxSliceLen+1)
	for i := range bigHistory {
		bigHistory[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"hi": bigHistory,
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored MonteCarlo
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized history must be rejected")
}

// TestMonteCarlo_SnapshotNegativeDeckDrawCntRejected rejects snapshots that
// would panic during Undo via deck[i] with i<0.
func TestMonteCarlo_SnapshotNegativeDeckDrawCntRejected(t *testing.T) {
	t.Parallel()

	payload := map[string]any{
		"tc": nil,
		"hi": []any{map[string]any{"dd": -1}},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored MonteCarlo
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "negative deckDrawCnt must be rejected")
}

// TestMonteCarlo_SnapshotNegativeActionLogLnRejected rejects snapshots that
// would panic during Undo via actionLog[:ll] with ll<0.
func TestMonteCarlo_SnapshotNegativeActionLogLnRejected(t *testing.T) {
	t.Parallel()

	payload := map[string]any{
		"tc": nil,
		"hi": []any{map[string]any{"ll": -1}},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored MonteCarlo
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "negative actionLogLn must be rejected")
}

// TestMonteCarlo_SnapshotExcessiveDeckDrawCntRejected rejects snapshots with
// DeckDrawCnt > MonteCarloDeckSize, which would cause an out-of-bounds panic
// in the deck slice inside Undo().
func TestMonteCarlo_SnapshotExcessiveDeckDrawCntRejected(t *testing.T) {
	t.Parallel()

	payload := map[string]any{
		"tc": nil,
		"hi": []any{map[string]any{"dd": MonteCarloDeckSize + 1}},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored MonteCarlo
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "excessive deckDrawCnt must be rejected")
}

// TestMonteCarlo_SnapshotExcessiveActionLogLnRejected rejects snapshots with
// ActionLogLn > monteCarloMaxSliceLen, consistent with the maxSliceLen defence
// used across the codebase.
func TestMonteCarlo_SnapshotExcessiveActionLogLnRejected(t *testing.T) {
	t.Parallel()

	payload := map[string]any{
		"tc": nil,
		"hi": []any{map[string]any{"ll": monteCarloMaxSliceLen + 1}},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored MonteCarlo
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "excessive actionLogLn must be rejected")
}
