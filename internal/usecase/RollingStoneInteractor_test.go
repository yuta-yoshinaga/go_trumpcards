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

func newMockRollingStoneGame() *interfaces.MockRollingStoneGame {
	return new(interfaces.MockRollingStoneGame)
}

func newMockRollingStonePresenter() *presenter.MockRollingStonePresenter {
	return new(presenter.MockRollingStonePresenter)
}

func TestNewRollingStoneInteractor(t *testing.T) {
	assert.NotNil(t, NewRollingStoneInteractor(newMockRollingStoneGame(), newMockRollingStonePresenter()))
}

func TestNewRollingStoneInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewRollingStoneInteractor(nil, newMockRollingStonePresenter()) })
	assert.Panics(t, func() { NewRollingStoneInteractor(newMockRollingStoneGame(), nil) })
}

func TestRollingStoneInteractorResetAdvancesToTheHuman(t *testing.T) {
	g := newMockRollingStoneGame()
	p := newMockRollingStonePresenter()
	i := NewRollingStoneInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(false).Times(3)
	g.On("IsHumanTurn").Return(true)
	g.On("CpuPlay").Return()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertNumberOfCalls(t, "CpuPlay", 3)
}

func TestRollingStoneInteractorStopsAtTheHumanTurn(t *testing.T) {
	g := newMockRollingStoneGame()
	p := newMockRollingStonePresenter()
	i := NewRollingStoneInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNotCalled(t, "CpuPlay")
}

// **引き取りも 1 手。** 打った後は同じように CPU を進める。
func TestRollingStoneInteractorPickUpAdvances(t *testing.T) {
	g := newMockRollingStoneGame()
	p := newMockRollingStonePresenter()
	i := NewRollingStoneInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true).Once()
	g.On("PlayerPickUp").Return(nil)
	g.On("IsHumanTurn").Return(false).Once()
	g.On("CpuPlay").Return()
	g.On("IsHumanTurn").Return(true)
	p.On("Output", g, nil).Return("picked")

	assert.Equal(t, "picked", i.PickUp())
	g.AssertNumberOfCalls(t, "PlayerPickUp", 1)
	g.AssertNumberOfCalls(t, "CpuPlay", 1)
}

func TestRollingStoneInteractorSurfacesErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		method string
		call   func(*RollingStoneInteractor) string
	}{
		{"play", "PlayerPlay", func(i *RollingStoneInteractor) string { return i.Play(3) }},
		{"pickup", "PlayerPickUp", func(i *RollingStoneInteractor) string { return i.PickUp() }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockRollingStoneGame()
			p := newMockRollingStonePresenter()
			i := NewRollingStoneInteractor(g, p)

			err := errors.New("must follow the led suit")
			g.On("GetGameEndFlag").Return(false)
			g.On("IsHumanTurn").Return(true)
			g.On("PlayerPlay", 3).Return(err)
			g.On("PlayerPickUp").Return(err)
			p.On("Output", g, err).Return("error_output")

			assert.Equal(t, "error_output", tc.call(i))
			g.AssertNotCalled(t, "CpuPlay")
		})
	}
}

func TestRollingStoneInteractorGuardsOnTurn(t *testing.T) {
	for _, tc := range []struct {
		name   string
		call   func(*RollingStoneInteractor) string
		method string
	}{
		{"play", func(i *RollingStoneInteractor) string { return i.Play(0) }, "PlayerPlay"},
		{"pickup", func(i *RollingStoneInteractor) string { return i.PickUp() }, "PlayerPickUp"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockRollingStoneGame()
			p := newMockRollingStonePresenter()
			i := NewRollingStoneInteractor(g, p)

			g.On("GetGameEndFlag").Return(false)
			g.On("IsHumanTurn").Return(false)
			p.On("Output", g, nil).Return("blocked")

			assert.Equal(t, "blocked", tc.call(i))
			g.AssertNotCalled(t, tc.method)
		})
	}
}

func TestRollingStoneInteractorGuardsAfterGameEnd(t *testing.T) {
	for _, tc := range []struct {
		name   string
		call   func(*RollingStoneInteractor) string
		method string
	}{
		{"play", func(i *RollingStoneInteractor) string { return i.Play(0) }, "PlayerPlay"},
		{"pickup", func(i *RollingStoneInteractor) string { return i.PickUp() }, "PlayerPickUp"},
		{"giveup", func(i *RollingStoneInteractor) string { return i.GiveUp() }, "GiveUp"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockRollingStoneGame()
			p := newMockRollingStonePresenter()
			i := NewRollingStoneInteractor(g, p)

			g.On("GetGameEndFlag").Return(true)
			g.On("IsHumanTurn").Return(false).Maybe()
			p.On("Output", g, nil).Return("ended")

			assert.Equal(t, "ended", tc.call(i))
			g.AssertNotCalled(t, tc.method)
		})
	}
}

func TestRollingStoneInteractorGiveUp(t *testing.T) {
	g := newMockRollingStoneGame()
	p := newMockRollingStonePresenter()
	i := NewRollingStoneInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.GiveUp())
	g.AssertNumberOfCalls(t, "GiveUp", 1)
}

func TestRollingStoneInteractorResetWithConfig(t *testing.T) {
	g := newMockRollingStoneGame()
	p := newMockRollingStonePresenter()
	i := NewRollingStoneInteractor(g, p)

	p.On("Output", g, mock.Anything).Return("cfg_error")
	assert.Equal(t, "cfg_error", i.ResetWithConfig(domain.RollingStoneConfig{PlayerCnt: 9}))
	g.AssertNotCalled(t, "SetConfig")
	g.AssertNotCalled(t, "Reset")

	g2 := newMockRollingStoneGame()
	p2 := newMockRollingStonePresenter()
	i2 := NewRollingStoneInteractor(g2, p2)
	cfg := domain.RollingStoneConfig{PlayerCnt: 6}
	g2.On("SetConfig", cfg).Return()
	g2.On("Reset").Return()
	g2.On("GetGameEndFlag").Return(false)
	g2.On("IsHumanTurn").Return(true)
	p2.On("Output", g2, nil).Return("reset")

	assert.Equal(t, "reset", i2.ResetWithConfig(cfg))
	g2.AssertCalled(t, "SetConfig", cfg)
}

func TestRollingStoneInteractorHintAndLogDelegate(t *testing.T) {
	g := newMockRollingStoneGame()
	p := newMockRollingStonePresenter()
	i := NewRollingStoneInteractor(g, p)

	p.On("HintOutput", g).Return("hint")
	p.On("ActionLogOutput", g).Return("log")

	assert.Equal(t, "hint", i.Hint())
	assert.Equal(t, "log", i.ActionLog())
}

func TestRollingStoneInteractorGetConfigDelegates(t *testing.T) {
	g := newMockRollingStoneGame()
	p := newMockRollingStonePresenter()
	i := NewRollingStoneInteractor(g, p)

	cfg := domain.RollingStoneConfig{PlayerCnt: 5}
	g.On("GetConfig").Return(cfg)
	assert.Equal(t, cfg, i.GetConfig())
}

func TestRestoreRollingStoneInteractor(t *testing.T) {
	src := domain.NewDefaultRollingStone()
	src.Reset()
	data, err := json.Marshal(src)
	require.NoError(t, err)

	restored, err := RestoreRollingStoneInteractor(data, new(presenter.MockRollingStonePresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
	assert.Equal(t, src.GetPlayerCnt(), restored.Game.GetPlayerCnt())
	assert.Equal(t, src.GetDeckSize(), restored.Game.GetDeckSize())
	assert.Equal(t, src.GetPlayer(0).GetCardsSize(), restored.Game.GetPlayer(0).GetCardsSize())

	_, err = RestoreRollingStoneInteractor([]byte("{"), new(presenter.MockRollingStonePresenter))
	assert.Error(t, err)
}
