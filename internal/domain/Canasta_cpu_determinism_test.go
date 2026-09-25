//go:build test

package domain

import "testing"

func TestCanastaCPUFindMeldsDeterministicRankOrder(t *testing.T) {
	p := NewCanastaPlayer(false)
	for _, c := range []*Card{NewCard(1, 8, false), NewCard(2, 8, false), NewCard(1, 4, false), NewCard(2, 4, false), NewCard(3, 6, false), NewCard(4, 6, false), NewCard(1, 6, false), NewCard(CardDesignSpade, 2, false)} {
		p.AddCard(c)
	}
	g := &Canasta{}
	var baseline [][]*Card
	for i := 0; i < 50; i++ {
		got := g.cpuFindMelds(p)
		if len(got) != 2 || got[0][0].GetValue() != 4 || got[1][0].GetValue() != 6 || got[0][2].GetValue() != 2 {
			t.Fatalf("iteration %d: unexpected groups: %#v", i, got)
		}
		if baseline != nil {
			for j := range got {
				for k := range got[j] {
					if got[j][k] != baseline[j][k] {
						t.Fatalf("iteration %d: group order/cards changed", i)
					}
				}
			}
		} else {
			baseline = got
		}
	}
}

func TestCanastaCPUFindBiribaMeldsDeterministicSuitOrder(t *testing.T) {
	p := NewCanastaPlayer(false)
	for _, c := range []*Card{NewCard(CardDesignHeart, 4, false), NewCard(CardDesignHeart, 5, false), NewCard(CardDesignHeart, 6, false), NewCard(CardDesignSpade, 4, false), NewCard(CardDesignSpade, 5, false), NewCard(CardDesignSpade, 6, false)} {
		p.AddCard(c)
	}
	g := &Canasta{config: CanastaConfig{UseBiriba: true}}
	var baseline [][]*Card
	for i := 0; i < 50; i++ {
		got := g.cpuFindMelds(p)
		if len(got) != 2 || got[0][0].GetDesign() != CardDesignSpade || got[1][0].GetDesign() != CardDesignHeart {
			t.Fatalf("iteration %d: expected spade then heart runs, got %#v", i, got)
		}
		if baseline != nil {
			for j := range got {
				for k := range got[j] {
					if got[j][k] != baseline[j][k] {
						t.Fatalf("iteration %d: meld order/cards changed", i)
					}
				}
			}
		} else {
			baseline = got
		}
	}
}
