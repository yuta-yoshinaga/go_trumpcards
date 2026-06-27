package presenter_test

import (
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func spBuildPlayedScopone(t *testing.T) *domain.Scopone {
	t.Helper()
	s := domain.NewDefaultScopone()
	s.Reset()
	return s
}

func TestScoponeCuiPresenter_HintOutput(t *testing.T) {
	p := &presenter.ScoponeCuiPresenter{}
	build := func(handVal, handSuit int, table []*domain.Card) *domain.Scopone {
		s := domain.NewDefaultScopone()
		s.SetPhase(domain.ScoponePhasePlayerTurn)
		s.SetCurrentTurn(0)
		s.GetPlayer(0).AddCard(domain.NewCard(handSuit, handVal, false))
		s.SetTableCards(table)
		return s
	}

	t.Run("scopa sweep", func(t *testing.T) {
		s := build(7, domain.CardDesignHeart, []*domain.Card{
			domain.NewCard(domain.CardDesignDiamond, 3, false),
			domain.NewCard(domain.CardDesignClover, 4, false),
		})
		if out := p.HintOutput(s); !strings.Contains(out, "スコパ") {
			t.Errorf("expected scopa hint, got: %s", out)
		}
	})

	t.Run("plain capture", func(t *testing.T) {
		s := build(7, domain.CardDesignHeart, []*domain.Card{
			domain.NewCard(domain.CardDesignDiamond, 7, false),
			domain.NewCard(domain.CardDesignClover, 5, false),
		})
		if out := p.HintOutput(s); !strings.Contains(out, "捕獲") {
			t.Errorf("expected capture hint, got: %s", out)
		}
	})

	t.Run("no capture", func(t *testing.T) {
		s := build(2, domain.CardDesignHeart, []*domain.Card{
			domain.NewCard(domain.CardDesignDiamond, 7, false),
			domain.NewCard(domain.CardDesignClover, 5, false),
		})
		if out := p.HintOutput(s); !strings.Contains(out, "捕獲できる手はありません") {
			t.Errorf("expected no-capture hint, got: %s", out)
		}
	})

	t.Run("none outside player turn", func(t *testing.T) {
		s := domain.NewDefaultScopone()
		s.SetPhase(domain.ScoponePhaseRoundEnd)
		if out := p.HintOutput(s); !strings.Contains(out, "ヒントはありません") {
			t.Errorf("expected none hint, got: %s", out)
		}
	})
}

func TestScoponeCuiPresenter_Output(t *testing.T) {
	p := &presenter.ScoponeCuiPresenter{}
	s := spBuildPlayedScopone(t)
	if out := p.Output(s, nil); out == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestScoponeCuiPresenter_Error(t *testing.T) {
	p := &presenter.ScoponeCuiPresenter{}
	s := spBuildPlayedScopone(t)
	out := p.Output(s, scoponeAssertErr{})
	if !strings.Contains(out, "boom") {
		t.Errorf("expected error text in output, got: %s", out)
	}
}

type scoponeAssertErr struct{}

func (scoponeAssertErr) Error() string { return "boom" }

func TestScoponeCuiPresenter_ActionLogOutput(t *testing.T) {
	p := &presenter.ScoponeCuiPresenter{}
	s := spBuildPlayedScopone(t)
	if out := p.ActionLogOutput(s); out == "" {
		t.Error("expected non-empty action log output")
	}
}

func TestScoponeCuiPresenter_OutputGameEnd(t *testing.T) {
	p := &presenter.ScoponeCuiPresenter{}
	s := spPlayedOutScopone(t) // all-CPU game driven to game end (defined in the web presenter test)
	out := p.Output(s, nil)
	if out == "" {
		t.Fatal("expected non-empty output at game end")
	}
}
