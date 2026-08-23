//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestTrappola() *domain.Trappola {
	players := []*domain.TrappolaPlayer{
		domain.NewTrappolaPlayer(true),  // 0 = human (team A)
		domain.NewTrappolaPlayer(false), // 1 = CPU  (team B)
		domain.NewTrappolaPlayer(false), // 2 = CPU  (team A)
		domain.NewTrappolaPlayer(false), // 3 = CPU  (team B)
	}
	return domain.NewTrappola(domain.NewTrumpCardsBriscola(), players, domain.DefaultTrappolaConfig())
}

func trapCard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

func trapSetHand(p *domain.TrappolaPlayer, cards ...*domain.Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// trapResolve sets up a complete trick at trickNum and resolves it.
func trapResolve(g *domain.Trappola, trickNum int, trick []*domain.TrickCard) {
	g.SetTrickNumber(trickNum)
	g.SetCurrentTrick(trick)
	g.SetPhase(domain.TrappolaPhaseTrickEnd)
	g.ResolveTrick()
}

// --- construction ---

func TestNewTrappola(t *testing.T) {
	g := newTestTrappola()
	assert.Equal(t, domain.TrappolaPhase(0), g.GetPhase())
	assert.Equal(t, 0, g.GetRoundNumber())
	assert.Equal(t, 4, g.GetPlayerCnt())
	assert.Equal(t, -1, g.GetWinnerTeam())
	assert.False(t, g.GetGameEndFlag())
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(99))
}

func TestNewDefaultTrappola(t *testing.T) {
	g := domain.NewDefaultTrappola()
	assert.Equal(t, 4, g.GetPlayerCnt())
	assert.True(t, g.GetPlayer(0).GetIsHuman())
	assert.False(t, g.GetPlayer(1).GetIsHuman())
}

func TestTrappolaTeamOf(t *testing.T) {
	assert.Equal(t, 0, domain.TrappolaTeamOf(0))
	assert.Equal(t, 1, domain.TrappolaTeamOf(1))
	assert.Equal(t, 0, domain.TrappolaTeamOf(2))
	assert.Equal(t, 1, domain.TrappolaTeamOf(3))
}

func TestTrappolaReset(t *testing.T) {
	g := newTestTrappola()
	g.Reset()
	assert.Equal(t, domain.TrappolaPhasePlay, g.GetPhase())
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

func TestTrappolaConfigValidate(t *testing.T) {
	assert.NoError(t, domain.DefaultTrappolaConfig().Validate())
	assert.Error(t, domain.TrappolaConfig{CpuDifficulty: 99, TargetPoints: 21}.Validate())
	assert.Error(t, domain.TrappolaConfig{CpuDifficulty: domain.TrappolaCpuDifficultyEasy, TargetPoints: 0}.Validate())
}

// --- trick resolution & strength ---

func TestTrappolaTrickWinnerHighestStrength(t *testing.T) {
	g := newTestTrappola()
	// **A が最強、3 が最弱。** クローン元のトレセッテは 3 > 2 > A だったので、
	// 同じ配りでも勝者が変わる。全部スペードなので A を出した席 0 が勝つ。
	trapResolve(g, 1, []*domain.TrickCard{
		{PlayerIdx: 0, Card: trapCard(domain.CardDesignSpade, 1)},  // A (最強)
		{PlayerIdx: 1, Card: trapCard(domain.CardDesignSpade, 4)},  // 4
		{PlayerIdx: 2, Card: trapCard(domain.CardDesignSpade, 3)},  // 3 (最弱)
		{PlayerIdx: 3, Card: trapCard(domain.CardDesignSpade, 13)}, // K
	})
	// 勝者 (席 0 = チーム A) の取り分: A(3) + 4(0) + 3(0) + K(1) = 4 thirds。
	thirds := g.GetTeamRoundThirds()
	assert.Equal(t, 4, thirds[0])
	assert.Equal(t, 0, thirds[1])
	assert.Equal(t, 0, g.GetLeadPlayerIdx(), "A を出した席が勝つ (3 ではない)")
	assert.Equal(t, domain.TrappolaPhaseTrickEnd, g.GetPhase())
}

// TestTrappolaRankOrderIsAceHigh は札位の順を**全ペアで**突き合わせる。
//
// 切り札が無いので、この 1 本がトリックの勝敗を全部決める。クローン元の
// トレセッテは 3 > 2 > A > K > Q > J > 7 > 6 > 5 > 4 という別の順を持つ。
func TestTrappolaRankOrderIsAceHigh(t *testing.T) {
	// 強い順。デッキに実在する札位だけを並べる。
	ranked := []int{1, 13, 12, 11, 7, 6, 5, 4, 3}
	require.Len(t, ranked, len(domain.TrappolaValues), "デッキの札位を全部並べること")
	for _, v := range ranked {
		require.Contains(t, domain.TrappolaValues, v, "札位 %d はデッキに無い", v)
	}
	for i := 1; i < len(ranked); i++ {
		assert.Greater(t, domain.TrappolaStrengthForTest(ranked[i-1]), domain.TrappolaStrengthForTest(ranked[i]),
			"%d は %d より強いこと", ranked[i-1], ranked[i])
	}
	// **3 は最弱。** クローン元では最強だった。
	for _, v := range ranked[:len(ranked)-1] {
		assert.Greater(t, domain.TrappolaStrengthForTest(v), domain.TrappolaStrengthForTest(3),
			"3 が %d より強い (クローン元の順が残っている)", v)
	}
}

func TestTrappolaOffsuitDoesNotWin(t *testing.T) {
	g := newTestTrappola()
	// Lead spade; player 1 throws a high heart that cannot win (no trump).
	trapResolve(g, 1, []*domain.TrickCard{
		{PlayerIdx: 0, Card: trapCard(domain.CardDesignSpade, 5)},
		{PlayerIdx: 1, Card: trapCard(domain.CardDesignHeart, 3)}, // off-suit
		{PlayerIdx: 2, Card: trapCard(domain.CardDesignSpade, 6)},
		{PlayerIdx: 3, Card: trapCard(domain.CardDesignSpade, 4)},
	})
	// Highest spade is 6 (player 2, team A).
	assert.Equal(t, 2, g.GetLeadPlayerIdx())
}

func TestTrappolaUltimaBonus(t *testing.T) {
	g := newTestTrappola()
	// **最終トリックは手札枚数と同じ 9。** クローン元は 10 枚配りだった。
	trapResolve(g, domain.TrappolaTrickCount, []*domain.TrickCard{
		{PlayerIdx: 0, Card: trapCard(domain.CardDesignSpade, 3)}, // 3 (最弱)
		{PlayerIdx: 1, Card: trapCard(domain.CardDesignSpade, 1)}, // A (最強) — チーム B
		{PlayerIdx: 2, Card: trapCard(domain.CardDesignSpade, 4)},
		{PlayerIdx: 3, Card: trapCard(domain.CardDesignSpade, 13)}, // K
	})
	// 勝者はチーム B: A(3) + 3(0) + 4(0) + K(1) = 4 thirds + ウルティマ 1 = 5。
	thirds := g.GetTeamRoundThirds()
	assert.Equal(t, 5, thirds[1], "ウルティマが乗ること")
	assert.Equal(t, 0, thirds[0])
	assert.Equal(t, domain.TrappolaPhaseRoundEnd, g.GetPhase())
}

// TestTrappolaUltimaFiresOnTheLastTrick は、**最終トリックでだけ**ボーナスが
// 乗ることを見る。
//
// クローン元の 10 をそのまま残すと 9 トリックのこのゲームでは条件に届かず、
// ウルティマが永久に発火しない (どのテストも「乗らない」まま緑になる)。
func TestTrappolaUltimaFiresOnTheLastTrick(t *testing.T) {
	// 最終の 1 つ手前では乗らない。
	before := newTestTrappola()
	trapResolve(before, domain.TrappolaTrickCount-1, []*domain.TrickCard{
		{PlayerIdx: 0, Card: trapCard(domain.CardDesignSpade, 13)}, // K = 1 third
		{PlayerIdx: 1, Card: trapCard(domain.CardDesignSpade, 3)},
		{PlayerIdx: 2, Card: trapCard(domain.CardDesignSpade, 4)},
		{PlayerIdx: 3, Card: trapCard(domain.CardDesignSpade, 5)},
	})
	assert.Equal(t, 1, before.GetTeamRoundThirds()[0], "最終トリックの前はボーナス無し")

	// 最終トリックでは +1。
	last := newTestTrappola()
	trapResolve(last, domain.TrappolaTrickCount, []*domain.TrickCard{
		{PlayerIdx: 0, Card: trapCard(domain.CardDesignSpade, 13)}, // K = 1 third
		{PlayerIdx: 1, Card: trapCard(domain.CardDesignSpade, 3)},
		{PlayerIdx: 2, Card: trapCard(domain.CardDesignSpade, 4)},
		{PlayerIdx: 3, Card: trapCard(domain.CardDesignSpade, 5)},
	})
	assert.Equal(t, 2, last.GetTeamRoundThirds()[0], "最終トリックでボーナスが乗ること")
}

func TestTrappolaResolveTrickGuards(t *testing.T) {
	g := newTestTrappola()
	g.SetPhase(domain.TrappolaPhasePlay) // wrong phase
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 0, Card: trapCard(domain.CardDesignSpade, 3)}})
	g.ResolveTrick()
	assert.Equal(t, domain.TrappolaPhasePlay, g.GetPhase()) // unchanged
}

// --- scoring ---

func TestTrappolaScoreRoundThirdsToPoints(t *testing.T) {
	g := newTestTrappola()
	// 最終トリックをチーム B (A を出した席 1) が取る: 4 thirds + ウルティマ 1 = 5。
	trapResolve(g, domain.TrappolaTrickCount, []*domain.TrickCard{
		{PlayerIdx: 0, Card: trapCard(domain.CardDesignSpade, 3)},
		{PlayerIdx: 1, Card: trapCard(domain.CardDesignSpade, 1)},
		{PlayerIdx: 2, Card: trapCard(domain.CardDesignSpade, 4)},
		{PlayerIdx: 3, Card: trapCard(domain.CardDesignSpade, 13)},
	})
	g.ScoreRound()
	scores := g.GetTeamScores()
	assert.Equal(t, 0, scores[0])
	assert.Equal(t, 1, scores[1]) // 5/3 = 1 (切り捨て)
	assert.False(t, g.GetGameEndFlag())
}

func TestTrappolaScoreRoundWrongPhaseNoop(t *testing.T) {
	g := newTestTrappola()
	g.SetPhase(domain.TrappolaPhasePlay)
	g.ScoreRound()
	assert.Equal(t, [2]int{0, 0}, g.GetTeamScores())
}

func TestTrappolaGameEnd(t *testing.T) {
	g := newTestTrappola()
	// **A が最強なので、この配りを取るのは席 1 (チーム B)。**
	g.SetTeamScores([2]int{0, 20})
	// チーム B が 5 thirds (= 1 点) を取り 21 ≥ 21 かつ相手より上 → ゲーム終了。
	trapResolve(g, domain.TrappolaTrickCount, []*domain.TrickCard{
		{PlayerIdx: 0, Card: trapCard(domain.CardDesignSpade, 3)},
		{PlayerIdx: 1, Card: trapCard(domain.CardDesignSpade, 1)},
		{PlayerIdx: 2, Card: trapCard(domain.CardDesignSpade, 4)},
		{PlayerIdx: 3, Card: trapCard(domain.CardDesignSpade, 13)},
	})
	g.ScoreRound()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 1, g.GetWinnerTeam())
	assert.Equal(t, domain.TrappolaPhaseGameEnd, g.GetPhase())
}

func TestTrappolaNoGameEndOnTie(t *testing.T) {
	g := newTestTrappola()
	g.SetTeamScores([2]int{21, 21})
	g.SetPhase(domain.TrappolaPhaseRoundEnd)
	g.ScoreRound()
	assert.False(t, g.GetGameEndFlag()) // equal scores → continue
}

// --- play flow & must-follow ---

func TestTrappolaPlayerPlayMustFollow(t *testing.T) {
	g := newTestTrappola()
	g.SetPhase(domain.TrappolaPhasePlay)
	g.SetCurrentPlayerIdx(0)
	trapSetHand(g.GetPlayer(0), trapCard(domain.CardDesignSpade, 1), trapCard(domain.CardDesignHeart, 7))
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: trapCard(domain.CardDesignSpade, 4)}})

	// Playing the heart while holding a spade is illegal.
	assert.Error(t, g.PlayerPlay(1))
	// Playing the spade is legal.
	assert.NoError(t, g.PlayerPlay(0))
}

func TestTrappolaPlayerPlayLeadAnySuit(t *testing.T) {
	g := newTestTrappola()
	g.SetPhase(domain.TrappolaPhasePlay)
	g.SetCurrentPlayerIdx(0)
	trapSetHand(g.GetPlayer(0), trapCard(domain.CardDesignHeart, 7))
	assert.NoError(t, g.PlayerPlay(0)) // leading: any card allowed
	assert.Len(t, g.GetCurrentTrick(), 1)
}

func TestTrappolaPlayerPlayErrors(t *testing.T) {
	g := newTestTrappola()

	g.SetPhase(domain.TrappolaPhaseTrickEnd)
	assert.ErrorIs(t, g.PlayerPlay(0), domain.ErrWrongPhase)

	g.SetPhase(domain.TrappolaPhasePlay)
	g.SetCurrentPlayerIdx(1) // CPU's turn
	assert.ErrorIs(t, g.PlayerPlay(0), domain.ErrNotHumanTurn)

	g.SetCurrentPlayerIdx(0)
	trapSetHand(g.GetPlayer(0), trapCard(domain.CardDesignSpade, 1))
	assert.Error(t, g.PlayerPlay(5)) // out of range

	g.SetPhase(domain.TrappolaPhaseGameEnd)
	g.SetCurrentPlayerIdx(0)
	// gameEndFlag drives ErrGameEnded
	g.SetTeamScores([2]int{99, 0})
}

func TestTrappolaPlayCompletesTrick(t *testing.T) {
	g := newTestTrappola()
	g.SetPhase(domain.TrappolaPhasePlay)
	g.SetCurrentPlayerIdx(0)
	for i := 0; i < 4; i++ {
		trapSetHand(g.GetPlayer(i), trapCard(domain.CardDesignSpade, 4+i))
	}
	// players 1,2,3 play first via direct trick set, human completes.
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: trapCard(domain.CardDesignSpade, 5)},
		{PlayerIdx: 2, Card: trapCard(domain.CardDesignSpade, 6)},
		{PlayerIdx: 3, Card: trapCard(domain.CardDesignSpade, 7)},
	})
	assert.NoError(t, g.PlayerPlay(0)) // 4th card → TrickEnd
	assert.Equal(t, domain.TrappolaPhaseTrickEnd, g.GetPhase())
}

func TestTrappolaCpuPlay(t *testing.T) {
	g := newTestTrappola()
	g.SetPhase(domain.TrappolaPhasePlay)
	g.SetCurrentPlayerIdx(1)
	trapSetHand(g.GetPlayer(1), trapCard(domain.CardDesignSpade, 4), trapCard(domain.CardDesignHeart, 7))
	g.CpuPlay()
	assert.Len(t, g.GetCurrentTrick(), 1)
}

func TestTrappolaCpuPlayFollowsAndWins(t *testing.T) {
	// Opponent leads a low spade with a point card; CPU should be able to act.
	cfg := domain.DefaultTrappolaConfig()
	cfg.CpuDifficulty = domain.TrappolaCpuDifficultyHard
	g := newTestTrappola()
	g.SetConfig(cfg)
	g.SetPhase(domain.TrappolaPhasePlay)
	g.SetCurrentPlayerIdx(1)
	trapSetHand(g.GetPlayer(1), trapCard(domain.CardDesignSpade, 3), trapCard(domain.CardDesignSpade, 4))
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 0, Card: trapCard(domain.CardDesignSpade, 1)}})
	g.CpuPlay()
	assert.Len(t, g.GetCurrentTrick(), 2)
}

func TestTrappolaCpuPlayEasyRandom(t *testing.T) {
	cfg := domain.DefaultTrappolaConfig()
	cfg.CpuDifficulty = domain.TrappolaCpuDifficultyEasy
	g := newTestTrappola()
	g.SetConfig(cfg)
	g.SetPhase(domain.TrappolaPhasePlay)
	g.SetCurrentPlayerIdx(1)
	trapSetHand(g.GetPlayer(1), trapCard(domain.CardDesignSpade, 4), trapCard(domain.CardDesignSpade, 5))
	g.CpuPlay()
	assert.Equal(t, 1, g.GetPlayer(1).GetCardsSize())
}

// --- next trick / next round ---

func TestTrappolaNextTrick(t *testing.T) {
	g := newTestTrappola()
	g.SetPhase(domain.TrappolaPhaseTrickEnd)
	g.SetLeadPlayerIdx(2)
	g.SetTrickNumber(3)
	g.NextTrick()
	assert.Equal(t, domain.TrappolaPhasePlay, g.GetPhase())
	assert.Equal(t, 2, g.GetCurrentPlayerIdx())
	assert.Equal(t, 4, g.GetTrickNumber())
	assert.Empty(t, g.GetCurrentTrick())
}

func TestTrappolaNextTrickWrongPhase(t *testing.T) {
	g := newTestTrappola()
	g.SetPhase(domain.TrappolaPhasePlay)
	g.NextTrick()
	assert.Equal(t, domain.TrappolaPhasePlay, g.GetPhase())
}

func TestTrappolaNextRound(t *testing.T) {
	g := newTestTrappola()
	g.SetRoundNumber(1)
	g.SetPhase(domain.TrappolaPhaseRoundEnd)
	g.NextRound()
	assert.Equal(t, 2, g.GetRoundNumber())
	assert.Equal(t, domain.TrappolaPhasePlay, g.GetPhase())
	for i := 0; i < 4; i++ {
		assert.Equal(t, 10, g.GetPlayer(i).GetCardsSize())
	}
}

func TestTrappolaNextRoundWrongPhase(t *testing.T) {
	g := newTestTrappola()
	g.SetRoundNumber(1)
	g.SetPhase(domain.TrappolaPhasePlay)
	g.NextRound()
	assert.Equal(t, 1, g.GetRoundNumber())
}

// --- hint & playable indices ---

func TestTrappolaHintLead(t *testing.T) {
	g := newTestTrappola()
	g.SetPhase(domain.TrappolaPhasePlay)
	g.SetCurrentPlayerIdx(0)
	trapSetHand(g.GetPlayer(0), trapCard(domain.CardDesignSpade, 4), trapCard(domain.CardDesignHeart, 1))
	hint := g.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "lead_low", hint.Reason)
	assert.Len(t, hint.CardIndices, 1)
}

func TestTrappolaHintNilWhenNotHumanTurn(t *testing.T) {
	g := newTestTrappola()
	g.SetPhase(domain.TrappolaPhasePlay)
	g.SetCurrentPlayerIdx(1)
	assert.Nil(t, g.GetHint())
}

func TestTrappolaHintFollowReasons(t *testing.T) {
	g := newTestTrappola()
	g.SetPhase(domain.TrappolaPhasePlay)
	g.SetCurrentPlayerIdx(0)
	trapSetHand(g.GetPlayer(0), trapCard(domain.CardDesignSpade, 1), trapCard(domain.CardDesignSpade, 4))

	// Opponent (player 3, team B) leads ♠K (a point card, strength 6) and the
	// human holds ♠A (strength 7) → worth winning the points.
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: trapCard(domain.CardDesignSpade, 13)}})
	assert.Equal(t, "follow_win", g.GetHint().Reason)

	// **相方が勝っている盤面を作るには、相方に強い札を出させる。**
	// 3 はこのゲームでは最弱なので、それでは相方は勝っていない
	// (クローン元では 3 が最強だった)。K を出させる。
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 2, Card: trapCard(domain.CardDesignSpade, 13)}})
	assert.Equal(t, "give_partner", g.GetHint().Reason)
}

func TestTrappolaHintDiscard(t *testing.T) {
	g := newTestTrappola()
	g.SetPhase(domain.TrappolaPhasePlay)
	g.SetCurrentPlayerIdx(0)
	trapSetHand(g.GetPlayer(0), trapCard(domain.CardDesignHeart, 4))
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: trapCard(domain.CardDesignSpade, 5)}})
	assert.Equal(t, "discard_low", g.GetHint().Reason)
}

func TestTrappolaPlayableIndices(t *testing.T) {
	g := newTestTrappola()
	g.SetPhase(domain.TrappolaPhasePlay)
	g.SetCurrentPlayerIdx(0)
	trapSetHand(g.GetPlayer(0), trapCard(domain.CardDesignSpade, 1), trapCard(domain.CardDesignHeart, 7))
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: trapCard(domain.CardDesignSpade, 4)}})
	// Only the spade (index 0) is playable.
	assert.Equal(t, []int{0}, g.GetPlayableIndices(0))
	assert.Nil(t, g.GetPlayableIndices(99))
}

// --- JSON ---

func TestTrappolaJSONRoundTrip(t *testing.T) {
	g := newTestTrappola()
	g.Reset()
	g.SetTeamScores([2]int{5, 3})
	data, err := json.Marshal(g)
	assert.NoError(t, err)

	var restored domain.Trappola
	assert.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, g.GetPhase(), restored.GetPhase())
	assert.Equal(t, g.GetRoundNumber(), restored.GetRoundNumber())
	assert.Equal(t, g.GetPlayerCnt(), restored.GetPlayerCnt())
	assert.Equal(t, [2]int{5, 3}, restored.GetTeamScores())
}

func TestTrappolaUnmarshalInvalid(t *testing.T) {
	var g domain.Trappola
	assert.Error(t, json.Unmarshal([]byte("not json"), &g))
}

// TestTrappolaRoundThirdsMatchesTheDeck は、1 ラウンドで奪い合う点の総和が
// **デッキから数えた値と一致する**ことを見る。
//
// 手で書いた定数は必ずずれる (Bauernschnapsen で 132 と書いて実際は 120
// だった)。デッキを走査して足し、最終トリックのボーナスを加えて突き合わせる。
func TestTrappolaRoundThirdsMatchesTheDeck(t *testing.T) {
	deck := domain.NewTrumpCardsTrappola()
	total := 0
	for i := 0; i < deck.GetTotalCount(); i++ {
		c := deck.DrawCard()
		require.NotNil(t, c)
		total += domain.TrappolaThirdsForTest(c.GetValue())
	}

	// **0 でないこと。** 全札 0 点なら下の比較はどちらも 1 (ボーナスのみ) で
	// 通ってしまう。
	require.Positive(t, total, "カード点が 1 つも無い")
	assert.Equal(t, domain.TrappolaRoundThirds, total+domain.TrappolaUltimaThirds,
		"カード点の合計 %d + ウルティマ %d が定数 %d と合わない",
		total, domain.TrappolaUltimaThirds, domain.TrappolaRoundThirds)

	// 2 はこのデッキに無いので、点を持っていても誰も取れない。
	assert.Equal(t, 0, domain.TrappolaThirdsForTest(2), "2 に点が付いている (デッキには無い札)")
}

// TestTrappolaDeclarations は配った時点で成立する役を、手札を組んで見る。
func TestTrappolaDeclarations(t *testing.T) {
	// 四つ揃い (A) — 三つ揃いと二重に数えない。
	t.Run("four of a kind is not also counted as three", func(t *testing.T) {
		g := newTestTrappola()
		trapSetHand(g.GetPlayer(0),
			trapCard(domain.CardDesignSpade, 1), trapCard(domain.CardDesignClover, 1),
			trapCard(domain.CardDesignHeart, 1), trapCard(domain.CardDesignDiamond, 1),
			trapCard(domain.CardDesignSpade, 5))
		ds := domain.TrappolaFindDeclarationsForTest(0, g.GetPlayer(0))
		require.Len(t, ds, 1, "四つ揃いと三つ揃いの両方が出ている: %+v", ds)
		assert.Equal(t, domain.TrappolaDeclarationFour, ds[0].Kind)
		assert.Equal(t, domain.TrappolaFourThirds, ds[0].Thirds)
	})

	// トラッポラ = 同スートの A+K+Q。
	t.Run("trappola needs all three of one suit", func(t *testing.T) {
		g := newTestTrappola()
		trapSetHand(g.GetPlayer(0),
			trapCard(domain.CardDesignHeart, 1), trapCard(domain.CardDesignHeart, 13),
			trapCard(domain.CardDesignHeart, 12), trapCard(domain.CardDesignSpade, 5))
		ds := domain.TrappolaFindDeclarationsForTest(0, g.GetPlayer(0))
		require.Len(t, ds, 1)
		assert.Equal(t, domain.TrappolaDeclarationTrappola, ds[0].Kind)
		assert.Equal(t, domain.CardDesignHeart, ds[0].Value)

		// **負のコントロール。** スートが割れていれば成立しない。
		trapSetHand(g.GetPlayer(0),
			trapCard(domain.CardDesignHeart, 1), trapCard(domain.CardDesignHeart, 13),
			trapCard(domain.CardDesignSpade, 12), trapCard(domain.CardDesignSpade, 5))
		assert.Empty(t, domain.TrappolaFindDeclarationsForTest(0, g.GetPlayer(0)),
			"スートが割れているのに成立している")
	})

	// 平札の揃いは役にならない。
	t.Run("plain ranks do not declare", func(t *testing.T) {
		g := newTestTrappola()
		trapSetHand(g.GetPlayer(0),
			trapCard(domain.CardDesignSpade, 7), trapCard(domain.CardDesignClover, 7),
			trapCard(domain.CardDesignHeart, 7), trapCard(domain.CardDesignDiamond, 7))
		assert.Empty(t, domain.TrappolaFindDeclarationsForTest(0, g.GetPlayer(0)),
			"平札 7 の四つ揃いが役になっている")
	})

	// 三つ揃いは 3 枚ちょうどで成立する。
	t.Run("three of a kind", func(t *testing.T) {
		g := newTestTrappola()
		trapSetHand(g.GetPlayer(0),
			trapCard(domain.CardDesignSpade, 13), trapCard(domain.CardDesignClover, 13),
			trapCard(domain.CardDesignHeart, 13), trapCard(domain.CardDesignSpade, 5))
		ds := domain.TrappolaFindDeclarationsForTest(0, g.GetPlayer(0))
		require.Len(t, ds, 1)
		assert.Equal(t, domain.TrappolaDeclarationThree, ds[0].Kind)
		assert.Equal(t, domain.TrappolaThreeThirds, ds[0].Thirds)
	})
}

// TestTrappolaDeclarationsScoreTheTeam は、配った時点で役の点がチームに
// 入っていることを**盤面を回して**見る。判定関数を直接叩くだけでは、
// startRound から呼ばれていないことに気づけない。
func TestTrappolaDeclarationsScoreTheTeam(t *testing.T) {
	// 36 枚 9 枚配りでは役が成立しない配りもあるので、複数ラウンド回して
	// 「一度でも役が付いたこと」と「点が動いたこと」を見る。
	seen := false
	for i := 0; i < 40 && !seen; i++ {
		g := domain.NewDefaultTrappola()
		g.Reset()
		ds := g.GetDeclarations()
		if len(ds) == 0 {
			continue
		}
		seen = true
		want := [2]int{}
		for _, d := range ds {
			want[domain.TrappolaTeamOf(d.PlayerIdx)] += d.Thirds
		}
		assert.Equal(t, want, g.GetTeamRoundThirds(),
			"役の点がチームに入っていない: %+v", ds)
	}
	require.True(t, seen, "40 ラウンド回して役が 1 つも出なかった —— 判定が死んでいる")
}

// TestTrappolaUnmarshalFallbackDeck は、デッキが欠けた JSON から復元したとき
// **コンストラクタと同じ 36 枚デッキ**が作られることを見る。
//
// クローン元の 40 枚ブリスコラデッキが残っていると 2 が混じる。この 36 枚に
// 2 は無いので、強さも点も既定の枝に落ちて静かにずれる (エラーにはならない)。
func TestTrappolaUnmarshalFallbackDeck(t *testing.T) {
	const okPlayers = `[{"gp":{},"th":{}},{"gp":{},"th":{}},{"gp":{},"th":{}},{"gp":{},"th":{}}]`
	blob := `{"ph":0,"ps":` + okPlayers + `,"cf":{"cd":1,"tp":21},"lt":-1,"wt":-1,"li":0,"tc":null}`

	var g domain.Trappola
	require.NoError(t, json.Unmarshal([]byte(blob), &g))

	deck := g.GetDeckForTest()
	require.NotNil(t, deck)
	assert.Equal(t, 36, deck.GetTotalCount(), "40 枚のクローン元デッキが作られている")

	// **2 が入っていないこと。** 枚数が合っていても札位が違えば同じ壊れ方をする。
	for i := 0; i < deck.GetTotalCount(); i++ {
		c := deck.DrawCard()
		require.NotNil(t, c)
		assert.NotEqual(t, 2, c.GetValue(), "このデッキに 2 は無い")
	}
}
