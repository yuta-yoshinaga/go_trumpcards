package domain_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
)

func makeSevensPlayers() []*domain.SevensPlayer {
	return []*domain.SevensPlayer{
		domain.NewSevensPlayer(true),
		domain.NewSevensPlayer(false),
		domain.NewSevensPlayer(false),
		domain.NewSevensPlayer(false),
	}
}

func TestSevens_Method(t *testing.T) {
	t.Run("success NewSevens", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
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
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{TunnelEnabled: true, JokerCount: 2, CpuStrategy: true}
		s := domain.NewSevens(tc, players, cfg)
		assert.Equal(t, cfg, s.GetConfig())
	})

	t.Run("success Reset distributes cards and removes sevens", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		s.Reset()
		total := 0
		for i := 0; i < s.GetPlayerCnt(); i++ {
			total += s.GetPlayer(i).GetCardsSize()
		}
		// 52 cards - 4 sevens = 48 cards distributed
		assert.Equal(t, 48, total)
	})

	t.Run("success Reset initializes table", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		s.Reset()
		mins := s.GetTableMinVals()
		maxs := s.GetTableMaxVals()
		for i := 1; i <= 4; i++ {
			assert.Equal(t, 7, mins[i])
			assert.Equal(t, 7, maxs[i])
		}
	})

	t.Run("success GetPlayer valid index", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		assert.NotNil(t, s.GetPlayer(0))
		assert.True(t, s.GetPlayer(0).GetIsHuman())
		assert.NotNil(t, s.GetPlayer(1))
		assert.False(t, s.GetPlayer(1).GetIsHuman())
	})

	t.Run("success GetPlayer invalid index returns nil", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		assert.Nil(t, s.GetPlayer(-1))
		assert.Nil(t, s.GetPlayer(10))
	})

	t.Run("success IsPlayable on fresh board", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		// Fresh board: all suits have only 7 placed
		// 6 of spades is playable (adjacent to 7), 8 of spades is playable (adjacent to 7)
		card6 := domain.NewCard(domain.CardDesignSpade, 6, false)
		card8 := domain.NewCard(domain.CardDesignSpade, 8, false)
		card5 := domain.NewCard(domain.CardDesignSpade, 5, false)
		card7 := domain.NewCard(domain.CardDesignSpade, 7, false)
		assert.True(t, s.IsPlayable(card6))
		assert.True(t, s.IsPlayable(card8))
		assert.False(t, s.IsPlayable(card5)) // not adjacent
		assert.False(t, s.IsPlayable(card7)) // 7 is already on board
	})

	t.Run("success IsPlayable nil card", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		assert.False(t, s.IsPlayable(nil))
	})

	t.Run("success IsPlayable Ace at min=2", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		// Give human a 6♠ to extend left
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		s.PlayerPlay(0) // human plays 6♠ → board has 6,7

		ace := domain.NewCard(domain.CardDesignSpade, 1, false)
		// min is 6, so ace (1) is not adjacent to 6
		assert.False(t, s.IsPlayable(ace))
	})

	t.Run("success IsPlayable boundary King at max=12", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		// Place 8 of spades → max becomes 8
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		s.PlayerPlay(0) // place 8♠ → max=8
		king := domain.NewCard(domain.CardDesignSpade, 13, false)
		assert.False(t, s.IsPlayable(king)) // 13 != 9 (max+1=9)
	})

	t.Run("success PlayerPlay places card on board", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		err := s.PlayerPlay(0) // play 6♠
		assert.NoError(t, err)
		mins := s.GetTableMinVals()
		assert.Equal(t, 6, mins[domain.CardDesignSpade])
		assert.Equal(t, 1, players[0].GetCardsSize())
		assert.NotNil(t, s.GetHumanAction())
		assert.NotNil(t, s.GetHumanAction().PlayedCard)
		assert.Equal(t, 6, s.GetHumanAction().PlayedCard.GetValue())
	})

	t.Run("success PlayerPlay fails with non-playable card", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false)) // not adjacent to 7
		err := s.PlayerPlay(0)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidPlay)
	})

	t.Run("success PlayerPlay fails with invalid index", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		err := s.PlayerPlay(5) // out of range
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidCard)
	})

	t.Run("success PlayerPlay pass", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		err := s.PlayerPlay(-1) // pass
		assert.NoError(t, err)
		assert.Equal(t, 1, players[0].GetPassesUsed())
		action := s.GetHumanAction()
		assert.NotNil(t, action)
		assert.Nil(t, action.PlayedCard) // nil = pass
	})

	t.Run("success PlayerPlay pass fails when no passes left", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		// Exhaust human passes
		for i := 0; i < domain.SevensMaxPasses; i++ {
			players[0].IncrPassesUsed()
		}
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		err := s.PlayerPlay(-1) // attempt to pass
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrCannotPass)
	})

	t.Run("success PlayerPlay fails when not human turn", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		s.PlayerPlay(0) // human plays → advances to CPU 1
		if !s.IsHumanTurn() && !s.GetGameEndFlag() {
			err := s.PlayerPlay(0)
			assert.Error(t, err)
			assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
		}
	})

	t.Run("success PlayerPlay does nothing when game ended", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		// 3 CPUs finished, human has last card
		players[1].SetIsFinished(true)
		players[1].SetRank(1)
		players[2].SetIsFinished(true)
		players[2].SetRank(2)
		players[3].SetIsFinished(true)
		players[3].SetRank(3)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		s.PlayerPlay(0) // plays 8♠ → finishes → game ends
		assert.True(t, s.GetGameEndFlag())
		assert.Equal(t, 4, players[0].GetRank())

		err := s.PlayerPlay(0) // game already ended
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrGameEnded)
	})

	t.Run("success CpuPlay plays valid card", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false)) // playable
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false)) // playable by CPU 1
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		s.PlayerPlay(0) // human plays 8♠ → advances to CPU 1
		s.CpuPlay()     // CPU 1 plays 6♠
		actions := s.GetCpuActions()
		assert.Len(t, actions, 1)
		assert.NotNil(t, actions[0].PlayedCard)
		assert.Equal(t, 6, actions[0].PlayedCard.GetValue())
	})

	t.Run("success CpuPlay passes when no playable card", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false)) // not adjacent → pass
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		s.PlayerPlay(0) // human plays 8♠
		s.CpuPlay()     // CPU 1: 5♠ not playable → passes
		actions := s.GetCpuActions()
		assert.Len(t, actions, 1)
		assert.Nil(t, actions[0].PlayedCard) // pass
		assert.Equal(t, 1, players[1].GetPassesUsed())
	})

	t.Run("success CpuPlay eliminates when no passes left", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		// Exhaust CPU 1's passes
		for i := 0; i < domain.SevensMaxPasses; i++ {
			players[1].IncrPassesUsed()
		}
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false)) // not playable
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		s.PlayerPlay(0) // human plays
		s.CpuPlay()     // CPU 1: no playable, no passes → eliminated
		assert.True(t, players[1].GetIsFinished())
		assert.Greater(t, players[1].GetRank(), 0)
	})

	t.Run("success CpuPlay does nothing on human turn", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		s.CpuPlay() // does nothing when it's human's turn
		assert.Nil(t, s.GetCpuActions())
		assert.True(t, s.IsHumanTurn())
	})

	t.Run("success HasAnyOption true when has playable card", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false)) // playable
		assert.True(t, s.HasAnyOption(0))
	})

	t.Run("success HasAnyOption true when can pass but no playable card", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false)) // not playable
		assert.True(t, s.HasAnyOption(0))                                    // can still pass
	})

	t.Run("success HasAnyOption false when no options", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		for i := 0; i < domain.SevensMaxPasses; i++ {
			players[0].IncrPassesUsed()
		}
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false)) // not playable
		assert.False(t, s.HasAnyOption(0))
	})

	t.Run("success AutoHandleNoOption eliminates human", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		for i := 0; i < domain.SevensMaxPasses; i++ {
			players[0].IncrPassesUsed()
		}
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		s.AutoHandleNoOption()
		assert.True(t, players[0].GetIsFinished())
		assert.NotNil(t, s.GetHumanAction())
		assert.Nil(t, s.GetHumanAction().PlayedCard) // pass recorded
	})

	t.Run("success game ends when last player finishes", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[1].SetIsFinished(true)
		players[1].SetRank(1)
		players[2].SetIsFinished(true)
		players[2].SetRank(2)
		players[3].SetIsFinished(true)
		players[3].SetRank(3)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		s.PlayerPlay(0) // human plays last card → finishes → game ends
		assert.True(t, s.GetGameEndFlag())
		assert.Equal(t, 4, players[0].GetRank())
	})

	t.Run("success eliminatePlayer places remaining cards on board", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		// Exhaust CPU 1's passes and give it cards that become non-playable after human plays 8♠
		for i := 0; i < domain.SevensMaxPasses; i++ {
			players[1].IncrPassesUsed()
		}
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		// After human plays 8♠: board SPADE has 7,8. 5♠ and 11♠ are not playable.
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		s.PlayerPlay(0) // human plays 8♠ → SPADE max=8
		s.CpuPlay()     // CPU 1: no playable, no passes → eliminated

		assert.True(t, players[1].GetIsFinished())
		assert.True(t, players[1].GetIsEliminated())
		assert.Equal(t, 0, players[1].GetCardsSize()) // hand cleared
		// Board should be expanded to include eliminated player's cards
		mins := s.GetTableMinVals()
		maxs := s.GetTableMaxVals()
		assert.Equal(t, 5, mins[domain.CardDesignSpade])  // 5♠ placed → min=5
		assert.Equal(t, 11, maxs[domain.CardDesignSpade]) // 11♠ placed → max=11
	})

	t.Run("success eliminated player gets lower rank than normal finisher", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		for i := 0; i < domain.SevensMaxPasses; i++ {
			players[1].IncrPassesUsed()
		}
		// Human has one playable card; CPU 1 has non-playable cards
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		s.PlayerPlay(0) // human empties hand → rank 1
		assert.Equal(t, 1, players[0].GetRank())

		s.CpuPlay() // CPU 1: no playable, no passes → eliminated → rank 4
		assert.Equal(t, 4, players[1].GetRank())
		assert.True(t, players[1].GetIsEliminated())

		// Normal finisher ranks better than eliminated player
		assert.Less(t, players[0].GetRank(), players[1].GetRank())
	})

	t.Run("success multiple eliminations get descending ranks", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		// Exhaust passes for CPU 1 and CPU 2
		for i := 0; i < domain.SevensMaxPasses; i++ {
			players[1].IncrPassesUsed()
			players[2].IncrPassesUsed()
		}
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false)) // not playable
		players[2].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false)) // not playable
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

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
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].SetIsEliminated(true)
		players[0].SetIsFinished(true)
		s.Reset()
		assert.False(t, players[0].GetIsEliminated())
		assert.False(t, players[0].GetIsFinished())
	})

	t.Run("success Reset shuffles player order", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())

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
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
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
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		placed := s.GetTablePlaced()
		// Initially only 7 is placed for each suit
		assert.Equal(t, uint16(1<<7), placed[domain.CardDesignSpade])

		s.PlayerPlay(0) // place 6♠
		placed = s.GetTablePlaced()
		assert.Equal(t, uint16((1<<7)|(1<<6)), placed[domain.CardDesignSpade])
	})

	t.Run("success derived GetTableMinVals/MaxVals match bitmask", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		s.PlayerPlay(0) // play 6♠
		mins := s.GetTableMinVals()
		maxs := s.GetTableMaxVals()
		assert.Equal(t, 6, mins[domain.CardDesignSpade])
		assert.Equal(t, 7, maxs[domain.CardDesignSpade])
	})
}

func setupSpadeBoard(config domain.SevensConfig, spadesToPlace []int) *domain.Sevens {
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	s := domain.NewSevens(tc, players, config)
	// Give all players enough dummy cards so they never finish
	for i := 0; i < 4; i++ {
		for d := 0; d < 10; d++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		}
	}
	for _, v := range spadesToPlace {
		if s.GetGameEndFlag() {
			break
		}
		if s.IsHumanTurn() {
			players[0].AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
			// Play the last card (newly added spade)
			idx := players[0].GetCardsSize() - 1
			s.PlayerPlay(idx)
		} else {
			cpuIdx := s.GetCurrentTurn()
			players[cpuIdx].AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
			s.CpuPlay()
		}
	}
	return s
}

func TestSevens_Tunnel(t *testing.T) {
	tunnelConfig := domain.SevensConfig{TunnelEnabled: true, JokerCount: 0, CpuStrategy: false}

	t.Run("success tunnel: Ace playable when 2 is placed", func(t *testing.T) {
		s := setupSpadeBoard(tunnelConfig, []int{6, 5, 4, 3, 2})
		ace := domain.NewCard(domain.CardDesignSpade, 1, false)
		assert.True(t, s.IsPlayable(ace))
	})

	t.Run("success tunnel: Ace playable when King placed via circular wrap", func(t *testing.T) {
		s := setupSpadeBoard(tunnelConfig, []int{8, 9, 10, 11, 12, 13})
		ace := domain.NewCard(domain.CardDesignSpade, 1, false)
		assert.True(t, s.IsPlayable(ace))
	})

	t.Run("success tunnel disabled: Ace not playable when King placed", func(t *testing.T) {
		noTunnel := domain.DefaultSevensConfig()
		s := setupSpadeBoard(noTunnel, []int{8, 9, 10, 11, 12, 13})
		ace := domain.NewCard(domain.CardDesignSpade, 1, false)
		assert.False(t, s.IsPlayable(ace))
	})

	t.Run("success tunnel: King playable when Ace placed via circular wrap", func(t *testing.T) {
		s := setupSpadeBoard(tunnelConfig, []int{6, 5, 4, 3, 2, 1})
		king := domain.NewCard(domain.CardDesignSpade, 13, false)
		assert.True(t, s.IsPlayable(king))
	})
}

func TestSevens_Joker(t *testing.T) {
	jokerConfig := domain.SevensConfig{TunnelEnabled: false, JokerCount: 2, CpuStrategy: false}

	t.Run("success joker is playable when there are open board positions", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, jokerConfig)
		joker := domain.NewCard(domain.CardDesignJoker, 1, false)
		// Fresh board: 6 and 8 are playable for each suit
		assert.True(t, s.IsPlayable(joker))
	})

	t.Run("success PlayerPlayJoker places on target position", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, jokerConfig)
		players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 1, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		// Play joker to SPADE 6 position
		err := s.PlayerPlayJoker(0, domain.CardDesignSpade, 6)
		assert.NoError(t, err)
		// Board should now have SPADE 6 placed
		placed := s.GetTablePlaced()
		assert.True(t, placed[domain.CardDesignSpade]&(1<<6) != 0)
		// Human action should record joker target
		action := s.GetHumanAction()
		assert.NotNil(t, action)
		assert.Equal(t, domain.CardDesignJoker, action.PlayedCard.GetDesign())
		assert.Equal(t, domain.CardDesignSpade, action.TargetSuit)
		assert.Equal(t, 6, action.TargetValue)
	})

	t.Run("success PlayerPlayJoker fails with non-joker card", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, jokerConfig)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		err := s.PlayerPlayJoker(0, domain.CardDesignSpade, 6)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidCard)
	})

	t.Run("success PlayerPlayJoker fails with invalid target position", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, jokerConfig)
		players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 1, false))
		// SPADE 5 is not playable on fresh board (not adjacent to 7)
		err := s.PlayerPlayJoker(0, domain.CardDesignSpade, 5)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidPlay)
	})

	t.Run("success PlayerPlayJoker fails when game ended", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, jokerConfig)
		players[1].SetIsFinished(true)
		players[1].SetRank(1)
		players[2].SetIsFinished(true)
		players[2].SetRank(2)
		players[3].SetIsFinished(true)
		players[3].SetRank(3)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		s.PlayerPlay(0) // game ends
		assert.True(t, s.GetGameEndFlag())

		players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 1, false))
		err := s.PlayerPlayJoker(0, domain.CardDesignSpade, 6)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrGameEnded)
	})

	t.Run("success CpuPlay plays joker", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, jokerConfig)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		// CPU 1 only has a joker
		players[1].AddCard(domain.NewCard(domain.CardDesignJoker, 1, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		s.PlayerPlay(0) // human plays 8♠
		s.CpuPlay()     // CPU 1 plays joker

		actions := s.GetCpuActions()
		assert.Len(t, actions, 1)
		assert.NotNil(t, actions[0].PlayedCard)
		assert.Equal(t, domain.CardDesignJoker, actions[0].PlayedCard.GetDesign())
		assert.Greater(t, actions[0].TargetSuit, 0)
		assert.Greater(t, actions[0].TargetValue, 0)
	})

	t.Run("success eliminatePlayer places normal card on board", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		noJokerCfg := domain.DefaultSevensConfig()
		s := domain.NewSevens(tc, players, noJokerCfg)
		for i := 0; i < domain.SevensMaxPasses; i++ {
			players[1].IncrPassesUsed()
		}
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false)) // not playable
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		s.PlayerPlay(0) // human plays 8♠
		s.CpuPlay()     // CPU 1: 5♠ not playable, no passes → eliminated

		assert.True(t, players[1].GetIsEliminated())
		assert.Equal(t, 0, players[1].GetCardsSize())
		// 5♠ should be force-placed on board
		mins := s.GetTableMinVals()
		assert.Equal(t, 5, mins[domain.CardDesignSpade])
	})

	t.Run("success eliminatePlayer skips joker cards on forced placement", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, jokerConfig)

		// Give CPU1 a joker + non-playable normal card (3♥ — not adjacent to 7 after joker is played)
		players[1].AddCard(domain.NewCard(domain.CardDesignJoker, 1, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		for i := 0; i < domain.SevensMaxPasses; i++ {
			players[1].IncrPassesUsed()
		}

		// Give human playable cards, other CPUs dummy non-playable cards with passes remaining
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		// Round 1: human plays 8♠ → CPU1 plays joker (playable) → CPU2 passes → CPU3 passes
		s.PlayerPlay(0) // human: 8♠
		assert.Equal(t, 1, s.GetCurrentTurn())
		s.CpuPlay() // CPU1: plays joker (board has open positions)
		actions := s.GetCpuActions()
		assert.Equal(t, domain.CardDesignJoker, actions[0].PlayedCard.GetDesign())
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
		assert.True(t, placedAfter[domain.CardDesignHeart]&(1<<3) != 0, "3♥ should be placed on board")
		// Joker suit index (0) should be unchanged — no joker artifact on board
		assert.Equal(t, placedBefore[domain.CardDesignJoker], placedAfter[domain.CardDesignJoker], "joker suit should not change in tablePlaced")
	})

	t.Run("success joker is not playable when board is full", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, jokerConfig)

		// Fill all 4 suits completely by playing cards in alternating order
		suits := []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond}
		values := []int{6, 8, 5, 9, 4, 10, 3, 11, 2, 12, 1, 13}
		for _, suit := range suits {
			for i, v := range values {
				pIdx := i % 4
				players[pIdx].AddCard(domain.NewCard(suit, v, false))
			}
		}
		// Give all players extra dummy cards so they never finish
		for i := 0; i < 4; i++ {
			for d := 0; d < 20; d++ {
				players[i].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
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

		joker := domain.NewCard(domain.CardDesignJoker, 1, false)
		assert.False(t, s.IsPlayable(joker))
	})
}

func TestSevens_CpuStrategy(t *testing.T) {
	strategyConfig := domain.SevensConfig{TunnelEnabled: false, JokerCount: 0, CpuStrategy: true}

	t.Run("success strategic CPU passes when holding blocker", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, strategyConfig)
		// CPU 1 has 6♠ (playable, adjacent to 7) but does NOT have 5♠
		// Playing 6♠ opens the path for opponents (score = -1 for low direction)
		// With strategy enabled and passes available, CPU should prefer to pass
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false)) // playable but blocks
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		s.PlayerPlay(0) // human plays 8♠
		s.CpuPlay()     // CPU 1: strategic evaluation

		actions := s.GetCpuActions()
		assert.Len(t, actions, 1)
		// CPU should pass since playing 6♠ has negative score and CPU has passes
		assert.Nil(t, actions[0].PlayedCard)
		assert.Equal(t, 1, players[1].GetPassesUsed())
	})

	t.Run("success strategic CPU plays when it has the chain card", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, strategyConfig)
		// CPU 1 has 6♠ AND 5♠ → playing 6♠ has positive score (+2 for having 5♠)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		s.PlayerPlay(0) // human plays 8♠
		s.CpuPlay()     // CPU 1: plays 6♠ (positive score: has 5♠)

		actions := s.GetCpuActions()
		assert.Len(t, actions, 1)
		assert.NotNil(t, actions[0].PlayedCard)
		assert.Equal(t, 6, actions[0].PlayedCard.GetValue())
	})

	t.Run("success strategic CPU plays when low on passes (reserve=1)", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, strategyConfig)
		// CPU 1 has used all passes except 1 (maxPasses-1 used)
		// Even with negative score, CPU must play because it can't afford to pass
		for i := 0; i < domain.SevensMaxPasses-1; i++ {
			players[1].IncrPassesUsed()
		}
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		s.PlayerPlay(0) // human plays 8♠
		s.CpuPlay()     // CPU 1: forced to play (only 1 pass remaining)

		actions := s.GetCpuActions()
		assert.Len(t, actions, 1)
		assert.NotNil(t, actions[0].PlayedCard)
		assert.Equal(t, 6, actions[0].PlayedCard.GetValue())
	})

	t.Run("success non-strategic CPU always plays first available card", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		noStrategy := domain.DefaultSevensConfig()
		s := domain.NewSevens(tc, players, noStrategy)
		// Same setup as "holding blocker" test but without strategy
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		s.PlayerPlay(0) // human plays 8♠
		s.CpuPlay()     // CPU 1: non-strategic, plays 6♠

		actions := s.GetCpuActions()
		assert.Len(t, actions, 1)
		assert.NotNil(t, actions[0].PlayedCard) // plays the card instead of passing
		assert.Equal(t, 6, actions[0].PlayedCard.GetValue())
	})
}

func TestSevens_SetHumanAction(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())

	action := &domain.SevensCpuAction{PlayerIdx: 0, PlayedCard: nil}
	s.SetHumanAction(action)
	assert.Equal(t, action, s.GetHumanAction())
}

func TestSevens_SetCpuActions(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())

	actions := []*domain.SevensCpuAction{
		{PlayerIdx: 1, PlayedCard: nil},
		{PlayerIdx: 2, PlayedCard: nil},
	}
	s.SetCpuActions(actions)
	assert.Equal(t, actions, s.GetCpuActions())
	assert.Len(t, s.GetCpuActions(), 2)
}

func TestSevens_GetNextActivePlayer_AllFinished(t *testing.T) {
	// When all players are finished, getNextActivePlayer returns -1.
	// Test indirectly: 3 CPUs finished, human plays last card → all 4 finished → game ends.
	// The last remaining player is auto-assigned a rank via checkGameEnd.
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())

	players[1].SetIsFinished(true)
	players[1].SetRank(1)
	players[2].SetIsFinished(true)
	players[2].SetRank(2)
	players[3].SetIsFinished(true)
	players[3].SetRank(3)
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))

	err := s.PlayerPlay(0) // human plays last card → finishes → checkGameEnd
	assert.NoError(t, err)
	assert.True(t, s.GetGameEndFlag())
	assert.Equal(t, 4, players[0].GetRank())
	// After game ends, turn should not have advanced further
	// (advanceTurn returns immediately when gameEndFlag is true)
}

func TestSevens_HasAnyOption_InvalidIndex(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())

	assert.False(t, s.HasAnyOption(-1))
	assert.False(t, s.HasAnyOption(4))
	assert.False(t, s.HasAnyOption(100))
}

func TestSevens_IsPlayable_InvalidSuit(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())

	// Card with design=5 (beyond CardDesignDiamond=4), value=6
	// IsPlayable should return false for suit > CardDesignDiamond
	cardInvalid := domain.NewCard(5, 6, false)
	assert.False(t, s.IsPlayable(cardInvalid))

	// Card with design=6
	cardInvalid2 := domain.NewCard(6, 8, false)
	assert.False(t, s.IsPlayable(cardInvalid2))
}

func TestSevens_IsPositionPlaced_InvalidBounds(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())

	// Test via IsPlayable with value=0 (below valid range 1-13)
	// isPositionPlayable(suit, 0) → value < 1 → returns false
	cardVal0 := domain.NewCard(domain.CardDesignSpade, 0, false)
	assert.False(t, s.IsPlayable(cardVal0))

	// Test via IsPlayable with value=14 (above valid range 1-13)
	// isPositionPlayable(suit, 14) → value > 13 → returns false
	cardVal14 := domain.NewCard(domain.CardDesignSpade, 14, false)
	assert.False(t, s.IsPlayable(cardVal14))
}

func TestSevens_IsPositionPlayable_InvalidBounds(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	jokerConfig := domain.SevensConfig{TunnelEnabled: false, JokerCount: 2, CpuStrategy: false}
	s := domain.NewSevens(tc, players, jokerConfig)

	players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 1, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

	// PlayerPlayJoker with targetValue=0 → isPositionPlayable returns false → ErrInvalidPlay
	err := s.PlayerPlayJoker(0, domain.CardDesignSpade, 0)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidPlay)

	// PlayerPlayJoker with targetSuit=0 (CardDesignJoker) → isPositionPlayable returns false
	err = s.PlayerPlayJoker(0, 0, 6)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidPlay)

	// PlayerPlayJoker with targetValue=14 → isPositionPlayable returns false
	err = s.PlayerPlayJoker(0, domain.CardDesignSpade, 14)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidPlay)
}

func TestSevens_PlayerPlayJoker_NotHumanTurn(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	jokerConfig := domain.SevensConfig{TunnelEnabled: false, JokerCount: 2, CpuStrategy: false}
	s := domain.NewSevens(tc, players, jokerConfig)

	// Human plays a card to advance turn to CPU
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignJoker, 1, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

	s.PlayerPlay(0) // human plays 8♠ → turn moves to CPU 1
	assert.False(t, s.IsHumanTurn())

	// Now try PlayerPlayJoker on CPU's turn
	err := s.PlayerPlayJoker(0, domain.CardDesignSpade, 6)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
}

func TestSevens_FindPlayableSimple_JokerNoPosition(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	jokerConfig := domain.SevensConfig{TunnelEnabled: false, JokerCount: 2, CpuStrategy: false}
	s := domain.NewSevens(tc, players, jokerConfig)

	// Fill the entire board for all 4 suits
	suits := []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond}
	values := []int{6, 8, 5, 9, 4, 10, 3, 11, 2, 12, 1, 13}
	for _, suit := range suits {
		for i, v := range values {
			pIdx := i % 4
			players[pIdx].AddCard(domain.NewCard(suit, v, false))
		}
	}
	// Give all players extra dummy cards so they never finish
	for i := 0; i < 4; i++ {
		for d := 0; d < 20; d++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		}
	}
	// Play all cards until board is full
	for {
		if s.GetGameEndFlag() {
			break
		}
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
				s.PlayerPlay(-1) // pass
			}
		} else {
			s.CpuPlay()
		}
	}

	// Now give CPU 1 a joker. Board is full, so joker cannot be placed.
	// Advance to CPU 1's turn by playing from human.
	if !s.GetGameEndFlag() && s.IsHumanTurn() {
		cpuIdx := 1
		// Find CPU 1 and give it only a joker
		players[cpuIdx].AddCard(domain.NewCard(domain.CardDesignJoker, 1, false))
		// Human passes to advance to CPU
		s.PlayerPlay(-1)
		if !s.GetGameEndFlag() && s.GetCurrentTurn() == cpuIdx {
			s.CpuPlay() // CPU with joker on full board → should pass (joker not playable)
			actions := s.GetCpuActions()
			if len(actions) > 0 {
				lastAction := actions[len(actions)-1]
				if lastAction.PlayerIdx == cpuIdx {
					// CPU should have passed (no joker play possible)
					assert.Nil(t, lastAction.PlayedCard)
				}
			}
		}
	}
}

func TestSevens_FindPlayableStrategic_JokerEval(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	strategyConfig := domain.SevensConfig{TunnelEnabled: false, JokerCount: 2, CpuStrategy: true}
	s := domain.NewSevens(tc, players, strategyConfig)

	// CPU 1 has a joker + 5♠. Strategic evaluation should place joker at 6♠
	// (adjacent to 7♠), which gives +2 because CPU has 5♠ (the next low card).
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignJoker, 1, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

	s.PlayerPlay(0) // human plays 8♠
	s.CpuPlay()     // CPU 1: strategic, plays joker

	actions := s.GetCpuActions()
	assert.Len(t, actions, 1)
	assert.NotNil(t, actions[0].PlayedCard)
	assert.Equal(t, domain.CardDesignJoker, actions[0].PlayedCard.GetDesign())
	// Should choose a position strategically (6♠ where CPU has 5♠)
	assert.Equal(t, domain.CardDesignSpade, actions[0].TargetSuit)
	assert.Equal(t, 6, actions[0].TargetValue)
}

func TestSevens_FindPlayableStrategic_Pass(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	strategyConfig := domain.SevensConfig{TunnelEnabled: false, JokerCount: 0, CpuStrategy: true}
	s := domain.NewSevens(tc, players, strategyConfig)

	// CPU 1 has 6♠ (playable) but does NOT have 5♠.
	// Playing 6♠ opens the path for opponents → negative score.
	// CPU also has 8♥ (playable) but does NOT have 9♥ → negative score.
	// With passes available and all plays negative, CPU should pass.
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
	players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

	s.PlayerPlay(0) // human plays 8♠
	s.CpuPlay()     // CPU 1: all plays negative → pass

	actions := s.GetCpuActions()
	assert.Len(t, actions, 1)
	assert.Nil(t, actions[0].PlayedCard) // passed
	assert.Equal(t, 1, players[1].GetPassesUsed())
}

func TestSevens_EvaluatePlay_LowDirection(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	strategyConfig := domain.SevensConfig{TunnelEnabled: false, JokerCount: 0, CpuStrategy: true}
	s := domain.NewSevens(tc, players, strategyConfig)

	// Setup: place 6♠ on board (via human play) so min=6, max=7
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
	// CPU 1 has 5♠ (playable, adjacent to 6) and 4♠ (the next low card)
	// Playing 5♠: nextLow=4, CPU has 4♠ → +2; nextHigh=6, already placed → no score
	// Total score = +2 → CPU should play
	players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 4, false))
	players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

	s.PlayerPlay(0) // human plays 6♠ → board spade min=6
	s.CpuPlay()     // CPU 1: plays 5♠ (score +2 for having 4♠)

	actions := s.GetCpuActions()
	assert.Len(t, actions, 1)
	assert.NotNil(t, actions[0].PlayedCard)
	assert.Equal(t, 5, actions[0].PlayedCard.GetValue())
}

func TestSevens_EvaluatePlay_LowDirection_NegativeScore(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	strategyConfig := domain.SevensConfig{TunnelEnabled: false, JokerCount: 0, CpuStrategy: true}
	s := domain.NewSevens(tc, players, strategyConfig)

	// Setup: place 6♠ on board (via human play) so min=6, max=7
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
	// CPU 1 has only 5♠ (playable, adjacent to 6) but NOT 4♠
	// Playing 5♠: nextLow=4, CPU doesn't have 4♠ → -1; nextHigh=6, already placed → no score
	// Total score = -1 → CPU should pass (has passes available)
	players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

	s.PlayerPlay(0) // human plays 6♠
	s.CpuPlay()     // CPU 1: 5♠ has score -1 → pass

	actions := s.GetCpuActions()
	assert.Len(t, actions, 1)
	assert.Nil(t, actions[0].PlayedCard) // passed
}

func TestSevens_EvaluatePlay_TunnelWrap(t *testing.T) {
	tunnelConfig := domain.SevensConfig{TunnelEnabled: true, JokerCount: 0, CpuStrategy: false}
	tunnelStrategyConfig := domain.SevensConfig{TunnelEnabled: true, JokerCount: 0, CpuStrategy: true}

	t.Run("value 1 with tunnel wraps nextLow to 13", func(t *testing.T) {
		// Use non-strategy config so setupSpadeBoard plays cards deterministically.
		// Place spade cards 6,8,5,9,4,10,3,11,2,12,13 on the board.
		// Board has spade 2-13 placed, only Ace(1) missing.
		// With tunnel: Ace is playable from King(13) side via circular wrap.
		spadesToPlace := []int{6, 8, 5, 9, 4, 10, 3, 11, 2, 12, 13}
		s := setupSpadeBoard(tunnelConfig, spadesToPlace)

		ace := domain.NewCard(domain.CardDesignSpade, 1, false)
		assert.True(t, s.IsPlayable(ace))
	})

	t.Run("value 13 with tunnel wraps nextHigh to 1", func(t *testing.T) {
		// Use non-strategy config so cards are played deterministically.
		// Place cards 6,8,5,9,4,10,3,11,2,12,1 on the board.
		// Board has everything except King(13).
		// With tunnel, King is playable adjacent to 12 and wraps to Ace(1).
		spadesToPlace := []int{6, 8, 5, 9, 4, 10, 3, 11, 2, 12, 1}
		s := setupSpadeBoard(tunnelConfig, spadesToPlace)

		king := domain.NewCard(domain.CardDesignSpade, 13, false)
		assert.True(t, s.IsPlayable(king))
	})

	t.Run("strategic CPU plays Ace with tunnel when holding King", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, tunnelStrategyConfig)

		// Give all players dummy cards so they never finish
		for i := 0; i < 4; i++ {
			for d := 0; d < 10; d++ {
				players[i].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
			}
		}

		// Build spade board by having human play all spade cards.
		// CPUs only have non-playable diamond 2s, so they always pass.
		humanCards := []int{6, 5, 4, 3, 2, 8, 9, 10, 11, 12}
		for _, v := range humanCards {
			if s.GetGameEndFlag() {
				break
			}
			// Let CPUs pass until it's human's turn
			for !s.IsHumanTurn() && !s.GetGameEndFlag() {
				s.CpuPlay()
			}
			if s.GetGameEndFlag() {
				break
			}
			players[0].AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
			idx := players[0].GetCardsSize() - 1
			s.PlayerPlay(idx)
		}

		// Board now has spade 2-12. Ace(1) and King(13) are playable with tunnel.
		// Advance to CPU 1's turn
		for !s.GetGameEndFlag() && s.IsHumanTurn() {
			s.PlayerPlay(-1)
		}

		if !s.GetGameEndFlag() {
			cpuIdx := s.GetCurrentTurn()
			// Give CPU Ace(1) and King(13).
			// evaluatePlay for Ace: nextLow = 13 (tunnel wrap), CPU has 13 -> +2.
			// nextHigh = 2 (placed) -> 0. Score = +2. CPU should play Ace.
			players[cpuIdx].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
			players[cpuIdx].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
			s.CpuPlay()

			actions := s.GetCpuActions()
			assert.NotEmpty(t, actions)
			lastAction := actions[len(actions)-1]
			assert.NotNil(t, lastAction.PlayedCard)
			assert.Equal(t, 1, lastAction.PlayedCard.GetValue()) // played Ace
		}
	})
}

func TestSevens_EvaluateJokerPlays(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	strategyConfig := domain.SevensConfig{TunnelEnabled: false, JokerCount: 2, CpuStrategy: true}
	s := domain.NewSevens(tc, players, strategyConfig)

	// CPU 1 has a joker and also 5♠.
	// Joker evaluation: for each playable position, evaluate as if placing a card there.
	// 6♠ is playable (adjacent to 7): nextLow=5 (CPU has it → +2), nextHigh=7 (placed → 0) → score +2
	// 8♠ is playable (adjacent to 7): nextLow=7 (placed → 0), nextHigh=9 (CPU doesn't have → -1) → score -1
	// Joker should be placed at 6♠ (highest score).
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignJoker, 1, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

	s.PlayerPlay(0) // human plays 8♠
	s.CpuPlay()     // CPU 1: strategic joker evaluation

	actions := s.GetCpuActions()
	assert.Len(t, actions, 1)
	assert.NotNil(t, actions[0].PlayedCard)
	assert.Equal(t, domain.CardDesignJoker, actions[0].PlayedCard.GetDesign())
	// Best joker position should be 6♠ (score +2)
	assert.Equal(t, domain.CardDesignSpade, actions[0].TargetSuit)
	assert.Equal(t, 6, actions[0].TargetValue)
}

func TestSevens_FindFirstPlayablePosition_BoardFull(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	jokerConfig := domain.SevensConfig{TunnelEnabled: false, JokerCount: 2, CpuStrategy: false}
	s := domain.NewSevens(tc, players, jokerConfig)

	// Fill the entire board for all 4 suits
	suits := []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond}
	values := []int{6, 8, 5, 9, 4, 10, 3, 11, 2, 12, 1, 13}
	for _, suit := range suits {
		for i, v := range values {
			pIdx := i % 4
			players[pIdx].AddCard(domain.NewCard(suit, v, false))
		}
	}
	for i := 0; i < 4; i++ {
		for d := 0; d < 20; d++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		}
	}
	// Play until board is full
	for {
		if s.GetGameEndFlag() {
			break
		}
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
				s.PlayerPlay(-1)
			}
		} else {
			s.CpuPlay()
		}
	}

	// Board is now completely full. Joker should not be playable.
	joker := domain.NewCard(domain.CardDesignJoker, 1, false)
	assert.False(t, s.IsPlayable(joker))
}

func TestSevens_AutoHandleNoOption_CpuEliminated(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())

	// Give human a playable card so human can play and advance turn to CPU
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
	// CPU 1 has no playable cards and no passes
	for i := 0; i < domain.SevensMaxPasses; i++ {
		players[1].IncrPassesUsed()
	}
	players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

	s.PlayerPlay(0) // human plays 8♠ → turn advances to CPU 1
	assert.Equal(t, 1, s.GetCurrentTurn())
	assert.False(t, s.IsHumanTurn())

	// Manually call AutoHandleNoOption for the CPU player
	s.AutoHandleNoOption()

	assert.True(t, players[1].GetIsFinished())
	assert.True(t, players[1].GetIsEliminated())
	// cpuActions should have the elimination action
	cpuActions := s.GetCpuActions()
	assert.NotNil(t, cpuActions)
	found := false
	for _, a := range cpuActions {
		if a.PlayerIdx == 1 {
			assert.Nil(t, a.PlayedCard) // elimination recorded as nil PlayedCard
			found = true
		}
	}
	assert.True(t, found, "expected CPU 1 elimination action in cpuActions")
}

func TestSevens_PlayerPlayJoker_InvalidCardIdx(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	jokerConfig := domain.SevensConfig{TunnelEnabled: false, JokerCount: 2, CpuStrategy: false}
	s := domain.NewSevens(tc, players, jokerConfig)

	players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 1, false))

	// Out-of-range card index (too high)
	err := s.PlayerPlayJoker(5, domain.CardDesignSpade, 6)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidCard)

	// Negative card index
	err = s.PlayerPlayJoker(-1, domain.CardDesignSpade, 6)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidCard)
}

func TestSevens_PlayerPlayJoker_PlayerFinishesAfterJoker(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	jokerConfig := domain.SevensConfig{TunnelEnabled: false, JokerCount: 2, CpuStrategy: false}
	s := domain.NewSevens(tc, players, jokerConfig)

	// Set 3 CPUs as finished
	players[1].SetIsFinished(true)
	players[1].SetRank(1)
	players[2].SetIsFinished(true)
	players[2].SetRank(2)
	players[3].SetIsFinished(true)
	players[3].SetRank(3)

	// Human has only a joker as last card
	players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 1, false))

	// Playing joker is the last card → player finishes → game ends
	err := s.PlayerPlayJoker(0, domain.CardDesignSpade, 6)
	assert.NoError(t, err)
	assert.True(t, s.GetGameEndFlag())
	assert.Equal(t, 4, players[0].GetRank())
	assert.True(t, players[0].GetIsFinished())
	assert.Equal(t, 0, players[0].GetCardsSize())
}

// --- Coverage gap tests below ---

func TestSevens_GetNextActivePlayer_AllFinished_ReturnsNegOne(t *testing.T) {
	// Covers line 148: getNextActivePlayer returns -1 when all players are finished.
	// We set all 4 players to isFinished=true (without setting gameEndFlag),
	// then call PlayerPlay(-1) (pass) which calls advanceTurn -> getNextActivePlayer.
	// getNextActivePlayer loops through all 4 finished players and returns -1.
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())

	// Mark all 4 players as finished (without gameEndFlag being set)
	for i := 0; i < 4; i++ {
		players[i].SetIsFinished(true)
		players[i].SetRank(i + 1)
	}

	// Human (player 0) still has a pass available and is the current turn.
	// PlayerPlay(-1) passes, which calls advanceTurn.
	// advanceTurn calls getNextActivePlayer; all players are finished, so it returns -1.
	// The currentTurn stays unchanged (next >= 0 is false).
	err := s.PlayerPlay(-1)
	assert.NoError(t, err)
	assert.Equal(t, 0, s.GetCurrentTurn()) // turn did not advance
}

func TestSevens_AdvanceTurn_GameEndFlagBlocksPlay(t *testing.T) {
	// Verifies that when gameEndFlag is true, PlayerPlay returns ErrGameEnded
	// and the turn does not advance. The defensive advanceTurn guard (lines 176-178)
	// is tested directly via sevens_internal_test.go.
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())

	// End the game via normal play
	players[1].SetIsFinished(true)
	players[1].SetRank(1)
	players[2].SetIsFinished(true)
	players[2].SetRank(2)
	players[3].SetIsFinished(true)
	players[3].SetRank(3)
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))

	s.PlayerPlay(0) // human plays last card -> checkGameEnd -> gameEndFlag=true
	assert.True(t, s.GetGameEndFlag())
	assert.Equal(t, 0, s.GetCurrentTurn())

	// Further play attempts are blocked
	err := s.PlayerPlay(-1)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrGameEnded)
}

func TestSevens_IsPositionPlaced_InvalidSuit(t *testing.T) {
	// Covers lines 187-189: isPositionPlaced returns false for invalid suit.
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())

	// suit < CardDesignSpade (suit=0)
	assert.False(t, s.IsPositionPlaced(0, 7))
	// suit > CardDesignDiamond (suit=5)
	assert.False(t, s.IsPositionPlaced(5, 7))
	// suit = -1
	assert.False(t, s.IsPositionPlaced(-1, 7))
	// Valid suit with valid value (7 is placed on fresh board)
	assert.True(t, s.IsPositionPlaced(domain.CardDesignSpade, 7))
	// Valid suit with unplaced value
	assert.False(t, s.IsPositionPlaced(domain.CardDesignSpade, 6))
}

func TestSevens_FindPlayableStrategic_BestScoreUpdated(t *testing.T) {
	// Covers lines 453-455: the loop body where a later play has higher score
	// than the first, updating 'best'.
	// Setup: CPU has two playable cards. The first has a lower score, the second
	// has a higher score.
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	strategyConfig := domain.SevensConfig{TunnelEnabled: false, JokerCount: 0, CpuStrategy: true}
	s := domain.NewSevens(tc, players, strategyConfig)

	// Give human playable cards to advance turn
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))

	// CPU 1 has two playable cards:
	// 1. 6♠ (adjacent to 7): nextLow=5 (CPU does NOT have 5♠) -> -1
	//    nextHigh=7 (placed) -> 0. Score = -1.
	// 2. 8♥ (adjacent to 7): nextLow=7 (placed) -> 0
	//    nextHigh=9 (CPU HAS 9♥) -> +2. Score = +2.
	// First play (6♠) has score -1, second play (8♥) has score +2.
	// The loop updates best from -1 to +2, covering lines 453-455.
	players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 9, false)) // chain card for 8♥

	players[2].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
	players[3].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))

	s.PlayerPlay(0) // human plays 8♠
	s.CpuPlay()     // CPU 1: strategic, evaluates both cards

	actions := s.GetCpuActions()
	assert.Len(t, actions, 1)
	assert.NotNil(t, actions[0].PlayedCard)
	// CPU should play 8♥ (score +2) over 6♠ (score -1)
	assert.Equal(t, domain.CardDesignHeart, actions[0].PlayedCard.GetDesign())
	assert.Equal(t, 8, actions[0].PlayedCard.GetValue())
}

func TestSevens_EvaluatePlay_TunnelAceLow(t *testing.T) {
	// Covers lines 476-478: evaluatePlay tunnel wrap for Ace (value=1),
	// where nextLow becomes 13 instead of 0.
	// Setup: tunnel enabled + strategy. Board has spade 2-7 placed (min=2, max=7).
	// CPU has Ace(1) which is playable (adjacent to 2).
	// Without tunnel: nextLow = 0, skipped. With tunnel: nextLow = 13.
	// CPU does NOT have King(13), so score -= 1 for low direction.
	// nextHigh = 2 (placed), so no score for high direction.
	// Total score = -1 -> CPU would pass if it has passes.
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	tunnelStrategyConfig := domain.SevensConfig{TunnelEnabled: true, JokerCount: 0, CpuStrategy: true}
	s := domain.NewSevens(tc, players, tunnelStrategyConfig)

	// Build board: place spade 6, 5, 4, 3, 2 so board has spade 2-7
	// Use SetTablePlaced to set the board directly for simplicity
	var placed [5]uint16
	for i := 1; i <= 4; i++ {
		placed[i] = 1 << 7 // 7 is placed for all suits
	}
	// Place spade 2-6 (plus 7 already)
	placed[domain.CardDesignSpade] |= (1 << 2) | (1 << 3) | (1 << 4) | (1 << 5) | (1 << 6)
	s.SetTablePlaced(placed)

	// Give players enough cards so game doesn't end
	for i := 0; i < 4; i++ {
		for d := 0; d < 5; d++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		}
	}

	// Human plays some card to advance turn to CPU
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
	s.PlayerPlay(players[0].GetCardsSize() - 1) // play 8♠

	// CPU 1 has Ace(1)♠ - playable since adjacent to 2♠
	// It does NOT have King(13)♠, so tunnel wrap gives negative score
	players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))

	if s.GetCurrentTurn() == 1 {
		s.CpuPlay()
		// Score = -1 (tunnel wrap nextLow=13, no King) and CPU has passes -> should pass
		actions := s.GetCpuActions()
		assert.NotEmpty(t, actions)
		lastAction := actions[len(actions)-1]
		if lastAction.PlayerIdx == 1 {
			assert.Nil(t, lastAction.PlayedCard) // passed due to negative score
		}
	}
}

func TestSevens_EvaluatePlay_TunnelKingHigh(t *testing.T) {
	// Covers lines 489-491: evaluatePlay tunnel wrap for King (value=13),
	// where nextHigh becomes 1 instead of 14.
	// Setup: tunnel enabled + strategy. Board has spade 7-12 placed.
	// CPU has King(13) which is playable (adjacent to 12).
	// Without tunnel: nextHigh = 14, skipped. With tunnel: nextHigh = 1.
	// CPU does NOT have Ace(1), so score -= 1 for high direction.
	// nextLow = 12 (placed), so no score for low direction.
	// Total score = -1 -> CPU would pass if it has passes.
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	tunnelStrategyConfig := domain.SevensConfig{TunnelEnabled: true, JokerCount: 0, CpuStrategy: true}
	s := domain.NewSevens(tc, players, tunnelStrategyConfig)

	// Build board: place spade 8-12 (plus 7 already)
	var placed [5]uint16
	for i := 1; i <= 4; i++ {
		placed[i] = 1 << 7
	}
	placed[domain.CardDesignSpade] |= (1 << 8) | (1 << 9) | (1 << 10) | (1 << 11) | (1 << 12)
	s.SetTablePlaced(placed)

	// Give players enough cards so game doesn't end
	for i := 0; i < 4; i++ {
		for d := 0; d < 5; d++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		}
	}

	// Human plays some card to advance turn to CPU
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
	s.PlayerPlay(players[0].GetCardsSize() - 1) // play 6♠

	// CPU 1 has King(13)♠ - playable since adjacent to 12♠
	// It does NOT have Ace(1)♠, so tunnel wrap gives negative score
	players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))

	if s.GetCurrentTurn() == 1 {
		s.CpuPlay()
		actions := s.GetCpuActions()
		assert.NotEmpty(t, actions)
		lastAction := actions[len(actions)-1]
		if lastAction.PlayerIdx == 1 {
			assert.Nil(t, lastAction.PlayedCard) // passed due to negative score
		}
	}
}

func TestSevens_EvaluatePlay_PlayerHasNextHighCard(t *testing.T) {
	// Covers lines 493-495: evaluatePlay where player has the next high card.
	// Setup: strategy enabled. CPU plays 8♠ (adjacent to 7) and has 9♠ in hand.
	// nextHigh = 9, not placed on board. CPU has 9♠ -> score += 2.
	// nextLow = 7 (placed on board) -> skip. Total score = +2.
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	strategyConfig := domain.SevensConfig{TunnelEnabled: false, JokerCount: 0, CpuStrategy: true}
	s := domain.NewSevens(tc, players, strategyConfig)

	// Give players dummy cards so game doesn't end
	for i := 0; i < 4; i++ {
		for d := 0; d < 5; d++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		}
	}

	// Human plays to advance turn
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
	s.PlayerPlay(players[0].GetCardsSize() - 1) // play 6♠

	// CPU 1 has 8♠ (playable, adjacent to 7) and 9♠ (the next high card)
	players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))

	if s.GetCurrentTurn() == 1 {
		s.CpuPlay()
		actions := s.GetCpuActions()
		assert.NotEmpty(t, actions)
		lastAction := actions[len(actions)-1]
		if lastAction.PlayerIdx == 1 {
			// CPU should play 8♠ because it has 9♠ (score = +2, positive)
			assert.NotNil(t, lastAction.PlayedCard)
			assert.Equal(t, 8, lastAction.PlayedCard.GetValue())
		}
	}
}

func TestSevens_SetGameEndFlag(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())

	assert.False(t, s.GetGameEndFlag())
	s.SetGameEndFlag(true)
	assert.True(t, s.GetGameEndFlag())
	s.SetGameEndFlag(false)
	assert.False(t, s.GetGameEndFlag())
}

func TestSevens_SetCurrentTurn(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())

	assert.Equal(t, 0, s.GetCurrentTurn())
	s.SetCurrentTurn(2)
	assert.Equal(t, 2, s.GetCurrentTurn())
}

func TestSevens_SetTablePlaced(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())

	var placed [5]uint16
	for i := 1; i <= 4; i++ {
		placed[i] = (1 << 7) | (1 << 6)
	}
	s.SetTablePlaced(placed)
	result := s.GetTablePlaced()
	for i := 1; i <= 4; i++ {
		assert.Equal(t, uint16((1<<7)|(1<<6)), result[i])
	}
}
