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

func newMockSlobberhannesGame() *interfaces.MockSlobberhannesGame {
	return new(interfaces.MockSlobberhannesGame)
}

func newMockSlobberhannesPresenter() *presenter.MockSlobberhannesPresenter {
	return new(presenter.MockSlobberhannesPresenter)
}

func TestNewSlobberhannesInteractor(t *testing.T) {
	assert.NotNil(t, NewSlobberhannesInteractor(newMockSlobberhannesGame(), newMockSlobberhannesPresenter()))
}

func TestNewSlobberhannesInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewSlobberhannesInteractor(nil, newMockSlobberhannesPresenter()) })
	assert.Panics(t, func() { NewSlobberhannesInteractor(newMockSlobberhannesGame(), nil) })
}

func TestSlobberhannesInteractorReset(t *testing.T) {
	g := newMockSlobberhannesGame()
	p := newMockSlobberhannesPresenter()
	i := NewSlobberhannesInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertCalled(t, "Reset")
	g.AssertNotCalled(t, "CpuPlay")
}

// 人間が最初のリードでなければ、手番が回るまで CPU が打つ。
func TestSlobberhannesInteractorRunsCpuUntilHumanTurn(t *testing.T) {
	g := newMockSlobberhannesGame()
	p := newMockSlobberhannesPresenter()
	i := NewSlobberhannesInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(false).Times(3)
	g.On("IsHumanTurn").Return(true)
	g.On("GetPhase").Return(domain.SlobberhannesPhasePlay)
	g.On("CpuPlay").Return()
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNumberOfCalls(t, "CpuPlay", 3)
}

// **ラウンド終了で必ず止まる。** 勝手に次のラウンドを配ってしまうと、
// プレイヤーが罰点の結果を見られない。
func TestSlobberhannesInteractorStopsAtRoundEnd(t *testing.T) {
	g := newMockSlobberhannesGame()
	p := newMockSlobberhannesPresenter()
	i := NewSlobberhannesInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(false)
	g.On("GetPhase").Return(domain.SlobberhannesPhaseRoundEnd)
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNotCalled(t, "CpuPlay")
	g.AssertNotCalled(t, "NextRound")
}

// 進まない CpuPlay でハングしない。
func TestSlobberhannesInteractorCpuLoopHasAnUpperBound(t *testing.T) {
	g := newMockSlobberhannesGame()
	p := newMockSlobberhannesPresenter()
	i := NewSlobberhannesInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(false)
	g.On("GetPhase").Return(domain.SlobberhannesPhasePlay)
	g.On("CpuPlay").Return()
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNumberOfCalls(t, "CpuPlay", maxCpuTurnsPerCall)
}

func TestSlobberhannesInteractorPlay(t *testing.T) {
	g := newMockSlobberhannesGame()
	p := newMockSlobberhannesPresenter()
	i := NewSlobberhannesInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	g.On("PlayerPlay", 2).Return(nil)
	p.On("Output", g, nil).Return("played")

	assert.Equal(t, "played", i.Play(2))
	g.AssertCalled(t, "PlayerPlay", 2)
}

func TestSlobberhannesInteractorPlayError(t *testing.T) {
	g := newMockSlobberhannesGame()
	p := newMockSlobberhannesPresenter()
	i := NewSlobberhannesInteractor(g, p)

	err := errors.New("must follow suit")
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	g.On("PlayerPlay", 2).Return(err)
	p.On("Output", g, err).Return("error_output")

	assert.Equal(t, "error_output", i.Play(2))
	g.AssertNotCalled(t, "CpuPlay")
}

func TestSlobberhannesInteractorPlayBlocked(t *testing.T) {
	for _, tc := range []struct {
		name    string
		ended   bool
		isHuman bool
	}{
		{"game over", true, true},
		{"not human turn", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockSlobberhannesGame()
			p := newMockSlobberhannesPresenter()
			i := NewSlobberhannesInteractor(g, p)

			g.On("GetGameEndFlag").Return(tc.ended)
			g.On("IsHumanTurn").Return(tc.isHuman)
			p.On("Output", g, nil).Return("blocked")

			assert.Equal(t, "blocked", i.Play(0))
			g.AssertNotCalled(t, "PlayerPlay", 0)
		})
	}
}

func TestSlobberhannesInteractorNextRound(t *testing.T) {
	g := newMockSlobberhannesGame()
	p := newMockSlobberhannesPresenter()
	i := NewSlobberhannesInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("NextRound").Return()
	g.On("IsHumanTurn").Return(true)
	p.On("Output", g, nil).Return("next")

	assert.Equal(t, "next", i.NextRound())
	g.AssertCalled(t, "NextRound")
}

func TestSlobberhannesInteractorNextRoundAfterGameEnd(t *testing.T) {
	g := newMockSlobberhannesGame()
	p := newMockSlobberhannesPresenter()
	i := NewSlobberhannesInteractor(g, p)

	g.On("GetGameEndFlag").Return(true)
	p.On("Output", g, nil).Return("over")

	assert.Equal(t, "over", i.NextRound())
	g.AssertNotCalled(t, "NextRound")
}

func TestSlobberhannesInteractorGiveUp(t *testing.T) {
	g := newMockSlobberhannesGame()
	p := newMockSlobberhannesPresenter()
	i := NewSlobberhannesInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("gave_up")

	assert.Equal(t, "gave_up", i.GiveUp())
	g.AssertCalled(t, "GiveUp")
}

func TestSlobberhannesInteractorGiveUpAfterGameEnd(t *testing.T) {
	g := newMockSlobberhannesGame()
	p := newMockSlobberhannesPresenter()
	i := NewSlobberhannesInteractor(g, p)

	g.On("GetGameEndFlag").Return(true)
	p.On("Output", g, nil).Return("already_over")

	assert.Equal(t, "already_over", i.GiveUp())
	g.AssertNotCalled(t, "GiveUp")
}

func TestSlobberhannesInteractorGetConfig(t *testing.T) {
	g := newMockSlobberhannesGame()
	p := newMockSlobberhannesPresenter()
	i := NewSlobberhannesInteractor(g, p)

	cfg := domain.SlobberhannesConfig{Rounds: 6}
	g.On("GetConfig").Return(cfg)
	assert.Equal(t, cfg, i.GetConfig())
}

// 範囲外のラウンド数は弾かれ、ゲームは配り直されない。
func TestSlobberhannesInteractorResetWithInvalidConfig(t *testing.T) {
	g := newMockSlobberhannesGame()
	p := newMockSlobberhannesPresenter()
	i := NewSlobberhannesInteractor(g, p)

	p.On("Output", mock.Anything, mock.Anything).Return("cfg_error")

	assert.Equal(t, "cfg_error", i.ResetWithConfig(domain.SlobberhannesConfig{Rounds: 0}))
	g.AssertNotCalled(t, "Reset")
	g.AssertNotCalled(t, "SetConfig", mock.Anything)
}

func TestSlobberhannesInteractorHintAndActionLog(t *testing.T) {
	g := newMockSlobberhannesGame()
	p := newMockSlobberhannesPresenter()
	i := NewSlobberhannesInteractor(g, p)

	p.On("HintOutput", g).Return("hint_output")
	p.On("ActionLogOutput", g).Return("log_output")

	assert.Equal(t, "hint_output", i.Hint())
	assert.Equal(t, "log_output", i.ActionLog())
}

// Worker はリクエストごとに KV から作り直す。**得点と罰の内訳が往復しないと
// ラウンド途中で罰が消える** (#4478)。
func TestSlobberhannesInteractorSnapshotRoundTrip(t *testing.T) {
	g := domain.NewDefaultSlobberhannes()
	g.Reset()
	g.GetPlayer(0).SetScore(-2)
	g.GetPlayer(1).SetScore(1)

	p := newMockSlobberhannesPresenter()
	i := NewSlobberhannesInteractor(g, p)

	data, err := i.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreSlobberhannesInteractor(data, p)
	require.NoError(t, err)

	assert.Equal(t, -2, restored.Game.GetPlayer(0).GetScore())
	assert.Equal(t, 1, restored.Game.GetPlayer(1).GetScore())
	assert.Equal(t, g.GetRoundNumber(), restored.Game.GetRoundNumber())
	assert.Equal(t, g.GetConfig().Rounds, restored.Game.GetConfig().Rounds)
}

func TestRestoreSlobberhannesInteractorRejectsGarbage(t *testing.T) {
	_, err := RestoreSlobberhannesInteractor([]byte("not json"), newMockSlobberhannesPresenter())
	assert.Error(t, err)
}
