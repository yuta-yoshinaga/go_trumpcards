package presenter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupOasisPokerCuiMockDefaults(m *interfaces.MockOasisPokerGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.OasisPokerPhaseBet).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteBet").Return(0).Maybe()
	m.On("GetJackpotBet").Return(0).Maybe()
	m.On("GetExchangeCount").Return(0).Maybe()
	m.On("GetExchangeFee").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetJackpotPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

func TestOasisPokerCuiPresenter_Output_BetPhase(t *testing.T) {
	p := new(OasisPokerCuiPresenter)
	m := new(interfaces.MockOasisPokerGame)
	setupOasisPokerCuiMockDefaults(m)

	result := p.Output(m, nil)
	assert.Contains(t, result, "チップ: 1000")
	assert.Contains(t, result, "フェーズ: BET")
}

func TestOasisPokerCuiPresenter_Output_ExchangePhase(t *testing.T) {
	p := new(OasisPokerCuiPresenter)
	m := new(interfaces.MockOasisPokerGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.OasisPokerPhaseExchange).Maybe()
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
	m.On("GetJackpotBet").Return(0).Maybe()
	m.On("GetExchangeCount").Return(0).Maybe()
	m.On("GetExchangeFee").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetJackpotPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetPlayerHandRank").Return(domain.PokerHandHighCard).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "フェーズ: EXCHANGE")
	assert.Contains(t, result, "PLAYER")
	assert.Contains(t, result, "DEALER")
	assert.Contains(t, result, "??")
}

func TestOasisPokerCuiPresenter_Output_ActionPhase_ShowsExchangeInfo(t *testing.T) {
	p := new(OasisPokerCuiPresenter)
	m := new(interfaces.MockOasisPokerGame)
	m.On("GetChips").Return(700).Maybe()
	m.On("GetPhase").Return(domain.OasisPokerPhaseAction).Maybe()
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
	m.On("GetJackpotBet").Return(0).Maybe()
	m.On("GetExchangeCount").Return(2).Maybe()
	m.On("GetExchangeFee").Return(200).Maybe()
	m.On("GetPlayBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetJackpotPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetPlayerHandRank").Return(domain.PokerHandHighCard).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "フェーズ: ACTION")
	// First dealer card visible
	assert.Contains(t, result, "HEART 13")
	// Remaining masked
	assert.Contains(t, result, "??")
}

func TestOasisPokerCuiPresenter_Output_EndPhase_PlayerWins(t *testing.T) {
	p := new(OasisPokerCuiPresenter)
	m := new(interfaces.MockOasisPokerGame)
	m.On("GetChips").Return(1400).Maybe()
	m.On("GetPhase").Return(domain.OasisPokerPhaseEnd).Maybe()
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
	m.On("GetJackpotBet").Return(0).Maybe()
	m.On("GetExchangeCount").Return(0).Maybe()
	m.On("GetExchangeFee").Return(0).Maybe()
	m.On("GetPlayBet").Return(200).Maybe()
	m.On("GetResult").Return(domain.GameResultWin).Maybe()
	m.On("GetAntePayout").Return(200).Maybe()
	m.On("GetPlayPayout").Return(800).Maybe()
	m.On("GetJackpotPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(1000).Maybe()
	m.On("GetDealerQualified").Return(true).Maybe()
	m.On("GetPlayerHandRank").Return(domain.PokerHandThreeOfAKind).Maybe()
	m.On("GetDealerHandRank").Return(domain.PokerHandOnePair).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "フェーズ: END")
	assert.Contains(t, result, "プレイヤーの勝ち")
	assert.Contains(t, result, "(Qualified)")
	assert.Contains(t, result, "合計払戻し: 1000")
}

func TestOasisPokerCuiPresenter_Output_EndPhase_Fold(t *testing.T) {
	p := new(OasisPokerCuiPresenter)
	m := new(interfaces.MockOasisPokerGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.OasisPokerPhaseEnd).Maybe()
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
	m.On("GetJackpotBet").Return(0).Maybe()
	m.On("GetExchangeCount").Return(0).Maybe()
	m.On("GetExchangeFee").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResultLose).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetJackpotPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetPlayerHandRank").Return(domain.PokerHandHighCard).Maybe()
	m.On("GetDealerHandRank").Return(domain.PokerHandHighCard).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "プレイヤーがフォールド")
}

func TestOasisPokerCuiPresenter_Output_Error(t *testing.T) {
	p := new(OasisPokerCuiPresenter)
	m := new(interfaces.MockOasisPokerGame)
	setupOasisPokerCuiMockDefaults(m)

	result := p.Output(m, domain.NewDomainError(domain.ErrWrongPhase, "wrong phase"))
	assert.Contains(t, result, "wrong phase")
}

func TestOasisPokerCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(OasisPokerCuiPresenter)
	m := new(interfaces.MockOasisPokerGame)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()

	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "棋譜はありません")
}

func TestOasisPokerCuiPresenter_HintOutput(t *testing.T) {
	p := new(OasisPokerCuiPresenter)
	c := func(suit, val int) *domain.Card { return domain.NewCard(suit, val, false) }

	t.Run("exchange lists low cards to swap", func(t *testing.T) {
		m := new(interfaces.MockOasisPokerGame)
		m.On("GetPhase").Return(domain.OasisPokerPhaseExchange)
		// Pair of Kings held; low singles 3,5,7 recommended for exchange.
		m.On("GetPlayerHand").Return([]*domain.Card{
			c(domain.CardDesignSpade, 13), c(domain.CardDesignHeart, 13),
			c(domain.CardDesignClover, 3), c(domain.CardDesignDiamond, 5), c(domain.CardDesignSpade, 7),
		})
		out := p.HintOutput(m)
		prefix := strings.SplitN(i18n.T("oasispoker.hintExchange"), "{{", 2)[0]
		assert.Contains(t, out, prefix)
		assert.Contains(t, out, "[2]")
		assert.NotContains(t, out, "[0]") // king held
	})

	t.Run("exchange recommends stand when all held", func(t *testing.T) {
		m := new(interfaces.MockOasisPokerGame)
		m.On("GetPhase").Return(domain.OasisPokerPhaseExchange)
		m.On("GetPlayerHand").Return([]*domain.Card{
			c(domain.CardDesignSpade, 1), c(domain.CardDesignHeart, 13),
			c(domain.CardDesignClover, 12), c(domain.CardDesignDiamond, 11), c(domain.CardDesignSpade, 1),
		})
		assert.Contains(t, p.HintOutput(m), i18n.T("oasispoker.hintStand"))
	})

	t.Run("action recommends play", func(t *testing.T) {
		m := new(interfaces.MockOasisPokerGame)
		m.On("GetPhase").Return(domain.OasisPokerPhaseAction)
		m.On("RecommendPlay").Return(true)
		assert.Contains(t, p.HintOutput(m), i18n.T("oasispoker.hintPlay"))
	})

	t.Run("action recommends fold", func(t *testing.T) {
		m := new(interfaces.MockOasisPokerGame)
		m.On("GetPhase").Return(domain.OasisPokerPhaseAction)
		m.On("RecommendPlay").Return(false)
		assert.Contains(t, p.HintOutput(m), i18n.T("oasispoker.hintFold"))
	})

	t.Run("no hint outside decision phases", func(t *testing.T) {
		m := new(interfaces.MockOasisPokerGame)
		m.On("GetPhase").Return(domain.OasisPokerPhaseBet)
		assert.Contains(t, p.HintOutput(m), i18n.T("oasispoker.hintNone"))
	})
}

// #5595: 成立/不成立のバッジは出ているのに、その条件はどこにも書かれておらず、
// アンティがプッシュになる理由が読めなかった。
func TestOasisPokerCuiPresenter_ExplainsTheQualifyRule(t *testing.T) {
	i18n.SetLang("ja")
	g := domain.NewDefaultOasisPoker()
	g.Reset()
	require.NoError(t, g.Bet(10, 0))
	require.NoError(t, g.Stand())
	require.NoError(t, g.Play()) // 勝負してディーラーの手を開く

	out := new(OasisPokerCuiPresenter).Output(g, nil)
	// 成立でも不成立でも、条件そのものは出ること。
	assert.Contains(t, out, i18n.T("oasispoker.qualifyRule"))
}

// 決着前には出さない。まだディーラーの手が伏せられている。
func TestOasisPokerCuiPresenter_HidesTheQualifyRuleBeforeTheEnd(t *testing.T) {
	i18n.SetLang("ja")
	g := domain.NewDefaultOasisPoker()
	g.Reset()

	assert.NotContains(t, new(OasisPokerCuiPresenter).Output(g, nil), i18n.T("oasispoker.qualifyRule"))
}
