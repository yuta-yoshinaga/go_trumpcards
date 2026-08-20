//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestBaloot(t *testing.T) *Baloot {
	t.Helper()
	b := NewDefaultBaloot()
	b.Reset()
	return b
}

// balootHandOf は指定プレイヤーの手札を固定の並びに差し替える。
func balootHandOf(b *Baloot, idx int, cards ...*Card) {
	p := b.GetPlayer(idx)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// settleMode は宣言フェーズを最後まで進めてモードを確定させる。
func settleMode(t *testing.T, b *Baloot) {
	t.Helper()
	for range BalootPlayerCnt * 2 {
		if b.GetPhase() != BalootPhaseDeclare {
			return
		}
		if b.IsHumanDeclareTurn() {
			require.NoError(t, b.DeclareSun())
			continue
		}
		b.CpuDeclare()
	}
	require.NotEqual(t, BalootPhaseDeclare, b.GetPhase(), "declaration never settled")
}

// --- 配りとデッキ ---

// **宣言の前は 5 枚。** 8 枚 × 4 人 = 32 枚はデッキ全部なので、配り切ってから
// 宣言させると手札を見て決める余地が無くなる。
func TestBaloot_FirstDealIsFiveCards(t *testing.T) {
	b := newTestBaloot(t)

	assert.Equal(t, BalootPhaseDeclare, b.GetPhase())
	for i := range BalootPlayerCnt {
		assert.Equal(t, BalootFirstDealSize, b.GetPlayer(i).GetCardsSize(), "player %d", i)
	}
}

// 宣言が決まると残り 3 枚が配られ、8 枚 × 4 人 = 32 枚でちょうど配り切る。
func TestBaloot_SecondDealCompletesEightCards(t *testing.T) {
	b := newTestBaloot(t)
	settleMode(t, b)

	assert.Equal(t, BalootPhasePlay, b.GetPhase())
	total := 0
	for i := range BalootPlayerCnt {
		assert.Equal(t, BalootHandSize, b.GetPlayer(i).GetCardsSize(), "player %d", i)
		total += b.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, 32, total)
	assert.Equal(t, 0, b.trumpCards.GetRemainingCount(), "the deck must be dealt out entirely")
}

// 32 枚デッキは各スート A,7〜K の 8 枚。
func TestBaloot_DeckIs32(t *testing.T) {
	b := newTestBaloot(t)
	settleMode(t, b)

	bySuit := map[int][]int{}
	for i := range BalootPlayerCnt {
		p := b.GetPlayer(i)
		for j := range p.GetCardsSize() {
			c := p.GetCard(j)
			bySuit[c.GetDesign()] = append(bySuit[c.GetDesign()], c.GetValue())
		}
	}
	require.Len(t, bySuit, 4)
	for suit, values := range bySuit {
		assert.Len(t, values, 8, "suit %d", suit)
		for _, v := range values {
			assert.True(t, v == 1 || (v >= 7 && v <= 13), "suit %d has rank %d", suit, v)
		}
	}
}

// --- 点数表 ---

// Sun の点数。A11/10=10/K4/Q3/J2、9・8・7 は 0。
func TestBaloot_SunCardPoints(t *testing.T) {
	for _, tc := range []struct{ value, want int }{
		{1, 11}, {10, 10}, {13, 4}, {12, 3}, {11, 2}, {9, 0}, {8, 0}, {7, 0},
	} {
		got := BalootCardPoints(NewCard(CardDesignSpade, tc.value, false), BalootModeSun, 0)
		assert.Equal(t, tc.want, got, "rank %d", tc.value)
	}
}

// Hokom の切り札は J20/9=14 が乗る。**同じ札が Sun では 2 点と 0 点。**
func TestBaloot_HokomTrumpCardPoints(t *testing.T) {
	trump := CardDesignHeart
	for _, tc := range []struct{ value, want int }{
		{11, 20}, {9, 14}, {1, 11}, {10, 10}, {13, 4}, {12, 3}, {8, 0}, {7, 0},
	} {
		got := BalootCardPoints(NewCard(trump, tc.value, false), BalootModeHokom, trump)
		assert.Equal(t, tc.want, got, "rank %d", tc.value)
	}
	assert.Equal(t, 2, BalootCardPoints(NewCard(trump, 11, false), BalootModeSun, 0))
	assert.Equal(t, 0, BalootCardPoints(NewCard(trump, 9, false), BalootModeSun, 0))
}

// Hokom でも切り札以外は Sun と同じ点数。
func TestBaloot_HokomNonTrumpUsesSunPoints(t *testing.T) {
	trump := CardDesignHeart
	for _, value := range []int{1, 7, 8, 9, 10, 11, 12, 13} {
		c := NewCard(CardDesignSpade, value, false)
		assert.Equal(t,
			BalootCardPoints(c, BalootModeSun, 0),
			BalootCardPoints(c, BalootModeHokom, trump),
			"rank %d", value)
	}
}

// nil 札は 0 点（復元途中の欠損で落ちないこと）。
func TestBaloot_CardPointsNilIsZero(t *testing.T) {
	assert.Equal(t, 0, BalootCardPoints(nil, BalootModeSun, 0))
	assert.Equal(t, 0, BalootCardPoints(nil, BalootModeHokom, CardDesignHeart))
}

// **デッキ全体の点数はモードで変わる。** Sun 120 / Hokom 152。
func TestBaloot_DeckTotalPointsPerMode(t *testing.T) {
	trump := CardDesignClover
	sun, hokom := 0, 0
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		for _, v := range []int{1, 7, 8, 9, 10, 11, 12, 13} {
			c := NewCard(suit, v, false)
			sun += BalootCardPoints(c, BalootModeSun, 0)
			hokom += BalootCardPoints(c, BalootModeHokom, trump)
		}
	}
	assert.Equal(t, 120, sun)
	assert.Equal(t, 152, hokom)
}

// --- 序列 ---

// Sun の序列は A>10>K>Q>J>9>8>7。
func TestBaloot_SunRankOrder(t *testing.T) {
	order := []int{1, 10, 13, 12, 11, 9, 8, 7}
	for i := 0; i < len(order)-1; i++ {
		hi := balootSunRank(NewCard(CardDesignSpade, order[i], false))
		lo := balootSunRank(NewCard(CardDesignSpade, order[i+1], false))
		assert.Greater(t, hi, lo, "%d must beat %d", order[i], order[i+1])
	}
	assert.Equal(t, 0, balootSunRank(nil))
}

// Hokom の切り札の序列は J>9>A>10>K>Q>8>7。**Sun とは J と 9 の位置が違う。**
func TestBaloot_HokomRankOrder(t *testing.T) {
	order := []int{11, 9, 1, 10, 13, 12, 8, 7}
	for i := 0; i < len(order)-1; i++ {
		hi := balootHokomRank(NewCard(CardDesignSpade, order[i], false))
		lo := balootHokomRank(NewCard(CardDesignSpade, order[i+1], false))
		assert.Greater(t, hi, lo, "%d must beat %d", order[i], order[i+1])
	}
	assert.Equal(t, 0, balootHokomRank(nil))
}

// **同じ 2 枚が、モードで勝敗が入れ替わる。** これがバルート固有の性質。
func TestBaloot_SameCardsFlipWinnerBetweenModes(t *testing.T) {
	trump := CardDesignHeart
	jack := NewCard(trump, 11, false)
	ace := NewCard(trump, 1, false)

	b := NewDefaultBaloot()
	b.mode = BalootModeSun
	assert.False(t, b.beats(jack, ace, trump), "Sun では A が J に勝つ")

	b.mode, b.trumpSuit = BalootModeHokom, trump
	assert.True(t, b.beats(jack, ace, trump), "Hokom では切り札の J が最強")
}

// Hokom では切り札が非切り札に常に勝つ。Sun には切り札という概念が無い。
func TestBaloot_TrumpOnlyBeatsInHokom(t *testing.T) {
	trump := CardDesignHeart
	lowTrump := NewCard(trump, 7, false)
	aceOfLead := NewCard(CardDesignSpade, 1, false)

	b := NewDefaultBaloot()
	b.mode, b.trumpSuit = BalootModeHokom, trump
	assert.True(t, b.beats(lowTrump, aceOfLead, CardDesignSpade))

	b.mode, b.trumpSuit = BalootModeSun, 0
	assert.False(t, b.beats(lowTrump, aceOfLead, CardDesignSpade),
		"Sun ではリード以外のスートは勝てない")
}

// --- 宣言 ---

// **親は宣言を降りられない。** 全員が降りるとモードが決まらないまま進めなくなる。
func TestBaloot_DealerCannotPass(t *testing.T) {
	b := newTestBaloot(t)
	b.dealerIdx = 0
	b.currentPlayerIdx = 0

	err := b.PassDeclaration()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "dealer")
	assert.Equal(t, BalootPhaseDeclare, b.GetPhase(), "宣言はまだ終わっていない")
}

// 親でなければ見送れて、手番が次へ回る。
func TestBaloot_PassAdvancesTheTurn(t *testing.T) {
	b := newTestBaloot(t)
	b.dealerIdx = 2
	b.currentPlayerIdx = 0

	require.NoError(t, b.PassDeclaration())
	assert.True(t, b.GetPlayer(0).GetDeclared())
	assert.Equal(t, 1, b.GetCurrentPlayerIdx())
}

// **全員が見送ると、最後の親が引き受ける。** 人間が親でない限り自動で決まる。
func TestBaloot_DealerIsForcedToDeclareWhenAllPass(t *testing.T) {
	b := newTestBaloot(t)
	b.dealerIdx = 3
	b.currentPlayerIdx = 0
	// **CPU が自発的に宣言しない手札にする。** そうでないと親まで回らない。
	for i := range BalootPlayerCnt {
		balootHandOf(b, i,
			NewCard(CardDesignSpade, 7, false), NewCard(CardDesignHeart, 8, false),
			NewCard(CardDesignClover, 7, false), NewCard(CardDesignDiamond, 8, false),
			NewCard(CardDesignSpade, 8, false))
	}

	require.NoError(t, b.PassDeclaration())
	for b.GetPhase() == BalootPhaseDeclare {
		b.CpuDeclare()
	}
	assert.Equal(t, BalootPhasePlay, b.GetPhase())
	assert.Equal(t, 3, b.GetDeclarerIdx(), "親が引き受ける")
	assert.Equal(t, BalootModeHokom, b.GetMode())
}

// Sun を宣言すると切り札は無い。
func TestBaloot_DeclareSun(t *testing.T) {
	b := newTestBaloot(t)
	b.currentPlayerIdx = 0

	require.NoError(t, b.DeclareSun())
	assert.Equal(t, BalootModeSun, b.GetMode())
	assert.Equal(t, 0, b.GetTrumpSuit())
	assert.Equal(t, 0, b.GetDeclarerIdx())
	assert.Equal(t, BalootPhasePlay, b.GetPhase())
}

// Hokom を宣言すると指定スートが切り札になる。
func TestBaloot_DeclareHokom(t *testing.T) {
	b := newTestBaloot(t)
	b.currentPlayerIdx = 0

	require.NoError(t, b.DeclareHokom(CardDesignHeart))
	assert.Equal(t, BalootModeHokom, b.GetMode())
	assert.Equal(t, CardDesignHeart, b.GetTrumpSuit())
}

// 存在しないスートは拒否する。
func TestBaloot_DeclareHokomRejectsInvalidSuit(t *testing.T) {
	b := newTestBaloot(t)
	b.currentPlayerIdx = 0

	require.Error(t, b.DeclareHokom(99))
	require.Error(t, b.DeclareHokom(-1))
	assert.Equal(t, BalootPhaseDeclare, b.GetPhase())
}

// 自分の番でない／フェーズ違い／終局後は宣言できない。
func TestBaloot_DeclareGuards(t *testing.T) {
	b := newTestBaloot(t)
	b.currentPlayerIdx = 1
	assert.Error(t, b.DeclareSun())
	assert.Error(t, b.PassDeclaration())

	b.currentPlayerIdx = 0
	b.phase = BalootPhasePlay
	assert.Error(t, b.DeclareSun())

	b.phase = BalootPhaseDeclare
	b.gameEndFlag = true
	assert.Error(t, b.DeclareSun())
}

// CPU は自分の番でなければ宣言しない。
func TestBaloot_CpuDeclareIgnoresHumanTurn(t *testing.T) {
	b := newTestBaloot(t)
	b.currentPlayerIdx = 0

	b.CpuDeclare()
	assert.Equal(t, BalootModeNone, b.GetMode())
	assert.Equal(t, 0, b.GetCurrentPlayerIdx())
}

// --- Baloot 役 ---

// **Baloot は Hokom だけ。** 切り札の K+Q で 20 点。
func TestBaloot_BalootBonusOnlyInHokom(t *testing.T) {
	trump := CardDesignHeart
	b := newTestBaloot(t)
	b.currentPlayerIdx = 0
	balootHandOf(b, 0,
		NewCard(trump, 13, false), NewCard(trump, 12, false),
		NewCard(CardDesignSpade, 7, false), NewCard(CardDesignSpade, 8, false),
		NewCard(CardDesignClover, 7, false))

	require.NoError(t, b.DeclareHokom(trump))
	assert.True(t, b.GetPlayer(0).GetHasBaloot())

	// 同じ手札でも Sun なら成立しない。
	b2 := newTestBaloot(t)
	b2.currentPlayerIdx = 0
	balootHandOf(b2, 0,
		NewCard(trump, 13, false), NewCard(trump, 12, false),
		NewCard(CardDesignSpade, 7, false), NewCard(CardDesignSpade, 8, false),
		NewCard(CardDesignClover, 7, false))
	require.NoError(t, b2.DeclareSun())
	assert.False(t, b2.GetPlayer(0).GetHasBaloot())
}

// K だけ／Q だけでは成立しない。切り札以外の K+Q でも成立しない。
// K だけ／Q だけでは成立しない。切り札以外の K+Q でも成立しない。
//
// **配りの上に手札を積んではいけない。** DeclareHokom は completeDeal で
// 手札を 8 枚まで補充するので、わざと外した切り札の Q が山から来てしまい、
// 11% ほどの確率で Baloot が成立して落ちる（[[feedback_fixture_on_top_of_reset]]）。
// 8 枚ちょうどを組んでから markBaloot を直接呼ぶ。
func TestBaloot_BalootNeedsBothTrumpHonours(t *testing.T) {
	trump := CardDesignHeart
	for _, tc := range []struct {
		name  string
		cards []*Card
	}{
		{"only the trump king", []*Card{
			NewCard(trump, 13, false), NewCard(trump, 7, false), NewCard(trump, 8, false),
			NewCard(CardDesignSpade, 12, false), NewCard(CardDesignSpade, 13, false),
			NewCard(CardDesignClover, 7, false), NewCard(CardDesignClover, 8, false),
			NewCard(CardDesignDiamond, 7, false),
		}},
		{"only the trump queen", []*Card{
			NewCard(trump, 12, false), NewCard(trump, 7, false), NewCard(trump, 8, false),
			NewCard(CardDesignSpade, 12, false), NewCard(CardDesignSpade, 13, false),
			NewCard(CardDesignClover, 7, false), NewCard(CardDesignClover, 8, false),
			NewCard(CardDesignDiamond, 7, false),
		}},
		{"king and queen of another suit", []*Card{
			NewCard(CardDesignSpade, 13, false), NewCard(CardDesignSpade, 12, false),
			NewCard(trump, 7, false), NewCard(trump, 8, false), NewCard(trump, 9, false),
			NewCard(CardDesignClover, 7, false), NewCard(CardDesignClover, 8, false),
			NewCard(CardDesignDiamond, 7, false),
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := newTestBaloot(t)
			b.mode, b.trumpSuit = BalootModeHokom, trump
			balootHandOf(b, 0, tc.cards...)
			b.markBaloot()
			assert.False(t, b.GetPlayer(0).GetHasBaloot())
		})
	}

	// 負のコントロール: 切り札の K+Q が揃っていれば成立する。
	b := newTestBaloot(t)
	b.mode, b.trumpSuit = BalootModeHokom, trump
	balootHandOf(b, 0,
		NewCard(trump, 13, false), NewCard(trump, 12, false), NewCard(trump, 7, false),
		NewCard(CardDesignSpade, 13, false), NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignClover, 7, false), NewCard(CardDesignClover, 8, false),
		NewCard(CardDesignDiamond, 7, false))
	b.markBaloot()
	assert.True(t, b.GetPlayer(0).GetHasBaloot())
}

// --- プレイ ---

// リードのスートを持っていれば必ずフォローする。
func TestBaloot_MustFollowSuit(t *testing.T) {
	b := newTestBaloot(t)
	b.currentPlayerIdx = 0
	require.NoError(t, b.DeclareSun())

	b.currentTrick = []*TrickCard{{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 13, false)}}
	b.currentPlayerIdx = 0
	balootHandOf(b, 0, NewCard(CardDesignHeart, 1, false), NewCard(CardDesignSpade, 7, false))

	require.Error(t, b.PlayerPlay(0), "スペードを持っているのにハートは出せない")
	require.NoError(t, b.PlayerPlay(1))
}

// リードのスートが無ければ何を出してもよい。
func TestBaloot_MayDiscardWhenVoid(t *testing.T) {
	b := newTestBaloot(t)
	b.currentPlayerIdx = 0
	require.NoError(t, b.DeclareSun())

	b.currentTrick = []*TrickCard{{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 13, false)}}
	b.currentPlayerIdx = 0
	balootHandOf(b, 0, NewCard(CardDesignHeart, 1, false), NewCard(CardDesignClover, 7, false))

	assert.Equal(t, []int{0, 1}, b.GetValidPlayIndices(0))
	require.NoError(t, b.PlayerPlay(0))
}

// 範囲外のインデックスは拒否する。
func TestBaloot_PlayRejectsInvalidIndex(t *testing.T) {
	b := newTestBaloot(t)
	b.currentPlayerIdx = 0
	require.NoError(t, b.DeclareSun())
	b.currentPlayerIdx = 0

	assert.Error(t, b.PlayerPlay(-1))
	assert.Error(t, b.PlayerPlay(99))
}

// 自分の番でない／フェーズ違い／終局後は出せない。
func TestBaloot_PlayGuards(t *testing.T) {
	b := newTestBaloot(t)
	assert.Error(t, b.PlayerPlay(0), "宣言フェーズでは出せない")

	b.currentPlayerIdx = 0
	require.NoError(t, b.DeclareSun())
	b.currentPlayerIdx = 1
	assert.Error(t, b.PlayerPlay(0))

	b.currentPlayerIdx = 0
	b.gameEndFlag = true
	assert.Error(t, b.PlayerPlay(0))
}

// 範囲外プレイヤーの合法手は nil。
func TestBaloot_ValidIndicesOutOfRange(t *testing.T) {
	b := newTestBaloot(t)
	assert.Nil(t, b.GetValidPlayIndices(-1))
	assert.Nil(t, b.GetValidPlayIndices(BalootPlayerCnt))
}

// トリックは 4 枚そろって解決し、勝者がリードを取る。
func TestBaloot_TrickResolution(t *testing.T) {
	b := newTestBaloot(t)
	b.currentPlayerIdx = 0
	require.NoError(t, b.DeclareHokom(CardDesignHeart))
	b.leadPlayerIdx, b.currentPlayerIdx = 0, 0

	balootHandOf(b, 0, NewCard(CardDesignSpade, 1, false))
	balootHandOf(b, 1, NewCard(CardDesignSpade, 13, false))
	balootHandOf(b, 2, NewCard(CardDesignHeart, 7, false)) // 切り札の 7
	balootHandOf(b, 3, NewCard(CardDesignSpade, 12, false))

	require.NoError(t, b.play(0, 0))
	require.NoError(t, b.play(1, 0))
	require.NoError(t, b.play(2, 0))
	require.NoError(t, b.play(3, 0))

	// **切り札の 7 が A に勝つ。** 11+4+0+3 = 18 点がチーム0 に入る。
	assert.Equal(t, 2, b.GetLeadPlayerIdx())
	assert.Equal(t, 18, b.GetRoundPoints(BalootTeamOf(2)))
	assert.Equal(t, 1, b.GetTrickNumber())
	assert.Empty(t, b.GetCurrentTrick())
}

// 向かい合う席が味方。
func TestBaloot_TeamAssignment(t *testing.T) {
	assert.Equal(t, BalootTeamOf(0), BalootTeamOf(2))
	assert.Equal(t, BalootTeamOf(1), BalootTeamOf(3))
	assert.NotEqual(t, BalootTeamOf(0), BalootTeamOf(1))
}

// **1 ラウンドの総得点はモードで決まる。** Sun 130 / Hokom 162（最終トリック +10 込み）。
func TestBaloot_RoundTotalMatchesMode(t *testing.T) {
	for range 20 {
		b := newTestBaloot(t)
		settleMode(t, b)
		mode := b.GetMode()

		for b.GetPhase() == BalootPhasePlay {
			if b.IsHumanTurn() {
				valid := b.GetValidPlayIndices(0)
				require.NotEmpty(t, valid)
				require.NoError(t, b.PlayerPlay(valid[0]))
				continue
			}
			b.CpuPlay()
		}

		want := 120 + BalootLastTrickBonus
		if mode == BalootModeHokom {
			want = 152 + BalootLastTrickBonus
		}
		for i := range BalootPlayerCnt {
			if b.GetPlayer(i).GetHasBaloot() {
				want += BalootBonus
			}
		}
		got := b.GetRoundPoints(0) + b.GetRoundPoints(1)
		assert.Equal(t, want, got, "mode=%d", mode)
		assert.Equal(t, BalootTricksPerRound, b.GetTrickNumber())
	}
}

// 目標点に届かなければ次ラウンドへ、届けば終局。
func TestBaloot_NextRoundAndGameEnd(t *testing.T) {
	b := newTestBaloot(t)
	b.SetConfig(BalootConfig{Target: BalootTargetMax})
	settleMode(t, b)
	for b.GetPhase() == BalootPhasePlay {
		if b.IsHumanTurn() {
			require.NoError(t, b.PlayerPlay(b.GetValidPlayIndices(0)[0]))
			continue
		}
		b.CpuPlay()
	}
	require.Equal(t, BalootPhaseRoundEnd, b.GetPhase())
	assert.Equal(t, 1, b.GetRoundNumber())

	b.NextRound()
	assert.Equal(t, 2, b.GetRoundNumber())
	assert.Equal(t, 1, b.GetDealerIdx(), "親は時計回りに動く")
	assert.Equal(t, BalootPhaseDeclare, b.GetPhase())
	assert.Equal(t, BalootModeNone, b.GetMode(), "モードはラウンドごとに宣言し直す")
}

// 目標点に届いたラウンドで終局し、勝ちチームが決まる。
func TestBaloot_GameEndsAtTarget(t *testing.T) {
	b := newTestBaloot(t)
	b.SetConfig(BalootConfig{Target: BalootTargetMin})
	settleMode(t, b)
	for b.GetPhase() == BalootPhasePlay {
		if b.IsHumanTurn() {
			require.NoError(t, b.PlayerPlay(b.GetValidPlayIndices(0)[0]))
			continue
		}
		b.CpuPlay()
	}

	assert.True(t, b.GetGameEndFlag())
	assert.Equal(t, BalootPhaseGameEnd, b.GetPhase())
	// **勝ちチームは総得点の高い方。** 引き分けだけが -1。
	switch {
	case b.GetScore(0) > b.GetScore(1):
		assert.Equal(t, 0, b.GetWinnerTeam())
	case b.GetScore(1) > b.GetScore(0):
		assert.Equal(t, 1, b.GetWinnerTeam())
	default:
		assert.Equal(t, -1, b.GetWinnerTeam())
	}

	// 終局後は NextRound が効かない。
	before := b.GetRoundNumber()
	b.NextRound()
	assert.Equal(t, before, b.GetRoundNumber())
}

// --- ヒント ---

// 宣言フェーズのヒントは札ではなくモードを勧める。
func TestBaloot_HintDuringDeclare(t *testing.T) {
	b := newTestBaloot(t)
	b.currentPlayerIdx = 0
	balootHandOf(b, 0,
		NewCard(CardDesignHeart, 11, false), NewCard(CardDesignHeart, 9, false),
		NewCard(CardDesignHeart, 1, false), NewCard(CardDesignSpade, 7, false),
		NewCard(CardDesignClover, 8, false))

	h := b.GetHint()
	require.NotNil(t, h)
	assert.Nil(t, h.CardIndex)
	assert.Equal(t, "balootDeclareHokom", h.Reason)
	assert.Equal(t, CardDesignHeart, h.Suit)
}

// 弱い手なら見送りを勧める。
func TestBaloot_HintSuggestsPass(t *testing.T) {
	b := newTestBaloot(t)
	b.currentPlayerIdx = 0
	balootHandOf(b, 0,
		NewCard(CardDesignSpade, 7, false), NewCard(CardDesignHeart, 8, false),
		NewCard(CardDesignClover, 7, false), NewCard(CardDesignDiamond, 8, false),
		NewCard(CardDesignSpade, 8, false))

	h := b.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "balootPassDeclare", h.Reason)
}

// A と 10 が厚ければ Sun を勧める。
func TestBaloot_HintSuggestsSun(t *testing.T) {
	b := newTestBaloot(t)
	b.currentPlayerIdx = 0
	balootHandOf(b, 0,
		NewCard(CardDesignSpade, 1, false), NewCard(CardDesignHeart, 1, false),
		NewCard(CardDesignClover, 1, false), NewCard(CardDesignDiamond, 10, false),
		NewCard(CardDesignSpade, 8, false))

	h := b.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "balootDeclareSun", h.Reason)
	assert.Equal(t, 0, h.Suit)
}

// プレイ中のヒントは合法な札を指す。
func TestBaloot_HintDuringPlay(t *testing.T) {
	b := newTestBaloot(t)
	b.currentPlayerIdx = 0
	require.NoError(t, b.DeclareSun())
	b.currentPlayerIdx = 0

	h := b.GetHint()
	require.NotNil(t, h)
	require.NotNil(t, h.CardIndex)
	assert.Contains(t, b.GetValidPlayIndices(0), *h.CardIndex)
}

// 自分の番でなければヒントは無い。
func TestBaloot_HintNilWhenNotHumanTurn(t *testing.T) {
	b := newTestBaloot(t)
	b.currentPlayerIdx = 0
	require.NoError(t, b.DeclareSun())
	b.currentPlayerIdx = 2
	assert.Nil(t, b.GetHint())

	b.gameEndFlag = true
	assert.Nil(t, b.GetHint())
}

// --- CPU ---

// **味方が勝っているときは点の高い札を乗せる。** 取りに行って味方を潰さない。
func TestBaloot_CpuFeedsWinningPartner(t *testing.T) {
	b := newTestBaloot(t)
	b.currentPlayerIdx = 0
	require.NoError(t, b.DeclareSun())

	// プレイヤー0（プレイヤー2 の味方）が A でリード済み。
	b.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 1, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 7, false)},
	}
	b.currentPlayerIdx = 2
	balootHandOf(b, 2,
		NewCard(CardDesignSpade, 8, false),  // 0 点
		NewCard(CardDesignSpade, 10, false)) // 10 点

	assert.True(t, b.partnerIsWinning(2))
	assert.Equal(t, 1, b.chooseCpuCard(2), "点の高い 10 を乗せる")
}

// 相手が勝っているなら、取れるなら一番安い札で取る。
func TestBaloot_CpuTakesTrickCheaply(t *testing.T) {
	b := newTestBaloot(t)
	b.currentPlayerIdx = 0
	require.NoError(t, b.DeclareSun())

	b.currentTrick = []*TrickCard{{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 12, false)}}
	b.currentPlayerIdx = 2
	balootHandOf(b, 2,
		NewCard(CardDesignSpade, 1, false),  // 勝てるが 11 点
		NewCard(CardDesignSpade, 13, false), // 勝てて 4 点
		NewCard(CardDesignSpade, 7, false))  // 勝てない

	assert.Equal(t, 1, b.chooseCpuCard(2), "K で足りる")
}

// 取れないなら一番安い札を捨てる。
func TestBaloot_CpuDucksWhenItCannotWin(t *testing.T) {
	b := newTestBaloot(t)
	b.currentPlayerIdx = 0
	require.NoError(t, b.DeclareSun())

	b.currentTrick = []*TrickCard{{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 1, false)}}
	b.currentPlayerIdx = 2
	balootHandOf(b, 2,
		NewCard(CardDesignSpade, 13, false), // 4 点
		NewCard(CardDesignSpade, 7, false))  // 0 点

	assert.Equal(t, 1, b.chooseCpuCard(2))
}

// リードなら一番強い札を出す。
func TestBaloot_CpuLeadsHighest(t *testing.T) {
	b := newTestBaloot(t)
	b.currentPlayerIdx = 0
	require.NoError(t, b.DeclareSun())
	b.currentTrick = nil
	b.currentPlayerIdx = 1
	balootHandOf(b, 1,
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignSpade, 11, false))

	assert.Equal(t, 1, b.chooseCpuCard(1))
}

// CPU は人間の番には打たない。
func TestBaloot_CpuPlayIgnoresHumanTurn(t *testing.T) {
	b := newTestBaloot(t)
	b.currentPlayerIdx = 0
	require.NoError(t, b.DeclareSun())
	b.currentPlayerIdx = 0
	size := b.GetPlayer(0).GetCardsSize()

	b.CpuPlay()
	assert.Equal(t, size, b.GetPlayer(0).GetCardsSize())
}

// --- その他 ---

// ギブアップで相手チームの勝ち。
func TestBaloot_GiveUp(t *testing.T) {
	b := newTestBaloot(t)
	b.GiveUp()
	assert.True(t, b.GetGameEndFlag())
	assert.Equal(t, 1, b.GetWinnerTeam())

	// 二度目は何も起きない。
	b.winnerTeam = 0
	b.GiveUp()
	assert.Equal(t, 0, b.GetWinnerTeam())
}

// 範囲外のアクセサは安全な値を返す。
func TestBaloot_AccessorsOutOfRange(t *testing.T) {
	b := newTestBaloot(t)
	assert.Nil(t, b.GetPlayer(-1))
	assert.Nil(t, b.GetPlayer(BalootPlayerCnt))
	assert.Equal(t, 0, b.GetScore(-1))
	assert.Equal(t, 0, b.GetScore(BalootTeamCnt))
	assert.Equal(t, 0, b.GetRoundPoints(-1))
	assert.Equal(t, 0, b.GetRoundPoints(BalootTeamCnt))
	assert.Equal(t, BalootPlayerCnt, b.GetPlayerCnt())

	b.SetScoreForTestUse(0, 42)
	assert.Equal(t, 42, b.GetScore(0))
	b.SetScoreForTestUse(-1, 7) // 何も起きない
	assert.Equal(t, 42, b.GetScore(0))
}

// 設定のバリデーション。
func TestBalootConfig_Validate(t *testing.T) {
	assert.NoError(t, DefaultBalootConfig().Validate())
	assert.NoError(t, BalootConfig{Target: BalootTargetMin}.Validate())
	assert.NoError(t, BalootConfig{Target: BalootTargetMax}.Validate())
	assert.Error(t, BalootConfig{Target: BalootTargetMin - 1}.Validate())
	assert.Error(t, BalootConfig{Target: BalootTargetMax + 1}.Validate())
}

// --- JSON 往復 ---

// **Worker はリクエストごとに KV から作り直す。** モードと切り札が往復しないと
// 札の強さが毎回変わる。
func TestBaloot_JSONRoundTrip(t *testing.T) {
	b := newTestBaloot(t)
	b.currentPlayerIdx = 0
	require.NoError(t, b.DeclareHokom(CardDesignHeart))
	b.currentPlayerIdx = 0
	require.NoError(t, b.PlayerPlay(b.GetValidPlayIndices(0)[0]))
	b.SetScoreForTestUse(0, 33)
	b.GetPlayer(1).SetHasBaloot(true)

	data, err := json.Marshal(b)
	require.NoError(t, err)

	var got Baloot
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, b.GetMode(), got.GetMode())
	assert.Equal(t, b.GetTrumpSuit(), got.GetTrumpSuit())
	assert.Equal(t, b.GetDeclarerIdx(), got.GetDeclarerIdx())
	assert.Equal(t, b.GetPhase(), got.GetPhase())
	assert.Equal(t, 33, got.GetScore(0))
	assert.Equal(t, b.GetCurrentPlayerIdx(), got.GetCurrentPlayerIdx())
	assert.Equal(t, b.GetDealerIdx(), got.GetDealerIdx())
	assert.Len(t, got.GetCurrentTrick(), len(b.GetCurrentTrick()))
	assert.True(t, got.GetPlayer(1).GetHasBaloot())
	for i := range BalootPlayerCnt {
		assert.Equal(t, b.GetPlayer(i).GetCardsSize(), got.GetPlayer(i).GetCardsSize(), "player %d", i)
	}
}

// 壊れたスナップショットは復元しない。
func TestBaloot_UnmarshalRejectsInvalid(t *testing.T) {
	valid := func() balootJSON {
		return balootJSON{
			Config:      DefaultBalootConfig(),
			Phase:       BalootPhasePlay,
			Mode:        BalootModeSun,
			RoundNumber: 1,
			DeclarerIdx: 0,
			WinnerTeam:  -1,
		}
	}
	cases := map[string]func(*balootJSON){
		"bad config": func(j *balootJSON) { j.Config.Target = 0 },
		"bad phase":  func(j *balootJSON) { j.Phase = BalootPhase(99) },
		"bad mode":   func(j *balootJSON) { j.Mode = BalootMode(99) },
		"bad trick":  func(j *balootJSON) { j.TrickNumber = BalootTricksPerRound + 1 },
		// **切り札はモードと整合していなければならない。** どちらの向きも踏む。
		"hokom without a suit": func(j *balootJSON) {
			j.Mode, j.TrumpSuit = BalootModeHokom, 0
		},
		"hokom with a bogus suit": func(j *balootJSON) {
			j.Mode, j.TrumpSuit = BalootModeHokom, 99
		},
		"sun carrying a trump suit": func(j *balootJSON) {
			j.Mode, j.TrumpSuit = BalootModeSun, CardDesignHeart
		},
		"undeclared carrying a trump suit": func(j *balootJSON) {
			j.Mode, j.TrumpSuit = BalootModeNone, CardDesignSpade
		},
		"bad round":    func(j *balootJSON) { j.RoundNumber = 0 },
		"bad current":  func(j *balootJSON) { j.CurrentPlayerIdx = BalootPlayerCnt },
		"bad lead":     func(j *balootJSON) { j.LeadPlayerIdx = -1 },
		"bad dealer":   func(j *balootJSON) { j.DealerIdx = BalootPlayerCnt },
		"bad declarer": func(j *balootJSON) { j.DeclarerIdx = BalootPlayerCnt },
		"bad winner":   func(j *balootJSON) { j.WinnerTeam = BalootTeamCnt },
		"long trick": func(j *balootJSON) {
			j.CurrentTrick = make([]*TrickCard, BalootPlayerCnt+1)
		},
		"long log": func(j *balootJSON) {
			j.ActionLog = make([]*ActionLogEntry, balootMaxSliceLen+1)
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			j := valid()
			mutate(&j)
			data, err := json.Marshal(j)
			require.NoError(t, err)
			var got Baloot
			assert.Error(t, json.Unmarshal(data, &got))
		})
	}

	var got Baloot
	assert.Error(t, got.UnmarshalJSON([]byte("{")))

	// 正しいスナップショットは通る（ガードの負のコントロール）。
	data, err := json.Marshal(valid())
	require.NoError(t, err)
	assert.NoError(t, json.Unmarshal(data, &got))

	// **Hokom + 実在するスートも通る。** 上のガードが Hokom を一律に弾いて
	// いないことを確かめる（負のコントロール 2 本目）。
	j := valid()
	j.Mode, j.TrumpSuit = BalootModeHokom, CardDesignDiamond
	hokom, err := json.Marshal(j)
	require.NoError(t, err)
	var okHokom Baloot
	require.NoError(t, json.Unmarshal(hokom, &okHokom))
	assert.Equal(t, CardDesignDiamond, okHokom.GetTrumpSuit())
}

// 棋譜が積まれること。
func TestBaloot_ActionLog(t *testing.T) {
	b := newTestBaloot(t)
	require.NotEmpty(t, b.actionLog, "配りが記録される")
	b.currentPlayerIdx = 0
	require.NoError(t, b.DeclareSun())
	b.currentPlayerIdx = 0
	require.NoError(t, b.PlayerPlay(b.GetValidPlayIndices(0)[0]))

	kinds := map[string]bool{}
	for _, e := range b.actionLog {
		kinds[e.ActionType] = true
	}
	assert.True(t, kinds["declare"])
	assert.True(t, kinds["play"])
}

// **配られた瞬間に相手の Baloot が割れるのは体験を壊す** (#5750)。
// 切り札の K か Q を実際に出した時点で初めて開く。
func TestBalootRevealsTheBonusOnlyWhenTheKingOrQueenIsPlayed(t *testing.T) {
	b := newTestBaloot(t)
	b.SetModeForTest(BalootModeHokom)
	b.SetTrumpSuitForTest(CardDesignSpade)
	b.SetPhaseForTest(BalootPhasePlay)
	b.SetCurrentPlayerIdxForTest(1)
	b.SetCurrentTrickForTest(nil)

	// CPU1 が Baloot 持ち。人間 (0) も持っているが、こちらは最初から見えている。
	balootHandOf(b, 1, NewCard(CardDesignSpade, 13, true), NewCard(CardDesignSpade, 12, true))
	balootHandOf(b, 0, NewCard(CardDesignHeart, 13, true), NewCard(CardDesignHeart, 12, true))
	b.GetPlayer(1).SetHasBaloot(true)
	b.GetPlayer(1).SetBalootRevealed(false)
	b.GetPlayer(0).SetHasBaloot(true)
	b.GetPlayer(0).SetBalootRevealed(true)

	// **印を付けるのは markBaloot。**そこで開示まで済ませてしまうと、配った
	// 瞬間に相手の手の内が割れる。実際に markBaloot を通して確かめる。
	marked := newTestBaloot(t)
	marked.SetModeForTest(BalootModeHokom)
	marked.SetTrumpSuitForTest(CardDesignSpade)
	balootHandOf(marked, 0, NewCard(CardDesignSpade, 13, true), NewCard(CardDesignSpade, 12, true))
	balootHandOf(marked, 1, NewCard(CardDesignSpade, 13, true), NewCard(CardDesignSpade, 12, true))
	marked.markBaloot()
	if !marked.GetPlayer(0).GetHasBaloot() || !marked.GetPlayer(1).GetHasBaloot() {
		t.Fatal("markBaloot must flag both hands holding the trump king and queen")
	}
	if !marked.GetPlayer(0).GetBalootRevealed() {
		t.Error("the human's own Baloot is visible from the moment it is dealt")
	}
	if marked.GetPlayer(1).GetBalootRevealed() {
		t.Error("a CPU's Baloot must not be public just because it was dealt")
	}

	if b.GetPlayer(1).GetBalootRevealed() {
		t.Fatal("the CPU's Baloot must stay hidden before it plays the king or the queen")
	}
	if !b.GetPlayer(0).GetBalootRevealed() {
		t.Fatal("the human's own Baloot is visible to the human from the start")
	}

	// 切り札でない札を出しても開かない (負のコントロール)。
	balootHandOf(b, 1, NewCard(CardDesignHeart, 7, true), NewCard(CardDesignSpade, 13, true))
	if err := b.play(1, 0); err != nil {
		t.Fatalf("playing a plain card: %v", err)
	}
	if b.GetPlayer(1).GetBalootRevealed() {
		t.Error("a plain card must not reveal the Baloot")
	}

	// 切り札の K を出したら開く。
	b.SetCurrentPlayerIdxForTest(1)
	b.SetCurrentTrickForTest(nil)
	balootHandOf(b, 1, NewCard(CardDesignSpade, 13, true))
	if err := b.play(1, 0); err != nil {
		t.Fatalf("playing the trump king: %v", err)
	}
	if !b.GetPlayer(1).GetBalootRevealed() {
		t.Error("playing the trump king must reveal the Baloot")
	}
}

// **精算では全員ぶんが開く。**加点の根拠になるので伏せたままにはできない。
func TestBalootRevealsEveryBonusAtTheSettlement(t *testing.T) {
	b := newTestBaloot(t)
	b.SetModeForTest(BalootModeHokom)
	b.SetTrumpSuitForTest(CardDesignSpade)
	b.GetPlayer(2).SetHasBaloot(true)
	b.GetPlayer(2).SetBalootRevealed(false)

	b.finishRound()

	if !b.GetPlayer(2).GetBalootRevealed() {
		t.Error("the settlement must reveal every Baloot")
	}
}
