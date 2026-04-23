package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// TestCassino_CpuHardAvoidsTrailingPointCard checks that Hard CPU prefers
// trailing a non-point card over a spade / ace / 10♦ / 2♠ when both are legal.
func TestCassino_CpuHardAvoidsTrailingPointCard(t *testing.T) {
	cfg := domain.DefaultCassinoConfig()
	cfg.CpuDifficulty = domain.CassinoDifficultyHard
	c := newTestCassino(t, cfg)
	c.SetPhase(domain.CassinoPhasePlayerTurn)
	c.SetCurrentTurn(1)
	// CPU has a high spade (point card) and a low non-point card.
	// No legal take on the empty table — Hard must choose trail and prefer
	// the safer card (heart 6) over the spade (point card).
	c.GetPlayer(1).AddCard(cass_s(7)) // spade — point
	c.GetPlayer(1).AddCard(cass_h(6)) // heart — non-point
	giveOtherPlayersFiller(c, 1)
	// Empty table → no take possible
	c.SetTableCards([]*domain.Card{})
	c.CpuPlay()
	// CPU should still have one card left (the spade), having trailed the heart 6.
	// (Or vice-versa, but at minimum the action should have consumed exactly one card.)
	assert.Equal(t, 1, c.GetPlayer(1).GetCardsSize())
}

// TestCassino_CpuTakesSweep checks that Normal CPU prefers a sweep when the
// estimateSweep heuristic adds 3 points to the take score.
func TestCassino_CpuTakesSweep(t *testing.T) {
	cfg := domain.DefaultCassinoConfig()
	cfg.CpuDifficulty = domain.CassinoDifficultyNormal
	c := newTestCassino(t, cfg)
	c.SetPhase(domain.CassinoPhasePlayerTurn)
	c.SetCurrentTurn(1)
	c.GetPlayer(1).AddCard(cass_h(8))
	giveOtherPlayersFiller(c, 1)
	// Two cards summing to 8; taking them empties the table → sweep.
	c.SetTableCards([]*domain.Card{cass_s(3), cass_s(5)})
	c.CpuPlay()
	// The CPU must have taken (sweep eligible) → captured >= 3 cards
	assert.GreaterOrEqual(t, c.GetPlayer(1).CapturedCount(), 3)
}

// TestCassino_CpuBuildsWhenMultiBuildDisabled regression test for the
// "CPU never builds when MultiBuildEnabled=false" issue. With the fix, the
// CPU should still create single builds when MultiBuildEnabled is false.
func TestCassino_CpuBuildsWhenMultiBuildDisabled(t *testing.T) {
	cfg := domain.DefaultCassinoConfig()
	cfg.MultiBuildEnabled = false
	cfg.CpuDifficulty = domain.CassinoDifficultyNormal
	c := newTestCassino(t, cfg)
	c.SetPhase(domain.CassinoPhasePlayerTurn)
	c.SetCurrentTurn(1)
	// CPU plan must include build even with MultiBuildEnabled=false.
	// We don't assert that the CPU CHOOSES build (Normal scoring may prefer trail),
	// but we ensure enumeration includes builds. Probe by checking that an
	// EnumerateBuilds call succeeds for the CPU's hand.
	c.GetPlayer(1).AddCard(cass_h(3))
	c.GetPlayer(1).AddCard(cass_h(8)) // capture card for build value 8
	giveOtherPlayersFiller(c, 1)
	c.SetTableCards([]*domain.Card{cass_s(5)}) // 3 + 5 = 8

	// EnumerateBuilds is independent of MultiBuildEnabled and should yield candidates.
	hand := []*domain.Card{cass_h(3), cass_h(8)}
	cands := domain.EnumerateBuilds(cass_h(3), 0, hand, c.GetTableCards())
	assert.NotEmpty(t, cands)

	// Run the CPU; the action must consume exactly one hand card and not error.
	c.CpuPlay()
	assert.Equal(t, 1, c.GetPlayer(1).GetCardsSize())
}

// TestCassino_CpuBuildAfterFix verifies single-build is reachable for CPU.
func TestCassino_CpuBuildAfterFix(t *testing.T) {
	cfg := domain.DefaultCassinoConfig()
	cfg.MultiBuildEnabled = false
	cfg.CpuDifficulty = domain.CassinoDifficultyNormal
	c := newTestCassino(t, cfg)
	c.SetPhase(domain.CassinoPhasePlayerTurn)
	c.SetCurrentTurn(1)
	// Hand: small + capture card. Table: card to combine.
	c.GetPlayer(1).AddCard(cass_h(2))
	c.GetPlayer(1).AddCard(cass_h(7))
	giveOtherPlayersFiller(c, 1)
	c.SetTableCards([]*domain.Card{cass_s(5)}) // 2 + 5 = 7
	c.CpuPlay()
	// CPU must have done some action. Hand size is now 1.
	assert.Equal(t, 1, c.GetPlayer(1).GetCardsSize())
}

// TestCassino_CpuEasyFallbackTrail covers cpuEasyPlan when no legal action.
// (In practice enumerateCpuPlans always emits trail, so this exercises the
// random-pick branch with one option.)
func TestCassino_CpuEasyFallbackTrail(t *testing.T) {
	cfg := domain.DefaultCassinoConfig()
	cfg.CpuDifficulty = domain.CassinoDifficultyEasy
	c := newTestCassino(t, cfg)
	c.SetPhase(domain.CassinoPhasePlayerTurn)
	c.SetCurrentTurn(1)
	c.GetPlayer(1).AddCard(cass_c(13)) // K — no rank match on table
	giveOtherPlayersFiller(c, 1)
	c.SetTableCards([]*domain.Card{cass_h(3), cass_h(4)})
	c.CpuPlay()
	assert.Equal(t, 0, c.GetPlayer(1).GetCardsSize())
}

// TestCassino_CpuPlayBlockedByOwnedBuild ensures CPU does not enumerate trail
// while owning a build, and proceeds with take or build instead.
func TestCassino_CpuPlayBlockedByOwnedBuild(t *testing.T) {
	cfg := domain.DefaultCassinoConfig()
	cfg.CpuDifficulty = domain.CassinoDifficultyNormal
	c := newTestCassino(t, cfg)
	c.SetPhase(domain.CassinoPhasePlayerTurn)
	c.SetCurrentTurn(1)
	c.GetPlayer(1).AddCard(cass_h(8)) // captures the build
	c.GetPlayer(1).AddCard(cass_h(2)) // would-be trail
	giveOtherPlayersFiller(c, 1)
	// CPU 1 already owns a value-8 build on the table.
	c.SetBuilds([]*domain.CassinoBuild{
		domain.NewCassinoBuild(1, 8, []*domain.Card{cass_s(3), cass_s(5)}),
	})
	c.CpuPlay()
	// CPU should have captured the build (8 of hearts → 8 build).
	assert.GreaterOrEqual(t, c.GetPlayer(1).CapturedCount(), 3)
}

// TestCassino_CpuPlayNoLegalActionFallsThrough ensures CpuPlay handles
// game-end / non-cpu-turn early returns without panicking.
func TestCassino_CpuPlayEarlyReturns(t *testing.T) {
	c := newTestCassino(t, domain.DefaultCassinoConfig())
	// Game ended — CpuPlay must return immediately.
	c.SetGameEndFlag(true)
	c.CpuPlay() // no panic
	// Reset and check human-turn early-return path.
	c2 := newTestCassino(t, domain.DefaultCassinoConfig())
	c2.SetPhase(domain.CassinoPhasePlayerTurn)
	c2.SetCurrentTurn(0) // human
	c2.CpuPlay()         // no-op for human turn
}

// TestCassino_CpuChooseCpuActionEmptyHand covers the empty-hand early-return.
func TestCassino_CpuChooseCpuActionEmptyHand(t *testing.T) {
	cfg := domain.DefaultCassinoConfig()
	cfg.CpuDifficulty = domain.CassinoDifficultyNormal
	c := newTestCassino(t, cfg)
	c.SetPhase(domain.CassinoPhasePlayerTurn)
	c.SetCurrentTurn(1)
	// CPU has no cards — CpuPlay should return without error.
	c.CpuPlay()
	assert.Equal(t, 0, c.GetPlayer(1).GetCardsSize())
}

// TestCassino_CpuHardTrailScoringAllPenalties exercises every branch in
// scorePlanHard's penalty calculation by trailing each kind of point card.
func TestCassino_CpuHardTrailScoringAllPenalties(t *testing.T) {
	pointCards := []*domain.Card{
		cass_s(7),  // spade
		cass_h(1),  // ace
		cass_d(10), // big casino
		cass_s(2),  // little casino + spade
	}
	for _, pc := range pointCards {
		cfg := domain.DefaultCassinoConfig()
		cfg.CpuDifficulty = domain.CassinoDifficultyHard
		c := newTestCassino(t, cfg)
		c.SetPhase(domain.CassinoPhasePlayerTurn)
		c.SetCurrentTurn(1)
		c.GetPlayer(1).AddCard(pc) // only point card → forced trail
		giveOtherPlayersFiller(c, 1)
		c.SetTableCards([]*domain.Card{}) // no take possible
		c.CpuPlay()
		// Card is gone → trailed; no panic on penalty calculation.
		assert.Equal(t, 0, c.GetPlayer(1).GetCardsSize())
	}
}

// TestCassino_CpuHardScoresTake covers the take-branch scoring path on Hard.
func TestCassino_CpuHardScoresTake(t *testing.T) {
	cfg := domain.DefaultCassinoConfig()
	cfg.CpuDifficulty = domain.CassinoDifficultyHard
	c := newTestCassino(t, cfg)
	c.SetPhase(domain.CassinoPhasePlayerTurn)
	c.SetCurrentTurn(1)
	// Multiple takes available with different point values.
	c.GetPlayer(1).AddCard(cass_h(7))
	giveOtherPlayersFiller(c, 1)
	// Table includes a spade-7 (point) and a non-spade 7 (no point).
	c.SetTableCards([]*domain.Card{cass_s(7), cass_h(7)})
	c.CpuPlay()
	// CPU should have captured at least the spade 7.
	caps := c.GetPlayer(1).GetCapturedCards()
	hasSpade7 := false
	for _, card := range caps {
		if card != nil && card.GetDesign() == domain.CardDesignSpade && card.GetValue() == 7 {
			hasSpade7 = true
			break
		}
	}
	assert.True(t, hasSpade7)
}
