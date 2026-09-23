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

func TestNewOmiInteractor_NilGuards(t *testing.T) {
	epMock := new(presenter.MockOmiPresenter)

	t.Run("panics when e is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "OmiInteractor: e must not be nil", func() {
			usecase.NewOmiInteractor(nil, epMock)
		})
	})

	t.Run("panics when ep is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockOmiGame)
		assert.PanicsWithValue(t, "OmiInteractor: ep must not be nil", func() {
			usecase.NewOmiInteractor(gameMock, nil)
		})
	})
}

func setupOmiInteractorMocks(phase domain.OmiPhase) (*interfaces.MockOmiGame, *presenter.MockOmiPresenter) {
	epMock := new(presenter.MockOmiPresenter)
	epMock.On("Output", mock.Anything, mock.Anything).Return(`{"phase":0}`)
	gameMock := new(interfaces.MockOmiGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(phase)
	gameMock.On("IsHumanCallTrumpTurn").Return(true)
	gameMock.On("IsHumanBidTurn").Return(true)
	gameMock.On("IsHumanTurn").Return(true)
	return gameMock, epMock
}

func TestOmiInteractor_Reset(t *testing.T) {
	t.Run("stays in call trump phase when human caller", func(t *testing.T) {
		gameMock, epMock := setupOmiInteractorMocks(domain.OmiPhaseCallTrump)
		gameMock.On("Reset").Return()

		ei := usecase.NewOmiInteractor(gameMock, epMock)
		result := ei.Reset()
		assert.Contains(t, result, "phase")
		gameMock.AssertCalled(t, "Reset")
	})

	t.Run("calls CpuCallTrump when CPU is caller", func(t *testing.T) {
		epMock := new(presenter.MockOmiPresenter)
		epMock.On("Output", mock.Anything, mock.Anything).Return(`{"phase":1}`)
		gameMock := new(interfaces.MockOmiGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.OmiPhaseCallTrump)
		gameMock.On("IsHumanCallTrumpTurn").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("Reset").Return()
		gameMock.On("CpuCallTrump").Return()

		ei := usecase.NewOmiInteractor(gameMock, epMock)
		result := ei.Reset()
		assert.Contains(t, result, "phase")
		gameMock.AssertCalled(t, "CpuCallTrump")
	})
}

func TestOmiInteractor_ResetWithConfig(t *testing.T) {
	t.Run("sets config then resets", func(t *testing.T) {
		gameMock, epMock := setupOmiInteractorMocks(domain.OmiPhaseCallTrump)
		cfg := domain.OmiConfig{CpuDifficulty: domain.OmiCpuDifficultyHard, PointLimit: 5}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()

		ei := usecase.NewOmiInteractor(gameMock, epMock)
		result := ei.ResetWithConfig(cfg)
		assert.Contains(t, result, "phase")
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config returns error", func(t *testing.T) {
		gameMock, epMock := setupOmiInteractorMocks(domain.OmiPhaseCallTrump)
		invalidCfg := domain.OmiConfig{CpuDifficulty: 99, PointLimit: 10}

		ei := usecase.NewOmiInteractor(gameMock, epMock)
		result := ei.ResetWithConfig(invalidCfg)
		assert.Contains(t, result, "phase")
	})
}

func TestOmiInteractor_CallTrump(t *testing.T) {
	t.Run("successful call", func(t *testing.T) {
		gameMock, epMock := setupOmiInteractorMocks(domain.OmiPhasePlay)
		gameMock.On("PlayerCallTrump", domain.CardDesignSpade).Return(nil)

		ei := usecase.NewOmiInteractor(gameMock, epMock)
		result := ei.CallTrump(domain.CardDesignSpade)
		assert.Contains(t, result, "phase")
		gameMock.AssertCalled(t, "PlayerCallTrump", domain.CardDesignSpade)
	})

	t.Run("call error", func(t *testing.T) {
		gameMock, epMock := setupOmiInteractorMocks(domain.OmiPhaseCallTrump)
		gameMock.On("PlayerCallTrump", 99).Return(errors.New("invalid suit"))

		ei := usecase.NewOmiInteractor(gameMock, epMock)
		result := ei.CallTrump(99)
		assert.Contains(t, result, "phase")
	})

	t.Run("game ended guard", func(t *testing.T) {
		epMock := new(presenter.MockOmiPresenter)
		epMock.On("Output", mock.Anything, mock.Anything).Return(`{"gameEnd":true}`)
		gameMock := new(interfaces.MockOmiGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ei := usecase.NewOmiInteractor(gameMock, epMock)
		result := ei.CallTrump(domain.CardDesignSpade)
		assert.Contains(t, result, "gameEnd")
	})
}

func TestOmiInteractor_Play(t *testing.T) {
	t.Run("successful play", func(t *testing.T) {
		gameMock, epMock := setupOmiInteractorMocks(domain.OmiPhasePlay)
		gameMock.On("PlayerPlay", 0).Return(nil)

		ei := usecase.NewOmiInteractor(gameMock, epMock)
		result := ei.Play(0)
		assert.Contains(t, result, "phase")
		gameMock.AssertCalled(t, "PlayerPlay", 0)
	})

	t.Run("play error", func(t *testing.T) {
		gameMock, epMock := setupOmiInteractorMocks(domain.OmiPhasePlay)
		gameMock.On("PlayerPlay", 0).Return(errors.New("follow suit"))

		ei := usecase.NewOmiInteractor(gameMock, epMock)
		result := ei.Play(0)
		assert.Contains(t, result, "phase")
	})

	t.Run("not human turn guard", func(t *testing.T) {
		epMock := new(presenter.MockOmiPresenter)
		epMock.On("Output", mock.Anything, mock.Anything).Return(`{"notHuman":true}`)
		gameMock := new(interfaces.MockOmiGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		ei := usecase.NewOmiInteractor(gameMock, epMock)
		result := ei.Play(0)
		assert.Contains(t, result, "notHuman")
	})

	t.Run("game ended guard", func(t *testing.T) {
		epMock := new(presenter.MockOmiPresenter)
		epMock.On("Output", mock.Anything, mock.Anything).Return(`{"gameEnd":true}`)
		gameMock := new(interfaces.MockOmiGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ei := usecase.NewOmiInteractor(gameMock, epMock)
		result := ei.Play(0)
		assert.Contains(t, result, "gameEnd")
	})
}

func TestOmiInteractor_NextTrick(t *testing.T) {
	t.Run("resolves and starts next trick", func(t *testing.T) {
		gameMock, epMock := setupOmiInteractorMocks(domain.OmiPhasePlay)
		gameMock.On("ResolveTrick").Return()
		gameMock.On("NextTrick").Return()

		ei := usecase.NewOmiInteractor(gameMock, epMock)
		result := ei.NextTrick()
		assert.Contains(t, result, "phase")
		gameMock.AssertCalled(t, "ResolveTrick")
		gameMock.AssertCalled(t, "NextTrick")
	})

	t.Run("game ended guard", func(t *testing.T) {
		epMock := new(presenter.MockOmiPresenter)
		epMock.On("Output", mock.Anything, mock.Anything).Return(`{"gameEnd":true}`)
		gameMock := new(interfaces.MockOmiGame)
		gameMock.On("ResolveTrick").Return()
		gameMock.On("GetGameEndFlag").Return(true)

		ei := usecase.NewOmiInteractor(gameMock, epMock)
		result := ei.NextTrick()
		assert.Contains(t, result, "gameEnd")
		gameMock.AssertCalled(t, "ResolveTrick")
	})
}

func TestOmiInteractor_NextRound(t *testing.T) {
	t.Run("scores and starts next round", func(t *testing.T) {
		gameMock, epMock := setupOmiInteractorMocks(domain.OmiPhaseCallTrump)
		gameMock.On("ScoreRound").Return()
		gameMock.On("NextRound").Return()

		ei := usecase.NewOmiInteractor(gameMock, epMock)
		result := ei.NextRound()
		assert.Contains(t, result, "phase")
		gameMock.AssertCalled(t, "ScoreRound")
		gameMock.AssertCalled(t, "NextRound")
	})

	t.Run("game ended guard", func(t *testing.T) {
		epMock := new(presenter.MockOmiPresenter)
		epMock.On("Output", mock.Anything, mock.Anything).Return(`{"gameEnd":true}`)
		gameMock := new(interfaces.MockOmiGame)
		gameMock.On("ScoreRound").Return()
		gameMock.On("GetGameEndFlag").Return(true)

		ei := usecase.NewOmiInteractor(gameMock, epMock)
		result := ei.NextRound()
		assert.Contains(t, result, "gameEnd")
		gameMock.AssertCalled(t, "ScoreRound")
	})
}

func TestOmiInteractor_GetConfig(t *testing.T) {
	gameMock, epMock := setupOmiInteractorMocks(domain.OmiPhaseCallTrump)
	cfg := domain.DefaultOmiConfig()
	gameMock.On("GetConfig").Return(cfg)

	ei := usecase.NewOmiInteractor(gameMock, epMock)
	assert.Equal(t, cfg, ei.GetConfig())
}

func TestOmiInteractor_Hint(t *testing.T) {
	gameMock, epMock := setupOmiInteractorMocks(domain.OmiPhaseCallTrump)
	epMock.On("HintOutput", gameMock).Return(`{"hint":"test"}`)

	ei := usecase.NewOmiInteractor(gameMock, epMock)
	result := ei.Hint()
	assert.Contains(t, result, "hint")
}

func TestOmiInteractor_ActionLog(t *testing.T) {
	gameMock, epMock := setupOmiInteractorMocks(domain.OmiPhaseCallTrump)
	epMock.On("ActionLogOutput", gameMock).Return(`[{"action":"test"}]`)

	ei := usecase.NewOmiInteractor(gameMock, epMock)
	result := ei.ActionLog()
	assert.Contains(t, result, "action")
}

func TestRestoreOmiInteractor(t *testing.T) {
	players := []*domain.OmiPlayer{
		domain.NewOmiPlayer(true, 0),
		domain.NewOmiPlayer(false, 1),
		domain.NewOmiPlayer(false, 0),
		domain.NewOmiPlayer(false, 1),
	}
	e := domain.NewOmi(domain.NewTrumpCards32(), players, domain.DefaultOmiConfig())
	epMock := new(presenter.MockOmiPresenter)
	ei := usecase.NewOmiInteractor(e, epMock)

	data, err := ei.Snapshot()
	assert.NoError(t, err)

	restored, err := usecase.RestoreOmiInteractor(data, epMock)
	assert.NoError(t, err)
	assert.NotNil(t, restored)
}

func TestRestoreOmiInteractor_InvalidJSON(t *testing.T) {
	epMock := new(presenter.MockOmiPresenter)
	_, err := usecase.RestoreOmiInteractor([]byte("invalid"), epMock)
	assert.Error(t, err)
}
