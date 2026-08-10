package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockFourSeasonsInteractor() *mockusecase.MockFourSeasonsInteractor {
	return new(mockusecase.MockFourSeasonsInteractor)
}

func TestFourSeasonsCuiController_NoArgCommands(t *testing.T) {
	tests := []struct {
		method  string
		aliases []string
	}{
		{"Draw", []string{"d", "draw"}},
		{"GiveUp", []string{"g", "giveup"}},
		{"AutoComplete", []string{"ac", "autocomplete"}},
		{"Undo", []string{"u", "undo"}},
		{"Hint", []string{"h", "hint"}},
		{"ActionLog", []string{"log", "l"}},
	}
	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			for _, alias := range tt.aliases {
				ci := newMockFourSeasonsInteractor()
				c := NewFourSeasonsCuiController(ci)
				ci.On(tt.method).Return("out")
				assert.Equal(t, "out", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

func TestFourSeasonsCuiController_Reset(t *testing.T) {
	ci := newMockFourSeasonsInteractor()
	c := NewFourSeasonsCuiController(ci)
	ci.On("Reset").Return("reset_out")
	assert.Equal(t, "reset_out", c.Exec("r"))
	assert.Equal(t, "reset_out", c.Exec("reset"))
}

func TestFourSeasonsCuiController_Quit(t *testing.T) {
	c := NewFourSeasonsCuiController(newMockFourSeasonsInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestFourSeasonsCuiController_Moves(t *testing.T) {
	t.Run("waste to foundation", func(t *testing.T) {
		ci := newMockFourSeasonsInteractor()
		c := NewFourSeasonsCuiController(ci)
		ci.On("MoveWasteToFoundation", 2).Return("ok")
		assert.Equal(t, "ok", c.Exec("m w f 2"))
	})
	t.Run("waste to tableau", func(t *testing.T) {
		ci := newMockFourSeasonsInteractor()
		c := NewFourSeasonsCuiController(ci)
		ci.On("MoveWasteToTableau", 3).Return("ok")
		assert.Equal(t, "ok", c.Exec("m w t 3"))
	})
	t.Run("tableau to foundation", func(t *testing.T) {
		ci := newMockFourSeasonsInteractor()
		c := NewFourSeasonsCuiController(ci)
		ci.On("MoveTableauToFoundation", 4, 1).Return("ok")
		assert.Equal(t, "ok", c.Exec("m t 4 f 1"))
	})
	t.Run("tableau to tableau", func(t *testing.T) {
		ci := newMockFourSeasonsInteractor()
		c := NewFourSeasonsCuiController(ci)
		ci.On("MoveTableauToTableau", 0, 2).Return("ok")
		assert.Equal(t, "ok", c.Exec("m t 0 t 2"))
	})
}

// An incomplete command asks for the rest rather than erroring — that is what
// the prompt prefix marks.
func TestFourSeasonsCuiController_MovePrompts(t *testing.T) {
	c := NewFourSeasonsCuiController(newMockFourSeasonsInteractor())
	for _, input := range []string{"m", "m w", "m w f", "m t", "m t 0", "m t 0 f"} {
		assert.Contains(t, c.Exec(input), cuiutil.PromptPrefix, "input %q", input)
	}
}

func TestFourSeasonsCuiController_MoveRejections(t *testing.T) {
	ci := newMockFourSeasonsInteractor()
	c := NewFourSeasonsCuiController(ci)
	for _, input := range []string{"m x f 1", "m w x 1", "m w f abc", "m t abc f 1", "m t 0 x 1", "m t 0 f abc"} {
		assert.NotEmpty(t, c.Exec(input), "input %q", input)
	}
	ci.AssertNotCalled(t, "MoveWasteToFoundation")
	ci.AssertNotCalled(t, "MoveTableauToFoundation")
}

func TestFourSeasonsCuiController_UnknownAndEmpty(t *testing.T) {
	c := NewFourSeasonsCuiController(newMockFourSeasonsInteractor())
	assert.NotEmpty(t, c.Exec("unknowncmd"))
	assert.NotEmpty(t, c.Exec(""))
}
