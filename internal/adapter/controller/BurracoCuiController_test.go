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

func TestBurracoCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockBurracoInteractor {
		m := new(mockUsecases.MockBurracoInteractor)
		m.On("GetConfig").Return(domain.DefaultBurracoConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("DrawFromStock").Return(mockOutput)
		m.On("DrawFromDiscard", mock.Anything).Return(mockOutput)
		m.On("Meld", mock.Anything).Return(mockOutput)
		m.On("SkipMeld").Return(mockOutput)
		m.On("Discard", mock.Anything).Return(mockOutput)
		m.On("GoOut").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	// quit
	t.Run("quit command q", func(t *testing.T) {
		c := controller.NewBurracoCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("quit command quit", func(t *testing.T) {
		c := controller.NewBurracoCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	// reset
	t.Run("reset command r preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewBurracoCuiController(m)
		result := c.Exec("r")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultBurracoConfig())
	})

	t.Run("reset command reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewBurracoCuiController(m)
		result := c.Exec("reset")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultBurracoConfig())
	})

	// drawstock
	t.Run("drawstock command ds", func(t *testing.T) {
		m := newMock()
		c := controller.NewBurracoCuiController(m)
		result := c.Exec("ds")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "DrawFromStock")
	})

	t.Run("drawstock command drawstock", func(t *testing.T) {
		m := newMock()
		c := controller.NewBurracoCuiController(m)
		result := c.Exec("drawstock")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "DrawFromStock")
	})

	// drawdiscard
	t.Run("drawdiscard command dd no args", func(t *testing.T) {
		m := newMock()
		c := controller.NewBurracoCuiController(m)
		result := c.Exec("dd")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "DrawFromDiscard", ([]int)(nil))
	})

	t.Run("drawdiscard command dd with indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewBurracoCuiController(m)
		result := c.Exec("dd 0,1")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "DrawFromDiscard", []int{0, 1})
	})

	t.Run("drawdiscard command drawdiscard", func(t *testing.T) {
		m := newMock()
		c := controller.NewBurracoCuiController(m)
		result := c.Exec("drawdiscard 2,3")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "DrawFromDiscard", []int{2, 3})
	})

	// meld
	t.Run("meld command m no args", func(t *testing.T) {
		m := newMock()
		c := controller.NewBurracoCuiController(m)
		result := c.Exec("m")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Meld", ([][]int)(nil))
	})

	t.Run("meld command m with groups", func(t *testing.T) {
		m := newMock()
		c := controller.NewBurracoCuiController(m)
		result := c.Exec("m 0,1,2;3,4,5")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Meld", [][]int{{0, 1, 2}, {3, 4, 5}})
	})

	t.Run("meld command meld", func(t *testing.T) {
		m := newMock()
		c := controller.NewBurracoCuiController(m)
		result := c.Exec("meld 0,1,2")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Meld", [][]int{{0, 1, 2}})
	})

	// skipmeld
	t.Run("skipmeld command sm", func(t *testing.T) {
		m := newMock()
		c := controller.NewBurracoCuiController(m)
		result := c.Exec("sm")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "SkipMeld")
	})

	t.Run("skipmeld command skipmeld", func(t *testing.T) {
		m := newMock()
		c := controller.NewBurracoCuiController(m)
		result := c.Exec("skipmeld")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "SkipMeld")
	})

	// discard
	t.Run("discard command d with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewBurracoCuiController(m)
		result := c.Exec("d 3")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Discard", 3)
	})

	t.Run("discard command discard with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewBurracoCuiController(m)
		result := c.Exec("discard 5")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Discard", 5)
	})

	t.Run("discard command d no args", func(t *testing.T) {
		c := controller.NewBurracoCuiController(newMock())
		result := c.Exec("d")
		assert.Contains(t, result, msgCardIndexRequired())
	})

	t.Run("discard command d invalid arg", func(t *testing.T) {
		c := controller.NewBurracoCuiController(newMock())
		result := c.Exec("d abc")
		assert.Contains(t, result, msgInvalidCardIndexPrefix())
	})

	// goout
	t.Run("goout command go", func(t *testing.T) {
		m := newMock()
		c := controller.NewBurracoCuiController(m)
		result := c.Exec("go")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "GoOut")
	})

	t.Run("goout command goout", func(t *testing.T) {
		m := newMock()
		c := controller.NewBurracoCuiController(m)
		result := c.Exec("goout")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "GoOut")
	})

	// nextround
	t.Run("nextround command nr", func(t *testing.T) {
		m := newMock()
		c := controller.NewBurracoCuiController(m)
		result := c.Exec("nr")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "NextRound")
	})

	t.Run("nextround command nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewBurracoCuiController(m)
		result := c.Exec("nextround")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "NextRound")
	})

	// setdifficulty
	t.Run("setdifficulty sd valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewBurracoCuiController(m)
		result := c.Exec("sd 2")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultBurracoConfig()
		expected.CpuDifficulty = domain.BurracoCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty long form", func(t *testing.T) {
		m := newMock()
		c := controller.NewBurracoCuiController(m)
		result := c.Exec("setdifficulty 0")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultBurracoConfig()
		expected.CpuDifficulty = domain.BurracoCpuDifficultyEasy
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty no args", func(t *testing.T) {
		c := controller.NewBurracoCuiController(newMock())
		result := c.Exec("sd")
		assert.Contains(t, result, msgCpuDifficultyRequired())
	})

	t.Run("setdifficulty invalid value", func(t *testing.T) {
		c := controller.NewBurracoCuiController(newMock())
		result := c.Exec("sd abc")
		assert.Contains(t, result, msgInvalidCpuDifficultyPrefix())
	})

	t.Run("setdifficulty negative", func(t *testing.T) {
		c := controller.NewBurracoCuiController(newMock())
		result := c.Exec("sd -1")
		assert.Contains(t, result, msgInvalidCpuDifficultyPrefix())
	})

	// setlimit
	t.Run("setlimit sl valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewBurracoCuiController(m)
		result := c.Exec("sl 7500")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultBurracoConfig()
		expected.PointLimit = 7500
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setlimit long form", func(t *testing.T) {
		m := newMock()
		c := controller.NewBurracoCuiController(m)
		result := c.Exec("setlimit 3000")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultBurracoConfig()
		expected.PointLimit = 3000
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setlimit no args", func(t *testing.T) {
		c := controller.NewBurracoCuiController(newMock())
		result := c.Exec("sl")
		assert.Contains(t, result, "required")
	})

	t.Run("setlimit invalid value", func(t *testing.T) {
		c := controller.NewBurracoCuiController(newMock())
		result := c.Exec("sl abc")
		assert.Contains(t, result, "Invalid point limit")
	})

	t.Run("setlimit zero is invalid", func(t *testing.T) {
		c := controller.NewBurracoCuiController(newMock())
		result := c.Exec("sl 0")
		assert.Contains(t, result, "Invalid point limit")
	})

	// log
	t.Run("log command log", func(t *testing.T) {
		m := newMock()
		c := controller.NewBurracoCuiController(m)
		result := c.Exec("log")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("log command l", func(t *testing.T) {
		m := newMock()
		c := controller.NewBurracoCuiController(m)
		result := c.Exec("l")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("hint command h", func(t *testing.T) {
		m := newMock()
		c := controller.NewBurracoCuiController(m)
		result := c.Exec("h")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Hint")
	})

	t.Run("hint command hint", func(t *testing.T) {
		m := newMock()
		c := controller.NewBurracoCuiController(m)
		result := c.Exec("hint")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Hint")
	})

	// unknown
	t.Run("unknown command returns unsupported", func(t *testing.T) {
		c := controller.NewBurracoCuiController(newMock())
		result := c.Exec("unknown")
		assert.Contains(t, result, "コマンドが不明です")
	})

	// empty command
	t.Run("empty command", func(t *testing.T) {
		c := controller.NewBurracoCuiController(newMock())
		result := c.Exec("")
		assert.Contains(t, result, "'help' でコマンド一覧を表示します。")
	})
}
