//go:build test

package domain

import "testing"

func tc(playerIdx, design, value int) *TrickCard {
	return &TrickCard{PlayerIdx: playerIdx, Card: NewCard(design, value, false)}
}

func TestResolveTrickWinner_Empty(t *testing.T) {
	if got := ResolveTrickWinner(nil, -1, nil); got != 0 {
		t.Fatalf("empty trick: want 0, got %d", got)
	}
}

func TestResolveTrickWinner_NoTrump(t *testing.T) {
	// Lead suit = Heart. Highest Heart wins; the higher off-suit Diamond loses.
	trick := []*TrickCard{
		tc(0, CardDesignHeart, 9),
		tc(1, CardDesignDiamond, 14),
		tc(2, CardDesignHeart, 11),
		tc(3, CardDesignClover, 13),
	}
	if got := ResolveTrickWinner(trick, -1, nil); got != 2 {
		t.Fatalf("no-trump: want seat 2, got %d", got)
	}
}

func TestResolveTrickWinner_TrumpBeatsNonTrump(t *testing.T) {
	// Lead Heart, trump = Spade. A low Spade beats the high Heart.
	trick := []*TrickCard{
		tc(0, CardDesignHeart, 14),
		tc(1, CardDesignSpade, 2),
		tc(2, CardDesignHeart, 13),
	}
	if got := ResolveTrickWinner(trick, CardDesignSpade, nil); got != 1 {
		t.Fatalf("trump beats non-trump: want seat 1, got %d", got)
	}
}

func TestResolveTrickWinner_HighestTrumpWins(t *testing.T) {
	trick := []*TrickCard{
		tc(0, CardDesignHeart, 14),
		tc(1, CardDesignSpade, 5),
		tc(2, CardDesignSpade, 10),
	}
	if got := ResolveTrickWinner(trick, CardDesignSpade, nil); got != 2 {
		t.Fatalf("highest trump: want seat 2, got %d", got)
	}
}

func TestResolveTrickWinner_TieKeepsEarlierCard(t *testing.T) {
	// Two equal-rank lead-suit cards: the earlier-played seat keeps the trick.
	trick := []*TrickCard{
		tc(0, CardDesignHeart, 10),
		tc(1, CardDesignHeart, 10),
	}
	if got := ResolveTrickWinner(trick, CardDesignSpade, nil); got != 0 {
		t.Fatalf("tie: want earlier seat 0, got %d", got)
	}
}

func TestResolveTrickWinner_CustomRank(t *testing.T) {
	// Inverted rank: lower raw value is stronger. Lead-suit 3 beats lead-suit 14.
	invert := func(c *Card) int { return -c.GetValue() }
	trick := []*TrickCard{
		tc(0, CardDesignClover, 14),
		tc(1, CardDesignClover, 3),
	}
	if got := ResolveTrickWinner(trick, -1, invert); got != 1 {
		t.Fatalf("custom rank: want seat 1, got %d", got)
	}
}
