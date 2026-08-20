package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockDiplomatInteractor() *mockusecase.MockDiplomatInteractor {
	return new(mockusecase.MockDiplomatInteractor)
}

func TestDiplomatCuiControllerSimpleCommands(t *testing.T) {
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
			ci := newMockDiplomatInteractor()
			c := NewDiplomatCuiController(ci)
			ci.On(tc.method).Return("output")
			for _, alias := range tc.aliases {
				assert.Equal(t, "output", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

func TestDiplomatCuiControllerQuit(t *testing.T) {
	c := NewDiplomatCuiController(newMockDiplomatInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestDiplomatCuiControllerMoves(t *testing.T) {
	t.Run("tableau to a foundation", func(t *testing.T) {
		ci := newMockDiplomatInteractor()
		c := NewDiplomatCuiController(ci)
		ci.On("MoveTableauToFoundation", 2).Return("tf")
		assert.Equal(t, "tf", c.Exec("m t 2 f"))
	})

	t.Run("between tableau piles", func(t *testing.T) {
		ci := newMockDiplomatInteractor()
		c := NewDiplomatCuiController(ci)
		ci.On("MoveTableauToTableau", 0, 5).Return("tt")
		assert.Equal(t, "tt", c.Exec("m t 0 t 5"))
	})

	t.Run("waste to a foundation", func(t *testing.T) {
		ci := newMockDiplomatInteractor()
		c := NewDiplomatCuiController(ci)
		ci.On("MoveWasteToFoundation").Return("wf")
		assert.Equal(t, "wf", c.Exec("m w f"))
	})

	t.Run("waste to a pile", func(t *testing.T) {
		ci := newMockDiplomatInteractor()
		c := NewDiplomatCuiController(ci)
		ci.On("MoveWasteToTableau", 2).Return("wt")
		assert.Equal(t, "wt", c.Exec("m w t 2"))
	})

	// An empty column is filled from another column or the waste -- never from
	// the stock, so `s` is not a source zone at all.
	t.Run("rejects the stock as a source", func(t *testing.T) {
		ci := newMockDiplomatInteractor()
		c := NewDiplomatCuiController(ci)
		assert.Equal(t, invalidArg("diplomat.invalidFromZone", "val", "s"), c.Exec("m s t 3"))
		ci.AssertExpectations(t)
	})
}

func TestDiplomatCuiControllerPrompts(t *testing.T) {
	for _, cmd := range []string{"m", "m t", "m t 0", "m t 0 t", "m w", "m w t"} {
		t.Run(cmd, func(t *testing.T) {
			c := NewDiplomatCuiController(newMockDiplomatInteractor())
			assert.True(t, cuiutil.IsPromptRequest(c.Exec(cmd)), "%q should prompt for more input", cmd)
		})
	}
}

func TestDiplomatCuiControllerErrors(t *testing.T) {
	for _, tc := range []struct{ cmd, contains string }{
		{"m x f", "x"},
		{"m s t 3", "s"},
		{"m t abc f", "abc"},
		{"m t 0 z", "z"},
		{"m t 0 t abc", "abc"},
		{"m w z", "z"},
		{"m w t abc", "abc"},
	} {
		t.Run(tc.cmd, func(t *testing.T) {
			c := NewDiplomatCuiController(newMockDiplomatInteractor())
			assert.Contains(t, c.Exec(tc.cmd), tc.contains)
		})
	}
}
