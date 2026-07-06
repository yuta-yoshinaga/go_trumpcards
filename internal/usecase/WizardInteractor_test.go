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

func TestNewWizardInteractor_NilGuards(t *testing.T) {
	opMock := new(presenter.MockWizardPresenter)

	t.Run("panics when o is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "WizardInteractor: o must not be nil", func() {
			usecase.NewWizardInteractor(nil, opMock)
		})
	})

	t.Run("panics when op is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockWizardGame)
		assert.PanicsWithValue(t, "WizardInteractor: op must not be nil", func() {
			usecase.NewWizardInteractor(gameMock, nil)
		})
	})
}

func TestWizardInteractor_Reset(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("reset stays in bid phase", func(t *testing.T) {
		opMock := new(presenter.MockWizardPresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockWizardGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.WizardPhaseBid)
		gameMock.On("IsHumanBidTurn").Return(true)

		oi := usecase.NewWizardInteractor(gameMock, opMock)
		result := oi.Reset()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "Reset")
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("reset transitions to play phase and runs CPU turns", func(t *testing.T) {
		opMock := new(presenter.MockWizardPresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockWizardGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.WizardPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		oi := usecase.NewWizardInteractor(gameMock, opMock)
		result := oi.Reset()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "Reset")
	})
}

func TestWizardInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("sets config then resets", func(t *testing.T) {
		opMock := new(presenter.MockWizardPresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockWizardGame)
		cfg := domain.WizardConfig{CpuDifficulty: domain.WizardCpuDifficultyHard}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.WizardPhaseBid)
		gameMock.On("IsHumanBidTurn").Return(true)

		oi := usecase.NewWizardInteractor(gameMock, opMock)
		result := oi.ResetWithConfig(cfg)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})
}

func TestWizardInteractor_ResetWithConfig_ValidationError(t *testing.T) {
	opMock := new(presenter.MockWizardPresenter)
	gameMock := new(interfaces.MockWizardGame)
	opMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")

	oi := usecase.NewWizardInteractor(gameMock, opMock)
	cfg := domain.WizardConfig{CpuDifficulty: domain.WizardCpuDifficulty(-1)}
	result := oi.ResetWithConfig(cfg)
	assert.Equal(t, "validation error", result)
	gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
}

func TestWizardInteractor_Bid(t *testing.T) {
	mockOutput := `{"phase":1}`

	t.Run("game ended returns output without bidding", func(t *testing.T) {
		opMock := new(presenter.MockWizardPresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockWizardGame)
		gameMock.On("GetGameEndFlag").Return(true)

		oi := usecase.NewWizardInteractor(gameMock, opMock)
		result := oi.Bid(3)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerBid")
	})

	t.Run("bid error returns error output", func(t *testing.T) {
		opMock := new(presenter.MockWizardPresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return("bid error")
		gameMock := new(interfaces.MockWizardGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerBid", 99).Return(errors.New("invalid bid"))

		oi := usecase.NewWizardInteractor(gameMock, opMock)
		result := oi.Bid(99)
		assert.Equal(t, "bid error", result)
	})

	t.Run("successful bid runs CPU bids", func(t *testing.T) {
		opMock := new(presenter.MockWizardPresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockWizardGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerBid", 3).Return(nil)
		gameMock.On("GetPhase").Return(domain.WizardPhaseBid)
		gameMock.On("IsHumanBidTurn").Return(true)

		oi := usecase.NewWizardInteractor(gameMock, opMock)
		result := oi.Bid(3)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerBid", 3)
	})
}

func TestWizardInteractor_Play(t *testing.T) {
	mockOutput := `{"phase":1}`

	t.Run("game ended returns output without playing", func(t *testing.T) {
		opMock := new(presenter.MockWizardPresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockWizardGame)
		gameMock.On("GetGameEndFlag").Return(true)

		oi := usecase.NewWizardInteractor(gameMock, opMock)
		result := oi.Play(0)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerPlay")
	})

	t.Run("not human turn returns output", func(t *testing.T) {
		opMock := new(presenter.MockWizardPresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockWizardGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		oi := usecase.NewWizardInteractor(gameMock, opMock)
		result := oi.Play(0)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerPlay")
	})

	t.Run("play error returns error output", func(t *testing.T) {
		opMock := new(presenter.MockWizardPresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return("play error")
		gameMock := new(interfaces.MockWizardGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerPlay", 99).Return(errors.New("invalid"))

		oi := usecase.NewWizardInteractor(gameMock, opMock)
		result := oi.Play(99)
		assert.Equal(t, "play error", result)
	})

	t.Run("successful play runs CPU turns", func(t *testing.T) {
		opMock := new(presenter.MockWizardPresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockWizardGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerPlay", 0).Return(nil)
		gameMock.On("GetPhase").Return(domain.WizardPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		oi := usecase.NewWizardInteractor(gameMock, opMock)
		result := oi.Play(0)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerPlay", 0)
		gameMock.AssertNotCalled(t, "ResolveTrick")
	})

	t.Run("human completes trick calls ResolveTrick", func(t *testing.T) {
		opMock := new(presenter.MockWizardPresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockWizardGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerPlay", 0).Return(nil)
		gameMock.On("GetPhase").Return(domain.WizardPhaseTrickEnd)
		gameMock.On("ResolveTrick").Return()

		oi := usecase.NewWizardInteractor(gameMock, opMock)
		result := oi.Play(0)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "ResolveTrick")
	})
}

func TestWizardInteractor_NextTrick(t *testing.T) {
	mockOutput := `{"phase":1}`

	opMock := new(presenter.MockWizardPresenter)
	opMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockWizardGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.WizardPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	oi := usecase.NewWizardInteractor(gameMock, opMock)
	result := oi.NextTrick()
	assert.Equal(t, mockOutput, result)
	gameMock.AssertCalled(t, "NextTrick")
}

func TestWizardInteractor_NextRound(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("scores round, then starts next", func(t *testing.T) {
		opMock := new(presenter.MockWizardPresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockWizardGame)
		gameMock.On("ScoreRound").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextRound").Return()
		gameMock.On("GetPhase").Return(domain.WizardPhaseBid)
		gameMock.On("IsHumanBidTurn").Return(true)

		oi := usecase.NewWizardInteractor(gameMock, opMock)
		result := oi.NextRound()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "ScoreRound")
		gameMock.AssertCalled(t, "NextRound")
	})

	t.Run("game ends after score", func(t *testing.T) {
		opMock := new(presenter.MockWizardPresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return(`{"gameEndFlag":true}`)
		gameMock := new(interfaces.MockWizardGame)
		gameMock.On("ScoreRound").Return()
		gameMock.On("GetGameEndFlag").Return(true)

		oi := usecase.NewWizardInteractor(gameMock, opMock)
		result := oi.NextRound()
		assert.Equal(t, `{"gameEndFlag":true}`, result)
		gameMock.AssertNotCalled(t, "NextRound")
	})
}

func TestWizardInteractor_GetConfig(t *testing.T) {
	opMock := new(presenter.MockWizardPresenter)
	gameMock := new(interfaces.MockWizardGame)
	cfg := domain.DefaultWizardConfig()
	gameMock.On("GetConfig").Return(cfg)

	oi := usecase.NewWizardInteractor(gameMock, opMock)
	assert.Equal(t, cfg, oi.GetConfig())
}

func TestWizardInteractor_Hint(t *testing.T) {
	opMock := new(presenter.MockWizardPresenter)
	gameMock := new(interfaces.MockWizardGame)
	opMock.On("HintOutput", gameMock).Return("hint result")

	oi := usecase.NewWizardInteractor(gameMock, opMock)
	assert.Equal(t, "hint result", oi.Hint())
}

func TestWizardInteractor_ActionLog(t *testing.T) {
	opMock := new(presenter.MockWizardPresenter)
	gameMock := new(interfaces.MockWizardGame)
	opMock.On("ActionLogOutput", gameMock).Return("log result")

	oi := usecase.NewWizardInteractor(gameMock, opMock)
	assert.Equal(t, "log result", oi.ActionLog())
}
