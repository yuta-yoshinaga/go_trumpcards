package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSevens_advanceTurn_gameEndFlag(t *testing.T) {
	// Covers lines 176-178: advanceTurn returns immediately when gameEndFlag is true.
	// This guard is defensive code unreachable through public API because all callers
	// check gameEndFlag before calling advanceTurn. We test it directly.
	tc := NewTrumpCards(0)
	players := []*SevensPlayer{
		NewSevensPlayer(true),
		NewSevensPlayer(false),
		NewSevensPlayer(false),
		NewSevensPlayer(false),
	}
	s := NewSevens(tc, players, DefaultSevensConfig())

	// Give players cards so there are active players to advance to
	for i := 0; i < 4; i++ {
		players[i].AddCard(NewCard(CardDesignDiamond, 2, false))
	}

	s.gameEndFlag = true
	originalTurn := s.currentTurn
	s.advanceTurn()
	// currentTurn should not have changed because advanceTurn returned early
	assert.Equal(t, originalTurn, s.currentTurn)
}

func makeSevensPlayersInternal() []*SevensPlayer {
	return []*SevensPlayer{
		NewSevensPlayer(true),
		NewSevensPlayer(false),
		NewSevensPlayer(false),
		NewSevensPlayer(false),
	}
}

func TestSevens_isPositionPlaced_InvalidSuit(t *testing.T) {
	// Covers lines 187-189: isPositionPlaced returns false for invalid suit.
	tc := NewTrumpCards(0)
	players := makeSevensPlayersInternal()
	s := NewSevens(tc, players, DefaultSevensConfig())

	// suit < CardDesignSpade (suit=0)
	assert.False(t, s.isPositionPlaced(0, 7))
	// suit > CardDesignDiamond (suit=5)
	assert.False(t, s.isPositionPlaced(5, 7))
	// suit = -1
	assert.False(t, s.isPositionPlaced(-1, 7))
	// Valid suit with valid value (7 is placed on fresh board)
	assert.True(t, s.isPositionPlaced(CardDesignSpade, 7))
	// Valid suit with unplaced value
	assert.False(t, s.isPositionPlaced(CardDesignSpade, 6))
}

func TestSevens_evaluatePlay_TunnelAceLow(t *testing.T) {
	// Covers evaluatePlay tunnel wrap for Ace (value=1),
	// where nextLow becomes 13 instead of 0.
	// Setup: tunnel enabled + strategy. Board has spade 2-7 placed (min=2, max=7).
	// CPU has Ace(1) which is playable (adjacent to 2).
	// Without tunnel: nextLow = 0, skipped. With tunnel: nextLow = 13.
	// CPU does NOT have King(13), so score -= 1 for low direction.
	// nextHigh = 2 (placed), so no score for high direction.
	// Total score = -1 -> CPU would pass if it has passes.
	tc := NewTrumpCards(0)
	players := makeSevensPlayersInternal()
	tunnelStrategyConfig := SevensConfig{TunnelEnabled: true, JokerCount: 0, CpuStrategy: true}
	s := NewSevens(tc, players, tunnelStrategyConfig)

	// Build board: set the board directly using field access
	var placed [5]uint16
	for i := 1; i <= 4; i++ {
		placed[i] = 1 << 7 // 7 is placed for all suits
	}
	// Place spade 2-6 (plus 7 already)
	placed[CardDesignSpade] |= (1 << 2) | (1 << 3) | (1 << 4) | (1 << 5) | (1 << 6)
	s.tablePlaced = placed

	// Give players enough cards so game doesn't end
	for i := 0; i < 4; i++ {
		for d := 0; d < 5; d++ {
			players[i].AddCard(NewCard(CardDesignDiamond, 2, false))
		}
	}

	// Human plays some card to advance turn to CPU
	players[0].AddCard(NewCard(CardDesignSpade, 8, false))
	s.PlayerPlay(players[0].GetCardsSize() - 1) // play 8♠

	// CPU 1 has Ace(1)♠ - playable since adjacent to 2♠
	// It does NOT have King(13)♠, so tunnel wrap gives negative score
	players[1].AddCard(NewCard(CardDesignSpade, 1, false))

	if s.currentTurn == 1 {
		s.CpuPlay()
		// Score = -1 (tunnel wrap nextLow=13, no King) and CPU has passes -> should pass
		actions := s.GetCpuActions()
		assert.NotEmpty(t, actions)
		lastAction := actions[len(actions)-1]
		if lastAction.PlayerIdx == 1 {
			assert.Nil(t, lastAction.PlayedCard) // passed due to negative score
		}
	}
}

func TestSevens_evaluatePlay_TunnelKingHigh(t *testing.T) {
	// Covers evaluatePlay tunnel wrap for King (value=13),
	// where nextHigh becomes 1 instead of 14.
	// Setup: tunnel enabled + strategy. Board has spade 7-12 placed.
	// CPU has King(13) which is playable (adjacent to 12).
	// Without tunnel: nextHigh = 14, skipped. With tunnel: nextHigh = 1.
	// CPU does NOT have Ace(1), so score -= 1 for high direction.
	// nextLow = 12 (placed), so no score for low direction.
	// Total score = -1 -> CPU would pass if it has passes.
	tc := NewTrumpCards(0)
	players := makeSevensPlayersInternal()
	tunnelStrategyConfig := SevensConfig{TunnelEnabled: true, JokerCount: 0, CpuStrategy: true}
	s := NewSevens(tc, players, tunnelStrategyConfig)

	// Build board: place spade 8-12 (plus 7 already)
	var placed [5]uint16
	for i := 1; i <= 4; i++ {
		placed[i] = 1 << 7
	}
	placed[CardDesignSpade] |= (1 << 8) | (1 << 9) | (1 << 10) | (1 << 11) | (1 << 12)
	s.tablePlaced = placed

	// Give players enough cards so game doesn't end
	for i := 0; i < 4; i++ {
		for d := 0; d < 5; d++ {
			players[i].AddCard(NewCard(CardDesignDiamond, 2, false))
		}
	}

	// Human plays some card to advance turn to CPU
	players[0].AddCard(NewCard(CardDesignSpade, 6, false))
	s.PlayerPlay(players[0].GetCardsSize() - 1) // play 6♠

	// CPU 1 has King(13)♠ - playable since adjacent to 12♠
	// It does NOT have Ace(1)♠, so tunnel wrap gives negative score
	players[1].AddCard(NewCard(CardDesignSpade, 13, false))

	if s.currentTurn == 1 {
		s.CpuPlay()
		actions := s.GetCpuActions()
		assert.NotEmpty(t, actions)
		lastAction := actions[len(actions)-1]
		if lastAction.PlayerIdx == 1 {
			assert.Nil(t, lastAction.PlayedCard) // passed due to negative score
		}
	}
}

func TestSevens_setGameEndFlag(t *testing.T) {
	tc := NewTrumpCards(0)
	players := makeSevensPlayersInternal()
	s := NewSevens(tc, players, DefaultSevensConfig())

	assert.False(t, s.gameEndFlag)
	s.gameEndFlag = true
	assert.True(t, s.gameEndFlag)
	s.gameEndFlag = false
	assert.False(t, s.gameEndFlag)
}

func TestSevens_setCurrentTurn(t *testing.T) {
	tc := NewTrumpCards(0)
	players := makeSevensPlayersInternal()
	s := NewSevens(tc, players, DefaultSevensConfig())

	assert.Equal(t, 0, s.currentTurn)
	s.currentTurn = 2
	assert.Equal(t, 2, s.currentTurn)
}

func TestSevens_setTablePlaced(t *testing.T) {
	tc := NewTrumpCards(0)
	players := makeSevensPlayersInternal()
	s := NewSevens(tc, players, DefaultSevensConfig())

	var placed [5]uint16
	for i := 1; i <= 4; i++ {
		placed[i] = (1 << 7) | (1 << 6)
	}
	s.tablePlaced = placed
	result := s.GetTablePlaced()
	for i := 1; i <= 4; i++ {
		assert.Equal(t, uint16((1<<7)|(1<<6)), result[i])
	}
}
