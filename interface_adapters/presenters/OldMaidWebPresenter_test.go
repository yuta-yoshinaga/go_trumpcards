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
		expected := `{"players":[{"id":0,"isHuman":true,"isFinished":false,"cardCount":2,"cards":[{"design":"SPADE","value":1},{"design":"CLOVER","value":2}]},{"id":1,"isHuman":false,"isFinished":false,"cardCount":1,"cards":[]},{"id":2,"isHuman":false,"isFinished":true,"cardCount":0,"cards":[]},{"id":3,"isHuman":false,"isFinished":false,"cardCount":1,"cards":[]}],"currentTurn":0,"nextDrawTargetIdx":1,"gameEndFlag":false,"loserIdx":-1,"lastDrawPlayerIdx":-1,"lastDrawFromIdx":-1,"lastDiscardedPairs":0,"hasDrawn":false,"message":""}`
		assert.Equal(t, expected, towp.Output(om))
	})

	t.Run("success Output game ended human loses", func(t *testing.T) {
		tc := entities.NewTrumpCards(1)
		players := makePlayers()
		om := entities.NewOldMaid(tc, players)
		// player 0 (human, turn=0): JOKER, SPADE 5, CLOVER 5
		// player 1 (CPU 1): HEART 7 — exactly 1 card (deterministic draw)
		// players 2,3 finished
		players[0].AddCard(entities.NewCard(entities.CardDesignJoker, entities.CardValueJoker, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 7, false))
		players[2].SetIsFinished(true)
		players[3].SetIsFinished(true)
		// PlayerDraw: draws HEART 7 (only card), discards SPADE5+CLOVER5 pair (1 pair)
		// player 0 left: JOKER, HEART 7; player 1 finished; game ends; loserIdx=0
		om.PlayerDraw()
		// getNextActivePlayer(0): only player 0 is active, so it returns 0 (itself)
		expected := `{"players":[{"id":0,"isHuman":true,"isFinished":false,"cardCount":2,"cards":[{"design":"JOKER","value":0},{"design":"HEART","value":7}]},{"id":1,"isHuman":false,"isFinished":true,"cardCount":0,"cards":[]},{"id":2,"isHuman":false,"isFinished":true,"cardCount":0,"cards":[]},{"id":3,"isHuman":false,"isFinished":true,"cardCount":0,"cards":[]}],"currentTurn":0,"nextDrawTargetIdx":0,"gameEndFlag":true,"loserIdx":0,"lastDrawPlayerIdx":0,"lastDrawFromIdx":1,"lastDiscardedPairs":1,"hasDrawn":true,"message":"ゲーム終了！ あなたの負け！"}`
		assert.Equal(t, expected, towp.Output(om))
	})

	t.Run("success Output game not ended with draw and no discard", func(t *testing.T) {
		tc := entities.NewTrumpCards(1)
		players := makePlayers()
		om := entities.NewOldMaid(tc, players)
		// player 0 (human, turn=0): SPADE 5 (1 card)
		// player 1 (CPU 1): CLOVER 7, DIAMOND 8 (2 cards) — non-deterministic draw, skip
		// Instead: player 1 has exactly 1 card CLOVER 7 (no pair with SPADE 5)
		// player 2: HEART 9 (1 card, active)
		// player 3: finished
		// PlayerDraw: player 0 draws from player 1 (1 card → index 0, deterministic)
		//   player 0 gets CLOVER 7 → SPADE 5, CLOVER 7 → no pair → 0 discarded
		//   player 1 has 0 cards → finished
		//   checkGameEnd: active = {0, 2} → 2 active → game not ended
		//   advanceTurn: nextActivePlayer(0) = 2 (player 1 is finished, player 2 is active)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignClover, 7, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 9, false))
		players[3].SetIsFinished(true)
		om.PlayerDraw()
		// After draw: player 0 has SPADE 5, CLOVER 7; player 1 finished; player 2 has HEART 9
		// currentTurn = 2 (advanced past finished player 1)
		// hasDrawn=true, lastDrawPlayerIdx=0, lastDrawFromIdx=1, lastDiscardedPairs=0
		result := towp.Output(om)
		assert.Contains(t, result, `"hasDrawn":true`)
		assert.Contains(t, result, `"lastDrawPlayerIdx":0`)
		assert.Contains(t, result, `"lastDrawFromIdx":1`)
		assert.Contains(t, result, `"lastDiscardedPairs":0`)
		assert.Contains(t, result, `"gameEndFlag":false`)
	})

	t.Run("success Output game ended cpu loses", func(t *testing.T) {
		tc := entities.NewTrumpCards(1)
		players := makePlayers()
		om := entities.NewOldMaid(tc, players)
		// Setup for CPU 2 losing:
		// player 0 (human, turn=0): SPADE 3 (1 card)
		// player 1 (CPU 1): finished
		// player 2 (CPU 2): JOKER, CLOVER 3 (2 cards)  → nextActive from 0 is 2
		// player 3: finished
		// PlayerDraw: player 0 draws from player 2 (next active after 0, skipping finished 1)
		//   player 2 has 2 cards → non-deterministic!
		// Use 1 card for player 2: JOKER (1 card, deterministic)
		// player 0 draws JOKER → SPADE 3, JOKER → no pair → 0 discarded
		// player 2 has 0 cards → finished
		// checkGameEnd: active = {0} (players 1,2,3 all finished) → gameEndFlag=true, loserIdx=0
		// This again gives human as loser. Let's try differently.
		//
		// The only way to get CPU losing deterministically via public API starting from turn=0:
		// player 0 draws, then finishes, then CPU with joker is last → but that requires player 0 to
		// finish after drawing (pair is created). Then only CPU with joker remains → CPU loses.
		//
		// Setup:
		// player 0 (human, turn=0): SPADE 3, CLOVER 3 (pair available after drawing matching card?)
		// No - the pair in player 0's hand would be discarded by DiscardPairs call.
		// player 0 needs to GAIN a card that pairs with something in hand.
		// player 0: SPADE 3 (1 card) - will pair with what they draw
		// player 1 (CPU 1): CLOVER 3 (1 card, next active after 0) - exactly 1 card (deterministic)
		// player 2 (CPU 2): JOKER (1 card)
		// player 3: finished
		// PlayerDraw: player 0 draws CLOVER 3 from player 1 (1 card → deterministic)
		//   player 0 now has SPADE 3, CLOVER 3 → DiscardPairs: 1 pair → player 0 has 0 cards → finished
		//   player 1 has 0 cards → finished
		//   checkGameEnd: active = {2} → gameEndFlag=true, loserIdx=2
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignClover, 3, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignJoker, entities.CardValueJoker, false))
		players[3].SetIsFinished(true)
		om.PlayerDraw()
		// getNextActivePlayer(0): player 0 finished, player 1 finished, player 2 not finished → returns 2
		expected := `{"players":[{"id":0,"isHuman":true,"isFinished":true,"cardCount":0,"cards":[]},{"id":1,"isHuman":false,"isFinished":true,"cardCount":0,"cards":[]},{"id":2,"isHuman":false,"isFinished":false,"cardCount":1,"cards":[]},{"id":3,"isHuman":false,"isFinished":true,"cardCount":0,"cards":[]}],"currentTurn":0,"nextDrawTargetIdx":2,"gameEndFlag":true,"loserIdx":2,"lastDrawPlayerIdx":0,"lastDrawFromIdx":1,"lastDiscardedPairs":1,"hasDrawn":true,"message":"ゲーム終了！ CPU 2の負け！"}`
		assert.Equal(t, expected, towp.Output(om))
	})
}
