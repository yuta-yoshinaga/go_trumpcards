//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAccordion_PersistsUndoHistory verifies issue #1654: when an
// Accordion game is round-tripped through JSON (e.g. a Cloudflare KV
// restore), the undo history must survive so the player can still step
// backward.
//
// Accordion's Move depends on rank/suit alignment of the random initial
// deal, so the test calls takeSnapshot() directly (same package). This
// still exercises the full Marshal/Unmarshal round-trip for the snapshot
// wire format.
func TestAccordion_PersistsUndoHistory(t *testing.T) {
	t.Parallel()

	original := NewDefaultAccordion()
	original.Reset()
	require.False(t, original.CanUndo(), "fresh game has no history")

	original.takeSnapshot()
	original.takeSnapshot()
	require.True(t, original.CanUndo(), "two snapshots should leave undoable history")
	originalHistoryLen := len(original.history)
	originalMoveCount := original.GetMoveCount()

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored Accordion
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, originalMoveCount, restored.GetMoveCount(), "moveCount must round-trip")
	assert.Equal(t, originalHistoryLen, len(restored.history), "history length must round-trip")
	assert.True(t, restored.CanUndo(), "restored game must allow Undo")
}

// TestAccordion_PersistsHistoryRestoresExactSnapshot ensures the snapshot
// fields are preserved exactly so an Undo on the restored game matches
// the pre-snapshot state.
func TestAccordion_PersistsHistoryRestoresExactSnapshot(t *testing.T) {
	t.Parallel()

	original := NewDefaultAccordion()
	original.Reset()

	// Snapshot the initial state, then mutate and snapshot again.
	original.takeSnapshot()
	preMutationPilesLen := len(original.piles)

	// Force a state change so Undo will visibly rewind.
	original.piles = original.piles[:len(original.piles)-1]
	original.takeSnapshot()

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored Accordion
	require.NoError(t, json.Unmarshal(data, &restored))

	// First Undo restores from second snapshot (post-mutation, one fewer pile);
	// second Undo restores from first snapshot (pre-mutation initial state).
	require.NoError(t, restored.Undo())
	require.NoError(t, restored.Undo())
	assert.Equal(t, preMutationPilesLen, len(restored.piles),
		"second Undo restores piles length")
}

// TestAccordion_HistoryRespectsMaxSliceLen rejects oversized history.
func TestAccordion_HistoryRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigHistory := make([]map[string]any, accordionMaxSliceLen+1)
	for i := range bigHistory {
		bigHistory[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"hi": bigHistory,
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Accordion
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized history must be rejected")
}

// TestAccordion_SnapshotPilesRespectsMaxSliceLen rejects payloads with an
// oversized Piles outer slice inside a history snapshot.
func TestAccordion_SnapshotPilesRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigPiles := make([]any, accordionMaxSliceLen+1)
	for i := range bigPiles {
		bigPiles[i] = []any{}
	}
	snapshot := map[string]any{
		"pl": bigPiles,
	}
	payload := map[string]any{
		"tc": nil,
		"hi": []any{snapshot},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Accordion
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized snapshot piles must be rejected")
}

// TestAccordion_SnapshotPileColumnRespectsMaxSliceLen rejects payloads
// with an oversized inner Pile entry inside a history snapshot.
func TestAccordion_SnapshotPileColumnRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigPile := make([]map[string]any, accordionMaxSliceLen+1)
	for i := range bigPile {
		bigPile[i] = map[string]any{}
	}
	snapshot := map[string]any{
		"pl": []any{bigPile},
	}
	payload := map[string]any{
		"tc": nil,
		"hi": []any{snapshot},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Accordion
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized snapshot pile entry must be rejected")
}
