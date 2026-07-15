package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeuceToSeven_SuggestExchange(t *testing.T) {
	newGame := func() (*DeuceToSeven, *DeuceToSevenPlayer) {
		players := []*DeuceToSevenPlayer{
			NewDeuceToSevenPlayer(true, DeuceToSevenStyleBalanced),
			NewDeuceToSevenPlayer(false, DeuceToSevenStyleBalanced),
		}
		g := NewDeuceToSeven(NewTrumpCards(0), players, DefaultDeuceToSevenConfig())
		return g, players[0]
	}

	t.Run("out-of-range index returns nil", func(t *testing.T) {
		g, _ := newGame()
		assert.Nil(t, g.SuggestExchange(-1))
		assert.Nil(t, g.SuggestExchange(99))
	})

	t.Run("made pat low returns nil (stand pat)", func(t *testing.T) {
		g, pl := newGame()
		ds := []int{CardDesignSpade, CardDesignHeart, CardDesignDiamond, CardDesignClover, CardDesignSpade}
		for i, v := range []int{7, 5, 4, 3, 2} {
			pl.AddCard(NewCard(ds[i], v, false))
		}
		assert.Nil(t, g.SuggestExchange(0))
	})

	t.Run("a flush is not treated as a pat low", func(t *testing.T) {
		g, pl := newGame()
		for _, v := range []int{7, 5, 4, 3, 2} { // all spades → flush (bad low)
			pl.AddCard(NewCard(CardDesignSpade, v, false))
		}
		assert.NotEmpty(t, g.SuggestExchange(0))
	})

	t.Run("weak hand returns a non-empty discard list", func(t *testing.T) {
		g, pl := newGame()
		for _, v := range []int{2, 2, 13, 12, 11} { // pair + high cards
			pl.AddCard(NewCard(CardDesignSpade, v, false))
		}
		assert.NotEmpty(t, g.SuggestExchange(0))
	})
}
