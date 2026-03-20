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

func TestNewGinRummyInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockGinRummyPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "GinRummyInteractor: g must not be nil", func() {
			usecase.NewGinRummyInteractor(nil, pMock)
		})
	})

	t.Run("panics when gp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockGinRummyGame)
		assert.PanicsWithValue(t, "GinRummyInteractor: gp must not be nil", func() {
			usecase.NewGinRummyInteractor(gameMock, nil)
		})
	})
}

func TestGinRummyInteractor_Reset(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("reset calls game Reset and returns output", func(t *testing.T) {
		pMock := new(presenter.MockGinRummyPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockGinRummyGame)
		gameMock.On("Reset").Return()
		// runCpuTurns: not ended, phase=Draw, human turn → break
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.GinRummyPhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewGinRummyInteractor(gameMock, pMock)
		result := ci.Reset()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "Reset")
	})
}

func TestGinRummyInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("valid config sets config then resets", func(t *testing.T) {
		pMock := new(presenter.MockGinRummyPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockGinRummyGame)
		cfg := domain.GinRummyConfig{CpuDifficulty: domain.GinRummyCpuDifficultyHard, PointLimit: 200}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.GinRummyPhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewGinRummyInteractor(gameMock, pMock)
		result := ci.ResetWithConfig(cfg)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "SetConfig", cfg)
		gameMock.AssertCalled(t, "Reset")
	})

	t.Run("invalid config returns error without resetting", func(t *testing.T) {
		pMock := new(presenter.MockGinRummyPresenter)
		gameMock := new(interfaces.MockGinRummyGame)
		pMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")

		ci := usecase.NewGinRummyInteractor(gameMock, pMock)
		cfg := domain.GinRummyConfig{CpuDifficulty: domain.GinRummyCpuDifficulty(-1), PointLimit: 100}
		result := ci.ResetWithConfig(cfg)
		assert.Equal(t, "validation error", result)
		gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
	})
}

func TestGinRummyInteractor_DrawFromStock(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ended returns output without drawing", func(t *testing.T) {
		pMock := new(presenter.MockGinRummyPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockGinRummyGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewGinRummyInteractor(gameMock, pMock)
		result := ci.DrawFromStock()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerDrawFromStock")
	})

	t.Run("not human turn returns output without drawing", func(t *testing.T) {
		pMock := new(presenter.MockGinRummyPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockGinRummyGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		ci := usecase.NewGinRummyInteractor(gameMock, pMock)
		result := ci.DrawFromStock()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerDrawFromStock")
	})

	t.Run("draw error is returned to presenter", func(t *testing.T) {
		drawErr := errors.New("draw error")
		pMock := new(presenter.MockGinRummyPresenter)
		pMock.On("Output", mock.Anything, drawErr).Return(mockOutput)
		gameMock := new(interfaces.MockGinRummyGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDrawFromStock").Return(drawErr)

		ci := usecase.NewGinRummyInteractor(gameMock, pMock)
		result := ci.DrawFromStock()
		assert.Equal(t, mockOutput, result)
	})

	t.Run("valid draw runs CPU turns", func(t *testing.T) {
		pMock := new(presenter.MockGinRummyPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockGinRummyGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerDrawFromStock").Return(nil)
		// runCpuTurns: human turn → break
		gameMock.On("GetPhase").Return(domain.GinRummyPhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewGinRummyInteractor(gameMock, pMock)
		result := ci.DrawFromStock()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerDrawFromStock")
	})
}

func TestGinRummyInteractor_DrawFromDiscard(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ended returns output", func(t *testing.T) {
		pMock := new(presenter.MockGinRummyPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockGinRummyGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewGinRummyInteractor(gameMock, pMock)
		result := ci.DrawFromDiscard()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerDrawFromDiscard")
	})

	t.Run("not human turn returns output", func(t *testing.T) {
		pMock := new(presenter.MockGinRummyPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockGinRummyGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		ci := usecase.NewGinRummyInteractor(gameMock, pMock)
		result := ci.DrawFromDiscard()
		assert.Equal(t, mockOutput, result)
	})

	t.Run("error is returned to presenter", func(t *testing.T) {
		drawErr := errors.New("draw error")
		pMock := new(presenter.MockGinRummyPresenter)
		pMock.On("Output", mock.Anything, drawErr).Return(mockOutput)
		gameMock := new(interfaces.MockGinRummyGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDrawFromDiscard").Return(drawErr)

		ci := usecase.NewGinRummyInteractor(gameMock, pMock)
		result := ci.DrawFromDiscard()
		assert.Equal(t, mockOutput, result)
	})

	t.Run("valid draw runs CPU turns", func(t *testing.T) {
		pMock := new(presenter.MockGinRummyPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockGinRummyGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerDrawFromDiscard").Return(nil)
		gameMock.On("GetPhase").Return(domain.GinRummyPhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewGinRummyInteractor(gameMock, pMock)
		result := ci.DrawFromDiscard()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerDrawFromDiscard")
	})
}

func TestGinRummyInteractor_Discard(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ended returns output", func(t *testing.T) {
		pMock := new(presenter.MockGinRummyPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockGinRummyGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewGinRummyInteractor(gameMock, pMock)
		result := ci.Discard(0)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerDiscard")
	})

	t.Run("not human turn returns output", func(t *testing.T) {
		pMock := new(presenter.MockGinRummyPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockGinRummyGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		ci := usecase.NewGinRummyInteractor(gameMock, pMock)
		result := ci.Discard(0)
		assert.Equal(t, mockOutput, result)
	})

	t.Run("error is returned", func(t *testing.T) {
		discardErr := errors.New("discard error")
		pMock := new(presenter.MockGinRummyPresenter)
		pMock.On("Output", mock.Anything, discardErr).Return(mockOutput)
		gameMock := new(interfaces.MockGinRummyGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDiscard", 0).Return(discardErr)

		ci := usecase.NewGinRummyInteractor(gameMock, pMock)
		result := ci.Discard(0)
		assert.Equal(t, mockOutput, result)
	})

	t.Run("valid discard runs CPU turns", func(t *testing.T) {
		pMock := new(presenter.MockGinRummyPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockGinRummyGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerDiscard", 3).Return(nil)
		gameMock.On("GetPhase").Return(domain.GinRummyPhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewGinRummyInteractor(gameMock, pMock)
		result := ci.Discard(3)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerDiscard", 3)
	})
}

func TestGinRummyInteractor_Knock(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ended returns output", func(t *testing.T) {
		pMock := new(presenter.MockGinRummyPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockGinRummyGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewGinRummyInteractor(gameMock, pMock)
		result := ci.Knock(0)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerKnock")
	})

	t.Run("not human turn returns output", func(t *testing.T) {
		pMock := new(presenter.MockGinRummyPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockGinRummyGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		ci := usecase.NewGinRummyInteractor(gameMock, pMock)
		result := ci.Knock(0)
		assert.Equal(t, mockOutput, result)
	})

	t.Run("error is returned", func(t *testing.T) {
		knockErr := errors.New("knock error")
		pMock := new(presenter.MockGinRummyPresenter)
		pMock.On("Output", mock.Anything, knockErr).Return(mockOutput)
		gameMock := new(interfaces.MockGinRummyGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerKnock", 0).Return(knockErr)

		ci := usecase.NewGinRummyInteractor(gameMock, pMock)
		result := ci.Knock(0)
		assert.Equal(t, mockOutput, result)
	})

	t.Run("valid knock runs CPU turns", func(t *testing.T) {
		pMock := new(presenter.MockGinRummyPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockGinRummyGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerKnock", 5).Return(nil)
		gameMock.On("GetPhase").Return(domain.GinRummyPhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewGinRummyInteractor(gameMock, pMock)
		result := ci.Knock(5)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerKnock", 5)
	})
}

func TestGinRummyInteractor_Layoff(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ended returns output", func(t *testing.T) {
		pMock := new(presenter.MockGinRummyPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockGinRummyGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewGinRummyInteractor(gameMock, pMock)
		result := ci.Layoff([]int{0})
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerLayoff")
	})

	t.Run("error is returned", func(t *testing.T) {
		layoffErr := errors.New("layoff error")
		pMock := new(presenter.MockGinRummyPresenter)
		pMock.On("Output", mock.Anything, layoffErr).Return(mockOutput)
		gameMock := new(interfaces.MockGinRummyGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerLayoff", []int{0}).Return(layoffErr)

		ci := usecase.NewGinRummyInteractor(gameMock, pMock)
		result := ci.Layoff([]int{0})
		assert.Equal(t, mockOutput, result)
	})

	t.Run("valid layoff runs CPU turns", func(t *testing.T) {
		pMock := new(presenter.MockGinRummyPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockGinRummyGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerLayoff", []int{1, 2}).Return(nil)
		gameMock.On("GetPhase").Return(domain.GinRummyPhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewGinRummyInteractor(gameMock, pMock)
		result := ci.Layoff([]int{1, 2})
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerLayoff", []int{1, 2})
	})
}

func TestGinRummyInteractor_NextRound(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ended returns output", func(t *testing.T) {
		pMock := new(presenter.MockGinRummyPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockGinRummyGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewGinRummyInteractor(gameMock, pMock)
		result := ci.NextRound()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "NextRound")
	})

	t.Run("valid next round", func(t *testing.T) {
		pMock := new(presenter.MockGinRummyPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockGinRummyGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextRound").Return()
		gameMock.On("GetPhase").Return(domain.GinRummyPhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewGinRummyInteractor(gameMock, pMock)
		result := ci.NextRound()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "NextRound")
	})
}

func TestGinRummyInteractor_GetConfig(t *testing.T) {
	t.Run("returns config from game", func(t *testing.T) {
		pMock := new(presenter.MockGinRummyPresenter)
		gameMock := new(interfaces.MockGinRummyGame)
		expected := domain.GinRummyConfig{CpuDifficulty: domain.GinRummyCpuDifficultyHard, PointLimit: 200}
		gameMock.On("GetConfig").Return(expected)

		ci := usecase.NewGinRummyInteractor(gameMock, pMock)
		result := ci.GetConfig()
		assert.Equal(t, expected, result)
	})
}

func TestGinRummyInteractor_ActionLog(t *testing.T) {
	pMock := new(presenter.MockGinRummyPresenter)
	gameMock := new(interfaces.MockGinRummyGame)
	pMock.On("ActionLogOutput", gameMock).Return(`{"entries":[]}`)

	ci := usecase.NewGinRummyInteractor(gameMock, pMock)
	result := ci.ActionLog()
	assert.Equal(t, `{"entries":[]}`, result)
	pMock.AssertExpectations(t)
}

func TestGinRummyInteractor_RunCpuTurns(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("stops when game ended", func(t *testing.T) {
		pMock := new(presenter.MockGinRummyPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockGinRummyGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewGinRummyInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("stops when phase is RoundEnd", func(t *testing.T) {
		pMock := new(presenter.MockGinRummyPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockGinRummyGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.GinRummyPhaseRoundEnd)

		ci := usecase.NewGinRummyInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("stops when phase is GameEnd", func(t *testing.T) {
		pMock := new(presenter.MockGinRummyPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockGinRummyGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.GinRummyPhaseGameEnd)

		ci := usecase.NewGinRummyInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("stops when human turn", func(t *testing.T) {
		pMock := new(presenter.MockGinRummyPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockGinRummyGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.GinRummyPhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewGinRummyInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("CPU plays then stops at human turn", func(t *testing.T) {
		pMock := new(presenter.MockGinRummyPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockGinRummyGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		// First iteration: Draw phase, not human → CpuPlay
		gameMock.On("GetPhase").Return(domain.GinRummyPhaseDraw)
		gameMock.On("IsHumanTurn").Return(false).Once()
		gameMock.On("CpuPlay").Return()
		// Second iteration: human turn → break
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewGinRummyInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertCalled(t, "CpuPlay")
	})
}
