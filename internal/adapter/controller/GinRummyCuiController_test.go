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

func TestGinRummyCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockGinRummyInteractor {
		m := new(mockUsecases.MockGinRummyInteractor)
		m.On("GetConfig").Return(domain.DefaultGinRummyConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("DrawFromStock").Return(mockOutput)
		m.On("DrawFromDiscard").Return(mockOutput)
		m.On("Discard", mock.Anything).Return(mockOutput)
		m.On("Knock", mock.Anything).Return(mockOutput)
		m.On("Layoff", mock.Anything).Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	// quit
	t.Run("quit command q", func(t *testing.T) {
		c := controller.NewGinRummyCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("quit command quit", func(t *testing.T) {
		c := controller.NewGinRummyCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	// reset
	t.Run("reset command r preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewGinRummyCuiController(m)
		result := c.Exec("r")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultGinRummyConfig())
	})

	t.Run("reset command reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewGinRummyCuiController(m)
		result := c.Exec("reset")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultGinRummyConfig())
	})

	// drawstock
	t.Run("drawstock command ds", func(t *testing.T) {
		m := newMock()
		c := controller.NewGinRummyCuiController(m)
		result := c.Exec("ds")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "DrawFromStock")
	})

	t.Run("drawstock command drawstock", func(t *testing.T) {
		m := newMock()
		c := controller.NewGinRummyCuiController(m)
		result := c.Exec("drawstock")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "DrawFromStock")
	})

	// drawdiscard
	t.Run("drawdiscard command dd", func(t *testing.T) {
		m := newMock()
		c := controller.NewGinRummyCuiController(m)
		result := c.Exec("dd")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "DrawFromDiscard")
	})

	t.Run("drawdiscard command drawdiscard", func(t *testing.T) {
		m := newMock()
		c := controller.NewGinRummyCuiController(m)
		result := c.Exec("drawdiscard")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "DrawFromDiscard")
	})

	// discard
	t.Run("discard command d with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewGinRummyCuiController(m)
		result := c.Exec("d 3")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Discard", 3)
	})

	t.Run("discard command discard with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewGinRummyCuiController(m)
		result := c.Exec("discard 5")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Discard", 5)
	})

	t.Run("discard command d no args", func(t *testing.T) {
		c := controller.NewGinRummyCuiController(newMock())
		result := c.Exec("d")
		assert.Contains(t, result, msgCardIndexRequired())
	})

	t.Run("discard command d invalid arg", func(t *testing.T) {
		c := controller.NewGinRummyCuiController(newMock())
		result := c.Exec("d abc")
		assert.Contains(t, result, msgInvalidCardIndexPrefix())
	})

	// knock
	t.Run("knock command k with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewGinRummyCuiController(m)
		result := c.Exec("k 2")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Knock", 2)
	})

	t.Run("knock command knock with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewGinRummyCuiController(m)
		result := c.Exec("knock 7")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Knock", 7)
	})

	t.Run("knock command k no args", func(t *testing.T) {
		c := controller.NewGinRummyCuiController(newMock())
		result := c.Exec("k")
		assert.Contains(t, result, msgCardIndexRequired())
	})

	t.Run("knock command k invalid arg", func(t *testing.T) {
		c := controller.NewGinRummyCuiController(newMock())
		result := c.Exec("k abc")
		assert.Contains(t, result, msgInvalidCardIndexPrefix())
	})

	// layoff
	t.Run("layoff command lo with indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewGinRummyCuiController(m)
		result := c.Exec("lo 1,2,3")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Layoff", []int{1, 2, 3})
	})

	t.Run("layoff command layoff no args", func(t *testing.T) {
		m := newMock()
		c := controller.NewGinRummyCuiController(m)
		result := c.Exec("layoff")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Layoff", ([]int)(nil))
	})

	t.Run("layoff command lo no args", func(t *testing.T) {
		m := newMock()
		c := controller.NewGinRummyCuiController(m)
		result := c.Exec("lo")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Layoff", ([]int)(nil))
	})

	t.Run("layoff command lo with trailing comma", func(t *testing.T) {
		m := newMock()
		m.On("Layoff", []int{1}).Return(mockOutput)
		c := controller.NewGinRummyCuiController(m)
		result := c.Exec("lo 1,")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Layoff", []int{1})
	})

	t.Run("layoff command lo with invalid values ignored", func(t *testing.T) {
		m := newMock()
		m.On("Layoff", []int{1}).Return(mockOutput)
		c := controller.NewGinRummyCuiController(m)
		result := c.Exec("lo 1,abc")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Layoff", []int{1})
	})

	// nextround
	t.Run("nextround command nr", func(t *testing.T) {
		m := newMock()
		c := controller.NewGinRummyCuiController(m)
		result := c.Exec("nr")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "NextRound")
	})

	t.Run("nextround command nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewGinRummyCuiController(m)
		result := c.Exec("nextround")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "NextRound")
	})

	// setdifficulty
	t.Run("setdifficulty sd valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewGinRummyCuiController(m)
		result := c.Exec("sd 2")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultGinRummyConfig()
		expected.CpuDifficulty = domain.GinRummyCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty long form", func(t *testing.T) {
		m := newMock()
		c := controller.NewGinRummyCuiController(m)
		result := c.Exec("setdifficulty 0")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultGinRummyConfig()
		expected.CpuDifficulty = domain.GinRummyCpuDifficultyEasy
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty no args", func(t *testing.T) {
		c := controller.NewGinRummyCuiController(newMock())
		result := c.Exec("sd")
		assert.Contains(t, result, msgCpuDifficultyRequired())
	})

	t.Run("setdifficulty invalid value", func(t *testing.T) {
		c := controller.NewGinRummyCuiController(newMock())
		result := c.Exec("sd abc")
		assert.Contains(t, result, msgInvalidCpuDifficultyPrefix())
	})

	t.Run("setdifficulty negative", func(t *testing.T) {
		c := controller.NewGinRummyCuiController(newMock())
		result := c.Exec("sd -1")
		assert.Equal(t, msgInvalidCpuDifficulty("-1"), result)
	})

	t.Run("setdifficulty over 2", func(t *testing.T) {
		c := controller.NewGinRummyCuiController(newMock())
		result := c.Exec("sd 3")
		assert.Equal(t, msgInvalidCpuDifficulty("3"), result)
	})

	// setlimit
	t.Run("setlimit sl valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewGinRummyCuiController(m)
		result := c.Exec("sl 300")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultGinRummyConfig()
		expected.PointLimit = 300
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setlimit long form", func(t *testing.T) {
		m := newMock()
		c := controller.NewGinRummyCuiController(m)
		result := c.Exec("setlimit 200")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultGinRummyConfig()
		expected.PointLimit = 200
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setlimit no args", func(t *testing.T) {
		c := controller.NewGinRummyCuiController(newMock())
		result := c.Exec("sl")
		assert.Contains(t, result, "required")
	})

	t.Run("setlimit invalid value", func(t *testing.T) {
		c := controller.NewGinRummyCuiController(newMock())
		result := c.Exec("sl abc")
		assert.Contains(t, result, "Invalid point limit")
	})

	t.Run("setlimit zero", func(t *testing.T) {
		c := controller.NewGinRummyCuiController(newMock())
		result := c.Exec("sl 0")
		assert.Equal(t, "Invalid point limit: 0. Please enter 1 or more.", result)
	})

	t.Run("setlimit negative", func(t *testing.T) {
		c := controller.NewGinRummyCuiController(newMock())
		result := c.Exec("sl -1")
		assert.Equal(t, "Invalid point limit: -1. Please enter 1 or more.", result)
	})

	// log
	t.Run("log command", func(t *testing.T) {
		m := newMock()
		c := controller.NewGinRummyCuiController(m)
		result := c.Exec("log")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("l command", func(t *testing.T) {
		m := newMock()
		c := controller.NewGinRummyCuiController(m)
		result := c.Exec("l")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ActionLog")
	})

	// unknown / empty
	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewGinRummyCuiController(newMock())
		result := c.Exec("unknown")
		assert.Contains(t, result, "コマンドが不明です")
	})

	t.Run("empty command", func(t *testing.T) {
		c := controller.NewGinRummyCuiController(newMock())
		result := c.Exec("")
		assert.Contains(t, result, "'help' でコマンド一覧を表示します。")
	})
}
