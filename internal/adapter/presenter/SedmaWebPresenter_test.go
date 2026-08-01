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

func setupSedmaWebMock() *interfaces.MockSedmaGame {
	m := new(interfaces.MockSedmaGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.SedmaPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTeamScores").Return([domain.SedmaTeamCnt]int{0, 0})
	m.On("GetRoundCardPoints").Return([domain.SedmaTeamCnt]int{0, 0})
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetPlayableIndices", 0).Return([]int{0})
	m.On("IsHumanTurn").Return(true)
	m.On("GetConfig").Return(domain.DefaultSedmaConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func setupSedmaWebMockWithPlayers() (*interfaces.MockSedmaGame, []*domain.SedmaPlayer) {
	m := setupSedmaWebMock()
	players := makeSedmaPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestSedmaWebPresenter_Output(t *testing.T) {
	p := new(presenter.SedmaWebPresenter)

	t.Run("initial state play phase lead", func(t *testing.T) {
		m, players := setupSedmaWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))

		result := p.Output(m, nil)
		var resObj controller.SedmaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.Players, 4)
		assert.Equal(t, int(domain.SedmaPhasePlay), resObj.Phase)
		assert.Equal(t, -1, resObj.WinnerTeam)
		assert.Equal(t, "sedma.playPhase.lead", resObj.MessageCode)
		// human cards visible, CPU hidden
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
		assert.Equal(t, []int{0}, resObj.PlayableIndices)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupSedmaWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.SedmaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, int(domain.SedmaCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, 101, resObj.Config.TargetPoints)
	})

	t.Run("play phase follow when trick has cards", func(t *testing.T) {
		m, _ := setupSedmaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetPhase").Return(domain.SedmaPhasePlay)
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},
		})
		result := p.Output(m, nil)
		var resObj controller.SedmaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.CurrentTrick, 1)
		assert.Equal(t, "sedma.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end message code", func(t *testing.T) {
		m, _ := setupSedmaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SedmaPhaseTrickEnd)
		result := p.Output(m, nil)
		var resObj controller.SedmaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "sedma.trickEnd", resObj.MessageCode)
	})

	t.Run("round end message code", func(t *testing.T) {
		m, _ := setupSedmaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SedmaPhaseRoundEnd)
		result := p.Output(m, nil)
		var resObj controller.SedmaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "sedma.roundEnd", resObj.MessageCode)
	})

	t.Run("error message takes priority", func(t *testing.T) {
		m, _ := setupSedmaWebMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		var resObj controller.SedmaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "boom", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human team wins", func(t *testing.T) {
		m, _ := setupSedmaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		// player 0 is human, team 0; winner = 0
		m.On("GetWinnerTeam").Return(0)
		result := p.Output(m, nil)
		var resObj controller.SedmaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "sedma.result.humanWin", resObj.MessageCode)
		assert.Nil(t, resObj.MessageParams)
	})

	t.Run("game end cpu team wins", func(t *testing.T) {
		m, _ := setupSedmaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		// player 0 is human, team 0; winner = team 1 (CPU)
		m.On("GetWinnerTeam").Return(1)
		result := p.Output(m, nil)
		var resObj controller.SedmaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "sedma.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"team": "B"}, resObj.MessageParams)
	})

	t.Run("team scores propagated to players", func(t *testing.T) {
		m, _ := setupSedmaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTeamScores")
		m.On("GetTeamScores").Return([domain.SedmaTeamCnt]int{30, 15})
		result := p.Output(m, nil)
		var resObj controller.SedmaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		// player 0 is team 0, score 30
		assert.Equal(t, 30, resObj.Players[0].TeamScore)
		// player 1 is team 1, score 15
		assert.Equal(t, 15, resObj.Players[1].TeamScore)
	})
}

func TestSedmaWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.SedmaWebPresenter)

	t.Run("hint with card indices", func(t *testing.T) {
		m, _ := setupSedmaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.SedmaHint{CardIndices: []int{2}, Reason: "capture"})
		result := p.HintOutput(m)
		var resObj controller.SedmaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, []int{2}, resObj.Hint.CardIndices)
		assert.Equal(t, "capture", resObj.Hint.Reason)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupSedmaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.SedmaHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.SedmaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Nil(t, resObj.Hint)
	})
}

func TestSedmaWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.SedmaWebPresenter)
	m := new(interfaces.MockSedmaGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, `"actionType":"play"`)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestSedmaWebPresenterOutputCarriesTheHint(t *testing.T) {
	sdg, _ := setupSedmaWebMockWithPlayers()
	sdg.ExpectedCalls = removeMockCall(sdg.ExpectedCalls, "GetHint")
	sdg.On("GetHint").Return(&domain.SedmaHint{CardIndices: []int{1}, Reason: "capture"})

	result := new(presenter.SedmaWebPresenter).Output(sdg, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	// **Output は「頼んだヒント」の印を付けない。**付けると CLI が毎回 HINT 行を出す。
	assert.NotContains(t, result, "sedma.hintRequested")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**このゲーム群の
// hintAvailable は画面のラベルとして埋まっているので別キーを使う (#4483)。
func TestSedmaWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	sdg, _ := setupSedmaWebMockWithPlayers()
	sdg.ExpectedCalls = removeMockCall(sdg.ExpectedCalls, "GetHint")
	sdg.On("GetHint").Return(&domain.SedmaHint{CardIndices: []int{1}, Reason: "capture"})
	assert.Contains(t, new(presenter.SedmaWebPresenter).HintOutput(sdg), "sedma.hintRequested")

	none, _ := setupSedmaWebMockWithPlayers()
	none.ExpectedCalls = removeMockCall(none.ExpectedCalls, "GetHint")
	none.On("GetHint").Return((*domain.SedmaHint)(nil))
	assert.Contains(t, new(presenter.SedmaWebPresenter).HintOutput(none), "sedma.noHint")
}
