package presenter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupCasinoHoldemWebMockDefaults(m *interfaces.MockCasinoHoldemGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.CasinoHoldemPhaseBet).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetCommunity").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteBet").Return(0).Maybe()
	m.On("GetBonusBet").Return(0).Maybe()
	m.On("GetCallBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetDealerQualify").Return(false).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetCallPayout").Return(0).Maybe()
	m.On("GetBonusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

func parseCasinoHoldemOutput(t *testing.T, jsonStr string) *controller.CasinoHoldemWebOutput {
	t.Helper()
	var out controller.CasinoHoldemWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

func TestCasinoHoldemWebPresenter_Output_BetPhase(t *testing.T) {
	p := new(CasinoHoldemWebPresenter)
	m := new(interfaces.MockCasinoHoldemGame)
	setupCasinoHoldemWebMockDefaults(m)

	result := parseCasinoHoldemOutput(t, p.Output(m, nil))
	assert.Equal(t, domain.CasinoHoldemPhaseBet, result.Phase)
	assert.Equal(t, 1000, result.Chips)
	assert.Empty(t, result.PlayerHand)
	assert.Empty(t, result.DealerHand)
	assert.Empty(t, result.Community)
	assert.Empty(t, result.Message)
}

func TestCasinoHoldemWebPresenter_Output_Error(t *testing.T) {
	p := new(CasinoHoldemWebPresenter)
	m := new(interfaces.MockCasinoHoldemGame)
	setupCasinoHoldemWebMockDefaults(m)

	result := parseCasinoHoldemOutput(t, p.Output(m, errors.New("test error")))
	assert.Equal(t, "test error", result.Message)
}

// フロップフェーズではディーラーは伏せられる
func TestCasinoHoldemWebPresenter_Output_Flop_DealerMasked(t *testing.T) {
	p := new(CasinoHoldemWebPresenter)
	m := new(interfaces.MockCasinoHoldemGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.CasinoHoldemPhaseFlop).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignSpade, 13, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignDiamond, 5, false),
	}).Maybe()
	m.On("GetCommunity").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 11, false),
		domain.NewCard(domain.CardDesignSpade, 12, false),
		domain.NewCard(domain.CardDesignSpade, 10, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetBonusBet").Return(0).Maybe()
	m.On("GetCallBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetDealerQualify").Return(false).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetCallPayout").Return(0).Maybe()
	m.On("GetBonusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := parseCasinoHoldemOutput(t, p.Output(m, nil))
	assert.Len(t, result.DealerHand, 2)
	for i := 0; i < 2; i++ {
		assert.Equal(t, "", result.DealerHand[i].Design)
		assert.Equal(t, 0, result.DealerHand[i].Value)
	}
}

// フォールド時もディーラーは公開しない
func TestCasinoHoldemWebPresenter_Output_Fold_DealerMasked(t *testing.T) {
	p := new(CasinoHoldemWebPresenter)
	m := new(interfaces.MockCasinoHoldemGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.CasinoHoldemPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignSpade, 13, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignDiamond, 5, false),
	}).Maybe()
	m.On("GetCommunity").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 11, false),
		domain.NewCard(domain.CardDesignSpade, 12, false),
		domain.NewCard(domain.CardDesignSpade, 10, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetBonusBet").Return(0).Maybe()
	m.On("GetCallBet").Return(0).Maybe() // フォールド時は callBet=0
	m.On("GetResult").Return(domain.GameResultLose).Maybe()
	m.On("GetDealerQualify").Return(false).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetCallPayout").Return(0).Maybe()
	m.On("GetBonusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := parseCasinoHoldemOutput(t, p.Output(m, nil))
	for i := 0; i < 2; i++ {
		assert.Equal(t, "", result.DealerHand[i].Design, "dealer hand should be masked on fold")
	}
	assert.Equal(t, "casinoholdem.result.fold", result.MessageCode)
}

func TestCasinoHoldemWebPresenter_Output_PlayerWins(t *testing.T) {
	p := new(CasinoHoldemWebPresenter)
	m := new(interfaces.MockCasinoHoldemGame)
	m.On("GetChips").Return(1500).Maybe()
	m.On("GetPhase").Return(domain.CasinoHoldemPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetCommunity").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetBonusBet").Return(0).Maybe()
	m.On("GetCallBet").Return(200).Maybe()
	m.On("GetResult").Return(domain.GameResultWin).Maybe()
	m.On("GetDealerQualify").Return(true).Maybe()
	m.On("GetAntePayout").Return(200).Maybe()
	m.On("GetCallPayout").Return(400).Maybe()
	m.On("GetBonusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(600).Maybe()
	m.On("GetPlayerHandRank").Return(1).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := parseCasinoHoldemOutput(t, p.Output(m, nil))
	assert.Equal(t, "Player wins!", result.Message)
	assert.Equal(t, "casinoholdem.result.playerWins", result.MessageCode)
	assert.True(t, result.DealerQualify)
}

func TestCasinoHoldemWebPresenter_Output_DealerWins(t *testing.T) {
	p := new(CasinoHoldemWebPresenter)
	m := new(interfaces.MockCasinoHoldemGame)
	m.On("GetChips").Return(700).Maybe()
	m.On("GetPhase").Return(domain.CasinoHoldemPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetCommunity").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetBonusBet").Return(0).Maybe()
	m.On("GetCallBet").Return(200).Maybe()
	m.On("GetResult").Return(domain.GameResultLose).Maybe()
	m.On("GetDealerQualify").Return(true).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetCallPayout").Return(0).Maybe()
	m.On("GetBonusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(1).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := parseCasinoHoldemOutput(t, p.Output(m, nil))
	assert.Equal(t, "Dealer wins!", result.Message)
	assert.Equal(t, "casinoholdem.result.dealerWins", result.MessageCode)
}

func TestCasinoHoldemWebPresenter_Output_Push(t *testing.T) {
	p := new(CasinoHoldemWebPresenter)
	m := new(interfaces.MockCasinoHoldemGame)
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.CasinoHoldemPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetCommunity").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetBonusBet").Return(0).Maybe()
	m.On("GetCallBet").Return(200).Maybe()
	m.On("GetResult").Return(domain.GameResultDraw).Maybe()
	m.On("GetDealerQualify").Return(true).Maybe()
	m.On("GetAntePayout").Return(100).Maybe()
	m.On("GetCallPayout").Return(200).Maybe()
	m.On("GetBonusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(300).Maybe()
	m.On("GetPlayerHandRank").Return(1).Maybe()
	m.On("GetDealerHandRank").Return(1).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := parseCasinoHoldemOutput(t, p.Output(m, nil))
	assert.Equal(t, "Push!", result.Message)
	assert.Equal(t, "casinoholdem.result.push", result.MessageCode)
}

func TestCasinoHoldemWebPresenter_Output_EndPhase_DealerVisibleAfterCall(t *testing.T) {
	p := new(CasinoHoldemWebPresenter)
	m := new(interfaces.MockCasinoHoldemGame)
	m.On("GetChips").Return(1500).Maybe()
	m.On("GetPhase").Return(domain.CasinoHoldemPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignClover, 1, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignDiamond, 5, false),
	}).Maybe()
	m.On("GetCommunity").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetBonusBet").Return(0).Maybe()
	m.On("GetCallBet").Return(200).Maybe() // コール経由でショーダウン
	m.On("GetResult").Return(domain.GameResultWin).Maybe()
	m.On("GetDealerQualify").Return(true).Maybe()
	m.On("GetAntePayout").Return(200).Maybe()
	m.On("GetCallPayout").Return(400).Maybe()
	m.On("GetBonusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(600).Maybe()
	m.On("GetPlayerHandRank").Return(1).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := parseCasinoHoldemOutput(t, p.Output(m, nil))
	assert.Len(t, result.DealerHand, 2)
	assert.NotEqual(t, "", result.DealerHand[0].Design)
	assert.NotEqual(t, 0, result.DealerHand[0].Value)
}

func TestCasinoHoldemWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(CasinoHoldemWebPresenter)
	m := new(interfaces.MockCasinoHoldemGame)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()

	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "entries")
}

func TestCasinoHoldemWebPresenter_HintOutput(t *testing.T) {
	m := new(interfaces.MockCasinoHoldemGame)
	setupCasinoHoldemWebMockDefaults(m)
	p := new(CasinoHoldemWebPresenter)
	// The web presenter computes hints client-side, so HintOutput mirrors Output.
	assert.Equal(t, p.Output(m, nil), p.HintOutput(m))
}
