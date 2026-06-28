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

func TestAllFoursCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockAllFoursInteractor {
		m := new(mockUsecases.MockAllFoursInteractor)
		m.On("GetConfig").Return(domain.DefaultAllFoursConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Beg", mock.Anything).Return(mockOutput)
		m.On("RespondBeg", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit q", func(t *testing.T) {
		c := controller.NewAllFoursCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})
	t.Run("reset r", func(t *testing.T) {
		m := newMock()
		c := controller.NewAllFoursCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultAllFoursConfig())
	})
	t.Run("stand", func(t *testing.T) {
		m := newMock()
		c := controller.NewAllFoursCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("stand"))
		m.AssertCalled(t, "Beg", false)
	})
	t.Run("beg", func(t *testing.T) {
		m := newMock()
		c := controller.NewAllFoursCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("beg"))
		m.AssertCalled(t, "Beg", true)
	})
	t.Run("gift", func(t *testing.T) {
		m := newMock()
		c := controller.NewAllFoursCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("gift"))
		m.AssertCalled(t, "RespondBeg", false)
	})
	t.Run("run", func(t *testing.T) {
		m := newMock()
		c := controller.NewAllFoursCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("run"))
		m.AssertCalled(t, "RespondBeg", true)
	})
	t.Run("play 0", func(t *testing.T) {
		m := newMock()
		c := controller.NewAllFoursCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 0"))
		m.AssertCalled(t, "Play", 0)
	})
	t.Run("play no args", func(t *testing.T) {
		c := controller.NewAllFoursCuiController(newMock())
		assert.NotEqual(t, mockOutput, c.Exec("p"))
	})
	t.Run("next n", func(t *testing.T) {
		m := newMock()
		c := controller.NewAllFoursCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		m.AssertCalled(t, "NextTrick")
	})
	t.Run("nextround nr", func(t *testing.T) {
		m := newMock()
		c := controller.NewAllFoursCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextRound")
	})
	t.Run("setdifficulty sd", func(t *testing.T) {
		m := newMock()
		c := controller.NewAllFoursCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultAllFoursConfig()
		expected.CpuDifficulty = domain.AllFoursCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})
	t.Run("setlimit sl", func(t *testing.T) {
		m := newMock()
		c := controller.NewAllFoursCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sl 9"))
		expected := domain.DefaultAllFoursConfig()
		expected.PointLimit = 9
		m.AssertCalled(t, "ResetWithConfig", expected)
	})
	t.Run("hint h", func(t *testing.T) {
		m := newMock()
		c := controller.NewAllFoursCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		m.AssertCalled(t, "Hint")
	})
	t.Run("log l", func(t *testing.T) {
		m := newMock()
		c := controller.NewAllFoursCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("l"))
		m.AssertCalled(t, "ActionLog")
	})
	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewAllFoursCuiController(newMock())
		assert.NotEqual(t, mockOutput, c.Exec("xyz"))
	})
}
