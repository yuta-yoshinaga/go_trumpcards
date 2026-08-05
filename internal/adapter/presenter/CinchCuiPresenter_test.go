package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestCinchCuiPresenter_Output_BidPhase(t *testing.T) {
	g := domain.NewDefaultCinch()
	g.Reset()
	p := new(presenter.CinchCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "[0]") // human indexed hand
}

func TestCinchCuiPresenter_Output_AllPhases(t *testing.T) {
	p := new(presenter.CinchCuiPresenter)

	g := domain.NewDefaultCinch()
	g.Reset()

	g.SetPhase(domain.CinchPhaseNameTrump)
	g.SetBidWinnerIdx(0)
	assert.NotEmpty(t, p.Output(g, nil))

	g.SetPhase(domain.CinchPhasePlay)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetCurrentTurn(0)
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: bcard(domain.CardDesignHeart, 9)}})
	assert.NotEmpty(t, p.Output(g, nil))

	g.SetPhase(domain.CinchPhaseTrickEnd)
	assert.NotEmpty(t, p.Output(g, nil))

	g.SetPhase(domain.CinchPhaseRoundEnd)
	assert.NotEmpty(t, p.Output(g, nil))

	// エラー出力。
	assert.NotEmpty(t, p.Output(g, errors.New("boom")))
}

func TestCinchCuiPresenter_GameEnd(t *testing.T) {
	g := domain.NewDefaultCinch()
	g.Reset()
	g.GetPlayer(0).AddScore(30)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetBidWinnerIdx(0)
	g.SetCurrentBid(1)
	g.SetPhase(domain.CinchPhaseRoundEnd)
	g.ScoreRound()

	p := new(presenter.CinchCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
}

func TestCinchCuiPresenter_HintOutput(t *testing.T) {
	p := new(presenter.CinchCuiPresenter)

	t.Run("bid hint", func(t *testing.T) {
		g := domain.NewDefaultCinch()
		g.Reset() // bid phase, human turn
		assert.NotEmpty(t, p.HintOutput(g))
	})

	t.Run("name trump hint", func(t *testing.T) {
		g := domain.NewDefaultCinch()
		g.Reset()
		g.SetPhase(domain.CinchPhaseNameTrump)
		g.SetBidWinnerIdx(0)
		assert.NotEmpty(t, p.HintOutput(g))
	})

	t.Run("play hint", func(t *testing.T) {
		g := domain.NewDefaultCinch()
		g.Reset()
		g.SetPhase(domain.CinchPhasePlay)
		g.SetTrumpSuit(domain.CardDesignHeart)
		g.SetCurrentTurn(0)
		g.SetLeadPlayerIdx(0)
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(bcard(domain.CardDesignHeart, 1))
		g.GetPlayer(0).AddCard(bcard(domain.CardDesignSpade, 2))
		assert.NotEmpty(t, p.HintOutput(g))
	})

	t.Run("no hint outside human turn", func(t *testing.T) {
		g := domain.NewDefaultCinch()
		g.Reset()
		g.SetPhase(domain.CinchPhaseTrickEnd)
		assert.NotEmpty(t, p.HintOutput(g)) // hintNone message
	})
}

func TestCinchCuiPresenter_ActionLog(t *testing.T) {
	g := domain.NewDefaultCinch()
	g.Reset()
	p := new(presenter.CinchCuiPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}

// **ビッドの定量的な材料。**Web は estimateCinchBidStrength でスート別の保持点と
// 最有力スートをビッドフェーズ中ずっと出しているのに、CUI は最高ビッド額しか
// 出していなかった (#4845)。
func TestCinchCuiPresenter_BidStrengthLine(t *testing.T) {
	// ♣ が最有力になる手札: ♣A(1) ♣5(Right Pedro=5) ♠5(Left Pedro=5) -> ♣ で 11 点。
	cards := []*domain.Card{
		domain.NewCard(domain.CardDesignClover, 1, false),
		domain.NewCard(domain.CardDesignClover, 5, false),
		domain.NewCard(domain.CardDesignSpade, 5, false),
	}
	points := domain.CinchHandPointsBySuit(cards)
	// ♣ 切り札: ♣A=1 + ♣5(Right)=5 + ♠5(Left)=5 = 11。
	assert.Equal(t, 11, points[domain.CardDesignClover])
	// ♠ 切り札: ♠5(Right)=5 + ♣5(Left)=5 = 10。♣A は点にならない。
	assert.Equal(t, 10, points[domain.CardDesignSpade])
	assert.Equal(t, 0, points[domain.CardDesignHeart])
	assert.Equal(t, domain.CardDesignClover, domain.CinchBestTrumpSuit(points))

	// 同点は小さいスート番号が勝つ (Web の estimateCinchBidStrength と同じ)。
	tie := domain.CinchHandPointsBySuit([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 1, false),
		domain.NewCard(domain.CardDesignDiamond, 1, false),
	})
	assert.Equal(t, tie[domain.CardDesignHeart], tie[domain.CardDesignDiamond])
	assert.Equal(t, domain.CardDesignHeart, domain.CinchBestTrumpSuit(tie))

	g := domain.NewDefaultCinch()
	g.Reset()
	out := new(presenter.CinchCuiPresenter).Output(g, nil)
	assert.Contains(t, out, "手札の点数:", "ビッドフェーズには常時出る")
	assert.Contains(t, out, "最有力:")

	// 他のフェーズには出さない。
	g.SetPhase(domain.CinchPhasePlay)
	assert.NotContains(t, new(presenter.CinchCuiPresenter).Output(g, nil), "手札の点数:")

	// 手札が無いとき (配り直し前など) も出さない。0 点の表を出しても意味が無い。
	g2 := domain.NewDefaultCinch()
	g2.Reset()
	for i := range g2.GetPlayerCnt() {
		if p := g2.GetPlayer(i); p != nil && p.GetIsHuman() {
			p.ResetDeal()
		}
	}
	assert.NotContains(t, new(presenter.CinchCuiPresenter).Output(g2, nil), "手札の点数:")
}
