//go:build !js || !wasm || classic

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestBrusquembillePlayer_New(t *testing.T) {
	p := domain.NewBrusquembillePlayer(true)
	if !p.GetIsHuman() {
		t.Error("expected human player")
	}
	if p.GetTrickCount() != 0 {
		t.Errorf("GetTrickCount() = %d, want 0", p.GetTrickCount())
	}
	if p.GetCardsSize() != 0 {
		t.Errorf("GetCardsSize() = %d, want 0", p.GetCardsSize())
	}

	p2 := domain.NewBrusquembillePlayer(false)
	if p2.GetIsHuman() {
		t.Error("expected non-human player")
	}
}

func TestBrusquembillePlayer_ResetGame(t *testing.T) {
	p := domain.NewBrusquembillePlayer(true)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
	p.AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 11, false)})
	p.SetIsFinished(true)

	p.ResetGame()

	if p.GetCardsSize() != 0 {
		t.Errorf("GetCardsSize() after ResetGame = %d, want 0", p.GetCardsSize())
	}
	if p.GetTrickCount() != 0 {
		t.Errorf("GetTrickCount() after ResetGame = %d, want 0", p.GetTrickCount())
	}
	if p.GetIsFinished() {
		t.Error("expected isFinished=false after ResetGame")
	}
}

func TestBrusquembillePlayer_JSONRoundtrip(t *testing.T) {
	p := domain.NewBrusquembillePlayer(true)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	p.AddTrick([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 3, false),
		domain.NewCard(domain.CardDesignDiamond, 13, false),
	})

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got domain.BrusquembillePlayer
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !got.GetIsHuman() {
		t.Error("isHuman lost in roundtrip")
	}
	if got.GetCardsSize() != 1 {
		t.Errorf("cards size = %d, want 1", got.GetCardsSize())
	}
	if got.GetTrickCount() != 1 {
		t.Errorf("trick count = %d, want 1", got.GetTrickCount())
	}
}
