package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvalFiveCardHand_AllRanks(t *testing.T) {
	tests := []struct {
		name  string
		cards []*Card
		want  int
	}{
		{
			name: "HighCard",
			cards: []*Card{
				NewCard(CardDesignSpade, 2, false),
				NewCard(CardDesignClover, 5, false),
				NewCard(CardDesignHeart, 7, false),
				NewCard(CardDesignDiamond, 9, false),
				NewCard(CardDesignSpade, 11, false),
			},
			want: PokerHandHighCard,
		},
		{
			name: "OnePair",
			cards: []*Card{
				NewCard(CardDesignSpade, 5, false),
				NewCard(CardDesignClover, 5, false),
				NewCard(CardDesignHeart, 7, false),
				NewCard(CardDesignDiamond, 9, false),
				NewCard(CardDesignSpade, 11, false),
			},
			want: PokerHandOnePair,
		},
		{
			name: "TwoPair",
			cards: []*Card{
				NewCard(CardDesignSpade, 5, false),
				NewCard(CardDesignClover, 5, false),
				NewCard(CardDesignHeart, 9, false),
				NewCard(CardDesignDiamond, 9, false),
				NewCard(CardDesignSpade, 11, false),
			},
			want: PokerHandTwoPair,
		},
		{
			name: "ThreeOfAKind",
			cards: []*Card{
				NewCard(CardDesignSpade, 7, false),
				NewCard(CardDesignClover, 7, false),
				NewCard(CardDesignHeart, 7, false),
				NewCard(CardDesignDiamond, 9, false),
				NewCard(CardDesignSpade, 11, false),
			},
			want: PokerHandThreeOfAKind,
		},
		{
			name: "Straight",
			cards: []*Card{
				NewCard(CardDesignSpade, 5, false),
				NewCard(CardDesignClover, 6, false),
				NewCard(CardDesignHeart, 7, false),
				NewCard(CardDesignDiamond, 8, false),
				NewCard(CardDesignSpade, 9, false),
			},
			want: PokerHandStraight,
		},
		{
			name: "Flush",
			cards: []*Card{
				NewCard(CardDesignSpade, 2, false),
				NewCard(CardDesignSpade, 5, false),
				NewCard(CardDesignSpade, 7, false),
				NewCard(CardDesignSpade, 9, false),
				NewCard(CardDesignSpade, 11, false),
			},
			want: PokerHandFlush,
		},
		{
			name: "FullHouse",
			cards: []*Card{
				NewCard(CardDesignSpade, 8, false),
				NewCard(CardDesignClover, 8, false),
				NewCard(CardDesignHeart, 8, false),
				NewCard(CardDesignDiamond, 3, false),
				NewCard(CardDesignSpade, 3, false),
			},
			want: PokerHandFullHouse,
		},
		{
			name: "FourOfAKind",
			cards: []*Card{
				NewCard(CardDesignSpade, 6, false),
				NewCard(CardDesignClover, 6, false),
				NewCard(CardDesignHeart, 6, false),
				NewCard(CardDesignDiamond, 6, false),
				NewCard(CardDesignSpade, 9, false),
			},
			want: PokerHandFourOfAKind,
		},
		{
			name: "StraightFlush",
			cards: []*Card{
				NewCard(CardDesignSpade, 3, false),
				NewCard(CardDesignSpade, 4, false),
				NewCard(CardDesignSpade, 5, false),
				NewCard(CardDesignSpade, 6, false),
				NewCard(CardDesignSpade, 7, false),
			},
			want: PokerHandStraightFlush,
		},
		{
			name: "RoyalFlush",
			cards: []*Card{
				NewCard(CardDesignSpade, 1, false),
				NewCard(CardDesignSpade, 10, false),
				NewCard(CardDesignSpade, 11, false),
				NewCard(CardDesignSpade, 12, false),
				NewCard(CardDesignSpade, 13, false),
			},
			want: PokerHandRoyalFlush,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, evalFiveCardHand(tt.cards))
		})
	}
}

func TestEvalFiveCardHand_LessThan5Cards(t *testing.T) {
	cards := []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignHeart, 2, false),
	}
	assert.Equal(t, PokerHandHighCard, evalFiveCardHand(cards))
}

func TestEvalFiveCardHand_Wheel(t *testing.T) {
	cards := []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignDiamond, 4, false),
		NewCard(CardDesignSpade, 5, false),
	}
	assert.Equal(t, PokerHandStraight, evalFiveCardHand(cards))
}

func TestEvalFiveCardHand_Broadway(t *testing.T) {
	// A-10-J-Q-K mixed suit = Straight
	cards := []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignClover, 10, false),
		NewCard(CardDesignHeart, 11, false),
		NewCard(CardDesignDiamond, 12, false),
		NewCard(CardDesignSpade, 13, false),
	}
	assert.Equal(t, PokerHandStraight, evalFiveCardHand(cards))
}

func TestCheckStraightValues(t *testing.T) {
	// Normal straight
	assert.True(t, checkStraightValues([]int{3, 4, 5, 6, 7}))
	// Ace-high straight (broadway)
	assert.True(t, checkStraightValues([]int{1, 10, 11, 12, 13}))
	// Wheel (A-2-3-4-5)
	assert.True(t, checkStraightValues([]int{1, 2, 3, 4, 5}))
	// Not a straight
	assert.False(t, checkStraightValues([]int{1, 3, 5, 7, 9}))
}

func TestCheckRoyalStraightValues(t *testing.T) {
	assert.True(t, checkRoyalStraightValues([]int{1, 10, 11, 12, 13}))
	assert.False(t, checkRoyalStraightValues([]int{2, 10, 11, 12, 13}))
	assert.False(t, checkRoyalStraightValues([]int{1, 10, 11, 12}))
}
