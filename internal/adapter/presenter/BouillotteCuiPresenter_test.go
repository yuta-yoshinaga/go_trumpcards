package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestBouillotteCuiPresenter_OutputBettingPhase(t *testing.T) {
	g := domain.NewDefaultBouillotte()
	p := new(presenter.BouillotteCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	// Title, round line and the shared retourne card should be present.
	assert.Contains(t, out, "ブイヨット")
	assert.Contains(t, out, "ラウンド")
	assert.Contains(t, out, "ルトゥルヌ")
}

func TestBouillotteCuiPresenter_OutputError(t *testing.T) {
	g := domain.NewDefaultBouillotte()
	p := new(presenter.BouillotteCuiPresenter)
	out := p.Output(g, errors.New("boom"))
	assert.Contains(t, out, "boom")
}

func TestBouillotteCuiPresenter_OutputResult(t *testing.T) {
	g := bouillotteResultGame(false) // human loses to a CPU brelan (shared helper)
	p := new(presenter.BouillotteCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "ラウンド")
}

func TestBouillotteCuiPresenter_OutputGameEnd(t *testing.T) {
	g := domain.NewDefaultBouillotte()
	cfg := g.GetConfig()
	cfg.TargetRounds = 1
	g.SetConfig(cfg)
	g.Reset()
	for i := 0; i < 100 && g.GetPhase() == domain.BouillottePhaseBetting && g.IsHumanTurn(); i++ {
		require.NoError(t, g.PlayerCall())
	}
	require.True(t, g.GetGameEndFlag())
	p := new(presenter.BouillotteCuiPresenter)
	assert.NotEmpty(t, p.Output(g, nil))
}

func TestBouillotteCuiPresenter_HintOutput(t *testing.T) {
	g := domain.NewDefaultBouillotte()
	p := new(presenter.BouillotteCuiPresenter)
	assert.NotEmpty(t, p.HintOutput(g))
}

func TestBouillotteCuiPresenter_HintOutputNone(t *testing.T) {
	g := bouillotteResultGame(false) // result phase → GetHint returns nil
	p := new(presenter.BouillotteCuiPresenter)
	assert.NotEmpty(t, p.HintOutput(g))
}

func TestBouillotteCuiPresenter_ActionLog(t *testing.T) {
	g := bouillotteResultGame(false)
	p := new(presenter.BouillotteCuiPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
