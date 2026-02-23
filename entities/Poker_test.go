package entities_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"

	"github.com/stretchr/testify/assert"
)

func TestPoker_Method(t *testing.T) {
	tc := entities.NewTrumpCards(0)
	player := entities.NewPokerPlayer()
	dealer := entities.NewPokerPlayer()
	tp := entities.NewPoker(tc, player, dealer)

	t.Run("success Reset", func(t *testing.T) {
		tp.Reset()
		assert.Equal(t, entities.PokerPhaseDeal, tp.GetPhase())
		assert.Equal(t, 5, tp.GetPlayer().GetCardsSize())
		assert.Equal(t, 5, tp.GetDealer().GetCardsSize())
	})

	t.Run("success Reset initializes chips", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		assert.Equal(t, entities.PokerDefaultChips-entities.PokerDefaultAnte, tp.GetPlayer().GetChips())
		assert.True(t, tp.GetPot() >= entities.PokerDefaultAnte*2)
	})

	t.Run("success GetPlayer", func(t *testing.T) {
		assert.NotEmpty(t, tp.GetPlayer())
	})

	t.Run("success GetDealer", func(t *testing.T) {
		assert.NotEmpty(t, tp.GetDealer())
	})

	t.Run("success GetAnte", func(t *testing.T) {
		assert.Equal(t, entities.PokerDefaultAnte, tp.GetAnte())
	})

	t.Run("success PlayerBet advances to Exchange phase", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		ok := tp.PlayerBet(entities.PokerMinBet)
		assert.True(t, ok)
		assert.Equal(t, entities.PokerPhaseExchange, tp.GetPhase())
	})

	t.Run("success PlayerCheck advances to Exchange phase when no dealer bet", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		// ディーラーがチェック (ハイカードの場合) なら dealerBet=0
		if tp.GetDealerBet() == 0 {
			ok := tp.PlayerCheck()
			assert.True(t, ok)
			assert.Equal(t, entities.PokerPhaseExchange, tp.GetPhase())
		}
	})

	t.Run("success PlayerCall matches dealer bet", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		if tp.GetDealerBet() > 0 {
			chipsBefore := tp.GetPlayer().GetChips()
			ok := tp.PlayerCall()
			assert.True(t, ok)
			assert.Equal(t, entities.PokerPhaseExchange, tp.GetPhase())
			assert.True(t, tp.GetPlayer().GetChips() < chipsBefore)
		}
	})

	t.Run("success PlayerFold ends game", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		tp.PlayerFold()
		assert.Equal(t, entities.PokerPhaseEnd, tp.GetPhase())
		assert.Equal(t, 1, tp.GetFolded())
		assert.Equal(t, 0, tp.GetPot())
	})

	t.Run("success PlayerBet rejected when not in betting phase", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		tp.PlayerFold()
		ok := tp.PlayerBet(entities.PokerMinBet)
		assert.False(t, ok)
	})

	t.Run("success PlayerBet rejected below minimum", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		ok := tp.PlayerBet(1)
		assert.False(t, ok)
	})

	t.Run("success PlayerBet rejected insufficient chips", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		ok := tp.PlayerBet(999999)
		assert.False(t, ok)
	})

	t.Run("success PlayerRaise rejected on overflow amount", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		ok := tp.PlayerRaise(1<<62)
		assert.False(t, ok)
		assert.Equal(t, entities.PokerPhaseDeal, tp.GetPhase())
	})

	t.Run("success PlayerCheck rejected when dealer has bet", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		if tp.GetDealerBet() > 0 {
			ok := tp.PlayerCheck()
			assert.False(t, ok)
		}
	})

	t.Run("success PlayerExchange moves to SecondBet phase", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		// Move to Exchange phase
		if tp.GetDealerBet() > 0 {
			tp.PlayerCall()
		} else {
			tp.PlayerCheck()
		}
		assert.Equal(t, entities.PokerPhaseExchange, tp.GetPhase())
		tp.PlayerExchange([]int{0, 1})
		assert.Equal(t, entities.PokerPhaseSecondBet, tp.GetPhase())
	})

	t.Run("success PlayerStand moves to SecondBet phase", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		if tp.GetDealerBet() > 0 {
			tp.PlayerCall()
		} else {
			tp.PlayerCheck()
		}
		assert.Equal(t, entities.PokerPhaseExchange, tp.GetPhase())
		tp.PlayerStand()
		assert.Equal(t, entities.PokerPhaseSecondBet, tp.GetPhase())
	})

	t.Run("success PlayerExchange ignored when not in Exchange phase", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		// In Deal phase, not Exchange
		tp.PlayerExchange([]int{0})
		assert.Equal(t, entities.PokerPhaseDeal, tp.GetPhase())
	})

	t.Run("success PlayerStand ignored when not in Exchange phase", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		// In Deal phase, not Exchange
		tp.PlayerStand()
		assert.Equal(t, entities.PokerPhaseDeal, tp.GetPhase())
	})

	t.Run("success Full game flow with showdown", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		// First bet
		if tp.GetDealerBet() > 0 {
			tp.PlayerCall()
		} else {
			tp.PlayerCheck()
		}
		// Exchange
		tp.PlayerStand()
		assert.Equal(t, entities.PokerPhaseSecondBet, tp.GetPhase())
		// Second bet
		if tp.GetDealerBet() > 0 {
			tp.PlayerCall()
		} else {
			tp.PlayerCheck()
		}
		assert.Equal(t, entities.PokerPhaseEnd, tp.GetPhase())
		assert.Equal(t, 0, tp.GetFolded())
	})

	t.Run("success GameJudgment player win higher rank", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		player.Reset()
		dealer.Reset()
		// player: Royal Flush
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 1, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 10, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 12, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 13, false))
		// dealer: High Card
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 2, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignHeart, 5, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignDiamond, 7, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 9, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignHeart, 11, false))
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
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		player.AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		player.AddCard(entities.NewCard(entities.CardDesignHeart, 7, false))
		player.AddCard(entities.NewCard(entities.CardDesignDiamond, 9, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		// dealer: One Pair
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignHeart, 5, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignDiamond, 7, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 9, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignHeart, 11, false))
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
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 7, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 9, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 2, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 7, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 9, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 11, false))
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
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 7, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 9, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 13, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 2, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 7, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 9, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 11, false))
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
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 7, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 9, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 2, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 7, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 9, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 13, false))
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
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 1, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 7, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 9, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 13, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 7, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 9, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 11, false))
		player.EvalHand()
		dealer.EvalHand()
		assert.Equal(t, 1, tp.GameJudgment())
	})

	t.Run("success PlayerRaise advances phase", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		ok := tp.PlayerRaise(entities.PokerMinBet)
		assert.True(t, ok)
		// Should advance or end depending on dealer response
		assert.True(t, tp.GetPhase() == entities.PokerPhaseExchange || tp.GetPhase() == entities.PokerPhaseEnd)
	})

	t.Run("success PlayerRaise rejected below minimum", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		ok := tp.PlayerRaise(1)
		assert.False(t, ok)
	})

	t.Run("success Showdown pot distribution player wins", func(t *testing.T) {
		player.SetChips(500)
		dealer.SetChips(500)
		tp.Reset()
		// Navigate to Exchange phase
		if tp.GetDealerBet() > 0 {
			tp.PlayerCall()
		} else {
			tp.PlayerCheck()
		}
		// Set up deterministic hands
		player.Reset()
		dealer.Reset()
		// player: Four of a Kind
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 10, false))
		player.AddCard(entities.NewCard(entities.CardDesignClover, 10, false))
		player.AddCard(entities.NewCard(entities.CardDesignHeart, 10, false))
		player.AddCard(entities.NewCard(entities.CardDesignDiamond, 10, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))
		// dealer: Full House (rank >= TwoPair -> no exchange)
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 8, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignHeart, 8, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignDiamond, 8, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 3, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignHeart, 3, false))
		tp.PlayerStand()
		// Second bet
		if tp.GetDealerBet() > 0 {
			tp.PlayerCall()
		} else {
			tp.PlayerCheck()
		}
		assert.Equal(t, entities.PokerPhaseEnd, tp.GetPhase())
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
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 2, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignHeart, 5, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignDiamond, 7, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 9, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignHeart, 11, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 1, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 10, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 12, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 13, false))
		// Large bet should cause dealer fold
		ok := tp.PlayerBet(entities.PokerMinBet * 3)
		assert.True(t, ok)
		if tp.GetFolded() == 2 {
			assert.Equal(t, entities.PokerPhaseEnd, tp.GetPhase())
			assert.Equal(t, 1, tp.GameJudgment())
		}
	})

	t.Run("success Flush draw exchange", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		// Move to exchange phase
		if tp.GetDealerBet() > 0 {
			tp.PlayerCall()
		} else {
			tp.PlayerCheck()
		}
		// Set up dealer with 4-card flush draw
		dealer.Reset()
		dealer.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignSpade, 9, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignHeart, 3, false)) // off-suit
		tp.PlayerStand()
		// After exchange, dealer should have replaced the off-suit card
		assert.Equal(t, entities.PokerPhaseSecondBet, tp.GetPhase())
	})

	t.Run("success Straight draw exchange", func(t *testing.T) {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		// Move to exchange phase
		if tp.GetDealerBet() > 0 {
			tp.PlayerCall()
		} else {
			tp.PlayerCheck()
		}
		// Set up dealer with 4-card straight draw (5-6-7-8 + off card)
		dealer.Reset()
		dealer.AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 6, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignHeart, 7, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignDiamond, 8, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignSpade, 12, false)) // outlier
		tp.PlayerStand()
		assert.Equal(t, entities.PokerPhaseSecondBet, tp.GetPhase())
	})
}
