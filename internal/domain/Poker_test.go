package domain_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestPoker_Method(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)

	t.Run("success Reset", func(t *testing.T) {
		tp.Reset()
		assert.Equal(t, domain.PokerPhaseDeal, tp.GetPhase())
		assert.Equal(t, 5, tp.GetPlayer().GetCardsSize())
		assert.Equal(t, 5, tp.GetDealer().GetCardsSize())
	})

	t.Run("success Reset initializes chips", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		assert.Equal(t, domain.PokerDefaultChips-domain.PokerDefaultAnte, tp.GetPlayer().GetChips())
		assert.True(t, tp.GetPot() >= domain.PokerDefaultAnte*2)
	})

	t.Run("success GetPlayer", func(t *testing.T) {
		assert.NotEmpty(t, tp.GetPlayer())
	})

	t.Run("success GetDealer", func(t *testing.T) {
		assert.NotEmpty(t, tp.GetDealer())
	})

	t.Run("success GetAnte", func(t *testing.T) {
		assert.Equal(t, domain.PokerDefaultAnte, tp.GetAnte())
	})

	t.Run("success PlayerBet advances to Exchange phase", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		err := tp.PlayerBet(domain.PokerMinBet)
		assert.NoError(t, err)
		assert.Equal(t, domain.PokerPhaseExchange, tp.GetPhase())
	})

	t.Run("success PlayerCheck advances to Exchange phase when no dealer bet", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		// ディーラーがチェック (ハイカードの場合) なら dealerBet=0
		if tp.GetDealerBet() == 0 {
			err := tp.PlayerCheck()
			assert.NoError(t, err)
			assert.Equal(t, domain.PokerPhaseExchange, tp.GetPhase())
		}
	})

	t.Run("success PlayerCall matches dealer bet", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		if tp.GetDealerBet() > 0 {
			chipsBefore := tp.GetPlayer().GetChips()
			err := tp.PlayerCall()
			assert.NoError(t, err)
			assert.Equal(t, domain.PokerPhaseExchange, tp.GetPhase())
			assert.True(t, tp.GetPlayer().GetChips() < chipsBefore)
		}
	})

	t.Run("success PlayerFold ends game", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		err := tp.PlayerFold()
		assert.NoError(t, err)
		assert.Equal(t, domain.PokerPhaseEnd, tp.GetPhase())
		assert.Equal(t, domain.PokerFoldByPlayer, tp.GetFolded())
		assert.Equal(t, 0, tp.GetPot())
	})

	t.Run("success PlayerBet rejected when not in betting phase", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		_ = tp.PlayerFold()
		err := tp.PlayerBet(domain.PokerMinBet)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrWrongPhase)
	})

	t.Run("success PlayerBet rejected below minimum", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		err := tp.PlayerBet(1)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidAmount)
	})

	t.Run("success PlayerBet rejected insufficient chips", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		err := tp.PlayerBet(999999)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInsufficientChips)
	})

	t.Run("success PlayerRaise rejected on overflow amount", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		err := tp.PlayerRaise(1 << 62)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInsufficientChips)
		assert.Equal(t, domain.PokerPhaseDeal, tp.GetPhase())
	})

	t.Run("success PlayerCheck rejected when dealer has bet", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		if tp.GetDealerBet() > 0 {
			err := tp.PlayerCheck()
			assert.Error(t, err)
			assert.ErrorIs(t, err, domain.ErrInvalidPlay)
		}
	})

	t.Run("success PlayerExchange moves to SecondBet phase", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		// Move to Exchange phase
		if tp.GetDealerBet() > 0 {
			_ = tp.PlayerCall()
		} else {
			_ = tp.PlayerCheck()
		}
		assert.Equal(t, domain.PokerPhaseExchange, tp.GetPhase())
		err := tp.PlayerExchange([]int{0, 1})
		assert.NoError(t, err)
		assert.Equal(t, domain.PokerPhaseSecondBet, tp.GetPhase())
	})

	t.Run("success PlayerStand moves to SecondBet phase", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		if tp.GetDealerBet() > 0 {
			_ = tp.PlayerCall()
		} else {
			_ = tp.PlayerCheck()
		}
		assert.Equal(t, domain.PokerPhaseExchange, tp.GetPhase())
		err := tp.PlayerStand()
		assert.NoError(t, err)
		assert.Equal(t, domain.PokerPhaseSecondBet, tp.GetPhase())
	})

	t.Run("success PlayerExchange rejected when not in Exchange phase", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		// In Deal phase, not Exchange
		err := tp.PlayerExchange([]int{0})
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrWrongPhase)
		assert.Equal(t, domain.PokerPhaseDeal, tp.GetPhase())
	})

	t.Run("success PlayerStand rejected when not in Exchange phase", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		// In Deal phase, not Exchange
		err := tp.PlayerStand()
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrWrongPhase)
		assert.Equal(t, domain.PokerPhaseDeal, tp.GetPhase())
	})

	t.Run("success Full game flow with showdown", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		// First bet
		if tp.GetDealerBet() > 0 {
			_ = tp.PlayerCall()
		} else {
			_ = tp.PlayerCheck()
		}
		// Exchange
		_ = tp.PlayerStand()
		assert.Equal(t, domain.PokerPhaseSecondBet, tp.GetPhase())
		// Second bet
		if tp.GetDealerBet() > 0 {
			_ = tp.PlayerCall()
		} else {
			_ = tp.PlayerCheck()
		}
		assert.Equal(t, domain.PokerPhaseEnd, tp.GetPhase())
		assert.Equal(t, domain.PokerFoldNone, tp.GetFolded())
	})

	t.Run("success GameJudgment player win higher rank", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		player.Reset()
		dealer.Reset()
		// player: Royal Flush
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		// dealer: High Card
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 2, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 9, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))
		player.EvalHand()
		dealer.EvalHand()
		assert.Equal(t, 1, tp.GameJudgment())
	})

	t.Run("success GameJudgment player lose lower rank", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		player.Reset()
		dealer.Reset()
		// player: High Card
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		player.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		player.AddCard(domain.NewCard(domain.CardDesignDiamond, 9, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		// dealer: One Pair
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 9, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))
		player.EvalHand()
		dealer.EvalHand()
		assert.Equal(t, -1, tp.GameJudgment())
	})

	t.Run("success GameJudgment draw same rank same high cards", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		player.Reset()
		dealer.Reset()
		// both: High Card with same values
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 2, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 7, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 9, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 11, false))
		player.EvalHand()
		dealer.EvalHand()
		assert.Equal(t, 0, tp.GameJudgment())
	})

	t.Run("success GameJudgment player win same rank higher cards", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		player.Reset()
		dealer.Reset()
		// both: High Card, player has higher top card
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 2, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 7, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 9, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 11, false))
		player.EvalHand()
		dealer.EvalHand()
		assert.Equal(t, 1, tp.GameJudgment())
	})

	t.Run("success GameJudgment player lose same rank lower cards", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		player.Reset()
		dealer.Reset()
		// both: High Card, player has lower top card
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 2, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 7, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 9, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 13, false))
		player.EvalHand()
		dealer.EvalHand()
		assert.Equal(t, -1, tp.GameJudgment())
	})

	t.Run("success GameJudgment ace treated as high card", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		player.Reset()
		dealer.Reset()
		// player has ace (high), dealer has king
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 13, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 7, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 9, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 11, false))
		player.EvalHand()
		dealer.EvalHand()
		assert.Equal(t, 1, tp.GameJudgment())
	})

	t.Run("success PlayerRaise advances phase", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		err := tp.PlayerRaise(domain.PokerMinBet)
		assert.NoError(t, err)
		// Should advance or end depending on dealer response
		assert.True(t, tp.GetPhase() == domain.PokerPhaseExchange || tp.GetPhase() == domain.PokerPhaseEnd)
	})

	t.Run("success PlayerRaise rejected below minimum", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		err := tp.PlayerRaise(1)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidAmount)
	})

	t.Run("success Showdown pot distribution player wins", func(t *testing.T) {
		player.SetChips(500)
		dealer.SetChips(500)
		tp.Reset()
		// Navigate to Exchange phase
		if tp.GetDealerBet() > 0 {
			_ = tp.PlayerCall()
		} else {
			_ = tp.PlayerCheck()
		}
		// Set up deterministic hands
		player.Reset()
		dealer.Reset()
		// player: Four of a Kind
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		player.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
		player.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
		player.AddCard(domain.NewCard(domain.CardDesignDiamond, 10, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		// dealer: Full House (rank >= TwoPair -> no exchange)
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 8, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 8, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 3, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		_ = tp.PlayerStand()
		// Second bet
		if tp.GetDealerBet() > 0 {
			_ = tp.PlayerCall()
		} else {
			_ = tp.PlayerCheck()
		}
		assert.Equal(t, domain.PokerPhaseEnd, tp.GetPhase())
		assert.Equal(t, 0, tp.GetPot())
		assert.Equal(t, 1, tp.GameJudgment())
	})

	t.Run("success Dealer fold when player makes large bet with high card dealer", func(t *testing.T) {
		player.SetChips(500)
		dealer.SetChips(500)
		tp.Reset()
		player.Reset()
		dealer.Reset()
		// dealer: High Card (will fold on large bet)
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 2, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 9, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		// Large bet should cause dealer fold
		err := tp.PlayerBet(domain.PokerMinBet * 3)
		assert.NoError(t, err)
		if tp.GetFolded() == domain.PokerFoldByDealer {
			assert.Equal(t, domain.PokerPhaseEnd, tp.GetPhase())
			assert.Equal(t, 1, tp.GameJudgment())
		}
	})

	t.Run("success Flush draw exchange", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		// Move to exchange phase
		if tp.GetDealerBet() > 0 {
			_ = tp.PlayerCall()
		} else {
			_ = tp.PlayerCheck()
		}
		// Set up dealer with 4-card flush draw
		dealer.Reset()
		dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 3, false)) // off-suit
		_ = tp.PlayerStand()
		// After exchange, dealer should have replaced the off-suit card
		assert.Equal(t, domain.PokerPhaseSecondBet, tp.GetPhase())
	})

	t.Run("success Straight draw exchange", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		// Move to exchange phase
		if tp.GetDealerBet() > 0 {
			_ = tp.PlayerCall()
		} else {
			_ = tp.PlayerCheck()
		}
		// Set up dealer with 4-card straight draw (5-6-7-8 + off card)
		dealer.Reset()
		dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 6, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 8, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 12, false)) // outlier
		_ = tp.PlayerStand()
		assert.Equal(t, domain.PokerPhaseSecondBet, tp.GetPhase())
	})
}
