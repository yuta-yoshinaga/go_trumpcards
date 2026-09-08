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

func TestOmiCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockOmiInteractor {
		m := new(mockUsecases.MockOmiInteractor)
		m.On("GetConfig").Return(domain.DefaultOmiConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("CallTrump", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	// quit
	t.Run("quit command q", func(t *testing.T) {
		c := controller.NewOmiCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("quit command quit", func(t *testing.T) {
		c := controller.NewOmiCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	// reset
	t.Run("reset command r preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewOmiCuiController(m)
		result := c.Exec("r")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultOmiConfig())
	})

	t.Run("reset command reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewOmiCuiController(m)
		result := c.Exec("reset")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultOmiConfig())
	})

	// removed commands are rejected as unknown
	t.Run("removed commands rejected as unknown", func(t *testing.T) {
		removedCmds := []string{
			"o", "orderup",
			"oa", "orderupalone",
			"pa", "pass",
			"ca", "callalone",
			"d", "discard",
		}
		for _, cmd := range removedCmds {
			c := controller.NewOmiCuiController(newMock())
			result := c.Exec(cmd)
			assert.Contains(t, result, "コマンドが不明です", "command %s should be unknown", cmd)
		}
	})

	// trump / call command
	t.Run("trump command t with suit", func(t *testing.T) {
		m := newMock()
		c := controller.NewOmiCuiController(m)
		result := c.Exec("t 2")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "CallTrump", 2)
	})

	t.Run("trump command trump with suit", func(t *testing.T) {
		m := newMock()
		c := controller.NewOmiCuiController(m)
		result := c.Exec("trump 3")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "CallTrump", 3)
	})

	t.Run("call command c with suit", func(t *testing.T) {
		m := newMock()
		c := controller.NewOmiCuiController(m)
		result := c.Exec("c 1")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "CallTrump", 1)
	})

	t.Run("call command call with suit", func(t *testing.T) {
		m := newMock()
		c := controller.NewOmiCuiController(m)
		result := c.Exec("call 4")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "CallTrump", 4)
	})

	t.Run("trump command t no args", func(t *testing.T) {
		c := controller.NewOmiCuiController(newMock())
		result := c.Exec("t")
		assert.Contains(t, result, msgStem("trumpSuitRequiredRange"))
	})

	t.Run("trump command t invalid arg", func(t *testing.T) {
		c := controller.NewOmiCuiController(newMock())
		result := c.Exec("t abc")
		assert.Contains(t, result, msgStem("invalidTrumpSuitRange"))
	})

	t.Run("trump command t out of range", func(t *testing.T) {
		c := controller.NewOmiCuiController(newMock())
		result := c.Exec("t 5")
		assert.Contains(t, result, msgKey("invalidTrumpSuitRange", "val", "5"))
	})

	t.Run("trump command t below range", func(t *testing.T) {
		c := controller.NewOmiCuiController(newMock())
		result := c.Exec("t 0")
		assert.Contains(t, result, msgKey("invalidTrumpSuitRange", "val", "0"))
	})

	// play
	t.Run("play command p with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewOmiCuiController(m)
		result := c.Exec("p 3")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("play command play with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewOmiCuiController(m)
		result := c.Exec("play 5")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", 5)
	})

	t.Run("play command p no args", func(t *testing.T) {
		c := controller.NewOmiCuiController(newMock())
		result := c.Exec("p")
		assert.Contains(t, result, msgCardIndexRequired())
	})

	t.Run("play command p invalid arg", func(t *testing.T) {
		c := controller.NewOmiCuiController(newMock())
		result := c.Exec("p abc")
		assert.Contains(t, result, msgInvalidCardIndexPrefix())
	})

	// next
	t.Run("next command n", func(t *testing.T) {
		m := newMock()
		c := controller.NewOmiCuiController(m)
		result := c.Exec("n")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "NextTrick")
	})

	t.Run("next command next", func(t *testing.T) {
		m := newMock()
		c := controller.NewOmiCuiController(m)
		result := c.Exec("next")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "NextTrick")
	})

	// nextround
	t.Run("nextround command nr", func(t *testing.T) {
		m := newMock()
		c := controller.NewOmiCuiController(m)
		result := c.Exec("nr")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "NextRound")
	})

	t.Run("nextround command nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewOmiCuiController(m)
		result := c.Exec("nextround")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "NextRound")
	})

	// setdifficulty
	t.Run("setdifficulty sd valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewOmiCuiController(m)
		result := c.Exec("sd 2")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultOmiConfig()
		expected.CpuDifficulty = domain.OmiCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty long form", func(t *testing.T) {
		m := newMock()
		c := controller.NewOmiCuiController(m)
		result := c.Exec("setdifficulty 0")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultOmiConfig()
		expected.CpuDifficulty = domain.OmiCpuDifficultyEasy
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty no args", func(t *testing.T) {
		c := controller.NewOmiCuiController(newMock())
		result := c.Exec("sd")
		assert.Contains(t, result, msgCpuDifficultyRequired())
	})

	t.Run("setdifficulty invalid value", func(t *testing.T) {
		c := controller.NewOmiCuiController(newMock())
		result := c.Exec("sd abc")
		assert.Contains(t, result, msgInvalidCpuDifficultyPrefix())
	})

	t.Run("setdifficulty negative", func(t *testing.T) {
		c := controller.NewOmiCuiController(newMock())
		result := c.Exec("sd -1")
		assert.Equal(t, msgInvalidCpuDifficulty("-1"), result)
	})

	t.Run("setdifficulty over 2", func(t *testing.T) {
		c := controller.NewOmiCuiController(newMock())
		result := c.Exec("sd 3")
		assert.Equal(t, msgInvalidCpuDifficulty("3"), result)
	})

	// setlimit
	t.Run("setlimit sl valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewOmiCuiController(m)
		result := c.Exec("sl 20")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultOmiConfig()
		expected.PointLimit = 20
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setlimit long form", func(t *testing.T) {
		m := newMock()
		c := controller.NewOmiCuiController(m)
		result := c.Exec("setlimit 15")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultOmiConfig()
		expected.PointLimit = 15
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setlimit no args", func(t *testing.T) {
		c := controller.NewOmiCuiController(newMock())
		result := c.Exec("sl")
		assert.Contains(t, result, msgPointLimitRequired())
	})

	t.Run("setlimit invalid value", func(t *testing.T) {
		c := controller.NewOmiCuiController(newMock())
		result := c.Exec("sl abc")
		assert.Contains(t, result, msgInvalidPointLimitPrefix())
	})

	t.Run("setlimit zero", func(t *testing.T) {
		c := controller.NewOmiCuiController(newMock())
		result := c.Exec("sl 0")
		assert.Equal(t, msgInvalidPointLimit("0"), result)
	})

	t.Run("setlimit negative", func(t *testing.T) {
		c := controller.NewOmiCuiController(newMock())
		result := c.Exec("sl -1")
		assert.Equal(t, msgInvalidPointLimit("-1"), result)
	})

	// log
	t.Run("log command", func(t *testing.T) {
		m := newMock()
		c := controller.NewOmiCuiController(m)
		result := c.Exec("log")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("l command", func(t *testing.T) {
		m := newMock()
		c := controller.NewOmiCuiController(m)
		result := c.Exec("l")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ActionLog")
	})

	// hint
	t.Run("hint command h", func(t *testing.T) {
		m := newMock()
		c := controller.NewOmiCuiController(m)
		result := c.Exec("h")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Hint")
	})

	t.Run("hint command hint", func(t *testing.T) {
		m := newMock()
		c := controller.NewOmiCuiController(m)
		result := c.Exec("hint")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Hint")
	})

	// unknown / empty
	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewOmiCuiController(newMock())
		result := c.Exec("unknown")
		assert.Contains(t, result, "コマンドが不明です")
	})

	t.Run("empty command", func(t *testing.T) {
		c := controller.NewOmiCuiController(newMock())
		result := c.Exec("")
		assert.Contains(t, result, "'help' でコマンド一覧を表示します。")
	})
}
