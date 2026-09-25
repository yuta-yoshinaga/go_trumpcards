//go:build test

package domain_test

import (
	"encoding/json"
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
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			base := tc.newGame()
			resetGame(base)
			data, err := json.Marshal(base)
			require.NoError(t, err)
			var wire map[string]any
			require.NoError(t, json.Unmarshal(data, &wire))
			wire[tc.firstKey], wire[tc.lastKey] = tc.invalid, tc.invalid
			bad, err := json.Marshal(wire)
			require.NoError(t, err)

			var firstError string
			for range 30 {
				game := tc.newGame()
				// Decode into a fresh zero value of the concrete game type.
				switch g := game.(type) {
				case *domain.AllFours:
					err = json.Unmarshal(bad, new(domain.AllFours))
				case *domain.Baloot:
					err = json.Unmarshal(bad, new(domain.Baloot))
				case *domain.Bhabhi:
					err = json.Unmarshal(bad, new(domain.Bhabhi))
				case *domain.BidEuchre:
					err = json.Unmarshal(bad, new(domain.BidEuchre))
				case *domain.Boston:
					err = json.Unmarshal(bad, new(domain.Boston))
				case *domain.Botifarra:
					err = json.Unmarshal(bad, new(domain.Botifarra))
				case *domain.ColourWhist:
					err = json.Unmarshal(bad, new(domain.ColourWhist))
				case *domain.DoubleAttackBlackjack:
					err = json.Unmarshal(bad, new(domain.DoubleAttackBlackjack))
				case *domain.Estimation:
					err = json.Unmarshal(bad, new(domain.Estimation))
				case *domain.Hasenpfeffer:
					err = json.Unmarshal(bad, new(domain.Hasenpfeffer))
				case *domain.Hokm:
					err = json.Unmarshal(bad, new(domain.Hokm))
				case *domain.HoneymoonBridge:
					err = json.Unmarshal(bad, new(domain.HoneymoonBridge))
				default:
					t.Fatalf("unsupported game type %T", g)
				}
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

func resetGame(game any) {
	switch g := game.(type) {
	case *domain.AllFours:
		g.Reset()
	case *domain.Baloot:
		g.Reset()
	case *domain.Bhabhi:
		g.Reset()
	case *domain.BidEuchre:
		g.Reset()
	case *domain.Boston:
		g.Reset()
	case *domain.Botifarra:
		g.Reset()
	case *domain.ColourWhist:
		g.Reset()
	case *domain.DoubleAttackBlackjack:
		g.Reset()
	case *domain.Estimation:
		g.Reset()
	case *domain.Hasenpfeffer:
		g.Reset()
	case *domain.Hokm:
		g.Reset()
	case *domain.HoneymoonBridge:
		g.Reset()
	}
}
