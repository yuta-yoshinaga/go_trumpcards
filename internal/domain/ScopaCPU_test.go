package domain

import "testing"

// cpuReady builds a drained-deck game where it is the CPU's (player 1) turn.
func cpuReady(t *testing.T, diff ScopaCpuDifficulty) *Scopa {
	t.Helper()
	cfg := DefaultScopaConfig()
	cfg.CpuDifficulty = diff
	s := ScopaTestNew(cfg)
	s.ScopaTestDrainDeck()
	s.ScopaTestSetPhase(ScopaPhasePlayerTurn)
	s.ScopaTestSetCurrentTurn(1)
	return s
}

func TestScopaCpuPlay_PrefersCapture(t *testing.T) {
	s := cpuReady(t, ScopaDifficultyNormal)
	s.GetPlayer(1).AddCard(NewCard(CardDesignDiamond, 7, false)) // can capture the 7♠
	s.GetPlayer(0).AddCard(NewCard(CardDesignClover, 4, false))  // keep hands non-empty
	s.ScopaTestSetTable([]*Card{
		NewCard(CardDesignSpade, 7, false),
		NewCard(CardDesignHeart, 2, false),
	})
	s.CpuPlay()
	if s.GetPlayer(1).CapturedCount() == 0 {
		t.Error("CPU should capture when possible")
	}
	acts := s.GetCpuActions()
	if len(acts) != 1 {
		t.Fatalf("expected 1 cpu action, got %d", len(acts))
	}
	if len(acts[0].CapturedCards) == 0 {
		t.Error("cpu action should record captured cards")
	}
}

func TestScopaCpuPlay_LaysWhenNoCapture(t *testing.T) {
	s := cpuReady(t, ScopaDifficultyNormal)
	s.GetPlayer(1).AddCard(NewCard(CardDesignClover, 4, false))
	s.GetPlayer(0).AddCard(NewCard(CardDesignClover, 5, false))
	s.ScopaTestSetTable([]*Card{NewCard(CardDesignHeart, 11, false)}) // value 8, no capture for a 4
	s.CpuPlay()
	if s.GetPlayer(1).CapturedCount() != 0 {
		t.Error("CPU cannot capture, should lay")
	}
	if len(s.GetTableCards()) != 2 {
		t.Errorf("table should grow to 2 after lay, got %d", len(s.GetTableCards()))
	}
}

func TestScopaCpuPlay_AllDifficulties(t *testing.T) {
	for _, diff := range []ScopaCpuDifficulty{ScopaDifficultyEasy, ScopaDifficultyNormal, ScopaDifficultyHard} {
		s := cpuReady(t, diff)
		s.GetPlayer(1).AddCard(NewCard(CardDesignDiamond, 5, false))
		s.GetPlayer(0).AddCard(NewCard(CardDesignClover, 4, false))
		s.ScopaTestSetTable([]*Card{NewCard(CardDesignSpade, 5, false)})
		s.CpuPlay()
		if len(s.GetCpuActions()) != 1 {
			t.Errorf("difficulty %d: expected a cpu move", diff)
		}
	}
}

func TestScopaCpuPlay_NoOpOffTurn(t *testing.T) {
	s := cpuReady(t, ScopaDifficultyNormal)
	s.ScopaTestSetCurrentTurn(0) // human turn
	s.CpuPlay()
	if len(s.GetCpuActions()) != 0 {
		t.Error("CpuPlay must be a no-op on the human's turn")
	}
}

func TestScopaCardValueScore(t *testing.T) {
	if scopaCardValueScore(nil) != 0 {
		t.Error("nil scores 0")
	}
	sette := scopaCardValueScore(NewCard(CardDesignDiamond, 7, false))
	plain := scopaCardValueScore(NewCard(CardDesignClover, 4, false))
	if sette <= plain {
		t.Error("settebello should score higher than a plain card")
	}
}
