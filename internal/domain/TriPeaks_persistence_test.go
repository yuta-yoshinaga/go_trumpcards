//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTriPeaks_PersistsUndoHistory verifies issue #1654: when a TriPeaks
// game is round-tripped through JSON (e.g. a Cloudflare KV restore), the
// undo history must survive so the player can still step backward. Prior
// to the fix UnmarshalJSON set history = nil, leaving a restored game
// with `cannot undo: no history`.
func TestTriPeaks_PersistsUndoHistory(t *testing.T) {
	t.Parallel()

	original := NewDefaultTriPeaks()
	original.Reset()
	require.False(t, original.CanUndo(), "fresh game has no history")

	// Draw is deterministic — after Reset the stock has cards regardless of shuffle.
	require.NoError(t, original.Draw())
	require.NoError(t, original.Draw())
	require.True(t, original.CanUndo(), "two draws should leave undoable history")
	originalHistoryLen := len(original.history)
	originalMoveCount := original.GetMoveCount()

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored TriPeaks
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, originalMoveCount, restored.GetMoveCount(), "moveCount must round-trip")
	assert.Equal(t, originalHistoryLen, len(restored.history), "history length must round-trip")
	assert.True(t, restored.CanUndo(), "restored game must allow Undo")

	require.NoError(t, restored.Undo())
	assert.Equal(t, originalMoveCount-1, restored.GetMoveCount(), "Undo should rewind one step")
}

// TestTriPeaks_PersistsHistoryRestoresExactSnapshot ensures the snapshot
// fields (layout, stock, waste) are preserved exactly so an Undo on a
// restored game returns the same state the game would have had before the
// action that produced the snapshot.
func TestTriPeaks_PersistsHistoryRestoresExactSnapshot(t *testing.T) {
	t.Parallel()

	original := NewDefaultTriPeaks()
	original.Reset()
	require.NoError(t, original.Draw())

	preDrawWasteLen := len(original.waste)
	preDrawStockLen := len(original.stock)

	require.NoError(t, original.Draw())

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored TriPeaks
	require.NoError(t, json.Unmarshal(data, &restored))

	require.NoError(t, restored.Undo())
	assert.Equal(t, preDrawWasteLen, len(restored.waste), "Undo restores waste length")
	assert.Equal(t, preDrawStockLen, len(restored.stock), "Undo restores stock length")
}

// TestTriPeaks_HistoryRespectsMaxSliceLen rejects malicious payloads that
// attempt to exhaust memory through an oversized history array — same
// defence as the existing maxSliceLen guards on stock/waste/actionLog.
func TestTriPeaks_HistoryRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigHistory := make([]map[string]any, triPeaksMaxSliceLen+1)
	for i := range bigHistory {
		bigHistory[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"hi": bigHistory,
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored TriPeaks
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized history must be rejected")
}

// TestTriPeaks_SnapshotStockRespectsMaxSliceLen rejects payloads that
// smuggle an oversized inner Stock slice inside a history snapshot.
func TestTriPeaks_SnapshotStockRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigStock := make([]map[string]any, triPeaksMaxSliceLen+1)
	for i := range bigStock {
		bigStock[i] = map[string]any{}
	}
	snapshot := map[string]any{
		"st": bigStock,
	}
	payload := map[string]any{
		"tc": nil,
		"hi": []any{snapshot},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored TriPeaks
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized snapshot stock must be rejected")
}

// TestTriPeaks_SnapshotWasteRespectsMaxSliceLen rejects payloads that
// smuggle an oversized inner Waste slice inside a history snapshot.
// The Stock and Waste branches share a combined `||` guard, so this is
// a separate test to exercise the Waste path explicitly.
func TestTriPeaks_SnapshotWasteRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigWaste := make([]map[string]any, triPeaksMaxSliceLen+1)
	for i := range bigWaste {
		bigWaste[i] = map[string]any{}
	}
	snapshot := map[string]any{
		"wa": bigWaste,
	}
	payload := map[string]any{
		"tc": nil,
		"hi": []any{snapshot},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored TriPeaks
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized snapshot waste must be rejected")
}
