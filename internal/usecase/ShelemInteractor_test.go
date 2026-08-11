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

func newMockShelemGame() *interfaces.MockShelemGame { return new(interfaces.MockShelemGame) }

func newMockShelemPresenter() *presenter.MockShelemPresenter {
	return new(presenter.MockShelemPresenter)
}

func TestNewShelemInteractor(t *testing.T) {
	assert.NotNil(t, NewShelemInteractor(newMockShelemGame(), newMockShelemPresenter()))
}

func TestNewShelemInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewShelemInteractor(nil, newMockShelemPresenter()) })
	assert.Panics(t, func() { NewShelemInteractor(newMockShelemGame(), nil) })
}

// **Reset は競りを回し、そのままプレイも進める。** CPU が落札するとウィドウ
// 交換まで競りの締めで終わるので、そこで止めると誰も打たない盤面を返す。
func TestShelemInteractorResetWalksBiddingThenPlay(t *testing.T) {
	g := newMockShelemGame()
	p := newMockShelemPresenter()
	i := NewShelemInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.ShelemPhaseBid).Times(3)
	g.On("IsHumanBidTurn").Return(false).Twice()
	g.On("IsHumanBidTurn").Return(true)
	g.On("CpuBid").Return()
	g.On("GetPhase").Return(domain.ShelemPhasePlay)
	g.On("IsHumanTurn").Return(true)
	g.On("GetHint").Return(nil).Maybe()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertNumberOfCalls(t, "CpuBid", 2)
}

// **人間が落札したら捨て札フェーズで止まる。** 勝手に捨てない。
func TestShelemInteractorStopsAtTheHumanDiscard(t *testing.T) {
	g := newMockShelemGame()
	p := newMockShelemPresenter()
	i := NewShelemInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.ShelemPhaseDiscard)
	g.On("IsHumanBidTurn").Return(false).Maybe()
	g.On("IsHumanTurn").Return(false).Maybe()
	g.On("GetHint").Return(nil).Maybe()
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNotCalled(t, "CpuBid")
	g.AssertNotCalled(t, "CpuPlay")
}

// **入札・Shelem・降りはそれぞれ別のドメイン操作に落ちる。**
func TestShelemInteractorBiddingCommandsAreDistinct(t *testing.T) {
	for _, tc := range []struct {
		name   string
		call   func(*ShelemInteractor) string
		method string
		args   []any
		others []string
	}{
		{"bid", func(i *ShelemInteractor) string { return i.Bid(80) }, "PlayerBid", []any{80}, []string{"PlayerBidShelem", "PlayerPass"}},
		{"shelem", func(i *ShelemInteractor) string { return i.BidShelem() }, "PlayerBidShelem", nil, []string{"PlayerBid", "PlayerPass"}},
		{"pass", func(i *ShelemInteractor) string { return i.Pass() }, "PlayerPass", nil, []string{"PlayerBid", "PlayerBidShelem"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockShelemGame()
			p := newMockShelemPresenter()
			i := NewShelemInteractor(g, p)

			g.On("GetGameEndFlag").Return(false)
			g.On("PlayerBid", 80).Return(nil)
			g.On("PlayerBidShelem").Return(nil)
			g.On("PlayerPass").Return(nil)
			g.On("GetPhase").Return(domain.ShelemPhasePlay)
			g.On("IsHumanBidTurn").Return(true)
			g.On("IsHumanTurn").Return(true)
			p.On("Output", g, nil).Return("out")

			assert.Equal(t, "out", tc.call(i))
			g.AssertCalled(t, tc.method, tc.args...)
			for _, other := range tc.others {
				g.AssertNotCalled(t, other)
			}
		})
	}
}

// **捨て札は 4 つのインデックスとスートをそのまま渡す。**
func TestShelemInteractorDiscardPassesBothArguments(t *testing.T) {
	g := newMockShelemGame()
	p := newMockShelemPresenter()
	i := NewShelemInteractor(g, p)

	indices := []int{0, 3, 7, 11}
	g.On("GetGameEndFlag").Return(false)
	g.On("PlayerDiscard", indices, domain.CardDesignHeart).Return(nil)
	g.On("GetPhase").Return(domain.ShelemPhasePlay)
	g.On("IsHumanBidTurn").Return(true)
	g.On("IsHumanTurn").Return(true)
	p.On("Output", g, nil).Return("discarded")

	assert.Equal(t, "discarded", i.Discard(indices, domain.CardDesignHeart))
	g.AssertCalled(t, "PlayerDiscard", indices, domain.CardDesignHeart)
}

func TestShelemInteractorRejectedBid(t *testing.T) {
	g := newMockShelemGame()
	p := newMockShelemPresenter()
	i := NewShelemInteractor(g, p)

	err := errors.New("bid must beat 90")
	g.On("GetGameEndFlag").Return(false)
	g.On("PlayerBid", 80).Return(err)
	p.On("Output", g, err).Return("bid_error")

	assert.Equal(t, "bid_error", i.Bid(80))
	g.AssertNotCalled(t, "CpuBid")
	g.AssertNotCalled(t, "CpuPlay")
}

func TestShelemInteractorPlay(t *testing.T) {
	g := newMockShelemGame()
	p := newMockShelemPresenter()
	i := NewShelemInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	g.On("PlayerPlay", 2).Return(nil)
	p.On("Output", g, nil).Return("played")

	assert.Equal(t, "played", i.Play(2))
	g.AssertCalled(t, "PlayerPlay", 2)
}

func TestShelemInteractorPlayBlocked(t *testing.T) {
	for _, tc := range []struct {
		name    string
		ended   bool
		isHuman bool
	}{
		{"game over", true, true},
		{"not human turn", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockShelemGame()
			p := newMockShelemPresenter()
			i := NewShelemInteractor(g, p)

			g.On("GetGameEndFlag").Return(tc.ended)
			g.On("IsHumanTurn").Return(tc.isHuman)
			p.On("Output", g, nil).Return("blocked")

			assert.Equal(t, "blocked", i.Play(0))
			g.AssertNotCalled(t, "PlayerPlay", 0)
		})
	}
}

func TestShelemInteractorCpuLoopsHaveAnUpperBound(t *testing.T) {
	t.Run("bids", func(t *testing.T) {
		g := newMockShelemGame()
		p := newMockShelemPresenter()
		i := NewShelemInteractor(g, p)

		g.On("Reset").Return()
		g.On("GetGameEndFlag").Return(false)
		g.On("GetPhase").Return(domain.ShelemPhaseBid)
		g.On("IsHumanBidTurn").Return(false)
		g.On("CpuBid").Return()
		g.On("IsHumanTurn").Return(true).Maybe()
		g.On("GetHint").Return(nil).Maybe()
		p.On("Output", g, nil).Return("out")

		assert.Equal(t, "out", i.Reset())
		g.AssertNumberOfCalls(t, "CpuBid", maxCpuTurnsPerCall)
	})

	t.Run("plays", func(t *testing.T) {
		g := newMockShelemGame()
		p := newMockShelemPresenter()
		i := NewShelemInteractor(g, p)

		g.On("GetGameEndFlag").Return(false)
		g.On("PlayerPass").Return(nil)
		g.On("GetPhase").Return(domain.ShelemPhasePlay)
		g.On("IsHumanBidTurn").Return(true)
		g.On("IsHumanTurn").Return(false)
		g.On("CpuPlay").Return()
		p.On("Output", g, nil).Return("out")

		assert.Equal(t, "out", i.Pass())
		g.AssertNumberOfCalls(t, "CpuPlay", maxCpuTurnsPerCall)
	})
}

func TestShelemInteractorNextRound(t *testing.T) {
	g := newMockShelemGame()
	p := newMockShelemPresenter()
	i := NewShelemInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("NextRound").Return()
	g.On("GetPhase").Return(domain.ShelemPhaseBid)
	g.On("IsHumanBidTurn").Return(true)
	g.On("IsHumanTurn").Return(false)
	p.On("Output", g, nil).Return("next")

	assert.Equal(t, "next", i.NextRound())
	g.AssertCalled(t, "NextRound")
}

func TestShelemInteractorGuardsAfterGameEnd(t *testing.T) {
	for _, tc := range []struct {
		name   string
		call   func(*ShelemInteractor) string
		method string
	}{
		{"next round", func(i *ShelemInteractor) string { return i.NextRound() }, "NextRound"},
		{"give up", func(i *ShelemInteractor) string { return i.GiveUp() }, "GiveUp"},
		{"bid", func(i *ShelemInteractor) string { return i.Bid(80) }, "PlayerBid"},
		{"shelem", func(i *ShelemInteractor) string { return i.BidShelem() }, "PlayerBidShelem"},
		{"pass", func(i *ShelemInteractor) string { return i.Pass() }, "PlayerPass"},
		{"discard", func(i *ShelemInteractor) string { return i.Discard([]int{0, 1, 2, 3}, 1) }, "PlayerDiscard"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockShelemGame()
			p := newMockShelemPresenter()
			i := NewShelemInteractor(g, p)

			g.On("GetGameEndFlag").Return(true)
			p.On("Output", g, nil).Return("over")

			assert.Equal(t, "over", tc.call(i))
			g.AssertNotCalled(t, tc.method)
		})
	}
}

func TestShelemInteractorGiveUp(t *testing.T) {
	g := newMockShelemGame()
	p := newMockShelemPresenter()
	i := NewShelemInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("gave_up")

	assert.Equal(t, "gave_up", i.GiveUp())
	g.AssertCalled(t, "GiveUp")
}

func TestShelemInteractorGetConfig(t *testing.T) {
	g := newMockShelemGame()
	i := NewShelemInteractor(g, newMockShelemPresenter())
	cfg := domain.ShelemConfig{Target: 700}
	g.On("GetConfig").Return(cfg)
	assert.Equal(t, cfg, i.GetConfig())
}

func TestShelemInteractorResetWithInvalidConfig(t *testing.T) {
	g := newMockShelemGame()
	p := newMockShelemPresenter()
	i := NewShelemInteractor(g, p)

	p.On("Output", mock.Anything, mock.Anything).Return("cfg_error")

	assert.Equal(t, "cfg_error", i.ResetWithConfig(domain.ShelemConfig{Target: 0}))
	g.AssertNotCalled(t, "Reset")
	g.AssertNotCalled(t, "SetConfig", mock.Anything)
}

func TestShelemInteractorHintAndActionLog(t *testing.T) {
	g := newMockShelemGame()
	p := newMockShelemPresenter()
	i := NewShelemInteractor(g, p)

	p.On("HintOutput", g).Return("hint_output")
	p.On("ActionLogOutput", g).Return("log_output")

	assert.Equal(t, "hint_output", i.Hint())
	assert.Equal(t, "log_output", i.ActionLog())
}

// Worker はリクエストごとに KV から作り直す。**契約と切り札が往復しないと
// 精算できない** (#4478)。
func TestShelemInteractorSnapshotRoundTrip(t *testing.T) {
	g := domain.NewDefaultShelem()
	g.Reset()
	g.SetContractForTest(0, 90, false)
	g.CloseBiddingForTest()
	require.NoError(t, g.PlayerDiscard([]int{0, 1, 2, 3}, domain.CardDesignClover))
	g.SetScoreForTestUse(0, 260)

	p := newMockShelemPresenter()
	i := NewShelemInteractor(g, p)

	data, err := i.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreShelemInteractor(data, p)
	require.NoError(t, err)

	assert.Equal(t, domain.CardDesignClover, restored.Game.GetTrumpSuit())
	assert.Equal(t, 90, restored.Game.GetContract())
	assert.Equal(t, 0, restored.Game.GetDeclarerIdx())
	assert.Equal(t, 260, restored.Game.GetScore(0))
	assert.Equal(t, g.GetConfig().Target, restored.Game.GetConfig().Target)
}

func TestRestoreShelemInteractorRejectsGarbage(t *testing.T) {
	_, err := RestoreShelemInteractor([]byte("not json"), newMockShelemPresenter())
	assert.Error(t, err)
}
