package domain

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvalBestFromDramaha_LessThan2HoleCards(t *testing.T) {
	t.Run("0 hole cards", func(t *testing.T) {
		community := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 13, false),
			NewCard(CardDesignClover, 5, false),
		}
		rank, best := evalBestFromDramaha(nil, community)
		assert.Equal(t, PokerHandHighCard, rank)
		assert.Nil(t, best)
	})

	t.Run("1 hole card", func(t *testing.T) {
		hole := []*Card{NewCard(CardDesignSpade, 1, false)}
		community := []*Card{
			NewCard(CardDesignHeart, 13, false),
			NewCard(CardDesignClover, 5, false),
			NewCard(CardDesignDiamond, 3, false),
		}
		rank, best := evalBestFromDramaha(hole, community)
		assert.Equal(t, PokerHandHighCard, rank)
		assert.Nil(t, best)
	})
}

func TestEvalBestFromDramaha_LessThan3CommunityCards(t *testing.T) {
	t.Run("0 community cards", func(t *testing.T) {
		hole := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 13, false),
		}
		rank, best := evalBestFromDramaha(hole, nil)
		assert.Equal(t, PokerHandHighCard, rank)
		assert.Nil(t, best)
	})

	t.Run("2 community cards", func(t *testing.T) {
		hole := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 13, false),
		}
		community := []*Card{
			NewCard(CardDesignClover, 5, false),
			NewCard(CardDesignDiamond, 3, false),
		}
		rank, best := evalBestFromDramaha(hole, community)
		assert.Equal(t, PokerHandHighCard, rank)
		assert.Nil(t, best)
	})
}

func TestEvalBestFromDramaha_Normal(t *testing.T) {
	t.Run("dramaha constraint: 4s in hand + royal on board = one pair", func(t *testing.T) {
		hole := []*Card{
			NewCard(CardDesignClover, 4, false),
			NewCard(CardDesignDiamond, 4, false),
			NewCard(CardDesignHeart, 4, false),
			NewCard(CardDesignSpade, 4, false),
		}
		community := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignSpade, 13, false),
			NewCard(CardDesignSpade, 12, false),
			NewCard(CardDesignSpade, 11, false),
			NewCard(CardDesignSpade, 10, false),
		}
		rank, best := evalBestFromDramaha(hole, community)
		// Must use exactly 2 hole cards (4,4) + 3 community → One Pair
		assert.Equal(t, PokerHandOnePair, rank)
		assert.Equal(t, 5, len(best))
	})
}

// TestCalcDramahaEquity_UsesTheTableHoleCardWidth pins the width the game
// actually plays at. Dramaha deals five, so the opponents the simulation deals
// against must hold five too -- modelling them on four hole cards
// systematically understates how often they beat the human.
func TestCalcDramahaEquity_UsesTheTableHoleCardWidth(t *testing.T) {
	humanCards := []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignHeart, 1, false),
		NewCard(CardDesignClover, 13, false),
		NewCard(CardDesignDiamond, 13, false),
		NewCard(CardDesignSpade, 12, false),
	}
	community := []*Card{
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignDiamond, 7, false),
		NewCard(CardDesignSpade, 9, false),
	}

	withFive := calcDramahaEquityWithHoleCount(humanCards, community, 3, 3000, rand.New(rand.NewSource(7)), DramahaHoleCards)
	withFour := calcDramahaEquityWithHoleCount(humanCards, community, 3, 3000, rand.New(rand.NewSource(7)), 4)

	assert.Greater(t, withFive.Equity, 0.0)
	assert.Less(t, withFive.Equity, 1.0)
	assert.NotEqual(t, withFour.Equity, withFive.Equity,
		"the hole-card width has to reach the simulation; if it did not, both runs would be identical")
	assert.Less(t, withFive.Equity, withFour.Equity,
		"five-card opponents make more hands, so the human's equity must fall")
}

// TestDramahaGetEquity_ModelsFiveCardOpponents drives the same rule from the
// game rather than from the helper: the table's own hole-card width has to be
// what reaches the simulation.
func TestDramahaGetEquity_ModelsFiveCardOpponents(t *testing.T) {
	o := newInternalTestDramaha()
	o.phase = DramahaPhaseFlop
	o.communityCards = []*Card{
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignDiamond, 7, false),
		NewCard(CardDesignSpade, 9, false),
	}
	o.players[0].Reset()
	for _, c := range []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignHeart, 1, false),
		NewCard(CardDesignClover, 13, false),
		NewCard(CardDesignDiamond, 13, false),
		NewCard(CardDesignSpade, 12, false),
	} {
		o.players[0].AddCard(c)
	}

	require.Equal(t, DramahaHoleCards, o.holeCardCount(),
		"GetEquity forwards holeCardCount(); this is the value it must forward")

	result := o.GetEquity()
	require.NotNil(t, result)
	assert.Greater(t, result.Equity, 0.0)
	assert.Less(t, result.Equity, 1.0)
}
