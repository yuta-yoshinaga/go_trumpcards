package presenters_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"
	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/presenters"

	"github.com/stretchr/testify/assert"
)

func TestOldMaidCuiPresenter_Method(t *testing.T) {
	top := presenters.NewOldMaidCuiPresenter()

	makePlayers := func() []*entities.OldMaidPlayer {
		return []*entities.OldMaidPlayer{
			entities.NewOldMaidPlayer(true),
			entities.NewOldMaidPlayer(false),
			entities.NewOldMaidPlayer(false),
			entities.NewOldMaidPlayer(false),
		}
	}

	t.Run("success Output initial state no draw no game end", func(t *testing.T) {
		tc := entities.NewTrumpCards(1)
		players := makePlayers()
		om := entities.NewOldMaid(tc, players)
		// Manually set cards
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 1, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignClover, 2, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 3, false))
		players[2].SetIsFinished(true)
		players[3].AddCard(entities.NewCard(entities.CardDesignDiamond, 4, false))
		expected := "==========\nOld Maid (ババ抜き)\n==========\n" +
			"[You]: 2枚\n[0]SPADE 1  [1]CLOVER 2\n" +
			"CPU 1: 1枚\n" +
			"CPU 2: 上がり\n" +
			"CPU 3: 1枚\n" +
			"----------\n" +
			"手番: あなた → CPU 1から引きます\n" +
			"==========\n"
		assert.Equal(t, expected, top.Output(om))
	})

	t.Run("success Output game ended human loses", func(t *testing.T) {
		tc := entities.NewTrumpCards(1)
		players := makePlayers()
		om := entities.NewOldMaid(tc, players)
		// Setup: player 0 (human, turn=0) has JOKER + SPADE 5 + CLOVER 5
		// player 1 (CPU 1) has HEART 7 — exactly 1 card (deterministic draw at index 0)
		// players 2,3 finished
		players[0].AddCard(entities.NewCard(entities.CardDesignJoker, entities.CardValueJoker, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 7, false))
		players[2].SetIsFinished(true)
		players[3].SetIsFinished(true)
		// PlayerDraw(0): draws card at index 0 from player 1 (HEART 7)
		// player 0 gains HEART 7 → hand: JOKER, SPADE 5, CLOVER 5, HEART 7
		// DiscardPairs: SPADE 5 + CLOVER 5 pair → discarded → player 0: JOKER, HEART 7 (shuffled order)
		// player 1 has 0 cards → finished
		// checkGameEnd: active = {0} → gameEndFlag=true, loserIdx=0
		om.PlayerDraw(0)
		result := top.Output(om)
		// Card display order in player's hand is non-deterministic due to ShuffleCards
		assert.Contains(t, result, "[You]: 2枚")
		assert.Contains(t, result, "JOKER")
		assert.Contains(t, result, "HEART 7")
		assert.Contains(t, result, "CPU 1: 上がり")
		assert.Contains(t, result, "CPU 2: 上がり")
		assert.Contains(t, result, "CPU 3: 上がり")
		assert.Contains(t, result, "あなたがCPU 1から1枚引きました (HEART 7)。1組捨てました")
		assert.Contains(t, result, "ゲーム終了！ あなたの負け！")
	})

	t.Run("success Output game ended cpu loses", func(t *testing.T) {
		tc := entities.NewTrumpCards(1)
		players := makePlayers()
		om := entities.NewOldMaid(tc, players)
		// player 0 (human, turn=0): SPADE 3 (1 card)
		// player 1 (CPU 1): CLOVER 3 (1 card, next active) → deterministic draw (index always 0)
		// player 2 (CPU 2): JOKER (1 card) → will be the loser
		// player 3: finished
		// PlayerDraw(0): player 0 draws CLOVER 3, forms pair with SPADE 3 → both players 0 and 1 finish
		// active = {2} → gameEndFlag=true, loserIdx=2
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignClover, 3, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignJoker, entities.CardValueJoker, false))
		players[3].SetIsFinished(true)
		om.PlayerDraw(0)
		expected := "==========\nOld Maid (ババ抜き)\n==========\n" +
			"[You]: 上がり\n" +
			"CPU 1: 上がり\n" +
			"CPU 2: 1枚\n" +
			"CPU 3: 上がり\n" +
			"----------\n" +
			"あなたがCPU 1から1枚引きました (CLOVER 3)。1組捨てました\n" +
			"ゲーム終了！ CPU 2の負け！\n" +
			"==========\n"
		assert.Equal(t, expected, top.Output(om))
	})

	t.Run("success Output human zero cards not finished", func(t *testing.T) {
		tc := entities.NewTrumpCards(1)
		players := makePlayers()
		om := entities.NewOldMaid(tc, players)
		// human player has 0 cards but is not marked finished
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 3, false))
		players[2].SetIsFinished(true)
		players[3].SetIsFinished(true)
		result := top.Output(om)
		assert.Contains(t, result, "[You]: 0枚")
		assert.Contains(t, result, "CPU 1: 1枚")
		assert.Contains(t, result, "CPU 2: 上がり")
		assert.Contains(t, result, "CPU 3: 上がり")
	})

	t.Run("success Output cpu actions drawn card not revealed", func(t *testing.T) {
		tc := entities.NewTrumpCards(1)
		// Use all CPU players to enable CpuDraw
		cpuPlayers := []*entities.OldMaidPlayer{
			entities.NewOldMaidPlayer(false),
			entities.NewOldMaidPlayer(false),
			entities.NewOldMaidPlayer(false),
			entities.NewOldMaidPlayer(false),
		}
		om := entities.NewOldMaid(tc, cpuPlayers)
		// Player 0: JOKER
		// Player 1: SPADE 5 (1 card, deterministic draw at index 0)
		// Players 2, 3: finished
		cpuPlayers[0].AddCard(entities.NewCard(entities.CardDesignJoker, entities.CardValueJoker, false))
		cpuPlayers[1].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		cpuPlayers[2].SetIsFinished(true)
		cpuPlayers[3].SetIsFinished(true)

		// Player 0 draws SPADE 5, no pair → keeps it
		om.CpuDraw()

		result := top.Output(om)
		// CPU action should show who drew from whom but NOT which card was drawn
		assert.Contains(t, result, "[CPUの行動]")
		assert.Contains(t, result, "CPU 0がCPU 1から1枚引きました")
		assert.NotContains(t, result, "SPADE 5")
	})

	t.Run("success Output cpu actions with discard does not reveal drawn card", func(t *testing.T) {
		tc := entities.NewTrumpCards(1)
		cpuPlayers := []*entities.OldMaidPlayer{
			entities.NewOldMaidPlayer(false),
			entities.NewOldMaidPlayer(false),
			entities.NewOldMaidPlayer(false),
			entities.NewOldMaidPlayer(false),
		}
		om := entities.NewOldMaid(tc, cpuPlayers)
		// Player 0: SPADE 10
		// Player 1: CLOVER 10 (1 card, deterministic draw at index 0)
		// Players 2, 3: finished
		cpuPlayers[0].AddCard(entities.NewCard(entities.CardDesignSpade, 10, false))
		cpuPlayers[1].AddCard(entities.NewCard(entities.CardDesignClover, 10, false))
		cpuPlayers[2].SetIsFinished(true)
		cpuPlayers[3].SetIsFinished(true)

		// Player 0 draws CLOVER 10, discards pair SPADE 10 + CLOVER 10
		om.CpuDraw()

		result := top.Output(om)
		assert.Contains(t, result, "[CPUの行動]")
		assert.Contains(t, result, "CPU 0がCPU 1から1枚引きました。1組捨てました")
		// Drawn card must not appear even when a pair was discarded
		assert.NotContains(t, result, "CLOVER 10")
	})
}
