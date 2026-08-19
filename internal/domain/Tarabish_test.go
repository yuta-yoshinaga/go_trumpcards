//go:build test

package domain

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestTarabish(t *testing.T) *Tarabish {
	t.Helper()
	tb := NewDefaultTarabish()
	tb.Reset()
	return tb
}

// handOf は指定プレイヤーの手札を固定の並びに差し替える。
func handOf(tb *Tarabish, idx int, cards ...*Card) {
	p := tb.GetPlayer(idx)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// --- デッキと配り ---

// **36枚 (A,6-K)。** 4 人に 9 枚ずつでちょうど配り切り、切り札候補が 1 枚残る。
func TestTarabish_DeckIs36(t *testing.T) {
	tb := newTestTarabish(t)

	bySuit := map[int][]int{}
	total := 0
	for i := range TarabishPlayerCnt {
		p := tb.GetPlayer(i)
		for j := range p.GetCardsSize() {
			c := p.GetCard(j)
			bySuit[c.GetDesign()] = append(bySuit[c.GetDesign()], c.GetValue())
			total++
		}
	}
	if up := tb.GetUpCard(); up != nil {
		bySuit[up.GetDesign()] = append(bySuit[up.GetDesign()], up.GetValue())
		total++
	}

	// **配りは 2 段構え。** 切り札が決まる前は 6 枚 × 4 + 表向き 1 = 25 枚だけ見える。
	assert.Equal(t, TarabishFirstDealSize*TarabishPlayerCnt+1, total, "初回配布 + 表向きの 1 枚")
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		for _, v := range bySuit[suit] {
			assert.NotContains(t, []int{2, 3, 4, 5}, v, "2〜5 はデッキに無い (suit %d)", suit)
		}
	}
}

func TestTarabish_ResetDealsNineEach(t *testing.T) {
	tb := newTestTarabish(t)

	// **切り札が決まるまでは 6 枚。** 9 枚 × 4 = 36 はデッキ全部なので、
	// 先に配り切ると表向きにする 1 枚が残らない。
	for i := range TarabishPlayerCnt {
		assert.Equal(t, TarabishFirstDealSize, tb.GetPlayer(i).GetCardsSize(), "player %d", i)
	}
	require.NotNil(t, tb.GetUpCard(), "切り札候補が表向きになる")
	// **配り直後は切り札の選択フェーズ。**
	assert.Equal(t, TarabishPhaseBid, tb.GetPhase())
	assert.Equal(t, -1, tb.GetTrumpTakerIdx())
	assert.Equal(t, 1, tb.GetRoundNumber())
}

// **向かい合う席が味方。**
func TestTarabishTeamOf(t *testing.T) {
	assert.Equal(t, 0, TarabishTeamOf(0))
	assert.Equal(t, 1, TarabishTeamOf(1))
	assert.Equal(t, 0, TarabishTeamOf(2), "0 と 2 が味方")
	assert.Equal(t, 1, TarabishTeamOf(3), "1 と 3 が味方")
}

// --- 点数 ---

// **切り札だけ序列が変わる。** J=20 と 9=14 が A を追い越す。
func TestTarabishCardPoints_Trump(t *testing.T) {
	trump := CardDesignHeart
	for _, tc := range []struct {
		value int
		want  int
		name  string
	}{
		{11, 20, "J (Jass)"}, {9, 14, "9 (Menel)"}, {1, 11, "A"},
		{10, 10, "10"}, {13, 4, "K"}, {12, 3, "Q"},
		{8, 0, "8"}, {7, 0, "7"}, {6, 0, "6"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, TarabishCardPoints(NewCard(trump, tc.value, false), trump))
		})
	}
}

// 非切り札では J は 2 点、9 は 0 点まで落ちる。
func TestTarabishCardPoints_Plain(t *testing.T) {
	trump := CardDesignHeart
	for _, tc := range []struct {
		value int
		want  int
		name  string
	}{
		{1, 11, "A"}, {10, 10, "10"}, {13, 4, "K"}, {12, 3, "Q"},
		{11, 2, "J"}, {9, 0, "9"}, {6, 0, "6"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, TarabishCardPoints(NewCard(CardDesignSpade, tc.value, false), trump))
		})
	}
	assert.Equal(t, 0, TarabishCardPoints(nil, trump))
}

// **カード点の合計は 152。** 切り札 62 + 非切り札 30×3。
func TestTarabish_TotalCardPointsIs152(t *testing.T) {
	trump := CardDesignHeart
	total := 0
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		for _, v := range ShortDeckValues {
			total += TarabishCardPoints(NewCard(suit, v, false), trump)
		}
	}
	assert.Equal(t, 152, total, "切り札62 + 非切り札30×3")
	assert.Equal(t, 162, total+TarabishLastTrickBonus, "最終トリック込みで162")
}

// --- 序列 ---

// **切り札の J が最強、次が 9。** 平の A より強い。
func TestTarabish_TrumpOrder(t *testing.T) {
	trump, lead := CardDesignHeart, CardDesignSpade
	jass := NewCard(trump, 11, false)
	menel := NewCard(trump, 9, false)
	trumpAce := NewCard(trump, 1, false)

	assert.True(t, tarabishBeats(jass, menel, lead, trump), "J は 9 に勝つ")
	assert.True(t, tarabishBeats(menel, trumpAce, lead, trump), "9 は切り札の A に勝つ")
	assert.False(t, tarabishBeats(trumpAce, menel, lead, trump))
}

func TestTarabish_TrumpBeatsPlain(t *testing.T) {
	trump, lead := CardDesignHeart, CardDesignSpade
	assert.True(t, tarabishBeats(NewCard(trump, 6, false), NewCard(lead, 1, false), lead, trump),
		"切り札の 6 がリードの A に勝つ")
	assert.False(t, tarabishBeats(NewCard(lead, 1, false), NewCard(trump, 6, false), lead, trump))
}

func TestTarabish_PlainOrder(t *testing.T) {
	trump, lead := CardDesignHeart, CardDesignSpade
	assert.True(t, tarabishBeats(NewCard(lead, 1, false), NewCard(lead, 13, false), lead, trump), "A > K")
	assert.True(t, tarabishBeats(NewCard(lead, 13, false), NewCard(lead, 11, false), lead, trump), "K > J")
	// 非切り札では J は 10 より弱い扱いではなく、札の並び通り J > 10。
	assert.True(t, tarabishBeats(NewCard(lead, 11, false), NewCard(lead, 10, false), lead, trump))
	assert.False(t, tarabishBeats(NewCard(CardDesignClover, 1, false), NewCard(lead, 6, false), lead, trump),
		"無関係のスートは勝てない")
}

// --- メルド ---

// **ランの連続判定は札の並び (A,K,Q,J,10,9,8,7,6)。** 切り札の J/9 が強い序列は
// トリックの勝敗の話で、ランには効かない。混ぜると J と 9 が飛び地になる。
func TestTarabish_RunUsesSequenceOrderNotTrumpOrder(t *testing.T) {
	tb := newTestTarabish(t)
	// ♥ の 9,10,J は札の並びで連続する 3 枚。切り札序列だと J は最強で飛び地になる。
	handOf(tb, 0,
		NewCard(CardDesignHeart, 9, false),
		NewCard(CardDesignHeart, 10, false),
		NewCard(CardDesignHeart, 11, false),
		NewCard(CardDesignSpade, 6, false),
	)
	pts, length := tarabishRunPoints(tb.GetPlayer(0))
	assert.Equal(t, TarabishRun3Bonus, pts, "9-10-J は 3 枚のラン")
	assert.Equal(t, 3, length)
}

func TestTarabish_RunLengths(t *testing.T) {
	tb := newTestTarabish(t)

	t.Run("two in sequence is not a run", func(t *testing.T) {
		handOf(tb, 0, NewCard(CardDesignSpade, 6, false), NewCard(CardDesignSpade, 7, false))
		pts, length := tarabishRunPoints(tb.GetPlayer(0))
		assert.Equal(t, 0, pts)
		assert.Equal(t, 0, length)
	})

	t.Run("three scores 20", func(t *testing.T) {
		handOf(tb, 0, NewCard(CardDesignSpade, 6, false), NewCard(CardDesignSpade, 7, false),
			NewCard(CardDesignSpade, 8, false))
		pts, length := tarabishRunPoints(tb.GetPlayer(0))
		assert.Equal(t, TarabishRun3Bonus, pts)
		assert.Equal(t, 3, length)
	})

	t.Run("four scores 50", func(t *testing.T) {
		handOf(tb, 0, NewCard(CardDesignSpade, 6, false), NewCard(CardDesignSpade, 7, false),
			NewCard(CardDesignSpade, 8, false), NewCard(CardDesignSpade, 9, false))
		pts, length := tarabishRunPoints(tb.GetPlayer(0))
		assert.Equal(t, TarabishRun4Bonus, pts)
		assert.Equal(t, 4, length)
	})

	// **スートをまたいだ連続はランではない。** 負のコントロール。
	t.Run("across suits is not a run", func(t *testing.T) {
		handOf(tb, 0, NewCard(CardDesignSpade, 6, false), NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignClover, 8, false))
		pts, _ := tarabishRunPoints(tb.GetPlayer(0))
		assert.Equal(t, 0, pts)
	})

	// K-A は 6 を挟まず連続する（A が K の上）。
	t.Run("Q-K-A is a run", func(t *testing.T) {
		handOf(tb, 0, NewCard(CardDesignSpade, 12, false), NewCard(CardDesignSpade, 13, false),
			NewCard(CardDesignSpade, 1, false))
		pts, _ := tarabishRunPoints(tb.GetPlayer(0))
		assert.Equal(t, TarabishRun3Bonus, pts, "A は K の上に続く")
	})
}

// **ベラは切り札の K+Q だけ。** 他スートの K+Q では付かない。
func TestTarabish_Bella(t *testing.T) {
	tb := newTestTarabish(t)
	trump := CardDesignHeart

	handOf(tb, 0, NewCard(trump, 13, false), NewCard(trump, 12, false))
	assert.True(t, tarabishHasBella(tb.GetPlayer(0), trump))

	handOf(tb, 0, NewCard(CardDesignSpade, 13, false), NewCard(CardDesignSpade, 12, false))
	assert.False(t, tarabishHasBella(tb.GetPlayer(0), trump), "非切り札の K+Q は無効")

	handOf(tb, 0, NewCard(trump, 13, false))
	assert.False(t, tarabishHasBella(tb.GetPlayer(0), trump), "K だけでは無効")
}

// メルド点は配り直後に確定し、ランとベラが合算される。
func TestTarabish_CountMeldsCombinesRunAndBella(t *testing.T) {
	tb := newTestTarabish(t)
	trump := CardDesignHeart
	tb.SetTrumpSuitForTest(trump)
	// ♥ Q-K-A のランかつ ♥ K+Q のベラ。
	handOf(tb, 0,
		NewCard(trump, 12, false), NewCard(trump, 13, false), NewCard(trump, 1, false))
	for i := 1; i < TarabishPlayerCnt; i++ {
		handOf(tb, i, NewCard(CardDesignSpade, 6, false))
	}

	tb.CountMeldsForTest()

	assert.Equal(t, TarabishRun3Bonus+TarabishBellaBonus, tb.GetPlayer(0).GetMeldPoints(),
		"ラン 20 + ベラ 20")
	assert.True(t, tb.GetPlayer(0).GetHasBella())
	assert.Equal(t, 0, tb.GetPlayer(1).GetMeldPoints())
}

// --- 切り札の選択 ---

func TestTarabish_TakeTrumpStartsPlay(t *testing.T) {
	tb := newTestTarabish(t)
	tb.SetCurrentPlayerIdxForTest(0)
	require.NoError(t, tb.TakeTrump())

	assert.Equal(t, 0, tb.GetTrumpTakerIdx())
	assert.Equal(t, tb.GetUpCard().GetDesign(), tb.GetTrumpSuit())
	assert.Equal(t, TarabishPhasePlay, tb.GetPhase())
	// **切り札が決まると残りが配られ、全員 9 枚になる。**
	for i := range TarabishPlayerCnt {
		assert.Equal(t, TarabishHandSize, tb.GetPlayer(i).GetCardsSize(), "player %d", i)
	}
}

// **表向きの 1 枚は親の手札に入る。** 24 + 1 + 2 + 9 = 36 でちょうど配り切る。
func TestTarabish_UpCardGoesToTheDealer(t *testing.T) {
	tb := newTestTarabish(t)
	up := tb.GetUpCard()
	require.NotNil(t, up)
	tb.SetCurrentPlayerIdxForTest(0)
	require.NoError(t, tb.TakeTrump())

	dealer := tb.GetPlayer(tb.GetDealerIdx())
	found := false
	for i := range dealer.GetCardsSize() {
		c := dealer.GetCard(i)
		if c.GetDesign() == up.GetDesign() && c.GetValue() == up.GetValue() {
			found = true
			break
		}
	}
	assert.True(t, found, "表向きの札が親の手札にある")

	// 36 枚がちょうど配り切られている。
	total := 0
	for i := range TarabishPlayerCnt {
		total += tb.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, TarabishHandSize*TarabishPlayerCnt, total, "36枚を配り切る")
}

func TestTarabish_PassMovesToTheNextSeat(t *testing.T) {
	tb := newTestTarabish(t)
	// 親でなければ見送れる。親を 3 番にして、人間 (0番) を最初の入札者にする。
	tb.SetDealerIdxForTest(3)
	tb.SetCurrentPlayerIdxForTest(0)
	require.NoError(t, tb.PassTrump())

	assert.Equal(t, 1, tb.GetCurrentPlayerIdx())
	assert.Equal(t, TarabishPhaseBid, tb.GetPhase(), "まだ選択フェーズ")
	assert.Equal(t, -1, tb.GetTrumpTakerIdx())
}

// **全員が見送っても切り札は必ず決まる。** そうしないと誰も切り札を決めない
// まま手番だけが回り続ける。
//
// 親が CPU のときは回ってきた時点で自動的に引き受ける。親が人間のときは
// 人間の番で止まり、見送りが拒否されるので「引き受ける」しか押せない。
func TestTarabish_TrumpAlwaysGetsSettled(t *testing.T) {
	for _, dealer := range []int{0, 1, 2, 3} {
		t.Run("dealer "+string(rune('0'+dealer)), func(t *testing.T) {
			tb := newTestTarabish(t)
			tb.SetDealerIdxForTest(dealer)
			tb.SetCurrentPlayerIdxForTest((dealer + 1) % TarabishPlayerCnt)

			guard := 0
			for tb.GetPhase() == TarabishPhaseBid && guard < 10 {
				guard++
				if tb.IsHumanBidTurn() {
					// 人間が親なら見送れないので引き受ける。
					if tb.GetDealerIdx() == 0 {
						require.NoError(t, tb.TakeTrump())
					} else {
						require.NoError(t, tb.PassTrump())
					}
					continue
				}
				tb.CpuBid()
			}

			assert.Equal(t, TarabishPhasePlay, tb.GetPhase(), "必ず切り札が決まる")
			assert.GreaterOrEqual(t, tb.GetTrumpTakerIdx(), 0)
		})
	}
}

// **親は見送れない。** 人間が親のときに見送りを許すと、切り札が決まらない。
func TestTarabish_DealerCannotPass(t *testing.T) {
	tb := newTestTarabish(t)
	require.Equal(t, 0, tb.GetDealerIdx(), "初回の親は 0 番")
	tb.SetCurrentPlayerIdxForTest(0)

	assert.Error(t, tb.PassTrump(), "親は見送れない")
	assert.NoError(t, tb.TakeTrump(), "引き受けは通る")
}

// **人間が親でも、engine が代わりに引き受けてしまわない。** 選択の余地が
// 無くても、引き受けたのが自分だと分かるように操作させる。
func TestTarabish_HumanDealerIsAskedNotAutoAccepted(t *testing.T) {
	tb := newTestTarabish(t)
	require.Equal(t, 0, tb.GetDealerIdx())

	// CPU 1,2,3 が全員見送るところまで回す。
	guard := 0
	for tb.GetPhase() == TarabishPhaseBid && !tb.IsHumanBidTurn() && guard < 10 {
		guard++
		tb.CpuBid()
	}

	if tb.GetPhase() == TarabishPhaseBid {
		// 人間の番で止まっている。まだ誰も引き受けていない。
		assert.True(t, tb.IsHumanBidTurn(), "人間の番で止まる")
		assert.Equal(t, -1, tb.GetTrumpTakerIdx(), "engine が勝手に引き受けない")
		assert.Error(t, tb.PassTrump(), "親なので見送れない")
	}
}

// CPU が親のときは、回ってきた時点で自動的に引き受ける（進まなくなるのを防ぐ）。
func TestTarabish_CpuDealerIsStuckAutomatically(t *testing.T) {
	tb := newTestTarabish(t)
	tb.SetDealerIdxForTest(2)
	tb.SetCurrentPlayerIdxForTest(1) // 親の 1 つ手前

	require.NoError(t, tb.PassTrumpForTest(1))

	assert.Equal(t, TarabishPhasePlay, tb.GetPhase(), "CPU の親が引き受けて進む")
	assert.Equal(t, 2, tb.GetTrumpTakerIdx())
}

func TestTarabish_BidRejections(t *testing.T) {
	tb := newTestTarabish(t)
	tb.SetCurrentPlayerIdxForTest(1)
	assert.Error(t, tb.TakeTrump(), "自分の番でなければ選べない")
	assert.Error(t, tb.PassTrump())

	tb.SetCurrentPlayerIdxForTest(0)
	require.NoError(t, tb.TakeTrump())
	assert.Error(t, tb.TakeTrump(), "プレイに入ったら選べない")
}

// --- プレイと集計 ---

func TestTarabish_TrickPointsGoToTheWinnerTeam(t *testing.T) {
	tb := newTestTarabish(t)
	trump := CardDesignHeart
	tb.SetTrumpSuitForTest(trump)
	tb.SetPhaseForTest(TarabishPhasePlay)
	tb.SetTrickNumberForTest(2)
	tb.leadPlayerIdx = 0
	tb.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 1, false)},  // 11
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 10, false)}, // 10
		{PlayerIdx: 2, Card: NewCard(trump, 11, false)},           // Jass 20、これが勝つ
		{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 13, false)}, // 4
	})
	tb.ResolveTrickForTest()

	// 2 番はチーム 0。11+10+20+4 = 45。
	assert.Equal(t, 45, tb.GetRoundPoints(0))
	assert.Equal(t, 0, tb.GetRoundPoints(1))
}

// **最終トリックに 10 点。**
func TestTarabish_LastTrickBonus(t *testing.T) {
	tb := newTestTarabish(t)
	tb.SetTrumpSuitForTest(CardDesignHeart)
	tb.SetPhaseForTest(TarabishPhasePlay)
	tb.SetTrickNumberForTest(TarabishTricksPerRound - 1)
	tb.leadPlayerIdx = 1
	tb.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 6, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignSpade, 7, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 8, false)},
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 9, false)}, // 0 が取る
	})
	tb.ResolveTrickForTest()

	// 札はすべて 0 点なので、加算されるのは最終トリックの 10 点だけ。
	assert.Equal(t, TarabishLastTrickBonus, tb.GetRoundPoints(0))
}

// メルド点はラウンド終了時にチーム得点へ乗る。
func TestTarabish_MeldPointsAddedAtRoundEnd(t *testing.T) {
	tb := newTestTarabish(t)
	tb.config.Target = 10000 // まだ終わらせない
	tb.GetPlayer(0).SetMeldPoints(50)
	tb.GetPlayer(1).SetMeldPoints(20)
	tb.SetRoundPointsForTest(0, 30)
	tb.SetRoundPointsForTest(1, 40)

	tb.FinishRoundForTest()

	assert.Equal(t, 80, tb.GetScore(0), "カード30 + メルド50")
	assert.Equal(t, 60, tb.GetScore(1), "カード40 + メルド20")
	assert.Equal(t, TarabishPhaseRoundEnd, tb.GetPhase())
}

func TestTarabish_ReachingTargetEndsTheGame(t *testing.T) {
	tb := newTestTarabish(t)
	tb.config.Target = 100
	tb.SetScoreForTestUse(0, 90)
	tb.SetRoundPointsForTest(0, 20)

	tb.FinishRoundForTest()

	assert.True(t, tb.GetGameEndFlag())
	assert.Equal(t, TarabishPhaseGameEnd, tb.GetPhase())
	assert.Equal(t, 0, tb.GetWinnerTeam())
}

// 両チームが同時に目標を超えたら高いほうが勝つ。
func TestTarabish_HigherScoreWinsWhenBothReachTarget(t *testing.T) {
	tb := newTestTarabish(t)
	tb.config.Target = 100
	tb.SetScoreForTestUse(0, 95)
	tb.SetScoreForTestUse(1, 95)
	tb.SetRoundPointsForTest(0, 10)
	tb.SetRoundPointsForTest(1, 30)

	tb.FinishRoundForTest()

	assert.Equal(t, 1, tb.GetWinnerTeam())
}

func TestTarabish_NextRoundRedealsAndKeepsScores(t *testing.T) {
	tb := newTestTarabish(t)
	tb.SetScoreForTestUse(0, 120)
	tb.SetPhaseForTest(TarabishPhaseRoundEnd)
	dealer := tb.GetDealerIdx()

	tb.NextRound()

	assert.Equal(t, 2, tb.GetRoundNumber())
	assert.Equal(t, TarabishPhaseBid, tb.GetPhase(), "次も切り札の選択から")
	assert.Equal(t, (dealer+1)%TarabishPlayerCnt, tb.GetDealerIdx())
	assert.Equal(t, 120, tb.GetScore(0), "累計得点は持ち越す")
	assert.Equal(t, -1, tb.GetTrumpTakerIdx(), "切り札はラウンドごとに決め直す")
	for i := range TarabishPlayerCnt {
		assert.Equal(t, TarabishFirstDealSize, tb.GetPlayer(i).GetCardsSize(), "切り札が決まるまでは 6 枚")
	}
}

func TestTarabish_MustFollowSuit(t *testing.T) {
	tb := newTestTarabish(t)
	handOf(tb, 1, NewCard(CardDesignSpade, 8, false), NewCard(CardDesignHeart, 9, false))
	tb.SetCurrentTrickForTest([]*TrickCard{{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 13, false)}})

	assert.Equal(t, []int{0}, tb.GetValidPlayIndices(1))
}

func TestTarabish_GetValidPlayIndicesOutOfRange(t *testing.T) {
	tb := newTestTarabish(t)
	assert.Nil(t, tb.GetValidPlayIndices(-1))
	assert.Nil(t, tb.GetValidPlayIndices(TarabishPlayerCnt))
}

func TestTarabish_PlayerPlayRejections(t *testing.T) {
	t.Run("bid phase", func(t *testing.T) {
		tb := newTestTarabish(t)
		assert.Error(t, tb.PlayerPlay(0), "切り札が決まる前は出せない")
	})
	t.Run("not your turn", func(t *testing.T) {
		tb := newTestTarabish(t)
		tb.SetPhaseForTest(TarabishPhasePlay)
		tb.SetCurrentPlayerIdxForTest(2)
		assert.Error(t, tb.PlayerPlay(0))
	})
	t.Run("game over", func(t *testing.T) {
		tb := newTestTarabish(t)
		tb.GiveUp()
		assert.Error(t, tb.PlayerPlay(0))
	})
	t.Run("index out of range", func(t *testing.T) {
		tb := newTestTarabish(t)
		tb.SetPhaseForTest(TarabishPhasePlay)
		tb.SetCurrentPlayerIdxForTest(0)
		assert.Error(t, tb.PlayerPlay(99))
		assert.Error(t, tb.PlayerPlay(-1))
	})
}

// CPU は合法手しか出さず、ゲームは必ず終わる。
func TestTarabish_CpuAlwaysPlaysLegally(t *testing.T) {
	for range 30 {
		tb := NewDefaultTarabish()
		tb.SetConfig(TarabishConfig{Target: 200})
		tb.Reset()
		guard := 0
		for !tb.GetGameEndFlag() && guard < 3000 {
			guard++
			switch {
			case tb.IsHumanBidTurn():
				require.NoError(t, tb.TakeTrump())
			case tb.GetPhase() == TarabishPhaseBid:
				tb.CpuBid()
			case tb.GetPhase() == TarabishPhaseRoundEnd:
				tb.NextRound()
			case tb.IsHumanTurn():
				valid := tb.GetValidPlayIndices(0)
				require.NotEmpty(t, valid)
				require.NoError(t, tb.PlayerPlay(valid[0]))
			default:
				idx := tb.GetCurrentPlayerIdx()
				before := tb.GetPlayer(idx).GetCardsSize()
				tb.CpuPlay()
				require.Equal(t, before-1, tb.GetPlayer(idx).GetCardsSize())
			}
		}
		require.True(t, tb.GetGameEndFlag())
	}
}

// **1 ラウンドのカード点は必ず 162。** 152 + 最終トリック 10。
func TestTarabish_RoundCardPointsAlwaysTotal162(t *testing.T) {
	for range 30 {
		tb := NewDefaultTarabish()
		tb.SetConfig(TarabishConfig{Target: 10000}) // 1 ラウンドで終わらせない
		tb.Reset()
		guard := 0
		for tb.GetPhase() != TarabishPhaseRoundEnd && guard < 200 {
			guard++
			switch {
			case tb.IsHumanBidTurn():
				require.NoError(t, tb.TakeTrump())
			case tb.GetPhase() == TarabishPhaseBid:
				tb.CpuBid()
			case tb.IsHumanTurn():
				valid := tb.GetValidPlayIndices(0)
				require.NoError(t, tb.PlayerPlay(valid[0]))
			default:
				tb.CpuPlay()
			}
		}
		require.Equal(t, TarabishPhaseRoundEnd, tb.GetPhase())

		// メルド点を除いたカード点だけの合計を見る。
		melds := 0
		for i := range TarabishPlayerCnt {
			melds += tb.GetPlayer(i).GetMeldPoints()
		}
		require.Equal(t, 162+melds, tb.GetScore(0)+tb.GetScore(1), "カード152 + 最終10 + メルド")
	}
}

func TestTarabish_GiveUp(t *testing.T) {
	tb := newTestTarabish(t)
	tb.GiveUp()
	assert.True(t, tb.GetGameEndFlag())
	assert.Equal(t, TarabishPhaseGameEnd, tb.GetPhase())
	assert.Equal(t, 1, tb.GetWinnerTeam(), "投了したので相手チームの勝ち")

	tb.GiveUp()
	assert.True(t, tb.GetGameEndFlag())
}

func TestTarabish_GetPlayerAndScoreOutOfRange(t *testing.T) {
	tb := newTestTarabish(t)
	assert.Nil(t, tb.GetPlayer(-1))
	assert.Nil(t, tb.GetPlayer(TarabishPlayerCnt))
	assert.Equal(t, 0, tb.GetScore(-1))
	assert.Equal(t, 0, tb.GetScore(TarabishTeamCnt))
	assert.Equal(t, 0, tb.GetRoundPoints(-1))
}

func TestTarabish_Config(t *testing.T) {
	tb := newTestTarabish(t)
	assert.Equal(t, TarabishTargetDefault, tb.GetConfig().Target)

	tb.SetConfig(TarabishConfig{Target: 300})
	assert.Equal(t, 300, tb.GetConfig().Target)

	assert.NoError(t, TarabishConfig{Target: TarabishTargetMin}.Validate())
	assert.NoError(t, TarabishConfig{Target: TarabishTargetMax}.Validate())
	assert.Error(t, TarabishConfig{Target: 50}.Validate())
	assert.Error(t, TarabishConfig{Target: 2000}.Validate())
}

// --- ヒント ---

// 選択フェーズでは札ではなく引き受けるかどうかを助言する。
func TestTarabish_GetHint_BidPhase(t *testing.T) {
	tb := newTestTarabish(t)
	tb.SetCurrentPlayerIdxForTest(0)

	h := tb.GetHint()
	require.NotNil(t, h)
	assert.Nil(t, h.CardIndex, "選択フェーズでは札を指さない")
	assert.Contains(t, []string{"tarabishTakeTrump", "tarabishPassTrump"}, h.Reason)
}

// 強い切り札候補なら引き受けを、弱ければ見送りを勧める。両側を踏む。
func TestTarabish_GetHint_BidBothWays(t *testing.T) {
	strong := newTestTarabish(t)
	strong.SetCurrentPlayerIdxForTest(0)
	suit := strong.GetUpCard().GetDesign()
	handOf(strong, 0,
		NewCard(suit, 11, false), NewCard(suit, 9, false),
		NewCard(suit, 1, false), NewCard(suit, 10, false))
	assert.Equal(t, "tarabishTakeTrump", strong.GetHint().Reason)

	weak := newTestTarabish(t)
	weak.SetCurrentPlayerIdxForTest(0)
	other := CardDesignSpade
	if weak.GetUpCard().GetDesign() == CardDesignSpade {
		other = CardDesignHeart
	}
	handOf(weak, 0, NewCard(other, 6, false), NewCard(other, 7, false))
	assert.Equal(t, "tarabishPassTrump", weak.GetHint().Reason)
}

func TestTarabish_GetHint_PlayPhase(t *testing.T) {
	tb := newTestTarabish(t)
	tb.SetCurrentPlayerIdxForTest(0)
	require.NoError(t, tb.TakeTrump())
	tb.SetCurrentPlayerIdxForTest(0) // リードは親の左隣なので手番を戻す

	h := tb.GetHint()
	if assert.NotNil(t, h) && assert.NotNil(t, h.CardIndex) {
		assert.Contains(t, tb.GetValidPlayIndices(0), *h.CardIndex)
		assert.Equal(t, "tarabishWinTrick", h.Reason)
	}
}

// **味方が勝っていれば狙いが変わる。** 取りに行かず点を乗せる。
func TestTarabish_GetHint_FeedPartner(t *testing.T) {
	tb := newTestTarabish(t)
	tb.SetCurrentPlayerIdxForTest(0)
	require.NoError(t, tb.TakeTrump())
	tb.SetCurrentPlayerIdxForTest(0)
	// 2 番（味方）が強い札で勝っている状況。
	tb.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 6, false)},
		{PlayerIdx: 2, Card: NewCard(tb.GetTrumpSuit(), 11, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 7, false)},
	})

	h := tb.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "tarabishFeedPartner", h.Reason)
}

func TestTarabish_GetHint_NilAfterGameEnd(t *testing.T) {
	tb := newTestTarabish(t)
	tb.GiveUp()
	assert.Nil(t, tb.GetHint())
}

// --- JSON 往復 ---

func TestTarabish_JSONRoundTrip(t *testing.T) {
	tb := newTestTarabish(t)
	tb.SetCurrentPlayerIdxForTest(0)
	require.NoError(t, tb.TakeTrump())
	tb.SetScoreForTestUse(0, 120)
	tb.SetScoreForTestUse(1, 80)
	tb.SetRoundPointsForTest(0, 30)
	tb.GetPlayer(0).SetMeldPoints(50)
	tb.GetPlayer(0).SetHasBella(true)

	data, err := json.Marshal(tb)
	require.NoError(t, err)

	restored := NewDefaultTarabish()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, 120, restored.GetScore(0))
	assert.Equal(t, 80, restored.GetScore(1))
	assert.Equal(t, 30, restored.GetRoundPoints(0))
	assert.Equal(t, 50, restored.GetPlayer(0).GetMeldPoints(), "メルド点が往復する")
	assert.True(t, restored.GetPlayer(0).GetHasBella())
	assert.Equal(t, tb.GetTrumpSuit(), restored.GetTrumpSuit(), "切り札が往復する")
	assert.Equal(t, tb.GetTrumpTakerIdx(), restored.GetTrumpTakerIdx())
	assert.Equal(t, tb.GetConfig().Target, restored.GetConfig().Target)
}

func TestTarabish_UnmarshalRejectsGarbage(t *testing.T) {
	assert.Error(t, json.Unmarshal([]byte("not json"), NewDefaultTarabish()))
}

func TestTarabish_ActionLog(t *testing.T) {
	tb := newTestTarabish(t)
	assert.NotEmpty(t, tb.GetActionLog())
}

// **切り札の 10 は K/Q に勝つ。** 素の GetValue() に落とすと 10 < 12 < 13 で
// 序列が点数表と食い違う（レビュー指摘 PR #5301）。
func TestTarabish_TrumpTenBeatsKingAndQueen(t *testing.T) {
	trump := CardDesignHeart
	ten := NewCard(trump, 10, false)
	king := NewCard(trump, 13, false)
	queen := NewCard(trump, 12, false)

	assert.True(t, tarabishBeats(ten, king, trump, trump))
	assert.True(t, tarabishBeats(ten, queen, trump, trump))
	assert.False(t, tarabishBeats(king, ten, trump, trump))
	assert.False(t, tarabishBeats(queen, ten, trump, trump))
}

// 切り札の序列を端から端まで踏む。J>9>A>10>K>Q>8>7>6。
func TestTarabish_FullTrumpOrder(t *testing.T) {
	trump := CardDesignHeart
	order := []int{11, 9, 1, 10, 13, 12, 8, 7, 6}
	for i := 0; i < len(order)-1; i++ {
		hi := NewCard(trump, order[i], false)
		lo := NewCard(trump, order[i+1], false)
		assert.True(t, tarabishBeats(hi, lo, trump, trump), "%d must beat %d", order[i], order[i+1])
		assert.False(t, tarabishBeats(lo, hi, trump, trump), "%d must not beat %d", order[i+1], order[i])
	}
}

// 壊れた設定のスナップショットは復元しない。
func TestTarabish_UnmarshalRejectsInvalidConfig(t *testing.T) {
	tb := newTestTarabish(t)
	data, err := json.Marshal(tb)
	require.NoError(t, err)

	var raw map[string]any
	require.NoError(t, json.Unmarshal(data, &raw))
	raw["cf"] = map[string]any{"tg": 0}
	broken, err := json.Marshal(raw)
	require.NoError(t, err)

	var got Tarabish
	assert.Error(t, json.Unmarshal(broken, &got))

	// 正のコントロール: 触っていないスナップショットは通る。
	var ok Tarabish
	assert.NoError(t, json.Unmarshal(data, &ok))
}

// **同じ 0 点なら切り札ではない方をリードする。** 点だけで選ぶと 0 点の切り札
// 8/7/6 が先に出て「切り札は温存する」というコメントと裏返る（レビュー指摘 PR #5301）。
func TestTarabish_CpuLeadPrefersNonTrumpOnTie(t *testing.T) {
	tb := newTestTarabish(t)
	tb.trumpSuit = CardDesignHeart
	tb.phase = TarabishPhasePlay
	tb.currentTrick = nil
	handOf(tb, 1,
		NewCard(CardDesignHeart, 7, false), // 0 点だが切り札
		NewCard(CardDesignSpade, 8, false)) // 0 点の非切り札

	assert.Equal(t, 1, tb.chooseCpuCard(1))
}

// **切り札だけ点数表が入れ替わる** (#5749)。手札に出す点と精算の点が
// 同じ関数から来ることを、TS と共有する黄金ベクタで固定する。
func TestTarabishCardPoints_GoldenVectors(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(
		"..", "..", "frontend", "src", "utils", "__fixtures__", "tarabishPoints.golden.json"))
	if err != nil {
		t.Fatalf("read the golden vectors: %v", err)
	}
	var golden struct {
		Cases []struct {
			Name      string `json:"name"`
			Design    string `json:"design"`
			Value     int    `json:"value"`
			TrumpSuit int    `json:"trumpSuit"`
			Points    int    `json:"points"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &golden); err != nil {
		t.Fatalf("parse the golden vectors: %v", err)
	}
	if len(golden.Cases) == 0 {
		t.Fatal("no vectors to check")
	}
	designs := map[string]int{
		"SPADE": CardDesignSpade, "CLOVER": CardDesignClover,
		"HEART": CardDesignHeart, "DIAMOND": CardDesignDiamond,
	}
	for _, c := range golden.Cases {
		design, ok := designs[c.Design]
		if !ok {
			t.Fatalf("%s: unknown design %q", c.Name, c.Design)
		}
		if got := TarabishCardPoints(NewCard(design, c.Value, true), c.TrumpSuit); got != c.Points {
			t.Errorf("%s: got %d, want %d", c.Name, got, c.Points)
		}
	}
}
