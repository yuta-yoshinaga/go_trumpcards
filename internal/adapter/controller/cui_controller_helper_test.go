//go:build test

package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnknownCommandMessage(t *testing.T) {
	assert.Equal(t, "コマンドが不明です: foo", unknownCommandMessage("foo"))
	assert.Equal(t, "コマンドが不明です: ", unknownCommandMessage(""))
}

func TestExecCuiCommand(t *testing.T) {
	resetFn := func(args []string) string {
		if len(args) > 0 {
			return "reset:" + args[0]
		}
		return "reset"
	}
	unknownMsg := func(cmd string) string { return "unknown:" + cmd }
	gameHandler := func(cmd string, args []string) (string, bool) {
		if cmd == "g" {
			return "game", true
		}
		return "", false
	}

	t.Run("empty input", func(t *testing.T) {
		assert.Equal(t, "unknown:", execCuiCommand("", resetFn, unknownMsg, gameHandler))
	})

	t.Run("whitespace only input", func(t *testing.T) {
		assert.Equal(t, "unknown:   ", execCuiCommand("   ", resetFn, unknownMsg, gameHandler))
	})

	t.Run("q command", func(t *testing.T) {
		assert.Equal(t, "bye.", execCuiCommand("q", resetFn, unknownMsg, gameHandler))
	})

	t.Run("quit command", func(t *testing.T) {
		assert.Equal(t, "bye.", execCuiCommand("quit", resetFn, unknownMsg, gameHandler))
	})

	t.Run("r command without args", func(t *testing.T) {
		assert.Equal(t, "reset", execCuiCommand("r", resetFn, unknownMsg, gameHandler))
	})

	t.Run("reset command without args", func(t *testing.T) {
		assert.Equal(t, "reset", execCuiCommand("reset", resetFn, unknownMsg, gameHandler))
	})

	t.Run("r command with args", func(t *testing.T) {
		assert.Equal(t, "reset:tunnel", execCuiCommand("r tunnel", resetFn, unknownMsg, gameHandler))
	})

	t.Run("handled game command", func(t *testing.T) {
		assert.Equal(t, "game", execCuiCommand("g", resetFn, unknownMsg, gameHandler))
	})

	t.Run("unhandled game command falls through to unknown", func(t *testing.T) {
		assert.Equal(t, "unknown:xyz", execCuiCommand("xyz", resetFn, unknownMsg, gameHandler))
	})
}
