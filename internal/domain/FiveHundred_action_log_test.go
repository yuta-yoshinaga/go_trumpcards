//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFiveHundredBidActionLogSplitsContractLabels(t *testing.T) {
	tests := []struct {
		name       string
		bid        FiveHundredBid
		bidCode    string
		winCode    string
		wantParams map[string]string
	}{
		{"suit", FiveHundredBid{Kind: FiveHundredContractSuit, Tricks: 7, Suit: CardDesignSpade}, "fivehundred.log.bidSuit", "fivehundred.log.winBidSuit", map[string]string{"name": "You", "tricks": "7", "value": "140", "suitKey": "common.suit.spade"}},
		{"no trump", FiveHundredBid{Kind: FiveHundredContractNoTrump, Tricks: 8, Suit: -1}, "fivehundred.log.bidNoTrump", "fivehundred.log.winBidNoTrump", map[string]string{"name": "You", "tricks": "8", "value": "320"}},
		{"misere", FiveHundredBid{Kind: FiveHundredContractMisere, Tricks: 0, Suit: -1}, "fivehundred.log.bidMisere", "fivehundred.log.winBidMisere", map[string]string{"name": "You", "tricks": "0", "value": "250"}},
		{"open misere", FiveHundredBid{Kind: FiveHundredContractOpenMisere, Tricks: 0, Suit: -1}, "fivehundred.log.bidOpenMisere", "fivehundred.log.winBidOpenMisere", map[string]string{"name": "You", "tricks": "0", "value": "520"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewFiveHundred(NewTrumpCardsFiveHundred(), []*FiveHundredPlayer{
				NewFiveHundredPlayer(true, 0), NewFiveHundredPlayer(false, 1),
				NewFiveHundredPlayer(false, 0), NewFiveHundredPlayer(false, 1),
			}, DefaultFiveHundredConfig())
			g.applyBid(0, tt.bid)
			bidEntry := g.actionLog[len(g.actionLog)-1]
			assert.Equal(t, tt.bidCode, bidEntry.DetailCode)
			assert.Equal(t, tt.wantParams, bidEntry.DetailParams)

			g.passed[1], g.passed[2], g.passed[3] = true, true, true
			g.finalizeBid()
			winEntry := g.actionLog[len(g.actionLog)-1]
			assert.Equal(t, tt.winCode, winEntry.DetailCode)
			assert.Equal(t, tt.wantParams, winEntry.DetailParams)
		})
	}
}

func TestFiveHundredBidLogNoneIsPass(t *testing.T) {
	for _, win := range []bool{false, true} {
		code, params := fiveHundredBidLog(FiveHundredBid{Kind: FiveHundredContractNone}, "You", win)
		assert.Equal(t, "fivehundred.log.pass", code)
		assert.Equal(t, map[string]string{"name": "You"}, params)
	}
}
