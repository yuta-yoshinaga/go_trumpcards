//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func contCard(design, value int) *Card { return NewCard(design, value, true) }
func contJoker() *Card                 { return NewCard(CardDesignJoker, CardValueJoker, true) }

// **認められる上がりの形は 3 通りだけ。**
func TestIsContinentalRummyLayout(t *testing.T) {
	for _, ok := range [][]int{
		{3, 3, 3, 3, 3},
		{4, 4, 4, 3},
		{5, 4, 3, 3},
		{3, 4, 4, 4}, // 並び順は問わない
		{3, 3, 4, 5},
	} {
		assert.True(t, IsContinentalRummyLayout(ok), "%v が認められていない", ok)
	}

	// **5 枚 3 組は合計 15 でも上がりにならない。** ここが #5464 の落とし穴。
	assert.False(t, IsContinentalRummyLayout([]int{5, 5, 5}), "5+5+5 が通ってしまっている")
	for _, ng := range [][]int{
		{5, 5, 4, 1},
		{6, 3, 3, 3},
		{3, 3, 3, 3},    // 12 枚
		{3, 3, 3, 3, 4}, // 16 枚
		{15},
		{},
	} {
		assert.False(t, IsContinentalRummyLayout(ng), "%v が通ってしまっている", ng)
	}
}

func TestContinentalRummyLayoutsIsACopy(t *testing.T) {
	got := ContinentalRummyLayouts()
	assert.Len(t, got, 3)
	got[0][0] = 99
	assert.Equal(t, 3, ContinentalRummyLayouts()[0][0], "内部の表が書き換えられている")
}

func TestIsContinentalRummyRun(t *testing.T) {
	t.Run("plain sequences of three to five", func(t *testing.T) {
		assert.True(t, IsContinentalRummyRun([]*Card{
			contCard(CardDesignSpade, 4), contCard(CardDesignSpade, 5), contCard(CardDesignSpade, 6)}))
		assert.True(t, IsContinentalRummyRun([]*Card{
			contCard(CardDesignHeart, 9), contCard(CardDesignHeart, 10),
			contCard(CardDesignHeart, 11), contCard(CardDesignHeart, 12), contCard(CardDesignHeart, 13)}))
		// 並び順は問わない。
		assert.True(t, IsContinentalRummyRun([]*Card{
			contCard(CardDesignClover, 6), contCard(CardDesignClover, 4), contCard(CardDesignClover, 5)}))
	})

	t.Run("length is bounded at three and five", func(t *testing.T) {
		assert.False(t, IsContinentalRummyRun([]*Card{
			contCard(CardDesignSpade, 4), contCard(CardDesignSpade, 5)}))
		assert.False(t, IsContinentalRummyRun([]*Card{
			contCard(CardDesignSpade, 2), contCard(CardDesignSpade, 3), contCard(CardDesignSpade, 4),
			contCard(CardDesignSpade, 5), contCard(CardDesignSpade, 6), contCard(CardDesignSpade, 7)}))
	})

	// **セットはメルドにならない。** ラミー系でここが唯一無二。
	t.Run("a set of the same rank is not a meld", func(t *testing.T) {
		assert.False(t, IsContinentalRummyRun([]*Card{
			contCard(CardDesignSpade, 7), contCard(CardDesignHeart, 7), contCard(CardDesignClover, 7)}),
			"同ランクの 3 枚が通ってしまっている")
	})

	t.Run("suits must match", func(t *testing.T) {
		assert.False(t, IsContinentalRummyRun([]*Card{
			contCard(CardDesignSpade, 4), contCard(CardDesignHeart, 5), contCard(CardDesignSpade, 6)}))
	})

	t.Run("jokers fill gaps and extend the ends", func(t *testing.T) {
		assert.True(t, IsContinentalRummyRun([]*Card{
			contCard(CardDesignSpade, 4), contJoker(), contCard(CardDesignSpade, 6)}))
		assert.True(t, IsContinentalRummyRun([]*Card{
			contCard(CardDesignSpade, 4), contCard(CardDesignSpade, 5), contJoker()}))
		// 穴が 2 つあってジョーカーが 1 枚では届かない。
		assert.False(t, IsContinentalRummyRun([]*Card{
			contCard(CardDesignSpade, 4), contJoker(), contCard(CardDesignSpade, 7)}))
		// 全部ジョーカーは何の並びか決まらない。
		assert.False(t, IsContinentalRummyRun([]*Card{contJoker(), contJoker(), contJoker()}))
	})

	t.Run("a duplicated card is not a sequence", func(t *testing.T) {
		assert.False(t, IsContinentalRummyRun([]*Card{
			contCard(CardDesignSpade, 5), contCard(CardDesignSpade, 5), contCard(CardDesignSpade, 6)}))
	})

	// **A は上端でも下端でもよいが、両方は兼ねない。**
	t.Run("the ace runs either way but never wraps", func(t *testing.T) {
		assert.True(t, IsContinentalRummyRun([]*Card{
			contCard(CardDesignDiamond, 1), contCard(CardDesignDiamond, 2), contCard(CardDesignDiamond, 3)}))
		assert.True(t, IsContinentalRummyRun([]*Card{
			contCard(CardDesignDiamond, 12), contCard(CardDesignDiamond, 13), contCard(CardDesignDiamond, 1)}))
		assert.False(t, IsContinentalRummyRun([]*Card{
			contCard(CardDesignDiamond, 13), contCard(CardDesignDiamond, 1), contCard(CardDesignDiamond, 2)}),
			"K-A-2 が繋がってしまっている")
	})

	t.Run("nil is not a card", func(t *testing.T) {
		assert.False(t, IsContinentalRummyRun([]*Card{
			contCard(CardDesignSpade, 4), nil, contCard(CardDesignSpade, 6)}))
	})
}

func TestContinentalRummyCardValue(t *testing.T) {
	assert.Equal(t, 50, ContinentalRummyCardValue(contJoker()))
	assert.Equal(t, 11, ContinentalRummyCardValue(contCard(CardDesignSpade, 1)))
	for _, v := range []int{10, 11, 12, 13} {
		assert.Equal(t, 10, ContinentalRummyCardValue(contCard(CardDesignSpade, v)))
	}
	assert.Equal(t, 7, ContinentalRummyCardValue(contCard(CardDesignHeart, 7)))
	assert.Equal(t, 0, ContinentalRummyCardValue(nil))
}
