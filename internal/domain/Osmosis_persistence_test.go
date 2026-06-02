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
	for i := 0; i < OsmosisFoundationCnt; i++ {
		assert.NotNil(t, restored.foundation[i])
	}
	assert.NotNil(t, restored.actionLog)
	assert.NotNil(t, restored.history)
}

// TestOsmosis_UnmarshalFiltersNilCardElements ensures null elements within card
// slices are silently dropped to prevent nil pointer dereferences in game logic.
func TestOsmosis_UnmarshalFiltersNilCardElements(t *testing.T) {
	t.Parallel()

	// JSON with explicit null entries in waste, foundation, reserve, and a snapshot.
	payload := `{
		"tc": null,
		"wa": [null, {"d":1,"v":5,"w":true}],
		"fd": [[null, {"d":1,"v":1,"w":true}], null, null, null],
		"rv": [[null, {"d":2,"v":8,"w":true}], null, null, null],
		"hi": [{"wa": [null], "fd": [null, null, null, null], "rv": [null, null, null, null]}]
	}`
	var restored Osmosis
	require.NoError(t, json.Unmarshal([]byte(payload), &restored))

	// Null waste element is dropped; only the real card survives.
	require.Len(t, restored.waste, 1)
	assert.Equal(t, 5, restored.waste[0].GetValue())

	// Null foundation element is dropped; only the real card survives.
	require.Len(t, restored.foundation[0], 1)
	assert.Equal(t, 1, restored.foundation[0][0].GetValue())

	// Null reserve element is dropped; only the real card survives.
	require.Len(t, restored.reserve[0], 1)
	assert.Equal(t, 8, restored.reserve[0][0].GetValue())

	// Snapshot's waste null element is also dropped.
	require.Len(t, restored.history, 1)
	assert.Empty(t, restored.history[0].waste)
}
