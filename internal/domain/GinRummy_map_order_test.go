//go:build test

package domain

import "testing"

func TestGinRummyFindAllPossibleMeldsMapOrder(t *testing.T) {
	cards := []*Card{NewCard(1, 9, false), NewCard(2, 9, false), NewCard(3, 9, false), NewCard(1, 4, false), NewCard(2, 4, false), NewCard(3, 4, false)}
	for i := 0; i < 50; i++ {
		got := findAllPossibleMelds(cards)
		if len(got) != 2 || got[0][0].GetValue() != 4 {
			t.Fatalf("iteration %d: first candidate %v", i, got)
		}
	}
}
