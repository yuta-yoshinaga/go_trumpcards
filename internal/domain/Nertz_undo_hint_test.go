//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestNertz_Undo_RestoresPreviousState(t *testing.T) {
	g := nertzGameForTest(t)
	clearFoundations(g)
	for _, p := range g.GetPlayers() {
		clearPlayerPiles(p)
	}
	p := g.GetPlayers()[0]
	p.PushNertz(newNertzCard(domain.CardDesignHeart, 5))
	p.PushNertz(newNertzCard(domain.CardDesignSpade, 1))
	require.True(t, g.GetPhase() == domain.NertzPhasePlaying)
	require.False(t, g.CanUndo())

	require.NoError(t, g.MoveNertzToFoundation(0, 0))
	assert.Equal(t, 1, p.NertzSize())
	assert.True(t, g.CanUndo())

	require.NoError(t, g.Undo())
	assert.False(t, g.CanUndo())
	// Undo replaces players slice, re-fetch
	p = g.GetPlayers()[0]
	assert.Equal(t, 2, p.NertzSize())
	assert.True(t, g.GetFoundations()[0].IsEmpty())
}

func TestNertz_Undo_NoHistoryReturnsError(t *testing.T) {
	g := nertzGameForTest(t)
	assert.Error(t, g.Undo())
}

func TestNertz_Undo_DoesNotRecordCpuActions(t *testing.T) {
	g := nertzGameForTest(t)
	clearFoundations(g)
	for _, p := range g.GetPlayers() {
		clearPlayerPiles(p)
	}
	cpu := g.GetPlayers()[1]
	cpu.PushNertz(newNertzCard(domain.CardDesignClover, 1))
	require.NoError(t, g.MoveNertzToFoundation(1, 0))
	assert.False(t, g.CanUndo(), "CPU actions must not be recorded for Undo")
}

func TestNertz_GetHint_FoundationBound(t *testing.T) {
	g := nertzGameForTest(t)
	clearFoundations(g)
	for _, p := range g.GetPlayers() {
		clearPlayerPiles(p)
	}
	g.GetPlayers()[0].PushNertz(newNertzCard(domain.CardDesignSpade, 1))

	hint := g.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "nertz", hint.FromZone)
	assert.Equal(t, "foundation", hint.ToZone)
}

func TestNertz_GetHint_NoMoveReturnsNil(t *testing.T) {
	g := nertzGameForTest(t)
	clearFoundations(g)
	for _, p := range g.GetPlayers() {
		clearPlayerPiles(p)
	}
	hint := g.GetHint()
	assert.Nil(t, hint)
}

func TestNertz_GetHint_NotPlaying(t *testing.T) {
	g := nertzGameForTest(t)
	g.SetPhase(domain.NertzPhaseRoundEnd)
	hint := g.GetHint()
	assert.Nil(t, hint)
}
