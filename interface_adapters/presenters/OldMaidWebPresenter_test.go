package presenters_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"
	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/presenters"

	"github.com/stretchr/testify/assert"
)

func TestOldMaidWebPresenter_Method(t *testing.T) {
	towp := presenters.NewOldMaidWebPresenter()

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
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 1, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignClover, 2, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 3, false))
		players[2].SetIsFinished(true)
		players[3].AddCard(entities.NewCard(entities.CardDesignDiamond, 4, false))
		expected := `{"players":[{"id":0,"isHuman":true,"isFinished":false,"cardCount":2,"cards":[{"design":"SPADE","value":1},{"design":"CLOVER","value":2}]},{"id":1,"isHuman":false,"isFinished":false,"cardCount":1,"cards":[]},{"id":2,"isHuman":false,"isFinished":true,"cardCount":0,"cards":[]},{"id":3,"isHuman":false,"isFinished":false,"cardCount":1,"cards":[]}],"currentTurn":0,"nextDrawTargetIdx":1,"gameEndFlag":false,"loserIdx":-1,"lastDrawPlayerIdx":-1,"lastDrawFromIdx":-1,"lastDrawCard":null,"lastDiscardedPairs":0,"lastDiscardedCards":[],"hasDrawn":false,"cpuActions":[],"humanAction":null,"message":""}`
		assert.Equal(t, expected, towp.Output(om))
	})

	t.Run("success Output game ended human loses", func(t *testing.T) {
		tc := entities.NewTrumpCards(1)
		players := makePlayers()
		om := entities.NewOldMaid(tc, players)
		// player 0 (human, turn=0): JOKER, SPADE 5, CLOVER 5
		// player 1 (CPU 1): HEART 7 — exactly 1 card (deterministic draw at index 0)
		// players 2,3 finished
		players[0].AddCard(entities.NewCard(entities.CardDesignJoker, entities.CardValueJoker, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 7, false))
		players[2].SetIsFinished(true)
		players[3].SetIsFinished(true)
		// PlayerDraw(0): draws HEART 7 (only card), discards SPADE5+CLOVER5 pair (1 pair)
		// player 0 left: JOKER, HEART 7; player 1 finished; game ends; loserIdx=0
		om.PlayerDraw(0)
		expected := `{"players":[{"id":0,"isHuman":true,"isFinished":false,"cardCount":2,"cards":[{"design":"JOKER","value":0},{"design":"HEART","value":7}]},{"id":1,"isHuman":false,"isFinished":true,"cardCount":0,"cards":[]},{"id":2,"isHuman":false,"isFinished":true,"cardCount":0,"cards":[]},{"id":3,"isHuman":false,"isFinished":true,"cardCount":0,"cards":[]}],"currentTurn":0,"nextDrawTargetIdx":0,"gameEndFlag":true,"loserIdx":0,"lastDrawPlayerIdx":0,"lastDrawFromIdx":1,"lastDrawCard":{"design":"HEART","value":7},"lastDiscardedPairs":1,"lastDiscardedCards":[{"design":"SPADE","value":5},{"design":"CLOVER","value":5}],"hasDrawn":true,"cpuActions":[],"humanAction":{"drawPlayerIdx":0,"drawFromIdx":1,"drawnCard":{"design":"HEART","value":7},"discardedPairs":1,"discardedCards":[{"design":"SPADE","value":5},{"design":"CLOVER","value":5}]},"message":"ゲーム終了！ あなたの負け！"}`
		assert.Equal(t, expected, towp.Output(om))
	})

	t.Run("success Output game not ended with draw and no discard", func(t *testing.T) {
		tc := entities.NewTrumpCards(1)
		players := makePlayers()
		om := entities.NewOldMaid(tc, players)
		// player 0 (human, turn=0): SPADE 5 (1 card)
		// player 1 (CPU 1): CLOVER 7 (1 card) — deterministic draw at index 0
		// player 2: HEART 9 (1 card, active)
		// player 3: finished
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignClover, 7, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 9, false))
		players[3].SetIsFinished(true)
		om.PlayerDraw(0)
		result := towp.Output(om)
		assert.Contains(t, result, `"hasDrawn":true`)
		assert.Contains(t, result, `"lastDrawPlayerIdx":0`)
		assert.Contains(t, result, `"lastDrawFromIdx":1`)
		assert.Contains(t, result, `"lastDiscardedPairs":0`)
		assert.Contains(t, result, `"gameEndFlag":false`)
		assert.Contains(t, result, `"lastDrawCard":{"design":"CLOVER","value":7}`)
		assert.Contains(t, result, `"cpuActions":[]`)
		assert.Contains(t, result, `"humanAction":{"drawPlayerIdx":0,"drawFromIdx":1`)
	})

	t.Run("success Output game ended cpu loses", func(t *testing.T) {
		tc := entities.NewTrumpCards(1)
		players := makePlayers()
		om := entities.NewOldMaid(tc, players)
		// player 0 (human, turn=0): SPADE 3 (1 card)
		// player 1 (CPU 1): CLOVER 3 (1 card, next active) → deterministic
		// player 2 (CPU 2): JOKER (1 card)
		// player 3: finished
		// PlayerDraw(0): player 0 draws CLOVER 3, forms pair → players 0,1 finish
		// active = {2} → gameEndFlag=true, loserIdx=2
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignClover, 3, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignJoker, entities.CardValueJoker, false))
		players[3].SetIsFinished(true)
		om.PlayerDraw(0)
		expected := `{"players":[{"id":0,"isHuman":true,"isFinished":true,"cardCount":0,"cards":[]},{"id":1,"isHuman":false,"isFinished":true,"cardCount":0,"cards":[]},{"id":2,"isHuman":false,"isFinished":false,"cardCount":1,"cards":[]},{"id":3,"isHuman":false,"isFinished":true,"cardCount":0,"cards":[]}],"currentTurn":0,"nextDrawTargetIdx":2,"gameEndFlag":true,"loserIdx":2,"lastDrawPlayerIdx":0,"lastDrawFromIdx":1,"lastDrawCard":{"design":"CLOVER","value":3},"lastDiscardedPairs":1,"lastDiscardedCards":[{"design":"SPADE","value":3},{"design":"CLOVER","value":3}],"hasDrawn":true,"cpuActions":[],"humanAction":{"drawPlayerIdx":0,"drawFromIdx":1,"drawnCard":{"design":"CLOVER","value":3},"discardedPairs":1,"discardedCards":[{"design":"SPADE","value":3},{"design":"CLOVER","value":3}]},"message":"ゲーム終了！ CPU 2の負け！"}`
		assert.Equal(t, expected, towp.Output(om))
	})

	t.Run("success Output with CpuActions and discarded cards", func(t *testing.T) {
		tc := entities.NewTrumpCards(1)
		// Use all CPU players for this test to enable CpuDraw
		players := []*entities.OldMaidPlayer{
			entities.NewOldMaidPlayer(false),
			entities.NewOldMaidPlayer(false),
			entities.NewOldMaidPlayer(false),
			entities.NewOldMaidPlayer(false),
		}
		om := entities.NewOldMaid(tc, players)
		// Player 0: SPADE 10
		// Player 1: CLOVER 10
		om.GetPlayer(0).AddCard(entities.NewCard(entities.CardDesignSpade, 10, false))
		om.GetPlayer(1).AddCard(entities.NewCard(entities.CardDesignClover, 10, false))
		om.GetPlayer(2).SetIsFinished(true)
		om.GetPlayer(3).SetIsFinished(true)

		// Turn is 0 (CPU). Call CpuDraw.
		// Player 0 draws CLOVER 10 from Player 1.
		// Player 0 discards SPADE 10 + CLOVER 10.
		om.CpuDraw()

		result := towp.Output(om)
		// drawnCard must be null in cpuActions to preserve game fairness
		assert.Contains(t, result, `"cpuActions":[{"drawPlayerIdx":0,"drawFromIdx":1,"drawnCard":null,"discardedPairs":1,"discardedCards":[{"design":"SPADE","value":10},{"design":"CLOVER","value":10}]}]`)
		// No human draw happened, so humanAction is null
		assert.Contains(t, result, `"humanAction":null`)
	})

	t.Run("success Output cpuActions drawnCard is nil when no pair discarded", func(t *testing.T) {
		tc := entities.NewTrumpCards(1)
		// Use all CPU players for this test to enable CpuDraw
		players := []*entities.OldMaidPlayer{
			entities.NewOldMaidPlayer(false),
			entities.NewOldMaidPlayer(false),
			entities.NewOldMaidPlayer(false),
			entities.NewOldMaidPlayer(false),
		}
		om := entities.NewOldMaid(tc, players)
		// Player 0: JOKER
		// Player 1: SPADE 5 (1 card, deterministic draw at index 0)
		// Players 2, 3: finished
		om.GetPlayer(0).AddCard(entities.NewCard(entities.CardDesignJoker, entities.CardValueJoker, false))
		om.GetPlayer(1).AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		om.GetPlayer(2).SetIsFinished(true)
		om.GetPlayer(3).SetIsFinished(true)

		// Player 0 draws SPADE 5, no pair → keeps it (drawnCard must not be revealed)
		om.CpuDraw()

		result := towp.Output(om)
		// drawnCard must be null regardless of whether a pair was discarded
		assert.Contains(t, result, `"drawnCard":null`)
		assert.NotContains(t, result, `"drawnCard":{"design`)
		assert.Contains(t, result, `"discardedPairs":0`)
	})
}
