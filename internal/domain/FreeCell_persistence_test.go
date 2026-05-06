//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFreeCell_PersistsUndoHistory verifies issue #1654: when a FreeCell
// game is round-tripped through JSON (e.g. a Cloudflare KV restore), the
// undo history must survive so the player can still step backward. Prior
// to the fix UnmarshalJSON set history = nil, leaving a restored game
// with `cannot undo: no history`.
func TestFreeCell_PersistsUndoHistory(t *testing.T) {
	t.Parallel()

	original := NewDefaultFreeCell()
	original.Reset()
	require.False(t, original.CanUndo(), "fresh game has no history")

	// Two deterministic moves: any tableau column has cards after Reset
	// and free cells 0/1 are empty, so these always succeed.
	require.NoError(t, original.MoveTableauToFreeCell(0, 0))
	require.NoError(t, original.MoveTableauToFreeCell(1, 1))
	require.True(t, original.CanUndo(), "two moves should leave undoable history")
	originalHistoryLen := len(original.history)
	originalMoveCount := original.GetMoveCount()

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored FreeCell
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, originalMoveCount, restored.GetMoveCount(), "moveCount must round-trip")
	assert.Equal(t, originalHistoryLen, len(restored.history), "history length must round-trip")
	assert.True(t, restored.CanUndo(), "restored game must allow Undo")

	require.NoError(t, restored.Undo())
	assert.Equal(t, originalMoveCount-1, restored.GetMoveCount(), "Undo should rewind one step")
}

// TestFreeCell_PersistsHistoryRestoresExactSnapshot ensures the snapshot
// fields (tableau, freeCells, foundation, moveCount) are preserved exactly
// so an Undo on a restored game returns to the same state the game would
// have had before the action that produced the snapshot.
func TestFreeCell_PersistsHistoryRestoresExactSnapshot(t *testing.T) {
	t.Parallel()

	original := NewDefaultFreeCell()
	original.Reset()
	require.NoError(t, original.MoveTableauToFreeCell(0, 0))

	// Capture the state we expect Undo on the restored game to recover.
	preMoveTableau0Len := len(original.tableau[0])
	preMoveFreeCell1 := original.freeCells[1]

	require.NoError(t, original.MoveTableauToFreeCell(1, 1))

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored FreeCell
	require.NoError(t, json.Unmarshal(data, &restored))

	require.NoError(t, restored.Undo())
	assert.Equal(t, preMoveTableau0Len, len(restored.tableau[0]), "Undo restores tableau length")
	assert.Equal(t, preMoveFreeCell1, restored.freeCells[1], "Undo restores free cell occupancy")
}

// TestFreeCell_HistoryRespectsMaxSliceLen rejects malicious payloads that
// attempt to exhaust memory through an oversized history array — same
// defence as the existing maxSliceLen guards on actionLog.
func TestFreeCell_HistoryRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigHistory := make([]map[string]any, freeCellMaxSliceLen+1)
	for i := range bigHistory {
		bigHistory[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"hi": bigHistory,
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored FreeCell
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized history must be rejected")
}

// TestFreeCell_SnapshotTableauColumnRespectsMaxSliceLen rejects payloads
// that smuggle an oversized inner Tableau column inside a history snapshot.
// Without the per-column guard inside freeCellSnapshot.UnmarshalJSON,
// 1000 history entries × 8 cols × 1000 cards would still allocate ~8M
// pointers despite the existing maxSliceLen check on the history length.
func TestFreeCell_SnapshotTableauColumnRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigCol := make([]map[string]any, freeCellMaxSliceLen+1)
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

	var restored FreeCell
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized snapshot tableau column must be rejected")
}

// TestFreeCell_SnapshotFoundationPileRespectsMaxSliceLen rejects payloads
// that smuggle an oversized inner Foundation pile inside a history snapshot.
func TestFreeCell_SnapshotFoundationPileRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigPile := make([]map[string]any, freeCellMaxSliceLen+1)
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

	var restored FreeCell
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized snapshot foundation pile must be rejected")
}
