package controller_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDaifugoCuiController_Exec(t *testing.T) {
	mockOutput := `{"players":[],"currentTurn":0,"tableCards":[],"lastPlayPlayerIdx":-1,"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":""}`

	newMock := func() *mockUsecases.MockDaifugoInteractor {
		m := new(mockUsecases.MockDaifugoInteractor)
		m.On("Reset").Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("Sort", mock.Anything).Return(mockOutput)
		return m
	}

	t.Run("quit command q", func(t *testing.T) {
		c := controller.NewDaifugoCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("quit command quit", func(t *testing.T) {
		c := controller.NewDaifugoCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	t.Run("reset command r", func(t *testing.T) {
		m := newMock()
		c := controller.NewDaifugoCuiController(m)
		result := c.Exec("r")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Reset")
	})

	t.Run("reset command reset", func(t *testing.T) {
		m := newMock()
		c := controller.NewDaifugoCuiController(m)
		result := c.Exec("reset")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Reset")
	})

	t.Run("play command p with no index (pass)", func(t *testing.T) {
		m := newMock()
		c := controller.NewDaifugoCuiController(m)
		result := c.Exec("p")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", []int{})
	})

	t.Run("play command play with no index (pass)", func(t *testing.T) {
		m := newMock()
		c := controller.NewDaifugoCuiController(m)
		result := c.Exec("play")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", []int{})
	})

	t.Run("play command p with one index", func(t *testing.T) {
		m := newMock()
		c := controller.NewDaifugoCuiController(m)
		result := c.Exec("p 2")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", []int{2})
	})

	t.Run("play command play with multiple indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewDaifugoCuiController(m)
		result := c.Exec("play 0 3 4")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", []int{0, 3, 4})
	})

	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewDaifugoCuiController(newMock())
		result := c.Exec("unknown")
		assert.Contains(t, result, "コマンドが不明です")
	})

	t.Run("empty command", func(t *testing.T) {
		c := controller.NewDaifugoCuiController(newMock())
		result := c.Exec("")
		assert.Contains(t, result, "コマンドが不明です")
	})

	t.Run("sort command default mode", func(t *testing.T) {
		m := newMock()
		c := controller.NewDaifugoCuiController(m)
		result := c.Exec("sort")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Sort", mock.MatchedBy(func(mode interface{}) bool {
			return mode == domain.DaifugoSortByStrength
		}))
	})

	t.Run("sort command with mode argument", func(t *testing.T) {
		m := newMock()
		c := controller.NewDaifugoCuiController(m)
		result := c.Exec("sort 1")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Sort", domain.DaifugoSortBySuit)
	})

	t.Run("sort command with invalid mode argument uses default", func(t *testing.T) {
		m := newMock()
		c := controller.NewDaifugoCuiController(m)
		result := c.Exec("sort abc")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Sort", domain.DaifugoSortByStrength)
	})

	t.Run("play command ignores non-numeric index", func(t *testing.T) {
		m := newMock()
		c := controller.NewDaifugoCuiController(m)
		result := c.Exec("p 0 abc 2")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", []int{0, 2})
	})
}
