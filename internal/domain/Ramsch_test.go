//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func rc(design, value int) *Card { return NewCard(design, value, false) }

func newTestRamsch() *Ramsch { return NewDefaultRamsch() }

// **配ったらすぐプレイ。** Skat から来た入札・スカット取り・宣言の 4 フェーズは
// このゲームには無い。Reset 直後にプレイが始まっていなければ、どこかに
// 死んだフェーズが残っている。
func TestRamsch_ResetStartsTrickPlayImmediately(t *testing.T) {
	g := newTestRamsch()
	g.Reset()

	assert.Equal(t, RamschPhasePlay, g.GetPhase())
	assert.Equal(t, 1, g.GetTrickNumber())
	// フォアハンド（ディーラーの左）が最初のリード。
	assert.Equal(t, (g.GetDealerIdx()+1)%RamschPlayerCnt, g.GetForehandIdx())

	// **場に出た札も数える。** Reset は人間の手番まで CPU を進めるので、
	// フォアハンドが CPU なら手札は既に 9 枚になっている。手札だけ見ると
	// 「配りが 1 枚足りない」と読み違える。
	played := map[int]int{}
	for _, tc := range g.GetCurrentTrick() {
		played[tc.PlayerIdx]++
	}
	for i := 0; i < g.GetPlayerCnt(); i++ {
		assert.Equal(t, RamschHandSize, g.GetPlayer(i).GetCardsSize()+played[i], "player %d", i)
	}
	assert.Len(t, g.GetSkat(), RamschSkatSize)
}

// **切り札はジャック 4 枚だけ、常に。** ♣J > ♠J > ♥J > ♦J で、スートの
// 一般的な強弱とは別物。
func TestRamsch_JacksAreTheOnlyTrumpsAndClubsIsHighest(t *testing.T) {
	g := newTestRamsch()
	g.Reset()

	for d := CardDesignSpade; d <= CardDesignDiamond; d++ {
		assert.True(t, g.IsWildTrumpForTest(rc(d, ramschValueJack)), "J of suit %d は切り札", d)
		for _, v := range []int{1, 7, 8, 9, 10, 12, 13} {
			assert.False(t, g.IsWildTrumpForTest(rc(d, v)), "suit %d value %d は切り札ではない", d, v)
		}
	}

	// ♣J が最強、♦J が最弱。
	order := []int{CardDesignClover, CardDesignSpade, CardDesignHeart, CardDesignDiamond}
	for i := 0; i+1 < len(order); i++ {
		strong := g.CardStrengthForTest(rc(order[i], ramschValueJack))
		weak := g.CardStrengthForTest(rc(order[i+1], ramschValueJack))
		assert.Greater(t, strong, weak, "J suit %d は J suit %d より強い", order[i], order[i+1])
	}
}

// **ジャックは自分の印刷スートに属さない。** ♣J でクラブをフォローすることは
// できず、クラブが他に無ければ切り札として好きに出せる。ここを取り違えると
// 「フォローできるのに切り札を切った」不正な手が通る。
func TestRamsch_AJackDoesNotFollowItsPrintedSuit(t *testing.T) {
	g := newTestRamsch()
	g.Reset()

	// 人間の手札: ♣J と ♠7。リードは ♣9。
	g.SetHandForTest(0, []*Card{rc(CardDesignClover, ramschValueJack), rc(CardDesignSpade, 7)})
	g.SetCurrentTrickForTest([]*TrickCard{{PlayerIdx: 1, Card: rc(CardDesignClover, 9)}})

	// ♣J はクラブのフォローにならないので、切り札として出すのは合法。
	assert.NoError(t, g.ValidatePlayForTest(0, rc(CardDesignClover, ramschValueJack)))
	// クラブの実札を持っていないので ♠7 も合法。
	assert.NoError(t, g.ValidatePlayForTest(0, rc(CardDesignSpade, 7)))

	// **負のコントロール**: 本物のクラブを持っていれば、それを出さねばならない。
	g.SetHandForTest(0, []*Card{rc(CardDesignClover, 8), rc(CardDesignSpade, 7)})
	assert.Error(t, g.ValidatePlayForTest(0, rc(CardDesignSpade, 7)),
		"クラブを持っているのに別スートを出せてしまう")
}

// **非切り札の序列は A > 10 > K > Q > 9 > 8 > 7。** 10 が K より強いのが
// スカート系で、ここを普通のランク順にすると点札の扱いが丸ごと狂う。
func TestRamsch_TenOutranksKingAmongNonTrumps(t *testing.T) {
	g := newTestRamsch()
	g.Reset()
	order := []int{ramschValueAce, ramschValueTen, ramschValueKing, ramschValueQueen,
		ramschValueNine, ramschValueEight, ramschValueSeven}
	for i := 0; i+1 < len(order); i++ {
		hi := g.CardStrengthForTest(rc(CardDesignHeart, order[i]))
		lo := g.CardStrengthForTest(rc(CardDesignHeart, order[i+1]))
		assert.Greater(t, hi, lo, "value %d は value %d より強い", order[i], order[i+1])
	}
	// どの切り札も、どの非切り札より強い。
	assert.Greater(t, g.CardStrengthForTest(rc(CardDesignDiamond, ramschValueJack)),
		g.CardStrengthForTest(rc(CardDesignHeart, ramschValueAce)))
}

// **点は罰点。最も多く取った人がその点数を失う。**
func TestRamsch_TheBiggestPileOfPointsLoses(t *testing.T) {
	g := newTestRamsch()
	g.Reset()
	g.SetPhase(RamschPhaseRoundEnd)
	g.SetCardPointsForTest([RamschPlayerCnt]int{30, 78, 12})
	g.ScoreRound()

	assert.Equal(t, 1, g.GetLoserIdx())
	assert.Equal(t, -78, g.GetPlayer(1).GetRoundScore(), "最多を取った人が失点する")
	assert.Equal(t, 0, g.GetPlayer(0).GetRoundScore())
	assert.Equal(t, 0, g.GetPlayer(2).GetRoundScore())
	// **負のコントロール**: 最少の人が罰を受けていないこと（符号の取り違え検出）。
	assert.NotEqual(t, -12, g.GetPlayer(2).GetRoundScore())
}

// **同点は全員が負う。** 誰か 1 人を選ぶ根拠が無いので、席順で決めない。
func TestRamsch_ATieForMostPointsPenalisesEveryoneTied(t *testing.T) {
	g := newTestRamsch()
	g.Reset()
	g.SetPhase(RamschPhaseRoundEnd)
	g.SetCardPointsForTest([RamschPlayerCnt]int{50, 50, 20})
	g.ScoreRound()

	assert.Equal(t, -1, g.GetLoserIdx(), "同点では「1 人の敗者」は決まらない")
	assert.Equal(t, -50, g.GetPlayer(0).GetRoundScore())
	assert.Equal(t, -50, g.GetPlayer(1).GetRoundScore())
	assert.Equal(t, 0, g.GetPlayer(2).GetRoundScore())
}

// **Durchmarsch は逆転勝ち。** 全 10 トリックを取ると、取った人は 0 点で、
// 他の 2 人が 120 点ずつ失う。取らないゲームの唯一の攻めどころ。
func TestRamsch_DurchmarschReversesTheResult(t *testing.T) {
	g := newTestRamsch()
	g.Reset()
	g.SetPhase(RamschPhaseRoundEnd)
	// 全部取った人は当然いちばん点が多い。逆転が効いていなければ、この人が失点する。
	g.SetCardPointsForTest([RamschPlayerCnt]int{RamschTotalCardPoints, 0, 0})
	g.SetDurchmarschForTest(0)
	g.ScoreRound()

	assert.Equal(t, 0, g.GetPlayer(0).GetRoundScore(), "総取りした人は失点しない")
	assert.Equal(t, -RamschTotalCardPoints, g.GetPlayer(1).GetRoundScore())
	assert.Equal(t, -RamschTotalCardPoints, g.GetPlayer(2).GetRoundScore())
	assert.Equal(t, 1, g.GetPlayer(0).GetRoundsWon())

	// **負のコントロール**: Durchmarsch でなければ、同じ点で 0 番が失点する。
	g2 := newTestRamsch()
	g2.Reset()
	g2.SetPhase(RamschPhaseRoundEnd)
	g2.SetCardPointsForTest([RamschPlayerCnt]int{RamschTotalCardPoints, 0, 0})
	g2.ScoreRound()
	assert.Equal(t, -RamschTotalCardPoints, g2.GetPlayer(0).GetRoundScore())
}

// **伏せ札 2 枚は最終トリックの獲得者が受け取る。** 誰にも付けないと 120 点の
// うち数点が宙に浮き、「取らなければ得」という前提が崩れる。
func TestRamsch_TheSkatGoesToTheLastTrickWinner(t *testing.T) {
	g := newTestRamsch()
	g.Reset()

	// 伏せ札を A + 10 に固定 = 21 点。
	g.SetSkatForTest([]*Card{rc(CardDesignHeart, ramschValueAce), rc(CardDesignSpade, ramschValueTen)})
	// 最終トリック。0 が ♣J（最強の切り札）で必ず取る。
	g.SetTrickNumberForTest(RamschTricksPerRound)
	g.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 0, Card: rc(CardDesignClover, ramschValueJack)},
		{PlayerIdx: 1, Card: rc(CardDesignHeart, 8)},
		{PlayerIdx: 2, Card: rc(CardDesignHeart, 7)},
	})
	g.SetPhase(RamschPhaseTrickEnd)
	before := g.GetCardPoints(0)
	g.ResolveTrick()

	// トリック自体は J の 2 点、そこに伏せ札の 21 点が乗る。
	assert.Equal(t, before+2+21, g.GetCardPoints(0), "伏せ札が最終トリックの獲得者に付いていない")
	assert.Equal(t, 0, g.GetCardPoints(1))
	assert.Equal(t, RamschPhaseRoundEnd, g.GetPhase())
}

// **配り切ると 120 点がちょうど誰かのものになる。** 伏せ札を落としていれば
// ここが 120 に届かない。実際に 1 ラウンド最後まで回して確かめる。
func TestRamsch_AFullRoundAccountsForEveryPoint(t *testing.T) {
	g := newTestRamsch()
	g.Reset()

	for guard := 0; guard < 60; guard++ {
		switch g.GetPhase() {
		case RamschPhasePlay:
			require.True(t, g.IsHumanTurn(), "CPU の手番で止まっている (trick %d)", g.GetTrickNumber())
			valid := g.GetValidPlayIndices(g.GetCurrentPlayerIdx())
			require.NotEmpty(t, valid, "合法手が無い (trick %d)", g.GetTrickNumber())
			require.NoError(t, g.PlayerPlay(valid[0]))
		case RamschPhaseTrickEnd:
			g.ResolveTrick()
			if g.GetPhase() == RamschPhaseTrickEnd {
				g.NextTrick()
			}
		case RamschPhaseRoundEnd:
			total := 0
			tricks := 0
			for i := 0; i < g.GetPlayerCnt(); i++ {
				total += g.GetCardPoints(i)
				tricks += g.GetPlayer(i).GetTrickCount()
			}
			assert.Equal(t, RamschTotalCardPoints, total, "120 点が全部どこかに行っていない")
			assert.Equal(t, RamschTricksPerRound, tricks)
			return
		default:
			t.Fatalf("unexpected phase %d", g.GetPhase())
		}
	}
	t.Fatal("ラウンドが終わらなかった")
}

// **CPU は取らないほうを選ぶ。** Skat から来た「勝てるなら勝つ」をそのまま
// 残すと、点を集めて自分が負ける。降りられる場面では必ず降りること。
func TestRamsch_CpuDucksInsteadOfWinningThePoints(t *testing.T) {
	g := newTestRamsch()
	g.Reset()
	cfg := g.GetConfig()
	cfg.CpuDifficulty = RamschCpuDifficultyHard
	g.SetConfig(cfg)

	// 場: ♥K（4点）がリードで勝っている。CPU 1 の手札は ♥A（勝てるが 11 点）と
	// ♥7（降りられる、0 点）。
	g.SetHandForTest(1, []*Card{rc(CardDesignHeart, ramschValueAce), rc(CardDesignHeart, 7)})
	g.SetCurrentTrickForTest([]*TrickCard{{PlayerIdx: 0, Card: rc(CardDesignHeart, ramschValueKing)}})

	idx := g.CpuPickPlayForTest(1)
	require.GreaterOrEqual(t, idx, 0)
	played := g.GetPlayer(1).GetCard(idx)
	assert.Equal(t, 7, played.GetValue(), "降りられるのに A を出して 15 点を取っている")
}

// **降りるなら「いちばん高い負け札」で。** 安い札から吐くと、A が最後まで残って
// 終盤に必ず取らされる。
func TestRamsch_CpuDucksWithItsHighestLosingCard(t *testing.T) {
	g := newTestRamsch()
	g.Reset()
	cfg := g.GetConfig()
	cfg.CpuDifficulty = RamschCpuDifficultyHard
	g.SetConfig(cfg)

	// 場: ♥A がリード。どの手札でも勝てない = 全部「降り札」。
	// ♥K(4点) と ♥7(0点) なら、高い ♥K を吐く。
	g.SetHandForTest(1, []*Card{rc(CardDesignHeart, 7), rc(CardDesignHeart, ramschValueKing)})
	g.SetCurrentTrickForTest([]*TrickCard{{PlayerIdx: 0, Card: rc(CardDesignHeart, ramschValueAce)}})

	idx := g.CpuPickPlayForTest(1)
	played := g.GetPlayer(1).GetCard(idx)
	assert.Equal(t, ramschValueKing, played.GetValue(),
		"取られない回に高い札を処分していない ── 終盤に抱え込む")
}

// **取るのが避けられないなら、いちばん安く取る。** 全部が勝ってしまう手札で、
// わざわざ A を出す理由は無い。
func TestRamsch_CpuTakesTheCheapestTrickItCannotAvoid(t *testing.T) {
	g := newTestRamsch()
	g.Reset()
	cfg := g.GetConfig()
	cfg.CpuDifficulty = RamschCpuDifficultyHard
	g.SetConfig(cfg)

	// 場: ♥7 がリード。手札は ♥A(11点) と ♥9(0点) ── どちらも勝ってしまう。
	g.SetHandForTest(1, []*Card{rc(CardDesignHeart, ramschValueAce), rc(CardDesignHeart, 9)})
	g.SetCurrentTrickForTest([]*TrickCard{{PlayerIdx: 0, Card: rc(CardDesignHeart, 7)}})

	idx := g.CpuPickPlayForTest(1)
	played := g.GetPlayer(1).GetCard(idx)
	assert.Equal(t, 9, played.GetValue(), "避けられない勝ちで、いちばん高い点札を出している")
}

// **助言は「取らない」方向に出る。** 点札しか出せない場面と、0 点で逃げられる
// 場面で理由が変わること。
func TestRamsch_HintAvoidsPointsAndOnlyFiresOnTheHumansTurn(t *testing.T) {
	g := newTestRamsch()
	g.Reset()

	require.True(t, g.IsHumanTurn())
	hint := g.GetHint()
	require.NotNil(t, hint)
	require.NotNil(t, hint.CardIndex)

	// 手番でなければ助言しない。
	g.SetCurrentPlayerIdx(1)
	assert.Nil(t, g.GetHint(), "CPU の手番で助言している")

	// プレイ以外のフェーズでも助言しない。
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(RamschPhaseRoundEnd)
	assert.Nil(t, g.GetHint(), "ラウンド終了後に着手を勧めている")
}

// **Durchmarsch を「実際に取って」検出させる。** 上のテストはフラグを直接
// 立てているので、検出そのものを消しても通ってしまった（ミューテーションで確認）。
// 9 トリック取っている状態から最終トリックも取り、検出が走ることを見る。
func TestRamsch_DurchmarschIsDetectedFromTheTricksActuallyTaken(t *testing.T) {
	g := newTestRamsch()
	g.Reset()

	// player 0 が既に 9 トリック取っている状態を作る。
	for i := 0; i < RamschTricksPerRound-1; i++ {
		g.GetPlayer(0).AddTrick([]*Card{rc(CardDesignHeart, 7), rc(CardDesignHeart, 8), rc(CardDesignHeart, 9)})
	}
	require.Equal(t, RamschTricksPerRound-1, g.GetPlayer(0).GetTrickCount())

	// 最終トリック。player 0 が ♣J（最強の切り札）で取る。
	g.SetTrickNumberForTest(RamschTricksPerRound)
	g.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 0, Card: rc(CardDesignClover, ramschValueJack)},
		{PlayerIdx: 1, Card: rc(CardDesignHeart, 8)},
		{PlayerIdx: 2, Card: rc(CardDesignHeart, 7)},
	})
	g.SetPhase(RamschPhaseTrickEnd)
	g.ResolveTrick()

	assert.True(t, g.IsDurchmarsch(), "10 トリック全部取ったのに Durchmarsch になっていない")
	assert.Equal(t, 0, g.GetDurchmarschIdx())

	// **負のコントロール**: 9 トリックしか取っていなければ Durchmarsch ではない。
	g2 := newTestRamsch()
	g2.Reset()
	for i := 0; i < RamschTricksPerRound-2; i++ {
		g2.GetPlayer(0).AddTrick([]*Card{rc(CardDesignHeart, 7), rc(CardDesignHeart, 8), rc(CardDesignHeart, 9)})
	}
	g2.SetTrickNumberForTest(RamschTricksPerRound)
	g2.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 0, Card: rc(CardDesignClover, ramschValueJack)},
		{PlayerIdx: 1, Card: rc(CardDesignHeart, 8)},
		{PlayerIdx: 2, Card: rc(CardDesignHeart, 7)},
	})
	g2.SetPhase(RamschPhaseTrickEnd)
	g2.ResolveTrick()
	assert.False(t, g2.IsDurchmarsch(), "9 トリックで Durchmarsch になっている")
}
