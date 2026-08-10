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

func TestFiveCardStudCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.FiveCardStudCuiPresenter)

	t.Run("initial state with door and hole cards", func(t *testing.T) {
		s, players := makeFiveCardStudForPresenter()
		s.SetPhase(domain.FiveCardStudPhaseSecondStreet)
		players[0].AddHoleCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		players[0].AddDoorCard(domain.NewCard(domain.CardDesignClover, 5, false))

		result := p.Output(s, nil)
		assert.Contains(t, result, "Five Card Stud")
		assert.Contains(t, result, "ディーラー:")
		assert.Contains(t, result, "ポット:")
		assert.Contains(t, result, "あなた")
		assert.Contains(t, result, "ホールカード:")
		assert.Contains(t, result, "♠10")
		assert.Contains(t, result, "ドアカード:")
		assert.Contains(t, result, "♣5")
	})

	t.Run("ante and bring-in info displayed", func(t *testing.T) {
		s, _ := makeFiveCardStudForPresenter()
		s.SetPhase(domain.FiveCardStudPhaseSecondStreet)

		result := p.Output(s, nil)
		assert.Contains(t, result, "Ante:1")
		assert.Contains(t, result, "BringIn:2")
		assert.Contains(t, result, "SmallBet:5")
		assert.Contains(t, result, "BigBet:10")
	})

	t.Run("CPU door cards always visible", func(t *testing.T) {
		s, players := makeFiveCardStudForPresenter()
		s.SetPhase(domain.FiveCardStudPhaseSecondStreet)
		players[1].AddDoorCard(domain.NewCard(domain.CardDesignHeart, 7, false))

		result := p.Output(s, nil)
		assert.Contains(t, result, "ドアカード:")
		assert.Contains(t, result, "♥7")
	})

	t.Run("CPU hole cards not shown", func(t *testing.T) {
		s, players := makeFiveCardStudForPresenter()
		s.SetPhase(domain.FiveCardStudPhaseSecondStreet)
		players[1].AddHoleCard(domain.NewCard(domain.CardDesignSpade, 3, false))

		result := p.Output(s, nil)
		assert.Contains(t, result, "CPU 1")
		assert.NotContains(t, result, "ホールカード:")
	})

	t.Run("folded player", func(t *testing.T) {
		s, players := makeFiveCardStudForPresenter()
		s.SetPhase(domain.FiveCardStudPhaseSecondStreet)
		players[1].SetFolded(true)

		result := p.Output(s, nil)
		assert.Contains(t, result, "[フォールド]")
	})

	t.Run("all-in player", func(t *testing.T) {
		s, players := makeFiveCardStudForPresenter()
		s.SetPhase(domain.FiveCardStudPhaseSecondStreet)
		players[1].SetAllIn(true)

		result := p.Output(s, nil)
		assert.Contains(t, result, "[オールイン]")
	})

	t.Run("player with current bet", func(t *testing.T) {
		s, players := makeFiveCardStudForPresenter()
		s.SetPhase(domain.FiveCardStudPhaseSecondStreet)
		players[0].SetCurrentBet(50)

		result := p.Output(s, nil)
		assert.Contains(t, result, "ベット:50")
	})

	t.Run("CPU actions displayed", func(t *testing.T) {
		s, _ := makeFiveCardStudForPresenter()
		s.SetPhase(domain.FiveCardStudPhaseSecondStreet)
		s.SetCpuActions([]domain.FiveCardStudCpuAction{
			{PlayerIdx: 1, Action: domain.FiveCardStudActionCall, Amount: 0},
			{PlayerIdx: 2, Action: domain.FiveCardStudActionRaise, Amount: 30},
		})

		result := p.Output(s, nil)
		assert.Contains(t, result, "[CPU行動]")
		assert.Contains(t, result, "Player 1: コール")
		assert.Contains(t, result, "Player 2: レイズ")
		assert.Contains(t, result, "(30)")
	})

	t.Run("no CPU actions hides section", func(t *testing.T) {
		s, _ := makeFiveCardStudForPresenter()
		s.SetPhase(domain.FiveCardStudPhaseSecondStreet)
		s.SetCpuActions([]domain.FiveCardStudCpuAction{})

		result := p.Output(s, nil)
		assert.NotContains(t, result, "[CPU行動]")
	})

	t.Run("showdown results with human winner", func(t *testing.T) {
		s, _ := makeFiveCardStudForPresenter()
		s.SetPhase(domain.FiveCardStudPhaseEnd)
		s.SetRoundResults([]domain.FiveCardStudResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100},
		})

		result := p.Output(s, nil)
		assert.Contains(t, result, "[結果]")
		assert.Contains(t, result, "あなた: フラッシュ")
		assert.Contains(t, result, "100チップ獲得")
	})

	t.Run("showdown results with kickers", func(t *testing.T) {
		s, _ := makeFiveCardStudForPresenter()
		s.SetPhase(domain.FiveCardStudPhaseEnd)
		s.SetRoundResults([]domain.FiveCardStudResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandOnePair, HandName: "One Pair", Kickers: []int{14, 12, 10}, WonAmount: 100},
		})

		result := p.Output(s, nil)
		assert.Contains(t, result, "あなた: ワンペア (キッカー: A, Q, 10)")
	})

	t.Run("showdown results with CPU winner", func(t *testing.T) {
		s, _ := makeFiveCardStudForPresenter()
		s.SetPhase(domain.FiveCardStudPhaseEnd)
		s.SetRoundResults([]domain.FiveCardStudResult{
			{PlayerIdx: 1, HandRank: domain.PokerHandOnePair, HandName: "One Pair", Kickers: []int{13, 12, 11}, WonAmount: 50},
		})

		result := p.Output(s, nil)
		assert.Contains(t, result, "CPU 1: ワンペア (キッカー: K, Q, J)")
	})

	t.Run("results not shown in non-end phase", func(t *testing.T) {
		s, _ := makeFiveCardStudForPresenter()
		s.SetPhase(domain.FiveCardStudPhaseSecondStreet)
		s.SetRoundResults([]domain.FiveCardStudResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100},
		})

		result := p.Output(s, nil)
		assert.NotContains(t, result, "[結果]")
	})

	t.Run("error message displayed", func(t *testing.T) {
		s, _ := makeFiveCardStudForPresenter()
		s.SetPhase(domain.FiveCardStudPhaseSecondStreet)

		result := p.Output(s, errors.New("test error"))
		assert.Contains(t, result, "test error")
	})

	t.Run("game end message displayed", func(t *testing.T) {
		s, _ := makeFiveCardStudForPresenter()
		s.SetGameEndFlag(true)

		result := p.Output(s, nil)
		assert.Contains(t, result, "ゲーム終了")
	})

	t.Run("mucked result displayed", func(t *testing.T) {
		s, _ := makeFiveCardStudForPresenter()
		s.SetPhase(domain.FiveCardStudPhaseEnd)
		s.SetRoundResults([]domain.FiveCardStudResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandOnePair, HandName: "One Pair", WonAmount: 0, Mucked: true},
			{PlayerIdx: 1, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100},
		})

		result := p.Output(s, nil)
		assert.Contains(t, result, "あなた: マック")
		assert.NotContains(t, result, "あなた: ワンペア")
	})

	t.Run("results shown in showdown phase", func(t *testing.T) {
		s, _ := makeFiveCardStudForPresenter()
		s.SetPhase(domain.FiveCardStudPhaseShowdown)
		s.SetRoundResults([]domain.FiveCardStudResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100},
		})

		result := p.Output(s, nil)
		assert.Contains(t, result, "[結果]")
		assert.Contains(t, result, "あなた: フラッシュ")
	})

	t.Run("HUD stats shown when totalHands > 0", func(t *testing.T) {
		s, players := makeFiveCardStudForPresenter()
		s.SetPhase(domain.FiveCardStudPhaseSecondStreet)
		players[0].SetTotalHands(2)
		players[0].SetVPIPCount(1)

		result := p.Output(s, nil)
		assert.Contains(t, result, "VPIP:50% PFR:0% 3Bet:0% AF:-")
	})

	t.Run("HUD stats not shown when totalHands is 0", func(t *testing.T) {
		s, _ := makeFiveCardStudForPresenter()
		s.SetPhase(domain.FiveCardStudPhaseSecondStreet)

		result := p.Output(s, nil)
		assert.NotContains(t, result, "VPIP:")
	})

	t.Run("tournament mode header shown when enabled", func(t *testing.T) {
		s, _ := makeFiveCardStudForPresenter()
		s.SetPhase(domain.FiveCardStudPhaseSecondStreet)
		cfg := domain.FiveCardStudConfig{
			Ante:           5,
			BringIn:        10,
			SmallBet:       10,
			BigBet:         20,
			InitChips:      1000,
			TournamentMode: true,
			AnteLevelHands: 5,
			AnteMultiplier: 200,
		}
		s.SetConfig(cfg)
		s.SetHandCount(3)

		result := p.Output(s, nil)
		assert.Contains(t, result, "トーナメント ハンド#3 Ante:5 BringIn:10 (レベルアップ:5ハンド毎)")
	})

	t.Run("tournament mode header not shown when disabled", func(t *testing.T) {
		s, _ := makeFiveCardStudForPresenter()
		s.SetPhase(domain.FiveCardStudPhaseSecondStreet)

		result := p.Output(s, nil)
		assert.NotContains(t, result, "トーナメント")
	})
}

func TestFiveCardStudCuiPresenter_Output_BettingLimitDisplay(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.FiveCardStudCuiPresenter)

	t.Run("displays Fixed limit", func(t *testing.T) {
		s, _ := makeFiveCardStudForPresenter()
		s.SetPhase(domain.FiveCardStudPhaseSecondStreet)
		result := p.Output(s, nil)
		assert.Contains(t, result, "Fixed")
	})
}

func TestFiveCardStudCuiPresenter_Output_TableSize(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.FiveCardStudCuiPresenter)

	t.Run("displays 4-max", func(t *testing.T) {
		s, _ := makeFiveCardStudForPresenter()
		s.SetPhase(domain.FiveCardStudPhaseSecondStreet)
		result := p.Output(s, nil)
		assert.Contains(t, result, "テーブル: 4-max")
	})
}

func TestFiveCardStudCuiPresenter_Output_RebuyAddon(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.FiveCardStudCuiPresenter)

	t.Run("tournament mode with rebuy enabled shows rebuy info", func(t *testing.T) {
		s, _ := makeFiveCardStudForPresenter()
		s.SetPhase(domain.FiveCardStudPhaseSecondStreet)
		cfg := domain.FiveCardStudConfig{
			Ante:             5,
			BringIn:          10,
			SmallBet:         10,
			BigBet:           20,
			InitChips:        1000,
			TournamentMode:   true,
			AnteLevelHands:   5,
			AnteMultiplier:   200,
			RebuyEnabled:     true,
			RebuyChips:       1000,
			RebuyMaxCount:    3,
			RebuyPeriodHands: 20,
		}
		s.SetConfig(cfg)
		s.SetHandCount(2)

		result := p.Output(s, nil)
		assert.Contains(t, result, "リバイ: 1000チップ (最大3回, 20ハンド目まで)")
	})

	t.Run("tournament mode with addon enabled shows addon info", func(t *testing.T) {
		s, _ := makeFiveCardStudForPresenter()
		s.SetPhase(domain.FiveCardStudPhaseSecondStreet)
		cfg := domain.FiveCardStudConfig{
			Ante:           5,
			BringIn:        10,
			SmallBet:       10,
			BigBet:         20,
			InitChips:      1000,
			TournamentMode: true,
			AnteLevelHands: 5,
			AnteMultiplier: 200,
			AddonEnabled:   true,
			AddonChips:     1500,
			AddonAfterHand: 20,
		}
		s.SetConfig(cfg)
		s.SetHandCount(2)

		result := p.Output(s, nil)
		assert.Contains(t, result, "アドオン: 1500チップ (20ハンド目に提供)")
	})

	t.Run("rebuy phase type 1 shows rebuy prompt", func(t *testing.T) {
		s, _ := makeFiveCardStudForPresenter()
		cfg := domain.FiveCardStudConfig{
			Ante:             5,
			BringIn:          10,
			SmallBet:         10,
			BigBet:           20,
			InitChips:        1000,
			TournamentMode:   true,
			AnteLevelHands:   5,
			AnteMultiplier:   200,
			RebuyEnabled:     true,
			RebuyChips:       1000,
			RebuyMaxCount:    3,
			RebuyPeriodHands: 20,
		}
		s.SetConfig(cfg)
		s.SetPhase(domain.FiveCardStudPhaseRebuy)
		s.SetRebuyPhaseType(1)
		s.SetRebuyCounts([]int{1, 0, 0, 0})

		result := p.Output(s, nil)
		assert.Contains(t, result, "リバイしますか?")
		assert.Contains(t, result, "1000チップ")
		assert.Contains(t, result, "1/3回使用済")
	})

	t.Run("rebuy phase type 2 shows addon prompt", func(t *testing.T) {
		s, _ := makeFiveCardStudForPresenter()
		cfg := domain.FiveCardStudConfig{
			Ante:           5,
			BringIn:        10,
			SmallBet:       10,
			BigBet:         20,
			InitChips:      1000,
			TournamentMode: true,
			AnteLevelHands: 5,
			AnteMultiplier: 200,
			AddonEnabled:   true,
			AddonChips:     1500,
			AddonAfterHand: 20,
		}
		s.SetConfig(cfg)
		s.SetPhase(domain.FiveCardStudPhaseRebuy)
		s.SetRebuyPhaseType(2)

		result := p.Output(s, nil)
		assert.Contains(t, result, "アドオンしますか?")
		assert.Contains(t, result, "1500チップ")
	})
}

func TestFiveCardStudCuiPresenter_Output_Muck(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.FiveCardStudCuiPresenter)

	t.Run("muck prompt displayed when available", func(t *testing.T) {
		s, _ := makeFiveCardStudForPresenter()
		s.SetPhase(domain.FiveCardStudPhaseShowdown)
		s.SetRoundResults([]domain.FiveCardStudResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandOnePair, HandName: "One Pair", WonAmount: 0},
			{PlayerIdx: 1, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100},
		})

		result := p.Output(s, nil)
		assert.Contains(t, result, "マックしますか? (m=マック / sh=ショー)")
	})
}

func TestFiveCardStudCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.FiveCardStudCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		mockGame := new(interfaces.MockFiveCardStudGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "raise", Detail: "raised to 100"},
		}
		mockGame.On("GetGameEndFlag").Return(true)
		mockGame.On("GetActionLog").Return(entries)

		result := p.ActionLogOutput(mockGame)
		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "raise")
		mockGame.AssertExpectations(t)
	})

	t.Run("game not ended", func(t *testing.T) {
		mockGame := new(interfaces.MockFiveCardStudGame)
		mockGame.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(mockGame)
		assert.Contains(t, result, "棋譜はありません")
		mockGame.AssertExpectations(t)
	})
}
