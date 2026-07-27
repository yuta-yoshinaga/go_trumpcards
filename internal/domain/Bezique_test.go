//go:build test

package domain_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestBezique(humanSeat0 bool) *domain.Bezique {
	players := []*domain.BeziquePlayer{
		domain.NewBeziquePlayer(humanSeat0),
		domain.NewBeziquePlayer(false),
	}
	return domain.NewBezique(domain.NewTrumpCardsBezique(), players, domain.DefaultBeziqueConfig())
}

func bzCard(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

func bzSetHand(p *domain.BeziquePlayer, cards ...*domain.Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func TestBeziqueConfig_Validate(t *testing.T) {
	cfg := domain.DefaultBeziqueConfig()
	assert.NoError(t, cfg.Validate())
	assert.Equal(t, domain.BeziqueDefaultTargetScore, cfg.TargetScore)
	assert.Error(t, domain.BeziqueConfig{CpuDifficulty: 99, TargetScore: 1000}.Validate())
	assert.Error(t, domain.BeziqueConfig{CpuDifficulty: domain.BeziqueCpuDifficultyNormal, TargetScore: 10}.Validate())
	assert.Error(t, domain.BeziqueConfig{CpuDifficulty: domain.BeziqueCpuDifficultyNormal, TargetScore: 99999}.Validate())
}

func TestBeziqueDeck64(t *testing.T) {
	d := domain.NewTrumpCardsBezique()
	assert.Equal(t, 64, d.GetRemainingCount())
}

func TestBeziqueCardPointsAndRank(t *testing.T) {
	assert.Equal(t, 11, domain.BeziqueCardPoints(bzCard(domain.CardDesignSpade, 1)))
	assert.Equal(t, 10, domain.BeziqueCardPoints(bzCard(domain.CardDesignSpade, 10)))
	assert.Equal(t, 0, domain.BeziqueCardPoints(bzCard(domain.CardDesignSpade, 7)))
	// A > 10 > K
	assert.Greater(t, domain.BeziqueRankOrder(bzCard(domain.CardDesignSpade, 1)), domain.BeziqueRankOrder(bzCard(domain.CardDesignSpade, 10)))
	assert.Greater(t, domain.BeziqueRankOrder(bzCard(domain.CardDesignSpade, 10)), domain.BeziqueRankOrder(bzCard(domain.CardDesignSpade, 13)))
	assert.Equal(t, 0, domain.BeziqueCardPoints(nil))
	assert.Equal(t, 0, domain.BeziqueRankOrder(nil))
}

func TestNewDefaultBezique(t *testing.T) {
	b := domain.NewDefaultBezique()
	require.NotNil(t, b)
	assert.Equal(t, domain.BeziquePlayerCnt, b.GetPlayerCnt())
	assert.True(t, b.GetPlayer(0).GetIsHuman())
	assert.False(t, b.GetPlayer(1).GetIsHuman())
	assert.Equal(t, -1, b.GetWinnerIdx())
	assert.Nil(t, b.GetPlayer(5))
}

func TestBezique_ResetDealsAndTurnsTrump(t *testing.T) {
	b := newTestBezique(true)
	b.Reset()
	assert.Equal(t, domain.BeziquePhasePlay, b.GetPhase())
	assert.Equal(t, 1, b.GetRoundNumber())
	assert.Equal(t, domain.BeziqueHandSize, b.GetPlayer(0).GetCardsSize())
	assert.Equal(t, domain.BeziqueHandSize, b.GetPlayer(1).GetCardsSize())
	require.NotNil(t, b.GetTrumpCard())
	assert.Equal(t, b.GetTrumpCard().GetDesign(), b.GetTrumpSuit())
	// 64 - 16 dealt - 1 trump card = 47 remaining in stock.
	assert.Equal(t, 47, b.GetStockRemaining())
}

func TestBezique_PlayMustFollowOnlyInEndgame(t *testing.T) {
	b := newTestBezique(true)
	b.SetPhase(domain.BeziquePhasePlay)
	b.SetTrumpSuit(domain.CardDesignSpade)
	b.SetCurrentPlayerIdx(0)
	// Phase 1 (stock present): any card is legal even off-suit.
	b.SetTrumpCard(bzCard(domain.CardDesignSpade, 1))
	b.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: bzCard(domain.CardDesignHeart, 10)}})
	bzSetHand(b.GetPlayer(0), bzCard(domain.CardDesignHeart, 13), bzCard(domain.CardDesignClover, 9))
	require.NoError(t, b.PlayerPlay(1)) // off-suit clover legal in phase 1
}

func TestBezique_TrickWinnerTrumpBeatsFailLead(t *testing.T) {
	b := newTestBezique(true)
	b.SetTrumpSuit(domain.CardDesignSpade)
	b.SetPhase(domain.BeziquePhasePlay)
	b.SetTrickNumber(1)
	b.SetTrumpCard(bzCard(domain.CardDesignSpade, 7)) // stock present → phase 1 meld after
	b.SetCurrentPlayerIdx(0)
	b.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: bzCard(domain.CardDesignHeart, 1)}}) // A♥ lead
	bzSetHand(b.GetPlayer(0), bzCard(domain.CardDesignSpade, 7))                                    // 7♠ trump
	require.NoError(t, b.PlayerPlay(0))
	// Trick resolved; seat 0 (trump) should have won → it's now seat 0's meld turn.
	assert.Equal(t, domain.BeziquePhaseMeld, b.GetPhase())
	assert.Equal(t, 0, b.GetCurrentPlayerIdx())
	assert.Equal(t, 1, b.GetPlayer(0).GetTrickCount())
}

func TestBezique_AvailableMelds(t *testing.T) {
	b := newTestBezique(true)
	b.SetPhase(domain.BeziquePhaseMeld)
	b.SetCurrentPlayerIdx(0)
	b.SetTrumpSuit(domain.CardDesignClover)
	// Hand: ♠K ♠Q (marriage 20), ♣K ♣Q (royal marriage 40, trump=clover), ♦J (with ♠Q → Bezique 40)
	bzSetHand(b.GetPlayer(0),
		bzCard(domain.CardDesignSpade, 13), bzCard(domain.CardDesignSpade, 12),
		bzCard(domain.CardDesignClover, 13), bzCard(domain.CardDesignClover, 12),
		bzCard(domain.CardDesignDiamond, 11))
	melds := b.GetAvailableMelds(0)
	require.NotEmpty(t, melds)
	var sawRoyal, sawMarriage, sawBezique bool
	for _, m := range melds {
		switch {
		case m.Type == domain.BeziqueMeldMarriage && m.Points == domain.BeziqueRoyalMarriagePoints:
			sawRoyal = true
		case m.Type == domain.BeziqueMeldMarriage && m.Points == domain.BeziqueMarriagePoints:
			sawMarriage = true
		case m.Type == domain.BeziqueMeldBezique:
			sawBezique = true
		}
	}
	assert.True(t, sawRoyal, "royal marriage")
	assert.True(t, sawMarriage, "plain marriage")
	assert.True(t, sawBezique, "bezique")
}

func TestBezique_FourAcesMeld(t *testing.T) {
	b := newTestBezique(true)
	b.SetPhase(domain.BeziquePhaseMeld)
	b.SetCurrentPlayerIdx(0)
	b.SetTrumpSuit(domain.CardDesignSpade)
	bzSetHand(b.GetPlayer(0),
		bzCard(domain.CardDesignSpade, 1), bzCard(domain.CardDesignClover, 1),
		bzCard(domain.CardDesignHeart, 1), bzCard(domain.CardDesignDiamond, 1))
	melds := b.GetAvailableMelds(0)
	require.Len(t, melds, 1)
	assert.Equal(t, domain.BeziqueMeldFourAces, melds[0].Type)
	assert.Equal(t, domain.BeziqueFourAcesPoints, melds[0].Points)

	startDeal := b.GetDealPoints(0)
	startMeld := b.GetDealMeldPoints(0)
	require.NoError(t, b.PlayerDeclareMeld(0))
	assert.Equal(t, startDeal+domain.BeziqueFourAcesPoints, b.GetDealPoints(0))
	// The meld points are also tracked separately for the score breakdown.
	assert.Equal(t, startMeld+domain.BeziqueFourAcesPoints, b.GetDealMeldPoints(0))
	// Same meld cannot be declared twice.
	b.SetPhase(domain.BeziquePhaseMeld)
	b.SetCurrentPlayerIdx(0)
	assert.Empty(t, b.GetAvailableMelds(0))
}

func TestBezique_SkipMeld(t *testing.T) {
	b := newTestBezique(true)
	b.Reset()
	// Drive to a meld phase by playing one full trick.
	b.SetCurrentPlayerIdx(0)
	b.SetCurrentTrick(nil)
	// seat 0 leads, seat 1 (CPU) follows automatically? Use explicit play:
	require.NoError(t, b.PlayerPlay(0))
	b.CpuPlay()
	// Whoever won, if it's seat 0's meld turn and human, skip is allowed.
	if b.GetPhase() == domain.BeziquePhaseMeld && b.GetCurrentPlayerIdx() == 0 {
		require.NoError(t, b.PlayerSkipMeld())
		assert.Equal(t, domain.BeziquePhasePlay, b.GetPhase())
	}
}

func TestBezique_SetMatchAndDealPoints(t *testing.T) {
	b := newTestBezique(true)
	b.SetDealPoints(0, 120)
	b.SetMatchScore(1, 150)
	assert.Equal(t, 120, b.GetDealPoints(0))
	assert.Equal(t, 150, b.GetMatchScore(1))
	assert.Equal(t, 0, b.GetDealPoints(9))
	assert.Equal(t, 0, b.GetMatchScore(9))
}

func TestBezique_FullCpuGame(t *testing.T) {
	b := newTestBezique(false)
	b.Reset()
	guard := 0
	for !b.GetGameEndFlag() && guard < 200000 {
		guard++
		switch b.GetPhase() {
		case domain.BeziquePhasePlay:
			b.CpuPlay()
		case domain.BeziquePhaseMeld:
			b.CpuMeld()
		case domain.BeziquePhaseRoundEnd:
			b.NextRound()
		}
	}
	assert.True(t, b.GetGameEndFlag(), "game must terminate")
	assert.Contains(t, []int{0, 1}, b.GetWinnerIdx())
	winner := b.GetWinnerIdx()
	assert.GreaterOrEqual(t, b.GetMatchScore(winner), b.GetConfig().TargetScore)
}

func TestBezique_Hint(t *testing.T) {
	b := newTestBezique(true)
	b.Reset()
	// After Reset the non-dealer (CPU seat 1) leads, so force the human's turn.
	b.SetPhase(domain.BeziquePhasePlay)
	b.SetCurrentPlayerIdx(0)
	b.SetCurrentTrick(nil)
	h := b.GetHint()
	require.NotNil(t, h)
	require.NotNil(t, h.CardIndex)
}

func TestBezique_JSONRoundTrip(t *testing.T) {
	b := newTestBezique(true)
	b.Reset()
	data, err := json.Marshal(b)
	require.NoError(t, err)
	var b2 domain.Bezique
	require.NoError(t, json.Unmarshal(data, &b2))
	assert.Equal(t, b.GetPhase(), b2.GetPhase())
	assert.Equal(t, b.GetTrumpSuit(), b2.GetTrumpSuit())
	assert.Equal(t, b.GetPlayerCnt(), b2.GetPlayerCnt())
	assert.Equal(t, b.GetStockRemaining(), b2.GetStockRemaining())
}

func TestBezique_UnmarshalRejectsInvalid(t *testing.T) {
	b := newTestBezique(true)
	b.Reset()
	data, err := json.Marshal(b)
	require.NoError(t, err)

	// Out-of-range current player index.
	tampered := strings.Replace(string(data), `"ci":1`, `"ci":9`, 1)
	if tampered == string(data) {
		tampered = strings.Replace(string(data), `"ci":0`, `"ci":9`, 1)
	}
	require.NotEqual(t, string(data), tampered)
	var bad domain.Bezique
	assert.Error(t, bad.UnmarshalJSON([]byte(tampered)))

	// Wrong player count.
	var bad2 domain.Bezique
	assert.Error(t, bad2.UnmarshalJSON([]byte(`{"ps":[null]}`)))

	// Malformed.
	var bad3 domain.Bezique
	assert.Error(t, bad3.UnmarshalJSON([]byte(`not json`)))
}
