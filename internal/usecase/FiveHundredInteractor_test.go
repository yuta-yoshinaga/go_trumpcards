//go:build test

package usecase_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newFiveHundredTestGame() *domain.FiveHundred {
	players := []*domain.FiveHundredPlayer{
		domain.NewFiveHundredPlayer(true, 0),
		domain.NewFiveHundredPlayer(false, 1),
		domain.NewFiveHundredPlayer(false, 0),
		domain.NewFiveHundredPlayer(false, 1),
	}
	return domain.NewFiveHundred(domain.NewTrumpCardsFiveHundred(), players, domain.DefaultFiveHundredConfig())
}

func newFiveHundredInteractor(g *domain.FiveHundred) (*usecase.FiveHundredInteractor, *presenter.MockFiveHundredPresenter) {
	fp := new(presenter.MockFiveHundredPresenter)
	fp.On("Output", mock.Anything, mock.Anything).Return("{}")
	fp.On("HintOutput", mock.Anything).Return("{hint}")
	fp.On("ActionLogOutput", mock.Anything).Return("[]")
	return usecase.NewFiveHundredInteractor(g, fp), fp
}

func TestFiveHundredInteractor_NilGuards(t *testing.T) {
	fp := new(presenter.MockFiveHundredPresenter)
	assert.PanicsWithValue(t, "FiveHundredInteractor: g must not be nil", func() {
		usecase.NewFiveHundredInteractor(nil, fp)
	})
	assert.PanicsWithValue(t, "FiveHundredInteractor: fp must not be nil", func() {
		usecase.NewFiveHundredInteractor(newFiveHundredTestGame(), nil)
	})
}

func TestFiveHundredInteractor_ResetAndHelpers(t *testing.T) {
	g := newFiveHundredTestGame()
	fi, fp := newFiveHundredInteractor(g)

	fi.Reset()
	// After Reset, CPUs auto-advance; phase must be a valid in-progress phase.
	if g.GetPhase() == domain.FiveHundredPhaseGameEnd {
		t.Errorf("game should not be over right after reset")
	}
	fp.AssertCalled(t, "Output", mock.Anything, mock.Anything)

	if got := fi.Hint(); got != "{hint}" {
		t.Errorf("Hint() = %q", got)
	}
	if got := fi.ActionLog(); got != "[]" {
		t.Errorf("ActionLog() = %q", got)
	}
	if fi.GetConfig().TargetScore != domain.FiveHundredDefaultTargetScore {
		t.Errorf("GetConfig target score wrong")
	}
	if _, err := fi.Snapshot(); err != nil {
		t.Errorf("Snapshot error: %v", err)
	}
}

func TestFiveHundredInteractor_ResetWithConfig_Invalid(t *testing.T) {
	g := newFiveHundredTestGame()
	fi, _ := newFiveHundredInteractor(g)
	out := fi.ResetWithConfig(domain.FiveHundredConfig{CpuDifficulty: 99, TargetScore: 0})
	if out == "" {
		t.Errorf("expected error output for invalid config")
	}
}

// TestFiveHundredInteractor_FullGame drives an entire game through the
// interactor, exercising bidding, kitty exchange, play, and scoring end to end.
func TestFiveHundredInteractor_FullGame(t *testing.T) {
	g := newFiveHundredTestGame()
	fi, _ := newFiveHundredInteractor(g)
	fi.Reset()

	const maxSteps = 200000
	steps := 0
	for !g.GetGameEndFlag() && steps < maxSteps {
		steps++
		switch g.GetPhase() {
		case domain.FiveHundredPhaseBid:
			if !g.IsHumanBidTurn() {
				// Should not happen — CPUs auto-advance — but guard anyway.
				t.Fatalf("stuck on CPU bid turn")
			}
			// Human opens with the lowest bid when nobody has bid yet, else passes.
			if g.GetHighestBid() == nil {
				fi.Bid(domain.FiveHundredContractSuit, 6, domain.CardDesignSpade)
			} else {
				fi.Pass()
			}
		case domain.FiveHundredPhaseKittyExchange:
			// Only the human declarer reaches here (CPU exchange is automatic).
			fi.ExchangeKitty([]int{0, 1, 2})
		case domain.FiveHundredPhasePlay:
			if !g.IsHumanTurn() {
				t.Fatalf("stuck on CPU play turn")
			}
			valid := g.GetValidPlayIndices(g.GetCurrentPlayerIdx())
			if len(valid) == 0 {
				t.Fatalf("human has no valid plays")
			}
			fi.Play(valid[0], -1)
		case domain.FiveHundredPhaseTrickEnd:
			fi.NextTrick()
		case domain.FiveHundredPhaseRoundEnd:
			fi.NextRound()
		default:
			t.Fatalf("unexpected phase %d", g.GetPhase())
		}
	}
	if !g.GetGameEndFlag() {
		t.Fatalf("game did not finish within %d steps", maxSteps)
	}
	if g.GetWinnerTeam() < 0 {
		t.Errorf("winner team not set")
	}
}

func TestFiveHundredInteractor_GuardAfterGameEnd(t *testing.T) {
	g := newFiveHundredTestGame()
	g.Reset()
	// Drive a winning round so the game-end flag is genuinely set.
	g.SetContract(domain.FiveHundredContractSuit, 7, domain.CardDesignSpade)
	g.SetDeclarerIdx(0)
	g.SetTeamScore(0, 400)
	for i := 0; i < 7; i++ {
		g.GetPlayer(i % 2 * 2).AddTrick([]*domain.Card{nil})
	}
	for i := 0; i < 3; i++ {
		g.GetPlayer(1).AddTrick([]*domain.Card{nil})
	}
	g.SetPhase(domain.FiveHundredPhaseRoundEnd)
	g.ScoreRound()
	if !g.GetGameEndFlag() {
		t.Fatalf("precondition: game should have ended")
	}
	fi, _ := newFiveHundredInteractor(g)
	// These should be no-ops returning output, not panicking.
	_ = fi.Bid(domain.FiveHundredContractSuit, 6, domain.CardDesignSpade)
	_ = fi.Pass()
	_ = fi.ExchangeKitty([]int{0, 1, 2})
	_ = fi.NextRound()
}

func TestFiveHundredInteractor_RestoreRoundTrip(t *testing.T) {
	g := newFiveHundredTestGame()
	fi, _ := newFiveHundredInteractor(g)
	fi.Reset()
	data, err := fi.Snapshot()
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	fp := new(presenter.MockFiveHundredPresenter)
	fp.On("Output", mock.Anything, mock.Anything).Return("{}")
	restored, err := usecase.RestoreFiveHundredInteractor(data, fp)
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if out := restored.Reset(); !strings.Contains(out, "{") {
		t.Errorf("restored interactor output unexpected: %q", out)
	}
}
