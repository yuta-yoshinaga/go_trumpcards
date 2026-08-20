package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

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

// #5563: 手詰まりの案内は「undo {{count}} または u を繰り返す」と**両ロケールで
// 既に言っている**のに、コントローラは引数を捨てて 1 手しか戻していなかった。
// 案内どおりに打った人は、脱出したつもりで手詰まりのまま置き去りにされる。
func TestTerraceCuiControllerUndoTakesACount(t *testing.T) {
	for _, alias := range []string{"u", "undo"} {
		t.Run(alias, func(t *testing.T) {
			ti := newMockTerraceInteractor()
			c := NewTerraceCuiController(ti)
			ti.On("UndoN", 5).Return("undone 5")
			assert.Equal(t, "undone 5", c.Exec(alias+" 5"))
			ti.AssertExpectations(t)
		})
	}
}

// 引数なしは今までどおり単発。UndoN(1) に寄せると、既存の Undo が死ぬ。
func TestTerraceCuiControllerBareUndoStillUndoesOnce(t *testing.T) {
	ti := newMockTerraceInteractor()
	c := NewTerraceCuiController(ti)
	ti.On("Undo").Return("undone")
	assert.Equal(t, "undone", c.Exec("undo"))
	ti.AssertNotCalled(t, "UndoN", mock.Anything)
}

// 不正な回数は案内を出して、何も戻さない。
func TestTerraceCuiControllerUndoRejectsABadCount(t *testing.T) {
	for _, arg := range []string{"0", "-1", "zz", "1.5"} {
		t.Run(arg, func(t *testing.T) {
			ti := newMockTerraceInteractor()
			c := NewTerraceCuiController(ti)
			out := c.Exec("undo " + arg)
			assert.Contains(t, out, i18n.Tf("terrace.invalidUndoCount", "val", arg))
			ti.AssertNotCalled(t, "UndoN", mock.Anything)
			ti.AssertNotCalled(t, "Undo")
		})
	}
}

// **回数の上限はここで決めない。**履歴より多い回数はドメインが「戻せる手が
// 足りない」と答える。コントローラで勝手に打ち切ると、Web の undo_n
// (上限なしで素通し) と答えが食い違う。
func TestTerraceCuiControllerUndoPassesLargeCountsThrough(t *testing.T) {
	ti := newMockTerraceInteractor()
	c := NewTerraceCuiController(ti)
	ti.On("UndoN", 9999).Return("not enough history")
	assert.Equal(t, "not enough history", c.Exec("undo 9999"))
	ti.AssertExpectations(t)
}
