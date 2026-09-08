//go:build test
// +build test

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

func TestNewTongitsInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockTongitsPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "TongitsInteractor: g must not be nil", func() {
			usecase.NewTongitsInteractor(nil, pMock)
		})
	})

	t.Run("panics when gp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockTongitsGame)
		assert.PanicsWithValue(t, "TongitsInteractor: gp must not be nil", func() {
			usecase.NewTongitsInteractor(gameMock, nil)
		})
	})
}

func TestTongitsInteractor_Reset(t *testing.T) {
	mockOutput := `{"phase":0}`

	pMock := new(presenter.MockTongitsPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockTongitsGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TongitsPhaseDraw)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewTongitsInteractor(gameMock, pMock)
	result := ci.Reset()
	assert.Equal(t, mockOutput, result)
	gameMock.AssertCalled(t, "Reset")
}

func TestTongitsInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockTongitsPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockTongitsGame)
		cfg := domain.TongitsConfig{CpuDifficulty: domain.TongitsCpuDifficultyHard, PointLimit: 100}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.TongitsPhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewTongitsInteractor(gameMock, pMock)
		result := ci.ResetWithConfig(cfg)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config returns error without resetting", func(t *testing.T) {
		pMock := new(presenter.MockTongitsPresenter)
		gameMock := new(interfaces.MockTongitsGame)
		pMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")

		ci := usecase.NewTongitsInteractor(gameMock, pMock)
		cfg := domain.TongitsConfig{CpuDifficulty: domain.TongitsCpuDifficulty(-1), PointLimit: 50}
		result := ci.ResetWithConfig(cfg)
		assert.Equal(t, "validation error", result)
		gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
	})
}

func TestTongitsInteractor_DrawFromStock(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ended", func(t *testing.T) {
		pMock := new(presenter.MockTongitsPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockTongitsGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewTongitsInteractor(gameMock, pMock)
		result := ci.DrawFromStock()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerDrawFromStock")
	})

	t.Run("not human turn", func(t *testing.T) {
		pMock := new(presenter.MockTongitsPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockTongitsGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		ci := usecase.NewTongitsInteractor(gameMock, pMock)
		result := ci.DrawFromStock()
		assert.Equal(t, mockOutput, result)
	})

	t.Run("draw error", func(t *testing.T) {
		drawErr := errors.New("draw error")
		pMock := new(presenter.MockTongitsPresenter)
		pMock.On("Output", mock.Anything, drawErr).Return(mockOutput)
		gameMock := new(interfaces.MockTongitsGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDrawFromStock").Return(drawErr)

		ci := usecase.NewTongitsInteractor(gameMock, pMock)
		result := ci.DrawFromStock()
		assert.Equal(t, mockOutput, result)
	})

	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockTongitsPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockTongitsGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerDrawFromStock").Return(nil)
		gameMock.On("GetPhase").Return(domain.TongitsPhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewTongitsInteractor(gameMock, pMock)
		result := ci.DrawFromStock()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerDrawFromStock")
	})
}

func TestTongitsInteractor_DrawFromDiscard(t *testing.T) {
	mockOutput := `{"phase":0}`

	pMock := new(presenter.MockTongitsPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockTongitsGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerDrawFromDiscard").Return(nil)
	gameMock.On("GetPhase").Return(domain.TongitsPhaseDraw)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewTongitsInteractor(gameMock, pMock)
	result := ci.DrawFromDiscard()
	assert.Equal(t, mockOutput, result)
}

func TestTongitsInteractor_DrawFromDiscard_Error(t *testing.T) {
	mockOutput := `{"phase":0}`
	drawErr := errors.New("err")
	pMock := new(presenter.MockTongitsPresenter)
	pMock.On("Output", mock.Anything, drawErr).Return(mockOutput)
	gameMock := new(interfaces.MockTongitsGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerDrawFromDiscard").Return(drawErr)

	ci := usecase.NewTongitsInteractor(gameMock, pMock)
	result := ci.DrawFromDiscard()
	assert.Equal(t, mockOutput, result)
}

func TestTongitsInteractor_DrawFromDiscard_GameEnded(t *testing.T) {
	mockOutput := `{}`
	pMock := new(presenter.MockTongitsPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockTongitsGame)
	gameMock.On("GetGameEndFlag").Return(true)

	ci := usecase.NewTongitsInteractor(gameMock, pMock)
	assert.Equal(t, mockOutput, ci.DrawFromDiscard())
}

func TestTongitsInteractor_DrawFromDiscard_NotHumanTurn(t *testing.T) {
	mockOutput := `{}`
	pMock := new(presenter.MockTongitsPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockTongitsGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	ci := usecase.NewTongitsInteractor(gameMock, pMock)
	assert.Equal(t, mockOutput, ci.DrawFromDiscard())
}

func TestTongitsInteractor_Discard(t *testing.T) {
	mockOutput := `{}`

	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockTongitsPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockTongitsGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerDiscard", 0).Return(nil)
		gameMock.On("GetPhase").Return(domain.TongitsPhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewTongitsInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Discard(0))
	})

	t.Run("error", func(t *testing.T) {
		err := errors.New("e")
		pMock := new(presenter.MockTongitsPresenter)
		pMock.On("Output", mock.Anything, err).Return(mockOutput)
		gameMock := new(interfaces.MockTongitsGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDiscard", 0).Return(err)

		ci := usecase.NewTongitsInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Discard(0))
	})

	t.Run("game ended", func(t *testing.T) {
		pMock := new(presenter.MockTongitsPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockTongitsGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewTongitsInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Discard(0))
	})
}

func TestTongitsInteractor_Knock(t *testing.T) {
	mockOutput := `{}`

	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockTongitsPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockTongitsGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerKnock", 1).Return(nil)
		gameMock.On("GetPhase").Return(domain.TongitsPhaseRoundEnd)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewTongitsInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Knock(1))
	})

	t.Run("error", func(t *testing.T) {
		err := errors.New("e")
		pMock := new(presenter.MockTongitsPresenter)
		pMock.On("Output", mock.Anything, err).Return(mockOutput)
		gameMock := new(interfaces.MockTongitsGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerKnock", 1).Return(err)

		ci := usecase.NewTongitsInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Knock(1))
	})

	t.Run("game ended", func(t *testing.T) {
		pMock := new(presenter.MockTongitsPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockTongitsGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewTongitsInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Knock(0))
	})
}

func TestTongitsInteractor_NextRound(t *testing.T) {
	mockOutput := `{}`

	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockTongitsPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockTongitsGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextRound").Return()
		gameMock.On("GetPhase").Return(domain.TongitsPhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewTongitsInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.NextRound())
		gameMock.AssertCalled(t, "NextRound")
	})

	t.Run("game ended", func(t *testing.T) {
		pMock := new(presenter.MockTongitsPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockTongitsGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewTongitsInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.NextRound())
		gameMock.AssertNotCalled(t, "NextRound")
	})
}

func TestTongitsInteractor_GetConfig(t *testing.T) {
	pMock := new(presenter.MockTongitsPresenter)
	gameMock := new(interfaces.MockTongitsGame)
	cfg := domain.TongitsConfig{CpuDifficulty: domain.TongitsCpuDifficultyHard, PointLimit: 200}
	gameMock.On("GetConfig").Return(cfg)

	ci := usecase.NewTongitsInteractor(gameMock, pMock)
	assert.Equal(t, cfg, ci.GetConfig())
}

func TestTongitsInteractor_ActionLog(t *testing.T) {
	pMock := new(presenter.MockTongitsPresenter)
	pMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockTongitsGame)

	ci := usecase.NewTongitsInteractor(gameMock, pMock)
	assert.Equal(t, "log", ci.ActionLog())
}

func TestTongitsInteractor_RunCpuTurns_LoopsThenExits(t *testing.T) {
	mockOutput := `{}`
	pMock := new(presenter.MockTongitsPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockTongitsGame)
	gameMock.On("Reset").Return()
	// CPU turn, then human turn
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TongitsPhaseDraw)
	gameMock.On("IsHumanTurn").Return(false).Once()
	gameMock.On("CpuPlay").Return().Once()
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewTongitsInteractor(gameMock, pMock)
	result := ci.Reset()
	assert.Equal(t, mockOutput, result)
	gameMock.AssertCalled(t, "CpuPlay")
}

func TestTongitsInteractor_RunCpuTurns_ExitsOnRoundEnd(t *testing.T) {
	mockOutput := `{}`
	pMock := new(presenter.MockTongitsPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockTongitsGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TongitsPhaseRoundEnd)

	ci := usecase.NewTongitsInteractor(gameMock, pMock)
	assert.Equal(t, mockOutput, ci.Reset())
	gameMock.AssertNotCalled(t, "CpuPlay")
}

func TestRestoreTongitsInteractor(t *testing.T) {
	pMock := new(presenter.MockTongitsPresenter)
	g := domain.NewDefaultTongits()
	data, err := g.MarshalJSON()
	assert.NoError(t, err)

	ci, err := usecase.RestoreTongitsInteractor(data, pMock)
	assert.NoError(t, err)
	assert.NotNil(t, ci)
}

func TestRestoreTongitsInteractor_InvalidJSON(t *testing.T) {
	pMock := new(presenter.MockTongitsPresenter)
	_, err := usecase.RestoreTongitsInteractor([]byte("not json"), pMock)
	assert.Error(t, err)
}
