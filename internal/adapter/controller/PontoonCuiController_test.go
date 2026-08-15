package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockPontoonInteractor() *mockusecase.MockPontoonInteractor {
	return new(mockusecase.MockPontoonInteractor)
}

func TestPontoonCuiControllerSimpleCommands(t *testing.T) {
	for _, tc := range []struct {
		method  string
		aliases []string
	}{
		{"Reset", []string{"r", "reset"}},
		{"Deal", []string{"deal"}},
		{"Stick", []string{"s", "stick"}},
		{"Twist", []string{"t", "twist"}},
		{"Split", []string{"sp", "split"}},
		{"BankerTwist", []string{"bt"}},
		{"BankerStay", []string{"bs"}},
		{"ActionLog", []string{"log", "l"}},
	} {
		t.Run(tc.method, func(t *testing.T) {
			pi := newMockPontoonInteractor()
			c := NewPontoonCuiController(pi)
			pi.On(tc.method).Return("output")
			for _, alias := range tc.aliases {
				assert.Equal(t, "output", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

func TestPontoonCuiControllerQuit(t *testing.T) {
	c := NewPontoonCuiController(newMockPontoonInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestPontoonCuiControllerBet(t *testing.T) {
	pi := newMockPontoonInteractor()
	c := NewPontoonCuiController(pi)
	pi.On("Bet", 100).Return("bet")

	assert.Equal(t, "bet", c.Exec("b 100"))
	assert.Equal(t, "bet", c.Exec("bet 100"))
}

func TestPontoonCuiControllerBuy(t *testing.T) {
	pi := newMockPontoonInteractor()
	c := NewPontoonCuiController(pi)
	pi.On("Buy", 50).Return("buy")

	assert.Equal(t, "buy", c.Exec("buy 50"))
}

func TestPontoonCuiControllerRejectsBadAmounts(t *testing.T) {
	for _, tc := range []struct{ cmd, contains string }{
		{"b", msgBetAmountRequired()},
		{"b abc", msgInvalidBetAmountPrefix()},
		{"buy", "required."},
		{"buy abc", "Invalid"},
	} {
		t.Run(tc.cmd, func(t *testing.T) {
			c := NewPontoonCuiController(newMockPontoonInteractor())
			assert.Contains(t, c.Exec(tc.cmd), tc.contains)
		})
	}
}

func TestPontoonCuiControllerUnknownCommand(t *testing.T) {
	c := NewPontoonCuiController(newMockPontoonInteractor())
	assert.Contains(t, c.Exec("zzzzz"), "zzzzz")
}
