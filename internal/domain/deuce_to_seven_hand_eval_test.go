//go:build test

package domain

import "testing"

// d27Cards spells out a hand in test tables. Each pair is (design, value);
// design uses CardDesignSpade/Clover/Heart/Diamond, value uses Ace = 1.
func d27Cards(pairs ...[2]int) []*Card {
	out := make([]*Card, len(pairs))
	for i, p := range pairs {
		out[i] = NewCard(p[0], p[1], true)
	}
	return out
}

func TestEvalDeuceToSevenHand(t *testing.T) {
	S, C, H, D := CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond
	tests := []struct {
		name string
		hand []*Card
		want int
	}{
		{
			name: "nut low 7-5-4-3-2 is high card",
			hand: d27Cards([2]int{S, 7}, [2]int{H, 5}, [2]int{D, 4}, [2]int{C, 3}, [2]int{S, 2}),
			want: PokerHandHighCard,
		},
		{
			name: "A-2-3-4-5 is NOT a straight (ace high) -> high card",
			hand: d27Cards([2]int{S, 1}, [2]int{H, 2}, [2]int{D, 3}, [2]int{C, 4}, [2]int{S, 5}),
			want: PokerHandHighCard,
		},
		{
			name: "2-3-4-5-6 is a straight",
			hand: d27Cards([2]int{S, 2}, [2]int{H, 3}, [2]int{D, 4}, [2]int{C, 5}, [2]int{S, 6}),
			want: PokerHandStraight,
		},
		{
			name: "broadway A-10-J-Q-K is a straight",
			hand: d27Cards([2]int{S, 1}, [2]int{H, 10}, [2]int{D, 11}, [2]int{C, 12}, [2]int{S, 13}),
			want: PokerHandStraight,
		},
		{
			name: "flush counts",
			hand: d27Cards([2]int{S, 2}, [2]int{S, 4}, [2]int{S, 6}, [2]int{S, 8}, [2]int{S, 10}),
			want: PokerHandFlush,
		},
		{
			name: "one pair",
			hand: d27Cards([2]int{S, 2}, [2]int{H, 2}, [2]int{D, 4}, [2]int{C, 6}, [2]int{S, 8}),
			want: PokerHandOnePair,
		},
		{
			name: "two pair",
			hand: d27Cards([2]int{S, 2}, [2]int{H, 2}, [2]int{D, 4}, [2]int{C, 4}, [2]int{S, 8}),
			want: PokerHandTwoPair,
		},
		{
			name: "three of a kind",
			hand: d27Cards([2]int{S, 2}, [2]int{H, 2}, [2]int{D, 2}, [2]int{C, 6}, [2]int{S, 8}),
			want: PokerHandThreeOfAKind,
		},
		{
			name: "full house",
			hand: d27Cards([2]int{S, 2}, [2]int{H, 2}, [2]int{D, 2}, [2]int{C, 6}, [2]int{S, 6}),
			want: PokerHandFullHouse,
		},
		{
			name: "four of a kind",
			hand: d27Cards([2]int{S, 2}, [2]int{H, 2}, [2]int{D, 2}, [2]int{C, 2}, [2]int{S, 8}),
			want: PokerHandFourOfAKind,
		},
		{
			name: "straight flush 2-6",
			hand: d27Cards([2]int{S, 2}, [2]int{S, 3}, [2]int{S, 4}, [2]int{S, 5}, [2]int{S, 6}),
			want: PokerHandStraightFlush,
		},
		{
			name: "royal flush (broadway straight flush)",
			hand: d27Cards([2]int{S, 1}, [2]int{S, 10}, [2]int{S, 11}, [2]int{S, 12}, [2]int{S, 13}),
			want: PokerHandRoyalFlush,
		},
		{
			name: "wrong length returns high card",
			hand: d27Cards([2]int{S, 2}, [2]int{H, 3}),
			want: PokerHandHighCard,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := evalDeuceToSevenHand(tt.hand); got != tt.want {
				t.Errorf("evalDeuceToSevenHand() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCompareDeuceToSevenCards(t *testing.T) {
	S, C, H, D := CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond
	nut := d27Cards([2]int{S, 7}, [2]int{H, 5}, [2]int{D, 4}, [2]int{C, 3}, [2]int{S, 2})        // 7-5-4-3-2
	eightLow := d27Cards([2]int{S, 8}, [2]int{H, 5}, [2]int{D, 4}, [2]int{C, 3}, [2]int{S, 2})   // 8-5-4-3-2
	sevenSix := d27Cards([2]int{S, 7}, [2]int{H, 6}, [2]int{D, 4}, [2]int{C, 3}, [2]int{S, 2})   // 7-6-4-3-2
	aceHigh := d27Cards([2]int{S, 1}, [2]int{H, 5}, [2]int{D, 4}, [2]int{C, 3}, [2]int{S, 2})    // A-5-4-3-2
	pairTwos := d27Cards([2]int{S, 2}, [2]int{H, 2}, [2]int{D, 4}, [2]int{C, 6}, [2]int{S, 8})   // pair of 2s
	pairThrees := d27Cards([2]int{S, 3}, [2]int{H, 3}, [2]int{D, 4}, [2]int{C, 6}, [2]int{S, 8}) // pair of 3s
	straight := d27Cards([2]int{S, 2}, [2]int{H, 3}, [2]int{D, 4}, [2]int{C, 5}, [2]int{S, 6})   // straight

	tests := []struct {
		name string
		a, b []*Card
		want int
	}{
		{"nut beats eight low", nut, eightLow, -1},
		{"nut beats seven-six", nut, sevenSix, -1},
		{"seven-six beats eight low", sevenSix, eightLow, -1},
		{"ace high loses to nut", aceHigh, nut, 1},
		{"no pair beats one pair", eightLow, pairTwos, -1},
		{"lower pair beats higher pair", pairTwos, pairThrees, -1},
		{"one pair beats straight", pairThrees, straight, -1},
		{"identical ranks tie", nut, d27Cards([2]int{D, 7}, [2]int{C, 5}, [2]int{H, 4}, [2]int{S, 3}, [2]int{D, 2}), 0},
		{"symmetry: eight low loses to nut", eightLow, nut, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compareDeuceToSevenCards(tt.a, tt.b); got != tt.want {
				t.Errorf("compareDeuceToSevenCards() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestDeuceLowStrength(t *testing.T) {
	S, C, H, D := CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond
	tests := []struct {
		name string
		hand []*Card
		want int
	}{
		{"nut low is 4", d27Cards([2]int{S, 7}, [2]int{H, 5}, [2]int{D, 4}, [2]int{C, 3}, [2]int{S, 2}), 4},
		{"eight low is 4", d27Cards([2]int{S, 8}, [2]int{H, 6}, [2]int{D, 4}, [2]int{C, 3}, [2]int{S, 2}), 4},
		{"ten high is 3", d27Cards([2]int{S, 10}, [2]int{H, 6}, [2]int{D, 4}, [2]int{C, 3}, [2]int{S, 2}), 3},
		{"jack high is 3", d27Cards([2]int{S, 11}, [2]int{H, 6}, [2]int{D, 4}, [2]int{C, 3}, [2]int{S, 2}), 3},
		{"king high is 2", d27Cards([2]int{S, 13}, [2]int{H, 6}, [2]int{D, 4}, [2]int{C, 3}, [2]int{S, 2}), 2},
		{"ace high is 2", d27Cards([2]int{S, 1}, [2]int{H, 6}, [2]int{D, 4}, [2]int{C, 3}, [2]int{S, 2}), 2},
		{"one pair is 1", d27Cards([2]int{S, 2}, [2]int{H, 2}, [2]int{D, 4}, [2]int{C, 6}, [2]int{S, 8}), 1},
		{"straight is 1", d27Cards([2]int{S, 2}, [2]int{H, 3}, [2]int{D, 4}, [2]int{C, 5}, [2]int{S, 6}), 1},
		{"wrong length is 1", d27Cards([2]int{S, 2}), 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := deuceLowStrength(tt.hand); got != tt.want {
				t.Errorf("deuceLowStrength() = %d, want %d", got, tt.want)
			}
		})
	}
}
