package controller_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSevensCuiController_Exec(t *testing.T) {
	mockOutput := `{"players":[],"currentTurn":0,"tableMinVals":[0,7,7,7,7],"tableMaxVals":[0,7,7,7,7],"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":""}`

	newMock := func() *mockUsecases.MockSevensInteractor {
		m := new(mockUsecases.MockSevensInteractor)
		m.On("Reset").Return(mockOutput)
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("PlayJoker", mock.Anything, mock.Anything, mock.Anything).Return(mockOutput)
		return m
	}

	t.Run("quit command q", func(t *testing.T) {
		c := controller.NewSevensCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("quit command quit", func(t *testing.T) {
		c := controller.NewSevensCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	t.Run("reset command r", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevensCuiController(m)
		result := c.Exec("r")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Reset")
	})

	t.Run("reset command reset", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevensCuiController(m)
		result := c.Exec("reset")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Reset")
	})

	t.Run("play command p with no index (pass)", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevensCuiController(m)
		result := c.Exec("p")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", -1) // pass = -1
	})

	t.Run("play command play with no index (pass)", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevensCuiController(m)
		result := c.Exec("play")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", -1)
	})

	t.Run("play command p with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevensCuiController(m)
		result := c.Exec("p 2")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", 2)
	})

	t.Run("play command play with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevensCuiController(m)
		result := c.Exec("play 0")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", 0)
	})

	t.Run("joker command j with args", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevensCuiController(m)
		result := c.Exec("j 0 1 6")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "PlayJoker", 0, 1, 6)
	})

	t.Run("joker command joker with args", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevensCuiController(m)
		result := c.Exec("joker 1 3 8")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "PlayJoker", 1, 3, 8)
	})

	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewSevensCuiController(newMock())
		result := c.Exec("unknown")
		assert.Contains(t, result, "コマンドが不明です")
	})

	t.Run("empty command", func(t *testing.T) {
		c := controller.NewSevensCuiController(newMock())
		result := c.Exec("")
		assert.Contains(t, result, "'help' でコマンド一覧を表示します。")
	})

	t.Run("reset with tunnel flag", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevensCuiController(m)
		result := c.Exec("r tunnel")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ResetWithConfig", domain.SevensConfig{TunnelEnabled: true, MaxPasses: 5})
	})

	t.Run("reset with joker=2 flag", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevensCuiController(m)
		result := c.Exec("r joker=2")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ResetWithConfig", domain.SevensConfig{JokerCount: 2, MaxPasses: 5})
	})

	t.Run("reset with all flags", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevensCuiController(m)
		result := c.Exec("r tunnel joker=1 strategy")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ResetWithConfig", domain.SevensConfig{TunnelEnabled: true, JokerCount: 1, CpuStrategy: domain.SevensCpuStrategic, MaxPasses: 5})
	})

	t.Run("reset with passes=3 flag", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevensCuiController(m)
		result := c.Exec("r passes=3")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ResetWithConfig", domain.SevensConfig{MaxPasses: 3})
	})

	t.Run("reset with passes=0 flag (unlimited)", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevensCuiController(m)
		result := c.Exec("r passes=0")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ResetWithConfig", domain.SevensConfig{MaxPasses: 0})
	})

	t.Run("reset with tunnel passes=10 strategy", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevensCuiController(m)
		result := c.Exec("r tunnel passes=10 strategy")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ResetWithConfig", domain.SevensConfig{TunnelEnabled: true, CpuStrategy: domain.SevensCpuStrategic, MaxPasses: 10})
	})

	t.Run("reset with harassment flag", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevensCuiController(m)
		result := c.Exec("r harassment")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ResetWithConfig", domain.SevensConfig{CpuStrategy: domain.SevensCpuHarassment, MaxPasses: 5})
	})

	t.Run("reset with nojokerfinish flag", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevensCuiController(m)
		result := c.Exec("r nojokerfinish")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ResetWithConfig", domain.SevensConfig{MaxPasses: 5, NoJokerFinish: true})
	})

	t.Run("reset with jokerreclaim flag", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevensCuiController(m)
		result := c.Exec("r jokerreclaim")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ResetWithConfig", domain.SevensConfig{MaxPasses: 5, JokerReclaimEnabled: true})
	})

	t.Run("reset with endstop flag", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevensCuiController(m)
		result := c.Exec("r endstop")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ResetWithConfig", domain.SevensConfig{MaxPasses: 5, EndStopEnabled: true})
	})

	t.Run("reset with jokerconsban flag", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevensCuiController(m)
		result := c.Exec("r jokerconsban")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ResetWithConfig", domain.SevensConfig{MaxPasses: 5, JokerConsecutiveBanned: true})
	})

	t.Run("reset with tunnelskip=3 flag", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevensCuiController(m)
		result := c.Exec("r tunnelskip=3")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ResetWithConfig", domain.SevensConfig{TunnelSkipWidth: 3, MaxPasses: 5})
	})

	t.Run("reset with tunnel and tunnelskip=4 flags", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevensCuiController(m)
		result := c.Exec("r tunnel tunnelskip=4")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ResetWithConfig", domain.SevensConfig{TunnelEnabled: true, TunnelSkipWidth: 4, MaxPasses: 5})
	})

	t.Run("plain r still calls Reset", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevensCuiController(m)
		result := c.Exec("r")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Reset")
	})
}

// 打ち間違えた任意引数は既定値に落とさず断る。落とすと `p abc` が 0 番を出し、
// プレイヤーが選んでいない札が場に出る (#5390)。
func TestSevensCuiController_RefusesMistypedOptionalArgs(t *testing.T) {
	mockOutput := `{"players":[],"currentTurn":0,"tableMinVals":[0,7,7,7,7],"tableMaxVals":[0,7,7,7,7],"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":""}`
	newMock := func() *mockUsecases.MockSevensInteractor {
		m := new(mockUsecases.MockSevensInteractor)
		m.On("Reset").Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("PlayJoker", mock.Anything, mock.Anything, mock.Anything).Return(mockOutput)
		return m
	}

	t.Run("play refuses a mistyped index", func(t *testing.T) {
		m := newMock()
		out := controller.NewSevensCuiController(m).Exec("p abc")
		assert.Equal(t, msgKey("invalidCardIndex", "val", "abc"), out)
		m.AssertNotCalled(t, "Play", mock.Anything)
	})

	// **引数なしはパス。** 既定値 -1 が残っていることを一緒に見ておかないと、
	// 断る側だけ直して省略形を壊しても気付けない。
	t.Run("play with no argument still passes", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewSevensCuiController(m).Exec("p"))
		m.AssertCalled(t, "Play", -1)
	})

	// **3 つの引数それぞれに分岐がある。** 真ん中だけ試すと、両端が既定値に
	// 落ちたままでもテストは緑になる。
	t.Run("joker refuses a mistyped card index", func(t *testing.T) {
		m := newMock()
		out := controller.NewSevensCuiController(m).Exec("j abc")
		assert.Equal(t, msgKey("invalidCardIndex", "val", "abc"), out)
		m.AssertNotCalled(t, "PlayJoker", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("joker refuses a mistyped target value", func(t *testing.T) {
		m := newMock()
		out := controller.NewSevensCuiController(m).Exec("j 0 1 zzz")
		assert.Equal(t, msgKey("invalidTargetValue", "val", "zzz"), out)
		m.AssertNotCalled(t, "PlayJoker", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("joker refuses a mistyped suit", func(t *testing.T) {
		m := newMock()
		out := controller.NewSevensCuiController(m).Exec("j 0 xyz")
		assert.Equal(t, msgKey("invalidSuit", "val", "xyz"), out)
		m.AssertNotCalled(t, "PlayJoker", mock.Anything, mock.Anything, mock.Anything)
	})
}
