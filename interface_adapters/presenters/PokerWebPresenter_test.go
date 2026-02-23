package presenters_test

import (
	"encoding/json"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"
	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/controllers"
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
		player.SetChips(0)
		dealer.SetChips(0)
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
		output := tpp.Output(tp)
		var result controllers.PokerWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, entities.PokerPhaseDeal, result.Phase)
		assert.Equal(t, 5, len(result.Player.Cards))
		assert.Equal(t, 0, len(result.Dealer.Cards))
		assert.Equal(t, "", result.Dealer.HandName)
		assert.True(t, result.Pot >= 0)
		assert.Equal(t, entities.PokerDefaultAnte, result.Ante)
	})

	t.Run("success Output end phase player wins", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		// Move to exchange phase
		if tp.GetDealerBet() > 0 {
			tp.PlayerCall()
		} else {
			tp.PlayerCheck()
		}
		player.Reset()
		dealer.Reset()
		// player: Straight Flush (3-4-5-6-7 same suit)
		player.AddCard(entities.NewCard(entities.CardDesignHeart, 3, false))
		player.AddCard(entities.NewCard(entities.CardDesignHeart, 4, false))
		player.AddCard(entities.NewCard(entities.CardDesignHeart, 5, false))
		player.AddCard(entities.NewCard(entities.CardDesignHeart, 6, false))
		player.AddCard(entities.NewCard(entities.CardDesignHeart, 7, false))
		// dealer: Full House (rank >= TwoPair -> no exchange)
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 8, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignSpade, 8, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignDiamond, 8, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 3, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))
		tp.PlayerStand()
		// Move to end
		if tp.GetDealerBet() > 0 {
			tp.PlayerCall()
		} else {
			tp.PlayerCheck()
		}
		output := tpp.Output(tp)
		var result controllers.PokerWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, entities.PokerPhaseEnd, result.Phase)
		assert.Equal(t, "Straight Flush", result.Player.HandName)
		assert.Equal(t, "Full House", result.Dealer.HandName)
		assert.Equal(t, "You are the winner.", result.Message)
		assert.Equal(t, 5, len(result.Dealer.Cards))
	})

	t.Run("success Output end phase player loses", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		if tp.GetDealerBet() > 0 {
			tp.PlayerCall()
		} else {
			tp.PlayerCheck()
		}
		player.Reset()
		dealer.Reset()
		// player: High Card
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		player.AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		player.AddCard(entities.NewCard(entities.CardDesignHeart, 7, false))
		player.AddCard(entities.NewCard(entities.CardDesignDiamond, 9, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		// dealer: Full House (rank >= TwoPair -> no exchange)
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 8, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignHeart, 8, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignDiamond, 8, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 3, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignHeart, 3, false))
		tp.PlayerStand()
		if tp.GetDealerBet() > 0 {
			tp.PlayerCall()
		} else {
			tp.PlayerCheck()
		}
		output := tpp.Output(tp)
		var result controllers.PokerWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, entities.PokerPhaseEnd, result.Phase)
		assert.Equal(t, "It is your loss.", result.Message)
	})

	t.Run("success Output end phase draw", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		if tp.GetDealerBet() > 0 {
			tp.PlayerCall()
		} else {
			tp.PlayerCheck()
		}
		player.Reset()
		dealer.Reset()
		// player: Two Pair
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		player.AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		player.AddCard(entities.NewCard(entities.CardDesignHeart, 9, false))
		player.AddCard(entities.NewCard(entities.CardDesignDiamond, 9, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		// dealer: Two Pair with same values (rank >= TwoPair -> no exchange)
		dealer.AddCard(entities.NewCard(entities.CardDesignHeart, 5, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignDiamond, 5, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 9, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignHeart, 9, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 11, false))
		tp.PlayerStand()
		if tp.GetDealerBet() > 0 {
			tp.PlayerCall()
		} else {
			tp.PlayerCheck()
		}
		output := tpp.Output(tp)
		var result controllers.PokerWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, entities.PokerPhaseEnd, result.Phase)
		assert.Equal(t, "It is a draw.", result.Message)
	})

	t.Run("success Output fold phase player folds", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		tp.PlayerFold()
		output := tpp.Output(tp)
		var result controllers.PokerWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, entities.PokerPhaseEnd, result.Phase)
		assert.Equal(t, "You folded.", result.Message)
	})

	t.Run("success Output includes chip and bet info", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		output := tpp.Output(tp)
		var result controllers.PokerWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.True(t, result.Player.Chips >= 0)
		assert.True(t, result.Dealer.Chips >= 0)
		assert.True(t, result.Pot >= 0)
		assert.Equal(t, entities.PokerDefaultAnte, result.Ante)
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
