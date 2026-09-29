//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFiveCardStudRankCards(t *testing.T) {
	card := func(suit, value int) *Card { return NewCard(suit, value, false) }
	tests := []struct {
		name string
		hand []*Card
		rank int
		soko bool
		want []*Card
	}{
		{"high card ace high", []*Card{card(0, 2), card(1, 1), card(2, 9), card(3, 5), card(0, 13)}, PokerHandHighCard, false, []*Card{card(1, 1)}},
		{"one pair", []*Card{card(0, 8), card(1, 8), card(2, 9), card(3, 5), card(0, 2)}, PokerHandOnePair, false, []*Card{card(0, 8), card(1, 8)}},
		{"two pair", []*Card{card(0, 8), card(1, 8), card(2, 9), card(3, 9), card(0, 2)}, PokerHandTwoPair, false, []*Card{card(0, 8), card(1, 8), card(2, 9), card(3, 9)}},
		{"three of a kind", []*Card{card(0, 8), card(1, 8), card(2, 8), card(3, 9), card(0, 2)}, PokerHandThreeOfAKind, false, []*Card{card(0, 8), card(1, 8), card(2, 8)}},
		{"straight", []*Card{card(0, 5), card(1, 6), card(2, 7), card(3, 8), card(0, 9)}, PokerHandStraight, false, []*Card{card(0, 5), card(1, 6), card(2, 7), card(3, 8), card(0, 9)}},
		{"flush", []*Card{card(0, 2), card(0, 5), card(0, 7), card(0, 9), card(0, 13)}, PokerHandFlush, false, []*Card{card(0, 2), card(0, 5), card(0, 7), card(0, 9), card(0, 13)}},
		{"full house", []*Card{card(0, 8), card(1, 8), card(2, 8), card(3, 9), card(0, 9)}, PokerHandFullHouse, false, []*Card{card(0, 8), card(1, 8), card(2, 8), card(3, 9), card(0, 9)}},
		{"four of a kind", []*Card{card(0, 8), card(1, 8), card(2, 8), card(3, 8), card(0, 9)}, PokerHandFourOfAKind, false, []*Card{card(0, 8), card(1, 8), card(2, 8), card(3, 8)}},
		{"straight flush", []*Card{card(0, 5), card(0, 6), card(0, 7), card(0, 8), card(0, 9)}, PokerHandStraightFlush, false, []*Card{card(0, 5), card(0, 6), card(0, 7), card(0, 8), card(0, 9)}},
		{"royal flush", []*Card{card(0, 1), card(0, 10), card(0, 11), card(0, 12), card(0, 13)}, PokerHandRoyalFlush, false, []*Card{card(0, 1), card(0, 10), card(0, 11), card(0, 12), card(0, 13)}},
		{"soko high card ace", []*Card{card(0, 2), card(1, 1), card(2, 9), card(3, 5), card(0, 13)}, SokoHandHighCard, true, []*Card{card(1, 1)}},
		{"soko high card king", []*Card{card(0, 12), card(1, 9), card(2, 13), card(3, 5), card(0, 2)}, SokoHandHighCard, true, []*Card{card(2, 13)}},
		{"soko one pair", []*Card{card(0, 8), card(1, 8), card(2, 9), card(3, 5), card(0, 2)}, SokoHandOnePair, true, []*Card{card(0, 8), card(1, 8)}},
		{"soko two pair", []*Card{card(0, 8), card(1, 8), card(2, 9), card(3, 9), card(0, 2)}, SokoHandTwoPair, true, []*Card{card(0, 8), card(1, 8), card(2, 9), card(3, 9)}},
		{"soko three of a kind", []*Card{card(0, 8), card(1, 8), card(2, 8), card(3, 9), card(0, 2)}, SokoHandThreeOfAKind, true, []*Card{card(0, 8), card(1, 8), card(2, 8)}},
		{"soko four of a kind", []*Card{card(0, 8), card(1, 8), card(2, 8), card(3, 8), card(0, 9)}, SokoHandFourOfAKind, true, []*Card{card(0, 8), card(1, 8), card(2, 8), card(3, 8)}},
		{"soko four straight ace low", []*Card{card(0, 1), card(1, 2), card(2, 3), card(3, 4), card(0, 9)}, SokoHandFourStraight, true, []*Card{card(0, 1), card(1, 2), card(2, 3), card(3, 4)}},
		{"soko four flush", []*Card{card(0, 2), card(0, 5), card(0, 7), card(0, 9), card(1, 13)}, SokoHandFourFlush, true, []*Card{card(0, 2), card(0, 5), card(0, 7), card(0, 9)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FiveCardStudRankCards(tt.hand, tt.rank, tt.soko)
			assertCardsEqual(t, tt.want, got)
		})
	}

	hand := []*Card{card(0, 8), card(1, 8), card(2, 9), card(3, 9), card(0, 2)}
	first := FiveCardStudRankCards(hand, PokerHandTwoPair, false)
	second := FiveCardStudRankCards(hand, PokerHandTwoPair, false)
	assertCardsEqual(t, first, second)
}

func assertCardsEqual(t *testing.T, want, got []*Card) {
	t.Helper()
	toPairs := func(cards []*Card) [][2]int {
		pairs := make([][2]int, 0, len(cards))
		for _, c := range cards {
			pairs = append(pairs, [2]int{c.GetDesign(), c.GetValue()})
		}
		return pairs
	}
	assert.Equal(t, toPairs(want), toPairs(got))
}
