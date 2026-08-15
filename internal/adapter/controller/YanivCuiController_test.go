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

func TestYanivCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockYanivInteractor {
		m := new(mockUsecases.MockYanivInteractor)
		m.On("GetConfig").Return(domain.DefaultYanivConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Discard", mock.Anything).Return(mockOutput)
		m.On("DeclareYaniv").Return(mockOutput)
		m.On("DrawFromStock").Return(mockOutput)
		m.On("DrawFromPickup", mock.Anything).Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		c := controller.NewYanivCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	t.Run("reset", func(t *testing.T) {
		m := newMock()
		c := controller.NewYanivCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		assert.Equal(t, mockOutput, c.Exec("reset"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultYanivConfig())
	})

	t.Run("discard combo", func(t *testing.T) {
		m := newMock()
		c := controller.NewYanivCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("d 0 2 4"))
		m.AssertCalled(t, "Discard", []int{0, 2, 4})
		assert.Equal(t, mockOutput, c.Exec("discard 1"))
		m.AssertCalled(t, "Discard", []int{1})
	})

	t.Run("yaniv", func(t *testing.T) {
		m := newMock()
		c := controller.NewYanivCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("y"))
		assert.Equal(t, mockOutput, c.Exec("yaniv"))
		m.AssertCalled(t, "DeclareYaniv")
	})

	t.Run("draw stock and pickup", func(t *testing.T) {
		m := newMock()
		c := controller.NewYanivCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("ds"))
		assert.Equal(t, mockOutput, c.Exec("dp 1"))
		m.AssertCalled(t, "DrawFromStock")
		m.AssertCalled(t, "DrawFromPickup", 1)
	})

	t.Run("drawpickup errors", func(t *testing.T) {
		c := controller.NewYanivCuiController(newMock())
		assert.Contains(t, c.Exec("dp"), "required")
		assert.Contains(t, c.Exec("dp 5"), "Invalid pickup end")
	})

	t.Run("nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewYanivCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("nr"))
		assert.Equal(t, mockOutput, c.Exec("nextround"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewYanivCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultYanivConfig()
		expected.CpuDifficulty = domain.YanivCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty errors", func(t *testing.T) {
		c := controller.NewYanivCuiController(newMock())
		assert.Contains(t, c.Exec("sd"), msgCpuDifficultyRequired())
		assert.Contains(t, c.Exec("sd 9"), msgInvalidCpuDifficultyPrefix())
	})

	t.Run("setlimit valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewYanivCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sl 100"))
		expected := domain.DefaultYanivConfig()
		expected.ScoreLimit = 100
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setlimit errors", func(t *testing.T) {
		c := controller.NewYanivCuiController(newMock())
		assert.Contains(t, c.Exec("sl"), "required")
		assert.Contains(t, c.Exec("sl 9"), "Invalid score limit")
	})

	t.Run("log", func(t *testing.T) {
		m := newMock()
		c := controller.NewYanivCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("log"))
		assert.Equal(t, mockOutput, c.Exec("l"))
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewYanivCuiController(newMock())
		assert.Contains(t, c.Exec("unknown"), "コマンドが不明です")
	})
}
