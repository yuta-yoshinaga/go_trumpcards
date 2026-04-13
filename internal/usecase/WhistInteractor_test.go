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

func TestNewWhistInteractor_NilGuards(t *testing.T) {
	wpMock := new(presenter.MockWhistPresenter)

	t.Run("panics when w is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "WhistInteractor: w must not be nil", func() {
			usecase.NewWhistInteractor(nil, wpMock)
		})
	})

	t.Run("panics when wp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockWhistGame)
		assert.PanicsWithValue(t, "WhistInteractor: wp must not be nil", func() {
			usecase.NewWhistInteractor(gameMock, nil)
		})
	})
}

func TestWhistInteractor_Reset(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("reset runs CPU turns when human is not first", func(t *testing.T) {
		wpMock := new(presenter.MockWhistPresenter)
		wpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockWhistGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.WhistPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		wi := usecase.NewWhistInteractor(gameMock, wpMock)
		result := wi.Reset()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "Reset")
	})
}

func TestWhistInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("sets config then resets", func(t *testing.T) {
		wpMock := new(presenter.MockWhistPresenter)
		wpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockWhistGame)
		cfg := domain.WhistConfig{CpuDifficulty: domain.WhistCpuDifficultyHard, PointLimit: 10}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.WhistPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		wi := usecase.NewWhistInteractor(gameMock, wpMock)
		result := wi.ResetWithConfig(cfg)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "SetConfig", cfg)
		gameMock.AssertCalled(t, "Reset")
	})
}

func TestWhistInteractor_ResetWithConfig_ValidationError(t *testing.T) {
	wpMock := new(presenter.MockWhistPresenter)
	gameMock := new(interfaces.MockWhistGame)
	wpMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")

	wi := usecase.NewWhistInteractor(gameMock, wpMock)
	invalidCfg := domain.WhistConfig{CpuDifficulty: 99, PointLimit: 5}
	result := wi.ResetWithConfig(invalidCfg)
	assert.Equal(t, "validation error", result)
	gameMock.AssertNotCalled(t, "Reset")
}

func TestWhistInteractor_Play(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("valid play", func(t *testing.T) {
		wpMock := new(presenter.MockWhistPresenter)
		wpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockWhistGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerPlay", 2).Return(nil)
		gameMock.On("GetPhase").Return(domain.WhistPhasePlay)

		wi := usecase.NewWhistInteractor(gameMock, wpMock)
		result := wi.Play(2)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerPlay", 2)
	})

	t.Run("play error", func(t *testing.T) {
		wpMock := new(presenter.MockWhistPresenter)
		wpMock.On("Output", mock.Anything, mock.Anything).Return("error output")
		gameMock := new(interfaces.MockWhistGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerPlay", 0).Return(errors.New("invalid"))

		wi := usecase.NewWhistInteractor(gameMock, wpMock)
		result := wi.Play(0)
		assert.Equal(t, "error output", result)
	})

	t.Run("game ended guard", func(t *testing.T) {
		wpMock := new(presenter.MockWhistPresenter)
		wpMock.On("Output", mock.Anything, mock.Anything).Return("ended")
		gameMock := new(interfaces.MockWhistGame)
		gameMock.On("GetGameEndFlag").Return(true)

		wi := usecase.NewWhistInteractor(gameMock, wpMock)
		result := wi.Play(0)
		assert.Equal(t, "ended", result)
		gameMock.AssertNotCalled(t, "PlayerPlay")
	})
}

func TestWhistInteractor_NextTrick(t *testing.T) {
	mockOutput := `{"phase":0}`
	wpMock := new(presenter.MockWhistPresenter)
	wpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockWhistGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.WhistPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	wi := usecase.NewWhistInteractor(gameMock, wpMock)
	result := wi.NextTrick()
	assert.Equal(t, mockOutput, result)
	gameMock.AssertCalled(t, "NextTrick")
}

func TestWhistInteractor_NextRound(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("scores and starts next round", func(t *testing.T) {
		wpMock := new(presenter.MockWhistPresenter)
		wpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockWhistGame)
		gameMock.On("ScoreRound").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextRound").Return()
		gameMock.On("GetPhase").Return(domain.WhistPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		wi := usecase.NewWhistInteractor(gameMock, wpMock)
		result := wi.NextRound()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "ScoreRound")
		gameMock.AssertCalled(t, "NextRound")
	})

	t.Run("game ends after scoring", func(t *testing.T) {
		wpMock := new(presenter.MockWhistPresenter)
		wpMock.On("Output", mock.Anything, mock.Anything).Return("game ended")
		gameMock := new(interfaces.MockWhistGame)
		gameMock.On("ScoreRound").Return()
		gameMock.On("GetGameEndFlag").Return(true)

		wi := usecase.NewWhistInteractor(gameMock, wpMock)
		result := wi.NextRound()
		assert.Equal(t, "game ended", result)
		gameMock.AssertCalled(t, "ScoreRound")
		gameMock.AssertNotCalled(t, "NextRound")
	})
}

func TestWhistInteractor_GetConfig(t *testing.T) {
	cfg := domain.DefaultWhistConfig()
	gameMock := new(interfaces.MockWhistGame)
	gameMock.On("GetConfig").Return(cfg)
	wpMock := new(presenter.MockWhistPresenter)

	wi := usecase.NewWhistInteractor(gameMock, wpMock)
	assert.Equal(t, cfg, wi.GetConfig())
}

func TestWhistInteractor_Hint(t *testing.T) {
	wpMock := new(presenter.MockWhistPresenter)
	gameMock := new(interfaces.MockWhistGame)
	wpMock.On("HintOutput", gameMock).Return("hint output")

	wi := usecase.NewWhistInteractor(gameMock, wpMock)
	assert.Equal(t, "hint output", wi.Hint())
}

func TestWhistInteractor_ActionLog(t *testing.T) {
	wpMock := new(presenter.MockWhistPresenter)
	gameMock := new(interfaces.MockWhistGame)
	wpMock.On("ActionLogOutput", gameMock).Return("log output")

	wi := usecase.NewWhistInteractor(gameMock, wpMock)
	assert.Equal(t, "log output", wi.ActionLog())
}
