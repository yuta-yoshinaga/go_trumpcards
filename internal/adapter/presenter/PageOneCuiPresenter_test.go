package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupPageOneCuiMock() *interfaces.MockPageOneGame {
	m := new(interfaces.MockPageOneGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(30)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.PageOnePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func makePageOnePlayers() []*domain.PageOnePlayer {
	return []*domain.PageOnePlayer{
		domain.NewPageOnePlayer(true),
		domain.NewPageOnePlayer(false),
		domain.NewPageOnePlayer(false),
		domain.NewPageOnePlayer(false),
	}
}

func setupPageOneCuiMockWithPlayers() (*interfaces.MockPageOneGame, []*domain.PageOnePlayer) {
	m := setupPageOneCuiMock()
	players := makePageOnePlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestPageOneCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.PageOneCuiPresenter)

	t.Run("initial play phase", func(t *testing.T) {
		m, players := setupPageOneCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 3, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "Page One (ページワン)")
		assert.Contains(t, result, "ラウンド: 1")
		assert.Contains(t, result, "山札: 30枚")
		assert.Contains(t, result, "あなた: 累積0点 ラウンド0点 1枚")
		assert.Contains(t, result, "CPU 1: 累積0点 ラウンド0点 1枚")
		assert.Contains(t, result, "手番: あなた")
		assert.Contains(t, result, "play <idx>")
	})

	t.Run("discard top shown", func(t *testing.T) {
		m, _ := setupPageOneCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		top := domain.NewCard(domain.CardDesignHeart, 7, false)
		m.On("GetDiscardTop").Return(top)

		result := p.Output(m, nil)
		assert.Contains(t, result, "捨て札: HEART 7")
		// Play condition references the matchable discard top during the play phase.
		assert.Contains(t, result, "出せる条件: HEART 7")
	})

	t.Run("must declare phase", func(t *testing.T) {
		m, _ := setupPageOneCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.PageOnePhaseMustDeclare)

		result := p.Output(m, nil)
		assert.Contains(t, result, "今すぐ「ページワン」を宣言")
		assert.Contains(t, result, "宣言フェーズ")
		assert.Contains(t, result, "declare")
		assert.Contains(t, result, "skip")
	})

	t.Run("round end phase", func(t *testing.T) {
		m, _ := setupPageOneCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.PageOnePhaseRoundEnd)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ラウンド終了")
	})

	t.Run("game end winner human", func(t *testing.T) {
		m, _ := setupPageOneCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！")
		assert.Contains(t, result, "あなたの勝利です！")
	})

	t.Run("error message shown", func(t *testing.T) {
		m, _ := setupPageOneCuiMockWithPlayers()
		result := p.Output(m, errors.New("invalid"))
		assert.Contains(t, result, "invalid")
	})

	t.Run("declared flag rendered", func(t *testing.T) {
		m, players := setupPageOneCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].SetHasDeclared(true)
		result := p.Output(m, nil)
		assert.Contains(t, result, "[PAGE ONE!]")
	})

	t.Run("last-card warning shown for undeclared single-card player", func(t *testing.T) {
		m, players := setupPageOneCuiMockWithPlayers()
		// CPU 1 sits on its last card without declaring → warning; others have 2 cards.
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 3, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "残り1枚！")
		// Only the single-card, undeclared player is warned (one occurrence).
		assert.Equal(t, 1, strings.Count(result, "残り1枚！"))
	})

	t.Run("no last-card warning once the player has declared", func(t *testing.T) {
		m, players := setupPageOneCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].SetHasDeclared(true)
		result := p.Output(m, nil)
		assert.NotContains(t, result, "残り1枚！")
		assert.Contains(t, result, "[PAGE ONE!]")
	})
}

func TestPageOneCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.PageOneCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockPageOneGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays SPADE 5"},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜")
	})

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockPageOneGame)
		m.On("GetGameEndFlag").Return(false)
		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜はありません")
	})
}

func TestPageOneCuiPresenter_HintOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.PageOneCuiPresenter)

	t.Run("lists only playable card indices", func(t *testing.T) {
		m, players := setupPageOneCuiMockWithPlayers()
		m.On("IsHumanTurn").Return(true)
		playable := domain.NewCard(domain.CardDesignHeart, 5, false)
		unplayable := domain.NewCard(domain.CardDesignSpade, 9, false)
		players[0].AddCard(playable)
		players[0].AddCard(unplayable)
		m.On("IsValidPlay", playable).Return(true)
		m.On("IsValidPlay", unplayable).Return(false)

		out := p.HintOutput(m)
		assert.Contains(t, out, i18n.Tf("pageone.hintPlayable", "cards", "[0]HEART 5"))
		assert.NotContains(t, out, "[1]")
	})

	t.Run("advises drawing when nothing is playable", func(t *testing.T) {
		m, players := setupPageOneCuiMockWithPlayers()
		m.On("IsHumanTurn").Return(true)
		c := domain.NewCard(domain.CardDesignSpade, 9, false)
		players[0].AddCard(c)
		m.On("IsValidPlay", c).Return(false)
		assert.Contains(t, p.HintOutput(m), i18n.T("pageone.hintDraw"))
	})

	t.Run("no hint on a CPU turn", func(t *testing.T) {
		m, _ := setupPageOneCuiMockWithPlayers()
		m.On("IsHumanTurn").Return(false)
		assert.Contains(t, p.HintOutput(m), i18n.T("pageone.hintNone"))
	})

	t.Run("no hint outside the play phase", func(t *testing.T) {
		m, _ := setupPageOneCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.PageOnePhaseRoundEnd)
		m.On("IsHumanTurn").Return(true)
		assert.Contains(t, p.HintOutput(m), i18n.T("pageone.hintNone"))
	})
}
