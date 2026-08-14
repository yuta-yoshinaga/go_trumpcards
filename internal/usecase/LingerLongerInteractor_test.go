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

func newMockLingerLongerGame() *interfaces.MockLingerLongerGame {
	return new(interfaces.MockLingerLongerGame)
}

func newMockLingerLongerPresenter() *presenter.MockLingerLongerPresenter {
	return new(presenter.MockLingerLongerPresenter)
}

func TestNewLingerLongerInteractor(t *testing.T) {
	assert.NotNil(t, NewLingerLongerInteractor(newMockLingerLongerGame(), newMockLingerLongerPresenter()))
}

func TestNewLingerLongerInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewLingerLongerInteractor(nil, newMockLingerLongerPresenter()) })
	assert.Panics(t, func() { NewLingerLongerInteractor(newMockLingerLongerGame(), nil) })
}

func TestLingerLongerInteractorResetAdvancesToTheHuman(t *testing.T) {
	g := newMockLingerLongerGame()
	p := newMockLingerLongerPresenter()
	i := NewLingerLongerInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(false).Times(3)
	g.On("IsHumanTurn").Return(true)
	g.On("CpuPlay").Return()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertNumberOfCalls(t, "CpuPlay", 3)
}

func TestLingerLongerInteractorStopsAtTheHumanTurn(t *testing.T) {
	g := newMockLingerLongerGame()
	p := newMockLingerLongerPresenter()
	i := NewLingerLongerInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNotCalled(t, "CpuPlay")
}

// **人間が脱落しても止まらない。**
//
// このゲームだけは、人間の手番が二度と来ないまま局が続きます。`IsHumanTurn` が
// 偽のままなので、終局フラグを見ていないと `maxCpuTurnsPerCall` まで空回りして
// 「決着したのに盤面が終わっていない」出力になります。
func TestLingerLongerInteractorRunsOnAfterTheHumanIsOut(t *testing.T) {
	g := newMockLingerLongerGame()
	p := newMockLingerLongerPresenter()
	i := NewLingerLongerInteractor(g, p)

	g.On("Reset").Return()
	g.On("IsHumanTurn").Return(false)
	g.On("GetGameEndFlag").Return(false).Times(5)
	g.On("GetGameEndFlag").Return(true)
	g.On("CpuPlay").Return()
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNumberOfCalls(t, "CpuPlay", 5)
}

func TestLingerLongerInteractorPlay(t *testing.T) {
	t.Run("打てたら CPU を進める", func(t *testing.T) {
		g := newMockLingerLongerGame()
		p := newMockLingerLongerPresenter()
		i := NewLingerLongerInteractor(g, p)

		g.On("GetGameEndFlag").Return(false)
		g.On("IsHumanTurn").Return(true).Once()
		g.On("PlayerPlay", 2).Return(nil)
		g.On("IsHumanTurn").Return(false).Once()
		g.On("CpuPlay").Return()
		g.On("IsHumanTurn").Return(true)
		p.On("Output", g, nil).Return("played")

		assert.Equal(t, "played", i.Play(2))
		g.AssertNumberOfCalls(t, "CpuPlay", 1)
	})

	t.Run("弾かれたら盤面は動かない", func(t *testing.T) {
		g := newMockLingerLongerGame()
		p := newMockLingerLongerPresenter()
		i := NewLingerLongerInteractor(g, p)

		err := errors.New("illegal")
		g.On("GetGameEndFlag").Return(false)
		g.On("IsHumanTurn").Return(true)
		g.On("PlayerPlay", 9).Return(err)
		p.On("Output", g, err).Return("bad")

		assert.Equal(t, "bad", i.Play(9))
		g.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("終局後は打てない", func(t *testing.T) {
		g := newMockLingerLongerGame()
		p := newMockLingerLongerPresenter()
		i := NewLingerLongerInteractor(g, p)

		g.On("GetGameEndFlag").Return(true)
		p.On("Output", g, mock.Anything).Return("ended")

		assert.Equal(t, "ended", i.Play(0))
		g.AssertNotCalled(t, "PlayerPlay")
	})
}

func TestLingerLongerInteractorGiveUp(t *testing.T) {
	g := newMockLingerLongerGame()
	p := newMockLingerLongerPresenter()
	i := NewLingerLongerInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("gave up")

	assert.Equal(t, "gave up", i.GiveUp())
	g.AssertNumberOfCalls(t, "GiveUp", 1)
}

func TestLingerLongerInteractorGiveUpAfterTheEndIsRejected(t *testing.T) {
	g := newMockLingerLongerGame()
	p := newMockLingerLongerPresenter()
	i := NewLingerLongerInteractor(g, p)

	g.On("GetGameEndFlag").Return(true)
	p.On("Output", g, mock.Anything).Return("ended")

	assert.Equal(t, "ended", i.GiveUp())
	g.AssertNotCalled(t, "GiveUp")
}

func TestLingerLongerInteractorResetWithConfig(t *testing.T) {
	t.Run("妥当な設定は通る", func(t *testing.T) {
		g := newMockLingerLongerGame()
		p := newMockLingerLongerPresenter()
		i := NewLingerLongerInteractor(g, p)

		cfg := domain.LingerLongerConfig{PlayerCnt: 5}
		g.On("SetConfig", cfg).Return()
		g.On("Reset").Return()
		g.On("GetGameEndFlag").Return(false)
		g.On("IsHumanTurn").Return(true)
		p.On("Output", g, nil).Return("reset")

		assert.Equal(t, "reset", i.ResetWithConfig(cfg))
		g.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("不正な人数は弾かれ、盤面はそのまま", func(t *testing.T) {
		g := newMockLingerLongerGame()
		p := newMockLingerLongerPresenter()
		i := NewLingerLongerInteractor(g, p)

		p.On("Output", g, mock.Anything).Return("bad config")
		assert.Equal(t, "bad config",
			i.ResetWithConfig(domain.LingerLongerConfig{PlayerCnt: domain.LingerLongerPlayerCntMax + 1}))
		g.AssertNotCalled(t, "SetConfig")
		g.AssertNotCalled(t, "Reset")
	})
}

func TestLingerLongerInteractorGetConfig(t *testing.T) {
	g := newMockLingerLongerGame()
	i := NewLingerLongerInteractor(g, newMockLingerLongerPresenter())
	cfg := domain.LingerLongerConfig{PlayerCnt: 6}
	g.On("GetConfig").Return(cfg)
	assert.Equal(t, cfg, i.GetConfig())
}

func TestLingerLongerInteractorHintAndActionLog(t *testing.T) {
	g := newMockLingerLongerGame()
	p := newMockLingerLongerPresenter()
	i := NewLingerLongerInteractor(g, p)

	p.On("HintOutput", g).Return("hint")
	p.On("ActionLogOutput", g).Return("log")

	assert.Equal(t, "hint", i.Hint())
	assert.Equal(t, "log", i.ActionLog())
}

// **KV に載らなければ Worker では毎リクエスト初期化される。**
func TestLingerLongerInteractorSnapshotRoundTrip(t *testing.T) {
	g := domain.NewDefaultLingerLonger()
	g.Reset()
	i := NewLingerLongerInteractor(g, newMockLingerLongerPresenter())

	data, err := i.Snapshot()
	require.NoError(t, err)

	back, err := RestoreLingerLongerInteractor(data, newMockLingerLongerPresenter())
	require.NoError(t, err)
	assert.Equal(t, g.GetStockSize(), back.Game.GetStockSize())
	assert.Equal(t, g.GetPlayerCnt(), back.Game.GetPlayerCnt())
	assert.Equal(t, g.GetPlayer(0).GetCardsSize(), back.Game.GetPlayer(0).GetCardsSize())
}

func TestRestoreLingerLongerInteractorRejectsGarbage(t *testing.T) {
	_, err := RestoreLingerLongerInteractor([]byte(`{"phase":`), newMockLingerLongerPresenter())
	assert.Error(t, err)

	// **壊れた状態は復元しない。** 席数と手札の数が食い違う保存データ。
	bad, err := json.Marshal(map[string]any{"config": map[string]any{"playerCnt": 4}, "players": []any{}})
	require.NoError(t, err)
	_, err = RestoreLingerLongerInteractor(bad, newMockLingerLongerPresenter())
	assert.Error(t, err)
}
