package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupRussianPokerCuiMockDefaults(m *interfaces.MockRussianPokerGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.RussianPokerPhaseBet).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteBet").Return(0).Maybe()
	m.On("GetExchangeCount").Return(0).Maybe()
	m.On("GetExchangeFee").Return(0).Maybe()
	m.On("GetBought6th").Return(false).Maybe()
	m.On("GetBuy6thFee").Return(0).Maybe()
	m.On("GetForceExchanged").Return(false).Maybe()
	m.On("GetForceExchangeFee").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

func TestRussianPokerCuiPresenter_Output_BetPhase(t *testing.T) {
	p := new(RussianPokerCuiPresenter)
	m := new(interfaces.MockRussianPokerGame)
	setupRussianPokerCuiMockDefaults(m)

	result := p.Output(m, nil)
	assert.Contains(t, result, "チップ: 1000")
	assert.Contains(t, result, "フェーズ: BET")
}

func TestRussianPokerCuiPresenter_Output_ActionPhase(t *testing.T) {
	p := new(RussianPokerCuiPresenter)
	m := new(interfaces.MockRussianPokerGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.RussianPokerPhaseAction).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignClover, 10, false),
		domain.NewCard(domain.CardDesignHeart, 5, false),
		domain.NewCard(domain.CardDesignDiamond, 7, false),
		domain.NewCard(domain.CardDesignSpade, 2, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 13, false),
		domain.NewCard(domain.CardDesignSpade, 5, false),
		domain.NewCard(domain.CardDesignClover, 3, false),
		domain.NewCard(domain.CardDesignDiamond, 8, false),
		domain.NewCard(domain.CardDesignHeart, 2, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetExchangeCount").Return(0).Maybe()
	m.On("GetExchangeFee").Return(0).Maybe()
	m.On("GetBought6th").Return(false).Maybe()
	m.On("GetBuy6thFee").Return(0).Maybe()
	m.On("GetForceExchanged").Return(false).Maybe()
	m.On("GetForceExchangeFee").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetPlayerHandRank").Return(domain.PokerHandHighCard).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "フェーズ: ACTION")
	assert.Contains(t, result, "PLAYER")
	assert.Contains(t, result, "DEALER")
	assert.Contains(t, result, "??")
}

func TestRussianPokerCuiPresenter_Output_PostActionPhase_ShowsExchangeInfo(t *testing.T) {
	p := new(RussianPokerCuiPresenter)
	m := new(interfaces.MockRussianPokerGame)
	m.On("GetChips").Return(700).Maybe()
	m.On("GetPhase").Return(domain.RussianPokerPhasePostAction).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignClover, 10, false),
		domain.NewCard(domain.CardDesignHeart, 5, false),
		domain.NewCard(domain.CardDesignDiamond, 7, false),
		domain.NewCard(domain.CardDesignSpade, 2, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 13, false),
		domain.NewCard(domain.CardDesignSpade, 5, false),
		domain.NewCard(domain.CardDesignClover, 3, false),
		domain.NewCard(domain.CardDesignDiamond, 8, false),
		domain.NewCard(domain.CardDesignHeart, 2, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetExchangeCount").Return(2).Maybe()
	m.On("GetExchangeFee").Return(200).Maybe()
	m.On("GetBought6th").Return(false).Maybe()
	m.On("GetBuy6thFee").Return(0).Maybe()
	m.On("GetForceExchanged").Return(false).Maybe()
	m.On("GetForceExchangeFee").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetPlayerHandRank").Return(domain.PokerHandHighCard).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "フェーズ: POST-ACTION")
	// First dealer card visible
	assert.Contains(t, result, "HEART 13")
	// Remaining masked
	assert.Contains(t, result, "??")
}

func TestRussianPokerCuiPresenter_Output_ForceQualifyPhase(t *testing.T) {
	p := new(RussianPokerCuiPresenter)
	m := new(interfaces.MockRussianPokerGame)
	m.On("GetChips").Return(800).Maybe()
	m.On("GetPhase").Return(domain.RussianPokerPhaseForceQualify).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetExchangeCount").Return(0).Maybe()
	m.On("GetExchangeFee").Return(0).Maybe()
	m.On("GetBought6th").Return(false).Maybe()
	m.On("GetBuy6thFee").Return(0).Maybe()
	m.On("GetForceExchanged").Return(false).Maybe()
	m.On("GetForceExchangeFee").Return(0).Maybe()
	m.On("GetPlayBet").Return(200).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "フェーズ: FORCE QUALIFY")
	assert.Contains(t, result, "ディーラー未クオリファイ")
}

func TestRussianPokerCuiPresenter_Output_EndPhase_PlayerWins(t *testing.T) {
	p := new(RussianPokerCuiPresenter)
	m := new(interfaces.MockRussianPokerGame)
	m.On("GetChips").Return(1400).Maybe()
	m.On("GetPhase").Return(domain.RussianPokerPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, false),
		domain.NewCard(domain.CardDesignClover, 7, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignDiamond, 4, false),
		domain.NewCard(domain.CardDesignSpade, 2, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 4, false),
		domain.NewCard(domain.CardDesignHeart, 4, false),
		domain.NewCard(domain.CardDesignClover, 6, false),
		domain.NewCard(domain.CardDesignSpade, 8, false),
		domain.NewCard(domain.CardDesignDiamond, 11, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetExchangeCount").Return(0).Maybe()
	m.On("GetExchangeFee").Return(0).Maybe()
	m.On("GetBought6th").Return(false).Maybe()
	m.On("GetBuy6thFee").Return(0).Maybe()
	m.On("GetForceExchanged").Return(false).Maybe()
	m.On("GetForceExchangeFee").Return(0).Maybe()
	m.On("GetPlayBet").Return(200).Maybe()
	m.On("GetResult").Return(domain.GameResultWin).Maybe()
	m.On("GetAntePayout").Return(200).Maybe()
	m.On("GetPlayPayout").Return(800).Maybe()
	m.On("GetTotalPayout").Return(1000).Maybe()
	m.On("GetDealerQualified").Return(true).Maybe()
	m.On("GetPlayerHandRank").Return(domain.PokerHandThreeOfAKind).Maybe()
	m.On("GetDealerHandRank").Return(domain.PokerHandOnePair).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "フェーズ: END")
	assert.Contains(t, result, "プレイヤーの勝ち")
	assert.Contains(t, result, "(Qualified)")
	assert.Contains(t, result, "アンテ払戻し: 200")
	assert.Contains(t, result, "プレイ払戻し: 800")
	assert.Contains(t, result, "合計払戻し: 1000")
}

func TestRussianPokerCuiPresenter_Output_SelectPhase_ShowsGuide(t *testing.T) {
	p := new(RussianPokerCuiPresenter)
	m := new(interfaces.MockRussianPokerGame)
	setupRussianPokerCuiMockDefaults(m)
	m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetPhase")
	m.On("GetPhase").Return(domain.RussianPokerPhaseSelect).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "捨てるカードの番号を選んでください")
}

func TestRussianPokerCuiPresenter_Output_EndPhase_ZeroBreakdownOmitted(t *testing.T) {
	p := new(RussianPokerCuiPresenter)
	m := new(interfaces.MockRussianPokerGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.RussianPokerPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetExchangeCount").Return(0).Maybe()
	m.On("GetExchangeFee").Return(0).Maybe()
	m.On("GetBought6th").Return(false).Maybe()
	m.On("GetBuy6thFee").Return(0).Maybe()
	m.On("GetForceExchanged").Return(false).Maybe()
	m.On("GetForceExchangeFee").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResultLose).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetPlayerHandRank").Return(domain.PokerHandHighCard).Maybe()
	m.On("GetDealerHandRank").Return(domain.PokerHandHighCard).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	// Zero ante/play payouts omit their breakdown lines.
	assert.NotContains(t, result, "アンテ払戻し")
	assert.NotContains(t, result, "プレイ払戻し")
	assert.Contains(t, result, "合計払戻し: 0")
}

func TestRussianPokerCuiPresenter_Output_EndPhase_Fold(t *testing.T) {
	p := new(RussianPokerCuiPresenter)
	m := new(interfaces.MockRussianPokerGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.RussianPokerPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 2, false),
		domain.NewCard(domain.CardDesignClover, 4, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignDiamond, 9, false),
		domain.NewCard(domain.CardDesignSpade, 10, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 12, false),
		domain.NewCard(domain.CardDesignHeart, 3, false),
		domain.NewCard(domain.CardDesignClover, 2, false),
		domain.NewCard(domain.CardDesignSpade, 5, false),
		domain.NewCard(domain.CardDesignDiamond, 7, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetExchangeCount").Return(0).Maybe()
	m.On("GetExchangeFee").Return(0).Maybe()
	m.On("GetBought6th").Return(false).Maybe()
	m.On("GetBuy6thFee").Return(0).Maybe()
	m.On("GetForceExchanged").Return(false).Maybe()
	m.On("GetForceExchangeFee").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResultLose).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetPlayerHandRank").Return(domain.PokerHandHighCard).Maybe()
	m.On("GetDealerHandRank").Return(domain.PokerHandHighCard).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "プレイヤーがフォールド")
}

func TestRussianPokerCuiPresenter_Output_EndPhase_ForceExchanged(t *testing.T) {
	p := new(RussianPokerCuiPresenter)
	m := new(interfaces.MockRussianPokerGame)
	m.On("GetChips").Return(1200).Maybe()
	m.On("GetPhase").Return(domain.RussianPokerPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, false),
		domain.NewCard(domain.CardDesignClover, 7, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignDiamond, 4, false),
		domain.NewCard(domain.CardDesignSpade, 2, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 4, false),
		domain.NewCard(domain.CardDesignHeart, 4, false),
		domain.NewCard(domain.CardDesignClover, 6, false),
		domain.NewCard(domain.CardDesignSpade, 8, false),
		domain.NewCard(domain.CardDesignDiamond, 11, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetExchangeCount").Return(0).Maybe()
	m.On("GetExchangeFee").Return(0).Maybe()
	m.On("GetBought6th").Return(false).Maybe()
	m.On("GetBuy6thFee").Return(0).Maybe()
	m.On("GetForceExchanged").Return(true).Maybe()
	m.On("GetForceExchangeFee").Return(100).Maybe()
	m.On("GetPlayBet").Return(200).Maybe()
	m.On("GetResult").Return(domain.GameResultWin).Maybe()
	m.On("GetAntePayout").Return(200).Maybe()
	m.On("GetPlayPayout").Return(800).Maybe()
	m.On("GetTotalPayout").Return(900).Maybe()
	m.On("GetDealerQualified").Return(true).Maybe()
	m.On("GetPlayerHandRank").Return(domain.PokerHandThreeOfAKind).Maybe()
	m.On("GetDealerHandRank").Return(domain.PokerHandOnePair).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "強制交換 (手数料: 100)")
}

func TestRussianPokerCuiPresenter_Output_Buy6th(t *testing.T) {
	p := new(RussianPokerCuiPresenter)
	m := new(interfaces.MockRussianPokerGame)
	m.On("GetChips").Return(800).Maybe()
	m.On("GetPhase").Return(domain.RussianPokerPhaseSelect).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, false),
		domain.NewCard(domain.CardDesignClover, 7, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignDiamond, 4, false),
		domain.NewCard(domain.CardDesignSpade, 2, false),
		domain.NewCard(domain.CardDesignHeart, 9, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 13, false),
		domain.NewCard(domain.CardDesignSpade, 5, false),
		domain.NewCard(domain.CardDesignClover, 3, false),
		domain.NewCard(domain.CardDesignDiamond, 8, false),
		domain.NewCard(domain.CardDesignHeart, 2, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetExchangeCount").Return(0).Maybe()
	m.On("GetExchangeFee").Return(0).Maybe()
	m.On("GetBought6th").Return(true).Maybe()
	m.On("GetBuy6thFee").Return(100).Maybe()
	m.On("GetForceExchanged").Return(false).Maybe()
	m.On("GetForceExchangeFee").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "フェーズ: SELECT")
	assert.Contains(t, result, "6枚目購入 (手数料: 100)")
}

func TestRussianPokerCuiPresenter_Output_Error(t *testing.T) {
	p := new(RussianPokerCuiPresenter)
	m := new(interfaces.MockRussianPokerGame)
	setupRussianPokerCuiMockDefaults(m)

	result := p.Output(m, domain.NewDomainError(domain.ErrWrongPhase, "wrong phase"))
	assert.Contains(t, result, "wrong phase")
}

func TestRussianPokerCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(RussianPokerCuiPresenter)
	m := new(interfaces.MockRussianPokerGame)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()

	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "棋譜はありません")
}
