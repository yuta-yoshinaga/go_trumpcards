package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupCatchTenWebMock() *interfaces.MockCatchTenGame {
	m := new(interfaces.MockCatchTenGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.CatchTenPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTeamScore", 0).Return(0)
	m.On("GetTeamScore", 1).Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultCatchTenConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func makeCatchTenPlayers() []*domain.CatchTenPlayer {
	return []*domain.CatchTenPlayer{
		domain.NewCatchTenPlayer(true, 0),
		domain.NewCatchTenPlayer(false, 1),
		domain.NewCatchTenPlayer(false, 0),
		domain.NewCatchTenPlayer(false, 1),
	}
}

func setupCatchTenWebMockWithPlayers() (*interfaces.MockCatchTenGame, []*domain.CatchTenPlayer) {
	m := setupCatchTenWebMock()
	players := makeCatchTenPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestCatchTenWebPresenter_Output(t *testing.T) {
	p := new(presenter.CatchTenWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupCatchTenWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := p.Output(m, nil)
		var resObj controller.CatchTenWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, 4, len(resObj.Players))
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, 0, resObj.Phase)
		assert.Equal(t, domain.CardDesignSpade, resObj.TrumpSuit)
		assert.Equal(t, -1, resObj.WinnerTeam)
		assert.Equal(t, "catchten.playPhase.lead", resObj.MessageCode)
	})

	t.Run("human cards shown, CPU hidden", func(t *testing.T) {
		m, players := setupCatchTenWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := p.Output(m, nil)
		var resObj controller.CatchTenWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
		assert.Equal(t, 0, resObj.Players[0].Team)
		assert.Equal(t, 1, resObj.Players[1].Team)
	})

	t.Run("error message", func(t *testing.T) {
		m, _ := setupCatchTenWebMockWithPlayers()
		result := p.Output(m, errors.New("test error"))
		var resObj controller.CatchTenWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "test error", resObj.Message)
	})

	t.Run("game end humanWin (team 0)", func(t *testing.T) {
		m, _ := setupCatchTenWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetWinnerTeam").Return(0)

		result := p.Output(m, nil)
		var resObj controller.CatchTenWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "catchten.result.humanWin", resObj.MessageCode)
	})

	t.Run("game end cpuWin (team 1)", func(t *testing.T) {
		m, _ := setupCatchTenWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetWinnerTeam").Return(1)

		result := p.Output(m, nil)
		var resObj controller.CatchTenWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "catchten.result.cpuWin", resObj.MessageCode)
	})

	t.Run("game end draw", func(t *testing.T) {
		m, _ := setupCatchTenWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetWinnerTeam").Return(domain.CatchTenDrawTeam)

		result := p.Output(m, nil)
		var resObj controller.CatchTenWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "catchten.result.draw", resObj.MessageCode)
	})

	t.Run("play phase follow message", func(t *testing.T) {
		m, _ := setupCatchTenWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 6, false)},
		})
		result := p.Output(m, nil)
		var resObj controller.CatchTenWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "catchten.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end message", func(t *testing.T) {
		m, _ := setupCatchTenWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CatchTenPhaseTrickEnd)
		result := p.Output(m, nil)
		var resObj controller.CatchTenWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "catchten.trickEnd", resObj.MessageCode)
	})

	t.Run("round end message", func(t *testing.T) {
		m, _ := setupCatchTenWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CatchTenPhaseRoundEnd)
		result := p.Output(m, nil)
		var resObj controller.CatchTenWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "catchten.roundEnd", resObj.MessageCode)
	})
}

func TestCatchTenWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.CatchTenWebPresenter)

	t.Run("with hint", func(t *testing.T) {
		m, _ := setupCatchTenWebMockWithPlayers()
		cardIdx := 3
		m.On("GetHint").Return(&domain.CatchTenHint{CardIndex: &cardIdx, Reason: "follow_suit"})
		result := p.HintOutput(m)
		var resObj controller.CatchTenWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, 3, *resObj.Hint.CardIndex)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupCatchTenWebMockWithPlayers()
		m.On("GetHint").Return((*domain.CatchTenHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.CatchTenWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Nil(t, resObj.Hint)
	})
}

func TestCatchTenWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.CatchTenWebPresenter)
	m := setupCatchTenWebMock()
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "test"},
	})
	assert.NotEmpty(t, p.ActionLogOutput(m))
}
