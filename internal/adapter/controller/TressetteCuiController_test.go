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

func TestTressetteCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockTressetteInteractor {
		m := new(mockUsecases.MockTressetteInteractor)
		m.On("GetConfig").Return(domain.DefaultTressetteConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewTressetteCuiController(newMock()).Exec("q"))
		assert.Equal(t, "bye.", controller.NewTressetteCuiController(newMock()).Exec("quit"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewTressetteCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultTressetteConfig())
	})

	t.Run("play", func(t *testing.T) {
		m := newMock()
		c := controller.NewTressetteCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 3"))
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("play no args", func(t *testing.T) {
		result := controller.NewTressetteCuiController(newMock()).Exec("p")
		assert.Contains(t, result, msgCardIndexRequired())
	})

	t.Run("next / nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewTressetteCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextTrick")
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewTressetteCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultTressetteConfig()
		expected.CpuDifficulty = domain.TressetteCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty invalid", func(t *testing.T) {
		result := controller.NewTressetteCuiController(newMock()).Exec("sd 9")
		assert.Contains(t, result, msgInvalidCpuDifficultyPrefix())
	})

	t.Run("settarget", func(t *testing.T) {
		m := newMock()
		c := controller.NewTressetteCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("st 31"))
		expected := domain.DefaultTressetteConfig()
		expected.TargetPoints = 31
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("settarget invalid", func(t *testing.T) {
		result := controller.NewTressetteCuiController(newMock()).Exec("st 0")
		assert.Contains(t, result, "Invalid target points")
	})

	t.Run("hint / log", func(t *testing.T) {
		m := newMock()
		c := controller.NewTressetteCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "Hint")
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		result := controller.NewTressetteCuiController(newMock()).Exec("zzz")
		assert.Contains(t, result, "コマンドが不明です")
	})
}
