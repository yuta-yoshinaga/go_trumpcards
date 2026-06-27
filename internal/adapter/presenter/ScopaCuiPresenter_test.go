package presenter_test

import (
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func buildPlayedScopa(t *testing.T) *domain.Scopa {
	t.Helper()
	s := domain.NewDefaultScopa()
	s.Reset()
	return s
}

func TestScopaCuiPresenter_Output(t *testing.T) {
	p := &presenter.ScopaCuiPresenter{}
	s := buildPlayedScopa(t)
	if out := p.Output(s, nil); out == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestScopaCuiPresenter_Error(t *testing.T) {
	p := &presenter.ScopaCuiPresenter{}
	s := buildPlayedScopa(t)
	out := p.Output(s, scopaAssertErr{})
	if !strings.Contains(out, "boom") {
		t.Errorf("expected error text in output, got: %s", out)
	}
}

type scopaAssertErr struct{}

func (scopaAssertErr) Error() string { return "boom" }

func TestScopaCuiPresenter_HintOutput(t *testing.T) {
	p := &presenter.ScopaCuiPresenter{}
	build := func(handVal, handSuit int, table []*domain.Card) *domain.Scopa {
		s := domain.ScopaTestNew(domain.DefaultScopaConfig())
		s.ScopaTestSetPhase(domain.ScopaPhasePlayerTurn)
		s.ScopaTestSetCurrentTurn(0)
		s.GetPlayer(0).AddCard(domain.NewCard(handSuit, handVal, false))
		s.ScopaTestSetTable(table)
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
		s := domain.ScopaTestNew(domain.DefaultScopaConfig())
		s.ScopaTestSetPhase(domain.ScopaPhaseRoundEnd)
		if out := p.HintOutput(s); !strings.Contains(out, "ヒントはありません") {
			t.Errorf("expected none hint, got: %s", out)
		}
	})
}

func TestScopaCuiPresenter_ActionLogOutput(t *testing.T) {
	p := &presenter.ScopaCuiPresenter{}
	s := buildPlayedScopa(t)
	if out := p.ActionLogOutput(s); out == "" {
		t.Error("expected non-empty action log output")
	}
}
