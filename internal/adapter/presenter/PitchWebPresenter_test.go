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

func setupPitchWebMock() *interfaces.MockPitchGame {
	m := new(interfaces.MockPitchGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetDealerIdx").Return(3)
	m.On("GetCurrentBid").Return(0)
	m.On("GetTrumpSuit").Return(0)
	m.On("GetBidWinnerIdx").Return(-1)
	m.On("GetCurrentTrick").Return([]*domain.PitchTrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.PitchPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultPitchConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetValidPlayIndices", 0).Return([]int{0, 1})
	return m
}

func setupPitchWebMockWithPlayers() (*interfaces.MockPitchGame, []*domain.PitchPlayer) {
	m := setupPitchWebMock()
	players := makePitchPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestPitchWebPresenter_Output(t *testing.T) {
	p := new(presenter.PitchWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupPitchWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))

		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
		var resObj controller.PitchWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, 4, len(resObj.Players))
		assert.Equal(t, 1, resObj.Phase) // PitchPhasePlay
		assert.Equal(t, 1, resObj.RoundNumber)
		assert.Equal(t, 3, resObj.DealerIdx)
		assert.Equal(t, -1, resObj.BidWinnerIdx)
		assert.Equal(t, 0, resObj.TrumpSuit)
		assert.Equal(t, []int{0, 1}, resObj.ValidPlayIndices)
		assert.Equal(t, "", resObj.Message)
	})

	t.Run("human cards shown, CPU cards hidden", func(t *testing.T) {
		m, players := setupPitchWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.PitchWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.True(t, resObj.Players[0].IsHuman)
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.False(t, resObj.Players[1].IsHuman)
		assert.Len(t, resObj.Players[1].Cards, 0)
		assert.Equal(t, 1, resObj.Players[1].CardCount)
	})

	t.Run("error message included", func(t *testing.T) {
		m, _ := setupPitchWebMockWithPlayers()
		result := p.Output(m, errors.New("invalid bid"))
		var resObj controller.PitchWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "invalid bid", resObj.Message)
	})

	t.Run("game end shows winner", func(t *testing.T) {
		m, _ := setupPitchWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)

		result := p.Output(m, nil)
		var resObj controller.PitchWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.True(t, resObj.GameEndFlag)
		assert.NotEmpty(t, resObj.MessageCode)
	})
}

func TestPitchWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.PitchWebPresenter)

	t.Run("nil hint", func(t *testing.T) {
		m, _ := setupPitchWebMockWithPlayers()
		m.On("GetHint").Return((*domain.PitchHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.PitchWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Nil(t, resObj.Hint)
	})

	t.Run("with hint", func(t *testing.T) {
		m, _ := setupPitchWebMockWithPlayers()
		bid := 3
		m.On("GetHint").Return(&domain.PitchHint{Bid: &bid, Reason: "bid_strong"})
		result := p.HintOutput(m)
		var resObj controller.PitchWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, "bid_strong", resObj.Hint.Reason)
		assert.Equal(t, 3, *resObj.Hint.Bid)
	})
}

func TestPitchWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.PitchWebPresenter)
	m := new(interfaces.MockPitchGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "bid", Detail: "You bid 3"},
	})
	out := p.ActionLogOutput(m)
	assert.Contains(t, out, "You bid 3")
}
