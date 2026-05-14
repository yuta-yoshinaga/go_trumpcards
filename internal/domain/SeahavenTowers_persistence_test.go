//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSeahavenTowers_PersistsUndoHistory verifies that JSON round-trip
// preserves undo history so a player can step backward after a KV restore.
func TestSeahavenTowers_PersistsUndoHistory(t *testing.T) {
	t.Parallel()

	original := NewDefaultSeahavenTowers()
	original.Reset()
	require.False(t, original.CanUndo())

	// After Reset both reserved cells are occupied; move both to columns we control.
	// Use a deterministic setup instead — clear and place known cards.
	clearTableauST(original)
	clearReservedST(original)
	original.tableau[0] = []*Card{makeCard(CardDesignSpade, 1)}
	original.tableau[1] = []*Card{makeCard(CardDesignSpade, 2)}
	require.NoError(t, original.MoveTableauToFoundation(0))
	require.NoError(t, original.MoveTableauToFoundation(1))
	require.True(t, original.CanUndo())
	originalHistoryLen := len(original.history)
	originalMoveCount := original.GetMoveCount()

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored SeahavenTowers
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, originalMoveCount, restored.GetMoveCount())
	assert.Equal(t, originalHistoryLen, len(restored.history))
	assert.True(t, restored.CanUndo())

	require.NoError(t, restored.Undo())
	assert.Equal(t, originalMoveCount-1, restored.GetMoveCount())
}

// TestSeahavenTowers_HistoryRespectsMaxSliceLen rejects oversized history arrays.
func TestSeahavenTowers_HistoryRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	big := make([]map[string]any, seahavenTowersMaxSliceLen+1)
	for i := range big {
		big[i] = map[string]any{}
	}
	payload := map[string]any{"tc": nil, "hi": big}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored SeahavenTowers
	require.Error(t, json.Unmarshal(data, &restored))
}

// TestSeahavenTowers_TopLevelTableauColumnRespectsMaxSliceLen rejects oversized tableau columns.
func TestSeahavenTowers_TopLevelTableauColumnRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigCol := make([]map[string]any, seahavenTowersMaxSliceLen+1)
	for i := range bigCol {
		bigCol[i] = map[string]any{}
	}
	cols := make([]any, SeahavenTowersTableauCnt)
	cols[0] = bigCol
	payload := map[string]any{"tc": nil, "tb": cols}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored SeahavenTowers
	require.Error(t, json.Unmarshal(data, &restored))
}

// TestSeahavenTowers_TopLevelFoundationPileRespectsMaxSliceLen rejects oversized foundation piles.
func TestSeahavenTowers_TopLevelFoundationPileRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigPile := make([]map[string]any, seahavenTowersMaxSliceLen+1)
	for i := range bigPile {
		bigPile[i] = map[string]any{}
	}
	piles := make([]any, SeahavenTowersFoundationCnt)
	piles[0] = bigPile
	payload := map[string]any{"tc": nil, "fd": piles}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored SeahavenTowers
	require.Error(t, json.Unmarshal(data, &restored))
}

// TestSeahavenTowers_SnapshotColumnsRespectMaxSliceLen rejects oversized
// inner tableau columns smuggled inside a history snapshot.
func TestSeahavenTowers_SnapshotColumnsRespectMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigCol := make([]map[string]any, seahavenTowersMaxSliceLen+1)
	for i := range bigCol {
		bigCol[i] = map[string]any{}
	}
	cols := make([]any, SeahavenTowersTableauCnt)
	cols[0] = bigCol
	snapshot := map[string]any{"tb": cols}
	payload := map[string]any{"tc": nil, "hi": []any{snapshot}}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored SeahavenTowers
	require.Error(t, json.Unmarshal(data, &restored))
}

// TestSeahavenTowers_SnapshotFoundationRespectsMaxSliceLen rejects oversized
// inner foundation piles smuggled inside a history snapshot.
func TestSeahavenTowers_SnapshotFoundationRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigPile := make([]map[string]any, seahavenTowersMaxSliceLen+1)
	for i := range bigPile {
		bigPile[i] = map[string]any{}
	}
	piles := make([]any, SeahavenTowersFoundationCnt)
	piles[0] = bigPile
	snapshot := map[string]any{"fd": piles}
	payload := map[string]any{"tc": nil, "hi": []any{snapshot}}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored SeahavenTowers
	require.Error(t, json.Unmarshal(data, &restored))
}

// TestSeahavenTowers_UnmarshalDefaultsAndActionLog covers the
// nil-history / nil-actionLog branches of UnmarshalJSON: a minimal payload
// without those arrays must still produce a usable game.
func TestSeahavenTowers_UnmarshalDefaultsAndActionLog(t *testing.T) {
	t.Parallel()

	payload := map[string]any{
		"tc": nil,
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored SeahavenTowers
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.NotNil(t, restored.GetActionLog())
	assert.NotNil(t, restored.history)
}
