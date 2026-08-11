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

func newMockSergeantMajorGame() *interfaces.MockSergeantMajorGame {
	return new(interfaces.MockSergeantMajorGame)
}

func newMockSergeantMajorPresenter() *presenter.MockSergeantMajorPresenter {
	return new(presenter.MockSergeantMajorPresenter)
}

func TestNewSergeantMajorInteractor(t *testing.T) {
	assert.NotNil(t, NewSergeantMajorInteractor(newMockSergeantMajorGame(), newMockSergeantMajorPresenter()))
}

func TestNewSergeantMajorInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewSergeantMajorInteractor(nil, newMockSergeantMajorPresenter()) })
	assert.Panics(t, func() { NewSergeantMajorInteractor(newMockSergeantMajorGame(), nil) })
}

// **宣言・捨て札・プレイの 3 段すべてを回す。** どれかで止めると人間が操作
// できない盤面を返してしまう。
func TestSergeantMajorInteractorResetWalksAllThreePhases(t *testing.T) {
	g := newMockSergeantMajorGame()
	p := newMockSergeantMajorPresenter()
	i := NewSergeantMajorInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.SergeantMajorPhaseTrump).Once()
	g.On("IsHumanTrumpTurn").Return(false).Once()
	g.On("CpuDeclareTrump").Return()
	g.On("GetPhase").Return(domain.SergeantMajorPhaseDiscard).Once()
	g.On("IsHumanDiscardTurn").Return(false).Once()
	g.On("CpuDiscard").Return()
	g.On("GetPhase").Return(domain.SergeantMajorPhasePlay)
	g.On("IsHumanTurn").Return(false).Twice()
	g.On("IsHumanTurn").Return(true)
	g.On("CpuPlay").Return()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertNumberOfCalls(t, "CpuDeclareTrump", 1)
	g.AssertNumberOfCalls(t, "CpuDiscard", 1)
	g.AssertNumberOfCalls(t, "CpuPlay", 2)
}

// **人間の番では止まる。** 3 段それぞれで確認する。
func TestSergeantMajorInteractorStopsAtEveryHumanTurn(t *testing.T) {
	for _, tc := range []struct {
		name    string
		phase   domain.SergeantMajorPhase
		flag    string
		skipped []string
	}{
		{"宣言", domain.SergeantMajorPhaseTrump, "IsHumanTrumpTurn", []string{"CpuDeclareTrump", "CpuDiscard", "CpuPlay"}},
		{"捨て札", domain.SergeantMajorPhaseDiscard, "IsHumanDiscardTurn", []string{"CpuDiscard", "CpuPlay"}},
		{"プレイ", domain.SergeantMajorPhasePlay, "IsHumanTurn", []string{"CpuPlay"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockSergeantMajorGame()
			p := newMockSergeantMajorPresenter()
			i := NewSergeantMajorInteractor(g, p)

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

// **ラウンド終了では止める。** 次のラウンドは next で明示的に。
func TestSergeantMajorInteractorStopsAtRoundEnd(t *testing.T) {
	g := newMockSergeantMajorGame()
	p := newMockSergeantMajorPresenter()
	i := NewSergeantMajorInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.SergeantMajorPhaseRoundEnd)
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNotCalled(t, "CpuPlay")
	g.AssertNotCalled(t, "NextRound")
}

func TestSergeantMajorInteractorDeclareTrumpSurfacesErrors(t *testing.T) {
	g := newMockSergeantMajorGame()
	p := newMockSergeantMajorPresenter()
	i := NewSergeantMajorInteractor(g, p)

	declErr := errors.New("invalid suit: 9")
	g.On("GetGameEndFlag").Return(false)
	g.On("PlayerDeclareTrump", 9).Return(declErr)
	p.On("Output", g, declErr).Return("trump_error")

	assert.Equal(t, "trump_error", i.DeclareTrump(9))
	g.AssertNotCalled(t, "CpuDiscard")
}

// **捨て札の枚数はドメインが弾き、その理由をそのまま返す。**
func TestSergeantMajorInteractorDiscardSurfacesErrors(t *testing.T) {
	g := newMockSergeantMajorGame()
	p := newMockSergeantMajorPresenter()
	i := NewSergeantMajorInteractor(g, p)

	discardErr := errors.New("must discard exactly 4 cards")
	bad := []int{0, 1}
	g.On("GetGameEndFlag").Return(false)
	g.On("PlayerDiscard", bad).Return(discardErr)
	p.On("Output", g, discardErr).Return("discard_error")

	assert.Equal(t, "discard_error", i.Discard(bad))
	g.AssertNotCalled(t, "CpuPlay")
}

func TestSergeantMajorInteractorDiscardAdvances(t *testing.T) {
	g := newMockSergeantMajorGame()
	p := newMockSergeantMajorPresenter()
	i := NewSergeantMajorInteractor(g, p)

	indices := []int{0, 1, 2, 3}
	g.On("GetGameEndFlag").Return(false)
	g.On("PlayerDiscard", indices).Return(nil)
	g.On("GetPhase").Return(domain.SergeantMajorPhasePlay)
	g.On("IsHumanTurn").Return(false).Once()
	g.On("IsHumanTurn").Return(true)
	g.On("CpuPlay").Return()
	p.On("Output", g, nil).Return("discarded")

	assert.Equal(t, "discarded", i.Discard(indices))
	g.AssertNumberOfCalls(t, "CpuPlay", 1)
}

func TestSergeantMajorInteractorPlayRejectsAndDoesNotAdvance(t *testing.T) {
	g := newMockSergeantMajorGame()
	p := newMockSergeantMajorPresenter()
	i := NewSergeantMajorInteractor(g, p)

	playErr := errors.New("must follow the led suit")
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.SergeantMajorPhasePlay)
	g.On("IsHumanTurn").Return(true)
	g.On("PlayerPlay", 3).Return(playErr)
	p.On("Output", g, playErr).Return("play_error")

	assert.Equal(t, "play_error", i.Play(3))
	g.AssertNotCalled(t, "CpuPlay")
}

func TestSergeantMajorInteractorPlayGuardsOnTurn(t *testing.T) {
	g := newMockSergeantMajorGame()
	p := newMockSergeantMajorPresenter()
	i := NewSergeantMajorInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.SergeantMajorPhaseTrump)
	g.On("IsHumanTurn").Return(false).Maybe()
	p.On("Output", g, nil).Return("blocked")

	assert.Equal(t, "blocked", i.Play(0))
	g.AssertNotCalled(t, "PlayerPlay")
}

func TestSergeantMajorInteractorGuardsAfterGameEnd(t *testing.T) {
	for _, tc := range []struct {
		name   string
		call   func(*SergeantMajorInteractor) string
		method string
	}{
		{"trump", func(i *SergeantMajorInteractor) string { return i.DeclareTrump(1) }, "PlayerDeclareTrump"},
		{"discard", func(i *SergeantMajorInteractor) string { return i.Discard([]int{0, 1, 2, 3}) }, "PlayerDiscard"},
		{"next", func(i *SergeantMajorInteractor) string { return i.NextRound() }, "NextRound"},
		{"giveup", func(i *SergeantMajorInteractor) string { return i.GiveUp() }, "GiveUp"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockSergeantMajorGame()
			p := newMockSergeantMajorPresenter()
			i := NewSergeantMajorInteractor(g, p)

			g.On("GetGameEndFlag").Return(true)
			p.On("Output", g, nil).Return("ended")

			assert.Equal(t, "ended", tc.call(i))
			g.AssertNotCalled(t, tc.method, mock.Anything)
			g.AssertNotCalled(t, tc.method)
		})
	}
}

func TestSergeantMajorInteractorNextRoundAdvances(t *testing.T) {
	g := newMockSergeantMajorGame()
	p := newMockSergeantMajorPresenter()
	i := NewSergeantMajorInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("NextRound").Return()
	g.On("GetPhase").Return(domain.SergeantMajorPhaseTrump)
	g.On("IsHumanTrumpTurn").Return(true)
	p.On("Output", g, nil).Return("next")

	assert.Equal(t, "next", i.NextRound())
	g.AssertCalled(t, "NextRound")
}

func TestSergeantMajorInteractorResetWithConfig(t *testing.T) {
	g := newMockSergeantMajorGame()
	p := newMockSergeantMajorPresenter()
	i := NewSergeantMajorInteractor(g, p)

	cfg := domain.SergeantMajorConfig{Rounds: 6}
	g.On("SetConfig", cfg).Return()
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.SergeantMajorPhaseTrump)
	g.On("IsHumanTrumpTurn").Return(true)
	p.On("Output", g, nil).Return("configured")

	assert.Equal(t, "configured", i.ResetWithConfig(cfg))
	g.AssertCalled(t, "SetConfig", cfg)
}

func TestSergeantMajorInteractorResetWithInvalidConfig(t *testing.T) {
	for _, n := range []int{domain.SergeantMajorRoundsMin - 1, domain.SergeantMajorRoundsMax + 1} {
		g := newMockSergeantMajorGame()
		p := newMockSergeantMajorPresenter()
		i := NewSergeantMajorInteractor(g, p)

		p.On("Output", mock.Anything, mock.Anything).Return("cfg_error")

		assert.Equal(t, "cfg_error", i.ResetWithConfig(domain.SergeantMajorConfig{Rounds: n}))
		g.AssertNotCalled(t, "Reset")
		g.AssertNotCalled(t, "SetConfig", mock.Anything)
	}
}

func TestSergeantMajorInteractorHintAndLogDelegate(t *testing.T) {
	g := newMockSergeantMajorGame()
	p := newMockSergeantMajorPresenter()
	i := NewSergeantMajorInteractor(g, p)

	p.On("HintOutput", g).Return("hint")
	p.On("ActionLogOutput", g).Return("log")

	assert.Equal(t, "hint", i.Hint())
	assert.Equal(t, "log", i.ActionLog())
}

func TestSergeantMajorInteractorGetConfigDelegates(t *testing.T) {
	g := newMockSergeantMajorGame()
	p := newMockSergeantMajorPresenter()
	i := NewSergeantMajorInteractor(g, p)

	cfg := domain.SergeantMajorConfig{Rounds: 9}
	g.On("GetConfig").Return(cfg)
	assert.Equal(t, cfg, i.GetConfig())
}

func TestRestoreSergeantMajorInteractor(t *testing.T) {
	src := domain.NewDefaultSergeantMajor()
	src.Reset()
	data, err := json.Marshal(src)
	require.NoError(t, err)

	restored, err := RestoreSergeantMajorInteractor(data, new(presenter.MockSergeantMajorPresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
	assert.Equal(t, src.GetRoundNumber(), restored.Game.GetRoundNumber())
	assert.Equal(t, src.GetKittySize(), restored.Game.GetKittySize())
	for i := range domain.SergeantMajorPlayerCnt {
		assert.Equal(t, src.GetPlayer(i).GetTarget(), restored.Game.GetPlayer(i).GetTarget(), "ノルマが消えない")
	}

	_, err = RestoreSergeantMajorInteractor([]byte("{"), new(presenter.MockSergeantMajorPresenter))
	assert.Error(t, err)
}
