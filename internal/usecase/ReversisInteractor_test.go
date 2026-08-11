package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockReversisGame() *interfaces.MockReversisGame { return new(interfaces.MockReversisGame) }

func newMockReversisPresenter() *presenter.MockReversisPresenter {
	return new(presenter.MockReversisPresenter)
}

func TestNewReversisInteractor(t *testing.T) {
	assert.NotNil(t, NewReversisInteractor(newMockReversisGame(), newMockReversisPresenter()))
}

func TestNewReversisInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewReversisInteractor(nil, newMockReversisPresenter()) })
	assert.Panics(t, func() { NewReversisInteractor(newMockReversisGame(), nil) })
}

func TestReversisInteractorReset(t *testing.T) {
	g := newMockReversisGame()
	p := newMockReversisPresenter()
	i := NewReversisInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertCalled(t, "Reset")
	g.AssertNotCalled(t, "CpuPlay")
}

func TestReversisInteractorRunsCpuUntilHumanTurn(t *testing.T) {
	g := newMockReversisGame()
	p := newMockReversisPresenter()
	i := NewReversisInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(false).Times(3)
	g.On("IsHumanTurn").Return(true)
	g.On("GetPhase").Return(domain.ReversisPhasePlay)
	g.On("CpuPlay").Return()
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNumberOfCalls(t, "CpuPlay", 3)
}

// **ラウンド終了で必ず止まる。** 誰がプールを取ったかを見せないまま
// 次を配ってしまうと、このゲームの見どころが飛ぶ。
func TestReversisInteractorStopsAtRoundEnd(t *testing.T) {
	g := newMockReversisGame()
	p := newMockReversisPresenter()
	i := NewReversisInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(false)
	g.On("GetPhase").Return(domain.ReversisPhaseRoundEnd)
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNotCalled(t, "CpuPlay")
	g.AssertNotCalled(t, "NextRound")
}

func TestReversisInteractorCpuLoopHasAnUpperBound(t *testing.T) {
	g := newMockReversisGame()
	p := newMockReversisPresenter()
	i := NewReversisInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(false)
	g.On("GetPhase").Return(domain.ReversisPhasePlay)
	g.On("CpuPlay").Return()
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNumberOfCalls(t, "CpuPlay", maxCpuTurnsPerCall)
}

func TestReversisInteractorPlay(t *testing.T) {
	g := newMockReversisGame()
	p := newMockReversisPresenter()
	i := NewReversisInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	g.On("PlayerPlay", 2).Return(nil)
	p.On("Output", g, nil).Return("played")

	assert.Equal(t, "played", i.Play(2))
	g.AssertCalled(t, "PlayerPlay", 2)
}

func TestReversisInteractorPlayError(t *testing.T) {
	g := newMockReversisGame()
	p := newMockReversisPresenter()
	i := NewReversisInteractor(g, p)

	err := errors.New("must follow suit")
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	g.On("PlayerPlay", 2).Return(err)
	p.On("Output", g, err).Return("error_output")

	assert.Equal(t, "error_output", i.Play(2))
	g.AssertNotCalled(t, "CpuPlay")
}

func TestReversisInteractorPlayBlocked(t *testing.T) {
	for _, tc := range []struct {
		name    string
		ended   bool
		isHuman bool
	}{
		{"game over", true, true},
		{"not human turn", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockReversisGame()
			p := newMockReversisPresenter()
			i := NewReversisInteractor(g, p)

			g.On("GetGameEndFlag").Return(tc.ended)
			g.On("IsHumanTurn").Return(tc.isHuman)
			p.On("Output", g, nil).Return("blocked")

			assert.Equal(t, "blocked", i.Play(0))
			g.AssertNotCalled(t, "PlayerPlay", 0)
		})
	}
}

func TestReversisInteractorNextRound(t *testing.T) {
	g := newMockReversisGame()
	p := newMockReversisPresenter()
	i := NewReversisInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("NextRound").Return()
	g.On("IsHumanTurn").Return(true)
	p.On("Output", g, nil).Return("next")

	assert.Equal(t, "next", i.NextRound())
	g.AssertCalled(t, "NextRound")
}

func TestReversisInteractorGuardsAfterGameEnd(t *testing.T) {
	for _, tc := range []struct {
		name   string
		call   func(*ReversisInteractor) string
		method string
	}{
		{"next round", func(i *ReversisInteractor) string { return i.NextRound() }, "NextRound"},
		{"give up", func(i *ReversisInteractor) string { return i.GiveUp() }, "GiveUp"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockReversisGame()
			p := newMockReversisPresenter()
			i := NewReversisInteractor(g, p)

			g.On("GetGameEndFlag").Return(true)
			p.On("Output", g, nil).Return("over")

			assert.Equal(t, "over", tc.call(i))
			g.AssertNotCalled(t, tc.method)
		})
	}
}

func TestReversisInteractorGiveUp(t *testing.T) {
	g := newMockReversisGame()
	p := newMockReversisPresenter()
	i := NewReversisInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("gave_up")

	assert.Equal(t, "gave_up", i.GiveUp())
	g.AssertCalled(t, "GiveUp")
}

func TestReversisInteractorGetConfig(t *testing.T) {
	g := newMockReversisGame()
	i := NewReversisInteractor(g, newMockReversisPresenter())
	cfg := domain.ReversisConfig{Rounds: 6}
	g.On("GetConfig").Return(cfg)
	assert.Equal(t, cfg, i.GetConfig())
}

func TestReversisInteractorResetWithInvalidConfig(t *testing.T) {
	g := newMockReversisGame()
	p := newMockReversisPresenter()
	i := NewReversisInteractor(g, p)

	p.On("Output", mock.Anything, mock.Anything).Return("cfg_error")

	assert.Equal(t, "cfg_error", i.ResetWithConfig(domain.ReversisConfig{Rounds: 0}))
	g.AssertNotCalled(t, "Reset")
	g.AssertNotCalled(t, "SetConfig", mock.Anything)
}

func TestReversisInteractorHintAndActionLog(t *testing.T) {
	g := newMockReversisGame()
	p := newMockReversisPresenter()
	i := NewReversisInteractor(g, p)

	p.On("HintOutput", g).Return("hint_output")
	p.On("ActionLogOutput", g).Return("log_output")

	assert.Equal(t, "hint_output", i.Hint())
	assert.Equal(t, "log_output", i.ActionLog())
}

// Worker はリクエストごとに KV から作り直す。**チップとプールが往復しないと
// 賭けが成立しない** (#4478)。
func TestReversisInteractorSnapshotRoundTrip(t *testing.T) {
	g := domain.NewDefaultReversis()
	g.Reset()
	g.GetPlayer(0).SetChips(88)
	g.SetPoolForTest(40)

	p := newMockReversisPresenter()
	i := NewReversisInteractor(g, p)

	data, err := i.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreReversisInteractor(data, p)
	require.NoError(t, err)

	assert.Equal(t, 88, restored.Game.GetPlayer(0).GetChips())
	assert.Equal(t, 40, restored.Game.GetPool())
	assert.Equal(t, g.GetConfig().Rounds, restored.Game.GetConfig().Rounds)
}

func TestRestoreReversisInteractorRejectsGarbage(t *testing.T) {
	_, err := RestoreReversisInteractor([]byte("not json"), newMockReversisPresenter())
	assert.Error(t, err)
}
