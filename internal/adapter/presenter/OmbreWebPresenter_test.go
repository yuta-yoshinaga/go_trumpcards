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

func setupOmbreWebMock() *interfaces.MockOmbreGame {
	m := new(interfaces.MockOmbreGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.OmbrePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetCurrentBidderIdx").Return(1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetForehandIdx").Return(1)
	m.On("GetOmbreIdx").Return(0)
	m.On("GetWinningBid").Return(domain.OmbreBidEntrar)
	m.On("GetTrumpSuit").Return(domain.CardDesignHeart)
	m.On("GetOutcome").Return(domain.OmbreOutcomeNone)
	m.On("GetResult").Return(domain.OmbreResultNone)
	m.On("GetPlayerScores").Return([domain.OmbrePlayerCnt]int{0, 0, 0})
	m.On("GetWinnerPlayer").Return(-1)
	m.On("GetPlayableIndices", 0).Return([]int{0})
	m.On("IsHumanTurn").Return(true)
	m.On("IsHumanBidTurn").Return(false)
	m.On("GetConfig").Return(domain.DefaultOmbreConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	// **base だけに置く。**removeMockCall は最初の 1 件しか外さない。
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func setupOmbreWebMockWithPlayers() (*interfaces.MockOmbreGame, []*domain.OmbrePlayer) {
	m := setupOmbreWebMock()
	players := makeOmbrePlayers()
	m.On("GetPlayerCnt").Return(3)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	return m, players
}

func TestOmbreWebPresenter_Output(t *testing.T) {
	p := new(presenter.OmbreWebPresenter)

	t.Run("initial state play phase lead", func(t *testing.T) {
		m, players := setupOmbreWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))

		result := p.Output(m, nil)
		var resObj controller.OmbreWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.Players, 3)
		assert.Equal(t, int(domain.OmbrePhasePlay), resObj.Phase)
		assert.Equal(t, -1, resObj.WinnerPlayer)
		assert.Equal(t, -1, resObj.LastTrickWinner)
		assert.Equal(t, "ombre.playPhase.lead", resObj.MessageCode)
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
		assert.True(t, resObj.Players[0].IsOmbre)
		assert.False(t, resObj.Players[1].IsOmbre)
		assert.Equal(t, []int{0}, resObj.PlayableIndices)
		assert.Equal(t, 0, resObj.OmbreIdx)
		assert.Equal(t, domain.CardDesignHeart, resObj.TrumpSuit)
		assert.Equal(t, 1, resObj.ForehandIdx)
		assert.Equal(t, int(domain.OmbreBidEntrar), resObj.WinningBid)
		assert.True(t, resObj.IsHumanTurn)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupOmbreWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.OmbreWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, int(domain.OmbreCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, domain.OmbreWinRounds, resObj.Config.TargetRounds)
	})

	t.Run("bid phase message code", func(t *testing.T) {
		m, _ := setupOmbreWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.OmbrePhaseBid)
		result := p.Output(m, nil)
		var resObj controller.OmbreWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "ombre.bidPhase", resObj.MessageCode)
	})

	t.Run("play phase follow when trick has cards", func(t *testing.T) {
		m, _ := setupOmbreWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetPhase").Return(domain.OmbrePhasePlay)
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},
		})
		result := p.Output(m, nil)
		var resObj controller.OmbreWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.CurrentTrick, 1)
		assert.Equal(t, "ombre.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end message code", func(t *testing.T) {
		m, _ := setupOmbreWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.OmbrePhaseTrickEnd)
		result := p.Output(m, nil)
		var resObj controller.OmbreWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "ombre.trickEnd", resObj.MessageCode)
	})

	t.Run("round end message code by outcome", func(t *testing.T) {
		cases := []struct {
			outcome domain.OmbreOutcome
			code    string
		}{
			{domain.OmbreOutcomeSacar, "ombre.roundEnd.sacar"},
			{domain.OmbreOutcomePuesta, "ombre.roundEnd.puesta"},
			{domain.OmbreOutcomeCodille, "ombre.roundEnd.codille"},
			{domain.OmbreOutcomeNone, "ombre.roundEnd"},
		}
		for _, c := range cases {
			m, _ := setupOmbreWebMockWithPlayers()
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetOutcome")
			m.On("GetPhase").Return(domain.OmbrePhaseRoundEnd)
			m.On("GetOutcome").Return(c.outcome)
			result := p.Output(m, nil)
			var resObj controller.OmbreWebOutput
			assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
			assert.Equal(t, c.code, resObj.MessageCode)
			assert.Equal(t, int(c.outcome), resObj.Outcome)
		}
	})

	t.Run("error message takes priority", func(t *testing.T) {
		m, _ := setupOmbreWebMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		var resObj controller.OmbreWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "boom", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human wins", func(t *testing.T) {
		m, _ := setupOmbreWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(0)
		result := p.Output(m, nil)
		var resObj controller.OmbreWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "ombre.result.humanWin", resObj.MessageCode)
		assert.Nil(t, resObj.MessageParams)
	})

	t.Run("game end cpu wins", func(t *testing.T) {
		m, _ := setupOmbreWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(1)
		result := p.Output(m, nil)
		var resObj controller.OmbreWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "ombre.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"player": "1"}, resObj.MessageParams)
	})

	t.Run("player scores propagated to players", func(t *testing.T) {
		m, _ := setupOmbreWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayerScores")
		m.On("GetPlayerScores").Return([domain.OmbrePlayerCnt]int{7, 3, 0})
		result := p.Output(m, nil)
		var resObj controller.OmbreWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, 7, resObj.Players[0].Score)
		assert.Equal(t, 3, resObj.Players[1].Score)
		assert.Equal(t, 0, resObj.Players[2].Score)
	})
}

func TestOmbreWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.OmbreWebPresenter)

	t.Run("hint with card indices", func(t *testing.T) {
		m, _ := setupOmbreWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.OmbreHint{CardIndices: []int{2}, Reason: "follow_win"})
		result := p.HintOutput(m)
		var resObj controller.OmbreWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, []int{2}, resObj.Hint.CardIndices)
		assert.Equal(t, "follow_win", resObj.Hint.Reason)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupOmbreWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.OmbreHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.OmbreWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Nil(t, resObj.Hint)
	})
}

func TestOmbreWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.OmbreWebPresenter)
	m := new(interfaces.MockOmbreGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, `"actionType":"play"`)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestOmbreWebPresenterOutputCarriesTheHint(t *testing.T) {
	obg, _ := setupOmbreWebMockWithPlayers()
	obg.ExpectedCalls = removeMockCall(obg.ExpectedCalls, "GetHint")
	obg.On("GetHint").Return(&domain.OmbreHint{CardIndices: []int{0}, Reason: "follow_suit"})

	result := new(presenter.OmbreWebPresenter).Output(obg, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	// **Output は「頼んだヒント」の印を付けない。**付けると CLI が毎回 HINT 行を出す。
	assert.NotContains(t, result, "ombre.hintRequested")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestOmbreWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	obg, _ := setupOmbreWebMockWithPlayers()
	obg.ExpectedCalls = removeMockCall(obg.ExpectedCalls, "GetHint")
	obg.On("GetHint").Return(&domain.OmbreHint{CardIndices: []int{0}, Reason: "follow_suit"})
	assert.Contains(t, new(presenter.OmbreWebPresenter).HintOutput(obg), "ombre.hintRequested")

	none, _ := setupOmbreWebMockWithPlayers()
	none.ExpectedCalls = removeMockCall(none.ExpectedCalls, "GetHint")
	none.On("GetHint").Return((*domain.OmbreHint)(nil))
	assert.Contains(t, new(presenter.OmbreWebPresenter).HintOutput(none), "ombre.noHint")
}
