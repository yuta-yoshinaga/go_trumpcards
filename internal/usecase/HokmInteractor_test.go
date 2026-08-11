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

func newMockHokmGame() *interfaces.MockHokmGame { return new(interfaces.MockHokmGame) }

func newMockHokmPresenter() *presenter.MockHokmPresenter { return new(presenter.MockHokmPresenter) }

func TestNewHokmInteractor(t *testing.T) {
	assert.NotNil(t, NewHokmInteractor(newMockHokmGame(), newMockHokmPresenter()))
}

func TestNewHokmInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewHokmInteractor(nil, newMockHokmPresenter()) })
	assert.Panics(t, func() { NewHokmInteractor(newMockHokmGame(), nil) })
}

// **CPU の親なら宣言させ、そのままプレイも進める。** 宣言で止めると
// リード（親）のまま誰も打たず、人間の手番が来ない盤面を返す。
func TestHokmInteractorResetDeclaresThenPlays(t *testing.T) {
	g := newMockHokmGame()
	p := newMockHokmPresenter()
	i := NewHokmInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.HokmPhaseTrump).Once()
	g.On("IsHumanTrumpTurn").Return(false)
	g.On("CpuDeclareTrump").Return()
	g.On("GetPhase").Return(domain.HokmPhasePlay)
	g.On("IsHumanTurn").Return(false).Times(3)
	g.On("IsHumanTurn").Return(true)
	g.On("CpuPlay").Return()
	g.On("GetHint").Return(nil).Maybe()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertCalled(t, "CpuDeclareTrump")
	g.AssertNumberOfCalls(t, "CpuPlay", 3)
}

// **人間が親なら勝手に宣言しない。** そこで止めて選ばせる。
func TestHokmInteractorResetStopsAtAHumanHakem(t *testing.T) {
	g := newMockHokmGame()
	p := newMockHokmPresenter()
	i := NewHokmInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.HokmPhaseTrump)
	g.On("IsHumanTrumpTurn").Return(true)
	g.On("IsHumanTurn").Return(false).Maybe()
	g.On("GetHint").Return(nil).Maybe()
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNotCalled(t, "CpuDeclareTrump")
	g.AssertNotCalled(t, "CpuPlay")
}

// 宣言はそのままドメインへ渡り、そのあと CPU が打つ。
func TestHokmInteractorDeclareTrump(t *testing.T) {
	g := newMockHokmGame()
	p := newMockHokmPresenter()
	i := NewHokmInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("PlayerDeclareTrump", domain.CardDesignHeart).Return(nil)
	g.On("GetPhase").Return(domain.HokmPhasePlay)
	g.On("IsHumanTurn").Return(true)
	p.On("Output", g, nil).Return("declared")

	assert.Equal(t, "declared", i.DeclareTrump(domain.CardDesignHeart))
	g.AssertCalled(t, "PlayerDeclareTrump", domain.CardDesignHeart)
}

func TestHokmInteractorDeclareTrumpRejected(t *testing.T) {
	g := newMockHokmGame()
	p := newMockHokmPresenter()
	i := NewHokmInteractor(g, p)

	err := errors.New("only the hakem declares the trump suit")
	g.On("GetGameEndFlag").Return(false)
	g.On("PlayerDeclareTrump", 1).Return(err)
	p.On("Output", g, err).Return("trump_error")

	assert.Equal(t, "trump_error", i.DeclareTrump(1))
	g.AssertNotCalled(t, "CpuPlay")
}

func TestHokmInteractorPlay(t *testing.T) {
	g := newMockHokmGame()
	p := newMockHokmPresenter()
	i := NewHokmInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	g.On("PlayerPlay", 2).Return(nil)
	p.On("Output", g, nil).Return("played")

	assert.Equal(t, "played", i.Play(2))
	g.AssertCalled(t, "PlayerPlay", 2)
}

func TestHokmInteractorPlayError(t *testing.T) {
	g := newMockHokmGame()
	p := newMockHokmPresenter()
	i := NewHokmInteractor(g, p)

	err := errors.New("must follow suit")
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	g.On("PlayerPlay", 2).Return(err)
	p.On("Output", g, err).Return("error_output")

	assert.Equal(t, "error_output", i.Play(2))
	g.AssertNotCalled(t, "CpuPlay")
}

func TestHokmInteractorPlayBlocked(t *testing.T) {
	for _, tc := range []struct {
		name    string
		ended   bool
		isHuman bool
	}{
		{"game over", true, true},
		{"not human turn", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockHokmGame()
			p := newMockHokmPresenter()
			i := NewHokmInteractor(g, p)

			g.On("GetGameEndFlag").Return(tc.ended)
			g.On("IsHumanTurn").Return(tc.isHuman)
			p.On("Output", g, nil).Return("blocked")

			assert.Equal(t, "blocked", i.Play(0))
			g.AssertNotCalled(t, "PlayerPlay", 0)
		})
	}
}

// **ハンドが終わっていればプレイのループは回らない。**
func TestHokmInteractorPlayLoopStopsOutsidePlayPhase(t *testing.T) {
	g := newMockHokmGame()
	p := newMockHokmPresenter()
	i := NewHokmInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("PlayerDeclareTrump", 1).Return(nil)
	g.On("GetPhase").Return(domain.HokmPhaseHandEnd)
	g.On("IsHumanTurn").Return(false)
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.DeclareTrump(1))
	g.AssertNotCalled(t, "CpuPlay")
}

func TestHokmInteractorCpuLoopHasAnUpperBound(t *testing.T) {
	g := newMockHokmGame()
	p := newMockHokmPresenter()
	i := NewHokmInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.HokmPhasePlay)
	g.On("IsHumanTrumpTurn").Return(true)
	g.On("IsHumanTurn").Return(false)
	g.On("CpuPlay").Return()
	g.On("GetHint").Return(nil).Maybe()
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNumberOfCalls(t, "CpuPlay", maxCpuTurnsPerCall)
}

func TestHokmInteractorNextHand(t *testing.T) {
	g := newMockHokmGame()
	p := newMockHokmPresenter()
	i := NewHokmInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("NextHand").Return()
	g.On("GetPhase").Return(domain.HokmPhaseTrump)
	g.On("IsHumanTrumpTurn").Return(true)
	g.On("IsHumanTurn").Return(false)
	p.On("Output", g, nil).Return("next")

	assert.Equal(t, "next", i.NextHand())
	g.AssertCalled(t, "NextHand")
}

func TestHokmInteractorGuardsAfterGameEnd(t *testing.T) {
	for _, tc := range []struct {
		name   string
		call   func(*HokmInteractor) string
		method string
	}{
		{"next hand", func(i *HokmInteractor) string { return i.NextHand() }, "NextHand"},
		{"give up", func(i *HokmInteractor) string { return i.GiveUp() }, "GiveUp"},
		{"declare trump", func(i *HokmInteractor) string { return i.DeclareTrump(1) }, "PlayerDeclareTrump"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockHokmGame()
			p := newMockHokmPresenter()
			i := NewHokmInteractor(g, p)

			g.On("GetGameEndFlag").Return(true)
			p.On("Output", g, nil).Return("over")

			assert.Equal(t, "over", tc.call(i))
			g.AssertNotCalled(t, tc.method)
		})
	}
}

func TestHokmInteractorGiveUp(t *testing.T) {
	g := newMockHokmGame()
	p := newMockHokmPresenter()
	i := NewHokmInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("gave_up")

	assert.Equal(t, "gave_up", i.GiveUp())
	g.AssertCalled(t, "GiveUp")
}

func TestHokmInteractorGetConfig(t *testing.T) {
	g := newMockHokmGame()
	i := NewHokmInteractor(g, newMockHokmPresenter())
	cfg := domain.HokmConfig{Target: 9}
	g.On("GetConfig").Return(cfg)
	assert.Equal(t, cfg, i.GetConfig())
}

func TestHokmInteractorResetWithInvalidConfig(t *testing.T) {
	g := newMockHokmGame()
	p := newMockHokmPresenter()
	i := NewHokmInteractor(g, p)

	p.On("Output", mock.Anything, mock.Anything).Return("cfg_error")

	assert.Equal(t, "cfg_error", i.ResetWithConfig(domain.HokmConfig{Target: 0}))
	g.AssertNotCalled(t, "Reset")
	g.AssertNotCalled(t, "SetConfig", mock.Anything)
}

func TestHokmInteractorHintAndActionLog(t *testing.T) {
	g := newMockHokmGame()
	p := newMockHokmPresenter()
	i := NewHokmInteractor(g, p)

	p.On("HintOutput", g).Return("hint_output")
	p.On("ActionLogOutput", g).Return("log_output")

	assert.Equal(t, "hint_output", i.Hint())
	assert.Equal(t, "log_output", i.ActionLog())
}

// Worker はリクエストごとに KV から作り直す。**獲得トリック数が往復しないと
// 7 先取も Kot も判定できない** (#4478)。
func TestHokmInteractorSnapshotRoundTrip(t *testing.T) {
	g := domain.NewDefaultHokm()
	g.Reset()
	g.SetHakemIdxForTest(0)
	require.NoError(t, g.PlayerDeclareTrump(domain.CardDesignClover))
	g.SetScoreForTestUse(0, 4)
	g.GiveTricksForTest(0, 3)

	p := newMockHokmPresenter()
	i := NewHokmInteractor(g, p)

	data, err := i.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreHokmInteractor(data, p)
	require.NoError(t, err)

	assert.Equal(t, domain.CardDesignClover, restored.Game.GetTrumpSuit())
	assert.Equal(t, 0, restored.Game.GetHakemIdx())
	assert.Equal(t, 4, restored.Game.GetScore(0))
	assert.Equal(t, 3, restored.Game.TeamTricks(0))
	assert.Equal(t, g.GetConfig().Target, restored.Game.GetConfig().Target)
}

func TestRestoreHokmInteractorRejectsGarbage(t *testing.T) {
	_, err := RestoreHokmInteractor([]byte("not json"), newMockHokmPresenter())
	assert.Error(t, err)
}
