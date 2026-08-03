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

func makeViraPlayers() []*domain.ViraPlayer {
	return []*domain.ViraPlayer{
		domain.NewViraPlayer(true),
		domain.NewViraPlayer(false),
		domain.NewViraPlayer(false),
	}
}

func setupViraWebMock() *interfaces.MockViraGame {
	m := new(interfaces.MockViraGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.ViraPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetDeclarerIdx").Return(0)
	m.On("GetContract").Return(domain.ViraBidGask)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetPot").Return(30)
	m.On("GetLastRoundDelta").Return([domain.ViraPlayerCnt]int{0, 0, 0})
	m.On("GetLastRoundMade").Return(false)
	m.On("GetBids").Return([domain.ViraPlayerCnt]domain.ViraBid{
		domain.ViraBidGask, domain.ViraBidPass, domain.ViraBidPass,
	})
	m.On("GetBidDone").Return([domain.ViraPlayerCnt]bool{true, true, true})
	m.On("GetPlayerScores").Return([domain.ViraPlayerCnt]int{0, 0, 0})
	m.On("GetRoundTricks").Return([domain.ViraPlayerCnt]int{0, 0, 0})
	m.On("GetWinnerPlayer").Return(-1)
	m.On("GetPlayableIndices", 0).Return([]int{0})
	m.On("IsHumanTurn").Return(true)
	m.On("IsHumanBidTurn").Return(false)
	m.On("GetConfig").Return(domain.DefaultViraConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// **base だけに置く。**removeMockCall は最初の 1 件しか外さない。
	m.On("GetHint").Return(nil).Maybe()
	return m
}

func setupViraWebMockWithPlayers() (*interfaces.MockViraGame, []*domain.ViraPlayer) {
	m := setupViraWebMock()
	players := makeViraPlayers()
	m.On("GetPlayerCnt").Return(3)
	for i := 0; i < 3; i++ {
		m.On("GetPlayer", i).Return(players[i])
	}
	return m, players
}

func TestViraWebPresenter_Output(t *testing.T) {
	p := new(presenter.ViraWebPresenter)

	t.Run("initial state play phase lead", func(t *testing.T) {
		m, players := setupViraWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))

		result := p.Output(m, nil)
		var resObj controller.ViraWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.Players, 3)
		assert.Equal(t, int(domain.ViraPhasePlay), resObj.Phase)
		assert.Equal(t, -1, resObj.WinnerPlayer)
		assert.Equal(t, "vira.playPhase.lead", resObj.MessageCode)
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
		assert.True(t, resObj.Players[0].IsDeclarer)
		assert.False(t, resObj.Players[1].IsDeclarer)
		assert.Equal(t, []int{0}, resObj.PlayableIndices)
		assert.Equal(t, domain.CardDesignSpade, resObj.TrumpSuit)
		assert.Equal(t, int(domain.ViraBidGask), resObj.Contract)
		assert.Equal(t, int(domain.ViraBidGask), resObj.Bids[0])
	})

	// The pot is the number a Vira player tracks between rounds; the response has
	// to carry it or the UI shows a stake of zero after a carried-forward round.
	t.Run("pot is carried into the response", func(t *testing.T) {
		m, _ := setupViraWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPot")
		m.On("GetPot").Return(120)
		result := p.Output(m, nil)
		var resObj controller.ViraWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, 120, resObj.Pot)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupViraWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.ViraWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, int(domain.ViraCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, 6, resObj.Config.TargetRounds)
	})

	t.Run("bid phase message code", func(t *testing.T) {
		m, _ := setupViraWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsHumanBidTurn")
		m.On("GetPhase").Return(domain.ViraPhaseBid)
		m.On("IsHumanBidTurn").Return(true)
		result := p.Output(m, nil)
		var resObj controller.ViraWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "vira.bidPhase", resObj.MessageCode)
		assert.True(t, resObj.IsHumanBidTurn)
	})

	t.Run("play phase follow when trick has cards", func(t *testing.T) {
		m, _ := setupViraWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},
		})
		result := p.Output(m, nil)
		var resObj controller.ViraWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.CurrentTrick, 1)
		assert.Equal(t, "vira.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end message code", func(t *testing.T) {
		m, _ := setupViraWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.ViraPhaseTrickEnd)
		result := p.Output(m, nil)
		var resObj controller.ViraWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "vira.trickEnd", resObj.MessageCode)
	})

	t.Run("round end message code carries the settlement", func(t *testing.T) {
		m, _ := setupViraWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetLastRoundDelta")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetLastRoundMade")
		m.On("GetPhase").Return(domain.ViraPhaseRoundEnd)
		m.On("GetLastRoundDelta").Return([domain.ViraPlayerCnt]int{60, -30, -30})
		m.On("GetLastRoundMade").Return(true)
		result := p.Output(m, nil)
		var resObj controller.ViraWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "vira.roundEnd", resObj.MessageCode)
		assert.Equal(t, [domain.ViraPlayerCnt]int{60, -30, -30}, resObj.LastRoundDelta)
		assert.True(t, resObj.LastRoundMade)
	})

	t.Run("error message takes priority", func(t *testing.T) {
		m, _ := setupViraWebMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		var resObj controller.ViraWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "boom", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human wins", func(t *testing.T) {
		m, _ := setupViraWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(0)
		result := p.Output(m, nil)
		var resObj controller.ViraWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "vira.result.humanWin", resObj.MessageCode)
		assert.Nil(t, resObj.MessageParams)
	})

	t.Run("game end cpu wins", func(t *testing.T) {
		m, _ := setupViraWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(1)
		result := p.Output(m, nil)
		var resObj controller.ViraWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "vira.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"player": "1"}, resObj.MessageParams)
	})

	t.Run("player scores propagated to players", func(t *testing.T) {
		m, _ := setupViraWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayerScores")
		m.On("GetPlayerScores").Return([domain.ViraPlayerCnt]int{40, 20, 0})
		result := p.Output(m, nil)
		var resObj controller.ViraWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, 40, resObj.Players[0].Score)
		assert.Equal(t, 20, resObj.Players[1].Score)
	})

	t.Run("no playable indices outside the human play turn", func(t *testing.T) {
		m, _ := setupViraWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsHumanTurn")
		m.On("IsHumanTurn").Return(false)
		result := p.Output(m, nil)
		var resObj controller.ViraWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Empty(t, resObj.PlayableIndices)
	})
}

func TestViraWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.ViraWebPresenter)

	t.Run("hint with card indices", func(t *testing.T) {
		m, _ := setupViraWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.ViraHint{CardIndices: []int{2}, Reason: "follow_win"})
		result := p.HintOutput(m)
		var resObj controller.ViraWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, []int{2}, resObj.Hint.CardIndices)
		assert.Equal(t, "follow_win", resObj.Hint.Reason)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupViraWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.ViraHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.ViraWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Nil(t, resObj.Hint)
	})
}

func TestViraWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.ViraWebPresenter)
	m := new(interfaces.MockViraGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	assert.Contains(t, p.ActionLogOutput(m), `"actionType":"play"`)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestViraWebPresenterOutputCarriesTheHint(t *testing.T) {
	g, _ := setupViraWebMockWithPlayers()
	g.ExpectedCalls = removeMockCall(g.ExpectedCalls, "GetHint")
	g.On("GetHint").Return(&domain.ViraHint{CardIndices: []int{0}, Reason: "follow_suit"})

	result := new(presenter.ViraWebPresenter).Output(g, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	assert.NotContains(t, result, "vira.hintRequested")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestViraWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	g, _ := setupViraWebMockWithPlayers()
	g.ExpectedCalls = removeMockCall(g.ExpectedCalls, "GetHint")
	g.On("GetHint").Return(&domain.ViraHint{CardIndices: []int{0}, Reason: "follow_suit"})
	assert.Contains(t, new(presenter.ViraWebPresenter).HintOutput(g), "vira.hintRequested")

	none, _ := setupViraWebMockWithPlayers()
	none.ExpectedCalls = removeMockCall(none.ExpectedCalls, "GetHint")
	none.On("GetHint").Return((*domain.ViraHint)(nil))
	assert.Contains(t, new(presenter.ViraWebPresenter).HintOutput(none), "vira.noHint")
}
