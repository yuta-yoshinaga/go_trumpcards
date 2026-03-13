package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupBaccaratCuiMockDefaults(m *interfaces.MockBaccaratGame) {
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
}

func TestBaccaratCuiPresenter_Output_BetPhase(t *testing.T) {
	color.SetNoColor(true)
	defer color.SetNoColor(false)
	p := new(BaccaratCuiPresenter)
	m := new(interfaces.MockBaccaratGame)
	setupBaccaratCuiMockDefaults(m)

	result := p.Output(m, nil)
	assert.Contains(t, result, "chips: 1000")
	assert.Contains(t, result, "phase: BET")
}

func TestBaccaratCuiPresenter_Output_EndPhase_PlayerWins(t *testing.T) {
	color.SetNoColor(true)
	defer color.SetNoColor(false)
	p := new(BaccaratCuiPresenter)
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

	result := p.Output(m, nil)
	assert.Contains(t, result, "phase: END")
	assert.Contains(t, result, "PLAYER")
	assert.Contains(t, result, "BANKER")
	assert.Contains(t, result, "Player wins!")
	assert.Contains(t, result, "payout: 200")
	assert.Contains(t, result, "SPADE 9")
}

func TestBaccaratCuiPresenter_Output_EndPhase_BankerWins(t *testing.T) {
	color.SetNoColor(true)
	defer color.SetNoColor(false)
	p := new(BaccaratCuiPresenter)
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

	result := p.Output(m, nil)
	assert.Contains(t, result, "Banker wins!")
	assert.Contains(t, result, "BANKER")
}

func TestBaccaratCuiPresenter_Output_EndPhase_Tie(t *testing.T) {
	color.SetNoColor(true)
	defer color.SetNoColor(false)
	p := new(BaccaratCuiPresenter)
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

	result := p.Output(m, nil)
	assert.Contains(t, result, "Tie!")
	assert.Contains(t, result, "TIE")
}

func TestBaccaratCuiPresenter_Output_Error(t *testing.T) {
	color.SetNoColor(true)
	defer color.SetNoColor(false)
	p := new(BaccaratCuiPresenter)
	m := new(interfaces.MockBaccaratGame)
	setupBaccaratCuiMockDefaults(m)

	result := p.Output(m, domain.NewDomainError(domain.ErrInvalidAmount, "Invalid bet amount."))
	assert.Contains(t, result, "Invalid bet amount.")
}

func TestBaccaratCuiPresenter_Output_UnknownPhase(t *testing.T) {
	color.SetNoColor(true)
	defer color.SetNoColor(false)
	p := new(BaccaratCuiPresenter)
	m := new(interfaces.MockBaccaratGame)
	setupBaccaratCuiMockDefaults(m)
	m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetPhase")
	m.On("GetPhase").Return(99).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "UNKNOWN")
}

func TestBaccaratCuiPresenter_Output_EndPhase_UnknownResult(t *testing.T) {
	color.SetNoColor(true)
	defer color.SetNoColor(false)
	p := new(BaccaratCuiPresenter)
	m := new(interfaces.MockBaccaratGame)
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.BaccaratPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetBankerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetPlayerHandValue").Return(0).Maybe()
	m.On("GetBankerHandValue").Return(0).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetBetAmount").Return(100).Maybe()
	m.On("GetBetType").Return(domain.BaccaratBetPlayer).Maybe()
	m.On("GetResult").Return(domain.GameResult(99)).Maybe()
	m.On("GetPayout").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "payout: 0")
}

func TestBaccaratCuiPresenter_Output_UnknownBetType(t *testing.T) {
	color.SetNoColor(true)
	defer color.SetNoColor(false)
	p := new(BaccaratCuiPresenter)
	m := new(interfaces.MockBaccaratGame)
	m.On("GetChips").Return(900).Maybe()
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
	m.On("GetBetType").Return(99).Maybe()
	m.On("GetResult").Return(domain.GameResultWin).Maybe()
	m.On("GetPayout").Return(200).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "UNKNOWN")
}

func TestBaccaratCuiPresenter_ActionLogOutput(t *testing.T) {
	color.SetNoColor(true)
	defer color.SetNoColor(false)
	p := new(BaccaratCuiPresenter)

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockBaccaratGame)
		m.On("GetGameEndFlag").Return(false)
		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜はありません")
	})

	t.Run("game ended with log", func(t *testing.T) {
		m := new(interfaces.MockBaccaratGame)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "bet", Detail: "bet 100 on player"},
		})
		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "bet 100 on player")
	})

	t.Run("game ended without log", func(t *testing.T) {
		m := new(interfaces.MockBaccaratGame)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜はありません")
	})
}
