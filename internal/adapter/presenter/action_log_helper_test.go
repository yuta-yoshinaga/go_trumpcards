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
				{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", DetailCode: "test.log.stub", DetailParams: map[string]string{"value": "1"}},
			},
		}
		result := actionLogOutputText(g)
		assert.Contains(t, result, "T1")
		assert.Contains(t, result, i18n.Tf("cuiActionLogPlayer", "idx", "0"))
	})
}

func TestLatestAction(t *testing.T) {
	entries := []*domain.ActionLogEntry{
		{ActionType: "payout"},
		{ActionType: "deal"},
	}
	assert.Nil(t, latestAction(entries, "payout"), "末尾が違う種別なら古い精算を返さない")
	assert.Equal(t, entries[1], latestAction(entries, "deal"))
	assert.Nil(t, latestAction(nil, "deal"))
	assert.Nil(t, latestAction([]*domain.ActionLogEntry{nil}, "deal"))
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
				{TurnNumber: 2, PlayerIdx: 1, ActionType: "draw", DetailCode: "test.log.stub", DetailParams: map[string]string{"value": "1"}},
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
				DetailCode: "test.log.stub", DetailParams: map[string]string{"value": "1"},
				Cards: []*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, true)},
			},
		}
		result := actionLogToJSON(entries)
		assert.Contains(t, result, `"turnNumber":1`)
		assert.Contains(t, result, `"playerIdx":0`)
		assert.Contains(t, result, `"actionType":"play"`)
		assert.Contains(t, result, `"detailCode":"test.log.stub"`)
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
				DetailCode: "test.log.stub", DetailParams: map[string]string{"value": "1"},
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
				DetailCode: "test.log.stub", DetailParams: map[string]string{"value": "1"},
				Cards: []*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, true)},
			},
		}
		result := actionLogToText(entries)
		assert.Contains(t, result, i18n.T("cuiActionLogHeader"))
		assert.Contains(t, result, "T1")
		assert.Contains(t, result, i18n.Tf("cuiActionLogPlayer", "idx", "0"))
		assert.Contains(t, result, "play")
		assert.Contains(t, result, "テスト用の棋譜行 1")
		assert.Contains(t, result, "SPADE 5")
	})

	t.Run("system entry shows SYSTEM", func(t *testing.T) {
		entries := []*domain.ActionLogEntry{
			{
				TurnNumber: 1,
				PlayerIdx:  -1,
				ActionType: "deal",
				DetailCode: "test.log.stub", DetailParams: map[string]string{"value": "1"},
			},
		}
		result := actionLogToText(entries)
		assert.Contains(t, result, i18n.T("cuiActionLogSystem"))
		assert.Contains(t, result, "deal")
		assert.Contains(t, result, "テスト用の棋譜行 1")
	})

	t.Run("entries without cards no card bracket", func(t *testing.T) {
		entries := []*domain.ActionLogEntry{
			{
				TurnNumber: 2,
				PlayerIdx:  1,
				ActionType: "pass",
				DetailCode: "test.log.stub", DetailParams: map[string]string{"value": "1"},
			},
		}
		result := actionLogToText(entries)
		assert.Contains(t, result, i18n.Tf("cuiActionLogPlayer", "idx", "1"))
		assert.Contains(t, result, "pass: テスト用の棋譜行 1")
		// No card bracket appended (only [Player X] and header brackets exist)
		assert.NotContains(t, result, "SPADE")
		assert.NotContains(t, result, "HEART")
	})
}

func TestActionLogToText_ResolvesSjavsTrumpSuitInBothLanguages(t *testing.T) {
	entry := &domain.ActionLogEntry{
		TurnNumber: 1, PlayerIdx: 0, ActionType: "trump", DetailCode: "sjavs.log.trump",
		DetailParams: map[string]string{"suitKey": "common.suit.spade"},
	}

	i18n.SetLang("ja")
	ja := actionLogToText([]*domain.ActionLogEntry{entry})
	assert.Contains(t, ja, "切り札のスート スペード を宣言しました")
	assert.NotContains(t, ja, "Spade")

	i18n.SetLang("en")
	en := actionLogToText([]*domain.ActionLogEntry{entry})
	assert.Contains(t, en, "declares trump suit Spade")

	i18n.SetLang("ja")
}

func TestActionLogToText_RendersTargetGameSuitsInBothLanguages(t *testing.T) {
	entries := []*domain.ActionLogEntry{
		{
			TurnNumber: 1, PlayerIdx: -1, ActionType: "turn_up", DetailCode: "gleek.log.turnUp",
			DetailParams: map[string]string{"card": "♠A", "suitKey": "common.suit.spade"},
		},
		{
			TurnNumber: 2, PlayerIdx: 0, ActionType: "marriage", DetailCode: "marjapussi.log.marriage",
			DetailParams: map[string]string{"name": "Player", "suitKey": "common.suit.spade", "points": "20"},
		},
		{
			TurnNumber: 3, PlayerIdx: -1, ActionType: "trump_set", DetailCode: "loo.log.trumpSet",
			DetailParams: map[string]string{"suitKey": "common.suit.spade", "turnUp": "SPADE 5"},
		},
		{
			TurnNumber: 4, PlayerIdx: 0, ActionType: "call_trump", DetailCode: "omi.log.callTrump",
			DetailParams: map[string]string{"name": "Player", "suitKey": "common.suit.spade"},
		},
		{
			TurnNumber: 5, PlayerIdx: 0, ActionType: "marriage", DetailCode: "tute.log.marriage",
			DetailParams: map[string]string{"name": "Player", "suitKey": "common.suit.spade", "points": "20"},
		},
		{
			TurnNumber: 6, PlayerIdx: 0, ActionType: "marriage", DetailCode: "bauernschnapsen.log.marriage",
			DetailParams: map[string]string{"player": "Player", "suitKey": "common.suit.spade", "bonus": "20", "team": "0"},
		},
		{
			TurnNumber: 7, PlayerIdx: 0, ActionType: "call_trump", DetailCode: "belote.log.callTrump",
			DetailParams: map[string]string{"name": "Player", "suitKey": "common.suit.spade"},
		},
		{
			TurnNumber: 8, PlayerIdx: 0, ActionType: "choose_trump", DetailCode: "jass.log.chooseTrump",
			DetailParams: map[string]string{"name": "Player", "suitKey": "common.suit.spade"},
		},
		{
			TurnNumber: 9, PlayerIdx: 0, ActionType: "declare_trump", DetailCode: "twotenjack.log.declareTrump",
			DetailParams: map[string]string{"name": "Player", "suitKey": "common.suit.spade"},
		},
		{
			TurnNumber: 10, PlayerIdx: 0, ActionType: "trump", DetailCode: "colourwhist.log.trump",
			DetailParams: map[string]string{"suitKey": "common.suit.club"},
		},
		{
			TurnNumber: 11, PlayerIdx: 0, ActionType: "trump", DetailCode: "dehlapakad.log.trump",
			DetailParams: map[string]string{"player": "0", "suitKey": "common.suit.heart"},
		},
	}

	t.Cleanup(func() { i18n.SetLang("ja") })

	i18n.SetLang("ja")
	ja := actionLogToText(entries)
	assert.Contains(t, ja, "スペード")
	assert.Contains(t, ja, "クラブ")
	assert.Contains(t, ja, "ハート")
	assert.NotContains(t, ja, "Spade")
	assert.NotContains(t, ja, "Club")
	assert.NotContains(t, ja, "Heart")
	assert.NotContains(t, ja, "spades")
	assert.NotContains(t, ja, "Spades")

	i18n.SetLang("en")
	en := actionLogToText(entries)
	assert.Contains(t, en, "Spade")
	assert.Contains(t, en, "Club")
	assert.Contains(t, en, "Heart")
	assert.NotContains(t, en, "スペード")
	assert.NotContains(t, en, "クラブ")
	assert.NotContains(t, en, "ハート")
	assert.NotContains(t, en, "spades")
	assert.NotContains(t, en, "Spades")
}

func TestActionLogToText_RendersRemainingGameSuitsInBothLanguages(t *testing.T) {
	entries := []*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "bid", DetailCode: "ombre.log.bid", DetailParams: map[string]string{"name": "You", "bid": "entrar", "trumpKey": "common.suit.spade"}},
		{TurnNumber: 2, PlayerIdx: 0, ActionType: "bid", DetailCode: "ulti.log.bid", DetailParams: map[string]string{"name": "You", "contract": "party", "trumpKey": "common.suit.spade"}},
		{TurnNumber: 3, PlayerIdx: 0, ActionType: "declare", DetailCode: "vira.log.declare", DetailParams: map[string]string{"bid": "normal", "trumpKey": "common.suit.spade"}},
		{TurnNumber: 4, PlayerIdx: 0, ActionType: "declare", DetailCode: "botifarra.log.declare", DetailParams: map[string]string{"suitKey": "common.suit.spade"}},
		{TurnNumber: 5, PlayerIdx: -1, ActionType: "trump", DetailCode: "bourre.log.trump", DetailParams: map[string]string{"suitKey": "common.suit.spade"}},
		{TurnNumber: 6, PlayerIdx: 0, ActionType: "joker", DetailCode: "sevens.log.joker", DetailParams: map[string]string{"suitKey": "common.suit.spade", "value": "7"}},
		{TurnNumber: 7, PlayerIdx: 0, ActionType: "declare", DetailCode: "botifarra.log.declare", DetailParams: map[string]string{"suitKey": "common.suit.notrump"}},
		{TurnNumber: 8, PlayerIdx: 0, ActionType: "bid", DetailCode: "honeymoonbridge.log.bid", DetailParams: map[string]string{"level": "1", "suitKey": "common.suit.spade"}},
		{TurnNumber: 9, PlayerIdx: 0, ActionType: "contract", DetailCode: "honeymoonbridge.log.contract", DetailParams: map[string]string{"level": "1", "suitKey": "common.suit.notrump", "need": "7"}},
	}

	t.Cleanup(func() { i18n.SetLang("ja") })

	i18n.SetLang("ja")
	ja := actionLogToText(entries)
	assert.Contains(t, ja, "スペード")
	assert.NotContains(t, ja, "Spade")
	assert.NotContains(t, ja, "spades")
	assert.NotContains(t, ja, "Spades")
	assert.Contains(t, ja, "切り札なし")
	assert.NotContains(t, ja, "NT")

	i18n.SetLang("en")
	en := actionLogToText(entries)
	assert.Contains(t, en, "Spade")
	assert.NotContains(t, en, "スペード")
	assert.NotContains(t, en, "クラブ")
	assert.Contains(t, en, "No trump")
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
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", DetailCode: "aluette.log.play", DetailParams: map[string]string{"name": "出した"}},
			{TurnNumber: 2, PlayerIdx: 1, ActionType: "play", DetailCode: "aluette.log.play", DetailParams: map[string]string{"name": "出した"}},
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
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", DetailCode: "aluette.log.play", DetailParams: map[string]string{"name": "出した"}},
		}
		result := actionLogToTextWithNames(entries, func(int) string { return "" })
		assert.Contains(t, result, "T1", "名前が引けなくても行そのものは出る")
	})
}

func TestActionLogDetailCodeIsTranslated(t *testing.T) {
	defer i18n.SetLang("ja")
	entry := &domain.ActionLogEntry{
		TurnNumber: 1, PlayerIdx: 0, ActionType: "play",
		DetailCode: "cuiActionLogPlayer", DetailParams: map[string]string{"idx": "7"},
	}

	i18n.SetLang("ja")
	ja := actionLogToText([]*domain.ActionLogEntry{entry})
	assert.Contains(t, ja, "座席7")

	i18n.SetLang("en")
	en := actionLogToText([]*domain.ActionLogEntry{entry})
	assert.Contains(t, en, "Player 7")
	assert.NotEqual(t, ja, en)
}

func TestActionLogDetailCodeIsRenderedAsTranslatedText(t *testing.T) {
	defer i18n.SetLang("ja")
	i18n.SetLang("ja")
	result := actionLogToText([]*domain.ActionLogEntry{{
		TurnNumber: 1, PlayerIdx: 0, ActionType: "play",
		DetailCode: "test.log.stub", DetailParams: map[string]string{"value": "1"},
	}})
	assert.NotContains(t, result, "test.log.stub")
	assert.Contains(t, result, "テスト用の棋譜行 1")
}

func TestActionLogEmptyDetailDoesNotExposeCode(t *testing.T) {
	result := actionLogToText([]*domain.ActionLogEntry{{
		TurnNumber: 1, PlayerIdx: 0, ActionType: "play",
	}})
	assert.NotContains(t, result, "DetailCode")
	assert.NotContains(t, result, "dc")
}

func TestLatestActionOnlyReturnsTheCurrentEvent(t *testing.T) {
	marked := &domain.ActionLogEntry{ActionType: "marked"}
	draw := &domain.ActionLogEntry{ActionType: "draw", Cards: []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, true),
		domain.NewCard(domain.CardDesignSpade, 2, true),
	}}

	assert.Same(t, marked, latestAction([]*domain.ActionLogEntry{marked}, "marked"))
	assert.Nil(t, latestAction([]*domain.ActionLogEntry{marked, draw}, "marked"), "次の操作で一時メッセージを消す")
	assert.Equal(t, "松·短", hachiHachiCapturedLabels(draw.Cards[1:]))
}

func TestActionLogToText_RendersOutcomeLabelsInBothLanguages(t *testing.T) {
	entries := []*domain.ActionLogEntry{
		{DetailCode: "schafkopf.log.roundScore", DetailParams: map[string]string{"round": "1", "points": "61", "outcomeKey": "schafkopf.log.outcome.pickerWins", "multiplier": "1"}},
		{DetailCode: "schafkopf.log.roundScore", DetailParams: map[string]string{"round": "2", "points": "59", "outcomeKey": "schafkopf.log.outcome.defendersWin", "multiplier": "1"}},
		{DetailCode: "sheepshead.log.roundScore", DetailParams: map[string]string{"round": "1", "points": "61", "outcomeKey": "sheepshead.log.outcome.pickerWins", "multiplier": "1"}},
		{DetailCode: "sheepshead.log.roundScore", DetailParams: map[string]string{"round": "2", "points": "59", "outcomeKey": "sheepshead.log.outcome.defendersWin", "multiplier": "1"}},
		{DetailCode: "doppelkopf.log.roundScore", DetailParams: map[string]string{"round": "1", "rePoints": "121", "outcomeKey": "doppelkopf.log.outcome.reWins", "gamePoints": "1"}},
		{DetailCode: "doppelkopf.log.roundScore", DetailParams: map[string]string{"round": "2", "rePoints": "119", "outcomeKey": "doppelkopf.log.outcome.kontraWins", "gamePoints": "1"}},
		{DetailCode: "germansolo.log.roundScore", DetailParams: map[string]string{"round": "1", "name": "You", "outcomeKey": "germansolo.log.outcome.made", "tricks": "7", "total": "10", "needed": "6", "stake": "1"}},
		{DetailCode: "germansolo.log.roundScore", DetailParams: map[string]string{"round": "2", "name": "You", "outcomeKey": "germansolo.log.outcome.failed", "tricks": "5", "total": "10", "needed": "6", "stake": "1"}},
		{DetailCode: "germansolo.log.roundScore", DetailParams: map[string]string{"round": "3", "name": "You", "outcomeKey": "germansolo.log.outcome.unknown", "tricks": "0", "total": "10", "needed": "6", "stake": "1"}},
		{DetailCode: "ulti.log.roundScore", DetailParams: map[string]string{"round": "1", "name": "You", "contract": "party", "outcomeKey": "ulti.log.outcome.win"}},
		{DetailCode: "ulti.log.roundScore", DetailParams: map[string]string{"round": "2", "name": "You", "contract": "party", "outcomeKey": "ulti.log.outcome.loss"}},
		{DetailCode: "ulti.log.roundScore", DetailParams: map[string]string{"round": "3", "name": "You", "contract": "party", "outcomeKey": "ulti.log.outcome.unknown"}},
		{DetailCode: "madrasso.log.roundScore", DetailParams: map[string]string{"round": "1", "teamAPoints": "60", "teamBPoints": "60", "resultKey": "madrasso.log.result.tie", "teamAScore": "1", "teamBScore": "1"}},
		{DetailCode: "madrasso.log.roundScore", DetailParams: map[string]string{"round": "2", "teamAPoints": "61", "teamBPoints": "59", "resultKey": "madrasso.log.result.teamA", "teamAScore": "2", "teamBScore": "1"}},
		{DetailCode: "madrasso.log.roundScore", DetailParams: map[string]string{"round": "3", "teamAPoints": "59", "teamBPoints": "61", "resultKey": "madrasso.log.result.teamB", "teamAScore": "2", "teamBScore": "2"}},
		{DetailCode: "sueca.log.roundScore", DetailParams: map[string]string{"round": "1", "cardsA": "60", "cardsB": "60", "teamKey": "sueca.log.team.draw", "points": "0", "totalA": "1", "totalB": "1"}},
		{DetailCode: "sueca.log.roundScore", DetailParams: map[string]string{"round": "2", "cardsA": "61", "cardsB": "59", "teamKey": "sueca.log.team.a", "points": "1", "totalA": "2", "totalB": "1"}},
		{DetailCode: "sueca.log.roundScore", DetailParams: map[string]string{"round": "3", "cardsA": "59", "cardsB": "61", "teamKey": "sueca.log.team.b", "points": "1", "totalA": "2", "totalB": "2"}},
		{DetailCode: "vira.log.settle", DetailParams: map[string]string{"bid": "normal", "outcomeKey": "vira.log.outcome.made", "tricks": "7", "pot": "10"}},
		{DetailCode: "vira.log.settle", DetailParams: map[string]string{"bid": "normal", "outcomeKey": "vira.log.outcome.failed", "tricks": "5", "pot": "10"}},
	}

	t.Cleanup(func() { i18n.SetLang("ja") })

	i18n.SetLang("ja")
	ja := actionLogToText(entries)
	for _, text := range []string{"ピッカー組の勝ち", "ディフェンス組の勝ち", "Re の勝ち", "Kontra の勝ち", "成功", "失敗", "引き分け", "TeamA の勝ち", "TeamB の勝ち", "チームA", "チームB"} {
		assert.Contains(t, ja, text)
	}
	for _, text := range []string{"picker team wins", "defenders win", "Re wins", "Kontra wins", "made it", "won", "lost", "nobody (tied)", "Draw", "Team A", "Team B", "failed", "made"} {
		assert.NotContains(t, ja, text)
	}

	i18n.SetLang("en")
	en := actionLogToText(entries)
	for _, text := range []string{"picker team wins", "defenders win", "Re wins", "Kontra wins", "made it", "won", "lost", "nobody (tied)", "Draw", "Team A", "Team B", "made", "failed"} {
		assert.Contains(t, en, text)
	}
	for _, text := range []string{"ピッカー組の勝ち", "ディフェンス組の勝ち", "Re の勝ち", "Kontra の勝ち", "成功", "失敗", "引き分け", "チームA", "チームB", "TeamA の勝ち", "TeamB の勝ち"} {
		assert.NotContains(t, en, text)
	}
}
