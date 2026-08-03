package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- BJSideBetResult.BetTypeName ---

func TestBJSideBetResult_BetTypeName(t *testing.T) {
	tests := []struct {
		name     string
		betType  int
		expected string
	}{
		{"PerfectPairs", BJSideBetPerfectPairs, "Perfect Pairs"},
		{"21Plus3", BJSideBet21Plus3, "Poker Hand Bonus"},
		{"Unknown", 99, "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &BJSideBetResult{BetType: tt.betType}
			assert.Equal(t, tt.expected, r.BetTypeName())
		})
	}
}

// --- EvaluatePerfectPairs ---

func TestEvaluatePerfectPairs(t *testing.T) {
	tests := []struct {
		name     string
		card1    *Card
		card2    *Card
		wantType int
		wantName string
	}{
		{
			"PerfectPair - same suit same value",
			NewCard(CardDesignSpade, 10, false),
			NewCard(CardDesignSpade, 10, false),
			BJPPPerfectPair, "Perfect Pair",
		},
		{
			"ColoredPair - same color different suit",
			NewCard(CardDesignSpade, 7, false),
			NewCard(CardDesignClover, 7, false),
			BJPPColoredPair, "Colored Pair",
		},
		{
			"ColoredPair - red suits",
			NewCard(CardDesignHeart, 5, false),
			NewCard(CardDesignDiamond, 5, false),
			BJPPColoredPair, "Colored Pair",
		},
		{
			"MixedPair - different color same value",
			NewCard(CardDesignSpade, 8, false),
			NewCard(CardDesignHeart, 8, false),
			BJPPMixedPair, "Mixed Pair",
		},
		{
			"NoPair - different values",
			NewCard(CardDesignSpade, 10, false),
			NewCard(CardDesignSpade, 9, false),
			BJPPNone, "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultType, resultName := EvaluatePerfectPairs(tt.card1, tt.card2)
			assert.Equal(t, tt.wantType, resultType)
			assert.Equal(t, tt.wantName, resultName)
		})
	}
}

// --- Evaluate21Plus3 ---

func TestEvaluate21Plus3(t *testing.T) {
	tests := []struct {
		name         string
		card1        *Card
		card2        *Card
		dealerUpcard *Card
		wantType     int
		wantName     string
	}{
		{
			"SuitedTrips - same suit same value",
			NewCard(CardDesignSpade, 7, false),
			NewCard(CardDesignSpade, 7, false),
			NewCard(CardDesignSpade, 7, false),
			BJT3SuitedTrips, "Suited Trips",
		},
		{
			"StraightFlush - same suit consecutive",
			NewCard(CardDesignHeart, 5, false),
			NewCard(CardDesignHeart, 6, false),
			NewCard(CardDesignHeart, 7, false),
			BJT3StraightFlush, "Straight Flush",
		},
		{
			"StraightFlush - A-2-3 wrap",
			NewCard(CardDesignDiamond, 1, false),
			NewCard(CardDesignDiamond, 2, false),
			NewCard(CardDesignDiamond, 3, false),
			BJT3StraightFlush, "Straight Flush",
		},
		{
			"StraightFlush - Q-K-A wrap",
			NewCard(CardDesignClover, 12, false),
			NewCard(CardDesignClover, 13, false),
			NewCard(CardDesignClover, 1, false),
			BJT3StraightFlush, "Straight Flush",
		},
		{
			"ThreeOfAKind - same value different suits",
			NewCard(CardDesignSpade, 9, false),
			NewCard(CardDesignHeart, 9, false),
			NewCard(CardDesignDiamond, 9, false),
			BJT3ThreeOfAKind, "Three of a Kind",
		},
		{
			"Straight - consecutive different suits",
			NewCard(CardDesignSpade, 4, false),
			NewCard(CardDesignHeart, 5, false),
			NewCard(CardDesignDiamond, 6, false),
			BJT3Straight, "Straight",
		},
		{
			"Straight - Q-K-A wrap different suits",
			NewCard(CardDesignSpade, 12, false),
			NewCard(CardDesignHeart, 13, false),
			NewCard(CardDesignDiamond, 1, false),
			BJT3Straight, "Straight",
		},
		{
			"Flush - same suit not consecutive not same value",
			NewCard(CardDesignClover, 2, false),
			NewCard(CardDesignClover, 5, false),
			NewCard(CardDesignClover, 9, false),
			BJT3Flush, "Flush",
		},
		{
			"None - no match",
			NewCard(CardDesignSpade, 2, false),
			NewCard(CardDesignHeart, 5, false),
			NewCard(CardDesignDiamond, 9, false),
			BJT3None, "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultType, resultName := Evaluate21Plus3(tt.card1, tt.card2, tt.dealerUpcard)
			assert.Equal(t, tt.wantType, resultType)
			assert.Equal(t, tt.wantName, resultName)
		})
	}
}

// --- PerfectPairsPayout ---

func TestPerfectPairsPayout(t *testing.T) {
	tests := []struct {
		name       string
		resultType int
		expected   int
	}{
		{"PerfectPair", BJPPPerfectPair, BJPPPerfectPairPayout},
		{"ColoredPair", BJPPColoredPair, BJPPColoredPairPayout},
		{"MixedPair", BJPPMixedPair, BJPPMixedPairPayout},
		{"None", BJPPNone, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, PerfectPairsPayout(tt.resultType))
		})
	}
}

// --- TwentyOnePlus3Payout ---

func TestTwentyOnePlus3Payout(t *testing.T) {
	tests := []struct {
		name       string
		resultType int
		expected   int
	}{
		{"SuitedTrips", BJT3SuitedTrips, BJT3SuitedTripsPayout},
		{"StraightFlush", BJT3StraightFlush, BJT3StraightFlushPayout},
		{"ThreeOfAKind", BJT3ThreeOfAKind, BJT3ThreeOfAKindPayout},
		{"Straight", BJT3Straight, BJT3StraightPayout},
		{"Flush", BJT3Flush, BJT3FlushPayout},
		{"None", BJT3None, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, TwentyOnePlus3Payout(tt.resultType))
		})
	}
}

// --- isSameSuit ---

func TestIsSameSuit(t *testing.T) {
	assert.True(t, isSameSuit(NewCard(CardDesignSpade, 1, false), NewCard(CardDesignSpade, 2, false)))
	assert.False(t, isSameSuit(NewCard(CardDesignSpade, 1, false), NewCard(CardDesignHeart, 1, false)))
}

// --- isSameColor ---

func TestIsSameColor(t *testing.T) {
	// black + black
	assert.True(t, isSameColor(NewCard(CardDesignSpade, 1, false), NewCard(CardDesignClover, 1, false)))
	// red + red
	assert.True(t, isSameColor(NewCard(CardDesignHeart, 1, false), NewCard(CardDesignDiamond, 1, false)))
	// black + red
	assert.False(t, isSameColor(NewCard(CardDesignSpade, 1, false), NewCard(CardDesignHeart, 1, false)))
}

// --- cardColor ---

func TestCardColor(t *testing.T) {
	assert.Equal(t, 0, cardColor(NewCard(CardDesignSpade, 1, false)))
	assert.Equal(t, 0, cardColor(NewCard(CardDesignClover, 1, false)))
	assert.Equal(t, 1, cardColor(NewCard(CardDesignHeart, 1, false)))
	assert.Equal(t, 1, cardColor(NewCard(CardDesignDiamond, 1, false)))
	assert.Equal(t, 0, cardColor(NewCard(CardDesignJoker, 0, false)))
}

// --- isSameValue ---

func TestIsSameValue(t *testing.T) {
	assert.True(t, isSameValue(NewCard(CardDesignSpade, 5, false), NewCard(CardDesignHeart, 5, false)))
	assert.False(t, isSameValue(NewCard(CardDesignSpade, 5, false), NewCard(CardDesignHeart, 6, false)))
}

// --- isFlush3 ---

func TestIsFlush3(t *testing.T) {
	assert.True(t, isFlush3(
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignSpade, 10, false),
	))
	assert.False(t, isFlush3(
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignSpade, 10, false),
	))
}

// --- isStraight3 ---

func TestIsStraight3(t *testing.T) {
	// normal straight
	assert.True(t, isStraight3(
		NewCard(CardDesignSpade, 4, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignDiamond, 6, false),
	))
	// A-2-3
	assert.True(t, isStraight3(
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignDiamond, 3, false),
	))
	// Q-K-A
	assert.True(t, isStraight3(
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignHeart, 13, false),
		NewCard(CardDesignDiamond, 1, false),
	))
	// not straight
	assert.False(t, isStraight3(
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignDiamond, 9, false),
	))
	// K-A-2 is NOT a straight (no wrap in middle)
	assert.False(t, isStraight3(
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignHeart, 1, false),
		NewCard(CardDesignDiamond, 2, false),
	))
}

// --- isThreeOfAKind3 ---

func TestIsThreeOfAKind3(t *testing.T) {
	assert.True(t, isThreeOfAKind3(
		NewCard(CardDesignSpade, 8, false),
		NewCard(CardDesignHeart, 8, false),
		NewCard(CardDesignDiamond, 8, false),
	))
	assert.False(t, isThreeOfAKind3(
		NewCard(CardDesignSpade, 8, false),
		NewCard(CardDesignHeart, 8, false),
		NewCard(CardDesignDiamond, 9, false),
	))
}
