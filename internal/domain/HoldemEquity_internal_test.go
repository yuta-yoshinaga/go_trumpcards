package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHoldem_GetEquity(t *testing.T) {
	setup := func(phase int) *Holdem {
		h := newTestHoldem()
		for _, p := range h.players {
			p.SetChips(1000)
		}
		h.phase = phase
		h.currentTurn = 0
		// Give human cards
		h.players[0].Reset()
		h.players[0].AddCard(NewCard(CardDesignSpade, 1, false))
		h.players[0].AddCard(NewCard(CardDesignHeart, 1, false))
		return h
	}

	t.Run("returns nil for init phase", func(t *testing.T) {
		h := setup(HoldemPhaseInit)
		assert.Nil(t, h.GetEquity())
	})

	t.Run("returns nil for showdown phase", func(t *testing.T) {
		h := setup(HoldemPhaseShowdown)
		assert.Nil(t, h.GetEquity())
	})

	t.Run("returns nil for end phase", func(t *testing.T) {
		h := setup(HoldemPhaseEnd)
		assert.Nil(t, h.GetEquity())
	})

	t.Run("returns nil for rebuy phase", func(t *testing.T) {
		h := setup(HoldemPhaseRebuy)
		assert.Nil(t, h.GetEquity())
	})

	t.Run("returns nil when human folded", func(t *testing.T) {
		h := setup(HoldemPhaseFlop)
		h.players[0].SetFolded(true)
		assert.Nil(t, h.GetEquity())
	})

	t.Run("returns result during preflop", func(t *testing.T) {
		h := setup(HoldemPhasePreFlop)
		result := h.GetEquity()
		assert.NotNil(t, result)
		assert.Greater(t, result.Equity, 0.0)
		assert.Len(t, result.HandOdds, len(PokerHandNames))
	})

	t.Run("returns result during flop", func(t *testing.T) {
		h := setup(HoldemPhaseFlop)
		h.communityCards = []*Card{
			NewCard(CardDesignClover, 2, false),
			NewCard(CardDesignDiamond, 7, false),
			NewCard(CardDesignSpade, 9, false),
		}
		result := h.GetEquity()
		assert.NotNil(t, result)
		assert.Greater(t, result.Equity, 0.0)
	})

	t.Run("returns result during turn", func(t *testing.T) {
		h := setup(HoldemPhaseTurn)
		h.communityCards = []*Card{
			NewCard(CardDesignClover, 2, false),
			NewCard(CardDesignDiamond, 7, false),
			NewCard(CardDesignSpade, 9, false),
			NewCard(CardDesignHeart, 4, false),
		}
		result := h.GetEquity()
		assert.NotNil(t, result)
	})

	t.Run("returns result during river", func(t *testing.T) {
		h := setup(HoldemPhaseRiver)
		h.communityCards = []*Card{
			NewCard(CardDesignClover, 2, false),
			NewCard(CardDesignDiamond, 7, false),
			NewCard(CardDesignSpade, 9, false),
			NewCard(CardDesignHeart, 4, false),
			NewCard(CardDesignClover, 6, false),
		}
		result := h.GetEquity()
		assert.NotNil(t, result)
	})
}

func TestHoldem_GetPotOdds(t *testing.T) {
	setup := func(phase int) *Holdem {
		h := newTestHoldem()
		for _, p := range h.players {
			p.SetChips(1000)
		}
		h.phase = phase
		h.pot = 100
		h.lastBet = 50
		h.players[0].SetCurrentBet(0)
		return h
	}

	t.Run("returns 0 for init phase", func(t *testing.T) {
		h := setup(HoldemPhaseInit)
		assert.Equal(t, 0.0, h.GetPotOdds())
	})

	t.Run("returns 0 for showdown phase", func(t *testing.T) {
		h := setup(HoldemPhaseShowdown)
		assert.Equal(t, 0.0, h.GetPotOdds())
	})

	t.Run("returns 0 for end phase", func(t *testing.T) {
		h := setup(HoldemPhaseEnd)
		assert.Equal(t, 0.0, h.GetPotOdds())
	})

	t.Run("returns correct pot odds during active phase", func(t *testing.T) {
		h := setup(HoldemPhasePreFlop)
		result := h.GetPotOdds()
		// pot=100, callAmount=50-0=50, potOdds = 50/(100+50)*100 = 33.33
		assert.InDelta(t, 33.33, result, 0.01)
	})

	t.Run("returns 0 when no outstanding bet", func(t *testing.T) {
		h := setup(HoldemPhaseFlop)
		h.lastBet = 0
		assert.Equal(t, 0.0, h.GetPotOdds())
	})

	t.Run("accounts for human current bet", func(t *testing.T) {
		h := setup(HoldemPhasePreFlop)
		h.players[0].SetCurrentBet(20)
		// callAmount = 50 - 20 = 30
		// potOdds = 30 / (100+30) * 100 = 23.08
		result := h.GetPotOdds()
		assert.InDelta(t, 23.08, result, 0.01)
	})
}
