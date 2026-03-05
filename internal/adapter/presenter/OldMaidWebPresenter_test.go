package presenter_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
)

// setupOldMaidWebTest creates an OldMaid game with standard setup (player[0] SPADE 1, player[1] HEART 3, players[2,3] finished).
func setupOldMaidWebTest() (*domain.OldMaid, []*domain.OldMaidPlayer) {
	tc := domain.NewTrumpCards(1)
	players := []*domain.OldMaidPlayer{
		domain.NewOldMaidPlayer(true),
		domain.NewOldMaidPlayer(false),
		domain.NewOldMaidPlayer(false),
		domain.NewOldMaidPlayer(false),
	}
	om := domain.NewOldMaid(tc, players)
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
	players[2].SetIsFinished(true)
	players[3].SetIsFinished(true)
	return om, players
}

func TestOldMaidWebPresenter_Method(t *testing.T) {
	towp := presenter.NewOldMaidWebPresenter()

	makePlayers := func() []*domain.OldMaidPlayer {
		return []*domain.OldMaidPlayer{
			domain.NewOldMaidPlayer(true),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
		}
	}

	t.Run("success Output initial state no draw no game end", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makePlayers()
		om := domain.NewOldMaid(tc, players)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 2, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		players[2].SetIsFinished(true)
		players[3].AddCard(domain.NewCard(domain.CardDesignDiamond, 4, false))
		expected := `{"players":[{"id":0,"isHuman":true,"isFinished":false,"cardCount":2,"cards":[{"design":"SPADE","value":1},{"design":"CLOVER","value":2}]},{"id":1,"isHuman":false,"isFinished":false,"cardCount":1,"cards":[]},{"id":2,"isHuman":false,"isFinished":true,"cardCount":0,"cards":[]},{"id":3,"isHuman":false,"isFinished":false,"cardCount":1,"cards":[]}],"currentTurn":0,"nextDrawTargetIdx":1,"gameEndFlag":false,"loserIdx":-1,"lastDrawPlayerIdx":-1,"lastDrawFromIdx":-1,"lastDrawCard":null,"lastDiscardedPairs":0,"lastDiscardedCards":[],"hasDrawn":false,"cpuActions":[],"humanAction":null,"cpuHighlightedCardIdx":-1,"removedCard":null,"mode":0,"message":""}`
		assert.Equal(t, expected, towp.Output(om, nil))
	})

	t.Run("success Output game ended human loses", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makePlayers()
		om := domain.NewOldMaid(tc, players)
		// player 0 (human, turn=0): JOKER, SPADE 5, CLOVER 5
		// player 1 (CPU 1): HEART 7 — exactly 1 card (deterministic draw at index 0)
		// players 2,3 finished
		players[0].AddCard(domain.NewCard(domain.CardDesignJoker, domain.CardValueJoker, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		players[2].SetIsFinished(true)
		players[3].SetIsFinished(true)
		// PlayerDraw(0): draws HEART 7 (only card), discards SPADE5+CLOVER5 pair (1 pair)
		// player 0 left: JOKER, HEART 7 (shuffled order); player 1 finished; game ends; loserIdx=0
		_ = om.PlayerDraw(0)
		result := towp.Output(om, nil)
		// Card order in player 0's hand is non-deterministic due to ShuffleCards; verify presence
		assert.Contains(t, result, `"cardCount":2`)
		assert.Contains(t, result, `{"design":"JOKER","value":0}`)
		assert.Contains(t, result, `{"design":"HEART","value":7}`)
		assert.Contains(t, result, `"gameEndFlag":true`)
		assert.Contains(t, result, `"loserIdx":0`)
		assert.Contains(t, result, `"lastDrawPlayerIdx":0`)
		assert.Contains(t, result, `"lastDrawFromIdx":1`)
		assert.Contains(t, result, `"lastDrawCard":{"design":"HEART","value":7}`)
		assert.Contains(t, result, `"lastDiscardedPairs":1`)
		assert.Contains(t, result, `"hasDrawn":true`)
		assert.Contains(t, result, `"humanAction":{"drawPlayerIdx":0,"drawFromIdx":1,"drawnCard":{"design":"HEART","value":7},"discardedPairs":1`)
		assert.Contains(t, result, `"message":"ゲーム終了！ あなたの負け！"`)
		assert.Contains(t, result, `"messageCode":"oldmaid.result.humanLose"`)
	})

	t.Run("success Output game not ended with draw and no discard", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makePlayers()
		om := domain.NewOldMaid(tc, players)
		// player 0 (human, turn=0): SPADE 5 (1 card)
		// player 1 (CPU 1): CLOVER 7 (1 card) — deterministic draw at index 0
		// player 2: HEART 9 (1 card, active)
		// player 3: finished
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 7, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
		players[3].SetIsFinished(true)
		_ = om.PlayerDraw(0)
		result := towp.Output(om, nil)
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
		tc := domain.NewTrumpCards(1)
		players := makePlayers()
		om := domain.NewOldMaid(tc, players)
		// player 0 (human, turn=0): SPADE 3 (1 card)
		// player 1 (CPU 1): CLOVER 3 (1 card, next active) → deterministic
		// player 2 (CPU 2): JOKER (1 card)
		// player 3: finished
		// PlayerDraw(0): player 0 draws CLOVER 3, forms pair → players 0,1 finish
		// active = {2} → gameEndFlag=true, loserIdx=2
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 3, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignJoker, domain.CardValueJoker, false))
		players[3].SetIsFinished(true)
		_ = om.PlayerDraw(0)
		result := towp.Output(om, nil)
		// discardedCards order is non-deterministic due to ShuffleCards before DiscardPairs
		assert.Contains(t, result, `"isFinished":true,"cardCount":0,"cards":[]`)
		assert.Contains(t, result, `"gameEndFlag":true`)
		assert.Contains(t, result, `"loserIdx":2`)
		assert.Contains(t, result, `"lastDrawPlayerIdx":0`)
		assert.Contains(t, result, `"lastDrawFromIdx":1`)
		assert.Contains(t, result, `"lastDrawCard":{"design":"CLOVER","value":3}`)
		assert.Contains(t, result, `"lastDiscardedPairs":1`)
		assert.Contains(t, result, `{"design":"SPADE","value":3}`)
		assert.Contains(t, result, `{"design":"CLOVER","value":3}`)
		assert.Contains(t, result, `"hasDrawn":true`)
		assert.Contains(t, result, `"message":"ゲーム終了！ CPU 2の負け！"`)
		assert.Contains(t, result, `"messageCode":"oldmaid.result.cpuLose"`)
		assert.Contains(t, result, `"messageParams":{"cpuId":"2"}`)
	})

	t.Run("success Output with CpuActions and discarded cards", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		// Use all CPU players for this test to enable CpuDraw
		players := []*domain.OldMaidPlayer{
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
		}
		om := domain.NewOldMaid(tc, players)
		// Player 0: SPADE 10
		// Player 1: CLOVER 10
		om.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		om.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
		om.GetPlayer(2).SetIsFinished(true)
		om.GetPlayer(3).SetIsFinished(true)

		// Turn is 0 (CPU). Call CpuDraw.
		// Player 0 draws CLOVER 10 from Player 1.
		// Player 0 discards SPADE 10 + CLOVER 10.
		_ = om.CpuDraw()

		result := towp.Output(om, nil)
		// drawnCard must be null in cpuActions to preserve game fairness
		// discardedCards order is non-deterministic due to ShuffleCards before DiscardPairs
		assert.Contains(t, result, `"cpuActions":[{"drawPlayerIdx":0,"drawFromIdx":1,"drawnCard":null,"discardedPairs":1,"discardedCards":[`)
		assert.Contains(t, result, `{"design":"SPADE","value":10}`)
		assert.Contains(t, result, `{"design":"CLOVER","value":10}`)
		// No human draw happened, so humanAction is null
		assert.Contains(t, result, `"humanAction":null`)
	})

	t.Run("success Output game ended human loses at non-zero index", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		// Human is at index 2 (simulates shuffled player order)
		players := []*domain.OldMaidPlayer{
			domain.NewOldMaidPlayer(false), // CPU at index 0
			domain.NewOldMaidPlayer(false), // CPU at index 1
			domain.NewOldMaidPlayer(true),  // Human at index 2
			domain.NewOldMaidPlayer(false), // CPU at index 3
		}
		om := domain.NewOldMaid(tc, players)
		// Player 0 (CPU, turn=0): SPADE 3
		// Player 1 (CPU): CLOVER 3 (1 card → deterministic draw)
		// Player 2 (Human): JOKER → will be the last remaining (loser)
		// Player 3: finished
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 3, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignJoker, domain.CardValueJoker, false))
		players[3].SetIsFinished(true)
		// CpuDraw: player 0 draws CLOVER 3, forms pair, both 0 and 1 finish → loserIdx=2 (human)
		_ = om.CpuDraw()
		result := towp.Output(om, nil)
		assert.Contains(t, result, `"message":"ゲーム終了！ あなたの負け！"`)
		assert.Contains(t, result, `"messageCode":"oldmaid.result.humanLose"`)
	})

	t.Run("success Output game ended cpu loses at index 0", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		// Human is at index 2 (simulates shuffled player order)
		players := []*domain.OldMaidPlayer{
			domain.NewOldMaidPlayer(false), // CPU at index 0
			domain.NewOldMaidPlayer(false), // CPU at index 1
			domain.NewOldMaidPlayer(true),  // Human at index 2
			domain.NewOldMaidPlayer(false), // CPU at index 3
		}
		om := domain.NewOldMaid(tc, players)
		// Player 0 (CPU, turn=0): JOKER → will be the last remaining (loser)
		// Player 1 (CPU): SPADE 3 (1 card → deterministic draw)
		// Player 2: Human, finished
		// Player 3: finished
		players[0].AddCard(domain.NewCard(domain.CardDesignJoker, domain.CardValueJoker, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[2].SetIsFinished(true)
		players[3].SetIsFinished(true)
		// CpuDraw: player 0 draws SPADE 3 from player 1, no pair → player 1 finishes (0 cards)
		// active = {player 0} → game ends, loserIdx=0 (CPU)
		_ = om.CpuDraw()
		result := towp.Output(om, nil)
		assert.Contains(t, result, `"message":"ゲーム終了！ CPU 0の負け！"`)
		assert.Contains(t, result, `"messageCode":"oldmaid.result.cpuLose"`)
		assert.Contains(t, result, `"messageParams":{"cpuId":"0"}`)
	})

	t.Run("success Output displays error message", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makePlayers()
		om := domain.NewOldMaid(tc, players)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		players[2].SetIsFinished(true)
		players[3].SetIsFinished(true)
		result := towp.Output(om, domain.ErrNotHumanTurn)
		assert.Contains(t, result, `"message":"not human player's turn"`)
	})

	t.Run("success Output cpuActions drawnCard is nil when no pair discarded", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		// Use all CPU players for this test to enable CpuDraw
		players := []*domain.OldMaidPlayer{
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
		}
		om := domain.NewOldMaid(tc, players)
		// Player 0: JOKER
		// Player 1: SPADE 5 (1 card, deterministic draw at index 0)
		// Players 2, 3: finished
		om.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignJoker, domain.CardValueJoker, false))
		om.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		om.GetPlayer(2).SetIsFinished(true)
		om.GetPlayer(3).SetIsFinished(true)

		// Player 0 draws SPADE 5, no pair → keeps it (drawnCard must not be revealed)
		_ = om.CpuDraw()

		result := towp.Output(om, nil)
		// drawnCard must be null regardless of whether a pair was discarded
		assert.Contains(t, result, `"drawnCard":null`)
		assert.NotContains(t, result, `"drawnCard":{"design`)
		assert.Contains(t, result, `"discardedPairs":0`)
	})

	t.Run("success Output lastDrawPlayer nil hides draw card", func(t *testing.T) {
		om, _ := setupOldMaidWebTest()
		// Simulate draw having happened with invalid player idx → GetPlayer returns nil
		om.SetHasDrawn(true)
		om.SetLastDrawPlayerIdx(-1)
		result := towp.Output(om, nil)
		// lastDrawCard should be null since lastDrawPlayer is nil
		assert.Contains(t, result, `"lastDrawCard":null`)
	})

	t.Run("success Output getCardObj nil card via humanAction", func(t *testing.T) {
		om, _ := setupOldMaidWebTest()
		// HumanAction with nil DrawnCard → exercises getCardObj(nil)
		om.SetHumanAction(&domain.OldMaidCpuAction{
			DrawPlayerIdx:  0,
			DrawFromIdx:    1,
			DrawnCard:      nil,
			DiscardedPairs: 0,
			DiscardedCards: []*domain.Card{},
		})
		result := towp.Output(om, nil)
		assert.Contains(t, result, `"humanAction":{`)
		assert.Contains(t, result, `"drawnCard":null`)
	})

	t.Run("success Output lastDrawPlayer is CPU hides draw card", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		cpuPlayers := []*domain.OldMaidPlayer{
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
		}
		om := domain.NewOldMaid(tc, cpuPlayers)
		// Player 0: JOKER
		// Player 1: SPADE 5
		cpuPlayers[0].AddCard(domain.NewCard(domain.CardDesignJoker, domain.CardValueJoker, false))
		cpuPlayers[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		cpuPlayers[2].SetIsFinished(true)
		cpuPlayers[3].SetIsFinished(true)
		_ = om.CpuDraw()
		result := towp.Output(om, nil)
		// lastDrawPlayer is CPU → lastDrawCard should be null
		assert.Contains(t, result, `"lastDrawCard":null`)
	})
}

func TestOldMaidWebPresenter_CpuHighlightedCardIdx(t *testing.T) {
	towp := presenter.NewOldMaidWebPresenter()
	tc := domain.NewTrumpCards(1)
	players := []*domain.OldMaidPlayer{
		domain.NewOldMaidPlayer(true),
		domain.NewOldMaidPlayer(false),
		domain.NewOldMaidPlayer(false),
		domain.NewOldMaidPlayer(false),
	}
	om := domain.NewOldMaid(tc, players)
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
	players[2].SetIsFinished(true)
	players[3].SetIsFinished(true)

	om.SetCpuHighlightedCardIdx(2)
	result := towp.Output(om, nil)
	assert.Contains(t, result, `"cpuHighlightedCardIdx":2`)
}

func TestOldMaidWebPresenter_Mode_Normal(t *testing.T) {
	towp := presenter.NewOldMaidWebPresenter()
	tc := domain.NewTrumpCards(1)
	players := []*domain.OldMaidPlayer{
		domain.NewOldMaidPlayer(true),
		domain.NewOldMaidPlayer(false),
		domain.NewOldMaidPlayer(false),
		domain.NewOldMaidPlayer(false),
	}
	om := domain.NewOldMaid(tc, players)
	// Default config is Normal (mode=0)
	result := towp.Output(om, nil)
	assert.Contains(t, result, `"mode":0`)
	assert.Contains(t, result, `"removedCard":null`)
}

func TestOldMaidWebPresenter_JijiNuki_GameEnd_RevealsRemovedCard(t *testing.T) {
	towp := presenter.NewOldMaidWebPresenter()
	tc := domain.NewTrumpCards(1)
	players := []*domain.OldMaidPlayer{
		domain.NewOldMaidPlayer(true),
		domain.NewOldMaidPlayer(false),
		domain.NewOldMaidPlayer(false),
		domain.NewOldMaidPlayer(false),
	}
	om := domain.NewOldMaid(tc, players)
	om.SetConfig(domain.OldMaidConfig{Mode: domain.OldMaidModeJijiNuki})
	// Simulate game end
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
	players[2].SetIsFinished(true)
	players[3].SetIsFinished(true)
	_ = om.PlayerDraw(0) // ends game

	// Manually set removedCard (Reset normally sets it but we're bypassing it)
	// We need to set it via a fresh reset with JijiNuki config
	om2 := domain.NewOldMaid(domain.NewTrumpCards(0), []*domain.OldMaidPlayer{
		domain.NewOldMaidPlayer(true),
		domain.NewOldMaidPlayer(false),
		domain.NewOldMaidPlayer(false),
		domain.NewOldMaidPlayer(false),
	})
	om2.SetConfig(domain.OldMaidConfig{Mode: domain.OldMaidModeJijiNuki})
	om2.Reset()
	// After reset, game likely not ended; force game end for test
	om2.SetGameEndFlag(true)
	result := towp.Output(om2, nil)
	assert.Contains(t, result, `"mode":1`)
	// removedCard should be revealed
	assert.NotContains(t, result, `"removedCard":null`)
}

func TestOldMaidWebPresenter_JijiNuki_GameNotEnd_NoRemovedCard(t *testing.T) {
	towp := presenter.NewOldMaidWebPresenter()
	tc := domain.NewTrumpCards(1)
	players := []*domain.OldMaidPlayer{
		domain.NewOldMaidPlayer(true),
		domain.NewOldMaidPlayer(false),
		domain.NewOldMaidPlayer(false),
		domain.NewOldMaidPlayer(false),
	}
	om := domain.NewOldMaid(tc, players)
	om.SetConfig(domain.OldMaidConfig{Mode: domain.OldMaidModeJijiNuki})
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
	players[2].SetIsFinished(true)
	players[3].SetIsFinished(true)
	result := towp.Output(om, nil)
	assert.Contains(t, result, `"mode":1`)
	// Game not ended → removedCard not revealed
	assert.Contains(t, result, `"removedCard":null`)
}
