//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPyramid_PersistsUndoHistory verifies issue #1654: when a Pyramid
// game is round-tripped through JSON (e.g. a Cloudflare KV restore), the
// undo history must survive so the player can still step backward.
func TestPyramid_PersistsUndoHistory(t *testing.T) {
	t.Parallel()

	original := NewDefaultPyramid()
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

	var restored Pyramid
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, originalMoveCount, restored.GetMoveCount(), "moveCount must round-trip")
	assert.Equal(t, originalHistoryLen, len(restored.history), "history length must round-trip")
	assert.True(t, restored.CanUndo(), "restored game must allow Undo")

	require.NoError(t, restored.Undo())
	assert.Equal(t, originalMoveCount-1, restored.GetMoveCount(), "Undo should rewind one step")
}

// TestPyramid_PersistsHistoryRestoresExactSnapshot ensures the snapshot
// fields (pyramid, stock, waste) are preserved exactly.
func TestPyramid_PersistsHistoryRestoresExactSnapshot(t *testing.T) {
	t.Parallel()

	original := NewDefaultPyramid()
	original.Reset()
	require.NoError(t, original.Draw())

	preDrawWasteLen := len(original.waste)
	preDrawStockLen := len(original.stock)

	require.NoError(t, original.Draw())

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored Pyramid
	require.NoError(t, json.Unmarshal(data, &restored))

	require.NoError(t, restored.Undo())
	assert.Equal(t, preDrawWasteLen, len(restored.waste), "Undo restores waste length")
	assert.Equal(t, preDrawStockLen, len(restored.stock), "Undo restores stock length")
}

// TestPyramid_HistoryRespectsMaxSliceLen rejects oversized history payloads.
func TestPyramid_HistoryRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigHistory := make([]map[string]any, pyramidMaxSliceLen+1)
	for i := range bigHistory {
		bigHistory[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"hi": bigHistory,
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Pyramid
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized history must be rejected")
}

// TestPyramid_SnapshotStockRespectsMaxSliceLen rejects payloads that
// smuggle an oversized inner Stock slice inside a history snapshot.
func TestPyramid_SnapshotStockRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigStock := make([]map[string]any, pyramidMaxSliceLen+1)
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

	var restored Pyramid
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized snapshot stock must be rejected")
}

// TestPyramid_SnapshotPyramidRowRespectsMaxSliceLen rejects payloads that
// smuggle an oversized Pyramid row slice inside a history snapshot.
func TestPyramid_SnapshotPyramidRowRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigRow := make([]map[string]any, pyramidMaxSliceLen+1)
	for i := range bigRow {
		bigRow[i] = map[string]any{}
	}
	snapshot := map[string]any{
		"py": []any{bigRow, nil, nil, nil, nil, nil, nil},
	}
	payload := map[string]any{
		"tc": nil,
		"hi": []any{snapshot},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Pyramid
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized snapshot pyramid row must be rejected")
}
