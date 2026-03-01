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

func TestBlackJack_ActionsOnFinishedHand(t *testing.T) {
	cases := []struct {
		name        string
		playerCards []*domain.Card
		action      func(bj *domain.BlackJack) error
	}{
		{
			name: "Stand on finished hand",
			playerCards: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 9, false),
				domain.NewCard(domain.CardDesignHeart, 8, false),
			},
			action: (*domain.BlackJack).PlayerStand,
		},
		{
			name: "Split on finished hand",
			playerCards: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 10, false),
				domain.NewCard(domain.CardDesignHeart, 10, false),
			},
			action: (*domain.BlackJack).PlayerSplit,
		},
	}

	dealerCards := []*domain.Card{
		domain.NewCard(domain.CardDesignClover, 10, false),
		domain.NewCard(domain.CardDesignClover, 7, false),
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bj, _, _ := setupDeterministicBJ(1000, tc.playerCards, dealerCards, 100)
			// Stand to finish the hand
			err := bj.PlayerStand()
			assert.NoError(t, err)
			// Reset phase to Action so the phase guard passes, to test the IsFinished guard
			bj.SetPhase(domain.BJPhaseAction)
			err = tc.action(bj)
			assert.Error(t, err)
			assert.ErrorIs(t, err, domain.ErrHandFinished)
		})
	}
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

		// デッキを全て引き切る（Reset後に新しいデッキが生成されるのでGetTrumpCardsを使う）
		deck := bj.GetTrumpCards()
		for i := 0; i < 52; i++ {
			deck.DrawCard()
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
		deck := bj.GetTrumpCards()
		for i := 0; i < 52; i++ {
			deck.DrawCard()
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
		deck := bj.GetTrumpCards()
		for i := 0; i < 52; i++ {
			deck.DrawCard()
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
		// This test verifies that when all player hands are busted,
		// the dealer skips drawing. Moved to blackjack_internal_test.go
		// (TestBlackJack_AllBustedSkipsDealerDraw) to allow deck stacking.
		t.Skip("moved to blackjack_internal_test.go for deck stacking")
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

func TestBlackJack_PlayerHit_NoBust(t *testing.T) {
	// Hit that doesn't bust: player score 11 (5+6), hit a 5 → score 16 < 22
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

	err := bj.PlayerHit()
	assert.NoError(t, err)
	assert.Equal(t, 3, hand.GetCardsSize())
	assert.False(t, hand.IsBusted(), "hand should not be busted")
	assert.False(t, hand.IsFinished(), "hand should not be finished")
}

func TestBlackJack_PlayerHit_SplitHandContinue(t *testing.T) {
	// After split, hit first hand to bust, then currentHandIdx should move to the second hand.
	tc := domain.NewTrumpCards(0)
	player := domain.NewBlackJackPlayer()
	dealer := domain.NewBlackJackPlayer()
	player.SetChips(1000)
	dealer.SetChips(1000)
	bj := domain.NewBlackJack(tc, player, dealer)
	bj.Reset()

	// Set up 2 hands manually (simulating a split)
	// hand0 has score 21 so any drawn card will bust it
	hand0 := domain.NewBlackJackHand()
	hand0.SetBet(100)
	hand0.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
	hand0.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
	hand0.AddCard(domain.NewCard(domain.CardDesignClover, 1, false)) // score = 21

	hand1 := domain.NewBlackJackHand()
	hand1.SetBet(100)
	hand1.AddCard(domain.NewCard(domain.CardDesignDiamond, 5, false))
	hand1.AddCard(domain.NewCard(domain.CardDesignDiamond, 6, false))

	bj.SetPlayerHands([]*domain.BlackJackHand{hand0, hand1})
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
	bj.SetPhase(domain.BJPhaseAction)

	// Hit on hand0 (score 21): any card will bust it (22+)
	err := bj.PlayerHit()
	assert.NoError(t, err)
	assert.True(t, hand0.IsBusted(), "first hand should be busted")
	// advanceHand should move to hand1 which is not finished
	assert.Equal(t, 1, bj.GetCurrentHandIdx(), "should advance to second hand")
	assert.False(t, hand1.IsFinished(), "second hand should not be finished")
}

func TestBlackJack_AdvanceHand_AllHandsFinished(t *testing.T) {
	// All hands finished triggers dealerPlay. Set up 2 stood hands, stand on the last one.
	tc := domain.NewTrumpCards(0)
	player := domain.NewBlackJackPlayer()
	dealer := domain.NewBlackJackPlayer()
	player.SetChips(800)
	dealer.SetChips(1000)
	bj := domain.NewBlackJack(tc, player, dealer)
	bj.Reset()

	hand0 := domain.NewBlackJackHand()
	hand0.SetBet(100)
	hand0.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
	hand0.AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
	hand0.SetStood(true) // already stood

	hand1 := domain.NewBlackJackHand()
	hand1.SetBet(100)
	hand1.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
	hand1.AddCard(domain.NewCard(domain.CardDesignDiamond, 8, false))

	bj.SetPlayerHands([]*domain.BlackJackHand{hand0, hand1})
	dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 9, false))
	bj.SetPhase(domain.BJPhaseAction)

	// currentHandIdx should point to hand1 (the first unfinished one)
	// Since hand0 is stood, we need to set the current hand to 1
	// Actually, we just call stand which sets the current hand's stood and advances.
	// The current index is 0 by default, but hand0 is already stood, so we need to
	// simulate the state where we are on hand1.
	// Let's just stand on hand1 by ensuring currentHandIdx points to it.
	// We can call PlayerStand which will stand on hand at currentHandIdx=0,
	// but hand0 is already stood. Let's use a clean approach:
	// Set hand0 as NOT stood, then stand twice.

	hand0.SetStood(false) // reset for clean test
	bj.SetPhase(domain.BJPhaseAction)

	// Stand on hand0
	err := bj.PlayerStand()
	assert.NoError(t, err)
	assert.True(t, hand0.IsStood())
	// advanceHand should move to hand1
	assert.Equal(t, 1, bj.GetCurrentHandIdx())
	assert.False(t, bj.GetGameEndFlag())

	// Stand on hand1 → all finished → dealerPlay → endGame
	err = bj.PlayerStand()
	assert.NoError(t, err)
	assert.True(t, hand1.IsStood())
	assert.True(t, bj.GetGameEndFlag(), "game should end after all hands finished")
	assert.Equal(t, domain.BJPhaseEnd, bj.GetPhase())
}

func TestBlackJack_DoubleDown_NoBust(t *testing.T) {
	// DD where score is 17-21 after drawn card → SetStood(true) branch
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
	// Score = 11, any card drawn will give 12-21 (won't bust since max card=10 → 11+10=21)
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
	bj.SetPhase(domain.BJPhaseAction)

	err := bj.PlayerDoubleDown()
	assert.NoError(t, err)
	assert.Equal(t, 200, hand.GetBet())
	assert.True(t, hand.IsDoubled())
	assert.Equal(t, 3, hand.GetCardsSize())
	// Score is 11 + (drawn card value) ≤ 21, so should be stood not busted
	assert.True(t, hand.IsStood(), "hand should be stood after DD without bust")
	assert.False(t, hand.IsBusted(), "hand should not be busted")
	// Single hand → game should end after DD
	assert.True(t, bj.GetGameEndFlag())
}

func TestBlackJack_PlayerSplit_DeckExhausted(t *testing.T) {
	// Deck exhausted before first draw → error returned and state reverted.
	tc := domain.NewTrumpCards(0)
	player := domain.NewBlackJackPlayer()
	dealer := domain.NewBlackJackPlayer()
	player.SetChips(1000)
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

	chipsBefore := player.GetChips()

	// Drain the deck completely (Reset rebuilds deck so use GetTrumpCards)
	deck := bj.GetTrumpCards()
	for i := 0; i < 52; i++ {
		deck.DrawCard()
	}

	err := bj.PlayerSplit()
	assert.ErrorIs(t, err, domain.ErrDeckExhausted)
	// State should be reverted: 1 hand with 2 original cards, chips refunded
	assert.Equal(t, 1, len(bj.GetPlayerHands()))
	assert.Equal(t, 2, bj.GetPlayerHands()[0].GetCardsSize())
	assert.Equal(t, 100, bj.GetPlayerHands()[0].GetBet())
	assert.Equal(t, chipsBefore, player.GetChips(), "split bet should be refunded")
	assert.False(t, bj.GetGameEndFlag())
}

func TestBlackJack_PlayerSplit_DeckExhaustedAfterFirstDraw(t *testing.T) {
	// First draw succeeds but second draw fails → error returned and state fully reverted.
	tc := domain.NewTrumpCards(0)
	player := domain.NewBlackJackPlayer()
	dealer := domain.NewBlackJackPlayer()
	player.SetChips(1000)
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

	chipsBefore := player.GetChips()

	// Drain the deck to leave exactly 1 card
	deck := bj.GetTrumpCards()
	for i := 0; i < 51; i++ {
		deck.DrawCard()
	}

	err := bj.PlayerSplit()
	assert.ErrorIs(t, err, domain.ErrDeckExhausted)
	// State should be reverted: 1 hand with 2 original cards, chips refunded
	assert.Equal(t, 1, len(bj.GetPlayerHands()))
	assert.Equal(t, 2, bj.GetPlayerHands()[0].GetCardsSize())
	assert.Equal(t, 100, bj.GetPlayerHands()[0].GetBet())
	assert.Equal(t, chipsBefore, player.GetChips(), "split bet should be refunded")
	assert.False(t, bj.GetGameEndFlag())
}

func TestBlackJack_PlayerDoubleDown_DeckExhausted(t *testing.T) {
	// Deck exhausted during double down → error returned and state reverted.
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

	chipsBefore := player.GetChips()

	// Drain the deck completely
	deck := bj.GetTrumpCards()
	for i := 0; i < 52; i++ {
		deck.DrawCard()
	}

	err := bj.PlayerDoubleDown()
	assert.ErrorIs(t, err, domain.ErrDeckExhausted)
	// State should be reverted: bet not doubled, chips not reduced, hand unchanged
	assert.Equal(t, chipsBefore, player.GetChips(), "extra bet should be refunded")
	assert.Equal(t, 100, hand.GetBet(), "bet should not be doubled")
	assert.Equal(t, 2, hand.GetCardsSize(), "no card should be added")
	assert.False(t, hand.IsDoubled(), "doubled flag should be reverted")
	assert.False(t, hand.IsStood(), "hand should not be stood")
	assert.False(t, hand.IsBusted(), "hand should not be busted")
	assert.False(t, bj.GetGameEndFlag())
}

func TestBlackJack_PlayerInsurance_InsufficientChips(t *testing.T) {
	// Insurance phase but player chips < bet/2.
	tc := domain.NewTrumpCards(0)
	player := domain.NewBlackJackPlayer()
	dealer := domain.NewBlackJackPlayer()
	player.SetChips(30) // less than bet/2 = 50
	dealer.SetChips(1000)
	bj := domain.NewBlackJack(tc, player, dealer)
	bj.Reset()

	hand := bj.GetPlayerHands()[0]
	hand.SetBet(100)
	hand.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
	hand.AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
	player.AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 10, false))
	bj.SetPhase(domain.BJPhaseInsurance)

	err := bj.PlayerInsurance()
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInsufficientChips)
	// Phase should still be insurance
	assert.Equal(t, domain.BJPhaseInsurance, bj.GetPhase())
	// Chips should not change
	assert.Equal(t, 30, player.GetChips())
}

func TestBlackJack_JudgeHand_SplitBJNo3to2(t *testing.T) {
	// When there are multiple hands (split), a natural BJ doesn't get 3:2 payout.
	tc := domain.NewTrumpCards(0)
	player := domain.NewBlackJackPlayer()
	dealer := domain.NewBlackJackPlayer()
	player.SetChips(800) // already subtracted 200 (100 per hand)
	dealer.SetChips(1000)
	bj := domain.NewBlackJack(tc, player, dealer)
	bj.Reset()

	// Hand 0: natural BJ (Ace + 10) in 2 cards — not yet stood
	hand0 := domain.NewBlackJackHand()
	hand0.SetBet(100)
	hand0.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	hand0.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))

	// Hand 1: normal hand (10 + 9 = 19) — not yet stood
	hand1 := domain.NewBlackJackHand()
	hand1.SetBet(100)
	hand1.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
	hand1.AddCard(domain.NewCard(domain.CardDesignDiamond, 9, false))

	bj.SetPlayerHands([]*domain.BlackJackHand{hand0, hand1})
	dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 9, false))
	bj.SetPhase(domain.BJPhaseAction)

	// Stand hand0 → advanceHand finds hand1 not finished, moves there
	err := bj.PlayerStand()
	assert.NoError(t, err)
	assert.Equal(t, 1, bj.GetCurrentHandIdx())

	// Stand hand1 → all hands finished → dealerPlay → endGame
	err = bj.PlayerStand()
	assert.NoError(t, err)
	assert.True(t, bj.GetGameEndFlag())

	// Dealer has 18. Hand0 has 21 (BJ), Hand1 has 19.
	// Both win. But hand0 should NOT get 3:2 because len(playerHands) == 2.
	// Hand0 win: normal payout = 100*2 = 200
	// Hand1 win: normal payout = 100*2 = 200
	// Total added: 400. Chips: 800 + 400 = 1200
	assert.Equal(t, 1200, player.GetChips())

	// If hand0 had gotten 3:2, it would be 100+150=250 instead of 200, total 1250.
	// So checking 1200 confirms no 3:2 payout for split BJ.
}

func TestBlackJack_PlayerBet_DealFailed(t *testing.T) {
	// Bet with partially exhausted deck - deal starts but some cards are nil.
	tc := domain.NewTrumpCards(0)
	player := domain.NewBlackJackPlayer()
	dealer := domain.NewBlackJackPlayer()
	player.SetChips(1000)
	dealer.SetChips(1000)
	bj := domain.NewBlackJack(tc, player, dealer)
	bj.Reset()

	// Drain deck to leave only 2 cards (need 4 for deal: 2 player + 2 dealer)
	// Reset rebuilds the deck, so use GetTrumpCards to get the current deck
	deck := bj.GetTrumpCards()
	for i := 0; i < 50; i++ {
		deck.DrawCard()
	}

	err := bj.PlayerBet(100)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrDeckExhausted)
	// Chips should be refunded
	assert.Equal(t, 1000, player.GetChips())
	assert.Equal(t, domain.BJPhaseBet, bj.GetPhase())
}

func TestBlackJack_PlayerHit_FinishedHand(t *testing.T) {
	// Hit on already-finished hand returns ErrHandFinished.
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
	hand.AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
	hand.SetStood(true) // mark as finished
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
	bj.SetPhase(domain.BJPhaseAction)

	err := bj.PlayerHit()
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrHandFinished)
	// Card count should not change
	assert.Equal(t, 2, hand.GetCardsSize())
}

func TestBlackJack_JudgeHand_DealerBJvsPlayerNonBJ(t *testing.T) {
	// Dealer has natural BJ (2 cards=21), player has 21 with 3+ cards → player loses.
	tc := domain.NewTrumpCards(0)
	player := domain.NewBlackJackPlayer()
	dealer := domain.NewBlackJackPlayer()
	player.SetChips(900)
	dealer.SetChips(1000)
	bj := domain.NewBlackJack(tc, player, dealer)
	bj.Reset()

	hand := bj.GetPlayerHands()[0]
	hand.SetBet(100)
	// Player: 7 + 7 + 7 = 21 (3 cards, not natural BJ)
	hand.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
	hand.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
	hand.AddCard(domain.NewCard(domain.CardDesignClover, 7, false))
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
	player.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
	player.AddCard(domain.NewCard(domain.CardDesignClover, 7, false))

	// Dealer: Ace + King = 21 (natural BJ, 2 cards)
	dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 1, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 13, false))
	bj.SetPhase(domain.BJPhaseAction)

	err := bj.PlayerStand()
	assert.NoError(t, err)
	assert.True(t, bj.GetGameEndFlag())

	// Both have score 21, but dealer has natural BJ and player does not → player loses
	assert.Equal(t, domain.GameResultLose, bj.GameJudgment())
	// Lose: nothing returned, chips stay at 900
	assert.Equal(t, 900, player.GetChips())
}

// TestBlackJack_PlayerBet_DealerAceTriggersInsurance moved to
// blackjack_internal_test.go as a deterministic test that stacks the deck
// instead of retrying random shuffles.

// TestBlackJack_PlayerDoubleDown_Bust moved to blackjack_internal_test.go
// as a deterministic test that stacks the deck instead of retrying random shuffles.

func TestBlackJack_PlayerSurrender(t *testing.T) {
	t.Run("surrender success returns half bet", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		hand.AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
		bj.GetDealer().AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
		bj.GetDealer().AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
		bj.SetPhase(domain.BJPhaseAction)
		startChips := bj.GetPlayer().GetChips()

		err := bj.PlayerSurrender()
		assert.NoError(t, err)
		// half bet returned
		assert.Equal(t, startChips+50, bj.GetPlayer().GetChips())
		assert.True(t, hand.IsSurrendered())
		// game ends (only one hand, all done)
		assert.True(t, bj.GetGameEndFlag())
	})
	t.Run("surrender wrong phase", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		err := bj.PlayerSurrender()
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrWrongPhase)
	})
	t.Run("surrender with 3 cards rejected", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		hand.AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
		hand.AddCard(domain.NewCard(domain.CardDesignClover, 7, false))
		bj.GetDealer().AddCard(domain.NewCard(domain.CardDesignDiamond, 10, false))
		bj.SetPhase(domain.BJPhaseAction)

		err := bj.PlayerSurrender()
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidPlay)
	})
	t.Run("surrender after stood rejected", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		hand.AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
		hand.SetStood(true)
		bj.GetDealer().AddCard(domain.NewCard(domain.CardDesignDiamond, 10, false))
		bj.SetPhase(domain.BJPhaseAction)

		err := bj.PlayerSurrender()
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidPlay)
	})
	t.Run("surrender with all-surrendered hands triggers game end without dealer play", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		hand.AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
		bj.GetDealer().AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
		bj.GetDealer().AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
		bj.SetPhase(domain.BJPhaseAction)

		err := bj.PlayerSurrender()
		assert.NoError(t, err)
		assert.True(t, bj.GetGameEndFlag())
		// dealer should have only 2 cards (not drawn more)
		assert.Equal(t, 2, bj.GetDealer().GetCardsSize())
	})
}

func TestBlackJack_SetDeckCount(t *testing.T) {
	t.Run("set valid deck counts", func(t *testing.T) {
		for _, count := range []int{1, 2, 4, 6, 8} {
			bj := domain.NewDefaultBlackJack()
			err := bj.SetDeckCount(count)
			assert.NoError(t, err, "count=%d", count)
			assert.Equal(t, count, bj.GetDeckCount())
		}
	})
	t.Run("invalid deck count rejected", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		err := bj.SetDeckCount(3)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidAmount)
	})
	t.Run("wrong phase rejected", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.SetPhase(domain.BJPhaseAction)
		err := bj.SetDeckCount(2)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrWrongPhase)
	})
	t.Run("deck count persists across reset", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		err := bj.SetDeckCount(6)
		assert.NoError(t, err)
		bj.Reset()
		assert.Equal(t, 6, bj.GetDeckCount())
	})
}

func TestBlackJack_GetDeckCount_Default(t *testing.T) {
	bj := domain.NewDefaultBlackJack()
	assert.Equal(t, 1, bj.GetDeckCount())
}

func TestBlackJack_ToggleHint(t *testing.T) {
	bj := domain.NewDefaultBlackJack()
	assert.False(t, bj.IsHintEnabled())
	bj.ToggleHint()
	assert.True(t, bj.IsHintEnabled())
	bj.ToggleHint()
	assert.False(t, bj.IsHintEnabled())
}

func TestBlackJack_GetBasicStrategySuggestion(t *testing.T) {
	t.Run("hint disabled returns none", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		assert.Equal(t, domain.BJSuggestNone, bj.GetBasicStrategySuggestion())
	})
	t.Run("bet phase returns none even with hint on", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.ToggleHint()
		assert.Equal(t, domain.BJSuggestNone, bj.GetBasicStrategySuggestion())
	})
	t.Run("insurance phase returns decline insurance", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.ToggleHint()
		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		hand.AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
		bj.GetDealer().AddCard(domain.NewCard(domain.CardDesignClover, 1, false)) // ace upcard
		bj.SetPhase(domain.BJPhaseInsurance)
		assert.Equal(t, domain.BJSuggestDeclineInsurance, bj.GetBasicStrategySuggestion())
	})
	t.Run("action phase with hard 16 vs 10 returns surrender", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.ToggleHint()
		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		hand.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		bj.GetDealer().AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
		bj.GetDealer().AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
		bj.SetPhase(domain.BJPhaseAction)
		assert.Equal(t, domain.BJSuggestSurrender, bj.GetBasicStrategySuggestion())
	})
	t.Run("action phase with finished hand returns none", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.ToggleHint()
		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		hand.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		hand.SetStood(true)
		bj.GetDealer().AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
		bj.SetPhase(domain.BJPhaseAction)
		assert.Equal(t, domain.BJSuggestNone, bj.GetBasicStrategySuggestion())
	})
	t.Run("action phase with no dealer upcard returns none", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.ToggleHint()
		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		hand.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		// no dealer card
		bj.SetPhase(domain.BJPhaseAction)
		assert.Equal(t, domain.BJSuggestNone, bj.GetBasicStrategySuggestion())
	})
}

func TestBlackJack_Reset_PreservesHintAndDeckCount(t *testing.T) {
	bj := domain.NewDefaultBlackJack()
	bj.ToggleHint()
	err := bj.SetDeckCount(4)
	assert.NoError(t, err)
	bj.Reset()
	assert.True(t, bj.IsHintEnabled())
	assert.Equal(t, 4, bj.GetDeckCount())
}

func TestBlackJack_Surrender_ResolvePayout_Skipped(t *testing.T) {
	// Surrendered hand should be skipped in payout; other hand is settled normally
	bj := domain.NewDefaultBlackJack()
	bj.Reset()
	// Give dealer a deterministic hand (17) so DealerHit stops immediately
	bj.GetDealer().AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
	bj.GetDealer().AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
	// Set up two hands: hand[0] surrendered, hand[1] wins
	hand0 := bj.GetPlayerHands()[0]
	hand0.SetBet(100)
	hand0.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
	hand0.AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
	hand1 := domain.NewBlackJackHand()
	hand1.SetBet(100)
	hand1.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
	hand1.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
	bj.SetPlayerHands([]*domain.BlackJackHand{hand0, hand1})

	// Start chips: player already paid 200 in bet (100+100)
	startChips := bj.GetPlayer().GetChips()

	// Surrender hand0 returns 50
	bj.SetPhase(domain.BJPhaseAction)
	err := bj.PlayerSurrender()
	assert.NoError(t, err)
	// Now currentHandIdx moves to 1 (hand1 not finished)
	// stand hand1 to trigger dealer play
	_ = bj.PlayerStand()

	// Dealer has score < 20 so player wins hand1 (dealer draws until >=17)
	// Just verify game ended and player got back hand1 winnings but not double hand0
	assert.True(t, bj.GetGameEndFlag())
	// Player should have gotten: 50 (from surrender) + some winnings from hand1
	assert.Greater(t, bj.GetPlayer().GetChips(), startChips+50)
}

func TestBlackJack_GetDeckCount_ZeroDefault(t *testing.T) {
	// NewBlackJack directly doesn't set deckCount, so GetDeckCount should return default
	tc := domain.NewTrumpCards(0)
	player := domain.NewBlackJackPlayer()
	dealer := domain.NewBlackJackPlayer()
	bj := domain.NewBlackJack(tc, player, dealer)
	assert.Equal(t, domain.BJDefaultDecks, bj.GetDeckCount())
}
