//go:build test
// +build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActionLogEntry(t *testing.T) {
	card := NewCard(1, 5, true)
	entry := &ActionLogEntry{
		TurnNumber: 1,
		PlayerIdx:  0,
		ActionType: "play",
		Detail:     "played SPADE 5",
		Cards:      []*Card{card},
	}

	assert.Equal(t, 1, entry.TurnNumber)
	assert.Equal(t, 0, entry.PlayerIdx)
	assert.Equal(t, "play", entry.ActionType)
	assert.Equal(t, "played SPADE 5", entry.Detail)
	assert.Len(t, entry.Cards, 1)
	assert.Equal(t, card, entry.Cards[0])
}

func TestActionLogEntrySystemEvent(t *testing.T) {
	entry := &ActionLogEntry{
		TurnNumber: 1,
		PlayerIdx:  -1,
		ActionType: "result",
		Detail:     "game ended",
		Cards:      nil,
	}

	assert.Equal(t, -1, entry.PlayerIdx)
	assert.Equal(t, "result", entry.ActionType)
	assert.Nil(t, entry.Cards)
}

func TestActionLogEntryJSONDetailCodeRoundTrip(t *testing.T) {
	original := &ActionLogEntry{
		TurnNumber: 3, PlayerIdx: 1, ActionType: "score",
		DetailCode: "player.cpu", DetailParams: map[string]string{"id": "2"},
		Cards: []*Card{NewCard(CardDesignSpade, 7, true)},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"dc":"player.cpu"`)
	assert.Contains(t, string(data), `"dp":{"id":"2"}`)

	var decoded ActionLogEntry
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, original, &decoded)
}

func TestActionLogEntryJSONOmitsEmptyDetailParams(t *testing.T) {
	data, err := json.Marshal(&ActionLogEntry{TurnNumber: 1, PlayerIdx: -1, ActionType: "deal", DetailCode: "deal"})
	require.NoError(t, err)
	assert.NotContains(t, string(data), `"dp"`)
	assert.NotContains(t, string(data), `"d"`)
}

func TestActionLogEntryJSONReadsLegacyDetail(t *testing.T) {
	var entry ActionLogEntry
	require.NoError(t, json.Unmarshal([]byte(`{"t":2,"p":0,"a":"play","d":"旧文言","c":null}`), &entry))
	assert.Equal(t, 2, entry.TurnNumber)
	assert.Equal(t, "旧文言", entry.Detail)
	assert.Empty(t, entry.DetailCode)
	assert.Nil(t, entry.DetailParams)
}
