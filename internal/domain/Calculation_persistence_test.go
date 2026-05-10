//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCalculation_PersistsUndoHistory verifies issue #1654.
func TestCalculation_PersistsUndoHistory(t *testing.T) {
	t.Parallel()

	original := NewDefaultCalculation()
	original.Reset()
	require.False(t, original.CanUndo(), "fresh game has no history")

	// PlayStockToWaste(0) is deterministic — after Reset stock has cards
	// regardless of shuffle, and waste 0 starts empty.
	require.NoError(t, original.PlayStockToWaste(0))
	require.NoError(t, original.PlayStockToWaste(1))
	require.True(t, original.CanUndo(), "two plays should leave undoable history")
	originalHistoryLen := len(original.history)
	originalMoveCount := original.GetMoveCount()

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored Calculation
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, originalMoveCount, restored.GetMoveCount(), "moveCount must round-trip")
	assert.Equal(t, originalHistoryLen, len(restored.history), "history length must round-trip")
	assert.True(t, restored.CanUndo(), "restored game must allow Undo")

	require.NoError(t, restored.Undo())
	assert.Equal(t, originalMoveCount-1, restored.GetMoveCount(), "Undo should rewind one step")
}

// TestCalculation_PersistsHistoryRestoresExactSnapshot ensures snapshot
// fields are preserved exactly.
func TestCalculation_PersistsHistoryRestoresExactSnapshot(t *testing.T) {
	t.Parallel()

	original := NewDefaultCalculation()
	original.Reset()
	require.NoError(t, original.PlayStockToWaste(0))

	prePlayWaste1Len := len(original.wastes[1])
	prePlayStockLen := len(original.stock)

	require.NoError(t, original.PlayStockToWaste(1))

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored Calculation
	require.NoError(t, json.Unmarshal(data, &restored))

	require.NoError(t, restored.Undo())
	assert.Equal(t, prePlayWaste1Len, len(restored.wastes[1]), "Undo restores waste 1 length")
	assert.Equal(t, prePlayStockLen, len(restored.stock), "Undo restores stock length")
}

// TestCalculation_HistoryRespectsMaxSliceLen rejects oversized history.
func TestCalculation_HistoryRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigHistory := make([]map[string]any, calculationMaxSliceLen+1)
	for i := range bigHistory {
		bigHistory[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"hi": bigHistory,
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Calculation
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized history must be rejected")
}

// TestCalculation_SnapshotStockRespectsMaxSliceLen rejects payloads
// with an oversized inner Stock slice inside a snapshot.
func TestCalculation_SnapshotStockRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigStock := make([]map[string]any, calculationMaxSliceLen+1)
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

	var restored Calculation
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized snapshot stock must be rejected")
}

// TestCalculation_SnapshotFoundationRespectsMaxSliceLen rejects payloads
// with an oversized foundation pile inside a snapshot.
func TestCalculation_SnapshotFoundationRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigPile := make([]map[string]any, calculationMaxSliceLen+1)
	for i := range bigPile {
		bigPile[i] = map[string]any{}
	}
	snapshot := map[string]any{
		"fd": []any{bigPile, nil, nil, nil},
	}
	payload := map[string]any{
		"tc": nil,
		"hi": []any{snapshot},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Calculation
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized snapshot foundation must be rejected")
}

// TestCalculation_SnapshotWasteRespectsMaxSliceLen rejects payloads with
// an oversized waste pile inside a snapshot.
func TestCalculation_SnapshotWasteRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigPile := make([]map[string]any, calculationMaxSliceLen+1)
	for i := range bigPile {
		bigPile[i] = map[string]any{}
	}
	snapshot := map[string]any{
		"wa": []any{bigPile, nil, nil, nil},
	}
	payload := map[string]any{
		"tc": nil,
		"hi": []any{snapshot},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Calculation
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized snapshot waste must be rejected")
}
