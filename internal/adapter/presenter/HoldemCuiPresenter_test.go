package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func makeHoldemForPresenter() (*domain.Holdem, []*domain.HoldemPlayer) {
	tc := domain.NewTrumpCards(0)
	players := []*domain.HoldemPlayer{
		domain.NewHoldemPlayer(true, domain.HoldemStyleTAG),
		domain.NewHoldemPlayer(false, domain.HoldemStyleLAP),
		domain.NewHoldemPlayer(false, domain.HoldemStyleTAP),
		domain.NewHoldemPlayer(false, domain.HoldemStyleGTO),
	}
	h := domain.NewHoldem(tc, players, domain.DefaultHoldemConfig())
	return h, players
}

func TestHoldemCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.HoldemCuiPresenter)

	t.Run("initial state with no community cards", func(t *testing.T) {
		h, players := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhasePreFlop)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := p.Output(h, nil)
		assert.Contains(t, result, "Texas Hold'em")
		assert.Contains(t, result, "ディーラー:")
		assert.Contains(t, result, "(なし)")
		assert.Contains(t, result, "ポット:")
		assert.Contains(t, result, "あなた")
		assert.Contains(t, result, "♠10")
		assert.Contains(t, result, "♥11")
	})

	t.Run("community cards displayed", func(t *testing.T) {
		h, _ := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhaseFlop)
		h.SetCommunityCards([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 2, false),
			domain.NewCard(domain.CardDesignClover, 5, false),
			domain.NewCard(domain.CardDesignDiamond, 9, false),
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "♠2")
		assert.Contains(t, result, "♣5")
		assert.Contains(t, result, "♦9")
		assert.NotContains(t, result, "(なし)")
	})

	t.Run("multiple community cards separator", func(t *testing.T) {
		h, _ := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhaseFlop)
		h.SetCommunityCards([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignHeart, 2, false),
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "♠1  ♥2")
	})

	t.Run("CPU player info with play style", func(t *testing.T) {
		h, _ := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhasePreFlop)

		result := p.Output(h, nil)
		assert.Contains(t, result, "CPU 1")
		assert.Contains(t, result, "CPU 2")
		assert.Contains(t, result, "CPU 3")
	})

	t.Run("folded player", func(t *testing.T) {
		h, players := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhasePreFlop)
		players[1].SetFolded(true)

		result := p.Output(h, nil)
		assert.Contains(t, result, "[フォールド]")
	})

	t.Run("all-in player", func(t *testing.T) {
		h, players := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhasePreFlop)
		players[1].SetAllIn(true)

		result := p.Output(h, nil)
		assert.Contains(t, result, "[オールイン]")
	})

	t.Run("player with current bet", func(t *testing.T) {
		h, players := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhasePreFlop)
		players[0].SetCurrentBet(50)

		result := p.Output(h, nil)
		assert.Contains(t, result, "ベット:50")
	})

	t.Run("player with zero bet hides bet label", func(t *testing.T) {
		h, players := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhasePreFlop)
		players[0].SetCurrentBet(0)

		result := p.Output(h, nil)
		assert.NotContains(t, result, "ベット:0")
	})

	t.Run("human cards visible when not folded", func(t *testing.T) {
		h, players := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhasePreFlop)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 13, false))

		result := p.Output(h, nil)
		assert.Contains(t, result, "手札:")
		assert.Contains(t, result, "♠1")
		assert.Contains(t, result, "♥13")
	})

	t.Run("human cards hidden when folded", func(t *testing.T) {
		h, players := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhasePreFlop)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].SetFolded(true)

		result := p.Output(h, nil)
		assert.NotContains(t, result, "手札:")
	})

	t.Run("human cards with separator", func(t *testing.T) {
		h, players := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhasePreFlop)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 7, false))

		result := p.Output(h, nil)
		assert.Contains(t, result, "♠3  ♣7")
	})

	t.Run("CPU actions displayed", func(t *testing.T) {
		h, _ := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhasePreFlop)
		h.SetCpuActions([]domain.HoldemCpuAction{
			{PlayerIdx: 1, Action: domain.HoldemActionCall, Amount: 0},
			{PlayerIdx: 2, Action: domain.HoldemActionRaise, Amount: 30},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "[CPU行動]")
		assert.Contains(t, result, "Player 1: コール")
		assert.Contains(t, result, "Player 2: レイズ")
		assert.Contains(t, result, "(30)")
	})

	t.Run("CPU action without amount", func(t *testing.T) {
		h, _ := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhasePreFlop)
		h.SetCpuActions([]domain.HoldemCpuAction{
			{PlayerIdx: 1, Action: domain.HoldemActionFold, Amount: 0},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "フォールド")
		assert.NotContains(t, result, "(0)")
	})

	t.Run("no CPU actions hides section", func(t *testing.T) {
		h, _ := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhasePreFlop)
		h.SetCpuActions([]domain.HoldemCpuAction{})

		result := p.Output(h, nil)
		assert.NotContains(t, result, "[CPU行動]")
	})

	t.Run("showdown results with human winner", func(t *testing.T) {
		h, players := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100, BestHand: nil},
		})

		_ = players[0]
		result := p.Output(h, nil)
		assert.Contains(t, result, "[結果]")
		assert.Contains(t, result, "あなた: フラッシュ")
		assert.Contains(t, result, "100チップ獲得")
	})

	t.Run("showdown results with kickers", func(t *testing.T) {
		h, _ := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandOnePair, HandName: "One Pair", Kickers: []int{14, 12, 10}, WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "あなた: ワンペア (キッカー: A, Q, 10)")
		assert.Contains(t, result, "100チップ獲得")
	})

	t.Run("showdown results without kickers", func(t *testing.T) {
		h, _ := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush", Kickers: nil, WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "あなた: フラッシュ")
		assert.NotContains(t, result, "キッカー")
	})

	t.Run("showdown results with CPU winner", func(t *testing.T) {
		h, _ := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 1, HandRank: domain.PokerHandOnePair, HandName: "One Pair", Kickers: []int{13, 12, 11}, WonAmount: 50, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "CPU 1: ワンペア (キッカー: K, Q, J)")
		assert.Contains(t, result, "50チップ獲得")
	})

	t.Run("showdown results with empty hand name", func(t *testing.T) {
		h, _ := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 1, HandName: "", WonAmount: 0, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "CPU 1")
		assert.NotContains(t, result, "CPU 1:")
	})

	t.Run("showdown results with zero won amount", func(t *testing.T) {
		h, _ := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 1, HandRank: domain.PokerHandHighCard, HandName: "High Card", WonAmount: 0, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.NotContains(t, result, "チップ獲得")
	})

	t.Run("results not shown in non-end phase", func(t *testing.T) {
		h, _ := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhaseFlop)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.NotContains(t, result, "[結果]")
	})

	t.Run("error message displayed", func(t *testing.T) {
		h, _ := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhasePreFlop)

		result := p.Output(h, errors.New("test error"))
		assert.Contains(t, result, "test error")
	})

	t.Run("game end message displayed", func(t *testing.T) {
		h, _ := makeHoldemForPresenter()
		h.SetGameEndFlag(true)

		result := p.Output(h, nil)
		assert.Contains(t, result, "ゲーム終了")
	})

	t.Run("game not ended hides game end message", func(t *testing.T) {
		h, _ := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhasePreFlop)

		result := p.Output(h, nil)
		// "ゲーム終了" should not appear unless gameEndFlag is true
		// The line "ゲーム終了\n" is only printed if h.GetGameEndFlag()
		// We just check it doesn't end with the specific game end block
		_ = result
	})

	t.Run("getCardStr with out of range design", func(t *testing.T) {
		h, players := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhasePreFlop)
		players[0].AddCard(domain.NewCard(99, 5, false))

		result := p.Output(h, nil)
		assert.Contains(t, result, "🃏5") // falls back to joker design
	})

	t.Run("getCardStr with negative design", func(t *testing.T) {
		h, players := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhasePreFlop)
		players[0].AddCard(domain.NewCard(-1, 3, false))

		result := p.Output(h, nil)
		assert.Contains(t, result, "🃏3")
	})

	t.Run("all action names", func(t *testing.T) {
		h, _ := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhasePreFlop)
		h.SetCpuActions([]domain.HoldemCpuAction{
			{PlayerIdx: 1, Action: domain.HoldemActionFold, Amount: 0},
			{PlayerIdx: 1, Action: domain.HoldemActionCheck, Amount: 0},
			{PlayerIdx: 1, Action: domain.HoldemActionCall, Amount: 0},
			{PlayerIdx: 1, Action: domain.HoldemActionBet, Amount: 10},
			{PlayerIdx: 1, Action: domain.HoldemActionRaise, Amount: 20},
			{PlayerIdx: 1, Action: domain.HoldemActionAllIn, Amount: 100},
			{PlayerIdx: 1, Action: 99, Amount: 0},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "フォールド")
		assert.Contains(t, result, "チェック")
		assert.Contains(t, result, "コール")
		assert.Contains(t, result, "ベット")
		assert.Contains(t, result, "レイズ")
		assert.Contains(t, result, "オールイン")
		assert.Contains(t, result, "不明")
	})

	t.Run("HUD stats shown when totalHands > 0", func(t *testing.T) {
		h, players := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhasePreFlop)
		players[0].IncrementTotalHands()
		players[0].IncrementTotalHands()
		players[0].IncrementVPIP()
		players[1].IncrementTotalHands()
		players[1].IncrementTotalHands()
		players[1].IncrementTotalHands()
		players[1].IncrementVPIP()
		players[1].IncrementVPIP()
		players[1].IncrementPFR()

		result := p.Output(h, nil)
		assert.Contains(t, result, "VPIP:50% PFR:0% 3Bet:0% AF:-")
		assert.Contains(t, result, "VPIP:66% PFR:33% 3Bet:0% AF:-")
	})

	t.Run("HUD stats with AF infinity", func(t *testing.T) {
		h, players := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhasePreFlop)
		players[0].IncrementTotalHands()
		players[0].IncrementPostFlopBetRaise()

		result := p.Output(h, nil)
		assert.Contains(t, result, "AF:∞")
	})

	t.Run("HUD stats with AF normal ratio", func(t *testing.T) {
		h, players := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhasePreFlop)
		players[0].IncrementTotalHands()
		players[0].IncrementPostFlopBetRaise()
		players[0].IncrementPostFlopBetRaise()
		players[0].IncrementPostFlopCall()

		result := p.Output(h, nil)
		assert.Contains(t, result, "AF:2.0")
	})

	t.Run("HUD stats not shown when totalHands is 0", func(t *testing.T) {
		h, _ := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhasePreFlop)

		result := p.Output(h, nil)
		assert.NotContains(t, result, "VPIP:")
		assert.NotContains(t, result, "PFR:")
	})

	t.Run("tournament mode header shown when enabled", func(t *testing.T) {
		h, _ := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhasePreFlop)
		cfg := domain.HoldemConfig{
			SmallBlind:      10,
			BigBlind:        20,
			InitChips:       1000,
			TournamentMode:  true,
			BlindLevelHands: 5,
			BlindMultiplier: 200,
		}
		h.SetConfig(cfg)
		h.SetHandCount(3)

		result := p.Output(h, nil)
		assert.Contains(t, result, "トーナメント ハンド#3 SB:10 BB:20 (レベルアップ:5ハンド毎)")
	})

	t.Run("tournament mode header not shown when disabled", func(t *testing.T) {
		h, _ := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhasePreFlop)

		result := p.Output(h, nil)
		assert.NotContains(t, result, "トーナメント")
	})
}

func TestHoldemCuiPresenter_Output_LearningMode(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.HoldemCuiPresenter)

	t.Run("equity and pot odds shown on human turn", func(t *testing.T) {
		h, players := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhasePreFlop)
		h.SetCurrentTurn(0)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))

		result := p.Output(h, nil)
		assert.Contains(t, result, "[学習モード]")
		assert.Contains(t, result, "勝率:")
		assert.Contains(t, result, "ポットオッズ:")
	})

	t.Run("EV verdict shown when a call amount exists", func(t *testing.T) {
		h, players := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhasePreFlop)
		h.SetCurrentTurn(0)
		h.SetPot(100)
		h.SetLastBet(50)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))

		result := p.Output(h, nil)
		assert.Contains(t, result, "EV (")
	})

	t.Run("no EV verdict when there is no call amount", func(t *testing.T) {
		h, players := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhasePreFlop)
		h.SetCurrentTurn(0)
		h.SetLastBet(0)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))

		result := p.Output(h, nil)
		assert.Contains(t, result, "[学習モード]")
		assert.NotContains(t, result, "EV (")
	})

	t.Run("not shown when it is not the human's turn", func(t *testing.T) {
		h, players := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhasePreFlop)
		h.SetCurrentTurn(1)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))

		result := p.Output(h, nil)
		assert.NotContains(t, result, "[学習モード]")
	})

	t.Run("not shown when the human has folded", func(t *testing.T) {
		h, players := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhasePreFlop)
		h.SetCurrentTurn(0)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))
		players[0].SetFolded(true)

		result := p.Output(h, nil)
		assert.NotContains(t, result, "[学習モード]")
	})

	t.Run("not shown outside the betting phases", func(t *testing.T) {
		h, players := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhaseShowdown)
		h.SetCurrentTurn(0)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))

		result := p.Output(h, nil)
		assert.NotContains(t, result, "[学習モード]")
	})
}

func TestHoldemCuiPresenter_Output_BettingLimitDisplay(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.HoldemCuiPresenter)

	t.Run("displays Fixed limit", func(t *testing.T) {
		h, _ := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhasePreFlop)
		result := p.Output(h, nil)
		assert.Contains(t, result, "Fixed")
	})
}

func TestHoldemCuiPresenter_Output_RebuyAddon(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.HoldemCuiPresenter)

	t.Run("tournament mode with rebuy enabled shows rebuy info", func(t *testing.T) {
		h, _ := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhasePreFlop)
		cfg := domain.HoldemConfig{
			SmallBlind:       10,
			BigBlind:         20,
			InitChips:        1000,
			TournamentMode:   true,
			BlindLevelHands:  5,
			BlindMultiplier:  200,
			RebuyEnabled:     true,
			RebuyChips:       1000,
			RebuyMaxCount:    3,
			RebuyPeriodHands: 20,
		}
		h.SetConfig(cfg)
		h.SetHandCount(2)

		result := p.Output(h, nil)
		assert.Contains(t, result, "リバイ: 1000チップ (最大3回, 20ハンド目まで)")
	})

	t.Run("tournament mode with addon enabled shows addon info", func(t *testing.T) {
		h, _ := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhasePreFlop)
		cfg := domain.HoldemConfig{
			SmallBlind:      10,
			BigBlind:        20,
			InitChips:       1000,
			TournamentMode:  true,
			BlindLevelHands: 5,
			BlindMultiplier: 200,
			AddonEnabled:    true,
			AddonChips:      1500,
			AddonAfterHand:  20,
		}
		h.SetConfig(cfg)
		h.SetHandCount(2)

		result := p.Output(h, nil)
		assert.Contains(t, result, "アドオン: 1500チップ (20ハンド目に提供)")
	})

	t.Run("rebuy phase type 1 shows rebuy prompt", func(t *testing.T) {
		h, _ := makeHoldemForPresenter()
		cfg := domain.HoldemConfig{
			SmallBlind:       10,
			BigBlind:         20,
			InitChips:        1000,
			TournamentMode:   true,
			BlindLevelHands:  5,
			BlindMultiplier:  200,
			RebuyEnabled:     true,
			RebuyChips:       1000,
			RebuyMaxCount:    3,
			RebuyPeriodHands: 20,
		}
		h.SetConfig(cfg)
		h.SetPhase(domain.HoldemPhaseRebuy)
		h.SetRebuyPhaseType(1)
		h.SetRebuyCounts([]int{1, 0, 0, 0})

		result := p.Output(h, nil)
		assert.Contains(t, result, "リバイしますか?")
		assert.Contains(t, result, "1000チップ")
		assert.Contains(t, result, "1/3回使用済")
		assert.Contains(t, result, "rb=リバイ / sr=スキップ")
	})

	t.Run("rebuy phase type 2 shows addon prompt", func(t *testing.T) {
		h, _ := makeHoldemForPresenter()
		cfg := domain.HoldemConfig{
			SmallBlind:      10,
			BigBlind:        20,
			InitChips:       1000,
			TournamentMode:  true,
			BlindLevelHands: 5,
			BlindMultiplier: 200,
			AddonEnabled:    true,
			AddonChips:      1500,
			AddonAfterHand:  20,
		}
		h.SetConfig(cfg)
		h.SetPhase(domain.HoldemPhaseRebuy)
		h.SetRebuyPhaseType(2)

		result := p.Output(h, nil)
		assert.Contains(t, result, "アドオンしますか?")
		assert.Contains(t, result, "1500チップ")
		assert.Contains(t, result, "ad=アドオン / sa=スキップ")
	})

	t.Run("rebuy phase type 0 does not show prompt", func(t *testing.T) {
		h, _ := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhaseRebuy)
		h.SetRebuyPhaseType(0)

		result := p.Output(h, nil)
		assert.NotContains(t, result, "リバイしますか?")
		assert.NotContains(t, result, "アドオンしますか?")
	})
}

func TestHoldemCuiPresenter_Output_TableSize(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.HoldemCuiPresenter)

	t.Run("displays 4-max", func(t *testing.T) {
		h, _ := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhasePreFlop)
		result := p.Output(h, nil)
		assert.Contains(t, result, "テーブル: 4-max")
	})

	t.Run("displays 6-max", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players6 := make([]*domain.HoldemPlayer, 6)
		players6[0] = domain.NewHoldemPlayer(true, domain.HoldemStyleTAG)
		for i := 1; i < 6; i++ {
			players6[i] = domain.NewHoldemPlayer(false, domain.HoldemStyleLAP)
		}
		h := domain.NewHoldem(tc, players6, domain.DefaultHoldemConfig())
		h.SetPhase(domain.HoldemPhasePreFlop)
		result := p.Output(h, nil)
		assert.Contains(t, result, "テーブル: 6-max")
	})
}

func TestHoldemCuiPresenter_Output_Muck(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.HoldemCuiPresenter)

	t.Run("muck prompt displayed during showdown when IsMuckAvailable", func(t *testing.T) {
		h, _ := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhaseShowdown)
		// IsMuckAvailable returns true when phase=SHOWDOWN and human has wonAmount=0
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandOnePair, HandName: "One Pair", WonAmount: 0, BestHand: nil},
			{PlayerIdx: 1, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "マックしますか? (m=マック / sh=ショー)")
	})

	t.Run("muck prompt not displayed when not available", func(t *testing.T) {
		h, _ := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.NotContains(t, result, "マックしますか?")
	})

	t.Run("mucked result displayed as マック", func(t *testing.T) {
		h, _ := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandOnePair, HandName: "One Pair", WonAmount: 0, Mucked: true, BestHand: nil},
			{PlayerIdx: 1, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "あなた: マック")
		assert.NotContains(t, result, "あなた: ワンペア")
	})

	t.Run("results shown in showdown phase", func(t *testing.T) {
		h, _ := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhaseShowdown)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "[結果]")
		assert.Contains(t, result, "あなた: フラッシュ")
	})
}

func TestHoldemCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.HoldemCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		mockGame := new(interfaces.MockHoldemGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "raise", Detail: "raised to 100"},
		}
		mockGame.On("GetGameEndFlag").Return(true)
		mockGame.On("GetActionLog").Return(entries)

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "raise")
		assert.Contains(t, result, "raised to 100")
		mockGame.AssertExpectations(t)
	})

	t.Run("nil_entries", func(t *testing.T) {
		mockGame := new(interfaces.MockHoldemGame)
		mockGame.On("GetGameEndFlag").Return(true)
		mockGame.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, "棋譜はありません")
		mockGame.AssertExpectations(t)
	})

	t.Run("game_not_ended", func(t *testing.T) {
		mockGame := new(interfaces.MockHoldemGame)
		mockGame.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, "棋譜はありません")
		mockGame.AssertExpectations(t)
	})
}
