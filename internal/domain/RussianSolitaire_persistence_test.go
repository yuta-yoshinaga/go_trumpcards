//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRussianSolitaire_PersistsUndoHistory verifies issue #1654: when a
// RussianSolitaire game is round-tripped through JSON (e.g. a Cloudflare
// KV restore), the undo history must survive so the player can still
// step backward.
//
// RussianSolitaire has no deterministic public action — every Move is
// shape-dependent on the random initial deal — so the test calls
// takeSnapshot() directly (same package). This still exercises the full
// Marshal/Unmarshal round-trip for the snapshot wire format.
func TestRussianSolitaire_PersistsUndoHistory(t *testing.T) {
	t.Parallel()

	original := NewDefaultRussianSolitaire()
	original.Reset()
	require.False(t, original.CanUndo(), "fresh game has no history")

	original.takeSnapshot()
	original.takeSnapshot()
	require.True(t, original.CanUndo(), "two snapshots should leave undoable history")
	originalHistoryLen := len(original.history)
	originalMoveCount := original.GetMoveCount()

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored RussianSolitaire
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, originalMoveCount, restored.GetMoveCount(), "moveCount must round-trip")
	assert.Equal(t, originalHistoryLen, len(restored.history), "history length must round-trip")
	assert.True(t, restored.CanUndo(), "restored game must allow Undo")
}

// TestRussianSolitaire_PersistsHistoryRestoresExactSnapshot ensures the
// snapshot fields are preserved exactly so an Undo on the restored game
// matches the pre-snapshot state.
func TestRussianSolitaire_PersistsHistoryRestoresExactSnapshot(t *testing.T) {
	t.Parallel()

	original := NewDefaultRussianSolitaire()
	original.Reset()

	// Snapshot the initial state, then mutate and snapshot again.
	original.takeSnapshot()
	preMutationTableau0Len := len(original.tableau[0])

	// Force a state change so Undo will visibly rewind.
	original.tableau[0] = original.tableau[0][:0]
	original.takeSnapshot()

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored RussianSolitaire
	require.NoError(t, json.Unmarshal(data, &restored))

	// The first Undo restores from the second snapshot (post-mutation, with
	// tableau[0] empty); the second Undo restores from the first snapshot
	// (pre-mutation initial state). Only after both pops should tableau[0]
	// regain its original length.
	require.NoError(t, restored.Undo())
	require.NoError(t, restored.Undo())
	assert.Equal(t, preMutationTableau0Len, len(restored.tableau[0]),
		"second Undo restores tableau[0] length")
}

// TestRussianSolitaire_HistoryRespectsMaxSliceLen rejects oversized history.
func TestRussianSolitaire_HistoryRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigHistory := make([]map[string]any, russianSolitaireMaxSliceLen+1)
	for i := range bigHistory {
		bigHistory[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"hi": bigHistory,
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored RussianSolitaire
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized history must be rejected")
}

// TestRussianSolitaire_TopLevelTableauColumnRespectsMaxSliceLen rejects
// payloads with an oversized Tableau column at the top level.
func TestRussianSolitaire_TopLevelTableauColumnRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigCol := make([]map[string]any, russianSolitaireMaxSliceLen+1)
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

	var restored RussianSolitaire
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized top-level tableau column must be rejected")
}

// TestRussianSolitaire_TopLevelFoundationPileRespectsMaxSliceLen rejects
// payloads with an oversized Foundation pile at the top level.
func TestRussianSolitaire_TopLevelFoundationPileRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigPile := make([]map[string]any, russianSolitaireMaxSliceLen+1)
	for i := range bigPile {
		bigPile[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"fd": []any{bigPile, nil, nil, nil},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored RussianSolitaire
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized top-level foundation pile must be rejected")
}

// TestRussianSolitaire_SnapshotTableauColumnRespectsMaxSliceLen rejects
// payloads with an oversized tableau column inside a history snapshot.
func TestRussianSolitaire_SnapshotTableauColumnRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigCol := make([]map[string]any, russianSolitaireMaxSliceLen+1)
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

	var restored RussianSolitaire
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized snapshot tableau column must be rejected")
}

// TestRussianSolitaire_SnapshotFoundationPileRespectsMaxSliceLen rejects
// payloads with an oversized foundation pile inside a history snapshot.
func TestRussianSolitaire_SnapshotFoundationPileRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigPile := make([]map[string]any, russianSolitaireMaxSliceLen+1)
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

	var restored RussianSolitaire
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized snapshot foundation pile must be rejected")
}
