//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newGaigel() *domain.Gaigel { return domain.NewDefaultGaigel() }

func TestGaigel_DeckAndPoints(t *testing.T) {
	g := newGaigel()
	deck := g.GetConfigDeckHelper()
	assert.Equal(t, 48, deck.GetTotalCount())

	// Card points.
	assert.Equal(t, 11, domain.GaigelCardPoints(domain.NewCard(domain.CardDesignSpade, 1, false)))
	assert.Equal(t, 10, domain.GaigelCardPoints(domain.NewCard(domain.CardDesignSpade, 10, false)))
	assert.Equal(t, 4, domain.GaigelCardPoints(domain.NewCard(domain.CardDesignSpade, 13, false)))
	assert.Equal(t, 3, domain.GaigelCardPoints(domain.NewCard(domain.CardDesignSpade, 12, false)))
	assert.Equal(t, 2, domain.GaigelCardPoints(domain.NewCard(domain.CardDesignSpade, 11, false)))
	assert.Equal(t, 0, domain.GaigelCardPoints(domain.NewCard(domain.CardDesignSpade, 7, false)))
	assert.Equal(t, 0, domain.GaigelCardPoints(nil))

	// Total = 240.
	total := (11 + 10 + 4 + 3 + 2 + 0) * 4 * 2
	assert.Equal(t, 240, total)
	assert.Equal(t, 240, domain.GaigelRoundCardPointsTotal)
}

func TestGaigel_RankOrder(t *testing.T) {
	// A>10>K>Q>J>7
	a := domain.GaigelRankOrder(domain.NewCard(domain.CardDesignSpade, 1, false))
	ten := domain.GaigelRankOrder(domain.NewCard(domain.CardDesignSpade, 10, false))
	k := domain.GaigelRankOrder(domain.NewCard(domain.CardDesignSpade, 13, false))
	q := domain.GaigelRankOrder(domain.NewCard(domain.CardDesignSpade, 12, false))
	j := domain.GaigelRankOrder(domain.NewCard(domain.CardDesignSpade, 11, false))
	seven := domain.GaigelRankOrder(domain.NewCard(domain.CardDesignSpade, 7, false))
	assert.True(t, a > ten && ten > k && k > q && q > j && j > seven)
	assert.Equal(t, 0, domain.GaigelRankOrder(nil))
}

func TestGaigel_ResetDeal(t *testing.T) {
	g := newGaigel()
	g.Reset()
	assert.Equal(t, domain.GaigelPhasePlay, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	// 4 players * 5 cards.
	totalHand := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		totalHand += g.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, 20, totalHand)
	assert.NotNil(t, g.GetTrumpCard())
	assert.True(t, g.GetTrumpSuit() >= 1 && g.GetTrumpSuit() <= 4)
	// 48 - 20 dealt - 1 trump card held = 27 stock.
	assert.Equal(t, 27, g.GetStockRemaining())
	assert.False(t, g.IsEndgame())
}

func TestGaigel_TrickWinner_DoubleDeckTie(t *testing.T) {
	g := newGaigel()
	g.Reset()
	g.SetTrumpSuit(domain.CardDesignHeart) // trump = heart, lead = spade
	// Two identical SPADE Aces: the earlier (player 0) wins the tie.
	a1 := domain.NewCard(domain.CardDesignSpade, 1, false)
	a2 := domain.NewCard(domain.CardDesignSpade, 1, false)
	low := domain.NewCard(domain.CardDesignSpade, 7, false)
	low2 := domain.NewCard(domain.CardDesignSpade, 7, false)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: a1},
		{PlayerIdx: 1, Card: low},
		{PlayerIdx: 2, Card: a2},
		{PlayerIdx: 3, Card: low2},
	})
	g.SetTrickNumber(1)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.GaigelPhaseTrickEnd)
	g.ResolveTrick()
	assert.Equal(t, 0, g.GetLeadPlayerIdx(), "earlier identical card should win")
}

func TestGaigel_TrumpBeatsNonTrump(t *testing.T) {
	g := newGaigel()
	g.Reset()
	g.SetTrumpSuit(domain.CardDesignDiamond)
	// Lead spade Ace, player 2 plays trump 7 -> trump wins.
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 10, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignDiamond, 7, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignSpade, 13, false)},
	})
	g.SetTrickNumber(1)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.GaigelPhaseTrickEnd)
	g.ResolveTrick()
	assert.Equal(t, 2, g.GetLeadPlayerIdx())
}

func TestGaigel_Marriage(t *testing.T) {
	g := newGaigel()
	g.Reset()
	g.SetTrumpSuit(domain.CardDesignHeart)
	// Give player 0 a non-trump marriage (spade K+Q) on lead.
	p0 := g.GetPlayer(0)
	p0.Reset()
	p0.AddCard(domain.NewCard(domain.CardDesignSpade, 12, false)) // Q idx0
	p0.AddCard(domain.NewCard(domain.CardDesignSpade, 13, false)) // K idx1
	p0.AddCard(domain.NewCard(domain.CardDesignClover, 7, false))
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	g.SetPhase(domain.GaigelPhasePlay)

	idxs := g.GetMarriageIndices(0)
	require.NotEmpty(t, idxs)

	err := g.PlayerDeclareMarriage(0) // declare via Q
	require.NoError(t, err)
	// Non-trump marriage = 20 to team 0.
	assert.Equal(t, 20, g.GetRoundMarriagePoints(0))
	// The Q was led.
	assert.Len(t, g.GetCurrentTrick(), 1)
	// Re-declaring the same suit is now blocked.
	assert.Empty(t, g.GetMarriageIndices(0))
}

func TestGaigel_RoyalMarriage(t *testing.T) {
	g := newGaigel()
	g.Reset()
	g.SetTrumpSuit(domain.CardDesignSpade)
	p0 := g.GetPlayer(0)
	p0.Reset()
	p0.AddCard(domain.NewCard(domain.CardDesignSpade, 12, false)) // trump Q
	p0.AddCard(domain.NewCard(domain.CardDesignSpade, 13, false)) // trump K
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	g.SetPhase(domain.GaigelPhasePlay)
	require.NoError(t, g.PlayerDeclareMarriage(0))
	assert.Equal(t, 40, g.GetRoundMarriagePoints(0))
}

func TestGaigel_MarriageRejectsNonStarter(t *testing.T) {
	g := newGaigel()
	g.Reset()
	g.SetTrumpSuit(domain.CardDesignHeart)
	p0 := g.GetPlayer(0)
	p0.Reset()
	p0.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false)) // not a K/Q
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	g.SetPhase(domain.GaigelPhasePlay)
	assert.Error(t, g.PlayerDeclareMarriage(0))
	assert.Error(t, g.PlayerDeclareMarriage(99)) // out of range
}

func TestGaigel_PhaseSwitch_Phase1Optional(t *testing.T) {
	g := newGaigel()
	g.Reset()
	// Phase 1 (stock not empty): following suit is optional — any card is valid.
	g.SetTrumpSuit(domain.CardDesignHeart)
	p0 := g.GetPlayer(0)
	p0.Reset()
	p0.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))  // lead suit
	p0.AddCard(domain.NewCard(domain.CardDesignClover, 7, false)) // off suit
	p0.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))  // trump
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 13, false)},
	})
	g.SetPhase(domain.GaigelPhasePlay)
	assert.False(t, g.IsEndgame())
	assert.Equal(t, 3, len(g.GetValidPlayIndices(0)), "phase 1 allows any card")
}

func TestGaigel_PhaseSwitch_Phase2MustFollow(t *testing.T) {
	// Drive real play until the stock + trump card are exhausted (endgame),
	// then verify must-follow restricts a player who can follow the led suit.
	g := newGaigel()
	g.Reset()
	guard := 0
	for !g.IsEndgame() && !g.GetGameEndFlag() && guard < 2000 {
		guard++
		switch g.GetPhase() {
		case domain.GaigelPhasePlay:
			if g.IsHumanTurn() {
				idxs := g.GetValidPlayIndices(g.GetCurrentPlayerIdx())
				require.NotEmpty(t, idxs)
				require.NoError(t, g.PlayerPlay(idxs[0]))
			} else {
				g.CpuPlay()
			}
		case domain.GaigelPhaseTrickEnd:
			g.ResolveTrick()
			g.NextTrick()
		default:
			guard = 2000
		}
	}
	require.True(t, g.IsEndgame(), "should reach endgame (stock drained)")

	// In endgame, construct a lead and assert a player holding the led suit
	// is restricted to following it.
	g.SetTrumpSuit(domain.CardDesignHeart)
	p := g.GetPlayer(g.GetCurrentPlayerIdx())
	p.Reset()
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))   // led suit
	p.AddCard(domain.NewCard(domain.CardDesignClover, 10, false)) // off suit
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: (g.GetCurrentPlayerIdx() + 3) % 4, Card: domain.NewCard(domain.CardDesignSpade, 7, false)},
	})
	g.SetPhase(domain.GaigelPhasePlay)
	valid := g.GetValidPlayIndices(g.GetCurrentPlayerIdx())
	assert.Equal(t, []int{0}, valid, "phase 2 forces following the led suit")
}

func TestGaigel_FullRoundFlow(t *testing.T) {
	g := newGaigel()
	g.Reset()
	guard := 0
	for !g.GetGameEndFlag() && guard < 2000 {
		guard++
		switch g.GetPhase() {
		case domain.GaigelPhasePlay:
			if g.IsHumanTurn() {
				idxs := g.GetValidPlayIndices(g.GetCurrentPlayerIdx())
				require.NotEmpty(t, idxs)
				require.NoError(t, g.PlayerPlay(idxs[0]))
			} else {
				g.CpuPlay()
			}
		case domain.GaigelPhaseTrickEnd:
			g.ResolveTrick()
			g.NextTrick()
		case domain.GaigelPhaseRoundEnd:
			g.ScoreRound()
			if !g.GetGameEndFlag() {
				g.NextRound()
			}
		case domain.GaigelPhaseGameEnd:
		}
	}
	// At least one round should have scored; total of both teams must be >0 eventually.
	assert.True(t, guard < 2000, "flow should terminate")
}

func TestGaigel_Getters(t *testing.T) {
	g := newGaigel()
	g.Reset()
	g.SetRoundNumber(3)
	assert.Equal(t, 3, g.GetRoundNumber())
	g.SetTrickNumber(2)
	assert.Equal(t, 2, g.GetTrickNumber())
	g.SetCurrentPlayerIdx(1)
	assert.Equal(t, 1, g.GetCurrentPlayerIdx())
	g.SetLeadPlayerIdx(2)
	assert.Equal(t, 2, g.GetLeadPlayerIdx())
	g.SetDealerIdx(3)
	assert.Equal(t, 3, g.GetDealerIdx())
	g.SetTeamScore(0, 55)
	assert.Equal(t, 55, g.GetTeamScore(0))
	assert.Equal(t, 0, g.GetTeamScore(99))
	g.AddRoundPointsForTest(1, 12)
	assert.Equal(t, 12, g.GetRoundPoints(1))
	assert.Equal(t, 0, g.GetRoundPoints(99))
	assert.Equal(t, 0, g.GetRoundMarriagePoints(99))
	assert.Nil(t, g.GetPlayer(99))
	assert.Equal(t, 4, g.GetPlayerCnt())
	g.SetPhase(domain.GaigelPhaseTrickEnd)
	assert.Equal(t, domain.GaigelPhaseTrickEnd, g.GetPhase())
	assert.NotNil(t, g.GetConfig())
	g.SetConfig(domain.DefaultGaigelConfig())
	assert.Equal(t, 2, g.CardPointsPublic(domain.NewCard(domain.CardDesignSpade, 11, false)))
	assert.True(t, g.CardRankPublic(domain.NewCard(domain.CardDesignSpade, 1, false)) > 0)
}

func TestGaigel_GameEnd(t *testing.T) {
	g := newGaigel()
	g.Reset()
	g.SetTeamScore(0, 200)
	g.SetPhase(domain.GaigelPhaseRoundEnd)
	g.ScoreRound()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 0, g.GetWinnerTeam())
	assert.Equal(t, domain.GaigelPhaseGameEnd, g.GetPhase())
}

func TestGaigel_GetHint(t *testing.T) {
	g := newGaigel()
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	g.SetPhase(domain.GaigelPhasePlay)
	// Lead hint when it is human's turn.
	if g.GetPlayer(0).GetCardsSize() > 0 {
		h := g.GetHint()
		assert.NotNil(t, h)
	}
	// Not human turn -> nil.
	g.SetCurrentPlayerIdx(1)
	assert.Nil(t, g.GetHint())
	// Wrong phase -> nil.
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.GaigelPhaseTrickEnd)
	assert.Nil(t, g.GetHint())
}

func TestGaigel_GetHint_Marriage(t *testing.T) {
	g := newGaigel()
	g.Reset()
	g.SetTrumpSuit(domain.CardDesignHeart)
	p0 := g.GetPlayer(0)
	p0.Reset()
	p0.AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))
	p0.AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	g.SetPhase(domain.GaigelPhasePlay)
	h := g.GetHint()
	require.NotNil(t, h)
	assert.True(t, h.IsMarriage)
	assert.Equal(t, "marriage", h.Reason)
}

func TestGaigel_JSON_RoundTrip(t *testing.T) {
	g := newGaigel()
	g.Reset()
	data, err := json.Marshal(g)
	require.NoError(t, err)

	var g2 domain.Gaigel
	require.NoError(t, json.Unmarshal(data, &g2))
	assert.Equal(t, g.GetPhase(), g2.GetPhase())
	assert.Equal(t, g.GetTrumpSuit(), g2.GetTrumpSuit())
	assert.Equal(t, g.GetPlayerCnt(), g2.GetPlayerCnt())
}

func TestGaigel_JSON_Invalid(t *testing.T) {
	cases := []string{
		`{"ph":99}`,                      // bad phase
		`{"ph":0,"ts":9}`,                // bad trump suit
		`{"ph":0,"pl":[null,null,null]}`, // wrong player count
		`not json`,                       // malformed
		`{"ph":0,"pl":[null,null,null,null],"cp":100}`, // currentPlayerIdx out of range
		`{"ph":0,"pl":[null,null,null,null],"cp":-1}`,  // currentPlayerIdx negative
		`{"ph":0,"pl":[null,null,null,null],"di":99}`,  // dealerIdx out of range
		`{"ph":0,"pl":[null,null,null,null],"li":99}`,  // leadPlayerIdx out of range
		`{"ph":0,"pl":[null,null,null,null],"lw":99}`,  // lastTrickWinner out of range
		`{"ph":0,"pl":[null,null,null,null],"li":-2}`,  // leadPlayerIdx below -1 sentinel
		`{"ph":0,"pl":[null,null,null,null],"wt":99}`,  // winnerTeam out of range
	}
	for _, c := range cases {
		var g domain.Gaigel
		assert.Error(t, json.Unmarshal([]byte(c), &g), c)
	}

	// Out-of-range player team is rejected.
	var p domain.GaigelPlayer
	assert.Error(t, json.Unmarshal([]byte(`{"tm":99}`), &p))

	// Valid trumpSuit=0 (undecided) with 4 players is accepted.
	valid := domain.NewDefaultGaigel()
	b, _ := json.Marshal(valid)
	var g2 domain.Gaigel
	assert.NoError(t, json.Unmarshal(b, &g2))
}

func TestGaigel_PlayerJSON(t *testing.T) {
	p := domain.NewGaigelPlayer(true, 1)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	b, err := json.Marshal(p)
	require.NoError(t, err)
	var p2 domain.GaigelPlayer
	require.NoError(t, json.Unmarshal(b, &p2))
	assert.Equal(t, 1, p2.GetTeam())
	p2.ResetRound()
	assert.Equal(t, 0, p2.GetCardsSize())
}

func TestGaigelConfig_Validate(t *testing.T) {
	assert.NoError(t, domain.DefaultGaigelConfig().Validate())
	assert.Error(t, domain.GaigelConfig{CpuDifficulty: 99, TargetScore: 101}.Validate())
	assert.Error(t, domain.GaigelConfig{CpuDifficulty: 1, TargetScore: 0}.Validate())
}
