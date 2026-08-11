//go:build test

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

func newMockIsraeliWhistGame() *interfaces.MockIsraeliWhistGame {
	return new(interfaces.MockIsraeliWhistGame)
}

func newMockIsraeliWhistPresenter() *presenter.MockIsraeliWhistPresenter {
	return new(presenter.MockIsraeliWhistPresenter)
}

func TestNewIsraeliWhistInteractor(t *testing.T) {
	assert.NotNil(t, NewIsraeliWhistInteractor(newMockIsraeliWhistGame(), newMockIsraeliWhistPresenter()))
}

func TestNewIsraeliWhistInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewIsraeliWhistInteractor(nil, newMockIsraeliWhistPresenter()) })
	assert.Panics(t, func() { NewIsraeliWhistInteractor(newMockIsraeliWhistGame(), nil) })
}

// **Reset はオークション・宣言・プレイの 3 段を順に進める。** 入札が 2 段階
// あるぶん、どこで止まっても人間の手番が来ない盤面になる。
func TestIsraeliWhistInteractorResetWalksAllThreeStages(t *testing.T) {
	g := newMockIsraeliWhistGame()
	p := newMockIsraeliWhistPresenter()
	i := NewIsraeliWhistInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.IsraeliWhistPhaseAuction).Once()
	g.On("IsHumanAuctionTurn").Return(false)
	g.On("CpuAuction").Return()
	g.On("GetPhase").Return(domain.IsraeliWhistPhaseBid).Times(3)
	g.On("IsHumanBidTurn").Return(false).Twice()
	g.On("IsHumanBidTurn").Return(true)
	g.On("CpuBid").Return()
	g.On("GetPhase").Return(domain.IsraeliWhistPhasePlay)
	g.On("IsHumanTurn").Return(true)
	g.On("GetHint").Return(nil).Maybe()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertCalled(t, "CpuAuction")
	g.AssertNumberOfCalls(t, "CpuBid", 2)
}

// **オークションのループは 4 手では終わらない。** 競り上げが続くぶん回る。
func TestIsraeliWhistInteractorAuctionLoopRunsUntilHumanTurn(t *testing.T) {
	g := newMockIsraeliWhistGame()
	p := newMockIsraeliWhistPresenter()
	i := NewIsraeliWhistInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.IsraeliWhistPhaseAuction)
	g.On("IsHumanAuctionTurn").Return(false).Times(6)
	g.On("IsHumanAuctionTurn").Return(true)
	g.On("CpuAuction").Return()
	g.On("IsHumanTurn").Return(true).Maybe()
	g.On("GetHint").Return(nil).Maybe()
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNumberOfCalls(t, "CpuAuction", 6)
}

// **入札・降り・宣言はそれぞれ別のドメイン操作に落ちる。**
func TestIsraeliWhistInteractorCommandsAreDistinct(t *testing.T) {
	g := newMockIsraeliWhistGame()
	p := newMockIsraeliWhistPresenter()
	i := NewIsraeliWhistInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("PlayerAuctionBid", 7, domain.CardDesignHeart).Return(nil)
	g.On("PlayerAuctionPass").Return(nil)
	g.On("PlayerBid", 4).Return(nil)
	g.On("GetPhase").Return(domain.IsraeliWhistPhasePlay)
	g.On("IsHumanAuctionTurn").Return(true)
	g.On("IsHumanBidTurn").Return(true)
	g.On("IsHumanTurn").Return(true)
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.AuctionBid(7, domain.CardDesignHeart))
	assert.Equal(t, "out", i.AuctionPass())
	assert.Equal(t, "out", i.Bid(4))
	g.AssertCalled(t, "PlayerAuctionBid", 7, domain.CardDesignHeart)
	g.AssertCalled(t, "PlayerAuctionPass")
	g.AssertCalled(t, "PlayerBid", 4)
}

// **入札の数とスートは両方そのまま渡る。** 取り違えると別の入札になる。
func TestIsraeliWhistInteractorAuctionPassesBothArguments(t *testing.T) {
	for _, tc := range []struct{ bid, suit int }{
		{5, domain.CardDesignSpade},
		{9, domain.CardDesignClover},
		{13, domain.CardDesignDiamond},
	} {
		g := newMockIsraeliWhistGame()
		p := newMockIsraeliWhistPresenter()
		i := NewIsraeliWhistInteractor(g, p)

		g.On("GetGameEndFlag").Return(false)
		g.On("PlayerAuctionBid", tc.bid, tc.suit).Return(nil)
		g.On("GetPhase").Return(domain.IsraeliWhistPhasePlay)
		g.On("IsHumanAuctionTurn").Return(true)
		g.On("IsHumanBidTurn").Return(true)
		g.On("IsHumanTurn").Return(true)
		p.On("Output", g, nil).Return("out")

		assert.Equal(t, "out", i.AuctionBid(tc.bid, tc.suit))
		g.AssertCalled(t, "PlayerAuctionBid", tc.bid, tc.suit)
	}
}

// ドメインが拒否したら error が presenter に渡り、CPU は進まない。
func TestIsraeliWhistInteractorRejectedBid(t *testing.T) {
	g := newMockIsraeliWhistGame()
	p := newMockIsraeliWhistPresenter()
	i := NewIsraeliWhistInteractor(g, p)

	err := errors.New("as the declarer you must bid at least 8")
	g.On("GetGameEndFlag").Return(false)
	g.On("PlayerBid", 3).Return(err)
	p.On("Output", g, err).Return("bid_error")

	assert.Equal(t, "bid_error", i.Bid(3))
	g.AssertNotCalled(t, "CpuAuction")
	g.AssertNotCalled(t, "CpuBid")
	g.AssertNotCalled(t, "CpuPlay")
}

func TestIsraeliWhistInteractorPlay(t *testing.T) {
	g := newMockIsraeliWhistGame()
	p := newMockIsraeliWhistPresenter()
	i := NewIsraeliWhistInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	g.On("PlayerPlay", 2).Return(nil)
	p.On("Output", g, nil).Return("played")

	assert.Equal(t, "played", i.Play(2))
	g.AssertCalled(t, "PlayerPlay", 2)
}

func TestIsraeliWhistInteractorPlayBlocked(t *testing.T) {
	for _, tc := range []struct {
		name    string
		ended   bool
		isHuman bool
	}{
		{"game over", true, true},
		{"not human turn", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockIsraeliWhistGame()
			p := newMockIsraeliWhistPresenter()
			i := NewIsraeliWhistInteractor(g, p)

			g.On("GetGameEndFlag").Return(tc.ended)
			g.On("IsHumanTurn").Return(tc.isHuman)
			p.On("Output", g, nil).Return("blocked")

			assert.Equal(t, "blocked", i.Play(0))
			g.AssertNotCalled(t, "PlayerPlay", 0)
		})
	}
}

func TestIsraeliWhistInteractorCpuLoopsHaveAnUpperBound(t *testing.T) {
	for _, tc := range []struct {
		name   string
		phase  domain.IsraeliWhistPhase
		method string
	}{
		{"auction", domain.IsraeliWhistPhaseAuction, "CpuAuction"},
		{"bids", domain.IsraeliWhistPhaseBid, "CpuBid"},
		{"plays", domain.IsraeliWhistPhasePlay, "CpuPlay"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockIsraeliWhistGame()
			p := newMockIsraeliWhistPresenter()
			i := NewIsraeliWhistInteractor(g, p)

			g.On("Reset").Return()
			g.On("GetGameEndFlag").Return(false)
			g.On("GetPhase").Return(tc.phase)
			g.On("IsHumanAuctionTurn").Return(false)
			g.On("IsHumanBidTurn").Return(false)
			g.On("IsHumanTurn").Return(false)
			g.On("CpuAuction").Return()
			g.On("CpuBid").Return()
			g.On("CpuPlay").Return()
			g.On("GetHint").Return(nil).Maybe()
			p.On("Output", g, nil).Return("out")

			assert.Equal(t, "out", i.Reset())
			g.AssertNumberOfCalls(t, tc.method, maxCpuTurnsPerCall)
		})
	}
}

func TestIsraeliWhistInteractorNextRound(t *testing.T) {
	g := newMockIsraeliWhistGame()
	p := newMockIsraeliWhistPresenter()
	i := NewIsraeliWhistInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("NextRound").Return()
	g.On("GetPhase").Return(domain.IsraeliWhistPhaseAuction)
	g.On("IsHumanAuctionTurn").Return(true)
	g.On("IsHumanTurn").Return(false)
	p.On("Output", g, nil).Return("next")

	assert.Equal(t, "next", i.NextRound())
	g.AssertCalled(t, "NextRound")
}

func TestIsraeliWhistInteractorGuardsAfterGameEnd(t *testing.T) {
	for _, tc := range []struct {
		name   string
		call   func(*IsraeliWhistInteractor) string
		method string
	}{
		{"next round", func(i *IsraeliWhistInteractor) string { return i.NextRound() }, "NextRound"},
		{"give up", func(i *IsraeliWhistInteractor) string { return i.GiveUp() }, "GiveUp"},
		{"auction bid", func(i *IsraeliWhistInteractor) string { return i.AuctionBid(7, 1) }, "PlayerAuctionBid"},
		{"auction pass", func(i *IsraeliWhistInteractor) string { return i.AuctionPass() }, "PlayerAuctionPass"},
		{"bid", func(i *IsraeliWhistInteractor) string { return i.Bid(3) }, "PlayerBid"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockIsraeliWhistGame()
			p := newMockIsraeliWhistPresenter()
			i := NewIsraeliWhistInteractor(g, p)

			g.On("GetGameEndFlag").Return(true)
			p.On("Output", g, nil).Return("over")

			assert.Equal(t, "over", tc.call(i))
			g.AssertNotCalled(t, tc.method)
		})
	}
}

func TestIsraeliWhistInteractorGiveUp(t *testing.T) {
	g := newMockIsraeliWhistGame()
	p := newMockIsraeliWhistPresenter()
	i := NewIsraeliWhistInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("gave_up")

	assert.Equal(t, "gave_up", i.GiveUp())
	g.AssertCalled(t, "GiveUp")
}

func TestIsraeliWhistInteractorGetConfig(t *testing.T) {
	g := newMockIsraeliWhistGame()
	i := NewIsraeliWhistInteractor(g, newMockIsraeliWhistPresenter())
	cfg := domain.IsraeliWhistConfig{Rounds: 6}
	g.On("GetConfig").Return(cfg)
	assert.Equal(t, cfg, i.GetConfig())
}

func TestIsraeliWhistInteractorResetWithInvalidConfig(t *testing.T) {
	g := newMockIsraeliWhistGame()
	p := newMockIsraeliWhistPresenter()
	i := NewIsraeliWhistInteractor(g, p)

	p.On("Output", mock.Anything, mock.Anything).Return("cfg_error")

	assert.Equal(t, "cfg_error", i.ResetWithConfig(domain.IsraeliWhistConfig{Rounds: 0}))
	g.AssertNotCalled(t, "Reset")
	g.AssertNotCalled(t, "SetConfig", mock.Anything)
}

func TestIsraeliWhistInteractorHintAndActionLog(t *testing.T) {
	g := newMockIsraeliWhistGame()
	p := newMockIsraeliWhistPresenter()
	i := NewIsraeliWhistInteractor(g, p)

	p.On("HintOutput", g).Return("hint_output")
	p.On("ActionLogOutput", g).Return("log_output")

	assert.Equal(t, "hint_output", i.Hint())
	assert.Equal(t, "log_output", i.ActionLog())
}

// Worker はリクエストごとに KV から作り直す。**2 段階ぶんの入札が往復しないと
// 入札がやり直しになる** (#4478)。
func TestIsraeliWhistInteractorSnapshotRoundTrip(t *testing.T) {
	g := domain.NewDefaultIsraeliWhist()
	g.Reset()
	g.SetAuctionPlayerIdxForTest(0)
	require.NoError(t, g.PlayerAuctionBid(8, domain.CardDesignClover))
	g.SetDeclarerForTest(0, 8, domain.CardDesignClover)
	g.CloseAuctionForTest()
	g.SetBidPlayerIdxForTest(0)
	require.NoError(t, g.PlayerBid(8))
	g.GetPlayer(0).SetTotalScore(74)

	p := newMockIsraeliWhistPresenter()
	i := NewIsraeliWhistInteractor(g, p)

	data, err := i.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreIsraeliWhistInteractor(data, p)
	require.NoError(t, err)

	assert.Equal(t, domain.CardDesignClover, restored.Game.GetTrumpSuit())
	assert.Equal(t, 8, restored.Game.GetHighBid())
	assert.Equal(t, 8, restored.Game.GetPlayer(0).GetBid())
	assert.Equal(t, 8, restored.Game.GetPlayer(0).GetAuctionBid())
	assert.Equal(t, 74, restored.Game.GetPlayer(0).GetTotalScore())
	assert.Equal(t, g.GetConfig().Rounds, restored.Game.GetConfig().Rounds)
}

func TestRestoreIsraeliWhistInteractorRejectsGarbage(t *testing.T) {
	_, err := RestoreIsraeliWhistInteractor([]byte("not json"), newMockIsraeliWhistPresenter())
	assert.Error(t, err)
}
