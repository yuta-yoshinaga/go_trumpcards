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

func newMockRamsGame() *interfaces.MockRamsGame { return new(interfaces.MockRamsGame) }

func newMockRamsPresenter() *presenter.MockRamsPresenter { return new(presenter.MockRamsPresenter) }

func TestNewRamsInteractor(t *testing.T) {
	assert.NotNil(t, NewRamsInteractor(newMockRamsGame(), newMockRamsPresenter()))
}

func TestNewRamsInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewRamsInteractor(nil, newMockRamsPresenter()) })
	assert.Panics(t, func() { NewRamsInteractor(newMockRamsGame(), nil) })
}

// **Reset は CPU を動かさない。** 配り終えた時点は選択フェーズ。
func TestRamsInteractorResetDoesNotRunCpu(t *testing.T) {
	g := newMockRamsGame()
	p := newMockRamsPresenter()
	i := NewRamsInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetHint").Return(nil).Maybe()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertCalled(t, "Reset")
	g.AssertNotCalled(t, "CpuPlay")
}

// Play / Pass はそれぞれ Decide(true) / Decide(false) に落ちる。取り違えたら
// 参加したはずのラウンドで降りることになる。
func TestRamsInteractorPlayAndPassAreDistinct(t *testing.T) {
	for _, tc := range []struct {
		name string
		call func(*RamsInteractor) string
		arg  bool
	}{
		{"play", func(i *RamsInteractor) string { return i.Play() }, true},
		{"pass", func(i *RamsInteractor) string { return i.Pass() }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockRamsGame()
			p := newMockRamsPresenter()
			i := NewRamsInteractor(g, p)

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

func TestRamsInteractorDecideError(t *testing.T) {
	g := newMockRamsGame()
	p := newMockRamsPresenter()
	i := NewRamsInteractor(g, p)

	err := errors.New("not the decision phase")
	g.On("GetGameEndFlag").Return(false)
	g.On("Decide", true).Return(err)
	p.On("Output", g, err).Return("decide_error")

	assert.Equal(t, "decide_error", i.Play())
	g.AssertNotCalled(t, "CpuPlay")
}

// **人間が降りたラウンドも最後まで回す。** 手番が回ってこないので
// IsHumanTurn() は永久に false、そこで止めると盤面が進まない。
func TestRamsInteractorPlaysOutARoundThePlayerDroppedFrom(t *testing.T) {
	g := newMockRamsGame()
	p := newMockRamsPresenter()
	i := NewRamsInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("Decide", false).Return(nil)
	g.On("IsHumanTurn").Return(false) // 降りているので永久に false
	g.On("GetPhase").Return(domain.RamsPhasePlay).Times(4)
	g.On("GetPhase").Return(domain.RamsPhaseRoundEnd)
	g.On("CpuPlay").Return()
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Pass())
	g.AssertNumberOfCalls(t, "CpuPlay", 4)
}

// ラウンド終了では止まる。
func TestRamsInteractorStopsAtRoundEnd(t *testing.T) {
	g := newMockRamsGame()
	p := newMockRamsPresenter()
	i := NewRamsInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("Decide", true).Return(nil)
	g.On("IsHumanTurn").Return(false)
	g.On("GetPhase").Return(domain.RamsPhaseRoundEnd)
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Play())
	g.AssertNotCalled(t, "CpuPlay")
	g.AssertNotCalled(t, "NextRound")
}

func TestRamsInteractorCpuLoopHasAnUpperBound(t *testing.T) {
	g := newMockRamsGame()
	p := newMockRamsPresenter()
	i := NewRamsInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("Decide", true).Return(nil)
	g.On("IsHumanTurn").Return(false)
	g.On("GetPhase").Return(domain.RamsPhasePlay)
	g.On("CpuPlay").Return()
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Play())
	g.AssertNumberOfCalls(t, "CpuPlay", maxCpuTurnsPerCall)
}

func TestRamsInteractorPlayCard(t *testing.T) {
	g := newMockRamsGame()
	p := newMockRamsPresenter()
	i := NewRamsInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	g.On("PlayerPlay", 2).Return(nil)
	p.On("Output", g, nil).Return("played")

	assert.Equal(t, "played", i.PlayCard(2))
	g.AssertCalled(t, "PlayerPlay", 2)
}

func TestRamsInteractorPlayCardError(t *testing.T) {
	g := newMockRamsGame()
	p := newMockRamsPresenter()
	i := NewRamsInteractor(g, p)

	err := errors.New("must follow suit")
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	g.On("PlayerPlay", 2).Return(err)
	p.On("Output", g, err).Return("error_output")

	assert.Equal(t, "error_output", i.PlayCard(2))
	g.AssertNotCalled(t, "CpuPlay")
}

func TestRamsInteractorPlayCardBlocked(t *testing.T) {
	for _, tc := range []struct {
		name    string
		ended   bool
		isHuman bool
	}{
		{"game over", true, true},
		{"not human turn", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockRamsGame()
			p := newMockRamsPresenter()
			i := NewRamsInteractor(g, p)

			g.On("GetGameEndFlag").Return(tc.ended)
			g.On("IsHumanTurn").Return(tc.isHuman)
			p.On("Output", g, nil).Return("blocked")

			assert.Equal(t, "blocked", i.PlayCard(0))
			g.AssertNotCalled(t, "PlayerPlay", 0)
		})
	}
}

// **NextRound も CPU を回さない。** 次のラウンドは選択フェーズから。
func TestRamsInteractorNextRound(t *testing.T) {
	g := newMockRamsGame()
	p := newMockRamsPresenter()
	i := NewRamsInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("NextRound").Return()
	p.On("Output", g, nil).Return("next")

	assert.Equal(t, "next", i.NextRound())
	g.AssertCalled(t, "NextRound")
	g.AssertNotCalled(t, "CpuPlay")
}

func TestRamsInteractorGuardsAfterGameEnd(t *testing.T) {
	for _, tc := range []struct {
		name   string
		call   func(*RamsInteractor) string
		method string
	}{
		{"next round", func(i *RamsInteractor) string { return i.NextRound() }, "NextRound"},
		{"give up", func(i *RamsInteractor) string { return i.GiveUp() }, "GiveUp"},
		{"play", func(i *RamsInteractor) string { return i.Play() }, "Decide"},
		{"pass", func(i *RamsInteractor) string { return i.Pass() }, "Decide"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockRamsGame()
			p := newMockRamsPresenter()
			i := NewRamsInteractor(g, p)

			g.On("GetGameEndFlag").Return(true)
			p.On("Output", g, nil).Return("over")

			assert.Equal(t, "over", tc.call(i))
			g.AssertNotCalled(t, tc.method, mock.Anything)
			g.AssertNotCalled(t, tc.method)
		})
	}
}

func TestRamsInteractorGiveUp(t *testing.T) {
	g := newMockRamsGame()
	p := newMockRamsPresenter()
	i := NewRamsInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("gave_up")

	assert.Equal(t, "gave_up", i.GiveUp())
	g.AssertCalled(t, "GiveUp")
}

func TestRamsInteractorGetConfig(t *testing.T) {
	g := newMockRamsGame()
	i := NewRamsInteractor(g, newMockRamsPresenter())
	cfg := domain.RamsConfig{PlayerCnt: 5, Rounds: 6}
	g.On("GetConfig").Return(cfg)
	assert.Equal(t, cfg, i.GetConfig())
}

// 範囲外の人数は弾かれ、配り直されない。
func TestRamsInteractorResetWithInvalidConfig(t *testing.T) {
	for _, cfg := range []domain.RamsConfig{
		{PlayerCnt: 2, Rounds: 4},
		{PlayerCnt: 6, Rounds: 4},
		{PlayerCnt: 4, Rounds: 0},
	} {
		g := newMockRamsGame()
		p := newMockRamsPresenter()
		i := NewRamsInteractor(g, p)
		p.On("Output", mock.Anything, mock.Anything).Return("cfg_error")

		assert.Equal(t, "cfg_error", i.ResetWithConfig(cfg))
		g.AssertNotCalled(t, "Reset")
		g.AssertNotCalled(t, "SetConfig", mock.Anything)
	}
}

func TestRamsInteractorHintAndActionLog(t *testing.T) {
	g := newMockRamsGame()
	p := newMockRamsPresenter()
	i := NewRamsInteractor(g, p)

	p.On("HintOutput", g).Return("hint_output")
	p.On("ActionLogOutput", g).Return("log_output")

	assert.Equal(t, "hint_output", i.Hint())
	assert.Equal(t, "log_output", i.ActionLog())
}

// Worker はリクエストごとに KV から作り直す。**参加状態が往復しないと、
// 降りたはずの人が参加者に戻る** (#4478)。
func TestRamsInteractorSnapshotRoundTrip(t *testing.T) {
	g := domain.NewDefaultRams()
	g.Reset()
	require.NoError(t, g.Decide(false)) // 降りる
	g.GetPlayer(0).SetChips(70)

	p := newMockRamsPresenter()
	i := NewRamsInteractor(g, p)

	data, err := i.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreRamsInteractor(data, p)
	require.NoError(t, err)

	assert.Equal(t, 70, restored.Game.GetPlayer(0).GetChips())
	assert.False(t, restored.Game.GetPlayer(0).GetInRound(), "降りたまま復元される")
	assert.Equal(t, g.GetPot(), restored.Game.GetPot())
	assert.Equal(t, g.GetConfig().PlayerCnt, restored.Game.GetConfig().PlayerCnt)
}

func TestRestoreRamsInteractorRejectsGarbage(t *testing.T) {
	_, err := RestoreRamsInteractor([]byte("not json"), newMockRamsPresenter())
	assert.Error(t, err)
}
