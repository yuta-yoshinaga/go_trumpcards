package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

	"github.com/stretchr/testify/assert"
)

func TestDoubtWebPresenter_Method(t *testing.T) {
	tdwp := presenter.NewDoubtWebPresenter()

	makeDPlayers := func() []*domain.DoubtPlayer {
		return []*domain.DoubtPlayer{
			domain.NewDoubtPlayer(true),
			domain.NewDoubtPlayer(false),
			domain.NewDoubtPlayer(false),
			domain.NewDoubtPlayer(false),
		}
	}

	setupDoubtWebTest := func() (*domain.Doubt, []*domain.DoubtPlayer) {
		tc := domain.NewTrumpCards(0)
		players := makeDPlayers()
		d := domain.NewDoubt(tc, players)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 4, false))
		return d, players
	}

	t.Run("success Output initial state", func(t *testing.T) {
		d, _ := setupDoubtWebTest()
		result := tdwp.Output(d, nil)
		assert.NotEmpty(t, result)

		var resObj controller.DoubtWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Equal(t, 4, len(resObj.Players))
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, 0, resObj.CurrentTurn)
		assert.Equal(t, 0, resObj.Phase)
		assert.Equal(t, 0, resObj.TableCardCount)
		assert.Equal(t, "", resObj.Message)
		assert.Equal(t, -1, resObj.WinnerIdx)
		assert.Nil(t, resObj.LastAction)
		assert.Nil(t, resObj.HumanAction)
		assert.Nil(t, resObj.LastDoubtResult)
		assert.Equal(t, []int{}, resObj.CpuDoubters)
	})

	t.Run("success Output shows human cards", func(t *testing.T) {
		d, _ := setupDoubtWebTest()
		result := tdwp.Output(d, nil)
		var resObj controller.DoubtWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		humanPlayer := resObj.Players[0]
		assert.True(t, humanPlayer.IsHuman)
		assert.Equal(t, 1, humanPlayer.CardCount)
		assert.Len(t, humanPlayer.Cards, 1)
		assert.Equal(t, "SPADE", humanPlayer.Cards[0].Design)
		assert.Equal(t, 5, humanPlayer.Cards[0].Value)
	})

	t.Run("success Output CPU cards hidden", func(t *testing.T) {
		d, _ := setupDoubtWebTest()
		result := tdwp.Output(d, nil)
		var resObj controller.DoubtWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		cpu1 := resObj.Players[1]
		assert.False(t, cpu1.IsHuman)
		assert.Equal(t, 1, cpu1.CardCount)
		assert.Len(t, cpu1.Cards, 0) // no cards shown for CPU
	})

	t.Run("success Output lastAction non-nil", func(t *testing.T) {
		d, players := setupDoubtWebTest()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
		// Human plays card at index 0, claims value 5
		_ = d.PlayerPlay([]int{0}, 5)
		result := tdwp.Output(d, nil)
		var resObj controller.DoubtWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.NotNil(t, resObj.LastAction)
		assert.Equal(t, 0, resObj.LastAction.PlayerIdx)
		assert.Equal(t, 5, resObj.LastAction.ClaimedValue)
		assert.Equal(t, 1, resObj.LastAction.CardCount)
	})

	t.Run("success Output lastAction nil", func(t *testing.T) {
		d, _ := setupDoubtWebTest()
		// lastAction is nil by default
		result := tdwp.Output(d, nil)
		var resObj controller.DoubtWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Nil(t, resObj.LastAction)
	})

	t.Run("success Output humanAction non-nil", func(t *testing.T) {
		d, players := setupDoubtWebTest()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
		_ = d.PlayerPlay([]int{0}, 5)
		result := tdwp.Output(d, nil)
		var resObj controller.DoubtWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.NotNil(t, resObj.HumanAction)
		assert.Equal(t, 0, resObj.HumanAction.PlayerIdx)
		assert.Equal(t, 5, resObj.HumanAction.ClaimedValue)
	})

	t.Run("success Output humanAction nil", func(t *testing.T) {
		d, _ := setupDoubtWebTest()
		// humanAction is nil by default
		result := tdwp.Output(d, nil)
		var resObj controller.DoubtWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Nil(t, resObj.HumanAction)
	})

	t.Run("success Output cpuActions non-empty", func(t *testing.T) {
		d, _ := setupDoubtWebTest()
		d.SetCpuActions([]*domain.DoubtCpuAction{
			{PlayerIdx: 1, ClaimedValue: 3, CardCount: 1, IsBluff: true},
		})
		result := tdwp.Output(d, nil)
		var resObj controller.DoubtWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Len(t, resObj.CpuActions, 1)
		assert.Equal(t, 1, resObj.CpuActions[0].PlayerIdx)
		assert.Equal(t, 3, resObj.CpuActions[0].ClaimedValue)
		assert.False(t, resObj.CpuActions[0].IsBluff) // IsBluff is not sent to client (hidden game state)
	})

	t.Run("success Output cpuActions HasTell true is passed through", func(t *testing.T) {
		d, _ := setupDoubtWebTest()
		d.SetCpuActions([]*domain.DoubtCpuAction{
			{PlayerIdx: 1, ClaimedValue: 3, CardCount: 1, IsBluff: true, HasTell: true},
		})
		result := tdwp.Output(d, nil)
		var resObj controller.DoubtWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Len(t, resObj.CpuActions, 1)
		assert.True(t, resObj.CpuActions[0].HasTell)
	})

	t.Run("success Output cpuActions HasTell false by default", func(t *testing.T) {
		d, _ := setupDoubtWebTest()
		d.SetCpuActions([]*domain.DoubtCpuAction{
			{PlayerIdx: 1, ClaimedValue: 3, CardCount: 1, IsBluff: true},
		})
		result := tdwp.Output(d, nil)
		var resObj controller.DoubtWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Len(t, resObj.CpuActions, 1)
		assert.False(t, resObj.CpuActions[0].HasTell)
	})

	t.Run("success Output humanAction HasTell is passed through", func(t *testing.T) {
		d, _ := setupDoubtWebTest()
		d.SetHumanAction(&domain.DoubtCpuAction{
			PlayerIdx: 0, ClaimedValue: 5, CardCount: 1, HasTell: false,
		})
		result := tdwp.Output(d, nil)
		var resObj controller.DoubtWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.NotNil(t, resObj.HumanAction)
		assert.False(t, resObj.HumanAction.HasTell)
	})

	t.Run("success Output cpuActions HesitationMs is passed through", func(t *testing.T) {
		d, _ := setupDoubtWebTest()
		d.SetCpuActions([]*domain.DoubtCpuAction{
			{PlayerIdx: 1, ClaimedValue: 3, CardCount: 1, HesitationMs: 750},
		})
		result := tdwp.Output(d, nil)
		var resObj controller.DoubtWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Len(t, resObj.CpuActions, 1)
		assert.Equal(t, 750, resObj.CpuActions[0].HesitationMs)
	})

	t.Run("success Output cpuActions HesitationMs zero by default", func(t *testing.T) {
		d, _ := setupDoubtWebTest()
		d.SetCpuActions([]*domain.DoubtCpuAction{
			{PlayerIdx: 1, ClaimedValue: 3, CardCount: 1},
		})
		result := tdwp.Output(d, nil)
		var resObj controller.DoubtWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Len(t, resObj.CpuActions, 1)
		assert.Equal(t, 0, resObj.CpuActions[0].HesitationMs)
	})

	t.Run("success Output cpuDoubters non-empty", func(t *testing.T) {
		d, _ := setupDoubtWebTest()
		d.SetCpuDoubters([]int{1, 2})
		result := tdwp.Output(d, nil)
		var resObj controller.DoubtWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, []int{1, 2}, resObj.CpuDoubters)
	})

	t.Run("success Output cpuDoubters nil converted to empty slice", func(t *testing.T) {
		d, _ := setupDoubtWebTest()
		// cpuDoubters is nil by default → should return []int{}
		result := tdwp.Output(d, nil)
		var resObj controller.DoubtWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, []int{}, resObj.CpuDoubters)
	})

	t.Run("success Output lastDoubtResult wasLying true", func(t *testing.T) {
		d, _ := setupDoubtWebTest()
		d.SetLastDoubtResult(&domain.DoubtDoubtResult{
			DoubterIdx:    1,
			CardPlayerIdx: 0,
			WasLying:      true,
			LoserIdx:      0,
			CardCount:     3,
			RevealedCards: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 5, false),
			},
		})
		result := tdwp.Output(d, nil)
		var resObj controller.DoubtWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.NotNil(t, resObj.LastDoubtResult)
		assert.Equal(t, 1, resObj.LastDoubtResult.DoubterIdx)
		assert.Equal(t, 0, resObj.LastDoubtResult.CardPlayerIdx)
		assert.True(t, resObj.LastDoubtResult.WasLying)
		assert.Equal(t, 0, resObj.LastDoubtResult.LoserIdx)
		assert.Equal(t, 3, resObj.LastDoubtResult.CardCount)
		assert.Len(t, resObj.LastDoubtResult.RevealedCards, 1)
		assert.Equal(t, "SPADE", resObj.LastDoubtResult.RevealedCards[0].Design)
		assert.Equal(t, 5, resObj.LastDoubtResult.RevealedCards[0].Value)
	})

	t.Run("success Output lastDoubtResult wasLying false", func(t *testing.T) {
		d, _ := setupDoubtWebTest()
		d.SetLastDoubtResult(&domain.DoubtDoubtResult{
			DoubterIdx:    2,
			CardPlayerIdx: 0,
			WasLying:      false,
			LoserIdx:      2,
			CardCount:     2,
			RevealedCards: []*domain.Card{},
		})
		result := tdwp.Output(d, nil)
		var resObj controller.DoubtWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.NotNil(t, resObj.LastDoubtResult)
		assert.False(t, resObj.LastDoubtResult.WasLying)
		assert.Equal(t, 2, resObj.LastDoubtResult.LoserIdx)
	})

	t.Run("success Output gameEndFlag human wins", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDPlayers()
		d := domain.NewDoubt(tc, players)
		// Human has only 1 card → wins after playing it
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 4, false))
		_ = d.PlayerPlay([]int{0}, 5) // human plays last card → wins
		result := tdwp.Output(d, nil)
		var resObj controller.DoubtWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.True(t, resObj.GameEndFlag)
		assert.Contains(t, resObj.Message, "ゲーム終了")
		assert.Contains(t, resObj.Message, "あなた")
		assert.Equal(t, "doubt.result.humanWin", resObj.MessageCode)
		assert.Nil(t, resObj.MessageParams)
	})

	t.Run("success Output gameEndFlag CPU wins", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDPlayers()
		d := domain.NewDoubt(tc, players)
		// Human has 2 cards, CPU 1 has 1 card (will win after playing)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 4, false))
		_ = d.PlayerPlay([]int{0}, 5) // human plays → DoubtPhase
		d.ResolveDoubt(nil)           // skip → currentTurn=1, Play phase
		d.CpuPlay()                   // CPU 1 plays 1 card → wins
		result := tdwp.Output(d, nil)
		var resObj controller.DoubtWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.True(t, resObj.GameEndFlag)
		assert.Contains(t, resObj.Message, "ゲーム終了")
		assert.Contains(t, resObj.Message, "CPU 1")
		assert.Equal(t, "doubt.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"cpuId": "1"}, resObj.MessageParams)
	})

	t.Run("success Output error message", func(t *testing.T) {
		d, _ := setupDoubtWebTest()
		result := tdwp.Output(d, errors.New("test error"))
		var resObj controller.DoubtWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Contains(t, resObj.Message, "test error")
	})

	t.Run("success Output doubtWindowSec reflects config default", func(t *testing.T) {
		d, _ := setupDoubtWebTest()
		result := tdwp.Output(d, nil)
		var resObj controller.DoubtWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Equal(t, 10, resObj.DoubtWindowSec)
	})

	t.Run("success Output doubtWindowSec reflects custom config", func(t *testing.T) {
		d, _ := setupDoubtWebTest()
		d.SetConfig(domain.DoubtConfig{DoubtWindowSec: 3, CpuMemoryLevel: domain.DoubtMemoryLevelHard})
		result := tdwp.Output(d, nil)
		var resObj controller.DoubtWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Equal(t, 3, resObj.DoubtWindowSec)
	})

	t.Run("success Output penaltyDrawLimit reflects config", func(t *testing.T) {
		d, _ := setupDoubtWebTest()
		d.SetConfig(domain.DoubtConfig{DoubtWindowSec: 10, CpuMemoryLevel: domain.DoubtMemoryLevelNormal, PenaltyDrawLimit: 5})
		result := tdwp.Output(d, nil)
		var resObj controller.DoubtWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Equal(t, 5, resObj.PenaltyDrawLimit)
	})

	t.Run("success Output penaltyDrawLimit default 0", func(t *testing.T) {
		d, _ := setupDoubtWebTest()
		result := tdwp.Output(d, nil)
		var resObj controller.DoubtWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Equal(t, 0, resObj.PenaltyDrawLimit)
	})

	t.Run("success Output lastDoubtResult discardedCount > 0", func(t *testing.T) {
		d, _ := setupDoubtWebTest()
		d.SetLastDoubtResult(&domain.DoubtDoubtResult{
			DoubterIdx:     1,
			CardPlayerIdx:  0,
			WasLying:       true,
			LoserIdx:       0,
			CardCount:      3,
			DiscardedCount: 2,
			RevealedCards:  []*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)},
		})
		result := tdwp.Output(d, nil)
		var resObj controller.DoubtWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.NotNil(t, resObj.LastDoubtResult)
		assert.Equal(t, 3, resObj.LastDoubtResult.CardCount)
		assert.Equal(t, 2, resObj.LastDoubtResult.DiscardedCount)
	})

	t.Run("success Output lastDoubtResult discardedCount 0", func(t *testing.T) {
		d, _ := setupDoubtWebTest()
		d.SetLastDoubtResult(&domain.DoubtDoubtResult{
			DoubterIdx:     1,
			CardPlayerIdx:  0,
			WasLying:       true,
			LoserIdx:       0,
			CardCount:      5,
			DiscardedCount: 0,
			RevealedCards:  []*domain.Card{},
		})
		result := tdwp.Output(d, nil)
		var resObj controller.DoubtWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.NotNil(t, resObj.LastDoubtResult)
		assert.Equal(t, 0, resObj.LastDoubtResult.DiscardedCount)
	})

	t.Run("success Output metaAI populated when profile exists", func(t *testing.T) {
		d, _ := setupDoubtWebTest()
		d.SetHumanProfile(&domain.DoubtHumanProfile{
			GamesPlayed:     3,
			BluffsByBracket: [3]struct{ Bluffs, Total int }{{1, 4}, {2, 5}, {0, 3}},
			DoubtCorrect:    3,
			DoubtTotal:      4,
		})
		result := tdwp.Output(d, nil)
		var resObj controller.DoubtWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.NotNil(t, resObj.MetaAI)
		assert.True(t, resObj.MetaAI.Enabled)
		assert.Equal(t, 3, resObj.MetaAI.GamesPlayed)
		assert.InDelta(t, 0.4, resObj.MetaAI.BluffRate, 0.001)      // 2/5
		assert.InDelta(t, 0.75, resObj.MetaAI.DoubtAccuracy, 0.001) // 3/4
	})

	t.Run("success Output metaAI omitted when profile is nil", func(t *testing.T) {
		d, _ := setupDoubtWebTest()
		// No SetHumanProfile → profile is nil
		result := tdwp.Output(d, nil)
		var resObj controller.DoubtWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Nil(t, resObj.MetaAI)
		// Also verify no metaAI key in raw JSON (omitempty)
		var raw map[string]interface{}
		_ = json.Unmarshal([]byte(result), &raw)
		_, hasMetaAI := raw["metaAI"]
		assert.False(t, hasMetaAI)
	})

	t.Run("success Output gameEndFlag nil player at winnerIdx", func(t *testing.T) {
		gameMock := new(interfaces.MockDoubtGame)
		gameMock.On("GetCurrentTurn").Return(0)
		gameMock.On("GetPhase").Return(domain.DoubtPhasePlay)
		gameMock.On("GetTableCardCount").Return(0)
		gameMock.On("GetGameEndFlag").Return(true)
		gameMock.On("GetWinnerIdx").Return(99)
		gameMock.On("GetCpuActions").Return([]*domain.DoubtCpuAction{})
		gameMock.On("GetHumanAction").Return((*domain.DoubtCpuAction)(nil))
		gameMock.On("GetLastAction").Return((*domain.DoubtAction)(nil))
		gameMock.On("GetCpuDoubters").Return([]int(nil))
		gameMock.On("GetLastDoubtResult").Return((*domain.DoubtDoubtResult)(nil))
		gameMock.On("GetPlayerCnt").Return(0)
		gameMock.On("GetPlayer", 99).Return((*domain.DoubtPlayer)(nil))
		gameMock.On("GetConfig").Return(domain.DefaultDoubtConfig())
		gameMock.On("GetHumanProfile").Return((*domain.DoubtHumanProfile)(nil))

		result := tdwp.Output(gameMock, nil)
		var resObj controller.DoubtWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.True(t, resObj.GameEndFlag)
		assert.Contains(t, resObj.Message, "ゲーム終了")
		assert.Contains(t, resObj.Message, "CPU 99")
		assert.Equal(t, "doubt.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"cpuId": "99"}, resObj.MessageParams)
	})
}

func TestDoubtWebPresenter_ActionLogOutput(t *testing.T) {
	p := presenter.NewDoubtWebPresenter()

	t.Run("with entries", func(t *testing.T) {
		mockGame := new(interfaces.MockDoubtGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 0, PlayerIdx: 0, ActionType: "play", Detail: "declared 5, played 1 card(s)", Cards: []*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, true)}},
		}
		mockGame.On("GetGameEndFlag").Return(true)
		mockGame.On("GetActionLog").Return(entries)

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, `"actionType":"play"`)
		assert.Contains(t, result, `"detail":"declared 5, played 1 card(s)"`)
		assert.Contains(t, result, `"turnNumber":0`)
		assert.Contains(t, result, `"playerIdx":0`)
		mockGame.AssertExpectations(t)
	})

	t.Run("nil_entries", func(t *testing.T) {
		mockGame := new(interfaces.MockDoubtGame)
		mockGame.On("GetGameEndFlag").Return(true)
		mockGame.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, `"entries":[]`)
		mockGame.AssertExpectations(t)
	})

	t.Run("game_not_ended", func(t *testing.T) {
		mockGame := new(interfaces.MockDoubtGame)
		mockGame.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, `"entries":[]`)
		mockGame.AssertExpectations(t)
	})
}
