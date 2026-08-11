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

func newBhabhiForCui(t *testing.T) *domain.Bhabhi {
	t.Helper()
	b := domain.NewDefaultBhabhi()
	b.Reset()
	return b
}

func TestBhabhiCuiPresenterOutput(t *testing.T) {
	p := new(BhabhiCuiPresenter)
	out := p.Output(newBhabhiForCui(t), nil)

	assert.Contains(t, out, i18n.T("bhabhi.helpTitle"))
	assert.Contains(t, out, fixedPart("bhabhi.header"))
	// **勝者ではなく敗者を決めるゲーム。** 目的を毎回書く。
	assert.Contains(t, out, i18n.T("bhabhi.rule"))
	assert.Contains(t, out, "[0]", "人間の手札は番号付き")
}

// リードは未開始と確定の両側を踏む。
func TestBhabhiCuiPresenterLeadLine(t *testing.T) {
	p := new(BhabhiCuiPresenter)

	fresh := newBhabhiForCui(t)
	assert.Contains(t, p.Output(fresh, nil), i18n.T("bhabhi.leadNone"))

	led := newBhabhiForCui(t)
	led.SetLeadSuitForTest(domain.CardDesignHeart)
	led.SetPileForTest([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
	})
	out := p.Output(led, nil)
	assert.Contains(t, out, fixedPart("bhabhi.leadLine"))
	assert.Contains(t, out, i18n.T("bhabhi.suitHeart"))
	assert.NotContains(t, out, i18n.T("bhabhi.leadNone"))
}

// **4 スートすべてに名前がある。** 既定の "?" に落ちない。
func TestBhabhiCuiPresenterNamesEverySuit(t *testing.T) {
	p := new(BhabhiCuiPresenter)
	for suit, key := range map[int]string{
		domain.CardDesignSpade:   "bhabhi.suitSpade",
		domain.CardDesignClover:  "bhabhi.suitClover",
		domain.CardDesignHeart:   "bhabhi.suitHeart",
		domain.CardDesignDiamond: "bhabhi.suitDiamond",
	} {
		b := newBhabhiForCui(t)
		b.SetLeadSuitForTest(suit)
		b.SetPileForTest([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(suit, 5, false)},
		})
		assert.Contains(t, p.Output(b, nil), i18n.T(key))
	}
	assert.Equal(t, "?", bhabhiSuitName(0))
}

// **上がった席と残っている席は見た目で区別が付く。**
func TestBhabhiCuiPresenterMarksFinishedSeats(t *testing.T) {
	p := new(BhabhiCuiPresenter)
	b := newBhabhiForCui(t)
	b.GetPlayer(1).Reset()
	b.GetPlayer(1).SetRank(1)

	out := p.Output(b, nil)
	assert.Contains(t, out, fixedPart("bhabhi.stateOut"))
	assert.Contains(t, out, fixedPart("bhabhi.stateCards"))
	// **ルール行にも「引き取り」が出る。** 席行だけを数えるため回数の形で拾う。
	assert.Len(t, regexp.MustCompile(`引き取り\d+回`).FindAllString(out, -1),
		domain.BhabhiDefaultPlayers, "全員の席行が出る")
}

// **直前の引き取りは盤面に痕跡が残らない。** 何枚どこへ行ったか言う。
func TestBhabhiCuiPresenterReportsTheLastPickup(t *testing.T) {
	p := new(BhabhiCuiPresenter)
	b := newBhabhiForCui(t)
	assert.NotContains(t, p.Output(b, nil), fixedPart("bhabhi.lastPickup"), "まだ引き取りは起きていない")

	b.SetLeadIdxForTest(0)
	b.SetCurrentIdxForTest(0)
	bhabhiCuiHand(b, 0, domain.NewCard(domain.CardDesignSpade, 5, false), domain.NewCard(domain.CardDesignClover, 4, false))
	bhabhiCuiHand(b, 1, domain.NewCard(domain.CardDesignHeart, 3, false), domain.NewCard(domain.CardDesignHeart, 8, false))
	bhabhiCuiHand(b, 2, domain.NewCard(domain.CardDesignSpade, 9, false), domain.NewCard(domain.CardDesignClover, 2, false))
	bhabhiCuiHand(b, 3, domain.NewCard(domain.CardDesignSpade, 2, false), domain.NewCard(domain.CardDesignClover, 6, false))
	require.NoError(t, b.PlayForTest(0, 0))
	require.NoError(t, b.PlayForTest(1, 0))

	assert.Contains(t, p.Output(b, nil), fixedPart("bhabhi.lastPickup"))
}

// bhabhiCuiHand は playerIdx の手札を cards ちょうどに置き換える。
func bhabhiCuiHand(b *domain.Bhabhi, playerIdx int, cards ...*domain.Card) {
	p := b.GetPlayer(playerIdx)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// **終わり方は 3 通りあり、それぞれ別の文言になる。**
func TestBhabhiCuiPresenterGameEndBanners(t *testing.T) {
	p := new(BhabhiCuiPresenter)

	you := newBhabhiForCui(t)
	you.GiveUp()
	assert.Contains(t, p.Output(you, nil), fixedPart("bhabhi.gameEndYou"))

	cpu := newBhabhiForCui(t)
	cpu.GetPlayer(0).Reset()
	cpu.GetPlayer(0).SetRank(1)
	cpu.GetPlayer(2).Reset()
	cpu.GetPlayer(2).SetRank(2)
	cpu.GetPlayer(3).Reset()
	cpu.GetPlayer(3).SetRank(3)
	cpu.FinishGameForTest()
	out := p.Output(cpu, nil)
	assert.Contains(t, out, fixedPart("bhabhi.gameEndCpu"))
	assert.NotContains(t, out, fixedPart("bhabhi.gameEndYou"))

	stale := newBhabhiForCui(t)
	stale.SetTrickNumberForTest(domain.BhabhiStalemateTricks)
	stale.FinishStalemateForTest()
	assert.Contains(t, p.Output(stale, nil), fixedPart("bhabhi.gameEndStalemate"))
}

func TestBhabhiCuiPresenterShowsErrors(t *testing.T) {
	p := new(BhabhiCuiPresenter)
	b := newBhabhiForCui(t)
	err := b.PlayerPlay(999)
	require.Error(t, err)
	assert.Contains(t, p.Output(b, err), err.Error())
}

func TestBhabhiCuiPresenterHint(t *testing.T) {
	p := new(BhabhiCuiPresenter)
	b := newBhabhiForCui(t)
	b.SetCurrentIdxForTest(0)

	out := p.HintOutput(b)
	assert.Contains(t, out, "HINT")
	for id := range bhabhiHintReasonKeys {
		assert.NotContains(t, out, id, "識別子がそのまま漏れていない")
	}

	b.FinishGameForTest()
	assert.Contains(t, p.HintOutput(b), i18n.T("bhabhi.hintNone"))
}

// **理由の識別子はすべて訳文を持つ。**
func TestBhabhiCuiPresenterHintReasonsAllTranslate(t *testing.T) {
	assert.NotEmpty(t, bhabhiHintReasonKeys)
	for id, key := range bhabhiHintReasonKeys {
		assert.NotEqual(t, key, i18n.T(key), "訳が無い: "+id)
	}
}

func TestBhabhiCuiPresenterActionLog(t *testing.T) {
	p := new(BhabhiCuiPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(newBhabhiForCui(t)))
}
