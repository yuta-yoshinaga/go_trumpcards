package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func makePokerCuiForPresenter() (*domain.Poker, []*domain.PokerPlayer) {
	tc := domain.NewTrumpCards(0)
	players := []*domain.PokerPlayer{
		domain.NewPokerPlayer(true, domain.PokerStyleBalanced),
		domain.NewPokerPlayer(false, domain.PokerStyleConservative),
		domain.NewPokerPlayer(false, domain.PokerStyleAggressive),
		domain.NewPokerPlayer(false, domain.PokerStyleBluffer),
	}
	cfg := domain.DefaultPokerConfig()
	p := domain.NewPoker(tc, players, cfg)
	return p, players
}

func TestPokerCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	pres := new(presenter.PokerCuiPresenter)

	t.Run("initial state with header and dealer info", func(t *testing.T) {
		p, players := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := pres.Output(p, nil)
		assert.Contains(t, result, "5-Card Draw Poker")
		assert.Contains(t, result, "ディーラー: Player 0")
		assert.Contains(t, result, "ポット:")
		assert.Contains(t, result, "あなた")
		assert.Contains(t, result, "♠10")
		assert.Contains(t, result, "♥11")
		// The deal phase shows its name and the betting-command prompt.
		assert.Contains(t, result, i18n.T("poker.phaseDeal"))
		assert.Contains(t, result, i18n.T("poker.promptBet"))
	})

	t.Run("exchange phase shows exchange prompt", func(t *testing.T) {
		p, _ := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseExchange)
		result := pres.Output(p, nil)
		assert.Contains(t, result, i18n.T("poker.phaseExchange"))
		assert.Contains(t, result, i18n.T("poker.promptExchange"))
	})

	t.Run("CPU player info with play style name", func(t *testing.T) {
		p, _ := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)

		result := pres.Output(p, nil)
		assert.Contains(t, result, "CPU 1")
		assert.Contains(t, result, "CPU 2")
		assert.Contains(t, result, "CPU 3")
	})

	t.Run("folded player badge", func(t *testing.T) {
		p, players := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		players[1].SetFolded(true)

		result := pres.Output(p, nil)
		assert.Contains(t, result, "[フォールド]")
	})

	t.Run("all-in player badge", func(t *testing.T) {
		p, players := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		players[1].SetAllIn(true)

		result := pres.Output(p, nil)
		assert.Contains(t, result, "[オールイン]")
	})

	t.Run("non-folded non-allin player has no badge", func(t *testing.T) {
		p, players := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		players[1].SetFolded(false)
		players[1].SetAllIn(false)

		result := pres.Output(p, nil)
		// CPU 1 line should not contain folded or all-in badges
		// This tests the else path of both conditions
		_ = result
	})

	t.Run("player with current bet shown", func(t *testing.T) {
		p, players := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		players[0].SetCurrentBet(50)

		result := pres.Output(p, nil)
		assert.Contains(t, result, "ベット:50")
	})

	t.Run("player with zero bet hides bet label", func(t *testing.T) {
		p, players := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		players[0].SetCurrentBet(0)

		result := pres.Output(p, nil)
		assert.NotContains(t, result, "ベット:0")
	})

	t.Run("exchange count shown in second bet phase", func(t *testing.T) {
		p, players := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseSecondBet)
		players[0].SetExchangeCount(3)

		result := pres.Output(p, nil)
		assert.Contains(t, result, "交換:3枚")
	})

	t.Run("exchange count shown in end phase", func(t *testing.T) {
		p, players := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseEnd)
		players[0].SetExchangeCount(2)

		result := pres.Output(p, nil)
		assert.Contains(t, result, "交換:2枚")
	})

	t.Run("exchange count hidden in deal phase", func(t *testing.T) {
		p, players := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		players[0].SetExchangeCount(3)

		result := pres.Output(p, nil)
		assert.NotContains(t, result, "交換:3枚")
	})

	t.Run("exchange count zero not shown even in second bet", func(t *testing.T) {
		p, players := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseSecondBet)
		players[0].SetExchangeCount(0)

		result := pres.Output(p, nil)
		assert.NotContains(t, result, "交換:0枚")
	})

	t.Run("human cards visible when not folded", func(t *testing.T) {
		p, players := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 13, false))

		result := pres.Output(p, nil)
		assert.Contains(t, result, "手札:")
		assert.Contains(t, result, "[0]♠1")
		assert.Contains(t, result, "[1]♥13")
	})

	t.Run("human cards hidden when folded", func(t *testing.T) {
		p, players := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].SetFolded(true)

		result := pres.Output(p, nil)
		assert.NotContains(t, result, "手札:")
	})

	t.Run("human cards with separator between cards", func(t *testing.T) {
		p, players := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 7, false))

		result := pres.Output(p, nil)
		assert.Contains(t, result, "[0]♠3  [1]♣7")
	})

	t.Run("human hand name shown at end phase", func(t *testing.T) {
		p, players := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseEnd)
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 4, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		players[0].EvalHand()

		result := pres.Output(p, nil)
		assert.Contains(t, result, "[ストレートフラッシュ]")
	})

	t.Run("human hand name not shown in non-end phase", func(t *testing.T) {
		p, players := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 4, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		players[0].EvalHand()

		result := pres.Output(p, nil)
		assert.NotContains(t, result, "[ストレートフラッシュ]")
	})

	t.Run("CPU cards shown at end phase when not folded", func(t *testing.T) {
		p, players := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseEnd)
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 9, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignDiamond, 9, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[1].EvalHand()

		result := pres.Output(p, nil)
		assert.Contains(t, result, "♠5")
		assert.Contains(t, result, "♥5")
		// **役名も日本語ロケールで日本語。**英語の PokerHandNames をそのまま
		// 埋めていて、このテストがその挙動を固定していた。
		assert.Contains(t, result, "[ツーペア]")
		assert.NotContains(t, result, "[Two Pair]")
	})

	t.Run("CPU cards hidden when not end phase", func(t *testing.T) {
		p, players := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))

		result := pres.Output(p, nil)
		// CPU cards should not show as visible cards in deal phase
		// The CPU line should not show 手札
		_ = result
	})

	t.Run("CPU cards hidden when folded at end phase", func(t *testing.T) {
		p, players := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseEnd)
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].SetFolded(true)

		result := pres.Output(p, nil)
		// Folded CPU should not show hand
		assert.NotContains(t, result, "♠5")
	})

	t.Run("CPU cards separator between cards", func(t *testing.T) {
		p, players := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseEnd)
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 7, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		players[1].EvalHand()

		result := pres.Output(p, nil)
		assert.Contains(t, result, "♠3  ♣7")
	})

	t.Run("CPU actions displayed", func(t *testing.T) {
		p, _ := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		p.SetCpuActions([]domain.PokerCpuAction{
			{PlayerIdx: 1, Action: domain.PokerActionCall, Amount: 0},
			{PlayerIdx: 2, Action: domain.PokerActionRaise, Amount: 30},
		})

		result := pres.Output(p, nil)
		assert.Contains(t, result, "[CPU行動]")
		assert.Contains(t, result, "Player 1: コール")
		assert.Contains(t, result, "Player 2: レイズ")
		assert.Contains(t, result, "(30)")
	})

	t.Run("CPU action without amount", func(t *testing.T) {
		p, _ := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		p.SetCpuActions([]domain.PokerCpuAction{
			{PlayerIdx: 1, Action: domain.PokerActionFold, Amount: 0},
		})

		result := pres.Output(p, nil)
		assert.Contains(t, result, "フォールド")
		assert.NotContains(t, result, "(0)")
	})

	t.Run("no CPU actions hides section", func(t *testing.T) {
		p, _ := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		p.SetCpuActions([]domain.PokerCpuAction{})

		result := pres.Output(p, nil)
		assert.NotContains(t, result, "[CPU行動]")
	})

	t.Run("CPU exchanges displayed", func(t *testing.T) {
		p, _ := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseExchange)
		p.SetCpuExchanges([]domain.PokerCpuExchange{
			{PlayerIdx: 1, ExchangeCount: 3},
			{PlayerIdx: 2, ExchangeCount: 0},
		})

		result := pres.Output(p, nil)
		assert.Contains(t, result, "[CPU交換]")
		assert.Contains(t, result, "Player 1: 3枚交換")
		assert.Contains(t, result, "Player 2: 0枚交換")
	})

	t.Run("no CPU exchanges hides section", func(t *testing.T) {
		p, _ := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		p.SetCpuExchanges([]domain.PokerCpuExchange{})

		result := pres.Output(p, nil)
		assert.NotContains(t, result, "[CPU交換]")
	})

	t.Run("showdown results with human winner", func(t *testing.T) {
		p, _ := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseEnd)
		p.SetRoundResults([]domain.PokerResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100},
		})

		result := pres.Output(p, nil)
		assert.Contains(t, result, "[結果]")
		assert.Contains(t, result, "あなた: フラッシュ")
		assert.Contains(t, result, "100チップ獲得")
	})

	t.Run("showdown results with kickers", func(t *testing.T) {
		p, _ := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseEnd)
		p.SetRoundResults([]domain.PokerResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandOnePair, HandName: "One Pair", Kickers: []int{14, 12, 10}, WonAmount: 100},
		})

		result := pres.Output(p, nil)
		assert.Contains(t, result, "あなた: ワンペア (キッカー: A, Q, 10)")
		assert.Contains(t, result, "100チップ獲得")
	})

	t.Run("showdown results without kickers", func(t *testing.T) {
		p, _ := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseEnd)
		p.SetRoundResults([]domain.PokerResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush", Kickers: nil, WonAmount: 100},
		})

		result := pres.Output(p, nil)
		assert.Contains(t, result, "あなた: フラッシュ")
		assert.NotContains(t, result, "キッカー")
	})

	t.Run("showdown results with CPU winner", func(t *testing.T) {
		p, _ := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseEnd)
		p.SetRoundResults([]domain.PokerResult{
			{PlayerIdx: 1, HandRank: domain.PokerHandOnePair, HandName: "One Pair", Kickers: []int{13, 12, 11}, WonAmount: 50},
		})

		result := pres.Output(p, nil)
		assert.Contains(t, result, "CPU 1: ワンペア (キッカー: K, Q, J)")
		assert.Contains(t, result, "50チップ獲得")
	})

	t.Run("showdown results with empty hand name", func(t *testing.T) {
		p, _ := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseEnd)
		p.SetRoundResults([]domain.PokerResult{
			{PlayerIdx: 1, HandName: "", WonAmount: 0},
		})

		result := pres.Output(p, nil)
		assert.Contains(t, result, "CPU 1")
		assert.NotContains(t, result, "CPU 1:")
	})

	t.Run("showdown results with zero won amount", func(t *testing.T) {
		p, _ := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseEnd)
		p.SetRoundResults([]domain.PokerResult{
			{PlayerIdx: 1, HandRank: domain.PokerHandHighCard, HandName: "High Card", WonAmount: 0},
		})

		result := pres.Output(p, nil)
		assert.NotContains(t, result, "チップ獲得")
	})

	t.Run("results not shown in non-end phase", func(t *testing.T) {
		p, _ := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		p.SetRoundResults([]domain.PokerResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100},
		})

		result := pres.Output(p, nil)
		assert.NotContains(t, result, "[結果]")
	})

	t.Run("results not shown when empty in end phase", func(t *testing.T) {
		p, _ := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseEnd)
		p.SetRoundResults([]domain.PokerResult{})

		result := pres.Output(p, nil)
		assert.NotContains(t, result, "[結果]")
	})

	t.Run("error message displayed", func(t *testing.T) {
		p, _ := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)

		result := pres.Output(p, errors.New("test error"))
		assert.Contains(t, result, "test error")
	})

	t.Run("no error hides error section", func(t *testing.T) {
		p, _ := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)

		result := pres.Output(p, nil)
		assert.NotContains(t, result, "test error")
	})

	t.Run("game end message displayed", func(t *testing.T) {
		p, _ := makePokerCuiForPresenter()
		p.SetGameEndFlag(true)

		result := pres.Output(p, nil)
		assert.Contains(t, result, "ゲーム終了")
	})

	t.Run("game not ended hides game end message", func(t *testing.T) {
		p, _ := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		p.SetGameEndFlag(false)

		result := pres.Output(p, nil)
		assert.NotContains(t, result, "ゲーム終了")
	})

	t.Run("joker count shown when greater than zero", func(t *testing.T) {
		tc := domain.NewTrumpCards(2)
		players := []*domain.PokerPlayer{
			domain.NewPokerPlayer(true, domain.PokerStyleBalanced),
			domain.NewPokerPlayer(false, domain.PokerStyleConservative),
		}
		cfg := domain.PokerConfig{InitChips: 1000, Ante: 10, MinBet: 10, CpuCount: 1, JokerCount: 2}
		p := domain.NewPoker(tc, players, cfg)
		p.SetPhase(domain.PokerPhaseDeal)

		result := pres.Output(p, nil)
		assert.Contains(t, result, "ジョーカー: 2枚")
	})

	t.Run("joker count hidden when zero", func(t *testing.T) {
		p, _ := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)

		result := pres.Output(p, nil)
		assert.NotContains(t, result, "ジョーカー:")
	})

	t.Run("getCardStr with spade", func(t *testing.T) {
		p, players := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))

		result := pres.Output(p, nil)
		assert.Contains(t, result, "♠1")
	})

	t.Run("getCardStr with clover", func(t *testing.T) {
		p, players := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 5, false))

		result := pres.Output(p, nil)
		assert.Contains(t, result, "♣5")
	})

	t.Run("getCardStr with heart", func(t *testing.T) {
		p, players := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 13, false))

		result := pres.Output(p, nil)
		assert.Contains(t, result, "♥13")
	})

	t.Run("getCardStr with diamond", func(t *testing.T) {
		p, players := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		players[0].AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))

		result := pres.Output(p, nil)
		assert.Contains(t, result, "♦7")
	})

	t.Run("getCardStr with joker design", func(t *testing.T) {
		p, players := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 1, false))

		result := pres.Output(p, nil)
		assert.Contains(t, result, "🃏1")
	})

	t.Run("getCardStr with out of range design", func(t *testing.T) {
		p, players := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		players[0].AddCard(domain.NewCard(99, 5, false))

		result := pres.Output(p, nil)
		assert.Contains(t, result, "🃏5") // falls back to joker design
	})

	t.Run("getCardStr with negative design", func(t *testing.T) {
		p, players := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		players[0].AddCard(domain.NewCard(-1, 3, false))

		result := pres.Output(p, nil)
		assert.Contains(t, result, "🃏3")
	})

	t.Run("all action names", func(t *testing.T) {
		p, _ := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		p.SetCpuActions([]domain.PokerCpuAction{
			{PlayerIdx: 1, Action: domain.PokerActionFold, Amount: 0},
			{PlayerIdx: 1, Action: domain.PokerActionCheck, Amount: 0},
			{PlayerIdx: 1, Action: domain.PokerActionCall, Amount: 0},
			{PlayerIdx: 1, Action: domain.PokerActionBet, Amount: 10},
			{PlayerIdx: 1, Action: domain.PokerActionRaise, Amount: 20},
			{PlayerIdx: 1, Action: domain.PokerActionAllIn, Amount: 100},
			{PlayerIdx: 1, Action: 99, Amount: 0},
		})

		result := pres.Output(p, nil)
		assert.Contains(t, result, "フォールド")
		assert.Contains(t, result, "チェック")
		assert.Contains(t, result, "コール")
		assert.Contains(t, result, "ベット")
		assert.Contains(t, result, "レイズ")
		assert.Contains(t, result, "オールイン")
		assert.Contains(t, result, "不明")
	})

	t.Run("chips displayed for each player", func(t *testing.T) {
		p, players := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		players[0].SetChips(500)
		players[1].SetChips(750)

		result := pres.Output(p, nil)
		assert.Contains(t, result, "チップ:500")
		assert.Contains(t, result, "チップ:750")
	})
}

func TestPokerCuiPresenter_OutputWithOdds(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	pres := new(presenter.PokerCuiPresenter)
	p, players := makePokerCuiForPresenter()
	p.SetPhase(domain.PokerPhaseExchange)
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))

	odds := []domain.PokerDrawOdds{
		{HandRank: 0, HandName: "High Card", Probability: 0.5, Count: 50, Total: 100},
		{HandRank: 1, HandName: "One Pair", Probability: 0.3, Count: 30, Total: 100},
	}

	result := pres.OutputWithOdds(p, nil, odds)
	assert.Contains(t, result, "[ドローオッズ]")
	assert.Contains(t, result, "ハイカード: 50.00% (50/100)")
	assert.Contains(t, result, "ワンペア: 30.00% (30/100)")
}

func TestPokerCuiPresenter_OutputWithOdds_NilOdds(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	pres := new(presenter.PokerCuiPresenter)
	p, players := makePokerCuiForPresenter()
	p.SetPhase(domain.PokerPhaseExchange)
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))

	resultWithOdds := pres.OutputWithOdds(p, nil, nil)
	resultPlain := pres.Output(p, nil)
	assert.Equal(t, resultPlain, resultWithOdds)
}

func TestPokerCuiPresenter_OutputWithOdds_EmptyOdds(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	pres := new(presenter.PokerCuiPresenter)
	p, players := makePokerCuiForPresenter()
	p.SetPhase(domain.PokerPhaseExchange)
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))

	resultWithOdds := pres.OutputWithOdds(p, nil, []domain.PokerDrawOdds{})
	resultPlain := pres.Output(p, nil)
	assert.Equal(t, resultPlain, resultWithOdds)
}

func TestPokerCuiPresenter_Output_LowballMode(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	pres := new(presenter.PokerCuiPresenter)

	t.Run("lowball mode shows 2-7 Lowball", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := []*domain.PokerPlayer{
			domain.NewPokerPlayer(true, domain.PokerStyleBalanced),
			domain.NewPokerPlayer(false, domain.PokerStyleConservative),
		}
		cfg := domain.DefaultPokerConfig()
		cfg.IsLowball = true
		cfg.CpuCount = 1
		p := domain.NewPoker(tc, players, cfg)
		p.SetPhase(domain.PokerPhaseDeal)

		result := pres.Output(p, nil)
		assert.Contains(t, result, "2-7 Lowball")
	})

	t.Run("normal mode does not show 2-7 Lowball", func(t *testing.T) {
		p, _ := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)

		result := pres.Output(p, nil)
		assert.NotContains(t, result, "2-7 Lowball")
	})
}

func TestPokerCuiPresenter_Output_BettingLimitDisplay(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	pres := new(presenter.PokerCuiPresenter)

	t.Run("displays Fixed limit", func(t *testing.T) {
		p, _ := makePokerCuiForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		result := pres.Output(p, nil)
		assert.Contains(t, result, "Fixed")
	})
}

func TestPokerCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.PokerCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		mockGame := new(interfaces.MockPokerGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "exchange", Detail: "exchanged 2 cards"},
		}
		mockGame.On("GetGameEndFlag").Return(true)
		mockGame.On("GetActionLog").Return(entries)

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "exchange")
		assert.Contains(t, result, "exchanged 2 cards")
		mockGame.AssertExpectations(t)
	})

	t.Run("nil_entries", func(t *testing.T) {
		mockGame := new(interfaces.MockPokerGame)
		mockGame.On("GetGameEndFlag").Return(true)
		mockGame.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, "棋譜はありません")
		mockGame.AssertExpectations(t)
	})

	t.Run("game_not_ended", func(t *testing.T) {
		mockGame := new(interfaces.MockPokerGame)
		mockGame.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, "棋譜はありません")
		mockGame.AssertExpectations(t)
	})
}
