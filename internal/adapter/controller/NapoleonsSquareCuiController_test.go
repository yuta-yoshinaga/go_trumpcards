package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockNapoleonsSquareInteractor() *mockusecase.MockNapoleonsSquareInteractor {
	return new(mockusecase.MockNapoleonsSquareInteractor)
}

func TestNapoleonsSquareCuiControllerSimpleCommands(t *testing.T) {
	cases := []struct {
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
	}
	for _, tc := range cases {
		t.Run(tc.method, func(t *testing.T) {
			ni := newMockNapoleonsSquareInteractor()
			c := NewNapoleonsSquareCuiController(ni)
			ni.On(tc.method).Return("output")
			for _, alias := range tc.aliases {
				assert.Equal(t, "output", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

func TestNapoleonsSquareCuiControllerQuit(t *testing.T) {
	c := NewNapoleonsSquareCuiController(newMockNapoleonsSquareInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestNapoleonsSquareCuiControllerMoves(t *testing.T) {
	t.Run("waste to foundation", func(t *testing.T) {
		ni := newMockNapoleonsSquareInteractor()
		c := NewNapoleonsSquareCuiController(ni)
		ni.On("MoveWasteToFoundation").Return("wf")
		assert.Equal(t, "wf", c.Exec("m w f"))
	})

	t.Run("waste to tableau", func(t *testing.T) {
		ni := newMockNapoleonsSquareInteractor()
		c := NewNapoleonsSquareCuiController(ni)
		ni.On("MoveWasteToTableau", 4).Return("wt")
		assert.Equal(t, "wt", c.Exec("m w t 4"))
	})

	t.Run("tableau to foundation", func(t *testing.T) {
		ni := newMockNapoleonsSquareInteractor()
		c := NewNapoleonsSquareCuiController(ni)
		ni.On("MoveTableauToFoundation", 2).Return("tf")
		assert.Equal(t, "tf", c.Exec("m t 2 f"))
	})

	t.Run("tableau to tableau defaults to the top card", func(t *testing.T) {
		ni := newMockNapoleonsSquareInteractor()
		c := NewNapoleonsSquareCuiController(ni)
		ni.On("MoveTableauToTableau", 0, -1, 5).Return("tt")
		assert.Equal(t, "tt", c.Exec("m t 0 t 5"))
	})

	t.Run("tableau to tableau carries an explicit run head", func(t *testing.T) {
		ni := newMockNapoleonsSquareInteractor()
		c := NewNapoleonsSquareCuiController(ni)
		ni.On("MoveTableauToTableau", 0, 2, 5).Return("tt")
		assert.Equal(t, "tt", c.Exec("m t 0 t 5 2"))
	})
}

func TestNapoleonsSquareCuiControllerMovePrompts(t *testing.T) {
	prompts := []string{"m", "m w", "m w t", "m t", "m t 0", "m t 0 t"}
	for _, cmd := range prompts {
		t.Run(cmd, func(t *testing.T) {
			c := NewNapoleonsSquareCuiController(newMockNapoleonsSquareInteractor())
			assert.True(t, cuiutil.IsPromptRequest(c.Exec(cmd)), "%q should prompt for more input", cmd)
		})
	}
}

func TestNapoleonsSquareCuiControllerMoveErrors(t *testing.T) {
	cases := []struct{ cmd, contains string }{
		{"m x f", "x"},
		{"m w z", "z"},
		{"m w t abc", "abc"},
		{"m t abc f", "abc"},
		{"m t 0 z", "z"},
		{"m t 0 t abc", "abc"},
		{"m t 0 t 5 abc", "abc"},
	}
	for _, tc := range cases {
		t.Run(tc.cmd, func(t *testing.T) {
			c := NewNapoleonsSquareCuiController(newMockNapoleonsSquareInteractor())
			assert.Contains(t, c.Exec(tc.cmd), tc.contains)
		})
	}
}
