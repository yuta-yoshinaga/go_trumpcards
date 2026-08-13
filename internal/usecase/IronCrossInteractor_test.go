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

func newIronCrossInteractorForTest() (*interfaces.MockIronCrossGame,
	*presenter.MockIronCrossPresenter, *IronCrossInteractor,
) {
	mg := new(interfaces.MockIronCrossGame)
	mp := new(presenter.MockIronCrossPresenter)
	return mg, mp, NewIronCrossInteractor(mg, mp)
}

func TestNewIronCrossInteractor_NilPanics(t *testing.T) {
	mp := new(presenter.MockIronCrossPresenter)
	assert.Panics(t, func() { NewIronCrossInteractor(nil, mp) })

	mg := new(interfaces.MockIronCrossGame)
	assert.Panics(t, func() { NewIronCrossInteractor(mg, nil) })
}

func TestIronCrossInteractor_Reset(t *testing.T) {
	mg, mp, ci := newIronCrossInteractorForTest()
	mg.On("Reset").Return()
	mp.On("Output", mg, nil).Return("reset output")

	assert.Equal(t, "reset output", ci.Reset())
	mg.AssertCalled(t, "Reset")
}

func TestIronCrossInteractor_ActionReachesTheDomain(t *testing.T) {
	mg, mp, ci := newIronCrossInteractorForTest()
	mg.On("GetGameEndFlag").Return(false)
	mg.On("PlayerAction", domain.IronCrossActionBet, 20).Return(nil)
	mg.On("CpuPlay").Return()
	mp.On("Output", mg, nil).Return("ok")

	assert.Equal(t, "ok", ci.Action(domain.IronCrossActionBet, 20))
	mg.AssertCalled(t, "PlayerAction", domain.IronCrossActionBet, 20)
}

// **弱いほうの列を選んでも直さない。** ここで「親切に」直すと、このゲームの
// 唯一の判断が消える。
func TestIronCrossInteractor_PassesTheChosenLineThrough(t *testing.T) {
	for _, line := range []domain.IronCrossLine{
		domain.IronCrossLineVertical, domain.IronCrossLineHorizontal,
	} {
		mg, mp, ci := newIronCrossInteractorForTest()
		mg.On("GetGameEndFlag").Return(false)
		mg.On("ChooseLine", line).Return(nil)
		mg.On("CpuPlay").Return()
		mp.On("Output", mg, nil).Return("ok")

		assert.Equal(t, "ok", ci.ChooseLine(int(line)))
		mg.AssertCalled(t, "ChooseLine", line)
	}
}

// **不正な列はドメインが弾く。** usecase では判定しない。
func TestIronCrossInteractor_LetsTheDomainRejectABadLine(t *testing.T) {
	mg, mp, ci := newIronCrossInteractorForTest()
	boom := errors.New("choose the vertical or the horizontal line")
	mg.On("GetGameEndFlag").Return(false)
	mg.On("ChooseLine", domain.IronCrossLine(99)).Return(boom)
	mg.On("CpuPlay").Return()
	mp.On("Output", mg, boom).Return("error output")

	assert.Equal(t, "error output", ci.ChooseLine(99))
	mg.AssertCalled(t, "ChooseLine", domain.IronCrossLine(99))
}

// **人間が 1 手指したら CPU が動く。**
func TestIronCrossInteractor_DrivesTheCpu(t *testing.T) {
	for _, tt := range []struct {
		name   string
		setup  func(*interfaces.MockIronCrossGame)
		invoke func(*IronCrossInteractor) string
	}{
		{
			name: "Action",
			setup: func(m *interfaces.MockIronCrossGame) {
				m.On("PlayerAction", domain.IronCrossActionCheck, 0).Return(nil)
			},
			invoke: func(ci *IronCrossInteractor) string {
				return ci.Action(domain.IronCrossActionCheck, 0)
			},
		},
		{
			name: "ChooseLine",
			setup: func(m *interfaces.MockIronCrossGame) {
				m.On("ChooseLine", domain.IronCrossLineVertical).Return(nil)
			},
			invoke: func(ci *IronCrossInteractor) string {
				return ci.ChooseLine(int(domain.IronCrossLineVertical))
			},
		},
		{
			name:   "NextHand",
			setup:  func(m *interfaces.MockIronCrossGame) { m.On("NextHand").Return(nil) },
			invoke: func(ci *IronCrossInteractor) string { return ci.NextHand() },
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mg, mp, ci := newIronCrossInteractorForTest()
			tt.setup(mg)
			mg.On("GetGameEndFlag").Return(false)
			mg.On("CpuPlay").Return()
			mp.On("Output", mg, nil).Return("ok")

			tt.invoke(ci)
			mg.AssertNumberOfCalls(t, "CpuPlay", 1)
		})
	}
}

func TestIronCrossInteractor_DoesNotDriveTheCpuOnError(t *testing.T) {
	mg, mp, ci := newIronCrossInteractorForTest()
	boom := errors.New("nope")
	mg.On("GetGameEndFlag").Return(false)
	mg.On("PlayerAction", domain.IronCrossActionCheck, 0).Return(boom)
	mg.On("CpuPlay").Return()
	mp.On("Output", mg, boom).Return("error output")

	assert.Equal(t, "error output", ci.Action(domain.IronCrossActionCheck, 0))
	mg.AssertNumberOfCalls(t, "CpuPlay", 0)
}

func TestIronCrossInteractor_BlocksAfterGameEnd(t *testing.T) {
	mg, mp, ci := newIronCrossInteractorForTest()
	mg.On("GetGameEndFlag").Return(true)
	mp.On("Output", mg, mock.Anything).Return("finished")

	assert.Equal(t, "finished", ci.Action(domain.IronCrossActionCheck, 0))
	assert.Equal(t, "finished", ci.ChooseLine(int(domain.IronCrossLineVertical)))
	assert.Equal(t, "finished", ci.NextHand())
	mg.AssertNotCalled(t, "PlayerAction", domain.IronCrossActionCheck, 0)
	mg.AssertNotCalled(t, "ChooseLine", domain.IronCrossLineVertical)
	mg.AssertNotCalled(t, "NextHand")
}

func TestIronCrossInteractor_HintAndLog(t *testing.T) {
	mg, mp, ci := newIronCrossInteractorForTest()
	mp.On("HintOutput", mg).Return("hint")
	mp.On("ActionLogOutput", mg).Return("log")

	assert.Equal(t, "hint", ci.Hint())
	assert.Equal(t, "log", ci.ActionLog())
}

func TestIronCrossInteractor_Config(t *testing.T) {
	mg, mp, ci := newIronCrossInteractorForTest()
	cfg := domain.DefaultIronCrossConfig()
	mg.On("GetConfig").Return(cfg)
	assert.Equal(t, cfg, ci.GetConfig())

	bad := domain.IronCrossConfig{Seats: 9, InitialChips: 1000, Ante: 10}
	mp.On("Output", mg, mock.Anything).Return("bad config")
	assert.Equal(t, "bad config", ci.ResetWithConfig(bad))
	mg.AssertNotCalled(t, "SetConfig", bad)

	good := domain.IronCrossConfig{Seats: 3, InitialChips: 500, Ante: 5}
	mg.On("SetConfig", good).Return()
	mg.On("Reset").Return()
	ci.ResetWithConfig(good)
	mg.AssertCalled(t, "SetConfig", good)
}

// --- 本物のドメインで駆動を確かめる ---

type ironCrossSilentPresenter struct{}

func (p *ironCrossSilentPresenter) Output(interfaces.IronCrossGame, error) string   { return "" }
func (p *ironCrossSilentPresenter) ActionLogOutput(interfaces.IronCrossGame) string { return "" }
func (p *ironCrossSilentPresenter) HintOutput(interfaces.IronCrossGame) string      { return "" }

// **人間の操作 1 回で盤面が人間の手番まで戻る。**
func TestIronCrossInteractor_OneActionReturnsTheTurn(t *testing.T) {
	for range 30 {
		g := domain.NewDefaultIronCross()
		g.Reset()
		ci := NewIronCrossInteractor(g, new(ironCrossSilentPresenter))
		require.True(t, g.IsHumanTurn(), "配った直後に人間の手番でない")

		ci.Action(domain.IronCrossActionCheck, 0)
		if g.GetPhase() != domain.IronCrossPhaseBetting {
			continue
		}
		require.True(t, g.IsHumanTurn(),
			"人間が 1 手指したのに席 %d (CPU) で盤面が止まっている", g.GetTurnSeat())
	}
}

func TestIronCrossInteractor_SnapshotRestore(t *testing.T) {
	g := domain.NewDefaultIronCross()
	g.Reset()
	ci := NewIronCrossInteractor(g, new(ironCrossSilentPresenter))
	ci.Action(domain.IronCrossActionCheck, 0)

	data, err := ci.Snapshot()
	require.NoError(t, err)
	assert.True(t, json.Valid(data))

	restored, err := RestoreIronCrossInteractor(data, new(ironCrossSilentPresenter))
	require.NoError(t, err)
	assert.Equal(t, g.GetPhase(), restored.Game.GetPhase())
	assert.Equal(t, g.GetRevealedCount(), restored.Game.GetRevealedCount())

	_, err = RestoreIronCrossInteractor([]byte(`{"ph":9}`), new(ironCrossSilentPresenter))
	assert.Error(t, err, "壊れた保存が復元できてしまった")
}
