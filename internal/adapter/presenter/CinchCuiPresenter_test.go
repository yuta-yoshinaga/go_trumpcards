package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
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

// **なぜ点がそう動いたのかが CUI からは読めなかった。**Web は `lastDealDetail` で
// 内訳を出し、ビッド未達の行を赤字にしているのに、CUI は「ディール終了」の
// 一行だけだった (#6488)。
func TestCinchCuiPresenter_ShowsTheDealBreakdown(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(false) // 強調そのものを見るので無効化しない
	defer color.SetNoColor(orig)
	assert.NotEqual(t, "x", color.Red("x"), "colour must be enabled for this test to measure anything")

	p := new(presenter.CinchCuiPresenter)
	withDetail := func(setBack bool) string {
		g := domain.NewDefaultCinch()
		g.Reset()
		g.SetPhase(domain.CinchPhaseRoundEnd)
		gained := map[int]int{0: 2, 1: 1, 2: 0, 3: 0}
		if setBack {
			gained[0] = -9
		}
		g.SetLastDealDetail(&domain.CinchDealDetail{
			TrumpSuit: domain.CardDesignSpade,
			BidderIdx: 0,
			Bid:       9,
			SetBack:   setBack,
			Points:    map[int]int{0: 4, 1: 3, 2: 2, 3: 0},
			Gained:    gained,
		})
		return p.Output(g, nil)
	}

	t.Run("names the bidder and the points each seat gained", func(t *testing.T) {
		out := withDetail(false)
		assert.Contains(t, out, i18n.T("cinch.dealResultTitle"))
		assert.Contains(t, out, i18n.Tf("cinch.dealResultGained",
			"name", color.Bold(i18n.T("cuiPlayerYou")), "points", "2"))
		assert.NotContains(t, out, "{{")
	})

	// **強調の条件は Web の `isSetBackRow` と同じ** ── ビッダーの行で、かつ
	// セットバックした場合だけ。
	t.Run("marks the bidder red when they were set back", func(t *testing.T) {
		out := withDetail(true)
		bidderRow := i18n.Tf("cinch.dealResultGained",
			"name", color.Bold(i18n.T("cuiPlayerYou")), "points", "-9")
		assert.Contains(t, out, color.Red(bidderRow))
		// 他家の行は赤くしない。
		other := i18n.Tf("cinch.dealResultGained",
			"name", color.Bold(i18n.Tf("cuiPlayerCpu", "idx", "1")), "points", "1")
		assert.Contains(t, out, other)
		assert.NotContains(t, out, color.Red(other))
	})

	t.Run("leaves the bidder plain when the bid was made", func(t *testing.T) {
		out := withDetail(false)
		bidderRow := i18n.Tf("cinch.dealResultGained",
			"name", color.Bold(i18n.T("cuiPlayerYou")), "points", "2")
		assert.NotContains(t, out, color.Red(bidderRow))
	})

	// 内訳が無い局面 (最初のディールの途中など) では何も出さない。
	t.Run("stays quiet without a recorded deal", func(t *testing.T) {
		g := domain.NewDefaultCinch()
		g.Reset()
		g.SetPhase(domain.CinchPhaseRoundEnd)
		assert.NotContains(t, p.Output(g, nil), i18n.T("cinch.dealResultTitle"))
	})
}
