package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestLiveBestHandIndices(t *testing.T) {
	hole := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 8, false),
		domain.NewCard(domain.CardDesignHeart, 9, false),
	}
	board := []*domain.Card{
		domain.NewCard(domain.CardDesignClover, 10, false),
		domain.NewCard(domain.CardDesignDiamond, 11, false),
		domain.NewCard(domain.CardDesignSpade, 12, false),
	}

	t.Run("maps pointer-identical cards to their indices", func(t *testing.T) {
		holeIdx, boardIdx := liveBestHandIndices(hole, board, []*domain.Card{hole[1], board[0], board[2]})
		assert.Equal(t, []int{1}, holeIdx)
		assert.Equal(t, []int{0, 2}, boardIdx)
	})

	t.Run("nil best hand returns nil indices", func(t *testing.T) {
		holeIdx, boardIdx := liveBestHandIndices(hole, board, nil)
		assert.Nil(t, holeIdx)
		assert.Nil(t, boardIdx)
	})
}
