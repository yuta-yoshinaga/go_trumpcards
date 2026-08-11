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

func newMockPasurGame() *interfaces.MockPasurGame { return new(interfaces.MockPasurGame) }

func newMockPasurPresenter() *presenter.MockPasurPresenter {
	return new(presenter.MockPasurPresenter)
}

func TestNewPasurInteractor(t *testing.T) {
	assert.NotNil(t, NewPasurInteractor(newMockPasurGame(), newMockPasurPresenter()))
}

func TestNewPasurInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewPasurInteractor(nil, newMockPasurPresenter()) })
	assert.Panics(t, func() { NewPasurInteractor(newMockPasurGame(), nil) })
}

// **人間の出番まで CPU を進める。** 途中に段は無い。
func TestPasurInteractorResetAdvancesToTheHuman(t *testing.T) {
	g := newMockPasurGame()
	p := newMockPasurPresenter()
	i := NewPasurInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(false).Times(3)
	g.On("IsHumanTurn").Return(true)
	g.On("CpuPlay").Return()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertNumberOfCalls(t, "CpuPlay", 3)
}

func TestPasurInteractorStopsAtTheHumanTurn(t *testing.T) {
	g := newMockPasurGame()
	p := newMockPasurPresenter()
	i := NewPasurInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNotCalled(t, "CpuPlay")
}

// **終局したら回さない。**
func TestPasurInteractorStopsWhenTheGameEnds(t *testing.T) {
	g := newMockPasurGame()
	p := newMockPasurPresenter()
	i := NewPasurInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(true)
	g.On("IsHumanTurn").Return(false).Maybe()
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNotCalled(t, "CpuPlay")
}

func TestPasurInteractorPlayRejectsAndDoesNotAdvance(t *testing.T) {
	g := newMockPasurGame()
	p := newMockPasurPresenter()
	i := NewPasurInteractor(g, p)

	playErr := errors.New("must capture when a capture is available")
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	g.On("PlayerPlay", 3, []int(nil)).Return(playErr)
	p.On("Output", g, playErr).Return("play_error")

	assert.Equal(t, "play_error", i.Play(3, nil))
	g.AssertNotCalled(t, "CpuPlay")
}

// **場札の指定はそのまま渡す。** 空はトレールという意味を持つ。
func TestPasurInteractorPlayPassesTheTableSelectionThrough(t *testing.T) {
	for _, tc := range []struct {
		name  string
		table []int
	}{
		{"取る", []int{0, 2}},
		{"置く（トレール）", nil},
		{"空スライスも置く", []int{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockPasurGame()
			p := newMockPasurPresenter()
			i := NewPasurInteractor(g, p)

			g.On("GetGameEndFlag").Return(false)
			g.On("IsHumanTurn").Return(true).Once()
			g.On("PlayerPlay", 1, tc.table).Return(nil)
			g.On("IsHumanTurn").Return(false).Once()
			g.On("CpuPlay").Return()
			g.On("IsHumanTurn").Return(true)
			p.On("Output", g, nil).Return("played")

			assert.Equal(t, "played", i.Play(1, tc.table))
			g.AssertCalled(t, "PlayerPlay", 1, tc.table)
		})
	}
}

func TestPasurInteractorPlayGuardsOnTurn(t *testing.T) {
	g := newMockPasurGame()
	p := newMockPasurPresenter()
	i := NewPasurInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(false)
	p.On("Output", g, nil).Return("blocked")

	assert.Equal(t, "blocked", i.Play(0, nil))
	g.AssertNotCalled(t, "PlayerPlay")
}

func TestPasurInteractorGuardsAfterGameEnd(t *testing.T) {
	for _, tc := range []struct {
		name   string
		call   func(*PasurInteractor) string
		method string
	}{
		{"play", func(i *PasurInteractor) string { return i.Play(0, nil) }, "PlayerPlay"},
		{"giveup", func(i *PasurInteractor) string { return i.GiveUp() }, "GiveUp"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockPasurGame()
			p := newMockPasurPresenter()
			i := NewPasurInteractor(g, p)

			g.On("GetGameEndFlag").Return(true)
			g.On("IsHumanTurn").Return(false).Maybe()
			p.On("Output", g, nil).Return("ended")

			assert.Equal(t, "ended", tc.call(i))
			g.AssertNotCalled(t, tc.method)
		})
	}
}

func TestPasurInteractorGiveUp(t *testing.T) {
	g := newMockPasurGame()
	p := newMockPasurPresenter()
	i := NewPasurInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.GiveUp())
	g.AssertNumberOfCalls(t, "GiveUp", 1)
}

// **人数が範囲外の設定は弾き、ゲームを作り直さない。**
func TestPasurInteractorResetWithConfig(t *testing.T) {
	g := newMockPasurGame()
	p := newMockPasurPresenter()
	i := NewPasurInteractor(g, p)

	p.On("Output", g, mock.Anything).Return("cfg_error")
	assert.Equal(t, "cfg_error", i.ResetWithConfig(domain.PasurConfig{PlayerCnt: 9}))
	g.AssertNotCalled(t, "SetConfig")
	g.AssertNotCalled(t, "Reset")

	g2 := newMockPasurGame()
	p2 := newMockPasurPresenter()
	i2 := NewPasurInteractor(g2, p2)
	cfg := domain.PasurConfig{PlayerCnt: 3}
	g2.On("SetConfig", cfg).Return()
	g2.On("Reset").Return()
	g2.On("GetGameEndFlag").Return(false)
	g2.On("IsHumanTurn").Return(true)
	p2.On("Output", g2, nil).Return("reset")

	assert.Equal(t, "reset", i2.ResetWithConfig(cfg))
	g2.AssertCalled(t, "SetConfig", cfg)
}

func TestPasurInteractorHintAndLogDelegate(t *testing.T) {
	g := newMockPasurGame()
	p := newMockPasurPresenter()
	i := NewPasurInteractor(g, p)

	p.On("HintOutput", g).Return("hint")
	p.On("ActionLogOutput", g).Return("log")

	assert.Equal(t, "hint", i.Hint())
	assert.Equal(t, "log", i.ActionLog())
}

func TestPasurInteractorGetConfigDelegates(t *testing.T) {
	g := newMockPasurGame()
	p := newMockPasurPresenter()
	i := NewPasurInteractor(g, p)

	cfg := domain.PasurConfig{PlayerCnt: 2}
	g.On("GetConfig").Return(cfg)
	assert.Equal(t, cfg, i.GetConfig())
}

func TestRestorePasurInteractor(t *testing.T) {
	src := domain.NewDefaultPasur()
	src.Reset()
	data, err := json.Marshal(src)
	require.NoError(t, err)

	restored, err := RestorePasurInteractor(data, new(presenter.MockPasurPresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
	assert.Equal(t, src.GetPlayerCnt(), restored.Game.GetPlayerCnt())
	assert.Len(t, restored.Game.GetTableCards(), len(src.GetTableCards()), "場の札が消えない")
	assert.Equal(t, src.GetDeckRemaining(), restored.Game.GetDeckRemaining())

	_, err = RestorePasurInteractor([]byte("{"), new(presenter.MockPasurPresenter))
	assert.Error(t, err)
}
