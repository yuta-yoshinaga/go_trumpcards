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

// setupBridgeWebMock creates a MockBridgeGame with sensible defaults for Web tests.
func setupBridgeWebMock() *interfaces.MockBridgeGame {
	m := new(interfaces.MockBridgeGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.BridgePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTrumpSuit").Return(1)
	m.On("GetContractLevel").Return(1)
	m.On("GetContractSuit").Return(3)
	m.On("GetDoubled").Return(0)
	m.On("GetDeclarerIdx").Return(0)
	m.On("GetDummyIdx").Return(2)
	m.On("GetBidHistory").Return([]*domain.BridgeBidEntry(nil))
	m.On("GetVulnerability", 0).Return(false)
	m.On("GetVulnerability", 1).Return(false)
	m.On("GetTeamScore", 0).Return(0)
	m.On("GetTeamScore", 1).Return(0)
	m.On("GetGamesWon", 0).Return(0)
	m.On("GetGamesWon", 1).Return(0)
	m.On("GetBelowLine", 0).Return(0)
	m.On("GetBelowLine", 1).Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("IsOpeningLeadDone").Return(false)
	m.On("GetDummyHand").Return(([]*domain.Card)(nil))
	m.On("GetConfig").Return(domain.DefaultBridgeConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	// **base だけに置く。**removeMockCall は最初の 1 件しか外さない。
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func makeBridgePlayers() []*domain.BridgePlayer {
	return []*domain.BridgePlayer{
		domain.NewBridgePlayer(true, 0),
		domain.NewBridgePlayer(false, 1),
		domain.NewBridgePlayer(false, 0),
		domain.NewBridgePlayer(false, 1),
	}
}

func setupBridgeWebMockWithPlayers() (*interfaces.MockBridgeGame, []*domain.BridgePlayer) {
	m := setupBridgeWebMock()
	players := makeBridgePlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestBridgeWebPresenter_Output(t *testing.T) {
	p := new(presenter.BridgeWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupBridgeWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		assert.NotEmpty(t, result)

		var resObj controller.BridgeWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Equal(t, 4, len(resObj.Players))
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, 0, resObj.CurrentPlayerIdx)
		assert.Equal(t, int(domain.BridgePhasePlay), resObj.Phase)
		assert.Equal(t, 1, resObj.RoundNumber)
		assert.Equal(t, 1, resObj.TrickNumber)
		assert.Equal(t, -1, resObj.WinnerTeam)
		assert.Equal(t, 0, resObj.LeadPlayerIdx)
		assert.Equal(t, "", resObj.Message)
		assert.Empty(t, resObj.CurrentTrick)
	})

	t.Run("human cards shown, CPU cards hidden", func(t *testing.T) {
		m, players := setupBridgeWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.BridgeWebOutput
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
		m, players := setupBridgeWebMockWithPlayers()
		players[1].AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 3, false)})

		result := p.Output(m, nil)
		var resObj controller.BridgeWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, 1, resObj.Players[1].Team)
		assert.Equal(t, 1, resObj.Players[1].TrickCount)
	})

	t.Run("current trick populated", func(t *testing.T) {
		m, _ := setupBridgeWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		trick := []*domain.TrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 3, false)},
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 7, false)},
		}
		m.On("GetCurrentTrick").Return(trick)

		result := p.Output(m, nil)
		var resObj controller.BridgeWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Len(t, resObj.CurrentTrick, 2)
		assert.Equal(t, 0, resObj.CurrentTrick[0].PlayerIdx)
		assert.Equal(t, "CLOVER", resObj.CurrentTrick[0].Card.Design)
		assert.Equal(t, 3, resObj.CurrentTrick[0].Card.Value)
		assert.Equal(t, 1, resObj.CurrentTrick[1].PlayerIdx)
	})

	t.Run("empty current trick", func(t *testing.T) {
		m, _ := setupBridgeWebMockWithPlayers()

		result := p.Output(m, nil)
		var resObj controller.BridgeWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Empty(t, resObj.CurrentTrick)
	})

	t.Run("team scores", func(t *testing.T) {
		m, _ := setupBridgeWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTeamScore")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTeamScore")
		m.On("GetTeamScore", 0).Return(5)
		m.On("GetTeamScore", 1).Return(3)

		result := p.Output(m, nil)
		var resObj controller.BridgeWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, [2]int{5, 3}, resObj.TeamScores)
	})

	t.Run("trump suit and dealer", func(t *testing.T) {
		m, _ := setupBridgeWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpSuit")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDealerIdx")
		m.On("GetTrumpSuit").Return(3)
		m.On("GetDealerIdx").Return(2)

		result := p.Output(m, nil)
		var resObj controller.BridgeWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, 3, resObj.TrumpSuit)
		assert.Equal(t, 2, resObj.DealerIdx)
	})

	t.Run("contract info", func(t *testing.T) {
		m, _ := setupBridgeWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetContractLevel")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetContractSuit")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDoubled")
		m.On("GetContractLevel").Return(3)
		m.On("GetContractSuit").Return(5)
		m.On("GetDoubled").Return(1)

		result := p.Output(m, nil)
		var resObj controller.BridgeWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, 3, resObj.ContractLevel)
		assert.Equal(t, 5, resObj.ContractSuit)
		assert.Equal(t, 1, resObj.Doubled)
	})

	t.Run("declarer and dummy", func(t *testing.T) {
		m, _ := setupBridgeWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDeclarerIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDummyIdx")
		m.On("GetDeclarerIdx").Return(1)
		m.On("GetDummyIdx").Return(3)

		result := p.Output(m, nil)
		var resObj controller.BridgeWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, 1, resObj.DeclarerIdx)
		assert.Equal(t, 3, resObj.DummyIdx)
	})

	t.Run("bid history", func(t *testing.T) {
		m, _ := setupBridgeWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetBidHistory")
		history := []*domain.BridgeBidEntry{
			{PlayerIdx: 0, BidType: domain.BridgeBidNormal, Level: 1, Suit: 3},
			{PlayerIdx: 1, BidType: domain.BridgeBidPass, Level: 0, Suit: 0},
		}
		m.On("GetBidHistory").Return(history)

		result := p.Output(m, nil)
		var resObj controller.BridgeWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Len(t, resObj.BidHistory, 2)
		assert.Equal(t, 0, resObj.BidHistory[0].PlayerIdx)
		assert.Equal(t, 1, resObj.BidHistory[0].BidType)
		assert.Equal(t, 1, resObj.BidHistory[0].Level)
		assert.Equal(t, 3, resObj.BidHistory[0].Suit)
		assert.Equal(t, 0, resObj.BidHistory[1].BidType)
	})

	t.Run("vulnerability", func(t *testing.T) {
		m, _ := setupBridgeWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetVulnerability")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetVulnerability")
		m.On("GetVulnerability", 0).Return(true)
		m.On("GetVulnerability", 1).Return(false)

		result := p.Output(m, nil)
		var resObj controller.BridgeWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, [2]bool{true, false}, resObj.Vulnerability)
	})

	t.Run("games won and below line", func(t *testing.T) {
		m, _ := setupBridgeWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGamesWon")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGamesWon")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetBelowLine")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetBelowLine")
		m.On("GetGamesWon", 0).Return(1)
		m.On("GetGamesWon", 1).Return(0)
		m.On("GetBelowLine", 0).Return(60)
		m.On("GetBelowLine", 1).Return(30)

		result := p.Output(m, nil)
		var resObj controller.BridgeWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, [2]int{1, 0}, resObj.GamesWon)
		assert.Equal(t, [2]int{60, 30}, resObj.BelowLine)
	})

	t.Run("opening lead done with dummy hand", func(t *testing.T) {
		m, _ := setupBridgeWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsOpeningLeadDone")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDummyHand")
		m.On("IsOpeningLeadDone").Return(true)
		m.On("GetDummyHand").Return([]*domain.Card{
			domain.NewCard(domain.CardDesignHeart, 10, false),
		})

		result := p.Output(m, nil)
		var resObj controller.BridgeWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.True(t, resObj.OpeningLeadDone)
		assert.Len(t, resObj.DummyHand, 1)
		assert.Equal(t, "HEART", resObj.DummyHand[0].Design)
		assert.Equal(t, 10, resObj.DummyHand[0].Value)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupBridgeWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetConfig")
		m.On("GetConfig").Return(domain.BridgeConfig{
			CpuDifficulty: domain.BridgeCpuDifficultyHard,
		})

		result := p.Output(m, nil)
		var resObj controller.BridgeWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, int(domain.BridgeCpuDifficultyHard), resObj.Config.CpuDifficulty)
	})

	t.Run("error message", func(t *testing.T) {
		m, _ := setupBridgeWebMockWithPlayers()

		result := p.Output(m, errors.New("test error"))
		var resObj controller.BridgeWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "test error", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end team 0 wins", func(t *testing.T) {
		m, _ := setupBridgeWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)

		result := p.Output(m, nil)
		var resObj controller.BridgeWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.True(t, resObj.GameEndFlag)
		assert.Contains(t, resObj.Message, "チーム0")
		assert.Equal(t, "bridge.result.team0Win", resObj.MessageCode)
		assert.Equal(t, map[string]string{"team": "0"}, resObj.MessageParams)
	})

	t.Run("game end team 1 wins", func(t *testing.T) {
		m, _ := setupBridgeWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(1)

		result := p.Output(m, nil)
		var resObj controller.BridgeWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.True(t, resObj.GameEndFlag)
		assert.Contains(t, resObj.Message, "チーム1")
		assert.Equal(t, "bridge.result.team1Win", resObj.MessageCode)
	})

	t.Run("bid phase messageCode", func(t *testing.T) {
		m, _ := setupBridgeWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BridgePhaseBid)

		result := p.Output(m, nil)
		var resObj controller.BridgeWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "bridge.bidPhase", resObj.MessageCode)
	})

	t.Run("play phase lead messageCode when trick empty", func(t *testing.T) {
		m, _ := setupBridgeWebMockWithPlayers()

		result := p.Output(m, nil)
		var resObj controller.BridgeWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "bridge.playPhase.lead", resObj.MessageCode)
	})

	t.Run("play phase follow messageCode when trick has cards", func(t *testing.T) {
		m, _ := setupBridgeWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		trick := []*domain.TrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 3, false)},
		}
		m.On("GetCurrentTrick").Return(trick)

		result := p.Output(m, nil)
		var resObj controller.BridgeWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "bridge.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end messageCode", func(t *testing.T) {
		m, _ := setupBridgeWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BridgePhaseTrickEnd)

		result := p.Output(m, nil)
		var resObj controller.BridgeWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "bridge.trickEnd", resObj.MessageCode)
	})

	t.Run("round end messageCode", func(t *testing.T) {
		m, _ := setupBridgeWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BridgePhaseRoundEnd)

		result := p.Output(m, nil)
		var resObj controller.BridgeWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "bridge.roundEnd", resObj.MessageCode)
	})

	t.Run("error takes priority over phase message", func(t *testing.T) {
		m, _ := setupBridgeWebMockWithPlayers()

		result := p.Output(m, errors.New("some error"))
		var resObj controller.BridgeWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "some error", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end phase no messageCode for unknown phases", func(t *testing.T) {
		m, _ := setupBridgeWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BridgePhaseGameEnd)
		// GetGameEndFlag remains false (default)

		result := p.Output(m, nil)
		var resObj controller.BridgeWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Empty(t, resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("default config values", func(t *testing.T) {
		m, _ := setupBridgeWebMockWithPlayers()

		result := p.Output(m, nil)
		var resObj controller.BridgeWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, int(domain.BridgeCpuDifficultyNormal), resObj.Config.CpuDifficulty)
	})
}

func TestBridgeWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.BridgeWebPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockBridgeGame)
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
		m := new(interfaces.MockBridgeGame)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"entries":[]`)
		m.AssertExpectations(t)
	})

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockBridgeGame)
		m.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"entries":[]`)
		m.AssertExpectations(t)
	})
}

func TestBridgeWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.BridgeWebPresenter)

	t.Run("hint available with card", func(t *testing.T) {
		idx := 2
		m, _ := setupBridgeWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.BridgeHint{
			CardIndex: &idx,
			Reason:    "follow_suit",
		})

		result := p.HintOutput(m)
		var resObj controller.BridgeWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, &idx, resObj.Hint.CardIndex)
		assert.Equal(t, "follow_suit", resObj.Hint.Reason)
		assert.Equal(t, "bridge.hintRequested", resObj.MessageCode)
	})

	t.Run("hint available with bid", func(t *testing.T) {
		bidType := 1
		bidLevel := 2
		bidSuit := 3
		m, _ := setupBridgeWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.BridgeHint{
			BidType:  &bidType,
			BidLevel: &bidLevel,
			BidSuit:  &bidSuit,
			Reason:   "strong_hand",
		})

		result := p.HintOutput(m)
		var resObj controller.BridgeWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, &bidType, resObj.Hint.BidType)
		assert.Equal(t, &bidLevel, resObj.Hint.BidLevel)
		assert.Equal(t, &bidSuit, resObj.Hint.BidSuit)
		assert.Equal(t, "strong_hand", resObj.Hint.Reason)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupBridgeWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.BridgeHint)(nil))

		result := p.HintOutput(m)
		var resObj controller.BridgeWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Nil(t, resObj.Hint)
		assert.Equal(t, "bridge.noHint", resObj.MessageCode)
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestBridgeWebPresenterOutputCarriesTheHint(t *testing.T) {
	idx := 0
	brg, _ := setupBridgeWebMockWithPlayers()
	brg.ExpectedCalls = removeMockCall(brg.ExpectedCalls, "GetHint")
	brg.On("GetHint").Return(&domain.BridgeHint{CardIndex: &idx, Reason: "follow_suit"})

	result := new(presenter.BridgeWebPresenter).Output(brg, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	assert.NotContains(t, result, "bridge.hintRequested")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestBridgeWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	idx := 0
	brg, _ := setupBridgeWebMockWithPlayers()
	brg.ExpectedCalls = removeMockCall(brg.ExpectedCalls, "GetHint")
	brg.On("GetHint").Return(&domain.BridgeHint{CardIndex: &idx, Reason: "follow_suit"})
	assert.Contains(t, new(presenter.BridgeWebPresenter).HintOutput(brg), "bridge.hintRequested")

	none, _ := setupBridgeWebMockWithPlayers()
	none.ExpectedCalls = removeMockCall(none.ExpectedCalls, "GetHint")
	none.On("GetHint").Return((*domain.BridgeHint)(nil))
	assert.Contains(t, new(presenter.BridgeWebPresenter).HintOutput(none), "bridge.noHint")
}
