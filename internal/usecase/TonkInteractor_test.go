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

func TestNewTonkInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockTonkPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "TonkInteractor: g must not be nil", func() {
			usecase.NewTonkInteractor(nil, pMock)
		})
	})

	t.Run("panics when gp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockTonkGame)
		assert.PanicsWithValue(t, "TonkInteractor: gp must not be nil", func() {
			usecase.NewTonkInteractor(gameMock, nil)
		})
	})
}

func TestTonkInteractor_Reset(t *testing.T) {
	mockOutput := `{"phase":0}`

	pMock := new(presenter.MockTonkPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockTonkGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TonkPhaseDraw)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewTonkInteractor(gameMock, pMock)
	result := ci.Reset()
	assert.Equal(t, mockOutput, result)
	gameMock.AssertCalled(t, "Reset")
}

func TestTonkInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockTonkPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockTonkGame)
		cfg := domain.TonkConfig{CpuDifficulty: domain.TonkCpuDifficultyHard, PointLimit: 100}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.TonkPhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewTonkInteractor(gameMock, pMock)
		result := ci.ResetWithConfig(cfg)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config returns error without resetting", func(t *testing.T) {
		pMock := new(presenter.MockTonkPresenter)
		gameMock := new(interfaces.MockTonkGame)
		pMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")

		ci := usecase.NewTonkInteractor(gameMock, pMock)
		cfg := domain.TonkConfig{CpuDifficulty: domain.TonkCpuDifficulty(-1), PointLimit: 50}
		result := ci.ResetWithConfig(cfg)
		assert.Equal(t, "validation error", result)
		gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
	})
}

func TestTonkInteractor_DrawFromStock(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ended", func(t *testing.T) {
		pMock := new(presenter.MockTonkPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockTonkGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewTonkInteractor(gameMock, pMock)
		result := ci.DrawFromStock()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerDrawFromStock")
	})

	t.Run("not human turn", func(t *testing.T) {
		pMock := new(presenter.MockTonkPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockTonkGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		ci := usecase.NewTonkInteractor(gameMock, pMock)
		result := ci.DrawFromStock()
		assert.Equal(t, mockOutput, result)
	})

	t.Run("draw error", func(t *testing.T) {
		drawErr := errors.New("draw error")
		pMock := new(presenter.MockTonkPresenter)
		pMock.On("Output", mock.Anything, drawErr).Return(mockOutput)
		gameMock := new(interfaces.MockTonkGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDrawFromStock").Return(drawErr)

		ci := usecase.NewTonkInteractor(gameMock, pMock)
		result := ci.DrawFromStock()
		assert.Equal(t, mockOutput, result)
	})

	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockTonkPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockTonkGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerDrawFromStock").Return(nil)
		gameMock.On("GetPhase").Return(domain.TonkPhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewTonkInteractor(gameMock, pMock)
		result := ci.DrawFromStock()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerDrawFromStock")
	})
}

func TestTonkInteractor_DrawFromDiscard(t *testing.T) {
	mockOutput := `{"phase":0}`

	pMock := new(presenter.MockTonkPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockTonkGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerDrawFromDiscard").Return(nil)
	gameMock.On("GetPhase").Return(domain.TonkPhaseDraw)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewTonkInteractor(gameMock, pMock)
	result := ci.DrawFromDiscard()
	assert.Equal(t, mockOutput, result)
}

func TestTonkInteractor_DrawFromDiscard_Error(t *testing.T) {
	mockOutput := `{"phase":0}`
	drawErr := errors.New("err")
	pMock := new(presenter.MockTonkPresenter)
	pMock.On("Output", mock.Anything, drawErr).Return(mockOutput)
	gameMock := new(interfaces.MockTonkGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerDrawFromDiscard").Return(drawErr)

	ci := usecase.NewTonkInteractor(gameMock, pMock)
	result := ci.DrawFromDiscard()
	assert.Equal(t, mockOutput, result)
}

func TestTonkInteractor_DrawFromDiscard_GameEnded(t *testing.T) {
	mockOutput := `{}`
	pMock := new(presenter.MockTonkPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockTonkGame)
	gameMock.On("GetGameEndFlag").Return(true)

	ci := usecase.NewTonkInteractor(gameMock, pMock)
	assert.Equal(t, mockOutput, ci.DrawFromDiscard())
}

func TestTonkInteractor_DrawFromDiscard_NotHumanTurn(t *testing.T) {
	mockOutput := `{}`
	pMock := new(presenter.MockTonkPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockTonkGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	ci := usecase.NewTonkInteractor(gameMock, pMock)
	assert.Equal(t, mockOutput, ci.DrawFromDiscard())
}

func TestTonkInteractor_Discard(t *testing.T) {
	mockOutput := `{}`

	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockTonkPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockTonkGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerDiscard", 0).Return(nil)
		gameMock.On("GetPhase").Return(domain.TonkPhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewTonkInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Discard(0))
	})

	t.Run("error", func(t *testing.T) {
		err := errors.New("e")
		pMock := new(presenter.MockTonkPresenter)
		pMock.On("Output", mock.Anything, err).Return(mockOutput)
		gameMock := new(interfaces.MockTonkGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDiscard", 0).Return(err)

		ci := usecase.NewTonkInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Discard(0))
	})

	t.Run("game ended", func(t *testing.T) {
		pMock := new(presenter.MockTonkPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockTonkGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewTonkInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Discard(0))
	})
}

func TestTonkInteractor_Knock(t *testing.T) {
	mockOutput := `{}`

	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockTonkPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockTonkGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerKnock", 1).Return(nil)
		gameMock.On("GetPhase").Return(domain.TonkPhaseRoundEnd)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewTonkInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Knock(1))
	})

	t.Run("error", func(t *testing.T) {
		err := errors.New("e")
		pMock := new(presenter.MockTonkPresenter)
		pMock.On("Output", mock.Anything, err).Return(mockOutput)
		gameMock := new(interfaces.MockTonkGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerKnock", 1).Return(err)

		ci := usecase.NewTonkInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Knock(1))
	})

	t.Run("game ended", func(t *testing.T) {
		pMock := new(presenter.MockTonkPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockTonkGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewTonkInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Knock(0))
	})
}

func TestTonkInteractor_NextRound(t *testing.T) {
	mockOutput := `{}`

	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockTonkPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockTonkGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextRound").Return()
		gameMock.On("GetPhase").Return(domain.TonkPhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewTonkInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.NextRound())
		gameMock.AssertCalled(t, "NextRound")
	})

	t.Run("game ended", func(t *testing.T) {
		pMock := new(presenter.MockTonkPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockTonkGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewTonkInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.NextRound())
		gameMock.AssertNotCalled(t, "NextRound")
	})
}

func TestTonkInteractor_GetConfig(t *testing.T) {
	pMock := new(presenter.MockTonkPresenter)
	gameMock := new(interfaces.MockTonkGame)
	cfg := domain.TonkConfig{CpuDifficulty: domain.TonkCpuDifficultyHard, PointLimit: 200}
	gameMock.On("GetConfig").Return(cfg)

	ci := usecase.NewTonkInteractor(gameMock, pMock)
	assert.Equal(t, cfg, ci.GetConfig())
}

func TestTonkInteractor_ActionLog(t *testing.T) {
	pMock := new(presenter.MockTonkPresenter)
	pMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockTonkGame)

	ci := usecase.NewTonkInteractor(gameMock, pMock)
	assert.Equal(t, "log", ci.ActionLog())
}

func TestTonkInteractor_RunCpuTurns_LoopsThenExits(t *testing.T) {
	mockOutput := `{}`
	pMock := new(presenter.MockTonkPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockTonkGame)
	gameMock.On("Reset").Return()
	// CPU turn, then human turn
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TonkPhaseDraw)
	gameMock.On("IsHumanTurn").Return(false).Once()
	gameMock.On("CpuPlay").Return().Once()
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewTonkInteractor(gameMock, pMock)
	result := ci.Reset()
	assert.Equal(t, mockOutput, result)
	gameMock.AssertCalled(t, "CpuPlay")
}

func TestTonkInteractor_RunCpuTurns_ExitsOnRoundEnd(t *testing.T) {
	mockOutput := `{}`
	pMock := new(presenter.MockTonkPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockTonkGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TonkPhaseRoundEnd)

	ci := usecase.NewTonkInteractor(gameMock, pMock)
	assert.Equal(t, mockOutput, ci.Reset())
	gameMock.AssertNotCalled(t, "CpuPlay")
}

func TestRestoreTonkInteractor(t *testing.T) {
	pMock := new(presenter.MockTonkPresenter)
	g := domain.NewDefaultTonk()
	data, err := g.MarshalJSON()
	assert.NoError(t, err)

	ci, err := usecase.RestoreTonkInteractor(data, pMock)
	assert.NoError(t, err)
	assert.NotNil(t, ci)
}

func TestRestoreTonkInteractor_InvalidJSON(t *testing.T) {
	pMock := new(presenter.MockTonkPresenter)
	_, err := usecase.RestoreTonkInteractor([]byte("not json"), pMock)
	assert.Error(t, err)
}
