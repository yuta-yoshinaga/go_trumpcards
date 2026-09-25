//go:build test

package domain

import "testing"

func TestPresidentGroupByValueMapOrder(t *testing.T) {
	p := NewPresidentPlayer(false)
	for _, c := range []*Card{NewCard(1, 9, false), NewCard(2, 9, false), NewCard(1, 4, false), NewCard(2, 4, false)} {
		p.AddCard(c)
	}
	g := &President{}
	for i := 0; i < 50; i++ {
		groups := g.groupByValue(p)
		if len(groups) != 2 || p.GetCard(groups[0][0]).GetValue() != 4 {
			t.Fatalf("iteration %d: %v", i, groups)
		}
	}
}
