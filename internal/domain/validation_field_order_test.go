//go:build test

package domain_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestValidationFieldOrder(t *testing.T) {
	tests := []struct {
		name, firstName, firstKey, lastKey string
		invalid                            float64
		newGame                            func() any
	}{
		{"AllFours/current player", "currentPlayerIdx", "ci", "ga", 99, func() any { return domain.NewDefaultAllFours() }},
		{"Baloot/current player", "current player", "cp", "di", 99, func() any { return domain.NewDefaultBaloot() }},
		{"Bhabhi/current player", "current player", "ci", "li", 99, func() any { return domain.NewDefaultBhabhi() }},
		{"Bhabhi/last pickup", "last pickup", "lpi", "bi", 99, func() any { return domain.NewDefaultBhabhi() }},
		{"BidEuchre/dealer", "dealer", "di", "bi", 99, func() any { return domain.NewDefaultBidEuchre() }},
		{"BidEuchre/declarer", "declarer", "de", "tl", 99, func() any { return domain.NewDefaultBidEuchre() }},
		{"Boston/dealer", "dealer", "di", "bi", 99, func() any { return domain.NewDefaultBoston() }},
		{"Boston/declarer", "declarer", "de", "wi", 99, func() any { return domain.NewDefaultBoston() }},
		{"Botifarra/dealer", "dealer", "di", "ct", 99, func() any { return domain.NewDefaultBotifarra() }},
		{"ColourWhist/dealer", "dealer", "di", "cu", 99, func() any { return domain.NewDefaultColourWhist() }},
		{"ColourWhist/declarer", "declarer", "dc", "wi", 99, func() any { return domain.NewDefaultColourWhist() }},
		{"DoubleAttackBlackjack/ante", "ante", "an", "po", -1, func() any { return domain.NewDefaultDoubleAttackBlackjack() }},
		{"Estimation/current player", "current player", "cp", "di", 99, func() any { return domain.NewDefaultEstimation() }},
		{"Hasenpfeffer/current player", "current player", "ci", "dl", 99, func() any { return domain.NewDefaultHasenpfeffer() }},
		{"Hokm/current player", "current player", "cp", "hk", 99, func() any { return domain.NewDefaultHokm() }},
		{"Hokm/winner team", "winner team", "wt", "lw", 99, func() any { return domain.NewDefaultHokm() }},
		{"HoneymoonBridge/current player", "current player", "ci", "dl", 99, func() any { return domain.NewDefaultHoneymoonBridge() }},
		{"IsraeliWhist/current player", "current player", "cp", "di", 99, func() any { return domain.NewDefaultIsraeliWhist() }},
		{"Julepe/current player", "current player", "cp", "di", 99, func() any { return domain.NewDefaultJulepe() }},
		{"Kaiser/dealer", "dealer", "di", "bi", 99, func() any { return domain.NewDefaultKaiser() }},
		{"Kaiser/declarer", "declarer", "de", "s3", 99, func() any { return domain.NewDefaultKaiser() }},
		{"Karnoffel/dealer", "dealer", "di", "tl", 99, func() any { return domain.NewDefaultKarnoffel() }},
		{"Klaberjass/dealer", "dealer", "di", "bi", 99, func() any { return domain.NewDefaultKlaberjass() }},
		{"Klaberjass/maker", "maker", "mi", "wi", 99, func() any { return domain.NewDefaultKlaberjass() }},
		{"LingerLonger/current player", "current player", "ci", "li", 99, func() any { return domain.NewDefaultLingerLonger() }},
		{"Mendikot/current player", "current player", "cp", "di", 99, func() any { return domain.NewDefaultMendikot() }},
		{"Mendikot/winner team", "winner team", "wt", "lw", 99, func() any { return domain.NewDefaultMendikot() }},
		{"Minibridge/current player", "current player", "ci", "dl", 99, func() any { return domain.NewDefaultMinibridge() }},
		{"Polignac/current player", "current player", "cp", "lp", 99, func() any { return domain.NewDefaultPolignac() }},
		{"Rams/current player", "current player", "cp", "di", 99, func() any { return domain.NewDefaultRams() }},
		{"Reversis/current player", "current player", "cp", "di", 99, func() any { return domain.NewDefaultReversis() }},
		{"Rikken/dealer", "dealer", "di", "cu", 99, func() any { return domain.NewDefaultRikken() }},
		{"Rikken/declarer", "declarer", "dc", "wi", 99, func() any { return domain.NewDefaultRikken() }},
		{"RollingStone/seat", "current player", "ci", "li", 99, func() any { return domain.NewDefaultRollingStone() }},
		{"SergeantMajor/seat", "current player", "ci", "dl", 99, func() any { return domain.NewDefaultSergeantMajor() }},
		{"Shelem/seat", "current player", "cp", "di", 99, func() any { return domain.NewDefaultShelem() }},
		{"SixBidSolo/seat", "dealer", "di", "tl", 99, func() any { return domain.NewDefaultSixBidSolo() }},
		{"SixBidSolo/declarer", "declarer", "de", "wi", 99, func() any { return domain.NewDefaultSixBidSolo() }},
		{"Skat/slice lengths", "skat", "sk", "dh", 99, func() any { return domain.NewDefaultSkat() }},
		{"Skat/seat", "currentPlayerIdx", "ci", "r1", 99, func() any { return domain.NewDefaultSkat() }},
		{"Slobberhannes/seat", "current player", "cp", "di", 99, func() any { return domain.NewDefaultSlobberhannes() }},
		{"TappTarock/seat", "current player", "cp", "di", 99, func() any { return domain.NewDefaultTappTarock() }},
		{"TappTarock/declarer", "declarer", "dc", "hr", 99, func() any { return domain.NewDefaultTappTarock() }},
		{"Tarabish/seat", "current player", "cp", "di", 99, func() any { return domain.NewDefaultTarabish() }},
		{"TeenDoPaanch/seat", "current player", "ci", "fi", 99, func() any { return domain.NewDefaultTeenDoPaanch() }},
		{"Troggu/seat", "current player", "cp", "di", 99, func() any { return domain.NewDefaultTroggu() }},
		{"Troggu/declarer", "declarer", "dc", "hr", 99, func() any { return domain.NewDefaultTroggu() }},
		{"Vint/seat", "dealer", "di", "bi", 99, func() any { return domain.NewDefaultVint() }},
		{"Vint/declarer", "declarer", "de", "tl", 99, func() any { return domain.NewDefaultVint() }},
		{"Zwanzigerrufen/seat", "current player", "cp", "di", 99, func() any { return domain.NewDefaultZwanzigerrufen() }},
		{"Zwanzigerrufen/declarer", "declarer", "dc", "hr", 99, func() any { return domain.NewDefaultZwanzigerrufen() }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			base := tc.newGame()
			resetGame(t, base)
			data, err := json.Marshal(base)
			require.NoError(t, err)
			var wire map[string]any
			require.NoError(t, json.Unmarshal(data, &wire))
			wire[tc.firstKey], wire[tc.lastKey] = tc.invalid, tc.invalid
			bad, err := json.Marshal(wire)
			require.NoError(t, err)

			var firstError string
			for range 30 {
				gameType := reflect.TypeOf(tc.newGame())
				fresh := reflect.New(gameType.Elem()).Interface()
				err = json.Unmarshal(bad, fresh)
				require.Error(t, err)
				if firstError == "" {
					firstError = err.Error()
				} else {
					assert.Equal(t, firstError, err.Error())
				}
			}
			assert.True(t, strings.Contains(firstError, tc.firstName), "error %q should contain %q", firstError, tc.firstName)
		})
	}
}

func resetGame(t *testing.T, game any) {
	t.Helper()
	resetter, ok := game.(interface{ Reset() })
	if !ok {
		t.Fatalf("%T does not implement Reset", game)
	}
	resetter.Reset()
}
