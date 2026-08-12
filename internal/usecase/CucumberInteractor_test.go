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

func newMockCucumberGame() *interfaces.MockCucumberGame { return new(interfaces.MockCucumberGame) }

func newMockCucumberPresenter() *presenter.MockCucumberPresenter {
	return new(presenter.MockCucumberPresenter)
}

func TestNewCucumberInteractor(t *testing.T) {
	assert.NotNil(t, NewCucumberInteractor(newMockCucumberGame(), newMockCucumberPresenter()))
}

func TestNewCucumberInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewCucumberInteractor(nil, newMockCucumberPresenter()) })
	assert.Panics(t, func() { NewCucumberInteractor(newMockCucumberGame(), nil) })
}

func TestCucumberInteractorResetAdvancesToTheHuman(t *testing.T) {
	g := newMockCucumberGame()
	p := newMockCucumberPresenter()
	i := NewCucumberInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(false).Times(3)
	g.On("IsHumanTurn").Return(true)
	g.On("CpuPlay").Return()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertNumberOfCalls(t, "CpuPlay", 3)
}

// **ラウンドの区切りでも止まります。**
//
// 失点はラウンドに 1 回だけの出来事で盤面に痕跡が残らないので、読む前に配り
// 直してはいけません。`IsHumanTurn` がラウンド終了で真を返すのはそのためです。
func TestCucumberInteractorStopsAtTheRoundBoundary(t *testing.T) {
	g := newMockCucumberGame()
	p := newMockCucumberPresenter()
	i := NewCucumberInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNotCalled(t, "CpuPlay")
}

func TestCucumberInteractorPlay(t *testing.T) {
	t.Run("出せたら CPU を進める", func(t *testing.T) {
		g := newMockCucumberGame()
		p := newMockCucumberPresenter()
		i := NewCucumberInteractor(g, p)

		g.On("GetGameEndFlag").Return(false)
		g.On("PlayerPlay", 2).Return(nil)
		g.On("IsHumanTurn").Return(false).Once()
		g.On("CpuPlay").Return()
		g.On("IsHumanTurn").Return(true)
		p.On("Output", g, nil).Return("played")

		assert.Equal(t, "played", i.Play(2))
		g.AssertNumberOfCalls(t, "CpuPlay", 1)
	})

	t.Run("弾かれたら盤面は動かない", func(t *testing.T) {
		g := newMockCucumberGame()
		p := newMockCucumberPresenter()
		i := NewCucumberInteractor(g, p)

		err := errors.New("いま出ている最高ランクを超える札を出してください")
		g.On("GetGameEndFlag").Return(false)
		g.On("PlayerPlay", 0).Return(err)
		p.On("Output", g, err).Return("bad")

		assert.Equal(t, "bad", i.Play(0))
		g.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("終局後は出せない", func(t *testing.T) {
		g := newMockCucumberGame()
		p := newMockCucumberPresenter()
		i := NewCucumberInteractor(g, p)

		g.On("GetGameEndFlag").Return(true)
		p.On("Output", g, mock.Anything).Return("ended")

		assert.Equal(t, "ended", i.Play(0))
		g.AssertNotCalled(t, "PlayerPlay")
	})
}

func TestCucumberInteractorNextRound(t *testing.T) {
	t.Run("次を配って CPU を進める", func(t *testing.T) {
		g := newMockCucumberGame()
		p := newMockCucumberPresenter()
		i := NewCucumberInteractor(g, p)

		g.On("GetGameEndFlag").Return(false)
		g.On("NextRound").Return(nil)
		g.On("IsHumanTurn").Return(true)
		p.On("Output", g, nil).Return("dealt")

		assert.Equal(t, "dealt", i.NextRound())
	})

	t.Run("区切りでないときは弾く", func(t *testing.T) {
		g := newMockCucumberGame()
		p := newMockCucumberPresenter()
		i := NewCucumberInteractor(g, p)

		err := errors.New("いまはラウンドの区切りではありません")
		g.On("GetGameEndFlag").Return(false)
		g.On("NextRound").Return(err)
		p.On("Output", g, err).Return("not now")

		assert.Equal(t, "not now", i.NextRound())
	})
}

func TestCucumberInteractorGiveUp(t *testing.T) {
	g := newMockCucumberGame()
	p := newMockCucumberPresenter()
	i := NewCucumberInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("gave up")

	assert.Equal(t, "gave up", i.GiveUp())
	g.AssertNumberOfCalls(t, "GiveUp", 1)
}

func TestCucumberInteractorResetWithConfig(t *testing.T) {
	t.Run("妥当な設定は通る", func(t *testing.T) {
		g := newMockCucumberGame()
		p := newMockCucumberPresenter()
		i := NewCucumberInteractor(g, p)

		cfg := domain.CucumberConfig{PlayerCnt: 5, TargetScore: 50}
		g.On("SetConfig", cfg).Return()
		g.On("Reset").Return()
		g.On("GetGameEndFlag").Return(false)
		g.On("IsHumanTurn").Return(true)
		p.On("Output", g, nil).Return("reset")

		assert.Equal(t, "reset", i.ResetWithConfig(cfg))
		g.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("不正な設定は弾かれ、盤面はそのまま", func(t *testing.T) {
		g := newMockCucumberGame()
		p := newMockCucumberPresenter()
		i := NewCucumberInteractor(g, p)

		p.On("Output", g, mock.Anything).Return("bad config")
		assert.Equal(t, "bad config",
			i.ResetWithConfig(domain.CucumberConfig{PlayerCnt: domain.CucumberPlayerCntMax + 1, TargetScore: 30}))
		g.AssertNotCalled(t, "SetConfig")
		g.AssertNotCalled(t, "Reset")
	})
}

func TestCucumberInteractorGetConfigHintAndLog(t *testing.T) {
	g := newMockCucumberGame()
	p := newMockCucumberPresenter()
	i := NewCucumberInteractor(g, p)

	cfg := domain.DefaultCucumberConfig()
	g.On("GetConfig").Return(cfg)
	p.On("HintOutput", g).Return("hint")
	p.On("ActionLogOutput", g).Return("log")

	assert.Equal(t, cfg, i.GetConfig())
	assert.Equal(t, "hint", i.Hint())
	assert.Equal(t, "log", i.ActionLog())
}

// **KV に載らなければ Worker では毎リクエスト初期化される。**
func TestCucumberInteractorSnapshotRoundTrip(t *testing.T) {
	g := domain.NewDefaultCucumber()
	g.Reset()
	i := NewCucumberInteractor(g, newMockCucumberPresenter())

	data, err := i.Snapshot()
	require.NoError(t, err)

	back, err := RestoreCucumberInteractor(data, newMockCucumberPresenter())
	require.NoError(t, err)
	assert.Equal(t, g.GetPlayerCnt(), back.Game.GetPlayerCnt())
	assert.Equal(t, g.GetRoundNumber(), back.Game.GetRoundNumber())
	assert.Equal(t, g.GetPlayer(0).GetCardsSize(), back.Game.GetPlayer(0).GetCardsSize())
}

func TestRestoreCucumberInteractorRejectsGarbage(t *testing.T) {
	_, err := RestoreCucumberInteractor([]byte(`{"ph":`), newMockCucumberPresenter())
	assert.Error(t, err)

	// 席数と設定が食い違う保存データ。
	_, err = RestoreCucumberInteractor([]byte(`{"cf":{"p":4,"ts":30},"pl":[]}`), newMockCucumberPresenter())
	assert.Error(t, err)
}
