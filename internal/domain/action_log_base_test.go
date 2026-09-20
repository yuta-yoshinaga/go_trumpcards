//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

	g.appendLogCode(0, "a", "a.code", nil, nil)
	assert.Len(t, g.GetActionLog(), 1)

	g.actionLog = nil
	assert.Empty(t, g.GetActionLog(), "assigning nil through the promoted field clears it")

	g.appendLogCode(1, "c", "c.code", nil, nil)
	assert.Equal(t, 1, g.GetActionLog()[0].TurnNumber, "numbering restarts after a clear")
}

func TestActionLogBaseAppendLogCode(t *testing.T) {
	params := map[string]string{"pile": "3"}
	card := NewCard(CardDesignSpade, 7, true)
	b := &actionLogBase{}

	b.appendLogCode(2, "step", "clocksolitaire.log.step", params, []*Card{card})

	assert.Equal(t, []*ActionLogEntry{{
		TurnNumber:   1,
		PlayerIdx:    2,
		ActionType:   "step",
		DetailCode:   "clocksolitaire.log.step",
		DetailParams: params,
		Cards:        []*Card{card},
	}}, b.actionLog)
}

func TestActionLogBaseAppendLogCodeAt(t *testing.T) {
	params := map[string]string{"amount": "25"}
	b := &actionLogBase{}

	b.appendLogCodeAt(9, 1, "raise", "sevencardstud.log.raise", params, nil)

	assert.Equal(t, 1, len(b.actionLog))
	assert.Equal(t, 9, b.actionLog[0].TurnNumber)
	assert.Equal(t, 1, b.actionLog[0].PlayerIdx)
	assert.Equal(t, "raise", b.actionLog[0].ActionType)
	assert.Equal(t, "sevencardstud.log.raise", b.actionLog[0].DetailCode)
	assert.Equal(t, params, b.actionLog[0].DetailParams)
	assert.Nil(t, b.actionLog[0].Cards)
}
