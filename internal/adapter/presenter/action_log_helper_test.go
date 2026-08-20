//go:build test
// +build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
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
		assert.Equal(t, i18n.T("cuiActionLogEmpty")+"\n", result)
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
		assert.Contains(t, result, i18n.Tf("cuiActionLogPlayer", "idx", "0"))
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
		assert.Equal(t, i18n.T("cuiActionLogEmpty")+"\n", result)
	})

	t.Run("nil entries", func(t *testing.T) {
		result := actionLogToText(nil)
		assert.Equal(t, i18n.T("cuiActionLogEmpty")+"\n", result)
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
		assert.Contains(t, result, i18n.T("cuiActionLogHeader"))
		assert.Contains(t, result, "T1")
		assert.Contains(t, result, i18n.Tf("cuiActionLogPlayer", "idx", "0"))
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
		assert.Contains(t, result, i18n.T("cuiActionLogSystem"))
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
		assert.Contains(t, result, i18n.Tf("cuiActionLogPlayer", "idx", "1"))
		assert.Contains(t, result, "pass: passed")
		// No card bracket appended (only [Player X] and header brackets exist)
		assert.NotContains(t, result, "SPADE")
		assert.NotContains(t, result, "HEART")
	})
}

// #5977: 棋譜だけ文字列が直書きで、`--lang en` でも日本語の見出しと
// 「棋譜はありません。」が出ていた。座席名も他の行 (cuiPlayerName) が
// 「あなた」「CPU 1」と出すのに対し、ここだけ英語固定の "Player 0" だった。
func TestActionLogTextIsTranslated(t *testing.T) {
	defer i18n.SetLang("ja")

	t.Run("empty log and header follow the language", func(t *testing.T) {
		i18n.SetLang("ja")
		ja := actionLogToText(nil)
		jaHeader := actionLogToText([]*domain.ActionLogEntry{{TurnNumber: 1, PlayerIdx: -1, ActionType: "deal"}})

		i18n.SetLang("en")
		en := actionLogToText(nil)
		enHeader := actionLogToText([]*domain.ActionLogEntry{{TurnNumber: 1, PlayerIdx: -1, ActionType: "deal"}})

		assert.NotEqual(t, ja, en, "空の棋譜が言語で変わらない")
		assert.NotEqual(t, jaHeader, enHeader, "見出しが言語で変わらない")
		// **キーがそのまま出ていないこと。**未定義キーは i18n.T がキー名を返す。
		assert.NotContains(t, en, "cuiActionLog")
		assert.NotContains(t, ja, "cuiActionLog")
	})

	t.Run("seat names match the rest of the screen", func(t *testing.T) {
		i18n.SetLang("ja")
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "出した"},
			{TurnNumber: 2, PlayerIdx: 1, ActionType: "play", Detail: "出した"},
		}
		players := []*domain.AluettePlayer{domain.NewAluettePlayer(true), domain.NewAluettePlayer(false)}

		result := actionLogToTextWithNames(entries, func(idx int) string {
			return cuiPlayerName(players[idx], idx)
		})
		assert.Contains(t, result, i18n.T("cuiPlayerYou"))
		assert.Contains(t, result, "CPU 1")
		assert.NotContains(t, result, i18n.Tf("cuiActionLogPlayer", "idx", "0"))
	})

	t.Run("a seat the resolver does not know still renders", func(t *testing.T) {
		i18n.SetLang("ja")
		entries := []*domain.ActionLogEntry{{TurnNumber: 1, PlayerIdx: 9, ActionType: "play", Detail: "出した"}}
		result := actionLogToTextWithNames(entries, func(int) string { return "" })
		assert.Contains(t, result, "T1", "名前が引けなくても行そのものは出る")
	})
}
