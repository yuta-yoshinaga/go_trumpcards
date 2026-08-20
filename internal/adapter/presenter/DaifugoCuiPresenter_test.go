package presenter_test

import (
	"errors"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func makeDaifugoPlayersForPresenter() []*domain.DaifugoPlayer {
	return []*domain.DaifugoPlayer{
		domain.NewDaifugoPlayer(true),
		domain.NewDaifugoPlayer(false),
		domain.NewDaifugoPlayer(false),
		domain.NewDaifugoPlayer(false),
	}
}

// setupDaifugoCuiTest creates a Daifugo game with DefaultDaifugoConfig and standard CPU cards.
func setupDaifugoCuiTest() (*domain.Daifugo, []*domain.DaifugoPlayer) {
	tc := domain.NewTrumpCards(0)
	players := makeDaifugoPlayersForPresenter()
	dg := domain.NewDaifugo(tc, players, domain.DefaultDaifugoConfig())
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
	players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
	players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))
	return dg, players
}

func TestDaifugoCuiPresenter_Method(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	tdp := new(presenter.DaifugoCuiPresenter)

	t.Run("success Output initial state", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDaifugoPlayersForPresenter()
		dg := domain.NewDaifugo(tc, players, domain.DaifugoConfig{})
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := tdp.Output(dg, nil)
		assert.Contains(t, result, "Daifugo (大富豪)")
		assert.Contains(t, result, "あなた: 2枚")
		assert.Contains(t, result, "[0]SPADE 3")
		assert.Contains(t, result, "CPU 1: 1枚")
		assert.Contains(t, result, "場: なし")
		assert.Contains(t, result, "手番: あなた")
	})

	t.Run("success Output with table cards", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDaifugoPlayersForPresenter()
		dg := domain.NewDaifugo(tc, players, domain.DaifugoConfig{})
		// Human plays a 5
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		_ = dg.PlayerPlay([]int{0}) // play 5
		// CPUs all pass (2 > 5 but 1 card vs 1 card needed — actually 2 is strongest, so CPUs WILL play it)
		// Instead just verify Output works after play
		result := tdp.Output(dg, nil)
		assert.Contains(t, result, "Daifugo (大富豪)")
		assert.NotEmpty(t, result)
	})

	t.Run("success Output game ended", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDaifugoPlayersForPresenter()
		dg := domain.NewDaifugo(tc, players, domain.DaifugoConfig{})
		// Simulate game end: finish all CPU players, human has 1 card, game ends
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[1].SetIsFinished(true)
		players[1].SetRank(1)
		players[2].SetIsFinished(true)
		players[2].SetRank(2)
		players[3].SetIsFinished(true)
		players[3].SetRank(3)
		// Play human's last card to end game
		_ = dg.PlayerPlay([]int{0})
		result := tdp.Output(dg, nil)
		assert.Contains(t, result, "ゲーム終了")
	})

	t.Run("success Output shows human action pass", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDaifugoPlayersForPresenter()
		dg := domain.NewDaifugo(tc, players, domain.DaifugoConfig{})
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		_ = dg.PlayerPlay([]int{}) // pass
		result := tdp.Output(dg, nil)
		assert.Contains(t, result, "パスしました")
	})

	t.Run("success Output shows revolution status when active", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDaifugoPlayersForPresenter()
		dg := domain.NewDaifugo(tc, players, domain.DaifugoConfig{})
		// Human plays four 5s → revolution active
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignDiamond, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false)) // spare
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		_ = dg.PlayerPlay([]int{0, 1, 2, 3}) // play four 5s
		result := tdp.Output(dg, nil)
		assert.Contains(t, result, "革命中")
	})

	t.Run("success Output does not show revolution status when not active", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDaifugoPlayersForPresenter()
		dg := domain.NewDaifugo(tc, players, domain.DaifugoConfig{})
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))
		result := tdp.Output(dg, nil)
		assert.NotContains(t, result, "革命中")
	})

	t.Run("success Output shows CPU actions", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDaifugoPlayersForPresenter()
		dg := domain.NewDaifugo(tc, players, domain.DaifugoConfig{})
		// Human plays 2 (strongest) → CPUs pass → back to human
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 4, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		_ = dg.PlayerPlay([]int{0})
		dg.CpuPlay()
		dg.CpuPlay()
		dg.CpuPlay()
		result := tdp.Output(dg, nil)
		assert.Contains(t, result, "[CPUの行動]")
	})

	t.Run("success Output shows error message when lastErr is non-nil", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDaifugoPlayersForPresenter()
		dg := domain.NewDaifugo(tc, players, domain.DaifugoConfig{})
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		testErr := errors.New("test error message")
		result := tdp.Output(dg, testErr)
		assert.Contains(t, result, "test error message")
	})

	t.Run("success Output eleven back active", func(t *testing.T) {
		dg, _ := setupDaifugoCuiTest()
		dg.SetElevenBackActive(true)
		result := tdp.Output(dg, nil)
		assert.Contains(t, result, "11バック")
	})

	t.Run("success Output suit locked all suits", func(t *testing.T) {
		suitTests := []struct {
			suit     int
			expected string
		}{
			{domain.CardDesignSpade, "SPADE"},
			{domain.CardDesignClover, "CLOVER"},
			{domain.CardDesignHeart, "HEART"},
			{domain.CardDesignDiamond, "DIAMOND"},
			{999, "UNKNOWN"},
		}
		for _, st := range suitTests {
			dg, _ := setupDaifugoCuiTest()
			dg.SetSuitLocked(true, st.suit)
			result := tdp.Output(dg, nil)
			assert.Contains(t, result, "スート縛り")
			assert.Contains(t, result, st.expected)
		}
	})

	t.Run("success Output table is sequence", func(t *testing.T) {
		dg, _ := setupDaifugoCuiTest()
		dg.SetTableIsSequence(true)
		result := tdp.Output(dg, nil)
		assert.Contains(t, result, "階段")
	})

	t.Run("success Output exchange actions", func(t *testing.T) {
		dg, _ := setupDaifugoCuiTest()
		dg.SetExchangeActions([]*domain.DaifugoExchangeAction{
			{FromPlayerIdx: 0, ToPlayerIdx: 3, Cards: []*domain.Card{domain.NewCard(domain.CardDesignSpade, 3, false)}},
		})
		result := tdp.Output(dg, nil)
		assert.Contains(t, result, "カード交換")
		assert.Contains(t, result, "あなた")
		assert.Contains(t, result, "CPU 3")
		assert.Contains(t, result, "SPADE 3")
	})

	t.Run("success Output human action with played cards", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDaifugoPlayersForPresenter()
		dg := domain.NewDaifugo(tc, players, domain.DefaultDaifugoConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))
		_ = dg.PlayerPlay([]int{0}) // play 3
		result := tdp.Output(dg, nil)
		assert.Contains(t, result, "を出しました")
	})

	t.Run("success Output getCardStr with all designs", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDaifugoPlayersForPresenter()
		dg := domain.NewDaifugo(tc, players, domain.DefaultDaifugoConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 4, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignDiamond, 6, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignJoker, domain.CardValueJoker, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))
		result := tdp.Output(dg, nil)
		assert.Contains(t, result, "SPADE 3")
		assert.Contains(t, result, "CLOVER 4")
		assert.Contains(t, result, "HEART 5")
		assert.Contains(t, result, "DIAMOND 6")
		assert.Contains(t, result, "JOKER")
	})

	t.Run("success Output getPlayerName nil player", func(t *testing.T) {
		dg, _ := setupDaifugoCuiTest()
		// Set a human action with out-of-bounds player idx to trigger nil player
		dg.SetHumanAction(&domain.DaifugoCpuAction{PlayerIdx: 99, PlayedCards: nil})
		result := tdp.Output(dg, nil)
		assert.Contains(t, result, "UNKNOWN")
	})

	t.Run("success Output rankName all values", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDaifugoPlayersForPresenter()
		dg := domain.NewDaifugo(tc, players, domain.DefaultDaifugoConfig())
		players[0].SetIsFinished(true)
		players[0].SetRank(1)
		players[1].SetIsFinished(true)
		players[1].SetRank(2)
		players[2].SetIsFinished(true)
		players[2].SetRank(3)
		players[3].SetIsFinished(true)
		players[3].SetRank(4)
		result := tdp.Output(dg, nil)
		assert.Contains(t, result, "大富豪")
		assert.Contains(t, result, "富豪")
		assert.Contains(t, result, "平民")
		assert.Contains(t, result, "大貧民")
	})

	t.Run("success Output rankName unknown rank", func(t *testing.T) {
		dg, players := setupDaifugoCuiTest()
		players[0].SetIsFinished(true)
		players[0].SetRank(0) // unknown rank → "不明"
		result := tdp.Output(dg, nil)
		assert.Contains(t, result, "不明")
	})

	t.Run("success Output getCardStr nil and unknown card via exchange", func(t *testing.T) {
		dg, _ := setupDaifugoCuiTest()
		dg.SetExchangeActions([]*domain.DaifugoExchangeAction{
			{FromPlayerIdx: 0, ToPlayerIdx: 1, Cards: []*domain.Card{nil}},
			{FromPlayerIdx: 1, ToPlayerIdx: 0, Cards: []*domain.Card{domain.NewCard(99, 1, false)}},
		})
		result := tdp.Output(dg, nil)
		assert.Contains(t, result, "??")
		assert.Contains(t, result, "UNKNOWN")
	})

	t.Run("success Output CPU action pass in cpuActions", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDaifugoPlayersForPresenter()
		dg := domain.NewDaifugo(tc, players, domain.DaifugoConfig{})
		// Human plays 2 (strongest) → CPUs pass → back to human
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 4, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 4, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		_ = dg.PlayerPlay([]int{0})
		dg.CpuPlay()
		dg.CpuPlay()
		dg.CpuPlay()
		result := tdp.Output(dg, nil)
		assert.Contains(t, result, "パスしました")
	})

	t.Run("success Output shows 7渡し pending action prompt", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDaifugoPlayersForPresenter()
		config := domain.DaifugoConfig{SevenPassEnabled: true}
		dg := domain.NewDaifugo(tc, players, config)
		// Human plays a 7; has spare card → SevenPass pending set
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false)) // spare
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		_ = dg.PlayerPlay([]int{0})
		result := tdp.Output(dg, nil)
		assert.Contains(t, result, "【7渡し】")
	})

	t.Run("success Output shows 10捨て pending action prompt", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDaifugoPlayersForPresenter()
		config := domain.DaifugoConfig{TenDiscardEnabled: true}
		dg := domain.NewDaifugo(tc, players, config)
		// Human plays a 10; has spare card → TenDiscard pending set
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false)) // spare
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		_ = dg.PlayerPlay([]int{0})
		result := tdp.Output(dg, nil)
		assert.Contains(t, result, "【10捨て】")
	})

	t.Run("success Output shows 9リバース badge", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDaifugoPlayersForPresenter()
		dg := domain.NewDaifugo(tc, players, domain.DaifugoConfig{})
		dg.SetReverseDirection(true)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		result := tdp.Output(dg, nil)
		assert.Contains(t, result, "【9リバース】")
	})

	t.Run("success Output shows 連番縛り badge", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDaifugoPlayersForPresenter()
		dg := domain.NewDaifugo(tc, players, domain.DaifugoConfig{})
		dg.SetNumberLocked(true)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		result := tdp.Output(dg, nil)
		assert.Contains(t, result, "【連番縛り】")
	})

	t.Run("success Output shows 階段縛り badge", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDaifugoPlayersForPresenter()
		dg := domain.NewDaifugo(tc, players, domain.DaifugoConfig{})
		dg.SetSequenceLocked(true)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		result := tdp.Output(dg, nil)
		assert.Contains(t, result, "【階段縛り】")
	})

	t.Run("success Output shows 反則上がり indicator for penalized player", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDaifugoPlayersForPresenter()
		dg := domain.NewDaifugo(tc, players, domain.DefaultDaifugoConfig())
		// Set up game end: finish 3 CPUs, human plays last card
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[1].SetIsFinished(true)
		players[1].SetRank(1)
		players[2].SetIsFinished(true)
		players[2].SetRank(2)
		players[3].SetIsFinished(true)
		players[3].SetRank(3)
		_ = dg.PlayerPlay([]int{0}) // human plays last card → game ends
		// Set penalty after game end
		players[0].SetIllegalFinishPenalty(true)
		result := tdp.Output(dg, nil)
		assert.Contains(t, result, "[反則上がり]")
		assert.Contains(t, result, "ゲーム終了")
	})

	t.Run("success Output shows 12ボンバー pending action prompt", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDaifugoPlayersForPresenter()
		config := domain.DaifugoConfig{QueenBomberEnabled: true}
		dg := domain.NewDaifugo(tc, players, config)
		// Human plays a 12; has spare card → QueenBomber pending set
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false)) // spare
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		_ = dg.PlayerPlay([]int{0})
		result := tdp.Output(dg, nil)
		assert.Contains(t, result, "【12ボンバー】")
	})

	t.Run("success Output does not show 反則上がり for non-penalized player", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDaifugoPlayersForPresenter()
		dg := domain.NewDaifugo(tc, players, domain.DefaultDaifugoConfig())
		players[0].SetIsFinished(true)
		players[0].SetRank(1)
		players[1].SetIsFinished(true)
		players[1].SetRank(2)
		players[2].SetIsFinished(true)
		players[2].SetRank(3)
		players[3].SetIsFinished(true)
		players[3].SetRank(4)
		result := tdp.Output(dg, nil)
		assert.NotContains(t, result, "[反則上がり]")
	})
}

func TestDaifugoCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.DaifugoCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		mockGame := new(interfaces.MockDaifugoGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "played 3 of spades"},
		}
		mockGame.On("GetGameEndFlag").Return(true)
		mockGame.On("GetActionLog").Return(entries)
		// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
		mockGame.On("GetPlayer", mock.Anything).Return(domain.NewDaifugoPlayer(true)).Maybe()

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "play")
		assert.Contains(t, result, "played 3 of spades")
		mockGame.AssertExpectations(t)
	})

	t.Run("nil_entries", func(t *testing.T) {
		mockGame := new(interfaces.MockDaifugoGame)
		mockGame.On("GetGameEndFlag").Return(true)
		mockGame.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
		// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
		mockGame.On("GetPlayer", mock.Anything).Return(domain.NewDaifugoPlayer(true)).Maybe()

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, "棋譜はありません")
		mockGame.AssertExpectations(t)
	})

	t.Run("game_not_ended", func(t *testing.T) {
		mockGame := new(interfaces.MockDaifugoGame)
		mockGame.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, "棋譜はありません")
		mockGame.AssertExpectations(t)
	})
}

// **どの手札が出せるかを自力で計算させていた (#4733)。**CrazyEights は
// 出せる札に "*" を付けているのに、遥かにルールが複雑な大富豪には無かった。
func TestDaifugoCuiPresenter_MarksPlayableCards(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	tdp := new(presenter.DaifugoCuiPresenter)

	newGame := func(hand ...*domain.Card) (*domain.Daifugo, []*domain.DaifugoPlayer) {
		players := makeDaifugoPlayersForPresenter()
		dg := domain.NewDaifugo(domain.NewTrumpCards(0), players, domain.DefaultDaifugoConfig())
		dg.SetCurrentTurn(0)
		for _, c := range hand {
			players[0].AddCard(c)
		}
		for i := 1; i < len(players); i++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		}
		return dg, players
	}
	card := func(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

	t.Run("stars the card that beats the table and leaves the other bare", func(t *testing.T) {
		dg, _ := newGame(card(domain.CardDesignSpade, 5), card(domain.CardDesignHeart, 12))
		dg.SetTableCards([]*domain.Card{card(domain.CardDesignClover, 9)})

		result := tdp.Output(dg, nil)
		assert.Contains(t, result, "[1]HEART 12*")
		// **付かないことも確かめる。**全部に星が付く実装でも「付く」だけなら通る。
		assert.NotContains(t, result, "[0]SPADE 5*")
	})

	t.Run("stars every card when the table is empty", func(t *testing.T) {
		dg, _ := newGame(card(domain.CardDesignSpade, 5), card(domain.CardDesignHeart, 12))

		result := tdp.Output(dg, nil)
		assert.Contains(t, result, "[0]SPADE 5*")
		assert.Contains(t, result, "[1]HEART 12*")
	})

	t.Run("stars nothing when no card can be played", func(t *testing.T) {
		dg, _ := newGame(card(domain.CardDesignSpade, 4), card(domain.CardDesignHeart, 5))
		dg.SetTableCards([]*domain.Card{card(domain.CardDesignClover, 13)})

		assert.NotContains(t, tdp.Output(dg, nil), "*")
	})

	// **CPU の手番では人間の手札に印を付けない。**手番プレイヤーの手札で
	// 計算したインデックスをそのまま人間の手札に当てると、出せない札に
	// 印が付く。
	t.Run("stars nothing while a CPU is to act", func(t *testing.T) {
		dg, _ := newGame(card(domain.CardDesignSpade, 4), card(domain.CardDesignHeart, 5))
		dg.SetTableCards([]*domain.Card{card(domain.CardDesignClover, 13)})
		dg.SetCurrentTurn(1)

		assert.NotContains(t, tdp.Output(dg, nil), "*")
	})

	// **革命中は印の付く札が入れ替わる。**場の状態を見ていない実装だと
	// ここで通らない。
	t.Run("a revolution moves the star to the weaker card", func(t *testing.T) {
		dg, _ := newGame(card(domain.CardDesignSpade, 5), card(domain.CardDesignHeart, 12))
		dg.SetTableCards([]*domain.Card{card(domain.CardDesignClover, 9)})
		dg.SetRevolutionActive(true)

		result := tdp.Output(dg, nil)
		assert.Contains(t, result, "[0]SPADE 5*")
		assert.NotContains(t, result, "[1]HEART 12*")
	})
}
