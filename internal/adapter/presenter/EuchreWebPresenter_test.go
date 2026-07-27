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

// setupEuchreWebMock creates a MockEuchreGame with sensible defaults for Web tests.
func setupEuchreWebMock() *interfaces.MockEuchreGame {
	m := new(interfaces.MockEuchreGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.EuchrePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTrumpSuit").Return(1)
	m.On("GetFaceUpCard").Return((*domain.Card)(nil))
	m.On("GetMakerTeam").Return(0)
	m.On("GetGoingAlone").Return(false)
	m.On("GetGoingAlonePlayerIdx").Return(-1)
	m.On("GetTeamScore", 0).Return(0)
	m.On("GetTeamScore", 1).Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultEuchreConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupEuchreWebMockWithPlayers() (*interfaces.MockEuchreGame, []*domain.EuchrePlayer) {
	m := setupEuchreWebMock()
	players := makeEuchrePlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestEuchreWebPresenter_Output(t *testing.T) {
	p := new(presenter.EuchreWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupEuchreWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		assert.NotEmpty(t, result)

		var resObj controller.EuchreWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Equal(t, 4, len(resObj.Players))
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, 0, resObj.CurrentPlayerIdx)
		assert.Equal(t, int(domain.EuchrePhasePlay), resObj.Phase)
		assert.Equal(t, 1, resObj.RoundNumber)
		assert.Equal(t, 1, resObj.TrickNumber)
		assert.Equal(t, -1, resObj.WinnerTeam)
		assert.Equal(t, 0, resObj.LeadPlayerIdx)
		assert.Equal(t, "", resObj.Message)
		assert.Empty(t, resObj.CurrentTrick)
	})

	t.Run("human cards shown, CPU cards hidden", func(t *testing.T) {
		m, players := setupEuchreWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.EuchreWebOutput
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

	t.Run("player team and tricks", func(t *testing.T) {
		m, players := setupEuchreWebMockWithPlayers()
		players[1].AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 3, false)})

		result := p.Output(m, nil)
		var resObj controller.EuchreWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, 1, resObj.Players[1].Team)
		assert.Equal(t, 1, resObj.Players[1].TrickCount)
	})

	t.Run("current trick populated", func(t *testing.T) {
		m, _ := setupEuchreWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		trick := []*domain.TrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 3, false)},
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 7, false)},
		}
		m.On("GetCurrentTrick").Return(trick)

		result := p.Output(m, nil)
		var resObj controller.EuchreWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Len(t, resObj.CurrentTrick, 2)
		assert.Equal(t, 0, resObj.CurrentTrick[0].PlayerIdx)
		assert.Equal(t, "CLOVER", resObj.CurrentTrick[0].Card.Design)
		assert.Equal(t, 3, resObj.CurrentTrick[0].Card.Value)
		assert.Equal(t, 1, resObj.CurrentTrick[1].PlayerIdx)
	})

	t.Run("empty current trick", func(t *testing.T) {
		m, _ := setupEuchreWebMockWithPlayers()

		result := p.Output(m, nil)
		var resObj controller.EuchreWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Empty(t, resObj.CurrentTrick)
	})

	t.Run("team scores", func(t *testing.T) {
		m, _ := setupEuchreWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTeamScore")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTeamScore")
		m.On("GetTeamScore", 0).Return(5)
		m.On("GetTeamScore", 1).Return(3)

		result := p.Output(m, nil)
		var resObj controller.EuchreWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, [2]int{5, 3}, resObj.TeamScores)
	})

	t.Run("trump suit and dealer", func(t *testing.T) {
		m, _ := setupEuchreWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpSuit")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDealerIdx")
		m.On("GetTrumpSuit").Return(3)
		m.On("GetDealerIdx").Return(2)

		result := p.Output(m, nil)
		var resObj controller.EuchreWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, 3, resObj.TrumpSuit)
		assert.Equal(t, 2, resObj.DealerIdx)
	})

	t.Run("face up card", func(t *testing.T) {
		m, _ := setupEuchreWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetFaceUpCard")
		m.On("GetFaceUpCard").Return(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := p.Output(m, nil)
		var resObj controller.EuchreWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.NotNil(t, resObj.FaceUpCard)
		assert.Equal(t, "HEART", resObj.FaceUpCard.Design)
		assert.Equal(t, 11, resObj.FaceUpCard.Value)
	})

	t.Run("going alone", func(t *testing.T) {
		m, _ := setupEuchreWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGoingAlone")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGoingAlonePlayerIdx")
		m.On("GetGoingAlone").Return(true)
		m.On("GetGoingAlonePlayerIdx").Return(2)

		result := p.Output(m, nil)
		var resObj controller.EuchreWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.True(t, resObj.GoingAlone)
		assert.Equal(t, 2, resObj.GoingAlonePlayerIdx)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupEuchreWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetConfig")
		m.On("GetConfig").Return(domain.EuchreConfig{
			CpuDifficulty: domain.EuchreCpuDifficultyHard,
			PointLimit:    20,
		})

		result := p.Output(m, nil)
		var resObj controller.EuchreWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, int(domain.EuchreCpuDifficultyHard), resObj.Config.CpuDifficulty)
		assert.Equal(t, 20, resObj.Config.PointLimit)
	})

	t.Run("error message", func(t *testing.T) {
		m, _ := setupEuchreWebMockWithPlayers()

		result := p.Output(m, errors.New("test error"))
		var resObj controller.EuchreWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "test error", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end team 0 wins", func(t *testing.T) {
		m, _ := setupEuchreWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)

		result := p.Output(m, nil)
		var resObj controller.EuchreWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.True(t, resObj.GameEndFlag)
		assert.Contains(t, resObj.Message, "チーム0")
		assert.Equal(t, "euchre.result.team0Win", resObj.MessageCode)
		assert.Equal(t, map[string]string{"team": "0"}, resObj.MessageParams)
	})

	t.Run("game end team 1 wins", func(t *testing.T) {
		m, _ := setupEuchreWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(1)

		result := p.Output(m, nil)
		var resObj controller.EuchreWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.True(t, resObj.GameEndFlag)
		assert.Contains(t, resObj.Message, "チーム1")
		assert.Equal(t, "euchre.result.team1Win", resObj.MessageCode)
	})

	t.Run("pickup phase messageCode", func(t *testing.T) {
		m, _ := setupEuchreWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.EuchrePhasePickUp)

		result := p.Output(m, nil)
		var resObj controller.EuchreWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "euchre.pickUpPhase", resObj.MessageCode)
	})

	t.Run("call trump phase messageCode", func(t *testing.T) {
		m, _ := setupEuchreWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.EuchrePhaseCallTrump)

		result := p.Output(m, nil)
		var resObj controller.EuchreWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "euchre.callTrumpPhase", resObj.MessageCode)
	})

	t.Run("discard phase messageCode", func(t *testing.T) {
		m, _ := setupEuchreWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.EuchrePhaseDiscard)

		result := p.Output(m, nil)
		var resObj controller.EuchreWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "euchre.discardPhase", resObj.MessageCode)
	})

	t.Run("play phase lead messageCode when trick empty", func(t *testing.T) {
		m, _ := setupEuchreWebMockWithPlayers()

		result := p.Output(m, nil)
		var resObj controller.EuchreWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "euchre.playPhase.lead", resObj.MessageCode)
	})

	t.Run("play phase follow messageCode when trick has cards", func(t *testing.T) {
		m, _ := setupEuchreWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		trick := []*domain.TrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 3, false)},
		}
		m.On("GetCurrentTrick").Return(trick)

		result := p.Output(m, nil)
		var resObj controller.EuchreWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "euchre.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end messageCode", func(t *testing.T) {
		m, _ := setupEuchreWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.EuchrePhaseTrickEnd)

		result := p.Output(m, nil)
		var resObj controller.EuchreWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "euchre.trickEnd", resObj.MessageCode)
	})

	t.Run("round end messageCode", func(t *testing.T) {
		m, _ := setupEuchreWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.EuchrePhaseRoundEnd)

		result := p.Output(m, nil)
		var resObj controller.EuchreWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "euchre.roundEnd", resObj.MessageCode)
	})

	t.Run("error takes priority over phase message", func(t *testing.T) {
		m, _ := setupEuchreWebMockWithPlayers()

		result := p.Output(m, errors.New("some error"))
		var resObj controller.EuchreWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "some error", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end phase no messageCode for unknown phases", func(t *testing.T) {
		m, _ := setupEuchreWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.EuchrePhaseGameEnd)
		// GetGameEndFlag remains false (default)

		result := p.Output(m, nil)
		var resObj controller.EuchreWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Empty(t, resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("default config values", func(t *testing.T) {
		m, _ := setupEuchreWebMockWithPlayers()

		result := p.Output(m, nil)
		var resObj controller.EuchreWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, int(domain.EuchreCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, 10, resObj.Config.PointLimit)
	})
}

func TestEuchreWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.EuchreWebPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockEuchreGame)
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
		m := new(interfaces.MockEuchreGame)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"entries":[]`)
		m.AssertExpectations(t)
	})

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockEuchreGame)
		m.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"entries":[]`)
		m.AssertExpectations(t)
	})
}

func TestEuchreWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.EuchreWebPresenter)

	t.Run("hint available with card", func(t *testing.T) {
		idx := 2
		m, _ := setupEuchreWebMockWithPlayers()
		m.On("GetHint").Return(&domain.EuchreHint{
			CardIndex: &idx,
			Reason:    "follow_suit",
		})

		result := p.HintOutput(m)
		var resObj controller.EuchreWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, &idx, resObj.Hint.CardIndex)
		assert.Equal(t, "follow_suit", resObj.Hint.Reason)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("hint available with orderUp", func(t *testing.T) {
		orderUp := true
		m, _ := setupEuchreWebMockWithPlayers()
		m.On("GetHint").Return(&domain.EuchreHint{
			OrderUp: &orderUp,
			Reason:  "strong_hand",
		})

		result := p.HintOutput(m)
		var resObj controller.EuchreWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, &orderUp, resObj.Hint.OrderUp)
		assert.Equal(t, "strong_hand", resObj.Hint.Reason)
	})

	t.Run("hint available with suit", func(t *testing.T) {
		suit := 3
		goAlone := true
		m, _ := setupEuchreWebMockWithPlayers()
		m.On("GetHint").Return(&domain.EuchreHint{
			Suit:    &suit,
			GoAlone: &goAlone,
			Reason:  "strong_hand",
		})

		result := p.HintOutput(m)
		var resObj controller.EuchreWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, &suit, resObj.Hint.Suit)
		assert.Equal(t, &goAlone, resObj.Hint.GoAlone)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupEuchreWebMockWithPlayers()
		m.On("GetHint").Return((*domain.EuchreHint)(nil))

		result := p.HintOutput(m)
		var resObj controller.EuchreWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Nil(t, resObj.Hint)
		assert.Empty(t, resObj.MessageCode)
	})
}
