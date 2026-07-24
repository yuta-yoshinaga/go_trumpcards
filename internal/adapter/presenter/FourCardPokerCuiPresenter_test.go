package presenter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupFourCardPokerCuiMockDefaults(m *interfaces.MockFourCardPokerGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.FourCardPokerPhaseBet).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetPlayerBest").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerBest").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerUpCard").Return((*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteBet").Return(0).Maybe()
	m.On("GetAcesUpBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe()
	m.On("GetPlayMultiplier").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetAnteBonusPayout").Return(0).Maybe()
	m.On("GetAcesUpPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

func TestFourCardPokerCuiPresenter_Output_BetPhase(t *testing.T) {
	p := new(FourCardPokerCuiPresenter)
	m := new(interfaces.MockFourCardPokerGame)
	setupFourCardPokerCuiMockDefaults(m)

	result := p.Output(m, nil)
	assert.Contains(t, result, "チップ: 1000")
	assert.Contains(t, result, "フェーズ: BET")
}

func TestFourCardPokerCuiPresenter_Output_ActionPhase_ShowsUpcard(t *testing.T) {
	p := new(FourCardPokerCuiPresenter)
	m := new(interfaces.MockFourCardPokerGame)
	up := domain.NewCard(domain.CardDesignSpade, 13, false)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.FourCardPokerPhaseAction).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignClover, 10, false),
		domain.NewCard(domain.CardDesignHeart, 5, false),
		domain.NewCard(domain.CardDesignDiamond, 7, false),
		domain.NewCard(domain.CardDesignClover, 2, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{up,
		domain.NewCard(domain.CardDesignSpade, 4, false),
		domain.NewCard(domain.CardDesignSpade, 6, false),
		domain.NewCard(domain.CardDesignSpade, 8, false),
		domain.NewCard(domain.CardDesignSpade, 9, false),
		domain.NewCard(domain.CardDesignSpade, 11, false),
	}).Maybe()
	m.On("GetPlayerBest").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerBest").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerUpCard").Return(up).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetAcesUpBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe()
	m.On("GetPlayMultiplier").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetAnteBonusPayout").Return(0).Maybe()
	m.On("GetAcesUpPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(domain.FourCardHandHighCard).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "フェーズ: ACTION")
	assert.Contains(t, result, "PLAYER")
	assert.Contains(t, result, "DEALER")
	assert.Contains(t, result, "アップカード")
}

func TestFourCardPokerCuiPresenter_Output_EndPhase_PlayerWins(t *testing.T) {
	p := new(FourCardPokerCuiPresenter)
	m := new(interfaces.MockFourCardPokerGame)
	m.On("GetChips").Return(1200).Maybe()
	m.On("GetPhase").Return(domain.FourCardPokerPhaseEnd).Maybe()
	playerHand := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 9, false),
		domain.NewCard(domain.CardDesignClover, 9, false),
		domain.NewCard(domain.CardDesignHeart, 9, false),
		domain.NewCard(domain.CardDesignDiamond, 5, false),
		domain.NewCard(domain.CardDesignClover, 2, false),
	}
	dealerHand := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 2, false),
		domain.NewCard(domain.CardDesignClover, 4, false),
		domain.NewCard(domain.CardDesignHeart, 6, false),
		domain.NewCard(domain.CardDesignDiamond, 7, false),
		domain.NewCard(domain.CardDesignClover, 8, false),
		domain.NewCard(domain.CardDesignSpade, 9, false),
	}
	m.On("GetPlayerHand").Return(playerHand).Maybe()
	m.On("GetDealerHand").Return(dealerHand).Maybe()
	m.On("GetPlayerBest").Return(playerHand[:4]).Maybe()
	m.On("GetDealerBest").Return(dealerHand[:4]).Maybe()
	m.On("GetDealerUpCard").Return(dealerHand[0]).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetAcesUpBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(100).Maybe()
	m.On("GetPlayMultiplier").Return(1).Maybe()
	m.On("GetResult").Return(domain.GameResultWin).Maybe()
	m.On("GetAntePayout").Return(200).Maybe()
	m.On("GetPlayPayout").Return(200).Maybe()
	m.On("GetAnteBonusPayout").Return(200).Maybe()
	m.On("GetAcesUpPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(600).Maybe()
	m.On("GetPlayerHandRank").Return(domain.FourCardHandThreeOfAKind).Maybe()
	m.On("GetDealerHandRank").Return(domain.FourCardHandHighCard).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "フェーズ: END")
	assert.Contains(t, result, "プレイヤーの勝ち")
	assert.Contains(t, result, "合計払戻し: 600")
	// Non-zero buckets are itemized; the zero Aces Up bucket is omitted.
	assert.Contains(t, result, i18n.Tf("fourcardpoker.antePayoutLine", "payout", "200"))
	assert.Contains(t, result, i18n.Tf("fourcardpoker.playPayoutLine", "payout", "200"))
	assert.Contains(t, result, i18n.Tf("fourcardpoker.anteBonusPayoutLine", "payout", "200"))
	acesUpPrefix := strings.SplitN(i18n.T("fourcardpoker.acesUpPayoutLine"), "{{", 2)[0]
	assert.NotContains(t, result, acesUpPrefix)
}

func TestFourCardPokerCuiPresenter_Output_EndPhase_Fold(t *testing.T) {
	p := new(FourCardPokerCuiPresenter)
	m := new(interfaces.MockFourCardPokerGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.FourCardPokerPhaseEnd).Maybe()
	playerHand := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 2, false),
		domain.NewCard(domain.CardDesignClover, 4, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignDiamond, 9, false),
		domain.NewCard(domain.CardDesignClover, 11, false),
	}
	dealerHand := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 13, false),
		domain.NewCard(domain.CardDesignClover, 12, false),
		domain.NewCard(domain.CardDesignHeart, 8, false),
		domain.NewCard(domain.CardDesignDiamond, 5, false),
		domain.NewCard(domain.CardDesignClover, 3, false),
		domain.NewCard(domain.CardDesignSpade, 6, false),
	}
	m.On("GetPlayerHand").Return(playerHand).Maybe()
	m.On("GetDealerHand").Return(dealerHand).Maybe()
	m.On("GetPlayerBest").Return(playerHand[:4]).Maybe()
	m.On("GetDealerBest").Return(dealerHand[:4]).Maybe()
	m.On("GetDealerUpCard").Return(dealerHand[0]).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetAcesUpBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe()
	m.On("GetPlayMultiplier").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResultLose).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetAnteBonusPayout").Return(0).Maybe()
	m.On("GetAcesUpPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(domain.FourCardHandHighCard).Maybe()
	m.On("GetDealerHandRank").Return(domain.FourCardHandHighCard).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "プレイヤーがフォールド")
}

func TestFourCardPokerCuiPresenter_Output_EndPhase_DealerWins(t *testing.T) {
	p := new(FourCardPokerCuiPresenter)
	m := new(interfaces.MockFourCardPokerGame)
	m.On("GetChips").Return(800).Maybe()
	m.On("GetPhase").Return(domain.FourCardPokerPhaseEnd).Maybe()
	playerHand := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 2, false),
		domain.NewCard(domain.CardDesignClover, 4, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignDiamond, 9, false),
		domain.NewCard(domain.CardDesignClover, 11, false),
	}
	dealerHand := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 13, false),
		domain.NewCard(domain.CardDesignClover, 13, false),
		domain.NewCard(domain.CardDesignHeart, 8, false),
		domain.NewCard(domain.CardDesignDiamond, 5, false),
		domain.NewCard(domain.CardDesignClover, 3, false),
		domain.NewCard(domain.CardDesignSpade, 6, false),
	}
	m.On("GetPlayerHand").Return(playerHand).Maybe()
	m.On("GetDealerHand").Return(dealerHand).Maybe()
	m.On("GetPlayerBest").Return(playerHand[:4]).Maybe()
	m.On("GetDealerBest").Return(dealerHand[:4]).Maybe()
	m.On("GetDealerUpCard").Return(dealerHand[0]).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetAcesUpBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(100).Maybe()
	m.On("GetPlayMultiplier").Return(1).Maybe()
	m.On("GetResult").Return(domain.GameResultLose).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetAnteBonusPayout").Return(0).Maybe()
	m.On("GetAcesUpPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(domain.FourCardHandHighCard).Maybe()
	m.On("GetDealerHandRank").Return(domain.FourCardHandPair).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "ディーラーの勝ち")
}

func TestFourCardPokerCuiPresenter_Output_EndPhase_Push(t *testing.T) {
	p := new(FourCardPokerCuiPresenter)
	m := new(interfaces.MockFourCardPokerGame)
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.FourCardPokerPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetPlayerBest").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerBest").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerUpCard").Return((*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetAcesUpBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(100).Maybe()
	m.On("GetPlayMultiplier").Return(1).Maybe()
	m.On("GetResult").Return(domain.GameResultDraw).Maybe()
	m.On("GetAntePayout").Return(100).Maybe()
	m.On("GetPlayPayout").Return(100).Maybe()
	m.On("GetAnteBonusPayout").Return(0).Maybe()
	m.On("GetAcesUpPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(200).Maybe()
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "プッシュ")
}

func TestFourCardPokerCuiPresenter_Output_Error(t *testing.T) {
	p := new(FourCardPokerCuiPresenter)
	m := new(interfaces.MockFourCardPokerGame)
	setupFourCardPokerCuiMockDefaults(m)

	result := p.Output(m, domain.NewDomainError(domain.ErrWrongPhase, "wrong phase"))
	assert.Contains(t, result, "wrong phase")
}

func TestFourCardPokerCuiPresenter_PhaseUnknown(t *testing.T) {
	p := new(FourCardPokerCuiPresenter)
	m := new(interfaces.MockFourCardPokerGame)
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(999).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetPlayerBest").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerBest").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerUpCard").Return((*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteBet").Return(0).Maybe()
	m.On("GetAcesUpBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe()
	m.On("GetPlayMultiplier").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetAnteBonusPayout").Return(0).Maybe()
	m.On("GetAcesUpPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "UNKNOWN")
}

func TestFourCardPokerCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(FourCardPokerCuiPresenter)
	m := new(interfaces.MockFourCardPokerGame)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()

	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "棋譜はありません")
}
