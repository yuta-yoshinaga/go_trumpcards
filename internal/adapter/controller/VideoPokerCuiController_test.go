//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockVideoPokerInteractor() *usecase.MockVideoPokerInteractor {
	m := new(usecase.MockVideoPokerInteractor)
	m.On("Reset").Return("reset result")
	m.On("Bet", 3).Return("bet result")
	m.On("Hold", []int{0, 2, 4}).Return("hold result")
	m.On("Hold", []int{}).Return("hold none result")
	m.On("ActionLog").Return("action log result")
	m.On("Hint").Return("hint result")
	return m
}

func TestVideoPokerCuiController_Quit(t *testing.T) {
	m := newMockVideoPokerInteractor()
	c := controller.NewVideoPokerCuiController(m)

	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestVideoPokerCuiController_Reset(t *testing.T) {
	m := newMockVideoPokerInteractor()
	c := controller.NewVideoPokerCuiController(m)

	assert.Equal(t, "reset result", c.Exec("r"))
	assert.Equal(t, "reset result", c.Exec("reset"))
}

func TestVideoPokerCuiController_Bet(t *testing.T) {
	m := newMockVideoPokerInteractor()
	c := controller.NewVideoPokerCuiController(m)

	assert.Equal(t, "bet result", c.Exec("b 3"))
	assert.Equal(t, "bet result", c.Exec("bet 3"))
}

func TestVideoPokerCuiController_Bet_Errors(t *testing.T) {
	m := newMockVideoPokerInteractor()
	c := controller.NewVideoPokerCuiController(m)

	t.Run("missing args", func(t *testing.T) {
		result := c.Exec("b")
		assert.Contains(t, result, "Bet amount is required (1-5)")
	})

	t.Run("invalid amount", func(t *testing.T) {
		result := c.Exec("b abc")
		assert.Contains(t, result, "Invalid bet amount")
	})

	t.Run("zero amount", func(t *testing.T) {
		result := c.Exec("b 0")
		assert.Contains(t, result, "Invalid bet amount")
	})

	t.Run("too high", func(t *testing.T) {
		result := c.Exec("b 6")
		assert.Contains(t, result, "Invalid bet amount")
	})
}

func TestVideoPokerCuiController_Hold(t *testing.T) {
	m := newMockVideoPokerInteractor()
	c := controller.NewVideoPokerCuiController(m)

	t.Run("hold specific", func(t *testing.T) {
		assert.Equal(t, "hold result", c.Exec("h 0 2 4"))
	})

	t.Run("hold none (no args)", func(t *testing.T) {
		assert.Equal(t, "hold none result", c.Exec("h"))
	})

	t.Run("hold long form", func(t *testing.T) {
		assert.Equal(t, "hold result", c.Exec("hold 0 2 4"))
	})
}

func TestVideoPokerCuiController_Hold_Errors(t *testing.T) {
	m := newMockVideoPokerInteractor()
	c := controller.NewVideoPokerCuiController(m)

	t.Run("invalid index text", func(t *testing.T) {
		result := c.Exec("h abc")
		assert.Contains(t, result, msgInvalidCardIndexPrefix())
	})

	t.Run("index out of range", func(t *testing.T) {
		result := c.Exec("h 5")
		assert.Contains(t, result, msgInvalidCardIndexPrefix())
	})

	t.Run("negative index", func(t *testing.T) {
		result := c.Exec("h -1")
		assert.Contains(t, result, msgInvalidCardIndexPrefix())
	})
}

func TestVideoPokerCuiController_ActionLog(t *testing.T) {
	m := newMockVideoPokerInteractor()
	c := controller.NewVideoPokerCuiController(m)

	assert.Equal(t, "action log result", c.Exec("log"))
	assert.Equal(t, "action log result", c.Exec("l"))
}

func TestVideoPokerCuiController_Unknown(t *testing.T) {
	m := newMockVideoPokerInteractor()
	c := controller.NewVideoPokerCuiController(m)

	result := c.Exec("xyz")
	assert.Contains(t, result, "コマンドが不明です")
}

func TestVideoPokerCuiController_Empty(t *testing.T) {
	m := newMockVideoPokerInteractor()
	c := controller.NewVideoPokerCuiController(m)

	result := c.Exec("")
	assert.Contains(t, result, "'help' でコマンド一覧を表示します。")
}

func TestVideoPokerCuiController_Hint(t *testing.T) {
	m := newMockVideoPokerInteractor()
	c := controller.NewVideoPokerCuiController(m)
	assert.Equal(t, "hint result", c.Exec("hint"))
	m.AssertCalled(t, "Hint")
}
