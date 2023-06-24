package presenters_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"
	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/presenters"

	"github.com/stretchr/testify/assert"
)

func TestBlackJackWebPresenters_Method(t *testing.T) {
	tbp := presenters.NewBlackJackWebPresenter()
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
		assert.Equal(t, `{"dealer":{"score":0,"cards":[{"design":"CLOVER","value":2}]},"player":{"score":22,"cards":[{"design":"SPADE","value":2},{"design":"SPADE","value":10},{"design":"SPADE","value":11}]},"message":""}`, tbp.Output(bj))
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
		assert.Equal(t, `{"dealer":{"score":22,"cards":[{"design":"CLOVER","value":2},{"design":"CLOVER","value":10},{"design":"CLOVER","value":11}]},"player":{"score":22,"cards":[{"design":"SPADE","value":2},{"design":"SPADE","value":10},{"design":"SPADE","value":11}]},"message":"It is your loss."}`, tbp.Output(bj))
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
		assert.Equal(t, `{"dealer":{"score":21,"cards":[{"design":"CLOVER","value":1},{"design":"CLOVER","value":10}]},"player":{"score":21,"cards":[{"design":"SPADE","value":1},{"design":"SPADE","value":10}]},"message":"It is a draw."}`, tbp.Output(bj))
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
		assert.Equal(t, `{"dealer":{"score":19,"cards":[{"design":"CLOVER","value":2},{"design":"CLOVER","value":10},{"design":"CLOVER","value":7}]},"player":{"score":21,"cards":[{"design":"SPADE","value":1},{"design":"SPADE","value":10}]},"message":"You are the winner."}`, tbp.Output(bj))
	})
	t.Run("success GetCardStr", func(t *testing.T) {
		card := tbp.GetCardObj(entities.NewCard(entities.CardDesignSpade, 1, false))
		assert.Equal(t, "SPADE", card.Design)
		assert.Equal(t, 1, card.Value)
	})
	t.Run("success GetCardStr", func(t *testing.T) {
		card := tbp.GetCardObj(entities.NewCard(entities.CardDesignClover, 1, false))
		assert.Equal(t, "CLOVER", card.Design)
		assert.Equal(t, 1, card.Value)
	})
	t.Run("success GetCardStr", func(t *testing.T) {
		card := tbp.GetCardObj(entities.NewCard(entities.CardDesignHeart, 1, false))
		assert.Equal(t, "HEART", card.Design)
		assert.Equal(t, 1, card.Value)
	})
	t.Run("success GetCardStr", func(t *testing.T) {
		card := tbp.GetCardObj(entities.NewCard(entities.CardDesignDiamond, 1, false))
		assert.Equal(t, "DIAMOND", card.Design)
		assert.Equal(t, 1, card.Value)
	})
	t.Run("success GetCardStr", func(t *testing.T) {
		card := tbp.GetCardObj(entities.NewCard(entities.CardDesignJoker, entities.CardValueJoker, false))
		assert.Equal(t, "Unsupported card", card.Design)
		assert.Equal(t, 0, card.Value)
	})
}
