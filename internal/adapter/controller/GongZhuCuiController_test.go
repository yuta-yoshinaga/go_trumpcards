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

func TestGongZhuCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockGongZhuInteractor {
		m := new(mockUsecases.MockGongZhuInteractor)
		m.On("GetConfig").Return(domain.DefaultGongZhuConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Expose", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewGongZhuCuiController(newMock()).Exec("q"))
		assert.Equal(t, "bye.", controller.NewGongZhuCuiController(newMock()).Exec("quit"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewGongZhuCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultGongZhuConfig())
	})

	t.Run("expose with indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewGongZhuCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("expose 0 1"))
		m.AssertCalled(t, "Expose", []int{0, 1})
	})

	t.Run("expose with no indices = expose nothing", func(t *testing.T) {
		m := newMock()
		c := controller.NewGongZhuCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("expose"))
		m.AssertCalled(t, "Expose", []int{})
	})

	t.Run("expose with invalid arg shows warning", func(t *testing.T) {
		m := newMock()
		c := controller.NewGongZhuCuiController(m)
		result := c.Exec("expose 0 x 1")
		assert.Contains(t, result, "'x'")
		m.AssertCalled(t, "Expose", []int{0, 1})
	})

	t.Run("play", func(t *testing.T) {
		m := newMock()
		c := controller.NewGongZhuCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 3"))
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("play no args", func(t *testing.T) {
		result := controller.NewGongZhuCuiController(newMock()).Exec("p")
		assert.Contains(t, result, msgCardIndexRequired())
	})

	t.Run("next / nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewGongZhuCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextTrick")
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewGongZhuCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultGongZhuConfig()
		expected.CpuDifficulty = domain.GongZhuCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty invalid", func(t *testing.T) {
		result := controller.NewGongZhuCuiController(newMock()).Exec("sd 9")
		assert.Contains(t, result, "Invalid CPU difficulty")
	})

	t.Run("setlimit", func(t *testing.T) {
		m := newMock()
		c := controller.NewGongZhuCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sl 500"))
		expected := domain.DefaultGongZhuConfig()
		expected.PointLimit = 500
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setlimit invalid", func(t *testing.T) {
		result := controller.NewGongZhuCuiController(newMock()).Exec("sl 0")
		assert.Contains(t, result, "Invalid point limit")
	})

	t.Run("hint / log", func(t *testing.T) {
		m := newMock()
		c := controller.NewGongZhuCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "Hint")
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		result := controller.NewGongZhuCuiController(newMock()).Exec("zzz")
		assert.Contains(t, result, "コマンドが不明です")
	})
}
