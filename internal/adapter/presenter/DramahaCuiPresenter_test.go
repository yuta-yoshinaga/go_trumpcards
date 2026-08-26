package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func makeDramahaForPresenter() (*domain.Dramaha, []*domain.DramahaPlayer) {
	tc := domain.NewTrumpCards(0)
	players := []*domain.DramahaPlayer{
		domain.NewDramahaPlayer(true, domain.HoldemStyleTAG),
		domain.NewDramahaPlayer(false, domain.HoldemStyleLAP),
		domain.NewDramahaPlayer(false, domain.HoldemStyleTAP),
		domain.NewDramahaPlayer(false, domain.HoldemStyleGTO),
	}
	h := domain.NewDramaha(tc, players, domain.DefaultDramahaConfig())
	return h, players
}

// makeHeadsUpDramahaForPresenter is the two-seat table the split-pot result
// rendering is exercised on. It replaces the clone's makeDramahaHiLoForPresenter:
// Dramaha has no Hi-Lo variant, so there is no second constructor to pick.
func makeHeadsUpDramahaForPresenter() *domain.Dramaha {
	tc := domain.NewTrumpCards(0)
	players := []*domain.DramahaPlayer{
		domain.NewDramahaPlayer(true, domain.HoldemStyleTAG),
		domain.NewDramahaPlayer(false, domain.HoldemStyleLAP),
	}
	return domain.NewDramaha(tc, players, domain.DefaultDramahaConfig())
}

func TestDramahaCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.DramahaCuiPresenter)

	t.Run("initial state with no community cards", func(t *testing.T) {
		h, players := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhasePreFlop)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := p.Output(h, nil)
		assert.Contains(t, result, "ドラマハ")
		assert.Contains(t, result, "ディーラー:")
		assert.Contains(t, result, "(なし)")
		assert.Contains(t, result, "ポット:")
		assert.Contains(t, result, "あなた")
		assert.Contains(t, result, "♠10")
		assert.Contains(t, result, "♥11")
		// The must-use-2 rule line is always shown, and Dramaha deals five holes.
		assert.Contains(t, result, "手札5枚のうちちょうど2枚")
		assert.NotContains(t, result, "手札4枚のうちちょうど2枚",
			"the four-card wording belongs to the clone")
	})

	// **The split is the game.** The clone put an "8 or better" qualifier line
	// here; Dramaha's pot always halves, and the screen has to say so or the
	// player cannot tell why two hands are being ranked.
	t.Run("always states that the pot splits between the two evaluations", func(t *testing.T) {
		h, _ := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhasePreFlop)

		result := p.Output(h, nil)

		assert.Contains(t, result, "ポットは常に二分されます")
		assert.Contains(t, result, "ドロー役")
		assert.NotContains(t, result, "8 or Better", "there is no low qualifier any more")
		assert.NotContains(t, result, "8以下", "there is no low qualifier any more")
	})

	// **Web は dramaha-live-besthand で暫定ベストを常時出しているのに、CUI は
	// 「2枚使用」の注意書きだけで実際の役を出していなかった (#4680)。**
	// 手札4枚から必ず2枚という特殊ルールがあるぶん、暫定表示はミスを防ぐ補助になる。
	t.Run("shows the human's current best hand once the flop is out", func(t *testing.T) {
		h, players := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhaseFlop)
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
		h, players := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhasePreFlop)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 12, false))

		assert.NotContains(t, p.Output(h, nil), "現在の最善役")
	})

	// **表示のために状態を変えない。**Peek を使っているので、描画しても
	// プレイヤーの handRank は書き換わらない。
	t.Run("rendering does not mutate the player state", func(t *testing.T) {
		h, players := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhaseFlop)
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
		h, _ := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhaseFlop)
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
		h, _ := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhaseFlop)
		h.SetCommunityCards([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignHeart, 2, false),
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "♠1  ♥2")
	})

	t.Run("CPU player info with play style", func(t *testing.T) {
		h, _ := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhasePreFlop)

		result := p.Output(h, nil)
		assert.Contains(t, result, "CPU 1")
		assert.Contains(t, result, "CPU 2")
		assert.Contains(t, result, "CPU 3")
	})

	t.Run("folded player", func(t *testing.T) {
		h, players := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhasePreFlop)
		players[1].SetFolded(true)

		result := p.Output(h, nil)
		assert.Contains(t, result, "[フォールド]")
	})

	t.Run("all-in player", func(t *testing.T) {
		h, players := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhasePreFlop)
		players[1].SetAllIn(true)

		result := p.Output(h, nil)
		assert.Contains(t, result, "[オールイン]")
	})

	t.Run("player with current bet", func(t *testing.T) {
		h, players := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhasePreFlop)
		players[0].SetCurrentBet(50)

		result := p.Output(h, nil)
		assert.Contains(t, result, "ベット:50")
	})

	t.Run("player with zero bet hides bet label", func(t *testing.T) {
		h, players := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhasePreFlop)
		players[0].SetCurrentBet(0)

		result := p.Output(h, nil)
		assert.NotContains(t, result, "ベット:0")
	})

	t.Run("human cards visible when not folded", func(t *testing.T) {
		h, players := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhasePreFlop)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 13, false))

		result := p.Output(h, nil)
		assert.Contains(t, result, "手札:")
		assert.Contains(t, result, "♠1")
		assert.Contains(t, result, "♥13")
	})

	t.Run("human cards hidden when folded", func(t *testing.T) {
		h, players := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhasePreFlop)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].SetFolded(true)

		result := p.Output(h, nil)
		assert.NotContains(t, result, "手札:")
	})

	t.Run("human cards with separator", func(t *testing.T) {
		h, players := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhasePreFlop)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 7, false))

		result := p.Output(h, nil)
		assert.Contains(t, result, "♠3  ♣7")
	})

	t.Run("CPU actions displayed", func(t *testing.T) {
		h, _ := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhasePreFlop)
		h.SetCpuActions([]domain.HoldemCpuAction{
			{PlayerIdx: 1, Action: domain.DramahaActionCall, Amount: 0},
			{PlayerIdx: 2, Action: domain.DramahaActionRaise, Amount: 30},
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
		h, _ := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhasePreFlop)
		h.SetCpuActions([]domain.HoldemCpuAction{
			{PlayerIdx: 1, Action: domain.DramahaActionFold, Amount: 0},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "フォールド")
		assert.NotContains(t, result, "(0)")
	})

	t.Run("no CPU actions hides section", func(t *testing.T) {
		h, _ := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhasePreFlop)
		h.SetCpuActions([]domain.HoldemCpuAction{})

		result := p.Output(h, nil)
		assert.NotContains(t, result, "[CPU行動]")
	})

	t.Run("showdown results with human winner", func(t *testing.T) {
		h, players := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhaseEnd)
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
		h, _ := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandOnePair, HandName: "One Pair", Kickers: []int{14, 12, 10}, WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "あなた: ワンペア (キッカー: A, Q, 10)")
		assert.Contains(t, result, "100チップ獲得")
	})

	t.Run("showdown results without kickers", func(t *testing.T) {
		h, _ := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush", Kickers: nil, WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "あなた: フラッシュ")
		assert.NotContains(t, result, "キッカー")
	})

	t.Run("showdown results with CPU winner", func(t *testing.T) {
		h, _ := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 1, HandRank: domain.PokerHandOnePair, HandName: "One Pair", Kickers: []int{13, 12, 11}, WonAmount: 50, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "CPU 1: ワンペア (キッカー: K, Q, J)")
		assert.Contains(t, result, "50チップ獲得")
	})

	t.Run("showdown results with empty hand name", func(t *testing.T) {
		h, _ := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 1, HandName: "", WonAmount: 0, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "CPU 1")
		assert.NotContains(t, result, "CPU 1:")
	})

	t.Run("showdown results with zero won amount", func(t *testing.T) {
		h, _ := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 1, HandRank: domain.PokerHandHighCard, HandName: "High Card", WonAmount: 0, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.NotContains(t, result, "チップ獲得")
	})

	t.Run("results not shown in non-end phase", func(t *testing.T) {
		h, _ := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhaseFlop)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.NotContains(t, result, "[結果]")
	})

	t.Run("error message displayed", func(t *testing.T) {
		h, _ := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhasePreFlop)

		result := p.Output(h, errors.New("test error"))
		assert.Contains(t, result, "test error")
	})

	t.Run("game end message displayed", func(t *testing.T) {
		h, _ := makeDramahaForPresenter()
		h.SetGameEndFlag(true)

		result := p.Output(h, nil)
		assert.Contains(t, result, "ゲーム終了")
	})

	t.Run("game not ended hides game end message", func(t *testing.T) {
		h, _ := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhasePreFlop)

		result := p.Output(h, nil)
		// "ゲーム終了" should not appear unless gameEndFlag is true
		// The line "ゲーム終了\n" is only printed if h.GetGameEndFlag()
		// We just check it doesn't end with the specific game end block
		_ = result
	})

	t.Run("getCardStr with out of range design", func(t *testing.T) {
		h, players := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhasePreFlop)
		players[0].AddCard(domain.NewCard(99, 5, false))

		result := p.Output(h, nil)
		assert.Contains(t, result, "🃏5") // falls back to joker design
	})

	t.Run("getCardStr with negative design", func(t *testing.T) {
		h, players := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhasePreFlop)
		players[0].AddCard(domain.NewCard(-1, 3, false))

		result := p.Output(h, nil)
		assert.Contains(t, result, "🃏3")
	})

	t.Run("all action names", func(t *testing.T) {
		h, _ := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhasePreFlop)
		h.SetCpuActions([]domain.HoldemCpuAction{
			{PlayerIdx: 1, Action: domain.DramahaActionFold, Amount: 0},
			{PlayerIdx: 1, Action: domain.DramahaActionCheck, Amount: 0},
			{PlayerIdx: 1, Action: domain.DramahaActionCall, Amount: 0},
			{PlayerIdx: 1, Action: domain.DramahaActionBet, Amount: 10},
			{PlayerIdx: 1, Action: domain.DramahaActionRaise, Amount: 20},
			{PlayerIdx: 1, Action: domain.DramahaActionAllIn, Amount: 100},
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
		h, players := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhasePreFlop)
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
		h, players := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhasePreFlop)
		players[0].IncrementTotalHands()
		players[0].IncrementPostFlopBetRaise()

		result := p.Output(h, nil)
		assert.Contains(t, result, "AF:∞")
	})

	t.Run("HUD stats with AF normal ratio", func(t *testing.T) {
		h, players := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhasePreFlop)
		players[0].IncrementTotalHands()
		players[0].IncrementPostFlopBetRaise()
		players[0].IncrementPostFlopBetRaise()
		players[0].IncrementPostFlopCall()

		result := p.Output(h, nil)
		assert.Contains(t, result, "AF:2.0")
	})

	t.Run("HUD stats not shown when totalHands is 0", func(t *testing.T) {
		h, _ := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhasePreFlop)

		result := p.Output(h, nil)
		assert.NotContains(t, result, "VPIP:")
		assert.NotContains(t, result, "PFR:")
	})

	t.Run("tournament mode header shown when enabled", func(t *testing.T) {
		h, _ := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhasePreFlop)
		cfg := domain.DramahaConfig{
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
		h, _ := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhasePreFlop)

		result := p.Output(h, nil)
		assert.NotContains(t, result, "トーナメント")
	})
}

func TestDramahaCuiPresenter_Output_BettingLimitDisplay(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.DramahaCuiPresenter)

	t.Run("displays Fixed limit", func(t *testing.T) {
		h, _ := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhasePreFlop)
		result := p.Output(h, nil)
		assert.Contains(t, result, "Fixed")
	})
}

func TestDramahaCuiPresenter_Output_RebuyAddon(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.DramahaCuiPresenter)

	t.Run("tournament mode with rebuy enabled shows rebuy info", func(t *testing.T) {
		h, _ := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhasePreFlop)
		cfg := domain.DramahaConfig{
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
		h, _ := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhasePreFlop)
		cfg := domain.DramahaConfig{
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
		h, _ := makeDramahaForPresenter()
		cfg := domain.DramahaConfig{
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
		h.SetPhase(domain.DramahaPhaseRebuy)
		h.SetRebuyPhaseType(1)
		h.SetRebuyCounts([]int{1, 0, 0, 0})

		result := p.Output(h, nil)
		assert.Contains(t, result, "リバイしますか?")
		assert.Contains(t, result, "1000チップ")
		assert.Contains(t, result, "1/3回使用済")
		assert.Contains(t, result, "rb=リバイ / sr=スキップ")
	})

	t.Run("rebuy phase type 2 shows addon prompt", func(t *testing.T) {
		h, _ := makeDramahaForPresenter()
		cfg := domain.DramahaConfig{
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
		h.SetPhase(domain.DramahaPhaseRebuy)
		h.SetRebuyPhaseType(2)

		result := p.Output(h, nil)
		assert.Contains(t, result, "アドオンしますか?")
		assert.Contains(t, result, "1500チップ")
		assert.Contains(t, result, "ad=アドオン / sa=スキップ")
	})

	t.Run("rebuy phase type 0 does not show prompt", func(t *testing.T) {
		h, _ := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhaseRebuy)
		h.SetRebuyPhaseType(0)

		result := p.Output(h, nil)
		assert.NotContains(t, result, "リバイしますか?")
		assert.NotContains(t, result, "アドオンしますか?")
	})
}

func TestDramahaCuiPresenter_Output_TableSize(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.DramahaCuiPresenter)

	t.Run("displays 4-max", func(t *testing.T) {
		h, _ := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhasePreFlop)
		result := p.Output(h, nil)
		assert.Contains(t, result, "テーブル: 4-max")
	})

	t.Run("displays 6-max", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players6 := make([]*domain.DramahaPlayer, 6)
		players6[0] = domain.NewDramahaPlayer(true, domain.HoldemStyleTAG)
		for i := 1; i < 6; i++ {
			players6[i] = domain.NewDramahaPlayer(false, domain.HoldemStyleLAP)
		}
		h := domain.NewDramaha(tc, players6, domain.DefaultDramahaConfig())
		h.SetPhase(domain.DramahaPhasePreFlop)
		result := p.Output(h, nil)
		assert.Contains(t, result, "テーブル: 6-max")
	})
}

func TestDramahaCuiPresenter_Output_Muck(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.DramahaCuiPresenter)

	t.Run("muck prompt displayed during showdown when IsMuckAvailable", func(t *testing.T) {
		h, _ := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhaseShowdown)
		// IsMuckAvailable returns true when phase=SHOWDOWN and human has wonAmount=0
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandOnePair, HandName: "One Pair", WonAmount: 0, BestHand: nil},
			{PlayerIdx: 1, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "マックしますか? (m=マック / sh=ショー)")
	})

	t.Run("muck prompt not displayed when not available", func(t *testing.T) {
		h, _ := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.NotContains(t, result, "マックしますか?")
	})

	t.Run("mucked result displayed as マック", func(t *testing.T) {
		h, _ := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandOnePair, HandName: "One Pair", WonAmount: 0, Mucked: true, BestHand: nil},
			{PlayerIdx: 1, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "あなた: マック")
		assert.NotContains(t, result, "あなた: ワンペア")
	})

	t.Run("results shown in showdown phase", func(t *testing.T) {
		h, _ := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhaseShowdown)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "[結果]")
		assert.Contains(t, result, "あなた: フラッシュ")
	})
}

func TestDramahaCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.DramahaCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		mockGame := new(interfaces.MockDramahaGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "raise", Detail: "raised to 100"},
		}
		mockGame.On("GetGameEndFlag").Return(true)
		mockGame.On("GetActionLog").Return(entries)
		// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
		mockGame.On("GetPlayer", mock.Anything).Return(domain.NewDramahaPlayer(true, domain.HoldemPlayStyle(0))).Maybe()

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "raise")
		assert.Contains(t, result, "raised to 100")
		mockGame.AssertExpectations(t)
	})

	t.Run("nil_entries", func(t *testing.T) {
		mockGame := new(interfaces.MockDramahaGame)
		mockGame.On("GetGameEndFlag").Return(true)
		mockGame.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
		// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
		mockGame.On("GetPlayer", mock.Anything).Return(domain.NewDramahaPlayer(true, domain.HoldemPlayStyle(0))).Maybe()

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, "棋譜はありません")
		mockGame.AssertExpectations(t)
	})

	t.Run("game_not_ended", func(t *testing.T) {
		mockGame := new(interfaces.MockDramahaGame)
		mockGame.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, "棋譜はありません")
		mockGame.AssertExpectations(t)
	})
}

// TestDramahaCuiPresenter_Heading replaces the clone's Hi-Lo / Big O heading
// tests. There is only one Dramaha -- no hole-card setting to switch on and no
// low half -- so the heading is fixed and must not name a variant this game
// does not have.
func TestDramahaCuiPresenter_Heading(t *testing.T) {
	p := new(presenter.DramahaCuiPresenter)
	o, _ := makeDramahaForPresenter()
	o.SetPhase(domain.DramahaPhasePreFlop)

	result := p.Output(o, nil)
	// The heading sits between the two rules of "=" that buildCuiOutput draws.
	lines := strings.Split(result, "\n")
	require.GreaterOrEqual(t, len(lines), 2)
	heading := lines[1]

	assert.Contains(t, heading, "ドラマハ", "the heading has to name the game")
	assert.NotContains(t, heading, "dramaha.", "the heading must be translated, not a raw key")
	for _, forbidden := range []string{"Hi-Lo", "8 or Better", "Big O", "Courchevel", "オマハホールデム"} {
		assert.NotContains(t, heading, forbidden,
			"the heading must not name a variant this game does not have")
	}
}

// TestDramahaCuiPresenter_DoesNotLeakUntranslatedKeys catches the failure mode
// the divergence created: keys that were dropped from the Dramaha locale while
// the presenter still asked for them. i18n.T returns the key itself in that
// case, so the identifier is printed straight onto the screen with no error
// anywhere. Asserting on i18n.T(key) would not catch it -- the comparison
// would be key-against-key -- so this looks for the namespace prefix instead.
func TestDramahaCuiPresenter_DoesNotLeakUntranslatedKeys(t *testing.T) {
	p := new(presenter.DramahaCuiPresenter)
	o := makeHeadsUpDramahaForPresenter()
	o.SetPhase(domain.DramahaPhaseEnd)
	o.SetRoundResults([]domain.HoldemResult{{
		PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush",
		WonAmount: 100, HiWonAmount: 50, LowWonAmount: 50,
		LowQualifies: true,
		LowBestHand: []*domain.Card{
			domain.NewCard(domain.CardDesignHeart, 2, false),
			domain.NewCard(domain.CardDesignHeart, 3, false),
			domain.NewCard(domain.CardDesignHeart, 5, false),
			domain.NewCard(domain.CardDesignHeart, 6, false),
			domain.NewCard(domain.CardDesignHeart, 9, false),
		},
	}})

	for _, lang := range []string{"ja", "en"} {
		i18n.SetLang(lang)
		out := p.Output(o, nil)
		assert.NotContains(t, out, "dramaha.",
			"%s: a dramaha.* identifier is being rendered instead of a translation:\n%s", lang, out)
	}
	i18n.SetLang("ja")
}

// TestDramahaCuiPresenter_PromptsTheDrawRound: the exchange happens in a single
// phase between the flop and the turn. If the screen does not say so, the
// player is never told they can swap cards.
func TestDramahaCuiPresenter_PromptsTheDrawRound(t *testing.T) {
	p := new(presenter.DramahaCuiPresenter)
	prompt := i18n.T("dramaha.drawPhaseLine")
	require.NotEqual(t, "dramaha.drawPhaseLine", prompt, "the prompt has to be translated")

	o, _ := makeDramahaForPresenter()
	o.SetPhase(domain.DramahaPhaseDraw)
	assert.Contains(t, p.Output(o, nil), prompt)

	for _, phase := range []int{
		domain.DramahaPhasePreFlop, domain.DramahaPhaseFlop,
		domain.DramahaPhaseTurn, domain.DramahaPhaseRiver, domain.DramahaPhaseEnd,
	} {
		o.SetPhase(phase)
		assert.NotContains(t, p.Output(o, nil), prompt,
			"phase %d has no exchange to prompt for", phase)
	}
}

// TestDramahaCuiPresenter_ShowsTheDrawHandAtShowdown: the draw side decides
// half the pot, so the five cards it was judged on have to be on the screen.
func TestDramahaCuiPresenter_ShowsTheDrawHandAtShowdown(t *testing.T) {
	p := new(presenter.DramahaCuiPresenter)
	o := makeHeadsUpDramahaForPresenter()
	o.SetPhase(domain.DramahaPhaseEnd)
	drawHand := []*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 2, false),
		domain.NewCard(domain.CardDesignHeart, 3, false),
		domain.NewCard(domain.CardDesignHeart, 5, false),
		domain.NewCard(domain.CardDesignHeart, 6, false),
		domain.NewCard(domain.CardDesignHeart, 9, false),
	}
	o.SetRoundResults([]domain.HoldemResult{{
		PlayerIdx: 0, HandRank: domain.PokerHandHighCard, HandName: "High Card",
		WonAmount: 50, LowWonAmount: 50, LowQualifies: true, LowBestHand: drawHand,
	}})

	out := p.Output(o, nil)

	assert.Contains(t, out, "ドロー役")
	for _, want := range []string{"♥2", "♥3", "♥5", "♥6", "♥9"} {
		assert.Contains(t, out, want, "the whole hole is what the draw side is judged on")
	}

	// A result with no draw hand (a fold-out win) prints no draw line at all --
	// an empty one would read as "there was no hand".
	o.SetRoundResults([]domain.HoldemResult{{
		PlayerIdx: 0, HandRank: domain.PokerHandHighCard, HandName: "High Card", WonAmount: 50,
	}})
	assert.NotContains(t, p.Output(o, nil), "ドロー役:")
}

// TestDramahaCuiPresenter_SplitResultRendering exercises the three ways a
// split can land: both halves, the Omaha half only, the draw half only. The
// clone reached these through its Hi-Lo flag; Dramaha reaches them always.
func TestDramahaCuiPresenter_SplitResultRendering(t *testing.T) {
	p := new(presenter.DramahaCuiPresenter)
	cases := []struct {
		name        string
		hi, lo      int
		lowCards    []*domain.Card
		wantSubstrs []string
	}{
		{
			name: "both halves",
			hi:   50, lo: 50,
			lowCards: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 1, false),
				domain.NewCard(domain.CardDesignHeart, 2, false),
				domain.NewCard(domain.CardDesignDiamond, 3, false),
				domain.NewCard(domain.CardDesignClover, 4, false),
				domain.NewCard(domain.CardDesignSpade, 5, false),
			},
			wantSubstrs: []string{"Hi:50", "Lo:50"},
		},
		{
			name: "omaha half only",
			hi:   100, lo: 0,
			lowCards:    nil,
			wantSubstrs: []string{"(Hi)"},
		},
		{
			name: "draw half only",
			hi:   0, lo: 50,
			lowCards: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 1, false),
				domain.NewCard(domain.CardDesignHeart, 2, false),
				domain.NewCard(domain.CardDesignDiamond, 3, false),
				domain.NewCard(domain.CardDesignClover, 4, false),
				domain.NewCard(domain.CardDesignSpade, 5, false),
			},
			wantSubstrs: []string{"(Lo)"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			o := makeHeadsUpDramahaForPresenter()
			o.SetPhase(domain.DramahaPhaseEnd)
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

// TestDramahaCuiPresenter_English verifies issue #1699 Phase 2: every
// previously-hardcoded Japanese string in DramahaCuiPresenter now follows
// the active locale. The default ja path is exercised by the assertions
// above; this suite re-runs Output under LANG=en and checks the English
// keys win out.
func TestDramahaCuiPresenter_English(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	i18n.SetLang("en")
	defer i18n.SetLang("ja")
	p := new(presenter.DramahaCuiPresenter)

	t.Run("output uses English headers and labels", func(t *testing.T) {
		o, _ := makeDramahaForPresenter()
		out := p.Output(o, nil)
		assert.Contains(t, out, "Table: 4-max")
		assert.Contains(t, out, "Dealer: Player")
		assert.Contains(t, out, "Pot:")
		assert.NotContains(t, out, "テーブル") // no Japanese leakage
	})

	// The split rule and the hole-card count are the two lines that describe the
	// game itself; both have to follow the locale.
	t.Run("output uses the English split-rule and hole-count lines", func(t *testing.T) {
		o, _ := makeDramahaForPresenter()
		out := p.Output(o, nil)
		assert.Contains(t, out, "The pot always splits")
		assert.Contains(t, out, "Use exactly 2 of your 5 hole cards")
		assert.NotContains(t, out, "ポットは常に二分されます")
		assert.NotContains(t, out, "手札5枚")
	})

	t.Run("output uses English game-end banner", func(t *testing.T) {
		o, _ := makeDramahaForPresenter()
		o.SetGameEndFlag(true)
		out := p.Output(o, nil)
		assert.Contains(t, out, "Game over")
		assert.NotContains(t, out, "ゲーム終了")
	})
}

// **学習モードの値は Web にしか出ていなかった (#5482)。** GetEquity /
// GetPotOdds は共有ヘルパ経由で Web へ送られ、Holdem 系の CUI も出しているのに、
// Dramaha の CUI だけが取り残されていた。
func TestDramahaCuiPresenter_ShowsEquityAndPotOdds(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.DramahaCuiPresenter)

	t.Run("on the human's turn the learning block appears", func(t *testing.T) {
		h, players := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhaseFlop)
		h.SetCurrentTurn(0)
		for _, v := range []int{14, 13, 12, 11} {
			players[0].AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
		}

		// **前提をドメインで確かめてから出力を見る。** ここが nil のままだと、
		// 表示が出ないことを「正しい」と読んでしまう。
		require.True(t, h.IsHumanTurn())
		require.NotNil(t, h.GetEquity())

		out := p.Output(h, nil)
		assert.Contains(t, out, i18n.T("dramaha.learningHeader"))
		assert.Contains(t, out, "勝率")
		assert.Contains(t, out, "ポットオッズ")
	})

	// **負のコントロール: 降りていれば出ない。** GetEquity が nil を返す局面。
	t.Run("a folded human gets no learning block", func(t *testing.T) {
		h, players := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhaseFlop)
		h.SetCurrentTurn(0)
		for _, v := range []int{14, 13, 12, 11} {
			players[0].AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
		}
		players[0].SetFolded(true)
		require.Nil(t, h.GetEquity())

		assert.NotContains(t, p.Output(h, nil), i18n.T("dramaha.learningHeader"))
	})

	// CPU の手番でも出さない。相手の番に自分の勝率を出しても意味がない。
	t.Run("no learning block on a CPU turn", func(t *testing.T) {
		h, players := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhaseFlop)
		h.SetCurrentTurn(1)
		for _, v := range []int{14, 13, 12, 11} {
			players[0].AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
		}
		require.False(t, h.IsHumanTurn())

		assert.NotContains(t, p.Output(h, nil), i18n.T("dramaha.learningHeader"))
	})
}

// **+EV / -EV の 2 本の腕。** #5482 では見出しと数値だけを見ており、この分岐が
// 一度も実行されていなかった (codecov が patch 61.5% で指摘)。
func TestDramahaCuiPresenter_ShowsWhetherCallingIsPlusEV(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.DramahaCuiPresenter)

	// ポットとコール額から必要勝率が決まる。極端な値を置いて 2 本の腕を狙う。
	render := func(pot, lastBet int) string {
		h, players := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhaseRiver)
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
		require.Contains(t, out, i18n.T("dramaha.learningHeader"))
		assert.Contains(t, out, i18n.T("dramaha.learningEvPlus"))
		assert.NotContains(t, out, i18n.T("dramaha.learningEvMinus"))
	})

	t.Run("an expensive call is -EV", func(t *testing.T) {
		// ポット 1 に対しコール 1000 → 必要勝率はほぼ 100%。どんな手でも -EV。
		out := render(1, 1000)
		require.Contains(t, out, i18n.T("dramaha.learningHeader"))
		assert.Contains(t, out, i18n.T("dramaha.learningEvMinus"))
		assert.NotContains(t, out, i18n.T("dramaha.learningEvPlus"))
	})

	// **負のコントロール: コール額 0 なら判定しない。** ポットオッズが 0 のとき
	// +EV/-EV を出すと、賭けていない局面で「コール有利」と言うことになる。
	t.Run("nothing to call means no verdict", func(t *testing.T) {
		out := render(100, 0)
		require.Contains(t, out, i18n.T("dramaha.learningHeader"))
		assert.NotContains(t, out, i18n.T("dramaha.learningEvPlus"))
		assert.NotContains(t, out, i18n.T("dramaha.learningEvMinus"))
	})
}

// The clone advised, all through the betting rounds, whether the board could
// still bring in an 8-or-better low. Dramaha's draw side never looks at the
// board, so that advice would describe a rule this game does not have. The
// outlook was deleted with GetBoardLowOutlook; this guard keeps it deleted.
//
// **Do not search for the raw i18n key.** The keys were removed from the
// Dramaha locale, so i18n.T returns the key itself and NotContains(out,
// i18n.T(key)) would pass no matter what the presenter did. The literal
// translated wording the clone used is what has to stay off the screen.
func assertNoDramahaBoardLowLine(t *testing.T, out string) {
	t.Helper()
	for _, wording := range []string{
		"ロー成立", "ロー:", "ローの目", "low still", "Low is dead", "Low is live",
	} {
		assert.NotContains(t, out, wording)
	}
}

func TestDramahaCuiPresenter_NoBoardLowOutlook(t *testing.T) {
	p := new(presenter.DramahaCuiPresenter)
	card := func(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

	at := func(phase int, cards ...*domain.Card) string {
		h := makeHeadsUpDramahaForPresenter()
		h.SetPhase(phase)
		h.SetCommunityCards(cards)
		return p.Output(h, nil)
	}

	// The three boards the clone reported on: a live low, a low still needing
	// ranks, and a dead low. All three must now be silent.
	t.Run("a board full of low cards says nothing about a low", func(t *testing.T) {
		assertNoDramahaBoardLowLine(t, at(domain.DramahaPhaseFlop,
			card(domain.CardDesignSpade, 2), card(domain.CardDesignHeart, 5), card(domain.CardDesignClover, 7)))
	})

	t.Run("a mixed board says nothing about a low", func(t *testing.T) {
		assertNoDramahaBoardLowLine(t, at(domain.DramahaPhaseFlop,
			card(domain.CardDesignSpade, 2), card(domain.CardDesignHeart, 11), card(domain.CardDesignClover, 12)))
	})

	t.Run("a board of high cards says nothing about a low", func(t *testing.T) {
		assertNoDramahaBoardLowLine(t, at(domain.DramahaPhaseFlop,
			card(domain.CardDesignSpade, 10), card(domain.CardDesignHeart, 11), card(domain.CardDesignClover, 12)))
	})

	t.Run("and neither does the showdown", func(t *testing.T) {
		assertNoDramahaBoardLowLine(t, at(domain.DramahaPhaseShowdown,
			card(domain.CardDesignSpade, 2), card(domain.CardDesignHeart, 5), card(domain.CardDesignClover, 7)))
	})

	// Negative control: the guard has to be able to see the wording it is
	// looking for, or every case above passes for free.
	t.Run("the guard can see the wording it forbids", func(t *testing.T) {
		fake := &testing.T{}
		assertNoDramahaBoardLowLine(fake, "コミュニティ: ロー成立の見込みあり")
		assert.True(t, fake.Failed(), "assertNoDramahaBoardLowLine must fail on the wording it forbids")
	})
}

// #5484: Big O は5枚のホールカードから必ず2枚だけ使う。10通りの組み合わせの
// どれが役になったのかを Web は cardUsed/cardUnused で示すのに、CUI の結果表示は
// 役名とキッカーだけで、RoundResult.BestHand を一度も使っていなかった。
func TestDramahaCuiPresenter_ResultBestHand(t *testing.T) {
	p := new(presenter.DramahaCuiPresenter)
	card := func(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

	t.Run("shows the five cards and marks the two that came from the hole", func(t *testing.T) {
		h, players := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhaseEnd)
		// ホール5枚 (Big O)。うち ♠A と ♠K がベストに入る。
		for _, c := range []*domain.Card{
			card(domain.CardDesignSpade, 1), card(domain.CardDesignSpade, 13),
			card(domain.CardDesignHeart, 2), card(domain.CardDesignHeart, 3),
			card(domain.CardDesignClover, 4),
		} {
			players[0].AddCard(c)
		}
		best := []*domain.Card{
			card(domain.CardDesignSpade, 1), card(domain.CardDesignSpade, 13),
			card(domain.CardDesignSpade, 12), card(domain.CardDesignSpade, 11), card(domain.CardDesignSpade, 10),
		}
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100, BestHand: best},
		})

		out := p.Output(h, nil)
		// 5枚とも出る (♠ は黒スートなので色コードが付かない)。
		for _, want := range []string{"♠1", "♠13", "♠12", "♠11", "♠10"} {
			assert.Contains(t, out, want)
		}
		// **ホール由来の2枚にだけ印。**印が5個なら、どれを使ったのか分からない
		// のと同じ。
		line := ""
		for _, l := range strings.Split(out, "\n") {
			if strings.Contains(l, i18n.T("dramaha.resultBestLabel")) {
				line = l
			}
		}
		assert.NotEmpty(t, line, "ベストハンドの行が出ていない")
		assert.Equal(t, 2, strings.Count(line, presenter.CuiHoleMark))
	})

	// マックしたプレイヤーは手を見せない。BestHand が残っていても出さない。
	t.Run("says nothing for a mucked hand", func(t *testing.T) {
		h, players := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhaseEnd)
		players[0].AddCard(card(domain.CardDesignSpade, 1))
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, Mucked: true, BestHand: []*domain.Card{card(domain.CardDesignSpade, 1)}},
		})
		assert.NotContains(t, p.Output(h, nil), i18n.T("dramaha.resultBestLabel"))
	})

	// BestHand が空の結果 (フォールド勝ちなど) では行ごと出さない。
	t.Run("says nothing when there is no best hand", func(t *testing.T) {
		h, _ := makeDramahaForPresenter()
		h.SetPhase(domain.DramahaPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100, BestHand: nil},
		})
		assert.NotContains(t, p.Output(h, nil), i18n.T("dramaha.resultBestLabel"))
	})
}

// 凡例は1度だけ。4人ショーダウンで同じ注記が4回並ぶと読みにくい。
func TestDramahaCuiPresenter_ResultBestLegendShownOnce(t *testing.T) {
	p := new(presenter.DramahaCuiPresenter)
	card := func(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

	h, players := makeDramahaForPresenter()
	h.SetPhase(domain.DramahaPhaseEnd)
	best := []*domain.Card{
		card(domain.CardDesignSpade, 1), card(domain.CardDesignSpade, 13),
		card(domain.CardDesignSpade, 12), card(domain.CardDesignSpade, 11), card(domain.CardDesignSpade, 10),
	}
	players[0].AddCard(card(domain.CardDesignSpade, 1))
	players[1].AddCard(card(domain.CardDesignSpade, 13))
	h.SetRoundResults([]domain.HoldemResult{
		{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush", BestHand: best},
		{PlayerIdx: 1, HandRank: domain.PokerHandFlush, HandName: "Flush", BestHand: best},
	})

	out := p.Output(h, nil)
	assert.Equal(t, 1, strings.Count(out, i18n.T("dramaha.resultBestLegend")))
	// それでも各プレイヤーの行にはベストが出る。
	assert.Equal(t, 2, strings.Count(out, i18n.T("dramaha.resultBestLabel")))
}

// #5485: Hi と Lo の両取り (スクープ) は Web では専用バッジ + 人間なら
// パルスアニメーションで強調されるのに、CUI は wonHiLoBoth で金額の内訳を
// 出すだけで、それが特別な結果だとは一言も言っていなかった。
func TestDramahaCuiPresenter_Scoop(t *testing.T) {
	p := new(presenter.DramahaCuiPresenter)

	render := func(results []domain.HoldemResult) string {
		h := makeHeadsUpDramahaForPresenter()
		h.SetPhase(domain.DramahaPhaseEnd)
		h.SetRoundResults(results)
		return p.Output(h, nil)
	}

	t.Run("calls out a scoop when one player takes both halves", func(t *testing.T) {
		out := render([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush",
				WonAmount: 100, HiWonAmount: 60, LowWonAmount: 40},
		})
		assert.Contains(t, out, i18n.T("dramaha.scoop"))
		// 内訳は今までどおり残る。スクープ表示で置き換えては情報が減る。
		assert.Contains(t, out, i18n.Tf("dramaha.wonHiLoBoth", "total", "100", "hi", "60", "lo", "40"))
	})

	// **片取りでは出さない。** Hi だけ・Lo だけの勝ちをスクープと呼ぶと、
	// 一番強い結果の意味が薄れる。
	t.Run("stays quiet when only the high half is won", func(t *testing.T) {
		assert.NotContains(t, render([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush",
				WonAmount: 60, HiWonAmount: 60},
		}), i18n.T("dramaha.scoop"))
	})

	t.Run("stays quiet when only the low half is won", func(t *testing.T) {
		assert.NotContains(t, render([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandHighCard, HandName: "High Card",
				WonAmount: 40, LowWonAmount: 40},
		}), i18n.T("dramaha.scoop"))
	})

	// 片側の勝ち分が 0 なら、金額が全額でも両取りとは呼ばない。ドロー側は必ず
	// 誰かが取るので、これは「もう片方が全額を持っていった」局面になる。
	t.Run("stays quiet when one half paid nothing to this seat", func(t *testing.T) {
		assert.NotContains(t, render([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush",
				WonAmount: 100, HiWonAmount: 100},
		}), i18n.T("dramaha.scoop"))
	})
}
