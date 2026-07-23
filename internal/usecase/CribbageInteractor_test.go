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

func TestNewCribbageInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockCribbagePresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "CribbageInteractor: g must not be nil", func() {
			usecase.NewCribbageInteractor(nil, pMock)
		})
	})

	t.Run("panics when gp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockCribbageGame)
		assert.PanicsWithValue(t, "CribbageInteractor: gp must not be nil", func() {
			usecase.NewCribbageInteractor(gameMock, nil)
		})
	})
}

func TestCribbageInteractor_Reset(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("reset calls game Reset and returns output", func(t *testing.T) {
		pMock := new(presenter.MockCribbagePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCribbageGame)
		gameMock.On("Reset").Return()
		// runCpuTurns: not ended, phase=Discard, human turn → break
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.CribbagePhaseDiscard)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewCribbageInteractor(gameMock, pMock)
		result := ci.Reset()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "Reset")
	})
}

func TestCribbageInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("valid config sets config then resets", func(t *testing.T) {
		pMock := new(presenter.MockCribbagePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCribbageGame)
		cfg := domain.CribbageConfig{CpuDifficulty: domain.CribbageCpuDifficultyHard, PointLimit: 200}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.CribbagePhaseDiscard)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewCribbageInteractor(gameMock, pMock)
		result := ci.ResetWithConfig(cfg)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "SetConfig", cfg)
		gameMock.AssertCalled(t, "Reset")
	})

	t.Run("invalid config returns error without resetting", func(t *testing.T) {
		pMock := new(presenter.MockCribbagePresenter)
		gameMock := new(interfaces.MockCribbageGame)
		pMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")

		ci := usecase.NewCribbageInteractor(gameMock, pMock)
		cfg := domain.CribbageConfig{CpuDifficulty: domain.CribbageCpuDifficulty(-1), PointLimit: 100}
		result := ci.ResetWithConfig(cfg)
		assert.Equal(t, "validation error", result)
		gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
	})
}

func TestCribbageInteractor_Discard(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ended returns output without discarding", func(t *testing.T) {
		pMock := new(presenter.MockCribbagePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCribbageGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewCribbageInteractor(gameMock, pMock)
		result := ci.Discard([]int{0, 1})
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerDiscard")
	})

	t.Run("not human turn returns output without discarding", func(t *testing.T) {
		pMock := new(presenter.MockCribbagePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCribbageGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		ci := usecase.NewCribbageInteractor(gameMock, pMock)
		result := ci.Discard([]int{0, 1})
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerDiscard")
	})

	t.Run("discard error is returned to presenter", func(t *testing.T) {
		discardErr := errors.New("discard error")
		pMock := new(presenter.MockCribbagePresenter)
		pMock.On("Output", mock.Anything, discardErr).Return(mockOutput)
		gameMock := new(interfaces.MockCribbageGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDiscard", []int{0, 1}).Return(discardErr)

		ci := usecase.NewCribbageInteractor(gameMock, pMock)
		result := ci.Discard([]int{0, 1})
		assert.Equal(t, mockOutput, result)
	})

	t.Run("valid discard runs CPU turns", func(t *testing.T) {
		pMock := new(presenter.MockCribbagePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCribbageGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerDiscard", []int{0, 1}).Return(nil)
		// runCpuTurns: human turn → break
		gameMock.On("GetPhase").Return(domain.CribbagePhaseDiscard)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewCribbageInteractor(gameMock, pMock)
		result := ci.Discard([]int{0, 1})
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerDiscard", []int{0, 1})
	})
}

func TestCribbageInteractor_Cut(t *testing.T) {
	mockOutput := `{"phase":1}`

	t.Run("game ended returns output without cutting", func(t *testing.T) {
		pMock := new(presenter.MockCribbagePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCribbageGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewCribbageInteractor(gameMock, pMock)
		result := ci.Cut()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerCut")
	})

	t.Run("not human turn returns output without cutting", func(t *testing.T) {
		pMock := new(presenter.MockCribbagePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCribbageGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		ci := usecase.NewCribbageInteractor(gameMock, pMock)
		result := ci.Cut()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerCut")
	})

	t.Run("cut error is returned to presenter", func(t *testing.T) {
		cutErr := errors.New("cut error")
		pMock := new(presenter.MockCribbagePresenter)
		pMock.On("Output", mock.Anything, cutErr).Return(mockOutput)
		gameMock := new(interfaces.MockCribbageGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerCut").Return(cutErr)

		ci := usecase.NewCribbageInteractor(gameMock, pMock)
		result := ci.Cut()
		assert.Equal(t, mockOutput, result)
	})

	t.Run("valid cut runs CPU turns", func(t *testing.T) {
		pMock := new(presenter.MockCribbagePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCribbageGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerCut").Return(nil)
		// runCpuTurns: pegging phase, human turn → break
		gameMock.On("GetPhase").Return(domain.CribbagePhasePegging)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewCribbageInteractor(gameMock, pMock)
		result := ci.Cut()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerCut")
	})
}

func TestCribbageInteractor_Peg(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ended returns output", func(t *testing.T) {
		pMock := new(presenter.MockCribbagePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCribbageGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewCribbageInteractor(gameMock, pMock)
		result := ci.Peg(0)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerPeg")
	})

	t.Run("not human turn returns output", func(t *testing.T) {
		pMock := new(presenter.MockCribbagePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCribbageGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		ci := usecase.NewCribbageInteractor(gameMock, pMock)
		result := ci.Peg(0)
		assert.Equal(t, mockOutput, result)
	})

	t.Run("error is returned to presenter", func(t *testing.T) {
		pegErr := errors.New("peg error")
		pMock := new(presenter.MockCribbagePresenter)
		pMock.On("Output", mock.Anything, pegErr).Return(mockOutput)
		gameMock := new(interfaces.MockCribbageGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerPeg", 0).Return(pegErr)

		ci := usecase.NewCribbageInteractor(gameMock, pMock)
		result := ci.Peg(0)
		assert.Equal(t, mockOutput, result)
	})

	t.Run("valid peg runs CPU turns", func(t *testing.T) {
		pMock := new(presenter.MockCribbagePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCribbageGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerPeg", 3).Return(nil)
		gameMock.On("GetPhase").Return(domain.CribbagePhasePegging)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewCribbageInteractor(gameMock, pMock)
		result := ci.Peg(3)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerPeg", 3)
	})
}

func TestCribbageInteractor_Go(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ended returns output", func(t *testing.T) {
		pMock := new(presenter.MockCribbagePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCribbageGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewCribbageInteractor(gameMock, pMock)
		result := ci.Go()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerGo")
	})

	t.Run("not human turn returns output", func(t *testing.T) {
		pMock := new(presenter.MockCribbagePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCribbageGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		ci := usecase.NewCribbageInteractor(gameMock, pMock)
		result := ci.Go()
		assert.Equal(t, mockOutput, result)
	})

	t.Run("error is returned to presenter", func(t *testing.T) {
		goErr := errors.New("go error")
		pMock := new(presenter.MockCribbagePresenter)
		pMock.On("Output", mock.Anything, goErr).Return(mockOutput)
		gameMock := new(interfaces.MockCribbageGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerGo").Return(goErr)

		ci := usecase.NewCribbageInteractor(gameMock, pMock)
		result := ci.Go()
		assert.Equal(t, mockOutput, result)
	})

	t.Run("valid go runs CPU turns", func(t *testing.T) {
		pMock := new(presenter.MockCribbagePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCribbageGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerGo").Return(nil)
		gameMock.On("GetPhase").Return(domain.CribbagePhasePegging)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewCribbageInteractor(gameMock, pMock)
		result := ci.Go()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerGo")
	})
}

func TestCribbageInteractor_ShowNext(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ended returns output", func(t *testing.T) {
		pMock := new(presenter.MockCribbagePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCribbageGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewCribbageInteractor(gameMock, pMock)
		result := ci.ShowNext()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "ShowNext")
	})

	t.Run("show next error is returned", func(t *testing.T) {
		showErr := errors.New("show error")
		pMock := new(presenter.MockCribbagePresenter)
		pMock.On("Output", mock.Anything, showErr).Return(mockOutput)
		gameMock := new(interfaces.MockCribbageGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("ShowNext").Return(showErr)

		ci := usecase.NewCribbageInteractor(gameMock, pMock)
		result := ci.ShowNext()
		assert.Equal(t, mockOutput, result)
	})

	t.Run("valid show next returns output", func(t *testing.T) {
		pMock := new(presenter.MockCribbagePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCribbageGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("ShowNext").Return(nil)

		ci := usecase.NewCribbageInteractor(gameMock, pMock)
		result := ci.ShowNext()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "ShowNext")
	})
}

func TestCribbageInteractor_NextRound(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ended returns output", func(t *testing.T) {
		pMock := new(presenter.MockCribbagePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCribbageGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewCribbageInteractor(gameMock, pMock)
		result := ci.NextRound()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "NextRound")
	})

	t.Run("valid next round", func(t *testing.T) {
		pMock := new(presenter.MockCribbagePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCribbageGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextRound").Return()
		gameMock.On("GetPhase").Return(domain.CribbagePhaseDiscard)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewCribbageInteractor(gameMock, pMock)
		result := ci.NextRound()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "NextRound")
	})
}

func TestCribbageInteractor_GetConfig(t *testing.T) {
	t.Run("returns config from game", func(t *testing.T) {
		pMock := new(presenter.MockCribbagePresenter)
		gameMock := new(interfaces.MockCribbageGame)
		expected := domain.CribbageConfig{CpuDifficulty: domain.CribbageCpuDifficultyHard, PointLimit: 200}
		gameMock.On("GetConfig").Return(expected)

		ci := usecase.NewCribbageInteractor(gameMock, pMock)
		result := ci.GetConfig()
		assert.Equal(t, expected, result)
	})
}

func TestCribbageInteractor_ActionLog(t *testing.T) {
	pMock := new(presenter.MockCribbagePresenter)
	gameMock := new(interfaces.MockCribbageGame)
	pMock.On("ActionLogOutput", gameMock).Return(`{"entries":[]}`)

	ci := usecase.NewCribbageInteractor(gameMock, pMock)
	result := ci.ActionLog()
	assert.Equal(t, `{"entries":[]}`, result)
	pMock.AssertExpectations(t)
}

func TestCribbageInteractor_RunCpuTurns(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("stops when game ended", func(t *testing.T) {
		pMock := new(presenter.MockCribbagePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCribbageGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewCribbageInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("stops when phase is RoundEnd", func(t *testing.T) {
		pMock := new(presenter.MockCribbagePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCribbageGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.CribbagePhaseRoundEnd)

		ci := usecase.NewCribbageInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("stops when phase is GameEnd", func(t *testing.T) {
		pMock := new(presenter.MockCribbagePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCribbageGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.CribbagePhaseGameEnd)

		ci := usecase.NewCribbageInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("stops when phase is Show", func(t *testing.T) {
		pMock := new(presenter.MockCribbagePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCribbageGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.CribbagePhaseShow)

		ci := usecase.NewCribbageInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("stops when human turn", func(t *testing.T) {
		pMock := new(presenter.MockCribbagePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCribbageGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.CribbagePhaseDiscard)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewCribbageInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("CPU plays then stops at human turn", func(t *testing.T) {
		pMock := new(presenter.MockCribbagePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCribbageGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		// First iteration: Discard phase, not human → CpuPlay
		gameMock.On("GetPhase").Return(domain.CribbagePhaseDiscard)
		gameMock.On("IsHumanTurn").Return(false).Once()
		gameMock.On("CpuPlay").Return()
		// Second iteration: human turn → break
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewCribbageInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertCalled(t, "CpuPlay")
	})
}

func TestCribbageInteractor_Hint(t *testing.T) {
	gameMock := new(interfaces.MockCribbageGame)
	pMock := new(presenter.MockCribbagePresenter)
	pMock.On("HintOutput", gameMock).Return("hint output")
	ci := usecase.NewCribbageInteractor(gameMock, pMock)
	assert.Equal(t, "hint output", ci.Hint())
	pMock.AssertCalled(t, "HintOutput", gameMock)
}
