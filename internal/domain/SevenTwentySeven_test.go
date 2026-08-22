//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestS27() *SevenTwentySeven { return NewDefaultSevenTwentySeven() }

// 配ったらすぐ引くフェーズ。手札は 2 枚から始まる。
func TestSevenTwentySeven_ResetDealsTwoCardsAndOpensTheDrawPhase(t *testing.T) {
	g := newTestS27()
	assert.Equal(t, SevenTwentySevenPhaseDraw, g.GetPhase())
	assert.Equal(t, 1, g.GetDrawRound())
	for i := 0; i < g.GetPlayerCnt(); i++ {
		assert.Equal(t, SevenTwentySevenHandSize, g.GetPlayer(i).GetCardsSize(), "player %d", i)
	}
	assert.Positive(t, g.GetPot(), "アンティがポットに入っていない")
}

// **止まった人には二度と配られない。** ここが効いていないと「止まる」判断に
// 意味が無くなり、超過が事故でしか起きなくなる。
func TestSevenTwentySeven_StandingPatStopsTheCards(t *testing.T) {
	g := newTestS27()
	g.SetStandingForTest(1, true)
	before := g.GetPlayer(1).GetCardsSize()
	beforeHuman := g.GetPlayer(0).GetCardsSize()

	require.NoError(t, g.TakeCard(true)) // 人間は引く

	assert.Equal(t, before, g.GetPlayer(1).GetCardsSize(), "止まった席に配られている")
	assert.Greater(t, g.GetPlayer(0).GetCardsSize(), beforeHuman, "引くと言った席に配られていない")
}

// 一度止まったら、そのラウンドはもう宣言できない。
func TestSevenTwentySeven_CannotActAfterStanding(t *testing.T) {
	g := newTestS27()
	g.SetStandingForTest(0, true)
	assert.Error(t, g.TakeCard(true))
}

// **人間が止まっても盤は進む。** 待ってしまうと、人間に打つ手が無いのに
// 全員が止まるまで進まず、ゲームが固まる。
func TestSevenTwentySeven_TheRoundFinishesItselfOnceTheHumanStands(t *testing.T) {
	g := newTestS27()
	require.NoError(t, g.TakeCard(false)) // 止まる
	assert.Equal(t, SevenTwentySevenPhaseResult, g.GetPhase(),
		"人間が止まったのにラウンドが決着していない")
}

// **7 側と 27 側で別々に勝者が出る。** 片方だけの実装では、ポットが割れない。
func TestSevenTwentySeven_BothSidesGetAWinnerAndSplitThePot(t *testing.T) {
	g := newTestS27()
	// p0: K Q K = 1.5 点 → 7 側に最も近い（27 側では遠い）
	g.SetHandForTest(0, []*Card{s27(CardDesignSpade, 13), s27(CardDesignHeart, 12), s27(CardDesignClover, 13)})
	// p1: 6 → 6 点。7 側でこちらが勝つ。
	g.SetHandForTest(1, []*Card{s27(CardDesignSpade, 6)})
	// p2: 10 9 8 = 27 ちょうど。27 側の勝者。
	g.SetHandForTest(2, []*Card{s27(CardDesignSpade, 10), s27(CardDesignHeart, 9), s27(CardDesignClover, 8)})
	// p3: 10 10 = 20。27 側では p2 に負ける。
	g.SetHandForTest(3, []*Card{s27(CardDesignDiamond, 10), s27(CardDesignSpade, 10)})

	before := []int{g.GetPlayer(0).GetChips(), g.GetPlayer(1).GetChips(),
		g.GetPlayer(2).GetChips(), g.GetPlayer(3).GetChips()}
	g.SetPotForTest(100)
	g.StandEveryoneForTest()
	g.SettleForTest()

	assert.Equal(t, 1, g.GetLowWinner(), "6 点の席が 7 側を取っていない")
	assert.Equal(t, 2, g.GetHighWinner(), "27 ちょうどの席が 27 側を取っていない")
	assert.Equal(t, before[1]+50, g.GetPlayer(1).GetChips(), "7 側が半分を受け取っていない")
	assert.Equal(t, before[2]+50, g.GetPlayer(2).GetChips(), "27 側が半分を受け取っていない")
	assert.Equal(t, before[0], g.GetPlayer(0).GetChips(), "どちらも取っていない席が貰っている")
	assert.Equal(t, before[3], g.GetPlayer(3).GetChips())
}

// **両取り (スクープ) は総取り。** A-A-5 は 7 にも 27 にもできる、このゲームの華。
func TestSevenTwentySeven_ScoopingBothSidesTakesTheWholePot(t *testing.T) {
	g := newTestS27()
	g.SetHandForTest(0, []*Card{s27(CardDesignSpade, 1), s27(CardDesignHeart, 1), s27(CardDesignClover, 5)})
	// 他は両側とも負ける手。
	for i := 1; i < g.GetPlayerCnt(); i++ {
		g.SetHandForTest(i, []*Card{s27(CardDesignSpade, 10), s27(CardDesignHeart, 8)}) // 18 点
	}
	before := g.GetPlayer(0).GetChips()
	g.SetPotForTest(100)
	g.StandEveryoneForTest()
	g.SettleForTest()

	assert.Equal(t, 0, g.GetLowWinner())
	assert.Equal(t, 0, g.GetHighWinner())
	assert.Equal(t, before+100, g.GetPlayer(0).GetChips(), "両取りなのに総取りになっていない")
	assert.Equal(t, SevenTwentySevenResultWin, g.GetResult())
}

// **片側が全滅したら、もう片方が総取り。** 半分を宙に浮かせない。
func TestSevenTwentySeven_AnEmptySideGivesTheWholePotToTheOther(t *testing.T) {
	g := newTestS27()
	// 全員 7 超え。27 側だけ生存。
	for i := 0; i < g.GetPlayerCnt(); i++ {
		g.SetHandForTest(i, []*Card{s27(CardDesignSpade, 9), s27(CardDesignHeart, 2)}) // 11 点
	}
	g.SetHandForTest(2, []*Card{s27(CardDesignSpade, 10), s27(CardDesignHeart, 9), s27(CardDesignClover, 7)}) // 26
	before := g.GetPlayer(2).GetChips()
	g.SetPotForTest(100)
	g.StandEveryoneForTest()
	g.SettleForTest()

	assert.Equal(t, -1, g.GetLowWinner(), "7 側に生存者がいることになっている")
	assert.Equal(t, 2, g.GetHighWinner())
	assert.Equal(t, before+100, g.GetPlayer(2).GetChips(), "片側全滅で総取りになっていない")
}

// **両側とも全滅ならポットは持ち越す。** チップが消えない。
func TestSevenTwentySeven_EverybodyBustingCarriesThePotOver(t *testing.T) {
	g := newTestS27()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		g.SetHandForTest(i, []*Card{s27(CardDesignSpade, 10), s27(CardDesignHeart, 10), s27(CardDesignClover, 10)})
	}
	chipsBefore := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		chipsBefore += g.GetPlayer(i).GetChips()
	}
	g.SetPotForTest(100)
	g.StandEveryoneForTest()
	g.SettleForTest()

	assert.Equal(t, -1, g.GetLowWinner())
	assert.Equal(t, -1, g.GetHighWinner())
	assert.Equal(t, 100, g.GetCarryPot(), "ポットが持ち越されていない")
	chipsAfter := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		chipsAfter += g.GetPlayer(i).GetChips()
	}
	assert.Equal(t, chipsBefore, chipsAfter, "誰も取っていないのにチップが動いている")
}

// 同点は分け合う。端数は先頭から 1 ずつ。
func TestSevenTwentySeven_TiesSplitTheSide(t *testing.T) {
	g := newTestS27()
	// p0 と p1 が 27 ちょうどで同点。7 側は全滅させる。
	for i := 0; i < g.GetPlayerCnt(); i++ {
		g.SetHandForTest(i, []*Card{s27(CardDesignSpade, 9), s27(CardDesignHeart, 9)}) // 18
	}
	for _, i := range []int{0, 1} {
		g.SetHandForTest(i, []*Card{s27(CardDesignSpade, 10), s27(CardDesignHeart, 9), s27(CardDesignClover, 8)})
	}
	b0, b1 := g.GetPlayer(0).GetChips(), g.GetPlayer(1).GetChips()
	g.SetPotForTest(100)
	g.StandEveryoneForTest()
	g.SettleForTest()

	assert.Equal(t, b0+50, g.GetPlayer(0).GetChips())
	assert.Equal(t, b1+50, g.GetPlayer(1).GetChips())
}

// **CPU は狙う側を選ぶ。** 常に 27 を追うと 7 側の勝負が消える
// （実測: 常時 27 狙いだと 7 側の生存者が 0 になった）。
func TestSevenTwentySeven_CpuProtectsALowHandInsteadOfChasingTwentySeven(t *testing.T) {
	g := newTestS27()
	// K Q K = 1.5 点。7 側にとても近い。27 側は絶望的に遠い。
	g.SetHandForTest(1, []*Card{s27(CardDesignSpade, 13), s27(CardDesignHeart, 12), s27(CardDesignClover, 13)})
	// 6.5 点 (6 + K) なら 7 側で止まるべき。
	g.SetHandForTest(2, []*Card{s27(CardDesignSpade, 6), s27(CardDesignHeart, 13)})
	assert.False(t, g.CpuDrawsForTest(2), "7 まであと 0.5 なのに引いている")

	// 逆に 20 点なら 27 を追って引く。
	g.SetHandForTest(3, []*Card{s27(CardDesignSpade, 10), s27(CardDesignHeart, 10)})
	assert.True(t, g.CpuDrawsForTest(3), "20 点で 27 を追わずに止まっている")
}

// 目標ちょうどなら必ず止まる。引けば壊すだけ。
func TestSevenTwentySeven_CpuStandsOnAnExactTarget(t *testing.T) {
	g := newTestS27()
	g.SetHandForTest(1, []*Card{s27(CardDesignSpade, 7)}) // 7 ちょうど
	assert.False(t, g.CpuDrawsForTest(1))
	g.SetHandForTest(2, []*Card{s27(CardDesignSpade, 10), s27(CardDesignHeart, 9), s27(CardDesignClover, 8)})
	assert.False(t, g.CpuDrawsForTest(2), "27 ちょうどで引いている")
}

// ヒントは狙っている側を示す。「引け」だけでは助言にならない。
func TestSevenTwentySeven_HintNamesTheSideItIsChasing(t *testing.T) {
	g := newTestS27()

	g.SetHandForTest(0, []*Card{s27(CardDesignSpade, 10), s27(CardDesignHeart, 9), s27(CardDesignClover, 8)})
	h := g.GetHint()
	require.NotNil(t, h)
	assert.False(t, h.Draw)
	assert.Equal(t, "exactly_twentyseven", h.Reason)

	g.SetHandForTest(0, []*Card{s27(CardDesignSpade, 7)})
	h = g.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "exactly_seven", h.Reason)

	g.SetHandForTest(0, []*Card{s27(CardDesignSpade, 10), s27(CardDesignHeart, 10)})
	h = g.GetHint()
	require.NotNil(t, h)
	assert.True(t, h.Draw)
	assert.Equal(t, "chase_twentyseven", h.Reason)

	// 両側とも超えていれば引く意味が無い。
	g.SetHandForTest(0, []*Card{s27(CardDesignSpade, 10), s27(CardDesignHeart, 10), s27(CardDesignClover, 10)})
	h = g.GetHint()
	require.NotNil(t, h)
	assert.False(t, h.Draw)
	assert.Equal(t, "bust_both", h.Reason)

	// 止まっていれば助言しない。
	g.SetStandingForTest(0, true)
	assert.Nil(t, g.GetHint())
}

// **KV 往復。** Worker はリクエストごとに状態を持たないので、ここが欠けると
// 本番だけで壊れる。復元した盤で指し続けられることを見る。
func TestSevenTwentySeven_SurvivesAKVRoundTrip(t *testing.T) {
	g := newTestS27()
	require.NoError(t, g.TakeCard(true))
	require.Equal(t, SevenTwentySevenPhaseDraw, g.GetPhase(), "1 手で決着してしまった")

	data, err := g.MarshalJSON()
	require.NoError(t, err)
	restored := NewDefaultSevenTwentySeven()
	require.NoError(t, restored.UnmarshalJSON(data))

	assert.Equal(t, g.GetPhase(), restored.GetPhase())
	assert.Equal(t, g.GetDrawRound(), restored.GetDrawRound())
	assert.Equal(t, g.GetPot(), restored.GetPot())
	assert.Equal(t, g.GetCarryPot(), restored.GetCarryPot())
	assert.Equal(t, g.GetCarryCount(), restored.GetCarryCount())
	for i := 0; i < g.GetPlayerCnt(); i++ {
		assert.Equal(t, g.GetPlayer(i).GetCardsSize(), restored.GetPlayer(i).GetCardsSize(), "player %d の手札", i)
		assert.Equal(t, g.GetPlayer(i).GetChips(), restored.GetPlayer(i).GetChips(), "player %d のチップ", i)
		assert.Equal(t, g.GetPlayer(i).GetStanding(), restored.GetPlayer(i).GetStanding(), "player %d の止まり", i)
	}
	// 復元した盤で続けられること。
	before := restored.GetPlayer(0).GetCardsSize()
	require.NoError(t, restored.TakeCard(true))
	assert.Greater(t, restored.GetPlayer(0).GetCardsSize(), before, "復元後に引けていない")
}

// **ラウンドをまたいで続くこと。** 持ち越し・アンティ・脱落・終了条件が
// 一巡して初めて「対局」になる。
func TestSevenTwentySeven_PlaysSeveralRoundsAndEnds(t *testing.T) {
	g := newTestS27()
	cfg := g.GetConfig()
	cfg.TargetRounds = 3
	g.SetConfig(cfg)
	g.Reset()

	rounds := 0
	for guard := 0; guard < 200 && !g.GetGameEndFlag(); guard++ {
		switch g.GetPhase() {
		case SevenTwentySevenPhaseDraw:
			require.NoError(t, g.TakeCard(g.CpuDrawsForTest(0)))
		case SevenTwentySevenPhaseResult:
			rounds++
			g.NextRound()
		default:
			t.Fatalf("unexpected phase %d", g.GetPhase())
		}
	}
	assert.True(t, g.GetGameEndFlag(), "%d ラウンド回っても終わらない", rounds)
	// 規定ラウンドに到達して終わったこと。最後の settle で終了するので、
	// NextRound は TargetRounds-1 回しか呼ばれない。
	assert.Equal(t, cfg.TargetRounds, g.GetRoundNumber())
	assert.GreaterOrEqual(t, rounds, cfg.TargetRounds-1)
	// **勝者はチップ最多。** 全員のチップ合計は変わらない（ポットは必ず誰かへ）。
	winner := g.GetMatchWinnerIdx()
	require.GreaterOrEqual(t, winner, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		assert.LessOrEqual(t, g.GetPlayer(i).GetChips(), g.GetPlayer(winner).GetChips(),
			"player %d が勝者よりチップを持っている", i)
	}
}

// **持ち越したポットは次のラウンドの種銭になる。** 消えないこと。
func TestSevenTwentySeven_TheCarriedPotSeedsTheNextRound(t *testing.T) {
	g := newTestS27()
	g.Reset()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		g.SetHandForTest(i, []*Card{rc(CardDesignSpade, 10), rc(CardDesignHeart, 10), rc(CardDesignClover, 10)})
	}
	g.SetPotForTest(100)
	g.StandEveryoneForTest()
	g.SettleForTest()
	require.Equal(t, 100, g.GetCarryPot())

	g.NextRound()
	// 新しいポット = 持ち越し + 全員のアンティ。
	want := 100 + g.GetAnte()*g.GetPlayerCnt()
	assert.Equal(t, want, g.GetPot(), "持ち越しが種銭になっていない")
	assert.Zero(t, g.GetCarryPot(), "持ち越しが二重に残っている")
}

// アンティを払えなくなったプレイヤーは脱落する。
func TestSevenTwentySeven_APlayerWhoCannotAnteDropsOut(t *testing.T) {
	g := newTestS27()
	g.Reset()
	g.GetPlayer(1).SetChips(0)
	g.SetPhase(SevenTwentySevenPhaseResult)
	g.NextRound()
	assert.True(t, g.GetPlayer(1).GetOut(), "アンティを払えないのに残っている")
}

// プレイヤーの JSON 往復。**Worker はここを通るので、落ちると本番だけで壊れる。**
func TestSevenTwentySevenPlayer_JSONRoundTrip(t *testing.T) {
	p := NewSevenTwentySevenPlayer(true, 200)
	p.SetStanding(true)
	p.SetOut(false)
	p.SetRoundBet(35)
	p.AddCard(rc(CardDesignSpade, 4))
	p.AddCard(rc(CardDesignHeart, 13))

	data, err := p.MarshalJSON()
	require.NoError(t, err)

	restored := NewSevenTwentySevenPlayer(false, 0)
	require.NoError(t, restored.UnmarshalJSON(data))
	assert.True(t, restored.GetIsHuman())
	assert.True(t, restored.GetStanding(), "止まりの状態が往復で消えている")
	assert.Equal(t, 200, restored.GetChips())
	assert.Equal(t, 35, restored.GetRoundBet())
	assert.Equal(t, 2, restored.GetCardsSize())
}

// 壊れたスナップショットは弾く（負のチップ・負の賭け）。
func TestSevenTwentySevenPlayer_RejectsInvalidSnapshots(t *testing.T) {
	for _, raw := range []string{`{"ch":-1}`, `{"rb":-5}`, `not json`} {
		p := NewSevenTwentySevenPlayer(false, 0)
		assert.Error(t, p.UnmarshalJSON([]byte(raw)), "raw=%s", raw)
	}
}

// ClearHand は手札だけを空にする（チップは残る）。
func TestSevenTwentySevenPlayer_ClearHand(t *testing.T) {
	p := NewSevenTwentySevenPlayer(true, 200)
	p.AddCard(rc(CardDesignSpade, 4))
	p.SetRoundBet(10)
	p.ClearHand()
	assert.Zero(t, p.GetCardsSize())
	assert.Equal(t, 200, p.GetChips(), "チップまで消えている")
}

// **持ち越し回数も往復すること。** 「carry #3」の表示に使うので、落ちると
// Worker では何度持ち越しても毎回 1 回目として出る。0 == 0 に退化しないよう、
// 実際に 2 回持ち越してから往復させる。
func TestSevenTwentySeven_TheCarryCountSurvivesAKVRoundTrip(t *testing.T) {
	g := newTestS27()
	g.Reset()

	bustEveryone := func() {
		for i := 0; i < g.GetPlayerCnt(); i++ {
			g.SetHandForTest(i, []*Card{rc(CardDesignSpade, 10), rc(CardDesignHeart, 10), rc(CardDesignClover, 10)})
		}
		g.SetPotForTest(100)
		g.StandEveryoneForTest()
		g.SettleForTest()
	}
	bustEveryone()
	g.NextRound()
	bustEveryone()
	require.Equal(t, 2, g.GetCarryCount(), "2 回連続で持ち越していない")

	data, err := g.MarshalJSON()
	require.NoError(t, err)
	restored := NewDefaultSevenTwentySeven()
	require.NoError(t, restored.UnmarshalJSON(data))

	assert.Equal(t, 2, restored.GetCarryCount(), "持ち越し回数が往復で消えている")
	assert.Equal(t, g.GetCarryPot(), restored.GetCarryPot())
}
