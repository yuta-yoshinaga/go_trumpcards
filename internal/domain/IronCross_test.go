//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newIronCrossForTest(t *testing.T) *IronCross {
	t.Helper()
	g := NewDefaultIronCross()
	g.Reset()
	return g
}

func icCard(design, value int) *Card { return NewCard(design, value, false) }

// icTotalChips は卓のチップ総量 (席 + ポット) を返す。
func icTotalChips(g *IronCross) int {
	total := g.GetPot()
	for _, p := range g.GetPlayers() {
		total += p.GetChips()
	}
	return total
}

// icPlayToChoose は選択の場面まで進める。
func icPlayToChoose(t *testing.T, g *IronCross) bool {
	t.Helper()
	for steps := 0; g.GetPhase() == IronCrossPhaseBetting; steps++ {
		require.Less(t, steps, 200, "ハンドが終わらない")
		if err := g.PlayerAction(IronCrossActionCheck, 0); err != nil {
			require.NoError(t, g.PlayerAction(IronCrossActionCall, 0))
		}
	}
	return g.GetPhase() == IronCrossPhaseChoose
}

// --- 十字の配置 ---

// **並びが配置そのもの。** 位置の番号を入れ替えると縦横の中身が変わる。
func TestIronCross_LineIndexes(t *testing.T) {
	t.Parallel()
	assert.Equal(t, []int{IronCrossTop, IronCrossCenter, IronCrossBottom},
		IronCrossLineIndexes(IronCrossLineVertical))
	assert.Equal(t, []int{IronCrossLeft, IronCrossCenter, IronCrossRight},
		IronCrossLineIndexes(IronCrossLineHorizontal))
	assert.Nil(t, IronCrossLineIndexes(IronCrossLineNone))

	// **中央だけが両方に入る。** ここが選択の妙。
	v := IronCrossLineIndexes(IronCrossLineVertical)
	h := IronCrossLineIndexes(IronCrossLineHorizontal)
	shared := 0
	for _, a := range v {
		for _, b := range h {
			if a == b {
				shared++
			}
		}
	}
	assert.Equal(t, 1, shared, "縦と横が共有する位置は中央 1 つだけ")
}

func TestIronCross_LineNames(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "vertical", IronCrossLineName(IronCrossLineVertical))
	assert.Equal(t, "horizontal", IronCrossLineName(IronCrossLineHorizontal))
	assert.Equal(t, "none", IronCrossLineName(IronCrossLineNone))
	assert.Equal(t, "none", IronCrossLineName(IronCrossLine(99)))
}

// **中央は最後に開く。** どちらを選ぶかが最後の 1 枚まで決まらない。
func TestIronCross_RevealsTheCentreLast(t *testing.T) {
	t.Parallel()
	g := newIronCrossForTest(t)
	seen := make([]int, 0, IronCrossCommunityCards)
	for i := range IronCrossCommunityCards {
		g.revealed = i
		seen = append(seen, g.revealOrder())
	}
	assert.Equal(t, IronCrossCenter, seen[len(seen)-1], "中央が最後に開いていない")
	assert.Len(t, seen, IronCrossCommunityCards)
	// 5 か所すべてを 1 回ずつ開く。
	assert.ElementsMatch(t,
		[]int{IronCrossCenter, IronCrossTop, IronCrossBottom, IronCrossLeft, IronCrossRight}, seen)
}

// --- 列の選択が結果を変える ---

// **選んだ列の 3 枚だけが使える。** 十字 5 枚から自由に選べる実装だと、
// このゲームの唯一の判断が消える。
func TestIronCross_OnlyTheChosenLineIsAvailable(t *testing.T) {
	t.Parallel()
	p := NewIronCrossPlayer("YOU", 1000, true)
	for _, c := range []*Card{
		icCard(CardDesignHeart, 2), icCard(CardDesignClover, 3),
		icCard(CardDesignDiamond, 4), icCard(CardDesignHeart, 6),
	} {
		p.AddCard(c)
	}
	// 縦 (上・中央・下) にスペードの A K Q、横 (左・中央・右) はばらばら。
	cross := make([]*Card, IronCrossCommunityCards)
	cross[IronCrossTop] = icCard(CardDesignSpade, 1)
	cross[IronCrossCenter] = icCard(CardDesignSpade, 13)
	cross[IronCrossBottom] = icCard(CardDesignSpade, 12)
	cross[IronCrossLeft] = icCard(CardDesignHeart, 8)
	cross[IronCrossRight] = icCard(CardDesignClover, 9)

	vRank, vBest := p.EvaluateLine(cross, IronCrossLineVertical)
	hRank, hBest := p.EvaluateLine(cross, IronCrossLineHorizontal)
	require.Len(t, vBest, IronCrossHandSize)
	require.Len(t, hBest, IronCrossHandSize)

	// **縦の札が横の評価に混ざっていないこと。**
	for _, c := range hBest {
		assert.NotEqual(t, 1, c.GetValue(), "縦にしかない A が横の手に入っている")
		assert.NotEqual(t, 12, c.GetValue(), "縦にしかない Q が横の手に入っている")
	}
	assert.GreaterOrEqual(t, vRank, PokerHandHighCard)
	assert.GreaterOrEqual(t, hRank, PokerHandHighCard)
}

// **強いほうの列が自動で選ばれる (CPU と未選択時)。**
func TestIronCross_EvaluateBestPicksTheStrongerLine(t *testing.T) {
	t.Parallel()
	p := NewIronCrossPlayer("YOU", 1000, true)
	for _, c := range []*Card{
		icCard(CardDesignSpade, 10), icCard(CardDesignSpade, 11),
		icCard(CardDesignHeart, 2), icCard(CardDesignClover, 3),
	} {
		p.AddCard(c)
	}
	// 縦にスペードを並べてフラッシュを作れるようにする。
	cross := make([]*Card, IronCrossCommunityCards)
	cross[IronCrossTop] = icCard(CardDesignSpade, 5)
	cross[IronCrossCenter] = icCard(CardDesignSpade, 7)
	cross[IronCrossBottom] = icCard(CardDesignSpade, 9)
	cross[IronCrossLeft] = icCard(CardDesignHeart, 4)
	cross[IronCrossRight] = icCard(CardDesignDiamond, 6)

	rank := p.EvaluateBest(cross)
	assert.Equal(t, IronCrossLineVertical, p.GetLine(), "強いほうの列を選んでいない")
	assert.GreaterOrEqual(t, rank, PokerHandFlush)
	for _, c := range p.GetBestHand() {
		assert.Equal(t, CardDesignSpade, c.GetDesign())
	}
}

// **人間が選んだ列は尊重する。** 弱いほうを選んでも勝手に直さない。
func TestIronCross_RespectsAnExplicitChoice(t *testing.T) {
	t.Parallel()
	p := NewIronCrossPlayer("YOU", 1000, true)
	for _, c := range []*Card{
		icCard(CardDesignSpade, 10), icCard(CardDesignSpade, 11),
		icCard(CardDesignHeart, 2), icCard(CardDesignClover, 3),
	} {
		p.AddCard(c)
	}
	cross := make([]*Card, IronCrossCommunityCards)
	cross[IronCrossTop] = icCard(CardDesignSpade, 5)
	cross[IronCrossCenter] = icCard(CardDesignSpade, 7)
	cross[IronCrossBottom] = icCard(CardDesignSpade, 9)
	cross[IronCrossLeft] = icCard(CardDesignHeart, 4)
	cross[IronCrossRight] = icCard(CardDesignDiamond, 6)

	p.SetLine(IronCrossLineHorizontal) // わざと弱いほう
	p.EvaluateBest(cross)
	assert.Equal(t, IronCrossLineHorizontal, p.GetLine(), "選んだ列が上書きされている")
	assert.Less(t, p.GetHandRank(), PokerHandFlush, "縦の札が混ざっている")
}

// --- 進行 ---

func TestIronCross_DealsFourEach(t *testing.T) {
	t.Parallel()
	g := newIronCrossForTest(t)
	assert.Equal(t, IronCrossPhaseBetting, g.GetPhase())
	for i, p := range g.GetPlayers() {
		assert.Equal(t, IronCrossHoleCards, p.GetCardsSize(), "席 %d の手札が 4 枚でない", i)
		assert.Equal(t, IronCrossLineNone, p.GetLine(), "配った直後に列が選ばれている")
	}
	assert.Zero(t, g.GetRevealedCount())
	assert.Positive(t, g.GetPot(), "アンティが集まっていない")
	// 十字は 5 か所ぶんの枠があり、まだ全部 nil。
	require.Len(t, g.GetCross(), IronCrossCommunityCards)
	for i, c := range g.GetCross() {
		assert.Nil(t, c, "位置 %d が配る前から開いている", i)
	}
}

// **5 枚そろってから選ぶ。** 全部見えている状態で選ばせる。
func TestIronCross_ChoosesAfterAllFiveAreUp(t *testing.T) {
	t.Parallel()
	for range 30 {
		g := newIronCrossForTest(t)
		if !icPlayToChoose(t, g) {
			continue // 全員降りて決着した
		}
		assert.Equal(t, IronCrossCommunityCards, g.GetRevealedCount(),
			"5 枚そろう前に選択の場面になっている")
		for i, c := range g.GetCross() {
			assert.NotNil(t, c, "位置 %d が伏せたまま選択の場面になっている", i)
		}
		assert.True(t, g.IsChoosing())
		return
	}
	t.Fatalf("30 回配っても選択の場面に届かなかった")
}

func TestIronCross_ChooseLineSettlesTheHand(t *testing.T) {
	t.Parallel()
	for range 30 {
		g := newIronCrossForTest(t)
		if !icPlayToChoose(t, g) {
			continue
		}
		require.NoError(t, g.ChooseLine(IronCrossLineVertical))
		assert.Equal(t, IronCrossPhaseShowdown, g.GetPhase())
		assert.Equal(t, IronCrossLineVertical, g.GetPlayers()[g.HumanSeat()].GetLine())
		assert.Len(t, g.GetResults(), len(g.GetPlayers()))
		assert.Zero(t, g.GetPot(), "ポットが残っている")
		return
	}
	t.Fatalf("30 回配っても選択の場面に届かなかった")
}

func TestIronCross_ChooseLineValidation(t *testing.T) {
	t.Parallel()
	g := newIronCrossForTest(t)
	// ベット中は選べない。
	assert.ErrorIs(t, g.ChooseLine(IronCrossLineVertical), errIronCrossWrongPhase)

	for range 30 {
		g = newIronCrossForTest(t)
		if !icPlayToChoose(t, g) {
			continue
		}
		assert.ErrorIs(t, g.ChooseLine(IronCrossLineNone), errIronCrossBadLine)
		assert.ErrorIs(t, g.ChooseLine(IronCrossLine(99)), errIronCrossBadLine)
		return
	}
	t.Fatalf("30 回配っても選択の場面に届かなかった")
}

// **チップは湧かないし消えない。**
func TestIronCross_ChipsAreConserved(t *testing.T) {
	t.Parallel()
	for range 30 {
		g := newIronCrossForTest(t)
		want := icTotalChips(g)

		for hand := 0; hand < 3 && !g.GetGameEndFlag(); hand++ {
			for steps := 0; g.GetPhase() == IronCrossPhaseBetting; steps++ {
				require.Less(t, steps, 200)
				assert.Equal(t, want, icTotalChips(g), "ベット中に総量が変わっている")
				if err := g.PlayerAction(IronCrossActionCheck, 0); err != nil {
					require.NoError(t, g.PlayerAction(IronCrossActionCall, 0))
				}
			}
			if g.IsChoosing() {
				require.NoError(t, g.ChooseLine(IronCrossLineVertical))
			}
			assert.Equal(t, want, icTotalChips(g), "決着後に総量が変わっている")
			assert.Zero(t, g.GetPot(), "ポットが残っている")
			if g.GetGameEndFlag() {
				break
			}
			require.NoError(t, g.NextHand())
		}
	}
}

// --- 入力の検証 ---

func TestIronCross_ActionValidation(t *testing.T) {
	t.Parallel()
	g := newIronCrossForTest(t)
	assert.ErrorIs(t, g.PlayerAction(9999, 0), errIronCrossBadAction)
	assert.ErrorIs(t, g.PlayerAction(IronCrossActionRaise, 10), errIronCrossCannotRaise)
	assert.ErrorIs(t, g.PlayerAction(IronCrossActionBet, 1), errIronCrossBetRange)
	assert.ErrorIs(t, g.PlayerAction(IronCrossActionBet, 999999), errIronCrossBetRange)
}

func TestIronCross_PhaseGuards(t *testing.T) {
	t.Parallel()
	g := newIronCrossForTest(t)
	assert.ErrorIs(t, g.NextHand(), errIronCrossWrongPhase)

	g.gameEndFlag = true
	assert.ErrorIs(t, g.PlayerAction(IronCrossActionCheck, 0), errIronCrossFinished)
	assert.ErrorIs(t, g.ChooseLine(IronCrossLineVertical), errIronCrossFinished)
	assert.ErrorIs(t, g.NextHand(), errIronCrossFinished)
}

func TestIronCross_NotYourTurn(t *testing.T) {
	t.Parallel()
	g := newIronCrossForTest(t)
	g.turn = (g.HumanSeat() + 1) % len(g.GetPlayers())
	assert.ErrorIs(t, g.PlayerAction(IronCrossActionCheck, 0), errIronCrossNotYourRun)
}

func TestIronCross_RaiseIsCapped(t *testing.T) {
	t.Parallel()
	g := newIronCrossForTest(t)
	g.currentBet = 10
	g.raiseCount = ironCrossMaxRaisesPerRound
	g.players[g.HumanSeat()].SetCurrentBet(0)
	assert.ErrorIs(t, g.PlayerAction(IronCrossActionRaise, 10), errIronCrossRaiseCapped)
	assert.False(t, g.CanRaise())
}

func TestIronCross_ConfigValidate(t *testing.T) {
	t.Parallel()
	assert.NoError(t, DefaultIronCrossConfig().Validate())
	assert.NoError(t, IronCrossConfig{Seats: 7, InitialChips: 1000, Ante: 10}.Validate())
	assert.ErrorIs(t, IronCrossConfig{Seats: 2, InitialChips: 1000, Ante: 10}.Validate(), errIronCrossSeatsRange)
	assert.ErrorIs(t, IronCrossConfig{Seats: 8, InitialChips: 1000, Ante: 10}.Validate(), errIronCrossSeatsRange)
	assert.ErrorIs(t, IronCrossConfig{Seats: 4, InitialChips: 1, Ante: 10}.Validate(), errIronCrossChipsRange)
	assert.ErrorIs(t, IronCrossConfig{Seats: 4, InitialChips: 1000, Ante: 1}.Validate(), errIronCrossAnteRange)
}

func TestIronCross_Accessors(t *testing.T) {
	t.Parallel()
	g := newIronCrossForTest(t)
	assert.Equal(t, 1, g.GetHandNumber())
	assert.Positive(t, g.GetRemainingCards())
	assert.NotEmpty(t, g.GetActionLog())
	assert.Zero(t, g.HumanSeat())
	assert.GreaterOrEqual(t, g.GetToCall(), 0)
	assert.GreaterOrEqual(t, g.GetCurrentBet(), 0)
	assert.GreaterOrEqual(t, g.GetRaiseCount(), 0)
	assert.Equal(t, "YOU", g.GetPlayers()[0].GetName())
	assert.True(t, g.GetPlayers()[0].GetIsHuman())
	assert.False(t, g.IsChoosing())

	g.SetConfig(IronCrossConfig{Seats: 3, InitialChips: 500, Ante: 5})
	assert.Equal(t, 500, g.GetConfig().InitialChips)

	g.players[2].SetChips(99999)
	assert.Equal(t, 2, g.WinnerSeat())
}

// **CpuPlay は人間の手番まで進める。**
func TestIronCross_CpuPlayAdvancesToTheHuman(t *testing.T) {
	t.Parallel()
	g := newIronCrossForTest(t)
	g.turn = (g.HumanSeat() + 1) % len(g.GetPlayers())
	g.CpuPlay()
	assert.True(t, g.IsHumanTurn() || g.GetPhase() != IronCrossPhaseBetting,
		"CpuPlay を呼んだのに CPU の席 %d で止まっている", g.GetTurnSeat())
}

// --- 助言 ---

// **選ぶ場面では強いほうの列を名指しする。** ここが唯一取り返しのつかない選択。
func TestIronCross_HintNamesTheBetterLine(t *testing.T) {
	t.Parallel()
	for range 30 {
		g := newIronCrossForTest(t)
		if !icPlayToChoose(t, g) {
			continue
		}
		h := g.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "line", h.Action)
		assert.Contains(t, []IronCrossLine{IronCrossLineVertical, IronCrossLineHorizontal}, h.Line)

		// 名指しした列が、実際に弱くないこと。
		p := g.GetPlayers()[g.HumanSeat()]
		named, _ := p.EvaluateLine(g.GetCross(), h.Line)
		other := IronCrossLineVertical
		if h.Line == IronCrossLineVertical {
			other = IronCrossLineHorizontal
		}
		otherRank, _ := p.EvaluateLine(g.GetCross(), other)
		assert.GreaterOrEqual(t, named, otherRank, "弱いほうの列を薦めている")
		return
	}
	t.Fatalf("30 回配っても選択の場面に届かなかった")
}

func TestIronCross_HintDuringBetting(t *testing.T) {
	t.Parallel()
	g := newIronCrossForTest(t)
	h := g.GetHint()
	require.NotNil(t, h)
	assert.Contains(t, []string{"fold", "check", "call", "bet", "raise"}, h.Action)

	g.gameEndFlag = true
	assert.Nil(t, g.GetHint())
}
