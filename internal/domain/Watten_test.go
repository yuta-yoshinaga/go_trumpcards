//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestWatten() *domain.Watten {
	players := []*domain.WattenPlayer{
		domain.NewWattenPlayer(true, 0),  // P0: human, team 0
		domain.NewWattenPlayer(false, 1), // P1: CPU, team 1
		domain.NewWattenPlayer(false, 0), // P2: CPU, team 0
		domain.NewWattenPlayer(false, 1), // P3: CPU, team 1
	}
	return domain.NewWatten(domain.NewDefaultWatten().GetConfigDeckHelper(), players, domain.DefaultWattenConfig())
}

func wattenSetHand(g *domain.Watten, playerIdx int, cards []*domain.Card) {
	p := g.GetPlayer(playerIdx)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func wattenCard(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

// --- Config ---

func TestWattenConfig_Default(t *testing.T) {
	cfg := domain.DefaultWattenConfig()
	assert.Equal(t, domain.WattenCpuDifficultyNormal, cfg.CpuDifficulty)
	assert.Equal(t, 15, cfg.TargetScore)
	assert.Equal(t, 5, cfg.MaxRaises)
	assert.NoError(t, cfg.Validate())
}

func TestWattenConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     domain.WattenConfig
		wantErr bool
	}{
		{"default", domain.DefaultWattenConfig(), false},
		{"bad difficulty", domain.WattenConfig{CpuDifficulty: 9, TargetScore: 15, MaxRaises: 5}, true},
		{"zero target", domain.WattenConfig{CpuDifficulty: domain.WattenCpuDifficultyNormal, TargetScore: 0, MaxRaises: 5}, true},
		{"neg raises", domain.WattenConfig{CpuDifficulty: domain.WattenCpuDifficultyNormal, TargetScore: 15, MaxRaises: -1}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				assert.Error(t, tt.cfg.Validate())
			} else {
				assert.NoError(t, tt.cfg.Validate())
			}
		})
	}
}

// --- Deck: 32 cards, 5 dealt each ---

func TestNewDefaultWatten_Deck(t *testing.T) {
	g := domain.NewDefaultWatten()
	g.Reset()
	total := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		total += g.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, 20, total) // 5 each × 4
	assert.Equal(t, 4, g.GetPlayerCnt())
}

// --- Ranking: Max > Belli > Spitz > Schlag > Critical > Plain ---

func TestWatten_Ranking_Tiers(t *testing.T) {
	g := newTestWatten()
	g.SetCriticalSuit(domain.CardDesignSpade)
	g.SetSchlagRank(10)

	maxC := g.CardRankPublic(wattenCard(domain.CardDesignHeart, 13))    // ♥K = Max
	belli := g.CardRankPublic(wattenCard(domain.CardDesignDiamond, 13)) // ♦K = Belli
	spitz := g.CardRankPublic(wattenCard(domain.CardDesignDiamond, 7))  // ♦7 = Spitz
	schlag := g.CardRankPublic(wattenCard(domain.CardDesignHeart, 10))  // 10♥ = Schlag
	crit := g.CardRankPublic(wattenCard(domain.CardDesignSpade, 1))     // A♠ = critical
	plain := g.CardRankPublic(wattenCard(domain.CardDesignClover, 1))   // A♣ = plain

	assert.Greater(t, maxC, belli)
	assert.Greater(t, belli, spitz)
	assert.Greater(t, spitz, schlag)
	assert.Greater(t, schlag, crit)
	assert.Greater(t, crit, plain)
}

// --- Schlag suit order ♥ > ♦ > ♠ > ♣ ---

func TestWatten_SchlagSuitOrder(t *testing.T) {
	g := newTestWatten()
	g.SetCriticalSuit(domain.CardDesignSpade)
	g.SetSchlagRank(10)
	h := g.CardRankPublic(wattenCard(domain.CardDesignHeart, 10))
	d := g.CardRankPublic(wattenCard(domain.CardDesignDiamond, 10))
	s := g.CardRankPublic(wattenCard(domain.CardDesignSpade, 10)) // schlag beats critical even though ♠ is critical suit
	c := g.CardRankPublic(wattenCard(domain.CardDesignClover, 10))
	assert.Greater(t, h, d)
	assert.Greater(t, d, s)
	assert.Greater(t, s, c)
	// A ♠10 is Schlag, not critical.
	assert.True(t, g.IsTrumpPublic(wattenCard(domain.CardDesignSpade, 10)))
}

// --- Critical-suit value order A>K>Q>J>10>9>8>7 ---

func TestWatten_CriticalValueOrder(t *testing.T) {
	g := newTestWatten()
	g.SetCriticalSuit(domain.CardDesignSpade)
	g.SetSchlagRank(11) // Jacks are schlag, so critical excludes J♠
	order := []int{1, 13, 12, 10, 9, 8, 7}
	for i := 0; i < len(order)-1; i++ {
		hi := g.CardRankPublic(wattenCard(domain.CardDesignSpade, order[i]))
		lo := g.CardRankPublic(wattenCard(domain.CardDesignSpade, order[i+1]))
		assert.Greater(t, hi, lo, "critical %d should beat %d", order[i], order[i+1])
	}
}

// --- Schlag=King: Max/Belli take precedence over schlag Kings ---

func TestWatten_SchlagKing(t *testing.T) {
	g := newTestWatten()
	g.SetCriticalSuit(domain.CardDesignSpade)
	g.SetSchlagRank(13)
	assert.Equal(t, 1000, g.CardRankPublic(wattenCard(domain.CardDesignHeart, 13)))  // Max
	assert.Equal(t, 999, g.CardRankPublic(wattenCard(domain.CardDesignDiamond, 13))) // Belli
	// K♠, K♣ are schlag.
	assert.Greater(t, g.CardRankPublic(wattenCard(domain.CardDesignSpade, 13)), 899)
	assert.Less(t, g.CardRankPublic(wattenCard(domain.CardDesignSpade, 13)), 998)
}

// --- isTrump membership ---

func TestWatten_IsTrump(t *testing.T) {
	g := newTestWatten()
	g.SetCriticalSuit(domain.CardDesignSpade)
	g.SetSchlagRank(10)
	assert.True(t, g.IsTrumpPublic(wattenCard(domain.CardDesignHeart, 13)))   // Max
	assert.True(t, g.IsTrumpPublic(wattenCard(domain.CardDesignDiamond, 13))) // Belli
	assert.True(t, g.IsTrumpPublic(wattenCard(domain.CardDesignDiamond, 7)))  // Spitz
	assert.True(t, g.IsTrumpPublic(wattenCard(domain.CardDesignClover, 10)))  // Schlag
	assert.True(t, g.IsTrumpPublic(wattenCard(domain.CardDesignSpade, 8)))    // critical
	assert.False(t, g.IsTrumpPublic(wattenCard(domain.CardDesignClover, 8)))  // plain
	assert.False(t, g.IsTrumpPublic(nil))
}

// --- Must-follow: trump lead ---

func TestWatten_ValidatePlay_TrumpLead(t *testing.T) {
	g := newTestWatten()
	g.SetCriticalSuit(domain.CardDesignHeart)
	g.SetSchlagRank(10)
	g.SetPhase(domain.WattenPhasePlay)
	g.SetCurrentPlayerIdx(1)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: wattenCard(domain.CardDesignHeart, 8)}, // heart = trump lead
	})
	// P1 holds a trump (♥9) and a plain (♠7): must follow trump.
	wattenSetHand(g, 1, []*domain.Card{wattenCard(domain.CardDesignHeart, 9), wattenCard(domain.CardDesignSpade, 7)})
	assert.Equal(t, []int{0}, g.GetValidPlayIndices(1))

	// P1 void of trump: may play anything.
	wattenSetHand(g, 1, []*domain.Card{wattenCard(domain.CardDesignSpade, 7), wattenCard(domain.CardDesignClover, 8)})
	assert.Len(t, g.GetValidPlayIndices(1), 2)
}

// --- Must-follow: plain-suit lead ---

func TestWatten_ValidatePlay_PlainLead(t *testing.T) {
	g := newTestWatten()
	g.SetCriticalSuit(domain.CardDesignHeart)
	g.SetSchlagRank(10)
	g.SetPhase(domain.WattenPhasePlay)
	g.SetCurrentPlayerIdx(1)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: wattenCard(domain.CardDesignSpade, 13)}, // K♠ plain lead
	})
	// P1 holds a plain spade (♠7) and a trump heart (♥7): must follow spade.
	wattenSetHand(g, 1, []*domain.Card{wattenCard(domain.CardDesignSpade, 7), wattenCard(domain.CardDesignHeart, 7)})
	assert.Equal(t, []int{0}, g.GetValidPlayIndices(1))

	// P1 void of plain spade: may play anything (incl. trump).
	wattenSetHand(g, 1, []*domain.Card{wattenCard(domain.CardDesignHeart, 7), wattenCard(domain.CardDesignClover, 8)})
	assert.Len(t, g.GetValidPlayIndices(1), 2)
}

// --- Trick winner: highest trump wins ---

func TestWatten_TrickWinner_Trump(t *testing.T) {
	g := newTestWatten()
	g.SetCriticalSuit(domain.CardDesignSpade)
	g.SetSchlagRank(10)
	g.SetPhase(domain.WattenPhaseTrickEnd)
	g.SetTrickNumber(1)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: wattenCard(domain.CardDesignClover, 1)}, // A♣ plain lead
		{PlayerIdx: 1, Card: wattenCard(domain.CardDesignClover, 13)},
		{PlayerIdx: 2, Card: wattenCard(domain.CardDesignSpade, 7)}, // ♠7 critical trump - wins
		{PlayerIdx: 3, Card: wattenCard(domain.CardDesignDiamond, 8)},
	})
	g.ResolveTrick()
	assert.Equal(t, 2, g.GetLeadPlayerIdx())
	assert.Equal(t, 1, g.GetTeamTricks(0)) // P2 is team 0
}

// --- Trick winner: Max beats all ---

func TestWatten_TrickWinner_Max(t *testing.T) {
	g := newTestWatten()
	g.SetCriticalSuit(domain.CardDesignSpade)
	g.SetSchlagRank(10)
	g.SetPhase(domain.WattenPhaseTrickEnd)
	g.SetTrickNumber(1)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: wattenCard(domain.CardDesignHeart, 13)},   // Max
		{PlayerIdx: 1, Card: wattenCard(domain.CardDesignSpade, 10)},   // schlag
		{PlayerIdx: 2, Card: wattenCard(domain.CardDesignDiamond, 13)}, // Belli
		{PlayerIdx: 3, Card: wattenCard(domain.CardDesignDiamond, 7)},  // Spitz
	})
	g.ResolveTrick()
	assert.Equal(t, 0, g.GetLeadPlayerIdx())
}

// --- Trick winner: plain led, higher of led suit ---

func TestWatten_TrickWinner_Plain(t *testing.T) {
	g := newTestWatten()
	g.SetCriticalSuit(domain.CardDesignSpade)
	g.SetSchlagRank(11)
	g.SetPhase(domain.WattenPhaseTrickEnd)
	g.SetTrickNumber(1)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: wattenCard(domain.CardDesignClover, 9)}, // lead
		{PlayerIdx: 1, Card: wattenCard(domain.CardDesignClover, 1)}, // A♣ - highest clover, wins
		{PlayerIdx: 2, Card: wattenCard(domain.CardDesignClover, 8)}, // follow low
		{PlayerIdx: 3, Card: wattenCard(domain.CardDesignHeart, 7)},  // off-suit plain (heart, R=11 so not trump; crit=spade)
	})
	g.ResolveTrick()
	assert.Equal(t, 1, g.GetLeadPlayerIdx())
}

// --- Last trick -> RoundEnd + deal scoring on phase entry ---

func TestWatten_LastTrick_DealScore(t *testing.T) {
	g := newTestWatten()
	g.SetCriticalSuit(domain.CardDesignSpade)
	g.SetSchlagRank(10)
	g.SetStake(3)
	g.SetPhase(domain.WattenPhaseTrickEnd)
	g.SetTrickNumber(domain.WattenHandSize) // last trick
	g.SetTeamTricksForTest(0, 2)            // team 0 already has 2
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: wattenCard(domain.CardDesignSpade, 1)}, // A♠ trump, P0 (team0) wins
		{PlayerIdx: 1, Card: wattenCard(domain.CardDesignClover, 7)},
		{PlayerIdx: 2, Card: wattenCard(domain.CardDesignClover, 8)},
		{PlayerIdx: 3, Card: wattenCard(domain.CardDesignClover, 9)},
	})
	g.ResolveTrick()
	assert.Equal(t, domain.WattenPhaseRoundEnd, g.GetPhase())
	assert.Equal(t, 0, g.GetDealWinnerTeam())
	assert.Equal(t, 3, g.GetTeamScore(0)) // stake awarded on entry
}

// --- Declaration: human ---

func TestWatten_PlayerDeclare(t *testing.T) {
	g := newTestWatten()
	g.SetDealerIdx(0) // human is dealer
	g.SetPhase(domain.WattenPhaseDeclare)
	g.SetCurrentPlayerIdx(0)
	wattenSetHand(g, 0, []*domain.Card{wattenCard(domain.CardDesignSpade, 10)})
	assert.True(t, g.IsHumanDeclareTurn())
	assert.NoError(t, g.PlayerDeclare(10, domain.CardDesignSpade))
	assert.Equal(t, 10, g.GetSchlagRank())
	assert.Equal(t, domain.CardDesignSpade, g.GetCriticalSuit())
	assert.Equal(t, domain.WattenPhasePlay, g.GetPhase())
	// Lead is left of dealer.
	assert.Equal(t, 1, g.GetLeadPlayerIdx())
}

func TestWatten_PlayerDeclare_Errors(t *testing.T) {
	g := newTestWatten()
	g.SetDealerIdx(0)
	g.SetPhase(domain.WattenPhaseDeclare)
	wattenSetHand(g, 0, []*domain.Card{wattenCard(domain.CardDesignSpade, 10)})
	assert.ErrorIs(t, g.PlayerDeclare(5, domain.CardDesignSpade), domain.ErrInvalidPlay) // bad rank
	assert.ErrorIs(t, g.PlayerDeclare(10, 9), domain.ErrInvalidPlay)                     // bad suit
	// Wrong phase.
	g.SetPhase(domain.WattenPhasePlay)
	assert.ErrorIs(t, g.PlayerDeclare(10, domain.CardDesignSpade), domain.ErrWrongPhase)
}

// --- Declaration: CPU picks most-held suit + rank ---

func TestWatten_CpuDeclare(t *testing.T) {
	g := newTestWatten()
	g.SetDealerIdx(1) // CPU dealer
	g.SetPhase(domain.WattenPhaseDeclare)
	wattenSetHand(g, 1, []*domain.Card{
		wattenCard(domain.CardDesignSpade, 10), wattenCard(domain.CardDesignSpade, 8),
		wattenCard(domain.CardDesignSpade, 7), wattenCard(domain.CardDesignClover, 10),
		wattenCard(domain.CardDesignHeart, 10),
	})
	g.CpuDeclare()
	assert.Equal(t, domain.CardDesignSpade, g.GetCriticalSuit()) // 3 spades
	assert.Equal(t, 10, g.GetSchlagRank())                       // three 10s
	assert.Equal(t, domain.WattenPhasePlay, g.GetPhase())
}

// --- Raise / hold ---

func TestWatten_Raise_Hold(t *testing.T) {
	g := newTestWatten()
	g.SetCriticalSuit(domain.CardDesignSpade)
	g.SetSchlagRank(10)
	g.SetPhase(domain.WattenPhasePlay)
	g.SetLeadPlayerIdx(0)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	g.SetStake(domain.WattenBaseStake)
	wattenSetHand(g, 0, []*domain.Card{wattenCard(domain.CardDesignSpade, 1)})

	assert.True(t, g.CanHumanRaise())
	assert.NoError(t, g.PlayerRaise())
	assert.Equal(t, domain.WattenPhaseRespond, g.GetPhase())
	assert.Equal(t, 3, g.GetPendingStake())
	// Responder is on team 1.
	assert.Equal(t, 1, g.GetRaiserTeam()^1) // raiser team 0 -> responder team 1
	assert.Equal(t, 1, g.GetResponderIdx())

	// Force human responder and hold.
	g.SetupRaiseForTest(3, 0, 0)
	g.SetStake(2)
	assert.NoError(t, g.PlayerRespond(true))
	assert.Equal(t, 3, g.GetStake())
	assert.Equal(t, 1, g.GetRaiseCount())
	assert.Equal(t, domain.WattenPhasePlay, g.GetPhase())
}

// --- Raise cap ---

func TestWatten_Raise_Cap(t *testing.T) {
	g := newTestWatten()
	g.SetPhase(domain.WattenPhasePlay)
	g.SetLeadPlayerIdx(0)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	g.SetRaiseCountForTest(5) // == default MaxRaises
	assert.False(t, g.CanHumanRaise())
	assert.Error(t, g.PlayerRaise())
}

// --- Fold: concede deal, raising team scores last-accepted stake ---

func TestWatten_Fold_Concede(t *testing.T) {
	g := newTestWatten()
	g.SetStake(2)
	g.SetTeamScore(1, 0)
	g.SetupRaiseForTest(3, 1, 0) // team 1 raised, human (team 0) responds
	assert.True(t, g.IsHumanRespondTurn())
	assert.NoError(t, g.PlayerRespond(false)) // fold
	assert.Equal(t, domain.WattenPhaseRoundEnd, g.GetPhase())
	assert.Equal(t, 1, g.GetDealWinnerTeam())
	assert.Equal(t, 2, g.GetTeamScore(1)) // last-accepted stake (2), not 3
}

func TestWatten_Respond_Errors(t *testing.T) {
	g := newTestWatten()
	g.SetPhase(domain.WattenPhasePlay)
	assert.ErrorIs(t, g.PlayerRespond(true), domain.ErrWrongPhase)
	g.SetupRaiseForTest(3, 1, 1) // responder is CPU
	assert.ErrorIs(t, g.PlayerRespond(true), domain.ErrNotHumanTurn)
}

// --- Match end to 15 ---

func TestWatten_MatchEnd(t *testing.T) {
	g := newTestWatten()
	g.SetCriticalSuit(domain.CardDesignSpade)
	g.SetSchlagRank(10)
	g.SetStake(2)
	g.SetTeamScore(0, 13)
	g.SetPhase(domain.WattenPhaseTrickEnd)
	g.SetTrickNumber(domain.WattenHandSize)
	g.SetTeamTricksForTest(0, 3)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: wattenCard(domain.CardDesignSpade, 1)},
		{PlayerIdx: 1, Card: wattenCard(domain.CardDesignClover, 7)},
		{PlayerIdx: 2, Card: wattenCard(domain.CardDesignClover, 8)},
		{PlayerIdx: 3, Card: wattenCard(domain.CardDesignClover, 9)},
	})
	g.ResolveTrick()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, domain.WattenPhaseGameEnd, g.GetPhase())
	assert.Equal(t, 0, g.GetWinnerTeam())
	assert.Equal(t, domain.WattenResultWin, g.GetResult())
}

func TestWatten_MatchEnd_TeamOneLose(t *testing.T) {
	g := newTestWatten()
	g.SetStake(2)
	g.SetTeamScore(1, 14)
	g.SetupRaiseForTest(3, 1, 0)
	assert.NoError(t, g.PlayerRespond(false)) // fold -> team1 +2 = 16
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 1, g.GetWinnerTeam())
	assert.Equal(t, domain.WattenResultLose, g.GetResult())
}

// --- Full deal drives CPU declare/play/respond, resolve, next trick ---

func TestWatten_FullDeal(t *testing.T) {
	g := domain.NewDefaultWatten()
	g.Reset()
	// Declaration phase.
	for g.GetPhase() == domain.WattenPhaseDeclare {
		if g.IsHumanDeclareTurn() {
			assert.NoError(t, g.PlayerDeclare(10, domain.CardDesignHeart))
		} else {
			g.CpuDeclare()
		}
	}
	assert.Equal(t, domain.WattenPhasePlay, g.GetPhase())

	// Play out the deal (auto-driving all phases) until it ends.
	for i := 0; i < 200 && g.GetPhase() != domain.WattenPhaseRoundEnd && g.GetPhase() != domain.WattenPhaseGameEnd; i++ {
		switch g.GetPhase() {
		case domain.WattenPhasePlay:
			if g.IsHumanTurn() {
				valid := g.GetValidPlayIndices(g.GetCurrentPlayerIdx())
				assert.NotEmpty(t, valid)
				_ = g.PlayerPlay(valid[0])
			} else {
				g.CpuPlay()
			}
		case domain.WattenPhaseRespond:
			if g.IsHumanRespondTurn() {
				_ = g.PlayerRespond(true)
			} else {
				g.CpuRespond()
			}
		case domain.WattenPhaseTrickEnd:
			g.ResolveTrick()
			g.NextTrick()
		}
	}
	phase := g.GetPhase()
	assert.True(t, phase == domain.WattenPhaseRoundEnd || phase == domain.WattenPhaseGameEnd)
}

// --- NextRound advances dealer and re-deals ---

func TestWatten_NextRound(t *testing.T) {
	g := newTestWatten()
	g.SetStake(2)
	g.SetTeamScore(0, 0)
	g.SetPhase(domain.WattenPhaseTrickEnd)
	g.SetTrickNumber(domain.WattenHandSize)
	g.SetTeamTricksForTest(0, 3)
	g.SetCriticalSuit(domain.CardDesignSpade)
	g.SetSchlagRank(10)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: wattenCard(domain.CardDesignSpade, 1)},
		{PlayerIdx: 1, Card: wattenCard(domain.CardDesignClover, 7)},
		{PlayerIdx: 2, Card: wattenCard(domain.CardDesignClover, 8)},
		{PlayerIdx: 3, Card: wattenCard(domain.CardDesignClover, 9)},
	})
	g.ResolveTrick()
	assert.Equal(t, domain.WattenPhaseRoundEnd, g.GetPhase())
	before := g.GetRoundNumber()
	g.ScoreRound() // idempotent
	g.NextRound()
	assert.Equal(t, before+1, g.GetRoundNumber())
	assert.Equal(t, domain.WattenPhaseDeclare, g.GetPhase())
	assert.Equal(t, 1, g.GetDealerIdx())
}

// --- Hints across phases ---

func TestWatten_GetHint(t *testing.T) {
	g := newTestWatten()

	// Declare hint (human dealer).
	g.SetDealerIdx(0)
	g.SetPhase(domain.WattenPhaseDeclare)
	wattenSetHand(g, 0, []*domain.Card{wattenCard(domain.CardDesignSpade, 10)})
	h := g.GetHint()
	assert.NotNil(t, h)
	assert.Equal(t, "declare", h.Action)

	// Declare but not human dealer -> nil.
	g.SetDealerIdx(1)
	assert.Nil(t, g.GetHint())

	// Play hint.
	g.SetCriticalSuit(domain.CardDesignSpade)
	g.SetSchlagRank(10)
	g.SetPhase(domain.WattenPhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.SetLeadPlayerIdx(2)
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 2, Card: wattenCard(domain.CardDesignClover, 9)}})
	wattenSetHand(g, 0, []*domain.Card{wattenCard(domain.CardDesignClover, 1), wattenCard(domain.CardDesignHeart, 8)})
	hp := g.GetHint()
	assert.NotNil(t, hp)
	assert.Contains(t, []string{"play", "raise"}, hp.Action)

	// Respond hint.
	g.SetupRaiseForTest(3, 1, 0)
	hr := g.GetHint()
	assert.NotNil(t, hr)
	assert.Contains(t, []string{"hold", "fold"}, hr.Action)

	// Respond but not human -> nil.
	g.SetupRaiseForTest(3, 1, 1)
	assert.Nil(t, g.GetHint())
}

// --- Getters coverage ---

func TestWatten_Getters(t *testing.T) {
	g := newTestWatten()
	g.Reset()
	_ = g.GetDealerIdx()
	_ = g.GetActionLog()
	_ = g.GetTeamScore(0)
	_ = g.GetTeamScore(-1)
	_ = g.GetTeamTricks(9)
	_ = g.GetTrickNumber()
	_ = g.GetStake()
	_ = g.GetPendingStake()
	_ = g.GetRaiseCount()
	_ = g.GetRaiserTeam()
	_ = g.GetResponderIdx()
	_ = g.GetDealWinnerTeam()
	_ = g.GetResult()
	assert.Nil(t, g.GetPlayer(99))
	assert.Equal(t, 15, g.GetConfig().TargetScore)
	assert.False(t, g.GetGameEndFlag())
}

// --- JSON round-trip + validation ---

func TestWatten_JSON_RoundTrip(t *testing.T) {
	g := domain.NewDefaultWatten()
	g.Reset()
	data, err := json.Marshal(g)
	assert.NoError(t, err)

	var restored domain.Watten
	assert.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, g.GetPhase(), restored.GetPhase())
	assert.Equal(t, g.GetPlayerCnt(), restored.GetPlayerCnt())
	assert.Equal(t, g.GetDealerIdx(), restored.GetDealerIdx())
}

func TestWatten_Unmarshal_Validation(t *testing.T) {
	var g domain.Watten
	// Invalid phase.
	assert.Error(t, json.Unmarshal([]byte(`{"ph":99,"pl":[{},{},{},{}]}`), &g))
	// Wrong player count.
	assert.Error(t, json.Unmarshal([]byte(`{"ph":1,"pl":[{},{}]}`), &g))
	// Play phase requires declaration.
	assert.Error(t, json.Unmarshal([]byte(`{"ph":1,"sr":0,"cs":0,"pl":[{},{},{},{}]}`), &g))
	// Bad schlag rank.
	assert.Error(t, json.Unmarshal([]byte(`{"ph":0,"sr":5,"pl":[{},{},{},{}]}`), &g))
	// Bad critical suit.
	assert.Error(t, json.Unmarshal([]byte(`{"ph":0,"cs":9,"pl":[{},{},{},{}]}`), &g))
	// OOB dealer.
	assert.Error(t, json.Unmarshal([]byte(`{"ph":0,"di":9,"pl":[{},{},{},{}]}`), &g))
	// Player team out of range (must be 0 or 1).
	assert.Error(t, json.Unmarshal([]byte(`{"ph":0,"pl":[{"tm":5},{},{},{}]}`), &g))
	// Winner team index validated against WattenTeamCnt (2), not the player count.
	assert.Error(t, json.Unmarshal([]byte(`{"ph":0,"pl":[{},{},{},{}],"wt":3}`), &g))
	// Raiser team index out of range.
	assert.Error(t, json.Unmarshal([]byte(`{"ph":0,"pl":[{},{},{},{}],"rt":2}`), &g))
	// Valid declare-phase state (undeclared allowed).
	assert.NoError(t, json.Unmarshal([]byte(`{"ph":0,"sr":0,"cs":0,"di":0,"cp":0,"pl":[{},{},{},{}]}`), &g))
}

// **Web の wattenTrumpCards と同じ分類。**Max/Belli/Spitz は宣言に関わらず切り札
// なので、スート別・ランク別の「増える枚数」には数えない (#4848)。
func TestWattenPreviewTrumps(t *testing.T) {
	pv := domain.WattenPreviewTrumps([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 13, false),   // Max
		domain.NewCard(domain.CardDesignDiamond, 13, false), // Belli
		domain.NewCard(domain.CardDesignDiamond, 7, false),  // Spitz
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignSpade, 10, false),
		domain.NewCard(domain.CardDesignClover, 10, false),
		nil,
	})
	assert.Equal(t, 3, pv.Permanent)
	assert.Equal(t, 2, pv.BySuit[domain.CardDesignSpade])
	assert.Equal(t, 1, pv.BySuit[domain.CardDesignClover])
	// ♥K/♦K/♦7 は常時側なので、スート別には残らない。
	assert.Equal(t, 0, pv.BySuit[domain.CardDesignHeart])
	assert.Equal(t, 0, pv.BySuit[domain.CardDesignDiamond])
	assert.Equal(t, 2, pv.ByRank[10])
	assert.Equal(t, 1, pv.ByRank[1])
	assert.NotContains(t, pv.ByRank, 13, "♥K/♦K は常時側なので K は増えない")

	empty := domain.WattenPreviewTrumps(nil)
	assert.Equal(t, 0, empty.Permanent)
	assert.Empty(t, empty.ByRank)
}
