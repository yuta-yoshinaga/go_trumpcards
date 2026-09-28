//go:build test

package domain

import "testing"

func TestSevenBridgeFindBestMeldIndicesMapOrder(t *testing.T) {
	p := NewSevenBridgePlayer(true)
	for _, c := range []*Card{NewCard(CardDesignSpade, 5, false), NewCard(CardDesignHeart, 5, false), NewCard(CardDesignClover, 5, false), NewCard(CardDesignSpade, 2, false), NewCard(CardDesignHeart, 2, false), NewCard(CardDesignClover, 2, false)} {
		p.AddCard(c)
	}
	g := NewDefaultSevenBridge()
	for i := 0; i < 50; i++ {
		got, ok := g.findBestMeldIndices(p)
		if !ok || p.GetCard(got[0]).GetValue() != 2 {
			t.Fatalf("iteration %d: indices=%v, want rank 2", i, got)
		}
	}
}
