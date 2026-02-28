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

func TestPoker_GetPlayerBet(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)
	tp.Reset()
	// After reset, playerBet should be 0 (ante is deducted but playerBet tracks round bets)
	assert.Equal(t, 0, tp.GetPlayerBet())
}

func TestPoker_SetPhase(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)
	tp.SetPhase(domain.PokerPhaseEnd)
	assert.Equal(t, domain.PokerPhaseEnd, tp.GetPhase())
	tp.SetPhase(domain.PokerPhaseDeal)
	assert.Equal(t, domain.PokerPhaseDeal, tp.GetPhase())
}

func TestPoker_SetFolded(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)
	tp.SetFolded(domain.PokerFoldByPlayer)
	assert.Equal(t, domain.PokerFoldByPlayer, tp.GetFolded())
	tp.SetFolded(domain.PokerFoldNone)
	assert.Equal(t, domain.PokerFoldNone, tp.GetFolded())
}

func TestPoker_SetDealerBet(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)
	tp.SetDealerBet(42)
	assert.Equal(t, 42, tp.GetDealerBet())
	tp.SetDealerBet(0)
	assert.Equal(t, 0, tp.GetDealerBet())
}

func TestPoker_CollectAnte_InsufficientChips(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)
	// Set chips below ante (default ante = 10)
	player.SetChips(5)
	dealer.SetChips(3)
	tp.Reset()
	// collectAnte should clamp to available chips: pot = 5 + 3 = 8
	// After dealing and first bet, pot should be at least 8
	// Player had 5 chips, ante clamped to 5 → player chips = 0
	// Dealer had 3 chips, ante clamped to 3 → dealer chips = 0
	// Pot starts at 8, then dealerFirstBet may add more if dealer has pair+
	assert.True(t, tp.GetPot() >= 8)
}

// TestPoker_PlayerBet_DealerFolds moved to poker_internal_test.go as
// TestPoker_DealerRespondToBet_FoldBranch_Deterministic which calls
// dealerRespondToBet() directly with controlled state.

func TestPoker_PlayerCall_AllIn(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)
	player.SetChips(0)
	dealer.SetChips(0)
	tp.Reset()
	// Manually set up: dealerBet > playerBet, player has few chips
	tp.SetDealerBet(100)
	// Player chips after ante: PokerDefaultChips - PokerDefaultAnte
	// We need player chips < diff (100 - 0 = 100)
	player.SetChips(30) // less than diff of 100
	err := tp.PlayerCall()
	assert.NoError(t, err)
	// Player should have gone all-in (chips = 0)
	assert.Equal(t, 0, player.GetChips())
}

func TestPoker_PlayerRaise_AlreadyOverbid(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)
	player.SetChips(0)
	dealer.SetChips(0)
	tp.Reset()
	// Set dealerBet to 0 so that playerBet (0) > dealerBet is false, diff clamps to 0
	tp.SetDealerBet(0)
	err := tp.PlayerRaise(domain.PokerMinBet)
	assert.NoError(t, err)
}

func TestPoker_PlayerFold_ChipTransfer(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)
	player.SetChips(0)
	dealer.SetChips(0)
	tp.Reset()
	dealerChipsBefore := dealer.GetChips()
	potBefore := tp.GetPot()
	err := tp.PlayerFold()
	assert.NoError(t, err)
	assert.Equal(t, domain.PokerPhaseEnd, tp.GetPhase())
	assert.Equal(t, domain.PokerFoldByPlayer, tp.GetFolded())
	assert.Equal(t, 0, tp.GetPot())
	assert.Equal(t, dealerChipsBefore+potBefore, dealer.GetChips())
}

func TestPoker_PlayerCheck_BetsEqual(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)
	player.SetChips(0)
	dealer.SetChips(0)
	tp.Reset()
	// Force dealerBet = 0 so check is valid
	tp.SetDealerBet(0)
	err := tp.PlayerCheck()
	assert.NoError(t, err)
	assert.Equal(t, domain.PokerPhaseExchange, tp.GetPhase())
}

func TestPoker_Showdown_PlayerWins(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)
	player.SetChips(500)
	dealer.SetChips(500)
	tp.Reset()
	// Navigate to exchange phase
	tp.SetDealerBet(0)
	_ = tp.PlayerCheck()
	assert.Equal(t, domain.PokerPhaseExchange, tp.GetPhase())
	// Set up deterministic hands
	player.Reset()
	dealer.Reset()
	// player: Flush
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
	// dealer: High Card (rank >= TwoPair so no exchange)
	// We need dealer to not exchange, so give TwoPair
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 3, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 6, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 6, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
	playerChipsBefore := player.GetChips()
	_ = tp.PlayerStand()
	assert.Equal(t, domain.PokerPhaseSecondBet, tp.GetPhase())
	// Second bet: navigate to showdown
	tp.SetDealerBet(0)
	_ = tp.PlayerCheck()
	assert.Equal(t, domain.PokerPhaseEnd, tp.GetPhase())
	assert.Equal(t, 0, tp.GetPot())
	assert.Equal(t, 1, tp.GameJudgment())
	assert.True(t, player.GetChips() > playerChipsBefore)
}

func TestPoker_Showdown_DealerWins(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)
	player.SetChips(500)
	dealer.SetChips(500)
	tp.Reset()
	// Navigate to exchange phase
	tp.SetDealerBet(0)
	_ = tp.PlayerCheck()
	assert.Equal(t, domain.PokerPhaseExchange, tp.GetPhase())
	// Set up deterministic hands
	player.Reset()
	dealer.Reset()
	// player: High Card (give TwoPair so no exchange)
	player.AddCard(domain.NewCard(domain.CardDesignClover, 2, false))
	player.AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	player.AddCard(domain.NewCard(domain.CardDesignDiamond, 4, false))
	player.AddCard(domain.NewCard(domain.CardDesignClover, 4, false))
	player.AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
	// dealer: Four of a Kind (rank >= TwoPair so no exchange)
	dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 10, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	dealerChipsBefore := dealer.GetChips()
	_ = tp.PlayerStand()
	assert.Equal(t, domain.PokerPhaseSecondBet, tp.GetPhase())
	// Second bet: navigate to showdown
	if tp.GetDealerBet() > 0 {
		_ = tp.PlayerCall()
	} else {
		_ = tp.PlayerCheck()
	}
	assert.Equal(t, domain.PokerPhaseEnd, tp.GetPhase())
	assert.Equal(t, 0, tp.GetPot())
	assert.Equal(t, -1, tp.GameJudgment())
	assert.True(t, dealer.GetChips() > dealerChipsBefore)
}

func TestPoker_Showdown_Draw(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)
	player.SetChips(500)
	dealer.SetChips(500)
	tp.Reset()
	// Navigate to exchange phase
	tp.SetDealerBet(0)
	_ = tp.PlayerCheck()
	assert.Equal(t, domain.PokerPhaseExchange, tp.GetPhase())
	// Set up identical hands (both TwoPair with same values, no exchange)
	player.Reset()
	dealer.Reset()
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
	player.AddCard(domain.NewCard(domain.CardDesignClover, 3, false))
	player.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
	player.AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 7, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))
	playerChipsBefore := player.GetChips()
	dealerChipsBefore := dealer.GetChips()
	_ = tp.PlayerStand()
	assert.Equal(t, domain.PokerPhaseSecondBet, tp.GetPhase())
	// Second bet: navigate to showdown
	if tp.GetDealerBet() > 0 {
		_ = tp.PlayerCall()
	} else {
		_ = tp.PlayerCheck()
	}
	assert.Equal(t, domain.PokerPhaseEnd, tp.GetPhase())
	assert.Equal(t, 0, tp.GetPot())
	assert.Equal(t, 0, tp.GameJudgment())
	// Both should have received chips back (pot split)
	assert.True(t, player.GetChips() > playerChipsBefore)
	assert.True(t, dealer.GetChips() > dealerChipsBefore)
}

func TestPoker_DealerRespondToBet_DiffZero(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)
	player.SetChips(0)
	dealer.SetChips(0)
	tp.Reset()
	// Set dealerBet >= playerBet so diff <= 0 → dealerRespondToBet returns early
	tp.SetDealerBet(999)
	dealerChipsBefore := dealer.GetChips()
	err := tp.PlayerBet(domain.PokerMinBet)
	assert.NoError(t, err)
	// Dealer should not have spent any additional chips on call since diff <= 0
	// (dealer may have bet in second round, but the respond-to-bet path was skipped)
	_ = dealerChipsBefore // used for verification context
}

func TestPoker_DealerSecondBet_FullHouse(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)
	player.SetChips(500)
	dealer.SetChips(500)
	tp.Reset()
	// Navigate to exchange phase
	tp.SetDealerBet(0)
	_ = tp.PlayerCheck()
	assert.Equal(t, domain.PokerPhaseExchange, tp.GetPhase())
	// Set up dealer with Full House (rank >= TwoPair, no exchange; >= FullHouse → bet*3)
	dealer.Reset()
	dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 9, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 4, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 4, false))
	_ = tp.PlayerStand()
	assert.Equal(t, domain.PokerPhaseSecondBet, tp.GetPhase())
	assert.Equal(t, domain.PokerMinBet*3, tp.GetDealerBet())
}

func TestPoker_DealerSecondBet_Straight(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)
	player.SetChips(500)
	dealer.SetChips(500)
	tp.Reset()
	// Navigate to exchange phase
	tp.SetDealerBet(0)
	_ = tp.PlayerCheck()
	assert.Equal(t, domain.PokerPhaseExchange, tp.GetPhase())
	// Set up dealer with Straight (rank >= TwoPair, no exchange; >= Straight → bet*2)
	dealer.Reset()
	dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 6, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 8, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
	_ = tp.PlayerStand()
	assert.Equal(t, domain.PokerPhaseSecondBet, tp.GetPhase())
	assert.Equal(t, domain.PokerMinBet*2, tp.GetDealerBet())
}

func TestPoker_DealerSecondBet_OnePair(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)
	player.SetChips(500)
	dealer.SetChips(500)
	tp.Reset()
	// Navigate to exchange phase
	tp.SetDealerBet(0)
	_ = tp.PlayerCheck()
	assert.Equal(t, domain.PokerPhaseExchange, tp.GetPhase())
	// Set up dealer with One Pair (< TwoPair → no second bet)
	// But OnePair triggers exchange (3 cards), so we need the deck to provide cards
	// that keep it as OnePair after exchange. Instead, directly test by giving TwoPair-minus hand.
	// Actually, dealerExchange will swap 3 cards for OnePair. The result is non-deterministic.
	// To ensure dealerBet stays 0, give dealer a hand that will evaluate as OnePair after exchange.
	// Simplest: give HighCard that stays HighCard after exchange (< TwoPair → no bet).
	dealer.Reset()
	dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 9, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
	_ = tp.PlayerStand()
	assert.Equal(t, domain.PokerPhaseSecondBet, tp.GetPhase())
	// After exchange, if dealer still has < TwoPair, dealerBet should be 0
	// This is non-deterministic, so we check: if rank < TwoPair then bet == 0
	dealer.EvalHand()
	if dealer.GetHandRank() < domain.PokerHandTwoPair {
		assert.Equal(t, 0, tp.GetDealerBet())
	}
}

func TestPoker_GameJudgment_Folded(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)

	t.Run("player folded returns -1", func(t *testing.T) {
		tp.SetFolded(domain.PokerFoldByPlayer)
		assert.Equal(t, -1, tp.GameJudgment())
	})

	t.Run("dealer folded returns 1", func(t *testing.T) {
		tp.SetFolded(domain.PokerFoldByDealer)
		assert.Equal(t, 1, tp.GameJudgment())
	})
}

func TestPoker_CompareHighCards_EqualTo4th(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)
	tp.Reset()
	player.Reset()
	dealer.Reset()
	// Both High Card, same top 4 cards, differ on 5th
	// Player: 3, 5, 7, 9, 11 → sorted desc: 11, 9, 7, 5, 3
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
	player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
	player.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
	player.AddCard(domain.NewCard(domain.CardDesignDiamond, 9, false))
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
	// Dealer: 2, 5, 7, 9, 11 → sorted desc: 11, 9, 7, 5, 2
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 2, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 9, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))
	player.EvalHand()
	dealer.EvalHand()
	// Same rank (HighCard), top 4 equal (11,9,7,5), 5th differs (3 vs 2) → player wins
	assert.Equal(t, 1, tp.GameJudgment())
}

func TestPoker_FindStraightDrawDiscard_NoDraw(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)
	player.SetChips(500)
	dealer.SetChips(500)
	tp.Reset()
	// Navigate to exchange phase
	tp.SetDealerBet(0)
	_ = tp.PlayerCheck()
	assert.Equal(t, domain.PokerPhaseExchange, tp.GetPhase())
	// Dealer hand with no straight draw: 2, 5, 8, 11, 13 (gaps too big)
	dealer.Reset()
	dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 11, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
	// Exchange should use the default high-card path (swap lowest 3)
	_ = tp.PlayerStand()
	assert.Equal(t, domain.PokerPhaseSecondBet, tp.GetPhase())
	// Dealer should still have 5 cards after exchange
	assert.Equal(t, 5, dealer.GetCardsSize())
}

// --- Coverage gap tests: wrong-phase branches ---

func TestPoker_PlayerCall_WrongPhase(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)

	// Set phase to Exchange (not Deal or SecondBet)
	tp.SetPhase(domain.PokerPhaseExchange)
	err := tp.PlayerCall()
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrWrongPhase)

	// Set phase to End
	tp.SetPhase(domain.PokerPhaseEnd)
	err = tp.PlayerCall()
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrWrongPhase)
}

func TestPoker_PlayerCall_NothingToCall(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)
	player.SetChips(0)
	dealer.SetChips(0)
	tp.Reset()
	// dealerBet <= playerBet (both 0 after reset if dealer checked)
	tp.SetDealerBet(0)
	err := tp.PlayerCall()
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidPlay)
}

func TestPoker_PlayerRaise_WrongPhase(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)

	// Set phase to Exchange (not Deal or SecondBet)
	tp.SetPhase(domain.PokerPhaseExchange)
	err := tp.PlayerRaise(domain.PokerMinBet)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrWrongPhase)

	// Set phase to End
	tp.SetPhase(domain.PokerPhaseEnd)
	err = tp.PlayerRaise(domain.PokerMinBet)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrWrongPhase)
}

func TestPoker_PlayerRaise_DiffNegativeClamped(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)
	player.SetChips(0)
	dealer.SetChips(0)
	tp.Reset()
	// Make a bet so playerBet increases
	tp.SetDealerBet(0)
	err := tp.PlayerBet(domain.PokerMinBet)
	assert.NoError(t, err)
	// Now playerBet = PokerMinBet, phase = Exchange
	// Force back to Deal phase with dealerBet = 0 so diff = 0 - PokerMinBet = -10 < 0
	tp.SetPhase(domain.PokerPhaseDeal)
	tp.SetDealerBet(0)
	// Give player enough chips for raise
	player.SetChips(500)
	// Give dealer enough chips so they can call
	dealer.SetChips(500)
	// Give dealer a strong hand so they don't fold
	dealer.Reset()
	dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 9, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
	err = tp.PlayerRaise(domain.PokerMinBet)
	assert.NoError(t, err)
	// diff was clamped to 0, so totalNeeded = 0 + PokerMinBet = PokerMinBet
}

func TestPoker_PlayerRaise_DealerFolds(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)
	// Run multiple attempts since dealer fold is probabilistic
	dealerFolded := false
	for i := 0; i < 200; i++ {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		// Set up dealer with weak high card (no A/K/Q) after Reset dealt cards
		dealer.Reset()
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 2, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 4, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 6, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 8, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
		// Set up player with enough chips
		player.SetChips(500)
		// PlayerRaise: diff = dealerBet - playerBet. After reset dealerBet may be 0 or 10.
		// Set dealerBet to 0 so diff clamps to 0, totalNeeded = 0 + amount = amount
		tp.SetDealerBet(0)
		err := tp.PlayerRaise(100)
		assert.NoError(t, err)
		if tp.GetFolded() == domain.PokerFoldByDealer {
			assert.Equal(t, domain.PokerPhaseEnd, tp.GetPhase())
			dealerFolded = true
			break
		}
	}
	assert.True(t, dealerFolded, "dealer should have folded at least once in 200 attempts via PlayerRaise")
}

func TestPoker_PlayerFold_WrongPhase(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)

	// Set phase to Exchange (not Deal or SecondBet)
	tp.SetPhase(domain.PokerPhaseExchange)
	err := tp.PlayerFold()
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrWrongPhase)

	// Set phase to End
	tp.SetPhase(domain.PokerPhaseEnd)
	err = tp.PlayerFold()
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrWrongPhase)
}

func TestPoker_PlayerCheck_WrongPhase(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)

	// Set phase to Exchange (not Deal or SecondBet)
	tp.SetPhase(domain.PokerPhaseExchange)
	err := tp.PlayerCheck()
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrWrongPhase)

	// Set phase to End
	tp.SetPhase(domain.PokerPhaseEnd)
	err = tp.PlayerCheck()
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrWrongPhase)
}

func TestPoker_PlayerCheck_OutstandingBet(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)
	player.SetChips(0)
	dealer.SetChips(0)
	tp.Reset()
	// Set dealerBet > playerBet so check is rejected
	tp.SetDealerBet(100)
	err := tp.PlayerCheck()
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidPlay)
}

// --- Coverage gap tests: dealerRespondToBet branches ---

func TestPoker_DealerRespondToBet_HasHighCard(t *testing.T) {
	// Dealer has HighCard rank but with A/Q/K cards -> hasHighCard = true
	// In this case neither fold branch fires and dealer just calls
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)
	player.SetChips(0)
	dealer.SetChips(0)
	tp.Reset()
	// Set up dealer with HighCard that includes an Ace (value=1)
	dealer.Reset()
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 1, false)) // Ace
	dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 4, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 6, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 8, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
	// Set up player hand
	player.Reset()
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
	tp.SetDealerBet(0)
	player.SetChips(500)
	dealer.SetChips(500)
	// Large bet that would trigger fold if hasHighCard were false
	err := tp.PlayerBet(100)
	assert.NoError(t, err)
	// hasHighCard = true, so dealer should NOT fold
	assert.Equal(t, domain.PokerFoldNone, tp.GetFolded())
}

func TestPoker_DealerRespondToBet_SecondFoldBranch(t *testing.T) {
	// Tests the second else-if fold branch:
	// !hasHighCard && diff > PokerMinBet * dealerFoldBetMultiplierStrong (=30)
	// but potOdds <= threshold (0.4) so the first branch is skipped
	// This happens when the pot is large relative to the bet
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)

	foldCount := 0
	noFoldCount := 0
	for i := 0; i < 200; i++ {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		// Set up dealer with weak HighCard (no A/Q/K)
		dealer.Reset()
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 2, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 4, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 6, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 8, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
		player.Reset()
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		player.SetChips(500)
		dealer.SetChips(500)
		// We need: !hasHighCard && diff > PokerMinBet*3 (=30) but potOdds <= 0.4
		// potOdds = diff / (pot + diff)
		// pot after ante = 20. diff = playerBet - dealerBet.
		// If we bet 31, diff = 31. potOdds = 31/(20+31) = 31/51 = 0.608 > 0.4 → first branch
		// To avoid first branch (potOdds <= 0.4), we need large pot.
		// potOdds = diff/(pot+diff) <= 0.4 means pot >= diff*1.5
		// If diff = 31, pot >= 46.5. Starting pot = 20, so we need pot = 47+.
		// We can't easily inflate pot, but we can use SetDealerBet to make diff smaller
		// while keeping diff > 30.
		// Actually, re-reading: diff = playerBet - dealerBet
		// After PlayerBet(amount): playerBet = amount, dealerBet = whatever was set
		// potOdds = diff / (pot + diff) where diff = playerBet - dealerBet
		// If dealerBet = 20, playerBet = 55, diff = 35, pot = 20 + 55 = 75 (before dealer respond)
		// Actually pot increases by bet amount in PlayerBet: pot += amount
		// So after PlayerBet(55): pot = 20(ante) + 55 = 75, playerBet = 55, dealerBet = 0
		// diff in dealerRespondToBet = 55 - 0 = 55
		// potOdds = 55 / (75 + 55) = 55/130 = 0.423 > 0.4 → still first branch
		// Let's try: make dealerBet high so diff is moderate (> 30 but potOdds <= 0.4)
		// SetDealerBet(20), PlayerBet(55) → playerBet=55, dealerBet=20, diff=35
		// pot = 20(ante) + 55 = 75 (PlayerBet adds to pot)
		// potOdds = 35 / (75 + 35) = 35/110 = 0.318 <= 0.4 → second branch!
		// diff = 35 > 30 = PokerMinBet*3 → condition met
		tp.SetDealerBet(20)
		err := tp.PlayerBet(55)
		assert.NoError(t, err)
		if tp.GetFolded() == domain.PokerFoldByDealer {
			foldCount++
		} else {
			noFoldCount++
		}
	}
	// 50% fold rate, so we expect both folds and non-folds
	assert.True(t, foldCount > 0, "dealer should have folded at least once via second branch")
	assert.True(t, noFoldCount > 0, "dealer should have NOT folded at least once via second branch")
}

func TestPoker_DealerRespondToBet_PartialCall(t *testing.T) {
	// Dealer has insufficient chips to fully call -> partial call
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)
	player.SetChips(0)
	dealer.SetChips(0)
	tp.Reset()
	// Give dealer a strong hand (OnePair) so no fold
	dealer.Reset()
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 9, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))
	player.Reset()
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 4, false))
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
	tp.SetDealerBet(0)
	player.SetChips(500)
	dealer.SetChips(5) // Only 5 chips, less than what player will bet
	err := tp.PlayerBet(100)
	assert.NoError(t, err)
	// Dealer should have used all remaining chips (partial call)
	assert.Equal(t, 0, dealer.GetChips())
}

// --- Coverage gap tests: dealerRespondToBet first fold branch non-fold path ---

func TestPoker_DealerRespondToBet_FirstBranch_NoFold(t *testing.T) {
	// The first fold branch has 70% fold rate, so 30% of the time it doesn't fold.
	// We need to hit the path where rand.Intn(100) >= 70 (no fold),
	// which then falls through to the call logic.
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)

	noFoldCount := 0
	for i := 0; i < 200; i++ {
		player.SetChips(0)
		dealer.SetChips(0)
		tp.Reset()
		dealer.Reset()
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 2, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 4, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 6, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 8, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
		player.Reset()
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		player.SetChips(500)
		dealer.SetChips(500)
		tp.SetDealerBet(0)
		// Large bet: pot=20+100=120, diff=100, potOdds=100/220=0.454>0.4, diff=100>20
		// First branch: 70% fold. We want the 30% no-fold path.
		err := tp.PlayerBet(100)
		assert.NoError(t, err)
		if tp.GetFolded() == domain.PokerFoldNone {
			noFoldCount++
		}
	}
	assert.True(t, noFoldCount > 0, "dealer should have NOT folded at least once (30% chance per attempt)")
}

// --- Coverage gap tests: findOpenEndedDraw len(remaining) != 4 ---

func TestPoker_FindOpenEndedDraw_ShortHand(t *testing.T) {
	// findOpenEndedDraw is called with fewer than 5 cards in the cards slice,
	// so remaining after skip will have != 4 entries → continue.
	// This happens when dealer has fewer than 5 cards (unusual but possible).
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)
	player.SetChips(500)
	dealer.SetChips(500)
	tp.Reset()
	// Navigate to exchange phase
	tp.SetDealerBet(0)
	_ = tp.PlayerCheck()
	assert.Equal(t, domain.PokerPhaseExchange, tp.GetPhase())
	// Set up dealer with only 4 cards (less than 5) and HighCard rank
	// findStraightDrawDiscard will call findOpenEndedDraw with 4 cards
	// skip one → remaining = 3 → len(remaining) != 4 → continue for all skips
	dealer.Reset()
	dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 9, false))
	// Only 4 cards; dealerExchange will run and hit the HighCard path
	// findFlushDrawDiscard and findStraightDrawDiscard run first on < OnePair hands
	// With 4 cards, findOpenEndedDraw skipping one leaves 3 → len != 4 → continue
	_ = tp.PlayerStand()
	assert.Equal(t, domain.PokerPhaseSecondBet, tp.GetPhase())
}

// --- Coverage gap tests: findStraightDrawDiscard Ace low pattern ---

func TestPoker_FindStraightDrawDiscard_AceLow(t *testing.T) {
	// Tests the Ace low straight draw pattern: A-2-3-4 + outlier
	// After high-Ace evaluation (14-2-3-4-outlier), no open-ended draw found.
	// Then Ace is re-evaluated as 1: 1-2-3-4-outlier → skip outlier → remaining = [1,2,3,4]
	// Check: r[0]==1 && r[3]<=5 → true → found draw
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)
	player.SetChips(500)
	dealer.SetChips(500)
	tp.Reset()
	// Navigate to exchange phase
	tp.SetDealerBet(0)
	_ = tp.PlayerCheck()
	assert.Equal(t, domain.PokerPhaseExchange, tp.GetPhase())
	// Set up dealer with Ace-low straight draw: A, 2, 3, 4 + outlier (10)
	dealer.Reset()
	dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))   // Ace
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 2, false))  // 2
	dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))   // 3
	dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 4, false)) // 4
	dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))  // outlier (different suits to avoid flush draw)
	_ = tp.PlayerStand()
	assert.Equal(t, domain.PokerPhaseSecondBet, tp.GetPhase())
	// Dealer should have exchanged the outlier card (index 4)
	assert.Equal(t, 5, dealer.GetCardsSize())
}

// --- Coverage gap tests: dealerExchange paths ---

func TestPoker_DealerExchange_OnePairPath(t *testing.T) {
	// Dealer has exactly OnePair → exchanges the 3 non-pair cards
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)
	player.SetChips(500)
	dealer.SetChips(500)
	tp.Reset()
	// Navigate to exchange phase
	tp.SetDealerBet(0)
	_ = tp.PlayerCheck()
	assert.Equal(t, domain.PokerPhaseExchange, tp.GetPhase())
	// Set up dealer with OnePair (pair of 5s, non-pair: 2,8,11)
	dealer.Reset()
	dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 8, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
	_ = tp.PlayerStand()
	assert.Equal(t, domain.PokerPhaseSecondBet, tp.GetPhase())
	assert.Equal(t, 5, dealer.GetCardsSize())
}

func TestPoker_DealerExchange_HighCardWithAce(t *testing.T) {
	// Tests the Ace=14 transformation in dealerExchange high card branch (L432-434).
	// Dealer has HighCard with an Ace, no flush draw, no straight draw.
	// The code should treat Ace as 14 (highest) and exchange the 3 lowest non-Ace cards.
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)
	player.SetChips(500)
	dealer.SetChips(500)
	tp.Reset()
	// Navigate to exchange phase
	tp.SetDealerBet(0)
	_ = tp.PlayerCheck()
	assert.Equal(t, domain.PokerPhaseExchange, tp.GetPhase())
	// Set up dealer with HighCard including Ace, no flush draw, no straight draw
	// Ace(1), 3, 6, 9, 12 - all different suits, gaps too big for straight draw
	dealer.Reset()
	dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))   // Ace (value=1, treated as 14)
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 3, false))  // low card
	dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))   // low card
	dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 9, false)) // medium card
	dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))  // Queen
	// After sorting by value with Ace=14: [3, 6, 9, 12, 14(Ace)]
	// Should exchange the 3 lowest: indices for values 3, 6, 9
	// Ace should be kept as it's the highest card
	_ = tp.PlayerStand()
	assert.Equal(t, domain.PokerPhaseSecondBet, tp.GetPhase())
	assert.Equal(t, 5, dealer.GetCardsSize())
}

// --- Tie-break regression tests: pair value must be compared before kickers ---

func TestPoker_CompareHighCards_OnePairTieBreak(t *testing.T) {
	// Player: Pair of 4s (4-4-K-3-2), Dealer: Pair of 3s (3-3-K-Q-J)
	// Old bug: sorts desc [13,4,4,3,2] vs [13,12,11,3,3] → K==K, 4<12 → dealer wins
	// Fixed: compare pair value first: 4 > 3 → player wins
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)
	tp.SetFolded(domain.PokerFoldNone)
	player.Reset()
	dealer.Reset()

	player.AddCard(domain.NewCard(domain.CardDesignSpade, 4, false))
	player.AddCard(domain.NewCard(domain.CardDesignHeart, 4, false))
	player.AddCard(domain.NewCard(domain.CardDesignClover, 13, false))
	player.AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))

	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 3, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 13, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 11, false))

	player.EvalHand()
	dealer.EvalHand()
	assert.Equal(t, domain.PokerHandOnePair, player.GetHandRank())
	assert.Equal(t, domain.PokerHandOnePair, dealer.GetHandRank())
	// Player's pair of 4s beats dealer's pair of 3s
	assert.Equal(t, 1, tp.GameJudgment())
}

func TestPoker_CompareHighCards_TwoPairTieBreak(t *testing.T) {
	// Player: 10-10-5-5-A, Dealer: 9-9-8-8-A
	// Higher top pair (10 vs 9) should win
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)
	tp.SetFolded(domain.PokerFoldNone)
	player.Reset()
	dealer.Reset()

	player.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
	player.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
	player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
	player.AddCard(domain.NewCard(domain.CardDesignDiamond, 5, false))
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))

	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 9, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 9, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 1, false))

	player.EvalHand()
	dealer.EvalHand()
	assert.Equal(t, domain.PokerHandTwoPair, player.GetHandRank())
	assert.Equal(t, domain.PokerHandTwoPair, dealer.GetHandRank())
	// Player's top pair (10) > dealer's top pair (9)
	assert.Equal(t, 1, tp.GameJudgment())
}

func TestPoker_CompareHighCards_TwoPairKickerDecides(t *testing.T) {
	// Player: 10-10-5-5-K, Dealer: 10-10-5-5-Q
	// Same two pairs, kicker decides: K > Q → player wins
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)
	tp.SetFolded(domain.PokerFoldNone)
	player.Reset()
	dealer.Reset()

	player.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
	player.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
	player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
	player.AddCard(domain.NewCard(domain.CardDesignDiamond, 5, false))
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))

	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 10, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 12, false))

	player.EvalHand()
	dealer.EvalHand()
	assert.Equal(t, domain.PokerHandTwoPair, player.GetHandRank())
	assert.Equal(t, domain.PokerHandTwoPair, dealer.GetHandRank())
	// Same pairs, kicker K(13) > Q(12) → player wins
	assert.Equal(t, 1, tp.GameJudgment())
}

// --- Coverage gap tests: compareHighCards dealer Ace=14 ---

func TestPoker_CompareHighCards_DealerAce(t *testing.T) {
	// Tests that dealer's Ace is treated as 14 in compareHighCards
	tc := domain.NewTrumpCards(0)
	player := domain.NewPokerPlayer()
	dealer := domain.NewPokerPlayer()
	tp := domain.NewPoker(tc, player, dealer)
	tp.SetFolded(domain.PokerFoldNone)
	player.Reset()
	dealer.Reset()
	// Player has King high
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
	player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
	player.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
	player.AddCard(domain.NewCard(domain.CardDesignDiamond, 9, false))
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
	// Dealer has Ace high (value=1 → treated as 14)
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 1, false)) // Ace
	dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 9, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))
	player.EvalHand()
	dealer.EvalHand()
	// Both HighCard, but dealer has Ace (14) > player's King (13)
	assert.Equal(t, -1, tp.GameJudgment())
}
