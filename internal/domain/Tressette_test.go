//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestTressette() *domain.Tressette {
	players := []*domain.TressettePlayer{
		domain.NewTressettePlayer(true),  // 0 = human (team A)
		domain.NewTressettePlayer(false), // 1 = CPU  (team B)
		domain.NewTressettePlayer(false), // 2 = CPU  (team A)
		domain.NewTressettePlayer(false), // 3 = CPU  (team B)
	}
	return domain.NewTressette(domain.NewTrumpCardsBriscola(), players, domain.DefaultTressetteConfig())
}

func trCard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

func trSetHand(p *domain.TressettePlayer, cards ...*domain.Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// trResolve sets up a complete trick at trickNum and resolves it.
func trResolve(g *domain.Tressette, trickNum int, trick []*domain.TrickCard) {
	g.SetTrickNumber(trickNum)
	g.SetCurrentTrick(trick)
	g.SetPhase(domain.TressettePhaseTrickEnd)
	g.ResolveTrick()
}

// --- construction ---

func TestNewTressette(t *testing.T) {
	g := newTestTressette()
	assert.Equal(t, domain.TressettePhase(0), g.GetPhase())
	assert.Equal(t, 0, g.GetRoundNumber())
	assert.Equal(t, 4, g.GetPlayerCnt())
	assert.Equal(t, -1, g.GetWinnerTeam())
	assert.False(t, g.GetGameEndFlag())
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(99))
}

func TestNewDefaultTressette(t *testing.T) {
	g := domain.NewDefaultTressette()
	assert.Equal(t, 4, g.GetPlayerCnt())
	assert.True(t, g.GetPlayer(0).GetIsHuman())
	assert.False(t, g.GetPlayer(1).GetIsHuman())
}

func TestTressetteTeamOf(t *testing.T) {
	assert.Equal(t, 0, domain.TressetteTeamOf(0))
	assert.Equal(t, 1, domain.TressetteTeamOf(1))
	assert.Equal(t, 0, domain.TressetteTeamOf(2))
	assert.Equal(t, 1, domain.TressetteTeamOf(3))
}

func TestTressetteReset(t *testing.T) {
	g := newTestTressette()
	g.Reset()
	assert.Equal(t, domain.TressettePhasePlay, g.GetPhase())
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

func TestTressetteConfigValidate(t *testing.T) {
	assert.NoError(t, domain.DefaultTressetteConfig().Validate())
	assert.Error(t, domain.TressetteConfig{CpuDifficulty: 99, TargetPoints: 21}.Validate())
	assert.Error(t, domain.TressetteConfig{CpuDifficulty: domain.TressetteCpuDifficultyEasy, TargetPoints: 0}.Validate())
}

// --- trick resolution & strength ---

func TestTressetteTrickWinnerHighestStrength(t *testing.T) {
	g := newTestTressette()
	// All spades; 3 (strength 9) is the strongest, played by player 2 (team A).
	trResolve(g, 1, []*domain.TrickCard{
		{PlayerIdx: 0, Card: trCard(domain.CardDesignSpade, 1)},  // A
		{PlayerIdx: 1, Card: trCard(domain.CardDesignSpade, 2)},  // 2
		{PlayerIdx: 2, Card: trCard(domain.CardDesignSpade, 3)},  // 3 (strongest)
		{PlayerIdx: 3, Card: trCard(domain.CardDesignSpade, 13)}, // K
	})
	// Winner team A: A(3) + 2(1) + 3(1) + K(1) = 6 thirds.
	thirds := g.GetTeamRoundThirds()
	assert.Equal(t, 6, thirds[0])
	assert.Equal(t, 0, thirds[1])
	assert.Equal(t, 2, g.GetLeadPlayerIdx())
	assert.Equal(t, domain.TressettePhaseTrickEnd, g.GetPhase())
}

func TestTressetteOffsuitDoesNotWin(t *testing.T) {
	g := newTestTressette()
	// Lead spade; player 1 throws a high heart that cannot win (no trump).
	trResolve(g, 1, []*domain.TrickCard{
		{PlayerIdx: 0, Card: trCard(domain.CardDesignSpade, 5)},
		{PlayerIdx: 1, Card: trCard(domain.CardDesignHeart, 3)}, // off-suit
		{PlayerIdx: 2, Card: trCard(domain.CardDesignSpade, 6)},
		{PlayerIdx: 3, Card: trCard(domain.CardDesignSpade, 4)},
	})
	// Highest spade is 6 (player 2, team A).
	assert.Equal(t, 2, g.GetLeadPlayerIdx())
}

func TestTressetteUltimaBonus(t *testing.T) {
	g := newTestTressette()
	trResolve(g, 10, []*domain.TrickCard{
		{PlayerIdx: 0, Card: trCard(domain.CardDesignSpade, 3)}, // strongest, team A
		{PlayerIdx: 1, Card: trCard(domain.CardDesignSpade, 1)}, // A
		{PlayerIdx: 2, Card: trCard(domain.CardDesignSpade, 2)},
		{PlayerIdx: 3, Card: trCard(domain.CardDesignSpade, 13)},
	})
	// A(3)+3(1)+2(1)+K(1)=6 thirds + 1 ultima = 7.
	thirds := g.GetTeamRoundThirds()
	assert.Equal(t, 7, thirds[0])
	assert.Equal(t, domain.TressettePhaseRoundEnd, g.GetPhase())
}

func TestTressetteResolveTrickGuards(t *testing.T) {
	g := newTestTressette()
	g.SetPhase(domain.TressettePhasePlay) // wrong phase
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 0, Card: trCard(domain.CardDesignSpade, 3)}})
	g.ResolveTrick()
	assert.Equal(t, domain.TressettePhasePlay, g.GetPhase()) // unchanged
}

// --- scoring ---

func TestTressetteScoreRoundThirdsToPoints(t *testing.T) {
	g := newTestTressette()
	// Team A wins a trick worth 7 thirds at trick 10 (incl. ultima) → RoundEnd.
	trResolve(g, 10, []*domain.TrickCard{
		{PlayerIdx: 0, Card: trCard(domain.CardDesignSpade, 3)},
		{PlayerIdx: 1, Card: trCard(domain.CardDesignSpade, 1)},
		{PlayerIdx: 2, Card: trCard(domain.CardDesignSpade, 2)},
		{PlayerIdx: 3, Card: trCard(domain.CardDesignSpade, 13)},
	})
	g.ScoreRound()
	scores := g.GetTeamScores()
	assert.Equal(t, 2, scores[0]) // 7/3 = 2 (truncated)
	assert.Equal(t, 0, scores[1])
	assert.False(t, g.GetGameEndFlag())
}

func TestTressetteScoreRoundWrongPhaseNoop(t *testing.T) {
	g := newTestTressette()
	g.SetPhase(domain.TressettePhasePlay)
	g.ScoreRound()
	assert.Equal(t, [2]int{0, 0}, g.GetTeamScores())
}

func TestTressetteGameEnd(t *testing.T) {
	g := newTestTressette()
	g.SetTeamScores([2]int{20, 0})
	// Team A wins 7 thirds (=2 pts) at trick 10 → 22 ≥ 21 and > 0 → game end.
	trResolve(g, 10, []*domain.TrickCard{
		{PlayerIdx: 0, Card: trCard(domain.CardDesignSpade, 3)},
		{PlayerIdx: 1, Card: trCard(domain.CardDesignSpade, 1)},
		{PlayerIdx: 2, Card: trCard(domain.CardDesignSpade, 2)},
		{PlayerIdx: 3, Card: trCard(domain.CardDesignSpade, 13)},
	})
	g.ScoreRound()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 0, g.GetWinnerTeam())
	assert.Equal(t, domain.TressettePhaseGameEnd, g.GetPhase())
}

func TestTressetteNoGameEndOnTie(t *testing.T) {
	g := newTestTressette()
	g.SetTeamScores([2]int{21, 21})
	g.SetPhase(domain.TressettePhaseRoundEnd)
	g.ScoreRound()
	assert.False(t, g.GetGameEndFlag()) // equal scores → continue
}

// --- play flow & must-follow ---

func TestTressettePlayerPlayMustFollow(t *testing.T) {
	g := newTestTressette()
	g.SetPhase(domain.TressettePhasePlay)
	g.SetCurrentPlayerIdx(0)
	trSetHand(g.GetPlayer(0), trCard(domain.CardDesignSpade, 1), trCard(domain.CardDesignHeart, 7))
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: trCard(domain.CardDesignSpade, 4)}})

	// Playing the heart while holding a spade is illegal.
	assert.Error(t, g.PlayerPlay(1))
	// Playing the spade is legal.
	assert.NoError(t, g.PlayerPlay(0))
}

func TestTressettePlayerPlayLeadAnySuit(t *testing.T) {
	g := newTestTressette()
	g.SetPhase(domain.TressettePhasePlay)
	g.SetCurrentPlayerIdx(0)
	trSetHand(g.GetPlayer(0), trCard(domain.CardDesignHeart, 7))
	assert.NoError(t, g.PlayerPlay(0)) // leading: any card allowed
	assert.Len(t, g.GetCurrentTrick(), 1)
}

func TestTressettePlayerPlayErrors(t *testing.T) {
	g := newTestTressette()

	g.SetPhase(domain.TressettePhaseTrickEnd)
	assert.ErrorIs(t, g.PlayerPlay(0), domain.ErrWrongPhase)

	g.SetPhase(domain.TressettePhasePlay)
	g.SetCurrentPlayerIdx(1) // CPU's turn
	assert.ErrorIs(t, g.PlayerPlay(0), domain.ErrNotHumanTurn)

	g.SetCurrentPlayerIdx(0)
	trSetHand(g.GetPlayer(0), trCard(domain.CardDesignSpade, 1))
	assert.Error(t, g.PlayerPlay(5)) // out of range

	g.SetPhase(domain.TressettePhaseGameEnd)
	g.SetCurrentPlayerIdx(0)
	// gameEndFlag drives ErrGameEnded
	g.SetTeamScores([2]int{99, 0})
}

func TestTressettePlayCompletesTrick(t *testing.T) {
	g := newTestTressette()
	g.SetPhase(domain.TressettePhasePlay)
	g.SetCurrentPlayerIdx(0)
	for i := 0; i < 4; i++ {
		trSetHand(g.GetPlayer(i), trCard(domain.CardDesignSpade, 4+i))
	}
	// players 1,2,3 play first via direct trick set, human completes.
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: trCard(domain.CardDesignSpade, 5)},
		{PlayerIdx: 2, Card: trCard(domain.CardDesignSpade, 6)},
		{PlayerIdx: 3, Card: trCard(domain.CardDesignSpade, 7)},
	})
	assert.NoError(t, g.PlayerPlay(0)) // 4th card → TrickEnd
	assert.Equal(t, domain.TressettePhaseTrickEnd, g.GetPhase())
}

func TestTressetteCpuPlay(t *testing.T) {
	g := newTestTressette()
	g.SetPhase(domain.TressettePhasePlay)
	g.SetCurrentPlayerIdx(1)
	trSetHand(g.GetPlayer(1), trCard(domain.CardDesignSpade, 4), trCard(domain.CardDesignHeart, 7))
	g.CpuPlay()
	assert.Len(t, g.GetCurrentTrick(), 1)
}

func TestTressetteCpuPlayFollowsAndWins(t *testing.T) {
	// Opponent leads a low spade with a point card; CPU should be able to act.
	cfg := domain.DefaultTressetteConfig()
	cfg.CpuDifficulty = domain.TressetteCpuDifficultyHard
	g := newTestTressette()
	g.SetConfig(cfg)
	g.SetPhase(domain.TressettePhasePlay)
	g.SetCurrentPlayerIdx(1)
	trSetHand(g.GetPlayer(1), trCard(domain.CardDesignSpade, 3), trCard(domain.CardDesignSpade, 4))
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 0, Card: trCard(domain.CardDesignSpade, 1)}})
	g.CpuPlay()
	assert.Len(t, g.GetCurrentTrick(), 2)
}

func TestTressetteCpuPlayEasyRandom(t *testing.T) {
	cfg := domain.DefaultTressetteConfig()
	cfg.CpuDifficulty = domain.TressetteCpuDifficultyEasy
	g := newTestTressette()
	g.SetConfig(cfg)
	g.SetPhase(domain.TressettePhasePlay)
	g.SetCurrentPlayerIdx(1)
	trSetHand(g.GetPlayer(1), trCard(domain.CardDesignSpade, 4), trCard(domain.CardDesignSpade, 5))
	g.CpuPlay()
	assert.Equal(t, 1, g.GetPlayer(1).GetCardsSize())
}

// --- next trick / next round ---

func TestTressetteNextTrick(t *testing.T) {
	g := newTestTressette()
	g.SetPhase(domain.TressettePhaseTrickEnd)
	g.SetLeadPlayerIdx(2)
	g.SetTrickNumber(3)
	g.NextTrick()
	assert.Equal(t, domain.TressettePhasePlay, g.GetPhase())
	assert.Equal(t, 2, g.GetCurrentPlayerIdx())
	assert.Equal(t, 4, g.GetTrickNumber())
	assert.Empty(t, g.GetCurrentTrick())
}

func TestTressetteNextTrickWrongPhase(t *testing.T) {
	g := newTestTressette()
	g.SetPhase(domain.TressettePhasePlay)
	g.NextTrick()
	assert.Equal(t, domain.TressettePhasePlay, g.GetPhase())
}

func TestTressetteNextRound(t *testing.T) {
	g := newTestTressette()
	g.SetRoundNumber(1)
	g.SetPhase(domain.TressettePhaseRoundEnd)
	g.NextRound()
	assert.Equal(t, 2, g.GetRoundNumber())
	assert.Equal(t, domain.TressettePhasePlay, g.GetPhase())
	for i := 0; i < 4; i++ {
		assert.Equal(t, 10, g.GetPlayer(i).GetCardsSize())
	}
}

func TestTressetteNextRoundWrongPhase(t *testing.T) {
	g := newTestTressette()
	g.SetRoundNumber(1)
	g.SetPhase(domain.TressettePhasePlay)
	g.NextRound()
	assert.Equal(t, 1, g.GetRoundNumber())
}

// --- hint & playable indices ---

func TestTressetteHintLead(t *testing.T) {
	g := newTestTressette()
	g.SetPhase(domain.TressettePhasePlay)
	g.SetCurrentPlayerIdx(0)
	trSetHand(g.GetPlayer(0), trCard(domain.CardDesignSpade, 4), trCard(domain.CardDesignHeart, 1))
	hint := g.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "lead_low", hint.Reason)
	assert.Len(t, hint.CardIndices, 1)
}

func TestTressetteHintNilWhenNotHumanTurn(t *testing.T) {
	g := newTestTressette()
	g.SetPhase(domain.TressettePhasePlay)
	g.SetCurrentPlayerIdx(1)
	assert.Nil(t, g.GetHint())
}

func TestTressetteHintFollowReasons(t *testing.T) {
	g := newTestTressette()
	g.SetPhase(domain.TressettePhasePlay)
	g.SetCurrentPlayerIdx(0)
	trSetHand(g.GetPlayer(0), trCard(domain.CardDesignSpade, 1), trCard(domain.CardDesignSpade, 4))

	// Opponent (player 3, team B) leads ♠K (a point card, strength 6) and the
	// human holds ♠A (strength 7) → worth winning the points.
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: trCard(domain.CardDesignSpade, 13)}})
	assert.Equal(t, "follow_win", g.GetHint().Reason)

	// Partner (player 2, team A) is winning → give partner points.
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 2, Card: trCard(domain.CardDesignSpade, 3)}})
	assert.Equal(t, "give_partner", g.GetHint().Reason)
}

func TestTressetteHintDiscard(t *testing.T) {
	g := newTestTressette()
	g.SetPhase(domain.TressettePhasePlay)
	g.SetCurrentPlayerIdx(0)
	trSetHand(g.GetPlayer(0), trCard(domain.CardDesignHeart, 4))
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: trCard(domain.CardDesignSpade, 5)}})
	assert.Equal(t, "discard_low", g.GetHint().Reason)
}

func TestTressettePlayableIndices(t *testing.T) {
	g := newTestTressette()
	g.SetPhase(domain.TressettePhasePlay)
	g.SetCurrentPlayerIdx(0)
	trSetHand(g.GetPlayer(0), trCard(domain.CardDesignSpade, 1), trCard(domain.CardDesignHeart, 7))
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: trCard(domain.CardDesignSpade, 4)}})
	// Only the spade (index 0) is playable.
	assert.Equal(t, []int{0}, g.GetPlayableIndices(0))
	assert.Nil(t, g.GetPlayableIndices(99))
}

// --- JSON ---

func TestTressetteJSONRoundTrip(t *testing.T) {
	g := newTestTressette()
	g.Reset()
	g.SetTeamScores([2]int{5, 3})
	data, err := json.Marshal(g)
	assert.NoError(t, err)

	var restored domain.Tressette
	assert.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, g.GetPhase(), restored.GetPhase())
	assert.Equal(t, g.GetRoundNumber(), restored.GetRoundNumber())
	assert.Equal(t, g.GetPlayerCnt(), restored.GetPlayerCnt())
	assert.Equal(t, [2]int{5, 3}, restored.GetTeamScores())
}

func TestTressetteUnmarshalInvalid(t *testing.T) {
	var g domain.Tressette
	assert.Error(t, json.Unmarshal([]byte("not json"), &g))
}
