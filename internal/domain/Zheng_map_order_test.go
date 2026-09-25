//go:build test

package domain

import "testing"

func TestZhengFindPairsMapOrder(t *testing.T) {
	p := NewZhengPlayer(false)
	for _, c := range []*Card{NewCard(1, 9, false), NewCard(2, 9, false), NewCard(1, 4, false), NewCard(2, 4, false)} {
		p.AddCard(c)
	}
	z := &Zheng{}
	for i := 0; i < 50; i++ {
		got := z.findPairs(p)
		if len(got) != 2 || p.GetCard(got[0][0]).GetValue() != 4 {
			t.Fatalf("iteration %d: %v", i, got)
		}
	}
}
