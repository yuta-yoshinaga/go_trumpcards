//go:build test

package domain

import "testing"

func TestHandAndFootCPUFindMeldsMapOrder(t *testing.T) {
	p := NewHandAndFootPlayer(false)
	for _, c := range []*Card{NewCard(1, 7, false), NewCard(2, 7, false), NewCard(0, 0, false), NewCard(1, 9, false), NewCard(2, 9, false), NewCard(0, 0, false)} {
		p.AddCard(c)
	}
	g := &HandAndFoot{}
	for i := 0; i < 50; i++ {
		got := g.cpuFindMelds(p, 0)
		if len(got) != 2 || got[0][0].GetValue() != 7 || got[0][2].GetDesign() != 0 || got[1][2].GetDesign() != 0 {
			t.Fatalf("iteration %d: melds %v", i, got)
		}
	}
}
