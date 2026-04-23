package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTrumpCardsSpanish21(t *testing.T) {
	t.Run("single deck has 48 cards (no 10s)", func(t *testing.T) {
		tc := NewTrumpCardsSpanish21(1)
		assert.Equal(t, SpanishDeckCardCount, tc.GetTotalCount())
		assert.Equal(t, SpanishDeckCardCount, tc.GetRemainingCount())

		// 全カードを引いて10が含まれていないことを確認
		seen := map[int]int{}
		for i := 0; i < SpanishDeckCardCount; i++ {
			c := tc.DrawCard()
			assert.NotNil(t, c)
			seen[c.GetValue()]++
			assert.NotEqual(t, 10, c.GetValue(), "Spanish 21 deck must not contain 10s")
		}
		// 各ランクが4枚ずつ存在 (10を除く全ランク)
		for _, v := range SpanishDeckValues {
			assert.Equal(t, 4, seen[v], "rank %d should appear 4 times", v)
		}
	})

	t.Run("multi deck = 48 * deckCount", func(t *testing.T) {
		tc := NewTrumpCardsSpanish21(6)
		assert.Equal(t, 6*SpanishDeckCardCount, tc.GetTotalCount())
	})

	t.Run("zero or negative deckCount falls back to 1", func(t *testing.T) {
		tc := NewTrumpCardsSpanish21(0)
		assert.Equal(t, SpanishDeckCardCount, tc.GetTotalCount())
		tc = NewTrumpCardsSpanish21(-3)
		assert.Equal(t, SpanishDeckCardCount, tc.GetTotalCount())
	})
}

func TestSpanish21Variant(t *testing.T) {
	v := Spanish21Variant()
	assert.Equal(t, BJVariantSpanish21, v.Name)
	assert.NotNil(t, v.DeckBuilder)
	assert.True(t, v.Player21AlwaysWins)
	assert.True(t, v.PlayerBJBeatsDealerBJ)
	assert.NotNil(t, v.BonusEval)
}

func TestResolveBlackJackVariant(t *testing.T) {
	t.Run("standard returns nil", func(t *testing.T) {
		assert.Nil(t, ResolveBlackJackVariant(BJVariantStandard))
		assert.Nil(t, ResolveBlackJackVariant(BlackJackVariantName("unknown")))
	})

	t.Run("spanish21 returns variant", func(t *testing.T) {
		v := ResolveBlackJackVariant(BJVariantSpanish21)
		assert.NotNil(t, v)
		assert.Equal(t, BJVariantSpanish21, v.Name)
	})
}

func TestBlackJackConfigValidateVariant(t *testing.T) {
	t.Run("standard variant accepted", func(t *testing.T) {
		c := DefaultBlackJackConfig()
		c.Variant = BJVariantStandard
		assert.NoError(t, c.Validate())
	})

	t.Run("spanish21 variant accepted", func(t *testing.T) {
		c := DefaultBlackJackConfig()
		c.Variant = BJVariantSpanish21
		assert.NoError(t, c.Validate())
	})

	t.Run("unknown variant rejected", func(t *testing.T) {
		c := DefaultBlackJackConfig()
		c.Variant = BlackJackVariantName("bogus")
		err := c.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown blackjack variant")
	})
}

func TestSpanish21BonusEval(t *testing.T) {
	makeHand := func(cards ...*Card) *BlackJackHand {
		h := NewBlackJackHand()
		for _, c := range cards {
			h.AddCard(c)
		}
		return h
	}

	tests := []struct {
		name          string
		hand          *BlackJackHand
		expectNil     bool
		expectKey     string
		expectMultNum int
		expectMultDen int
	}{
		{
			name:      "non-21 hand returns nil",
			hand:      makeHand(NewCard(CardDesignSpade, 9, false), NewCard(CardDesignClover, 9, false)),
			expectNil: true,
		},
		{
			name:      "natural 21 (2 cards) returns nil — handled by BJ payout",
			hand:      makeHand(NewCard(CardDesignSpade, 1, false), NewCard(CardDesignClover, 13, false)),
			expectNil: true,
		},
		{
			name:      "3-card 21 without trio returns nil (e.g., 7-5-9)",
			hand:      makeHand(NewCard(CardDesignSpade, 7, false), NewCard(CardDesignClover, 5, false), NewCard(CardDesignHeart, 9, false)),
			expectNil: true,
		},
		{
			name:          "5-card 21 → 3:2",
			hand:          makeHand(NewCard(CardDesignSpade, 2, false), NewCard(CardDesignClover, 3, false), NewCard(CardDesignHeart, 4, false), NewCard(CardDesignDiamond, 5, false), NewCard(CardDesignSpade, 7, false)),
			expectKey:     "spanish21.bonus.fivecard21",
			expectMultNum: 3, expectMultDen: 2,
		},
		{
			name:          "6-card 21 → 2:1",
			hand:          makeHand(NewCard(CardDesignSpade, 2, false), NewCard(CardDesignClover, 2, false), NewCard(CardDesignHeart, 3, false), NewCard(CardDesignDiamond, 4, false), NewCard(CardDesignSpade, 5, false), NewCard(CardDesignClover, 5, false)),
			expectKey:     "spanish21.bonus.sixcard21",
			expectMultNum: 2, expectMultDen: 1,
		},
		{
			name: "7-card 21 → 3:1",
			hand: makeHand(
				NewCard(CardDesignSpade, 2, false), NewCard(CardDesignClover, 2, false),
				NewCard(CardDesignHeart, 2, false), NewCard(CardDesignDiamond, 2, false),
				NewCard(CardDesignSpade, 3, false), NewCard(CardDesignClover, 3, false),
				NewCard(CardDesignHeart, 7, false),
			),
			expectKey:     "spanish21.bonus.sevencard21",
			expectMultNum: 3, expectMultDen: 1,
		},
		{
			name:          "6-7-8 mixed suits → 3:2",
			hand:          makeHand(NewCard(CardDesignSpade, 6, false), NewCard(CardDesignClover, 7, false), NewCard(CardDesignHeart, 8, false)),
			expectKey:     "spanish21.bonus.678.mixed",
			expectMultNum: 3, expectMultDen: 2,
		},
		{
			name:          "6-7-8 same suit (hearts) → 2:1",
			hand:          makeHand(NewCard(CardDesignHeart, 6, false), NewCard(CardDesignHeart, 7, false), NewCard(CardDesignHeart, 8, false)),
			expectKey:     "spanish21.bonus.678.samesuit",
			expectMultNum: 2, expectMultDen: 1,
		},
		{
			name:          "6-7-8 spades → 3:1",
			hand:          makeHand(NewCard(CardDesignSpade, 6, false), NewCard(CardDesignSpade, 7, false), NewCard(CardDesignSpade, 8, false)),
			expectKey:     "spanish21.bonus.678.spade",
			expectMultNum: 3, expectMultDen: 1,
		},
		{
			name:          "7-7-7 mixed → 3:2",
			hand:          makeHand(NewCard(CardDesignSpade, 7, false), NewCard(CardDesignClover, 7, false), NewCard(CardDesignHeart, 7, false)),
			expectKey:     "spanish21.bonus.777.mixed",
			expectMultNum: 3, expectMultDen: 2,
		},
		{
			name:          "7-7-7 same suit (diamonds) → 2:1",
			hand:          makeHand(NewCard(CardDesignDiamond, 7, false), NewCard(CardDesignDiamond, 7, false), NewCard(CardDesignDiamond, 7, false)),
			expectKey:     "spanish21.bonus.777.samesuit",
			expectMultNum: 2, expectMultDen: 1,
		},
		{
			name:          "7-7-7 spades → 3:1",
			hand:          makeHand(NewCard(CardDesignSpade, 7, false), NewCard(CardDesignSpade, 7, false), NewCard(CardDesignSpade, 7, false)),
			expectKey:     "spanish21.bonus.777.spade",
			expectMultNum: 3, expectMultDen: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := spanish21BonusEval(tt.hand, nil)
			if tt.expectNil {
				assert.Nil(t, got)
				return
			}
			assert.NotNil(t, got)
			assert.Equal(t, tt.expectKey, got.NameKey)
			assert.Equal(t, tt.expectMultNum, got.MultiplierNum)
			assert.Equal(t, tt.expectMultDen, got.MultiplierDen)
		})
	}

	t.Run("nil hand returns nil", func(t *testing.T) {
		assert.Nil(t, spanish21BonusEval(nil, nil))
	})

	t.Run("trio bonuses preferred over multi-card on 3 cards", func(t *testing.T) {
		// 3-card 21 that's also 7-7-7 spades — trio bonus should fire (not nil)
		h := NewBlackJackHand()
		h.AddCard(NewCard(CardDesignSpade, 7, false))
		h.AddCard(NewCard(CardDesignSpade, 7, false))
		h.AddCard(NewCard(CardDesignSpade, 7, false))
		bonus := spanish21BonusEval(h, nil)
		assert.NotNil(t, bonus)
		assert.Equal(t, "spanish21.bonus.777.spade", bonus.NameKey)
	})
}
