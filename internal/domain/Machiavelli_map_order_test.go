//go:build test

package domain

import "testing"

func TestMachiavelliCPUFindRunsMapOrder(t *testing.T) {
	g := &Machiavelli{}
	hand := []*Card{NewCard(1, 5, false), NewCard(1, 6, false), NewCard(1, 7, false), NewCard(2, 2, false), NewCard(2, 3, false), NewCard(2, 4, false)}
	for i := 0; i < 50; i++ {
		got := g.cpuFindRuns(hand, make([]bool, len(hand)))
		if len(got) != 2 || hand[got[0][0]].GetDesign() != 1 {
			t.Fatalf("iteration %d: runs %v", i, got)
		}
	}
}
func TestMachiavelliCPUBuildPlayMapOrder(t *testing.T) {
	p := NewMachiavelliPlayer(false)
	for _, c := range []*Card{NewCard(1, 9, false), NewCard(2, 9, false), NewCard(3, 9, false), NewCard(1, 4, false), NewCard(2, 4, false), NewCard(3, 4, false)} {
		p.AddCard(c)
	}
	g := NewMachiavelli(nil, []*MachiavelliPlayer{p}, DefaultMachiavelliConfig())
	for i := 0; i < 50; i++ {
		_, added, ok := g.cpuBuildPlay()
		if !ok || len(added) < 3 {
			t.Fatalf("iteration %d: added %v", i, added)
		}
		hand := machiavelliCollectCards(p)
		if hand[added[0]].GetValue() != 4 {
			t.Fatalf("iteration %d: first rank %d", i, hand[added[0]].GetValue())
		}
	}
}
