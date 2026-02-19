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
		// player 1 (CPU 1) has HEART 7 — exactly 1 card (deterministic draw)
		// players 2,3 finished
		players[0].AddCard(entities.NewCard(entities.CardDesignJoker, entities.CardValueJoker, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 7, false))
		players[2].SetIsFinished(true)
		players[3].SetIsFinished(true)
		// PlayerDraw: player 0 draws from player 1 (1 card → index always 0, deterministic)
		// player 0 gains HEART 7 → hand: JOKER, SPADE 5, CLOVER 5, HEART 7
		// DiscardPairs: SPADE 5 + CLOVER 5 pair → discarded → player 0: JOKER, HEART 7
		// player 1 has 0 cards → finished
		// checkGameEnd: active = {0} → gameEndFlag=true, loserIdx=0
		om.PlayerDraw()
		expected := "==========\nOld Maid (ババ抜き)\n==========\n" +
			"[You]: 2枚\n[0]JOKER  [1]HEART 7\n" +
			"CPU 1: 上がり\n" +
			"CPU 2: 上がり\n" +
			"CPU 3: 上がり\n" +
			"----------\n" +
			"あなたがCPU 1から1枚引きました。1組捨てました\n" +
			"ゲーム終了！ あなたの負け！\n" +
			"==========\n"
		assert.Equal(t, expected, top.Output(om))
	})

	t.Run("success Output game ended cpu loses", func(t *testing.T) {
		tc := entities.NewTrumpCards(1)
		players := makePlayers()
		om := entities.NewOldMaid(tc, players)
		// player 0 (human, turn=0): SPADE 3 (1 card)
		// player 1 (CPU 1): CLOVER 3 (1 card, next active) → deterministic draw (index always 0)
		// player 2 (CPU 2): JOKER (1 card) → will be the loser
		// player 3: finished
		// PlayerDraw: player 0 draws CLOVER 3, forms pair with SPADE 3 → both players 0 and 1 finish
		// active = {2} → gameEndFlag=true, loserIdx=2
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignClover, 3, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignJoker, entities.CardValueJoker, false))
		players[3].SetIsFinished(true)
		om.PlayerDraw()
		expected := "==========\nOld Maid (ババ抜き)\n==========\n" +
			"[You]: 上がり\n" +
			"CPU 1: 上がり\n" +
			"CPU 2: 1枚\n" +
			"CPU 3: 上がり\n" +
			"----------\n" +
			"あなたがCPU 1から1枚引きました。1組捨てました\n" +
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
}
