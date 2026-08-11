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

func newMockHoneymoonBridgeGame() *interfaces.MockHoneymoonBridgeGame {
	return new(interfaces.MockHoneymoonBridgeGame)
}

func newMockHoneymoonBridgePresenter() *presenter.MockHoneymoonBridgePresenter {
	return new(presenter.MockHoneymoonBridgePresenter)
}

func TestNewHoneymoonBridgeInteractor(t *testing.T) {
	assert.NotNil(t, NewHoneymoonBridgeInteractor(newMockHoneymoonBridgeGame(), newMockHoneymoonBridgePresenter()))
}

func TestNewHoneymoonBridgeInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewHoneymoonBridgeInteractor(nil, newMockHoneymoonBridgePresenter()) })
	assert.Panics(t, func() { NewHoneymoonBridgeInteractor(newMockHoneymoonBridgeGame(), nil) })
}

// **引き合い・競り・本番の 3 段すべてを回す。** 競りを落とすと CPU の手番の
// 盤面を返してしまう。
func TestHoneymoonBridgeInteractorResetWalksAllThreePhases(t *testing.T) {
	g := newMockHoneymoonBridgeGame()
	p := newMockHoneymoonBridgePresenter()
	i := NewHoneymoonBridgeInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.HoneymoonBridgePhaseDraw).Once()
	g.On("IsHumanTurn").Return(false).Once()
	g.On("GetPhase").Return(domain.HoneymoonBridgePhaseBid).Once()
	g.On("IsHumanBidTurn").Return(false).Once()
	g.On("CpuBid").Return()
	g.On("GetPhase").Return(domain.HoneymoonBridgePhasePlay)
	g.On("IsHumanTurn").Return(false).Once()
	g.On("IsHumanTurn").Return(true)
	g.On("CpuPlay").Return()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertNumberOfCalls(t, "CpuBid", 1)
	// **引き合いと本番の両方で CPU を進める。**
	g.AssertNumberOfCalls(t, "CpuPlay", 2)
}

// **人間の番では止まる。** 3 段それぞれで確認する。
func TestHoneymoonBridgeInteractorStopsAtEveryHumanTurn(t *testing.T) {
	for _, tc := range []struct {
		name    string
		phase   domain.HoneymoonBridgePhase
		flag    string
		skipped []string
	}{
		{"引き合い", domain.HoneymoonBridgePhaseDraw, "IsHumanTurn", []string{"CpuPlay", "CpuBid"}},
		{"競り", domain.HoneymoonBridgePhaseBid, "IsHumanBidTurn", []string{"CpuBid", "CpuPlay"}},
		{"本番", domain.HoneymoonBridgePhasePlay, "IsHumanTurn", []string{"CpuPlay"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockHoneymoonBridgeGame()
			p := newMockHoneymoonBridgePresenter()
			i := NewHoneymoonBridgeInteractor(g, p)

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
func TestHoneymoonBridgeInteractorStopsAtRoundEnd(t *testing.T) {
	g := newMockHoneymoonBridgeGame()
	p := newMockHoneymoonBridgePresenter()
	i := NewHoneymoonBridgeInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.HoneymoonBridgePhaseRoundEnd)
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNotCalled(t, "CpuPlay")
	g.AssertNotCalled(t, "NextRound")
}

func TestHoneymoonBridgeInteractorBidSurfacesErrors(t *testing.T) {
	g := newMockHoneymoonBridgeGame()
	p := newMockHoneymoonBridgePresenter()
	i := NewHoneymoonBridgeInteractor(g, p)

	bidErr := errors.New("bid does not outbid the current contract")
	g.On("GetGameEndFlag").Return(false)
	g.On("PlayerBid", 2, 1).Return(bidErr)
	p.On("Output", g, bidErr).Return("bid_error")

	assert.Equal(t, "bid_error", i.Bid(2, 1))
	g.AssertNotCalled(t, "CpuBid")
}

func TestHoneymoonBridgeInteractorBidAndPassAdvance(t *testing.T) {
	for _, tc := range []struct {
		name   string
		call   func(*HoneymoonBridgeInteractor) string
		method string
	}{
		{"bid", func(i *HoneymoonBridgeInteractor) string { return i.Bid(3, 0) }, "PlayerBid"},
		{"pass", func(i *HoneymoonBridgeInteractor) string { return i.Pass() }, "PlayerPass"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockHoneymoonBridgeGame()
			p := newMockHoneymoonBridgePresenter()
			i := NewHoneymoonBridgeInteractor(g, p)

			g.On("GetGameEndFlag").Return(false)
			g.On("PlayerBid", 3, 0).Return(nil)
			g.On("PlayerPass").Return(nil)
			g.On("GetPhase").Return(domain.HoneymoonBridgePhasePlay)
			g.On("IsHumanTurn").Return(false).Once()
			g.On("IsHumanTurn").Return(true)
			g.On("CpuPlay").Return()
			p.On("Output", g, nil).Return("advanced")

			assert.Equal(t, "advanced", tc.call(i))
			g.AssertNumberOfCalls(t, "CpuPlay", 1)
			g.AssertNumberOfCalls(t, tc.method, 1)
		})
	}
}

func TestHoneymoonBridgeInteractorPassSurfacesErrors(t *testing.T) {
	g := newMockHoneymoonBridgeGame()
	p := newMockHoneymoonBridgePresenter()
	i := NewHoneymoonBridgeInteractor(g, p)

	passErr := errors.New("not your bidding turn")
	g.On("GetGameEndFlag").Return(false)
	g.On("PlayerPass").Return(passErr)
	p.On("Output", g, passErr).Return("pass_error")

	assert.Equal(t, "pass_error", i.Pass())
	g.AssertNotCalled(t, "CpuBid")
}

func TestHoneymoonBridgeInteractorPlayRejectsAndDoesNotAdvance(t *testing.T) {
	g := newMockHoneymoonBridgeGame()
	p := newMockHoneymoonBridgePresenter()
	i := NewHoneymoonBridgeInteractor(g, p)

	playErr := errors.New("must follow the led suit")
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	g.On("PlayerPlay", 3).Return(playErr)
	p.On("Output", g, playErr).Return("play_error")

	assert.Equal(t, "play_error", i.Play(3))
	g.AssertNotCalled(t, "CpuPlay")
}

func TestHoneymoonBridgeInteractorPlayGuardsOnTurn(t *testing.T) {
	g := newMockHoneymoonBridgeGame()
	p := newMockHoneymoonBridgePresenter()
	i := NewHoneymoonBridgeInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(false)
	p.On("Output", g, nil).Return("blocked")

	assert.Equal(t, "blocked", i.Play(0))
	g.AssertNotCalled(t, "PlayerPlay")
}

func TestHoneymoonBridgeInteractorGuardsAfterGameEnd(t *testing.T) {
	for _, tc := range []struct {
		name   string
		call   func(*HoneymoonBridgeInteractor) string
		method string
	}{
		{"bid", func(i *HoneymoonBridgeInteractor) string { return i.Bid(1, 0) }, "PlayerBid"},
		{"pass", func(i *HoneymoonBridgeInteractor) string { return i.Pass() }, "PlayerPass"},
		{"play", func(i *HoneymoonBridgeInteractor) string { return i.Play(0) }, "PlayerPlay"},
		{"next", func(i *HoneymoonBridgeInteractor) string { return i.NextRound() }, "NextRound"},
		{"giveup", func(i *HoneymoonBridgeInteractor) string { return i.GiveUp() }, "GiveUp"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockHoneymoonBridgeGame()
			p := newMockHoneymoonBridgePresenter()
			i := NewHoneymoonBridgeInteractor(g, p)

			g.On("GetGameEndFlag").Return(true)
			p.On("Output", g, nil).Return("ended")

			assert.Equal(t, "ended", tc.call(i))
			g.AssertNotCalled(t, tc.method)
		})
	}
}

func TestHoneymoonBridgeInteractorNextRoundAndGiveUp(t *testing.T) {
	g := newMockHoneymoonBridgeGame()
	p := newMockHoneymoonBridgePresenter()
	i := NewHoneymoonBridgeInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("NextRound").Return()
	g.On("GiveUp").Return()
	g.On("GetPhase").Return(domain.HoneymoonBridgePhaseDraw)
	g.On("IsHumanTurn").Return(true)
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.NextRound())
	assert.Equal(t, "out", i.GiveUp())
	g.AssertNumberOfCalls(t, "NextRound", 1)
	g.AssertNumberOfCalls(t, "GiveUp", 1)
}

// **不正な設定は弾き、ゲームを作り直さない。**
func TestHoneymoonBridgeInteractorResetWithConfig(t *testing.T) {
	g := newMockHoneymoonBridgeGame()
	p := newMockHoneymoonBridgePresenter()
	i := NewHoneymoonBridgeInteractor(g, p)

	p.On("Output", g, mock.Anything).Return("cfg_error")
	assert.Equal(t, "cfg_error", i.ResetWithConfig(domain.HoneymoonBridgeConfig{Target: 0}))
	g.AssertNotCalled(t, "SetConfig")
	g.AssertNotCalled(t, "Reset")

	g2 := newMockHoneymoonBridgeGame()
	p2 := newMockHoneymoonBridgePresenter()
	i2 := NewHoneymoonBridgeInteractor(g2, p2)
	cfg := domain.HoneymoonBridgeConfig{Target: 200}
	g2.On("SetConfig", cfg).Return()
	g2.On("Reset").Return()
	g2.On("GetGameEndFlag").Return(false)
	g2.On("GetPhase").Return(domain.HoneymoonBridgePhaseDraw)
	g2.On("IsHumanTurn").Return(true)
	p2.On("Output", g2, nil).Return("reset")

	assert.Equal(t, "reset", i2.ResetWithConfig(cfg))
	g2.AssertCalled(t, "SetConfig", cfg)
}

func TestHoneymoonBridgeInteractorHintAndLogDelegate(t *testing.T) {
	g := newMockHoneymoonBridgeGame()
	p := newMockHoneymoonBridgePresenter()
	i := NewHoneymoonBridgeInteractor(g, p)

	p.On("HintOutput", g).Return("hint")
	p.On("ActionLogOutput", g).Return("log")

	assert.Equal(t, "hint", i.Hint())
	assert.Equal(t, "log", i.ActionLog())
}

func TestHoneymoonBridgeInteractorGetConfigDelegates(t *testing.T) {
	g := newMockHoneymoonBridgeGame()
	p := newMockHoneymoonBridgePresenter()
	i := NewHoneymoonBridgeInteractor(g, p)

	cfg := domain.HoneymoonBridgeConfig{Target: 150}
	g.On("GetConfig").Return(cfg)
	assert.Equal(t, cfg, i.GetConfig())
}

func TestRestoreHoneymoonBridgeInteractor(t *testing.T) {
	src := domain.NewDefaultHoneymoonBridge()
	src.Reset()
	data, err := json.Marshal(src)
	require.NoError(t, err)

	restored, err := RestoreHoneymoonBridgeInteractor(data, new(presenter.MockHoneymoonBridgePresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
	assert.Equal(t, src.GetRoundNumber(), restored.Game.GetRoundNumber())
	assert.Equal(t, src.GetStockSize(), restored.Game.GetStockSize(), "山札が消えない")
	for i := range domain.HoneymoonBridgePlayerCnt {
		assert.Equal(t, src.GetPlayer(i).GetCardsSize(), restored.Game.GetPlayer(i).GetCardsSize())
	}

	_, err = RestoreHoneymoonBridgeInteractor([]byte("{"), new(presenter.MockHoneymoonBridgePresenter))
	assert.Error(t, err)
}
