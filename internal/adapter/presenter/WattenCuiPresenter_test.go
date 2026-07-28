package presenter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupWattenCuiMock() *interfaces.MockWattenGame {
	m := new(interfaces.MockWattenGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.WattenPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetSchlagRank").Return(10)
	m.On("GetCriticalSuit").Return(1)
	m.On("GetStake").Return(2)
	m.On("GetPendingStake").Return(3)
	m.On("GetRaiserTeam").Return(1)
	m.On("GetResponderIdx").Return(0)
	m.On("CanHumanRaise").Return(false)
	m.On("GetTeamScore", 0).Return(0)
	m.On("GetTeamScore", 1).Return(0)
	m.On("GetDealWinnerTeam").Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func makeWattenPlayers() []*domain.WattenPlayer {
	return []*domain.WattenPlayer{
		domain.NewWattenPlayer(true, 0),
		domain.NewWattenPlayer(false, 1),
		domain.NewWattenPlayer(false, 0),
		domain.NewWattenPlayer(false, 1),
	}
}

func setupWattenCuiMockWithPlayers() (*interfaces.MockWattenGame, []*domain.WattenPlayer) {
	m := setupWattenCuiMock()
	players := makeWattenPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestWattenCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.WattenCuiPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupWattenCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 13, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "Watten")
		assert.Contains(t, result, "ディール: 1")
		assert.Contains(t, result, "Schlag=10")
		assert.Contains(t, result, "SPADE")
		assert.Contains(t, result, "ステーク: 2")
	})

	t.Run("declared schlag rank shows a card-face label", func(t *testing.T) {
		m, _ := setupWattenCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetSchlagRank")
		m.On("GetSchlagRank").Return(11)

		result := p.Output(m, nil)
		// Rank 11 is shown as "J", not the raw number.
		assert.Contains(t, result, "Schlag=J")
		assert.NotContains(t, result, "Schlag=11")
	})

	t.Run("undeclared", func(t *testing.T) {
		m, _ := setupWattenCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCriticalSuit")
		m.On("GetCriticalSuit").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "未決定")
	})

	t.Run("phase: declare", func(t *testing.T) {
		m, _ := setupWattenCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.WattenPhaseDeclare)

		result := p.Output(m, nil)
		assert.Contains(t, result, "宣言")
	})

	t.Run("phase: respond", func(t *testing.T) {
		m, _ := setupWattenCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.WattenPhaseRespond)

		result := p.Output(m, nil)
		assert.Contains(t, result, "応答")
	})

	t.Run("phase: round end", func(t *testing.T) {
		m, _ := setupWattenCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.WattenPhaseRoundEnd)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ディール終了")
	})

	t.Run("play phase with raise available", func(t *testing.T) {
		m, _ := setupWattenCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "CanHumanRaise")
		m.On("CanHumanRaise").Return(true)

		result := p.Output(m, nil)
		assert.Contains(t, result, "手番")
	})

	t.Run("game end", func(t *testing.T) {
		m, _ := setupWattenCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了")
	})
}

func TestWattenCuiPresenter_HintOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.WattenCuiPresenter)

	t.Run("nil hint", func(t *testing.T) {
		m := setupWattenCuiMock()
		m.On("GetHint").Return((*domain.WattenHint)(nil))
		assert.Contains(t, p.HintOutput(m), "ヒントはありません")
	})

	t.Run("declare hint", func(t *testing.T) {
		m := setupWattenCuiMock()
		rank, suit := 10, 2
		m.On("GetHint").Return(&domain.WattenHint{Action: "declare", Rank: &rank, Suit: &suit, Reason: "declare_strong"})
		assert.Contains(t, p.HintOutput(m), "Schlag")
	})

	t.Run("raise hint", func(t *testing.T) {
		m := setupWattenCuiMock()
		m.On("GetHint").Return(&domain.WattenHint{Action: "raise", Reason: "raise_strong"})
		assert.Contains(t, p.HintOutput(m), "HINT")
	})

	t.Run("hold hint", func(t *testing.T) {
		m := setupWattenCuiMock()
		m.On("GetHint").Return(&domain.WattenHint{Action: "hold", Reason: "hold_ok"})
		assert.Contains(t, p.HintOutput(m), "hold")
	})

	t.Run("fold hint", func(t *testing.T) {
		m := setupWattenCuiMock()
		m.On("GetHint").Return(&domain.WattenHint{Action: "fold", Reason: "fold_weak"})
		assert.Contains(t, p.HintOutput(m), "fold")
	})

	t.Run("card hint", func(t *testing.T) {
		m, players := setupWattenCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		idx := 0
		m.On("GetHint").Return(&domain.WattenHint{Action: "play", CardIndex: &idx, Reason: "follow_win"})
		assert.Contains(t, p.HintOutput(m), "HINT")
	})
}

func TestWattenCuiPresenter_ActionLogOutput(t *testing.T) {
	m := setupWattenCuiMock()
	p := new(presenter.WattenCuiPresenter)
	assert.NotNil(t, p.ActionLogOutput(m))
}
