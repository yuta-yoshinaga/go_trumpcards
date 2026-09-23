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
		assert.Contains(t, result, "♠5")
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

func TestActionLogToText_RendersSeparateBidPassAndNilCodes(t *testing.T) {
	entries := []*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "bid", DetailCode: "pitch.log.bidPass", DetailParams: map[string]string{"name": "Player"}},
		{TurnNumber: 2, PlayerIdx: 1, ActionType: "bid", DetailCode: "cinch.log.bidPass", DetailParams: map[string]string{"name": "Player"}},
		{TurnNumber: 3, PlayerIdx: 2, ActionType: "bid", DetailCode: "tarneeb.log.bidPass", DetailParams: map[string]string{"name": "Player"}},
		{TurnNumber: 4, PlayerIdx: 3, ActionType: "bid", DetailCode: "spades.log.bidNil", DetailParams: map[string]string{"name": "Player"}},
	}

	t.Cleanup(func() { i18n.SetLang("ja") })
	i18n.SetLang("ja")
	ja := actionLogToText(entries)
	assert.Contains(t, ja, "Player がパスしました")
	assert.Contains(t, ja, "Player がパス")
	assert.Contains(t, ja, "Playerがパスしました")
	assert.Contains(t, ja, "Player がニルをビッド")
	assert.NotContains(t, ja, "pass")
	assert.NotContains(t, ja, "Pass")
	assert.NotContains(t, ja, "Nil")

	i18n.SetLang("en")
	en := actionLogToText(entries)
	assert.Contains(t, en, "Player passes")
	assert.Contains(t, en, "Player bids Nil")
}

func TestActionLogToText_RendersChangedSuitParamsInBothLanguages(t *testing.T) {
	entries := []*domain.ActionLogEntry{
		{DetailCode: "binokel.log.trump", DetailParams: map[string]string{"suitKey": "common.suit.heart"}},
		{DetailCode: "pinochle.log.trump", DetailParams: map[string]string{"suitKey": "common.suit.heart"}},
		{DetailCode: "mighty.log.declareTrump", DetailParams: map[string]string{"name": "Player", "suitKey": "common.suit.heart"}},
		{DetailCode: "mighty.log.playJokerLead", DetailParams: map[string]string{"name": "Player", "card": "🃏", "suitKey": "common.suit.heart"}},
		{DetailCode: "napoleon.log.declareTrump", DetailParams: map[string]string{"name": "Player", "suitKey": "common.suit.heart"}},
	}

	t.Cleanup(func() { i18n.SetLang("ja") })
	i18n.SetLang("ja")
	ja := actionLogToText(entries)
	assert.Contains(t, ja, "ハート")
	assert.NotContains(t, ja, "Heart")

	i18n.SetLang("en")
	en := actionLogToText(entries)
	assert.Contains(t, en, "Heart")
	assert.NotContains(t, en, "ハート")
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

func TestActionLogToText_RendersMeldAndBetTypesInBothLanguages(t *testing.T) {
	entries := []*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "meld", DetailCode: "bolivia.log.meld", DetailParams: map[string]string{"name": "Player", "typeKey": "bolivia.meldTypeSequence", "cards": "3"}},
		{TurnNumber: 2, PlayerIdx: 0, ActionType: "meld", DetailCode: "bolivia.log.meld", DetailParams: map[string]string{"name": "Player", "typeKey": "bolivia.meldTypeSet", "cards": "3"}},
		{TurnNumber: 3, PlayerIdx: 0, ActionType: "canasta", DetailCode: "bolivia.log.canasta", DetailParams: map[string]string{"name": "Player", "typeKey": "bolivia.meldTypeNatural"}},
		{TurnNumber: 4, PlayerIdx: 0, ActionType: "canasta", DetailCode: "bolivia.log.canasta", DetailParams: map[string]string{"name": "Player", "typeKey": "bolivia.meldTypeMixed"}},
		{TurnNumber: 5, PlayerIdx: 0, ActionType: "canasta", DetailCode: "samba.log.canasta", DetailParams: map[string]string{"name": "Player", "typeKey": "samba.meldTypeNatural"}},
		{TurnNumber: 6, PlayerIdx: 0, ActionType: "canasta", DetailCode: "samba.log.canasta", DetailParams: map[string]string{"name": "Player", "typeKey": "samba.meldTypeMixed"}},
		{TurnNumber: 7, PlayerIdx: 0, ActionType: "meld", DetailCode: "samba.log.meld", DetailParams: map[string]string{"name": "Player", "typeKey": "samba.meldTypeSequence", "cards": "3"}},
		{TurnNumber: 8, PlayerIdx: 0, ActionType: "meld", DetailCode: "samba.log.meld", DetailParams: map[string]string{"name": "Player", "typeKey": "samba.meldTypeSet", "cards": "3"}},
		{TurnNumber: 9, PlayerIdx: 0, ActionType: "canasta", DetailCode: "canasta.log.canasta", DetailParams: map[string]string{"name": "Player", "typeKey": "canasta.meldTypeNatural"}},
		{TurnNumber: 10, PlayerIdx: 0, ActionType: "canasta", DetailCode: "canasta.log.canasta", DetailParams: map[string]string{"name": "Player", "typeKey": "canasta.meldTypeMixed"}},
		{TurnNumber: 11, PlayerIdx: 0, ActionType: "canasta", DetailCode: "handandfoot.log.canasta", DetailParams: map[string]string{"name": "Player", "typeKey": "handandfoot.meldTypeNatural"}},
		{TurnNumber: 12, PlayerIdx: 0, ActionType: "canasta", DetailCode: "handandfoot.log.canasta", DetailParams: map[string]string{"name": "Player", "typeKey": "handandfoot.meldTypeMixed"}},
		{TurnNumber: 13, PlayerIdx: 0, ActionType: "bet", DetailCode: "baccarat.log.bet", DetailParams: map[string]string{"amount": "100", "typeKey": "baccarat.sidePlayer"}},
		{TurnNumber: 14, PlayerIdx: 0, ActionType: "bet", DetailCode: "baccarat.log.bet", DetailParams: map[string]string{"amount": "100", "typeKey": "baccarat.sideBanker"}},
		{TurnNumber: 15, PlayerIdx: 0, ActionType: "bet", DetailCode: "baccarat.log.bet", DetailParams: map[string]string{"amount": "100", "typeKey": "baccarat.sideTie"}},
		{TurnNumber: 16, PlayerIdx: 0, ActionType: "bet", DetailCode: "baccarat.log.bet", DetailParams: map[string]string{"amount": "100", "typeKey": "baccarat.sideUnknown"}},
	}

	t.Cleanup(func() { i18n.SetLang("ja") })

	i18n.SetLang("ja")
	ja := actionLogToText(entries)
	for _, want := range []string{"シーケンス", "セット", "ナチュラル", "ミックス", "プレイヤー", "バンカー", "タイ", "不明"} {
		assert.Contains(t, ja, want)
	}
	for _, unwanted := range []string{"sequence", "set", "natural", "mixed", "player", "banker", "tie", "bolivia.meldTypeSet", "{{"} {
		assert.NotContains(t, ja, unwanted)
	}
	assert.NotContains(t, ja, "100をタイ！に賭けました")

	i18n.SetLang("en")
	en := actionLogToText(entries)
	for _, want := range []string{"sequence", "set", "natural", "mixed", "Player", "Banker", "Tie", "Unknown"} {
		assert.Contains(t, en, want)
	}
	for _, unwanted := range []string{"シーケンス", "セット", "ナチュラル", "ミックス", "プレイヤー", "バンカー", "タイ", "不明", "bolivia.meldTypeSet", "{{"} {
		assert.NotContains(t, en, unwanted)
	}
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

func TestActionLogKeyParamsResolveInBothLanguages(t *testing.T) {
	defer i18n.SetLang("ja")
	entries := []*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "showdown", DetailCode: "poker.log.showdown", DetailParams: map[string]string{"handKey": "pokerhand.highCard"}},
		{TurnNumber: 2, PlayerIdx: 0, ActionType: "result", DetailCode: "videopoker.log.result", DetailParams: map[string]string{"handKey": "pokerhand.jacksOrBetter", "payout": "5"}},
		{TurnNumber: 3, PlayerIdx: 0, ActionType: "exchange", DetailCode: "rook.log.exchange", DetailParams: map[string]string{"name": "Player", "count": "5", "trumpKey": "rook.colorRed"}},
	}

	i18n.SetLang("ja")
	ja := actionLogToText(entries)
	assert.Contains(t, ja, "ハイカード")
	assert.Contains(t, ja, "ジャックス・オア・ベター")
	assert.Contains(t, ja, "赤")
	assert.NotContains(t, ja, "High Card")
	assert.NotContains(t, ja, "Jacks or Better")
	assert.NotContains(t, ja, "Red")

	i18n.SetLang("en")
	en := actionLogToText(entries)
	assert.Contains(t, en, "High Card")
	assert.Contains(t, en, "Jacks or Better")
	assert.Contains(t, en, "Red")
	assert.NotContains(t, en, "ハイカード")
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

func TestActionLogToText_RendersGameIdentifiersInBothLanguages(t *testing.T) {
	entries := []*domain.ActionLogEntry{
		{DetailCode: "boston.log.bid", DetailParams: map[string]string{"bidKey": "boston.bid.chelem"}},
		{DetailCode: "tapptarock.log.bid", DetailParams: map[string]string{"name": "You", "bidKey": "tapptarock.contract.trischaken"}},
		{DetailCode: "zwanzigerrufen.log.bid", DetailParams: map[string]string{"name": "You", "bidKey": "zwanzigerrufen.contract.rufer"}},
		{DetailCode: "troggu.log.contract", DetailParams: map[string]string{"name": "You", "contractKey": "troggu.contractShort.solo"}},
		{DetailCode: "horse.log.hand", DetailParams: map[string]string{"letter": "H", "hand": "1", "nameKey": "horse.discipline.holdem"}},
		{DetailCode: "ironcross.log.line", DetailParams: map[string]string{"seat": "0", "lineKey": "ironcross.line.vertical"}},
		{DetailCode: "kingo.log.deal", DetailParams: map[string]string{"rankKey": "kingo.rank.arashi"}},
		{DetailCode: "unsunkaruta.log.deal", DetailParams: map[string]string{"round": "1", "suitKey": "unsunkaruta.suit.pao"}},
		{DetailCode: "nainjaune.log.award", DetailParams: map[string]string{"boxKey": "nainjaune.box.dwarf", "chips": "1"}},
		{DetailCode: "poch.log.staking", DetailParams: map[string]string{"poolKey": "poch.pool.marriage", "amount": "1"}},
	}

	t.Cleanup(func() { i18n.SetLang("ja") })

	i18n.SetLang("ja")
	ja := actionLogToText(entries)
	for _, text := range []string{"シュレム（13トリック）", "トリシャーケン", "20番呼び", "ソロ", "テキサスホールデム", "縦", "嵐", "ぱお", "♦7（黄色い小人）", "マリッジ"} {
		assert.Contains(t, ja, text)
	}
	for _, text := range []string{"trischaken", "rufer", "holdem", "pao", "arashi", "vertical", "chelem", "marriage", "{{target}}"} {
		assert.NotContains(t, ja, text)
	}

	i18n.SetLang("en")
	en := actionLogToText(entries)
	for _, text := range []string{"Chelem (13 tricks)", "Trischaken", "Call the XX", "Solo", "Texas Hold'em", "vertical", "arashi (three alike)", "Pao", "7 of diamonds (the dwarf)", "Marriage"} {
		assert.Contains(t, en, text)
	}
	for _, text := range []string{"シュレム（13トリック）", "トリシャーケン", "20番呼び", "テキサスホールデム", "ぱお", "嵐", "マリッジ", "{{target}}"} {
		assert.NotContains(t, en, text)
	}
}

func TestActionLogToText_RendersBarbuAndKingContractsInBothLanguages(t *testing.T) {
	entries := []*domain.ActionLogEntry{
		{DetailCode: "barbu.log.selectContract", DetailParams: map[string]string{"dealer": "0", "contractKey": "barbu.cNoTricks"}},
		{DetailCode: "barbu.log.selectContract", DetailParams: map[string]string{"dealer": "0", "contractKey": "barbu.cNoHearts"}},
		{DetailCode: "barbu.log.selectContract", DetailParams: map[string]string{"dealer": "0", "contractKey": "barbu.cNoQueens"}},
		{DetailCode: "barbu.log.selectContract", DetailParams: map[string]string{"dealer": "0", "contractKey": "barbu.cBarbu"}},
		{DetailCode: "barbu.log.selectContract", DetailParams: map[string]string{"dealer": "0", "contractKey": "barbu.cNoLastTrick"}},
		{DetailCode: "barbu.log.selectContract", DetailParams: map[string]string{"dealer": "0", "contractKey": "barbu.cTrumps"}},
		{DetailCode: "barbu.log.selectContract", DetailParams: map[string]string{"dealer": "0", "contractKey": "barbu.cDominoes"}},
		{DetailCode: "king.log.selectContract", DetailParams: map[string]string{"dealer": "0", "contractKey": "king.contractShort.noTricks"}},
		{DetailCode: "king.log.selectContract", DetailParams: map[string]string{"dealer": "0", "contractKey": "king.contractShort.noHearts"}},
		{DetailCode: "king.log.selectContract", DetailParams: map[string]string{"dealer": "0", "contractKey": "king.contractShort.noQueens"}},
		{DetailCode: "king.log.selectContract", DetailParams: map[string]string{"dealer": "0", "contractKey": "king.contractShort.noKingHeart"}},
		{DetailCode: "king.log.selectContract", DetailParams: map[string]string{"dealer": "0", "contractKey": "king.contractShort.noLastTwo"}},
		{DetailCode: "king.log.selectContract", DetailParams: map[string]string{"dealer": "0", "contractKey": "king.contractShort.noMen"}},
		{DetailCode: "king.log.selectContract", DetailParams: map[string]string{"dealer": "0", "contractKey": "king.contractShort.kingTrump"}},
	}

	t.Cleanup(func() { i18n.SetLang("ja") })

	i18n.SetLang("ja")
	ja := actionLogToText(entries)
	for _, text := range []string{"ノー・トリック", "ノー・ハーツ", "ノー・クイーン", "バルブ", "ノー・ラストトリック", "トランプ", "ドミノ", "ノートリック", "ノーハート", "ノークイーン", "ノーキングハート", "ノーラスト2", "ノーメン", "キング（切り札あり）"} {
		assert.Contains(t, ja, text)
	}
	for _, text := range []string{"No Tricks", "No Hearts", "No Queens", "Barbu", "No Last Trick", "Trumps", "Dominoes", "No King of Hearts", "No Last Two Tricks", "No Men", "King (Trump)"} {
		assert.NotContains(t, ja, text)
	}
	for _, key := range []string{"barbu.cNoTricks", "barbu.cNoHearts", "barbu.cNoQueens", "barbu.cBarbu", "barbu.cNoLastTrick", "barbu.cTrumps", "barbu.cDominoes", "king.contractShort.noTricks", "king.contractShort.noHearts", "king.contractShort.noQueens", "king.contractShort.noKingHeart", "king.contractShort.noLastTwo", "king.contractShort.noMen", "king.contractShort.kingTrump"} {
		assert.NotContains(t, ja, key)
	}
	assert.NotContains(t, ja, "（取ったトリックごとに減点）")

	i18n.SetLang("en")
	en := actionLogToText(entries)
	for _, text := range []string{"No Tricks", "No Hearts", "No Queens", "Barbu (K of Hearts)", "No Last Trick", "Trumps", "Dominoes", "No King of Hearts", "No Last Two Tricks", "No Men", "King (Trump)"} {
		assert.Contains(t, en, text)
	}
	for _, text := range []string{"ノー・トリック", "ノー・ハーツ", "ノー・クイーン", "バルブ", "ノー・ラストトリック", "トランプ", "ドミノ", "ノートリック", "ノーハート", "ノークイーン", "ノーキングハート", "ノーラスト2", "ノーメン", "キング（切り札あり）"} {
		assert.NotContains(t, en, text)
	}
	for _, key := range []string{"barbu.cNoTricks", "barbu.cNoHearts", "barbu.cNoQueens", "barbu.cBarbu", "barbu.cNoLastTrick", "barbu.cTrumps", "barbu.cDominoes", "king.contractShort.noTricks", "king.contractShort.noHearts", "king.contractShort.noQueens", "king.contractShort.noKingHeart", "king.contractShort.noLastTwo", "king.contractShort.noMen", "king.contractShort.kingTrump"} {
		assert.NotContains(t, en, key)
	}
}

func TestActionLogToText_RendersSliceSevenContractsAndBidsInBothLanguages(t *testing.T) {
	entries := []*domain.ActionLogEntry{
		{DetailCode: "frenchtarot.log.bid", DetailParams: map[string]string{"player": "Player", "bidKey": "frenchtarot.bidPetite"}},
		{DetailCode: "frenchtarot.log.bid", DetailParams: map[string]string{"player": "Player", "bidKey": "frenchtarot.bidGarde"}},
		{DetailCode: "frenchtarot.log.bid", DetailParams: map[string]string{"player": "Player", "bidKey": "frenchtarot.bidGardeSans"}},
		{DetailCode: "frenchtarot.log.bid", DetailParams: map[string]string{"player": "Player", "bidKey": "frenchtarot.bidGardeContre"}},
		{DetailCode: "frenchtarot.log.bid", DetailParams: map[string]string{"player": "Player", "bidKey": "frenchtarot.bidPass"}},
		{DetailCode: "calabresella.log.bid", DetailParams: map[string]string{"name": "Player", "bidKey": "calabresella.bidChiamo"}},
		{DetailCode: "calabresella.log.bid", DetailParams: map[string]string{"name": "Player", "bidKey": "calabresella.bidSolo"}},
		{DetailCode: "calabresella.log.bid", DetailParams: map[string]string{"name": "Player", "bidKey": "calabresella.bidPass"}},
		{DetailCode: "koenigrufen.log.bid", DetailParams: map[string]string{"name": "Player", "bidKey": "koenigrufen.bidRufer"}},
		{DetailCode: "koenigrufen.log.bid", DetailParams: map[string]string{"name": "Player", "bidKey": "koenigrufen.bidPass"}},
		{DetailCode: "cego.log.bid", DetailParams: map[string]string{"name": "Player", "bidKey": "cego.bidPlay"}},
		{DetailCode: "cego.log.bid", DetailParams: map[string]string{"name": "Player", "bidKey": "cego.bidPass"}},
		{DetailCode: "cego.log.contract", DetailParams: map[string]string{"name": "Player", "contractKey": "cego.contractCego"}},
		{DetailCode: "cego.log.contract", DetailParams: map[string]string{"name": "Player", "contractKey": "cego.contractHandspiel"}},
		{DetailCode: "cego.log.contract", DetailParams: map[string]string{"name": "Player", "contractKey": "cego.contractNone"}},
		{DetailCode: "colourwhist.log.bid", DetailParams: map[string]string{"contractKey": "colourwhist.contractShort.samen"}},
		{DetailCode: "colourwhist.log.bid", DetailParams: map[string]string{"contractKey": "colourwhist.contractShort.alleen"}},
		{DetailCode: "colourwhist.log.bid", DetailParams: map[string]string{"contractKey": "colourwhist.contractShort.miserie"}},
		{DetailCode: "colourwhist.log.bid", DetailParams: map[string]string{"contractKey": "colourwhist.contractShort.troel"}},
		{DetailCode: "colourwhist.log.bid", DetailParams: map[string]string{"contractKey": "colourwhist.contractShort.none"}},
	}

	t.Cleanup(func() { i18n.SetLang("ja") })

	i18n.SetLang("ja")
	ja := actionLogToText(entries)
	for _, text := range []string{"プティット", "ガルド", "ガルド・サン", "ガルド・コントル", "パス", "キアーモ", "ソロ", "ルーファー", "プレイ", "Cego (場札交換)", "Handspiel (手札のまま)", "-", "サーメン", "アレーン", "ミゼリー", "トルール", "未定"} {
		assert.Contains(t, ja, text)
	}
	for _, text := range []string{"petite", "garde", "garde-sans", "garde-contre", "chiamo", "solo", "rufer", "play", "samen", "alleen", "miserie", "troel"} {
		assert.NotContains(t, ja, text)
	}
	for _, key := range []string{"frenchtarot.bidPetite", "frenchtarot.bidGarde", "frenchtarot.bidGardeSans", "frenchtarot.bidGardeContre", "frenchtarot.bidPass", "calabresella.bidChiamo", "calabresella.bidSolo", "calabresella.bidPass", "koenigrufen.bidRufer", "koenigrufen.bidPass", "cego.bidPlay", "cego.bidPass", "cego.contractCego", "cego.contractHandspiel", "cego.contractNone", "colourwhist.contractShort.samen", "colourwhist.contractShort.alleen", "colourwhist.contractShort.miserie", "colourwhist.contractShort.troel", "colourwhist.contractShort.none"} {
		assert.NotContains(t, ja, key)
	}
	assert.NotContains(t, ja, "{{")
	assert.NotContains(t, ja, "（相方と8）")
	assert.NotContains(t, ja, "（エース3枚・強制）")

	i18n.SetLang("en")
	en := actionLogToText(entries)
	for _, text := range []string{"Petite", "Garde", "Garde Sans", "Garde Contre", "Pass", "chiamo", "solo", "Rufer", "Play", "Cego (swap with blind)", "Handspiel (keep hand)", "-", "Samen", "Alleen", "Miserie", "Troel", "undecided"} {
		assert.Contains(t, en, text)
	}
	for _, text := range []string{"プティット", "ガルド", "ガルド・サン", "ガルド・コントル", "キアーモ", "ソロ", "ルーファー", "プレイ", "サーメン", "アレーン", "ミゼリー", "トルール", "未定"} {
		assert.NotContains(t, en, text)
	}
	for _, key := range []string{"frenchtarot.bidPetite", "frenchtarot.bidGarde", "frenchtarot.bidGardeSans", "frenchtarot.bidGardeContre", "frenchtarot.bidPass", "calabresella.bidChiamo", "calabresella.bidSolo", "calabresella.bidPass", "koenigrufen.bidRufer", "koenigrufen.bidPass", "cego.bidPlay", "cego.bidPass", "cego.contractCego", "cego.contractHandspiel", "cego.contractNone", "colourwhist.contractShort.samen", "colourwhist.contractShort.alleen", "colourwhist.contractShort.miserie", "colourwhist.contractShort.troel", "colourwhist.contractShort.none"} {
		assert.NotContains(t, en, key)
	}
	assert.NotContains(t, en, "{{")
	assert.NotContains(t, en, "（相方と8）")
	assert.NotContains(t, en, "（エース3枚・強制）")
}

func TestActionLogToText_RendersHanafudaCaptureLabelsInBothLanguages(t *testing.T) {
	entries := []*domain.ActionLogEntry{
		{DetailCode: "gostop.log.play", DetailParams: map[string]string{"name": "You", "card": "三月·光", "resultKey": "gostop.log.result.captured"}},
		{DetailCode: "gostop.log.draw", DetailParams: map[string]string{"name": "You", "card": "三月·光", "resultKey": "gostop.log.result.toField"}},
		{DetailCode: "hachihachi.log.plays", DetailParams: map[string]string{"name": "You", "card": "三月·光", "resultKey": "hachihachi.log.result.captured"}},
		{DetailCode: "hachihachi.log.draws", DetailParams: map[string]string{"name": "You", "card": "三月·光", "resultKey": "hachihachi.log.result.toField"}},
		{DetailCode: "koikoi.log.play", DetailParams: map[string]string{"name": "You", "card": "三月·光", "capturedKey": "koikoi.log.result.captured"}},
		{DetailCode: "koikoi.log.draw", DetailParams: map[string]string{"name": "You", "card": "三月·光", "capturedKey": "koikoi.log.result.toField"}},
		{DetailCode: "sakura.log.play", DetailParams: map[string]string{"player": "You", "card": "三月·光", "tookKey": "sakura.log.took.captured"}},
		{DetailCode: "sakura.log.draw", DetailParams: map[string]string{"player": "You", "card": "三月·光", "tookKey": "sakura.log.took.discarded"}},
	}

	t.Cleanup(func() { i18n.SetLang("ja") })

	i18n.SetLang("ja")
	ja := actionLogToText(entries)
	for _, text := range []string{"取りました", "場に置きました", "捨てました"} {
		assert.Contains(t, ja, text)
	}
	for _, text := range []string{"captures", "to field", "captured", "discarded", "gostop.log.result.captured", "gostop.log.result.toField", "hachihachi.log.result.captured", "hachihachi.log.result.toField", "koikoi.log.result.captured", "koikoi.log.result.toField", "sakura.log.took.captured", "sakura.log.took.discarded"} {
		assert.NotContains(t, ja, text)
	}

	i18n.SetLang("en")
	en := actionLogToText(entries)
	for _, text := range []string{"captures", "to field", "captured", "discarded"} {
		assert.Contains(t, en, text)
	}
	for _, text := range []string{"取りました", "場に置きました", "捨てました", "gostop.log.result.captured", "gostop.log.result.toField", "hachihachi.log.result.captured", "hachihachi.log.result.toField", "koikoi.log.result.captured", "koikoi.log.result.toField", "sakura.log.took.captured", "sakura.log.took.discarded"} {
		assert.NotContains(t, en, text)
	}
}

func TestActionLogToText_RendersNapPreferenceSoloWhistAndTwentyNineLabelsInBothLanguages(t *testing.T) {
	entries := []*domain.ActionLogEntry{
		{DetailCode: "nap.log.bid", DetailParams: map[string]string{"name": "You", "bidKey": "nap.bid.two"}},
		{DetailCode: "nap.log.contract", DetailParams: map[string]string{"name": "You", "contractKey": "nap.bid.nap", "trump": "0"}},
		{DetailCode: "nap.log.roundScore", DetailParams: map[string]string{"round": "1", "contractKey": "nap.bid.nap", "resultKey": "nap.log.outcome.made", "won": "5", "target": "5"}},
		{DetailCode: "preference.log.bid", DetailParams: map[string]string{"name": "You", "bidKey": "preference.bid.misere"}},
		{DetailCode: "preference.log.contract", DetailParams: map[string]string{"name": "You", "contractKey": "preference.bid.misere", "trump": "0"}},
		{DetailCode: "preference.log.roundScore", DetailParams: map[string]string{"round": "1", "contractKey": "preference.bid.misere", "outcomeKey": "preference.log.outcome.failed", "tricks": "0", "target": "0"}},
		{DetailCode: "solowhist.log.bid", DetailParams: map[string]string{"name": "You", "bidKey": "solowhist.bid.abundance"}},
		{DetailCode: "solowhist.log.contract", DetailParams: map[string]string{"name": "You", "contractKey": "solowhist.bid.abundance", "trump": "0"}},
		{DetailCode: "solowhist.log.roundScore", DetailParams: map[string]string{"round": "1", "contractKey": "solowhist.bid.abundance", "outcomeKey": "solowhist.log.outcome.made", "tricks": "9"}},
		{DetailCode: "twentynine.log.roundScore", DetailParams: map[string]string{"round": "1", "team": "A", "bid": "28", "points": "29", "resultKey": "twentynine.log.result.set"}},
	}

	t.Cleanup(func() { i18n.SetLang("ja") })

	i18n.SetLang("ja")
	ja := actionLogToText(entries)
	for _, text := range []string{"ツー", "ナップ", "ミゼール", "アバンダンス", "成功", "失敗"} {
		assert.Contains(t, ja, text)
	}
	for _, text := range []string{"Two", "Nap", "Abundance", "made", "failed", "set", "nap.bid.two", "twentynine.log.result.set"} {
		assert.NotContains(t, ja, text)
	}

	i18n.SetLang("en")
	en := actionLogToText(entries)
	for _, text := range []string{"Two", "Nap", "Misère", "Abundance", "made", "failed", "set"} {
		assert.Contains(t, en, text)
	}
	for _, text := range []string{"ツー", "ナップ", "成功", "失敗", "nap.bid.two", "twentynine.log.result.set"} {
		assert.NotContains(t, en, text)
	}
}

func TestActionLogToText_RendersCompositeGameLabelsInBothLanguages(t *testing.T) {
	entries := []*domain.ActionLogEntry{
		{DetailCode: "skat.log.declareGameSuit", DetailParams: map[string]string{"name": "You", "trumpKey": "common.suit.spade"}},
		{DetailCode: "skat.log.declareGame", DetailParams: map[string]string{"name": "You", "gameKey": "skat.gameTypeGrand"}},
		{DetailCode: "skat.log.declareGame", DetailParams: map[string]string{"name": "You", "gameKey": "skat.gameTypeNull"}},
		{DetailCode: "bezique.log.meldMarriage", DetailParams: map[string]string{"name": "You", "suitKey": "common.suit.spade", "points": "20"}},
		{DetailCode: "bezique.log.meld", DetailParams: map[string]string{"name": "You", "meldKey": "bezique.meld.royalMarriage", "points": "40"}},
		{DetailCode: "bezique.log.meld", DetailParams: map[string]string{"name": "You", "meldKey": "bezique.meld.bezique", "points": "40"}},
		{DetailCode: "bezique.log.meld", DetailParams: map[string]string{"name": "You", "meldKey": "bezique.meld.fourAces", "points": "100"}},
		{DetailCode: "trappola.log.declarationTrappola", DetailParams: map[string]string{"name": "You", "suitKey": "common.suit.spade", "thirds": "4"}},
		{DetailCode: "trappola.log.declarationFour", DetailParams: map[string]string{"name": "You", "rank": "1", "thirds": "6"}},
		{DetailCode: "trappola.log.declarationThree", DetailParams: map[string]string{"name": "You", "rank": "13", "thirds": "3"}},
	}

	t.Cleanup(func() { i18n.SetLang("ja") })

	i18n.SetLang("ja")
	ja := actionLogToText(entries)
	for _, text := range []string{"スペード", "グランド", "ヌル", "マリッジ", "ロイヤルマリッジ", "ベジック", "エース4枚", "トラッポラ"} {
		assert.Contains(t, ja, text)
	}
	for _, text := range []string{"Spade", "Grand", "Null", "Marriage", "Royal Marriage", "Bezique", "Four Aces", "trappola in", "Suit (trump=", "Spades"} {
		assert.NotContains(t, ja, text)
	}

	i18n.SetLang("en")
	en := actionLogToText(entries)
	for _, text := range []string{"Spade", "Grand", "Null", "Marriage", "Royal Marriage", "Bezique", "Four Aces", "trappola in", "Suit (trump"} {
		assert.Contains(t, en, text)
	}
	for _, text := range []string{"スペード", "グランド", "ヌル", "マリッジ", "ロイヤルマリッジ", "ベジック", "エース4枚", "トラッポラ"} {
		assert.NotContains(t, en, text)
	}
	for _, output := range []string{ja, en} {
		for _, raw := range []string{"skat.log.declareGameSuit", "bezique.meld.bezique", "common.suit.spade", "{{"} {
			assert.NotContains(t, output, raw)
		}
	}
}

func TestActionLogToText_RendersFiveHundredAndGongZhuLabelsInBothLanguages(t *testing.T) {
	entries := []*domain.ActionLogEntry{
		{DetailCode: "fivehundred.log.bidSuit", DetailParams: map[string]string{"name": "P1", "tricks": "7", "suitKey": "common.suit.spade", "value": "140"}},
		{DetailCode: "fivehundred.log.bidNoTrump", DetailParams: map[string]string{"name": "P1", "tricks": "8", "value": "320"}},
		{DetailCode: "fivehundred.log.bidMisere", DetailParams: map[string]string{"name": "P1", "tricks": "0", "value": "250"}},
		{DetailCode: "fivehundred.log.bidOpenMisere", DetailParams: map[string]string{"name": "P1", "tricks": "0", "value": "520"}},
		{DetailCode: "fivehundred.log.winBidSuit", DetailParams: map[string]string{"name": "P1", "tricks": "7", "suitKey": "common.suit.spade", "value": "140"}},
		{DetailCode: "fivehundred.log.winBidNoTrump", DetailParams: map[string]string{"name": "P1", "tricks": "8", "value": "320"}},
		{DetailCode: "fivehundred.log.winBidMisere", DetailParams: map[string]string{"name": "P1", "tricks": "0", "value": "250"}},
		{DetailCode: "fivehundred.log.winBidOpenMisere", DetailParams: map[string]string{"name": "P1", "tricks": "0", "value": "520"}},
		{DetailCode: "gongzhu.log.exposeNone", DetailParams: map[string]string{"round": "3"}},
		{DetailCode: "gongzhu.log.exposeCards", DetailParams: map[string]string{"round": "3", "cards": "♠Q, ♦J"}},
	}

	t.Cleanup(func() { i18n.SetLang("ja") })
	i18n.SetLang("ja")
	ja := actionLogToText(entries)
	for _, text := range []string{"P1が7スペード (140)をビッド", "P1が8NT (320)をビッド", "P1がミゼール (250)をビッド", "P1がオープンミゼール (520)をビッド", "P1が契約を獲得: 7スペード (140)", "ラウンド3: 公開なし", "ラウンド3: 公開 ♠Q, ♦J"} {
		assert.Contains(t, ja, text)
	}
	assert.NotContains(t, ja, "Misere")

	i18n.SetLang("en")
	en := actionLogToText(entries)
	for _, text := range []string{"P1 bids 7Spade (140)", "P1 bids 8NT (320)", "P1 bids Misere (250)", "P1 bids Open Misere (520)", "P1 wins the contract: 7Spade (140)", "round 3: no cards exposed", "round 3: exposed: ♠Q, ♦J"} {
		assert.Contains(t, en, text)
	}
	assert.NotContains(t, en, "スペード")
}

func TestActionLogToText_RendersAndarBaharMusAndViraLabelsInBothLanguages(t *testing.T) {
	entries := []*domain.ActionLogEntry{
		{DetailCode: "andarbahar.log.joker", DetailParams: map[string]string{"columnKey": "andarbahar.columnAndar"}},
		{DetailCode: "andarbahar.log.bet", DetailParams: map[string]string{"columnKey": "andarbahar.columnBahar", "amount": "100"}},
		{DetailCode: "andarbahar.log.match", DetailParams: map[string]string{"columnKey": "andarbahar.columnUnknown", "count": "3"}},
		{DetailCode: "mus.log.showdown", DetailParams: map[string]string{"roundKey": "mus.roundGrande", "team": "A", "stake": "1"}},
		{DetailCode: "mus.log.showdown", DetailParams: map[string]string{"roundKey": "mus.roundChica", "team": "A", "stake": "1"}},
		{DetailCode: "mus.log.showdown", DetailParams: map[string]string{"roundKey": "mus.roundPares", "team": "A", "stake": "1"}},
		{DetailCode: "mus.log.showdown", DetailParams: map[string]string{"roundKey": "mus.roundJuego", "team": "A", "stake": "1"}},
		{DetailCode: "mus.log.showdown", DetailParams: map[string]string{"roundKey": "mus.roundUnknown", "team": "A", "stake": "1"}},
		{DetailCode: "vira.log.bid", DetailParams: map[string]string{"bidKey": "vira.bidShort.pass"}},
		{DetailCode: "vira.log.bid", DetailParams: map[string]string{"bidKey": "vira.bidShort.gask"}},
		{DetailCode: "vira.log.declare", DetailParams: map[string]string{"bidKey": "vira.bidShort.solo", "trumpKey": "common.suit.spade"}},
		{DetailCode: "vira.log.declare", DetailParams: map[string]string{"bidKey": "vira.bidShort.misere", "trumpKey": "common.suit.spade"}},
		{DetailCode: "vira.log.settle", DetailParams: map[string]string{"bidKey": "vira.bidShort.vira", "outcomeKey": "vira.log.outcome.made", "tricks": "10", "pot": "10"}},
	}

	t.Cleanup(func() { i18n.SetLang("ja") })

	i18n.SetLang("ja")
	ja := actionLogToText(entries)
	for _, text := range []string{"アンダー", "バハール", "グランデ", "チカ", "パレス", "フエゴ", "ガスク", "ソロ", "ミゼール", "ヴィーラ"} {
		assert.Contains(t, ja, text)
	}
	for _, text := range []string{"andar", "bahar", "Grande", "Chica", "Pares", "Juego", "Gask", "Solo", "Misère", "Vira", "vira.bidShort.gask", "{{", "(7 トリック)", "(8 トリック)", "(10 トリック)"} {
		assert.NotContains(t, ja, text)
	}

	i18n.SetLang("en")
	en := actionLogToText(entries)
	for _, text := range []string{"Andar", "Bahar", "Grande", "Chica", "Pares", "Juego", "Gask", "Solo", "Misère", "Vira", "Team A"} {
		assert.Contains(t, en, text)
	}
	for _, text := range []string{"アンダー", "バハール", "グランデ", "チカ", "パレス", "フエゴ", "ガスク", "ソロ", "ミゼール", "ヴィーラ", "vira.bidShort.gask", "{{", "(7 トリック)", "(8 トリック)", "(10 トリック)"} {
		assert.NotContains(t, en, text)
	}
}

func TestActionLogToText_RendersTargetContractKeysInBothLanguages(t *testing.T) {
	entries := []*domain.ActionLogEntry{
		{DetailCode: "germansolo.log.bid", DetailParams: map[string]string{"name": "P1", "bidKey": "germansolo.bidSolo", "trumpKey": "common.suit.heart"}},
		{DetailCode: "ombre.log.bid", DetailParams: map[string]string{"name": "P1", "bidKey": "ombre.bidEntrar", "trumpKey": "common.suit.heart"}},
		{DetailCode: "ombre.log.roundScore", DetailParams: map[string]string{"round": "1", "name": "P1", "outcomeKey": "ombre.outcomeSacar", "stake": "1"}},
		{DetailCode: "quadrille.log.bid", DetailParams: map[string]string{"name": "P1", "bidKey": "quadrille.bidSolo", "trumpKey": "common.suit.heart"}},
		{DetailCode: "quadrille.log.roundScore", DetailParams: map[string]string{"round": "1", "name": "P1", "outcomeKey": "quadrille.outcomeCodille", "stake": "1"}},
		{DetailCode: "ulti.log.bid", DetailParams: map[string]string{"name": "P1", "contractKey": "ulti.contractParty", "trumpKey": "common.suit.heart"}},
		{DetailCode: "truco.log.truco", DetailParams: map[string]string{"name": "P1", "levelKey": "truco.levelRetruco"}},
		{DetailCode: "put.log.put", DetailParams: map[string]string{"name": "P1", "levelKey": "put.levelPut"}},
	}

	t.Cleanup(func() { i18n.SetLang("ja") })
	i18n.SetLang("ja")
	ja := actionLogToText(entries)
	assert.Contains(t, ja, "ラウンド 1: オンブル(P1) サカール（オンブル勝ち） (賭け金=1)")
	for _, text := range []string{"ソロ", "エントラール", "サカール（オンブル勝ち）", "コディール（連合勝ち）", "パルティ", "レトルーコ", "プット"} {
		assert.Contains(t, ja, text)
	}
	for _, raw := range []string{"entrar", "solo", "sacar", "codille", "party", "Retruco", "Put", "ombre.bidEntrar"} {
		assert.NotContains(t, ja, raw)
	}

	i18n.SetLang("en")
	en := actionLogToText(entries)
	for _, text := range []string{"Solo", "entrar", "Sacar", "Codille", "Party", "Retruco", "Put"} {
		assert.Contains(t, en, text)
	}
}
