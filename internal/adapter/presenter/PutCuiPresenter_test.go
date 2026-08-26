//go:build test && (!js || !wasm || extra4)

package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestPutCuiPresenter_Output_Phases(t *testing.T) {
	p := new(presenter.PutCuiPresenter)
	phases := []domain.PutPhase{
		domain.PutPhasePlay,
		domain.PutPhaseRespond,
		domain.PutPhaseTrickEnd,
		domain.PutPhaseHandEnd,
		domain.PutPhaseGameEnd,
	}
	for _, ph := range phases {
		g := domain.NewDefaultPut()
		g.Reset()
		g.SetPhase(ph)
		if ph == domain.PutPhaseRespond {
			g.SetPendingLevel(domain.PutLevelPut)
			g.SetPutCallerIdx(1)
			g.SetResponderIdx(0)
		}
		if ph == domain.PutPhaseHandEnd {
			g.SetHandWinnerIdx(0)
		}
		out := p.Output(g, nil)
		assert.NotEmpty(t, out, "phase %d output should be non-empty", ph)
	}
}

func TestPutCuiPresenter_Output_Error(t *testing.T) {
	p := new(presenter.PutCuiPresenter)
	g := domain.NewDefaultPut()
	g.Reset()
	out := p.Output(g, errors.New("boom"))
	assert.Contains(t, out, "boom")
}

func TestPutCuiPresenter_Output_GameEndBanner(t *testing.T) {
	p := new(presenter.PutCuiPresenter)
	g := domain.NewDefaultPut()
	g.Reset()
	g.SetGameEndFlag(true)
	g.SetPhase(domain.PutPhaseGameEnd)
	g.SetPlayerMatchPoints(0, 15)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
}

func TestPutCuiPresenter_HintOutput(t *testing.T) {
	p := new(presenter.PutCuiPresenter)

	t.Run("play hint", func(t *testing.T) {
		g := domain.NewDefaultPut()
		g.Reset()
		g.SetPhase(domain.PutPhasePlay)
		g.SetCurrentPlayerIdx(0)
		g.SetCurrentTrick(nil)
		assert.NotEmpty(t, p.HintOutput(g))
	})

	t.Run("respond hint", func(t *testing.T) {
		g := domain.NewDefaultPut()
		g.Reset()
		g.SetPhase(domain.PutPhaseRespond)
		g.SetResponderIdx(0)
		assert.NotEmpty(t, p.HintOutput(g))
	})

	t.Run("no hint", func(t *testing.T) {
		g := domain.NewDefaultPut()
		g.Reset()
		g.SetPhase(domain.PutPhaseTrickEnd)
		out := p.HintOutput(g)
		assert.True(t, strings.TrimSpace(out) != "")
	})
}

func TestPutCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.PutCuiPresenter)
	g := domain.NewDefaultPut()
	g.Reset()
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
