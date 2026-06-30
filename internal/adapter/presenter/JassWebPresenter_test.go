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

func setupJassWebMock() *interfaces.MockJassGame {
	m := new(interfaces.MockJassGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.JassTrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.JassPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetForehandIdx").Return(1)
	m.On("GetTrumpSuit").Return(1)
	m.On("GetSchieben").Return(false)
	m.On("GetMakerTeam").Return(0)
	m.On("GetMakerPlayerIdx").Return(0)
	m.On("GetTeamScore", 0).Return(0)
	m.On("GetTeamScore", 1).Return(0)
	m.On("GetRoundPoints", 0).Return(0)
	m.On("GetRoundPoints", 1).Return(0)
	m.On("GetRoundWeisPoints", 0).Return(0)
	m.On("GetRoundWeisPoints", 1).Return(0)
	m.On("GetRoundStockPoints", 0).Return(0)
	m.On("GetRoundStockPoints", 1).Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultJassConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupJassWebMockWithPlayers() (*interfaces.MockJassGame, []*domain.JassPlayer) {
	m := setupJassWebMock()
	players := makeJassPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestJassWebPresenter_Output(t *testing.T) {
	p := new(presenter.JassWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupJassWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))

		result := p.Output(m, nil)
		var resObj controller.JassWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, 4, len(resObj.Players))
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, int(domain.JassPhasePlay), resObj.Phase)
		assert.Equal(t, -1, resObj.WinnerTeam)
	})

	t.Run("with error", func(t *testing.T) {
		m, _ := setupJassWebMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		var resObj controller.JassWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "boom", resObj.Message)
	})

	t.Run("bid trump phase", func(t *testing.T) {
		m, _ := setupJassWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.JassPhaseBidTrump)
		result := p.Output(m, nil)
		assert.Contains(t, result, "jass.bidTrumpPhase")
	})

	t.Run("game end", func(t *testing.T) {
		m, _ := setupJassWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)
		result := p.Output(m, nil)
		assert.Contains(t, result, "jass.result.team0Win")
	})
}

func TestJassWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.JassWebPresenter)
	m, _ := setupJassWebMockWithPlayers()
	suit := 2
	m.On("GetHint").Return(&domain.JassHint{Suit: &suit, Reason: "strategic_trump"})
	result := p.HintOutput(m)
	var resObj controller.JassWebOutput
	assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
	assert.NotNil(t, resObj.Hint)
	assert.Equal(t, "strategic_trump", resObj.Hint.Reason)
}

func TestJassWebPresenter_HintOutput_Nil(t *testing.T) {
	p := new(presenter.JassWebPresenter)
	m, _ := setupJassWebMockWithPlayers()
	m.On("GetHint").Return((*domain.JassHint)(nil))
	result := p.HintOutput(m)
	var resObj controller.JassWebOutput
	assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
	assert.Nil(t, resObj.Hint)
}

func TestJassWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.JassWebPresenter)
	m := setupJassWebMock()
	assert.NotNil(t, p.ActionLogOutput(m))
}
