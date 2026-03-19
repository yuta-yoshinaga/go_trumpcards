package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvalBestFromOmaha_LessThan2HoleCards(t *testing.T) {
	t.Run("0 hole cards", func(t *testing.T) {
		community := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 13, false),
			NewCard(CardDesignClover, 5, false),
		}
		rank, best := evalBestFromOmaha(nil, community)
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
		rank, best := evalBestFromOmaha(hole, community)
		assert.Equal(t, PokerHandHighCard, rank)
		assert.Nil(t, best)
	})
}

func TestEvalBestFromOmaha_LessThan3CommunityCards(t *testing.T) {
	t.Run("0 community cards", func(t *testing.T) {
		hole := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 13, false),
		}
		rank, best := evalBestFromOmaha(hole, nil)
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
		rank, best := evalBestFromOmaha(hole, community)
		assert.Equal(t, PokerHandHighCard, rank)
		assert.Nil(t, best)
	})
}

func TestEvalBestFromOmaha_Normal(t *testing.T) {
	t.Run("omaha constraint: 4s in hand + royal on board = one pair", func(t *testing.T) {
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
		rank, best := evalBestFromOmaha(hole, community)
		// Must use exactly 2 hole cards (4,4) + 3 community → One Pair
		assert.Equal(t, PokerHandOnePair, rank)
		assert.Equal(t, 5, len(best))
	})
}
