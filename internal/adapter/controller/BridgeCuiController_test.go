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
		assert.Contains(t, result, msgStem("bidTypeRequired"))
	})

	t.Run("bid command b invalid arg", func(t *testing.T) {
		c := controller.NewBridgeCuiController(newMock())
		result := c.Exec("b abc")
		assert.Contains(t, result, msgStem("invalidBidType"))
	})

	t.Run("bid command b out of range", func(t *testing.T) {
		c := controller.NewBridgeCuiController(newMock())
		result := c.Exec("b 4")
		assert.Contains(t, result, msgKey("invalidBidType", "val", "4"))
	})

	t.Run("bid command b below range", func(t *testing.T) {
		c := controller.NewBridgeCuiController(newMock())
		result := c.Exec("b -1")
		assert.Contains(t, result, msgKey("invalidBidType", "val", "-1"))
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
		assert.Contains(t, result, msgCpuDifficultyRequired())
	})

	t.Run("setdifficulty invalid value", func(t *testing.T) {
		c := controller.NewBridgeCuiController(newMock())
		result := c.Exec("sd abc")
		assert.Contains(t, result, msgInvalidCpuDifficultyPrefix())
	})

	t.Run("setdifficulty negative", func(t *testing.T) {
		c := controller.NewBridgeCuiController(newMock())
		result := c.Exec("sd -1")
		assert.Equal(t, msgInvalidCpuDifficulty("-1"), result)
	})

	t.Run("setdifficulty over 2", func(t *testing.T) {
		c := controller.NewBridgeCuiController(newMock())
		result := c.Exec("sd 3")
		assert.Equal(t, msgInvalidCpuDifficulty("3"), result)
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

// #5390: `b 0 abc` は入札レベルを 0 に落として通っていた。
func TestBridgeCuiController_BidRefusesMistypedOptionalArgs(t *testing.T) {
	newMock := func() *mockUsecases.MockBridgeInteractor {
		m := new(mockUsecases.MockBridgeInteractor)
		m.On("Reset").Return("ok")
		m.On("Bid", mock.Anything, mock.Anything, mock.Anything).Return("bid-ok")
		return m
	}

	t.Run("bid level", func(t *testing.T) {
		m := newMock()
		out := controller.NewBridgeCuiController(m).Exec("b 0 abc")
		assert.Equal(t, msgKey("invalidBidLevel", "val", "abc"), out)
		m.AssertNotCalled(t, "Bid", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("bid suit", func(t *testing.T) {
		m := newMock()
		out := controller.NewBridgeCuiController(m).Exec("b 0 1 xyz")
		assert.Equal(t, msgKey("invalidSuit", "val", "xyz"), out)
		m.AssertNotCalled(t, "Bid", mock.Anything, mock.Anything, mock.Anything)
	})

	// **省略形も一緒に見る。** 断る側だけ直して既定値を消しても緑になるので。
	t.Run("both omitted still bids", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, "bid-ok", controller.NewBridgeCuiController(m).Exec("b 0"))
		m.AssertCalled(t, "Bid", 0, 0, 0)
	})
}
