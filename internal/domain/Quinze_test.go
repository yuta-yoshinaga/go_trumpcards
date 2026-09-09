//go:build test

package domain

import "testing"

func quinzeTestCard(value int) *Card {
	return NewCard(CardDesignSpade, value, true)
}

func quinzeTestHand(bet int, values ...int) *QuinzeHand {
	cards := make([]*Card, 0, len(values))
	for _, value := range values {
		cards = append(cards, quinzeTestCard(value))
	}
	return &QuinzeHand{cards: cards, bet: bet}
}

func quinzeTestGame() *Quinze {
	game := NewDefaultQuinze()
	game.Reset()
	return game
}

func TestQuinzeResetUsesTheExpectedDefaults(t *testing.T) {
	game := quinzeTestGame()
	if game.GetPhase() != QuinzePhaseBet {
		t.Fatalf("phase = %d, want %d", game.GetPhase(), QuinzePhaseBet)
	}
	if game.GetChips() != QuinzeDefaultChips {
		t.Fatalf("chips = %d, want %d", game.GetChips(), QuinzeDefaultChips)
	}
	if len(game.GetSeats()) != QuinzeSeatCnt {
		t.Fatalf("seats = %d, want %d", len(game.GetSeats()), QuinzeSeatCnt)
	}
}

func TestQuinzeDeckHas52Cards(t *testing.T) {
	deck := NewTrumpCards(0)
	total := 0
	for deck.DrawCard() != nil {
		total++
	}
	if total != QuinzeDeckSize {
		t.Fatalf("deck = %d cards, want %d", total, QuinzeDeckSize)
	}
}

func TestQuinzeCardValues(t *testing.T) {
	tests := []struct {
		name  string
		value int
		want  int
	}{
		{name: "ace is one", value: 1, want: 1},
		{name: "number cards keep their value", value: 7, want: 7},
		{name: "ten is ten", value: 10, want: 10},
		{name: "jack is ten", value: 11, want: 10},
		{name: "queen is ten", value: 12, want: 10},
		{name: "king is ten", value: 13, want: 10},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := quinzeCardPoints(quinzeTestCard(test.value)); got != test.want {
				t.Fatalf("card value = %d, want %d", got, test.want)
			}
		})
	}
}

func TestQuinzeTargetAndBust(t *testing.T) {
	game := quinzeTestGame()
	exact := quinzeTestHand(10, 10, 5)
	if got := game.handPoints(exact); got != QuinzeTarget {
		t.Fatalf("exact total = %d, want %d", got, QuinzeTarget)
	}
	if got := game.settleHand(exact, 14, false); got != exact.bet {
		t.Fatalf("exact target payout = %d, want %d", got, exact.bet)
	}

	bust := quinzeTestHand(10, 10, 6)
	if got := game.handPoints(bust); got != 16 {
		t.Fatalf("bust total = %d, want 16", got)
	}
	if got := game.settleHand(bust, 14, false); got != -bust.bet {
		t.Fatalf("bust payout = %d, want %d", got, -bust.bet)
	}
}

func TestQuinzeTieGoesToTheBanker(t *testing.T) {
	game := quinzeTestGame()
	hand := quinzeTestHand(25, 10, 5)
	if got := game.settleHand(hand, QuinzeTarget, false); got != -hand.bet {
		t.Fatalf("tie payout = %d, want %d", got, -hand.bet)
	}
}

func TestQuinzeHumanCanStandWithAnExplicitHand(t *testing.T) {
	game := quinzeTestGame()
	game.banker = 1
	game.seats[0].hand = quinzeTestHand(100, 9)
	game.seats[2].hand = quinzeTestHand(20, 2)
	game.bankerHand = quinzeTestHand(0, 2)
	game.phase = QuinzePhasePlayerTurn
	game.activeSeat = 0

	if !game.CanStand() {
		t.Fatal("the human should be able to stand")
	}
	if err := game.Stand(); err != nil {
		t.Fatalf("Stand: %v", err)
	}
	if !game.seats[0].hand.stood {
		t.Fatal("the human hand should be stood")
	}
}
