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

func TestScartoCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":1}`

	newMock := func() *mockUsecases.MockScartoInteractor {
		m := new(mockUsecases.MockScartoInteractor)
		m.On("GetConfig").Return(domain.DefaultScartoConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Discard", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewScartoCuiController(newMock()).Exec("q"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewScartoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultScartoConfig())
	})

	t.Run("scarto three cards", func(t *testing.T) {
		m := newMock()
		c := controller.NewScartoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("scarto 0 1 2"))
		m.AssertCalled(t, "Discard", []int{0, 1, 2})
	})

	t.Run("discard alias", func(t *testing.T) {
		m := newMock()
		c := controller.NewScartoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("discard 3 4 5"))
		m.AssertCalled(t, "Discard", []int{3, 4, 5})
	})

	t.Run("scarto too few", func(t *testing.T) {
		result := controller.NewScartoCuiController(newMock()).Exec("scarto 0 1")
		assert.Contains(t, result, "Three card indices are required")
	})

	t.Run("scarto invalid index", func(t *testing.T) {
		result := controller.NewScartoCuiController(newMock()).Exec("scarto 0 1 x")
		assert.Contains(t, result, msgInvalidCardIndexPrefix())
	})

	t.Run("play card", func(t *testing.T) {
		m := newMock()
		c := controller.NewScartoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("play 3"))
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("play no args", func(t *testing.T) {
		result := controller.NewScartoCuiController(newMock()).Exec("play")
		assert.Contains(t, result, msgCardIndexRequired())
	})

	t.Run("next / nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewScartoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextTrick")
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewScartoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultScartoConfig()
		expected.CpuDifficulty = domain.ScartoCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty invalid", func(t *testing.T) {
		result := controller.NewScartoCuiController(newMock()).Exec("sd 9")
		assert.Contains(t, result, msgInvalidCpuDifficultyPrefix())
	})

	t.Run("hint / log", func(t *testing.T) {
		m := newMock()
		c := controller.NewScartoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "Hint")
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		result := controller.NewScartoCuiController(newMock()).Exec("zzz")
		assert.Contains(t, result, "コマンドが不明です")
	})
}
