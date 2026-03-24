//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvalBestFromShortDeck_LessThan5Cards(t *testing.T) {
	t.Run("0 cards", func(t *testing.T) {
		rank, best := evalBestFromShortDeck(nil)
		assert.Equal(t, ShortDeckHandHighCard, rank)
		assert.Nil(t, best)
	})

	t.Run("3 cards", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 13, false),
			NewCard(CardDesignClover, 8, false),
		}
		rank, best := evalBestFromShortDeck(cards)
		assert.Equal(t, ShortDeckHandHighCard, rank)
		assert.Nil(t, best)
	})
}

func TestShortDeck_GetEquity(t *testing.T) {
	setup := func(phase int) *ShortDeck {
		players := []*ShortDeckPlayer{
			NewShortDeckPlayer(true, HoldemStyleTAG),
			NewShortDeckPlayer(false, HoldemStyleTAG),
			NewShortDeckPlayer(false, HoldemStyleLAP),
			NewShortDeckPlayer(false, HoldemStyleLAG),
		}
		cfg := DefaultShortDeckConfig()
		tc := NewTrumpCardsShortDeck()
		sd := NewShortDeck(tc, players, cfg)
		for _, p := range sd.players {
			p.SetChips(1000)
		}
		sd.phase = phase
		sd.currentTurn = 0
		// Give human cards
		sd.players[0].Reset()
		sd.players[0].AddCard(NewCard(CardDesignSpade, 1, false))
		sd.players[0].AddCard(NewCard(CardDesignHeart, 1, false))
		return sd
	}

	t.Run("returns nil for init phase", func(t *testing.T) {
		sd := setup(ShortDeckPhaseInit)
		assert.Nil(t, sd.GetEquity())
	})

	t.Run("returns nil for showdown phase", func(t *testing.T) {
		sd := setup(ShortDeckPhaseShowdown)
		assert.Nil(t, sd.GetEquity())
	})

	t.Run("returns nil for end phase", func(t *testing.T) {
		sd := setup(ShortDeckPhaseEnd)
		assert.Nil(t, sd.GetEquity())
	})

	t.Run("returns nil for rebuy phase", func(t *testing.T) {
		sd := setup(ShortDeckPhaseRebuy)
		assert.Nil(t, sd.GetEquity())
	})

	t.Run("returns nil when human folded", func(t *testing.T) {
		sd := setup(ShortDeckPhaseFlop)
		sd.players[0].SetFolded(true)
		assert.Nil(t, sd.GetEquity())
	})

	t.Run("returns nil when no human player", func(t *testing.T) {
		players := []*ShortDeckPlayer{
			NewShortDeckPlayer(false, HoldemStyleTAG),
			NewShortDeckPlayer(false, HoldemStyleLAP),
		}
		cfg := DefaultShortDeckConfig()
		tc := NewTrumpCardsShortDeck()
		sd := NewShortDeck(tc, players, cfg)
		sd.phase = ShortDeckPhasePreFlop
		for _, p := range sd.players {
			p.SetChips(1000)
			p.AddCard(NewCard(CardDesignSpade, 1, false))
			p.AddCard(NewCard(CardDesignHeart, 13, false))
		}
		assert.Nil(t, sd.GetEquity())
	})

	t.Run("returns result during preflop", func(t *testing.T) {
		sd := setup(ShortDeckPhasePreFlop)
		result := sd.GetEquity()
		assert.NotNil(t, result)
		assert.Greater(t, result.Equity, 0.0)
		assert.Len(t, result.HandOdds, len(ShortDeckHandNames))
	})

	t.Run("returns result during flop", func(t *testing.T) {
		sd := setup(ShortDeckPhaseFlop)
		sd.communityCards = []*Card{
			NewCard(CardDesignClover, 6, false),
			NewCard(CardDesignDiamond, 7, false),
			NewCard(CardDesignSpade, 9, false),
		}
		result := sd.GetEquity()
		assert.NotNil(t, result)
		assert.Greater(t, result.Equity, 0.0)
	})

	t.Run("returns result during turn", func(t *testing.T) {
		sd := setup(ShortDeckPhaseTurn)
		sd.communityCards = []*Card{
			NewCard(CardDesignClover, 6, false),
			NewCard(CardDesignDiamond, 7, false),
			NewCard(CardDesignSpade, 9, false),
			NewCard(CardDesignHeart, 8, false),
		}
		result := sd.GetEquity()
		assert.NotNil(t, result)
	})

	t.Run("returns result during river", func(t *testing.T) {
		sd := setup(ShortDeckPhaseRiver)
		sd.communityCards = []*Card{
			NewCard(CardDesignClover, 6, false),
			NewCard(CardDesignDiamond, 7, false),
			NewCard(CardDesignSpade, 9, false),
			NewCard(CardDesignHeart, 8, false),
			NewCard(CardDesignClover, 10, false),
		}
		result := sd.GetEquity()
		assert.NotNil(t, result)
	})
}

func TestShortDeck_GetPotOdds(t *testing.T) {
	setup := func(phase int) *ShortDeck {
		players := []*ShortDeckPlayer{
			NewShortDeckPlayer(true, HoldemStyleTAG),
			NewShortDeckPlayer(false, HoldemStyleTAG),
			NewShortDeckPlayer(false, HoldemStyleLAP),
			NewShortDeckPlayer(false, HoldemStyleLAG),
		}
		cfg := DefaultShortDeckConfig()
		tc := NewTrumpCardsShortDeck()
		sd := NewShortDeck(tc, players, cfg)
		for _, p := range sd.players {
			p.SetChips(1000)
		}
		sd.phase = phase
		sd.pot = 100
		sd.lastBet = 50
		sd.players[0].SetCurrentBet(0)
		return sd
	}

	t.Run("returns 0 for init phase", func(t *testing.T) {
		sd := setup(ShortDeckPhaseInit)
		assert.Equal(t, 0.0, sd.GetPotOdds())
	})

	t.Run("returns 0 for showdown phase", func(t *testing.T) {
		sd := setup(ShortDeckPhaseShowdown)
		assert.Equal(t, 0.0, sd.GetPotOdds())
	})

	t.Run("returns 0 for end phase", func(t *testing.T) {
		sd := setup(ShortDeckPhaseEnd)
		assert.Equal(t, 0.0, sd.GetPotOdds())
	})

	t.Run("returns correct pot odds during active phase", func(t *testing.T) {
		sd := setup(ShortDeckPhasePreFlop)
		result := sd.GetPotOdds()
		// pot=100, callAmount=50-0=50, potOdds = 50/(100+50)*100 = 33.33
		assert.InDelta(t, 33.33, result, 0.01)
	})

	t.Run("returns 0 when no outstanding bet", func(t *testing.T) {
		sd := setup(ShortDeckPhaseFlop)
		sd.lastBet = 0
		assert.Equal(t, 0.0, sd.GetPotOdds())
	})

	t.Run("accounts for human current bet", func(t *testing.T) {
		sd := setup(ShortDeckPhasePreFlop)
		sd.players[0].SetCurrentBet(20)
		// callAmount = 50 - 20 = 30
		// potOdds = 30 / (100+30) * 100 = 23.08
		result := sd.GetPotOdds()
		assert.InDelta(t, 23.08, result, 0.01)
	})

	t.Run("returns 0 when humanCurrentBet exceeds lastBet", func(t *testing.T) {
		sd := setup(ShortDeckPhasePreFlop)
		sd.lastBet = 10
		sd.players[0].SetCurrentBet(50)
		// callAmount = 10 - 50 = -40 -> clamped to 0
		assert.Equal(t, 0.0, sd.GetPotOdds())
	})
}
