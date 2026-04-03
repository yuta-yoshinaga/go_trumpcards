//go:build test

package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func TestNewPinochleInteractor_NilGuards(t *testing.T) {
	ppMock := new(presenter.MockPinochlePresenter)

	t.Run("panics when p is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "PinochleInteractor: p must not be nil", func() {
			usecase.NewPinochleInteractor(nil, ppMock)
		})
	})

	t.Run("panics when pp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockPinochleGame)
		assert.PanicsWithValue(t, "PinochleInteractor: pp must not be nil", func() {
			usecase.NewPinochleInteractor(gameMock, nil)
		})
	})
}

func setupPinochleInteractorMocks(phase domain.PinochlePhase) (*interfaces.MockPinochleGame, *presenter.MockPinochlePresenter) {
	ppMock := new(presenter.MockPinochlePresenter)
	ppMock.On("Output", mock.Anything, mock.Anything).Return(`{"phase":0}`)
	gameMock := new(interfaces.MockPinochleGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(phase)
	gameMock.On("IsHumanBidTurn").Return(true)
	gameMock.On("IsHumanTurn").Return(true)
	return gameMock, ppMock
}

func TestPinochleInteractor_Reset(t *testing.T) {
	gameMock, ppMock := setupPinochleInteractorMocks(domain.PinochlePhaseBid)
	gameMock.On("Reset").Return()

	pi := usecase.NewPinochleInteractor(gameMock, ppMock)
	result := pi.Reset()
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "Reset")
}

func TestPinochleInteractor_ResetWithConfig(t *testing.T) {
	t.Run("sets config then resets", func(t *testing.T) {
		gameMock, ppMock := setupPinochleInteractorMocks(domain.PinochlePhaseBid)
		cfg := domain.PinochleConfig{CpuDifficulty: domain.PinochleCpuDifficultyHard, PointLimit: 500}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()

		pi := usecase.NewPinochleInteractor(gameMock, ppMock)
		result := pi.ResetWithConfig(cfg)
		assert.Contains(t, result, "phase")
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config returns error", func(t *testing.T) {
		gameMock, ppMock := setupPinochleInteractorMocks(domain.PinochlePhaseBid)
		invalidCfg := domain.PinochleConfig{CpuDifficulty: 99, PointLimit: 10}

		pi := usecase.NewPinochleInteractor(gameMock, ppMock)
		result := pi.ResetWithConfig(invalidCfg)
		assert.Contains(t, result, "phase")
	})
}

func TestPinochleInteractor_Bid(t *testing.T) {
	gameMock, ppMock := setupPinochleInteractorMocks(domain.PinochlePhaseBid)
	gameMock.On("PlayerBid", 25).Return(nil)

	pi := usecase.NewPinochleInteractor(gameMock, ppMock)
	result := pi.Bid(25)
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "PlayerBid", 25)
}

func TestPinochleInteractor_Bid_Error(t *testing.T) {
	gameMock, ppMock := setupPinochleInteractorMocks(domain.PinochlePhaseBid)
	gameMock.On("PlayerBid", 5).Return(domain.NewDomainError(domain.ErrInvalidAmount, "too low"))

	pi := usecase.NewPinochleInteractor(gameMock, ppMock)
	result := pi.Bid(5)
	assert.Contains(t, result, "phase")
}

func TestPinochleInteractor_Pass(t *testing.T) {
	gameMock, ppMock := setupPinochleInteractorMocks(domain.PinochlePhaseBid)
	gameMock.On("PlayerPass").Return(nil)

	pi := usecase.NewPinochleInteractor(gameMock, ppMock)
	result := pi.Pass()
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "PlayerPass")
}

func TestPinochleInteractor_CallTrump(t *testing.T) {
	gameMock, ppMock := setupPinochleInteractorMocks(domain.PinochlePhaseTrump)
	gameMock.On("PlayerCallTrump", 3).Return(nil)

	pi := usecase.NewPinochleInteractor(gameMock, ppMock)
	result := pi.CallTrump(3)
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "PlayerCallTrump", 3)
}

func TestPinochleInteractor_ConfirmMelds(t *testing.T) {
	gameMock, ppMock := setupPinochleInteractorMocks(domain.PinochlePhaseMeld)
	gameMock.On("ConfirmMelds").Return()

	pi := usecase.NewPinochleInteractor(gameMock, ppMock)
	result := pi.ConfirmMelds()
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "ConfirmMelds")
}

func TestPinochleInteractor_Play(t *testing.T) {
	gameMock, ppMock := setupPinochleInteractorMocks(domain.PinochlePhasePlay)
	gameMock.On("PlayerPlay", 2).Return(nil)

	pi := usecase.NewPinochleInteractor(gameMock, ppMock)
	result := pi.Play(2)
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "PlayerPlay", 2)
}

func TestPinochleInteractor_NextTrick(t *testing.T) {
	gameMock, ppMock := setupPinochleInteractorMocks(domain.PinochlePhaseTrickEnd)
	gameMock.On("ResolveTrick").Return()
	gameMock.On("NextTrick").Return()

	pi := usecase.NewPinochleInteractor(gameMock, ppMock)
	result := pi.NextTrick()
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "ResolveTrick")
	gameMock.AssertCalled(t, "NextTrick")
}

func TestPinochleInteractor_NextRound(t *testing.T) {
	gameMock, ppMock := setupPinochleInteractorMocks(domain.PinochlePhaseRoundEnd)
	gameMock.On("NextRound").Return()

	pi := usecase.NewPinochleInteractor(gameMock, ppMock)
	result := pi.NextRound()
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "NextRound")
}

func TestPinochleInteractor_Hint(t *testing.T) {
	gameMock, ppMock := setupPinochleInteractorMocks(domain.PinochlePhaseBid)
	ppMock.On("HintOutput", mock.Anything).Return(`{"hint":"bid"}`)

	pi := usecase.NewPinochleInteractor(gameMock, ppMock)
	result := pi.Hint()
	assert.Contains(t, result, "hint")
}

func TestPinochleInteractor_ActionLog(t *testing.T) {
	gameMock, ppMock := setupPinochleInteractorMocks(domain.PinochlePhaseBid)
	ppMock.On("ActionLogOutput", mock.Anything).Return(`[]`)

	pi := usecase.NewPinochleInteractor(gameMock, ppMock)
	result := pi.ActionLog()
	assert.Equal(t, "[]", result)
}

func TestPinochleInteractor_GetConfig(t *testing.T) {
	gameMock, ppMock := setupPinochleInteractorMocks(domain.PinochlePhaseBid)
	cfg := domain.DefaultPinochleConfig()
	gameMock.On("GetConfig").Return(cfg)

	pi := usecase.NewPinochleInteractor(gameMock, ppMock)
	result := pi.GetConfig()
	assert.Equal(t, cfg, result)
}

func TestPinochleInteractor_GameEndGuard(t *testing.T) {
	ppMock := new(presenter.MockPinochlePresenter)
	ppMock.On("Output", mock.Anything, mock.Anything).Return(`{"gameEndFlag":true}`)
	gameMock := new(interfaces.MockPinochleGame)
	gameMock.On("GetGameEndFlag").Return(true)

	pi := usecase.NewPinochleInteractor(gameMock, ppMock)

	assert.Contains(t, pi.Bid(20), "gameEndFlag")
	assert.Contains(t, pi.Pass(), "gameEndFlag")
	assert.Contains(t, pi.CallTrump(1), "gameEndFlag")
	assert.Contains(t, pi.NextRound(), "gameEndFlag")
	assert.Contains(t, pi.ConfirmMelds(), "gameEndFlag")
}
