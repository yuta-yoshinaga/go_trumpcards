//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPokerSquares_PersistsUndoHistory verifies issue #1860: when a
// PokerSquares game is round-tripped through JSON (e.g. a Cloudflare KV
// restore), the undo history must survive so the player can still step
// backward. Prior to the fix UnmarshalJSON set history = nil, leaving a
// restored game with `cannot undo: no history`.
func TestPokerSquares_PersistsUndoHistory(t *testing.T) {
	t.Parallel()

	original := NewDefaultPokerSquares()
	original.Reset()
	require.False(t, original.CanUndo(), "fresh game has no history")

	require.NoError(t, original.Place(0, 0))
	require.NoError(t, original.Place(0, 1))
	require.True(t, original.CanUndo(), "two placements should leave undoable history")

	originalHistoryLen := len(original.history)
	originalPlacedCount := original.GetPlacedCount()

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored PokerSquares
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, originalPlacedCount, restored.GetPlacedCount(), "placedCount must round-trip")
	assert.Equal(t, originalHistoryLen, len(restored.history), "history length must round-trip")
	assert.True(t, restored.CanUndo(), "restored game must allow Undo")

	require.NoError(t, restored.Undo())
	assert.Equal(t, originalPlacedCount-1, restored.GetPlacedCount(), "Undo should rewind one placement")
}

// TestPokerSquares_PersistsHistoryRestoresExactSnapshot ensures Undo on a
// restored game returns the board, currentCard, and deckDrawCnt to the
// exact pre-Place state.
func TestPokerSquares_PersistsHistoryRestoresExactSnapshot(t *testing.T) {
	t.Parallel()

	original := NewDefaultPokerSquares()
	original.Reset()

	// Snapshot of state we expect Undo on the restored game to recover.
	preCurrentCard := original.GetCurrentCard()
	preStockCount := original.trumpCards.GetRemainingCount()

	require.NoError(t, original.Place(2, 3))

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored PokerSquares
	require.NoError(t, json.Unmarshal(data, &restored))

	require.NoError(t, restored.Undo())
	assert.Nil(t, restored.GetBoard()[2][3], "Undo clears the placed cell")
	assert.Equal(t, preCurrentCard, restored.GetCurrentCard(), "Undo restores currentCard")
	assert.Equal(t, preStockCount, restored.trumpCards.GetRemainingCount(), "Undo restores deck draw count")
	assert.Equal(t, 0, restored.GetPlacedCount(), "Undo restores placedCount")
}

// TestPokerSquares_HistoryRespectsMaxSliceLen rejects oversized history
// payloads.
func TestPokerSquares_HistoryRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigHistory := make([]map[string]any, pokerSquaresMaxSliceLen+1)
	for i := range bigHistory {
		bigHistory[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"hi": bigHistory,
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored PokerSquares
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized history must be rejected")
}

// TestPokerSquares_SnapshotNegativeDeckDrawCntRejected rejects snapshots
// that would panic during Undo via deck[i] with i<0.
func TestPokerSquares_SnapshotNegativeDeckDrawCntRejected(t *testing.T) {
	t.Parallel()

	payload := map[string]any{
		"tc": nil,
		"hi": []any{map[string]any{"dd": -1}},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored PokerSquares
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "negative deckDrawCnt must be rejected")
}

// TestPokerSquares_SnapshotNegativeActionLogLnRejected rejects snapshots
// that would panic during Undo via actionLog[:ll] with ll<0.
func TestPokerSquares_SnapshotNegativeActionLogLnRejected(t *testing.T) {
	t.Parallel()

	payload := map[string]any{
		"tc": nil,
		"hi": []any{map[string]any{"ll": -1}},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored PokerSquares
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "negative actionLogLn must be rejected")
}
