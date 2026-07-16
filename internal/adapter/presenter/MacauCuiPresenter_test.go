package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupMacauCuiMock() *interfaces.MockMacauGame {
	m := new(interfaces.MockMacauGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(30)
	m.On("GetDirection").Return(1)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetChosenSuit").Return(-1)
	m.On("GetPenaltyDrawCount").Return(0)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.MacauPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func makeMacauPlayers() []*domain.MacauPlayer {
	return []*domain.MacauPlayer{
		domain.NewMacauPlayer(true),
		domain.NewMacauPlayer(false),
		domain.NewMacauPlayer(false),
		domain.NewMacauPlayer(false),
	}
}

func setupMacauCuiMockWithPlayers() (*interfaces.MockMacauGame, []*domain.MacauPlayer) {
	m := setupMacauCuiMock()
	players := makeMacauPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestMacauCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.MacauCuiPresenter)

	t.Run("initial state header and players", func(t *testing.T) {
		m, players := setupMacauCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 3, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "Macau")
		assert.Contains(t, result, "ラウンド: 1")
		assert.Contains(t, result, "山札: 30枚")
		assert.Contains(t, result, "あなた: 累積0点 ラウンド0点 1枚")
		assert.Contains(t, result, "[0]SPADE 1")
		assert.Contains(t, result, "手番: あなた")
	})

	t.Run("discard top shown", func(t *testing.T) {
		m, _ := setupMacauCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 7, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "捨て札: HEART 7")
	})

	t.Run("chosen suit shown", func(t *testing.T) {
		m, _ := setupMacauCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetChosenSuit")
		m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignSpade, 8, false))
		m.On("GetChosenSuit").Return(domain.CardDesignHeart)
		result := p.Output(m, nil)
		assert.Contains(t, result, "(指定スート: ♥)")
	})

	t.Run("penalty stack shown", func(t *testing.T) {
		m, _ := setupMacauCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPenaltyDrawCount")
		m.On("GetPenaltyDrawCount").Return(4)
		result := p.Output(m, nil)
		assert.Contains(t, result, "4")
	})

	t.Run("reverse direction shown", func(t *testing.T) {
		m, _ := setupMacauCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDirection")
		m.On("GetDirection").Return(-1)
		result := p.Output(m, nil)
		assert.Contains(t, result, "←")
	})

	t.Run("error message shown", func(t *testing.T) {
		m, _ := setupMacauCuiMockWithPlayers()
		result := p.Output(m, errors.New("invalid card index"))
		assert.Contains(t, result, "invalid card index")
	})

	t.Run("game ended shows winner", func(t *testing.T) {
		m, _ := setupMacauCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！")
		assert.Contains(t, result, "あなたの勝利です！")
	})

	t.Run("choose suit phase shows commands", func(t *testing.T) {
		m, _ := setupMacauCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MacauPhaseChooseSuit)
		result := p.Output(m, nil)
		assert.Contains(t, result, "スート選択フェーズ")
	})

	t.Run("must declare phase shows commands", func(t *testing.T) {
		m, _ := setupMacauCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MacauPhaseMustDeclare)
		result := p.Output(m, nil)
		assert.Contains(t, result, "宣言フェーズ")
		assert.Contains(t, result, "declare")
	})

	t.Run("round end phase shows next command", func(t *testing.T) {
		m, _ := setupMacauCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MacauPhaseRoundEnd)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ラウンド終了")
		assert.Contains(t, result, "nr / nextround")
	})
}

func TestMacauCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.MacauCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockMacauGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays SPADE 5"},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)
		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "You plays SPADE 5")
	})

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockMacauGame)
		m.On("GetGameEndFlag").Return(false)
		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜はありません")
	})
}

func TestMacauCuiPresenter_HintOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.MacauCuiPresenter)

	t.Run("lists only playable card indices", func(t *testing.T) {
		m, players := setupMacauCuiMockWithPlayers()
		m.On("IsHumanTurn").Return(true)
		playable := domain.NewCard(domain.CardDesignHeart, 5, false)
		unplayable := domain.NewCard(domain.CardDesignSpade, 9, false)
		players[0].AddCard(playable)
		players[0].AddCard(unplayable)
		m.On("IsValidPlay", playable).Return(true)
		m.On("IsValidPlay", unplayable).Return(false)

		out := p.HintOutput(m)
		assert.Contains(t, out, i18n.Tf("macau.hintPlayable", "cards", "[0]HEART 5"))
		assert.NotContains(t, out, "[1]")
	})

	t.Run("advises drawing when nothing is playable", func(t *testing.T) {
		m, players := setupMacauCuiMockWithPlayers()
		m.On("IsHumanTurn").Return(true)
		c := domain.NewCard(domain.CardDesignSpade, 9, false)
		players[0].AddCard(c)
		m.On("IsValidPlay", c).Return(false)

		out := p.HintOutput(m)
		assert.Contains(t, out, i18n.T("macau.hintDraw"))
	})

	t.Run("advises accepting the penalty mid-chain", func(t *testing.T) {
		m, players := setupMacauCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPenaltyDrawCount")
		m.On("GetPenaltyDrawCount").Return(2)
		m.On("IsHumanTurn").Return(true)
		c := domain.NewCard(domain.CardDesignSpade, 9, false)
		players[0].AddCard(c)
		m.On("IsValidPlay", c).Return(false)

		out := p.HintOutput(m)
		assert.Contains(t, out, i18n.T("macau.hintReceivePenalty"))
	})

	t.Run("no hint outside the human's play turn", func(t *testing.T) {
		m, _ := setupMacauCuiMockWithPlayers()
		m.On("IsHumanTurn").Return(false)
		assert.Contains(t, p.HintOutput(m), i18n.T("macau.hintNone"))
	})
}
