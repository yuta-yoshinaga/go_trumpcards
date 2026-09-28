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

	t.Run("lists every hand card candidate and mandatory capture", func(t *testing.T) {
		s := build(7, domain.CardDesignHeart, []*domain.Card{
			domain.NewCard(domain.CardDesignClover, 3, false),
			domain.NewCard(domain.CardDesignSpade, 4, false),
		})
		s.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		s.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		out := p.HintOutput(s)
		if recommendation, list := strings.Index(out, "推奨:"), strings.Index(out, "取得候補一覧"); recommendation < 0 || list < 0 || recommendation > list {
			t.Errorf("expected recommendation before candidate list, got: %s", out)
		}
		if !strings.Contains(out, "全取り") {
			t.Errorf("expected scopa recommendation to be prioritized, got: %s", out)
		}
		for _, want := range []string{"手札0", "手札1", "手札2", "捕獲必須", "取得候補なし", "場に置けます", "♣3", "♠4"} {
			if !strings.Contains(out, want) {
				t.Errorf("expected candidate hint to contain %q, got: %s", want, out)
			}
		}
	})

	t.Run("no capture lists every hand as unavailable", func(t *testing.T) {
		s := build(2, domain.CardDesignHeart, []*domain.Card{
			domain.NewCard(domain.CardDesignDiamond, 7, false),
			domain.NewCard(domain.CardDesignClover, 5, false),
		})
		s.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		out := p.HintOutput(s)
		if !strings.Contains(out, "捕獲できる手はありません") {
			t.Errorf("expected no-capture recommendation first, got: %s", out)
		}
		if strings.Count(out, "取得候補なし") != 2 {
			t.Errorf("expected no-capture line for both hand cards, got: %s", out)
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
	s.SetPhase(domain.ScoponePhasePlayerTurn)
	out := p.Output(s, nil)
	if out == "" {
		t.Fatal("expected non-empty output")
	}
	const scoreRules = "得点: 最多カード/最多ダイヤ/最多の7/セッテベッロ(7♦)は各1点、スコパは1回1点（最多が同数なら加点なし）"
	if !strings.Contains(out, scoreRules) {
		t.Errorf("expected score rules in prompt, got: %s", out)
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
