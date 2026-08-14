//go:build test

package domain

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// c4pCard は指定のスート・数字の札を返す。
func c4pCard(design, value int) *Card { return NewCard(design, value, false) }

// c4pHand は 4 枚の手を組み立てる。
func c4pHand(cards ...*Card) []*Card { return cards }

// newCrazyFourPokerForTest は既定の卓を返す。
func newCrazyFourPokerForTest(t *testing.T) *CrazyFourPoker {
	t.Helper()
	g := NewDefaultCrazyFourPoker()
	g.Reset()
	return g
}

// --- 配当表: 全列挙で出した実測値を守る ---

// **配当表は出典から写さず、実際の分布から決めた。**
//
// C(52,5)=2,598,960 を全列挙して得た頻度がこれ。列挙そのものは 29 秒かかるので
// CI では回さず、**定数として残した実測値**を後続のサンプリング検査と突き合わせる。
var crazyFourPokerExactCounts = map[int]int{
	FourCardHandFourOfAKind:   624,
	FourCardHandStraightFlush: 2072,
	FourCardHandThreeOfAKind:  58656,
	FourCardHandFlush:         114616,
	FourCardHandStraight:      101808,
	FourCardHandTwoPair:       123552,
	FourCardHandPair:          1047552,
	FourCardHandHighCard:      1150080,
}

const crazyFourPokerTotalHands = 2598960

// 実測値そのものが整合しているか (合計が C(52,5) になるか)。
func TestCrazyFourPoker_ExactCountsAreConsistent(t *testing.T) {
	t.Parallel()

	total := 0
	for _, n := range crazyFourPokerExactCounts {
		total += n
	}
	assert.Equal(t, crazyFourPokerTotalHands, total, "全列挙の内訳が C(52,5) に合わない")
}

// **Queens Up のハウスエッジは 4.5203%。**
//
// 配当表を変えたらこの値が動く。動いたことに気付かないまま出荷しないための錨。
func TestCrazyFourPoker_QueensUpHouseEdge(t *testing.T) {
	t.Parallel()

	// クイーン以上のペアの数 = (QQ 以上の全体) - (ペアより上の役すべて)
	const queensOrBetter = 644244
	above := 0
	for rank, n := range crazyFourPokerExactCounts {
		if rank > FourCardHandPair {
			above += n
		}
	}
	pairsQQ := queensOrBetter - above
	require.Positive(t, pairsQQ)

	ret := pairsQQ * crazyFourPokerQueensUpPayouts[FourCardHandPair]
	for rank, n := range crazyFourPokerExactCounts {
		if rank > FourCardHandPair {
			ret += n * crazyFourPokerQueensUpPayouts[rank]
		}
	}
	lose := crazyFourPokerTotalHands - queensOrBetter
	ev := ret - lose

	// -117,480 / 2,598,960 = -4.5203%
	assert.Equal(t, -117480, ev, "Queens Up の期待値が変わった (配当表を触った?)")
	edge := -float64(ev) / float64(crazyFourPokerTotalHands) * 100
	assert.InDelta(t, 4.5203, edge, 0.0001)
}

// **配当は希少度と同じ向きに並んでいる。**
//
// 4 枚役ではフォーカード (624) のほうがストレートフラッシュ (2072) より希少なので、
// 5 枚役の感覚で SF を上に置くと向きが食い違う。両方の表で守られていることを見る。
func TestCrazyFourPoker_PayoutsFollowRarity(t *testing.T) {
	t.Parallel()

	require.Less(t, crazyFourPokerExactCounts[FourCardHandFourOfAKind],
		crazyFourPokerExactCounts[FourCardHandStraightFlush],
		"前提が崩れている: 4 枚役ではフォーカードのほうが希少なはず")

	assert.Greater(t, crazyFourPokerQueensUpPayouts[FourCardHandFourOfAKind],
		crazyFourPokerQueensUpPayouts[FourCardHandStraightFlush],
		"Queens Up: より希少なフォーカードの配当が低い")
	assert.Greater(t, crazyFourPokerSuperBonusPayouts[FourCardHandFourOfAKind],
		crazyFourPokerSuperBonusPayouts[FourCardHandStraightFlush],
		"Super Bonus: より希少なフォーカードの配当が低い")

	// **配当は役の強さの順に並ぶ。** 希少度の順ではない。
	//
	// 4 枚役では**ストレート (101,808) のほうがフラッシュ (114,616) より希少**なのに、
	// 既存の評価器はフラッシュを上に置いている (Four Card Poker の公式順と同じ)。
	// 希少度で並べ替えると既存ゲームの勝敗が変わってしまうので、順位はそのままにして
	// **配当は順位に従わせる**。逆転しているのはこの 1 組だけ。
	order := []int{
		FourCardHandPair, FourCardHandTwoPair, FourCardHandStraight,
		FourCardHandFlush, FourCardHandThreeOfAKind,
		FourCardHandStraightFlush, FourCardHandFourOfAKind,
	}
	for i := 1; i < len(order); i++ {
		hi, lo := order[i], order[i-1]
		assert.GreaterOrEqual(t, crazyFourPokerQueensUpPayouts[hi], crazyFourPokerQueensUpPayouts[lo],
			"Queens Up の配当が役の強さの順になっていない: %s", FourCardHandNames[hi])
		assert.GreaterOrEqual(t, crazyFourPokerSuperBonusPayouts[hi], crazyFourPokerSuperBonusPayouts[lo],
			"Super Bonus の配当が役の強さの順になっていない: %s", FourCardHandNames[hi])
	}
}

// **ストレートとフラッシュだけは希少度と順位が逆。**
//
// 実測で straight 101,808 < flush 114,616。順位はフラッシュが上なので、より希少な
// ストレートのほうが安い。これは既存の評価器 (Four Card Poker の公式順) に合わせた
// 意図的な仕様で、事故ではないことをここに固定しておく。
func TestCrazyFourPoker_StraightIsRarerThanFlushButRanksLower(t *testing.T) {
	t.Parallel()

	assert.Less(t, crazyFourPokerExactCounts[FourCardHandStraight],
		crazyFourPokerExactCounts[FourCardHandFlush], "ストレートのほうが希少なはず")
	assert.Less(t, FourCardHandStraight, FourCardHandFlush, "順位はフラッシュが上")
	assert.Less(t, crazyFourPokerQueensUpPayouts[FourCardHandStraight],
		crazyFourPokerQueensUpPayouts[FourCardHandFlush], "配当は順位に従う")
}

// **評価器が実測の分布と一致する。**
//
// 定数だけを検査すると「表は綺麗だが評価器が別物」を見逃す。全列挙は遅いので、
// 固定 seed のサンプリングで頻度が実測値に寄ることを確かめる。
func TestCrazyFourPoker_EvaluatorMatchesTheMeasuredDistribution(t *testing.T) {
	const samples = 120000
	rng := rand.New(rand.NewSource(20260813))

	deck := make([]*Card, 0, 52)
	for d := CardDesignSpade; d <= CardDesignDiamond; d++ {
		for v := 1; v <= 13; v++ {
			deck = append(deck, c4pCard(d, v))
		}
	}

	seen := map[int]int{}
	for range samples {
		rng.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
		seen[evalFourCardHand(pickBestFour(deck[:5]))]++
	}

	for rank, exact := range crazyFourPokerExactCounts {
		want := float64(exact) / crazyFourPokerTotalHands
		got := float64(seen[rank]) / samples
		// 稀な役ほど相対誤差が大きいので、絶対誤差で見る。
		assert.InDelta(t, want, got, 0.008,
			"%s の出現率が実測値から離れている (want %.5f, got %.5f)",
			FourCardHandNames[rank], want, got)
	}
}

// --- ディーラーのクオリファイ ---

// **キング以上で成立する。** 境界をずらすと控除率が静かに変わる。
func TestCrazyFourPoker_DealerQualification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		best []*Card
		want bool
	}{
		{
			name: "キングハイは成立する",
			best: c4pHand(c4pCard(CardDesignSpade, 13), c4pCard(CardDesignHeart, 8),
				c4pCard(CardDesignClover, 5), c4pCard(CardDesignDiamond, 2)),
			want: true,
		},
		{
			name: "エースハイは成立する",
			best: c4pHand(c4pCard(CardDesignSpade, 1), c4pCard(CardDesignHeart, 8),
				c4pCard(CardDesignClover, 5), c4pCard(CardDesignDiamond, 2)),
			want: true,
		},
		{
			name: "クイーンハイは成立しない",
			best: c4pHand(c4pCard(CardDesignSpade, 12), c4pCard(CardDesignHeart, 8),
				c4pCard(CardDesignClover, 5), c4pCard(CardDesignDiamond, 2)),
			want: false,
		},
		{
			name: "ジャックハイは成立しない",
			best: c4pHand(c4pCard(CardDesignSpade, 11), c4pCard(CardDesignHeart, 8),
				c4pCard(CardDesignClover, 5), c4pCard(CardDesignDiamond, 2)),
			want: false,
		},
		{
			name: "低いペアでも成立する",
			best: c4pHand(c4pCard(CardDesignSpade, 2), c4pCard(CardDesignHeart, 2),
				c4pCard(CardDesignClover, 5), c4pCard(CardDesignDiamond, 7)),
			want: true,
		},
		{
			name: "手が無ければ成立しない",
			best: nil,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, crazyFourPokerQualifies(tt.best))
		})
	}
}

// --- 3 倍ルール ---

// **倍率を動かせること自体がエースのペア以上の特典。**
func TestCrazyFourPoker_MaxPlayMultiplier(t *testing.T) {
	t.Parallel()

	assert.Equal(t, CrazyFourPokerPlayAcesMax, CrazyFourPokerMaxPlayMultiplier(true))
	assert.Equal(t, CrazyFourPokerPlayNormalMax, CrazyFourPokerMaxPlayMultiplier(false))
	assert.Equal(t, 1, CrazyFourPokerMaxPlayMultiplier(false), "エース未満で 2 倍を許してはいけない")
}

func TestCrazyFourPoker_PairAtLeast(t *testing.T) {
	t.Parallel()

	aces := c4pHand(c4pCard(CardDesignSpade, 1), c4pCard(CardDesignHeart, 1),
		c4pCard(CardDesignClover, 5), c4pCard(CardDesignDiamond, 7))
	kings := c4pHand(c4pCard(CardDesignSpade, 13), c4pCard(CardDesignHeart, 13),
		c4pCard(CardDesignClover, 5), c4pCard(CardDesignDiamond, 7))
	queens := c4pHand(c4pCard(CardDesignSpade, 12), c4pCard(CardDesignHeart, 12),
		c4pCard(CardDesignClover, 5), c4pCard(CardDesignDiamond, 7))
	jacks := c4pHand(c4pCard(CardDesignSpade, 11), c4pCard(CardDesignHeart, 11),
		c4pCard(CardDesignClover, 5), c4pCard(CardDesignDiamond, 7))
	twoPair := c4pHand(c4pCard(CardDesignSpade, 3), c4pCard(CardDesignHeart, 3),
		c4pCard(CardDesignClover, 5), c4pCard(CardDesignDiamond, 5))

	// エース基準 (Super Bonus / 3 倍ルール)
	assert.True(t, crazyFourPokerPairAtLeast(aces, CrazyFourPokerSuperBonusMinPair))
	assert.False(t, crazyFourPokerPairAtLeast(kings, CrazyFourPokerSuperBonusMinPair),
		"キングのペアはエース未満")
	assert.True(t, crazyFourPokerPairAtLeast(twoPair, CrazyFourPokerSuperBonusMinPair),
		"ツーペアはペアより上の役なので通る")

	// クイーン基準 (Queens Up)
	assert.True(t, crazyFourPokerPairAtLeast(queens, CrazyFourPokerQueensUpMinPair))
	assert.True(t, crazyFourPokerPairAtLeast(kings, CrazyFourPokerQueensUpMinPair))
	assert.False(t, crazyFourPokerPairAtLeast(jacks, CrazyFourPokerQueensUpMinPair))
	assert.False(t, crazyFourPokerPairAtLeast(nil, CrazyFourPokerQueensUpMinPair))
}

// --- 賭けの検証 ---

func TestCrazyFourPoker_PlaceBetValidation(t *testing.T) {
	tests := []struct {
		name     string
		ante     int
		queensUp int
		wantErr  error
	}{
		{"最低額は通る", CrazyFourPokerAnteMin, 0, nil},
		{"上限額は通る", CrazyFourPokerAnteMax, 0, nil},
		{"低すぎるアンティ", CrazyFourPokerAnteMin - 1, 0, errCrazyFourPokerAnteRange},
		{"高すぎるアンティ", CrazyFourPokerAnteMax + 1, 0, errCrazyFourPokerAnteRange},
		{"刻みに合わないアンティ", 15, 0, errCrazyFourPokerAnteUnit},
		{"負の Queens Up", 50, -10, errCrazyFourPokerSideRange},
		{"刻みに合わない Queens Up", 50, 15, errCrazyFourPokerAnteUnit},
		{"Queens Up は 0 でよい", 50, 0, nil},
		{"Queens Up を置ける", 50, 20, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newCrazyFourPokerForTest(t)
			g.SetChips(CrazyFourPokerChipsMax)
			err := g.PlaceBet(tt.ante, tt.queensUp)
			if tt.wantErr == nil {
				require.NoError(t, err)
				return
			}
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

// **アンティと Super Bonus は同額で必須。** 選ばせない。
func TestCrazyFourPoker_SuperBonusAlwaysMatchesTheAnte(t *testing.T) {
	g := newCrazyFourPokerForTest(t)
	before := g.GetChips()

	require.NoError(t, g.PlaceBet(50, 20))
	assert.Equal(t, 50, g.GetAnteBet())
	assert.Equal(t, 50, g.GetSuperBet(), "Super Bonus がアンティと違う")
	assert.Equal(t, 20, g.GetQueensUpBet())
	assert.Equal(t, before-120, g.GetChips(), "アンティ + Super Bonus + Queens Up が引かれていない")
}

func TestCrazyFourPoker_PlaceBetRejectsInsufficientChips(t *testing.T) {
	g := newCrazyFourPokerForTest(t)
	g.SetChips(50)
	assert.ErrorIs(t, g.PlaceBet(50, 0), errCrazyFourPokerNotEnough, "50 では 2 口 (100) 置けない")
	assert.Equal(t, 50, g.GetChips(), "失敗した賭けでチップが減っている")
}

func TestCrazyFourPoker_PhaseGuards(t *testing.T) {
	g := newCrazyFourPokerForTest(t)

	assert.ErrorIs(t, g.Play(1), errCrazyFourPokerWrongPhase)
	assert.ErrorIs(t, g.Fold(), errCrazyFourPokerWrongPhase)
	assert.ErrorIs(t, g.NextRound(), errCrazyFourPokerWrongPhase)

	require.NoError(t, g.PlaceBet(50, 0))
	assert.ErrorIs(t, g.PlaceBet(50, 0), errCrazyFourPokerWrongPhase)
	assert.ErrorIs(t, g.NextRound(), errCrazyFourPokerWrongPhase)
}

// --- 決着 ---

// c4pStaged は指定の手を積んだ、判断待ちの卓を返す。
//
// **配りは乱数のままでは固定できない**ので、賭けまで進めてから手を差し替える。
func c4pStaged(t *testing.T, ante, queensUp int, player, dealer []*Card) *CrazyFourPoker {
	t.Helper()
	g := newCrazyFourPokerForTest(t)
	g.SetChips(CrazyFourPokerChipsMax)
	require.NoError(t, g.PlaceBet(ante, queensUp))
	g.playerHand = player
	g.dealerHand = dealer
	g.playerBest = pickBestFour(player)
	g.dealerBest = pickBestFour(dealer)
	return g
}

// 手の見本。
func c4pAcePair() []*Card {
	return c4pHand(c4pCard(CardDesignSpade, 1), c4pCard(CardDesignHeart, 1),
		c4pCard(CardDesignClover, 5), c4pCard(CardDesignDiamond, 7))
}

func c4pKingHigh() []*Card {
	return c4pHand(c4pCard(CardDesignSpade, 13), c4pCard(CardDesignHeart, 9),
		c4pCard(CardDesignClover, 6), c4pCard(CardDesignDiamond, 3))
}

func c4pJackHigh() []*Card {
	return c4pHand(c4pCard(CardDesignSpade, 11), c4pCard(CardDesignHeart, 9),
		c4pCard(CardDesignClover, 6), c4pCard(CardDesignDiamond, 3))
}

// **3 倍はエースのペア以上でしか置けない。**
func TestCrazyFourPoker_PlayMultiplierIsGatedByTheHand(t *testing.T) {
	t.Run("エースのペアなら 3 倍まで", func(t *testing.T) {
		g := c4pStaged(t, 50, 0, c4pAcePair(), c4pKingHigh())
		require.True(t, g.PlayerHasAcesOrBetter())
		assert.Equal(t, CrazyFourPokerPlayAcesMax, g.MaxPlayMultiplier())
		require.NoError(t, g.Play(3))
		assert.Equal(t, 150, g.GetPlayBet())
	})

	t.Run("エース未満は同額のみ", func(t *testing.T) {
		for _, mult := range []int{2, 3} {
			g := c4pStaged(t, 50, 0, c4pKingHigh(), c4pJackHigh())
			require.False(t, g.PlayerHasAcesOrBetter())
			assert.Equal(t, CrazyFourPokerPlayNormalMax, g.MaxPlayMultiplier())
			assert.ErrorIs(t, g.Play(mult), errCrazyFourPokerMultiplier,
				"エース未満で %d 倍が通った", mult)
		}
	})

	t.Run("0 倍以下は通らない", func(t *testing.T) {
		g := c4pStaged(t, 50, 0, c4pAcePair(), c4pKingHigh())
		assert.ErrorIs(t, g.Play(0), errCrazyFourPokerMultiplier)
		assert.ErrorIs(t, g.Play(-1), errCrazyFourPokerMultiplier)
	})
}

// **ディーラー不成立ならアンティ 1:1、プレイベットはプッシュ。**
func TestCrazyFourPoker_DealerNotQualified(t *testing.T) {
	g := c4pStaged(t, 50, 0, c4pKingHigh(), c4pJackHigh())
	before := g.GetChips()

	require.NoError(t, g.Play(1))
	assert.Equal(t, CrazyFourPokerResultDealerNotQualified, g.GetResult())

	// アンティ 50 -> 100 が戻り、プレイ 50 はそのまま返り、Super Bonus 50 も返る。
	assert.Equal(t, before-50+100+50+50, g.GetChips())
}

func TestCrazyFourPoker_WinLosePush(t *testing.T) {
	tests := []struct {
		name           string
		player, dealer []*Card
		want           CrazyFourPokerResult
	}{
		{"エースのペアはキングハイに勝つ", c4pAcePair(), c4pKingHigh(), CrazyFourPokerResultWin},
		{"キングハイはエースのペアに負ける", c4pKingHigh(), c4pAcePair(), CrazyFourPokerResultLose},
		{"同じ手はプッシュ", c4pKingHigh(), c4pKingHigh(), CrazyFourPokerResultPush},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := c4pStaged(t, 50, 0, tt.player, tt.dealer)
			require.NoError(t, g.Play(1))
			assert.Equal(t, tt.want, g.GetResult())
		})
	}
}

// **Super Bonus はエース以上なら勝敗に関係なく払う。**
//
// このゲームでいちばん誤解されやすい賭け。「強い手なら負けても報われる」。
func TestCrazyFourPoker_SuperBonusPaysRegardlessWhenStrong(t *testing.T) {
	// エースのペアを持って、より強いツーペアに負ける。
	strongerDealer := c4pHand(c4pCard(CardDesignSpade, 13), c4pCard(CardDesignHeart, 13),
		c4pCard(CardDesignClover, 12), c4pCard(CardDesignDiamond, 12))
	g := c4pStaged(t, 50, 0, c4pAcePair(), strongerDealer)
	before := g.GetChips()

	require.NoError(t, g.Play(1))
	require.Equal(t, CrazyFourPokerResultLose, g.GetResult(), "前提: 負けている")

	// アンティ 50 とプレイ 50 は没収。Super Bonus 50 は 1:1 で 100 戻る。
	assert.Equal(t, before-50+100, g.GetChips(),
		"負けたのに Super Bonus が払われていない (または多い)")
}

// **エース未満の Super Bonus は、勝ち/引き分けで返却・負けで没収。**
func TestCrazyFourPoker_SuperBonusPushesOrLosesWhenWeak(t *testing.T) {
	tests := []struct {
		name           string
		player, dealer []*Card
		wantSuper      int
	}{
		{"勝てば返却", c4pKingHigh(), c4pHand(c4pCard(CardDesignSpade, 13),
			c4pCard(CardDesignHeart, 8), c4pCard(CardDesignClover, 6),
			c4pCard(CardDesignDiamond, 3)), 50},
		{"引き分けでも返却", c4pKingHigh(), c4pKingHigh(), 50},
		{"負ければ没収", c4pKingHigh(), c4pHand(c4pCard(CardDesignSpade, 4),
			c4pCard(CardDesignHeart, 4), c4pCard(CardDesignClover, 6),
			c4pCard(CardDesignDiamond, 3)), 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := c4pStaged(t, 50, 0, tt.player, tt.dealer)
			require.False(t, g.PlayerHasAcesOrBetter(), "前提: エース未満")
			cmp := compareFourCardHands(g.GetPlayerBest(), g.GetDealerBest())
			assert.Equal(t, tt.wantSuper, g.superBonusPayout(cmp, g.DealerQualifies()))
		})
	}
}

// **Queens Up は自分の役だけで決まる。** 降りても生きている。
func TestCrazyFourPoker_QueensUpIsIndependent(t *testing.T) {
	t.Run("降りても払われる", func(t *testing.T) {
		g := c4pStaged(t, 50, 20, c4pAcePair(), c4pKingHigh())
		before := g.GetChips()
		require.NoError(t, g.Fold())

		assert.Equal(t, CrazyFourPokerResultFold, g.GetResult())
		// エースのペアはクイーン以上なので 1:1 -> 40 戻る。アンティと Super は没収。
		assert.Equal(t, before+40, g.GetChips())
	})

	t.Run("クイーン未満なら払われない", func(t *testing.T) {
		g := c4pStaged(t, 50, 20, c4pJackHigh(), c4pKingHigh())
		before := g.GetChips()
		require.NoError(t, g.Fold())
		assert.Equal(t, before, g.GetChips(), "クイーン未満で Queens Up が払われた")
	})

	t.Run("置いていなければ 0", func(t *testing.T) {
		g := c4pStaged(t, 50, 0, c4pAcePair(), c4pKingHigh())
		assert.Zero(t, g.queensUpPayout())
	})
}

// 4 枚のエースは Super Bonus の別枠。
func TestCrazyFourPoker_FourAcesPaysTheSpecialRate(t *testing.T) {
	fourAces := c4pHand(c4pCard(CardDesignSpade, 1), c4pCard(CardDesignHeart, 1),
		c4pCard(CardDesignClover, 1), c4pCard(CardDesignDiamond, 1))
	g := c4pStaged(t, 50, 0, fourAces, c4pKingHigh())

	mult, ok := g.superBonusMultiplier()
	require.True(t, ok)
	assert.Equal(t, CrazyFourPokerFourAcesPayout, mult)

	fourKings := c4pHand(c4pCard(CardDesignSpade, 13), c4pCard(CardDesignHeart, 13),
		c4pCard(CardDesignClover, 13), c4pCard(CardDesignDiamond, 13))
	g2 := c4pStaged(t, 50, 0, fourKings, c4pKingHigh())
	mult2, ok2 := g2.superBonusMultiplier()
	require.True(t, ok2)
	assert.Equal(t, crazyFourPokerSuperBonusPayouts[FourCardHandFourOfAKind], mult2)
	assert.Less(t, mult2, mult, "エース 4 枚が他のフォーカードより高くない")
}

// **1.5:1 が整数で割り切れる。** 刻みを 10 に固定しているため。
func TestCrazyFourPoker_FractionalPayoutIsExact(t *testing.T) {
	flush := c4pHand(c4pCard(CardDesignSpade, 1), c4pCard(CardDesignSpade, 9),
		c4pCard(CardDesignSpade, 6), c4pCard(CardDesignSpade, 3))
	g := c4pStaged(t, 10, 0, flush, c4pKingHigh())

	mult, ok := g.superBonusMultiplier()
	require.True(t, ok)
	require.Equal(t, 15, mult, "フラッシュは 1.5:1")

	// 賭け 10 に対し 10 + 10*15/10 = 25。端数は出ない。
	assert.Equal(t, 25, g.superBet+g.superBet*mult/CrazyFourPokerPayoutScale)
}

// --- 進行 ---

func TestCrazyFourPoker_NextRoundAndGameEnd(t *testing.T) {
	g := newCrazyFourPokerForTest(t)
	require.NoError(t, g.PlaceBet(CrazyFourPokerAnteMin, 0))
	require.NoError(t, g.Fold())
	require.NoError(t, g.NextRound())

	assert.Equal(t, 2, g.GetRoundNumber())
	assert.Equal(t, CrazyFourPokerPhaseBet, g.GetPhase())
	assert.False(t, g.GetGameEndFlag())

}

// 資金が尽きたらゲームが終わる。
func TestCrazyFourPoker_EndsWhenOutOfChips(t *testing.T) {
	g := newCrazyFourPokerForTest(t)
	require.NoError(t, g.PlaceBet(CrazyFourPokerAnteMin, 0))
	require.NoError(t, g.Fold())

	g.SetChips(CrazyFourPokerAnteMin) // 1 ラウンド分に足りない
	require.NoError(t, g.NextRound())
	assert.True(t, g.GetGameEndFlag())

	assert.ErrorIs(t, g.PlaceBet(CrazyFourPokerAnteMin, 0), errCrazyFourPokerGameFinished)
	assert.ErrorIs(t, g.NextRound(), errCrazyFourPokerGameFinished)
}

// --- ヒント ---

func TestCrazyFourPoker_GetHint(t *testing.T) {
	t.Run("判断どころでなければ nil", func(t *testing.T) {
		g := newCrazyFourPokerForTest(t)
		assert.Nil(t, g.GetHint())
	})

	t.Run("エース以上なら上限倍率を薦める", func(t *testing.T) {
		g := c4pStaged(t, 50, 0, c4pAcePair(), c4pKingHigh())
		h := g.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, CrazyFourPokerPlayAcesMax, h.Multiplier)
		assert.Equal(t, "acesOrBetter", h.Reason)
	})

	t.Run("キング以上なら同額で乗る", func(t *testing.T) {
		g := c4pStaged(t, 50, 0, c4pKingHigh(), c4pJackHigh())
		h := g.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, CrazyFourPokerPlayMin, h.Multiplier)
		assert.Equal(t, "marginal", h.Reason)
	})

	t.Run("弱ければフォールドを薦める", func(t *testing.T) {
		g := c4pStaged(t, 50, 0, c4pJackHigh(), c4pKingHigh())
		h := g.GetHint()
		require.NotNil(t, h)
		assert.Zero(t, h.Multiplier)
		assert.Equal(t, "fold", h.Reason)
	})
}

// --- 名前と設定 ---

func TestCrazyFourPokerNames(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "bet", CrazyFourPokerPhaseName(CrazyFourPokerPhaseBet))
	assert.Equal(t, "decide", CrazyFourPokerPhaseName(CrazyFourPokerPhaseDecide))
	assert.Equal(t, "result", CrazyFourPokerPhaseName(CrazyFourPokerPhaseResult))

	assert.Equal(t, "fold", CrazyFourPokerResultName(CrazyFourPokerResultFold))
	assert.Equal(t, "win", CrazyFourPokerResultName(CrazyFourPokerResultWin))
	assert.Equal(t, "lose", CrazyFourPokerResultName(CrazyFourPokerResultLose))
	assert.Equal(t, "push", CrazyFourPokerResultName(CrazyFourPokerResultPush))
	assert.Equal(t, "dealerNotQualified",
		CrazyFourPokerResultName(CrazyFourPokerResultDealerNotQualified))
	assert.Equal(t, "none", CrazyFourPokerResultName(CrazyFourPokerResultNone))
}

func TestCrazyFourPokerConfig_Validate(t *testing.T) {
	t.Parallel()

	assert.NoError(t, DefaultCrazyFourPokerConfig().Validate())

	for _, tt := range []struct {
		name string
		cfg  CrazyFourPokerConfig
	}{
		{"チップが少なすぎる", CrazyFourPokerConfig{InitialChips: CrazyFourPokerChipsMin - 1, DefaultAnte: 50}},
		{"チップが多すぎる", CrazyFourPokerConfig{InitialChips: CrazyFourPokerChipsMax + 1, DefaultAnte: 50}},
		{"アンティが少なすぎる", CrazyFourPokerConfig{InitialChips: 1000, DefaultAnte: CrazyFourPokerAnteMin - 1}},
		{"アンティが多すぎる", CrazyFourPokerConfig{InitialChips: 1000, DefaultAnte: CrazyFourPokerAnteMax + 1}},
	} {
		t.Run(tt.name, func(t *testing.T) { assert.Error(t, tt.cfg.Validate()) })
	}
}

func TestCrazyFourPoker_Accessors(t *testing.T) {
	g := newCrazyFourPokerForTest(t)
	assert.Equal(t, 52, g.GetRemainingCards())
	assert.NotNil(t, g.GetPlayer())
	assert.Equal(t, CrazyFourPokerDefaultChips, g.GetConfig().InitialChips)
	assert.NotEmpty(t, g.GetActionLog())
	assert.Zero(t, g.GetPlayerHandRank(), "配る前は役なし")
	assert.Zero(t, g.GetDealerHandRank())

	require.NoError(t, g.PlaceBet(50, 0))
	assert.Len(t, g.GetPlayerHand(), CrazyFourPokerHandSize)
	assert.Len(t, g.GetDealerHand(), CrazyFourPokerHandSize)
	assert.Len(t, g.GetPlayerBest(), CrazyFourPokerBestSize)
	assert.Len(t, g.GetDealerBest(), CrazyFourPokerBestSize)
	assert.Positive(t, g.GetPlayerHandRank())
	assert.Equal(t, 52-2*CrazyFourPokerHandSize, g.GetRemainingCards())

	g.SetConfig(CrazyFourPokerConfig{InitialChips: 500, DefaultAnte: 20})
	assert.Equal(t, 500, g.GetConfig().InitialChips)
}

func TestCrazyFourPoker_ActionLogIsBounded(t *testing.T) {
	g := newCrazyFourPokerForTest(t)
	for range crazyFourPokerMaxSliceLen + 50 {
		g.appendLog("noise", "x", nil)
	}
	assert.Len(t, g.GetActionLog(), crazyFourPokerMaxSliceLen)
}
