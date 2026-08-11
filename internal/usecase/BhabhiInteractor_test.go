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

func newMockBhabhiGame() *interfaces.MockBhabhiGame { return new(interfaces.MockBhabhiGame) }

func newMockBhabhiPresenter() *presenter.MockBhabhiPresenter {
	return new(presenter.MockBhabhiPresenter)
}

func TestNewBhabhiInteractor(t *testing.T) {
	assert.NotNil(t, NewBhabhiInteractor(newMockBhabhiGame(), newMockBhabhiPresenter()))
}

func TestNewBhabhiInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewBhabhiInteractor(nil, newMockBhabhiPresenter()) })
	assert.Panics(t, func() { NewBhabhiInteractor(newMockBhabhiGame(), nil) })
}

// **Reset は人間の番が来るまで CPU に打たせる。**
func TestBhabhiInteractorResetWalksToTheHumanTurn(t *testing.T) {
	g := newMockBhabhiGame()
	p := newMockBhabhiPresenter()
	i := NewBhabhiInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(false).Twice()
	g.On("IsHumanTurn").Return(true)
	g.On("CpuPlay").Return()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertNumberOfCalls(t, "CpuPlay", 2)
}

// **人間が上がったあとも最後の 1 人が決まるまで回し切る。**
//
// ここで打ち切ると、まだ CPU の手番の盤面を返してしまい画面が固まる。
// 上限は共通の 1000 ではなく「膠着上限 × 最大人数」から導いている。
func TestBhabhiInteractorRunsPastTheSharedTurnCap(t *testing.T) {
	g := newMockBhabhiGame()
	p := newMockBhabhiPresenter()
	i := NewBhabhiInteractor(g, p)

	const played = maxCpuTurnsPerCall + 500
	require.Less(t, played, bhabhiMaxCpuTurns, "共通上限より大きく、専用上限より小さい手数で試す")

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(false).Times(played)
	g.On("IsHumanTurn").Return(true)
	g.On("CpuPlay").Return()
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNumberOfCalls(t, "CpuPlay", played)
}

// **専用上限に達したら止まる。** 無限ループにはしない。
func TestBhabhiInteractorStopsAtItsOwnCap(t *testing.T) {
	g := newMockBhabhiGame()
	p := newMockBhabhiPresenter()
	i := NewBhabhiInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(false)
	g.On("CpuPlay").Return()
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNumberOfCalls(t, "CpuPlay", bhabhiMaxCpuTurns)
}

// **ゲームが終わったら CPU を回さない。**
func TestBhabhiInteractorStopsAtGameEnd(t *testing.T) {
	g := newMockBhabhiGame()
	p := newMockBhabhiPresenter()
	i := NewBhabhiInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(true)
	g.On("IsHumanTurn").Return(false).Maybe()
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNotCalled(t, "CpuPlay")
}

// **不正なプレイはドメインのエラーをそのまま返し、CPU は動かさない。**
func TestBhabhiInteractorPlayRejectsAndDoesNotAdvance(t *testing.T) {
	g := newMockBhabhiGame()
	p := newMockBhabhiPresenter()
	i := NewBhabhiInteractor(g, p)

	playErr := errors.New("must follow the led suit")
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	g.On("PlayerPlay", 3).Return(playErr)
	p.On("Output", g, playErr).Return("play_error")

	assert.Equal(t, "play_error", i.Play(3))
	g.AssertNotCalled(t, "CpuPlay")
}

// **人間の番でなければドメインには触らせない。**
func TestBhabhiInteractorPlayGuardsOnTurn(t *testing.T) {
	g := newMockBhabhiGame()
	p := newMockBhabhiPresenter()
	i := NewBhabhiInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(false).Maybe()
	p.On("Output", g, nil).Return("blocked")

	assert.Equal(t, "blocked", i.Play(0))
	g.AssertNotCalled(t, "PlayerPlay")
}

func TestBhabhiInteractorPlayAdvancesToTheNextHumanTurn(t *testing.T) {
	g := newMockBhabhiGame()
	p := newMockBhabhiPresenter()
	i := NewBhabhiInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true).Once()
	g.On("PlayerPlay", 1).Return(nil)
	g.On("IsHumanTurn").Return(false).Times(3)
	g.On("IsHumanTurn").Return(true)
	g.On("CpuPlay").Return()
	p.On("Output", g, nil).Return("played")

	assert.Equal(t, "played", i.Play(1))
	g.AssertNumberOfCalls(t, "CpuPlay", 3)
}

// **ゲームが終わっていたら giveup はドメインに届かない。**
func TestBhabhiInteractorGiveUpGuardsAfterGameEnd(t *testing.T) {
	g := newMockBhabhiGame()
	p := newMockBhabhiPresenter()
	i := NewBhabhiInteractor(g, p)

	g.On("GetGameEndFlag").Return(true)
	p.On("Output", g, nil).Return("ended")

	assert.Equal(t, "ended", i.GiveUp())
	g.AssertNotCalled(t, "GiveUp")
}

// **負のコントロール: 続行中なら届く。**
func TestBhabhiInteractorGiveUpReachesTheDomain(t *testing.T) {
	g := newMockBhabhiGame()
	p := newMockBhabhiPresenter()
	i := NewBhabhiInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true).Maybe()
	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("ok")

	assert.Equal(t, "ok", i.GiveUp())
	g.AssertCalled(t, "GiveUp")
}

func TestBhabhiInteractorResetWithConfig(t *testing.T) {
	g := newMockBhabhiGame()
	p := newMockBhabhiPresenter()
	i := NewBhabhiInteractor(g, p)

	cfg := domain.BhabhiConfig{PlayerCnt: 5}
	g.On("SetConfig", cfg).Return()
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	p.On("Output", g, nil).Return("configured")

	assert.Equal(t, "configured", i.ResetWithConfig(cfg))
	g.AssertCalled(t, "SetConfig", cfg)
}

// **人数が範囲外なら弾き、ドメインには載せない。**
func TestBhabhiInteractorResetWithInvalidConfig(t *testing.T) {
	for _, n := range []int{domain.BhabhiMinPlayers - 1, domain.BhabhiMaxPlayers + 1} {
		g := newMockBhabhiGame()
		p := newMockBhabhiPresenter()
		i := NewBhabhiInteractor(g, p)

		p.On("Output", mock.Anything, mock.Anything).Return("cfg_error")

		assert.Equal(t, "cfg_error", i.ResetWithConfig(domain.BhabhiConfig{PlayerCnt: n}))
		g.AssertNotCalled(t, "Reset")
		g.AssertNotCalled(t, "SetConfig", mock.Anything)
	}
}

func TestBhabhiInteractorHintAndLogDelegate(t *testing.T) {
	g := newMockBhabhiGame()
	p := newMockBhabhiPresenter()
	i := NewBhabhiInteractor(g, p)

	p.On("HintOutput", g).Return("hint")
	p.On("ActionLogOutput", g).Return("log")

	assert.Equal(t, "hint", i.Hint())
	assert.Equal(t, "log", i.ActionLog())
}

func TestBhabhiInteractorGetConfigDelegates(t *testing.T) {
	g := newMockBhabhiGame()
	p := newMockBhabhiPresenter()
	i := NewBhabhiInteractor(g, p)

	cfg := domain.BhabhiConfig{PlayerCnt: 6}
	g.On("GetConfig").Return(cfg)
	assert.Equal(t, cfg, i.GetConfig())
}

// **KV から戻したインタラクタが同じ盤面を持っていること。**
func TestRestoreBhabhiInteractor(t *testing.T) {
	src := domain.NewDefaultBhabhi()
	src.Reset()
	data, err := json.Marshal(src)
	require.NoError(t, err)

	restored, err := RestoreBhabhiInteractor(data, new(presenter.MockBhabhiPresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
	assert.Equal(t, src.GetPlayerCnt(), restored.Game.GetPlayerCnt())
	assert.Equal(t, src.GetLeadSuit(), restored.Game.GetLeadSuit())
	assert.Equal(t, src.GetPlayer(0).GetCardsSize(), restored.Game.GetPlayer(0).GetCardsSize())

	_, err = RestoreBhabhiInteractor([]byte("{"), new(presenter.MockBhabhiPresenter))
	assert.Error(t, err)
}
