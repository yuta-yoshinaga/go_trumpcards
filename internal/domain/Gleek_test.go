//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// newTestGleek returns a fresh, reset Gleek game (human seat 0 + 2 CPUs).
func newTestGleek() *domain.Gleek {
	g := domain.NewDefaultGleek()
	g.Reset()
	return g
}

// gleekCard is a shorthand constructor for a face-up card.
func gleekCard(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

// setGleekHand replaces player i's hand with the supplied cards deterministically.
func setGleekHand(g *domain.Gleek, i int, cards ...*domain.Card) {
	p := g.GetPlayer(i)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func TestGleek_ResetDeal(t *testing.T) {
	g := newTestGleek()
	assert.Equal(t, domain.GleekPhaseBid, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, domain.GleekPlayerCnt, g.GetPlayerCnt())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, -1, g.GetWinnerPlayer())

	for i := 0; i < g.GetPlayerCnt(); i++ {
		assert.Equal(t, domain.GleekHandSize, g.GetPlayer(i).GetCardsSize())
	}
	// **切り札は配った時点で決まる。** ストックの一番上を表にするので、
	// 競りが始まる前から全員が知っている。
	require.NotNil(t, g.GetTurnUp())
	assert.Equal(t, g.GetTurnUp().GetDesign(), g.GetTrumpSuit())
}

// **人間がエルダー。** ディーラーが席 0 のままだと、人間は毎回最後に競り、
// 最初のリードも取れない。
func TestGleek_ElderIsTheHumanOnTheOpeningDeal(t *testing.T) {
	g := newTestGleek()
	assert.Equal(t, domain.GleekPlayerCnt-1, g.GetDealerIdx())
	assert.Equal(t, 0, g.GetElderIdx())
}

func TestGleek_DeckIs44WithoutDeucesOrTreys(t *testing.T) {
	g := newTestGleek()
	seen := map[int]bool{}
	count := 0
	add := func(c *domain.Card) {
		key := c.GetDesign()*100 + c.GetValue()
		assert.False(t, seen[key], "duplicate card %d", key)
		seen[key] = true
		count++
	}
	for i := 0; i < g.GetPlayerCnt(); i++ {
		p := g.GetPlayer(i)
		for j := 0; j < p.GetCardsSize(); j++ {
			add(p.GetCard(j))
		}
	}
	assert.Equal(t, domain.GleekHandSize*domain.GleekPlayerCnt, count)
	for k := range seen {
		rank := k % 100
		assert.NotEqual(t, 2, rank, "2 は抜いたはず")
		assert.NotEqual(t, 3, rank, "3 は抜いたはず")
	}
	// 44 - 36 = 8 がストックに残る。
	assert.Equal(t, 8, domain.GleekStockSize)
	assert.Equal(t, domain.GleekDeckSize-domain.GleekHandSize*domain.GleekPlayerCnt, domain.GleekStockSize)
}

// **基準点はそのディールから数える。** 表向きの札と落札者の捨て札は卓に出ないので、
// 名札がそこに落ちたディールは上限より小さくなる。上限を基準点にすると、その分だけ
// 全員が届かず卓から点が消え続ける。
func TestGleek_ParIsReadFromTheDealNotFromTheCeiling(t *testing.T) {
	assert.Equal(t, domain.GleekTrickCount*domain.GleekTrickPoints+domain.GleekHonourTotal,
		domain.GleekMaxDealPoints, "上限はトリック点 + 名札")

	g := newTestGleek()
	g.SetTrickPointsForTest([domain.GleekPlayerCnt]int{30, 24, 21})
	assert.Equal(t, 75, g.DealPoints())
	assert.Equal(t, 25, g.Par(), "そのディールの合計を人数で割る")
	assert.NotEqual(t, domain.GleekMaxDealPoints/domain.GleekPlayerCnt, g.Par(),
		"上限から決め打った基準点とは違う値になること")
}

// **名札の合計は表から数える。** 定数を書き写すと、片方を変えたときに基準点だけ
// 古いままになる。
func TestGleek_HonourTableSumsToTheDeclaredTotal(t *testing.T) {
	trump := domain.CardDesignHeart
	total := 0
	for _, v := range []int{1, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4} {
		total += domain.GleekHonourValueForTest(gleekCard(trump, v), trump)
	}
	assert.Equal(t, domain.GleekHonourTotal, total)

	// 切り札以外の同じ札には点が付かない。
	for _, v := range []int{1, 11, 6, 5, 4} {
		assert.Zero(t, domain.GleekHonourValueForTest(gleekCard(domain.CardDesignSpade, v), trump))
	}
}

// --- bidding -------------------------------------------------------------

// **エルダーは降りられない。** 12 は自動で置かれ、卓には必ず買い手が出る。
func TestGleek_ElderOpensAtTheMinimumAndCannotPass(t *testing.T) {
	g := newTestGleek()
	assert.Equal(t, domain.GleekMinBid, g.HighestBid())
	assert.Equal(t, domain.GleekMinBid, g.GetBids()[g.GetElderIdx()])
	assert.NotEqual(t, g.GetElderIdx(), g.GetCurrentBidderIdx(), "エルダーの次から競る")
}

func TestGleek_RaisesGoUpInSteps(t *testing.T) {
	g := newTestGleek()
	assert.Equal(t, domain.GleekMinBid+domain.GleekBidStep, g.NextBidAmount())
	// **半額が整数になる刻み。** 1 刻みだと 13 の半分が端数になり、
	// 各相手に払う額の合計が落札額と合わなくなる。
	assert.Zero(t, domain.GleekBidStep%2)
	assert.Zero(t, domain.GleekMinBid%2)
}

func TestGleek_BidRejectsAnAmountThatIsNotTheNextStep(t *testing.T) {
	g := newTestGleek()
	// 人間 (席 0) がエルダーなので、次の手番は CPU。人間の手番になるまで回す。
	for g.GetPhase() == domain.GleekPhaseBid && !g.IsHumanBidTurn() {
		g.CpuBid()
	}
	if g.GetPhase() != domain.GleekPhaseBid {
		t.Skip("CPU が両方降りて競りが閉じた")
	}
	require.Error(t, g.PlayerBid(g.HighestBid()), "同額では競り上げられない")
	require.Error(t, g.PlayerBid(g.HighestBid()+1), "刻みを外れた額は置けない")
	require.NoError(t, g.PlayerBid(0), "降りることはいつでもできる")
}

// **落札者は半額を「各相手に」払う。** 1 回だけ払う実装にすると動く点が半分になる。
func TestGleek_BuyerPaysHalfToEachOpponent(t *testing.T) {
	g := newTestGleek()
	for g.GetPhase() == domain.GleekPhaseBid {
		if g.IsHumanBidTurn() {
			require.NoError(t, g.PlayerBid(0))
			continue
		}
		g.CpuBid()
	}
	buyer := g.GetBuyerIdx()
	require.True(t, buyer >= 0, "買い手が決まること")
	half := g.GetWinningBid() / 2
	scores := g.GetPlayerScores()
	// Tiddy がめくれた場合はディーラーへの支払いが混ざるので、そのディールは飛ばす。
	if g.GetTurnUp() != nil && g.GetTurnUp().GetValue() == 4 {
		t.Skip("Tiddy turn-up deal")
	}
	assert.Equal(t, -half*(domain.GleekPlayerCnt-1), scores[buyer])
	for i := 0; i < domain.GleekPlayerCnt; i++ {
		if i == buyer {
			continue
		}
		assert.Equal(t, half, scores[i])
	}
	assert.Zero(t, gleekSum(scores), "点はゼロ和で動く")
}

func TestGleek_AuctionAlwaysProducesABuyer(t *testing.T) {
	for attempt := 0; attempt < 50; attempt++ {
		g := newTestGleek()
		for step := 0; step < 100 && g.GetPhase() == domain.GleekPhaseBid; step++ {
			if g.IsHumanBidTurn() {
				require.NoError(t, g.PlayerBid(0))
				continue
			}
			g.CpuBid()
		}
		require.NotEqual(t, domain.GleekPhaseBid, g.GetPhase(), "競りが閉じない")
		require.True(t, g.GetBuyerIdx() >= 0)
		assert.True(t, g.GetWinningBid() >= domain.GleekMinBid)
		assert.True(t, g.GetWinningBid() <= domain.GleekMaxBid)
	}
}

// **降りた席はストックを買わない。** 降りても置いた額は `bids` に残るので、
// そこだけを見ると「降りたのに最高額」の席が落札してしまう。競りの手順では
// そこへ到達しないが、**保存した盤を復元すればその形は作れる**ので、
// 復元経路から確かめる。
func TestGleek_ASeatThatDroppedOutNeverBuysTheStock(t *testing.T) {
	g := newTestGleek()
	require.Equal(t, 0, g.GetElderIdx(), "人間がエルダー")

	data, err := json.Marshal(g)
	require.NoError(t, err)
	var raw map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &raw))
	// 席 0 は 16 を置いた後で降りていて、席 1 が 14 で残っている盤。
	raw["bd"] = json.RawMessage(`[16,14,0]`)
	raw["pa"] = json.RawMessage(`[true,false,false]`)
	raw["cbi"] = json.RawMessage(`2`)
	body, err := json.Marshal(raw)
	require.NoError(t, err)

	restored := new(domain.Gleek)
	require.NoError(t, json.Unmarshal(body, restored))
	require.Equal(t, domain.GleekPhaseBid, restored.GetPhase())

	// 席 2 が降りると残るのは席 1 だけ。落札するのは 14 を置いた席 1。
	for restored.GetPhase() == domain.GleekPhaseBid {
		if restored.IsHumanBidTurn() {
			require.NoError(t, restored.PlayerBid(0))
			continue
		}
		restored.CpuBid()
	}
	buyer := restored.GetBuyerIdx()
	require.True(t, buyer >= 0)
	assert.False(t, restored.GetPassed()[buyer], "降りた席が落札してはいけない")
	assert.Equal(t, 1, buyer, "降りた席 0 の 16 でなく、残った席 1 の 14 が落札する")
	assert.Equal(t, 14, restored.GetWinningBid())
}

// --- discard -------------------------------------------------------------

// buyGleekStock drives the auction until the human owns the stock, or reports false.
func buyGleekStock(t *testing.T, g *domain.Gleek) bool {
	t.Helper()
	for g.GetPhase() == domain.GleekPhaseBid {
		if g.IsHumanBidTurn() {
			if n := g.NextBidAmount(); n > 0 {
				require.NoError(t, g.PlayerBid(n))
				continue
			}
			require.NoError(t, g.PlayerBid(0))
			continue
		}
		g.CpuBid()
	}
	return g.GetBuyerIdx() == 0
}

func TestGleek_BuyerTakesTheStockAndDiscardsExactlySeven(t *testing.T) {
	var g *domain.Gleek
	for attempt := 0; attempt < 60; attempt++ {
		g = newTestGleek()
		if buyGleekStock(t, g) {
			break
		}
		g = nil
	}
	require.NotNil(t, g, "人間が落札するディールが引けない")

	require.Equal(t, domain.GleekPhaseDiscard, g.GetPhase())
	assert.Equal(t, domain.GleekHandSize+domain.GleekSwapSize, g.GetPlayer(0).GetCardsSize(),
		"表向きの札を除くストックが手札に入る")
	assert.True(t, g.IsHumanDiscardTurn())

	require.Error(t, g.PlayerDiscard([]int{0, 1, 2}), "枚数が足りない")
	require.Error(t, g.PlayerDiscard([]int{0, 0, 1, 2, 3, 4, 5}), "同じ札は二度選べない")
	require.Error(t, g.PlayerDiscard([]int{0, 1, 2, 3, 4, 5, 99}), "範囲外")

	require.NoError(t, g.PlayerDiscard([]int{0, 1, 2, 3, 4, 5, 6}))
	assert.Equal(t, domain.GleekHandSize, g.GetPlayer(0).GetCardsSize())
	assert.Equal(t, domain.GleekPhasePlay, g.GetPhase())
}

// **大きい索引から抜く。** 小さい方から抜くと後ろがずれて、指定したのと違う札が
// 捨てられる。
func TestGleek_DiscardRemovesExactlyTheChosenCards(t *testing.T) {
	var g *domain.Gleek
	for attempt := 0; attempt < 60; attempt++ {
		g = newTestGleek()
		if buyGleekStock(t, g) {
			break
		}
		g = nil
	}
	require.NotNil(t, g)

	p := g.GetPlayer(0)
	drop := []int{0, 2, 4, 6, 8, 10, 12}
	dropped := map[string]bool{}
	kept := map[string]bool{}
	for i := 0; i < p.GetCardsSize(); i++ {
		key := gleekCardKey(p.GetCard(i))
		if gleekContainsInt(drop, i) {
			dropped[key] = true
		} else {
			kept[key] = true
		}
	}
	require.NoError(t, g.PlayerDiscard(drop))

	for i := 0; i < p.GetCardsSize(); i++ {
		key := gleekCardKey(p.GetCard(i))
		assert.False(t, dropped[key], "捨てたはずの札が残っている: %s", key)
		assert.True(t, kept[key], "残すはずでない札が居る: %s", key)
	}
}

// --- ruff and melds ------------------------------------------------------

func TestGleek_RuffGoesToTheBestSingleSuitTotal(t *testing.T) {
	g := newTestGleek()
	g.SetTrumpSuit(domain.CardDesignSpade)
	// 席 1 にハートを固める (A+K+Q = 31)。他の席は散らす。
	setGleekHand(g, 0,
		gleekCard(domain.CardDesignSpade, 4), gleekCard(domain.CardDesignClover, 5),
		gleekCard(domain.CardDesignHeart, 6), gleekCard(domain.CardDesignDiamond, 7))
	setGleekHand(g, 1,
		gleekCard(domain.CardDesignHeart, 1), gleekCard(domain.CardDesignHeart, 13),
		gleekCard(domain.CardDesignHeart, 12), gleekCard(domain.CardDesignSpade, 5))
	setGleekHand(g, 2,
		gleekCard(domain.CardDesignClover, 10), gleekCard(domain.CardDesignClover, 9),
		gleekCard(domain.CardDesignDiamond, 8), gleekCard(domain.CardDesignSpade, 6))
	g.SetBuyerIdx(0)
	// **段階の点だけを見る。** めくれた札が Tiddy だとディーラーへの支払いが
	// 先に入っていて、ラフの取り分と混ざる。
	g.SetPlayerScores([domain.GleekPlayerCnt]int{})
	g.StartPlayForTest()

	assert.Equal(t, 1, g.GetRuffWinnerIdx())
	ruffs := g.GetRuffs()
	require.Len(t, ruffs, domain.GleekPlayerCnt)
	assert.Equal(t, domain.CardDesignHeart, ruffs[1].Suit)
	assert.Equal(t, 31, ruffs[1].Total, "A=11 K=10 Q=10")

	scores := g.GetPlayerScores()
	assert.Equal(t, domain.GleekRuffStake*(domain.GleekPlayerCnt-1), scores[1])
	assert.Equal(t, -domain.GleekRuffStake, scores[0])
	assert.Equal(t, -domain.GleekRuffStake, scores[2])
}

// **エース 4 枚のマーニヴァルはどんなラフにも勝つ。** 散らばっていて合計が低くても
// ラフを取る。
func TestGleek_MournivalOfAcesBeatsAnyRuff(t *testing.T) {
	g := newTestGleek()
	g.SetTrumpSuit(domain.CardDesignSpade)
	setGleekHand(g, 0,
		gleekCard(domain.CardDesignHeart, 1), gleekCard(domain.CardDesignHeart, 13),
		gleekCard(domain.CardDesignHeart, 12), gleekCard(domain.CardDesignHeart, 11))
	setGleekHand(g, 1,
		gleekCard(domain.CardDesignSpade, 1), gleekCard(domain.CardDesignClover, 1),
		gleekCard(domain.CardDesignDiamond, 1), gleekCard(domain.CardDesignSpade, 4))
	setGleekHand(g, 2, gleekCard(domain.CardDesignClover, 4))
	// 席 1 は ♥A を席 0 が持っているので 4 枚に足りない。まず負けることを確かめる。
	g.SetBuyerIdx(0)
	// **段階の点だけを見る。** めくれた札が Tiddy だとディーラーへの支払いが
	// 先に入っていて、ラフの取り分と混ざる。
	g.SetPlayerScores([domain.GleekPlayerCnt]int{})
	g.StartPlayForTest()
	assert.Equal(t, 0, g.GetRuffWinnerIdx(), "エースが 3 枚では普通のラフ勝負")

	// ♥A を席 1 に移すと 4 枚そろい、合計で負けていてもラフを取る。
	h := newTestGleek()
	h.SetTrumpSuit(domain.CardDesignSpade)
	setGleekHand(h, 0,
		gleekCard(domain.CardDesignHeart, 13), gleekCard(domain.CardDesignHeart, 12),
		gleekCard(domain.CardDesignHeart, 11), gleekCard(domain.CardDesignHeart, 10))
	setGleekHand(h, 1,
		gleekCard(domain.CardDesignHeart, 1), gleekCard(domain.CardDesignSpade, 1),
		gleekCard(domain.CardDesignClover, 1), gleekCard(domain.CardDesignDiamond, 1))
	setGleekHand(h, 2, gleekCard(domain.CardDesignClover, 4))
	h.SetBuyerIdx(0)
	h.SetPlayerScores([domain.GleekPlayerCnt]int{})
	h.StartPlayForTest()
	assert.Equal(t, 1, h.GetRuffWinnerIdx())
}

func TestGleek_MeldsPayEachOpponentAndMournivalDoublesTheGleek(t *testing.T) {
	g := newTestGleek()
	g.SetTrumpSuit(domain.CardDesignSpade)
	setGleekHand(g, 0,
		gleekCard(domain.CardDesignSpade, 13), gleekCard(domain.CardDesignClover, 13),
		gleekCard(domain.CardDesignHeart, 13), gleekCard(domain.CardDesignDiamond, 4))
	setGleekHand(g, 1,
		gleekCard(domain.CardDesignSpade, 11), gleekCard(domain.CardDesignClover, 11),
		gleekCard(domain.CardDesignHeart, 11), gleekCard(domain.CardDesignDiamond, 11))
	setGleekHand(g, 2, gleekCard(domain.CardDesignClover, 5))
	g.SetBuyerIdx(0)
	g.SetPlayerScores([domain.GleekPlayerCnt]int{})
	g.ScoreMeldsForTest()

	melds := g.GetMelds()
	require.Len(t, melds, 2)
	byPlayer := map[int]*domain.GleekMeld{}
	for _, m := range melds {
		byPlayer[m.PlayerIdx] = m
	}
	require.Contains(t, byPlayer, 0)
	require.Contains(t, byPlayer, 1)
	assert.Equal(t, 3, byPlayer[0].Count, "K が 3 枚 = グリーク")
	assert.Equal(t, 3, byPlayer[0].Value, "K のグリークは 3")
	assert.Equal(t, 4, byPlayer[1].Count, "J が 4 枚 = マーニヴァル")
	assert.Equal(t, 2, byPlayer[1].Value, "J のマーニヴァルはグリーク 1 の倍")

	scores := g.GetPlayerScores()
	assert.Equal(t, 3*2-2, scores[0], "K グリークで +6、J マーニヴァルに -2")
	assert.Equal(t, 2*2-3, scores[1], "J マーニヴァルで +4、K グリークに -3")
	assert.Equal(t, -3-2, scores[2])
	assert.Zero(t, gleekSum(scores))
}

func TestGleek_TwoCardsOfARankAreNotAMeld(t *testing.T) {
	g := newTestGleek()
	g.SetTrumpSuit(domain.CardDesignSpade)
	for i := 0; i < domain.GleekPlayerCnt; i++ {
		setGleekHand(g, i, gleekCard(domain.CardDesignSpade, 4+i), gleekCard(domain.CardDesignClover, 4+i))
	}
	setGleekHand(g, 0, gleekCard(domain.CardDesignSpade, 1), gleekCard(domain.CardDesignClover, 1))
	g.ScoreMeldsForTest()
	assert.Empty(t, g.GetMelds())
}

// --- play ----------------------------------------------------------------

func TestGleek_MustFollowTheLedSuit(t *testing.T) {
	g := newTestGleek()
	g.SetTrumpSuit(domain.CardDesignSpade)
	g.SetPhase(domain.GleekPhasePlay)
	setGleekHand(g, 1,
		gleekCard(domain.CardDesignHeart, 9), gleekCard(domain.CardDesignSpade, 1),
		gleekCard(domain.CardDesignClover, 10))
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: gleekCard(domain.CardDesignHeart, 13)},
	})
	g.SetCurrentPlayerIdx(1)
	assert.Equal(t, []int{0}, g.GetPlayableIndices(1), "ハートを持っていればハートしか出せない")

	setGleekHand(g, 1, gleekCard(domain.CardDesignSpade, 1), gleekCard(domain.CardDesignClover, 10))
	assert.Equal(t, []int{0, 1}, g.GetPlayableIndices(1), "ボイドなら何でも出せる")
}

func TestGleek_TrumpBeatsTheLedSuitAndHonoursScore(t *testing.T) {
	g := newTestGleek()
	trump := domain.CardDesignSpade
	g.SetTrumpSuit(trump)
	g.SetPhase(domain.GleekPhaseTrickEnd)
	g.SetTrickNumber(1)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: gleekCard(domain.CardDesignHeart, 1)}, // 平札の A
		{PlayerIdx: 1, Card: gleekCard(trump, 4)},                  // Tiddy = 切り札の 4
		{PlayerIdx: 2, Card: gleekCard(domain.CardDesignHeart, 13)},
	})
	g.ResolveTrick()
	assert.Equal(t, 1, g.GetPlayer(1).GetTrickCount(), "切り札が平札の A に勝つ")
	// 3 点 + Tiddy 4 点。
	assert.Equal(t, domain.GleekTrickPoints+4, g.GetTrickPoints()[1])
}

func TestGleek_OffSuitCardsCannotWin(t *testing.T) {
	g := newTestGleek()
	g.SetTrumpSuit(domain.CardDesignSpade)
	g.SetPhase(domain.GleekPhaseTrickEnd)
	g.SetTrickNumber(1)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: gleekCard(domain.CardDesignHeart, 4)},
		{PlayerIdx: 1, Card: gleekCard(domain.CardDesignClover, 1)}, // 場外の A
		{PlayerIdx: 2, Card: gleekCard(domain.CardDesignDiamond, 1)},
	})
	g.ResolveTrick()
	assert.Equal(t, 1, g.GetPlayer(0).GetTrickCount(), "リードスートの 4 が場外の A に勝つ")
}

// **同じトリックを二度精算しない。**
func TestGleek_ResolveTrickIsIdempotent(t *testing.T) {
	g := newTestGleek()
	g.SetTrumpSuit(domain.CardDesignSpade)
	g.SetPhase(domain.GleekPhaseTrickEnd)
	g.SetTrickNumber(1)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: gleekCard(domain.CardDesignHeart, 13)},
		{PlayerIdx: 1, Card: gleekCard(domain.CardDesignHeart, 9)},
		{PlayerIdx: 2, Card: gleekCard(domain.CardDesignHeart, 8)},
	})
	g.ResolveTrick()
	g.ResolveTrick()
	total := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		total += g.GetPlayer(i).GetTrickCount()
	}
	assert.Equal(t, 1, total)
	assert.Equal(t, domain.GleekTrickPoints, g.GetTrickPoints()[0])
}

// --- whole deals ---------------------------------------------------------

// **卓が最後まで回ることを見る。** 段階が 4 つあるので、どこか 1 つでも入力面が
// 欠けると盤が止まる。
func TestGleek_PlaysOutFullMatches(t *testing.T) {
	for attempt := 0; attempt < 20; attempt++ {
		g := newTestGleek()
		for step := 0; step < 6000 && !g.GetGameEndFlag(); step++ {
			switch g.GetPhase() {
			case domain.GleekPhaseBid:
				if g.IsHumanBidTurn() {
					require.NoError(t, g.PlayerBid(0))
					continue
				}
				g.CpuBid()
			case domain.GleekPhaseDiscard:
				if g.IsHumanDiscardTurn() {
					drop := g.GetDiscardHint()
					require.Len(t, drop, domain.GleekSwapSize)
					require.NoError(t, g.PlayerDiscard(drop))
					continue
				}
				g.CpuDiscard()
			case domain.GleekPhasePlay:
				if g.IsHumanTurn() {
					valid := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
					require.NotEmpty(t, valid)
					require.NoError(t, g.PlayerPlay(valid[0]))
					continue
				}
				g.CpuPlay()
			case domain.GleekPhaseTrickEnd:
				g.ResolveTrick()
				g.NextTrick()
			case domain.GleekPhaseRoundEnd:
				g.NextRound()
			case domain.GleekPhaseGameEnd:
			}
		}
		require.True(t, g.GetGameEndFlag(), "マッチが終わらない")
		// **競り・ラフ・メルドはすべて相手との受け渡しなので厳密にゼロ和。**
		// トリックの精算だけは、そのディールの合計を人数で割った余り (0..2) が
		// 残る。5 ディールでも卓の合計はその程度しか動かない。
		total := gleekSum(g.GetPlayerScores())
		assert.LessOrEqual(t, total, (domain.GleekPlayerCnt-1)*g.GetRoundNumber(),
			"卓の点が想定以上に増えている")
		assert.GreaterOrEqual(t, total, 0, "卓の点が減ってはいけない")
	}
}

// **12 トリックで必ず 81 点が配られる。** 名札を数え落とすと基準点との差が
// 合わなくなる。
func TestGleek_EveryDealDistributesTheFullDealPoints(t *testing.T) {
	for attempt := 0; attempt < 20; attempt++ {
		g := newTestGleek()
		for step := 0; step < 3000 && g.GetPhase() != domain.GleekPhaseRoundEnd; step++ {
			switch g.GetPhase() {
			case domain.GleekPhaseBid:
				if g.IsHumanBidTurn() {
					require.NoError(t, g.PlayerBid(0))
					continue
				}
				g.CpuBid()
			case domain.GleekPhaseDiscard:
				if g.IsHumanDiscardTurn() {
					require.NoError(t, g.PlayerDiscard(g.GetDiscardHint()))
					continue
				}
				g.CpuDiscard()
			case domain.GleekPhasePlay:
				if g.IsHumanTurn() {
					valid := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
					require.NotEmpty(t, valid)
					require.NoError(t, g.PlayerPlay(valid[0]))
					continue
				}
				g.CpuPlay()
			case domain.GleekPhaseTrickEnd:
				g.ResolveTrick()
				g.NextTrick()
			}
		}
		require.Equal(t, domain.GleekPhaseRoundEnd, g.GetPhase())
		total := 0
		for _, v := range g.GetTrickPoints() {
			total += v
		}
		assert.Equal(t, g.DealPoints(), total)
		// トリック点 36 は必ず配られ、名札は場外に落ちた分だけ欠ける。
		assert.GreaterOrEqual(t, total, domain.GleekTrickCount*domain.GleekTrickPoints)
		assert.LessOrEqual(t, total, domain.GleekMaxDealPoints)
	}
}

// --- JSON ----------------------------------------------------------------

func TestGleek_JSONRoundTripKeepsTheStageState(t *testing.T) {
	g := newTestGleek()
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetBuyerIdx(1)
	g.SetPhase(domain.GleekPhasePlay)
	g.SetLeadPlayerIdx(0)
	g.ScoreMeldsForTest()

	data, err := json.Marshal(g)
	require.NoError(t, err)
	restored := new(domain.Gleek)
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, domain.CardDesignHeart, restored.GetTrumpSuit())
	assert.Equal(t, 1, restored.GetBuyerIdx())
	assert.Equal(t, g.GetPlayerScores(), restored.GetPlayerScores())
	assert.Equal(t, len(g.GetMelds()), len(restored.GetMelds()))
}

func TestGleek_JSONRejectsOutOfRangeValues(t *testing.T) {
	base := newTestGleek()
	base.SetBuyerIdx(0)
	base.SetPhase(domain.GleekPhasePlay)
	base.SetLeadPlayerIdx(0)
	data, err := json.Marshal(base)
	require.NoError(t, err)

	var raw map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &raw))

	for name, tc := range map[string]struct{ field, value string }{
		"phase":  {"ph", "99"},
		"trump":  {"ts", "-1"},
		"buyer":  {"bi", "-1"},
		"lead":   {"li", "-1"},
		"result": {"rs", "9"},
	} {
		t.Run(name, func(t *testing.T) {
			mutated := map[string]json.RawMessage{}
			for k, v := range raw {
				mutated[k] = v
			}
			mutated[tc.field] = json.RawMessage(tc.value)
			body, err := json.Marshal(mutated)
			require.NoError(t, err)
			require.Error(t, json.Unmarshal(body, new(domain.Gleek)))
		})
	}
}

// --- helpers -------------------------------------------------------------

func gleekSum(s [domain.GleekPlayerCnt]int) int {
	total := 0
	for _, v := range s {
		total += v
	}
	return total
}

func gleekCardKey(c *domain.Card) string {
	return string(rune('A'+c.GetDesign())) + string(rune('a'+c.GetValue()))
}

func gleekContainsInt(list []int, v int) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}
