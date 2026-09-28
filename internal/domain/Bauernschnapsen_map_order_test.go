//go:build test

package domain

import "testing"

func TestBauernschnapsenPickDefaultTrumpMapOrder(t *testing.T) {
	p := NewBauernschnapsenPlayer(false, 0)
	for _, c := range []*Card{NewCard(CardDesignHeart, 3, false), NewCard(CardDesignHeart, 4, false), NewCard(CardDesignSpade, 5, false), NewCard(CardDesignSpade, 6, false)} {
		p.AddCard(c)
	}
	g := &Bauernschnapsen{players: []*BauernschnapsenPlayer{p}}
	for i := 0; i < 50; i++ {
		if got := g.pickDefaultTrump(0); got != CardDesignSpade {
			t.Fatalf("iteration %d: got %d", i, got)
		}
	}
}
