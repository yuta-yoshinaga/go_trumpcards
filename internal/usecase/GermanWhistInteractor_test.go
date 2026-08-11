package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockGermanWhistGame() *interfaces.MockGermanWhistGame {
	return new(interfaces.MockGermanWhistGame)
}

func newMockGermanWhistPresenter() *presenter.MockGermanWhistPresenter {
	return new(presenter.MockGermanWhistPresenter)
}

func TestNewGermanWhistInteractor(t *testing.T) {
	assert.NotNil(t, NewGermanWhistInteractor(newMockGermanWhistGame(), newMockGermanWhistPresenter()))
}

func TestNewGermanWhistInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewGermanWhistInteractor(nil, newMockGermanWhistPresenter()) })
	assert.Panics(t, func() { NewGermanWhistInteractor(newMockGermanWhistGame(), nil) })
}

func TestGermanWhistInteractorReset(t *testing.T) {
	g := newMockGermanWhistGame()
	p := newMockGermanWhistPresenter()
	i := NewGermanWhistInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertCalled(t, "Reset")
	g.AssertNotCalled(t, "CpuPlay")
}

// 配り手が CPU 側なら、人間の手番になるまで CPU が打つ。
func TestGermanWhistInteractorResetRunsCpuUntilHumanTurn(t *testing.T) {
	g := newMockGermanWhistGame()
	p := newMockGermanWhistPresenter()
	i := NewGermanWhistInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(false).Twice()
	g.On("IsHumanTurn").Return(true)
	g.On("CpuPlay").Return()
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNumberOfCalls(t, "CpuPlay", 2)
}

// 終局していれば CPU は一切打たない。
func TestGermanWhistInteractorCpuLoopStopsAtGameEnd(t *testing.T) {
	g := newMockGermanWhistGame()
	p := newMockGermanWhistPresenter()
	i := NewGermanWhistInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(true)
	g.On("IsHumanTurn").Return(false)
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNotCalled(t, "CpuPlay")
}

// **進まない CpuPlay でハングしない。**盤面が変わらないまま呼ばれ続けても
// maxCpuTurnsPerCall で必ず抜ける (#4607 と同じ防御)。
func TestGermanWhistInteractorCpuLoopHasAnUpperBound(t *testing.T) {
	g := newMockGermanWhistGame()
	p := newMockGermanWhistPresenter()
	i := NewGermanWhistInteractor(g, p)

	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(false) // 永久に人間の手番にならない
	g.On("CpuPlay").Return()
	p.On("Output", g, nil).Return("out")

	assert.Equal(t, "out", i.Reset())
	g.AssertNumberOfCalls(t, "CpuPlay", maxCpuTurnsPerCall)
}

func TestGermanWhistInteractorPlay(t *testing.T) {
	g := newMockGermanWhistGame()
	p := newMockGermanWhistPresenter()
	i := NewGermanWhistInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	g.On("PlayerPlay", 2).Return(nil)
	p.On("Output", g, nil).Return("played")

	assert.Equal(t, "played", i.Play(2))
	g.AssertCalled(t, "PlayerPlay", 2)
}

// ドメインが返したエラーはプレゼンターに渡る。
func TestGermanWhistInteractorPlayError(t *testing.T) {
	g := newMockGermanWhistGame()
	p := newMockGermanWhistPresenter()
	i := NewGermanWhistInteractor(g, p)

	err := errors.New("must follow suit")
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	g.On("PlayerPlay", 2).Return(err)
	p.On("Output", g, err).Return("error_output")

	assert.Equal(t, "error_output", i.Play(2))
	g.AssertNotCalled(t, "CpuPlay")
}

// 人間の手番でなければ PlayerPlay まで到達しない。
func TestGermanWhistInteractorPlayBlockedWhenNotHumanTurn(t *testing.T) {
	for _, tc := range []struct {
		name    string
		ended   bool
		isHuman bool
	}{
		{"game over", true, true},
		{"cpu's turn", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newMockGermanWhistGame()
			p := newMockGermanWhistPresenter()
			i := NewGermanWhistInteractor(g, p)

			g.On("GetGameEndFlag").Return(tc.ended)
			g.On("IsHumanTurn").Return(tc.isHuman)
			p.On("Output", g, nil).Return("blocked")

			assert.Equal(t, "blocked", i.Play(0))
			g.AssertNotCalled(t, "PlayerPlay", 0)
		})
	}
}

func TestGermanWhistInteractorGiveUp(t *testing.T) {
	g := newMockGermanWhistGame()
	p := newMockGermanWhistPresenter()
	i := NewGermanWhistInteractor(g, p)

	g.On("GetGameEndFlag").Return(false)
	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("gave_up")

	assert.Equal(t, "gave_up", i.GiveUp())
	g.AssertCalled(t, "GiveUp")
}

func TestGermanWhistInteractorGiveUpAfterGameEnd(t *testing.T) {
	g := newMockGermanWhistGame()
	p := newMockGermanWhistPresenter()
	i := NewGermanWhistInteractor(g, p)

	g.On("GetGameEndFlag").Return(true)
	p.On("Output", g, nil).Return("already_over")

	assert.Equal(t, "already_over", i.GiveUp())
	g.AssertNotCalled(t, "GiveUp")
}

func TestGermanWhistInteractorHintAndActionLog(t *testing.T) {
	g := newMockGermanWhistGame()
	p := newMockGermanWhistPresenter()
	i := NewGermanWhistInteractor(g, p)

	p.On("HintOutput", g).Return("hint_output")
	p.On("ActionLogOutput", g).Return("log_output")

	assert.Equal(t, "hint_output", i.Hint())
	assert.Equal(t, "log_output", i.ActionLog())
}

// Worker はリクエストごとに KV から状態を作り直す。**後半のトリック数が
// 往復しないと、得点がハンド途中で黙って 0 に戻る** (#4478)。
func TestGermanWhistInteractorSnapshotRoundTrip(t *testing.T) {
	g := domain.NewDefaultGermanWhist()
	g.Reset()
	g.GetPlayer(0).SetScoringTricks(5)
	g.GetPlayer(1).SetScoringTricks(3)
	g.SetPhase(domain.GermanWhistPhaseScoring)

	p := newMockGermanWhistPresenter()
	i := NewGermanWhistInteractor(g, p)

	data, err := i.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreGermanWhistInteractor(data, p)
	require.NoError(t, err)

	rg := restored.Game
	assert.Equal(t, 5, rg.GetPlayer(0).GetScoringTricks(), "後半の得点が往復する")
	assert.Equal(t, 3, rg.GetPlayer(1).GetScoringTricks())
	assert.Equal(t, domain.GermanWhistPhaseScoring, rg.GetPhase())
	assert.Equal(t, g.GetTrumpSuit(), rg.GetTrumpSuit(), "切り札が往復する")
	assert.Equal(t, g.GetStockCount(), rg.GetStockCount(), "山札の残りが往復する")
	assert.Equal(t, g.GetPlayer(0).GetCardsSize(), rg.GetPlayer(0).GetCardsSize())

	// 表向きの札も往復する。ここが落ちると前半の駆け引きが成立しない。
	if up := g.GetUpCard(); up != nil {
		require.NotNil(t, rg.GetUpCard())
		assert.Equal(t, up.GetDesign(), rg.GetUpCard().GetDesign())
		assert.Equal(t, up.GetValue(), rg.GetUpCard().GetValue())
	}
}

func TestRestoreGermanWhistInteractorRejectsGarbage(t *testing.T) {
	_, err := RestoreGermanWhistInteractor([]byte("not json"), newMockGermanWhistPresenter())
	assert.Error(t, err)
}
