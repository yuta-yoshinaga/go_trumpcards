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
