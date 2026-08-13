//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCincinnatiForTest(t *testing.T) *Cincinnati {
	t.Helper()
	g := NewDefaultCincinnati()
	g.Reset()
	return g
}

func cinCard(design, value int) *Card { return NewCard(design, value, false) }

// cinTotalChips は卓のチップ総量 (席 + ポット) を返す。
//
// **ポットも数える。** 席だけ足すと、ベット中は必ず減って見える。
func cinTotalChips(g *Cincinnati) int {
	total := g.GetPot()
	for _, p := range g.GetPlayers() {
		total += p.GetChips()
	}
	return total
}

// cinPlayOutHand は人間が降りてハンドを閉じる。
func cinPlayOutHand(t *testing.T, g *Cincinnati) {
	t.Helper()
	for steps := 0; g.GetPhase() == CincinnatiPhaseBetting; steps++ {
		require.Less(t, steps, 200, "ハンドが終わらない")
		require.True(t, g.IsHumanTurn(), "CPU の手番のまま盤面が止まっている (席 %d)", g.GetTurnSeat())
		require.NoError(t, g.PlayerAction(CincinnatiActionFold, 0))
	}
}

// --- 配りと形 ---

func TestCincinnati_DealsFiveEach(t *testing.T) {
	t.Parallel()
	g := newCincinnatiForTest(t)

	assert.Equal(t, CincinnatiPhaseBetting, g.GetPhase())
	for i, p := range g.GetPlayers() {
		assert.Equal(t, CincinnatiHoleCards, p.GetCardsSize(), "席 %d の手札が 5 枚でない", i)
	}
	// **配った直後はコミュニティが 0 枚。** 1 枚目のベットは手札だけで決める。
	assert.Empty(t, g.GetCommunityCards())
	assert.Zero(t, g.GetRevealedCount())
	assert.Positive(t, g.GetPot(), "アンティが集まっていない")
}

// **山が足りない席数を設定で弾く。** 配っている途中で尽きると nil が手札に入る。
func TestCincinnati_ConfigRejectsATableTheDeckCannotServe(t *testing.T) {
	t.Parallel()
	assert.NoError(t, DefaultCincinnatiConfig().Validate())
	// 7 席 = 40 枚でぎりぎり足りる。
	assert.NoError(t, CincinnatiConfig{Seats: 7, InitialChips: 1000, Ante: 10}.Validate())
	assert.ErrorIs(t, CincinnatiConfig{Seats: 8, InitialChips: 1000, Ante: 10}.Validate(),
		errCincinnatiSeatsRange)
	assert.ErrorIs(t, CincinnatiConfig{Seats: 1, InitialChips: 1000, Ante: 10}.Validate(),
		errCincinnatiSeatsRange)
	assert.ErrorIs(t, CincinnatiConfig{Seats: 4, InitialChips: 1, Ante: 10}.Validate(),
		errCincinnatiChipsRange)
	assert.ErrorIs(t, CincinnatiConfig{Seats: 4, InitialChips: 1000, Ante: 1}.Validate(),
		errCincinnatiAnteRange)
}

// --- 役の選択 ---

// **手札 5 枚だけで役が完成しうる。** Holdem と違い、コミュニティを 1 枚も
// 使わない選択が正しい場面がある ── ここを取りこぼすと強い手が弱く出る。
func TestCincinnati_BestHandMayUseNoCommunityCards(t *testing.T) {
	t.Parallel()
	p := NewCincinnatiPlayer("YOU", 1000, true)
	// 手札だけでストレートフラッシュ。
	for _, c := range []*Card{
		cinCard(CardDesignSpade, 5), cinCard(CardDesignSpade, 6), cinCard(CardDesignSpade, 7),
		cinCard(CardDesignSpade, 8), cinCard(CardDesignSpade, 9),
	} {
		p.AddCard(c)
	}
	community := []*Card{
		cinCard(CardDesignHeart, 2), cinCard(CardDesignClover, 3), cinCard(CardDesignDiamond, 4),
		cinCard(CardDesignHeart, 11), cinCard(CardDesignClover, 12),
	}

	rank := p.EvaluateBest(community)
	assert.Equal(t, PokerHandStraightFlush, rank, "手札だけの役を選べていない")
	require.Len(t, p.GetBestHand(), CincinnatiHandSize)
	for _, c := range p.GetBestHand() {
		assert.Equal(t, CardDesignSpade, c.GetDesign(), "コミュニティの札が混ざっている")
	}
}

// **コミュニティだけで役が完成する場合も選べる。** 逆側も踏む。
func TestCincinnati_BestHandMayUseOnlyCommunityCards(t *testing.T) {
	t.Parallel()
	p := NewCincinnatiPlayer("YOU", 1000, true)
	for _, c := range []*Card{
		cinCard(CardDesignHeart, 2), cinCard(CardDesignClover, 3), cinCard(CardDesignDiamond, 4),
		cinCard(CardDesignHeart, 6), cinCard(CardDesignClover, 8),
	} {
		p.AddCard(c)
	}
	community := []*Card{
		cinCard(CardDesignSpade, 10), cinCard(CardDesignSpade, 11), cinCard(CardDesignSpade, 12),
		cinCard(CardDesignSpade, 13), cinCard(CardDesignSpade, 1),
	}

	assert.Equal(t, PokerHandRoyalFlush, p.EvaluateBest(community), "コミュニティだけの役を選べていない")
}

// **10 枚から選ぶので、混ぜた組み合わせも当然選ぶ。**
func TestCincinnati_BestHandMixesBothSources(t *testing.T) {
	t.Parallel()
	p := NewCincinnatiPlayer("YOU", 1000, true)
	for _, c := range []*Card{
		cinCard(CardDesignSpade, 1), cinCard(CardDesignHeart, 1), cinCard(CardDesignClover, 1),
		cinCard(CardDesignDiamond, 2), cinCard(CardDesignHeart, 3),
	} {
		p.AddCard(c)
	}
	community := []*Card{
		cinCard(CardDesignDiamond, 1), cinCard(CardDesignSpade, 5), cinCard(CardDesignHeart, 7),
		cinCard(CardDesignClover, 9), cinCard(CardDesignDiamond, 12),
	}

	// 手札の A 3 枚 + コミュニティの A で four of a kind。
	assert.Equal(t, PokerHandFourOfAKind, p.EvaluateBest(community))
}

func TestCincinnati_EvaluateWithTooFewCards(t *testing.T) {
	t.Parallel()
	p := NewCincinnatiPlayer("YOU", 1000, true)
	p.AddCard(cinCard(CardDesignSpade, 1))
	assert.Equal(t, PokerHandHighCard, p.EvaluateBest(nil))
	assert.Nil(t, p.GetBestHand())
}

// --- 進行 ---

// **コミュニティは 1 枚ずつ 5 回めくる。** Holdem の 3-1-1 とは違う。
func TestCincinnati_RevealsOneCardPerRound(t *testing.T) {
	t.Parallel()
	g := newCincinnatiForTest(t)

	seen := []int{g.GetRevealedCount()}
	for steps := 0; g.GetPhase() == CincinnatiPhaseBetting; steps++ {
		require.Less(t, steps, 200)
		// **CPU が賭けているとチェックできない。** 通す手は場況で変わるので、
		// 「必ず合法な手」を 1 つに決められない ── 拒まれたらコールに落とす。
		if err := g.PlayerAction(CincinnatiActionCheck, 0); err != nil {
			require.NoError(t, g.PlayerAction(CincinnatiActionCall, 0))
		}
		if n := g.GetRevealedCount(); n != seen[len(seen)-1] {
			seen = append(seen, n)
		}
	}
	// 0 →1→2→3→4→5 と 1 枚ずつ増える。
	assert.Equal(t, []int{0, 1, 2, 3, 4, 5}, seen, "1 枚ずつめくれていない")
	assert.Len(t, g.GetCommunityCards(), CincinnatiCommunityCards)
}

// **チェックだけで進めてもショーダウンに着く。** 止まらない規則を書いていない。
func TestCincinnati_AlwaysReachesShowdown(t *testing.T) {
	t.Parallel()
	for range 100 {
		g := newCincinnatiForTest(t)
		for steps := 0; g.GetPhase() == CincinnatiPhaseBetting; steps++ {
			require.Less(t, steps, 200, "ハンドが終わらない")
			require.True(t, g.IsHumanTurn())
			if err := g.PlayerAction(CincinnatiActionCheck, 0); err != nil {
				require.NoError(t, g.PlayerAction(CincinnatiActionCall, 0))
			}
		}
		assert.Equal(t, CincinnatiPhaseShowdown, g.GetPhase())
		assert.Len(t, g.GetResults(), len(g.GetPlayers()))
	}
}

// **チップは湧かないし消えない。** ポットも数に入れて確かめる。
func TestCincinnati_ChipsAreConserved(t *testing.T) {
	t.Parallel()
	for range 50 {
		g := newCincinnatiForTest(t)
		want := cinTotalChips(g)

		for hand := 0; hand < 4 && !g.GetGameEndFlag(); hand++ {
			for steps := 0; g.GetPhase() == CincinnatiPhaseBetting; steps++ {
				require.Less(t, steps, 200)
				assert.Equal(t, want, cinTotalChips(g), "ベット中に総量が変わっている")
				if err := g.PlayerAction(CincinnatiActionCheck, 0); err != nil {
					require.NoError(t, g.PlayerAction(CincinnatiActionCall, 0))
				}
			}
			assert.Equal(t, want, cinTotalChips(g), "決着後に総量が変わっている")
			// **ポットは配り切る。** 端数がポットに残ると卓から消える。
			assert.Zero(t, g.GetPot(), "ポットが残っている")
			if g.GetGameEndFlag() {
				break
			}
			require.NoError(t, g.NextHand())
		}
	}
}

// --- 入力の検証 ---

func TestCincinnati_ActionValidation(t *testing.T) {
	t.Parallel()
	g := newCincinnatiForTest(t)

	assert.ErrorIs(t, g.PlayerAction(9999, 0), errCincinnatiBadAction)
	// 賭けの無いところでレイズはできない。
	assert.ErrorIs(t, g.PlayerAction(CincinnatiActionRaise, 10), errCincinnatiCannotRaise)
	// 額が範囲外のベットは弾く。
	assert.ErrorIs(t, g.PlayerAction(CincinnatiActionBet, 1), errCincinnatiBetRange)
	assert.ErrorIs(t, g.PlayerAction(CincinnatiActionBet, 999999), errCincinnatiBetRange)
}

func TestCincinnati_PhaseGuards(t *testing.T) {
	t.Parallel()
	g := newCincinnatiForTest(t)
	assert.ErrorIs(t, g.NextHand(), errCincinnatiWrongPhase)

	cinPlayOutHand(t, g)
	assert.Equal(t, CincinnatiPhaseShowdown, g.GetPhase())
	assert.ErrorIs(t, g.PlayerAction(CincinnatiActionCheck, 0), errCincinnatiWrongPhase)

	g.gameEndFlag = true
	assert.ErrorIs(t, g.PlayerAction(CincinnatiActionCheck, 0), errCincinnatiFinished)
	assert.ErrorIs(t, g.NextHand(), errCincinnatiFinished)
}

// **他人の手番では動かせない。**
func TestCincinnati_NotYourTurn(t *testing.T) {
	t.Parallel()
	g := newCincinnatiForTest(t)
	g.turn = (g.HumanSeat() + 1) % len(g.GetPlayers())
	assert.ErrorIs(t, g.PlayerAction(CincinnatiActionCheck, 0), errCincinnatiNotYourRun)
}

// **レイズには上限がある。** 無いと 2 人で延々と上げ続けられる。
func TestCincinnati_RaiseIsCapped(t *testing.T) {
	t.Parallel()
	g := newCincinnatiForTest(t)
	g.currentBet = 10
	g.raiseCount = cincinnatiMaxRaisesPerRound
	g.players[g.HumanSeat()].SetCurrentBet(0)
	assert.ErrorIs(t, g.PlayerAction(CincinnatiActionRaise, 10), errCincinnatiRaiseCapped)
	assert.False(t, g.CanRaise())
}

func TestCincinnati_Accessors(t *testing.T) {
	t.Parallel()
	g := newCincinnatiForTest(t)
	assert.Equal(t, 1, g.GetHandNumber())
	assert.Positive(t, g.GetRemainingCards())
	assert.NotEmpty(t, g.GetActionLog())
	assert.Zero(t, g.HumanSeat())
	assert.GreaterOrEqual(t, g.GetToCall(), 0)
	assert.GreaterOrEqual(t, g.GetCurrentBet(), 0)
	assert.GreaterOrEqual(t, g.GetRaiseCount(), 0)
	assert.Equal(t, "YOU", g.GetPlayers()[0].GetName())
	assert.True(t, g.GetPlayers()[0].GetIsHuman())

	g.SetConfig(CincinnatiConfig{Seats: 3, InitialChips: 500, Ante: 5})
	assert.Equal(t, 500, g.GetConfig().InitialChips)
}

// **人間が賭けられなくなったら終わる。**
func TestCincinnati_EndsWhenTheHumanIsBroke(t *testing.T) {
	t.Parallel()
	g := newCincinnatiForTest(t)
	cinPlayOutHand(t, g)
	g.players[g.HumanSeat()].SetChips(0)
	require.NoError(t, g.NextHand())
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, CincinnatiPhaseGameEnd, g.GetPhase())
}

// **同点なら山分けし、端数まで配り切る。**
//
// 端数をポットに残すと、そのぶんが卓から消える。ランダムな配りでは同点が
// めったに出ないので、**総量の検査だけでは端数の取りこぼしを検出できない**
// (端数を配らない変異が素通りした)。同点を作って直接確かめる。
func TestCincinnati_SplitPotDistributesTheRemainder(t *testing.T) {
	t.Parallel()
	g := NewCincinnati(NewTrumpCards(0),
		NewCincinnatiPlayersForTable(3, 1000), CincinnatiConfig{Seats: 3, InitialChips: 1000, Ante: 10})

	// 3 席のうち 2 席が完全に同じ役 (コミュニティのロイヤルフラッシュ)。
	g.community = []*Card{
		cinCard(CardDesignSpade, 10), cinCard(CardDesignSpade, 11), cinCard(CardDesignSpade, 12),
		cinCard(CardDesignSpade, 13), cinCard(CardDesignSpade, 1),
	}
	for i, p := range g.players {
		p.ResetForHand()
		for _, c := range []*Card{
			cinCard(CardDesignHeart, 2), cinCard(CardDesignClover, 3), cinCard(CardDesignDiamond, 4),
			cinCard(CardDesignHeart, 6), cinCard(CardDesignClover, 8),
		} {
			p.AddCard(c)
		}
		if i == 2 {
			p.SetFolded(true) // 2 席だけで分ける
		}
	}
	// **割り切れない額にする。** 端数を落とす実装ならここで消える。
	g.pot = 101
	before := 0
	for _, p := range g.players {
		before += p.GetChips()
	}

	g.showdown()

	after := 0
	won := 0
	for _, p := range g.players {
		after += p.GetChips()
	}
	for _, r := range g.GetResults() {
		won += r.WonAmount
	}
	assert.Equal(t, 101, won, "配当の合計がポットと合わない (端数が消えている)")
	assert.Equal(t, before+101, after, "端数が卓から消えている")
	assert.Zero(t, g.GetPot())
	// 51 と 50 に割れる。
	assert.Equal(t, 51, g.GetResults()[0].WonAmount)
	assert.Equal(t, 50, g.GetResults()[1].WonAmount)
}

// **賭けに応じていない席がいる間は次の札をめくらない。**
//
// 「全員が 1 度動いた」だけで閉じると、レイズを受けた席が応じないまま
// コミュニティが進む。総量は合ったままなので、保存則の検査では出ない
// (ラウンド完了の判定を潰す変異が素通りした)。
func TestCincinnati_DoesNotRevealUntilBetsAreMatched(t *testing.T) {
	t.Parallel()
	g := NewCincinnati(NewTrumpCards(0),
		NewCincinnatiPlayersForTable(3, 1000), CincinnatiConfig{Seats: 3, InitialChips: 1000, Ante: 10})
	g.startHand()
	// 人間の手番に戻す (CPU が先に動いていることがある)。
	g.turn = g.HumanSeat()
	g.currentBet = 0
	g.raiseCount = 0
	g.actedFlags = make([]bool, len(g.players))
	for _, p := range g.players {
		p.SetCurrentBet(0)
		p.SetFolded(false)
	}
	revealedBefore := g.GetRevealedCount()

	// 人間だけがベットした直後は、まだ誰も応じていない。
	require.NoError(t, g.applyAction(g.HumanSeat(), CincinnatiActionBet, 10))
	assert.False(t, g.bettingRoundComplete(),
		"賭けたばかりで誰も応じていないのにラウンドが閉じている")
	assert.Equal(t, revealedBefore, g.GetRevealedCount(), "応じる前に札がめくれている")

	// 1 席が応じてもまだ残っている。
	require.NoError(t, g.applyAction(1, CincinnatiActionCall, 0))
	assert.False(t, g.bettingRoundComplete(), "まだ応じていない席が残っている")

	// 全員が応じて初めて閉じる。
	require.NoError(t, g.applyAction(2, CincinnatiActionCall, 0))
	assert.True(t, g.bettingRoundComplete(), "全員応じたのに閉じない")
}

// **降りた席と全ツッパの席は応じる必要が無い。** 上の規則で止まらないこと。
func TestCincinnati_FoldedAndAllInSeatsDoNotBlockTheRound(t *testing.T) {
	t.Parallel()
	g := NewCincinnati(NewTrumpCards(0),
		NewCincinnatiPlayersForTable(3, 1000), CincinnatiConfig{Seats: 3, InitialChips: 1000, Ante: 10})
	g.startHand()
	g.currentBet = 50
	g.actedFlags = []bool{true, false, false}
	g.players[0].SetCurrentBet(50)
	g.players[1].SetFolded(true)
	g.players[2].SetAllIn(true)

	assert.True(t, g.bettingRoundComplete(),
		"降りた席や全ツッパの席がラウンドを止めている")
}

// **一度動いた席でも、後からレイズを受けたら応じ直す必要がある。**
//
// 「まだ動いていない」か「額が揃っていない」かのどちらか一方だけで判定すると
// ここが漏れる ── 実際、`||` を `&&` に変えた変異は他のテストを全部素通りした。
// 額が揃っていないのに閉じると、レイズに応じていない席を残して札がめくれる。
func TestCincinnati_ActedSeatMustAnswerALaterRaise(t *testing.T) {
	t.Parallel()
	g := NewCincinnati(NewTrumpCards(0),
		NewCincinnatiPlayersForTable(3, 1000), CincinnatiConfig{Seats: 3, InitialChips: 1000, Ante: 10})
	g.startHand()
	g.currentBet = 0
	g.raiseCount = 0
	g.actedFlags = make([]bool, len(g.players))
	for _, p := range g.players {
		p.SetCurrentBet(0)
		p.SetFolded(false)
		p.SetAllIn(false)
	}

	require.NoError(t, g.applyAction(0, CincinnatiActionBet, 10))
	require.NoError(t, g.applyAction(1, CincinnatiActionCall, 0))
	// ここで席 2 がレイズ。席 0 と 1 は「動いた」が額が足りない。
	require.NoError(t, g.applyAction(2, CincinnatiActionRaise, 20))

	require.True(t, g.actedFlags[0], "席 0 は動いている前提")
	require.True(t, g.actedFlags[1], "席 1 は動いている前提")
	require.Less(t, g.players[0].GetCurrentBet(), g.GetCurrentBet(), "席 0 の額が足りていない前提")
	assert.False(t, g.bettingRoundComplete(),
		"レイズに応じていない席が残っているのにラウンドが閉じている")

	require.NoError(t, g.applyAction(0, CincinnatiActionCall, 0))
	assert.False(t, g.bettingRoundComplete(), "まだ席 1 が応じていない")
	require.NoError(t, g.applyAction(1, CincinnatiActionCall, 0))
	assert.True(t, g.bettingRoundComplete(), "全員応じたのに閉じない")
}
