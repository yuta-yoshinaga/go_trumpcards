package domain

import (
	"errors"
	"testing"
)

func TestTrickPlayerPlayRejectsNonHuman(t *testing.T) {
	player := NewGamePlayer(false)
	called := false
	err := trickPlayerPlay(0, player, 0, "test.outOfRange", func(int, *Card) error { called = true; return nil }, func(int, *Card) { called = true })
	if !errors.Is(err, ErrNotHumanTurn) {
		t.Fatalf("got %v, want ErrNotHumanTurn", err)
	}
	if called {
		t.Fatal("validation or play called for non-human")
	}
}

func TestTrickPlayerPlayRejectsOutOfRange(t *testing.T) {
	for _, index := range []int{-1, 1} {
		t.Run(map[bool]string{true: "negative", false: "at_size"}[index < 0], func(t *testing.T) {
			player := NewGamePlayer(true)
			player.AddCard(NewCard(CardDesignSpade, 1, false))
			err := trickPlayerPlay(0, player, index, "test.outOfRange", func(int, *Card) error { return nil }, func(int, *Card) {})
			if !errors.Is(err, ErrInvalidCard) {
				t.Fatalf("got %v, want ErrInvalidCard", err)
			}
			var domainErr *DomainError
			if !errors.As(err, &domainErr) || domainErr.Code != "test.outOfRange" {
				t.Fatalf("got error code %v, want test.outOfRange", err)
			}
			if player.GetCardsSize() != 1 {
				t.Fatalf("hand size = %d, want 1", player.GetCardsSize())
			}
		})
	}
}

func TestTrickPlayerPlayReturnsValidationErrorWithoutRemovingCard(t *testing.T) {
	player := NewGamePlayer(true)
	card := NewCard(CardDesignHeart, 7, false)
	player.AddCard(card)
	want := errors.New("invalid play")
	got := trickPlayerPlay(2, player, 0, "test.outOfRange", func(seat int, gotCard *Card) error {
		if seat != 2 || gotCard != card {
			t.Fatalf("validate got seat=%d card=%p", seat, gotCard)
		}
		return want
	}, func(int, *Card) { t.Fatal("play called after validation error") })
	if got != want {
		t.Fatalf("got %v, want original error", got)
	}
	if player.GetCardsSize() != 1 {
		t.Fatalf("hand size = %d, want 1", player.GetCardsSize())
	}
}

func TestTrickPlayerPlayRemovesAndPlaysSelectedCard(t *testing.T) {
	player := NewGamePlayer(true)
	first, selected := NewCard(CardDesignSpade, 1, false), NewCard(CardDesignHeart, 7, false)
	player.AddCard(first)
	player.AddCard(selected)
	var played *Card
	got := trickPlayerPlay(3, player, 1, "test.outOfRange", func(seat int, card *Card) error {
		if seat != 3 || card != selected {
			t.Fatalf("validate got seat=%d card=%p", seat, card)
		}
		return nil
	}, func(seat int, card *Card) {
		if seat != 3 {
			t.Fatalf("play seat = %d, want 3", seat)
		}
		played = card
	})
	if got != nil {
		t.Fatalf("got %v, want nil", got)
	}
	if played != selected {
		t.Fatalf("played %p, want selected %p", played, selected)
	}
	if player.GetCardsSize() != 1 || player.GetCard(0) != first {
		t.Fatal("selected card was not removed from hand")
	}
}
