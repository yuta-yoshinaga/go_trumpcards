//go:build test

package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupBinokelWebMock() *interfaces.MockBinokelGame {
	m := new(interfaces.MockBinokelGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.BinokelPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTrumpSuit").Return(1)
	m.On("GetHighestBid").Return(150)
	m.On("GetHighestBidder").Return(0)
	m.On("GetScores").Return([domain.BinokelPlayerCnt]int{0, 0, 0})
	m.On("GetScore", 0).Return(0)
	m.On("GetScore", 1).Return(0)
	m.On("GetScore", 2).Return(0)
	m.On("GetWinnerPlayer").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultBinokelConfig())
	m.On("GetPlayerMelds").Return([domain.BinokelPlayerCnt][]*domain.BinokelMeld{})
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetDabb").Return([]*domain.Card(nil))
	m.On("GetDabbDiscarded").Return([]*domain.Card(nil))
	m.On("IsHumanTurn").Return(true)
	m.On("GetValidPlayIndices", 0).Return([]int{0, 1})
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func setupBinokelWebMockWithPlayers() (*interfaces.MockBinokelGame, []*domain.BinokelPlayer) {
	m := setupBinokelWebMock()
	players := []*domain.BinokelPlayer{
		domain.NewBinokelPlayer(true),
		domain.NewBinokelPlayer(false),
		domain.NewBinokelPlayer(false),
	}
	m.On("GetPlayerCnt").Return(3)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	return m, players
}

func TestBinokelWebPresenter_Output(t *testing.T) {
	p := new(presenter.BinokelWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupBinokelWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))

		result := p.Output(m, nil)
		assert.NotEmpty(t, result)

		var resObj controller.BinokelWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(resObj.Players))
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, int(domain.BinokelPhasePlay), resObj.Phase)
		assert.Equal(t, -1, resObj.WinnerPlayer)
	})

	t.Run("human cards shown, CPU cards hidden", func(t *testing.T) {
		m, players := setupBinokelWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))

		result := p.Output(m, nil)
		var resObj controller.BinokelWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, 1, len(resObj.Players[0].Cards))
		assert.Equal(t, 0, len(resObj.Players[1].Cards))
	})

	t.Run("error message", func(t *testing.T) {
		m, _ := setupBinokelWebMockWithPlayers()

		result := p.Output(m, errors.New("test error"))
		var resObj controller.BinokelWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "test error", resObj.Message)
	})

	t.Run("game end message", func(t *testing.T) {
		m, _ := setupBinokelWebMockWithPlayers()
		m.ExpectedCalls = nil
		m.On("GetRoundNumber").Return(1)
		m.On("GetTrickNumber").Return(15)
		m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
		m.On("GetGameEndFlag").Return(true)
		m.On("GetPhase").Return(domain.BinokelPhaseGameEnd)
		m.On("GetCurrentPlayerIdx").Return(0)
		m.On("GetBidPlayerIdx").Return(0)
		m.On("GetDealerIdx").Return(0)
		m.On("GetTrumpSuit").Return(1)
		m.On("GetHighestBid").Return(200)
		m.On("GetHighestBidder").Return(0)
		m.On("GetScores").Return([domain.BinokelPlayerCnt]int{1000, 800, 600})
		m.On("GetScore", 0).Return(1000)
		m.On("GetScore", 1).Return(800)
		m.On("GetScore", 2).Return(600)
		m.On("GetWinnerPlayer").Return(0)
		m.On("GetLeadPlayerIdx").Return(0)
		m.On("GetConfig").Return(domain.DefaultBinokelConfig())
		m.On("GetPlayerMelds").Return([domain.BinokelPlayerCnt][]*domain.BinokelMeld{})
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
		m.On("GetDabb").Return([]*domain.Card(nil))
		m.On("GetDabbDiscarded").Return([]*domain.Card(nil))
		m.On("IsHumanTurn").Return(false)
		m.On("GetPlayerCnt").Return(3)
		m.On("GetPlayer", 0).Return(domain.NewBinokelPlayer(true))
		m.On("GetPlayer", 1).Return(domain.NewBinokelPlayer(false))
		m.On("GetPlayer", 2).Return(domain.NewBinokelPlayer(false))
		m.On("GetHint").Return(nil).Maybe()

		result := p.Output(m, nil)
		var resObj controller.BinokelWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, 0, resObj.WinnerPlayer)
		assert.Contains(t, resObj.MessageCode, "binokel.result.player0Win")
	})

	t.Run("valid play indices included in play phase", func(t *testing.T) {
		m, _ := setupBinokelWebMockWithPlayers()

		result := p.Output(m, nil)
		var resObj controller.BinokelWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, []int{0, 1}, resObj.ValidPlayIndices)
	})

	t.Run("dabb and dabbDiscarded included", func(t *testing.T) {
		m, _ := setupBinokelWebMockWithPlayers()
		dabbCards := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignHeart, 10, false),
			domain.NewCard(domain.CardDesignDiamond, 11, false),
		}
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDabb")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDabbDiscarded")
		m.On("GetDabb").Return(dabbCards)
		m.On("GetDabbDiscarded").Return(dabbCards)

		result := p.Output(m, nil)
		var resObj controller.BinokelWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Len(t, resObj.Dabb, 3)
		assert.Len(t, resObj.DabbDiscarded, 3)
	})

	t.Run("players bid and hasPassed state distinguishing unbid, passed, and declared", func(t *testing.T) {
		m, players := setupBinokelWebMockWithPlayers()
		// Player 0 (Human): has not bid yet
		players[0].SetBid(0)
		players[0].SetHasPassed(false)
		// Player 1 (CPU 1): has passed
		players[1].SetBid(0)
		players[1].SetHasPassed(true)
		// Player 2 (CPU 2): declared 150
		players[2].SetBid(150)
		players[2].SetHasPassed(false)

		result := p.Output(m, nil)
		var resObj controller.BinokelWebOutput
		require.NoError(t, json.Unmarshal([]byte(result), &resObj))
		require.Len(t, resObj.Players, 3)

		// Unbid seat
		assert.Equal(t, 0, resObj.Players[0].Bid)
		assert.False(t, resObj.Players[0].HasPassed)

		// Passed seat
		assert.Equal(t, 0, resObj.Players[1].Bid)
		assert.True(t, resObj.Players[1].HasPassed)

		// Declared seat
		assert.Equal(t, 150, resObj.Players[2].Bid)
		assert.False(t, resObj.Players[2].HasPassed)
	})
}

func TestBinokelWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.BinokelWebPresenter)

	t.Run("returns hint", func(t *testing.T) {
		m, _ := setupBinokelWebMockWithPlayers()
		bid := 150
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.BinokelHint{BidAmount: &bid, Reason: "hint_bid"})

		result := p.HintOutput(m)
		assert.Contains(t, result, "hint_bid")
	})

	t.Run("returns nil hint", func(t *testing.T) {
		m, _ := setupBinokelWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.BinokelHint)(nil))

		result := p.HintOutput(m)
		assert.Contains(t, result, "binokel.noHint")
	})
}

func TestBinokelWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.BinokelWebPresenter)
	m := setupBinokelWebMock()

	result := p.ActionLogOutput(m)
	assert.NotEmpty(t, result)
}

func TestBinokelWebPresenterOutputCarriesTheHint(t *testing.T) {
	idx := 0
	png, _ := setupBinokelWebMockWithPlayers()
	png.ExpectedCalls = removeMockCall(png.ExpectedCalls, "GetHint")
	png.On("GetHint").Return(&domain.BinokelHint{CardIndex: &idx, Reason: "follow_suit"})

	result := new(presenter.BinokelWebPresenter).Output(png, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	assert.NotContains(t, result, "binokel.hintRequested")
}

func TestBinokelWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	idx := 0
	png, _ := setupBinokelWebMockWithPlayers()
	png.ExpectedCalls = removeMockCall(png.ExpectedCalls, "GetHint")
	png.On("GetHint").Return(&domain.BinokelHint{CardIndex: &idx, Reason: "follow_suit"})
	got := new(presenter.BinokelWebPresenter).HintOutput(png)
	assert.Contains(t, got, "binokel.hintRequested")
	assert.Contains(t, got, `"hint"`)
	assert.Contains(t, got, `"players"`, "HintOutput must return a state response, not a bare hint")

	none, _ := setupBinokelWebMockWithPlayers()
	none.ExpectedCalls = removeMockCall(none.ExpectedCalls, "GetHint")
	none.On("GetHint").Return((*domain.BinokelHint)(nil))
	assert.Contains(t, new(presenter.BinokelWebPresenter).HintOutput(none), "binokel.noHint")
}

func TestBinokelWebPresenterOutputCarriesTheMeldTable(t *testing.T) {
	p := new(presenter.BinokelWebPresenter)
	m, _ := setupBinokelWebMockWithPlayers()

	var resObj controller.BinokelWebOutput
	require.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &resObj))

	table := domain.BinokelMeldTable()
	require.Len(t, resObj.MeldTable, len(table))
	for i, e := range table {
		assert.Equal(t, int(e.Type), resObj.MeldTable[i].Type, "entry %d type", i)
		assert.Equal(t, e.Points, resObj.MeldTable[i].Points, "entry %d points", i)
	}
}
