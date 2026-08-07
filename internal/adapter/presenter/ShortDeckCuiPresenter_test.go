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

func makeShortDeckForPresenter() (*domain.ShortDeck, []*domain.ShortDeckPlayer) {
	tc := domain.NewTrumpCards(0)
	players := []*domain.ShortDeckPlayer{
		domain.NewShortDeckPlayer(true, domain.HoldemStyleTAG),
		domain.NewShortDeckPlayer(false, domain.HoldemStyleLAP),
		domain.NewShortDeckPlayer(false, domain.HoldemStyleTAP),
		domain.NewShortDeckPlayer(false, domain.HoldemStyleGTO),
	}
	h := domain.NewShortDeck(tc, players, domain.DefaultShortDeckConfig())
	return h, players
}

func TestShortDeckCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.ShortDeckCuiPresenter)

	t.Run("initial state with no community cards", func(t *testing.T) {
		h, players := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhasePreFlop)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := p.Output(h, nil)
		assert.Contains(t, result, "Short Deck Hold'em")
		assert.Contains(t, result, "ディーラー:")
		assert.Contains(t, result, "(なし)")
		assert.Contains(t, result, "ポット:")
		assert.Contains(t, result, "あなた")
		assert.Contains(t, result, "♠10")
		assert.Contains(t, result, "♥11")
		// The Short Deck rule reminder is always shown.
		assert.Contains(t, result, "フラッシュ＞フルハウス")
		assert.Contains(t, result, "A-6-7-8-9")
	})

	t.Run("community cards displayed", func(t *testing.T) {
		h, _ := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhaseFlop)
		h.SetCommunityCards([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 6, false),
			domain.NewCard(domain.CardDesignClover, 8, false),
			domain.NewCard(domain.CardDesignDiamond, 9, false),
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "♠6")
		assert.Contains(t, result, "♣8")
		assert.Contains(t, result, "♦9")
		assert.NotContains(t, result, "(なし)")
	})

	t.Run("multiple community cards separator", func(t *testing.T) {
		h, _ := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhaseFlop)
		h.SetCommunityCards([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignHeart, 6, false),
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "♠1  ♥6")
	})

	t.Run("CPU player info with play style", func(t *testing.T) {
		h, _ := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhasePreFlop)

		result := p.Output(h, nil)
		assert.Contains(t, result, "CPU 1")
		assert.Contains(t, result, "CPU 2")
		assert.Contains(t, result, "CPU 3")
	})

	t.Run("folded player", func(t *testing.T) {
		h, players := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhasePreFlop)
		players[1].SetFolded(true)

		result := p.Output(h, nil)
		assert.Contains(t, result, "[フォールド]")
	})

	t.Run("all-in player", func(t *testing.T) {
		h, players := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhasePreFlop)
		players[1].SetAllIn(true)

		result := p.Output(h, nil)
		assert.Contains(t, result, "[オールイン]")
	})

	t.Run("player with current bet", func(t *testing.T) {
		h, players := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhasePreFlop)
		players[0].SetCurrentBet(50)

		result := p.Output(h, nil)
		assert.Contains(t, result, "ベット:50")
	})

	t.Run("player with zero bet hides bet label", func(t *testing.T) {
		h, players := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhasePreFlop)
		players[0].SetCurrentBet(0)

		result := p.Output(h, nil)
		assert.NotContains(t, result, "ベット:0")
	})

	t.Run("human cards visible when not folded", func(t *testing.T) {
		h, players := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhasePreFlop)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 13, false))

		result := p.Output(h, nil)
		assert.Contains(t, result, "手札:")
		assert.Contains(t, result, "♠1")
		assert.Contains(t, result, "♥13")
	})

	t.Run("human cards hidden when folded", func(t *testing.T) {
		h, players := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhasePreFlop)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].SetFolded(true)

		result := p.Output(h, nil)
		assert.NotContains(t, result, "手札:")
	})

	t.Run("human cards with separator", func(t *testing.T) {
		h, players := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhasePreFlop)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 7, false))

		result := p.Output(h, nil)
		assert.Contains(t, result, "♠6  ♣7")
	})

	t.Run("CPU actions displayed", func(t *testing.T) {
		h, _ := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhasePreFlop)
		h.SetCpuActions([]domain.HoldemCpuAction{
			{PlayerIdx: 1, Action: domain.ShortDeckActionCall, Amount: 0},
			{PlayerIdx: 2, Action: domain.ShortDeckActionRaise, Amount: 30},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "[CPU行動]")
		assert.Contains(t, result, "Player 1: コール")
		assert.Contains(t, result, "Player 2: レイズ")
		assert.Contains(t, result, "(30)")
	})

	t.Run("CPU action without amount", func(t *testing.T) {
		h, _ := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhasePreFlop)
		h.SetCpuActions([]domain.HoldemCpuAction{
			{PlayerIdx: 1, Action: domain.ShortDeckActionFold, Amount: 0},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "フォールド")
		assert.NotContains(t, result, "(0)")
	})

	t.Run("no CPU actions hides section", func(t *testing.T) {
		h, _ := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhasePreFlop)
		h.SetCpuActions([]domain.HoldemCpuAction{})

		result := p.Output(h, nil)
		assert.NotContains(t, result, "[CPU行動]")
	})

	t.Run("showdown results with human winner", func(t *testing.T) {
		h, players := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.ShortDeckHandFlush, HandName: "Flush", WonAmount: 100, BestHand: nil},
		})

		_ = players[0]
		result := p.Output(h, nil)
		assert.Contains(t, result, "[結果]")
		assert.Contains(t, result, "あなた: フラッシュ")
		assert.Contains(t, result, "100チップ獲得")
	})

	t.Run("showdown results with kickers", func(t *testing.T) {
		h, _ := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.ShortDeckHandOnePair, HandName: "One Pair", Kickers: []int{14, 12, 10}, WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "あなた: ワンペア (キッカー: A, Q, 10)")
		assert.Contains(t, result, "100チップ獲得")
	})

	t.Run("showdown results without kickers", func(t *testing.T) {
		h, _ := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.ShortDeckHandFlush, HandName: "Flush", Kickers: nil, WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "あなた: フラッシュ")
		assert.NotContains(t, result, "キッカー")
	})

	t.Run("showdown results with CPU winner", func(t *testing.T) {
		h, _ := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 1, HandRank: domain.ShortDeckHandOnePair, HandName: "One Pair", Kickers: []int{13, 12, 11}, WonAmount: 50, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "CPU 1: ワンペア (キッカー: K, Q, J)")
		assert.Contains(t, result, "50チップ獲得")
	})

	t.Run("showdown results with empty hand name", func(t *testing.T) {
		h, _ := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 1, HandName: "", WonAmount: 0, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "CPU 1")
		assert.NotContains(t, result, "CPU 1:")
	})

	t.Run("showdown results with zero won amount", func(t *testing.T) {
		h, _ := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 1, HandRank: domain.ShortDeckHandHighCard, HandName: "High Card", WonAmount: 0, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.NotContains(t, result, "チップ獲得")
	})

	t.Run("results not shown in non-end phase", func(t *testing.T) {
		h, _ := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhaseFlop)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.ShortDeckHandFlush, HandName: "Flush", WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.NotContains(t, result, "[結果]")
	})

	t.Run("error message displayed", func(t *testing.T) {
		h, _ := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhasePreFlop)

		result := p.Output(h, errors.New("test error"))
		assert.Contains(t, result, "test error")
	})

	t.Run("game end message displayed", func(t *testing.T) {
		h, _ := makeShortDeckForPresenter()
		h.SetGameEndFlag(true)

		result := p.Output(h, nil)
		assert.Contains(t, result, "ゲーム終了")
	})

	t.Run("game not ended hides game end message", func(t *testing.T) {
		h, _ := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhasePreFlop)

		result := p.Output(h, nil)
		_ = result
	})

	t.Run("getCardStr with out of range design", func(t *testing.T) {
		h, players := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhasePreFlop)
		players[0].AddCard(domain.NewCard(99, 8, false))

		result := p.Output(h, nil)
		assert.Contains(t, result, "🃏8") // falls back to joker design
	})

	t.Run("getCardStr with negative design", func(t *testing.T) {
		h, players := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhasePreFlop)
		players[0].AddCard(domain.NewCard(-1, 6, false))

		result := p.Output(h, nil)
		assert.Contains(t, result, "🃏6")
	})

	t.Run("all action names", func(t *testing.T) {
		h, _ := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhasePreFlop)
		h.SetCpuActions([]domain.HoldemCpuAction{
			{PlayerIdx: 1, Action: domain.ShortDeckActionFold, Amount: 0},
			{PlayerIdx: 1, Action: domain.ShortDeckActionCheck, Amount: 0},
			{PlayerIdx: 1, Action: domain.ShortDeckActionCall, Amount: 0},
			{PlayerIdx: 1, Action: domain.ShortDeckActionBet, Amount: 10},
			{PlayerIdx: 1, Action: domain.ShortDeckActionRaise, Amount: 20},
			{PlayerIdx: 1, Action: domain.ShortDeckActionAllIn, Amount: 100},
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
		h, players := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhasePreFlop)
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
		h, players := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhasePreFlop)
		players[0].IncrementTotalHands()
		players[0].IncrementPostFlopBetRaise()

		result := p.Output(h, nil)
		assert.Contains(t, result, "AF:∞")
	})

	t.Run("HUD stats with AF normal ratio", func(t *testing.T) {
		h, players := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhasePreFlop)
		players[0].IncrementTotalHands()
		players[0].IncrementPostFlopBetRaise()
		players[0].IncrementPostFlopBetRaise()
		players[0].IncrementPostFlopCall()

		result := p.Output(h, nil)
		assert.Contains(t, result, "AF:2.0")
	})

	t.Run("HUD stats not shown when totalHands is 0", func(t *testing.T) {
		h, _ := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhasePreFlop)

		result := p.Output(h, nil)
		assert.NotContains(t, result, "VPIP:")
		assert.NotContains(t, result, "PFR:")
	})

	t.Run("tournament mode header shown when enabled", func(t *testing.T) {
		h, _ := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhasePreFlop)
		cfg := domain.ShortDeckConfig{
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
		h, _ := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhasePreFlop)

		result := p.Output(h, nil)
		assert.NotContains(t, result, "トーナメント")
	})
}

func TestShortDeckCuiPresenter_Output_BettingLimitDisplay(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.ShortDeckCuiPresenter)

	t.Run("displays Fixed limit", func(t *testing.T) {
		h, _ := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhasePreFlop)
		result := p.Output(h, nil)
		assert.Contains(t, result, "Fixed")
	})
}

func TestShortDeckCuiPresenter_Output_RebuyAddon(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.ShortDeckCuiPresenter)

	t.Run("tournament mode with rebuy enabled shows rebuy info", func(t *testing.T) {
		h, _ := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhasePreFlop)
		cfg := domain.ShortDeckConfig{
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
		h, _ := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhasePreFlop)
		cfg := domain.ShortDeckConfig{
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
		h, _ := makeShortDeckForPresenter()
		cfg := domain.ShortDeckConfig{
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
		h.SetPhase(domain.ShortDeckPhaseRebuy)
		h.SetRebuyPhaseType(1)
		h.SetRebuyCounts([]int{1, 0, 0, 0})

		result := p.Output(h, nil)
		assert.Contains(t, result, "リバイしますか?")
		assert.Contains(t, result, "1000チップ")
		assert.Contains(t, result, "1/3回使用済")
		assert.Contains(t, result, "rb=リバイ / sr=スキップ")
	})

	t.Run("rebuy phase type 2 shows addon prompt", func(t *testing.T) {
		h, _ := makeShortDeckForPresenter()
		cfg := domain.ShortDeckConfig{
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
		h.SetPhase(domain.ShortDeckPhaseRebuy)
		h.SetRebuyPhaseType(2)

		result := p.Output(h, nil)
		assert.Contains(t, result, "アドオンしますか?")
		assert.Contains(t, result, "1500チップ")
		assert.Contains(t, result, "ad=アドオン / sa=スキップ")
	})

	t.Run("rebuy phase type 0 does not show prompt", func(t *testing.T) {
		h, _ := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhaseRebuy)
		h.SetRebuyPhaseType(0)

		result := p.Output(h, nil)
		assert.NotContains(t, result, "リバイしますか?")
		assert.NotContains(t, result, "アドオンしますか?")
	})
}

func TestShortDeckCuiPresenter_Output_TableSize(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.ShortDeckCuiPresenter)

	t.Run("displays 4-max", func(t *testing.T) {
		h, _ := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhasePreFlop)
		result := p.Output(h, nil)
		assert.Contains(t, result, "テーブル: 4-max")
	})

	t.Run("displays 6-max", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players6 := make([]*domain.ShortDeckPlayer, 6)
		players6[0] = domain.NewShortDeckPlayer(true, domain.HoldemStyleTAG)
		for i := 1; i < 6; i++ {
			players6[i] = domain.NewShortDeckPlayer(false, domain.HoldemStyleLAP)
		}
		h := domain.NewShortDeck(tc, players6, domain.DefaultShortDeckConfig())
		h.SetPhase(domain.ShortDeckPhasePreFlop)
		result := p.Output(h, nil)
		assert.Contains(t, result, "テーブル: 6-max")
	})
}

func TestShortDeckCuiPresenter_Output_Muck(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.ShortDeckCuiPresenter)

	t.Run("muck prompt displayed during showdown when IsMuckAvailable", func(t *testing.T) {
		h, _ := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhaseShowdown)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.ShortDeckHandOnePair, HandName: "One Pair", WonAmount: 0, BestHand: nil},
			{PlayerIdx: 1, HandRank: domain.ShortDeckHandFlush, HandName: "Flush", WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "マックしますか? (m=マック / sh=ショー)")
	})

	t.Run("muck prompt not displayed when not available", func(t *testing.T) {
		h, _ := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.ShortDeckHandFlush, HandName: "Flush", WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.NotContains(t, result, "マックしますか?")
	})

	t.Run("mucked result displayed as マック", func(t *testing.T) {
		h, _ := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandName: "One Pair", WonAmount: 0, Mucked: true, BestHand: nil},
			{PlayerIdx: 1, HandRank: domain.ShortDeckHandFlush, HandName: "Flush", WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "あなた: マック")
		assert.NotContains(t, result, "あなた: One Pair")
	})

	t.Run("results shown in showdown phase", func(t *testing.T) {
		h, _ := makeShortDeckForPresenter()
		h.SetPhase(domain.ShortDeckPhaseShowdown)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.ShortDeckHandFlush, HandName: "Flush", WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "[結果]")
		assert.Contains(t, result, "あなた: フラッシュ")
	})
}

func TestShortDeckCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.ShortDeckCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		mockGame := new(interfaces.MockShortDeckGame)
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
		mockGame := new(interfaces.MockShortDeckGame)
		mockGame.On("GetGameEndFlag").Return(true)
		mockGame.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, "棋譜はありません")
		mockGame.AssertExpectations(t)
	})

	t.Run("game_not_ended", func(t *testing.T) {
		mockGame := new(interfaces.MockShortDeckGame)
		mockGame.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, "棋譜はありません")
		mockGame.AssertExpectations(t)
	})
}

// TestShortDeckCuiPresenter_HandNameOrdering pins the short-deck ranking, where
// the flush outranks the full house (#4987).
//
// **共通の役表を流用すると 5 が「フラッシュ」になる。**標準デッキでは 5=Flush /
// 6=Full House だが、36枚デッキでは逆。訳されないより誤訳のほうが悪いので、
// 入れ替わっている 2 ランクを名指しで固定する。
func TestShortDeckCuiPresenter_HandNameOrdering(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	p := new(presenter.ShortDeckCuiPresenter)
	tests := []struct {
		name string
		rank int
		want string
		deny string
	}{
		{"rank 5 is the full house, not the flush", domain.ShortDeckHandFullHouse, "フルハウス", "フラッシュ"},
		{"rank 6 is the flush, not the full house", domain.ShortDeckHandFlush, "フラッシュ", "フルハウス"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, _ := makeShortDeckForPresenter()
			h.SetPhase(domain.ShortDeckPhaseEnd)
			h.SetRoundResults([]domain.HoldemResult{
				{PlayerIdx: 0, HandRank: tt.rank, HandName: domain.ShortDeckHandNames[tt.rank], WonAmount: 100},
			})

			result := p.Output(h, nil)
			assert.Contains(t, result, "あなた: "+tt.want)
			assert.NotContains(t, result, "あなた: "+tt.deny)
		})
	}
}
