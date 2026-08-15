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

func TestGaigelCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockGaigelInteractor {
		m := new(mockUsecases.MockGaigelInteractor)
		m.On("GetConfig").Return(domain.DefaultGaigelConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("DeclareMarriage", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit q", func(t *testing.T) {
		c := controller.NewGaigelCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("reset r preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewGaigelCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "GetConfig")
	})

	t.Run("play p with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewGaigelCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 3"))
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("play no args", func(t *testing.T) {
		c := controller.NewGaigelCuiController(newMock())
		assert.Contains(t, c.Exec("p"), msgCardIndexRequired())
	})

	t.Run("marriage m with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewGaigelCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("m 2"))
		m.AssertCalled(t, "DeclareMarriage", 2)
	})

	t.Run("marriage long form", func(t *testing.T) {
		m := newMock()
		c := controller.NewGaigelCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("marriage 0"))
		m.AssertCalled(t, "DeclareMarriage", 0)
	})

	t.Run("marriage no args", func(t *testing.T) {
		c := controller.NewGaigelCuiController(newMock())
		assert.Contains(t, c.Exec("m"), msgCardIndexRequired())
	})

	t.Run("next n", func(t *testing.T) {
		m := newMock()
		c := controller.NewGaigelCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		m.AssertCalled(t, "NextTrick")
	})

	t.Run("nextround nr", func(t *testing.T) {
		m := newMock()
		c := controller.NewGaigelCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewGaigelCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultGaigelConfig()
		expected.CpuDifficulty = domain.GaigelCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("settarget valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewGaigelCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("st 51"))
		expected := domain.DefaultGaigelConfig()
		expected.TargetScore = 51
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("settarget invalid", func(t *testing.T) {
		c := controller.NewGaigelCuiController(newMock())
		assert.Contains(t, c.Exec("st 0"), "Invalid target score: 0")
	})

	t.Run("hint h", func(t *testing.T) {
		m := newMock()
		c := controller.NewGaigelCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		m.AssertCalled(t, "Hint")
	})

	t.Run("log", func(t *testing.T) {
		m := newMock()
		c := controller.NewGaigelCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewGaigelCuiController(newMock())
		assert.Contains(t, c.Exec("unknown"), "コマンドが不明")
	})

	t.Run("empty command", func(t *testing.T) {
		c := controller.NewGaigelCuiController(newMock())
		assert.Contains(t, c.Exec(""), "'help'")
	})
}
