package presenters_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"
	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/presenters"

	"github.com/stretchr/testify/assert"
)

func TestPokerCuiPresenter_Method(t *testing.T) {
	tpp := presenters.NewPokerCuiPresenter()
	tc := entities.NewTrumpCards(0)
	player := entities.NewPokerPlayer()
	dealer := entities.NewPokerPlayer()
	tp := entities.NewPoker(tc, player, dealer)

	t.Run("success Output deal phase", func(t *testing.T) {
		tp.Reset()
		player.Reset()
		dealer.Reset()
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		player.AddCard(entities.NewCard(entities.CardDesignClover, 6, false))
		player.AddCard(entities.NewCard(entities.CardDesignHeart, 7, false))
		player.AddCard(entities.NewCard(entities.CardDesignDiamond, 8, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 9, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignSpade, 8, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 8, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignHeart, 8, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignDiamond, 3, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))
		expected := "----------\nplayer hand\n[0]SPADE 5,[1]CLOVER 6,[2]HEART 7,[3]DIAMOND 8,[4]SPADE 9\n----------\ndealer hand\n----------\n"
		assert.Equal(t, expected, tpp.Output(tp))
	})

	t.Run("success Output end phase player wins", func(t *testing.T) {
		tp.Reset()
		player.Reset()
		dealer.Reset()
		// player: Straight Flush (3-4-5-6-7 same suit)
		player.AddCard(entities.NewCard(entities.CardDesignHeart, 3, false))
		player.AddCard(entities.NewCard(entities.CardDesignHeart, 4, false))
		player.AddCard(entities.NewCard(entities.CardDesignHeart, 5, false))
		player.AddCard(entities.NewCard(entities.CardDesignHeart, 6, false))
		player.AddCard(entities.NewCard(entities.CardDesignHeart, 7, false))
		// dealer: Full House (rank >= TwoPair → no exchange when PlayerStand called)
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 8, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignSpade, 8, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignDiamond, 8, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 3, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))
		tp.PlayerStand()
		expected := "----------\nplayer hand [Straight Flush]\n[0]HEART 3,[1]HEART 4,[2]HEART 5,[3]HEART 6,[4]HEART 7\n----------\ndealer hand [Full House]\nCLOVER 8,SPADE 8,DIAMOND 8,CLOVER 3,SPADE 3\n----------\nYou are the winner.\n----------\n"
		assert.Equal(t, expected, tpp.Output(tp))
	})

	t.Run("success Output end phase player loses", func(t *testing.T) {
		tp.Reset()
		player.Reset()
		dealer.Reset()
		// player: High Card
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		player.AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		player.AddCard(entities.NewCard(entities.CardDesignHeart, 7, false))
		player.AddCard(entities.NewCard(entities.CardDesignDiamond, 9, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		// dealer: Full House (rank >= TwoPair → no exchange)
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 8, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignHeart, 8, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignDiamond, 8, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 3, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignHeart, 3, false))
		tp.PlayerStand()
		expected := "----------\nplayer hand [High Card]\n[0]SPADE 2,[1]CLOVER 5,[2]HEART 7,[3]DIAMOND 9,[4]SPADE 11\n----------\ndealer hand [Full House]\nCLOVER 8,HEART 8,DIAMOND 8,CLOVER 3,HEART 3\n----------\nIt is your loss.\n----------\n"
		assert.Equal(t, expected, tpp.Output(tp))
	})

	t.Run("success Output end phase draw", func(t *testing.T) {
		tp.Reset()
		player.Reset()
		dealer.Reset()
		// player: Two Pair
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		player.AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		player.AddCard(entities.NewCard(entities.CardDesignHeart, 9, false))
		player.AddCard(entities.NewCard(entities.CardDesignDiamond, 9, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		// dealer: Two Pair with same values (rank >= TwoPair → no exchange)
		dealer.AddCard(entities.NewCard(entities.CardDesignHeart, 5, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignDiamond, 5, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 9, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignHeart, 9, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 11, false))
		tp.PlayerStand()
		expected := "----------\nplayer hand [Two Pair]\n[0]SPADE 5,[1]CLOVER 5,[2]HEART 9,[3]DIAMOND 9,[4]SPADE 11\n----------\ndealer hand [Two Pair]\nHEART 5,DIAMOND 5,CLOVER 9,HEART 9,CLOVER 11\n----------\nIt is a draw.\n----------\n"
		assert.Equal(t, expected, tpp.Output(tp))
	})

	t.Run("success GetCardStr SPADE", func(t *testing.T) {
		assert.Equal(t, "SPADE 1", tpp.GetCardStr(entities.NewCard(entities.CardDesignSpade, 1, false)))
	})
	t.Run("success GetCardStr CLOVER", func(t *testing.T) {
		assert.Equal(t, "CLOVER 1", tpp.GetCardStr(entities.NewCard(entities.CardDesignClover, 1, false)))
	})
	t.Run("success GetCardStr HEART", func(t *testing.T) {
		assert.Equal(t, "HEART 1", tpp.GetCardStr(entities.NewCard(entities.CardDesignHeart, 1, false)))
	})
	t.Run("success GetCardStr DIAMOND", func(t *testing.T) {
		assert.Equal(t, "DIAMOND 1", tpp.GetCardStr(entities.NewCard(entities.CardDesignDiamond, 1, false)))
	})
	t.Run("success GetCardStr JOKER", func(t *testing.T) {
		assert.Equal(t, "Unsupported card 0", tpp.GetCardStr(entities.NewCard(entities.CardDesignJoker, entities.CardValueJoker, false)))
	})
}
