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

func makeOhHellPlayers() []*domain.OhHellPlayer {
	return []*domain.OhHellPlayer{
		domain.NewOhHellPlayer(true),
		domain.NewOhHellPlayer(false),
		domain.NewOhHellPlayer(false),
		domain.NewOhHellPlayer(false),
	}
}

func setupOhHellWebMock() *interfaces.MockOhHellGame {
	m := new(interfaces.MockOhHellGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTotalRounds").Return(19)
	m.On("GetHandSize").Return(10)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.OhHellPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTrumpCard").Return(domain.NewCard(domain.CardDesignHeart, 5, false))
	m.On("GetTrumpSuit").Return(domain.CardDesignHeart)
	m.On("GetRestrictedBid").Return(-1)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultOhHellConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	// **base だけに置く。**removeMockCall は最初の 1 件しか外さない。
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func setupOhHellWebMockWithPlayers() (*interfaces.MockOhHellGame, []*domain.OhHellPlayer) {
	m := setupOhHellWebMock()
	players := makeOhHellPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestOhHellWebPresenter_Output(t *testing.T) {
	p := new(presenter.OhHellWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupOhHellWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))

		result := p.Output(m, nil)
		assert.NotEmpty(t, result)

		var resObj controller.OhHellWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Equal(t, 4, len(resObj.Players))
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, 1, resObj.Phase)
		assert.Equal(t, 1, resObj.RoundNumber)
		assert.Equal(t, 19, resObj.TotalRounds)
		assert.Equal(t, 10, resObj.HandSize)
		assert.Equal(t, domain.CardDesignHeart, resObj.TrumpSuit)
		assert.NotNil(t, resObj.TrumpCard)
		assert.Equal(t, -1, resObj.RestrictedBid)
		assert.Equal(t, -1, resObj.WinnerIdx)
	})

	t.Run("error message", func(t *testing.T) {
		m, _ := setupOhHellWebMockWithPlayers()

		result := p.Output(m, errors.New("test error"))
		var resObj controller.OhHellWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "test error", resObj.Message)
	})

	t.Run("bid phase message", func(t *testing.T) {
		m, _ := setupOhHellWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.OhHellPhaseBid)

		result := p.Output(m, nil)
		var resObj controller.OhHellWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "ohhell.bidPhase", resObj.MessageCode)
	})

	t.Run("play phase lead message", func(t *testing.T) {
		m, _ := setupOhHellWebMockWithPlayers()

		result := p.Output(m, nil)
		var resObj controller.OhHellWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "ohhell.playPhase.lead", resObj.MessageCode)
	})

	t.Run("play phase follow message", func(t *testing.T) {
		m, _ := setupOhHellWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 3, false)},
		})

		result := p.Output(m, nil)
		var resObj controller.OhHellWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "ohhell.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end message", func(t *testing.T) {
		m, _ := setupOhHellWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.OhHellPhaseTrickEnd)

		result := p.Output(m, nil)
		var resObj controller.OhHellWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "ohhell.trickEnd", resObj.MessageCode)
	})

	t.Run("round end message", func(t *testing.T) {
		m, _ := setupOhHellWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.OhHellPhaseRoundEnd)

		result := p.Output(m, nil)
		var resObj controller.OhHellWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "ohhell.roundEnd", resObj.MessageCode)
	})

	t.Run("game end message", func(t *testing.T) {
		m, players := setupOhHellWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		players[0].SetBid(3)

		result := p.Output(m, nil)
		var resObj controller.OhHellWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.True(t, resObj.GameEndFlag)
		assert.NotEmpty(t, resObj.Message)
	})
}

func TestOhHellWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.OhHellWebPresenter)

	t.Run("with hint", func(t *testing.T) {
		m, _ := setupOhHellWebMockWithPlayers()
		bid := 3
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.OhHellHint{Bid: &bid, Reason: "strategic_bid"})

		result := p.HintOutput(m)
		var resObj controller.OhHellWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, 3, *resObj.Hint.Bid)
	})

	t.Run("nil hint", func(t *testing.T) {
		m, _ := setupOhHellWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.OhHellHint)(nil))

		result := p.HintOutput(m)
		var resObj controller.OhHellWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Nil(t, resObj.Hint)
	})
}

func TestOhHellWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.OhHellWebPresenter)
	m := setupOhHellWebMock()
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "bid", Detail: "test"},
	})

	result := p.ActionLogOutput(m)
	assert.NotEmpty(t, result)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestOhHellWebPresenterOutputCarriesTheHint(t *testing.T) {
	idx := 0
	ohg, _ := setupOhHellWebMockWithPlayers()
	ohg.ExpectedCalls = removeMockCall(ohg.ExpectedCalls, "GetHint")
	ohg.On("GetHint").Return(&domain.OhHellHint{CardIndex: &idx, Reason: "follow_suit"})

	result := new(presenter.OhHellWebPresenter).Output(ohg, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	assert.NotContains(t, result, "ohhell.hintRequested")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestOhHellWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	idx := 0
	ohg, _ := setupOhHellWebMockWithPlayers()
	ohg.ExpectedCalls = removeMockCall(ohg.ExpectedCalls, "GetHint")
	ohg.On("GetHint").Return(&domain.OhHellHint{CardIndex: &idx, Reason: "follow_suit"})
	assert.Contains(t, new(presenter.OhHellWebPresenter).HintOutput(ohg), "ohhell.hintRequested")

	none, _ := setupOhHellWebMockWithPlayers()
	none.ExpectedCalls = removeMockCall(none.ExpectedCalls, "GetHint")
	none.On("GetHint").Return((*domain.OhHellHint)(nil))
	assert.Contains(t, new(presenter.OhHellWebPresenter).HintOutput(none), "ohhell.noHint")
}
