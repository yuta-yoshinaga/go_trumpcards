package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newMockBraidInteractor() *mockusecase.MockBraidInteractor {
	return new(mockusecase.MockBraidInteractor)
}

func TestBraidCuiControllerSimpleCommands(t *testing.T) {
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
			bi := newMockBraidInteractor()
			c := NewBraidCuiController(bi)
			bi.On(tc.method).Return("output")
			for _, alias := range tc.aliases {
				assert.Equal(t, "output", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

func TestBraidCuiControllerQuit(t *testing.T) {
	c := NewBraidCuiController(newMockBraidInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestBraidCuiControllerDirection(t *testing.T) {
	for _, tc := range []struct {
		cmd  string
		want bool
	}{
		{"dir a", true},
		{"dir asc", true},
		{"dir up", true},
		{"direction a", true},
		{"dir d", false},
		{"dir desc", false},
		{"dir down", false},
	} {
		t.Run(tc.cmd, func(t *testing.T) {
			bi := newMockBraidInteractor()
			c := NewBraidCuiController(bi)
			bi.On("ChooseDirection", tc.want).Return("dir")
			assert.Equal(t, "dir", c.Exec(tc.cmd))
			bi.AssertCalled(t, "ChooseDirection", tc.want)
		})
	}
}

func TestBraidCuiControllerMoves(t *testing.T) {
	t.Run("braid to a foundation", func(t *testing.T) {
		bi := newMockBraidInteractor()
		c := NewBraidCuiController(bi)
		bi.On("MoveBraidToFoundation").Return("bf")
		assert.Equal(t, "bf", c.Exec("m br f"))
	})

	t.Run("braid field to a foundation", func(t *testing.T) {
		bi := newMockBraidInteractor()
		c := NewBraidCuiController(bi)
		bi.On("MoveFieldToFoundation", 2).Return("ff")
		assert.Equal(t, "ff", c.Exec("m fd 2 f"))
	})

	// The braid and the slots have exactly one destination, so anything but `f`
	// is refused. Both refusals go through invalidArg, so the reply carries the
	// error marker and names the token that was rejected.
	t.Run("rejects a braid destination other than a foundation", func(t *testing.T) {
		c := NewBraidCuiController(newMockBraidInteractor())

		body, isErr := i18n.StripErrorPrefix(c.Exec("m br t"))

		assert.True(t, isErr, "a refused destination must be marked as an error")
		assert.Equal(t, i18n.Tf("braid.onlyToFoundation", "val", "t"), body)
	})

	t.Run("rejects a slot destination other than a foundation", func(t *testing.T) {
		c := NewBraidCuiController(newMockBraidInteractor())

		body, isErr := i18n.StripErrorPrefix(c.Exec("m fd 2 t"))

		assert.True(t, isErr, "a refused destination must be marked as an error")
		assert.Equal(t, i18n.Tf("braid.onlyToFoundation", "val", "t"), body)
	})

	t.Run("helper to a foundation", func(t *testing.T) {
		bi := newMockBraidInteractor()
		c := NewBraidCuiController(bi)
		bi.On("MoveHelperToFoundation", 5).Return("hf")
		assert.Equal(t, "hf", c.Exec("m hp 5 f"))
	})

	t.Run("waste to a foundation", func(t *testing.T) {
		bi := newMockBraidInteractor()
		c := NewBraidCuiController(bi)
		bi.On("MoveWasteToFoundation").Return("wf")
		assert.Equal(t, "wf", c.Exec("m w f"))
	})

	t.Run("waste to a helper", func(t *testing.T) {
		bi := newMockBraidInteractor()
		c := NewBraidCuiController(bi)
		bi.On("MoveWasteToHelper", 3).Return("wh")
		assert.Equal(t, "wh", c.Exec("m w hp 3"))
	})

	// 基礎札以外への行き先は無い。黙って読み替えず、はっきり断る。
	t.Run("rejects a non-foundation destination", func(t *testing.T) {
		c := NewBraidCuiController(newMockBraidInteractor())
		assert.Contains(t, c.Exec("m br hp"), "hp")
		assert.Contains(t, c.Exec("m fd 0 hp"), "hp")
		assert.Contains(t, c.Exec("m hp 0 br"), "br")
	})
}

func TestBraidCuiControllerPrompts(t *testing.T) {
	for _, cmd := range []string{"dir", "m", "m br", "m fd", "m fd 0", "m hp", "m hp 1", "m w", "m w hp"} {
		t.Run(cmd, func(t *testing.T) {
			c := NewBraidCuiController(newMockBraidInteractor())
			assert.True(t, cuiutil.IsPromptRequest(c.Exec(cmd)), "%q should prompt for more input", cmd)
		})
	}
}

func TestBraidCuiControllerErrors(t *testing.T) {
	for _, tc := range []struct{ cmd, contains string }{
		{"dir x", "x"},
		{"m x f", "x"},
		{"m fd abc f", "abc"},
		{"m hp abc f", "abc"},
		{"m w z", "z"},
		{"m w hp abc", "abc"},
	} {
		t.Run(tc.cmd, func(t *testing.T) {
			c := NewBraidCuiController(newMockBraidInteractor())
			assert.Contains(t, c.Exec(tc.cmd), tc.contains)
		})
	}
}
