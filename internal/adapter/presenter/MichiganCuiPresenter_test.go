package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestMichiganCuiPresenter_OutputBetPhase(t *testing.T) {
	g := domain.NewDefaultMichigan()
	p := new(presenter.MichiganCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "ミシガン")
	assert.Contains(t, out, "ラウンド")
	assert.Contains(t, out, "ブードル")
}

func TestMichiganCuiPresenter_OutputBetHint(t *testing.T) {
	p := new(presenter.MichiganCuiPresenter)

	// Held: the human holds the card that claims boodle 0 → it is recommended.
	g := domain.NewDefaultMichigan()
	bc := g.GetBoodle(0).GetCard()
	g.GetPlayer(0).AddCard(domain.NewCard(bc.GetDesign(), bc.GetValue(), false))
	out := p.Output(g, nil)
	assert.Contains(t, out, "推奨")
	assert.Contains(t, out, "boodle0")

	// None: an empty human hand holds no boodle cards → the even-spread tip shows.
	g2 := domain.NewDefaultMichigan()
	g2.GetPlayer(0).ClearHand()
	out2 := p.Output(g2, nil)
	assert.Contains(t, out2, "均等")
}

func TestMichiganCuiPresenter_OutputPlayPhase(t *testing.T) {
	g := domain.NewDefaultMichigan()
	require.NoError(t, g.PlaceHumanBet(michiganWebEvenBet(g.GetBetBudget())))
	p := new(presenter.MichiganCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "デッドハンド")
}

func TestMichiganCuiPresenter_OutputError(t *testing.T) {
	g := domain.NewDefaultMichigan()
	p := new(presenter.MichiganCuiPresenter)
	out := p.Output(g, errors.New("boom"))
	assert.Contains(t, out, "boom")
}

func TestMichiganCuiPresenter_OutputResult(t *testing.T) {
	g := michiganResultGame(false)
	p := new(presenter.MichiganCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "ラウンド")
}

func TestMichiganCuiPresenter_OutputGameEnd(t *testing.T) {
	g := domain.NewDefaultMichigan()
	cfg := g.GetConfig()
	cfg.TargetRounds = 1
	g.SetConfig(cfg)
	g.Reset()
	require.NoError(t, g.PlaceHumanBet(michiganWebEvenBet(g.GetBetBudget())))
	for i := 0; i < 300 && g.GetPhase() == domain.MichiganPhasePlay && g.IsHumanTurn(); i++ {
		pi := g.GetPlayableIndices()
		require.NotEmpty(t, pi)
		require.NoError(t, g.PlayCard(pi[0]))
	}
	require.True(t, g.GetGameEndFlag())
	p := new(presenter.MichiganCuiPresenter)
	assert.NotEmpty(t, p.Output(g, nil))
}

func TestMichiganCuiPresenter_HintOutput(t *testing.T) {
	g := domain.NewDefaultMichigan()
	require.NoError(t, g.PlaceHumanBet(michiganWebEvenBet(g.GetBetBudget())))
	p := new(presenter.MichiganCuiPresenter)
	assert.NotEmpty(t, p.HintOutput(g))
}

func TestMichiganCuiPresenter_HintOutputNone(t *testing.T) {
	g := michiganResultGame(false) // result phase -> GetHint returns nil
	p := new(presenter.MichiganCuiPresenter)
	assert.NotEmpty(t, p.HintOutput(g))
}

func TestMichiganCuiPresenter_ActionLog(t *testing.T) {
	g := michiganResultGame(false)
	p := new(presenter.MichiganCuiPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
