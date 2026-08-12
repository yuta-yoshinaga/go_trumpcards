//go:build test

package presenter

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newCucumberForCui(t *testing.T) *domain.Cucumber {
	t.Helper()
	c := domain.NewDefaultCucumber()
	c.Reset()
	return c
}

func TestCucumberCuiPresenterOutput(t *testing.T) {
	p := new(CucumberCuiPresenter)
	c := newCucumberForCui(t)
	out := p.Output(c, nil)

	assert.Contains(t, out, i18n.T("cucumber.helpTitle"))
	assert.Contains(t, out, fixedPart("cucumber.header"))
	// **スート無関係・失点は最終トリックだけ、が規則そのもの。**
	assert.Contains(t, out, i18n.T("cucumber.rule"))
	assert.Len(t, regexp.MustCompile(`手札\d+枚 失点\d+点`).FindAllString(out, -1),
		domain.CucumberDefaultPlayerCnt, "全員の席行に手札と失点が出る")
	// リードなので基準は出ない。
	assert.NotContains(t, out, fixedPart("cucumber.highest"))
	assert.Contains(t, out, i18n.T("cucumber.promptLead"))
}

// **超える基準を出します。** 盤面から数えさせません。
func TestCucumberCuiPresenterShowsTheThreshold(t *testing.T) {
	p := new(CucumberCuiPresenter)
	c := newCucumberForCui(t)
	c.SetCurrentPlayerIdxForTest(0)
	c.GiveHandForTest(0,
		domain.NewCard(domain.CardDesignSpade, 3, false),
		domain.NewCard(domain.CardDesignHeart, 10, false))
	c.SetCurrentTrickForTest([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignDiamond, 9, false)},
	})

	out := p.Output(c, nil)
	assert.Contains(t, out, fixedPart("cucumber.highest"))
	assert.Contains(t, out, fixedPart("cucumber.promptBeat"))
	assert.NotContains(t, out, i18n.T("cucumber.promptForced"))
	assert.NotContains(t, out, i18n.T("cucumber.promptLead"))
}

// **合法手が 1 つ = 更新できない、ではありません。**
func TestCucumberCuiPresenterDistinguishesForcedFromASingleBeat(t *testing.T) {
	p := new(CucumberCuiPresenter)
	c := newCucumberForCui(t)
	c.SetCurrentPlayerIdxForTest(0)
	c.GiveHandForTest(0,
		domain.NewCard(domain.CardDesignSpade, 3, false),
		domain.NewCard(domain.CardDesignHeart, 5, false))
	c.SetCurrentTrickForTest([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignDiamond, 13, false)},
	})
	assert.Contains(t, p.Output(c, nil), i18n.T("cucumber.promptForced"))

	// **負のコントロール: 1 枚でも更新できるなら forced ではない。**
	c.GiveHandForTest(0,
		domain.NewCard(domain.CardDesignSpade, 3, false),
		domain.NewCard(domain.CardDesignHeart, 14, false))
	out := p.Output(c, nil)
	assert.NotContains(t, out, i18n.T("cucumber.promptForced"))
	assert.Contains(t, out, fixedPart("cucumber.promptBeat"))
}

// **失点はラウンドに 1 回だけの出来事。** 配り直す前に読ませます。
func TestCucumberCuiPresenterShowsTheRoundResult(t *testing.T) {
	p := new(CucumberCuiPresenter)
	c := newCucumberForCui(t)
	for i := range c.GetPlayerCnt() {
		c.GiveHandForTest(i, domain.NewCard(domain.CardDesignSpade, 5+i, false))
	}
	c.SetCurrentPlayerIdxForTest(0)
	c.SetCurrentTrickForTest(nil)
	for i := range c.GetPlayerCnt() {
		require.NoError(t, c.PlayForTest(i, 0))
	}
	require.Equal(t, domain.CucumberPhaseRoundEnd, c.GetPhase())

	out := p.Output(c, nil)
	assert.Contains(t, out, fixedPart("cucumber.promptRoundEnd"))
	assert.Contains(t, out, i18n.T("cucumber.promptNext"))
	assert.Contains(t, out, fixedPart("cucumber.roleLastTrick"))
	assert.NotContains(t, out, i18n.T("cucumber.promptPlay"), "区切りでは出す促しをしない")
}

func TestCucumberCuiPresenterGameEndBanners(t *testing.T) {
	p := new(CucumberCuiPresenter)

	won := newCucumberForCui(t)
	for i := 1; i < won.GetPlayerCnt(); i++ {
		won.GetPlayer(i).SetPenalty(won.GetConfig().TargetScore)
	}
	won.GiveHandForTest(0, domain.NewCard(domain.CardDesignSpade, 2, false))
	for i := 1; i < won.GetPlayerCnt(); i++ {
		won.GiveHandForTest(i, domain.NewCard(domain.CardDesignHeart, 3+i, false))
	}
	won.SetCurrentPlayerIdxForTest(0)
	won.SetCurrentTrickForTest(nil)
	for i := range won.GetPlayerCnt() {
		require.NoError(t, won.PlayForTest(i, 0))
	}
	require.True(t, won.GetGameEndFlag())
	out := p.Output(won, nil)
	assert.Contains(t, out, fixedPart("cucumber.gameEndYou"))
	assert.NotContains(t, out, i18n.T("cucumber.promptPlay"), "終局後は促さない")

	lost := newCucumberForCui(t)
	lost.GiveUp()
	assert.Contains(t, p.Output(lost, nil), fixedPart("cucumber.gameEndCpu"))
}

func TestCucumberCuiPresenterShowsErrors(t *testing.T) {
	p := new(CucumberCuiPresenter)
	assert.Contains(t, p.Output(newCucumberForCui(t), assert.AnError), assert.AnError.Error())
}

func TestCucumberCuiPresenterHintOutput(t *testing.T) {
	p := new(CucumberCuiPresenter)
	c := newCucumberForCui(t)
	c.SetCurrentPlayerIdxForTest(0)
	c.SetCurrentTrickForTest(nil)

	out := p.HintOutput(c)
	assert.Contains(t, out, fixedPart("cucumber.hintCard"))
	assert.Contains(t, out, i18n.T("cucumber.hintReasonLead"))

	c.GiveHandForTest(0,
		domain.NewCard(domain.CardDesignSpade, 3, false),
		domain.NewCard(domain.CardDesignHeart, 10, false))
	c.SetCurrentTrickForTest([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignDiamond, 9, false)},
	})
	assert.Contains(t, p.HintOutput(c), i18n.T("cucumber.hintReasonBeat"))

	c.SetCurrentTrickForTest([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignDiamond, 13, false)},
	})
	assert.Contains(t, p.HintOutput(c), i18n.T("cucumber.hintReasonForced"))

	c.GiveUp()
	assert.Contains(t, p.HintOutput(c), i18n.T("cucumber.hintNone"))
}

func TestCucumberCuiPresenterActionLogOutput(t *testing.T) {
	p := new(CucumberCuiPresenter)
	c := newCucumberForCui(t)
	c.GiveUp()
	assert.NotEmpty(t, p.ActionLogOutput(c))
}
