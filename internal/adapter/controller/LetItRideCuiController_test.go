//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockLetItRideInteractor() *usecase.MockLetItRideInteractor {
	m := new(usecase.MockLetItRideInteractor)
	m.On("Reset").Return("reset result")
	m.On("Bet", 100).Return("bet result")
	m.On("Pull").Return("pull result")
	m.On("LetItRide").Return("letitride result")
	m.On("ActionLog").Return("action log result")
	return m
}

func TestLetItRideCuiController_Quit(t *testing.T) {
	m := newMockLetItRideInteractor()
	c := controller.NewLetItRideCuiController(m)

	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestLetItRideCuiController_Reset(t *testing.T) {
	m := newMockLetItRideInteractor()
	c := controller.NewLetItRideCuiController(m)

	assert.Equal(t, "reset result", c.Exec("r"))
	assert.Equal(t, "reset result", c.Exec("reset"))
}

func TestLetItRideCuiController_Bet(t *testing.T) {
	m := newMockLetItRideInteractor()
	c := controller.NewLetItRideCuiController(m)

	t.Run("short", func(t *testing.T) {
		assert.Equal(t, "bet result", c.Exec("b 100"))
	})

	t.Run("long", func(t *testing.T) {
		assert.Equal(t, "bet result", c.Exec("bet 100"))
	})
}

func TestLetItRideCuiController_Bet_Errors(t *testing.T) {
	m := newMockLetItRideInteractor()
	c := controller.NewLetItRideCuiController(m)

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

	t.Run("negative amount", func(t *testing.T) {
		result := c.Exec("b -10")
		assert.Contains(t, result, "Invalid bet amount")
	})
}

func TestLetItRideCuiController_Pull(t *testing.T) {
	m := newMockLetItRideInteractor()
	c := controller.NewLetItRideCuiController(m)

	assert.Equal(t, "pull result", c.Exec("p"))
	assert.Equal(t, "pull result", c.Exec("pull"))
}

func TestLetItRideCuiController_LetItRide(t *testing.T) {
	m := newMockLetItRideInteractor()
	c := controller.NewLetItRideCuiController(m)

	assert.Equal(t, "letitride result", c.Exec("l"))
	assert.Equal(t, "letitride result", c.Exec("letitride"))
}

func TestLetItRideCuiController_ActionLog(t *testing.T) {
	m := newMockLetItRideInteractor()
	c := controller.NewLetItRideCuiController(m)

	assert.Equal(t, "action log result", c.Exec("log"))
}

func TestLetItRideCuiController_Unknown(t *testing.T) {
	m := newMockLetItRideInteractor()
	c := controller.NewLetItRideCuiController(m)

	result := c.Exec("xyz")
	assert.Contains(t, result, "コマンドが不明です")
}

func TestLetItRideCuiController_Empty(t *testing.T) {
	m := newMockLetItRideInteractor()
	c := controller.NewLetItRideCuiController(m)

	result := c.Exec("")
	assert.Contains(t, result, "'help' でコマンド一覧を表示します。")
}
