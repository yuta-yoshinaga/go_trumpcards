package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockMissMilliganInteractor() *mockusecase.MockMissMilliganInteractor {
	return new(mockusecase.MockMissMilliganInteractor)
}

func TestMissMilliganCuiControllerSimpleCommands(t *testing.T) {
	for _, tc := range []struct {
		method  string
		aliases []string
	}{
		{"Reset", []string{"r", "reset"}},
		{"Deal", []string{"d", "deal"}},
		{"GiveUp", []string{"g", "giveup"}},
		{"Hint", []string{"h", "hint"}},
		{"AutoComplete", []string{"ac", "autocomplete"}},
		{"ActionLog", []string{"log", "l"}},
		{"Undo", []string{"u", "undo"}},
	} {
		t.Run(tc.method, func(t *testing.T) {
			mi := newMockMissMilliganInteractor()
			c := NewMissMilliganCuiController(mi)
			mi.On(tc.method).Return("output")
			for _, alias := range tc.aliases {
				assert.Equal(t, "output", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

func TestMissMilliganCuiControllerQuit(t *testing.T) {
	c := NewMissMilliganCuiController(newMockMissMilliganInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestMissMilliganCuiControllerMoves(t *testing.T) {
	t.Run("tableau to foundation", func(t *testing.T) {
		mi := newMockMissMilliganInteractor()
		c := NewMissMilliganCuiController(mi)
		mi.On("MoveTableauToFoundation", 2).Return("tf")
		assert.Equal(t, "tf", c.Exec("m t 2 f"))
	})

	t.Run("tableau to tableau defaults to the top card", func(t *testing.T) {
		mi := newMockMissMilliganInteractor()
		c := NewMissMilliganCuiController(mi)
		mi.On("MoveTableauToTableau", 0, -1, 5).Return("tt")
		assert.Equal(t, "tt", c.Exec("m t 0 t 5"))
	})

	t.Run("tableau to tableau carries an explicit run head", func(t *testing.T) {
		mi := newMockMissMilliganInteractor()
		c := NewMissMilliganCuiController(mi)
		mi.On("MoveTableauToTableau", 0, 2, 5).Return("tt")
		assert.Equal(t, "tt", c.Exec("m t 0 t 5 2"))
	})

	t.Run("waived back to the tableau", func(t *testing.T) {
		mi := newMockMissMilliganInteractor()
		c := NewMissMilliganCuiController(mi)
		mi.On("PlaceWaived", 4).Return("wt")
		assert.Equal(t, "wt", c.Exec("m w t 4"))
	})

	t.Run("waived to a foundation", func(t *testing.T) {
		mi := newMockMissMilliganInteractor()
		c := NewMissMilliganCuiController(mi)
		mi.On("MoveWaivedToFoundation").Return("wf")
		assert.Equal(t, "wf", c.Exec("m w f"))
	})
}

func TestMissMilliganCuiControllerWaive(t *testing.T) {
	t.Run("defaults to the top card", func(t *testing.T) {
		mi := newMockMissMilliganInteractor()
		c := NewMissMilliganCuiController(mi)
		mi.On("Waive", 3, -1).Return("wv")
		assert.Equal(t, "wv", c.Exec("wv 3"))
		assert.Equal(t, "wv", c.Exec("waive 3"))
	})

	t.Run("takes an explicit run head", func(t *testing.T) {
		mi := newMockMissMilliganInteractor()
		c := NewMissMilliganCuiController(mi)
		mi.On("Waive", 3, 1).Return("wv")
		assert.Equal(t, "wv", c.Exec("wv 3 1"))
	})
}

func TestMissMilliganCuiControllerPrompts(t *testing.T) {
	for _, cmd := range []string{"m", "m w", "m w t", "m t", "m t 0", "m t 0 t", "wv"} {
		t.Run(cmd, func(t *testing.T) {
			c := NewMissMilliganCuiController(newMockMissMilliganInteractor())
			assert.True(t, cuiutil.IsPromptRequest(c.Exec(cmd)), "%q should prompt for more input", cmd)
		})
	}
}

func TestMissMilliganCuiControllerErrors(t *testing.T) {
	for _, tc := range []struct{ cmd, contains string }{
		{"m x f", "x"},
		{"m w z", "z"},
		{"m w t abc", "abc"},
		{"m t abc f", "abc"},
		{"m t 0 z", "z"},
		{"m t 0 t abc", "abc"},
		{"m t 0 t 5 abc", "abc"},
		{"wv abc", "abc"},
		{"wv 3 abc", "abc"},
	} {
		t.Run(tc.cmd, func(t *testing.T) {
			c := NewMissMilliganCuiController(newMockMissMilliganInteractor())
			assert.Contains(t, c.Exec(tc.cmd), tc.contains)
		})
	}
}
