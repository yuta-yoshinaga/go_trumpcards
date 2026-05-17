package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvalFourCardHand_AllRanks(t *testing.T) {
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
			},
			want: FourCardHandHighCard,
		},
		{
			name: "OnePair",
			cards: []*Card{
				NewCard(CardDesignSpade, 5, false),
				NewCard(CardDesignClover, 5, false),
				NewCard(CardDesignHeart, 7, false),
				NewCard(CardDesignDiamond, 9, false),
			},
			want: FourCardHandPair,
		},
		{
			name: "TwoPair",
			cards: []*Card{
				NewCard(CardDesignSpade, 5, false),
				NewCard(CardDesignClover, 5, false),
				NewCard(CardDesignHeart, 9, false),
				NewCard(CardDesignDiamond, 9, false),
			},
			want: FourCardHandTwoPair,
		},
		{
			name: "Straight",
			cards: []*Card{
				NewCard(CardDesignSpade, 5, false),
				NewCard(CardDesignClover, 6, false),
				NewCard(CardDesignHeart, 7, false),
				NewCard(CardDesignDiamond, 8, false),
			},
			want: FourCardHandStraight,
		},
		{
			name: "Flush",
			cards: []*Card{
				NewCard(CardDesignSpade, 2, false),
				NewCard(CardDesignSpade, 5, false),
				NewCard(CardDesignSpade, 7, false),
				NewCard(CardDesignSpade, 11, false),
			},
			want: FourCardHandFlush,
		},
		{
			name: "ThreeOfAKind",
			cards: []*Card{
				NewCard(CardDesignSpade, 7, false),
				NewCard(CardDesignClover, 7, false),
				NewCard(CardDesignHeart, 7, false),
				NewCard(CardDesignDiamond, 9, false),
			},
			want: FourCardHandThreeOfAKind,
		},
		{
			name: "StraightFlush",
			cards: []*Card{
				NewCard(CardDesignSpade, 5, false),
				NewCard(CardDesignSpade, 6, false),
				NewCard(CardDesignSpade, 7, false),
				NewCard(CardDesignSpade, 8, false),
			},
			want: FourCardHandStraightFlush,
		},
		{
			name: "FourOfAKind",
			cards: []*Card{
				NewCard(CardDesignSpade, 9, false),
				NewCard(CardDesignClover, 9, false),
				NewCard(CardDesignHeart, 9, false),
				NewCard(CardDesignDiamond, 9, false),
			},
			want: FourCardHandFourOfAKind,
		},
		{
			name: "StraightWheelA234",
			cards: []*Card{
				NewCard(CardDesignSpade, 1, false),
				NewCard(CardDesignClover, 2, false),
				NewCard(CardDesignHeart, 3, false),
				NewCard(CardDesignDiamond, 4, false),
			},
			want: FourCardHandStraight,
		},
		{
			name: "StraightBroadwayJQKA",
			cards: []*Card{
				NewCard(CardDesignSpade, 11, false),
				NewCard(CardDesignClover, 12, false),
				NewCard(CardDesignHeart, 13, false),
				NewCard(CardDesignDiamond, 1, false),
			},
			want: FourCardHandStraight,
		},
		{
			name: "StraightFlushBroadway",
			cards: []*Card{
				NewCard(CardDesignSpade, 11, false),
				NewCard(CardDesignSpade, 12, false),
				NewCard(CardDesignSpade, 13, false),
				NewCard(CardDesignSpade, 1, false),
			},
			want: FourCardHandStraightFlush,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := evalFourCardHand(tt.cards)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEvalFourCardHand_InvalidLength(t *testing.T) {
	cards := []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignClover, 5, false),
	}
	assert.Equal(t, FourCardHandHighCard, evalFourCardHand(cards))
}

func TestCompareFourCardHands_DifferentRanks(t *testing.T) {
	pair := []*Card{
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignClover, 5, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignDiamond, 9, false),
	}
	flush := []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignSpade, 7, false),
		NewCard(CardDesignSpade, 11, false),
	}
	assert.Equal(t, 1, compareFourCardHands(flush, pair))
	assert.Equal(t, -1, compareFourCardHands(pair, flush))
}

func TestCompareFourCardHands_HighCardTie(t *testing.T) {
	a := []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignClover, 5, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignDiamond, 9, false),
	}
	b := []*Card{
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignClover, 5, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignDiamond, 9, false),
	}
	assert.Equal(t, 1, compareFourCardHands(b, a))
	assert.Equal(t, -1, compareFourCardHands(a, b))
}

func TestCompareFourCardHands_HighCardEqual(t *testing.T) {
	a := []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignClover, 5, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignDiamond, 9, false),
	}
	b := []*Card{
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignDiamond, 5, false),
		NewCard(CardDesignSpade, 7, false),
		NewCard(CardDesignClover, 9, false),
	}
	assert.Equal(t, 0, compareFourCardHands(a, b))
}

func TestCompareFourCardHands_PairTieKicker(t *testing.T) {
	// Both have pair of 5; kickers differ
	a := []*Card{
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignClover, 5, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignDiamond, 9, false),
	}
	b := []*Card{
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignDiamond, 5, false),
		NewCard(CardDesignSpade, 8, false),
		NewCard(CardDesignClover, 9, false),
	}
	assert.Equal(t, -1, compareFourCardHands(a, b))
	assert.Equal(t, 1, compareFourCardHands(b, a))
}

func TestCompareFourCardHands_TwoPairTie(t *testing.T) {
	// Both 9s and 5s; identical kickers (none for 4-card)
	a := []*Card{
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignClover, 5, false),
		NewCard(CardDesignHeart, 9, false),
		NewCard(CardDesignDiamond, 9, false),
	}
	b := []*Card{
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignDiamond, 5, false),
		NewCard(CardDesignSpade, 9, false),
		NewCard(CardDesignClover, 9, false),
	}
	assert.Equal(t, 0, compareFourCardHands(a, b))
}

func TestCompareFourCardHands_TwoPairCompareHigh(t *testing.T) {
	// 9s+5s vs 8s+7s — higher top pair wins
	a := []*Card{
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignClover, 5, false),
		NewCard(CardDesignHeart, 9, false),
		NewCard(CardDesignDiamond, 9, false),
	}
	b := []*Card{
		NewCard(CardDesignSpade, 7, false),
		NewCard(CardDesignClover, 7, false),
		NewCard(CardDesignHeart, 8, false),
		NewCard(CardDesignDiamond, 8, false),
	}
	assert.Equal(t, 1, compareFourCardHands(a, b))
}

func TestCompareFourCardHands_StraightWheelLowerThanRegular(t *testing.T) {
	wheel := []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignDiamond, 4, false),
	}
	regular := []*Card{
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignClover, 6, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignDiamond, 8, false),
	}
	assert.Equal(t, -1, compareFourCardHands(wheel, regular))
}

func TestCompareFourCardHands_StraightBroadwayBeatsRegular(t *testing.T) {
	broadway := []*Card{
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignClover, 12, false),
		NewCard(CardDesignHeart, 13, false),
		NewCard(CardDesignDiamond, 1, false),
	}
	regular := []*Card{
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignClover, 6, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignDiamond, 8, false),
	}
	assert.Equal(t, 1, compareFourCardHands(broadway, regular))
}

func TestCompareFourCardHands_ThreeOfAKindCompareTrips(t *testing.T) {
	a := []*Card{
		NewCard(CardDesignSpade, 7, false),
		NewCard(CardDesignClover, 7, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignDiamond, 9, false),
	}
	b := []*Card{
		NewCard(CardDesignSpade, 8, false),
		NewCard(CardDesignClover, 8, false),
		NewCard(CardDesignHeart, 8, false),
		NewCard(CardDesignDiamond, 9, false),
	}
	assert.Equal(t, -1, compareFourCardHands(a, b))
}

func TestCompareFourCardHands_ThreeOfAKindKicker(t *testing.T) {
	a := []*Card{
		NewCard(CardDesignSpade, 7, false),
		NewCard(CardDesignClover, 7, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignDiamond, 9, false),
	}
	b := []*Card{
		NewCard(CardDesignSpade, 7, false),
		NewCard(CardDesignClover, 7, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignDiamond, 10, false),
	}
	assert.Equal(t, -1, compareFourCardHands(a, b))
}

func TestCompareFourCardHands_FourOfAKindCompare(t *testing.T) {
	a := []*Card{
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignClover, 5, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignDiamond, 5, false),
	}
	b := []*Card{
		NewCard(CardDesignSpade, 9, false),
		NewCard(CardDesignClover, 9, false),
		NewCard(CardDesignHeart, 9, false),
		NewCard(CardDesignDiamond, 9, false),
	}
	assert.Equal(t, -1, compareFourCardHands(a, b))
	aceQuads := []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignClover, 1, false),
		NewCard(CardDesignHeart, 1, false),
		NewCard(CardDesignDiamond, 1, false),
	}
	assert.Equal(t, 1, compareFourCardHands(aceQuads, b))
}

func TestCompareFourCardHands_FlushCompareHighCards(t *testing.T) {
	a := []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignSpade, 7, false),
		NewCard(CardDesignSpade, 11, false),
	}
	b := []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignSpade, 8, false),
		NewCard(CardDesignSpade, 11, false),
	}
	assert.Equal(t, -1, compareFourCardHands(a, b))
}

// pickBestFour must select the optimal 4-card hand out of 5.
func TestPickBestFour_PicksFlushOverStraight(t *testing.T) {
	// 4 spades present (forming flush) + one off-suit straight filler — flush should win.
	cards := []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignSpade, 9, false),
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignClover, 6, false),
	}
	best := pickBestFour(cards)
	assert.Equal(t, FourCardHandFlush, evalFourCardHand(best))
}

func TestPickBestFourFromFive_FromSix(t *testing.T) {
	// Five cards passed should still pick the best four; six-card variant for dealer
	cards := []*Card{
		NewCard(CardDesignSpade, 9, false),
		NewCard(CardDesignClover, 9, false),
		NewCard(CardDesignHeart, 9, false),
		NewCard(CardDesignDiamond, 9, false),
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignClover, 7, false),
	}
	best := pickBestFour(cards)
	assert.Equal(t, FourCardHandFourOfAKind, evalFourCardHand(best))
}
