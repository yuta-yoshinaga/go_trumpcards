package domain

import "testing"

// testPlayer is a minimal cardAdder for testing dealAllCards.
type testPlayer struct {
	cards []*Card
}

func (p *testPlayer) AddCard(card *Card) {
	p.cards = append(p.cards, card)
}

func TestDealAllCards(t *testing.T) {
	t.Run("normal dealing distributes round-robin", func(t *testing.T) {
		deck := NewTrumpCards(0) // 52 cards
		players := make([]*testPlayer, 4)
		for i := range players {
			players[i] = &testPlayer{}
		}

		dealAllCards(deck, players)

		// 52 cards / 4 players = 13 each
		for i, p := range players {
			if len(p.cards) != 13 {
				t.Errorf("player %d got %d cards, want 13", i, len(p.cards))
			}
		}
	})

	t.Run("empty deck deals no cards", func(t *testing.T) {
		deck := NewTrumpCards(0)
		// Drain all cards
		for deck.DrawCard() != nil {
		}

		players := make([]*testPlayer, 3)
		for i := range players {
			players[i] = &testPlayer{}
		}

		dealAllCards(deck, players)

		for i, p := range players {
			if len(p.cards) != 0 {
				t.Errorf("player %d got %d cards, want 0", i, len(p.cards))
			}
		}
	})

	t.Run("no players does not panic", func(t *testing.T) {
		deck := NewTrumpCards(0) // 52 cards
		var players []*testPlayer

		dealAllCards(deck, players)

		if deck.GetRemainingCount() != 52 {
			t.Errorf("deck should not be drawn from when there are no players, remaining: %d", deck.GetRemainingCount())
		}
	})

	t.Run("nil deck does not panic", func(t *testing.T) {
		players := make([]*testPlayer, 2)
		for i := range players {
			players[i] = &testPlayer{}
		}

		dealAllCards(nil, players)

		for i, p := range players {
			if len(p.cards) != 0 {
				t.Errorf("player %d got %d cards, want 0", i, len(p.cards))
			}
		}
	})

	t.Run("uneven distribution", func(t *testing.T) {
		deck := NewTrumpCards(1) // 53 cards
		players := make([]*testPlayer, 4)
		for i := range players {
			players[i] = &testPlayer{}
		}

		dealAllCards(deck, players)

		// 53 cards / 4 players: first player gets 14, rest get 13
		if len(players[0].cards) != 14 {
			t.Errorf("player 0 got %d cards, want 14", len(players[0].cards))
		}
		for i := 1; i < 4; i++ {
			if len(players[i].cards) != 13 {
				t.Errorf("player %d got %d cards, want 13", i, len(players[i].cards))
			}
		}

		// Verify total
		total := 0
		for _, p := range players {
			total += len(p.cards)
		}
		if total != 53 {
			t.Errorf("total cards dealt = %d, want 53", total)
		}
	})
}
