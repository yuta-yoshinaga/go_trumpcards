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
		assert.Contains(t, result, "コマンドが不明です")
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
		m.AssertCalled(t, "ResetWithConfig", domain.SevensConfig{TunnelEnabled: true, JokerCount: 1, CpuStrategy: true, MaxPasses: 5})
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
		m.AssertCalled(t, "ResetWithConfig", domain.SevensConfig{TunnelEnabled: true, CpuStrategy: true, MaxPasses: 10})
	})

	t.Run("reset with nojokerfinish flag", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevensCuiController(m)
		result := c.Exec("r nojokerfinish")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ResetWithConfig", domain.SevensConfig{MaxPasses: 5, NoJokerFinish: true})
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

	t.Run("plain r still calls Reset", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevensCuiController(m)
		result := c.Exec("r")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Reset")
	})
}
