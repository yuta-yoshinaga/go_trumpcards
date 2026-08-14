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

func newMockHasenpfefferGame() *interfaces.MockHasenpfefferGame {
	return new(interfaces.MockHasenpfefferGame)
}

func newMockHasenpfefferPresenter() *presenter.MockHasenpfefferPresenter {
	return new(presenter.MockHasenpfefferPresenter)
}

func TestNewHasenpfefferInteractor(t *testing.T) {
	assert.NotNil(t, NewHasenpfefferInteractor(newMockHasenpfefferGame(), newMockHasenpfefferPresenter()))
}

func TestNewHasenpfefferInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewHasenpfefferInteractor(nil, newMockHasenpfefferPresenter()) })
	assert.Panics(t, func() { NewHasenpfefferInteractor(newMockHasenpfefferGame(), nil) })
}

// **競り・捨て札・プレイの 3 段すべてを回す。** どれか 1 つで止めると、人間が
// 操作できない盤面を返してしまう。
func TestHasenpfefferInteractorResetWalksAllThreePhases(t *testing.T) {
	g := newMockHasenpfefferGame()
	p := newMockHasenpfefferPresenter()
	i := NewHasenpfefferInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.HasenpfefferPhaseBid).Times(2)
	g.On("IsHumanBidTurn").Return(false).Twice()
	g.On("CpuBid").Return()
	g.On("GetPhase").Return(domain.HasenpfefferPhaseDiscard).Once()
	g.On("IsHumanDiscardTurn").Return(false).Once()
	g.On("CpuDiscard").Return()
	g.On("GetPhase").Return(domain.HasenpfefferPhasePlay)
	g.On("IsHumanTurn").Return(false).Once()
	g.On("IsHumanTurn").Return(true)
	g.On("CpuPlay").Return()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertNumberOfCalls(t, "CpuBid", 2)
	g.AssertNumberOfCalls(t, "CpuDiscard", 1)
	g.AssertNumberOfCalls(t, "CpuPlay", 1)
}

// **人間の番では止まる。** 3 段それぞれで確認する。
func TestHasenpfefferInteractorStopsAtEveryHumanTurn(t *testing.T) {
	for _, tc := range []struct {
		name    string
		phase   domain.HasenpfefferPhase
		flag    string
		skipped []string
	}{
		{"競り", domain.HasenpfefferPhaseBid, "IsHumanBidTurn", []string{"CpuBid", "CpuDiscard", "CpuPlay"}},
		{"捨て札", domain.HasenpfefferPhaseDiscard, "IsHumanDiscardTurn", []string{"CpuDiscard", "CpuPlay"}},
		{"プレイ", domain.HasenpfefferPhasePlay, "IsHumanTurn", []string{"CpuPlay"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockHasenpfefferGame()
			p := newMockHasenpfefferPresenter()
			i := NewHasenpfefferInteractor(g, p)

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

// **ハンド終了では止める。** 次のハンドは next で明示的に始める。
func TestHasenpfefferInteractorStopsAtHandEnd(t *testing.T) {
	g := newMockHasenpfefferGame()
	p := newMockHasenpfefferPresenter()
	i := NewHasenpfefferInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.HasenpfefferPhaseHandEnd)
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNotCalled(t, "CpuPlay")
	g.AssertNotCalled(t, "NextHand")
}

// **範囲外の宣言はドメインが弾き、その理由をそのまま返す。**
func TestHasenpfefferInteractorBidSurfacesErrors(t *testing.T) {
	g := newMockHasenpfefferGame()
	p := newMockHasenpfefferPresenter()
	i := NewHasenpfefferInteractor(g, p)

	bidErr := errors.New("bid must be 4..6")
	g.On("GetGameEndFlag").Return(false)
	g.On("PlayerBid", 2).Return(bidErr)
	p.On("Output", g, bidErr).Return("bid_error")

	assert.Equal(t, "bid_error", i.Bid(2))
	g.AssertNotCalled(t, "CpuBid")
}

// **0（降りる）もドメインに届く。** 未指定と取り違えない。
func TestHasenpfefferInteractorPassReachesTheDomain(t *testing.T) {
	g := newMockHasenpfefferGame()
	p := newMockHasenpfefferPresenter()
	i := NewHasenpfefferInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("PlayerBid", 0).Return(nil)
	g.On("GetPhase").Return(domain.HasenpfefferPhaseBid)
	g.On("IsHumanBidTurn").Return(true)
	p.On("Output", g, nil).Return("passed")

	assert.Equal(t, "passed", i.Bid(0))
	g.AssertCalled(t, "PlayerBid", 0)
}

func TestHasenpfefferInteractorDiscardSurfacesErrors(t *testing.T) {
	g := newMockHasenpfefferGame()
	p := newMockHasenpfefferPresenter()
	i := NewHasenpfefferInteractor(g, p)

	discardErr := errors.New("invalid trump suit: 9")
	g.On("GetGameEndFlag").Return(false)
	g.On("PlayerDiscard", 0, 9).Return(discardErr)
	p.On("Output", g, discardErr).Return("discard_error")

	assert.Equal(t, "discard_error", i.Discard(0, 9))
	g.AssertNotCalled(t, "CpuPlay")
}

func TestHasenpfefferInteractorPlayRejectsAndDoesNotAdvance(t *testing.T) {
	g := newMockHasenpfefferGame()
	p := newMockHasenpfefferPresenter()
	i := NewHasenpfefferInteractor(g, p)

	playErr := errors.New("must follow the led suit")
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.HasenpfefferPhasePlay)
	g.On("IsHumanTurn").Return(true)
	g.On("PlayerPlay", 3).Return(playErr)
	p.On("Output", g, playErr).Return("play_error")

	assert.Equal(t, "play_error", i.Play(3))
	g.AssertNotCalled(t, "CpuPlay")
}

func TestHasenpfefferInteractorPlayGuardsOnTurn(t *testing.T) {
	g := newMockHasenpfefferGame()
	p := newMockHasenpfefferPresenter()
	i := NewHasenpfefferInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.HasenpfefferPhaseBid)
	g.On("IsHumanTurn").Return(false).Maybe()
	p.On("Output", g, nil).Return("blocked")

	assert.Equal(t, "blocked", i.Play(0))
	g.AssertNotCalled(t, "PlayerPlay")
}

func TestHasenpfefferInteractorGuardsAfterGameEnd(t *testing.T) {
	for _, tc := range []struct {
		name   string
		call   func(*HasenpfefferInteractor) string
		method string
	}{
		{"bid", func(i *HasenpfefferInteractor) string { return i.Bid(3) }, "PlayerBid"},
		{"discard", func(i *HasenpfefferInteractor) string { return i.Discard(0, 1) }, "PlayerDiscard"},
		{"next", func(i *HasenpfefferInteractor) string { return i.NextHand() }, "NextHand"},
		{"giveup", func(i *HasenpfefferInteractor) string { return i.GiveUp() }, "GiveUp"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockHasenpfefferGame()
			p := newMockHasenpfefferPresenter()
			i := NewHasenpfefferInteractor(g, p)

			g.On("GetGameEndFlag").Return(true)
			p.On("Output", g, nil).Return("ended")

			assert.Equal(t, "ended", tc.call(i))
			g.AssertNotCalled(t, tc.method, mock.Anything)
			g.AssertNotCalled(t, tc.method)
		})
	}
}

func TestHasenpfefferInteractorNextHandAdvances(t *testing.T) {
	g := newMockHasenpfefferGame()
	p := newMockHasenpfefferPresenter()
	i := NewHasenpfefferInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("NextHand").Return()
	g.On("GetPhase").Return(domain.HasenpfefferPhaseBid)
	g.On("IsHumanBidTurn").Return(true)
	p.On("Output", g, nil).Return("next")

	assert.Equal(t, "next", i.NextHand())
	g.AssertCalled(t, "NextHand")
}

func TestHasenpfefferInteractorResetWithConfig(t *testing.T) {
	g := newMockHasenpfefferGame()
	p := newMockHasenpfefferPresenter()
	i := NewHasenpfefferInteractor(g, p)

	cfg := domain.HasenpfefferConfig{Target: 15}
	g.On("SetConfig", cfg).Return()
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.HasenpfefferPhaseBid)
	g.On("IsHumanBidTurn").Return(true)
	p.On("Output", g, nil).Return("configured")

	assert.Equal(t, "configured", i.ResetWithConfig(cfg))
	g.AssertCalled(t, "SetConfig", cfg)
}

func TestHasenpfefferInteractorResetWithInvalidConfig(t *testing.T) {
	for _, n := range []int{domain.HasenpfefferTargetMin - 1, domain.HasenpfefferTargetMax + 1} {
		g := newMockHasenpfefferGame()
		p := newMockHasenpfefferPresenter()
		i := NewHasenpfefferInteractor(g, p)

		p.On("Output", mock.Anything, mock.Anything).Return("cfg_error")

		assert.Equal(t, "cfg_error", i.ResetWithConfig(domain.HasenpfefferConfig{Target: n}))
		g.AssertNotCalled(t, "Reset")
		g.AssertNotCalled(t, "SetConfig", mock.Anything)
	}
}

func TestHasenpfefferInteractorHintAndLogDelegate(t *testing.T) {
	g := newMockHasenpfefferGame()
	p := newMockHasenpfefferPresenter()
	i := NewHasenpfefferInteractor(g, p)

	p.On("HintOutput", g).Return("hint")
	p.On("ActionLogOutput", g).Return("log")

	assert.Equal(t, "hint", i.Hint())
	assert.Equal(t, "log", i.ActionLog())
}

func TestHasenpfefferInteractorGetConfigDelegates(t *testing.T) {
	g := newMockHasenpfefferGame()
	p := newMockHasenpfefferPresenter()
	i := NewHasenpfefferInteractor(g, p)

	cfg := domain.HasenpfefferConfig{Target: 12}
	g.On("GetConfig").Return(cfg)
	assert.Equal(t, cfg, i.GetConfig())
}

func TestRestoreHasenpfefferInteractor(t *testing.T) {
	src := domain.NewDefaultHasenpfeffer()
	src.Reset()
	data, err := json.Marshal(src)
	require.NoError(t, err)

	restored, err := RestoreHasenpfefferInteractor(data, new(presenter.MockHasenpfefferPresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
	assert.Equal(t, src.GetHandNumber(), restored.Game.GetHandNumber())
	assert.Equal(t, src.GetBlindSize(), restored.Game.GetBlindSize())
	assert.Equal(t, src.GetPlayer(0).GetCardsSize(), restored.Game.GetPlayer(0).GetCardsSize())

	_, err = RestoreHasenpfefferInteractor([]byte("{"), new(presenter.MockHasenpfefferPresenter))
	assert.Error(t, err)
}
