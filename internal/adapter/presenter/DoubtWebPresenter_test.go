package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

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
		assert.True(t, resObj.CpuActions[0].IsBluff)
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
	})

	t.Run("success Output error message", func(t *testing.T) {
		d, _ := setupDoubtWebTest()
		result := tdwp.Output(d, errors.New("test error"))
		var resObj controller.DoubtWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Contains(t, resObj.Message, "test error")
	})
}
