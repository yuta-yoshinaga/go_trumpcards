//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCruel_PersistsUndoHistory verifies that when a Cruel game is
// round-tripped through JSON (e.g. a Cloudflare KV restore), undo history
// survives so the player can still step backward.
func TestCruel_PersistsUndoHistory(t *testing.T) {
	t.Parallel()

	original := NewDefaultCruel()
	original.Reset()
	require.False(t, original.CanUndo(), "fresh game has no history")

	original.takeSnapshot()
	original.takeSnapshot()
	require.True(t, original.CanUndo(), "two snapshots should leave undoable history")
	originalHistoryLen := len(original.history)
	originalMoveCount := original.GetMoveCount()

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored Cruel
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, originalMoveCount, restored.GetMoveCount(), "moveCount must round-trip")
	assert.Equal(t, originalHistoryLen, len(restored.history), "history length must round-trip")
	assert.True(t, restored.CanUndo(), "restored game must allow Undo")
}

// TestCruel_PersistsHistoryRestoresExactSnapshot ensures snapshot fields
// are preserved so an Undo on the restored game matches the pre-snapshot state.
func TestCruel_PersistsHistoryRestoresExactSnapshot(t *testing.T) {
	t.Parallel()

	original := NewDefaultCruel()
	original.Reset()

	original.takeSnapshot()
	preMutationLen := len(original.tableau[0])

	original.tableau[0] = original.tableau[0][:0]
	original.takeSnapshot()

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored Cruel
	require.NoError(t, json.Unmarshal(data, &restored))

	require.NoError(t, restored.Undo())
	require.NoError(t, restored.Undo())
	assert.Equal(t, preMutationLen, len(restored.tableau[0]),
		"second Undo restores tableau[0] length")
}

// TestCruel_HistoryRespectsMaxSliceLen rejects oversized history payloads.
func TestCruel_HistoryRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigHistory := make([]map[string]any, cruelMaxSliceLen+1)
	for i := range bigHistory {
		bigHistory[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"hi": bigHistory,
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Cruel
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized history must be rejected")
}

// TestCruel_TopLevelTableauColumnRespectsMaxSliceLen rejects payloads with
// an oversized tableau column at the top level.
func TestCruel_TopLevelTableauColumnRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigCol := make([]map[string]any, cruelMaxSliceLen+1)
	for i := range bigCol {
		bigCol[i] = map[string]any{}
	}
	tableau := make([]any, CruelTableauCnt)
	tableau[0] = bigCol
	payload := map[string]any{
		"tc": nil,
		"tb": tableau,
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Cruel
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized top-level tableau column must be rejected")
}

// TestCruel_TopLevelFoundationPileRespectsMaxSliceLen rejects payloads
// with an oversized foundation pile at the top level.
func TestCruel_TopLevelFoundationPileRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigPile := make([]map[string]any, cruelMaxSliceLen+1)
	for i := range bigPile {
		bigPile[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"fd": []any{bigPile, nil, nil, nil},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Cruel
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized top-level foundation pile must be rejected")
}

// TestCruel_SnapshotTableauColumnRespectsMaxSliceLen rejects payloads with
// an oversized tableau column inside a history snapshot.
func TestCruel_SnapshotTableauColumnRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigCol := make([]map[string]any, cruelMaxSliceLen+1)
	for i := range bigCol {
		bigCol[i] = map[string]any{}
	}
	tableau := make([]any, CruelTableauCnt)
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

	var restored Cruel
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized snapshot tableau column must be rejected")
}

// TestCruel_SnapshotFoundationPileRespectsMaxSliceLen rejects payloads
// with an oversized foundation pile inside a history snapshot.
func TestCruel_SnapshotFoundationPileRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigPile := make([]map[string]any, cruelMaxSliceLen+1)
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

	var restored Cruel
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized snapshot foundation pile must be rejected")
}

// TestCruel_TopLevelTableauNilCardRejected ensures a JSON payload with a
// `null` tableau-card entry is rejected, preventing nil dereference panics
// in downstream code (presenters, stalemate check, undo restore).
func TestCruel_TopLevelTableauNilCardRejected(t *testing.T) {
	t.Parallel()

	tableau := make([]any, CruelTableauCnt)
	tableau[0] = []any{nil}
	payload := map[string]any{
		"tc": nil,
		"tb": tableau,
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Cruel
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "nil tableau card must be rejected")
	assert.Contains(t, err.Error(), "nil card")
}

// TestCruel_TopLevelFoundationNilCardRejected does the same for foundations.
func TestCruel_TopLevelFoundationNilCardRejected(t *testing.T) {
	t.Parallel()

	payload := map[string]any{
		"tc": nil,
		"fd": []any{[]any{nil}, nil, nil, nil},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Cruel
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "nil foundation card must be rejected")
	assert.Contains(t, err.Error(), "nil card")
}

// TestCruel_SnapshotTableauNilCardRejected guards the same surface inside
// history snapshots — these go through cruelSnapshot.UnmarshalJSON.
func TestCruel_SnapshotTableauNilCardRejected(t *testing.T) {
	t.Parallel()

	tableau := make([]any, CruelTableauCnt)
	tableau[0] = []any{nil}
	snapshot := map[string]any{"tb": tableau}
	payload := map[string]any{
		"tc": nil,
		"hi": []any{snapshot},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Cruel
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "nil snapshot tableau card must be rejected")
	assert.Contains(t, err.Error(), "nil card")
}

// TestCruel_SnapshotFoundationNilCardRejected mirrors the above for snapshot foundations.
func TestCruel_SnapshotFoundationNilCardRejected(t *testing.T) {
	t.Parallel()

	snapshot := map[string]any{"fd": []any{[]any{nil}, nil, nil, nil}}
	payload := map[string]any{
		"tc": nil,
		"hi": []any{snapshot},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Cruel
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "nil snapshot foundation card must be rejected")
	assert.Contains(t, err.Error(), "nil card")
}

// TestCruel_RoundTripEmptyState ensures an unstarted game (nil action log,
// nil history) roundtrips cleanly without panicking.
func TestCruel_RoundTripEmptyState(t *testing.T) {
	t.Parallel()

	original := NewDefaultCruel()
	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored Cruel
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.NotNil(t, restored.GetActionLog())
}
