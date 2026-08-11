//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestBhabhi(t *testing.T) *Bhabhi {
	t.Helper()
	b := NewDefaultBhabhi()
	b.Reset()
	return b
}

func TestBhabhi_ResetDealsTheWholeDeck(t *testing.T) {
	for _, n := range []int{BhabhiMinPlayers, 4, 5, 6, BhabhiMaxPlayers} {
		b := NewDefaultBhabhi()
		b.SetConfig(BhabhiConfig{PlayerCnt: n})
		b.Reset()

		require.Equal(t, n, b.GetPlayerCnt(), "設定した人数で席を作り直す")
		total := 0
		sizes := make([]int, 0, n)
		for i := range n {
			total += b.GetPlayer(i).GetCardsSize()
			sizes = append(sizes, b.GetPlayer(i).GetCardsSize())
		}
		// **山札は残さない。** 52 枚を配り切る。
		assert.Equal(t, BhabhiDeckSize, total, "%d 人でも 52 枚すべて配る", n)
		// **52 は 3/5/6/7 人で割り切れない。** 差は 1 枚まで。
		lo, hi := sizes[0], sizes[0]
		for _, s := range sizes {
			lo, hi = min(lo, s), max(hi, s)
		}
		assert.LessOrEqual(t, hi-lo, 1, "手札枚数の差は 1 枚まで (%d 人)", n)
	}
}

func TestBhabhi_ResetStartsInPlay(t *testing.T) {
	b := newTestBhabhi(t)
	assert.Equal(t, BhabhiPhasePlay, b.GetPhase())
	assert.False(t, b.GetGameEndFlag())
	assert.Equal(t, 0, b.GetLeadSuit(), "リードスートはまだ無い")
	assert.Empty(t, b.GetPile())
	assert.Equal(t, -1, b.GetBhabhiIdx())
	assert.Equal(t, -1, b.GetLastPickupIdx())
	assert.Equal(t, BhabhiDefaultPlayers, b.GetAliveCount())
}

// **フォローできるならそのスートしか出せない。**
func TestBhabhi_FollowSuitIsCompulsory(t *testing.T) {
	b := newTestBhabhi(t)
	b.SetLeadIdxForTest(0)
	b.SetCurrentIdxForTest(0)
	bhabhiHandOf(b, 0, NewCard(CardDesignSpade, 5, false))
	bhabhiHandOf(b, 1, NewCard(CardDesignSpade, 7, false), NewCard(CardDesignHeart, 3, false))
	bhabhiHandOf(b, 2, NewCard(CardDesignSpade, 9, false))
	bhabhiHandOf(b, 3, NewCard(CardDesignSpade, 2, false))

	require.NoError(t, b.PlayForTest(0, 0))
	assert.Equal(t, CardDesignSpade, b.GetLeadSuit())

	// プレイヤー1 は ♠ を持っているので ♥ は出せない。
	assert.Equal(t, []int{0}, b.GetValidPlayIndices(1))
	assert.Error(t, b.PlayForTest(1, 1), "フォローできるのに別スートは通さない")
}

// **フォローできないときは何を出してもよく、場札を全部引き取る。**
func TestBhabhi_CannotFollowPicksUpTheWholePile(t *testing.T) {
	b := newTestBhabhi(t)
	b.SetLeadIdxForTest(0)
	b.SetCurrentIdxForTest(0)
	bhabhiHandOf(b, 0, NewCard(CardDesignSpade, 5, false), NewCard(CardDesignClover, 4, false))
	bhabhiHandOf(b, 1, NewCard(CardDesignSpade, 7, false), NewCard(CardDesignClover, 6, false))
	// プレイヤー2 は ♠ を持っていない。
	bhabhiHandOf(b, 2, NewCard(CardDesignHeart, 3, false), NewCard(CardDesignHeart, 8, false))
	bhabhiHandOf(b, 3, NewCard(CardDesignSpade, 9, false), NewCard(CardDesignClover, 2, false))

	require.NoError(t, b.PlayForTest(0, 0))
	require.NoError(t, b.PlayForTest(1, 0))
	assert.Len(t, b.GetPile(), 2)

	require.NoError(t, b.PlayForTest(2, 0)) // ♥3 を出してフォロー失敗
	// **出した 1 枚も含めて全部引き取る。** 手札 2 枚 -1 +3 = 4 枚。
	assert.Equal(t, 4, b.GetPlayer(2).GetCardsSize(), "自分が出した札も戻ってくる")
	assert.Empty(t, b.GetPile(), "場は空になる")
	assert.Equal(t, 0, b.GetLeadSuit(), "リードスートも消える")
	assert.Equal(t, 2, b.GetLastPickupIdx())
	assert.Equal(t, 3, b.GetLastPickupSize())
	assert.Equal(t, 1, b.GetPlayer(2).GetPickups())
	// **引き取った人は次のリードを取らない。** 詳細は
	// TestBhabhi_ThePickerDoesNotLeadNext。
	assert.Equal(t, 3, b.GetCurrentPlayerIdx())
	assert.Equal(t, 3, b.GetLeadPlayerIdx())
	// プレイヤー3 は打っていない。
	assert.Equal(t, 2, b.GetPlayer(3).GetCardsSize())
}

// **全員フォローできたトリックは捨てる。引き取らない。**
//
// これが issue #5244 の規則 4 との違い。引き取らせると場札が必ず誰かの手札へ
// 戻るので、**手札の総数が 52 枚のまま減らず、誰も上がれずゲームが終わらない**。
func TestBhabhi_ACompletedTrickLeavesPlayForGood(t *testing.T) {
	b := newTestBhabhi(t)
	b.SetLeadIdxForTest(0)
	b.SetCurrentIdxForTest(0)
	for i := range BhabhiDefaultPlayers {
		bhabhiHandOf(b, i,
			NewCard(CardDesignSpade, 3+i, false),
			NewCard(CardDesignHeart, 3+i, false))
	}
	before := 0
	for i := range BhabhiDefaultPlayers {
		before += b.GetPlayer(i).GetCardsSize()
	}

	for i := range BhabhiDefaultPlayers {
		require.NoError(t, b.PlayForTest(i, 0))
	}

	after := 0
	for i := range BhabhiDefaultPlayers {
		after += b.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, before-BhabhiDefaultPlayers, after, "出した 4 枚は場から消える")
	assert.Empty(t, b.GetPile())
	assert.Equal(t, 1, b.GetTrickNumber())
	assert.Equal(t, -1, b.GetLastPickupIdx(), "引き取りは起きていない")
}

// **最強札を出した人が次のリード。** A が最強。
func TestBhabhi_HighestOfTheLedSuitLeadsNext(t *testing.T) {
	b := newTestBhabhi(t)
	b.SetLeadIdxForTest(0)
	b.SetCurrentIdxForTest(0)
	bhabhiHandOf(b, 0, NewCard(CardDesignSpade, 13, false), NewCard(CardDesignHeart, 2, false))
	bhabhiHandOf(b, 1, NewCard(CardDesignSpade, 1, false), NewCard(CardDesignHeart, 3, false)) // ♠A
	bhabhiHandOf(b, 2, NewCard(CardDesignSpade, 12, false), NewCard(CardDesignHeart, 4, false))
	bhabhiHandOf(b, 3, NewCard(CardDesignSpade, 11, false), NewCard(CardDesignHeart, 5, false))

	for i := range BhabhiDefaultPlayers {
		require.NoError(t, b.PlayForTest(i, 0))
	}
	assert.Equal(t, 1, b.GetLeadPlayerIdx(), "♠A の 1 が次のリード")
	assert.Equal(t, 1, b.GetCurrentPlayerIdx())
}

// **切り札は無い。** 別スートの高い札はトリックを取らない。
func TestBhabhi_OffSuitNeverWinsTheTrick(t *testing.T) {
	b := newTestBhabhi(t)
	b.SetLeadSuitForTest(CardDesignSpade)
	b.SetPileForTest([]*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 3, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 1, false)}, // ♥A でも取らない
		{PlayerIdx: 2, Card: NewCard(CardDesignSpade, 2, false)},
	})
	assert.Equal(t, 0, b.TrickWinnerForTest())
}

// **手札を出し切ったら上がり。** 上がった人は以降の手番に来ない。
func TestBhabhi_EmptyingYourHandFinishesYou(t *testing.T) {
	b := newTestBhabhi(t)
	b.SetLeadIdxForTest(0)
	b.SetCurrentIdxForTest(0)
	bhabhiHandOf(b, 0, NewCard(CardDesignSpade, 5, false)) // これが最後の 1 枚
	bhabhiHandOf(b, 1, NewCard(CardDesignSpade, 7, false), NewCard(CardDesignHeart, 3, false))
	bhabhiHandOf(b, 2, NewCard(CardDesignSpade, 9, false), NewCard(CardDesignHeart, 4, false))
	bhabhiHandOf(b, 3, NewCard(CardDesignSpade, 2, false), NewCard(CardDesignHeart, 5, false))

	require.NoError(t, b.PlayForTest(0, 0))
	assert.True(t, b.GetPlayer(0).IsOut(), "手札 0 枚で上がり")
	assert.Equal(t, 1, b.GetPlayer(0).GetRank(), "上がった順に順位が付く")
	assert.Equal(t, BhabhiDefaultPlayers-1, b.GetAliveCount())

	for i := 1; i < BhabhiDefaultPlayers; i++ {
		require.NoError(t, b.PlayForTest(i, 0))
	}
	assert.NotEqual(t, 0, b.GetCurrentPlayerIdx(), "上がった席に手番は来ない")
	assert.Empty(t, b.GetValidPlayIndices(0), "上がった席に出せる札は無い")
}

// **最後に 1 人だけ残った人が Bhabhi。**
func TestBhabhi_TheLastPlayerHoldingCardsIsTheBhabhi(t *testing.T) {
	b := newTestBhabhi(t)
	b.SetLeadIdxForTest(0)
	b.SetCurrentIdxForTest(0)
	bhabhiHandOf(b, 0, NewCard(CardDesignSpade, 5, false))
	bhabhiHandOf(b, 1, NewCard(CardDesignSpade, 7, false))
	bhabhiHandOf(b, 2, NewCard(CardDesignSpade, 9, false))
	bhabhiHandOf(b, 3, NewCard(CardDesignSpade, 2, false), NewCard(CardDesignHeart, 5, false))

	// **3 人が上がった時点で決着する。** 残り 1 人が打つ前に敗者は確定している。
	for i := range BhabhiDefaultPlayers - 1 {
		require.NoError(t, b.PlayForTest(i, 0))
	}
	assert.True(t, b.GetGameEndFlag())
	assert.Equal(t, BhabhiPhaseGameEnd, b.GetPhase())
	assert.Equal(t, 3, b.GetBhabhiIdx(), "手札が残っている 3 が敗者")
	assert.Equal(t, 1, b.GetAliveCount())
	assert.False(t, b.IsStalemate())
}

// **引き取った人は絶対に上がれない。** 引き取り直後は必ず手札がある。
func TestBhabhi_PickingUpKeepsYouInTheGame(t *testing.T) {
	b := newTestBhabhi(t)
	b.SetLeadIdxForTest(0)
	b.SetCurrentIdxForTest(0)
	bhabhiHandOf(b, 0, NewCard(CardDesignSpade, 5, false), NewCard(CardDesignClover, 4, false))
	bhabhiHandOf(b, 1, NewCard(CardDesignHeart, 3, false)) // ♠ が無い最後の 1 枚
	bhabhiHandOf(b, 2, NewCard(CardDesignSpade, 9, false), NewCard(CardDesignClover, 2, false))
	bhabhiHandOf(b, 3, NewCard(CardDesignSpade, 2, false), NewCard(CardDesignClover, 6, false))

	require.NoError(t, b.PlayForTest(0, 0))
	require.NoError(t, b.PlayForTest(1, 0))
	// 最後の 1 枚を出したが、フォローできないので場札ごと戻ってくる。
	assert.False(t, b.GetPlayer(1).IsOut(), "引き取ったので上がっていない")
	assert.Equal(t, 2, b.GetPlayer(1).GetCardsSize())
	assert.False(t, b.GetGameEndFlag())
}

// **リードは何を出してもよい。** リード自身がフォロー失敗になることは無い。
func TestBhabhi_TheLeaderCanPlayAnything(t *testing.T) {
	b := newTestBhabhi(t)
	b.SetLeadIdxForTest(0)
	b.SetCurrentIdxForTest(0)
	bhabhiHandOf(b, 0, NewCard(CardDesignSpade, 5, false), NewCard(CardDesignHeart, 9, false))
	for i := 1; i < BhabhiDefaultPlayers; i++ {
		bhabhiHandOf(b, i, NewCard(CardDesignHeart, 2+i, false))
	}
	assert.Equal(t, []int{0, 1}, b.GetValidPlayIndices(0), "リードは全部出せる")

	require.NoError(t, b.PlayForTest(0, 1)) // ♥9 をリード
	assert.Equal(t, CardDesignHeart, b.GetLeadSuit())
	assert.Equal(t, -1, b.GetLastPickupIdx(), "リードは引き取らない")
	assert.Len(t, b.GetPile(), 1)
}

func TestBhabhi_RejectsOutOfTurnAndBadIndices(t *testing.T) {
	b := newTestBhabhi(t)
	b.SetCurrentIdxForTest(0)
	assert.Error(t, b.PlayForTest(1, 0), "手番でない席は打てない")
	assert.Error(t, b.PlayForTest(0, -1))
	assert.Error(t, b.PlayForTest(0, 999))

	b.FinishGameForTest()
	assert.Error(t, b.PlayForTest(0, 0), "終局後は打てない")
}

func TestBhabhi_PlayerPlayGuardsOnTheHumanTurn(t *testing.T) {
	b := newTestBhabhi(t)
	b.SetCurrentIdxForTest(1)
	assert.Error(t, b.PlayerPlay(0))
	assert.False(t, b.IsHumanTurn())

	b.SetCurrentIdxForTest(0)
	assert.True(t, b.IsHumanTurn())
	assert.NoError(t, b.PlayerPlay(0))
}

// **CPU はフォローできないなら高い札を落とす。** どのみち引き取るので、
// 抱えたくない札を先に処分する場面。
func TestBhabhi_CpuDumpsItsHighestWhenItCannotFollow(t *testing.T) {
	b := newTestBhabhi(t)
	b.SetLeadSuitForTest(CardDesignSpade)
	b.SetCurrentIdxForTest(1)
	b.SetPileForTest([]*TrickCard{{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 5, false)}})
	bhabhiHandOf(b, 1,
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignHeart, 1, false), // ♥A がいちばん高い
		NewCard(CardDesignClover, 9, false))
	assert.Equal(t, 1, b.CpuChoiceForTest(1))
}

// **フォローできるなら安く出す。** 取ってしまうと次のリードを押し付けられる。
func TestBhabhi_CpuDucksWhenItCanFollow(t *testing.T) {
	b := newTestBhabhi(t)
	b.SetLeadSuitForTest(CardDesignSpade)
	b.SetCurrentIdxForTest(1)
	b.SetPileForTest([]*TrickCard{{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 5, false)}})
	bhabhiHandOf(b, 1,
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignSpade, 2, false), // これがいちばん安い
		NewCard(CardDesignSpade, 9, false))
	assert.Equal(t, 1, b.CpuChoiceForTest(1))
}

// **CPU は必ず合法手を返す。** 配り依存で違法手を選ばないことを回して確認。
func TestBhabhi_CpuAlwaysChoosesALegalCard(t *testing.T) {
	for range 200 {
		b := NewDefaultBhabhi()
		b.Reset()
		for range 60 {
			if b.GetGameEndFlag() {
				break
			}
			idx := b.GetCurrentPlayerIdx()
			choice := b.CpuChoiceForTest(idx)
			require.Contains(t, b.GetValidPlayIndices(idx), choice)
			require.NoError(t, b.PlayForTest(idx, choice))
		}
	}
}

// **どの局も必ず終わり、必ず敗者が 1 人決まる。**
//
// 普通は最後の 1 人が Bhabhi。膠着（BhabhiStalemateTricks）で打ち切られた
// ときも、**手札がいちばん多い人**が Bhabhi になり「誰も負けないまま止まる」
// ことは無い。どちらの経路でも敗者は手札を持っている。
func TestBhabhi_EveryGameEndsWithABhabhi(t *testing.T) {
	for _, n := range []int{BhabhiMinPlayers, 4, 5, BhabhiMaxPlayers} {
		for range 40 {
			b := NewDefaultBhabhi()
			b.SetConfig(BhabhiConfig{PlayerCnt: n})
			b.Reset()
			for turns := 0; !b.GetGameEndFlag(); turns++ {
				require.Less(t, turns, 20000, "%d 人で終わらない", n)
				idx := b.GetCurrentPlayerIdx()
				require.NoError(t, b.PlayForTest(idx, b.CpuChoiceForTest(idx)))
			}
			require.GreaterOrEqual(t, b.GetBhabhiIdx(), 0, "敗者が決まる")
			assert.Positive(t, b.GetPlayer(b.GetBhabhiIdx()).GetCardsSize(), "敗者は手札を持っている")
			assert.False(t, b.GetPlayer(b.GetBhabhiIdx()).IsOut(), "上がった人を敗者にしない")
			if b.IsStalemate() {
				assert.Equal(t, BhabhiStalemateTricks, b.GetTrickNumber())
				// 膠着なら「いちばん多い」であること。
				for i := range n {
					if p := b.GetPlayer(i); !p.IsOut() {
						assert.LessOrEqual(t, p.GetCardsSize(), b.GetPlayer(b.GetBhabhiIdx()).GetCardsSize())
					}
				}
			} else {
				assert.Equal(t, 1, b.GetAliveCount(), "決着なら残るのはちょうど 1 人")
			}
		}
	}
}

// **膠着は起こりうる。** 場から札が落ちるのは全員フォローできたトリックだけ
// なので、常に誰かがフォローできない配置になると引き取りだけが続く。
// 上限に達したら手札のいちばん多い人を Bhabhi にする。
func TestBhabhi_StalemateNamesTheFullestHand(t *testing.T) {
	b := newTestBhabhi(t)
	b.SetTrickNumberForTest(BhabhiStalemateTricks - 1)
	b.SetLeadIdxForTest(0)
	b.SetCurrentIdxForTest(0)
	// 2 が最多になるように積む。
	bhabhiHandOf(b, 0, NewCard(CardDesignSpade, 5, false), NewCard(CardDesignClover, 4, false))
	bhabhiHandOf(b, 1, NewCard(CardDesignSpade, 7, false), NewCard(CardDesignClover, 6, false))
	bhabhiHandOf(b, 2,
		NewCard(CardDesignHeart, 3, false), NewCard(CardDesignHeart, 8, false),
		NewCard(CardDesignHeart, 9, false), NewCard(CardDesignHeart, 4, false))
	bhabhiHandOf(b, 3, NewCard(CardDesignSpade, 9, false), NewCard(CardDesignClover, 2, false))

	require.NoError(t, b.PlayForTest(0, 0))
	require.NoError(t, b.PlayForTest(1, 0))
	require.NoError(t, b.PlayForTest(2, 0)) // ♠ が無い → 引き取り、ここで上限に達する

	assert.True(t, b.GetGameEndFlag())
	assert.True(t, b.IsStalemate())
	assert.Equal(t, 2, b.GetBhabhiIdx(), "手札がいちばん多い 2 が Bhabhi")
	assert.Greater(t, b.GetAliveCount(), 1, "決着ではないので複数残っている")
}

// **負のコントロール: 普通の決着では膠着フラグは立たない。**
func TestBhabhi_NormalFinishIsNotAStalemate(t *testing.T) {
	b := newTestBhabhi(t)
	b.SetLeadIdxForTest(0)
	b.SetCurrentIdxForTest(0)
	bhabhiHandOf(b, 0, NewCard(CardDesignSpade, 5, false))
	bhabhiHandOf(b, 1, NewCard(CardDesignSpade, 7, false))
	bhabhiHandOf(b, 2, NewCard(CardDesignSpade, 9, false))
	bhabhiHandOf(b, 3, NewCard(CardDesignSpade, 2, false), NewCard(CardDesignHeart, 5, false))
	for i := range BhabhiDefaultPlayers - 1 {
		require.NoError(t, b.PlayForTest(i, 0))
	}
	assert.True(t, b.GetGameEndFlag())
	assert.False(t, b.IsStalemate())
	assert.Equal(t, 3, b.GetBhabhiIdx())
}

// **引き取った人は次のリードを取らない。** 取らせると終わらない局が出る:
// あるスートが場に 1 枚しか残っていないと、その札が 2 人のあいだを往復し続ける。
func TestBhabhi_ThePickerDoesNotLeadNext(t *testing.T) {
	b := newTestBhabhi(t)
	b.SetLeadIdxForTest(0)
	b.SetCurrentIdxForTest(0)
	bhabhiHandOf(b, 0, NewCard(CardDesignSpade, 5, false), NewCard(CardDesignClover, 4, false))
	bhabhiHandOf(b, 1, NewCard(CardDesignHeart, 3, false), NewCard(CardDesignHeart, 8, false))
	bhabhiHandOf(b, 2, NewCard(CardDesignSpade, 9, false), NewCard(CardDesignClover, 2, false))
	bhabhiHandOf(b, 3, NewCard(CardDesignSpade, 2, false), NewCard(CardDesignClover, 6, false))

	require.NoError(t, b.PlayForTest(0, 0))
	require.NoError(t, b.PlayForTest(1, 0)) // 1 が引き取る

	assert.Equal(t, 1, b.GetLastPickupIdx())
	assert.Equal(t, 2, b.GetLeadPlayerIdx(), "引き取った 1 ではなく次の 2 がリード")
	assert.Equal(t, 2, b.GetCurrentPlayerIdx())
}

// **カードは増えも減りもしない。** 手札 + 場札は常に 52 枚。
func TestBhabhi_CardsAreConserved(t *testing.T) {
	b := newTestBhabhi(t)
	count := func() int {
		n := len(b.GetPile())
		for i := range b.GetPlayerCnt() {
			n += b.GetPlayer(i).GetCardsSize()
		}
		return n
	}
	// **上がった人の手札 0 枚も込みで数える。** 捨てられた札は場からも手札からも
	// 消えるので、ここでは「まだ生きている札」を数えている。
	seen := BhabhiDeckSize
	for range 80 {
		if b.GetGameEndFlag() {
			break
		}
		before := count()
		idx := b.GetCurrentPlayerIdx()
		require.NoError(t, b.PlayForTest(idx, b.CpuChoiceForTest(idx)))
		after := count()
		assert.LessOrEqual(t, after, before, "札が湧かない")
		seen = after
	}
	assert.LessOrEqual(t, seen, BhabhiDeckSize)
}

func TestBhabhi_GiveUpMakesYouTheBhabhi(t *testing.T) {
	b := newTestBhabhi(t)
	b.GiveUp()
	assert.True(t, b.GetGameEndFlag())
	assert.Equal(t, BhabhiPhaseGameEnd, b.GetPhase())
	assert.Equal(t, 0, b.GetBhabhiIdx())

	// 二度目は何も起きない。
	b.GiveUp()
	assert.Equal(t, 0, b.GetBhabhiIdx())
}

func TestBhabhi_HintNamesALegalCard(t *testing.T) {
	b := newTestBhabhi(t)
	b.SetCurrentIdxForTest(0)
	h := b.GetHint()
	require.NotNil(t, h)
	require.NotNil(t, h.CardIndex)
	assert.Contains(t, b.GetValidPlayIndices(0), *h.CardIndex)
	assert.Equal(t, "bhabhiLead", h.Reason, "リード時はリードの助言")

	b.SetCurrentIdxForTest(1)
	assert.Nil(t, b.GetHint(), "人間の手番でなければ助言しない")

	b.SetCurrentIdxForTest(0)
	b.FinishGameForTest()
	assert.Nil(t, b.GetHint(), "終局後は助言しない")
}

// **助言の理由は場面ごとに変わる。**
func TestBhabhi_HintReasonTracksTheSituation(t *testing.T) {
	b := newTestBhabhi(t)
	b.SetCurrentIdxForTest(0)
	b.SetLeadSuitForTest(CardDesignSpade)
	b.SetPileForTest([]*TrickCard{{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 5, false)}})

	bhabhiHandOf(b, 0, NewCard(CardDesignSpade, 9, false), NewCard(CardDesignSpade, 2, false))
	assert.Equal(t, "bhabhiDuck", b.GetHint().Reason, "フォローできるなら安く")

	bhabhiHandOf(b, 0, NewCard(CardDesignHeart, 9, false), NewCard(CardDesignHeart, 2, false))
	assert.Equal(t, "bhabhiDumpHigh", b.GetHint().Reason, "フォローできないなら高い札を落とす")
}

func TestBhabhi_JSONRoundTrip(t *testing.T) {
	b := newTestBhabhi(t)
	for range 12 {
		if b.GetGameEndFlag() {
			break
		}
		idx := b.GetCurrentPlayerIdx()
		require.NoError(t, b.PlayForTest(idx, b.CpuChoiceForTest(idx)))
	}

	data, err := json.Marshal(b)
	require.NoError(t, err)

	var restored Bhabhi
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, b.GetPlayerCnt(), restored.GetPlayerCnt())
	assert.Equal(t, b.GetLeadSuit(), restored.GetLeadSuit())
	assert.Equal(t, b.GetTrickNumber(), restored.GetTrickNumber())
	assert.Equal(t, b.GetCurrentPlayerIdx(), restored.GetCurrentPlayerIdx())
	assert.Equal(t, b.GetLastPickupIdx(), restored.GetLastPickupIdx())
	assert.Equal(t, b.GetAliveCount(), restored.GetAliveCount(), "上がりが取り消されていない")
	for i := range b.GetPlayerCnt() {
		assert.Equal(t, b.GetPlayer(i).GetCardsSize(), restored.GetPlayer(i).GetCardsSize())
		assert.Equal(t, b.GetPlayer(i).GetRank(), restored.GetPlayer(i).GetRank())
		assert.Equal(t, b.GetPlayer(i).GetPickups(), restored.GetPlayer(i).GetPickups())
	}
}

// **壊れたスナップショットは弾く。** 素通しすると規則が黙って消える。
func TestBhabhi_UnmarshalRejectsBrokenSnapshots(t *testing.T) {
	base := func(t *testing.T) map[string]any {
		t.Helper()
		b := newTestBhabhi(t)
		data, err := json.Marshal(b)
		require.NoError(t, err)
		var m map[string]any
		require.NoError(t, json.Unmarshal(data, &m))
		return m
	}

	for _, tc := range []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"phase out of range", func(m map[string]any) { m["ph"] = 9 }},
		{"led suit out of range", func(m map[string]any) { m["ls"] = 9 }},
		{"led suit with an empty pile", func(m map[string]any) { m["ls"] = 3 }},
		// **逆向きも弾く。** 場札があるのにリードスートが無いのは、フォロー
		// 義務がどの札にも掛からない盤面——規則が黙って消える (レビュー指摘 PR #5308)。
		{"a pile with no led suit", func(m map[string]any) {
			m["pi"] = []any{map[string]any{"pi": 1, "c": map[string]any{"d": 1, "v": 5, "j": false}}}
		}},
		// **膠着は終局後にしか立たない。**
		{"stalemate before the game ended", func(m map[string]any) { m["sm"] = true }},
		{"current player out of range", func(m map[string]any) { m["ci"] = 99 }},
		{"lead player out of range", func(m map[string]any) { m["li"] = -1 }},
		{"bhabhi before the game ended", func(m map[string]any) { m["bi"] = 2 }},
		{"pickup index out of range", func(m map[string]any) { m["lpi"] = 99 }},
		{"negative trick number", func(m map[string]any) { m["tn"] = -1 }},
		{"finished count above the table", func(m map[string]any) { m["fc"] = 99 }},
		{"config out of range", func(m map[string]any) { m["cf"] = map[string]any{"pc": 99} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := base(t)
			tc.mutate(m)
			data, err := json.Marshal(m)
			require.NoError(t, err)
			var restored Bhabhi
			assert.Error(t, json.Unmarshal(data, &restored))
		})
	}

	// **負のコントロール: 手を加えていないスナップショットは通る。**
	m := base(t)
	data, err := json.Marshal(m)
	require.NoError(t, err)
	var ok Bhabhi
	assert.NoError(t, json.Unmarshal(data, &ok))
}

// **席数と設定は食い違ってはいけない。**
func TestBhabhi_UnmarshalRejectsASeatCountMismatch(t *testing.T) {
	b := NewDefaultBhabhi()
	b.SetConfig(BhabhiConfig{PlayerCnt: 5})
	b.Reset()
	data, err := json.Marshal(b)
	require.NoError(t, err)

	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	m["cf"] = map[string]any{"pc": 4} // 席は 5 つあるのに設定は 4

	bad, err := json.Marshal(m)
	require.NoError(t, err)
	var restored Bhabhi
	assert.Error(t, json.Unmarshal(bad, &restored))
}

func TestBhabhiConfig_Validate(t *testing.T) {
	assert.NoError(t, DefaultBhabhiConfig().Validate())
	assert.NoError(t, BhabhiConfig{PlayerCnt: BhabhiMinPlayers}.Validate())
	assert.NoError(t, BhabhiConfig{PlayerCnt: BhabhiMaxPlayers}.Validate())
	assert.Error(t, BhabhiConfig{PlayerCnt: BhabhiMinPlayers - 1}.Validate())
	assert.Error(t, BhabhiConfig{PlayerCnt: BhabhiMaxPlayers + 1}.Validate())
}

func TestBhabhi_ActionLogRecordsPickupsAndFinishes(t *testing.T) {
	b := newTestBhabhi(t)
	for range 40 {
		if b.GetGameEndFlag() {
			break
		}
		idx := b.GetCurrentPlayerIdx()
		require.NoError(t, b.PlayForTest(idx, b.CpuChoiceForTest(idx)))
	}
	kinds := map[string]bool{}
	for _, e := range b.GetActionLog() {
		kinds[e.ActionType] = true
	}
	assert.True(t, kinds["deal"], "配りが記録される")
	assert.True(t, kinds["play"], "プレイが記録される")
}

// **CpuPlay は CPU の手番でだけ動く。**
func TestBhabhi_CpuPlayOnlyMovesOnACpuTurn(t *testing.T) {
	b := newTestBhabhi(t)
	b.SetCurrentIdxForTest(0) // 人間
	before := b.GetPlayer(0).GetCardsSize()
	b.CpuPlay()
	assert.Equal(t, before, b.GetPlayer(0).GetCardsSize(), "人間の手番では動かない")

	b.SetCurrentIdxForTest(1)
	beforeCpu := b.GetPlayer(1).GetCardsSize()
	b.CpuPlay()
	assert.Equal(t, beforeCpu-1, b.GetPlayer(1).GetCardsSize(), "CPU は 1 枚出す")

	b.FinishGameForTest()
	after := b.GetPlayer(2).GetCardsSize()
	b.CpuPlay()
	assert.Equal(t, after, b.GetPlayer(2).GetCardsSize(), "終局後は動かない")
}

func TestBhabhi_ConfigRoundTrips(t *testing.T) {
	b := NewDefaultBhabhi()
	assert.Equal(t, BhabhiDefaultPlayers, b.GetConfig().PlayerCnt)
	b.SetConfig(BhabhiConfig{PlayerCnt: 6})
	assert.Equal(t, 6, b.GetConfig().PlayerCnt)
	b.Reset()
	assert.Equal(t, 6, b.GetPlayerCnt(), "設定を変えたら席も作り直す")
}

// **人数を減らしても席が残らない。**
func TestBhabhi_ResetShrinksTheTable(t *testing.T) {
	b := NewDefaultBhabhi()
	b.SetConfig(BhabhiConfig{PlayerCnt: 7})
	b.Reset()
	require.Equal(t, 7, b.GetPlayerCnt())
	b.SetConfig(BhabhiConfig{PlayerCnt: BhabhiMinPlayers})
	b.Reset()
	assert.Equal(t, BhabhiMinPlayers, b.GetPlayerCnt())
	assert.True(t, b.GetPlayer(0).GetIsHuman(), "先頭は必ず人間")
	for i := 1; i < b.GetPlayerCnt(); i++ {
		assert.False(t, b.GetPlayer(i).GetIsHuman())
	}
}

func TestBhabhi_GetPlayerBounds(t *testing.T) {
	b := newTestBhabhi(t)
	assert.Nil(t, b.GetPlayer(-1))
	assert.Nil(t, b.GetPlayer(99))
	assert.NotNil(t, b.GetPlayer(0))
	assert.Empty(t, b.GetValidPlayIndices(-1))
	assert.Empty(t, b.GetValidPlayIndices(99))
}
