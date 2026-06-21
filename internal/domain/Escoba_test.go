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

func newTestEscoba(humanSeat0 bool) *domain.Escoba {
	players := make([]*domain.ScopaPlayer, domain.EscobaPlayerCnt)
	for i := range players {
		players[i] = domain.NewScopaPlayer(i == 0 && humanSeat0)
	}
	return domain.NewEscoba(domain.NewTrumpCardsScopa(), players, domain.DefaultEscobaConfig())
}

func ebCard(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

func ebSetHand(p *domain.ScopaPlayer, cards ...*domain.Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func TestEscobaConfig_Validate(t *testing.T) {
	cfg := domain.DefaultEscobaConfig()
	assert.NoError(t, cfg.Validate())
	assert.Equal(t, 10, cfg.TargetScore)
	assert.Error(t, domain.EscobaConfig{CpuDifficulty: 99, TargetScore: 10}.Validate())
	assert.Error(t, domain.EscobaConfig{CpuDifficulty: 0, TargetScore: 0}.Validate())
}

func TestNewDefaultEscoba(t *testing.T) {
	e := domain.NewDefaultEscoba()
	require.NotNil(t, e)
	assert.Equal(t, domain.EscobaPlayerCnt, e.GetPlayerCnt())
	assert.True(t, e.GetPlayer(0).GetIsHuman())
	assert.Equal(t, -1, e.GetWinnerIdx())
	assert.Nil(t, e.GetPlayer(9))
}

func TestEscoba_ResetDeals4TableAnd3Each(t *testing.T) {
	e := newTestEscoba(true)
	e.Reset()
	assert.Equal(t, domain.EscobaPhasePlayerTurn, e.GetPhase())
	assert.Equal(t, domain.EscobaInitialTable, len(e.GetTableCards()))
	for i := 0; i < domain.EscobaPlayerCnt; i++ {
		assert.Equal(t, domain.EscobaPackSize, e.GetPlayer(i).GetCardsSize(), "player %d", i)
	}
	// 40 - 4 table - 12 dealt = 24 in stock.
	assert.Equal(t, 24, e.GetStockRemaining())
}

func TestEscobaCaptures_SumTo15(t *testing.T) {
	// Played 7 (value 7) + table {5,3} = 15 -> capture [5,3]; a lone 6 stays.
	table := []*domain.Card{ebCard(domain.CardDesignSpade, 5), ebCard(domain.CardDesignHeart, 3), ebCard(domain.CardDesignDiamond, 6)}
	caps := domain.EscobaCaptures(ebCard(domain.CardDesignClover, 7), table)
	require.NotEmpty(t, caps)
	// figure value: King(13) -> 10, so K + a lone 5 = 15.
	caps2 := domain.EscobaCaptures(ebCard(domain.CardDesignClover, 13), []*domain.Card{ebCard(domain.CardDesignSpade, 5)})
	require.Len(t, caps2, 1)
	assert.Nil(t, domain.EscobaCaptures(nil, table))
}

func TestEscoba_CaptureAndPlace(t *testing.T) {
	e := newTestEscoba(true)
	e.SetPhase(domain.EscobaPhasePlayerTurn)
	e.SetCurrentTurn(0)
	e.SetTableCards([]*domain.Card{ebCard(domain.CardDesignSpade, 5), ebCard(domain.CardDesignHeart, 3), ebCard(domain.CardDesignDiamond, 6)})
	ebSetHand(e.GetPlayer(0), ebCard(domain.CardDesignClover, 7), ebCard(domain.CardDesignClover, 2))
	e.GetPlayer(1).AddCard(ebCard(domain.CardDesignSpade, 1)) // keep round going
	// invalid capture (5+6=11, +7=18 != 15) -> error
	assert.Error(t, e.PlayerPlay(0, []int{0, 2}))
	// valid: 7 + 5 + 3 = 15
	require.NoError(t, e.PlayerPlay(0, []int{0, 1}))
	assert.Equal(t, 1, len(e.GetTableCards())) // only the 6 remains
	assert.Equal(t, 3, e.GetPlayer(0).CapturedCount())
}

func TestEscoba_PlaceWhenNoFifteen(t *testing.T) {
	e := newTestEscoba(true)
	e.SetPhase(domain.EscobaPhasePlayerTurn)
	e.SetCurrentTurn(0)
	e.SetTableCards([]*domain.Card{ebCard(domain.CardDesignSpade, 5)})
	ebSetHand(e.GetPlayer(0), ebCard(domain.CardDesignClover, 2), ebCard(domain.CardDesignHeart, 4))
	e.GetPlayer(1).AddCard(ebCard(domain.CardDesignSpade, 1))
	require.NoError(t, e.PlayerPlay(0, nil)) // 2 cannot make 15 with a lone 5 -> placed
	assert.Equal(t, 2, len(e.GetTableCards()))
}

func TestEscoba_ForcedCapture(t *testing.T) {
	// A 15-combo (K=10 + 5) exists on the table, so laying (nil) is illegal and
	// must not consume the played card.
	e := newTestEscoba(true)
	e.SetPhase(domain.EscobaPhasePlayerTurn)
	e.SetCurrentTurn(0)
	e.SetTableCards([]*domain.Card{ebCard(domain.CardDesignSpade, 5)})
	ebSetHand(e.GetPlayer(0), ebCard(domain.CardDesignHeart, 13), ebCard(domain.CardDesignClover, 2))
	e.GetPlayer(1).AddCard(ebCard(domain.CardDesignSpade, 1))
	assert.Error(t, e.PlayerPlay(0, nil)) // must capture the K+5 -> error
	assert.Equal(t, 2, e.GetPlayer(0).GetCardsSize(), "hand card must not be lost on a rejected lay")
	assert.Equal(t, 1, len(e.GetTableCards()))
}

func TestEscoba_EscobaSweep(t *testing.T) {
	e := newTestEscoba(true)
	e.SetPhase(domain.EscobaPhasePlayerTurn)
	e.SetCurrentTurn(0)
	e.SetTableCards([]*domain.Card{ebCard(domain.CardDesignSpade, 5)})
	ebSetHand(e.GetPlayer(0), ebCard(domain.CardDesignHeart, 13), ebCard(domain.CardDesignClover, 2)) // K worth 10
	e.GetPlayer(1).AddCard(ebCard(domain.CardDesignSpade, 1))                                         // round not over
	require.NoError(t, e.PlayerPlay(0, []int{0}))                                                     // 10 + 5 = 15, sweeps table
	assert.Equal(t, 1, e.GetPlayer(0).GetScopaCount())
}

// newDrainedEscoba builds a game whose stock is already empty, so that emptying
// all hands triggers finishRound (Escoba's round ends only when hands AND stock
// are exhausted).
func newDrainedEscoba() *domain.Escoba {
	d := domain.NewTrumpCardsScopa()
	for d.GetRemainingCount() > 0 {
		d.DrawCard()
	}
	players := make([]*domain.ScopaPlayer, domain.EscobaPlayerCnt)
	for i := range players {
		players[i] = domain.NewScopaPlayer(i == 0)
	}
	return domain.NewEscoba(d, players, domain.DefaultEscobaConfig())
}

func TestEscoba_RoundScoring(t *testing.T) {
	e := newDrainedEscoba()
	e.SetCurrentTurn(0)
	e.SetPhase(domain.EscobaPhasePlayerTurn)
	// Pre-load seat 0 with the A♠, 7♠, several oros and sevens, lots of cards.
	pile := []*domain.Card{ebCard(domain.CardDesignSpade, 1), ebCard(domain.CardDesignSpade, 7),
		ebCard(domain.CardDesignDiamond, 7), ebCard(domain.CardDesignDiamond, 2), ebCard(domain.CardDesignDiamond, 3)}
	e.GetPlayer(0).AddCaptured(pile)
	// Final play -> finishRound. seat 0 captures to set lastCapture and empty its hand.
	e.SetTableCards([]*domain.Card{ebCard(domain.CardDesignClover, 11)}) // J -> value 8
	ebSetHand(e.GetPlayer(0), ebCard(domain.CardDesignClover, 7))        // 7 + 8 = 15
	require.NoError(t, e.PlayerPlay(0, []int{0}))
	assert.True(t, e.GetPhase() == domain.EscobaPhaseRoundEnd || e.GetGameEndFlag())
	require.NotNil(t, e.GetLastRoundDetail())
	assert.Equal(t, 0, e.GetLastRoundDetail().AceEsp)  // seat 0 holds A♠
	assert.Equal(t, 0, e.GetLastRoundDetail().SeteEsp) // seat 0 holds 7♠
	assert.Greater(t, e.GetPlayer(0).GetTotalScore(), 0)
}

func TestEscoba_FullCpuGame(t *testing.T) {
	e := newTestEscoba(false)
	e.Reset()
	guard := 0
	for !e.GetGameEndFlag() && guard < 200000 {
		guard++
		switch e.GetPhase() {
		case domain.EscobaPhasePlayerTurn:
			e.CpuPlay()
		case domain.EscobaPhaseRoundEnd:
			e.NextRound()
		}
	}
	assert.True(t, e.GetGameEndFlag(), "match must terminate")
	assert.GreaterOrEqual(t, e.GetWinnerIdx(), 0)
	assert.GreaterOrEqual(t, e.GetPlayer(e.GetWinnerIdx()).GetTotalScore(), e.GetConfig().TargetScore)
}

func TestEscoba_GettersAndNextRoundNoop(t *testing.T) {
	e := newTestEscoba(true)
	e.Reset()
	assert.Equal(t, 0, e.GetDealerIdx())
	assert.Equal(t, -1, e.GetLastCaptureIdx())
	assert.NotNil(t, e.GetActionLog())
	cfg := domain.EscobaConfig{CpuDifficulty: domain.EscobaCpuDifficultyHard, TargetScore: 21}
	e.SetConfig(cfg)
	assert.Equal(t, cfg, e.GetConfig())
	e.SetPhase(domain.EscobaPhasePlayerTurn)
	r := e.GetRoundNumber()
	e.NextRound() // no-op outside roundEnd
	assert.Equal(t, r, e.GetRoundNumber())
	assert.Nil(t, e.GetValidCaptures(99))
}

func TestEscoba_JSONRoundTrip(t *testing.T) {
	e := newTestEscoba(true)
	e.Reset()
	data, err := json.Marshal(e)
	require.NoError(t, err)
	var e2 domain.Escoba
	require.NoError(t, json.Unmarshal(data, &e2))
	assert.Equal(t, e.GetPhase(), e2.GetPhase())
	assert.Equal(t, e.GetPlayerCnt(), e2.GetPlayerCnt())
	assert.Equal(t, e.GetStockRemaining(), e2.GetStockRemaining())
}

func TestEscoba_UnmarshalRejectsInvalid(t *testing.T) {
	e := newTestEscoba(true)
	e.Reset()
	data, err := json.Marshal(e)
	require.NoError(t, err)
	tampered := strings.Replace(string(data), `"ct":1`, `"ct":9`, 1)
	if tampered == string(data) {
		tampered = strings.Replace(string(data), `"ct":2`, `"ct":9`, 1)
	}
	require.NotEqual(t, string(data), tampered)
	var bad domain.Escoba
	assert.Error(t, bad.UnmarshalJSON([]byte(tampered)))
	var bad2 domain.Escoba
	assert.Error(t, bad2.UnmarshalJSON([]byte(`{"ps":[null]}`)))
	var bad3 domain.Escoba
	assert.Error(t, bad3.UnmarshalJSON([]byte(`not json`)))
}
