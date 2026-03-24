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

func TestNewEuchreInteractor_NilGuards(t *testing.T) {
	epMock := new(presenter.MockEuchrePresenter)

	t.Run("panics when e is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "EuchreInteractor: e must not be nil", func() {
			usecase.NewEuchreInteractor(nil, epMock)
		})
	})

	t.Run("panics when ep is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockEuchreGame)
		assert.PanicsWithValue(t, "EuchreInteractor: ep must not be nil", func() {
			usecase.NewEuchreInteractor(gameMock, nil)
		})
	})
}

func setupEuchreInteractorMocks(phase domain.EuchrePhase) (*interfaces.MockEuchreGame, *presenter.MockEuchrePresenter) {
	epMock := new(presenter.MockEuchrePresenter)
	epMock.On("Output", mock.Anything, mock.Anything).Return(`{"phase":0}`)
	gameMock := new(interfaces.MockEuchreGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(phase)
	gameMock.On("IsHumanBidTurn").Return(true)
	gameMock.On("IsHumanTurn").Return(true)
	return gameMock, epMock
}

func TestEuchreInteractor_Reset(t *testing.T) {
	t.Run("stays in pickup phase", func(t *testing.T) {
		gameMock, epMock := setupEuchreInteractorMocks(domain.EuchrePhasePickUp)
		gameMock.On("Reset").Return()

		ei := usecase.NewEuchreInteractor(gameMock, epMock)
		result := ei.Reset()
		assert.Contains(t, result, "phase")
		gameMock.AssertCalled(t, "Reset")
	})
}

func TestEuchreInteractor_ResetWithConfig(t *testing.T) {
	t.Run("sets config then resets", func(t *testing.T) {
		gameMock, epMock := setupEuchreInteractorMocks(domain.EuchrePhasePickUp)
		cfg := domain.EuchreConfig{CpuDifficulty: domain.EuchreCpuDifficultyHard, PointLimit: 5}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()

		ei := usecase.NewEuchreInteractor(gameMock, epMock)
		result := ei.ResetWithConfig(cfg)
		assert.Contains(t, result, "phase")
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config returns error", func(t *testing.T) {
		gameMock, epMock := setupEuchreInteractorMocks(domain.EuchrePhasePickUp)
		invalidCfg := domain.EuchreConfig{CpuDifficulty: 99, PointLimit: 10}

		ei := usecase.NewEuchreInteractor(gameMock, epMock)
		result := ei.ResetWithConfig(invalidCfg)
		assert.Contains(t, result, "phase")
	})
}

func TestEuchreInteractor_PickUp(t *testing.T) {
	t.Run("successful pickup", func(t *testing.T) {
		gameMock, epMock := setupEuchreInteractorMocks(domain.EuchrePhasePickUp)
		gameMock.On("PlayerPickUp", true, false).Return(nil)

		ei := usecase.NewEuchreInteractor(gameMock, epMock)
		result := ei.PickUp(true, false)
		assert.Contains(t, result, "phase")
	})

	t.Run("pickup error", func(t *testing.T) {
		gameMock, epMock := setupEuchreInteractorMocks(domain.EuchrePhasePickUp)
		gameMock.On("PlayerPickUp", true, false).Return(errors.New("test error"))

		ei := usecase.NewEuchreInteractor(gameMock, epMock)
		result := ei.PickUp(true, false)
		assert.Contains(t, result, "phase")
	})

	t.Run("game ended guard", func(t *testing.T) {
		epMock := new(presenter.MockEuchrePresenter)
		epMock.On("Output", mock.Anything, mock.Anything).Return(`{"gameEnd":true}`)
		gameMock := new(interfaces.MockEuchreGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ei := usecase.NewEuchreInteractor(gameMock, epMock)
		result := ei.PickUp(true, false)
		assert.Contains(t, result, "gameEnd")
	})
}

func TestEuchreInteractor_CallTrump(t *testing.T) {
	t.Run("successful call", func(t *testing.T) {
		gameMock, epMock := setupEuchreInteractorMocks(domain.EuchrePhasePlay)
		gameMock.On("PlayerCallTrump", domain.CardDesignSpade, false).Return(nil)

		ei := usecase.NewEuchreInteractor(gameMock, epMock)
		result := ei.CallTrump(domain.CardDesignSpade, false)
		assert.Contains(t, result, "phase")
	})

	t.Run("call error", func(t *testing.T) {
		gameMock, epMock := setupEuchreInteractorMocks(domain.EuchrePhaseCallTrump)
		gameMock.On("PlayerCallTrump", 99, false).Return(errors.New("invalid suit"))

		ei := usecase.NewEuchreInteractor(gameMock, epMock)
		result := ei.CallTrump(99, false)
		assert.Contains(t, result, "phase")
	})
}

func TestEuchreInteractor_PassCall(t *testing.T) {
	t.Run("successful pass", func(t *testing.T) {
		gameMock, epMock := setupEuchreInteractorMocks(domain.EuchrePhasePickUp)
		gameMock.On("PlayerPassCall").Return(nil)

		ei := usecase.NewEuchreInteractor(gameMock, epMock)
		result := ei.PassCall()
		assert.Contains(t, result, "phase")
	})
}

func TestEuchreInteractor_Pass(t *testing.T) {
	t.Run("pickup phase delegates to PickUp", func(t *testing.T) {
		gameMock, epMock := setupEuchreInteractorMocks(domain.EuchrePhasePickUp)
		gameMock.On("PlayerPickUp", false, false).Return(nil)

		ei := usecase.NewEuchreInteractor(gameMock, epMock)
		result := ei.Pass()
		assert.Contains(t, result, "phase")
	})

	t.Run("call trump phase delegates to PassCall", func(t *testing.T) {
		gameMock, epMock := setupEuchreInteractorMocks(domain.EuchrePhaseCallTrump)
		gameMock.On("PlayerPassCall").Return(nil)

		ei := usecase.NewEuchreInteractor(gameMock, epMock)
		result := ei.Pass()
		assert.Contains(t, result, "phase")
	})

	t.Run("wrong phase returns error", func(t *testing.T) {
		gameMock, epMock := setupEuchreInteractorMocks(domain.EuchrePhasePlay)

		ei := usecase.NewEuchreInteractor(gameMock, epMock)
		result := ei.Pass()
		assert.Contains(t, result, "phase")
	})
}

func TestEuchreInteractor_Discard(t *testing.T) {
	t.Run("successful discard", func(t *testing.T) {
		gameMock, epMock := setupEuchreInteractorMocks(domain.EuchrePhasePlay)
		gameMock.On("PlayerDiscard", 0).Return(nil)

		ei := usecase.NewEuchreInteractor(gameMock, epMock)
		result := ei.Discard(0)
		assert.Contains(t, result, "phase")
	})
}

func TestEuchreInteractor_Play(t *testing.T) {
	t.Run("successful play", func(t *testing.T) {
		gameMock, epMock := setupEuchreInteractorMocks(domain.EuchrePhasePlay)
		gameMock.On("PlayerPlay", 0).Return(nil)

		ei := usecase.NewEuchreInteractor(gameMock, epMock)
		result := ei.Play(0)
		assert.Contains(t, result, "phase")
	})

	t.Run("play error", func(t *testing.T) {
		gameMock, epMock := setupEuchreInteractorMocks(domain.EuchrePhasePlay)
		gameMock.On("PlayerPlay", 0).Return(errors.New("follow suit"))

		ei := usecase.NewEuchreInteractor(gameMock, epMock)
		result := ei.Play(0)
		assert.Contains(t, result, "phase")
	})
}

func TestEuchreInteractor_NextTrick(t *testing.T) {
	gameMock, epMock := setupEuchreInteractorMocks(domain.EuchrePhasePlay)
	gameMock.On("ResolveTrick").Return()
	gameMock.On("NextTrick").Return()

	ei := usecase.NewEuchreInteractor(gameMock, epMock)
	result := ei.NextTrick()
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "ResolveTrick")
	gameMock.AssertCalled(t, "NextTrick")
}

func TestEuchreInteractor_NextRound(t *testing.T) {
	t.Run("scores and starts next round", func(t *testing.T) {
		gameMock, epMock := setupEuchreInteractorMocks(domain.EuchrePhasePickUp)
		gameMock.On("ScoreRound").Return()
		gameMock.On("NextRound").Return()

		ei := usecase.NewEuchreInteractor(gameMock, epMock)
		result := ei.NextRound()
		assert.Contains(t, result, "phase")
		gameMock.AssertCalled(t, "ScoreRound")
		gameMock.AssertCalled(t, "NextRound")
	})
}

func TestEuchreInteractor_GetConfig(t *testing.T) {
	gameMock, epMock := setupEuchreInteractorMocks(domain.EuchrePhasePickUp)
	cfg := domain.DefaultEuchreConfig()
	gameMock.On("GetConfig").Return(cfg)

	ei := usecase.NewEuchreInteractor(gameMock, epMock)
	assert.Equal(t, cfg, ei.GetConfig())
}

func TestEuchreInteractor_Hint(t *testing.T) {
	gameMock, epMock := setupEuchreInteractorMocks(domain.EuchrePhasePickUp)
	epMock.On("HintOutput", gameMock).Return(`{"hint":"test"}`)

	ei := usecase.NewEuchreInteractor(gameMock, epMock)
	result := ei.Hint()
	assert.Contains(t, result, "hint")
}

func TestEuchreInteractor_ActionLog(t *testing.T) {
	gameMock, epMock := setupEuchreInteractorMocks(domain.EuchrePhasePickUp)
	epMock.On("ActionLogOutput", gameMock).Return(`[{"action":"test"}]`)

	ei := usecase.NewEuchreInteractor(gameMock, epMock)
	result := ei.ActionLog()
	assert.Contains(t, result, "action")
}
