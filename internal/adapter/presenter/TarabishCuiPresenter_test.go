//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newTarabishForCui(t *testing.T) *domain.Tarabish {
	t.Helper()
	tb := domain.NewDefaultTarabish()
	tb.Reset()
	return tb
}

func TestTarabishCuiPresenterOutput(t *testing.T) {
	p := new(TarabishCuiPresenter)
	tb := newTarabishForCui(t)

	out := p.Output(tb, nil)
	assert.Contains(t, out, i18n.T("tarabish.helpTitle"))
	// **切り札の序列はこの系統の肝。** 盤面から読めないので常時出す。
	assert.Contains(t, out, i18n.T("tarabish.orderLine"))
	assert.Contains(t, out, fixedPart("tarabish.scoreLine"))
	// 入札前なので切り札候補が出る。
	assert.Contains(t, out, fixedPart("tarabish.upCardLine"))
}

// **親のときは見送れないので案内が変わる。** 両側を踏む。
func TestTarabishCuiPresenterBidPrompts(t *testing.T) {
	p := new(TarabishCuiPresenter)

	dealer := newTarabishForCui(t)
	dealer.SetDealerIdxForTest(0)
	out := p.Output(dealer, nil)
	assert.Contains(t, out, i18n.T("tarabish.promptBidDealer"))
	assert.NotContains(t, out, i18n.T("tarabish.promptBidHelp"))

	other := newTarabishForCui(t)
	other.SetDealerIdxForTest(2)
	out2 := p.Output(other, nil)
	assert.Contains(t, out2, i18n.T("tarabish.promptBidHelp"))
	assert.NotContains(t, out2, i18n.T("tarabish.promptBidDealer"))
}

// 切り札が決まると候補表示から切り札表示に変わる。
func TestTarabishCuiPresenterShowsTrumpOnceTaken(t *testing.T) {
	p := new(TarabishCuiPresenter)
	tb := newTarabishForCui(t)
	tb.SetCurrentPlayerIdxForTest(0)
	require.NoError(t, tb.TakeTrump())

	out := p.Output(tb, nil)
	assert.Contains(t, out, fixedPart("tarabish.trumpLine"))
	assert.NotContains(t, out, fixedPart("tarabish.upCardLine"))
}

// メルドの内訳が出る。無い場合との両側を踏む。
func TestTarabishCuiPresenterMeldSummary(t *testing.T) {
	p := new(TarabishCuiPresenter)

	none := newTarabishForCui(t)
	for i := range domain.TarabishPlayerCnt {
		none.GetPlayer(i).SetMeldPoints(0)
	}
	assert.Contains(t, p.Output(none, nil), i18n.T("tarabish.meldNone"))

	some := newTarabishForCui(t)
	some.GetPlayer(0).SetMeldPoints(70)
	some.GetPlayer(0).SetRunLen(4)
	some.GetPlayer(0).SetHasBella(true)
	out := p.Output(some, nil)
	assert.Contains(t, out, i18n.T("tarabish.meldBella"))
	assert.Contains(t, out, fixedPart("tarabish.meldRun"))
}

func TestTarabishCuiPresenterRoundEnd(t *testing.T) {
	p := new(TarabishCuiPresenter)
	tb := newTarabishForCui(t)
	tb.SetPhaseForTest(domain.TarabishPhaseRoundEnd)

	out := p.Output(tb, nil)
	assert.Contains(t, out, i18n.T("tarabish.promptRoundEnd"))
	assert.Contains(t, out, i18n.T("tarabish.promptNext"))
	assert.NotContains(t, out, i18n.T("tarabish.promptPlay"))
}

func TestTarabishCuiPresenterError(t *testing.T) {
	p := new(TarabishCuiPresenter)
	assert.Contains(t, p.Output(newTarabishForCui(t), assert.AnError), assert.AnError.Error())
}

func TestTarabishCuiPresenterGameEnd(t *testing.T) {
	p := new(TarabishCuiPresenter)

	for _, tc := range []struct {
		name    string
		t0, t1  int
		wantKey string
	}{
		{"your team", 520, 300, "tarabish.gameEndTeam0"},
		{"the other team", 300, 520, "tarabish.gameEndTeam1"},
		{"a tie", 500, 500, "tarabish.gameEndTie"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tb := newTarabishForCui(t)
			tb.SetScoreForTestUse(0, tc.t0)
			tb.SetScoreForTestUse(1, tc.t1)
			tb.FinishGameForTest()

			out := p.Output(tb, nil)
			assert.Contains(t, out, fixedPart(tc.wantKey))
			assert.NotContains(t, out, i18n.T("tarabish.promptPlay"))
		})
	}
}

func TestTarabishCuiPresenterHintInBidPhase(t *testing.T) {
	p := new(TarabishCuiPresenter)
	tb := newTarabishForCui(t)
	tb.SetCurrentPlayerIdxForTest(0)

	out := p.HintOutput(tb)
	assert.Contains(t, out, "HINT")
	assert.NotContains(t, out, "tarabishTakeTrump", "生のキーが出ていたら未登録")
	assert.NotContains(t, out, "tarabishPassTrump")
}

// プレイ中の 2 つの理由キーが両方とも文言に解決される。
func TestTarabishCuiPresenterHintInPlayPhase(t *testing.T) {
	p := new(TarabishCuiPresenter)

	solo := newTarabishForCui(t)
	solo.SetCurrentPlayerIdxForTest(0)
	require.NoError(t, solo.TakeTrump())
	solo.SetCurrentPlayerIdxForTest(0)
	assert.Contains(t, p.HintOutput(solo), i18n.T("tarabish.hintReasonWinTrick"))

	partner := newTarabishForCui(t)
	partner.SetCurrentPlayerIdxForTest(0)
	require.NoError(t, partner.TakeTrump())
	partner.SetCurrentPlayerIdxForTest(0)
	// 味方 (2番) が切り札の J で勝っている。
	partner.SetCurrentTrickForTest([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 6, false)},
		{PlayerIdx: 2, Card: domain.NewCard(partner.GetTrumpSuit(), 11, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignSpade, 7, false)},
	})
	assert.Contains(t, p.HintOutput(partner), i18n.T("tarabish.hintReasonFeedPartner"))
}

func TestTarabishCuiPresenterHintNoneAfterGameEnd(t *testing.T) {
	p := new(TarabishCuiPresenter)
	tb := newTarabishForCui(t)
	tb.GiveUp()
	assert.Contains(t, p.HintOutput(tb), i18n.T("tarabish.hintNone"))
}

// スート名は 4 つとも解決される。
func TestTarabishSuitName(t *testing.T) {
	for _, suit := range []int{
		domain.CardDesignSpade, domain.CardDesignClover,
		domain.CardDesignHeart, domain.CardDesignDiamond,
	} {
		assert.NotEqual(t, "?", tarabishSuitName(suit))
	}
	assert.Equal(t, "?", tarabishSuitName(domain.CardDesignJoker))
}

func TestTarabishCuiPresenterActionLogOutput(t *testing.T) {
	p := new(TarabishCuiPresenter)
	tb := newTarabishForCui(t)
	tb.GiveUp()
	require.NotEmpty(t, p.ActionLogOutput(tb))
}
