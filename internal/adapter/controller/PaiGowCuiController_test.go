//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockPaiGowInteractor() *usecase.MockPaiGowInteractor {
	m := new(usecase.MockPaiGowInteractor)
	m.On("Reset").Return("reset result")
	m.On("Bet", 100).Return("bet result")
	m.On("SetHands", 0, 1).Return("set result")
	m.On("ActionLog").Return("action log result")
	return m
}

func TestPaiGowCuiController_Quit(t *testing.T) {
	m := newMockPaiGowInteractor()
	c := controller.NewPaiGowCuiController(m)

	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestPaiGowCuiController_Reset(t *testing.T) {
	m := newMockPaiGowInteractor()
	c := controller.NewPaiGowCuiController(m)

	assert.Equal(t, "reset result", c.Exec("r"))
	assert.Equal(t, "reset result", c.Exec("reset"))
}

func TestPaiGowCuiController_Bet(t *testing.T) {
	m := newMockPaiGowInteractor()
	c := controller.NewPaiGowCuiController(m)

	assert.Equal(t, "bet result", c.Exec("b 100"))
	assert.Equal(t, "bet result", c.Exec("bet 100"))
}

func TestPaiGowCuiController_Bet_Errors(t *testing.T) {
	m := newMockPaiGowInteractor()
	c := controller.NewPaiGowCuiController(m)

	t.Run("missing args", func(t *testing.T) {
		result := c.Exec("b")
		assert.Contains(t, result, "Bet amount is required")
	})

	t.Run("invalid amount", func(t *testing.T) {
		result := c.Exec("b abc")
		assert.Contains(t, result, "Invalid bet amount")
	})

	t.Run("zero amount", func(t *testing.T) {
		result := c.Exec("b 0")
		assert.Contains(t, result, "Invalid bet amount")
	})
}

func TestPaiGowCuiController_Set(t *testing.T) {
	m := newMockPaiGowInteractor()
	c := controller.NewPaiGowCuiController(m)

	assert.Equal(t, "set result", c.Exec("s 0 1"))
	assert.Equal(t, "set result", c.Exec("set 0 1"))
}

func TestPaiGowCuiController_Set_Errors(t *testing.T) {
	m := newMockPaiGowInteractor()
	c := controller.NewPaiGowCuiController(m)

	t.Run("missing args", func(t *testing.T) {
		result := c.Exec("s")
		assert.Contains(t, result, "Two card indices are required")
	})

	t.Run("missing second", func(t *testing.T) {
		result := c.Exec("s 0")
		assert.Contains(t, result, "Two card indices are required")
	})

	t.Run("invalid first", func(t *testing.T) {
		result := c.Exec("s abc 1")
		assert.Contains(t, result, "Invalid first index")
	})

	t.Run("invalid second", func(t *testing.T) {
		result := c.Exec("s 0 abc")
		assert.Contains(t, result, "Invalid second index")
	})
}

func TestPaiGowCuiController_ActionLog(t *testing.T) {
	m := newMockPaiGowInteractor()
	c := controller.NewPaiGowCuiController(m)

	assert.Equal(t, "action log result", c.Exec("log"))
	assert.Equal(t, "action log result", c.Exec("l"))
}

func TestPaiGowCuiController_Unknown(t *testing.T) {
	m := newMockPaiGowInteractor()
	c := controller.NewPaiGowCuiController(m)

	result := c.Exec("xyz")
	assert.Contains(t, result, "コマンドが不明です")
}

func TestPaiGowCuiController_Empty(t *testing.T) {
	m := newMockPaiGowInteractor()
	c := controller.NewPaiGowCuiController(m)

	result := c.Exec("")
	assert.Contains(t, result, "'help' でコマンド一覧を表示します。")
}
