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

func TestZhengCuiController_Exec(t *testing.T) {
	mockOutput := `{"players":[]}`

	newMock := func() *mockUsecases.MockZhengInteractor {
		m := new(mockUsecases.MockZhengInteractor)
		m.On("GetConfig").Return(domain.DefaultZhengConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		c := controller.NewZhengCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	t.Run("reset", func(t *testing.T) {
		m := newMock()
		c := controller.NewZhengCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		assert.Equal(t, mockOutput, c.Exec("reset"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultZhengConfig())
	})

	t.Run("play with indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewZhengCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 0 1"))
		assert.Equal(t, mockOutput, c.Exec("play 2"))
		m.AssertCalled(t, "Play", []int{0, 1})
		m.AssertCalled(t, "Play", []int{2})
	})

	t.Run("play with no indices is a pass", func(t *testing.T) {
		m := newMock()
		c := controller.NewZhengCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p"))
		m.AssertCalled(t, "Play", []int{})
	})

	t.Run("setdifficulty valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewZhengCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultZhengConfig()
		expected.CpuDifficulty = domain.ZhengDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty errors", func(t *testing.T) {
		c := controller.NewZhengCuiController(newMock())
		assert.Contains(t, c.Exec("sd"), "required")
		assert.Contains(t, c.Exec("sd abc"), "Invalid CPU difficulty")
		assert.Contains(t, c.Exec("sd 9"), "Invalid CPU difficulty")
	})

	t.Run("log", func(t *testing.T) {
		m := newMock()
		c := controller.NewZhengCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("log"))
		assert.Equal(t, mockOutput, c.Exec("l"))
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewZhengCuiController(newMock())
		assert.Contains(t, c.Exec("unknown"), "コマンドが不明です")
	})

	t.Run("empty command", func(t *testing.T) {
		c := controller.NewZhengCuiController(newMock())
		assert.Contains(t, c.Exec(""), "'help' でコマンド一覧を表示します。")
	})
}
