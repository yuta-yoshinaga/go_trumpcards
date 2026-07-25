//go:build test

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

func makeNinetyNinePlayers() []*domain.NinetyNinePlayer {
	return []*domain.NinetyNinePlayer{
		domain.NewNinetyNinePlayer(true),
		domain.NewNinetyNinePlayer(false),
		domain.NewNinetyNinePlayer(false),
	}
}

func setupNinetyNineWebMock() *interfaces.MockNinetyNineGame {
	m := new(interfaces.MockNinetyNineGame)
	m.On("GetDealNumber").Return(1)
	m.On("GetTargetScore").Return(100)
	m.On("GetHandSize").Return(9)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.NinetyNinePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTrumpSuit").Return(domain.CardDesignHeart)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultNinetyNineConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupNinetyNineWebMockWithPlayers() (*interfaces.MockNinetyNineGame, []*domain.NinetyNinePlayer) {
	m := setupNinetyNineWebMock()
	players := makeNinetyNinePlayers()
	m.On("GetPlayerCnt").Return(3)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	return m, players
}

func TestNinetyNineWebPresenter_Output(t *testing.T) {
	p := new(presenter.NinetyNineWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupNinetyNineWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))

		result := p.Output(m, nil)
		var resObj controller.NinetyNineWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, 3, len(resObj.Players))
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, 1, resObj.Phase)
		assert.Equal(t, 1, resObj.DealNumber)
		assert.Equal(t, 100, resObj.TargetScore)
		assert.Equal(t, 9, resObj.HandSize)
		assert.Equal(t, domain.CardDesignHeart, resObj.TrumpSuit)
		assert.Equal(t, -1, resObj.WinnerIdx)
	})

	t.Run("error message", func(t *testing.T) {
		m, _ := setupNinetyNineWebMockWithPlayers()
		result := p.Output(m, errors.New("test error"))
		var resObj controller.NinetyNineWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "test error", resObj.Message)
	})

	t.Run("bid phase message", func(t *testing.T) {
		m, _ := setupNinetyNineWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.NinetyNinePhaseBid)
		result := p.Output(m, nil)
		var resObj controller.NinetyNineWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "ninetynine.bidPhase", resObj.MessageCode)
	})

	t.Run("play phase lead message", func(t *testing.T) {
		m, _ := setupNinetyNineWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.NinetyNineWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "ninetynine.playPhase.lead", resObj.MessageCode)
	})

	t.Run("play phase follow message", func(t *testing.T) {
		m, _ := setupNinetyNineWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 7, false)},
		})
		result := p.Output(m, nil)
		var resObj controller.NinetyNineWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "ninetynine.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end message", func(t *testing.T) {
		m, _ := setupNinetyNineWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.NinetyNinePhaseTrickEnd)
		result := p.Output(m, nil)
		var resObj controller.NinetyNineWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "ninetynine.trickEnd", resObj.MessageCode)
	})

	t.Run("round end message", func(t *testing.T) {
		m, _ := setupNinetyNineWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.NinetyNinePhaseRoundEnd)
		result := p.Output(m, nil)
		var resObj controller.NinetyNineWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "ninetynine.roundEnd", resObj.MessageCode)
	})

	t.Run("game end message", func(t *testing.T) {
		m, _ := setupNinetyNineWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		result := p.Output(m, nil)
		var resObj controller.NinetyNineWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.True(t, resObj.GameEndFlag)
		assert.NotEmpty(t, resObj.MessageCode)
	})
}

func TestNinetyNineWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.NinetyNineWebPresenter)

	t.Run("with bury hint", func(t *testing.T) {
		m, _ := setupNinetyNineWebMockWithPlayers()
		m.On("GetHint").Return(&domain.NinetyNineHint{BuryIndices: []int{0, 1, 2}, Reason: "strategic_bury"})
		result := p.HintOutput(m)
		var resObj controller.NinetyNineWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, []int{0, 1, 2}, resObj.Hint.BuryIndices)
	})

	t.Run("nil hint", func(t *testing.T) {
		m, _ := setupNinetyNineWebMockWithPlayers()
		m.On("GetHint").Return((*domain.NinetyNineHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.NinetyNineWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Nil(t, resObj.Hint)
	})
}

func TestNinetyNineWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.NinetyNineWebPresenter)
	m := setupNinetyNineWebMock()
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "bid", Detail: "test"},
	})
	assert.NotEmpty(t, p.ActionLogOutput(m))
}
