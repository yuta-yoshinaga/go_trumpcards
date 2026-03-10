//go:build test
// +build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// stubGameEndLogger is a test double for gameEndLogger.
type stubGameEndLogger struct {
	ended     bool
	actionLog []*domain.ActionLogEntry
}

func (s *stubGameEndLogger) GetGameEndFlag() bool                   { return s.ended }
func (s *stubGameEndLogger) GetActionLog() []*domain.ActionLogEntry { return s.actionLog }

func TestActionLogOutputText(t *testing.T) {
	t.Run("game not ended returns empty log", func(t *testing.T) {
		g := &stubGameEndLogger{ended: false}
		result := actionLogOutputText(g)
		assert.Equal(t, "棋譜はありません。\n", result)
	})

	t.Run("game ended returns log", func(t *testing.T) {
		g := &stubGameEndLogger{
			ended: true,
			actionLog: []*domain.ActionLogEntry{
				{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "test"},
			},
		}
		result := actionLogOutputText(g)
		assert.Contains(t, result, "T1")
		assert.Contains(t, result, "Player 0")
	})
}

func TestActionLogOutputJSON(t *testing.T) {
	t.Run("game not ended returns empty log", func(t *testing.T) {
		g := &stubGameEndLogger{ended: false}
		result := actionLogOutputJSON(g)
		assert.Contains(t, result, `"entries":[]`)
	})

	t.Run("game ended returns log", func(t *testing.T) {
		g := &stubGameEndLogger{
			ended: true,
			actionLog: []*domain.ActionLogEntry{
				{TurnNumber: 2, PlayerIdx: 1, ActionType: "draw", Detail: "drew a card"},
			},
		}
		result := actionLogOutputJSON(g)
		assert.Contains(t, result, `"turnNumber":2`)
		assert.Contains(t, result, `"playerIdx":1`)
	})
}

func TestActionLogToJSON(t *testing.T) {
	t.Run("empty entries", func(t *testing.T) {
		result := actionLogToJSON([]*domain.ActionLogEntry{})
		assert.Contains(t, result, `"entries":[]`)
	})

	t.Run("nil entries", func(t *testing.T) {
		result := actionLogToJSON(nil)
		assert.Contains(t, result, `"entries":[]`)
	})

	t.Run("entries with cards", func(t *testing.T) {
		entries := []*domain.ActionLogEntry{
			{
				TurnNumber: 1,
				PlayerIdx:  0,
				ActionType: "play",
				Detail:     "played a card",
				Cards:      []*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, true)},
			},
		}
		result := actionLogToJSON(entries)
		assert.Contains(t, result, `"turnNumber":1`)
		assert.Contains(t, result, `"playerIdx":0`)
		assert.Contains(t, result, `"actionType":"play"`)
		assert.Contains(t, result, `"detail":"played a card"`)
		assert.Contains(t, result, `"cards":[`)
		assert.Contains(t, result, `"design":"SPADE"`)
		assert.Contains(t, result, `"value":5`)
	})

	t.Run("system entry playerIdx -1", func(t *testing.T) {
		entries := []*domain.ActionLogEntry{
			{
				TurnNumber: 1,
				PlayerIdx:  -1,
				ActionType: "deal",
				Detail:     "dealt cards",
			},
		}
		result := actionLogToJSON(entries)
		assert.Contains(t, result, `"playerIdx":-1`)
		assert.Contains(t, result, `"actionType":"deal"`)
	})
}

func TestActionLogToText(t *testing.T) {
	t.Run("empty entries", func(t *testing.T) {
		result := actionLogToText([]*domain.ActionLogEntry{})
		assert.Equal(t, "棋譜はありません。\n", result)
	})

	t.Run("nil entries", func(t *testing.T) {
		result := actionLogToText(nil)
		assert.Equal(t, "棋譜はありません。\n", result)
	})

	t.Run("entries with cards", func(t *testing.T) {
		entries := []*domain.ActionLogEntry{
			{
				TurnNumber: 1,
				PlayerIdx:  0,
				ActionType: "play",
				Detail:     "played a card",
				Cards:      []*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, true)},
			},
		}
		result := actionLogToText(entries)
		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "T1")
		assert.Contains(t, result, "Player 0")
		assert.Contains(t, result, "play")
		assert.Contains(t, result, "played a card")
		assert.Contains(t, result, "SPADE 5")
	})

	t.Run("system entry shows SYSTEM", func(t *testing.T) {
		entries := []*domain.ActionLogEntry{
			{
				TurnNumber: 1,
				PlayerIdx:  -1,
				ActionType: "deal",
				Detail:     "dealt cards",
			},
		}
		result := actionLogToText(entries)
		assert.Contains(t, result, "SYSTEM")
		assert.Contains(t, result, "deal")
		assert.Contains(t, result, "dealt cards")
	})

	t.Run("entries without cards no card bracket", func(t *testing.T) {
		entries := []*domain.ActionLogEntry{
			{
				TurnNumber: 2,
				PlayerIdx:  1,
				ActionType: "pass",
				Detail:     "passed",
			},
		}
		result := actionLogToText(entries)
		assert.Contains(t, result, "Player 1")
		assert.Contains(t, result, "pass: passed")
		// No card bracket appended (only [Player X] and header brackets exist)
		assert.NotContains(t, result, "SPADE")
		assert.NotContains(t, result, "HEART")
	})
}
