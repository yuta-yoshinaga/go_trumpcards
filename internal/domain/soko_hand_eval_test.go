//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// The card shorthands sp/he/di/cl are the package-level helpers already defined
// in Piquet_test.go -- a fourth spelling of "make me a spade" would be one more
// thing to keep in sync.

// The two ranks Soko inserts are the whole point of the game: a four-card
// straight beats a pair, and a four-card flush beats a four-card straight.
// Everything else keeps standard poker order, shifted up by the two insertions.
func TestEvalSokoHand_Ranks(t *testing.T) {
	tests := []struct {
		name  string
		cards []*Card
		want  int
	}{
		{"high card", []*Card{sp(2), he(5), di(9), cl(11), sp(13)}, SokoHandHighCard},
		{"one pair", []*Card{sp(5), he(5), di(9), cl(11), sp(13)}, SokoHandOnePair},
		// exactly four in sequence, fifth card unconnected, no four-flush
		{"four-card straight", []*Card{sp(5), he(6), di(7), cl(8), he(13)}, SokoHandFourStraight},
		// exactly four of one suit, fifth off-suit, no pair
		{"four-card flush", []*Card{sp(2), sp(5), sp(9), sp(12), he(7)}, SokoHandFourFlush},
		{"two pair", []*Card{sp(5), he(5), di(9), cl(9), sp(13)}, SokoHandTwoPair},
		{"three of a kind", []*Card{sp(5), he(5), di(5), cl(9), he(13)}, SokoHandThreeOfAKind},
		{"straight", []*Card{sp(5), he(6), di(7), cl(8), he(9)}, SokoHandStraight},
		{"flush", []*Card{sp(2), sp(5), sp(9), sp(12), sp(7)}, SokoHandFlush},
		{"full house", []*Card{sp(5), he(5), di(5), cl(9), he(9)}, SokoHandFullHouse},
		{"four of a kind", []*Card{sp(5), he(5), di(5), cl(5), he(9)}, SokoHandFourOfAKind},
		{"straight flush", []*Card{sp(5), sp(6), sp(7), sp(8), sp(9)}, SokoHandStraightFlush},
		{"royal flush", []*Card{sp(1), sp(10), sp(11), sp(12), sp(13)}, SokoHandRoyalFlush},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, evalSokoHand(tt.cards))
		})
	}
}

// The inserted ranks must sit strictly between one pair and two pair, and the
// flush must outrank the straight. Asserting the ORDER (not just the values)
// is what fails if someone renumbers the constants.
func TestEvalSokoHand_InsertedRanksSitBetweenPairAndTwoPair(t *testing.T) {
	assert.Less(t, SokoHandOnePair, SokoHandFourStraight)
	assert.Less(t, SokoHandFourStraight, SokoHandFourFlush)
	assert.Less(t, SokoHandFourFlush, SokoHandTwoPair)
}

// A hand can be a pair AND a four-flush at once (a pair leaves four distinct
// ranks, four of which may share a suit). Soko ranks the four-flush higher, so
// the hand must play as the four-flush -- taking the standard evaluator's word
// for it would under-rank it.
func TestEvalSokoHand_FourFlushBeatsCoexistingPair(t *testing.T) {
	// ♠2 ♠5 ♠9 ♠K ♥K -- pair of kings, and four spades
	cards := []*Card{sp(2), sp(5), sp(9), sp(13), he(13)}
	assert.Equal(t, PokerHandOnePair, evalFiveCardHand(cards), "the standard evaluator sees only the pair")
	assert.Equal(t, SokoHandFourFlush, evalSokoHand(cards), "Soko must see the four-flush")
}

// Same for a four-straight coexisting with a pair: 5-6-7-8 plus a paired 8.
func TestEvalSokoHand_FourStraightBeatsCoexistingPair(t *testing.T) {
	cards := []*Card{sp(5), he(6), di(7), cl(8), he(8)}
	assert.Equal(t, PokerHandOnePair, evalFiveCardHand(cards))
	assert.Equal(t, SokoHandFourStraight, evalSokoHand(cards))
}

// Ace is low in A-2-3-4 and high in J-Q-K-A; both are four-card straights.
func TestEvalSokoHand_FourStraightAceBothEnds(t *testing.T) {
	low := []*Card{sp(1), he(2), di(3), cl(4), he(9)}
	high := []*Card{sp(11), he(12), di(13), cl(1), he(6)}
	assert.Equal(t, SokoHandFourStraight, evalSokoHand(low), "A-2-3-4")
	assert.Equal(t, SokoHandFourStraight, evalSokoHand(high), "J-Q-K-A")
}

// A gap of two is not a four-straight -- the negative control for the run
// detector, without which "any four distinct ranks" would pass.
func TestEvalSokoHand_NotAFourStraight(t *testing.T) {
	cards := []*Card{sp(5), he(6), di(7), cl(9), he(13)}
	assert.Equal(t, SokoHandHighCard, evalSokoHand(cards))
}

// Three of one suit is not a four-flush -- the negative control for the suit
// counter.
func TestEvalSokoHand_NotAFourFlush(t *testing.T) {
	cards := []*Card{sp(2), sp(5), sp(9), he(12), di(7)}
	assert.Equal(t, SokoHandHighCard, evalSokoHand(cards))
}

// A real five-card flush must NOT be demoted to a four-flush.
func TestEvalSokoHand_FiveCardFlushIsNotFourFlush(t *testing.T) {
	cards := []*Card{sp(2), sp(5), sp(9), sp(12), sp(7)}
	assert.Equal(t, SokoHandFlush, evalSokoHand(cards))
	assert.Greater(t, evalSokoHand(cards), SokoHandFourFlush)
}

func TestEvalSokoHand_WrongCardCount(t *testing.T) {
	assert.Equal(t, SokoHandHighCard, evalSokoHand([]*Card{sp(2)}))
	assert.Equal(t, SokoHandHighCard, evalSokoHand(nil))
}

func TestSokoHandName(t *testing.T) {
	assert.Equal(t, "Four-Card Flush", sokoHandName(SokoHandFourFlush))
	assert.Equal(t, "Four-Card Straight", sokoHandName(SokoHandFourStraight))
	assert.Equal(t, "Two Pair", sokoHandName(SokoHandTwoPair))
	assert.Equal(t, "Unknown", sokoHandName(-1))
	assert.Equal(t, "Unknown", sokoHandName(999))
}

// Every Soko rank must have a name: a positional slice that is one short shows
// up as "Unknown" on a real hand rather than as a compile error.
func TestSokoHandNames_CoverEveryRank(t *testing.T) {
	for r := SokoHandHighCard; r <= SokoHandRoyalFlush; r++ {
		assert.NotEqual(t, "Unknown", sokoHandName(r), "rank %d has no name", r)
	}
}
