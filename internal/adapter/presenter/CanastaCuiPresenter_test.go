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

func setupCanastaCuiMock() *interfaces.MockCanastaGame {
	m := new(interfaces.MockCanastaGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(54)
	m.On("GetDiscardPileCount").Return(0)
	m.On("GetIsFrozen").Return(false)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.CanastaPhaseDraw)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupCanastaCuiMockWithPlayers() (*interfaces.MockCanastaGame, []*domain.CanastaPlayer) {
	m := setupCanastaCuiMock()
	players := makeCanastaPlayers()
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	return m, players
}

func TestCanastaCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.CanastaCuiPresenter)

	t.Run("initial state with header and player info", func(t *testing.T) {
		m, players := setupCanastaCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "Canasta (カナスタ)")
		assert.Contains(t, result, "ラウンド: 1")
		assert.Contains(t, result, "山札: 54枚")
		assert.Contains(t, result, "あなた: 累積0点 ラウンド0点 1枚")
		assert.Contains(t, result, "[0]SPADE 5")
		assert.Contains(t, result, "CPU 1: 累積0点 ラウンド0点 1枚")
		assert.Contains(t, result, "手番: あなた")
		assert.Contains(t, result, "ds")
		assert.Contains(t, result, "dd")
	})

	t.Run("frozen pile shown", func(t *testing.T) {
		m, _ := setupCanastaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetIsFrozen")
		m.On("GetIsFrozen").Return(true)

		result := p.Output(m, nil)
		assert.Contains(t, result, "[フリーズ]")
		// Draw phase with a frozen pile also shows the pickup-restriction note.
		assert.Contains(t, result, "捨て札は凍結中")
	})

	t.Run("meld phase shows the score-tiered initial-meld minimum", func(t *testing.T) {
		m, players := setupCanastaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CanastaPhaseMeld)
		// Human (player 0), no initial meld, cumulative 0 -> 50-point band.
		players[0].SetCumulativeScore(0)
		players[0].SetHasInitMeld(false)

		result := p.Output(m, nil)
		assert.Contains(t, result, "初回メルド最低点: 50点")
	})

	t.Run("meld phase minimum rises with the score band", func(t *testing.T) {
		m, players := setupCanastaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CanastaPhaseMeld)
		players[0].SetCumulativeScore(1500) // 1500-2999 band -> 90
		players[0].SetHasInitMeld(false)

		result := p.Output(m, nil)
		assert.Contains(t, result, "初回メルド最低点: 90点")
	})

	t.Run("meld phase hides the minimum once the initial meld is done", func(t *testing.T) {
		m, players := setupCanastaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CanastaPhaseMeld)
		players[0].SetHasInitMeld(true)

		result := p.Output(m, nil)
		assert.NotContains(t, result, "初回メルド最低点")
	})

	t.Run("discard top shown", func(t *testing.T) {
		m, _ := setupCanastaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		top := domain.NewCard(domain.CardDesignHeart, 7, false)
		m.On("GetDiscardTop").Return(top)

		result := p.Output(m, nil)
		assert.Contains(t, result, "捨て札: HEART 7")
	})

	t.Run("discard top nil hides section", func(t *testing.T) {
		m, _ := setupCanastaCuiMockWithPlayers()

		result := p.Output(m, nil)
		assert.NotContains(t, result, "捨て札:")
	})

	t.Run("player with scores", func(t *testing.T) {
		m, players := setupCanastaCuiMockWithPlayers()
		players[1].SetCumulativeScore(300)
		players[1].SetRoundScore(100)

		result := p.Output(m, nil)
		assert.Contains(t, result, "CPU 1: 累積300点 ラウンド100点 0枚")
	})

	t.Run("player with meld shown", func(t *testing.T) {
		m, players := setupCanastaCuiMockWithPlayers()
		meld := &domain.CanastaMeld{
			Cards: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 7, false),
				domain.NewCard(domain.CardDesignHeart, 7, false),
				domain.NewCard(domain.CardDesignClover, 7, false),
			},
			IsNatural: true,
		}
		players[0].SetMelds([]*domain.CanastaMeld{meld})

		result := p.Output(m, nil)
		assert.Contains(t, result, "ナチュラル")
		assert.Contains(t, result, "SPADE 7")
	})

	t.Run("error message shown", func(t *testing.T) {
		m, _ := setupCanastaCuiMockWithPlayers()
		testErr := errors.New("invalid card index")

		result := p.Output(m, testErr)
		assert.Contains(t, result, "invalid card index")
	})

	t.Run("game ended shows winner human", func(t *testing.T) {
		m, _ := setupCanastaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！")
		assert.Contains(t, result, "あなたの勝利です！")
	})

	t.Run("game ended shows winner CPU", func(t *testing.T) {
		m, _ := setupCanastaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(1)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！")
		assert.Contains(t, result, "CPU 1の勝利です！")
	})

	t.Run("draw phase shows current player CPU", func(t *testing.T) {
		m, _ := setupCanastaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentPlayerIdx")
		m.On("GetCurrentPlayerIdx").Return(1)

		result := p.Output(m, nil)
		assert.Contains(t, result, "手番: CPU 1")
	})

	t.Run("meld phase shows commands", func(t *testing.T) {
		m, _ := setupCanastaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CanastaPhaseMeld)

		result := p.Output(m, nil)
		assert.Contains(t, result, "メルドフェーズ")
		assert.Contains(t, result, "m ")
		assert.Contains(t, result, "sm")
	})

	t.Run("discard phase shows commands", func(t *testing.T) {
		m, _ := setupCanastaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CanastaPhaseDiscard)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ディスカードフェーズ")
		assert.Contains(t, result, "d <idx>")
		assert.Contains(t, result, "go")
	})

	t.Run("round end phase shows next command", func(t *testing.T) {
		m, _ := setupCanastaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CanastaPhaseRoundEnd)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ラウンド終了")
		assert.Contains(t, result, "nr / nextround")
	})

	t.Run("red 3 tag shown", func(t *testing.T) {
		m, players := setupCanastaCuiMockWithPlayers()
		players[0].AddRed3(domain.NewCard(domain.CardDesignHeart, 3, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "赤3: 1枚")
	})

	t.Run("canasta star tag shown when player holds a canasta meld", func(t *testing.T) {
		m, players := setupCanastaCuiMockWithPlayers()
		// A 7-card natural meld of 4s satisfies HasCanasta() (>=7 cards).
		// IsNatural=true also exercises the "ナチュラル" meld-type label;
		// the m.IsCanasta() branch attaches "カナスタ" to that label.
		cards := make([]*domain.Card, 7)
		for i := range cards {
			cards[i] = domain.NewCard(domain.CardDesignSpade, 4, false)
		}
		players[0].AddMeld(&domain.CanastaMeld{Cards: cards, IsNatural: true})

		result := p.Output(m, nil)
		assert.Contains(t, result, "★カナスタ")
		assert.Contains(t, result, "ナチュラルカナスタ")
	})
}

func TestCanastaCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.CanastaCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockCanastaGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "draw_stock", Detail: "drew from stock"},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "draw_stock")
		m.AssertExpectations(t)
	})

	t.Run("game not ended returns empty", func(t *testing.T) {
		m := new(interfaces.MockCanastaGame)
		m.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(m)
		assert.NotEmpty(t, result)
		m.AssertExpectations(t)
	})
}
