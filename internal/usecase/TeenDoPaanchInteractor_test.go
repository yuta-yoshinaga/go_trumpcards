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

func newMockTeenDoPaanchGame() *interfaces.MockTeenDoPaanchGame {
	return new(interfaces.MockTeenDoPaanchGame)
}

func newMockTeenDoPaanchPresenter() *presenter.MockTeenDoPaanchPresenter {
	return new(presenter.MockTeenDoPaanchPresenter)
}

func TestNewTeenDoPaanchInteractor(t *testing.T) {
	assert.NotNil(t, NewTeenDoPaanchInteractor(newMockTeenDoPaanchGame(), newMockTeenDoPaanchPresenter()))
}

func TestNewTeenDoPaanchInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewTeenDoPaanchInteractor(nil, newMockTeenDoPaanchPresenter()) })
	assert.Panics(t, func() { NewTeenDoPaanchInteractor(newMockTeenDoPaanchGame(), nil) })
}

// **切り札を決めるのがノルマ 5 の席で、それが CPU なら宣言まで進める。**
// 宣言で止めると人間が打てない盤面を返してしまう。
func TestTeenDoPaanchInteractorResetWalksTrumpThenPlay(t *testing.T) {
	g := newMockTeenDoPaanchGame()
	p := newMockTeenDoPaanchPresenter()
	i := NewTeenDoPaanchInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.TeenDoPaanchPhaseTrump).Once()
	g.On("IsHumanTrumpTurn").Return(false).Once()
	g.On("CpuDeclareTrump").Return()
	g.On("GetPhase").Return(domain.TeenDoPaanchPhasePlay)
	g.On("IsHumanTurn").Return(false).Twice()
	g.On("IsHumanTurn").Return(true)
	g.On("CpuPlay").Return()
	g.On("GetHint").Return(nil).Maybe()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertNumberOfCalls(t, "CpuDeclareTrump", 1)
	g.AssertNumberOfCalls(t, "CpuPlay", 2)
}

// **人間がノルマ 5 なら宣言フェーズで止まる。** 勝手に決めない。
func TestTeenDoPaanchInteractorStopsAtTheHumanTrumpCall(t *testing.T) {
	g := newMockTeenDoPaanchGame()
	p := newMockTeenDoPaanchPresenter()
	i := NewTeenDoPaanchInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.TeenDoPaanchPhaseTrump)
	g.On("IsHumanTrumpTurn").Return(true)
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNotCalled(t, "CpuDeclareTrump")
	g.AssertNotCalled(t, "CpuPlay")
}

// **ラウンド終了では止める。** 次のラウンドは next で明示的に始める。
func TestTeenDoPaanchInteractorStopsAtRoundEnd(t *testing.T) {
	g := newMockTeenDoPaanchGame()
	p := newMockTeenDoPaanchPresenter()
	i := NewTeenDoPaanchInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.TeenDoPaanchPhaseRoundEnd)
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNotCalled(t, "CpuPlay")
}

// **範囲外のスートはドメインが弾き、その理由をそのまま返す。**
func TestTeenDoPaanchInteractorDeclareTrumpSurfacesErrors(t *testing.T) {
	g := newMockTeenDoPaanchGame()
	p := newMockTeenDoPaanchPresenter()
	i := NewTeenDoPaanchInteractor(g, p)

	declErr := errors.New("invalid suit: 9")
	g.On("GetGameEndFlag").Return(false)
	g.On("PlayerDeclareTrump", 9).Return(declErr)
	p.On("Output", g, declErr).Return("trump_error")

	assert.Equal(t, "trump_error", i.DeclareTrump(9))
	g.AssertNotCalled(t, "CpuPlay")
}

func TestTeenDoPaanchInteractorDeclareTrumpAdvances(t *testing.T) {
	g := newMockTeenDoPaanchGame()
	p := newMockTeenDoPaanchPresenter()
	i := NewTeenDoPaanchInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("PlayerDeclareTrump", 3).Return(nil)
	g.On("GetPhase").Return(domain.TeenDoPaanchPhasePlay)
	g.On("IsHumanTurn").Return(false).Once()
	g.On("IsHumanTurn").Return(true)
	g.On("CpuPlay").Return()
	p.On("Output", g, nil).Return("declared")

	assert.Equal(t, "declared", i.DeclareTrump(3))
	g.AssertNumberOfCalls(t, "CpuPlay", 1)
}

func TestTeenDoPaanchInteractorPlayRejectsAndDoesNotAdvance(t *testing.T) {
	g := newMockTeenDoPaanchGame()
	p := newMockTeenDoPaanchPresenter()
	i := NewTeenDoPaanchInteractor(g, p)

	playErr := errors.New("must follow the led suit")
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.TeenDoPaanchPhasePlay)
	g.On("IsHumanTurn").Return(true)
	g.On("PlayerPlay", 3).Return(playErr)
	p.On("Output", g, playErr).Return("play_error")

	assert.Equal(t, "play_error", i.Play(3))
	g.AssertNotCalled(t, "CpuPlay")
}

func TestTeenDoPaanchInteractorPlayGuardsOnTurn(t *testing.T) {
	g := newMockTeenDoPaanchGame()
	p := newMockTeenDoPaanchPresenter()
	i := NewTeenDoPaanchInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.TeenDoPaanchPhaseTrump)
	g.On("IsHumanTurn").Return(false).Maybe()
	p.On("Output", g, nil).Return("blocked")

	assert.Equal(t, "blocked", i.Play(0))
	g.AssertNotCalled(t, "PlayerPlay")
}

func TestTeenDoPaanchInteractorGuardsAfterGameEnd(t *testing.T) {
	for _, tc := range []struct {
		name   string
		call   func(*TeenDoPaanchInteractor) string
		method string
	}{
		{"trump", func(i *TeenDoPaanchInteractor) string { return i.DeclareTrump(1) }, "PlayerDeclareTrump"},
		{"next", func(i *TeenDoPaanchInteractor) string { return i.NextRound() }, "NextRound"},
		{"giveup", func(i *TeenDoPaanchInteractor) string { return i.GiveUp() }, "GiveUp"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockTeenDoPaanchGame()
			p := newMockTeenDoPaanchPresenter()
			i := NewTeenDoPaanchInteractor(g, p)

			g.On("GetGameEndFlag").Return(true)
			p.On("Output", g, nil).Return("ended")

			assert.Equal(t, "ended", tc.call(i))
			g.AssertNotCalled(t, tc.method, mock.Anything)
			g.AssertNotCalled(t, tc.method)
		})
	}
}

func TestTeenDoPaanchInteractorNextRoundAdvances(t *testing.T) {
	g := newMockTeenDoPaanchGame()
	p := newMockTeenDoPaanchPresenter()
	i := NewTeenDoPaanchInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("NextRound").Return()
	g.On("GetPhase").Return(domain.TeenDoPaanchPhaseTrump)
	g.On("IsHumanTrumpTurn").Return(true)
	p.On("Output", g, nil).Return("next")

	assert.Equal(t, "next", i.NextRound())
	g.AssertCalled(t, "NextRound")
}

func TestTeenDoPaanchInteractorResetWithConfig(t *testing.T) {
	g := newMockTeenDoPaanchGame()
	p := newMockTeenDoPaanchPresenter()
	i := NewTeenDoPaanchInteractor(g, p)

	cfg := domain.TeenDoPaanchConfig{Rounds: 6}
	g.On("SetConfig", cfg).Return()
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.TeenDoPaanchPhaseTrump)
	g.On("IsHumanTrumpTurn").Return(true)
	p.On("Output", g, nil).Return("configured")

	assert.Equal(t, "configured", i.ResetWithConfig(cfg))
	g.AssertCalled(t, "SetConfig", cfg)
}

func TestTeenDoPaanchInteractorResetWithInvalidConfig(t *testing.T) {
	for _, n := range []int{domain.TeenDoPaanchRoundsMin - 1, domain.TeenDoPaanchRoundsMax + 1} {
		g := newMockTeenDoPaanchGame()
		p := newMockTeenDoPaanchPresenter()
		i := NewTeenDoPaanchInteractor(g, p)

		p.On("Output", mock.Anything, mock.Anything).Return("cfg_error")

		assert.Equal(t, "cfg_error", i.ResetWithConfig(domain.TeenDoPaanchConfig{Rounds: n}))
		g.AssertNotCalled(t, "Reset")
		g.AssertNotCalled(t, "SetConfig", mock.Anything)
	}
}

func TestTeenDoPaanchInteractorHintAndLogDelegate(t *testing.T) {
	g := newMockTeenDoPaanchGame()
	p := newMockTeenDoPaanchPresenter()
	i := NewTeenDoPaanchInteractor(g, p)

	p.On("HintOutput", g).Return("hint")
	p.On("ActionLogOutput", g).Return("log")

	assert.Equal(t, "hint", i.Hint())
	assert.Equal(t, "log", i.ActionLog())
}

func TestTeenDoPaanchInteractorGetConfigDelegates(t *testing.T) {
	g := newMockTeenDoPaanchGame()
	p := newMockTeenDoPaanchPresenter()
	i := NewTeenDoPaanchInteractor(g, p)

	cfg := domain.TeenDoPaanchConfig{Rounds: 9}
	g.On("GetConfig").Return(cfg)
	assert.Equal(t, cfg, i.GetConfig())
}

func TestRestoreTeenDoPaanchInteractor(t *testing.T) {
	src := domain.NewDefaultTeenDoPaanch()
	src.Reset()
	data, err := json.Marshal(src)
	require.NoError(t, err)

	restored, err := RestoreTeenDoPaanchInteractor(data, new(presenter.MockTeenDoPaanchPresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
	assert.Equal(t, src.GetRoundNumber(), restored.Game.GetRoundNumber())
	assert.Equal(t, src.GetFivePlayerIdx(), restored.Game.GetFivePlayerIdx())
	for i := range domain.TeenDoPaanchPlayerCnt {
		assert.Equal(t, src.GetPlayer(i).GetTarget(), restored.Game.GetPlayer(i).GetTarget(), "ノルマが消えない")
	}

	_, err = RestoreTeenDoPaanchInteractor([]byte("{"), new(presenter.MockTeenDoPaanchPresenter))
	assert.Error(t, err)
}
