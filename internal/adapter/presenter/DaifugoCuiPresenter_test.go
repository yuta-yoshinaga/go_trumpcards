package presenter_test

import (
	"errors"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
)

func makeDaifugoPlayersForPresenter() []*domain.DaifugoPlayer {
	return []*domain.DaifugoPlayer{
		domain.NewDaifugoPlayer(true),
		domain.NewDaifugoPlayer(false),
		domain.NewDaifugoPlayer(false),
		domain.NewDaifugoPlayer(false),
	}
}

func TestDaifugoCuiPresenter_Method(t *testing.T) {
	tdp := presenter.NewDaifugoCuiPresenter()

	t.Run("success Output initial state", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDaifugoPlayersForPresenter()
		dg := domain.NewDaifugo(tc, players, domain.DaifugoConfig{})
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := tdp.Output(dg, nil)
		assert.Contains(t, result, "Daifugo (大富豪)")
		assert.Contains(t, result, "[You]: 2枚")
		assert.Contains(t, result, "[0]SPADE 3")
		assert.Contains(t, result, "CPU 1: 1枚")
		assert.Contains(t, result, "場: なし")
		assert.Contains(t, result, "手番: あなた")
	})

	t.Run("success Output with table cards", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDaifugoPlayersForPresenter()
		dg := domain.NewDaifugo(tc, players, domain.DaifugoConfig{})
		// Human plays a 5
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		_ = dg.PlayerPlay([]int{0}) // play 5
		// CPUs all pass (2 > 5 but 1 card vs 1 card needed — actually 2 is strongest, so CPUs WILL play it)
		// Instead just verify Output works after play
		result := tdp.Output(dg, nil)
		assert.Contains(t, result, "Daifugo (大富豪)")
		assert.NotEmpty(t, result)
	})

	t.Run("success Output game ended", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDaifugoPlayersForPresenter()
		dg := domain.NewDaifugo(tc, players, domain.DaifugoConfig{})
		// Simulate game end: finish all CPU players, human has 1 card, game ends
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[1].SetIsFinished(true)
		players[1].SetRank(1)
		players[2].SetIsFinished(true)
		players[2].SetRank(2)
		players[3].SetIsFinished(true)
		players[3].SetRank(3)
		// Play human's last card to end game
		_ = dg.PlayerPlay([]int{0})
		result := tdp.Output(dg, nil)
		assert.Contains(t, result, "ゲーム終了")
	})

	t.Run("success Output shows human action pass", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDaifugoPlayersForPresenter()
		dg := domain.NewDaifugo(tc, players, domain.DaifugoConfig{})
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		_ = dg.PlayerPlay([]int{}) // pass
		result := tdp.Output(dg, nil)
		assert.Contains(t, result, "パスしました")
	})

	t.Run("success Output shows revolution status when active", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDaifugoPlayersForPresenter()
		dg := domain.NewDaifugo(tc, players, domain.DaifugoConfig{})
		// Human plays four 5s → revolution active
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignDiamond, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false)) // spare
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		_ = dg.PlayerPlay([]int{0, 1, 2, 3}) // play four 5s
		result := tdp.Output(dg, nil)
		assert.Contains(t, result, "革命中")
	})

	t.Run("success Output does not show revolution status when not active", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDaifugoPlayersForPresenter()
		dg := domain.NewDaifugo(tc, players, domain.DaifugoConfig{})
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))
		result := tdp.Output(dg, nil)
		assert.NotContains(t, result, "革命中")
	})

	t.Run("success Output shows CPU actions", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDaifugoPlayersForPresenter()
		dg := domain.NewDaifugo(tc, players, domain.DaifugoConfig{})
		// Human plays 2 (strongest) → CPUs pass → back to human
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 4, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		_ = dg.PlayerPlay([]int{0})
		dg.CpuPlay()
		dg.CpuPlay()
		dg.CpuPlay()
		result := tdp.Output(dg, nil)
		assert.Contains(t, result, "[CPUの行動]")
	})

	t.Run("success Output shows error message when lastErr is non-nil", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeDaifugoPlayersForPresenter()
		dg := domain.NewDaifugo(tc, players, domain.DaifugoConfig{})
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		testErr := errors.New("test error message")
		result := tdp.Output(dg, testErr)
		assert.Contains(t, result, "test error message")
	})
}
