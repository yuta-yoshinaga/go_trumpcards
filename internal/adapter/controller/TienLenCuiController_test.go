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

func TestTienLenCuiController_Exec(t *testing.T) {
	mockOutput := `{"players":[]}`

	newMock := func() *mockUsecases.MockTienLenInteractor {
		m := new(mockUsecases.MockTienLenInteractor)
		m.On("GetConfig").Return(domain.DefaultTienLenConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		c := controller.NewTienLenCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	t.Run("reset", func(t *testing.T) {
		m := newMock()
		c := controller.NewTienLenCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		assert.Equal(t, mockOutput, c.Exec("reset"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultTienLenConfig())
	})

	t.Run("play with indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewTienLenCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 0 1"))
		assert.Equal(t, mockOutput, c.Exec("play 2"))
		m.AssertCalled(t, "Play", []int{0, 1})
		m.AssertCalled(t, "Play", []int{2})
	})

	t.Run("play with no indices is a pass", func(t *testing.T) {
		m := newMock()
		c := controller.NewTienLenCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p"))
		m.AssertCalled(t, "Play", []int{})
	})

	t.Run("setdifficulty valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewTienLenCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultTienLenConfig()
		expected.CpuDifficulty = domain.TienLenDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty errors", func(t *testing.T) {
		c := controller.NewTienLenCuiController(newMock())
		assert.Contains(t, c.Exec("sd"), msgCpuDifficultyRequiredAlt())
		assert.Contains(t, c.Exec("sd abc"), msgInvalidCpuDifficultyPrefix())
		assert.Contains(t, c.Exec("sd 9"), msgInvalidCpuDifficultyPrefix())
	})

	t.Run("log", func(t *testing.T) {
		m := newMock()
		c := controller.NewTienLenCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("log"))
		assert.Equal(t, mockOutput, c.Exec("l"))
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewTienLenCuiController(newMock())
		assert.Contains(t, c.Exec("unknown"), "コマンドが不明です")
	})

	t.Run("empty command", func(t *testing.T) {
		c := controller.NewTienLenCuiController(newMock())
		assert.Contains(t, c.Exec(""), "'help' でコマンド一覧を表示します。")
	})
}
