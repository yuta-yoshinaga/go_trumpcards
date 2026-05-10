//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSpider_PersistsUndoHistory verifies issue #1654: when a Spider
// game is round-tripped through JSON (e.g. a Cloudflare KV restore), the
// undo history must survive so the player can still step backward.
func TestSpider_PersistsUndoHistory(t *testing.T) {
	t.Parallel()

	original := NewDefaultSpider()
	original.Reset()
	require.False(t, original.CanUndo(), "fresh game has no history")

	// Deal is deterministic: after Reset every tableau column is non-empty
	// and the stock has enough cards for at least one Deal.
	require.NoError(t, original.Deal())
	require.True(t, original.CanUndo(), "Deal should leave undoable history")
	originalHistoryLen := len(original.history)
	originalMoveCount := original.GetMoveCount()

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored Spider
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, originalMoveCount, restored.GetMoveCount(), "moveCount must round-trip")
	assert.Equal(t, originalHistoryLen, len(restored.history), "history length must round-trip")
	assert.True(t, restored.CanUndo(), "restored game must allow Undo")

	require.NoError(t, restored.Undo())
	assert.Equal(t, originalMoveCount-1, restored.GetMoveCount(), "Undo should rewind one step")
}

// TestSpider_PersistsHistoryRestoresExactSnapshot ensures the snapshot
// fields (tableau, stock, completedSuits, score) are preserved exactly.
func TestSpider_PersistsHistoryRestoresExactSnapshot(t *testing.T) {
	t.Parallel()

	original := NewDefaultSpider()
	original.Reset()

	preDealStockLen := len(original.stock)
	preDealScore := original.score

	require.NoError(t, original.Deal())

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored Spider
	require.NoError(t, json.Unmarshal(data, &restored))

	require.NoError(t, restored.Undo())
	assert.Equal(t, preDealStockLen, len(restored.stock), "Undo restores stock length")
	assert.Equal(t, preDealScore, restored.score, "Undo restores score")
}

// TestSpider_HistoryRespectsMaxSliceLen rejects oversized history payloads.
func TestSpider_HistoryRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigHistory := make([]map[string]any, spiderMaxSliceLen+1)
	for i := range bigHistory {
		bigHistory[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"hi": bigHistory,
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Spider
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized history must be rejected")
}

// TestSpider_SnapshotStockRespectsMaxSliceLen rejects payloads that
// smuggle an oversized inner Stock slice inside a history snapshot.
func TestSpider_SnapshotStockRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigStock := make([]map[string]any, spiderMaxSliceLen+1)
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

	var restored Spider
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized snapshot stock must be rejected")
}

// TestSpider_SnapshotTableauColumnRespectsMaxSliceLen rejects payloads
// that smuggle an oversized Tableau column inside a history snapshot.
func TestSpider_SnapshotTableauColumnRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigCol := make([]map[string]any, spiderMaxSliceLen+1)
	for i := range bigCol {
		bigCol[i] = map[string]any{}
	}
	snapshot := map[string]any{
		"tb": []any{bigCol, nil, nil, nil, nil, nil, nil, nil, nil, nil},
	}
	payload := map[string]any{
		"tc": nil,
		"hi": []any{snapshot},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Spider
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized snapshot tableau column must be rejected")
}
