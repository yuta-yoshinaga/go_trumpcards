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

func TestNewPiquetInteractor_NilGuards(t *testing.T) {
	ppMock := new(presenter.MockPiquetPresenter)

	t.Run("panics when p is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "PiquetInteractor: p must not be nil", func() {
			usecase.NewPiquetInteractor(nil, ppMock)
		})
	})

	t.Run("panics when pp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockPiquetGame)
		assert.PanicsWithValue(t, "PiquetInteractor: pp must not be nil", func() {
			usecase.NewPiquetInteractor(gameMock, nil)
		})
	})
}

// setupPiquetMocks sets up the mock game in a phase where CPU autorun
// will not engage (returns IsHumanTurn=true and ExchangeTurn=Done so
// neither autorun loop will progress).
func setupPiquetMocks(phase domain.PiquetPhase) (*interfaces.MockPiquetGame, *presenter.MockPiquetPresenter) {
	ppMock := new(presenter.MockPiquetPresenter)
	ppMock.On("Output", mock.Anything, mock.Anything).Return(`{"phase":0}`)
	gameMock := new(interfaces.MockPiquetGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(phase)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("GetExchangeTurn").Return(domain.PiquetExchangeTurnDone)
	return gameMock, ppMock
}

func TestPiquetInteractor_Reset(t *testing.T) {
	gameMock, ppMock := setupPiquetMocks(domain.PiquetPhaseExchange)
	gameMock.On("Reset").Return()

	pi := usecase.NewPiquetInteractor(gameMock, ppMock)
	result := pi.Reset()
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "Reset")
}

func TestPiquetInteractor_ResetWithConfig_Valid(t *testing.T) {
	gameMock, ppMock := setupPiquetMocks(domain.PiquetPhaseExchange)
	cfg := domain.PiquetConfig{CpuDifficulty: domain.PiquetCpuDifficultyHard, DealsPerPartie: 6}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()

	pi := usecase.NewPiquetInteractor(gameMock, ppMock)
	result := pi.ResetWithConfig(cfg)
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestPiquetInteractor_ResetWithConfig_Invalid(t *testing.T) {
	gameMock, ppMock := setupPiquetMocks(domain.PiquetPhaseExchange)
	invalid := domain.PiquetConfig{CpuDifficulty: domain.PiquetCpuDifficultyNormal, DealsPerPartie: 0}

	pi := usecase.NewPiquetInteractor(gameMock, ppMock)
	result := pi.ResetWithConfig(invalid)
	assert.Contains(t, result, "phase")
	gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
}

func TestPiquetInteractor_ExchangeElder(t *testing.T) {
	gameMock, ppMock := setupPiquetMocks(domain.PiquetPhaseExchange)
	gameMock.On("ExchangeElder", []int{0, 1, 2}).Return(nil)

	pi := usecase.NewPiquetInteractor(gameMock, ppMock)
	result := pi.ExchangeElder([]int{0, 1, 2})
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "ExchangeElder", []int{0, 1, 2})
}

func TestPiquetInteractor_ExchangeElder_Error(t *testing.T) {
	gameMock, ppMock := setupPiquetMocks(domain.PiquetPhaseExchange)
	gameMock.On("ExchangeElder", []int{}).Return(errors.New("must exchange 1..5"))

	pi := usecase.NewPiquetInteractor(gameMock, ppMock)
	result := pi.ExchangeElder([]int{})
	assert.Contains(t, result, "phase")
}

func TestPiquetInteractor_ExchangeElder_GameEndShortCircuit(t *testing.T) {
	ppMock := new(presenter.MockPiquetPresenter)
	ppMock.On("Output", mock.Anything, mock.Anything).Return(`done`)
	gameMock := new(interfaces.MockPiquetGame)
	gameMock.On("GetGameEndFlag").Return(true)

	pi := usecase.NewPiquetInteractor(gameMock, ppMock)
	result := pi.ExchangeElder([]int{0})
	assert.Equal(t, "done", result)
	gameMock.AssertNotCalled(t, "ExchangeElder", mock.Anything)
}

func TestPiquetInteractor_ExchangeYounger(t *testing.T) {
	gameMock, ppMock := setupPiquetMocks(domain.PiquetPhaseExchange)
	gameMock.On("ExchangeYounger", []int{0}).Return(nil)

	pi := usecase.NewPiquetInteractor(gameMock, ppMock)
	result := pi.ExchangeYounger([]int{0})
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "ExchangeYounger", []int{0})
}

func TestPiquetInteractor_ResolveDeclaration(t *testing.T) {
	gameMock, ppMock := setupPiquetMocks(domain.PiquetPhaseDeclaration)
	gameMock.On("ResolveDeclaration").Return(&domain.PiquetDeclarationResult{Score: 4}, nil)

	pi := usecase.NewPiquetInteractor(gameMock, ppMock)
	result := pi.ResolveDeclaration()
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "ResolveDeclaration")
}

func TestPiquetInteractor_ResolveDeclaration_Error(t *testing.T) {
	gameMock, ppMock := setupPiquetMocks(domain.PiquetPhaseExchange)
	gameMock.On("ResolveDeclaration").Return((*domain.PiquetDeclarationResult)(nil), errors.New("wrong phase"))

	pi := usecase.NewPiquetInteractor(gameMock, ppMock)
	result := pi.ResolveDeclaration()
	assert.Contains(t, result, "phase")
}

func TestPiquetInteractor_Play(t *testing.T) {
	gameMock, ppMock := setupPiquetMocks(domain.PiquetPhasePlay)
	gameMock.On("PlayCard", 3).Return(nil)

	pi := usecase.NewPiquetInteractor(gameMock, ppMock)
	result := pi.Play(3)
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "PlayCard", 3)
}

func TestPiquetInteractor_Play_NotHumanTurn(t *testing.T) {
	ppMock := new(presenter.MockPiquetPresenter)
	ppMock.On("Output", mock.Anything, mock.Anything).Return(`waiting`)
	gameMock := new(interfaces.MockPiquetGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	pi := usecase.NewPiquetInteractor(gameMock, ppMock)
	result := pi.Play(0)
	assert.Equal(t, "waiting", result)
	gameMock.AssertNotCalled(t, "PlayCard", mock.Anything)
}

func TestPiquetInteractor_Play_Error(t *testing.T) {
	gameMock, ppMock := setupPiquetMocks(domain.PiquetPhasePlay)
	gameMock.On("PlayCard", 0).Return(errors.New("illegal play"))

	pi := usecase.NewPiquetInteractor(gameMock, ppMock)
	result := pi.Play(0)
	assert.Contains(t, result, "phase")
}

func TestPiquetInteractor_NextDeal(t *testing.T) {
	gameMock, ppMock := setupPiquetMocks(domain.PiquetPhaseScore)
	gameMock.On("NextDeal").Return()

	pi := usecase.NewPiquetInteractor(gameMock, ppMock)
	result := pi.NextDeal()
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "NextDeal")
}

func TestPiquetInteractor_NextDeal_GameEnd(t *testing.T) {
	ppMock := new(presenter.MockPiquetPresenter)
	ppMock.On("Output", mock.Anything, mock.Anything).Return(`finished`)
	gameMock := new(interfaces.MockPiquetGame)
	gameMock.On("GetGameEndFlag").Return(true)

	pi := usecase.NewPiquetInteractor(gameMock, ppMock)
	result := pi.NextDeal()
	assert.Equal(t, "finished", result)
	gameMock.AssertNotCalled(t, "NextDeal")
}

func TestPiquetInteractor_Hint(t *testing.T) {
	ppMock := new(presenter.MockPiquetPresenter)
	ppMock.On("HintOutput", mock.Anything).Return(`{"hint":"play"}`)
	gameMock := new(interfaces.MockPiquetGame)

	pi := usecase.NewPiquetInteractor(gameMock, ppMock)
	result := pi.Hint()
	assert.Equal(t, `{"hint":"play"}`, result)
	ppMock.AssertCalled(t, "HintOutput", mock.Anything)
}

func TestPiquetInteractor_ActionLog(t *testing.T) {
	ppMock := new(presenter.MockPiquetPresenter)
	ppMock.On("ActionLogOutput", mock.Anything).Return(`[]`)
	gameMock := new(interfaces.MockPiquetGame)

	pi := usecase.NewPiquetInteractor(gameMock, ppMock)
	result := pi.ActionLog()
	assert.Equal(t, "[]", result)
}

func TestPiquetInteractor_GetConfig(t *testing.T) {
	gameMock := new(interfaces.MockPiquetGame)
	gameMock.On("GetConfig").Return(domain.PiquetConfig{DealsPerPartie: 6})
	ppMock := new(presenter.MockPiquetPresenter)

	pi := usecase.NewPiquetInteractor(gameMock, ppMock)
	cfg := pi.GetConfig()
	assert.Equal(t, 6, cfg.DealsPerPartie)
}

// Live-domain integration smoke: Reset → exchange → declarations → play through a real Piquet.
func TestPiquetInteractor_RealGameFlow(t *testing.T) {
	players := []*domain.PiquetPlayer{
		domain.NewPiquetPlayer(false), // both CPU so the interactor auto-advances
		domain.NewPiquetPlayer(false),
	}
	game := domain.NewPiquet(domain.NewTrumpCardsBelote(), players,
		domain.PiquetConfig{DealsPerPartie: 1, CpuDifficulty: domain.PiquetCpuDifficultyNormal})
	pp := new(presenter.MockPiquetPresenter)
	pp.On("Output", mock.Anything, mock.Anything).Return(`out`)
	pp.On("HintOutput", mock.Anything).Return(`hint`)

	pi := usecase.NewPiquetInteractor(game, pp)
	pi.Reset()
	// Both players are CPU → exchange phase autoruns to declaration
	assert.Equal(t, domain.PiquetPhaseDeclaration, game.GetPhase())

	// Resolve all 3 declarations (Point/Sequence/Set)
	pi.ResolveDeclaration()
	pi.ResolveDeclaration()
	pi.ResolveDeclaration()
	// After all 3, transitioned to play, then runCpuPlay drove the whole deal home
	assert.True(t, game.GetPhase() == domain.PiquetPhaseScore || game.GetPhase() == domain.PiquetPhaseGameEnd)
}

// Snapshot/Restore round trip
func TestPiquetInteractor_SnapshotRestore(t *testing.T) {
	players := []*domain.PiquetPlayer{
		domain.NewPiquetPlayer(true),
		domain.NewPiquetPlayer(false),
	}
	game := domain.NewPiquet(domain.NewTrumpCardsBelote(), players, domain.DefaultPiquetConfig())
	pp := new(presenter.MockPiquetPresenter)
	pp.On("Output", mock.Anything, mock.Anything).Return(`out`)

	pi := usecase.NewPiquetInteractor(game, pp)
	pi.Reset()
	data, err := pi.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	restored, err := usecase.RestorePiquetInteractor(data, pp)
	assert.NoError(t, err)
	assert.NotNil(t, restored)
	assert.Equal(t, game.GetConfig().DealsPerPartie, restored.GetConfig().DealsPerPartie)
}
