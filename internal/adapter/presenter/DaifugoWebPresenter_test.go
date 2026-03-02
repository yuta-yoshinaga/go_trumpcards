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

func TestDaifugoWebPresenter_Method(t *testing.T) {
	tdwp := presenter.NewDaifugoWebPresenter()

	makeDGPlayers := func() []*domain.DaifugoPlayer {
		return []*domain.DaifugoPlayer{
			domain.NewDaifugoPlayer(true),
			domain.NewDaifugoPlayer(false),
			domain.NewDaifugoPlayer(false),
			domain.NewDaifugoPlayer(false),
		}
	}

	setupDGWebTest := func() (*domain.Daifugo, []*domain.DaifugoPlayer) {
		tc := domain.NewTrumpCards(0)
		players := makeDGPlayers()
		dg := domain.NewDaifugo(tc, players, domain.DefaultDaifugoConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))
		return dg, players
	}

	t.Run("success Output initial state", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDGPlayers()
		dg := domain.NewDaifugo(tc, players, domain.DaifugoConfig{})
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := tdwp.Output(dg, nil)
		assert.NotEmpty(t, result)

		var resObj controller.DaifugoWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Equal(t, 4, len(resObj.Players))
		assert.False(t, resObj.GameEndFlag)
		assert.Empty(t, resObj.TableCards) // table is clear → serialized as []
		assert.Equal(t, 0, resObj.CurrentTurn)
		assert.Equal(t, -1, resObj.LastPlayPlayerIdx)
		assert.Equal(t, "", resObj.Message)
	})

	t.Run("success Output shows human cards", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDGPlayers()
		dg := domain.NewDaifugo(tc, players, domain.DaifugoConfig{})
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 9, false))

		result := tdwp.Output(dg, nil)
		var resObj controller.DaifugoWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		humanPlayer := resObj.Players[0]
		assert.True(t, humanPlayer.IsHuman)
		assert.Equal(t, 2, humanPlayer.CardCount)
		assert.Len(t, humanPlayer.Cards, 2)
		assert.Equal(t, "SPADE", humanPlayer.Cards[0].Design)
		assert.Equal(t, 3, humanPlayer.Cards[0].Value)
	})

	t.Run("success Output shows CPU card count but not cards", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDGPlayers()
		dg := domain.NewDaifugo(tc, players, domain.DaifugoConfig{})
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))

		result := tdwp.Output(dg, nil)
		var resObj controller.DaifugoWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		cpu1 := resObj.Players[1]
		assert.False(t, cpu1.IsHuman)
		assert.Equal(t, 2, cpu1.CardCount)
		assert.Len(t, cpu1.Cards, 0) // no cards shown for CPU
	})

	t.Run("success Output table cards after play", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDGPlayers()
		dg := domain.NewDaifugo(tc, players, domain.DaifugoConfig{})
		// Human has 2 cards → play index 0 → table gets [5], human keeps [7]
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		_ = dg.PlayerPlay([]int{0}) // play 5

		result := tdwp.Output(dg, nil)
		var resObj controller.DaifugoWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Len(t, resObj.TableCards, 1)
		assert.Equal(t, 5, resObj.TableCards[0].Value)
		assert.Equal(t, 0, resObj.LastPlayPlayerIdx)
	})

	t.Run("success Output game ended message", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDGPlayers()
		dg := domain.NewDaifugo(tc, players, domain.DaifugoConfig{})
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[1].SetIsFinished(true)
		players[1].SetRank(1)
		players[2].SetIsFinished(true)
		players[2].SetRank(2)
		players[3].SetIsFinished(true)
		players[3].SetRank(3)
		_ = dg.PlayerPlay([]int{0}) // human plays last card → game ends

		result := tdwp.Output(dg, nil)
		var resObj controller.DaifugoWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.True(t, resObj.GameEndFlag)
		assert.Contains(t, resObj.Message, "ゲーム終了")
	})

	t.Run("success Output human action after play", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDGPlayers()
		dg := domain.NewDaifugo(tc, players, domain.DaifugoConfig{})
		// Human has 2 cards → play index 0 (5)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		_ = dg.PlayerPlay([]int{0}) // play 5

		result := tdwp.Output(dg, nil)
		var resObj controller.DaifugoWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.NotNil(t, resObj.HumanAction)
		assert.Equal(t, 0, resObj.HumanAction.PlayerIdx)
		assert.Len(t, resObj.HumanAction.PlayedCards, 1)
		assert.Equal(t, 5, resObj.HumanAction.PlayedCards[0].Value)
	})

	t.Run("success Output human action pass has nil cards", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDGPlayers()
		dg := domain.NewDaifugo(tc, players, domain.DaifugoConfig{})
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		_ = dg.PlayerPlay([]int{}) // pass

		result := tdwp.Output(dg, nil)
		var resObj controller.DaifugoWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.NotNil(t, resObj.HumanAction)
		assert.Nil(t, resObj.HumanAction.PlayedCards) // pass → null in JSON
	})

	t.Run("success Output CPU actions all pass after unbeatable play", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDGPlayers()
		dg := domain.NewDaifugo(tc, players, domain.DaifugoConfig{})
		// Human has 2 cards: [2 (idx0), 3 (idx1)] — play 2 (strongest), keep 3
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false)) // idx0 → play
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false)) // idx1 → kept
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 4, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		_ = dg.PlayerPlay([]int{0}) // play 2 → human keeps [3], not finished
		dg.CpuPlay()                // CPU 1 passes
		dg.CpuPlay()                // CPU 2 passes
		dg.CpuPlay()                // CPU 3 passes → table clears

		result := tdwp.Output(dg, nil)
		var resObj controller.DaifugoWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Len(t, resObj.CpuActions, 3)
		for _, action := range resObj.CpuActions {
			assert.Nil(t, action.PlayedCards) // all passed
		}
	})

	t.Run("success Output revolutionActive is false initially", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDGPlayers()
		dg := domain.NewDaifugo(tc, players, domain.DaifugoConfig{})
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := tdwp.Output(dg, nil)
		var resObj controller.DaifugoWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.False(t, resObj.RevolutionActive)
	})

	t.Run("success Output revolutionActive is true after 4-card play", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDGPlayers()
		dg := domain.NewDaifugo(tc, players, domain.DaifugoConfig{})
		// Human plays four 5s → revolution
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignDiamond, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false)) // spare
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		_ = dg.PlayerPlay([]int{0, 1, 2, 3}) // play four 5s

		result := tdwp.Output(dg, nil)
		var resObj controller.DaifugoWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.True(t, resObj.RevolutionActive)
	})

	t.Run("success Output player rank reflects finished order", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDGPlayers()
		dg := domain.NewDaifugo(tc, players, domain.DaifugoConfig{})
		// 3 CPUs already finished (manually), human plays last card → gets rank 4
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[1].SetIsFinished(true)
		players[1].SetRank(1)
		players[2].SetIsFinished(true)
		players[2].SetRank(2)
		players[3].SetIsFinished(true)
		players[3].SetRank(3)
		_ = dg.PlayerPlay([]int{0}) // human plays last card → countFinished=3 → rank=4

		result := tdwp.Output(dg, nil)
		var resObj controller.DaifugoWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		humanPlayer := resObj.Players[0]
		assert.True(t, humanPlayer.IsFinished)
		assert.Equal(t, 4, humanPlayer.Rank)
	})

	t.Run("success Output shows error message when lastErr is non-nil", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDGPlayers()
		dg := domain.NewDaifugo(tc, players, domain.DaifugoConfig{})
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		testErr := errors.New("test error message")
		result := tdwp.Output(dg, testErr)
		var resObj controller.DaifugoWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Equal(t, "test error message", resObj.Message)
	})

	t.Run("success Output elevenBack suitLocked tableIsSequence flags", func(t *testing.T) {
		dg, _ := setupDGWebTest()
		dg.SetElevenBackActive(true)
		dg.SetSuitLocked(true, domain.CardDesignHeart)
		dg.SetTableIsSequence(true)
		result := tdwp.Output(dg, nil)
		var resObj controller.DaifugoWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.True(t, resObj.ElevenBackActive)
		assert.True(t, resObj.SuitLocked)
		assert.Equal(t, "HEART", resObj.LockedSuit)
		assert.True(t, resObj.TableIsSequence)
	})

	t.Run("success Output reverseDirection numberLocked sortMode fields", func(t *testing.T) {
		dg, _ := setupDGWebTest()
		dg.SetReverseDirection(true)
		dg.SetNumberLocked(true)
		dg.SetSortMode(domain.DaifugoSortBySuit)
		result := tdwp.Output(dg, nil)
		var resObj controller.DaifugoWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.True(t, resObj.ReverseDirection)
		assert.True(t, resObj.NumberLocked)
		assert.Equal(t, 1, resObj.SortMode)
	})

	t.Run("success Output new config fields mapped", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDGPlayers()
		cfg := domain.DefaultDaifugoConfig()
		cfg.NineReverseEnabled = true
		cfg.CoupDetatEnabled = true
		cfg.IntenseLockEnabled = true
		dg := domain.NewDaifugo(tc, players, cfg)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))
		result := tdwp.Output(dg, nil)
		var resObj controller.DaifugoWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.True(t, resObj.Config.NineReverseEnabled)
		assert.True(t, resObj.Config.CoupDetatEnabled)
		assert.True(t, resObj.Config.IntenseLockEnabled)
	})

	t.Run("success Output lockedSuit all suits", func(t *testing.T) {
		suitTests := []struct {
			suit     int
			expected string
		}{
			{domain.CardDesignSpade, "SPADE"},
			{domain.CardDesignClover, "CLOVER"},
			{domain.CardDesignDiamond, "DIAMOND"},
			{999, ""},
		}
		for _, st := range suitTests {
			dg, _ := setupDGWebTest()
			dg.SetSuitLocked(true, st.suit)
			result := tdwp.Output(dg, nil)
			var resObj controller.DaifugoWebOutput
			err := json.Unmarshal([]byte(result), &resObj)
			assert.NoError(t, err)
			assert.Equal(t, st.expected, resObj.LockedSuit)
		}
	})

	t.Run("success Output exchange actions in JSON", func(t *testing.T) {
		dg, _ := setupDGWebTest()
		dg.SetExchangeActions([]*domain.DaifugoExchangeAction{
			{FromPlayerIdx: 0, ToPlayerIdx: 3, Cards: []*domain.Card{domain.NewCard(domain.CardDesignSpade, 3, false)}},
		})
		result := tdwp.Output(dg, nil)
		var resObj controller.DaifugoWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Len(t, resObj.ExchangeActions, 1)
		assert.Equal(t, 0, resObj.ExchangeActions[0].FromPlayerIdx)
		assert.Equal(t, 3, resObj.ExchangeActions[0].ToPlayerIdx)
		assert.Len(t, resObj.ExchangeActions[0].Cards, 1)
		assert.Equal(t, "SPADE", resObj.ExchangeActions[0].Cards[0].Design)
	})

	t.Run("success Output buildResultMessage with rank out of bounds", func(t *testing.T) {
		dg, players := setupDGWebTest()
		players[1].SetIsFinished(true)
		players[1].SetRank(0) // out of bounds → skip in buildResultMessage
		players[2].SetIsFinished(true)
		players[2].SetRank(2)
		players[3].SetIsFinished(true)
		players[3].SetRank(3)
		_ = dg.PlayerPlay([]int{0})
		result := tdwp.Output(dg, nil)
		var resObj controller.DaifugoWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Contains(t, resObj.Message, "ゲーム終了")
		// CPU 1 with rank 0 should be skipped
		assert.NotContains(t, resObj.Message, "CPU 1")
	})

	t.Run("success Output getCardObj nil card", func(t *testing.T) {
		dg, _ := setupDGWebTest()
		// Set exchange action with nil card to exercise getCardObj nil branch
		dg.SetExchangeActions([]*domain.DaifugoExchangeAction{
			{FromPlayerIdx: 0, ToPlayerIdx: 3, Cards: []*domain.Card{nil}},
		})
		result := tdwp.Output(dg, nil)
		var resObj controller.DaifugoWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Len(t, resObj.ExchangeActions, 1)
		assert.Nil(t, resObj.ExchangeActions[0].Cards[0])
	})

	t.Run("success Output getCardObjs nil input", func(t *testing.T) {
		dg, _ := setupDGWebTest()
		// HumanAction with nil PlayedCards → exercises getCardObjs nil branch
		dg.SetHumanAction(&domain.DaifugoCpuAction{PlayerIdx: 0, PlayedCards: nil})
		result := tdwp.Output(dg, nil)
		var resObj controller.DaifugoWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.NotNil(t, resObj.HumanAction)
		assert.Nil(t, resObj.HumanAction.PlayedCards)
	})

	t.Run("success Output buildResultMessage with rank > 4", func(t *testing.T) {
		dg, players := setupDGWebTest()
		players[1].SetIsFinished(true)
		players[1].SetRank(5) // out of bounds (> 4) → skip in buildResultMessage
		players[2].SetIsFinished(true)
		players[2].SetRank(2)
		players[3].SetIsFinished(true)
		players[3].SetRank(3)
		_ = dg.PlayerPlay([]int{0})
		result := tdwp.Output(dg, nil)
		var resObj controller.DaifugoWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Contains(t, resObj.Message, "ゲーム終了")
		assert.NotContains(t, resObj.Message, "CPU 1")
	})

	setupDGMock := func() *interfaces.MockDaifugoGame {
		m := new(interfaces.MockDaifugoGame)
		m.On("GetCurrentTurn").Return(0)
		m.On("GetLastPlayPlayerIdx").Return(-1)
		m.On("GetRevolutionActive").Return(false)
		m.On("GetElevenBackActive").Return(false)
		m.On("GetSuitLocked").Return(false)
		m.On("GetLockedSuit").Return(0)
		m.On("GetTableIsSequence").Return(false)
		m.On("GetConfig").Return(domain.DaifugoConfig{})
		m.On("GetExchangeActions").Return([]*domain.DaifugoExchangeAction(nil))
		m.On("GetTableCards").Return([]*domain.Card(nil))
		m.On("GetCpuActions").Return([]*domain.DaifugoCpuAction(nil))
		m.On("GetHumanAction").Return((*domain.DaifugoCpuAction)(nil))
		m.On("GetReverseDirection").Return(false)
		m.On("GetNumberLocked").Return(false)
		m.On("GetSortMode").Return(domain.DaifugoSortByStrength)
		return m
	}

	t.Run("success Output skips nil player in player loop", func(t *testing.T) {
		m := setupDGMock()
		m.On("GetPendingActionType").Return(domain.DaifugoPendingNone)
		m.On("GetPendingActionTarget").Return(-1)
		m.On("GetGameEndFlag").Return(false)
		m.On("GetPlayerCnt").Return(1)
		m.On("GetPlayer", 0).Return((*domain.DaifugoPlayer)(nil))

		result := tdwp.Output(m, nil)
		var resObj controller.DaifugoWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Len(t, resObj.Players, 0) // nil player skipped
	})

	t.Run("success buildResultMessage skips nil player", func(t *testing.T) {
		m := setupDGMock()
		m.On("GetPendingActionType").Return(domain.DaifugoPendingNone)
		m.On("GetPendingActionTarget").Return(-1)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetPlayerCnt").Return(1)
		m.On("GetPlayer", 0).Return((*domain.DaifugoPlayer)(nil))

		result := tdwp.Output(m, nil)
		var resObj controller.DaifugoWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Len(t, resObj.Players, 0)           // nil player skipped in Output loop
		assert.Equal(t, "ゲーム終了！ ", resObj.Message) // buildResultMessage called
	})

	t.Run("success Output pendingAction sevenPass", func(t *testing.T) {
		m := setupDGMock()
		m.On("GetPendingActionType").Return(domain.DaifugoPendingSevenPass)
		m.On("GetPendingActionTarget").Return(1)
		m.On("GetGameEndFlag").Return(false)
		m.On("GetPlayerCnt").Return(0)

		result := tdwp.Output(m, nil)
		var resObj controller.DaifugoWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Equal(t, "sevenPass", resObj.PendingAction)
		assert.Equal(t, 1, resObj.PendingActionTarget)
	})

	t.Run("success Output pendingAction tenDiscard", func(t *testing.T) {
		m := setupDGMock()
		m.On("GetPendingActionType").Return(domain.DaifugoPendingTenDiscard)
		m.On("GetPendingActionTarget").Return(-1)
		m.On("GetGameEndFlag").Return(false)
		m.On("GetPlayerCnt").Return(0)

		result := tdwp.Output(m, nil)
		var resObj controller.DaifugoWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Equal(t, "tenDiscard", resObj.PendingAction)
	})
}
