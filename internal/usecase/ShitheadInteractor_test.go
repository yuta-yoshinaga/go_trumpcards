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

func newTestShithead() *domain.Shithead {
	return domain.NewDefaultShithead()
}

func TestNewShitheadInteractor_NilGuards(t *testing.T) {
	ppMock := new(presenter.MockShitheadPresenter)
	t.Run("panics when sg is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "ShitheadInteractor: sg must not be nil", func() {
			usecase.NewShitheadInteractor(nil, ppMock)
		})
	})
	t.Run("panics when pp is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "ShitheadInteractor: pp must not be nil", func() {
			usecase.NewShitheadInteractor(newTestShithead(), nil)
		})
	})
}

func TestShitheadInteractor_Method(t *testing.T) {
	mockOutput := `{"players":[]}`
	ppMock := new(presenter.MockShitheadPresenter)
	ppMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	si := usecase.NewShitheadInteractor(newTestShithead(), ppMock)

	t.Run("Reset", func(t *testing.T) {
		assert.Equal(t, mockOutput, si.Reset())
	})

	t.Run("Play pickup", func(t *testing.T) {
		assert.Equal(t, mockOutput, si.Play([]int{}))
	})

	t.Run("Play indices", func(t *testing.T) {
		assert.Equal(t, mockOutput, si.Play([]int{0}))
	})

	t.Run("ResetWithConfig valid", func(t *testing.T) {
		cfg := domain.DefaultShitheadConfig()
		assert.Equal(t, mockOutput, si.ResetWithConfig(cfg))
	})

	t.Run("GetConfig", func(t *testing.T) {
		cfg := si.GetConfig()
		assert.NotEmpty(t, domain.ShitheadDifficultyNames[cfg.CpuDifficulty])
	})
}

func TestShitheadInteractor_ActionLog(t *testing.T) {
	ppMock := new(presenter.MockShitheadPresenter)
	gameMock := new(interfaces.MockShitheadGame)
	ppMock.On("ActionLogOutput", gameMock).Return(`{"entries":[]}`)

	si := usecase.NewShitheadInteractor(gameMock, ppMock)
	result := si.ActionLog()
	assert.Equal(t, `{"entries":[]}`, result)
	ppMock.AssertExpectations(t)
}

func TestShitheadInteractor_MockGame(t *testing.T) {
	mockOutput := `{"players":[]}`
	ppMock := new(presenter.MockShitheadPresenter)
	ppMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)

	gameMock := new(interfaces.MockShitheadGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("CpuPlay").Return()
	gameMock.On("PlayerPlay", mock.Anything).Return(nil)
	gameMock.On("SetConfig", mock.Anything).Return()

	si := usecase.NewShitheadInteractor(gameMock, ppMock)

	t.Run("Reset calls game.Reset", func(t *testing.T) {
		result := si.Reset()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "Reset")
	})

	t.Run("Play calls PlayerPlay on human turn", func(t *testing.T) {
		result := si.Play([]int{0})
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerPlay", []int{0})
	})

	t.Run("Play returns early when not human turn", func(t *testing.T) {
		cpuMock := new(interfaces.MockShitheadGame)
		cpuMock.On("GetGameEndFlag").Return(false)
		cpuMock.On("IsHumanTurn").Return(false)
		cpuMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		siCpu := usecase.NewShitheadInteractor(cpuMock, ppMock)
		result := siCpu.Play([]int{0})
		assert.Equal(t, mockOutput, result)
		cpuMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
	})

	t.Run("GetConfig delegates", func(t *testing.T) {
		gameMock.On("GetConfig").Return(domain.DefaultShitheadConfig())
		cfg := si.GetConfig()
		assert.Equal(t, domain.DefaultShitheadConfig(), cfg)
		gameMock.AssertCalled(t, "GetConfig")
	})
}

func TestShitheadInteractor_ResetWithConfig_Invalid(t *testing.T) {
	ppMock := new(presenter.MockShitheadPresenter)
	ppMock.On("Output", mock.Anything, mock.Anything).Return("invalid")
	si := usecase.NewShitheadInteractor(newTestShithead(), ppMock)
	bad := domain.ShitheadConfig{CpuDifficulty: 99}
	result := si.ResetWithConfig(bad)
	assert.Equal(t, "invalid", result)
}

func TestShitheadInteractor_Snapshot(t *testing.T) {
	ppMock := new(presenter.MockShitheadPresenter)
	si := usecase.NewShitheadInteractor(newTestShithead(), ppMock)
	data, err := si.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreShitheadInteractor(t *testing.T) {
	ppMock := new(presenter.MockShitheadPresenter)
	ppMock.On("Output", mock.Anything, mock.Anything).Return("ok")
	si := usecase.NewShitheadInteractor(newTestShithead(), ppMock)
	si.Reset()
	data, err := si.Snapshot()
	assert.NoError(t, err)

	si2, err := usecase.RestoreShitheadInteractor(data, ppMock)
	assert.NoError(t, err)
	assert.NotNil(t, si2)
}
