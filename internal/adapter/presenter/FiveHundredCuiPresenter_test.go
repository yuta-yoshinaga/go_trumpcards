package presenter_test

import (
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestFiveHundredCuiPresenter_Output_Phases(t *testing.T) {
	p := &presenter.FiveHundredCuiPresenter{}

	t.Run("bid phase", func(t *testing.T) {
		g := newFiveHundredGame()
		g.Reset()
		out := p.Output(g, nil)
		if out == "" || !strings.Contains(out, "500") {
			t.Errorf("bid output missing title: %q", out)
		}
	})

	t.Run("kitty exchange phase", func(t *testing.T) {
		g := newFiveHundredGame()
		g.SetContract(domain.FiveHundredContractSuit, 7, domain.CardDesignSpade)
		g.SetDeclarerIdx(0)
		g.GetPlayer(0).SetIsDeclarer(true)
		g.SetPhase(domain.FiveHundredPhaseKittyExchange)
		if out := p.Output(g, nil); out == "" {
			t.Errorf("kitty exchange output empty")
		}
	})

	t.Run("play phase with trick", func(t *testing.T) {
		g := newFiveHundredGame()
		g.SetContract(domain.FiveHundredContractNoTrump, 7, -1)
		g.SetPhase(domain.FiveHundredPhasePlay)
		g.SetCurrentTrick([]*domain.FiveHundredTrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 1, false)},
		})
		if out := p.Output(g, nil); out == "" {
			t.Errorf("play output empty")
		}
	})

	t.Run("game end", func(t *testing.T) {
		g := newFiveHundredGame()
		g.SetTeamScore(0, 500)
		g.SetContract(domain.FiveHundredContractSuit, 7, domain.CardDesignSpade)
		g.SetDeclarerIdx(0)
		g.SetPhase(domain.FiveHundredPhaseRoundEnd)
		g.ScoreRound()
		out := p.Output(g, nil)
		if !strings.Contains(out, "0") {
			t.Errorf("game end banner missing winner: %q", out)
		}
	})
}

func TestFiveHundredCuiPresenter_HintOutput(t *testing.T) {
	p := &presenter.FiveHundredCuiPresenter{}

	t.Run("play hint", func(t *testing.T) {
		g := newFiveHundredGame()
		g.SetContract(domain.FiveHundredContractSuit, 7, domain.CardDesignSpade)
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		g.SetPhase(domain.FiveHundredPhasePlay)
		g.SetCurrentPlayerIdx(0)
		if out := p.HintOutput(g); out == "" {
			t.Errorf("play hint empty")
		}
	})

	t.Run("no hint when not human turn", func(t *testing.T) {
		g := newFiveHundredGame()
		g.SetContract(domain.FiveHundredContractSuit, 7, domain.CardDesignSpade)
		g.SetPhase(domain.FiveHundredPhasePlay)
		g.SetCurrentPlayerIdx(1) // CPU turn
		out := p.HintOutput(g)
		if !strings.Contains(out, "ヒント") {
			t.Errorf("expected 'no hint' message, got %q", out)
		}
	})
}

func TestFiveHundredCuiPresenter_ActionLog(t *testing.T) {
	g := newFiveHundredGame()
	g.Reset()
	p := &presenter.FiveHundredCuiPresenter{}
	if out := p.ActionLogOutput(g); out == "" {
		t.Errorf("action log empty")
	}
}
