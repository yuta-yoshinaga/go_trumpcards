//go:build test

package presenter

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newSnapForCui(t *testing.T) *domain.Snap {
	t.Helper()
	g := domain.NewDefaultSnap()
	g.SetClock(func() time.Time { return time.UnixMilli(0) })
	g.SetRand(rand.New(rand.NewSource(1)))
	g.Reset()
	return g
}

func TestSnapCuiPresenterOutput(t *testing.T) {
	p := new(SnapCuiPresenter)
	out := p.Output(newSnapForCui(t), nil)

	assert.Contains(t, out, i18n.T("snap.helpTitle"))
	assert.Contains(t, out, fixedPart("snap.header"))
	// **トリガーが動くことが規則そのもの。** 毎回書く。
	assert.Contains(t, out, i18n.T("snap.rule"))
	assert.Contains(t, out, i18n.T("snap.pileEmpty"))
	assert.Contains(t, out, i18n.T("snap.promptPlay"))
}

// **成立しているかどうかは一目で分かる必要がある。** 反射ゲームなので。
func TestSnapCuiPresenterShoutsWhenSnapIsOn(t *testing.T) {
	p := new(SnapCuiPresenter)
	g := newSnapForCui(t)

	g.SetCenterPileForTest([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)})
	out := p.Output(g, nil)
	assert.Contains(t, out, fixedPart("snap.topLine"))
	assert.NotContains(t, out, i18n.T("snap.availableLine"), "**1 枚では成立しない**")

	g.SetCenterPileForTest([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
	})
	assert.Contains(t, p.Output(g, nil), i18n.T("snap.availableLine"))
}

// **直近に何が起きたかを出す。** 盤面だけでは、誰が取ったのか読めない。
func TestSnapCuiPresenterReportsTheLastEvent(t *testing.T) {
	p := new(SnapCuiPresenter)

	stepped := newSnapForCui(t)
	stepped.StepForTest(0)
	assert.Contains(t, p.Output(stepped, nil), fixedPart("snap.eventStep"))

	wrong := newSnapForCui(t)
	wrong.SetCenterPileForTest([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)})
	require.NoError(t, wrong.PlayerSnap())
	assert.Contains(t, p.Output(wrong, nil), fixedPart("snap.eventSnapWrong"))

	correct := newSnapForCui(t)
	correct.SetCenterPileForTest([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
	})
	require.NoError(t, correct.PlayerSnap())
	assert.Contains(t, p.Output(correct, nil), fixedPart("snap.eventSnapCorrect"))

	eliminated := newSnapForCui(t)
	eliminated.GiveStockForTest(1)
	eliminated.SetCurrentTurnIdxForTest(1)
	eliminated.StepForTest(1)
	assert.Contains(t, p.Output(eliminated, nil), fixedPart("snap.eventEliminated"))
}

func TestSnapCuiPresenterShowsEveryStock(t *testing.T) {
	p := new(SnapCuiPresenter)
	out := p.Output(newSnapForCui(t), nil)
	assert.Contains(t, out, fixedPart("snap.playerLine"))
	assert.Contains(t, out, "26", "2 人なら 26 枚ずつ")
}

func TestSnapCuiPresenterGameEndBanners(t *testing.T) {
	p := new(SnapCuiPresenter)

	lost := newSnapForCui(t)
	lost.GiveUp()
	out := p.Output(lost, nil)
	assert.Contains(t, out, fixedPart("snap.gameEndCpu"))
	assert.NotContains(t, out, i18n.T("snap.promptPlay"), "終局後は促さない")

	won := newSnapForCui(t)
	won.GiveStockForTest(1)
	won.SetCenterPileForTest(nil)
	won.SetCurrentTurnIdxForTest(1)
	won.StepForTest(1)
	require.True(t, won.GetGameEndFlag())
	assert.Contains(t, p.Output(won, nil), i18n.T("snap.gameEndYou"))
}

func TestSnapCuiPresenterShowsErrors(t *testing.T) {
	p := new(SnapCuiPresenter)
	assert.Contains(t, p.Output(newSnapForCui(t), assert.AnError), assert.AnError.Error())
}

func TestSnapCuiPresenterHintOutput(t *testing.T) {
	p := new(SnapCuiPresenter)
	g := newSnapForCui(t)

	g.SetCenterPileForTest([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
	})
	out := p.HintOutput(g)
	assert.Contains(t, out, fixedPart("snap.hintSnap"))
	assert.Contains(t, out, i18n.T("snap.hintReasonDeclare"))

	g.SetCenterPileForTest([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)})
	g.SetCurrentTurnIdxForTest(0)
	out = p.HintOutput(g)
	assert.Contains(t, out, fixedPart("snap.hintWait"))
	assert.Contains(t, out, i18n.T("snap.hintReasonStep"))

	g.SetCurrentTurnIdxForTest(1)
	assert.Contains(t, p.HintOutput(g), i18n.T("snap.hintReasonWait"))

	g.GiveUp()
	assert.Contains(t, p.HintOutput(g), i18n.T("snap.hintNone"))
}

func TestSnapCuiPresenterActionLogOutput(t *testing.T) {
	p := new(SnapCuiPresenter)
	g := newSnapForCui(t)
	g.GiveUp()
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
