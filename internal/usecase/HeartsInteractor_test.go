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

func TestNewHeartsInteractor_NilGuards(t *testing.T) {
	hpMock := new(presenter.MockHeartsPresenter)

	t.Run("panics when h is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "HeartsInteractor: h must not be nil", func() {
			usecase.NewHeartsInteractor(nil, hpMock)
		})
	})

	t.Run("panics when hp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockHeartsGame)
		assert.PanicsWithValue(t, "HeartsInteractor: hp must not be nil", func() {
			usecase.NewHeartsInteractor(gameMock, nil)
		})
	})
}

func TestHeartsInteractor_Reset(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("reset with pass direction (not PassNone) skips runCpuTurns", func(t *testing.T) {
		hpMock := new(presenter.MockHeartsPresenter)
		hpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockHeartsGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetPassDirection").Return(domain.HeartsPassLeft)

		hi := usecase.NewHeartsInteractor(gameMock, hpMock)
		result := hi.Reset()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "Reset")
		gameMock.AssertNotCalled(t, "GetGameEndFlag")
	})

	t.Run("reset with PassNone runs CPU turns until human turn", func(t *testing.T) {
		hpMock := new(presenter.MockHeartsPresenter)
		hpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockHeartsGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetPassDirection").Return(domain.HeartsPassNone)
		// runCpuTurns: not ended, phase=Play, not human → CpuPlay
		// then phase stays Play, not ended, human → break
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.HeartsPhasePlay)
		gameMock.On("IsHumanTurn").Return(false).Once()
		gameMock.On("CpuPlay").Return()
		gameMock.On("IsHumanTurn").Return(true)

		hi := usecase.NewHeartsInteractor(gameMock, hpMock)
		result := hi.Reset()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "CpuPlay")
	})
}

func TestHeartsInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("sets config then resets", func(t *testing.T) {
		hpMock := new(presenter.MockHeartsPresenter)
		hpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockHeartsGame)
		cfg := domain.HeartsConfig{CpuDifficulty: domain.HeartsCpuDifficultyHard, PointLimit: 50}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetPassDirection").Return(domain.HeartsPassLeft)

		hi := usecase.NewHeartsInteractor(gameMock, hpMock)
		result := hi.ResetWithConfig(cfg)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "SetConfig", cfg)
		gameMock.AssertCalled(t, "Reset")
	})
}

func TestHeartsInteractor_ResetWithConfig_ValidationError(t *testing.T) {
	hpMock := new(presenter.MockHeartsPresenter)
	gameMock := new(interfaces.MockHeartsGame)
	hpMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")

	hi := usecase.NewHeartsInteractor(gameMock, hpMock)
	cfg := domain.HeartsConfig{CpuDifficulty: domain.HeartsCpuDifficulty(-1), PointLimit: 100}
	result := hi.ResetWithConfig(cfg)
	assert.Equal(t, "validation error", result)
	gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
}

func TestHeartsInteractor_Pass(t *testing.T) {
	mockOutput := `{"phase":1}`

	t.Run("game ended returns output without passing", func(t *testing.T) {
		hpMock := new(presenter.MockHeartsPresenter)
		hpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockHeartsGame)
		gameMock.On("GetGameEndFlag").Return(true)

		hi := usecase.NewHeartsInteractor(gameMock, hpMock)
		result := hi.Pass([]int{0, 1, 2})
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerPass")
	})

	t.Run("pass error is returned to presenter", func(t *testing.T) {
		passErr := errors.New("invalid pass")
		hpMock := new(presenter.MockHeartsPresenter)
		hpMock.On("Output", mock.Anything, passErr).Return(mockOutput)
		gameMock := new(interfaces.MockHeartsGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerPass", []int{0, 1, 2}).Return(passErr)

		hi := usecase.NewHeartsInteractor(gameMock, hpMock)
		result := hi.Pass([]int{0, 1, 2})
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "CpuPass")
	})

	t.Run("valid pass executes full pass flow and runs CPU turns", func(t *testing.T) {
		hpMock := new(presenter.MockHeartsPresenter)
		hpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockHeartsGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerPass", []int{0, 1, 2}).Return(nil)
		gameMock.On("CpuPass").Return()
		gameMock.On("ExecutePass").Return()
		// runCpuTurns: human turn → break immediately
		gameMock.On("GetPhase").Return(domain.HeartsPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		hi := usecase.NewHeartsInteractor(gameMock, hpMock)
		result := hi.Pass([]int{0, 1, 2})
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "CpuPass")
		gameMock.AssertCalled(t, "ExecutePass")
	})
}

func TestHeartsInteractor_Play(t *testing.T) {
	mockOutput := `{"phase":1}`

	t.Run("game ended returns output without playing", func(t *testing.T) {
		hpMock := new(presenter.MockHeartsPresenter)
		hpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockHeartsGame)
		gameMock.On("GetGameEndFlag").Return(true)

		hi := usecase.NewHeartsInteractor(gameMock, hpMock)
		result := hi.Play(0)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerPlay")
	})

	t.Run("not human turn returns output without playing", func(t *testing.T) {
		hpMock := new(presenter.MockHeartsPresenter)
		hpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockHeartsGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		hi := usecase.NewHeartsInteractor(gameMock, hpMock)
		result := hi.Play(0)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerPlay")
	})

	t.Run("play error is returned to presenter", func(t *testing.T) {
		playErr := errors.New("invalid play")
		hpMock := new(presenter.MockHeartsPresenter)
		hpMock.On("Output", mock.Anything, playErr).Return(mockOutput)
		gameMock := new(interfaces.MockHeartsGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerPlay", 0).Return(playErr)

		hi := usecase.NewHeartsInteractor(gameMock, hpMock)
		result := hi.Play(0)
		assert.Equal(t, mockOutput, result)
	})

	t.Run("valid play runs CPU turns", func(t *testing.T) {
		hpMock := new(presenter.MockHeartsPresenter)
		hpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockHeartsGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerPlay", 2).Return(nil)
		// runCpuTurns: human turn → break
		gameMock.On("GetPhase").Return(domain.HeartsPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		hi := usecase.NewHeartsInteractor(gameMock, hpMock)
		result := hi.Play(2)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerPlay", 2)
	})
}

func TestHeartsInteractor_NextTrick(t *testing.T) {
	mockOutput := `{"phase":1}`

	t.Run("calls NextTrick and runs CPU turns", func(t *testing.T) {
		hpMock := new(presenter.MockHeartsPresenter)
		hpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockHeartsGame)
		gameMock.On("NextTrick").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.HeartsPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		hi := usecase.NewHeartsInteractor(gameMock, hpMock)
		result := hi.NextTrick()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "NextTrick")
	})
}

func TestHeartsInteractor_NextRound(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ends after scoring", func(t *testing.T) {
		hpMock := new(presenter.MockHeartsPresenter)
		hpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockHeartsGame)
		gameMock.On("ScoreRound").Return()
		gameMock.On("GetGameEndFlag").Return(true)

		hi := usecase.NewHeartsInteractor(gameMock, hpMock)
		result := hi.NextRound()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "ScoreRound")
		gameMock.AssertNotCalled(t, "NextRound")
	})

	t.Run("game continues to next round", func(t *testing.T) {
		hpMock := new(presenter.MockHeartsPresenter)
		hpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockHeartsGame)
		gameMock.On("ScoreRound").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextRound").Return()

		hi := usecase.NewHeartsInteractor(gameMock, hpMock)
		result := hi.NextRound()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "ScoreRound")
		gameMock.AssertCalled(t, "NextRound")
	})
}

func TestHeartsInteractor_GetConfig(t *testing.T) {
	t.Run("returns config from game", func(t *testing.T) {
		hpMock := new(presenter.MockHeartsPresenter)
		gameMock := new(interfaces.MockHeartsGame)
		expected := domain.HeartsConfig{CpuDifficulty: domain.HeartsCpuDifficultyHard, PointLimit: 50}
		gameMock.On("GetConfig").Return(expected)

		hi := usecase.NewHeartsInteractor(gameMock, hpMock)
		result := hi.GetConfig()
		assert.Equal(t, expected, result)
	})
}

func TestHeartsInteractor_Hint(t *testing.T) {
	hpMock := new(presenter.MockHeartsPresenter)
	gameMock := new(interfaces.MockHeartsGame)
	hpMock.On("HintOutput", gameMock).Return(`{"hint":{"cardIndices":[0]}}`)

	hi := usecase.NewHeartsInteractor(gameMock, hpMock)
	result := hi.Hint()
	assert.Equal(t, `{"hint":{"cardIndices":[0]}}`, result)
	hpMock.AssertExpectations(t)
}

func TestHeartsInteractor_ActionLog(t *testing.T) {
	hpMock := new(presenter.MockHeartsPresenter)
	gameMock := new(interfaces.MockHeartsGame)
	hpMock.On("ActionLogOutput", gameMock).Return(`{"entries":[]}`)

	hi := usecase.NewHeartsInteractor(gameMock, hpMock)
	result := hi.ActionLog()
	assert.Equal(t, `{"entries":[]}`, result)
	hpMock.AssertExpectations(t)
}

func TestHeartsInteractor_RunCpuTurns(t *testing.T) {
	mockOutput := `{"phase":1}`

	t.Run("stops when game ended", func(t *testing.T) {
		hpMock := new(presenter.MockHeartsPresenter)
		hpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockHeartsGame)
		gameMock.On("NextTrick").Return()
		gameMock.On("GetGameEndFlag").Return(true)

		hi := usecase.NewHeartsInteractor(gameMock, hpMock)
		hi.NextTrick()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("stops when phase is TrickEnd", func(t *testing.T) {
		hpMock := new(presenter.MockHeartsPresenter)
		hpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockHeartsGame)
		gameMock.On("NextTrick").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.HeartsPhaseTrickEnd)

		hi := usecase.NewHeartsInteractor(gameMock, hpMock)
		hi.NextTrick()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("stops when phase is RoundEnd", func(t *testing.T) {
		hpMock := new(presenter.MockHeartsPresenter)
		hpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockHeartsGame)
		gameMock.On("NextTrick").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.HeartsPhaseRoundEnd)

		hi := usecase.NewHeartsInteractor(gameMock, hpMock)
		hi.NextTrick()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("stops when phase is GameEnd", func(t *testing.T) {
		hpMock := new(presenter.MockHeartsPresenter)
		hpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockHeartsGame)
		gameMock.On("NextTrick").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.HeartsPhaseGameEnd)

		hi := usecase.NewHeartsInteractor(gameMock, hpMock)
		hi.NextTrick()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("stops when phase is not Play (e.g. Pass)", func(t *testing.T) {
		hpMock := new(presenter.MockHeartsPresenter)
		hpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockHeartsGame)
		gameMock.On("NextTrick").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.HeartsPhasePass)

		hi := usecase.NewHeartsInteractor(gameMock, hpMock)
		hi.NextTrick()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("CPU plays then trick ends with ResolveTrick leading to RoundEnd", func(t *testing.T) {
		hpMock := new(presenter.MockHeartsPresenter)
		hpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockHeartsGame)
		gameMock.On("NextTrick").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		// First iteration: Play phase, not human → CpuPlay
		gameMock.On("GetPhase").Return(domain.HeartsPhasePlay).Once()
		gameMock.On("IsHumanTurn").Return(false).Once()
		gameMock.On("CpuPlay").Return().Once()
		// After CpuPlay: phase becomes TrickEnd → ResolveTrick
		gameMock.On("GetPhase").Return(domain.HeartsPhaseTrickEnd).Once()
		gameMock.On("ResolveTrick").Return()
		// After ResolveTrick: phase becomes RoundEnd → break
		gameMock.On("GetPhase").Return(domain.HeartsPhaseRoundEnd)

		hi := usecase.NewHeartsInteractor(gameMock, hpMock)
		hi.NextTrick()
		gameMock.AssertCalled(t, "CpuPlay")
		gameMock.AssertCalled(t, "ResolveTrick")
		// NextTrick is called once (from the interactor method), but not from runCpuTurns
		gameMock.AssertNumberOfCalls(t, "NextTrick", 1)
	})

	t.Run("CPU plays then trick ends with ResolveTrick then continues", func(t *testing.T) {
		hpMock := new(presenter.MockHeartsPresenter)
		hpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockHeartsGame)
		gameMock.On("NextTrick").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		// First iteration: Play phase, not human → CpuPlay
		gameMock.On("GetPhase").Return(domain.HeartsPhasePlay).Once()
		gameMock.On("IsHumanTurn").Return(false).Once()
		gameMock.On("CpuPlay").Return()
		// After CpuPlay: phase becomes TrickEnd → ResolveTrick
		gameMock.On("GetPhase").Return(domain.HeartsPhaseTrickEnd).Once()
		gameMock.On("ResolveTrick").Return()
		// After ResolveTrick: phase is Play → NextTrick inside runCpuTurns
		gameMock.On("GetPhase").Return(domain.HeartsPhasePlay).Once()
		// NextTrick called inside runCpuTurns, then loop continues
		// Next iteration: phase=Play, human turn → break
		gameMock.On("GetPhase").Return(domain.HeartsPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		hi := usecase.NewHeartsInteractor(gameMock, hpMock)
		hi.NextTrick()
		gameMock.AssertCalled(t, "CpuPlay")
		gameMock.AssertCalled(t, "ResolveTrick")
	})

	t.Run("CPU play does not trigger trick end", func(t *testing.T) {
		hpMock := new(presenter.MockHeartsPresenter)
		hpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockHeartsGame)
		gameMock.On("NextTrick").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		// First iteration: Play phase, not human → CpuPlay
		gameMock.On("GetPhase").Return(domain.HeartsPhasePlay).Once()
		gameMock.On("IsHumanTurn").Return(false).Once()
		gameMock.On("CpuPlay").Return()
		// After CpuPlay: phase stays Play (not TrickEnd)
		gameMock.On("GetPhase").Return(domain.HeartsPhasePlay)
		// Next iteration: human turn → break
		gameMock.On("IsHumanTurn").Return(true)

		hi := usecase.NewHeartsInteractor(gameMock, hpMock)
		hi.NextTrick()
		gameMock.AssertCalled(t, "CpuPlay")
		gameMock.AssertNotCalled(t, "ResolveTrick")
	})
}
