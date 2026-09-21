//go:build test

package presenter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestPhaseLabels(t *testing.T) {
	t.Cleanup(func() { i18n.SetLang("ja") })

	labels := map[string]map[string]string{
		"andarbahar": {
			"phaseBet": "賭け", "phaseEnd": "終了", "phaseUnknown": "不明",
		},
		"baccarat": {
			"phaseBet": "賭け", "phaseEnd": "終了", "phaseUnknown": "不明",
		},
		"banluck": {
			"phaseBet": "賭け", "phaseGameEnd": "ゲーム終了", "phasePlay": "プレイ",
			"phaseRoundEnd": "ラウンド終了", "phaseUnknown": "不明",
		},
		"baseballpoker": {
			"phaseBetting": "賭け", "phaseBuyIn": "ポット購入", "phaseGameEnd": "ゲーム終了",
			"phaseShowdown": "ショーダウン", "phaseUnknown": "不明",
		},
		"blackjack": {
			"phaseAction": "アクション", "phaseBet": "賭け", "phaseDeal": "配り",
			"phaseEarlySurrender": "アーリーサレンダー", "phaseEnd": "終了",
			"phaseInsurance": "インシュランス", "phaseUnknown": "不明",
		},
		"blackjackswitch": {
			"phaseAction": "アクション", "phaseBet": "賭け", "phaseEnd": "終了",
			"phaseSwitch": "スイッチ", "phaseUnknown": "不明",
		},
		"botifarra": {
			"phaseDeclare": "宣言", "phaseDelegated": "委任済み", "phaseDouble": "ダブル",
			"phaseGameEnd": "ゲーム終了", "phasePlay": "プレイ", "phaseRoundEnd": "ラウンド終了",
			"phaseUnknown": "不明",
		},
		"caribbeanstud": {
			"phaseAction": "アクション", "phaseBet": "賭け", "phaseEnd": "終了", "phaseUnknown": "不明",
		},
		"casinoholdem": {
			"phaseBet": "賭け", "phaseEnd": "終了", "phaseFlop": "フロップ", "phaseUnknown": "不明",
		},
		"casinowar": {
			"phaseBet": "賭け", "phaseEnd": "終了", "phaseInitialDealt": "初期配り済み",
			"phaseTieDecision": "引き分け判断", "phaseUnknown": "不明", "phaseWarDealt": "ウォー配り済み",
		},
		"chemindefer": {
			"phaseBankerDraw": "バンカー引き", "phaseBet": "賭け", "phasePunterDraw": "プレイヤー引き",
			"phaseRoundEnd": "ラウンド終了", "phaseStake": "賭け金", "phaseUnknown": "不明",
		},
		"chinesepoker": {
			"phaseBet": "賭け", "phaseEnd": "終了", "phaseSetHands": "手役セット", "phaseUnknown": "不明",
		},
		"cincinnati": {
			"phaseBetting": "賭け", "phaseDeal": "配り", "phaseGameEnd": "ゲーム終了",
			"phaseShowdown": "ショーダウン", "phaseUnknown": "不明",
		},
		"colourwhist": {
			"phaseBid": "入札", "phaseCall": "コール", "phaseGameEnd": "ゲーム終了", "phasePlay": "プレイ",
			"phaseRoundEnd": "ラウンド終了", "phaseUnknown": "不明",
		},
		"crazyfourpoker": {
			"phaseBet": "賭け", "phaseDecide": "判断", "phaseResult": "結果", "phaseUnknown": "不明",
		},
		"doubleattack": {
			"phaseAttack": "攻撃", "phaseBet": "賭け", "phasePlay": "プレイ", "phaseResult": "結果",
			"phaseUnknown": "不明",
		},
		"dragontiger": {
			"phaseBet": "賭け", "phaseEnd": "終了", "phaseUnknown": "不明",
		},
		"faro": {
			"phaseBetting": "賭け", "phaseCall": "コール", "phaseGameEnd": "ゲーム終了",
			"phaseRoundEnd": "ラウンド終了", "phaseTurn": "ターン", "phaseUnknown": "不明",
		},
		"fourcardpoker": {
			"phaseAction": "アクション", "phaseBet": "賭け", "phaseEnd": "終了", "phaseUnknown": "不明",
		},
		"freebet": {
			"phaseBet": "賭け", "phasePlay": "プレイ", "phaseResult": "結果", "phaseUnknown": "不明",
		},
		"highcardflush": {
			"phaseAction": "アクション", "phaseBet": "賭け", "phaseEnd": "終了", "phaseUnknown": "不明",
		},
		"ironcross": {
			"phaseBetting": "賭け", "phaseChoose": "ライン選択", "phaseGameEnd": "ゲーム終了",
			"phaseShowdown": "ショーダウン", "phaseUnknown": "不明",
		},
		"letitride": {
			"phaseBet": "賭け", "phaseEnd": "終了", "phaseFirstDecision": "1回目の判断",
			"phaseSecondDecision": "2回目の判断", "phaseUnknown": "不明",
		},
		"mississippistud": {
			"phaseAnte": "アンティ", "phaseEnd": "終了", "phaseFifthSt": "5th ストリート",
			"phaseFourthSt": "4th ストリート", "phaseThirdSt": "3rd ストリート", "phaseUnknown": "不明",
		},
		"montebank": {
			"phaseBet": "賭け", "phaseGameEnd": "ゲーム終了", "phaseResult": "結果", "phaseUnknown": "不明",
		},
		"oasispoker": {
			"phaseAction": "アクション", "phaseBet": "賭け", "phaseEnd": "終了", "phaseExchange": "交換",
			"phaseUnknown": "不明",
		},
		"oichokabu": {
			"phaseBet": "賭け", "phaseDraw": "ドロー", "phaseEnd": "終了", "phaseUnknown": "不明",
		},
		"paigow": {
			"phaseBet": "賭け", "phaseEnd": "終了", "phaseSetHands": "手役セット", "phaseUnknown": "不明",
		},
		"reddog": {
			"phaseBet": "賭け", "phaseEnd": "終了", "phaseInitialDealt": "初期配り済み",
			"phaseSpreadDecision": "スプレッド判断", "phaseUnknown": "不明",
		},
		"rikken": {
			"phaseBid": "入札", "phaseCall": "コール", "phaseGameEnd": "ゲーム終了", "phasePlay": "プレイ",
			"phaseRoundEnd": "ラウンド終了", "phaseUnknown": "不明",
		},
		"russianpoker": {
			"phaseAction": "アクション", "phaseBet": "賭け", "phaseEnd": "終了",
			"phaseForceQualify": "強制クオリファイ", "phasePostAction": "アクション後", "phaseSelect": "選択",
			"phaseUnknown": "不明",
		},
		"texasholdembonus": {
			"phaseBet": "賭け", "phaseEnd": "終了", "phaseFlop": "フロップ", "phasePreFlop": "プリフロップ",
			"phaseTurn": "ターン", "phaseUnknown": "不明",
		},
		"threecard": {
			"phaseAction": "アクション", "phaseBet": "賭け", "phaseEnd": "終了", "phaseUnknown": "不明",
		},
		"ultimatetexasholdem": {
			"phaseBet": "賭け", "phaseEnd": "終了", "phaseFlop": "フロップ", "phasePreFlop": "プリフロップ",
			"phaseRiver": "リバー", "phaseUnknown": "不明",
		},
	}

	oldEnglish := []string{
		"BET", "BETTING", "BUY THE POT", "GAME END", "PLAY", "ROUND END", "UNKNOWN", "SHOWDOWN",
		"ACTION", "DEAL", "EARLY SURRENDER", "END", "INSURANCE", "SWITCH", "DECLARE", "DELEGATED",
		"DOUBLE", "FLOP", "INITIAL DEALT", "TIE DECISION", "WAR DEALT", "BANKER DRAW", "PUNTER DRAW",
		"STAKE", "SET HANDS", "BID", "CALL", "DECIDE", "RESULT", "ATTACK", "TURN", "CHOOSE LINE",
		"FIRST DECISION", "SECOND DECISION", "ANTE", "5TH STREET", "4TH STREET", "3RD STREET", "EXCHANGE",
		"DRAW", "SPREAD DECISION", "SELECT", "POST-ACTION", "FORCE QUALIFY", "PRE-FLOP", "RIVER",
	}

	i18n.SetLang("ja")
	for game, gameLabels := range labels {
		for phase, want := range gameLabels {
			key := game + "." + phase
			got := i18n.T(key)
			assert.Equal(t, want, got, key)
			for _, english := range oldEnglish {
				assert.False(t, strings.Contains(got, english), key+" contains "+english)
			}
		}
	}

	i18n.SetLang("en")
	for game, gameLabels := range labels {
		for phase, japanese := range gameLabels {
			key := game + "." + phase
			assert.NotContains(t, i18n.T(key), japanese, key+" contains Japanese translation")
		}
	}
}

func TestTranslatedPresenterLabels(t *testing.T) {
	t.Cleanup(func() { i18n.SetLang("ja") })

	expected := map[string]string{
		"caribbeanstud.qualified":    "(クオリファイ)",
		"caribbeanstud.notQualified": "(クオリファイなし)",
		"highcardflush.qualified":    "(クオリファイ)",
		"highcardflush.notQualified": "(クオリファイなし)",
		"oasispoker.qualified":       "(クオリファイ)",
		"oasispoker.notQualified":    "(クオリファイなし)",
		"russianpoker.qualified":     "(クオリファイ)",
		"russianpoker.notQualified":  "(クオリファイなし)",
		"threecard.qualified":        "(クオリファイ)",
		"threecard.notQualified":     "(クオリファイなし)",
		"agnes.baseRank":             "ベースランク: {{rank}}",
		"canfield.baseRank":          "ベースランク: {{rank}}",
		"penguin.baseRankLabel":      "ベースランク: {{rank}}",
	}

	jaForbidden := []string{"Qualified", "Not Qualified", "Base rank", "BaseRank"}
	jaJapanese := []string{"クオリファイ", "ベースランク"}

	i18n.SetLang("ja")
	for key, want := range expected {
		got := i18n.T(key)
		assert.Equal(t, want, got, key)
		for _, forbidden := range jaForbidden {
			assert.NotContains(t, got, forbidden, key+" contains "+forbidden)
		}
	}

	i18n.SetLang("en")
	for key := range expected {
		got := i18n.T(key)
		for _, japanese := range jaJapanese {
			assert.NotContains(t, got, japanese, key+" contains "+japanese)
		}
	}
}
