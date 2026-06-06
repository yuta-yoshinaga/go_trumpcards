package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTichuAllCpu() *domain.Tichu {
	players := []*domain.TichuPlayer{
		domain.NewTichuPlayer(false),
		domain.NewTichuPlayer(false),
		domain.NewTichuPlayer(false),
		domain.NewTichuPlayer(false),
	}
	return domain.NewTichu(domain.NewTrumpCards(domain.TichuJokerCount), players, domain.DefaultTichuConfig())
}

func TestTichuWebPresenter_DeclarePhase(t *testing.T) {
	tg := newTichuAllCpu()
	tg.Reset()

	p := new(presenter.TichuWebPresenter)
	var resp controller.TichuWebOutput
	require.NoError(t, json.Unmarshal([]byte(p.Output(tg, nil)), &resp))
	assert.Equal(t, "declare", resp.Phase)
	assert.Len(t, resp.Players, domain.TichuPlayerCnt)
	assert.False(t, resp.GameEndFlag)
	// teams alternate
	assert.Equal(t, 0, resp.Players[0].Team)
	assert.Equal(t, 1, resp.Players[1].Team)
}

func TestTichuWebPresenter_PlayAndEnd(t *testing.T) {
	tg := newTichuAllCpu()
	tg.Reset()
	for tg.GetPhase() == domain.TichuPhaseDeclare {
		tg.CpuPlay()
	}
	p := new(presenter.TichuWebPresenter)
	var play controller.TichuWebOutput
	require.NoError(t, json.Unmarshal([]byte(p.Output(tg, nil)), &play))
	assert.Equal(t, "play", play.Phase)

	for !tg.GetGameEndFlag() {
		tg.CpuPlay()
	}
	var end controller.TichuWebOutput
	require.NoError(t, json.Unmarshal([]byte(p.Output(tg, nil)), &end))
	assert.Equal(t, "end", end.Phase)
	assert.True(t, end.GameEndFlag)
	assert.NotEmpty(t, end.Message)
}

func TestTichuWebPresenter_Error(t *testing.T) {
	tg := newTichuAllCpu()
	tg.Reset()
	p := new(presenter.TichuWebPresenter)
	var resp controller.TichuWebOutput
	require.NoError(t, json.Unmarshal([]byte(p.Output(tg, errors.New("boom"))), &resp))
	assert.Equal(t, "boom", resp.Message)
}

func TestTichuWebPresenter_ActionLog(t *testing.T) {
	tg := newTichuAllCpu()
	tg.Reset()
	p := new(presenter.TichuWebPresenter)
	out := p.ActionLogOutput(tg)
	assert.NotEmpty(t, out)
}
