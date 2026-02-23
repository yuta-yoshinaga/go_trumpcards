package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestOldMaid_cpuSelectCardIdx_AllBranches covers all branches of cpuSelectCardIdx:
// - size <= 1 returns 0
// - 30% chance of edge selection (first or last card)
// - 70% chance of random selection
func TestOldMaid_cpuSelectCardIdx_AllBranches(t *testing.T) {
	// Test size <= 1 branch
	assert.Equal(t, 0, cpuSelectCardIdx(0))
	assert.Equal(t, 0, cpuSelectCardIdx(1))

	// Run many times with size > 1 to statistically hit all branches
	size := 10
	hitFirst := false  // return 0 from edge select (line 260)
	hitLast := false   // return size-1 from edge select (line 262)
	hitRandom := false // return rand.Intn(size) (line 264)

	for i := 0; i < 5000; i++ {
		idx := cpuSelectCardIdx(size)
		if idx == 0 {
			hitFirst = true
		} else if idx == size-1 {
			hitLast = true
		} else {
			hitRandom = true
		}
		if hitFirst && hitLast && hitRandom {
			break
		}
	}
	assert.True(t, hitFirst, "should hit first card (edge select)")
	assert.True(t, hitLast, "should hit last card (edge select)")
	assert.True(t, hitRandom, "should hit random middle card")
}

// TestOldMaid_drawCard_GameEndFlag covers drawCard returning nil when gameEndFlag is true.
func TestOldMaid_drawCard_GameEndFlag(t *testing.T) {
	tc := NewTrumpCards(1)
	players := []*OldMaidPlayer{
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
	}
	om := NewOldMaid(tc, players)

	players[0].AddCard(NewCard(CardDesignSpade, 2, false))
	players[1].AddCard(NewCard(CardDesignClover, 3, false))

	// Set gameEndFlag to true, then call drawCard directly
	om.gameEndFlag = true
	result := om.drawCard(0, 0)
	assert.Nil(t, result, "drawCard should return nil when gameEndFlag is true")
}

// TestOldMaid_drawCard_FinishedPlayerReturnsNil tests drawCard when the calling player is finished.
func TestOldMaid_drawCard_FinishedPlayerReturnsNil(t *testing.T) {
	tc := NewTrumpCards(1)
	players := []*OldMaidPlayer{
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
	}
	om := NewOldMaid(tc, players)

	players[0].AddCard(NewCard(CardDesignSpade, 2, false))
	players[1].AddCard(NewCard(CardDesignClover, 3, false))
	players[0].SetIsFinished(true)

	result := om.drawCard(0, 0)
	assert.Nil(t, result, "drawCard should return nil when player is finished")
}

// TestOldMaid_currentTurn_DirectAccess verifies direct field access for currentTurn.
func TestOldMaid_currentTurn_DirectAccess(t *testing.T) {
	tc := NewTrumpCards(1)
	players := []*OldMaidPlayer{
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
	}
	om := NewOldMaid(tc, players)

	assert.Equal(t, 0, om.GetCurrentTurn())
	om.currentTurn = 2
	assert.Equal(t, 2, om.GetCurrentTurn())
}

// TestOldMaid_advanceTurn_GameEndFlag covers advanceTurn returning early when gameEndFlag is true.
func TestOldMaid_advanceTurn_GameEndFlag(t *testing.T) {
	tc := NewTrumpCards(1)
	players := []*OldMaidPlayer{
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
	}
	om := NewOldMaid(tc, players)

	players[0].AddCard(NewCard(CardDesignSpade, 2, false))
	players[1].AddCard(NewCard(CardDesignClover, 3, false))

	// Set gameEndFlag to true, then call advanceTurn directly
	om.gameEndFlag = true
	turnBefore := om.GetCurrentTurn()
	om.advanceTurn()
	assert.Equal(t, turnBefore, om.GetCurrentTurn(),
		"advanceTurn should not change currentTurn when gameEndFlag is true")
}

// TestOldMaid_Reset_PlayerZeroCardsAfterDiscardPairs covers lines 97-99 and 107-108
// in Reset: when a player has 0 cards after DiscardPairs, they are set as finished,
// and the currentTurn advancement loop skips finished players.
//
// Strategy: Replace the TrumpCards deck with exactly 5 cards of the same value.
// With 5 cards dealt round-robin to 4 players, player at index 0 receives 2 cards
// (at positions 0 and 4) while players at indices 1-3 receive 1 card each.
// Since all cards share the same value, DiscardPairs matches player 0's 2 cards
// as a pair, leaving 0 cards → SetIsFinished(true).
// The remaining 3 players each keep 1 card (no pair possible) → active count = 3
// → gameEndFlag stays false → the currentTurn advancement loop executes.
func TestOldMaid_Reset_PlayerZeroCardsAfterDiscardPairs(t *testing.T) {
	tc := NewTrumpCards(0)
	// Replace the deck with exactly 5 cards, all value 1.
	// Shuffle order is irrelevant because all cards have the same value.
	tc.deck = []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignClover, 1, false),
		NewCard(CardDesignHeart, 1, false),
		NewCard(CardDesignDiamond, 1, false),
		NewCard(CardDesignSpade, 1, false),
	}
	tc.deckCnt = 5
	tc.deckDrawCnt = 0

	players := []*OldMaidPlayer{
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
	}
	om := NewOldMaid(tc, players)
	om.Reset()

	// After Reset:
	// - Player at index 0 received 2 same-value cards → paired → 0 cards → finished
	// - Players at indices 1-3 each received 1 card → no pair → 1 card → not finished
	// - Active players = 3, so gameEndFlag is false
	// - currentTurn was 0 (finished), loop advanced to next non-finished player

	// Verify at least one player has 0 cards and is finished (lines 97-99)
	foundFinished := false
	for i := 0; i < OldMaidPlayerCnt; i++ {
		p := om.players[i]
		if p.GetCardsSize() == 0 && p.GetIsFinished() {
			foundFinished = true
		}
	}
	assert.True(t, foundFinished,
		"player at index 0 should have 0 cards and be finished after DiscardPairs")

	// Verify game has not ended (3 active players remain)
	assert.False(t, om.gameEndFlag,
		"gameEndFlag should be false with 3 active players")

	// Verify currentTurn was advanced past the finished player (lines 107-108)
	currentPlayer := om.players[om.currentTurn]
	assert.False(t, currentPlayer.GetIsFinished(),
		"currentTurn should point to a non-finished player after Reset")
}

// TestOldMaid_drawCard_NoActiveTargetWith5Players covers drawCard returning nil
// when getNextActivePlayer returns -1 (no active target).
func TestOldMaid_drawCard_NoActiveTargetWith5Players(t *testing.T) {
	tc := NewTrumpCards(1)
	players := []*OldMaidPlayer{
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false), // Extra player at index 4
	}
	om := NewOldMaid(tc, players)

	// Player 4 is not finished and has a card
	players[4].AddCard(NewCard(CardDesignSpade, 2, false))
	// Players 0-3 are all finished
	players[0].SetIsFinished(true)
	players[1].SetIsFinished(true)
	players[2].SetIsFinished(true)
	players[3].SetIsFinished(true)

	// drawCard(4, 0): player 4 is not finished (passes check),
	// getNextActivePlayer(4) checks indices 1,2,3,0 - all finished → returns -1.
	// targetIdx = -1 < 0 → return nil.
	result := om.drawCard(4, 0)
	assert.Nil(t, result, "drawCard should return nil when no active target is found")
}
