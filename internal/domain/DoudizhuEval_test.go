//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func dzCard(design, value int) *Card { return NewCard(design, value, false) }
func dzSpade(v int) *Card            { return dzCard(CardDesignSpade, v) }
func dzHeart(v int) *Card            { return dzCard(CardDesignHeart, v) }
func dzDiamond(v int) *Card          { return dzCard(CardDesignDiamond, v) }
func dzClover(v int) *Card           { return dzCard(CardDesignClover, v) }
func dzSmallJoker() *Card            { return dzCard(CardDesignJoker, 1) }
func dzBigJoker() *Card              { return dzCard(CardDesignJoker, 2) }

func TestDoudizhuCardStrength(t *testing.T) {
	tests := []struct {
		name     string
		card     *Card
		expected int
	}{
		{"3", dzSpade(3), 3},
		{"4", dzSpade(4), 4},
		{"10", dzSpade(10), 10},
		{"J", dzSpade(11), 11},
		{"Q", dzSpade(12), 12},
		{"K", dzSpade(13), 13},
		{"A", dzSpade(1), 14},
		{"2", dzSpade(2), 15},
		{"Small Joker", dzSmallJoker(), 16},
		{"Big Joker", dzBigJoker(), 17},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, DoudizhuCardStrength(tt.card))
		})
	}
}

func TestIsBigJoker(t *testing.T) {
	assert.True(t, IsBigJoker(dzBigJoker()))
	assert.False(t, IsBigJoker(dzSmallJoker()))
	assert.False(t, IsBigJoker(dzSpade(1)))
}

func TestIsSmallJoker(t *testing.T) {
	assert.True(t, IsSmallJoker(dzSmallJoker()))
	assert.False(t, IsSmallJoker(dzBigJoker()))
	assert.False(t, IsSmallJoker(dzSpade(1)))
}

func TestClassifyCombo_Single(t *testing.T) {
	combo := DoudizhuClassifyCombo([]*Card{dzSpade(3)})
	assert.NotNil(t, combo)
	assert.Equal(t, DoudizhuComboSingle, combo.Type)
	assert.Equal(t, 3, combo.Rank)
}

func TestClassifyCombo_Pair(t *testing.T) {
	combo := DoudizhuClassifyCombo([]*Card{dzSpade(5), dzHeart(5)})
	assert.NotNil(t, combo)
	assert.Equal(t, DoudizhuComboPair, combo.Type)
	assert.Equal(t, 5, combo.Rank)
}

func TestClassifyCombo_PairInvalid(t *testing.T) {
	combo := DoudizhuClassifyCombo([]*Card{dzSpade(5), dzHeart(6)})
	assert.Nil(t, combo)
}

func TestClassifyCombo_Trio(t *testing.T) {
	combo := DoudizhuClassifyCombo([]*Card{dzSpade(7), dzHeart(7), dzDiamond(7)})
	assert.NotNil(t, combo)
	assert.Equal(t, DoudizhuComboTrio, combo.Type)
	assert.Equal(t, 7, combo.Rank)
}

func TestClassifyCombo_TrioSingle(t *testing.T) {
	combo := DoudizhuClassifyCombo([]*Card{dzSpade(8), dzHeart(8), dzDiamond(8), dzClover(3)})
	assert.NotNil(t, combo)
	assert.Equal(t, DoudizhuComboTrioSingle, combo.Type)
	assert.Equal(t, 8, combo.Rank)
}

func TestClassifyCombo_TrioPair(t *testing.T) {
	combo := DoudizhuClassifyCombo([]*Card{dzSpade(9), dzHeart(9), dzDiamond(9), dzClover(4), dzSpade(4)})
	assert.NotNil(t, combo)
	assert.Equal(t, DoudizhuComboTrioPair, combo.Type)
	assert.Equal(t, 9, combo.Rank)
}

func TestClassifyCombo_TrioPairInvalid_KickerNotPair(t *testing.T) {
	combo := DoudizhuClassifyCombo([]*Card{dzSpade(9), dzHeart(9), dzDiamond(9), dzClover(4), dzSpade(5)})
	assert.Nil(t, combo)
}

func TestClassifyCombo_Bomb(t *testing.T) {
	combo := DoudizhuClassifyCombo([]*Card{dzSpade(10), dzHeart(10), dzDiamond(10), dzClover(10)})
	assert.NotNil(t, combo)
	assert.Equal(t, DoudizhuComboBomb, combo.Type)
	assert.Equal(t, 10, combo.Rank)
}

func TestClassifyCombo_Rocket(t *testing.T) {
	combo := DoudizhuClassifyCombo([]*Card{dzSmallJoker(), dzBigJoker()})
	assert.NotNil(t, combo)
	assert.Equal(t, DoudizhuComboRocket, combo.Type)
}

func TestClassifyCombo_Straight(t *testing.T) {
	combo := DoudizhuClassifyCombo([]*Card{dzSpade(3), dzHeart(4), dzDiamond(5), dzClover(6), dzSpade(7)})
	assert.NotNil(t, combo)
	assert.Equal(t, DoudizhuComboStraight, combo.Type)
	assert.Equal(t, 3, combo.Rank)
	assert.Equal(t, 5, combo.Length)
}

func TestClassifyCombo_StraightLong(t *testing.T) {
	cards := []*Card{dzSpade(3), dzHeart(4), dzDiamond(5), dzClover(6), dzSpade(7), dzHeart(8), dzDiamond(9), dzClover(10), dzSpade(11), dzHeart(12), dzDiamond(13), dzClover(1)}
	combo := DoudizhuClassifyCombo(cards)
	assert.NotNil(t, combo)
	assert.Equal(t, DoudizhuComboStraight, combo.Type)
	assert.Equal(t, 3, combo.Rank)
	assert.Equal(t, 12, combo.Length)
}

func TestClassifyCombo_StraightInvalid_Contains2(t *testing.T) {
	combo := DoudizhuClassifyCombo([]*Card{dzSpade(10), dzHeart(11), dzDiamond(12), dzClover(13), dzSpade(1), dzHeart(2)})
	assert.Nil(t, combo)
}

func TestClassifyCombo_StraightInvalid_ContainsJoker(t *testing.T) {
	combo := DoudizhuClassifyCombo([]*Card{dzSpade(10), dzHeart(11), dzDiamond(12), dzClover(13), dzSmallJoker()})
	assert.Nil(t, combo)
}

func TestClassifyCombo_StraightInvalid_TooShort(t *testing.T) {
	combo := DoudizhuClassifyCombo([]*Card{dzSpade(3), dzHeart(4), dzDiamond(5), dzClover(6)})
	assert.Nil(t, combo)
}

func TestClassifyCombo_StraightInvalid_Gap(t *testing.T) {
	combo := DoudizhuClassifyCombo([]*Card{dzSpade(3), dzHeart(4), dzDiamond(5), dzClover(6), dzSpade(8)})
	assert.Nil(t, combo)
}

func TestClassifyCombo_ConsecutivePair(t *testing.T) {
	combo := DoudizhuClassifyCombo([]*Card{dzSpade(3), dzHeart(3), dzDiamond(4), dzClover(4), dzSpade(5), dzHeart(5)})
	assert.NotNil(t, combo)
	assert.Equal(t, DoudizhuComboConsecutivePair, combo.Type)
	assert.Equal(t, 3, combo.Rank)
	assert.Equal(t, 3, combo.Length)
}

func TestClassifyCombo_ConsecutivePairInvalid_Contains2(t *testing.T) {
	combo := DoudizhuClassifyCombo([]*Card{dzSpade(13), dzHeart(13), dzDiamond(1), dzClover(1), dzSpade(2), dzHeart(2)})
	assert.Nil(t, combo)
}

func TestClassifyCombo_ConsecutivePairInvalid_TwoSetsOnly(t *testing.T) {
	combo := DoudizhuClassifyCombo([]*Card{dzSpade(3), dzHeart(3), dzDiamond(4), dzClover(4)})
	assert.Nil(t, combo)
}

func TestClassifyCombo_Airplane(t *testing.T) {
	combo := DoudizhuClassifyCombo([]*Card{
		dzSpade(5), dzHeart(5), dzDiamond(5),
		dzClover(6), dzSpade(6), dzHeart(6),
	})
	assert.NotNil(t, combo)
	assert.Equal(t, DoudizhuComboAirplane, combo.Type)
	assert.Equal(t, 5, combo.Rank)
	assert.Equal(t, 2, combo.Length)
}

func TestClassifyCombo_AirplaneSingle(t *testing.T) {
	combo := DoudizhuClassifyCombo([]*Card{
		dzSpade(5), dzHeart(5), dzDiamond(5),
		dzClover(6), dzSpade(6), dzHeart(6),
		dzDiamond(3), dzClover(9),
	})
	assert.NotNil(t, combo)
	assert.Equal(t, DoudizhuComboAirplaneSingle, combo.Type)
	assert.Equal(t, 5, combo.Rank)
	assert.Equal(t, 2, combo.Length)
}

func TestClassifyCombo_AirplanePair(t *testing.T) {
	combo := DoudizhuClassifyCombo([]*Card{
		dzSpade(5), dzHeart(5), dzDiamond(5),
		dzClover(6), dzSpade(6), dzHeart(6),
		dzDiamond(3), dzClover(3), dzSpade(9), dzHeart(9),
	})
	assert.NotNil(t, combo)
	assert.Equal(t, DoudizhuComboAirplanePair, combo.Type)
	assert.Equal(t, 5, combo.Rank)
	assert.Equal(t, 2, combo.Length)
}

func TestClassifyCombo_AirplaneInvalid_Contains2(t *testing.T) {
	combo := DoudizhuClassifyCombo([]*Card{
		dzSpade(1), dzHeart(1), dzDiamond(1),
		dzClover(2), dzSpade(2), dzHeart(2),
	})
	assert.Nil(t, combo)
}

func TestClassifyCombo_Empty(t *testing.T) {
	assert.Nil(t, DoudizhuClassifyCombo([]*Card{}))
}

func TestCanBeat_SameTypeSingleHigher(t *testing.T) {
	table := &DoudizhuCombo{Type: DoudizhuComboSingle, Rank: 5, Length: 1}
	play := &DoudizhuCombo{Type: DoudizhuComboSingle, Rank: 8, Length: 1}
	assert.True(t, DoudizhuCanBeat(play, table))
}

func TestCanBeat_SameTypeSingleLower(t *testing.T) {
	table := &DoudizhuCombo{Type: DoudizhuComboSingle, Rank: 8, Length: 1}
	play := &DoudizhuCombo{Type: DoudizhuComboSingle, Rank: 5, Length: 1}
	assert.False(t, DoudizhuCanBeat(play, table))
}

func TestCanBeat_DifferentType(t *testing.T) {
	table := &DoudizhuCombo{Type: DoudizhuComboSingle, Rank: 5, Length: 1}
	play := &DoudizhuCombo{Type: DoudizhuComboPair, Rank: 8, Length: 1}
	assert.False(t, DoudizhuCanBeat(play, table))
}

func TestCanBeat_DifferentLength(t *testing.T) {
	table := &DoudizhuCombo{Type: DoudizhuComboStraight, Rank: 3, Length: 5}
	play := &DoudizhuCombo{Type: DoudizhuComboStraight, Rank: 5, Length: 6}
	assert.False(t, DoudizhuCanBeat(play, table))
}

func TestCanBeat_BombBeatsSingle(t *testing.T) {
	table := &DoudizhuCombo{Type: DoudizhuComboSingle, Rank: 15, Length: 1}
	play := &DoudizhuCombo{Type: DoudizhuComboBomb, Rank: 3, Length: 1}
	assert.True(t, DoudizhuCanBeat(play, table))
}

func TestCanBeat_BombVsBomb(t *testing.T) {
	table := &DoudizhuCombo{Type: DoudizhuComboBomb, Rank: 5, Length: 1}
	play := &DoudizhuCombo{Type: DoudizhuComboBomb, Rank: 8, Length: 1}
	assert.True(t, DoudizhuCanBeat(play, table))

	play2 := &DoudizhuCombo{Type: DoudizhuComboBomb, Rank: 3, Length: 1}
	assert.False(t, DoudizhuCanBeat(play2, table))
}

func TestCanBeat_RocketBeatsEverything(t *testing.T) {
	rocket := &DoudizhuCombo{Type: DoudizhuComboRocket, Rank: doudizhuStrengthBigJoker, Length: 1}

	bomb := &DoudizhuCombo{Type: DoudizhuComboBomb, Rank: 15, Length: 1}
	assert.True(t, DoudizhuCanBeat(rocket, bomb))

	single := &DoudizhuCombo{Type: DoudizhuComboSingle, Rank: 15, Length: 1}
	assert.True(t, DoudizhuCanBeat(rocket, single))
}

func TestCanBeat_BombDoesNotBeatRocket(t *testing.T) {
	rocket := &DoudizhuCombo{Type: DoudizhuComboRocket, Rank: doudizhuStrengthBigJoker, Length: 1}
	bomb := &DoudizhuCombo{Type: DoudizhuComboBomb, Rank: 15, Length: 1}
	assert.False(t, DoudizhuCanBeat(bomb, rocket))
}

func TestCanBeat_PairVsPair(t *testing.T) {
	table := &DoudizhuCombo{Type: DoudizhuComboPair, Rank: 7, Length: 1}
	play := &DoudizhuCombo{Type: DoudizhuComboPair, Rank: 14, Length: 1}
	assert.True(t, DoudizhuCanBeat(play, table))
}

func TestClassifyCombo_AirplaneThreeChain(t *testing.T) {
	combo := DoudizhuClassifyCombo([]*Card{
		dzSpade(7), dzHeart(7), dzDiamond(7),
		dzClover(8), dzSpade(8), dzHeart(8),
		dzDiamond(9), dzClover(9), dzSpade(9),
	})
	assert.NotNil(t, combo)
	assert.Equal(t, DoudizhuComboAirplane, combo.Type)
	assert.Equal(t, 7, combo.Rank)
	assert.Equal(t, 3, combo.Length)
}
