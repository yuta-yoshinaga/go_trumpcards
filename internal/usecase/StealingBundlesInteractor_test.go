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

func newMockStealingBundlesGame() *interfaces.MockStealingBundlesGame {
	return new(interfaces.MockStealingBundlesGame)
}

func newMockStealingBundlesPresenter() *presenter.MockStealingBundlesPresenter {
	return new(presenter.MockStealingBundlesPresenter)
}

func TestNewStealingBundlesInteractor(t *testing.T) {
	assert.NotNil(t, NewStealingBundlesInteractor(newMockStealingBundlesGame(), newMockStealingBundlesPresenter()))
}

func TestNewStealingBundlesInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewStealingBundlesInteractor(nil, newMockStealingBundlesPresenter()) })
	assert.Panics(t, func() { NewStealingBundlesInteractor(newMockStealingBundlesGame(), nil) })
}

func TestStealingBundlesInteractorResetAdvancesToTheHuman(t *testing.T) {
	g := newMockStealingBundlesGame()
	p := newMockStealingBundlesPresenter()
	i := NewStealingBundlesInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(false).Times(3)
	g.On("IsHumanTurn").Return(true)
	g.On("CpuPlay").Return()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertNumberOfCalls(t, "CpuPlay", 3)
}

func TestStealingBundlesInteractorStopsAtTheHumanTurn(t *testing.T) {
	g := newMockStealingBundlesGame()
	p := newMockStealingBundlesPresenter()
	i := NewStealingBundlesInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNotCalled(t, "CpuPlay")
}

// **3 つの手はどれも 1 手。** 打てたら CPU を進め、弾かれたら盤面は動かない。
func TestStealingBundlesInteractorMoves(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(g *interfaces.MockStealingBundlesGame)
		call  func(i *StealingBundlesInteractor) string
	}{
		{
			"take",
			func(g *interfaces.MockStealingBundlesGame) { g.On("PlayerTake", 2).Return(nil) },
			func(i *StealingBundlesInteractor) string { return i.Take(2) },
		},
		{
			"steal",
			func(g *interfaces.MockStealingBundlesGame) { g.On("PlayerSteal", 1, 3).Return(nil) },
			func(i *StealingBundlesInteractor) string { return i.Steal(1, 3) },
		},
		{
			"trail",
			func(g *interfaces.MockStealingBundlesGame) { g.On("PlayerTrail", 0).Return(nil) },
			func(i *StealingBundlesInteractor) string { return i.Trail(0) },
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockStealingBundlesGame()
			p := newMockStealingBundlesPresenter()
			i := NewStealingBundlesInteractor(g, p)

			g.On("GetGameEndFlag").Return(false)
			g.On("IsHumanTurn").Return(true).Once()
			tc.setup(g)
			g.On("IsHumanTurn").Return(false).Once()
			g.On("CpuPlay").Return()
			g.On("IsHumanTurn").Return(true)
			p.On("Output", g, nil).Return("done")

			assert.Equal(t, "done", tc.call(i))
			g.AssertNumberOfCalls(t, "CpuPlay", 1)
		})
	}
}

func TestStealingBundlesInteractorRejectedMoveLeavesTheBoard(t *testing.T) {
	g := newMockStealingBundlesGame()
	p := newMockStealingBundlesPresenter()
	i := NewStealingBundlesInteractor(g, p)

	err := errors.New("取れる手があるときは場に置けません")
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	g.On("PlayerTrail", 0).Return(err)
	p.On("Output", g, err).Return("blocked")

	assert.Equal(t, "blocked", i.Trail(0))
	g.AssertNotCalled(t, "CpuPlay")
}

func TestStealingBundlesInteractorMovesAfterTheEndAreRejected(t *testing.T) {
	g := newMockStealingBundlesGame()
	p := newMockStealingBundlesPresenter()
	i := NewStealingBundlesInteractor(g, p)

	g.On("GetGameEndFlag").Return(true)
	p.On("Output", g, mock.Anything).Return("ended")

	assert.Equal(t, "ended", i.Take(0))
	assert.Equal(t, "ended", i.Steal(0, 1))
	assert.Equal(t, "ended", i.Trail(0))
	assert.Equal(t, "ended", i.GiveUp())
	g.AssertNotCalled(t, "PlayerTake")
	g.AssertNotCalled(t, "PlayerSteal")
	g.AssertNotCalled(t, "PlayerTrail")
	g.AssertNotCalled(t, "GiveUp")
}

func TestStealingBundlesInteractorGiveUp(t *testing.T) {
	g := newMockStealingBundlesGame()
	p := newMockStealingBundlesPresenter()
	i := NewStealingBundlesInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("gave up")

	assert.Equal(t, "gave up", i.GiveUp())
	g.AssertNumberOfCalls(t, "GiveUp", 1)
}

func TestStealingBundlesInteractorResetWithConfig(t *testing.T) {
	t.Run("妥当な設定は通る", func(t *testing.T) {
		g := newMockStealingBundlesGame()
		p := newMockStealingBundlesPresenter()
		i := NewStealingBundlesInteractor(g, p)

		cfg := domain.StealingBundlesConfig{PlayerCnt: 2}
		g.On("SetConfig", cfg).Return()
		g.On("Reset").Return()
		g.On("GetGameEndFlag").Return(false)
		g.On("IsHumanTurn").Return(true)
		p.On("Output", g, nil).Return("reset")

		assert.Equal(t, "reset", i.ResetWithConfig(cfg))
		g.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("不正な人数は弾かれ、盤面はそのまま", func(t *testing.T) {
		g := newMockStealingBundlesGame()
		p := newMockStealingBundlesPresenter()
		i := NewStealingBundlesInteractor(g, p)

		p.On("Output", g, mock.Anything).Return("bad config")
		assert.Equal(t, "bad config",
			i.ResetWithConfig(domain.StealingBundlesConfig{PlayerCnt: domain.StealingBundlesPlayerCntMax + 1}))
		g.AssertNotCalled(t, "SetConfig")
		g.AssertNotCalled(t, "Reset")
	})
}

func TestStealingBundlesInteractorGetConfigHintAndLog(t *testing.T) {
	g := newMockStealingBundlesGame()
	p := newMockStealingBundlesPresenter()
	i := NewStealingBundlesInteractor(g, p)

	cfg := domain.DefaultStealingBundlesConfig()
	g.On("GetConfig").Return(cfg)
	p.On("HintOutput", g).Return("hint")
	p.On("ActionLogOutput", g).Return("log")

	assert.Equal(t, cfg, i.GetConfig())
	assert.Equal(t, "hint", i.Hint())
	assert.Equal(t, "log", i.ActionLog())
}

// **KV に載らなければ Worker では毎リクエスト初期化される。**
func TestStealingBundlesInteractorSnapshotRoundTrip(t *testing.T) {
	g := domain.NewDefaultStealingBundles()
	g.Reset()
	i := NewStealingBundlesInteractor(g, newMockStealingBundlesPresenter())

	data, err := i.Snapshot()
	require.NoError(t, err)

	back, err := RestoreStealingBundlesInteractor(data, newMockStealingBundlesPresenter())
	require.NoError(t, err)
	assert.Equal(t, g.GetPlayerCnt(), back.Game.GetPlayerCnt())
	assert.Equal(t, g.GetDeckRemaining(), back.Game.GetDeckRemaining())
	assert.Len(t, back.Game.GetTableCards(), len(g.GetTableCards()))
}

func TestRestoreStealingBundlesInteractorRejectsGarbage(t *testing.T) {
	_, err := RestoreStealingBundlesInteractor([]byte(`{"ph":`), newMockStealingBundlesPresenter())
	assert.Error(t, err)

	// 席数と設定が食い違う保存データ。
	_, err = RestoreStealingBundlesInteractor([]byte(`{"cf":{"p":4},"pl":[]}`), newMockStealingBundlesPresenter())
	assert.Error(t, err)
}
