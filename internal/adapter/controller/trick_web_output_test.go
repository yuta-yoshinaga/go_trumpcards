//go:build test

package controller

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWebOutputTrickCard_WireShape pins the REST field names.
//
// 60 games' trick display now serializes through this one type (issue #4432),
// each of which previously declared its own struct with these exact tags.
// Renaming either would change the REST response for all 60 at once, and the
// frontend reads `playerIdx` / `card` directly — so the failure would surface as
// blank trick cards in the browser, not as a Go test failure.
func TestWebOutputTrickCard_WireShape(t *testing.T) {
	data, err := json.Marshal(&WebOutputTrickCard{
		PlayerIdx: 2,
		Card:      &WebOutputCard{Design: "SPADE", Value: 11},
	})
	require.NoError(t, err)

	var raw map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &raw))
	assert.Len(t, raw, 2, "must serialize exactly playerIdx + card; got %s", data)
	assert.Contains(t, raw, "playerIdx")
	assert.Contains(t, raw, "card")
}

// TestWebOutputCardHint_WireShape pins the same contract for the shared hint.
//
// Only 22 of the 115 *WebOutputHint types share this shape; the rest are
// genuinely different (solitaire zone/column moves, bid hints) and keep their
// own declarations. See the type's doc comment.
func TestWebOutputCardHint_WireShape(t *testing.T) {
	data, err := json.Marshal(&WebOutputCardHint{CardIndices: []int{0, 3}, Reason: "lead"})
	require.NoError(t, err)

	var raw map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &raw))
	assert.Len(t, raw, 2, "must serialize exactly cardIndices + reason; got %s", data)
	assert.Contains(t, raw, "cardIndices")
	assert.Contains(t, raw, "reason")
}
