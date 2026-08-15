package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newMockTerraceInteractor() *mockusecase.MockTerraceInteractor {
	return new(mockusecase.MockTerraceInteractor)
}

func TestTerraceCuiControllerSimpleCommands(t *testing.T) {
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
			ti := newMockTerraceInteractor()
			c := NewTerraceCuiController(ti)
			ti.On(tc.method).Return("output")
			for _, alias := range tc.aliases {
				assert.Equal(t, "output", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

func TestTerraceCuiControllerQuit(t *testing.T) {
	c := NewTerraceCuiController(newMockTerraceInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestTerraceCuiControllerMoves(t *testing.T) {
	t.Run("terrace to a foundation", func(t *testing.T) {
		ti := newMockTerraceInteractor()
		c := NewTerraceCuiController(ti)
		ti.On("MoveReserveToFoundation").Return("rf")
		assert.Equal(t, "rf", c.Exec("m r f"))
	})

	// The terrace has exactly one destination, so a tableau target is refused
	// rather than quietly reinterpreted.
	//
	// The refusal is asserted through i18n rather than with Contains(out, "t"),
	// which the previous version used: every rendering of every message in this
	// package contains a "t", so it held whether the command was refused or run.
	t.Run("rejects the terrace to a pile", func(t *testing.T) {
		c := NewTerraceCuiController(newMockTerraceInteractor())

		out := c.Exec("m r t 2")

		body, isErr := i18n.StripErrorPrefix(out)
		assert.True(t, isErr, "a refused destination must be marked as an error")
		assert.Equal(t, i18n.Tf("terrace.reserveOnlyToFoundation", "val", "t"), body)
	})

	t.Run("waste to a foundation", func(t *testing.T) {
		ti := newMockTerraceInteractor()
		c := NewTerraceCuiController(ti)
		ti.On("MoveWasteToFoundation").Return("wf")
		assert.Equal(t, "wf", c.Exec("m w f"))
	})

	t.Run("waste to a pile", func(t *testing.T) {
		ti := newMockTerraceInteractor()
		c := NewTerraceCuiController(ti)
		ti.On("MoveWasteToTableau", 2).Return("wt")
		assert.Equal(t, "wt", c.Exec("m w t 2"))
	})

	t.Run("tableau to a foundation", func(t *testing.T) {
		ti := newMockTerraceInteractor()
		c := NewTerraceCuiController(ti)
		ti.On("MoveTableauToFoundation", 3).Return("tf")
		assert.Equal(t, "tf", c.Exec("m t 3 f"))
	})

	t.Run("between piles", func(t *testing.T) {
		ti := newMockTerraceInteractor()
		c := NewTerraceCuiController(ti)
		ti.On("MoveTableauToTableau", 0, 5).Return("tt")
		assert.Equal(t, "tt", c.Exec("m t 0 t 5"))
	})
}

func TestTerraceCuiControllerPrompts(t *testing.T) {
	for _, cmd := range []string{"m", "m r", "m w", "m w t", "m t", "m t 0", "m t 0 t"} {
		t.Run(cmd, func(t *testing.T) {
			c := NewTerraceCuiController(newMockTerraceInteractor())
			assert.True(t, cuiutil.IsPromptRequest(c.Exec(cmd)), "%q should prompt for more input", cmd)
		})
	}
}

func TestTerraceCuiControllerErrors(t *testing.T) {
	for _, tc := range []struct{ cmd, contains string }{
		{"m x f", "x"},
		{"m w z", "z"},
		{"m w t abc", "abc"},
		{"m t abc f", "abc"},
		{"m t 0 z", "z"},
		{"m t 0 t abc", "abc"},
	} {
		t.Run(tc.cmd, func(t *testing.T) {
			c := NewTerraceCuiController(newMockTerraceInteractor())
			assert.Contains(t, c.Exec(tc.cmd), tc.contains)
		})
	}
}
