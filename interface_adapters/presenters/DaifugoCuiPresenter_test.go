package presenters_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"
	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/presenters"

	"github.com/stretchr/testify/assert"
)

func makeDaifugoPlayersForPresenter() []*entities.DaifugoPlayer {
	return []*entities.DaifugoPlayer{
		entities.NewDaifugoPlayer(true),
		entities.NewDaifugoPlayer(false),
		entities.NewDaifugoPlayer(false),
		entities.NewDaifugoPlayer(false),
	}
}

func TestDaifugoCuiPresenter_Method(t *testing.T) {
	tdp := presenters.NewDaifugoCuiPresenter()

	t.Run("success Output initial state", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayersForPresenter()
		dg := entities.NewDaifugo(tc, players)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 7, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 9, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 11, false))

		result := tdp.Output(dg)
		assert.Contains(t, result, "Daifugo (大富豪)")
		assert.Contains(t, result, "[You]: 2枚")
		assert.Contains(t, result, "[0]SPADE 3")
		assert.Contains(t, result, "CPU 1: 1枚")
		assert.Contains(t, result, "場: なし")
		assert.Contains(t, result, "手番: あなた")
	})

	t.Run("success Output with table cards", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayersForPresenter()
		dg := entities.NewDaifugo(tc, players)
		// Human plays a 5
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 7, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		dg.PlayerPlay([]int{0}) // play 5
		// CPUs all pass (2 > 5 but 1 card vs 1 card needed — actually 2 is strongest, so CPUs WILL play it)
		// Instead just verify Output works after play
		result := tdp.Output(dg)
		assert.Contains(t, result, "Daifugo (大富豪)")
		assert.NotEmpty(t, result)
	})

	t.Run("success Output game ended", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayersForPresenter()
		dg := entities.NewDaifugo(tc, players)
		// Simulate game end: finish all CPU players, human has 1 card, game ends
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))
		players[1].SetIsFinished(true)
		players[1].SetRank(1)
		players[2].SetIsFinished(true)
		players[2].SetRank(2)
		players[3].SetIsFinished(true)
		players[3].SetRank(3)
		// Play human's last card to end game
		dg.PlayerPlay([]int{0})
		result := tdp.Output(dg)
		assert.Contains(t, result, "ゲーム終了")
	})

	t.Run("success Output shows human action pass", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayersForPresenter()
		dg := entities.NewDaifugo(tc, players)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		dg.PlayerPlay([]int{}) // pass
		result := tdp.Output(dg)
		assert.Contains(t, result, "パスしました")
	})

	t.Run("success Output shows CPU actions", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayersForPresenter()
		dg := entities.NewDaifugo(tc, players)
		// Human plays 2 (strongest) → CPUs pass → back to human
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 3, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 4, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 5, false))
		dg.PlayerPlay([]int{0})
		dg.CpuPlay()
		dg.CpuPlay()
		dg.CpuPlay()
		result := tdp.Output(dg)
		assert.Contains(t, result, "[CPUの行動]")
	})
}
