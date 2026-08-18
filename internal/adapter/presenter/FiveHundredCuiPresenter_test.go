package presenter_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
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
		g.SetCurrentTrick([]*domain.TrickCard{
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

func TestFiveHundredCuiPresenter_BidLabelsAndHighestBid(t *testing.T) {
	p := &presenter.FiveHundredCuiPresenter{}

	// Player bid labels (suit / NT / misère / open misère) + pass + declarer.
	g := newFiveHundredGame()
	g.GetPlayer(0).SetBid(&domain.FiveHundredBid{Kind: domain.FiveHundredContractSuit, Tricks: 7, Suit: domain.CardDesignSpade})
	g.GetPlayer(1).SetBid(&domain.FiveHundredBid{Kind: domain.FiveHundredContractNoTrump, Tricks: 8})
	g.GetPlayer(2).SetBid(&domain.FiveHundredBid{Kind: domain.FiveHundredContractMisere})
	g.GetPlayer(3).SetBid(&domain.FiveHundredBid{Kind: domain.FiveHundredContractOpenMisere})
	g.SetPhase(domain.FiveHundredPhaseBid)
	if out := p.Output(g, nil); out == "" {
		t.Errorf("bid-label output empty")
	}

	// Highest-bid display (contract undecided, a bid stands).
	g2 := newFiveHundredGame()
	g2.Reset()
	g2.SetBidPlayerIdx(0)
	if err := g2.PlayerBid(domain.FiveHundredContractSuit, 6, domain.CardDesignSpade); err != nil {
		t.Fatalf("bid failed: %v", err)
	}
	if out := p.Output(g2, nil); out == "" {
		t.Errorf("highest-bid output empty")
	}
}

func TestFiveHundredCuiPresenter_HintVariants(t *testing.T) {
	p := &presenter.FiveHundredCuiPresenter{}

	// Pass hint (weak bid hand).
	g := newFiveHundredGame()
	g.SetPhase(domain.FiveHundredPhaseBid)
	g.SetBidPlayerIdx(0)
	for v := 5; v <= 9; v++ {
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, v, false))
	}
	if out := p.HintOutput(g); out == "" {
		t.Errorf("pass hint empty")
	}

	// Bid hint (strong hand).
	g2 := newFiveHundredGame()
	g2.SetPhase(domain.FiveHundredPhaseBid)
	g2.SetBidPlayerIdx(0)
	for _, c := range []*domain.Card{
		domain.NewCard(domain.CardDesignJoker, 1, false),
		domain.NewCard(domain.CardDesignSpade, 11, false),
		domain.NewCard(domain.CardDesignClover, 11, false),
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignSpade, 13, false),
		domain.NewCard(domain.CardDesignSpade, 12, false),
		domain.NewCard(domain.CardDesignSpade, 10, false),
		domain.NewCard(domain.CardDesignSpade, 9, false),
		domain.NewCard(domain.CardDesignHeart, 1, false),
		domain.NewCard(domain.CardDesignDiamond, 1, false),
	} {
		g2.GetPlayer(0).AddCard(c)
	}
	if out := p.HintOutput(g2); out == "" {
		t.Errorf("bid hint empty")
	}

	// Discard hint (declarer in kitty exchange).
	g3 := newFiveHundredGame()
	g3.SetContract(domain.FiveHundredContractSuit, 7, domain.CardDesignSpade)
	g3.SetDeclarerIdx(0)
	g3.SetPhase(domain.FiveHundredPhaseKittyExchange)
	for i := 0; i < 13; i++ {
		g3.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, (i%9)+5, false))
	}
	if out := p.HintOutput(g3); out == "" {
		t.Errorf("discard hint empty")
	}
}

// #5626: ノートランプでジョーカーをリードすると、リーダーが追走スートを指名する。
// ドメインは `GetJokerLeadSuit()` で保持し presenter (Web) もレスポンスに載せて
// いたが、**どちらの画面もその値を出していなかった**ので、他の3人は「いま何に
// 従うのか」を確認できなかった。
func TestFiveHundredCuiPresenterShowsTheNominatedSuit(t *testing.T) {
	p := &presenter.FiveHundredCuiPresenter{}

	g := newFiveHundredGame()
	g.SetContract(domain.FiveHundredContractNoTrump, 7, -1)
	g.SetPhase(domain.FiveHundredPhasePlay)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignJoker, 0, false)},
	})
	g.SetJokerLeadSuit(domain.CardDesignHeart)

	out := p.Output(g, nil)
	assert.Contains(t, out, i18n.Tf("fivehundred.jokerLeadSuit", "suit", "HEART"))
}

// 指名が無いとき (通常のリード) は行そのものを出さない。
func TestFiveHundredCuiPresenterOmitsTheNominatedSuitWhenThereIsNone(t *testing.T) {
	p := &presenter.FiveHundredCuiPresenter{}

	g := newFiveHundredGame()
	g.SetContract(domain.FiveHundredContractNoTrump, 7, -1)
	g.SetPhase(domain.FiveHundredPhasePlay)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
	})

	out := p.Output(g, nil)
	assert.NotContains(t, out, i18n.Tf("fivehundred.jokerLeadSuit", "suit", "HEART"))
}
