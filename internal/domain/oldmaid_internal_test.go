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

// TestOldMaid_advancePastFinished_SkipsFinishedPlayer covers the advancePastFinished
// code path where currentTurn points to a finished player and must advance.
func TestOldMaid_advancePastFinished_SkipsFinishedPlayer(t *testing.T) {
	tc := NewTrumpCards(1)
	players := []*OldMaidPlayer{
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
	}
	om := NewOldMaid(tc, players)

	// player[0] is finished, player[1] is active
	om.players[0].SetIsFinished(true)
	om.players[1].AddCard(NewCard(CardDesignSpade, 2, false))
	om.currentTurn = 0

	om.advancePastFinished()

	assert.Equal(t, 1, om.currentTurn, "advancePastFinished should skip finished player[0] and land on player[1]")
}

// TestOldMaid_advancePastFinished_GameEndFlag covers the guard that prevents an
// infinite loop when all players are finished and gameEndFlag is true.
func TestOldMaid_advancePastFinished_GameEndFlag(t *testing.T) {
	tc := NewTrumpCards(1)
	players := []*OldMaidPlayer{
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
	}
	om := NewOldMaid(tc, players)

	// All players finished + gameEndFlag set
	for _, p := range players {
		p.SetIsFinished(true)
	}
	om.gameEndFlag = true
	om.currentTurn = 0

	// Must return without hanging (would infinite-loop without the guard)
	om.advancePastFinished()
	assert.Equal(t, 0, om.currentTurn, "advancePastFinished should not change currentTurn when gameEndFlag is true")
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
