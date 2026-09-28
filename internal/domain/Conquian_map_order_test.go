//go:build test

package domain

import "testing"

func TestConquianFindMeldMapOrder(t *testing.T) {
	cards := []*Card{
		NewCard(CardDesignSpade, 9, false), NewCard(CardDesignHeart, 9, false), NewCard(CardDesignDiamond, 9, false),
		NewCard(CardDesignSpade, 4, false), NewCard(CardDesignHeart, 4, false), NewCard(CardDesignDiamond, 4, false),
	}
	for i := 0; i < 50; i++ {
		got := conquianFindMeld(cards)
		if len(got) != 3 || got[0].GetValue() != 4 {
			t.Fatalf("iteration %d: meld = %v, want rank 4 set", i, got)
		}
	}
}
