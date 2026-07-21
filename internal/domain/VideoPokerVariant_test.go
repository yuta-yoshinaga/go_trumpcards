package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- JacksOrBetter GetResult tests ---

func TestJacksOrBetterGetResult(t *testing.T) {
	cfg := JacksOrBetterConfig()
	tests := []struct {
		name      string
		hand      []*Card
		betAmount int
		wantRank  int
		wantMult  int
		wantName  string
	}{
		{
			name: "RoyalFlush_5coins_800x",
			hand: []*Card{
				NewCard(CardDesignSpade, 1, false),
				NewCard(CardDesignSpade, 10, false),
				NewCard(CardDesignSpade, 11, false),
				NewCard(CardDesignSpade, 12, false),
				NewCard(CardDesignSpade, 13, false),
			},
			betAmount: 5, wantRank: PokerHandRoyalFlush, wantMult: 800, wantName: "Royal Flush",
		},
		{
			name: "RoyalFlush_1coin_250x",
			hand: []*Card{
				NewCard(CardDesignSpade, 1, false),
				NewCard(CardDesignSpade, 10, false),
				NewCard(CardDesignSpade, 11, false),
				NewCard(CardDesignSpade, 12, false),
				NewCard(CardDesignSpade, 13, false),
			},
			betAmount: 1, wantRank: PokerHandRoyalFlush, wantMult: 250, wantName: "Royal Flush",
		},
		{
			name: "JacksOrBetter_pair",
			hand: []*Card{
				NewCard(CardDesignSpade, 11, false),
				NewCard(CardDesignClover, 11, false),
				NewCard(CardDesignHeart, 3, false),
				NewCard(CardDesignDiamond, 7, false),
				NewCard(CardDesignSpade, 9, false),
			},
			betAmount: 1, wantRank: PokerHandOnePair, wantMult: 1, wantName: "Jacks or Better",
		},
		{
			name: "LowPair_noPayout",
			hand: []*Card{
				NewCard(CardDesignSpade, 5, false),
				NewCard(CardDesignClover, 5, false),
				NewCard(CardDesignHeart, 3, false),
				NewCard(CardDesignDiamond, 7, false),
				NewCard(CardDesignSpade, 9, false),
			},
			betAmount: 1, wantRank: PokerHandOnePair, wantMult: 0, wantName: "",
		},
		{
			name: "HighCard_noPayout",
			hand: []*Card{
				NewCard(CardDesignSpade, 3, false),
				NewCard(CardDesignClover, 5, false),
				NewCard(CardDesignHeart, 7, false),
				NewCard(CardDesignDiamond, 9, false),
				NewCard(CardDesignSpade, 11, false),
			},
			betAmount: 1, wantRank: PokerHandHighCard, wantMult: 0, wantName: "",
		},
		{
			name: "FullHouse_9x",
			hand: []*Card{
				NewCard(CardDesignSpade, 8, false),
				NewCard(CardDesignClover, 8, false),
				NewCard(CardDesignHeart, 8, false),
				NewCard(CardDesignDiamond, 3, false),
				NewCard(CardDesignSpade, 3, false),
			},
			betAmount: 1, wantRank: PokerHandFullHouse, wantMult: 9, wantName: "Full House",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rank, mult, name := cfg.GetResult(tt.hand, tt.betAmount)
			assert.Equal(t, tt.wantRank, rank)
			assert.Equal(t, tt.wantMult, mult)
			assert.Equal(t, tt.wantName, name)
		})
	}
}

// --- DeucesWild GetResult tests ---

func TestDeucesWildGetResult(t *testing.T) {
	cfg := DeucesWildConfig()
	tests := []struct {
		name      string
		hand      []*Card
		betAmount int
		wantRank  int
		wantMult  int
		wantName  string
	}{
		{
			name: "NaturalRoyalFlush_5coins",
			hand: []*Card{
				NewCard(CardDesignSpade, 1, false),
				NewCard(CardDesignSpade, 10, false),
				NewCard(CardDesignSpade, 11, false),
				NewCard(CardDesignSpade, 12, false),
				NewCard(CardDesignSpade, 13, false),
			},
			betAmount: 5, wantRank: PokerHandRoyalFlush, wantMult: 800, wantName: "Natural Royal Flush",
		},
		{
			name: "NaturalRoyalFlush_1coin",
			hand: []*Card{
				NewCard(CardDesignSpade, 1, false),
				NewCard(CardDesignSpade, 10, false),
				NewCard(CardDesignSpade, 11, false),
				NewCard(CardDesignSpade, 12, false),
				NewCard(CardDesignSpade, 13, false),
			},
			betAmount: 1, wantRank: PokerHandRoyalFlush, wantMult: 250, wantName: "Natural Royal Flush",
		},
		{
			name: "FourDeuces_200x",
			hand: []*Card{
				NewCard(CardDesignSpade, 2, false),
				NewCard(CardDesignClover, 2, false),
				NewCard(CardDesignHeart, 2, false),
				NewCard(CardDesignDiamond, 2, false),
				NewCard(CardDesignSpade, 7, false),
			},
			betAmount: 1, wantRank: PokerHandFourOfAKind, wantMult: 200, wantName: "Four Deuces",
		},
		{
			name: "WildRoyalFlush_25x",
			hand: []*Card{
				NewCard(CardDesignSpade, 2, false), // wild
				NewCard(CardDesignHeart, 1, false),
				NewCard(CardDesignHeart, 10, false),
				NewCard(CardDesignHeart, 12, false),
				NewCard(CardDesignHeart, 13, false),
			},
			betAmount: 1, wantRank: PokerHandRoyalFlush, wantMult: 25, wantName: "Wild Royal Flush",
		},
		{
			name: "FiveOfAKind_15x",
			hand: []*Card{
				NewCard(CardDesignSpade, 2, false),  // wild
				NewCard(CardDesignClover, 2, false), // wild
				NewCard(CardDesignHeart, 2, false),  // wild
				NewCard(CardDesignDiamond, 9, false),
				NewCard(CardDesignSpade, 9, false),
			},
			betAmount: 1, wantRank: PokerHandFiveOfAKind, wantMult: 15, wantName: "Five of a Kind",
		},
		{
			name: "StraightFlush_9x",
			hand: []*Card{
				NewCard(CardDesignSpade, 2, false), // wild
				NewCard(CardDesignSpade, 5, false),
				NewCard(CardDesignSpade, 6, false),
				NewCard(CardDesignSpade, 7, false),
				NewCard(CardDesignSpade, 8, false),
			},
			betAmount: 1, wantRank: PokerHandStraightFlush, wantMult: 9, wantName: "Straight Flush",
		},
		{
			name: "FourOfAKind_5x",
			hand: []*Card{
				NewCard(CardDesignSpade, 2, false), // wild
				NewCard(CardDesignClover, 8, false),
				NewCard(CardDesignHeart, 8, false),
				NewCard(CardDesignDiamond, 8, false),
				NewCard(CardDesignSpade, 5, false),
			},
			betAmount: 1, wantRank: PokerHandFourOfAKind, wantMult: 5, wantName: "Four of a Kind",
		},
		{
			name: "FullHouse_3x",
			hand: []*Card{
				NewCard(CardDesignSpade, 2, false), // wild
				NewCard(CardDesignClover, 8, false),
				NewCard(CardDesignHeart, 8, false),
				NewCard(CardDesignDiamond, 5, false),
				NewCard(CardDesignSpade, 5, false),
			},
			betAmount: 1, wantRank: PokerHandFullHouse, wantMult: 3, wantName: "Full House",
		},
		{
			name: "ThreeOfAKind_1x",
			hand: []*Card{
				NewCard(CardDesignSpade, 2, false), // wild
				NewCard(CardDesignClover, 8, false),
				NewCard(CardDesignHeart, 8, false),
				NewCard(CardDesignDiamond, 5, false),
				NewCard(CardDesignSpade, 11, false),
			},
			betAmount: 1, wantRank: PokerHandThreeOfAKind, wantMult: 1, wantName: "Three of a Kind",
		},
		{
			name: "OnePair_noPayout_deucesWild",
			hand: []*Card{
				NewCard(CardDesignSpade, 3, false),
				NewCard(CardDesignClover, 3, false),
				NewCard(CardDesignHeart, 5, false),
				NewCard(CardDesignDiamond, 7, false),
				NewCard(CardDesignSpade, 11, false),
			},
			betAmount: 1, wantRank: PokerHandOnePair, wantMult: 0, wantName: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rank, mult, name := cfg.GetResult(tt.hand, tt.betAmount)
			assert.Equal(t, tt.wantRank, rank)
			assert.Equal(t, tt.wantMult, mult)
			assert.Equal(t, tt.wantName, name)
		})
	}
}

// --- JokerPoker GetResult tests ---

func TestJokerPokerGetResult(t *testing.T) {
	cfg := JokerPokerConfig()
	tests := []struct {
		name      string
		hand      []*Card
		betAmount int
		wantRank  int
		wantMult  int
		wantName  string
	}{
		{
			name: "NaturalRoyalFlush_5coins",
			hand: []*Card{
				NewCard(CardDesignSpade, 1, false),
				NewCard(CardDesignSpade, 10, false),
				NewCard(CardDesignSpade, 11, false),
				NewCard(CardDesignSpade, 12, false),
				NewCard(CardDesignSpade, 13, false),
			},
			betAmount: 5, wantRank: PokerHandRoyalFlush, wantMult: 800, wantName: "Natural Royal Flush",
		},
		{
			name: "NaturalRoyalFlush_1coin",
			hand: []*Card{
				NewCard(CardDesignSpade, 1, false),
				NewCard(CardDesignSpade, 10, false),
				NewCard(CardDesignSpade, 11, false),
				NewCard(CardDesignSpade, 12, false),
				NewCard(CardDesignSpade, 13, false),
			},
			betAmount: 1, wantRank: PokerHandRoyalFlush, wantMult: 250, wantName: "Natural Royal Flush",
		},
		{
			name: "FiveOfAKind_200x",
			hand: []*Card{
				NewCard(CardDesignJoker, 1, false), // joker
				NewCard(CardDesignSpade, 7, false),
				NewCard(CardDesignClover, 7, false),
				NewCard(CardDesignHeart, 7, false),
				NewCard(CardDesignDiamond, 7, false),
			},
			betAmount: 1, wantRank: PokerHandFiveOfAKind, wantMult: 200, wantName: "Five of a Kind",
		},
		{
			name: "WildRoyalFlush_100x",
			hand: []*Card{
				NewCard(CardDesignJoker, 1, false), // joker
				NewCard(CardDesignHeart, 1, false),
				NewCard(CardDesignHeart, 10, false),
				NewCard(CardDesignHeart, 12, false),
				NewCard(CardDesignHeart, 13, false),
			},
			betAmount: 1, wantRank: PokerHandRoyalFlush, wantMult: 100, wantName: "Wild Royal Flush",
		},
		{
			name: "StraightFlush_50x",
			hand: []*Card{
				NewCard(CardDesignJoker, 1, false), // joker
				NewCard(CardDesignSpade, 5, false),
				NewCard(CardDesignSpade, 6, false),
				NewCard(CardDesignSpade, 7, false),
				NewCard(CardDesignSpade, 8, false),
			},
			betAmount: 1, wantRank: PokerHandStraightFlush, wantMult: 50, wantName: "Straight Flush",
		},
		{
			name: "FourOfAKind_20x",
			hand: []*Card{
				NewCard(CardDesignJoker, 1, false), // joker
				NewCard(CardDesignSpade, 9, false),
				NewCard(CardDesignClover, 9, false),
				NewCard(CardDesignHeart, 9, false),
				NewCard(CardDesignDiamond, 5, false),
			},
			betAmount: 1, wantRank: PokerHandFourOfAKind, wantMult: 20, wantName: "Four of a Kind",
		},
		{
			name: "FullHouse_7x",
			hand: []*Card{
				NewCard(CardDesignSpade, 8, false),
				NewCard(CardDesignClover, 8, false),
				NewCard(CardDesignHeart, 8, false),
				NewCard(CardDesignDiamond, 3, false),
				NewCard(CardDesignSpade, 3, false),
			},
			betAmount: 1, wantRank: PokerHandFullHouse, wantMult: 7, wantName: "Full House",
		},
		{
			name: "TwoPair_1x",
			hand: []*Card{
				NewCard(CardDesignSpade, 5, false),
				NewCard(CardDesignClover, 5, false),
				NewCard(CardDesignHeart, 9, false),
				NewCard(CardDesignDiamond, 9, false),
				NewCard(CardDesignSpade, 11, false),
			},
			betAmount: 1, wantRank: PokerHandTwoPair, wantMult: 1, wantName: "Two Pair",
		},
		{
			name: "KingsOrBetter_1x",
			hand: []*Card{
				NewCard(CardDesignSpade, 13, false),
				NewCard(CardDesignClover, 13, false),
				NewCard(CardDesignHeart, 3, false),
				NewCard(CardDesignDiamond, 7, false),
				NewCard(CardDesignSpade, 9, false),
			},
			betAmount: 1, wantRank: PokerHandOnePair, wantMult: 1, wantName: "Kings or Better",
		},
		{
			name: "AcesOrBetter_1x",
			hand: []*Card{
				NewCard(CardDesignSpade, 1, false),
				NewCard(CardDesignClover, 1, false),
				NewCard(CardDesignHeart, 3, false),
				NewCard(CardDesignDiamond, 7, false),
				NewCard(CardDesignSpade, 9, false),
			},
			betAmount: 1, wantRank: PokerHandOnePair, wantMult: 1, wantName: "Kings or Better",
		},
		{
			name: "QueensPair_noPayout",
			hand: []*Card{
				NewCard(CardDesignSpade, 12, false),
				NewCard(CardDesignClover, 12, false),
				NewCard(CardDesignHeart, 3, false),
				NewCard(CardDesignDiamond, 7, false),
				NewCard(CardDesignSpade, 9, false),
			},
			betAmount: 1, wantRank: PokerHandOnePair, wantMult: 0, wantName: "",
		},
		{
			name: "HighCard_noPayout",
			hand: []*Card{
				NewCard(CardDesignSpade, 3, false),
				NewCard(CardDesignClover, 5, false),
				NewCard(CardDesignHeart, 7, false),
				NewCard(CardDesignDiamond, 9, false),
				NewCard(CardDesignSpade, 11, false),
			},
			betAmount: 1, wantRank: PokerHandHighCard, wantMult: 0, wantName: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rank, mult, name := cfg.GetResult(tt.hand, tt.betAmount)
			assert.Equal(t, tt.wantRank, rank)
			assert.Equal(t, tt.wantMult, mult)
			assert.Equal(t, tt.wantName, name)
		})
	}
}

func TestVariantConfigFactories(t *testing.T) {
	t.Run("JacksOrBetter", func(t *testing.T) {
		cfg := JacksOrBetterConfig()
		assert.Equal(t, "jacksorbetter", cfg.Name)
		assert.Equal(t, 0, cfg.JokerCount)
		assert.Nil(t, cfg.IsWild)
		assert.NotNil(t, cfg.GetResult)
	})
	t.Run("DeucesWild", func(t *testing.T) {
		cfg := DeucesWildConfig()
		assert.Equal(t, "deuceswild", cfg.Name)
		assert.Equal(t, 0, cfg.JokerCount)
		assert.NotNil(t, cfg.IsWild)
		assert.True(t, cfg.IsWild(NewCard(CardDesignSpade, 2, false)))
		assert.False(t, cfg.IsWild(NewCard(CardDesignSpade, 3, false)))
		assert.NotNil(t, cfg.GetResult)
	})
	t.Run("JokerPoker", func(t *testing.T) {
		cfg := JokerPokerConfig()
		assert.Equal(t, "jokerpoker", cfg.Name)
		assert.Equal(t, 1, cfg.JokerCount)
		assert.NotNil(t, cfg.IsWild)
		assert.True(t, cfg.IsWild(NewCard(CardDesignJoker, 1, false)))
		assert.False(t, cfg.IsWild(NewCard(CardDesignSpade, 1, false)))
		assert.NotNil(t, cfg.GetResult)
	})
}

// --- VideoPokerPaytable tests ---

func TestVideoPokerPaytable(t *testing.T) {
	t.Run("jokerpoker rows and order", func(t *testing.T) {
		rows := VideoPokerPaytable("jokerpoker")
		assert.Len(t, rows, 11)
		assert.Equal(t, "ptNaturalRoyalFlush", rows[0].HandKey)
		assert.Equal(t, 250, rows[0].Multiplier)
		assert.True(t, rows[0].RoyalJackpot)
		assert.Equal(t, "ptKingsOrBetter", rows[len(rows)-1].HandKey)
	})

	t.Run("deuceswild rows", func(t *testing.T) {
		rows := VideoPokerPaytable("deuceswild")
		assert.Len(t, rows, 10)
		assert.Equal(t, "ptFourDeuces", rows[1].HandKey)
		assert.Equal(t, 200, rows[1].Multiplier)
	})

	t.Run("unknown variant falls back to jacks or better", func(t *testing.T) {
		rows := VideoPokerPaytable("does-not-exist")
		assert.Len(t, rows, 9)
		assert.Equal(t, "ptRoyalFlush", rows[0].HandKey)
		assert.Equal(t, "ptJacksOrBetter", rows[len(rows)-1].HandKey)
	})

	t.Run("multipliers match GetResult (SSoT consistency)", func(t *testing.T) {
		// Joker + four Kings evaluates to Five of a Kind, which the paytable lists at 200x.
		cfg := JokerPokerConfig()
		hand := []*Card{
			NewCard(CardDesignJoker, 1, false),
			NewCard(CardDesignSpade, 13, false),
			NewCard(CardDesignHeart, 13, false),
			NewCard(CardDesignClover, 13, false),
			NewCard(CardDesignDiamond, 13, false),
		}
		_, multiplier, _ := cfg.GetResult(hand, 1)
		var fiveOfAKind int
		for _, row := range VideoPokerPaytable("jokerpoker") {
			if row.HandKey == "ptFiveOfAKind" {
				fiveOfAKind = row.Multiplier
			}
		}
		assert.Equal(t, fiveOfAKind, multiplier)
	})
}
