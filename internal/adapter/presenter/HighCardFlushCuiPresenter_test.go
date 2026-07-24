package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupHighCardFlushCuiMockDefaults(m *interfaces.MockHighCardFlushGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.HighCardFlushPhaseBet).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteBet").Return(0).Maybe()
	m.On("GetFlushBonusBet").Return(0).Maybe()
	m.On("GetStraightFlushBet").Return(0).Maybe()
	m.On("GetRaiseBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetRaisePayout").Return(0).Maybe()
	m.On("GetFlushBonusPayout").Return(0).Maybe()
	m.On("GetStraightFlushPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetPlayerFlushLen").Return(0).Maybe()
	m.On("GetDealerFlushLen").Return(0).Maybe()
	m.On("GetPlayerStraightFlushLen").Return(0).Maybe()
	m.On("MaxRaiseMultiplier").Return(1).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

func TestHighCardFlushCuiPresenter_Output_BetPhase(t *testing.T) {
	p := new(HighCardFlushCuiPresenter)
	m := new(interfaces.MockHighCardFlushGame)
	setupHighCardFlushCuiMockDefaults(m)

	result := p.Output(m, nil)
	assert.Contains(t, result, "チップ: 1000")
	assert.Contains(t, result, "フェーズ: BET")
}

func TestHighCardFlushCuiPresenter_Output_ActionPhase(t *testing.T) {
	p := new(HighCardFlushCuiPresenter)
	m := new(interfaces.MockHighCardFlushGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.HighCardFlushPhaseAction).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignSpade, 5, false),
		domain.NewCard(domain.CardDesignSpade, 9, false),
		domain.NewCard(domain.CardDesignClover, 10, false),
		domain.NewCard(domain.CardDesignHeart, 5, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignDiamond, 3, false),
	}).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetFlushBonusBet").Return(0).Maybe()
	m.On("GetStraightFlushBet").Return(0).Maybe()
	m.On("GetRaiseBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetRaisePayout").Return(0).Maybe()
	m.On("GetFlushBonusPayout").Return(0).Maybe()
	m.On("GetStraightFlushPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetPlayerFlushLen").Return(3).Maybe()
	m.On("GetDealerFlushLen").Return(0).Maybe()
	m.On("GetPlayerStraightFlushLen").Return(0).Maybe()
	m.On("MaxRaiseMultiplier").Return(1).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "フェーズ: ACTION")
	assert.Contains(t, result, "PLAYER")
	assert.Contains(t, result, "最長フラッシュ: 3")
}

func TestHighCardFlushCuiPresenter_Output_EndPhase_PlayerWins(t *testing.T) {
	p := new(HighCardFlushCuiPresenter)
	m := new(interfaces.MockHighCardFlushGame)
	m.On("GetChips").Return(1200).Maybe()
	m.On("GetPhase").Return(domain.HighCardFlushPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignSpade, 5, false),
		domain.NewCard(domain.CardDesignSpade, 9, false),
		domain.NewCard(domain.CardDesignSpade, 12, false),
		domain.NewCard(domain.CardDesignHeart, 5, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignDiamond, 3, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 12, false),
		domain.NewCard(domain.CardDesignDiamond, 3, false),
		domain.NewCard(domain.CardDesignDiamond, 1, false),
		domain.NewCard(domain.CardDesignHeart, 8, false),
		domain.NewCard(domain.CardDesignSpade, 6, false),
		domain.NewCard(domain.CardDesignClover, 2, false),
		domain.NewCard(domain.CardDesignClover, 4, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetFlushBonusBet").Return(0).Maybe()
	m.On("GetStraightFlushBet").Return(0).Maybe()
	m.On("GetRaiseBet").Return(100).Maybe()
	m.On("GetResult").Return(domain.GameResultWin).Maybe()
	m.On("GetAntePayout").Return(200).Maybe()
	m.On("GetRaisePayout").Return(200).Maybe()
	m.On("GetFlushBonusPayout").Return(0).Maybe()
	m.On("GetStraightFlushPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(400).Maybe()
	m.On("GetDealerQualified").Return(true).Maybe()
	m.On("GetPlayerFlushLen").Return(4).Maybe()
	m.On("GetDealerFlushLen").Return(3).Maybe()
	m.On("GetPlayerStraightFlushLen").Return(0).Maybe()
	m.On("MaxRaiseMultiplier").Return(1).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "フェーズ: END")
	assert.Contains(t, result, "プレイヤーの勝ち")
	assert.Contains(t, result, "DEALER")
	assert.Contains(t, result, "(Qualified)")
	assert.Contains(t, result, "合計払戻し: 400")
}

func TestHighCardFlushCuiPresenter_Output_EndPhase_Fold(t *testing.T) {
	p := new(HighCardFlushCuiPresenter)
	m := new(interfaces.MockHighCardFlushGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.HighCardFlushPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 2, false),
		domain.NewCard(domain.CardDesignClover, 4, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignDiamond, 12, false),
		domain.NewCard(domain.CardDesignHeart, 3, false),
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignClover, 10, false),
	}).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetFlushBonusBet").Return(0).Maybe()
	m.On("GetStraightFlushBet").Return(0).Maybe()
	m.On("GetRaiseBet").Return(0).Maybe() // fold
	m.On("GetResult").Return(domain.GameResultLose).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetRaisePayout").Return(0).Maybe()
	m.On("GetFlushBonusPayout").Return(0).Maybe()
	m.On("GetStraightFlushPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetPlayerFlushLen").Return(2).Maybe()
	m.On("GetDealerFlushLen").Return(0).Maybe()
	m.On("GetPlayerStraightFlushLen").Return(0).Maybe()
	m.On("MaxRaiseMultiplier").Return(1).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "プレイヤーがフォールド")
}

func TestHighCardFlushCuiPresenter_Output_EndPhase_Push(t *testing.T) {
	p := new(HighCardFlushCuiPresenter)
	m := new(interfaces.MockHighCardFlushGame)
	setupHighCardFlushCuiMockDefaults(m)
	// Override to push state
	m2 := new(interfaces.MockHighCardFlushGame)
	m2.On("GetChips").Return(1000).Maybe()
	m2.On("GetPhase").Return(domain.HighCardFlushPhaseEnd).Maybe()
	m2.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
	}).Maybe()
	m2.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 1, false),
	}).Maybe()
	m2.On("GetGameEndFlag").Return(true).Maybe()
	m2.On("GetAnteBet").Return(100).Maybe()
	m2.On("GetFlushBonusBet").Return(0).Maybe()
	m2.On("GetStraightFlushBet").Return(0).Maybe()
	m2.On("GetRaiseBet").Return(100).Maybe()
	m2.On("GetResult").Return(domain.GameResultDraw).Maybe()
	m2.On("GetAntePayout").Return(100).Maybe()
	m2.On("GetRaisePayout").Return(100).Maybe()
	m2.On("GetFlushBonusPayout").Return(0).Maybe()
	m2.On("GetStraightFlushPayout").Return(0).Maybe()
	m2.On("GetTotalPayout").Return(200).Maybe()
	m2.On("GetDealerQualified").Return(true).Maybe()
	m2.On("GetPlayerFlushLen").Return(3).Maybe()
	m2.On("GetDealerFlushLen").Return(3).Maybe()
	m2.On("GetPlayerStraightFlushLen").Return(0).Maybe()
	m2.On("MaxRaiseMultiplier").Return(1).Maybe()
	m2.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m2, nil)
	assert.Contains(t, result, "プッシュ")
	_ = m
}

func TestHighCardFlushCuiPresenter_Output_DealerWins(t *testing.T) {
	p := new(HighCardFlushCuiPresenter)
	m := new(interfaces.MockHighCardFlushGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.HighCardFlushPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 2, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 1, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetFlushBonusBet").Return(0).Maybe()
	m.On("GetStraightFlushBet").Return(0).Maybe()
	m.On("GetRaiseBet").Return(100).Maybe()
	m.On("GetResult").Return(domain.GameResultLose).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetRaisePayout").Return(0).Maybe()
	m.On("GetFlushBonusPayout").Return(0).Maybe()
	m.On("GetStraightFlushPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetDealerQualified").Return(true).Maybe()
	m.On("GetPlayerFlushLen").Return(2).Maybe()
	m.On("GetDealerFlushLen").Return(4).Maybe()
	m.On("GetPlayerStraightFlushLen").Return(0).Maybe()
	m.On("MaxRaiseMultiplier").Return(1).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	result := p.Output(m, nil)
	assert.Contains(t, result, "ディーラーの勝ち")
}

func TestHighCardFlushCuiPresenter_Output_Error(t *testing.T) {
	p := new(HighCardFlushCuiPresenter)
	m := new(interfaces.MockHighCardFlushGame)
	setupHighCardFlushCuiMockDefaults(m)

	result := p.Output(m, domain.NewDomainError(domain.ErrWrongPhase, "wrong phase"))
	assert.Contains(t, result, "wrong phase")
}

func TestHighCardFlushCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(HighCardFlushCuiPresenter)
	m := new(interfaces.MockHighCardFlushGame)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()

	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "棋譜はありません")
}

func TestHighCardFlushCuiPresenter_HintOutput(t *testing.T) {
	p := new(HighCardFlushCuiPresenter)

	t.Run("no hint outside the action phase", func(t *testing.T) {
		m := new(interfaces.MockHighCardFlushGame)
		m.On("GetPhase").Return(domain.HighCardFlushPhaseBet)
		assert.Contains(t, p.HintOutput(m), i18n.T("highcardflush.hintNone"))
	})

	t.Run("raise when the flush qualifies", func(t *testing.T) {
		m := new(interfaces.MockHighCardFlushGame)
		m.On("GetPhase").Return(domain.HighCardFlushPhaseAction)
		m.On("GetPlayerFlushLen").Return(domain.HighCardFlushDealerMinFlushLen)
		assert.Contains(t, p.HintOutput(m), i18n.T("highcardflush.hintRaise"))
	})

	t.Run("fold when the flush is too short", func(t *testing.T) {
		m := new(interfaces.MockHighCardFlushGame)
		m.On("GetPhase").Return(domain.HighCardFlushPhaseAction)
		m.On("GetPlayerFlushLen").Return(domain.HighCardFlushDealerMinFlushLen - 1)
		assert.Contains(t, p.HintOutput(m), i18n.T("highcardflush.hintFold"))
	})
}
