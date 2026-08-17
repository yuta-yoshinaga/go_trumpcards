package presenter_test

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestSevenCardStudCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.SevenCardStudCuiPresenter)

	t.Run("initial state with door and hole cards", func(t *testing.T) {
		s, players := makeSevenCardStudForPresenter()
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)
		players[0].AddHoleCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		players[0].AddHoleCard(domain.NewCard(domain.CardDesignHeart, 11, false))
		players[0].AddDoorCard(domain.NewCard(domain.CardDesignClover, 5, false))

		result := p.Output(s, nil)
		assert.Contains(t, result, "Seven Card Stud")
		assert.Contains(t, result, "ディーラー:")
		assert.Contains(t, result, "ポット:")
		assert.Contains(t, result, "あなた")
		assert.Contains(t, result, "ホールカード:")
		assert.Contains(t, result, "♠10")
		assert.Contains(t, result, "♥11")
		assert.Contains(t, result, "ドアカード:")
		assert.Contains(t, result, "♣5")
	})

	// **Web は常時「いまの最善役」を出しているのに、ハイ戦の CUI はショーダウン
	// まで役名を一切出していなかった (#4695)。**3rd〜7th street の間ずっと
	// 自分の手が何に達しているか分からない。
	t.Run("high game shows the human's current best hand", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := []*domain.SevenCardStudPlayer{
			domain.NewSevenCardStudPlayer(true, domain.HoldemStyleTAG),
			domain.NewSevenCardStudPlayer(false, domain.HoldemStyleLAP),
		}
		s := domain.NewSevenCardStud(tc, players, domain.DefaultSevenCardStudConfig())
		s.SetPhase(domain.SevenCardStudPhaseFifthStreet)
		// フラッシュ: ♠2 ♠5 ♠7 ♠9 ♠13
		players[0].AddHoleCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		players[0].AddHoleCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[0].AddDoorCard(domain.NewCard(domain.CardDesignSpade, 7, false))
		players[0].AddDoorCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		players[0].AddDoorCard(domain.NewCard(domain.CardDesignSpade, 13, false))

		result := p.Output(s, nil)
		assert.Contains(t, result, "現在の最善役")
		assert.Contains(t, result, "フラッシュ")
	})

	// **5枚に満たないうちは出さない。**3rd street は3枚しかなく、そこで
	// 「ハイカード」と出しても情報が無い。
	t.Run("high game shows nothing before five cards", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := []*domain.SevenCardStudPlayer{
			domain.NewSevenCardStudPlayer(true, domain.HoldemStyleTAG),
			domain.NewSevenCardStudPlayer(false, domain.HoldemStyleLAP),
		}
		s := domain.NewSevenCardStud(tc, players, domain.DefaultSevenCardStudConfig())
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)
		players[0].AddHoleCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		players[0].AddHoleCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		players[0].AddDoorCard(domain.NewCard(domain.CardDesignClover, 7, false))

		assert.NotContains(t, p.Output(s, nil), "現在の最善役")
	})

	// **Razz では出さない。**ロー狙いなのにハイ役を出すと逆の助言になる。
	t.Run("razz does not show the high-hand line", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := []*domain.SevenCardStudPlayer{
			domain.NewSevenCardStudPlayer(true, domain.HoldemStyleTAG),
			domain.NewSevenCardStudPlayer(false, domain.HoldemStyleLAP),
		}
		s := domain.NewRazz(tc, players, domain.DefaultSevenCardStudConfig())
		s.SetPhase(domain.SevenCardStudPhaseFifthStreet)
		players[0].AddHoleCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		players[0].AddHoleCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[0].AddDoorCard(domain.NewCard(domain.CardDesignSpade, 7, false))
		players[0].AddDoorCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		players[0].AddDoorCard(domain.NewCard(domain.CardDesignSpade, 13, false))

		assert.NotContains(t, p.Output(s, nil), "現在の最善役")
	})

	t.Run("razz mode shows the human's current best low", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := []*domain.SevenCardStudPlayer{
			domain.NewSevenCardStudPlayer(true, domain.HoldemStyleTAG),
			domain.NewSevenCardStudPlayer(false, domain.HoldemStyleLAP),
		}
		s := domain.NewRazz(tc, players, domain.DefaultSevenCardStudConfig())
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)
		// A completed low: 2-3-4-5-7.
		players[0].AddHoleCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		players[0].AddHoleCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		players[0].AddDoorCard(domain.NewCard(domain.CardDesignClover, 4, false))
		players[0].AddDoorCard(domain.NewCard(domain.CardDesignDiamond, 5, false))
		players[0].AddDoorCard(domain.NewCard(domain.CardDesignSpade, 7, false))

		result := p.Output(s, nil)
		assert.Contains(t, result, "現在のベストロー")
	})

	t.Run("razz mode shows incomplete low with fewer than five cards", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := []*domain.SevenCardStudPlayer{
			domain.NewSevenCardStudPlayer(true, domain.HoldemStyleTAG),
			domain.NewSevenCardStudPlayer(false, domain.HoldemStyleLAP),
		}
		s := domain.NewRazz(tc, players, domain.DefaultSevenCardStudConfig())
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)
		players[0].AddHoleCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		players[0].AddDoorCard(domain.NewCard(domain.CardDesignClover, 4, false))

		result := p.Output(s, nil)
		assert.Contains(t, result, "未完成")
	})

	t.Run("action prompt shows call amount on human betting turn", func(t *testing.T) {
		s, players := makeSevenCardStudForPresenter()
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)
		s.SetCurrentTurn(0)
		s.SetLastBet(5)
		players[0].SetCurrentBet(0)

		result := p.Output(s, nil)
		assert.Contains(t, result, "コール 5")
		assert.Contains(t, result, "最低レイズ")
	})

	t.Run("action prompt shows check when nothing to call", func(t *testing.T) {
		s, players := makeSevenCardStudForPresenter()
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)
		s.SetCurrentTurn(0)
		s.SetLastBet(0)
		players[0].SetCurrentBet(0)

		result := p.Output(s, nil)
		assert.Contains(t, result, "チェック可")
	})

	t.Run("no action prompt on a CPU betting turn", func(t *testing.T) {
		s := func() *domain.SevenCardStud { s, _ := makeSevenCardStudForPresenter(); return s }()
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)
		s.SetCurrentTurn(1)
		s.SetLastBet(5)

		result := p.Output(s, nil)
		assert.NotContains(t, result, "あなたの手番")
	})

	t.Run("ante and bring-in info displayed", func(t *testing.T) {
		s, _ := makeSevenCardStudForPresenter()
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)

		result := p.Output(s, nil)
		assert.Contains(t, result, "Ante:1")
		assert.Contains(t, result, "BringIn:2")
		assert.Contains(t, result, "SmallBet:5")
		assert.Contains(t, result, "BigBet:10")
	})

	t.Run("CPU door cards always visible", func(t *testing.T) {
		s, players := makeSevenCardStudForPresenter()
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)
		players[1].AddDoorCard(domain.NewCard(domain.CardDesignHeart, 7, false))

		result := p.Output(s, nil)
		assert.Contains(t, result, "ドアカード:")
		assert.Contains(t, result, "♥7")
	})

	t.Run("CPU hole cards not shown", func(t *testing.T) {
		s, players := makeSevenCardStudForPresenter()
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)
		players[1].AddHoleCard(domain.NewCard(domain.CardDesignSpade, 3, false))

		result := p.Output(s, nil)
		// CPU player name is shown but ホールカード section should not appear for CPU
		assert.Contains(t, result, "CPU 1")
		// No ホールカード section for CPU (human has no hole cards either in this test)
		assert.NotContains(t, result, "ホールカード:")
	})

	t.Run("folded player", func(t *testing.T) {
		s, players := makeSevenCardStudForPresenter()
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)
		players[1].SetFolded(true)

		result := p.Output(s, nil)
		assert.Contains(t, result, "[フォールド]")
	})

	t.Run("all-in player", func(t *testing.T) {
		s, players := makeSevenCardStudForPresenter()
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)
		players[1].SetAllIn(true)

		result := p.Output(s, nil)
		assert.Contains(t, result, "[オールイン]")
	})

	t.Run("player with current bet", func(t *testing.T) {
		s, players := makeSevenCardStudForPresenter()
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)
		players[0].SetCurrentBet(50)

		result := p.Output(s, nil)
		assert.Contains(t, result, "ベット:50")
	})

	t.Run("CPU actions displayed", func(t *testing.T) {
		s, _ := makeSevenCardStudForPresenter()
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)
		s.SetCpuActions([]domain.SevenCardStudCpuAction{
			{PlayerIdx: 1, Action: domain.SevenCardStudActionCall, Amount: 0},
			{PlayerIdx: 2, Action: domain.SevenCardStudActionRaise, Amount: 30},
		})

		result := p.Output(s, nil)
		assert.Contains(t, result, "[CPU行動]")
		assert.Contains(t, result, "Player 1: コール")
		assert.Contains(t, result, "Player 2: レイズ")
		assert.Contains(t, result, "(30)")
	})

	t.Run("no CPU actions hides section", func(t *testing.T) {
		s, _ := makeSevenCardStudForPresenter()
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)
		s.SetCpuActions([]domain.SevenCardStudCpuAction{})

		result := p.Output(s, nil)
		assert.NotContains(t, result, "[CPU行動]")
	})

	t.Run("showdown results with human winner", func(t *testing.T) {
		s, _ := makeSevenCardStudForPresenter()
		s.SetPhase(domain.SevenCardStudPhaseEnd)
		s.SetRoundResults([]domain.SevenCardStudResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100},
		})

		result := p.Output(s, nil)
		assert.Contains(t, result, "[結果]")
		assert.Contains(t, result, "あなた: フラッシュ")
		assert.Contains(t, result, "100チップ獲得")
	})

	t.Run("showdown results with kickers", func(t *testing.T) {
		s, _ := makeSevenCardStudForPresenter()
		s.SetPhase(domain.SevenCardStudPhaseEnd)
		s.SetRoundResults([]domain.SevenCardStudResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandOnePair, HandName: "One Pair", Kickers: []int{14, 12, 10}, WonAmount: 100},
		})

		result := p.Output(s, nil)
		assert.Contains(t, result, "あなた: ワンペア (キッカー: A, Q, 10)")
	})

	t.Run("showdown results with CPU winner", func(t *testing.T) {
		s, _ := makeSevenCardStudForPresenter()
		s.SetPhase(domain.SevenCardStudPhaseEnd)
		s.SetRoundResults([]domain.SevenCardStudResult{
			{PlayerIdx: 1, HandRank: domain.PokerHandOnePair, HandName: "One Pair", Kickers: []int{13, 12, 11}, WonAmount: 50},
		})

		result := p.Output(s, nil)
		assert.Contains(t, result, "CPU 1: ワンペア (キッカー: K, Q, J)")
	})

	t.Run("results not shown in non-end phase", func(t *testing.T) {
		s, _ := makeSevenCardStudForPresenter()
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)
		s.SetRoundResults([]domain.SevenCardStudResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100},
		})

		result := p.Output(s, nil)
		assert.NotContains(t, result, "[結果]")
	})

	t.Run("error message displayed", func(t *testing.T) {
		s, _ := makeSevenCardStudForPresenter()
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)

		result := p.Output(s, errors.New("test error"))
		assert.Contains(t, result, "test error")
	})

	t.Run("game end message displayed", func(t *testing.T) {
		s, _ := makeSevenCardStudForPresenter()
		s.SetGameEndFlag(true)

		result := p.Output(s, nil)
		assert.Contains(t, result, "ゲーム終了")
	})

	t.Run("mucked result displayed", func(t *testing.T) {
		s, _ := makeSevenCardStudForPresenter()
		s.SetPhase(domain.SevenCardStudPhaseEnd)
		s.SetRoundResults([]domain.SevenCardStudResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandOnePair, HandName: "One Pair", WonAmount: 0, Mucked: true},
			{PlayerIdx: 1, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100},
		})

		result := p.Output(s, nil)
		assert.Contains(t, result, "あなた: マック")
		assert.NotContains(t, result, "あなた: ワンペア")
	})

	t.Run("results shown in showdown phase", func(t *testing.T) {
		s, _ := makeSevenCardStudForPresenter()
		s.SetPhase(domain.SevenCardStudPhaseShowdown)
		s.SetRoundResults([]domain.SevenCardStudResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100},
		})

		result := p.Output(s, nil)
		assert.Contains(t, result, "[結果]")
		assert.Contains(t, result, "あなた: フラッシュ")
	})

	t.Run("HUD stats shown when totalHands > 0", func(t *testing.T) {
		s, players := makeSevenCardStudForPresenter()
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)
		players[0].SetTotalHands(2)
		players[0].SetVPIPCount(1)

		result := p.Output(s, nil)
		assert.Contains(t, result, "VPIP:50% PFR:0% 3Bet:0% AF:-")
	})

	t.Run("HUD stats not shown when totalHands is 0", func(t *testing.T) {
		s, _ := makeSevenCardStudForPresenter()
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)

		result := p.Output(s, nil)
		assert.NotContains(t, result, "VPIP:")
	})

	t.Run("tournament mode header shown when enabled", func(t *testing.T) {
		s, _ := makeSevenCardStudForPresenter()
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)
		cfg := domain.SevenCardStudConfig{
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
		s, _ := makeSevenCardStudForPresenter()
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)

		result := p.Output(s, nil)
		assert.NotContains(t, result, "トーナメント")
	})
}

func TestSevenCardStudCuiPresenter_Output_BettingLimitDisplay(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.SevenCardStudCuiPresenter)

	t.Run("displays Fixed limit", func(t *testing.T) {
		s, _ := makeSevenCardStudForPresenter()
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)
		result := p.Output(s, nil)
		assert.Contains(t, result, "Fixed")
	})
}

func TestSevenCardStudCuiPresenter_Output_TableSize(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.SevenCardStudCuiPresenter)

	t.Run("displays 4-max", func(t *testing.T) {
		s, _ := makeSevenCardStudForPresenter()
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)
		result := p.Output(s, nil)
		assert.Contains(t, result, "テーブル: 4-max")
	})
}

func TestSevenCardStudCuiPresenter_Output_RebuyAddon(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.SevenCardStudCuiPresenter)

	t.Run("tournament mode with rebuy enabled shows rebuy info", func(t *testing.T) {
		s, _ := makeSevenCardStudForPresenter()
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)
		cfg := domain.SevenCardStudConfig{
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
		s, _ := makeSevenCardStudForPresenter()
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)
		cfg := domain.SevenCardStudConfig{
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
		s, _ := makeSevenCardStudForPresenter()
		cfg := domain.SevenCardStudConfig{
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
		s.SetPhase(domain.SevenCardStudPhaseRebuy)
		s.SetRebuyPhaseType(1)
		s.SetRebuyCounts([]int{1, 0, 0, 0})

		result := p.Output(s, nil)
		assert.Contains(t, result, "リバイしますか?")
		assert.Contains(t, result, "1000チップ")
		assert.Contains(t, result, "1/3回使用済")
	})

	t.Run("rebuy phase type 2 shows addon prompt", func(t *testing.T) {
		s, _ := makeSevenCardStudForPresenter()
		cfg := domain.SevenCardStudConfig{
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
		s.SetPhase(domain.SevenCardStudPhaseRebuy)
		s.SetRebuyPhaseType(2)

		result := p.Output(s, nil)
		assert.Contains(t, result, "アドオンしますか?")
		assert.Contains(t, result, "1500チップ")
	})
}

func TestSevenCardStudCuiPresenter_Output_Muck(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.SevenCardStudCuiPresenter)

	t.Run("muck prompt displayed when available", func(t *testing.T) {
		s, _ := makeSevenCardStudForPresenter()
		s.SetPhase(domain.SevenCardStudPhaseShowdown)
		s.SetRoundResults([]domain.SevenCardStudResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandOnePair, HandName: "One Pair", WonAmount: 0},
			{PlayerIdx: 1, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100},
		})

		result := p.Output(s, nil)
		assert.Contains(t, result, "マックしますか? (m=マック / sh=ショー)")
	})
}

func TestSevenCardStudCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.SevenCardStudCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		mockGame := new(interfaces.MockSevenCardStudGame)
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
		mockGame := new(interfaces.MockSevenCardStudGame)
		mockGame.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(mockGame)
		assert.Contains(t, result, "棋譜はありません")
		mockGame.AssertExpectations(t)
	})
}

func TestSevenCardStudCuiPresenter_HintOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.SevenCardStudCuiPresenter)

	c := func(suit, val int) *domain.Card { return domain.NewCard(suit, val, false) }
	set := func(s *domain.SevenCardStud, players []*domain.SevenCardStudPlayer, cards ...*domain.Card) {
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)
		s.SetCurrentTurn(0)
		players[0].ClearCards()
		players[0].AddHoleCard(cards[0])
		players[0].AddHoleCard(cards[1])
		players[0].AddDoorCard(cards[2])
	}

	cases := []struct {
		name      string
		cards     []*domain.Card
		continue_ bool
		reason    string
	}{
		{"pair", []*domain.Card{c(domain.CardDesignSpade, 8), c(domain.CardDesignHeart, 8), c(domain.CardDesignClover, 2)}, true, "sevencardstud.hintReasonPair"},
		{"three flush", []*domain.Card{c(domain.CardDesignSpade, 4), c(domain.CardDesignSpade, 8), c(domain.CardDesignSpade, 12)}, true, "sevencardstud.hintReasonFlush"},
		{"three straight", []*domain.Card{c(domain.CardDesignSpade, 6), c(domain.CardDesignHeart, 7), c(domain.CardDesignClover, 8)}, true, "sevencardstud.hintReasonStraight"},
		{"Q-K-A wheel straight", []*domain.Card{c(domain.CardDesignSpade, 1), c(domain.CardDesignHeart, 12), c(domain.CardDesignClover, 13)}, true, "sevencardstud.hintReasonStraight"},
		{"three high", []*domain.Card{c(domain.CardDesignSpade, 1), c(domain.CardDesignHeart, 13), c(domain.CardDesignClover, 11)}, true, "sevencardstud.hintReasonHigh"},
		{"junk folds", []*domain.Card{c(domain.CardDesignSpade, 2), c(domain.CardDesignHeart, 5), c(domain.CardDesignClover, 9)}, false, "sevencardstud.hintReasonFold"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, players := makeSevenCardStudForPresenter()
			set(s, players, tc.cards...)
			out := p.HintOutput(s)
			assert.Contains(t, out, i18n.T(tc.reason))
			if tc.continue_ {
				assert.Contains(t, out, i18n.T("sevencardstud.hintContinue"))
			} else {
				assert.Contains(t, out, i18n.T("sevencardstud.hintFold"))
			}
		})
	}

	t.Run("no hint on a CPU turn", func(t *testing.T) {
		s, _ := makeSevenCardStudForPresenter()
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)
		s.SetCurrentTurn(1) // CPU
		assert.Contains(t, p.HintOutput(s), i18n.T("sevencardstud.hintNone"))
	})

	t.Run("no hint outside third street", func(t *testing.T) {
		s, _ := makeSevenCardStudForPresenter()
		s.SetPhase(domain.SevenCardStudPhaseFifthStreet)
		s.SetCurrentTurn(0)
		assert.Contains(t, p.HintOutput(s), i18n.T("sevencardstud.hintNone"))
	})

	// **ラズは以前ここで常に「ヒントなし」だった (#4703)。**Web は razzHint.ts で
	// fold/call/raise を助言しているのに、CUI だけ黙っていた。この t.Run は
	// その挙動を「仕様」として固定してしまっていたので、中身を入れ替えた。
	newRazz := func(cards ...*domain.Card) *domain.SevenCardStud {
		players := []*domain.SevenCardStudPlayer{
			domain.NewSevenCardStudPlayer(true, domain.HoldemStyleTAG),
			domain.NewSevenCardStudPlayer(false, domain.HoldemStyleLAP),
		}
		s := domain.NewRazz(domain.NewTrumpCards(0), players, domain.DefaultSevenCardStudConfig())
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)
		s.SetCurrentTurn(0)
		for i, c := range cards {
			if i < 2 {
				players[0].AddHoleCard(c)
			} else {
				players[0].AddDoorCard(c)
			}
		}
		return s
	}
	card := func(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

	// **ラズは役の強弱が逆。**ペアはハイなら続行材料だが、ローでは弱い。
	t.Run("Razz folds a pair instead of continuing on it", func(t *testing.T) {
		out := p.HintOutput(newRazz(
			card(domain.CardDesignSpade, 8), card(domain.CardDesignHeart, 8),
			card(domain.CardDesignClover, 2)))
		assert.Contains(t, out, i18n.T("sevencardstud.hintFold"))
		assert.Contains(t, out, i18n.T("sevencardstud.hintReasonRazzPair"))
		assert.NotContains(t, out, i18n.T("sevencardstud.hintNone"))
	})

	t.Run("Razz raises on three low cards", func(t *testing.T) {
		out := p.HintOutput(newRazz(
			card(domain.CardDesignSpade, 1), card(domain.CardDesignHeart, 3),
			card(domain.CardDesignClover, 5)))
		assert.Contains(t, out, i18n.T("sevencardstud.hintRaise"))
	})

	t.Run("Razz folds a hand made of high cards", func(t *testing.T) {
		out := p.HintOutput(newRazz(
			card(domain.CardDesignSpade, 12), card(domain.CardDesignHeart, 11),
			card(domain.CardDesignClover, 9)))
		assert.Contains(t, out, i18n.T("sevencardstud.hintFold"))
		assert.Contains(t, out, i18n.T("sevencardstud.hintReasonRazzWeak"))
	})

	// **ハイと違って全ストリートで出す。**引くたびにロー札の枚数が変わる。
	t.Run("Razz still advises on fifth street", func(t *testing.T) {
		s := newRazz(
			card(domain.CardDesignSpade, 1), card(domain.CardDesignHeart, 3),
			card(domain.CardDesignClover, 5), card(domain.CardDesignDiamond, 7))
		s.SetPhase(domain.SevenCardStudPhaseFifthStreet)
		assert.NotContains(t, p.HintOutput(s), i18n.T("sevencardstud.hintNone"))
	})

	// **Hi-Lo にハイ専用の基本戦略をそのまま当てていた (#4704)。**ポットの
	// 半分がローに行くので、ハイとして弱くてもロー札がそろっていれば続ける
	// 価値がある。
	newHiLo := func(cards ...*domain.Card) *domain.SevenCardStud {
		players := []*domain.SevenCardStudPlayer{
			domain.NewSevenCardStudPlayer(true, domain.HoldemStyleTAG),
			domain.NewSevenCardStudPlayer(false, domain.HoldemStyleLAP),
		}
		s := domain.NewSevenCardStudHiLo(domain.NewTrumpCards(0), players, domain.DefaultSevenCardStudConfig())
		s.SetPhase(domain.SevenCardStudPhaseFifthStreet)
		s.SetCurrentTurn(0)
		for i, c := range cards {
			if i < 2 {
				players[0].AddHoleCard(c)
			} else {
				players[0].AddDoorCard(c)
			}
		}
		return s
	}

	t.Run("Hi-Lo keeps going on five low cards a high-only read would fold", func(t *testing.T) {
		// 2 3 4 6 8: ハイとしてはハイカードすら無い (ハイ専用判定なら降りる)。
		out := p.HintOutput(newHiLo(
			card(domain.CardDesignSpade, 2), card(domain.CardDesignHeart, 3),
			card(domain.CardDesignClover, 4), card(domain.CardDesignDiamond, 6),
			card(domain.CardDesignSpade, 8)))
		assert.Contains(t, out, i18n.T("sevencardstud.hintContinue"))
		assert.Contains(t, out, i18n.T("sevencardstud.hintReasonHiLoLow"))
	})

	// **ロー札が4枚では足りない。**5枚そろって初めてローが成立する。
	t.Run("Hi-Lo does not claim a low on four low cards", func(t *testing.T) {
		out := p.HintOutput(newHiLo(
			card(domain.CardDesignSpade, 2), card(domain.CardDesignHeart, 3),
			card(domain.CardDesignClover, 4), card(domain.CardDesignDiamond, 6),
			card(domain.CardDesignSpade, 9)))
		assert.NotContains(t, out, i18n.T("sevencardstud.hintReasonHiLoLow"))
	})

	t.Run("Hi-Lo still leads with a pair", func(t *testing.T) {
		out := p.HintOutput(newHiLo(
			card(domain.CardDesignSpade, 7), card(domain.CardDesignHeart, 7),
			card(domain.CardDesignClover, 4), card(domain.CardDesignDiamond, 6),
			card(domain.CardDesignSpade, 8)))
		assert.Contains(t, out, i18n.T("sevencardstud.hintReasonPair"))
	})

	t.Run("Hi-Lo folds a hand with neither a low nor a high card", func(t *testing.T) {
		out := p.HintOutput(newHiLo(
			card(domain.CardDesignSpade, 2), card(domain.CardDesignHeart, 4),
			card(domain.CardDesignClover, 6), card(domain.CardDesignDiamond, 9),
			card(domain.CardDesignSpade, 3)))
		assert.Contains(t, out, i18n.T("sevencardstud.hintFold"))
	})

	t.Run("Razz gives no hint at showdown", func(t *testing.T) {
		s := newRazz(
			card(domain.CardDesignSpade, 1), card(domain.CardDesignHeart, 3),
			card(domain.CardDesignClover, 5))
		s.SetPhase(domain.SevenCardStudPhaseShowdown)
		assert.Contains(t, p.HintOutput(s), i18n.T("sevencardstud.hintNone"))
	})
}

// #5543: Hi-Lo の結果は合計額しか出ておらず、ハイを取ったのかローを取ったのか、
// スクープしたのかが CUI からは読み取れなかった。Web は StudHiLoSplit が
// 3 通りを別バッジで出している。
func TestSevenCardStudCuiPresenter_Output_HiLoSplit(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.SevenCardStudCuiPresenter)

	outputWith := func(hiLo bool, r domain.SevenCardStudResult) string {
		var s *domain.SevenCardStud
		if hiLo {
			cfg := domain.DefaultSevenCardStudConfig()
			s = domain.NewSevenCardStudHiLo(domain.NewTrumpCards(0),
				domain.NewSevenCardStudPlayersForTable(cfg.TableSize), cfg)
		} else {
			s, _ = makeSevenCardStudForPresenter()
		}
		s.SetPhase(domain.SevenCardStudPhaseEnd)
		s.SetRoundResults([]domain.SevenCardStudResult{r})
		return p.Output(s, nil)
	}

	base := domain.SevenCardStudResult{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush"}

	t.Run("splits the pot into the high and low shares", func(t *testing.T) {
		r := base
		r.WonAmount, r.WonLow = 100, 40
		out := outputWith(true, r)
		assert.Contains(t, out, i18n.Tf("sevencardstud.wonSplit",
			"high", strconv.Itoa(60), "low", strconv.Itoa(40)))
	})

	t.Run("says scoop when the same player took both", func(t *testing.T) {
		r := base
		r.WonAmount, r.WonLow = 100, 40
		assert.Contains(t, outputWith(true, r), i18n.T("sevencardstud.wonScoop"))
	})

	// ローだけ取ったときはスクープではない。
	t.Run("is not a scoop when only the low was won", func(t *testing.T) {
		r := base
		r.WonAmount, r.WonLow = 40, 40
		out := outputWith(true, r)
		assert.NotContains(t, out, i18n.T("sevencardstud.wonScoop"))
		assert.Contains(t, out, i18n.Tf("sevencardstud.wonSplit",
			"high", strconv.Itoa(0), "low", strconv.Itoa(40)))
	})

	// **Hi-Lo でないゲームでは何も変わらない。**
	t.Run("leaves plain seven card stud alone", func(t *testing.T) {
		r := base
		r.WonAmount = 100
		out := outputWith(false, r)
		assert.Contains(t, out, i18n.Tf("sevencardstud.wonAmount", "total", "100"))
		assert.NotContains(t, out, strings.SplitN(i18n.T("sevencardstud.wonSplit"), "{{", 2)[0])
	})
}

// #5542: Web は 3rd street でブリングイン (強制ベットを払い最初に動く席) に
// バッジを出すのに、CUI は誰なのかを知る手段が無かった。Razz は「一番強い
// ドアカード」という逆転ルールなので、なおさら判断材料になる。
func TestSevenCardStudCuiPresenter_Output_BringIn(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.SevenCardStudCuiPresenter)

	outputWith := func(phase, bringIn int) string {
		s, players := makeSevenCardStudForPresenter()
		s.SetPhase(phase)
		s.SetBringInPlayerIdx(bringIn)
		players[0].AddDoorCard(domain.NewCard(domain.CardDesignClover, 5, false))
		return p.Output(s, nil)
	}

	line := func(idx int) string {
		return i18n.Tf("sevencardstud.bringInLine", "name", i18n.Tf("cuiPlayerCpu", "idx", strconv.Itoa(idx)))
	}

	out := outputWith(domain.SevenCardStudPhaseThirdStreet, 2)
	assert.Contains(t, out, line(2))

	// **他ストリートでは出さない。**強制ベットは 3rd street だけの話。
	header := strings.SplitN(i18n.T("sevencardstud.bringInLine"), "{{", 2)[0]
	assert.NotContains(t, outputWith(domain.SevenCardStudPhaseFourthStreet, 2), header)
	// 未確定 (-1) のときも出さない。
	assert.NotContains(t, outputWith(domain.SevenCardStudPhaseThirdStreet, -1), header)
}
