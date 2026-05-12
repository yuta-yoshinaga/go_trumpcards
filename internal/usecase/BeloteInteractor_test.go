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

func TestNewBeloteInteractor_NilGuards(t *testing.T) {
	bpMock := new(presenter.MockBelotePresenter)

	t.Run("panics when b is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "BeloteInteractor: b must not be nil", func() {
			usecase.NewBeloteInteractor(nil, bpMock)
		})
	})

	t.Run("panics when bp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockBeloteGame)
		assert.PanicsWithValue(t, "BeloteInteractor: bp must not be nil", func() {
			usecase.NewBeloteInteractor(gameMock, nil)
		})
	})
}

func setupBeloteInteractorMocks(phase domain.BelotePhase) (*interfaces.MockBeloteGame, *presenter.MockBelotePresenter) {
	bpMock := new(presenter.MockBelotePresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(`{"phase":0}`)
	gameMock := new(interfaces.MockBeloteGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(phase)
	gameMock.On("IsHumanBidTurn").Return(true)
	gameMock.On("IsHumanTurn").Return(true)
	return gameMock, bpMock
}

func TestBeloteInteractor_Reset(t *testing.T) {
	gameMock, bpMock := setupBeloteInteractorMocks(domain.BelotePhaseBidPickUp)
	gameMock.On("Reset").Return()

	bi := usecase.NewBeloteInteractor(gameMock, bpMock)
	result := bi.Reset()
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "Reset")
}

func TestBeloteInteractor_ResetWithConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		gameMock, bpMock := setupBeloteInteractorMocks(domain.BelotePhaseBidPickUp)
		cfg := domain.BeloteConfig{
			CpuDifficulty:        domain.BeloteCpuDifficultyHard,
			TargetScore:          500,
			DixDeDer:             10,
			EnableBeloteRebelote: true,
		}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()

		bi := usecase.NewBeloteInteractor(gameMock, bpMock)
		result := bi.ResetWithConfig(cfg)
		assert.Contains(t, result, "phase")
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config returns presenter output", func(t *testing.T) {
		gameMock, bpMock := setupBeloteInteractorMocks(domain.BelotePhaseBidPickUp)
		invalidCfg := domain.BeloteConfig{
			CpuDifficulty: 99,
			TargetScore:   1000,
			DixDeDer:      10,
		}

		bi := usecase.NewBeloteInteractor(gameMock, bpMock)
		result := bi.ResetWithConfig(invalidCfg)
		assert.Contains(t, result, "phase")
	})
}

func TestBeloteInteractor_PickUp(t *testing.T) {
	t.Run("successful", func(t *testing.T) {
		gameMock, bpMock := setupBeloteInteractorMocks(domain.BelotePhaseBidPickUp)
		gameMock.On("PlayerPickUp", true).Return(nil)

		bi := usecase.NewBeloteInteractor(gameMock, bpMock)
		result := bi.PickUp(true)
		assert.Contains(t, result, "phase")
	})

	t.Run("returns error", func(t *testing.T) {
		gameMock, bpMock := setupBeloteInteractorMocks(domain.BelotePhaseBidPickUp)
		gameMock.On("PlayerPickUp", true).Return(errors.New("test"))

		bi := usecase.NewBeloteInteractor(gameMock, bpMock)
		result := bi.PickUp(true)
		assert.Contains(t, result, "phase")
	})

	t.Run("game ended", func(t *testing.T) {
		bpMock := new(presenter.MockBelotePresenter)
		bpMock.On("Output", mock.Anything, mock.Anything).Return(`{"gameEnd":true}`)
		gameMock := new(interfaces.MockBeloteGame)
		gameMock.On("GetGameEndFlag").Return(true)

		bi := usecase.NewBeloteInteractor(gameMock, bpMock)
		result := bi.PickUp(true)
		assert.Contains(t, result, "gameEnd")
	})
}

func TestBeloteInteractor_CallTrump(t *testing.T) {
	t.Run("successful", func(t *testing.T) {
		gameMock, bpMock := setupBeloteInteractorMocks(domain.BelotePhasePlay)
		gameMock.On("PlayerCallTrump", domain.CardDesignSpade).Return(nil)

		bi := usecase.NewBeloteInteractor(gameMock, bpMock)
		result := bi.CallTrump(domain.CardDesignSpade)
		assert.Contains(t, result, "phase")
	})

	t.Run("error", func(t *testing.T) {
		gameMock, bpMock := setupBeloteInteractorMocks(domain.BelotePhaseBidCallTrump)
		gameMock.On("PlayerCallTrump", 99).Return(errors.New("invalid suit"))

		bi := usecase.NewBeloteInteractor(gameMock, bpMock)
		result := bi.CallTrump(99)
		assert.Contains(t, result, "phase")
	})

	t.Run("game ended", func(t *testing.T) {
		bpMock := new(presenter.MockBelotePresenter)
		bpMock.On("Output", mock.Anything, mock.Anything).Return(`{"gameEnd":true}`)
		gameMock := new(interfaces.MockBeloteGame)
		gameMock.On("GetGameEndFlag").Return(true)

		bi := usecase.NewBeloteInteractor(gameMock, bpMock)
		result := bi.CallTrump(domain.CardDesignSpade)
		assert.Contains(t, result, "gameEnd")
	})
}

func TestBeloteInteractor_PassCall(t *testing.T) {
	gameMock, bpMock := setupBeloteInteractorMocks(domain.BelotePhaseBidPickUp)
	gameMock.On("PlayerPassCall").Return(nil)

	bi := usecase.NewBeloteInteractor(gameMock, bpMock)
	result := bi.PassCall()
	assert.Contains(t, result, "phase")
}

func TestBeloteInteractor_PassCall_GameEnd(t *testing.T) {
	bpMock := new(presenter.MockBelotePresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(`{"gameEnd":true}`)
	gameMock := new(interfaces.MockBeloteGame)
	gameMock.On("GetGameEndFlag").Return(true)

	bi := usecase.NewBeloteInteractor(gameMock, bpMock)
	result := bi.PassCall()
	assert.Contains(t, result, "gameEnd")
}

func TestBeloteInteractor_Pass(t *testing.T) {
	t.Run("pickup phase delegates to PickUp", func(t *testing.T) {
		gameMock, bpMock := setupBeloteInteractorMocks(domain.BelotePhaseBidPickUp)
		gameMock.On("PlayerPickUp", false).Return(nil)

		bi := usecase.NewBeloteInteractor(gameMock, bpMock)
		result := bi.Pass()
		assert.Contains(t, result, "phase")
	})

	t.Run("call trump phase delegates to PassCall", func(t *testing.T) {
		gameMock, bpMock := setupBeloteInteractorMocks(domain.BelotePhaseBidCallTrump)
		gameMock.On("PlayerPassCall").Return(nil)

		bi := usecase.NewBeloteInteractor(gameMock, bpMock)
		result := bi.Pass()
		assert.Contains(t, result, "phase")
	})

	t.Run("wrong phase returns error output", func(t *testing.T) {
		gameMock, bpMock := setupBeloteInteractorMocks(domain.BelotePhasePlay)

		bi := usecase.NewBeloteInteractor(gameMock, bpMock)
		result := bi.Pass()
		assert.Contains(t, result, "phase")
	})

	t.Run("game ended", func(t *testing.T) {
		bpMock := new(presenter.MockBelotePresenter)
		bpMock.On("Output", mock.Anything, mock.Anything).Return(`{"gameEnd":true}`)
		gameMock := new(interfaces.MockBeloteGame)
		gameMock.On("GetGameEndFlag").Return(true)

		bi := usecase.NewBeloteInteractor(gameMock, bpMock)
		result := bi.Pass()
		assert.Contains(t, result, "gameEnd")
	})
}

func TestBeloteInteractor_Play(t *testing.T) {
	t.Run("successful", func(t *testing.T) {
		gameMock, bpMock := setupBeloteInteractorMocks(domain.BelotePhasePlay)
		gameMock.On("PlayerPlay", 0).Return(nil)

		bi := usecase.NewBeloteInteractor(gameMock, bpMock)
		result := bi.Play(0)
		assert.Contains(t, result, "phase")
	})

	t.Run("error", func(t *testing.T) {
		gameMock, bpMock := setupBeloteInteractorMocks(domain.BelotePhasePlay)
		gameMock.On("PlayerPlay", 0).Return(errors.New("must follow"))

		bi := usecase.NewBeloteInteractor(gameMock, bpMock)
		result := bi.Play(0)
		assert.Contains(t, result, "phase")
	})

	t.Run("not human turn", func(t *testing.T) {
		bpMock := new(presenter.MockBelotePresenter)
		bpMock.On("Output", mock.Anything, mock.Anything).Return(`{"notHuman":true}`)
		gameMock := new(interfaces.MockBeloteGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		bi := usecase.NewBeloteInteractor(gameMock, bpMock)
		result := bi.Play(0)
		assert.Contains(t, result, "notHuman")
	})
}

func TestBeloteInteractor_NextTrick(t *testing.T) {
	gameMock, bpMock := setupBeloteInteractorMocks(domain.BelotePhasePlay)
	gameMock.On("ResolveTrick").Return()
	gameMock.On("NextTrick").Return()

	bi := usecase.NewBeloteInteractor(gameMock, bpMock)
	result := bi.NextTrick()
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "ResolveTrick")
	gameMock.AssertCalled(t, "NextTrick")
}

func TestBeloteInteractor_NextTrick_GameEnd(t *testing.T) {
	bpMock := new(presenter.MockBelotePresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(`{"gameEnd":true}`)
	gameMock := new(interfaces.MockBeloteGame)
	gameMock.On("ResolveTrick").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	bi := usecase.NewBeloteInteractor(gameMock, bpMock)
	result := bi.NextTrick()
	assert.Contains(t, result, "gameEnd")
}

func TestBeloteInteractor_NextRound(t *testing.T) {
	gameMock, bpMock := setupBeloteInteractorMocks(domain.BelotePhaseBidPickUp)
	gameMock.On("ScoreRound").Return()
	gameMock.On("NextRound").Return()

	bi := usecase.NewBeloteInteractor(gameMock, bpMock)
	result := bi.NextRound()
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "ScoreRound")
	gameMock.AssertCalled(t, "NextRound")
}

func TestBeloteInteractor_NextRound_GameEnd(t *testing.T) {
	bpMock := new(presenter.MockBelotePresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(`{"gameEnd":true}`)
	gameMock := new(interfaces.MockBeloteGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	bi := usecase.NewBeloteInteractor(gameMock, bpMock)
	result := bi.NextRound()
	assert.Contains(t, result, "gameEnd")
}

func TestBeloteInteractor_GetConfig(t *testing.T) {
	gameMock, bpMock := setupBeloteInteractorMocks(domain.BelotePhaseBidPickUp)
	cfg := domain.DefaultBeloteConfig()
	gameMock.On("GetConfig").Return(cfg)

	bi := usecase.NewBeloteInteractor(gameMock, bpMock)
	assert.Equal(t, cfg, bi.GetConfig())
}

func TestBeloteInteractor_Hint(t *testing.T) {
	gameMock, bpMock := setupBeloteInteractorMocks(domain.BelotePhaseBidPickUp)
	bpMock.On("HintOutput", gameMock).Return(`{"hint":"x"}`)

	bi := usecase.NewBeloteInteractor(gameMock, bpMock)
	assert.Contains(t, bi.Hint(), "hint")
}

func TestBeloteInteractor_ActionLog(t *testing.T) {
	gameMock, bpMock := setupBeloteInteractorMocks(domain.BelotePhaseBidPickUp)
	bpMock.On("ActionLogOutput", gameMock).Return(`[{"a":"b"}]`)

	bi := usecase.NewBeloteInteractor(gameMock, bpMock)
	assert.Contains(t, bi.ActionLog(), "a")
}

func TestRestoreBeloteInteractor(t *testing.T) {
	players := []*domain.BelotePlayer{
		domain.NewBelotePlayer(true, 0),
		domain.NewBelotePlayer(false, 1),
		domain.NewBelotePlayer(false, 0),
		domain.NewBelotePlayer(false, 1),
	}
	b := domain.NewBelote(domain.NewTrumpCardsBelote(), players, domain.DefaultBeloteConfig())
	bpMock := new(presenter.MockBelotePresenter)
	bi := usecase.NewBeloteInteractor(b, bpMock)

	data, err := bi.Snapshot()
	assert.NoError(t, err)

	restored, err := usecase.RestoreBeloteInteractor(data, bpMock)
	assert.NoError(t, err)
	assert.NotNil(t, restored)
}

func TestRestoreBeloteInteractor_InvalidJSON(t *testing.T) {
	bpMock := new(presenter.MockBelotePresenter)
	_, err := usecase.RestoreBeloteInteractor([]byte("invalid"), bpMock)
	assert.Error(t, err)
}
