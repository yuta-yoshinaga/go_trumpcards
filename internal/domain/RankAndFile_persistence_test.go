//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRankAndFile_PersistsUndoHistory verifies issue #1654: when a
// RankAndFile game is round-tripped through JSON (e.g. a Cloudflare KV
// restore), the undo history must survive so the player can still step
// backward.
func TestRankAndFile_PersistsUndoHistory(t *testing.T) {
	t.Parallel()

	original := NewDefaultRankAndFile()
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

	var restored RankAndFile
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, originalMoveCount, restored.GetMoveCount(), "moveCount must round-trip")
	assert.Equal(t, originalHistoryLen, len(restored.history), "history length must round-trip")
	assert.True(t, restored.CanUndo(), "restored game must allow Undo")

	require.NoError(t, restored.Undo())
	assert.Equal(t, originalMoveCount-1, restored.GetMoveCount(), "Undo should rewind one step")
}

// TestRankAndFile_PersistsHistoryRestoresExactSnapshot ensures the
// snapshot fields are preserved exactly.
func TestRankAndFile_PersistsHistoryRestoresExactSnapshot(t *testing.T) {
	t.Parallel()

	original := NewDefaultRankAndFile()
	original.Reset()
	require.NoError(t, original.Draw())

	preDrawWasteLen := len(original.waste)
	preDrawStockLen := len(original.stock)

	require.NoError(t, original.Draw())

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored RankAndFile
	require.NoError(t, json.Unmarshal(data, &restored))

	require.NoError(t, restored.Undo())
	assert.Equal(t, preDrawWasteLen, len(restored.waste), "Undo restores waste length")
	assert.Equal(t, preDrawStockLen, len(restored.stock), "Undo restores stock length")
}

// TestRankAndFile_HistoryRespectsMaxSliceLen rejects oversized history.
func TestRankAndFile_HistoryRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigHistory := make([]map[string]any, rankAndFileMaxSliceLen+1)
	for i := range bigHistory {
		bigHistory[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"hi": bigHistory,
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored RankAndFile
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized history must be rejected")
}

// TestRankAndFile_SnapshotStockRespectsMaxSliceLen rejects payloads
// with an oversized inner Stock slice inside a history snapshot.
func TestRankAndFile_SnapshotStockRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigStock := make([]map[string]any, rankAndFileMaxSliceLen+1)
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

	var restored RankAndFile
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized snapshot stock must be rejected")
}

// TestRankAndFile_SnapshotTableauColumnRespectsMaxSliceLen rejects
// payloads with an oversized inner tableau column inside a snapshot.
func TestRankAndFile_SnapshotTableauColumnRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigCol := make([]map[string]any, rankAndFileMaxSliceLen+1)
	for i := range bigCol {
		bigCol[i] = map[string]any{}
	}
	// RankAndFile has 10 tableau columns.
	tableau := make([]any, 10)
	tableau[0] = bigCol
	snapshot := map[string]any{
		"tb": tableau,
	}
	payload := map[string]any{
		"tc": nil,
		"hi": []any{snapshot},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored RankAndFile
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized snapshot tableau column must be rejected")
}

// TestRankAndFile_TopLevelTableauColumnRespectsMaxSliceLen rejects
// payloads with an oversized Tableau column at the top level.
func TestRankAndFile_TopLevelTableauColumnRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigCol := make([]map[string]any, rankAndFileMaxSliceLen+1)
	for i := range bigCol {
		bigCol[i] = map[string]any{}
	}
	tableau := make([]any, 10)
	tableau[0] = bigCol
	payload := map[string]any{
		"tc": nil,
		"tb": tableau,
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored RankAndFile
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized top-level tableau column must be rejected")
}

// TestRankAndFile_TopLevelFoundationPileRespectsMaxSliceLen rejects
// payloads with an oversized Foundation pile at the top level.
func TestRankAndFile_TopLevelFoundationPileRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigPile := make([]map[string]any, rankAndFileMaxSliceLen+1)
	for i := range bigPile {
		bigPile[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"fd": []any{bigPile, nil, nil, nil, nil, nil, nil, nil},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored RankAndFile
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized top-level foundation pile must be rejected")
}
