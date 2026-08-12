//go:build test

package domain

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newGoofspielForTest(t *testing.T, n int) *Goofspiel {
	t.Helper()
	cfg := DefaultGoofspielConfig()
	cfg.PlayerCnt = n
	g := NewGoofspiel(newGoofspielSeats(n), cfg)
	g.SetRand(rand.New(rand.NewSource(1)))
	g.Reset()
	return g
}

// **入札札は自分のスート 13 枚で固定。** 配りに乱数は要りません。
func TestGoofspielResetDealsOneSuitEach(t *testing.T) {
	for n := GoofspielPlayerCntMin; n <= GoofspielPlayerCntMax; n++ {
		g := newGoofspielForTest(t, n)
		require.Equal(t, GoofspielPhaseBid, g.GetPhase())
		for i := range n {
			p := g.GetPlayer(i)
			assert.Equal(t, GoofspielRounds, p.GetCardsSize(), "%d 人: 席 %d", n, i)
			assert.Zero(t, p.GetScore())
			// **席ごとに別のスート。** 3 人卓は賞札 1 + 入札 3 = 4 スート要ります。
			for k := range p.GetCardsSize() {
				assert.Equal(t, GoofspielBidSuit(i), p.GetCard(k).GetDesign(), "%d 人: 席 %d", n, i)
			}
		}
		// 賞札は 1 枚めくられ、残りは 12 枚。
		assert.NotNil(t, g.GetCurrentPrize())
		assert.Equal(t, GoofspielPrizeSuit(), g.GetCurrentPrize().GetDesign())
		assert.Equal(t, GoofspielRounds-1, g.GetPrizeRemaining())
		assert.Equal(t, 1, g.GetRoundNumber())
	}
}

// **3 人卓は 4 スート要る。** issue の「3 スートに分割」では足りません。
func TestGoofspielThreePlayersNeedFourSuits(t *testing.T) {
	used := map[int]bool{GoofspielPrizeSuit(): true}
	for i := range GoofspielPlayerCntMax {
		s := GoofspielBidSuit(i)
		assert.False(t, used[s], "席 %d のスートが重複した", i)
		used[s] = true
	}
	assert.Len(t, used, GoofspielPlayerCntMax+1, "賞札 1 + 入札 3 = 4 スート")
}

// **同時に伏せて、同時に公開する。**
func TestGoofspielBidsAreSimultaneous(t *testing.T) {
	g := newGoofspielForTest(t, 2)
	require.False(t, g.HasBid(0))
	require.False(t, g.HasBid(1))

	require.NoError(t, g.BidForTest(0, 0))
	assert.True(t, g.HasBid(0))
	// **相手の入札はまだ見えない。** 公開は全員が伏せてから。
	assert.Empty(t, g.GetRevealedBids())
	assert.Equal(t, GoofspielPhaseBid, g.GetPhase())

	require.NoError(t, g.BidForTest(1, 0))
	g.ResolveForTest()
	assert.Equal(t, GoofspielPhaseReveal, g.GetPhase())
	assert.Len(t, g.GetRevealedBids(), 2)
}

// **最高額が賞札のランクぶん取る。**
func TestGoofspielHighestBidTakesThePrize(t *testing.T) {
	g := newGoofspielForTest(t, 2)
	g.SetCurrentPrizeForTest(NewCard(GoofspielPrizeSuit(), 9, false))
	// 席 0 は 5、席 1 は 12 を出す (手札は昇順なので index = rank-1)。
	require.NoError(t, g.BidForTest(0, 4))
	require.NoError(t, g.BidForTest(1, 11))
	g.ResolveForTest()

	assert.Equal(t, 1, g.GetLastWinnerIdx())
	assert.Equal(t, 9, g.GetLastGained())
	assert.Equal(t, 9, g.GetPlayer(1).GetScore())
	assert.Zero(t, g.GetPlayer(0).GetScore())
	// **使った札は戻りません。**
	assert.Equal(t, GoofspielRounds-1, g.GetPlayer(0).GetCardsSize())
}

// **同点は誰も取りません。**
func TestGoofspielTieDiscardsThePrize(t *testing.T) {
	g := newGoofspielForTest(t, 2)
	g.SetCurrentPrizeForTest(NewCard(GoofspielPrizeSuit(), 9, false))
	require.NoError(t, g.BidForTest(0, 6))
	require.NoError(t, g.BidForTest(1, 6))
	g.ResolveForTest()

	assert.Equal(t, -1, g.GetLastWinnerIdx())
	assert.Zero(t, g.GetLastGained())
	assert.Zero(t, g.GetPlayer(0).GetScore())
	assert.Zero(t, g.GetPlayer(1).GetScore())
	assert.Empty(t, g.GetCarriedPrizes(), "既定では流れる")
}

// **持ち越し設定では、次の賞に上乗せされます。**
func TestGoofspielTieCanCarryOver(t *testing.T) {
	cfg := DefaultGoofspielConfig()
	cfg.TieRule = GoofspielTieCarryOver
	g := NewGoofspiel(newGoofspielSeats(2), cfg)
	g.SetRand(rand.New(rand.NewSource(1)))
	g.Reset()
	g.SetCurrentPrizeForTest(NewCard(GoofspielPrizeSuit(), 9, false))

	require.NoError(t, g.BidForTest(0, 6))
	require.NoError(t, g.BidForTest(1, 6))
	g.ResolveForTest()
	require.Len(t, g.GetCarriedPrizes(), 1)

	// 次のラウンドは 9 点が上乗せされる。
	require.NoError(t, g.NextRound())
	prize := goofspielRank(g.GetCurrentPrize())
	assert.Equal(t, prize+9, g.PrizeValue())

	// 決着すると持ち越しは消える。
	require.NoError(t, g.BidForTest(0, 0))
	require.NoError(t, g.BidForTest(1, 1))
	g.ResolveForTest()
	assert.Empty(t, g.GetCarriedPrizes())
	assert.Equal(t, prize+9, g.GetPlayer(1).GetScore())
}

// **13 ラウンドちょうどで終わります。**
func TestGoofspielRunsExactlyThirteenRounds(t *testing.T) {
	g := newGoofspielForTest(t, 2)
	rounds := 0
	for !g.GetGameEndFlag() {
		rounds++
		require.LessOrEqual(t, rounds, GoofspielRounds+1, "13 ラウンドで終わる")
		require.NoError(t, g.PlayerBid(0))
		if g.GetPhase() == GoofspielPhaseReveal && !g.GetGameEndFlag() {
			require.NoError(t, g.NextRound())
		}
	}
	assert.Equal(t, GoofspielRounds, rounds)
	assert.Equal(t, GoofspielRounds, g.GetRoundNumber())
	assert.Zero(t, g.GetPlayer(0).GetCardsSize(), "入札札を使い切る")
	// **賞札の合計は 1..13 の総和。** 同点で流れたぶんだけ減ります。
	total := 0
	for i := range g.GetPlayerCnt() {
		total += g.GetPlayer(i).GetScore()
	}
	assert.LessOrEqual(t, total, GoofspielRounds*(GoofspielRounds+1)/2)
}

func TestGoofspielPlayerBidRejectsBadInput(t *testing.T) {
	g := newGoofspielForTest(t, 2)
	assert.Error(t, g.PlayerBid(-1))
	assert.Error(t, g.PlayerBid(99))
	require.NoError(t, g.PlayerBid(0))
	// 公開後は入札できない。
	assert.Error(t, g.PlayerBid(0))
}

func TestGoofspielGiveUp(t *testing.T) {
	g := newGoofspielForTest(t, 2)
	g.GiveUp()
	assert.True(t, g.GetGameEndFlag())
	assert.NotEqual(t, 0, g.GetWinnerIdx(), "投了した人は勝たない")

	g.GiveUp()
	assert.True(t, g.GetGameEndFlag())
}

// **高い賞には高い札。** 残りの賞札の中での順位に合わせます。
func TestGoofspielCpuMatchesThePrizeRank(t *testing.T) {
	g := newGoofspielForTest(t, 2)
	// 賞札 13 が出ていて、山に 1..12 が残っているなら最高札を出す。
	g.SetCurrentPrizeForTest(NewCard(GoofspielPrizeSuit(), 13, false))
	pile := make([]*Card, 0, 12)
	for v := 1; v <= 12; v++ {
		pile = append(pile, NewCard(GoofspielPrizeSuit(), v, false))
	}
	g.SetPrizePileForTest(pile)
	assert.Equal(t, GoofspielRounds-1, g.CpuChoiceForTest(1), "いちばん高い札")

	// 賞札 1 なら最低札。
	g.SetCurrentPrizeForTest(NewCard(GoofspielPrizeSuit(), 1, false))
	pile = pile[:0]
	for v := 2; v <= 13; v++ {
		pile = append(pile, NewCard(GoofspielPrizeSuit(), v, false))
	}
	g.SetPrizePileForTest(pile)
	assert.Equal(t, 0, g.CpuChoiceForTest(1), "いちばん低い札")
}

func TestGoofspielHint(t *testing.T) {
	g := newGoofspielForTest(t, 2)
	g.SetCurrentPrizeForTest(NewCard(GoofspielPrizeSuit(), 12, false))
	h := g.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "goofspielHighPrize", h.Reason)

	g.SetCurrentPrizeForTest(NewCard(GoofspielPrizeSuit(), 2, false))
	assert.Equal(t, "goofspielLowPrize", g.GetHint().Reason)

	g.SetCurrentPrizeForTest(NewCard(GoofspielPrizeSuit(), 7, false))
	assert.Equal(t, "goofspielMatch", g.GetHint().Reason)

	// 伏せたあとは助言しない。
	require.NoError(t, g.PlayerBid(0))
	assert.Nil(t, g.GetHint())

	g.GiveUp()
	assert.Nil(t, g.GetHint(), "終局後は助言しない")
}

// **全人数を終局まで指して、1 手ごとに保存して読み直す。**
func TestGoofspiel_EveryReachableStateSurvivesARoundTrip(t *testing.T) {
	for n := GoofspielPlayerCntMin; n <= GoofspielPlayerCntMax; n++ {
		for _, tie := range []GoofspielTieRule{GoofspielTieDiscard, GoofspielTieCarryOver} {
			for seed := range 5 {
				cfg := GoofspielConfig{PlayerCnt: n, TieRule: tie}
				g := NewGoofspiel(newGoofspielSeats(n), cfg)
				g.SetRand(rand.New(rand.NewSource(int64(seed) + 1)))
				g.Reset()

				for turns := 0; ; turns++ {
					require.Less(t, turns, 100, "%d 人: 終わらない", n)

					data, err := json.Marshal(g)
					require.NoError(t, err)
					var back Goofspiel
					require.NoError(t, json.Unmarshal(data, &back),
						"%d 人 tie=%d %d 手目: 書き込み側が codec の不変条件を破った", n, tie, turns)

					if g.GetGameEndFlag() {
						break
					}
					switch g.GetPhase() {
					case GoofspielPhaseBid:
						require.NoError(t, g.PlayerBid(g.CpuChoiceForTest(0)))
					case GoofspielPhaseReveal:
						require.NoError(t, g.NextRound())
					default:
						t.Fatalf("%d 人: 進めないフェーズ %d", n, g.GetPhase())
					}
				}
			}
		}
	}
}
