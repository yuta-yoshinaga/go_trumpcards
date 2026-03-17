//go:build test

package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockPlayableGame struct {
	gameEnd   bool
	humanTurn bool
}

func (m *mockPlayableGame) GetGameEndFlag() bool { return m.gameEnd }
func (m *mockPlayableGame) IsHumanTurn() bool    { return m.humanTurn }

type mockGameEndChecker struct {
	gameEnd bool
}

func (m *mockGameEndChecker) GetGameEndFlag() bool { return m.gameEnd }

type mockPresenter[G any] struct {
	output string
}

func (m *mockPresenter[G]) Output(_ G, _ error) string { return m.output }

func TestGuardNotPlayable(t *testing.T) {
	p := &mockPresenter[*mockPlayableGame]{output: "blocked"}

	t.Run("game ended", func(t *testing.T) {
		g := &mockPlayableGame{gameEnd: true, humanTurn: true}
		out, blocked := guardNotPlayable(g, p)
		assert.True(t, blocked)
		assert.Equal(t, "blocked", out)
	})

	t.Run("not human turn", func(t *testing.T) {
		g := &mockPlayableGame{gameEnd: false, humanTurn: false}
		out, blocked := guardNotPlayable(g, p)
		assert.True(t, blocked)
		assert.Equal(t, "blocked", out)
	})

	t.Run("playable", func(t *testing.T) {
		g := &mockPlayableGame{gameEnd: false, humanTurn: true}
		out, blocked := guardNotPlayable(g, p)
		assert.False(t, blocked)
		assert.Equal(t, "", out)
	})
}

func TestGuardGameEnd(t *testing.T) {
	p := &mockPresenter[*mockGameEndChecker]{output: "ended"}

	t.Run("game ended", func(t *testing.T) {
		g := &mockGameEndChecker{gameEnd: true}
		out, blocked := guardGameEnd(g, p)
		assert.True(t, blocked)
		assert.Equal(t, "ended", out)
	})

	t.Run("game not ended", func(t *testing.T) {
		g := &mockGameEndChecker{gameEnd: false}
		out, blocked := guardGameEnd(g, p)
		assert.False(t, blocked)
		assert.Equal(t, "", out)
	})
}
