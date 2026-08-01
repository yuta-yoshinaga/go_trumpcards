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

func setupTysiacWebMock() *interfaces.MockTysiacGame {
	m := new(interfaces.MockTysiacGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.TysiacPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetForehandIdx").Return(1)
	m.On("GetDeclarerIdx").Return(0)
	m.On("GetContract").Return(100)
	m.On("GetCurrentBid").Return(100)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetPlayerScores").Return([domain.TysiacPlayerCnt]int{0, 0, 0})
	m.On("GetRoundCardPoints").Return([domain.TysiacPlayerCnt]int{0, 0, 0})
	m.On("GetRoundMarriage").Return([domain.TysiacPlayerCnt]int{0, 0, 0})
	m.On("GetWinnerPlayer").Return(-1)
	m.On("GetPlayableIndices", 0).Return([]int{0})
	m.On("IsHumanTurn").Return(true)
	m.On("GetConfig").Return(domain.DefaultTysiacConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	// **base だけに置く。**removeMockCall は最初の 1 件しか外さない。
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func setupTysiacWebMockWithPlayers() (*interfaces.MockTysiacGame, []*domain.TysiacPlayer) {
	m := setupTysiacWebMock()
	players := makeTysiacPlayers()
	m.On("GetPlayerCnt").Return(3)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	return m, players
}

func TestTysiacWebPresenter_Output(t *testing.T) {
	p := new(presenter.TysiacWebPresenter)

	t.Run("initial state play phase lead", func(t *testing.T) {
		m, players := setupTysiacWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))

		result := p.Output(m, nil)
		var resObj controller.TysiacWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.Players, 3)
		assert.Equal(t, int(domain.TysiacPhasePlay), resObj.Phase)
		assert.Equal(t, -1, resObj.WinnerPlayer)
		assert.Equal(t, -1, resObj.LastTrickWinner)
		assert.Equal(t, "tysiac.playPhase.lead", resObj.MessageCode)
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
		assert.True(t, resObj.Players[0].IsDeclarer)
		assert.False(t, resObj.Players[1].IsDeclarer)
		assert.Equal(t, []int{0}, resObj.PlayableIndices)
		assert.Equal(t, domain.CardDesignSpade, resObj.TrumpSuit)
		assert.Equal(t, 0, resObj.DeclarerIdx)
		assert.Equal(t, 1, resObj.ForehandIdx)
		assert.Equal(t, 100, resObj.Contract)
		assert.Equal(t, 100, resObj.CurrentBid)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupTysiacWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.TysiacWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, int(domain.TysiacCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, domain.TysiacWinTarget, resObj.Config.TargetPoints)
	})

	t.Run("bid phase message code", func(t *testing.T) {
		m, _ := setupTysiacWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TysiacPhaseBid)
		result := p.Output(m, nil)
		var resObj controller.TysiacWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "tysiac.bidPhase", resObj.MessageCode)
	})

	t.Run("talon phase message code", func(t *testing.T) {
		m, _ := setupTysiacWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TysiacPhaseTalon)
		result := p.Output(m, nil)
		var resObj controller.TysiacWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "tysiac.talonPhase", resObj.MessageCode)
	})

	t.Run("play phase follow when trick has cards", func(t *testing.T) {
		m, _ := setupTysiacWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetPhase").Return(domain.TysiacPhasePlay)
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},
		})
		result := p.Output(m, nil)
		var resObj controller.TysiacWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.CurrentTrick, 1)
		assert.Equal(t, "tysiac.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end message code", func(t *testing.T) {
		m, _ := setupTysiacWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TysiacPhaseTrickEnd)
		result := p.Output(m, nil)
		var resObj controller.TysiacWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "tysiac.trickEnd", resObj.MessageCode)
	})

	t.Run("round end message code", func(t *testing.T) {
		m, _ := setupTysiacWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TysiacPhaseRoundEnd)
		result := p.Output(m, nil)
		var resObj controller.TysiacWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "tysiac.roundEnd", resObj.MessageCode)
	})

	t.Run("error message takes priority", func(t *testing.T) {
		m, _ := setupTysiacWebMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		var resObj controller.TysiacWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "boom", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human wins", func(t *testing.T) {
		m, _ := setupTysiacWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(0)
		result := p.Output(m, nil)
		var resObj controller.TysiacWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "tysiac.result.humanWin", resObj.MessageCode)
		assert.Nil(t, resObj.MessageParams)
	})

	t.Run("game end cpu wins", func(t *testing.T) {
		m, _ := setupTysiacWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(1)
		result := p.Output(m, nil)
		var resObj controller.TysiacWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "tysiac.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"player": "1"}, resObj.MessageParams)
	})

	t.Run("player scores propagated to players", func(t *testing.T) {
		m, _ := setupTysiacWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayerScores")
		m.On("GetPlayerScores").Return([domain.TysiacPlayerCnt]int{40, 20, 0})
		result := p.Output(m, nil)
		var resObj controller.TysiacWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, 40, resObj.Players[0].Score)
		assert.Equal(t, 20, resObj.Players[1].Score)
		assert.Equal(t, 0, resObj.Players[2].Score)
	})
}

func TestTysiacWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.TysiacWebPresenter)

	t.Run("hint with card indices", func(t *testing.T) {
		m, _ := setupTysiacWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.TysiacHint{CardIndices: []int{2}, Reason: "follow_win"})
		result := p.HintOutput(m)
		var resObj controller.TysiacWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, []int{2}, resObj.Hint.CardIndices)
		assert.Equal(t, "follow_win", resObj.Hint.Reason)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupTysiacWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.TysiacHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.TysiacWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Nil(t, resObj.Hint)
	})
}

func TestTysiacWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.TysiacWebPresenter)
	m := new(interfaces.MockTysiacGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, `"actionType":"play"`)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestTysiacWebPresenterOutputCarriesTheHint(t *testing.T) {
	tsg, _ := setupTysiacWebMockWithPlayers()
	tsg.ExpectedCalls = removeMockCall(tsg.ExpectedCalls, "GetHint")
	tsg.On("GetHint").Return(&domain.TysiacHint{CardIndices: []int{0}, Reason: "follow_suit"})

	result := new(presenter.TysiacWebPresenter).Output(tsg, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	// **Output は「頼んだヒント」の印を付けない。**付けると CLI が毎回 HINT 行を出す。
	assert.NotContains(t, result, "tysiac.hintRequested")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestTysiacWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	tsg, _ := setupTysiacWebMockWithPlayers()
	tsg.ExpectedCalls = removeMockCall(tsg.ExpectedCalls, "GetHint")
	tsg.On("GetHint").Return(&domain.TysiacHint{CardIndices: []int{0}, Reason: "follow_suit"})
	assert.Contains(t, new(presenter.TysiacWebPresenter).HintOutput(tsg), "tysiac.hintRequested")

	none, _ := setupTysiacWebMockWithPlayers()
	none.ExpectedCalls = removeMockCall(none.ExpectedCalls, "GetHint")
	none.On("GetHint").Return((*domain.TysiacHint)(nil))
	assert.Contains(t, new(presenter.TysiacWebPresenter).HintOutput(none), "tysiac.noHint")
}
