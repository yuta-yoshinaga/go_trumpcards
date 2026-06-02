//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOsmosis_PersistsUndoHistory verifies the undo history survives a
// JSON round-trip (issue #1654 rollout).
func TestOsmosis_PersistsUndoHistory(t *testing.T) {
	t.Parallel()

	original := NewDefaultOsmosis()
	original.Reset()
	require.False(t, original.CanUndo(), "fresh game has no history")

	require.NoError(t, original.Draw())
	require.NoError(t, original.Draw())
	require.True(t, original.CanUndo(), "two draws should leave undoable history")
	originalHistoryLen := len(original.history)
	originalMoveCount := original.GetMoveCount()

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored Osmosis
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, originalMoveCount, restored.GetMoveCount(), "moveCount must round-trip")
	assert.Equal(t, originalHistoryLen, len(restored.history), "history length must round-trip")
	assert.True(t, restored.CanUndo(), "restored game must allow Undo")

	require.NoError(t, restored.Undo())
	assert.Equal(t, originalMoveCount-1, restored.GetMoveCount(), "Undo should rewind one step")
}

// TestOsmosis_PersistsHistoryRestoresExactSnapshot ensures snapshot fields are
// preserved exactly across serialisation.
func TestOsmosis_PersistsHistoryRestoresExactSnapshot(t *testing.T) {
	t.Parallel()

	original := NewDefaultOsmosis()
	original.Reset()
	require.NoError(t, original.Draw())

	preDrawWasteLen := len(original.waste)
	preDrawStockLen := len(original.stock)

	require.NoError(t, original.Draw())

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored Osmosis
	require.NoError(t, json.Unmarshal(data, &restored))

	require.NoError(t, restored.Undo())
	assert.Equal(t, preDrawWasteLen, len(restored.waste), "Undo restores waste length")
	assert.Equal(t, preDrawStockLen, len(restored.stock), "Undo restores stock length")
}

// TestOsmosis_HistoryRespectsMaxSliceLen rejects oversized history.
func TestOsmosis_HistoryRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigHistory := make([]map[string]any, osmosisMaxSliceLen+1)
	for i := range bigHistory {
		bigHistory[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"hi": bigHistory,
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Osmosis
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized history must be rejected")
}

// TestOsmosis_SnapshotStockRespectsMaxSliceLen rejects an oversized Stock slice
// inside a history snapshot.
func TestOsmosis_SnapshotStockRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigStock := make([]map[string]any, osmosisMaxSliceLen+1)
	for i := range bigStock {
		bigStock[i] = map[string]any{}
	}
	snapshot := map[string]any{"st": bigStock}
	payload := map[string]any{
		"tc": nil,
		"hi": []any{snapshot},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Osmosis
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized snapshot stock must be rejected")
}

// TestOsmosis_TopLevelReservePileRespectsMaxSliceLen rejects an oversized
// reserve pile at the top level.
func TestOsmosis_TopLevelReservePileRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigPile := make([]map[string]any, osmosisMaxSliceLen+1)
	for i := range bigPile {
		bigPile[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"rv": []any{bigPile, nil, nil, nil},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Osmosis
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized top-level reserve pile must be rejected")
}

// TestOsmosis_TopLevelFoundationPileRespectsMaxSliceLen rejects an oversized
// foundation pile at the top level.
func TestOsmosis_TopLevelFoundationPileRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigPile := make([]map[string]any, osmosisMaxSliceLen+1)
	for i := range bigPile {
		bigPile[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"fd": []any{bigPile, nil, nil, nil},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Osmosis
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized top-level foundation pile must be rejected")
}

// TestOsmosis_UnmarshalNormalizesNilSlices ensures nil slices deserialise to
// non-nil empties (defensive against null JSON fields).
func TestOsmosis_UnmarshalNormalizesNilSlices(t *testing.T) {
	t.Parallel()

	var restored Osmosis
	require.NoError(t, json.Unmarshal([]byte(`{"tc":null}`), &restored))
	assert.NotNil(t, restored.stock)
	assert.NotNil(t, restored.waste)
	for i := 0; i < OsmosisReserveCnt; i++ {
		assert.NotNil(t, restored.reserve[i])
	}
	assert.NotNil(t, restored.actionLog)
	assert.NotNil(t, restored.history)
}
