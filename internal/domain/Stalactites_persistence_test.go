//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStalactites_PersistsUndoHistory verifies issue #1654: when a Stalactites
// game is round-tripped through JSON (e.g. a Cloudflare KV restore), the
// undo history must survive so the player can still step backward. Prior
// to the fix UnmarshalJSON set history = nil, leaving a restored game
// with `cannot undo: no history`.
func TestStalactites_PersistsUndoHistory(t *testing.T) {
	t.Parallel()

	original := NewDefaultStalactites()
	original.Reset()
	require.False(t, original.CanUndo(), "fresh game has no history")

	// Two deterministic moves. Stalactites deals INTO the cells, so unlike
	// FreeCell none is free after Reset -- empty two of them first, or both
	// moves fail with "free cell is occupied".
	original.cells[0] = nil
	original.cells[1] = nil
	require.NoError(t, original.MoveTableauToStalactites(0, 0))
	require.NoError(t, original.MoveTableauToStalactites(1, 1))
	require.True(t, original.CanUndo(), "two moves should leave undoable history")
	originalHistoryLen := len(original.history)
	originalMoveCount := original.GetMoveCount()

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored Stalactites
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, originalMoveCount, restored.GetMoveCount(), "moveCount must round-trip")
	assert.Equal(t, originalHistoryLen, len(restored.history), "history length must round-trip")
	assert.True(t, restored.CanUndo(), "restored game must allow Undo")

	require.NoError(t, restored.Undo())
	assert.Equal(t, originalMoveCount-1, restored.GetMoveCount(), "Undo should rewind one step")
}

// TestStalactites_PersistsHistoryRestoresExactSnapshot ensures the snapshot
// fields (tableau, cells, foundation, moveCount) are preserved exactly
// so an Undo on a restored game returns to the same state the game would
// have had before the action that produced the snapshot.
func TestStalactites_PersistsHistoryRestoresExactSnapshot(t *testing.T) {
	t.Parallel()

	original := NewDefaultStalactites()
	original.Reset()
	// Stalactites deals INTO the cells, so free the two this test moves into.
	original.cells[0] = nil
	original.cells[1] = nil
	require.NoError(t, original.MoveTableauToStalactites(0, 0))

	// Capture the state we expect Undo on the restored game to recover.
	preMoveTableau0Len := len(original.tableau[0])
	preMoveStalactites1 := original.cells[1]

	require.NoError(t, original.MoveTableauToStalactites(1, 1))

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored Stalactites
	require.NoError(t, json.Unmarshal(data, &restored))

	require.NoError(t, restored.Undo())
	assert.Equal(t, preMoveTableau0Len, len(restored.tableau[0]), "Undo restores tableau length")
	assert.Equal(t, preMoveStalactites1, restored.cells[1], "Undo restores free cell occupancy")
}

// TestStalactites_HistoryRespectsMaxSliceLen rejects malicious payloads that
// attempt to exhaust memory through an oversized history array — same
// defence as the existing maxSliceLen guards on actionLog.
func TestStalactites_HistoryRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigHistory := make([]map[string]any, stalactitesMaxSliceLen+1)
	for i := range bigHistory {
		bigHistory[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"hi": bigHistory,
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Stalactites
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized history must be rejected")
}

// TestStalactites_SnapshotTableauColumnRespectsMaxSliceLen rejects payloads
// that smuggle an oversized inner Tableau column inside a history snapshot.
// Without the per-column guard inside stalactitesSnapshot.UnmarshalJSON,
// 1000 history entries × 8 cols × 1000 cards would still allocate ~8M
// pointers despite the existing maxSliceLen check on the history length.
func TestStalactites_SnapshotTableauColumnRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigCol := make([]map[string]any, stalactitesMaxSliceLen+1)
	for i := range bigCol {
		bigCol[i] = map[string]any{}
	}
	snapshot := map[string]any{
		"tb": []any{bigCol, nil, nil, nil, nil, nil, nil, nil},
	}
	payload := map[string]any{
		"tc": nil,
		"hi": []any{snapshot},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Stalactites
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized snapshot tableau column must be rejected")
}

// TestStalactites_TopLevelTableauColumnRespectsMaxSliceLen rejects payloads
// with an oversized Tableau column at the top level (not nested inside a
// snapshot). Mirrors the snapshot-level guard so callers cannot bypass
// the OOM defence by inflating the live tableau directly.
func TestStalactites_TopLevelTableauColumnRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigCol := make([]map[string]any, stalactitesMaxSliceLen+1)
	for i := range bigCol {
		bigCol[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"tb": []any{bigCol, nil, nil, nil, nil, nil, nil, nil},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Stalactites
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized top-level tableau column must be rejected")
}

// TestStalactites_TopLevelFoundationPileRespectsMaxSliceLen rejects payloads
// with an oversized Foundation pile at the top level.
func TestStalactites_TopLevelFoundationPileRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigPile := make([]map[string]any, stalactitesMaxSliceLen+1)
	for i := range bigPile {
		bigPile[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"fd": []any{bigPile, nil, nil, nil},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Stalactites
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized top-level foundation pile must be rejected")
}

// TestStalactites_SnapshotFoundationPileRespectsMaxSliceLen rejects payloads
// that smuggle an oversized inner Foundation pile inside a history snapshot.
func TestStalactites_SnapshotFoundationPileRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigPile := make([]map[string]any, stalactitesMaxSliceLen+1)
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

	var restored Stalactites
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized snapshot foundation pile must be rejected")
}
