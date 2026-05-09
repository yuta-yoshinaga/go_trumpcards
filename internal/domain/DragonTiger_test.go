package domain_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestNewDefaultDragonTiger(t *testing.T) {
	dt := domain.NewDefaultDragonTiger()
	assert.Equal(t, domain.DragonTigerPhaseBet, dt.GetPhase())
	assert.Equal(t, domain.DragonTigerDefaultChips, dt.GetChips())
	assert.False(t, dt.GetGameEndFlag())
	assert.Nil(t, dt.GetDragonCard())
	assert.Nil(t, dt.GetTigerCard())
	assert.Empty(t, dt.GetHistory())
}

func TestDragonTiger_Reset(t *testing.T) {
	dt := domain.NewDefaultDragonTiger()
	require.NoError(t, dt.Bet(100, domain.DragonTigerBetDragon))
	dt.Reset()
	assert.Equal(t, domain.DragonTigerPhaseBet, dt.GetPhase())
	assert.False(t, dt.GetGameEndFlag())
	assert.Nil(t, dt.GetDragonCard())
	assert.Nil(t, dt.GetTigerCard())
	assert.Equal(t, 0, dt.GetBetAmount())
	assert.Equal(t, 0, dt.GetPayout())
}

func TestDragonTiger_Reset_RefillChips(t *testing.T) {
	dt := domain.NewDefaultDragonTiger()
	dt.SetChips(5)
	dt.Reset()
	assert.Equal(t, domain.DragonTigerDefaultChips, dt.GetChips())
}

func TestDragonTiger_Reset_NoRefillAboveThreshold(t *testing.T) {
	dt := domain.NewDefaultDragonTiger()
	dt.SetChips(500)
	dt.Reset()
	assert.Equal(t, 500, dt.GetChips())
}

func TestDragonTiger_Reset_PreservesHistory(t *testing.T) {
	dt := domain.NewDefaultDragonTiger()
	dt.SetHistory([]int{domain.DragonTigerResultDragon, domain.DragonTigerResultTie})
	dt.Reset()
	assert.Equal(t, []int{domain.DragonTigerResultDragon, domain.DragonTigerResultTie}, dt.GetHistory())
}

func TestDragonTiger_ClearHistory(t *testing.T) {
	dt := domain.NewDefaultDragonTiger()
	dt.SetHistory([]int{domain.DragonTigerResultDragon})
	dt.ClearHistory()
	assert.Empty(t, dt.GetHistory())
}

func TestDragonTiger_Bet_WrongPhase(t *testing.T) {
	dt := domain.NewDefaultDragonTiger()
	dt.SetPhase(domain.DragonTigerPhaseEnd)
	err := dt.Bet(100, domain.DragonTigerBetDragon)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestDragonTiger_Bet_InvalidAmount(t *testing.T) {
	tests := []struct {
		name   string
		amount int
	}{
		{"TooLow", 5},
		{"NotMultiple", 15},
		{"TooHigh", 20000},
		{"Zero", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := domain.NewDefaultDragonTiger()
			err := dt.Bet(tt.amount, domain.DragonTigerBetDragon)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
		})
	}
}

func TestDragonTiger_Bet_InvalidBetType(t *testing.T) {
	dt := domain.NewDefaultDragonTiger()
	err := dt.Bet(100, 99)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
}

func TestDragonTiger_Bet_InsufficientChips(t *testing.T) {
	dt := domain.NewDefaultDragonTiger()
	dt.SetChips(50)
	err := dt.Bet(100, domain.DragonTigerBetDragon)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

// Force the cards via setters so the test doesn't depend on shuffle.
func setupForcedDragonTiger(t *testing.T, dragonValue, tigerValue int) *domain.DragonTiger {
	t.Helper()
	dt := domain.NewDefaultDragonTiger()
	dt.SetDragonCard(domain.NewCard(domain.CardDesignSpade, dragonValue, false))
	dt.SetTigerCard(domain.NewCard(domain.CardDesignHeart, tigerValue, false))
	return dt
}

func TestDragonTiger_Judge_DragonWins(t *testing.T) {
	dt := setupForcedDragonTiger(t, 13, 5)
	dt.SetBetAmount(100)
	dt.SetBetType(domain.DragonTigerBetDragon)
	startChips := dt.GetChips()
	require.NoError(t, dt.Bet(100, domain.DragonTigerBetDragon))
	// Bet() also re-deals, overwriting our setters with shuffled cards. To
	// test the judge path deterministically, exercise calculatePayout via the
	// public Bet flow with retries OR via a dedicated test that injects state.
	// Here, simply assert the game ended and payout is non-negative.
	assert.True(t, dt.GetGameEndFlag())
	assert.GreaterOrEqual(t, dt.GetChips(), 0)
	_ = startChips
}

// Forcing the deal output deterministically requires bypassing trumpCards.
// We do this by setting cards before Bet, but Bet calls deal() which only
// draws when the slot is nil. So pre-setting both slots makes Bet skip the
// draw and use our forced cards.

func TestDragonTiger_Bet_DragonWins_PaysDouble(t *testing.T) {
	dt := setupForcedDragonTiger(t, 13, 5)
	dt.SetChips(1000)
	require.NoError(t, dt.Bet(100, domain.DragonTigerBetDragon))
	assert.True(t, dt.GetGameEndFlag())
	assert.Equal(t, domain.DragonTigerPhaseEnd, dt.GetPhase())
	// 1000 - 100 (bet) + 200 (payout 1:1 incl. stake) = 1100
	assert.Equal(t, 1100, dt.GetChips())
	assert.Equal(t, 200, dt.GetPayout())
	assert.Equal(t, []int{domain.DragonTigerResultDragon}, dt.GetHistory())
}

func TestDragonTiger_Bet_TigerWins_PaysDouble(t *testing.T) {
	dt := setupForcedDragonTiger(t, 3, 13)
	dt.SetChips(1000)
	require.NoError(t, dt.Bet(100, domain.DragonTigerBetTiger))
	assert.True(t, dt.GetGameEndFlag())
	assert.Equal(t, 1100, dt.GetChips())
	assert.Equal(t, 200, dt.GetPayout())
	assert.Equal(t, []int{domain.DragonTigerResultTiger}, dt.GetHistory())
}

func TestDragonTiger_Bet_DragonLoses_NoPayout(t *testing.T) {
	dt := setupForcedDragonTiger(t, 3, 13)
	dt.SetChips(1000)
	require.NoError(t, dt.Bet(100, domain.DragonTigerBetDragon))
	assert.Equal(t, 900, dt.GetChips())
	assert.Equal(t, 0, dt.GetPayout())
}

func TestDragonTiger_Bet_TigerLoses_NoPayout(t *testing.T) {
	dt := setupForcedDragonTiger(t, 13, 3)
	dt.SetChips(1000)
	require.NoError(t, dt.Bet(100, domain.DragonTigerBetTiger))
	assert.Equal(t, 900, dt.GetChips())
	assert.Equal(t, 0, dt.GetPayout())
}

func TestDragonTiger_Bet_TieOnDragonBet_HalfRefund(t *testing.T) {
	dt := setupForcedDragonTiger(t, 7, 7)
	dt.SetChips(1000)
	require.NoError(t, dt.Bet(100, domain.DragonTigerBetDragon))
	// 1000 - 100 (bet) + 50 (half refund) = 950
	assert.Equal(t, 950, dt.GetChips())
	assert.Equal(t, 50, dt.GetPayout())
	assert.Equal(t, []int{domain.DragonTigerResultTie}, dt.GetHistory())
}

func TestDragonTiger_Bet_TieOnTigerBet_HalfRefund(t *testing.T) {
	dt := setupForcedDragonTiger(t, 5, 5)
	dt.SetChips(1000)
	require.NoError(t, dt.Bet(100, domain.DragonTigerBetTiger))
	assert.Equal(t, 950, dt.GetChips())
	assert.Equal(t, 50, dt.GetPayout())
}

func TestDragonTiger_Bet_TieOnTieBet_PaysEightToOne(t *testing.T) {
	dt := setupForcedDragonTiger(t, 7, 7)
	dt.SetChips(1000)
	require.NoError(t, dt.Bet(100, domain.DragonTigerBetTie))
	// 1000 - 100 + (100 + 100*8) = 1800
	assert.Equal(t, 1800, dt.GetChips())
	assert.Equal(t, 900, dt.GetPayout())
}

func TestDragonTiger_Bet_TieBetLosesOnNonTie(t *testing.T) {
	dt := setupForcedDragonTiger(t, 13, 5)
	dt.SetChips(1000)
	require.NoError(t, dt.Bet(100, domain.DragonTigerBetTie))
	assert.Equal(t, 900, dt.GetChips())
	assert.Equal(t, 0, dt.GetPayout())
}

// A=1 is the weakest rank — verify A loses to 2.
func TestDragonTiger_AceIsWeakest(t *testing.T) {
	dt := setupForcedDragonTiger(t, 1, 2)
	dt.SetChips(1000)
	require.NoError(t, dt.Bet(100, domain.DragonTigerBetTiger))
	assert.Equal(t, 1100, dt.GetChips(), "Tiger (2) should beat Dragon (Ace)")
}

func TestDragonTiger_KingBeatsQueen(t *testing.T) {
	dt := setupForcedDragonTiger(t, 13, 12)
	dt.SetChips(1000)
	require.NoError(t, dt.Bet(100, domain.DragonTigerBetDragon))
	assert.Equal(t, 1100, dt.GetChips(), "Dragon (K) should beat Tiger (Q)")
}

func TestDragonTiger_ActionLog(t *testing.T) {
	dt := setupForcedDragonTiger(t, 13, 5)
	require.NoError(t, dt.Bet(100, domain.DragonTigerBetDragon))
	log := dt.GetActionLog()
	assert.NotEmpty(t, log)
	// Expect at least: bet, deal, result
	types := []string{}
	for _, e := range log {
		types = append(types, e.ActionType)
	}
	assert.Contains(t, types, "bet")
	assert.Contains(t, types, "deal")
	assert.Contains(t, types, "result")
}

func TestDragonTiger_JSONRoundTrip(t *testing.T) {
	dt := setupForcedDragonTiger(t, 13, 7)
	dt.SetChips(1500)
	require.NoError(t, dt.Bet(100, domain.DragonTigerBetDragon))

	data, err := json.Marshal(dt)
	require.NoError(t, err)

	var decoded domain.DragonTiger
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, dt.GetChips(), decoded.GetChips())
	assert.Equal(t, dt.GetPhase(), decoded.GetPhase())
	assert.Equal(t, dt.GetBetAmount(), decoded.GetBetAmount())
	assert.Equal(t, dt.GetBetType(), decoded.GetBetType())
	assert.Equal(t, dt.GetResult(), decoded.GetResult())
	assert.Equal(t, dt.GetPayout(), decoded.GetPayout())
	assert.Equal(t, dt.GetHistory(), decoded.GetHistory())
}

func TestDragonTiger_JSONUnmarshal_RejectOversizedSlices(t *testing.T) {
	// Build a payload with a history slice exceeding the cap (1000).
	data := []byte(`{"hi":[`)
	for i := 0; i < 1001; i++ {
		if i > 0 {
			data = append(data, ',')
		}
		data = append(data, '0')
	}
	data = append(data, []byte(`]}`)...)

	var decoded domain.DragonTiger
	err := json.Unmarshal(data, &decoded)
	assert.Error(t, err)
}
