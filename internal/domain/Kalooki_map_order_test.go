//go:build test

package domain

import "testing"

func TestFindKalookiMeldMapOrder(t *testing.T) {
	cards := []*Card{NewCard(CardDesignSpade, 5, false), NewCard(CardDesignHeart, 5, false), NewCard(CardDesignClover, 5, false), NewCard(CardDesignSpade, 2, false), NewCard(CardDesignHeart, 2, false), NewCard(CardDesignClover, 2, false)}
	for i := 0; i < 50; i++ {
		got := findKalookiMeld(cards)
		if len(got) != 3 || got[0].GetValue() != 2 {
			t.Fatalf("iteration %d: meld=%v, want rank 2", i, got)
		}
	}
}
