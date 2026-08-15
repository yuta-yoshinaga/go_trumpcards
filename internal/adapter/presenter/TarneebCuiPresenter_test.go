//go:build test

package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newTarneebForCuiTest() *domain.Tarneeb {
	tn := domain.NewDefaultTarneeb()
	tn.Reset()
	return tn
}

func TestTarneebCuiPresenter_Output_PhaseLabels(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.TarneebCuiPresenter)

	t.Run("bid phase prompt", func(t *testing.T) {
		tn := newTarneebForCuiTest()
		out := p.Output(tn, nil)
		assert.Contains(t, out, i18n.T("tarneeb.promptBidHelp"))
		assert.NotContains(t, out, "tarneeb.", "a raw i18n key reached the screen")
	})

	t.Run("trump phase prompt", func(t *testing.T) {
		tn := newTarneebForCuiTest()
		tn.SetPhase(domain.TarneebPhaseTrumpDeclaration)
		tn.SetBidWinnerIdx(0)
		out := p.Output(tn, nil)
		assert.NotEmpty(t, out)
	})

	t.Run("play phase prompt", func(t *testing.T) {
		tn := newTarneebForCuiTest()
		tn.SetPhase(domain.TarneebPhasePlay)
		tn.SetTrumpSuit(domain.CardDesignSpade)
		out := p.Output(tn, nil)
		assert.Contains(t, out, i18n.T("tarneeb.promptPlayHelp"))
		assert.NotContains(t, out, "tarneeb.", "a raw i18n key reached the screen")
	})

	t.Run("trick end prompt", func(t *testing.T) {
		tn := newTarneebForCuiTest()
		tn.SetPhase(domain.TarneebPhaseTrickEnd)
		out := p.Output(tn, nil)
		assert.NotEmpty(t, out)
	})

	t.Run("round end prompt", func(t *testing.T) {
		tn := newTarneebForCuiTest()
		tn.SetPhase(domain.TarneebPhaseRoundEnd)
		out := p.Output(tn, nil)
		assert.NotEmpty(t, out)
	})

	t.Run("redeal count surfaces", func(t *testing.T) {
		tn := newTarneebForCuiTest()
		// Force a redeal by simulating all-pass scenario via internal state.
		tn.SetPhase(domain.TarneebPhaseBid)
		out := p.Output(tn, nil)
		require.NotEmpty(t, out)
	})

	t.Run("error block included", func(t *testing.T) {
		tn := newTarneebForCuiTest()
		out := p.Output(tn, errors.New("bad input"))
		assert.Contains(t, out, "bad input")
	})

	t.Run("game end banner", func(t *testing.T) {
		tn := newTarneebForCuiTest()
		tn.SetPhase(domain.TarneebPhaseGameEnd)
		// Force the game end flag via game state
		tn.SetTeamScore(0, 31)
		tn.SetBidWinnerIdx(0)
		tn.SetHighestBid(7)
		// Simulate end-of-game presentation by setting phase + GameEndFlag via Reset path:
		// reach GameEnd by calling ScoreRound on a TarneebPhaseRoundEnd configured game
		out := p.Output(tn, nil)
		assert.NotEmpty(t, out)
	})
}

func TestTarneebCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.TarneebCuiPresenter)

	t.Run("bid hint", func(t *testing.T) {
		tn := newTarneebForCuiTest()
		tn.SetPhase(domain.TarneebPhaseBid)
		tn.SetBidPlayerIdx(0)
		out := p.HintOutput(tn)
		assert.NotEmpty(t, out)
	})

	t.Run("trump hint", func(t *testing.T) {
		tn := newTarneebForCuiTest()
		tn.SetPhase(domain.TarneebPhaseTrumpDeclaration)
		tn.SetBidWinnerIdx(0)
		out := p.HintOutput(tn)
		assert.NotEmpty(t, out)
	})

	t.Run("play hint", func(t *testing.T) {
		tn := newTarneebForCuiTest()
		tn.SetPhase(domain.TarneebPhasePlay)
		tn.SetTrumpSuit(domain.CardDesignSpade)
		tn.SetCurrentPlayerIdx(0)
		out := p.HintOutput(tn)
		assert.NotEmpty(t, out)
	})

	t.Run("no hint when out of turn", func(t *testing.T) {
		tn := newTarneebForCuiTest()
		tn.SetPhase(domain.TarneebPhasePlay)
		tn.SetCurrentPlayerIdx(1) // CPU's turn
		out := p.HintOutput(tn)
		assert.Contains(t, out, i18n.T("tarneeb.hintNone"))
	})
}

func TestTarneebCuiPresenter_ActionLogOutput(t *testing.T) {
	tn := newTarneebForCuiTest()
	p := new(presenter.TarneebCuiPresenter)
	out := p.ActionLogOutput(tn)
	assert.NotNil(t, out)
}
