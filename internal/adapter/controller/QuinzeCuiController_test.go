package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockQuinzeInteractor() *mockusecase.MockQuinzeInteractor {
	return new(mockusecase.MockQuinzeInteractor)
}

func TestQuinzeCuiControllerSimpleCommands(t *testing.T) {
	for _, tc := range []struct {
		method  string
		aliases []string
	}{
		{"Reset", []string{"r", "reset"}},
		{"Deal", []string{"deal"}},
		{"Hit", []string{"h", "hit"}},
		{"Stand", []string{"s", "stand"}},
		{"BankerHit", []string{"bh"}},
		{"BankerStand", []string{"bs"}},
		{"ActionLog", []string{"log", "l"}},
	} {
		t.Run(tc.method, func(t *testing.T) {
			si := newMockQuinzeInteractor()
			c := NewQuinzeCuiController(si)
			si.On(tc.method).Return("output")
			for _, alias := range tc.aliases {
				assert.Equal(t, "output", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

func TestQuinzeCuiControllerQuit(t *testing.T) {
	c := NewQuinzeCuiController(newMockQuinzeInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestQuinzeCuiControllerBet(t *testing.T) {
	si := newMockQuinzeInteractor()
	c := NewQuinzeCuiController(si)
	si.On("Bet", 100).Return("bet")

	assert.Equal(t, "bet", c.Exec("b 100"))
	assert.Equal(t, "bet", c.Exec("bet 100"))
}

func TestQuinzeCuiControllerRejectsBadInput(t *testing.T) {
	for _, tc := range []struct{ cmd, contains string }{
		{"b", msgBetAmountRequired()},

		// 10 と 1〜7 以外は取れない。

	} {
		t.Run(tc.cmd, func(t *testing.T) {
			c := NewQuinzeCuiController(newMockQuinzeInteractor())
			assert.Contains(t, c.Exec(tc.cmd), tc.contains)
		})
	}
}

func TestQuinzeCuiControllerUnknownCommand(t *testing.T) {
	c := NewQuinzeCuiController(newMockQuinzeInteractor())
	assert.Contains(t, c.Exec("zzzzz"), "zzzzz")
}
