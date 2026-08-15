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

// advanceRound consolidates 22 byte-identical NextRound implementations
// (Burraco, Canasta, Carioca, Chinchon, Conquian, ContractRummy, …) that
// differed only in receiver name and concrete type — see issue #5368.
type mockRoundGame struct {
	gameEnd    bool
	roundCalls int
}

func (m *mockRoundGame) GetGameEndFlag() bool { return m.gameEnd }
func (m *mockRoundGame) NextRound()           { m.roundCalls++ }

func TestAdvanceRound(t *testing.T) {
	t.Run("advances the round and runs the CPU", func(t *testing.T) {
		g := &mockRoundGame{}
		p := &mockPresenter[*mockRoundGame]{output: "out"}
		cpuRuns := 0

		got := advanceRound(g, p, func() { cpuRuns++ })

		assert.Equal(t, "out", got)
		assert.Equal(t, 1, g.roundCalls)
		assert.Equal(t, 1, cpuRuns)
	})

	// The guard is the whole reason these 22 bodies were identical: without it
	// a finished game would deal another round.
	t.Run("does nothing once the game has ended", func(t *testing.T) {
		g := &mockRoundGame{gameEnd: true}
		p := &mockPresenter[*mockRoundGame]{output: "ended"}
		cpuRuns := 0

		got := advanceRound(g, p, func() { cpuRuns++ })

		assert.Equal(t, "ended", got)
		assert.Equal(t, 0, g.roundCalls, "the round must not advance after the game ends")
		assert.Equal(t, 0, cpuRuns, "the CPU must not play after the game ends")
	})

	// Order matters: the CPU must act on the new round, not the old one.
	t.Run("runs the CPU after the round advances", func(t *testing.T) {
		g := &mockRoundGame{}
		var order []string
		p := &mockPresenter[*mockRoundGame]{output: "out"}
		orderedGame := &orderRecordingGame{mockRoundGame: g, order: &order}

		advanceRound(orderedGame, p2(p), func() { order = append(order, "cpu") })

		assert.Equal(t, []string{"nextRound", "cpu"}, order)
	})
}

// orderRecordingGame records when NextRound runs relative to the CPU callback.
type orderRecordingGame struct {
	*mockRoundGame
	order *[]string
}

func (o *orderRecordingGame) NextRound() {
	*o.order = append(*o.order, "nextRound")
	o.mockRoundGame.NextRound()
}

// p2 adapts the presenter's type parameter to the wrapper type.
func p2(_ *mockPresenter[*mockRoundGame]) *mockPresenter[*orderRecordingGame] {
	return &mockPresenter[*orderRecordingGame]{output: "out"}
}
