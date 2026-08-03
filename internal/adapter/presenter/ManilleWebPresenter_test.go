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

func setupManilleWebMock() *interfaces.MockManilleGame {
	m := new(interfaces.MockManilleGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.ManillePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetTeamScores").Return([domain.ManilleTeamCnt]int{0, 0})
	m.On("GetRoundCardPoints").Return([domain.ManilleTeamCnt]int{0, 0})
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetPlayableIndices", 0).Return([]int{0})
	m.On("IsHumanTurn").Return(true)
	m.On("GetConfig").Return(domain.DefaultManilleConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	// **base だけに置く。**removeMockCall は最初の 1 件しか外さない。
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func setupManilleWebMockWithPlayers() (*interfaces.MockManilleGame, []*domain.ManillePlayer) {
	m := setupManilleWebMock()
	players := makeManillePlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestManilleWebPresenter_Output(t *testing.T) {
	p := new(presenter.ManilleWebPresenter)

	t.Run("initial state play phase lead", func(t *testing.T) {
		m, players := setupManilleWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))

		result := p.Output(m, nil)
		var resObj controller.ManilleWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.Players, 4)
		assert.Equal(t, int(domain.ManillePhasePlay), resObj.Phase)
		assert.Equal(t, -1, resObj.WinnerTeam)
		assert.Equal(t, "manille.playPhase.lead", resObj.MessageCode)
		// human cards visible, CPU hidden
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
		assert.Equal(t, []int{0}, resObj.PlayableIndices)
		assert.Equal(t, domain.CardDesignSpade, resObj.TrumpSuit)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupManilleWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.ManilleWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, int(domain.ManilleCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, 101, resObj.Config.TargetPoints)
	})

	t.Run("play phase follow when trick has cards", func(t *testing.T) {
		m, _ := setupManilleWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetPhase").Return(domain.ManillePhasePlay)
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},
		})
		result := p.Output(m, nil)
		var resObj controller.ManilleWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.CurrentTrick, 1)
		assert.Equal(t, "manille.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end message code", func(t *testing.T) {
		m, _ := setupManilleWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.ManillePhaseTrickEnd)
		result := p.Output(m, nil)
		var resObj controller.ManilleWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "manille.trickEnd", resObj.MessageCode)
	})

	t.Run("round end message code", func(t *testing.T) {
		m, _ := setupManilleWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.ManillePhaseRoundEnd)
		result := p.Output(m, nil)
		var resObj controller.ManilleWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "manille.roundEnd", resObj.MessageCode)
	})

	t.Run("error message takes priority", func(t *testing.T) {
		m, _ := setupManilleWebMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		var resObj controller.ManilleWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "boom", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human team wins", func(t *testing.T) {
		m, _ := setupManilleWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		// player 0 is human, team 0; winner = 0
		m.On("GetWinnerTeam").Return(0)
		result := p.Output(m, nil)
		var resObj controller.ManilleWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "manille.result.humanWin", resObj.MessageCode)
		assert.Nil(t, resObj.MessageParams)
	})

	t.Run("game end cpu team wins", func(t *testing.T) {
		m, _ := setupManilleWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		// player 0 is human, team 0; winner = team 1 (CPU)
		m.On("GetWinnerTeam").Return(1)
		result := p.Output(m, nil)
		var resObj controller.ManilleWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "manille.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"team": "1"}, resObj.MessageParams)
	})

	t.Run("team scores propagated to players", func(t *testing.T) {
		m, _ := setupManilleWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTeamScores")
		m.On("GetTeamScores").Return([domain.ManilleTeamCnt]int{30, 15})
		result := p.Output(m, nil)
		var resObj controller.ManilleWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		// player 0 is team 0, score 30
		assert.Equal(t, 30, resObj.Players[0].TeamScore)
		// player 1 is team 1, score 15
		assert.Equal(t, 15, resObj.Players[1].TeamScore)
	})
}

func TestManilleWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.ManilleWebPresenter)

	t.Run("hint with card indices", func(t *testing.T) {
		m, _ := setupManilleWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.ManilleHint{CardIndices: []int{2}, Reason: "follow_win"})
		result := p.HintOutput(m)
		var resObj controller.ManilleWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, []int{2}, resObj.Hint.CardIndices)
		assert.Equal(t, "follow_win", resObj.Hint.Reason)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupManilleWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.ManilleHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.ManilleWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Nil(t, resObj.Hint)
	})
}

func TestManilleWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.ManilleWebPresenter)
	m := new(interfaces.MockManilleGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, `"actionType":"play"`)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestManilleWebPresenterOutputCarriesTheHint(t *testing.T) {
	mng, _ := setupManilleWebMockWithPlayers()
	mng.ExpectedCalls = removeMockCall(mng.ExpectedCalls, "GetHint")
	mng.On("GetHint").Return(&domain.ManilleHint{CardIndices: []int{0}, Reason: "follow_suit"})

	result := new(presenter.ManilleWebPresenter).Output(mng, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	// **Output は「頼んだヒント」の印を付けない。**付けると CLI が毎回 HINT 行を出す。
	assert.NotContains(t, result, "manille.hintRequested")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**このゲーム群の
// hintAvailable は画面のラベルとして埋まっているので別キーを使う (#4483)。
func TestManilleWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	mng, _ := setupManilleWebMockWithPlayers()
	mng.ExpectedCalls = removeMockCall(mng.ExpectedCalls, "GetHint")
	mng.On("GetHint").Return(&domain.ManilleHint{CardIndices: []int{0}, Reason: "follow_suit"})
	assert.Contains(t, new(presenter.ManilleWebPresenter).HintOutput(mng), "manille.hintRequested")

	none, _ := setupManilleWebMockWithPlayers()
	none.ExpectedCalls = removeMockCall(none.ExpectedCalls, "GetHint")
	none.On("GetHint").Return((*domain.ManilleHint)(nil))
	assert.Contains(t, new(presenter.ManilleWebPresenter).HintOutput(none), "manille.noHint")
}
