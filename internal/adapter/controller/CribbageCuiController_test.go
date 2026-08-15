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

func TestCribbageCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockCribbageInteractor {
		m := new(mockUsecases.MockCribbageInteractor)
		m.On("GetConfig").Return(domain.DefaultCribbageConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Discard", mock.Anything).Return(mockOutput)
		m.On("Cut").Return(mockOutput)
		m.On("Peg", mock.Anything).Return(mockOutput)
		m.On("Go").Return(mockOutput)
		m.On("ShowNext").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	// quit
	t.Run("quit command q", func(t *testing.T) {
		c := controller.NewCribbageCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("quit command quit", func(t *testing.T) {
		c := controller.NewCribbageCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	// reset
	t.Run("reset command r preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewCribbageCuiController(m)
		result := c.Exec("r")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultCribbageConfig())
	})

	t.Run("reset command reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewCribbageCuiController(m)
		result := c.Exec("reset")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultCribbageConfig())
	})

	// discard
	t.Run("discard command d with indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewCribbageCuiController(m)
		result := c.Exec("d 1,3")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Discard", []int{1, 3})
	})

	t.Run("discard command discard with indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewCribbageCuiController(m)
		result := c.Exec("discard 0,5")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Discard", []int{0, 5})
	})

	t.Run("discard command d no args", func(t *testing.T) {
		m := newMock()
		c := controller.NewCribbageCuiController(m)
		result := c.Exec("d")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Discard", ([]int)(nil))
	})

	t.Run("discard command d with trailing comma", func(t *testing.T) {
		m := newMock()
		m.On("Discard", []int{1}).Return(mockOutput)
		c := controller.NewCribbageCuiController(m)
		result := c.Exec("d 1,")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Discard", []int{1})
	})

	t.Run("discard command d with invalid values ignored", func(t *testing.T) {
		m := newMock()
		m.On("Discard", []int{1}).Return(mockOutput)
		c := controller.NewCribbageCuiController(m)
		result := c.Exec("d 1,abc")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Discard", []int{1})
	})

	// cut
	t.Run("cut command c", func(t *testing.T) {
		m := newMock()
		c := controller.NewCribbageCuiController(m)
		result := c.Exec("c")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Cut")
	})

	t.Run("cut command cut", func(t *testing.T) {
		m := newMock()
		c := controller.NewCribbageCuiController(m)
		result := c.Exec("cut")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Cut")
	})

	// peg
	t.Run("peg command p with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewCribbageCuiController(m)
		result := c.Exec("p 2")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Peg", 2)
	})

	t.Run("peg command peg with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewCribbageCuiController(m)
		result := c.Exec("peg 4")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Peg", 4)
	})

	t.Run("peg command p no args", func(t *testing.T) {
		c := controller.NewCribbageCuiController(newMock())
		result := c.Exec("p")
		assert.Contains(t, result, msgCardIndexRequired())
	})

	t.Run("peg command p invalid arg", func(t *testing.T) {
		c := controller.NewCribbageCuiController(newMock())
		result := c.Exec("p abc")
		assert.Contains(t, result, msgInvalidCardIndexPrefix())
	})

	// go
	t.Run("go command", func(t *testing.T) {
		m := newMock()
		c := controller.NewCribbageCuiController(m)
		result := c.Exec("go")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Go")
	})

	// shownext
	t.Run("shownext command sn", func(t *testing.T) {
		m := newMock()
		c := controller.NewCribbageCuiController(m)
		result := c.Exec("sn")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ShowNext")
	})

	t.Run("shownext command shownext", func(t *testing.T) {
		m := newMock()
		c := controller.NewCribbageCuiController(m)
		result := c.Exec("shownext")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ShowNext")
	})

	// nextround
	t.Run("nextround command nr", func(t *testing.T) {
		m := newMock()
		c := controller.NewCribbageCuiController(m)
		result := c.Exec("nr")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "NextRound")
	})

	t.Run("nextround command nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewCribbageCuiController(m)
		result := c.Exec("nextround")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "NextRound")
	})

	// setdifficulty
	t.Run("setdifficulty sd valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewCribbageCuiController(m)
		result := c.Exec("sd 2")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultCribbageConfig()
		expected.CpuDifficulty = domain.CribbageCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty long form", func(t *testing.T) {
		m := newMock()
		c := controller.NewCribbageCuiController(m)
		result := c.Exec("setdifficulty 0")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultCribbageConfig()
		expected.CpuDifficulty = domain.CribbageCpuDifficultyEasy
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty no args", func(t *testing.T) {
		c := controller.NewCribbageCuiController(newMock())
		result := c.Exec("sd")
		assert.Contains(t, result, msgCpuDifficultyRequired())
	})

	t.Run("setdifficulty invalid value", func(t *testing.T) {
		c := controller.NewCribbageCuiController(newMock())
		result := c.Exec("sd abc")
		assert.Contains(t, result, msgInvalidCpuDifficultyPrefix())
	})

	t.Run("setdifficulty negative", func(t *testing.T) {
		c := controller.NewCribbageCuiController(newMock())
		result := c.Exec("sd -1")
		assert.Equal(t, msgInvalidCpuDifficulty("-1"), result)
	})

	t.Run("setdifficulty over 2", func(t *testing.T) {
		c := controller.NewCribbageCuiController(newMock())
		result := c.Exec("sd 3")
		assert.Equal(t, msgInvalidCpuDifficulty("3"), result)
	})

	// setlimit
	t.Run("setlimit sl valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewCribbageCuiController(m)
		result := c.Exec("sl 200")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultCribbageConfig()
		expected.PointLimit = 200
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setlimit long form", func(t *testing.T) {
		m := newMock()
		c := controller.NewCribbageCuiController(m)
		result := c.Exec("setlimit 300")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultCribbageConfig()
		expected.PointLimit = 300
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setlimit no args", func(t *testing.T) {
		c := controller.NewCribbageCuiController(newMock())
		result := c.Exec("sl")
		assert.Contains(t, result, "required")
	})

	t.Run("setlimit invalid value", func(t *testing.T) {
		c := controller.NewCribbageCuiController(newMock())
		result := c.Exec("sl abc")
		assert.Contains(t, result, "Invalid point limit")
	})

	t.Run("setlimit zero", func(t *testing.T) {
		c := controller.NewCribbageCuiController(newMock())
		result := c.Exec("sl 0")
		assert.Equal(t, "Invalid point limit: 0. Please enter 1 or more.", result)
	})

	t.Run("setlimit negative", func(t *testing.T) {
		c := controller.NewCribbageCuiController(newMock())
		result := c.Exec("sl -1")
		assert.Equal(t, "Invalid point limit: -1. Please enter 1 or more.", result)
	})

	// log
	t.Run("log command", func(t *testing.T) {
		m := newMock()
		c := controller.NewCribbageCuiController(m)
		result := c.Exec("log")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("l command", func(t *testing.T) {
		m := newMock()
		c := controller.NewCribbageCuiController(m)
		result := c.Exec("l")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ActionLog")
	})

	// unknown / empty
	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewCribbageCuiController(newMock())
		result := c.Exec("unknown")
		assert.Contains(t, result, "コマンドが不明です")
	})

	t.Run("empty command", func(t *testing.T) {
		c := controller.NewCribbageCuiController(newMock())
		result := c.Exec("")
		assert.Contains(t, result, "'help' でコマンド一覧を表示します。")
	})
}

func TestCribbageCuiController_Hint(t *testing.T) {
	mockOutput := "hint output"
	for _, cmd := range []string{"h", "hint"} {
		t.Run("hint command "+cmd, func(t *testing.T) {
			m := new(mockUsecases.MockCribbageInteractor)
			m.On("Hint").Return(mockOutput)
			c := controller.NewCribbageCuiController(m)
			assert.Equal(t, mockOutput, c.Exec(cmd))
			m.AssertCalled(t, "Hint")
		})
	}
}
