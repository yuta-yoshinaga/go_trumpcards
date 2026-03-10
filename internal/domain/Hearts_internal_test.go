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
	assert.Equal(t, "Player -1", h.playerName(-1))
	assert.Equal(t, "Player 10", h.playerName(10))
}

func TestHearts_startPlayPhase_NotFirstTrick(t *testing.T) {
	h := newInternalTestHearts()
	h.trickNumber = 5
	h.leadPlayerIdx = 2
	h.startPlayPhase()
	assert.Equal(t, 2, h.currentPlayerIdx)
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
	h.currentTrick = []*HeartsTrickCard{
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
