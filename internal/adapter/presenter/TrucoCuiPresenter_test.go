//go:build test

package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestTrucoCuiPresenter_Output_Phases(t *testing.T) {
	p := new(presenter.TrucoCuiPresenter)
	phases := []domain.TrucoPhase{
		domain.TrucoPhasePlay,
		domain.TrucoPhaseRespond,
		domain.TrucoPhaseTrickEnd,
		domain.TrucoPhaseHandEnd,
		domain.TrucoPhaseGameEnd,
	}
	for _, ph := range phases {
		g := domain.NewDefaultTruco()
		g.Reset()
		g.SetPhase(ph)
		if ph == domain.TrucoPhaseRespond {
			g.SetPendingLevel(domain.TrucoLevelTruco)
			g.SetTrucoCallerIdx(1)
			g.SetResponderIdx(0)
		}
		if ph == domain.TrucoPhaseHandEnd {
			g.SetHandWinnerIdx(0)
		}
		out := p.Output(g, nil)
		assert.NotEmpty(t, out, "phase %d output should be non-empty", ph)
	}
}

func TestTrucoCuiPresenter_Output_Error(t *testing.T) {
	p := new(presenter.TrucoCuiPresenter)
	g := domain.NewDefaultTruco()
	g.Reset()
	out := p.Output(g, errors.New("boom"))
	assert.Contains(t, out, "boom")
}

func TestTrucoCuiPresenter_Output_GameEndBanner(t *testing.T) {
	p := new(presenter.TrucoCuiPresenter)
	g := domain.NewDefaultTruco()
	g.Reset()
	g.SetGameEndFlag(true)
	g.SetPhase(domain.TrucoPhaseGameEnd)
	g.SetPlayerMatchPoints(0, 15)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
}

func TestTrucoCuiPresenter_HintOutput(t *testing.T) {
	p := new(presenter.TrucoCuiPresenter)

	t.Run("play hint", func(t *testing.T) {
		g := domain.NewDefaultTruco()
		g.Reset()
		g.SetPhase(domain.TrucoPhasePlay)
		g.SetCurrentPlayerIdx(0)
		g.SetCurrentTrick(nil)
		assert.NotEmpty(t, p.HintOutput(g))
	})

	t.Run("respond hint", func(t *testing.T) {
		g := domain.NewDefaultTruco()
		g.Reset()
		g.SetPhase(domain.TrucoPhaseRespond)
		g.SetResponderIdx(0)
		assert.NotEmpty(t, p.HintOutput(g))
	})

	t.Run("no hint", func(t *testing.T) {
		g := domain.NewDefaultTruco()
		g.Reset()
		g.SetPhase(domain.TrucoPhaseTrickEnd)
		out := p.HintOutput(g)
		assert.True(t, strings.TrimSpace(out) != "")
	})
}

// 応答側もその場で引き上げられるのに、respond の分岐だけ CanDeclareTruco を
// 見ておらず accept/decline の 2 択しか案内していなかった。両方向で見る。
func TestTrucoCuiPresenter_TellsTheResponderTheyCanRaise(t *testing.T) {
	i18n.SetLang("ja")
	p := new(presenter.TrucoCuiPresenter)

	respondAt := func(level int) *domain.Truco {
		g := domain.NewDefaultTruco()
		g.Reset()
		g.SetPhase(domain.TrucoPhaseRespond)
		g.SetTrucoCallerIdx(1)
		g.SetResponderIdx(0)
		g.SetPendingLevel(level)
		return g
	}

	// まだ上がある水準なら引き上げを案内する。
	canRaise := respondAt(domain.TrucoLevelTruco)
	require.True(t, canRaise.CanDeclareTruco(), "fixture must actually allow a raise")
	assert.Contains(t, p.Output(canRaise, nil), i18n.T("truco.promptCanTruco"))

	// 上限に達していれば案内しない。
	maxed := respondAt(domain.TrucoMaxLevel)
	require.False(t, maxed.CanDeclareTruco(), "fixture must actually forbid a raise")
	assert.NotContains(t, p.Output(maxed, nil), i18n.T("truco.promptCanTruco"))
}

func TestTrucoCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.TrucoCuiPresenter)
	g := domain.NewDefaultTruco()
	g.Reset()
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
