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

func newMockPolignacGame() *interfaces.MockPolignacGame { return new(interfaces.MockPolignacGame) }

func newMockPolignacPresenter() *presenter.MockPolignacPresenter {
	return new(presenter.MockPolignacPresenter)
}

func TestNewPolignacInteractor(t *testing.T) {
	assert.NotNil(t, NewPolignacInteractor(newMockPolignacGame(), newMockPolignacPresenter()))
}

func TestNewPolignacInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewPolignacInteractor(nil, newMockPolignacPresenter()) })
	assert.Panics(t, func() { NewPolignacInteractor(newMockPolignacGame(), nil) })
}

// **Reset は CPU を一切動かさない。** 配り終えた時点は宣言フェーズで、
// 人間が capot を選ぶまで誰も打たない。
func TestPolignacInteractorResetDoesNotRunCpu(t *testing.T) {
	g := newMockPolignacGame()
	p := newMockPolignacPresenter()
	i := NewPolignacInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetHint").Return(nil).Maybe()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertCalled(t, "Reset")
	g.AssertNotCalled(t, "CpuPlay")
}

func TestPolignacInteractorDeclareCapot(t *testing.T) {
	g := newMockPolignacGame()
	p := newMockPolignacPresenter()
	i := NewPolignacInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("DeclareCapot").Return(nil)
	g.On("IsHumanTurn").Return(true)
	p.On("Output", g, nil).Return("declared")

	assert.Equal(t, "declared", i.DeclareCapot())
	g.AssertCalled(t, "DeclareCapot")
}

func TestPolignacInteractorDeclareCapotError(t *testing.T) {
	g := newMockPolignacGame()
	p := newMockPolignacPresenter()
	i := NewPolignacInteractor(g, p)

	err := errors.New("not the declaration phase")
	g.On("GetGameEndFlag").Return(false)
	g.On("DeclareCapot").Return(err)
	p.On("Output", g, err).Return("declare_error")

	assert.Equal(t, "declare_error", i.DeclareCapot())
	g.AssertNotCalled(t, "CpuPlay")
}

// pass のあと、人間の手番になるまで CPU が打つ。
func TestPolignacInteractorPassRunsCpuUntilHumanTurn(t *testing.T) {
	g := newMockPolignacGame()
	p := newMockPolignacPresenter()
	i := NewPolignacInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("PassDeclaration").Return(nil)
	g.On("IsHumanTurn").Return(false).Times(2)
	g.On("IsHumanTurn").Return(true)
	g.On("GetPhase").Return(domain.PolignacPhasePlay)
	g.On("CpuPlay").Return()
	p.On("Output", g, nil).Return("passed")

	assert.Equal(t, "passed", i.Pass())
	g.AssertNumberOfCalls(t, "CpuPlay", 2)
}

func TestPolignacInteractorPassError(t *testing.T) {
	g := newMockPolignacGame()
	p := newMockPolignacPresenter()
	i := NewPolignacInteractor(g, p)

	err := errors.New("not the declaration phase")
	g.On("GetGameEndFlag").Return(false)
	g.On("PassDeclaration").Return(err)
	p.On("Output", g, err).Return("pass_error")

	assert.Equal(t, "pass_error", i.Pass())
}

// **ラウンド終了で必ず止まる。** 勝手に配り直すと失点を確認できない。
func TestPolignacInteractorStopsAtRoundEnd(t *testing.T) {
	g := newMockPolignacGame()
	p := newMockPolignacPresenter()
	i := NewPolignacInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("PassDeclaration").Return(nil)
	g.On("IsHumanTurn").Return(false)
	g.On("GetPhase").Return(domain.PolignacPhaseRoundEnd)
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Pass())
	g.AssertNotCalled(t, "CpuPlay")
	g.AssertNotCalled(t, "NextRound")
}

func TestPolignacInteractorCpuLoopHasAnUpperBound(t *testing.T) {
	g := newMockPolignacGame()
	p := newMockPolignacPresenter()
	i := NewPolignacInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("PassDeclaration").Return(nil)
	g.On("IsHumanTurn").Return(false)
	g.On("GetPhase").Return(domain.PolignacPhasePlay)
	g.On("CpuPlay").Return()
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Pass())
	g.AssertNumberOfCalls(t, "CpuPlay", maxCpuTurnsPerCall)
}

func TestPolignacInteractorPlay(t *testing.T) {
	g := newMockPolignacGame()
	p := newMockPolignacPresenter()
	i := NewPolignacInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	g.On("PlayerPlay", 2).Return(nil)
	p.On("Output", g, nil).Return("played")

	assert.Equal(t, "played", i.Play(2))
	g.AssertCalled(t, "PlayerPlay", 2)
}

func TestPolignacInteractorPlayError(t *testing.T) {
	g := newMockPolignacGame()
	p := newMockPolignacPresenter()
	i := NewPolignacInteractor(g, p)

	err := errors.New("must follow suit")
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	g.On("PlayerPlay", 2).Return(err)
	p.On("Output", g, err).Return("error_output")

	assert.Equal(t, "error_output", i.Play(2))
	g.AssertNotCalled(t, "CpuPlay")
}

func TestPolignacInteractorPlayBlocked(t *testing.T) {
	for _, tc := range []struct {
		name    string
		ended   bool
		isHuman bool
	}{
		{"game over", true, true},
		{"not human turn", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockPolignacGame()
			p := newMockPolignacPresenter()
			i := NewPolignacInteractor(g, p)

			g.On("GetGameEndFlag").Return(tc.ended)
			g.On("IsHumanTurn").Return(tc.isHuman)
			p.On("Output", g, nil).Return("blocked")

			assert.Equal(t, "blocked", i.Play(0))
			g.AssertNotCalled(t, "PlayerPlay", 0)
		})
	}
}

// **NextRound も CPU を回さない。** 次のラウンドは宣言フェーズから始まる。
func TestPolignacInteractorNextRound(t *testing.T) {
	g := newMockPolignacGame()
	p := newMockPolignacPresenter()
	i := NewPolignacInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("NextRound").Return()
	p.On("Output", g, nil).Return("next")

	assert.Equal(t, "next", i.NextRound())
	g.AssertCalled(t, "NextRound")
	g.AssertNotCalled(t, "CpuPlay")
}

func TestPolignacInteractorGuardsAfterGameEnd(t *testing.T) {
	for _, tc := range []struct {
		name   string
		call   func(*PolignacInteractor) string
		method string
	}{
		{"next round", func(i *PolignacInteractor) string { return i.NextRound() }, "NextRound"},
		{"give up", func(i *PolignacInteractor) string { return i.GiveUp() }, "GiveUp"},
		{"declare capot", func(i *PolignacInteractor) string { return i.DeclareCapot() }, "DeclareCapot"},
		{"pass", func(i *PolignacInteractor) string { return i.Pass() }, "PassDeclaration"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockPolignacGame()
			p := newMockPolignacPresenter()
			i := NewPolignacInteractor(g, p)

			g.On("GetGameEndFlag").Return(true)
			p.On("Output", g, nil).Return("over")

			assert.Equal(t, "over", tc.call(i))
			g.AssertNotCalled(t, tc.method)
		})
	}
}

func TestPolignacInteractorGiveUp(t *testing.T) {
	g := newMockPolignacGame()
	p := newMockPolignacPresenter()
	i := NewPolignacInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("gave_up")

	assert.Equal(t, "gave_up", i.GiveUp())
	g.AssertCalled(t, "GiveUp")
}

func TestPolignacInteractorGetConfig(t *testing.T) {
	g := newMockPolignacGame()
	i := NewPolignacInteractor(g, newMockPolignacPresenter())
	cfg := domain.PolignacConfig{Rounds: 6}
	g.On("GetConfig").Return(cfg)
	assert.Equal(t, cfg, i.GetConfig())
}

func TestPolignacInteractorResetWithInvalidConfig(t *testing.T) {
	g := newMockPolignacGame()
	p := newMockPolignacPresenter()
	i := NewPolignacInteractor(g, p)

	p.On("Output", mock.Anything, mock.Anything).Return("cfg_error")

	assert.Equal(t, "cfg_error", i.ResetWithConfig(domain.PolignacConfig{Rounds: 0}))
	g.AssertNotCalled(t, "Reset")
	g.AssertNotCalled(t, "SetConfig", mock.Anything)
}

func TestPolignacInteractorHintAndActionLog(t *testing.T) {
	g := newMockPolignacGame()
	p := newMockPolignacPresenter()
	i := NewPolignacInteractor(g, p)

	p.On("HintOutput", g).Return("hint_output")
	p.On("ActionLogOutput", g).Return("log_output")

	assert.Equal(t, "hint_output", i.Hint())
	assert.Equal(t, "log_output", i.ActionLog())
}

// Worker はリクエストごとに KV から作り直す (#4478)。
func TestPolignacInteractorSnapshotRoundTrip(t *testing.T) {
	g := domain.NewDefaultPolignac()
	g.Reset()
	require.NoError(t, g.DeclareCapot())
	g.GetPlayer(0).SetScore(4)
	g.GetPlayer(1).SetScore(2)

	p := newMockPolignacPresenter()
	i := NewPolignacInteractor(g, p)

	data, err := i.Snapshot()
	require.NoError(t, err)

	restored, err := RestorePolignacInteractor(data, p)
	require.NoError(t, err)

	assert.Equal(t, 4, restored.Game.GetPlayer(0).GetScore())
	assert.Equal(t, 2, restored.Game.GetPlayer(1).GetScore())
	assert.Equal(t, 0, restored.Game.GetCapotIdx(), "capot 宣言が往復する")
	assert.True(t, restored.Game.GetPlayer(0).GetDeclaredCapot())
	assert.Equal(t, g.GetConfig().Rounds, restored.Game.GetConfig().Rounds)
}

func TestRestorePolignacInteractorRejectsGarbage(t *testing.T) {
	_, err := RestorePolignacInteractor([]byte("not json"), newMockPolignacPresenter())
	assert.Error(t, err)
}
