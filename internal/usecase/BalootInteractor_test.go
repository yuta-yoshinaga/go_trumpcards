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

func newMockBalootGame() *interfaces.MockBalootGame { return new(interfaces.MockBalootGame) }

func newMockBalootPresenter() *presenter.MockBalootPresenter {
	return new(presenter.MockBalootPresenter)
}

func TestNewBalootInteractor(t *testing.T) {
	assert.NotNil(t, NewBalootInteractor(newMockBalootGame(), newMockBalootPresenter()))
}

func TestNewBalootInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewBalootInteractor(nil, newMockBalootPresenter()) })
	assert.Panics(t, func() { NewBalootInteractor(newMockBalootGame(), nil) })
}

// Reset のあと、人間の宣言番になるまで CPU が宣言する。
func TestBalootInteractorResetRunsCpuDeclaresToHumanTurn(t *testing.T) {
	g := newMockBalootGame()
	p := newMockBalootPresenter()
	i := NewBalootInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.BalootPhaseDeclare)
	g.On("IsHumanDeclareTurn").Return(false).Twice()
	g.On("IsHumanDeclareTurn").Return(true)
	g.On("CpuDeclare").Return()
	// Reset は宣言のあとプレイも進めるので、手番判定も呼ばれる。
	g.On("IsHumanTurn").Return(true)
	g.On("GetHint").Return(nil).Maybe()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertNumberOfCalls(t, "CpuDeclare", 2)
	g.AssertNotCalled(t, "CpuPlay")
}

// **CPU がモードを決めたら、そのままプレイも進める。**
//
// ここで止めると、リード（親の左隣＝CPU）のまま誰も打たず、人間の手番が
// 永久に来ない盤面を返してしまう（tarabish で e2e が 3 回に 1 回踏んだ形）。
func TestBalootInteractorResetPlaysOnWhenACpuDeclares(t *testing.T) {
	g := newMockBalootGame()
	p := newMockBalootPresenter()
	i := NewBalootInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.BalootPhasePlay)
	g.On("IsHumanDeclareTurn").Return(true)
	g.On("IsHumanTurn").Return(false).Times(3)
	g.On("IsHumanTurn").Return(true)
	g.On("CpuPlay").Return()
	g.On("GetHint").Return(nil).Maybe()
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNumberOfCalls(t, "CpuPlay", 3)
}

// **Sun / Hokom / Pass はそれぞれ別のドメイン操作に落ちる。** 取り違えると
// 宣言していないモードでラウンドが進む。
func TestBalootInteractorDeclareCommandsAreDistinct(t *testing.T) {
	for _, tc := range []struct {
		name   string
		call   func(*BalootInteractor) string
		method string
		others []string
	}{
		{"sun", func(i *BalootInteractor) string { return i.DeclareSun() }, "DeclareSun", []string{"DeclareHokom", "PassDeclaration"}},
		{"pass", func(i *BalootInteractor) string { return i.PassDeclaration() }, "PassDeclaration", []string{"DeclareSun", "DeclareHokom"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockBalootGame()
			p := newMockBalootPresenter()
			i := NewBalootInteractor(g, p)

			g.On("GetGameEndFlag").Return(false)
			g.On(tc.method).Return(nil)
			g.On("GetPhase").Return(domain.BalootPhasePlay)
			g.On("IsHumanTurn").Return(true)
			p.On("Output", g, nil).Return("declare_output")

			assert.Equal(t, "declare_output", tc.call(i))
			g.AssertCalled(t, tc.method)
			for _, other := range tc.others {
				g.AssertNotCalled(t, other)
			}
		})
	}
}

// **Hokom は選んだスートをそのままドメインへ渡す。** 既定値で埋めると
// プレイヤーが選んでいないスートが切り札になる。
func TestBalootInteractorDeclareHokomPassesTheSuitThrough(t *testing.T) {
	g := newMockBalootGame()
	p := newMockBalootPresenter()
	i := NewBalootInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("DeclareHokom", domain.CardDesignDiamond).Return(nil)
	g.On("GetPhase").Return(domain.BalootPhasePlay)
	g.On("IsHumanTurn").Return(true)
	p.On("Output", g, nil).Return("hokom")

	assert.Equal(t, "hokom", i.DeclareHokom(domain.CardDesignDiamond))
	g.AssertCalled(t, "DeclareHokom", domain.CardDesignDiamond)
	g.AssertNotCalled(t, "DeclareSun")
}

// **親が見送ろうとしたらドメインが拒否し、その error が presenter に渡る。**
func TestBalootInteractorPassRejectedForDealer(t *testing.T) {
	g := newMockBalootGame()
	p := newMockBalootPresenter()
	i := NewBalootInteractor(g, p)

	err := errors.New("the dealer must declare")
	g.On("GetGameEndFlag").Return(false)
	g.On("PassDeclaration").Return(err)
	p.On("Output", g, err).Return("pass_error")

	assert.Equal(t, "pass_error", i.PassDeclaration())
	g.AssertNotCalled(t, "CpuDeclare")
	g.AssertNotCalled(t, "CpuPlay")
}

func TestBalootInteractorPlay(t *testing.T) {
	g := newMockBalootGame()
	p := newMockBalootPresenter()
	i := NewBalootInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	g.On("PlayerPlay", 2).Return(nil)
	p.On("Output", g, nil).Return("played")

	assert.Equal(t, "played", i.Play(2))
	g.AssertCalled(t, "PlayerPlay", 2)
}

func TestBalootInteractorPlayError(t *testing.T) {
	g := newMockBalootGame()
	p := newMockBalootPresenter()
	i := NewBalootInteractor(g, p)

	err := errors.New("must follow suit")
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	g.On("PlayerPlay", 2).Return(err)
	p.On("Output", g, err).Return("error_output")

	assert.Equal(t, "error_output", i.Play(2))
	g.AssertNotCalled(t, "CpuPlay")
}

func TestBalootInteractorPlayBlocked(t *testing.T) {
	for _, tc := range []struct {
		name    string
		ended   bool
		isHuman bool
	}{
		{"game over", true, true},
		{"not human turn", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockBalootGame()
			p := newMockBalootPresenter()
			i := NewBalootInteractor(g, p)

			g.On("GetGameEndFlag").Return(tc.ended)
			g.On("IsHumanTurn").Return(tc.isHuman)
			p.On("Output", g, nil).Return("blocked")

			assert.Equal(t, "blocked", i.Play(0))
			g.AssertNotCalled(t, "PlayerPlay", 0)
		})
	}
}

// **プレイ中は宣言ループを回さない。** フェーズが違えば即座に抜ける。
func TestBalootInteractorDeclareLoopStopsOutsideDeclarePhase(t *testing.T) {
	g := newMockBalootGame()
	p := newMockBalootPresenter()
	i := NewBalootInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("DeclareSun").Return(nil)
	g.On("GetPhase").Return(domain.BalootPhaseRoundEnd)
	g.On("IsHumanTurn").Return(false)
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.DeclareSun())
	g.AssertNotCalled(t, "CpuDeclare")
	g.AssertNotCalled(t, "CpuPlay")
}

func TestBalootInteractorCpuLoopsHaveAnUpperBound(t *testing.T) {
	t.Run("declares", func(t *testing.T) {
		g := newMockBalootGame()
		p := newMockBalootPresenter()
		i := NewBalootInteractor(g, p)

		g.On("Reset").Return()
		g.On("GetGameEndFlag").Return(false)
		g.On("GetPhase").Return(domain.BalootPhaseDeclare)
		g.On("IsHumanDeclareTurn").Return(false)
		g.On("CpuDeclare").Return()
		// 宣言ループが上限で抜けたあと、Reset はプレイ側も試す。
		g.On("IsHumanTurn").Return(true)
		g.On("GetHint").Return(nil).Maybe()
		p.On("Output", g, nil).Return("out")

		assert.Equal(t, "out", i.Reset())
		g.AssertNumberOfCalls(t, "CpuDeclare", maxCpuTurnsPerCall)
	})

	t.Run("plays", func(t *testing.T) {
		g := newMockBalootGame()
		p := newMockBalootPresenter()
		i := NewBalootInteractor(g, p)

		g.On("GetGameEndFlag").Return(false)
		g.On("DeclareSun").Return(nil)
		g.On("GetPhase").Return(domain.BalootPhasePlay)
		g.On("IsHumanDeclareTurn").Return(true)
		g.On("IsHumanTurn").Return(false)
		g.On("CpuPlay").Return()
		p.On("Output", g, nil).Return("out")

		assert.Equal(t, "out", i.DeclareSun())
		g.AssertNumberOfCalls(t, "CpuPlay", maxCpuTurnsPerCall)
	})
}

func TestBalootInteractorNextRound(t *testing.T) {
	g := newMockBalootGame()
	p := newMockBalootPresenter()
	i := NewBalootInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("NextRound").Return()
	g.On("GetPhase").Return(domain.BalootPhaseDeclare)
	g.On("IsHumanDeclareTurn").Return(true)
	g.On("IsHumanTurn").Return(false)
	p.On("Output", g, nil).Return("next")

	assert.Equal(t, "next", i.NextRound())
	g.AssertCalled(t, "NextRound")
}

func TestBalootInteractorGuardsAfterGameEnd(t *testing.T) {
	for _, tc := range []struct {
		name   string
		call   func(*BalootInteractor) string
		method string
	}{
		{"next round", func(i *BalootInteractor) string { return i.NextRound() }, "NextRound"},
		{"give up", func(i *BalootInteractor) string { return i.GiveUp() }, "GiveUp"},
		{"sun", func(i *BalootInteractor) string { return i.DeclareSun() }, "DeclareSun"},
		{"hokom", func(i *BalootInteractor) string { return i.DeclareHokom(domain.CardDesignHeart) }, "DeclareHokom"},
		{"pass", func(i *BalootInteractor) string { return i.PassDeclaration() }, "PassDeclaration"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockBalootGame()
			p := newMockBalootPresenter()
			i := NewBalootInteractor(g, p)

			g.On("GetGameEndFlag").Return(true)
			p.On("Output", g, nil).Return("over")

			assert.Equal(t, "over", tc.call(i))
			g.AssertNotCalled(t, tc.method)
		})
	}
}

func TestBalootInteractorGiveUp(t *testing.T) {
	g := newMockBalootGame()
	p := newMockBalootPresenter()
	i := NewBalootInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("gave_up")

	assert.Equal(t, "gave_up", i.GiveUp())
	g.AssertCalled(t, "GiveUp")
}

func TestBalootInteractorGetConfig(t *testing.T) {
	g := newMockBalootGame()
	i := NewBalootInteractor(g, newMockBalootPresenter())
	cfg := domain.BalootConfig{Target: 200}
	g.On("GetConfig").Return(cfg)
	assert.Equal(t, cfg, i.GetConfig())
}

func TestBalootInteractorResetWithInvalidConfig(t *testing.T) {
	g := newMockBalootGame()
	p := newMockBalootPresenter()
	i := NewBalootInteractor(g, p)

	p.On("Output", mock.Anything, mock.Anything).Return("cfg_error")

	assert.Equal(t, "cfg_error", i.ResetWithConfig(domain.BalootConfig{Target: 5}))
	g.AssertNotCalled(t, "Reset")
	g.AssertNotCalled(t, "SetConfig", mock.Anything)
}

func TestBalootInteractorHintAndActionLog(t *testing.T) {
	g := newMockBalootGame()
	p := newMockBalootPresenter()
	i := NewBalootInteractor(g, p)

	p.On("HintOutput", g).Return("hint_output")
	p.On("ActionLogOutput", g).Return("log_output")

	assert.Equal(t, "hint_output", i.Hint())
	assert.Equal(t, "log_output", i.ActionLog())
}

// Worker はリクエストごとに KV から作り直す。**モードと切り札が往復しないと
// 札の強さが毎リクエスト変わる** (#4478)。
func TestBalootInteractorSnapshotRoundTrip(t *testing.T) {
	g := domain.NewDefaultBaloot()
	g.Reset()
	g.SetCurrentPlayerIdxForTest(0)
	require.NoError(t, g.DeclareHokom(domain.CardDesignDiamond))
	g.SetScoreForTestUse(0, 140)
	g.GetPlayer(0).SetHasBaloot(true)

	p := newMockBalootPresenter()
	i := NewBalootInteractor(g, p)

	data, err := i.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreBalootInteractor(data, p)
	require.NoError(t, err)

	assert.Equal(t, 140, restored.Game.GetScore(0))
	assert.True(t, restored.Game.GetPlayer(0).GetHasBaloot())
	assert.Equal(t, domain.BalootModeHokom, restored.Game.GetMode())
	assert.Equal(t, domain.CardDesignDiamond, restored.Game.GetTrumpSuit())
	assert.Equal(t, g.GetDeclarerIdx(), restored.Game.GetDeclarerIdx())
	assert.Equal(t, g.GetConfig().Target, restored.Game.GetConfig().Target)
}

func TestRestoreBalootInteractorRejectsGarbage(t *testing.T) {
	_, err := RestoreBalootInteractor([]byte("not json"), newMockBalootPresenter())
	assert.Error(t, err)
}
