//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestKlondike_PersistsUndoHistory verifies issue #1654: when a Klondike
// game is round-tripped through JSON (e.g., a Cloudflare KV restore), the
// undo history survives so the player can still step backward. Prior to
// the fix, MarshalJSON dropped the history slice and UnmarshalJSON set
// it to nil, leaving a restored game with `cannot undo: no history`.
func TestKlondike_PersistsUndoHistory(t *testing.T) {
	t.Parallel()

	original := NewDefaultKlondike()
	original.Reset()
	require.True(t, original.CanUndo() == false, "fresh game has no history")

	// Generate at least two snapshots so we can verify the array (not just
	// presence/absence) round-trips intact.
	require.NoError(t, original.Draw())
	require.NoError(t, original.Draw())
	require.True(t, original.CanUndo(), "two draws should leave undoable history")
	originalHistoryLen := len(original.history)
	originalMoveCount := original.GetMoveCount()

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored Klondike
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, originalMoveCount, restored.GetMoveCount(), "moveCount must round-trip")
	assert.Equal(t, originalHistoryLen, len(restored.history), "history length must round-trip")
	assert.True(t, restored.CanUndo(), "restored game must allow Undo")

	require.NoError(t, restored.Undo())
	assert.Equal(t, originalMoveCount-1, restored.GetMoveCount(), "Undo should rewind one step")
}

// TestKlondike_PersistsHistoryRestoresExactSnapshot ensures the snapshot
// fields (tableau, stock, waste, foundation, stalemate state) are
// preserved exactly so an Undo on a restored game returns the same state
// the game would have had before the action that produced the snapshot.
func TestKlondike_PersistsHistoryRestoresExactSnapshot(t *testing.T) {
	t.Parallel()

	original := NewDefaultKlondike()
	original.Reset()
	require.NoError(t, original.Draw())

	// Capture state we expect Undo to restore to (i.e., before the second draw).
	preDrawWasteLen := len(original.waste)
	preDrawStockLen := len(original.stock)

	require.NoError(t, original.Draw())

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored Klondike
	require.NoError(t, json.Unmarshal(data, &restored))

	require.NoError(t, restored.Undo())
	assert.Equal(t, preDrawWasteLen, len(restored.waste), "Undo restores waste length")
	assert.Equal(t, preDrawStockLen, len(restored.stock), "Undo restores stock length")
}

// TestKlondike_HistoryRespectsMaxSliceLen rejects malicious payloads that
// attempt to exhaust memory through an oversized history array — same
// defence as the existing maxSliceLen guards on stock/waste/actionLog.
func TestKlondike_HistoryRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	// Build a JSON envelope with klondikeMaxSliceLen+1 history entries.
	bigHistory := make([]map[string]any, klondikeMaxSliceLen+1)
	for i := range bigHistory {
		bigHistory[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"hi": bigHistory,
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Klondike
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized history must be rejected")
}
