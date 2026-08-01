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

func setupTwentyNineWebMock() *interfaces.MockTwentyNineGame {
	m := new(interfaces.MockTwentyNineGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.TwentyNinePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetDeclarerIdx").Return(0)
	m.On("GetContract").Return(domain.TwentyNineBidTwenty)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetTrumpRevealed").Return(false)
	m.On("GetBids").Return([domain.TwentyNinePlayerCnt]domain.TwentyNineBid{domain.TwentyNineBidTwenty, domain.TwentyNineBidPass, domain.TwentyNineBidPass, domain.TwentyNineBidPass})
	m.On("GetTeamScores").Return([domain.TwentyNineTeamCnt]int{0, 0})
	m.On("GetRoundTeamPoints").Return([domain.TwentyNineTeamCnt]int{0, 0})
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetPlayableIndices", 0).Return([]int{0})
	m.On("IsHumanTurn").Return(true)
	m.On("IsHumanBidTurn").Return(false)
	m.On("GetConfig").Return(domain.DefaultTwentyNineConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	// **base だけに置く。**removeMockCall は最初の 1 件しか外さない。
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func setupTwentyNineWebMockWithPlayers() (*interfaces.MockTwentyNineGame, []*domain.TwentyNinePlayer) {
	m := setupTwentyNineWebMock()
	players := makeTwentyNinePlayers()
	m.On("GetPlayerCnt").Return(4)
	for i := 0; i < 4; i++ {
		m.On("GetPlayer", i).Return(players[i])
	}
	return m, players
}

func TestTwentyNineWebPresenter_Output(t *testing.T) {
	p := new(presenter.TwentyNineWebPresenter)

	t.Run("initial state play phase lead", func(t *testing.T) {
		m, players := setupTwentyNineWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))

		result := p.Output(m, nil)
		var resObj controller.TwentyNineWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.Players, 4)
		assert.Equal(t, int(domain.TwentyNinePhasePlay), resObj.Phase)
		assert.Equal(t, -1, resObj.WinnerTeam)
		assert.Equal(t, "twentynine.playPhase.lead", resObj.MessageCode)
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
		assert.True(t, resObj.Players[0].IsDeclarer)
		assert.False(t, resObj.Players[1].IsDeclarer)
		assert.Equal(t, []int{0}, resObj.PlayableIndices)
		assert.Equal(t, domain.CardDesignSpade, resObj.TrumpSuit)
		assert.False(t, resObj.TrumpRevealed)
		assert.Equal(t, int(domain.TwentyNineBidTwenty), resObj.Contract)
		assert.Equal(t, int(domain.TwentyNineBidTwenty), resObj.Bids[0])
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupTwentyNineWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.TwentyNineWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, int(domain.TwentyNineCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, 6, resObj.Config.TargetPoints)
	})

	t.Run("trump revealed propagated", func(t *testing.T) {
		m, _ := setupTwentyNineWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpRevealed")
		m.On("GetTrumpRevealed").Return(true)
		result := p.Output(m, nil)
		var resObj controller.TwentyNineWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.True(t, resObj.TrumpRevealed)
	})

	t.Run("bid phase message code", func(t *testing.T) {
		m, _ := setupTwentyNineWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsHumanBidTurn")
		m.On("GetPhase").Return(domain.TwentyNinePhaseBid)
		m.On("IsHumanBidTurn").Return(true)
		result := p.Output(m, nil)
		var resObj controller.TwentyNineWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "twentynine.bidPhase", resObj.MessageCode)
		assert.True(t, resObj.IsHumanBidTurn)
	})

	t.Run("play phase follow when trick has cards", func(t *testing.T) {
		m, _ := setupTwentyNineWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},
		})
		result := p.Output(m, nil)
		var resObj controller.TwentyNineWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.CurrentTrick, 1)
		assert.Equal(t, "twentynine.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end message code", func(t *testing.T) {
		m, _ := setupTwentyNineWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TwentyNinePhaseTrickEnd)
		result := p.Output(m, nil)
		var resObj controller.TwentyNineWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "twentynine.trickEnd", resObj.MessageCode)
	})

	t.Run("round end message code", func(t *testing.T) {
		m, _ := setupTwentyNineWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TwentyNinePhaseRoundEnd)
		result := p.Output(m, nil)
		var resObj controller.TwentyNineWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "twentynine.roundEnd", resObj.MessageCode)
	})

	t.Run("error message takes priority", func(t *testing.T) {
		m, _ := setupTwentyNineWebMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		var resObj controller.TwentyNineWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "boom", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human team wins", func(t *testing.T) {
		m, _ := setupTwentyNineWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)
		result := p.Output(m, nil)
		var resObj controller.TwentyNineWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "twentynine.result.humanWin", resObj.MessageCode)
		assert.Nil(t, resObj.MessageParams)
	})

	t.Run("game end cpu team wins", func(t *testing.T) {
		m, _ := setupTwentyNineWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(1)
		result := p.Output(m, nil)
		var resObj controller.TwentyNineWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "twentynine.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"team": "B"}, resObj.MessageParams)
	})

	t.Run("team scores propagated to players", func(t *testing.T) {
		m, _ := setupTwentyNineWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTeamScores")
		m.On("GetTeamScores").Return([domain.TwentyNineTeamCnt]int{3, 2})
		result := p.Output(m, nil)
		var resObj controller.TwentyNineWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		// player 0 -> team 0, player 1 -> team 1
		assert.Equal(t, 3, resObj.Players[0].TeamScore)
		assert.Equal(t, 2, resObj.Players[1].TeamScore)
		assert.Equal(t, [domain.TwentyNineTeamCnt]int{3, 2}, resObj.TeamScores)
	})
}

func TestTwentyNineWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.TwentyNineWebPresenter)

	t.Run("hint with card indices", func(t *testing.T) {
		m, _ := setupTwentyNineWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.TwentyNineHint{CardIndices: []int{2}, Reason: "follow_win"})
		result := p.HintOutput(m)
		var resObj controller.TwentyNineWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, []int{2}, resObj.Hint.CardIndices)
		assert.Equal(t, "follow_win", resObj.Hint.Reason)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupTwentyNineWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.TwentyNineHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.TwentyNineWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Nil(t, resObj.Hint)
	})
}

func TestTwentyNineWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.TwentyNineWebPresenter)
	m := new(interfaces.MockTwentyNineGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, `"actionType":"play"`)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
//
// Output 側にゲートは置きません。TwentyNine.GetHint() が「人間の手番で、かつ
// 行動を選べる状態か」を自分で確かめて nil を返します。
func TestTwentyNineWebPresenterOutputCarriesTheHint(t *testing.T) {
	tng, _ := setupTwentyNineWebMockWithPlayers()
	tng.ExpectedCalls = removeMockCall(tng.ExpectedCalls, "GetHint")
	tng.On("GetHint").Return(&domain.TwentyNineHint{CardIndices: []int{0}, Reason: "follow_suit"})

	result := new(presenter.TwentyNineWebPresenter).Output(tng, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
}
