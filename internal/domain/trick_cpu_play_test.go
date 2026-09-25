//go:build test

package domain

import "testing"

func TestTrickCpuPlaySkipsHuman(t *testing.T) {
	player := NewGamePlayer(true)
	player.AddCard(NewCard(CardDesignSpade, 3, false))
	chooseCalled, playCalled := false, false
	trickCpuPlay(0, player, func(int) int {
		chooseCalled = true
		return 0
	}, func(int, *Card) { playCalled = true })
	if chooseCalled || playCalled {
		t.Fatalf("human turn invoked selector=%v play=%v", chooseCalled, playCalled)
	}
	if player.GetCardsSize() != 1 {
		t.Fatalf("hand size = %d, want 1", player.GetCardsSize())
	}
}

func TestTrickCpuPlaySkipsEmptyHand(t *testing.T) {
	player := NewGamePlayer(false)
	playCalled := false
	trickCpuPlay(0, player, func(int) int { return 0 }, func(int, *Card) { playCalled = true })
	if playCalled {
		t.Fatal("play was called for an empty hand")
	}
	if player.GetCardsSize() != 0 {
		t.Fatalf("hand size = %d, want 0", player.GetCardsSize())
	}
}

func TestTrickCpuPlayRemovesAndPlaysChosenCard(t *testing.T) {
	player := NewGamePlayer(false)
	first := NewCard(CardDesignSpade, 3, false)
	chosen := NewCard(CardDesignHeart, 8, false)
	player.AddCard(first)
	player.AddCard(chosen)
	played := false
	trickCpuPlay(2, player, func(seat int) int {
		if seat != 2 {
			t.Fatalf("choose seat = %d, want 2", seat)
		}
		return 1
	}, func(seat int, card *Card) {
		played = true
		if seat != 2 {
			t.Errorf("play seat = %d, want 2", seat)
		}
		if card != chosen {
			t.Errorf("played card = %p, want %p", card, chosen)
		}
	})
	if !played {
		t.Fatal("play was not called")
	}
	if player.GetCardsSize() != 1 || player.GetCard(0) != first {
		t.Fatal("chosen card was not removed from hand")
	}
}
