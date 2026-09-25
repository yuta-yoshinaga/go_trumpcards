//go:build test

package domain

import "testing"

func TestBuraCpuChooseLeadMapOrder(t *testing.T) {
	g := NewDefaultBura()
	g.trumpSuit = CardDesignDiamond
	hand := []*Card{
		NewCard(CardDesignHeart, 1, false), NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignSpade, 1, false), NewCard(CardDesignSpade, 2, false),
	}
	for i := 0; i < 50; i++ {
		got := g.cpuChooseLead(hand)
		if len(got) != 2 || got[0] != 2 || got[1] != 3 {
			t.Fatalf("iteration %d: lead indices = %v, want [2 3] (lowest suit on tie)", i, got)
		}
	}
}
