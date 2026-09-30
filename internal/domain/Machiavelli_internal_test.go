//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func machiavelliInternalCard(design, value int) *Card {
	return NewCard(design, value, false)
}

func TestMachiavelli_IsSet(t *testing.T) {
	// 有効: 同ランク別スート 3 枚
	assert.True(t, machiavelliIsValidMeld([]*Card{
		machiavelliInternalCard(CardDesignSpade, 9),
		machiavelliInternalCard(CardDesignHeart, 9),
		machiavelliInternalCard(CardDesignClover, 9),
	}))
	// 有効: 4 枚（全スート）
	assert.True(t, machiavelliIsValidMeld([]*Card{
		machiavelliInternalCard(CardDesignSpade, 9),
		machiavelliInternalCard(CardDesignHeart, 9),
		machiavelliInternalCard(CardDesignClover, 9),
		machiavelliInternalCard(CardDesignDiamond, 9),
	}))
	// 無効: スート重複
	assert.False(t, machiavelliIsValidMeld([]*Card{
		machiavelliInternalCard(CardDesignSpade, 9),
		machiavelliInternalCard(CardDesignSpade, 9),
		machiavelliInternalCard(CardDesignHeart, 9),
	}))
	// 無効: ランク不一致（かつラン不成立）
	assert.False(t, machiavelliIsValidMeld([]*Card{
		machiavelliInternalCard(CardDesignSpade, 9),
		machiavelliInternalCard(CardDesignHeart, 9),
		machiavelliInternalCard(CardDesignClover, 5),
	}))
	// 無効: 2 枚（枚数不足）
	assert.False(t, machiavelliIsValidMeld([]*Card{
		machiavelliInternalCard(CardDesignSpade, 9),
		machiavelliInternalCard(CardDesignHeart, 9),
	}))
	// 無効: ジョーカーは使えない
	assert.False(t, machiavelliIsValidMeld([]*Card{
		machiavelliInternalCard(CardDesignJoker, 0),
		machiavelliInternalCard(CardDesignHeart, 9),
		machiavelliInternalCard(CardDesignClover, 9),
	}))
}

func TestMachiavelli_IsRun(t *testing.T) {
	// 有効: 同スート連続 3 枚
	assert.True(t, machiavelliIsValidMeld([]*Card{
		machiavelliInternalCard(CardDesignSpade, 4),
		machiavelliInternalCard(CardDesignSpade, 5),
		machiavelliInternalCard(CardDesignSpade, 6),
	}))
	// 有効: Ace-low (A-2-3)
	assert.True(t, machiavelliIsValidMeld([]*Card{
		machiavelliInternalCard(CardDesignHeart, 1),
		machiavelliInternalCard(CardDesignHeart, 2),
		machiavelliInternalCard(CardDesignHeart, 3),
	}))
	// 有効: Ace-high (Q-K-A)
	assert.True(t, machiavelliIsValidMeld([]*Card{
		machiavelliInternalCard(CardDesignDiamond, 12),
		machiavelliInternalCard(CardDesignDiamond, 13),
		machiavelliInternalCard(CardDesignDiamond, 1),
	}))
	// 無効: ラップアラウンド (K-A-2)
	assert.False(t, machiavelliIsValidMeld([]*Card{
		machiavelliInternalCard(CardDesignDiamond, 13),
		machiavelliInternalCard(CardDesignDiamond, 1),
		machiavelliInternalCard(CardDesignDiamond, 2),
	}))
	// 無効: スート不一致
	assert.False(t, machiavelliIsValidMeld([]*Card{
		machiavelliInternalCard(CardDesignSpade, 4),
		machiavelliInternalCard(CardDesignHeart, 5),
		machiavelliInternalCard(CardDesignSpade, 6),
	}))
	// 無効: 連続でない
	assert.False(t, machiavelliIsValidMeld([]*Card{
		machiavelliInternalCard(CardDesignSpade, 4),
		machiavelliInternalCard(CardDesignSpade, 5),
		machiavelliInternalCard(CardDesignSpade, 7),
	}))
	// 無効: 値重複
	assert.False(t, machiavelliIsValidMeld([]*Card{
		machiavelliInternalCard(CardDesignSpade, 4),
		machiavelliInternalCard(CardDesignSpade, 4),
		machiavelliInternalCard(CardDesignSpade, 5),
	}))
}

func TestMachiavelli_CardPoints(t *testing.T) {
	assert.Equal(t, 1, machiavelliCardPoints(machiavelliInternalCard(CardDesignSpade, 1)))   // Ace
	assert.Equal(t, 7, machiavelliCardPoints(machiavelliInternalCard(CardDesignSpade, 7)))   // pip
	assert.Equal(t, 10, machiavelliCardPoints(machiavelliInternalCard(CardDesignSpade, 10))) // Ten
	assert.Equal(t, 10, machiavelliCardPoints(machiavelliInternalCard(CardDesignSpade, 13))) // King
	assert.Equal(t, 0, machiavelliCardPoints(nil))
}
