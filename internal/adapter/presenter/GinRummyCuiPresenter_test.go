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

func setupGinRummyCuiMock() *interfaces.MockGinRummyGame {
	m := new(interfaces.MockGinRummyGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(31)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.GinRummyPhaseDraw)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetKnockerMelds").Return(([][]*domain.Card)(nil)).Maybe()
	return m
}

func makeGinRummyPlayers() []*domain.GinRummyPlayer {
	return []*domain.GinRummyPlayer{
		domain.NewGinRummyPlayer(true),
		domain.NewGinRummyPlayer(false),
	}
}

func setupGinRummyCuiMockWithPlayers() (*interfaces.MockGinRummyGame, []*domain.GinRummyPlayer) {
	m := setupGinRummyCuiMock()
	players := makeGinRummyPlayers()
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	return m, players
}

func TestGinRummyCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.GinRummyCuiPresenter)

	t.Run("initial state with header and player info", func(t *testing.T) {
		m, players := setupGinRummyCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 3, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "Gin Rummy (ジンラミー)")
		assert.Contains(t, result, "ラウンド: 1")
		assert.Contains(t, result, "山札: 31枚")
		assert.Contains(t, result, "あなた: 累積0点 ラウンド0点 2枚")
		assert.Contains(t, result, "[0]SPADE 1")
		assert.Contains(t, result, "[1]HEART 5")
		assert.Contains(t, result, "CPU 1: 累積0点 ラウンド0点 1枚")
		assert.Contains(t, result, "手番: あなた")
		assert.Contains(t, result, "ds")
		assert.Contains(t, result, "dd")
	})

	t.Run("discard top shown", func(t *testing.T) {
		m, _ := setupGinRummyCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		top := domain.NewCard(domain.CardDesignHeart, 7, false)
		m.On("GetDiscardTop").Return(top)

		result := p.Output(m, nil)
		assert.Contains(t, result, "捨て札: HEART 7")
	})

	t.Run("discard top nil hides section", func(t *testing.T) {
		m, _ := setupGinRummyCuiMockWithPlayers()

		result := p.Output(m, nil)
		assert.NotContains(t, result, "捨て札:")
	})

	t.Run("human with no cards does not print cards line", func(t *testing.T) {
		m, _ := setupGinRummyCuiMockWithPlayers()

		result := p.Output(m, nil)
		assert.Contains(t, result, "あなた: 累積0点 ラウンド0点 0枚")
		assert.NotContains(t, result, "[0]")
	})

	t.Run("player with scores", func(t *testing.T) {
		m, players := setupGinRummyCuiMockWithPlayers()
		players[1].SetCumulativeScore(150)
		players[1].SetRoundScore(30)

		result := p.Output(m, nil)
		assert.Contains(t, result, "CPU 1: 累積150点 ラウンド30点 0枚")
	})

	t.Run("error message shown", func(t *testing.T) {
		m, _ := setupGinRummyCuiMockWithPlayers()
		testErr := errors.New("invalid card index")

		result := p.Output(m, testErr)
		assert.Contains(t, result, "invalid card index")
	})

	t.Run("game ended shows winner human", func(t *testing.T) {
		m, _ := setupGinRummyCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！")
		assert.Contains(t, result, "あなたの勝利です！")
	})

	t.Run("game ended shows winner CPU", func(t *testing.T) {
		m, _ := setupGinRummyCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(1)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！")
		assert.Contains(t, result, "CPU 1の勝利です！")
	})

	t.Run("draw phase shows current player CPU", func(t *testing.T) {
		m, _ := setupGinRummyCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentPlayerIdx")
		m.On("GetCurrentPlayerIdx").Return(1)

		result := p.Output(m, nil)
		assert.Contains(t, result, "手番: CPU 1")
	})

	t.Run("discard phase shows commands", func(t *testing.T) {
		m, _ := setupGinRummyCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.GinRummyPhaseDiscard)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ディスカードフェーズ")
		assert.Contains(t, result, "d <idx>")
		assert.Contains(t, result, "k <idx>")
	})

	t.Run("discard phase shows deadwood and knock-available for a low hand", func(t *testing.T) {
		m, players := setupGinRummyCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.GinRummyPhaseDiscard)
		// Two runs + two low strays: discarding one stray leaves deadwood <= 10.
		for _, v := range []int{1, 2, 3, 4, 5, 6} {
			players[0].AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
		}
		for _, v := range []int{7, 8, 9} {
			players[0].AddCard(domain.NewCard(domain.CardDesignHeart, v, false))
		}
		players[0].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 3, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "最小デッドウッド(1枚捨て後):")
		assert.Contains(t, result, "ノック可能")
	})

	t.Run("discard phase hides knock-available for a high hand", func(t *testing.T) {
		m, players := setupGinRummyCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.GinRummyPhaseDiscard)
		// Scattered high cards with no runs or sets: deadwood stays well above 10.
		highCards := []struct {
			d, v int
		}{
			{domain.CardDesignSpade, 13}, {domain.CardDesignHeart, 12}, {domain.CardDesignClover, 11},
			{domain.CardDesignDiamond, 10}, {domain.CardDesignSpade, 9}, {domain.CardDesignHeart, 8},
			{domain.CardDesignClover, 7}, {domain.CardDesignDiamond, 6}, {domain.CardDesignSpade, 4},
			{domain.CardDesignHeart, 3}, {domain.CardDesignClover, 2},
		}
		for _, c := range highCards {
			players[0].AddCard(domain.NewCard(c.d, c.v, false))
		}

		result := p.Output(m, nil)
		assert.Contains(t, result, "最小デッドウッド(1枚捨て後):")
		assert.NotContains(t, result, "ノック可能")
	})

	t.Run("layoff phase shows commands", func(t *testing.T) {
		m, _ := setupGinRummyCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.GinRummyPhaseLayoff)

		result := p.Output(m, nil)
		assert.Contains(t, result, "レイオフフェーズ")
		assert.Contains(t, result, "lo")
	})

	t.Run("layoff phase lists knocker melds with set/run labels", func(t *testing.T) {
		m, _ := setupGinRummyCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetKnockerMelds")
		m.On("GetPhase").Return(domain.GinRummyPhaseLayoff)
		m.On("GetKnockerMelds").Return([][]*domain.Card{
			{ // set of 7s
				domain.NewCard(domain.CardDesignSpade, 7, false),
				domain.NewCard(domain.CardDesignHeart, 7, false),
				domain.NewCard(domain.CardDesignDiamond, 7, false),
			},
			{ // run of clubs 4-5-6
				domain.NewCard(domain.CardDesignClover, 4, false),
				domain.NewCard(domain.CardDesignClover, 5, false),
				domain.NewCard(domain.CardDesignClover, 6, false),
			},
		})

		result := p.Output(m, nil)
		assert.Contains(t, result, "[ノッカーのメルド]")
		assert.Contains(t, result, "メルド1(セット):")
		assert.Contains(t, result, "メルド2(ラン):")
	})

	t.Run("round end phase shows next command", func(t *testing.T) {
		m, _ := setupGinRummyCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.GinRummyPhaseRoundEnd)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ラウンド終了")
		assert.Contains(t, result, "nr / nextround・・・次のラウンドへ")
	})
}

func TestGinRummyCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.GinRummyCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockGinRummyGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "draw_stock", Detail: "You draws from stock"},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "draw_stock")
		assert.Contains(t, result, "You draws from stock")
		m.AssertExpectations(t)
	})

	t.Run("nil entries", func(t *testing.T) {
		m := new(interfaces.MockGinRummyGame)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜はありません")
		m.AssertExpectations(t)
	})

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockGinRummyGame)
		m.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜はありません")
		m.AssertExpectations(t)
	})
}
