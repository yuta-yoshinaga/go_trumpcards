package domain_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestNewDefaultCasinoWar(t *testing.T) {
	cw := domain.NewDefaultCasinoWar()
	assert.Equal(t, domain.CasinoWarPhaseBet, cw.GetPhase())
	assert.Equal(t, domain.CasinoWarDefaultChips, cw.GetChips())
	assert.False(t, cw.GetGameEndFlag())
	assert.Nil(t, cw.GetPlayerCard())
	assert.Nil(t, cw.GetDealerCard())
	assert.Nil(t, cw.GetPlayerWarCard())
	assert.Nil(t, cw.GetDealerWarCard())
	assert.Empty(t, cw.GetBurnCards())
}

func TestCasinoWar_Reset(t *testing.T) {
	cw := domain.NewDefaultCasinoWar()
	require.NoError(t, cw.Bet(100))
	cw.SetPhase(domain.CasinoWarPhaseEnd)
	cw.Reset()
	assert.Equal(t, domain.CasinoWarPhaseBet, cw.GetPhase())
	assert.False(t, cw.GetGameEndFlag())
	assert.Nil(t, cw.GetPlayerCard())
	assert.Nil(t, cw.GetDealerCard())
	assert.Nil(t, cw.GetPlayerWarCard())
	assert.Nil(t, cw.GetDealerWarCard())
	assert.Empty(t, cw.GetBurnCards())
	assert.Equal(t, 0, cw.GetAnte())
	assert.Equal(t, 0, cw.GetWarBet())
	assert.Equal(t, 0, cw.GetTotalPayout())
}

func TestCasinoWar_Reset_RefillChips(t *testing.T) {
	cw := domain.NewDefaultCasinoWar()
	cw.SetChips(5)
	cw.Reset()
	assert.Equal(t, domain.CasinoWarDefaultChips, cw.GetChips())
}

func TestCasinoWar_Reset_NoRefillAboveThreshold(t *testing.T) {
	cw := domain.NewDefaultCasinoWar()
	cw.SetChips(500)
	cw.Reset()
	assert.Equal(t, 500, cw.GetChips())
}

func TestCasinoWar_Bet_WrongPhase(t *testing.T) {
	cw := domain.NewDefaultCasinoWar()
	cw.SetPhase(domain.CasinoWarPhaseTieDecision)
	err := cw.Bet(100)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestCasinoWar_Bet_InvalidAmount(t *testing.T) {
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
			cw := domain.NewDefaultCasinoWar()
			err := cw.Bet(tt.amount)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
		})
	}
}

func TestCasinoWar_Bet_InsufficientChips(t *testing.T) {
	cw := domain.NewDefaultCasinoWar()
	cw.SetChips(50)
	err := cw.Bet(100)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

func TestCasinoWar_Bet_Success_NaturalFlow(t *testing.T) {
	cw := domain.NewDefaultCasinoWar()
	require.NoError(t, cw.Bet(100))
	assert.Equal(t, 100, cw.GetAnte())
	assert.NotNil(t, cw.GetPlayerCard())
	assert.NotNil(t, cw.GetDealerCard())
	assert.Equal(t, domain.CasinoWarPhaseInitialDealt, cw.GetPhase())
	assert.Equal(t, domain.CasinoWarDefaultChips-100, cw.GetChips())
}

func TestCasinoWar_ResolveInitial_PlayerWins(t *testing.T) {
	cw := domain.NewDefaultCasinoWar()
	require.NoError(t, cw.Bet(100))
	cw.SetPlayerCard(domain.NewCard(domain.CardDesignSpade, 13, true)) // K
	cw.SetDealerCard(domain.NewCard(domain.CardDesignClover, 7, true))
	cw.ResolveInitial()
	assert.Equal(t, domain.CasinoWarPhaseEnd, cw.GetPhase())
	assert.True(t, cw.GetGameEndFlag())
	assert.Equal(t, domain.GameResultWin, cw.GetResult())
	// 1:1 on ante: returned ante + ante
	assert.Equal(t, 200, cw.GetTotalPayout())
	assert.Equal(t, domain.CasinoWarDefaultChips-100+200, cw.GetChips())
}

func TestCasinoWar_ResolveInitial_PlayerWins_AceHigh(t *testing.T) {
	cw := domain.NewDefaultCasinoWar()
	require.NoError(t, cw.Bet(100))
	cw.SetPlayerCard(domain.NewCard(domain.CardDesignSpade, 1, true))   // Ace
	cw.SetDealerCard(domain.NewCard(domain.CardDesignClover, 13, true)) // K
	cw.ResolveInitial()
	assert.Equal(t, domain.GameResultWin, cw.GetResult())
}

func TestCasinoWar_ResolveInitial_DealerWins(t *testing.T) {
	cw := domain.NewDefaultCasinoWar()
	require.NoError(t, cw.Bet(100))
	cw.SetPlayerCard(domain.NewCard(domain.CardDesignSpade, 5, true))
	cw.SetDealerCard(domain.NewCard(domain.CardDesignClover, 9, true))
	cw.ResolveInitial()
	assert.Equal(t, domain.CasinoWarPhaseEnd, cw.GetPhase())
	assert.True(t, cw.GetGameEndFlag())
	assert.Equal(t, domain.GameResultLose, cw.GetResult())
	assert.Equal(t, 0, cw.GetTotalPayout())
	assert.Equal(t, domain.CasinoWarDefaultChips-100, cw.GetChips())
}

func TestCasinoWar_ResolveInitial_Tie_EntersTieDecision(t *testing.T) {
	cw := domain.NewDefaultCasinoWar()
	require.NoError(t, cw.Bet(100))
	cw.SetPlayerCard(domain.NewCard(domain.CardDesignSpade, 7, true))
	cw.SetDealerCard(domain.NewCard(domain.CardDesignClover, 7, true))
	cw.ResolveInitial()
	assert.Equal(t, domain.CasinoWarPhaseTieDecision, cw.GetPhase())
	assert.False(t, cw.GetGameEndFlag())
	assert.Equal(t, 0, cw.GetTotalPayout())
}

func TestCasinoWar_ResolveInitial_NoOpWithoutCards(t *testing.T) {
	cw := domain.NewDefaultCasinoWar()
	cw.ResolveInitial()
	assert.Equal(t, domain.CasinoWarPhaseBet, cw.GetPhase())
}

func TestCasinoWar_Surrender_Success(t *testing.T) {
	cw := domain.NewDefaultCasinoWar()
	require.NoError(t, cw.Bet(100))
	cw.SetPlayerCard(domain.NewCard(domain.CardDesignSpade, 7, true))
	cw.SetDealerCard(domain.NewCard(domain.CardDesignClover, 7, true))
	cw.ResolveInitial()
	require.NoError(t, cw.Surrender())
	assert.Equal(t, domain.CasinoWarPhaseEnd, cw.GetPhase())
	assert.True(t, cw.GetGameEndFlag())
	assert.Equal(t, domain.GameResultLose, cw.GetResult())
	// Surrender refunds half ante: 50
	assert.Equal(t, 50, cw.GetTotalPayout())
	assert.Equal(t, domain.CasinoWarDefaultChips-100+50, cw.GetChips())
}

func TestCasinoWar_Surrender_WrongPhase(t *testing.T) {
	cw := domain.NewDefaultCasinoWar()
	err := cw.Surrender()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestCasinoWar_GoToWar_PlayerWins(t *testing.T) {
	cw := domain.NewDefaultCasinoWar()
	require.NoError(t, cw.Bet(100))
	cw.SetPlayerCard(domain.NewCard(domain.CardDesignSpade, 7, true))
	cw.SetDealerCard(domain.NewCard(domain.CardDesignClover, 7, true))
	cw.ResolveInitial()
	cw.SetPlayerWarCard(domain.NewCard(domain.CardDesignHeart, 13, true))
	cw.SetDealerWarCard(domain.NewCard(domain.CardDesignDiamond, 5, true))
	require.NoError(t, cw.GoToWar())
	assert.Equal(t, domain.CasinoWarPhaseEnd, cw.GetPhase())
	assert.True(t, cw.GetGameEndFlag())
	assert.Equal(t, domain.GameResultWin, cw.GetResult())
	assert.Equal(t, 100, cw.GetWarBet())
	// War win: ante push (100) + war bet returned (100) + 1:1 on war bet (100) = 300
	assert.Equal(t, 300, cw.GetTotalPayout())
	assert.Equal(t, domain.CasinoWarDefaultChips-100-100+300, cw.GetChips())
	assert.Len(t, cw.GetBurnCards(), domain.CasinoWarBurnCount)
}

func TestCasinoWar_GoToWar_TieAfterWar_CountedAsWin(t *testing.T) {
	cw := domain.NewDefaultCasinoWar()
	require.NoError(t, cw.Bet(100))
	cw.SetPlayerCard(domain.NewCard(domain.CardDesignSpade, 7, true))
	cw.SetDealerCard(domain.NewCard(domain.CardDesignClover, 7, true))
	cw.ResolveInitial()
	cw.SetPlayerWarCard(domain.NewCard(domain.CardDesignHeart, 9, true))
	cw.SetDealerWarCard(domain.NewCard(domain.CardDesignDiamond, 9, true))
	require.NoError(t, cw.GoToWar())
	assert.Equal(t, domain.GameResultWin, cw.GetResult())
	// Same payout as war win
	assert.Equal(t, 300, cw.GetTotalPayout())
}

func TestCasinoWar_GoToWar_PlayerLoses(t *testing.T) {
	cw := domain.NewDefaultCasinoWar()
	require.NoError(t, cw.Bet(100))
	cw.SetPlayerCard(domain.NewCard(domain.CardDesignSpade, 7, true))
	cw.SetDealerCard(domain.NewCard(domain.CardDesignClover, 7, true))
	cw.ResolveInitial()
	cw.SetPlayerWarCard(domain.NewCard(domain.CardDesignHeart, 5, true))
	cw.SetDealerWarCard(domain.NewCard(domain.CardDesignDiamond, 13, true))
	require.NoError(t, cw.GoToWar())
	assert.Equal(t, domain.CasinoWarPhaseEnd, cw.GetPhase())
	assert.True(t, cw.GetGameEndFlag())
	assert.Equal(t, domain.GameResultLose, cw.GetResult())
	assert.Equal(t, 0, cw.GetTotalPayout())
	assert.Equal(t, domain.CasinoWarDefaultChips-100-100, cw.GetChips())
}

func TestCasinoWar_GoToWar_WrongPhase(t *testing.T) {
	cw := domain.NewDefaultCasinoWar()
	err := cw.GoToWar()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestCasinoWar_GoToWar_InsufficientChips(t *testing.T) {
	cw := domain.NewDefaultCasinoWar()
	cw.SetChips(150)
	require.NoError(t, cw.Bet(100))
	cw.SetPlayerCard(domain.NewCard(domain.CardDesignSpade, 7, true))
	cw.SetDealerCard(domain.NewCard(domain.CardDesignClover, 7, true))
	cw.ResolveInitial()
	err := cw.GoToWar()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

func TestCasinoWar_GetActionLog(t *testing.T) {
	cw := domain.NewDefaultCasinoWar()
	require.NoError(t, cw.Bet(100))
	assert.NotEmpty(t, cw.GetActionLog())
}

func TestCasinoWar_JSONRoundTrip(t *testing.T) {
	cw := domain.NewDefaultCasinoWar()
	require.NoError(t, cw.Bet(100))
	cw.SetPlayerCard(domain.NewCard(domain.CardDesignSpade, 7, true))
	cw.SetDealerCard(domain.NewCard(domain.CardDesignClover, 7, true))
	cw.ResolveInitial()
	cw.SetPlayerWarCard(domain.NewCard(domain.CardDesignHeart, 13, true))
	cw.SetDealerWarCard(domain.NewCard(domain.CardDesignDiamond, 5, true))
	require.NoError(t, cw.GoToWar())

	data, err := json.Marshal(cw)
	require.NoError(t, err)
	var restored domain.CasinoWar
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, cw.GetPhase(), restored.GetPhase())
	assert.Equal(t, cw.GetChips(), restored.GetChips())
	assert.Equal(t, cw.GetAnte(), restored.GetAnte())
	assert.Equal(t, cw.GetWarBet(), restored.GetWarBet())
	assert.Equal(t, cw.GetResult(), restored.GetResult())
	assert.Equal(t, cw.GetTotalPayout(), restored.GetTotalPayout())
	assert.Equal(t, cw.GetGameEndFlag(), restored.GetGameEndFlag())
	assert.Len(t, restored.GetBurnCards(), domain.CasinoWarBurnCount)
}

func TestCasinoWar_UnmarshalJSON_InvalidData(t *testing.T) {
	var cw domain.CasinoWar
	err := cw.UnmarshalJSON([]byte("not json"))
	assert.Error(t, err)
}

func TestCasinoWar_UnmarshalJSON_NilFields(t *testing.T) {
	data := []byte(`{"tc":null,"pc":null,"dc":null,"pw":null,"dw":null,"bc":null,"ch":null,"an":0,"wb":0,"ps":1,"ge":false,"gr":0,"tp":0,"al":null}`)
	var cw domain.CasinoWar
	require.NoError(t, json.Unmarshal(data, &cw))
	assert.NotNil(t, cw.GetActionLog())
	assert.NotNil(t, cw.GetBurnCards())
}
