//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCardValueDisplayName(t *testing.T) {
	tests := []struct {
		value    int
		expected string
	}{
		{1, "A"},
		{2, "2"},
		{3, "3"},
		{4, "4"},
		{5, "5"},
		{6, "6"},
		{7, "7"},
		{8, "8"},
		{9, "9"},
		{10, "10"},
		{11, "J"},
		{12, "Q"},
		{13, "K"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, CardValueDisplayName(tt.value))
	}
}

func TestExtractKickers(t *testing.T) {
	t.Run("HighCard returns nil", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 5, false),
			NewCard(CardDesignClover, 7, false),
			NewCard(CardDesignDiamond, 9, false),
			NewCard(CardDesignSpade, 11, false),
		}
		assert.Nil(t, ExtractKickers(cards, PokerHandHighCard))
	})

	t.Run("OnePair returns 3 kickers", func(t *testing.T) {
		// Pair of 5s, kickers A, Q, 10
		cards := []*Card{
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignHeart, 5, false),
			NewCard(CardDesignClover, 1, false),
			NewCard(CardDesignDiamond, 12, false),
			NewCard(CardDesignSpade, 10, false),
		}
		kickers := ExtractKickers(cards, PokerHandOnePair)
		assert.Equal(t, []int{14, 12, 10}, kickers)
	})

	t.Run("TwoPair returns 1 kicker", func(t *testing.T) {
		// Pair of Ks, pair of 5s, kicker 8
		cards := []*Card{
			NewCard(CardDesignSpade, 13, false),
			NewCard(CardDesignHeart, 13, false),
			NewCard(CardDesignClover, 5, false),
			NewCard(CardDesignDiamond, 5, false),
			NewCard(CardDesignSpade, 8, false),
		}
		kickers := ExtractKickers(cards, PokerHandTwoPair)
		assert.Equal(t, []int{8}, kickers)
	})

	t.Run("ThreeOfAKind returns 2 kickers", func(t *testing.T) {
		// Three 7s, kickers A, 3
		cards := []*Card{
			NewCard(CardDesignSpade, 7, false),
			NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignClover, 7, false),
			NewCard(CardDesignDiamond, 1, false),
			NewCard(CardDesignSpade, 3, false),
		}
		kickers := ExtractKickers(cards, PokerHandThreeOfAKind)
		assert.Equal(t, []int{14, 3}, kickers)
	})

	t.Run("Straight returns nil", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignHeart, 6, false),
			NewCard(CardDesignClover, 7, false),
			NewCard(CardDesignDiamond, 8, false),
			NewCard(CardDesignSpade, 9, false),
		}
		assert.Nil(t, ExtractKickers(cards, PokerHandStraight))
	})

	t.Run("Flush returns nil", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 2, false),
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignSpade, 7, false),
			NewCard(CardDesignSpade, 9, false),
			NewCard(CardDesignSpade, 11, false),
		}
		assert.Nil(t, ExtractKickers(cards, PokerHandFlush))
	})

	t.Run("FullHouse returns nil", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 10, false),
			NewCard(CardDesignHeart, 10, false),
			NewCard(CardDesignClover, 10, false),
			NewCard(CardDesignDiamond, 5, false),
			NewCard(CardDesignSpade, 5, false),
		}
		assert.Nil(t, ExtractKickers(cards, PokerHandFullHouse))
	})

	t.Run("FourOfAKind returns 1 kicker", func(t *testing.T) {
		// Four 9s, kicker A
		cards := []*Card{
			NewCard(CardDesignSpade, 9, false),
			NewCard(CardDesignHeart, 9, false),
			NewCard(CardDesignClover, 9, false),
			NewCard(CardDesignDiamond, 9, false),
			NewCard(CardDesignSpade, 1, false),
		}
		kickers := ExtractKickers(cards, PokerHandFourOfAKind)
		assert.Equal(t, []int{14}, kickers)
	})

	t.Run("StraightFlush returns nil", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignHeart, 5, false),
			NewCard(CardDesignHeart, 6, false),
			NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignHeart, 8, false),
			NewCard(CardDesignHeart, 9, false),
		}
		assert.Nil(t, ExtractKickers(cards, PokerHandStraightFlush))
	})

	t.Run("RoyalFlush returns nil", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignSpade, 10, false),
			NewCard(CardDesignSpade, 11, false),
			NewCard(CardDesignSpade, 12, false),
			NewCard(CardDesignSpade, 13, false),
		}
		assert.Nil(t, ExtractKickers(cards, PokerHandRoyalFlush))
	})

	t.Run("FiveOfAKind returns nil", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 1, false),
			NewCard(CardDesignClover, 1, false),
			NewCard(CardDesignDiamond, 1, false),
			NewCard(CardDesignJoker, 0, false),
		}
		assert.Nil(t, ExtractKickers(cards, PokerHandFiveOfAKind))
	})

	t.Run("nil cards returns nil", func(t *testing.T) {
		assert.Nil(t, ExtractKickers(nil, PokerHandOnePair))
	})

	t.Run("wrong card count returns nil", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 5, false),
		}
		assert.Nil(t, ExtractKickers(cards, PokerHandOnePair))
	})

	t.Run("TwoPair with ace kicker", func(t *testing.T) {
		// Pair of 10s, pair of 3s, kicker A
		cards := []*Card{
			NewCard(CardDesignSpade, 10, false),
			NewCard(CardDesignHeart, 10, false),
			NewCard(CardDesignClover, 3, false),
			NewCard(CardDesignDiamond, 3, false),
			NewCard(CardDesignSpade, 1, false),
		}
		kickers := ExtractKickers(cards, PokerHandTwoPair)
		assert.Equal(t, []int{14}, kickers)
	})
}

func TestFormatKickers(t *testing.T) {
	t.Run("empty kickers", func(t *testing.T) {
		assert.Equal(t, "", FormatKickers(nil))
		assert.Equal(t, "", FormatKickers([]int{}))
	})

	t.Run("single kicker", func(t *testing.T) {
		assert.Equal(t, "8", FormatKickers([]int{8}))
	})

	t.Run("multiple kickers", func(t *testing.T) {
		assert.Equal(t, "A, Q, 10", FormatKickers([]int{14, 12, 10}))
	})

	t.Run("ace kicker", func(t *testing.T) {
		assert.Equal(t, "A", FormatKickers([]int{14}))
	})

	t.Run("face cards", func(t *testing.T) {
		assert.Equal(t, "K, J", FormatKickers([]int{13, 11}))
	})
}
