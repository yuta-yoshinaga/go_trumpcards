//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestJass() *domain.Jass {
	players := []*domain.JassPlayer{
		domain.NewJassPlayer(true, 0),  // P0: human, team 0
		domain.NewJassPlayer(false, 1), // P1: CPU, team 1
		domain.NewJassPlayer(false, 0), // P2: CPU, team 0
		domain.NewJassPlayer(false, 1), // P3: CPU, team 1
	}
	return domain.NewJass(domain.NewDefaultJass().GetConfigDeckHelper(), players, domain.DefaultJassConfig())
}

func setJassHand(g *domain.Jass, playerIdx int, cards []*domain.Card) {
	p := g.GetPlayer(playerIdx)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func jcard(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

// --- Config ---

func TestJassConfig_Default(t *testing.T) {
	cfg := domain.DefaultJassConfig()
	assert.Equal(t, domain.JassCpuDifficultyNormal, cfg.CpuDifficulty)
	assert.Equal(t, 1000, cfg.TargetScore)
	assert.Equal(t, 5, cfg.LastTrickBonus)
	assert.True(t, cfg.EnableWeis)
	assert.NoError(t, cfg.Validate())
}

func TestJassConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     domain.JassConfig
		wantErr bool
	}{
		{"default", domain.DefaultJassConfig(), false},
		{"bad difficulty", domain.JassConfig{CpuDifficulty: 9, TargetScore: 1000, LastTrickBonus: 5}, true},
		{"zero target", domain.JassConfig{CpuDifficulty: domain.JassCpuDifficultyNormal, TargetScore: 0, LastTrickBonus: 5}, true},
		{"neg bonus", domain.JassConfig{CpuDifficulty: domain.JassCpuDifficultyNormal, TargetScore: 1000, LastTrickBonus: -1}, true},
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

// --- Deck ---

func TestNewDefaultJass_Deck(t *testing.T) {
	g := domain.NewDefaultJass()
	g.Reset()
	// After Reset, all 36 cards are dealt (9 each to 4 players).
	total := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		total += g.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, 36, total)
	assert.Equal(t, 4, g.GetPlayerCnt())
}

// --- Trump rank ordering: J > 9 > A > K > Q > 10 > 8 > 7 > 6 ---

func TestJass_TrumpRank(t *testing.T) {
	g := newTestJass()
	g.SetTrumpSuit(domain.CardDesignHeart)
	order := []int{11, 9, 1, 13, 12, 10, 8, 7, 6} // J,9,A,K,Q,10,8,7,6
	for i := 0; i < len(order)-1; i++ {
		hi := g.CardRankPublic(jcard(domain.CardDesignHeart, order[i]))
		lo := g.CardRankPublic(jcard(domain.CardDesignHeart, order[i+1]))
		assert.Greater(t, hi, lo, "trump %d should beat %d", order[i], order[i+1])
	}
	// Any trump beats any non-trump.
	assert.Greater(t,
		g.CardRankPublic(jcard(domain.CardDesignHeart, 6)),
		g.CardRankPublic(jcard(domain.CardDesignSpade, 1)))
}

// --- Non-trump rank: A > K > Q > J > 10 > 9 > 8 > 7 > 6 ---

func TestJass_NonTrumpRank(t *testing.T) {
	g := newTestJass()
	g.SetTrumpSuit(domain.CardDesignHeart)
	order := []int{1, 13, 12, 11, 10, 9, 8, 7, 6}
	for i := 0; i < len(order)-1; i++ {
		hi := g.CardRankPublic(jcard(domain.CardDesignSpade, order[i]))
		lo := g.CardRankPublic(jcard(domain.CardDesignSpade, order[i+1]))
		assert.Greater(t, hi, lo, "non-trump %d should beat %d", order[i], order[i+1])
	}
}

// --- Card points ---

func TestJass_CardPoints(t *testing.T) {
	g := newTestJass()
	g.SetTrumpSuit(domain.CardDesignHeart)
	// Trump points.
	assert.Equal(t, 20, g.CardPointsPublic(jcard(domain.CardDesignHeart, 11))) // J
	assert.Equal(t, 14, g.CardPointsPublic(jcard(domain.CardDesignHeart, 9)))
	assert.Equal(t, 11, g.CardPointsPublic(jcard(domain.CardDesignHeart, 1))) // A
	assert.Equal(t, 10, g.CardPointsPublic(jcard(domain.CardDesignHeart, 10)))
	assert.Equal(t, 4, g.CardPointsPublic(jcard(domain.CardDesignHeart, 13))) // K
	assert.Equal(t, 3, g.CardPointsPublic(jcard(domain.CardDesignHeart, 12))) // Q
	assert.Equal(t, 0, g.CardPointsPublic(jcard(domain.CardDesignHeart, 8)))
	// Non-trump points.
	assert.Equal(t, 11, g.CardPointsPublic(jcard(domain.CardDesignSpade, 1))) // A
	assert.Equal(t, 10, g.CardPointsPublic(jcard(domain.CardDesignSpade, 10)))
	assert.Equal(t, 4, g.CardPointsPublic(jcard(domain.CardDesignSpade, 13)))
	assert.Equal(t, 3, g.CardPointsPublic(jcard(domain.CardDesignSpade, 12)))
	assert.Equal(t, 2, g.CardPointsPublic(jcard(domain.CardDesignSpade, 11))) // non-trump J
	assert.Equal(t, 0, g.CardPointsPublic(jcard(domain.CardDesignSpade, 9)))
	assert.Equal(t, 0, g.CardPointsPublic(nil))
}

// --- Total card points sum to 152 (157 with last trick) ---

func TestJass_TotalCardPointsIs152(t *testing.T) {
	g := newTestJass()
	g.SetTrumpSuit(domain.CardDesignHeart)
	suits := []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond}
	values := []int{1, 6, 7, 8, 9, 10, 11, 12, 13}
	total := 0
	for _, s := range suits {
		for _, v := range values {
			total += g.CardPointsPublic(jcard(s, v))
		}
	}
	assert.Equal(t, 152, total)
}

// --- Must-follow suit; void may play anything including trump ---

func TestJass_ValidatePlay_MustFollow(t *testing.T) {
	g := newTestJass()
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetPhase(domain.JassPhasePlay)
	g.SetCurrentPlayerIdx(1)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: jcard(domain.CardDesignSpade, 13)}, // lead spade
	})
	// P1 holds a spade and a heart (trump). Must follow spade -> only spade valid.
	setJassHand(g, 1, []*domain.Card{jcard(domain.CardDesignSpade, 7), jcard(domain.CardDesignHeart, 6)})
	valid := g.GetValidPlayIndices(1)
	assert.Equal(t, []int{0}, valid)

	// P1 void in spade: may play anything (heart trump or clover).
	setJassHand(g, 1, []*domain.Card{jcard(domain.CardDesignHeart, 6), jcard(domain.CardDesignClover, 7)})
	valid = g.GetValidPlayIndices(1)
	assert.Len(t, valid, 2)
}

// --- Lead is trump: must follow trump (unless only trump is the Jack) ---

func TestJass_ValidatePlay_TrumpLead(t *testing.T) {
	g := newTestJass()
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetPhase(domain.JassPhasePlay)
	g.SetCurrentPlayerIdx(1)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: jcard(domain.CardDesignHeart, 13)}, // lead trump
	})
	// Holds two hearts: must follow trump.
	setJassHand(g, 1, []*domain.Card{jcard(domain.CardDesignHeart, 6), jcard(domain.CardDesignSpade, 7)})
	valid := g.GetValidPlayIndices(1)
	assert.Equal(t, []int{0}, valid)

	// Only trump is the Jack (Bauer): exempt — may play anything.
	setJassHand(g, 1, []*domain.Card{jcard(domain.CardDesignHeart, 11), jcard(domain.CardDesignSpade, 7)})
	valid = g.GetValidPlayIndices(1)
	assert.Len(t, valid, 2)
}

// --- Trick winner: highest trump wins; else highest of led suit ---

func TestJass_TrickWinner(t *testing.T) {
	g := newTestJass()
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetPhase(domain.JassPhaseTrickEnd)
	g.SetTrickNumber(1)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: jcard(domain.CardDesignSpade, 1)},  // A spade (lead)
		{PlayerIdx: 1, Card: jcard(domain.CardDesignSpade, 13)}, // K spade
		{PlayerIdx: 2, Card: jcard(domain.CardDesignHeart, 6)},  // 6 trump - wins
		{PlayerIdx: 3, Card: jcard(domain.CardDesignClover, 1)}, // off-suit
	})
	g.ResolveTrick()
	assert.Equal(t, 2, g.GetLeadPlayerIdx())
	// Trump J beats trump 6.
	g.SetPhase(domain.JassPhaseTrickEnd)
	g.SetTrickNumber(1)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: jcard(domain.CardDesignHeart, 1)},  // A trump
		{PlayerIdx: 1, Card: jcard(domain.CardDesignHeart, 11)}, // J trump - wins
		{PlayerIdx: 2, Card: jcard(domain.CardDesignHeart, 9)},  // 9 trump
		{PlayerIdx: 3, Card: jcard(domain.CardDesignHeart, 13)}, // K trump
	})
	g.ResolveTrick()
	assert.Equal(t, 1, g.GetLeadPlayerIdx())
}

// --- Last trick bonus +5 ---

func TestJass_LastTrickBonus(t *testing.T) {
	g := newTestJass()
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetPhase(domain.JassPhaseTrickEnd)
	g.SetTrickNumber(domain.JassHandSize) // 9 = last trick
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: jcard(domain.CardDesignSpade, 1)}, // 11 pts, wins (team 0)
		{PlayerIdx: 1, Card: jcard(domain.CardDesignSpade, 6)},
		{PlayerIdx: 2, Card: jcard(domain.CardDesignSpade, 7)},
		{PlayerIdx: 3, Card: jcard(domain.CardDesignSpade, 8)},
	})
	g.ResolveTrick()
	assert.Equal(t, domain.JassPhaseRoundEnd, g.GetPhase())
	// Team 0 (P0): 11 (A) + 5 (last trick) = 16.
	assert.Equal(t, 16, g.GetRoundPoints(0))
}

// --- Weis: sequence points ---

func TestJass_Weis_Sequence(t *testing.T) {
	g := newTestJass()
	// P0 has a 4-card spade sequence 6-7-8-9 (=50). Others nothing.
	setJassHand(g, 0, []*domain.Card{
		jcard(domain.CardDesignSpade, 6), jcard(domain.CardDesignSpade, 7),
		jcard(domain.CardDesignSpade, 8), jcard(domain.CardDesignSpade, 9),
		jcard(domain.CardDesignHeart, 1),
	})
	setJassHand(g, 1, []*domain.Card{jcard(domain.CardDesignClover, 1)})
	setJassHand(g, 2, []*domain.Card{jcard(domain.CardDesignDiamond, 1)})
	setJassHand(g, 3, []*domain.Card{jcard(domain.CardDesignClover, 6)})
	g.ResolveWeisForTest(domain.CardDesignHeart, 0)
	assert.Equal(t, 50, g.GetRoundWeisPoints(0))
	assert.Equal(t, 0, g.GetRoundWeisPoints(1))
}

// --- Weis: four of a kind beats sequence; winning team scores all members ---

func TestJass_Weis_FourOfAKind(t *testing.T) {
	g := newTestJass()
	// P1 (team 1) has four Jacks (=200). P0 (team 0) has a 3-seq (=20). P2 (team 0)
	// has a 3-seq too (=20). Team 1 wins; team 1 scores only P1's 200.
	setJassHand(g, 0, []*domain.Card{
		jcard(domain.CardDesignSpade, 6), jcard(domain.CardDesignSpade, 7), jcard(domain.CardDesignSpade, 8),
	})
	setJassHand(g, 1, []*domain.Card{
		jcard(domain.CardDesignSpade, 11), jcard(domain.CardDesignClover, 11),
		jcard(domain.CardDesignHeart, 11), jcard(domain.CardDesignDiamond, 11),
	})
	setJassHand(g, 2, []*domain.Card{
		jcard(domain.CardDesignClover, 6), jcard(domain.CardDesignClover, 7), jcard(domain.CardDesignClover, 8),
	})
	setJassHand(g, 3, []*domain.Card{jcard(domain.CardDesignDiamond, 1)})
	g.ResolveWeisForTest(domain.CardDesignHeart, 0)
	assert.Equal(t, 200, g.GetRoundWeisPoints(1))
	assert.Equal(t, 0, g.GetRoundWeisPoints(0))
}

// --- Stöck: trump K+Q held by a player = +20 for their team ---

func TestJass_Stock(t *testing.T) {
	g := newTestJass()
	setJassHand(g, 2, []*domain.Card{
		jcard(domain.CardDesignHeart, 13), jcard(domain.CardDesignHeart, 12), // trump K+Q
		jcard(domain.CardDesignSpade, 1),
	})
	setJassHand(g, 0, []*domain.Card{jcard(domain.CardDesignSpade, 6)})
	setJassHand(g, 1, []*domain.Card{jcard(domain.CardDesignClover, 6)})
	setJassHand(g, 3, []*domain.Card{jcard(domain.CardDesignDiamond, 6)})
	g.ResolveStockForTest(domain.CardDesignHeart)
	assert.Equal(t, domain.JassStockBonus, g.GetRoundStockPoints(0)) // P2 is team 0
	assert.Equal(t, 0, g.GetRoundStockPoints(1))
}

// --- Schieben: forehand passes choice to partner ---

func TestJass_Schieben(t *testing.T) {
	g := domain.NewDefaultJass()
	g.Reset()
	// Force human (P0) to be forehand bidder.
	g.SetDealerIdxForTest(3) // forehand = (3+1)%4 = 0
	g.RebeginRoundForTest()
	assert.Equal(t, domain.JassPhaseBidTrump, g.GetPhase())
	assert.Equal(t, 0, g.GetBidPlayerIdx())
	assert.True(t, g.IsHumanBidTurn())

	err := g.PlayerSchieben()
	assert.NoError(t, err)
	assert.True(t, g.GetSchieben())
	assert.Equal(t, domain.JassPhaseBidPartner, g.GetPhase())
	assert.Equal(t, 2, g.GetBidPlayerIdx()) // partner of P0

	// Partner P2 is CPU and must choose a trump (cannot schieben again).
	g.CpuBid()
	assert.Greater(t, g.GetTrumpSuit(), 0)
	assert.Equal(t, domain.JassPhasePlay, g.GetPhase())
}

// --- Round / match scoring to 1000 ---

func TestJass_ScoreRound_GameEnd(t *testing.T) {
	g := newTestJass()
	g.SetPhase(domain.JassPhaseRoundEnd)
	g.SetTeamScore(0, 995)
	g.AddRoundPointsForTest(0, 10) // pushes team 0 over 1000
	g.ScoreRound()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, domain.JassPhaseGameEnd, g.GetPhase())
	assert.Equal(t, 0, g.GetWinnerTeam())
}

// --- CPU play produces a valid move ---

func TestJass_CpuPlay_Valid(t *testing.T) {
	g := domain.NewDefaultJass()
	g.Reset()
	for g.GetPhase() == domain.JassPhaseBidTrump || g.GetPhase() == domain.JassPhaseBidPartner {
		if g.IsHumanBidTurn() {
			_ = g.PlayerChooseTrump(domain.CardDesignHeart)
		} else {
			g.CpuBid()
		}
	}
	assert.Equal(t, domain.JassPhasePlay, g.GetPhase())
	// Play out a full trick via CPU + forced human play.
	for i := 0; i < 4; i++ {
		if g.IsHumanTurn() {
			valid := g.GetValidPlayIndices(g.GetCurrentPlayerIdx())
			assert.NotEmpty(t, valid)
			_ = g.PlayerPlay(valid[0])
		} else {
			g.CpuPlay()
		}
	}
	assert.Equal(t, domain.JassPhaseTrickEnd, g.GetPhase())
	assert.Len(t, g.GetCurrentTrick(), 4)
}

// --- JSON round-trip ---

func TestJass_JSON_RoundTrip(t *testing.T) {
	g := domain.NewDefaultJass()
	g.Reset()
	data, err := json.Marshal(g)
	assert.NoError(t, err)

	var restored domain.Jass
	assert.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, g.GetPhase(), restored.GetPhase())
	assert.Equal(t, g.GetPlayerCnt(), restored.GetPlayerCnt())
}

func TestJass_Unmarshal_Validation(t *testing.T) {
	var g domain.Jass
	// Invalid phase.
	assert.Error(t, json.Unmarshal([]byte(`{"ph":99,"pl":[{},{},{},{}]}`), &g))
	// Invalid trump suit.
	assert.Error(t, json.Unmarshal([]byte(`{"ph":2,"ts":9,"pl":[{},{},{},{}]}`), &g))
	// Wrong player count.
	assert.Error(t, json.Unmarshal([]byte(`{"ph":2,"pl":[{},{}]}`), &g))
	// trumpSuit=0 (unset) is allowed during bidding.
	assert.NoError(t, json.Unmarshal([]byte(`{"ph":0,"ts":0,"pl":[{},{},{},{}]}`), &g))
}

// TestJass_GetHint_Coverage exercises GetHint across the bid and play phases
// (including the not-your-turn early returns), covering playHintReason.
func TestJass_GetHint_Coverage(t *testing.T) {
	g := newTestJass()
	g.Reset()

	// BidTrump, human (forehand) to bid -> a hint (suit or schieben).
	g.SetPhase(domain.JassPhaseBidTrump)
	g.SetBidPlayerIdx(0)
	assert.NotNil(t, g.GetHint())

	// BidTrump but not the human's turn -> nil.
	g.SetBidPlayerIdx(1)
	assert.Nil(t, g.GetHint())

	// BidPartner, human to bid -> suit hint.
	g.SetPhase(domain.JassPhaseBidPartner)
	g.SetBidPlayerIdx(0)
	hp := g.GetHint()
	assert.NotNil(t, hp)
	assert.NotNil(t, hp.Suit)

	// Play, human's turn -> card-index hint (covers playHintReason).
	g.SetTrumpSuit(domain.CardDesignSpade)
	g.SetPhase(domain.JassPhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	hpl := g.GetHint()
	assert.NotNil(t, hpl)
	assert.NotNil(t, hpl.CardIndex)

	// Play but not the human's turn -> nil.
	g.SetCurrentPlayerIdx(1)
	assert.Nil(t, g.GetHint())
}

// TestJass_Getters_Coverage exercises the simple accessors and the
// NextTrick/NextRound phase advancers.
func TestJass_Getters_Coverage(t *testing.T) {
	g := newTestJass()
	g.Reset()

	_ = g.GetForehandIdx()
	_ = g.GetDealerIdx()
	_ = g.GetActionLog()
	_ = g.GetTeamScore(0)
	_ = g.GetRoundNumber()
	_ = g.GetTrickNumber()
	_ = g.GetMakerTeam()
	_ = g.GetMakerPlayerIdx()
	_ = g.GetLeadPlayerIdx()
	_ = g.GetBidPlayerIdx()
	assert.Equal(t, 1000, g.GetConfig().TargetScore)

	// NextTrick advances from TrickEnd to Play.
	g.SetPhase(domain.JassPhaseTrickEnd)
	g.SetLeadPlayerIdx(2)
	g.NextTrick()
	assert.Equal(t, domain.JassPhasePlay, g.GetPhase())

	// NextRound re-deals when the round has ended.
	g.SetPhase(domain.JassPhaseRoundEnd)
	before := g.GetRoundNumber()
	g.NextRound()
	assert.Equal(t, before+1, g.GetRoundNumber())
}
