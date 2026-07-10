package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestHachiHachiCuiPresenter_Output_PlayPhase(t *testing.T) {
	g := domain.NewDefaultHachiHachi()
	g.Reset()
	g.SetCurrentTurn(0)
	p := new(presenter.HachiHachiCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "[0]") // human indexed hand + indexed field
}

func TestHachiHachiCuiPresenter_Output_AllPhases(t *testing.T) {
	p := new(presenter.HachiHachiCuiPresenter)

	g := domain.NewDefaultHachiHachi()
	g.Reset()
	g.SetCurrentTurn(0)
	assert.NotEmpty(t, p.Output(g, nil))

	g.SetPhase(domain.HachiHachiPhaseRoundEnd)
	assert.NotEmpty(t, p.Output(g, nil))

	g.SetPhase(domain.HachiHachiPhaseGameEnd)
	assert.NotEmpty(t, p.Output(g, nil))

	// エラー出力。
	assert.NotEmpty(t, p.Output(g, errors.New("boom")))

	// 空フィールド表示。
	g.SetPhase(domain.HachiHachiPhasePlay)
	g.SetFieldCards(nil)
	assert.NotEmpty(t, p.Output(g, nil))
}

func TestHachiHachiCuiPresenter_HintOutput(t *testing.T) {
	p := new(presenter.HachiHachiCuiPresenter)

	g := domain.NewDefaultHachiHachi()
	g.Reset()
	g.SetCurrentTurn(0)
	assert.NotEmpty(t, p.HintOutput(g))

	// ヒント無し (CPU 手番)。
	g.SetCurrentTurn(1)
	assert.NotEmpty(t, p.HintOutput(g))
}

func TestHachiHachiCuiPresenter_ActionLog(t *testing.T) {
	g := domain.NewDefaultHachiHachi()
	g.Reset()
	p := new(presenter.HachiHachiCuiPresenter)
	assert.NotNil(t, p.ActionLogOutput(g))
}

func TestHachiHachiCuiPresenter_RoundResult(t *testing.T) {
	g := domain.NewDefaultHachiHachi()
	cfg := domain.DefaultHachiHachiConfig()
	cfg.TargetRounds = 3
	g.SetConfig(cfg)
	g.Reset()
	for step := 0; step < 20000 && g.GetPhase() == domain.HachiHachiPhasePlay; step++ {
		if g.IsHumanTurn() {
			require.NoError(t, g.PlayerPlay(0, -1))
		} else {
			g.CpuPlay()
		}
	}
	require.Equal(t, domain.HachiHachiPhaseRoundEnd, g.GetPhase())

	p := new(presenter.HachiHachiCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
}
