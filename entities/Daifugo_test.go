package entities_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"

	"github.com/stretchr/testify/assert"
)

func makeDaifugoPlayers() []*entities.DaifugoPlayer {
	return []*entities.DaifugoPlayer{
		entities.NewDaifugoPlayer(true),
		entities.NewDaifugoPlayer(false),
		entities.NewDaifugoPlayer(false),
		entities.NewDaifugoPlayer(false),
	}
}

func TestDaifugo_Method(t *testing.T) {
	t.Run("success NewDaifugo", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		assert.NotNil(t, dg)
		assert.Equal(t, 4, dg.GetPlayerCnt())
		assert.False(t, dg.GetGameEndFlag())
		assert.Nil(t, dg.GetTableCards())
		assert.Equal(t, -1, dg.GetLastPlayPlayerIdx())
		assert.Equal(t, 0, dg.GetCurrentTurn())
	})

	t.Run("success Reset distributes 52 cards", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		dg.Reset()
		total := 0
		for i := 0; i < dg.GetPlayerCnt(); i++ {
			total += dg.GetPlayer(i).GetCardsSize()
		}
		assert.Equal(t, 52, total)
	})

	t.Run("success Reset clears state", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		dg.Reset()
		assert.False(t, dg.GetGameEndFlag())
		assert.Nil(t, dg.GetTableCards())
		assert.Equal(t, -1, dg.GetLastPlayPlayerIdx())
		assert.Equal(t, 0, dg.GetPassCount())
		assert.Nil(t, dg.GetHumanAction())
		assert.Nil(t, dg.GetCpuActions())
	})

	t.Run("success GetPlayer valid index", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		assert.NotNil(t, dg.GetPlayer(0))
		assert.True(t, dg.GetPlayer(0).GetIsHuman())
		assert.NotNil(t, dg.GetPlayer(1))
		assert.False(t, dg.GetPlayer(1).GetIsHuman())
	})

	t.Run("success GetPlayer invalid index returns nil", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		assert.Nil(t, dg.GetPlayer(-1))
		assert.Nil(t, dg.GetPlayer(10))
	})

	t.Run("success IsHumanTurn at start", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		assert.True(t, dg.IsHumanTurn())
	})

	t.Run("success PlayerPlay on clear table", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		// Human has 2 cards so they don't finish when playing one
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 7, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))

		ok := dg.PlayerPlay([]int{0}) // play 3
		assert.True(t, ok)
		assert.NotNil(t, dg.GetTableCards())
		assert.Equal(t, 3, dg.GetTableCards()[0].GetValue())
		assert.Equal(t, 0, dg.GetLastPlayPlayerIdx())
		assert.Equal(t, 1, players[0].GetCardsSize())
	})

	t.Run("success PlayerPlay fails with invalid index", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))

		ok := dg.PlayerPlay([]int{5}) // out of range
		assert.False(t, ok)
	})

	t.Run("success PlayerPlay fails with different values", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))

		ok := dg.PlayerPlay([]int{0, 1}) // different values → invalid
		assert.False(t, ok)
	})

	t.Run("success PlayerPlay table card stays after valid play", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 7, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		ok := dg.PlayerPlay([]int{0}) // play 7
		assert.True(t, ok)
		assert.Equal(t, 7, dg.GetTableCards()[0].GetValue())
	})

	t.Run("success PlayerPlay pass", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))

		ok := dg.PlayerPlay([]int{}) // pass
		assert.True(t, ok)
		assert.Equal(t, 1, dg.GetPassCount())
		assert.NotNil(t, dg.GetHumanAction())
		assert.Nil(t, dg.GetHumanAction().PlayedCards) // pass → nil
	})

	t.Run("success PlayerPlay does nothing when not human turn", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 7, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		dg.PlayerPlay([]int{0}) // advance to CPU turn
		if !dg.IsHumanTurn() && !dg.GetGameEndFlag() {
			ok := dg.PlayerPlay([]int{0})
			assert.False(t, ok)
		}
	})

	t.Run("success CpuPlay passes on table with unbeatable card", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		// Human has 2 cards: [2 (idx0), 3 (idx1)]  — play the 2 (strongest), keep 3
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 2, false)) // idx0 → play this
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 3, false)) // idx1 → kept
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 3, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 4, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 5, false))
		dg.PlayerPlay([]int{0}) // play 2 → human keeps [3], not finished
		// CPUs all pass (can't beat 2) → table clears → back to human
		dg.CpuPlay() // CPU 1 passes
		dg.CpuPlay() // CPU 2 passes
		dg.CpuPlay() // CPU 3 passes → checkPassClear triggers, table clears
		assert.Nil(t, dg.GetTableCards())
		assert.True(t, dg.IsHumanTurn())
	})

	t.Run("success CpuPlay does nothing on human turn", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		dg.CpuPlay() // does nothing on human turn
		assert.Nil(t, dg.GetTableCards())
		assert.True(t, dg.IsHumanTurn())
	})

	t.Run("success game ends when only 1 player remains", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		// 3 CPUs already finished, human has 1 card left
		players[1].SetIsFinished(true)
		players[1].SetRank(1)
		players[2].SetIsFinished(true)
		players[2].SetRank(2)
		players[3].SetIsFinished(true)
		players[3].SetRank(3)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))
		dg.PlayerPlay([]int{0}) // human plays last card → finishes → game ends
		assert.True(t, dg.GetGameEndFlag())
		// countFinished was 3 before human finished → rank = 4
		assert.Equal(t, 4, players[0].GetRank())
	})

	t.Run("success GetHumanAction after play", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 7, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		dg.PlayerPlay([]int{0}) // play 5
		action := dg.GetHumanAction()
		assert.NotNil(t, action)
		assert.Equal(t, 0, action.PlayerIdx)
		assert.Len(t, action.PlayedCards, 1)
		assert.Equal(t, 5, action.PlayedCards[0].GetValue())
	})

	t.Run("success GetCpuActions is nil at start", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		assert.Nil(t, dg.GetCpuActions())
	})

	t.Run("success pair play on clear table keeps table alive", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		// Human has 3 cards: pair of 5s + extra 3 (human doesn't finish after playing pair)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))  // idx0
		players[0].AddCard(entities.NewCard(entities.CardDesignHeart, 5, false))  // idx1
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))  // idx2 (kept)
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		ok := dg.PlayerPlay([]int{0, 1}) // play pair of 5s
		assert.True(t, ok)
		assert.Len(t, dg.GetTableCards(), 2)
		assert.Equal(t, 1, players[0].GetCardsSize()) // 1 card (3) remains
	})

	t.Run("success finishPlayer rank based on already-finished count", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		players[1].SetIsFinished(true)
		players[1].SetRank(1)
		players[2].SetIsFinished(true)
		players[2].SetRank(2)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		// 2 already finished → human finishes → gets rank 3
		dg.PlayerPlay([]int{0})
		assert.Equal(t, 3, players[0].GetRank())
	})
}
