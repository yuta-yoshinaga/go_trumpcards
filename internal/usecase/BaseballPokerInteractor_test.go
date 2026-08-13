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

func newBaseballInteractorForTest() (*interfaces.MockBaseballPokerGame,
	*presenter.MockBaseballPokerPresenter, *BaseballPokerInteractor,
) {
	mg := new(interfaces.MockBaseballPokerGame)
	mp := new(presenter.MockBaseballPokerPresenter)
	return mg, mp, NewBaseballPokerInteractor(mg, mp)
}

func TestNewBaseballPokerInteractor_NilPanics(t *testing.T) {
	mp := new(presenter.MockBaseballPokerPresenter)
	assert.Panics(t, func() { NewBaseballPokerInteractor(nil, mp) })

	mg := new(interfaces.MockBaseballPokerGame)
	assert.Panics(t, func() { NewBaseballPokerInteractor(mg, nil) })
}

// **配った直後から CPU が動くことがある。** 表の 3 で買い増しを迫られるのが
// CPU なら、その返事まで進めないと人間の手番に戻らない。
func TestBaseballPokerInteractor_ResetDrivesTheCpu(t *testing.T) {
	mg, mp, ci := newBaseballInteractorForTest()
	mg.On("Reset").Return()
	mg.On("CpuPlay").Return()
	mp.On("Output", mg, nil).Return("reset output")

	assert.Equal(t, "reset output", ci.Reset())
	mg.AssertCalled(t, "Reset")
	mg.AssertNumberOfCalls(t, "CpuPlay", 1)
}

func TestBaseballPokerInteractor_ActionReachesTheDomain(t *testing.T) {
	mg, mp, ci := newBaseballInteractorForTest()
	mg.On("GetGameEndFlag").Return(false)
	mg.On("PlayerAction", domain.BaseballActionBet, 20).Return(nil)
	mg.On("CpuPlay").Return()
	mp.On("Output", mg, nil).Return("ok")

	assert.Equal(t, "ok", ci.Action(domain.BaseballActionBet, 20))
	mg.AssertCalled(t, "PlayerAction", domain.BaseballActionBet, 20)
}

// **払うか降りるかはそのまま通す。** 手が強いからと勝手に払わせない。
func TestBaseballPokerInteractor_PassesTheBuyInAnswerThrough(t *testing.T) {
	for _, answer := range []int{domain.BaseballBuyPay, domain.BaseballBuyFold} {
		mg, mp, ci := newBaseballInteractorForTest()
		mg.On("GetGameEndFlag").Return(false)
		mg.On("AnswerBuyIn", answer).Return(nil)
		mg.On("CpuPlay").Return()
		mp.On("Output", mg, nil).Return("ok")

		assert.Equal(t, "ok", ci.AnswerBuyIn(answer))
		mg.AssertCalled(t, "AnswerBuyIn", answer)
	}
}

// **不正な返事はドメインが弾く。** usecase では判定しない。
func TestBaseballPokerInteractor_LetsTheDomainRejectABadAnswer(t *testing.T) {
	mg, mp, ci := newBaseballInteractorForTest()
	boom := errors.New("answer pay or fold")
	mg.On("GetGameEndFlag").Return(false)
	mg.On("AnswerBuyIn", 99).Return(boom)
	mg.On("CpuPlay").Return()
	mp.On("Output", mg, boom).Return("error output")

	assert.Equal(t, "error output", ci.AnswerBuyIn(99))
	mg.AssertCalled(t, "AnswerBuyIn", 99)
}

// **人間が 1 手指したら CPU が動く。**
func TestBaseballPokerInteractor_DrivesTheCpu(t *testing.T) {
	for _, tt := range []struct {
		name   string
		setup  func(*interfaces.MockBaseballPokerGame)
		invoke func(*BaseballPokerInteractor) string
	}{
		{
			name: "Action",
			setup: func(m *interfaces.MockBaseballPokerGame) {
				m.On("PlayerAction", domain.BaseballActionCheck, 0).Return(nil)
			},
			invoke: func(ci *BaseballPokerInteractor) string {
				return ci.Action(domain.BaseballActionCheck, 0)
			},
		},
		{
			name: "AnswerBuyIn",
			setup: func(m *interfaces.MockBaseballPokerGame) {
				m.On("AnswerBuyIn", domain.BaseballBuyPay).Return(nil)
			},
			invoke: func(ci *BaseballPokerInteractor) string {
				return ci.AnswerBuyIn(domain.BaseballBuyPay)
			},
		},
		{
			name:   "NextHand",
			setup:  func(m *interfaces.MockBaseballPokerGame) { m.On("NextHand").Return(nil) },
			invoke: func(ci *BaseballPokerInteractor) string { return ci.NextHand() },
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mg, mp, ci := newBaseballInteractorForTest()
			tt.setup(mg)
			mg.On("GetGameEndFlag").Return(false)
			mg.On("CpuPlay").Return()
			mp.On("Output", mg, nil).Return("ok")

			tt.invoke(ci)
			mg.AssertNumberOfCalls(t, "CpuPlay", 1)
		})
	}
}

func TestBaseballPokerInteractor_DoesNotDriveTheCpuOnError(t *testing.T) {
	mg, mp, ci := newBaseballInteractorForTest()
	boom := errors.New("nope")
	mg.On("GetGameEndFlag").Return(false)
	mg.On("PlayerAction", domain.BaseballActionCheck, 0).Return(boom)
	mg.On("CpuPlay").Return()
	mp.On("Output", mg, boom).Return("error output")

	assert.Equal(t, "error output", ci.Action(domain.BaseballActionCheck, 0))
	mg.AssertNumberOfCalls(t, "CpuPlay", 0)
}

func TestBaseballPokerInteractor_BlocksAfterGameEnd(t *testing.T) {
	mg, mp, ci := newBaseballInteractorForTest()
	mg.On("GetGameEndFlag").Return(true)
	mp.On("Output", mg, mock.Anything).Return("finished")

	assert.Equal(t, "finished", ci.Action(domain.BaseballActionCheck, 0))
	assert.Equal(t, "finished", ci.AnswerBuyIn(domain.BaseballBuyPay))
	assert.Equal(t, "finished", ci.NextHand())
	mg.AssertNotCalled(t, "PlayerAction", domain.BaseballActionCheck, 0)
	mg.AssertNotCalled(t, "AnswerBuyIn", domain.BaseballBuyPay)
	mg.AssertNotCalled(t, "NextHand")
}

func TestBaseballPokerInteractor_HintAndLog(t *testing.T) {
	mg, mp, ci := newBaseballInteractorForTest()
	mp.On("HintOutput", mg).Return("hint")
	mp.On("ActionLogOutput", mg).Return("log")

	assert.Equal(t, "hint", ci.Hint())
	assert.Equal(t, "log", ci.ActionLog())
}

func TestBaseballPokerInteractor_Config(t *testing.T) {
	mg, mp, ci := newBaseballInteractorForTest()
	cfg := domain.DefaultBaseballPokerConfig()
	mg.On("GetConfig").Return(cfg)
	assert.Equal(t, cfg, ci.GetConfig())

	// **山が足りない席数は弾く。** 7 席は範囲でも山でも通らない。
	bad := domain.BaseballPokerConfig{Seats: 7, InitialChips: 1000, Ante: 10}
	mp.On("Output", mg, mock.Anything).Return("bad config")
	assert.Equal(t, "bad config", ci.ResetWithConfig(bad))
	mg.AssertNotCalled(t, "SetConfig", bad)

	good := domain.BaseballPokerConfig{Seats: 3, InitialChips: 500, Ante: 5}
	mg.On("SetConfig", good).Return()
	mg.On("Reset").Return()
	mg.On("CpuPlay").Return()
	ci.ResetWithConfig(good)
	mg.AssertCalled(t, "SetConfig", good)
}

// --- 本物のドメインで駆動を確かめる ---

type baseballSilentPresenter struct{}

func (p *baseballSilentPresenter) Output(interfaces.BaseballPokerGame, error) string   { return "" }
func (p *baseballSilentPresenter) ActionLogOutput(interfaces.BaseballPokerGame) string { return "" }
func (p *baseballSilentPresenter) HintOutput(interfaces.BaseballPokerGame) string      { return "" }

// **人間の操作 1 回で、盤面は人間の判断が要る場面まで戻る。**
//
// テストが自分でループを回すと、駆動が抜けていても気づけない ── 人間の操作を
// 1 回だけ実行して、どこで止まったかを見る。
func TestBaseballPokerInteractor_OneActionReturnsControl(t *testing.T) {
	for range 30 {
		g := domain.NewDefaultBaseballPoker()
		ci := NewBaseballPokerInteractor(g, new(baseballSilentPresenter))
		ci.Reset()

		switch {
		case g.IsHumanBuying():
			ci.AnswerBuyIn(domain.BaseballBuyPay)
		case g.IsHumanTurn():
			ci.Action(domain.BaseballActionCheck, 0)
		default:
			t.Fatalf("Reset の直後に人間の手番でも買い増しでもない (席 %d)", g.GetTurnSeat())
		}

		if g.GetPhase() == domain.BaseballPhaseShowdown || g.GetGameEndFlag() {
			continue
		}
		require.True(t, g.IsHumanTurn() || g.IsHumanBuying(),
			"人間が 1 手指したのに席 %d (CPU) で盤面が止まっている", g.GetTurnSeat())
	}
}

func TestBaseballPokerInteractor_SnapshotRestore(t *testing.T) {
	g := domain.NewDefaultBaseballPoker()
	ci := NewBaseballPokerInteractor(g, new(baseballSilentPresenter))
	ci.Reset()

	data, err := ci.Snapshot()
	require.NoError(t, err)
	assert.True(t, json.Valid(data))

	restored, err := RestoreBaseballPokerInteractor(data, new(baseballSilentPresenter))
	require.NoError(t, err)
	assert.Equal(t, g.GetPhase(), restored.Game.GetPhase())
	assert.Equal(t, g.GetStreet(), restored.Game.GetStreet())
	assert.Equal(t, g.GetBuyerSeat(), restored.Game.GetBuyerSeat())

	_, err = RestoreBaseballPokerInteractor([]byte(`{"ph":9}`), new(baseballSilentPresenter))
	assert.Error(t, err, "壊れた保存が復元できてしまった")
}
