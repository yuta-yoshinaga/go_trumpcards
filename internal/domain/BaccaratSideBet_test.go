package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBacSideBetResult_BetTypeName(t *testing.T) {
	t.Run("player pair", func(t *testing.T) {
		r := &BacSideBetResult{BetType: BacSideBetPlayerPair}
		assert.Equal(t, "Player Pair", r.BetTypeName())
	})
	t.Run("banker pair", func(t *testing.T) {
		r := &BacSideBetResult{BetType: BacSideBetBankerPair}
		assert.Equal(t, "Banker Pair", r.BetTypeName())
	})
	t.Run("unknown", func(t *testing.T) {
		r := &BacSideBetResult{BetType: 99}
		assert.Equal(t, "Unknown", r.BetTypeName())
	})
}

func TestEvaluateBaccaratPair(t *testing.T) {
	t.Run("pair match", func(t *testing.T) {
		c1 := NewCard(CardDesignSpade, 5, false)
		c2 := NewCard(CardDesignHeart, 5, false)
		resultType, resultName := EvaluateBaccaratPair(c1, c2)
		assert.Equal(t, BacPairMatch, resultType)
		assert.Equal(t, "Pair", resultName)
	})
	t.Run("no pair", func(t *testing.T) {
		c1 := NewCard(CardDesignSpade, 5, false)
		c2 := NewCard(CardDesignHeart, 6, false)
		resultType, resultName := EvaluateBaccaratPair(c1, c2)
		assert.Equal(t, BacPairNone, resultType)
		assert.Equal(t, "", resultName)
	})
}

func TestBacPairPayout(t *testing.T) {
	t.Run("match pays 11x", func(t *testing.T) {
		assert.Equal(t, BacPairPayoutRate, BacPairPayout(BacPairMatch))
	})
	t.Run("none pays 0", func(t *testing.T) {
		assert.Equal(t, 0, BacPairPayout(BacPairNone))
	})
}
