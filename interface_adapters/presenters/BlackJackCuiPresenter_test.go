package presenters_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"
	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/presenters"

	"github.com/stretchr/testify/assert"
)

func TestBlackJackCuiPresenters_Method(t *testing.T) {
	tbp := presenters.NewBlackJackCuiPresenter()
	tc := entities.NewTrumpCards(0)
	player := entities.NewBlackJackPlayer()
	dealer := entities.NewBlackJackPlayer()
	bj := entities.NewBlackJack(tc, player, dealer)
	t.Run("success Output", func(t *testing.T) {
		bj.Reset()
		player.Reset()
		dealer.Reset()
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 10, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 2, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 10, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 11, false))
		assert.Equal(t, "----------\ndealer score \nCLOVER 2,\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\n", tbp.Output(bj))
	})
	t.Run("success Output", func(t *testing.T) {
		bj.Reset()
		player.Reset()
		dealer.Reset()
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 10, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 2, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 10, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 11, false))
		bj.PlayerStand()
		assert.Equal(t, "----------\ndealer score 22\nCLOVER 2,CLOVER 10,CLOVER 11\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\nIt is your loss.\n\n----------\n", tbp.Output(bj))
	})
	t.Run("success Output", func(t *testing.T) {
		bj.Reset()
		player.Reset()
		dealer.Reset()
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 1, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 10, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 1, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 10, false))
		bj.PlayerStand()
		assert.Equal(t, "----------\ndealer score 21\nCLOVER 1,CLOVER 10\n----------\nplayer score 21\nSPADE 1,SPADE 10\n----------\nIt is a draw.\n\n----------\n", tbp.Output(bj))
	})
	t.Run("success Output", func(t *testing.T) {
		bj.Reset()
		player.Reset()
		dealer.Reset()
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 1, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 10, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 2, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 10, false))
		bj.PlayerStand()
		assert.Equal(t, "----------\ndealer score 19\nCLOVER 2,CLOVER 10,CLOVER 7\n----------\nplayer score 21\nSPADE 1,SPADE 10\n----------\nYou are the winner.\n\n----------\n", tbp.Output(bj))
	})
	t.Run("success GetCardStr", func(t *testing.T) {
		assert.Equal(t, "SPADE 1", tbp.GetCardStr(entities.NewCard(entities.CardDesignSpade, 1, false)))
	})
	t.Run("success GetCardStr", func(t *testing.T) {
		assert.Equal(t, "CLOVER 1", tbp.GetCardStr(entities.NewCard(entities.CardDesignClover, 1, false)))
	})
	t.Run("success GetCardStr", func(t *testing.T) {
		assert.Equal(t, "HEART 1", tbp.GetCardStr(entities.NewCard(entities.CardDesignHeart, 1, false)))
	})
	t.Run("success GetCardStr", func(t *testing.T) {
		assert.Equal(t, "DIAMOND 1", tbp.GetCardStr(entities.NewCard(entities.CardDesignDiamond, 1, false)))
	})
	t.Run("success GetCardStr", func(t *testing.T) {
		assert.Equal(t, "Unsupported card 0", tbp.GetCardStr(entities.NewCard(entities.CardDesignJoker, entities.CardValueJoker, false)))
	})
}
