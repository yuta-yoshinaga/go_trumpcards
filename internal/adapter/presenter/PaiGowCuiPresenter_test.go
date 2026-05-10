package presenter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupPaiGowCuiMockDefaults(m *interfaces.MockPaiGowGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.PaiGowPhaseBet).Maybe()
	m.On("GetPlayerCards").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerCards").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetPlayerHighHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetPlayerLowHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHighHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerLowHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetHighHandResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetLowHandResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetPayout").Return(0).Maybe()
	m.On("GetCommission").Return(0).Maybe()
	m.On("GetPlayerHighRank").Return(0).Maybe()
	m.On("GetPlayerLowRank").Return(0).Maybe()
	m.On("GetDealerHighRank").Return(0).Maybe()
	m.On("GetDealerLowRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

func TestPaiGowCuiPresenter_Output_BetPhase(t *testing.T) {
	p := new(PaiGowCuiPresenter)
	m := new(interfaces.MockPaiGowGame)
	setupPaiGowCuiMockDefaults(m)

	result := p.Output(m, nil)
	assert.Contains(t, result, "チップ: 1000")
	assert.Contains(t, result, "フェーズ: BET")
}

func TestPaiGowCuiPresenter_Output_SetHandsPhase(t *testing.T) {
	p := new(PaiGowCuiPresenter)
	m := new(interfaces.MockPaiGowGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.PaiGowPhaseSetHands).Maybe()
	m.On("GetPlayerCards").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignClover, 10, false),
		domain.NewCard(domain.CardDesignHeart, 5, false),
		domain.NewCard(domain.CardDesignDiamond, 13, false),
		domain.NewCard(domain.CardDesignSpade, 8, false),
		domain.NewCard(domain.CardDesignClover, 3, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
	}).Maybe()
	m.On("GetDealerCards").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetPlayerHighHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetPlayerLowHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHighHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerLowHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetBet").Return(100).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetHighHandResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetLowHandResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetPayout").Return(0).Maybe()
	m.On("GetCommission").Return(0).Maybe()
	m.On("GetPlayerHighRank").Return(0).Maybe()
	m.On("GetPlayerLowRank").Return(0).Maybe()
	m.On("GetDealerHighRank").Return(0).Maybe()
	m.On("GetDealerLowRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "フェーズ: SET HANDS")
	assert.Contains(t, result, "[0]")
	assert.Contains(t, result, "[6]")
}

func TestPaiGowCuiPresenter_Output_Error(t *testing.T) {
	p := new(PaiGowCuiPresenter)
	m := new(interfaces.MockPaiGowGame)
	setupPaiGowCuiMockDefaults(m)

	result := p.Output(m, errors.New("test error"))
	assert.Contains(t, result, "test error")
}

func TestPaiGowCuiPresenter_Output_PlayerWins(t *testing.T) {
	p := new(PaiGowCuiPresenter)
	m := new(interfaces.MockPaiGowGame)
	m.On("GetChips").Return(1190).Maybe()
	m.On("GetPhase").Return(domain.PaiGowPhaseEnd).Maybe()
	m.On("GetPlayerCards").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerCards").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetPlayerHighHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetPlayerLowHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHighHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerLowHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetBet").Return(100).Maybe()
	m.On("GetResult").Return(domain.GameResultWin).Maybe()
	m.On("GetHighHandResult").Return(domain.GameResultWin).Maybe()
	m.On("GetLowHandResult").Return(domain.GameResultWin).Maybe()
	m.On("GetPayout").Return(190).Maybe()
	m.On("GetCommission").Return(10).Maybe()
	m.On("GetPlayerHighRank").Return(0).Maybe()
	m.On("GetPlayerLowRank").Return(0).Maybe()
	m.On("GetDealerHighRank").Return(0).Maybe()
	m.On("GetDealerLowRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "プレイヤーの勝ち")
	assert.Contains(t, result, "払戻し: 190")
	assert.Contains(t, result, "手数料: 10")
}

func TestPaiGowCuiPresenter_Output_DealerWins(t *testing.T) {
	p := new(PaiGowCuiPresenter)
	m := new(interfaces.MockPaiGowGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.PaiGowPhaseEnd).Maybe()
	m.On("GetPlayerCards").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerCards").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetPlayerHighHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetPlayerLowHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHighHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerLowHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetBet").Return(100).Maybe()
	m.On("GetResult").Return(domain.GameResultLose).Maybe()
	m.On("GetHighHandResult").Return(domain.GameResultLose).Maybe()
	m.On("GetLowHandResult").Return(domain.GameResultLose).Maybe()
	m.On("GetPayout").Return(0).Maybe()
	m.On("GetCommission").Return(0).Maybe()
	m.On("GetPlayerHighRank").Return(0).Maybe()
	m.On("GetPlayerLowRank").Return(0).Maybe()
	m.On("GetDealerHighRank").Return(0).Maybe()
	m.On("GetDealerLowRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "ディーラーの勝ち")
}

func TestPaiGowCuiPresenter_Output_Push(t *testing.T) {
	p := new(PaiGowCuiPresenter)
	m := new(interfaces.MockPaiGowGame)
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.PaiGowPhaseEnd).Maybe()
	m.On("GetPlayerCards").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerCards").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetPlayerHighHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetPlayerLowHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHighHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerLowHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetBet").Return(100).Maybe()
	m.On("GetResult").Return(domain.GameResultDraw).Maybe()
	m.On("GetHighHandResult").Return(domain.GameResultWin).Maybe()
	m.On("GetLowHandResult").Return(domain.GameResultLose).Maybe()
	m.On("GetPayout").Return(100).Maybe()
	m.On("GetCommission").Return(0).Maybe()
	m.On("GetPlayerHighRank").Return(0).Maybe()
	m.On("GetPlayerLowRank").Return(0).Maybe()
	m.On("GetDealerHighRank").Return(0).Maybe()
	m.On("GetDealerLowRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "プッシュ")
	assert.Contains(t, result, "払戻し: 100")
}

func TestPaiGowCuiPresenter_Output_EndPhaseWithHands(t *testing.T) {
	p := new(PaiGowCuiPresenter)
	m := new(interfaces.MockPaiGowGame)
	m.On("GetChips").Return(1195).Maybe()
	m.On("GetPhase").Return(domain.PaiGowPhaseEnd).Maybe()
	m.On("GetPlayerCards").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerCards").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetPlayerHighHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignHeart, 13, false),
		domain.NewCard(domain.CardDesignClover, 12, false),
		domain.NewCard(domain.CardDesignDiamond, 11, false),
		domain.NewCard(domain.CardDesignSpade, 10, false),
	}).Maybe()
	m.On("GetPlayerLowHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 5, false),
		domain.NewCard(domain.CardDesignClover, 3, false),
	}).Maybe()
	m.On("GetDealerHighHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 9, false),
		domain.NewCard(domain.CardDesignSpade, 8, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignClover, 6, false),
		domain.NewCard(domain.CardDesignDiamond, 4, false),
	}).Maybe()
	m.On("GetDealerLowHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 2, false),
		domain.NewCard(domain.CardDesignHeart, 3, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetBet").Return(100).Maybe()
	m.On("GetResult").Return(domain.GameResultWin).Maybe()
	m.On("GetHighHandResult").Return(domain.GameResultWin).Maybe()
	m.On("GetLowHandResult").Return(domain.GameResultWin).Maybe()
	m.On("GetPayout").Return(195).Maybe()
	m.On("GetCommission").Return(5).Maybe()
	m.On("GetPlayerHighRank").Return(domain.PokerHandStraight).Maybe()
	m.On("GetPlayerLowRank").Return(domain.PaiGowLowHandHighCard).Maybe()
	m.On("GetDealerHighRank").Return(domain.PokerHandHighCard).Maybe()
	m.On("GetDealerLowRank").Return(domain.PaiGowLowHandHighCard).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "ハイ:")
	assert.Contains(t, result, "ロー:")
	assert.Contains(t, result, "DEALER")
	assert.Contains(t, result, "Straight")
	assert.Contains(t, result, "High Card")
}

func TestPaiGowCuiPresenter_PhaseStr_Unknown(t *testing.T) {
	p := new(PaiGowCuiPresenter)
	m := new(interfaces.MockPaiGowGame)
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(99).Maybe()
	m.On("GetPlayerCards").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerCards").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetPlayerHighHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetPlayerLowHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHighHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerLowHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetHighHandResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetLowHandResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetPayout").Return(0).Maybe()
	m.On("GetCommission").Return(0).Maybe()
	m.On("GetPlayerHighRank").Return(0).Maybe()
	m.On("GetPlayerLowRank").Return(0).Maybe()
	m.On("GetDealerHighRank").Return(0).Maybe()
	m.On("GetDealerLowRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "UNKNOWN")
}

func TestPaiGowCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(PaiGowCuiPresenter)
	m := new(interfaces.MockPaiGowGame)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()

	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "棋譜")
}
