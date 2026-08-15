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

func TestHeartsCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockHeartsInteractor {
		m := new(mockUsecases.MockHeartsInteractor)
		m.On("GetConfig").Return(domain.DefaultHeartsConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Pass", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	// quit
	t.Run("quit command q", func(t *testing.T) {
		c := controller.NewHeartsCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("quit command quit", func(t *testing.T) {
		c := controller.NewHeartsCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	// reset
	t.Run("reset command r preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewHeartsCuiController(m)
		result := c.Exec("r")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultHeartsConfig())
	})

	t.Run("reset command reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewHeartsCuiController(m)
		result := c.Exec("reset")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultHeartsConfig())
	})

	// pass
	t.Run("pass with 3 indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewHeartsCuiController(m)
		result := c.Exec("pass 0 1 2")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Pass", []int{0, 1, 2})
	})

	t.Run("pass with fewer than 3 indices", func(t *testing.T) {
		c := controller.NewHeartsCuiController(newMock())
		result := c.Exec("pass 0 1")
		assert.Contains(t, result, "exactly 3")
	})

	t.Run("pass with more than 3 indices", func(t *testing.T) {
		c := controller.NewHeartsCuiController(newMock())
		result := c.Exec("pass 0 1 2 3")
		assert.Contains(t, result, "exactly 3")
	})

	t.Run("pass with no indices", func(t *testing.T) {
		c := controller.NewHeartsCuiController(newMock())
		result := c.Exec("pass")
		assert.Contains(t, result, "exactly 3")
	})

	t.Run("pass refuses non-numeric args", func(t *testing.T) {
		c := controller.NewHeartsCuiController(newMock())
		result := c.Exec("pass a b c")
		assert.Contains(t, result, "a")
		assert.NotContains(t, result, "exactly 3",
			"the count is only meaningful once every index parsed")
	})

	// **落として数えない。** `pass 0 a 2 3` は a を捨てると 3 枚になり、
	// プレイヤーが選んでいない組み合わせを渡してしまう (issue #5390)。
	t.Run("pass refuses a mistyped index instead of dropping it", func(t *testing.T) {
		m := newMock()
		c := controller.NewHeartsCuiController(m)

		result := c.Exec("pass 0 a 2 3")

		assert.Contains(t, result, "a")
		m.AssertNotCalled(t, "Pass", mock.Anything)
	})

	// This is the case the drop-and-continue behaviour was worst for: `abc` is
	// discarded, the remaining three happen to be a legal pass, and a hand the
	// player never chose goes across (issue #5390).
	t.Run("pass refuses even when the survivors number three", func(t *testing.T) {
		m := newMock()
		c := controller.NewHeartsCuiController(m)

		result := c.Exec("pass 0 abc 1 2")

		assert.Contains(t, result, "abc")
		m.AssertNotCalled(t, "Pass", mock.Anything)
	})

	// play
	t.Run("play command p with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewHeartsCuiController(m)
		result := c.Exec("p 3")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("play command play with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewHeartsCuiController(m)
		result := c.Exec("play 5")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", 5)
	})

	t.Run("play command p no args", func(t *testing.T) {
		c := controller.NewHeartsCuiController(newMock())
		result := c.Exec("p")
		assert.Contains(t, result, msgCardIndexRequired())
	})

	t.Run("play command p invalid arg", func(t *testing.T) {
		c := controller.NewHeartsCuiController(newMock())
		result := c.Exec("p abc")
		assert.Contains(t, result, msgInvalidCardIndexPrefix())
	})

	// next
	t.Run("next command n", func(t *testing.T) {
		m := newMock()
		c := controller.NewHeartsCuiController(m)
		result := c.Exec("n")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "NextTrick")
	})

	t.Run("next command next", func(t *testing.T) {
		m := newMock()
		c := controller.NewHeartsCuiController(m)
		result := c.Exec("next")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "NextTrick")
	})

	// nextround
	t.Run("nextround command nr", func(t *testing.T) {
		m := newMock()
		c := controller.NewHeartsCuiController(m)
		result := c.Exec("nr")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "NextRound")
	})

	t.Run("nextround command nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewHeartsCuiController(m)
		result := c.Exec("nextround")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "NextRound")
	})

	// setdifficulty
	t.Run("setdifficulty sd valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewHeartsCuiController(m)
		result := c.Exec("sd 2")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultHeartsConfig()
		expected.CpuDifficulty = domain.HeartsCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty long form", func(t *testing.T) {
		m := newMock()
		c := controller.NewHeartsCuiController(m)
		result := c.Exec("setdifficulty 0")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultHeartsConfig()
		expected.CpuDifficulty = domain.HeartsCpuDifficultyEasy
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty no args", func(t *testing.T) {
		c := controller.NewHeartsCuiController(newMock())
		result := c.Exec("sd")
		assert.Contains(t, result, msgCpuDifficultyRequired())
	})

	t.Run("setdifficulty invalid value", func(t *testing.T) {
		c := controller.NewHeartsCuiController(newMock())
		result := c.Exec("sd abc")
		assert.Contains(t, result, msgInvalidCpuDifficultyPrefix())
	})

	t.Run("setdifficulty negative", func(t *testing.T) {
		c := controller.NewHeartsCuiController(newMock())
		result := c.Exec("sd -1")
		assert.Equal(t, msgInvalidCpuDifficulty("-1"), result)
	})

	t.Run("setdifficulty over 2", func(t *testing.T) {
		c := controller.NewHeartsCuiController(newMock())
		result := c.Exec("sd 3")
		assert.Equal(t, msgInvalidCpuDifficulty("3"), result)
	})

	// setlimit
	t.Run("setlimit sl valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewHeartsCuiController(m)
		result := c.Exec("sl 50")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultHeartsConfig()
		expected.PointLimit = 50
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setlimit long form", func(t *testing.T) {
		m := newMock()
		c := controller.NewHeartsCuiController(m)
		result := c.Exec("setlimit 200")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultHeartsConfig()
		expected.PointLimit = 200
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setlimit no args", func(t *testing.T) {
		c := controller.NewHeartsCuiController(newMock())
		result := c.Exec("sl")
		assert.Contains(t, result, msgPointLimitRequired())
	})

	t.Run("setlimit invalid value", func(t *testing.T) {
		c := controller.NewHeartsCuiController(newMock())
		result := c.Exec("sl abc")
		assert.Contains(t, result, msgInvalidPointLimitPrefix())
	})

	t.Run("setlimit zero", func(t *testing.T) {
		c := controller.NewHeartsCuiController(newMock())
		result := c.Exec("sl 0")
		assert.Equal(t, msgInvalidPointLimit("0"), result)
	})

	t.Run("setlimit negative", func(t *testing.T) {
		c := controller.NewHeartsCuiController(newMock())
		result := c.Exec("sl -1")
		assert.Equal(t, msgInvalidPointLimit("-1"), result)
	})

	// log
	t.Run("log command", func(t *testing.T) {
		m := newMock()
		c := controller.NewHeartsCuiController(m)
		result := c.Exec("log")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("l command", func(t *testing.T) {
		m := newMock()
		c := controller.NewHeartsCuiController(m)
		result := c.Exec("l")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ActionLog")
	})

	// hint
	t.Run("hint command h", func(t *testing.T) {
		m := newMock()
		c := controller.NewHeartsCuiController(m)
		result := c.Exec("h")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Hint")
	})

	t.Run("hint command hint", func(t *testing.T) {
		m := newMock()
		c := controller.NewHeartsCuiController(m)
		result := c.Exec("hint")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Hint")
	})

	// unknown / empty
	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewHeartsCuiController(newMock())
		result := c.Exec("unknown")
		assert.Contains(t, result, "コマンドが不明です")
	})

	t.Run("empty command", func(t *testing.T) {
		c := controller.NewHeartsCuiController(newMock())
		result := c.Exec("")
		assert.Contains(t, result, "'help' でコマンド一覧を表示します。")
	})
}
