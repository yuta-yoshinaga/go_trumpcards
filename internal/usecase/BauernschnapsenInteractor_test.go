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

func TestNewBauernschnapsenInteractor_NilGuards(t *testing.T) {
	gpMock := new(presenter.MockBauernschnapsenPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "BauernschnapsenInteractor: g must not be nil", func() {
			usecase.NewBauernschnapsenInteractor(nil, gpMock)
		})
	})

	t.Run("panics when gp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockBauernschnapsenGame)
		assert.PanicsWithValue(t, "BauernschnapsenInteractor: gp must not be nil", func() {
			usecase.NewBauernschnapsenInteractor(gameMock, nil)
		})
	})
}

func setupBauernschnapsenMocks(phase domain.BauernschnapsenPhase) (*interfaces.MockBauernschnapsenGame, *presenter.MockBauernschnapsenPresenter) {
	gpMock := new(presenter.MockBauernschnapsenPresenter)
	gpMock.On("Output", mock.Anything, mock.Anything).Return(`{"phase":0}`)
	gameMock := new(interfaces.MockBauernschnapsenGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(phase)
	gameMock.On("IsHumanTurn").Return(true)
	return gameMock, gpMock
}

func TestBauernschnapsenInteractor_Reset(t *testing.T) {
	gameMock, gpMock := setupBauernschnapsenMocks(domain.BauernschnapsenPhasePlay)
	gameMock.On("Reset").Return()

	gi := usecase.NewBauernschnapsenInteractor(gameMock, gpMock)
	result := gi.Reset()
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "Reset")
}

func TestBauernschnapsenInteractor_ResetWithConfig(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		gameMock, gpMock := setupBauernschnapsenMocks(domain.BauernschnapsenPhasePlay)
		cfg := domain.DefaultBauernschnapsenConfig()
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()

		gi := usecase.NewBauernschnapsenInteractor(gameMock, gpMock)
		assert.Contains(t, gi.ResetWithConfig(cfg), "phase")
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config", func(t *testing.T) {
		gameMock, gpMock := setupBauernschnapsenMocks(domain.BauernschnapsenPhasePlay)
		bad := domain.BauernschnapsenConfig{CpuDifficulty: 99, TargetScore: 101}

		gi := usecase.NewBauernschnapsenInteractor(gameMock, gpMock)
		assert.Contains(t, gi.ResetWithConfig(bad), "phase")
	})
}

func TestBauernschnapsenInteractor_Play(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		gameMock, gpMock := setupBauernschnapsenMocks(domain.BauernschnapsenPhasePlay)
		gameMock.On("PlayerPlay", 0).Return(nil)

		gi := usecase.NewBauernschnapsenInteractor(gameMock, gpMock)
		assert.Contains(t, gi.Play(0), "phase")
	})

	t.Run("error", func(t *testing.T) {
		gameMock, gpMock := setupBauernschnapsenMocks(domain.BauernschnapsenPhasePlay)
		gameMock.On("PlayerPlay", 99).Return(errors.New("bad"))

		gi := usecase.NewBauernschnapsenInteractor(gameMock, gpMock)
		assert.Contains(t, gi.Play(99), "phase")
	})

	t.Run("game ended", func(t *testing.T) {
		gpMock := new(presenter.MockBauernschnapsenPresenter)
		gpMock.On("Output", mock.Anything, mock.Anything).Return(`{"gameEnd":true}`)
		gameMock := new(interfaces.MockBauernschnapsenGame)
		gameMock.On("GetGameEndFlag").Return(true)
		gameMock.On("GetPhase").Return(domain.BauernschnapsenPhaseGameEnd)

		gi := usecase.NewBauernschnapsenInteractor(gameMock, gpMock)
		assert.Contains(t, gi.Play(0), "gameEnd")
	})
}

func TestBauernschnapsenInteractor_DeclareMarriage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		gameMock, gpMock := setupBauernschnapsenMocks(domain.BauernschnapsenPhasePlay)
		gameMock.On("PlayerDeclareMarriage", 0).Return(nil)

		gi := usecase.NewBauernschnapsenInteractor(gameMock, gpMock)
		assert.Contains(t, gi.DeclareMarriage(0), "phase")
	})

	t.Run("error", func(t *testing.T) {
		gameMock, gpMock := setupBauernschnapsenMocks(domain.BauernschnapsenPhasePlay)
		gameMock.On("PlayerDeclareMarriage", 1).Return(errors.New("bad"))

		gi := usecase.NewBauernschnapsenInteractor(gameMock, gpMock)
		assert.Contains(t, gi.DeclareMarriage(1), "phase")
	})

	t.Run("game ended", func(t *testing.T) {
		gpMock := new(presenter.MockBauernschnapsenPresenter)
		gpMock.On("Output", mock.Anything, mock.Anything).Return(`{"gameEnd":true}`)
		gameMock := new(interfaces.MockBauernschnapsenGame)
		gameMock.On("GetGameEndFlag").Return(true)
		gameMock.On("GetPhase").Return(domain.BauernschnapsenPhaseGameEnd)

		gi := usecase.NewBauernschnapsenInteractor(gameMock, gpMock)
		assert.Contains(t, gi.DeclareMarriage(0), "gameEnd")
	})
}

func TestBauernschnapsenInteractor_NextTrick(t *testing.T) {
	gameMock, gpMock := setupBauernschnapsenMocks(domain.BauernschnapsenPhasePlay)
	gameMock.On("ResolveTrick").Return()
	gameMock.On("NextTrick").Return()

	gi := usecase.NewBauernschnapsenInteractor(gameMock, gpMock)
	assert.Contains(t, gi.NextTrick(), "phase")
}

func TestBauernschnapsenInteractor_NextRound(t *testing.T) {
	gameMock, gpMock := setupBauernschnapsenMocks(domain.BauernschnapsenPhasePlay)
	gameMock.On("ScoreRound").Return()
	gameMock.On("NextRound").Return()

	gi := usecase.NewBauernschnapsenInteractor(gameMock, gpMock)
	assert.Contains(t, gi.NextRound(), "phase")
}

func TestBauernschnapsenInteractor_GetConfig(t *testing.T) {
	gameMock := new(interfaces.MockBauernschnapsenGame)
	cfg := domain.DefaultBauernschnapsenConfig()
	gameMock.On("GetConfig").Return(cfg)
	gpMock := new(presenter.MockBauernschnapsenPresenter)

	gi := usecase.NewBauernschnapsenInteractor(gameMock, gpMock)
	assert.Equal(t, cfg, gi.GetConfig())
}

func TestBauernschnapsenInteractor_HintAndLog(t *testing.T) {
	gameMock := new(interfaces.MockBauernschnapsenGame)
	gpMock := new(presenter.MockBauernschnapsenPresenter)
	gpMock.On("HintOutput", gameMock).Return(`{"hint":1}`)
	gpMock.On("ActionLogOutput", gameMock).Return(`{"log":1}`)

	gi := usecase.NewBauernschnapsenInteractor(gameMock, gpMock)
	assert.Contains(t, gi.Hint(), "hint")
	assert.Contains(t, gi.ActionLog(), "log")
}

func TestBauernschnapsenInteractor_RealReset(t *testing.T) {
	g := domain.NewDefaultBauernschnapsen()
	gi := usecase.NewBauernschnapsenInteractor(g, new(presenter.MockBauernschnapsenPresenter))
	assert.NotNil(t, gi)
}
