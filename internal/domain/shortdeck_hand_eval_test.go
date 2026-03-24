//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvalShortDeckFiveCardHand(t *testing.T) {
	tests := []struct {
		name     string
		cards    []*Card
		expected int
	}{
		{
			name: "Royal Flush",
			cards: []*Card{
				NewCard(CardDesignSpade, 1, false),
				NewCard(CardDesignSpade, 10, false),
				NewCard(CardDesignSpade, 11, false),
				NewCard(CardDesignSpade, 12, false),
				NewCard(CardDesignSpade, 13, false),
			},
			expected: ShortDeckHandRoyalFlush,
		},
		{
			name: "Straight Flush",
			cards: []*Card{
				NewCard(CardDesignHeart, 6, false),
				NewCard(CardDesignHeart, 7, false),
				NewCard(CardDesignHeart, 8, false),
				NewCard(CardDesignHeart, 9, false),
				NewCard(CardDesignHeart, 10, false),
			},
			expected: ShortDeckHandStraightFlush,
		},
		{
			name: "Straight Flush with Short Deck wheel A-6-7-8-9",
			cards: []*Card{
				NewCard(CardDesignClover, 1, false),
				NewCard(CardDesignClover, 6, false),
				NewCard(CardDesignClover, 7, false),
				NewCard(CardDesignClover, 8, false),
				NewCard(CardDesignClover, 9, false),
			},
			expected: ShortDeckHandStraightFlush,
		},
		{
			name: "Four of a Kind",
			cards: []*Card{
				NewCard(CardDesignSpade, 10, false),
				NewCard(CardDesignClover, 10, false),
				NewCard(CardDesignHeart, 10, false),
				NewCard(CardDesignDiamond, 10, false),
				NewCard(CardDesignSpade, 13, false),
			},
			expected: ShortDeckHandFourOfAKind,
		},
		{
			name: "Flush beats Full House in Short Deck",
			cards: []*Card{
				NewCard(CardDesignSpade, 6, false),
				NewCard(CardDesignSpade, 8, false),
				NewCard(CardDesignSpade, 10, false),
				NewCard(CardDesignSpade, 12, false),
				NewCard(CardDesignSpade, 13, false),
			},
			expected: ShortDeckHandFlush,
		},
		{
			name: "Full House",
			cards: []*Card{
				NewCard(CardDesignSpade, 9, false),
				NewCard(CardDesignClover, 9, false),
				NewCard(CardDesignHeart, 9, false),
				NewCard(CardDesignSpade, 13, false),
				NewCard(CardDesignClover, 13, false),
			},
			expected: ShortDeckHandFullHouse,
		},
		{
			name:     "Flush rank is higher than Full House rank",
			cards:    []*Card{},
			expected: -1, // sentinel: tested via assertion below
		},
		{
			name: "Straight with Short Deck wheel A-6-7-8-9",
			cards: []*Card{
				NewCard(CardDesignSpade, 1, false),
				NewCard(CardDesignClover, 6, false),
				NewCard(CardDesignHeart, 7, false),
				NewCard(CardDesignDiamond, 8, false),
				NewCard(CardDesignSpade, 9, false),
			},
			expected: ShortDeckHandStraight,
		},
		{
			name: "Straight 6-7-8-9-10",
			cards: []*Card{
				NewCard(CardDesignSpade, 6, false),
				NewCard(CardDesignClover, 7, false),
				NewCard(CardDesignHeart, 8, false),
				NewCard(CardDesignDiamond, 9, false),
				NewCard(CardDesignSpade, 10, false),
			},
			expected: ShortDeckHandStraight,
		},
		{
			name: "Broadway straight A-10-J-Q-K",
			cards: []*Card{
				NewCard(CardDesignSpade, 1, false),
				NewCard(CardDesignClover, 10, false),
				NewCard(CardDesignHeart, 11, false),
				NewCard(CardDesignDiamond, 12, false),
				NewCard(CardDesignSpade, 13, false),
			},
			expected: ShortDeckHandStraight,
		},
		{
			name: "Three of a Kind",
			cards: []*Card{
				NewCard(CardDesignSpade, 8, false),
				NewCard(CardDesignClover, 8, false),
				NewCard(CardDesignHeart, 8, false),
				NewCard(CardDesignDiamond, 10, false),
				NewCard(CardDesignSpade, 13, false),
			},
			expected: ShortDeckHandThreeOfAKind,
		},
		{
			name: "Two Pair",
			cards: []*Card{
				NewCard(CardDesignSpade, 9, false),
				NewCard(CardDesignClover, 9, false),
				NewCard(CardDesignHeart, 13, false),
				NewCard(CardDesignDiamond, 13, false),
				NewCard(CardDesignSpade, 6, false),
			},
			expected: ShortDeckHandTwoPair,
		},
		{
			name: "One Pair",
			cards: []*Card{
				NewCard(CardDesignSpade, 1, false),
				NewCard(CardDesignClover, 1, false),
				NewCard(CardDesignHeart, 7, false),
				NewCard(CardDesignDiamond, 9, false),
				NewCard(CardDesignSpade, 12, false),
			},
			expected: ShortDeckHandOnePair,
		},
		{
			name: "High Card",
			cards: []*Card{
				NewCard(CardDesignSpade, 1, false),
				NewCard(CardDesignClover, 6, false),
				NewCard(CardDesignHeart, 8, false),
				NewCard(CardDesignDiamond, 10, false),
				NewCard(CardDesignSpade, 12, false),
			},
			expected: ShortDeckHandHighCard,
		},
		{
			name: "Less than 5 cards returns HighCard",
			cards: []*Card{
				NewCard(CardDesignSpade, 1, false),
				NewCard(CardDesignClover, 6, false),
			},
			expected: ShortDeckHandHighCard,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "Flush rank is higher than Full House rank" {
				assert.Greater(t, ShortDeckHandFlush, ShortDeckHandFullHouse,
					"In Short Deck, Flush must rank higher than Full House")
				return
			}
			got := evalShortDeckFiveCardHand(tt.cards)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestCheckShortDeckStraightValues(t *testing.T) {
	tests := []struct {
		name     string
		values   []int
		expected bool
	}{
		{"Short Deck wheel A-6-7-8-9", []int{1, 6, 7, 8, 9}, true},
		{"Broadway A-10-J-Q-K", []int{1, 10, 11, 12, 13}, true},
		{"Normal straight 6-7-8-9-10", []int{6, 7, 8, 9, 10}, true},
		{"Normal straight 9-10-J-Q-K", []int{9, 10, 11, 12, 13}, true},
		{"Not a straight A-6-8-10-K", []int{1, 6, 8, 10, 13}, false},
		{"Not a straight 6-7-9-10-J", []int{6, 7, 9, 10, 11}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkShortDeckStraightValues(tt.values)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestIsShortDeckWheelHand(t *testing.T) {
	wheel := []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignClover, 6, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignDiamond, 8, false),
		NewCard(CardDesignSpade, 9, false),
	}
	assert.True(t, isShortDeckWheelHand(wheel))

	notWheel := []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignClover, 10, false),
		NewCard(CardDesignHeart, 11, false),
		NewCard(CardDesignDiamond, 12, false),
		NewCard(CardDesignSpade, 13, false),
	}
	assert.False(t, isShortDeckWheelHand(notWheel))

	// fewer than 5 cards
	assert.False(t, isShortDeckWheelHand([]*Card{NewCard(CardDesignSpade, 1, false)}))
}

func TestCompareShortDeckHighCardsSlice(t *testing.T) {
	// Wheel (A-6-7-8-9) should lose to a higher straight (6-7-8-9-10)
	wheel := []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignClover, 6, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignDiamond, 8, false),
		NewCard(CardDesignSpade, 9, false),
	}
	higher := []*Card{
		NewCard(CardDesignSpade, 6, false),
		NewCard(CardDesignClover, 7, false),
		NewCard(CardDesignHeart, 8, false),
		NewCard(CardDesignDiamond, 9, false),
		NewCard(CardDesignSpade, 10, false),
	}
	assert.Equal(t, -1, compareShortDeckHighCardsSlice(wheel, higher))
	assert.Equal(t, 1, compareShortDeckHighCardsSlice(higher, wheel))

	// Empty slices
	assert.Equal(t, 0, compareShortDeckHighCardsSlice(nil, higher))
	assert.Equal(t, 0, compareShortDeckHighCardsSlice(wheel, nil))
}

func TestNewTrumpCardsShortDeck(t *testing.T) {
	deck := NewTrumpCardsShortDeck()
	assert.Equal(t, 36, deck.GetTotalCount())

	validValues := map[int]bool{1: true, 6: true, 7: true, 8: true, 9: true, 10: true, 11: true, 12: true, 13: true}
	suitCounts := make(map[int]int)
	for range 36 {
		card := deck.DrawCard()
		assert.NotNil(t, card)
		assert.True(t, validValues[card.GetValue()], "unexpected value: %d", card.GetValue())
		assert.True(t, card.GetDesign() >= CardDesignSpade && card.GetDesign() <= CardDesignDiamond)
		suitCounts[card.GetDesign()]++
	}
	// 37枚目はnil
	assert.Nil(t, deck.DrawCard())
	// 各スート9枚
	for _, cnt := range suitCounts {
		assert.Equal(t, 9, cnt)
	}
}

func TestShortDeckHandNames(t *testing.T) {
	assert.Equal(t, "Full House", ShortDeckHandNames[ShortDeckHandFullHouse])
	assert.Equal(t, "Flush", ShortDeckHandNames[ShortDeckHandFlush])
	assert.Equal(t, len(ShortDeckHandNames), 10)
}
