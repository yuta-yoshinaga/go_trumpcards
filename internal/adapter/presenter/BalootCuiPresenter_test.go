//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newBalootForCui(t *testing.T) *domain.Baloot {
	t.Helper()
	b := domain.NewDefaultBaloot()
	b.Reset()
	return b
}

func TestBalootCuiPresenterOutput(t *testing.T) {
	p := new(BalootCuiPresenter)
	b := newBalootForCui(t)

	out := p.Output(b, nil)
	assert.Contains(t, out, i18n.T("baloot.helpTitle"))
	assert.Contains(t, out, fixedPart("baloot.scoreLine"))
	// 宣言前はモード未定の案内。
	assert.Contains(t, out, i18n.T("baloot.modeUndecided"))
}

// **序列はモードで入れ替わるので、有効な方だけを出す。** 両方出すと
// どちらが効いているのか画面から読めない。
func TestBalootCuiPresenterShowsTheActiveOrderOnly(t *testing.T) {
	p := new(BalootCuiPresenter)

	sun := newBalootForCui(t)
	sun.SetCurrentPlayerIdxForTest(0)
	require.NoError(t, sun.DeclareSun())
	outSun := p.Output(sun, nil)
	assert.Contains(t, outSun, i18n.T("baloot.orderSun"))
	assert.NotContains(t, outSun, i18n.T("baloot.orderHokom"))

	hokom := newBalootForCui(t)
	hokom.SetCurrentPlayerIdxForTest(0)
	require.NoError(t, hokom.DeclareHokom(domain.CardDesignHeart))
	outHokom := p.Output(hokom, nil)
	assert.Contains(t, outHokom, i18n.T("baloot.orderHokom"))
	assert.NotContains(t, outHokom, i18n.T("baloot.orderSun"))
	assert.Contains(t, outHokom, fixedPart("baloot.modeHokomLine"))
}

// **宣言者が居ないのにモードだけ立っている状態でも落ちない。**
func TestBalootCuiPresenterModeWithoutDeclarer(t *testing.T) {
	p := new(BalootCuiPresenter)
	b := newBalootForCui(t)
	b.SetModeForTest(domain.BalootModeSun)
	b.SetDeclarerIdxForTest(-1)

	assert.Contains(t, p.Output(b, nil), i18n.T("baloot.modeUndecided"))
}

// **親のときは見送れないので案内が変わる。** 両側を踏む。
func TestBalootCuiPresenterDeclarePrompts(t *testing.T) {
	p := new(BalootCuiPresenter)

	dealer := newBalootForCui(t)
	dealer.SetDealerIdxForTest(0)
	out := p.Output(dealer, nil)
	assert.Contains(t, out, i18n.T("baloot.promptDeclareDealer"))
	assert.NotContains(t, out, i18n.T("baloot.promptDeclareHelp"))

	other := newBalootForCui(t)
	other.SetDealerIdxForTest(2)
	out2 := p.Output(other, nil)
	assert.Contains(t, out2, i18n.T("baloot.promptDeclareHelp"))
	assert.NotContains(t, out2, i18n.T("baloot.promptDeclareDealer"))
}

// Baloot 役の有無で表示が変わる。両側を踏む。
func TestBalootCuiPresenterBalootBonus(t *testing.T) {
	p := new(BalootCuiPresenter)

	none := newBalootForCui(t)
	for i := range domain.BalootPlayerCnt {
		none.GetPlayer(i).SetHasBaloot(false)
	}
	assert.Contains(t, p.Output(none, nil), i18n.T("baloot.balootNone"))

	held := newBalootForCui(t)
	held.GetPlayer(0).SetHasBaloot(true)
	assert.Contains(t, p.Output(held, nil), fixedPart("baloot.balootHeld"))
}

func TestBalootCuiPresenterRoundEnd(t *testing.T) {
	p := new(BalootCuiPresenter)
	b := newBalootForCui(t)
	b.SetPhaseForTest(domain.BalootPhaseRoundEnd)

	out := p.Output(b, nil)
	assert.Contains(t, out, i18n.T("baloot.promptRoundEnd"))
	assert.Contains(t, out, i18n.T("baloot.promptNext"))
	assert.NotContains(t, out, i18n.T("baloot.promptPlay"))
}

func TestBalootCuiPresenterPlayPrompt(t *testing.T) {
	p := new(BalootCuiPresenter)
	b := newBalootForCui(t)
	b.SetCurrentPlayerIdxForTest(0)
	require.NoError(t, b.DeclareSun())

	out := p.Output(b, nil)
	assert.Contains(t, out, i18n.T("baloot.promptPlay"))
	assert.Contains(t, out, fixedPart("baloot.promptCurrentPlayer"))
}

func TestBalootCuiPresenterError(t *testing.T) {
	p := new(BalootCuiPresenter)
	assert.Contains(t, p.Output(newBalootForCui(t), assert.AnError), assert.AnError.Error())
}

func TestBalootCuiPresenterGameEnd(t *testing.T) {
	p := new(BalootCuiPresenter)

	for _, tc := range []struct {
		name    string
		t0, t1  int
		wantKey string
	}{
		{"your team", 160, 100, "baloot.gameEndTeam0"},
		{"the other team", 100, 160, "baloot.gameEndTeam1"},
		{"a tie", 152, 152, "baloot.gameEndTie"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := newBalootForCui(t)
			b.SetScoreForTestUse(0, tc.t0)
			b.SetScoreForTestUse(1, tc.t1)
			b.FinishGameForTest()

			out := p.Output(b, nil)
			assert.Contains(t, out, fixedPart(tc.wantKey))
			assert.NotContains(t, out, i18n.T("baloot.promptPlay"))
		})
	}
}

// 宣言フェーズのヒントは 3 つの理由キーが全部文言に解決される。
func TestBalootCuiPresenterHintInDeclarePhase(t *testing.T) {
	p := new(BalootCuiPresenter)
	b := newBalootForCui(t)
	b.SetCurrentPlayerIdxForTest(0)

	out := p.HintOutput(b)
	assert.Contains(t, out, "HINT")
	assert.NotContains(t, out, "balootDeclareSun", "生のキーが出ていたら未登録")
	assert.NotContains(t, out, "balootDeclareHokom")
	assert.NotContains(t, out, "balootPassDeclare")
}

// **Hokom を勧めるときはスート名まで出す。** 「1 スートが厚い」だけでは
// どのスートを宣言すればよいか分からない。
func TestBalootCuiPresenterHintNamesTheHokomSuit(t *testing.T) {
	p := new(BalootCuiPresenter)
	b := newBalootForCui(t)
	b.SetCurrentPlayerIdxForTest(0)
	hand := b.GetPlayer(0)
	hand.Reset()
	for _, c := range []*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 11, false),
		domain.NewCard(domain.CardDesignHeart, 9, false),
		domain.NewCard(domain.CardDesignHeart, 1, false),
		domain.NewCard(domain.CardDesignSpade, 7, false),
		domain.NewCard(domain.CardDesignClover, 8, false),
	} {
		hand.AddCard(c)
	}

	out := p.HintOutput(b)
	assert.Contains(t, out, i18n.T("baloot.suitHeart"))
}

// プレイ中の 2 つの理由キーが両方とも文言に解決される。
func TestBalootCuiPresenterHintInPlayPhase(t *testing.T) {
	p := new(BalootCuiPresenter)

	solo := newBalootForCui(t)
	solo.SetCurrentPlayerIdxForTest(0)
	require.NoError(t, solo.DeclareSun())
	solo.SetCurrentPlayerIdxForTest(0)
	assert.Contains(t, p.HintOutput(solo), i18n.T("baloot.hintReasonWinTrick"))

	partner := newBalootForCui(t)
	partner.SetCurrentPlayerIdxForTest(0)
	require.NoError(t, partner.DeclareHokom(domain.CardDesignHeart))
	partner.SetCurrentPlayerIdxForTest(0)
	// 味方 (2番) が切り札の J で勝っている。
	partner.SetCurrentTrickForTest([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 7, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 11, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignSpade, 8, false)},
	})
	assert.Contains(t, p.HintOutput(partner), i18n.T("baloot.hintReasonFeedPartner"))
}

func TestBalootCuiPresenterHintNoneAfterGameEnd(t *testing.T) {
	p := new(BalootCuiPresenter)
	b := newBalootForCui(t)
	b.GiveUp()
	assert.Contains(t, p.HintOutput(b), i18n.T("baloot.hintNone"))
}

// スート名は 4 つとも解決される。
func TestBalootSuitName(t *testing.T) {
	for _, suit := range []int{
		domain.CardDesignSpade, domain.CardDesignClover,
		domain.CardDesignHeart, domain.CardDesignDiamond,
	} {
		assert.NotEqual(t, "?", balootSuitName(suit))
	}
	assert.Equal(t, "?", balootSuitName(domain.CardDesignJoker))
}

func TestBalootCuiPresenterActionLogOutput(t *testing.T) {
	p := new(BalootCuiPresenter)
	b := newBalootForCui(t)
	b.GiveUp()
	require.NotEmpty(t, p.ActionLogOutput(b))
}
