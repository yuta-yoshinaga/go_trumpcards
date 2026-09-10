//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestMatrimony() *Matrimony {
	c := NewDefaultMatrimony()
	c.Reset()
	c.stock = nil
	c.waste = nil
	c.history = nil
	c.moveCount = 0
	c.phase = MatrimonyPhasePlaying
	for i := range MatrimonyTableauCnt {
		c.tableau[i] = nil
	}
	for i := range MatrimonyFoundationCnt {
		c.foundation[i] = nil
	}
	return c
}

func TestMatrimonyResetUsesTwoDecksAndSixteenTableauSlots(t *testing.T) {
	c := NewDefaultMatrimony()
	c.Reset()
	assert.Len(t, c.GetTableau(), MatrimonyTableauCnt)
	assert.Equal(t, 88, c.GetStockCount())
	assert.Equal(t, MatrimonyTotalCards, c.GetStockCount()+MatrimonyTableauCnt)
}

func TestMatrimonyFoundationsStartCorrectly(t *testing.T) {
	c := newTestMatrimony()
	assert.True(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, matrimonyQueen, true), 0))
	assert.True(t, c.canPlaceOnFoundation(NewCard(CardDesignDiamond, matrimonyJack, true), 2))
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignHeart, matrimonyQueen, true), 0))
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, matrimonyJack, true), 2))
}

func TestMatrimonyFoundationWrapsBothDirections(t *testing.T) {
	c := newTestMatrimony()
	for value := matrimonyQueen; value >= matrimonyAce; value-- {
		c.foundation[0] = append(c.foundation[0], NewCard(CardDesignSpade, value, true))
	}
	assert.True(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, matrimonyKing, true), 0))
	for value := matrimonyJack; value <= matrimonyKing; value++ {
		c.foundation[2] = append(c.foundation[2], NewCard(CardDesignDiamond, value, true))
	}
	assert.True(t, c.canPlaceOnFoundation(NewCard(CardDesignDiamond, matrimonyAce, true), 2))
}

func TestMatrimonyMoveAndRedeal(t *testing.T) {
	c := newTestMatrimony()
	c.tableau[0] = NewCard(CardDesignSpade, matrimonyQueen, true)
	require.NoError(t, c.MoveTableauToFoundation(0))
	assert.Len(t, c.GetFoundation()[0], 1)
	c.waste = []*Card{NewCard(CardDesignHeart, 5, true)}
	for range MatrimonyMaxRedeals {
		c.stock = []*Card{NewCard(CardDesignClover, 4, true)}
		require.NoError(t, c.Draw())
		c.stock = nil
		require.NoError(t, c.Draw())
		assert.Equal(t, 1, c.GetStockCount())
	}
	c.stock = nil
	c.waste = []*Card{NewCard(CardDesignHeart, 5, true)}
	assert.Error(t, c.Draw())
}

func TestMatrimonyClearWhenAllCardsAreOnFoundations(t *testing.T) {
	c := newTestMatrimony()
	for i := range MatrimonyFoundationCnt {
		start := matrimonyFoundationStart[i]
		for n := 0; n < MatrimonyFoundationTarget; n++ {
			value := start.value
			for j := 0; j < n; j++ {
				if start.desc {
					value = matrimonyPreviousRank(value)
				} else {
					value = matrimonyNextRank(value)
				}
			}
			c.foundation[i] = append(c.foundation[i], NewCard(start.design, value, true))
		}
	}
	c.checkGameClear()
	assert.Equal(t, MatrimonyPhaseGameClear, c.GetPhase())
}
