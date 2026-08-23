//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestMadrasso() *domain.Madrasso {
	players := []*domain.MadrassoPlayer{
		domain.NewMadrassoPlayer(true),  // 0 = human (team A)
		domain.NewMadrassoPlayer(false), // 1 = CPU  (team B)
		domain.NewMadrassoPlayer(false), // 2 = CPU  (team A)
		domain.NewMadrassoPlayer(false), // 3 = CPU  (team B)
	}
	return domain.NewMadrasso(domain.NewTrumpCardsBriscola(), players, domain.DefaultMadrassoConfig())
}

func madCard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

func madSetHand(p *domain.MadrassoPlayer, cards ...*domain.Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// madResolve sets up a complete trick at trickNum and resolves it.
func madResolve(g *domain.Madrasso, trickNum int, trick []*domain.TrickCard) {
	g.SetTrickNumber(trickNum)
	g.SetCurrentTrick(trick)
	g.SetPhase(domain.MadrassoPhaseTrickEnd)
	g.ResolveTrick()
}

// --- construction ---

func TestNewMadrasso(t *testing.T) {
	g := newTestMadrasso()
	assert.Equal(t, domain.MadrassoPhase(0), g.GetPhase())
	assert.Equal(t, 0, g.GetRoundNumber())
	assert.Equal(t, 4, g.GetPlayerCnt())
	assert.Equal(t, -1, g.GetWinnerTeam())
	assert.False(t, g.GetGameEndFlag())
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(99))
}

func TestNewDefaultMadrasso(t *testing.T) {
	g := domain.NewDefaultMadrasso()
	assert.Equal(t, 4, g.GetPlayerCnt())
	assert.True(t, g.GetPlayer(0).GetIsHuman())
	assert.False(t, g.GetPlayer(1).GetIsHuman())
}

func TestMadrassoTeamOf(t *testing.T) {
	assert.Equal(t, 0, domain.MadrassoTeamOf(0))
	assert.Equal(t, 1, domain.MadrassoTeamOf(1))
	assert.Equal(t, 0, domain.MadrassoTeamOf(2))
	assert.Equal(t, 1, domain.MadrassoTeamOf(3))
}

func TestMadrassoReset(t *testing.T) {
	g := newTestMadrasso()
	g.Reset()
	assert.Equal(t, domain.MadrassoPhasePlay, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, 1, g.GetTrickNumber())
	assert.False(t, g.GetGameEndFlag())
	total := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		assert.Equal(t, 10, g.GetPlayer(i).GetCardsSize())
		total += g.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, 40, total)
	// Round 1 leader is player 0 and it is their turn.
	assert.Equal(t, 0, g.GetCurrentPlayerIdx())
}

// --- config ---

func TestMadrassoConfigValidate(t *testing.T) {
	assert.NoError(t, domain.DefaultMadrassoConfig().Validate())
	assert.Error(t, domain.MadrassoConfig{CpuDifficulty: 99, TargetPoints: 21}.Validate())
	assert.Error(t, domain.MadrassoConfig{CpuDifficulty: domain.MadrassoCpuDifficultyEasy, TargetPoints: 0}.Validate())
}

// --- trick resolution & strength ---

func TestMadrassoTrickWinnerHighestStrength(t *testing.T) {
	g := newTestMadrasso()
	g.SetTrumpSuitForTest(domain.CardDesignHeart) // 場に切り札は出ていない
	// **A が最強。** クローン元のトレセッテは 3 > 2 > A なので、同じ配りでも
	// 勝者が変わる。全部スペードなので A を出した席 0 (チーム A) が勝つ。
	madResolve(g, 1, []*domain.TrickCard{
		{PlayerIdx: 0, Card: madCard(domain.CardDesignSpade, 1)},  // A (最強)
		{PlayerIdx: 1, Card: madCard(domain.CardDesignSpade, 2)},  // 2 (最弱)
		{PlayerIdx: 2, Card: madCard(domain.CardDesignSpade, 3)},  // 3
		{PlayerIdx: 3, Card: madCard(domain.CardDesignSpade, 13)}, // K
	})
	// 整数点: A(11) + 2(0) + 3(10) + K(4) = 25。
	pts := g.GetTeamRoundPoints()
	assert.Equal(t, 25, pts[0])
	assert.Equal(t, 0, pts[1])
	assert.Equal(t, 0, g.GetLeadPlayerIdx(), "A を出した席が勝つ (3 ではない)")
	assert.Equal(t, domain.MadrassoPhaseTrickEnd, g.GetPhase())
}

// TestMadrassoRankOrderFollowsThePoints は札位の順を**全ペアで**突き合わせる。
//
// クローン元のトレセッテは 3 > 2 > A、こちらはブリスコラ系の A > 3 > K > ...。
// **点の高い札ほど強い**という関係も併せて固定する ── 片方だけ直すと、
// 一番高い札が一番弱いという盤面ができる。
func TestMadrassoRankOrderFollowsThePoints(t *testing.T) {
	ranked := []int{1, 3, 13, 12, 11, 7, 6, 5, 4, 2}
	for i := 1; i < len(ranked); i++ {
		assert.Greater(t, domain.MadrassoStrengthForTest(ranked[i-1]), domain.MadrassoStrengthForTest(ranked[i]),
			"%d は %d より強いこと", ranked[i-1], ranked[i])
	}
	// 点を持つ札 (A,3,K,Q,J) の中では、強い順と点の高い順が一致する。
	scoring := []int{1, 3, 13, 12, 11}
	for i := 1; i < len(scoring); i++ {
		assert.Greater(t, domain.MadrassoPointsForTest(scoring[i-1]), domain.MadrassoPointsForTest(scoring[i]),
			"%d は %d より点が高いこと", scoring[i-1], scoring[i])
	}
	// **2 は最弱かつ 0 点。** クローン元では A より強かった。
	assert.Equal(t, 0, domain.MadrassoStrengthForTest(2))
	assert.Equal(t, 0, domain.MadrassoPointsForTest(2))
}

// TestMadrassoTrumpBeatsAnyPlainCard は、配りで決まった切り札が平札に勝つことを見る。
//
// クローン元のトレセッテに切り札は無いので、この経路は完全に新規。
func TestMadrassoTrumpBeatsAnyPlainCard(t *testing.T) {
	g := newTestMadrasso()
	g.SetTrumpSuitForTest(domain.CardDesignHeart)
	// リードはスペードの A (平札としては最強)。席 2 が切り札の最弱札で取る。
	madResolve(g, 1, []*domain.TrickCard{
		{PlayerIdx: 0, Card: madCard(domain.CardDesignSpade, 1)},  // 平札 A
		{PlayerIdx: 1, Card: madCard(domain.CardDesignSpade, 3)},  // 平札 3
		{PlayerIdx: 2, Card: madCard(domain.CardDesignHeart, 2)},  // 切り札の 2 (最弱)
		{PlayerIdx: 3, Card: madCard(domain.CardDesignSpade, 13)}, // 平札 K
	})
	assert.Equal(t, 2, g.GetLeadPlayerIdx(), "切り札はどの平札にも勝つ")

	// **負のコントロール。** 切り札でなければ 2 は最弱のまま。
	h := newTestMadrasso()
	h.SetTrumpSuitForTest(domain.CardDesignDiamond)
	madResolve(h, 1, []*domain.TrickCard{
		{PlayerIdx: 0, Card: madCard(domain.CardDesignSpade, 1)},
		{PlayerIdx: 1, Card: madCard(domain.CardDesignSpade, 3)},
		{PlayerIdx: 2, Card: madCard(domain.CardDesignHeart, 2)},
		{PlayerIdx: 3, Card: madCard(domain.CardDesignSpade, 13)},
	})
	assert.Equal(t, 0, h.GetLeadPlayerIdx(), "切り札でないハートは勝てない")
}

func TestMadrassoOffsuitDoesNotWin(t *testing.T) {
	g := newTestMadrasso()
	// Lead spade; player 1 throws a high heart that cannot win (no trump).
	madResolve(g, 1, []*domain.TrickCard{
		{PlayerIdx: 0, Card: madCard(domain.CardDesignSpade, 5)},
		{PlayerIdx: 1, Card: madCard(domain.CardDesignHeart, 3)}, // off-suit
		{PlayerIdx: 2, Card: madCard(domain.CardDesignSpade, 6)},
		{PlayerIdx: 3, Card: madCard(domain.CardDesignSpade, 4)},
	})
	// Highest spade is 6 (player 2, team A).
	assert.Equal(t, 2, g.GetLeadPlayerIdx())
}

func TestMadrassoUltimaBonus(t *testing.T) {
	g := newTestMadrasso()
	g.SetTrumpSuitForTest(domain.CardDesignHeart)
	// **A が最強なので、この配りを取るのは席 1 (チーム B)。**
	madResolve(g, domain.MadrassoTrickCount, []*domain.TrickCard{
		{PlayerIdx: 0, Card: madCard(domain.CardDesignSpade, 3)},  // 3 = 10 点
		{PlayerIdx: 1, Card: madCard(domain.CardDesignSpade, 1)},  // A = 11 点 (最強)
		{PlayerIdx: 2, Card: madCard(domain.CardDesignSpade, 2)},  // 0 点
		{PlayerIdx: 3, Card: madCard(domain.CardDesignSpade, 13)}, // K = 4 点
	})
	// 整数点: 11 + 10 + 0 + 4 = 25 + ウルティマ 1 = 26。
	pts := g.GetTeamRoundPoints()
	assert.Equal(t, 26, pts[1], "ウルティマが乗ること")
	assert.Equal(t, 0, pts[0])
	assert.Equal(t, domain.MadrassoPhaseRoundEnd, g.GetPhase())
}

// TestMadrassoUltimaFiresOnTheLastTrick は、**最終トリックでだけ**ボーナスが
// 乗ることを見る。定数がずれると永久に発火しない。
func TestMadrassoUltimaFiresOnTheLastTrick(t *testing.T) {
	trick := []*domain.TrickCard{
		{PlayerIdx: 0, Card: madCard(domain.CardDesignSpade, 13)}, // K = 4 点
		{PlayerIdx: 1, Card: madCard(domain.CardDesignSpade, 2)},
		{PlayerIdx: 2, Card: madCard(domain.CardDesignSpade, 4)},
		{PlayerIdx: 3, Card: madCard(domain.CardDesignSpade, 5)},
	}
	before := newTestMadrasso()
	before.SetTrumpSuitForTest(domain.CardDesignHeart)
	madResolve(before, domain.MadrassoTrickCount-1, trick)
	assert.Equal(t, 4, before.GetTeamRoundPoints()[0], "最終トリックの前はボーナス無し")

	last := newTestMadrasso()
	last.SetTrumpSuitForTest(domain.CardDesignHeart)
	madResolve(last, domain.MadrassoTrickCount, trick)
	assert.Equal(t, 5, last.GetTeamRoundPoints()[0], "最終トリックでボーナスが乗ること")
}

func TestMadrassoResolveTrickGuards(t *testing.T) {
	g := newTestMadrasso()
	g.SetPhase(domain.MadrassoPhasePlay) // wrong phase
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 0, Card: madCard(domain.CardDesignSpade, 3)}})
	g.ResolveTrick()
	assert.Equal(t, domain.MadrassoPhasePlay, g.GetPhase()) // unchanged
}

// --- scoring ---

// TestMadrassoScoreRoundPointsToScore は、**カード点の過半を取った側が
// そのディールを取り 1 点**という数え方を見る。
//
// クローン元のトレセッテは 1/3 点を 3 で割って積むが、こちらのカード点は
// 1 ディールで 120 点あり、そのまま積むと 1 ディールで目標を越える。
func TestMadrassoScoreRoundPointsToScore(t *testing.T) {
	// 過半 (61 点以上) を取った側が 1 点。
	g := newTestMadrasso()
	g.SetPhase(domain.MadrassoPhaseRoundEnd)
	g.SetTeamRoundPointsForTest(0, 70)
	g.SetTeamRoundPointsForTest(1, 50)
	g.ScoreRound()
	assert.Equal(t, 1, g.GetTeamScores()[0])
	assert.Equal(t, 0, g.GetTeamScores()[1])
	assert.False(t, g.GetGameEndFlag())

	// **ちょうど半分では誰も取らない。** 60-60 は引き分け。
	h := newTestMadrasso()
	h.SetPhase(domain.MadrassoPhaseRoundEnd)
	h.SetTeamRoundPointsForTest(0, 60)
	h.SetTeamRoundPointsForTest(1, 60)
	h.ScoreRound()
	assert.Equal(t, [2]int{0, 0}, h.GetTeamScores(), "引き分けのディールは誰も取らない")
}

func TestMadrassoScoreRoundWrongPhaseNoop(t *testing.T) {
	g := newTestMadrasso()
	g.SetPhase(domain.MadrassoPhasePlay)
	g.ScoreRound()
	assert.Equal(t, [2]int{0, 0}, g.GetTeamScores())
}

func TestMadrassoGameEnd(t *testing.T) {
	g := newTestMadrasso()
	g.SetTeamScores([2]int{20, 0})
	// Team A wins 7 thirds (=2 pts) at trick 10 → 22 ≥ 21 and > 0 → game end.
	madResolve(g, 10, []*domain.TrickCard{
		{PlayerIdx: 0, Card: madCard(domain.CardDesignSpade, 3)},
		{PlayerIdx: 1, Card: madCard(domain.CardDesignSpade, 1)},
		{PlayerIdx: 2, Card: madCard(domain.CardDesignSpade, 2)},
		{PlayerIdx: 3, Card: madCard(domain.CardDesignSpade, 13)},
	})
	g.ScoreRound()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 0, g.GetWinnerTeam())
	assert.Equal(t, domain.MadrassoPhaseGameEnd, g.GetPhase())
}

func TestMadrassoNoGameEndOnTie(t *testing.T) {
	g := newTestMadrasso()
	g.SetTeamScores([2]int{21, 21})
	g.SetPhase(domain.MadrassoPhaseRoundEnd)
	g.ScoreRound()
	assert.False(t, g.GetGameEndFlag()) // equal scores → continue
}

// --- play flow & must-follow ---

func TestMadrassoPlayerPlayMustFollow(t *testing.T) {
	g := newTestMadrasso()
	g.SetPhase(domain.MadrassoPhasePlay)
	g.SetCurrentPlayerIdx(0)
	madSetHand(g.GetPlayer(0), madCard(domain.CardDesignSpade, 1), madCard(domain.CardDesignHeart, 7))
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: madCard(domain.CardDesignSpade, 4)}})

	// Playing the heart while holding a spade is illegal.
	assert.Error(t, g.PlayerPlay(1))
	// Playing the spade is legal.
	assert.NoError(t, g.PlayerPlay(0))
}

func TestMadrassoPlayerPlayLeadAnySuit(t *testing.T) {
	g := newTestMadrasso()
	g.SetPhase(domain.MadrassoPhasePlay)
	g.SetCurrentPlayerIdx(0)
	madSetHand(g.GetPlayer(0), madCard(domain.CardDesignHeart, 7))
	assert.NoError(t, g.PlayerPlay(0)) // leading: any card allowed
	assert.Len(t, g.GetCurrentTrick(), 1)
}

func TestMadrassoPlayerPlayErrors(t *testing.T) {
	g := newTestMadrasso()

	g.SetPhase(domain.MadrassoPhaseTrickEnd)
	assert.ErrorIs(t, g.PlayerPlay(0), domain.ErrWrongPhase)

	g.SetPhase(domain.MadrassoPhasePlay)
	g.SetCurrentPlayerIdx(1) // CPU's turn
	assert.ErrorIs(t, g.PlayerPlay(0), domain.ErrNotHumanTurn)

	g.SetCurrentPlayerIdx(0)
	madSetHand(g.GetPlayer(0), madCard(domain.CardDesignSpade, 1))
	assert.Error(t, g.PlayerPlay(5)) // out of range

	g.SetPhase(domain.MadrassoPhaseGameEnd)
	g.SetCurrentPlayerIdx(0)
	// gameEndFlag drives ErrGameEnded
	g.SetTeamScores([2]int{99, 0})
}

func TestMadrassoPlayCompletesTrick(t *testing.T) {
	g := newTestMadrasso()
	g.SetPhase(domain.MadrassoPhasePlay)
	g.SetCurrentPlayerIdx(0)
	for i := 0; i < 4; i++ {
		madSetHand(g.GetPlayer(i), madCard(domain.CardDesignSpade, 4+i))
	}
	// players 1,2,3 play first via direct trick set, human completes.
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: madCard(domain.CardDesignSpade, 5)},
		{PlayerIdx: 2, Card: madCard(domain.CardDesignSpade, 6)},
		{PlayerIdx: 3, Card: madCard(domain.CardDesignSpade, 7)},
	})
	assert.NoError(t, g.PlayerPlay(0)) // 4th card → TrickEnd
	assert.Equal(t, domain.MadrassoPhaseTrickEnd, g.GetPhase())
}

func TestMadrassoCpuPlay(t *testing.T) {
	g := newTestMadrasso()
	g.SetPhase(domain.MadrassoPhasePlay)
	g.SetCurrentPlayerIdx(1)
	madSetHand(g.GetPlayer(1), madCard(domain.CardDesignSpade, 4), madCard(domain.CardDesignHeart, 7))
	g.CpuPlay()
	assert.Len(t, g.GetCurrentTrick(), 1)
}

func TestMadrassoCpuPlayFollowsAndWins(t *testing.T) {
	// Opponent leads a low spade with a point card; CPU should be able to act.
	cfg := domain.DefaultMadrassoConfig()
	cfg.CpuDifficulty = domain.MadrassoCpuDifficultyHard
	g := newTestMadrasso()
	g.SetConfig(cfg)
	g.SetPhase(domain.MadrassoPhasePlay)
	g.SetCurrentPlayerIdx(1)
	madSetHand(g.GetPlayer(1), madCard(domain.CardDesignSpade, 3), madCard(domain.CardDesignSpade, 4))
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 0, Card: madCard(domain.CardDesignSpade, 1)}})
	g.CpuPlay()
	assert.Len(t, g.GetCurrentTrick(), 2)
}

func TestMadrassoCpuPlayEasyRandom(t *testing.T) {
	cfg := domain.DefaultMadrassoConfig()
	cfg.CpuDifficulty = domain.MadrassoCpuDifficultyEasy
	g := newTestMadrasso()
	g.SetConfig(cfg)
	g.SetPhase(domain.MadrassoPhasePlay)
	g.SetCurrentPlayerIdx(1)
	madSetHand(g.GetPlayer(1), madCard(domain.CardDesignSpade, 4), madCard(domain.CardDesignSpade, 5))
	g.CpuPlay()
	assert.Equal(t, 1, g.GetPlayer(1).GetCardsSize())
}

// --- next trick / next round ---

func TestMadrassoNextTrick(t *testing.T) {
	g := newTestMadrasso()
	g.SetPhase(domain.MadrassoPhaseTrickEnd)
	g.SetLeadPlayerIdx(2)
	g.SetTrickNumber(3)
	g.NextTrick()
	assert.Equal(t, domain.MadrassoPhasePlay, g.GetPhase())
	assert.Equal(t, 2, g.GetCurrentPlayerIdx())
	assert.Equal(t, 4, g.GetTrickNumber())
	assert.Empty(t, g.GetCurrentTrick())
}

func TestMadrassoNextTrickWrongPhase(t *testing.T) {
	g := newTestMadrasso()
	g.SetPhase(domain.MadrassoPhasePlay)
	g.NextTrick()
	assert.Equal(t, domain.MadrassoPhasePlay, g.GetPhase())
}

func TestMadrassoNextRound(t *testing.T) {
	g := newTestMadrasso()
	g.SetRoundNumber(1)
	g.SetPhase(domain.MadrassoPhaseRoundEnd)
	g.NextRound()
	assert.Equal(t, 2, g.GetRoundNumber())
	assert.Equal(t, domain.MadrassoPhasePlay, g.GetPhase())
	for i := 0; i < 4; i++ {
		assert.Equal(t, 10, g.GetPlayer(i).GetCardsSize())
	}
}

func TestMadrassoNextRoundWrongPhase(t *testing.T) {
	g := newTestMadrasso()
	g.SetRoundNumber(1)
	g.SetPhase(domain.MadrassoPhasePlay)
	g.NextRound()
	assert.Equal(t, 1, g.GetRoundNumber())
}

// --- hint & playable indices ---

func TestMadrassoHintLead(t *testing.T) {
	g := newTestMadrasso()
	g.SetPhase(domain.MadrassoPhasePlay)
	g.SetCurrentPlayerIdx(0)
	madSetHand(g.GetPlayer(0), madCard(domain.CardDesignSpade, 4), madCard(domain.CardDesignHeart, 1))
	hint := g.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "lead_low", hint.Reason)
	assert.Len(t, hint.CardIndices, 1)
}

func TestMadrassoHintNilWhenNotHumanTurn(t *testing.T) {
	g := newTestMadrasso()
	g.SetPhase(domain.MadrassoPhasePlay)
	g.SetCurrentPlayerIdx(1)
	assert.Nil(t, g.GetHint())
}

func TestMadrassoHintFollowReasons(t *testing.T) {
	g := newTestMadrasso()
	g.SetPhase(domain.MadrassoPhasePlay)
	g.SetCurrentPlayerIdx(0)
	madSetHand(g.GetPlayer(0), madCard(domain.CardDesignSpade, 1), madCard(domain.CardDesignSpade, 4))

	// Opponent (player 3, team B) leads ♠K (a point card, strength 6) and the
	// human holds ♠A (strength 7) → worth winning the points.
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: madCard(domain.CardDesignSpade, 13)}})
	assert.Equal(t, "follow_win", g.GetHint().Reason)

	// Partner (player 2, team A) is winning → give partner points.
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 2, Card: madCard(domain.CardDesignSpade, 3)}})
	assert.Equal(t, "give_partner", g.GetHint().Reason)
}

func TestMadrassoHintDiscard(t *testing.T) {
	g := newTestMadrasso()
	g.SetPhase(domain.MadrassoPhasePlay)
	g.SetCurrentPlayerIdx(0)
	madSetHand(g.GetPlayer(0), madCard(domain.CardDesignHeart, 4))
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: madCard(domain.CardDesignSpade, 5)}})
	assert.Equal(t, "discard_low", g.GetHint().Reason)
}

func TestMadrassoPlayableIndices(t *testing.T) {
	g := newTestMadrasso()
	g.SetPhase(domain.MadrassoPhasePlay)
	g.SetCurrentPlayerIdx(0)
	madSetHand(g.GetPlayer(0), madCard(domain.CardDesignSpade, 1), madCard(domain.CardDesignHeart, 7))
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: madCard(domain.CardDesignSpade, 4)}})
	// Only the spade (index 0) is playable.
	assert.Equal(t, []int{0}, g.GetPlayableIndices(0))
	assert.Nil(t, g.GetPlayableIndices(99))
}

// --- JSON ---

func TestMadrassoJSONRoundTrip(t *testing.T) {
	g := newTestMadrasso()
	g.Reset()
	g.SetTeamScores([2]int{5, 3})
	data, err := json.Marshal(g)
	assert.NoError(t, err)

	var restored domain.Madrasso
	assert.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, g.GetPhase(), restored.GetPhase())
	assert.Equal(t, g.GetRoundNumber(), restored.GetRoundNumber())
	assert.Equal(t, g.GetPlayerCnt(), restored.GetPlayerCnt())
	assert.Equal(t, [2]int{5, 3}, restored.GetTeamScores())
}

func TestMadrassoUnmarshalInvalid(t *testing.T) {
	var g domain.Madrasso
	assert.Error(t, json.Unmarshal([]byte("not json"), &g))
}

// TestMadrassoTrumpComesFromTheDeal は、配り終えた時点で切り札が決まって
// いることを見る。
//
// クローン元のトレセッテに切り札は無いので、Reset がこれを設定し忘れても
// 「切り札 -1」のまま静かに進み、madrassoBeats はどの札も切り札と見なさない。
func TestMadrassoTrumpComesFromTheDeal(t *testing.T) {
	seen := map[int]bool{}
	for i := 0; i < 40; i++ {
		g := domain.NewDefaultMadrasso()
		g.Reset()
		trump := g.GetTrumpSuit()

		// **本物のスートであること。** -1 のままだと切り札が機能しない。
		require.GreaterOrEqual(t, trump, domain.CardDesignSpade, "切り札が未確定のまま")
		require.LessOrEqual(t, trump, domain.CardDesignMax)

		// 最後に配られた札は最後の席に入るので、その席が切り札スートを持つ。
		last := g.GetPlayer(domain.MadrassoPlayerCnt - 1)
		held := false
		for j := 0; j < last.GetCardsSize(); j++ {
			if c := last.GetCard(j); c != nil && c.GetDesign() == trump {
				held = true
				break
			}
		}
		assert.True(t, held, "切り札スートが最後の席の手札に無い (配りから取っていない)")

		seen[trump] = true
	}

	// **固定値になっていないこと。** 定数を返すだけの実装でも上は通る。
	assert.Greater(t, len(seen), 1, "40 回配って切り札が %d 種類しか出ていない", len(seen))
}
