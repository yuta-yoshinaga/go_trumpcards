package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockRoyalCotillionInteractor() *mockusecase.MockRoyalCotillionInteractor {
	return new(mockusecase.MockRoyalCotillionInteractor)
}

func TestRoyalCotillionCuiControllerSimpleCommands(t *testing.T) {
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
			ci := newMockRoyalCotillionInteractor()
			c := NewRoyalCotillionCuiController(ci)
			ci.On(tc.method).Return("output")
			for _, alias := range tc.aliases {
				assert.Equal(t, "output", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

func TestRoyalCotillionCuiControllerQuit(t *testing.T) {
	c := NewRoyalCotillionCuiController(newMockRoyalCotillionInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestRoyalCotillionCuiControllerMoves(t *testing.T) {
	t.Run("tableau to a foundation", func(t *testing.T) {
		ci := newMockRoyalCotillionInteractor()
		c := NewRoyalCotillionCuiController(ci)
		ci.On("MoveTableauToFoundation", 2).Return("tf")
		assert.Equal(t, "tf", c.Exec("m t 2 f"))
	})

	// タブロー枠もリザーブも行き先は基礎札だけなので、ゾーンを尋ねない。
	t.Run("tableau without a destination zone", func(t *testing.T) {
		ci := newMockRoyalCotillionInteractor()
		c := NewRoyalCotillionCuiController(ci)
		ci.On("MoveTableauToFoundation", 2).Return("tf")
		assert.Equal(t, "tf", c.Exec("m t 2"))
	})

	t.Run("reserve to a foundation", func(t *testing.T) {
		ci := newMockRoyalCotillionInteractor()
		c := NewRoyalCotillionCuiController(ci)
		ci.On("MoveReserveToFoundation", 1).Return("rf")
		assert.Equal(t, "rf", c.Exec("m r 1"))
	})

	t.Run("waste to a foundation", func(t *testing.T) {
		ci := newMockRoyalCotillionInteractor()
		c := NewRoyalCotillionCuiController(ci)
		ci.On("MoveWasteToFoundation").Return("wf")
		assert.Equal(t, "wf", c.Exec("m w f"))
	})

	t.Run("waste to a pile", func(t *testing.T) {
		ci := newMockRoyalCotillionInteractor()
		c := NewRoyalCotillionCuiController(ci)
		ci.On("MoveWasteToTableau", 2).Return("wt")
		assert.Equal(t, "wt", c.Exec("m w t 2"))
	})

	t.Run("stock straight into an empty pile", func(t *testing.T) {
		ci := newMockRoyalCotillionInteractor()
		c := NewRoyalCotillionCuiController(ci)
		ci.On("MoveStockToTableau", 3).Return("st")
		assert.Equal(t, "st", c.Exec("m s t 3"))
	})

	// The stock can only fill a gap; it never goes straight to a foundation.
	t.Run("rejects the stock to a foundation", func(t *testing.T) {
		c := NewRoyalCotillionCuiController(newMockRoyalCotillionInteractor())
		assert.Contains(t, c.Exec("m s f"), "f")
	})
}

func TestRoyalCotillionCuiControllerPrompts(t *testing.T) {
	for _, cmd := range []string{"m", "m t", "m r", "m w", "m w t", "m s", "m s t"} {
		t.Run(cmd, func(t *testing.T) {
			c := NewRoyalCotillionCuiController(newMockRoyalCotillionInteractor())
			assert.True(t, cuiutil.IsPromptRequest(c.Exec(cmd)), "%q should prompt for more input", cmd)
		})
	}
}

func TestRoyalCotillionCuiControllerErrors(t *testing.T) {
	for _, tc := range []struct{ cmd, contains string }{
		{"m x f", "x"},
		{"m t abc f", "abc"},
		{"m t 0 z", "z"},
		{"m t 0 t", "t"},
		{"m r abc", "abc"},
		{"m w z", "z"},
		{"m w t abc", "abc"},
		{"m s t abc", "abc"},
	} {
		t.Run(tc.cmd, func(t *testing.T) {
			c := NewRoyalCotillionCuiController(newMockRoyalCotillionInteractor())
			assert.Contains(t, c.Exec(tc.cmd), tc.contains)
		})
	}
}
