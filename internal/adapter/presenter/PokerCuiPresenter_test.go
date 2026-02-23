package presenter_test

import (
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestPokerCuiPresenter_Method(t *testing.T) {
	tpp := presenter.NewPokerCuiPresenter()
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
		dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 8, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		output := tpp.Output(tp, nil)
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
		// dealer: Full House (rank >= TwoPair -> no exchange when PlayerStand called)
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 8, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 8, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 3, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		_ = tp.PlayerStand()
		if tp.GetDealerBet() > 0 {
			_ = tp.PlayerCall()
		} else {
			_ = tp.PlayerCheck()
		}
		output := tpp.Output(tp, nil)
		assert.Contains(t, output, "player hand [Straight Flush]")
		assert.Contains(t, output, "dealer hand [Full House]")
		assert.Contains(t, output, "You are the winner.")
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
		assert.Contains(t, output, "It is your loss.")
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
		assert.Contains(t, output, "It is a draw.")
	})

	t.Run("success Output fold", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		_ = tp.PlayerFold()
		output := tpp.Output(tp, nil)
		assert.Contains(t, output, "You folded.")
	})

	t.Run("success Output shows dealer bet", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		output := tpp.Output(tp, nil)
		if tp.GetDealerBet() > 0 {
			assert.True(t, strings.Contains(output, "Dealer Bet:"))
		}
	})

	t.Run("success Output displays error message", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		testErr := domain.NewDomainError(domain.ErrWrongPhase, "Bet is not allowed now.")
		output := tpp.Output(tp, testErr)
		assert.Contains(t, output, "Bet is not allowed now.")
		assert.NotContains(t, output, "wrong game phase")
	})

	t.Run("success GetCardStr SPADE", func(t *testing.T) {
		assert.Equal(t, "SPADE 1", tpp.GetCardStr(domain.NewCard(domain.CardDesignSpade, 1, false)))
	})
	t.Run("success GetCardStr CLOVER", func(t *testing.T) {
		assert.Equal(t, "CLOVER 1", tpp.GetCardStr(domain.NewCard(domain.CardDesignClover, 1, false)))
	})
	t.Run("success GetCardStr HEART", func(t *testing.T) {
		assert.Equal(t, "HEART 1", tpp.GetCardStr(domain.NewCard(domain.CardDesignHeart, 1, false)))
	})
	t.Run("success GetCardStr DIAMOND", func(t *testing.T) {
		assert.Equal(t, "DIAMOND 1", tpp.GetCardStr(domain.NewCard(domain.CardDesignDiamond, 1, false)))
	})
	t.Run("success GetCardStr JOKER", func(t *testing.T) {
		assert.Equal(t, "Unsupported card 0", tpp.GetCardStr(domain.NewCard(domain.CardDesignJoker, domain.CardValueJoker, false)))
	})

	t.Run("success Output dealer fold", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		// Force dealer fold via setter
		tp.SetPhase(domain.PokerPhaseEnd)
		tp.SetFolded(domain.PokerFoldByDealer)
		output := tpp.Output(tp, nil)
		assert.Contains(t, output, "Dealer folded. You win!")
	})

	t.Run("success Output dealer bet zero", func(t *testing.T) {
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
		// When dealer bet is 0, "Dealer Bet:" line should not appear
		if tp.GetDealerBet() == 0 {
			output := tpp.Output(tp, nil)
			assert.NotContains(t, output, "Dealer Bet:")
		}
	})

	t.Run("success Output dealer hand hidden in non-end phase", func(t *testing.T) {
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
		// In deal phase, dealer hand name should NOT be shown
		assert.NotContains(t, output, "Full House")
		// "dealer hand\n" without hand name
		assert.Contains(t, output, "dealer hand\n")
	})
}
