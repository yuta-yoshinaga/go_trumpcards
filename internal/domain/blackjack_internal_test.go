package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupInternalTestBJ(playerChips, dealerChips int) (*BlackJack, *BlackJackPlayer, *BlackJackPlayer) {
	tc := NewTrumpCards(0)
	player := NewBlackJackPlayer()
	dealer := NewBlackJackPlayer()
	player.SetChips(playerChips)
	dealer.SetChips(dealerChips)
	bj := NewBlackJack(tc, player, dealer)
	return bj, player, dealer
}

// TestBlackJack_PlayerDoubleDown_Bust_Deterministic exercises lines 234-236 of
// BlackJack.go: when score >= 22 after the double-down draw, the hand should be
// marked as busted. Instead of retrying random shuffles, the deck is directly
// stacked so the next drawn card is guaranteed to be a King (BJ value 10),
// pushing the player's score from 20 to 30 (bust).
func TestBlackJack_PlayerDoubleDown_Bust_Deterministic(t *testing.T) {
	bj, _, dealer := setupInternalTestBJ(1000, 1000)

	// Set up player hand with score 20 (10 + King=10)
	hand := bj.playerHands[0]
	hand.SetBet(100)
	hand.AddCard(NewCard(CardDesignSpade, 10, false))
	hand.AddCard(NewCard(CardDesignHeart, 13, false)) // King = BJ value 10, score = 20

	// Set up dealer cards (score 17, so dealer won't draw if reached)
	dealer.AddCard(NewCard(CardDesignClover, 10, false))
	dealer.AddCard(NewCard(CardDesignDiamond, 7, false))

	// Stack the deck: place a King at the current draw position so that
	// drawCard() returns a card with BJ value 10, causing 20 + 10 = 30 (bust).
	bj.trumpCards.deck[0] = NewCard(CardDesignClover, 13, false) // King = BJ value 10
	bj.trumpCards.deckDrawCnt = 0

	// Set phase to action so PlayerDoubleDown is allowed
	bj.phase = BJPhaseAction

	err := bj.PlayerDoubleDown()
	assert.NoError(t, err)

	// Verify double-down mechanics
	assert.Equal(t, 200, hand.GetBet(), "bet should be doubled")
	assert.True(t, hand.IsDoubled(), "hand should be marked as doubled")
	assert.Equal(t, 3, hand.GetCardsSize(), "hand should have 3 cards after DD")

	// Verify bust-specific state (lines 234-236)
	assert.True(t, hand.IsBusted(), "hand should be busted when score >= 22")
	assert.False(t, hand.IsStood(), "busted hand should not be stood")
	assert.True(t, hand.GetScore() >= 22, "score should be >= 22 for bust")

	// Game should end (single hand, all finished → dealerPlay → allBusted → endGame)
	assert.True(t, bj.gameEndFlag)
	assert.Equal(t, BJPhaseEnd, bj.phase)

	// Player should lose (busted)
	assert.Equal(t, GameResultLose, bj.GameJudgment())
}

// --- hiLoValue tests ---

func TestHiLoValue(t *testing.T) {
	t.Run("nil card returns 0", func(t *testing.T) {
		assert.Equal(t, 0, hiLoValue(nil))
	})

	// Hi-Lo: 2-6 = +1, 7-9 = 0, 10/J/Q/K/A = -1
	tests := []struct {
		value    int
		expected int
		label    string
	}{
		{1, -1, "Ace"},
		{2, 1, "2"},
		{3, 1, "3"},
		{4, 1, "4"},
		{5, 1, "5"},
		{6, 1, "6"},
		{7, 0, "7"},
		{8, 0, "8"},
		{9, 0, "9"},
		{10, -1, "10"},
		{11, -1, "J"},
		{12, -1, "Q"},
		{13, -1, "K"},
	}

	for _, tc := range tests {
		t.Run(tc.label, func(t *testing.T) {
			card := NewCard(CardDesignSpade, tc.value, false)
			assert.Equal(t, tc.expected, hiLoValue(card), "hiLoValue for %s (value=%d)", tc.label, tc.value)
		})
	}
}

// --- dealerShouldHit tests ---

func TestDealerShouldHit(t *testing.T) {
	t.Run("score below 17 should hit", func(t *testing.T) {
		bj, _, dealer := setupInternalTestBJ(1000, 1000)
		dealer.AddCard(NewCard(CardDesignSpade, 10, false))
		dealer.AddCard(NewCard(CardDesignHeart, 5, false)) // score 15
		assert.True(t, bj.dealerShouldHit())
	})

	t.Run("hard 17 should not hit (S17)", func(t *testing.T) {
		bj, _, dealer := setupInternalTestBJ(1000, 1000)
		dealer.AddCard(NewCard(CardDesignSpade, 10, false))
		dealer.AddCard(NewCard(CardDesignHeart, 7, false)) // score 17, hard
		assert.False(t, bj.dealerShouldHit())
	})

	t.Run("soft 17 should not hit with S17", func(t *testing.T) {
		bj, _, dealer := setupInternalTestBJ(1000, 1000)
		dealer.AddCard(NewCard(CardDesignSpade, 1, false))
		dealer.AddCard(NewCard(CardDesignHeart, 6, false)) // soft 17
		bj.config.DealerHitsSoft17 = false
		assert.False(t, bj.dealerShouldHit())
	})

	t.Run("soft 17 should hit with H17", func(t *testing.T) {
		bj, _, dealer := setupInternalTestBJ(1000, 1000)
		dealer.AddCard(NewCard(CardDesignSpade, 1, false))
		dealer.AddCard(NewCard(CardDesignHeart, 6, false)) // soft 17
		bj.config.DealerHitsSoft17 = true
		assert.True(t, bj.dealerShouldHit())
	})

	t.Run("score above 17 should not hit", func(t *testing.T) {
		bj, _, dealer := setupInternalTestBJ(1000, 1000)
		dealer.AddCard(NewCard(CardDesignSpade, 10, false))
		dealer.AddCard(NewCard(CardDesignHeart, 8, false)) // score 18
		assert.False(t, bj.dealerShouldHit())
	})

	t.Run("soft 18 should not hit even with H17", func(t *testing.T) {
		bj, _, dealer := setupInternalTestBJ(1000, 1000)
		dealer.AddCard(NewCard(CardDesignSpade, 1, false))
		dealer.AddCard(NewCard(CardDesignHeart, 7, false)) // soft 18
		bj.config.DealerHitsSoft17 = true
		assert.False(t, bj.dealerShouldHit(), "soft 18 should stand even with H17")
	})
}

// --- allPlayerHandsDone tests ---

func TestAllPlayerHandsDone(t *testing.T) {
	t.Run("not all done when one hand is active", func(t *testing.T) {
		bj, _, _ := setupInternalTestBJ(1000, 1000)
		hand := bj.playerHands[0]
		hand.AddCard(NewCard(CardDesignSpade, 10, false))
		hand.AddCard(NewCard(CardDesignHeart, 8, false))
		assert.False(t, bj.allPlayerHandsDone())
	})

	t.Run("all done when all hands busted", func(t *testing.T) {
		bj, _, _ := setupInternalTestBJ(1000, 1000)
		hand := bj.playerHands[0]
		hand.SetBusted(true)
		assert.True(t, bj.allPlayerHandsDone())
	})

	t.Run("all done when all hands surrendered", func(t *testing.T) {
		bj, _, _ := setupInternalTestBJ(1000, 1000)
		hand := bj.playerHands[0]
		hand.SetSurrendered(true)
		assert.True(t, bj.allPlayerHandsDone())
	})

	t.Run("CPU hand not busted prevents all done", func(t *testing.T) {
		bj, _, _ := setupInternalTestBJ(1000, 1000)
		hand := bj.playerHands[0]
		hand.SetBusted(true)

		cpu := NewBlackJackCpuSeat()
		cpuHand := cpu.GetHands()[0]
		cpuHand.AddCard(NewCard(CardDesignSpade, 10, false))
		cpuHand.AddCard(NewCard(CardDesignHeart, 8, false))
		bj.cpuPlayers = []*BlackJackCpuSeat{cpu}

		assert.False(t, bj.allPlayerHandsDone(), "CPU active hand should prevent all-done")
	})

	t.Run("CPU hand busted allows all done", func(t *testing.T) {
		bj, _, _ := setupInternalTestBJ(1000, 1000)
		hand := bj.playerHands[0]
		hand.SetBusted(true)

		cpu := NewBlackJackCpuSeat()
		cpuHand := cpu.GetHands()[0]
		cpuHand.SetBusted(true)
		bj.cpuPlayers = []*BlackJackCpuSeat{cpu}

		assert.True(t, bj.allPlayerHandsDone())
	})
}

// --- updateRunningCount tests ---

func TestUpdateRunningCount(t *testing.T) {
	t.Run("counting disabled does nothing", func(t *testing.T) {
		bj, _, _ := setupInternalTestBJ(1000, 1000)
		bj.config.CountingEnabled = false
		bj.runningCount = 0
		bj.updateRunningCount(NewCard(CardDesignSpade, 2, false))
		assert.Equal(t, 0, bj.runningCount, "should not update when counting disabled")
	})

	t.Run("counting enabled updates RC", func(t *testing.T) {
		bj, _, _ := setupInternalTestBJ(1000, 1000)
		bj.config.CountingEnabled = true
		bj.runningCount = 0

		// Low card (2-6) should add +1
		bj.updateRunningCount(NewCard(CardDesignSpade, 3, false))
		assert.Equal(t, 1, bj.runningCount)

		// Neutral card (7-9) should add 0
		bj.updateRunningCount(NewCard(CardDesignSpade, 8, false))
		assert.Equal(t, 1, bj.runningCount)

		// High card (10/J/Q/K/A) should add -1
		bj.updateRunningCount(NewCard(CardDesignSpade, 10, false))
		assert.Equal(t, 0, bj.runningCount)

		// Ace should add -1
		bj.updateRunningCount(NewCard(CardDesignSpade, 1, false))
		assert.Equal(t, -1, bj.runningCount)
	})
}

// --- GetTrueCount internal tests ---

func TestGetTrueCountInternal(t *testing.T) {
	t.Run("nil trumpCards returns 0", func(t *testing.T) {
		bj := &BlackJack{}
		bj.trumpCards = nil
		assert.Equal(t, 0.0, bj.GetTrueCount())
	})

	t.Run("less than 52 remaining uses 1.0 as floor", func(t *testing.T) {
		bj, _, _ := setupInternalTestBJ(1000, 1000)
		bj.runningCount = 10
		// Draw most of the cards to get remaining < 52
		for i := 0; i < 42; i++ {
			bj.trumpCards.DrawCard()
		}
		remaining := bj.trumpCards.GetRemainingCount()
		assert.Less(t, remaining, 52)
		tc := bj.GetTrueCount()
		assert.Equal(t, float64(bj.runningCount)/1.0, tc)
	})

	t.Run("multi-deck correct calculation", func(t *testing.T) {
		tc := NewTrumpCardsWithDecks(4, 0)
		player := NewBlackJackPlayer()
		dealer := NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := NewBlackJack(tc, player, dealer)
		bj.runningCount = 8
		// 4 decks = 208 cards, decks remaining = 208/52 = 4.0
		expected := float64(8) / 4.0
		assert.InDelta(t, expected, bj.GetTrueCount(), 0.01)
	})
}

// --- judgeHandCore tests ---

func TestJudgeHandCore(t *testing.T) {
	t.Run("player bust loses", func(t *testing.T) {
		bj, _, dealer := setupInternalTestBJ(1000, 1000)
		dealer.AddCard(NewCard(CardDesignSpade, 10, false))
		dealer.AddCard(NewCard(CardDesignHeart, 7, false)) // 17

		hand := NewBlackJackHand()
		hand.AddCard(NewCard(CardDesignClover, 10, false))
		hand.AddCard(NewCard(CardDesignDiamond, 10, false))
		hand.AddCard(NewCard(CardDesignSpade, 5, false)) // 25, bust

		assert.Equal(t, GameResultLose, bj.judgeHandCore(hand, 1))
	})

	t.Run("dealer bust, player wins", func(t *testing.T) {
		bj, _, dealer := setupInternalTestBJ(1000, 1000)
		dealer.AddCard(NewCard(CardDesignSpade, 10, false))
		dealer.AddCard(NewCard(CardDesignHeart, 10, false))
		dealer.AddCard(NewCard(CardDesignClover, 5, false)) // 25, bust

		hand := NewBlackJackHand()
		hand.AddCard(NewCard(CardDesignDiamond, 10, false))
		hand.AddCard(NewCard(CardDesignSpade, 8, false)) // 18

		assert.Equal(t, GameResultWin, bj.judgeHandCore(hand, 1))
	})

	t.Run("player higher score wins", func(t *testing.T) {
		bj, _, dealer := setupInternalTestBJ(1000, 1000)
		dealer.AddCard(NewCard(CardDesignSpade, 10, false))
		dealer.AddCard(NewCard(CardDesignHeart, 7, false)) // 17

		hand := NewBlackJackHand()
		hand.AddCard(NewCard(CardDesignClover, 10, false))
		hand.AddCard(NewCard(CardDesignDiamond, 9, false)) // 19

		assert.Equal(t, GameResultWin, bj.judgeHandCore(hand, 1))
	})

	t.Run("dealer higher score, player loses", func(t *testing.T) {
		bj, _, dealer := setupInternalTestBJ(1000, 1000)
		dealer.AddCard(NewCard(CardDesignSpade, 10, false))
		dealer.AddCard(NewCard(CardDesignHeart, 9, false)) // 19

		hand := NewBlackJackHand()
		hand.AddCard(NewCard(CardDesignClover, 10, false))
		hand.AddCard(NewCard(CardDesignDiamond, 7, false)) // 17

		assert.Equal(t, GameResultLose, bj.judgeHandCore(hand, 1))
	})

	t.Run("equal score, neither BJ, draw", func(t *testing.T) {
		bj, _, dealer := setupInternalTestBJ(1000, 1000)
		dealer.AddCard(NewCard(CardDesignSpade, 10, false))
		dealer.AddCard(NewCard(CardDesignHeart, 8, false)) // 18

		hand := NewBlackJackHand()
		hand.AddCard(NewCard(CardDesignClover, 10, false))
		hand.AddCard(NewCard(CardDesignDiamond, 8, false)) // 18

		assert.Equal(t, GameResultDraw, bj.judgeHandCore(hand, 1))
	})

	t.Run("both BJ draw", func(t *testing.T) {
		bj, _, dealer := setupInternalTestBJ(1000, 1000)
		dealer.AddCard(NewCard(CardDesignSpade, 1, false))
		dealer.AddCard(NewCard(CardDesignHeart, 13, false)) // BJ

		hand := NewBlackJackHand()
		hand.AddCard(NewCard(CardDesignClover, 1, false))
		hand.AddCard(NewCard(CardDesignDiamond, 10, false)) // BJ

		assert.Equal(t, GameResultDraw, bj.judgeHandCore(hand, 1))
	})

	t.Run("player BJ beats dealer non-BJ at 21 (handCount=1)", func(t *testing.T) {
		bj, _, dealer := setupInternalTestBJ(1000, 1000)
		dealer.AddCard(NewCard(CardDesignSpade, 7, false))
		dealer.AddCard(NewCard(CardDesignHeart, 7, false))
		dealer.AddCard(NewCard(CardDesignClover, 7, false)) // 21, but 3 cards

		hand := NewBlackJackHand()
		hand.AddCard(NewCard(CardDesignDiamond, 1, false))
		hand.AddCard(NewCard(CardDesignSpade, 13, false)) // BJ

		assert.Equal(t, GameResultWin, bj.judgeHandCore(hand, 1))
	})

	t.Run("dealer BJ beats player non-BJ at 21", func(t *testing.T) {
		bj, _, dealer := setupInternalTestBJ(1000, 1000)
		dealer.AddCard(NewCard(CardDesignSpade, 1, false))
		dealer.AddCard(NewCard(CardDesignHeart, 13, false)) // BJ

		hand := NewBlackJackHand()
		hand.AddCard(NewCard(CardDesignClover, 7, false))
		hand.AddCard(NewCard(CardDesignDiamond, 7, false))
		hand.AddCard(NewCard(CardDesignSpade, 7, false)) // 21, but 3 cards

		assert.Equal(t, GameResultLose, bj.judgeHandCore(hand, 1))
	})

	t.Run("handCount>1 suppresses player BJ, dealer not BJ, draw", func(t *testing.T) {
		bj, _, dealer := setupInternalTestBJ(1000, 1000)
		dealer.AddCard(NewCard(CardDesignSpade, 7, false))
		dealer.AddCard(NewCard(CardDesignHeart, 7, false))
		dealer.AddCard(NewCard(CardDesignClover, 7, false)) // 21, 3 cards

		hand := NewBlackJackHand()
		hand.AddCard(NewCard(CardDesignDiamond, 1, false))
		hand.AddCard(NewCard(CardDesignSpade, 13, false)) // 21, 2 cards but handCount=2

		assert.Equal(t, GameResultDraw, bj.judgeHandCore(hand, 2))
	})

	t.Run("handCount>1 suppresses player BJ, dealer BJ, lose", func(t *testing.T) {
		bj, _, dealer := setupInternalTestBJ(1000, 1000)
		dealer.AddCard(NewCard(CardDesignSpade, 1, false))
		dealer.AddCard(NewCard(CardDesignHeart, 13, false)) // BJ

		hand := NewBlackJackHand()
		hand.AddCard(NewCard(CardDesignClover, 1, false))
		hand.AddCard(NewCard(CardDesignDiamond, 10, false)) // 21, 2 cards but handCount=2

		assert.Equal(t, GameResultLose, bj.judgeHandCore(hand, 2))
	})
}

// --- payoutHand tests ---

func TestPayoutHand(t *testing.T) {
	t.Run("win with BJ and handCount=1 gets 3:2 payout", func(t *testing.T) {
		player := NewBlackJackPlayer()
		player.SetChips(900)

		hand := NewBlackJackHand()
		hand.AddCard(NewCard(CardDesignClover, 1, false))
		hand.AddCard(NewCard(CardDesignDiamond, 13, false)) // BJ
		hand.SetBet(100)

		payoutHand(player, hand, 1, GameResultWin)
		assert.Equal(t, 900+250, player.GetChips()) // 100 + 150 = 250
	})

	t.Run("win with BJ but handCount>1 gets 2x payout", func(t *testing.T) {
		player := NewBlackJackPlayer()
		player.SetChips(900)

		hand := NewBlackJackHand()
		hand.AddCard(NewCard(CardDesignClover, 1, false))
		hand.AddCard(NewCard(CardDesignDiamond, 13, false)) // BJ
		hand.SetBet(100)

		payoutHand(player, hand, 2, GameResultWin)
		assert.Equal(t, 900+200, player.GetChips()) // normal 2x
	})

	t.Run("win without BJ gets 2x payout", func(t *testing.T) {
		player := NewBlackJackPlayer()
		player.SetChips(900)

		hand := NewBlackJackHand()
		hand.AddCard(NewCard(CardDesignClover, 10, false))
		hand.AddCard(NewCard(CardDesignDiamond, 9, false)) // 19, not BJ
		hand.SetBet(100)

		payoutHand(player, hand, 1, GameResultWin)
		assert.Equal(t, 900+200, player.GetChips())
	})

	t.Run("draw returns bet", func(t *testing.T) {
		player := NewBlackJackPlayer()
		player.SetChips(900)

		hand := NewBlackJackHand()
		hand.AddCard(NewCard(CardDesignClover, 10, false))
		hand.AddCard(NewCard(CardDesignDiamond, 8, false))
		hand.SetBet(100)

		payoutHand(player, hand, 1, GameResultDraw)
		assert.Equal(t, 900+100, player.GetChips())
	})

	t.Run("lose adds nothing", func(t *testing.T) {
		player := NewBlackJackPlayer()
		player.SetChips(900)

		hand := NewBlackJackHand()
		hand.AddCard(NewCard(CardDesignClover, 10, false))
		hand.AddCard(NewCard(CardDesignDiamond, 7, false))
		hand.SetBet(100)

		payoutHand(player, hand, 1, GameResultLose)
		assert.Equal(t, 900, player.GetChips())
	})
}

// --- judgeCpuHand tests ---

func TestJudgeCpuHand(t *testing.T) {
	t.Run("CPU bust loses", func(t *testing.T) {
		bj, _, dealer := setupInternalTestBJ(1000, 1000)
		dealer.AddCard(NewCard(CardDesignSpade, 10, false))
		dealer.AddCard(NewCard(CardDesignHeart, 7, false))

		hand := NewBlackJackHand()
		hand.AddCard(NewCard(CardDesignClover, 10, false))
		hand.AddCard(NewCard(CardDesignDiamond, 10, false))
		hand.AddCard(NewCard(CardDesignSpade, 5, false)) // 25, bust

		assert.Equal(t, GameResultLose, bj.judgeCpuHand(hand))
	})

	t.Run("dealer bust, CPU wins", func(t *testing.T) {
		bj, _, dealer := setupInternalTestBJ(1000, 1000)
		dealer.AddCard(NewCard(CardDesignSpade, 10, false))
		dealer.AddCard(NewCard(CardDesignHeart, 10, false))
		dealer.AddCard(NewCard(CardDesignClover, 5, false)) // 25, bust

		hand := NewBlackJackHand()
		hand.AddCard(NewCard(CardDesignDiamond, 10, false))
		hand.AddCard(NewCard(CardDesignSpade, 8, false)) // 18

		assert.Equal(t, GameResultWin, bj.judgeCpuHand(hand))
	})

	t.Run("CPU higher score wins", func(t *testing.T) {
		bj, _, dealer := setupInternalTestBJ(1000, 1000)
		dealer.AddCard(NewCard(CardDesignSpade, 10, false))
		dealer.AddCard(NewCard(CardDesignHeart, 7, false)) // 17

		hand := NewBlackJackHand()
		hand.AddCard(NewCard(CardDesignClover, 10, false))
		hand.AddCard(NewCard(CardDesignDiamond, 9, false)) // 19

		assert.Equal(t, GameResultWin, bj.judgeCpuHand(hand))
	})

	t.Run("dealer higher score, CPU loses", func(t *testing.T) {
		bj, _, dealer := setupInternalTestBJ(1000, 1000)
		dealer.AddCard(NewCard(CardDesignSpade, 10, false))
		dealer.AddCard(NewCard(CardDesignHeart, 9, false)) // 19

		hand := NewBlackJackHand()
		hand.AddCard(NewCard(CardDesignClover, 10, false))
		hand.AddCard(NewCard(CardDesignDiamond, 7, false)) // 17

		assert.Equal(t, GameResultLose, bj.judgeCpuHand(hand))
	})

	t.Run("same score draw", func(t *testing.T) {
		bj, _, dealer := setupInternalTestBJ(1000, 1000)
		dealer.AddCard(NewCard(CardDesignSpade, 10, false))
		dealer.AddCard(NewCard(CardDesignHeart, 8, false)) // 18

		hand := NewBlackJackHand()
		hand.AddCard(NewCard(CardDesignClover, 10, false))
		hand.AddCard(NewCard(CardDesignDiamond, 8, false)) // 18

		assert.Equal(t, GameResultDraw, bj.judgeCpuHand(hand))
	})

	t.Run("CPU BJ beats dealer non-BJ at 21", func(t *testing.T) {
		bj, _, dealer := setupInternalTestBJ(1000, 1000)
		dealer.AddCard(NewCard(CardDesignSpade, 7, false))
		dealer.AddCard(NewCard(CardDesignHeart, 7, false))
		dealer.AddCard(NewCard(CardDesignClover, 7, false)) // 21, but 3 cards

		hand := NewBlackJackHand()
		hand.AddCard(NewCard(CardDesignDiamond, 1, false))
		hand.AddCard(NewCard(CardDesignSpade, 13, false)) // BJ (2 cards, 21)

		assert.Equal(t, GameResultWin, bj.judgeCpuHand(hand))
	})

	t.Run("dealer BJ beats CPU non-BJ at 21", func(t *testing.T) {
		bj, _, dealer := setupInternalTestBJ(1000, 1000)
		dealer.AddCard(NewCard(CardDesignSpade, 1, false))
		dealer.AddCard(NewCard(CardDesignHeart, 13, false)) // BJ

		hand := NewBlackJackHand()
		hand.AddCard(NewCard(CardDesignClover, 7, false))
		hand.AddCard(NewCard(CardDesignDiamond, 7, false))
		hand.AddCard(NewCard(CardDesignSpade, 7, false)) // 21, but 3 cards

		assert.Equal(t, GameResultLose, bj.judgeCpuHand(hand))
	})

	t.Run("both BJ is draw", func(t *testing.T) {
		bj, _, dealer := setupInternalTestBJ(1000, 1000)
		dealer.AddCard(NewCard(CardDesignSpade, 1, false))
		dealer.AddCard(NewCard(CardDesignHeart, 13, false)) // BJ

		hand := NewBlackJackHand()
		hand.AddCard(NewCard(CardDesignClover, 1, false))
		hand.AddCard(NewCard(CardDesignDiamond, 10, false)) // BJ

		assert.Equal(t, GameResultDraw, bj.judgeCpuHand(hand))
	})
}

// --- resolvePayoutsCpu internal tests ---

func TestResolvePayoutsCpuInternal(t *testing.T) {
	t.Run("CPU with 0 cards skipped", func(t *testing.T) {
		bj, _, dealer := setupInternalTestBJ(1000, 1000)
		dealer.AddCard(NewCard(CardDesignSpade, 10, false))
		dealer.AddCard(NewCard(CardDesignHeart, 7, false))

		cpu := NewBlackJackCpuSeat()
		bj.cpuPlayers = []*BlackJackCpuSeat{cpu}
		initialChips := cpu.GetPlayer().GetChips()

		bj.resolvePayoutsCpu()

		assert.Equal(t, initialChips, cpu.GetPlayer().GetChips(), "empty hand CPU should not change chips")
	})

	t.Run("CPU surrendered hand skipped", func(t *testing.T) {
		bj, _, dealer := setupInternalTestBJ(1000, 1000)
		dealer.AddCard(NewCard(CardDesignSpade, 10, false))
		dealer.AddCard(NewCard(CardDesignHeart, 7, false))

		cpu := NewBlackJackCpuSeat()
		cpuHand := cpu.GetHands()[0]
		cpuHand.AddCard(NewCard(CardDesignClover, 10, false))
		cpuHand.AddCard(NewCard(CardDesignDiamond, 6, false))
		cpuHand.SetBet(50)
		cpuHand.SetSurrendered(true)
		bj.cpuPlayers = []*BlackJackCpuSeat{cpu}
		initialChips := cpu.GetPlayer().GetChips()

		bj.resolvePayoutsCpu()

		assert.Equal(t, initialChips, cpu.GetPlayer().GetChips(), "surrendered CPU hand should not add chips")
	})

	t.Run("CPU BJ with single hand gets 3:2 payout", func(t *testing.T) {
		bj, _, dealer := setupInternalTestBJ(1000, 1000)
		dealer.AddCard(NewCard(CardDesignSpade, 10, false))
		dealer.AddCard(NewCard(CardDesignHeart, 7, false)) // 17

		cpu := NewBlackJackCpuSeat()
		cpuHand := cpu.GetHands()[0]
		cpuHand.AddCard(NewCard(CardDesignClover, 1, false))
		cpuHand.AddCard(NewCard(CardDesignDiamond, 13, false)) // BJ
		cpuHand.SetBet(100)
		cpu.GetPlayer().SetChips(900) // 1000 - 100 bet
		bj.cpuPlayers = []*BlackJackCpuSeat{cpu}

		bj.resolvePayoutsCpu()

		// 3:2 payout: 100 + 150 = 250 added to chips
		assert.Equal(t, 900+250, cpu.GetPlayer().GetChips())
	})

	t.Run("CPU normal win gets 2x payout", func(t *testing.T) {
		bj, _, dealer := setupInternalTestBJ(1000, 1000)
		dealer.AddCard(NewCard(CardDesignSpade, 10, false))
		dealer.AddCard(NewCard(CardDesignHeart, 7, false)) // 17

		cpu := NewBlackJackCpuSeat()
		cpuHand := cpu.GetHands()[0]
		cpuHand.AddCard(NewCard(CardDesignClover, 10, false))
		cpuHand.AddCard(NewCard(CardDesignDiamond, 9, false)) // 19
		cpuHand.SetBet(100)
		cpu.GetPlayer().SetChips(900)
		bj.cpuPlayers = []*BlackJackCpuSeat{cpu}

		bj.resolvePayoutsCpu()

		assert.Equal(t, 900+200, cpu.GetPlayer().GetChips())
	})

	t.Run("CPU draw gets bet back", func(t *testing.T) {
		bj, _, dealer := setupInternalTestBJ(1000, 1000)
		dealer.AddCard(NewCard(CardDesignSpade, 10, false))
		dealer.AddCard(NewCard(CardDesignHeart, 8, false)) // 18

		cpu := NewBlackJackCpuSeat()
		cpuHand := cpu.GetHands()[0]
		cpuHand.AddCard(NewCard(CardDesignClover, 10, false))
		cpuHand.AddCard(NewCard(CardDesignDiamond, 8, false)) // 18
		cpuHand.SetBet(100)
		cpu.GetPlayer().SetChips(900)
		bj.cpuPlayers = []*BlackJackCpuSeat{cpu}

		bj.resolvePayoutsCpu()

		assert.Equal(t, 900+100, cpu.GetPlayer().GetChips())
	})

	t.Run("CPU lose gets nothing", func(t *testing.T) {
		bj, _, dealer := setupInternalTestBJ(1000, 1000)
		dealer.AddCard(NewCard(CardDesignSpade, 10, false))
		dealer.AddCard(NewCard(CardDesignHeart, 9, false)) // 19

		cpu := NewBlackJackCpuSeat()
		cpuHand := cpu.GetHands()[0]
		cpuHand.AddCard(NewCard(CardDesignClover, 10, false))
		cpuHand.AddCard(NewCard(CardDesignDiamond, 7, false)) // 17
		cpuHand.SetBet(100)
		cpu.GetPlayer().SetChips(900)
		bj.cpuPlayers = []*BlackJackCpuSeat{cpu}

		bj.resolvePayoutsCpu()

		assert.Equal(t, 900, cpu.GetPlayer().GetChips(), "CPU lose should not add chips")
	})
}

// --- cpuPlaySeat internal tests ---

func TestCpuPlaySeatInternal(t *testing.T) {
	t.Run("empty hand skipped", func(t *testing.T) {
		bj, _, _ := setupInternalTestBJ(1000, 1000)
		cpu := NewBlackJackCpuSeat()
		dealerUpcard := NewCard(CardDesignSpade, 10, false)
		bj.cpuPlaySeat(cpu, dealerUpcard)
		assert.Equal(t, 0, cpu.GetHands()[0].GetCardsSize())
	})

	t.Run("finished hand skipped", func(t *testing.T) {
		bj, _, _ := setupInternalTestBJ(1000, 1000)
		cpu := NewBlackJackCpuSeat()
		cpuHand := cpu.GetHands()[0]
		cpuHand.AddCard(NewCard(CardDesignClover, 10, false))
		cpuHand.AddCard(NewCard(CardDesignDiamond, 7, false))
		cpuHand.SetStood(true)
		cpuHand.SetBet(50)

		dealerUpcard := NewCard(CardDesignSpade, 6, false)
		bj.cpuPlaySeat(cpu, dealerUpcard)
		assert.True(t, cpuHand.IsStood())
		assert.Equal(t, 2, cpuHand.GetCardsSize(), "finished hand should not get more cards")
	})

	t.Run("nil dealer upcard skips CPU play", func(t *testing.T) {
		bj, _, _ := setupInternalTestBJ(1000, 1000)
		cpu := NewBlackJackCpuSeat()
		cpuHand := cpu.GetHands()[0]
		cpuHand.AddCard(NewCard(CardDesignClover, 10, false))
		cpuHand.AddCard(NewCard(CardDesignDiamond, 7, false))
		cpuHand.SetBet(50)

		bj.cpuPlayers = []*BlackJackCpuSeat{cpu}
		bj.dealer.Reset() // no cards -> GetCard(0) returns nil
		bj.cpuPlay()
		assert.False(t, cpuHand.IsFinished())
	})
}

// --- cpuHit internal tests ---

func TestCpuHitInternal(t *testing.T) {
	t.Run("deck exhausted during hit causes stand", func(t *testing.T) {
		tc := NewTrumpCards(0)
		player := NewBlackJackPlayer()
		dealer := NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := NewBlackJack(tc, player, dealer)

		for tc.DrawCard() != nil {
		}

		hand := NewBlackJackHand()
		hand.AddCard(NewCard(CardDesignSpade, 10, false))
		hand.AddCard(NewCard(CardDesignHeart, 5, false))

		bj.cpuHit(hand)
		assert.True(t, hand.IsStood(), "should stand when deck is exhausted")
		assert.Equal(t, 2, hand.GetCardsSize(), "should not have drawn a card")
	})

	t.Run("bust after hit", func(t *testing.T) {
		tc := NewTrumpCards(0)
		player := NewBlackJackPlayer()
		dealer := NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := NewBlackJack(tc, player, dealer)

		bj.trumpCards.deck[0] = NewCard(CardDesignClover, 13, false) // King
		bj.trumpCards.deckDrawCnt = 0

		hand := NewBlackJackHand()
		hand.AddCard(NewCard(CardDesignSpade, 10, false))
		hand.AddCard(NewCard(CardDesignHeart, 10, false)) // 20

		bj.cpuHit(hand)
		assert.True(t, hand.IsBusted())
		assert.Equal(t, 3, hand.GetCardsSize())
	})

	t.Run("no bust after hit", func(t *testing.T) {
		tc := NewTrumpCards(0)
		player := NewBlackJackPlayer()
		dealer := NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := NewBlackJack(tc, player, dealer)

		bj.trumpCards.deck[0] = NewCard(CardDesignClover, 2, false) // 2
		bj.trumpCards.deckDrawCnt = 0

		hand := NewBlackJackHand()
		hand.AddCard(NewCard(CardDesignSpade, 10, false))
		hand.AddCard(NewCard(CardDesignHeart, 5, false)) // 15

		bj.cpuHit(hand)
		assert.False(t, hand.IsBusted())
		assert.Equal(t, 3, hand.GetCardsSize())
		assert.Equal(t, 17, hand.GetScore())
	})
}

// --- cpuDoubleDown internal tests ---

func TestCpuDoubleDownInternal(t *testing.T) {
	t.Run("deck exhausted during double down", func(t *testing.T) {
		tc := NewTrumpCards(0)
		player := NewBlackJackPlayer()
		dealer := NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := NewBlackJack(tc, player, dealer)

		for tc.DrawCard() != nil {
		}

		cpu := NewBlackJackCpuSeat()
		hand := cpu.GetHands()[0]
		hand.AddCard(NewCard(CardDesignSpade, 5, false))
		hand.AddCard(NewCard(CardDesignHeart, 6, false)) // 11
		hand.SetBet(50)

		bj.cpuDoubleDown(cpu, hand)

		assert.Equal(t, 50, hand.GetBet())
		assert.False(t, hand.IsDoubled())
		assert.True(t, hand.IsStood())
	})

	t.Run("successful double down no bust", func(t *testing.T) {
		tc := NewTrumpCards(0)
		player := NewBlackJackPlayer()
		dealer := NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := NewBlackJack(tc, player, dealer)

		bj.trumpCards.deck[0] = NewCard(CardDesignClover, 3, false)
		bj.trumpCards.deckDrawCnt = 0

		cpu := NewBlackJackCpuSeat()
		hand := cpu.GetHands()[0]
		hand.AddCard(NewCard(CardDesignSpade, 5, false))
		hand.AddCard(NewCard(CardDesignHeart, 6, false)) // 11
		hand.SetBet(50)

		bj.cpuDoubleDown(cpu, hand)

		assert.Equal(t, 100, hand.GetBet())
		assert.True(t, hand.IsDoubled())
		assert.Equal(t, 3, hand.GetCardsSize())
		assert.True(t, hand.IsStood())
		assert.False(t, hand.IsBusted())
	})

	t.Run("double down bust", func(t *testing.T) {
		tc := NewTrumpCards(0)
		player := NewBlackJackPlayer()
		dealer := NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := NewBlackJack(tc, player, dealer)

		bj.trumpCards.deck[0] = NewCard(CardDesignClover, 13, false) // King
		bj.trumpCards.deckDrawCnt = 0

		cpu := NewBlackJackCpuSeat()
		hand := cpu.GetHands()[0]
		hand.AddCard(NewCard(CardDesignSpade, 10, false))
		hand.AddCard(NewCard(CardDesignHeart, 5, false)) // 15
		hand.SetBet(50)

		bj.cpuDoubleDown(cpu, hand)

		assert.Equal(t, 100, hand.GetBet())
		assert.True(t, hand.IsDoubled())
		assert.True(t, hand.IsBusted())
	})
}

// --- cpuSplit internal tests ---

func TestCpuSplitInternal(t *testing.T) {
	t.Run("ace split auto-stands both hands", func(t *testing.T) {
		tc := NewTrumpCardsWithDecks(2, 0)
		player := NewBlackJackPlayer()
		dealer := NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := NewBlackJack(tc, player, dealer)

		cpu := NewBlackJackCpuSeat()
		hand := cpu.GetHands()[0]
		hand.AddCard(NewCard(CardDesignSpade, 1, false))
		hand.AddCard(NewCard(CardDesignHeart, 1, false)) // pair of aces
		hand.SetBet(50)

		dealerUpcard := NewCard(CardDesignClover, 6, false)
		bj.cpuSplit(cpu, hand, 0, dealerUpcard)

		for _, h := range cpu.GetHands() {
			assert.True(t, h.IsStood(), "ace split should auto-stand")
		}
		assert.Equal(t, 2, len(cpu.GetHands()))
	})

	t.Run("non-ace split does not auto-stand", func(t *testing.T) {
		tc := NewTrumpCardsWithDecks(2, 0)
		player := NewBlackJackPlayer()
		dealer := NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := NewBlackJack(tc, player, dealer)

		cpu := NewBlackJackCpuSeat()
		hand := cpu.GetHands()[0]
		hand.AddCard(NewCard(CardDesignSpade, 8, false))
		hand.AddCard(NewCard(CardDesignHeart, 8, false)) // pair of 8s
		hand.SetBet(50)

		dealerUpcard := NewCard(CardDesignClover, 6, false)
		bj.cpuSplit(cpu, hand, 0, dealerUpcard)

		assert.Equal(t, 2, len(cpu.GetHands()))
	})

	t.Run("split with partial deck exhaustion rolls back", func(t *testing.T) {
		tc := NewTrumpCards(0)
		player := NewBlackJackPlayer()
		dealer := NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := NewBlackJack(tc, player, dealer)

		// Leave only 1 card in deck (first draw succeeds, second returns nil)
		for i := 0; i < 51; i++ {
			tc.DrawCard()
		}

		cpu := NewBlackJackCpuSeat()
		cpu.GetPlayer().SetChips(200)
		hand := cpu.GetHands()[0]
		hand.AddCard(NewCard(CardDesignSpade, 8, false))
		hand.AddCard(NewCard(CardDesignHeart, 8, false))
		hand.SetBet(50)

		dealerUpcard := NewCard(CardDesignClover, 6, false)
		bj.cpuSplit(cpu, hand, 0, dealerUpcard)

		// Rollback: should remain 1 hand with original 2 cards restored
		// cpuSplit subtracts bet (50) at entry → 150, then rollback adds it back → 200
		assert.Equal(t, 1, len(cpu.GetHands()))
		assert.Equal(t, 2, hand.GetCardsSize())
		assert.True(t, hand.IsStood(), "hand should be stood after rollback")
		assert.Equal(t, 200, cpu.GetPlayer().GetChips(), "bet should be refunded")
	})

	t.Run("split with full deck exhaustion rolls back", func(t *testing.T) {
		tc := NewTrumpCards(0)
		player := NewBlackJackPlayer()
		dealer := NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := NewBlackJack(tc, player, dealer)

		// Empty the deck completely
		for tc.DrawCard() != nil {
		}

		cpu := NewBlackJackCpuSeat()
		cpu.GetPlayer().SetChips(200)
		hand := cpu.GetHands()[0]
		hand.AddCard(NewCard(CardDesignSpade, 8, false))
		hand.AddCard(NewCard(CardDesignHeart, 8, false))
		hand.SetBet(50)

		dealerUpcard := NewCard(CardDesignClover, 6, false)
		bj.cpuSplit(cpu, hand, 0, dealerUpcard)

		// Rollback: should remain 1 hand with original 2 cards restored
		assert.Equal(t, 1, len(cpu.GetHands()))
		assert.Equal(t, 2, hand.GetCardsSize())
		assert.True(t, hand.IsStood(), "hand should be stood after rollback")
		assert.Equal(t, 200, cpu.GetPlayer().GetChips(), "bet should be refunded")
	})

	t.Run("split partial exhaustion restores running count", func(t *testing.T) {
		tc := NewTrumpCards(0)
		player := NewBlackJackPlayer()
		dealer := NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := NewBlackJack(tc, player, dealer)
		bj.config.CountingEnabled = true

		// Leave only 1 card in deck
		for i := 0; i < 51; i++ {
			tc.DrawCard()
		}

		cpu := NewBlackJackCpuSeat()
		cpu.GetPlayer().SetChips(200)
		hand := cpu.GetHands()[0]
		hand.AddCard(NewCard(CardDesignSpade, 8, false))
		hand.AddCard(NewCard(CardDesignHeart, 8, false))
		hand.SetBet(50)

		countBefore := bj.runningCount
		dealerUpcard := NewCard(CardDesignClover, 6, false)
		bj.cpuSplit(cpu, hand, 0, dealerUpcard)

		// Running count should be restored after rollback
		assert.Equal(t, countBefore, bj.runningCount, "running count should be rolled back")
	})
}

// --- dealerPlay with counting ---

func TestDealerPlaySkipsDraw_AllBustedWithCounting(t *testing.T) {
	bj, _, dealer := setupInternalTestBJ(1000, 1000)
	bj.config.CountingEnabled = true
	bj.holeCardCounted = false

	dealer.AddCard(NewCard(CardDesignSpade, 10, false))
	dealer.AddCard(NewCard(CardDesignHeart, 7, false))

	hand := bj.playerHands[0]
	hand.SetBusted(true)

	bj.dealerPlay()

	assert.True(t, bj.holeCardCounted)
	assert.True(t, bj.gameEndFlag)
}

// --- DealerHit with deck exhaustion ---

func TestDealerHitDeckExhausted(t *testing.T) {
	tc := NewTrumpCards(0)
	player := NewBlackJackPlayer()
	dealer := NewBlackJackPlayer()
	player.SetChips(1000)
	dealer.SetChips(1000)
	bj := NewBlackJack(tc, player, dealer)

	for tc.DrawCard() != nil {
	}

	dealer.AddCard(NewCard(CardDesignSpade, 10, false))
	dealer.AddCard(NewCard(CardDesignHeart, 5, false))

	bj.DealerHit()

	assert.True(t, bj.gameEndFlag)
}

// --- cpuBetAndDeal with low chips ---

func TestCpuBetAndDealLowChips(t *testing.T) {
	t.Run("CPU with chips between MinBet and CpuBetAmount", func(t *testing.T) {
		tc := NewTrumpCardsWithDecks(2, 0)
		player := NewBlackJackPlayer()
		dealer := NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := NewBlackJack(tc, player, dealer)

		cpu := NewBlackJackCpuSeat()
		// Set chips to 30 (between BJMinBet=10 and BJCpuBetAmount=50)
		cpu.GetPlayer().SetChips(30)
		bj.cpuPlayers = []*BlackJackCpuSeat{cpu}

		bj.cpuBetAndDeal()

		// Should bet 30 rounded down to multiple of BJMinBet = 30
		cpuHand := cpu.GetHands()[0]
		assert.Equal(t, 30, cpuHand.GetBet())
		assert.Equal(t, 2, cpuHand.GetCardsSize())
		assert.Equal(t, 0, cpu.GetPlayer().GetChips())
	})

	t.Run("CPU with chips below MinBet is skipped", func(t *testing.T) {
		tc := NewTrumpCardsWithDecks(2, 0)
		player := NewBlackJackPlayer()
		dealer := NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := NewBlackJack(tc, player, dealer)

		cpu := NewBlackJackCpuSeat()
		cpu.GetPlayer().SetChips(5) // below BJMinBet
		bj.cpuPlayers = []*BlackJackCpuSeat{cpu}

		bj.cpuBetAndDeal()

		cpuHand := cpu.GetHands()[0]
		assert.Equal(t, 0, cpuHand.GetBet())
		assert.Equal(t, 0, cpuHand.GetCardsSize())
	})

	t.Run("CPU bet rounds down to multiple of MinBet", func(t *testing.T) {
		tc := NewTrumpCardsWithDecks(2, 0)
		player := NewBlackJackPlayer()
		dealer := NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := NewBlackJack(tc, player, dealer)

		cpu := NewBlackJackCpuSeat()
		cpu.GetPlayer().SetChips(45) // 45 / 10 * 10 = 40
		bj.cpuPlayers = []*BlackJackCpuSeat{cpu}

		bj.cpuBetAndDeal()

		cpuHand := cpu.GetHands()[0]
		assert.Equal(t, 40, cpuHand.GetBet())
		assert.Equal(t, 5, cpu.GetPlayer().GetChips())
	})
}

// --- cpuBetAndDeal with deck exhaustion ---

func TestCpuBetAndDealDeckExhaustion(t *testing.T) {
	t.Run("full deck exhaustion refunds chips and resets hand", func(t *testing.T) {
		tc := NewTrumpCards(0)
		player := NewBlackJackPlayer()
		dealer := NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := NewBlackJack(tc, player, dealer)

		// Empty the deck completely
		for tc.DrawCard() != nil {
		}

		cpu := NewBlackJackCpuSeat()
		cpu.GetPlayer().SetChips(200)
		bj.cpuPlayers = []*BlackJackCpuSeat{cpu}

		bj.cpuBetAndDeal()

		cpuHand := cpu.GetHands()[0]
		assert.Equal(t, 0, cpuHand.GetCardsSize(), "hand should be reset")
		assert.Equal(t, 200, cpu.GetPlayer().GetChips(), "chips should be refunded")
	})

	t.Run("partial deck exhaustion refunds chips and resets hand", func(t *testing.T) {
		tc := NewTrumpCards(0)
		player := NewBlackJackPlayer()
		dealer := NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := NewBlackJack(tc, player, dealer)

		// Leave only 1 card in deck
		for i := 0; i < 51; i++ {
			tc.DrawCard()
		}

		cpu := NewBlackJackCpuSeat()
		cpu.GetPlayer().SetChips(200)
		bj.cpuPlayers = []*BlackJackCpuSeat{cpu}

		bj.cpuBetAndDeal()

		cpuHand := cpu.GetHands()[0]
		assert.Equal(t, 0, cpuHand.GetCardsSize(), "hand should be reset")
		assert.Equal(t, 200, cpu.GetPlayer().GetChips(), "chips should be refunded")
	})
}

// TestBlackJack_AllBustedSkipsDealerDraw verifies that when all player hands
// are busted, dealerPlay skips drawing cards. Uses internal access to stack the
// deck and call PlayerHit to bust the hand naturally.
func TestBlackJack_AllBustedSkipsDealerDraw(t *testing.T) {
	bj, player, dealer := setupInternalTestBJ(1000, 1000)

	// Set up player hand with score 20 (10 + 10) and a 100 chip bet
	hand := bj.playerHands[0]
	hand.SetBet(100)
	player.SubtractChips(100)
	hand.AddCard(NewCard(CardDesignSpade, 10, false))
	hand.AddCard(NewCard(CardDesignHeart, 10, false))

	// Set up dealer with score 11 (5 + 6); would draw if not all-busted
	dealer.AddCard(NewCard(CardDesignDiamond, 5, false))
	dealer.AddCard(NewCard(CardDesignDiamond, 6, false))

	// Stack the deck so PlayerHit draws a King (BJ value 10): 20 + 10 = 30 → bust
	bj.trumpCards.deck[0] = NewCard(CardDesignClover, 13, false)
	bj.trumpCards.deckDrawCnt = 0
	bj.phase = BJPhaseAction

	err := bj.PlayerHit()
	assert.NoError(t, err)

	// Hand should be busted, game should end
	assert.True(t, hand.IsBusted())
	assert.True(t, bj.gameEndFlag)
	assert.Equal(t, BJPhaseEnd, bj.phase)
	// Dealer should NOT have drawn any additional cards (still 2)
	assert.Equal(t, 2, dealer.GetCardsSize())
}

// TestBlackJack_PlayerBet_DealerAceTriggersInsurance_Deterministic exercises
// lines 131-134 of BlackJack.go: when the dealer's first card is an Ace
// (value==1), insuranceAvailable is set to true and phase becomes
// BJPhaseInsurance. Instead of relying on random shuffles, the deck is
// directly stacked so the dealer's first dealt card is guaranteed to be an Ace.
func TestBlackJack_PlayerBet_DealerAceTriggersInsurance_Deterministic(t *testing.T) {
	bj, _, _ := setupInternalTestBJ(BJDefaultChips, BJDefaultChips)

	// Stack the deck so that the deal produces a known layout.
	// PlayerBet deals cards in this order (2 iterations of the loop):
	//   deck[0] -> player, deck[1] -> dealer,
	//   deck[2] -> player, deck[3] -> dealer
	// We place an Ace at deck[1] so dealer's first card is an Ace.
	bj.trumpCards.deck[0] = NewCard(CardDesignSpade, 10, false)  // player card 1
	bj.trumpCards.deck[1] = NewCard(CardDesignHeart, 1, false)   // dealer card 1 (Ace)
	bj.trumpCards.deck[2] = NewCard(CardDesignClover, 9, false)  // player card 2
	bj.trumpCards.deck[3] = NewCard(CardDesignDiamond, 7, false) // dealer card 2 (non-10 to avoid natural BJ)
	bj.trumpCards.deckDrawCnt = 0

	// Reset phase to Bet so PlayerBet can proceed
	bj.phase = BJPhaseBet

	err := bj.PlayerBet(BJMinBet, 0, 0)
	assert.NoError(t, err)

	// Verify insurance-specific state
	assert.True(t, bj.insuranceAvailable, "insuranceAvailable should be true when dealer shows Ace")
	assert.Equal(t, BJPhaseInsurance, bj.phase)

	// Confirm dealer's first card is an Ace
	dealerCard := bj.dealer.GetCard(0)
	assert.NotNil(t, dealerCard)
	assert.Equal(t, 1, dealerCard.GetValue(), "dealer's first card should be an Ace")

	// Confirm game has not ended yet (insurance decision pending)
	assert.False(t, bj.gameEndFlag)
}

// --- PlayerBet with counting + CPU players ---

func TestPlayerBet_CountingWithCpuPlayers(t *testing.T) {
	t.Run("counting enabled counts CPU cards after deal", func(t *testing.T) {
		bj, _, _ := setupInternalTestBJ(BJDefaultChips, BJDefaultChips)
		bj.config.CountingEnabled = true
		bj.config.CpuPlayerCount = 1
		bj.phase = BJPhaseBet
		bj.runningCount = 0

		// Initialize CPU players
		bj.initCpuPlayers()

		// Stack the deck:
		// deck[0] -> player card 1
		// deck[1] -> dealer card 1 (non-Ace to skip insurance)
		// deck[2] -> player card 2
		// deck[3] -> dealer card 2
		// deck[4] -> CPU card 1 (via cpuBetAndDeal)
		// deck[5] -> CPU card 2 (via cpuBetAndDeal)
		bj.trumpCards.deck[0] = NewCard(CardDesignSpade, 5, false)   // player +1
		bj.trumpCards.deck[1] = NewCard(CardDesignHeart, 10, false)  // dealer upcard -1
		bj.trumpCards.deck[2] = NewCard(CardDesignClover, 3, false)  // player +1
		bj.trumpCards.deck[3] = NewCard(CardDesignDiamond, 8, false) // dealer hole (not counted yet)
		bj.trumpCards.deck[4] = NewCard(CardDesignSpade, 2, false)   // CPU card 1 +1
		bj.trumpCards.deck[5] = NewCard(CardDesignHeart, 4, false)   // CPU card 2 +1
		bj.trumpCards.deckDrawCnt = 0

		err := bj.PlayerBet(BJMinBet, 0, 0)
		assert.NoError(t, err)

		// Running count should include: player 2 cards + dealer upcard + CPU 2 cards
		// 5(+1) + 10(-1) + 3(+1) + 2(+1) + 4(+1) = +3
		assert.Equal(t, 3, bj.runningCount)
		assert.Equal(t, BJPhaseAction, bj.phase) // non-Ace dealer, no insurance
	})
}

// --- DealerHit with counting ---

func TestDealerHit_CountingHoleCard(t *testing.T) {
	t.Run("hole card counted during DealerHit when counting enabled", func(t *testing.T) {
		bj, _, dealer := setupInternalTestBJ(1000, 1000)
		bj.config.CountingEnabled = true
		bj.holeCardCounted = false
		bj.runningCount = 0

		// Dealer has score 15 (10+5) so will hit
		dealer.AddCard(NewCard(CardDesignSpade, 10, false))
		dealer.AddCard(NewCard(CardDesignHeart, 5, false))

		// Player stands at 20 (not busted)
		hand := bj.playerHands[0]
		hand.AddCard(NewCard(CardDesignClover, 10, false))
		hand.AddCard(NewCard(CardDesignDiamond, 10, false))
		hand.SetStood(true)

		// Stack deck so dealer draws a King (10) -> 25 bust
		bj.trumpCards.deck[0] = NewCard(CardDesignClover, 13, false)
		bj.trumpCards.deckDrawCnt = 0

		bj.DealerHit()

		assert.True(t, bj.holeCardCounted, "hole card should be counted during DealerHit")
		assert.True(t, bj.gameEndFlag)
		// Running count: hole card (5 -> +1), drawn King (-1) = 0
		assert.Equal(t, 0, bj.runningCount)
	})
}

// --- cpuPlaySeat comprehensive branch tests ---

func TestCpuPlaySeat_DoubleAction(t *testing.T) {
	t.Run("CPU double with 2 cards and enough chips", func(t *testing.T) {
		tc := NewTrumpCardsWithDecks(4, 0)
		player := NewBlackJackPlayer()
		dealer := NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := NewBlackJack(tc, player, dealer)

		cpu := NewBlackJackCpuSeat()
		hand := cpu.GetHands()[0]
		// Hand: 5+6 = hard 11, dealer upcard = 6 -> strategy returns Double
		hand.AddCard(NewCard(CardDesignSpade, 5, false))
		hand.AddCard(NewCard(CardDesignHeart, 6, false))
		hand.SetBet(50)
		cpu.GetPlayer().SetChips(950) // enough for double

		dealerUpcard := NewCard(CardDesignClover, 6, false)
		bj.cpuPlaySeat(cpu, dealerUpcard)

		assert.True(t, hand.IsDoubled(), "should have doubled down")
		assert.True(t, hand.IsFinished(), "hand should be finished after double")
	})

	t.Run("CPU double falls back to hit when not enough chips", func(t *testing.T) {
		tc := NewTrumpCardsWithDecks(4, 0)
		player := NewBlackJackPlayer()
		dealer := NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := NewBlackJack(tc, player, dealer)

		cpu := NewBlackJackCpuSeat()
		hand := cpu.GetHands()[0]
		// Hand: 5+6 = hard 11, dealer upcard = 6 -> strategy returns Double
		hand.AddCard(NewCard(CardDesignSpade, 5, false))
		hand.AddCard(NewCard(CardDesignHeart, 6, false))
		hand.SetBet(50)
		cpu.GetPlayer().SetChips(10) // not enough for double (need 50)

		dealerUpcard := NewCard(CardDesignClover, 6, false)
		bj.cpuPlaySeat(cpu, dealerUpcard)

		assert.False(t, hand.IsDoubled(), "should not have doubled (insufficient chips)")
		assert.True(t, hand.IsFinished(), "hand should be finished")
	})

}

func TestCpuPlaySeat_SplitAction(t *testing.T) {
	t.Run("CPU split when conditions are met", func(t *testing.T) {
		tc := NewTrumpCardsWithDecks(4, 0)
		player := NewBlackJackPlayer()
		dealer := NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := NewBlackJack(tc, player, dealer)

		cpu := NewBlackJackCpuSeat()
		hand := cpu.GetHands()[0]
		// Pair of 8s vs dealer 6 -> strategy returns Split
		hand.AddCard(NewCard(CardDesignSpade, 8, false))
		hand.AddCard(NewCard(CardDesignHeart, 8, false))
		hand.SetBet(50)
		cpu.GetPlayer().SetChips(950) // enough for split

		dealerUpcard := NewCard(CardDesignClover, 6, false)
		bj.cpuPlaySeat(cpu, dealerUpcard)

		assert.GreaterOrEqual(t, len(cpu.GetHands()), 2, "should have split into 2+ hands")
	})

	t.Run("CPU split falls back to hit when not enough chips", func(t *testing.T) {
		tc := NewTrumpCardsWithDecks(4, 0)
		player := NewBlackJackPlayer()
		dealer := NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := NewBlackJack(tc, player, dealer)

		cpu := NewBlackJackCpuSeat()
		hand := cpu.GetHands()[0]
		// Pair of 8s vs dealer 6 -> strategy returns Split
		hand.AddCard(NewCard(CardDesignSpade, 8, false))
		hand.AddCard(NewCard(CardDesignHeart, 8, false))
		hand.SetBet(50)
		cpu.GetPlayer().SetChips(10) // not enough for split (need 50)

		dealerUpcard := NewCard(CardDesignClover, 6, false)
		bj.cpuPlaySeat(cpu, dealerUpcard)

		assert.Equal(t, 1, len(cpu.GetHands()), "should not have split (insufficient chips)")
		assert.True(t, hand.IsFinished())
	})

	t.Run("CPU split falls back to hit when max hands reached", func(t *testing.T) {
		tc := NewTrumpCardsWithDecks(4, 0)
		player := NewBlackJackPlayer()
		dealer := NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := NewBlackJack(tc, player, dealer)

		cpu := NewBlackJackCpuSeat()
		// Fill up to BJMaxHands (4)
		hand0 := NewBlackJackHand()
		hand0.SetBet(50)
		hand0.SetStood(true)
		hand0.AddCard(NewCard(CardDesignSpade, 10, false))
		hand0.AddCard(NewCard(CardDesignHeart, 9, false))
		hand1 := NewBlackJackHand()
		hand1.SetBet(50)
		hand1.SetStood(true)
		hand1.AddCard(NewCard(CardDesignClover, 10, false))
		hand1.AddCard(NewCard(CardDesignDiamond, 9, false))
		hand2 := NewBlackJackHand()
		hand2.SetBet(50)
		hand2.SetStood(true)
		hand2.AddCard(NewCard(CardDesignSpade, 10, false))
		hand2.AddCard(NewCard(CardDesignHeart, 7, false))
		hand3 := NewBlackJackHand()
		hand3.SetBet(50)
		// Pair of 8s -> strategy returns Split
		hand3.AddCard(NewCard(CardDesignClover, 8, false))
		hand3.AddCard(NewCard(CardDesignDiamond, 8, false))
		cpu.SetHands([]*BlackJackHand{hand0, hand1, hand2, hand3})
		cpu.GetPlayer().SetChips(800)

		dealerUpcard := NewCard(CardDesignClover, 6, false)
		bj.cpuPlaySeat(cpu, dealerUpcard)

		assert.Equal(t, 4, len(cpu.GetHands()), "should not split when at max hands")
		assert.True(t, hand3.IsFinished())
	})
}

func TestCpuPlaySeat_SurrenderAction(t *testing.T) {
	t.Run("CPU surrender when allowed", func(t *testing.T) {
		tc := NewTrumpCardsWithDecks(4, 0)
		player := NewBlackJackPlayer()
		dealer := NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := NewBlackJack(tc, player, dealer)

		cpu := NewBlackJackCpuSeat()
		hand := cpu.GetHands()[0]
		// Hard 16 vs dealer 9 -> strategy returns Surrender (Rh)
		hand.AddCard(NewCard(CardDesignSpade, 10, false))
		hand.AddCard(NewCard(CardDesignHeart, 6, false))
		hand.SetBet(100)
		cpu.GetPlayer().SetChips(900)

		dealerUpcard := NewCard(CardDesignClover, 9, false)
		bj.cpuPlaySeat(cpu, dealerUpcard)

		assert.True(t, hand.IsSurrendered(), "should have surrendered")
		assert.Equal(t, 950, cpu.GetPlayer().GetChips(), "should get half bet back")
	})

	t.Run("CPU surrender falls back to hit when not allowed", func(t *testing.T) {
		tc := NewTrumpCardsWithDecks(4, 0)
		player := NewBlackJackPlayer()
		dealer := NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := NewBlackJack(tc, player, dealer)
		bj.config.DealerHitsSoft17 = true // H17 to get surrender on hard 15 vs A

		cpu := NewBlackJackCpuSeat()
		hand := cpu.GetHands()[0]
		// Hard 15 vs dealer Ace with H17 -> Surrender
		// But make the hand already have 3 cards so CanSurrender returns false
		hand.AddCard(NewCard(CardDesignSpade, 5, false))
		hand.AddCard(NewCard(CardDesignHeart, 3, false))
		hand.AddCard(NewCard(CardDesignClover, 7, false)) // 15, 3 cards -> can't surrender
		hand.SetBet(100)
		cpu.GetPlayer().SetChips(900)

		dealerUpcard := NewCard(CardDesignSpade, 1, false) // Ace
		bj.cpuPlaySeat(cpu, dealerUpcard)

		assert.False(t, hand.IsSurrendered(), "should not surrender with 3 cards")
		assert.True(t, hand.IsFinished())
	})
}

func TestCpuPlaySeat_DoubleStandAction(t *testing.T) {
	t.Run("DoubleStand with enough chips doubles", func(t *testing.T) {
		bj, _, _ := setupInternalTestBJ(1000, 1000)
		cpu := NewBlackJackCpuSeat()
		hand := cpu.GetHands()[0]
		hand.SetBet(50)
		// A+7 = soft 18 vs dealer 2 → DoubleStand (Ds)
		hand.AddCard(NewCard(CardDesignSpade, 1, false))
		hand.AddCard(NewCard(CardDesignHeart, 7, false))
		cpu.GetPlayer().SetChips(100) // enough to double

		dealerUpcard := NewCard(CardDesignClover, 2, false)
		bj.cpuPlaySeat(cpu, dealerUpcard)

		assert.True(t, hand.IsDoubled(), "should have doubled down")
	})

	t.Run("DoubleStand without enough chips stands", func(t *testing.T) {
		bj, _, _ := setupInternalTestBJ(1000, 1000)
		cpu := NewBlackJackCpuSeat()
		hand := cpu.GetHands()[0]
		hand.SetBet(50)
		// A+7 = soft 18 vs dealer 2 → DoubleStand (Ds)
		hand.AddCard(NewCard(CardDesignSpade, 1, false))
		hand.AddCard(NewCard(CardDesignHeart, 7, false))
		cpu.GetPlayer().SetChips(10) // insufficient to double

		dealerUpcard := NewCard(CardDesignClover, 2, false)
		bj.cpuPlaySeat(cpu, dealerUpcard)

		assert.False(t, hand.IsDoubled(), "should not have doubled (insufficient chips)")
		assert.True(t, hand.IsStood(), "should stand as fallback (not hit)")
		assert.Equal(t, 2, hand.GetCardsSize(), "should not have drawn extra cards")
	})

	t.Run("DoubleStand with 3 cards stands", func(t *testing.T) {
		bj, _, _ := setupInternalTestBJ(1000, 1000)
		cpu := NewBlackJackCpuSeat()
		hand := cpu.GetHands()[0]
		hand.SetBet(50)
		// A+4+3 = soft 18 (not 2 cards) vs dealer 2 → DoubleStand
		hand.AddCard(NewCard(CardDesignSpade, 1, false))
		hand.AddCard(NewCard(CardDesignHeart, 4, false))
		hand.AddCard(NewCard(CardDesignClover, 3, false))
		cpu.GetPlayer().SetChips(100)

		dealerUpcard := NewCard(CardDesignClover, 2, false)
		bj.cpuPlaySeat(cpu, dealerUpcard)

		assert.False(t, hand.IsDoubled(), "can't double with 3 cards")
		assert.True(t, hand.IsStood(), "should stand as fallback")
	})
}

func TestPlayerSplit_RunningCountRollbackOnPartialFailure(t *testing.T) {
	t.Run("running count is corrected when card2 draw fails", func(t *testing.T) {
		bj := NewDefaultBlackJack()
		cfg := BlackJackConfig{CountingEnabled: true}
		_ = bj.SetConfig(cfg)

		// Create a tiny deck: enough for deal + 1 split card but not 2
		tc := NewTrumpCardsWithDecks(1, 0)
		bj.trumpCards = tc
		bj.Reset()

		// Set up manually: player has pair of 5s, dealer has 10+7
		bj.player.Reset()
		bj.dealer.Reset()
		for _, h := range bj.playerHands {
			h.Reset()
		}
		bj.playerHands[0].AddCard(NewCard(CardDesignSpade, 5, false))
		bj.playerHands[0].AddCard(NewCard(CardDesignHeart, 5, false))
		bj.playerHands[0].SetBet(10)
		bj.dealer.AddCard(NewCard(CardDesignClover, 10, false))
		bj.dealer.AddCard(NewCard(CardDesignDiamond, 7, false))
		bj.player.SetChips(1000)
		bj.phase = BJPhaseAction
		bj.runningCount = 0

		// Drain the deck until only 1 card remains
		for tc.GetRemainingCount() > 1 {
			tc.DrawCard()
		}

		// Now there's exactly 1 card left - split will draw card1 but fail on card2
		countBefore := bj.runningCount
		err := bj.PlayerSplit()
		assert.Error(t, err, "should fail due to deck exhaustion on 2nd card")

		// Running count should be back to what it was before the split attempt
		assert.Equal(t, countBefore, bj.runningCount, "running count should be rolled back")
	})
}
