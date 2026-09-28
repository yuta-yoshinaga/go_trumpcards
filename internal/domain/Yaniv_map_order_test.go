//go:build test

package domain

import "testing"

func TestBestYanivDiscardMapOrder(t *testing.T) {
	cards := []*Card{NewCard(CardDesignSpade, 2, false), NewCard(CardDesignSpade, 3, false), NewCard(CardDesignSpade, 4, false), NewCard(CardDesignHeart, 2, false), NewCard(CardDesignHeart, 3, false), NewCard(CardDesignHeart, 4, false)}
	for i := 0; i < 50; i++ {
		got := bestYanivDiscard(cards)
		if len(got) != 3 || got[0] != 0 || got[1] != 1 || got[2] != 2 {
			t.Fatalf("iteration %d: %v, want lowest suit run [0 1 2]", i, got)
		}
	}
}

func TestBestYanivDiscardSameValueTieMapOrder(t *testing.T) {
	cards := []*Card{
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignHeart, 11, false),
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignHeart, 12, false),
	}
	for i := 0; i < 50; i++ {
		got := bestYanivDiscard(cards)
		if len(got) != 2 || got[0] != 0 || got[1] != 1 {
			t.Fatalf("iteration %d: %v, want J pair [0 1] before tied Q pair", i, got)
		}
	}
}
