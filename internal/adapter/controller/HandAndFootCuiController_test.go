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

func TestHandAndFootCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockHandAndFootInteractor {
		m := new(mockUsecases.MockHandAndFootInteractor)
		m.On("GetConfig").Return(domain.DefaultHandAndFootConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("DrawFromStock").Return(mockOutput)
		m.On("DrawFromDiscard", mock.Anything).Return(mockOutput)
		m.On("Meld", mock.Anything).Return(mockOutput)
		m.On("SkipMeld").Return(mockOutput)
		m.On("Discard", mock.Anything).Return(mockOutput)
		m.On("GoOut").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		return m
	}

	t.Run("quit command q", func(t *testing.T) {
		c := controller.NewHandAndFootCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("reset command r preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewHandAndFootCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultHandAndFootConfig())
	})

	t.Run("drawstock ds", func(t *testing.T) {
		m := newMock()
		c := controller.NewHandAndFootCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("ds"))
		m.AssertCalled(t, "DrawFromStock")
	})

	t.Run("hint h", func(t *testing.T) {
		m := newMock()
		c := controller.NewHandAndFootCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		assert.Equal(t, mockOutput, c.Exec("hint"))
		m.AssertCalled(t, "Hint")
	})

	t.Run("drawdiscard dd with indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewHandAndFootCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("dd 0,1"))
		m.AssertCalled(t, "DrawFromDiscard", []int{0, 1})
	})

	t.Run("meld m with groups", func(t *testing.T) {
		m := newMock()
		c := controller.NewHandAndFootCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("m 0,1,2;3,4,5"))
		m.AssertCalled(t, "Meld", [][]int{{0, 1, 2}, {3, 4, 5}})
	})

	t.Run("skipmeld sm", func(t *testing.T) {
		m := newMock()
		c := controller.NewHandAndFootCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sm"))
		m.AssertCalled(t, "SkipMeld")
	})

	t.Run("discard d with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewHandAndFootCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("d 3"))
		m.AssertCalled(t, "Discard", 3)
	})

	t.Run("discard d no args", func(t *testing.T) {
		c := controller.NewHandAndFootCuiController(newMock())
		assert.Contains(t, c.Exec("d"), "Card index is required")
	})

	t.Run("discard d invalid arg", func(t *testing.T) {
		c := controller.NewHandAndFootCuiController(newMock())
		assert.Contains(t, c.Exec("d abc"), "Invalid card index")
	})

	t.Run("goout go", func(t *testing.T) {
		m := newMock()
		c := controller.NewHandAndFootCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("go"))
		m.AssertCalled(t, "GoOut")
	})

	t.Run("nextround nr", func(t *testing.T) {
		m := newMock()
		c := controller.NewHandAndFootCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty sd valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewHandAndFootCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultHandAndFootConfig()
		expected.CpuDifficulty = domain.HandAndFootCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty invalid", func(t *testing.T) {
		c := controller.NewHandAndFootCuiController(newMock())
		assert.Contains(t, c.Exec("sd abc"), "Invalid CPU difficulty")
	})

	t.Run("setlimit sl valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewHandAndFootCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sl 7500"))
		expected := domain.DefaultHandAndFootConfig()
		expected.PointLimit = 7500
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setlimit zero invalid", func(t *testing.T) {
		c := controller.NewHandAndFootCuiController(newMock())
		assert.Contains(t, c.Exec("sl 0"), "Invalid point limit")
	})

	t.Run("log l", func(t *testing.T) {
		m := newMock()
		c := controller.NewHandAndFootCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("l"))
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewHandAndFootCuiController(newMock())
		assert.Contains(t, c.Exec("unknown"), "コマンドが不明です")
	})

	t.Run("empty command", func(t *testing.T) {
		c := controller.NewHandAndFootCuiController(newMock())
		assert.Contains(t, c.Exec(""), "'help' でコマンド一覧を表示します。")
	})
}
