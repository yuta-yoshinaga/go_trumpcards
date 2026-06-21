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

func TestNewCuckooInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockCuckooPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "CuckooInteractor: g must not be nil", func() {
			usecase.NewCuckooInteractor(nil, pMock)
		})
	})

	t.Run("panics when gp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockCuckooGame)
		assert.PanicsWithValue(t, "CuckooInteractor: gp must not be nil", func() {
			usecase.NewCuckooInteractor(gameMock, nil)
		})
	})
}

func TestCuckooInteractor_Reset(t *testing.T) {
	mockOutput := `{"phase":0}`
	pMock := new(presenter.MockCuckooPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockCuckooGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.CuckooPhaseTurn)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewCuckooInteractor(gameMock, pMock)
	assert.Equal(t, mockOutput, ci.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestCuckooInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockCuckooPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCuckooGame)
		cfg := domain.CuckooConfig{CpuDifficulty: domain.CuckooCpuDifficultyHard, InitialLives: 5}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.CuckooPhaseTurn)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewCuckooInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.ResetWithConfig(cfg))
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config returns error without resetting", func(t *testing.T) {
		pMock := new(presenter.MockCuckooPresenter)
		gameMock := new(interfaces.MockCuckooGame)
		pMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")

		ci := usecase.NewCuckooInteractor(gameMock, pMock)
		cfg := domain.CuckooConfig{CpuDifficulty: domain.CuckooCpuDifficulty(-1), InitialLives: 3}
		assert.Equal(t, "validation error", ci.ResetWithConfig(cfg))
		gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
	})
}

func TestCuckooInteractor_Keep(t *testing.T) {
	mockOutput := `{}`

	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockCuckooPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCuckooGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerKeep").Return(nil)
		gameMock.On("GetPhase").Return(domain.CuckooPhaseTurn)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewCuckooInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Keep())
		gameMock.AssertCalled(t, "PlayerKeep")
	})

	t.Run("error", func(t *testing.T) {
		err := errors.New("e")
		pMock := new(presenter.MockCuckooPresenter)
		pMock.On("Output", mock.Anything, err).Return(mockOutput)
		gameMock := new(interfaces.MockCuckooGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerKeep").Return(err)

		ci := usecase.NewCuckooInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Keep())
	})

	t.Run("game ended", func(t *testing.T) {
		pMock := new(presenter.MockCuckooPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCuckooGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewCuckooInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Keep())
		gameMock.AssertNotCalled(t, "PlayerKeep")
	})
}

func TestCuckooInteractor_SwapRefuseAccept(t *testing.T) {
	mockOutput := `{}`
	newGame := func() (*interfaces.MockCuckooGame, *presenter.MockCuckooPresenter) {
		pMock := new(presenter.MockCuckooPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCuckooGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.CuckooPhaseTurn)
		gameMock.On("IsHumanTurn").Return(true)
		return gameMock, pMock
	}

	t.Run("swap", func(t *testing.T) {
		gameMock, pMock := newGame()
		gameMock.On("PlayerSwap").Return(nil)
		ci := usecase.NewCuckooInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Swap())
		gameMock.AssertCalled(t, "PlayerSwap")
	})

	t.Run("refuse", func(t *testing.T) {
		gameMock, pMock := newGame()
		gameMock.On("PlayerRefuse").Return(nil)
		ci := usecase.NewCuckooInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Refuse())
		gameMock.AssertCalled(t, "PlayerRefuse")
	})

	t.Run("accept", func(t *testing.T) {
		gameMock, pMock := newGame()
		gameMock.On("PlayerAcceptSwap").Return(nil)
		ci := usecase.NewCuckooInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.AcceptSwap())
		gameMock.AssertCalled(t, "PlayerAcceptSwap")
	})
}

func TestCuckooInteractor_NextRound(t *testing.T) {
	mockOutput := `{}`

	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockCuckooPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCuckooGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextRound").Return()
		gameMock.On("GetPhase").Return(domain.CuckooPhaseTurn)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewCuckooInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.NextRound())
		gameMock.AssertCalled(t, "NextRound")
	})

	t.Run("game ended", func(t *testing.T) {
		pMock := new(presenter.MockCuckooPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCuckooGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewCuckooInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.NextRound())
		gameMock.AssertNotCalled(t, "NextRound")
	})
}

func TestCuckooInteractor_RunCpuTurns(t *testing.T) {
	mockOutput := `{}`
	pMock := new(presenter.MockCuckooPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockCuckooGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	// Two CPU turns then a human turn.
	gameMock.On("GetPhase").Return(domain.CuckooPhaseTurn)
	// Two CPU turns (IsHumanTurn=false) then a human turn (true).
	gameMock.On("IsHumanTurn").Return(false).Once()
	gameMock.On("IsHumanTurn").Return(false).Once()
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("CpuPlay").Return()

	ci := usecase.NewCuckooInteractor(gameMock, pMock)
	assert.Equal(t, mockOutput, ci.Reset())
	gameMock.AssertNumberOfCalls(t, "CpuPlay", 2)
}

func TestCuckooInteractor_GetConfigAndLog(t *testing.T) {
	pMock := new(presenter.MockCuckooPresenter)
	pMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockCuckooGame)
	cfg := domain.CuckooConfig{CpuDifficulty: domain.CuckooCpuDifficultyHard, InitialLives: 2}
	gameMock.On("GetConfig").Return(cfg)

	ci := usecase.NewCuckooInteractor(gameMock, pMock)
	assert.Equal(t, cfg, ci.GetConfig())
	assert.Equal(t, "log", ci.ActionLog())
}

func TestRestoreCuckooInteractor(t *testing.T) {
	pMock := new(presenter.MockCuckooPresenter)
	g := domain.NewDefaultCuckoo()
	g.Reset()
	data, err := g.MarshalJSON()
	assert.NoError(t, err)

	ci, err := usecase.RestoreCuckooInteractor(data, pMock)
	assert.NoError(t, err)
	assert.NotNil(t, ci)
}

func TestRestoreCuckooInteractor_InvalidJSON(t *testing.T) {
	pMock := new(presenter.MockCuckooPresenter)
	_, err := usecase.RestoreCuckooInteractor([]byte("not json"), pMock)
	assert.Error(t, err)
}
