package presenter_test

import (
	"encoding/json"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestPokerWebPresenter_Method(t *testing.T) {
	tpp := presenter.NewPokerWebPresenter()
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)

	t.Run("success Output deal phase", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		player.Reset()
		dealer.Reset()
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		player.AddCard(domain.NewCard(domain.CardDesignClover, 6, false))
		player.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		player.AddCard(domain.NewCard(domain.CardDesignDiamond, 8, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 8, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 8, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 3, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		output := tpp.Output(tp, nil)
		var result controller.PokerWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, domain.PokerPhaseDeal, result.Phase)
		assert.Equal(t, 5, len(result.Player.Cards))
		assert.Equal(t, 0, len(result.Dealer.Cards))
		assert.Equal(t, "", result.Dealer.HandName)
		assert.True(t, result.Pot >= 0)
		assert.Equal(t, domain.PokerDefaultAnte, result.Ante)
	})

	t.Run("success Output end phase player wins", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		// Move to exchange phase
		if tp.GetDealerBet() > 0 {
			_ = tp.PlayerCall()
		} else {
			_ = tp.PlayerCheck()
		}
		player.Reset()
		dealer.Reset()
		// player: Straight Flush (3-4-5-6-7 same suit)
		player.AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		player.AddCard(domain.NewCard(domain.CardDesignHeart, 4, false))
		player.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		player.AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
		player.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		// dealer: Full House (rank >= TwoPair -> no exchange)
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 8, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 8, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 3, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		_ = tp.PlayerStand()
		// Move to end
		if tp.GetDealerBet() > 0 {
			_ = tp.PlayerCall()
		} else {
			_ = tp.PlayerCheck()
		}
		output := tpp.Output(tp, nil)
		var result controller.PokerWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, domain.PokerPhaseEnd, result.Phase)
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
			_ = tp.PlayerCall()
		} else {
			_ = tp.PlayerCheck()
		}
		player.Reset()
		dealer.Reset()
		// player: High Card
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		player.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		player.AddCard(domain.NewCard(domain.CardDesignDiamond, 9, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		// dealer: Full House (rank >= TwoPair -> no exchange)
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 8, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 8, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 3, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		_ = tp.PlayerStand()
		if tp.GetDealerBet() > 0 {
			_ = tp.PlayerCall()
		} else {
			_ = tp.PlayerCheck()
		}
		output := tpp.Output(tp, nil)
		var result controller.PokerWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, domain.PokerPhaseEnd, result.Phase)
		assert.Equal(t, "It is your loss.", result.Message)
	})

	t.Run("success Output end phase draw", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		if tp.GetDealerBet() > 0 {
			_ = tp.PlayerCall()
		} else {
			_ = tp.PlayerCheck()
		}
		player.Reset()
		dealer.Reset()
		// player: Two Pair
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		player.AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
		player.AddCard(domain.NewCard(domain.CardDesignDiamond, 9, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		// dealer: Two Pair with same values (rank >= TwoPair -> no exchange)
		dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 5, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 9, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 11, false))
		_ = tp.PlayerStand()
		if tp.GetDealerBet() > 0 {
			_ = tp.PlayerCall()
		} else {
			_ = tp.PlayerCheck()
		}
		output := tpp.Output(tp, nil)
		var result controller.PokerWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, domain.PokerPhaseEnd, result.Phase)
		assert.Equal(t, "It is a draw.", result.Message)
	})

	t.Run("success Output fold phase player folds", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		_ = tp.PlayerFold()
		output := tpp.Output(tp, nil)
		var result controller.PokerWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, domain.PokerPhaseEnd, result.Phase)
		assert.Equal(t, "You folded.", result.Message)
	})

	t.Run("success Output includes chip and bet info", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		output := tpp.Output(tp, nil)
		var result controller.PokerWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.True(t, result.Player.Chips >= 0)
		assert.True(t, result.Dealer.Chips >= 0)
		assert.True(t, result.Pot >= 0)
		assert.Equal(t, domain.PokerDefaultAnte, result.Ante)
	})

	t.Run("success Output displays error message in JSON", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		testErr := domain.NewDomainError(domain.ErrWrongPhase, "Bet is not allowed now.")
		output := tpp.Output(tp, testErr)
		var result controller.PokerWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, "Bet is not allowed now.", result.Message)
	})

	t.Run("success Output dealer fold", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		tp.SetPhase(domain.PokerPhaseEnd)
		tp.SetFolded(domain.PokerFoldByDealer)
		output := tpp.Output(tp, nil)
		var result controller.PokerWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, domain.PokerPhaseEnd, result.Phase)
		assert.Equal(t, "Dealer folded. You win!", result.Message)
	})

	t.Run("success Output dealer cards hidden in non-end phase", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		player.Reset()
		dealer.Reset()
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		player.AddCard(domain.NewCard(domain.CardDesignClover, 6, false))
		player.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		player.AddCard(domain.NewCard(domain.CardDesignDiamond, 8, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 8, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 8, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 3, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		output := tpp.Output(tp, nil)
		var result controller.PokerWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		// In non-end phase, dealer cards should be empty
		assert.Equal(t, 0, len(result.Dealer.Cards))
		assert.Equal(t, "", result.Dealer.HandName)
	})
}
