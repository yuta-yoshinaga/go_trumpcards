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

func newMockGoofspielGame() *interfaces.MockGoofspielGame { return new(interfaces.MockGoofspielGame) }

func newMockGoofspielPresenter() *presenter.MockGoofspielPresenter {
	return new(presenter.MockGoofspielPresenter)
}

func TestNewGoofspielInteractor(t *testing.T) {
	assert.NotNil(t, NewGoofspielInteractor(newMockGoofspielGame(), newMockGoofspielPresenter()))
}

func TestNewGoofspielInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewGoofspielInteractor(nil, newMockGoofspielPresenter()) })
	assert.Panics(t, func() { NewGoofspielInteractor(newMockGoofspielGame(), nil) })
}

func TestGoofspielInteractorReset(t *testing.T) {
	g := newMockGoofspielGame()
	p := newMockGoofspielPresenter()
	i := NewGoofspielInteractor(g, p)

	g.On("Reset").Return()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertNumberOfCalls(t, "Reset", 1)
}

// **CPU を先に動かしません。**
//
// 入札は同時なので、ここで `CpuPlay` を呼ぶと公開前に相手の札が出力に載ります。
// ドメインの `PlayerBid` が CPU の入札と一斉公開まで面倒を見ます。
func TestGoofspielInteractorBidDoesNotDriveTheCpuItself(t *testing.T) {
	g := newMockGoofspielGame()
	p := newMockGoofspielPresenter()
	i := NewGoofspielInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("PlayerBid", 3).Return(nil)
	p.On("Output", g, nil).Return("bid")

	assert.Equal(t, "bid", i.Bid(3))
	g.AssertNotCalled(t, "CpuPlay")
}

func TestGoofspielInteractorBidRejected(t *testing.T) {
	g := newMockGoofspielGame()
	p := newMockGoofspielPresenter()
	i := NewGoofspielInteractor(g, p)

	err := errors.New("すでに入札しています")
	g.On("GetGameEndFlag").Return(false)
	g.On("PlayerBid", 0).Return(err)
	p.On("Output", g, err).Return("bad")

	assert.Equal(t, "bad", i.Bid(0))
}

func TestGoofspielInteractorNextRound(t *testing.T) {
	t.Run("次をめくる", func(t *testing.T) {
		g := newMockGoofspielGame()
		p := newMockGoofspielPresenter()
		i := NewGoofspielInteractor(g, p)

		g.On("GetGameEndFlag").Return(false)
		g.On("NextRound").Return(nil)
		p.On("Output", g, nil).Return("dealt")

		assert.Equal(t, "dealt", i.NextRound())
	})

	t.Run("区切りでないときは弾く", func(t *testing.T) {
		g := newMockGoofspielGame()
		p := newMockGoofspielPresenter()
		i := NewGoofspielInteractor(g, p)

		err := errors.New("いまはラウンドの区切りではありません")
		g.On("GetGameEndFlag").Return(false)
		g.On("NextRound").Return(err)
		p.On("Output", g, err).Return("not now")

		assert.Equal(t, "not now", i.NextRound())
	})
}

func TestGoofspielInteractorMovesAfterTheEndAreRejected(t *testing.T) {
	g := newMockGoofspielGame()
	p := newMockGoofspielPresenter()
	i := NewGoofspielInteractor(g, p)

	g.On("GetGameEndFlag").Return(true)
	p.On("Output", g, mock.Anything).Return("ended")

	assert.Equal(t, "ended", i.Bid(0))
	assert.Equal(t, "ended", i.NextRound())
	assert.Equal(t, "ended", i.GiveUp())
	g.AssertNotCalled(t, "PlayerBid")
	g.AssertNotCalled(t, "NextRound")
	g.AssertNotCalled(t, "GiveUp")
}

func TestGoofspielInteractorGiveUp(t *testing.T) {
	g := newMockGoofspielGame()
	p := newMockGoofspielPresenter()
	i := NewGoofspielInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("gave up")

	assert.Equal(t, "gave up", i.GiveUp())
	g.AssertNumberOfCalls(t, "GiveUp", 1)
}

func TestGoofspielInteractorResetWithConfig(t *testing.T) {
	t.Run("妥当な設定は通る", func(t *testing.T) {
		g := newMockGoofspielGame()
		p := newMockGoofspielPresenter()
		i := NewGoofspielInteractor(g, p)

		cfg := domain.GoofspielConfig{PlayerCnt: 3, TieRule: domain.GoofspielTieCarryOver}
		g.On("SetConfig", cfg).Return()
		g.On("Reset").Return()
		p.On("Output", g, nil).Return("reset")

		assert.Equal(t, "reset", i.ResetWithConfig(cfg))
		g.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("不正な人数は弾かれ、盤面はそのまま", func(t *testing.T) {
		g := newMockGoofspielGame()
		p := newMockGoofspielPresenter()
		i := NewGoofspielInteractor(g, p)

		p.On("Output", g, mock.Anything).Return("bad config")
		assert.Equal(t, "bad config",
			i.ResetWithConfig(domain.GoofspielConfig{PlayerCnt: domain.GoofspielPlayerCntMax + 1}))
		g.AssertNotCalled(t, "SetConfig")
		g.AssertNotCalled(t, "Reset")
	})
}

func TestGoofspielInteractorGetConfigHintAndLog(t *testing.T) {
	g := newMockGoofspielGame()
	p := newMockGoofspielPresenter()
	i := NewGoofspielInteractor(g, p)

	cfg := domain.DefaultGoofspielConfig()
	g.On("GetConfig").Return(cfg)
	p.On("HintOutput", g).Return("hint")
	p.On("ActionLogOutput", g).Return("log")

	assert.Equal(t, cfg, i.GetConfig())
	assert.Equal(t, "hint", i.Hint())
	assert.Equal(t, "log", i.ActionLog())
}

// **KV に載らなければ Worker では毎リクエスト初期化される。**
func TestGoofspielInteractorSnapshotRoundTrip(t *testing.T) {
	g := domain.NewDefaultGoofspiel()
	g.Reset()
	i := NewGoofspielInteractor(g, newMockGoofspielPresenter())

	data, err := i.Snapshot()
	require.NoError(t, err)

	back, err := RestoreGoofspielInteractor(data, newMockGoofspielPresenter())
	require.NoError(t, err)
	assert.Equal(t, g.GetPlayerCnt(), back.Game.GetPlayerCnt())
	assert.Equal(t, g.GetRoundNumber(), back.Game.GetRoundNumber())
	assert.Equal(t, g.GetPrizeRemaining(), back.Game.GetPrizeRemaining())
}

func TestRestoreGoofspielInteractorRejectsGarbage(t *testing.T) {
	_, err := RestoreGoofspielInteractor([]byte(`{"ph":`), newMockGoofspielPresenter())
	assert.Error(t, err)

	// 席数と設定が食い違う保存データ。
	_, err = RestoreGoofspielInteractor([]byte(`{"cf":{"p":2,"tr":0},"pl":[]}`), newMockGoofspielPresenter())
	assert.Error(t, err)
}
