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

// setupOmiWebMock creates a MockOmiGame with sensible defaults for Web tests.
func setupOmiWebMock() *interfaces.MockOmiGame {
	m := new(interfaces.MockOmiGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.OmiPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetTrumpCallerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTrumpSuit").Return(1)
	m.On("GetMakerTeam").Return(0)
	m.On("GetTeamScore", 0).Return(0)
	m.On("GetTeamScore", 1).Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultOmiConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func setupOmiWebMockWithPlayers() (*interfaces.MockOmiGame, []*domain.OmiPlayer) {
	m := setupOmiWebMock()
	players := makeOmiPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestOmiWebPresenter_Output(t *testing.T) {
	p := new(presenter.OmiWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupOmiWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		assert.NotEmpty(t, result)

		var resObj controller.OmiWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Equal(t, 4, len(resObj.Players))
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, 0, resObj.CurrentPlayerIdx)
		assert.Equal(t, int(domain.OmiPhasePlay), resObj.Phase)
		assert.Equal(t, 1, resObj.RoundNumber)
		assert.Equal(t, 1, resObj.TrickNumber)
		assert.Equal(t, -1, resObj.WinnerTeam)
		assert.Equal(t, 0, resObj.LeadPlayerIdx)
		assert.Equal(t, 0, resObj.TrumpCallerIdx)
		assert.Equal(t, 0, resObj.BidPlayerIdx)
		assert.Equal(t, 2, resObj.DealStage)
		assert.Equal(t, [2]int{0, 0}, resObj.TeamTricks)
		assert.Nil(t, resObj.FaceUpCard)
		assert.False(t, resObj.GoingAlone)
		assert.Equal(t, -1, resObj.GoingAlonePlayerIdx)
		assert.Equal(t, "", resObj.Message)
		assert.Empty(t, resObj.CurrentTrick)
	})

	t.Run("human cards shown, CPU cards hidden", func(t *testing.T) {
		m, players := setupOmiWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.OmiWebOutput
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
		m, players := setupOmiWebMockWithPlayers()
		players[1].AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 3, false)})

		result := p.Output(m, nil)
		var resObj controller.OmiWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, 1, resObj.Players[1].Team)
		assert.Equal(t, 1, resObj.Players[1].TrickCount)
		assert.Equal(t, [2]int{0, 1}, resObj.TeamTricks)
	})

	t.Run("current trick populated", func(t *testing.T) {
		m, _ := setupOmiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		trick := []*domain.TrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 3, false)},
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 7, false)},
		}
		m.On("GetCurrentTrick").Return(trick)

		result := p.Output(m, nil)
		var resObj controller.OmiWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Len(t, resObj.CurrentTrick, 2)
		assert.Equal(t, 0, resObj.CurrentTrick[0].PlayerIdx)
		assert.Equal(t, "CLOVER", resObj.CurrentTrick[0].Card.Design)
		assert.Equal(t, 3, resObj.CurrentTrick[0].Card.Value)
		assert.Equal(t, 1, resObj.CurrentTrick[1].PlayerIdx)
	})

	t.Run("empty current trick", func(t *testing.T) {
		m, _ := setupOmiWebMockWithPlayers()

		result := p.Output(m, nil)
		var resObj controller.OmiWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Empty(t, resObj.CurrentTrick)
	})

	t.Run("team scores", func(t *testing.T) {
		m, _ := setupOmiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTeamScore")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTeamScore")
		m.On("GetTeamScore", 0).Return(5)
		m.On("GetTeamScore", 1).Return(3)

		result := p.Output(m, nil)
		var resObj controller.OmiWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, [2]int{5, 3}, resObj.TeamScores)
	})

	t.Run("trump suit and dealer", func(t *testing.T) {
		m, _ := setupOmiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpSuit")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDealerIdx")
		m.On("GetTrumpSuit").Return(3)
		m.On("GetDealerIdx").Return(2)

		result := p.Output(m, nil)
		var resObj controller.OmiWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, 3, resObj.TrumpSuit)
		assert.Equal(t, 2, resObj.DealerIdx)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupOmiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetConfig")
		m.On("GetConfig").Return(domain.OmiConfig{
			CpuDifficulty: domain.OmiCpuDifficultyHard,
			PointLimit:    20,
		})

		result := p.Output(m, nil)
		var resObj controller.OmiWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, int(domain.OmiCpuDifficultyHard), resObj.Config.CpuDifficulty)
		assert.Equal(t, 20, resObj.Config.PointLimit)
	})

	t.Run("error message", func(t *testing.T) {
		m, _ := setupOmiWebMockWithPlayers()

		result := p.Output(m, errors.New("test error"))
		var resObj controller.OmiWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "test error", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end team 0 wins", func(t *testing.T) {
		m, _ := setupOmiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)

		result := p.Output(m, nil)
		var resObj controller.OmiWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.True(t, resObj.GameEndFlag)
		assert.Contains(t, resObj.Message, "チーム0")
		assert.Equal(t, "omi.result.team0Win", resObj.MessageCode)
		assert.Equal(t, map[string]string{"team": "0"}, resObj.MessageParams)
	})

	t.Run("game end team 1 wins", func(t *testing.T) {
		m, _ := setupOmiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(1)

		result := p.Output(m, nil)
		var resObj controller.OmiWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.True(t, resObj.GameEndFlag)
		assert.Contains(t, resObj.Message, "チーム1")
		assert.Equal(t, "omi.result.team1Win", resObj.MessageCode)
	})

	t.Run("call trump phase messageCode and dealStage", func(t *testing.T) {
		m, _ := setupOmiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.OmiPhaseCallTrump)

		result := p.Output(m, nil)
		var resObj controller.OmiWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "omi.callTrumpPhase", resObj.MessageCode)
		assert.Equal(t, 1, resObj.DealStage)
	})

	t.Run("play phase lead messageCode when trick empty", func(t *testing.T) {
		m, _ := setupOmiWebMockWithPlayers()

		result := p.Output(m, nil)
		var resObj controller.OmiWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "omi.playPhase.lead", resObj.MessageCode)
		assert.Equal(t, 2, resObj.DealStage)
	})

	t.Run("play phase follow messageCode when trick has cards", func(t *testing.T) {
		m, _ := setupOmiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		trick := []*domain.TrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 3, false)},
		}
		m.On("GetCurrentTrick").Return(trick)

		result := p.Output(m, nil)
		var resObj controller.OmiWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "omi.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end messageCode", func(t *testing.T) {
		m, _ := setupOmiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.OmiPhaseTrickEnd)

		result := p.Output(m, nil)
		var resObj controller.OmiWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "omi.trickEnd", resObj.MessageCode)
	})

	t.Run("round end messageCode", func(t *testing.T) {
		m, _ := setupOmiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.OmiPhaseRoundEnd)

		result := p.Output(m, nil)
		var resObj controller.OmiWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "omi.roundEnd", resObj.MessageCode)
	})

	t.Run("error takes priority over phase message", func(t *testing.T) {
		m, _ := setupOmiWebMockWithPlayers()

		result := p.Output(m, errors.New("some error"))
		var resObj controller.OmiWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "some error", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end phase no messageCode for unknown phases", func(t *testing.T) {
		m, _ := setupOmiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.OmiPhaseGameEnd)

		result := p.Output(m, nil)
		var resObj controller.OmiWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Empty(t, resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("default config values", func(t *testing.T) {
		m, _ := setupOmiWebMockWithPlayers()

		result := p.Output(m, nil)
		var resObj controller.OmiWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, int(domain.OmiCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, 10, resObj.Config.PointLimit)
	})
}

func TestOmiWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.OmiWebPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockOmiGame)
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
		m := new(interfaces.MockOmiGame)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"entries":[]`)
		m.AssertExpectations(t)
	})

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockOmiGame)
		m.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"entries":[]`)
		m.AssertExpectations(t)
	})
}

func TestOmiWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.OmiWebPresenter)

	t.Run("hint available with card", func(t *testing.T) {
		idx := 2
		m, _ := setupOmiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.OmiHint{
			CardIndex: &idx,
			Reason:    "follow_suit",
		})

		result := p.HintOutput(m)
		var resObj controller.OmiWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, &idx, resObj.Hint.CardIndex)
		assert.Equal(t, "follow_suit", resObj.Hint.Reason)
		assert.Equal(t, "omi.hintRequested", resObj.MessageCode)
	})

	t.Run("hint available with suit", func(t *testing.T) {
		suit := 3
		m, _ := setupOmiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.OmiHint{
			Suit:   &suit,
			Reason: "strategic_call",
		})

		result := p.HintOutput(m)
		var resObj controller.OmiWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, &suit, resObj.Hint.Suit)
		assert.Equal(t, "strategic_call", resObj.Hint.Reason)
		assert.Equal(t, "omi.hintRequested", resObj.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupOmiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.OmiHint)(nil))

		result := p.HintOutput(m)
		var resObj controller.OmiWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Nil(t, resObj.Hint)
		assert.Equal(t, "omi.noHint", resObj.MessageCode)
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestOmiWebPresenterOutputCarriesTheHint(t *testing.T) {
	idx := 0
	ecg, _ := setupOmiWebMockWithPlayers()
	ecg.ExpectedCalls = removeMockCall(ecg.ExpectedCalls, "GetHint")
	ecg.On("GetHint").Return(&domain.OmiHint{CardIndex: &idx, Reason: "follow_suit"})

	result := new(presenter.OmiWebPresenter).Output(ecg, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	assert.NotContains(t, result, "omi.hintRequested")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestOmiWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	idx := 0
	ecg, _ := setupOmiWebMockWithPlayers()
	ecg.ExpectedCalls = removeMockCall(ecg.ExpectedCalls, "GetHint")
	ecg.On("GetHint").Return(&domain.OmiHint{CardIndex: &idx, Reason: "follow_suit"})
	assert.Contains(t, new(presenter.OmiWebPresenter).HintOutput(ecg), "omi.hintRequested")

	none, _ := setupOmiWebMockWithPlayers()
	none.ExpectedCalls = removeMockCall(none.ExpectedCalls, "GetHint")
	none.On("GetHint").Return((*domain.OmiHint)(nil))
	assert.Contains(t, new(presenter.OmiWebPresenter).HintOutput(none), "omi.noHint")
}
