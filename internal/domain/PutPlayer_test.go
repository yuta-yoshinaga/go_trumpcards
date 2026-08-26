//go:build !js || !wasm || extra4

package domain

import (
	"encoding/json"
	"testing"
)

func TestPutPlayerResetGame(t *testing.T) {
	p := NewPutPlayer(true)
	p.AddCard(NewCard(CardDesignSpade, 1, false))
	p.AddTrick([]*Card{NewCard(CardDesignHeart, 4, false)})
	p.SetIsFinished(true)

	p.ResetGame()

	if p.GetCardsSize() != 0 {
		t.Errorf("hand size after reset = %d, want 0", p.GetCardsSize())
	}
	if p.GetTrickCount() != 0 {
		t.Errorf("trick count after reset = %d, want 0", p.GetTrickCount())
	}
	if p.GetIsFinished() {
		t.Error("isFinished after reset = true, want false")
	}
}

func TestPutPlayerJSONRoundTrip(t *testing.T) {
	p := NewPutPlayer(false)
	p.AddCard(NewCard(CardDesignSpade, 7, false))
	p.AddTrick([]*Card{NewCard(CardDesignDiamond, 7, false)})

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got PutPlayer
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.GetCardsSize() != 1 || got.GetCard(0).GetValue() != 7 {
		t.Errorf("hand not restored: %+v", got)
	}
	if got.GetTrickCount() != 1 {
		t.Errorf("trick count = %d, want 1", got.GetTrickCount())
	}
	if got.GetIsHuman() {
		t.Error("isHuman = true, want false")
	}
}

func TestPutPlayerUnmarshalNilGamePlayer(t *testing.T) {
	var got PutPlayer
	if err := json.Unmarshal([]byte(`{}`), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.GamePlayer == nil {
		t.Error("GamePlayer should default to non-nil")
	}
}
