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

func newMockJulepeGame() *interfaces.MockJulepeGame { return new(interfaces.MockJulepeGame) }

func newMockJulepePresenter() *presenter.MockJulepePresenter {
	return new(presenter.MockJulepePresenter)
}

func TestNewJulepeInteractor(t *testing.T) {
	assert.NotNil(t, NewJulepeInteractor(newMockJulepeGame(), newMockJulepePresenter()))
}

func TestNewJulepeInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewJulepeInteractor(nil, newMockJulepePresenter()) })
	assert.Panics(t, func() { NewJulepeInteractor(newMockJulepeGame(), nil) })
}

// **Reset は CPU を動かさない。** 配り終えた時点は選択フェーズ。
func TestJulepeInteractorResetDoesNotRunCpu(t *testing.T) {
	g := newMockJulepeGame()
	p := newMockJulepePresenter()
	i := NewJulepeInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetHint").Return(nil).Maybe()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertCalled(t, "Reset")
	g.AssertNotCalled(t, "CpuPlay")
}

// Play / Pass はそれぞれ Decide(true) / Decide(false) に落ちる。取り違えたら
// 参加したはずのラウンドで降りることになる。
func TestJulepeInteractorPlayAndPassAreDistinct(t *testing.T) {
	for _, tc := range []struct {
		name string
		call func(*JulepeInteractor) string
		arg  bool
	}{
		{"play", func(i *JulepeInteractor) string { return i.Play() }, true},
		{"pass", func(i *JulepeInteractor) string { return i.Pass() }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockJulepeGame()
			p := newMockJulepePresenter()
			i := NewJulepeInteractor(g, p)

			g.On("GetGameEndFlag").Return(false)
			g.On("Decide", tc.arg).Return(nil)
			g.On("IsHumanTurn").Return(true)
			p.On("Output", g, nil).Return("decided")

			assert.Equal(t, "decided", tc.call(i))
			g.AssertCalled(t, "Decide", tc.arg)
			g.AssertNotCalled(t, "Decide", !tc.arg)
		})
	}
}

func TestJulepeInteractorDecideError(t *testing.T) {
	g := newMockJulepeGame()
	p := newMockJulepePresenter()
	i := NewJulepeInteractor(g, p)

	err := errors.New("not the decision phase")
	g.On("GetGameEndFlag").Return(false)
	g.On("Decide", true).Return(err)
	p.On("Output", g, err).Return("decide_error")

	assert.Equal(t, "decide_error", i.Play())
	g.AssertNotCalled(t, "CpuPlay")
}

// **人間が降りたラウンドも最後まで回す。** 手番が回ってこないので
// IsHumanTurn() は永久に false、そこで止めると盤面が進まない。
func TestJulepeInteractorPlaysOutARoundThePlayerDroppedFrom(t *testing.T) {
	g := newMockJulepeGame()
	p := newMockJulepePresenter()
	i := NewJulepeInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("Decide", false).Return(nil)
	g.On("IsHumanTurn").Return(false) // 降りているので永久に false
	g.On("GetPhase").Return(domain.JulepePhasePlay).Times(4)
	g.On("GetPhase").Return(domain.JulepePhaseRoundEnd)
	g.On("CpuPlay").Return()
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Pass())
	g.AssertNumberOfCalls(t, "CpuPlay", 4)
}

// ラウンド終了では止まる。
func TestJulepeInteractorStopsAtRoundEnd(t *testing.T) {
	g := newMockJulepeGame()
	p := newMockJulepePresenter()
	i := NewJulepeInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("Decide", true).Return(nil)
	g.On("IsHumanTurn").Return(false)
	g.On("GetPhase").Return(domain.JulepePhaseRoundEnd)
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Play())
	g.AssertNotCalled(t, "CpuPlay")
	g.AssertNotCalled(t, "NextRound")
}

func TestJulepeInteractorCpuLoopHasAnUpperBound(t *testing.T) {
	g := newMockJulepeGame()
	p := newMockJulepePresenter()
	i := NewJulepeInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("Decide", true).Return(nil)
	g.On("IsHumanTurn").Return(false)
	g.On("GetPhase").Return(domain.JulepePhasePlay)
	g.On("CpuPlay").Return()
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Play())
	g.AssertNumberOfCalls(t, "CpuPlay", maxCpuTurnsPerCall)
}

func TestJulepeInteractorPlayCard(t *testing.T) {
	g := newMockJulepeGame()
	p := newMockJulepePresenter()
	i := NewJulepeInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	g.On("PlayerPlay", 2).Return(nil)
	p.On("Output", g, nil).Return("played")

	assert.Equal(t, "played", i.PlayCard(2))
	g.AssertCalled(t, "PlayerPlay", 2)
}

func TestJulepeInteractorPlayCardError(t *testing.T) {
	g := newMockJulepeGame()
	p := newMockJulepePresenter()
	i := NewJulepeInteractor(g, p)

	err := errors.New("must follow suit")
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	g.On("PlayerPlay", 2).Return(err)
	p.On("Output", g, err).Return("error_output")

	assert.Equal(t, "error_output", i.PlayCard(2))
	g.AssertNotCalled(t, "CpuPlay")
}

func TestJulepeInteractorPlayCardBlocked(t *testing.T) {
	for _, tc := range []struct {
		name    string
		ended   bool
		isHuman bool
	}{
		{"game over", true, true},
		{"not human turn", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockJulepeGame()
			p := newMockJulepePresenter()
			i := NewJulepeInteractor(g, p)

			g.On("GetGameEndFlag").Return(tc.ended)
			g.On("IsHumanTurn").Return(tc.isHuman)
			p.On("Output", g, nil).Return("blocked")

			assert.Equal(t, "blocked", i.PlayCard(0))
			g.AssertNotCalled(t, "PlayerPlay", 0)
		})
	}
}

// **NextRound も CPU を回さない。** 次のラウンドは選択フェーズから。
func TestJulepeInteractorNextRound(t *testing.T) {
	g := newMockJulepeGame()
	p := newMockJulepePresenter()
	i := NewJulepeInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("NextRound").Return()
	p.On("Output", g, nil).Return("next")

	assert.Equal(t, "next", i.NextRound())
	g.AssertCalled(t, "NextRound")
	g.AssertNotCalled(t, "CpuPlay")
}

func TestJulepeInteractorGuardsAfterGameEnd(t *testing.T) {
	for _, tc := range []struct {
		name   string
		call   func(*JulepeInteractor) string
		method string
	}{
		{"next round", func(i *JulepeInteractor) string { return i.NextRound() }, "NextRound"},
		{"give up", func(i *JulepeInteractor) string { return i.GiveUp() }, "GiveUp"},
		{"play", func(i *JulepeInteractor) string { return i.Play() }, "Decide"},
		{"pass", func(i *JulepeInteractor) string { return i.Pass() }, "Decide"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockJulepeGame()
			p := newMockJulepePresenter()
			i := NewJulepeInteractor(g, p)

			g.On("GetGameEndFlag").Return(true)
			p.On("Output", g, nil).Return("over")

			assert.Equal(t, "over", tc.call(i))
			g.AssertNotCalled(t, tc.method, mock.Anything)
			g.AssertNotCalled(t, tc.method)
		})
	}
}

func TestJulepeInteractorGiveUp(t *testing.T) {
	g := newMockJulepeGame()
	p := newMockJulepePresenter()
	i := NewJulepeInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("gave_up")

	assert.Equal(t, "gave_up", i.GiveUp())
	g.AssertCalled(t, "GiveUp")
}

func TestJulepeInteractorGetConfig(t *testing.T) {
	g := newMockJulepeGame()
	i := NewJulepeInteractor(g, newMockJulepePresenter())
	cfg := domain.JulepeConfig{PlayerCnt: 5, Rounds: 6}
	g.On("GetConfig").Return(cfg)
	assert.Equal(t, cfg, i.GetConfig())
}

// 範囲外の人数は弾かれ、配り直されない。
func TestJulepeInteractorResetWithInvalidConfig(t *testing.T) {
	for _, cfg := range []domain.JulepeConfig{
		{PlayerCnt: 2, Rounds: 4},
		{PlayerCnt: 6, Rounds: 4},
		{PlayerCnt: 4, Rounds: 0},
	} {
		g := newMockJulepeGame()
		p := newMockJulepePresenter()
		i := NewJulepeInteractor(g, p)
		p.On("Output", mock.Anything, mock.Anything).Return("cfg_error")

		assert.Equal(t, "cfg_error", i.ResetWithConfig(cfg))
		g.AssertNotCalled(t, "Reset")
		g.AssertNotCalled(t, "SetConfig", mock.Anything)
	}
}

func TestJulepeInteractorHintAndActionLog(t *testing.T) {
	g := newMockJulepeGame()
	p := newMockJulepePresenter()
	i := NewJulepeInteractor(g, p)

	p.On("HintOutput", g).Return("hint_output")
	p.On("ActionLogOutput", g).Return("log_output")

	assert.Equal(t, "hint_output", i.Hint())
	assert.Equal(t, "log_output", i.ActionLog())
}

// Worker はリクエストごとに KV から作り直す。**参加状態が往復しないと、
// 降りたはずの人が参加者に戻る** (#4478)。
func TestJulepeInteractorSnapshotRoundTrip(t *testing.T) {
	g := domain.NewDefaultJulepe()
	g.Reset()
	require.NoError(t, g.Decide(false)) // 降りる
	g.GetPlayer(0).SetChips(70)

	p := newMockJulepePresenter()
	i := NewJulepeInteractor(g, p)

	data, err := i.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreJulepeInteractor(data, p)
	require.NoError(t, err)

	assert.Equal(t, 70, restored.Game.GetPlayer(0).GetChips())
	assert.False(t, restored.Game.GetPlayer(0).GetInRound(), "降りたまま復元される")
	assert.Equal(t, g.GetPot(), restored.Game.GetPot())
	assert.Equal(t, g.GetConfig().PlayerCnt, restored.Game.GetConfig().PlayerCnt)
}

func TestRestoreJulepeInteractorRejectsGarbage(t *testing.T) {
	_, err := RestoreJulepeInteractor([]byte("not json"), newMockJulepePresenter())
	assert.Error(t, err)
}
