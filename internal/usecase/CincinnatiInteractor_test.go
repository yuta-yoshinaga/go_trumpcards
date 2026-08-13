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

func newCincinnatiInteractorForTest() (*interfaces.MockCincinnatiGame,
	*presenter.MockCincinnatiPresenter, *CincinnatiInteractor,
) {
	mg := new(interfaces.MockCincinnatiGame)
	mp := new(presenter.MockCincinnatiPresenter)
	return mg, mp, NewCincinnatiInteractor(mg, mp)
}

func TestNewCincinnatiInteractor_NilPanics(t *testing.T) {
	mp := new(presenter.MockCincinnatiPresenter)
	assert.Panics(t, func() { NewCincinnatiInteractor(nil, mp) })

	mg := new(interfaces.MockCincinnatiGame)
	assert.Panics(t, func() { NewCincinnatiInteractor(mg, nil) })
}

func TestCincinnatiInteractor_Reset(t *testing.T) {
	mg, mp, ci := newCincinnatiInteractorForTest()
	mg.On("Reset").Return()
	mp.On("Output", mg, nil).Return("reset output")

	assert.Equal(t, "reset output", ci.Reset())
	mg.AssertCalled(t, "Reset")
}

// **手はそのままドメインへ渡る。** 合法性はここで判定しない。
func TestCincinnatiInteractor_ActionReachesTheDomain(t *testing.T) {
	mg, mp, ci := newCincinnatiInteractorForTest()
	mg.On("GetGameEndFlag").Return(false)
	mg.On("PlayerAction", domain.CincinnatiActionBet, 20).Return(nil)
	mg.On("CpuPlay").Return()
	mp.On("Output", mg, nil).Return("ok")

	assert.Equal(t, "ok", ci.Action(domain.CincinnatiActionBet, 20))
	mg.AssertCalled(t, "PlayerAction", domain.CincinnatiActionBet, 20)
}

// **人間が 1 手指したら CPU が動く。** 忘れると盤面が止まる。
func TestCincinnatiInteractor_DrivesTheCpu(t *testing.T) {
	for _, tt := range []struct {
		name   string
		setup  func(*interfaces.MockCincinnatiGame)
		invoke func(*CincinnatiInteractor) string
	}{
		{
			name: "Action",
			setup: func(m *interfaces.MockCincinnatiGame) {
				m.On("PlayerAction", domain.CincinnatiActionCheck, 0).Return(nil)
			},
			invoke: func(ci *CincinnatiInteractor) string {
				return ci.Action(domain.CincinnatiActionCheck, 0)
			},
		},
		{
			name:   "NextHand",
			setup:  func(m *interfaces.MockCincinnatiGame) { m.On("NextHand").Return(nil) },
			invoke: func(ci *CincinnatiInteractor) string { return ci.NextHand() },
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mg, mp, ci := newCincinnatiInteractorForTest()
			tt.setup(mg)
			mg.On("GetGameEndFlag").Return(false)
			mg.On("CpuPlay").Return()
			mp.On("Output", mg, nil).Return("ok")

			tt.invoke(ci)
			mg.AssertNumberOfCalls(t, "CpuPlay", 1)
		})
	}
}

// **拒まれた手では CPU を進めない。**
func TestCincinnatiInteractor_DoesNotDriveTheCpuOnError(t *testing.T) {
	mg, mp, ci := newCincinnatiInteractorForTest()
	boom := errors.New("nope")
	mg.On("GetGameEndFlag").Return(false)
	mg.On("PlayerAction", domain.CincinnatiActionCheck, 0).Return(boom)
	mg.On("CpuPlay").Return()
	mp.On("Output", mg, boom).Return("error output")

	assert.Equal(t, "error output", ci.Action(domain.CincinnatiActionCheck, 0))
	mg.AssertNumberOfCalls(t, "CpuPlay", 0)
}

func TestCincinnatiInteractor_BlocksAfterGameEnd(t *testing.T) {
	mg, mp, ci := newCincinnatiInteractorForTest()
	mg.On("GetGameEndFlag").Return(true)
	mp.On("Output", mg, mock.Anything).Return("finished")

	assert.Equal(t, "finished", ci.Action(domain.CincinnatiActionCheck, 0))
	assert.Equal(t, "finished", ci.NextHand())
	mg.AssertNotCalled(t, "PlayerAction", domain.CincinnatiActionCheck, 0)
	mg.AssertNotCalled(t, "NextHand")
	mg.AssertNotCalled(t, "CpuPlay")
}

func TestCincinnatiInteractor_HintAndLog(t *testing.T) {
	mg, mp, ci := newCincinnatiInteractorForTest()
	mp.On("HintOutput", mg).Return("hint")
	mp.On("ActionLogOutput", mg).Return("log")

	assert.Equal(t, "hint", ci.Hint())
	assert.Equal(t, "log", ci.ActionLog())
}

func TestCincinnatiInteractor_Config(t *testing.T) {
	mg, mp, ci := newCincinnatiInteractorForTest()
	cfg := domain.DefaultCincinnatiConfig()
	mg.On("GetConfig").Return(cfg)
	assert.Equal(t, cfg, ci.GetConfig())

	// **山が足りない席数は弾く。**
	bad := domain.CincinnatiConfig{Seats: 8, InitialChips: 1000, Ante: 10}
	mp.On("Output", mg, mock.Anything).Return("bad config")
	assert.Equal(t, "bad config", ci.ResetWithConfig(bad))
	mg.AssertNotCalled(t, "SetConfig", bad)

	good := domain.CincinnatiConfig{Seats: 3, InitialChips: 500, Ante: 5}
	mg.On("SetConfig", good).Return()
	mg.On("Reset").Return()
	ci.ResetWithConfig(good)
	mg.AssertCalled(t, "SetConfig", good)
}

// --- 本物のドメインで駆動を確かめる ---

type cincinnatiSilentPresenter struct{}

func (p *cincinnatiSilentPresenter) Output(interfaces.CincinnatiGame, error) string   { return "" }
func (p *cincinnatiSilentPresenter) ActionLogOutput(interfaces.CincinnatiGame) string { return "" }
func (p *cincinnatiSilentPresenter) HintOutput(interfaces.CincinnatiGame) string      { return "" }

// **人間の操作 1 回で盤面が人間の手番まで戻る。**
//
// テストが自分でループを回すと、誰も CPU を進めていなくても緑になる。
// 実物を 1 手だけ動かして、どこで止まったかを見る。
func TestCincinnatiInteractor_OneActionReturnsTheTurn(t *testing.T) {
	for range 30 {
		g := domain.NewDefaultCincinnati()
		g.Reset()
		ci := NewCincinnatiInteractor(g, new(cincinnatiSilentPresenter))
		require.True(t, g.IsHumanTurn(), "配った直後に人間の手番でない")

		ci.Action(domain.CincinnatiActionCheck, 0)
		if g.GetPhase() != domain.CincinnatiPhaseBetting {
			continue // そのハンドは決着した
		}
		require.True(t, g.IsHumanTurn(),
			"人間が 1 手指したのに席 %d (CPU) で盤面が止まっている", g.GetTurnSeat())
	}
}

func TestCincinnatiInteractor_SnapshotRestore(t *testing.T) {
	g := domain.NewDefaultCincinnati()
	g.Reset()
	ci := NewCincinnatiInteractor(g, new(cincinnatiSilentPresenter))
	ci.Action(domain.CincinnatiActionCheck, 0)

	data, err := ci.Snapshot()
	require.NoError(t, err)
	assert.True(t, json.Valid(data))

	restored, err := RestoreCincinnatiInteractor(data, new(cincinnatiSilentPresenter))
	require.NoError(t, err)
	assert.Equal(t, g.GetPhase(), restored.Game.GetPhase())
	assert.Equal(t, g.GetRevealedCount(), restored.Game.GetRevealedCount())
	assert.Equal(t, g.GetPot(), restored.Game.GetPot())

	_, err = RestoreCincinnatiInteractor([]byte(`{"ph":9}`), new(cincinnatiSilentPresenter))
	assert.Error(t, err, "壊れた保存が復元できてしまった")
}
