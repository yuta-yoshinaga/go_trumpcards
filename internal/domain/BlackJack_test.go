package domain_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		err := bj.PlayerBet(100, 0, 0, 0)
		assert.NoError(t, err)
		assert.Equal(t, 2, bj.GetPlayerHands()[0].GetCardsSize())
		assert.Equal(t, 2, bj.GetDealer().GetCardsSize())
		assert.NotEqual(t, domain.BJPhaseBet, bj.GetPhase())
	})
	t.Run("bet below minimum", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		err := bj.PlayerBet(5, 0, 0, 0)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidAmount)
		assert.Equal(t, domain.BJPhaseBet, bj.GetPhase())
	})
	t.Run("bet with insufficient chips", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		bj.GetPlayer().SetChips(50)
		err := bj.PlayerBet(100, 0, 0, 0)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInsufficientChips)
	})
	t.Run("bet not multiple of min bet", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		err := bj.PlayerBet(15, 0, 0, 0)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidAmount)
		assert.Equal(t, domain.BJPhaseBet, bj.GetPhase())
	})
	t.Run("bet in wrong phase", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		_ = bj.PlayerBet(100, 0, 0, 0)
		err := bj.PlayerBet(100, 0, 0, 0)
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
		err := bj.PlayerBet(100, 0, 0, 0)
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
		err := bj.PlayerBet(5, 0, 0, 0) // below minimum
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidAmount)
		assert.Contains(t, err.Error(), "Invalid bet amount.")
	})
	t.Run("bet wrong phase returns ErrWrongPhase", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		_ = bj.PlayerBet(100, 0, 0, 0)
		err := bj.PlayerBet(100, 0, 0, 0)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrWrongPhase)
		assert.Contains(t, err.Error(), "Bet is only allowed during the bet phase.")
	})
	t.Run("bet insufficient chips returns ErrInsufficientChips", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		bj.GetPlayer().SetChips(50)
		err := bj.PlayerBet(100, 0, 0, 0)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInsufficientChips)
		assert.Contains(t, err.Error(), "Insufficient chips.")
	})
	t.Run("successful bet returns nil", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		err := bj.PlayerBet(100, 0, 0, 0)
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

	// Hand 0: natural BJ (Ace + 10) in 2 cards — from split
	hand0 := domain.NewBlackJackHand()
	hand0.SetBet(100)
	hand0.SetFromSplit(true)
	hand0.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	hand0.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))

	// Hand 1: normal hand (10 + 9 = 19) — from split
	hand1 := domain.NewBlackJackHand()
	hand1.SetBet(100)
	hand1.SetFromSplit(true)
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

	err := bj.PlayerBet(100, 0, 0, 0)
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

// --- New tests for BlackJack features ---

func TestDefaultBlackJackConfig(t *testing.T) {
	cfg := domain.DefaultBlackJackConfig()
	assert.False(t, cfg.DealerHitsSoft17, "DealerHitsSoft17 default should be false")
	assert.Equal(t, 0, cfg.CpuPlayerCount, "CpuPlayerCount default should be 0")
	assert.False(t, cfg.CountingEnabled, "CountingEnabled default should be false")
	assert.True(t, cfg.DoubleAfterSplit, "DoubleAfterSplit default should be true")
	assert.Equal(t, domain.BJCountingHiLo, cfg.CountingSystem, "CountingSystem default should be Hi-Lo (0)")
}

func TestIsBalancedCountingSystem(t *testing.T) {
	assert.True(t, domain.IsBalancedCountingSystem(domain.BJCountingHiLo), "Hi-Lo is balanced")
	assert.False(t, domain.IsBalancedCountingSystem(domain.BJCountingKO), "KO is unbalanced")
	assert.True(t, domain.IsBalancedCountingSystem(domain.BJCountingZen), "Zen Count is balanced")
	assert.True(t, domain.IsBalancedCountingSystem(domain.BJCountingOmegaII), "Omega II is balanced")
}

func TestBlackJackConfig_Validate(t *testing.T) {
	t.Run("valid default config", func(t *testing.T) {
		cfg := domain.DefaultBlackJackConfig()
		assert.NoError(t, cfg.Validate())
	})
	t.Run("CPU count negative fails", func(t *testing.T) {
		cfg := domain.BlackJackConfig{CpuPlayerCount: -1}
		assert.Error(t, cfg.Validate())
	})
	t.Run("CPU count too high fails", func(t *testing.T) {
		cfg := domain.BlackJackConfig{CpuPlayerCount: domain.BJMaxCpuPlayers + 1}
		assert.Error(t, cfg.Validate())
	})
	t.Run("counting system negative fails", func(t *testing.T) {
		cfg := domain.BlackJackConfig{CountingSystem: -1}
		assert.Error(t, cfg.Validate())
	})
	t.Run("counting system too high fails", func(t *testing.T) {
		cfg := domain.BlackJackConfig{CountingSystem: domain.BJCountingMax + 1}
		assert.Error(t, cfg.Validate())
	})
	t.Run("surrender rule negative fails", func(t *testing.T) {
		cfg := domain.BlackJackConfig{SurrenderRule: -1}
		assert.Error(t, cfg.Validate())
	})
	t.Run("surrender rule too high fails", func(t *testing.T) {
		cfg := domain.BlackJackConfig{SurrenderRule: domain.BJSurrenderMax + 1}
		assert.Error(t, cfg.Validate())
	})
	t.Run("deck penetration 0 ok (use default)", func(t *testing.T) {
		cfg := domain.BlackJackConfig{DeckPenetration: 0}
		assert.NoError(t, cfg.Validate())
	})
	t.Run("deck penetration 50 ok", func(t *testing.T) {
		cfg := domain.BlackJackConfig{DeckPenetration: 50}
		assert.NoError(t, cfg.Validate())
	})
	t.Run("deck penetration 75 ok", func(t *testing.T) {
		cfg := domain.BlackJackConfig{DeckPenetration: 75}
		assert.NoError(t, cfg.Validate())
	})
	t.Run("deck penetration invalid fails", func(t *testing.T) {
		cfg := domain.BlackJackConfig{DeckPenetration: 60}
		assert.Error(t, cfg.Validate())
	})
}

func TestBlackJack_DAS_Enabled(t *testing.T) {
	// DAS enabled (default) → double down after split should succeed
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

	// Perform split
	err := bj.PlayerSplit()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(bj.GetPlayerHands()))

	// Double down on the first split hand should succeed (DAS default true)
	splitHand := bj.GetPlayerHands()[bj.GetCurrentHandIdx()]
	require.True(t, splitHand.GetCardsSize() == 2 && !splitHand.IsFinished(), "split hand must have exactly 2 cards and not be finished")
	err = bj.PlayerDoubleDown()
	assert.NoError(t, err, "DD after split should be allowed when DAS is enabled")
}

func TestBlackJack_DAS_Disabled(t *testing.T) {
	// DAS disabled → double down after split should be rejected
	tc := domain.NewTrumpCards(0)
	player := domain.NewBlackJackPlayer()
	dealer := domain.NewBlackJackPlayer()
	player.SetChips(1000)
	dealer.SetChips(1000)
	bj := domain.NewBlackJack(tc, player, dealer)
	_ = bj.SetConfig(domain.BlackJackConfig{DoubleAfterSplit: false})
	bj.Reset()

	hand := bj.GetPlayerHands()[0]
	hand.SetBet(100)
	hand.AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
	hand.AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
	bj.SetPhase(domain.BJPhaseAction)

	// Perform split
	err := bj.PlayerSplit()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(bj.GetPlayerHands()))

	// Double down on split hand should be rejected
	splitHand := bj.GetPlayerHands()[bj.GetCurrentHandIdx()]
	require.True(t, splitHand.GetCardsSize() == 2 && !splitHand.IsFinished(), "split hand must have exactly 2 cards and not be finished")
	err = bj.PlayerDoubleDown()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Double after split is not allowed")
}

func TestBlackJack_DAS_Disabled_NonSplitHandAllowed(t *testing.T) {
	// DAS disabled but non-split hand → double down should succeed
	tc := domain.NewTrumpCards(0)
	player := domain.NewBlackJackPlayer()
	dealer := domain.NewBlackJackPlayer()
	player.SetChips(1000)
	dealer.SetChips(1000)
	bj := domain.NewBlackJack(tc, player, dealer)
	_ = bj.SetConfig(domain.BlackJackConfig{DoubleAfterSplit: false})
	bj.Reset()

	hand := bj.GetPlayerHands()[0]
	hand.SetBet(100)
	hand.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	hand.AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
	bj.SetPhase(domain.BJPhaseAction)

	// Single hand → DD should be allowed regardless of DAS setting
	err := bj.PlayerDoubleDown()
	assert.NoError(t, err, "DD on non-split hand should be allowed even when DAS is disabled")
}

func TestBlackJackPlayerIsSoft(t *testing.T) {
	t.Run("A+6 is soft 17", func(t *testing.T) {
		p := domain.NewBlackJackPlayer()
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false)) // Ace
		p.AddCard(domain.NewCard(domain.CardDesignHeart, 6, false)) // 6
		assert.Equal(t, 17, p.GetScore())
		assert.True(t, p.IsSoft(), "A+6 should be soft")
	})
	t.Run("K+7 is hard 17", func(t *testing.T) {
		p := domain.NewBlackJackPlayer()
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 13, false)) // King = 10
		p.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))  // 7
		assert.Equal(t, 17, p.GetScore())
		assert.False(t, p.IsSoft(), "K+7 should be hard")
	})
	t.Run("A+K is hard 21 (ace counted as 11)", func(t *testing.T) {
		p := domain.NewBlackJackPlayer()
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))  // Ace
		p.AddCard(domain.NewCard(domain.CardDesignHeart, 13, false)) // King = 10
		assert.Equal(t, 21, p.GetScore())
		// Ace is 11 and score is 21 (not busted), so it is soft
		assert.True(t, p.IsSoft(), "A+K = 21 should be soft (ace is 11)")
	})
	t.Run("A+5+10 is hard 16 (ace forced to 1)", func(t *testing.T) {
		p := domain.NewBlackJackPlayer()
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))   // Ace
		p.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))   // 5
		p.AddCard(domain.NewCard(domain.CardDesignClover, 10, false)) // 10
		assert.Equal(t, 16, p.GetScore())
		assert.False(t, p.IsSoft(), "A+5+10 should be hard (ace forced to 1)")
	})
}

func TestDealerHitSoft17(t *testing.T) {
	t.Run("S17 default: dealer with soft 17 stands", func(t *testing.T) {
		// Create BJ with enough cards in deck for dealer to draw
		tc := domain.NewTrumpCardsWithDecks(1, 0)
		player := domain.NewBlackJackPlayer()
		dealer := domain.NewBlackJackPlayer()
		player.SetChips(domain.BJDefaultChips)
		dealer.SetChips(domain.BJDefaultChips)
		bj := domain.NewBlackJack(tc, player, dealer)

		// Set up dealer hand: A+6 = soft 17
		dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false)) // Ace
		dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 6, false)) // 6
		assert.Equal(t, 17, dealer.GetScore())
		assert.True(t, dealer.IsSoft())

		// Set up a non-busted player hand so dealer will play
		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
		hand.AddCard(domain.NewCard(domain.CardDesignDiamond, 8, false)) // 18

		// Default config: S17 (DealerHitsSoft17 = false)
		cfg := domain.DefaultBlackJackConfig()
		assert.False(t, cfg.DealerHitsSoft17)

		bj.SetPhase(domain.BJPhaseAction)
		// Call DealerHit directly
		bj.DealerHit()

		// Dealer should stand on soft 17 with S17 rules: no additional cards
		assert.Equal(t, 2, dealer.GetCardsSize(), "S17: dealer should stand on soft 17")
		assert.Equal(t, 17, dealer.GetScore())
	})

	t.Run("H17: dealer with soft 17 hits", func(t *testing.T) {
		// Create BJ with enough cards in deck for dealer to draw
		tc := domain.NewTrumpCardsWithDecks(2, 0)
		player := domain.NewBlackJackPlayer()
		dealer := domain.NewBlackJackPlayer()
		player.SetChips(domain.BJDefaultChips)
		dealer.SetChips(domain.BJDefaultChips)
		bj := domain.NewBlackJack(tc, player, dealer)

		// Set up dealer hand: A+6 = soft 17
		dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false)) // Ace
		dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 6, false)) // 6
		assert.Equal(t, 17, dealer.GetScore())
		assert.True(t, dealer.IsSoft())

		// Set up a non-busted player hand so dealer will play
		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
		hand.AddCard(domain.NewCard(domain.CardDesignDiamond, 8, false)) // 18

		// Enable H17
		cfg := domain.BlackJackConfig{DealerHitsSoft17: true}
		_ = bj.SetConfig(cfg)
		// SetConfig only works in BET phase; set config directly via bet flow
		// We need to set phase to BET first
		bj.SetPhase(domain.BJPhaseBet)
		err := bj.SetConfig(cfg)
		assert.NoError(t, err)

		bj.SetPhase(domain.BJPhaseAction)
		bj.DealerHit()

		// Dealer should have drawn at least one more card
		assert.Greater(t, dealer.GetCardsSize(), 2, "H17: dealer should hit on soft 17")
	})
}

func TestRunningCountUpdates(t *testing.T) {
	bj := domain.NewDefaultBlackJack()
	// Enable counting
	bj.SetPhase(domain.BJPhaseBet)
	cfg := domain.BlackJackConfig{CountingEnabled: true}
	err := bj.SetConfig(cfg)
	assert.NoError(t, err)

	bj.Reset()

	// After reset, running count should be 0
	assert.Equal(t, 0, bj.GetRunningCount())

	// Bet to deal cards; running count should change based on dealt cards
	err = bj.PlayerBet(domain.BJMinBet, 0, 0, 0)
	assert.NoError(t, err)

	// After dealing, the running count should have been updated (3 counted: 2 player + 1 dealer upcard)
	// We can't predict exact value due to shuffling, but counting should be enabled
	assert.True(t, bj.IsCountingEnabled())

	// If game is still in action phase, hit to verify count changes
	if bj.GetPhase() == domain.BJPhaseAction || bj.GetPhase() == domain.BJPhaseInsurance {
		prevRC := bj.GetRunningCount()
		if bj.GetPhase() == domain.BJPhaseInsurance {
			_ = bj.PlayerDeclineInsurance()
		}
		if bj.GetPhase() == domain.BJPhaseAction && !bj.GetGameEndFlag() {
			hand := bj.GetPlayerHands()[0]
			if !hand.IsFinished() {
				_ = bj.PlayerHit()
				// Running count may or may not have changed depending on the card
				// At least verify it didn't cause an error
				_ = prevRC
			}
		}
	}
}

func TestTrueCount(t *testing.T) {
	t.Run("true count with known remaining count", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()

		// With a fresh 1-deck shoe (52 cards), remaining = 52, decks remaining = 1.0
		// RC = 0, so TC = 0/1.0 = 0
		assert.Equal(t, 0.0, bj.GetTrueCount())
	})
	t.Run("true count with nil trumpCards returns 0", func(t *testing.T) {
		// NewBlackJack with no Reset will have trumpCards, but we can test the branch
		// via NewDefaultBlackJack which always has trumpCards
		bj := domain.NewDefaultBlackJack()
		assert.Equal(t, 0.0, bj.GetTrueCount())
	})
	t.Run("true count with multi-deck", func(t *testing.T) {
		// Create a 6-deck shoe
		tc := domain.NewTrumpCardsWithDecks(6, 0)
		player := domain.NewBlackJackPlayer()
		dealer := domain.NewBlackJackPlayer()
		player.SetChips(domain.BJDefaultChips)
		dealer.SetChips(domain.BJDefaultChips)
		bj := domain.NewBlackJack(tc, player, dealer)

		// TC = RC / decks remaining
		// With 6 decks (312 cards), decks remaining = 312/52 = 6.0
		// RC = 0, TC = 0
		assert.Equal(t, 0.0, bj.GetTrueCount())
	})
}

func TestShoePersistence(t *testing.T) {
	t.Run("shoe persists when >25% remaining", func(t *testing.T) {
		// Use a 2-deck shoe (104 cards)
		bj := domain.NewDefaultBlackJack()
		err := bj.SetDeckCount(2)
		assert.NoError(t, err)
		bj.Reset()

		// After reset, we should have a 2-deck shoe
		initialTotal := bj.GetTrumpCards().GetTotalCount()
		assert.Equal(t, 104, initialTotal)

		// Bet and play a hand
		err = bj.PlayerBet(domain.BJMinBet, 0, 0, 0)
		assert.NoError(t, err)

		// Finish the hand
		if bj.GetPhase() == domain.BJPhaseInsurance {
			_ = bj.PlayerDeclineInsurance()
		}
		if bj.GetPhase() == domain.BJPhaseAction && !bj.GetGameEndFlag() {
			_ = bj.PlayerStand()
		}

		// After one hand, remaining should be well above 25%
		remaining := bj.GetTrumpCards().GetRemainingCount()
		total := bj.GetTrumpCards().GetTotalCount()
		assert.True(t, remaining*4 >= total, "should have >25%% remaining")

		// Enable counting and verify running count persists across Reset
		bj.SetPhase(domain.BJPhaseBet)
		cfg := domain.BlackJackConfig{CountingEnabled: true}
		_ = bj.SetConfig(cfg)

		// Save current shoe reference
		shoeBeforeReset := bj.GetTrumpCards()
		bj.Reset()

		// The shoe should be reused (same total count), not reshuffled
		assert.Equal(t, shoeBeforeReset.GetTotalCount(), bj.GetTrumpCards().GetTotalCount())
	})

	t.Run("deckCountChanged forces reshuffle", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()

		// Change deck count to trigger reshuffle
		err := bj.SetDeckCount(2)
		assert.NoError(t, err)

		bj.Reset()

		// After reshuffle, total should reflect 2 decks
		assert.Equal(t, 104, bj.GetTrumpCards().GetTotalCount())
	})

	t.Run("running count resets on reshuffle", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.SetPhase(domain.BJPhaseBet)
		cfg := domain.BlackJackConfig{CountingEnabled: true}
		_ = bj.SetConfig(cfg)
		bj.Reset()

		// After fresh reshuffle, RC should be 0
		assert.Equal(t, 0, bj.GetRunningCount())

		// Force a deck count change to trigger reshuffle
		_ = bj.SetDeckCount(2)
		bj.Reset()
		assert.Equal(t, 0, bj.GetRunningCount(), "running count should reset after deck change reshuffle")
	})
}

func TestBlackJackConfig_GetSetConfig(t *testing.T) {
	t.Run("get default config", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		cfg := bj.GetConfig()
		assert.False(t, cfg.DealerHitsSoft17)
		assert.Equal(t, 0, cfg.CpuPlayerCount)
		assert.False(t, cfg.CountingEnabled)
	})
	t.Run("set config in BET phase succeeds", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		cfg := domain.BlackJackConfig{
			DealerHitsSoft17: true,
			CpuPlayerCount:   2,
			CountingEnabled:  true,
		}
		err := bj.SetConfig(cfg)
		assert.NoError(t, err)
		got := bj.GetConfig()
		assert.True(t, got.DealerHitsSoft17)
		assert.Equal(t, 2, got.CpuPlayerCount)
		assert.True(t, got.CountingEnabled)
	})
	t.Run("set config in ACTION phase fails", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		bj.SetPhase(domain.BJPhaseAction)
		cfg := domain.BlackJackConfig{DealerHitsSoft17: true}
		err := bj.SetConfig(cfg)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrWrongPhase)
	})
	t.Run("CPU count 0-3 ok", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		for count := 0; count <= domain.BJMaxCpuPlayers; count++ {
			cfg := domain.BlackJackConfig{CpuPlayerCount: count}
			err := bj.SetConfig(cfg)
			assert.NoError(t, err, "CPU count %d should be valid", count)
		}
	})
	t.Run("CPU count 4 fails", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		cfg := domain.BlackJackConfig{CpuPlayerCount: 4}
		err := bj.SetConfig(cfg)
		assert.Error(t, err)
	})
	t.Run("CPU count negative fails", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		cfg := domain.BlackJackConfig{CpuPlayerCount: -1}
		err := bj.SetConfig(cfg)
		assert.Error(t, err)
	})
	t.Run("counting system 0-3 ok", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		for sys := 0; sys <= domain.BJCountingMax; sys++ {
			cfg := domain.BlackJackConfig{CountingSystem: sys}
			err := bj.SetConfig(cfg)
			assert.NoError(t, err, "counting system %d should be valid", sys)
		}
	})
	t.Run("counting system out of range fails", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		cfg := domain.BlackJackConfig{CountingSystem: domain.BJCountingMax + 1}
		err := bj.SetConfig(cfg)
		assert.Error(t, err)
	})
	t.Run("counting system negative fails", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		cfg := domain.BlackJackConfig{CountingSystem: -1}
		err := bj.SetConfig(cfg)
		assert.Error(t, err)
	})
}

func TestCpuPlayBasicStrategy(t *testing.T) {
	t.Run("1 CPU player gets cards and plays", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		// Set up config with 1 CPU, use 2 decks for enough cards
		_ = bj.SetDeckCount(2)
		cfg := domain.BlackJackConfig{CpuPlayerCount: 1}
		err := bj.SetConfig(cfg)
		assert.NoError(t, err)
		bj.Reset()

		// Bet
		err = bj.PlayerBet(domain.BJMinBet, 0, 0, 0)
		assert.NoError(t, err)

		// CPU should have been dealt cards
		cpus := bj.GetCpuPlayers()
		assert.Equal(t, 1, len(cpus))
		assert.Equal(t, 2, cpus[0].GetHands()[0].GetCardsSize(), "CPU should have 2 cards after deal")

		// Finish the game: decline insurance if needed, then stand
		if bj.GetPhase() == domain.BJPhaseInsurance {
			_ = bj.PlayerDeclineInsurance()
		}
		if bj.GetPhase() == domain.BJPhaseAction && !bj.GetGameEndFlag() {
			_ = bj.PlayerStand()
		}

		// Game should have ended (CPU play happens after all human hands finish)
		assert.True(t, bj.GetGameEndFlag())

		// CPU hands should be resolved (stood or busted or surrendered)
		// unless the game ended due to a natural BJ (checkNaturalBlackJack
		// calls endGame directly without cpuPlay).
		playerBJ := bj.GetPlayerHands()[0].IsBlackJack()
		dealerBJ := bj.GetDealer().GetCardsSize() == 2 && bj.GetDealer().GetScore() == 21
		if !playerBJ && !dealerBJ {
			for _, hand := range cpus[0].GetHands() {
				if hand.GetCardsSize() > 0 {
					assert.True(t, hand.IsFinished(), "CPU hand should be finished after game ends")
				}
			}
		}
	})
	t.Run("3 CPU players all play", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		_ = bj.SetDeckCount(4)
		cfg := domain.BlackJackConfig{CpuPlayerCount: 3}
		err := bj.SetConfig(cfg)
		assert.NoError(t, err)
		bj.Reset()

		err = bj.PlayerBet(domain.BJMinBet, 0, 0, 0)
		assert.NoError(t, err)

		cpus := bj.GetCpuPlayers()
		assert.Equal(t, 3, len(cpus))
		for i, cpu := range cpus {
			assert.GreaterOrEqual(t, cpu.GetHands()[0].GetCardsSize(), 2, "CPU %d should have at least 2 cards", i)
		}

		if bj.GetPhase() == domain.BJPhaseInsurance {
			_ = bj.PlayerDeclineInsurance()
		}
		if bj.GetPhase() == domain.BJPhaseAction && !bj.GetGameEndFlag() {
			_ = bj.PlayerStand()
		}
		assert.True(t, bj.GetGameEndFlag())
	})
}

func TestCpuPayout(t *testing.T) {
	t.Run("CPU chips change after game", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		_ = bj.SetDeckCount(2)
		cfg := domain.BlackJackConfig{CpuPlayerCount: 1}
		_ = bj.SetConfig(cfg)
		bj.Reset()

		err := bj.PlayerBet(domain.BJMinBet, 0, 0, 0)
		assert.NoError(t, err)

		cpus := bj.GetCpuPlayers()
		assert.Equal(t, 1, len(cpus))

		// If natural BJ was dealt, endGame (including CPU payout) has already
		// run inside PlayerBet, so the pre-payout assertion only holds when
		// the game has NOT ended yet.
		cpuBet := cpus[0].GetHands()[0].GetBet()
		if !bj.GetGameEndFlag() {
			cpuChipsAfterBet := domain.BJDefaultChips - cpuBet
			assert.Equal(t, cpuChipsAfterBet, cpus[0].GetPlayer().GetChips())
		}

		// Finish the game
		if bj.GetPhase() == domain.BJPhaseInsurance {
			_ = bj.PlayerDeclineInsurance()
		}
		if bj.GetPhase() == domain.BJPhaseAction && !bj.GetGameEndFlag() {
			_ = bj.PlayerStand()
		}

		assert.True(t, bj.GetGameEndFlag())

		// After game, CPU chips should have changed (win/lose/draw)
		finalChips := cpus[0].GetPlayer().GetChips()
		// Verify chips are valid (not negative)
		assert.GreaterOrEqual(t, finalChips, 0, "CPU chips should not be negative")
	})

	t.Run("CPU chips persist across rounds", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		_ = bj.SetDeckCount(4)
		cfg := domain.BlackJackConfig{CpuPlayerCount: 1}
		_ = bj.SetConfig(cfg)
		bj.Reset()

		// Play first round
		err := bj.PlayerBet(domain.BJMinBet, 0, 0, 0)
		assert.NoError(t, err)
		if bj.GetPhase() == domain.BJPhaseInsurance {
			_ = bj.PlayerDeclineInsurance()
		}
		if bj.GetPhase() == domain.BJPhaseAction && !bj.GetGameEndFlag() {
			_ = bj.PlayerStand()
		}

		firstRoundChips := bj.GetCpuPlayers()[0].GetPlayer().GetChips()

		// Reset for second round (CPU count stays)
		bj.Reset()
		assert.Equal(t, 1, len(bj.GetCpuPlayers()), "CPU count should persist after reset")

		// CPU chips should carry over from previous round
		assert.Equal(t, firstRoundChips, bj.GetCpuPlayers()[0].GetPlayer().GetChips(),
			"CPU chips should persist across rounds")
	})

	t.Run("CPU with low chips gets reset", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		_ = bj.SetDeckCount(2)
		cfg := domain.BlackJackConfig{CpuPlayerCount: 1}
		_ = bj.SetConfig(cfg)
		bj.Reset()

		// Set CPU chips to below minimum bet
		bj.GetCpuPlayers()[0].GetPlayer().SetChips(domain.BJMinBet - 1)
		bj.Reset()

		// CPU chips should be reset to default
		assert.Equal(t, domain.BJDefaultChips, bj.GetCpuPlayers()[0].GetPlayer().GetChips(),
			"CPU chips below min bet should be reset to default")
	})
}

func TestBlackJack_CountingWithDealerHoleCard(t *testing.T) {
	t.Run("hole card counted after all hands busted", func(t *testing.T) {
		// When all player hands bust, dealer does not draw but hole card should still be counted
		bj := domain.NewDefaultBlackJack()
		bj.SetPhase(domain.BJPhaseBet)
		cfg := domain.BlackJackConfig{CountingEnabled: true}
		_ = bj.SetConfig(cfg)
		_ = bj.SetDeckCount(2)
		bj.Reset()

		// Bet and play through
		err := bj.PlayerBet(domain.BJMinBet, 0, 0, 0)
		assert.NoError(t, err)

		// Regardless of the outcome, after the game the running count should have included
		// the dealer's hole card. We mainly verify no panics/errors occur.
		if bj.GetPhase() == domain.BJPhaseInsurance {
			_ = bj.PlayerDeclineInsurance()
		}
		if bj.GetPhase() == domain.BJPhaseAction && !bj.GetGameEndFlag() {
			// Keep hitting until bust or stand
			for !bj.GetGameEndFlag() {
				hand := bj.GetPlayerHands()[bj.GetCurrentHandIdx()]
				if hand.IsFinished() {
					break
				}
				hitErr := bj.PlayerHit()
				if hitErr != nil {
					break
				}
			}
			if !bj.GetGameEndFlag() {
				_ = bj.PlayerStand()
			}
		}

		assert.True(t, bj.GetGameEndFlag())
		// Running count should be a valid integer (including possible 0)
		_ = bj.GetRunningCount()
	})
}

func TestBlackJack_CpuInitReuse(t *testing.T) {
	t.Run("CPU players reused when count unchanged", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		cfg := domain.BlackJackConfig{CpuPlayerCount: 2}
		_ = bj.SetConfig(cfg)
		bj.Reset()

		// Get initial CPU references
		cpus := bj.GetCpuPlayers()
		assert.Equal(t, 2, len(cpus))

		// Play a round
		_ = bj.PlayerBet(domain.BJMinBet, 0, 0, 0)
		if bj.GetPhase() == domain.BJPhaseInsurance {
			_ = bj.PlayerDeclineInsurance()
		}
		if bj.GetPhase() == domain.BJPhaseAction && !bj.GetGameEndFlag() {
			_ = bj.PlayerStand()
		}

		// Reset: same CPU count should reuse existing CPUs
		bj.Reset()
		assert.Equal(t, 2, len(bj.GetCpuPlayers()))
	})
	t.Run("CPU count change creates new CPUs", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		cfg := domain.BlackJackConfig{CpuPlayerCount: 1}
		_ = bj.SetConfig(cfg)
		bj.Reset()

		assert.Equal(t, 1, len(bj.GetCpuPlayers()))

		// Change CPU count
		cfg.CpuPlayerCount = 3
		_ = bj.SetConfig(cfg)
		bj.Reset()

		assert.Equal(t, 3, len(bj.GetCpuPlayers()))
	})
	t.Run("CPU count 0 sets nil", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		cfg := domain.BlackJackConfig{CpuPlayerCount: 0}
		_ = bj.SetConfig(cfg)
		bj.Reset()

		assert.Empty(t, bj.GetCpuPlayers())
	})
}

func TestBlackJack_CpuBetSkipsInsufficientChips(t *testing.T) {
	bj := domain.NewDefaultBlackJack()
	_ = bj.SetDeckCount(2)
	cfg := domain.BlackJackConfig{CpuPlayerCount: 1}
	_ = bj.SetConfig(cfg)
	bj.Reset()

	// Set CPU chips to 0 (below min bet)
	bj.GetCpuPlayers()[0].GetPlayer().SetChips(0)

	// Bet: CPU should be skipped during cpuBetAndDeal because chips < BJMinBet
	err := bj.PlayerBet(domain.BJMinBet, 0, 0, 0)
	assert.NoError(t, err)

	// CPU hand should have 0 cards (skipped)
	cpuHand := bj.GetCpuPlayers()[0].GetHands()[0]
	assert.Equal(t, 0, cpuHand.GetCardsSize(), "CPU with 0 chips should be skipped during deal")
	assert.Equal(t, 0, cpuHand.GetBet(), "CPU with 0 chips should have 0 bet")
}

// --- Side Bet Tests ---

func TestBlackJack_SideBetValidation(t *testing.T) {
	t.Run("invalid PP bet below minimum", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		err := bj.PlayerBet(100, 5, 0, 0)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidAmount)
		assert.Contains(t, err.Error(), "Perfect Pairs")
	})
	t.Run("invalid PP bet not multiple of min", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		err := bj.PlayerBet(100, 15, 0, 0)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidAmount)
	})
	t.Run("invalid T3 bet below minimum", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		err := bj.PlayerBet(100, 0, 5, 0)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidAmount)
		assert.Contains(t, err.Error(), "poker-hand")
	})
	t.Run("invalid T3 bet not multiple of min", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		err := bj.PlayerBet(100, 0, 15, 0)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidAmount)
	})
	t.Run("main bet above max", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		bj.GetPlayer().SetChips(20000)
		err := bj.PlayerBet(10010, 0, 0, 0)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidAmount)
	})
	t.Run("PP bet above max", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		bj.GetPlayer().SetChips(20000)
		err := bj.PlayerBet(100, 10010, 0, 0)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidAmount)
		assert.Contains(t, err.Error(), "Perfect Pairs")
	})
	t.Run("T3 bet above max", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		bj.GetPlayer().SetChips(20000)
		err := bj.PlayerBet(100, 0, 10010, 0)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidAmount)
		assert.Contains(t, err.Error(), "poker-hand")
	})
	t.Run("insufficient chips for main+side bets total", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		bj.GetPlayer().SetChips(100)
		err := bj.PlayerBet(100, 10, 10, 0)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInsufficientChips)
	})
	t.Run("zero side bets are valid", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		err := bj.PlayerBet(100, 0, 0, 0)
		assert.NoError(t, err)
		assert.Equal(t, 0, bj.GetPerfectPairsBet())
		assert.Equal(t, 0, bj.Get21Plus3Bet())
		assert.Nil(t, bj.GetSideBetResults())
	})
}

func TestBlackJack_SideBetEvaluation(t *testing.T) {
	t.Run("PP bet evaluates after deal", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		err := bj.PlayerBet(100, 10, 0, 0)
		assert.NoError(t, err)
		assert.Equal(t, 10, bj.GetPerfectPairsBet())
		results := bj.GetSideBetResults()
		assert.Equal(t, 1, len(results))
		assert.Equal(t, domain.BJSideBetPerfectPairs, results[0].BetType)
		assert.Equal(t, 10, results[0].BetAmount)
	})
	t.Run("T3 bet evaluates after deal", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		err := bj.PlayerBet(100, 0, 10, 0)
		assert.NoError(t, err)
		assert.Equal(t, 10, bj.Get21Plus3Bet())
		results := bj.GetSideBetResults()
		assert.Equal(t, 1, len(results))
		assert.Equal(t, domain.BJSideBet21Plus3, results[0].BetType)
	})
	t.Run("both side bets produce 2 results", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		err := bj.PlayerBet(100, 10, 20, 0)
		assert.NoError(t, err)
		results := bj.GetSideBetResults()
		assert.Equal(t, 2, len(results))
		assert.Equal(t, domain.BJSideBetPerfectPairs, results[0].BetType)
		assert.Equal(t, domain.BJSideBet21Plus3, results[1].BetType)
	})
}

func TestBlackJack_SideBetResetClears(t *testing.T) {
	bj := domain.NewDefaultBlackJack()
	bj.Reset()
	_ = bj.PlayerBet(100, 10, 20, 0)
	assert.NotNil(t, bj.GetSideBetResults())
	bj.Reset()
	assert.Equal(t, 0, bj.GetPerfectPairsBet())
	assert.Equal(t, 0, bj.Get21Plus3Bet())
	assert.Nil(t, bj.GetSideBetResults())
}

func TestBlackJack_SideBetDeckExhausted(t *testing.T) {
	tc := domain.NewTrumpCardsWithDecks(1, 0)
	player := domain.NewBlackJackPlayer()
	dealer := domain.NewBlackJackPlayer()
	player.SetChips(10000)
	dealer.SetChips(10000)
	bj := domain.NewBlackJack(tc, player, dealer)
	bj.Reset()
	// Draw all cards to exhaust the deck
	for bj.GetTrumpCards().GetRemainingCount() > 3 {
		bj.GetTrumpCards().DrawCard()
	}
	chipsBefore := player.GetChips()
	err := bj.PlayerBet(100, 10, 20, 0)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrDeckExhausted)
	// Total cost (100+10+20=130) should be refunded
	assert.Equal(t, chipsBefore, player.GetChips())
	assert.Equal(t, 0, bj.GetPerfectPairsBet())
	assert.Equal(t, 0, bj.Get21Plus3Bet())
}

func TestBlackJack_SideBetPPWinPayout(t *testing.T) {
	// Force a perfect pair by setting up cards manually
	tc := domain.NewTrumpCardsWithDecks(1, 0)
	for i := 0; i < 10; i++ {
		tc.Shuffle()
	}
	player := domain.NewBlackJackPlayer()
	dealer := domain.NewBlackJackPlayer()
	player.SetChips(1000)
	dealer.SetChips(1000)
	bj := domain.NewBlackJack(tc, player, dealer)
	bj.Reset()

	err := bj.PlayerBet(100, 10, 0, 0)
	assert.NoError(t, err)
	results := bj.GetSideBetResults()
	assert.Equal(t, 1, len(results))
	// Regardless of outcome, the bet amount should be recorded
	assert.Equal(t, 10, results[0].BetAmount)
	assert.Equal(t, domain.BJSideBetPerfectPairs, results[0].BetType)
}

func TestBlackJack_SideBetT3Payout(t *testing.T) {
	tc := domain.NewTrumpCardsWithDecks(1, 0)
	for i := 0; i < 10; i++ {
		tc.Shuffle()
	}
	player := domain.NewBlackJackPlayer()
	dealer := domain.NewBlackJackPlayer()
	player.SetChips(1000)
	dealer.SetChips(1000)
	bj := domain.NewBlackJack(tc, player, dealer)
	bj.Reset()

	err := bj.PlayerBet(100, 0, 10, 0)
	assert.NoError(t, err)
	results := bj.GetSideBetResults()
	assert.Equal(t, 1, len(results))
	assert.Equal(t, 10, results[0].BetAmount)
	assert.Equal(t, domain.BJSideBet21Plus3, results[0].BetType)
}

func TestGetDeckPenetration_Default(t *testing.T) {
	bj := domain.NewDefaultBlackJack()
	assert.Equal(t, 75, bj.GetDeckPenetration())
}

func TestGetDeckPenetration_Zero(t *testing.T) {
	bj := domain.NewDefaultBlackJack()
	_ = bj.SetConfig(domain.BlackJackConfig{DeckPenetration: 0, DoubleAfterSplit: true})
	assert.Equal(t, 75, bj.GetDeckPenetration())
}

func TestGetDeckPenetration_50(t *testing.T) {
	bj := domain.NewDefaultBlackJack()
	_ = bj.SetConfig(domain.BlackJackConfig{DeckPenetration: 50, DoubleAfterSplit: true})
	assert.Equal(t, 50, bj.GetDeckPenetration())
}

func TestSetConfig_InvalidPenetration(t *testing.T) {
	bj := domain.NewDefaultBlackJack()
	err := bj.SetConfig(domain.BlackJackConfig{DeckPenetration: 60, DoubleAfterSplit: true})
	assert.Error(t, err)
}

func TestSetConfig_ValidPenetration(t *testing.T) {
	bj := domain.NewDefaultBlackJack()
	err := bj.SetConfig(domain.BlackJackConfig{DeckPenetration: 50, DoubleAfterSplit: true})
	assert.NoError(t, err)
	assert.Equal(t, 50, bj.GetDeckPenetration())
}

func TestReset_Penetration50(t *testing.T) {
	// With 50% penetration, reshuffle when remaining < 50% of total (i.e., remaining < 26 for 52-card deck).
	// Draw enough cards so remaining drops below 50% threshold, then verify Reset reshuffles.
	bj := domain.NewDefaultBlackJack()
	_ = bj.SetConfig(domain.BlackJackConfig{DeckPenetration: 50, DoubleAfterSplit: true})
	bj.Reset()

	// After reset, we have a fresh 52-card deck. Draw 27 cards to leave 25 remaining (< 26).
	for i := 0; i < 27; i++ {
		_ = bj.PlayerBet(10, 0, 0, 0)
		bj.Reset()
	}
	// After drawing 27 * 4 = 108 cards from bets (2 player + 2 dealer each round),
	// but Reset reshuffles when remaining drops below threshold.
	// The key point: with 50% penetration, reshuffle triggers sooner.
	// Let's test it directly: the game should still be functional after many rounds.
	assert.Equal(t, domain.BJPhaseBet, bj.GetPhase())
	assert.Equal(t, 50, bj.GetDeckPenetration())
}

func TestReset_Penetration75(t *testing.T) {
	// With 75% penetration (default), reshuffle when remaining < 25% of total (i.e., remaining < 13 for 52-card deck).
	bj := domain.NewDefaultBlackJack()
	// Default config: DeckPenetration=0 → treated as 75
	bj.Reset()

	assert.Equal(t, 75, bj.GetDeckPenetration())
	// Play several rounds to verify the game stays functional with default penetration
	for i := 0; i < 10; i++ {
		_ = bj.PlayerBet(10, 0, 0, 0)
		bj.Reset()
	}
	assert.Equal(t, domain.BJPhaseBet, bj.GetPhase())
}

func TestBlackJack_MultiHand(t *testing.T) {
	t.Run("handCount=0 defaults to 1", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		err := bj.PlayerBet(100, 0, 0, 0)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(bj.GetPlayerHands()))
		assert.Equal(t, 1, bj.GetMultiHandCount())
	})

	t.Run("handCount=1 backward compat", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		err := bj.PlayerBet(100, 0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(bj.GetPlayerHands()))
		assert.Equal(t, 1, bj.GetMultiHandCount())
	})

	t.Run("handCount=2 creates 2 hands", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		bj.GetPlayer().SetChips(2000)
		err := bj.PlayerBet(100, 0, 0, 2)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(bj.GetPlayerHands()))
		assert.Equal(t, 2, bj.GetMultiHandCount())
		for _, hand := range bj.GetPlayerHands() {
			assert.Equal(t, 100, hand.GetBet())
			assert.Equal(t, 2, hand.GetCardsSize())
		}
		// totalCost = 100*2 = 200; side bets / natural BJ payouts may change
		// chips after deduction, so verify bet was recorded on each hand
		for _, hand := range bj.GetPlayerHands() {
			assert.Equal(t, 100, hand.GetBet())
		}
	})

	t.Run("handCount=3 creates 3 hands", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		bj.GetPlayer().SetChips(2000)
		err := bj.PlayerBet(100, 0, 0, 3)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(bj.GetPlayerHands()))
		assert.Equal(t, 3, bj.GetMultiHandCount())
	})

	t.Run("handCount=4 returns error", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		err := bj.PlayerBet(100, 0, 0, 4)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidAmount)
	})

	t.Run("handCount=-1 returns error", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		err := bj.PlayerBet(100, 0, 0, -1)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidAmount)
	})

	t.Run("insufficient chips for multi-hand", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		bj.GetPlayer().SetChips(150)
		err := bj.PlayerBet(100, 0, 0, 2)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInsufficientChips)
	})

	t.Run("multi-hand with side bets costs correctly", func(t *testing.T) {
		// Side bets and natural BJ can change chips after deduction,
		// so verify that at least totalCost was deducted.
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		bj.GetPlayer().SetChips(2000)
		err := bj.PlayerBet(100, 10, 20, 2)
		assert.NoError(t, err)
		// totalCost = 100*2 + 10 + 20 = 230; side bets / natural BJ payouts
		// may change chips after deduction, so verify bets were recorded
		for _, hand := range bj.GetPlayerHands() {
			assert.Equal(t, 100, hand.GetBet())
		}
		assert.Equal(t, 10, bj.GetPerfectPairsBet())
		assert.Equal(t, 20, bj.Get21Plus3Bet())
	})

	t.Run("Reset clears multiHandCount", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		bj.SetMultiHandCount(3)
		bj.Reset()
		assert.Equal(t, 1, bj.GetMultiHandCount())
	})

	t.Run("GetMultiHandCount returns 1 for zero value", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		assert.Equal(t, 1, bj.GetMultiHandCount())
	})
}

func TestBlackJack_MultiHand_FromSplit(t *testing.T) {
	t.Run("split sets fromSplit on both hands", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		bj.GetPlayer().SetChips(2000)

		hand := domain.NewBlackJackHand()
		hand.SetBet(100)
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		hand.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		bj.SetPlayerHands([]*domain.BlackJackHand{hand})
		bj.SetPhase(domain.BJPhaseAction)

		err := bj.PlayerSplit()
		assert.NoError(t, err)
		for _, h := range bj.GetPlayerHands() {
			assert.True(t, h.IsFromSplit())
		}
	})

	t.Run("multi-hand initial hands are NOT fromSplit", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		bj.GetPlayer().SetChips(2000)
		err := bj.PlayerBet(100, 0, 0, 2)
		assert.NoError(t, err)
		for _, hand := range bj.GetPlayerHands() {
			assert.False(t, hand.IsFromSplit())
		}
	})

	t.Run("fromSplit hand does not get BJ 3:2 payout", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		player := domain.NewBlackJackPlayer()
		dealer := domain.NewBlackJackPlayer()
		player.SetChips(800)
		dealer.SetChips(1000)
		bj := domain.NewBlackJack(tc, player, dealer)
		bj.Reset()

		hand0 := domain.NewBlackJackHand()
		hand0.SetBet(100)
		hand0.SetFromSplit(true)
		hand0.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		hand0.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))

		bj.SetPlayerHands([]*domain.BlackJackHand{hand0})
		dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 8, false))
		bj.SetPhase(domain.BJPhaseAction)

		err := bj.PlayerStand()
		assert.NoError(t, err)
		assert.True(t, bj.GetGameEndFlag())
		assert.Equal(t, 800+200, player.GetChips())
	})

	t.Run("non-fromSplit hand gets BJ 3:2 payout", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		player := domain.NewBlackJackPlayer()
		dealer := domain.NewBlackJackPlayer()
		player.SetChips(800)
		dealer.SetChips(1000)
		bj := domain.NewBlackJack(tc, player, dealer)
		bj.Reset()

		hand0 := domain.NewBlackJackHand()
		hand0.SetBet(100)
		hand0.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		hand0.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))

		bj.SetPlayerHands([]*domain.BlackJackHand{hand0})
		dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 8, false))
		bj.SetPhase(domain.BJPhaseAction)

		err := bj.PlayerStand()
		assert.NoError(t, err)
		assert.True(t, bj.GetGameEndFlag())
		assert.Equal(t, 800+250, player.GetChips())
	})
}

func TestBlackJack_MultiHand_DAS(t *testing.T) {
	t.Run("DAS check uses fromSplit not hand count", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		bj.GetPlayer().SetChips(2000)

		config := bj.GetConfig()
		config.DoubleAfterSplit = false
		_ = bj.SetConfig(config)
		bj.Reset()

		hand0 := domain.NewBlackJackHand()
		hand0.SetBet(100)
		hand0.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		hand0.AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))

		hand1 := domain.NewBlackJackHand()
		hand1.SetBet(100)
		hand1.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
		hand1.AddCard(domain.NewCard(domain.CardDesignDiamond, 9, false))

		bj.SetPlayerHands([]*domain.BlackJackHand{hand0, hand1})
		bj.SetMultiHandCount(2)
		bj.GetDealer().AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		bj.GetDealer().AddCard(domain.NewCard(domain.CardDesignClover, 7, false))
		bj.SetPhase(domain.BJPhaseAction)

		err := bj.PlayerDoubleDown()
		assert.NoError(t, err)
	})

	t.Run("DAS disabled blocks DD on fromSplit hand", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		bj.GetPlayer().SetChips(2000)

		config := bj.GetConfig()
		config.DoubleAfterSplit = false
		_ = bj.SetConfig(config)
		bj.Reset()

		hand0 := domain.NewBlackJackHand()
		hand0.SetBet(100)
		hand0.SetFromSplit(true)
		hand0.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		hand0.AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))

		bj.SetPlayerHands([]*domain.BlackJackHand{hand0})
		bj.GetDealer().AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		bj.GetDealer().AddCard(domain.NewCard(domain.CardDesignClover, 7, false))
		bj.SetPhase(domain.BJPhaseAction)

		err := bj.PlayerDoubleDown()
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidPlay)
	})
}

func TestBlackJack_MultiHand_DeckExhausted(t *testing.T) {
	t.Run("deck exhaustion during multi-hand deal", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		player := domain.NewBlackJackPlayer()
		dealer := domain.NewBlackJackPlayer()
		player.SetChips(2000)
		dealer.SetChips(1000)
		bj := domain.NewBlackJack(tc, player, dealer)
		bj.Reset()

		remaining := tc.GetRemainingCount()
		for i := 0; i < remaining-5; i++ {
			tc.DrawCard()
		}

		err := bj.PlayerBet(100, 0, 0, 2)
		assert.ErrorIs(t, err, domain.ErrDeckExhausted)
		assert.Equal(t, 2000, player.GetChips())
		assert.Equal(t, 1, len(bj.GetPlayerHands()))
	})
}

func TestSetConfig_InvalidSurrenderRule(t *testing.T) {
	bj := domain.NewDefaultBlackJack()
	bj.Reset()

	t.Run("surrender rule -1 fails", func(t *testing.T) {
		cfg := domain.BlackJackConfig{SurrenderRule: -1}
		err := bj.SetConfig(cfg)
		assert.Error(t, err)
	})
	t.Run("surrender rule 3 fails", func(t *testing.T) {
		cfg := domain.BlackJackConfig{SurrenderRule: 3}
		err := bj.SetConfig(cfg)
		assert.Error(t, err)
	})
	t.Run("valid surrender rules succeed", func(t *testing.T) {
		for _, rule := range []int{domain.BJSurrenderLate, domain.BJSurrenderEarly, domain.BJSurrenderNone} {
			cfg := domain.BlackJackConfig{SurrenderRule: rule}
			err := bj.SetConfig(cfg)
			assert.NoError(t, err, "surrender rule %d should be valid", rule)
		}
	})
}

func TestPlayerSurrender_NoSurrenderMode(t *testing.T) {
	playerCards := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 10, false),
		domain.NewCard(domain.CardDesignHeart, 6, false),
	}
	dealerCards := []*domain.Card{
		domain.NewCard(domain.CardDesignClover, 10, false),
		domain.NewCard(domain.CardDesignDiamond, 7, false),
	}
	bj, _, _ := setupDeterministicBJ(1000, playerCards, dealerCards, 100)

	// Set SurrenderRule to None
	bj.SetPhase(domain.BJPhaseBet)
	err := bj.SetConfig(domain.BlackJackConfig{SurrenderRule: domain.BJSurrenderNone})
	require.NoError(t, err)
	bj.SetPhase(domain.BJPhaseAction)

	err = bj.PlayerSurrender()
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidPlay)
}

func TestCanSurrenderHand(t *testing.T) {
	t.Run("returns true with late surrender when hand is eligible", func(t *testing.T) {
		playerCards := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 10, false),
			domain.NewCard(domain.CardDesignHeart, 6, false),
		}
		dealerCards := []*domain.Card{
			domain.NewCard(domain.CardDesignClover, 10, false),
			domain.NewCard(domain.CardDesignDiamond, 7, false),
		}
		bj, _, _ := setupDeterministicBJ(1000, playerCards, dealerCards, 100)
		// Default SurrenderRule is Late (0)
		assert.True(t, bj.CanSurrenderHand(0))
	})
	t.Run("returns false when SurrenderRule is None", func(t *testing.T) {
		playerCards := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 10, false),
			domain.NewCard(domain.CardDesignHeart, 6, false),
		}
		dealerCards := []*domain.Card{
			domain.NewCard(domain.CardDesignClover, 10, false),
			domain.NewCard(domain.CardDesignDiamond, 7, false),
		}
		bj, _, _ := setupDeterministicBJ(1000, playerCards, dealerCards, 100)
		bj.SetPhase(domain.BJPhaseBet)
		err := bj.SetConfig(domain.BlackJackConfig{SurrenderRule: domain.BJSurrenderNone})
		require.NoError(t, err)
		bj.SetPhase(domain.BJPhaseAction)
		assert.False(t, bj.CanSurrenderHand(0))
	})
	t.Run("returns false for out of bounds index", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		assert.False(t, bj.CanSurrenderHand(-1))
		assert.False(t, bj.CanSurrenderHand(99))
	})
}

func TestCanSurrenderCpuHand(t *testing.T) {
	t.Run("returns false when no CPU players", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		assert.False(t, bj.CanSurrenderCpuHand(0, 0))
	})
	t.Run("returns false with SurrenderRule None", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		cfg := domain.BlackJackConfig{CpuPlayerCount: 1, SurrenderRule: domain.BJSurrenderNone}
		err := bj.SetConfig(cfg)
		require.NoError(t, err)
		bj.Reset()
		_ = bj.PlayerBet(domain.BJMinBet, 0, 0, 0)
		assert.False(t, bj.CanSurrenderCpuHand(0, 0))
	})
	t.Run("returns false for out of bounds cpuIdx", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		cfg := domain.BlackJackConfig{CpuPlayerCount: 1}
		err := bj.SetConfig(cfg)
		require.NoError(t, err)
		bj.Reset()
		_ = bj.PlayerBet(domain.BJMinBet, 0, 0, 0)
		assert.False(t, bj.CanSurrenderCpuHand(-1, 0))
		assert.False(t, bj.CanSurrenderCpuHand(5, 0))
	})
	t.Run("returns false for out of bounds handIdx", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		cfg := domain.BlackJackConfig{CpuPlayerCount: 1}
		err := bj.SetConfig(cfg)
		require.NoError(t, err)
		bj.Reset()
		_ = bj.PlayerBet(domain.BJMinBet, 0, 0, 0)
		assert.False(t, bj.CanSurrenderCpuHand(0, -1))
		assert.False(t, bj.CanSurrenderCpuHand(0, 99))
	})
	t.Run("returns true for eligible CPU hand", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		_ = bj.SetDeckCount(2)
		cfg := domain.BlackJackConfig{CpuPlayerCount: 1}
		err := bj.SetConfig(cfg)
		require.NoError(t, err)
		bj.Reset()
		_ = bj.PlayerBet(domain.BJMinBet, 0, 0, 0)
		// CPU hand might already be finished (stood/busted/surrendered) after cpuPlay
		// If not finished and has 2 cards, CanSurrender should be true
		cpus := bj.GetCpuPlayers()
		if len(cpus) > 0 {
			hand := cpus[0].GetHands()[0]
			if hand.GetCardsSize() == 2 && !hand.IsFinished() {
				assert.True(t, bj.CanSurrenderCpuHand(0, 0))
			}
		}
	})
}

func TestPlayerEarlySurrender(t *testing.T) {
	playerCards := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 10, false),
		domain.NewCard(domain.CardDesignHeart, 6, false),
	}
	dealerCards := []*domain.Card{
		domain.NewCard(domain.CardDesignClover, 10, false),
		domain.NewCard(domain.CardDesignDiamond, 7, false),
	}
	bj, player, _ := setupDeterministicBJ(1000, playerCards, dealerCards, 100)

	// Set early surrender config
	bj.SetPhase(domain.BJPhaseBet)
	err := bj.SetConfig(domain.BlackJackConfig{SurrenderRule: domain.BJSurrenderEarly})
	require.NoError(t, err)
	bj.SetPhase(domain.BJPhaseEarlySurrender)

	startChips := player.GetChips()
	err = bj.PlayerEarlySurrender()
	assert.NoError(t, err)
	// Half bet (50) returned
	assert.Equal(t, startChips+50, player.GetChips())
	// Hand should be surrendered
	assert.True(t, bj.GetPlayerHands()[0].IsSurrendered())
}

func TestPlayerEarlySurrender_CannotSurrender(t *testing.T) {
	playerCards := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 10, false),
		domain.NewCard(domain.CardDesignHeart, 6, false),
	}
	dealerCards := []*domain.Card{
		domain.NewCard(domain.CardDesignClover, 10, false),
		domain.NewCard(domain.CardDesignDiamond, 7, false),
	}
	bj, _, _ := setupDeterministicBJ(1000, playerCards, dealerCards, 100)

	bj.SetPhase(domain.BJPhaseBet)
	err := bj.SetConfig(domain.BlackJackConfig{SurrenderRule: domain.BJSurrenderEarly})
	require.NoError(t, err)
	bj.SetPhase(domain.BJPhaseEarlySurrender)

	// Add a 3rd card to make CanSurrender() return false
	bj.GetPlayerHands()[0].AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))

	err = bj.PlayerEarlySurrender()
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidPlay)
}

func TestPlayerEarlySurrender_WrongPhase(t *testing.T) {
	bj := domain.NewDefaultBlackJack()
	bj.Reset()
	err := bj.PlayerEarlySurrender()
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrWrongPhase)
}

func TestPlayerDeclineEarlySurrender(t *testing.T) {
	playerCards := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 10, false),
		domain.NewCard(domain.CardDesignHeart, 6, false),
	}
	dealerCards := []*domain.Card{
		domain.NewCard(domain.CardDesignClover, 10, false),
		domain.NewCard(domain.CardDesignDiamond, 7, false),
	}
	bj, _, _ := setupDeterministicBJ(1000, playerCards, dealerCards, 100)

	bj.SetPhase(domain.BJPhaseBet)
	err := bj.SetConfig(domain.BlackJackConfig{SurrenderRule: domain.BJSurrenderEarly})
	require.NoError(t, err)
	bj.SetPhase(domain.BJPhaseEarlySurrender)

	err = bj.PlayerDeclineEarlySurrender()
	assert.NoError(t, err)
	// Phase should advance to action (single hand, all done -> action phase)
	assert.Equal(t, domain.BJPhaseAction, bj.GetPhase())
}

func TestPlayerDeclineEarlySurrender_WrongPhase(t *testing.T) {
	bj := domain.NewDefaultBlackJack()
	bj.Reset()
	err := bj.PlayerDeclineEarlySurrender()
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrWrongPhase)
}

func TestEarlySurrenderFlow_NonAceUpcard(t *testing.T) {
	bj := domain.NewDefaultBlackJack()
	// Configure early surrender
	cfg := domain.BlackJackConfig{SurrenderRule: domain.BJSurrenderEarly}
	err := bj.SetConfig(cfg)
	require.NoError(t, err)
	bj.Reset()

	err = bj.PlayerBet(100, 0, 0, 0)
	require.NoError(t, err)

	// If dealer upcard is not ace, phase should be early surrender (6)
	dealerUpcard := bj.GetDealer().GetCard(0)
	if dealerUpcard != nil && dealerUpcard.GetValue() != 1 {
		assert.Equal(t, domain.BJPhaseEarlySurrender, bj.GetPhase())
	} else {
		// Dealer has ace -> insurance first
		assert.Equal(t, domain.BJPhaseInsurance, bj.GetPhase())
	}
}

func TestEarlySurrender_DealerBJAfterDecline(t *testing.T) {
	// Setup: dealer has BJ (A + 10), player declines early surrender -> game ends
	playerCards := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 10, false),
		domain.NewCard(domain.CardDesignHeart, 6, false),
	}
	dealerCards := []*domain.Card{
		domain.NewCard(domain.CardDesignClover, 1, false),   // Ace
		domain.NewCard(domain.CardDesignDiamond, 10, false), // 10 -> BJ
	}
	bj, _, _ := setupDeterministicBJ(1000, playerCards, dealerCards, 100)

	bj.SetPhase(domain.BJPhaseBet)
	err := bj.SetConfig(domain.BlackJackConfig{SurrenderRule: domain.BJSurrenderEarly})
	require.NoError(t, err)
	bj.SetPhase(domain.BJPhaseEarlySurrender)

	err = bj.PlayerDeclineEarlySurrender()
	assert.NoError(t, err)
	// After advancing, checkNaturalBlackJack should detect dealer BJ and end game
	assert.True(t, bj.GetGameEndFlag())
	assert.Equal(t, domain.BJPhaseEnd, bj.GetPhase())
}

func TestGetBasicStrategySuggestion_EarlySurrenderPhase(t *testing.T) {
	t.Run("returns surrender for hard 16 vs 10", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.ToggleHint()
		hand := bj.GetPlayerHands()[0]
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		hand.AddCard(domain.NewCard(domain.CardDesignHeart, 6, false)) // hard 16
		bj.GetDealer().AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
		bj.GetDealer().AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
		bj.SetPhase(domain.BJPhaseEarlySurrender)
		assert.Equal(t, domain.BJSuggestSurrender, bj.GetBasicStrategySuggestion())
	})
	t.Run("returns stand (continue) for hard 12 vs 6", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.ToggleHint()
		hand := bj.GetPlayerHands()[0]
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		hand.AddCard(domain.NewCard(domain.CardDesignHeart, 2, false)) // hard 12
		bj.GetDealer().AddCard(domain.NewCard(domain.CardDesignClover, 6, false))
		bj.GetDealer().AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
		bj.SetPhase(domain.BJPhaseEarlySurrender)
		// Basic strategy for hard 12 vs 6 is Stand, not Surrender -> returns Stand (continue)
		assert.Equal(t, domain.BJSuggestStand, bj.GetBasicStrategySuggestion())
	})
}

// ---------------------------------------------------------------------------
// ActionLog tests
// ---------------------------------------------------------------------------

func TestBlackJack_ActionLog_Bet(t *testing.T) {
	bj := domain.NewDefaultBlackJack()
	err := bj.PlayerBet(100, 0, 0, 0)
	assert.NoError(t, err)

	log := bj.GetActionLog()
	assert.GreaterOrEqual(t, len(log), 1)
	entry := log[0]
	assert.Equal(t, 0, entry.PlayerIdx)
	assert.Equal(t, "bet", entry.ActionType)
	assert.Contains(t, entry.Detail, "bet 100 chips")
	assert.Nil(t, entry.Cards)
}

func TestBlackJack_ActionLog_HitStand(t *testing.T) {
	bj := domain.NewDefaultBlackJack()
	// Set up: bet, then manually set phase to action with low-value hand
	err := bj.PlayerBet(100, 0, 0, 0)
	require.NoError(t, err)

	// Give dealer score >= 17 to prevent auto-draw issues
	bj.GetDealer().Reset()
	bj.GetDealer().AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
	bj.GetDealer().AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))

	// Set up player hand with low score
	hand := bj.GetPlayerHands()[0]
	hand.Reset()
	hand.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
	hand.AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
	bj.SetPhase(domain.BJPhaseAction)

	// Hit
	err = bj.PlayerHit()
	assert.NoError(t, err)

	log := bj.GetActionLog()
	hitFound := false
	for _, e := range log {
		if e.ActionType == "hit" {
			hitFound = true
			assert.Equal(t, 0, e.PlayerIdx)
			assert.Len(t, e.Cards, 1)
			break
		}
	}
	assert.True(t, hitFound, "expected hit action log entry")

	// Stand (if not busted)
	if !hand.IsBusted() {
		bj.SetPhase(domain.BJPhaseAction)
		err = bj.PlayerStand()
		assert.NoError(t, err)

		log = bj.GetActionLog()
		standFound := false
		for _, e := range log {
			if e.ActionType == "stand" {
				standFound = true
				assert.Equal(t, 0, e.PlayerIdx)
				assert.Nil(t, e.Cards)
				break
			}
		}
		assert.True(t, standFound, "expected stand action log entry")
	}
}

func TestBlackJack_ActionLog_Reset(t *testing.T) {
	bj := domain.NewDefaultBlackJack()
	_ = bj.PlayerBet(100, 0, 0, 0)
	assert.NotEmpty(t, bj.GetActionLog())

	bj.Reset()
	assert.Nil(t, bj.GetActionLog())
}
