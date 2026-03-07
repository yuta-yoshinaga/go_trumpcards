package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// SetTableCards テーブルカード設定（テスト用）
func (d *Doubt) SetTableCards(cards []*Card) { d.tableCards = cards }

// SetTurnCounter ターンカウンター設定（テスト用）
func (d *Doubt) SetTurnCounter(v int) { d.turnCounter = v }

// TestDecideCpuDoubters_NilLastAction lastAction が nil のとき早期リターンする
func TestDecideCpuDoubters_NilLastAction(t *testing.T) {
	players := []*DoubtPlayer{
		NewDoubtPlayer(true),
		NewDoubtPlayer(false),
		NewDoubtPlayer(false),
		NewDoubtPlayer(false),
	}
	d := NewDoubt(NewTrumpCards(0), players)
	// lastAction is nil (initial state)
	d.decideCpuDoubters()
	// Should be nil (early return)
	assert.Nil(t, d.cpuDoubters)
}

// TestMemoryRetentionChance_UnknownLevel 未知のレベルは Normal と同じ保持率を返す
func TestMemoryRetentionChance_UnknownLevel(t *testing.T) {
	unknown := DoubtMemoryLevel(99)
	assert.Equal(t, retentionChanceNormal, memoryRetentionChance(unknown))
}

// TestCheckLying_NilLastAction lastAction が nil のとき false を返す
func TestCheckLying_NilLastAction(t *testing.T) {
	players := []*DoubtPlayer{
		NewDoubtPlayer(true),
		NewDoubtPlayer(false),
		NewDoubtPlayer(false),
		NewDoubtPlayer(false),
	}
	d := NewDoubt(NewTrumpCards(0), players)
	// lastAction is nil (initial state)
	result := d.checkLying()
	assert.False(t, result)
}

// TestCalcBluffChance 動的ブラフ確率の計算テスト
func TestCalcBluffChance(t *testing.T) {
	makeGame := func() *Doubt {
		players := []*DoubtPlayer{
			NewDoubtPlayer(true),
			NewDoubtPlayer(false),
			NewDoubtPlayer(false),
			NewDoubtPlayer(false),
		}
		return NewDoubt(NewTrumpCards(0), players)
	}

	t.Run("normal case - base chance", func(t *testing.T) {
		d := makeGame()
		chance := d.calcBluffChance(5, 0)
		assert.InDelta(t, bluffChanceBase, chance, 0.001) // 0.4
	})

	t.Run("last card - reduced to bluffChanceLastCard", func(t *testing.T) {
		d := makeGame()
		chance := d.calcBluffChance(1, 0)
		assert.InDelta(t, bluffChanceLastCard, chance, 0.001) // 0.1
	})

	t.Run("handSize=0 - also uses bluffChanceLastCard", func(t *testing.T) {
		d := makeGame()
		chance := d.calcBluffChance(0, 0)
		assert.InDelta(t, bluffChanceLastCard, chance, 0.001)
	})

	t.Run("large table (>=20) - penalty applied", func(t *testing.T) {
		d := makeGame()
		chance := d.calcBluffChance(5, 20)
		assert.InDelta(t, bluffChanceBase-bluffPenaltyLargeTable, chance, 0.001) // 0.4 - 0.15 = 0.25
	})

	t.Run("medium table (>=10, <20) - penalty applied", func(t *testing.T) {
		d := makeGame()
		chance := d.calcBluffChance(5, 10)
		assert.InDelta(t, bluffChanceBase-bluffPenaltyMediumTable, chance, 0.001) // 0.4 - 0.10 = 0.30
	})

	t.Run("table exactly 19 - medium penalty", func(t *testing.T) {
		d := makeGame()
		chance := d.calcBluffChance(5, 19)
		assert.InDelta(t, bluffChanceBase-bluffPenaltyMediumTable, chance, 0.001)
	})

	t.Run("table below 10 - no penalty", func(t *testing.T) {
		d := makeGame()
		chance := d.calcBluffChance(5, 9)
		assert.InDelta(t, bluffChanceBase, chance, 0.001)
	})

	t.Run("very large table still uses large table penalty", func(t *testing.T) {
		d := makeGame()
		chance := d.calcBluffChance(5, 100) // very large table (>=20)
		assert.InDelta(t, bluffChanceBase-bluffPenaltyLargeTable, chance, 0.001)
	})
}

// TestCalcTellChance テル表示確率のテスト
func TestCalcTellChance(t *testing.T) {
	t.Run("easy", func(t *testing.T) {
		assert.InDelta(t, tellChanceEasy, calcTellChance(DoubtMemoryLevelEasy), 0.001)
	})

	t.Run("normal", func(t *testing.T) {
		assert.InDelta(t, tellChanceNormal, calcTellChance(DoubtMemoryLevelNormal), 0.001)
	})

	t.Run("hard", func(t *testing.T) {
		assert.InDelta(t, tellChanceHard, calcTellChance(DoubtMemoryLevelHard), 0.001)
	})

	t.Run("unknown level defaults to normal", func(t *testing.T) {
		assert.InDelta(t, tellChanceNormal, calcTellChance(DoubtMemoryLevel(99)), 0.001)
	})
}

// TestMixedBluffChance 混合ブラフ確率の定数値テスト
func TestMixedBluffChance(t *testing.T) {
	assert.InDelta(t, 0.3, mixedBluffChance, 0.001)
}

// TestSelectMixedCards 混合カード選択ヘルパーのテスト
func TestSelectMixedCards(t *testing.T) {
	t.Run("normal case - has both matching and non-matching", func(t *testing.T) {
		player := NewDoubtPlayer(false)
		player.AddCard(NewCard(CardDesignSpade, 5, false))   // 0: matches
		player.AddCard(NewCard(CardDesignHeart, 5, false))   // 1: matches
		player.AddCard(NewCard(CardDesignClover, 3, false))  // 2: non-matching
		player.AddCard(NewCard(CardDesignDiamond, 7, false)) // 3: non-matching

		played := selectMixedCards(player, 5, 3)
		assert.Len(t, played, 3)

		hasMatch := false
		hasNonMatch := false
		for _, card := range played {
			if card.GetValue() == 5 {
				hasMatch = true
			} else {
				hasNonMatch = true
			}
		}
		assert.True(t, hasMatch, "should contain at least one matching card")
		assert.True(t, hasNonMatch, "should contain at least one non-matching card")
	})

	t.Run("fallback - all cards match claimed value", func(t *testing.T) {
		player := NewDoubtPlayer(false)
		player.AddCard(NewCard(CardDesignSpade, 5, false))
		player.AddCard(NewCard(CardDesignHeart, 5, false))
		player.AddCard(NewCard(CardDesignClover, 5, false))

		played := selectMixedCards(player, 5, 2)
		assert.Len(t, played, 2)
		for _, card := range played {
			assert.Equal(t, 5, card.GetValue())
		}
	})

	t.Run("fallback - no matching cards", func(t *testing.T) {
		player := NewDoubtPlayer(false)
		player.AddCard(NewCard(CardDesignSpade, 3, false))
		player.AddCard(NewCard(CardDesignHeart, 7, false))
		player.AddCard(NewCard(CardDesignClover, 9, false))

		played := selectMixedCards(player, 5, 2)
		assert.Len(t, played, 2)
	})

	t.Run("edge case - only 1 non-matching card", func(t *testing.T) {
		player := NewDoubtPlayer(false)
		player.AddCard(NewCard(CardDesignSpade, 5, false))  // matches
		player.AddCard(NewCard(CardDesignHeart, 5, false))  // matches
		player.AddCard(NewCard(CardDesignClover, 3, false)) // non-matching (only 1)

		played := selectMixedCards(player, 5, 3)
		assert.Len(t, played, 3)

		nonMatchCount := 0
		for _, card := range played {
			if card.GetValue() != 5 {
				nonMatchCount++
			}
		}
		assert.Equal(t, 1, nonMatchCount, "should have exactly 1 non-matching card")
	})
}

// TestMemoryDecayRate 記憶減衰率のテスト
func TestMemoryDecayRate(t *testing.T) {
	t.Run("easy", func(t *testing.T) {
		assert.InDelta(t, decayRateEasy, memoryDecayRate(DoubtMemoryLevelEasy), 0.001)
	})

	t.Run("normal", func(t *testing.T) {
		assert.InDelta(t, decayRateNormal, memoryDecayRate(DoubtMemoryLevelNormal), 0.001)
	})

	t.Run("hard", func(t *testing.T) {
		assert.InDelta(t, decayRateHard, memoryDecayRate(DoubtMemoryLevelHard), 0.001)
	})

	t.Run("unknown level defaults to normal", func(t *testing.T) {
		assert.InDelta(t, decayRateNormal, memoryDecayRate(DoubtMemoryLevel(99)), 0.001)
	})
}
