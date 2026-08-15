//go:build test

package controller

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestUnknownCommandMessage(t *testing.T) {
	t.Run("no suggestion", func(t *testing.T) {
		assert.Equal(t, i18n.ErrorPrefix+"コマンドが不明です: foo", unknownCommandMessage("foo", nil))
	})
	t.Run("empty command", func(t *testing.T) {
		assert.Equal(t, i18n.ErrorPrefix+"コマンドが不明です: ", unknownCommandMessage("", nil))
	})
	t.Run("with suggestion", func(t *testing.T) {
		assert.Equal(t, i18n.ErrorPrefix+"コマンドが不明です: hti。もしかして 'hit' ですか？", unknownCommandMessage("hti", []string{"hit", "stand"}))
	})
	t.Run("no close match", func(t *testing.T) {
		assert.Equal(t, i18n.ErrorPrefix+"コマンドが不明です: zzzzzzz", unknownCommandMessage("zzzzzzz", []string{"hit", "stand"}))
	})
}

func TestExecCuiCommand(t *testing.T) {
	resetFn := func(args []string) string {
		if len(args) > 0 {
			return "reset:" + args[0]
		}
		return "reset"
	}
	validCmds := []string{"g", "game"}
	gameHandler := func(cmd string, args []string) (string, bool) {
		if cmd == "g" {
			return "game", true
		}
		return "", false
	}

	t.Run("empty input", func(t *testing.T) {
		result := execCuiCommand("", resetFn, validCmds, gameHandler)
		assert.Equal(t, "'help' でコマンド一覧を表示します。", result)
	})

	t.Run("whitespace only input", func(t *testing.T) {
		result := execCuiCommand("   ", resetFn, validCmds, gameHandler)
		assert.Equal(t, "'help' でコマンド一覧を表示します。", result)
	})

	t.Run("q command", func(t *testing.T) {
		assert.Equal(t, "bye.", execCuiCommand("q", resetFn, validCmds, gameHandler))
	})

	t.Run("quit command", func(t *testing.T) {
		assert.Equal(t, "bye.", execCuiCommand("quit", resetFn, validCmds, gameHandler))
	})

	t.Run("exit command", func(t *testing.T) {
		assert.Equal(t, "bye.", execCuiCommand("exit", resetFn, validCmds, gameHandler))
	})

	t.Run("r command without args", func(t *testing.T) {
		assert.Equal(t, "reset", execCuiCommand("r", resetFn, validCmds, gameHandler))
	})

	t.Run("reset command without args", func(t *testing.T) {
		assert.Equal(t, "reset", execCuiCommand("reset", resetFn, validCmds, gameHandler))
	})

	t.Run("r command with args", func(t *testing.T) {
		assert.Equal(t, "reset:tunnel", execCuiCommand("r tunnel", resetFn, validCmds, gameHandler))
	})

	t.Run("handled game command", func(t *testing.T) {
		assert.Equal(t, "game", execCuiCommand("g", resetFn, validCmds, gameHandler))
	})

	t.Run("unhandled game command with suggestion", func(t *testing.T) {
		result := execCuiCommand("gam", resetFn, validCmds, gameHandler)
		assert.Equal(t, i18n.ErrorPrefix+"コマンドが不明です: gam。もしかして 'game' ですか？", result)
	})

	t.Run("unhandled game command no suggestion", func(t *testing.T) {
		result := execCuiCommand("zzzzzzz", resetFn, validCmds, gameHandler)
		assert.Equal(t, i18n.ErrorPrefix+"コマンドが不明です: zzzzzzz", result)
	})
}

// execSolitaireCui consolidates 6 byte-identical Exec bodies across the
// tableau solitaires (BakersDozen, BeleagueredCastle, Bisley, FlowerGarden,
// KingAlbert, StreetsAndAlleys). They differed only in receiver and in which
// interactor the closures called — see issue #5368.
func TestExecSolitaireCui(t *testing.T) {
	fns := func(calls *[]string) solitaireCuiFns {
		rec := func(name string) func() string {
			return func() string { *calls = append(*calls, name); return name }
		}
		return solitaireCuiFns{
			reset:        rec("reset"),
			giveUp:       rec("giveup"),
			autoComplete: rec("autocomplete"),
			undo:         rec("undo"),
			hint:         rec("hint"),
			actionLog:    rec("log"),
			move: func(args []string) string {
				*calls = append(*calls, "move:"+strings.Join(args, ","))
				return "moved"
			},
		}
	}

	cases := []struct{ cmd, want string }{
		{"r", "reset"}, {"reset", "reset"},
		{"g", "giveup"}, {"giveup", "giveup"},
		{"ac", "autocomplete"}, {"autocomplete", "autocomplete"},
		{"u", "undo"}, {"undo", "undo"},
		{"h", "hint"}, {"hint", "hint"},
		{"l", "log"}, {"log", "log"},
	}
	for _, tc := range cases {
		t.Run(tc.cmd, func(t *testing.T) {
			var calls []string
			execSolitaireCui(tc.cmd, fns(&calls))
			assert.Equal(t, []string{tc.want}, calls)
		})
	}

	t.Run("move forwards its arguments", func(t *testing.T) {
		var calls []string
		got := execSolitaireCui("m w 1", fns(&calls))
		assert.Equal(t, "moved", got)
		assert.Equal(t, []string{"move:w,1"}, calls)
	})

	// An unknown command must reach the shared suggestion path rather than
	// being answered here: that is what produces "もしかして…" for a typo.
	t.Run("unknown command is not answered by any of the callbacks", func(t *testing.T) {
		var calls []string
		out := execSolitaireCui("frobnicate", fns(&calls))
		assert.Empty(t, calls, "no interactor method should have run")
		assert.NotEmpty(t, out, "the shared helper still answers with a suggestion")
	})
}
