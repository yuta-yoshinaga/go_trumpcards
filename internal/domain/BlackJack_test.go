package domain_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestNewDefaultBlackJack(t *testing.T) {
	bj := domain.NewDefaultBlackJack()
	assert.NotNil(t, bj)
	assert.NotNil(t, bj.GetPlayer())
	assert.NotNil(t, bj.GetDealer())
	assert.Equal(t, false, bj.GetGameEndFlag())
	assert.Equal(t, domain.BJPhaseBet, bj.GetPhase())
	assert.Equal(t, domain.BJDefaultChips, bj.GetPlayer().GetChips())
	assert.Equal(t, domain.BJDefaultChips, bj.GetDealer().GetChips())
}

func TestBlackJack_Reset(t *testing.T) {
	bj := domain.NewDefaultBlackJack()
	bj.Reset()
	assert.Equal(t, domain.BJPhaseBet, bj.GetPhase())
	assert.False(t, bj.GetGameEndFlag())
	assert.Equal(t, 0, bj.GetCurrentHandIdx())
	assert.Equal(t, 0, bj.GetInsuranceBet())
	assert.False(t, bj.IsInsuranceAvailable())
	assert.Equal(t, 1, len(bj.GetPlayerHands()))
	assert.Equal(t, 0, bj.GetPlayer().GetCardsSize())
}

func TestBlackJack_ResetChipsIfZero(t *testing.T) {
	bj := domain.NewDefaultBlackJack()
	bj.GetPlayer().SetChips(0)
	bj.GetDealer().SetChips(0)
	bj.Reset()
	assert.Equal(t, domain.BJDefaultChips, bj.GetPlayer().GetChips())
	assert.Equal(t, domain.BJDefaultChips, bj.GetDealer().GetChips())
}

func TestBlackJack_ResetChipsBelowMinBet(t *testing.T) {
	bj := domain.NewDefaultBlackJack()
	// チップが最低ベット額未満ならリセットされる
	bj.GetPlayer().SetChips(domain.BJMinBet - 1)
	bj.GetDealer().SetChips(domain.BJMinBet - 1)
	bj.Reset()
	assert.Equal(t, domain.BJDefaultChips, bj.GetPlayer().GetChips())
	assert.Equal(t, domain.BJDefaultChips, bj.GetDealer().GetChips())

	// ちょうど最低ベット額ならリセットされない
	bj.GetPlayer().SetChips(domain.BJMinBet)
	bj.GetDealer().SetChips(domain.BJMinBet)
	bj.Reset()
	assert.Equal(t, domain.BJMinBet, bj.GetPlayer().GetChips())
	assert.Equal(t, domain.BJMinBet, bj.GetDealer().GetChips())
}

func TestBlackJack_PlayerBet(t *testing.T) {
	t.Run("successful bet", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		err := bj.PlayerBet(100)
		assert.NoError(t, err)
		assert.Equal(t, 2, bj.GetPlayerHands()[0].GetCardsSize())
		assert.Equal(t, 2, bj.GetDealer().GetCardsSize())
		assert.NotEqual(t, domain.BJPhaseBet, bj.GetPhase())
	})
	t.Run("bet below minimum", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		err := bj.PlayerBet(5)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidAmount)
		assert.Equal(t, domain.BJPhaseBet, bj.GetPhase())
	})
	t.Run("bet with insufficient chips", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		bj.GetPlayer().SetChips(50)
		err := bj.PlayerBet(100)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInsufficientChips)
	})
	t.Run("bet not multiple of min bet", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		err := bj.PlayerBet(15)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidAmount)
		assert.Equal(t, domain.BJPhaseBet, bj.GetPhase())
	})
	t.Run("bet in wrong phase", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		_ = bj.PlayerBet(100)
		err := bj.PlayerBet(100)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrWrongPhase)
	})
}

func TestBlackJack_InsurancePhase(t *testing.T) {
	t.Run("accept insurance in wrong phase returns error", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		err := bj.PlayerInsurance()
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrWrongPhase)
	})
	t.Run("decline insurance in wrong phase returns error", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		err := bj.PlayerDeclineInsurance()
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrWrongPhase)
		assert.Equal(t, domain.BJPhaseBet, bj.GetPhase())
	})
}

func TestBlackJack_PlayerHitInWrongPhase(t *testing.T) {
	bj := domain.NewDefaultBlackJack()
	bj.Reset()
	err := bj.PlayerHit()
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrWrongPhase)
	assert.Equal(t, domain.BJPhaseBet, bj.GetPhase())
}

func TestBlackJack_PlayerStandViaHand(t *testing.T) {
	playerCards := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 9, false),
		domain.NewCard(domain.CardDesignHeart, 8, false),
	}
	dealerCards := []*domain.Card{
		domain.NewCard(domain.CardDesignClover, 2, false),
		domain.NewCard(domain.CardDesignClover, 10, false),
		domain.NewCard(domain.CardDesignClover, 11, false),
	}
	bj, _, _ := setupDeterministicBJ(1000, playerCards, dealerCards, 100)
	err := bj.PlayerStand()
	assert.NoError(t, err)
	// Player score 17, dealer score 22 (bust) → player wins
	assert.Equal(t, domain.GameResultWin, bj.GameJudgment())
}

func TestBlackJack_GameJudgmentCases(t *testing.T) {
	t.Run("player lose bust", func(t *testing.T) {
		playerCards := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 2, false),
			domain.NewCard(domain.CardDesignSpade, 10, false),
			domain.NewCard(domain.CardDesignSpade, 11, false),
		}
		dealerCards := []*domain.Card{
			domain.NewCard(domain.CardDesignClover, 2, false),
			domain.NewCard(domain.CardDesignClover, 10, false),
			domain.NewCard(domain.CardDesignClover, 11, false),
		}
		bj, _, _ := setupDeterministicBJ(1000, playerCards, dealerCards, 100)
		assert.Equal(t, domain.GameResultLose, bj.GameJudgment())
	})
	t.Run("player lose dealer higher", func(t *testing.T) {
		playerCards := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 2, false),
			domain.NewCard(domain.CardDesignSpade, 10, false),
			domain.NewCard(domain.CardDesignSpade, 11, false),
		}
		dealerCards := []*domain.Card{
			domain.NewCard(domain.CardDesignClover, 1, false),
			domain.NewCard(domain.CardDesignClover, 10, false),
			domain.NewCard(domain.CardDesignClover, 11, false),
		}
		bj, _, _ := setupDeterministicBJ(1000, playerCards, dealerCards, 100)
		assert.Equal(t, domain.GameResultLose, bj.GameJudgment())
	})
	t.Run("player win dealer bust", func(t *testing.T) {
		playerCards := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignSpade, 10, false),
			domain.NewCard(domain.CardDesignSpade, 11, false),
		}
		dealerCards := []*domain.Card{
			domain.NewCard(domain.CardDesignClover, 2, false),
			domain.NewCard(domain.CardDesignClover, 10, false),
			domain.NewCard(domain.CardDesignClover, 11, false),
		}
		bj, _, _ := setupDeterministicBJ(1000, playerCards, dealerCards, 100)
		assert.Equal(t, domain.GameResultWin, bj.GameJudgment())
	})
	t.Run("draw", func(t *testing.T) {
		playerCards := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignSpade, 10, false),
		}
		dealerCards := []*domain.Card{
			domain.NewCard(domain.CardDesignClover, 1, false),
			domain.NewCard(domain.CardDesignClover, 10, false),
		}
		bj, _, _ := setupDeterministicBJ(1000, playerCards, dealerCards, 100)
		assert.Equal(t, domain.GameResultDraw, bj.GameJudgment())
	})
	t.Run("player win natural BJ vs dealer 3 cards", func(t *testing.T) {
		playerCards := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignSpade, 10, false),
		}
		dealerCards := []*domain.Card{
			domain.NewCard(domain.CardDesignClover, 1, false),
			domain.NewCard(domain.CardDesignClover, 10, false),
			domain.NewCard(domain.CardDesignClover, 11, false),
		}
		bj, _, _ := setupDeterministicBJ(1000, playerCards, dealerCards, 100)
		assert.Equal(t, domain.GameResultWin, bj.GameJudgment())
	})
	t.Run("player win higher score", func(t *testing.T) {
		playerCards := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignSpade, 10, false),
		}
		dealerCards := []*domain.Card{
			domain.NewCard(domain.CardDesignClover, 9, false),
			domain.NewCard(domain.CardDesignClover, 10, false),
		}
		bj, _, _ := setupDeterministicBJ(1000, playerCards, dealerCards, 100)
		assert.Equal(t, domain.GameResultWin, bj.GameJudgment())
	})
	t.Run("player lose dealer higher score", func(t *testing.T) {
		playerCards := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 9, false),
			domain.NewCard(domain.CardDesignSpade, 10, false),
		}
		dealerCards := []*domain.Card{
			domain.NewCard(domain.CardDesignClover, 1, false),
			domain.NewCard(domain.CardDesignClover, 10, false),
		}
		bj, _, _ := setupDeterministicBJ(1000, playerCards, dealerCards, 100)
		assert.Equal(t, domain.GameResultLose, bj.GameJudgment())
	})
}

func TestBlackJack_DoubleDown_WrongPhase(t *testing.T) {
	bj := domain.NewDefaultBlackJack()
	bj.Reset()
	err := bj.PlayerDoubleDown()
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrWrongPhase)
}

func TestBlackJack_Split_WrongPhase(t *testing.T) {
	bj := domain.NewDefaultBlackJack()
	bj.Reset()
	err := bj.PlayerSplit()
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrWrongPhase)
}

func TestBlackJack_GetterMethods(t *testing.T) {
	bj := domain.NewDefaultBlackJack()
	bj.Reset()
	assert.NotNil(t, bj.GetPlayer())
	assert.NotNil(t, bj.GetDealer())
	assert.NotNil(t, bj.GetPlayerHands())
	assert.NotNil(t, bj.GetTrumpCards())
	assert.Equal(t, 0, bj.GetCurrentHandIdx())
	assert.Equal(t, 0, bj.GetInsuranceBet())
	assert.False(t, bj.IsInsuranceAvailable())
	assert.Equal(t, domain.BJPhaseBet, bj.GetPhase())
}

func TestBlackJack_GameJudgmentForHand(t *testing.T) {
	bj := domain.NewDefaultBlackJack()
	t.Run("invalid hand index negative", func(t *testing.T) {
		assert.Equal(t, domain.GameResultLose, bj.GameJudgmentForHand(-1))
	})
	t.Run("invalid hand index out of range", func(t *testing.T) {
		assert.Equal(t, domain.GameResultLose, bj.GameJudgmentForHand(100))
	})
}

func TestBlackJack_DrawCardNilSafety(t *testing.T) {
	t.Run("bet with exhausted deck refunds chips", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		player := domain.NewBlackJackPlayer()
		dealer := domain.NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := domain.NewBlackJack(tc, player, dealer)
		bj.Reset()

		// デッキを全て引き切る
		for i := 0; i < 52; i++ {
			tc.DrawCard()
		}

		// デッキ枯渇時のベットはエラーを返し、チップが返却される
		err := bj.PlayerBet(100)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrDeckExhausted)
		assert.Equal(t, 1000, player.GetChips())
		assert.Equal(t, domain.BJPhaseBet, bj.GetPhase())
	})

	t.Run("hit with exhausted deck does not panic", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		player := domain.NewBlackJackPlayer()
		dealer := domain.NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := domain.NewBlackJack(tc, player, dealer)
		bj.Reset()

		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		hand.AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
		bj.SetPhase(domain.BJPhaseAction)

		// デッキを全て引き切る
		for i := 0; i < 52; i++ {
			tc.DrawCard()
		}

		// デッキ枯渇後のヒットでパニックしない
		assert.NotPanics(t, func() {
			err := bj.PlayerHit()
			assert.Error(t, err)
			assert.ErrorIs(t, err, domain.ErrDeckExhausted)
		})
		// カードが追加されないことを確認
		assert.Equal(t, 2, hand.GetCardsSize())
	})

	t.Run("dealer hit with exhausted deck does not panic", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		player := domain.NewBlackJackPlayer()
		dealer := domain.NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := domain.NewBlackJack(tc, player, dealer)
		bj.Reset()

		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		hand.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 6, false))
		bj.SetPhase(domain.BJPhaseAction)

		// デッキを全て引き切る
		for i := 0; i < 52; i++ {
			tc.DrawCard()
		}

		// デッキ枯渇後のスタンド→ディーラーヒットでパニックしない
		assert.NotPanics(t, func() {
			_ = bj.PlayerStand()
		})
		assert.True(t, bj.GetGameEndFlag())
	})
}

func TestBlackJack_ErrorReturns(t *testing.T) {
	t.Run("bet failure returns ErrInvalidAmount", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		err := bj.PlayerBet(5) // below minimum
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidAmount)
		assert.Contains(t, err.Error(), "Invalid bet amount.")
	})
	t.Run("bet wrong phase returns ErrWrongPhase", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		_ = bj.PlayerBet(100)
		err := bj.PlayerBet(100)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrWrongPhase)
		assert.Contains(t, err.Error(), "Bet is only allowed during the bet phase.")
	})
	t.Run("bet insufficient chips returns ErrInsufficientChips", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		bj.GetPlayer().SetChips(50)
		err := bj.PlayerBet(100)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInsufficientChips)
		assert.Contains(t, err.Error(), "Insufficient chips.")
	})
	t.Run("successful bet returns nil", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		err := bj.PlayerBet(100)
		assert.NoError(t, err)
	})
	t.Run("double down failure returns ErrWrongPhase", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		err := bj.PlayerDoubleDown()
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrWrongPhase)
		assert.Contains(t, err.Error(), "Double down is not allowed now.")
	})
	t.Run("split failure returns ErrWrongPhase", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		err := bj.PlayerSplit()
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrWrongPhase)
		assert.Contains(t, err.Error(), "Split is not allowed now.")
	})
	t.Run("insurance failure returns ErrWrongPhase", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		err := bj.PlayerInsurance()
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrWrongPhase)
		assert.Contains(t, err.Error(), "Insurance is not available now.")
	})
}

func TestBlackJack_MaxSplitHandsLimit(t *testing.T) {
	// Test that split is rejected when playerHands already has BJMaxHands entries.
	// We simulate this by manually adding dummy hands to the slice.
	tc := domain.NewTrumpCards(0)
	player := domain.NewBlackJackPlayer()
	dealer := domain.NewBlackJackPlayer()
	player.SetChips(10000)
	dealer.SetChips(10000)
	bj := domain.NewBlackJack(tc, player, dealer)
	bj.Reset()

	hand := bj.GetPlayerHands()[0]
	hand.SetBet(100)
	hand.AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
	hand.AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
	bj.SetPhase(domain.BJPhaseAction)

	// First split should succeed (1 → 2 hands)
	err := bj.PlayerSplit()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(bj.GetPlayerHands()))

	// Manually add dummy hands to reach BJMaxHands
	for len(bj.GetPlayerHands()) < domain.BJMaxHands {
		dummyHand := domain.NewBlackJackHand()
		dummyHand.SetBet(100)
		dummyHand.SetStood(true)
		bj.SetPlayerHands(append(bj.GetPlayerHands(), dummyHand))
	}
	assert.Equal(t, domain.BJMaxHands, len(bj.GetPlayerHands()))

	// Now split should fail with max hands error
	err = bj.PlayerSplit()
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidPlay)
	assert.Contains(t, err.Error(), "Maximum number of hands reached.")
}

func TestBlackJack_OldFlow(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	player := domain.NewBlackJackPlayer()
	dealer := domain.NewBlackJackPlayer()
	tb := domain.NewBlackJack(tc, player, dealer)
	t.Run("success Reset", func(t *testing.T) {
		tb.Reset()
	})
	t.Run("success GetGameEndFlag", func(t *testing.T) {
		assert.Equal(t, false, tb.GetGameEndFlag())
	})
	t.Run("success GetPlayer", func(t *testing.T) {
		assert.NotEmpty(t, tb.GetPlayer())
	})
	t.Run("success GetDealer", func(t *testing.T) {
		assert.NotEmpty(t, tb.GetDealer())
	})
	t.Run("success PlayerHit", func(t *testing.T) {
		_ = tb.PlayerHit()
	})
	t.Run("success PlayerStand", func(t *testing.T) {
		_ = tb.PlayerStand()
	})
	t.Run("success DealerHit", func(t *testing.T) {
		tb.DealerHit()
	})
	t.Run("success DealerStand", func(t *testing.T) {
		tb.DealerStand()
	})
}

// setupDeterministicBJ sets up a BlackJack in BJPhaseAction with specific cards.
// This avoids the randomness from PlayerBet's card dealing.
func setupDeterministicBJ(
	playerChips int,
	playerCards []*domain.Card,
	dealerCards []*domain.Card,
	bet int,
) (*domain.BlackJack, *domain.BlackJackPlayer, *domain.BlackJackPlayer) {
	tc := domain.NewTrumpCards(0)
	player := domain.NewBlackJackPlayer()
	dealer := domain.NewBlackJackPlayer()
	player.SetChips(playerChips)
	dealer.SetChips(1000)
	bj := domain.NewBlackJack(tc, player, dealer)
	bj.Reset()

	// Set up hand
	hand := bj.GetPlayerHands()[0]
	hand.SetBet(bet)
	for _, c := range playerCards {
		hand.AddCard(c)
	}
	// Sync player cards
	for _, c := range playerCards {
		player.AddCard(c)
	}
	for _, c := range dealerCards {
		dealer.AddCard(c)
	}
	bj.SetPhase(domain.BJPhaseAction)

	return bj, player, dealer
}

func TestBlackJack_FullBettingFlow(t *testing.T) {
	t.Run("normal win payout", func(t *testing.T) {
		bj, player, _ := setupDeterministicBJ(
			900, // after 100 bet
			[]*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 10, false),
				domain.NewCard(domain.CardDesignHeart, 10, false),
			},
			[]*domain.Card{
				domain.NewCard(domain.CardDesignClover, 9, false),
				domain.NewCard(domain.CardDesignDiamond, 10, false),
			},
			100,
		)
		assert.NoError(t, bj.PlayerStand())
		assert.True(t, bj.GetGameEndFlag())
		assert.Equal(t, domain.BJPhaseEnd, bj.GetPhase())
		assert.Equal(t, domain.GameResultWin, bj.GameJudgment())
		// Normal win: bet*2 = 200 returned, 900 + 200 = 1100
		assert.Equal(t, 1100, player.GetChips())
	})

	t.Run("natural BJ 3:2 payout", func(t *testing.T) {
		bj, player, _ := setupDeterministicBJ(
			900,
			[]*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 1, false),
				domain.NewCard(domain.CardDesignHeart, 10, false),
			},
			[]*domain.Card{
				domain.NewCard(domain.CardDesignClover, 9, false),
				domain.NewCard(domain.CardDesignDiamond, 10, false),
			},
			100,
		)
		assert.NoError(t, bj.PlayerStand())
		assert.True(t, bj.GetGameEndFlag())
		assert.Equal(t, domain.GameResultWin, bj.GameJudgment())
		// Natural BJ 3:2: 100 + 150 = 250 returned, 900 + 250 = 1150
		assert.Equal(t, 1150, player.GetChips())
	})

	t.Run("draw payout", func(t *testing.T) {
		bj, player, _ := setupDeterministicBJ(
			900,
			[]*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 10, false),
				domain.NewCard(domain.CardDesignHeart, 9, false),
			},
			[]*domain.Card{
				domain.NewCard(domain.CardDesignClover, 9, false),
				domain.NewCard(domain.CardDesignDiamond, 10, false),
			},
			100,
		)
		assert.NoError(t, bj.PlayerStand())
		assert.True(t, bj.GetGameEndFlag())
		assert.Equal(t, domain.GameResultDraw, bj.GameJudgment())
		// Draw: bet returned, 900 + 100 = 1000
		assert.Equal(t, 1000, player.GetChips())
	})

	t.Run("lose payout", func(t *testing.T) {
		bj, player, _ := setupDeterministicBJ(
			900,
			[]*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 8, false),
				domain.NewCard(domain.CardDesignHeart, 9, false),
			},
			[]*domain.Card{
				domain.NewCard(domain.CardDesignClover, 10, false),
				domain.NewCard(domain.CardDesignDiamond, 10, false),
			},
			100,
		)
		assert.NoError(t, bj.PlayerStand())
		assert.True(t, bj.GetGameEndFlag())
		assert.Equal(t, domain.GameResultLose, bj.GameJudgment())
		// Lose: nothing returned, chips stay at 900
		assert.Equal(t, 900, player.GetChips())
	})

	t.Run("double down", func(t *testing.T) {
		bj, _, _ := setupDeterministicBJ(
			900,
			[]*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 5, false),
				domain.NewCard(domain.CardDesignHeart, 6, false),
			},
			[]*domain.Card{
				domain.NewCard(domain.CardDesignClover, 10, false),
				domain.NewCard(domain.CardDesignDiamond, 7, false),
			},
			100,
		)
		err := bj.PlayerDoubleDown()
		assert.NoError(t, err)
		hand := bj.GetPlayerHands()[0]
		assert.Equal(t, 200, hand.GetBet())
		assert.True(t, hand.IsDoubled())
		assert.Equal(t, 3, hand.GetCardsSize())
		// Game finishes after double down (single hand auto-resolved)
		assert.True(t, bj.GetGameEndFlag())
	})

	t.Run("double down insufficient chips", func(t *testing.T) {
		bj, _, _ := setupDeterministicBJ(
			50,
			[]*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 5, false),
				domain.NewCard(domain.CardDesignHeart, 6, false),
			},
			[]*domain.Card{
				domain.NewCard(domain.CardDesignClover, 10, false),
				domain.NewCard(domain.CardDesignDiamond, 7, false),
			},
			100,
		)
		err := bj.PlayerDoubleDown()
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInsufficientChips)
	})

	t.Run("double down with 3 cards", func(t *testing.T) {
		bj, _, _ := setupDeterministicBJ(
			900,
			[]*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 2, false),
				domain.NewCard(domain.CardDesignHeart, 3, false),
				domain.NewCard(domain.CardDesignClover, 4, false),
			},
			[]*domain.Card{
				domain.NewCard(domain.CardDesignDiamond, 10, false),
				domain.NewCard(domain.CardDesignDiamond, 7, false),
			},
			100,
		)
		err := bj.PlayerDoubleDown()
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidPlay)
	})

	t.Run("double down on finished hand", func(t *testing.T) {
		bj, _, _ := setupDeterministicBJ(
			900,
			[]*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 5, false),
				domain.NewCard(domain.CardDesignHeart, 6, false),
			},
			[]*domain.Card{
				domain.NewCard(domain.CardDesignClover, 10, false),
				domain.NewCard(domain.CardDesignDiamond, 7, false),
			},
			100,
		)
		bj.GetPlayerHands()[0].SetStood(true)
		err := bj.PlayerDoubleDown()
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrHandFinished)
	})

	t.Run("split pair", func(t *testing.T) {
		bj, player, _ := setupDeterministicBJ(
			900,
			[]*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 8, false),
				domain.NewCard(domain.CardDesignHeart, 8, false),
			},
			[]*domain.Card{
				domain.NewCard(domain.CardDesignClover, 10, false),
				domain.NewCard(domain.CardDesignDiamond, 7, false),
			},
			100,
		)
		err := bj.PlayerSplit()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(bj.GetPlayerHands()))
		assert.Equal(t, 2, bj.GetPlayerHands()[0].GetCardsSize())
		assert.Equal(t, 2, bj.GetPlayerHands()[1].GetCardsSize())
		assert.Equal(t, 100, bj.GetPlayerHands()[0].GetBet())
		assert.Equal(t, 100, bj.GetPlayerHands()[1].GetBet())
		// chips: 900 - 100 (split) = 800
		assert.Equal(t, 800, player.GetChips())
	})

	t.Run("split aces auto-stand", func(t *testing.T) {
		bj, _, _ := setupDeterministicBJ(
			900,
			[]*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 1, false),
				domain.NewCard(domain.CardDesignHeart, 1, false),
			},
			[]*domain.Card{
				domain.NewCard(domain.CardDesignClover, 10, false),
				domain.NewCard(domain.CardDesignDiamond, 7, false),
			},
			100,
		)
		err := bj.PlayerSplit()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(bj.GetPlayerHands()))
		assert.True(t, bj.GetPlayerHands()[0].IsStood())
		assert.True(t, bj.GetPlayerHands()[1].IsStood())
	})

	t.Run("split insufficient chips", func(t *testing.T) {
		bj, _, _ := setupDeterministicBJ(
			50,
			[]*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 8, false),
				domain.NewCard(domain.CardDesignHeart, 8, false),
			},
			[]*domain.Card{
				domain.NewCard(domain.CardDesignClover, 10, false),
				domain.NewCard(domain.CardDesignDiamond, 7, false),
			},
			100,
		)
		err := bj.PlayerSplit()
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInsufficientChips)
	})

	t.Run("split non-pair", func(t *testing.T) {
		bj, _, _ := setupDeterministicBJ(
			900,
			[]*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 5, false),
				domain.NewCard(domain.CardDesignHeart, 8, false),
			},
			[]*domain.Card{
				domain.NewCard(domain.CardDesignClover, 10, false),
				domain.NewCard(domain.CardDesignDiamond, 7, false),
			},
			100,
		)
		err := bj.PlayerSplit()
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidPlay)
	})

	t.Run("all busted skips dealer draw", func(t *testing.T) {
		bj, _, dealer := setupDeterministicBJ(
			900,
			[]*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 10, false),
				domain.NewCard(domain.CardDesignHeart, 10, false),
				domain.NewCard(domain.CardDesignClover, 5, false),
			},
			[]*domain.Card{
				domain.NewCard(domain.CardDesignDiamond, 5, false),
				domain.NewCard(domain.CardDesignDiamond, 6, false),
			},
			100,
		)
		bj.GetPlayerHands()[0].SetBusted(true)
		_ = bj.PlayerStand()
		assert.True(t, bj.GetGameEndFlag())
		assert.Equal(t, 2, dealer.GetCardsSize())
	})

	t.Run("hit then stand flow", func(t *testing.T) {
		bj, _, _ := setupDeterministicBJ(
			900,
			[]*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 5, false),
				domain.NewCard(domain.CardDesignHeart, 6, false),
			},
			[]*domain.Card{
				domain.NewCard(domain.CardDesignClover, 10, false),
				domain.NewCard(domain.CardDesignDiamond, 7, false),
			},
			100,
		)
		_ = bj.PlayerHit()
		_ = bj.PlayerStand()
		assert.True(t, bj.GetGameEndFlag())
		assert.Equal(t, domain.BJPhaseEnd, bj.GetPhase())
	})

	t.Run("insurance win with dealer BJ", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		player := domain.NewBlackJackPlayer()
		dealer := domain.NewBlackJackPlayer()
		player.SetChips(850) // 1000 - 100 bet - 50 insurance
		dealer.SetChips(1000)
		bj := domain.NewBlackJack(tc, player, dealer)
		bj.Reset()
		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		hand.AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		player.AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 10, false))
		bj.SetPhase(domain.BJPhaseInsurance)
		assert.NoError(t, bj.PlayerInsurance()) // cost = 50, chips: 800
		assert.Equal(t, domain.BJPhaseEnd, bj.GetPhase())
		// Dealer has BJ, so insurance pays 3*50=150. Player loses hand (17 vs 21).
		// chips: 800 + 150 (insurance win) + 0 (hand loss) = 950
		assert.Equal(t, 950, player.GetChips())
	})

	t.Run("insurance lose without dealer BJ", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		player := domain.NewBlackJackPlayer()
		dealer := domain.NewBlackJackPlayer()
		player.SetChips(850)
		dealer.SetChips(1000)
		bj := domain.NewBlackJack(tc, player, dealer)
		bj.Reset()
		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		hand.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		player.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
		bj.SetPhase(domain.BJPhaseInsurance)
		assert.NoError(t, bj.PlayerInsurance()) // cost = 50, chips: 800
		assert.Equal(t, domain.BJPhaseAction, bj.GetPhase())
		assert.NoError(t, bj.PlayerStand())
		assert.True(t, bj.GetGameEndFlag())
		// Dealer: A + 7 = 18, Player: 20, player wins
		// Insurance lost (no dealer BJ), hand wins: 800 + 0 (insurance) + 200 (hand win) = 1000
		assert.Equal(t, 1000, player.GetChips())
	})

	t.Run("decline insurance", func(t *testing.T) {
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
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		player.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
		bj.SetPhase(domain.BJPhaseInsurance)
		assert.NoError(t, bj.PlayerDeclineInsurance())
		assert.Equal(t, domain.BJPhaseAction, bj.GetPhase())
		assert.Equal(t, 0, bj.GetInsuranceBet())
	})

	t.Run("dealer BJ vs player BJ is draw", func(t *testing.T) {
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
		hand.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		player.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 13, false))
		bj.SetPhase(domain.BJPhaseAction)
		_ = bj.PlayerStand()
		assert.True(t, bj.GetGameEndFlag())
		assert.Equal(t, domain.GameResultDraw, bj.GameJudgment())
		// Draw: bet returned, 900 + 100 = 1000
		assert.Equal(t, 1000, player.GetChips())
	})

	t.Run("SetPhase sets gameEndFlag correctly", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.SetPhase(domain.BJPhaseEnd)
		assert.True(t, bj.GetGameEndFlag())
		bj.SetPhase(domain.BJPhaseAction)
		assert.False(t, bj.GetGameEndFlag())
	})
}
