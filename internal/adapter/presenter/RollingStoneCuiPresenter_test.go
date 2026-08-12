//go:build test

package presenter

import (
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newRollingStoneForCui(t *testing.T) *domain.RollingStone {
	t.Helper()
	r := domain.NewDefaultRollingStone()
	r.Reset()
	return r
}

func TestRollingStoneCuiPresenterOutput(t *testing.T) {
	p := new(RollingStoneCuiPresenter)
	out := p.Output(newRollingStoneForCui(t), nil)

	assert.Contains(t, out, i18n.T("rollingstone.helpTitle"))
	assert.Contains(t, out, fixedPart("rollingstone.header"))
	// **勝利条件が逆さまなのが規則そのもの。** 毎回書く。
	assert.Contains(t, out, i18n.T("rollingstone.rule"))
	// 「手札」はルール行にも出るので、席行の並び（手札N枚 引き取りN回）で数える。
	assert.Len(t, regexp.MustCompile(`手札\d+枚 引き取り\d+回`).FindAllString(out, -1),
		domain.RollingStoneDefaultPlayerCnt, "全員の席行に手札と引き取り回数が出る")
	// **デッキ枚数は人数で変わる。**
	assert.Contains(t, out, strconv.Itoa(domain.RollingStoneDeckSize(domain.RollingStoneDefaultPlayerCnt)))
}

// **引き取った席と上がった席を印で出す。** どちらも盤面に痕跡が残らない。
func TestRollingStoneCuiPresenterMarksPickupsAndFinishers(t *testing.T) {
	p := new(RollingStoneCuiPresenter)
	r := newRollingStoneForCui(t)
	assert.NotContains(t, p.Output(r, nil), i18n.T("rollingstone.rolePickedUp"))

	r.SetLeadPlayerIdxForTest(0)
	r.SetCurrentPlayerIdxForTest(0)
	r.GiveHandForTest(0, domain.NewCard(domain.CardDesignSpade, 8, false),
		domain.NewCard(domain.CardDesignSpade, 9, false))
	r.GiveHandForTest(1, domain.NewCard(domain.CardDesignHeart, 8, false))
	require.NoError(t, r.PlayForTest(0, 0))
	require.NoError(t, r.PickUpForTest(1))
	assert.Contains(t, p.Output(r, nil), i18n.T("rollingstone.rolePickedUp"))

	// 上がった席には順位が付く。
	fin := newRollingStoneForCui(t)
	fin.GetPlayer(2).SetFinishedAt(1)
	assert.Contains(t, p.Output(fin, nil), fixedPart("rollingstone.roleFinished"))
}

// **出せる札が無い局面はそう名乗る。** 黙っていると打てない理由が分からない。
func TestRollingStoneCuiPresenterPromptsForAPickUp(t *testing.T) {
	p := new(RollingStoneCuiPresenter)
	r := newRollingStoneForCui(t)
	r.SetLeadPlayerIdxForTest(1)
	r.SetCurrentPlayerIdxForTest(0)
	r.GiveHandForTest(0, domain.NewCard(domain.CardDesignHeart, 8, false))
	r.SetCurrentTrickForTest([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 9, false)},
	})

	out := p.Output(r, nil)
	assert.Contains(t, out, fixedPart("rollingstone.promptPickUp"))
	assert.NotContains(t, out, i18n.T("rollingstone.promptPlay"))

	// **負のコントロール: フォローできるなら通常の促し。**
	r.GiveHandForTest(0, domain.NewCard(domain.CardDesignSpade, 8, false))
	out = p.Output(r, nil)
	assert.Contains(t, out, i18n.T("rollingstone.promptPlay"))
	assert.NotContains(t, out, fixedPart("rollingstone.promptPickUp"))
}

func TestRollingStoneCuiPresenterGameEndBanners(t *testing.T) {
	p := new(RollingStoneCuiPresenter)

	// 上がりで決着。
	won := newRollingStoneForCui(t)
	for i := range won.GetPlayerCnt() {
		won.GiveHandForTest(i, domain.NewCard(domain.CardDesignSpade, 7+i, false))
	}
	won.SetLeadPlayerIdxForTest(0)
	won.SetCurrentPlayerIdxForTest(0)
	for i := range won.GetPlayerCnt() {
		if won.GetGameEndFlag() {
			break
		}
		require.NoError(t, won.PlayForTest(i, 0))
	}
	require.True(t, won.GetGameEndFlag())
	out := p.Output(won, nil)
	assert.Contains(t, out, i18n.T("rollingstone.gameEndYou"))
	assert.NotContains(t, out, i18n.T("rollingstone.promptPlay"), "終局後は促さない")

	// **上限で切った局は「上がった」わけではない。** 投了も手札が残る形。
	stale := newRollingStoneForCui(t)
	stale.GiveUp()
	assert.Contains(t, p.Output(stale, nil), fixedPart("rollingstone.gameEndStalemate"))
}

func TestRollingStoneCuiPresenterShowsErrors(t *testing.T) {
	p := new(RollingStoneCuiPresenter)
	assert.Contains(t, p.Output(newRollingStoneForCui(t), assert.AnError), assert.AnError.Error())
}

func TestRollingStoneCuiPresenterHintOutput(t *testing.T) {
	p := new(RollingStoneCuiPresenter)
	r := newRollingStoneForCui(t)
	r.SetCurrentPlayerIdxForTest(0)

	out := p.HintOutput(r)
	assert.Contains(t, out, fixedPart("rollingstone.hintCard"))
	assert.Contains(t, out, i18n.T("rollingstone.hintReasonLead"))

	// **引き取るしかない場面は札を指さない。**
	r.SetLeadPlayerIdxForTest(1)
	r.GiveHandForTest(0, domain.NewCard(domain.CardDesignHeart, 8, false))
	r.SetCurrentTrickForTest([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 9, false)},
	})
	out = p.HintOutput(r)
	assert.Contains(t, out, fixedPart("rollingstone.hintPickUp"))
	assert.Contains(t, out, i18n.T("rollingstone.hintReasonPickUp"))
	assert.NotContains(t, out, fixedPart("rollingstone.hintCard"))

	// フォローできる場面。
	r.GiveHandForTest(0, domain.NewCard(domain.CardDesignSpade, 8, false))
	assert.Contains(t, p.HintOutput(r), i18n.T("rollingstone.hintReasonFollow"))

	r.GiveUp()
	assert.Contains(t, p.HintOutput(r), i18n.T("rollingstone.hintNone"))
}

func TestRollingStoneCuiPresenterActionLogOutput(t *testing.T) {
	p := new(RollingStoneCuiPresenter)
	r := newRollingStoneForCui(t)
	r.GiveUp()
	assert.NotEmpty(t, p.ActionLogOutput(r))
}
