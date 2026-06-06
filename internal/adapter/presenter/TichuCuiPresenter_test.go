package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestTichuCuiPresenter_Output(t *testing.T) {
	tg := newTichuAllCpu()
	tg.Reset()
	p := new(presenter.TichuCuiPresenter)

	declare := p.Output(tg, nil)
	assert.NotEmpty(t, declare)

	for tg.GetPhase() == domain.TichuPhaseDeclare {
		tg.CpuPlay()
	}
	play := p.Output(tg, nil)
	assert.Contains(t, play, "----------")

	for !tg.GetGameEndFlag() {
		tg.CpuPlay()
	}
	end := p.Output(tg, nil)
	assert.NotEmpty(t, end)

	withErr := p.Output(tg, errors.New("boom"))
	assert.True(t, strings.Contains(withErr, "boom"))
}

func TestTichuCuiPresenter_ActionLog(t *testing.T) {
	tg := newTichuAllCpu()
	tg.Reset()
	p := new(presenter.TichuCuiPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(tg))
}
