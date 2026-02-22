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
		s := entities.NewSevens(tc, players, entities.DefaultSevensConfig())
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

	t.Run("success NewSevens stores config", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		cfg := entities.SevensConfig{TunnelEnabled: true, JokerCount: 2, CpuStrategy: true}
		s := entities.NewSevens(tc, players, cfg)
		assert.Equal(t, cfg, s.GetConfig())
	})

	t.Run("success Reset distributes cards and removes sevens", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players, entities.DefaultSevensConfig())
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
		s := entities.NewSevens(tc, players, entities.DefaultSevensConfig())
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
		s := entities.NewSevens(tc, players, entities.DefaultSevensConfig())
		assert.NotNil(t, s.GetPlayer(0))
		assert.True(t, s.GetPlayer(0).GetIsHuman())
		assert.NotNil(t, s.GetPlayer(1))
		assert.False(t, s.GetPlayer(1).GetIsHuman())
	})

	t.Run("success GetPlayer invalid index returns nil", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players, entities.DefaultSevensConfig())
		assert.Nil(t, s.GetPlayer(-1))
		assert.Nil(t, s.GetPlayer(10))
	})

	t.Run("success IsPlayable on fresh board", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players, entities.DefaultSevensConfig())
		// Fresh board: all suits have only 7 placed
		// 6 of spades is playable (adjacent to 7), 8 of spades is playable (adjacent to 7)
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
		s := entities.NewSevens(tc, players, entities.DefaultSevensConfig())
		assert.False(t, s.IsPlayable(nil))
	})

	t.Run("success IsPlayable Ace at min=2", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players, entities.DefaultSevensConfig())
		// Give human a 6♠ to extend left
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 6, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		s.PlayerPlay(0) // human plays 6♠ → board has 6,7

		ace := entities.NewCard(entities.CardDesignSpade, 1, false)
		// min is 6, so ace (1) is not adjacent to 6
		assert.False(t, s.IsPlayable(ace))
	})

	t.Run("success IsPlayable boundary King at max=12", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players, entities.DefaultSevensConfig())
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
		s := entities.NewSevens(tc, players, entities.DefaultSevensConfig())
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
		s := entities.NewSevens(tc, players, entities.DefaultSevensConfig())
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false)) // not adjacent to 7
		ok := s.PlayerPlay(0)
		assert.False(t, ok)
	})

	t.Run("success PlayerPlay fails with invalid index", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players, entities.DefaultSevensConfig())
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 6, false))
		ok := s.PlayerPlay(5) // out of range
		assert.False(t, ok)
	})

	t.Run("success PlayerPlay pass", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players, entities.DefaultSevensConfig())
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
		s := entities.NewSevens(tc, players, entities.DefaultSevensConfig())
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
		s := entities.NewSevens(tc, players, entities.DefaultSevensConfig())
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
		s := entities.NewSevens(tc, players, entities.DefaultSevensConfig())
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
		s := entities.NewSevens(tc, players, entities.DefaultSevensConfig())
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 8, false)) // playable
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 9, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignSpade, 6, false)) // playable by CPU 1
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))

		s.PlayerPlay(0) // human plays 8♠ → advances to CPU 1
		s.CpuPlay()     // CPU 1 plays 6♠
		actions := s.GetCpuActions()
		assert.Len(t, actions, 1)
		assert.NotNil(t, actions[0].PlayedCard)
		assert.Equal(t, 6, actions[0].PlayedCard.GetValue())
	})

	t.Run("success CpuPlay passes when no playable card", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players, entities.DefaultSevensConfig())
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 8, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 9, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false)) // not adjacent → pass
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))

		s.PlayerPlay(0) // human plays 8♠
		s.CpuPlay()     // CPU 1: 5♠ not playable → passes
		actions := s.GetCpuActions()
		assert.Len(t, actions, 1)
		assert.Nil(t, actions[0].PlayedCard) // pass
		assert.Equal(t, 1, players[1].GetPassesUsed())
	})

	t.Run("success CpuPlay eliminates when no passes left", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players, entities.DefaultSevensConfig())
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
		s.CpuPlay()     // CPU 1: no playable, no passes → eliminated
		assert.True(t, players[1].GetIsFinished())
		assert.Greater(t, players[1].GetRank(), 0)
	})

	t.Run("success CpuPlay does nothing on human turn", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players, entities.DefaultSevensConfig())
		s.CpuPlay() // does nothing when it's human's turn
		assert.Nil(t, s.GetCpuActions())
		assert.True(t, s.IsHumanTurn())
	})

	t.Run("success HasAnyOption true when has playable card", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players, entities.DefaultSevensConfig())
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 6, false)) // playable
		assert.True(t, s.HasAnyOption(0))
	})

	t.Run("success HasAnyOption true when can pass but no playable card", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players, entities.DefaultSevensConfig())
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false)) // not playable
		assert.True(t, s.HasAnyOption(0)) // can still pass
	})

	t.Run("success HasAnyOption false when no options", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players, entities.DefaultSevensConfig())
		for i := 0; i < entities.SevensMaxPasses; i++ {
			players[0].IncrPassesUsed()
		}
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false)) // not playable
		assert.False(t, s.HasAnyOption(0))
	})

	t.Run("success AutoHandleNoOption eliminates human", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players, entities.DefaultSevensConfig())
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
		s := entities.NewSevens(tc, players, entities.DefaultSevensConfig())
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
		s := entities.NewSevens(tc, players, entities.DefaultSevensConfig())
		// Exhaust CPU 1's passes and give it cards that become non-playable after human plays 8♠
		for i := 0; i < entities.SevensMaxPasses; i++ {
			players[1].IncrPassesUsed()
		}
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 8, false))
		// After human plays 8♠: board SPADE has 7,8. 5♠ and 11♠ are not playable.
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
		s := entities.NewSevens(tc, players, entities.DefaultSevensConfig())
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
		s := entities.NewSevens(tc, players, entities.DefaultSevensConfig())
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

		if !s.GetGameEndFlag() && !s.IsHumanTurn() {
			s.CpuPlay() // CPU 2 eliminated → rank 3
		}
		if players[2].GetIsEliminated() {
			assert.Equal(t, 3, players[2].GetRank())
			assert.Less(t, players[2].GetRank(), players[1].GetRank())
		}
	})

	t.Run("success Reset clears isEliminated flag", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players, entities.DefaultSevensConfig())
		players[0].SetIsEliminated(true)
		players[0].SetIsFinished(true)
		s.Reset()
		assert.False(t, players[0].GetIsEliminated())
		assert.False(t, players[0].GetIsFinished())
	})

	t.Run("success Reset shuffles player order", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players, entities.DefaultSevensConfig())

		humanNotAtZero := false
		for i := 0; i < 50; i++ {
			s.Reset()
			if !s.GetPlayer(0).GetIsHuman() {
				humanNotAtZero = true
				break
			}
		}
		assert.True(t, humanNotAtZero, "player order should be randomized after Reset")
	})

	t.Run("success Reset preserves all players after shuffle", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players, entities.DefaultSevensConfig())
		s.Reset()

		humanCnt := 0
		cpuCnt := 0
		for i := 0; i < s.GetPlayerCnt(); i++ {
			if s.GetPlayer(i).GetIsHuman() {
				humanCnt++
			} else {
				cpuCnt++
			}
		}
		assert.Equal(t, 1, humanCnt)
		assert.Equal(t, 3, cpuCnt)
	})

	t.Run("success bitmask board: GetTablePlaced reflects placed cards", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players, entities.DefaultSevensConfig())
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 6, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))

		placed := s.GetTablePlaced()
		// Initially only 7 is placed for each suit
		assert.Equal(t, uint16(1<<7), placed[entities.CardDesignSpade])

		s.PlayerPlay(0) // place 6♠
		placed = s.GetTablePlaced()
		assert.Equal(t, uint16((1<<7)|(1<<6)), placed[entities.CardDesignSpade])
	})

	t.Run("success derived GetTableMinVals/MaxVals match bitmask", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players, entities.DefaultSevensConfig())
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 6, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 8, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))

		s.PlayerPlay(0) // play 6♠
		mins := s.GetTableMinVals()
		maxs := s.GetTableMaxVals()
		assert.Equal(t, 6, mins[entities.CardDesignSpade])
		assert.Equal(t, 7, maxs[entities.CardDesignSpade])
	})
}

func setupSpadeBoard(config entities.SevensConfig, spadesToPlace []int) *entities.Sevens {
	tc := entities.NewTrumpCards(0)
	players := makeSevensPlayers()
	s := entities.NewSevens(tc, players, config)
	// Give all players enough dummy cards so they never finish
	for i := 0; i < 4; i++ {
		for d := 0; d < 10; d++ {
			players[i].AddCard(entities.NewCard(entities.CardDesignDiamond, 2, false))
		}
	}
	for _, v := range spadesToPlace {
		if s.GetGameEndFlag() {
			break
		}
		if s.IsHumanTurn() {
			players[0].AddCard(entities.NewCard(entities.CardDesignSpade, v, false))
			// Play the last card (newly added spade)
			idx := players[0].GetCardsSize() - 1
			s.PlayerPlay(idx)
		} else {
			cpuIdx := s.GetCurrentTurn()
			players[cpuIdx].AddCard(entities.NewCard(entities.CardDesignSpade, v, false))
			s.CpuPlay()
		}
	}
	return s
}

func TestSevens_Tunnel(t *testing.T) {
	tunnelConfig := entities.SevensConfig{TunnelEnabled: true, JokerCount: 0, CpuStrategy: false}

	t.Run("success tunnel: Ace playable when 2 is placed", func(t *testing.T) {
		s := setupSpadeBoard(tunnelConfig, []int{6, 5, 4, 3, 2})
		ace := entities.NewCard(entities.CardDesignSpade, 1, false)
		assert.True(t, s.IsPlayable(ace))
	})

	t.Run("success tunnel: Ace playable when King placed via circular wrap", func(t *testing.T) {
		s := setupSpadeBoard(tunnelConfig, []int{8, 9, 10, 11, 12, 13})
		ace := entities.NewCard(entities.CardDesignSpade, 1, false)
		assert.True(t, s.IsPlayable(ace))
	})

	t.Run("success tunnel disabled: Ace not playable when King placed", func(t *testing.T) {
		noTunnel := entities.DefaultSevensConfig()
		s := setupSpadeBoard(noTunnel, []int{8, 9, 10, 11, 12, 13})
		ace := entities.NewCard(entities.CardDesignSpade, 1, false)
		assert.False(t, s.IsPlayable(ace))
	})

	t.Run("success tunnel: King playable when Ace placed via circular wrap", func(t *testing.T) {
		s := setupSpadeBoard(tunnelConfig, []int{6, 5, 4, 3, 2, 1})
		king := entities.NewCard(entities.CardDesignSpade, 13, false)
		assert.True(t, s.IsPlayable(king))
	})
}

func TestSevens_Joker(t *testing.T) {
	jokerConfig := entities.SevensConfig{TunnelEnabled: false, JokerCount: 2, CpuStrategy: false}

	t.Run("success joker is playable when there are open board positions", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players, jokerConfig)
		joker := entities.NewCard(entities.CardDesignJoker, 1, false)
		// Fresh board: 6 and 8 are playable for each suit
		assert.True(t, s.IsPlayable(joker))
	})

	t.Run("success PlayerPlayJoker places on target position", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players, jokerConfig)
		players[0].AddCard(entities.NewCard(entities.CardDesignJoker, 1, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))

		// Play joker to SPADE 6 position
		ok := s.PlayerPlayJoker(0, entities.CardDesignSpade, 6)
		assert.True(t, ok)
		// Board should now have SPADE 6 placed
		placed := s.GetTablePlaced()
		assert.True(t, placed[entities.CardDesignSpade]&(1<<6) != 0)
		// Human action should record joker target
		action := s.GetHumanAction()
		assert.NotNil(t, action)
		assert.Equal(t, entities.CardDesignJoker, action.PlayedCard.GetDesign())
		assert.Equal(t, entities.CardDesignSpade, action.TargetSuit)
		assert.Equal(t, 6, action.TargetValue)
	})

	t.Run("success PlayerPlayJoker fails with non-joker card", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players, jokerConfig)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 6, false))
		ok := s.PlayerPlayJoker(0, entities.CardDesignSpade, 6)
		assert.False(t, ok)
	})

	t.Run("success PlayerPlayJoker fails with invalid target position", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players, jokerConfig)
		players[0].AddCard(entities.NewCard(entities.CardDesignJoker, 1, false))
		// SPADE 5 is not playable on fresh board (not adjacent to 7)
		ok := s.PlayerPlayJoker(0, entities.CardDesignSpade, 5)
		assert.False(t, ok)
	})

	t.Run("success PlayerPlayJoker fails when game ended", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players, jokerConfig)
		players[1].SetIsFinished(true)
		players[1].SetRank(1)
		players[2].SetIsFinished(true)
		players[2].SetRank(2)
		players[3].SetIsFinished(true)
		players[3].SetRank(3)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 8, false))
		s.PlayerPlay(0) // game ends
		assert.True(t, s.GetGameEndFlag())

		players[0].AddCard(entities.NewCard(entities.CardDesignJoker, 1, false))
		ok := s.PlayerPlayJoker(0, entities.CardDesignSpade, 6)
		assert.False(t, ok)
	})

	t.Run("success CpuPlay plays joker", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players, jokerConfig)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 8, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 9, false))
		// CPU 1 only has a joker
		players[1].AddCard(entities.NewCard(entities.CardDesignJoker, 1, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))

		s.PlayerPlay(0) // human plays 8♠
		s.CpuPlay()     // CPU 1 plays joker

		actions := s.GetCpuActions()
		assert.Len(t, actions, 1)
		assert.NotNil(t, actions[0].PlayedCard)
		assert.Equal(t, entities.CardDesignJoker, actions[0].PlayedCard.GetDesign())
		assert.Greater(t, actions[0].TargetSuit, 0)
		assert.Greater(t, actions[0].TargetValue, 0)
	})

	t.Run("success eliminatePlayer places normal card on board", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		noJokerCfg := entities.DefaultSevensConfig()
		s := entities.NewSevens(tc, players, noJokerCfg)
		for i := 0; i < entities.SevensMaxPasses; i++ {
			players[1].IncrPassesUsed()
		}
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 8, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 9, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false)) // not playable
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))

		s.PlayerPlay(0) // human plays 8♠
		s.CpuPlay()     // CPU 1: 5♠ not playable, no passes → eliminated

		assert.True(t, players[1].GetIsEliminated())
		assert.Equal(t, 0, players[1].GetCardsSize())
		// 5♠ should be force-placed on board
		mins := s.GetTableMinVals()
		assert.Equal(t, 5, mins[entities.CardDesignSpade])
	})

	t.Run("success eliminatePlayer skips joker cards on forced placement", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players, jokerConfig)

		// Give CPU1 a joker + non-playable normal card (3♥ — not adjacent to 7 after joker is played)
		players[1].AddCard(entities.NewCard(entities.CardDesignJoker, 1, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 3, false))
		for i := 0; i < entities.SevensMaxPasses; i++ {
			players[1].IncrPassesUsed()
		}

		// Give human playable cards, other CPUs dummy non-playable cards with passes remaining
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 8, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 9, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))

		// Round 1: human plays 8♠ → CPU1 plays joker (playable) → CPU2 passes → CPU3 passes
		s.PlayerPlay(0) // human: 8♠
		assert.Equal(t, 1, s.GetCurrentTurn())
		s.CpuPlay() // CPU1: plays joker (board has open positions)
		actions := s.GetCpuActions()
		assert.Equal(t, entities.CardDesignJoker, actions[0].PlayedCard.GetDesign())
		// CPU1 now has only 5♠ (not playable) and no passes left
		assert.Equal(t, 1, players[1].GetCardsSize())

		s.CpuPlay() // CPU2: passes
		s.CpuPlay() // CPU3: passes

		// Round 2: human plays 9♠ → CPU1: 5♠ not playable, no passes → eliminated
		assert.Equal(t, 0, s.GetCurrentTurn())
		placedBefore := s.GetTablePlaced()
		s.PlayerPlay(0) // human: 9♠
		assert.Equal(t, 1, s.GetCurrentTurn())
		s.CpuPlay() // CPU1: eliminated

		assert.True(t, players[1].GetIsEliminated())
		assert.Equal(t, 0, players[1].GetCardsSize())
		// 3♥ should be force-placed on board (normal card placement)
		placedAfter := s.GetTablePlaced()
		assert.True(t, placedAfter[entities.CardDesignHeart]&(1<<3) != 0, "3♥ should be placed on board")
		// Joker suit index (0) should be unchanged — no joker artifact on board
		assert.Equal(t, placedBefore[entities.CardDesignJoker], placedAfter[entities.CardDesignJoker], "joker suit should not change in tablePlaced")
	})

	t.Run("success joker is not playable when board is full", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players, jokerConfig)

		// Fill all 4 suits completely by playing cards in alternating order
		suits := []int{entities.CardDesignSpade, entities.CardDesignClover, entities.CardDesignHeart, entities.CardDesignDiamond}
		values := []int{6, 8, 5, 9, 4, 10, 3, 11, 2, 12, 1, 13}
		for _, suit := range suits {
			for i, v := range values {
				pIdx := i % 4
				players[pIdx].AddCard(entities.NewCard(suit, v, false))
			}
		}
		// Give all players extra dummy cards so they never finish
		for i := 0; i < 4; i++ {
			for d := 0; d < 20; d++ {
				players[i].AddCard(entities.NewCard(entities.CardDesignDiamond, 2, false))
			}
		}
		// Play all rounds until no more playable cards
		for !s.GetGameEndFlag() {
			if s.IsHumanTurn() {
				played := false
				for i := 0; i < players[0].GetCardsSize(); i++ {
					if s.IsPlayable(players[0].GetCard(i)) {
						s.PlayerPlay(i)
						played = true
						break
					}
				}
				if !played {
					s.PlayerPlay(-1) // pass to advance
				}
			} else {
				s.CpuPlay()
			}
			// Safety: check if all suits are full
			placed := s.GetTablePlaced()
			allFull := true
			for _, suit := range suits {
				for v := 1; v <= 13; v++ {
					if placed[suit]&(1<<uint(v)) == 0 {
						allFull = false
						break
					}
				}
				if !allFull {
					break
				}
			}
			if allFull {
				break
			}
		}

		joker := entities.NewCard(entities.CardDesignJoker, 1, false)
		assert.False(t, s.IsPlayable(joker))
	})
}

func TestSevens_CpuStrategy(t *testing.T) {
	strategyConfig := entities.SevensConfig{TunnelEnabled: false, JokerCount: 0, CpuStrategy: true}

	t.Run("success strategic CPU passes when holding blocker", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players, strategyConfig)
		// CPU 1 has 6♠ (playable, adjacent to 7) but does NOT have 5♠
		// Playing 6♠ opens the path for opponents (score = -1 for low direction)
		// With strategy enabled and passes available, CPU should prefer to pass
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 8, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 9, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignSpade, 6, false)) // playable but blocks
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))

		s.PlayerPlay(0) // human plays 8♠
		s.CpuPlay()     // CPU 1: strategic evaluation

		actions := s.GetCpuActions()
		assert.Len(t, actions, 1)
		// CPU should pass since playing 6♠ has negative score and CPU has passes
		assert.Nil(t, actions[0].PlayedCard)
		assert.Equal(t, 1, players[1].GetPassesUsed())
	})

	t.Run("success strategic CPU plays when it has the chain card", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players, strategyConfig)
		// CPU 1 has 6♠ AND 5♠ → playing 6♠ has positive score (+2 for having 5♠)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 8, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 9, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignSpade, 6, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))

		s.PlayerPlay(0) // human plays 8♠
		s.CpuPlay()     // CPU 1: plays 6♠ (positive score: has 5♠)

		actions := s.GetCpuActions()
		assert.Len(t, actions, 1)
		assert.NotNil(t, actions[0].PlayedCard)
		assert.Equal(t, 6, actions[0].PlayedCard.GetValue())
	})

	t.Run("success strategic CPU plays when low on passes (reserve=1)", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := entities.NewSevens(tc, players, strategyConfig)
		// CPU 1 has used all passes except 1 (maxPasses-1 used)
		// Even with negative score, CPU must play because it can't afford to pass
		for i := 0; i < entities.SevensMaxPasses-1; i++ {
			players[1].IncrPassesUsed()
		}
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 8, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 9, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignSpade, 6, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))

		s.PlayerPlay(0) // human plays 8♠
		s.CpuPlay()     // CPU 1: forced to play (only 1 pass remaining)

		actions := s.GetCpuActions()
		assert.Len(t, actions, 1)
		assert.NotNil(t, actions[0].PlayedCard)
		assert.Equal(t, 6, actions[0].PlayedCard.GetValue())
	})

	t.Run("success non-strategic CPU always plays first available card", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeSevensPlayers()
		noStrategy := entities.DefaultSevensConfig()
		s := entities.NewSevens(tc, players, noStrategy)
		// Same setup as "holding blocker" test but without strategy
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 8, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 9, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignSpade, 6, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))

		s.PlayerPlay(0) // human plays 8♠
		s.CpuPlay()     // CPU 1: non-strategic, plays 6♠

		actions := s.GetCpuActions()
		assert.Len(t, actions, 1)
		assert.NotNil(t, actions[0].PlayedCard) // plays the card instead of passing
		assert.Equal(t, 6, actions[0].PlayedCard.GetValue())
	})
}
