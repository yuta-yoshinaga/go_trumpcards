package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupThreeCardCuiMockDefaults(m *interfaces.MockThreeCardGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.ThreeCardPhaseBet).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteBet").Return(0).Maybe()
	m.On("GetPairPlusBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetAnteBonusPayout").Return(0).Maybe()
	m.On("GetPairPlusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

func TestThreeCardCuiPresenter_Output_BetPhase(t *testing.T) {
	p := new(ThreeCardCuiPresenter)
	m := new(interfaces.MockThreeCardGame)
	setupThreeCardCuiMockDefaults(m)

	result := p.Output(m, nil)
	assert.Contains(t, result, "チップ: 1000")
	assert.Contains(t, result, "フェーズ: BET")
}

func TestThreeCardCuiPresenter_HintOutput(t *testing.T) {
	p := new(ThreeCardCuiPresenter)

	actionMock := func(rank int, cards ...*domain.Card) *interfaces.MockThreeCardGame {
		m := new(interfaces.MockThreeCardGame)
		m.On("GetPhase").Return(domain.ThreeCardPhaseAction).Maybe()
		m.On("GetPlayerHand").Return(cards).Maybe()
		m.On("GetPlayerHandRank").Return(rank).Maybe()
		return m
	}
	c := func(v int) *domain.Card { return domain.NewCard(domain.CardDesignSpade, v, false) }
	mixed := func(v int, d int) *domain.Card { return domain.NewCard(d, v, false) }

	t.Run("pair always plays", func(t *testing.T) {
		m := actionMock(domain.ThreeCardHandPair, c(3), c(3), c(9))
		assert.Contains(t, p.HintOutput(m), "プレイ")
	})

	t.Run("Q-6-4 high card plays (boundary)", func(t *testing.T) {
		m := actionMock(domain.ThreeCardHandHighCard, mixed(12, domain.CardDesignSpade), mixed(6, domain.CardDesignHeart), mixed(4, domain.CardDesignClover))
		assert.Contains(t, p.HintOutput(m), "プレイ")
	})

	t.Run("Q-6-3 high card folds (just below boundary)", func(t *testing.T) {
		m := actionMock(domain.ThreeCardHandHighCard, mixed(12, domain.CardDesignSpade), mixed(6, domain.CardDesignHeart), mixed(3, domain.CardDesignClover))
		assert.Contains(t, p.HintOutput(m), "フォールド")
	})

	t.Run("Ace high plays", func(t *testing.T) {
		m := actionMock(domain.ThreeCardHandHighCard, mixed(1, domain.CardDesignSpade), mixed(9, domain.CardDesignHeart), mixed(2, domain.CardDesignClover))
		assert.Contains(t, p.HintOutput(m), "プレイ")
	})

	t.Run("Jack high folds", func(t *testing.T) {
		m := actionMock(domain.ThreeCardHandHighCard, mixed(11, domain.CardDesignSpade), mixed(9, domain.CardDesignHeart), mixed(8, domain.CardDesignClover))
		assert.Contains(t, p.HintOutput(m), "フォールド")
	})

	t.Run("no hint outside the action phase", func(t *testing.T) {
		m := new(interfaces.MockThreeCardGame)
		m.On("GetPhase").Return(domain.ThreeCardPhaseBet).Maybe()
		assert.Contains(t, p.HintOutput(m), "ヒントはありません")
	})

	t.Run("incomplete hand folds (guard)", func(t *testing.T) {
		// An action-phase hand without exactly 3 cards is treated as fold.
		m := actionMock(domain.ThreeCardHandHighCard, c(12), c(6))
		assert.Contains(t, p.HintOutput(m), "フォールド")
	})
}

func TestThreeCardCuiPresenter_Output_ActionPhase(t *testing.T) {
	p := new(ThreeCardCuiPresenter)
	m := new(interfaces.MockThreeCardGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.ThreeCardPhaseAction).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignClover, 10, false),
		domain.NewCard(domain.CardDesignHeart, 5, false),
	}).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetPairPlusBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetAnteBonusPayout").Return(0).Maybe()
	m.On("GetPairPlusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetPlayerHandRank").Return(domain.ThreeCardHandHighCard).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "フェーズ: ACTION")
	assert.Contains(t, result, "PLAYER")
	// 役名は日本語ロケールで日本語。以前は英語の表示名配列をそのまま
	// 埋めていて、このテストがその挙動を固定していた (#4694)。
	assert.Contains(t, result, "ハイカード")
	assert.NotContains(t, result, "High Card")
}

func TestThreeCardCuiPresenter_Output_EndPhase_PlayerWins(t *testing.T) {
	p := new(ThreeCardCuiPresenter)
	m := new(interfaces.MockThreeCardGame)
	m.On("GetChips").Return(1200).Maybe()
	m.On("GetPhase").Return(domain.ThreeCardPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignClover, 10, false),
		domain.NewCard(domain.CardDesignHeart, 5, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 12, false),
		domain.NewCard(domain.CardDesignHeart, 3, false),
		domain.NewCard(domain.CardDesignClover, 2, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetPairPlusBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(100).Maybe()
	m.On("GetResult").Return(domain.GameResultWin).Maybe()
	m.On("GetAntePayout").Return(200).Maybe()
	m.On("GetPlayPayout").Return(200).Maybe()
	m.On("GetAnteBonusPayout").Return(0).Maybe()
	m.On("GetPairPlusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(400).Maybe()
	m.On("GetDealerQualified").Return(true).Maybe()
	m.On("GetPlayerHandRank").Return(domain.ThreeCardHandHighCard).Maybe()
	m.On("GetDealerHandRank").Return(domain.ThreeCardHandHighCard).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "フェーズ: END")
	assert.Contains(t, result, "プレイヤーの勝ち")
	assert.Contains(t, result, "DEALER")
	assert.Contains(t, result, "(Qualified)")
	assert.Contains(t, result, "合計払戻し: 400")
	// Zero side-bet/bonus payouts omit their breakdown lines (backward compatible).
	assert.NotContains(t, result, "ペアプラス配当")
	assert.NotContains(t, result, "アンテボーナス配当")
}

func TestThreeCardCuiPresenter_Output_EndPhase_PayoutBreakdown(t *testing.T) {
	p := new(ThreeCardCuiPresenter)
	m := new(interfaces.MockThreeCardGame)
	m.On("GetChips").Return(1500).Maybe()
	m.On("GetPhase").Return(domain.ThreeCardPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 10, false),
		domain.NewCard(domain.CardDesignSpade, 11, false),
		domain.NewCard(domain.CardDesignSpade, 12, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 5, false),
		domain.NewCard(domain.CardDesignHeart, 3, false),
		domain.NewCard(domain.CardDesignClover, 2, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetPairPlusBet").Return(100).Maybe()
	m.On("GetPlayBet").Return(100).Maybe()
	m.On("GetResult").Return(domain.GameResultWin).Maybe()
	m.On("GetAntePayout").Return(200).Maybe()
	m.On("GetPlayPayout").Return(200).Maybe()
	m.On("GetAnteBonusPayout").Return(500).Maybe()
	m.On("GetPairPlusPayout").Return(4000).Maybe()
	m.On("GetTotalPayout").Return(4900).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetPlayerHandRank").Return(domain.ThreeCardHandStraightFlush).Maybe()
	m.On("GetDealerHandRank").Return(domain.ThreeCardHandHighCard).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	// **役名も日本語で出す。**`domain.ThreeCardHandNames` は英語の表示名配列で、
	// そのまま埋めていたので日本語ロケールでも Straight Flush と出ていた (#4694)。
	assert.Contains(t, result, "ストレートフラッシュ")
	assert.Contains(t, result, "ハイカード")
	assert.NotContains(t, result, "Straight Flush")
	assert.NotContains(t, result, "High Card")
	assert.Contains(t, result, "アンテボーナス配当: 500")
	assert.Contains(t, result, "ペアプラス配当: 4000")
	assert.Contains(t, result, "合計払戻し: 4900")
}

func TestThreeCardCuiPresenter_Output_EndPhase_Fold(t *testing.T) {
	p := new(ThreeCardCuiPresenter)
	m := new(interfaces.MockThreeCardGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.ThreeCardPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 2, false),
		domain.NewCard(domain.CardDesignClover, 4, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 12, false),
		domain.NewCard(domain.CardDesignHeart, 3, false),
		domain.NewCard(domain.CardDesignClover, 2, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetPairPlusBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe() // fold
	m.On("GetResult").Return(domain.GameResultLose).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetAnteBonusPayout").Return(0).Maybe()
	m.On("GetPairPlusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetPlayerHandRank").Return(domain.ThreeCardHandHighCard).Maybe()
	m.On("GetDealerHandRank").Return(domain.ThreeCardHandHighCard).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "プレイヤーがフォールド")
}

func TestThreeCardCuiPresenter_Output_Error(t *testing.T) {
	p := new(ThreeCardCuiPresenter)
	m := new(interfaces.MockThreeCardGame)
	setupThreeCardCuiMockDefaults(m)

	result := p.Output(m, domain.NewDomainError(domain.ErrWrongPhase, "wrong phase"))
	assert.Contains(t, result, "wrong phase")
}

func TestThreeCardCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(ThreeCardCuiPresenter)
	m := new(interfaces.MockThreeCardGame)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()

	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "棋譜はありません")
}
