//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newShelemForCui(t *testing.T) *domain.Shelem {
	t.Helper()
	s := domain.NewDefaultShelem()
	s.Reset()
	return s
}

func TestShelemCuiPresenterOutput(t *testing.T) {
	p := new(ShelemCuiPresenter)
	s := newShelemForCui(t)

	out := p.Output(s, nil)
	assert.Contains(t, out, i18n.T("shelem.helpTitle"))
	// **点になるのは A/10/5 だけ。** 盤面から読めないので常時出す。
	assert.Contains(t, out, i18n.T("shelem.pointTable"))
	assert.Contains(t, out, i18n.T("shelem.contractUndecided"))
	assert.Contains(t, out, fixedPart("shelem.widowLine"))
}

// 契約は未定・通常・Shelem の 3 通りを踏む。
func TestShelemCuiPresenterContractLine(t *testing.T) {
	p := new(ShelemCuiPresenter)

	normal := newShelemForCui(t)
	normal.SetContractForTest(1, 90, false)
	out := p.Output(normal, nil)
	assert.Contains(t, out, fixedPart("shelem.contractLine"))
	assert.NotContains(t, out, i18n.T("shelem.contractUndecided"))

	shelem := newShelemForCui(t)
	shelem.SetContractForTest(1, domain.ShelemMaxBid, true)
	assert.Contains(t, p.Output(shelem, nil), fixedPart("shelem.contractShelem"))
}

// 競りでの立場が席ごとに出る。5 種すべて踏む。
func TestShelemCuiPresenterRoles(t *testing.T) {
	p := new(ShelemCuiPresenter)
	s := newShelemForCui(t)
	s.SetContractForTest(0, 90, false)
	s.GetPlayer(0).SetBid(90)
	s.GetPlayer(1).SetPassed(true)
	s.GetPlayer(2).SetBid(80)

	out := p.Output(s, nil)
	assert.Contains(t, out, fixedPart("shelem.roleDeclarer"))
	assert.Contains(t, out, i18n.T("shelem.rolePassed"))
	assert.Contains(t, out, fixedPart("shelem.roleBid"))
	assert.Contains(t, out, i18n.T("shelem.roleActive"))

	declared := newShelemForCui(t)
	declared.SetContractForTest(0, domain.ShelemMaxBid, true)
	declared.GetPlayer(0).SetDeclaredShelem(true)
	assert.Contains(t, p.Output(declared, nil), i18n.T("shelem.roleShelem"))
}

// **入札できる番と待つ番で案内が変わる。** 両側を踏む。
func TestShelemCuiPresenterBidPrompts(t *testing.T) {
	p := new(ShelemCuiPresenter)

	mine := newShelemForCui(t)
	mine.SetBidPlayerIdxForTest(0)
	assert.Contains(t, p.Output(mine, nil), fixedPart("shelem.promptBid"))

	theirs := newShelemForCui(t)
	theirs.SetBidPlayerIdxForTest(2)
	out := p.Output(theirs, nil)
	assert.Contains(t, out, i18n.T("shelem.promptBidWait"))
	assert.NotContains(t, out, fixedPart("shelem.promptBid"))
}

// **捨て札は落札者だけ。** 両側を踏む。
func TestShelemCuiPresenterDiscardPrompts(t *testing.T) {
	p := new(ShelemCuiPresenter)

	mine := newShelemForCui(t)
	mine.SetContractForTest(0, 90, false)
	mine.SetPhaseForTest(domain.ShelemPhaseDiscard)
	assert.Contains(t, p.Output(mine, nil), fixedPart("shelem.promptDiscard"))

	theirs := newShelemForCui(t)
	theirs.SetContractForTest(2, 90, false)
	theirs.SetPhaseForTest(domain.ShelemPhaseDiscard)
	assert.Contains(t, p.Output(theirs, nil), i18n.T("shelem.promptDiscardWait"))
}

func TestShelemCuiPresenterRoundEnd(t *testing.T) {
	p := new(ShelemCuiPresenter)
	s := newShelemForCui(t)
	s.SetPhaseForTest(domain.ShelemPhaseRoundEnd)

	out := p.Output(s, nil)
	assert.Contains(t, out, i18n.T("shelem.promptRoundEnd"))
	assert.Contains(t, out, i18n.T("shelem.promptNext"))
	assert.NotContains(t, out, i18n.T("shelem.promptPlay"))
}

func TestShelemCuiPresenterPlayPrompt(t *testing.T) {
	p := new(ShelemCuiPresenter)
	s := newShelemForCui(t)
	s.SetPhaseForTest(domain.ShelemPhasePlay)
	s.SetTrumpSuitForTest(domain.CardDesignHeart)
	s.SetCurrentPlayerIdxForTest(0)

	out := p.Output(s, nil)
	assert.Contains(t, out, i18n.T("shelem.promptPlay"))
	assert.Contains(t, out, fixedPart("shelem.trumpLine"))
}

func TestShelemCuiPresenterError(t *testing.T) {
	p := new(ShelemCuiPresenter)
	assert.Contains(t, p.Output(newShelemForCui(t), assert.AnError), assert.AnError.Error())
}

func TestShelemCuiPresenterGameEnd(t *testing.T) {
	p := new(ShelemCuiPresenter)

	for _, tc := range []struct {
		name    string
		t0, t1  int
		wantKey string
	}{
		{"your team", 520, 300, "shelem.gameEndTeam0"},
		{"the other team", 300, 520, "shelem.gameEndTeam1"},
		{"a tie", 500, 500, "shelem.gameEndTie"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newShelemForCui(t)
			s.SetScoreForTestUse(0, tc.t0)
			s.SetScoreForTestUse(1, tc.t1)
			s.FinishGameForTest()

			out := p.Output(s, nil)
			assert.Contains(t, out, fixedPart(tc.wantKey))
			assert.NotContains(t, out, i18n.T("shelem.promptPlay"))
		})
	}
}

// 競りのヒントは点数まで出す。
func TestShelemCuiPresenterHintDuringBidding(t *testing.T) {
	p := new(ShelemCuiPresenter)
	s := newShelemForCui(t)
	s.SetBidPlayerIdxForTest(0)

	out := p.HintOutput(s)
	assert.Contains(t, out, "HINT")
	assert.NotContains(t, out, "shelemBid", "生のキーが出ていたら未登録")
	assert.NotContains(t, out, "shelemPass")
}

// 捨て札のヒントはスート名まで出す。
func TestShelemCuiPresenterHintDuringDiscard(t *testing.T) {
	p := new(ShelemCuiPresenter)
	s := newShelemForCui(t)
	s.SetContractForTest(0, 90, false)
	s.CloseBiddingForTest()

	out := p.HintOutput(s)
	assert.Contains(t, out, "HINT")
	assert.NotContains(t, out, "shelemDiscard")
}

// プレイ中の 2 つの理由キーが両方とも文言に解決される。
func TestShelemCuiPresenterHintDuringPlay(t *testing.T) {
	p := new(ShelemCuiPresenter)

	solo := newShelemForCui(t)
	solo.SetContractForTest(0, 90, false)
	solo.CloseBiddingForTest()
	require.NoError(t, solo.PlayerDiscard([]int{0, 1, 2, 3}, domain.CardDesignHeart))
	solo.SetCurrentPlayerIdxForTest(0)
	assert.Contains(t, p.HintOutput(solo), i18n.T("shelem.hintReasonWinTrick"))

	partner := newShelemForCui(t)
	partner.SetContractForTest(0, 90, false)
	partner.CloseBiddingForTest()
	require.NoError(t, partner.PlayerDiscard([]int{0, 1, 2, 3}, domain.CardDesignHeart))
	partner.SetCurrentPlayerIdxForTest(0)
	// 味方 (2番) が切り札の A で勝っている。
	partner.SetCurrentTrickForTest([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 7, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 1, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignSpade, 8, false)},
	})
	assert.Contains(t, p.HintOutput(partner), i18n.T("shelem.hintReasonFeedPartner"))
}

func TestShelemCuiPresenterHintNoneAfterGameEnd(t *testing.T) {
	p := new(ShelemCuiPresenter)
	s := newShelemForCui(t)
	s.GiveUp()
	assert.Contains(t, p.HintOutput(s), i18n.T("shelem.hintNone"))
}

// スート名は 4 つとも解決される。
func TestShelemSuitName(t *testing.T) {
	for _, suit := range []int{
		domain.CardDesignSpade, domain.CardDesignClover,
		domain.CardDesignHeart, domain.CardDesignDiamond,
	} {
		assert.NotEqual(t, "?", shelemSuitName(suit))
	}
	assert.Equal(t, "?", shelemSuitName(domain.CardDesignJoker))
}

func TestShelemCuiPresenterActionLogOutput(t *testing.T) {
	p := new(ShelemCuiPresenter)
	s := newShelemForCui(t)
	s.GiveUp()
	require.NotEmpty(t, p.ActionLogOutput(s))
}
