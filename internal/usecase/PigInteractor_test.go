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

func newMockPigGame() *interfaces.MockPigGame { return new(interfaces.MockPigGame) }

func newMockPigPresenter() *presenter.MockPigPresenter { return new(presenter.MockPigPresenter) }

func TestNewPigInteractor(t *testing.T) {
	assert.NotNil(t, NewPigInteractor(newMockPigGame(), newMockPigPresenter()))
}

func TestNewPigInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewPigInteractor(nil, newMockPigPresenter()) })
	assert.Panics(t, func() { NewPigInteractor(newMockPigGame(), nil) })
}

func TestPigInteractorResetAdvancesToTheHuman(t *testing.T) {
	g := newMockPigGame()
	p := newMockPigPresenter()
	i := NewPigInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(false).Times(3)
	g.On("IsHumanTurn").Return(true)
	g.On("CpuPlay").Return()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertNumberOfCalls(t, "CpuPlay", 3)
}

// **合図の場面で回し続けると人間が黙って負けます。**
//
// 遅れた 1 人が文字を受け取る規則なので、`IsHumanTurn` が真のあいだに CPU を
// 進めてしまうと、名乗る機会のないまま罰が確定します。
func TestPigInteractorStopsWhileTheHumanCanStillSignal(t *testing.T) {
	g := newMockPigGame()
	p := newMockPigPresenter()
	i := NewPigInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNotCalled(t, "CpuPlay")
}

// **人間が脱落しても止まりません。**
func TestPigInteractorRunsOnAfterTheHumanIsOut(t *testing.T) {
	g := newMockPigGame()
	p := newMockPigPresenter()
	i := NewPigInteractor(g, p)

	g.On("Reset").Return()
	g.On("IsHumanTurn").Return(false)
	g.On("GetGameEndFlag").Return(false).Times(4)
	g.On("GetGameEndFlag").Return(true)
	g.On("CpuPlay").Return()
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNumberOfCalls(t, "CpuPlay", 4)
}

func TestPigInteractorPass(t *testing.T) {
	t.Run("渡せたら CPU を進める", func(t *testing.T) {
		g := newMockPigGame()
		p := newMockPigPresenter()
		i := NewPigInteractor(g, p)

		g.On("GetGameEndFlag").Return(false)
		g.On("PlayerPass", 2).Return(nil)
		g.On("IsHumanTurn").Return(false).Once()
		g.On("CpuPlay").Return()
		g.On("IsHumanTurn").Return(true)
		p.On("Output", g, nil).Return("passed")

		assert.Equal(t, "passed", i.Pass(2))
		g.AssertNumberOfCalls(t, "CpuPlay", 1)
	})

	t.Run("弾かれたら盤面は動かない", func(t *testing.T) {
		g := newMockPigGame()
		p := newMockPigPresenter()
		i := NewPigInteractor(g, p)

		err := errors.New("illegal")
		g.On("GetGameEndFlag").Return(false)
		g.On("PlayerPass", 9).Return(err)
		p.On("Output", g, err).Return("bad")

		assert.Equal(t, "bad", i.Pass(9))
		g.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("終局後は渡せない", func(t *testing.T) {
		g := newMockPigGame()
		p := newMockPigPresenter()
		i := NewPigInteractor(g, p)

		g.On("GetGameEndFlag").Return(true)
		p.On("Output", g, mock.Anything).Return("ended")

		assert.Equal(t, "ended", i.Pass(0))
		g.AssertNotCalled(t, "PlayerPass")
	})
}

func TestPigInteractorSignal(t *testing.T) {
	t.Run("名乗れたら CPU を進める", func(t *testing.T) {
		g := newMockPigGame()
		p := newMockPigPresenter()
		i := NewPigInteractor(g, p)

		g.On("GetGameEndFlag").Return(false)
		g.On("PlayerSignal").Return(nil)
		g.On("IsHumanTurn").Return(true)
		p.On("Output", g, nil).Return("signalled")

		assert.Equal(t, "signalled", i.Signal())
	})

	// **合図が出ていないのに名乗るのは反則。** 早とちりは盤面を動かしません。
	t.Run("早とちりは弾く", func(t *testing.T) {
		g := newMockPigGame()
		p := newMockPigPresenter()
		i := NewPigInteractor(g, p)

		err := errors.New("まだ誰も合図していません")
		g.On("GetGameEndFlag").Return(false)
		g.On("PlayerSignal").Return(err)
		p.On("Output", g, err).Return("too early")

		assert.Equal(t, "too early", i.Signal())
		g.AssertNotCalled(t, "CpuPlay")
	})
}

func TestPigInteractorNextRound(t *testing.T) {
	t.Run("次を配って CPU を進める", func(t *testing.T) {
		g := newMockPigGame()
		p := newMockPigPresenter()
		i := NewPigInteractor(g, p)

		g.On("GetGameEndFlag").Return(false)
		g.On("NextRound").Return(nil)
		g.On("IsHumanTurn").Return(true)
		p.On("Output", g, nil).Return("dealt")

		assert.Equal(t, "dealt", i.NextRound())
	})

	t.Run("区切りでないときは弾く", func(t *testing.T) {
		g := newMockPigGame()
		p := newMockPigPresenter()
		i := NewPigInteractor(g, p)

		err := errors.New("いまはラウンドの区切りではありません")
		g.On("GetGameEndFlag").Return(false)
		g.On("NextRound").Return(err)
		p.On("Output", g, err).Return("not now")

		assert.Equal(t, "not now", i.NextRound())
	})
}

func TestPigInteractorGiveUp(t *testing.T) {
	g := newMockPigGame()
	p := newMockPigPresenter()
	i := NewPigInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("gave up")

	assert.Equal(t, "gave up", i.GiveUp())
	g.AssertNumberOfCalls(t, "GiveUp", 1)
}

func TestPigInteractorResetWithConfig(t *testing.T) {
	t.Run("妥当な設定は通る", func(t *testing.T) {
		g := newMockPigGame()
		p := newMockPigPresenter()
		i := NewPigInteractor(g, p)

		cfg := domain.PigConfig{PlayerCnt: 5, CpuDifficulty: domain.PigCpuHard}
		g.On("SetConfig", cfg).Return()
		g.On("Reset").Return()
		g.On("GetGameEndFlag").Return(false)
		g.On("IsHumanTurn").Return(true)
		p.On("Output", g, nil).Return("reset")

		assert.Equal(t, "reset", i.ResetWithConfig(cfg))
		g.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("不正な人数は弾かれ、盤面はそのまま", func(t *testing.T) {
		g := newMockPigGame()
		p := newMockPigPresenter()
		i := NewPigInteractor(g, p)

		p.On("Output", g, mock.Anything).Return("bad config")
		assert.Equal(t, "bad config",
			i.ResetWithConfig(domain.PigConfig{PlayerCnt: domain.PigPlayerCntMax + 1}))
		g.AssertNotCalled(t, "SetConfig")
		g.AssertNotCalled(t, "Reset")
	})
}

func TestPigInteractorGetConfigHintAndLog(t *testing.T) {
	g := newMockPigGame()
	p := newMockPigPresenter()
	i := NewPigInteractor(g, p)

	cfg := domain.DefaultPigConfig()
	g.On("GetConfig").Return(cfg)
	p.On("HintOutput", g).Return("hint")
	p.On("ActionLogOutput", g).Return("log")

	assert.Equal(t, cfg, i.GetConfig())
	assert.Equal(t, "hint", i.Hint())
	assert.Equal(t, "log", i.ActionLog())
}

// **KV に載らなければ Worker では毎リクエスト初期化される。**
func TestPigInteractorSnapshotRoundTrip(t *testing.T) {
	g := domain.NewDefaultPig()
	g.Reset()
	i := NewPigInteractor(g, newMockPigPresenter())

	data, err := i.Snapshot()
	require.NoError(t, err)

	back, err := RestorePigInteractor(data, newMockPigPresenter())
	require.NoError(t, err)
	assert.Equal(t, g.GetPlayerCnt(), back.Game.GetPlayerCnt())
	assert.Equal(t, g.GetRoundNumber(), back.Game.GetRoundNumber())
	assert.Equal(t, g.GetPlayer(0).GetCardsSize(), back.Game.GetPlayer(0).GetCardsSize())
}

func TestRestorePigInteractorRejectsGarbage(t *testing.T) {
	_, err := RestorePigInteractor([]byte(`{"ph":`), newMockPigPresenter())
	assert.Error(t, err)

	// 席数と設定が食い違う保存データ。
	_, err = RestorePigInteractor([]byte(`{"cf":{"p":4,"cd":1},"pl":[]}`), newMockPigPresenter())
	assert.Error(t, err)
}
