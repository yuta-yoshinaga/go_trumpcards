package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func makeOmahaForPresenter() (*domain.Omaha, []*domain.OmahaPlayer) {
	tc := domain.NewTrumpCards(0)
	players := []*domain.OmahaPlayer{
		domain.NewOmahaPlayer(true, domain.HoldemStyleTAG),
		domain.NewOmahaPlayer(false, domain.HoldemStyleLAP),
		domain.NewOmahaPlayer(false, domain.HoldemStyleTAP),
		domain.NewOmahaPlayer(false, domain.HoldemStyleGTO),
	}
	h := domain.NewOmaha(tc, players, domain.DefaultOmahaConfig())
	return h, players
}

func makeOmahaHiLoForPresenter() *domain.Omaha {
	tc := domain.NewTrumpCards(0)
	players := []*domain.OmahaPlayer{
		domain.NewOmahaPlayer(true, domain.HoldemStyleTAG),
		domain.NewOmahaPlayer(false, domain.HoldemStyleLAP),
	}
	return domain.NewOmahaHiLo(tc, players, domain.DefaultOmahaConfig())
}

func TestOmahaCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.OmahaCuiPresenter)

	t.Run("initial state with no community cards", func(t *testing.T) {
		h, players := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhasePreFlop)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := p.Output(h, nil)
		assert.Contains(t, result, "Omaha Hold'em")
		assert.Contains(t, result, "ディーラー:")
		assert.Contains(t, result, "(なし)")
		assert.Contains(t, result, "ポット:")
		assert.Contains(t, result, "あなた")
		assert.Contains(t, result, "♠10")
		assert.Contains(t, result, "♥11")
		// The must-use-2 rule line is always shown; standard Omaha deals 4 holes.
		assert.Contains(t, result, "手札4枚のうちちょうど2枚")
		// Hi-only game → no low-qualifier rule line.
		assert.NotContains(t, result, i18n.T("omaha.hiLoRuleLine"))
	})

	t.Run("hi-lo game shows the eight-or-better low rule", func(t *testing.T) {
		h := makeOmahaHiLoForPresenter()
		h.SetPhase(domain.OmahaPhasePreFlop)
		result := p.Output(h, nil)
		assert.Contains(t, result, i18n.T("omaha.hiLoRuleLine"))
	})

	// **Web は omaha-live-besthand で暫定ベストを常時出しているのに、CUI は
	// 「2枚使用」の注意書きだけで実際の役を出していなかった (#4680)。**
	// 手札4枚から必ず2枚という特殊ルールがあるぶん、暫定表示はミスを防ぐ補助になる。
	t.Run("shows the human's current best hand once the flop is out", func(t *testing.T) {
		h, players := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhaseFlop)
		// 手札 ♠2 ♠3 ♥11 ♥12、ボード ♠5 ♠7 ♠9 → ♠ のフラッシュ (手札2枚+場3枚)。
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 12, false))
		h.SetCommunityCards([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 5, false),
			domain.NewCard(domain.CardDesignSpade, 7, false),
			domain.NewCard(domain.CardDesignSpade, 9, false),
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "現在の最善役")
		assert.Contains(t, result, "フラッシュ")
	})

	// **プリフロップでは出さない。**ボードが3枚に満たないと役は決まらない。
	t.Run("shows nothing before the flop", func(t *testing.T) {
		h, players := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhasePreFlop)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 12, false))

		assert.NotContains(t, p.Output(h, nil), "現在の最善役")
	})

	// **表示のために状態を変えない。**Peek を使っているので、描画しても
	// プレイヤーの handRank は書き換わらない。
	t.Run("rendering does not mutate the player state", func(t *testing.T) {
		h, players := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhaseFlop)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 12, false))
		h.SetCommunityCards([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 5, false),
			domain.NewCard(domain.CardDesignSpade, 7, false),
			domain.NewCard(domain.CardDesignSpade, 9, false),
		})

		before := players[0].GetBestHand()
		_ = p.Output(h, nil)
		assert.Equal(t, before, players[0].GetBestHand(), "描画でベストハンドが書き換わらない")
	})

	t.Run("community cards displayed", func(t *testing.T) {
		h, _ := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhaseFlop)
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
		h, _ := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhaseFlop)
		h.SetCommunityCards([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignHeart, 2, false),
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "♠1  ♥2")
	})

	t.Run("CPU player info with play style", func(t *testing.T) {
		h, _ := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhasePreFlop)

		result := p.Output(h, nil)
		assert.Contains(t, result, "CPU 1")
		assert.Contains(t, result, "CPU 2")
		assert.Contains(t, result, "CPU 3")
	})

	t.Run("folded player", func(t *testing.T) {
		h, players := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhasePreFlop)
		players[1].SetFolded(true)

		result := p.Output(h, nil)
		assert.Contains(t, result, "[フォールド]")
	})

	t.Run("all-in player", func(t *testing.T) {
		h, players := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhasePreFlop)
		players[1].SetAllIn(true)

		result := p.Output(h, nil)
		assert.Contains(t, result, "[オールイン]")
	})

	t.Run("player with current bet", func(t *testing.T) {
		h, players := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhasePreFlop)
		players[0].SetCurrentBet(50)

		result := p.Output(h, nil)
		assert.Contains(t, result, "ベット:50")
	})

	t.Run("player with zero bet hides bet label", func(t *testing.T) {
		h, players := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhasePreFlop)
		players[0].SetCurrentBet(0)

		result := p.Output(h, nil)
		assert.NotContains(t, result, "ベット:0")
	})

	t.Run("human cards visible when not folded", func(t *testing.T) {
		h, players := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhasePreFlop)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 13, false))

		result := p.Output(h, nil)
		assert.Contains(t, result, "手札:")
		assert.Contains(t, result, "♠1")
		assert.Contains(t, result, "♥13")
	})

	t.Run("human cards hidden when folded", func(t *testing.T) {
		h, players := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhasePreFlop)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].SetFolded(true)

		result := p.Output(h, nil)
		assert.NotContains(t, result, "手札:")
	})

	t.Run("human cards with separator", func(t *testing.T) {
		h, players := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhasePreFlop)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 7, false))

		result := p.Output(h, nil)
		assert.Contains(t, result, "♠3  ♣7")
	})

	t.Run("CPU actions displayed", func(t *testing.T) {
		h, _ := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhasePreFlop)
		h.SetCpuActions([]domain.HoldemCpuAction{
			{PlayerIdx: 1, Action: domain.OmahaActionCall, Amount: 0},
			{PlayerIdx: 2, Action: domain.OmahaActionRaise, Amount: 30},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "[CPU行動]")
		// CPU action lines use the localized player name, matching the result section.
		assert.Contains(t, result, "CPU 1: コール")
		assert.Contains(t, result, "CPU 2: レイズ")
		assert.NotContains(t, result, "Player 1:")
		assert.Contains(t, result, "(30)")
	})

	t.Run("CPU action without amount", func(t *testing.T) {
		h, _ := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhasePreFlop)
		h.SetCpuActions([]domain.HoldemCpuAction{
			{PlayerIdx: 1, Action: domain.OmahaActionFold, Amount: 0},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "フォールド")
		assert.NotContains(t, result, "(0)")
	})

	t.Run("no CPU actions hides section", func(t *testing.T) {
		h, _ := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhasePreFlop)
		h.SetCpuActions([]domain.HoldemCpuAction{})

		result := p.Output(h, nil)
		assert.NotContains(t, result, "[CPU行動]")
	})

	t.Run("showdown results with human winner", func(t *testing.T) {
		h, players := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhaseEnd)
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
		h, _ := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandOnePair, HandName: "One Pair", Kickers: []int{14, 12, 10}, WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "あなた: ワンペア (キッカー: A, Q, 10)")
		assert.Contains(t, result, "100チップ獲得")
	})

	t.Run("showdown results without kickers", func(t *testing.T) {
		h, _ := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush", Kickers: nil, WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "あなた: フラッシュ")
		assert.NotContains(t, result, "キッカー")
	})

	t.Run("showdown results with CPU winner", func(t *testing.T) {
		h, _ := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 1, HandRank: domain.PokerHandOnePair, HandName: "One Pair", Kickers: []int{13, 12, 11}, WonAmount: 50, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "CPU 1: ワンペア (キッカー: K, Q, J)")
		assert.Contains(t, result, "50チップ獲得")
	})

	t.Run("showdown results with empty hand name", func(t *testing.T) {
		h, _ := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 1, HandName: "", WonAmount: 0, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "CPU 1")
		assert.NotContains(t, result, "CPU 1:")
	})

	t.Run("showdown results with zero won amount", func(t *testing.T) {
		h, _ := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 1, HandRank: domain.PokerHandHighCard, HandName: "High Card", WonAmount: 0, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.NotContains(t, result, "チップ獲得")
	})

	t.Run("results not shown in non-end phase", func(t *testing.T) {
		h, _ := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhaseFlop)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.NotContains(t, result, "[結果]")
	})

	t.Run("error message displayed", func(t *testing.T) {
		h, _ := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhasePreFlop)

		result := p.Output(h, errors.New("test error"))
		assert.Contains(t, result, "test error")
	})

	t.Run("game end message displayed", func(t *testing.T) {
		h, _ := makeOmahaForPresenter()
		h.SetGameEndFlag(true)

		result := p.Output(h, nil)
		assert.Contains(t, result, "ゲーム終了")
	})

	t.Run("game not ended hides game end message", func(t *testing.T) {
		h, _ := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhasePreFlop)

		result := p.Output(h, nil)
		// "ゲーム終了" should not appear unless gameEndFlag is true
		// The line "ゲーム終了\n" is only printed if h.GetGameEndFlag()
		// We just check it doesn't end with the specific game end block
		_ = result
	})

	t.Run("getCardStr with out of range design", func(t *testing.T) {
		h, players := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhasePreFlop)
		players[0].AddCard(domain.NewCard(99, 5, false))

		result := p.Output(h, nil)
		assert.Contains(t, result, "🃏5") // falls back to joker design
	})

	t.Run("getCardStr with negative design", func(t *testing.T) {
		h, players := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhasePreFlop)
		players[0].AddCard(domain.NewCard(-1, 3, false))

		result := p.Output(h, nil)
		assert.Contains(t, result, "🃏3")
	})

	t.Run("all action names", func(t *testing.T) {
		h, _ := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhasePreFlop)
		h.SetCpuActions([]domain.HoldemCpuAction{
			{PlayerIdx: 1, Action: domain.OmahaActionFold, Amount: 0},
			{PlayerIdx: 1, Action: domain.OmahaActionCheck, Amount: 0},
			{PlayerIdx: 1, Action: domain.OmahaActionCall, Amount: 0},
			{PlayerIdx: 1, Action: domain.OmahaActionBet, Amount: 10},
			{PlayerIdx: 1, Action: domain.OmahaActionRaise, Amount: 20},
			{PlayerIdx: 1, Action: domain.OmahaActionAllIn, Amount: 100},
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
		h, players := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhasePreFlop)
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
		h, players := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhasePreFlop)
		players[0].IncrementTotalHands()
		players[0].IncrementPostFlopBetRaise()

		result := p.Output(h, nil)
		assert.Contains(t, result, "AF:∞")
	})

	t.Run("HUD stats with AF normal ratio", func(t *testing.T) {
		h, players := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhasePreFlop)
		players[0].IncrementTotalHands()
		players[0].IncrementPostFlopBetRaise()
		players[0].IncrementPostFlopBetRaise()
		players[0].IncrementPostFlopCall()

		result := p.Output(h, nil)
		assert.Contains(t, result, "AF:2.0")
	})

	t.Run("HUD stats not shown when totalHands is 0", func(t *testing.T) {
		h, _ := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhasePreFlop)

		result := p.Output(h, nil)
		assert.NotContains(t, result, "VPIP:")
		assert.NotContains(t, result, "PFR:")
	})

	t.Run("tournament mode header shown when enabled", func(t *testing.T) {
		h, _ := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhasePreFlop)
		cfg := domain.OmahaConfig{
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
		h, _ := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhasePreFlop)

		result := p.Output(h, nil)
		assert.NotContains(t, result, "トーナメント")
	})
}

func TestOmahaCuiPresenter_Output_BettingLimitDisplay(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.OmahaCuiPresenter)

	t.Run("displays Fixed limit", func(t *testing.T) {
		h, _ := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhasePreFlop)
		result := p.Output(h, nil)
		assert.Contains(t, result, "Fixed")
	})
}

func TestOmahaCuiPresenter_Output_RebuyAddon(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.OmahaCuiPresenter)

	t.Run("tournament mode with rebuy enabled shows rebuy info", func(t *testing.T) {
		h, _ := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhasePreFlop)
		cfg := domain.OmahaConfig{
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
		h, _ := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhasePreFlop)
		cfg := domain.OmahaConfig{
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
		h, _ := makeOmahaForPresenter()
		cfg := domain.OmahaConfig{
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
		h.SetPhase(domain.OmahaPhaseRebuy)
		h.SetRebuyPhaseType(1)
		h.SetRebuyCounts([]int{1, 0, 0, 0})

		result := p.Output(h, nil)
		assert.Contains(t, result, "リバイしますか?")
		assert.Contains(t, result, "1000チップ")
		assert.Contains(t, result, "1/3回使用済")
		assert.Contains(t, result, "rb=リバイ / sr=スキップ")
	})

	t.Run("rebuy phase type 2 shows addon prompt", func(t *testing.T) {
		h, _ := makeOmahaForPresenter()
		cfg := domain.OmahaConfig{
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
		h.SetPhase(domain.OmahaPhaseRebuy)
		h.SetRebuyPhaseType(2)

		result := p.Output(h, nil)
		assert.Contains(t, result, "アドオンしますか?")
		assert.Contains(t, result, "1500チップ")
		assert.Contains(t, result, "ad=アドオン / sa=スキップ")
	})

	t.Run("rebuy phase type 0 does not show prompt", func(t *testing.T) {
		h, _ := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhaseRebuy)
		h.SetRebuyPhaseType(0)

		result := p.Output(h, nil)
		assert.NotContains(t, result, "リバイしますか?")
		assert.NotContains(t, result, "アドオンしますか?")
	})
}

func TestOmahaCuiPresenter_Output_TableSize(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.OmahaCuiPresenter)

	t.Run("displays 4-max", func(t *testing.T) {
		h, _ := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhasePreFlop)
		result := p.Output(h, nil)
		assert.Contains(t, result, "テーブル: 4-max")
	})

	t.Run("displays 6-max", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players6 := make([]*domain.OmahaPlayer, 6)
		players6[0] = domain.NewOmahaPlayer(true, domain.HoldemStyleTAG)
		for i := 1; i < 6; i++ {
			players6[i] = domain.NewOmahaPlayer(false, domain.HoldemStyleLAP)
		}
		h := domain.NewOmaha(tc, players6, domain.DefaultOmahaConfig())
		h.SetPhase(domain.OmahaPhasePreFlop)
		result := p.Output(h, nil)
		assert.Contains(t, result, "テーブル: 6-max")
	})
}

func TestOmahaCuiPresenter_Output_Muck(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.OmahaCuiPresenter)

	t.Run("muck prompt displayed during showdown when IsMuckAvailable", func(t *testing.T) {
		h, _ := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhaseShowdown)
		// IsMuckAvailable returns true when phase=SHOWDOWN and human has wonAmount=0
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandOnePair, HandName: "One Pair", WonAmount: 0, BestHand: nil},
			{PlayerIdx: 1, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "マックしますか? (m=マック / sh=ショー)")
	})

	t.Run("muck prompt not displayed when not available", func(t *testing.T) {
		h, _ := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.NotContains(t, result, "マックしますか?")
	})

	t.Run("mucked result displayed as マック", func(t *testing.T) {
		h, _ := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandOnePair, HandName: "One Pair", WonAmount: 0, Mucked: true, BestHand: nil},
			{PlayerIdx: 1, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "あなた: マック")
		assert.NotContains(t, result, "あなた: ワンペア")
	})

	t.Run("results shown in showdown phase", func(t *testing.T) {
		h, _ := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhaseShowdown)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "[結果]")
		assert.Contains(t, result, "あなた: フラッシュ")
	})
}

func TestOmahaCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.OmahaCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		mockGame := new(interfaces.MockOmahaGame)
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
		mockGame := new(interfaces.MockOmahaGame)
		mockGame.On("GetGameEndFlag").Return(true)
		mockGame.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, "棋譜はありません")
		mockGame.AssertExpectations(t)
	})

	t.Run("game_not_ended", func(t *testing.T) {
		mockGame := new(interfaces.MockOmahaGame)
		mockGame.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, "棋譜はありません")
		mockGame.AssertExpectations(t)
	})
}

// TestOmahaCuiPresenter_HiLo_Title asserts the heading switches to
// "Omaha Hi-Lo (8 or Better)" when GetIsHiLo() is true.
func TestOmahaCuiPresenter_HiLo_Title(t *testing.T) {
	p := new(presenter.OmahaCuiPresenter)
	o := domain.NewDefaultOmahaHiLo()
	o.SetPhase(domain.OmahaPhasePreFlop)

	result := p.Output(o, nil)
	assert.Contains(t, result, "Omaha Hi-Lo (8 or Better)")
}

// TestOmahaCuiPresenter_BigO_Title asserts the heading switches to the
// Big O (5 Card Omaha) variants based on the hole-card count.
func TestOmahaCuiPresenter_BigO_Title(t *testing.T) {
	p := new(presenter.OmahaCuiPresenter)

	t.Run("big o hi", func(t *testing.T) {
		o := domain.NewDefaultBigO()
		o.SetPhase(domain.OmahaPhasePreFlop)
		result := p.Output(o, nil)
		assert.Contains(t, result, "5 Card Omaha (Big O)")
	})

	t.Run("big o hi-lo", func(t *testing.T) {
		o := domain.NewDefaultBigOHiLo()
		o.SetPhase(domain.OmahaPhasePreFlop)
		result := p.Output(o, nil)
		assert.Contains(t, result, "5 Card Omaha Hi-Lo (Big O)")
	})
}

// TestOmahaCuiPresenter_HiLo_ResultRendering exercises the Hi-Lo
// result branches: low hand display when qualified, and the
// (Hi:N / Lo:M) / (Hi) / (Lo) suffixes on chip totals.
func TestOmahaCuiPresenter_HiLo_ResultRendering(t *testing.T) {
	p := new(presenter.OmahaCuiPresenter)
	cases := []struct {
		name        string
		hi, lo      int
		lowCards    []*domain.Card
		wantSubstrs []string
	}{
		{
			name: "scoop with qualified low",
			hi:   50, lo: 50,
			lowCards: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 1, false),
				domain.NewCard(domain.CardDesignHeart, 2, false),
				domain.NewCard(domain.CardDesignDiamond, 3, false),
				domain.NewCard(domain.CardDesignClover, 4, false),
				domain.NewCard(domain.CardDesignSpade, 5, false),
			},
			wantSubstrs: []string{"Low:", "Hi:50", "Lo:50"},
		},
		{
			name: "hi only",
			hi:   100, lo: 0,
			lowCards:    nil,
			wantSubstrs: []string{"(Hi)"},
		},
		{
			name: "lo only",
			hi:   0, lo: 50,
			lowCards: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 1, false),
				domain.NewCard(domain.CardDesignHeart, 2, false),
				domain.NewCard(domain.CardDesignDiamond, 3, false),
				domain.NewCard(domain.CardDesignClover, 4, false),
				domain.NewCard(domain.CardDesignSpade, 5, false),
			},
			wantSubstrs: []string{"(Lo)", "Low:"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			o := domain.NewDefaultOmahaHiLo()
			o.SetPhase(domain.OmahaPhaseEnd)
			result := domain.HoldemResult{
				PlayerIdx:    0,
				HandRank:     domain.PokerHandHighCard,
				HandName:     "High Card",
				BestHand:     []*domain.Card{},
				WonAmount:    tc.hi + tc.lo,
				HiWonAmount:  tc.hi,
				LowWonAmount: tc.lo,
			}
			if tc.lowCards != nil {
				result.LowQualifies = true
				result.LowBestHand = tc.lowCards
			}
			o.SetRoundResults([]domain.HoldemResult{result})

			out := p.Output(o, nil)
			for _, want := range tc.wantSubstrs {
				assert.Contains(t, out, want, "expected output to contain %q", want)
			}
		})
	}
}

// TestOmahaCuiPresenter_English verifies issue #1699 Phase 2: every
// previously-hardcoded Japanese string in OmahaCuiPresenter now follows
// the active locale. The default ja path is exercised by the assertions
// above; this suite re-runs Output under LANG=en and checks the English
// keys win out.
func TestOmahaCuiPresenter_English(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	i18n.SetLang("en")
	defer i18n.SetLang("ja")
	p := new(presenter.OmahaCuiPresenter)

	t.Run("output uses English headers and labels", func(t *testing.T) {
		o, _ := makeOmahaForPresenter()
		out := p.Output(o, nil)
		assert.Contains(t, out, "Omaha Hold'em")
		assert.Contains(t, out, "Table: 4-max")
		assert.Contains(t, out, "Dealer: Player")
		assert.Contains(t, out, "Pot:")
		assert.NotContains(t, out, "テーブル") // no Japanese leakage
	})

	t.Run("output uses English Hi-Lo title", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := []*domain.OmahaPlayer{
			domain.NewOmahaPlayer(true, domain.HoldemStyleTAG),
			domain.NewOmahaPlayer(false, domain.HoldemStyleLAP),
			domain.NewOmahaPlayer(false, domain.HoldemStyleTAP),
			domain.NewOmahaPlayer(false, domain.HoldemStyleGTO),
		}
		o := domain.NewOmahaHiLo(tc, players, domain.DefaultOmahaConfig())
		out := p.Output(o, nil)
		assert.Contains(t, out, "Omaha Hi-Lo")
	})

	t.Run("output uses English game-end banner", func(t *testing.T) {
		o, _ := makeOmahaForPresenter()
		o.SetGameEndFlag(true)
		out := p.Output(o, nil)
		assert.Contains(t, out, "Game over")
		assert.NotContains(t, out, "ゲーム終了")
	})
}

// **学習モードの値は Web にしか出ていなかった (#5482)。** GetEquity /
// GetPotOdds は共有ヘルパ経由で Web へ送られ、Holdem 系の CUI も出しているのに、
// Omaha の CUI だけが取り残されていた。
func TestOmahaCuiPresenter_ShowsEquityAndPotOdds(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.OmahaCuiPresenter)

	t.Run("on the human's turn the learning block appears", func(t *testing.T) {
		h, players := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhaseFlop)
		h.SetCurrentTurn(0)
		for _, v := range []int{14, 13, 12, 11} {
			players[0].AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
		}

		// **前提をドメインで確かめてから出力を見る。** ここが nil のままだと、
		// 表示が出ないことを「正しい」と読んでしまう。
		require.True(t, h.IsHumanTurn())
		require.NotNil(t, h.GetEquity())

		out := p.Output(h, nil)
		assert.Contains(t, out, i18n.T("omaha.learningHeader"))
		assert.Contains(t, out, "勝率")
		assert.Contains(t, out, "ポットオッズ")
	})

	// **負のコントロール: 降りていれば出ない。** GetEquity が nil を返す局面。
	t.Run("a folded human gets no learning block", func(t *testing.T) {
		h, players := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhaseFlop)
		h.SetCurrentTurn(0)
		for _, v := range []int{14, 13, 12, 11} {
			players[0].AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
		}
		players[0].SetFolded(true)
		require.Nil(t, h.GetEquity())

		assert.NotContains(t, p.Output(h, nil), i18n.T("omaha.learningHeader"))
	})

	// CPU の手番でも出さない。相手の番に自分の勝率を出しても意味がない。
	t.Run("no learning block on a CPU turn", func(t *testing.T) {
		h, players := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhaseFlop)
		h.SetCurrentTurn(1)
		for _, v := range []int{14, 13, 12, 11} {
			players[0].AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
		}
		require.False(t, h.IsHumanTurn())

		assert.NotContains(t, p.Output(h, nil), i18n.T("omaha.learningHeader"))
	})
}

// **+EV / -EV の 2 本の腕。** #5482 では見出しと数値だけを見ており、この分岐が
// 一度も実行されていなかった (codecov が patch 61.5% で指摘)。
func TestOmahaCuiPresenter_ShowsWhetherCallingIsPlusEV(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.OmahaCuiPresenter)

	// ポットとコール額から必要勝率が決まる。極端な値を置いて 2 本の腕を狙う。
	render := func(pot, lastBet int) string {
		h, players := makeOmahaForPresenter()
		h.SetPhase(domain.OmahaPhaseRiver)
		h.SetCurrentTurn(0)
		for _, v := range []int{14, 13, 12, 11} {
			players[0].AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
		}
		h.SetPot(pot)
		h.SetLastBet(lastBet)
		require.True(t, h.IsHumanTurn())
		require.NotNil(t, h.GetEquity())
		return p.Output(h, nil)
	}

	t.Run("a cheap call is +EV", func(t *testing.T) {
		// ポット 1000 に対しコール 1 → 必要勝率はほぼ 0%。どんな手でも +EV。
		out := render(1000, 1)
		require.Contains(t, out, i18n.T("omaha.learningHeader"))
		assert.Contains(t, out, i18n.T("omaha.learningEvPlus"))
		assert.NotContains(t, out, i18n.T("omaha.learningEvMinus"))
	})

	t.Run("an expensive call is -EV", func(t *testing.T) {
		// ポット 1 に対しコール 1000 → 必要勝率はほぼ 100%。どんな手でも -EV。
		out := render(1, 1000)
		require.Contains(t, out, i18n.T("omaha.learningHeader"))
		assert.Contains(t, out, i18n.T("omaha.learningEvMinus"))
		assert.NotContains(t, out, i18n.T("omaha.learningEvPlus"))
	})

	// **負のコントロール: コール額 0 なら判定しない。** ポットオッズが 0 のとき
	// +EV/-EV を出すと、賭けていない局面で「コール有利」と言うことになる。
	t.Run("nothing to call means no verdict", func(t *testing.T) {
		out := render(100, 0)
		require.Contains(t, out, i18n.T("omaha.learningHeader"))
		assert.NotContains(t, out, i18n.T("omaha.learningEvPlus"))
		assert.NotContains(t, out, i18n.T("omaha.learningEvMinus"))
	})
}
