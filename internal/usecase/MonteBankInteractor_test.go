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

func newMonteBankInteractorForTest() (*interfaces.MockMonteBankGame,
	*presenter.MockMonteBankPresenter, *MonteBankInteractor,
) {
	mg := new(interfaces.MockMonteBankGame)
	mp := new(presenter.MockMonteBankPresenter)
	return mg, mp, NewMonteBankInteractor(mg, mp)
}

func TestNewMonteBankInteractor_NilPanics(t *testing.T) {
	mp := new(presenter.MockMonteBankPresenter)
	assert.Panics(t, func() { NewMonteBankInteractor(nil, mp) })

	mg := new(interfaces.MockMonteBankGame)
	assert.Panics(t, func() { NewMonteBankInteractor(mg, nil) })
}

func TestMonteBankInteractor_Reset(t *testing.T) {
	mg, mp, ci := newMonteBankInteractorForTest()
	mg.On("Reset").Return()
	mp.On("Output", mg, nil).Return("reset output")

	assert.Equal(t, "reset output", ci.Reset())
	mg.AssertCalled(t, "Reset")
}

// **賭けはインデックスと額をそのままドメインへ渡す。**
func TestMonteBankInteractor_PlaceBetReachesTheDomain(t *testing.T) {
	mg, mp, ci := newMonteBankInteractorForTest()
	mg.On("GetGameEndFlag").Return(false)
	mg.On("PlaceBet", 2, 50).Return(nil)
	mp.On("Output", mg, nil).Return("ok")

	assert.Equal(t, "ok", ci.PlaceBet(2, 50))
	mg.AssertCalled(t, "PlaceBet", 2, 50)
}

// **損な賭けも拒まない。** 場札に同じスートが 2 枚出ていても、それを選ぶのは
// プレイヤーの自由 ── ここで弾くと「選べる手」が黙って消える。
func TestMonteBankInteractor_DoesNotBlockABadBet(t *testing.T) {
	mg, mp, ci := newMonteBankInteractorForTest()
	mg.On("GetGameEndFlag").Return(false)
	mg.On("PlaceBet", 0, 50).Return(nil)
	mg.On("SuitCountInLayout", mock.Anything).Return(4) // 最悪の賭け
	mp.On("Output", mg, nil).Return("ok")

	assert.Equal(t, "ok", ci.PlaceBet(0, 50))
	mg.AssertCalled(t, "PlaceBet", 0, 50)
}

func TestMonteBankInteractor_NextRound(t *testing.T) {
	mg, mp, ci := newMonteBankInteractorForTest()
	mg.On("GetGameEndFlag").Return(false)
	mg.On("NextRound").Return(nil)
	mp.On("Output", mg, nil).Return("ok")

	assert.Equal(t, "ok", ci.NextRound())
	mg.AssertCalled(t, "NextRound")
}

func TestMonteBankInteractor_PassesErrorsToThePresenter(t *testing.T) {
	mg, mp, ci := newMonteBankInteractorForTest()
	boom := errors.New("nope")
	mg.On("GetGameEndFlag").Return(false)
	mg.On("PlaceBet", 0, 50).Return(boom)
	mp.On("Output", mg, boom).Return("error output")

	assert.Equal(t, "error output", ci.PlaceBet(0, 50))
}

// **終局後はドメインに触らない。**
func TestMonteBankInteractor_BlocksAfterGameEnd(t *testing.T) {
	mg, mp, ci := newMonteBankInteractorForTest()
	mg.On("GetGameEndFlag").Return(true)
	mp.On("Output", mg, mock.Anything).Return("finished")

	assert.Equal(t, "finished", ci.PlaceBet(0, 50))
	assert.Equal(t, "finished", ci.NextRound())
	mg.AssertNotCalled(t, "PlaceBet", 0, 50)
	mg.AssertNotCalled(t, "NextRound")
}

func TestMonteBankInteractor_HintAndLog(t *testing.T) {
	mg, mp, ci := newMonteBankInteractorForTest()
	mp.On("HintOutput", mg).Return("hint")
	mp.On("ActionLogOutput", mg).Return("log")

	assert.Equal(t, "hint", ci.Hint())
	assert.Equal(t, "log", ci.ActionLog())
}

func TestMonteBankInteractor_Config(t *testing.T) {
	mg, mp, ci := newMonteBankInteractorForTest()
	cfg := domain.DefaultMonteBankConfig()
	mg.On("GetConfig").Return(cfg)
	assert.Equal(t, cfg, ci.GetConfig())

	// **範囲外の設定は弾く。**
	bad := domain.MonteBankConfig{InitialChips: 1, DefaultBet: 50}
	mp.On("Output", mg, mock.Anything).Return("bad config")
	assert.Equal(t, "bad config", ci.ResetWithConfig(bad))
	mg.AssertNotCalled(t, "SetConfig", bad)

	good := domain.MonteBankConfig{InitialChips: 500, DefaultBet: 20}
	mg.On("SetConfig", good).Return()
	mg.On("Reset").Return()
	ci.ResetWithConfig(good)
	mg.AssertCalled(t, "SetConfig", good)
}

// --- 本物のドメインで一巡させる ---

// monteBankSilentPresenter は進行だけを見るための無音プレゼンタ。
type monteBankSilentPresenter struct{}

func (p *monteBankSilentPresenter) Output(interfaces.MonteBankGame, error) string   { return "" }
func (p *monteBankSilentPresenter) ActionLogOutput(interfaces.MonteBankGame) string { return "" }
func (p *monteBankSilentPresenter) HintOutput(interfaces.MonteBankGame) string      { return "" }

// **山を使い切るまで回り切る。** 途中で詰まる経路が無いことを実物で見る。
func TestMonteBankInteractor_PlaysThroughTheDeck(t *testing.T) {
	for range 50 {
		g := domain.NewDefaultMonteBank()
		g.Reset()
		ci := NewMonteBankInteractor(g, new(monteBankSilentPresenter))

		rounds := 0
		for !g.GetGameEndFlag() {
			rounds++
			require.Less(t, rounds, 100, "局が終わらない")
			require.Len(t, g.GetLayout(), domain.MonteBankLayoutSize,
				"場札が %d 枚に満たないまま賭けさせている", domain.MonteBankLayoutSize)
			ci.PlaceBet(0, domain.MonteBankMinBet)
			ci.NextRound()
		}
		assert.Equal(t, domain.MonteBankDeckSize/domain.MonteBankCardsPerRound, rounds)
	}
}

func TestMonteBankInteractor_SnapshotRestore(t *testing.T) {
	g := domain.NewDefaultMonteBank()
	g.Reset()
	ci := NewMonteBankInteractor(g, new(monteBankSilentPresenter))
	ci.PlaceBet(0, domain.MonteBankDefaultBet)

	data, err := ci.Snapshot()
	require.NoError(t, err)
	assert.True(t, json.Valid(data))

	restored, err := RestoreMonteBankInteractor(data, new(monteBankSilentPresenter))
	require.NoError(t, err)
	assert.Equal(t, g.GetPhase(), restored.Game.GetPhase())
	assert.Equal(t, g.GetChips(), restored.Game.GetChips())
	assert.Equal(t, g.GetResult(), restored.Game.GetResult())

	_, err = RestoreMonteBankInteractor([]byte(`{"ph":9}`), new(monteBankSilentPresenter))
	assert.Error(t, err, "壊れた保存が復元できてしまった")
}
