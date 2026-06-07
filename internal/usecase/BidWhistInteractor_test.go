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

func newBidWhistTestGame() *domain.BidWhist {
	players := []*domain.BidWhistPlayer{
		domain.NewBidWhistPlayer(true, 0),
		domain.NewBidWhistPlayer(false, 1),
		domain.NewBidWhistPlayer(false, 0),
		domain.NewBidWhistPlayer(false, 1),
	}
	return domain.NewBidWhist(domain.NewTrumpCards(2), players, domain.DefaultBidWhistConfig())
}

func newBidWhistInteractor(g *domain.BidWhist) (*usecase.BidWhistInteractor, *presenter.MockBidWhistPresenter) {
	fp := new(presenter.MockBidWhistPresenter)
	fp.On("Output", mock.Anything, mock.Anything).Return("{}")
	fp.On("HintOutput", mock.Anything).Return("{hint}")
	fp.On("ActionLogOutput", mock.Anything).Return("[]")
	return usecase.NewBidWhistInteractor(g, fp), fp
}

func TestBidWhistInteractor_NilGuards(t *testing.T) {
	fp := new(presenter.MockBidWhistPresenter)
	assert.PanicsWithValue(t, "BidWhistInteractor: g must not be nil", func() {
		usecase.NewBidWhistInteractor(nil, fp)
	})
	assert.PanicsWithValue(t, "BidWhistInteractor: fp must not be nil", func() {
		usecase.NewBidWhistInteractor(newBidWhistTestGame(), nil)
	})
}

func TestBidWhistInteractor_ResetAndHelpers(t *testing.T) {
	g := newBidWhistTestGame()
	bi, fp := newBidWhistInteractor(g)

	bi.Reset()
	if g.GetPhase() == domain.BidWhistPhaseGameEnd {
		t.Errorf("game should not be over right after reset")
	}
	fp.AssertCalled(t, "Output", mock.Anything, mock.Anything)

	if got := bi.Hint(); got != "{hint}" {
		t.Errorf("Hint() = %q", got)
	}
	if got := bi.ActionLog(); got != "[]" {
		t.Errorf("ActionLog() = %q", got)
	}
	if bi.GetConfig().TargetScore != domain.BidWhistDefaultTargetScore {
		t.Errorf("GetConfig target score wrong")
	}
	if _, err := bi.Snapshot(); err != nil {
		t.Errorf("Snapshot error: %v", err)
	}
}

func TestBidWhistInteractor_ResetWithConfig_Invalid(t *testing.T) {
	g := newBidWhistTestGame()
	bi, _ := newBidWhistInteractor(g)
	out := bi.ResetWithConfig(domain.BidWhistConfig{CpuDifficulty: 99, TargetScore: 0})
	if out == "" {
		t.Errorf("expected error output for invalid config")
	}
}

// TestBidWhistInteractor_FullGame drives an entire game through the interactor,
// exercising bidding, trump declaration, kitty exchange, play, and scoring.
func TestBidWhistInteractor_FullGame(t *testing.T) {
	g := newBidWhistTestGame()
	bi, _ := newBidWhistInteractor(g)
	bi.Reset()

	const maxSteps = 300000
	steps := 0
	for !g.GetGameEndFlag() && steps < maxSteps {
		steps++
		switch g.GetPhase() {
		case domain.BidWhistPhaseBid:
			if !g.IsHumanBidTurn() {
				t.Fatalf("stuck on CPU bid turn")
			}
			if g.GetHighestBid() == nil {
				bi.Bid(domain.BidWhistMinBid, domain.BidWhistDirectionUptown)
			} else {
				bi.Pass()
			}
		case domain.BidWhistPhaseTrumpDeclaration:
			bi.DeclareTrump(domain.CardDesignSpade)
		case domain.BidWhistPhaseKittyExchange:
			bi.ExchangeKitty([]int{0, 1, 2, 3, 4, 5})
		case domain.BidWhistPhasePlay:
			if !g.IsHumanTurn() {
				t.Fatalf("stuck on CPU play turn")
			}
			valid := g.GetValidPlayIndices(g.GetCurrentPlayerIdx())
			if len(valid) == 0 {
				t.Fatalf("human has no valid plays")
			}
			bi.Play(valid[0])
		case domain.BidWhistPhaseTrickEnd:
			bi.NextTrick()
		case domain.BidWhistPhaseRoundEnd:
			bi.NextRound()
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

func TestBidWhistInteractor_GuardAfterGameEnd(t *testing.T) {
	g := newBidWhistTestGame()
	g.Reset()
	// Bid 1 needs 6+1=7 tricks; sweeping all 12 makes it for +6, pushing team 0
	// from 6 to 12 (>= target 7) so the game genuinely ends.
	g.SetContract(1, domain.BidWhistDirectionUptown, domain.CardDesignSpade)
	g.SetDeclarerIdx(0)
	g.GetPlayer(0).SetIsDeclarer(true)
	g.SetTeamScore(0, 6)
	for i := 0; i < 12; i++ {
		g.GetPlayer(0).AddTrick([]*domain.Card{nil})
	}
	g.SetPhase(domain.BidWhistPhaseRoundEnd)
	g.ScoreRound()
	if !g.GetGameEndFlag() {
		t.Fatalf("precondition: game should have ended")
	}
	bi, _ := newBidWhistInteractor(g)
	_ = bi.Bid(1, domain.BidWhistDirectionUptown)
	_ = bi.Pass()
	_ = bi.DeclareTrump(domain.CardDesignSpade)
	_ = bi.ExchangeKitty([]int{0, 1, 2, 3, 4, 5})
	_ = bi.NextRound()
}

func TestBidWhistInteractor_RestoreRoundTrip(t *testing.T) {
	g := newBidWhistTestGame()
	bi, _ := newBidWhistInteractor(g)
	bi.Reset()
	data, err := bi.Snapshot()
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	fp := new(presenter.MockBidWhistPresenter)
	fp.On("Output", mock.Anything, mock.Anything).Return("{}")
	restored, err := usecase.RestoreBidWhistInteractor(data, fp)
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if out := restored.Reset(); !strings.Contains(out, "{") {
		t.Errorf("restored interactor output unexpected: %q", out)
	}
}
