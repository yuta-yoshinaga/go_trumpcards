package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupCasinoHoldemCuiMockDefaults(m *interfaces.MockCasinoHoldemGame) {
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
	m.On("GetPlayerBest").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerBest").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

func TestCasinoHoldemCuiPresenter_Output_BetPhase(t *testing.T) {
	p := new(CasinoHoldemCuiPresenter)
	m := new(interfaces.MockCasinoHoldemGame)
	setupCasinoHoldemCuiMockDefaults(m)

	result := p.Output(m, nil)
	assert.Contains(t, result, "チップ: 1000")
	assert.Contains(t, result, "フェーズ: BET")
}

func TestCasinoHoldemCuiPresenter_Output_FlopPhase(t *testing.T) {
	p := new(CasinoHoldemCuiPresenter)
	m := new(interfaces.MockCasinoHoldemGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.CasinoHoldemPhaseFlop).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignClover, 13, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 5, false),
		domain.NewCard(domain.CardDesignDiamond, 7, false),
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

	result := p.Output(m, nil)
	assert.Contains(t, result, "フェーズ: FLOP")
	assert.Contains(t, result, "BOARD")
	assert.Contains(t, result, "PLAYER")
	assert.Contains(t, result, "DEALER")
	// Dealer cards hidden during flop
	assert.Contains(t, result, "??")
}

func TestCasinoHoldemCuiPresenter_Output_EndPhase_PlayerWinsCall(t *testing.T) {
	p := new(CasinoHoldemCuiPresenter)
	m := new(interfaces.MockCasinoHoldemGame)
	m.On("GetChips").Return(2000).Maybe()
	m.On("GetPhase").Return(domain.CasinoHoldemPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignClover, 1, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 13, false),
		domain.NewCard(domain.CardDesignDiamond, 13, false),
	}).Maybe()
	m.On("GetCommunity").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 2, false),
		domain.NewCard(domain.CardDesignClover, 4, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignSpade, 9, false),
		domain.NewCard(domain.CardDesignDiamond, 12, false),
	}).Maybe()
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
	m.On("GetPlayerHandRank").Return(domain.PokerHandOnePair).Maybe()
	m.On("GetDealerHandRank").Return(domain.PokerHandOnePair).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "フェーズ: END")
	assert.Contains(t, result, "プレイヤーの勝ち")
	assert.Contains(t, result, "合計払戻し: 600")
	assert.Contains(t, result, "ディーラークオリファイ")
}

func TestCasinoHoldemCuiPresenter_Output_EndPhase_Fold(t *testing.T) {
	p := new(CasinoHoldemCuiPresenter)
	m := new(interfaces.MockCasinoHoldemGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.CasinoHoldemPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, false),
		domain.NewCard(domain.CardDesignClover, 2, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 13, false),
		domain.NewCard(domain.CardDesignDiamond, 13, false),
	}).Maybe()
	m.On("GetCommunity").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 5, false),
		domain.NewCard(domain.CardDesignClover, 8, false),
		domain.NewCard(domain.CardDesignHeart, 11, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetBonusBet").Return(0).Maybe()
	m.On("GetCallBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResultLose).Maybe()
	m.On("GetDealerQualify").Return(false).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetCallPayout").Return(0).Maybe()
	m.On("GetBonusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "プレイヤーがフォールド")
}

func TestCasinoHoldemCuiPresenter_Output_EndPhase_Push(t *testing.T) {
	p := new(CasinoHoldemCuiPresenter)
	m := new(interfaces.MockCasinoHoldemGame)
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.CasinoHoldemPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 2, false),
		domain.NewCard(domain.CardDesignSpade, 3, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignClover, 4, false),
		domain.NewCard(domain.CardDesignClover, 5, false),
	}).Maybe()
	m.On("GetCommunity").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 6, false),
		domain.NewCard(domain.CardDesignDiamond, 8, false),
		domain.NewCard(domain.CardDesignSpade, 11, false),
		domain.NewCard(domain.CardDesignDiamond, 12, false),
		domain.NewCard(domain.CardDesignClover, 13, false),
	}).Maybe()
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
	m.On("GetPlayerHandRank").Return(domain.PokerHandHighCard).Maybe()
	m.On("GetDealerHandRank").Return(domain.PokerHandHighCard).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "プッシュ")
}

func TestCasinoHoldemCuiPresenter_Output_EndPhase_DealerWins(t *testing.T) {
	p := new(CasinoHoldemCuiPresenter)
	m := new(interfaces.MockCasinoHoldemGame)
	m.On("GetChips").Return(700).Maybe()
	m.On("GetPhase").Return(domain.CasinoHoldemPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 2, false),
		domain.NewCard(domain.CardDesignClover, 3, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 1, false),
		domain.NewCard(domain.CardDesignDiamond, 1, false),
	}).Maybe()
	m.On("GetCommunity").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 5, false),
		domain.NewCard(domain.CardDesignClover, 7, false),
		domain.NewCard(domain.CardDesignHeart, 9, false),
		domain.NewCard(domain.CardDesignSpade, 11, false),
		domain.NewCard(domain.CardDesignClover, 13, false),
	}).Maybe()
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
	m.On("GetPlayerHandRank").Return(domain.PokerHandHighCard).Maybe()
	m.On("GetDealerHandRank").Return(domain.PokerHandOnePair).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "ディーラーの勝ち")
}

// ノークオリファイ：コール後にディーラー不適格
func TestCasinoHoldemCuiPresenter_Output_EndPhase_NoQualify(t *testing.T) {
	p := new(CasinoHoldemCuiPresenter)
	m := new(interfaces.MockCasinoHoldemGame)
	m.On("GetChips").Return(1300).Maybe()
	m.On("GetPhase").Return(domain.CasinoHoldemPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignSpade, 13, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 2, false),
		domain.NewCard(domain.CardDesignDiamond, 7, false),
	}).Maybe()
	m.On("GetCommunity").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 5, false),
		domain.NewCard(domain.CardDesignClover, 8, false),
		domain.NewCard(domain.CardDesignHeart, 11, false),
		domain.NewCard(domain.CardDesignClover, 3, false),
		domain.NewCard(domain.CardDesignDiamond, 4, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetBonusBet").Return(0).Maybe()
	m.On("GetCallBet").Return(200).Maybe()
	m.On("GetResult").Return(domain.GameResultWin).Maybe()
	m.On("GetDealerQualify").Return(false).Maybe()
	m.On("GetAntePayout").Return(200).Maybe()
	m.On("GetCallPayout").Return(200).Maybe()
	m.On("GetBonusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(400).Maybe()
	m.On("GetPlayerHandRank").Return(domain.PokerHandHighCard).Maybe()
	m.On("GetDealerHandRank").Return(domain.PokerHandHighCard).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "ディーラーノークオリファイ")
}

func TestCasinoHoldemCuiPresenter_Output_Error(t *testing.T) {
	p := new(CasinoHoldemCuiPresenter)
	m := new(interfaces.MockCasinoHoldemGame)
	setupCasinoHoldemCuiMockDefaults(m)

	result := p.Output(m, domain.NewDomainError(domain.ErrWrongPhase, "wrong phase"))
	assert.Contains(t, result, "wrong phase")
}

func TestCasinoHoldemCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(CasinoHoldemCuiPresenter)
	m := new(interfaces.MockCasinoHoldemGame)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()

	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "棋譜はありません")
}

func TestCasinoHoldemCuiPresenter_PhaseStr_Unknown(t *testing.T) {
	p := new(CasinoHoldemCuiPresenter)
	m := new(interfaces.MockCasinoHoldemGame)
	m.On("GetChips").Return(0).Maybe()
	m.On("GetPhase").Return(999).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetCommunity").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "フェーズ: UNKNOWN")
}

func TestCasinoHoldemCuiPresenter_HintOutput(t *testing.T) {
	p := new(CasinoHoldemCuiPresenter)

	t.Run("recommends call when advised", func(t *testing.T) {
		m := new(interfaces.MockCasinoHoldemGame)
		m.On("GetPhase").Return(domain.CasinoHoldemPhaseFlop)
		m.On("RecommendCall").Return(true)
		assert.Contains(t, p.HintOutput(m), i18n.T("casinoholdem.hintCall"))
	})

	t.Run("recommends fold when not advised", func(t *testing.T) {
		m := new(interfaces.MockCasinoHoldemGame)
		m.On("GetPhase").Return(domain.CasinoHoldemPhaseFlop)
		m.On("RecommendCall").Return(false)
		assert.Contains(t, p.HintOutput(m), i18n.T("casinoholdem.hintFold"))
	})

	t.Run("no hint outside the flop phase", func(t *testing.T) {
		m := new(interfaces.MockCasinoHoldemGame)
		m.On("GetPhase").Return(domain.CasinoHoldemPhaseBet)
		assert.Contains(t, p.HintOutput(m), i18n.T("casinoholdem.hintNone"))
	})
}
