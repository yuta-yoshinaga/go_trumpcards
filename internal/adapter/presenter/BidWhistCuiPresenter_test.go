package presenter_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestBidWhistCuiPresenter_Output_Phases(t *testing.T) {
	p := &presenter.BidWhistCuiPresenter{}

	t.Run("bid phase", func(t *testing.T) {
		g := newBidWhistGame()
		g.Reset()
		out := p.Output(g, nil)
		if out == "" || !strings.Contains(out, "Bid Whist") {
			t.Errorf("bid output missing title: %q", out)
		}
	})

	t.Run("trump declaration phase", func(t *testing.T) {
		g := newBidWhistGame()
		g.SetContract(3, domain.BidWhistDirectionUptown, -1)
		g.SetDeclarerIdx(0)
		g.GetPlayer(0).SetIsDeclarer(true)
		g.SetPhase(domain.BidWhistPhaseTrumpDeclaration)
		if out := p.Output(g, nil); out == "" {
			t.Errorf("trump declaration output empty")
		}
	})

	t.Run("kitty exchange phase", func(t *testing.T) {
		g := newBidWhistGame()
		g.SetContract(3, domain.BidWhistDirectionUptown, domain.CardDesignSpade)
		g.SetTrumpSuit(domain.CardDesignSpade)
		g.SetDeclarerIdx(0)
		g.GetPlayer(0).SetIsDeclarer(true)
		g.SetPhase(domain.BidWhistPhaseKittyExchange)
		if out := p.Output(g, nil); out == "" {
			t.Errorf("kitty exchange output empty")
		}
	})

	t.Run("play phase with trick (no trump)", func(t *testing.T) {
		g := newBidWhistGame()
		g.SetContract(3, domain.BidWhistDirectionNoTrump, -1)
		g.SetDeclarerIdx(0)
		g.SetPhase(domain.BidWhistPhasePlay)
		g.SetCurrentTrick([]*domain.TrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 1, false)},
		})
		if out := p.Output(g, nil); out == "" {
			t.Errorf("play output empty")
		}
	})

	t.Run("game end", func(t *testing.T) {
		g := newBidWhistGame()
		g.SetTeamScore(0, 6)
		g.SetContract(1, domain.BidWhistDirectionUptown, domain.CardDesignSpade)
		g.SetDeclarerIdx(0)
		for i := 0; i < 12; i++ {
			g.GetPlayer(0).AddTrick([]*domain.Card{nil})
		}
		g.SetPhase(domain.BidWhistPhaseRoundEnd)
		g.ScoreRound()
		out := p.Output(g, nil)
		if !strings.Contains(out, "0") {
			t.Errorf("game end banner missing winner: %q", out)
		}
	})
}

func TestBidWhistCuiPresenter_HintOutput(t *testing.T) {
	p := &presenter.BidWhistCuiPresenter{}

	t.Run("play hint", func(t *testing.T) {
		g := newBidWhistGame()
		g.SetContract(3, domain.BidWhistDirectionUptown, domain.CardDesignSpade)
		g.SetTrumpSuit(domain.CardDesignSpade)
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		g.SetPhase(domain.BidWhistPhasePlay)
		g.SetCurrentPlayerIdx(0)
		if out := p.HintOutput(g); out == "" {
			t.Errorf("play hint empty")
		}
	})

	t.Run("no hint when not human turn", func(t *testing.T) {
		g := newBidWhistGame()
		g.SetContract(3, domain.BidWhistDirectionUptown, domain.CardDesignSpade)
		g.SetTrumpSuit(domain.CardDesignSpade)
		g.SetPhase(domain.BidWhistPhasePlay)
		g.SetCurrentPlayerIdx(1) // CPU turn
		out := p.HintOutput(g)
		if !strings.Contains(out, "ヒント") {
			t.Errorf("expected 'no hint' message, got %q", out)
		}
	})
}

func TestBidWhistCuiPresenter_ActionLog(t *testing.T) {
	g := newBidWhistGame()
	g.Reset()
	p := &presenter.BidWhistCuiPresenter{}
	if out := p.ActionLogOutput(g); out == "" {
		t.Errorf("action log empty")
	}
}

func TestBidWhistCuiPresenter_BidLabelsAndHints(t *testing.T) {
	p := &presenter.BidWhistCuiPresenter{}

	// Player bid labels (each direction) + declarer mark, shown in the bid phase.
	g := newBidWhistGame()
	g.GetPlayer(0).SetBid(&domain.BidWhistBid{Tricks: 4, Direction: domain.BidWhistDirectionUptown})
	g.GetPlayer(1).SetBid(&domain.BidWhistBid{Tricks: 5, Direction: domain.BidWhistDirectionDowntown})
	g.GetPlayer(2).SetBid(&domain.BidWhistBid{Tricks: 6, Direction: domain.BidWhistDirectionNoTrump})
	g.GetPlayer(3).SetPassed(true)
	g.SetPhase(domain.BidWhistPhaseBid)
	if out := p.Output(g, nil); out == "" {
		t.Errorf("bid-label output empty")
	}

	// Trump declaration hint (declarer's turn).
	g2 := newBidWhistGame()
	g2.SetDeclarerIdx(0)
	g2.SetContract(3, domain.BidWhistDirectionUptown, -1)
	for v := 9; v <= 13; v++ {
		g2.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
	}
	g2.SetPhase(domain.BidWhistPhaseTrumpDeclaration)
	if out := p.HintOutput(g2); out == "" {
		t.Errorf("trump hint empty")
	}

	// Discard hint (declarer in kitty exchange).
	g3 := newBidWhistGame()
	g3.SetContract(3, domain.BidWhistDirectionUptown, domain.CardDesignSpade)
	g3.SetTrumpSuit(domain.CardDesignSpade)
	g3.SetDeclarerIdx(0)
	g3.SetPhase(domain.BidWhistPhaseKittyExchange)
	for i := 0; i < 18; i++ {
		g3.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, (i%12)+2, false))
	}
	if out := p.HintOutput(g3); out == "" {
		t.Errorf("discard hint empty")
	}
}

// #5632: キティ交換では落札者の手札に山札由来の6枚が混ざる。Web はバッジで
// 区別しているのに、CUI は素の番号付き一覧だけで、どの6枚が入ってきたのかを
// 記憶するしかなかった。
func TestBidWhistCuiPresenterMarksTheKittyCards(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.BidWhistCuiPresenter)

	setup := func(t *testing.T) *domain.BidWhist {
		t.Helper()
		g := newBidWhistGame()
		g.Reset() // 手札を配る
		g.SetContract(3, domain.BidWhistDirectionUptown, domain.CardDesignSpade)
		g.SetTrumpSuit(domain.CardDesignSpade)
		g.SetDeclarerIdx(0)
		g.GetPlayer(0).SetIsDeclarer(true)
		g.SetPhase(domain.BidWhistPhaseKittyExchange)
		// 手札の先頭2枚がキティ由来だったことにする。
		human := g.GetPlayer(0)
		require.Positive(t, human.GetCardsSize(), "前提: 手札が配られている")
		g.SetDeclarerKitty([]*domain.Card{human.GetCard(0), human.GetCard(1)})
		return g
	}

	t.Run("marks exactly the cards that came from the kitty", func(t *testing.T) {
		g := setup(t)
		want := g.GetKittyIndices()
		require.Len(t, want, 2, "前提: 2枚がキティ由来")

		out := p.Output(g, nil)
		assert.Equal(t, len(want), strings.Count(out, presenter.CuiKittyMark),
			"印の数がキティ由来の枚数と一致する")
	})

	// **合法手の印とは別の記号。**同じ画面で意味の違う印が同じ形だと、
	// どちらの話をしているのか読めない。
	t.Run("uses a different mark from the legal-play one", func(t *testing.T) {
		assert.NotEqual(t, presenter.CuiLegalMark, presenter.CuiKittyMark)
	})

	// 交換フェーズ以外では印を出さない (キティはもう手札の一部)。
	t.Run("marks nothing outside the exchange phase", func(t *testing.T) {
		g := setup(t)
		g.SetPhase(domain.BidWhistPhasePlay)

		assert.NotContains(t, p.Output(g, nil), presenter.CuiKittyMark)
	})
}
