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

func TestGutsCuiPresenter_OutputDeclarePhase(t *testing.T) {
	g := domain.NewDefaultGuts()
	p := new(presenter.GutsCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	// Round line + declare prompt should be present.
	assert.Contains(t, out, "ラウンド")
	assert.Contains(t, out, "宣言")
}

func TestGutsCuiPresenter_OutputError(t *testing.T) {
	g := domain.NewDefaultGuts()
	p := new(presenter.GutsCuiPresenter)
	out := p.Output(g, errors.New("boom"))
	assert.Contains(t, out, "boom")
}

func TestGutsCuiPresenter_OutputResultLose(t *testing.T) {
	g := gutsWebResultGame() // human loses to a CPU pair (shared helper)
	p := new(presenter.GutsCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
}

func TestGutsCuiPresenter_OutputResultCarry(t *testing.T) {
	g := domain.NewDefaultGuts()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		g.GetPlayer(i).SetIn(false)
	}
	g.SettleForTest()
	p := new(presenter.GutsCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	// The carry line reports the carried-over pot and the consecutive-carry count.
	assert.Equal(t, 1, g.GetCarryCount())
	assert.Contains(t, out, strconv.Itoa(g.GetCarryPot()))
	assert.Contains(t, out, "持ち越し")
}

func TestGutsCuiPresenter_OutputGameEnd(t *testing.T) {
	g := domain.NewDefaultGuts()
	cfg := g.GetConfig()
	cfg.TargetRounds = 1
	g.SetConfig(cfg)
	g.Reset()
	require.NoError(t, g.Declare(true))
	require.True(t, g.GetGameEndFlag())
	p := new(presenter.GutsCuiPresenter)
	assert.NotEmpty(t, p.Output(g, nil))
}

func TestGutsCuiPresenter_HintOutput(t *testing.T) {
	g := domain.NewDefaultGuts()
	p := new(presenter.GutsCuiPresenter)
	assert.NotEmpty(t, p.HintOutput(g))
}

func TestGutsCuiPresenter_HintOutputNone(t *testing.T) {
	g := domain.NewDefaultGuts()
	require.NoError(t, g.Declare(true)) // result phase → GetHint returns nil
	p := new(presenter.GutsCuiPresenter)
	assert.NotEmpty(t, p.HintOutput(g))
}

func TestGutsCuiPresenter_ActionLog(t *testing.T) {
	g := domain.NewDefaultGuts()
	require.NoError(t, g.Declare(true))
	p := new(presenter.GutsCuiPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
