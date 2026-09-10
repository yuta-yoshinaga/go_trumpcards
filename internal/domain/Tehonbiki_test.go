//go:build test

package domain

import "testing"

func TestTehonbikiPayoutsAndLosses(t *testing.T) {
	tests := []struct {
		name    string
		kind    TehonbikiBetType
		numbers []int
		payout  int
	}{
		{"single", TehonbikiBetSingle, []int{4}, 225},
		{"double", TehonbikiBetDouble, []int{2, 4}, 90},
		{"triple", TehonbikiBetTriple, []int{1, 4, 6}, 45},
		{"half", TehonbikiBetHalf, []int{3, 4, 5}, 45},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewDefaultTehonbiki()
			if err := g.SetParentCard(4); err != nil {
				t.Fatal(err)
			}
			if err := g.PlaceBet(tt.numbers, tt.kind, 50); err != nil {
				t.Fatal(err)
			}
			if g.GetResult() != TehonbikiResultWin || g.GetPayout() != tt.payout {
				t.Fatalf("result=%v payout=%d", g.GetResult(), g.GetPayout())
			}
		})
	}
	g := NewDefaultTehonbiki()
	if err := g.SetParentCard(6); err != nil {
		t.Fatal(err)
	}
	if err := g.PlaceBet([]int{1, 2}, TehonbikiBetDouble, 50); err != nil {
		t.Fatal(err)
	}
	if g.GetResult() != TehonbikiResultLose || g.GetPayout() != 0 || g.GetChips() != 950 {
		t.Fatalf("loss state: result=%v payout=%d chips=%d", g.GetResult(), g.GetPayout(), g.GetChips())
	}
}

func TestTehonbikiParentCardRange(t *testing.T) {
	g := NewDefaultTehonbiki()
	for _, n := range []int{0, 7} {
		if err := g.SetParentCard(n); err == nil {
			t.Fatalf("accepted parent card %d", n)
		}
	}
	for range 20 {
		g.Reset()
		if g.GetParentCard() < 1 || g.GetParentCard() > 6 {
			t.Fatalf("parent card=%d", g.GetParentCard())
		}
	}
}
