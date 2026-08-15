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

func TestSpadesCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockSpadesInteractor {
		m := new(mockUsecases.MockSpadesInteractor)
		m.On("GetConfig").Return(domain.DefaultSpadesConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Bid", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	// quit
	t.Run("quit command q", func(t *testing.T) {
		c := controller.NewSpadesCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("quit command quit", func(t *testing.T) {
		c := controller.NewSpadesCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	// reset
	t.Run("reset command r preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewSpadesCuiController(m)
		result := c.Exec("r")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultSpadesConfig())
	})

	t.Run("reset command reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewSpadesCuiController(m)
		result := c.Exec("reset")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultSpadesConfig())
	})

	// bid
	t.Run("bid command b with value", func(t *testing.T) {
		m := newMock()
		c := controller.NewSpadesCuiController(m)
		result := c.Exec("b 3")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Bid", 3)
	})

	t.Run("bid command bid with value", func(t *testing.T) {
		m := newMock()
		c := controller.NewSpadesCuiController(m)
		result := c.Exec("bid 5")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Bid", 5)
	})

	t.Run("bid command b no args", func(t *testing.T) {
		c := controller.NewSpadesCuiController(newMock())
		result := c.Exec("b")
		assert.Contains(t, result, "Bid value is required")
	})

	t.Run("bid command b invalid arg", func(t *testing.T) {
		c := controller.NewSpadesCuiController(newMock())
		result := c.Exec("b abc")
		assert.Contains(t, result, "Invalid bid value")
	})

	t.Run("bid command b out of range negative", func(t *testing.T) {
		c := controller.NewSpadesCuiController(newMock())
		result := c.Exec("b -1")
		assert.Contains(t, result, "Invalid bid value: -1")
	})

	t.Run("bid command b out of range over 13", func(t *testing.T) {
		c := controller.NewSpadesCuiController(newMock())
		result := c.Exec("b 14")
		assert.Contains(t, result, "Invalid bid value: 14")
	})

	// play
	t.Run("play command p with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewSpadesCuiController(m)
		result := c.Exec("p 3")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("play command play with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewSpadesCuiController(m)
		result := c.Exec("play 5")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", 5)
	})

	t.Run("play command p no args", func(t *testing.T) {
		c := controller.NewSpadesCuiController(newMock())
		result := c.Exec("p")
		assert.Contains(t, result, msgCardIndexRequired())
	})

	t.Run("play command p invalid arg", func(t *testing.T) {
		c := controller.NewSpadesCuiController(newMock())
		result := c.Exec("p abc")
		assert.Contains(t, result, msgInvalidCardIndexPrefix())
	})

	// next
	t.Run("next command n", func(t *testing.T) {
		m := newMock()
		c := controller.NewSpadesCuiController(m)
		result := c.Exec("n")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "NextTrick")
	})

	t.Run("next command next", func(t *testing.T) {
		m := newMock()
		c := controller.NewSpadesCuiController(m)
		result := c.Exec("next")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "NextTrick")
	})

	// nextround
	t.Run("nextround command nr", func(t *testing.T) {
		m := newMock()
		c := controller.NewSpadesCuiController(m)
		result := c.Exec("nr")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "NextRound")
	})

	t.Run("nextround command nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewSpadesCuiController(m)
		result := c.Exec("nextround")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "NextRound")
	})

	// setdifficulty
	t.Run("setdifficulty sd valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewSpadesCuiController(m)
		result := c.Exec("sd 2")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultSpadesConfig()
		expected.CpuDifficulty = domain.SpadesCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty long form", func(t *testing.T) {
		m := newMock()
		c := controller.NewSpadesCuiController(m)
		result := c.Exec("setdifficulty 0")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultSpadesConfig()
		expected.CpuDifficulty = domain.SpadesCpuDifficultyEasy
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty no args", func(t *testing.T) {
		c := controller.NewSpadesCuiController(newMock())
		result := c.Exec("sd")
		assert.Contains(t, result, "required")
	})

	t.Run("setdifficulty invalid value", func(t *testing.T) {
		c := controller.NewSpadesCuiController(newMock())
		result := c.Exec("sd abc")
		assert.Contains(t, result, "Invalid CPU difficulty")
	})

	t.Run("setdifficulty negative", func(t *testing.T) {
		c := controller.NewSpadesCuiController(newMock())
		result := c.Exec("sd -1")
		assert.Equal(t, "Invalid CPU difficulty: -1. Please enter 0-2.", result)
	})

	t.Run("setdifficulty over 2", func(t *testing.T) {
		c := controller.NewSpadesCuiController(newMock())
		result := c.Exec("sd 3")
		assert.Equal(t, "Invalid CPU difficulty: 3. Please enter 0-2.", result)
	})

	// setlimit
	t.Run("setlimit sl valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewSpadesCuiController(m)
		result := c.Exec("sl 300")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultSpadesConfig()
		expected.PointLimit = 300
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setlimit long form", func(t *testing.T) {
		m := newMock()
		c := controller.NewSpadesCuiController(m)
		result := c.Exec("setlimit 200")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultSpadesConfig()
		expected.PointLimit = 200
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setlimit no args", func(t *testing.T) {
		c := controller.NewSpadesCuiController(newMock())
		result := c.Exec("sl")
		assert.Contains(t, result, "required")
	})

	t.Run("setlimit invalid value", func(t *testing.T) {
		c := controller.NewSpadesCuiController(newMock())
		result := c.Exec("sl abc")
		assert.Contains(t, result, "Invalid point limit")
	})

	t.Run("setlimit zero", func(t *testing.T) {
		c := controller.NewSpadesCuiController(newMock())
		result := c.Exec("sl 0")
		assert.Equal(t, "Invalid point limit: 0. Please enter 1 or more.", result)
	})

	t.Run("setlimit negative", func(t *testing.T) {
		c := controller.NewSpadesCuiController(newMock())
		result := c.Exec("sl -1")
		assert.Equal(t, "Invalid point limit: -1. Please enter 1 or more.", result)
	})

	// log
	t.Run("log command", func(t *testing.T) {
		m := newMock()
		c := controller.NewSpadesCuiController(m)
		result := c.Exec("log")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("l command", func(t *testing.T) {
		m := newMock()
		c := controller.NewSpadesCuiController(m)
		result := c.Exec("l")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ActionLog")
	})

	// hint
	t.Run("hint command h", func(t *testing.T) {
		m := newMock()
		c := controller.NewSpadesCuiController(m)
		result := c.Exec("h")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Hint")
	})

	t.Run("hint command hint", func(t *testing.T) {
		m := newMock()
		c := controller.NewSpadesCuiController(m)
		result := c.Exec("hint")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Hint")
	})

	// unknown / empty
	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewSpadesCuiController(newMock())
		result := c.Exec("unknown")
		assert.Contains(t, result, "コマンドが不明です")
	})

	t.Run("empty command", func(t *testing.T) {
		c := controller.NewSpadesCuiController(newMock())
		result := c.Exec("")
		assert.Contains(t, result, "'help' でコマンド一覧を表示します。")
	})
}
