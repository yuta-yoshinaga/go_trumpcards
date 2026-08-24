//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// newTestGermanSolo returns a fresh, reset GermanSolo game (human seat 0 + 3 CPUs).
func newTestGermanSolo() *domain.GermanSolo {
	g := domain.NewDefaultGermanSolo()
	g.Reset()
	return g
}

// germanSoloCard is a shorthand constructor for a face-up card.
func germanSoloCard(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

// setGermanSoloHand replaces player i's hand with the supplied cards deterministically.
func setGermanSoloHand(g *domain.GermanSolo, i int, cards ...*domain.Card) {
	p := g.GetPlayer(i)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// germanSoloGiveTricks awards player i exactly n tricks by adding dummy trick groups.
func germanSoloGiveTricks(g *domain.GermanSolo, i, n int) {
	p := g.GetPlayer(i)
	for k := 0; k < n; k++ {
		p.AddTrick([]*domain.Card{germanSoloCard(domain.CardDesignHeart, 8)})
	}
}

func TestGermanSolo_ResetDeal(t *testing.T) {
	g := newTestGermanSolo()
	assert.Equal(t, domain.GermanSoloPhaseBid, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, domain.GermanSoloPlayerCnt, g.GetPlayerCnt())
	assert.Equal(t, -1, g.GetDeclarerIdx())
	assert.Equal(t, -1, g.GetTrumpSuit())
	assert.Equal(t, domain.GermanSoloBidNone, g.GetWinningBid())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, -1, g.GetWinnerPlayer())

	total := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		assert.Equal(t, domain.GermanSoloHandSize, g.GetPlayer(i).GetCardsSize())
		total += g.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, domain.GermanSoloDeckSize, total, "32 枚を配り切りストックは残らない")
	assert.Equal(t, domain.GermanSoloHandSize, domain.GermanSoloTrickCount,
		"配った枚数とトリック数が食い違うと手札が余る")
}

// **人間が最初に喋る。** ディーラーが席 0 のままだと人間は毎回最後の宣言者になり、
// 開幕ディールで高い契約を選べない。
func TestGermanSolo_ForehandIsTheHumanOnTheOpeningDeal(t *testing.T) {
	g := newTestGermanSolo()
	assert.Equal(t, domain.GermanSoloPlayerCnt-1, g.GetDealerIdx())
	assert.Equal(t, 0, g.GetForehandIdx())
	assert.Equal(t, 0, g.GetCurrentBidderIdx())
	assert.True(t, g.IsHumanBidTurn())
}

func TestGermanSolo_DeckIs32SkatPack(t *testing.T) {
	g := newTestGermanSolo()
	seen := map[int]bool{}
	count := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		p := g.GetPlayer(i)
		for j := 0; j < p.GetCardsSize(); j++ {
			c := p.GetCard(j)
			key := c.GetDesign()*100 + c.GetValue()
			assert.False(t, seen[key], "duplicate card %d", key)
			seen[key] = true
			count++
		}
	}
	assert.Equal(t, domain.GermanSoloDeckSize, count)
	// A,K,Q,J,10,9,8,7 のみ。2..6 が 1 枚でも混ざると 7 が Manille でなくなる。
	valid := map[int]bool{1: true, 7: true, 8: true, 9: true, 10: true, 11: true, 12: true, 13: true}
	for k := range seen {
		assert.True(t, valid[k%100], "unexpected rank %d", k%100)
	}
}

// --- card ranking -------------------------------------------------------

func TestGermanSolo_MatadorRankIsFixedRegardlessOfTrump(t *testing.T) {
	for _, trump := range []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond} {
		assert.Equal(t, 1, domain.GermanSoloMatadorRank(germanSoloCard(domain.CardDesignClover, 12), trump),
			"♣Q は常に Spadille")
		assert.Equal(t, 2, domain.GermanSoloMatadorRank(germanSoloCard(trump, 7), trump),
			"切り札の 7 は常に Manille")
		assert.Equal(t, 3, domain.GermanSoloMatadorRank(germanSoloCard(domain.CardDesignSpade, 12), trump),
			"♠Q は常に Basta")
		assert.Equal(t, 0, domain.GermanSoloMatadorRank(germanSoloCard(domain.CardDesignHeart, 13), trump))
	}
	// 切り札未確定ならマニーユは決まらないので一切示さない。
	assert.Equal(t, 0, domain.GermanSoloMatadorRank(germanSoloCard(domain.CardDesignClover, 12), -1))
	assert.Equal(t, 0, domain.GermanSoloMatadorRank(nil, domain.CardDesignHeart))
}

// trickBetween plays a two-card trick and reports the winner seat.
func trickBetween(t *testing.T, trump int, a, b *domain.Card) int {
	t.Helper()
	g := newTestGermanSolo()
	g.SetTrumpSuit(trump)
	g.SetDeclarerIdx(0)
	g.SetPhase(domain.GermanSoloPhasePlay)
	setGermanSoloHand(g, 0, a)
	setGermanSoloHand(g, 1, b)
	g.SetLeadPlayerIdx(0)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: a},
		{PlayerIdx: 1, Card: b},
	})
	g.SetCurrentPlayerIdx(2)
	g.SetPhase(domain.GermanSoloPhaseTrickEnd)
	// 4 人分そろっていないと ResolveTrick は何もしないので、残り 2 席を弱い平札で埋める。
	filler := []*domain.TrickCard{
		{PlayerIdx: 2, Card: germanSoloCard(domain.CardDesignDiamond, 8)},
		{PlayerIdx: 3, Card: germanSoloCard(domain.CardDesignDiamond, 9)},
	}
	g.SetCurrentTrick(append(g.GetCurrentTrick(), filler...))
	g.SetTrickNumber(1)
	g.ResolveTrick()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if g.GetPlayer(i).GetTrickCount() > 0 {
			return i
		}
	}
	return -1
}

func TestGermanSolo_TrumpOrderIsSpadilleManilleBasta(t *testing.T) {
	trump := domain.CardDesignHeart
	spadille := germanSoloCard(domain.CardDesignClover, 12)
	manille := germanSoloCard(domain.CardDesignHeart, 7)
	basta := germanSoloCard(domain.CardDesignSpade, 12)
	trumpAce := germanSoloCard(domain.CardDesignHeart, 1)

	assert.Equal(t, 0, trickBetween(t, trump, spadille, manille), "Spadille > Manille")
	assert.Equal(t, 0, trickBetween(t, trump, manille, basta), "Manille > Basta")
	assert.Equal(t, 0, trickBetween(t, trump, basta, trumpAce), "Basta > 切り札の A")
	assert.Equal(t, 1, trickBetween(t, trump, trumpAce, spadille), "Spadille はどこから出ても最強")
	// 黒の Q はスペード/クラブが平札でも切り札として平札の A に勝つ。
	assert.Equal(t, 1, trickBetween(t, trump, germanSoloCard(domain.CardDesignDiamond, 1), basta),
		"♠Q は切り札なので平札の A に勝つ")
}

func TestGermanSolo_PlainSuitRanksAceHighSevenLow(t *testing.T) {
	trump := domain.CardDesignSpade
	// 同じ平札スート内では A > K > Q > J > 10 > 9 > 8 > 7。
	order := []int{1, 13, 12, 11, 10, 9, 8, 7}
	for i := 0; i+1 < len(order); i++ {
		hi := germanSoloCard(domain.CardDesignHeart, order[i])
		lo := germanSoloCard(domain.CardDesignHeart, order[i+1])
		assert.Equal(t, 0, trickBetween(t, trump, hi, lo), "rank %d > %d", order[i], order[i+1])
	}
}

// **切り札が赤なら Q は普通の切り札として残る。** ここを黒と同じ扱いにすると、
// 赤切り札の Q が平札に落ちて切り札が 1 枚減る。
func TestGermanSolo_RedTrumpQueenIsAnOrdinaryTrump(t *testing.T) {
	trump := domain.CardDesignHeart
	trumpQueen := germanSoloCard(domain.CardDesignHeart, 12)
	trumpKing := germanSoloCard(domain.CardDesignHeart, 13)
	plainAce := germanSoloCard(domain.CardDesignDiamond, 1)
	assert.Equal(t, 1, trickBetween(t, trump, trumpQueen, trumpKing), "切り札 K > 切り札 Q")
	assert.Equal(t, 0, trickBetween(t, trump, trumpQueen, plainAce), "切り札 Q は平札の A に勝つ")
}

// --- bidding ------------------------------------------------------------

func TestGermanSolo_MussfrageCannotBeDeclared(t *testing.T) {
	g := newTestGermanSolo()
	err := g.PlayerBid(domain.GermanSoloBidMussfrage, domain.CardDesignHeart)
	require.Error(t, err, "Mussfrage は卓が押し付ける契約で、宣言はできない")
	assert.Equal(t, domain.GermanSoloBidNone, g.GetHighestBid())
}

func TestGermanSolo_BidMustExceedTheStandingBid(t *testing.T) {
	g := newTestGermanSolo()
	require.NoError(t, g.PlayerBid(domain.GermanSoloBidSolo, domain.CardDesignHeart))
	assert.Equal(t, domain.GermanSoloBidSolo, g.GetHighestBid())
	assert.Equal(t, []int{int(domain.GermanSoloBidTout)}, g.GetBiddableBids(),
		"Solo が立っていれば残る選択肢は Tout だけ")
	// 人間の後は CPU が順に宣言し、全員が喋ると競りが閉じる。
	for g.GetPhase() == domain.GermanSoloPhaseBid {
		g.CpuBid()
	}
	assert.True(t, g.GetWinningBid() >= domain.GermanSoloBidSolo,
		"確定した契約が Solo を下回ってはいけない")
}

func TestGermanSolo_BidRequiresATrumpSuit(t *testing.T) {
	g := newTestGermanSolo()
	err := g.PlayerBid(domain.GermanSoloBidFrage, 0)
	require.Error(t, err)
	assert.Equal(t, domain.GermanSoloPhaseBid, g.GetPhase())
	// パスには切り札が要らない。
	require.NoError(t, g.PlayerBid(domain.GermanSoloBidNone, -1))
}

func TestGermanSolo_BiddableBidsShrinkAsTheAuctionRises(t *testing.T) {
	g := newTestGermanSolo()
	assert.Equal(t, []int{
		int(domain.GermanSoloBidFrage), int(domain.GermanSoloBidSolo), int(domain.GermanSoloBidTout),
	}, g.GetBiddableBids(), "開幕は 3 段すべて宣言できる")

	g.SetPhase(domain.GermanSoloPhasePlay)
	assert.Nil(t, g.GetBiddableBids(), "競りの外では選択肢を出さない")
}

// **全員パスは配り直しではなく Mussfrage。** 配り直しにすると CPU が
// ビッドしない手を引き続ける限りディールが終わらない。
func TestGermanSolo_AllPassForcesTheMussfrageOnTheSpadilleHolder(t *testing.T) {
	g := newTestGermanSolo()
	cfg := g.GetConfig()
	cfg.CpuDifficulty = domain.GermanSoloCpuDifficultyEasy // Easy CPU は常にパスする
	g.SetConfig(cfg)
	g.Reset()

	// 席 2 に Spadille (♣Q) を握らせる。
	for i := 0; i < g.GetPlayerCnt(); i++ {
		base := domain.CardDesignDiamond
		setGermanSoloHand(g, i,
			germanSoloCard(base, 8), germanSoloCard(base, 9), germanSoloCard(base, 10),
			germanSoloCard(base, 11), germanSoloCard(domain.CardDesignHeart, 8),
			germanSoloCard(domain.CardDesignHeart, 9), germanSoloCard(domain.CardDesignHeart, 10),
			germanSoloCard(domain.CardDesignHeart, 11))
	}
	setGermanSoloHand(g, 2,
		germanSoloCard(domain.CardDesignClover, 12), germanSoloCard(domain.CardDesignClover, 8),
		germanSoloCard(domain.CardDesignClover, 9), germanSoloCard(domain.CardDesignClover, 10),
		germanSoloCard(domain.CardDesignClover, 11), germanSoloCard(domain.CardDesignSpade, 8),
		germanSoloCard(domain.CardDesignSpade, 9), germanSoloCard(domain.CardDesignSpade, 10))

	require.NoError(t, g.PlayerBid(domain.GermanSoloBidNone, -1))
	for g.GetPhase() == domain.GermanSoloPhaseBid {
		g.CpuBid()
	}
	assert.Equal(t, domain.GermanSoloBidMussfrage, g.GetWinningBid())
	assert.Equal(t, 2, g.GetDeclarerIdx(), "♣Q を持つ席が引き受ける")
	assert.Equal(t, domain.GermanSoloMakeTricks, g.RequiredTricks())
}

// --- ace call -----------------------------------------------------------

func TestGermanSolo_SoloAndToutSkipTheAceCall(t *testing.T) {
	for _, bid := range []domain.GermanSoloBid{domain.GermanSoloBidSolo, domain.GermanSoloBidTout} {
		g := newTestGermanSolo()
		require.NoError(t, g.PlayerBid(bid, domain.CardDesignHeart))
		for g.GetPhase() == domain.GermanSoloPhaseBid {
			g.CpuBid()
		}
		if g.GetWinningBid() != bid {
			continue // CPU がさらに上を宣言した (Tout の上は無いので Solo のときだけ)
		}
		assert.NotEqual(t, domain.GermanSoloPhaseAceCall, g.GetPhase(), "単独契約はエースを呼ばない")
		assert.True(t, g.IsPlayingAlone())
		assert.Equal(t, 1, g.GetDeclarerSideSize())
		assert.Equal(t, -1, g.GetCalledAceSuit())
	}
}

func TestGermanSolo_RequiredTricksIsEightOnlyForTout(t *testing.T) {
	g := newTestGermanSolo()
	g.SetWinningBid(domain.GermanSoloBidTout)
	assert.Equal(t, domain.GermanSoloTrickCount, g.RequiredTricks())
	for _, bid := range []domain.GermanSoloBid{domain.GermanSoloBidMussfrage, domain.GermanSoloBidFrage, domain.GermanSoloBidSolo} {
		g.SetWinningBid(bid)
		assert.Equal(t, domain.GermanSoloMakeTricks, g.RequiredTricks())
	}
}

// **呼べるのは自分が持っていない・切り札以外のエース。** 切り札の A を呼べると
// 味方探しでなく上から 4 枚目の切り札の補充になる。
func TestGermanSolo_CallableAcesExcludeHeldAndTrumpSuit(t *testing.T) {
	g := newTestGermanSolo()
	g.SetTrumpSuit(domain.CardDesignHeart)
	setGermanSoloHand(g, 0,
		germanSoloCard(domain.CardDesignSpade, 1), // 自分で持っている ♠A
		germanSoloCard(domain.CardDesignHeart, 7), germanSoloCard(domain.CardDesignHeart, 8),
		germanSoloCard(domain.CardDesignHeart, 9), germanSoloCard(domain.CardDesignHeart, 10),
		germanSoloCard(domain.CardDesignClover, 8), germanSoloCard(domain.CardDesignClover, 9),
		germanSoloCard(domain.CardDesignDiamond, 8))
	callable := g.GetCallableAceSuitsForTest(0)
	assert.NotContains(t, callable, domain.CardDesignSpade, "持っているエースは呼べない")
	assert.NotContains(t, callable, domain.CardDesignHeart, "切り札のエースは呼べない")
	assert.ElementsMatch(t, []int{domain.CardDesignClover, domain.CardDesignDiamond}, callable)
}

func TestGermanSolo_DeclareAceRejectsBadInput(t *testing.T) {
	g := newTestGermanSolo()
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetDeclarerIdx(0)
	g.SetPhase(domain.GermanSoloPhasePlay)
	require.Error(t, g.DeclareAce(0, domain.CardDesignClover), "プレイ中は呼べない")

	g.SetPhase(domain.GermanSoloPhaseAceCall)
	require.Error(t, g.DeclareAce(1, domain.CardDesignClover), "落札者以外は呼べない")
	require.Error(t, g.DeclareAce(0, 0), "スートが不正")
	require.Error(t, g.DeclareAce(0, domain.CardDesignHeart), "切り札のエースは呼べない")
}

// **味方はそのエースが場に出るまで伏せる。** ここが公開されると
// Frage の駆け引きが最初のトリックで終わる。
func TestGermanSolo_PartnerStaysHiddenUntilTheCalledAceIsPlayed(t *testing.T) {
	g := newTestGermanSolo()
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetDeclarerIdx(0)
	g.SetWinningBid(domain.GermanSoloBidFrage)
	// **配られた札を残さない。** 席 2 に ♣A を持たせても、配りで別の席に
	// 入った ♣A がそのままなら findAceHolder は先に見つけた方を味方にする。
	// 4 席とも置き換えて、♣A が卓に 1 枚だけになるようにする。
	setGermanSoloHand(g, 0,
		germanSoloCard(domain.CardDesignHeart, 7), germanSoloCard(domain.CardDesignHeart, 8))
	setGermanSoloHand(g, 1, germanSoloCard(domain.CardDesignSpade, 9))
	setGermanSoloHand(g, 2, germanSoloCard(domain.CardDesignClover, 1))
	setGermanSoloHand(g, 3, germanSoloCard(domain.CardDesignDiamond, 9))
	g.SetPhase(domain.GermanSoloPhaseAceCall)
	require.NoError(t, g.DeclareAce(0, domain.CardDesignClover))

	assert.Equal(t, domain.CardDesignClover, g.GetCalledAceSuit())
	assert.Equal(t, -1, g.GetPartnerIdx(), "まだ伏せられている")
	assert.False(t, g.IsPlayingAlone())
	assert.Equal(t, 2, g.GetDeclarerSideSize(), "伏せていても勝敗の計算では味方として数える")

	// 呼ばれたエースが出た瞬間に公開される。
	g.SetPhase(domain.GermanSoloPhasePlay)
	g.SetCurrentTrick(nil)
	g.SetCurrentPlayerIdx(2)
	g.CpuPlay() // 席 2 の手札は呼ばれた ♣A の 1 枚だけなので必ずそれが出る
	assert.Equal(t, 2, g.GetPartnerIdx())
}

// --- scoring ------------------------------------------------------------

// germanSoloScoreRound runs a deal end with the supplied trick split and returns the scores.
func germanSoloScoreRound(t *testing.T, bid domain.GermanSoloBid, partner int, declarerTricks int) [domain.GermanSoloPlayerCnt]int {
	t.Helper()
	g := newTestGermanSolo()
	g.SetDeclarerIdx(0)
	g.SetWinningBid(bid)
	g.SetTrumpSuit(domain.CardDesignHeart)
	if partner >= 0 {
		g.SetCalledAceSuitForTest(domain.CardDesignClover)
		g.SetPartnerForTest(partner, true)
	} else {
		g.SetPlaysAloneForTest(true)
	}
	remaining := declarerTricks
	if partner >= 0 {
		half := remaining / 2
		germanSoloGiveTricks(g, 0, remaining-half)
		germanSoloGiveTricks(g, partner, half)
	} else {
		germanSoloGiveTricks(g, 0, remaining)
	}
	defenders := 0
	for i := 0; i < domain.GermanSoloPlayerCnt; i++ {
		if i == 0 || i == partner {
			continue
		}
		if defenders < domain.GermanSoloTrickCount-declarerTricks {
			germanSoloGiveTricks(g, i, domain.GermanSoloTrickCount-declarerTricks-defenders)
			defenders = domain.GermanSoloTrickCount - declarerTricks
		}
	}
	g.SetPhase(domain.GermanSoloPhaseRoundEnd)
	g.ScoreRound()
	return g.GetPlayerScores()
}

func TestGermanSolo_SoloMadePaysThreeTimesTheContract(t *testing.T) {
	scores := germanSoloScoreRound(t, domain.GermanSoloBidSolo, -1, 5)
	assert.Equal(t, 12, scores[0], "Solo の契約値 4 × 守備 3 人")
	for i := 1; i < domain.GermanSoloPlayerCnt; i++ {
		assert.Equal(t, -4, scores[i])
	}
	assert.Equal(t, 0, sumScores(scores), "点はゼロ和で動く")
}

func TestGermanSolo_SoloFailedReversesTheSign(t *testing.T) {
	scores := germanSoloScoreRound(t, domain.GermanSoloBidSolo, -1, 4)
	assert.Equal(t, -12, scores[0], "4 トリックでは 5 に届かない")
	for i := 1; i < domain.GermanSoloPlayerCnt; i++ {
		assert.Equal(t, 4, scores[i])
	}
	assert.Equal(t, 0, sumScores(scores))
}

func TestGermanSolo_FrageSplitsTheStakeBetweenThePartners(t *testing.T) {
	scores := germanSoloScoreRound(t, domain.GermanSoloBidFrage, 2, 5)
	assert.Equal(t, 2, scores[0], "Frage の契約値 2 × 守備 2 人 ÷ 宣言側 2 人")
	assert.Equal(t, 2, scores[2], "味方も同じ側の点を受け取る")
	assert.Equal(t, -2, scores[1])
	assert.Equal(t, -2, scores[3])
	assert.Equal(t, 0, sumScores(scores))
}

// **Tout は 8 取らないと失敗。** 7 で成功にすると、8 点払う最上位契約が
// Solo と同じ条件になり階段が壊れる。
func TestGermanSolo_ToutNeedsEveryTrick(t *testing.T) {
	seven := germanSoloScoreRound(t, domain.GermanSoloBidTout, -1, 7)
	assert.Equal(t, -24, seven[0], "7 トリックでも失敗")
	eight := germanSoloScoreRound(t, domain.GermanSoloBidTout, -1, 8)
	assert.Equal(t, 24, eight[0], "8 トリックで成功")
}

func TestGermanSolo_MussfrageMovesTheSmallestStake(t *testing.T) {
	scores := germanSoloScoreRound(t, domain.GermanSoloBidMussfrage, 2, 5)
	assert.Equal(t, 1, scores[0], "Mussfrage の契約値は 1")
	assert.Equal(t, -1, scores[1])
}

func sumScores(s [domain.GermanSoloPlayerCnt]int) int {
	total := 0
	for _, v := range s {
		total += v
	}
	return total
}

// --- play ---------------------------------------------------------------

func TestGermanSolo_MustFollowTheLeadSuitWithTrumpsAsOneSuit(t *testing.T) {
	g := newTestGermanSolo()
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetDeclarerIdx(0)
	g.SetPhase(domain.GermanSoloPhasePlay)
	setGermanSoloHand(g, 1,
		germanSoloCard(domain.CardDesignDiamond, 9), // リードスート
		germanSoloCard(domain.CardDesignClover, 12), // Spadille = 切り札
		germanSoloCard(domain.CardDesignSpade, 10))
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: germanSoloCard(domain.CardDesignDiamond, 13)},
	})
	g.SetCurrentPlayerIdx(1)
	assert.Equal(t, []int{0}, g.GetPlayableIndices(1), "ダイヤを持っていればダイヤしか出せない")

	// 切り札リードには切り札グループ (♣Q・♠Q・切り札の 7 を含む) で従う。
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: germanSoloCard(domain.CardDesignHeart, 13)},
	})
	assert.Equal(t, []int{1}, g.GetPlayableIndices(1), "♣Q は切り札なのでフォロー義務がある")
}

// **リードは forehand。** 落札者が先に打てると、単独契約でも自分の切り札から
// 好きな順に叩ける。
func TestGermanSolo_ForehandLeadsTheFirstTrick(t *testing.T) {
	g := newTestGermanSolo()
	require.NoError(t, g.PlayerBid(domain.GermanSoloBidTout, domain.CardDesignHeart))
	for g.GetPhase() == domain.GermanSoloPhaseBid {
		g.CpuBid()
	}
	if g.GetPhase() == domain.GermanSoloPhaseAceCall {
		g.CpuDeclareAce()
	}
	require.Equal(t, domain.GermanSoloPhasePlay, g.GetPhase())
	assert.Equal(t, g.GetForehandIdx(), g.GetLeadPlayerIdx())
}

// --- CPU / hint ---------------------------------------------------------

// **届かない契約には乗らない。** 一段上を返すだけの実装だと、Frage がやっとの
// 手札が Tout を掴み、ヒントも同じ助言を返す。
func TestGermanSolo_CpuFoldsWhenTheHandCannotBeatTheStandingBid(t *testing.T) {
	g := newTestGermanSolo()
	// 弱い手 (切り札 2 枚・マタドール無し) を席 1 に持たせる。
	setGermanSoloHand(g, 1,
		germanSoloCard(domain.CardDesignHeart, 8), germanSoloCard(domain.CardDesignHeart, 9),
		germanSoloCard(domain.CardDesignSpade, 7), germanSoloCard(domain.CardDesignSpade, 8),
		germanSoloCard(domain.CardDesignSpade, 9), germanSoloCard(domain.CardDesignDiamond, 7),
		germanSoloCard(domain.CardDesignDiamond, 8), germanSoloCard(domain.CardDesignDiamond, 9))
	require.NoError(t, g.PlayerBid(domain.GermanSoloBidTout, domain.CardDesignClover))
	// 人間が Tout を宣言済み。席 1 の CPU はこの手札で上回れないのでパスするしかない。
	g.CpuBid()
	assert.Equal(t, domain.GermanSoloBidTout, g.GetHighestBid(),
		"弱い手が Tout の上に押し上げられてはいけない")
}

func TestGermanSolo_HintFollowsTheSameRuleAsTheCpu(t *testing.T) {
	g := newTestGermanSolo()
	hint := g.GetHint()
	require.NotNil(t, hint, "人間の宣言手番ではヒントが出る")
	assert.Contains(t, []string{"bid_pass", "bid_frage", "bid_solo", "bid_tout"}, hint.Reason)
}

func TestGermanSolo_HintSuggestsAnAceInTheAceCallPhase(t *testing.T) {
	g := newTestGermanSolo()
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetDeclarerIdx(0)
	// **手札を置き換える。** 配られたままだと、非切り札のエース 3 枚が
	// 全部この席に入ったディール (実測 1.1%) で呼べるエースが無くなり、
	// ヒントが nil になってテストだけが落ちる。
	setGermanSoloHand(g, 0,
		germanSoloCard(domain.CardDesignHeart, 7), germanSoloCard(domain.CardDesignHeart, 8),
		germanSoloCard(domain.CardDesignClover, 9), germanSoloCard(domain.CardDesignClover, 10),
		germanSoloCard(domain.CardDesignSpade, 8), germanSoloCard(domain.CardDesignSpade, 9),
		germanSoloCard(domain.CardDesignDiamond, 8), germanSoloCard(domain.CardDesignDiamond, 9))
	g.SetPhase(domain.GermanSoloPhaseAceCall)
	hint := g.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "call_ace", hint.Reason)
	assert.Contains(t, g.GetCallableAceSuitsForTest(0), hint.SuitHint,
		"勧めたスートは実際に呼べること")
}

// --- whole deals --------------------------------------------------------

// **卓が最後まで回ることを見る。** 単体の分岐が全部緑でも、フェーズが繋がって
// いなければ CUI も Web も最初の手番で固まる。
func TestGermanSolo_PlaysOutFullMatches(t *testing.T) {
	for attempt := 0; attempt < 30; attempt++ {
		g := newTestGermanSolo()
		for step := 0; step < 4000 && !g.GetGameEndFlag(); step++ {
			switch g.GetPhase() {
			case domain.GermanSoloPhaseBid:
				if g.IsHumanBidTurn() {
					require.NoError(t, g.PlayerBid(domain.GermanSoloBidNone, -1))
					continue
				}
				g.CpuBid()
			case domain.GermanSoloPhaseAceCall:
				if g.IsHumanAceCallTurn() {
					suits := g.GetCallableAceSuits()
					require.NotEmpty(t, suits)
					require.NoError(t, g.DeclareAce(g.GetDeclarerIdx(), suits[0]))
					continue
				}
				g.CpuDeclareAce()
			case domain.GermanSoloPhasePlay:
				if g.IsHumanTurn() {
					valid := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
					require.NotEmpty(t, valid, "出せる札が無い手番は作らない")
					require.NoError(t, g.PlayerPlay(valid[0]))
					continue
				}
				g.CpuPlay()
			case domain.GermanSoloPhaseTrickEnd:
				g.ResolveTrick()
				g.NextTrick()
			case domain.GermanSoloPhaseRoundEnd:
				g.NextRound()
			case domain.GermanSoloPhaseGameEnd:
			}
		}
		require.True(t, g.GetGameEndFlag(), "マッチが終わらない")
		assert.Equal(t, 0, sumScores(g.GetPlayerScores()), "累積点もゼロ和のまま")
	}
}

// --- JSON ---------------------------------------------------------------

func TestGermanSolo_JSONRoundTripKeepsTheAceCall(t *testing.T) {
	g := newTestGermanSolo()
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetDeclarerIdx(0)
	g.SetWinningBid(domain.GermanSoloBidFrage)
	g.SetCalledAceSuitForTest(domain.CardDesignClover)
	g.SetPartnerForTest(2, true)
	g.SetPhase(domain.GermanSoloPhasePlay)
	g.SetLeadPlayerIdx(0)

	data, err := json.Marshal(g)
	require.NoError(t, err)
	restored := new(domain.GermanSolo)
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, domain.CardDesignClover, restored.GetCalledAceSuit())
	assert.Equal(t, 2, restored.GetPartnerIdx())
	assert.False(t, restored.IsPlayingAlone())
	assert.Equal(t, 2, restored.GetDeclarerSideSize())
	assert.Equal(t, domain.GermanSoloBidFrage, restored.GetWinningBid())
	assert.Equal(t, domain.CardDesignHeart, restored.GetTrumpSuit())
}

func TestGermanSolo_JSONRejectsOutOfRangeValues(t *testing.T) {
	base := newTestGermanSolo()
	data, err := json.Marshal(base)
	require.NoError(t, err)

	var raw map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &raw))

	for _, tc := range []struct {
		name  string
		field string
		value string
	}{
		{"bid", "wb", "99"},
		{"outcome", "oc", "99"},
		{"phase", "ph", "99"},
		{"trump", "ts", "9"},
		{"partner", "pi", "9"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mutated := map[string]json.RawMessage{}
			for k, v := range raw {
				mutated[k] = v
			}
			mutated[tc.field] = json.RawMessage(tc.value)
			body, err := json.Marshal(mutated)
			require.NoError(t, err)
			require.Error(t, json.Unmarshal(body, new(domain.GermanSolo)))
		})
	}
}

// **単独プレイに味方はいない。** 復元時に整合を取らないと、フィールドの
// 欠けた JSON が席 0 を味方として数え、勝敗の集計が静かに変わる。
func TestGermanSolo_JSONNormalisesAnInconsistentAceCall(t *testing.T) {
	g := newTestGermanSolo()
	g.SetDeclarerIdx(1)
	g.SetPlaysAloneForTest(true)
	g.SetPartnerForTest(0, true)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetLeadPlayerIdx(0)
	g.SetPhase(domain.GermanSoloPhasePlay)

	data, err := json.Marshal(g)
	require.NoError(t, err)
	restored := new(domain.GermanSolo)
	require.NoError(t, json.Unmarshal(data, restored))
	assert.True(t, restored.IsPlayingAlone())
	assert.Equal(t, -1, restored.GetPartnerIdx())
	assert.Equal(t, 1, restored.GetDeclarerSideSize())
}

// **Tout は「勝てる手」ではなく「1 枚も落とさない手」でしか宣言しない。**
// 評価値のしきい値で代用すると 5 回に 1 回しか成立しない手が 24 点を賭ける。
// 切り札序列の上から連続した 8 枚なら、リードを誰が取っても全部取れる。
func TestGermanSolo_ToutIsDeclaredOnlyOnAProvableSweep(t *testing.T) {
	g := newTestGermanSolo()
	trump := domain.CardDesignHeart
	sweep := []*domain.Card{
		germanSoloCard(domain.CardDesignClover, 12), // Spadille
		germanSoloCard(trump, 7),                    // Manille
		germanSoloCard(domain.CardDesignSpade, 12),  // Basta
		germanSoloCard(trump, 1), germanSoloCard(trump, 13), germanSoloCard(trump, 12),
		germanSoloCard(trump, 11), germanSoloCard(trump, 10),
	}
	setGermanSoloHand(g, 1, sweep...)
	g.CpuBidForTest(1)
	assert.Equal(t, domain.GermanSoloBidTout, g.GetHighestBid(), "上から連続 8 枚なら Tout")

	// 1 枚でも欠けると宣言しない: ♥10 を平札のエースに差し替える。
	g2 := newTestGermanSolo()
	broken := append([]*domain.Card{}, sweep[:len(sweep)-1]...)
	broken = append(broken, germanSoloCard(domain.CardDesignDiamond, 1))
	setGermanSoloHand(g2, 1, broken...)
	g2.CpuBidForTest(1)
	assert.NotEqual(t, domain.GermanSoloBidTout, g2.GetHighestBid(),
		"平札のエースは切り札で殺されるので Tout の根拠にならない")
}

// **宣言した Tout は実際に 8 トリック取れる。** 条件が机上のものでないことを、
// 卓を最後まで回して確かめる。
func TestGermanSolo_ASweepHandActuallyTakesEveryTrick(t *testing.T) {
	g := newTestGermanSolo()
	trump := domain.CardDesignHeart
	setGermanSoloHand(g, 0,
		germanSoloCard(domain.CardDesignClover, 12), germanSoloCard(trump, 7),
		germanSoloCard(domain.CardDesignSpade, 12), germanSoloCard(trump, 1),
		germanSoloCard(trump, 13), germanSoloCard(trump, 12),
		germanSoloCard(trump, 11), germanSoloCard(trump, 10))
	// 残り 3 席には切り札序列の下と平札を配る。
	setGermanSoloHand(g, 1,
		germanSoloCard(trump, 9), germanSoloCard(trump, 8),
		germanSoloCard(domain.CardDesignDiamond, 1), germanSoloCard(domain.CardDesignDiamond, 13),
		germanSoloCard(domain.CardDesignDiamond, 11), germanSoloCard(domain.CardDesignDiamond, 10),
		germanSoloCard(domain.CardDesignDiamond, 9), germanSoloCard(domain.CardDesignDiamond, 8))
	setGermanSoloHand(g, 2,
		germanSoloCard(domain.CardDesignSpade, 1), germanSoloCard(domain.CardDesignSpade, 13),
		germanSoloCard(domain.CardDesignSpade, 11), germanSoloCard(domain.CardDesignSpade, 10),
		germanSoloCard(domain.CardDesignSpade, 9), germanSoloCard(domain.CardDesignSpade, 8),
		germanSoloCard(domain.CardDesignSpade, 7), germanSoloCard(domain.CardDesignDiamond, 7))
	setGermanSoloHand(g, 3,
		germanSoloCard(domain.CardDesignClover, 1), germanSoloCard(domain.CardDesignClover, 13),
		germanSoloCard(domain.CardDesignClover, 11), germanSoloCard(domain.CardDesignClover, 10),
		germanSoloCard(domain.CardDesignClover, 9), germanSoloCard(domain.CardDesignClover, 8),
		germanSoloCard(domain.CardDesignClover, 7), germanSoloCard(domain.CardDesignDiamond, 12))

	g.SetDeclarerIdx(0)
	g.SetWinningBid(domain.GermanSoloBidTout)
	g.SetTrumpSuit(trump)
	g.SetPlaysAloneForTest(true)
	g.StartPlayForTest()

	for step := 0; step < 200 && g.GetPhase() != domain.GermanSoloPhaseRoundEnd; step++ {
		switch g.GetPhase() {
		case domain.GermanSoloPhasePlay:
			if g.IsHumanTurn() {
				valid := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
				require.NotEmpty(t, valid)
				require.NoError(t, g.PlayerPlay(valid[0]))
				continue
			}
			g.CpuPlay()
		case domain.GermanSoloPhaseTrickEnd:
			g.ResolveTrick()
			g.NextTrick()
		}
	}
	declarerTricks, defenderTricks := g.GetSideTrickCounts()
	assert.Equal(t, domain.GermanSoloTrickCount, declarerTricks, "8 トリック全部を取る")
	assert.Equal(t, 0, defenderTricks)
	assert.Equal(t, domain.GermanSoloOutcomeMade, g.GetOutcome())
}

// **同じトリックを二度精算しない。** ResolveTrick はトリック終了フェーズが続く間
// 何度でも呼べてしまうので、二度目は何もしないこと。
func TestGermanSolo_ResolveTrickIsIdempotent(t *testing.T) {
	g := newTestGermanSolo()
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetDeclarerIdx(0)
	g.SetTrickNumber(1)
	g.SetPhase(domain.GermanSoloPhaseTrickEnd)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: germanSoloCard(domain.CardDesignHeart, 13)},
		{PlayerIdx: 1, Card: germanSoloCard(domain.CardDesignHeart, 9)},
		{PlayerIdx: 2, Card: germanSoloCard(domain.CardDesignHeart, 8)},
		{PlayerIdx: 3, Card: germanSoloCard(domain.CardDesignDiamond, 8)},
	})
	g.ResolveTrick()
	g.ResolveTrick()
	total := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		total += g.GetPlayer(i).GetTrickCount()
	}
	assert.Equal(t, 1, total, "二度呼んでもトリックは 1 つ")
}

// **プレイ中のヒントは理由まで出す。** 出す札だけ返しても、なぜその札なのかが
// 読めないと助言として使えない。宣言側と守備側で理由が変わることも押さえる。
func TestGermanSolo_PlayHintExplainsTheChoice(t *testing.T) {
	newHintGame := func(declarer int) *domain.GermanSolo {
		g := newTestGermanSolo()
		g.SetTrumpSuit(domain.CardDesignHeart)
		g.SetDeclarerIdx(declarer)
		g.SetPlaysAloneForTest(true)
		g.SetWinningBid(domain.GermanSoloBidSolo)
		g.SetPhase(domain.GermanSoloPhasePlay)
		g.SetCurrentPlayerIdx(0)
		g.SetCurrentTrick(nil)
		setGermanSoloHand(g, 0,
			germanSoloCard(domain.CardDesignClover, 12), // Spadille
			germanSoloCard(domain.CardDesignDiamond, 8),
			germanSoloCard(domain.CardDesignSpade, 9))
		return g
	}

	// 宣言側のリードは強い札で主導する。
	declaring := newHintGame(0)
	hint := declaring.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "lead_high", hint.Reason)
	require.Len(t, hint.CardIndices, 1)

	// 守備側のリードは温存する。
	defending := newHintGame(1)
	hint = defending.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "lead_low", hint.Reason)

	// フォローできない手番は捨てる理由になる。
	discarding := newHintGame(0)
	discarding.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: germanSoloCard(domain.CardDesignHeart, 13)},
	})
	discarding.SetCurrentPlayerIdx(0)
	setGermanSoloHand(discarding, 0,
		germanSoloCard(domain.CardDesignDiamond, 8), germanSoloCard(domain.CardDesignSpade, 9))
	hint = discarding.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "discard_low", hint.Reason)

	// 勝てる札を持っていれば取りに行く。
	winning := newHintGame(0)
	winning.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: germanSoloCard(domain.CardDesignDiamond, 9)},
	})
	winning.SetCurrentPlayerIdx(0)
	setGermanSoloHand(winning, 0,
		germanSoloCard(domain.CardDesignDiamond, 1), germanSoloCard(domain.CardDesignDiamond, 8))
	hint = winning.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "follow_win", hint.Reason)

	// 味方が勝っていれば上書きしない。
	partnered := newHintGame(0)
	partnered.SetPlaysAloneForTest(false)
	partnered.SetCalledAceSuitForTest(domain.CardDesignClover)
	partnered.SetPartnerForTest(1, true)
	partnered.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: germanSoloCard(domain.CardDesignDiamond, 1)},
	})
	partnered.SetCurrentPlayerIdx(0)
	setGermanSoloHand(partnered, 0,
		germanSoloCard(domain.CardDesignDiamond, 9), germanSoloCard(domain.CardDesignDiamond, 8))
	hint = partnered.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "give_partner", hint.Reason)

	// 勝てず味方も勝っていなければダック。
	ducking := newHintGame(0)
	ducking.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: germanSoloCard(domain.CardDesignDiamond, 1)},
	})
	ducking.SetCurrentPlayerIdx(0)
	setGermanSoloHand(ducking, 0,
		germanSoloCard(domain.CardDesignDiamond, 9), germanSoloCard(domain.CardDesignDiamond, 8))
	hint = ducking.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "follow_duck", hint.Reason)
}

// **状態の読み出しは復元後も同じ値を返す。** 表示だけの getter でも、
// 落ちれば画面がゼロ値を出す。
func TestGermanSolo_StateAccessorsRoundTrip(t *testing.T) {
	g := newTestGermanSolo()
	g.SetRoundNumber(3)
	g.SetTrickNumber(4)
	g.SetPlayerScores([domain.GermanSoloPlayerCnt]int{5, -1, -2, -2})
	assert.Equal(t, 3, g.GetRoundNumber())
	assert.Equal(t, 4, g.GetTrickNumber())
	assert.Equal(t, [domain.GermanSoloPlayerCnt]int{5, -1, -2, -2}, g.GetPlayerScores())
	assert.Equal(t, domain.GermanSoloResultNone, g.GetResult())
}
