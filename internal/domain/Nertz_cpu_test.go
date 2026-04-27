//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// --- Scoring & NextRound ---

func TestNertz_RoundScoring_FoundationsAndPenalty(t *testing.T) {
	g := nertzGameForTest(t)
	clearFoundations(g)
	for _, p := range g.GetPlayers() {
		clearPlayerPiles(p)
	}

	// Player 0: contributes A, 2 to foundation 0; nertz pile empty (winner).
	// Player 1: leaves 5 cards in nertz pile.
	// Player 2: leaves 3 cards.
	// Player 3: leaves 0 cards (no contribution).
	founds := g.GetFoundations()
	require.NoError(t, founds[0].Push(newNertzCard(domain.CardDesignSpade, 1), 0))
	require.NoError(t, founds[0].Push(newNertzCard(domain.CardDesignSpade, 2), 0))
	g.SetFoundations(founds)

	for i := 0; i < 5; i++ {
		g.GetPlayers()[1].PushNertz(newNertzCard(domain.CardDesignDiamond, i+1))
	}
	for i := 0; i < 3; i++ {
		g.GetPlayers()[2].PushNertz(newNertzCard(domain.CardDesignClover, i+1))
	}
	// Player 0 has empty nertz, will trigger round end via direct phase signal.
	g.GetPlayers()[0].PushNertz(newNertzCard(domain.CardDesignHeart, 1))
	require.NoError(t, g.MoveNertzToFoundation(0, 1)) // foundation 1 was empty, accept Ace

	assert.Equal(t, domain.NertzPhaseRoundEnd, g.GetPhase())
	// Score: p0 = +2 foundation0 + +1 foundation1 = +3
	// p1 = -2 * 5 = -10
	// p2 = -2 * 3 = -6
	// p3 = 0
	assert.Equal(t, 3, g.GetPlayers()[0].GetScore())
	assert.Equal(t, -10, g.GetPlayers()[1].GetScore())
	assert.Equal(t, -6, g.GetPlayers()[2].GetScore())
	assert.Equal(t, 0, g.GetPlayers()[3].GetScore())
}

func TestNertz_NextRound_PreservesScores(t *testing.T) {
	g := nertzGameForTest(t)
	for i, p := range g.GetPlayers() {
		p.SetScore(10 * (i + 1))
	}
	g.SetPhase(domain.NertzPhaseRoundEnd)
	g.NextRound()
	assert.Equal(t, domain.NertzPhasePlaying, g.GetPhase())
	assert.Equal(t, 2, g.GetRoundNo())
	for i, p := range g.GetPlayers() {
		assert.Equal(t, 10*(i+1), p.GetScore())
	}
}

func TestNertz_GameEndsWhenTargetReached(t *testing.T) {
	g := nertzGameForTest(t)
	cfg := g.GetConfig()
	cfg.TargetScore = domain.NertzTargetScoreMin // 25
	g.ResetWithConfig(cfg)

	clearFoundations(g)
	for _, p := range g.GetPlayers() {
		clearPlayerPiles(p)
	}
	// Pre-load p0 score so a small foundation contribution tips them over
	g.GetPlayers()[0].SetScore(domain.NertzTargetScoreMin - 1)
	g.GetPlayers()[0].PushNertz(newNertzCard(domain.CardDesignSpade, 1))

	require.NoError(t, g.MoveNertzToFoundation(0, 0))
	assert.Equal(t, domain.NertzPhaseGameEnd, g.GetPhase())
	assert.Equal(t, 0, g.GetMatchWinner())
}

func TestNertz_NextRound_NoOpAfterMatchEnd(t *testing.T) {
	g := nertzGameForTest(t)
	cfg := g.GetConfig()
	cfg.TargetScore = domain.NertzTargetScoreMin
	g.ResetWithConfig(cfg)
	clearFoundations(g)
	for _, p := range g.GetPlayers() {
		clearPlayerPiles(p)
	}
	g.GetPlayers()[0].SetScore(domain.NertzTargetScoreMin - 1)
	g.GetPlayers()[0].PushNertz(newNertzCard(domain.CardDesignSpade, 1))
	require.NoError(t, g.MoveNertzToFoundation(0, 0))
	require.Equal(t, domain.NertzPhaseGameEnd, g.GetPhase())

	roundBefore := g.GetRoundNo()
	g.NextRound()
	assert.Equal(t, roundBefore, g.GetRoundNo())
	assert.Equal(t, domain.NertzPhaseGameEnd, g.GetPhase())
}

// --- CPU AI ---

func TestNertz_FindCpuMove_PrefersFoundation(t *testing.T) {
	g := nertzGameForTest(t)
	clearFoundations(g)
	cpu := g.GetPlayers()[1]
	clearPlayerPiles(cpu)
	cpu.PushNertz(newNertzCard(domain.CardDesignSpade, 1)) // playable Ace

	move := g.FindCpuMove(1)
	require.NotNil(t, move)
	assert.Equal(t, "moveNF", move.ActionType)
	assert.Equal(t, 1, move.PlayerIdx)
}

func TestNertz_FindCpuMove_ReturnsNilWhenStuck(t *testing.T) {
	g := nertzGameForTest(t)
	clearFoundations(g)
	cpu := g.GetPlayers()[1]
	clearPlayerPiles(cpu)
	move := g.FindCpuMove(1)
	assert.Nil(t, move)
}

func TestNertz_Tick_AdvancesCpuPlayers(t *testing.T) {
	g := nertzGameForTest(t)
	clearFoundations(g)
	for _, p := range g.GetPlayers() {
		clearPlayerPiles(p)
	}
	// Give CPU 1 a playable Ace; others have nothing.
	g.GetPlayers()[1].PushNertz(newNertzCard(domain.CardDesignClover, 1))

	actions := g.Tick()
	require.NotEmpty(t, actions)
	// CPU 1 should have played its Ace
	assert.Equal(t, 0, g.GetPlayers()[1].NertzSize())
}

func TestNertz_Tick_NoOpWhenNotPlaying(t *testing.T) {
	g := nertzGameForTest(t)
	g.SetPhase(domain.NertzPhaseRoundEnd)
	actions := g.Tick()
	assert.Empty(t, actions)
}

func TestNertz_Tick_DoesNotMoveHumanPlayer(t *testing.T) {
	g := nertzGameForTest(t)
	clearFoundations(g)
	for _, p := range g.GetPlayers() {
		clearPlayerPiles(p)
	}
	human := g.GetPlayers()[0]
	human.PushNertz(newNertzCard(domain.CardDesignHeart, 1))

	g.Tick()
	assert.Equal(t, 1, human.NertzSize(), "Tick must not advance the human player")
}

// TestNertz_FindCpuMove_NoOscillationBetweenEmptyColumns guards against the
// PR #1528 review issue where step 6 (tableau→tableau) would shuffle a whole
// column into an empty destination indefinitely. With an empty Nertz pile and
// no foundation/waste plays, the CPU must fall through to draw or return nil
// rather than emit a moveTT that produces no progress.
func TestNertz_FindCpuMove_NoOscillationBetweenEmptyColumns(t *testing.T) {
	g := nertzGameForTest(t)
	clearFoundations(g)
	cpu := g.GetPlayers()[1]
	clearPlayerPiles(cpu)
	// Column 0 holds a single card, the rest are empty. Nertz pile is empty
	// so a whole-column move into another empty column would be infinite.
	cpu.PushTableau(0, &domain.NertzTableauCard{
		Card:   newNertzCard(domain.CardDesignSpade, 5),
		FaceUp: true,
	})

	move := g.FindCpuMove(1)
	if move != nil {
		assert.NotEqual(t, "moveTT", move.ActionType,
			"CPU must not emit a moveTT when the move only swaps cards between empty columns")
	}
}
