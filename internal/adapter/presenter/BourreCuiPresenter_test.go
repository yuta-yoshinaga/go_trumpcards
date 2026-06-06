package presenter_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestBourreCuiPresenter_Output(t *testing.T) {
	bg := newBourreAllCpu()
	bg.Reset()
	p := new(presenter.BourreCuiPresenter)

	// Decide phase render
	out := p.Output(bg, nil)
	assert.Contains(t, out, "==========")

	// Drive to game end, rendering along the way
	for i := 0; i < 300000 && !bg.GetGameEndFlag(); i++ {
		if bg.GetPhase() == domain.BourrePhaseRoundEnd {
			bg.NextHand()
			continue
		}
		_ = p.Output(bg, nil)
		bg.CpuPlay()
	}
	endOut := p.Output(bg, nil)
	assert.Contains(t, endOut, "==========")
	assert.NotEmpty(t, strings.TrimSpace(endOut))
}

func TestBourreCuiPresenter_Error(t *testing.T) {
	bg := newBourreAllCpu()
	bg.Reset()
	p := new(presenter.BourreCuiPresenter)
	out := p.Output(bg, assertErr{})
	assert.Contains(t, out, "boom")
}

type assertErr struct{}

func (assertErr) Error() string { return "boom" }

func TestBourreCuiPresenter_ActionLog(t *testing.T) {
	bg := newBourreAllCpu()
	bg.Reset()
	p := new(presenter.BourreCuiPresenter)
	assert.NotNil(t, p.ActionLogOutput(bg))
}
