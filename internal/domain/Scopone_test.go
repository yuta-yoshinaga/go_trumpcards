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

func newTestScopone(humanSeat0 bool) *domain.Scopone {
	players := make([]*domain.ScopaPlayer, domain.ScoponePlayerCnt)
	for i := range players {
		players[i] = domain.NewScopaPlayer(i == 0 && humanSeat0)
	}
	return domain.NewScopone(domain.NewTrumpCardsScopa(), players, domain.DefaultScoponeConfig())
}

func spCard(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

func spSetHand(p *domain.ScopaPlayer, cards ...*domain.Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func TestScoponeConfig_Validate(t *testing.T) {
	cfg := domain.DefaultScoponeConfig()
	assert.NoError(t, cfg.Validate())
	assert.Equal(t, 11, cfg.TargetScore)
	assert.Error(t, domain.ScoponeConfig{CpuDifficulty: 99, TargetScore: 11}.Validate())
	assert.Error(t, domain.ScoponeConfig{CpuDifficulty: 0, TargetScore: 0}.Validate())
}

func TestScoponeDeck40(t *testing.T) {
	assert.Equal(t, 40, domain.NewTrumpCardsScopa().GetRemainingCount())
}

func TestScoponeTeamOf(t *testing.T) {
	assert.Equal(t, 0, domain.ScoponeTeamOf(0))
	assert.Equal(t, 1, domain.ScoponeTeamOf(1))
	assert.Equal(t, 0, domain.ScoponeTeamOf(2))
	assert.Equal(t, 1, domain.ScoponeTeamOf(3))
}

func TestNewDefaultScopone(t *testing.T) {
	s := domain.NewDefaultScopone()
	require.NotNil(t, s)
	assert.Equal(t, domain.ScoponePlayerCnt, s.GetPlayerCnt())
	assert.True(t, s.GetPlayer(0).GetIsHuman())
	assert.Equal(t, -1, s.GetWinnerTeam())
	assert.Nil(t, s.GetPlayer(9))
}

func TestScopone_ResetDealsAllCards(t *testing.T) {
	s := newTestScopone(true)
	s.Reset()
	assert.Equal(t, domain.ScoponePhasePlayerTurn, s.GetPhase())
	for i := 0; i < domain.ScoponePlayerCnt; i++ {
		assert.Equal(t, domain.ScoponeHandSize, s.GetPlayer(i).GetCardsSize(), "player %d", i)
	}
	assert.Empty(t, s.GetTableCards()) // Scopone starts with no table cards
	assert.Equal(t, 0, len(s.GetTableCards()))
}

func TestScopone_PlaceWhenNoCapture(t *testing.T) {
	s := newTestScopone(true)
	s.SetPhase(domain.ScoponePhasePlayerTurn)
	s.SetCurrentTurn(0)
	s.SetTableCards([]*domain.Card{spCard(domain.CardDesignSpade, 3)}) // a 3 on the table
	spSetHand(s.GetPlayer(0), spCard(domain.CardDesignClover, 9), spCard(domain.CardDesignHeart, 2))
	// 9 cannot capture a lone 3 -> placed on table.
	require.NoError(t, s.PlayerPlay(0, nil))
	assert.Equal(t, 2, len(s.GetTableCards()))
}

func TestScopone_ForcedCapture(t *testing.T) {
	s := newTestScopone(true)
	s.SetPhase(domain.ScoponePhasePlayerTurn)
	s.SetCurrentTurn(0)
	s.SetTableCards([]*domain.Card{spCard(domain.CardDesignSpade, 7)})
	spSetHand(s.GetPlayer(0), spCard(domain.CardDesignClover, 7), spCard(domain.CardDesignHeart, 2))
	// A capture is available (7 matches 7) -> placing (nil) must be rejected.
	assert.Error(t, s.PlayerPlay(0, nil))
	require.NoError(t, s.PlayerPlay(0, []int{0})) // capture the 7
	assert.Equal(t, 0, len(s.GetTableCards()))
}

func TestScopone_SumCapture(t *testing.T) {
	s := newTestScopone(true)
	s.SetPhase(domain.ScoponePhasePlayerTurn)
	s.SetCurrentTurn(0)
	s.SetTableCards([]*domain.Card{spCard(domain.CardDesignSpade, 3), spCard(domain.CardDesignHeart, 4), spCard(domain.CardDesignClover, 9)})
	// give seat 0 some extra cards so the round isn't over (no scopa on last play).
	spSetHand(s.GetPlayer(0), spCard(domain.CardDesignClover, 7), spCard(domain.CardDesignHeart, 2))
	s.GetPlayer(1).AddCard(spCard(domain.CardDesignSpade, 5)) // keep a non-empty hand elsewhere
	caps := s.GetValidCaptures(0)
	require.NotEmpty(t, caps) // 7 = 3+4
	require.NoError(t, s.PlayerPlay(0, []int{0, 1}))
	// 3 + 4 captured (the lone 9 stays); seat 0's pile holds the played 7 + the two captured.
	assert.Equal(t, 1, len(s.GetTableCards()))
	assert.Equal(t, 3, s.GetPlayer(0).CapturedCount())
	assert.Equal(t, 1, s.GetCurrentTurn()) // turn advanced to the next player
}

func TestScopone_ScopaSweep(t *testing.T) {
	s := newTestScopone(true)
	s.SetPhase(domain.ScoponePhasePlayerTurn)
	s.SetCurrentTurn(0)
	s.SetTableCards([]*domain.Card{spCard(domain.CardDesignSpade, 7)})
	// seat 0 keeps a second card and others have cards -> not the last play, so sweeping is a scopa.
	spSetHand(s.GetPlayer(0), spCard(domain.CardDesignClover, 7), spCard(domain.CardDesignHeart, 2))
	s.GetPlayer(1).AddCard(spCard(domain.CardDesignSpade, 5))
	require.NoError(t, s.PlayerPlay(0, []int{0}))
	assert.Equal(t, 1, s.GetPlayer(0).GetScopaCount())
}

func TestScopone_RoundScoringTeams(t *testing.T) {
	s := newTestScopone(true)
	s.SetCurrentTurn(0)
	s.SetPhase(domain.ScoponePhasePlayerTurn)
	// Pre-load team 0 (seat 0) with a large capture pile incl. the settebello (7 of diamonds).
	pile := []*domain.Card{spCard(domain.CardDesignDiamond, 7)} // settebello
	for v := 1; v <= 9; v++ {
		pile = append(pile, spCard(domain.CardDesignDiamond, v)) // lots of diamonds + cards
	}
	s.GetPlayer(0).AddCaptured(pile)
	// Drive the final play so finishRound runs: only seat 0 holds 1 card; capture to set lastCapture.
	s.SetTableCards([]*domain.Card{spCard(domain.CardDesignSpade, 7)})
	spSetHand(s.GetPlayer(0), spCard(domain.CardDesignClover, 7))
	require.NoError(t, s.PlayerPlay(0, []int{0})) // captures -> all hands empty -> finishRound
	assert.True(t, s.GetPhase() == domain.ScoponePhaseRoundEnd || s.GetGameEndFlag())
	// Team 0 should out-score team 1 (most cards, most diamonds, settebello).
	assert.Greater(t, s.GetTeamScore(0), s.GetTeamScore(1))
	require.NotNil(t, s.GetLastRoundDetail())
	assert.Equal(t, 0, s.GetLastRoundDetail().SettebelloTm)
}

func TestScopone_FullCpuGame(t *testing.T) {
	s := newTestScopone(false)
	s.Reset()
	guard := 0
	for !s.GetGameEndFlag() && guard < 200000 {
		guard++
		switch s.GetPhase() {
		case domain.ScoponePhasePlayerTurn:
			s.CpuPlay()
		case domain.ScoponePhaseRoundEnd:
			s.NextRound()
		}
	}
	assert.True(t, s.GetGameEndFlag(), "match must terminate")
	assert.Contains(t, []int{0, 1}, s.GetWinnerTeam())
	assert.GreaterOrEqual(t, s.GetTeamScore(s.GetWinnerTeam()), s.GetConfig().TargetScore)
}

func TestScopone_JSONRoundTrip(t *testing.T) {
	s := newTestScopone(true)
	s.Reset()
	data, err := json.Marshal(s)
	require.NoError(t, err)
	var s2 domain.Scopone
	require.NoError(t, json.Unmarshal(data, &s2))
	assert.Equal(t, s.GetPhase(), s2.GetPhase())
	assert.Equal(t, s.GetPlayerCnt(), s2.GetPlayerCnt())
}

func TestScopone_UnmarshalRejectsInvalid(t *testing.T) {
	s := newTestScopone(true)
	s.Reset()
	data, err := json.Marshal(s)
	require.NoError(t, err)
	tampered := strings.Replace(string(data), `"ct":1`, `"ct":9`, 1)
	if tampered == string(data) {
		tampered = strings.Replace(string(data), `"ct":2`, `"ct":9`, 1)
	}
	require.NotEqual(t, string(data), tampered)
	var bad domain.Scopone
	assert.Error(t, bad.UnmarshalJSON([]byte(tampered)))
	var bad2 domain.Scopone
	assert.Error(t, bad2.UnmarshalJSON([]byte(`{"ps":[null]}`)))
	var bad3 domain.Scopone
	assert.Error(t, bad3.UnmarshalJSON([]byte(`not json`)))
}
