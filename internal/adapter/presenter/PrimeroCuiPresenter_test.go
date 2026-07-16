package presenter_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestPrimeroCuiPresenter_OutputBettingPhase(t *testing.T) {
	g := domain.NewDefaultPrimero()
	p := new(presenter.PrimeroCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	// Title and round line should be present.
	assert.Contains(t, out, "プリメロ")
	assert.Contains(t, out, "ラウンド")
}

func TestPrimeroCuiPresenter_OutputBettingShowsCallNeed(t *testing.T) {
	g := domain.NewDefaultPrimero()
	require.Equal(t, domain.PrimeroPhaseBetting, g.GetPhase())
	p := new(presenter.PrimeroCuiPresenter)
	out := p.Output(g, nil)

	actor := g.GetPlayer(g.GetCurrentPlayerIdx())
	need := g.GetCurrentBet() - actor.GetRoundBet()
	raiseTo := g.GetCurrentBet() + g.GetAnte()
	assert.Contains(t, out, "必要 "+strconv.Itoa(need))
	assert.Contains(t, out, strconv.Itoa(raiseTo))
}

func TestPrimeroCuiPresenter_OutputError(t *testing.T) {
	g := domain.NewDefaultPrimero()
	p := new(presenter.PrimeroCuiPresenter)
	out := p.Output(g, errors.New("boom"))
	assert.Contains(t, out, "boom")
}

func TestPrimeroCuiPresenter_OutputResult(t *testing.T) {
	g := primeroResultGame(false) // human loses to a CPU fluxus (shared helper)
	p := new(presenter.PrimeroCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "ラウンド")
}

func TestPrimeroCuiPresenter_OutputGameEnd(t *testing.T) {
	g := domain.NewDefaultPrimero()
	cfg := g.GetConfig()
	cfg.TargetRounds = 1
	g.SetConfig(cfg)
	g.Reset()
	for i := 0; i < 100 && g.GetPhase() == domain.PrimeroPhaseBetting && g.IsHumanTurn(); i++ {
		require.NoError(t, g.PlayerCall())
	}
	require.True(t, g.GetGameEndFlag())
	p := new(presenter.PrimeroCuiPresenter)
	assert.NotEmpty(t, p.Output(g, nil))
}

func TestPrimeroCuiPresenter_HintOutput(t *testing.T) {
	g := domain.NewDefaultPrimero()
	p := new(presenter.PrimeroCuiPresenter)
	assert.NotEmpty(t, p.HintOutput(g))
}

func TestPrimeroCuiPresenter_HintOutputNone(t *testing.T) {
	g := primeroResultGame(false) // result phase → GetHint returns nil
	p := new(presenter.PrimeroCuiPresenter)
	assert.NotEmpty(t, p.HintOutput(g))
}

func TestPrimeroCuiPresenter_ActionLog(t *testing.T) {
	g := primeroResultGame(false)
	p := new(presenter.PrimeroCuiPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
