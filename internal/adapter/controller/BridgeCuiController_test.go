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

func TestBridgeCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockBridgeInteractor {
		m := new(mockUsecases.MockBridgeInteractor)
		m.On("GetConfig").Return(domain.DefaultBridgeConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Bid", mock.Anything, mock.Anything, mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	// quit
	t.Run("quit command q", func(t *testing.T) {
		c := controller.NewBridgeCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("quit command quit", func(t *testing.T) {
		c := controller.NewBridgeCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	// reset
	t.Run("reset command r preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewBridgeCuiController(m)
		result := c.Exec("r")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultBridgeConfig())
	})

	t.Run("reset command reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewBridgeCuiController(m)
		result := c.Exec("reset")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultBridgeConfig())
	})

	// bid
	t.Run("bid command b with type only (pass)", func(t *testing.T) {
		m := newMock()
		c := controller.NewBridgeCuiController(m)
		result := c.Exec("b 0")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Bid", 0, 0, 0)
	})

	t.Run("bid command bid with type level suit", func(t *testing.T) {
		m := newMock()
		c := controller.NewBridgeCuiController(m)
		result := c.Exec("bid 1 2 3")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Bid", 1, 2, 3)
	})

	t.Run("bid command b no args", func(t *testing.T) {
		c := controller.NewBridgeCuiController(newMock())
		result := c.Exec("b")
		assert.Contains(t, result, "Bid type is required")
	})

	t.Run("bid command b invalid arg", func(t *testing.T) {
		c := controller.NewBridgeCuiController(newMock())
		result := c.Exec("b abc")
		assert.Contains(t, result, "Invalid bid type")
	})

	t.Run("bid command b out of range", func(t *testing.T) {
		c := controller.NewBridgeCuiController(newMock())
		result := c.Exec("b 4")
		assert.Contains(t, result, "Invalid bid type: 4")
	})

	t.Run("bid command b below range", func(t *testing.T) {
		c := controller.NewBridgeCuiController(newMock())
		result := c.Exec("b -1")
		assert.Contains(t, result, "Invalid bid type: -1")
	})

	// play
	t.Run("play command p with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewBridgeCuiController(m)
		result := c.Exec("p 3")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("play command play with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewBridgeCuiController(m)
		result := c.Exec("play 5")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", 5)
	})

	t.Run("play command p no args", func(t *testing.T) {
		c := controller.NewBridgeCuiController(newMock())
		result := c.Exec("p")
		assert.Contains(t, result, msgCardIndexRequired())
	})

	t.Run("play command p invalid arg", func(t *testing.T) {
		c := controller.NewBridgeCuiController(newMock())
		result := c.Exec("p abc")
		assert.Contains(t, result, msgInvalidCardIndexPrefix())
	})

	// next
	t.Run("next command n", func(t *testing.T) {
		m := newMock()
		c := controller.NewBridgeCuiController(m)
		result := c.Exec("n")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "NextTrick")
	})

	t.Run("next command next", func(t *testing.T) {
		m := newMock()
		c := controller.NewBridgeCuiController(m)
		result := c.Exec("next")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "NextTrick")
	})

	// nextround
	t.Run("nextround command nr", func(t *testing.T) {
		m := newMock()
		c := controller.NewBridgeCuiController(m)
		result := c.Exec("nr")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "NextRound")
	})

	t.Run("nextround command nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewBridgeCuiController(m)
		result := c.Exec("nextround")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "NextRound")
	})

	// setdifficulty
	t.Run("setdifficulty sd valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewBridgeCuiController(m)
		result := c.Exec("sd 2")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultBridgeConfig()
		expected.CpuDifficulty = domain.BridgeCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty long form", func(t *testing.T) {
		m := newMock()
		c := controller.NewBridgeCuiController(m)
		result := c.Exec("setdifficulty 0")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultBridgeConfig()
		expected.CpuDifficulty = domain.BridgeCpuDifficultyEasy
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty no args", func(t *testing.T) {
		c := controller.NewBridgeCuiController(newMock())
		result := c.Exec("sd")
		assert.Contains(t, result, "required")
	})

	t.Run("setdifficulty invalid value", func(t *testing.T) {
		c := controller.NewBridgeCuiController(newMock())
		result := c.Exec("sd abc")
		assert.Contains(t, result, "Invalid CPU difficulty")
	})

	t.Run("setdifficulty negative", func(t *testing.T) {
		c := controller.NewBridgeCuiController(newMock())
		result := c.Exec("sd -1")
		assert.Equal(t, "Invalid CPU difficulty: -1. Please enter 0-2.", result)
	})

	t.Run("setdifficulty over 2", func(t *testing.T) {
		c := controller.NewBridgeCuiController(newMock())
		result := c.Exec("sd 3")
		assert.Equal(t, "Invalid CPU difficulty: 3. Please enter 0-2.", result)
	})

	// log
	t.Run("log command", func(t *testing.T) {
		m := newMock()
		c := controller.NewBridgeCuiController(m)
		result := c.Exec("log")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("l command", func(t *testing.T) {
		m := newMock()
		c := controller.NewBridgeCuiController(m)
		result := c.Exec("l")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ActionLog")
	})

	// hint
	t.Run("hint command h", func(t *testing.T) {
		m := newMock()
		c := controller.NewBridgeCuiController(m)
		result := c.Exec("h")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Hint")
	})

	t.Run("hint command hint", func(t *testing.T) {
		m := newMock()
		c := controller.NewBridgeCuiController(m)
		result := c.Exec("hint")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Hint")
	})

	// unknown / empty
	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewBridgeCuiController(newMock())
		result := c.Exec("unknown")
		assert.Contains(t, result, "コマンドが不明です")
	})

	t.Run("empty command", func(t *testing.T) {
		c := controller.NewBridgeCuiController(newMock())
		result := c.Exec("")
		assert.Contains(t, result, "'help' でコマンド一覧を表示します。")
	})
}
