package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// isDeuce returns true if the card is a 2 (for Deuces Wild testing).
func isDeuce(c *Card) bool {
	return c.GetValue() == 2
}

// isJoker returns true if the card is a joker (for Joker Poker testing).
func isJoker(c *Card) bool {
	return c.GetDesign() == CardDesignJoker
}

func TestEvalWildHand_NoWilds_DelegatesToStandard(t *testing.T) {
	tests := []struct {
		name  string
		cards []*Card
		rank  int
	}{
		{
			name: "HighCard_noWild",
			cards: []*Card{
				NewCard(CardDesignSpade, 3, false),
				NewCard(CardDesignClover, 5, false),
				NewCard(CardDesignHeart, 7, false),
				NewCard(CardDesignDiamond, 9, false),
				NewCard(CardDesignSpade, 11, false),
			},
			rank: PokerHandHighCard,
		},
		{
			name: "RoyalFlush_noWild",
			cards: []*Card{
				NewCard(CardDesignSpade, 1, false),
				NewCard(CardDesignSpade, 10, false),
				NewCard(CardDesignSpade, 11, false),
				NewCard(CardDesignSpade, 12, false),
				NewCard(CardDesignSpade, 13, false),
			},
			rank: PokerHandRoyalFlush,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rank, usedWilds := evalWildHand(tt.cards, isDeuce)
			assert.Equal(t, tt.rank, rank)
			assert.False(t, usedWilds)
		})
	}
}

func TestEvalWildHand_OneJoker(t *testing.T) {
	tests := []struct {
		name      string
		cards     []*Card
		rank      int
		usedWilds bool
	}{
		{
			name: "Joker_completes_RoyalFlush",
			cards: []*Card{
				NewCard(CardDesignJoker, 1, false), // joker
				NewCard(CardDesignSpade, 1, false),
				NewCard(CardDesignSpade, 10, false),
				NewCard(CardDesignSpade, 12, false),
				NewCard(CardDesignSpade, 13, false),
			},
			rank:      PokerHandRoyalFlush,
			usedWilds: true,
		},
		{
			name: "Joker_makes_FourOfAKind_from_ThreeOfAKind",
			cards: []*Card{
				NewCard(CardDesignJoker, 1, false), // joker
				NewCard(CardDesignSpade, 8, false),
				NewCard(CardDesignClover, 8, false),
				NewCard(CardDesignHeart, 8, false),
				NewCard(CardDesignDiamond, 3, false),
			},
			rank:      PokerHandFourOfAKind,
			usedWilds: true,
		},
		{
			name: "Joker_makes_Straight",
			cards: []*Card{
				NewCard(CardDesignJoker, 1, false), // joker
				NewCard(CardDesignSpade, 5, false),
				NewCard(CardDesignClover, 6, false),
				NewCard(CardDesignHeart, 8, false),
				NewCard(CardDesignDiamond, 9, false),
			},
			rank:      PokerHandStraight,
			usedWilds: true,
		},
		{
			name: "Joker_makes_Flush",
			cards: []*Card{
				NewCard(CardDesignJoker, 1, false), // joker
				NewCard(CardDesignHeart, 3, false),
				NewCard(CardDesignHeart, 5, false),
				NewCard(CardDesignHeart, 9, false),
				NewCard(CardDesignHeart, 11, false),
			},
			rank:      PokerHandFlush,
			usedWilds: true,
		},
		{
			name: "Joker_makes_ThreeOfAKind_from_Pair",
			cards: []*Card{
				NewCard(CardDesignJoker, 1, false), // joker
				NewCard(CardDesignSpade, 6, false),
				NewCard(CardDesignClover, 6, false),
				NewCard(CardDesignHeart, 9, false),
				NewCard(CardDesignDiamond, 12, false),
			},
			rank:      PokerHandThreeOfAKind,
			usedWilds: true,
		},
		{
			name: "Joker_makes_FiveOfAKind_from_FourOfAKind",
			cards: []*Card{
				NewCard(CardDesignJoker, 1, false), // joker
				NewCard(CardDesignSpade, 7, false),
				NewCard(CardDesignClover, 7, false),
				NewCard(CardDesignHeart, 7, false),
				NewCard(CardDesignDiamond, 7, false),
			},
			rank:      PokerHandFiveOfAKind,
			usedWilds: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rank, usedWilds := evalWildHand(tt.cards, isJoker)
			assert.Equal(t, tt.rank, rank)
			assert.Equal(t, tt.usedWilds, usedWilds)
		})
	}
}

func TestEvalWildHand_DeucesWild(t *testing.T) {
	tests := []struct {
		name      string
		cards     []*Card
		rank      int
		usedWilds bool
	}{
		{
			name: "One_deuce_makes_ThreeOfAKind",
			cards: []*Card{
				NewCard(CardDesignSpade, 2, false), // wild
				NewCard(CardDesignClover, 8, false),
				NewCard(CardDesignHeart, 8, false),
				NewCard(CardDesignDiamond, 5, false),
				NewCard(CardDesignSpade, 11, false),
			},
			rank:      PokerHandThreeOfAKind,
			usedWilds: true,
		},
		{
			name: "Two_deuces_make_FourOfAKind",
			cards: []*Card{
				NewCard(CardDesignSpade, 2, false),  // wild
				NewCard(CardDesignClover, 2, false), // wild
				NewCard(CardDesignHeart, 9, false),
				NewCard(CardDesignDiamond, 9, false),
				NewCard(CardDesignSpade, 5, false),
			},
			rank:      PokerHandFourOfAKind,
			usedWilds: true,
		},
		{
			name: "Three_deuces_make_FiveOfAKind",
			cards: []*Card{
				NewCard(CardDesignSpade, 2, false),  // wild
				NewCard(CardDesignClover, 2, false), // wild
				NewCard(CardDesignHeart, 2, false),  // wild
				NewCard(CardDesignDiamond, 10, false),
				NewCard(CardDesignSpade, 10, false),
			},
			rank:      PokerHandFiveOfAKind,
			usedWilds: true,
		},
		{
			name: "Four_deuces_FiveOfAKind",
			cards: []*Card{
				NewCard(CardDesignSpade, 2, false),   // wild
				NewCard(CardDesignClover, 2, false),  // wild
				NewCard(CardDesignHeart, 2, false),   // wild
				NewCard(CardDesignDiamond, 2, false), // wild
				NewCard(CardDesignSpade, 7, false),
			},
			rank:      PokerHandFiveOfAKind,
			usedWilds: true,
		},
		{
			name: "Two_deuces_make_StraightFlush",
			cards: []*Card{
				NewCard(CardDesignSpade, 2, false), // wild
				NewCard(CardDesignHeart, 2, false), // wild
				NewCard(CardDesignSpade, 5, false),
				NewCard(CardDesignSpade, 6, false),
				NewCard(CardDesignSpade, 7, false),
			},
			rank:      PokerHandStraightFlush,
			usedWilds: true,
		},
		{
			name: "One_deuce_makes_RoyalFlush",
			cards: []*Card{
				NewCard(CardDesignSpade, 2, false), // wild
				NewCard(CardDesignHeart, 1, false),
				NewCard(CardDesignHeart, 10, false),
				NewCard(CardDesignHeart, 12, false),
				NewCard(CardDesignHeart, 13, false),
			},
			rank:      PokerHandRoyalFlush,
			usedWilds: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rank, usedWilds := evalWildHand(tt.cards, isDeuce)
			assert.Equal(t, tt.rank, rank)
			assert.Equal(t, tt.usedWilds, usedWilds)
		})
	}
}

func TestEvalWildHand_NilIsWild_TreatsAsNoWilds(t *testing.T) {
	cards := []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignSpade, 13, false),
	}
	rank, usedWilds := evalWildHand(cards, nil)
	assert.Equal(t, PokerHandRoyalFlush, rank)
	assert.False(t, usedWilds)
}

func TestCountWilds(t *testing.T) {
	cards := []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignDiamond, 9, false),
		NewCard(CardDesignSpade, 11, false),
	}
	assert.Equal(t, 2, countWilds(cards, isDeuce))
	assert.Equal(t, 0, countWilds(cards, isJoker))
	assert.Equal(t, 0, countWilds(cards, nil))
}
