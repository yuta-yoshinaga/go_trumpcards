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

func TestNewWattenInteractor_NilGuards(t *testing.T) {
	wpMock := new(presenter.MockWattenPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "WattenInteractor: g must not be nil", func() {
			usecase.NewWattenInteractor(nil, wpMock)
		})
	})

	t.Run("panics when wp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockWattenGame)
		assert.PanicsWithValue(t, "WattenInteractor: wp must not be nil", func() {
			usecase.NewWattenInteractor(gameMock, nil)
		})
	})
}

func setupWattenMocks(phase domain.WattenPhase) (*interfaces.MockWattenGame, *presenter.MockWattenPresenter) {
	wpMock := new(presenter.MockWattenPresenter)
	wpMock.On("Output", mock.Anything, mock.Anything).Return(`{"phase":0}`)
	gameMock := new(interfaces.MockWattenGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(phase)
	gameMock.On("IsHumanDeclareTurn").Return(true)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("IsHumanRespondTurn").Return(true)
	return gameMock, wpMock
}

func gameEndedWatten() (*interfaces.MockWattenGame, *presenter.MockWattenPresenter) {
	wpMock := new(presenter.MockWattenPresenter)
	wpMock.On("Output", mock.Anything, mock.Anything).Return(`{"gameEnd":true}`)
	gameMock := new(interfaces.MockWattenGame)
	gameMock.On("GetGameEndFlag").Return(true)
	gameMock.On("GetPhase").Return(domain.WattenPhaseGameEnd)
	return gameMock, wpMock
}

func TestWattenInteractor_Reset(t *testing.T) {
	gameMock, wpMock := setupWattenMocks(domain.WattenPhaseDeclare)
	gameMock.On("Reset").Return()
	gameMock.On("CpuDeclare").Return()

	wi := usecase.NewWattenInteractor(gameMock, wpMock)
	assert.Contains(t, wi.Reset(), "phase")
	gameMock.AssertCalled(t, "Reset")
}

func TestWattenInteractor_ResetWithConfig(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		gameMock, wpMock := setupWattenMocks(domain.WattenPhaseDeclare)
		cfg := domain.DefaultWattenConfig()
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()

		wi := usecase.NewWattenInteractor(gameMock, wpMock)
		assert.Contains(t, wi.ResetWithConfig(cfg), "phase")
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config", func(t *testing.T) {
		gameMock, wpMock := setupWattenMocks(domain.WattenPhaseDeclare)
		bad := domain.WattenConfig{CpuDifficulty: 99, TargetScore: 15, MaxRaises: 5}

		wi := usecase.NewWattenInteractor(gameMock, wpMock)
		assert.Contains(t, wi.ResetWithConfig(bad), "phase")
	})
}

func TestWattenInteractor_Declare(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		gameMock, wpMock := setupWattenMocks(domain.WattenPhasePlay)
		gameMock.On("PlayerDeclare", 10, domain.CardDesignSpade).Return(nil)

		wi := usecase.NewWattenInteractor(gameMock, wpMock)
		assert.Contains(t, wi.Declare(10, domain.CardDesignSpade), "phase")
	})

	t.Run("error", func(t *testing.T) {
		gameMock, wpMock := setupWattenMocks(domain.WattenPhaseDeclare)
		gameMock.On("PlayerDeclare", 5, 9).Return(errors.New("bad"))

		wi := usecase.NewWattenInteractor(gameMock, wpMock)
		assert.Contains(t, wi.Declare(5, 9), "phase")
	})

	t.Run("game ended", func(t *testing.T) {
		gameMock, wpMock := gameEndedWatten()
		wi := usecase.NewWattenInteractor(gameMock, wpMock)
		assert.Contains(t, wi.Declare(10, domain.CardDesignSpade), "gameEnd")
	})
}

func TestWattenInteractor_Play(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		gameMock, wpMock := setupWattenMocks(domain.WattenPhasePlay)
		gameMock.On("PlayerPlay", 0).Return(nil)

		wi := usecase.NewWattenInteractor(gameMock, wpMock)
		assert.Contains(t, wi.Play(0), "phase")
	})

	t.Run("error", func(t *testing.T) {
		gameMock, wpMock := setupWattenMocks(domain.WattenPhasePlay)
		gameMock.On("PlayerPlay", 99).Return(errors.New("bad"))

		wi := usecase.NewWattenInteractor(gameMock, wpMock)
		assert.Contains(t, wi.Play(99), "phase")
	})

	t.Run("not play phase", func(t *testing.T) {
		gameMock, wpMock := setupWattenMocks(domain.WattenPhaseRespond)
		wi := usecase.NewWattenInteractor(gameMock, wpMock)
		assert.Contains(t, wi.Play(0), "phase")
	})

	t.Run("game ended", func(t *testing.T) {
		gameMock, wpMock := gameEndedWatten()
		wi := usecase.NewWattenInteractor(gameMock, wpMock)
		assert.Contains(t, wi.Play(0), "gameEnd")
	})
}

func TestWattenInteractor_Raise(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		gameMock, wpMock := setupWattenMocks(domain.WattenPhaseRespond)
		gameMock.On("PlayerRaise").Return(nil)

		wi := usecase.NewWattenInteractor(gameMock, wpMock)
		assert.Contains(t, wi.Raise(), "phase")
	})

	t.Run("error", func(t *testing.T) {
		gameMock, wpMock := setupWattenMocks(domain.WattenPhasePlay)
		gameMock.On("PlayerRaise").Return(errors.New("bad"))

		wi := usecase.NewWattenInteractor(gameMock, wpMock)
		assert.Contains(t, wi.Raise(), "phase")
	})

	t.Run("game ended", func(t *testing.T) {
		gameMock, wpMock := gameEndedWatten()
		wi := usecase.NewWattenInteractor(gameMock, wpMock)
		assert.Contains(t, wi.Raise(), "gameEnd")
	})
}

func TestWattenInteractor_Respond(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		gameMock, wpMock := setupWattenMocks(domain.WattenPhasePlay)
		gameMock.On("PlayerRespond", true).Return(nil)

		wi := usecase.NewWattenInteractor(gameMock, wpMock)
		assert.Contains(t, wi.Respond(true), "phase")
	})

	t.Run("error", func(t *testing.T) {
		gameMock, wpMock := setupWattenMocks(domain.WattenPhaseRespond)
		gameMock.On("PlayerRespond", false).Return(errors.New("bad"))

		wi := usecase.NewWattenInteractor(gameMock, wpMock)
		assert.Contains(t, wi.Respond(false), "phase")
	})

	t.Run("game ended", func(t *testing.T) {
		gameMock, wpMock := gameEndedWatten()
		wi := usecase.NewWattenInteractor(gameMock, wpMock)
		assert.Contains(t, wi.Respond(true), "gameEnd")
	})
}

func TestWattenInteractor_NextRound(t *testing.T) {
	gameMock, wpMock := setupWattenMocks(domain.WattenPhaseDeclare)
	gameMock.On("ScoreRound").Return()
	gameMock.On("NextRound").Return()
	gameMock.On("CpuDeclare").Return()

	wi := usecase.NewWattenInteractor(gameMock, wpMock)
	assert.Contains(t, wi.NextRound(), "phase")
}

func TestWattenInteractor_GetConfig(t *testing.T) {
	gameMock := new(interfaces.MockWattenGame)
	cfg := domain.DefaultWattenConfig()
	gameMock.On("GetConfig").Return(cfg)
	wpMock := new(presenter.MockWattenPresenter)

	wi := usecase.NewWattenInteractor(gameMock, wpMock)
	assert.Equal(t, cfg, wi.GetConfig())
}

func TestWattenInteractor_HintAndLog(t *testing.T) {
	gameMock := new(interfaces.MockWattenGame)
	wpMock := new(presenter.MockWattenPresenter)
	wpMock.On("HintOutput", gameMock).Return(`{"hint":1}`)
	wpMock.On("ActionLogOutput", gameMock).Return(`{"log":1}`)

	wi := usecase.NewWattenInteractor(gameMock, wpMock)
	assert.Contains(t, wi.Hint(), "hint")
	assert.Contains(t, wi.ActionLog(), "log")
}

func TestWattenInteractor_RealReset(t *testing.T) {
	g := domain.NewDefaultWatten()
	wi := usecase.NewWattenInteractor(g, new(presenter.MockWattenPresenter))
	assert.NotNil(t, wi)
}
