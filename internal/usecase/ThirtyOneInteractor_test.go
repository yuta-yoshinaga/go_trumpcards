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

func TestNewThirtyOneInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockThirtyOnePresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "ThirtyOneInteractor: g must not be nil", func() {
			usecase.NewThirtyOneInteractor(nil, pMock)
		})
	})

	t.Run("panics when gp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockThirtyOneGame)
		assert.PanicsWithValue(t, "ThirtyOneInteractor: gp must not be nil", func() {
			usecase.NewThirtyOneInteractor(gameMock, nil)
		})
	})
}

func TestThirtyOneInteractor_Reset(t *testing.T) {
	mockOutput := `{"phase":0}`
	pMock := new(presenter.MockThirtyOnePresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockThirtyOneGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.ThirtyOnePhaseDraw)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewThirtyOneInteractor(gameMock, pMock)
	assert.Equal(t, mockOutput, ci.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestThirtyOneInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockThirtyOnePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockThirtyOneGame)
		cfg := domain.ThirtyOneConfig{CpuDifficulty: domain.ThirtyOneCpuDifficultyHard, InitialLives: 5}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.ThirtyOnePhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewThirtyOneInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.ResetWithConfig(cfg))
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config returns error without resetting", func(t *testing.T) {
		pMock := new(presenter.MockThirtyOnePresenter)
		gameMock := new(interfaces.MockThirtyOneGame)
		pMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")

		ci := usecase.NewThirtyOneInteractor(gameMock, pMock)
		cfg := domain.ThirtyOneConfig{CpuDifficulty: domain.ThirtyOneCpuDifficulty(-1), InitialLives: 3}
		assert.Equal(t, "validation error", ci.ResetWithConfig(cfg))
		gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
	})
}

func TestThirtyOneInteractor_DrawFromStock(t *testing.T) {
	mockOutput := `{}`

	t.Run("game ended", func(t *testing.T) {
		pMock := new(presenter.MockThirtyOnePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockThirtyOneGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewThirtyOneInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.DrawFromStock())
		gameMock.AssertNotCalled(t, "PlayerDrawFromStock")
	})

	t.Run("not human turn", func(t *testing.T) {
		pMock := new(presenter.MockThirtyOnePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockThirtyOneGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		ci := usecase.NewThirtyOneInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.DrawFromStock())
	})

	t.Run("draw error", func(t *testing.T) {
		drawErr := errors.New("draw error")
		pMock := new(presenter.MockThirtyOnePresenter)
		pMock.On("Output", mock.Anything, drawErr).Return(mockOutput)
		gameMock := new(interfaces.MockThirtyOneGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDrawFromStock").Return(drawErr)

		ci := usecase.NewThirtyOneInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.DrawFromStock())
	})

	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockThirtyOnePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockThirtyOneGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerDrawFromStock").Return(nil)
		gameMock.On("GetPhase").Return(domain.ThirtyOnePhaseDiscard)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewThirtyOneInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.DrawFromStock())
		gameMock.AssertCalled(t, "PlayerDrawFromStock")
	})
}

func TestThirtyOneInteractor_DrawFromDiscard(t *testing.T) {
	mockOutput := `{}`

	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockThirtyOnePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockThirtyOneGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerDrawFromDiscard").Return(nil)
		gameMock.On("GetPhase").Return(domain.ThirtyOnePhaseDiscard)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewThirtyOneInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.DrawFromDiscard())
	})

	t.Run("error", func(t *testing.T) {
		drawErr := errors.New("err")
		pMock := new(presenter.MockThirtyOnePresenter)
		pMock.On("Output", mock.Anything, drawErr).Return(mockOutput)
		gameMock := new(interfaces.MockThirtyOneGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDrawFromDiscard").Return(drawErr)

		ci := usecase.NewThirtyOneInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.DrawFromDiscard())
	})

	t.Run("game ended", func(t *testing.T) {
		pMock := new(presenter.MockThirtyOnePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockThirtyOneGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewThirtyOneInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.DrawFromDiscard())
	})
}

func TestThirtyOneInteractor_Discard(t *testing.T) {
	mockOutput := `{}`

	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockThirtyOnePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockThirtyOneGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerDiscard", 0).Return(nil)
		gameMock.On("GetPhase").Return(domain.ThirtyOnePhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewThirtyOneInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Discard(0))
	})

	t.Run("error", func(t *testing.T) {
		err := errors.New("e")
		pMock := new(presenter.MockThirtyOnePresenter)
		pMock.On("Output", mock.Anything, err).Return(mockOutput)
		gameMock := new(interfaces.MockThirtyOneGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDiscard", 0).Return(err)

		ci := usecase.NewThirtyOneInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Discard(0))
	})
}

func TestThirtyOneInteractor_Knock(t *testing.T) {
	mockOutput := `{}`

	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockThirtyOnePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockThirtyOneGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerKnock").Return(nil)
		gameMock.On("GetPhase").Return(domain.ThirtyOnePhaseRoundEnd)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewThirtyOneInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Knock())
		gameMock.AssertCalled(t, "PlayerKnock")
	})

	t.Run("error", func(t *testing.T) {
		err := errors.New("e")
		pMock := new(presenter.MockThirtyOnePresenter)
		pMock.On("Output", mock.Anything, err).Return(mockOutput)
		gameMock := new(interfaces.MockThirtyOneGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerKnock").Return(err)

		ci := usecase.NewThirtyOneInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Knock())
	})

	t.Run("game ended", func(t *testing.T) {
		pMock := new(presenter.MockThirtyOnePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockThirtyOneGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewThirtyOneInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Knock())
	})
}

func TestThirtyOneInteractor_NextRound(t *testing.T) {
	mockOutput := `{}`

	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockThirtyOnePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockThirtyOneGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextRound").Return()
		gameMock.On("GetPhase").Return(domain.ThirtyOnePhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewThirtyOneInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.NextRound())
		gameMock.AssertCalled(t, "NextRound")
	})

	t.Run("game ended", func(t *testing.T) {
		pMock := new(presenter.MockThirtyOnePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockThirtyOneGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewThirtyOneInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.NextRound())
		gameMock.AssertNotCalled(t, "NextRound")
	})
}

func TestThirtyOneInteractor_GetConfigAndLog(t *testing.T) {
	pMock := new(presenter.MockThirtyOnePresenter)
	pMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockThirtyOneGame)
	cfg := domain.ThirtyOneConfig{CpuDifficulty: domain.ThirtyOneCpuDifficultyHard, InitialLives: 2}
	gameMock.On("GetConfig").Return(cfg)

	ci := usecase.NewThirtyOneInteractor(gameMock, pMock)
	assert.Equal(t, cfg, ci.GetConfig())
	assert.Equal(t, "log", ci.ActionLog())
}

func TestThirtyOneInteractor_RunCpuTurns_ExitsOnRoundEnd(t *testing.T) {
	mockOutput := `{}`
	pMock := new(presenter.MockThirtyOnePresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockThirtyOneGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.ThirtyOnePhaseRoundEnd)

	ci := usecase.NewThirtyOneInteractor(gameMock, pMock)
	assert.Equal(t, mockOutput, ci.Reset())
	gameMock.AssertNotCalled(t, "CpuPlay")
}

func TestRestoreThirtyOneInteractor(t *testing.T) {
	pMock := new(presenter.MockThirtyOnePresenter)
	g := domain.NewDefaultThirtyOne()
	data, err := g.MarshalJSON()
	assert.NoError(t, err)

	ci, err := usecase.RestoreThirtyOneInteractor(data, pMock)
	assert.NoError(t, err)
	assert.NotNil(t, ci)
}

func TestRestoreThirtyOneInteractor_InvalidJSON(t *testing.T) {
	pMock := new(presenter.MockThirtyOnePresenter)
	_, err := usecase.RestoreThirtyOneInteractor([]byte("not json"), pMock)
	assert.Error(t, err)
}

// **ヒントがプレゼンターまで通っていること (#4806)。**Barbu / Macau と同じ配線。
func TestThirtyOneInteractor_Hint(t *testing.T) {
	pMock := new(presenter.MockThirtyOnePresenter)
	gameMock := new(interfaces.MockThirtyOneGame)
	pMock.On("HintOutput", gameMock).Return("hint")

	ci := usecase.NewThirtyOneInteractor(gameMock, pMock)
	assert.Equal(t, "hint", ci.Hint())
	pMock.AssertExpectations(t)
}
