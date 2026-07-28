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

func setupBeloteWebMock() *interfaces.MockBeloteGame {
	m := new(interfaces.MockBeloteGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.BelotePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTrumpSuit").Return(1)
	m.On("GetFaceUpCard").Return((*domain.Card)(nil))
	m.On("GetMakerTeam").Return(0)
	m.On("GetMakerPlayerIdx").Return(0)
	m.On("GetTeamScore", 0).Return(0)
	m.On("GetTeamScore", 1).Return(0)
	m.On("GetRoundPoints", 0).Return(0)
	m.On("GetRoundPoints", 1).Return(0)
	m.On("GetRoundBeloteBonus", 0).Return(0)
	m.On("GetRoundBeloteBonus", 1).Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultBeloteConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupBeloteWebMockWithPlayers() (*interfaces.MockBeloteGame, []*domain.BelotePlayer) {
	m := setupBeloteWebMock()
	players := makeBelotePlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestBeloteWebPresenter_Output(t *testing.T) {
	p := new(presenter.BeloteWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupBeloteWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))

		result := p.Output(m, nil)
		assert.NotEmpty(t, result)

		var resObj controller.BeloteWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Equal(t, 4, len(resObj.Players))
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, int(domain.BelotePhasePlay), resObj.Phase)
		assert.Equal(t, -1, resObj.WinnerTeam)
	})

	t.Run("CPU cards hidden", func(t *testing.T) {
		m, players := setupBeloteWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))

		result := p.Output(m, nil)
		var resObj controller.BeloteWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, 1, len(resObj.Players[0].Cards), "human cards visible")
		assert.Equal(t, 0, len(resObj.Players[1].Cards), "CPU cards hidden")
	})

	t.Run("last error message", func(t *testing.T) {
		m, _ := setupBeloteWebMockWithPlayers()
		result := p.Output(m, errors.New("invalid play"))
		var resObj controller.BeloteWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "invalid play", resObj.Message)
	})

	t.Run("game-end message", func(t *testing.T) {
		m, _ := setupBeloteWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(1)

		result := p.Output(m, nil)
		var resObj controller.BeloteWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Contains(t, resObj.Message, "ゲーム終了")
		assert.Equal(t, 1, resObj.WinnerTeam)
	})

	t.Run("each phase produces message code", func(t *testing.T) {
		cases := []struct {
			phase domain.BelotePhase
			code  string
		}{
			{domain.BelotePhaseBidPickUp, "belote.pickUpPhase"},
			{domain.BelotePhaseBidCallTrump, "belote.callTrumpPhase"},
			{domain.BelotePhasePlay, "belote.playPhase.lead"},
			{domain.BelotePhaseTrickEnd, "belote.trickEnd"},
			{domain.BelotePhaseRoundEnd, "belote.roundEnd"},
		}
		for _, tc := range cases {
			m, _ := setupBeloteWebMockWithPlayers()
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
			m.On("GetPhase").Return(tc.phase)
			result := p.Output(m, nil)
			var resObj controller.BeloteWebOutput
			_ = json.Unmarshal([]byte(result), &resObj)
			assert.Equal(t, tc.code, resObj.MessageCode, "phase %d", tc.phase)
		}
	})
}

func TestBeloteWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.BeloteWebPresenter)

	t.Run("nil hint", func(t *testing.T) {
		m, _ := setupBeloteWebMockWithPlayers()
		m.On("GetHint").Return((*domain.BeloteHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.BeloteWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Nil(t, resObj.Hint)
	})

	t.Run("with hint", func(t *testing.T) {
		m, _ := setupBeloteWebMockWithPlayers()
		ok := true
		m.On("GetHint").Return(&domain.BeloteHint{OrderUp: &ok, Reason: "strategic_pickup"})

		result := p.HintOutput(m)
		var resObj controller.BeloteWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.NotNil(t, resObj.Hint)
		assert.NotNil(t, resObj.Hint.OrderUp)
		assert.Equal(t, "strategic_pickup", resObj.Hint.Reason)
	})
}

func TestBeloteWebPresenter_ActionLogOutput(t *testing.T) {
	m, _ := setupBeloteWebMockWithPlayers()
	p := new(presenter.BeloteWebPresenter)
	result := p.ActionLogOutput(m)
	assert.NotEmpty(t, result)
}
