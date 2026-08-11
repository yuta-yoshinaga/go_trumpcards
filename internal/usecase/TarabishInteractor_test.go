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

func newMockTarabishGame() *interfaces.MockTarabishGame { return new(interfaces.MockTarabishGame) }

func newMockTarabishPresenter() *presenter.MockTarabishPresenter {
	return new(presenter.MockTarabishPresenter)
}

func TestNewTarabishInteractor(t *testing.T) {
	assert.NotNil(t, NewTarabishInteractor(newMockTarabishGame(), newMockTarabishPresenter()))
}

func TestNewTarabishInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewTarabishInteractor(nil, newMockTarabishPresenter()) })
	assert.Panics(t, func() { NewTarabishInteractor(newMockTarabishGame(), nil) })
}

// Reset のあと、人間の入札番になるまで CPU が入札する。
func TestTarabishInteractorResetRunsCpuBidsToHumanTurn(t *testing.T) {
	g := newMockTarabishGame()
	p := newMockTarabishPresenter()
	i := NewTarabishInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.TarabishPhaseBid)
	g.On("IsHumanBidTurn").Return(false).Twice()
	g.On("IsHumanBidTurn").Return(true)
	g.On("CpuBid").Return()
	// Reset は入札のあとプレイも進めるので、手番判定も呼ばれる。
	g.On("IsHumanTurn").Return(true)
	g.On("GetHint").Return(nil).Maybe()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertNumberOfCalls(t, "CpuBid", 2)
	g.AssertNotCalled(t, "CpuPlay")
}

// **CPU が入札で切り札を取ったら、そのままプレイも進める。**
//
// ここで止めると、リード（親の左隣＝CPU）のまま誰も打たず、人間の手番が
// 永久に来ない盤面を返してしまう。e2e がこれを 3 回に 1 回踏んでいた。
func TestTarabishInteractorResetPlaysOnWhenACpuTakesTrump(t *testing.T) {
	g := newMockTarabishGame()
	p := newMockTarabishPresenter()
	i := NewTarabishInteractor(g, p)

	// Reset の直後には既に切り札が決まっている（CPU が取った）状態を作る。
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.TarabishPhasePlay)
	g.On("IsHumanBidTurn").Return(true)
	g.On("IsHumanTurn").Return(false).Times(3)
	g.On("IsHumanTurn").Return(true)
	g.On("CpuPlay").Return()
	g.On("GetHint").Return(nil).Maybe()
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNumberOfCalls(t, "CpuPlay", 3)
}

// **Take と Pass はそれぞれ別のドメイン操作に落ちる。** 取り違えると
// 引き受けたはずの切り札を見送ることになる。
func TestTarabishInteractorTakeAndPassAreDistinct(t *testing.T) {
	for _, tc := range []struct {
		name   string
		call   func(*TarabishInteractor) string
		method string
		other  string
	}{
		{"take", func(i *TarabishInteractor) string { return i.TakeTrump() }, "TakeTrump", "PassTrump"},
		{"pass", func(i *TarabishInteractor) string { return i.PassTrump() }, "PassTrump", "TakeTrump"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockTarabishGame()
			p := newMockTarabishPresenter()
			i := NewTarabishInteractor(g, p)

			g.On("GetGameEndFlag").Return(false)
			g.On(tc.method).Return(nil)
			g.On("GetPhase").Return(domain.TarabishPhasePlay)
			g.On("IsHumanTurn").Return(true)
			p.On("Output", g, nil).Return("bid_output")

			assert.Equal(t, "bid_output", tc.call(i))
			g.AssertCalled(t, tc.method)
			g.AssertNotCalled(t, tc.other)
		})
	}
}

// **親が見送ろうとしたらドメインが拒否し、その error が presenter に渡る。**
func TestTarabishInteractorPassRejectedForDealer(t *testing.T) {
	g := newMockTarabishGame()
	p := newMockTarabishPresenter()
	i := NewTarabishInteractor(g, p)

	err := errors.New("the dealer must take the trump")
	g.On("GetGameEndFlag").Return(false)
	g.On("PassTrump").Return(err)
	p.On("Output", g, err).Return("pass_error")

	assert.Equal(t, "pass_error", i.PassTrump())
	g.AssertNotCalled(t, "CpuBid")
	g.AssertNotCalled(t, "CpuPlay")
}

func TestTarabishInteractorPlay(t *testing.T) {
	g := newMockTarabishGame()
	p := newMockTarabishPresenter()
	i := NewTarabishInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	g.On("PlayerPlay", 2).Return(nil)
	p.On("Output", g, nil).Return("played")

	assert.Equal(t, "played", i.Play(2))
	g.AssertCalled(t, "PlayerPlay", 2)
}

func TestTarabishInteractorPlayError(t *testing.T) {
	g := newMockTarabishGame()
	p := newMockTarabishPresenter()
	i := NewTarabishInteractor(g, p)

	err := errors.New("must follow suit")
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	g.On("PlayerPlay", 2).Return(err)
	p.On("Output", g, err).Return("error_output")

	assert.Equal(t, "error_output", i.Play(2))
	g.AssertNotCalled(t, "CpuPlay")
}

func TestTarabishInteractorPlayBlocked(t *testing.T) {
	for _, tc := range []struct {
		name    string
		ended   bool
		isHuman bool
	}{
		{"game over", true, true},
		{"not human turn", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockTarabishGame()
			p := newMockTarabishPresenter()
			i := NewTarabishInteractor(g, p)

			g.On("GetGameEndFlag").Return(tc.ended)
			g.On("IsHumanTurn").Return(tc.isHuman)
			p.On("Output", g, nil).Return("blocked")

			assert.Equal(t, "blocked", i.Play(0))
			g.AssertNotCalled(t, "PlayerPlay", 0)
		})
	}
}

// **プレイ中は入札ループを回さない。** フェーズが違えば即座に抜ける。
func TestTarabishInteractorBidLoopStopsOutsideBidPhase(t *testing.T) {
	g := newMockTarabishGame()
	p := newMockTarabishPresenter()
	i := NewTarabishInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("TakeTrump").Return(nil)
	g.On("GetPhase").Return(domain.TarabishPhaseRoundEnd)
	g.On("IsHumanTurn").Return(false)
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.TakeTrump())
	g.AssertNotCalled(t, "CpuBid")
	g.AssertNotCalled(t, "CpuPlay")
}

func TestTarabishInteractorCpuLoopsHaveAnUpperBound(t *testing.T) {
	t.Run("bids", func(t *testing.T) {
		g := newMockTarabishGame()
		p := newMockTarabishPresenter()
		i := NewTarabishInteractor(g, p)

		g.On("Reset").Return()
		g.On("GetGameEndFlag").Return(false)
		g.On("GetPhase").Return(domain.TarabishPhaseBid)
		g.On("IsHumanBidTurn").Return(false)
		g.On("CpuBid").Return()
		// 入札ループが上限で抜けたあと、Reset はプレイ側も試す。
		g.On("IsHumanTurn").Return(true)
		g.On("GetHint").Return(nil).Maybe()
		p.On("Output", g, nil).Return("out")

		assert.Equal(t, "out", i.Reset())
		g.AssertNumberOfCalls(t, "CpuBid", maxCpuTurnsPerCall)
	})

	t.Run("plays", func(t *testing.T) {
		g := newMockTarabishGame()
		p := newMockTarabishPresenter()
		i := NewTarabishInteractor(g, p)

		g.On("GetGameEndFlag").Return(false)
		g.On("TakeTrump").Return(nil)
		g.On("GetPhase").Return(domain.TarabishPhasePlay)
		g.On("IsHumanBidTurn").Return(true)
		g.On("IsHumanTurn").Return(false)
		g.On("CpuPlay").Return()
		p.On("Output", g, nil).Return("out")

		assert.Equal(t, "out", i.TakeTrump())
		g.AssertNumberOfCalls(t, "CpuPlay", maxCpuTurnsPerCall)
	})
}

func TestTarabishInteractorNextRound(t *testing.T) {
	g := newMockTarabishGame()
	p := newMockTarabishPresenter()
	i := NewTarabishInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("NextRound").Return()
	g.On("GetPhase").Return(domain.TarabishPhaseBid)
	g.On("IsHumanBidTurn").Return(true)
	g.On("IsHumanTurn").Return(false)
	p.On("Output", g, nil).Return("next")

	assert.Equal(t, "next", i.NextRound())
	g.AssertCalled(t, "NextRound")
}

func TestTarabishInteractorGuardsAfterGameEnd(t *testing.T) {
	for _, tc := range []struct {
		name   string
		call   func(*TarabishInteractor) string
		method string
	}{
		{"next round", func(i *TarabishInteractor) string { return i.NextRound() }, "NextRound"},
		{"give up", func(i *TarabishInteractor) string { return i.GiveUp() }, "GiveUp"},
		{"take trump", func(i *TarabishInteractor) string { return i.TakeTrump() }, "TakeTrump"},
		{"pass trump", func(i *TarabishInteractor) string { return i.PassTrump() }, "PassTrump"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockTarabishGame()
			p := newMockTarabishPresenter()
			i := NewTarabishInteractor(g, p)

			g.On("GetGameEndFlag").Return(true)
			p.On("Output", g, nil).Return("over")

			assert.Equal(t, "over", tc.call(i))
			g.AssertNotCalled(t, tc.method)
		})
	}
}

func TestTarabishInteractorGiveUp(t *testing.T) {
	g := newMockTarabishGame()
	p := newMockTarabishPresenter()
	i := NewTarabishInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("gave_up")

	assert.Equal(t, "gave_up", i.GiveUp())
	g.AssertCalled(t, "GiveUp")
}

func TestTarabishInteractorGetConfig(t *testing.T) {
	g := newMockTarabishGame()
	i := NewTarabishInteractor(g, newMockTarabishPresenter())
	cfg := domain.TarabishConfig{Target: 300}
	g.On("GetConfig").Return(cfg)
	assert.Equal(t, cfg, i.GetConfig())
}

func TestTarabishInteractorResetWithInvalidConfig(t *testing.T) {
	g := newMockTarabishGame()
	p := newMockTarabishPresenter()
	i := NewTarabishInteractor(g, p)

	p.On("Output", mock.Anything, mock.Anything).Return("cfg_error")

	assert.Equal(t, "cfg_error", i.ResetWithConfig(domain.TarabishConfig{Target: 5}))
	g.AssertNotCalled(t, "Reset")
	g.AssertNotCalled(t, "SetConfig", mock.Anything)
}

func TestTarabishInteractorHintAndActionLog(t *testing.T) {
	g := newMockTarabishGame()
	p := newMockTarabishPresenter()
	i := NewTarabishInteractor(g, p)

	p.On("HintOutput", g).Return("hint_output")
	p.On("ActionLogOutput", g).Return("log_output")

	assert.Equal(t, "hint_output", i.Hint())
	assert.Equal(t, "log_output", i.ActionLog())
}

// Worker はリクエストごとに KV から作り直す。**メルド点と切り札が往復しないと
// 得点が合わなくなる** (#4478)。
func TestTarabishInteractorSnapshotRoundTrip(t *testing.T) {
	g := domain.NewDefaultTarabish()
	g.Reset()
	g.SetCurrentPlayerIdxForTest(0)
	require.NoError(t, g.TakeTrump())
	g.SetScoreForTestUse(0, 140)
	g.GetPlayer(0).SetMeldPoints(50)

	p := newMockTarabishPresenter()
	i := NewTarabishInteractor(g, p)

	data, err := i.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreTarabishInteractor(data, p)
	require.NoError(t, err)

	assert.Equal(t, 140, restored.Game.GetScore(0))
	assert.Equal(t, 50, restored.Game.GetPlayer(0).GetMeldPoints())
	assert.Equal(t, g.GetTrumpSuit(), restored.Game.GetTrumpSuit())
	assert.Equal(t, g.GetTrumpTakerIdx(), restored.Game.GetTrumpTakerIdx())
	assert.Equal(t, g.GetConfig().Target, restored.Game.GetConfig().Target)
}

func TestRestoreTarabishInteractorRejectsGarbage(t *testing.T) {
	_, err := RestoreTarabishInteractor([]byte("not json"), newMockTarabishPresenter())
	assert.Error(t, err)
}
