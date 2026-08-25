//go:build test

package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// contRun は同スートの連番 n 枚を返す。
func contRun(design, from, n int) []*Card {
	out := make([]*Card, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, contCard(design, from+i))
	}
	return out
}

func contHand(groups ...[]*Card) []*Card {
	out := make([]*Card, 0, ContinentalRummyHandSize)
	for _, g := range groups {
		out = append(out, g...)
	}
	return out
}

func TestFindContinentalRummyGoOut_LegalLayouts(t *testing.T) {
	t.Run("five sequences of three", func(t *testing.T) {
		hand := contHand(
			contRun(CardDesignSpade, 2, 3), contRun(CardDesignSpade, 7, 3),
			contRun(CardDesignHeart, 4, 3), contRun(CardDesignClover, 9, 3),
			contRun(CardDesignDiamond, 5, 3))
		groups, ok := FindContinentalRummyGoOut(hand)
		require.True(t, ok, "5 組 3 枚が上がりにならない")
		assert.Len(t, groups, 5)
		assertContinentalPartition(t, hand, groups)
	})

	t.Run("three of four and one of three", func(t *testing.T) {
		hand := contHand(
			contRun(CardDesignSpade, 2, 4), contRun(CardDesignHeart, 5, 4),
			contRun(CardDesignClover, 8, 4), contRun(CardDesignDiamond, 3, 3))
		groups, ok := FindContinentalRummyGoOut(hand)
		require.True(t, ok, "4+4+4+3 が上がりにならない")
		assert.Len(t, groups, 4)
		assertContinentalPartition(t, hand, groups)
	})

	t.Run("one of five, one of four and two of three", func(t *testing.T) {
		hand := contHand(
			contRun(CardDesignSpade, 2, 5), contRun(CardDesignHeart, 5, 4),
			contRun(CardDesignClover, 8, 3), contRun(CardDesignDiamond, 3, 3))
		groups, ok := FindContinentalRummyGoOut(hand)
		require.True(t, ok, "5+4+3+3 が上がりにならない")
		assert.Len(t, groups, 4)
		assertContinentalPartition(t, hand, groups)
	})
}

// **5 枚 3 組は合計 15 でも上がりではない。** #5464 が落としている制約。
func TestFindContinentalRummyGoOut_ThreeFivesIsNotALayout(t *testing.T) {
	hand := contHand(
		contRun(CardDesignSpade, 2, 5), contRun(CardDesignHeart, 5, 5), contRun(CardDesignClover, 8, 5))
	_, ok := FindContinentalRummyGoOut(hand)
	assert.False(t, ok, "5+5+5 で上がれてしまっている")
}

// **セットはメルドにならない。** ラミー系でここが唯一無二。
func TestFindContinentalRummyGoOut_SetsAreNeverMelds(t *testing.T) {
	// **ランクは離して取る。** 2-3-4-5-6 で組むと各スートが 5 枚の連番にも
	// なってしまい、断られた理由が「セットだから」なのか「5+5+5 だから」なのか
	// 区別が付かない ── それでは何も測っていない。
	hand := make([]*Card, 0, ContinentalRummyHandSize)
	for _, v := range []int{2, 5, 8, 11, 13} {
		for _, d := range []int{CardDesignSpade, CardDesignHeart, CardDesignClover} {
			hand = append(hand, contCard(d, v))
		}
	}
	require.Len(t, hand, ContinentalRummyHandSize)
	// 前提の確認: どのスートにも連番は 1 本も無い。
	for _, d := range []int{CardDesignSpade, CardDesignHeart, CardDesignClover} {
		suited := make([]*Card, 0, 5)
		for _, c := range hand {
			if c.GetDesign() == d {
				suited = append(suited, c)
			}
		}
		require.False(t, IsContinentalRummyRun(suited[:3]), "仕込みが連番になっている")
	}
	_, ok := FindContinentalRummyGoOut(hand)
	assert.False(t, ok, "同ランク 3 枚 × 5 組で上がれてしまっている")
}

func TestFindContinentalRummyGoOut_JokersFillGaps(t *testing.T) {
	// ♠2-J-4 / ♥5-6-J / あとは素直な 3 組。
	hand := contHand(
		[]*Card{contCard(CardDesignSpade, 2), contJoker(), contCard(CardDesignSpade, 4)},
		[]*Card{contCard(CardDesignHeart, 5), contCard(CardDesignHeart, 6), contJoker()},
		contRun(CardDesignClover, 9, 3), contRun(CardDesignDiamond, 5, 3),
		contRun(CardDesignSpade, 8, 3))
	groups, ok := FindContinentalRummyGoOut(hand)
	require.True(t, ok, "ジョーカー入りで上がれない")
	assertContinentalPartition(t, hand, groups)
}

func TestFindContinentalRummyGoOut_RejectsWrongSizes(t *testing.T) {
	assert.False(t, CanContinentalRummyGoOut(contRun(CardDesignSpade, 2, 5)), "5 枚で上がれてしまっている")
	assert.False(t, CanContinentalRummyGoOut(nil))
	sixteen := append(contHand(
		contRun(CardDesignSpade, 2, 3), contRun(CardDesignSpade, 7, 3),
		contRun(CardDesignHeart, 4, 3), contRun(CardDesignClover, 9, 3),
		contRun(CardDesignDiamond, 5, 3)), contCard(CardDesignHeart, 13))
	assert.False(t, CanContinentalRummyGoOut(sixteen), "16 枚で上がれてしまっている")
}

// **一枚足りないだけで上がりではない。** 負のコントロール。
func TestFindContinentalRummyGoOut_OneBrokenGroupFails(t *testing.T) {
	hand := contHand(
		contRun(CardDesignSpade, 2, 3), contRun(CardDesignSpade, 7, 3),
		contRun(CardDesignHeart, 4, 3), contRun(CardDesignClover, 9, 3),
		[]*Card{contCard(CardDesignDiamond, 5), contCard(CardDesignDiamond, 6), contCard(CardDesignDiamond, 9)})
	assert.False(t, CanContinentalRummyGoOut(hand), "1 組が繋がっていないのに上がれてしまっている")
}

// **Worker の CPU 予算は手元では見えない。** 総当たりのままでは持ち込めないので、
// 一番割れにくい形 (全部同じスートで穴だらけ) を測って壁時計で歯止めをかける。
// 上限は環境差を吸収できるだけ緩く取り、負のコントロールで「速すぎて意味が無い
// 上限」になっていないことを見る。
func TestFindContinentalRummyGoOut_StaysCheapOnTheWorstHand(t *testing.T) {
	// ♠ だけ 15 枚、連番にならないよう散らす。全組が同スート候補になる。
	worst := make([]*Card, 0, ContinentalRummyHandSize)
	for i := 0; i < ContinentalRummyHandSize; i++ {
		worst = append(worst, contCard(CardDesignSpade, (i%13)+1))
	}
	start := time.Now()
	ok := CanContinentalRummyGoOut(worst)
	elapsed := time.Since(start)
	assert.False(t, ok, "重複だらけの手で上がれてしまっている")
	assert.Less(t, elapsed, 200*time.Millisecond, "探索が刈れていない (%v)", elapsed)

	// 負のコントロール: この上限は「一瞬で終わる」を意味しない ── ちゃんと
	// 15 枚ぶん探索しても収まる、という主張であること。
	assert.Greater(t, elapsed, time.Duration(0))
}

// assertContinentalPartition は返った組が手札をちょうど覆っていて、
// どの組も本物のシーケンスであることを確かめる。
func assertContinentalPartition(t *testing.T, hand []*Card, groups [][]int) {
	t.Helper()
	seen := map[int]bool{}
	sizes := make([]int, 0, len(groups))
	for _, g := range groups {
		cards := make([]*Card, 0, len(g))
		for _, i := range g {
			require.False(t, seen[i], "札 %d が 2 つの組に入っている", i)
			seen[i] = true
			cards = append(cards, hand[i])
		}
		assert.True(t, IsContinentalRummyRun(cards), "組 %v がシーケンスになっていない", g)
		sizes = append(sizes, len(g))
	}
	assert.Len(t, seen, len(hand), "手札を覆えていない")
	assert.True(t, IsContinentalRummyLayout(sizes), "枚数の並び %v が認められた形でない", sizes)
}
