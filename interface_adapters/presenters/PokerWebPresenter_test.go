package presenters_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"
	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/presenters"

	"github.com/stretchr/testify/assert"
)

func TestPokerWebPresenter_Method(t *testing.T) {
	tpp := presenters.NewPokerWebPresenter()
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
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 8, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignHeart, 8, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignDiamond, 8, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 3, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignHeart, 3, false))
		expected := `{"dealer":{"handRank":0,"handName":"","cards":[]},"player":{"handRank":0,"handName":"High Card","cards":[{"design":"SPADE","value":5},{"design":"CLOVER","value":6},{"design":"HEART","value":7},{"design":"DIAMOND","value":8},{"design":"SPADE","value":9}]},"phase":1,"message":""}`
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
		// dealer: Full House (rank >= TwoPair → no exchange)
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 8, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignSpade, 8, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignDiamond, 8, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 3, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))
		tp.PlayerStand()
		expected := `{"dealer":{"handRank":6,"handName":"Full House","cards":[{"design":"CLOVER","value":8},{"design":"SPADE","value":8},{"design":"DIAMOND","value":8},{"design":"CLOVER","value":3},{"design":"SPADE","value":3}]},"player":{"handRank":8,"handName":"Straight Flush","cards":[{"design":"HEART","value":3},{"design":"HEART","value":4},{"design":"HEART","value":5},{"design":"HEART","value":6},{"design":"HEART","value":7}]},"phase":2,"message":"You are the winner."}`
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
		expected := `{"dealer":{"handRank":6,"handName":"Full House","cards":[{"design":"CLOVER","value":8},{"design":"HEART","value":8},{"design":"DIAMOND","value":8},{"design":"CLOVER","value":3},{"design":"HEART","value":3}]},"player":{"handRank":0,"handName":"High Card","cards":[{"design":"SPADE","value":2},{"design":"CLOVER","value":5},{"design":"HEART","value":7},{"design":"DIAMOND","value":9},{"design":"SPADE","value":11}]},"phase":2,"message":"It is your loss."}`
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
		expected := `{"dealer":{"handRank":2,"handName":"Two Pair","cards":[{"design":"HEART","value":5},{"design":"DIAMOND","value":5},{"design":"CLOVER","value":9},{"design":"HEART","value":9},{"design":"CLOVER","value":11}]},"player":{"handRank":2,"handName":"Two Pair","cards":[{"design":"SPADE","value":5},{"design":"CLOVER","value":5},{"design":"HEART","value":9},{"design":"DIAMOND","value":9},{"design":"SPADE","value":11}]},"phase":2,"message":"It is a draw."}`
		assert.Equal(t, expected, tpp.Output(tp))
	})

	t.Run("success GetCardObj SPADE", func(t *testing.T) {
		card := tpp.GetCardObj(entities.NewCard(entities.CardDesignSpade, 1, false))
		assert.Equal(t, "SPADE", card.Design)
		assert.Equal(t, 1, card.Value)
	})
	t.Run("success GetCardObj CLOVER", func(t *testing.T) {
		card := tpp.GetCardObj(entities.NewCard(entities.CardDesignClover, 1, false))
		assert.Equal(t, "CLOVER", card.Design)
		assert.Equal(t, 1, card.Value)
	})
	t.Run("success GetCardObj HEART", func(t *testing.T) {
		card := tpp.GetCardObj(entities.NewCard(entities.CardDesignHeart, 1, false))
		assert.Equal(t, "HEART", card.Design)
		assert.Equal(t, 1, card.Value)
	})
	t.Run("success GetCardObj DIAMOND", func(t *testing.T) {
		card := tpp.GetCardObj(entities.NewCard(entities.CardDesignDiamond, 1, false))
		assert.Equal(t, "DIAMOND", card.Design)
		assert.Equal(t, 1, card.Value)
	})
	t.Run("success GetCardObj JOKER", func(t *testing.T) {
		card := tpp.GetCardObj(entities.NewCard(entities.CardDesignJoker, entities.CardValueJoker, false))
		assert.Equal(t, "Unsupported card", card.Design)
		assert.Equal(t, 0, card.Value)
	})
}
