//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEightOff_PersistsUndoHistory verifies undo history survives JSON round-trip.
func TestEightOff_PersistsUndoHistory(t *testing.T) {
	t.Parallel()

	original := NewDefaultEightOff()
	original.Reset()
	require.False(t, original.CanUndo())

	// Move the bottom card of column 0 into free cell 4 (always empty after Reset).
	require.NoError(t, original.MoveTableauToFreeCell(0, 4))
	require.NoError(t, original.MoveTableauToFreeCell(1, 5))
	require.True(t, original.CanUndo())
	originalHistoryLen := len(original.history)
	originalMoveCount := original.GetMoveCount()

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored EightOff
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, originalMoveCount, restored.GetMoveCount())
	assert.Equal(t, originalHistoryLen, len(restored.history))
	assert.True(t, restored.CanUndo())

	require.NoError(t, restored.Undo())
	assert.Equal(t, originalMoveCount-1, restored.GetMoveCount())
}

func TestEightOff_PersistsHistoryRestoresExactSnapshot(t *testing.T) {
	t.Parallel()

	original := NewDefaultEightOff()
	original.Reset()
	require.NoError(t, original.MoveTableauToFreeCell(0, 4))

	preMoveTableau0Len := len(original.tableau[0])
	preMoveFreeCell5 := original.freeCells[5]

	require.NoError(t, original.MoveTableauToFreeCell(1, 5))

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored EightOff
	require.NoError(t, json.Unmarshal(data, &restored))

	require.NoError(t, restored.Undo())
	assert.Equal(t, preMoveTableau0Len, len(restored.tableau[0]))
	assert.Equal(t, preMoveFreeCell5, restored.freeCells[5])
}

func TestEightOff_HistoryRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigHistory := make([]map[string]any, eightOffMaxSliceLen+1)
	for i := range bigHistory {
		bigHistory[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"hi": bigHistory,
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored EightOff
	err = json.Unmarshal(data, &restored)
	require.Error(t, err)
}

func TestEightOff_TopLevelTableauColumnRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigCol := make([]map[string]any, eightOffMaxSliceLen+1)
	for i := range bigCol {
		bigCol[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"tb": []any{bigCol, nil, nil, nil, nil, nil, nil, nil},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored EightOff
	err = json.Unmarshal(data, &restored)
	require.Error(t, err)
}

func TestEightOff_TopLevelFoundationPileRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigPile := make([]map[string]any, eightOffMaxSliceLen+1)
	for i := range bigPile {
		bigPile[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"fd": []any{bigPile, nil, nil, nil},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored EightOff
	err = json.Unmarshal(data, &restored)
	require.Error(t, err)
}

func TestEightOff_SnapshotTableauColumnRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigCol := make([]map[string]any, eightOffMaxSliceLen+1)
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

	var restored EightOff
	err = json.Unmarshal(data, &restored)
	require.Error(t, err)
}

func TestEightOff_SnapshotFoundationPileRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigPile := make([]map[string]any, eightOffMaxSliceLen+1)
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

	var restored EightOff
	err = json.Unmarshal(data, &restored)
	require.Error(t, err)
}
