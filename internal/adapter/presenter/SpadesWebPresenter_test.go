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

// setupSpadesWebMock creates a MockSpadesGame with sensible defaults for Web tests.
func setupSpadesWebMock() *interfaces.MockSpadesGame {
	m := new(interfaces.MockSpadesGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetSpadesBroken").Return(false)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.SpadesPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultSpadesConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupSpadesWebMockWithPlayers() (*interfaces.MockSpadesGame, []*domain.SpadesPlayer) {
	m := setupSpadesWebMock()
	players := makeSpadesPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func removeSpadesWebMockCall(calls []*mock.Call, method string) []*mock.Call {
	return removeMockCall(calls, method)
}

func TestSpadesWebPresenter_Output(t *testing.T) {
	p := new(presenter.SpadesWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupSpadesWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		assert.NotEmpty(t, result)

		var resObj controller.SpadesWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Equal(t, 4, len(resObj.Players))
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, 0, resObj.CurrentPlayerIdx)
		assert.Equal(t, 1, resObj.Phase) // SpadesPhasePlay
		assert.Equal(t, 1, resObj.RoundNumber)
		assert.Equal(t, 1, resObj.TrickNumber)
		assert.False(t, resObj.SpadesBroken)
		assert.Equal(t, -1, resObj.WinnerIdx)
		assert.Equal(t, 0, resObj.LeadPlayerIdx)
		assert.Equal(t, "", resObj.Message)
		assert.Empty(t, resObj.CurrentTrick)
	})

	t.Run("human cards shown, CPU cards hidden", func(t *testing.T) {
		m, players := setupSpadesWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.SpadesWebOutput
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

	t.Run("player bid, scores, tricks and bags", func(t *testing.T) {
		m, players := setupSpadesWebMockWithPlayers()
		players[1].SetCumulativeScore(200)
		players[1].SetRoundScore(50)
		players[1].SetBags(3)
		players[1].SetBid(4)
		players[1].AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 3, false)})

		result := p.Output(m, nil)
		var resObj controller.SpadesWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, 200, resObj.Players[1].CumulativeScore)
		assert.Equal(t, 50, resObj.Players[1].RoundScore)
		assert.Equal(t, 1, resObj.Players[1].TrickCount)
		assert.Equal(t, 3, resObj.Players[1].Bags)
		assert.Equal(t, 4, resObj.Players[1].Bid)
	})

	t.Run("current trick populated", func(t *testing.T) {
		m, _ := setupSpadesWebMockWithPlayers()
		m.ExpectedCalls = removeSpadesWebMockCall(m.ExpectedCalls, "GetCurrentTrick")
		trick := []*domain.TrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 3, false)},
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 7, false)},
		}
		m.On("GetCurrentTrick").Return(trick)

		result := p.Output(m, nil)
		var resObj controller.SpadesWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Len(t, resObj.CurrentTrick, 2)
		assert.Equal(t, 0, resObj.CurrentTrick[0].PlayerIdx)
		assert.Equal(t, "CLOVER", resObj.CurrentTrick[0].Card.Design)
		assert.Equal(t, 3, resObj.CurrentTrick[0].Card.Value)
		assert.Equal(t, 1, resObj.CurrentTrick[1].PlayerIdx)
	})

	t.Run("empty current trick", func(t *testing.T) {
		m, _ := setupSpadesWebMockWithPlayers()

		result := p.Output(m, nil)
		var resObj controller.SpadesWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Empty(t, resObj.CurrentTrick)
	})

	t.Run("spades broken true", func(t *testing.T) {
		m, _ := setupSpadesWebMockWithPlayers()
		m.ExpectedCalls = removeSpadesWebMockCall(m.ExpectedCalls, "GetSpadesBroken")
		m.On("GetSpadesBroken").Return(true)

		result := p.Output(m, nil)
		var resObj controller.SpadesWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.True(t, resObj.SpadesBroken)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupSpadesWebMockWithPlayers()
		m.ExpectedCalls = removeSpadesWebMockCall(m.ExpectedCalls, "GetConfig")
		m.On("GetConfig").Return(domain.SpadesConfig{
			CpuDifficulty:       domain.SpadesCpuDifficultyHard,
			PointLimit:          300,
			NilBonus:            200,
			BagPenaltyThreshold: 5,
		})

		result := p.Output(m, nil)
		var resObj controller.SpadesWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, int(domain.SpadesCpuDifficultyHard), resObj.Config.CpuDifficulty)
		assert.Equal(t, 300, resObj.Config.PointLimit)
		assert.Equal(t, 200, resObj.Config.NilBonus)
		assert.Equal(t, 5, resObj.Config.BagPenaltyThreshold)
	})

	t.Run("error message", func(t *testing.T) {
		m, _ := setupSpadesWebMockWithPlayers()

		result := p.Output(m, errors.New("test error"))
		var resObj controller.SpadesWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "test error", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human wins", func(t *testing.T) {
		m, _ := setupSpadesWebMockWithPlayers()
		m.ExpectedCalls = removeSpadesWebMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeSpadesWebMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)

		result := p.Output(m, nil)
		var resObj controller.SpadesWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.True(t, resObj.GameEndFlag)
		assert.Contains(t, resObj.Message, "あなた")
		assert.Equal(t, "spades.result.humanWin", resObj.MessageCode)
		assert.Nil(t, resObj.MessageParams)
	})

	t.Run("game end CPU wins", func(t *testing.T) {
		m, _ := setupSpadesWebMockWithPlayers()
		m.ExpectedCalls = removeSpadesWebMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeSpadesWebMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(2)

		result := p.Output(m, nil)
		var resObj controller.SpadesWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.True(t, resObj.GameEndFlag)
		assert.Contains(t, resObj.Message, "CPU 2")
		assert.Equal(t, "spades.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"cpuId": "2"}, resObj.MessageParams)
	})

	t.Run("game end nil player at winnerIdx", func(t *testing.T) {
		m := setupSpadesWebMock()
		m.On("GetPlayerCnt").Return(0)
		m.ExpectedCalls = removeSpadesWebMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeSpadesWebMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(99)
		m.On("GetPlayer", 99).Return((*domain.SpadesPlayer)(nil))

		result := p.Output(m, nil)
		var resObj controller.SpadesWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)

		assert.True(t, resObj.GameEndFlag)
		assert.Contains(t, resObj.Message, "CPU 99")
		assert.Equal(t, "spades.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"cpuId": "99"}, resObj.MessageParams)
	})

	t.Run("bid phase messageCode", func(t *testing.T) {
		m, _ := setupSpadesWebMockWithPlayers()
		m.ExpectedCalls = removeSpadesWebMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SpadesPhaseBid)

		result := p.Output(m, nil)
		var resObj controller.SpadesWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "spades.bidPhase", resObj.MessageCode)
	})

	t.Run("play phase lead messageCode when trick empty", func(t *testing.T) {
		m, _ := setupSpadesWebMockWithPlayers()
		// Default: phase=Play, currentTrick=nil (empty)

		result := p.Output(m, nil)
		var resObj controller.SpadesWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "spades.playPhase.lead", resObj.MessageCode)
	})

	t.Run("play phase follow messageCode when trick has cards", func(t *testing.T) {
		m, _ := setupSpadesWebMockWithPlayers()
		m.ExpectedCalls = removeSpadesWebMockCall(m.ExpectedCalls, "GetCurrentTrick")
		trick := []*domain.TrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 3, false)},
		}
		m.On("GetCurrentTrick").Return(trick)

		result := p.Output(m, nil)
		var resObj controller.SpadesWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "spades.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end messageCode", func(t *testing.T) {
		m, _ := setupSpadesWebMockWithPlayers()
		m.ExpectedCalls = removeSpadesWebMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SpadesPhaseTrickEnd)

		result := p.Output(m, nil)
		var resObj controller.SpadesWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "spades.trickEnd", resObj.MessageCode)
	})

	t.Run("round end messageCode", func(t *testing.T) {
		m, _ := setupSpadesWebMockWithPlayers()
		m.ExpectedCalls = removeSpadesWebMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SpadesPhaseRoundEnd)

		result := p.Output(m, nil)
		var resObj controller.SpadesWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "spades.roundEnd", resObj.MessageCode)
	})

	t.Run("error takes priority over phase message", func(t *testing.T) {
		m, _ := setupSpadesWebMockWithPlayers()

		result := p.Output(m, errors.New("some error"))
		var resObj controller.SpadesWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "some error", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end phase no messageCode for phases", func(t *testing.T) {
		m, _ := setupSpadesWebMockWithPlayers()
		m.ExpectedCalls = removeSpadesWebMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeSpadesWebMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeSpadesWebMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetPhase").Return(domain.SpadesPhaseGameEnd)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)

		result := p.Output(m, nil)
		var resObj controller.SpadesWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "spades.result.humanWin", resObj.MessageCode)
	})

	t.Run("unrecognized phase no messageCode", func(t *testing.T) {
		m, _ := setupSpadesWebMockWithPlayers()
		m.ExpectedCalls = removeSpadesWebMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SpadesPhaseGameEnd)
		// GetGameEndFlag remains false (default)

		result := p.Output(m, nil)
		var resObj controller.SpadesWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Empty(t, resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("default config values", func(t *testing.T) {
		m, _ := setupSpadesWebMockWithPlayers()

		result := p.Output(m, nil)
		var resObj controller.SpadesWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, int(domain.SpadesCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, 500, resObj.Config.PointLimit)
		assert.Equal(t, 100, resObj.Config.NilBonus)
		assert.Equal(t, 10, resObj.Config.BagPenaltyThreshold)
	})
}

func TestSpadesWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.SpadesWebPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockSpadesGame)
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
		m := new(interfaces.MockSpadesGame)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"entries":[]`)
		m.AssertExpectations(t)
	})

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockSpadesGame)
		m.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"entries":[]`)
		m.AssertExpectations(t)
	})
}

func TestSpadesWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.SpadesWebPresenter)

	t.Run("hint available with card", func(t *testing.T) {
		idx := 2
		m, _ := setupSpadesWebMockWithPlayers()
		m.On("GetHint").Return(&domain.SpadesHint{
			CardIndex: &idx,
			Reason:    "follow_suit",
		})

		result := p.HintOutput(m)
		var resObj controller.SpadesWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, &idx, resObj.Hint.CardIndex)
		assert.Equal(t, "follow_suit", resObj.Hint.Reason)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("hint available with bid", func(t *testing.T) {
		bid := 3
		m, _ := setupSpadesWebMockWithPlayers()
		m.On("GetHint").Return(&domain.SpadesHint{
			Bid:    &bid,
			Reason: "strategic_bid",
		})

		result := p.HintOutput(m)
		var resObj controller.SpadesWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, &bid, resObj.Hint.Bid)
		assert.Equal(t, "strategic_bid", resObj.Hint.Reason)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupSpadesWebMockWithPlayers()
		m.On("GetHint").Return((*domain.SpadesHint)(nil))

		result := p.HintOutput(m)
		var resObj controller.SpadesWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Nil(t, resObj.Hint)
		assert.Empty(t, resObj.MessageCode)
	})
}
