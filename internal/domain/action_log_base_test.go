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

// appendLogAt is what lets the solitaires share this base: they number entries
// by move count, not by how many entries exist, so the turn number has to come
// from the caller. Pinning it separately from appendLog because the two now
// share a body and a regression in the shared half would be easy to miss.
func TestActionLogBase_AppendLogAtUsesCallerTurnNumber(t *testing.T) {
	var b actionLogBase

	b.appendLogAt(7, 0, "move", "t1->f0", nil)
	b.appendLogAt(7, 0, "move", "same turn again", nil)
	b.appendLogAt(9, 3, "draw", "later", nil)

	got := b.GetActionLog()
	assert.Len(t, got, 3)
	assert.Equal(t, 7, got[0].TurnNumber)
	assert.Equal(t, 7, got[1].TurnNumber, "the caller's number is used verbatim, not incremented")
	assert.Equal(t, 9, got[2].TurnNumber, "gaps are preserved")
	assert.Equal(t, 3, got[2].PlayerIdx)
}

// appendLog must keep numbering from the entry count even though it now
// delegates -- that is the contract the 122 phase-1 games depend on.
func TestActionLogBase_AppendLogStillNumbersByCount(t *testing.T) {
	var b actionLogBase
	b.appendLogAt(99, 0, "seed", "", nil)
	b.appendLog(0, "next", "", nil)

	got := b.GetActionLog()
	assert.Equal(t, 99, got[0].TurnNumber)
	assert.Equal(t, 2, got[1].TurnNumber, "appendLog counts entries, ignoring the seeded number")
}
