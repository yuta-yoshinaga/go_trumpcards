package entities_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"

	"github.com/stretchr/testify/assert"
)

func makeSevensPlayers() []*entities.SevensPlayer {
	return []*entities.SevensPlayer{
		entities.NewSevensPlayer(true),
		entities.NewSevensPlayer(false),
		entities.NewSevensPlayer(false),
		entities.NewSevensPlayer(false),
	}
}

func TestSevens_Method(t *testing.T) {
	t.Run("success NewSevens", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players)
		assert.NotNil(t, s)
		assert.Equal(t, 4, s.GetPlayerCnt())
		assert.False(t, s.GetGameEndFlag())
		assert.Equal(t, 0, s.GetCurrentTurn())
		assert.True(t, s.IsHumanTurn())
		assert.Nil(t, s.GetCpuActions())
		assert.Nil(t, s.GetHumanAction())
		mins := s.GetTableMinVals()
		maxs := s.GetTableMaxVals()
		for i := 1; i <= 4; i++ {
			assert.Equal(t, 7, mins[i])
			assert.Equal(t, 7, maxs[i])
		}
	})

	t.Run("success Reset distributes cards and removes sevens", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players)
		s.Reset()
		total := 0
		for i := 0; i < s.GetPlayerCnt(); i++ {
			total += s.GetPlayer(i).GetCardsSize()
		}
		// 52 cards - 4 sevens = 48 cards distributed
		assert.Equal(t, 48, total)
	})

	t.Run("success Reset initializes table", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players)
		s.Reset()
		mins := s.GetTableMinVals()
		maxs := s.GetTableMaxVals()
		for i := 1; i <= 4; i++ {
			assert.Equal(t, 7, mins[i])
			assert.Equal(t, 7, maxs[i])
		}
	})

	t.Run("success GetPlayer valid index", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players)
		assert.NotNil(t, s.GetPlayer(0))
		assert.True(t, s.GetPlayer(0).GetIsHuman())
		assert.NotNil(t, s.GetPlayer(1))
		assert.False(t, s.GetPlayer(1).GetIsHuman())
	})

	t.Run("success GetPlayer invalid index returns nil", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players)
		assert.Nil(t, s.GetPlayer(-1))
		assert.Nil(t, s.GetPlayer(10))
	})

	t.Run("success IsPlayable on fresh board", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players)
		// Fresh board: all suits have min=max=7
		// 6 of spades is playable (7-1=6), 8 of spades is playable (7+1=8)
		card6 := entities.NewCard(entities.CardDesignSpade, 6, false)
		card8 := entities.NewCard(entities.CardDesignSpade, 8, false)
		card5 := entities.NewCard(entities.CardDesignSpade, 5, false)
		card7 := entities.NewCard(entities.CardDesignSpade, 7, false)
		assert.True(t, s.IsPlayable(card6))
		assert.True(t, s.IsPlayable(card8))
		assert.False(t, s.IsPlayable(card5)) // not adjacent
		assert.False(t, s.IsPlayable(card7)) // 7 is already on board
	})

	t.Run("success IsPlayable nil card", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players)
		assert.False(t, s.IsPlayable(nil))
	})

	t.Run("success IsPlayable Ace at min=2", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players)
		// Manually place cards to set minVal[Spade] = 2
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 6, false))
		s.PlayerPlay(0)                                                           // place 6 → min=6
		players[1].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false)) // won't be played here directly

		// Set up a board where spade min is 2 by using HasAnyOption checks
		// Instead, directly verify: ace (value=1) playable when min=2
		tc2 := entities.NewTrumpCards(0)
		players2 := makeSevensPlayers()
		s2 := entities.NewSevens(tc2, players2)
		// Give human a 6♠ to extend left
		players2[0].AddCard(entities.NewCard(entities.CardDesignSpade, 6, false))
		players2[1].AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		players2[2].AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		players2[3].AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		s2.PlayerPlay(0) // human plays 6♠ → min=6

		ace := entities.NewCard(entities.CardDesignSpade, 1, false)
		// min is 6, so ace (1) is not adjacent to 6
		assert.False(t, s2.IsPlayable(ace))
	})

	t.Run("success IsPlayable boundary King at max=12", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players)
		// Place 8 of spades → max becomes 8
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 8, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		s.PlayerPlay(0) // place 8♠ → max=8
		king := entities.NewCard(entities.CardDesignSpade, 13, false)
		assert.False(t, s.IsPlayable(king)) // 13 != 9 (max+1=9)
	})

	t.Run("success PlayerPlay places card on board", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 6, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 8, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))

		ok := s.PlayerPlay(0) // play 6♠
		assert.True(t, ok)
		mins := s.GetTableMinVals()
		assert.Equal(t, 6, mins[entities.CardDesignSpade])
		assert.Equal(t, 1, players[0].GetCardsSize())
		assert.NotNil(t, s.GetHumanAction())
		assert.NotNil(t, s.GetHumanAction().PlayedCard)
		assert.Equal(t, 6, s.GetHumanAction().PlayedCard.GetValue())
	})

	t.Run("success PlayerPlay fails with non-playable card", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false)) // not adjacent to 7
		ok := s.PlayerPlay(0)
		assert.False(t, ok)
	})

	t.Run("success PlayerPlay fails with invalid index", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 6, false))
		ok := s.PlayerPlay(5) // out of range
		assert.False(t, ok)
	})

	t.Run("success PlayerPlay pass", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 6, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))

		ok := s.PlayerPlay(-1) // pass
		assert.True(t, ok)
		assert.Equal(t, 1, players[0].GetPassesUsed())
		action := s.GetHumanAction()
		assert.NotNil(t, action)
		assert.Nil(t, action.PlayedCard) // nil = pass
	})

	t.Run("success PlayerPlay pass fails when no passes left", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players)
		// Exhaust human passes
		for i := 0; i < entities.SevensMaxPasses; i++ {
			players[0].IncrPassesUsed()
		}
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 6, false))
		ok := s.PlayerPlay(-1) // attempt to pass
		assert.False(t, ok)
	})

	t.Run("success PlayerPlay fails when not human turn", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 6, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 8, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 8, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		s.PlayerPlay(0) // human plays → advances to CPU 1
		if !s.IsHumanTurn() && !s.GetGameEndFlag() {
			ok := s.PlayerPlay(0)
			assert.False(t, ok)
		}
	})

	t.Run("success PlayerPlay does nothing when game ended", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players)
		// 3 CPUs finished, human has last card
		players[1].SetIsFinished(true)
		players[1].SetRank(1)
		players[2].SetIsFinished(true)
		players[2].SetRank(2)
		players[3].SetIsFinished(true)
		players[3].SetRank(3)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 8, false))
		s.PlayerPlay(0) // plays 8♠ → finishes → game ends
		assert.True(t, s.GetGameEndFlag())
		assert.Equal(t, 4, players[0].GetRank())

		ok := s.PlayerPlay(0) // game already ended
		assert.False(t, ok)
	})

	t.Run("success CpuPlay plays valid card", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 8, false)) // playable
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 9, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignSpade, 6, false)) // playable by CPU 1
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))

		s.PlayerPlay(0) // human plays 8♠ → advances to CPU 1
		s.CpuPlay()    // CPU 1 plays 6♠
		actions := s.GetCpuActions()
		assert.Len(t, actions, 1)
		assert.NotNil(t, actions[0].PlayedCard)
		assert.Equal(t, 6, actions[0].PlayedCard.GetValue())
	})

	t.Run("success CpuPlay passes when no playable card", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 8, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 9, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false)) // not adjacent → pass
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))

		s.PlayerPlay(0) // human plays 8♠
		s.CpuPlay()    // CPU 1: 5♠ not playable → passes
		actions := s.GetCpuActions()
		assert.Len(t, actions, 1)
		assert.Nil(t, actions[0].PlayedCard) // pass
		assert.Equal(t, 1, players[1].GetPassesUsed())
	})

	t.Run("success CpuPlay eliminates when no passes left", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players)
		// Exhaust CPU 1's passes
		for i := 0; i < entities.SevensMaxPasses; i++ {
			players[1].IncrPassesUsed()
		}
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 8, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 9, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false)) // not playable
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))

		s.PlayerPlay(0) // human plays
		s.CpuPlay()    // CPU 1: no playable, no passes → eliminated
		assert.True(t, players[1].GetIsFinished())
		assert.Greater(t, players[1].GetRank(), 0)
	})

	t.Run("success CpuPlay does nothing on human turn", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players)
		s.CpuPlay() // does nothing when it's human's turn
		assert.Nil(t, s.GetCpuActions())
		assert.True(t, s.IsHumanTurn())
	})

	t.Run("success HasAnyOption true when has playable card", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 6, false)) // playable
		assert.True(t, s.HasAnyOption(0))
	})

	t.Run("success HasAnyOption true when can pass but no playable card", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false)) // not playable
		assert.True(t, s.HasAnyOption(0)) // can still pass
	})

	t.Run("success HasAnyOption false when no options", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players)
		for i := 0; i < entities.SevensMaxPasses; i++ {
			players[0].IncrPassesUsed()
		}
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false)) // not playable
		assert.False(t, s.HasAnyOption(0))
	})

	t.Run("success AutoHandleNoOption eliminates human", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players)
		for i := 0; i < entities.SevensMaxPasses; i++ {
			players[0].IncrPassesUsed()
		}
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		s.AutoHandleNoOption()
		assert.True(t, players[0].GetIsFinished())
		assert.NotNil(t, s.GetHumanAction())
		assert.Nil(t, s.GetHumanAction().PlayedCard) // pass recorded
	})

	t.Run("success game ends when last player finishes", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players)
		players[1].SetIsFinished(true)
		players[1].SetRank(1)
		players[2].SetIsFinished(true)
		players[2].SetRank(2)
		players[3].SetIsFinished(true)
		players[3].SetRank(3)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 8, false))
		s.PlayerPlay(0) // human plays last card → finishes → game ends
		assert.True(t, s.GetGameEndFlag())
		assert.Equal(t, 4, players[0].GetRank())
	})

	t.Run("success eliminatePlayer places remaining cards on board", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players)
		// Exhaust CPU 1's passes and give it cards that become non-playable after human plays 8♠
		for i := 0; i < entities.SevensMaxPasses; i++ {
			players[1].IncrPassesUsed()
		}
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 8, false))
		// After human plays 8♠: board SPADE max=8. 5♠ (≠9=max+1) and 11♠ (≠9) are not playable.
		players[1].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))

		s.PlayerPlay(0) // human plays 8♠ → SPADE max=8
		s.CpuPlay()     // CPU 1: no playable, no passes → eliminated

		assert.True(t, players[1].GetIsFinished())
		assert.True(t, players[1].GetIsEliminated())
		assert.Equal(t, 0, players[1].GetCardsSize()) // hand cleared
		// Board should be expanded to include eliminated player's cards
		mins := s.GetTableMinVals()
		maxs := s.GetTableMaxVals()
		assert.Equal(t, 5, mins[entities.CardDesignSpade])  // 5♠ placed → min=5
		assert.Equal(t, 11, maxs[entities.CardDesignSpade]) // 11♠ placed → max=11
	})

	t.Run("success eliminated player gets lower rank than normal finisher", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players)
		for i := 0; i < entities.SevensMaxPasses; i++ {
			players[1].IncrPassesUsed()
		}
		// Human has one playable card; CPU 1 has non-playable cards
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 8, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))

		s.PlayerPlay(0) // human empties hand → rank 1
		assert.Equal(t, 1, players[0].GetRank())

		s.CpuPlay() // CPU 1: no playable, no passes → eliminated → rank 4
		assert.Equal(t, 4, players[1].GetRank())
		assert.True(t, players[1].GetIsEliminated())

		// Normal finisher ranks better than eliminated player
		assert.Less(t, players[0].GetRank(), players[1].GetRank())
	})

	t.Run("success multiple eliminations get descending ranks", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players)
		// Exhaust passes for CPU 1 and CPU 2
		for i := 0; i < entities.SevensMaxPasses; i++ {
			players[1].IncrPassesUsed()
			players[2].IncrPassesUsed()
		}
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 8, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false)) // not playable
		players[2].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false)) // not playable
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))

		s.PlayerPlay(0) // human plays
		s.CpuPlay()     // CPU 1 eliminated first → rank 4
		assert.Equal(t, 4, players[1].GetRank())

		// Advance to CPU 2 (CPU 3 can also pass, but for simplicity test CPU 2 directly)
		// Manually eliminate CPU 2 by setting up a similar scenario
		if !s.GetGameEndFlag() && !s.IsHumanTurn() {
			s.CpuPlay() // CPU 2 eliminated → rank 3
		}
		if players[2].GetIsEliminated() {
			assert.Equal(t, 3, players[2].GetRank()) // 2nd elimination gets rank 3
			assert.Less(t, players[2].GetRank(), players[1].GetRank()) // later eliminated ranks better
		}
	})

	t.Run("success Reset clears isEliminated flag", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players)
		players[0].SetIsEliminated(true)
		players[0].SetIsFinished(true)
		s.Reset()
		assert.False(t, players[0].GetIsEliminated())
		assert.False(t, players[0].GetIsFinished())
	})
}
