//go:build test

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

func setupSevenBridgeCuiMock() *interfaces.MockSevenBridgeGame {
	m := new(interfaces.MockSevenBridgeGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(37)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.SevenBridgePhaseDraw)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func makeSevenBridgePlayers() []*domain.SevenBridgePlayer {
	return []*domain.SevenBridgePlayer{
		domain.NewSevenBridgePlayer(true),
		domain.NewSevenBridgePlayer(false),
	}
}

func setupSevenBridgeCuiMockWithPlayers() (*interfaces.MockSevenBridgeGame, []*domain.SevenBridgePlayer) {
	m := setupSevenBridgeCuiMock()
	players := makeSevenBridgePlayers()
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	return m, players
}

func TestSevenBridgeCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.SevenBridgeCuiPresenter)

	t.Run("initial header/state", func(t *testing.T) {
		m, players := setupSevenBridgeCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 3, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "Seven Bridge (セブンブリッジ)")
		assert.Contains(t, result, "ラウンド: 1")
		assert.Contains(t, result, "山札: 37枚")
		assert.Contains(t, result, "あなた: 累積0点 ラウンド0点 2枚")
		assert.Contains(t, result, "CPU 1: 累積0点 ラウンド0点 1枚")
		assert.Contains(t, result, "手番: あなた")
		assert.Contains(t, result, "ds")
		assert.Contains(t, result, "pon")
		assert.Contains(t, result, "chi")
	})

	t.Run("discard top shown", func(t *testing.T) {
		m, _ := setupSevenBridgeCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 7, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "捨て札: HEART 7")
	})

	t.Run("meld rendered for player", func(t *testing.T) {
		m, players := setupSevenBridgeCuiMockWithPlayers()
		players[0].AppendMeld([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 3, false),
			domain.NewCard(domain.CardDesignClover, 3, false),
			domain.NewCard(domain.CardDesignHeart, 3, false),
		})

		result := p.Output(m, nil)
		assert.Contains(t, result, "場: ")
		assert.Contains(t, result, "SPADE 3")
	})

	t.Run("error shown", func(t *testing.T) {
		m, _ := setupSevenBridgeCuiMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})

	t.Run("game ended → winner human", func(t *testing.T) {
		m, _ := setupSevenBridgeCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！")
		assert.Contains(t, result, "あなたの勝利です！")
	})

	t.Run("game ended → winner CPU", func(t *testing.T) {
		m, _ := setupSevenBridgeCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(1)

		result := p.Output(m, nil)
		assert.Contains(t, result, "CPU 1の勝利です！")
	})

	t.Run("play phase commands", func(t *testing.T) {
		m, _ := setupSevenBridgeCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SevenBridgePhasePlay)

		result := p.Output(m, nil)
		assert.Contains(t, result, "プレイフェーズ")
		assert.Contains(t, result, "m <idx")
		assert.Contains(t, result, "lo <pIdx>")
		assert.Contains(t, result, "d <idx>")
	})

	t.Run("round end phase", func(t *testing.T) {
		m, _ := setupSevenBridgeCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SevenBridgePhaseRoundEnd)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ラウンド終了")
		assert.Contains(t, result, "nr / nextround")
	})

	t.Run("draw phase CPU turn", func(t *testing.T) {
		m, _ := setupSevenBridgeCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentPlayerIdx")
		m.On("GetCurrentPlayerIdx").Return(1)

		result := p.Output(m, nil)
		assert.Contains(t, result, "手番: CPU 1")
	})
}

func TestSevenBridgeCuiPresenter_HintOutput(t *testing.T) {
	p := new(presenter.SevenBridgeCuiPresenter)

	newHintMock := func() *interfaces.MockSevenBridgeGame {
		m := new(interfaces.MockSevenBridgeGame)
		human := domain.NewSevenBridgePlayer(true)
		for _, v := range []int{3, 3, 3, 8} {
			human.AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
		}
		m.On("GetPhase").Return(domain.SevenBridgePhasePlay)
		m.On("IsHumanTurn").Return(true)
		m.On("GetCurrentPlayerIdx").Return(0)
		m.On("GetPlayer", 0).Return(human)
		return m
	}

	t.Run("recommends a meld when one is available", func(t *testing.T) {
		m := newHintMock()
		m.On("SuggestMeld", 0).Return([]int{0, 1, 2})
		assert.Contains(t, p.HintOutput(m), "メルド")
	})

	t.Run("recommends a discard when no meld is available", func(t *testing.T) {
		m := newHintMock()
		m.On("SuggestMeld", 0).Return(([]int)(nil))
		m.On("SuggestDiscard", 0).Return(3)
		assert.Contains(t, p.HintOutput(m), "捨てる")
	})

	t.Run("declines outside the human's play phase", func(t *testing.T) {
		m := new(interfaces.MockSevenBridgeGame)
		m.On("GetPhase").Return(domain.SevenBridgePhaseDraw)
		assert.Contains(t, p.HintOutput(m), "プレイフェーズではありません")
	})
}

func TestSevenBridgeCuiPresenter_ActionLogOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.SevenBridgeCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockSevenBridgeGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "draw_stock", Detail: "You draws from stock"},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "draw_stock")
		m.AssertExpectations(t)
	})

	t.Run("no entries", func(t *testing.T) {
		m := new(interfaces.MockSevenBridgeGame)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜はありません")
		m.AssertExpectations(t)
	})

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockSevenBridgeGame)
		m.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜はありません")
		m.AssertExpectations(t)
	})
}
