package domain_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestOldMaid_Method(t *testing.T) {
	makePlayers := func() []*domain.OldMaidPlayer {
		return []*domain.OldMaidPlayer{
			domain.NewOldMaidPlayer(true),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
		}
	}

	t.Run("success NewOldMaid", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makePlayers()
		om := domain.NewOldMaid(tc, players)
		assert.NotNil(t, om)
		assert.Equal(t, 4, om.GetPlayerCnt())
		assert.False(t, om.GetGameEndFlag())
		assert.Equal(t, -1, om.GetLoserIdx())
		assert.Nil(t, om.GetLastDiscardedCards())
		assert.Nil(t, om.GetHumanAction())
	})

	t.Run("success Reset distributes cards", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makePlayers()
		om := domain.NewOldMaid(tc, players)
		om.Reset()
		totalCards := 0
		for i := 0; i < om.GetPlayerCnt(); i++ {
			totalCards += om.GetPlayer(i).GetCardsSize()
		}
		assert.True(t, totalCards > 0)
	})

	t.Run("success Reset clears state", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makePlayers()
		om := domain.NewOldMaid(tc, players)
		om.Reset()
		assert.Equal(t, -1, om.GetLastDrawPlayerIdx())
		assert.Equal(t, -1, om.GetLastDrawFromIdx())
		assert.Equal(t, 0, om.GetLastDiscardedPairs())
		assert.Nil(t, om.GetLastDiscardedCards())
		assert.False(t, om.GetHasDrawn())
		assert.Nil(t, om.GetHumanAction())
	})

	t.Run("success GetPlayer valid index", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makePlayers()
		om := domain.NewOldMaid(tc, players)
		assert.NotNil(t, om.GetPlayer(0))
		assert.True(t, om.GetPlayer(0).GetIsHuman())
		assert.NotNil(t, om.GetPlayer(1))
		assert.False(t, om.GetPlayer(1).GetIsHuman())
	})

	t.Run("success GetPlayer invalid index returns nil", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makePlayers()
		om := domain.NewOldMaid(tc, players)
		assert.Nil(t, om.GetPlayer(-1))
		assert.Nil(t, om.GetPlayer(10))
	})

	t.Run("success IsHumanTurn at start", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makePlayers()
		om := domain.NewOldMaid(tc, players)
		om.Reset()
		_ = om.IsHumanTurn()
	})

	t.Run("success GetCurrentTurn", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makePlayers()
		om := domain.NewOldMaid(tc, players)
		turn := om.GetCurrentTurn()
		assert.True(t, turn >= 0 && turn < domain.OldMaidPlayerCnt)
	})

	t.Run("success PlayerDraw returns ErrGameEnded when game ended", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makePlayers()
		om := domain.NewOldMaid(tc, players)
		// Set up a game-ending scenario: player 0 draws a pair, everyone else finished
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		players[2].SetIsFinished(true)
		players[3].SetIsFinished(true)
		// First draw ends the game (pair discarded, all finish except joker holder)
		err := om.PlayerDraw(0)
		assert.NoError(t, err)
		assert.True(t, om.GetGameEndFlag())
		// Second draw should return ErrGameEnded
		err = om.PlayerDraw(0)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrGameEnded)
	})

	t.Run("success PlayerDraw returns ErrNotHumanTurn when not human turn", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		// All CPU players so PlayerDraw should return ErrNotHumanTurn
		cpuPlayers := []*domain.OldMaidPlayer{
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
		}
		om := domain.NewOldMaid(tc, cpuPlayers)
		cpuPlayers[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		cpuPlayers[1].AddCard(domain.NewCard(domain.CardDesignClover, 3, false))
		cpuPlayers[2].SetIsFinished(true)
		cpuPlayers[3].SetIsFinished(true)
		err := om.PlayerDraw(0)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
	})

	t.Run("success CpuDraw returns ErrGameEnded when game ended", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		cpuPlayers := []*domain.OldMaidPlayer{
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
		}
		om := domain.NewOldMaid(tc, cpuPlayers)
		// Set up a game-ending scenario
		cpuPlayers[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		cpuPlayers[1].AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		cpuPlayers[2].SetIsFinished(true)
		cpuPlayers[3].SetIsFinished(true)
		// First draw ends the game
		err := om.CpuDraw()
		assert.NoError(t, err)
		assert.True(t, om.GetGameEndFlag())
		// Second draw should return ErrGameEnded
		err = om.CpuDraw()
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrGameEnded)
	})

	t.Run("success CpuDraw returns ErrNotHumanTurn when human turn", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makePlayers()
		om := domain.NewOldMaid(tc, players)
		om.Reset()
		// Force turn to 0 (human)
		// (Actually Reset might set turn to 0)
		if om.IsHumanTurn() {
			prevTurn := om.GetCurrentTurn()
			err := om.CpuDraw()
			assert.Error(t, err)
			assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
			assert.Equal(t, prevTurn, om.GetCurrentTurn())
		}
	})

	t.Run("success GetNextDrawTargetIdx returns valid player", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makePlayers()
		om := domain.NewOldMaid(tc, players)
		om.Reset()
		if !om.GetGameEndFlag() {
			targetIdx := om.GetNextDrawTargetIdx()
			assert.True(t, targetIdx >= 0 && targetIdx < domain.OldMaidPlayerCnt)
		}
	})

	t.Run("success GetLoserIdx before game end", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makePlayers()
		om := domain.NewOldMaid(tc, players)
		assert.Equal(t, -1, om.GetLoserIdx())
	})

	t.Run("success PlayerDraw populates LastDiscardedCards", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makePlayers()
		om := domain.NewOldMaid(tc, players)
		// Player 0: SPADE 5
		// Player 1: CLOVER 5
		// Players 2,3: Finished
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		players[2].SetIsFinished(true)
		players[3].SetIsFinished(true)

		// Player 0 draws CLOVER 5 from Player 1
		err := om.PlayerDraw(0)
		assert.NoError(t, err)

		assert.True(t, om.GetHasDrawn())
		assert.Equal(t, 1, om.GetLastDiscardedPairs())
		discarded := om.GetLastDiscardedCards()
		assert.Equal(t, 2, len(discarded))
		// Check that we have the two 5s
		values := []int{discarded[0].GetValue(), discarded[1].GetValue()}
		assert.Equal(t, 5, values[0])
		assert.Equal(t, 5, values[1])
		// Check humanAction is populated
		ha := om.GetHumanAction()
		assert.NotNil(t, ha)
		assert.Equal(t, 0, ha.DrawPlayerIdx)
		assert.Equal(t, 1, ha.DrawFromIdx)
		assert.Equal(t, 1, ha.DiscardedPairs)
		assert.Equal(t, 2, len(ha.DiscardedCards))
	})

	t.Run("success CpuDraw populates DiscardedCards in action", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		// Custom players for this test
		cpuPlayers := []*domain.OldMaidPlayer{
			domain.NewOldMaidPlayer(false), // Player 0 is CPU
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
		}
		omCpu := domain.NewOldMaid(tc, cpuPlayers)
		// Player 0: HEART 10
		// Player 1: DIAMOND 10
		omCpu.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
		omCpu.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignDiamond, 10, false))
		omCpu.GetPlayer(2).SetIsFinished(true)
		omCpu.GetPlayer(3).SetIsFinished(true)

		// Now turn is 0 (CPU). CpuDraw should work.
		err := omCpu.CpuDraw()
		assert.NoError(t, err)

		actions := omCpu.GetCpuActions()
		assert.Equal(t, 1, len(actions))
		assert.Equal(t, 1, actions[0].DiscardedPairs)
		assert.Equal(t, 2, len(actions[0].DiscardedCards))
		assert.Equal(t, 10, actions[0].DiscardedCards[0].GetValue())
	})

	t.Run("success Reset shuffles player order", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makePlayers()
		om := domain.NewOldMaid(tc, players)

		// Run Reset many times and check that the human player
		// does not always end up at index 0.
		humanNotAtZero := false
		for i := 0; i < 50; i++ {
			om.Reset()
			if !om.GetPlayer(0).GetIsHuman() {
				humanNotAtZero = true
				break
			}
		}
		assert.True(t, humanNotAtZero, "player order should be randomized after Reset")
	})

	t.Run("success PlayerDraw shuffles hand after draw", func(t *testing.T) {
		// Player 0 の手に複数のカードがあるとき、引いたカードが末尾固定にならないことを確認
		notAlwaysLast := false
		for attempt := 0; attempt < 100; attempt++ {
			tc := domain.NewTrumpCards(1)
			players := []*domain.OldMaidPlayer{
				domain.NewOldMaidPlayer(true),
				domain.NewOldMaidPlayer(false),
				domain.NewOldMaidPlayer(false),
				domain.NewOldMaidPlayer(false),
			}
			om := domain.NewOldMaid(tc, players)
			// Player 0: cards 2,3,4,6,8 (odd number, no pairs)
			players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
			players[0].AddCard(domain.NewCard(domain.CardDesignClover, 3, false))
			players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 4, false))
			players[0].AddCard(domain.NewCard(domain.CardDesignDiamond, 6, false))
			players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
			// Player 1: Joker (no pairs possible)
			players[1].AddCard(domain.NewCard(domain.CardDesignJoker, domain.CardValueJoker, false))
			players[2].SetIsFinished(true)
			players[3].SetIsFinished(true)

			_ = om.PlayerDraw(0)

			// After draw player 0 has 6 cards (5 + joker, no pair discarded)
			if om.GetPlayer(0).GetCardsSize() > 0 {
				// Check if last card is NOT the joker (meaning it was shuffled)
				lastCard := om.GetPlayer(0).GetCard(om.GetPlayer(0).GetCardsSize() - 1)
				if lastCard.GetDesign() != domain.CardDesignJoker {
					notAlwaysLast = true
					break
				}
			}
		}
		assert.True(t, notAlwaysLast, "drawn card should not always be at the last position after shuffle")
	})

	t.Run("success CpuDraw with improved strategy still draws and records action", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		cpuPlayers := []*domain.OldMaidPlayer{
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
		}
		om := domain.NewOldMaid(tc, cpuPlayers)
		// Player 0: spade 2
		// Player 1: spade 3, heart 3 (will pair up immediately if drawn)
		// Players 2,3: finished
		cpuPlayers[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		cpuPlayers[1].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		cpuPlayers[1].AddCard(domain.NewCard(domain.CardDesignHeart, 4, false))
		cpuPlayers[1].AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
		cpuPlayers[2].SetIsFinished(true)
		cpuPlayers[3].SetIsFinished(true)

		err := om.CpuDraw()
		assert.NoError(t, err)

		actions := om.GetCpuActions()
		assert.Equal(t, 1, len(actions))
		assert.Equal(t, 0, actions[0].DrawPlayerIdx)
		assert.Equal(t, 1, actions[0].DrawFromIdx)
	})

	t.Run("success CpuDraw edge card selection covers first and last positions", func(t *testing.T) {
		firstSelected := false
		lastSelected := false
		for attempt := 0; attempt < 200; attempt++ {
			tc := domain.NewTrumpCards(1)
			cpuPlayers := []*domain.OldMaidPlayer{
				domain.NewOldMaidPlayer(false),
				domain.NewOldMaidPlayer(false),
				domain.NewOldMaidPlayer(false),
				domain.NewOldMaidPlayer(false),
			}
			om := domain.NewOldMaid(tc, cpuPlayers)
			// Player 0 draws from Player 1 who has 5 cards (distinct values, no pairs with player 0)
			cpuPlayers[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
			// Player 1: 5 cards with distinct values that won't pair with spade 2
			cpuPlayers[1].AddCard(domain.NewCard(domain.CardDesignSpade, 4, false))
			cpuPlayers[1].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
			cpuPlayers[1].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
			cpuPlayers[1].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
			cpuPlayers[1].AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))
			cpuPlayers[2].SetIsFinished(true)
			cpuPlayers[3].SetIsFinished(true)

			// Record which card (by value) is at index 0 and last in player 1's hand
			firstVal := cpuPlayers[1].GetCard(0).GetValue()
			lastVal := cpuPlayers[1].GetCard(cpuPlayers[1].GetCardsSize() - 1).GetValue()

			_ = om.CpuDraw()

			actions := om.GetCpuActions()
			if len(actions) == 1 && actions[0].DrawnCard != nil {
				drawnVal := actions[0].DrawnCard.GetValue()
				if drawnVal == firstVal {
					firstSelected = true
				}
				if drawnVal == lastVal {
					lastSelected = true
				}
			}
			if firstSelected && lastSelected {
				break
			}
		}
		assert.True(t, firstSelected, "CPU should sometimes select the first card")
		assert.True(t, lastSelected, "CPU should sometimes select the last card")
	})

	t.Run("success Reset preserves all players after shuffle", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makePlayers()
		om := domain.NewOldMaid(tc, players)
		om.Reset()

		humanCnt := 0
		cpuCnt := 0
		for i := 0; i < om.GetPlayerCnt(); i++ {
			if om.GetPlayer(i).GetIsHuman() {
				humanCnt++
			} else {
				cpuCnt++
			}
		}
		assert.Equal(t, 1, humanCnt)
		assert.Equal(t, 3, cpuCnt)
	})
}
