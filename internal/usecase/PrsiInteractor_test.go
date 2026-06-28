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

func TestNewPrsiInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockPrsiPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "PrsiInteractor: g must not be nil", func() {
			usecase.NewPrsiInteractor(nil, pMock)
		})
	})

	t.Run("panics when gp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockPrsiGame)
		assert.PanicsWithValue(t, "PrsiInteractor: gp must not be nil", func() {
			usecase.NewPrsiInteractor(gameMock, nil)
		})
	})
}

func TestPrsiInteractor_Reset(t *testing.T) {
	mockOutput := `{"phase":0}`
	pMock := new(presenter.MockPrsiPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockPrsiGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.PrsiPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewPrsiInteractor(gameMock, pMock)
	assert.Equal(t, mockOutput, ci.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestPrsiInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("valid config sets config then resets", func(t *testing.T) {
		pMock := new(presenter.MockPrsiPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockPrsiGame)
		cfg := domain.PrsiConfig{CpuDifficulty: domain.PrsiCpuDifficultyHard}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.PrsiPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewPrsiInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.ResetWithConfig(cfg))
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config returns error without resetting", func(t *testing.T) {
		pMock := new(presenter.MockPrsiPresenter)
		gameMock := new(interfaces.MockPrsiGame)
		pMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")

		ci := usecase.NewPrsiInteractor(gameMock, pMock)
		cfg := domain.PrsiConfig{CpuDifficulty: domain.PrsiCpuDifficulty(-1)}
		assert.Equal(t, "validation error", ci.ResetWithConfig(cfg))
		gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
	})
}

func TestPrsiInteractor_Play(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ended returns output without playing", func(t *testing.T) {
		pMock := new(presenter.MockPrsiPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockPrsiGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewPrsiInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Play(0))
		gameMock.AssertNotCalled(t, "PlayerPlay")
	})

	t.Run("not human turn returns output without playing", func(t *testing.T) {
		pMock := new(presenter.MockPrsiPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockPrsiGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		ci := usecase.NewPrsiInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Play(0))
		gameMock.AssertNotCalled(t, "PlayerPlay")
	})

	t.Run("play error returned to presenter", func(t *testing.T) {
		playErr := errors.New("invalid play")
		pMock := new(presenter.MockPrsiPresenter)
		pMock.On("Output", mock.Anything, playErr).Return(mockOutput)
		gameMock := new(interfaces.MockPrsiGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerPlay", 0).Return(playErr)

		ci := usecase.NewPrsiInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Play(0))
	})

	t.Run("valid play runs CPU turns", func(t *testing.T) {
		pMock := new(presenter.MockPrsiPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockPrsiGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerPlay", 2).Return(nil)
		gameMock.On("GetPhase").Return(domain.PrsiPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewPrsiInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Play(2))
		gameMock.AssertCalled(t, "PlayerPlay", 2)
	})
}

func TestPrsiInteractor_Draw(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ended returns output without drawing", func(t *testing.T) {
		pMock := new(presenter.MockPrsiPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockPrsiGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewPrsiInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Draw())
		gameMock.AssertNotCalled(t, "PlayerDraw")
	})

	t.Run("draw error returned to presenter", func(t *testing.T) {
		drawErr := errors.New("draw error")
		pMock := new(presenter.MockPrsiPresenter)
		pMock.On("Output", mock.Anything, drawErr).Return(mockOutput)
		gameMock := new(interfaces.MockPrsiGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDraw").Return(drawErr)

		ci := usecase.NewPrsiInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Draw())
	})

	t.Run("valid draw runs CPU turns", func(t *testing.T) {
		pMock := new(presenter.MockPrsiPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockPrsiGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerDraw").Return(nil)
		gameMock.On("GetPhase").Return(domain.PrsiPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewPrsiInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Draw())
		gameMock.AssertCalled(t, "PlayerDraw")
	})
}

func TestPrsiInteractor_GetConfig(t *testing.T) {
	pMock := new(presenter.MockPrsiPresenter)
	gameMock := new(interfaces.MockPrsiGame)
	expected := domain.PrsiConfig{CpuDifficulty: domain.PrsiCpuDifficultyHard}
	gameMock.On("GetConfig").Return(expected)

	ci := usecase.NewPrsiInteractor(gameMock, pMock)
	assert.Equal(t, expected, ci.GetConfig())
}

func TestPrsiInteractor_ActionLog(t *testing.T) {
	pMock := new(presenter.MockPrsiPresenter)
	gameMock := new(interfaces.MockPrsiGame)
	pMock.On("ActionLogOutput", gameMock).Return(`{"entries":[]}`)

	ci := usecase.NewPrsiInteractor(gameMock, pMock)
	assert.Equal(t, `{"entries":[]}`, ci.ActionLog())
	pMock.AssertExpectations(t)
}

func TestPrsiInteractor_RunCpuTurns(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("stops when game ended", func(t *testing.T) {
		pMock := new(presenter.MockPrsiPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockPrsiGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(true)

		usecase.NewPrsiInteractor(gameMock, pMock).Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("stops when phase is not Play (GameEnd)", func(t *testing.T) {
		pMock := new(presenter.MockPrsiPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockPrsiGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.PrsiPhaseGameEnd)

		usecase.NewPrsiInteractor(gameMock, pMock).Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("Play phase with human turn breaks", func(t *testing.T) {
		pMock := new(presenter.MockPrsiPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockPrsiGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.PrsiPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		usecase.NewPrsiInteractor(gameMock, pMock).Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("Play phase with CPU turn calls CpuPlay then stops at human", func(t *testing.T) {
		pMock := new(presenter.MockPrsiPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockPrsiGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.PrsiPhasePlay)
		gameMock.On("IsHumanTurn").Return(false).Once()
		gameMock.On("CpuPlay").Return()
		gameMock.On("IsHumanTurn").Return(true)

		usecase.NewPrsiInteractor(gameMock, pMock).Reset()
		gameMock.AssertCalled(t, "CpuPlay")
	})
}
