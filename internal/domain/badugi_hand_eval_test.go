//go:build test

package domain

import (
	"strings"
	"testing"
)

// cardList is a compact helper for spelling out hands in test tables.
// Each pair is (design, value); design uses CardDesignSpade/Clover/Heart/Diamond.
func badugiCards(pairs ...[2]int) []*Card {
	out := make([]*Card, len(pairs))
	for i, p := range pairs {
		out[i] = NewCard(p[0], p[1], true)
	}
	return out
}

func TestEvalBadugiHand(t *testing.T) {
	S, C, H, D := CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond
	tests := []struct {
		name     string
		hand     []*Card
		wantSize int
		wantTop  int // highest value in the selected subset (A=1)
	}{
		{
			name:     "perfect badugi A-2-3-4 four suits",
			hand:     badugiCards([2]int{S, 1}, [2]int{H, 2}, [2]int{D, 3}, [2]int{C, 4}),
			wantSize: 4,
			wantTop:  4,
		},
		{
			name:     "perfect badugi 2-3-4-5",
			hand:     badugiCards([2]int{S, 2}, [2]int{H, 3}, [2]int{D, 4}, [2]int{C, 5}),
			wantSize: 4,
			wantTop:  5,
		},
		{
			name:     "three-card with rank duplicate drops pair",
			hand:     badugiCards([2]int{S, 1}, [2]int{H, 1}, [2]int{D, 3}, [2]int{C, 4}),
			wantSize: 3,
			wantTop:  4,
		},
		{
			name:     "three-card with suit duplicate drops higher card",
			hand:     badugiCards([2]int{S, 1}, [2]int{S, 5}, [2]int{D, 3}, [2]int{C, 4}),
			wantSize: 3,
			wantTop:  4,
		},
		{
			name:     "two-card alternating suits",
			hand:     badugiCards([2]int{S, 1}, [2]int{H, 2}, [2]int{S, 3}, [2]int{H, 4}),
			wantSize: 2,
			wantTop:  2,
		},
		{
			name:     "one-card all same suit picks lowest",
			hand:     badugiCards([2]int{S, 1}, [2]int{S, 2}, [2]int{S, 3}, [2]int{S, 4}),
			wantSize: 1,
			wantTop:  1,
		},
		{
			name:     "one-card all same rank picks lowest (here A)",
			hand:     badugiCards([2]int{S, 1}, [2]int{C, 1}, [2]int{H, 1}, [2]int{D, 1}),
			wantSize: 1,
			wantTop:  1,
		},
		{
			name:     "three-card with King high",
			hand:     badugiCards([2]int{S, 1}, [2]int{H, 2}, [2]int{D, 13}, [2]int{C, 2}),
			wantSize: 3,
			wantTop:  13,
		},
		{
			name: "tie-break prefers lower top card within same size",
			// hand has two 3-card subsets after dropping 10♠ or 5♣:
			//   {A♠,2♥,10♠} invalid (two spades)
			//   {A♠,2♥,5♣} valid, top=5
			//   {A♠,10♠,5♣} invalid (two spades)
			//   {2♥,10♠,5♣} valid, top=10
			// best should be the top=5 subset
			hand:     badugiCards([2]int{S, 1}, [2]int{H, 2}, [2]int{S, 10}, [2]int{C, 5}),
			wantSize: 3,
			wantTop:  5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := evalBadugiHand(tt.hand)
			if got.Size != tt.wantSize {
				t.Fatalf("size = %d, want %d (got cards=%v)", got.Size, tt.wantSize, badugiDebug(got.Cards))
			}
			if got.Size == 0 {
				return
			}
			// Cards sorted descending; first is the top card.
			if got.Cards[0].GetValue() != tt.wantTop {
				t.Fatalf("top card = %d, want %d (got cards=%v)", got.Cards[0].GetValue(), tt.wantTop, badugiDebug(got.Cards))
			}
		})
	}
}

func TestEvalBadugiHandInvalidInput(t *testing.T) {
	// Too few / too many cards: returns empty hand (Size 0) without panic.
	if got := evalBadugiHand(nil); got.Size != 0 {
		t.Errorf("nil hand Size = %d, want 0", got.Size)
	}
	if got := evalBadugiHand(badugiCards([2]int{CardDesignSpade, 1})); got.Size != 0 {
		t.Errorf("1-card hand Size = %d, want 0", got.Size)
	}
	big := badugiCards(
		[2]int{CardDesignSpade, 1}, [2]int{CardDesignHeart, 2},
		[2]int{CardDesignDiamond, 3}, [2]int{CardDesignClover, 4},
		[2]int{CardDesignSpade, 5},
	)
	if got := evalBadugiHand(big); got.Size != 0 {
		t.Errorf("5-card hand Size = %d, want 0", got.Size)
	}
}

func TestCompareBadugiHands(t *testing.T) {
	S, C, H, D := CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond

	perfect := evalBadugiHand(badugiCards([2]int{S, 1}, [2]int{H, 2}, [2]int{D, 3}, [2]int{C, 4}))
	sevenHigh := evalBadugiHand(badugiCards([2]int{S, 1}, [2]int{H, 2}, [2]int{D, 3}, [2]int{C, 7}))
	threeCard := evalBadugiHand(badugiCards([2]int{S, 1}, [2]int{H, 1}, [2]int{D, 3}, [2]int{C, 4}))
	twoCard := evalBadugiHand(badugiCards([2]int{S, 1}, [2]int{H, 2}, [2]int{S, 3}, [2]int{H, 4}))
	oneCard := evalBadugiHand(badugiCards([2]int{S, 1}, [2]int{S, 2}, [2]int{S, 3}, [2]int{S, 4}))

	tests := []struct {
		name string
		a, b BadugiHand
		want int
	}{
		{"4-card beats 3-card", perfect, threeCard, -1},
		{"3-card beats 2-card", threeCard, twoCard, -1},
		{"2-card beats 1-card", twoCard, oneCard, -1},
		{"4-card perfect beats 4-card seven-high", perfect, sevenHigh, -1},
		{"same 4-card hand is tie", perfect, perfect, 0},
		{"reflected comparison", sevenHigh, perfect, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compareBadugiHands(tt.a, tt.b); got != tt.want {
				t.Fatalf("compareBadugiHands = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCompareBadugiHandsTieBreakChain(t *testing.T) {
	// Two 3-card hands with same top card but different second card.
	// A-2-5 should beat A-4-5 because 2<4.
	S, C, H, D := CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond
	a := evalBadugiHand(badugiCards([2]int{S, 1}, [2]int{H, 2}, [2]int{D, 5}, [2]int{C, 2}))
	b := evalBadugiHand(badugiCards([2]int{S, 1}, [2]int{H, 4}, [2]int{D, 5}, [2]int{C, 4}))
	if a.Size != 3 || b.Size != 3 {
		t.Fatalf("precondition: a.Size=%d b.Size=%d want both 3", a.Size, b.Size)
	}
	if got := compareBadugiHands(a, b); got != -1 {
		t.Errorf("A-2-5 vs A-4-5: compareBadugiHands = %d, want -1", got)
	}
}

// badugiDebug renders cards in a short form for test failure messages.
func badugiDebug(cards []*Card) string {
	if len(cards) == 0 {
		return "∅"
	}
	var b strings.Builder
	for _, c := range cards {
		suit := "?"
		switch c.GetDesign() {
		case CardDesignSpade:
			suit = "S"
		case CardDesignClover:
			suit = "C"
		case CardDesignHeart:
			suit = "H"
		case CardDesignDiamond:
			suit = "D"
		}
		b.WriteString(suit)
		b.WriteString(itoaBadugi(c.GetValue()))
		b.WriteString(" ")
	}
	return b.String()
}

func itoaBadugi(v int) string {
	switch v {
	case 1:
		return "A"
	case 11:
		return "J"
	case 12:
		return "Q"
	case 13:
		return "K"
	}
	if v < 10 {
		return string(rune('0' + v))
	}
	return "10"
}
