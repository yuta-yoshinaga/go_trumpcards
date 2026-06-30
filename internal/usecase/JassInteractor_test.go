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

func TestNewJassInteractor_NilGuards(t *testing.T) {
	jpMock := new(presenter.MockJassPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "JassInteractor: g must not be nil", func() {
			usecase.NewJassInteractor(nil, jpMock)
		})
	})

	t.Run("panics when jp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockJassGame)
		assert.PanicsWithValue(t, "JassInteractor: jp must not be nil", func() {
			usecase.NewJassInteractor(gameMock, nil)
		})
	})
}

func setupJassMocks(phase domain.JassPhase) (*interfaces.MockJassGame, *presenter.MockJassPresenter) {
	jpMock := new(presenter.MockJassPresenter)
	jpMock.On("Output", mock.Anything, mock.Anything).Return(`{"phase":0}`)
	gameMock := new(interfaces.MockJassGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(phase)
	gameMock.On("IsHumanBidTurn").Return(true)
	gameMock.On("IsHumanTurn").Return(true)
	return gameMock, jpMock
}

func TestJassInteractor_Reset(t *testing.T) {
	gameMock, jpMock := setupJassMocks(domain.JassPhaseBidTrump)
	gameMock.On("Reset").Return()

	ji := usecase.NewJassInteractor(gameMock, jpMock)
	result := ji.Reset()
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "Reset")
}

func TestJassInteractor_ResetWithConfig(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		gameMock, jpMock := setupJassMocks(domain.JassPhaseBidTrump)
		cfg := domain.DefaultJassConfig()
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()

		ji := usecase.NewJassInteractor(gameMock, jpMock)
		result := ji.ResetWithConfig(cfg)
		assert.Contains(t, result, "phase")
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config", func(t *testing.T) {
		gameMock, jpMock := setupJassMocks(domain.JassPhaseBidTrump)
		bad := domain.JassConfig{CpuDifficulty: 99, TargetScore: 1000, LastTrickBonus: 5}

		ji := usecase.NewJassInteractor(gameMock, jpMock)
		result := ji.ResetWithConfig(bad)
		assert.Contains(t, result, "phase")
	})
}

func TestJassInteractor_ChooseTrump(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		gameMock, jpMock := setupJassMocks(domain.JassPhasePlay)
		gameMock.On("PlayerChooseTrump", domain.CardDesignSpade).Return(nil)

		ji := usecase.NewJassInteractor(gameMock, jpMock)
		assert.Contains(t, ji.ChooseTrump(domain.CardDesignSpade), "phase")
	})

	t.Run("error", func(t *testing.T) {
		gameMock, jpMock := setupJassMocks(domain.JassPhaseBidTrump)
		gameMock.On("PlayerChooseTrump", 99).Return(errors.New("bad"))

		ji := usecase.NewJassInteractor(gameMock, jpMock)
		assert.Contains(t, ji.ChooseTrump(99), "phase")
	})

	t.Run("game ended", func(t *testing.T) {
		jpMock := new(presenter.MockJassPresenter)
		jpMock.On("Output", mock.Anything, mock.Anything).Return(`{"gameEnd":true}`)
		gameMock := new(interfaces.MockJassGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ji := usecase.NewJassInteractor(gameMock, jpMock)
		assert.Contains(t, ji.ChooseTrump(domain.CardDesignSpade), "gameEnd")
	})
}

func TestJassInteractor_Schieben(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		gameMock, jpMock := setupJassMocks(domain.JassPhaseBidPartner)
		gameMock.On("PlayerSchieben").Return(nil)
		gameMock.On("CpuBid").Return()

		ji := usecase.NewJassInteractor(gameMock, jpMock)
		assert.Contains(t, ji.Schieben(), "phase")
	})

	t.Run("error", func(t *testing.T) {
		gameMock, jpMock := setupJassMocks(domain.JassPhaseBidTrump)
		gameMock.On("PlayerSchieben").Return(errors.New("bad"))

		ji := usecase.NewJassInteractor(gameMock, jpMock)
		assert.Contains(t, ji.Schieben(), "phase")
	})

	t.Run("game ended", func(t *testing.T) {
		jpMock := new(presenter.MockJassPresenter)
		jpMock.On("Output", mock.Anything, mock.Anything).Return(`{"gameEnd":true}`)
		gameMock := new(interfaces.MockJassGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ji := usecase.NewJassInteractor(gameMock, jpMock)
		assert.Contains(t, ji.Schieben(), "gameEnd")
	})
}

func TestJassInteractor_Play(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		gameMock, jpMock := setupJassMocks(domain.JassPhasePlay)
		gameMock.On("PlayerPlay", 0).Return(nil)

		ji := usecase.NewJassInteractor(gameMock, jpMock)
		assert.Contains(t, ji.Play(0), "phase")
	})

	t.Run("error", func(t *testing.T) {
		gameMock, jpMock := setupJassMocks(domain.JassPhasePlay)
		gameMock.On("PlayerPlay", 99).Return(errors.New("bad"))

		ji := usecase.NewJassInteractor(gameMock, jpMock)
		assert.Contains(t, ji.Play(99), "phase")
	})

	t.Run("game ended", func(t *testing.T) {
		jpMock := new(presenter.MockJassPresenter)
		jpMock.On("Output", mock.Anything, mock.Anything).Return(`{"gameEnd":true}`)
		gameMock := new(interfaces.MockJassGame)
		gameMock.On("GetGameEndFlag").Return(true)
		gameMock.On("GetPhase").Return(domain.JassPhaseGameEnd)

		ji := usecase.NewJassInteractor(gameMock, jpMock)
		assert.Contains(t, ji.Play(0), "gameEnd")
	})
}

func TestJassInteractor_NextTrick(t *testing.T) {
	gameMock, jpMock := setupJassMocks(domain.JassPhasePlay)
	gameMock.On("ResolveTrick").Return()
	gameMock.On("NextTrick").Return()

	ji := usecase.NewJassInteractor(gameMock, jpMock)
	assert.Contains(t, ji.NextTrick(), "phase")
}

func TestJassInteractor_NextRound(t *testing.T) {
	gameMock, jpMock := setupJassMocks(domain.JassPhaseBidTrump)
	gameMock.On("ScoreRound").Return()
	gameMock.On("NextRound").Return()

	ji := usecase.NewJassInteractor(gameMock, jpMock)
	assert.Contains(t, ji.NextRound(), "phase")
}

func TestJassInteractor_GetConfig(t *testing.T) {
	gameMock := new(interfaces.MockJassGame)
	cfg := domain.DefaultJassConfig()
	gameMock.On("GetConfig").Return(cfg)
	jpMock := new(presenter.MockJassPresenter)

	ji := usecase.NewJassInteractor(gameMock, jpMock)
	assert.Equal(t, cfg, ji.GetConfig())
}

func TestJassInteractor_HintAndLog(t *testing.T) {
	gameMock := new(interfaces.MockJassGame)
	jpMock := new(presenter.MockJassPresenter)
	jpMock.On("HintOutput", gameMock).Return(`{"hint":1}`)
	jpMock.On("ActionLogOutput", gameMock).Return(`{"log":1}`)

	ji := usecase.NewJassInteractor(gameMock, jpMock)
	assert.Contains(t, ji.Hint(), "hint")
	assert.Contains(t, ji.ActionLog(), "log")
}

func TestJassInteractor_RealReset(t *testing.T) {
	g := domain.NewDefaultJass()
	ji := usecase.NewJassInteractor(g, new(presenter.MockJassPresenter))
	_ = ji // ensure real game satisfies the interface
	assert.NotNil(t, ji)
}
