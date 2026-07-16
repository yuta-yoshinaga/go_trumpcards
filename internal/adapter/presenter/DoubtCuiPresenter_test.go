package presenter_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func makeDoubtGameForPresenter() (*domain.Doubt, []*domain.DoubtPlayer) {
	players := []*domain.DoubtPlayer{
		domain.NewDoubtPlayer(true),
		domain.NewDoubtPlayer(false),
		domain.NewDoubtPlayer(false),
		domain.NewDoubtPlayer(false),
	}
	tc := domain.NewTrumpCards(0)
	game := domain.NewDoubt(tc, players)
	return game, players
}

func TestDoubtCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.DoubtCuiPresenter)

	t.Run("initial state", func(t *testing.T) {
		game, players := makeDoubtGameForPresenter()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))

		result := p.Output(game, nil)
		assert.Contains(t, result, "Doubt (ダウト)")
		assert.Contains(t, result, "あなた: 2枚")
		assert.Contains(t, result, "[0]SPADE 1")
		assert.Contains(t, result, "CPU 1: 1枚")
		assert.Contains(t, result, "テーブル: 0枚")
		assert.Contains(t, result, "手番: あなた")
		assert.Contains(t, result, "p <値> <idx...>")
	})

	t.Run("shows card designs for human hand", func(t *testing.T) {
		game, players := makeDoubtGameForPresenter()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 2, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignDiamond, 4, false))

		result := p.Output(game, nil)
		assert.Contains(t, result, "SPADE 1")
		assert.Contains(t, result, "CLOVER 2")
		assert.Contains(t, result, "HEART 3")
		assert.Contains(t, result, "DIAMOND 4")
	})

	t.Run("nil card shows ??", func(t *testing.T) {
		game, players := makeDoubtGameForPresenter()
		players[0].AddCard(nil)

		result := p.Output(game, nil)
		assert.Contains(t, result, "??")
	})

	t.Run("unknown design shows UNKNOWN", func(t *testing.T) {
		game, players := makeDoubtGameForPresenter()
		players[0].AddCard(domain.NewCard(99, 1, false))

		result := p.Output(game, nil)
		assert.Contains(t, result, "UNKNOWN")
	})

	t.Run("finished player shows 上がり", func(t *testing.T) {
		game, players := makeDoubtGameForPresenter()
		players[1].SetIsFinished(true)

		result := p.Output(game, nil)
		assert.Contains(t, result, "上がり")
	})

	t.Run("last action shown", func(t *testing.T) {
		game, _ := makeDoubtGameForPresenter()
		game.SetLastAction(&domain.DoubtAction{
			PlayerIdx:    0,
			ClaimedValue: 5,
			CardCount:    2,
			PlayedCards:  []*domain.Card{},
		})

		result := p.Output(game, nil)
		assert.Contains(t, result, "最後のプレイ")
		assert.Contains(t, result, "「5」を2枚")
		assert.Contains(t, result, "あなた")
	})

	t.Run("doubt result - was lying", func(t *testing.T) {
		game, _ := makeDoubtGameForPresenter()
		revealedCard := domain.NewCard(domain.CardDesignSpade, 7, false)
		game.SetLastDoubtResult(&domain.DoubtDoubtResult{
			DoubterIdx:    1,
			CardPlayerIdx: 0,
			WasLying:      true,
			LoserIdx:      0,
			CardCount:     3,
			RevealedCards: []*domain.Card{revealedCard},
		})

		result := p.Output(game, nil)
		assert.Contains(t, result, "ダウト")
		assert.Contains(t, result, "嘘つき")
		assert.Contains(t, result, "3枚引き取りました")
		assert.Contains(t, result, "公開カード")
		assert.Contains(t, result, "SPADE 7")
	})

	t.Run("doubt result - was honest", func(t *testing.T) {
		game, _ := makeDoubtGameForPresenter()
		game.SetLastDoubtResult(&domain.DoubtDoubtResult{
			DoubterIdx:    1,
			CardPlayerIdx: 0,
			WasLying:      false,
			LoserIdx:      1,
			CardCount:     2,
			RevealedCards: []*domain.Card{},
		})

		result := p.Output(game, nil)
		assert.Contains(t, result, "正直者")
		assert.Contains(t, result, "2枚引き取りました")
	})

	t.Run("doubt result with nil revealed cards", func(t *testing.T) {
		game, _ := makeDoubtGameForPresenter()
		game.SetLastDoubtResult(&domain.DoubtDoubtResult{
			DoubterIdx:    1,
			CardPlayerIdx: 0,
			WasLying:      true,
			LoserIdx:      0,
			CardCount:     1,
			RevealedCards: nil,
		})

		result := p.Output(game, nil)
		assert.Contains(t, result, "嘘つき")
		assert.NotContains(t, result, "公開カード")
	})

	t.Run("doubt result with multiple revealed cards shows comma separator", func(t *testing.T) {
		game, _ := makeDoubtGameForPresenter()
		game.SetLastDoubtResult(&domain.DoubtDoubtResult{
			DoubterIdx:    1,
			CardPlayerIdx: 0,
			WasLying:      true,
			LoserIdx:      0,
			CardCount:     2,
			RevealedCards: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 3, false),
				domain.NewCard(domain.CardDesignHeart, 7, false),
			},
		})

		result := p.Output(game, nil)
		assert.Contains(t, result, "公開カード")
		assert.Contains(t, result, "SPADE 3")
		assert.Contains(t, result, "HEART 7")
		assert.Contains(t, result, ", ")
	})

	t.Run("human action shown", func(t *testing.T) {
		game, _ := makeDoubtGameForPresenter()
		game.SetHumanAction(&domain.DoubtCpuAction{
			PlayerIdx:    0,
			ClaimedValue: 7,
			CardCount:    3,
		})

		result := p.Output(game, nil)
		assert.Contains(t, result, "あなたの行動")
		assert.Contains(t, result, "「7」を3枚")
	})

	t.Run("CPU actions shown", func(t *testing.T) {
		game, _ := makeDoubtGameForPresenter()
		game.SetCpuActions([]*domain.DoubtCpuAction{
			{PlayerIdx: 1, ClaimedValue: 3, CardCount: 2},
			{PlayerIdx: 2, ClaimedValue: 5, CardCount: 1},
		})

		result := p.Output(game, nil)
		assert.Contains(t, result, "CPUの行動")
		assert.Contains(t, result, "CPU 1")
		assert.Contains(t, result, "「3」を2枚")
		assert.Contains(t, result, "CPU 2")
		assert.Contains(t, result, "「5」を1枚")
	})

	t.Run("CPU tell shown only when HasTell is true", func(t *testing.T) {
		game, _ := makeDoubtGameForPresenter()
		game.SetCpuActions([]*domain.DoubtCpuAction{
			{PlayerIdx: 1, ClaimedValue: 3, CardCount: 2, HasTell: true},
			{PlayerIdx: 2, ClaimedValue: 5, CardCount: 1, HasTell: false},
		})

		result := p.Output(game, nil)
		tell := i18n.T("doubt.tell")
		// The tell appears exactly once — for the HasTell CPU only.
		assert.Equal(t, 1, strings.Count(result, tell))
	})

	t.Run("error message shown", func(t *testing.T) {
		game, players := makeDoubtGameForPresenter()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))

		result := p.Output(game, domain.ErrInvalidPlay)
		assert.Contains(t, result, domain.ErrInvalidPlay.Error())
	})

	t.Run("game ended shows winner", func(t *testing.T) {
		game, players := makeDoubtGameForPresenter()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		_ = game.PlayerPlay([]int{0}, 1, 0) // game ends
		game.SetWinnerIdx(0)

		result := p.Output(game, nil)
		assert.Contains(t, result, "ゲーム終了")
		assert.Contains(t, result, "あなたの勝利")
	})

	t.Run("doubt phase shows doubt prompt", func(t *testing.T) {
		game, _ := makeDoubtGameForPresenter()
		// Set up last action so the player name is shown
		game.SetLastAction(&domain.DoubtAction{
			PlayerIdx:    1,
			ClaimedValue: 3,
			CardCount:    1,
			PlayedCards:  []*domain.Card{},
		})
		game.SetPhase(domain.DoubtPhaseDoubt)

		result := p.Output(game, nil)
		assert.Contains(t, result, "ダウトフェーズ")
		assert.Contains(t, result, "d <idx...>")
		assert.Contains(t, result, "s・・・スキップ")
	})

	t.Run("doubt phase with nil lastAction shows generic message", func(t *testing.T) {
		game, _ := makeDoubtGameForPresenter()
		game.SetPhase(domain.DoubtPhaseDoubt)
		// lastAction is nil

		result := p.Output(game, nil)
		assert.Contains(t, result, "ダウトフェーズ")
		assert.Contains(t, result, "d <idx...>")
		assert.NotContains(t, result, "のプレイにダウト")
	})

	t.Run("doubt result with discardedCount > 0 shows discard message", func(t *testing.T) {
		game, _ := makeDoubtGameForPresenter()
		game.SetLastDoubtResult(&domain.DoubtDoubtResult{
			DoubterIdx:     1,
			CardPlayerIdx:  0,
			WasLying:       true,
			LoserIdx:       0,
			CardCount:      3,
			DiscardedCount: 2,
			RevealedCards:  nil,
		})

		result := p.Output(game, nil)
		assert.Contains(t, result, "2枚がゲームから除外されました")
	})

	t.Run("doubt result with discardedCount 0 does not show discard message", func(t *testing.T) {
		game, _ := makeDoubtGameForPresenter()
		game.SetLastDoubtResult(&domain.DoubtDoubtResult{
			DoubterIdx:     1,
			CardPlayerIdx:  0,
			WasLying:       true,
			LoserIdx:       0,
			CardCount:      5,
			DiscardedCount: 0,
			RevealedCards:  nil,
		})

		result := p.Output(game, nil)
		assert.NotContains(t, result, "ゲームから除外されました")
	})

	t.Run("metaAI status line shown when profile exists", func(t *testing.T) {
		game, players := makeDoubtGameForPresenter()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		game.SetHumanProfile(&domain.DoubtHumanProfile{
			GamesPlayed:     2,
			BluffsByBracket: [3]struct{ Bluffs, Total int }{{1, 2}, {3, 6}, {0, 1}},
			DoubtCorrect:    1,
			DoubtTotal:      2,
		})
		result := p.Output(game, nil)
		assert.Contains(t, result, "[メタAI]")
		assert.Contains(t, result, "適応中")
		assert.Contains(t, result, "ゲーム数: 2")
		assert.Contains(t, result, "ブラフ率: 50%")   // 3/6 = 50%
		assert.Contains(t, result, "ダウト正解率: 50%") // 1/2 = 50%
	})

	t.Run("no metaAI line when profile is nil", func(t *testing.T) {
		game, players := makeDoubtGameForPresenter()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		// No SetHumanProfile → profile is nil
		result := p.Output(game, nil)
		assert.NotContains(t, result, "[メタAI]")
	})

	t.Run("getPlayerName for out-of-range index returns 不明", func(t *testing.T) {
		game, _ := makeDoubtGameForPresenter()
		game.SetLastDoubtResult(&domain.DoubtDoubtResult{
			DoubterIdx:    99,
			CardPlayerIdx: 0,
			WasLying:      true,
			LoserIdx:      99,
			CardCount:     1,
			RevealedCards: nil,
		})

		result := p.Output(game, nil)
		assert.Contains(t, result, "UNKNOWN")
	})
}

func TestDoubtCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.DoubtCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		mockGame := new(interfaces.MockDoubtGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "declared 5, played 1 card(s)"},
		}
		mockGame.On("GetGameEndFlag").Return(true)
		mockGame.On("GetActionLog").Return(entries)

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "play")
		assert.Contains(t, result, "declared 5, played 1 card(s)")
		mockGame.AssertExpectations(t)
	})

	t.Run("nil_entries", func(t *testing.T) {
		mockGame := new(interfaces.MockDoubtGame)
		mockGame.On("GetGameEndFlag").Return(true)
		mockGame.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, "棋譜はありません")
		mockGame.AssertExpectations(t)
	})

	t.Run("game_not_ended", func(t *testing.T) {
		mockGame := new(interfaces.MockDoubtGame)
		mockGame.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, "棋譜はありません")
		mockGame.AssertExpectations(t)
	})
}
