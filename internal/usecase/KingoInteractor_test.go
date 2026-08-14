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

func newKingoInteractorForTest() (*interfaces.MockKingoGame,
	*presenter.MockKingoPresenter, *KingoInteractor,
) {
	mg := new(interfaces.MockKingoGame)
	mp := new(presenter.MockKingoPresenter)
	return mg, mp, NewKingoInteractor(mg, mp)
}

func TestNewKingoInteractor_NilPanics(t *testing.T) {
	mp := new(presenter.MockKingoPresenter)
	assert.Panics(t, func() { NewKingoInteractor(nil, mp) })

	mg := new(interfaces.MockKingoGame)
	assert.Panics(t, func() { NewKingoInteractor(mg, nil) })
}

func TestKingoInteractor_Reset(t *testing.T) {
	mg, mp, ci := newKingoInteractorForTest()
	mg.On("Reset").Return()
	mp.On("Output", mg, nil).Return("reset output")

	assert.Equal(t, "reset output", ci.Reset())
	mg.AssertCalled(t, "Reset")
}

// **額はそのまま通す。** 手札は配る前なので、丸めると見えていない情報を
// 根拠にしたことになる。
func TestKingoInteractor_PassesTheBetThrough(t *testing.T) {
	for _, amount := range []int{10, 50, 500} {
		mg, mp, ci := newKingoInteractorForTest()
		mg.On("GetGameEndFlag").Return(false)
		mg.On("PlaceBet", amount).Return(nil)
		mp.On("Output", mg, nil).Return("ok")

		assert.Equal(t, "ok", ci.Bet(amount))
		mg.AssertCalled(t, "PlaceBet", amount)
	}
}

// **範囲外はドメインが弾く。** usecase では判定しない。
func TestKingoInteractor_LetsTheDomainRejectABadBet(t *testing.T) {
	mg, mp, ci := newKingoInteractorForTest()
	boom := errors.New("kingo: bet out of range")
	mg.On("GetGameEndFlag").Return(false)
	mg.On("PlaceBet", 7).Return(boom)
	mp.On("Output", mg, boom).Return("error output")

	assert.Equal(t, "error output", ci.Bet(7))
	mg.AssertCalled(t, "PlaceBet", 7)
}

func TestKingoInteractor_DealAndNextRound(t *testing.T) {
	mg, mp, ci := newKingoInteractorForTest()
	mg.On("GetGameEndFlag").Return(false)
	mg.On("Deal").Return(nil)
	mg.On("NextRound").Return(nil)
	mp.On("Output", mg, nil).Return("ok")

	assert.Equal(t, "ok", ci.Deal())
	assert.Equal(t, "ok", ci.NextRound())
	mg.AssertCalled(t, "Deal")
	mg.AssertCalled(t, "NextRound")
}

func TestKingoInteractor_BlocksAfterGameEnd(t *testing.T) {
	mg, mp, ci := newKingoInteractorForTest()
	mg.On("GetGameEndFlag").Return(true)
	mp.On("Output", mg, mock.Anything).Return("finished")

	assert.Equal(t, "finished", ci.Bet(10))
	assert.Equal(t, "finished", ci.Deal())
	assert.Equal(t, "finished", ci.NextRound())
	mg.AssertNotCalled(t, "PlaceBet", 10)
	mg.AssertNotCalled(t, "Deal")
	mg.AssertNotCalled(t, "NextRound")
}

func TestKingoInteractor_HintAndLog(t *testing.T) {
	mg, mp, ci := newKingoInteractorForTest()
	mp.On("HintOutput", mg).Return("hint")
	mp.On("ActionLogOutput", mg).Return("log")

	assert.Equal(t, "hint", ci.Hint())
	assert.Equal(t, "log", ci.ActionLog())
}

func TestKingoInteractor_Config(t *testing.T) {
	mg, mp, ci := newKingoInteractorForTest()
	cfg := domain.DefaultKingoConfig()
	mg.On("GetConfig").Return(cfg)
	assert.Equal(t, cfg, ci.GetConfig())

	// **親が回らない設定は弾く。** ラウンド数 < 席数。
	bad := domain.KingoConfig{Seats: 5, InitialChips: 1000, MinBet: 10, Rounds: 4}
	mp.On("Output", mg, mock.Anything).Return("bad config")
	assert.Equal(t, "bad config", ci.ResetWithConfig(bad))
	mg.AssertNotCalled(t, "SetConfig", bad)

	good := domain.KingoConfig{Seats: 3, InitialChips: 500, MinBet: 20, Rounds: 5}
	mg.On("SetConfig", good).Return()
	mg.On("Reset").Return()
	ci.ResetWithConfig(good)
	mg.AssertCalled(t, "SetConfig", good)
}

// --- 本物のドメインで駆動を確かめる ---

type kingoSilentPresenter struct{}

func (p *kingoSilentPresenter) Output(interfaces.KingoGame, error) string   { return "" }
func (p *kingoSilentPresenter) ActionLogOutput(interfaces.KingoGame) string { return "" }
func (p *kingoSilentPresenter) HintOutput(interfaces.KingoGame) string      { return "" }

// **人間の操作 1 回でラウンドが閉じる。**
//
// テストが自分でループを回すと、駆動が抜けていても気づけない ── 人間の操作を
// 1 回だけ実行して、どこで止まったかを見る。
func TestKingoInteractor_OneActionClosesTheRound(t *testing.T) {
	for range 20 {
		g := domain.NewDefaultKingo()
		ci := NewKingoInteractor(g, new(kingoSilentPresenter))
		ci.Reset()
		require.True(t, g.IsHumanTurn(), "配った直後に人間の入力待ちでない")

		if g.IsHumanBanker() {
			ci.Deal()
		} else {
			ci.Bet(g.GetConfig().MinBet)
		}
		require.Equal(t, domain.KingoPhaseResult, g.GetPhase(),
			"人間が 1 手指したのにラウンドが閉じていない")
	}
}

func TestKingoInteractor_SnapshotRestore(t *testing.T) {
	g := domain.NewDefaultKingo()
	ci := NewKingoInteractor(g, new(kingoSilentPresenter))
	ci.Reset()

	data, err := ci.Snapshot()
	require.NoError(t, err)
	assert.True(t, json.Valid(data))

	restored, err := RestoreKingoInteractor(data, new(kingoSilentPresenter))
	require.NoError(t, err)
	assert.Equal(t, g.GetPhase(), restored.Game.GetPhase())
	assert.Equal(t, g.GetBankerSeat(), restored.Game.GetBankerSeat())
	assert.Equal(t, g.GetRoundNumber(), restored.Game.GetRoundNumber())

	_, err = RestoreKingoInteractor([]byte(`{"ph":9}`), new(kingoSilentPresenter))
	assert.Error(t, err, "壊れた保存が復元できてしまった")
}
