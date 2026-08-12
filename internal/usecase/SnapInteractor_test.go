//go:build test

package usecase

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockSnapGame() *interfaces.MockSnapGame { return new(interfaces.MockSnapGame) }

func newMockSnapPresenter() *presenter.MockSnapPresenter { return new(presenter.MockSnapPresenter) }

func TestNewSnapInteractor(t *testing.T) {
	assert.NotNil(t, NewSnapInteractor(newMockSnapGame(), newMockSnapPresenter()))
}

func TestNewSnapInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewSnapInteractor(nil, newMockSnapPresenter()) })
	assert.Panics(t, func() { NewSnapInteractor(newMockSnapGame(), nil) })
}

// **Reset は CPU を進めない。** 反射ゲームなので、時間が動くのは Tick だけ。
func TestSnapInteractorResetDoesNotRunTheCpu(t *testing.T) {
	g := newMockSnapGame()
	p := newMockSnapPresenter()
	i := NewSnapInteractor(g, p)

	g.On("Reset").Return()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertNotCalled(t, "Tick")
}

// **宣言は CPU を先に進めない。** ここで Tick を回すと、人間の宣言より前に
// CPU の予約が発火して反射ゲームとして成立しない。
func TestSnapInteractorSnapDoesNotTickFirst(t *testing.T) {
	g := newMockSnapGame()
	p := newMockSnapPresenter()
	i := NewSnapInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("PlayerSnap").Return(nil)
	p.On("Output", g, nil).Return("snapped")

	assert.Equal(t, "snapped", i.Snap())
	g.AssertNotCalled(t, "Tick")
	g.AssertNumberOfCalls(t, "PlayerSnap", 1)
}

func TestSnapInteractorStepAndSnapSurfaceErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		method string
		call   func(*SnapInteractor) string
	}{
		{"step", "PlayerStep", func(i *SnapInteractor) string { return i.Step() }},
		{"snap", "PlayerSnap", func(i *SnapInteractor) string { return i.Snap() }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockSnapGame()
			p := newMockSnapPresenter()
			i := NewSnapInteractor(g, p)

			err := errors.New("not your turn")
			g.On("GetGameEndFlag").Return(false)
			g.On(tc.method).Return(err)
			p.On("Output", g, err).Return("error_output")

			assert.Equal(t, "error_output", tc.call(i))
		})
	}
}

func TestSnapInteractorTickDelegates(t *testing.T) {
	g := newMockSnapGame()
	p := newMockSnapPresenter()
	i := NewSnapInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("Tick").Return(domain.SnapPendingSnap)
	p.On("Output", g, nil).Return("ticked")

	assert.Equal(t, "ticked", i.Tick())
	g.AssertNumberOfCalls(t, "Tick", 1)
}

func TestSnapInteractorGuardsAfterGameEnd(t *testing.T) {
	for _, tc := range []struct {
		name   string
		call   func(*SnapInteractor) string
		method string
	}{
		{"step", func(i *SnapInteractor) string { return i.Step() }, "PlayerStep"},
		{"snap", func(i *SnapInteractor) string { return i.Snap() }, "PlayerSnap"},
		{"tick", func(i *SnapInteractor) string { return i.Tick() }, "Tick"},
		{"giveup", func(i *SnapInteractor) string { return i.GiveUp() }, "GiveUp"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockSnapGame()
			p := newMockSnapPresenter()
			i := NewSnapInteractor(g, p)

			g.On("GetGameEndFlag").Return(true)
			p.On("Output", g, nil).Return("ended")

			assert.Equal(t, "ended", tc.call(i))
			g.AssertNotCalled(t, tc.method)
		})
	}
}

func TestSnapInteractorStepAndGiveUp(t *testing.T) {
	g := newMockSnapGame()
	p := newMockSnapPresenter()
	i := NewSnapInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("PlayerStep").Return(nil)
	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Step())
	assert.Equal(t, "out", i.GiveUp())
	g.AssertNumberOfCalls(t, "PlayerStep", 1)
	g.AssertNumberOfCalls(t, "GiveUp", 1)
}

// **範囲外の設定は弾き、ゲームを作り直さない。**
func TestSnapInteractorResetWithConfig(t *testing.T) {
	g := newMockSnapGame()
	p := newMockSnapPresenter()
	i := NewSnapInteractor(g, p)

	p.On("Output", g, mock.Anything).Return("cfg_error")
	assert.Equal(t, "cfg_error", i.ResetWithConfig(domain.SnapConfig{PlayerCnt: 9}))
	g.AssertNotCalled(t, "SetConfig")
	g.AssertNotCalled(t, "Reset")

	g2 := newMockSnapGame()
	p2 := newMockSnapPresenter()
	i2 := NewSnapInteractor(g2, p2)
	cfg := domain.SnapConfig{PlayerCnt: 3, CpuDifficulty: domain.SnapCpuHard}
	g2.On("SetConfig", cfg).Return()
	g2.On("Reset").Return()
	p2.On("Output", g2, nil).Return("reset")

	assert.Equal(t, "reset", i2.ResetWithConfig(cfg))
	g2.AssertCalled(t, "SetConfig", cfg)
}

func TestSnapInteractorHintAndLogDelegate(t *testing.T) {
	g := newMockSnapGame()
	p := newMockSnapPresenter()
	i := NewSnapInteractor(g, p)

	p.On("HintOutput", g).Return("hint")
	p.On("ActionLogOutput", g).Return("log")

	assert.Equal(t, "hint", i.Hint())
	assert.Equal(t, "log", i.ActionLog())
}

func TestSnapInteractorGetConfigDelegates(t *testing.T) {
	g := newMockSnapGame()
	p := newMockSnapPresenter()
	i := NewSnapInteractor(g, p)

	cfg := domain.SnapConfig{PlayerCnt: 4, CpuDifficulty: domain.SnapCpuEasy}
	g.On("GetConfig").Return(cfg)
	assert.Equal(t, cfg, i.GetConfig())
}

func TestRestoreSnapInteractor(t *testing.T) {
	src := domain.NewDefaultSnap()
	src.Reset()
	data, err := json.Marshal(src)
	require.NoError(t, err)

	restored, err := RestoreSnapInteractor(data, new(presenter.MockSnapPresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
	assert.Equal(t, src.GetPlayerCnt(), restored.Game.GetPlayerCnt())
	assert.Equal(t, src.GetPlayer(0).GetStockSize(), restored.Game.GetPlayer(0).GetStockSize())

	_, err = RestoreSnapInteractor([]byte("{"), new(presenter.MockSnapPresenter))
	assert.Error(t, err)
}
