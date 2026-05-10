//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGolf_PersistsUndoHistory verifies issue #1654.
func TestGolf_PersistsUndoHistory(t *testing.T) {
	t.Parallel()

	original := NewDefaultGolf()
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

	var restored Golf
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, originalMoveCount, restored.GetMoveCount(), "moveCount must round-trip")
	assert.Equal(t, originalHistoryLen, len(restored.history), "history length must round-trip")
	assert.True(t, restored.CanUndo(), "restored game must allow Undo")

	require.NoError(t, restored.Undo())
	assert.Equal(t, originalMoveCount-1, restored.GetMoveCount(), "Undo should rewind one step")
}

// TestGolf_PersistsHistoryRestoresExactSnapshot ensures snapshot fields
// are preserved exactly.
func TestGolf_PersistsHistoryRestoresExactSnapshot(t *testing.T) {
	t.Parallel()

	original := NewDefaultGolf()
	original.Reset()
	require.NoError(t, original.Draw())

	preDrawWasteLen := len(original.waste)
	preDrawStockLen := len(original.stock)

	require.NoError(t, original.Draw())

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored Golf
	require.NoError(t, json.Unmarshal(data, &restored))

	require.NoError(t, restored.Undo())
	assert.Equal(t, preDrawWasteLen, len(restored.waste), "Undo restores waste length")
	assert.Equal(t, preDrawStockLen, len(restored.stock), "Undo restores stock length")
}

// TestGolf_HistoryRespectsMaxSliceLen rejects oversized history.
func TestGolf_HistoryRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigHistory := make([]map[string]any, golfMaxSliceLen+1)
	for i := range bigHistory {
		bigHistory[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"hi": bigHistory,
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Golf
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized history must be rejected")
}

// TestGolf_SnapshotStockRespectsMaxSliceLen rejects payloads with an
// oversized inner Stock slice inside a snapshot.
func TestGolf_SnapshotStockRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigStock := make([]map[string]any, golfMaxSliceLen+1)
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

	var restored Golf
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized snapshot stock must be rejected")
}

// TestGolf_SnapshotWasteRespectsMaxSliceLen rejects payloads with an
// oversized inner Waste slice inside a snapshot.
func TestGolf_SnapshotWasteRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigWaste := make([]map[string]any, golfMaxSliceLen+1)
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

	var restored Golf
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized snapshot waste must be rejected")
}
