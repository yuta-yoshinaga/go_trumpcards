package presenters_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"
	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/presenters"

	"github.com/stretchr/testify/assert"
)

func makeSevensPlayersForPresenter() []*entities.SevensPlayer {
	return []*entities.SevensPlayer{
		entities.NewSevensPlayer(true),
		entities.NewSevensPlayer(false),
		entities.NewSevensPlayer(false),
		entities.NewSevensPlayer(false),
	}
}

func TestSevensCuiPresenter_Method(t *testing.T) {
	tsp := presenters.NewSevensCuiPresenter()

	t.Run("success Output initial state", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		s := entities.NewSevens(tc, players)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 6, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignHeart, 8, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignClover, 9, false))

		result := tsp.Output(s)
		assert.Contains(t, result, "Sevens (7並べ)")
		assert.Contains(t, result, "[You]: 2枚")
		assert.Contains(t, result, "[0]SPADE 6")
		assert.Contains(t, result, "CPU 1: 1枚")
		assert.Contains(t, result, "ボード:")
		assert.Contains(t, result, "SPADE: 7〜7")
		assert.Contains(t, result, "手番: あなた")
	})

	t.Run("success Output shows pass count", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		s := entities.NewSevens(tc, players)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false)) // not playable
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 8, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))

		s.PlayerPlay(-1) // human passes

		result := tsp.Output(s)
		assert.Contains(t, result, "パス: 1/5")
		assert.Contains(t, result, "パスしました")
	})

	t.Run("success Output shows board state after play", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		s := entities.NewSevens(tc, players)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 6, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 8, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		s.PlayerPlay(0) // play 6♠ → minVal[Spade] = 6

		result := tsp.Output(s)
		assert.Contains(t, result, "SPADE: 6〜7")
	})

	t.Run("success Output game ended", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		s := entities.NewSevens(tc, players)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 8, false))
		players[1].SetIsFinished(true)
		players[1].SetRank(1)
		players[2].SetIsFinished(true)
		players[2].SetRank(2)
		players[3].SetIsFinished(true)
		players[3].SetRank(3)
		s.PlayerPlay(0)

		result := tsp.Output(s)
		assert.Contains(t, result, "ゲーム終了")
	})

	t.Run("success Output shows CPU actions", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		s := entities.NewSevens(tc, players)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 8, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 9, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false)) // not playable → pass
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		s.PlayerPlay(0) // human plays 8♠
		s.CpuPlay()    // CPU 1 passes

		result := tsp.Output(s)
		assert.Contains(t, result, "[CPUの行動]")
		assert.Contains(t, result, "パスしました")
	})

	t.Run("success Output finished player shows rank", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		s := entities.NewSevens(tc, players)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 8, false))
		players[1].SetIsFinished(true)
		players[1].SetRank(1)
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))

		result := tsp.Output(s)
		assert.Contains(t, result, "上がり/失格 (ランク: 1位)")
	})
}
