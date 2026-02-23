package presenter_test

import (
	"encoding/json"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"

	"github.com/stretchr/testify/assert"
)

func TestBlackJackWebPresenters_Method(t *testing.T) {
	tbp := presenter.NewBlackJackWebPresenter()

	t.Run("success Output bet phase (no cards)", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		output := tbp.Output(bj)
		var result controller.BlackJackWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, domain.BJPhaseBet, result.Phase)
		assert.Equal(t, domain.BJDefaultChips, result.Player.Chips)
		assert.Equal(t, domain.BJDefaultChips, result.Dealer.Chips)
		assert.Equal(t, 0, len(result.Dealer.Cards))
		assert.Equal(t, 1, len(result.Hands))
		assert.Equal(t, 0, result.Hands[0].Score)
	})
	t.Run("success Output action phase", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		player := domain.NewBlackJackPlayer()
		dealer := domain.NewBlackJackPlayer()
		player.SetChips(900)
		dealer.SetChips(1000)
		bj := domain.NewBlackJack(tc, player, dealer)
		bj.Reset()
		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 2, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 11, false))
		bj.SetPhase(domain.BJPhaseAction)
		output := tbp.Output(bj)
		var result controller.BlackJackWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, 0, result.Dealer.Score) // Not game end, so score hidden
		assert.Equal(t, 1, len(result.Dealer.Cards))
		assert.Equal(t, 900, result.Player.Chips)
		assert.Equal(t, 1000, result.Dealer.Chips)
		assert.Equal(t, domain.BJPhaseAction, result.Phase)
		assert.Equal(t, 1, len(result.Hands))
		assert.Equal(t, 100, result.Hands[0].Bet)
	})
	t.Run("success Output end phase lose", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		player := domain.NewBlackJackPlayer()
		dealer := domain.NewBlackJackPlayer()
		player.SetChips(900)
		dealer.SetChips(1000)
		bj := domain.NewBlackJack(tc, player, dealer)
		bj.Reset()
		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 2, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 11, false))
		bj.SetPhase(domain.BJPhaseAction)
		bj.PlayerStand()
		output := tbp.Output(bj)
		var result controller.BlackJackWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, 22, result.Dealer.Score)
		assert.Equal(t, 3, len(result.Dealer.Cards))
		assert.Equal(t, "It is your loss.", result.Message)
		assert.Equal(t, domain.BJPhaseEnd, result.Phase)
	})
	t.Run("success Output end phase draw", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		player := domain.NewBlackJackPlayer()
		dealer := domain.NewBlackJackPlayer()
		player.SetChips(900)
		dealer.SetChips(1000)
		bj := domain.NewBlackJack(tc, player, dealer)
		bj.Reset()
		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
		bj.SetPhase(domain.BJPhaseAction)
		bj.PlayerStand()
		output := tbp.Output(bj)
		var result controller.BlackJackWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, "It is a draw.", result.Message)
	})
	t.Run("success Output end phase win", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		player := domain.NewBlackJackPlayer()
		dealer := domain.NewBlackJackPlayer()
		player.SetChips(900)
		dealer.SetChips(1000)
		bj := domain.NewBlackJack(tc, player, dealer)
		bj.Reset()
		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		// Dealer score 19 (>= 17)
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 9, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
		bj.SetPhase(domain.BJPhaseAction)
		bj.PlayerStand()
		output := tbp.Output(bj)
		var result controller.BlackJackWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, "You are the winner.", result.Message)
	})
	t.Run("success Output hands fields", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		player := domain.NewBlackJackPlayer()
		dealer := domain.NewBlackJackPlayer()
		player.SetChips(900)
		dealer.SetChips(1000)
		bj := domain.NewBlackJack(tc, player, dealer)
		bj.Reset()
		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.SetDoubled(true)
		hand.SetBusted(true)
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		hand.AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
		hand.AddCard(domain.NewCard(domain.CardDesignClover, 8, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 10, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
		bj.SetPhase(domain.BJPhaseAction)
		output := tbp.Output(bj)
		var result controller.BlackJackWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(result.Hands))
		assert.True(t, result.Hands[0].Doubled)
		assert.True(t, result.Hands[0].Busted)
		assert.Equal(t, 100, result.Hands[0].Bet)
		assert.Equal(t, 3, len(result.Hands[0].Cards))
	})
	t.Run("success Output insurance fields", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		player := domain.NewBlackJackPlayer()
		dealer := domain.NewBlackJackPlayer()
		player.SetChips(900)
		dealer.SetChips(1000)
		bj := domain.NewBlackJack(tc, player, dealer)
		bj.Reset()
		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		hand.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 10, false))
		bj.SetPhase(domain.BJPhaseInsurance)
		bj.PlayerInsurance() // cost = 50
		output := tbp.Output(bj)
		var result controller.BlackJackWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, 50, result.InsuranceBet)
	})
	t.Run("success Output canSplit and isBlackJack", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		player := domain.NewBlackJackPlayer()
		dealer := domain.NewBlackJackPlayer()
		player.SetChips(900)
		dealer.SetChips(1000)
		bj := domain.NewBlackJack(tc, player, dealer)
		bj.Reset()
		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		hand.AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
		bj.SetPhase(domain.BJPhaseAction)
		output := tbp.Output(bj)
		var result controller.BlackJackWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.True(t, result.Hands[0].CanSplit)
		assert.False(t, result.Hands[0].IsBlackJack)
	})
	t.Run("success Output error message on failed bet", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		bj.PlayerBet(5) // invalid amount, sets lastError
		output := tbp.Output(bj)
		var result controller.BlackJackWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, domain.BJPhaseBet, result.Phase)
		assert.Equal(t, "Invalid bet amount.", result.Message)
	})
	t.Run("success GetCardObj SPADE", func(t *testing.T) {
		card := tbp.GetCardObj(domain.NewCard(domain.CardDesignSpade, 1, false))
		assert.Equal(t, "SPADE", card.Design)
		assert.Equal(t, 1, card.Value)
	})
	t.Run("success GetCardObj CLOVER", func(t *testing.T) {
		card := tbp.GetCardObj(domain.NewCard(domain.CardDesignClover, 1, false))
		assert.Equal(t, "CLOVER", card.Design)
		assert.Equal(t, 1, card.Value)
	})
	t.Run("success GetCardObj HEART", func(t *testing.T) {
		card := tbp.GetCardObj(domain.NewCard(domain.CardDesignHeart, 1, false))
		assert.Equal(t, "HEART", card.Design)
		assert.Equal(t, 1, card.Value)
	})
	t.Run("success GetCardObj DIAMOND", func(t *testing.T) {
		card := tbp.GetCardObj(domain.NewCard(domain.CardDesignDiamond, 1, false))
		assert.Equal(t, "DIAMOND", card.Design)
		assert.Equal(t, 1, card.Value)
	})
	t.Run("success GetCardObj unsupported", func(t *testing.T) {
		card := tbp.GetCardObj(domain.NewCard(domain.CardDesignJoker, domain.CardValueJoker, false))
		assert.Equal(t, "Unsupported card", card.Design)
		assert.Equal(t, 0, card.Value)
	})
}
