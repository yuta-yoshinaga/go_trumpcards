//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newQuodlibetForTest(t *testing.T) *Quodlibet {
	t.Helper()
	q := NewDefaultQuodlibet()
	q.Reset()
	return q
}

// quodlibetPlayDeal は 1 ディールを最後まで打つ。合法手の先頭を出し続ける。
func quodlibetPlayDeal(t *testing.T, q *Quodlibet) {
	t.Helper()
	for step := 0; step < 2000; step++ {
		switch q.GetPhase() {
		case QuodlibetPhaseSelectContract:
			if q.IsHumanTurn() {
				avail := q.GetAvailableContracts()
				require.NotEmpty(t, avail)
				require.NoError(t, q.SelectContract(avail[0]))
				continue
			}
			q.CpuSelectContract()
		case QuodlibetPhasePlay:
			if q.IsHumanTurn() {
				idx := -1
				if valid := q.GetPlayableIndices(q.GetCurrentTurn()); len(valid) > 0 {
					idx = valid[0]
				}
				require.NoError(t, q.PlayerPlay(idx))
				continue
			}
			q.CpuPlay()
		default:
			return
		}
	}
	t.Fatal("ディールが終わらない")
}

// quodlibetPlayMatch は 12 ディールを終局まで打ち切る。
func quodlibetPlayMatch(t *testing.T, q *Quodlibet) {
	t.Helper()
	for deal := 0; deal < QuodlibetTotalDeals+1 && !q.GetGameEndFlag(); deal++ {
		quodlibetPlayDeal(t, q)
		require.Equal(t, QuodlibetPhaseDealEnd, q.GetPhase())
		q.NextDeal()
	}
	require.True(t, q.GetGameEndFlag(), "12 ディールで終局しない")
}

// **32 枚を 4 人に 8 枚ずつ。** 余りもタロンも無い。
func TestQuodlibet_DealsEightEach(t *testing.T) {
	q := newQuodlibetForTest(t)
	total := 0
	for i, p := range q.GetPlayers() {
		assert.Equal(t, QuodlibetHandSize, p.GetCardsSize(), "席 %d の手札", i)
		total += p.GetCardsSize()
	}
	assert.Equal(t, 32, total)
	assert.Equal(t, QuodlibetPhaseSelectContract, q.GetPhase())
}

// **12 ディールは 3 つの輪 × 4 種目。** ディーラー 1 人につき 1 種目ではない。
func TestQuodlibet_TwelveDealsInThreeRounds(t *testing.T) {
	assert.Equal(t, 12, QuodlibetTotalDeals)
	assert.Equal(t, 3, QuodlibetRoundCnt)
	for c := 0; c < QuodlibetContractCnt; c++ {
		assert.Equal(t, c/4, QuodlibetRoundOf(c), "コントラクト %d の輪", c)
	}
	assert.Equal(t, -1, QuodlibetRoundOf(-1))
	assert.Equal(t, -1, QuodlibetRoundOf(QuodlibetContractCnt))
}

// **選べるのはその輪の 4 種目だけ。** 輪をまたいで選べると構成が崩れる。
func TestQuodlibet_ContractChoicesStayInsideTheRound(t *testing.T) {
	q := newQuodlibetForTest(t)
	for _, c := range q.GetAvailableContracts() {
		assert.Equal(t, 0, QuodlibetRoundOf(c), "第 1 の輪の外が選べる: %d", c)
	}
	// 第 2 の輪へ進めると、選べる種目もそちらへ移る。
	err := q.SelectContract(QuodlibetOberUnter)
	assert.Error(t, err, "別の輪の種目が選べてしまう")
}

func TestQuodlibet_UsedContractsShrinkTheChoices(t *testing.T) {
	q := newQuodlibetForTest(t)
	require.Len(t, q.GetAvailableContracts(), QuodlibetContractsPerRound)
	require.NoError(t, q.SelectContract(QuodlibetMinus))
	quodlibetPlayDeal(t, q)
	q.NextDeal()
	avail := q.GetAvailableContracts()
	assert.Len(t, avail, QuodlibetContractsPerRound-1)
	assert.NotContains(t, avail, QuodlibetMinus)
}

// **切り札は無い。** 台札スートの最強札だけが取る。
func TestQuodlibet_NoTrumpTheLedSuitDecides(t *testing.T) {
	q := newQuodlibetForTest(t)
	q.currentContract = QuodlibetMinus
	q.phase = QuodlibetPhasePlay
	q.leadPlayer = 0
	q.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 9, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 13, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 1, false)}, // 別スートのエースは無力
		{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 10, false)},
	}
	assert.Equal(t, 1, q.trickWinner())
}

// エースが最強 (値は 1 だが強さは 14)。
func TestQuodlibetCardStrength(t *testing.T) {
	assert.Equal(t, 14, QuodlibetCardStrength(NewCard(CardDesignSpade, 1, false)))
	assert.Greater(t, QuodlibetCardStrength(NewCard(CardDesignSpade, 1, false)),
		QuodlibetCardStrength(NewCard(CardDesignSpade, 13, false)))
	assert.Greater(t, QuodlibetCardStrength(NewCard(CardDesignSpade, 13, false)),
		QuodlibetCardStrength(NewCard(CardDesignSpade, 7, false)))
	assert.Equal(t, -1, QuodlibetCardStrength(nil))
}

// **右隣は次の手番の席。** 反時計回りなので、向きを取り違えると悪い隣人の
// 罰点が真逆の人に付く。
func TestQuodlibetRightOf(t *testing.T) {
	assert.Equal(t, 1, QuodlibetRightOf(0))
	assert.Equal(t, 0, QuodlibetRightOf(3))
}

func TestQuodlibet_ScorePlus(t *testing.T) {
	q := newQuodlibetForTest(t)
	q.currentContract = QuodlibetPlus
	quodlibetGiveTricks(q, []int{4, 3, 1, 0})
	points := quodlibetScorePlus(q)
	assert.Equal(t, 40, points[0], "8-4 トリック")
	assert.Equal(t, 50, points[1])
	assert.Equal(t, 70, points[2])
	// **1 つも取れなければ 80 ではなく 100。**
	assert.Equal(t, 100, points[3])
}

func TestQuodlibet_ScoreMinus(t *testing.T) {
	q := newQuodlibetForTest(t)
	q.currentContract = QuodlibetMinus
	quodlibetGiveTricks(q, []int{4, 3, 1, 0})
	points := quodlibetScoreMinus(q)
	assert.Equal(t, 40, points[0])
	assert.Equal(t, 0, points[3])

	quodlibetGiveTricks(q, []int{8, 0, 0, 0})
	// **全部取ると 80 ではなく 100。**
	assert.Equal(t, 100, quodlibetScoreMinus(q)[0])
}

// **悪い隣人の点は右隣に流れる。** 自分の取ったトリックで隣が沈む。
func TestQuodlibet_ScoreBadNeighbourPaysTheRightHandSeat(t *testing.T) {
	q := newQuodlibetForTest(t)
	q.currentContract = QuodlibetBadNeighbour
	quodlibetGiveTricks(q, []int{3, 0, 0, 5})
	points := quodlibetScoreBadNeighbour(q)
	assert.Equal(t, 30, points[1], "席 0 の 3 トリックは右隣の席 1 へ")
	assert.Equal(t, 50, points[0], "席 3 の 5 トリックは右隣の席 0 へ")
	assert.Equal(t, 0, points[2], "誰の左隣でもない席には付かない")
	assert.Equal(t, 0, points[3], "自分の取ったトリックでは沈まない")
	// マイナスなら同じ配りで席 0 が 30、席 3 が 50 を負う ── つまり左右が入れ替わる。
	minus := quodlibetScoreMinus(q)
	assert.Equal(t, 30, minus[0])
	assert.Equal(t, 50, minus[3])
}

// **K♥ と Q♦ を同じトリックで取ると 80 ではなく 100。**
func TestQuodlibet_ScoreAlarich(t *testing.T) {
	q := newQuodlibetForTest(t)
	q.currentContract = QuodlibetAlarich
	king := NewCard(CardDesignHeart, 13, false)
	ruffian := NewCard(CardDesignDiamond, 12, false)
	plain := NewCard(CardDesignSpade, 7, false)

	quodlibetResetTricks(q)
	q.GetPlayer(0).AddTrick([]*Card{king, plain, plain, plain})
	q.GetPlayer(1).AddTrick([]*Card{ruffian, plain, plain, plain})
	points := quodlibetScoreAlarich(q)
	assert.Equal(t, 50, points[0])
	assert.Equal(t, 30, points[1])

	quodlibetResetTricks(q)
	q.GetPlayer(2).AddTrick([]*Card{king, ruffian, plain, plain})
	assert.Equal(t, 100, quodlibetScoreAlarich(q)[2], "同じトリックなら 80 ではなく 100")

	// 別々のトリックなら合計 80。
	quodlibetResetTricks(q)
	q.GetPlayer(3).AddTrick([]*Card{king, plain, plain, plain})
	q.GetPlayer(3).AddTrick([]*Card{ruffian, plain, plain, plain})
	assert.Equal(t, 80, quodlibetScoreAlarich(q)[3])
}

// **途中の 4 トリックは無料。** 1・2・3 と最終だけに点が付く。
func TestQuodlibet_ScoreFirstThreeAndLast(t *testing.T) {
	q := newQuodlibetForTest(t)
	q.currentContract = QuodlibetFirstThreeAndLast
	q.trickWinners = []int{0, 1, 2, 3, 3, 3, 3, 0}
	points := quodlibetScoreFirstThreeAndLast(q)
	assert.Equal(t, 10+100, points[0], "第 1 と最終")
	assert.Equal(t, 20, points[1])
	assert.Equal(t, 30, points[2])
	assert.Equal(t, 0, points[3], "第 4-7 トリックは無料")
}

// **低いハートのほうが重い。** 7-10 が 20 点、J-A が 10 点 ── 普通の
// ハート系と逆で、ここを写し間違えると安い札を押しつける遊びが消える。
func TestQuodlibet_ScoreNoRedsChargesLowHeartsMore(t *testing.T) {
	q := newQuodlibetForTest(t)
	q.currentContract = QuodlibetNoReds
	plain := NewCard(CardDesignSpade, 7, false)

	quodlibetResetTricks(q)
	q.GetPlayer(0).AddTrick([]*Card{NewCard(CardDesignHeart, 7, false), plain, plain, plain})
	q.GetPlayer(1).AddTrick([]*Card{NewCard(CardDesignHeart, 1, false), plain, plain, plain})
	points := quodlibetScoreNoReds(q)
	assert.Equal(t, 20, points[0], "ハートの 7")
	assert.Equal(t, 10, points[1], "ハートのエース")
	assert.Greater(t, points[0], points[1], "低い札のほうが重い")

	// 8 枚すべてを同じトリックで取ると 120 ではなく 100。
	quodlibetResetTricks(q)
	all := make([]*Card, 0, QuodlibetHandSize)
	for _, v := range quodlibetRankOrder {
		all = append(all, NewCard(CardDesignHeart, v, false))
	}
	q.GetPlayer(2).AddTrick(all)
	assert.Equal(t, 100, quodlibetScoreNoReds(q)[2])
}

// **Q と J を同じトリックで取ると 50 ではなく 100。**
func TestQuodlibet_ScoreOberUnter(t *testing.T) {
	q := newQuodlibetForTest(t)
	q.currentContract = QuodlibetOberUnter
	plain := NewCard(CardDesignSpade, 7, false)
	queen := NewCard(CardDesignHeart, 12, false)
	jack := NewCard(CardDesignClover, 11, false)

	quodlibetResetTricks(q)
	q.GetPlayer(0).AddTrick([]*Card{queen, plain, plain, plain})
	q.GetPlayer(1).AddTrick([]*Card{jack, plain, plain, plain})
	q.GetPlayer(2).AddTrick([]*Card{queen, jack, plain, plain})
	points := quodlibetScoreOberUnter(q)
	assert.Equal(t, 30, points[0])
	assert.Equal(t, 20, points[1])
	assert.Equal(t, 100, points[2])
}

// **賄賂は取らなかった人にも点が付く。** 最低の札を出した席が 20 を負う。
func TestQuodlibet_ScoreBribeChargesTheLowestCardToo(t *testing.T) {
	q := newQuodlibetForTest(t)
	q.currentContract = QuodlibetBribe
	q.trickWinners = []int{1}
	q.trickRecord = [][]*TrickCard{{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 7, false)}, // 最低
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 1, false)}, // 勝ち
		{PlayerIdx: 2, Card: NewCard(CardDesignSpade, 9, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 10, false)},
	}}
	points := quodlibetScoreBribe(q)
	assert.Equal(t, 20, points[0], "最低の札を出した席")
	assert.Equal(t, 30, points[1], "取った席")
	assert.Equal(t, 0, points[2])
}

// **最低の札で取ると 50 ではなく 100。**
func TestQuodlibet_ScoreBribeWinningWithTheLowestCardIsTheDisaster(t *testing.T) {
	q := newQuodlibetForTest(t)
	q.currentContract = QuodlibetBribe
	q.trickWinners = []int{0}
	// 全員がスートを外し、リードの 7 がそのまま残る。
	q.trickRecord = [][]*TrickCard{{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 7, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 1, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 13, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignHeart, 12, false)},
	}}
	points := quodlibetScoreBribe(q)
	assert.Equal(t, 100, points[0])
	assert.Equal(t, 0, points[1], "他の席には何も付かない")
}

// **上がりが遅いほど 1 枚が高くつく。** 最初に上がった人は 0。
func TestQuodlibet_ScoreShedding(t *testing.T) {
	q := newQuodlibetForTest(t)
	q.currentContract = QuodlibetSnack
	for i, p := range q.GetPlayers() {
		p.Reset()
		p.SetOutRank(i + 1)
	}
	q.GetPlayer(1).AddCard(NewCard(CardDesignSpade, 7, false))
	q.GetPlayer(2).AddCard(NewCard(CardDesignSpade, 8, false))
	q.GetPlayer(2).AddCard(NewCard(CardDesignSpade, 9, false))
	q.GetPlayer(3).AddCard(NewCard(CardDesignSpade, 10, false))
	points := quodlibetScoreShedding(q)
	assert.Equal(t, 0, points[0], "最初に上がった人は無罰")
	assert.Equal(t, 10, points[1], "2 位は 1 枚 10 点")
	assert.Equal(t, 40, points[2], "3 位は 1 枚 20 点 × 2 枚")
	assert.Equal(t, 30, points[3], "4 位は 1 枚 30 点")
}

// **第 3 の輪では手札の見え方そのものが規則。**
func TestQuodlibetHandVisibility(t *testing.T) {
	// 開いたズボン: 自分の手札だけが見えない。
	assert.False(t, QuodlibetHandVisibility(QuodlibetOpen, 0, 0))
	assert.True(t, QuodlibetHandVisibility(QuodlibetOpen, 0, 1))
	// 狩猟: 全員の手札が見える。
	assert.True(t, QuodlibetHandVisibility(QuodlibetHunt, 0, 0))
	assert.True(t, QuodlibetHandVisibility(QuodlibetHunt, 0, 3))
	// それ以外: 自分の手札だけが見える。
	assert.True(t, QuodlibetHandVisibility(QuodlibetMinus, 0, 0))
	assert.False(t, QuodlibetHandVisibility(QuodlibetMinus, 0, 2))
}

// **フォロー義務はある。** 台札スートを持っていれば、それしか出せない。
func TestQuodlibet_MustFollowSuit(t *testing.T) {
	q := newQuodlibetForTest(t)
	q.currentContract = QuodlibetMinus
	q.phase = QuodlibetPhasePlay
	p := q.GetPlayer(0)
	p.Reset()
	p.AddCard(NewCard(CardDesignSpade, 7, false))
	p.AddCard(NewCard(CardDesignHeart, 9, false))
	q.currentTrick = []*TrickCard{{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 13, false)}}
	assert.Equal(t, []int{0}, q.GetPlayableIndices(0))

	// 台札スートを持っていなければ何でも出せる。
	p.Reset()
	p.AddCard(NewCard(CardDesignHeart, 9, false))
	p.AddCard(NewCard(CardDesignClover, 8, false))
	assert.Equal(t, []int{0, 1}, q.GetPlayableIndices(0))
}

// **四分は同じスートのちょうど 3 つ上でしか重ねられない。**
func TestQuodlibet_QuadratureNeedsExactlyThreeHigher(t *testing.T) {
	q := newQuodlibetForTest(t)
	q.currentContract = QuodlibetQuadrature
	q.stack = []*Card{NewCard(CardDesignSpade, 7, false)}
	assert.True(t, q.isQuadratureCardPlayable(NewCard(CardDesignSpade, 10, false)), "7 の 3 つ上は 10")
	assert.False(t, q.isQuadratureCardPlayable(NewCard(CardDesignSpade, 9, false)), "2 つ上は不可")
	assert.False(t, q.isQuadratureCardPlayable(NewCard(CardDesignSpade, 11, false)), "4 つ上は不可")
	assert.False(t, q.isQuadratureCardPlayable(NewCard(CardDesignHeart, 10, false)), "別スートは不可")

	// エースは値が 1 でも並びの最上位。J の 3 つ上。
	q.stack = []*Card{NewCard(CardDesignSpade, 11, false)}
	assert.True(t, q.isQuadratureCardPlayable(NewCard(CardDesignSpade, 1, false)))

	// 重ねが空なら何でもリードできる。
	q.stack = nil
	assert.True(t, q.isQuadratureCardPlayable(NewCard(CardDesignHeart, 12, false)))
}

// **小食いの起点は J。** 32 枚デッキは 7 から A なので 7 ではない。
func TestQuodlibet_SnackBuildsOutFromTheJack(t *testing.T) {
	q := newQuodlibetForTest(t)
	q.currentContract = QuodlibetSnack
	assert.True(t, q.isSnackCardPlayable(NewCard(CardDesignSpade, 11, false)), "J は常に置ける")
	assert.False(t, q.isSnackCardPlayable(NewCard(CardDesignSpade, 7, false)), "隣が空なら置けない")

	q.tablePlaced[CardDesignSpade] |= 1 << uint(QuodlibetSnackAnchorIndex)
	assert.True(t, q.isSnackCardPlayable(NewCard(CardDesignSpade, 10, false)), "J の下")
	assert.True(t, q.isSnackCardPlayable(NewCard(CardDesignSpade, 12, false)), "J の上")
	assert.False(t, q.isSnackCardPlayable(NewCard(CardDesignSpade, 11, false)), "置いた札は置けない")
	assert.False(t, q.isSnackCardPlayable(NewCard(CardDesignHeart, 10, false)), "別スートはまだ J から")
}

func TestQuodlibetRankIndex(t *testing.T) {
	assert.Equal(t, 0, QuodlibetRankIndex(7))
	assert.Equal(t, 4, QuodlibetRankIndex(11))
	// **エースは値 1 でも並びの最上位。**
	assert.Equal(t, 7, QuodlibetRankIndex(1))
	assert.Equal(t, -1, QuodlibetRankIndex(5), "32 枚デッキに無い値")
}

// **1 ゲームを通しで打てる。** 12 ディールすべてが終局に届く。
func TestQuodlibet_PlaysAFullMatch(t *testing.T) {
	q := newQuodlibetForTest(t)
	quodlibetPlayMatch(t, q)
	assert.Equal(t, QuodlibetPhaseGameEnd, q.GetPhase())
	assert.Len(t, q.GetDealHistory(), QuodlibetTotalDeals)
	// 12 種目すべてが 1 回ずつ打たれている。
	seen := map[int]int{}
	for _, d := range q.GetDealHistory() {
		seen[d.Contract]++
	}
	assert.Len(t, seen, QuodlibetContractCnt)
	for c, n := range seen {
		assert.Equal(t, 1, n, "%s が %d 回打たれている", QuodlibetContractName(c), n)
	}
}

// **勝つのは罰点が一番少ない人。** 多いほうではない。
func TestQuodlibet_LowestPenaltyWins(t *testing.T) {
	q := newQuodlibetForTest(t)
	for _, p := range q.GetPlayers() {
		p.ResetPenalty()
	}
	q.GetPlayer(0).AddPenalty(300)
	q.GetPlayer(1).AddPenalty(120)
	q.GetPlayer(2).AddPenalty(120)
	q.GetPlayer(3).AddPenalty(400)
	assert.Equal(t, []int{1, 2}, q.GetWinners())
}

func TestQuodlibet_ContractNames(t *testing.T) {
	seen := map[string]bool{}
	for c := 0; c < QuodlibetContractCnt; c++ {
		name := QuodlibetContractName(c)
		assert.NotEqual(t, "unknown", name, "コントラクト %d に名前が無い", c)
		assert.False(t, seen[name], "%q が重複している", name)
		seen[name] = true
	}
	assert.Equal(t, "unknown", QuodlibetContractName(QuodlibetContractCnt))
}

func TestQuodlibet_HintFollowsThePhase(t *testing.T) {
	q := newQuodlibetForTest(t)
	// 席 0 が最初のディーラーなので、開幕はコントラクト選択。
	hint := q.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "pick_contract", hint.Reason)
	assert.Contains(t, q.GetAvailableContracts(), hint.Contract)

	require.NoError(t, q.SelectContract(QuodlibetMinus))
	for q.GetPhase() == QuodlibetPhasePlay && !q.IsHumanTurn() {
		q.CpuPlay()
	}
	if q.GetPhase() == QuodlibetPhasePlay {
		hint = q.GetHint()
		require.Len(t, hint.CardIndices, 1)
		assert.Contains(t, q.GetPlayableIndices(0), hint.CardIndices[0])
	}
}

// **助言は CPU の難易度に引きずられない。** Easy でランダムに選ぶ関数を
// そのまま使うと、Easy を選んだ人にだけでたらめな札を勧めることになる。
func TestQuodlibet_HintIgnoresCpuDifficulty(t *testing.T) {
	q := NewQuodlibet(NewTrumpCards32(), quodlibetTestPlayers(), QuodlibetConfig{
		CpuDifficulty: QuodlibetCpuDifficultyEasy,
	})
	q.Reset()
	require.NoError(t, q.SelectContract(QuodlibetMinus))

	// **人間のリードに固定する。** 配り任せだと台札スートの持ち札が 1 枚の
	// ディールに当たり、合法手が 1 つしか無い ＝ ランダムでも同じ札が返る
	// 局面で「ぶれない」を確かめたことにしてしまう。
	q.currentPlayer = 0
	q.leadPlayer = 0
	q.currentTrick = nil
	require.Greater(t, len(q.GetPlayableIndices(0)), 1, "リードなら手札すべてが合法手のはず")

	want := q.GetHint().CardIndices[0]
	for i := 0; i < 20; i++ {
		hint := q.GetHint()
		require.Len(t, hint.CardIndices, 1)
		assert.Equal(t, want, hint.CardIndices[0], "%d 回目でぶれた", i+1)
	}

	// コントラクト選択の助言も同じ ── Easy でもぶれない。
	q.phase = QuodlibetPhaseSelectContract
	q.dealerIdx = 0
	wantContract := q.GetHint().Contract
	for i := 0; i < 20; i++ {
		assert.Equal(t, wantContract, q.GetHint().Contract, "%d 回目で種目がぶれた", i+1)
	}
}

func TestQuodlibet_RejectsOutOfTurnAndBadIndex(t *testing.T) {
	q := newQuodlibetForTest(t)
	assert.ErrorIs(t, q.PlayerPlay(0), ErrWrongPhase, "選択フェーズでは出せない")

	require.NoError(t, q.SelectContract(QuodlibetMinus))
	for q.GetPhase() == QuodlibetPhasePlay && !q.IsHumanTurn() {
		q.CpuPlay()
	}
	if q.GetPhase() != QuodlibetPhasePlay {
		t.Skip("配りによっては人間の手番の前にトリックが揃う")
	}
	assert.Error(t, q.PlayerPlay(99))
	assert.Error(t, q.PlayerPlay(-1))
}

func TestQuodlibet_SelectContractGuards(t *testing.T) {
	q := newQuodlibetForTest(t)
	assert.Error(t, q.SelectContract(-1))
	assert.Error(t, q.SelectContract(QuodlibetContractCnt))
	require.NoError(t, q.SelectContract(QuodlibetPlus))
	assert.ErrorIs(t, q.SelectContract(QuodlibetMinus), ErrWrongPhase)
}

// **CPU がディーラーの輪でも盤面は進む。** 選ばれないまま止まると
// 何も打てない。
func TestQuodlibet_CpuDealerPicksAContract(t *testing.T) {
	q := newQuodlibetForTest(t)
	require.NoError(t, q.SelectContract(QuodlibetPlus))
	quodlibetPlayDeal(t, q)
	q.NextDeal()
	require.False(t, q.GetPlayer(q.GetDealerIdx()).GetIsHuman(), "第 2 ディールのディーラーは CPU")
	require.Equal(t, QuodlibetPhaseSelectContract, q.GetPhase())
	q.CpuSelectContract()
	assert.Equal(t, QuodlibetPhasePlay, q.GetPhase())
	assert.GreaterOrEqual(t, q.GetCurrentContract(), 0)
}

// **自動選択にすると人間にも訊かない。** 12 回の選択を省く遊び方。
func TestQuodlibet_AutoSelectSkipsTheChoice(t *testing.T) {
	cfg := DefaultQuodlibetConfig()
	cfg.AutoSelectContract = true
	q := NewQuodlibet(NewTrumpCards32(), quodlibetTestPlayers(), cfg)
	q.Reset()
	assert.Equal(t, QuodlibetPhasePlay, q.GetPhase())
	assert.Equal(t, QuodlibetPlus, q.GetCurrentContract())
}

// **保存した盤で指し続けられる。** 非公開フィールドだけの型は MarshalJSON が
// 無いと `{}` になる。
func TestQuodlibet_SaveRestoreKeepsPlaying(t *testing.T) {
	q := newQuodlibetForTest(t)
	require.NoError(t, q.SelectContract(QuodlibetPlus))
	for q.GetPhase() == QuodlibetPhasePlay && !q.IsHumanTurn() {
		q.CpuPlay()
	}
	data, err := json.Marshal(q)
	require.NoError(t, err)
	require.Greater(t, len(data), 2, "空の JSON になっている")

	restored := new(Quodlibet)
	require.NoError(t, json.Unmarshal(data, restored))
	assert.Equal(t, q.GetPhase(), restored.GetPhase())
	assert.Equal(t, q.GetCurrentContract(), restored.GetCurrentContract())
	assert.Equal(t, q.GetDealerIdx(), restored.GetDealerIdx())
	assert.Equal(t, q.GetCurrentTurn(), restored.GetCurrentTurn())
	assert.Len(t, restored.GetPlayers(), QuodlibetPlayerCnt)
	for i := range q.GetPlayers() {
		assert.Equal(t, q.GetPlayer(i).GetCardsSize(), restored.GetPlayer(i).GetCardsSize(), "席 %d", i)
	}
	// 復元した盤で最後まで打てる。
	quodlibetPlayMatch(t, restored)
}

func TestQuodlibet_RejectsTamperedSnapshot(t *testing.T) {
	restored := new(Quodlibet)
	assert.Error(t, restored.UnmarshalJSON([]byte("{")))
	assert.Error(t, restored.UnmarshalJSON([]byte(`{"pl":[]}`)))
}

func TestQuodlibetConfig_Validate(t *testing.T) {
	assert.NoError(t, DefaultQuodlibetConfig().Validate())
	assert.Error(t, QuodlibetConfig{CpuDifficulty: -1}.Validate())
	assert.Error(t, QuodlibetConfig{CpuDifficulty: 3}.Validate())
}

// quodlibetTestPlayers は席 0 が人間の 4 人を返す。
func quodlibetTestPlayers() []*QuodlibetPlayer {
	players := make([]*QuodlibetPlayer, QuodlibetPlayerCnt)
	players[0] = NewQuodlibetPlayer(true)
	for i := 1; i < QuodlibetPlayerCnt; i++ {
		players[i] = NewQuodlibetPlayer(false)
	}
	return players
}

// quodlibetResetTricks は全席の獲得トリックを空にする。
func quodlibetResetTricks(q *Quodlibet) {
	for _, p := range q.GetPlayers() {
		p.ResetTricks()
	}
}

// quodlibetGiveTricks は席ごとに指定枚数のダミートリックを持たせる。
func quodlibetGiveTricks(q *Quodlibet, counts []int) {
	quodlibetResetTricks(q)
	dummy := []*Card{NewCard(CardDesignSpade, 7, false)}
	for i, n := range counts {
		for j := 0; j < n; j++ {
			q.GetPlayer(i).AddTrick(dummy)
		}
	}
}
