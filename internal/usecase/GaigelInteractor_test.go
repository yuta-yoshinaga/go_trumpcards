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

func TestNewGaigelInteractor_NilGuards(t *testing.T) {
	gpMock := new(presenter.MockGaigelPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "GaigelInteractor: g must not be nil", func() {
			usecase.NewGaigelInteractor(nil, gpMock)
		})
	})

	t.Run("panics when gp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockGaigelGame)
		assert.PanicsWithValue(t, "GaigelInteractor: gp must not be nil", func() {
			usecase.NewGaigelInteractor(gameMock, nil)
		})
	})
}

func setupGaigelMocks(phase domain.GaigelPhase) (*interfaces.MockGaigelGame, *presenter.MockGaigelPresenter) {
	gpMock := new(presenter.MockGaigelPresenter)
	gpMock.On("Output", mock.Anything, mock.Anything).Return(`{"phase":0}`)
	gameMock := new(interfaces.MockGaigelGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(phase)
	gameMock.On("IsHumanTurn").Return(true)
	return gameMock, gpMock
}

func TestGaigelInteractor_Reset(t *testing.T) {
	gameMock, gpMock := setupGaigelMocks(domain.GaigelPhasePlay)
	gameMock.On("Reset").Return()

	gi := usecase.NewGaigelInteractor(gameMock, gpMock)
	result := gi.Reset()
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "Reset")
}

func TestGaigelInteractor_ResetWithConfig(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		gameMock, gpMock := setupGaigelMocks(domain.GaigelPhasePlay)
		cfg := domain.DefaultGaigelConfig()
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()

		gi := usecase.NewGaigelInteractor(gameMock, gpMock)
		assert.Contains(t, gi.ResetWithConfig(cfg), "phase")
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config", func(t *testing.T) {
		gameMock, gpMock := setupGaigelMocks(domain.GaigelPhasePlay)
		bad := domain.GaigelConfig{CpuDifficulty: 99, TargetScore: 101}

		gi := usecase.NewGaigelInteractor(gameMock, gpMock)
		assert.Contains(t, gi.ResetWithConfig(bad), "phase")
	})
}

func TestGaigelInteractor_Play(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		gameMock, gpMock := setupGaigelMocks(domain.GaigelPhasePlay)
		gameMock.On("PlayerPlay", 0).Return(nil)

		gi := usecase.NewGaigelInteractor(gameMock, gpMock)
		assert.Contains(t, gi.Play(0), "phase")
	})

	t.Run("error", func(t *testing.T) {
		gameMock, gpMock := setupGaigelMocks(domain.GaigelPhasePlay)
		gameMock.On("PlayerPlay", 99).Return(errors.New("bad"))

		gi := usecase.NewGaigelInteractor(gameMock, gpMock)
		assert.Contains(t, gi.Play(99), "phase")
	})

	t.Run("game ended", func(t *testing.T) {
		gpMock := new(presenter.MockGaigelPresenter)
		gpMock.On("Output", mock.Anything, mock.Anything).Return(`{"gameEnd":true}`)
		gameMock := new(interfaces.MockGaigelGame)
		gameMock.On("GetGameEndFlag").Return(true)
		gameMock.On("GetPhase").Return(domain.GaigelPhaseGameEnd)

		gi := usecase.NewGaigelInteractor(gameMock, gpMock)
		assert.Contains(t, gi.Play(0), "gameEnd")
	})
}

func TestGaigelInteractor_DeclareMarriage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		gameMock, gpMock := setupGaigelMocks(domain.GaigelPhasePlay)
		gameMock.On("PlayerDeclareMarriage", 0).Return(nil)

		gi := usecase.NewGaigelInteractor(gameMock, gpMock)
		assert.Contains(t, gi.DeclareMarriage(0), "phase")
	})

	t.Run("error", func(t *testing.T) {
		gameMock, gpMock := setupGaigelMocks(domain.GaigelPhasePlay)
		gameMock.On("PlayerDeclareMarriage", 1).Return(errors.New("bad"))

		gi := usecase.NewGaigelInteractor(gameMock, gpMock)
		assert.Contains(t, gi.DeclareMarriage(1), "phase")
	})

	t.Run("game ended", func(t *testing.T) {
		gpMock := new(presenter.MockGaigelPresenter)
		gpMock.On("Output", mock.Anything, mock.Anything).Return(`{"gameEnd":true}`)
		gameMock := new(interfaces.MockGaigelGame)
		gameMock.On("GetGameEndFlag").Return(true)
		gameMock.On("GetPhase").Return(domain.GaigelPhaseGameEnd)

		gi := usecase.NewGaigelInteractor(gameMock, gpMock)
		assert.Contains(t, gi.DeclareMarriage(0), "gameEnd")
	})
}

func TestGaigelInteractor_NextTrick(t *testing.T) {
	gameMock, gpMock := setupGaigelMocks(domain.GaigelPhasePlay)
	gameMock.On("ResolveTrick").Return()
	gameMock.On("NextTrick").Return()

	gi := usecase.NewGaigelInteractor(gameMock, gpMock)
	assert.Contains(t, gi.NextTrick(), "phase")
}

func TestGaigelInteractor_NextRound(t *testing.T) {
	gameMock, gpMock := setupGaigelMocks(domain.GaigelPhasePlay)
	gameMock.On("ScoreRound").Return()
	gameMock.On("NextRound").Return()

	gi := usecase.NewGaigelInteractor(gameMock, gpMock)
	assert.Contains(t, gi.NextRound(), "phase")
}

func TestGaigelInteractor_GetConfig(t *testing.T) {
	gameMock := new(interfaces.MockGaigelGame)
	cfg := domain.DefaultGaigelConfig()
	gameMock.On("GetConfig").Return(cfg)
	gpMock := new(presenter.MockGaigelPresenter)

	gi := usecase.NewGaigelInteractor(gameMock, gpMock)
	assert.Equal(t, cfg, gi.GetConfig())
}

func TestGaigelInteractor_HintAndLog(t *testing.T) {
	gameMock := new(interfaces.MockGaigelGame)
	gpMock := new(presenter.MockGaigelPresenter)
	gpMock.On("HintOutput", gameMock).Return(`{"hint":1}`)
	gpMock.On("ActionLogOutput", gameMock).Return(`{"log":1}`)

	gi := usecase.NewGaigelInteractor(gameMock, gpMock)
	assert.Contains(t, gi.Hint(), "hint")
	assert.Contains(t, gi.ActionLog(), "log")
}

func TestGaigelInteractor_RealReset(t *testing.T) {
	g := domain.NewDefaultGaigel()
	gi := usecase.NewGaigelInteractor(g, new(presenter.MockGaigelPresenter))
	assert.NotNil(t, gi)
}
