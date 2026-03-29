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

func TestEvalThreeCardHand_AllRanks(t *testing.T) {
	tests := []struct {
		name  string
		cards []*Card
		want  int
	}{
		{
			name: "StraightFlush",
			cards: []*Card{
				NewCard(CardDesignSpade, 5, false),
				NewCard(CardDesignSpade, 6, false),
				NewCard(CardDesignSpade, 7, false),
			},
			want: ThreeCardHandStraightFlush,
		},
		{
			name: "StraightFlush_AceHighQKA",
			cards: []*Card{
				NewCard(CardDesignHeart, 12, false),
				NewCard(CardDesignHeart, 13, false),
				NewCard(CardDesignHeart, 1, false),
			},
			want: ThreeCardHandStraightFlush,
		},
		{
			name: "StraightFlush_AceLowA23",
			cards: []*Card{
				NewCard(CardDesignDiamond, 1, false),
				NewCard(CardDesignDiamond, 2, false),
				NewCard(CardDesignDiamond, 3, false),
			},
			want: ThreeCardHandStraightFlush,
		},
		{
			name: "ThreeOfAKind",
			cards: []*Card{
				NewCard(CardDesignSpade, 8, false),
				NewCard(CardDesignClover, 8, false),
				NewCard(CardDesignHeart, 8, false),
			},
			want: ThreeCardHandThreeOfAKind,
		},
		{
			name: "Straight",
			cards: []*Card{
				NewCard(CardDesignSpade, 4, false),
				NewCard(CardDesignClover, 5, false),
				NewCard(CardDesignHeart, 6, false),
			},
			want: ThreeCardHandStraight,
		},
		{
			name: "Straight_AceHighQKA",
			cards: []*Card{
				NewCard(CardDesignSpade, 12, false),
				NewCard(CardDesignClover, 13, false),
				NewCard(CardDesignHeart, 1, false),
			},
			want: ThreeCardHandStraight,
		},
		{
			name: "Straight_AceLowA23",
			cards: []*Card{
				NewCard(CardDesignSpade, 1, false),
				NewCard(CardDesignClover, 2, false),
				NewCard(CardDesignHeart, 3, false),
			},
			want: ThreeCardHandStraight,
		},
		{
			name: "Flush",
			cards: []*Card{
				NewCard(CardDesignSpade, 2, false),
				NewCard(CardDesignSpade, 7, false),
				NewCard(CardDesignSpade, 11, false),
			},
			want: ThreeCardHandFlush,
		},
		{
			name: "Pair",
			cards: []*Card{
				NewCard(CardDesignSpade, 13, false),
				NewCard(CardDesignClover, 13, false),
				NewCard(CardDesignHeart, 3, false),
			},
			want: ThreeCardHandPair,
		},
		{
			name: "HighCard",
			cards: []*Card{
				NewCard(CardDesignSpade, 2, false),
				NewCard(CardDesignClover, 7, false),
				NewCard(CardDesignHeart, 11, false),
			},
			want: ThreeCardHandHighCard,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, evalThreeCardHand(tt.cards))
		})
	}
}

func TestEvalThreeCardHand_InvalidInput(t *testing.T) {
	// Not 3 cards
	cards := []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignHeart, 2, false),
	}
	assert.Equal(t, ThreeCardHandHighCard, evalThreeCardHand(cards))
}

func TestCheckThreeCardStraight(t *testing.T) {
	assert.True(t, checkThreeCardStraight([]int{4, 5, 6}))
	assert.True(t, checkThreeCardStraight([]int{1, 2, 3}))    // A-2-3
	assert.True(t, checkThreeCardStraight([]int{1, 12, 13}))  // Q-K-A
	assert.True(t, checkThreeCardStraight([]int{11, 12, 13})) // J-Q-K
	assert.False(t, checkThreeCardStraight([]int{1, 5, 9}))   // not a straight
	assert.False(t, checkThreeCardStraight([]int{1, 11, 12})) // J-Q-A is not straight
}

func TestCompareThreeCardHands(t *testing.T) {
	tests := []struct {
		name string
		a    []*Card
		b    []*Card
		want int
	}{
		{
			name: "HigherRankWins",
			a: []*Card{
				NewCard(CardDesignSpade, 5, false),
				NewCard(CardDesignSpade, 6, false),
				NewCard(CardDesignSpade, 7, false),
			}, // straight flush
			b: []*Card{
				NewCard(CardDesignSpade, 8, false),
				NewCard(CardDesignClover, 8, false),
				NewCard(CardDesignHeart, 8, false),
			}, // three of a kind
			want: 1,
		},
		{
			name: "SameRank_HigherKickerWins",
			a: []*Card{
				NewCard(CardDesignSpade, 1, false),
				NewCard(CardDesignClover, 7, false),
				NewCard(CardDesignHeart, 3, false),
			}, // high card Ace
			b: []*Card{
				NewCard(CardDesignSpade, 13, false),
				NewCard(CardDesignClover, 7, false),
				NewCard(CardDesignHeart, 3, false),
			}, // high card King
			want: 1,
		},
		{
			name: "SameRank_SameKickers_Tie",
			a: []*Card{
				NewCard(CardDesignSpade, 10, false),
				NewCard(CardDesignClover, 7, false),
				NewCard(CardDesignHeart, 3, false),
			},
			b: []*Card{
				NewCard(CardDesignDiamond, 10, false),
				NewCard(CardDesignHeart, 7, false),
				NewCard(CardDesignClover, 3, false),
			},
			want: 0,
		},
		{
			name: "LowerRankLoses",
			a: []*Card{
				NewCard(CardDesignSpade, 5, false),
				NewCard(CardDesignClover, 5, false),
				NewCard(CardDesignHeart, 3, false),
			}, // pair
			b: []*Card{
				NewCard(CardDesignSpade, 4, false),
				NewCard(CardDesignSpade, 5, false),
				NewCard(CardDesignSpade, 6, false),
			}, // flush
			want: -1,
		},
		{
			name: "PairVsPair_HigherPairWins",
			a: []*Card{
				NewCard(CardDesignSpade, 1, false),
				NewCard(CardDesignClover, 1, false),
				NewCard(CardDesignHeart, 5, false),
			}, // pair of aces
			b: []*Card{
				NewCard(CardDesignSpade, 13, false),
				NewCard(CardDesignClover, 13, false),
				NewCard(CardDesignHeart, 5, false),
			}, // pair of kings
			want: 1,
		},
		{
			name: "StraightVsStraight_HigherWins",
			a: []*Card{
				NewCard(CardDesignSpade, 12, false),
				NewCard(CardDesignClover, 13, false),
				NewCard(CardDesignHeart, 1, false),
			}, // Q-K-A straight (ace high)
			b: []*Card{
				NewCard(CardDesignSpade, 1, false),
				NewCard(CardDesignClover, 2, false),
				NewCard(CardDesignHeart, 3, false),
			}, // A-2-3 straight (3 high)
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, compareThreeCardHands(tt.a, tt.b))
		})
	}
}
