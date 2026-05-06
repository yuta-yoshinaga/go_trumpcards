//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestScorpion_PersistsUndoHistory verifies issue #1654: when a Scorpion
// game is round-tripped through JSON (e.g. a Cloudflare KV restore), the
// undo history must survive so the player can still step backward.
func TestScorpion_PersistsUndoHistory(t *testing.T) {
	t.Parallel()

	original := NewDefaultScorpion()
	original.Reset()
	require.False(t, original.CanUndo(), "fresh game has no history")

	// Deal is deterministic — after Reset the stock has 3 cards regardless
	// of shuffle, and Deal places one card per face-down column.
	require.NoError(t, original.Deal())
	require.True(t, original.CanUndo(), "Deal should leave undoable history")
	originalHistoryLen := len(original.history)
	originalMoveCount := original.GetMoveCount()

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored Scorpion
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, originalMoveCount, restored.GetMoveCount(), "moveCount must round-trip")
	assert.Equal(t, originalHistoryLen, len(restored.history), "history length must round-trip")
	assert.True(t, restored.CanUndo(), "restored game must allow Undo")

	require.NoError(t, restored.Undo())
	assert.Equal(t, originalMoveCount-1, restored.GetMoveCount(), "Undo should rewind one step")
}

// TestScorpion_PersistsHistoryRestoresExactSnapshot ensures the snapshot
// fields are preserved exactly so an Undo on a restored game returns to
// the same state as the pre-action snapshot.
func TestScorpion_PersistsHistoryRestoresExactSnapshot(t *testing.T) {
	t.Parallel()

	original := NewDefaultScorpion()
	original.Reset()

	// Capture pre-Deal state
	preDealStockLen := len(original.stock)

	require.NoError(t, original.Deal())

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored Scorpion
	require.NoError(t, json.Unmarshal(data, &restored))

	require.NoError(t, restored.Undo())
	assert.Equal(t, preDealStockLen, len(restored.stock), "Undo restores stock length")
}

// TestScorpion_HistoryRespectsMaxSliceLen rejects oversized history.
func TestScorpion_HistoryRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigHistory := make([]map[string]any, scorpionMaxSliceLen+1)
	for i := range bigHistory {
		bigHistory[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"hi": bigHistory,
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Scorpion
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized history must be rejected")
}

// TestScorpion_TopLevelTableauColumnRespectsMaxSliceLen rejects payloads
// with an oversized Tableau column at the top level.
func TestScorpion_TopLevelTableauColumnRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigCol := make([]map[string]any, scorpionMaxSliceLen+1)
	for i := range bigCol {
		bigCol[i] = map[string]any{}
	}
	tableau := make([]any, 7)
	tableau[0] = bigCol
	payload := map[string]any{
		"tc": nil,
		"tb": tableau,
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Scorpion
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized top-level tableau column must be rejected")
}

// TestScorpion_SnapshotTableauColumnRespectsMaxSliceLen rejects payloads
// with an oversized tableau column inside a history snapshot.
func TestScorpion_SnapshotTableauColumnRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigCol := make([]map[string]any, scorpionMaxSliceLen+1)
	for i := range bigCol {
		bigCol[i] = map[string]any{}
	}
	tableau := make([]any, 7)
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

	var restored Scorpion
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized snapshot tableau column must be rejected")
}

// TestScorpion_SnapshotStockRespectsMaxSliceLen rejects payloads with
// an oversized inner Stock slice inside a history snapshot.
func TestScorpion_SnapshotStockRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigStock := make([]map[string]any, scorpionMaxSliceLen+1)
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

	var restored Scorpion
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized snapshot stock must be rejected")
}
