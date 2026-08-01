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

// setupNapoleonWebMock creates a MockNapoleonGame with sensible defaults for Web tests.
func setupNapoleonWebMock() *interfaces.MockNapoleonGame {
	m := new(interfaces.MockNapoleonGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetTrumpSuit").Return(0)
	m.On("GetAdjutantCard").Return((*domain.Card)(nil))
	m.On("GetNapoleonIdx").Return(-1)
	m.On("GetAdjutantIdx").Return(-1)
	m.On("GetAdjutantRevealed").Return(false)
	m.On("GetHighestBid").Return(0)
	m.On("GetHighestBidder").Return(-1)
	m.On("GetKitty").Return(([]*domain.Card)(nil))
	m.On("GetCurrentTrick").Return([]*domain.NapoleonTrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.NapoleonPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetWinnerTeam").Return(domain.NapoleonWinnerUndecided)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultNapoleonConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	// **base だけに置く。**removeMockCall は最初の 1 件しか外さない。
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func setupNapoleonWebMockWithPlayers() (*interfaces.MockNapoleonGame, []*domain.NapoleonPlayer) {
	m := setupNapoleonWebMock()
	players := makeNapoleonPlayers()
	m.On("GetPlayerCnt").Return(5)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	m.On("GetPlayer", 4).Return(players[4])
	return m, players
}

func removeNapoleonWebMockCall(calls []*mock.Call, method string) []*mock.Call {
	return removeMockCall(calls, method)
}

func TestNapoleonWebPresenter_Output(t *testing.T) {
	p := new(presenter.NapoleonWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupNapoleonWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		assert.NotEmpty(t, result)

		var resObj controller.NapoleonWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Equal(t, 5, len(resObj.Players))
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, 0, resObj.CurrentPlayerIdx)
		assert.Equal(t, 3, resObj.Phase) // NapoleonPhasePlay
		assert.Equal(t, 1, resObj.RoundNumber)
		assert.Equal(t, 1, resObj.TrickNumber)
		assert.Equal(t, -1, resObj.WinnerTeam)
		assert.Equal(t, 0, resObj.LeadPlayerIdx)
		assert.Equal(t, "", resObj.Message)
		assert.Empty(t, resObj.CurrentTrick)
	})

	t.Run("human cards shown, CPU cards hidden", func(t *testing.T) {
		m, players := setupNapoleonWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.NapoleonWebOutput
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

	t.Run("player bid, scores, tricks and pictureCards", func(t *testing.T) {
		m, players := setupNapoleonWebMockWithPlayers()
		players[1].SetCumulativeScore(50)
		players[1].SetRoundScore(10)
		players[1].SetPictureCards(3)
		players[1].SetBid(14)
		players[1].AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 3, false)})

		result := p.Output(m, nil)
		var resObj controller.NapoleonWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, 50, resObj.Players[1].CumulativeScore)
		assert.Equal(t, 10, resObj.Players[1].RoundScore)
		assert.Equal(t, 1, resObj.Players[1].TrickCount)
		assert.Equal(t, 3, resObj.Players[1].PictureCards)
		assert.Equal(t, 14, resObj.Players[1].Bid)
	})

	t.Run("napoleon and adjutant flags", func(t *testing.T) {
		m, players := setupNapoleonWebMockWithPlayers()
		m.ExpectedCalls = removeNapoleonWebMockCall(m.ExpectedCalls, "GetAdjutantRevealed")
		m.On("GetAdjutantRevealed").Return(true)
		players[0].SetIsNapoleon(true)
		players[1].SetIsAdjutant(true)

		result := p.Output(m, nil)
		var resObj controller.NapoleonWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.True(t, resObj.Players[0].IsNapoleon)
		assert.True(t, resObj.Players[1].IsAdjutant)
		assert.True(t, resObj.AdjutantRevealed)
	})

	t.Run("adjutant not shown when not revealed", func(t *testing.T) {
		m, players := setupNapoleonWebMockWithPlayers()
		players[1].SetIsAdjutant(true)

		result := p.Output(m, nil)
		var resObj controller.NapoleonWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.False(t, resObj.Players[1].IsAdjutant)
	})

	t.Run("current trick populated", func(t *testing.T) {
		m, _ := setupNapoleonWebMockWithPlayers()
		m.ExpectedCalls = removeNapoleonWebMockCall(m.ExpectedCalls, "GetCurrentTrick")
		trick := []*domain.NapoleonTrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 3, false)},
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 7, false)},
		}
		m.On("GetCurrentTrick").Return(trick)

		result := p.Output(m, nil)
		var resObj controller.NapoleonWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Len(t, resObj.CurrentTrick, 2)
		assert.Equal(t, 0, resObj.CurrentTrick[0].PlayerIdx)
		assert.Equal(t, "CLOVER", resObj.CurrentTrick[0].Card.Design)
		assert.Equal(t, 3, resObj.CurrentTrick[0].Card.Value)
		assert.Equal(t, 1, resObj.CurrentTrick[1].PlayerIdx)
	})

	t.Run("empty current trick", func(t *testing.T) {
		m, _ := setupNapoleonWebMockWithPlayers()

		result := p.Output(m, nil)
		var resObj controller.NapoleonWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Empty(t, resObj.CurrentTrick)
	})

	t.Run("adjutant card set", func(t *testing.T) {
		m, _ := setupNapoleonWebMockWithPlayers()
		m.ExpectedCalls = removeNapoleonWebMockCall(m.ExpectedCalls, "GetAdjutantCard")
		m.On("GetAdjutantCard").Return(domain.NewCard(domain.CardDesignHeart, 13, false))

		result := p.Output(m, nil)
		var resObj controller.NapoleonWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.NotNil(t, resObj.AdjutantCard)
		assert.Equal(t, "HEART", resObj.AdjutantCard.Design)
		assert.Equal(t, 13, resObj.AdjutantCard.Value)
	})

	t.Run("kitty shown in exchange phase", func(t *testing.T) {
		m, _ := setupNapoleonWebMockWithPlayers()
		m.ExpectedCalls = removeNapoleonWebMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeNapoleonWebMockCall(m.ExpectedCalls, "GetKitty")
		m.On("GetPhase").Return(domain.NapoleonPhaseKittyExchange)
		kitty := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 5, false),
			domain.NewCard(domain.CardDesignDiamond, 10, false),
		}
		m.On("GetKitty").Return(kitty)

		result := p.Output(m, nil)
		var resObj controller.NapoleonWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Len(t, resObj.Kitty, 2)
		assert.Equal(t, "SPADE", resObj.Kitty[0].Design)
		assert.Equal(t, 5, resObj.Kitty[0].Value)
	})

	t.Run("kitty not shown in play phase", func(t *testing.T) {
		m, _ := setupNapoleonWebMockWithPlayers()

		result := p.Output(m, nil)
		var resObj controller.NapoleonWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Nil(t, resObj.Kitty)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupNapoleonWebMockWithPlayers()
		m.ExpectedCalls = removeNapoleonWebMockCall(m.ExpectedCalls, "GetConfig")
		m.On("GetConfig").Return(domain.NapoleonConfig{
			CpuDifficulty: domain.NapoleonCpuDifficultyHard,
			MinBid:        15,
			PointLimit:    200,
		})

		result := p.Output(m, nil)
		var resObj controller.NapoleonWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, int(domain.NapoleonCpuDifficultyHard), resObj.Config.CpuDifficulty)
		assert.Equal(t, 15, resObj.Config.MinBid)
		assert.Equal(t, 200, resObj.Config.PointLimit)
	})

	t.Run("error message", func(t *testing.T) {
		m, _ := setupNapoleonWebMockWithPlayers()

		result := p.Output(m, errors.New("test error"))
		var resObj controller.NapoleonWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "test error", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end napoleon wins", func(t *testing.T) {
		m, _ := setupNapoleonWebMockWithPlayers()
		m.ExpectedCalls = removeNapoleonWebMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeNapoleonWebMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(domain.NapoleonWinnerNapoleon)

		result := p.Output(m, nil)
		var resObj controller.NapoleonWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "napoleon.gameEnd.napoleonWins", resObj.MessageCode)
	})

	t.Run("game end allied wins", func(t *testing.T) {
		m, _ := setupNapoleonWebMockWithPlayers()
		m.ExpectedCalls = removeNapoleonWebMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeNapoleonWebMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(domain.NapoleonWinnerAllied)

		result := p.Output(m, nil)
		var resObj controller.NapoleonWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "napoleon.gameEnd.alliedWins", resObj.MessageCode)
	})

	t.Run("bid phase messageCode", func(t *testing.T) {
		m, _ := setupNapoleonWebMockWithPlayers()
		m.ExpectedCalls = removeNapoleonWebMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.NapoleonPhaseBid)

		result := p.Output(m, nil)
		var resObj controller.NapoleonWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "napoleon.bidPhase", resObj.MessageCode)
	})

	t.Run("trump declaration phase messageCode", func(t *testing.T) {
		m, _ := setupNapoleonWebMockWithPlayers()
		m.ExpectedCalls = removeNapoleonWebMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.NapoleonPhaseTrumpDeclaration)

		result := p.Output(m, nil)
		var resObj controller.NapoleonWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "napoleon.trumpDeclaration", resObj.MessageCode)
	})

	t.Run("kitty exchange phase messageCode", func(t *testing.T) {
		m, _ := setupNapoleonWebMockWithPlayers()
		m.ExpectedCalls = removeNapoleonWebMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.NapoleonPhaseKittyExchange)

		result := p.Output(m, nil)
		var resObj controller.NapoleonWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "napoleon.kittyExchange", resObj.MessageCode)
	})

	t.Run("play phase lead messageCode when trick empty", func(t *testing.T) {
		m, _ := setupNapoleonWebMockWithPlayers()
		// Default: phase=Play, currentTrick=nil (empty)

		result := p.Output(m, nil)
		var resObj controller.NapoleonWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "napoleon.playPhase.lead", resObj.MessageCode)
	})

	t.Run("play phase follow messageCode when trick has cards", func(t *testing.T) {
		m, _ := setupNapoleonWebMockWithPlayers()
		m.ExpectedCalls = removeNapoleonWebMockCall(m.ExpectedCalls, "GetCurrentTrick")
		trick := []*domain.NapoleonTrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 3, false)},
		}
		m.On("GetCurrentTrick").Return(trick)

		result := p.Output(m, nil)
		var resObj controller.NapoleonWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "napoleon.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end messageCode", func(t *testing.T) {
		m, _ := setupNapoleonWebMockWithPlayers()
		m.ExpectedCalls = removeNapoleonWebMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.NapoleonPhaseTrickEnd)

		result := p.Output(m, nil)
		var resObj controller.NapoleonWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "napoleon.trickEnd", resObj.MessageCode)
	})

	t.Run("round end messageCode", func(t *testing.T) {
		m, _ := setupNapoleonWebMockWithPlayers()
		m.ExpectedCalls = removeNapoleonWebMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.NapoleonPhaseRoundEnd)

		result := p.Output(m, nil)
		var resObj controller.NapoleonWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "napoleon.roundEnd", resObj.MessageCode)
	})

	t.Run("error takes priority over phase message", func(t *testing.T) {
		m, _ := setupNapoleonWebMockWithPlayers()

		result := p.Output(m, errors.New("some error"))
		var resObj controller.NapoleonWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "some error", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end phase messageCode", func(t *testing.T) {
		m, _ := setupNapoleonWebMockWithPlayers()
		m.ExpectedCalls = removeNapoleonWebMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeNapoleonWebMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeNapoleonWebMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetPhase").Return(domain.NapoleonPhaseGameEnd)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(domain.NapoleonWinnerNapoleon)

		result := p.Output(m, nil)
		var resObj controller.NapoleonWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "napoleon.gameEnd.napoleonWins", resObj.MessageCode)
	})

	t.Run("unrecognized phase no messageCode", func(t *testing.T) {
		m, _ := setupNapoleonWebMockWithPlayers()
		m.ExpectedCalls = removeNapoleonWebMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.NapoleonPhaseGameEnd)
		// GetGameEndFlag remains false (default)

		result := p.Output(m, nil)
		var resObj controller.NapoleonWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Empty(t, resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("default config values", func(t *testing.T) {
		m, _ := setupNapoleonWebMockWithPlayers()

		result := p.Output(m, nil)
		var resObj controller.NapoleonWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, int(domain.NapoleonCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, 12, resObj.Config.MinBid)
		assert.Equal(t, 100, resObj.Config.PointLimit)
	})
}

func TestNapoleonWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.NapoleonWebPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockNapoleonGame)
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
		m := new(interfaces.MockNapoleonGame)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"entries":[]`)
		m.AssertExpectations(t)
	})

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockNapoleonGame)
		m.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"entries":[]`)
		m.AssertExpectations(t)
	})
}

func TestNapoleonWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.NapoleonWebPresenter)

	t.Run("hint available with card", func(t *testing.T) {
		idx := 2
		m, _ := setupNapoleonWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.NapoleonHint{
			CardIndex: &idx,
			Reason:    "follow_suit",
		})

		result := p.HintOutput(m)
		var resObj controller.NapoleonWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, &idx, resObj.Hint.CardIndex)
		assert.Equal(t, "follow_suit", resObj.Hint.Reason)
		assert.Equal(t, "napoleon.hintRequested", resObj.MessageCode)
	})

	t.Run("hint available with bid", func(t *testing.T) {
		bid := 14
		m, _ := setupNapoleonWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.NapoleonHint{
			Bid:    &bid,
			Reason: "strategic_bid",
		})

		result := p.HintOutput(m)
		var resObj controller.NapoleonWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, &bid, resObj.Hint.Bid)
		assert.Equal(t, "strategic_bid", resObj.Hint.Reason)
		assert.Equal(t, "napoleon.hintRequested", resObj.MessageCode)
	})

	t.Run("hint available with trump suit", func(t *testing.T) {
		suit := 1
		adjSuit := 3
		adjVal := 13
		m, _ := setupNapoleonWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.NapoleonHint{
			TrumpSuit:     &suit,
			AdjutantSuit:  &adjSuit,
			AdjutantValue: &adjVal,
			Reason:        "strategic_declare",
		})

		result := p.HintOutput(m)
		var resObj controller.NapoleonWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, &suit, resObj.Hint.TrumpSuit)
		assert.Equal(t, &adjSuit, resObj.Hint.AdjutantSuit)
		assert.Equal(t, &adjVal, resObj.Hint.AdjutantValue)
		assert.Equal(t, "strategic_declare", resObj.Hint.Reason)
	})

	t.Run("hint available with discard", func(t *testing.T) {
		idx := 3
		m, _ := setupNapoleonWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.NapoleonHint{
			DiscardIndex: &idx,
			Reason:       "strategic_discard",
		})

		result := p.HintOutput(m)
		var resObj controller.NapoleonWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, &idx, resObj.Hint.DiscardIndex)
		assert.Equal(t, "strategic_discard", resObj.Hint.Reason)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupNapoleonWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.NapoleonHint)(nil))

		result := p.HintOutput(m)
		var resObj controller.NapoleonWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Nil(t, resObj.Hint)
		assert.Equal(t, "napoleon.noHint", resObj.MessageCode)
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestNapoleonWebPresenterOutputCarriesTheHint(t *testing.T) {
	idx := 0
	npg, _ := setupNapoleonWebMockWithPlayers()
	npg.ExpectedCalls = removeMockCall(npg.ExpectedCalls, "GetHint")
	npg.On("GetHint").Return(&domain.NapoleonHint{CardIndex: &idx, Reason: "follow_suit"})

	result := new(presenter.NapoleonWebPresenter).Output(npg, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	assert.NotContains(t, result, "napoleon.hintRequested")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestNapoleonWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	idx := 0
	npg, _ := setupNapoleonWebMockWithPlayers()
	npg.ExpectedCalls = removeMockCall(npg.ExpectedCalls, "GetHint")
	npg.On("GetHint").Return(&domain.NapoleonHint{CardIndex: &idx, Reason: "follow_suit"})
	assert.Contains(t, new(presenter.NapoleonWebPresenter).HintOutput(npg), "napoleon.hintRequested")

	none, _ := setupNapoleonWebMockWithPlayers()
	none.ExpectedCalls = removeMockCall(none.ExpectedCalls, "GetHint")
	none.On("GetHint").Return((*domain.NapoleonHint)(nil))
	assert.Contains(t, new(presenter.NapoleonWebPresenter).HintOutput(none), "napoleon.noHint")
}
