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

// setupHeartsWebMock creates a MockHeartsGame with sensible defaults for Web tests.
func setupHeartsWebMock() *interfaces.MockHeartsGame {
	m := new(interfaces.MockHeartsGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetHeartsBroken").Return(false)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.HeartsPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetPassDirection").Return(domain.HeartsPassLeft)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultHeartsConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupHeartsWebMockWithPlayers() (*interfaces.MockHeartsGame, []*domain.HeartsPlayer) {
	m := setupHeartsWebMock()
	players := makeHeartsPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func removeWebMockCall(calls []*mock.Call, method string) []*mock.Call {
	return removeMockCall(calls, method)
}

func TestHeartsWebPresenter_Output(t *testing.T) {
	p := new(presenter.HeartsWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupHeartsWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		assert.NotEmpty(t, result)

		var resObj controller.HeartsWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Equal(t, 4, len(resObj.Players))
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, 0, resObj.CurrentPlayerIdx)
		assert.Equal(t, 1, resObj.Phase) // HeartsPhasePlay
		assert.Equal(t, 1, resObj.RoundNumber)
		assert.Equal(t, 1, resObj.TrickNumber)
		assert.False(t, resObj.HeartsBroken)
		assert.Equal(t, 0, resObj.PassDirection) // HeartsPassLeft
		assert.Equal(t, -1, resObj.WinnerIdx)
		assert.Equal(t, 0, resObj.LeadPlayerIdx)
		assert.Equal(t, "", resObj.Message)
		assert.Empty(t, resObj.CurrentTrick)
	})

	t.Run("human cards shown, CPU cards hidden", func(t *testing.T) {
		m, players := setupHeartsWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.HeartsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		humanPlayer := resObj.Players[0]
		assert.True(t, humanPlayer.IsHuman)
		assert.Equal(t, 1, humanPlayer.CardCount)
		assert.Len(t, humanPlayer.Cards, 1)
		assert.Equal(t, "SPADE", humanPlayer.Cards[0].Design)
		assert.Equal(t, 5, humanPlayer.Cards[0].Value)

		cpu1 := resObj.Players[1]
		assert.False(t, cpu1.IsHuman)
		assert.Equal(t, 1, cpu1.CardCount)
		assert.Len(t, cpu1.Cards, 0)
	})

	t.Run("player scores and trick count", func(t *testing.T) {
		m, players := setupHeartsWebMockWithPlayers()
		players[1].SetCumulativeScore(20)
		players[1].SetRoundScore(5)
		players[1].AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 3, false)})

		result := p.Output(m, nil)
		var resObj controller.HeartsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, 20, resObj.Players[1].CumulativeScore)
		assert.Equal(t, 5, resObj.Players[1].RoundScore)
		assert.Equal(t, 1, resObj.Players[1].TrickCount)
	})

	t.Run("penalty cards filtered and sorted", func(t *testing.T) {
		m, players := setupHeartsWebMockWithPlayers()
		// Player 1 captured two tricks containing hearts, the Q♠, plus
		// non-penalty cards (a low spade, a club) and an omnibus J♦ which
		// must NOT be counted as a penalty card.
		players[1].AddTrick([]*domain.Card{
			domain.NewCard(domain.CardDesignHeart, 10, false),
			domain.NewCard(domain.CardDesignClover, 4, false),
			domain.NewCard(domain.CardDesignSpade, 12, false), // Q♠
			domain.NewCard(domain.CardDesignSpade, 3, false),  // low spade (not penalty)
		})
		players[1].AddTrick([]*domain.Card{
			domain.NewCard(domain.CardDesignHeart, 2, false),
			domain.NewCard(domain.CardDesignDiamond, 11, false), // J♦ (omnibus bonus, not penalty)
		})

		result := p.Output(m, nil)
		var resObj controller.HeartsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		// Player 0 has no tricks: empty (never nil) slice.
		assert.NotNil(t, resObj.Players[0].PenaltyCards)
		assert.Len(t, resObj.Players[0].PenaltyCards, 0)

		// Player 1: two hearts (2, 10) sorted ascending, then Q♠.
		pen := resObj.Players[1].PenaltyCards
		assert.Len(t, pen, 3)
		assert.Equal(t, "HEART", pen[0].Design)
		assert.Equal(t, 2, pen[0].Value)
		assert.Equal(t, "HEART", pen[1].Design)
		assert.Equal(t, 10, pen[1].Value)
		assert.Equal(t, "SPADE", pen[2].Design)
		assert.Equal(t, 12, pen[2].Value)
	})

	t.Run("current trick populated", func(t *testing.T) {
		m, _ := setupHeartsWebMockWithPlayers()
		m.ExpectedCalls = removeWebMockCall(m.ExpectedCalls, "GetCurrentTrick")
		trick := []*domain.TrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 3, false)},
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 7, false)},
		}
		m.On("GetCurrentTrick").Return(trick)

		result := p.Output(m, nil)
		var resObj controller.HeartsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Len(t, resObj.CurrentTrick, 2)
		assert.Equal(t, 0, resObj.CurrentTrick[0].PlayerIdx)
		assert.Equal(t, "CLOVER", resObj.CurrentTrick[0].Card.Design)
		assert.Equal(t, 3, resObj.CurrentTrick[0].Card.Value)
		assert.Equal(t, 1, resObj.CurrentTrick[1].PlayerIdx)
	})

	t.Run("empty current trick", func(t *testing.T) {
		m, _ := setupHeartsWebMockWithPlayers()

		result := p.Output(m, nil)
		var resObj controller.HeartsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Empty(t, resObj.CurrentTrick)
	})

	t.Run("hearts broken true", func(t *testing.T) {
		m, _ := setupHeartsWebMockWithPlayers()
		m.ExpectedCalls = removeWebMockCall(m.ExpectedCalls, "GetHeartsBroken")
		m.On("GetHeartsBroken").Return(true)

		result := p.Output(m, nil)
		var resObj controller.HeartsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.True(t, resObj.HeartsBroken)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupHeartsWebMockWithPlayers()
		m.ExpectedCalls = removeWebMockCall(m.ExpectedCalls, "GetConfig")
		m.On("GetConfig").Return(domain.HeartsConfig{
			CpuDifficulty: domain.HeartsCpuDifficultyHard,
			PointLimit:    50,
		})

		result := p.Output(m, nil)
		var resObj controller.HeartsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, int(domain.HeartsCpuDifficultyHard), resObj.Config.CpuDifficulty)
		assert.Equal(t, 50, resObj.Config.PointLimit)
	})

	t.Run("error message", func(t *testing.T) {
		m, _ := setupHeartsWebMockWithPlayers()

		result := p.Output(m, errors.New("test error"))
		var resObj controller.HeartsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "test error", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human wins", func(t *testing.T) {
		m, _ := setupHeartsWebMockWithPlayers()
		m.ExpectedCalls = removeWebMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeWebMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)

		result := p.Output(m, nil)
		var resObj controller.HeartsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.True(t, resObj.GameEndFlag)
		assert.Contains(t, resObj.Message, "あなた")
		assert.Equal(t, "hearts.result.humanWin", resObj.MessageCode)
		assert.Nil(t, resObj.MessageParams)
	})

	t.Run("game end CPU wins", func(t *testing.T) {
		m, _ := setupHeartsWebMockWithPlayers()
		m.ExpectedCalls = removeWebMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeWebMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(2)

		result := p.Output(m, nil)
		var resObj controller.HeartsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.True(t, resObj.GameEndFlag)
		assert.Contains(t, resObj.Message, "CPU 2")
		assert.Equal(t, "hearts.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"cpuId": "2"}, resObj.MessageParams)
	})

	t.Run("game end nil player at winnerIdx", func(t *testing.T) {
		m := setupHeartsWebMock()
		m.On("GetPlayerCnt").Return(0)
		m.ExpectedCalls = removeWebMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeWebMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(99)
		m.On("GetPlayer", 99).Return((*domain.HeartsPlayer)(nil))

		result := p.Output(m, nil)
		var resObj controller.HeartsWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)

		assert.True(t, resObj.GameEndFlag)
		assert.Contains(t, resObj.Message, "CPU 99")
		assert.Equal(t, "hearts.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"cpuId": "99"}, resObj.MessageParams)
	})

	t.Run("pass phase messageCode", func(t *testing.T) {
		m, _ := setupHeartsWebMockWithPlayers()
		m.ExpectedCalls = removeWebMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.HeartsPhasePass)

		result := p.Output(m, nil)
		var resObj controller.HeartsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "hearts.passPhase", resObj.MessageCode)
	})

	t.Run("play phase lead messageCode when trick empty", func(t *testing.T) {
		m, _ := setupHeartsWebMockWithPlayers()
		// Default: phase=Play, currentTrick=nil (empty)

		result := p.Output(m, nil)
		var resObj controller.HeartsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "hearts.playPhase.lead", resObj.MessageCode)
	})

	t.Run("play phase follow messageCode when trick has cards", func(t *testing.T) {
		m, _ := setupHeartsWebMockWithPlayers()
		m.ExpectedCalls = removeWebMockCall(m.ExpectedCalls, "GetCurrentTrick")
		trick := []*domain.TrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 3, false)},
		}
		m.On("GetCurrentTrick").Return(trick)

		result := p.Output(m, nil)
		var resObj controller.HeartsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "hearts.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end messageCode", func(t *testing.T) {
		m, _ := setupHeartsWebMockWithPlayers()
		m.ExpectedCalls = removeWebMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.HeartsPhaseTrickEnd)

		result := p.Output(m, nil)
		var resObj controller.HeartsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "hearts.trickEnd", resObj.MessageCode)
	})

	t.Run("round end messageCode", func(t *testing.T) {
		m, _ := setupHeartsWebMockWithPlayers()
		m.ExpectedCalls = removeWebMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.HeartsPhaseRoundEnd)

		result := p.Output(m, nil)
		var resObj controller.HeartsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "hearts.roundEnd", resObj.MessageCode)
	})

	t.Run("error takes priority over phase message", func(t *testing.T) {
		m, _ := setupHeartsWebMockWithPlayers()

		result := p.Output(m, errors.New("some error"))
		var resObj controller.HeartsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "some error", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end phase no messageCode for phases", func(t *testing.T) {
		m, _ := setupHeartsWebMockWithPlayers()
		m.ExpectedCalls = removeWebMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeWebMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeWebMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetPhase").Return(domain.HeartsPhaseGameEnd)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)

		result := p.Output(m, nil)
		var resObj controller.HeartsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "hearts.result.humanWin", resObj.MessageCode)
	})

	t.Run("unrecognized phase no messageCode", func(t *testing.T) {
		m, _ := setupHeartsWebMockWithPlayers()
		m.ExpectedCalls = removeWebMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.HeartsPhaseGameEnd)
		// GetGameEndFlag remains false (default)

		result := p.Output(m, nil)
		var resObj controller.HeartsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Empty(t, resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})
	t.Run("default config values", func(t *testing.T) {
		m, _ := setupHeartsWebMockWithPlayers()

		result := p.Output(m, nil)
		var resObj controller.HeartsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, int(domain.HeartsCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, 100, resObj.Config.PointLimit)
		assert.False(t, resObj.Config.OmnibusJD)
	})

	t.Run("omnibus config enabled", func(t *testing.T) {
		m, _ := setupHeartsWebMockWithPlayers()
		m.ExpectedCalls = removeWebMockCall(m.ExpectedCalls, "GetConfig")
		m.On("GetConfig").Return(domain.HeartsConfig{
			CpuDifficulty: domain.HeartsCpuDifficultyNormal,
			PointLimit:    100,
			OmnibusJD:     true,
		})

		result := p.Output(m, nil)
		var resObj controller.HeartsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.True(t, resObj.Config.OmnibusJD)
	})
}

func TestHeartsWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.HeartsWebPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockHeartsGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "played SPADE 5", Cards: []*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, true)}},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"actionType":"play"`)
		assert.Contains(t, result, `"detail":"played SPADE 5"`)
		assert.Contains(t, result, `"turnNumber":1`)
		assert.Contains(t, result, `"playerIdx":0`)
		m.AssertExpectations(t)
	})

	t.Run("nil entries", func(t *testing.T) {
		m := new(interfaces.MockHeartsGame)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"entries":[]`)
		m.AssertExpectations(t)
	})

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockHeartsGame)
		m.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"entries":[]`)
		m.AssertExpectations(t)
	})
}

func TestHeartsWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.HeartsWebPresenter)

	t.Run("hint available", func(t *testing.T) {
		m, _ := setupHeartsWebMockWithPlayers()
		m.On("GetHint").Return(&domain.HeartsHint{
			CardIndices: []int{2},
			Reason:      "follow_suit",
		})

		result := p.HintOutput(m)
		var resObj controller.HeartsWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, []int{2}, resObj.Hint.CardIndices)
		assert.Equal(t, "follow_suit", resObj.Hint.Reason)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupHeartsWebMockWithPlayers()
		m.On("GetHint").Return((*domain.HeartsHint)(nil))

		result := p.HintOutput(m)
		var resObj controller.HeartsWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Nil(t, resObj.Hint)
		assert.Empty(t, resObj.MessageCode)
	})
}
