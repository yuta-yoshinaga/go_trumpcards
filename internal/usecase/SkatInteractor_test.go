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

const skatMockOutput = `{"phase":0}`

// newSkatMocks builds presenter and game mocks with the most common stubs.
func newSkatMocks() (*presenter.MockSkatPresenter, *interfaces.MockSkatGame) {
	sp := new(presenter.MockSkatPresenter)
	sp.On("Output", mock.Anything, mock.Anything).Return(skatMockOutput)
	g := new(interfaces.MockSkatGame)
	return sp, g
}

func TestNewSkatInteractor_NilGuards(t *testing.T) {
	sp := new(presenter.MockSkatPresenter)
	t.Run("panics when game is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "SkatInteractor: s must not be nil", func() {
			usecase.NewSkatInteractor(nil, sp)
		})
	})
	t.Run("panics when presenter is nil", func(t *testing.T) {
		g := new(interfaces.MockSkatGame)
		assert.PanicsWithValue(t, "SkatInteractor: sp must not be nil", func() {
			usecase.NewSkatInteractor(g, nil)
		})
	})
}

func TestSkatInteractor_Reset(t *testing.T) {
	sp, g := newSkatMocks()
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.SkatPhaseBid)
	g.On("IsHumanBidTurn").Return(true)

	si := usecase.NewSkatInteractor(g, sp)
	got := si.Reset()
	assert.Equal(t, skatMockOutput, got)
	g.AssertCalled(t, "Reset")
}

func TestSkatInteractor_ResetWithConfig(t *testing.T) {
	t.Run("invalid config returns error output", func(t *testing.T) {
		sp, g := newSkatMocks()
		si := usecase.NewSkatInteractor(g, sp)
		got := si.ResetWithConfig(domain.SkatConfig{TargetScore: 0})
		assert.Equal(t, skatMockOutput, got)
		g.AssertNotCalled(t, "Reset")
	})

	t.Run("valid config triggers Reset", func(t *testing.T) {
		sp, g := newSkatMocks()
		cfg := domain.SkatConfig{CpuDifficulty: domain.SkatCpuDifficultyHard, TargetScore: 250}
		g.On("SetConfig", cfg).Return()
		g.On("Reset").Return()
		g.On("GetGameEndFlag").Return(false)
		g.On("GetPhase").Return(domain.SkatPhaseBid)
		g.On("IsHumanBidTurn").Return(true)
		si := usecase.NewSkatInteractor(g, sp)
		got := si.ResetWithConfig(cfg)
		assert.Equal(t, skatMockOutput, got)
		g.AssertCalled(t, "Reset")
	})
}

func TestSkatInteractor_BidGuards(t *testing.T) {
	t.Run("game ended short-circuits", func(t *testing.T) {
		sp, g := newSkatMocks()
		g.On("GetGameEndFlag").Return(true)
		si := usecase.NewSkatInteractor(g, sp)
		got := si.Bid(true)
		assert.Equal(t, skatMockOutput, got)
		g.AssertNotCalled(t, "PlayerBid", mock.Anything)
	})

	t.Run("PlayerBid error short-circuits CPU loop", func(t *testing.T) {
		sp, g := newSkatMocks()
		g.On("GetGameEndFlag").Return(false)
		g.On("PlayerBid", true).Return(errors.New("bad"))
		si := usecase.NewSkatInteractor(g, sp)
		got := si.Bid(true)
		assert.Equal(t, skatMockOutput, got)
		g.AssertNotCalled(t, "GetPhase")
	})

	t.Run("Bid succeeds and CPU phases run until human bid turn", func(t *testing.T) {
		sp, g := newSkatMocks()
		g.On("GetGameEndFlag").Return(false)
		g.On("PlayerBid", true).Return(nil)
		g.On("GetPhase").Return(domain.SkatPhaseBid)
		g.On("IsHumanBidTurn").Return(true)
		si := usecase.NewSkatInteractor(g, sp)
		got := si.Bid(true)
		assert.Equal(t, skatMockOutput, got)
		g.AssertCalled(t, "PlayerBid", true)
	})
}

func TestSkatInteractor_PickSkatDiscardDeclare(t *testing.T) {
	t.Run("PickSkat success", func(t *testing.T) {
		sp, g := newSkatMocks()
		g.On("GetGameEndFlag").Return(false)
		g.On("PlayerPickSkat", true).Return(nil)
		g.On("GetPhase").Return(domain.SkatPhaseDiscard)
		g.On("IsHumanDeclarerTurn").Return(true)
		si := usecase.NewSkatInteractor(g, sp)
		assert.Equal(t, skatMockOutput, si.PickSkat(true))
		g.AssertCalled(t, "PlayerPickSkat", true)
	})

	t.Run("Discard success", func(t *testing.T) {
		sp, g := newSkatMocks()
		g.On("GetGameEndFlag").Return(false)
		g.On("PlayerDiscard", 0, 1).Return(nil)
		g.On("GetPhase").Return(domain.SkatPhaseGameDeclaration)
		g.On("IsHumanDeclarerTurn").Return(true)
		si := usecase.NewSkatInteractor(g, sp)
		assert.Equal(t, skatMockOutput, si.Discard(0, 1))
		g.AssertCalled(t, "PlayerDiscard", 0, 1)
	})

	t.Run("DeclareGame success", func(t *testing.T) {
		sp, g := newSkatMocks()
		g.On("GetGameEndFlag").Return(false)
		g.On("PlayerDeclareGame", domain.SkatGameSuit, domain.CardDesignSpade).Return(nil)
		g.On("GetPhase").Return(domain.SkatPhasePlay)
		g.On("IsHumanTurn").Return(true)
		si := usecase.NewSkatInteractor(g, sp)
		assert.Equal(t, skatMockOutput, si.DeclareGame(domain.SkatGameSuit, domain.CardDesignSpade))
		g.AssertCalled(t, "PlayerDeclareGame", domain.SkatGameSuit, domain.CardDesignSpade)
	})

	t.Run("PickSkat error returned", func(t *testing.T) {
		sp, g := newSkatMocks()
		g.On("GetGameEndFlag").Return(false)
		g.On("PlayerPickSkat", false).Return(errors.New("bad phase"))
		si := usecase.NewSkatInteractor(g, sp)
		assert.Equal(t, skatMockOutput, si.PickSkat(false))
	})
}

func TestSkatInteractor_PlayAndNextTrickRound(t *testing.T) {
	t.Run("Play guard blocks when game ended", func(t *testing.T) {
		sp, g := newSkatMocks()
		g.On("GetGameEndFlag").Return(true)
		g.On("IsHumanTurn").Return(false)
		si := usecase.NewSkatInteractor(g, sp)
		assert.Equal(t, skatMockOutput, si.Play(0))
		g.AssertNotCalled(t, "PlayerPlay", mock.Anything)
	})

	t.Run("Play success drives CPU turns", func(t *testing.T) {
		sp, g := newSkatMocks()
		g.On("GetGameEndFlag").Return(false)
		g.On("IsHumanTurn").Return(true)
		g.On("PlayerPlay", 1).Return(nil)
		g.On("GetPhase").Return(domain.SkatPhasePlay)
		si := usecase.NewSkatInteractor(g, sp)
		assert.Equal(t, skatMockOutput, si.Play(1))
		g.AssertCalled(t, "PlayerPlay", 1)
	})

	t.Run("NextTrick", func(t *testing.T) {
		sp, g := newSkatMocks()
		g.On("GetGameEndFlag").Return(false)
		g.On("NextTrick").Return()
		g.On("GetPhase").Return(domain.SkatPhasePlay)
		g.On("IsHumanTurn").Return(true)
		si := usecase.NewSkatInteractor(g, sp)
		assert.Equal(t, skatMockOutput, si.NextTrick())
	})

	t.Run("NextRound", func(t *testing.T) {
		sp, g := newSkatMocks()
		g.On("ScoreRound").Return()
		g.On("GetGameEndFlag").Return(false)
		g.On("NextRound").Return()
		g.On("GetPhase").Return(domain.SkatPhaseBid)
		g.On("IsHumanBidTurn").Return(true)
		si := usecase.NewSkatInteractor(g, sp)
		assert.Equal(t, skatMockOutput, si.NextRound())
	})
}

func TestSkatInteractor_HintAndActionLogAndConfig(t *testing.T) {
	sp, g := newSkatMocks()
	cfg := domain.DefaultSkatConfig()
	g.On("GetConfig").Return(cfg)
	sp.On("HintOutput", mock.Anything).Return("hint")
	sp.On("ActionLogOutput", mock.Anything).Return("log")
	si := usecase.NewSkatInteractor(g, sp)
	assert.Equal(t, cfg, si.GetConfig())
	assert.Equal(t, "hint", si.Hint())
	assert.Equal(t, "log", si.ActionLog())
}

func TestSkatInteractor_RestoreFromJSON(t *testing.T) {
	src := domain.NewDefaultSkat()
	src.Reset()
	sp := new(presenter.MockSkatPresenter)
	tempInteractor := usecase.NewSkatInteractor(src, sp)
	data, err := tempInteractor.Snapshot()
	assert.NoError(t, err)
	si, err := usecase.RestoreSkatInteractor(data, sp)
	assert.NoError(t, err)
	assert.NotNil(t, si)
}
