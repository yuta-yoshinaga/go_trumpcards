package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockAmericanToadInteractor() *mockusecase.MockAmericanToadInteractor {
	return new(mockusecase.MockAmericanToadInteractor)
}

func TestAmericanToadCuiControllerSimpleCommands(t *testing.T) {
	for _, tc := range []struct {
		method  string
		aliases []string
	}{
		{"Reset", []string{"r", "reset"}},
		{"Draw", []string{"d", "draw"}},
		{"GiveUp", []string{"g", "giveup"}},
		{"Hint", []string{"h", "hint"}},
		{"AutoComplete", []string{"ac", "autocomplete"}},
		{"ActionLog", []string{"log", "l"}},
		{"Undo", []string{"u", "undo"}},
	} {
		t.Run(tc.method, func(t *testing.T) {
			ai := newMockAmericanToadInteractor()
			c := NewAmericanToadCuiController(ai)
			ai.On(tc.method).Return("output")
			for _, alias := range tc.aliases {
				assert.Equal(t, "output", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

func TestAmericanToadCuiControllerQuit(t *testing.T) {
	c := NewAmericanToadCuiController(newMockAmericanToadInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestAmericanToadCuiControllerMoves(t *testing.T) {
	t.Run("reserve to a foundation", func(t *testing.T) {
		ai := newMockAmericanToadInteractor()
		c := NewAmericanToadCuiController(ai)
		ai.On("MoveReserveToFoundation").Return("rf")
		assert.Equal(t, "rf", c.Exec("m r f"))
	})

	t.Run("reserve to a column", func(t *testing.T) {
		ai := newMockAmericanToadInteractor()
		c := NewAmericanToadCuiController(ai)
		ai.On("MoveReserveToTableau", 3).Return("rt")
		assert.Equal(t, "rt", c.Exec("m r t 3"))
	})

	t.Run("waste to a foundation", func(t *testing.T) {
		ai := newMockAmericanToadInteractor()
		c := NewAmericanToadCuiController(ai)
		ai.On("MoveWasteToFoundation").Return("wf")
		assert.Equal(t, "wf", c.Exec("m w f"))
	})

	t.Run("waste to a column", func(t *testing.T) {
		ai := newMockAmericanToadInteractor()
		c := NewAmericanToadCuiController(ai)
		ai.On("MoveWasteToTableau", 2).Return("wt")
		assert.Equal(t, "wt", c.Exec("m w t 2"))
	})

	t.Run("tableau to a foundation", func(t *testing.T) {
		ai := newMockAmericanToadInteractor()
		c := NewAmericanToadCuiController(ai)
		ai.On("MoveTableauToFoundation", 2).Return("tf")
		assert.Equal(t, "tf", c.Exec("m t 2 f"))
	})

	t.Run("tableau to tableau defaults to the top card", func(t *testing.T) {
		ai := newMockAmericanToadInteractor()
		c := NewAmericanToadCuiController(ai)
		ai.On("MoveTableauToTableau", 0, -1, 5).Return("tt")
		assert.Equal(t, "tt", c.Exec("m t 0 t 5"))
	})

	t.Run("tableau to tableau carries an explicit run head", func(t *testing.T) {
		ai := newMockAmericanToadInteractor()
		c := NewAmericanToadCuiController(ai)
		ai.On("MoveTableauToTableau", 0, 2, 5).Return("tt")
		assert.Equal(t, "tt", c.Exec("m t 0 t 5 2"))
	})
}

func TestAmericanToadCuiControllerPrompts(t *testing.T) {
	for _, cmd := range []string{"m", "m r", "m r t", "m w", "m w t", "m t", "m t 0", "m t 0 t"} {
		t.Run(cmd, func(t *testing.T) {
			c := NewAmericanToadCuiController(newMockAmericanToadInteractor())
			assert.True(t, cuiutil.IsPromptRequest(c.Exec(cmd)), "%q should prompt for more input", cmd)
		})
	}
}

func TestAmericanToadCuiControllerErrors(t *testing.T) {
	for _, tc := range []struct{ cmd, contains string }{
		{"m x f", "x"},
		{"m r z", "z"},
		{"m r t abc", "abc"},
		{"m w z", "z"},
		{"m w t abc", "abc"},
		{"m t abc f", "abc"},
		{"m t 0 z", "z"},
		{"m t 0 t abc", "abc"},
		{"m t 0 t 5 abc", "abc"},
	} {
		t.Run(tc.cmd, func(t *testing.T) {
			c := NewAmericanToadCuiController(newMockAmericanToadInteractor())
			assert.Contains(t, c.Exec(tc.cmd), tc.contains)
		})
	}
}
