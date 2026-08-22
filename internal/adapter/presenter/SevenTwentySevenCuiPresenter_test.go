//go:build test

package presenter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func s27NoColor(t *testing.T) {
	t.Helper()
	orig := color.NoColor()
	color.SetNoColor(true)
	t.Cleanup(func() { color.SetNoColor(orig) })
}

func s27Card(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

// **2 つの目標と自分の両側の得点を毎回出す。** 7 と 27 のどちらに寄せるかが
// このゲームそのもので、書いていなければ何を選んでいるのか読めない。
func TestSevenTwentySevenCuiPresenter_Output_AlwaysShowsBothTargetsAndScores(t *testing.T) {
	s27NoColor(t)
	p := new(presenter.SevenTwentySevenCuiPresenter)
	g := domain.NewDefaultSevenTwentySeven()
	// 4 + K = 4.5 点。どちらの側も生きている。
	g.SetHandForTest(0, []*domain.Card{s27Card(domain.CardDesignSpade, 4), s27Card(domain.CardDesignHeart, 13)})

	out := p.Output(g, nil)
	assert.Contains(t, out, i18n.T("seventwentyseven.targetsNote"))
	assert.Contains(t, out, i18n.Tf("seventwentyseven.yourScore", "score", "4.5 / 4.5"))
	assert.NotContains(t, out, "seventwentyseven.targetsNote", "キーがそのまま出ている")
}

// **超過した側は「-」。** 数字を出すと、超えているのに勝負が残っているように読める。
func TestSevenTwentySevenCuiPresenter_Output_MarksABustedSide(t *testing.T) {
	s27NoColor(t)
	p := new(presenter.SevenTwentySevenCuiPresenter)
	g := domain.NewDefaultSevenTwentySeven()
	// 10 + 9 = 19 点。7 側は失格、27 側は生存。
	g.SetHandForTest(0, []*domain.Card{s27Card(domain.CardDesignSpade, 10), s27Card(domain.CardDesignHeart, 9)})

	out := p.Output(g, nil)
	assert.Contains(t, out, i18n.Tf("seventwentyseven.yourScore", "score", "- / 19"))
}

// **両側の勝者を名指しする。** どちらを取ったのかが分からないと、
// なぜ半分なのかが読めない。
func TestSevenTwentySevenCuiPresenter_Output_NamesBothSideWinners(t *testing.T) {
	s27NoColor(t)
	p := new(presenter.SevenTwentySevenCuiPresenter)
	g := domain.NewDefaultSevenTwentySeven()
	g.SetHandForTest(0, []*domain.Card{s27Card(domain.CardDesignSpade, 6)}) // 6 → 7 側
	g.SetHandForTest(1, []*domain.Card{s27Card(domain.CardDesignSpade, 10), s27Card(domain.CardDesignHeart, 9),
		s27Card(domain.CardDesignClover, 8)}) // 27 ちょうど
	g.SetHandForTest(2, []*domain.Card{s27Card(domain.CardDesignSpade, 10), s27Card(domain.CardDesignHeart, 10),
		s27Card(domain.CardDesignClover, 10)}) // 両方超過
	g.SetHandForTest(3, []*domain.Card{s27Card(domain.CardDesignDiamond, 10), s27Card(domain.CardDesignSpade, 10),
		s27Card(domain.CardDesignHeart, 10)})
	g.SetPotForTest(100)
	g.StandEveryoneForTest()
	g.SettleForTest()

	out := p.Output(g, nil)
	assert.Contains(t, out, i18n.Tf("seventwentyseven.result.lowWinner", "name", "あなた", "score", "6"))
	assert.Contains(t, out, i18n.Tf("seventwentyseven.result.highWinner", "name", "CPU 1", "score", "27"))
	assert.NotContains(t, out, i18n.T("seventwentyseven.result.lowEmpty"))
}

// 総取りは専用の文言で出す。半分ずつの表記だと誤解する。
func TestSevenTwentySevenCuiPresenter_Output_AnnouncesAScoop(t *testing.T) {
	s27NoColor(t)
	p := new(presenter.SevenTwentySevenCuiPresenter)
	g := domain.NewDefaultSevenTwentySeven()
	// A A 5 = 7 にも 27 にもできる。
	g.SetHandForTest(0, []*domain.Card{s27Card(domain.CardDesignSpade, 1), s27Card(domain.CardDesignHeart, 1),
		s27Card(domain.CardDesignClover, 5)})
	for i := 1; i < g.GetPlayerCnt(); i++ {
		g.SetHandForTest(i, []*domain.Card{s27Card(domain.CardDesignSpade, 10), s27Card(domain.CardDesignHeart, 8)})
	}
	g.SetPotForTest(100)
	g.StandEveryoneForTest()
	g.SettleForTest()

	out := p.Output(g, nil)
	assert.Contains(t, out, i18n.Tf("seventwentyseven.result.scoop", "name", "あなた"))
	assert.NotContains(t, out, i18n.Tf("seventwentyseven.result.lowWinner", "name", "あなた", "score", "7"))
}

// 全員が両側とも超えたら持ち越しと言う。
func TestSevenTwentySevenCuiPresenter_Output_SaysThePotCarriesOver(t *testing.T) {
	s27NoColor(t)
	p := new(presenter.SevenTwentySevenCuiPresenter)
	g := domain.NewDefaultSevenTwentySeven()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		g.SetHandForTest(i, []*domain.Card{s27Card(domain.CardDesignSpade, 10), s27Card(domain.CardDesignHeart, 10),
			s27Card(domain.CardDesignClover, 10)})
	}
	g.SetPotForTest(100)
	g.StandEveryoneForTest()
	g.SettleForTest()

	out := p.Output(g, nil)
	assert.Contains(t, out, i18n.Tf("seventwentyseven.result.carry", "pot", "100", "count", "1"))
}

// ヒントは狙っている側を言う。「引け」だけでは助言にならない。
func TestSevenTwentySevenCuiPresenter_HintOutput_NamesTheSide(t *testing.T) {
	s27NoColor(t)
	p := new(presenter.SevenTwentySevenCuiPresenter)
	g := domain.NewDefaultSevenTwentySeven()

	g.SetHandForTest(0, []*domain.Card{s27Card(domain.CardDesignSpade, 10), s27Card(domain.CardDesignHeart, 10)})
	out := p.HintOutput(g)
	assert.Contains(t, out, i18n.T("seventwentyseven.takeCard"))
	assert.Contains(t, out, i18n.T("seventwentyseven.hintReasonChaseTwentySeven"))

	g.SetHandForTest(0, []*domain.Card{s27Card(domain.CardDesignSpade, 7)})
	out = p.HintOutput(g)
	assert.Contains(t, out, i18n.T("seventwentyseven.stand"))
	assert.Contains(t, out, i18n.T("seventwentyseven.hintReasonExactlySeven"))

	// 止まっていれば助言しない。
	g.SetStandingForTest(0, true)
	assert.Equal(t, i18n.T("seventwentyseven.hintNone")+"\n", p.HintOutput(g))
}

func TestSevenTwentySevenCuiPresenter_Output_ShowsTheError(t *testing.T) {
	s27NoColor(t)
	p := new(presenter.SevenTwentySevenCuiPresenter)
	g := domain.NewDefaultSevenTwentySeven()
	out := p.Output(g, domain.ErrWrongPhase)
	assert.Contains(t, out, domain.ErrWrongPhase.Error())
}

func TestSevenTwentySevenCuiPresenter_ActionLogOutput(t *testing.T) {
	s27NoColor(t)
	p := new(presenter.SevenTwentySevenCuiPresenter)
	g := domain.NewDefaultSevenTwentySeven()
	require.NotEmpty(t, p.ActionLogOutput(g))
}
