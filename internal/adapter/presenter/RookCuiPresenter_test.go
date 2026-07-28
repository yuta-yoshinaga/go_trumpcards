package presenter_test

import (
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestRookCuiPresenter_Output_Phases(t *testing.T) {
	p := &presenter.RookCuiPresenter{}

	t.Run("bid phase", func(t *testing.T) {
		g := newRookGame()
		g.Reset()
		out := p.Output(g, nil)
		if out == "" || !strings.Contains(out, "Rook") {
			t.Errorf("bid output missing title: %q", out)
		}
	})

	t.Run("nest exchange phase", func(t *testing.T) {
		g := newRookGame()
		g.SetContractBid(80)
		g.SetDeclarerIdx(0)
		g.GetPlayer(0).SetIsDeclarer(true)
		g.SetNest([]*domain.Card{domain.NewCard(1, 5, false)})
		g.SetPhase(domain.RookPhaseNestExchange)
		if out := p.Output(g, nil); out == "" {
			t.Errorf("nest exchange output empty")
		}
	})

	t.Run("play phase with trick", func(t *testing.T) {
		g := newRookGame()
		g.SetContractBid(80)
		g.SetDeclarerIdx(0)
		g.SetTrumpColor(2)
		g.SetPhase(domain.RookPhasePlay)
		g.SetCurrentTrick([]*domain.TrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(1, 1, false)},
			{PlayerIdx: 1, Card: domain.NewCard(domain.RookBirdDesign, domain.RookBirdValue, false)},
		})
		if out := p.Output(g, nil); out == "" {
			t.Errorf("play output empty")
		}
	})

	t.Run("game end", func(t *testing.T) {
		g := newRookGame()
		g.SetTeamScore(0, 500)
		g.SetDeclarerIdx(0)
		g.SetContractBid(70)
		g.SetPhase(domain.RookPhaseRoundEnd)
		g.ScoreRound()
		out := p.Output(g, nil)
		if !strings.Contains(out, "0") {
			t.Errorf("game end banner missing winner: %q", out)
		}
	})
}

func TestRookCuiPresenter_HintOutput(t *testing.T) {
	p := &presenter.RookCuiPresenter{}

	t.Run("play hint", func(t *testing.T) {
		g := newRookGame()
		g.SetTrumpColor(1)
		g.SetDeclarerIdx(0)
		g.GetPlayer(0).AddCard(domain.NewCard(1, 1, false))
		g.SetPhase(domain.RookPhasePlay)
		g.SetCurrentPlayerIdx(0)
		if out := p.HintOutput(g); out == "" {
			t.Errorf("play hint empty")
		}
	})

	t.Run("no hint when not human turn", func(t *testing.T) {
		g := newRookGame()
		g.SetTrumpColor(1)
		g.SetPhase(domain.RookPhasePlay)
		g.SetCurrentPlayerIdx(1) // CPU turn
		out := p.HintOutput(g)
		if !strings.Contains(out, "ヒント") {
			t.Errorf("expected 'no hint' message, got %q", out)
		}
	})
}

func TestRookCuiPresenter_ActionLog(t *testing.T) {
	g := newRookGame()
	g.Reset()
	p := &presenter.RookCuiPresenter{}
	if out := p.ActionLogOutput(g); out == "" {
		t.Errorf("action log empty")
	}
}

func TestRookCuiPresenter_BidAndHintVariants(t *testing.T) {
	p := &presenter.RookCuiPresenter{}

	// Player bid labels + pass + declarer status.
	g := newRookGame()
	g.GetPlayer(0).SetBid(75)
	g.GetPlayer(1).SetPassed(true)
	g.GetPlayer(2).SetIsDeclarer(true)
	g.SetPhase(domain.RookPhaseBid)
	if out := p.Output(g, nil); out == "" {
		t.Errorf("bid-label output empty")
	}

	// Highest-bid display (contract undecided, a bid stands).
	g2 := newRookGame()
	g2.Reset()
	g2.SetBidPlayerIdx(0)
	if err := g2.PlayerBid(70); err != nil {
		t.Fatalf("bid failed: %v", err)
	}
	if out := p.Output(g2, nil); out == "" {
		t.Errorf("highest-bid output empty")
	}

	// Pass hint (weak hand).
	g3 := newRookGame()
	g3.SetPhase(domain.RookPhaseBid)
	g3.SetBidPlayerIdx(0)
	for v := 2; v <= 6; v++ {
		g3.GetPlayer(0).AddCard(domain.NewCard(3, v, false))
	}
	if out := p.HintOutput(g3); out == "" {
		t.Errorf("pass hint empty")
	}

	// Bid hint (strong hand) + discard hint (declarer in nest exchange).
	g4 := newRookGame()
	g4.SetDeclarerIdx(0)
	g4.SetContractBid(80)
	g4.SetPhase(domain.RookPhaseNestExchange)
	for i := 0; i < 18; i++ {
		g4.GetPlayer(0).AddCard(domain.NewCard((i%4)+1, (i%14)+1, false))
	}
	if out := p.HintOutput(g4); out == "" {
		t.Errorf("discard hint empty")
	}
}
