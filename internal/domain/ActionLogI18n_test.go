//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestActionLogDetailCodesRenderInBothLanguages(t *testing.T) {
	chemindefer := newChemindeFerAllCpu(t, 1)
	chemindefer.bankerIdx = 0
	chemindefer.bankerHand = chemindeFerHandWorth(7)
	chemindefer.punterHand = chemindeFerHandWorth(2)
	chemindefer.players[1].SetBet(100)
	chemindefer.resolve()
	var resultLog *ActionLogEntry
	for _, entry := range chemindefer.GetActionLog() {
		if entry.ActionType == "result" {
			resultLog = entry
			break
		}
	}
	require.NotNil(t, resultLog)

	blackjack := NewSpanish21BlackJack()
	hand := NewBlackJackHand()
	for _, value := range []int{2, 3, 4, 5, 7} {
		hand.AddCard(NewCard(CardDesignSpade, value, false))
	}
	hand.SetBet(100)
	blackjack.playerHands = []*BlackJackHand{hand}
	blackjack.dealer.AddCard(NewCard(CardDesignClover, 10, false))
	blackjack.dealer.AddCard(NewCard(CardDesignDiamond, 7, false))
	blackjack.resolvePayouts()
	var bonusLog *ActionLogEntry
	for _, entry := range blackjack.GetActionLog() {
		if entry.ActionType == "bonus" {
			bonusLog = entry
			break
		}
	}
	require.NotNil(t, bonusLog)

	want := map[string]map[string]string{
		"ja": {
			"chemindefer.log.resultBanker": "親の勝ち",
			"chemindefer.log.resultPunter": "子側の勝ち",
			"chemindefer.log.resultTie":    "引き分け（égalité）",
			"spanish21.bonus.fivecard21":   "5枚で21 ボーナス",
		},
		"en": {
			"chemindefer.log.resultBanker": "banker wins",
			"chemindefer.log.resultPunter": "punters win",
			"chemindefer.log.resultTie":    "a tie (egalite)",
			"spanish21.bonus.fivecard21":   "5-card 21 bonus",
		},
	}

	for lang, translations := range want {
		i18n.SetLang(lang)
		for _, entry := range []*ActionLogEntry{resultLog, bonusLog} {
			rendered := i18n.Tf(entry.DetailCode)
			assert.Equal(t, translations[entry.DetailCode], rendered)
			assert.NotEqual(t, entry.DetailCode, rendered)
			assert.NotContains(t, rendered, "spanish21.bonus.fivecard21")
		}
	}
	i18n.SetLang("ja")
}
