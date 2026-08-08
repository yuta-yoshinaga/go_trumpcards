//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestActionLogBase_AppendNumbersFromOne(t *testing.T) {
	var b actionLogBase

	assert.Empty(t, b.GetActionLog(), "a fresh log is empty")

	b.appendLog(0, "play", "first", nil)
	b.appendLog(2, "pass", "second", []*Card{NewCard(0, 1, true)})

	got := b.GetActionLog()
	assert.Len(t, got, 2)

	assert.Equal(t, 1, got[0].TurnNumber, "first entry is turn 1, not 0")
	assert.Equal(t, 0, got[0].PlayerIdx)
	assert.Equal(t, "play", got[0].ActionType)
	assert.Equal(t, "first", got[0].Detail)
	assert.Nil(t, got[0].Cards)

	assert.Equal(t, 2, got[1].TurnNumber)
	assert.Equal(t, 2, got[1].PlayerIdx)
	assert.Len(t, got[1].Cards, 1)
}

// Reset implementations clear the log by assigning to the promoted field. That
// only compiles and behaves correctly because the field is promoted, so it is
// worth pinning: renaming or unexporting it differently would break ~122 games
// at once.
func TestActionLogBase_PromotedFieldStaysAssignable(t *testing.T) {
	type game struct {
		actionLogBase
		name string
	}
	g := &game{name: "x"}

	g.appendLog(0, "a", "b", nil)
	assert.Len(t, g.GetActionLog(), 1)

	g.actionLog = nil
	assert.Empty(t, g.GetActionLog(), "assigning nil through the promoted field clears it")

	g.appendLog(1, "c", "d", nil)
	assert.Equal(t, 1, g.GetActionLog()[0].TurnNumber, "numbering restarts after a clear")
}
