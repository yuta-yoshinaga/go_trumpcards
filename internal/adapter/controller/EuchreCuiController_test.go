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

func TestEuchreCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockEuchreInteractor {
		m := new(mockUsecases.MockEuchreInteractor)
		m.On("GetConfig").Return(domain.DefaultEuchreConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("PickUp", mock.Anything, mock.Anything).Return(mockOutput)
		m.On("CallTrump", mock.Anything, mock.Anything).Return(mockOutput)
		m.On("Pass").Return(mockOutput)
		m.On("Discard", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	// quit
	t.Run("quit command q", func(t *testing.T) {
		c := controller.NewEuchreCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("quit command quit", func(t *testing.T) {
		c := controller.NewEuchreCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	// reset
	t.Run("reset command r preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewEuchreCuiController(m)
		result := c.Exec("r")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultEuchreConfig())
	})

	t.Run("reset command reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewEuchreCuiController(m)
		result := c.Exec("reset")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultEuchreConfig())
	})

	// orderup
	t.Run("orderup command o", func(t *testing.T) {
		m := newMock()
		c := controller.NewEuchreCuiController(m)
		result := c.Exec("o")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "PickUp", true, false)
	})

	t.Run("orderup command orderup", func(t *testing.T) {
		m := newMock()
		c := controller.NewEuchreCuiController(m)
		result := c.Exec("orderup")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "PickUp", true, false)
	})

	// orderup alone
	t.Run("orderupalone command oa", func(t *testing.T) {
		m := newMock()
		c := controller.NewEuchreCuiController(m)
		result := c.Exec("oa")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "PickUp", true, true)
	})

	t.Run("orderupalone command orderupalone", func(t *testing.T) {
		m := newMock()
		c := controller.NewEuchreCuiController(m)
		result := c.Exec("orderupalone")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "PickUp", true, true)
	})

	// pass
	t.Run("pass command pa", func(t *testing.T) {
		m := newMock()
		c := controller.NewEuchreCuiController(m)
		result := c.Exec("pa")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Pass")
	})

	t.Run("pass command pass", func(t *testing.T) {
		m := newMock()
		c := controller.NewEuchreCuiController(m)
		result := c.Exec("pass")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Pass")
	})

	// call
	t.Run("call command c with suit", func(t *testing.T) {
		m := newMock()
		c := controller.NewEuchreCuiController(m)
		result := c.Exec("c 2")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "CallTrump", 2, false)
	})

	t.Run("call command call with suit", func(t *testing.T) {
		m := newMock()
		c := controller.NewEuchreCuiController(m)
		result := c.Exec("call 3")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "CallTrump", 3, false)
	})

	t.Run("call command c no args", func(t *testing.T) {
		c := controller.NewEuchreCuiController(newMock())
		result := c.Exec("c")
		assert.Contains(t, result, "Suit is required")
	})

	t.Run("call command c invalid arg", func(t *testing.T) {
		c := controller.NewEuchreCuiController(newMock())
		result := c.Exec("c abc")
		assert.Contains(t, result, "Invalid suit")
	})

	t.Run("call command c out of range", func(t *testing.T) {
		c := controller.NewEuchreCuiController(newMock())
		result := c.Exec("c 5")
		assert.Contains(t, result, "Invalid suit: 5")
	})

	t.Run("call command c below range", func(t *testing.T) {
		c := controller.NewEuchreCuiController(newMock())
		result := c.Exec("c 0")
		assert.Contains(t, result, "Invalid suit: 0")
	})

	// callalone
	t.Run("callalone command ca with suit", func(t *testing.T) {
		m := newMock()
		c := controller.NewEuchreCuiController(m)
		result := c.Exec("ca 1")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "CallTrump", 1, true)
	})

	t.Run("callalone command callalone with suit", func(t *testing.T) {
		m := newMock()
		c := controller.NewEuchreCuiController(m)
		result := c.Exec("callalone 4")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "CallTrump", 4, true)
	})

	t.Run("callalone command ca no args", func(t *testing.T) {
		c := controller.NewEuchreCuiController(newMock())
		result := c.Exec("ca")
		assert.Contains(t, result, "Suit is required")
	})

	// discard
	t.Run("discard command d with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewEuchreCuiController(m)
		result := c.Exec("d 2")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Discard", 2)
	})

	t.Run("discard command discard with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewEuchreCuiController(m)
		result := c.Exec("discard 3")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Discard", 3)
	})

	t.Run("discard command d no args", func(t *testing.T) {
		c := controller.NewEuchreCuiController(newMock())
		result := c.Exec("d")
		assert.Contains(t, result, msgCardIndexRequired())
	})

	t.Run("discard command d invalid arg", func(t *testing.T) {
		c := controller.NewEuchreCuiController(newMock())
		result := c.Exec("d abc")
		assert.Contains(t, result, msgInvalidCardIndexPrefix())
	})

	// play
	t.Run("play command p with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewEuchreCuiController(m)
		result := c.Exec("p 3")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("play command play with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewEuchreCuiController(m)
		result := c.Exec("play 5")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", 5)
	})

	t.Run("play command p no args", func(t *testing.T) {
		c := controller.NewEuchreCuiController(newMock())
		result := c.Exec("p")
		assert.Contains(t, result, msgCardIndexRequired())
	})

	t.Run("play command p invalid arg", func(t *testing.T) {
		c := controller.NewEuchreCuiController(newMock())
		result := c.Exec("p abc")
		assert.Contains(t, result, msgInvalidCardIndexPrefix())
	})

	// next
	t.Run("next command n", func(t *testing.T) {
		m := newMock()
		c := controller.NewEuchreCuiController(m)
		result := c.Exec("n")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "NextTrick")
	})

	t.Run("next command next", func(t *testing.T) {
		m := newMock()
		c := controller.NewEuchreCuiController(m)
		result := c.Exec("next")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "NextTrick")
	})

	// nextround
	t.Run("nextround command nr", func(t *testing.T) {
		m := newMock()
		c := controller.NewEuchreCuiController(m)
		result := c.Exec("nr")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "NextRound")
	})

	t.Run("nextround command nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewEuchreCuiController(m)
		result := c.Exec("nextround")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "NextRound")
	})

	// setdifficulty
	t.Run("setdifficulty sd valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewEuchreCuiController(m)
		result := c.Exec("sd 2")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultEuchreConfig()
		expected.CpuDifficulty = domain.EuchreCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty long form", func(t *testing.T) {
		m := newMock()
		c := controller.NewEuchreCuiController(m)
		result := c.Exec("setdifficulty 0")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultEuchreConfig()
		expected.CpuDifficulty = domain.EuchreCpuDifficultyEasy
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty no args", func(t *testing.T) {
		c := controller.NewEuchreCuiController(newMock())
		result := c.Exec("sd")
		assert.Contains(t, result, msgCpuDifficultyRequired())
	})

	t.Run("setdifficulty invalid value", func(t *testing.T) {
		c := controller.NewEuchreCuiController(newMock())
		result := c.Exec("sd abc")
		assert.Contains(t, result, msgInvalidCpuDifficultyPrefix())
	})

	t.Run("setdifficulty negative", func(t *testing.T) {
		c := controller.NewEuchreCuiController(newMock())
		result := c.Exec("sd -1")
		assert.Equal(t, msgInvalidCpuDifficulty("-1"), result)
	})

	t.Run("setdifficulty over 2", func(t *testing.T) {
		c := controller.NewEuchreCuiController(newMock())
		result := c.Exec("sd 3")
		assert.Equal(t, msgInvalidCpuDifficulty("3"), result)
	})

	// setlimit
	t.Run("setlimit sl valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewEuchreCuiController(m)
		result := c.Exec("sl 20")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultEuchreConfig()
		expected.PointLimit = 20
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setlimit long form", func(t *testing.T) {
		m := newMock()
		c := controller.NewEuchreCuiController(m)
		result := c.Exec("setlimit 15")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultEuchreConfig()
		expected.PointLimit = 15
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setlimit no args", func(t *testing.T) {
		c := controller.NewEuchreCuiController(newMock())
		result := c.Exec("sl")
		assert.Contains(t, result, msgPointLimitRequired())
	})

	t.Run("setlimit invalid value", func(t *testing.T) {
		c := controller.NewEuchreCuiController(newMock())
		result := c.Exec("sl abc")
		assert.Contains(t, result, msgInvalidPointLimitPrefix())
	})

	t.Run("setlimit zero", func(t *testing.T) {
		c := controller.NewEuchreCuiController(newMock())
		result := c.Exec("sl 0")
		assert.Equal(t, msgInvalidPointLimit("0"), result)
	})

	t.Run("setlimit negative", func(t *testing.T) {
		c := controller.NewEuchreCuiController(newMock())
		result := c.Exec("sl -1")
		assert.Equal(t, msgInvalidPointLimit("-1"), result)
	})

	// log
	t.Run("log command", func(t *testing.T) {
		m := newMock()
		c := controller.NewEuchreCuiController(m)
		result := c.Exec("log")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("l command", func(t *testing.T) {
		m := newMock()
		c := controller.NewEuchreCuiController(m)
		result := c.Exec("l")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ActionLog")
	})

	// hint
	t.Run("hint command h", func(t *testing.T) {
		m := newMock()
		c := controller.NewEuchreCuiController(m)
		result := c.Exec("h")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Hint")
	})

	t.Run("hint command hint", func(t *testing.T) {
		m := newMock()
		c := controller.NewEuchreCuiController(m)
		result := c.Exec("hint")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Hint")
	})

	// unknown / empty
	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewEuchreCuiController(newMock())
		result := c.Exec("unknown")
		assert.Contains(t, result, "コマンドが不明です")
	})

	t.Run("empty command", func(t *testing.T) {
		c := controller.NewEuchreCuiController(newMock())
		result := c.Exec("")
		assert.Contains(t, result, "'help' でコマンド一覧を表示します。")
	})
}
