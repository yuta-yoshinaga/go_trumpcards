package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockNiuNiuInteractor() *mockusecase.MockNiuNiuInteractor {
	return new(mockusecase.MockNiuNiuInteractor)
}

func TestNiuNiuCuiControllerSimpleCommands(t *testing.T) {
	for _, tc := range []struct {
		method  string
		aliases []string
	}{
		{"Reset", []string{"r", "reset"}},
		{"ActionLog", []string{"log", "l"}},
	} {
		t.Run(tc.method, func(t *testing.T) {
			ni := newMockNiuNiuInteractor()
			c := NewNiuNiuCuiController(ni)
			ni.On(tc.method).Return("output")
			for _, alias := range tc.aliases {
				assert.Equal(t, "output", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

func TestNiuNiuCuiControllerQuit(t *testing.T) {
	c := NewNiuNiuCuiController(newMockNiuNiuInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestNiuNiuCuiControllerBet(t *testing.T) {
	ni := newMockNiuNiuInteractor()
	c := NewNiuNiuCuiController(ni)
	ni.On("Bet", 100).Return("bet")

	assert.Equal(t, "bet", c.Exec("b 100"))
	assert.Equal(t, "bet", c.Exec("bet 100"))
}

func TestNiuNiuCuiControllerRejectsBadAmounts(t *testing.T) {
	for _, tc := range []struct{ cmd, contains string }{
		{"b", "required."},
		{"b abc", "Invalid"},
	} {
		t.Run(tc.cmd, func(t *testing.T) {
			c := NewNiuNiuCuiController(newMockNiuNiuInteractor())
			assert.Contains(t, c.Exec(tc.cmd), tc.contains)
		})
	}
}

func TestNiuNiuCuiControllerUnknownCommand(t *testing.T) {
	c := NewNiuNiuCuiController(newMockNiuNiuInteractor())
	assert.Contains(t, c.Exec("zzzzz"), "zzzzz")
}
