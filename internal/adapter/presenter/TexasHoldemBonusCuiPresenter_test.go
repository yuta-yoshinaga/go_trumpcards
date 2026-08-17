package presenter

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupTexasHoldemBonusCuiMockDefaults(m *interfaces.MockTexasHoldemBonusGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.TexasHoldemBonusPhaseBet).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetCommunity").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteBet").Return(0).Maybe()
	m.On("GetNextBetCost").Return(0).Maybe()
	m.On("GetBonusBet").Return(0).Maybe()
	m.On("GetFlopBet").Return(0).Maybe()
	m.On("GetTurnBet").Return(0).Maybe()
	m.On("GetRiverBet").Return(0).Maybe()
	m.On("GetTotalPlayBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetBonusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetPlayerBest").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerBest").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

func TestTexasHoldemBonusCuiPresenter_Output_BetPhase(t *testing.T) {
	p := new(TexasHoldemBonusCuiPresenter)
	m := new(interfaces.MockTexasHoldemBonusGame)
	setupTexasHoldemBonusCuiMockDefaults(m)

	result := p.Output(m, nil)
	assert.Contains(t, result, "チップ: 1000")
	assert.Contains(t, result, "フェーズ: BET")
}

func TestTexasHoldemBonusCuiPresenter_Output_PreFlopPhase(t *testing.T) {
	p := new(TexasHoldemBonusCuiPresenter)
	m := new(interfaces.MockTexasHoldemBonusGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.TexasHoldemBonusPhasePreFlop).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignClover, 13, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 5, false),
		domain.NewCard(domain.CardDesignDiamond, 7, false),
	}).Maybe()
	m.On("GetCommunity").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetNextBetCost").Return(0).Maybe()
	m.On("GetBonusBet").Return(0).Maybe()
	m.On("GetFlopBet").Return(0).Maybe()
	m.On("GetTurnBet").Return(0).Maybe()
	m.On("GetRiverBet").Return(0).Maybe()
	m.On("GetTotalPlayBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetBonusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "フェーズ: PRE-FLOP")
	assert.Contains(t, result, "PLAYER")
	assert.Contains(t, result, "DEALER")
	// Dealer cards hidden in pre-flop
	assert.Contains(t, result, "??")
}

func TestTexasHoldemBonusCuiPresenter_Output_FlopPhase(t *testing.T) {
	p := new(TexasHoldemBonusCuiPresenter)
	m := new(interfaces.MockTexasHoldemBonusGame)
	m.On("GetChips").Return(700).Maybe()
	m.On("GetPhase").Return(domain.TexasHoldemBonusPhaseFlop).Maybe()
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
	m.On("GetNextBetCost").Return(0).Maybe()
	m.On("GetBonusBet").Return(0).Maybe()
	m.On("GetFlopBet").Return(200).Maybe()
	m.On("GetTurnBet").Return(0).Maybe()
	m.On("GetRiverBet").Return(0).Maybe()
	m.On("GetTotalPlayBet").Return(200).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetBonusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "フェーズ: FLOP")
	assert.Contains(t, result, "BOARD")
}

func TestTexasHoldemBonusCuiPresenter_Output_EndPhase_PlayerWins(t *testing.T) {
	p := new(TexasHoldemBonusCuiPresenter)
	m := new(interfaces.MockTexasHoldemBonusGame)
	m.On("GetChips").Return(1500).Maybe()
	m.On("GetPhase").Return(domain.TexasHoldemBonusPhaseEnd).Maybe()
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
	m.On("GetNextBetCost").Return(0).Maybe()
	m.On("GetBonusBet").Return(0).Maybe()
	m.On("GetFlopBet").Return(200).Maybe()
	m.On("GetTurnBet").Return(0).Maybe()
	m.On("GetRiverBet").Return(0).Maybe()
	m.On("GetTotalPlayBet").Return(200).Maybe()
	m.On("GetResult").Return(domain.GameResultWin).Maybe()
	m.On("GetAntePayout").Return(200).Maybe()
	m.On("GetPlayPayout").Return(400).Maybe()
	m.On("GetBonusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(600).Maybe()
	m.On("GetPlayerHandRank").Return(domain.PokerHandOnePair).Maybe()
	m.On("GetDealerHandRank").Return(domain.PokerHandOnePair).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "フェーズ: END")
	assert.Contains(t, result, "プレイヤーの勝ち")
	assert.Contains(t, result, "合計払戻し: 600")
}

func TestTexasHoldemBonusCuiPresenter_Output_EndPhase_Fold(t *testing.T) {
	p := new(TexasHoldemBonusCuiPresenter)
	m := new(interfaces.MockTexasHoldemBonusGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.TexasHoldemBonusPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, false),
		domain.NewCard(domain.CardDesignClover, 2, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 13, false),
		domain.NewCard(domain.CardDesignDiamond, 13, false),
	}).Maybe()
	m.On("GetCommunity").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetNextBetCost").Return(0).Maybe()
	m.On("GetBonusBet").Return(0).Maybe()
	m.On("GetFlopBet").Return(0).Maybe()
	m.On("GetTurnBet").Return(0).Maybe()
	m.On("GetRiverBet").Return(0).Maybe()
	m.On("GetTotalPlayBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResultLose).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetBonusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "プレイヤーがフォールド")
}

func TestTexasHoldemBonusCuiPresenter_Output_EndPhase_Push(t *testing.T) {
	p := new(TexasHoldemBonusCuiPresenter)
	m := new(interfaces.MockTexasHoldemBonusGame)
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.TexasHoldemBonusPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 2, false),
		domain.NewCard(domain.CardDesignSpade, 3, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignClover, 2, false),
		domain.NewCard(domain.CardDesignClover, 3, false),
	}).Maybe()
	m.On("GetCommunity").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 5, false),
		domain.NewCard(domain.CardDesignDiamond, 6, false),
		domain.NewCard(domain.CardDesignSpade, 7, false),
		domain.NewCard(domain.CardDesignDiamond, 8, false),
		domain.NewCard(domain.CardDesignClover, 9, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetNextBetCost").Return(0).Maybe()
	m.On("GetBonusBet").Return(0).Maybe()
	m.On("GetFlopBet").Return(200).Maybe()
	m.On("GetTurnBet").Return(0).Maybe()
	m.On("GetRiverBet").Return(0).Maybe()
	m.On("GetTotalPlayBet").Return(200).Maybe()
	m.On("GetResult").Return(domain.GameResultDraw).Maybe()
	m.On("GetAntePayout").Return(200).Maybe()
	m.On("GetPlayPayout").Return(200).Maybe()
	m.On("GetBonusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(400).Maybe()
	m.On("GetPlayerHandRank").Return(domain.PokerHandStraight).Maybe()
	m.On("GetDealerHandRank").Return(domain.PokerHandStraight).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "プッシュ")
}

func TestTexasHoldemBonusCuiPresenter_Output_EndPhase_DealerWins(t *testing.T) {
	p := new(TexasHoldemBonusCuiPresenter)
	m := new(interfaces.MockTexasHoldemBonusGame)
	m.On("GetChips").Return(700).Maybe()
	m.On("GetPhase").Return(domain.TexasHoldemBonusPhaseEnd).Maybe()
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
	m.On("GetNextBetCost").Return(0).Maybe()
	m.On("GetBonusBet").Return(0).Maybe()
	m.On("GetFlopBet").Return(200).Maybe()
	m.On("GetTurnBet").Return(0).Maybe()
	m.On("GetRiverBet").Return(0).Maybe()
	m.On("GetTotalPlayBet").Return(200).Maybe()
	m.On("GetResult").Return(domain.GameResultLose).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetBonusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(domain.PokerHandHighCard).Maybe()
	m.On("GetDealerHandRank").Return(domain.PokerHandOnePair).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "ディーラーの勝ち")
}

func TestTexasHoldemBonusCuiPresenter_Output_Error(t *testing.T) {
	p := new(TexasHoldemBonusCuiPresenter)
	m := new(interfaces.MockTexasHoldemBonusGame)
	setupTexasHoldemBonusCuiMockDefaults(m)

	result := p.Output(m, domain.NewDomainError(domain.ErrWrongPhase, "wrong phase"))
	assert.Contains(t, result, "wrong phase")
}

func TestTexasHoldemBonusCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(TexasHoldemBonusCuiPresenter)
	m := new(interfaces.MockTexasHoldemBonusGame)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()

	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "棋譜はありません")
}

func TestTexasHoldemBonusCuiPresenter_PhaseStr_Unknown(t *testing.T) {
	p := new(TexasHoldemBonusCuiPresenter)
	m := new(interfaces.MockTexasHoldemBonusGame)
	m.On("GetChips").Return(0).Maybe()
	m.On("GetPhase").Return(999).Maybe()
	m.On("GetNextBetCost").Return(0).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetCommunity").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "フェーズ: UNKNOWN")
}

// **アクション中はアンテ額も実コストも画面のどこにも出ていなかった (#4698)。**
// Web は Play/Raise ボタンに ante×倍率をラベル表示している。
func TestTexasHoldemBonusCuiPresenter_BetCostLine(t *testing.T) {
	p := new(TexasHoldemBonusCuiPresenter)
	game := func(phase, ante, cost int) *interfaces.MockTexasHoldemBonusGame {
		m := new(interfaces.MockTexasHoldemBonusGame)
		// **先に登録した期待が勝つ。**defaults より前に置く。
		m.On("GetPhase").Return(phase)
		m.On("GetAnteBet").Return(ante)
		m.On("GetNextBetCost").Return(cost)
		setupTexasHoldemBonusCuiMockDefaults(m)
		return m
	}

	t.Run("pre-flop shows the ante and the 2x Play cost", func(t *testing.T) {
		out := p.Output(game(domain.TexasHoldemBonusPhasePreFlop, 100, 200), nil)
		assert.Contains(t, out, "アンテ: 100")
		assert.Contains(t, out, "プレイ: 200")
	})

	// **プリフロップとフロップで倍率が違う。**同じ文言だと 2× と 1× の
	// 区別が付かない。
	t.Run("flop shows the 1x Raise cost under its own label", func(t *testing.T) {
		out := p.Output(game(domain.TexasHoldemBonusPhaseFlop, 100, 100), nil)
		assert.Contains(t, out, "レイズ: 100")
		assert.NotContains(t, out, "プレイ: ")
	})

	t.Run("turn shows the raise cost too", func(t *testing.T) {
		assert.Contains(t,
			p.Output(game(domain.TexasHoldemBonusPhaseTurn, 50, 50), nil),
			"レイズ: 50")
	})

	// **END の anteLine と重ならないこと。**結果表示の行とは別物。
	t.Run("no cost line once the round is over", func(t *testing.T) {
		out := p.Output(game(domain.TexasHoldemBonusPhaseEnd, 100, 0), nil)
		assert.NotContains(t, out, "レイズ: ")
		assert.NotContains(t, out, "プレイ: ")
	})

	t.Run("no cost line before the ante is placed", func(t *testing.T) {
		out := p.Output(game(domain.TexasHoldemBonusPhaseBet, 0, 0), nil)
		assert.NotContains(t, out, "チップ / ")
	})
}

// #5529: Web は BET フェーズにボーナスの配当表を出しているのに、CUI は
// "bet <ante> <bonus>" で実チップを賭けさせながら何が当たるのか言っていなかった。
func TestTexasHoldemBonusCuiPresenter_Output_BonusPaytable(t *testing.T) {
	p := new(TexasHoldemBonusCuiPresenter)

	outputInPhase := func(phase int) string {
		m := new(interfaces.MockTexasHoldemBonusGame)
		m.On("GetPhase").Return(phase)
		setupTexasHoldemBonusCuiMockDefaults(m)
		return p.Output(m, nil)
	}

	betOut := outputInPhase(domain.TexasHoldemBonusPhaseBet)
	assert.Contains(t, betOut, i18n.T("texasholdembonus.bonusPayHeader"))

	// **倍率はドメインの定数から出す。**画面に書き写すと、配当を1つ直したとき
	// 表だけが古いまま残る。
	for _, tc := range []struct {
		key  string
		mult int
	}{
		{"texasholdembonus.bonusPayAA", domain.TexasHoldemBonusBonusPayAA},
		{"texasholdembonus.bonusPayAKSuited", domain.TexasHoldemBonusBonusPayAKSuited},
		{"texasholdembonus.bonusPayAQAJSuited", domain.TexasHoldemBonusBonusPayAQAJSuited},
		{"texasholdembonus.bonusPayAKOff", domain.TexasHoldemBonusBonusPayAKOff},
		{"texasholdembonus.bonusPayKKQQJJ", domain.TexasHoldemBonusBonusPayKKQQJJ},
		{"texasholdembonus.bonusPayAQAJOff", domain.TexasHoldemBonusBonusPayAQAJOff},
		{"texasholdembonus.bonusPayMediumPair", domain.TexasHoldemBonusBonusPayMediumPair},
	} {
		assert.Contains(t, betOut, i18n.Tf(tc.key, "mult", strconv.Itoa(tc.mult)), tc.key)
	}

	// **賭け終わった後には出さない。**もう選べないものの説明は場所を取るだけ。
	assert.NotContains(t, outputInPhase(domain.TexasHoldemBonusPhasePreFlop),
		i18n.T("texasholdembonus.bonusPayHeader"))
	assert.NotContains(t, outputInPhase(domain.TexasHoldemBonusPhaseEnd),
		i18n.T("texasholdembonus.bonusPayHeader"))
}
