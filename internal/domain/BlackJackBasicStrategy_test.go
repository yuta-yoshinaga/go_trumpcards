package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestGetBasicStrategyAction_Pairs(t *testing.T) {
	// Ace pair always splits
	hand := domain.NewBlackJackHand()
	hand.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	hand.AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))
	for _, dealerVal := range []int{2, 5, 10, 1} {
		upcard := domain.NewCard(domain.CardDesignClover, dealerVal, false)
		assert.Equal(t, domain.BJSuggestSplit, domain.GetBasicStrategyAction(hand, upcard, false), "A,A vs %d", dealerVal)
	}

	// 8,8 always splits
	hand8 := domain.NewBlackJackHand()
	hand8.AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
	hand8.AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
	assert.Equal(t, domain.BJSuggestSplit, domain.GetBasicStrategyAction(hand8, domain.NewCard(domain.CardDesignClover, 9, false), false))

	// 10,10 never splits
	hand10 := domain.NewBlackJackHand()
	hand10.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
	hand10.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
	assert.Equal(t, domain.BJSuggestStand, domain.GetBasicStrategyAction(hand10, domain.NewCard(domain.CardDesignClover, 6, false), false))

	// 5,5 treated as hard 10 (double vs 2-9, hit vs 10/A)
	hand5 := domain.NewBlackJackHand()
	hand5.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	hand5.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
	assert.Equal(t, domain.BJSuggestDouble, domain.GetBasicStrategyAction(hand5, domain.NewCard(domain.CardDesignClover, 6, false), false))
	assert.Equal(t, domain.BJSuggestHit, domain.GetBasicStrategyAction(hand5, domain.NewCard(domain.CardDesignClover, 10, false), false))

	// 2,2 split vs 2-7, hit vs 8-A
	hand2 := domain.NewBlackJackHand()
	hand2.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
	hand2.AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	assert.Equal(t, domain.BJSuggestSplit, domain.GetBasicStrategyAction(hand2, domain.NewCard(domain.CardDesignClover, 3, false), false))
	assert.Equal(t, domain.BJSuggestHit, domain.GetBasicStrategyAction(hand2, domain.NewCard(domain.CardDesignClover, 8, false), false))

	// 3,3 split vs 2-7, hit vs 8-A
	hand3 := domain.NewBlackJackHand()
	hand3.AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
	hand3.AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
	assert.Equal(t, domain.BJSuggestSplit, domain.GetBasicStrategyAction(hand3, domain.NewCard(domain.CardDesignClover, 7, false), false))
	assert.Equal(t, domain.BJSuggestHit, domain.GetBasicStrategyAction(hand3, domain.NewCard(domain.CardDesignClover, 9, false), false))

	// 4,4 split vs 5-6, hit otherwise
	hand4 := domain.NewBlackJackHand()
	hand4.AddCard(domain.NewCard(domain.CardDesignSpade, 4, false))
	hand4.AddCard(domain.NewCard(domain.CardDesignHeart, 4, false))
	assert.Equal(t, domain.BJSuggestSplit, domain.GetBasicStrategyAction(hand4, domain.NewCard(domain.CardDesignClover, 5, false), false))
	assert.Equal(t, domain.BJSuggestHit, domain.GetBasicStrategyAction(hand4, domain.NewCard(domain.CardDesignClover, 2, false), false))

	// 6,6 split vs 2-6, hit vs 7-A
	hand6 := domain.NewBlackJackHand()
	hand6.AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
	hand6.AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
	assert.Equal(t, domain.BJSuggestSplit, domain.GetBasicStrategyAction(hand6, domain.NewCard(domain.CardDesignClover, 2, false), false))
	assert.Equal(t, domain.BJSuggestHit, domain.GetBasicStrategyAction(hand6, domain.NewCard(domain.CardDesignClover, 7, false), false))

	// 7,7 split vs 2-7, hit vs 8-A
	hand7 := domain.NewBlackJackHand()
	hand7.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
	hand7.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
	assert.Equal(t, domain.BJSuggestSplit, domain.GetBasicStrategyAction(hand7, domain.NewCard(domain.CardDesignClover, 7, false), false))
	assert.Equal(t, domain.BJSuggestHit, domain.GetBasicStrategyAction(hand7, domain.NewCard(domain.CardDesignClover, 8, false), false))

	// 9,9 split vs 2-9 (not 7,10,A), stand vs 7,10,A
	hand9 := domain.NewBlackJackHand()
	hand9.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
	hand9.AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
	assert.Equal(t, domain.BJSuggestSplit, domain.GetBasicStrategyAction(hand9, domain.NewCard(domain.CardDesignClover, 2, false), false))
	assert.Equal(t, domain.BJSuggestStand, domain.GetBasicStrategyAction(hand9, domain.NewCard(domain.CardDesignClover, 7, false), false))
	assert.Equal(t, domain.BJSuggestStand, domain.GetBasicStrategyAction(hand9, domain.NewCard(domain.CardDesignClover, 10, false), false))
	assert.Equal(t, domain.BJSuggestStand, domain.GetBasicStrategyAction(hand9, domain.NewCard(domain.CardDesignClover, 1, false), false))

	// J,J pair (bjValue 10 = index 9) stands
	handJJ := domain.NewBlackJackHand()
	handJJ.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
	handJJ.AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))
	assert.Equal(t, domain.BJSuggestStand, domain.GetBasicStrategyAction(handJJ, domain.NewCard(domain.CardDesignClover, 5, false), false))
}

func TestGetBasicStrategyAction_SoftHands(t *testing.T) {
	mkSoft := func(aceVal, otherVal int) *domain.BlackJackHand {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false)) // Ace
		h.AddCard(domain.NewCard(domain.CardDesignHeart, otherVal, false))
		return h
	}

	// Soft 13 (A+2): double vs 5-6, hit otherwise
	s13 := mkSoft(1, 2)
	assert.Equal(t, domain.BJSuggestDouble, domain.GetBasicStrategyAction(s13, domain.NewCard(domain.CardDesignClover, 5, false), false))
	assert.Equal(t, domain.BJSuggestHit, domain.GetBasicStrategyAction(s13, domain.NewCard(domain.CardDesignClover, 2, false), false))

	// Soft 14 (A+3): double vs 5-6
	s14 := mkSoft(1, 3)
	assert.Equal(t, domain.BJSuggestDouble, domain.GetBasicStrategyAction(s14, domain.NewCard(domain.CardDesignClover, 6, false), false))
	assert.Equal(t, domain.BJSuggestHit, domain.GetBasicStrategyAction(s14, domain.NewCard(domain.CardDesignClover, 7, false), false))

	// Soft 15 (A+4): double vs 4-6
	s15 := mkSoft(1, 4)
	assert.Equal(t, domain.BJSuggestDouble, domain.GetBasicStrategyAction(s15, domain.NewCard(domain.CardDesignClover, 4, false), false))
	assert.Equal(t, domain.BJSuggestHit, domain.GetBasicStrategyAction(s15, domain.NewCard(domain.CardDesignClover, 3, false), false))

	// Soft 16 (A+5): double vs 4-6
	s16 := mkSoft(1, 5)
	assert.Equal(t, domain.BJSuggestDouble, domain.GetBasicStrategyAction(s16, domain.NewCard(domain.CardDesignClover, 4, false), false))
	assert.Equal(t, domain.BJSuggestHit, domain.GetBasicStrategyAction(s16, domain.NewCard(domain.CardDesignClover, 7, false), false))

	// Soft 17 (A+6): double vs 3-6, hit vs 2 and 7+
	s17 := mkSoft(1, 6)
	assert.Equal(t, domain.BJSuggestDouble, domain.GetBasicStrategyAction(s17, domain.NewCard(domain.CardDesignClover, 3, false), false))
	assert.Equal(t, domain.BJSuggestHit, domain.GetBasicStrategyAction(s17, domain.NewCard(domain.CardDesignClover, 2, false), false))
	assert.Equal(t, domain.BJSuggestHit, domain.GetBasicStrategyAction(s17, domain.NewCard(domain.CardDesignClover, 7, false), false))

	// Soft 18 (A+7): Ds (double-else-stand) vs 2-6, stand vs 7-8, hit vs 9-A
	s18 := mkSoft(1, 7)
	assert.Equal(t, domain.BJSuggestDoubleStand, domain.GetBasicStrategyAction(s18, domain.NewCard(domain.CardDesignClover, 2, false), false))
	assert.Equal(t, domain.BJSuggestStand, domain.GetBasicStrategyAction(s18, domain.NewCard(domain.CardDesignClover, 7, false), false))
	assert.Equal(t, domain.BJSuggestHit, domain.GetBasicStrategyAction(s18, domain.NewCard(domain.CardDesignClover, 9, false), false))

	// Soft 19 (A+8): S17 では常にスタンド。6 に対するダブルは H17 のときだけ (#4705)。
	s19 := mkSoft(1, 8)
	assert.Equal(t, domain.BJSuggestStand, domain.GetBasicStrategyAction(s19, domain.NewCard(domain.CardDesignClover, 6, false), false))
	assert.Equal(t, domain.BJSuggestStand, domain.GetBasicStrategyAction(s19, domain.NewCard(domain.CardDesignClover, 7, false), false))

	// Soft 20 (A+9): always stand
	s20 := mkSoft(1, 9)
	assert.Equal(t, domain.BJSuggestStand, domain.GetBasicStrategyAction(s20, domain.NewCard(domain.CardDesignClover, 5, false), false))
	assert.Equal(t, domain.BJSuggestStand, domain.GetBasicStrategyAction(s20, domain.NewCard(domain.CardDesignClover, 1, false), false))
}

func TestGetBasicStrategyAction_HardHands(t *testing.T) {
	mkHard := func(v1, v2 int) *domain.BlackJackHand {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, v1, false))
		h.AddCard(domain.NewCard(domain.CardDesignHeart, v2, false))
		return h
	}

	// Hard <=8: always hit
	h8 := mkHard(3, 5)
	assert.Equal(t, domain.BJSuggestHit, domain.GetBasicStrategyAction(h8, domain.NewCard(domain.CardDesignClover, 6, false), false))

	// Hard 9: double vs 3-6, hit otherwise
	h9 := mkHard(5, 4)
	assert.Equal(t, domain.BJSuggestDouble, domain.GetBasicStrategyAction(h9, domain.NewCard(domain.CardDesignClover, 3, false), false))
	assert.Equal(t, domain.BJSuggestHit, domain.GetBasicStrategyAction(h9, domain.NewCard(domain.CardDesignClover, 2, false), false))
	assert.Equal(t, domain.BJSuggestHit, domain.GetBasicStrategyAction(h9, domain.NewCard(domain.CardDesignClover, 7, false), false))

	// Hard 10: double vs 2-9, hit vs 10/A
	h10 := mkHard(6, 4)
	assert.Equal(t, domain.BJSuggestDouble, domain.GetBasicStrategyAction(h10, domain.NewCard(domain.CardDesignClover, 9, false), false))
	assert.Equal(t, domain.BJSuggestHit, domain.GetBasicStrategyAction(h10, domain.NewCard(domain.CardDesignClover, 10, false), false))
	assert.Equal(t, domain.BJSuggestHit, domain.GetBasicStrategyAction(h10, domain.NewCard(domain.CardDesignClover, 1, false), false))

	// Hard 11: double vs 2-10, hit vs A
	h11 := mkHard(7, 4)
	assert.Equal(t, domain.BJSuggestDouble, domain.GetBasicStrategyAction(h11, domain.NewCard(domain.CardDesignClover, 10, false), false))
	assert.Equal(t, domain.BJSuggestHit, domain.GetBasicStrategyAction(h11, domain.NewCard(domain.CardDesignClover, 1, false), false))

	// Hard 12: stand vs 4-6, hit otherwise
	h12 := mkHard(8, 4)
	assert.Equal(t, domain.BJSuggestStand, domain.GetBasicStrategyAction(h12, domain.NewCard(domain.CardDesignClover, 4, false), false))
	assert.Equal(t, domain.BJSuggestHit, domain.GetBasicStrategyAction(h12, domain.NewCard(domain.CardDesignClover, 2, false), false))
	assert.Equal(t, domain.BJSuggestHit, domain.GetBasicStrategyAction(h12, domain.NewCard(domain.CardDesignClover, 7, false), false))

	// Hard 13-14: stand vs 2-6, hit otherwise
	h13 := mkHard(9, 4)
	assert.Equal(t, domain.BJSuggestStand, domain.GetBasicStrategyAction(h13, domain.NewCard(domain.CardDesignClover, 2, false), false))
	assert.Equal(t, domain.BJSuggestHit, domain.GetBasicStrategyAction(h13, domain.NewCard(domain.CardDesignClover, 7, false), false))

	h14 := mkHard(8, 6)
	assert.Equal(t, domain.BJSuggestStand, domain.GetBasicStrategyAction(h14, domain.NewCard(domain.CardDesignClover, 2, false), false))
	assert.Equal(t, domain.BJSuggestHit, domain.GetBasicStrategyAction(h14, domain.NewCard(domain.CardDesignClover, 7, false), false))

	// Hard 15: stand vs 2-6, surrender vs 10, hit vs 7-9 and A
	h15 := mkHard(8, 7)
	assert.Equal(t, domain.BJSuggestStand, domain.GetBasicStrategyAction(h15, domain.NewCard(domain.CardDesignClover, 2, false), false))
	assert.Equal(t, domain.BJSuggestSurrender, domain.GetBasicStrategyAction(h15, domain.NewCard(domain.CardDesignClover, 10, false), false))
	assert.Equal(t, domain.BJSuggestHit, domain.GetBasicStrategyAction(h15, domain.NewCard(domain.CardDesignClover, 7, false), false))
	assert.Equal(t, domain.BJSuggestHit, domain.GetBasicStrategyAction(h15, domain.NewCard(domain.CardDesignClover, 1, false), false))

	// Hard 16: stand vs 2-6, surrender vs 9/10/A, hit vs 7-8
	h16 := mkHard(9, 7)
	assert.Equal(t, domain.BJSuggestStand, domain.GetBasicStrategyAction(h16, domain.NewCard(domain.CardDesignClover, 2, false), false))
	assert.Equal(t, domain.BJSuggestSurrender, domain.GetBasicStrategyAction(h16, domain.NewCard(domain.CardDesignClover, 9, false), false))
	assert.Equal(t, domain.BJSuggestSurrender, domain.GetBasicStrategyAction(h16, domain.NewCard(domain.CardDesignClover, 10, false), false))
	assert.Equal(t, domain.BJSuggestSurrender, domain.GetBasicStrategyAction(h16, domain.NewCard(domain.CardDesignClover, 1, false), false))
	assert.Equal(t, domain.BJSuggestHit, domain.GetBasicStrategyAction(h16, domain.NewCard(domain.CardDesignClover, 7, false), false))

	// Hard 17+: always stand
	h17 := mkHard(10, 7)
	assert.Equal(t, domain.BJSuggestStand, domain.GetBasicStrategyAction(h17, domain.NewCard(domain.CardDesignClover, 1, false), false))

	// Hard 20 (clamped to 17 row): always stand
	h20 := mkHard(10, 10)
	assert.Equal(t, domain.BJSuggestStand, domain.GetBasicStrategyAction(h20, domain.NewCard(domain.CardDesignClover, 6, false), false))

	// Hard 3-card hand (not a pair, not soft): hard 7 (clamped to 5 in table -> hit)
	h5 := domain.NewBlackJackHand()
	h5.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
	h5.AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
	h5.AddCard(domain.NewCard(domain.CardDesignClover, 2, false)) // score = 7 (hard, no ace)
	assert.Equal(t, domain.BJSuggestHit, domain.GetBasicStrategyAction(h5, domain.NewCard(domain.CardDesignDiamond, 6, false), false))
}

func TestGetBasicStrategyAction_DealerUpcardIndexes(t *testing.T) {
	// Test all dealer upcard scenarios to ensure full table coverage
	mkHard16 := func() *domain.BlackJackHand {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		h.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		return h
	}
	// Dealer 2-9: stand for hard 16 vs 2-6, hit vs 7-8
	expected := map[int]domain.BJSuggestedAction{
		2: domain.BJSuggestStand,
		3: domain.BJSuggestStand,
		4: domain.BJSuggestStand,
		5: domain.BJSuggestStand,
		6: domain.BJSuggestStand,
		7: domain.BJSuggestHit,
		8: domain.BJSuggestHit,
		9: domain.BJSuggestSurrender,
	}
	for dealerVal, want := range expected {
		h := mkHard16()
		upcard := domain.NewCard(domain.CardDesignClover, dealerVal, false)
		assert.Equal(t, want, domain.GetBasicStrategyAction(h, upcard, false), "hard 16 vs %d", dealerVal)
	}
	// J,Q,K -> same as 10
	for _, faceVal := range []int{10, 11, 12, 13} {
		h := mkHard16()
		upcard := domain.NewCard(domain.CardDesignClover, faceVal, false)
		assert.Equal(t, domain.BJSuggestSurrender, domain.GetBasicStrategyAction(h, upcard, false), "hard 16 vs face %d", faceVal)
	}
}

func TestNewTrumpCardsWithDecks(t *testing.T) {
	t.Run("single deck has 52 cards", func(t *testing.T) {
		tc := domain.NewTrumpCardsWithDecks(1, 0)
		count := 0
		for tc.DrawCard() != nil {
			count++
		}
		assert.Equal(t, 52, count)
	})
	t.Run("two decks has 104 cards", func(t *testing.T) {
		tc := domain.NewTrumpCardsWithDecks(2, 0)
		count := 0
		for tc.DrawCard() != nil {
			count++
		}
		assert.Equal(t, 104, count)
	})
	t.Run("six decks has 312 cards", func(t *testing.T) {
		tc := domain.NewTrumpCardsWithDecks(6, 0)
		count := 0
		for tc.DrawCard() != nil {
			count++
		}
		assert.Equal(t, 312, count)
	})
	t.Run("NewTrumpCards delegates to NewTrumpCardsWithDecks(1,...)", func(t *testing.T) {
		tc1 := domain.NewTrumpCards(0)
		tc2 := domain.NewTrumpCardsWithDecks(1, 0)
		count1, count2 := 0, 0
		for tc1.DrawCard() != nil {
			count1++
		}
		for tc2.DrawCard() != nil {
			count2++
		}
		assert.Equal(t, count2, count1)
	})
}

func TestGetBasicStrategyAction_EdgeCases(t *testing.T) {
	// hasSoftAce v >= 10 branch: A + face card (King=13, bjValue=10)
	// Not a pair (bjValue(A)=1 != bjValue(K)=10), so falls to soft hand check
	// A+K = soft 21, clamped to soft 20 (idx=7) → Stand
	t.Run("soft A+King is soft 21 clamped to soft20 stand", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))  // Ace
		h.AddCard(domain.NewCard(domain.CardDesignHeart, 13, false)) // King
		upcard := domain.NewCard(domain.CardDesignClover, 6, false)
		assert.Equal(t, domain.BJSuggestStand, domain.GetBasicStrategyAction(h, upcard, false))
	})

	// softStrategy idx < 0: A+A = pair (goes to pairStrategy), so use 3 aces
	// Actually A+A is a pair, so try single Ace or different combo
	// Two non-pair aces: impossible (A bjValue = 1, A bjValue = 1 → pair)
	// Use A+Q+K = 1+12+13: all bjValue: A=1,Q=10,K=10, not simple pair detection (3 cards)
	// hasSoftAce: 11+10+10=31, reduce→21, aces=1 → (true, 21), idx=8>7 → clamped to 7
	t.Run("soft 3-card A+Q+K = soft 21 clamped to soft20 stand", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		h.AddCard(domain.NewCard(domain.CardDesignHeart, 12, false))  // Queen
		h.AddCard(domain.NewCard(domain.CardDesignClover, 13, false)) // King
		upcard := domain.NewCard(domain.CardDesignDiamond, 5, false)
		assert.Equal(t, domain.BJSuggestStand, domain.GetBasicStrategyAction(h, upcard, false))
	})

	// softStrategy idx < 0: A+9+A = 11+9+11=31, reduce→21, aces=1 (soft 21, idx=8>7)
	// Different from above... let me find idx < 0 case
	// For idx < 0, softTotal < 13: A alone (1 card)
	t.Run("single Ace hand: soft with idx<0 clamped to soft13 hit", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false)) // Ace alone
		// softTotal = 11, idx = 11-13 = -2 → clamped to 0 (soft13 row) → H vs dealer 6
		upcard := domain.NewCard(domain.CardDesignClover, 2, false)
		// soft13 vs 2 → H (hit)
		assert.Equal(t, domain.BJSuggestHit, domain.GetBasicStrategyAction(h, upcard, false))
	})

	// hardStrategy clamped < 5: single card hand with value 3
	t.Run("single card hard 3: clamped to 5 → hit", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 3, false)) // value 3, hard
		upcard := domain.NewCard(domain.CardDesignClover, 6, false)
		assert.Equal(t, domain.BJSuggestHit, domain.GetBasicStrategyAction(h, upcard, false))
	})

	// hardStrategy clamped > 17: non-pair non-soft hand with score > 17
	// 8+10 = 18 (bjValue 8 != bjValue 10, not a pair; no ace → hard)
	t.Run("hard 18 (8+10): clamped to 17 → stand", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		h.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
		upcard := domain.NewCard(domain.CardDesignClover, 1, false)
		assert.Equal(t, domain.BJSuggestStand, domain.GetBasicStrategyAction(h, upcard, false))
	})
}

func TestGetBasicStrategyAction_H17Overrides(t *testing.T) {
	aceUpcard := domain.NewCard(domain.CardDesignClover, 1, false)

	t.Run("hard 15 vs A: S17=Hit, H17=Surrender", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		h.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false)) // hard 15

		// S17: hard 15 vs A → Hit
		assert.Equal(t, domain.BJSuggestHit, domain.GetBasicStrategyAction(h, aceUpcard, false),
			"S17: hard 15 vs A should be Hit")
		// H17: hard 15 vs A → Surrender
		assert.Equal(t, domain.BJSuggestSurrender, domain.GetBasicStrategyAction(h, aceUpcard, true),
			"H17: hard 15 vs A should be Surrender")
	})

	t.Run("hard 17 vs A: S17=Stand, H17=Surrender", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		h.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false)) // hard 17

		// S17: hard 17 vs A → Stand
		assert.Equal(t, domain.BJSuggestStand, domain.GetBasicStrategyAction(h, aceUpcard, false),
			"S17: hard 17 vs A should be Stand")
		// H17: hard 17 vs A → Surrender
		assert.Equal(t, domain.BJSuggestSurrender, domain.GetBasicStrategyAction(h, aceUpcard, true),
			"H17: hard 17 vs A should be Surrender")
	})

	// **両側が同じ override テストは、override が死んでいても通る。**
	// 以前はここが S17=Ds / H17=Ds で、基本表が既に Ds だったため
	// softH17Override は何も変えていなかった (#4705)。S17 は S、H17 は Ds が正しい。
	t.Run("soft 19 vs 6: S17=Stand, H17=DoubleStand", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		h.AddCard(domain.NewCard(domain.CardDesignHeart, 8, false)) // soft 19
		upcard6 := domain.NewCard(domain.CardDesignClover, 6, false)

		assert.Equal(t, domain.BJSuggestStand, domain.GetBasicStrategyAction(h, upcard6, false),
			"S17: soft 19 vs 6 should be Stand")
		assert.Equal(t, domain.BJSuggestDoubleStand, domain.GetBasicStrategyAction(h, upcard6, true),
			"H17: soft 19 vs 6 should be DoubleStand")
	})

	t.Run("hard 11 vs A: S17=Hit, H17=Double", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
		h.AddCard(domain.NewCard(domain.CardDesignHeart, 4, false)) // hard 11

		// S17: hard 11 vs A → Hit
		assert.Equal(t, domain.BJSuggestHit, domain.GetBasicStrategyAction(h, aceUpcard, false),
			"S17: hard 11 vs A should be Hit")
		// H17: hard 11 vs A → Double
		assert.Equal(t, domain.BJSuggestDouble, domain.GetBasicStrategyAction(h, aceUpcard, true),
			"H17: hard 11 vs A should be Double")
	})

	t.Run("H17 no change for non-Ace dealer", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		h.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false)) // hard 15
		upcard10 := domain.NewCard(domain.CardDesignClover, 10, false)

		// H17 override only applies to Ace (di=9)
		// hard 15 vs 10 → Surrender (same for S17 and H17)
		assert.Equal(t, domain.BJSuggestSurrender, domain.GetBasicStrategyAction(h, upcard10, false))
		assert.Equal(t, domain.BJSuggestSurrender, domain.GetBasicStrategyAction(h, upcard10, true))
	})

	t.Run("soft H17 override only for soft 19 vs 6", func(t *testing.T) {
		// Soft 18 vs 6: S17=DoubleStand, H17=DoubleStand (no change)
		h18 := domain.NewBlackJackHand()
		h18.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		h18.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false)) // soft 18
		upcard6 := domain.NewCard(domain.CardDesignClover, 6, false)

		assert.Equal(t, domain.BJSuggestDoubleStand, domain.GetBasicStrategyAction(h18, upcard6, false))
		assert.Equal(t, domain.BJSuggestDoubleStand, domain.GetBasicStrategyAction(h18, upcard6, true))

		// Soft 19 vs 5: S17=Stand (table: S,S,S,S,Ds → di=3 is S), H17=Stand (no override for di!=4)
		h19 := domain.NewBlackJackHand()
		h19.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		h19.AddCard(domain.NewCard(domain.CardDesignHeart, 8, false)) // soft 19
		upcard5 := domain.NewCard(domain.CardDesignClover, 5, false)

		assert.Equal(t, domain.BJSuggestStand, domain.GetBasicStrategyAction(h19, upcard5, false),
			"S17: soft 19 vs 5 should be Stand")
		assert.Equal(t, domain.BJSuggestStand, domain.GetBasicStrategyAction(h19, upcard5, true),
			"H17: soft 19 vs 5 should still be Stand (override only for di=4)")
	})

	t.Run("hard H17 default case returns s17Action", func(t *testing.T) {
		// Hard 12 vs A: S17=Hit, H17=Hit (default case, no override)
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		h.AddCard(domain.NewCard(domain.CardDesignHeart, 4, false)) // hard 12

		assert.Equal(t, domain.BJSuggestHit, domain.GetBasicStrategyAction(h, aceUpcard, false))
		assert.Equal(t, domain.BJSuggestHit, domain.GetBasicStrategyAction(h, aceUpcard, true),
			"H17: hard 12 vs A should still be Hit (default case)")
	})
}
