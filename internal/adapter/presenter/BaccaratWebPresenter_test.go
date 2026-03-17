package presenter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupBaccaratWebMockDefaults(m *interfaces.MockBaccaratGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.BaccaratPhaseBet).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetBankerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetPlayerHandValue").Return(0).Maybe()
	m.On("GetBankerHandValue").Return(0).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetBetAmount").Return(0).Maybe()
	m.On("GetBetType").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetPayout").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetHistory").Return(([]int)(nil)).Maybe()
	m.On("GetPlayerPairBet").Return(0).Maybe()
	m.On("GetBankerPairBet").Return(0).Maybe()
	m.On("GetSideBetResults").Return(([]*domain.BacSideBetResult)(nil)).Maybe()
}

func parseBaccaratOutput(t *testing.T, jsonStr string) *controller.BaccaratWebOutput {
	t.Helper()
	var out controller.BaccaratWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

func TestBaccaratWebPresenter_Output_BetPhase(t *testing.T) {
	p := new(BaccaratWebPresenter)
	m := new(interfaces.MockBaccaratGame)
	setupBaccaratWebMockDefaults(m)

	result := parseBaccaratOutput(t, p.Output(m, nil))
	assert.Equal(t, domain.BaccaratPhaseBet, result.Phase)
	assert.Equal(t, 1000, result.Chips)
	assert.Empty(t, result.PlayerHand)
	assert.Empty(t, result.BankerHand)
	assert.Empty(t, result.Message)
	assert.Empty(t, result.History)
	assert.Empty(t, result.SideBetResults)
}

func TestBaccaratWebPresenter_Output_EndPhase_PlayerWins(t *testing.T) {
	p := new(BaccaratWebPresenter)
	m := new(interfaces.MockBaccaratGame)
	m.On("GetChips").Return(1100).Maybe()
	m.On("GetPhase").Return(domain.BaccaratPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 9, false),
		domain.NewCard(domain.CardDesignHeart, 3, false),
	}).Maybe()
	m.On("GetBankerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignClover, 5, false),
		domain.NewCard(domain.CardDesignDiamond, 2, false),
	}).Maybe()
	m.On("GetPlayerHandValue").Return(2).Maybe()
	m.On("GetBankerHandValue").Return(7).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetBetAmount").Return(100).Maybe()
	m.On("GetBetType").Return(domain.BaccaratBetPlayer).Maybe()
	m.On("GetResult").Return(domain.GameResultWin).Maybe()
	m.On("GetPayout").Return(200).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetHistory").Return([]int{domain.BaccaratResultPlayer}).Maybe()
	m.On("GetPlayerPairBet").Return(0).Maybe()
	m.On("GetBankerPairBet").Return(0).Maybe()
	m.On("GetSideBetResults").Return(([]*domain.BacSideBetResult)(nil)).Maybe()

	result := parseBaccaratOutput(t, p.Output(m, nil))
	assert.Equal(t, domain.BaccaratPhaseEnd, result.Phase)
	assert.Equal(t, "Player wins!", result.Message)
	assert.Equal(t, "baccarat.result.playerWins", result.MessageCode)
	assert.Equal(t, 2, len(result.PlayerHand))
	assert.Equal(t, 2, len(result.BankerHand))
	assert.Equal(t, 200, result.Payout)
	assert.Equal(t, []int{domain.BaccaratResultPlayer}, result.History)
}

func TestBaccaratWebPresenter_Output_EndPhase_BankerWins(t *testing.T) {
	p := new(BaccaratWebPresenter)
	m := new(interfaces.MockBaccaratGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.BaccaratPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 3, false),
	}).Maybe()
	m.On("GetBankerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 8, false),
	}).Maybe()
	m.On("GetPlayerHandValue").Return(3).Maybe()
	m.On("GetBankerHandValue").Return(8).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetBetAmount").Return(100).Maybe()
	m.On("GetBetType").Return(domain.BaccaratBetBanker).Maybe()
	m.On("GetResult").Return(domain.GameResultLose).Maybe()
	m.On("GetPayout").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetHistory").Return([]int{domain.BaccaratResultBanker}).Maybe()
	m.On("GetPlayerPairBet").Return(0).Maybe()
	m.On("GetBankerPairBet").Return(0).Maybe()
	m.On("GetSideBetResults").Return(([]*domain.BacSideBetResult)(nil)).Maybe()

	result := parseBaccaratOutput(t, p.Output(m, nil))
	assert.Equal(t, "Banker wins!", result.Message)
	assert.Equal(t, "baccarat.result.bankerWins", result.MessageCode)
}

func TestBaccaratWebPresenter_Output_EndPhase_Tie(t *testing.T) {
	p := new(BaccaratWebPresenter)
	m := new(interfaces.MockBaccaratGame)
	m.On("GetChips").Return(1900).Maybe()
	m.On("GetPhase").Return(domain.BaccaratPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 5, false),
	}).Maybe()
	m.On("GetBankerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 5, false),
	}).Maybe()
	m.On("GetPlayerHandValue").Return(5).Maybe()
	m.On("GetBankerHandValue").Return(5).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetBetAmount").Return(100).Maybe()
	m.On("GetBetType").Return(domain.BaccaratBetTie).Maybe()
	m.On("GetResult").Return(domain.GameResultDraw).Maybe()
	m.On("GetPayout").Return(900).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetHistory").Return(([]int)(nil)).Maybe()
	m.On("GetPlayerPairBet").Return(0).Maybe()
	m.On("GetBankerPairBet").Return(0).Maybe()
	m.On("GetSideBetResults").Return(([]*domain.BacSideBetResult)(nil)).Maybe()

	result := parseBaccaratOutput(t, p.Output(m, nil))
	assert.Equal(t, "Tie!", result.Message)
	assert.Equal(t, "baccarat.result.tie", result.MessageCode)
}

func TestBaccaratWebPresenter_Output_EndPhase_UnknownResult(t *testing.T) {
	p := new(BaccaratWebPresenter)
	m := new(interfaces.MockBaccaratGame)
	setupBaccaratWebMockDefaults(m)
	m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetGameEndFlag")
	m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetResult")
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetResult").Return(domain.GameResult(99)).Maybe()

	result := parseBaccaratOutput(t, p.Output(m, nil))
	assert.Empty(t, result.Message)
}

func TestBaccaratWebPresenter_Output_Error(t *testing.T) {
	p := new(BaccaratWebPresenter)
	m := new(interfaces.MockBaccaratGame)
	setupBaccaratWebMockDefaults(m)

	result := parseBaccaratOutput(t, p.Output(m, domain.NewDomainError(domain.ErrInvalidAmount, "Invalid bet amount.")))
	assert.Equal(t, "Invalid bet amount.", result.Message)
}

func TestBaccaratWebPresenter_Output_WithSideBetResults(t *testing.T) {
	p := new(BaccaratWebPresenter)
	m := new(interfaces.MockBaccaratGame)
	setupBaccaratWebMockDefaults(m)
	m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetSideBetResults")
	m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetPlayerPairBet")
	m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetBankerPairBet")
	m.On("GetPlayerPairBet").Return(10).Maybe()
	m.On("GetBankerPairBet").Return(20).Maybe()
	m.On("GetSideBetResults").Return([]*domain.BacSideBetResult{
		{BetType: domain.BacSideBetPlayerPair, ResultType: domain.BacPairMatch, ResultName: "Pair", BetAmount: 10, Payout: 120},
		{BetType: domain.BacSideBetBankerPair, ResultType: domain.BacPairNone, ResultName: "", BetAmount: 20, Payout: 0},
	}).Maybe()

	result := parseBaccaratOutput(t, p.Output(m, nil))
	assert.Equal(t, 10, result.PlayerPairBet)
	assert.Equal(t, 20, result.BankerPairBet)
	assert.Len(t, result.SideBetResults, 2)
	assert.Equal(t, domain.BacSideBetPlayerPair, result.SideBetResults[0].BetType)
	assert.Equal(t, 120, result.SideBetResults[0].Payout)
	assert.Equal(t, domain.BacSideBetBankerPair, result.SideBetResults[1].BetType)
	assert.Equal(t, 0, result.SideBetResults[1].Payout)
}

func TestBaccaratWebPresenter_Output_WithHistory(t *testing.T) {
	p := new(BaccaratWebPresenter)
	m := new(interfaces.MockBaccaratGame)
	setupBaccaratWebMockDefaults(m)
	m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetHistory")
	m.On("GetHistory").Return([]int{domain.BaccaratResultPlayer, domain.BaccaratResultBanker, domain.BaccaratResultTie}).Maybe()

	result := parseBaccaratOutput(t, p.Output(m, nil))
	assert.Equal(t, []int{0, 1, 2}, result.History)
}

func TestBaccaratWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(BaccaratWebPresenter)

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockBaccaratGame)
		m.On("GetGameEndFlag").Return(false)
		jsonStr := p.ActionLogOutput(m)
		var out controller.ActionLogWebOutput
		err := json.Unmarshal([]byte(jsonStr), &out)
		assert.NoError(t, err)
		assert.Empty(t, out.Entries)
	})

	t.Run("game ended with log", func(t *testing.T) {
		m := new(interfaces.MockBaccaratGame)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "bet", Detail: "bet 100 on player"},
		})
		jsonStr := p.ActionLogOutput(m)
		var out controller.ActionLogWebOutput
		err := json.Unmarshal([]byte(jsonStr), &out)
		assert.NoError(t, err)
		assert.Len(t, out.Entries, 1)
		assert.Equal(t, "bet", out.Entries[0].ActionType)
	})

	t.Run("game ended without log", func(t *testing.T) {
		m := new(interfaces.MockBaccaratGame)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
		jsonStr := p.ActionLogOutput(m)
		var out controller.ActionLogWebOutput
		err := json.Unmarshal([]byte(jsonStr), &out)
		assert.NoError(t, err)
		assert.Empty(t, out.Entries)
	})
}
