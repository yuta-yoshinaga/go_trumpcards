//go:build test
// +build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
		TurnNumber: 0,
		PlayerIdx:  -1,
		ActionType: "result",
		Detail:     "game ended",
		Cards:      nil,
	}

	assert.Equal(t, -1, entry.PlayerIdx)
	assert.Equal(t, "result", entry.ActionType)
	assert.Nil(t, entry.Cards)
}
