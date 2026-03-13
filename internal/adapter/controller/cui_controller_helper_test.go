//go:build test

package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnknownCommandMessage(t *testing.T) {
	t.Run("no suggestion", func(t *testing.T) {
		assert.Equal(t, "コマンドが不明です: foo", unknownCommandMessage("foo", nil))
	})
	t.Run("empty command", func(t *testing.T) {
		assert.Equal(t, "コマンドが不明です: ", unknownCommandMessage("", nil))
	})
	t.Run("with suggestion", func(t *testing.T) {
		assert.Equal(t, "コマンドが不明です: hti。もしかして 'hit' ですか？", unknownCommandMessage("hti", []string{"hit", "stand"}))
	})
	t.Run("no close match", func(t *testing.T) {
		assert.Equal(t, "コマンドが不明です: zzzzzzz", unknownCommandMessage("zzzzzzz", []string{"hit", "stand"}))
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
		assert.Equal(t, "コマンドが不明です: ", result)
	})

	t.Run("whitespace only input", func(t *testing.T) {
		result := execCuiCommand("   ", resetFn, validCmds, gameHandler)
		assert.Equal(t, "コマンドが不明です:    ", result)
	})

	t.Run("q command", func(t *testing.T) {
		assert.Equal(t, "bye.", execCuiCommand("q", resetFn, validCmds, gameHandler))
	})

	t.Run("quit command", func(t *testing.T) {
		assert.Equal(t, "bye.", execCuiCommand("quit", resetFn, validCmds, gameHandler))
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
		assert.Equal(t, "コマンドが不明です: gam。もしかして 'game' ですか？", result)
	})

	t.Run("unhandled game command no suggestion", func(t *testing.T) {
		result := execCuiCommand("zzzzzzz", resetFn, validCmds, gameHandler)
		assert.Equal(t, "コマンドが不明です: zzzzzzz", result)
	})
}
