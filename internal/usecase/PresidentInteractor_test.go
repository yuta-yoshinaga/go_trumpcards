package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newTestPresident() *domain.President {
	return domain.NewDefaultPresident()
}

func TestNewPresidentInteractor_NilGuards(t *testing.T) {
	ppMock := new(presenter.MockPresidentPresenter)
	t.Run("panics when pg is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "PresidentInteractor: pg must not be nil", func() {
			usecase.NewPresidentInteractor(nil, ppMock)
		})
	})
	t.Run("panics when pp is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "PresidentInteractor: pp must not be nil", func() {
			usecase.NewPresidentInteractor(newTestPresident(), nil)
		})
	})
}

func TestPresidentInteractor_Method(t *testing.T) {
	mockOutput := `{"players":[]}`
	ppMock := new(presenter.MockPresidentPresenter)
	ppMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	pi := usecase.NewPresidentInteractor(newTestPresident(), ppMock)

	t.Run("Reset", func(t *testing.T) {
		assert.Equal(t, mockOutput, pi.Reset())
	})

	t.Run("Hint delegates to the presenter", func(t *testing.T) {
		ppMock.On("HintOutput", mock.Anything).Return("hint_output")
		assert.Equal(t, "hint_output", pi.Hint())
	})

	t.Run("Play pass", func(t *testing.T) {
		assert.Equal(t, mockOutput, pi.Play([]int{}))
	})

	t.Run("Play indices", func(t *testing.T) {
		assert.Equal(t, mockOutput, pi.Play([]int{0}))
	})

	t.Run("ResetWithConfig", func(t *testing.T) {
		cfg := domain.DefaultPresidentConfig()
		assert.Equal(t, mockOutput, pi.ResetWithConfig(cfg))
	})

	t.Run("GetConfig", func(t *testing.T) {
		cfg := pi.GetConfig()
		assert.NotEmpty(t, domain.PresidentDifficultyNames[cfg.CpuDifficulty])
	})
}

func TestPresidentInteractor_ActionLog(t *testing.T) {
	ppMock := new(presenter.MockPresidentPresenter)
	gameMock := new(interfaces.MockPresidentGame)
	ppMock.On("ActionLogOutput", gameMock).Return(`{"entries":[]}`)

	pi := usecase.NewPresidentInteractor(gameMock, ppMock)
	result := pi.ActionLog()
	assert.Equal(t, `{"entries":[]}`, result)
	ppMock.AssertExpectations(t)
}

func TestPresidentInteractor_MockGame(t *testing.T) {
	mockOutput := `{"players":[]}`
	ppMock := new(presenter.MockPresidentPresenter)
	ppMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)

	gameMock := new(interfaces.MockPresidentGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("CpuPlay").Return()
	gameMock.On("PlayerPlay", mock.Anything).Return(nil)
	gameMock.On("SetConfig", mock.Anything).Return()

	pi := usecase.NewPresidentInteractor(gameMock, ppMock)

	t.Run("Reset calls game.Reset", func(t *testing.T) {
		result := pi.Reset()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "Reset")
	})

	t.Run("Play calls PlayerPlay on human turn", func(t *testing.T) {
		result := pi.Play([]int{0})
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerPlay", []int{0})
	})

	t.Run("Play returns early when not human turn", func(t *testing.T) {
		cpuMock := new(interfaces.MockPresidentGame)
		cpuMock.On("GetGameEndFlag").Return(false)
		cpuMock.On("IsHumanTurn").Return(false)
		cpuMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		piCpu := usecase.NewPresidentInteractor(cpuMock, ppMock)
		result := piCpu.Play([]int{0})
		assert.Equal(t, mockOutput, result)
		cpuMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
	})

	t.Run("GetConfig delegates", func(t *testing.T) {
		gameMock.On("GetConfig").Return(domain.DefaultPresidentConfig())
		cfg := pi.GetConfig()
		assert.Equal(t, domain.DefaultPresidentConfig(), cfg)
		gameMock.AssertCalled(t, "GetConfig")
	})
}

func TestPresidentInteractor_ResetWithConfig_InvalidReturnsOutput(t *testing.T) {
	ppMock := new(presenter.MockPresidentPresenter)
	ppMock.On("Output", mock.Anything, mock.Anything).Return("invalid")
	pi := usecase.NewPresidentInteractor(newTestPresident(), ppMock)
	// Invalid: CpuDifficulty out of range
	bad := domain.PresidentConfig{CpuDifficulty: 99}
	result := pi.ResetWithConfig(bad)
	assert.Equal(t, "invalid", result)
}
