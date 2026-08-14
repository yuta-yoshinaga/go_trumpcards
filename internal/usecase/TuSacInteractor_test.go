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

func newTuSacInteractorForTest() (*interfaces.MockTuSacGame,
	*presenter.MockTuSacPresenter, *TuSacInteractor,
) {
	mg := new(interfaces.MockTuSacGame)
	mp := new(presenter.MockTuSacPresenter)
	return mg, mp, NewTuSacInteractor(mg, mp)
}

func TestNewTuSacInteractor_NilPanics(t *testing.T) {
	mp := new(presenter.MockTuSacPresenter)
	assert.Panics(t, func() { NewTuSacInteractor(nil, mp) })

	mg := new(interfaces.MockTuSacGame)
	assert.Panics(t, func() { NewTuSacInteractor(mg, nil) })
}

func TestTuSacInteractor_Reset(t *testing.T) {
	mg, mp, ci := newTuSacInteractorForTest()
	mg.On("Reset").Return()
	mp.On("Output", mg, nil).Return("reset output")

	assert.Equal(t, "reset output", ci.Reset())
	mg.AssertCalled(t, "Reset")
}

// **山と捨て札のどちらから引くかはそのまま通す。**
func TestTuSacInteractor_PassesTheDrawSourceThrough(t *testing.T) {
	for _, fromDiscard := range []bool{false, true} {
		mg, mp, ci := newTuSacInteractorForTest()
		mg.On("GetGameEndFlag").Return(false)
		mg.On("Draw", fromDiscard).Return(nil)
		mp.On("Output", mg, nil).Return("ok")

		assert.Equal(t, "ok", ci.Draw(fromDiscard))
		mg.AssertCalled(t, "Draw", fromDiscard)
	}
}

// **出す組み合わせは選ばせる。** 勝手に全部出さない。
func TestTuSacInteractor_PassesTheChosenMeldThrough(t *testing.T) {
	mg, mp, ci := newTuSacInteractorForTest()
	idx := []int{1, 4, 7}
	mg.On("GetGameEndFlag").Return(false)
	mg.On("Meld", idx).Return(nil)
	mp.On("Output", mg, nil).Return("ok")

	assert.Equal(t, "ok", ci.Meld(idx))
	mg.AssertCalled(t, "Meld", idx)
}

// **組み合わせでない札はドメインが弾く。** usecase では判定しない。
func TestTuSacInteractor_LetsTheDomainRejectABadMeld(t *testing.T) {
	mg, mp, ci := newTuSacInteractorForTest()
	boom := errors.New("tusac: those cards are not a valid combination")
	idx := []int{0, 1, 2}
	mg.On("GetGameEndFlag").Return(false)
	mg.On("Meld", idx).Return(boom)
	mp.On("Output", mg, boom).Return("error output")

	assert.Equal(t, "error output", ci.Meld(idx))
	mg.AssertCalled(t, "Meld", idx)
}

func TestTuSacInteractor_DiscardAndNextRound(t *testing.T) {
	mg, mp, ci := newTuSacInteractorForTest()
	mg.On("GetGameEndFlag").Return(false)
	mg.On("Discard", 3).Return(nil)
	mg.On("NextRound").Return(nil)
	mp.On("Output", mg, nil).Return("ok")

	assert.Equal(t, "ok", ci.Discard(3))
	assert.Equal(t, "ok", ci.NextRound())
	mg.AssertCalled(t, "Discard", 3)
	mg.AssertCalled(t, "NextRound")
}

func TestTuSacInteractor_BlocksAfterGameEnd(t *testing.T) {
	mg, mp, ci := newTuSacInteractorForTest()
	mg.On("GetGameEndFlag").Return(true)
	mp.On("Output", mg, mock.Anything).Return("finished")

	assert.Equal(t, "finished", ci.Draw(false))
	assert.Equal(t, "finished", ci.Meld([]int{0, 1, 2}))
	assert.Equal(t, "finished", ci.Discard(0))
	assert.Equal(t, "finished", ci.NextRound())
	mg.AssertNotCalled(t, "Draw", false)
	mg.AssertNotCalled(t, "Meld", []int{0, 1, 2})
	mg.AssertNotCalled(t, "Discard", 0)
	mg.AssertNotCalled(t, "NextRound")
}

func TestTuSacInteractor_HintAndLog(t *testing.T) {
	mg, mp, ci := newTuSacInteractorForTest()
	mp.On("HintOutput", mg).Return("hint")
	mp.On("ActionLogOutput", mg).Return("log")

	assert.Equal(t, "hint", ci.Hint())
	assert.Equal(t, "log", ci.ActionLog())
}

func TestTuSacInteractor_Config(t *testing.T) {
	mg, mp, ci := newTuSacInteractorForTest()
	cfg := domain.DefaultTuSacConfig()
	mg.On("GetConfig").Return(cfg)
	assert.Equal(t, cfg, ci.GetConfig())

	bad := domain.TuSacConfig{Seats: 9, Rounds: 5}
	mp.On("Output", mg, mock.Anything).Return("bad config")
	assert.Equal(t, "bad config", ci.ResetWithConfig(bad))
	mg.AssertNotCalled(t, "SetConfig", bad)

	good := domain.TuSacConfig{Seats: 2, Rounds: 3}
	mg.On("SetConfig", good).Return()
	mg.On("Reset").Return()
	ci.ResetWithConfig(good)
	mg.AssertCalled(t, "SetConfig", good)
}

// --- 本物のドメインで駆動を確かめる ---

type tuSacSilentPresenter struct{}

func (p *tuSacSilentPresenter) Output(interfaces.TuSacGame, error) string   { return "" }
func (p *tuSacSilentPresenter) ActionLogOutput(interfaces.TuSacGame) string { return "" }
func (p *tuSacSilentPresenter) HintOutput(interfaces.TuSacGame) string      { return "" }

// **人間が捨てたら、盤面は人間の手番まで戻る。**
//
// テストが自分でループを回すと、駆動が抜けていても気づけない ── 人間の操作を
// 引いて捨てるの 2 回だけ実行して、どこで止まったかを見る。
func TestTuSacInteractor_OneTurnReturnsControl(t *testing.T) {
	for range 20 {
		g := domain.NewDefaultTuSac()
		ci := NewTuSacInteractor(g, new(tuSacSilentPresenter))
		ci.Reset()
		require.True(t, g.IsHumanTurn(), "配った直後に人間の手番でない")
		require.Equal(t, domain.TuSacPhaseDraw, g.GetPhase())

		ci.Draw(false)
		require.Equal(t, domain.TuSacPhaseDiscard, g.GetPhase())
		ci.Discard(0)

		if g.GetPhase() == domain.TuSacPhaseRoundEnd || g.GetGameEndFlag() {
			continue
		}
		require.True(t, g.IsHumanTurn(),
			"人間が 1 手番を終えたのに席 %d (CPU) で盤面が止まっている", g.GetTurnSeat())
		require.Equal(t, domain.TuSacPhaseDraw, g.GetPhase())
	}
}

func TestTuSacInteractor_SnapshotRestore(t *testing.T) {
	g := domain.NewDefaultTuSac()
	ci := NewTuSacInteractor(g, new(tuSacSilentPresenter))
	ci.Reset()
	ci.Draw(false)

	data, err := ci.Snapshot()
	require.NoError(t, err)
	assert.True(t, json.Valid(data))

	restored, err := RestoreTuSacInteractor(data, new(tuSacSilentPresenter))
	require.NoError(t, err)
	assert.Equal(t, g.GetPhase(), restored.Game.GetPhase())
	assert.Equal(t, g.GetStockCount(), restored.Game.GetStockCount())
	assert.Equal(t, g.GetRoundNumber(), restored.Game.GetRoundNumber())

	_, err = RestoreTuSacInteractor([]byte(`{"ph":9}`), new(tuSacSilentPresenter))
	assert.Error(t, err, "壊れた保存が復元できてしまった")
}
