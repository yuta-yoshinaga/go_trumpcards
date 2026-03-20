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

func TestNewCrazyEightsInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockCrazyEightsPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "CrazyEightsInteractor: g must not be nil", func() {
			usecase.NewCrazyEightsInteractor(nil, pMock)
		})
	})

	t.Run("panics when gp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockCrazyEightsGame)
		assert.PanicsWithValue(t, "CrazyEightsInteractor: gp must not be nil", func() {
			usecase.NewCrazyEightsInteractor(gameMock, nil)
		})
	})
}

func TestCrazyEightsInteractor_Reset(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("reset calls game Reset and returns output", func(t *testing.T) {
		pMock := new(presenter.MockCrazyEightsPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCrazyEightsGame)
		gameMock.On("Reset").Return()
		// runCpuTurns: not ended, phase=Play, human turn → break
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.CrazyEightsPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewCrazyEightsInteractor(gameMock, pMock)
		result := ci.Reset()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "Reset")
	})
}

func TestCrazyEightsInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("valid config sets config then resets", func(t *testing.T) {
		pMock := new(presenter.MockCrazyEightsPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCrazyEightsGame)
		cfg := domain.CrazyEightsConfig{CpuDifficulty: domain.CrazyEightsCpuDifficultyHard, PointLimit: 300}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.CrazyEightsPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewCrazyEightsInteractor(gameMock, pMock)
		result := ci.ResetWithConfig(cfg)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "SetConfig", cfg)
		gameMock.AssertCalled(t, "Reset")
	})

	t.Run("invalid config returns error without resetting", func(t *testing.T) {
		pMock := new(presenter.MockCrazyEightsPresenter)
		gameMock := new(interfaces.MockCrazyEightsGame)
		pMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")

		ci := usecase.NewCrazyEightsInteractor(gameMock, pMock)
		cfg := domain.CrazyEightsConfig{CpuDifficulty: domain.CrazyEightsCpuDifficulty(-1), PointLimit: 200}
		result := ci.ResetWithConfig(cfg)
		assert.Equal(t, "validation error", result)
		gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
	})
}

func TestCrazyEightsInteractor_Play(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ended returns output without playing", func(t *testing.T) {
		pMock := new(presenter.MockCrazyEightsPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCrazyEightsGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewCrazyEightsInteractor(gameMock, pMock)
		result := ci.Play(0)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerPlay")
	})

	t.Run("not human turn returns output without playing", func(t *testing.T) {
		pMock := new(presenter.MockCrazyEightsPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCrazyEightsGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		ci := usecase.NewCrazyEightsInteractor(gameMock, pMock)
		result := ci.Play(0)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerPlay")
	})

	t.Run("play error is returned to presenter", func(t *testing.T) {
		playErr := errors.New("invalid play")
		pMock := new(presenter.MockCrazyEightsPresenter)
		pMock.On("Output", mock.Anything, playErr).Return(mockOutput)
		gameMock := new(interfaces.MockCrazyEightsGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerPlay", 0).Return(playErr)

		ci := usecase.NewCrazyEightsInteractor(gameMock, pMock)
		result := ci.Play(0)
		assert.Equal(t, mockOutput, result)
	})

	t.Run("valid play runs CPU turns", func(t *testing.T) {
		pMock := new(presenter.MockCrazyEightsPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCrazyEightsGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerPlay", 2).Return(nil)
		// runCpuTurns: human turn → break
		gameMock.On("GetPhase").Return(domain.CrazyEightsPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewCrazyEightsInteractor(gameMock, pMock)
		result := ci.Play(2)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerPlay", 2)
	})
}

func TestCrazyEightsInteractor_ChooseSuit(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ended returns output without choosing", func(t *testing.T) {
		pMock := new(presenter.MockCrazyEightsPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCrazyEightsGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewCrazyEightsInteractor(gameMock, pMock)
		result := ci.ChooseSuit(1)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerChooseSuit")
	})

	t.Run("error is returned to presenter", func(t *testing.T) {
		suitErr := errors.New("invalid suit")
		pMock := new(presenter.MockCrazyEightsPresenter)
		pMock.On("Output", mock.Anything, suitErr).Return(mockOutput)
		gameMock := new(interfaces.MockCrazyEightsGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerChooseSuit", 5).Return(suitErr)

		ci := usecase.NewCrazyEightsInteractor(gameMock, pMock)
		result := ci.ChooseSuit(5)
		assert.Equal(t, mockOutput, result)
	})

	t.Run("valid suit runs CPU turns", func(t *testing.T) {
		pMock := new(presenter.MockCrazyEightsPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCrazyEightsGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerChooseSuit", 1).Return(nil)
		// runCpuTurns: human turn → break
		gameMock.On("GetPhase").Return(domain.CrazyEightsPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewCrazyEightsInteractor(gameMock, pMock)
		result := ci.ChooseSuit(1)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerChooseSuit", 1)
	})
}

func TestCrazyEightsInteractor_Draw(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ended returns output without drawing", func(t *testing.T) {
		pMock := new(presenter.MockCrazyEightsPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCrazyEightsGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewCrazyEightsInteractor(gameMock, pMock)
		result := ci.Draw()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerDraw")
	})

	t.Run("not human turn returns output without drawing", func(t *testing.T) {
		pMock := new(presenter.MockCrazyEightsPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCrazyEightsGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		ci := usecase.NewCrazyEightsInteractor(gameMock, pMock)
		result := ci.Draw()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerDraw")
	})

	t.Run("draw error is returned to presenter", func(t *testing.T) {
		drawErr := errors.New("draw error")
		pMock := new(presenter.MockCrazyEightsPresenter)
		pMock.On("Output", mock.Anything, drawErr).Return(mockOutput)
		gameMock := new(interfaces.MockCrazyEightsGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDraw").Return(drawErr)

		ci := usecase.NewCrazyEightsInteractor(gameMock, pMock)
		result := ci.Draw()
		assert.Equal(t, mockOutput, result)
	})

	t.Run("valid draw runs CPU turns", func(t *testing.T) {
		pMock := new(presenter.MockCrazyEightsPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCrazyEightsGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerDraw").Return(nil)
		// runCpuTurns: human turn → break
		gameMock.On("GetPhase").Return(domain.CrazyEightsPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewCrazyEightsInteractor(gameMock, pMock)
		result := ci.Draw()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerDraw")
	})
}

func TestCrazyEightsInteractor_NextRound(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ends after scoring", func(t *testing.T) {
		pMock := new(presenter.MockCrazyEightsPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCrazyEightsGame)
		gameMock.On("ScoreRound").Return()
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewCrazyEightsInteractor(gameMock, pMock)
		result := ci.NextRound()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "ScoreRound")
		gameMock.AssertNotCalled(t, "NextRound")
	})

	t.Run("game continues to next round", func(t *testing.T) {
		pMock := new(presenter.MockCrazyEightsPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCrazyEightsGame)
		gameMock.On("ScoreRound").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextRound").Return()
		// runCpuTurns: human turn → break
		gameMock.On("GetPhase").Return(domain.CrazyEightsPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewCrazyEightsInteractor(gameMock, pMock)
		result := ci.NextRound()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "ScoreRound")
		gameMock.AssertCalled(t, "NextRound")
	})
}

func TestCrazyEightsInteractor_GetConfig(t *testing.T) {
	t.Run("returns config from game", func(t *testing.T) {
		pMock := new(presenter.MockCrazyEightsPresenter)
		gameMock := new(interfaces.MockCrazyEightsGame)
		expected := domain.CrazyEightsConfig{CpuDifficulty: domain.CrazyEightsCpuDifficultyHard, PointLimit: 300}
		gameMock.On("GetConfig").Return(expected)

		ci := usecase.NewCrazyEightsInteractor(gameMock, pMock)
		result := ci.GetConfig()
		assert.Equal(t, expected, result)
	})
}

func TestCrazyEightsInteractor_ActionLog(t *testing.T) {
	pMock := new(presenter.MockCrazyEightsPresenter)
	gameMock := new(interfaces.MockCrazyEightsGame)
	pMock.On("ActionLogOutput", gameMock).Return(`{"entries":[]}`)

	ci := usecase.NewCrazyEightsInteractor(gameMock, pMock)
	result := ci.ActionLog()
	assert.Equal(t, `{"entries":[]}`, result)
	pMock.AssertExpectations(t)
}

func TestCrazyEightsInteractor_RunCpuTurns(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("stops when game ended", func(t *testing.T) {
		pMock := new(presenter.MockCrazyEightsPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCrazyEightsGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewCrazyEightsInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("stops when phase is RoundEnd", func(t *testing.T) {
		pMock := new(presenter.MockCrazyEightsPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCrazyEightsGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.CrazyEightsPhaseRoundEnd)

		ci := usecase.NewCrazyEightsInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("stops when phase is GameEnd", func(t *testing.T) {
		pMock := new(presenter.MockCrazyEightsPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCrazyEightsGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.CrazyEightsPhaseGameEnd)

		ci := usecase.NewCrazyEightsInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("ChooseSuit phase with human turn breaks", func(t *testing.T) {
		pMock := new(presenter.MockCrazyEightsPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCrazyEightsGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.CrazyEightsPhaseChooseSuit)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewCrazyEightsInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertNotCalled(t, "CpuChooseSuit")
	})

	t.Run("ChooseSuit phase with CPU turn calls CpuChooseSuit and continues", func(t *testing.T) {
		pMock := new(presenter.MockCrazyEightsPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCrazyEightsGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		// First iteration: ChooseSuit, CPU turn → CpuChooseSuit
		gameMock.On("GetPhase").Return(domain.CrazyEightsPhaseChooseSuit).Once()
		gameMock.On("IsHumanTurn").Return(false).Once()
		gameMock.On("CpuChooseSuit").Return()
		// Second iteration: Play phase, human turn → break
		gameMock.On("GetPhase").Return(domain.CrazyEightsPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewCrazyEightsInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertCalled(t, "CpuChooseSuit")
	})

	t.Run("Play phase with non-Play phase (unknown) breaks", func(t *testing.T) {
		pMock := new(presenter.MockCrazyEightsPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCrazyEightsGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		// Phase is something other than Play/ChooseSuit/RoundEnd/GameEnd
		gameMock.On("GetPhase").Return(domain.CrazyEightsPhase(99))

		ci := usecase.NewCrazyEightsInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("Play phase with human turn breaks", func(t *testing.T) {
		pMock := new(presenter.MockCrazyEightsPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCrazyEightsGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.CrazyEightsPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewCrazyEightsInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("Play phase with CPU turn calls CpuPlay then stops at human turn", func(t *testing.T) {
		pMock := new(presenter.MockCrazyEightsPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCrazyEightsGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		// First iteration: Play phase, not human → CpuPlay
		gameMock.On("GetPhase").Return(domain.CrazyEightsPhasePlay)
		gameMock.On("IsHumanTurn").Return(false).Once()
		gameMock.On("CpuPlay").Return()
		// Second iteration: human turn → break
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewCrazyEightsInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertCalled(t, "CpuPlay")
	})
}
