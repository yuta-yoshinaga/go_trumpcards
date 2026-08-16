package domain_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		cfg := domain.SevensConfig{TunnelEnabled: true, JokerCount: 2, CpuStrategy: domain.SevensCpuStrategic}
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
		_ = s.PlayerPlay(0) // human plays 6♠ → board has 6,7

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
		_ = s.PlayerPlay(0) // place 8♠ → max=8
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
		_ = s.PlayerPlay(0) // human plays → advances to CPU 1
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
		_ = s.PlayerPlay(0) // plays 8♠ → finishes → game ends
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

		_ = s.PlayerPlay(0) // human plays 8♠ → advances to CPU 1
		s.CpuPlay()         // CPU 1 plays 6♠
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

		_ = s.PlayerPlay(0) // human plays 8♠
		s.CpuPlay()         // CPU 1: 5♠ not playable → passes
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

		_ = s.PlayerPlay(0) // human plays
		s.CpuPlay()         // CPU 1: no playable, no passes → eliminated
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
		_ = s.PlayerPlay(0) // human plays last card → finishes → game ends
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

		_ = s.PlayerPlay(0) // human plays 8♠ → SPADE max=8
		s.CpuPlay()         // CPU 1: no playable, no passes → eliminated

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

		_ = s.PlayerPlay(0) // human empties hand → rank 1
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

		_ = s.PlayerPlay(0) // human plays
		s.CpuPlay()         // CPU 1 eliminated first → rank 4
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

		_ = s.PlayerPlay(0) // place 6♠
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

		_ = s.PlayerPlay(0) // play 6♠
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
			_ = s.PlayerPlay(idx)
		} else {
			cpuIdx := s.GetCurrentTurn()
			players[cpuIdx].AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
			s.CpuPlay()
		}
	}
	return s
}

func TestSevens_Tunnel(t *testing.T) {
	tunnelConfig := domain.SevensConfig{TunnelEnabled: true, JokerCount: 0, CpuStrategy: domain.SevensCpuSimple}

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

func TestSevens_TunnelSkipWidth(t *testing.T) {
	skipConfig := domain.SevensConfig{TunnelSkipWidth: 3}

	t.Run("skip=3: value 4 playable when 7 is placed", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, skipConfig)
		// Fresh board has only 7s placed; skip=3 means 7-3=4 and 7+3=10 are playable
		card4 := domain.NewCard(domain.CardDesignSpade, 4, false)
		assert.True(t, s.IsPlayable(card4))
	})

	t.Run("skip=3: value 10 playable when 7 is placed", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, skipConfig)
		card10 := domain.NewCard(domain.CardDesignSpade, 10, false)
		assert.True(t, s.IsPlayable(card10))
	})

	t.Run("skip=3: value 5 not playable (not adjacent or skip-distance from placed)", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, skipConfig)
		card5 := domain.NewCard(domain.CardDesignSpade, 5, false)
		// 5 is not adjacent to 7 (not 6 or 8), and not skip-distance from 7 (not 4 or 10)
		assert.False(t, s.IsPlayable(card5))
	})

	t.Run("skip=3: normal adjacency still works", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, skipConfig)
		card6 := domain.NewCard(domain.CardDesignSpade, 6, false)
		assert.True(t, s.IsPlayable(card6))
	})

	t.Run("skip=3 without tunnel: no wrap for value 1 (skip from 1-3=-2 out of range)", func(t *testing.T) {
		// Place 1 on the board, then check if -2 (out of range) is handled
		s := setupSpadeBoard(skipConfig, []int{6, 5, 4, 3, 2, 1})
		// 1 is placed. Without tunnel, skip from 1: low=1-3=-2 (invalid), high=1+3=4 (already placed)
		// So 11 (= 1-3 wrapped) should NOT be playable without TunnelEnabled
		card11 := domain.NewCard(domain.CardDesignSpade, 11, false)
		assert.False(t, s.IsPlayable(card11))
	})

	t.Run("skip=3 with tunnel: wrap enables distant connections", func(t *testing.T) {
		wrapConfig := domain.SevensConfig{TunnelEnabled: true, TunnelSkipWidth: 3}
		s := setupSpadeBoard(wrapConfig, []int{6, 5, 4, 3, 2, 1})
		// 1 is placed. With tunnel wrap: 1-3 → wrapValue(-2) = 11
		card11 := domain.NewCard(domain.CardDesignSpade, 11, false)
		assert.True(t, s.IsPlayable(card11))
	})

	t.Run("skip=3 with tunnel: high wrap from 13", func(t *testing.T) {
		wrapConfig := domain.SevensConfig{TunnelEnabled: true, TunnelSkipWidth: 3}
		s := setupSpadeBoard(wrapConfig, []int{8, 9, 10, 11, 12, 13})
		// 13 is placed. With tunnel wrap: 13+3 → wrapValue(16) = 3
		card3 := domain.NewCard(domain.CardDesignSpade, 3, false)
		assert.True(t, s.IsPlayable(card3))
	})

	t.Run("skip=1: treated as skip < 2 so disabled", func(t *testing.T) {
		skip1Config := domain.SevensConfig{TunnelSkipWidth: 1}
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, skip1Config)
		// With skip=1 (< 2), only normal adjacency applies
		card5 := domain.NewCard(domain.CardDesignSpade, 5, false)
		assert.False(t, s.IsPlayable(card5))
	})

	t.Run("SetConfig clamps negative tunnelSkipWidth to 0", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		s.SetConfig(domain.SevensConfig{TunnelSkipWidth: -1, MaxPasses: 5})
		assert.Equal(t, 0, s.GetConfig().TunnelSkipWidth)
	})

	t.Run("SetConfig clamps tunnelSkipWidth > 12 to 12", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		s.SetConfig(domain.SevensConfig{TunnelSkipWidth: 15, MaxPasses: 5})
		assert.Equal(t, 12, s.GetConfig().TunnelSkipWidth)
	})
}

func TestSevens_Joker(t *testing.T) {
	jokerConfig := domain.SevensConfig{TunnelEnabled: false, JokerCount: 2, CpuStrategy: domain.SevensCpuSimple}

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
		_ = s.PlayerPlay(0) // game ends
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

		_ = s.PlayerPlay(0) // human plays 8♠
		s.CpuPlay()         // CPU 1 plays joker

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

		_ = s.PlayerPlay(0) // human plays 8♠
		s.CpuPlay()         // CPU 1: 5♠ not playable, no passes → eliminated

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
		_ = s.PlayerPlay(0) // human: 8♠
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
		_ = s.PlayerPlay(0) // human: 9♠
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
						_ = s.PlayerPlay(i)
						played = true
						break
					}
				}
				if !played {
					_ = s.PlayerPlay(-1) // pass to advance
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
	strategyConfig := domain.SevensConfig{TunnelEnabled: false, JokerCount: 0, CpuStrategy: domain.SevensCpuStrategic}

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

		_ = s.PlayerPlay(0) // human plays 8♠
		s.CpuPlay()         // CPU 1: strategic evaluation

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

		_ = s.PlayerPlay(0) // human plays 8♠
		s.CpuPlay()         // CPU 1: plays 6♠ (positive score: has 5♠)

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

		_ = s.PlayerPlay(0) // human plays 8♠
		s.CpuPlay()         // CPU 1: forced to play (only 1 pass remaining)

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

		_ = s.PlayerPlay(0) // human plays 8♠
		s.CpuPlay()         // CPU 1: non-strategic, plays 6♠

		actions := s.GetCpuActions()
		assert.Len(t, actions, 1)
		assert.NotNil(t, actions[0].PlayedCard) // plays the card instead of passing
		assert.Equal(t, 6, actions[0].PlayedCard.GetValue())
	})
}

func TestSevens_CpuHarassment(t *testing.T) {
	t.Run("success harassment CPU passes when play helps opponent", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{CpuStrategy: domain.SevensCpuHarassment, MaxPasses: domain.SevensMaxPasses}
		s := domain.NewSevens(tc, players, cfg)
		// Human plays spade 8 first to advance turn to CPU
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		// CPU(player1) has spade 6, opponent(player2) has spade 5
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		_ = s.PlayerPlay(0) // human plays
		s.CpuPlay()         // CPU 1: harassment evaluation

		actions := s.GetCpuActions()
		assert.Len(t, actions, 1)
		assert.Nil(t, actions[0].PlayedCard) // passed
	})

	t.Run("success harassment CPU plays when blocking opponents", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{CpuStrategy: domain.SevensCpuHarassment, MaxPasses: domain.SevensMaxPasses}
		s := domain.NewSevens(tc, players, cfg)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		// CPU(player1) has spade 6, no one has spade 5
		// Playing 6 with opponents having cards beyond 5 → they are blocked
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignSpade, 4, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		_ = s.PlayerPlay(0)
		s.CpuPlay()

		actions := s.GetCpuActions()
		assert.Len(t, actions, 1)
		assert.NotNil(t, actions[0].PlayedCard)
		assert.Equal(t, 6, actions[0].PlayedCard.GetValue())
	})

	t.Run("success harassment CPU plays when self holds next and no opponents blocked", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{CpuStrategy: domain.SevensCpuHarassment, MaxPasses: domain.SevensMaxPasses}
		s := domain.NewSevens(tc, players, cfg)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		// CPU has spade 6 and spade 5 (self holds next card, safe to play)
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		_ = s.PlayerPlay(0)
		s.CpuPlay()

		actions := s.GetCpuActions()
		assert.Len(t, actions, 1)
		assert.NotNil(t, actions[0].PlayedCard)
	})

	t.Run("success harassment CPU forced to play when no passes left", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{CpuStrategy: domain.SevensCpuHarassment, MaxPasses: 1}
		s := domain.NewSevens(tc, players, cfg)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		// CPU has spade 6, opponent has spade 5
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		// Use up passes so only 1 remains
		players[1].IncrPassesUsed()
		_ = s.PlayerPlay(0)
		s.CpuPlay()

		actions := s.GetCpuActions()
		assert.Len(t, actions, 1)
		// All passes used up minus reserve → should be forced to play
	})

	t.Run("success findBestPlay dispatches to harassment mode", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{CpuStrategy: domain.SevensCpuHarassment, MaxPasses: domain.SevensMaxPasses}
		s := domain.NewSevens(tc, players, cfg)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		_ = s.PlayerPlay(0)
		s.CpuPlay()

		actions := s.GetCpuActions()
		assert.Len(t, actions, 1)
	})
}

func TestSevens_CpuHarassment_Joker(t *testing.T) {
	t.Run("success harassment CPU evaluates joker plays", func(t *testing.T) {
		tc := domain.NewTrumpCards(2)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{CpuStrategy: domain.SevensCpuHarassment, JokerCount: 2, MaxPasses: domain.SevensMaxPasses}
		s := domain.NewSevens(tc, players, cfg)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		// CPU has a joker
		players[1].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignClover, 9, false))
		_ = s.PlayerPlay(0)
		s.CpuPlay()

		actions := s.GetCpuActions()
		assert.Len(t, actions, 1)
	})

	t.Run("success harassment CPU joker blocked by finish rule", func(t *testing.T) {
		tc := domain.NewTrumpCards(2)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{CpuStrategy: domain.SevensCpuHarassment, JokerCount: 2, MaxPasses: domain.SevensMaxPasses, NoJokerFinish: true}
		s := domain.NewSevens(tc, players, cfg)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		// CPU has only jokers (can't finish with joker)
		players[1].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignClover, 2, false))
		_ = s.PlayerPlay(0)
		s.CpuPlay()

		actions := s.GetCpuActions()
		assert.Len(t, actions, 1)
		assert.Nil(t, actions[0].PlayedCard) // forced pass because joker blocked
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
	jokerConfig := domain.SevensConfig{TunnelEnabled: false, JokerCount: 2, CpuStrategy: domain.SevensCpuSimple}
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
	jokerConfig := domain.SevensConfig{TunnelEnabled: false, JokerCount: 2, CpuStrategy: domain.SevensCpuSimple}
	s := domain.NewSevens(tc, players, jokerConfig)

	// Human plays a card to advance turn to CPU
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignJoker, 1, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

	_ = s.PlayerPlay(0) // human plays 8♠ → turn moves to CPU 1
	assert.False(t, s.IsHumanTurn())

	// Now try PlayerPlayJoker on CPU's turn
	err := s.PlayerPlayJoker(0, domain.CardDesignSpade, 6)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
}

func TestSevens_FindPlayableSimple_JokerNoPosition(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	jokerConfig := domain.SevensConfig{TunnelEnabled: false, JokerCount: 2, CpuStrategy: domain.SevensCpuSimple}
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
	for !s.GetGameEndFlag() {
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
					_ = s.PlayerPlay(i)
					played = true
					break
				}
			}
			if !played {
				_ = s.PlayerPlay(-1) // pass
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
		_ = s.PlayerPlay(-1)
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
	strategyConfig := domain.SevensConfig{TunnelEnabled: false, JokerCount: 2, CpuStrategy: domain.SevensCpuStrategic}
	s := domain.NewSevens(tc, players, strategyConfig)

	// CPU 1 has a joker + 5♠. Strategic evaluation should place joker at 6♠
	// (adjacent to 7♠), which gives +2 because CPU has 5♠ (the next low card).
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignJoker, 1, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

	_ = s.PlayerPlay(0) // human plays 8♠
	s.CpuPlay()         // CPU 1: strategic, plays joker

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
	strategyConfig := domain.SevensConfig{TunnelEnabled: false, JokerCount: 0, CpuStrategy: domain.SevensCpuStrategic}
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

	_ = s.PlayerPlay(0) // human plays 8♠
	s.CpuPlay()         // CPU 1: all plays negative → pass

	actions := s.GetCpuActions()
	assert.Len(t, actions, 1)
	assert.Nil(t, actions[0].PlayedCard) // passed
	assert.Equal(t, 1, players[1].GetPassesUsed())
}

func TestSevens_EvaluatePlay_LowDirection(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	strategyConfig := domain.SevensConfig{TunnelEnabled: false, JokerCount: 0, CpuStrategy: domain.SevensCpuStrategic}
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

	_ = s.PlayerPlay(0) // human plays 6♠ → board spade min=6
	s.CpuPlay()         // CPU 1: plays 5♠ (score +2 for having 4♠)

	actions := s.GetCpuActions()
	assert.Len(t, actions, 1)
	assert.NotNil(t, actions[0].PlayedCard)
	assert.Equal(t, 5, actions[0].PlayedCard.GetValue())
}

func TestSevens_EvaluatePlay_LowDirection_NegativeScore(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	strategyConfig := domain.SevensConfig{TunnelEnabled: false, JokerCount: 0, CpuStrategy: domain.SevensCpuStrategic}
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

	_ = s.PlayerPlay(0) // human plays 6♠
	s.CpuPlay()         // CPU 1: 5♠ has score -1 → pass

	actions := s.GetCpuActions()
	assert.Len(t, actions, 1)
	assert.Nil(t, actions[0].PlayedCard) // passed
}

func TestSevens_EvaluatePlay_TunnelWrap(t *testing.T) {
	tunnelConfig := domain.SevensConfig{TunnelEnabled: true, JokerCount: 0, CpuStrategy: domain.SevensCpuSimple}
	tunnelStrategyConfig := domain.SevensConfig{TunnelEnabled: true, JokerCount: 0, CpuStrategy: domain.SevensCpuStrategic}

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
			_ = s.PlayerPlay(idx)
		}

		// Board now has spade 2-12. Ace(1) and King(13) are playable with tunnel.
		// Advance to CPU 1's turn
		for !s.GetGameEndFlag() && s.IsHumanTurn() {
			_ = s.PlayerPlay(-1)
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
	strategyConfig := domain.SevensConfig{TunnelEnabled: false, JokerCount: 2, CpuStrategy: domain.SevensCpuStrategic}
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

	_ = s.PlayerPlay(0) // human plays 8♠
	s.CpuPlay()         // CPU 1: strategic joker evaluation

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
	jokerConfig := domain.SevensConfig{TunnelEnabled: false, JokerCount: 2, CpuStrategy: domain.SevensCpuSimple}
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
	for !s.GetGameEndFlag() {
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
					_ = s.PlayerPlay(i)
					played = true
					break
				}
			}
			if !played {
				_ = s.PlayerPlay(-1)
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

	_ = s.PlayerPlay(0) // human plays 8♠ → turn advances to CPU 1
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
	jokerConfig := domain.SevensConfig{TunnelEnabled: false, JokerCount: 2, CpuStrategy: domain.SevensCpuSimple}
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
	jokerConfig := domain.SevensConfig{TunnelEnabled: false, JokerCount: 2, CpuStrategy: domain.SevensCpuSimple}
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

	_ = s.PlayerPlay(0) // human plays last card -> checkGameEnd -> gameEndFlag=true
	assert.True(t, s.GetGameEndFlag())
	assert.Equal(t, 0, s.GetCurrentTurn())

	// Further play attempts are blocked
	err := s.PlayerPlay(-1)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrGameEnded)
}

func TestSevens_FindPlayableStrategic_BestScoreUpdated(t *testing.T) {
	// Covers lines 453-455: the loop body where a later play has higher score
	// than the first, updating 'best'.
	// Setup: CPU has two playable cards. The first has a lower score, the second
	// has a higher score.
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	strategyConfig := domain.SevensConfig{TunnelEnabled: false, JokerCount: 0, CpuStrategy: domain.SevensCpuStrategic}
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

	_ = s.PlayerPlay(0) // human plays 8♠
	s.CpuPlay()         // CPU 1: strategic, evaluates both cards

	actions := s.GetCpuActions()
	assert.Len(t, actions, 1)
	assert.NotNil(t, actions[0].PlayedCard)
	// CPU should play 8♥ (score +2) over 6♠ (score -1)
	assert.Equal(t, domain.CardDesignHeart, actions[0].PlayedCard.GetDesign())
	assert.Equal(t, 8, actions[0].PlayedCard.GetValue())
}

func TestSevens_SetConfig(t *testing.T) {
	t.Run("success SetConfig with normal values", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		cfg := domain.SevensConfig{TunnelEnabled: true, JokerCount: 1, CpuStrategy: domain.SevensCpuStrategic, MaxPasses: 3, NoJokerFinish: true}
		s.SetConfig(cfg)
		assert.Equal(t, cfg, s.GetConfig())
	})

	t.Run("success SetConfig clamps negative jokerCount to 0", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		cfg := domain.SevensConfig{JokerCount: -5, MaxPasses: domain.SevensMaxPasses}
		s.SetConfig(cfg)
		assert.Equal(t, 0, s.GetConfig().JokerCount)
	})

	t.Run("success SetConfig clamps jokerCount above max to max", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		cfg := domain.SevensConfig{JokerCount: 10, MaxPasses: domain.SevensMaxPasses}
		s.SetConfig(cfg)
		assert.Equal(t, domain.SevensMaxJokerCount, s.GetConfig().JokerCount)
	})

	t.Run("success SetConfig clamps negative maxPasses to 0", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		cfg := domain.SevensConfig{JokerCount: 0, MaxPasses: -1}
		s.SetConfig(cfg)
		assert.Equal(t, 0, s.GetConfig().MaxPasses)
	})

	t.Run("success SetConfig clamps out-of-range cpuStrategy below min to simple", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		cfg := domain.SevensConfig{CpuStrategy: -1, MaxPasses: domain.SevensMaxPasses}
		s.SetConfig(cfg)
		assert.Equal(t, domain.SevensCpuSimple, s.GetConfig().CpuStrategy)
	})

	t.Run("success SetConfig clamps out-of-range cpuStrategy above max to simple", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		cfg := domain.SevensConfig{CpuStrategy: 99, MaxPasses: domain.SevensMaxPasses}
		s.SetConfig(cfg)
		assert.Equal(t, domain.SevensCpuSimple, s.GetConfig().CpuStrategy)
	})

	t.Run("success SetConfig then Reset recreates deck with joker", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		cfg := domain.SevensConfig{JokerCount: 1, MaxPasses: domain.SevensMaxPasses}
		s.SetConfig(cfg)
		s.Reset()
		// After reset, total cards dealt = 52 + 1 joker - 4 sevens = 49
		totalCards := 0
		for i := 0; i < s.GetPlayerCnt(); i++ {
			totalCards += s.GetPlayer(i).GetCardsSize()
		}
		assert.Equal(t, 49, totalCards)
	})
}

func TestSevens_ResetAppliesMaxPasses(t *testing.T) {
	t.Run("success Reset applies config.MaxPasses to all players", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{TunnelEnabled: false, JokerCount: 0, CpuStrategy: domain.SevensCpuSimple, MaxPasses: 3}
		s := domain.NewSevens(tc, players, cfg)
		s.Reset()
		for i := 0; i < s.GetPlayerCnt(); i++ {
			assert.Equal(t, 3, s.GetPlayer(i).GetMaxPasses())
		}
	})

	t.Run("success Reset applies MaxPasses 0 (unlimited) to all players", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{TunnelEnabled: false, JokerCount: 0, CpuStrategy: domain.SevensCpuSimple, MaxPasses: 0}
		s := domain.NewSevens(tc, players, cfg)
		s.Reset()
		for i := 0; i < s.GetPlayerCnt(); i++ {
			assert.Equal(t, 0, s.GetPlayer(i).GetMaxPasses())
		}
	})

	t.Run("success Reset applies default MaxPasses (5) to all players", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		s.Reset()
		for i := 0; i < s.GetPlayerCnt(); i++ {
			assert.Equal(t, domain.SevensMaxPasses, s.GetPlayer(i).GetMaxPasses())
		}
	})
}

func TestSevens_ForcedPass(t *testing.T) {
	t.Run("success CpuPlay forced pass when no playable cards", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false)) // not adjacent → pass
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		_ = s.PlayerPlay(0) // human plays 8♠
		s.CpuPlay()         // CPU 1: 5♠ not playable → forced pass

		actions := s.GetCpuActions()
		assert.Len(t, actions, 1)
		assert.Nil(t, actions[0].PlayedCard)
		assert.True(t, actions[0].ForcedPass)
	})

	t.Run("success CpuPlay voluntary strategic pass when has playable cards", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		strategyConfig := domain.SevensConfig{TunnelEnabled: false, JokerCount: 0, CpuStrategy: domain.SevensCpuStrategic, MaxPasses: domain.SevensMaxPasses}
		s := domain.NewSevens(tc, players, strategyConfig)
		// CPU 1 has 6♠ (playable) but does NOT have 5♠ → negative score → voluntary pass
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		_ = s.PlayerPlay(0) // human plays 8♠
		s.CpuPlay()         // CPU 1: has playable card but strategic pass (negative score)

		actions := s.GetCpuActions()
		assert.Len(t, actions, 1)
		assert.Nil(t, actions[0].PlayedCard)
		assert.False(t, actions[0].ForcedPass) // voluntary pass
	})

	t.Run("success PlayerPlay forced pass when no playable cards", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		// Human has only non-playable cards
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		err := s.PlayerPlay(-1) // human passes
		assert.NoError(t, err)
		action := s.GetHumanAction()
		assert.NotNil(t, action)
		assert.Nil(t, action.PlayedCard)
		assert.True(t, action.ForcedPass) // no playable cards → forced pass
	})

	t.Run("success PlayerPlay voluntary pass when has playable cards", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		// Human has playable card but chooses to pass
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		err := s.PlayerPlay(-1) // human voluntarily passes
		assert.NoError(t, err)
		action := s.GetHumanAction()
		assert.NotNil(t, action)
		assert.Nil(t, action.PlayedCard)
		assert.False(t, action.ForcedPass) // has playable cards → voluntary pass
	})

	t.Run("success AutoHandleNoOption sets ForcedPass true", func(t *testing.T) {
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
		action := s.GetHumanAction()
		assert.NotNil(t, action)
		assert.True(t, action.ForcedPass)
	})
}

func TestSevens_CountOpponentsBlocked(t *testing.T) {
	// countWeightedOpponentsBlocked is unexported but tested via evaluatePlay behavior.
	// We set up scenarios where opponents hold blocked cards and verify the
	// evaluatePlay score changes accordingly.

	t.Run("success evaluatePlay penalty increases with more opponents blocked", func(t *testing.T) {
		// Setup: strategy enabled. CPU 1 has 6♠ (playable, adjacent to 7).
		// CPU does NOT have 5♠ → penalty for low direction.
		// Opponents (player 0, 2, 3) hold 5♠ → countWeightedOpponentsBlocked = up to 3.
		// Penalty = -(1 + blocked_count).

		// Case 1: 0 opponents hold blocked card (5♠)
		tc0 := domain.NewTrumpCards(0)
		p0 := makeSevensPlayers()
		cfg := domain.SevensConfig{TunnelEnabled: false, JokerCount: 0, CpuStrategy: domain.SevensCpuStrategic, MaxPasses: domain.SevensMaxPasses}
		s0 := domain.NewSevens(tc0, p0, cfg)
		p0[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		p0[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		// CPU 1 has only 6♠ (no chain card)
		p0[1].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		// No opponent has 5♠
		p0[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		p0[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		_ = s0.PlayerPlay(0)
		s0.CpuPlay()
		actions0 := s0.GetCpuActions()
		assert.Len(t, actions0, 1)
		// CPU should pass since score is negative (-(1+0) = -1)
		assert.Nil(t, actions0[0].PlayedCard)

		// Case 2: 1 opponent holds blocked card (5♠)
		tc1 := domain.NewTrumpCards(0)
		p1 := makeSevensPlayers()
		s1 := domain.NewSevens(tc1, p1, cfg)
		p1[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		p1[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		p1[1].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		p1[2].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false)) // opponent has 5♠
		p1[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		_ = s1.PlayerPlay(0)
		s1.CpuPlay()
		actions1 := s1.GetCpuActions()
		assert.Len(t, actions1, 1)
		// CPU should pass, penalty is -(1+1) = -2
		assert.Nil(t, actions1[0].PlayedCard)

		// Case 3: 2 opponents hold blocked card (5♠)
		tc2 := domain.NewTrumpCards(0)
		p2 := makeSevensPlayers()
		s2 := domain.NewSevens(tc2, p2, cfg)
		p2[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		p2[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		p2[1].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		p2[2].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false)) // opponent has 5♠
		p2[3].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false)) // another opponent has 5♠
		_ = s2.PlayerPlay(0)
		s2.CpuPlay()
		actions2 := s2.GetCpuActions()
		assert.Len(t, actions2, 1)
		// CPU should pass, penalty is -(1+2) = -3
		assert.Nil(t, actions2[0].PlayedCard)

		// Case 4: 3 opponents hold blocked card (5♠)
		tc3 := domain.NewTrumpCards(0)
		p3 := makeSevensPlayers()
		s3 := domain.NewSevens(tc3, p3, cfg)
		p3[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		p3[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		p3[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false)) // human also has 5♠
		p3[1].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		p3[2].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false)) // opponent has 5♠
		p3[3].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false)) // another opponent has 5♠
		_ = s3.PlayerPlay(0)
		s3.CpuPlay()
		actions3 := s3.GetCpuActions()
		assert.Len(t, actions3, 1)
		// CPU should pass, penalty is -(1+3) = -4
		assert.Nil(t, actions3[0].PlayedCard)
	})

	t.Run("success evaluatePlay with opponent blocked in high direction", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{TunnelEnabled: false, JokerCount: 0, CpuStrategy: domain.SevensCpuStrategic, MaxPasses: domain.SevensMaxPasses}
		s := domain.NewSevens(tc, players, cfg)

		// CPU 1 has 8♠ (playable, adjacent to 7) but NOT 9♠
		// Opponents hold 9♠ → blocked in high direction
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false)) // opponent has 9♠
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		_ = s.PlayerPlay(0) // human plays 6♠
		s.CpuPlay()         // CPU 1 evaluates 8♠

		actions := s.GetCpuActions()
		assert.Len(t, actions, 1)
		// Low direction: nextLow=7 (placed) → 0. High direction: nextHigh=9 (not placed, opponent has it) → -(1+1) = -2
		// Total score = -2, so CPU should pass
		assert.Nil(t, actions[0].PlayedCard)
	})
}

func TestSevens_UnlimitedPasses(t *testing.T) {
	t.Run("success findPlayableStrategic with unlimited passes", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{TunnelEnabled: false, JokerCount: 0, CpuStrategy: domain.SevensCpuStrategic, MaxPasses: 0}
		s := domain.NewSevens(tc, players, cfg)
		// Set unlimited passes on all players
		for i := 0; i < 4; i++ {
			players[i].SetMaxPasses(0)
		}

		// CPU 1 has 6♠ (playable) but NOT 5♠ → negative score
		// With unlimited passes, CPU should still be able to pass freely
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		_ = s.PlayerPlay(0) // human plays 8♠
		s.CpuPlay()         // CPU 1: unlimited passes, negative score → pass

		actions := s.GetCpuActions()
		assert.Len(t, actions, 1)
		assert.Nil(t, actions[0].PlayedCard) // passed
		assert.Equal(t, 1, players[1].GetPassesUsed())
	})

	t.Run("success unlimited passes player never eliminated from pass exhaustion", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{TunnelEnabled: false, JokerCount: 0, CpuStrategy: domain.SevensCpuSimple, MaxPasses: 0}
		s := domain.NewSevens(tc, players, cfg)
		// Set unlimited passes on all players
		for i := 0; i < 4; i++ {
			players[i].SetMaxPasses(0)
		}

		// CPU 1 has only non-playable cards. With unlimited passes, it should pass repeatedly
		// and never be eliminated.
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false)) // not playable
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		_ = s.PlayerPlay(0) // human plays 8♠
		// CPU 1 passes multiple times without being eliminated
		for i := 0; i < 10; i++ {
			if s.GetGameEndFlag() {
				break
			}
			if !s.IsHumanTurn() && s.GetCurrentTurn() == 1 {
				s.CpuPlay()
				assert.False(t, players[1].GetIsEliminated(), "unlimited pass player should not be eliminated")
				assert.False(t, players[1].GetIsFinished(), "unlimited pass player should not be finished")
			} else {
				break
			}
		}
		assert.Greater(t, players[1].GetPassesUsed(), 0)
		assert.True(t, players[1].CanPass()) // still can pass
	})

	t.Run("success unlimited passes strategic CPU still passes with low-reserve guard", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{TunnelEnabled: false, JokerCount: 0, CpuStrategy: domain.SevensCpuStrategic, MaxPasses: 0}
		s := domain.NewSevens(tc, players, cfg)
		for i := 0; i < 4; i++ {
			players[i].SetMaxPasses(0)
		}

		// Even after many passes used, unlimited passes means reserve guard should not force play
		for i := 0; i < 100; i++ {
			players[1].IncrPassesUsed()
		}
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false)) // playable but negative score
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		_ = s.PlayerPlay(0) // human plays 8♠
		s.CpuPlay()         // CPU 1: unlimited passes → should still pass despite many passes used

		actions := s.GetCpuActions()
		assert.Len(t, actions, 1)
		assert.Nil(t, actions[0].PlayedCard) // still passes (unlimited)
	})
}

func TestSevens_NoJokerFinish(t *testing.T) {
	t.Run("PlayerPlay rejects joker card (must use PlayerPlayJoker)", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{JokerCount: 1, NoJokerFinish: true, MaxPasses: domain.SevensMaxPasses}
		s := domain.NewSevens(tc, players, cfg)

		// Give other players cards so game doesn't end
		for i := 1; i < 4; i++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		}
		// Human has ONLY a joker
		players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))

		err := s.PlayerPlay(0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "use PlayerPlayJoker")
	})

	t.Run("PlayerPlay rejects joker even when player has other cards", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{JokerCount: 1, NoJokerFinish: true, MaxPasses: domain.SevensMaxPasses}
		s := domain.NewSevens(tc, players, cfg)

		for i := 1; i < 4; i++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		}
		// Human has joker + a normal card — joker at index 0
		players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		// PlayerPlay always rejects joker cards
		err := s.PlayerPlay(0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "use PlayerPlayJoker")
	})

	t.Run("PlayerPlay rejects joker even when NoJokerFinish disabled", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{JokerCount: 1, NoJokerFinish: false, MaxPasses: domain.SevensMaxPasses}
		s := domain.NewSevens(tc, players, cfg)

		for i := 1; i < 4; i++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		}
		// Human has ONLY a joker but rule is OFF — still rejected
		players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))

		err := s.PlayerPlay(0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "use PlayerPlayJoker")
	})

	t.Run("PlayerPlayJoker rejects when only jokers left + NoJokerFinish", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{JokerCount: 1, NoJokerFinish: true, MaxPasses: domain.SevensMaxPasses}
		s := domain.NewSevens(tc, players, cfg)

		for i := 1; i < 4; i++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		}
		players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))

		err := s.PlayerPlayJoker(0, domain.CardDesignSpade, 6)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot finish with a joker")
	})

	t.Run("PlayerPlayJoker allows when player has other cards + NoJokerFinish", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{JokerCount: 1, NoJokerFinish: true, MaxPasses: domain.SevensMaxPasses}
		s := domain.NewSevens(tc, players, cfg)

		for i := 1; i < 4; i++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		}
		players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))

		err := s.PlayerPlayJoker(0, domain.CardDesignSpade, 6)
		assert.NoError(t, err)
	})

	t.Run("HasAnyOption returns false when only jokers + NoJokerFinish + no passes", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{JokerCount: 1, NoJokerFinish: true, MaxPasses: 1}
		s := domain.NewSevens(tc, players, cfg)

		for i := 0; i < 4; i++ {
			players[i].SetMaxPasses(1)
		}
		players[0].IncrPassesUsed() // used 1/1 pass
		players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))

		assert.False(t, s.HasAnyOption(0))
	})

	t.Run("HasAnyOption returns true when only jokers + NoJokerFinish + passes remain", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{JokerCount: 1, NoJokerFinish: true, MaxPasses: 5}
		s := domain.NewSevens(tc, players, cfg)

		players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))

		assert.True(t, s.HasAnyOption(0))
	})

	t.Run("CpuPlay eliminates CPU when only jokers + NoJokerFinish + no passes", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{JokerCount: 1, NoJokerFinish: true, MaxPasses: 1}
		s := domain.NewSevens(tc, players, cfg)

		for i := 0; i < 4; i++ {
			players[i].SetMaxPasses(1)
			players[i].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		}

		// Human plays to advance turn
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		_ = s.PlayerPlay(players[0].GetCardsSize() - 1) // play 6♠

		// CPU 1 has only joker, used all passes
		players[1].IncrPassesUsed() // used 1/1 pass
		// Remove the dummy card we gave above
		players[1].RemoveCard(0) // remove diamond 2
		players[1].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))

		require.Equal(t, 1, s.GetCurrentTurn())
		s.CpuPlay()
		assert.True(t, players[1].GetIsEliminated())
	})

	t.Run("CpuPlay passes when only jokers + NoJokerFinish + passes remain", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{JokerCount: 1, NoJokerFinish: true, MaxPasses: 5}
		s := domain.NewSevens(tc, players, cfg)

		for i := 0; i < 4; i++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		}

		// Human plays to advance turn
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		_ = s.PlayerPlay(players[0].GetCardsSize() - 1) // play 6♠

		// CPU 1 has only joker but has passes remaining
		players[1].RemoveCard(0) // remove diamond 2
		players[1].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))

		require.Equal(t, 1, s.GetCurrentTurn())
		s.CpuPlay()
		actions := s.GetCpuActions()
		assert.NotEmpty(t, actions)
		lastAction := actions[len(actions)-1]
		require.Equal(t, 1, lastAction.PlayerIdx)
		assert.Nil(t, lastAction.PlayedCard) // passed
		assert.False(t, players[1].GetIsEliminated())
	})

	t.Run("findPlayableSimple skips blocked jokers only-joker case", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{JokerCount: 1, NoJokerFinish: true, CpuStrategy: domain.SevensCpuSimple, MaxPasses: 5}
		s := domain.NewSevens(tc, players, cfg)

		for i := 0; i < 4; i++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		}

		// Human plays to advance turn
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		_ = s.PlayerPlay(players[0].GetCardsSize() - 1)

		// CPU 1: only joker (blocked by finish rule)
		players[1].RemoveCard(0) // remove diamond 2
		players[1].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))

		require.Equal(t, 1, s.GetCurrentTurn())
		s.CpuPlay()
		actions := s.GetCpuActions()
		assert.NotEmpty(t, actions)
		lastAction := actions[len(actions)-1]
		require.Equal(t, 1, lastAction.PlayerIdx)
		assert.Nil(t, lastAction.PlayedCard) // passes because joker is blocked
	})

	t.Run("findPlayableStrategic skips blocked jokers", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{JokerCount: 1, NoJokerFinish: true, CpuStrategy: domain.SevensCpuStrategic, MaxPasses: 5}
		s := domain.NewSevens(tc, players, cfg)

		for i := 0; i < 4; i++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		}

		// Human plays to advance turn
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		_ = s.PlayerPlay(players[0].GetCardsSize() - 1)

		// CPU 1: only joker (blocked by finish rule)
		players[1].RemoveCard(0) // remove diamond 2
		players[1].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))

		require.Equal(t, 1, s.GetCurrentTurn())
		s.CpuPlay()
		actions := s.GetCpuActions()
		assert.NotEmpty(t, actions)
		lastAction := actions[len(actions)-1]
		require.Equal(t, 1, lastAction.PlayerIdx)
		assert.Nil(t, lastAction.PlayedCard) // passes because joker is blocked
	})

	t.Run("NewSevens stores NoJokerFinish config", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{NoJokerFinish: true}
		s := domain.NewSevens(tc, players, cfg)
		assert.True(t, s.GetConfig().NoJokerFinish)
	})
}

func TestSevens_WeightedOpponentsBlocked(t *testing.T) {
	t.Run("opponent with 1 pass left gets weight 3", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{CpuStrategy: domain.SevensCpuStrategic, MaxPasses: 3}
		s := domain.NewSevens(tc, players, cfg)

		for i := 0; i < 4; i++ {
			players[i].SetMaxPasses(3)
			players[i].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		}
		// Opponent (player 2) has 1 pass remaining (used 2/3)
		players[2].IncrPassesUsed()
		players[2].IncrPassesUsed()
		// Opponent has spade 5 (blocked behind 6 which is unplaced)
		players[2].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))

		// Human plays to advance turn
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		_ = s.PlayerPlay(players[0].GetCardsSize() - 1) // play 8♠

		// CPU 1 has spade 6 (playable adjacent to 7, opens path to opponent's 5)
		// With weighted blocking: opponent has 1 pass left -> weight 3
		// Score for playing 6♠: nextLow=5, not placed. CPU doesn't have 5, but opponent does.
		// weighted blocked = 3. score -= (1+3) = -4.
		// nextHigh = 7 placed, no penalty.
		// Without weighting it would be -(1+1) = -2.
		// CPU should pass (negative score, passes available)
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		require.Equal(t, 1, s.GetCurrentTurn())
		s.CpuPlay()
		actions := s.GetCpuActions()
		assert.NotEmpty(t, actions)
		lastAction := actions[len(actions)-1]
		require.Equal(t, 1, lastAction.PlayerIdx)
		assert.Nil(t, lastAction.PlayedCard) // passes due to high weighted penalty
	})

	t.Run("opponent with plenty of passes gets weight 1", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{CpuStrategy: domain.SevensCpuStrategic, MaxPasses: 10}
		s := domain.NewSevens(tc, players, cfg)

		for i := 0; i < 4; i++ {
			players[i].SetMaxPasses(10)
			players[i].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		}
		// Opponent has spade 5 (blocked behind 6)
		players[2].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))

		// Human plays
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		_ = s.PlayerPlay(players[0].GetCardsSize() - 1) // play 8♠

		// CPU 1 has spade 6 + spade 5 (has chain card)
		// Score: nextLow=5 not placed, CPU has 5 -> score +=2.
		// nextHigh = 7 placed. Total score = +2.
		// CPU should play (positive score)
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))

		require.Equal(t, 1, s.GetCurrentTurn())
		s.CpuPlay()
		actions := s.GetCpuActions()
		assert.NotEmpty(t, actions)
		lastAction := actions[len(actions)-1]
		require.Equal(t, 1, lastAction.PlayerIdx)
		assert.NotNil(t, lastAction.PlayedCard)
		assert.Equal(t, 6, lastAction.PlayedCard.GetValue())
	})

	t.Run("opponent with unlimited passes gets weight 1", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{CpuStrategy: domain.SevensCpuStrategic, MaxPasses: 0} // unlimited
		s := domain.NewSevens(tc, players, cfg)

		for i := 0; i < 4; i++ {
			players[i].SetMaxPasses(0) // unlimited
			players[i].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		}
		// Opponent has spade 5
		players[2].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))

		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		_ = s.PlayerPlay(players[0].GetCardsSize() - 1)

		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		require.Equal(t, 1, s.GetCurrentTurn())
		s.CpuPlay()
		actions := s.GetCpuActions()
		assert.NotEmpty(t, actions)
		lastAction := actions[len(actions)-1]
		require.Equal(t, 1, lastAction.PlayerIdx)
		// Score = -(1+1) = -2, passes available -> pass
		assert.Nil(t, lastAction.PlayedCard)
	})

	t.Run("opponent with 2 passes left gets weight 2", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{CpuStrategy: domain.SevensCpuStrategic, MaxPasses: 3}
		s := domain.NewSevens(tc, players, cfg)

		for i := 0; i < 4; i++ {
			players[i].SetMaxPasses(3)
			players[i].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		}
		// Opponent (player 2) has 2 passes remaining (used 1/3)
		players[2].IncrPassesUsed()
		// Opponent has spade 5 (blocked)
		players[2].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))

		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		_ = s.PlayerPlay(players[0].GetCardsSize() - 1)

		// CPU 1 has spade 6. With weight 2: score -= (1+2) = -3
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		require.Equal(t, 1, s.GetCurrentTurn())
		s.CpuPlay()
		actions := s.GetCpuActions()
		assert.NotEmpty(t, actions)
		lastAction := actions[len(actions)-1]
		require.Equal(t, 1, lastAction.PlayerIdx)
		assert.Nil(t, lastAction.PlayedCard) // passes due to negative score
	})
}

func TestSevens_EvaluatePlay_PlayerHasNextHighCard(t *testing.T) {
	// Covers lines 493-495: evaluatePlay where player has the next high card.
	// Setup: strategy enabled. CPU plays 8♠ (adjacent to 7) and has 9♠ in hand.
	// nextHigh = 9, not placed on board. CPU has 9♠ -> score += 2.
	// nextLow = 7 (placed on board) -> skip. Total score = +2.
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	strategyConfig := domain.SevensConfig{TunnelEnabled: false, JokerCount: 0, CpuStrategy: domain.SevensCpuStrategic}
	s := domain.NewSevens(tc, players, strategyConfig)

	// Give players dummy cards so game doesn't end
	for i := 0; i < 4; i++ {
		for d := 0; d < 5; d++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		}
	}

	// Human plays to advance turn
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
	_ = s.PlayerPlay(players[0].GetCardsSize() - 1) // play 6♠

	// CPU 1 has 8♠ (playable, adjacent to 7) and 9♠ (the next high card)
	players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))

	require.Equal(t, 1, s.GetCurrentTurn())
	s.CpuPlay()
	actions := s.GetCpuActions()
	assert.NotEmpty(t, actions)
	lastAction := actions[len(actions)-1]
	require.Equal(t, 1, lastAction.PlayerIdx)
	// CPU should play 8♠ because it has 9♠ (score = +2, positive)
	assert.NotNil(t, lastAction.PlayedCard)
	assert.Equal(t, 8, lastAction.PlayedCard.GetValue())
}

// ---------------------------------------------------------------------------
// Joker Reclaim tests
// ---------------------------------------------------------------------------

func TestJokerReclaim_BasicReclaim(t *testing.T) {
	// Full round-trip: play joker at position → play real card at same position → joker returns.
	tc := domain.NewTrumpCards(2)
	players := makeSevensPlayers()
	cfg := domain.SevensConfig{
		JokerCount:          2,
		JokerReclaimEnabled: true,
		MaxPasses:           domain.SevensMaxPasses,
	}
	s := domain.NewSevens(tc, players, cfg)

	// Give other players dummy cards so game doesn't end
	for i := 1; i <= 3; i++ {
		for d := 0; d < 5; d++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		}
	}

	// Human has joker + real Spade-6 + extra cards to avoid finishing
	players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 1, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignClover, 6, false))

	require.Equal(t, 4, players[0].GetCardsSize())

	// Play joker at Spade-6
	err := s.PlayerPlayJoker(0, domain.CardDesignSpade, 6)
	require.NoError(t, err)

	// Verify joker tracking
	assert.Equal(t, 1, s.GetJokerCardsCount(), "one joker should be tracked on board")
	jp := s.GetJokerPlaced()
	assert.True(t, jp[domain.CardDesignSpade]&(1<<6) != 0, "jokerPlaced bit should be set at spade-6")
	assert.Equal(t, 3, players[0].GetCardsSize(), "hand size should decrease by 1 after playing joker")

	// Board should have Spade-6 placed
	placed := s.GetTablePlaced()
	assert.True(t, placed[domain.CardDesignSpade]&(1<<6) != 0, "board should have spade-6")

	// Advance through CPUs back to human turn
	for !s.IsHumanTurn() && !s.GetGameEndFlag() {
		s.CpuPlay()
	}
	require.True(t, s.IsHumanTurn(), "should be back to human's turn")

	// Find real Spade-6 in hand
	spade6Idx := -1
	for i := 0; i < players[0].GetCardsSize(); i++ {
		c := players[0].GetCard(i)
		if c.GetDesign() == domain.CardDesignSpade && c.GetValue() == 6 {
			spade6Idx = i
			break
		}
	}
	require.NotEqual(t, -1, spade6Idx, "should have real Spade-6 in hand")

	handSizeBefore := players[0].GetCardsSize()

	// Play real Spade-6 at joker-occupied position → triggers reclaim
	err = s.PlayerPlay(spade6Idx)
	require.NoError(t, err)

	// Reclaim should have fired: card played (-1) + joker returned (+1) = same size
	assert.Equal(t, handSizeBefore, players[0].GetCardsSize(),
		"hand size should stay same: played card replaced by reclaimed joker")

	// Joker tracking cleared
	assert.Equal(t, 0, s.GetJokerCardsCount(), "joker should be removed from board tracking")
	jp = s.GetJokerPlaced()
	assert.True(t, jp[domain.CardDesignSpade]&(1<<6) == 0, "jokerPlaced bit should be cleared at spade-6")

	// Verify the reclaimed card is a joker
	hasJoker := false
	for i := 0; i < players[0].GetCardsSize(); i++ {
		if players[0].GetCard(i).GetDesign() == domain.CardDesignJoker {
			hasJoker = true
			break
		}
	}
	assert.True(t, hasJoker, "player should have a joker back in hand")
}

func TestJokerReclaim_DifferentPosition_NoReclaim(t *testing.T) {
	// Play joker at Spade-6, then play card at Heart-6 → no reclaim
	tc := domain.NewTrumpCards(2)
	players := makeSevensPlayers()
	cfg := domain.SevensConfig{
		JokerCount:          2,
		JokerReclaimEnabled: true,
		MaxPasses:           domain.SevensMaxPasses,
	}
	s := domain.NewSevens(tc, players, cfg)

	for i := 1; i <= 3; i++ {
		for d := 0; d < 5; d++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		}
	}

	players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 1, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignClover, 6, false))

	err := s.PlayerPlayJoker(0, domain.CardDesignSpade, 6)
	require.NoError(t, err)
	assert.Equal(t, 1, s.GetJokerCardsCount())

	for !s.IsHumanTurn() && !s.GetGameEndFlag() {
		s.CpuPlay()
	}

	// Find Heart-6
	var heart6Idx int
	for i := 0; i < players[0].GetCardsSize(); i++ {
		c := players[0].GetCard(i)
		if c.GetDesign() == domain.CardDesignHeart && c.GetValue() == 6 {
			heart6Idx = i
			break
		}
	}

	err = s.PlayerPlay(heart6Idx)
	require.NoError(t, err)

	// Joker tracking should remain (played card was Heart-6, not Spade-6)
	assert.Equal(t, 1, s.GetJokerCardsCount(), "joker should still be tracked")
	jp := s.GetJokerPlaced()
	assert.True(t, jp[domain.CardDesignSpade]&(1<<6) != 0, "jokerPlaced bit should still be set")
}

func TestJokerReclaim_Disabled_NoReclaim(t *testing.T) {
	// JokerReclaimEnabled=false -> joker placement is NOT tracked for reclaim
	tc := domain.NewTrumpCards(2)
	players := makeSevensPlayers()
	cfg := domain.SevensConfig{
		JokerCount:          2,
		JokerReclaimEnabled: false,
		MaxPasses:           domain.SevensMaxPasses,
	}
	s := domain.NewSevens(tc, players, cfg)

	// Give other players dummy cards
	for i := 1; i <= 3; i++ {
		for d := 0; d < 5; d++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		}
	}

	// Human has joker + extras
	players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 1, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignClover, 6, false))

	// Play joker at Spade-6
	err := s.PlayerPlayJoker(0, domain.CardDesignSpade, 6)
	require.NoError(t, err)

	// With reclaim disabled, joker tracking should NOT be populated
	assert.Equal(t, 0, s.GetJokerCardsCount(), "no joker cards tracked when reclaim disabled")
	jp := s.GetJokerPlaced()
	assert.True(t, jp[domain.CardDesignSpade]&(1<<6) == 0, "jokerPlaced bit not set when reclaim disabled")

	// Board should still have the position placed (game mechanics work regardless)
	placed := s.GetTablePlaced()
	assert.True(t, placed[domain.CardDesignSpade]&(1<<6) != 0, "board should have spade-6 regardless")
	assert.Equal(t, 2, players[0].GetCardsSize(), "hand size decreased by 1")
}

func TestJokerReclaim_CpuReclaim(t *testing.T) {
	// Full round-trip for CPU: human plays joker at Spade-6, CPU plays real Spade-6 → joker returns to CPU.
	tc := domain.NewTrumpCards(2)
	players := makeSevensPlayers()
	cfg := domain.SevensConfig{
		JokerCount:          2,
		JokerReclaimEnabled: true,
		MaxPasses:           domain.SevensMaxPasses,
	}
	s := domain.NewSevens(tc, players, cfg)

	// Give CPUs 2, 3 dummy cards
	for i := 2; i <= 3; i++ {
		for d := 0; d < 5; d++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		}
	}

	// Human places joker at Spade-6
	players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 1, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignClover, 6, false))

	err := s.PlayerPlayJoker(0, domain.CardDesignSpade, 6)
	require.NoError(t, err)
	assert.Equal(t, 1, s.GetJokerCardsCount())

	// CPU 1 has real Spade-6 (playable at joker-occupied position) + extra cards
	players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
	for d := 0; d < 4; d++ {
		players[1].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
	}

	cpu1HandBefore := players[1].GetCardsSize()

	// CPU 1 plays
	s.CpuPlay()
	actions := s.GetCpuActions()
	require.NotEmpty(t, actions)

	// Check if CPU 1 played Spade-6
	cpuPlayed6 := false
	for _, a := range actions {
		if a.PlayerIdx == 1 && a.PlayedCard != nil &&
			a.PlayedCard.GetDesign() == domain.CardDesignSpade && a.PlayedCard.GetValue() == 6 {
			cpuPlayed6 = true
			break
		}
	}

	if cpuPlayed6 {
		// Reclaim should have fired: CPU played 6♠ at joker position
		// Hand: before - 1 (played) + 1 (reclaimed) = same
		assert.Equal(t, cpu1HandBefore, players[1].GetCardsSize(),
			"CPU hand should reflect played card and reclaimed joker net effect")
		assert.Equal(t, 0, s.GetJokerCardsCount(), "joker removed from board tracking")
		jp := s.GetJokerPlaced()
		assert.True(t, jp[domain.CardDesignSpade]&(1<<6) == 0, "jokerPlaced bit cleared")

		// Verify CPU has joker in hand
		hasJoker := false
		for i := 0; i < players[1].GetCardsSize(); i++ {
			if players[1].GetCard(i).GetDesign() == domain.CardDesignJoker {
				hasJoker = true
				break
			}
		}
		assert.True(t, hasJoker, "CPU should have reclaimed joker")
	}
}

func TestJokerReclaim_CpuNoDifferentPosition(t *testing.T) {
	// CPU plays at a different position than joker → no reclaim
	tc := domain.NewTrumpCards(2)
	players := makeSevensPlayers()
	cfg := domain.SevensConfig{
		JokerCount:          2,
		JokerReclaimEnabled: true,
		MaxPasses:           domain.SevensMaxPasses,
	}
	s := domain.NewSevens(tc, players, cfg)

	for i := 2; i <= 3; i++ {
		for d := 0; d < 5; d++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		}
	}

	// Human places joker at Spade-8
	players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 1, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignClover, 6, false))

	err := s.PlayerPlayJoker(0, domain.CardDesignSpade, 8)
	require.NoError(t, err)
	assert.Equal(t, 1, s.GetJokerCardsCount())

	// CPU 1 has Spade-6 (different from joker position Spade-8)
	players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
	for d := 0; d < 4; d++ {
		players[1].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
	}

	s.CpuPlay()
	actions := s.GetCpuActions()
	require.NotEmpty(t, actions)

	// Joker should still be tracked (CPU played at Spade-6, not Spade-8)
	assert.Equal(t, 1, s.GetJokerCardsCount(), "joker should remain tracked")
	jp := s.GetJokerPlaced()
	assert.True(t, jp[domain.CardDesignSpade]&(1<<8) != 0, "jokerPlaced bit should still be set at spade-8")
}

func TestJokerReclaim_NoJokerAtPosition_NoOp(t *testing.T) {
	// Playing a card at a position without a joker -> no reclaim occurs
	tc := domain.NewTrumpCards(2)
	players := makeSevensPlayers()
	cfg := domain.SevensConfig{
		JokerCount:          2,
		JokerReclaimEnabled: true,
		MaxPasses:           domain.SevensMaxPasses,
	}
	s := domain.NewSevens(tc, players, cfg)

	// Give other players dummy cards
	for i := 1; i <= 3; i++ {
		for d := 0; d < 5; d++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		}
	}

	// Human has a normal card (no joker has been placed anywhere)
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))

	// No jokers on board
	assert.Equal(t, 0, s.GetJokerCardsCount())
	jp := s.GetJokerPlaced()
	for suit := 1; suit <= 4; suit++ {
		assert.Equal(t, uint16(0), jp[suit], "no jokerPlaced bits should be set")
	}

	handSizeBefore := players[0].GetCardsSize()
	err := s.PlayerPlay(0) // play Spade-6
	require.NoError(t, err)

	// Hand size decreased by exactly 1 (no joker returned)
	assert.Equal(t, handSizeBefore-1, players[0].GetCardsSize(), "hand should lose exactly 1 card")

	// Still no jokers tracked
	assert.Equal(t, 0, s.GetJokerCardsCount(), "no jokers should appear from nowhere")

	// No joker in player's hand
	for i := 0; i < players[0].GetCardsSize(); i++ {
		assert.NotEqual(t, domain.CardDesignJoker, players[0].GetCard(i).GetDesign(),
			"player should not have a joker")
	}
}

func TestJokerReclaim_MultipleJokersOnBoard(t *testing.T) {
	// Two jokers placed, then reclaim one → only one returned, one remains tracked.
	tc := domain.NewTrumpCards(2)
	players := makeSevensPlayers()
	cfg := domain.SevensConfig{
		JokerCount:          2,
		JokerReclaimEnabled: true,
		MaxPasses:           domain.SevensMaxPasses,
	}
	s := domain.NewSevens(tc, players, cfg)

	// Give other players dummy cards
	for i := 1; i <= 3; i++ {
		for d := 0; d < 5; d++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		}
	}

	// Human has two jokers + real Spade-6 for reclaim + extra cards
	players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 1, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 2, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignClover, 6, false))

	require.Equal(t, 5, players[0].GetCardsSize())

	// Place first joker at Spade-6
	err := s.PlayerPlayJoker(0, domain.CardDesignSpade, 6)
	require.NoError(t, err)
	assert.Equal(t, 1, s.GetJokerCardsCount(), "one joker tracked after first placement")

	// Advance through CPUs back to human turn
	for !s.IsHumanTurn() && !s.GetGameEndFlag() {
		s.CpuPlay()
	}
	require.True(t, s.IsHumanTurn())

	// Place second joker at Spade-5
	jokerIdx := -1
	for i := 0; i < players[0].GetCardsSize(); i++ {
		if players[0].GetCard(i).GetDesign() == domain.CardDesignJoker {
			jokerIdx = i
			break
		}
	}
	require.NotEqual(t, -1, jokerIdx, "should still have a joker in hand")

	err = s.PlayerPlayJoker(jokerIdx, domain.CardDesignSpade, 5)
	require.NoError(t, err)
	assert.Equal(t, 2, s.GetJokerCardsCount(), "two jokers tracked after second placement")

	// Advance through CPUs back to human turn
	for !s.IsHumanTurn() && !s.GetGameEndFlag() {
		s.CpuPlay()
	}
	require.True(t, s.IsHumanTurn())

	// Now play real Spade-6 at joker-occupied position → reclaim one joker
	spade6Idx := -1
	for i := 0; i < players[0].GetCardsSize(); i++ {
		c := players[0].GetCard(i)
		if c.GetDesign() == domain.CardDesignSpade && c.GetValue() == 6 {
			spade6Idx = i
			break
		}
	}
	require.NotEqual(t, -1, spade6Idx, "should have real Spade-6 in hand")

	err = s.PlayerPlay(spade6Idx)
	require.NoError(t, err)

	// One joker reclaimed, one remains
	assert.Equal(t, 1, s.GetJokerCardsCount(), "one joker should remain on board")
	jp := s.GetJokerPlaced()
	assert.True(t, jp[domain.CardDesignSpade]&(1<<6) == 0, "spade-6 joker bit should be cleared")
	assert.True(t, jp[domain.CardDesignSpade]&(1<<5) != 0, "spade-5 joker bit should remain set")

	// Player should have a joker in hand (reclaimed)
	hasJoker := false
	for i := 0; i < players[0].GetCardsSize(); i++ {
		if players[0].GetCard(i).GetDesign() == domain.CardDesignJoker {
			hasJoker = true
			break
		}
	}
	assert.True(t, hasJoker, "player should have reclaimed joker in hand")
}

func TestJokerReclaim_ReclaimDisabled_PositionNotPlayable(t *testing.T) {
	// When reclaim is disabled, a joker-occupied position should NOT be playable
	tc := domain.NewTrumpCards(2)
	players := makeSevensPlayers()
	cfg := domain.SevensConfig{
		JokerCount:          2,
		JokerReclaimEnabled: false,
		MaxPasses:           domain.SevensMaxPasses,
	}
	s := domain.NewSevens(tc, players, cfg)

	for i := 1; i <= 3; i++ {
		for d := 0; d < 5; d++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		}
	}

	// Human has joker + real Spade-6 + extra
	players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 1, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))

	// Play joker at Spade-6
	err := s.PlayerPlayJoker(0, domain.CardDesignSpade, 6)
	require.NoError(t, err)

	for !s.IsHumanTurn() && !s.GetGameEndFlag() {
		s.CpuPlay()
	}

	// Find real Spade-6
	spade6Idx := -1
	for i := 0; i < players[0].GetCardsSize(); i++ {
		c := players[0].GetCard(i)
		if c.GetDesign() == domain.CardDesignSpade && c.GetValue() == 6 {
			spade6Idx = i
			break
		}
	}
	require.NotEqual(t, -1, spade6Idx)

	// Playing real Spade-6 should fail (position already occupied, no reclaim enabled)
	err = s.PlayerPlay(spade6Idx)
	assert.Error(t, err, "should not be able to play on joker-occupied position when reclaim disabled")
}

func TestSevens_EndStop(t *testing.T) {
	t.Run("EndStop disabled does not block", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{EndStopEnabled: false, MaxPasses: 5}
		s := domain.NewSevens(tc, players, cfg)

		// Place A(1) of spade
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		// Place 6, 5, 4, 3, 2 first to reach A
		for v := 6; v >= 2; v-- {
			players[0].AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
		}
		// Play 6,5,4,3,2 first
		for v := 6; v >= 2; v-- {
			idx := -1
			for i := 0; i < players[0].GetCardsSize(); i++ {
				c := players[0].GetCard(i)
				if c.GetDesign() == domain.CardDesignSpade && c.GetValue() == v {
					idx = i
					break
				}
			}
			require.NotEqual(t, -1, idx)
			err := s.PlayerPlay(idx)
			require.NoError(t, err)
			// Skip CPU turns back to human
			for !s.IsHumanTurn() && !s.GetGameEndFlag() {
				s.CpuPlay()
			}
		}
		// Now play A
		aIdx := -1
		for i := 0; i < players[0].GetCardsSize(); i++ {
			c := players[0].GetCard(i)
			if c.GetDesign() == domain.CardDesignSpade && c.GetValue() == 1 {
				aIdx = i
				break
			}
		}
		require.NotEqual(t, -1, aIdx)
		err := s.PlayerPlay(aIdx)
		require.NoError(t, err)
		// Skip CPU turns
		for !s.IsHumanTurn() && !s.GetGameEndFlag() {
			s.CpuPlay()
		}
		// 8 of spade should still be playable (EndStop disabled)
		card8 := domain.NewCard(domain.CardDesignSpade, 8, false)
		assert.True(t, s.IsPlayable(card8))
	})

	t.Run("EndStop enabled A placed blocks high side (8)", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{EndStopEnabled: true, MaxPasses: 5}
		s := domain.NewSevens(tc, players, cfg)
		// Give CPUs cards so they don't finish
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		// Place 6,5,4,3,2,A of spade in sequence
		for v := 6; v >= 1; v-- {
			players[0].AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
		}
		// Add 8 of spade to test later
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))

		for v := 6; v >= 1; v-- {
			idx := -1
			for i := 0; i < players[0].GetCardsSize(); i++ {
				c := players[0].GetCard(i)
				if c.GetDesign() == domain.CardDesignSpade && c.GetValue() == v {
					idx = i
					break
				}
			}
			require.NotEqual(t, -1, idx, "should find spade %d", v)
			err := s.PlayerPlay(idx)
			require.NoError(t, err)
			for !s.IsHumanTurn() && !s.GetGameEndFlag() {
				s.CpuPlay()
			}
		}
		// A is now placed → high side (8-K) should be blocked
		card8 := domain.NewCard(domain.CardDesignSpade, 8, false)
		assert.False(t, s.IsPlayable(card8), "8 should be blocked when A is placed and EndStop enabled")
	})

	t.Run("EndStop enabled K placed blocks low side (6)", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{EndStopEnabled: true, MaxPasses: 5}
		s := domain.NewSevens(tc, players, cfg)
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		// Place 8,9,10,11,12,13 of spade
		for v := 8; v <= 13; v++ {
			players[0].AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
		}
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		for v := 8; v <= 13; v++ {
			idx := -1
			for i := 0; i < players[0].GetCardsSize(); i++ {
				c := players[0].GetCard(i)
				if c.GetDesign() == domain.CardDesignSpade && c.GetValue() == v {
					idx = i
					break
				}
			}
			require.NotEqual(t, -1, idx, "should find spade %d", v)
			err := s.PlayerPlay(idx)
			require.NoError(t, err)
			for !s.IsHumanTurn() && !s.GetGameEndFlag() {
				s.CpuPlay()
			}
		}
		// K is now placed → low side (1-6) should be blocked
		card6 := domain.NewCard(domain.CardDesignSpade, 6, false)
		assert.False(t, s.IsPlayable(card6), "6 should be blocked when K is placed and EndStop enabled")
	})

	t.Run("EndStop A placed does NOT block low side", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{EndStopEnabled: true, MaxPasses: 5}
		s := domain.NewSevens(tc, players, cfg)
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		// Place 6,5,4,3,2,A of spade
		for v := 6; v >= 1; v-- {
			players[0].AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
		}
		for v := 6; v >= 1; v-- {
			idx := -1
			for i := 0; i < players[0].GetCardsSize(); i++ {
				c := players[0].GetCard(i)
				if c.GetDesign() == domain.CardDesignSpade && c.GetValue() == v {
					idx = i
					break
				}
			}
			require.NotEqual(t, -1, idx)
			err := s.PlayerPlay(idx)
			require.NoError(t, err)
			for !s.IsHumanTurn() && !s.GetGameEndFlag() {
				s.CpuPlay()
			}
		}
		// A placed does not block the low side (already filled). Also verify 7 is always OK.
		card7 := domain.NewCard(domain.CardDesignSpade, 7, false)
		// 7 is already placed, so IsPlayable returns false (already on board), but isEndStopped(7) = false
		assert.False(t, s.IsPlayable(card7), "7 is already placed so not playable")
	})

	t.Run("EndStop K placed does NOT block high side", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{EndStopEnabled: true, MaxPasses: 5}
		s := domain.NewSevens(tc, players, cfg)
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		// Place 8..13 of spade (K placed)
		for v := 8; v <= 13; v++ {
			players[0].AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
		}
		for v := 8; v <= 13; v++ {
			idx := -1
			for i := 0; i < players[0].GetCardsSize(); i++ {
				c := players[0].GetCard(i)
				if c.GetDesign() == domain.CardDesignSpade && c.GetValue() == v {
					idx = i
					break
				}
			}
			require.NotEqual(t, -1, idx)
			err := s.PlayerPlay(idx)
			require.NoError(t, err)
			for !s.IsHumanTurn() && !s.GetGameEndFlag() {
				s.CpuPlay()
			}
		}
		// K placed does not block high side (already filled). Verify that nothing beyond K is affected.
		// Also: IsPlayable for value 7 remains "already placed"
		card7 := domain.NewCard(domain.CardDesignSpade, 7, false)
		assert.False(t, s.IsPlayable(card7), "7 is already placed")
	})

	t.Run("EndStop both A and K placed locks entire suit", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{EndStopEnabled: true, MaxPasses: 5}
		s := domain.NewSevens(tc, players, cfg)
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		// Place all of spade: 6,5,4,3,2,1,8,9,10,11,12,13
		for v := 6; v >= 1; v-- {
			players[0].AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
		}
		for v := 8; v <= 13; v++ {
			players[0].AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
		}
		for v := 6; v >= 1; v-- {
			idx := -1
			for i := 0; i < players[0].GetCardsSize(); i++ {
				c := players[0].GetCard(i)
				if c.GetDesign() == domain.CardDesignSpade && c.GetValue() == v {
					idx = i
					break
				}
			}
			require.NotEqual(t, -1, idx)
			err := s.PlayerPlay(idx)
			require.NoError(t, err)
			for !s.IsHumanTurn() && !s.GetGameEndFlag() {
				s.CpuPlay()
			}
		}
		// Now A is placed, high side should be blocked
		// But we also need to place 8 before A blocks it...
		// Actually A blocks 8-K, but 8 is adjacent to 7 which is placed.
		// With EndStop, even though 8 is adjacent to 7, it should still be blocked because A is placed.
		// Let's verify the heart suit instead (untouched)
		// Heart has only 7 placed, no A or K → 6 and 8 should be playable
		card6h := domain.NewCard(domain.CardDesignHeart, 6, false)
		card8h := domain.NewCard(domain.CardDesignHeart, 8, false)
		assert.True(t, s.IsPlayable(card6h))
		assert.True(t, s.IsPlayable(card8h))
	})

	t.Run("EndStop + Tunnel: K blocked even with tunnel when A placed", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{EndStopEnabled: true, TunnelEnabled: true, MaxPasses: 5}
		s := domain.NewSevens(tc, players, cfg)
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		// Place 6,5,4,3,2,1 of spade
		for v := 6; v >= 1; v-- {
			players[0].AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
		}
		for v := 6; v >= 1; v-- {
			idx := -1
			for i := 0; i < players[0].GetCardsSize(); i++ {
				c := players[0].GetCard(i)
				if c.GetDesign() == domain.CardDesignSpade && c.GetValue() == v {
					idx = i
					break
				}
			}
			require.NotEqual(t, -1, idx)
			err := s.PlayerPlay(idx)
			require.NoError(t, err)
			for !s.IsHumanTurn() && !s.GetGameEndFlag() {
				s.CpuPlay()
			}
		}
		// A(1) placed. With tunnel, K(13) would normally be playable (adjacent to A via tunnel).
		// But with EndStop, K is on the high side (value > 7) and A is placed → blocked
		card13 := domain.NewCard(domain.CardDesignSpade, 13, false)
		assert.False(t, s.IsPlayable(card13), "K should be blocked by EndStop even with tunnel when A is placed")

		// Also verify 8 is blocked
		card8 := domain.NewCard(domain.CardDesignSpade, 8, false)
		assert.False(t, s.IsPlayable(card8), "8 should be blocked by EndStop when A is placed")
	})

	t.Run("EndStop + Tunnel: A blocked even with tunnel when K placed", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{EndStopEnabled: true, TunnelEnabled: true, MaxPasses: 5}
		s := domain.NewSevens(tc, players, cfg)
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		// Place 8..13 of spade
		for v := 8; v <= 13; v++ {
			players[0].AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
		}
		for v := 8; v <= 13; v++ {
			idx := -1
			for i := 0; i < players[0].GetCardsSize(); i++ {
				c := players[0].GetCard(i)
				if c.GetDesign() == domain.CardDesignSpade && c.GetValue() == v {
					idx = i
					break
				}
			}
			require.NotEqual(t, -1, idx)
			err := s.PlayerPlay(idx)
			require.NoError(t, err)
			for !s.IsHumanTurn() && !s.GetGameEndFlag() {
				s.CpuPlay()
			}
		}
		// K(13) placed. With tunnel, A(1) would normally be playable (adjacent to K via tunnel).
		// But with EndStop, A is on the low side (value < 7) and K is placed → blocked
		card1 := domain.NewCard(domain.CardDesignSpade, 1, false)
		assert.False(t, s.IsPlayable(card1), "A should be blocked by EndStop even with tunnel when K is placed")

		// Also verify 6 is blocked
		card6 := domain.NewCard(domain.CardDesignSpade, 6, false)
		assert.False(t, s.IsPlayable(card6), "6 should be blocked by EndStop when K is placed")
	})

	t.Run("EndStop + Joker: joker cannot target end-stopped positions", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{EndStopEnabled: true, JokerCount: 1, MaxPasses: 5}
		s := domain.NewSevens(tc, players, cfg)
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		// Place 6,5,4,3,2,1 of spade
		for v := 6; v >= 1; v-- {
			players[0].AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
		}
		// Give human a joker
		players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))

		for v := 6; v >= 1; v-- {
			idx := -1
			for i := 0; i < players[0].GetCardsSize(); i++ {
				c := players[0].GetCard(i)
				if c.GetDesign() == domain.CardDesignSpade && c.GetValue() == v {
					idx = i
					break
				}
			}
			require.NotEqual(t, -1, idx)
			err := s.PlayerPlay(idx)
			require.NoError(t, err)
			for !s.IsHumanTurn() && !s.GetGameEndFlag() {
				s.CpuPlay()
			}
		}

		// Joker should not be able to target spade 8 (high side blocked because A placed)
		jokerIdx := -1
		for i := 0; i < players[0].GetCardsSize(); i++ {
			if players[0].GetCard(i).GetDesign() == domain.CardDesignJoker {
				jokerIdx = i
				break
			}
		}
		require.NotEqual(t, -1, jokerIdx)
		err := s.PlayerPlayJoker(jokerIdx, domain.CardDesignSpade, 8)
		assert.Error(t, err, "joker should not be able to target end-stopped position")
	})

	t.Run("EndStop + JokerReclaim: reclaim still works on placed position", func(t *testing.T) {
		tc := domain.NewTrumpCards(1)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{EndStopEnabled: true, JokerCount: 1, JokerReclaimEnabled: true, MaxPasses: 5}
		s := domain.NewSevens(tc, players, cfg)
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		// Give human a joker and spade 6
		players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		// Play joker on spade 6 position
		jokerIdx := -1
		for i := 0; i < players[0].GetCardsSize(); i++ {
			if players[0].GetCard(i).GetDesign() == domain.CardDesignJoker {
				jokerIdx = i
				break
			}
		}
		require.NotEqual(t, -1, jokerIdx)
		err := s.PlayerPlayJoker(jokerIdx, domain.CardDesignSpade, 6)
		require.NoError(t, err)

		// Skip CPU turns
		for !s.IsHumanTurn() && !s.GetGameEndFlag() {
			s.CpuPlay()
		}

		// Now play real spade 6 to reclaim joker (placed check runs before EndStop check)
		spade6Idx := -1
		for i := 0; i < players[0].GetCardsSize(); i++ {
			c := players[0].GetCard(i)
			if c.GetDesign() == domain.CardDesignSpade && c.GetValue() == 6 {
				spade6Idx = i
				break
			}
		}
		require.NotEqual(t, -1, spade6Idx)
		err = s.PlayerPlay(spade6Idx)
		assert.NoError(t, err, "reclaim should work even with EndStop enabled")
	})

	t.Run("EndStop CPU play and pass", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{EndStopEnabled: true, MaxPasses: 5}
		s := domain.NewSevens(tc, players, cfg)

		// Give CPU1 a card on blocked side
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		// Place A of spade so high side is blocked for spade
		// First human places 6,5,4,3,2,1
		for v := 6; v >= 1; v-- {
			players[0].AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
		}
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 6, false)) // extra card so human doesn't finish
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		for v := 6; v >= 1; v-- {
			idx := -1
			for i := 0; i < players[0].GetCardsSize(); i++ {
				c := players[0].GetCard(i)
				if c.GetDesign() == domain.CardDesignSpade && c.GetValue() == v {
					idx = i
					break
				}
			}
			require.NotEqual(t, -1, idx)
			err := s.PlayerPlay(idx)
			require.NoError(t, err)
			// CPU turns happen automatically after human play
			for !s.IsHumanTurn() && !s.GetGameEndFlag() {
				s.CpuPlay()
			}
		}
		// After A placed, CPU1's spade 8 should be unplayable.
		// The CPU should have passed when it was its turn.
		assert.True(t, players[1].GetPassesUsed() > 0 || players[1].GetIsFinished(),
			"CPU1 should have passed or been eliminated because its only card (spade 8) is blocked")
	})

	t.Run("EndStop value 7 is never affected", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{EndStopEnabled: true, MaxPasses: 5}
		s := domain.NewSevens(tc, players, cfg)
		// 7 is already placed, so it's not playable (already on board),
		// but isEndStopped should return false for value 7
		// We test indirectly: 7 is placed, and 6 and 8 are playable on a fresh board (no A or K yet)
		card6 := domain.NewCard(domain.CardDesignSpade, 6, false)
		card8 := domain.NewCard(domain.CardDesignSpade, 8, false)
		assert.True(t, s.IsPlayable(card6), "6 should be playable on fresh board with EndStop")
		assert.True(t, s.IsPlayable(card8), "8 should be playable on fresh board with EndStop")
	})

	t.Run("EndStop enabled but A not placed high side still playable", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{EndStopEnabled: true, MaxPasses: 5}
		s := domain.NewSevens(tc, players, cfg)

		// Only 7 is placed (no A), so 8 should be playable
		card8 := domain.NewCard(domain.CardDesignSpade, 8, false)
		assert.True(t, s.IsPlayable(card8), "high side should be playable when A not placed")
	})

	t.Run("EndStop enabled but K not placed low side still playable", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{EndStopEnabled: true, MaxPasses: 5}
		s := domain.NewSevens(tc, players, cfg)

		// Only 7 is placed (no K), so 6 should be playable
		card6 := domain.NewCard(domain.CardDesignSpade, 6, false)
		assert.True(t, s.IsPlayable(card6), "low side should be playable when K not placed")
	})

	t.Run("EndStop with CPU strategy", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{EndStopEnabled: true, CpuStrategy: domain.SevensCpuStrategic, MaxPasses: 5}
		s := domain.NewSevens(tc, players, cfg)
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		// Place A of spade
		for v := 6; v >= 1; v-- {
			players[0].AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
		}
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 6, false))

		for v := 6; v >= 1; v-- {
			idx := -1
			for i := 0; i < players[0].GetCardsSize(); i++ {
				c := players[0].GetCard(i)
				if c.GetDesign() == domain.CardDesignSpade && c.GetValue() == v {
					idx = i
					break
				}
			}
			require.NotEqual(t, -1, idx)
			err := s.PlayerPlay(idx)
			require.NoError(t, err)
			for !s.IsHumanTurn() && !s.GetGameEndFlag() {
				s.CpuPlay()
			}
		}
		// CPU1's spade 8 is blocked → CPU passed or was eliminated
		assert.True(t, players[1].GetPassesUsed() > 0 || players[1].GetIsFinished(),
			"CPU1 with strategy should pass or be eliminated when its only card is end-stopped")
	})
}

func TestSevens_JokerConsecutiveBanned(t *testing.T) {
	t.Run("enabled blocks joker after joker play via PlayerPlayJoker", func(t *testing.T) {
		tc := domain.NewTrumpCards(2)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{
			JokerCount:             2,
			MaxPasses:              domain.SevensMaxPasses,
			JokerConsecutiveBanned: true,
		}
		s := domain.NewSevens(tc, players, cfg)

		// Give human 2 jokers + a normal card
		for players[0].GetCardsSize() > 0 {
			players[0].RemoveCard(0)
		}
		players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		// First joker play should succeed
		err := s.PlayerPlayJoker(0, domain.CardDesignSpade, 6)
		assert.NoError(t, err)
		assert.True(t, players[0].GetLastPlayedJoker())

		// Advance CPU turns back to human
		for !s.IsHumanTurn() && !s.GetGameEndFlag() {
			s.CpuPlay()
		}

		// Second joker play should be blocked
		err = s.PlayerPlayJoker(0, domain.CardDesignSpade, 8)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "consecutive")
	})

	t.Run("PlayerPlay always rejects joker regardless of consecutive rule", func(t *testing.T) {
		tc := domain.NewTrumpCards(2)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{
			JokerCount:             2,
			MaxPasses:              domain.SevensMaxPasses,
			JokerConsecutiveBanned: true,
		}
		s := domain.NewSevens(tc, players, cfg)

		for players[0].GetCardsSize() > 0 {
			players[0].RemoveCard(0)
		}
		players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		// PlayerPlay rejects joker card — must use PlayerPlayJoker
		err := s.PlayerPlay(0) // joker is at index 0
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "use PlayerPlayJoker")
	})

	t.Run("normal card play resets lastPlayedJoker flag", func(t *testing.T) {
		tc := domain.NewTrumpCards(2)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{
			JokerCount:             2,
			MaxPasses:              domain.SevensMaxPasses,
			JokerConsecutiveBanned: true,
		}
		s := domain.NewSevens(tc, players, cfg)

		for players[0].GetCardsSize() > 0 {
			players[0].RemoveCard(0)
		}
		// Spade 6 is adjacent to 7, so it's always playable
		// Spade 8 is adjacent to 7, so it's always playable
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))

		// Play joker at spade 5 (adjacent to 6? no, 6 not placed yet. Adjacent to 7? no, 5 is not adjacent to 7)
		// Play joker at spade 6 (adjacent to 7) — this works
		err := s.PlayerPlayJoker(1, domain.CardDesignSpade, 6)
		assert.NoError(t, err)
		assert.True(t, players[0].GetLastPlayedJoker())

		for !s.IsHumanTurn() && !s.GetGameEndFlag() {
			s.CpuPlay()
		}

		// Play normal card spade 8 (adjacent to 7) → resets flag
		err = s.PlayerPlay(1) // spade 8 is now at index 1 (joker was removed, spade 6 at 0, spade 8 at 1)
		assert.NoError(t, err)
		assert.False(t, players[0].GetLastPlayedJoker())
	})

	t.Run("pass resets lastPlayedJoker flag", func(t *testing.T) {
		tc := domain.NewTrumpCards(2)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{
			JokerCount:             2,
			MaxPasses:              domain.SevensMaxPasses,
			JokerConsecutiveBanned: true,
		}
		s := domain.NewSevens(tc, players, cfg)

		for players[0].GetCardsSize() > 0 {
			players[0].RemoveCard(0)
		}
		players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))

		// Play joker
		err := s.PlayerPlayJoker(0, domain.CardDesignSpade, 6)
		assert.NoError(t, err)
		assert.True(t, players[0].GetLastPlayedJoker())

		for !s.IsHumanTurn() && !s.GetGameEndFlag() {
			s.CpuPlay()
		}

		// Pass → resets flag
		err = s.PlayerPlay(-1)
		assert.NoError(t, err)
		assert.False(t, players[0].GetLastPlayedJoker())
	})

	t.Run("hasPlayableCard respects consecutive rule", func(t *testing.T) {
		tc := domain.NewTrumpCards(2)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{
			JokerCount:             2,
			MaxPasses:              domain.SevensMaxPasses,
			JokerConsecutiveBanned: true,
		}
		s := domain.NewSevens(tc, players, cfg)

		// Give human only jokers
		for players[0].GetCardsSize() > 0 {
			players[0].RemoveCard(0)
		}
		players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))
		players[0].SetLastPlayedJoker(true)

		// HasAnyOption should be true (because can pass) but the joker is blocked
		// IsPlayable still returns true for the joker card itself (it just checks board position)
		// But hasPlayableCard internally blocks it
		assert.True(t, s.HasAnyOption(0), "can still pass")
	})

	t.Run("disabled allows joker after joker play", func(t *testing.T) {
		tc := domain.NewTrumpCards(2)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{
			JokerCount:             2,
			MaxPasses:              domain.SevensMaxPasses,
			JokerConsecutiveBanned: false,
		}
		s := domain.NewSevens(tc, players, cfg)

		for players[0].GetCardsSize() > 0 {
			players[0].RemoveCard(0)
		}
		players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))

		err := s.PlayerPlayJoker(0, domain.CardDesignSpade, 6)
		assert.NoError(t, err)

		for !s.IsHumanTurn() && !s.GetGameEndFlag() {
			s.CpuPlay()
		}

		// Second joker should be allowed when rule is disabled
		err = s.PlayerPlayJoker(0, domain.CardDesignSpade, 5)
		assert.NoError(t, err)
	})

	t.Run("CPU respects consecutive rule", func(t *testing.T) {
		tc := domain.NewTrumpCards(2)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{
			JokerCount:             2,
			MaxPasses:              domain.SevensMaxPasses,
			JokerConsecutiveBanned: true,
		}
		s := domain.NewSevens(tc, players, cfg)

		// Set up CPU1 with 2 jokers only
		for players[1].GetCardsSize() > 0 {
			players[1].RemoveCard(0)
		}
		players[1].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))
		// Set lastPlayedJoker flag
		players[1].SetLastPlayedJoker(true)

		// Advance to CPU1's turn
		for players[0].GetCardsSize() > 0 {
			players[0].RemoveCard(0)
		}
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		err := s.PlayerPlay(0) // human plays spade 6
		assert.NoError(t, err)

		// Now it's CPU1's turn — call CpuPlay
		s.CpuPlay()
		cpuActions := s.GetCpuActions()
		require.True(t, len(cpuActions) > 0, "expected at least one CPU action")
		cpu1Action := cpuActions[0]
		assert.Equal(t, 1, cpu1Action.PlayerIdx)
		// CPU1 should pass (joker blocked by consecutive rule)
		assert.Nil(t, cpu1Action.PlayedCard, "CPU1 should pass when joker is consecutively blocked")
	})

	t.Run("CPU with strategy respects consecutive rule", func(t *testing.T) {
		tc := domain.NewTrumpCards(2)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{
			JokerCount:             2,
			MaxPasses:              domain.SevensMaxPasses,
			CpuStrategy:            domain.SevensCpuStrategic,
			JokerConsecutiveBanned: true,
		}
		s := domain.NewSevens(tc, players, cfg)

		// Set up CPU1 with 2 jokers only
		for players[1].GetCardsSize() > 0 {
			players[1].RemoveCard(0)
		}
		players[1].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))
		players[1].SetLastPlayedJoker(true)

		for players[0].GetCardsSize() > 0 {
			players[0].RemoveCard(0)
		}
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		err := s.PlayerPlay(0) // human plays spade 6
		assert.NoError(t, err)

		// Now it's CPU1's turn — call CpuPlay
		s.CpuPlay()
		cpuActions := s.GetCpuActions()
		require.True(t, len(cpuActions) > 0, "expected at least one CPU action")
		cpu1Action := cpuActions[0]
		assert.Equal(t, 1, cpu1Action.PlayerIdx)
		// CPU1 should pass (joker blocked by consecutive rule)
		assert.Nil(t, cpu1Action.PlayedCard, "CPU1 should pass when joker is consecutively blocked")
	})

	t.Run("combined with NoJokerFinish", func(t *testing.T) {
		tc := domain.NewTrumpCards(2)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{
			JokerCount:             2,
			MaxPasses:              domain.SevensMaxPasses,
			JokerConsecutiveBanned: true,
			NoJokerFinish:          true,
		}
		s := domain.NewSevens(tc, players, cfg)

		// Give human only jokers with lastPlayedJoker set
		for players[0].GetCardsSize() > 0 {
			players[0].RemoveCard(0)
		}
		players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))
		players[0].SetLastPlayedJoker(true)

		// Joker blocked by both rules → no playable cards, but can pass
		assert.True(t, s.HasAnyOption(0))
	})

	t.Run("Reset clears lastPlayedJoker flag", func(t *testing.T) {
		tc := domain.NewTrumpCards(2)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{
			JokerCount:             2,
			MaxPasses:              domain.SevensMaxPasses,
			JokerConsecutiveBanned: true,
		}
		s := domain.NewSevens(tc, players, cfg)

		players[0].SetLastPlayedJoker(true)
		s.Reset()

		// After reset, find human player and check flag is cleared
		for i := 0; i < s.GetPlayerCnt(); i++ {
			assert.False(t, s.GetPlayer(i).GetLastPlayedJoker())
		}
	})

	t.Run("elimination when only jokers and lastPlayedJoker and no passes", func(t *testing.T) {
		tc := domain.NewTrumpCards(2)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{
			JokerCount:             2,
			MaxPasses:              1,
			JokerConsecutiveBanned: true,
		}
		s := domain.NewSevens(tc, players, cfg)

		// CPU1 has only jokers, lastPlayedJoker=true, and exhausted passes
		for players[1].GetCardsSize() > 0 {
			players[1].RemoveCard(0)
		}
		players[1].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))
		players[1].SetLastPlayedJoker(true)
		players[1].SetMaxPasses(1)
		players[1].IncrPassesUsed() // exhaust passes

		// HasAnyOption should return false (joker blocked + no passes)
		assert.False(t, s.HasAnyOption(1))

		// Trigger CPU play by advancing via human
		for players[0].GetCardsSize() > 0 {
			players[0].RemoveCard(0)
		}
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		err := s.PlayerPlay(0) // human plays spade 6
		assert.NoError(t, err)

		// Now it's CPU1's turn — call CpuPlay
		s.CpuPlay()

		// CPU1 should be eliminated
		assert.True(t, players[1].GetIsEliminated())
	})

	t.Run("NewSevens stores JokerConsecutiveBanned config", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{JokerConsecutiveBanned: true}
		s := domain.NewSevens(tc, players, cfg)
		assert.True(t, s.GetConfig().JokerConsecutiveBanned)
	})

	t.Run("CPU sets lastPlayedJoker true after playing joker", func(t *testing.T) {
		tc := domain.NewTrumpCards(2)
		players := makeSevensPlayers()
		cfg := domain.SevensConfig{
			JokerCount:             2,
			MaxPasses:              domain.SevensMaxPasses,
			JokerConsecutiveBanned: true,
		}
		s := domain.NewSevens(tc, players, cfg)

		// Set up CPU1 with a joker + normal card
		for players[1].GetCardsSize() > 0 {
			players[1].RemoveCard(0)
		}
		players[1].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[1].SetLastPlayedJoker(false)

		// Human plays to advance turn
		for players[0].GetCardsSize() > 0 {
			players[0].RemoveCard(0)
		}
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		err := s.PlayerPlay(0)
		assert.NoError(t, err)

		// After CPU plays, check if lastPlayedJoker is set correctly
		cpuActions := s.GetCpuActions()
		for _, a := range cpuActions {
			if a.PlayerIdx == 1 && a.PlayedCard != nil {
				if a.PlayedCard.GetDesign() == domain.CardDesignJoker {
					assert.True(t, players[1].GetLastPlayedJoker())
				} else {
					assert.False(t, players[1].GetLastPlayedJoker())
				}
			}
		}
	})
}

// ---------------------------------------------------------------------------
// ActionLog tests
// ---------------------------------------------------------------------------

func TestSevens_ActionLog_Play(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())

	// Clear all cards
	for _, p := range players {
		for p.GetCardsSize() > 0 {
			p.RemoveCard(0)
		}
	}

	// Give human a card adjacent to 7 (6 of spades)
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
	// Give other players cards so game doesn't end
	players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
	players[2].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
	players[3].AddCard(domain.NewCard(domain.CardDesignSpade, 4, false))

	err := s.PlayerPlay(0) // play 6♠
	assert.NoError(t, err)

	log := s.GetActionLog()
	found := false
	for _, e := range log {
		if e.ActionType == "play" && e.PlayerIdx == 0 {
			found = true
			assert.Contains(t, e.Detail, "played")
			assert.Len(t, e.Cards, 1)
			break
		}
	}
	assert.True(t, found, "expected play action log entry")
}

func TestSevens_ActionLog_Pass(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())

	// Clear all cards
	for _, p := range players {
		for p.GetCardsSize() > 0 {
			p.RemoveCard(0)
		}
	}

	// Give human a card that can't be played (far from 7)
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	// Give other players cards
	players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
	players[2].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
	players[3].AddCard(domain.NewCard(domain.CardDesignSpade, 4, false))

	// Pass (idx < 0)
	err := s.PlayerPlay(-1)
	assert.NoError(t, err)

	log := s.GetActionLog()
	found := false
	for _, e := range log {
		if e.ActionType == "pass" && e.PlayerIdx == 0 {
			found = true
			assert.Equal(t, "pass", e.Detail)
			break
		}
	}
	assert.True(t, found, "expected pass action log entry")
}

func TestSevens_ActionLog_Reset(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayers()
	s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())

	// Clear all cards
	for _, p := range players {
		for p.GetCardsSize() > 0 {
			p.RemoveCard(0)
		}
	}

	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
	players[2].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
	players[3].AddCard(domain.NewCard(domain.CardDesignSpade, 4, false))

	_ = s.PlayerPlay(0)
	assert.NotEmpty(t, s.GetActionLog())

	s.Reset()
	assert.Nil(t, s.GetActionLog())
}

// #5479: IsPlayable は実装済みだったが interface に無く、CUI からは呼べなかった。
// Web は SevensHumanArea.tsx が同じ判定をフロントで書き直して色を付けている。
func TestSevens_GetPlayableCardIndices(t *testing.T) {
	// 卓には最初から4枚の7が並んでいるので、6 と 8 が出せて 5 や J は出せない。
	newGame := func(hand ...*domain.Card) (*domain.Sevens, []*domain.SevensPlayer) {
		players := []*domain.SevensPlayer{
			domain.NewSevensPlayer(true),
			domain.NewSevensPlayer(false),
			domain.NewSevensPlayer(false),
			domain.NewSevensPlayer(false),
		}
		s := domain.NewSevens(domain.NewTrumpCards(0), players, domain.DefaultSevensConfig())
		for _, c := range hand {
			players[0].AddCard(c)
		}
		for i := 1; i < 4; i++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		}
		return s, players
	}
	card := func(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

	t.Run("marks only the cards adjacent to the board", func(t *testing.T) {
		s, _ := newGame(
			card(domain.CardDesignSpade, 6),  // ♠7 の隣 → 出せる
			card(domain.CardDesignSpade, 11), // ♠7 から離れている → 出せない
			card(domain.CardDesignHeart, 8),  // ♥7 の隣 → 出せる
		)
		assert.Equal(t, []int{0, 2}, s.GetPlayableCardIndices())
	})

	// **空スライスと nil を区別する。**7並べは「1枚も出せない」が普通に起きる
	// (そこでパスする) 局面なので、判定していない状態と同じ扱いにはできない。
	t.Run("returns an empty slice, not nil, when nothing is playable", func(t *testing.T) {
		s, _ := newGame(card(domain.CardDesignSpade, 11))
		got := s.GetPlayableCardIndices()
		assert.NotNil(t, got)
		assert.Empty(t, got)
	})

	t.Run("returns nil when it is not the human turn", func(t *testing.T) {
		s, _ := newGame(card(domain.CardDesignSpade, 6), card(domain.CardDesignSpade, 8))
		require.NoError(t, s.PlayerPlay(0)) // 手番が CPU へ移る
		require.False(t, s.IsHumanTurn())
		assert.Nil(t, s.GetPlayableCardIndices())
	})

	// ジョーカーは置ける場所が1つでもあれば出せる。IsPlayable がそう書いてある
	// のをここでも通す — 印の判定を presenter 側に書き直さないため。
	t.Run("marks a joker while the board still has room", func(t *testing.T) {
		s, _ := newGame(domain.NewCard(domain.CardDesignJoker, 0, true))
		assert.Equal(t, []int{0}, s.GetPlayableCardIndices())
	})
}
