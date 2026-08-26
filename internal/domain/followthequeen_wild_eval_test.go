//go:build test

package domain

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// bruteForceWildRank は 52^w の総当たりで最良の役位を返す ── **オラクル**。
//
// 本番の `evalFiveCardHandWithWilds` は探索を 60 倍以上削っているので、
// 「削ったことで答えが変わっていない」ことを機械的に確かめる相手が要る。
// 遅いのは承知の上で、ここでしか使わない。
func bruteForceWildRank(cards []*Card, isWild func(*Card) bool) int {
	wildIdx := []int{}
	for i, c := range cards {
		if isWild(c) {
			wildIdx = append(wildIdx, i)
		}
	}
	work := make([]*Card, 5)
	copy(work, cards)
	best := PokerHandHighCard
	var rec func(depth int)
	rec = func(depth int) {
		if depth == len(wildIdx) {
			if r := evalFiveCardHand(work); r > best {
				best = r
			}
			return
		}
		for d := CardDesignSpade; d <= CardDesignDiamond; d++ {
			for v := 1; v <= CardValueMax; v++ {
				work[wildIdx[depth]] = NewCard(d, v, false)
				rec(depth + 1)
			}
		}
	}
	rec(0)
	return best
}

// **削った探索が答えを変えていないこと。** 乱数手を総当たりと突き合わせる。
// ワイルドは 0〜3 枚（4 枚以上は必ずファイブカードで、別の subtest で見る）。
func TestFollowTheQueenWildEval_MatchesBruteForce(t *testing.T) {
	rng := rand.New(rand.NewSource(20260822)) //nolint:gosec // 再現可能な固定シード

	for _, wildRank := range []int{0, 7, 1, 13} {
		isWild := func(c *Card) bool {
			v := c.GetValue()
			return v == FollowTheQueenQueenValue || (wildRank != 0 && v == wildRank)
		}
		for i := 0; i < 500; i++ {
			deck := make([]*Card, 0, 52)
			for d := CardDesignSpade; d <= CardDesignDiamond; d++ {
				for v := 1; v <= CardValueMax; v++ {
					deck = append(deck, NewCard(d, v, false))
				}
			}
			rng.Shuffle(len(deck), func(a, b int) { deck[a], deck[b] = deck[b], deck[a] })
			hand := deck[:5]

			wilds := 0
			for _, c := range hand {
				if isWild(c) {
					wilds++
				}
			}
			if wilds > 3 {
				continue
			}

			got, cards := evalFiveCardHandWithWilds(hand, isWild)
			want := bruteForceWildRank(hand, isWild)
			require.Equal(t, want, got, "wildRank=%d hand=%s", wildRank, cuiDebugHand(hand))

			// **返した 5 枚が、本当にその役位になること。** 役位だけ合っていて
			// 札が別物なら、同位の比較が壊れる（それがこの関数を書き直した理由）。
			require.Len(t, cards, 5)
			require.Equal(t, got, evalFiveCardHand(cards),
				"返した 5 枚 %s が役位 %d になっていない", cuiDebugHand(cards), got)
		}
	}
}

// **ワイルドで作った役が、本物の同位の役に勝ってはいけない。**
// レビューで見つかった実際の形: Q♠Q♦3♣3♦9♥（Q は常時ワイルド ⇒ 3 のフォーカード）
// が 5♠5♦5♣5♥8♠（本物の 5 のフォーカード）にポットを取っていた。
// ワイルドランクすら要らない ── 普通の配りで起きる。
func TestFollowTheQueenWildEval_SubstitutedCardsDecideTheTieBreak(t *testing.T) {
	isWild := func(c *Card) bool { return c.GetValue() == FollowTheQueenQueenValue }

	wildThrees := []*Card{
		NewCard(CardDesignSpade, FollowTheQueenQueenValue, false),
		NewCard(CardDesignDiamond, FollowTheQueenQueenValue, false),
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignDiamond, 3, false),
		NewCard(CardDesignHeart, 9, false),
	}
	naturalFives := []*Card{
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignDiamond, 5, false),
		NewCard(CardDesignClover, 5, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignSpade, 8, false),
	}

	rA, handA := evalFiveCardHandWithWilds(wildThrees, isWild)
	rB, handB := evalFiveCardHandWithWilds(naturalFives, isWild)
	require.Equal(t, PokerHandFourOfAKind, rA, "Q2枚 + 3が2枚 は 3 のフォーカード")
	require.Equal(t, PokerHandFourOfAKind, rB)

	// **置換後の札で比較する。** 元の 5 枚で比べると Q(12) が 5 に勝ってしまう。
	assert.Negative(t, compareHighCardsSlice(handA, handB),
		"3 のフォーカードが 5 のフォーカードに勝っている")

	// 負のコントロール: 元の札のままだと本当に逆転すること（この検査が
	// 「たまたま通っている」のではないことを示す）。
	assert.Positive(t, compareHighCardsSlice(wildThrees, naturalFives),
		"印刷された額面で比べると逆転する ── これが直した誤り")
}

// ワイルド 4 枚以上は必ずファイブカード。返す 5 枚も本当にそうなっていること。
func TestFollowTheQueenWildEval_FourOrMoreWildsGiveFiveOfAKind(t *testing.T) {
	isWild := func(c *Card) bool { return c.GetValue() == FollowTheQueenQueenValue }
	hand := []*Card{
		NewCard(CardDesignSpade, FollowTheQueenQueenValue, false),
		NewCard(CardDesignHeart, FollowTheQueenQueenValue, false),
		NewCard(CardDesignClover, FollowTheQueenQueenValue, false),
		NewCard(CardDesignDiamond, FollowTheQueenQueenValue, false),
		NewCard(CardDesignSpade, 7, false),
	}
	r, cards := evalFiveCardHandWithWilds(hand, isWild)
	assert.Equal(t, PokerHandFiveOfAKind, r)
	require.Len(t, cards, 5)
	assert.Equal(t, PokerHandFiveOfAKind, evalFiveCardHand(cards))
	for _, c := range cards {
		assert.Equal(t, 7, c.GetValue(), "実札の 7 に合わせるはず")
	}
}

// **フラッシュがフルハウスを隠さないこと。** ワイルドに固定スートを割り当てると、
// 実札 3 枚が同スートの手で必ずフラッシュ (5) になり、フルハウス (6) を取り逃がす。
// 貪欲にスートを散らしている根拠がこれ。
func TestFollowTheQueenWildEval_PrefersFullHouseOverAnAccidentalFlush(t *testing.T) {
	isWild := func(c *Card) bool { return c.GetValue() == FollowTheQueenQueenValue }
	hand := []*Card{
		NewCard(CardDesignSpade, 4, false),
		NewCard(CardDesignSpade, 4, false),
		NewCard(CardDesignSpade, 9, false),
		NewCard(CardDesignHeart, FollowTheQueenQueenValue, false),
		NewCard(CardDesignClover, FollowTheQueenQueenValue, false),
	}
	r, cards := evalFiveCardHandWithWilds(hand, isWild)
	// 4-4-9 + ワイルド2枚 → 4 のフォーカード (7) が最良。フラッシュ (5) より上。
	assert.Equal(t, PokerHandFourOfAKind, r)
	assert.Equal(t, r, evalFiveCardHand(cards))
	assert.Equal(t, bruteForceWildRank(hand, isWild), r)
}

// cuiDebugHand はテスト失敗時に手札を読める形で出す。
func cuiDebugHand(cards []*Card) string {
	out := ""
	for _, c := range cards {
		if c == nil {
			out += "?? "
			continue
		}
		out += string(rune('0'+c.GetDesign())) + ":" + itoaSmall(c.GetValue()) + " "
	}
	return out
}

func itoaSmall(v int) string {
	if v < 10 {
		return string(rune('0' + v))
	}
	return string(rune('0'+v/10)) + string(rune('0'+v%10))
}

// mustWildRank は役位だけを見たいテストのための薄い橋渡し。
// 置換後の札まで見る検査は上の oracle テストが担当する。
func mustWildRank(rank int, _ []*Card) int { return rank }

// **探索が膨らんだら落ちるガード。**
//
// 素朴な 52^w 総当たりは w=3 の 7 枚手 1 つで **278 ms** かかっていた。Worker は
// 1 リクエストで CPU 全員分の評価を回すので、無料枠の CPU 予算 (10 ms) を一発で
// 超えて殺される ── 手元の CLI では「少し重い」としか感じられず、テストも全部通る。
//
// 上限は**わざと緩い (50 ms)**。狙いはミリ秒の測定ではなく「桁が戻っていないこと」で、
// 混んだ CI ランナーでも通る幅を取ってある。現在の実測は約 1.6 ms。
func TestFollowTheQueenWildEval_StaysWithinAWorkerCPUBudget(t *testing.T) {
	p := NewFollowTheQueenPlayer(true, 0)
	p.SetWildRank(7)
	// ワイルド 3 枚 (Q×3) + 実札 4 枚 ── 最悪に近い形。
	p.SetHoleCards([]*Card{
		NewCard(CardDesignSpade, FollowTheQueenQueenValue, false),
		NewCard(CardDesignHeart, FollowTheQueenQueenValue, false),
		NewCard(CardDesignClover, FollowTheQueenQueenValue, false),
	})
	p.SetDoorCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignClover, 9, false),
		NewCard(CardDesignDiamond, 11, false),
	})

	start := time.Now()
	const runs = 5
	for i := 0; i < runs; i++ {
		p.EvalBestHand()
	}
	per := time.Since(start) / runs

	assert.Less(t, per, 50*time.Millisecond,
		"ワイルド3枚の評価に %v かかっている ── 総当たりに戻っていないか", per)
}
