package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestScopaActionJSONRoundTrip(t *testing.T) {
	a := &domain.ScopaAction{
		PlayerIdx:     1,
		PlayedCard:    domain.NewCard(domain.CardDesignDiamond, 7, false),
		CapturedCards: []*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)},
		IsScopa:       true,
	}
	data, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got domain.ScopaAction
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.PlayerIdx != a.PlayerIdx || !got.IsScopa || got.PlayedCard.GetValue() != 7 {
		t.Errorf("action round-trip mismatch: %+v", got)
	}
	if len(got.CapturedCards) != 1 {
		t.Errorf("captured cards lost in round-trip")
	}
}

func TestScopaScoreDetailJSONRoundTrip(t *testing.T) {
	d := &domain.ScopaScoreDetail{
		Cards:         map[int]int{0: 20, 1: 20},
		Diamonds:      map[int]int{0: 6, 1: 4},
		Sevens:        map[int]int{0: 3, 1: 1},
		HasSetteBello: 0,
		Scopas:        map[int]int{0: 1, 1: 0},
		Gained:        map[int]int{0: 4, 1: 0},
	}
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got domain.ScopaScoreDetail
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.HasSetteBello != 0 || got.Gained[0] != 4 || got.Diamonds[0] != 6 {
		t.Errorf("score detail round-trip mismatch: %+v", got)
	}
}

func TestScopaPlayerJSONRoundTrip(t *testing.T) {
	p := domain.NewScopaPlayer(true)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
	p.AddCaptured([]*domain.Card{domain.NewCard(domain.CardDesignDiamond, 7, false)})
	p.IncrementScopa()
	p.AddScore(5)
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got domain.ScopaPlayer
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.GetScopaCount() != 1 || got.GetTotalScore() != 5 || got.CapturedCount() != 1 {
		t.Errorf("player round-trip mismatch: scopa=%d score=%d captured=%d",
			got.GetScopaCount(), got.GetTotalScore(), got.CapturedCount())
	}
	if !got.GetIsHuman() {
		t.Error("isHuman lost in round-trip")
	}
}

func TestScopaGettersAndSetters(t *testing.T) {
	s := domain.NewDefaultScopa()
	s.Reset()
	if s.GetCurrentTurn() != 0 {
		t.Errorf("current turn = %d, want 0", s.GetCurrentTurn())
	}
	if s.GetPacksDealt() < 1 {
		t.Error("packs dealt should be >= 1 after reset")
	}
	if s.GetActionLog() == nil {
		t.Error("action log should be non-nil after reset")
	}
	cfg := s.GetConfig()
	cfg.CpuDifficulty = domain.ScopaDifficultyHard
	s.SetConfig(cfg)
	if s.GetConfig().CpuDifficulty != domain.ScopaDifficultyHard {
		t.Error("SetConfig did not persist")
	}
}
