//go:build test

package presenter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newGermanWhistForCui(t *testing.T) *domain.GermanWhist {
	t.Helper()
	g := domain.NewDefaultGermanWhist()
	g.Reset()
	return g
}

func TestGermanWhistCuiPresenterOutput(t *testing.T) {
	p := new(GermanWhistCuiPresenter)
	g := newGermanWhistForCui(t)

	out := p.Output(g, nil)

	assert.Contains(t, out, i18n.T("germanwhist.helpTitle"))
	assert.Contains(t, out, i18n.T("germanwhist.phaseFirst"))
	assert.Contains(t, out, i18n.T("germanwhist.promptPlay"))
	// 前半は表向きの札が出ている。
	assert.Contains(t, out, strings.SplitN(i18n.T("germanwhist.upCardLine"), "{{", 2)[0])
	assert.NotContains(t, out, i18n.T("germanwhist.upCardNone"))
}

// 山札が尽きたら「表向きの札なし」に切り替わる。両方向を踏む。
func TestGermanWhistCuiPresenterUpCardNone(t *testing.T) {
	p := new(GermanWhistCuiPresenter)
	g := newGermanWhistForCui(t)
	g.SetUpCard(nil)

	out := p.Output(g, nil)
	assert.Contains(t, out, i18n.T("germanwhist.upCardNone"))
}

func TestGermanWhistCuiPresenterSecondHalf(t *testing.T) {
	p := new(GermanWhistCuiPresenter)
	g := newGermanWhistForCui(t)
	g.SetPhase(domain.GermanWhistPhaseScoring)

	out := p.Output(g, nil)
	assert.Contains(t, out, i18n.T("germanwhist.phaseSecond"))
	assert.NotContains(t, out, i18n.T("germanwhist.phaseFirst"))
}

func TestGermanWhistCuiPresenterError(t *testing.T) {
	p := new(GermanWhistCuiPresenter)
	g := newGermanWhistForCui(t)

	assert.Contains(t, p.Output(g, assert.AnError), assert.AnError.Error())
}

func TestGermanWhistCuiPresenterGameEnd(t *testing.T) {
	for _, tc := range []struct {
		name    string
		s0, s1  int
		wantKey string
	}{
		{"human wins", 7, 6, "germanwhist.gameEndP0"},
		{"cpu wins", 6, 7, "germanwhist.gameEndP1"},
		{"tie", 6, 6, "germanwhist.gameEndTie"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := new(GermanWhistCuiPresenter)
			g := newGermanWhistForCui(t)
			g.GetPlayer(0).SetScoringTricks(tc.s0)
			g.GetPlayer(1).SetScoringTricks(tc.s1)
			g.Finish()

			out := p.Output(g, nil)
			// 色コードが混ざるので、置換前の固定部分で照合する。
			assert.Contains(t, out, strings.SplitN(i18n.T(tc.wantKey), "{{", 2)[0])
			// 終局したら手番の案内は出さない。
			assert.NotContains(t, out, i18n.T("germanwhist.promptPlay"))
		})
	}
}

func TestGermanWhistCuiPresenterHintOutput(t *testing.T) {
	p := new(GermanWhistCuiPresenter)
	g := newGermanWhistForCui(t)

	out := p.HintOutput(g)
	assert.Contains(t, out, "HINT")
	// 理由キーは i18n に解決されている。生のキーが出ていたら未登録。
	assert.NotContains(t, out, "germanWhistTakeUpCard")
	assert.NotContains(t, out, "germanWhistDuck")
}

// 前半の 2 つの狙いがそれぞれ日本語文言になる。
func TestGermanWhistCuiPresenterHintReasons(t *testing.T) {
	p := new(GermanWhistCuiPresenter)

	g := newGermanWhistForCui(t)
	g.SetUpCard(domain.NewCard(g.GetTrumpSuit(), 1, false)) // 切り札のA
	assert.Contains(t, p.HintOutput(g), i18n.T("germanwhist.hintReasonTakeUpCard"))

	g2 := newGermanWhistForCui(t)
	// 切り札ではない最弱の札。
	weak := domain.CardDesignHeart
	if g2.GetTrumpSuit() == domain.CardDesignHeart {
		weak = domain.CardDesignSpade
	}
	g2.SetUpCard(domain.NewCard(weak, 2, false))
	assert.Contains(t, p.HintOutput(g2), i18n.T("germanwhist.hintReasonDuck"))
}

func TestGermanWhistCuiPresenterHintNoneAfterGameEnd(t *testing.T) {
	p := new(GermanWhistCuiPresenter)
	g := newGermanWhistForCui(t)
	g.GiveUp()

	assert.Contains(t, p.HintOutput(g), i18n.T("germanwhist.hintNone"))
}

func TestGermanWhistCuiPresenterActionLogOutput(t *testing.T) {
	p := new(GermanWhistCuiPresenter)
	g := newGermanWhistForCui(t)
	g.GiveUp()

	out := p.ActionLogOutput(g)
	require.NotEmpty(t, out)
}

// スート名は 4 つとも解決される。"?" が出たら未登録。
func TestGermanWhistSuitName(t *testing.T) {
	for _, suit := range []int{
		domain.CardDesignSpade, domain.CardDesignClover,
		domain.CardDesignHeart, domain.CardDesignDiamond,
	} {
		assert.NotEqual(t, "?", germanWhistSuitName(suit))
	}
	assert.Equal(t, "?", germanWhistSuitName(domain.CardDesignJoker))
}
