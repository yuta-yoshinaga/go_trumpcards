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

func TestGoStopCuiPresenter_Output_PlayPhase(t *testing.T) {
	g := domain.NewDefaultGoStop()
	g.Reset()
	g.SetCurrentTurn(0)
	p := new(presenter.GoStopCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "[0]")
}

func TestGoStopCuiPresenter_Output_AllPhases(t *testing.T) {
	p := new(presenter.GoStopCuiPresenter)

	g := domain.NewDefaultGoStop()
	g.Reset()
	g.SetCurrentTurn(0)
	assert.NotEmpty(t, p.Output(g, nil))

	g.SetPhase(domain.GoStopPhaseGoDecision)
	assert.NotEmpty(t, p.Output(g, nil))

	g.SetPhase(domain.GoStopPhaseRoundEnd)
	assert.NotEmpty(t, p.Output(g, nil))

	g.SetPhase(domain.GoStopPhaseGameEnd)
	assert.NotEmpty(t, p.Output(g, nil))

	assert.NotEmpty(t, p.Output(g, errors.New("boom")))

	g.SetPhase(domain.GoStopPhasePlay)
	g.SetFieldCards(nil)
	assert.NotEmpty(t, p.Output(g, nil))
}

func TestGoStopCuiPresenter_CategoryCounts(t *testing.T) {
	g := domain.NewDefaultGoStop()
	g.Reset()
	g.SetCurrentTurn(0)
	g.GetPlayer(0).AddCaptured(gostopGwang3Cards()) // 3 光 (Gwang)
	p := new(presenter.GoStopCuiPresenter)
	out := p.Output(g, nil)
	assert.Contains(t, out, strings.Split(i18n.T("gostop.categoryCounts"), "{{")[0])
	// The three captured Gwang are counted in the category line.
	assert.Contains(t, out, "光3")
}

func TestGoStopCuiPresenter_HintOutput(t *testing.T) {
	p := new(presenter.GoStopCuiPresenter)

	g := domain.NewDefaultGoStop()
	g.Reset()
	g.SetCurrentTurn(0)
	assert.NotEmpty(t, p.HintOutput(g))

	g.SetPhase(domain.GoStopPhaseGoDecision)
	assert.NotEmpty(t, p.HintOutput(g))
}

func TestGoStopCuiPresenter_ActionLog(t *testing.T) {
	g := domain.NewDefaultGoStop()
	g.Reset()
	p := new(presenter.GoStopCuiPresenter)
	assert.NotNil(t, p.ActionLogOutput(g))
}

func gostopGwang3Cards() []*domain.Card {
	return []*domain.Card{domain.NewCard(1, 1, false), domain.NewCard(3, 1, false), domain.NewCard(12, 1, false)}
}

func TestGoStopCuiPresenter_RoundResult_HumanWinner(t *testing.T) {
	g := domain.NewDefaultGoStop()
	g.Reset()
	g.SetCurrentTurn(0)
	g.SetPhase(domain.GoStopPhaseGoDecision)
	g.GetPlayer(0).AddCaptured(gostopGwang3Cards())
	require.NoError(t, g.PlayerDecide(false))
	require.Equal(t, domain.GoStopPhaseRoundEnd, g.GetPhase())

	p := new(presenter.GoStopCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
}

func TestGoStopCuiPresenter_RoundResult_CpuWinner(t *testing.T) {
	g := domain.NewDefaultGoStop()
	g.Reset()
	cfg := domain.DefaultGoStopConfig()
	cfg.CpuDifficulty = domain.GoStopCpuDifficultyEasy // Easy always stops
	g.SetConfig(cfg)
	g.SetCurrentTurn(1)
	g.SetPhase(domain.GoStopPhaseGoDecision)
	g.GetPlayer(1).AddCaptured(gostopGwang3Cards())
	g.CpuDecide()
	require.Equal(t, domain.GoStopPhaseRoundEnd, g.GetPhase())

	p := new(presenter.GoStopCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
}
