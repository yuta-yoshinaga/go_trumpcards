package domain_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

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
