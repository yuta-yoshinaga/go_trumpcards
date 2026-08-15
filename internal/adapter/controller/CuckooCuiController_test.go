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

func TestCuckooCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockCuckooInteractor {
		m := new(mockUsecases.MockCuckooInteractor)
		m.On("GetConfig").Return(domain.DefaultCuckooConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Keep").Return(mockOutput)
		m.On("Swap").Return(mockOutput)
		m.On("Refuse").Return(mockOutput)
		m.On("AcceptSwap").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		c := controller.NewCuckooCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	t.Run("reset", func(t *testing.T) {
		m := newMock()
		c := controller.NewCuckooCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		assert.Equal(t, mockOutput, c.Exec("reset"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultCuckooConfig())
	})

	t.Run("keep and swap", func(t *testing.T) {
		m := newMock()
		c := controller.NewCuckooCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("k"))
		assert.Equal(t, mockOutput, c.Exec("keep"))
		assert.Equal(t, mockOutput, c.Exec("s"))
		assert.Equal(t, mockOutput, c.Exec("swap"))
		m.AssertCalled(t, "Keep")
		m.AssertCalled(t, "Swap")
	})

	t.Run("refuse and accept", func(t *testing.T) {
		m := newMock()
		c := controller.NewCuckooCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("rf"))
		assert.Equal(t, mockOutput, c.Exec("refuse"))
		assert.Equal(t, mockOutput, c.Exec("ac"))
		assert.Equal(t, mockOutput, c.Exec("accept"))
		m.AssertCalled(t, "Refuse")
		m.AssertCalled(t, "AcceptSwap")
	})

	t.Run("nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewCuckooCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("nr"))
		assert.Equal(t, mockOutput, c.Exec("nextround"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewCuckooCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultCuckooConfig()
		expected.CpuDifficulty = domain.CuckooCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty errors", func(t *testing.T) {
		c := controller.NewCuckooCuiController(newMock())
		assert.Contains(t, c.Exec("sd"), msgCpuDifficultyRequired())
		assert.Contains(t, c.Exec("sd abc"), msgInvalidCpuDifficultyPrefix())
		assert.Contains(t, c.Exec("sd 9"), msgInvalidCpuDifficultyPrefix())
	})

	t.Run("setlives valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewCuckooCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sv 5"))
		expected := domain.DefaultCuckooConfig()
		expected.InitialLives = 5
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setlives errors", func(t *testing.T) {
		c := controller.NewCuckooCuiController(newMock())
		assert.Contains(t, c.Exec("sv"), "required")
		assert.Contains(t, c.Exec("sv abc"), "Invalid lives")
		assert.Contains(t, c.Exec("sv 0"), "Invalid lives")
	})

	t.Run("log", func(t *testing.T) {
		m := newMock()
		c := controller.NewCuckooCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("log"))
		assert.Equal(t, mockOutput, c.Exec("l"))
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewCuckooCuiController(newMock())
		assert.Contains(t, c.Exec("unknown"), "コマンドが不明です")
	})
}
