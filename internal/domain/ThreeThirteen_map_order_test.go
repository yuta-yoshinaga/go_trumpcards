//go:build test

package domain

import "testing"

func TestThreeThirteenAllMeldsMapOrder(t *testing.T) {
	cards := []*Card{NewCard(CardDesignSpade, 5, false), NewCard(CardDesignHeart, 5, false), NewCard(CardDesignClover, 5, false), NewCard(CardDesignSpade, 2, false), NewCard(CardDesignHeart, 2, false), NewCard(CardDesignClover, 2, false)}
	for i := 0; i < 50; i++ {
		got := threeThirteenAllMelds(cards, 13)
		if len(got) < 2 || got[0][0].GetValue() != 2 {
			t.Fatalf("iteration %d: first meld=%v, want rank 2", i, got[0])
		}
	}
}
