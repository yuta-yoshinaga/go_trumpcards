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

func newMockMinibridgeGame() *interfaces.MockMinibridgeGame {
	return new(interfaces.MockMinibridgeGame)
}

func newMockMinibridgePresenter() *presenter.MockMinibridgePresenter {
	return new(presenter.MockMinibridgePresenter)
}

func TestNewMinibridgeInteractor(t *testing.T) {
	assert.NotNil(t, NewMinibridgeInteractor(newMockMinibridgeGame(), newMockMinibridgePresenter()))
}

func TestNewMinibridgeInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewMinibridgeInteractor(nil, newMockMinibridgePresenter()) })
	assert.Panics(t, func() { NewMinibridgeInteractor(newMockMinibridgeGame(), nil) })
}

// **契約とプレイの 2 段を回す。** 契約で止めると人間が操作できない盤面を返す。
func TestMinibridgeInteractorResetWalksBothPhases(t *testing.T) {
	g := newMockMinibridgeGame()
	p := newMockMinibridgePresenter()
	i := NewMinibridgeInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.MinibridgePhaseContract).Once()
	g.On("IsHumanContractTurn").Return(false).Once()
	g.On("CpuSelectContract").Return()
	g.On("GetPhase").Return(domain.MinibridgePhasePlay)
	g.On("IsHumanTurn").Return(false).Twice()
	g.On("IsHumanTurn").Return(true)
	g.On("CpuPlay").Return()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertNumberOfCalls(t, "CpuSelectContract", 1)
	g.AssertNumberOfCalls(t, "CpuPlay", 2)
}

// **人間の番では止まる。** 2 段それぞれで確認する。
func TestMinibridgeInteractorStopsAtEveryHumanTurn(t *testing.T) {
	for _, tc := range []struct {
		name    string
		phase   domain.MinibridgePhase
		flag    string
		skipped []string
	}{
		{"契約", domain.MinibridgePhaseContract, "IsHumanContractTurn", []string{"CpuSelectContract", "CpuPlay"}},
		{"プレイ", domain.MinibridgePhasePlay, "IsHumanTurn", []string{"CpuPlay"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockMinibridgeGame()
			p := newMockMinibridgePresenter()
			i := NewMinibridgeInteractor(g, p)

			g.On("Reset").Return()
			g.On("GetGameEndFlag").Return(false)
			g.On("GetPhase").Return(tc.phase)
			g.On(tc.flag).Return(true)
			p.On("Output", g, nil).Return("out")

			assert.Equal(t, "out", i.Reset())
			for _, m := range tc.skipped {
				g.AssertNotCalled(t, m)
			}
		})
	}
}

// **ディール終了では止める。** 次のディールは next で明示的に。
func TestMinibridgeInteractorStopsAtRoundEnd(t *testing.T) {
	g := newMockMinibridgeGame()
	p := newMockMinibridgePresenter()
	i := NewMinibridgeInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.MinibridgePhaseRoundEnd)
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNotCalled(t, "CpuPlay")
	g.AssertNotCalled(t, "NextRound")
}

func TestMinibridgeInteractorContractSurfacesErrors(t *testing.T) {
	g := newMockMinibridgeGame()
	p := newMockMinibridgePresenter()
	i := NewMinibridgeInteractor(g, p)

	err := errors.New("invalid contract level: 9")
	g.On("GetGameEndFlag").Return(false)
	g.On("PlayerSelectContract", 9, 0).Return(err)
	p.On("Output", g, err).Return("contract_error")

	assert.Equal(t, "contract_error", i.Contract(9, 0))
	g.AssertNotCalled(t, "CpuPlay")
}

func TestMinibridgeInteractorContractAdvances(t *testing.T) {
	g := newMockMinibridgeGame()
	p := newMockMinibridgePresenter()
	i := NewMinibridgeInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("PlayerSelectContract", 3, 0).Return(nil)
	g.On("GetPhase").Return(domain.MinibridgePhasePlay)
	g.On("IsHumanTurn").Return(false).Once()
	g.On("IsHumanTurn").Return(true)
	g.On("CpuPlay").Return()
	p.On("Output", g, nil).Return("contracted")

	assert.Equal(t, "contracted", i.Contract(3, 0))
	g.AssertNumberOfCalls(t, "CpuPlay", 1)
}

func TestMinibridgeInteractorPlayRejectsAndDoesNotAdvance(t *testing.T) {
	g := newMockMinibridgeGame()
	p := newMockMinibridgePresenter()
	i := NewMinibridgeInteractor(g, p)

	playErr := errors.New("must follow the led suit")
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	g.On("PlayerPlay", 3).Return(playErr)
	p.On("Output", g, playErr).Return("play_error")

	assert.Equal(t, "play_error", i.Play(3))
	g.AssertNotCalled(t, "CpuPlay")
}

func TestMinibridgeInteractorPlayGuardsOnTurn(t *testing.T) {
	g := newMockMinibridgeGame()
	p := newMockMinibridgePresenter()
	i := NewMinibridgeInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(false)
	p.On("Output", g, nil).Return("blocked")

	assert.Equal(t, "blocked", i.Play(0))
	g.AssertNotCalled(t, "PlayerPlay")
}

func TestMinibridgeInteractorGuardsAfterGameEnd(t *testing.T) {
	for _, tc := range []struct {
		name   string
		call   func(*MinibridgeInteractor) string
		method string
	}{
		{"contract", func(i *MinibridgeInteractor) string { return i.Contract(1, 0) }, "PlayerSelectContract"},
		{"play", func(i *MinibridgeInteractor) string { return i.Play(0) }, "PlayerPlay"},
		{"next", func(i *MinibridgeInteractor) string { return i.NextRound() }, "NextRound"},
		{"giveup", func(i *MinibridgeInteractor) string { return i.GiveUp() }, "GiveUp"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockMinibridgeGame()
			p := newMockMinibridgePresenter()
			i := NewMinibridgeInteractor(g, p)

			g.On("GetGameEndFlag").Return(true)
			p.On("Output", g, nil).Return("ended")

			assert.Equal(t, "ended", tc.call(i))
			g.AssertNotCalled(t, tc.method)
		})
	}
}

func TestMinibridgeInteractorNextRoundAndGiveUp(t *testing.T) {
	g := newMockMinibridgeGame()
	p := newMockMinibridgePresenter()
	i := NewMinibridgeInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("NextRound").Return()
	g.On("GiveUp").Return()
	g.On("GetPhase").Return(domain.MinibridgePhaseContract)
	g.On("IsHumanContractTurn").Return(true)
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.NextRound())
	assert.Equal(t, "out", i.GiveUp())
	g.AssertNumberOfCalls(t, "NextRound", 1)
	g.AssertNumberOfCalls(t, "GiveUp", 1)
}

// **4 の倍数でないラウンド数は弾き、ゲームを作り直さない。**
func TestMinibridgeInteractorResetWithConfig(t *testing.T) {
	g := newMockMinibridgeGame()
	p := newMockMinibridgePresenter()
	i := NewMinibridgeInteractor(g, p)

	p.On("Output", g, mock.Anything).Return("cfg_error")
	assert.Equal(t, "cfg_error", i.ResetWithConfig(domain.MinibridgeConfig{Rounds: 6}))
	g.AssertNotCalled(t, "SetConfig")
	g.AssertNotCalled(t, "Reset")

	g2 := newMockMinibridgeGame()
	p2 := newMockMinibridgePresenter()
	i2 := NewMinibridgeInteractor(g2, p2)
	cfg := domain.MinibridgeConfig{Rounds: 8}
	g2.On("SetConfig", cfg).Return()
	g2.On("Reset").Return()
	g2.On("GetGameEndFlag").Return(false)
	g2.On("GetPhase").Return(domain.MinibridgePhaseContract)
	g2.On("IsHumanContractTurn").Return(true)
	p2.On("Output", g2, nil).Return("reset")

	assert.Equal(t, "reset", i2.ResetWithConfig(cfg))
	g2.AssertCalled(t, "SetConfig", cfg)
}

func TestMinibridgeInteractorHintAndLogDelegate(t *testing.T) {
	g := newMockMinibridgeGame()
	p := newMockMinibridgePresenter()
	i := NewMinibridgeInteractor(g, p)

	p.On("HintOutput", g).Return("hint")
	p.On("ActionLogOutput", g).Return("log")

	assert.Equal(t, "hint", i.Hint())
	assert.Equal(t, "log", i.ActionLog())
}

func TestMinibridgeInteractorGetConfigDelegates(t *testing.T) {
	g := newMockMinibridgeGame()
	p := newMockMinibridgePresenter()
	i := NewMinibridgeInteractor(g, p)

	cfg := domain.MinibridgeConfig{Rounds: 12}
	g.On("GetConfig").Return(cfg)
	assert.Equal(t, cfg, i.GetConfig())
}

func TestRestoreMinibridgeInteractor(t *testing.T) {
	src := domain.NewDefaultMinibridge()
	src.Reset()
	data, err := json.Marshal(src)
	require.NoError(t, err)

	restored, err := RestoreMinibridgeInteractor(data, new(presenter.MockMinibridgePresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
	assert.Equal(t, src.GetRoundNumber(), restored.Game.GetRoundNumber())
	assert.Equal(t, src.GetDeclarerIdx(), restored.Game.GetDeclarerIdx(), "落札者が消えない")
	for i := range domain.MinibridgePlayerCnt {
		assert.Equal(t, src.GetPlayer(i).GetHcp(), restored.Game.GetPlayer(i).GetHcp(), "HCP が消えない")
	}

	_, err = RestoreMinibridgeInteractor([]byte("{"), new(presenter.MockMinibridgePresenter))
	assert.Error(t, err)
}
