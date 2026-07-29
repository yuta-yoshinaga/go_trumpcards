package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockDuchessInteractor() *mockusecase.MockDuchessInteractor {
	return new(mockusecase.MockDuchessInteractor)
}

func TestDuchessCuiControllerSimpleCommands(t *testing.T) {
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
			mi := newMockDuchessInteractor()
			c := NewDuchessCuiController(mi)
			mi.On(tc.method).Return("output")
			for _, alias := range tc.aliases {
				assert.Equal(t, "output", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

func TestDuchessCuiControllerQuit(t *testing.T) {
	c := NewDuchessCuiController(newMockDuchessInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestDuchessCuiControllerChooseBaseRank(t *testing.T) {
	mi := newMockDuchessInteractor()
	c := NewDuchessCuiController(mi)
	mi.On("ChooseBaseRank", 2).Return("base")
	assert.Equal(t, "base", c.Exec("b 2"))
	assert.Equal(t, "base", c.Exec("base 2"))
}

func TestDuchessCuiControllerMoves(t *testing.T) {
	t.Run("reserve to foundation", func(t *testing.T) {
		mi := newMockDuchessInteractor()
		c := NewDuchessCuiController(mi)
		mi.On("MoveReserveToFoundation", 1).Return("rf")
		assert.Equal(t, "rf", c.Exec("m r 1 f"))
	})

	t.Run("reserve to tableau", func(t *testing.T) {
		mi := newMockDuchessInteractor()
		c := NewDuchessCuiController(mi)
		mi.On("MoveReserveToTableau", 3, 0).Return("rt")
		assert.Equal(t, "rt", c.Exec("m r 3 t 0"))
	})

	t.Run("waste to foundation", func(t *testing.T) {
		mi := newMockDuchessInteractor()
		c := NewDuchessCuiController(mi)
		mi.On("MoveWasteToFoundation").Return("wf")
		assert.Equal(t, "wf", c.Exec("m w f"))
	})

	t.Run("waste to tableau", func(t *testing.T) {
		mi := newMockDuchessInteractor()
		c := NewDuchessCuiController(mi)
		mi.On("MoveWasteToTableau", 2).Return("wt")
		assert.Equal(t, "wt", c.Exec("m w t 2"))
	})

	t.Run("tableau to foundation", func(t *testing.T) {
		mi := newMockDuchessInteractor()
		c := NewDuchessCuiController(mi)
		mi.On("MoveTableauToFoundation", 2).Return("tf")
		assert.Equal(t, "tf", c.Exec("m t 2 f"))
	})

	t.Run("tableau to tableau defaults to the top card", func(t *testing.T) {
		mi := newMockDuchessInteractor()
		c := NewDuchessCuiController(mi)
		mi.On("MoveTableauToTableau", 0, -1, 3).Return("tt")
		assert.Equal(t, "tt", c.Exec("m t 0 t 3"))
	})

	t.Run("tableau to tableau carries an explicit run head", func(t *testing.T) {
		mi := newMockDuchessInteractor()
		c := NewDuchessCuiController(mi)
		mi.On("MoveTableauToTableau", 0, 2, 3).Return("tt")
		assert.Equal(t, "tt", c.Exec("m t 0 t 3 2"))
	})
}

func TestDuchessCuiControllerPrompts(t *testing.T) {
	for _, cmd := range []string{"b", "m", "m r", "m r 1", "m r 1 t", "m w", "m w t", "m t", "m t 0", "m t 0 t"} {
		t.Run(cmd, func(t *testing.T) {
			c := NewDuchessCuiController(newMockDuchessInteractor())
			assert.True(t, cuiutil.IsPromptRequest(c.Exec(cmd)), "%q should prompt for more input", cmd)
		})
	}
}

func TestDuchessCuiControllerErrors(t *testing.T) {
	for _, tc := range []struct{ cmd, contains string }{
		{"b abc", "abc"},
		{"m x f", "x"},
		{"m r abc f", "abc"},
		{"m r 1 z", "z"},
		{"m r 1 t abc", "abc"},
		{"m w z", "z"},
		{"m w t abc", "abc"},
		{"m t abc f", "abc"},
		{"m t 0 z", "z"},
		{"m t 0 t abc", "abc"},
		{"m t 0 t 3 abc", "abc"},
	} {
		t.Run(tc.cmd, func(t *testing.T) {
			c := NewDuchessCuiController(newMockDuchessInteractor())
			assert.Contains(t, c.Exec(tc.cmd), tc.contains)
		})
	}
}
