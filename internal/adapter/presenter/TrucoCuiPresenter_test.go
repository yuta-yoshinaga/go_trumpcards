//go:build test

package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
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

func TestTrucoCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.TrucoCuiPresenter)
	g := domain.NewDefaultTruco()
	g.Reset()
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
