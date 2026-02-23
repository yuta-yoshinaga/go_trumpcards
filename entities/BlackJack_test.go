package entities_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"

	"github.com/stretchr/testify/assert"
)

func TestNewDefaultBlackJack(t *testing.T) {
	bj := entities.NewDefaultBlackJack()
	assert.NotNil(t, bj)
	assert.NotNil(t, bj.GetPlayer())
	assert.NotNil(t, bj.GetDealer())
	assert.Equal(t, false, bj.GetGameEndFlag())
	assert.Equal(t, entities.BJPhaseBet, bj.GetPhase())
	assert.Equal(t, entities.BJDefaultChips, bj.GetPlayer().GetChips())
	assert.Equal(t, entities.BJDefaultChips, bj.GetDealer().GetChips())
}

func TestBlackJack_Reset(t *testing.T) {
	bj := entities.NewDefaultBlackJack()
	bj.Reset()
	assert.Equal(t, entities.BJPhaseBet, bj.GetPhase())
	assert.False(t, bj.GetGameEndFlag())
	assert.Equal(t, 0, bj.GetCurrentHandIdx())
	assert.Equal(t, 0, bj.GetInsuranceBet())
	assert.False(t, bj.IsInsuranceAvailable())
	assert.Equal(t, 1, len(bj.GetPlayerHands()))
	assert.Equal(t, 0, bj.GetPlayer().GetCardsSize())
}

func TestBlackJack_ResetChipsIfZero(t *testing.T) {
	bj := entities.NewDefaultBlackJack()
	bj.GetPlayer().SetChips(0)
	bj.GetDealer().SetChips(0)
	bj.Reset()
	assert.Equal(t, entities.BJDefaultChips, bj.GetPlayer().GetChips())
	assert.Equal(t, entities.BJDefaultChips, bj.GetDealer().GetChips())
}

func TestBlackJack_ResetChipsBelowMinBet(t *testing.T) {
	bj := entities.NewDefaultBlackJack()
	// チップが最低ベット額未満ならリセットされる
	bj.GetPlayer().SetChips(entities.BJMinBet - 1)
	bj.GetDealer().SetChips(entities.BJMinBet - 1)
	bj.Reset()
	assert.Equal(t, entities.BJDefaultChips, bj.GetPlayer().GetChips())
	assert.Equal(t, entities.BJDefaultChips, bj.GetDealer().GetChips())

	// ちょうど最低ベット額ならリセットされない
	bj.GetPlayer().SetChips(entities.BJMinBet)
	bj.GetDealer().SetChips(entities.BJMinBet)
	bj.Reset()
	assert.Equal(t, entities.BJMinBet, bj.GetPlayer().GetChips())
	assert.Equal(t, entities.BJMinBet, bj.GetDealer().GetChips())
}

func TestBlackJack_PlayerBet(t *testing.T) {
	t.Run("successful bet", func(t *testing.T) {
		bj := entities.NewDefaultBlackJack()
		bj.Reset()
		ok := bj.PlayerBet(100)
		assert.True(t, ok)
		assert.Equal(t, 2, bj.GetPlayerHands()[0].GetCardsSize())
		assert.Equal(t, 2, bj.GetDealer().GetCardsSize())
		assert.NotEqual(t, entities.BJPhaseBet, bj.GetPhase())
	})
	t.Run("bet below minimum", func(t *testing.T) {
		bj := entities.NewDefaultBlackJack()
		bj.Reset()
		ok := bj.PlayerBet(5)
		assert.False(t, ok)
		assert.Equal(t, entities.BJPhaseBet, bj.GetPhase())
	})
	t.Run("bet with insufficient chips", func(t *testing.T) {
		bj := entities.NewDefaultBlackJack()
		bj.Reset()
		bj.GetPlayer().SetChips(50)
		ok := bj.PlayerBet(100)
		assert.False(t, ok)
	})
	t.Run("bet not multiple of min bet", func(t *testing.T) {
		bj := entities.NewDefaultBlackJack()
		bj.Reset()
		ok := bj.PlayerBet(15)
		assert.False(t, ok)
		assert.Equal(t, entities.BJPhaseBet, bj.GetPhase())
	})
	t.Run("bet in wrong phase", func(t *testing.T) {
		bj := entities.NewDefaultBlackJack()
		bj.Reset()
		bj.PlayerBet(100)
		ok := bj.PlayerBet(100)
		assert.False(t, ok)
	})
}

func TestBlackJack_InsurancePhase(t *testing.T) {
	t.Run("accept insurance in wrong phase returns false", func(t *testing.T) {
		bj := entities.NewDefaultBlackJack()
		bj.Reset()
		ok := bj.PlayerInsurance()
		assert.False(t, ok)
	})
	t.Run("decline insurance in wrong phase does nothing", func(t *testing.T) {
		bj := entities.NewDefaultBlackJack()
		bj.Reset()
		bj.PlayerDeclineInsurance()
		assert.Equal(t, entities.BJPhaseBet, bj.GetPhase())
	})
}

func TestBlackJack_PlayerHitInWrongPhase(t *testing.T) {
	bj := entities.NewDefaultBlackJack()
	bj.Reset()
	bj.PlayerHit()
	assert.Equal(t, entities.BJPhaseBet, bj.GetPhase())
}

func TestBlackJack_PlayerStandViaHand(t *testing.T) {
	playerCards := []*entities.Card{
		entities.NewCard(entities.CardDesignSpade, 9, false),
		entities.NewCard(entities.CardDesignHeart, 8, false),
	}
	dealerCards := []*entities.Card{
		entities.NewCard(entities.CardDesignClover, 2, false),
		entities.NewCard(entities.CardDesignClover, 10, false),
		entities.NewCard(entities.CardDesignClover, 11, false),
	}
	bj, _, _ := setupDeterministicBJ(1000, playerCards, dealerCards, 100)
	bj.PlayerStand()
	// Player score 17, dealer score 22 (bust) → player wins
	assert.Equal(t, entities.GameResultWin, bj.GameJudgment())
}

func TestBlackJack_GameJudgmentCases(t *testing.T) {
	t.Run("player lose bust", func(t *testing.T) {
		playerCards := []*entities.Card{
			entities.NewCard(entities.CardDesignSpade, 2, false),
			entities.NewCard(entities.CardDesignSpade, 10, false),
			entities.NewCard(entities.CardDesignSpade, 11, false),
		}
		dealerCards := []*entities.Card{
			entities.NewCard(entities.CardDesignClover, 2, false),
			entities.NewCard(entities.CardDesignClover, 10, false),
			entities.NewCard(entities.CardDesignClover, 11, false),
		}
		bj, _, _ := setupDeterministicBJ(1000, playerCards, dealerCards, 100)
		assert.Equal(t, entities.GameResultLose, bj.GameJudgment())
	})
	t.Run("player lose dealer higher", func(t *testing.T) {
		playerCards := []*entities.Card{
			entities.NewCard(entities.CardDesignSpade, 2, false),
			entities.NewCard(entities.CardDesignSpade, 10, false),
			entities.NewCard(entities.CardDesignSpade, 11, false),
		}
		dealerCards := []*entities.Card{
			entities.NewCard(entities.CardDesignClover, 1, false),
			entities.NewCard(entities.CardDesignClover, 10, false),
			entities.NewCard(entities.CardDesignClover, 11, false),
		}
		bj, _, _ := setupDeterministicBJ(1000, playerCards, dealerCards, 100)
		assert.Equal(t, entities.GameResultLose, bj.GameJudgment())
	})
	t.Run("player win dealer bust", func(t *testing.T) {
		playerCards := []*entities.Card{
			entities.NewCard(entities.CardDesignSpade, 1, false),
			entities.NewCard(entities.CardDesignSpade, 10, false),
			entities.NewCard(entities.CardDesignSpade, 11, false),
		}
		dealerCards := []*entities.Card{
			entities.NewCard(entities.CardDesignClover, 2, false),
			entities.NewCard(entities.CardDesignClover, 10, false),
			entities.NewCard(entities.CardDesignClover, 11, false),
		}
		bj, _, _ := setupDeterministicBJ(1000, playerCards, dealerCards, 100)
		assert.Equal(t, entities.GameResultWin, bj.GameJudgment())
	})
	t.Run("draw", func(t *testing.T) {
		playerCards := []*entities.Card{
			entities.NewCard(entities.CardDesignSpade, 1, false),
			entities.NewCard(entities.CardDesignSpade, 10, false),
		}
		dealerCards := []*entities.Card{
			entities.NewCard(entities.CardDesignClover, 1, false),
			entities.NewCard(entities.CardDesignClover, 10, false),
		}
		bj, _, _ := setupDeterministicBJ(1000, playerCards, dealerCards, 100)
		assert.Equal(t, entities.GameResultDraw, bj.GameJudgment())
	})
	t.Run("player win natural BJ vs dealer 3 cards", func(t *testing.T) {
		playerCards := []*entities.Card{
			entities.NewCard(entities.CardDesignSpade, 1, false),
			entities.NewCard(entities.CardDesignSpade, 10, false),
		}
		dealerCards := []*entities.Card{
			entities.NewCard(entities.CardDesignClover, 1, false),
			entities.NewCard(entities.CardDesignClover, 10, false),
			entities.NewCard(entities.CardDesignClover, 11, false),
		}
		bj, _, _ := setupDeterministicBJ(1000, playerCards, dealerCards, 100)
		assert.Equal(t, entities.GameResultWin, bj.GameJudgment())
	})
	t.Run("player win higher score", func(t *testing.T) {
		playerCards := []*entities.Card{
			entities.NewCard(entities.CardDesignSpade, 1, false),
			entities.NewCard(entities.CardDesignSpade, 10, false),
		}
		dealerCards := []*entities.Card{
			entities.NewCard(entities.CardDesignClover, 9, false),
			entities.NewCard(entities.CardDesignClover, 10, false),
		}
		bj, _, _ := setupDeterministicBJ(1000, playerCards, dealerCards, 100)
		assert.Equal(t, entities.GameResultWin, bj.GameJudgment())
	})
	t.Run("player lose dealer higher score", func(t *testing.T) {
		playerCards := []*entities.Card{
			entities.NewCard(entities.CardDesignSpade, 9, false),
			entities.NewCard(entities.CardDesignSpade, 10, false),
		}
		dealerCards := []*entities.Card{
			entities.NewCard(entities.CardDesignClover, 1, false),
			entities.NewCard(entities.CardDesignClover, 10, false),
		}
		bj, _, _ := setupDeterministicBJ(1000, playerCards, dealerCards, 100)
		assert.Equal(t, entities.GameResultLose, bj.GameJudgment())
	})
}

func TestBlackJack_DoubleDown_WrongPhase(t *testing.T) {
	bj := entities.NewDefaultBlackJack()
	bj.Reset()
	ok := bj.PlayerDoubleDown()
	assert.False(t, ok)
}

func TestBlackJack_Split_WrongPhase(t *testing.T) {
	bj := entities.NewDefaultBlackJack()
	bj.Reset()
	ok := bj.PlayerSplit()
	assert.False(t, ok)
}

func TestBlackJack_GetterMethods(t *testing.T) {
	bj := entities.NewDefaultBlackJack()
	bj.Reset()
	assert.NotNil(t, bj.GetPlayer())
	assert.NotNil(t, bj.GetDealer())
	assert.NotNil(t, bj.GetPlayerHands())
	assert.NotNil(t, bj.GetTrumpCards())
	assert.Equal(t, 0, bj.GetCurrentHandIdx())
	assert.Equal(t, 0, bj.GetInsuranceBet())
	assert.False(t, bj.IsInsuranceAvailable())
	assert.Equal(t, entities.BJPhaseBet, bj.GetPhase())
}

func TestBlackJack_GameJudgmentForHand(t *testing.T) {
	bj := entities.NewDefaultBlackJack()
	t.Run("invalid hand index negative", func(t *testing.T) {
		assert.Equal(t, entities.GameResultLose, bj.GameJudgmentForHand(-1))
	})
	t.Run("invalid hand index out of range", func(t *testing.T) {
		assert.Equal(t, entities.GameResultLose, bj.GameJudgmentForHand(100))
	})
}

func TestBlackJack_DrawCardNilSafety(t *testing.T) {
	t.Run("bet with exhausted deck refunds chips", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		player := entities.NewBlackJackPlayer()
		dealer := entities.NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := entities.NewBlackJack(tc, player, dealer)
		bj.Reset()

		// デッキを全て引き切る
		for i := 0; i < 52; i++ {
			tc.DrawCard()
		}

		// デッキ枯渇時のベットはfalseを返し、チップが返却される
		ok := bj.PlayerBet(100)
		assert.False(t, ok)
		assert.Equal(t, 1000, player.GetChips())
		assert.Equal(t, entities.BJPhaseBet, bj.GetPhase())
	})

	t.Run("hit with exhausted deck does not panic", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		player := entities.NewBlackJackPlayer()
		dealer := entities.NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := entities.NewBlackJack(tc, player, dealer)
		bj.Reset()

		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		hand.AddCard(entities.NewCard(entities.CardDesignHeart, 6, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 10, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignDiamond, 7, false))
		bj.SetPhase(entities.BJPhaseAction)

		// デッキを全て引き切る
		for i := 0; i < 52; i++ {
			tc.DrawCard()
		}

		// デッキ枯渇後のヒットでパニックしない
		assert.NotPanics(t, func() {
			bj.PlayerHit()
		})
		// カードが追加されないことを確認
		assert.Equal(t, 2, hand.GetCardsSize())
	})

	t.Run("dealer hit with exhausted deck does not panic", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		player := entities.NewBlackJackPlayer()
		dealer := entities.NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := entities.NewBlackJack(tc, player, dealer)
		bj.Reset()

		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(entities.NewCard(entities.CardDesignSpade, 10, false))
		hand.AddCard(entities.NewCard(entities.CardDesignHeart, 10, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignDiamond, 6, false))
		bj.SetPhase(entities.BJPhaseAction)

		// デッキを全て引き切る
		for i := 0; i < 52; i++ {
			tc.DrawCard()
		}

		// デッキ枯渇後のスタンド→ディーラーヒットでパニックしない
		assert.NotPanics(t, func() {
			bj.PlayerStand()
		})
		assert.True(t, bj.GetGameEndFlag())
	})
}

func TestBlackJack_LastError(t *testing.T) {
	t.Run("bet failure sets lastError", func(t *testing.T) {
		bj := entities.NewDefaultBlackJack()
		bj.Reset()
		ok := bj.PlayerBet(5) // below minimum
		assert.False(t, ok)
		assert.Equal(t, "Invalid bet amount.", bj.GetLastError())
		// GetLastError clears the message
		assert.Equal(t, "", bj.GetLastError())
	})
	t.Run("bet wrong phase sets lastError", func(t *testing.T) {
		bj := entities.NewDefaultBlackJack()
		bj.Reset()
		bj.PlayerBet(100)
		ok := bj.PlayerBet(100)
		assert.False(t, ok)
		errMsg := bj.GetLastError()
		assert.Equal(t, "Bet is only allowed during the bet phase.", errMsg)
	})
	t.Run("bet insufficient chips sets lastError", func(t *testing.T) {
		bj := entities.NewDefaultBlackJack()
		bj.Reset()
		bj.GetPlayer().SetChips(50)
		ok := bj.PlayerBet(100)
		assert.False(t, ok)
		assert.Equal(t, "Insufficient chips.", bj.GetLastError())
	})
	t.Run("successful bet clears lastError", func(t *testing.T) {
		bj := entities.NewDefaultBlackJack()
		bj.Reset()
		ok := bj.PlayerBet(100)
		assert.True(t, ok)
		assert.Equal(t, "", bj.GetLastError())
	})
	t.Run("double down failure sets lastError", func(t *testing.T) {
		bj := entities.NewDefaultBlackJack()
		bj.Reset()
		ok := bj.PlayerDoubleDown()
		assert.False(t, ok)
		assert.Equal(t, "Double down is not allowed now.", bj.GetLastError())
	})
	t.Run("split failure sets lastError", func(t *testing.T) {
		bj := entities.NewDefaultBlackJack()
		bj.Reset()
		ok := bj.PlayerSplit()
		assert.False(t, ok)
		assert.Equal(t, "Split is not allowed now.", bj.GetLastError())
	})
	t.Run("insurance failure sets lastError", func(t *testing.T) {
		bj := entities.NewDefaultBlackJack()
		bj.Reset()
		ok := bj.PlayerInsurance()
		assert.False(t, ok)
		assert.Equal(t, "Insurance is not available now.", bj.GetLastError())
	})
}

func TestBlackJack_MaxSplitHandsLimit(t *testing.T) {
	// Test that split is rejected when playerHands already has BJMaxHands entries.
	// We simulate this by manually adding dummy hands to the slice.
	tc := entities.NewTrumpCards(0)
	player := entities.NewBlackJackPlayer()
	dealer := entities.NewBlackJackPlayer()
	player.SetChips(10000)
	dealer.SetChips(10000)
	bj := entities.NewBlackJack(tc, player, dealer)
	bj.Reset()

	hand := bj.GetPlayerHands()[0]
	hand.SetBet(100)
	hand.AddCard(entities.NewCard(entities.CardDesignSpade, 8, false))
	hand.AddCard(entities.NewCard(entities.CardDesignHeart, 8, false))
	dealer.AddCard(entities.NewCard(entities.CardDesignClover, 10, false))
	dealer.AddCard(entities.NewCard(entities.CardDesignDiamond, 7, false))
	bj.SetPhase(entities.BJPhaseAction)

	// First split should succeed (1 → 2 hands)
	ok := bj.PlayerSplit()
	assert.True(t, ok)
	assert.Equal(t, 2, len(bj.GetPlayerHands()))

	// Manually add dummy hands to reach BJMaxHands
	for len(bj.GetPlayerHands()) < entities.BJMaxHands {
		dummyHand := entities.NewBlackJackHand()
		dummyHand.SetBet(100)
		dummyHand.SetStood(true)
		bj.SetPlayerHands(append(bj.GetPlayerHands(), dummyHand))
	}
	assert.Equal(t, entities.BJMaxHands, len(bj.GetPlayerHands()))

	// Now split should fail with max hands error
	ok = bj.PlayerSplit()
	assert.False(t, ok)
	assert.Equal(t, "Maximum number of hands reached.", bj.GetLastError())
}

func TestBlackJack_OldFlow(t *testing.T) {
	tc := entities.NewTrumpCards(0)
	player := entities.NewBlackJackPlayer()
	dealer := entities.NewBlackJackPlayer()
	tb := entities.NewBlackJack(tc, player, dealer)
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
		tb.PlayerHit()
	})
	t.Run("success PlayerStand", func(t *testing.T) {
		tb.PlayerStand()
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
	playerCards []*entities.Card,
	dealerCards []*entities.Card,
	bet int,
) (*entities.BlackJack, *entities.BlackJackPlayer, *entities.BlackJackPlayer) {
	tc := entities.NewTrumpCards(0)
	player := entities.NewBlackJackPlayer()
	dealer := entities.NewBlackJackPlayer()
	player.SetChips(playerChips)
	dealer.SetChips(1000)
	bj := entities.NewBlackJack(tc, player, dealer)
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
	bj.SetPhase(entities.BJPhaseAction)

	return bj, player, dealer
}

func TestBlackJack_FullBettingFlow(t *testing.T) {
	t.Run("normal win payout", func(t *testing.T) {
		bj, player, _ := setupDeterministicBJ(
			900, // after 100 bet
			[]*entities.Card{
				entities.NewCard(entities.CardDesignSpade, 10, false),
				entities.NewCard(entities.CardDesignHeart, 10, false),
			},
			[]*entities.Card{
				entities.NewCard(entities.CardDesignClover, 9, false),
				entities.NewCard(entities.CardDesignDiamond, 10, false),
			},
			100,
		)
		bj.PlayerStand()
		assert.True(t, bj.GetGameEndFlag())
		assert.Equal(t, entities.BJPhaseEnd, bj.GetPhase())
		assert.Equal(t, entities.GameResultWin, bj.GameJudgment())
		// Normal win: bet*2 = 200 returned, 900 + 200 = 1100
		assert.Equal(t, 1100, player.GetChips())
	})

	t.Run("natural BJ 3:2 payout", func(t *testing.T) {
		bj, player, _ := setupDeterministicBJ(
			900,
			[]*entities.Card{
				entities.NewCard(entities.CardDesignSpade, 1, false),
				entities.NewCard(entities.CardDesignHeart, 10, false),
			},
			[]*entities.Card{
				entities.NewCard(entities.CardDesignClover, 9, false),
				entities.NewCard(entities.CardDesignDiamond, 10, false),
			},
			100,
		)
		bj.PlayerStand()
		assert.True(t, bj.GetGameEndFlag())
		assert.Equal(t, entities.GameResultWin, bj.GameJudgment())
		// Natural BJ 3:2: 100 + 150 = 250 returned, 900 + 250 = 1150
		assert.Equal(t, 1150, player.GetChips())
	})

	t.Run("draw payout", func(t *testing.T) {
		bj, player, _ := setupDeterministicBJ(
			900,
			[]*entities.Card{
				entities.NewCard(entities.CardDesignSpade, 10, false),
				entities.NewCard(entities.CardDesignHeart, 9, false),
			},
			[]*entities.Card{
				entities.NewCard(entities.CardDesignClover, 9, false),
				entities.NewCard(entities.CardDesignDiamond, 10, false),
			},
			100,
		)
		bj.PlayerStand()
		assert.True(t, bj.GetGameEndFlag())
		assert.Equal(t, entities.GameResultDraw, bj.GameJudgment())
		// Draw: bet returned, 900 + 100 = 1000
		assert.Equal(t, 1000, player.GetChips())
	})

	t.Run("lose payout", func(t *testing.T) {
		bj, player, _ := setupDeterministicBJ(
			900,
			[]*entities.Card{
				entities.NewCard(entities.CardDesignSpade, 8, false),
				entities.NewCard(entities.CardDesignHeart, 9, false),
			},
			[]*entities.Card{
				entities.NewCard(entities.CardDesignClover, 10, false),
				entities.NewCard(entities.CardDesignDiamond, 10, false),
			},
			100,
		)
		bj.PlayerStand()
		assert.True(t, bj.GetGameEndFlag())
		assert.Equal(t, entities.GameResultLose, bj.GameJudgment())
		// Lose: nothing returned, chips stay at 900
		assert.Equal(t, 900, player.GetChips())
	})

	t.Run("double down", func(t *testing.T) {
		bj, _, _ := setupDeterministicBJ(
			900,
			[]*entities.Card{
				entities.NewCard(entities.CardDesignSpade, 5, false),
				entities.NewCard(entities.CardDesignHeart, 6, false),
			},
			[]*entities.Card{
				entities.NewCard(entities.CardDesignClover, 10, false),
				entities.NewCard(entities.CardDesignDiamond, 7, false),
			},
			100,
		)
		ok := bj.PlayerDoubleDown()
		assert.True(t, ok)
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
			[]*entities.Card{
				entities.NewCard(entities.CardDesignSpade, 5, false),
				entities.NewCard(entities.CardDesignHeart, 6, false),
			},
			[]*entities.Card{
				entities.NewCard(entities.CardDesignClover, 10, false),
				entities.NewCard(entities.CardDesignDiamond, 7, false),
			},
			100,
		)
		ok := bj.PlayerDoubleDown()
		assert.False(t, ok)
	})

	t.Run("double down with 3 cards", func(t *testing.T) {
		bj, _, _ := setupDeterministicBJ(
			900,
			[]*entities.Card{
				entities.NewCard(entities.CardDesignSpade, 2, false),
				entities.NewCard(entities.CardDesignHeart, 3, false),
				entities.NewCard(entities.CardDesignClover, 4, false),
			},
			[]*entities.Card{
				entities.NewCard(entities.CardDesignDiamond, 10, false),
				entities.NewCard(entities.CardDesignDiamond, 7, false),
			},
			100,
		)
		ok := bj.PlayerDoubleDown()
		assert.False(t, ok)
	})

	t.Run("double down on finished hand", func(t *testing.T) {
		bj, _, _ := setupDeterministicBJ(
			900,
			[]*entities.Card{
				entities.NewCard(entities.CardDesignSpade, 5, false),
				entities.NewCard(entities.CardDesignHeart, 6, false),
			},
			[]*entities.Card{
				entities.NewCard(entities.CardDesignClover, 10, false),
				entities.NewCard(entities.CardDesignDiamond, 7, false),
			},
			100,
		)
		bj.GetPlayerHands()[0].SetStood(true)
		ok := bj.PlayerDoubleDown()
		assert.False(t, ok)
	})

	t.Run("split pair", func(t *testing.T) {
		bj, player, _ := setupDeterministicBJ(
			900,
			[]*entities.Card{
				entities.NewCard(entities.CardDesignSpade, 8, false),
				entities.NewCard(entities.CardDesignHeart, 8, false),
			},
			[]*entities.Card{
				entities.NewCard(entities.CardDesignClover, 10, false),
				entities.NewCard(entities.CardDesignDiamond, 7, false),
			},
			100,
		)
		ok := bj.PlayerSplit()
		assert.True(t, ok)
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
			[]*entities.Card{
				entities.NewCard(entities.CardDesignSpade, 1, false),
				entities.NewCard(entities.CardDesignHeart, 1, false),
			},
			[]*entities.Card{
				entities.NewCard(entities.CardDesignClover, 10, false),
				entities.NewCard(entities.CardDesignDiamond, 7, false),
			},
			100,
		)
		ok := bj.PlayerSplit()
		assert.True(t, ok)
		assert.Equal(t, 2, len(bj.GetPlayerHands()))
		assert.True(t, bj.GetPlayerHands()[0].IsStood())
		assert.True(t, bj.GetPlayerHands()[1].IsStood())
	})

	t.Run("split insufficient chips", func(t *testing.T) {
		bj, _, _ := setupDeterministicBJ(
			50,
			[]*entities.Card{
				entities.NewCard(entities.CardDesignSpade, 8, false),
				entities.NewCard(entities.CardDesignHeart, 8, false),
			},
			[]*entities.Card{
				entities.NewCard(entities.CardDesignClover, 10, false),
				entities.NewCard(entities.CardDesignDiamond, 7, false),
			},
			100,
		)
		ok := bj.PlayerSplit()
		assert.False(t, ok)
	})

	t.Run("split non-pair", func(t *testing.T) {
		bj, _, _ := setupDeterministicBJ(
			900,
			[]*entities.Card{
				entities.NewCard(entities.CardDesignSpade, 5, false),
				entities.NewCard(entities.CardDesignHeart, 8, false),
			},
			[]*entities.Card{
				entities.NewCard(entities.CardDesignClover, 10, false),
				entities.NewCard(entities.CardDesignDiamond, 7, false),
			},
			100,
		)
		ok := bj.PlayerSplit()
		assert.False(t, ok)
	})

	t.Run("all busted skips dealer draw", func(t *testing.T) {
		bj, _, dealer := setupDeterministicBJ(
			900,
			[]*entities.Card{
				entities.NewCard(entities.CardDesignSpade, 10, false),
				entities.NewCard(entities.CardDesignHeart, 10, false),
				entities.NewCard(entities.CardDesignClover, 5, false),
			},
			[]*entities.Card{
				entities.NewCard(entities.CardDesignDiamond, 5, false),
				entities.NewCard(entities.CardDesignDiamond, 6, false),
			},
			100,
		)
		bj.GetPlayerHands()[0].SetBusted(true)
		bj.PlayerStand()
		assert.True(t, bj.GetGameEndFlag())
		assert.Equal(t, 2, dealer.GetCardsSize())
	})

	t.Run("hit then stand flow", func(t *testing.T) {
		bj, _, _ := setupDeterministicBJ(
			900,
			[]*entities.Card{
				entities.NewCard(entities.CardDesignSpade, 5, false),
				entities.NewCard(entities.CardDesignHeart, 6, false),
			},
			[]*entities.Card{
				entities.NewCard(entities.CardDesignClover, 10, false),
				entities.NewCard(entities.CardDesignDiamond, 7, false),
			},
			100,
		)
		bj.PlayerHit()
		bj.PlayerStand()
		assert.True(t, bj.GetGameEndFlag())
		assert.Equal(t, entities.BJPhaseEnd, bj.GetPhase())
	})

	t.Run("insurance win with dealer BJ", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		player := entities.NewBlackJackPlayer()
		dealer := entities.NewBlackJackPlayer()
		player.SetChips(850) // 1000 - 100 bet - 50 insurance
		dealer.SetChips(1000)
		bj := entities.NewBlackJack(tc, player, dealer)
		bj.Reset()
		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(entities.NewCard(entities.CardDesignSpade, 9, false))
		hand.AddCard(entities.NewCard(entities.CardDesignHeart, 8, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 9, false))
		player.AddCard(entities.NewCard(entities.CardDesignHeart, 8, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 1, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignDiamond, 10, false))
		bj.SetPhase(entities.BJPhaseInsurance)
		bj.PlayerInsurance() // cost = 50, chips: 800
		assert.Equal(t, entities.BJPhaseEnd, bj.GetPhase())
		// Dealer has BJ, so insurance pays 3*50=150. Player loses hand (17 vs 21).
		// chips: 800 + 150 (insurance win) + 0 (hand loss) = 950
		assert.Equal(t, 950, player.GetChips())
	})

	t.Run("insurance lose without dealer BJ", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		player := entities.NewBlackJackPlayer()
		dealer := entities.NewBlackJackPlayer()
		player.SetChips(850)
		dealer.SetChips(1000)
		bj := entities.NewBlackJack(tc, player, dealer)
		bj.Reset()
		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(entities.NewCard(entities.CardDesignSpade, 10, false))
		hand.AddCard(entities.NewCard(entities.CardDesignHeart, 10, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 10, false))
		player.AddCard(entities.NewCard(entities.CardDesignHeart, 10, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 1, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignDiamond, 7, false))
		bj.SetPhase(entities.BJPhaseInsurance)
		bj.PlayerInsurance() // cost = 50, chips: 800
		assert.Equal(t, entities.BJPhaseAction, bj.GetPhase())
		bj.PlayerStand()
		assert.True(t, bj.GetGameEndFlag())
		// Dealer: A + 7 = 18, Player: 20, player wins
		// Insurance lost (no dealer BJ), hand wins: 800 + 0 (insurance) + 200 (hand win) = 1000
		assert.Equal(t, 1000, player.GetChips())
	})

	t.Run("decline insurance", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		player := entities.NewBlackJackPlayer()
		dealer := entities.NewBlackJackPlayer()
		player.SetChips(900)
		dealer.SetChips(1000)
		bj := entities.NewBlackJack(tc, player, dealer)
		bj.Reset()
		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(entities.NewCard(entities.CardDesignSpade, 10, false))
		hand.AddCard(entities.NewCard(entities.CardDesignHeart, 10, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 10, false))
		player.AddCard(entities.NewCard(entities.CardDesignHeart, 10, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 1, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignDiamond, 7, false))
		bj.SetPhase(entities.BJPhaseInsurance)
		bj.PlayerDeclineInsurance()
		assert.Equal(t, entities.BJPhaseAction, bj.GetPhase())
		assert.Equal(t, 0, bj.GetInsuranceBet())
	})

	t.Run("dealer BJ vs player BJ is draw", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		player := entities.NewBlackJackPlayer()
		dealer := entities.NewBlackJackPlayer()
		player.SetChips(900)
		dealer.SetChips(1000)
		bj := entities.NewBlackJack(tc, player, dealer)
		bj.Reset()
		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(entities.NewCard(entities.CardDesignSpade, 1, false))
		hand.AddCard(entities.NewCard(entities.CardDesignHeart, 10, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 1, false))
		player.AddCard(entities.NewCard(entities.CardDesignHeart, 10, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 1, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignDiamond, 13, false))
		bj.SetPhase(entities.BJPhaseAction)
		bj.PlayerStand()
		assert.True(t, bj.GetGameEndFlag())
		assert.Equal(t, entities.GameResultDraw, bj.GameJudgment())
		// Draw: bet returned, 900 + 100 = 1000
		assert.Equal(t, 1000, player.GetChips())
	})

	t.Run("SetPhase sets gameEndFlag correctly", func(t *testing.T) {
		bj := entities.NewDefaultBlackJack()
		bj.SetPhase(entities.BJPhaseEnd)
		assert.True(t, bj.GetGameEndFlag())
		bj.SetPhase(entities.BJPhaseAction)
		assert.False(t, bj.GetGameEndFlag())
	})
}
