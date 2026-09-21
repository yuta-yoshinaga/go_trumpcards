//go:build test

package presenter

import (
	"encoding/json"
	"os"
	"path/filepath"
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

	boardLabels := map[string]string{
		"agnes.foundationHeader":             "組札: ",
		"alaska.foundationHeader":            "組札: ",
		"bakersdozen.foundationHeader":       "組札: ",
		"bakersgame.freeCellHeader":          "フリーセル: ",
		"bakersgame.foundationHeader":        "組札: ",
		"beleagueredcastle.foundationHeader": "組札: ",
		"canfield.foundationHeader":          "組札: ",
		"citadel.foundationHeader":           "組札: ",
		"cruel.foundationHeader":             "組札: ",
		"easthaven.foundationHeader":         "組札: ",
		"eightoff.freeCellHeader":            "フリーセル: ",
		"eightoff.foundationHeader":          "組札: ",
		"flowergarden.foundationHeader":      "組札: ",
		"fortress.foundationHeader":          "組札: ",
		"fortyandeight.foundationHeader":     "組札: ",
		"fortythieves.foundationHeader":      "組札: ",
		"freecell.freeCellHeader":            "フリーセル: ",
		"freecell.foundationHeader":          "組札: ",
		"kingalbert.foundationHeader":        "組札: ",
		"klondike.foundationHeader":          "組札: ",
		"penguin.freeCellHeader":             "フリーセル: ",
		"penguin.foundationHeader":           "組札: ",
		"perseverance.foundationHeader":      "組札: ",
		"rankandfile.foundationHeader":       "組札: ",
		"russiansolitaire.foundationHeader":  "組札: ",
		"seahaventowers.reservedHeader":      "リザーブ: ",
		"seahaventowers.foundationHeader":    "組札: ",
		"somerset.foundationHeader":          "組札: ",
		"stalactites.foundationHeader":       "組札: ",
		"streetsandalleys.foundationHeader":  "組札: ",
		"sultan.foundationHeader":            "組札: ",
		"sultan.divanHeader":                 "ディヴァン:",
		"whitehead.foundationHeader":         "組札: ",
		"yukon.foundationHeader":             "組札: ",
		"blackjack.handStatusBust":           "[バースト]",
		"blackjack.handStatusStand":          "[スタンド]",
		"blackjack.handStatusSurrender":      "[サレンダー]",
	}
	englishForbidden := []string{"Foundation", "FreeCells", "Reserved", "Divan", "BUST", "STAND", "SURRENDER"}
	japaneseForbidden := []string{"組札", "フリーセル", "リザーブ", "ディヴァン", "バースト", "スタンド", "サレンダー"}

	i18n.SetLang("ja")
	for key, want := range boardLabels {
		got := i18n.T(key)
		assert.Equal(t, want, got, key)
		for _, forbidden := range englishForbidden {
			assert.NotContains(t, got, forbidden, key+" contains "+forbidden)
		}
	}

	i18n.SetLang("en")
	for key := range boardLabels {
		got := i18n.T(key)
		for _, forbidden := range japaneseForbidden {
			assert.NotContains(t, got, forbidden, key+" contains "+forbidden)
		}
	}
}

func TestHintTranslationsUseTheSelectedLanguage(t *testing.T) {
	t.Cleanup(func() { i18n.SetLang("ja") })

	localeValues := func(lang string) map[string]string {
		values := map[string]string{}
		dir := filepath.Join("..", "..", "i18n", "locales", lang)
		entries, err := os.ReadDir(dir)
		if !assert.NoError(t, err, lang) {
			return values
		}
		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
				continue
			}
			path := filepath.Join(dir, entry.Name())
			data, readErr := os.ReadFile(path)
			if !assert.NoError(t, readErr, path) {
				continue
			}
			var translations map[string]string
			if !assert.NoError(t, json.Unmarshal(data, &translations), path) {
				continue
			}
			for key, value := range translations {
				values[lang+"."+key] = value
			}
		}
		return values
	}

	i18n.SetLang("ja")
	for key, value := range localeValues("ja") {
		assert.NotContains(t, value, "HINT", key)
	}
	assert.Equal(t, "[ヒント: パス ({{reason}})]", i18n.T("honeymoonbridge.hintPass"))
	assert.Equal(t, "[ヒント: ホールド ({{reason}})]", i18n.T("watten.hintHold"))
	assert.Equal(t, "[ヒント: フォールド ({{reason}})]", i18n.T("watten.hintFold"))

	i18n.SetLang("en")
	for key, value := range localeValues("en") {
		assert.NotContains(t, value, "ヒント", key)
	}
}

func TestJapaneseRemainingLocaleLabels(t *testing.T) {
	t.Cleanup(func() { i18n.SetLang("ja") })

	// blackjack.countingHiLo は Hi-Lo のまま残す (兄弟の countingKO / countingZen /
	// countingOmegaII と同じくカウンティング手法の固有名)。英語を含まない検査には掛けない。
	expected := map[string]string{
		"blackjack.suggestHit": "ヒット", "blackjack.suggestStand": "スタンド",
		"blackjack.suggestSplit": "スプリット", "blackjack.suggestSurrender": "サレンダー",
		"blackjack.suggestDouble": "ダブル", "blackjack.suggestDeclineInsurance": "インシュランスを断る",
		"baccarat.betTypePlayer": "プレイヤー", "baccarat.betTypeBanker": "バンカー",
		"baccarat.betTypeTie": "タイ", "baccarat.betTypeUnknown": "不明",
		"andarbahar.bandUnknown": "不明", "dragontiger.betTypeTie": "タイ",
		"dragontiger.betTypeUnknown": "不明", "chinesepoker.rankUnknown": "不明",
		"paigow.rankUnknown": "不明", "cuiPlayerUnknown": "不明",
		"nertz.foundationEmpty": "(空)", "nertz.nertzEmpty": "  ナッツ: (空)",
		"nertz.tableauEmpty": "(空)", "nertz.wasteEmpty": "  ウェイスト: (空)  ストック: {{stock}}枚",
		"spiteandmalice.foundationEmpty": "(空)", "spiteandmalice.goalEmpty": "ゴール: (空)",
		"spiteandmalice.humanHandEmpty": "(空)", "spiteandmalice.sideEmpty": "(空)",
		"realtime.keySpace":         "スペース",
		"dragontiger.betTypeDragon": "ドラゴン", "dragontiger.betTypeTiger": "タイガー",
		"letitride.betStatusRide": "ライド", "letitride.betStatusPull": "プル",
		"piquet.roleElder": "エルダー", "piquet.roleYounger": "ヤンガー",
		"piquet.declKindPoint": "ポイント", "piquet.declKindSequence": "シークエンス", "piquet.declKindSet": "セット",
		"put.levelPut": "プット", "slapjack.difficultyEasy": "イージー",
		"slapjack.difficultyNormal": "ノーマル", "slapjack.difficultyHard": "ハード",
		"truco.levelTruco": "トルーコ", "truco.levelRetruco": "レトルーコ", "truco.levelValeCuatro": "バレ・クアトロ",
		"klondike.scoringVegas": "ベガス", "whitehead.scoringVegas": "ベガス",
		"clocksolitaire.actionLogTitle": "クロックソリティア行動ログ", "shithead.playerTurnSuffix": " ← 手番",
	}

	oldEnglish := []string{"HIT", "STAND", "SPLIT", "SURRENDER", "PLAYER", "BANKER", "TIE", "UNKNOWN", "(empty)"}
	i18n.SetLang("ja")
	for key, want := range expected {
		got := i18n.T(key)
		assert.Equal(t, want, got, key)
		for _, english := range oldEnglish {
			assert.NotContains(t, got, english, key+" contains "+english)
		}
	}

	i18n.SetLang("en")
	for key, japanese := range expected {
		assert.NotContains(t, i18n.T(key), japanese, key+" contains Japanese translation")
	}
}
