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

func makeMarjapussiWebPlayers() []*domain.MarjapussiPlayer {
	return []*domain.MarjapussiPlayer{
		domain.NewMarjapussiPlayer(true),
		domain.NewMarjapussiPlayer(false),
		domain.NewMarjapussiPlayer(false),
		domain.NewMarjapussiPlayer(false),
	}
}

func setupMarjapussiWebMock() *interfaces.MockMarjapussiGame {
	m := new(interfaces.MockMarjapussiGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.MarjapussiPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetPlayerScores").Return([domain.MarjapussiPlayerCnt]int{0, 0, 0, 0})
	m.On("GetTeamScores").Return([domain.MarjapussiTeamCnt]int{0, 0})
	m.On("GetRoundCardPoints").Return([domain.MarjapussiTeamCnt]int{0, 0})
	m.On("GetRoundMarriage").Return([domain.MarjapussiTeamCnt]int{0, 0})
	m.On("GetPussi").Return(([]*domain.Card)(nil))
	m.On("GetWinnerPlayer").Return(-1)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetPlayableIndices", 0).Return([]int{0})
	m.On("IsHumanTurn").Return(true)
	m.On("GetConfig").Return(domain.DefaultMarjapussiConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	// **base だけに置く。**removeMockCall は最初の 1 件しか外さない。
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func setupMarjapussiWebMockWithPlayers() (*interfaces.MockMarjapussiGame, []*domain.MarjapussiPlayer) {
	m := setupMarjapussiWebMock()
	players := makeMarjapussiWebPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestMarjapussiWebPresenter_Output(t *testing.T) {
	p := new(presenter.MarjapussiWebPresenter)

	t.Run("initial state play phase lead", func(t *testing.T) {
		m, players := setupMarjapussiWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))

		result := p.Output(m, nil)
		var resObj controller.MarjapussiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.Players, 4)
		assert.Equal(t, int(domain.MarjapussiPhasePlay), resObj.Phase)
		assert.Equal(t, -1, resObj.WinnerPlayer)
		assert.Equal(t, -1, resObj.WinnerTeam)
		assert.Equal(t, -1, resObj.LastTrickWinner)
		assert.Equal(t, -1, resObj.PussiWinnerTeam)
		assert.Equal(t, "marjapussi.playPhase.lead", resObj.MessageCode)
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
		assert.Equal(t, 0, resObj.Players[0].TeamID)
		assert.Equal(t, 1, resObj.Players[1].TeamID)
		assert.Equal(t, []int{0}, resObj.PlayableIndices)
		assert.Equal(t, domain.CardDesignSpade, resObj.TrumpSuit)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupMarjapussiWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.MarjapussiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, int(domain.MarjapussiCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, domain.MarjapussiDefaultPointLimit, resObj.Config.TargetPoints)
	})

	t.Run("play phase follow when trick has cards", func(t *testing.T) {
		m, _ := setupMarjapussiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetPhase").Return(domain.MarjapussiPhasePlay)
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},
		})
		result := p.Output(m, nil)
		var resObj controller.MarjapussiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.CurrentTrick, 1)
		assert.Equal(t, "marjapussi.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end message code", func(t *testing.T) {
		m, _ := setupMarjapussiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MarjapussiPhaseTrickEnd)
		result := p.Output(m, nil)
		var resObj controller.MarjapussiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "marjapussi.trickEnd", resObj.MessageCode)
	})

	t.Run("round end message code and pussi reveal", func(t *testing.T) {
		m, _ := setupMarjapussiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MarjapussiPhaseRoundEnd)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPussi")
		pussi := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignHeart, 10, false),
		}
		m.On("GetPussi").Return(pussi)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetActionLog")
		m.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{ActionType: "pussi_win", PlayerIdx: 2, Detail: "team 0 wins the pussi (+21)"},
		})

		result := p.Output(m, nil)
		var resObj controller.MarjapussiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "marjapussi.roundEnd", resObj.MessageCode)
		assert.Equal(t, 2, resObj.PussiCount)
		assert.Len(t, resObj.Pussi, 2)
		assert.Equal(t, 2, resObj.LastTrickWinner)
		assert.Equal(t, 0, resObj.PussiWinnerTeam)
	})

	t.Run("error message takes priority", func(t *testing.T) {
		m, _ := setupMarjapussiWebMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		var resObj controller.MarjapussiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "boom", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human wins", func(t *testing.T) {
		m, _ := setupMarjapussiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)
		result := p.Output(m, nil)
		var resObj controller.MarjapussiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "marjapussi.result.humanWin", resObj.MessageCode)
		assert.Nil(t, resObj.MessageParams)
	})

	t.Run("game end cpu wins", func(t *testing.T) {
		m, _ := setupMarjapussiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(1)
		result := p.Output(m, nil)
		var resObj controller.MarjapussiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "marjapussi.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"team": "1"}, resObj.MessageParams)
	})

	t.Run("player scores propagated to players", func(t *testing.T) {
		m, _ := setupMarjapussiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayerScores")
		m.On("GetPlayerScores").Return([domain.MarjapussiPlayerCnt]int{40, 20, 40, 20})
		result := p.Output(m, nil)
		var resObj controller.MarjapussiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, 40, resObj.Players[0].Score)
		assert.Equal(t, 20, resObj.Players[1].Score)
		assert.Equal(t, 40, resObj.Players[2].Score)
		assert.Equal(t, 20, resObj.Players[3].Score)
	})
}

func TestMarjapussiWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.MarjapussiWebPresenter)

	t.Run("hint with card indices", func(t *testing.T) {
		m, _ := setupMarjapussiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.MarjapussiHint{CardIndices: []int{2}, Reason: "follow_win"})
		result := p.HintOutput(m)
		var resObj controller.MarjapussiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, []int{2}, resObj.Hint.CardIndices)
		assert.Equal(t, "follow_win", resObj.Hint.Reason)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupMarjapussiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.MarjapussiHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.MarjapussiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Nil(t, resObj.Hint)
	})
}

func TestMarjapussiWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.MarjapussiWebPresenter)
	m := new(interfaces.MockMarjapussiGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, `"actionType":"play"`)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestMarjapussiWebPresenterOutputCarriesTheHint(t *testing.T) {
	tsg, _ := setupMarjapussiWebMockWithPlayers()
	tsg.ExpectedCalls = removeMockCall(tsg.ExpectedCalls, "GetHint")
	tsg.On("GetHint").Return(&domain.MarjapussiHint{CardIndices: []int{0}, Reason: "follow_suit"})

	result := new(presenter.MarjapussiWebPresenter).Output(tsg, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	// **Output は「頼んだヒント」の印を付けない。**付けると CLI が毎回 HINT 行を出す。
	assert.NotContains(t, result, "marjapussi.hintRequested")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestMarjapussiWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	tsg, _ := setupMarjapussiWebMockWithPlayers()
	tsg.ExpectedCalls = removeMockCall(tsg.ExpectedCalls, "GetHint")
	tsg.On("GetHint").Return(&domain.MarjapussiHint{CardIndices: []int{0}, Reason: "follow_suit"})
	assert.Contains(t, new(presenter.MarjapussiWebPresenter).HintOutput(tsg), "marjapussi.hintRequested")

	none, _ := setupMarjapussiWebMockWithPlayers()
	none.ExpectedCalls = removeMockCall(none.ExpectedCalls, "GetHint")
	none.On("GetHint").Return((*domain.MarjapussiHint)(nil))
	assert.Contains(t, new(presenter.MarjapussiWebPresenter).HintOutput(none), "marjapussi.noHint")
}
