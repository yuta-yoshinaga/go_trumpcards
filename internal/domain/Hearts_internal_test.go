//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newInternalTestHearts() *Hearts {
	players := []*HeartsPlayer{
		NewHeartsPlayer(true),
		NewHeartsPlayer(false),
		NewHeartsPlayer(false),
		NewHeartsPlayer(false),
	}
	cfg := DefaultHeartsConfig()
	tc := NewTrumpCards(0)
	return NewHearts(tc, players, cfg)
}

func TestHearts_trickWinner_EmptyTrick(t *testing.T) {
	h := newInternalTestHearts()
	h.currentTrick = nil
	assert.Equal(t, 0, h.trickWinner())
}

func TestHearts_cpuSelectPlayCard_EmptyValidIndices(t *testing.T) {
	h := newInternalTestHearts()
	h.phase = HeartsPhasePlay
	// Player with no cards → getValidPlayIndices returns empty
	h.players[1].Reset()
	h.currentPlayerIdx = 1
	h.currentTrick = nil
	result := h.cpuSelectPlayCard(1)
	assert.Equal(t, 0, result)
}

func TestHearts_passTarget_DefaultCase(t *testing.T) {
	h := newInternalTestHearts()
	// HeartsPassNone returns the same player
	result := h.passTarget(2, HeartsPassNone)
	assert.Equal(t, 2, result)
}

func TestHearts_passDirectionStr_DefaultCase(t *testing.T) {
	h := newInternalTestHearts()
	result := h.passDirectionStr(HeartsPassNone)
	assert.Equal(t, "none", result)
}

func TestHearts_playerName_OutOfBounds(t *testing.T) {
	h := newInternalTestHearts()
	assert.Equal(t, "Player -1", playerName(h.players, -1))
	assert.Equal(t, "Player 10", playerName(h.players, 10))
}

func TestHearts_startPlayPhase_NotFirstTrick(t *testing.T) {
	h := newInternalTestHearts()
	h.trickNumber = 5
	h.leadPlayerIdx = 2
	h.startPlayPhase()
	assert.Equal(t, 2, h.currentPlayerIdx)
}

// --- Omnibus J♦ internal tests ---

func TestCardPoints_OmnibusJD(t *testing.T) {
	jd := NewCard(CardDesignDiamond, 11, false)
	// Without omnibus: J♦ = 0 points
	assert.Equal(t, 0, cardPoints(jd, false))
	// With omnibus: J♦ = -10 points
	assert.Equal(t, -10, cardPoints(jd, true))

	// Hearts still 1, Q♠ still 13 regardless of omnibus
	h3 := NewCard(CardDesignHeart, 3, false)
	assert.Equal(t, 1, cardPoints(h3, false))
	assert.Equal(t, 1, cardPoints(h3, true))

	qs := NewCard(CardDesignSpade, 12, false)
	assert.Equal(t, 13, cardPoints(qs, false))
	assert.Equal(t, 13, cardPoints(qs, true))

	// Other cards still 0
	c5 := NewCard(CardDesignClover, 5, false)
	assert.Equal(t, 0, cardPoints(c5, false))
	assert.Equal(t, 0, cardPoints(c5, true))
}

func TestIsPointCard_OmnibusJD(t *testing.T) {
	jd := NewCard(CardDesignDiamond, 11, false)
	assert.False(t, isPointCard(jd, false))
	assert.True(t, isPointCard(jd, true))

	// Hearts always point cards
	h3 := NewCard(CardDesignHeart, 3, false)
	assert.True(t, isPointCard(h3, false))
	assert.True(t, isPointCard(h3, true))

	// Q♠ always point card
	qs := NewCard(CardDesignSpade, 12, false)
	assert.True(t, isPointCard(qs, false))
	assert.True(t, isPointCard(qs, true))
}

func TestIsPenaltyCard(t *testing.T) {
	// Hearts are penalty
	assert.True(t, isPenaltyCard(NewCard(CardDesignHeart, 5, false)))
	// Q♠ is penalty
	assert.True(t, isPenaltyCard(NewCard(CardDesignSpade, 12, false)))
	// J♦ is NOT penalty (even though it's a point card under omnibus)
	assert.False(t, isPenaltyCard(NewCard(CardDesignDiamond, 11, false)))
	// Other cards are not penalty
	assert.False(t, isPenaltyCard(NewCard(CardDesignClover, 5, false)))
}

func TestHearts_cpuPlayHard_FollowWithNonLeadSuitCards(t *testing.T) {
	// Exercise the continue branch in cpuPlayHard follow-suit when
	// validIndices contains non-lead-suit cards (shouldn't happen in practice
	// but the code handles it)
	h := newInternalTestHearts()
	h.phase = HeartsPhasePlay
	h.currentPlayerIdx = 1
	h.config.CpuDifficulty = HeartsCpuDifficultyHard

	// Set up trick with clover lead
	h.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignClover, 10, false)},
	}

	// Give CPU lead-suit card + non-lead-suit card
	player := h.players[1]
	player.Reset()
	player.AddCard(NewCard(CardDesignClover, 5, false)) // lead suit
	player.AddCard(NewCard(CardDesignHeart, 8, false))  // non-lead suit (should be skipped)

	result := h.cpuPlayHard(1, []int{0, 1})
	// Should pick clover 5 (lead suit), skipping heart 8
	assert.Equal(t, 0, result)
}

func TestHearts_playerTookCard(t *testing.T) {
	h := newInternalTestHearts()
	jd := NewCard(CardDesignDiamond, 11, false)
	h.players[0].AddTrick([]*Card{jd})

	// Found: J♦ is in player 0's tricksTaken
	assert.True(t, h.playerTookCard(0, CardDesignDiamond, 11))
	// Not found: wrong value
	assert.False(t, h.playerTookCard(0, CardDesignDiamond, 10))
	// Not found: wrong design
	assert.False(t, h.playerTookCard(0, CardDesignSpade, 11))
	// Not found: player 1 has no tricks
	assert.False(t, h.playerTookCard(1, CardDesignDiamond, 11))
}
