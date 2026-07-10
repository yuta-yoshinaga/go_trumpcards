//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSultan_PersistsUndoHistory verifies that when a Sultan game is
// round-tripped through JSON (e.g. a Cloudflare KV restore), the undo history
// survives so the player can still step backward.
func TestSultan_PersistsUndoHistory(t *testing.T) {
	t.Parallel()

	original := NewDefaultSultan()
	original.Reset()
	require.False(t, original.CanUndo(), "fresh game has no history")

	require.NoError(t, original.Draw())
	require.NoError(t, original.Draw())
	require.True(t, original.CanUndo())
	originalHistoryLen := len(original.history)
	originalMoveCount := original.GetMoveCount()

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored Sultan
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, originalMoveCount, restored.GetMoveCount())
	assert.Equal(t, originalHistoryLen, len(restored.history))
	assert.True(t, restored.CanUndo())

	require.NoError(t, restored.Undo())
	assert.Equal(t, originalMoveCount-1, restored.GetMoveCount())
}

// TestSultan_PersistsExactSnapshot ensures snapshot fields are preserved exactly.
func TestSultan_PersistsExactSnapshot(t *testing.T) {
	t.Parallel()

	original := NewDefaultSultan()
	original.Reset()
	require.NoError(t, original.Draw())

	preWasteLen := len(original.waste)
	preStockLen := len(original.stock)

	require.NoError(t, original.Draw())

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored Sultan
	require.NoError(t, json.Unmarshal(data, &restored))

	require.NoError(t, restored.Undo())
	assert.Equal(t, preWasteLen, len(restored.waste))
	assert.Equal(t, preStockLen, len(restored.stock))
}

// TestSultan_PreservesDivanRoundTrip ensures the divan (including nil played
// slots) round-trips through JSON.
func TestSultan_PreservesDivanRoundTrip(t *testing.T) {
	t.Parallel()

	original := NewDefaultSultan()
	original.Reset()
	// Force a played, unrefilled divan slot by emptying stock and replacing a
	// divan card with one that lands on a foundation.
	original.SetStock(nil)
	original.SetFoundation([SultanFoundationCnt][]*Card{
		{NewCard(CardDesignSpade, CardValueMax, false)},
		{NewCard(CardDesignClover, CardValueMax, false)},
		{NewCard(CardDesignHeart, CardValueMax, false)},
		{NewCard(CardDesignDiamond, CardValueMax, false)},
		{NewCard(CardDesignSpade, CardValueMax, false)},
		{NewCard(CardDesignClover, CardValueMax, false)},
		{NewCard(CardDesignHeart, CardValueMax, false)},
		{NewCard(CardDesignDiamond, CardValueMax, false)},
	})
	divan := make([]*Card, SultanDivanCnt)
	divan[0] = NewCard(CardDesignSpade, 1, false)
	original.SetDivan(divan)
	require.NoError(t, original.MoveDivanToFoundation(0))
	require.Nil(t, original.GetDivan()[0])

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored Sultan
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Len(t, restored.GetDivan(), SultanDivanCnt)
	assert.Nil(t, restored.GetDivan()[0], "nil divan slot must survive round-trip")
}

// TestSultan_HistoryRespectsMaxSliceLen rejects oversized history.
func TestSultan_HistoryRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigHistory := make([]map[string]any, sultanMaxSliceLen+1)
	for i := range bigHistory {
		bigHistory[i] = map[string]any{}
	}
	payload := map[string]any{"tc": nil, "hi": bigHistory}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Sultan
	require.Error(t, json.Unmarshal(data, &restored))
}

// TestSultan_SnapshotStockRespectsMaxSliceLen rejects oversized snapshot stock.
func TestSultan_SnapshotStockRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigStock := make([]map[string]any, sultanMaxSliceLen+1)
	for i := range bigStock {
		bigStock[i] = map[string]any{}
	}
	payload := map[string]any{"tc": nil, "hi": []any{map[string]any{"st": bigStock}}}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Sultan
	require.Error(t, json.Unmarshal(data, &restored))
}

// TestSultan_SnapshotDivanRespectsMaxSliceLen rejects oversized snapshot divan.
func TestSultan_SnapshotDivanRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigDivan := make([]map[string]any, sultanMaxSliceLen+1)
	for i := range bigDivan {
		bigDivan[i] = map[string]any{}
	}
	payload := map[string]any{"tc": nil, "hi": []any{map[string]any{"dv": bigDivan}}}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Sultan
	require.Error(t, json.Unmarshal(data, &restored))
}

// TestSultan_TopLevelFoundationPileRespectsMaxSliceLen rejects an oversized
// top-level foundation pile.
func TestSultan_TopLevelFoundationPileRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigPile := make([]map[string]any, sultanMaxSliceLen+1)
	for i := range bigPile {
		bigPile[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"fd": []any{bigPile, nil, nil, nil, nil, nil, nil, nil},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Sultan
	require.Error(t, json.Unmarshal(data, &restored))
}

// TestSultan_TopLevelDivanRespectsMaxSliceLen rejects an oversized top-level divan.
func TestSultan_TopLevelDivanRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigDivan := make([]map[string]any, sultanMaxSliceLen+1)
	for i := range bigDivan {
		bigDivan[i] = map[string]any{}
	}
	payload := map[string]any{"tc": nil, "dv": bigDivan}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Sultan
	require.Error(t, json.Unmarshal(data, &restored))
}

// TestSultan_RejectsNilStockCard ensures top-level UnmarshalJSON rejects a nil
// stock element.
func TestSultan_RejectsNilStockCard(t *testing.T) {
	t.Parallel()

	payload := map[string]any{"tc": nil, "st": []any{nil}}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Sultan
	require.Error(t, json.Unmarshal(data, &restored))
}

// TestSultan_RejectsNilWasteCard ensures top-level UnmarshalJSON rejects a nil
// waste element.
func TestSultan_RejectsNilWasteCard(t *testing.T) {
	t.Parallel()

	payload := map[string]any{"tc": nil, "wa": []any{nil}}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Sultan
	require.Error(t, json.Unmarshal(data, &restored))
}

// TestSultan_RejectsNilFoundationCard ensures top-level UnmarshalJSON rejects a
// nil foundation element.
func TestSultan_RejectsNilFoundationCard(t *testing.T) {
	t.Parallel()

	payload := map[string]any{
		"tc": nil,
		"fd": []any{[]any{nil}, nil, nil, nil, nil, nil, nil, nil},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Sultan
	require.Error(t, json.Unmarshal(data, &restored))
}

// TestSultan_SnapshotRejectsNilStockCard ensures snapshot UnmarshalJSON rejects
// a nil stock element.
func TestSultan_SnapshotRejectsNilStockCard(t *testing.T) {
	t.Parallel()

	payload := map[string]any{"tc": nil, "hi": []any{map[string]any{"st": []any{nil}}}}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Sultan
	require.Error(t, json.Unmarshal(data, &restored))
}

// TestSultan_AllowsNilDivanSlot ensures a nil divan slot is accepted (a played,
// unrefilled slot is intentionally nil).
func TestSultan_AllowsNilDivanSlot(t *testing.T) {
	t.Parallel()

	payload := map[string]any{"tc": nil, "dv": []any{nil}}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored Sultan
	require.NoError(t, json.Unmarshal(data, &restored), "nil divan slot is valid")
}
