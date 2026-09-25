//go:build test

package domain

import "testing"

func TestCariocaMapOrder(t *testing.T) {
	t.Run("set candidates rank order", func(t *testing.T) {
		cards := []*Card{cariocaCard(1, 9), cariocaCard(2, 9), cariocaCard(3, 9), cariocaCard(1, 4), cariocaCard(2, 4), cariocaCard(3, 4)}
		for i := 0; i < 50; i++ {
			got := cariocaFindSetCandidates(3, cards, make([]bool, len(cards)))
			if len(got) == 0 || got[0][0] != 3 {
				t.Fatalf("iteration %d: first candidate = %v, want rank 4 first", i, got)
			}
		}
	})
	t.Run("run candidates suit order", func(t *testing.T) {
		cards := []*Card{cariocaCard(3, 4), cariocaCard(3, 5), cariocaCard(3, 6), cariocaCard(1, 4), cariocaCard(1, 5), cariocaCard(1, 6)}
		for i := 0; i < 50; i++ {
			got := cariocaFindRunCandidates(3, cards, make([]bool, len(cards)))
			if len(got) == 0 || got[0][0] != 3 {
				t.Fatalf("iteration %d: first candidate = %v, want lowest suit first", i, got)
			}
		}
	})
	t.Run("extra meld rank order", func(t *testing.T) {
		cards := []*Card{cariocaCard(1, 9), cariocaCard(2, 9), cariocaCard(3, 9), cariocaCard(1, 4), cariocaCard(2, 4), cariocaCard(3, 4)}
		for i := 0; i < 50; i++ {
			got, ok := cariocaFindExtraMeld(cards)
			if !ok || got[0].GetValue() != 4 {
				t.Fatalf("iteration %d: meld = %v, ok=%v; want rank 4 set", i, got, ok)
			}
		}
	})
}
