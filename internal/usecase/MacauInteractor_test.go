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

func TestNewMacauInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockMacauPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "MacauInteractor: g must not be nil", func() {
			usecase.NewMacauInteractor(nil, pMock)
		})
	})

	t.Run("panics when gp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockMacauGame)
		assert.PanicsWithValue(t, "MacauInteractor: gp must not be nil", func() {
			usecase.NewMacauInteractor(gameMock, nil)
		})
	})
}

func TestMacauInteractor_Reset(t *testing.T) {
	mockOutput := `{"phase":0}`
	pMock := new(presenter.MockMacauPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockMacauGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.MacauPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewMacauInteractor(gameMock, pMock)
	assert.Equal(t, mockOutput, ci.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestMacauInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("valid config sets then resets", func(t *testing.T) {
		pMock := new(presenter.MockMacauPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockMacauGame)
		cfg := domain.MacauConfig{CpuDifficulty: domain.MacauCpuDifficultyHard, PointLimit: 300}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.MacauPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewMacauInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.ResetWithConfig(cfg))
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config returns error without resetting", func(t *testing.T) {
		pMock := new(presenter.MockMacauPresenter)
		gameMock := new(interfaces.MockMacauGame)
		pMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")

		ci := usecase.NewMacauInteractor(gameMock, pMock)
		cfg := domain.MacauConfig{CpuDifficulty: domain.MacauCpuDifficulty(-1), PointLimit: 200}
		assert.Equal(t, "validation error", ci.ResetWithConfig(cfg))
		gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
	})
}

func TestMacauInteractor_Play(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ended skips playing", func(t *testing.T) {
		pMock := new(presenter.MockMacauPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockMacauGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewMacauInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Play(0))
		gameMock.AssertNotCalled(t, "PlayerPlay")
	})

	t.Run("not human turn skips playing", func(t *testing.T) {
		pMock := new(presenter.MockMacauPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockMacauGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		ci := usecase.NewMacauInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Play(0))
		gameMock.AssertNotCalled(t, "PlayerPlay")
	})

	t.Run("play error returned to presenter", func(t *testing.T) {
		playErr := errors.New("invalid play")
		pMock := new(presenter.MockMacauPresenter)
		pMock.On("Output", mock.Anything, playErr).Return(mockOutput)
		gameMock := new(interfaces.MockMacauGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerPlay", 0).Return(playErr)

		ci := usecase.NewMacauInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Play(0))
	})

	t.Run("valid play runs CPU turns", func(t *testing.T) {
		pMock := new(presenter.MockMacauPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockMacauGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerPlay", 2).Return(nil)
		gameMock.On("GetPhase").Return(domain.MacauPhasePlay)

		ci := usecase.NewMacauInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Play(2))
		gameMock.AssertCalled(t, "PlayerPlay", 2)
	})
}

func TestMacauInteractor_ChooseSuit(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ended skips", func(t *testing.T) {
		pMock := new(presenter.MockMacauPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockMacauGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewMacauInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.ChooseSuit(1))
		gameMock.AssertNotCalled(t, "PlayerChooseSuit")
	})

	t.Run("error returned to presenter", func(t *testing.T) {
		suitErr := errors.New("invalid suit")
		pMock := new(presenter.MockMacauPresenter)
		pMock.On("Output", mock.Anything, suitErr).Return(mockOutput)
		gameMock := new(interfaces.MockMacauGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerChooseSuit", 5).Return(suitErr)

		ci := usecase.NewMacauInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.ChooseSuit(5))
	})

	t.Run("valid suit runs CPU turns", func(t *testing.T) {
		pMock := new(presenter.MockMacauPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockMacauGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerChooseSuit", 1).Return(nil)
		gameMock.On("GetPhase").Return(domain.MacauPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewMacauInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.ChooseSuit(1))
		gameMock.AssertCalled(t, "PlayerChooseSuit", 1)
	})
}

func TestMacauInteractor_Draw(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("not human turn skips", func(t *testing.T) {
		pMock := new(presenter.MockMacauPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockMacauGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		ci := usecase.NewMacauInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Draw())
		gameMock.AssertNotCalled(t, "PlayerDraw")
	})

	t.Run("draw error returned to presenter", func(t *testing.T) {
		drawErr := errors.New("draw error")
		pMock := new(presenter.MockMacauPresenter)
		pMock.On("Output", mock.Anything, drawErr).Return(mockOutput)
		gameMock := new(interfaces.MockMacauGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDraw").Return(drawErr)

		ci := usecase.NewMacauInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Draw())
	})

	t.Run("valid draw runs CPU turns", func(t *testing.T) {
		pMock := new(presenter.MockMacauPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockMacauGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDraw").Return(nil)
		gameMock.On("GetPhase").Return(domain.MacauPhasePlay)

		ci := usecase.NewMacauInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Draw())
		gameMock.AssertCalled(t, "PlayerDraw")
	})
}

func TestMacauInteractor_Declare(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ended skips", func(t *testing.T) {
		pMock := new(presenter.MockMacauPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockMacauGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewMacauInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Declare())
		gameMock.AssertNotCalled(t, "PlayerDeclare")
	})

	t.Run("error returned to presenter", func(t *testing.T) {
		declErr := errors.New("wrong phase")
		pMock := new(presenter.MockMacauPresenter)
		pMock.On("Output", mock.Anything, declErr).Return(mockOutput)
		gameMock := new(interfaces.MockMacauGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerDeclare").Return(declErr)

		ci := usecase.NewMacauInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Declare())
	})

	t.Run("valid declare runs CPU turns", func(t *testing.T) {
		pMock := new(presenter.MockMacauPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockMacauGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerDeclare").Return(nil)
		gameMock.On("GetPhase").Return(domain.MacauPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewMacauInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Declare())
		gameMock.AssertCalled(t, "PlayerDeclare")
	})
}

func TestMacauInteractor_SkipDeclare(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ended skips", func(t *testing.T) {
		pMock := new(presenter.MockMacauPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockMacauGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewMacauInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.SkipDeclare())
		gameMock.AssertNotCalled(t, "PlayerSkipDeclare")
	})

	t.Run("error returned to presenter", func(t *testing.T) {
		skErr := errors.New("wrong phase")
		pMock := new(presenter.MockMacauPresenter)
		pMock.On("Output", mock.Anything, skErr).Return(mockOutput)
		gameMock := new(interfaces.MockMacauGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerSkipDeclare").Return(skErr)

		ci := usecase.NewMacauInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.SkipDeclare())
	})

	t.Run("valid skip runs CPU turns", func(t *testing.T) {
		pMock := new(presenter.MockMacauPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockMacauGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerSkipDeclare").Return(nil)
		gameMock.On("GetPhase").Return(domain.MacauPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewMacauInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.SkipDeclare())
		gameMock.AssertCalled(t, "PlayerSkipDeclare")
	})
}

func TestMacauInteractor_NextRound(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ends after scoring", func(t *testing.T) {
		pMock := new(presenter.MockMacauPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockMacauGame)
		gameMock.On("ScoreRound").Return()
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewMacauInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.NextRound())
		gameMock.AssertCalled(t, "ScoreRound")
		gameMock.AssertNotCalled(t, "NextRound")
	})

	t.Run("game continues to next round", func(t *testing.T) {
		pMock := new(presenter.MockMacauPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockMacauGame)
		gameMock.On("ScoreRound").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextRound").Return()
		gameMock.On("GetPhase").Return(domain.MacauPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewMacauInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.NextRound())
		gameMock.AssertCalled(t, "NextRound")
	})
}

func TestMacauInteractor_GetConfig(t *testing.T) {
	pMock := new(presenter.MockMacauPresenter)
	gameMock := new(interfaces.MockMacauGame)
	expected := domain.MacauConfig{CpuDifficulty: domain.MacauCpuDifficultyHard, PointLimit: 300}
	gameMock.On("GetConfig").Return(expected)

	ci := usecase.NewMacauInteractor(gameMock, pMock)
	assert.Equal(t, expected, ci.GetConfig())
}

func TestMacauInteractor_ActionLog(t *testing.T) {
	pMock := new(presenter.MockMacauPresenter)
	gameMock := new(interfaces.MockMacauGame)
	pMock.On("ActionLogOutput", gameMock).Return(`{"entries":[]}`)

	ci := usecase.NewMacauInteractor(gameMock, pMock)
	assert.Equal(t, `{"entries":[]}`, ci.ActionLog())
	pMock.AssertExpectations(t)
}

func TestMacauInteractor_Hint(t *testing.T) {
	pMock := new(presenter.MockMacauPresenter)
	gameMock := new(interfaces.MockMacauGame)
	pMock.On("HintOutput", gameMock).Return("hint")

	ci := usecase.NewMacauInteractor(gameMock, pMock)
	assert.Equal(t, "hint", ci.Hint())
	pMock.AssertExpectations(t)
}

func TestMacauInteractor_IsHumanTurnHelpers(t *testing.T) {
	t.Run("IsHumanChooseSuitTurn", func(t *testing.T) {
		pMock := new(presenter.MockMacauPresenter)
		gameMock := new(interfaces.MockMacauGame)
		gameMock.On("GetPhase").Return(domain.MacauPhaseChooseSuit)
		gameMock.On("IsHumanTurn").Return(true)
		ci := usecase.NewMacauInteractor(gameMock, pMock)
		assert.True(t, ci.IsHumanChooseSuitTurn())
	})

	t.Run("IsHumanDeclareTurn", func(t *testing.T) {
		pMock := new(presenter.MockMacauPresenter)
		gameMock := new(interfaces.MockMacauGame)
		gameMock.On("GetPhase").Return(domain.MacauPhaseMustDeclare)
		gameMock.On("IsHumanTurn").Return(true)
		ci := usecase.NewMacauInteractor(gameMock, pMock)
		assert.True(t, ci.IsHumanDeclareTurn())
	})

	t.Run("not declare turn in play phase", func(t *testing.T) {
		pMock := new(presenter.MockMacauPresenter)
		gameMock := new(interfaces.MockMacauGame)
		gameMock.On("GetPhase").Return(domain.MacauPhasePlay)
		ci := usecase.NewMacauInteractor(gameMock, pMock)
		assert.False(t, ci.IsHumanDeclareTurn())
	})
}

func TestMacauInteractor_RunCpuTurns(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("stops when game ended", func(t *testing.T) {
		pMock := new(presenter.MockMacauPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockMacauGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewMacauInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("stops in RoundEnd phase", func(t *testing.T) {
		pMock := new(presenter.MockMacauPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockMacauGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.MacauPhaseRoundEnd)

		ci := usecase.NewMacauInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("ChooseSuit phase with CPU turn calls CpuChooseSuit then stops at human", func(t *testing.T) {
		pMock := new(presenter.MockMacauPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockMacauGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.MacauPhaseChooseSuit).Once()
		gameMock.On("IsHumanTurn").Return(false).Once()
		gameMock.On("CpuChooseSuit").Return()
		gameMock.On("GetPhase").Return(domain.MacauPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewMacauInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertCalled(t, "CpuChooseSuit")
	})

	t.Run("MustDeclare phase with CPU turn calls CpuDeclare then stops at human", func(t *testing.T) {
		pMock := new(presenter.MockMacauPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockMacauGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.MacauPhaseMustDeclare).Once()
		gameMock.On("IsHumanTurn").Return(false).Once()
		gameMock.On("CpuDeclare").Return()
		gameMock.On("GetPhase").Return(domain.MacauPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewMacauInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertCalled(t, "CpuDeclare")
	})

	t.Run("Play phase with CPU turn calls CpuPlay then stops at human", func(t *testing.T) {
		pMock := new(presenter.MockMacauPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockMacauGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.MacauPhasePlay)
		gameMock.On("IsHumanTurn").Return(false).Once()
		gameMock.On("CpuPlay").Return()
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewMacauInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertCalled(t, "CpuPlay")
	})

	t.Run("unknown phase breaks", func(t *testing.T) {
		pMock := new(presenter.MockMacauPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockMacauGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.MacauPhase(99))
		gameMock.On("IsHumanTurn").Return(false)

		ci := usecase.NewMacauInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})
}
