//go:build test

package domain

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newBotifarraWithHuman は席 0 が人間の卓を返し、宣言と倍付けを済ませてプレイ直前まで進める。
//
// **観測には人間席が要ります。** 全席 CPU だと advanceCpu がラウンドを一気に
// 進めてしまい、トリックの途中を一度も見られません。
func newBotifarraWithHuman(t *testing.T, seed int64) *Botifarra {
	t.Helper()
	g := NewDefaultBotifarra()
	g.SetRand(rand.New(rand.NewSource(seed)))
	g.Reset()
	if g.GetPhase() == BotifarraPhaseDeclare || g.GetPhase() == BotifarraPhaseDelegated {
		require.NoError(t, g.Declare(g.longestSuitOf(0)))
	}
	for g.GetPhase() == BotifarraPhaseDouble && g.IsHumanTurn() {
		require.NoError(t, g.PassDouble())
	}
	return g
}

// botifarraDriveHuman は人間の手番で observe を呼び、合法な札を1枚出して進める。
// 戻り値は observe が呼ばれた回数。
func botifarraDriveHuman(t *testing.T, g *Botifarra, rounds int, observe func(*Botifarra)) int {
	t.Helper()
	seen := 0
	for step := 0; step < rounds && !g.GetGameEndFlag(); step++ {
		if g.GetPhase() == BotifarraPhaseRoundEnd {
			require.NoError(t, g.NextRound())
			if g.GetPhase() == BotifarraPhaseDeclare || g.GetPhase() == BotifarraPhaseDelegated {
				require.NoError(t, g.Declare(g.longestSuitOf(0)))
			}
			for g.GetPhase() == BotifarraPhaseDouble && g.IsHumanTurn() {
				require.NoError(t, g.PassDouble())
			}
			continue
		}
		if g.GetPhase() != BotifarraPhasePlay || !g.IsHumanTurn() {
			g.CpuPlay()
			continue
		}
		observe(g)
		seen++
		valid := g.GetValidPlayIndices(0)
		require.NotEmpty(t, valid, "人間に出せる札が無い")
		require.NoError(t, g.PlayCard(valid[0]))
	}
	return seen
}

// newBotifarraAllCpu は全席 CPU の卓を返す (自動進行の観測用)。
func newBotifarraAllCpu(t *testing.T, seed int64) *Botifarra {
	t.Helper()
	seats := make([]*BotifarraPlayer, BotifarraPlayerCnt)
	for i := range seats {
		seats[i] = NewBotifarraPlayer(false)
	}
	g := NewBotifarra(NewTrumpCardsReversis(), seats, DefaultBotifarraConfig())
	g.SetRand(rand.New(rand.NewSource(seed)))
	return g
}

// **デッキはスペイン式 48 枚で、4 人にちょうど配り切れる。**
func TestBotifarraDealsTwelveEachWithNothingLeft(t *testing.T) {
	t.Parallel()

	// **人間席が要ります。** 全席 CPU だと Reset() の中でラウンドが終わり切って
	// しまい、手札が 0 枚になったところを見ることになります。
	g := NewDefaultBotifarra()
	g.SetRand(rand.New(rand.NewSource(1)))
	g.Reset()
	require.Equal(t, BotifarraPhaseDeclare, g.GetPhase(), "人間の宣言待ちで止まる")
	for i := range BotifarraPlayerCnt {
		assert.Equal(t, BotifarraHandSize, g.GetPlayer(i).GetCardsSize(), "席 %d", i)
	}
	// 10 は入っていない。
	for i := range BotifarraPlayerCnt {
		p := g.GetPlayer(i)
		for k := range p.GetCardsSize() {
			assert.NotEqual(t, 10, p.GetCard(k).GetValue(), "スペイン式に 10 は無い")
		}
	}
}

// **点の合計は定数と一致する。** デッキや点表を変えたらここで落ちます。
func TestBotifarraDeckPointsMatchTheConstant(t *testing.T) {
	t.Parallel()

	require.NoError(t, botifarraValidateDeckPoints())

	deck := NewTrumpCardsReversis()
	total := 0
	for range 48 {
		total += BotifarraCardPoint(deck.DrawCard())
	}
	assert.Equal(t, BotifarraCardPoints, total, "札の点の合計")
	assert.Equal(t, 72, BotifarraTotalPoints, "札 60 + トリック 12")
	assert.Equal(t, 36, BotifarraHalfPoints)
}

// **点札がそのまま上位 5 枚。** マニラ (9) がいちばん強く、いちばん高い。
func TestBotifarraRankOrder(t *testing.T) {
	t.Parallel()

	order := []int{9, 1, 13, 12, 11, 8, 7, 6, 5, 4, 3, 2}
	for i := 1; i < len(order); i++ {
		prev := NewCard(CardDesignSpade, order[i-1], false)
		cur := NewCard(CardDesignSpade, order[i], false)
		assert.Greater(t, BotifarraRank(prev), BotifarraRank(cur),
			"%d は %d より強い", order[i-1], order[i])
	}
	assert.Zero(t, BotifarraRank(nil))
	assert.Zero(t, BotifarraCardPoint(nil))
	assert.Equal(t, 5, BotifarraCardPoint(NewCard(CardDesignHeart, 9, false)), "マニラは 5 点")
	assert.Zero(t, BotifarraCardPoint(NewCard(CardDesignHeart, 2, false)), "2 は点にならない")
}

// **相方は向かい。** 席 0-2 と 1-3 が組です。
func TestBotifarraSeating(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 2, BotifarraPartnerOf(0))
	assert.Equal(t, 3, BotifarraPartnerOf(1))
	assert.Equal(t, 0, BotifarraPartnerOf(2))
	assert.Equal(t, BotifarraTeamOf(0), BotifarraTeamOf(2))
	assert.Equal(t, BotifarraTeamOf(1), BotifarraTeamOf(3))
	assert.NotEqual(t, BotifarraTeamOf(0), BotifarraTeamOf(1))
}

// **CPU 同士で最後まで進み切ることを数える。**
//
// 「勝てるなら勝たなければならない」という義務があるので、出せる札の集合が空に
// なると盤面が止まります。テストが全部緑でも止まっているだけ、ということが
// あり得るので、**打ち切りに落ちた回数を実際に数えます**。
func TestBotifarraAlwaysTerminates(t *testing.T) {
	const games = 60
	stalled := 0

	for seed := range games {
		g := newBotifarraAllCpu(t, int64(seed)+1)
		g.Reset()

		steps := 0
		for !g.GetGameEndFlag() {
			steps++
			if steps > 2000 {
				stalled++
				break
			}
			if g.GetPhase() == BotifarraPhaseRoundEnd {
				require.NoError(t, g.NextRound())
				continue
			}
			// 全席 CPU なので、進んでいなければ止まっている。
			before := g.GetTrickCount() + g.GetScore(0) + g.GetScore(1)
			g.CpuPlay()
			after := g.GetTrickCount() + g.GetScore(0) + g.GetScore(1)
			if before == after && g.GetPhase() == BotifarraPhasePlay && len(g.GetTrick()) == 0 {
				stalled++
				break
			}
		}
	}
	assert.Zero(t, stalled, "%d 局中 %d 局が進まなくなった", games, stalled)
}

// **1 ラウンドで動く点はちょうど 72。** 取りこぼしも二重計上もありません。
//
// **全席 CPU だと Reset() の中で 1 ラウンド目が終わり切ります**（人間の入力待ちが
// 無いため）。さらに recontrar が乗ると (72-36)*4 = 144 点入るので、**1 ラウンドで
// 101 点に届いてゲームが終わることがあります**——そこも数えられるように、
// まず「Reset 直後の 1 ラウンド」を検査してから続きを回します。
func TestBotifarraRoundPointsAlwaysSumTo72(t *testing.T) {
	for seed := range 40 {
		g := newBotifarraAllCpu(t, int64(seed)+1)
		g.Reset()

		rounds := 0
		for {
			require.Equal(t, BotifarraTrickCnt, g.GetTrickCount(),
				"seed %d ラウンド %d: 12 トリックで終わらなかった", seed, rounds)
			sum := g.GetRoundPoints(0) + g.GetRoundPoints(1)
			require.Equal(t, BotifarraTotalPoints, sum,
				"seed %d ラウンド %d: 点の合計が %d", seed, rounds, sum)
			rounds++

			if g.GetGameEndFlag() || rounds >= 30 {
				break
			}
			require.NoError(t, g.NextRound())
			for g.GetPhase() != BotifarraPhaseRoundEnd && !g.GetGameEndFlag() {
				g.CpuPlay()
			}
		}
		require.Positive(t, rounds)
	}
}

// **出せる札が無くなることは無い。** 義務が強すぎると盤面が止まります。
func TestBotifarraAlwaysHasALegalPlay(t *testing.T) {
	for seed := range 30 {
		g := newBotifarraAllCpu(t, int64(seed)+1)
		g.Reset()

		for !g.GetGameEndFlag() {
			if g.GetPhase() == BotifarraPhaseRoundEnd {
				require.NoError(t, g.NextRound())
				continue
			}
			if g.GetPhase() == BotifarraPhasePlay {
				idx := g.GetCurrentTurn()
				if g.GetPlayer(idx).GetCardsSize() > 0 {
					require.NotEmpty(t, g.GetValidPlayIndices(idx),
						"seed %d: 席 %d に出せる札が無い", seed, idx)
				}
			}
			g.CpuPlay()
		}
	}
}

// **勝てるなら勝たなければならない。** 上回れる札があるのに安い札は出せません。
func TestBotifarraMustBeatWhenAble(t *testing.T) {
	checked := 0
	for seed := range 12 {
		g := newBotifarraWithHuman(t, int64(seed)+1)
		checked += botifarraDriveHuman(t, g, 400, func(g *Botifarra) {
			if len(g.GetTrick()) == 0 {
				return
			}
			p := g.GetPlayer(0)
			winner := ResolveTrickWinner(g.GetTrick(), g.GetTrumpSuit(), BotifarraRank)
			if BotifarraTeamOf(winner) == BotifarraTeamOf(0) || g.trickWonByTrump() {
				return
			}
			leadSuit := g.GetTrick()[0].Card.GetDesign()
			best := g.bestRankInTrick()
			canBeat := false
			for k := range p.GetCardsSize() {
				c := p.GetCard(k)
				if c.GetDesign() == leadSuit && BotifarraRank(c) > best {
					canBeat = true
					break
				}
			}
			if !canBeat {
				return
			}
			// **上回れるなら、選べるのは上回る札だけ。**
			for _, v := range g.GetValidPlayIndices(0) {
				assert.Greater(t, BotifarraRank(p.GetCard(v)), best,
					"勝てるのに勝たない札が選べてしまう")
			}
		})
	}
	assert.Positive(t, checked, "人間の手番を一度も踏めなかった")
}

// **フォローできるならフォローする。**
func TestBotifarraMustFollowSuit(t *testing.T) {
	checked := 0
	for seed := range 12 {
		g := newBotifarraWithHuman(t, int64(seed)+100)
		checked += botifarraDriveHuman(t, g, 400, func(g *Botifarra) {
			if len(g.GetTrick()) == 0 {
				return
			}
			p := g.GetPlayer(0)
			leadSuit := g.GetTrick()[0].Card.GetDesign()
			has := false
			for k := range p.GetCardsSize() {
				if p.GetCard(k).GetDesign() == leadSuit {
					has = true
					break
				}
			}
			if !has {
				return
			}
			for _, v := range g.GetValidPlayIndices(0) {
				assert.Equal(t, leadSuit, p.GetCard(v).GetDesign(), "フォローを外せてしまう")
			}
		})
	}
	assert.Positive(t, checked)
}

// **36 を超えたぶんだけが得点になる。**
func TestBotifarraOnlyThePointsAboveHalfScore(t *testing.T) {
	for seed := range 25 {
		// **全席 CPU だと Reset() の中で 1 ラウンド目が終わり切ります。**
		// なので「1 ラウンド後」の状態をそのまま読みます。
		g := newBotifarraAllCpu(t, int64(seed)+1)
		g.Reset()
		require.Equal(t, BotifarraTrickCnt, g.GetTrickCount(), "seed %d", seed)
		gained := [2]int{g.GetScore(0), g.GetScore(1)}

		// 片方だけが得点する (36-36 なら両方 0)。
		assert.False(t, gained[0] > 0 && gained[1] > 0, "seed %d: 両チームが得点した", seed)
		for team := range 2 {
			want := g.GetRoundPoints(team) - BotifarraHalfPoints
			if want <= 0 {
				assert.Zero(t, gained[team], "seed %d: 半分以下なのに得点した", seed)
				continue
			}
			assert.Equal(t, want*g.GetMultiplier(), gained[team], "seed %d: チーム %d", seed, team)
		}
	}
}

// **切り札なしでも成立する。** ボティファラ宣言では誰も切り札を持ちません。
//
// 盤面を手で組まず、**人間が実際に宣言する経路**で踏みます。組んだ状態は
// 到達不能なことがあり、そこで落ちても本番のバグではありません。
func TestBotifarraNoTrumpRound(t *testing.T) {
	g := NewDefaultBotifarra()
	g.SetRand(rand.New(rand.NewSource(5)))
	g.Reset()
	require.Equal(t, BotifarraPhaseDeclare, g.GetPhase())

	require.NoError(t, g.Declare(BotifarraNoTrump))
	assert.Equal(t, BotifarraNoTrump, g.GetTrumpSuit())
	for g.GetPhase() == BotifarraPhaseDouble && g.IsHumanTurn() {
		require.NoError(t, g.PassDouble())
	}

	// **このラウンドだけを見ます。** 次のラウンドでは CPU が普通の切り札を
	// 宣言するので、そこまで回すと「切り札なし」の性質は当然崩れます。
	for step := 0; step < 200 && g.GetPhase() == BotifarraPhasePlay; step++ {
		if !g.IsHumanTurn() {
			g.CpuPlay()
			continue
		}
		// 切り札が無いので、切り札で取る義務は発生しません。
		assert.False(t, g.trickWonByTrump())
		valid := g.GetValidPlayIndices(0)
		require.NotEmpty(t, valid)
		require.NoError(t, g.PlayCard(valid[0]))
	}

	assert.Equal(t, BotifarraTrickCnt, g.GetTrickCount())
	assert.Equal(t, BotifarraTotalPoints, g.GetRoundPoints(0)+g.GetRoundPoints(1))
}

func TestBotifarraConfigValidate(t *testing.T) {
	t.Parallel()

	assert.NoError(t, DefaultBotifarraConfig().Validate())
	assert.Error(t, BotifarraConfig{TargetScore: BotifarraTargetMin - 1}.Validate())
	assert.Error(t, BotifarraConfig{TargetScore: BotifarraTargetMax + 1}.Validate())

	// **壊れた設定は既定に落とす。**
	g := NewBotifarra(nil, nil, BotifarraConfig{TargetScore: 9999})
	assert.Equal(t, DefaultBotifarraConfig(), g.GetConfig())
	assert.Equal(t, BotifarraPlayerCnt, g.GetPlayerCnt())
	assert.True(t, g.GetPlayer(0).GetIsHuman(), "席 0 は人間")

	g.SetConfig(BotifarraConfig{TargetScore: 151, AllowDoubling: false})
	assert.Equal(t, 151, g.GetConfig().TargetScore)
	g.SetConfig(BotifarraConfig{TargetScore: 1})
	assert.Equal(t, 151, g.GetConfig().TargetScore, "範囲外は無視する")
}

func TestBotifarraDeclareRejectsBadInput(t *testing.T) {
	g := NewDefaultBotifarra()
	g.SetRand(rand.New(rand.NewSource(1)))
	g.Reset()

	// 親は席 0 (人間) なので宣言の番。
	require.Equal(t, BotifarraPhaseDeclare, g.GetPhase())
	require.Equal(t, 0, g.GetCurrentTurn())
	assert.Error(t, g.Declare(99), "スートが範囲外")
	assert.Error(t, g.PlayCard(0), "まだプレイの場面ではない")
	assert.Error(t, g.NextRound(), "まだラウンドの区切りではない")

	require.NoError(t, g.Declare(CardDesignSpade))
	assert.Equal(t, CardDesignSpade, g.GetTrumpSuit())
	assert.Equal(t, 0, g.GetDeclarerIdx())
	assert.Error(t, g.Declare(CardDesignHeart), "二度は宣言できない")
}

// **委ねられた側は必ず宣言する。** 降りる選択肢はありません。
func TestBotifarraDelegatePassesToThePartner(t *testing.T) {
	g := NewDefaultBotifarra()
	g.SetRand(rand.New(rand.NewSource(2)))
	g.Reset()
	require.Equal(t, BotifarraPhaseDeclare, g.GetPhase())

	require.NoError(t, g.Delegate())
	// 相方 (席 2) は CPU なので、その場で宣言まで進みます。
	assert.NotEqual(t, -1, g.GetDeclarerIdx(), "委ねた先が宣言していない")
	assert.Equal(t, BotifarraPartnerOf(0), g.GetDeclarerIdx())
	assert.Error(t, g.Delegate(), "二度は委ねられない")
}

func TestBotifarraGiveUp(t *testing.T) {
	g := NewDefaultBotifarra()
	g.SetRand(rand.New(rand.NewSource(1)))
	g.Reset()

	g.GiveUp()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 1, g.GetWinnerTeam(), "投了した側は勝たない")
	assert.False(t, g.IsHumanTurn())
	assert.Nil(t, g.GetHint())

	g.GiveUp()
	assert.True(t, g.GetGameEndFlag())
}

func TestBotifarraAccessorsRejectOutOfRange(t *testing.T) {
	t.Parallel()

	g := NewDefaultBotifarra()
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(99))
	assert.Zero(t, g.GetRoundPoints(-1))
	assert.Zero(t, g.GetRoundPoints(9))
	assert.Zero(t, g.GetScore(-1))
	assert.Zero(t, g.GetScore(9))
	assert.Nil(t, g.GetValidPlayIndices(-1))
	assert.Nil(t, g.GetValidPlayIndices(99))

	assert.Equal(t, "spade", botifarraTrumpName(CardDesignSpade))
	assert.Equal(t, "clover", botifarraTrumpName(CardDesignClover))
	assert.Equal(t, "heart", botifarraTrumpName(CardDesignHeart))
	assert.Equal(t, "diamond", botifarraTrumpName(CardDesignDiamond))
	assert.Equal(t, "notrump", botifarraTrumpName(BotifarraNoTrump))
}

func TestBotifarraHint(t *testing.T) {
	g := NewDefaultBotifarra()
	g.SetRand(rand.New(rand.NewSource(1)))
	g.Reset()

	h := g.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "botifarraDeclareLongest", h.Reason)
	require.NotNil(t, h.Suit)
	assert.Nil(t, h.CardIndex)

	require.NoError(t, g.Declare(*h.Suit))
	for g.GetPhase() == BotifarraPhaseDouble {
		require.NoError(t, g.PassDouble())
	}
	if g.GetPhase() == BotifarraPhasePlay && g.IsHumanTurn() {
		h = g.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "botifarraMustWin", h.Reason)
		require.NotNil(t, h.CardIndex)
		assert.Contains(t, g.GetValidPlayIndices(0), *h.CardIndex, "助言が違法な札を指している")
	}
}

// **切り札がリードされたときも「勝てるなら勝つ」義務は効く。**
//
// `trickWonByTrump()` はリード札が切り札でも真になるので、素直に除外すると
// **切り札リードのトリックだけ義務が外れ**、上位の切り札を持っていても安い切り札で
// 逃げられてしまいます。切り札を引き出すためのリードは定石なので、めったに起きない
// 場面ではありません。
//
// 既存の TestBotifarraMustBeatWhenAble は `trickWonByTrump()` が真だと観測を
// 打ち切るため、この経路を一度も踏んでいませんでした。
func TestBotifarraMustBeatWhenTrumpIsLed(t *testing.T) {
	t.Parallel()

	g := newBotifarraAllCpu(t, 1)
	g.Reset()

	trump := CardDesignSpade
	g.trumpSuit = trump
	g.phase = BotifarraPhasePlay
	g.declarerIdx = 0

	// 席 1 に「勝てる切り札」と「勝てない切り札」を1枚ずつだけ持たせる。
	seat := g.players[1]
	seat.Reset()
	manille := NewCard(trump, 9, false) // rank 12 — 場を上回れる
	sota := NewCard(trump, 11, false)   // rank 8  — 上回れない
	seat.AddCard(sota)
	seat.AddCard(manille)

	// 席 0（相手チーム）が騎士の切り札でリード。rank 9。
	lead := NewCard(trump, 12, false)
	g.trick = []*TrickCard{{PlayerIdx: 0, Card: lead}}
	g.currentTurn = 1

	require.NotEqual(t, BotifarraTeamOf(0), BotifarraTeamOf(1), "味方ではない")
	require.True(t, g.trickWonByTrump(), "リード札が切り札なので真になる")
	require.Equal(t, BotifarraRank(lead), g.bestRankInTrick())

	valid := g.GetValidPlayIndices(1)
	require.NotEmpty(t, valid)
	for _, v := range valid {
		c := seat.GetCard(v)
		assert.Greater(t, BotifarraRank(c), BotifarraRank(lead),
			"勝てるのに安い切り札で逃げられてしまう")
	}
	assert.Len(t, valid, 1, "上回れる札はマニラだけ")
}

// **リードとは別のスートの切り札で取られたら、フォローでは勝てない。**
// このときだけ「上回る義務」は外れます。
func TestBotifarraNoObligationWhenRuffedByAnotherSuit(t *testing.T) {
	t.Parallel()

	g := newBotifarraAllCpu(t, 2)
	g.Reset()

	trump := CardDesignSpade
	lead := CardDesignHeart
	g.trumpSuit = trump
	g.phase = BotifarraPhasePlay
	g.declarerIdx = 0

	seat := g.players[1]
	seat.Reset()
	seat.AddCard(NewCard(lead, 11, false))
	seat.AddCard(NewCard(lead, 9, false)) // リードスートの最強でも切り札には勝てない

	g.trick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(lead, 12, false)},
		{PlayerIdx: 3, Card: NewCard(trump, 2, false)}, // 別スートの切り札で取られた
	}
	g.currentTurn = 1

	require.True(t, g.trickWonByTrump())
	// フォローはするが、上回る義務は無い（勝てないので）。
	valid := g.GetValidPlayIndices(1)
	assert.Len(t, valid, 2, "リードスートの札は両方出せる")
}

// **倍付けは配りの価値に掛かる。** 契約側かどうかで変わりません。
func TestBotifarraMultiplierAppliesToWhicheverTeamScores(t *testing.T) {
	t.Parallel()

	for _, declarer := range []int{0, 1} {
		g := newBotifarraAllCpu(t, 3)
		g.Reset()
		g.roundPoints = [2]int{50, 22}
		g.scores = [2]int{}
		g.declarerIdx = declarer
		g.multiplier = BotifarraMultiplierContrar
		g.trickCount = BotifarraTrickCnt
		g.finishRound()

		// 50 - 36 = 14、×2 = 28。宣言者がどちらでも同じ。
		assert.Equal(t, 28, g.GetScore(0), "declarer=%d", declarer)
		assert.Zero(t, g.GetScore(1), "declarer=%d", declarer)
	}
}
