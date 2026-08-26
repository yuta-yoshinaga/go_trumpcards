//go:build test

package usecase_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

const gleekMockOutput = `{"phase":0}`

// newGleekMocks wires a presenter and a game mock that sit still in the play
// phase, so a test only has to describe the call it cares about.
func newGleekMocks() (*presenter.MockGleekPresenter, *interfaces.MockGleekGame) {
	tp := new(presenter.MockGleekPresenter)
	tp.On("Output", mock.Anything, mock.Anything).Return(gleekMockOutput)
	g := new(interfaces.MockGleekGame)
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.GleekPhasePlay)
	g.On("IsHumanTurn").Return(true)
	g.On("IsHumanBidTurn").Return(true)
	g.On("IsHumanDiscardTurn").Return(true)
	return tp, g
}

func TestGleekInteractor_Bid(t *testing.T) {
	tp, g := newGleekMocks()
	g.On("PlayerBid", 14).Return(nil)

	ci := usecase.NewGleekInteractor(g, tp)
	assert.Equal(t, gleekMockOutput, ci.Bid(14))
	g.AssertCalled(t, "PlayerBid", 14)
}

func TestGleekInteractor_BidError(t *testing.T) {
	tp, g := newGleekMocks()
	g.On("PlayerBid", 13).Return(errors.New("bad amount"))

	ci := usecase.NewGleekInteractor(g, tp)
	assert.Equal(t, gleekMockOutput, ci.Bid(13))
	// 失敗したら盤は進めない。
	g.AssertNotCalled(t, "CpuPlay")
}

func TestGleekInteractor_BidGameEnded(t *testing.T) {
	tp := new(presenter.MockGleekPresenter)
	tp.On("Output", mock.Anything, mock.Anything).Return(gleekMockOutput)
	g := new(interfaces.MockGleekGame)
	g.On("GetGameEndFlag").Return(true)

	ci := usecase.NewGleekInteractor(g, tp)
	assert.Equal(t, gleekMockOutput, ci.Bid(14))
	g.AssertNotCalled(t, "PlayerBid", mock.Anything)
}

// **捨て札フェーズを抜ける唯一の入口。** ここが動かないと落札の直後で盤が
// 固まり、play は「フェーズが違う」で弾かれ続ける。
func TestGleekInteractor_Discard(t *testing.T) {
	tp, g := newGleekMocks()
	indices := []int{0, 1, 2, 3, 4, 5, 6}
	g.On("PlayerDiscard", indices).Return(nil)

	ci := usecase.NewGleekInteractor(g, tp)
	assert.Equal(t, gleekMockOutput, ci.Discard(indices))
	g.AssertCalled(t, "PlayerDiscard", indices)
}

func TestGleekInteractor_DiscardError(t *testing.T) {
	tp, g := newGleekMocks()
	g.On("PlayerDiscard", []int{0}).Return(errors.New("wrong count"))

	ci := usecase.NewGleekInteractor(g, tp)
	assert.Equal(t, gleekMockOutput, ci.Discard([]int{0}))
	g.AssertNotCalled(t, "CpuPlay")
}

func TestGleekInteractor_DiscardGameEnded(t *testing.T) {
	tp := new(presenter.MockGleekPresenter)
	tp.On("Output", mock.Anything, mock.Anything).Return(gleekMockOutput)
	g := new(interfaces.MockGleekGame)
	g.On("GetGameEndFlag").Return(true)

	ci := usecase.NewGleekInteractor(g, tp)
	assert.Equal(t, gleekMockOutput, ci.Discard([]int{0}))
	g.AssertNotCalled(t, "PlayerDiscard", mock.Anything)
}

func TestGleekInteractor_Play(t *testing.T) {
	tp, g := newGleekMocks()
	g.On("PlayerPlay", 2).Return(nil)

	ci := usecase.NewGleekInteractor(g, tp)
	assert.Equal(t, gleekMockOutput, ci.Play(2))
	g.AssertCalled(t, "PlayerPlay", 2)
}

// **人間が最後の札を出したトリックはその場で解決する。** 解決しないと、
// 3 枚そろっているのに誰も勝者を決めないまま止まる。
func TestGleekInteractor_PlayResolvesTheCompletedTrick(t *testing.T) {
	tp := new(presenter.MockGleekPresenter)
	tp.On("Output", mock.Anything, mock.Anything).Return(gleekMockOutput)
	g := new(interfaces.MockGleekGame)
	g.On("GetGameEndFlag").Return(false)
	g.On("IsHumanTurn").Return(true)
	g.On("IsHumanBidTurn").Return(false)
	g.On("IsHumanDiscardTurn").Return(false)
	g.On("PlayerPlay", 0).Return(nil)
	g.On("GetPhase").Return(domain.GleekPhaseTrickEnd)
	g.On("ResolveTrick").Return()

	ci := usecase.NewGleekInteractor(g, tp)
	assert.Equal(t, gleekMockOutput, ci.Play(0))
	g.AssertCalled(t, "ResolveTrick")
}

func TestGleekInteractor_PlayError(t *testing.T) {
	tp, g := newGleekMocks()
	g.On("PlayerPlay", 9).Return(errors.New("illegal"))

	ci := usecase.NewGleekInteractor(g, tp)
	assert.Equal(t, gleekMockOutput, ci.Play(9))
	g.AssertNotCalled(t, "ResolveTrick")
}

func TestGleekInteractor_NextTrickAndNextRound(t *testing.T) {
	tp, g := newGleekMocks()
	g.On("NextTrick").Return()
	g.On("ScoreRound").Return()
	g.On("NextRound").Return()

	ci := usecase.NewGleekInteractor(g, tp)
	assert.Equal(t, gleekMockOutput, ci.NextTrick())
	g.AssertCalled(t, "NextTrick")
	assert.Equal(t, gleekMockOutput, ci.NextRound())
	g.AssertCalled(t, "ScoreRound")
	g.AssertCalled(t, "NextRound")
}

// **終わったゲームでは次のディールを配らない。** ScoreRound だけ走って止まる。
func TestGleekInteractor_NextRoundStopsAtGameEnd(t *testing.T) {
	tp := new(presenter.MockGleekPresenter)
	tp.On("Output", mock.Anything, mock.Anything).Return(gleekMockOutput)
	g := new(interfaces.MockGleekGame)
	g.On("ScoreRound").Return()
	g.On("GetGameEndFlag").Return(true)

	ci := usecase.NewGleekInteractor(g, tp)
	assert.Equal(t, gleekMockOutput, ci.NextRound())
	g.AssertCalled(t, "ScoreRound")
	g.AssertNotCalled(t, "NextRound")
}

func TestGleekInteractor_ResetAppliesTheConfig(t *testing.T) {
	tp, g := newGleekMocks()
	cfg := domain.DefaultGleekConfig()
	cfg.CpuDifficulty = domain.GleekCpuDifficultyHard
	g.On("SetConfig", cfg).Return()
	g.On("Reset").Return()
	g.On("GetConfig").Return(cfg)

	ci := usecase.NewGleekInteractor(g, tp)
	assert.Equal(t, gleekMockOutput, ci.ResetWithConfig(cfg))
	g.AssertCalled(t, "SetConfig", cfg)
	g.AssertCalled(t, "Reset")
	assert.Equal(t, cfg, ci.GetConfig())
}

func TestGleekInteractor_HintAndActionLog(t *testing.T) {
	tp, g := newGleekMocks()
	tp.On("HintOutput", mock.Anything).Return("hint")
	tp.On("ActionLogOutput", mock.Anything).Return("log")

	ci := usecase.NewGleekInteractor(g, tp)
	assert.Equal(t, "hint", ci.Hint())
	assert.Equal(t, "log", ci.ActionLog())
}
