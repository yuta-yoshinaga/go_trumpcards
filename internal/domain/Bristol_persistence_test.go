//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBristol_PersistsUndoHistory verifies the undo history survives a
// JSON round-trip (issue #1654 rollout).
func TestBristol_PersistsUndoHistory(t *testing.T) {
	t.Parallel()

	original := NewDefaultBristol()
	original.Reset()
	require.False(t, original.CanUndo(), "fresh game has no history")

	require.NoError(t, original.Draw())
	require.NoError(t, original.Draw())
	require.True(t, original.CanUndo(), "two draws should leave undoable history")
	originalHistoryLen := len(original.history)
	originalMoveCount := original.GetMoveCount()

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored Bristol
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, originalMoveCount, restored.GetMoveCount(), "moveCount must round-trip")
	assert.Equal(t, originalHistoryLen, len(restored.history), "history length must round-trip")
	assert.True(t, restored.CanUndo(), "restored game must allow Undo")

	require.NoError(t, restored.Undo())
	assert.Equal(t, originalMoveCount-1, restored.GetMoveCount(), "Undo should rewind one step")
}

// TestBristol_PersistsHistoryRestoresExactSnapshot ensures snapshot fields are
// preserved exactly across serialisation.
func TestBristol_PersistsHistoryRestoresExactSnapshot(t *testing.T) {
	t.Parallel()

	original := NewDefaultBristol()
	original.Reset()
	require.NoError(t, original.Draw())

	preStockLen := original.GetStockCount()

	require.NoError(t, original.Draw())

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored Bristol
	require.NoError(t, json.Unmarshal(data, &restored))

	require.NoError(t, restored.Undo())
	assert.Equal(t, preStockLen, restored.GetStockCount(), "Undo restores stock length")
}

// TestBristol_HistoryRespectsMaxSliceLen rejects oversized history.
func TestBristol_HistoryRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigHistory := make([]map[string]any, bristolMaxSliceLen+1)
	for i := range bigHistory {
		bigHistory[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"hi": bigHistory,
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Bristol
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized history must be rejected")
}

// TestBristol_SnapshotStockRespectsMaxSliceLen rejects an oversized Stock slice
// inside a history snapshot.
func TestBristol_SnapshotStockRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigStock := make([]map[string]any, bristolMaxSliceLen+1)
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

	var restored Bristol
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized snapshot stock must be rejected")
}

// TestBristol_TopLevelTableauColumnRespectsMaxSliceLen rejects an oversized
// tableau column at the top level.
func TestBristol_TopLevelTableauColumnRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigCol := make([]map[string]any, bristolMaxSliceLen+1)
	for i := range bigCol {
		bigCol[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"tb": []any{bigCol, nil, nil, nil, nil, nil, nil, nil},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Bristol
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized top-level tableau column must be rejected")
}

// TestBristol_TopLevelFanRespectsMaxSliceLen rejects an oversized fan at the
// top level.
func TestBristol_TopLevelFanRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigFan := make([]map[string]any, bristolMaxSliceLen+1)
	for i := range bigFan {
		bigFan[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"fn": []any{bigFan, nil, nil},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Bristol
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized top-level fan must be rejected")
}

// TestBristol_UnmarshalNormalizesNilSlices ensures nil slices deserialise to
// non-nil empties (defensive against null JSON fields).
func TestBristol_UnmarshalNormalizesNilSlices(t *testing.T) {
	t.Parallel()

	var restored Bristol
	require.NoError(t, json.Unmarshal([]byte(`{"tc":null}`), &restored))
	assert.NotNil(t, restored.stock)
	for i := 0; i < BristolTableauCnt; i++ {
		assert.NotNil(t, restored.tableau[i])
	}
	for i := 0; i < BristolFanCnt; i++ {
		assert.NotNil(t, restored.fan[i])
	}
	for i := 0; i < BristolFoundationCnt; i++ {
		assert.NotNil(t, restored.foundation[i])
	}
	assert.NotNil(t, restored.actionLog)
	assert.NotNil(t, restored.history)
}

// TestBristol_UnmarshalFiltersNilCardElements ensures null elements within card
// slices are silently dropped to prevent nil pointer dereferences in game logic.
func TestBristol_UnmarshalFiltersNilCardElements(t *testing.T) {
	t.Parallel()

	payload := `{
		"tc": null,
		"tb": [[null, {"d":1,"v":5,"w":true}], null, null, null, null, null, null, null],
		"fn": [[null, {"d":2,"v":8,"w":true}], null, null],
		"fd": [[null, {"d":1,"v":1,"w":true}], null, null, null],
		"hi": [{"st": [null], "tb": [null,null,null,null,null,null,null,null], "fn": [null,null,null], "fd": [null,null,null,null]}]
	}`
	var restored Bristol
	require.NoError(t, json.Unmarshal([]byte(payload), &restored))

	require.Len(t, restored.tableau[0], 1)
	assert.Equal(t, 5, restored.tableau[0][0].GetValue())
	require.Len(t, restored.fan[0], 1)
	assert.Equal(t, 8, restored.fan[0][0].GetValue())
	require.Len(t, restored.foundation[0], 1)
	assert.Equal(t, 1, restored.foundation[0][0].GetValue())
	require.Len(t, restored.history, 1)
	assert.Empty(t, restored.history[0].stock)
}
