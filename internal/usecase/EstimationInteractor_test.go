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

func newMockEstimationGame() *interfaces.MockEstimationGame {
	return new(interfaces.MockEstimationGame)
}

func newMockEstimationPresenter() *presenter.MockEstimationPresenter {
	return new(presenter.MockEstimationPresenter)
}

func TestNewEstimationInteractor(t *testing.T) {
	assert.NotNil(t, NewEstimationInteractor(newMockEstimationGame(), newMockEstimationPresenter()))
}

func TestNewEstimationInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewEstimationInteractor(nil, newMockEstimationPresenter()) })
	assert.Panics(t, func() { NewEstimationInteractor(newMockEstimationGame(), nil) })
}

// **Reset は切り札選択・宣言・プレイの 3 段を順に進める。** どれかを省くと
// CPU の途中で止まった盤面を返し、人間の手番が永久に来ない。
func TestEstimationInteractorResetWalksAllThreeStages(t *testing.T) {
	g := newMockEstimationGame()
	p := newMockEstimationPresenter()
	i := NewEstimationInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	// 切り札 → 宣言 → プレイ とフェーズが進む。
	g.On("GetPhase").Return(domain.EstimationPhaseTrump).Once()
	g.On("IsHumanTrumpTurn").Return(false)
	g.On("CpuSelectTrump").Return()
	g.On("GetPhase").Return(domain.EstimationPhaseBid).Times(3)
	g.On("IsHumanBidTurn").Return(false).Twice()
	g.On("IsHumanBidTurn").Return(true)
	g.On("CpuBid").Return()
	g.On("GetPhase").Return(domain.EstimationPhasePlay)
	g.On("IsHumanTurn").Return(true)
	g.On("GetHint").Return(nil).Maybe()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertCalled(t, "CpuSelectTrump")
	g.AssertNumberOfCalls(t, "CpuBid", 2)
}

// **CPU が宣言まで終えたら、そのままプレイも進める。**
func TestEstimationInteractorResetPlaysOnAfterBidding(t *testing.T) {
	g := newMockEstimationGame()
	p := newMockEstimationPresenter()
	i := NewEstimationInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.EstimationPhasePlay)
	g.On("IsHumanTrumpTurn").Return(true)
	g.On("IsHumanBidTurn").Return(true)
	g.On("IsHumanTurn").Return(false).Times(3)
	g.On("IsHumanTurn").Return(true)
	g.On("CpuPlay").Return()
	g.On("GetHint").Return(nil).Maybe()
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNumberOfCalls(t, "CpuPlay", 3)
	g.AssertNotCalled(t, "CpuSelectTrump")
}

// **切り札と宣言は別の値を別のドメイン操作に渡す。** 取り違えると
// 選んでいないスートが切り札になったり、意図しない数を宣言する。
func TestEstimationInteractorSelectTrumpAndBidAreDistinct(t *testing.T) {
	g := newMockEstimationGame()
	p := newMockEstimationPresenter()
	i := NewEstimationInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("SelectTrump", domain.CardDesignHeart).Return(nil)
	g.On("PlayerBid", 4).Return(nil)
	g.On("GetPhase").Return(domain.EstimationPhasePlay)
	g.On("IsHumanTrumpTurn").Return(true)
	g.On("IsHumanBidTurn").Return(true)
	g.On("IsHumanTurn").Return(true)
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.SelectTrump(domain.CardDesignHeart))
	assert.Equal(t, "out", i.Bid(4))
	g.AssertCalled(t, "SelectTrump", domain.CardDesignHeart)
	g.AssertCalled(t, "PlayerBid", 4)
}

// **0 の宣言はそのまま 0 として渡る。** Dash Call は「未指定」ではない。
func TestEstimationInteractorZeroBidReachesTheDomain(t *testing.T) {
	g := newMockEstimationGame()
	p := newMockEstimationPresenter()
	i := NewEstimationInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("PlayerBid", 0).Return(nil)
	g.On("GetPhase").Return(domain.EstimationPhasePlay)
	g.On("IsHumanTrumpTurn").Return(true)
	g.On("IsHumanBidTurn").Return(true)
	g.On("IsHumanTurn").Return(true)
	p.On("Output", g, nil).Return("dash")

	assert.Equal(t, "dash", i.Bid(0))
	g.AssertCalled(t, "PlayerBid", 0)
}

// ドメインが宣言を拒否したら、その error が presenter に渡り CPU は進まない。
func TestEstimationInteractorBidRejected(t *testing.T) {
	g := newMockEstimationGame()
	p := newMockEstimationPresenter()
	i := NewEstimationInteractor(g, p)

	err := errors.New("the last bidder cannot make the total 13")
	g.On("GetGameEndFlag").Return(false)
	g.On("PlayerBid", 1).Return(err)
	p.On("Output", g, err).Return("bid_error")

	assert.Equal(t, "bid_error", i.Bid(1))
	g.AssertNotCalled(t, "CpuBid")
	g.AssertNotCalled(t, "CpuPlay")
}

func TestEstimationInteractorPlay(t *testing.T) {
	g := newMockEstimationGame()
	p := newMockEstimationPresenter()
	i := NewEstimationInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	g.On("PlayerPlay", 2).Return(nil)
	p.On("Output", g, nil).Return("played")

	assert.Equal(t, "played", i.Play(2))
	g.AssertCalled(t, "PlayerPlay", 2)
}

func TestEstimationInteractorPlayError(t *testing.T) {
	g := newMockEstimationGame()
	p := newMockEstimationPresenter()
	i := NewEstimationInteractor(g, p)

	err := errors.New("must follow suit")
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	g.On("PlayerPlay", 2).Return(err)
	p.On("Output", g, err).Return("error_output")

	assert.Equal(t, "error_output", i.Play(2))
	g.AssertNotCalled(t, "CpuPlay")
}

func TestEstimationInteractorPlayBlocked(t *testing.T) {
	for _, tc := range []struct {
		name    string
		ended   bool
		isHuman bool
	}{
		{"game over", true, true},
		{"not human turn", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockEstimationGame()
			p := newMockEstimationPresenter()
			i := NewEstimationInteractor(g, p)

			g.On("GetGameEndFlag").Return(tc.ended)
			g.On("IsHumanTurn").Return(tc.isHuman)
			p.On("Output", g, nil).Return("blocked")

			assert.Equal(t, "blocked", i.Play(0))
			g.AssertNotCalled(t, "PlayerPlay", 0)
		})
	}
}

// フェーズが違えば各ループは即座に抜ける。
func TestEstimationInteractorLoopsStopOutsideTheirPhase(t *testing.T) {
	g := newMockEstimationGame()
	p := newMockEstimationPresenter()
	i := NewEstimationInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("SelectTrump", mock.Anything).Return(nil)
	g.On("GetPhase").Return(domain.EstimationPhaseRoundEnd)
	g.On("IsHumanTurn").Return(false)
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.SelectTrump(domain.CardDesignSpade))
	g.AssertNotCalled(t, "CpuSelectTrump")
	g.AssertNotCalled(t, "CpuBid")
	g.AssertNotCalled(t, "CpuPlay")
}

func TestEstimationInteractorCpuLoopsHaveAnUpperBound(t *testing.T) {
	t.Run("bids", func(t *testing.T) {
		g := newMockEstimationGame()
		p := newMockEstimationPresenter()
		i := NewEstimationInteractor(g, p)

		g.On("Reset").Return()
		g.On("GetGameEndFlag").Return(false)
		g.On("GetPhase").Return(domain.EstimationPhaseBid)
		g.On("IsHumanTrumpTurn").Return(true)
		g.On("IsHumanBidTurn").Return(false)
		g.On("CpuBid").Return()
		g.On("IsHumanTurn").Return(true)
		g.On("GetHint").Return(nil).Maybe()
		p.On("Output", g, nil).Return("out")

		assert.Equal(t, "out", i.Reset())
		g.AssertNumberOfCalls(t, "CpuBid", maxCpuTurnsPerCall)
	})

	t.Run("plays", func(t *testing.T) {
		g := newMockEstimationGame()
		p := newMockEstimationPresenter()
		i := NewEstimationInteractor(g, p)

		g.On("GetGameEndFlag").Return(false)
		g.On("PlayerBid", 3).Return(nil)
		g.On("GetPhase").Return(domain.EstimationPhasePlay)
		g.On("IsHumanTrumpTurn").Return(true)
		g.On("IsHumanBidTurn").Return(true)
		g.On("IsHumanTurn").Return(false)
		g.On("CpuPlay").Return()
		p.On("Output", g, nil).Return("out")

		assert.Equal(t, "out", i.Bid(3))
		g.AssertNumberOfCalls(t, "CpuPlay", maxCpuTurnsPerCall)
	})
}

func TestEstimationInteractorNextRound(t *testing.T) {
	g := newMockEstimationGame()
	p := newMockEstimationPresenter()
	i := NewEstimationInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("NextRound").Return()
	g.On("GetPhase").Return(domain.EstimationPhaseTrump)
	g.On("IsHumanTrumpTurn").Return(true)
	g.On("IsHumanTurn").Return(false)
	p.On("Output", g, nil).Return("next")

	assert.Equal(t, "next", i.NextRound())
	g.AssertCalled(t, "NextRound")
}

func TestEstimationInteractorGuardsAfterGameEnd(t *testing.T) {
	for _, tc := range []struct {
		name   string
		call   func(*EstimationInteractor) string
		method string
	}{
		{"next round", func(i *EstimationInteractor) string { return i.NextRound() }, "NextRound"},
		{"give up", func(i *EstimationInteractor) string { return i.GiveUp() }, "GiveUp"},
		{"select trump", func(i *EstimationInteractor) string { return i.SelectTrump(1) }, "SelectTrump"},
		{"bid", func(i *EstimationInteractor) string { return i.Bid(3) }, "PlayerBid"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockEstimationGame()
			p := newMockEstimationPresenter()
			i := NewEstimationInteractor(g, p)

			g.On("GetGameEndFlag").Return(true)
			p.On("Output", g, nil).Return("over")

			assert.Equal(t, "over", tc.call(i))
			g.AssertNotCalled(t, tc.method)
		})
	}
}

func TestEstimationInteractorGiveUp(t *testing.T) {
	g := newMockEstimationGame()
	p := newMockEstimationPresenter()
	i := NewEstimationInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("gave_up")

	assert.Equal(t, "gave_up", i.GiveUp())
	g.AssertCalled(t, "GiveUp")
}

func TestEstimationInteractorGetConfig(t *testing.T) {
	g := newMockEstimationGame()
	i := NewEstimationInteractor(g, newMockEstimationPresenter())
	cfg := domain.EstimationConfig{Rounds: 7}
	g.On("GetConfig").Return(cfg)
	assert.Equal(t, cfg, i.GetConfig())
}

func TestEstimationInteractorResetWithInvalidConfig(t *testing.T) {
	g := newMockEstimationGame()
	p := newMockEstimationPresenter()
	i := NewEstimationInteractor(g, p)

	p.On("Output", mock.Anything, mock.Anything).Return("cfg_error")

	assert.Equal(t, "cfg_error", i.ResetWithConfig(domain.EstimationConfig{Rounds: 0}))
	g.AssertNotCalled(t, "Reset")
	g.AssertNotCalled(t, "SetConfig", mock.Anything)
}

func TestEstimationInteractorHintAndActionLog(t *testing.T) {
	g := newMockEstimationGame()
	p := newMockEstimationPresenter()
	i := NewEstimationInteractor(g, p)

	p.On("HintOutput", g).Return("hint_output")
	p.On("ActionLogOutput", g).Return("log_output")

	assert.Equal(t, "hint_output", i.Hint())
	assert.Equal(t, "log_output", i.ActionLog())
}

// Worker はリクエストごとに KV から作り直す。**宣言・種別・累計が往復しないと
// 得点が消える** (#4478)。
func TestEstimationInteractorSnapshotRoundTrip(t *testing.T) {
	g := domain.NewDefaultEstimation()
	g.Reset()
	g.SetDealerIdxForTest(0)
	require.NoError(t, g.SelectTrump(domain.CardDesignClover))
	require.NoError(t, g.PlayerBid(5))
	g.GetPlayer(0).SetTotalScore(48)

	p := newMockEstimationPresenter()
	i := NewEstimationInteractor(g, p)

	data, err := i.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreEstimationInteractor(data, p)
	require.NoError(t, err)

	assert.Equal(t, domain.CardDesignClover, restored.Game.GetTrumpSuit())
	assert.Equal(t, 5, restored.Game.GetPlayer(0).GetBid())
	assert.Equal(t, 48, restored.Game.GetPlayer(0).GetTotalScore())
	assert.Equal(t, g.GetConfig().Rounds, restored.Game.GetConfig().Rounds)
	assert.Equal(t, g.GetBidPlayerIdx(), restored.Game.GetBidPlayerIdx())
}

func TestRestoreEstimationInteractorRejectsGarbage(t *testing.T) {
	_, err := RestoreEstimationInteractor([]byte("not json"), newMockEstimationPresenter())
	assert.Error(t, err)
}
