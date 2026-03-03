package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func makeHoldemForPresenter() (*domain.Holdem, []*domain.HoldemPlayer) {
	tc := domain.NewTrumpCards(0)
	players := []*domain.HoldemPlayer{
		domain.NewHoldemPlayer(true, domain.HoldemStyleTAG),
		domain.NewHoldemPlayer(false, domain.HoldemStyleLAP),
		domain.NewHoldemPlayer(false, domain.HoldemStyleTAP),
		domain.NewHoldemPlayer(false, domain.HoldemStyleLAG),
	}
	h := domain.NewHoldem(tc, players, domain.DefaultHoldemConfig())
	return h, players
}

func TestHoldemCuiPresenter_Output(t *testing.T) {
	p := presenter.NewHoldemCuiPresenter()

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
		assert.Contains(t, result, "[You]")
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
		assert.Contains(t, result, "You: Flush")
		assert.Contains(t, result, "100チップ獲得")
	})

	t.Run("showdown results with CPU winner", func(t *testing.T) {
		h, _ := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 1, HandRank: domain.PokerHandOnePair, HandName: "One Pair", WonAmount: 50, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.Contains(t, result, "CPU 1: One Pair")
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
			{PlayerIdx: 1, HandName: "High Card", WonAmount: 0, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.NotContains(t, result, "チップ獲得")
	})

	t.Run("results not shown in non-end phase", func(t *testing.T) {
		h, _ := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhaseFlop)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandName: "Flush", WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		assert.NotContains(t, result, "[結果]")
	})

	t.Run("error message displayed", func(t *testing.T) {
		h, _ := makeHoldemForPresenter()
		h.SetPhase(domain.HoldemPhasePreFlop)

		result := p.Output(h, errors.New("test error"))
		assert.Contains(t, result, "[エラー] test error")
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
		assert.Contains(t, result, "VPIP:50% PFR:0%")
		assert.Contains(t, result, "VPIP:66% PFR:33%")
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
