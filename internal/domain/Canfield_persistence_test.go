//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCanfield_PersistsUndoHistory verifies issue #1654.
func TestCanfield_PersistsUndoHistory(t *testing.T) {
	t.Parallel()

	original := NewDefaultCanfield()
	original.Reset()
	require.False(t, original.CanUndo(), "fresh game has no history")

	require.NoError(t, original.Draw())
	require.NoError(t, original.Draw())
	require.True(t, original.CanUndo(), "two draws should leave undoable history")
	originalHistoryLen := len(original.history)
	originalMoveCount := original.GetMoveCount()

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored Canfield
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, originalMoveCount, restored.GetMoveCount(), "moveCount must round-trip")
	assert.Equal(t, originalHistoryLen, len(restored.history), "history length must round-trip")
	assert.True(t, restored.CanUndo(), "restored game must allow Undo")

	require.NoError(t, restored.Undo())
	assert.Equal(t, originalMoveCount-1, restored.GetMoveCount(), "Undo should rewind one step")
}

// TestCanfield_PersistsHistoryRestoresExactSnapshot ensures the snapshot
// fields are preserved exactly.
func TestCanfield_PersistsHistoryRestoresExactSnapshot(t *testing.T) {
	t.Parallel()

	original := NewDefaultCanfield()
	original.Reset()
	require.NoError(t, original.Draw())

	preDrawWasteLen := len(original.waste)
	preDrawStockLen := len(original.stock)

	require.NoError(t, original.Draw())

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored Canfield
	require.NoError(t, json.Unmarshal(data, &restored))

	require.NoError(t, restored.Undo())
	assert.Equal(t, preDrawWasteLen, len(restored.waste), "Undo restores waste length")
	assert.Equal(t, preDrawStockLen, len(restored.stock), "Undo restores stock length")
}

// TestCanfield_HistoryRespectsMaxSliceLen rejects oversized history.
func TestCanfield_HistoryRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigHistory := make([]map[string]any, canfieldMaxSliceLen+1)
	for i := range bigHistory {
		bigHistory[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"hi": bigHistory,
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Canfield
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized history must be rejected")
}

// TestCanfield_SnapshotReserveRespectsMaxSliceLen rejects payloads with
// an oversized inner Reserve slice inside a history snapshot.
func TestCanfield_SnapshotReserveRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigReserve := make([]map[string]any, canfieldMaxSliceLen+1)
	for i := range bigReserve {
		bigReserve[i] = map[string]any{}
	}
	snapshot := map[string]any{
		"rv": bigReserve,
	}
	payload := map[string]any{
		"tc": nil,
		"hi": []any{snapshot},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Canfield
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized snapshot reserve must be rejected")
}

// TestCanfield_SnapshotStockRespectsMaxSliceLen rejects payloads with
// an oversized inner Stock slice inside a history snapshot.
func TestCanfield_SnapshotStockRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigStock := make([]map[string]any, canfieldMaxSliceLen+1)
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

	var restored Canfield
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized snapshot stock must be rejected")
}

// TestCanfield_TopLevelTableauColumnRespectsMaxSliceLen rejects payloads
// with an oversized Tableau column at the top level.
func TestCanfield_TopLevelTableauColumnRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigCol := make([]map[string]any, canfieldMaxSliceLen+1)
	for i := range bigCol {
		bigCol[i] = map[string]any{}
	}
	// Canfield has 4 tableau columns.
	payload := map[string]any{
		"tc": nil,
		"tb": []any{bigCol, nil, nil, nil},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Canfield
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized top-level tableau column must be rejected")
}

// TestCanfield_TopLevelFoundationPileRespectsMaxSliceLen rejects payloads
// with an oversized Foundation pile at the top level.
func TestCanfield_TopLevelFoundationPileRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigPile := make([]map[string]any, canfieldMaxSliceLen+1)
	for i := range bigPile {
		bigPile[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"fd": []any{bigPile, nil, nil, nil},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Canfield
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized top-level foundation pile must be rejected")
}
