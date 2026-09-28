//go:build test

package domain

import "testing"

func TestBoliviaCPUFindMeldsDeterministicRankOrder(t *testing.T) {
	p := NewBoliviaPlayer(false, 0)
	for _, c := range []*Card{NewCard(1, 8, false), NewCard(2, 8, false), NewCard(1, 4, false), NewCard(2, 4, false), NewCard(3, 6, false), NewCard(4, 6, false), NewCard(1, 6, false), NewCard(CardDesignSpade, 2, false)} {
		p.AddCard(c)
	}
	g := &Bolivia{}
	var baseline []boliviaCpuGroup
	for i := 0; i < 50; i++ {
		got := g.cpuFindMelds(p)
		if len(got) != 2 || got[0].cards[0].GetValue() != 4 || got[1].cards[0].GetValue() != 6 || got[0].cards[2].GetValue() != 2 {
			t.Fatalf("iteration %d: unexpected groups", i)
		}
		if baseline != nil {
			for j := range got {
				for k := range got[j].cards {
					if got[j].cards[k] != baseline[j].cards[k] {
						t.Fatalf("iteration %d: group order/cards changed", i)
					}
				}
			}
		} else {
			baseline = got
		}
	}
}

func TestBoliviaCPUFindEscaleraMeldsDeterministicSuitOrder(t *testing.T) {
	p := NewBoliviaPlayer(false, 0)
	for _, c := range []*Card{NewCard(CardDesignHeart, 4, false), NewCard(CardDesignHeart, 5, false), NewCard(CardDesignHeart, 6, false), NewCard(CardDesignSpade, 4, false), NewCard(CardDesignSpade, 5, false), NewCard(CardDesignSpade, 6, false)} {
		p.AddCard(c)
	}
	got := (&Bolivia{}).cpuFindMelds(p)
	if len(got) != 2 || got[0].cards[0].GetDesign() != CardDesignSpade || got[1].cards[0].GetDesign() != CardDesignHeart {
		t.Fatalf("expected spade then heart runs, got %#v", got)
	}
	for i := 0; i < 50; i++ {
		repeated := (&Bolivia{}).cpuFindMelds(p)
		if len(repeated) != len(got) {
			t.Fatalf("iteration %d: groups changed", i)
		}
		for j := range got {
			for k := range got[j].cards {
				if repeated[j].cards[k] != got[j].cards[k] {
					t.Fatalf("iteration %d: group order/cards changed", i)
				}
			}
		}
	}
}
