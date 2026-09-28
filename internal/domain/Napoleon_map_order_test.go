//go:build test

package domain

import "testing"

func TestNapoleonCPUSelectTrumpNormalMapOrder(t *testing.T) {
	p := NewNapoleonPlayer(false)
	for _, c := range []*Card{NewCard(CardDesignHeart, 3, false), NewCard(CardDesignHeart, 4, false), NewCard(CardDesignSpade, 5, false), NewCard(CardDesignSpade, 6, false)} {
		p.AddCard(c)
	}
	n := &Napoleon{players: []*NapoleonPlayer{p}}
	for i := 0; i < 50; i++ {
		trump, _, _ := n.cpuSelectTrumpNormal(0)
		if trump != CardDesignSpade {
			t.Fatalf("iteration %d: trump %d", i, trump)
		}
	}
}
