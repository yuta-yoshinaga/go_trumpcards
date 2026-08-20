//go:build test

package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func makeUltiPlayers() []*domain.UltiPlayer {
	return []*domain.UltiPlayer{
		domain.NewUltiPlayer(true),
		domain.NewUltiPlayer(false),
		domain.NewUltiPlayer(false),
	}
}

func setupUltiWebMock() *interfaces.MockUltiGame {
	m := new(interfaces.MockUltiGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.UltiPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetDeclarerIdx").Return(0)
	m.On("GetContract").Return(domain.UltiContractParty)
	m.On("GetTrumpSuit").Return(domain.CardDesignHeart)
	m.On("GetTalonCount").Return(0)
	m.On("GetTalonTaken").Return(true)
	m.On("GetDiscardCount").Return(2)
	m.On("GetOutcome").Return(domain.UltiOutcomeNone)
	m.On("GetResult").Return(domain.UltiResultNone)
	m.On("GetPlayerCoins").Return([domain.UltiPlayerCnt]int{0, 0, 0})
	m.On("GetLastDealCoins").Return([domain.UltiPlayerCnt]int{})
	m.On("GetCardPoints", mock.Anything).Return(0)
	m.On("GetWinnerPlayer").Return(-1)
	m.On("GetPlayableIndices", 0).Return([]int{0})
	m.On("IsHumanTurn").Return(true)
	m.On("IsHumanBidTurn").Return(false)
	m.On("GetConfig").Return(domain.DefaultUltiConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	// **base だけに置く。**removeMockCall は最初の 1 件しか外さない。
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func setupUltiWebMockWithPlayers() (*interfaces.MockUltiGame, []*domain.UltiPlayer) {
	m := setupUltiWebMock()
	players := makeUltiPlayers()
	m.On("GetPlayerCnt").Return(3)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	return m, players
}

func TestUltiWebPresenter_Output(t *testing.T) {
	p := new(presenter.UltiWebPresenter)

	t.Run("initial state play phase lead", func(t *testing.T) {
		m, players := setupUltiWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))

		result := p.Output(m, nil)
		var resObj controller.UltiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.Players, 3)
		assert.Equal(t, int(domain.UltiPhasePlay), resObj.Phase)
		assert.Equal(t, -1, resObj.WinnerPlayer)
		assert.Equal(t, "ulti.playPhase.lead", resObj.MessageCode)
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
		assert.True(t, resObj.Players[0].IsDeclarer)
		assert.False(t, resObj.Players[1].IsDeclarer)
		assert.Equal(t, []int{0}, resObj.PlayableIndices)
		assert.Equal(t, 0, resObj.DeclarerIdx)
		assert.Equal(t, domain.CardDesignHeart, resObj.TrumpSuit)
		assert.Equal(t, int(domain.UltiContractParty), resObj.Contract)
		assert.True(t, resObj.TalonTaken)
		assert.Equal(t, 2, resObj.DiscardCount)
		assert.True(t, resObj.IsHumanTurn)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupUltiWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.UltiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, int(domain.UltiCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, domain.UltiWinRounds, resObj.Config.TargetRounds)
	})

	t.Run("bid phase message code", func(t *testing.T) {
		m, _ := setupUltiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.UltiPhaseBid)
		result := p.Output(m, nil)
		var resObj controller.UltiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "ulti.bidPhase", resObj.MessageCode)
	})

	t.Run("discard phase message code", func(t *testing.T) {
		m, _ := setupUltiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.UltiPhaseDiscard)
		result := p.Output(m, nil)
		var resObj controller.UltiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "ulti.discardPhase", resObj.MessageCode)
	})

	t.Run("play phase follow when trick has cards", func(t *testing.T) {
		m, _ := setupUltiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},
		})
		result := p.Output(m, nil)
		var resObj controller.UltiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.CurrentTrick, 1)
		assert.Equal(t, "ulti.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end message code", func(t *testing.T) {
		m, _ := setupUltiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.UltiPhaseTrickEnd)
		result := p.Output(m, nil)
		var resObj controller.UltiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "ulti.trickEnd", resObj.MessageCode)
	})

	t.Run("round end message code by outcome", func(t *testing.T) {
		cases := []struct {
			outcome domain.UltiOutcome
			code    string
		}{
			{domain.UltiOutcomeWin, "ulti.roundEnd.win"},
			{domain.UltiOutcomeLoss, "ulti.roundEnd.loss"},
			{domain.UltiOutcomeNone, "ulti.roundEnd"},
		}
		for _, c := range cases {
			m, _ := setupUltiWebMockWithPlayers()
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetOutcome")
			m.On("GetPhase").Return(domain.UltiPhaseRoundEnd)
			m.On("GetOutcome").Return(c.outcome)
			result := p.Output(m, nil)
			var resObj controller.UltiWebOutput
			assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
			assert.Equal(t, c.code, resObj.MessageCode)
			assert.Equal(t, int(c.outcome), resObj.Outcome)
		}
	})

	t.Run("error message takes priority", func(t *testing.T) {
		m, _ := setupUltiWebMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		var resObj controller.UltiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "boom", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human wins", func(t *testing.T) {
		m, _ := setupUltiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(0)
		result := p.Output(m, nil)
		var resObj controller.UltiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "ulti.result.humanWin", resObj.MessageCode)
		assert.Nil(t, resObj.MessageParams)
	})

	t.Run("game end cpu wins", func(t *testing.T) {
		m, _ := setupUltiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(1)
		result := p.Output(m, nil)
		var resObj controller.UltiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "ulti.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"player": "1"}, resObj.MessageParams)
	})

	t.Run("coins propagated to players", func(t *testing.T) {
		m, _ := setupUltiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayerCoins")
		m.On("GetPlayerCoins").Return([domain.UltiPlayerCnt]int{7, 3, 0})
		result := p.Output(m, nil)
		var resObj controller.UltiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, 7, resObj.Players[0].Coins)
		assert.Equal(t, 3, resObj.Players[1].Coins)
		assert.Equal(t, 0, resObj.Players[2].Coins)
	})

	// #5690: 累積とは別に、今回のディールで動いた額そのものを返す。
	// Web はこれ以前、累積を ref に退避して差分を取っていた。
	t.Run("last deal settlement propagated", func(t *testing.T) {
		m, _ := setupUltiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayerCoins")
		m.On("GetPlayerCoins").Return([domain.UltiPlayerCnt]int{7, 3, 0})
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetLastDealCoins")
		m.On("GetLastDealCoins").Return([domain.UltiPlayerCnt]int{4, -2, -2})
		result := p.Output(m, nil)
		var resObj controller.UltiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, [domain.UltiPlayerCnt]int{4, -2, -2}, resObj.LastDealCoins)
	})
}

func TestUltiWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.UltiWebPresenter)

	t.Run("hint with card indices", func(t *testing.T) {
		m, _ := setupUltiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.UltiHint{CardIndices: []int{2}, Reason: "follow_win"})
		result := p.HintOutput(m)
		var resObj controller.UltiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, []int{2}, resObj.Hint.CardIndices)
		assert.Equal(t, "follow_win", resObj.Hint.Reason)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupUltiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.UltiHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.UltiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Nil(t, resObj.Hint)
	})
}

func TestUltiWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.UltiWebPresenter)
	m := new(interfaces.MockUltiGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, `"actionType":"play"`)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestUltiWebPresenterOutputCarriesTheHint(t *testing.T) {
	ulg, _ := setupUltiWebMockWithPlayers()
	ulg.ExpectedCalls = removeMockCall(ulg.ExpectedCalls, "GetHint")
	ulg.On("GetHint").Return(&domain.UltiHint{CardIndices: []int{0}, Reason: "follow_suit"})

	result := new(presenter.UltiWebPresenter).Output(ulg, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	// **Output は「頼んだヒント」の印を付けない。**付けると CLI が毎回 HINT 行を出す。
	assert.NotContains(t, result, "ulti.hintRequested")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestUltiWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	ulg, _ := setupUltiWebMockWithPlayers()
	ulg.ExpectedCalls = removeMockCall(ulg.ExpectedCalls, "GetHint")
	ulg.On("GetHint").Return(&domain.UltiHint{CardIndices: []int{0}, Reason: "follow_suit"})
	assert.Contains(t, new(presenter.UltiWebPresenter).HintOutput(ulg), "ulti.hintRequested")

	none, _ := setupUltiWebMockWithPlayers()
	none.ExpectedCalls = removeMockCall(none.ExpectedCalls, "GetHint")
	none.On("GetHint").Return((*domain.UltiHint)(nil))
	assert.Contains(t, new(presenter.UltiWebPresenter).HintOutput(none), "ulti.noHint")
}
