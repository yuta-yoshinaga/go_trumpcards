//go:build test

package domain

import (
	"encoding/json"
	"testing"
)

func TestSkatPlayerAccessors(t *testing.T) {
	p := NewSkatPlayer(true)
	if !p.GetIsHuman() {
		t.Fatal("expected human")
	}
	if p.GetBid() != -1 {
		t.Fatalf("default bid = %d, want -1", p.GetBid())
	}
	p.SetBid(20)
	if p.GetBid() != 20 {
		t.Fatalf("bid = %d, want 20", p.GetBid())
	}
	p.SetIsDeclarer(true)
	if !p.GetIsDeclarer() {
		t.Fatal("declarer not set")
	}
	p.SetCardPoints(45)
	if p.GetCardPoints() != 45 {
		t.Fatalf("points = %d, want 45", p.GetCardPoints())
	}
	p.IncRoundsWon()
	if p.GetRoundsWon() != 1 {
		t.Fatalf("rounds won = %d, want 1", p.GetRoundsWon())
	}
	p.SetRoundsWon(5)
	if p.GetRoundsWon() != 5 {
		t.Fatalf("rounds won = %d, want 5", p.GetRoundsWon())
	}
	p.IncRoundsLost()
	if p.GetRoundsLost() != 1 {
		t.Fatalf("rounds lost = %d, want 1", p.GetRoundsLost())
	}
	p.SetRoundsLost(3)
	if p.GetRoundsLost() != 3 {
		t.Fatalf("rounds lost = %d, want 3", p.GetRoundsLost())
	}
}

func TestSkatPlayerResetRound(t *testing.T) {
	p := NewSkatPlayer(false)
	p.SetBid(20)
	p.SetIsDeclarer(true)
	p.SetCardPoints(50)
	p.SetRoundScore(40)
	p.AddCard(NewCard(CardDesignSpade, skatValueAce, false))
	p.AddTrick([]*Card{NewCard(CardDesignSpade, skatValueKing, false)})

	p.ResetRound()

	if p.GetBid() != -1 || p.GetIsDeclarer() || p.GetCardPoints() != 0 || p.GetCardsSize() != 0 || p.GetTrickCount() != 0 {
		t.Fatalf("ResetRound did not clear all per-round state: %+v", p)
	}
}

func TestSkatPlayerJSONRoundTrip(t *testing.T) {
	p := NewSkatPlayer(true)
	p.SetBid(24)
	p.SetIsDeclarer(true)
	p.SetCardPoints(61)
	p.SetRoundScore(36)
	p.SetCumulativeScore(120)
	p.IncRoundsWon()
	p.IncRoundsLost()
	p.AddCard(NewCard(CardDesignClover, skatValueJack, false))

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got SkatPlayer
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.GetBid() != 24 || !got.GetIsDeclarer() || got.GetCardPoints() != 61 ||
		got.GetRoundScore() != 36 || got.GetCumulativeScore() != 120 ||
		got.GetRoundsWon() != 1 || got.GetRoundsLost() != 1 {
		t.Fatalf("roundtrip mismatch: %+v", got)
	}
}
