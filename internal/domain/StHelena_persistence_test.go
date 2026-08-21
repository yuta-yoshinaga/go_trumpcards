//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStHelena_PersistsUndoHistory ensures the JSON wire format round-trips
// the undo history so a player who reloaded a session can still Undo.
func TestStHelena_PersistsUndoHistory(t *testing.T) {
	t.Parallel()

	original := NewDefaultStHelena()
	original.Reset()
	require.False(t, original.CanUndo())
	require.NoError(t, original.Redeal())
	require.NoError(t, original.Redeal())
	require.True(t, original.CanUndo())

	originalHistoryLen := len(original.history)
	originalMoveCount := original.GetMoveCount()
	originalRedeals := original.GetRedealsRemaining()

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored StHelena
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, originalMoveCount, restored.GetMoveCount())
	assert.Equal(t, originalRedeals, restored.GetRedealsRemaining())
	assert.Equal(t, originalHistoryLen, len(restored.history))
	require.True(t, restored.CanUndo())

	require.NoError(t, restored.Undo())
	assert.Equal(t, originalRedeals+1, restored.GetRedealsRemaining(), "Undo restores redeal counter")
}

// TestStHelena_PersistsSnapshotExact restores the exact snapshot pre-move.
func TestStHelena_PersistsSnapshotExact(t *testing.T) {
	t.Parallel()

	original := NewDefaultStHelena()
	original.Reset()

	preRedeals := original.GetRedealsRemaining()
	preTabRef := original.GetTableau()
	var preTab [StHelenaTableauCnt][]*StHelenaTableauCard
	for i := range StHelenaTableauCnt {
		preTab[i] = make([]*StHelenaTableauCard, len(preTabRef[i]))
		for j, tc := range preTabRef[i] {
			preTab[i][j] = &StHelenaTableauCard{Card: NewCard(tc.Card.GetDesign(), tc.Card.GetValue(), false), FaceUp: tc.FaceUp}
		}
	}

	require.NoError(t, original.Redeal())

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored StHelena
	require.NoError(t, json.Unmarshal(data, &restored))

	require.NoError(t, restored.Undo())
	assert.Equal(t, preRedeals, restored.GetRedealsRemaining())
	got := restored.GetTableau()
	for i := range StHelenaTableauCnt {
		require.Len(t, got[i], len(preTab[i]))
		for j := range got[i] {
			assert.Equal(t, preTab[i][j].Card.GetDesign(), got[i][j].Card.GetDesign())
			assert.Equal(t, preTab[i][j].Card.GetValue(), got[i][j].Card.GetValue())
		}
	}
}

// TestStHelena_HistoryRespectsMaxSliceLen rejects oversized history.
func TestStHelena_HistoryRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigHistory := make([]map[string]any, stHelenaMaxSliceLen+1)
	for i := range bigHistory {
		bigHistory[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"hi": bigHistory,
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored StHelena
	err = json.Unmarshal(data, &restored)
	require.Error(t, err)
}

// TestStHelena_SnapshotTableauColumnRespectsMaxSliceLen rejects oversized snapshot tableau.
func TestStHelena_SnapshotTableauColumnRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigCol := make([]map[string]any, stHelenaMaxSliceLen+1)
	for i := range bigCol {
		bigCol[i] = map[string]any{}
	}
	tableau := make([]any, StHelenaTableauCnt)
	tableau[0] = bigCol
	snapshot := map[string]any{"tb": tableau}
	payload := map[string]any{
		"tc": nil,
		"hi": []any{snapshot},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored StHelena
	err = json.Unmarshal(data, &restored)
	require.Error(t, err)
}

// TestStHelena_TopLevelTableauColumnRespectsMaxSliceLen rejects oversized top-level tableau.
func TestStHelena_TopLevelTableauColumnRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigCol := make([]map[string]any, stHelenaMaxSliceLen+1)
	for i := range bigCol {
		bigCol[i] = map[string]any{}
	}
	tableau := make([]any, StHelenaTableauCnt)
	tableau[0] = bigCol
	payload := map[string]any{
		"tc": nil,
		"tb": tableau,
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored StHelena
	err = json.Unmarshal(data, &restored)
	require.Error(t, err)
}

// TestStHelena_TopLevelFoundationPileRespectsMaxSliceLen rejects oversized foundation.
func TestStHelena_TopLevelFoundationPileRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigPile := make([]map[string]any, stHelenaMaxSliceLen+1)
	for i := range bigPile {
		bigPile[i] = map[string]any{}
	}
	foundation := make([]any, StHelenaFoundationCnt)
	foundation[0] = bigPile
	payload := map[string]any{
		"tc": nil,
		"fd": foundation,
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored StHelena
	err = json.Unmarshal(data, &restored)
	require.Error(t, err)
}

// TestStHelena_UnmarshalNilTrumpCardsDefaults ensures that a payload with a nil
// "tc" key produces a usable StHelena (deck regenerated with two combined decks).
func TestStHelena_UnmarshalNilTrumpCardsDefaults(t *testing.T) {
	t.Parallel()

	payload := map[string]any{
		"tc": nil,
		"tb": []any{nil},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored StHelena
	require.NoError(t, json.Unmarshal(data, &restored))
	require.NotNil(t, restored.trumpCards)
}
