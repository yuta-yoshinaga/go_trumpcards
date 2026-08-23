package domain_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestNewTrumpCardsScopa(t *testing.T) {
	deck := domain.NewTrumpCardsScopa()
	if got := deck.GetTotalCount(); got != 40 {
		t.Fatalf("GetTotalCount() = %d, want 40", got)
	}

	valid := map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 7: true, 11: true, 12: true, 13: true}
	suits := map[int]int{}
	values := map[int]int{}
	for i := range 40 {
		c := deck.DrawCard()
		if c == nil {
			t.Fatalf("expected card at index %d but got nil", i)
		}
		if !valid[c.GetValue()] {
			t.Errorf("unexpected value %d at index %d", c.GetValue(), i)
		}
		if c.GetDesign() < domain.CardDesignSpade || c.GetDesign() > domain.CardDesignDiamond {
			t.Errorf("unexpected design %d at index %d", c.GetDesign(), i)
		}
		suits[c.GetDesign()]++
		values[c.GetValue()]++
	}
	for s := domain.CardDesignSpade; s <= domain.CardDesignDiamond; s++ {
		if suits[s] != 10 {
			t.Errorf("suit %d count = %d, want 10", s, suits[s])
		}
	}
	for v := range valid {
		if values[v] != 4 {
			t.Errorf("value %d count = %d, want 4", v, values[v])
		}
	}
	if extra := deck.DrawCard(); extra != nil {
		t.Errorf("expected nil after 40 draws, got %+v", extra)
	}
}

func TestNewTrumpCardsBriscola(t *testing.T) {
	deck := domain.NewTrumpCardsBriscola()
	if got := deck.GetTotalCount(); got != 40 {
		t.Fatalf("GetTotalCount() = %d, want 40", got)
	}

	valid := map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 7: true, 11: true, 12: true, 13: true}
	suits := map[int]int{}
	values := map[int]int{}
	for i := range 40 {
		c := deck.DrawCard()
		if c == nil {
			t.Fatalf("expected card at index %d but got nil", i)
		}
		if !valid[c.GetValue()] {
			t.Errorf("unexpected value %d at index %d", c.GetValue(), i)
		}
		if c.GetDesign() < domain.CardDesignSpade || c.GetDesign() > domain.CardDesignDiamond {
			t.Errorf("unexpected design %d at index %d", c.GetDesign(), i)
		}
		suits[c.GetDesign()]++
		values[c.GetValue()]++
	}
	for s := domain.CardDesignSpade; s <= domain.CardDesignDiamond; s++ {
		if suits[s] != 10 {
			t.Errorf("suit %d count = %d, want 10", s, suits[s])
		}
	}
	for v := range valid {
		if values[v] != 4 {
			t.Errorf("value %d count = %d, want 4", v, values[v])
		}
	}
	if extra := deck.DrawCard(); extra != nil {
		t.Errorf("expected nil after 40 draws, got %+v", extra)
	}
}

func TestTrumpCards_Method(t *testing.T) {
	t.Run("success Shuffle", func(t *testing.T) {
		tc := domain.NewTrumpCards(2)
		tc.Shuffle()
	})
	t.Run("success DrawCard", func(t *testing.T) {
		tc := domain.NewTrumpCards(2)
		card := tc.DrawCard()
		if card == nil {
			t.Fatal("expected a card but got nil")
		}
	})
}

func TestTrumpCards_cardsInit(t *testing.T) {
	t.Run("deck size with 0 jokers", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		// Draw all 52 cards
		for i := 0; i < 52; i++ {
			card := tc.DrawCard()
			if card == nil {
				t.Fatalf("expected card at index %d but got nil", i)
			}
		}
		// 53rd draw should return nil
		if card := tc.DrawCard(); card != nil {
			t.Fatal("expected nil after drawing all cards")
		}
	})

	t.Run("deck size with 2 jokers", func(t *testing.T) {
		tc := domain.NewTrumpCards(2)
		// Draw all 54 cards
		for i := 0; i < 54; i++ {
			card := tc.DrawCard()
			if card == nil {
				t.Fatalf("expected card at index %d but got nil", i)
			}
		}
		// 55th draw should return nil
		if card := tc.DrawCard(); card != nil {
			t.Fatal("expected nil after drawing all cards")
		}
	})

	t.Run("suit and value correctness", func(t *testing.T) {
		tc := domain.NewTrumpCards(2)

		// Expected suits in order: Spade, Clover, Heart, Diamond
		expectedDesigns := []int{
			domain.CardDesignSpade,
			domain.CardDesignClover,
			domain.CardDesignHeart,
			domain.CardDesignDiamond,
		}

		// Verify 4 suits x 13 cards
		for _, expectedDesign := range expectedDesigns {
			for val := 1; val <= domain.CardValueMax; val++ {
				card := tc.DrawCard()
				if card == nil {
					t.Fatalf("expected card with design=%d value=%d but got nil", expectedDesign, val)
				}
				if card.GetDesign() != expectedDesign {
					t.Errorf("expected design %d but got %d", expectedDesign, card.GetDesign())
				}
				if card.GetValue() != val {
					t.Errorf("expected value %d but got %d", val, card.GetValue())
				}
			}
		}

		// Verify 2 jokers
		for i := 1; i <= 2; i++ {
			card := tc.DrawCard()
			if card == nil {
				t.Fatalf("expected joker %d but got nil", i)
			}
			if card.GetDesign() != domain.CardDesignJoker {
				t.Errorf("expected joker design %d but got %d", domain.CardDesignJoker, card.GetDesign())
			}
			if card.GetValue() != i {
				t.Errorf("expected joker value %d but got %d", i, card.GetValue())
			}
		}
	})
}

func TestTrumpCards_DrawCard(t *testing.T) {
	t.Run("returns nil when deck is exhausted", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		for i := 0; i < 52; i++ {
			tc.DrawCard()
		}
		card := tc.DrawCard()
		if card != nil {
			t.Fatal("expected nil when deck is exhausted")
		}
	})

	t.Run("drawn card has draw flag set", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		card := tc.DrawCard()
		if card == nil {
			t.Fatal("expected a card but got nil")
		}
		if !card.GetDraw() {
			t.Error("expected draw flag to be true")
		}
	})
}

func TestTrumpCards_Replenish(t *testing.T) {
	t.Run("makes drawn cards available again without changing deck order", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)

		// Capture the initial deck order by drawing all cards.
		drawn := make([]*domain.Card, 0, 52)
		for {
			c := tc.DrawCard()
			if c == nil {
				break
			}
			drawn = append(drawn, c)
		}
		if tc.GetRemainingCount() != 0 {
			t.Fatalf("setup: expected exhausted deck, got %d remaining", tc.GetRemainingCount())
		}

		// Replenish should restore full availability without shuffling.
		tc.Replenish()
		if got := tc.GetRemainingCount(); got != 52 {
			t.Fatalf("expected 52 remaining after Replenish, got %d", got)
		}

		for i := 0; i < 52; i++ {
			c := tc.DrawCard()
			if c == nil {
				t.Fatalf("DrawCard returned nil at index %d after Replenish", i)
			}
			if c.GetDesign() != drawn[i].GetDesign() || c.GetValue() != drawn[i].GetValue() {
				t.Errorf("card %d differs from pre-Replenish order: got (%v,%d) want (%v,%d)",
					i, c.GetDesign(), c.GetValue(), drawn[i].GetDesign(), drawn[i].GetValue())
			}
		}
	})
}

func TestTrumpCards_GetRemainingCount(t *testing.T) {
	t.Run("initial remaining equals total", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		if tc.GetRemainingCount() != 52 {
			t.Errorf("expected 52 remaining, got %d", tc.GetRemainingCount())
		}
	})

	t.Run("remaining decreases after draw", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		tc.DrawCard()
		if tc.GetRemainingCount() != 51 {
			t.Errorf("expected 51 remaining after 1 draw, got %d", tc.GetRemainingCount())
		}
	})

	t.Run("remaining is 0 after all drawn", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		for i := 0; i < 52; i++ {
			tc.DrawCard()
		}
		if tc.GetRemainingCount() != 0 {
			t.Errorf("expected 0 remaining, got %d", tc.GetRemainingCount())
		}
	})

	t.Run("multi-deck remaining count", func(t *testing.T) {
		tc := domain.NewTrumpCardsWithDecks(2, 0)
		if tc.GetRemainingCount() != 104 {
			t.Errorf("expected 104 remaining for 2 decks, got %d", tc.GetRemainingCount())
		}
		for i := 0; i < 10; i++ {
			tc.DrawCard()
		}
		if tc.GetRemainingCount() != 94 {
			t.Errorf("expected 94 remaining after 10 draws, got %d", tc.GetRemainingCount())
		}
	})

	t.Run("remaining with jokers", func(t *testing.T) {
		tc := domain.NewTrumpCards(2)
		if tc.GetRemainingCount() != 54 {
			t.Errorf("expected 54 remaining with 2 jokers, got %d", tc.GetRemainingCount())
		}
	})
}

func TestTrumpCards_GetTotalCount(t *testing.T) {
	t.Run("single deck no jokers", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		if tc.GetTotalCount() != 52 {
			t.Errorf("expected total 52, got %d", tc.GetTotalCount())
		}
	})

	t.Run("single deck with 2 jokers", func(t *testing.T) {
		tc := domain.NewTrumpCards(2)
		if tc.GetTotalCount() != 54 {
			t.Errorf("expected total 54, got %d", tc.GetTotalCount())
		}
	})

	t.Run("multi-deck total count", func(t *testing.T) {
		tc := domain.NewTrumpCardsWithDecks(6, 0)
		if tc.GetTotalCount() != 312 {
			t.Errorf("expected total 312 for 6 decks, got %d", tc.GetTotalCount())
		}
	})

	t.Run("total count unchanged after draws", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		for i := 0; i < 20; i++ {
			tc.DrawCard()
		}
		if tc.GetTotalCount() != 52 {
			t.Errorf("expected total 52 unchanged after draws, got %d", tc.GetTotalCount())
		}
	})

	t.Run("total count reset after shuffle", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		for i := 0; i < 20; i++ {
			tc.DrawCard()
		}
		tc.Shuffle()
		if tc.GetTotalCount() != 52 {
			t.Errorf("expected total 52 after shuffle, got %d", tc.GetTotalCount())
		}
		if tc.GetRemainingCount() != 52 {
			t.Errorf("expected remaining 52 after shuffle, got %d", tc.GetRemainingCount())
		}
	})
}

// TestNewTrumpCardsTrappola は 36 枚が**4 スートに均等**であることを見る。
//
// issue #5423 は「36 枚は NewTrumpCardsWithSuits(36, suits) で構成できる」と
// 書いていたが、あの関数はスートごとに値 1..13 を回して指定枚数で打ち切るので
// 13+13+10+0 になり **ダイヤが 1 枚も入らない**。Skat が同じ形でダイヤ 0 枚の
// まま出荷された前例があるので、スート別の枚数を明示的に数える。
func TestNewTrumpCardsTrappola(t *testing.T) {
	deck := domain.NewTrumpCardsTrappola()
	if got := deck.GetTotalCount(); got != 36 {
		t.Fatalf("GetTotalCount() = %d, want 36", got)
	}

	perSuit := map[int]int{}
	perValue := map[int]int{}
	for i := 0; i < 36; i++ {
		c := deck.DrawCard()
		if c == nil {
			t.Fatalf("DrawCard() returned nil at %d", i)
		}
		perSuit[c.GetDesign()]++
		perValue[c.GetValue()]++
	}

	// **4 スートが同じ枚数。** ここが崩れるのが例の壊れ方。
	for suit := domain.CardDesignSpade; suit <= domain.CardDesignMax; suit++ {
		if got, want := perSuit[suit], len(domain.TrappolaValues); got != want {
			t.Errorf("suit %d has %d cards, want %d", suit, got, want)
		}
	}

	// 札位は A,3,4,5,6,7,J,Q,K のみ。2 と 8,9,10 は入らない。
	for _, v := range domain.TrappolaValues {
		if got := perValue[v]; got != 4 {
			t.Errorf("value %d appears %d times, want 4", v, got)
		}
	}
	for _, v := range []int{2, 8, 9, 10} {
		if got := perValue[v]; got != 0 {
			t.Errorf("value %d appears %d times, want 0", v, got)
		}
	}

	// **負のコントロール。** issue の推奨する作り方は実際に壊れる ——
	// これが通らなくなったら NewTrumpCardsWithSuits の挙動が変わったので、
	// 上のコメントごと見直すこと。
	suits := []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond}
	bad := domain.NewTrumpCardsWithSuits(36, suits)
	badDiamonds := 0
	for i := 0; i < 36; i++ {
		if c := bad.DrawCard(); c != nil && c.GetDesign() == domain.CardDesignDiamond {
			badDiamonds++
		}
	}
	if badDiamonds != 0 {
		t.Errorf("NewTrumpCardsWithSuits(36) now yields %d diamonds -- revisit the comment above", badDiamonds)
	}
}
