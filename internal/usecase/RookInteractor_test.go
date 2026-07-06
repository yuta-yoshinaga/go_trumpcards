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

func newRookTestGame() *domain.Rook {
	players := []*domain.RookPlayer{
		domain.NewRookPlayer(true, 0),
		domain.NewRookPlayer(false, 1),
		domain.NewRookPlayer(false, 0),
		domain.NewRookPlayer(false, 1),
	}
	return domain.NewRook(players, domain.DefaultRookConfig())
}

func newRookInteractor(g *domain.Rook) (*usecase.RookInteractor, *presenter.MockRookPresenter) {
	fp := new(presenter.MockRookPresenter)
	fp.On("Output", mock.Anything, mock.Anything).Return("{}")
	fp.On("HintOutput", mock.Anything).Return("{hint}")
	fp.On("ActionLogOutput", mock.Anything).Return("[]")
	return usecase.NewRookInteractor(g, fp), fp
}

func TestRookInteractor_NilGuards(t *testing.T) {
	fp := new(presenter.MockRookPresenter)
	assert.PanicsWithValue(t, "RookInteractor: g must not be nil", func() {
		usecase.NewRookInteractor(nil, fp)
	})
	assert.PanicsWithValue(t, "RookInteractor: fp must not be nil", func() {
		usecase.NewRookInteractor(newRookTestGame(), nil)
	})
}

func TestRookInteractor_ResetAndHelpers(t *testing.T) {
	g := newRookTestGame()
	fi, fp := newRookInteractor(g)

	fi.Reset()
	if g.GetPhase() == domain.RookPhaseGameEnd {
		t.Errorf("game should not be over right after reset")
	}
	fp.AssertCalled(t, "Output", mock.Anything, mock.Anything)

	if got := fi.Hint(); got != "{hint}" {
		t.Errorf("Hint() = %q", got)
	}
	if got := fi.ActionLog(); got != "[]" {
		t.Errorf("ActionLog() = %q", got)
	}
	if fi.GetConfig().TargetScore != domain.RookDefaultTargetScore {
		t.Errorf("GetConfig target score wrong")
	}
	if _, err := fi.Snapshot(); err != nil {
		t.Errorf("Snapshot error: %v", err)
	}
}

func TestRookInteractor_ResetWithConfig_Invalid(t *testing.T) {
	g := newRookTestGame()
	fi, _ := newRookInteractor(g)
	out := fi.ResetWithConfig(domain.RookConfig{CpuDifficulty: 99, TargetScore: 0})
	if out == "" {
		t.Errorf("expected error output for invalid config")
	}
}

func TestRookInteractor_FullGame(t *testing.T) {
	g := newRookTestGame()
	fi, _ := newRookInteractor(g)
	fi.Reset()

	const maxSteps = 200000
	steps := 0
	for !g.GetGameEndFlag() && steps < maxSteps {
		steps++
		switch g.GetPhase() {
		case domain.RookPhaseBid:
			if !g.IsHumanBidTurn() {
				t.Fatalf("stuck on CPU bid turn")
			}
			if g.GetHighestBid() == 0 {
				fi.Bid(domain.RookMinBid)
			} else {
				fi.Pass()
			}
		case domain.RookPhaseNestExchange:
			// Only a human declarer reaches here (CPU exchange is automatic).
			fi.ExchangeNest([]int{0, 1, 2, 3, 4}, 1)
		case domain.RookPhasePlay:
			if !g.IsHumanTurn() {
				t.Fatalf("stuck on CPU play turn")
			}
			valid := g.GetValidPlayIndices(g.GetCurrentPlayerIdx())
			if len(valid) == 0 {
				t.Fatalf("human has no valid plays")
			}
			fi.Play(valid[0])
		case domain.RookPhaseTrickEnd:
			fi.NextTrick()
		case domain.RookPhaseRoundEnd:
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

func TestRookInteractor_GuardAfterGameEnd(t *testing.T) {
	g := newRookTestGame()
	g.Reset()
	g.SetDeclarerIdx(0)
	g.SetContractBid(70)
	g.SetTeamScore(0, 450)
	g.GetPlayer(0).SetPoints(100)
	g.SetPhase(domain.RookPhaseRoundEnd)
	g.ScoreRound()
	if !g.GetGameEndFlag() {
		t.Fatalf("precondition: game should have ended")
	}
	fi, _ := newRookInteractor(g)
	_ = fi.Bid(70)
	_ = fi.Pass()
	_ = fi.ExchangeNest([]int{0, 1, 2, 3, 4}, 1)
	_ = fi.NextRound()
}

func TestRookInteractor_RestoreRoundTrip(t *testing.T) {
	g := newRookTestGame()
	fi, _ := newRookInteractor(g)
	fi.Reset()
	data, err := fi.Snapshot()
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	fp := new(presenter.MockRookPresenter)
	fp.On("Output", mock.Anything, mock.Anything).Return("{}")
	restored, err := usecase.RestoreRookInteractor(data, fp)
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if out := restored.Reset(); !strings.Contains(out, "{") {
		t.Errorf("restored interactor output unexpected: %q", out)
	}
}
