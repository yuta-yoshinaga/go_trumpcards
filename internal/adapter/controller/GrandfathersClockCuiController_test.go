package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockGrandfathersClockInteractor() *mockusecase.MockGrandfathersClockInteractor {
	return new(mockusecase.MockGrandfathersClockInteractor)
}

func TestGrandfathersClockCuiControllerSimpleCommands(t *testing.T) {
	cases := []struct {
		method  string
		aliases []string
	}{
		{"Reset", []string{"r", "reset"}},
		{"GiveUp", []string{"g", "giveup"}},
		{"Hint", []string{"h", "hint"}},
		{"AutoComplete", []string{"ac", "autocomplete"}},
		{"ActionLog", []string{"log", "l"}},
		{"Undo", []string{"u", "undo"}},
	}
	for _, tc := range cases {
		t.Run(tc.method, func(t *testing.T) {
			gi := newMockGrandfathersClockInteractor()
			c := NewGrandfathersClockCuiController(gi)
			gi.On(tc.method).Return("output")
			for _, alias := range tc.aliases {
				assert.Equal(t, "output", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

func TestGrandfathersClockCuiControllerQuit(t *testing.T) {
	c := NewGrandfathersClockCuiController(newMockGrandfathersClockInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestGrandfathersClockCuiControllerMoves(t *testing.T) {
	t.Run("tableau to clock face", func(t *testing.T) {
		gi := newMockGrandfathersClockInteractor()
		c := NewGrandfathersClockCuiController(gi)
		gi.On("MoveTableauToFoundation", 2, 7).Return("tf")
		assert.Equal(t, "tf", c.Exec("m 2 f 7"))
	})

	t.Run("tableau to tableau", func(t *testing.T) {
		gi := newMockGrandfathersClockInteractor()
		c := NewGrandfathersClockCuiController(gi)
		gi.On("MoveTableauToTableau", 0, 5).Return("tt")
		assert.Equal(t, "tt", c.Exec("m 0 5"))
	})
}

func TestGrandfathersClockCuiControllerMovePrompts(t *testing.T) {
	// The clock face index is not optional: twelve faces can hold the same suit,
	// so it cannot be derived the way Bisley derives its foundation.
	for _, cmd := range []string{"m", "m 0", "m 0 f"} {
		t.Run(cmd, func(t *testing.T) {
			c := NewGrandfathersClockCuiController(newMockGrandfathersClockInteractor())
			assert.True(t, cuiutil.IsPromptRequest(c.Exec(cmd)), "%q should prompt for more input", cmd)
		})
	}
}

func TestGrandfathersClockCuiControllerMoveErrors(t *testing.T) {
	for _, tc := range []struct{ cmd, contains string }{
		{"m abc", "abc"},
		{"m 0 xyz", "xyz"},
		{"m 0 f xyz", "xyz"},
	} {
		t.Run(tc.cmd, func(t *testing.T) {
			c := NewGrandfathersClockCuiController(newMockGrandfathersClockInteractor())
			assert.Contains(t, c.Exec(tc.cmd), tc.contains)
		})
	}
}
