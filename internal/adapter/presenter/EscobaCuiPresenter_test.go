package presenter_test

import (
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func escobaBuildCuiGame(t *testing.T) *domain.Escoba {
	t.Helper()
	e := domain.NewDefaultEscoba()
	e.Reset()
	return e
}

func TestEscobaCuiPresenter_HintOutput(t *testing.T) {
	p := &presenter.EscobaCuiPresenter{}
	build := func(handVal, handSuit int, table []*domain.Card) *domain.Escoba {
		e := domain.NewDefaultEscoba()
		e.SetPhase(domain.EscobaPhasePlayerTurn)
		e.SetCurrentTurn(0)
		e.GetPlayer(0).AddCard(domain.NewCard(handSuit, handVal, false))
		e.SetTableCards(table)
		return e
	}

	t.Run("escoba sweep", func(t *testing.T) {
		// played 5 (target 10) + table 3+7 = 10, clears the table.
		e := build(5, domain.CardDesignHeart, []*domain.Card{
			domain.NewCard(domain.CardDesignDiamond, 3, false),
			domain.NewCard(domain.CardDesignClover, 7, false),
		})
		if out := p.HintOutput(e); !strings.Contains(out, "エスコバ") {
			t.Errorf("expected escoba hint, got: %s", out)
		}
	})

	t.Run("plain capture", func(t *testing.T) {
		// played 5 (target 10) + table 3+7 = 10, but a 9 remains on the table.
		e := build(5, domain.CardDesignHeart, []*domain.Card{
			domain.NewCard(domain.CardDesignDiamond, 3, false),
			domain.NewCard(domain.CardDesignClover, 7, false),
			domain.NewCard(domain.CardDesignSpade, 9, false),
		})
		if out := p.HintOutput(e); !strings.Contains(out, "捕獲") {
			t.Errorf("expected capture hint, got: %s", out)
		}
	})

	t.Run("no capture", func(t *testing.T) {
		// played 2 (target 13); no subset of {3,7} sums to 13.
		e := build(2, domain.CardDesignHeart, []*domain.Card{
			domain.NewCard(domain.CardDesignDiamond, 3, false),
			domain.NewCard(domain.CardDesignClover, 7, false),
		})
		if out := p.HintOutput(e); !strings.Contains(out, "捕獲できる手はありません") {
			t.Errorf("expected no-capture hint, got: %s", out)
		}
	})

	t.Run("none outside player turn", func(t *testing.T) {
		e := domain.NewDefaultEscoba()
		e.SetPhase(domain.EscobaPhaseRoundEnd)
		if out := p.HintOutput(e); !strings.Contains(out, "ヒントはありません") {
			t.Errorf("expected none hint, got: %s", out)
		}
	})
}

func TestEscobaCuiPresenter_Output(t *testing.T) {
	p := &presenter.EscobaCuiPresenter{}
	e := escobaBuildCuiGame(t)
	if out := p.Output(e, nil); out == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestEscobaCuiPresenter_Error(t *testing.T) {
	p := &presenter.EscobaCuiPresenter{}
	e := escobaBuildCuiGame(t)
	out := p.Output(e, escobaAssertErr{})
	if !strings.Contains(out, "boom") {
		t.Errorf("expected error text in output, got: %s", out)
	}
}

type escobaAssertErr struct{}

func (escobaAssertErr) Error() string { return "boom" }

func TestEscobaCuiPresenter_ActionLogOutput(t *testing.T) {
	p := &presenter.EscobaCuiPresenter{}
	e := escobaBuildCuiGame(t)
	if out := p.ActionLogOutput(e); out == "" {
		t.Error("expected non-empty action log output")
	}
}

func TestEscobaCuiPresenter_OutputGameEnd(t *testing.T) {
	p := &presenter.EscobaCuiPresenter{}
	e := escobaPlayedOutGame(t) // all-CPU game driven to game end (defined in the web presenter test)
	out := p.Output(e, nil)
	if out == "" {
		t.Fatal("expected non-empty output at game end")
	}
}
