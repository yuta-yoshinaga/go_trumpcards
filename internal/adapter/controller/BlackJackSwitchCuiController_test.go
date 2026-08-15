//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockBlackJackSwitchInteractor() *usecase.MockBlackJackSwitchInteractor {
	m := new(usecase.MockBlackJackSwitchInteractor)
	m.On("Reset").Return("reset")
	m.On("Bet", 100).Return("bet")
	m.On("Switch").Return("switch")
	m.On("Keep").Return("keep")
	m.On("Hit").Return("hit")
	m.On("Stand").Return("stand")
	m.On("DoubleDown").Return("dd")
	m.On("ActionLog").Return("log")
	return m
}

func TestBlackJackSwitchCuiController_Quit(t *testing.T) {
	c := controller.NewBlackJackSwitchCuiController(newMockBlackJackSwitchInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestBlackJackSwitchCuiController_Reset(t *testing.T) {
	c := controller.NewBlackJackSwitchCuiController(newMockBlackJackSwitchInteractor())
	assert.Equal(t, "reset", c.Exec("r"))
	assert.Equal(t, "reset", c.Exec("reset"))
}

func TestBlackJackSwitchCuiController_Bet(t *testing.T) {
	c := controller.NewBlackJackSwitchCuiController(newMockBlackJackSwitchInteractor())
	assert.Equal(t, "bet", c.Exec("b 100"))
	assert.Equal(t, "bet", c.Exec("bet 100"))
}

func TestBlackJackSwitchCuiController_Bet_Errors(t *testing.T) {
	c := controller.NewBlackJackSwitchCuiController(newMockBlackJackSwitchInteractor())
	assert.Contains(t, c.Exec("b"), msgBetAmountRequired())
	assert.Contains(t, c.Exec("b abc"), msgInvalidBetAmountPrefix())
}

func TestBlackJackSwitchCuiController_PlayActions(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"sw", "switch"},
		{"switch", "switch"},
		{"k", "keep"},
		{"keep", "keep"},
		{"h", "hit"},
		{"hit", "hit"},
		{"s", "stand"},
		{"stand", "stand"},
		{"dd", "dd"},
		{"doubledown", "dd"},
		{"log", "log"},
		{"l", "log"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			c := controller.NewBlackJackSwitchCuiController(newMockBlackJackSwitchInteractor())
			assert.Equal(t, tc.want, c.Exec(tc.input))
		})
	}
}

func TestBlackJackSwitchCuiController_Unknown(t *testing.T) {
	c := controller.NewBlackJackSwitchCuiController(newMockBlackJackSwitchInteractor())
	assert.Contains(t, c.Exec("xyz"), "コマンドが不明です")
}

func TestBlackJackSwitchCuiController_Empty(t *testing.T) {
	c := controller.NewBlackJackSwitchCuiController(newMockBlackJackSwitchInteractor())
	assert.NotEmpty(t, c.Exec(""))
}
