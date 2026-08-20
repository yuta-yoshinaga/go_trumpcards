package presenter_test

import (
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// setupOldMaidCuiTest creates an OldMaid game with standard CPU card setup (player[1] has HEART 5, players[2,3] finished).
func setupOldMaidCuiTest() (*domain.OldMaid, []*domain.OldMaidPlayer) {
	tc := domain.NewTrumpCards(1)
	players := []*domain.OldMaidPlayer{
		domain.NewOldMaidPlayer(true),
		domain.NewOldMaidPlayer(false),
		domain.NewOldMaidPlayer(false),
		domain.NewOldMaidPlayer(false),
	}
	om := domain.NewOldMaid(tc, players)
	players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
	players[2].SetIsFinished(true)
	players[3].SetIsFinished(true)
	return om, players
}

func TestOldMaidCuiPresenter_Method(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	top := new(presenter.OldMaidCuiPresenter)

	makePlayers := func() []*domain.OldMaidPlayer {
		return []*domain.OldMaidPlayer{
			domain.NewOldMaidPlayer(true),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
		}
	}

	t.Run("success Output initial state no draw no game end", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makePlayers()
		om := domain.NewOldMaid(tc, players)
		// Manually set cards
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 2, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		players[2].SetIsFinished(true)
		players[3].AddCard(domain.NewCard(domain.CardDesignDiamond, 4, false))
		expected := "==========\nOld Maid (ババ抜き)\n==========\n" +
			"あなた: 2枚\n[0]SPADE 1  [1]CLOVER 2\n" +
			"CPU 1: 1枚\n" +
			"CPU 2: 上がり\n" +
			"CPU 3: 1枚\n" +
			"----------\n" +
			"手番: あなた → CPU 1から引きます\n" +
			"==========\n"
		assert.Equal(t, expected, top.Output(om, nil))
	})

	t.Run("success Output game ended human loses", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makePlayers()
		om := domain.NewOldMaid(tc, players)
		// Setup: player 0 (human, turn=0) has JOKER + SPADE 5 + CLOVER 5
		// player 1 (CPU 1) has HEART 7 — exactly 1 card (deterministic draw at index 0)
		// players 2,3 finished
		players[0].AddCard(domain.NewCard(domain.CardDesignJoker, domain.CardValueJoker, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		players[2].SetIsFinished(true)
		players[3].SetIsFinished(true)
		// PlayerDraw(0): draws card at index 0 from player 1 (HEART 7)
		// player 0 gains HEART 7 → hand: JOKER, SPADE 5, CLOVER 5, HEART 7
		// DiscardPairs: SPADE 5 + CLOVER 5 pair → discarded → player 0: JOKER, HEART 7 (shuffled order)
		// player 1 has 0 cards → finished
		// checkGameEnd: active = {0} → gameEndFlag=true, loserIdx=0
		_ = om.PlayerDraw(0)
		result := top.Output(om, nil)
		// Card display order in player's hand is non-deterministic due to ShuffleCards
		assert.Contains(t, result, "あなた: 2枚")
		assert.Contains(t, result, "JOKER")
		assert.Contains(t, result, "HEART 7")
		assert.Contains(t, result, "CPU 1: 上がり")
		assert.Contains(t, result, "CPU 2: 上がり")
		assert.Contains(t, result, "CPU 3: 上がり")
		assert.Contains(t, result, "あなたがCPU 1から1枚引きました (HEART 7)。1組捨てました")
		assert.Contains(t, result, "ゲーム終了！ あなたの負け！")
	})

	t.Run("success Output game ended cpu loses", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makePlayers()
		om := domain.NewOldMaid(tc, players)
		// player 0 (human, turn=0): SPADE 3 (1 card)
		// player 1 (CPU 1): CLOVER 3 (1 card, next active) → deterministic draw (index always 0)
		// player 2 (CPU 2): JOKER (1 card) → will be the loser
		// player 3: finished
		// PlayerDraw(0): player 0 draws CLOVER 3, forms pair with SPADE 3 → both players 0 and 1 finish
		// active = {2} → gameEndFlag=true, loserIdx=2
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 3, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignJoker, domain.CardValueJoker, false))
		players[3].SetIsFinished(true)
		_ = om.PlayerDraw(0)
		expected := "==========\nOld Maid (ババ抜き)\n==========\n" +
			"あなた: 上がり\n" +
			"CPU 1: 上がり\n" +
			"CPU 2: 1枚\n" +
			"CPU 3: 上がり\n" +
			"----------\n" +
			"あなたがCPU 1から1枚引きました (CLOVER 3)。1組捨てました\n" +
			"[引き履歴]\n" +
			"1. あなたがCPU 1から引いた (1組捨て) [あなた上がり] [CPU 1上がり]\n" +
			"ゲーム終了！ CPU 2の負け！\n" +
			"==========\n"
		assert.Equal(t, expected, top.Output(om, nil))
	})

	t.Run("success Output human zero cards not finished", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makePlayers()
		om := domain.NewOldMaid(tc, players)
		// human player has 0 cards but is not marked finished
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		players[2].SetIsFinished(true)
		players[3].SetIsFinished(true)
		result := top.Output(om, nil)
		assert.Contains(t, result, "あなた: 0枚")
		assert.Contains(t, result, "CPU 1: 1枚")
		assert.Contains(t, result, "CPU 2: 上がり")
		assert.Contains(t, result, "CPU 3: 上がり")
	})

	t.Run("success Output cpu actions drawn card not revealed", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		// Use all CPU players to enable CpuDraw
		cpuPlayers := []*domain.OldMaidPlayer{
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
		}
		om := domain.NewOldMaid(tc, cpuPlayers)
		// Player 0: JOKER
		// Player 1: SPADE 5 (1 card, deterministic draw at index 0)
		// Players 2, 3: finished
		cpuPlayers[0].AddCard(domain.NewCard(domain.CardDesignJoker, domain.CardValueJoker, false))
		cpuPlayers[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		cpuPlayers[2].SetIsFinished(true)
		cpuPlayers[3].SetIsFinished(true)

		// Player 0 draws SPADE 5, no pair → keeps it
		_ = om.CpuDraw()

		result := top.Output(om, nil)
		// CPU action should show who drew from whom but NOT which card was drawn
		assert.Contains(t, result, "[CPUの行動]")
		assert.Contains(t, result, "CPU 0がCPU 1から1枚引きました")
		assert.NotContains(t, result, "SPADE 5")
	})

	t.Run("success Output displays error message", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makePlayers()
		om := domain.NewOldMaid(tc, players)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		players[2].SetIsFinished(true)
		players[3].SetIsFinished(true)
		result := top.Output(om, domain.ErrNotHumanTurn)
		assert.Contains(t, result, "not human player's turn")
	})

	t.Run("success Output cpu actions with discard does not reveal drawn card", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		cpuPlayers := []*domain.OldMaidPlayer{
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
		}
		om := domain.NewOldMaid(tc, cpuPlayers)
		// Player 0: SPADE 10
		// Player 1: CLOVER 10 (1 card, deterministic draw at index 0)
		// Players 2, 3: finished
		cpuPlayers[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		cpuPlayers[1].AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
		cpuPlayers[2].SetIsFinished(true)
		cpuPlayers[3].SetIsFinished(true)

		// Player 0 draws CLOVER 10, discards pair SPADE 10 + CLOVER 10
		_ = om.CpuDraw()

		result := top.Output(om, nil)
		assert.Contains(t, result, "[CPUの行動]")
		assert.Contains(t, result, "CPU 0がCPU 1から1枚引きました。1組捨てました")
		// Drawn card must not appear even when a pair was discarded
		assert.NotContains(t, result, "CLOVER 10")
	})

	t.Run("success Output getCardStr all designs", func(t *testing.T) {
		om, players := setupOldMaidCuiTest()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 2, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignDiamond, 4, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignJoker, domain.CardValueJoker, false))
		result := top.Output(om, nil)
		assert.Contains(t, result, "SPADE 1")
		assert.Contains(t, result, "CLOVER 2")
		assert.Contains(t, result, "HEART 3")
		assert.Contains(t, result, "DIAMOND 4")
		assert.Contains(t, result, "JOKER")
	})

	t.Run("success Output targetIdx negative no draw target", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makePlayers()
		om := domain.NewOldMaid(tc, players)
		// Manually set all players as finished
		players[0].SetIsFinished(true)
		players[1].SetIsFinished(true)
		players[2].SetIsFinished(true)
		players[3].SetIsFinished(true)
		// Force gameEndFlag to false to exercise the targetIdx < 0 branch
		om.SetGameEndFlag(false)
		result := top.Output(om, nil)
		assert.Contains(t, result, "手番: あなた\n")
		assert.NotContains(t, result, "から引きます")
	})

	t.Run("success Output getCardStr nil and unknown design", func(t *testing.T) {
		om, players := setupOldMaidCuiTest()
		players[0].AddCard(nil)
		players[0].AddCard(domain.NewCard(99, 1, false))
		result := top.Output(om, nil)
		assert.Contains(t, result, "??")
		assert.Contains(t, result, "UNKNOWN")
	})

	t.Run("success Output getPlayerName nil player in human action", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		cpuPlayers := []*domain.OldMaidPlayer{
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
		}
		om := domain.NewOldMaid(tc, cpuPlayers)
		cpuPlayers[0].AddCard(domain.NewCard(domain.CardDesignJoker, domain.CardValueJoker, false))
		cpuPlayers[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		cpuPlayers[2].SetIsFinished(true)
		cpuPlayers[3].SetIsFinished(true)
		// Simulate draw with invalid player idx to trigger nil player name
		om.SetLastDrawPlayerIdx(99)
		om.SetHasDrawn(true)
		result := top.Output(om, nil)
		assert.Contains(t, result, "UNKNOWN")
	})
}

func TestOldMaidCuiPresenter_MetaAI(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	top := new(presenter.OldMaidCuiPresenter)

	t.Run("metaAI status line shown when profile exists", func(t *testing.T) {
		om, _ := setupOldMaidCuiTest()
		om.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		om.SetHumanProfile(&domain.OldMaidHumanProfile{
			GamesPlayed:     3,
			PositionBuckets: [3]int{4, 2, 6},
			TotalPicks:      12,
		})
		result := top.Output(om, nil)
		assert.Contains(t, result, "[メタAI]")
		assert.Contains(t, result, "適応中")
		assert.Contains(t, result, "ゲーム数: 3")
		// EdgePickRate = (4/12 + 6/12)*100 ≈ 83%
		assert.Contains(t, result, "端ピック率: 83%")
	})

	t.Run("no metaAI line when profile is nil", func(t *testing.T) {
		om, _ := setupOldMaidCuiTest()
		om.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		// No SetHumanProfile → profile is nil
		result := top.Output(om, nil)
		assert.NotContains(t, result, "[メタAI]")
	})
}

func TestOldMaidCuiPresenter_DrawHistory(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	top := new(presenter.OldMaidCuiPresenter)

	t.Run("no history section when empty", func(t *testing.T) {
		om, _ := setupOldMaidCuiTest()
		om.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		result := top.Output(om, nil)
		assert.NotContains(t, result, "[引き履歴]")
	})

	t.Run("history section shown after draw", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := []*domain.OldMaidPlayer{
			domain.NewOldMaidPlayer(true),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
		}
		om := domain.NewOldMaid(tc, players)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 7, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
		players[3].SetIsFinished(true)
		_ = om.PlayerDraw(0)
		result := top.Output(om, nil)
		assert.Contains(t, result, "[引き履歴]")
		assert.Contains(t, result, "1. あなたがCPU 1から引いた")
		assert.Contains(t, result, "[CPU 1上がり]")
	})

	t.Run("history via test helper with discardedPairs and drawerFinished", func(t *testing.T) {
		om, _ := setupOldMaidCuiTest()
		om.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		om.SetDrawHistory([]*domain.OldMaidDrawHistoryEntry{
			{DrawPlayerIdx: 0, DrawFromIdx: 1, DiscardedPairs: 2, DrawerFinished: true, TargetFinished: false},
			{DrawPlayerIdx: 2, DrawFromIdx: 3, DiscardedPairs: 0, DrawerFinished: false, TargetFinished: false},
		})
		result := top.Output(om, nil)
		assert.Contains(t, result, "[引き履歴]")
		assert.Contains(t, result, "1. あなたがCPU 1から引いた (2組捨て) [あなた上がり]")
		assert.Contains(t, result, "2. CPU 2がCPU 3から引いた")
		assert.NotContains(t, result, "2. CPU 2がCPU 3から引いた (")
	})

	t.Run("exactly 10 entries shown in full without omission summary", func(t *testing.T) {
		om, _ := setupOldMaidCuiTest()
		om.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		entries := make([]*domain.OldMaidDrawHistoryEntry, 0, 10)
		for range 10 {
			entries = append(entries, &domain.OldMaidDrawHistoryEntry{DrawPlayerIdx: 0, DrawFromIdx: 1})
		}
		om.SetDrawHistory(entries)
		result := top.Output(om, nil)
		assert.Contains(t, result, "1. あなたがCPU 1から引いた")
		assert.Contains(t, result, "10. あなたがCPU 1から引いた")
		assert.NotContains(t, result, "…他")
	})

	t.Run("over 10 entries shows only latest 10 plus omission summary", func(t *testing.T) {
		om, _ := setupOldMaidCuiTest()
		om.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		entries := make([]*domain.OldMaidDrawHistoryEntry, 0, 13)
		for range 13 {
			entries = append(entries, &domain.OldMaidDrawHistoryEntry{DrawPlayerIdx: 0, DrawFromIdx: 1})
		}
		om.SetDrawHistory(entries)
		result := top.Output(om, nil)
		// 3 oldest entries omitted; entries 4..13 (real 1-based indices) remain.
		assert.Contains(t, result, "…他3件（棋譜は log コマンドで参照可）")
		assert.Equal(t, 10, strings.Count(result, "から引いた"))
		assert.Contains(t, result, "4. あなたがCPU 1から引いた")  // first shown = real index 4
		assert.Contains(t, result, "13. あなたがCPU 1から引いた") // last shown = real index 13
	})
}

func TestOldMaidCuiPresenter_JijiNuki_Header(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	top := new(presenter.OldMaidCuiPresenter)
	tc := domain.NewTrumpCards(1)
	players := []*domain.OldMaidPlayer{
		domain.NewOldMaidPlayer(true),
		domain.NewOldMaidPlayer(false),
		domain.NewOldMaidPlayer(false),
		domain.NewOldMaidPlayer(false),
	}
	om := domain.NewOldMaid(tc, players)
	om.SetConfig(domain.OldMaidConfig{Mode: domain.OldMaidModeJijiNuki})
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
	players[2].SetIsFinished(true)
	players[3].SetIsFinished(true)
	result := top.Output(om, nil)
	assert.Contains(t, result, "Old Maid (ジジ抜き)")
	assert.NotContains(t, result, "Old Maid (ババ抜き)")
}

func TestOldMaidCuiPresenter_Normal_Header(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	top := new(presenter.OldMaidCuiPresenter)
	tc := domain.NewTrumpCards(1)
	players := []*domain.OldMaidPlayer{
		domain.NewOldMaidPlayer(true),
		domain.NewOldMaidPlayer(false),
		domain.NewOldMaidPlayer(false),
		domain.NewOldMaidPlayer(false),
	}
	om := domain.NewOldMaid(tc, players)
	// Default config is Normal
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
	players[2].SetIsFinished(true)
	players[3].SetIsFinished(true)
	result := top.Output(om, nil)
	assert.Contains(t, result, "Old Maid (ババ抜き)")
	assert.NotContains(t, result, "Old Maid (ジジ抜き)")
}

func TestOldMaidCuiPresenter_JijiNuki_GameEnd_ShowsRemovedCard(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	top := new(presenter.OldMaidCuiPresenter)
	tc := domain.NewTrumpCards(0)
	players := []*domain.OldMaidPlayer{
		domain.NewOldMaidPlayer(true),
		domain.NewOldMaidPlayer(false),
		domain.NewOldMaidPlayer(false),
		domain.NewOldMaidPlayer(false),
	}
	om := domain.NewOldMaid(tc, players)
	om.SetConfig(domain.OldMaidConfig{Mode: domain.OldMaidModeJijiNuki})

	// Deterministic setup: Heart 7 was the removed card (explains why Clover 7 is unpairable).
	om.SetRemovedCard(domain.NewCard(domain.CardDesignHeart, 7, false))
	// player[0] holds Spade 5; player[1] holds Clover 5 (will be drawn and paired).
	// player[2] holds Clover 7 (unpaired → loser); player[3] is already finished.
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
	players[2].AddCard(domain.NewCard(domain.CardDesignClover, 7, false))
	players[3].SetIsFinished(true)

	// player[0] draws Clover 5 from player[1] → pair → both finish; player[2] alone → game ends.
	_ = om.PlayerDraw(0)

	assert.True(t, om.GetGameEndFlag())
	assert.Equal(t, 2, om.GetLoserIdx())
	result := top.Output(om, nil)
	assert.Contains(t, result, "（除外カード: HEART 7）")
}

func TestOldMaidCuiPresenter_Normal_GameEnd_NoRemovedCard(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	top := new(presenter.OldMaidCuiPresenter)
	tc := domain.NewTrumpCards(1)
	players := []*domain.OldMaidPlayer{
		domain.NewOldMaidPlayer(true),
		domain.NewOldMaidPlayer(false),
		domain.NewOldMaidPlayer(false),
		domain.NewOldMaidPlayer(false),
	}
	om := domain.NewOldMaid(tc, players)
	// Normal mode, removedCard is nil
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignClover, 3, false))
	players[2].AddCard(domain.NewCard(domain.CardDesignJoker, domain.CardValueJoker, false))
	players[3].SetIsFinished(true)
	_ = om.PlayerDraw(0) // ends game
	result := top.Output(om, nil)
	assert.NotContains(t, result, "（除外カード:")
	assert.Contains(t, result, "ゲーム終了！")
}

func TestOldMaidCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.OldMaidCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		mockGame := new(interfaces.MockOldMaidGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 1, ActionType: "draw", Detail: "drew a card"},
		}
		mockGame.On("GetGameEndFlag").Return(true)
		mockGame.On("GetActionLog").Return(entries)
		// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
		mockGame.On("GetPlayer", mock.Anything).Return(domain.NewOldMaidPlayer(true)).Maybe()

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "draw")
		assert.Contains(t, result, "drew a card")
		mockGame.AssertExpectations(t)
	})

	t.Run("nil_entries", func(t *testing.T) {
		mockGame := new(interfaces.MockOldMaidGame)
		mockGame.On("GetGameEndFlag").Return(true)
		mockGame.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
		// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
		mockGame.On("GetPlayer", mock.Anything).Return(domain.NewOldMaidPlayer(true)).Maybe()

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, "棋譜はありません")
		mockGame.AssertExpectations(t)
	})

	t.Run("game_not_ended", func(t *testing.T) {
		mockGame := new(interfaces.MockOldMaidGame)
		mockGame.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, "棋譜はありません")
		mockGame.AssertExpectations(t)
	})
}
