//go:build test

package domain

import "testing"

func TestContractRummyFindExtraMeldMapOrder(t *testing.T) {
	cards := []*Card{NewCard(CardDesignSpade, 9, false), NewCard(CardDesignHeart, 9, false), NewCard(CardDesignDiamond, 9, false), NewCard(CardDesignSpade, 4, false), NewCard(CardDesignHeart, 4, false), NewCard(CardDesignDiamond, 4, false)}
	for i := 0; i < 50; i++ {
		got := findExtraMeld(cards)
		if len(got) != 3 || got[0].GetValue() != 4 {
			t.Fatalf("iteration %d: got %v, want rank 4", i, got)
		}
	}
}
func TestContractRummyFindSetCandidatesMapOrder(t *testing.T) {
	cards := []*Card{NewCard(1, 9, false), NewCard(2, 9, false), NewCard(3, 9, false), NewCard(1, 4, false), NewCard(2, 4, false), NewCard(3, 4, false)}
	for i := 0; i < 50; i++ {
		got := findSetCandidates(3, cards, make([]bool, len(cards)))
		if len(got) != 2 || cards[got[0][0]].GetValue() != 4 {
			t.Fatalf("iteration %d: candidates %v, want rank 4 first", i, got)
		}
	}
}
