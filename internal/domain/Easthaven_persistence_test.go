//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEasthaven_PersistsUndoHistory verifies issue #1654: when an Easthaven
// game is round-tripped through JSON (e.g. a Cloudflare KV restore), the undo
// history must survive so the player can still step backward.
func TestEasthaven_PersistsUndoHistory(t *testing.T) {
	t.Parallel()

	original := NewDefaultEasthaven()
	original.Reset()
	require.False(t, original.CanUndo(), "fresh game has no history")

	original.takeSnapshot()
	original.takeSnapshot()
	require.True(t, original.CanUndo(), "two snapshots should leave undoable history")
	originalHistoryLen := len(original.history)
	originalMoveCount := original.GetMoveCount()

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored Easthaven
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, originalMoveCount, restored.GetMoveCount(), "moveCount must round-trip")
	assert.Equal(t, originalHistoryLen, len(restored.history), "history length must round-trip")
	assert.True(t, restored.CanUndo(), "restored game must allow Undo")
}

// TestEasthaven_PersistsHistoryRestoresExactSnapshot ensures the snapshot
// fields are preserved exactly so an Undo on the restored game matches the
// pre-snapshot state.
func TestEasthaven_PersistsHistoryRestoresExactSnapshot(t *testing.T) {
	t.Parallel()

	original := NewDefaultEasthaven()
	original.Reset()

	original.takeSnapshot()
	preMutationTableau0Len := len(original.tableau[0])

	original.tableau[0] = original.tableau[0][:0]
	original.takeSnapshot()

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored Easthaven
	require.NoError(t, json.Unmarshal(data, &restored))

	require.NoError(t, restored.Undo())
	require.NoError(t, restored.Undo())
	assert.Equal(t, preMutationTableau0Len, len(restored.tableau[0]),
		"second Undo restores tableau[0] length")
}

// TestEasthaven_HistoryRespectsMaxSliceLen rejects oversized history.
func TestEasthaven_HistoryRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigHistory := make([]map[string]any, easthavenMaxSliceLen+1)
	for i := range bigHistory {
		bigHistory[i] = map[string]any{}
	}
	payload := map[string]any{"tc": nil, "hi": bigHistory}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Easthaven
	require.Error(t, json.Unmarshal(data, &restored), "oversized history must be rejected")
}

// TestEasthaven_TopLevelStockRespectsMaxSliceLen rejects an oversized stock.
func TestEasthaven_TopLevelStockRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigStock := make([]map[string]any, easthavenMaxSliceLen+1)
	for i := range bigStock {
		bigStock[i] = map[string]any{}
	}
	payload := map[string]any{"tc": nil, "st": bigStock}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Easthaven
	require.Error(t, json.Unmarshal(data, &restored), "oversized stock must be rejected")
}

// TestEasthaven_TopLevelTableauColumnRespectsMaxSliceLen rejects payloads with
// an oversized tableau column at the top level.
func TestEasthaven_TopLevelTableauColumnRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigCol := make([]map[string]any, easthavenMaxSliceLen+1)
	for i := range bigCol {
		bigCol[i] = map[string]any{}
	}
	tableau := make([]any, 7)
	tableau[0] = bigCol
	payload := map[string]any{"tc": nil, "tb": tableau}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Easthaven
	require.Error(t, json.Unmarshal(data, &restored), "oversized top-level tableau column must be rejected")
}

// TestEasthaven_TopLevelFoundationPileRespectsMaxSliceLen rejects payloads with
// an oversized foundation pile at the top level.
func TestEasthaven_TopLevelFoundationPileRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigPile := make([]map[string]any, easthavenMaxSliceLen+1)
	for i := range bigPile {
		bigPile[i] = map[string]any{}
	}
	payload := map[string]any{"tc": nil, "fd": []any{bigPile, nil, nil, nil}}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Easthaven
	require.Error(t, json.Unmarshal(data, &restored), "oversized top-level foundation pile must be rejected")
}

// TestEasthaven_SnapshotColumnsRespectMaxSliceLen rejects payloads with an
// oversized tableau column or foundation pile inside a history snapshot.
func TestEasthaven_SnapshotColumnsRespectMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigCol := make([]map[string]any, easthavenMaxSliceLen+1)
	for i := range bigCol {
		bigCol[i] = map[string]any{}
	}
	tableau := make([]any, 7)
	tableau[0] = bigCol
	payload := map[string]any{"tc": nil, "hi": []any{map[string]any{"tb": tableau}}}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Easthaven
	require.Error(t, json.Unmarshal(data, &restored), "oversized snapshot tableau column must be rejected")

	bigPile := make([]map[string]any, easthavenMaxSliceLen+1)
	for i := range bigPile {
		bigPile[i] = map[string]any{}
	}
	payload2 := map[string]any{"tc": nil, "hi": []any{map[string]any{"fd": []any{bigPile, nil, nil, nil}}}}
	data2, err := json.Marshal(payload2)
	require.NoError(t, err)

	var restored2 Easthaven
	require.Error(t, json.Unmarshal(data2, &restored2), "oversized snapshot foundation pile must be rejected")
}
