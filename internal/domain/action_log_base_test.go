//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestActionLogBaseCapsAndMarks(t *testing.T) {
	var b actionLogBase
	for i := 0; i < 250; i++ {
		b.appendLogCode(0, "move", "", nil, nil)
	}
	require.Len(t, b.actionLog, MaxActionLog)
	require.Equal(t, 51, b.actionLog[0].TurnNumber)
	require.Equal(t, 250, b.actionLog[len(b.actionLog)-1].TurnNumber)
	require.Equal(t, 50, b.dropped)
	require.Equal(t, 251, b.nextTurnNumber())
	mark := b.actionLogMark()
	for i := 0; i < 3; i++ {
		b.appendLogCode(0, "move", "", nil, nil)
	}
	b.truncateActionLog(mark)
	require.Len(t, b.actionLog, MaxActionLog-3)
	b.truncateActionLog(0)
	require.Empty(t, b.actionLog)
	b.dropped = 0
	for i := 0; i < 1000; i++ {
		b.actionLog = append(b.actionLog, &ActionLogEntry{TurnNumber: i + 1})
	}
	b.appendLogCode(0, "move", "", nil, nil)
	require.Len(t, b.actionLog, MaxActionLog)
	require.Equal(t, 801, b.dropped)
}

func TestActionLogBaseUsesLegacyNumberingBelowCap(t *testing.T) {
	var b actionLogBase
	require.Equal(t, 1, b.nextTurnNumber())
	b.appendLogCodeAt(9, 0, "custom", "", nil, nil)
	require.Equal(t, 2, b.nextTurnNumber())
	b.appendLogCode(0, "next", "", nil, nil)
	require.Equal(t, 2, b.actionLog[1].TurnNumber)
}

func TestActionLogDroppedJSONRoundTrip(t *testing.T) {
	// Each codec has game-specific state validation, so round-trip initialized states.
	spiderette := NewSpiderette(NewTrumpCards(0))
	spiderette.Reset()
	spiderette.dropped = 7
	wisp := NewWillOTheWisp(NewTrumpCards(0))
	wisp.Reset()
	wisp.dropped = 7
	crib := NewCribbageSquares(NewTrumpCards(0))
	crib.Reset()
	crib.dropped = 7
	poker := NewPokerSquares(NewTrumpCards(0))
	poker.Reset()
	poker.dropped = 7
	monte := NewMonteCarlo(NewTrumpCards(0))
	monte.Reset()
	monte.dropped = 7
	fourteen := NewFourteenOut(NewTrumpCards(0))
	fourteen.Reset()
	fourteen.dropped = 7
	tests := []struct {
		name      string
		marshal   func() ([]byte, error)
		unmarshal func([]byte) (int, error)
	}{
		{"Spiderette", func() ([]byte, error) { return json.Marshal(spiderette) }, func(b []byte) (int, error) { var g Spiderette; err := json.Unmarshal(b, &g); return g.dropped, err }},
		{"WillOTheWisp", func() ([]byte, error) { return json.Marshal(wisp) }, func(b []byte) (int, error) { var g WillOTheWisp; err := json.Unmarshal(b, &g); return g.dropped, err }},
		{"CribbageSquares", func() ([]byte, error) { return json.Marshal(crib) }, func(b []byte) (int, error) {
			var g CribbageSquares
			err := json.Unmarshal(b, &g)
			return g.dropped, err
		}},
		{"PokerSquares", func() ([]byte, error) { return json.Marshal(poker) }, func(b []byte) (int, error) { var g PokerSquares; err := json.Unmarshal(b, &g); return g.dropped, err }},
		{"MonteCarlo", func() ([]byte, error) { return json.Marshal(monte) }, func(b []byte) (int, error) { var g MonteCarlo; err := json.Unmarshal(b, &g); return g.dropped, err }},
		{"FourteenOut", func() ([]byte, error) { return json.Marshal(fourteen) }, func(b []byte) (int, error) { var g FourteenOut; err := json.Unmarshal(b, &g); return g.dropped, err }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.marshal()
			require.NoError(t, err)
			dropped, err := tt.unmarshal(data)
			require.NoError(t, err)
			require.Equal(t, 7, dropped)
		})
	}
	poker.dropped = 0
	zeroData, err := json.Marshal(poker)
	require.NoError(t, err)
	var zeroPayload map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(zeroData, &zeroPayload))
	require.NotContains(t, zeroPayload, "actionLogDropped")
}
