package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBettingPlayerBase_HandRank(t *testing.T) {
	b := &bettingPlayerBase{}
	assert.Equal(t, 0, b.GetHandRank())

	b.SetHandRank(PokerHandRoyalFlush)
	assert.Equal(t, PokerHandRoyalFlush, b.GetHandRank())
}

func TestBettingPlayerBase_Folded(t *testing.T) {
	b := &bettingPlayerBase{}
	assert.False(t, b.GetFolded())

	b.SetFolded(true)
	assert.True(t, b.GetFolded())

	b.SetFolded(false)
	assert.False(t, b.GetFolded())
}

func TestBettingPlayerBase_AllIn(t *testing.T) {
	b := &bettingPlayerBase{}
	assert.False(t, b.GetAllIn())

	b.SetAllIn(true)
	assert.True(t, b.GetAllIn())

	b.SetAllIn(false)
	assert.False(t, b.GetAllIn())
}

func TestBettingPlayerBase_CurrentBet(t *testing.T) {
	b := &bettingPlayerBase{}
	assert.Equal(t, 0, b.GetCurrentBet())

	b.SetCurrentBet(100)
	assert.Equal(t, 100, b.GetCurrentBet())
}
