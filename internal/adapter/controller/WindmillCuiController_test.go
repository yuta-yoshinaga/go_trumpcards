package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockWindmillInteractor() *mockusecase.MockWindmillInteractor {
	return new(mockusecase.MockWindmillInteractor)
}

func TestWindmillCuiControllerSimpleCommands(t *testing.T) {
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
			wi := newMockWindmillInteractor()
			c := NewWindmillCuiController(wi)
			wi.On(tc.method).Return("output")
			for _, alias := range tc.aliases {
				assert.Equal(t, "output", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

func TestWindmillCuiControllerQuit(t *testing.T) {
	c := NewWindmillCuiController(newMockWindmillInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestWindmillCuiControllerMoves(t *testing.T) {
	t.Run("sail to the centre", func(t *testing.T) {
		wi := newMockWindmillInteractor()
		c := NewWindmillCuiController(wi)
		wi.On("MoveSailToCenter", 3).Return("sc")
		assert.Equal(t, "sc", c.Exec("m s 3 c"))
	})

	t.Run("sail to a corner", func(t *testing.T) {
		wi := newMockWindmillInteractor()
		c := NewWindmillCuiController(wi)
		wi.On("MoveSailToCorner", 3, 1).Return("sk")
		assert.Equal(t, "sk", c.Exec("m s 3 k 1"))
	})

	t.Run("waste to the centre", func(t *testing.T) {
		wi := newMockWindmillInteractor()
		c := NewWindmillCuiController(wi)
		wi.On("MoveWasteToCenter").Return("wc")
		assert.Equal(t, "wc", c.Exec("m w c"))
	})

	t.Run("waste to a corner", func(t *testing.T) {
		wi := newMockWindmillInteractor()
		c := NewWindmillCuiController(wi)
		wi.On("MoveWasteToCorner", 2).Return("wk")
		assert.Equal(t, "wk", c.Exec("m w k 2"))
	})

	t.Run("corner back to the centre", func(t *testing.T) {
		wi := newMockWindmillInteractor()
		c := NewWindmillCuiController(wi)
		wi.On("MoveCornerToCenter", 0).Return("kc")
		assert.Equal(t, "kc", c.Exec("m k 0 c"))
	})

	// The rescue runs one way only, so there is no corner-to-corner syntax to
	// accept -- "m k 0 k 1" must be rejected rather than silently reinterpreted.
	t.Run("rejects a corner-to-corner move", func(t *testing.T) {
		c := NewWindmillCuiController(newMockWindmillInteractor())
		assert.Contains(t, c.Exec("m k 0 k 1"), "k")
	})
}

func TestWindmillCuiControllerPrompts(t *testing.T) {
	for _, cmd := range []string{"m", "m s", "m s 3", "m s 3 k", "m w", "m w k", "m k", "m k 0"} {
		t.Run(cmd, func(t *testing.T) {
			c := NewWindmillCuiController(newMockWindmillInteractor())
			assert.True(t, cuiutil.IsPromptRequest(c.Exec(cmd)), "%q should prompt for more input", cmd)
		})
	}
}

func TestWindmillCuiControllerErrors(t *testing.T) {
	for _, tc := range []struct{ cmd, contains string }{
		{"m x c", "x"},
		{"m s abc c", "abc"},
		{"m s 3 z", "z"},
		{"m s 3 k abc", "abc"},
		{"m w z", "z"},
		{"m w k abc", "abc"},
		{"m k abc c", "abc"},
		{"m k 0 z", "z"},
	} {
		t.Run(tc.cmd, func(t *testing.T) {
			c := NewWindmillCuiController(newMockWindmillInteractor())
			assert.Contains(t, c.Exec(tc.cmd), tc.contains)
		})
	}
}
