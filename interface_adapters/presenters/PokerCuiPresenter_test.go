package presenters_test

import (
	"strings"
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
		dealer.AddCard(entities.NewCard(entities.CardDesignSpade, 8, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 8, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignHeart, 8, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignDiamond, 3, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))
		output := tpp.Output(tp)
		assert.Contains(t, output, "Pot:")
		assert.Contains(t, output, "Player Chips:")
		assert.Contains(t, output, "Dealer Chips:")
		assert.Contains(t, output, "player hand")
		assert.Contains(t, output, "[0]SPADE 5")
		assert.Contains(t, output, "dealer hand")
		assert.NotContains(t, output, "You are the winner.")
	})

	t.Run("success Output end phase player wins", func(t *testing.T) {
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
		// player: Straight Flush (3-4-5-6-7 same suit)
		player.AddCard(entities.NewCard(entities.CardDesignHeart, 3, false))
		player.AddCard(entities.NewCard(entities.CardDesignHeart, 4, false))
		player.AddCard(entities.NewCard(entities.CardDesignHeart, 5, false))
		player.AddCard(entities.NewCard(entities.CardDesignHeart, 6, false))
		player.AddCard(entities.NewCard(entities.CardDesignHeart, 7, false))
		// dealer: Full House (rank >= TwoPair -> no exchange when PlayerStand called)
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 8, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignSpade, 8, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignDiamond, 8, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 3, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))
		tp.PlayerStand()
		if tp.GetDealerBet() > 0 {
			tp.PlayerCall()
		} else {
			tp.PlayerCheck()
		}
		output := tpp.Output(tp)
		assert.Contains(t, output, "player hand [Straight Flush]")
		assert.Contains(t, output, "dealer hand [Full House]")
		assert.Contains(t, output, "You are the winner.")
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
		assert.Contains(t, output, "It is your loss.")
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
		assert.Contains(t, output, "It is a draw.")
	})

	t.Run("success Output fold", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		tp.PlayerFold()
		output := tpp.Output(tp)
		assert.Contains(t, output, "You folded.")
	})

	t.Run("success Output shows dealer bet", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		output := tpp.Output(tp)
		if tp.GetDealerBet() > 0 {
			assert.True(t, strings.Contains(output, "Dealer Bet:"))
		}
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
