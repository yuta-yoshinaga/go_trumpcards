//go:build test

package domain

import "testing"

func TestPokerCPUDecideExchangeMapOrder(t *testing.T) {
	p := NewPokerPlayer(false, PokerStyleConservative)
	for _, c := range []*Card{NewCard(1, 2, false), NewCard(2, 2, false), NewCard(1, 4, false), NewCard(2, 7, false), NewCard(3, 9, false)} {
		p.AddCard(c)
	}
	g := &Poker{players: []*PokerPlayer{p}}
	for i := 0; i < 50; i++ {
		got := g.cpuDecideExchange(0)
		if len(got) != 3 || got[0] != 2 || got[1] != 3 || got[2] != 4 {
			t.Fatalf("iteration %d: %v", i, got)
		}
	}
}

func TestPokerCPUDecideExchangeLowballMapOrder(t *testing.T) {
	p := NewPokerPlayer(false, PokerStyleConservative)
	for _, c := range []*Card{NewCard(1, 9, false), NewCard(2, 9, false), NewCard(1, 4, false), NewCard(2, 4, false), NewCard(3, 6, false)} {
		p.AddCard(c)
	}
	g := &Poker{players: []*PokerPlayer{p}}
	for i := 0; i < 50; i++ {
		got := g.cpuDecideExchangeLowball(0)
		if len(got) != 2 || got[0] != 3 || got[1] != 1 {
			t.Fatalf("iteration %d: %v, want [3 1]", i, got)
		}
	}
}
