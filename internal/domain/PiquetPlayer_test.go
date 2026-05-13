//go:build test

package domain

import (
	"encoding/json"
	"testing"
)

func TestNewPiquetPlayer(t *testing.T) {
	p := NewPiquetPlayer(true)
	if !p.GetIsHuman() {
		t.Error("expected human player")
	}
	if p.GetDeclScore() != 0 || p.GetTrickScore() != 0 || p.GetBonusScore() != 0 {
		t.Errorf("expected zero scores, got decl=%d trick=%d bonus=%d",
			p.GetDeclScore(), p.GetTrickScore(), p.GetBonusScore())
	}
	if p.GetMatchScore() != 0 {
		t.Errorf("expected zero matchScore, got %d", p.GetMatchScore())
	}
}

func TestPiquetPlayerScoreOps(t *testing.T) {
	p := NewPiquetPlayer(false)
	p.AddDeclScore(5)
	p.AddDeclScore(15)
	p.AddTrickScore(7)
	p.AddBonusScore(30)
	if p.GetDeclScore() != 20 {
		t.Errorf("DeclScore = %d, want 20", p.GetDeclScore())
	}
	if p.GetTrickScore() != 7 {
		t.Errorf("TrickScore = %d, want 7", p.GetTrickScore())
	}
	if p.GetBonusScore() != 30 {
		t.Errorf("BonusScore = %d, want 30", p.GetBonusScore())
	}
	if p.GetRoundScore() != 57 {
		t.Errorf("RoundScore = %d, want 57", p.GetRoundScore())
	}
	p.AddMatchScore(57)
	p.AddMatchScore(43)
	if p.GetMatchScore() != 100 {
		t.Errorf("MatchScore = %d, want 100", p.GetMatchScore())
	}
}

func TestPiquetPlayerSetters(t *testing.T) {
	p := NewPiquetPlayer(false)
	p.SetDeclScore(10)
	p.SetTrickScore(20)
	p.SetBonusScore(30)
	p.SetMatchScore(40)
	if p.GetDeclScore() != 10 || p.GetTrickScore() != 20 || p.GetBonusScore() != 30 || p.GetMatchScore() != 40 {
		t.Errorf("setters did not persist values: %+v", p)
	}
}

func TestPiquetPlayerResetRound(t *testing.T) {
	p := NewPiquetPlayer(true)
	p.AddCard(NewCard(CardDesignSpade, 13, false))
	p.AddTrick([]*Card{NewCard(CardDesignHeart, 1, false)})
	p.AddDeclScore(20)
	p.AddTrickScore(8)
	p.AddBonusScore(10)
	p.AddMatchScore(100)
	p.SetIsFinished(true)

	p.ResetRound()

	if p.GetCardsSize() != 0 {
		t.Errorf("ResetRound did not clear hand: %d cards", p.GetCardsSize())
	}
	if p.GetTrickCount() != 0 {
		t.Errorf("ResetRound did not clear tricks: %d", p.GetTrickCount())
	}
	if p.GetDeclScore() != 0 || p.GetTrickScore() != 0 || p.GetBonusScore() != 0 {
		t.Errorf("ResetRound did not clear round scores")
	}
	if p.GetIsFinished() {
		t.Error("ResetRound did not clear isFinished")
	}
	if p.GetMatchScore() != 100 {
		t.Errorf("ResetRound must not clear matchScore, got %d", p.GetMatchScore())
	}
}

func TestPiquetPlayerResetMatch(t *testing.T) {
	p := NewPiquetPlayer(true)
	p.AddMatchScore(150)
	p.AddDeclScore(20)
	p.ResetMatch()
	if p.GetMatchScore() != 0 {
		t.Errorf("ResetMatch should clear matchScore, got %d", p.GetMatchScore())
	}
	if p.GetDeclScore() != 0 {
		t.Errorf("ResetMatch should clear round scores")
	}
}

func TestPiquetPlayerJSONRoundTrip(t *testing.T) {
	orig := NewPiquetPlayer(true)
	orig.AddCard(NewCard(CardDesignSpade, 13, false))
	orig.AddCard(NewCard(CardDesignHeart, 1, false))
	orig.AddTrick([]*Card{NewCard(CardDesignClover, 12, false)})
	orig.SetDeclScore(15)
	orig.SetTrickScore(7)
	orig.SetBonusScore(40)
	orig.SetMatchScore(120)

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got := &PiquetPlayer{}
	if err := json.Unmarshal(data, got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.GetIsHuman() != orig.GetIsHuman() {
		t.Errorf("isHuman: got %v, want %v", got.GetIsHuman(), orig.GetIsHuman())
	}
	if got.GetDeclScore() != orig.GetDeclScore() || got.GetTrickScore() != orig.GetTrickScore() ||
		got.GetBonusScore() != orig.GetBonusScore() || got.GetMatchScore() != orig.GetMatchScore() {
		t.Errorf("scores mismatch")
	}
	if got.GetCardsSize() != 2 {
		t.Errorf("hand size: got %d, want 2", got.GetCardsSize())
	}
	if got.GetTrickCount() != 1 {
		t.Errorf("trick count: got %d, want 1", got.GetTrickCount())
	}
}
