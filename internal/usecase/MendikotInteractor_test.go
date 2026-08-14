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

func newMockMendikotGame() *interfaces.MockMendikotGame { return new(interfaces.MockMendikotGame) }

func newMockMendikotPresenter() *presenter.MockMendikotPresenter {
	return new(presenter.MockMendikotPresenter)
}

func TestNewMendikotInteractor(t *testing.T) {
	assert.NotNil(t, NewMendikotInteractor(newMockMendikotGame(), newMockMendikotPresenter()))
}

func TestNewMendikotInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewMendikotInteractor(nil, newMockMendikotPresenter()) })
	assert.Panics(t, func() { NewMendikotInteractor(newMockMendikotGame(), nil) })
}

// **Reset は人間の番が来るまで CPU に打たせる。** リードは親の左隣なので、
// 人間が親でなければ CPU が先に出す。ここで止めると誰も打てない盤面を返す。
func TestMendikotInteractorResetWalksToTheHumanTurn(t *testing.T) {
	g := newMockMendikotGame()
	p := newMockMendikotPresenter()
	i := NewMendikotInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.MendikotPhasePlay)
	g.On("IsHumanTurn").Return(false).Twice()
	g.On("IsHumanTurn").Return(true)
	g.On("CpuPlay").Return()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertNumberOfCalls(t, "CpuPlay", 2)
}

// **ハンドが終わったら CPU を回さない。** 次のハンドは next で明示的に始める。
func TestMendikotInteractorStopsAtHandEnd(t *testing.T) {
	g := newMockMendikotGame()
	p := newMockMendikotPresenter()
	i := NewMendikotInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(false)
	g.On("GetPhase").Return(domain.MendikotPhaseHandEnd)
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNotCalled(t, "CpuPlay")
}

// **不正なプレイはドメインのエラーをそのまま返し、CPU は動かさない。**
func TestMendikotInteractorPlayRejectsAndDoesNotAdvance(t *testing.T) {
	g := newMockMendikotGame()
	p := newMockMendikotPresenter()
	i := NewMendikotInteractor(g, p)

	playErr := errors.New("follow suit")
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.MendikotPhasePlay)
	g.On("IsHumanTurn").Return(true)
	g.On("PlayerPlay", 3).Return(playErr)
	p.On("Output", g, playErr).Return("play_error")

	assert.Equal(t, "play_error", i.Play(3))
	g.AssertNotCalled(t, "CpuPlay")
}

// **人間の番でなければドメインには触らせない。**
func TestMendikotInteractorPlayGuardsOnTurn(t *testing.T) {
	g := newMockMendikotGame()
	p := newMockMendikotPresenter()
	i := NewMendikotInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.MendikotPhaseHandEnd)
	g.On("IsHumanTurn").Return(false).Maybe()
	p.On("Output", g, nil).Return("blocked")

	assert.Equal(t, "blocked", i.Play(0))
	g.AssertNotCalled(t, "PlayerPlay")
}

// **成功したプレイはそのままトリックを解決し、次の人間の番まで進める。**
func TestMendikotInteractorPlayAdvancesToTheNextHumanTurn(t *testing.T) {
	g := newMockMendikotGame()
	p := newMockMendikotPresenter()
	i := NewMendikotInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.MendikotPhasePlay)
	g.On("IsHumanTurn").Return(true).Once()
	g.On("PlayerPlay", 1).Return(nil)
	g.On("IsHumanTurn").Return(false).Times(3)
	g.On("IsHumanTurn").Return(true)
	g.On("CpuPlay").Return()
	p.On("Output", g, nil).Return("played")

	assert.Equal(t, "played", i.Play(1))
	g.AssertNumberOfCalls(t, "CpuPlay", 3)
}

// **ゲームが終わっていたら next も giveup もドメインに届かない。**
func TestMendikotInteractorGuardsAfterGameEnd(t *testing.T) {
	for _, tc := range []struct {
		name   string
		call   func(*MendikotInteractor) string
		method string
	}{
		{"next", func(i *MendikotInteractor) string { return i.NextHand() }, "NextHand"},
		{"giveup", func(i *MendikotInteractor) string { return i.GiveUp() }, "GiveUp"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockMendikotGame()
			p := newMockMendikotPresenter()
			i := NewMendikotInteractor(g, p)

			g.On("GetGameEndFlag").Return(true)
			p.On("Output", g, nil).Return("ended")

			assert.Equal(t, "ended", tc.call(i))
			g.AssertNotCalled(t, tc.method)
		})
	}
}

// **負のコントロール: 続行中ならどちらもドメインに届く。**
func TestMendikotInteractorNextAndGiveUpReachTheDomain(t *testing.T) {
	for _, tc := range []struct {
		name   string
		call   func(*MendikotInteractor) string
		method string
	}{
		{"next", func(i *MendikotInteractor) string { return i.NextHand() }, "NextHand"},
		{"giveup", func(i *MendikotInteractor) string { return i.GiveUp() }, "GiveUp"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockMendikotGame()
			p := newMockMendikotPresenter()
			i := NewMendikotInteractor(g, p)

			g.On("GetGameEndFlag").Return(false)
			g.On("GetPhase").Return(domain.MendikotPhasePlay).Maybe()
			g.On("IsHumanTurn").Return(true).Maybe()
			g.On(tc.method).Return()
			p.On("Output", g, nil).Return("ok")

			assert.Equal(t, "ok", tc.call(i))
			g.AssertCalled(t, tc.method)
		})
	}
}

// **設定は検証を通ってからドメインに載る。**
func TestMendikotInteractorResetWithConfig(t *testing.T) {
	g := newMockMendikotGame()
	p := newMockMendikotPresenter()
	i := NewMendikotInteractor(g, p)

	cfg := domain.MendikotConfig{Target: 5}
	g.On("SetConfig", cfg).Return()
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.MendikotPhasePlay).Maybe()
	g.On("IsHumanTurn").Return(true)
	p.On("Output", g, nil).Return("configured")

	assert.Equal(t, "configured", i.ResetWithConfig(cfg))
	g.AssertCalled(t, "SetConfig", cfg)
}

// **範囲外の目標点は弾き、ドメインには載せない。**
func TestMendikotInteractorResetWithConfigRejectsOutOfRange(t *testing.T) {
	g := newMockMendikotGame()
	p := newMockMendikotPresenter()
	i := NewMendikotInteractor(g, p)

	p.On("Output", mock.Anything, mock.Anything).Return("cfg_error")

	assert.Equal(t, "cfg_error", i.ResetWithConfig(domain.MendikotConfig{Target: 0}))
	g.AssertNotCalled(t, "Reset")
	g.AssertNotCalled(t, "SetConfig", mock.Anything)
}

func TestMendikotInteractorHintAndLogDelegate(t *testing.T) {
	g := newMockMendikotGame()
	p := newMockMendikotPresenter()
	i := NewMendikotInteractor(g, p)

	p.On("HintOutput", g).Return("hint")
	p.On("ActionLogOutput", g).Return("log")

	assert.Equal(t, "hint", i.Hint())
	assert.Equal(t, "log", i.ActionLog())
}

func TestMendikotInteractorGetConfigDelegates(t *testing.T) {
	g := newMockMendikotGame()
	p := newMockMendikotPresenter()
	i := NewMendikotInteractor(g, p)

	cfg := domain.MendikotConfig{Target: 7}
	g.On("GetConfig").Return(cfg)
	assert.Equal(t, cfg, i.GetConfig())
}

// **KV から戻したインタラクタが同じ盤面を持っていること。** Worker は毎
// リクエスト JSON から作り直すので、ここが欠けると状態が毎回消える。
func TestRestoreMendikotInteractor(t *testing.T) {
	src := domain.NewDefaultMendikot()
	src.Reset()
	data, err := json.Marshal(src)
	require.NoError(t, err)

	restored, err := RestoreMendikotInteractor(data, new(presenter.MockMendikotPresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
	assert.Equal(t, src.GetHandNumber(), restored.Game.GetHandNumber())
	assert.Equal(t, src.GetTrumpSuit(), restored.Game.GetTrumpSuit())
	assert.Equal(t, src.GetPlayer(0).GetCardsSize(), restored.Game.GetPlayer(0).GetCardsSize())

	_, err = RestoreMendikotInteractor([]byte("{"), new(presenter.MockMendikotPresenter))
	assert.Error(t, err)
}
