package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newMockColoradoInteractor() *mockusecase.MockColoradoInteractor {
	return new(mockusecase.MockColoradoInteractor)
}

func TestColoradoCuiControllerSimpleCommands(t *testing.T) {
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
			ci := newMockColoradoInteractor()
			c := NewColoradoCuiController(ci)
			ci.On(tc.method).Return("output")
			for _, alias := range tc.aliases {
				assert.Equal(t, "output", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

func TestColoradoCuiControllerQuit(t *testing.T) {
	c := NewColoradoCuiController(newMockColoradoInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestColoradoCuiControllerMoves(t *testing.T) {
	t.Run("tableau to a foundation", func(t *testing.T) {
		ci := newMockColoradoInteractor()
		c := NewColoradoCuiController(ci)
		ci.On("MoveTableauToFoundation", 2).Return("tf")
		assert.Equal(t, "tf", c.Exec("m t 2 f"))
	})

	// A tableau card has one legal destination, so the destination zone is
	// optional -- and anything other than `f` in that slot is a mistake, not a
	// silently-ignored extra word.
	t.Run("tableau without a destination zone", func(t *testing.T) {
		ci := newMockColoradoInteractor()
		c := NewColoradoCuiController(ci)
		ci.On("MoveTableauToFoundation", 2).Return("tf")
		assert.Equal(t, "tf", c.Exec("m t 2"))
	})

	t.Run("tableau to a tableau is rejected", func(t *testing.T) {
		ci := newMockColoradoInteractor()
		c := NewColoradoCuiController(ci)
		// 1 文字の Contains は本文がほぼ何であっても通ってしまうので、
		// 実際に出る文言そのものを見る。
		assert.Equal(t, i18n.Tf("colorado.invalidToZone", "val", "t"), c.Exec("m t 0 t 5"))
		ci.AssertNotCalled(t, "MoveTableauToFoundation", 0)
	})

	t.Run("waste to a foundation", func(t *testing.T) {
		ci := newMockColoradoInteractor()
		c := NewColoradoCuiController(ci)
		ci.On("MoveWasteToFoundation").Return("wf")
		assert.Equal(t, "wf", c.Exec("m w f"))
	})

	t.Run("waste to a pile", func(t *testing.T) {
		ci := newMockColoradoInteractor()
		c := NewColoradoCuiController(ci)
		ci.On("MoveWasteToTableau", 2).Return("wt")
		assert.Equal(t, "wt", c.Exec("m w t 2"))
	})

	t.Run("stock straight into an empty pile", func(t *testing.T) {
		ci := newMockColoradoInteractor()
		c := NewColoradoCuiController(ci)
		ci.On("MoveStockToTableau", 3).Return("st")
		assert.Equal(t, "st", c.Exec("m s t 3"))
	})

	// The stock can only fill a gap; it never goes straight to a foundation.
	t.Run("rejects the stock to a foundation", func(t *testing.T) {
		ci := newMockColoradoInteractor()
		c := NewColoradoCuiController(ci)
		assert.Equal(t, i18n.Tf("colorado.invalidToZone", "val", "f"), c.Exec("m s f"))
		ci.AssertNotCalled(t, "MoveStockToTableau", 0)
	})
}

func TestColoradoCuiControllerPrompts(t *testing.T) {
	for _, cmd := range []string{"m", "m t", "m w", "m w t", "m s", "m s t"} {
		t.Run(cmd, func(t *testing.T) {
			c := NewColoradoCuiController(newMockColoradoInteractor())
			assert.True(t, cuiutil.IsPromptRequest(c.Exec(cmd)), "%q should prompt for more input", cmd)
		})
	}
}

func TestColoradoCuiControllerErrors(t *testing.T) {
	// 期待値は完全一致で持つ。部分一致だと "t" のような 1 文字が
	// 「たまたま含まれている」だけで通り、壊れても気付けない。
	for _, tc := range []struct{ cmd, want string }{
		{"m x f", i18n.Tf("colorado.invalidFromZone", "val", "x")},
		{"m t abc f", i18n.Tf("colorado.invalidPile", "val", "abc")},
		{"m t 0 z", i18n.Tf("colorado.invalidToZone", "val", "z")},
		{"m t 0 t", i18n.Tf("colorado.invalidToZone", "val", "t")},
		{"m w z", i18n.Tf("colorado.invalidToZone", "val", "z")},
		{"m w t abc", i18n.Tf("colorado.invalidPile", "val", "abc")},
		{"m s t abc", i18n.Tf("colorado.invalidPile", "val", "abc")},
	} {
		t.Run(tc.cmd, func(t *testing.T) {
			ci := newMockColoradoInteractor()
			c := NewColoradoCuiController(ci)
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			ci.AssertExpectations(t)
		})
	}
}
