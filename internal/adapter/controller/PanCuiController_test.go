//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestPanCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockPanInteractor {
		m := new(mockUsecases.MockPanInteractor)
		m.On("GetConfig").Return(domain.DefaultPanConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("DrawFromStock").Return(mockOutput)
		m.On("DrawFromDiscard").Return(mockOutput)
		m.On("Meld", mock.Anything).Return(mockOutput)
		m.On("Layoff", mock.Anything, mock.Anything, mock.Anything).Return(mockOutput)
		m.On("Discard", mock.Anything).Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit q", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewPanCuiController(newMock()).Exec("q"))
	})

	t.Run("reset r preserves config", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewPanCuiController(m).Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultPanConfig())
	})

	t.Run("drawstock ds", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewPanCuiController(m).Exec("ds"))
		m.AssertCalled(t, "DrawFromStock")
	})

	t.Run("drawdiscard dd", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewPanCuiController(m).Exec("dd"))
		m.AssertCalled(t, "DrawFromDiscard")
	})

	t.Run("meld m with indices", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewPanCuiController(m).Exec("m 0 1 2"))
		m.AssertCalled(t, "Meld", []int{0, 1, 2})
	})

	t.Run("meld with comma-separated indices", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewPanCuiController(m).Exec("meld 0,1,2"))
		m.AssertCalled(t, "Meld", []int{0, 1, 2})
	})

	t.Run("layoff lo with 3 ints", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewPanCuiController(m).Exec("lo 1 0 2"))
		m.AssertCalled(t, "Layoff", 1, 0, 2)
	})

	t.Run("layoff lo missing args", func(t *testing.T) {
		result := controller.NewPanCuiController(newMock()).Exec("lo 0 1")
		assert.Contains(t, result, "layoff requires")
	})

	t.Run("discard d with index", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewPanCuiController(m).Exec("d 3"))
		m.AssertCalled(t, "Discard", 3)
	})

	t.Run("discard d no args", func(t *testing.T) {
		result := controller.NewPanCuiController(newMock()).Exec("d")
		assert.Contains(t, result, msgCardIndexRequired())
	})

	t.Run("nextround nr", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewPanCuiController(m).Exec("nr"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setplayers pc valid", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewPanCuiController(m).Exec("pc 5"))
		cfg := domain.DefaultPanConfig()
		cfg.PlayerCount = 5
		m.AssertCalled(t, "ResetWithConfig", cfg)
	})

	t.Run("setplayers pc out of range", func(t *testing.T) {
		result := controller.NewPanCuiController(newMock()).Exec("pc 9")
		assert.Contains(t, result, "Invalid player count")
	})

	t.Run("setdifficulty sd valid", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewPanCuiController(m).Exec("sd 2"))
		cfg := domain.DefaultPanConfig()
		cfg.CpuDifficulty = domain.PanCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", cfg)
	})

	t.Run("setdifficulty sd out of range", func(t *testing.T) {
		result := controller.NewPanCuiController(newMock()).Exec("sd 9")
		assert.Contains(t, result, "Invalid CPU difficulty")
	})

	t.Run("setrounds sr valid", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewPanCuiController(m).Exec("sr 5"))
		cfg := domain.DefaultPanConfig()
		cfg.TargetRounds = 5
		m.AssertCalled(t, "ResetWithConfig", cfg)
	})

	t.Run("setrounds sr invalid", func(t *testing.T) {
		result := controller.NewPanCuiController(newMock()).Exec("sr 0")
		assert.Contains(t, result, "Invalid target rounds")
	})

	t.Run("action log l", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewPanCuiController(m).Exec("l"))
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		assert.NotEmpty(t, controller.NewPanCuiController(newMock()).Exec("garbage"))
	})

	t.Run("empty command", func(t *testing.T) {
		assert.NotEmpty(t, controller.NewPanCuiController(newMock()).Exec(""))
	})
}
