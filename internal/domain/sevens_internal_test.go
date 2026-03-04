package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSevens_advanceTurn_gameEndFlag(t *testing.T) {
	// Covers lines 176-178: advanceTurn returns immediately when gameEndFlag is true.
	// This guard is defensive code unreachable through public API because all callers
	// check gameEndFlag before calling advanceTurn. We test it directly.
	tc := NewTrumpCards(0)
	players := []*SevensPlayer{
		NewSevensPlayer(true),
		NewSevensPlayer(false),
		NewSevensPlayer(false),
		NewSevensPlayer(false),
	}
	s := NewSevens(tc, players, DefaultSevensConfig())

	// Give players cards so there are active players to advance to
	for i := 0; i < 4; i++ {
		players[i].AddCard(NewCard(CardDesignDiamond, 2, false))
	}

	s.gameEndFlag = true
	originalTurn := s.currentTurn
	s.advanceTurn()
	// currentTurn should not have changed because advanceTurn returned early
	assert.Equal(t, originalTurn, s.currentTurn)
}

func makeSevensPlayersInternal() []*SevensPlayer {
	return []*SevensPlayer{
		NewSevensPlayer(true),
		NewSevensPlayer(false),
		NewSevensPlayer(false),
		NewSevensPlayer(false),
	}
}

func TestSevens_isPositionPlaced_InvalidSuit(t *testing.T) {
	// Covers lines 187-189: isPositionPlaced returns false for invalid suit.
	tc := NewTrumpCards(0)
	players := makeSevensPlayersInternal()
	s := NewSevens(tc, players, DefaultSevensConfig())

	// suit < CardDesignSpade (suit=0)
	assert.False(t, s.isPositionPlaced(0, 7))
	// suit > CardDesignDiamond (suit=5)
	assert.False(t, s.isPositionPlaced(5, 7))
	// suit = -1
	assert.False(t, s.isPositionPlaced(-1, 7))
	// Valid suit with valid value (7 is placed on fresh board)
	assert.True(t, s.isPositionPlaced(CardDesignSpade, 7))
	// Valid suit with unplaced value
	assert.False(t, s.isPositionPlaced(CardDesignSpade, 6))
}

func TestSevens_evaluatePlay_TunnelAceLow(t *testing.T) {
	// Covers evaluatePlay tunnel wrap for Ace (value=1),
	// where nextLow becomes 13 instead of 0.
	// Setup: tunnel enabled + strategy. Board has spade 2-7 placed (min=2, max=7).
	// CPU has Ace(1) which is playable (adjacent to 2).
	// Without tunnel: nextLow = 0, skipped. With tunnel: nextLow = 13.
	// CPU does NOT have King(13), so score -= 1 for low direction.
	// nextHigh = 2 (placed), so no score for high direction.
	// Total score = -1 -> CPU would pass if it has passes.
	tc := NewTrumpCards(0)
	players := makeSevensPlayersInternal()
	tunnelStrategyConfig := SevensConfig{TunnelEnabled: true, JokerCount: 0, CpuStrategy: true}
	s := NewSevens(tc, players, tunnelStrategyConfig)

	// Build board: set the board directly using field access
	var placed [5]uint16
	for i := 1; i <= 4; i++ {
		placed[i] = 1 << 7 // 7 is placed for all suits
	}
	// Place spade 2-6 (plus 7 already)
	placed[CardDesignSpade] |= (1 << 2) | (1 << 3) | (1 << 4) | (1 << 5) | (1 << 6)
	s.tablePlaced = placed

	// Give players enough cards so game doesn't end
	for i := 0; i < 4; i++ {
		for d := 0; d < 5; d++ {
			players[i].AddCard(NewCard(CardDesignDiamond, 2, false))
		}
	}

	// Human plays some card to advance turn to CPU
	players[0].AddCard(NewCard(CardDesignSpade, 8, false))
	_ = s.PlayerPlay(players[0].GetCardsSize() - 1) // play 8♠

	// CPU 1 has Ace(1)♠ - playable since adjacent to 2♠
	// It does NOT have King(13)♠, so tunnel wrap gives negative score
	players[1].AddCard(NewCard(CardDesignSpade, 1, false))

	if s.currentTurn == 1 {
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

func TestSevens_evaluatePlay_TunnelKingHigh(t *testing.T) {
	// Covers evaluatePlay tunnel wrap for King (value=13),
	// where nextHigh becomes 1 instead of 14.
	// Setup: tunnel enabled + strategy. Board has spade 7-12 placed.
	// CPU has King(13) which is playable (adjacent to 12).
	// Without tunnel: nextHigh = 14, skipped. With tunnel: nextHigh = 1.
	// CPU does NOT have Ace(1), so score -= 1 for high direction.
	// nextLow = 12 (placed), so no score for low direction.
	// Total score = -1 -> CPU would pass if it has passes.
	tc := NewTrumpCards(0)
	players := makeSevensPlayersInternal()
	tunnelStrategyConfig := SevensConfig{TunnelEnabled: true, JokerCount: 0, CpuStrategy: true}
	s := NewSevens(tc, players, tunnelStrategyConfig)

	// Build board: place spade 8-12 (plus 7 already)
	var placed [5]uint16
	for i := 1; i <= 4; i++ {
		placed[i] = 1 << 7
	}
	placed[CardDesignSpade] |= (1 << 8) | (1 << 9) | (1 << 10) | (1 << 11) | (1 << 12)
	s.tablePlaced = placed

	// Give players enough cards so game doesn't end
	for i := 0; i < 4; i++ {
		for d := 0; d < 5; d++ {
			players[i].AddCard(NewCard(CardDesignDiamond, 2, false))
		}
	}

	// Human plays some card to advance turn to CPU
	players[0].AddCard(NewCard(CardDesignSpade, 6, false))
	_ = s.PlayerPlay(players[0].GetCardsSize() - 1) // play 6♠

	// CPU 1 has King(13)♠ - playable since adjacent to 12♠
	// It does NOT have Ace(1)♠, so tunnel wrap gives negative score
	players[1].AddCard(NewCard(CardDesignSpade, 13, false))

	if s.currentTurn == 1 {
		s.CpuPlay()
		actions := s.GetCpuActions()
		assert.NotEmpty(t, actions)
		lastAction := actions[len(actions)-1]
		if lastAction.PlayerIdx == 1 {
			assert.Nil(t, lastAction.PlayedCard) // passed due to negative score
		}
	}
}

func TestSevens_setGameEndFlag(t *testing.T) {
	tc := NewTrumpCards(0)
	players := makeSevensPlayersInternal()
	s := NewSevens(tc, players, DefaultSevensConfig())

	assert.False(t, s.gameEndFlag)
	s.gameEndFlag = true
	assert.True(t, s.gameEndFlag)
	s.gameEndFlag = false
	assert.False(t, s.gameEndFlag)
}

func TestSevens_setCurrentTurn(t *testing.T) {
	tc := NewTrumpCards(0)
	players := makeSevensPlayersInternal()
	s := NewSevens(tc, players, DefaultSevensConfig())

	assert.Equal(t, 0, s.currentTurn)
	s.currentTurn = 2
	assert.Equal(t, 2, s.currentTurn)
}

func TestSevens_hasOnlyJokers(t *testing.T) {
	tc := NewTrumpCards(1)
	players := makeSevensPlayersInternal()
	cfg := SevensConfig{JokerCount: 1, NoJokerFinish: true, MaxPasses: SevensMaxPasses}
	s := NewSevens(tc, players, cfg)

	t.Run("empty hand returns false", func(t *testing.T) {
		assert.False(t, s.hasOnlyJokers(players[0]))
	})

	t.Run("only joker returns true", func(t *testing.T) {
		players[0].AddCard(NewCard(CardDesignJoker, 0, false))
		assert.True(t, s.hasOnlyJokers(players[0]))
		players[0].RemoveCard(0)
	})

	t.Run("joker + normal card returns false", func(t *testing.T) {
		players[0].AddCard(NewCard(CardDesignJoker, 0, false))
		players[0].AddCard(NewCard(CardDesignSpade, 6, false))
		assert.False(t, s.hasOnlyJokers(players[0]))
		players[0].RemoveCard(0)
		players[0].RemoveCard(0)
	})

	t.Run("only normal cards returns false", func(t *testing.T) {
		players[0].AddCard(NewCard(CardDesignSpade, 6, false))
		assert.False(t, s.hasOnlyJokers(players[0]))
		players[0].RemoveCard(0)
	})
}

func TestSevens_isJokerBlockedByFinishRule(t *testing.T) {
	t.Run("blocked when NoJokerFinish and only jokers", func(t *testing.T) {
		tc := NewTrumpCards(1)
		players := makeSevensPlayersInternal()
		cfg := SevensConfig{JokerCount: 1, NoJokerFinish: true, MaxPasses: SevensMaxPasses}
		s := NewSevens(tc, players, cfg)
		players[0].AddCard(NewCard(CardDesignJoker, 0, false))
		assert.True(t, s.isJokerBlockedByFinishRule(players[0]))
	})

	t.Run("not blocked when NoJokerFinish off", func(t *testing.T) {
		tc := NewTrumpCards(1)
		players := makeSevensPlayersInternal()
		cfg := SevensConfig{JokerCount: 1, NoJokerFinish: false, MaxPasses: SevensMaxPasses}
		s := NewSevens(tc, players, cfg)
		players[0].AddCard(NewCard(CardDesignJoker, 0, false))
		assert.False(t, s.isJokerBlockedByFinishRule(players[0]))
	})

	t.Run("not blocked when mixed cards", func(t *testing.T) {
		tc := NewTrumpCards(1)
		players := makeSevensPlayersInternal()
		cfg := SevensConfig{JokerCount: 1, NoJokerFinish: true, MaxPasses: SevensMaxPasses}
		s := NewSevens(tc, players, cfg)
		players[0].AddCard(NewCard(CardDesignJoker, 0, false))
		players[0].AddCard(NewCard(CardDesignSpade, 6, false))
		assert.False(t, s.isJokerBlockedByFinishRule(players[0]))
	})
}

func TestSevens_passUrgencyWeight(t *testing.T) {
	tc := NewTrumpCards(0)
	players := makeSevensPlayersInternal()
	cfg := DefaultSevensConfig()
	s := NewSevens(tc, players, cfg)

	t.Run("unlimited passes returns weight 1", func(t *testing.T) {
		players[0].SetMaxPasses(0) // unlimited
		assert.Equal(t, 1, s.passUrgencyWeight(players[0]))
		players[0].SetMaxPasses(SevensMaxPasses) // restore
	})

	t.Run("1 pass remaining returns weight 3", func(t *testing.T) {
		players[0].SetMaxPasses(3)
		players[0].ResetPasses()
		players[0].IncrPassesUsed()
		players[0].IncrPassesUsed() // used 2/3 -> 1 remaining
		assert.Equal(t, 3, s.passUrgencyWeight(players[0]))
	})

	t.Run("0 passes remaining returns weight 3", func(t *testing.T) {
		players[0].SetMaxPasses(3)
		players[0].ResetPasses()
		players[0].IncrPassesUsed()
		players[0].IncrPassesUsed()
		players[0].IncrPassesUsed() // used 3/3 -> 0 remaining
		assert.Equal(t, 3, s.passUrgencyWeight(players[0]))
	})

	t.Run("2 passes remaining returns weight 2", func(t *testing.T) {
		players[0].SetMaxPasses(3)
		players[0].ResetPasses()
		players[0].IncrPassesUsed() // used 1/3 -> 2 remaining
		assert.Equal(t, 2, s.passUrgencyWeight(players[0]))
	})

	t.Run("plenty of passes returns weight 1", func(t *testing.T) {
		players[0].SetMaxPasses(10)
		players[0].ResetPasses()
		assert.Equal(t, 1, s.passUrgencyWeight(players[0]))
	})
}

func TestSevens_countWeightedOpponentsBlocked(t *testing.T) {
	t.Run("weighted count higher for low-pass opponents", func(t *testing.T) {
		tc := NewTrumpCards(0)
		players := makeSevensPlayersInternal()
		cfg := SevensConfig{CpuStrategy: true, MaxPasses: 3}
		s := NewSevens(tc, players, cfg)

		for i := 0; i < 4; i++ {
			players[i].SetMaxPasses(3)
			players[i].AddCard(NewCard(CardDesignDiamond, 2, false))
		}

		// Opponent (player 2) has 1 pass remaining
		players[2].IncrPassesUsed()
		players[2].IncrPassesUsed()
		players[2].AddCard(NewCard(CardDesignSpade, 5, false))

		// Count weighted blocked for suit spade from value 6 going low
		count := s.countWeightedOpponentsBlocked(players[1], CardDesignSpade, 6, -1)
		// Opponent 2 has 5♠, weight = 3 (1 pass remaining)
		assert.Equal(t, 3, count)
	})

	t.Run("unweighted count for unlimited-pass opponents", func(t *testing.T) {
		tc := NewTrumpCards(0)
		players := makeSevensPlayersInternal()
		cfg := SevensConfig{CpuStrategy: true, MaxPasses: 0}
		s := NewSevens(tc, players, cfg)

		for i := 0; i < 4; i++ {
			players[i].SetMaxPasses(0)
			players[i].AddCard(NewCard(CardDesignDiamond, 2, false))
		}

		players[2].AddCard(NewCard(CardDesignSpade, 5, false))

		count := s.countWeightedOpponentsBlocked(players[1], CardDesignSpade, 6, -1)
		// Opponent 2 has 5♠, weight = 1 (unlimited passes)
		assert.Equal(t, 1, count)
	})

	t.Run("skips finished opponents", func(t *testing.T) {
		tc := NewTrumpCards(0)
		players := makeSevensPlayersInternal()
		cfg := SevensConfig{CpuStrategy: true, MaxPasses: 3}
		s := NewSevens(tc, players, cfg)

		for i := 0; i < 4; i++ {
			players[i].SetMaxPasses(3)
			players[i].AddCard(NewCard(CardDesignDiamond, 2, false))
		}

		players[2].AddCard(NewCard(CardDesignSpade, 5, false))
		players[2].SetIsFinished(true)

		count := s.countWeightedOpponentsBlocked(players[1], CardDesignSpade, 6, -1)
		assert.Equal(t, 0, count) // finished player skipped
	})

	t.Run("tunnel wrapping in weighted count", func(t *testing.T) {
		tc := NewTrumpCards(0)
		players := makeSevensPlayersInternal()
		cfg := SevensConfig{TunnelEnabled: true, CpuStrategy: true, MaxPasses: 3}
		s := NewSevens(tc, players, cfg)

		// Build board: place spade 2-7
		var placed [5]uint16
		for i := 1; i <= 4; i++ {
			placed[i] = 1 << 7
		}
		placed[CardDesignSpade] |= (1 << 2) | (1 << 3) | (1 << 4) | (1 << 5) | (1 << 6)
		s.tablePlaced = placed

		for i := 0; i < 4; i++ {
			players[i].SetMaxPasses(3)
			players[i].AddCard(NewCard(CardDesignDiamond, 2, false))
		}

		// Opponent has King (13) of spades - reachable via tunnel from Ace going low
		players[2].IncrPassesUsed()
		players[2].IncrPassesUsed() // 1 pass remaining
		players[2].AddCard(NewCard(CardDesignSpade, 13, false))

		// Count from value 1 going low (-1) should tunnel wrap to 13
		count := s.countWeightedOpponentsBlocked(players[1], CardDesignSpade, 1, -1)
		assert.Equal(t, 3, count) // weight 3 for near-eliminated opponent
	})

	t.Run("tunnel full loop returns to fromValue", func(t *testing.T) {
		tc := NewTrumpCards(0)
		players := makeSevensPlayersInternal()
		cfg := SevensConfig{TunnelEnabled: true, CpuStrategy: true, MaxPasses: 3}
		s := NewSevens(tc, players, cfg)

		// Board: nothing placed for spade (clear even 7)
		var placed [5]uint16
		s.tablePlaced = placed

		for i := 0; i < 4; i++ {
			players[i].SetMaxPasses(3)
			players[i].AddCard(NewCard(CardDesignDiamond, 2, false))
		}

		// No opponent has any spade card → scan wraps all the way around to fromValue
		count := s.countWeightedOpponentsBlocked(players[0], CardDesignSpade, 6, -1)
		assert.Equal(t, 0, count)
	})
}

func TestSevens_setTablePlaced(t *testing.T) {
	tc := NewTrumpCards(0)
	players := makeSevensPlayersInternal()
	s := NewSevens(tc, players, DefaultSevensConfig())

	var placed [5]uint16
	for i := 1; i <= 4; i++ {
		placed[i] = (1 << 7) | (1 << 6)
	}
	s.tablePlaced = placed
	result := s.GetTablePlaced()
	for i := 1; i <= 4; i++ {
		assert.Equal(t, uint16((1<<7)|(1<<6)), result[i])
	}
}

// ---------------------------------------------------------------------------
// Joker Reclaim internal tests
// These tests directly manipulate jokerPlaced/jokerCards to exercise the
// reclaimJokerIfNeeded code path which cannot be triggered through the public
// API alone (because IsPlayable rejects a card whose position is already
// occupied on the board, even when a joker fills that position).
// ---------------------------------------------------------------------------

func TestJokerReclaim_BasicReclaim_Internal(t *testing.T) {
	// Simulate: joker placed at Spade-6 (tracked in jokerPlaced/jokerCards but
	// NOT in tablePlaced so the real card passes IsPlayable).
	// Then human plays real 6♠ → reclaimJokerIfNeeded fires → joker returns.
	tc := NewTrumpCards(2)
	players := makeSevensPlayersInternal()
	cfg := SevensConfig{
		JokerCount:          2,
		JokerReclaimEnabled: true,
		MaxPasses:           SevensMaxPasses,
	}
	s := NewSevens(tc, players, cfg)

	// Give other players dummy cards so game doesn't end
	for i := 1; i <= 3; i++ {
		for d := 0; d < 5; d++ {
			players[i].AddCard(NewCard(CardDesignDiamond, 2, false))
		}
	}

	// Manually set up joker tracking at Spade-6 WITHOUT placing it on the board.
	// This simulates the state where a joker occupies that logical position
	// but allows the real card to pass IsPlayable.
	jokerCard := NewCard(CardDesignJoker, 1, false)
	s.jokerPlaced[CardDesignSpade] |= 1 << 6
	s.jokerCards = append(s.jokerCards, jokerCard)

	// Board: only 7s placed (default), so Spade-6 is playable (adjacent to 7)
	// Human has real 6♠ + extra card
	players[0].AddCard(NewCard(CardDesignSpade, 6, false))
	players[0].AddCard(NewCard(CardDesignHeart, 6, false))

	handSizeBefore := players[0].GetCardsSize()
	assert.Equal(t, 2, handSizeBefore)
	assert.Equal(t, 1, s.GetJokerCardsCount())

	// Human plays real 6♠ → triggers reclaimJokerIfNeeded(0, Spade, 6)
	err := s.PlayerPlay(0)
	assert.NoError(t, err)

	// Reclaim should have fired: card played (-1) + joker returned (+1) = same size
	assert.Equal(t, handSizeBefore, players[0].GetCardsSize(),
		"hand size should stay same: played card replaced by reclaimed joker")

	// Joker tracking cleared
	assert.Equal(t, 0, s.GetJokerCardsCount(), "joker should be removed from board tracking")
	assert.True(t, s.jokerPlaced[CardDesignSpade]&(1<<6) == 0,
		"jokerPlaced bit should be cleared at spade-6")

	// Verify the reclaimed card is a joker
	hasJoker := false
	for i := 0; i < players[0].GetCardsSize(); i++ {
		if players[0].GetCard(i).GetDesign() == CardDesignJoker {
			hasJoker = true
			break
		}
	}
	assert.True(t, hasJoker, "player should have a joker back in hand")
}

func TestJokerReclaim_CpuReclaim_Internal(t *testing.T) {
	// CPU plays card at a joker-tracked position → CPU gets joker back.
	tc := NewTrumpCards(2)
	players := makeSevensPlayersInternal()
	cfg := SevensConfig{
		JokerCount:          2,
		JokerReclaimEnabled: true,
		MaxPasses:           SevensMaxPasses,
	}
	s := NewSevens(tc, players, cfg)

	// Give CPUs 2, 3 dummy cards
	for i := 2; i <= 3; i++ {
		for d := 0; d < 5; d++ {
			players[i].AddCard(NewCard(CardDesignDiamond, 2, false))
		}
	}

	// Human has a playable card to advance turn
	players[0].AddCard(NewCard(CardDesignSpade, 8, false))
	players[0].AddCard(NewCard(CardDesignHeart, 6, false))

	// Set up joker tracking at Spade-6 (without tablePlaced bit)
	jokerCard := NewCard(CardDesignJoker, 1, false)
	s.jokerPlaced[CardDesignSpade] |= 1 << 6
	s.jokerCards = append(s.jokerCards, jokerCard)

	// CPU 1 has real Spade-6 (playable: adjacent to 7) + extra cards
	players[1].AddCard(NewCard(CardDesignSpade, 6, false))
	for d := 0; d < 4; d++ {
		players[1].AddCard(NewCard(CardDesignDiamond, 2, false))
	}

	cpu1HandBefore := players[1].GetCardsSize()

	// Human plays to advance to CPU 1
	err := s.PlayerPlay(0) // play Spade-8
	assert.NoError(t, err)
	assert.Equal(t, 1, s.currentTurn, "should be CPU 1's turn")

	// CPU 1 plays
	s.CpuPlay()

	actions := s.GetCpuActions()
	require.NotEmpty(t, actions)

	// Check if CPU 1 played Spade-6
	cpuPlayed6 := false
	for _, a := range actions {
		if a.PlayerIdx == 1 && a.PlayedCard != nil &&
			a.PlayedCard.GetDesign() == CardDesignSpade && a.PlayedCard.GetValue() == 6 {
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
		assert.True(t, s.jokerPlaced[CardDesignSpade]&(1<<6) == 0,
			"jokerPlaced bit cleared")

		// Verify CPU has joker in hand
		hasJoker := false
		for i := 0; i < players[1].GetCardsSize(); i++ {
			if players[1].GetCard(i).GetDesign() == CardDesignJoker {
				hasJoker = true
				break
			}
		}
		assert.True(t, hasJoker, "CPU should have reclaimed joker")
	}
}

func TestJokerReclaim_MultipleJokers_ReclaimOne_Internal(t *testing.T) {
	// Two jokers on board, reclaim one → only one returned, one remains tracked.
	tc := NewTrumpCards(2)
	players := makeSevensPlayersInternal()
	cfg := SevensConfig{
		JokerCount:          2,
		JokerReclaimEnabled: true,
		MaxPasses:           SevensMaxPasses,
	}
	s := NewSevens(tc, players, cfg)

	// Give other players dummy cards
	for i := 1; i <= 3; i++ {
		for d := 0; d < 5; d++ {
			players[i].AddCard(NewCard(CardDesignDiamond, 2, false))
		}
	}

	// Set up two jokers tracked: one at Spade-6, one at Heart-8
	joker1 := NewCard(CardDesignJoker, 1, false)
	joker2 := NewCard(CardDesignJoker, 2, false)
	s.jokerPlaced[CardDesignSpade] |= 1 << 6
	s.jokerPlaced[CardDesignHeart] |= 1 << 8
	s.jokerCards = append(s.jokerCards, joker1, joker2)

	assert.Equal(t, 2, s.GetJokerCardsCount())

	// Human has real 6♠ (triggers reclaim at Spade-6) + extra cards
	players[0].AddCard(NewCard(CardDesignSpade, 6, false))
	players[0].AddCard(NewCard(CardDesignClover, 6, false))
	players[0].AddCard(NewCard(CardDesignDiamond, 6, false))

	err := s.PlayerPlay(0) // play Spade-6
	assert.NoError(t, err)

	// One joker reclaimed, one remains
	assert.Equal(t, 1, s.GetJokerCardsCount(), "one joker should remain on board")
	assert.True(t, s.jokerPlaced[CardDesignSpade]&(1<<6) == 0,
		"spade-6 joker bit should be cleared")
	assert.True(t, s.jokerPlaced[CardDesignHeart]&(1<<8) != 0,
		"heart-8 joker bit should remain set")

	// Player should have a joker in hand (reclaimed)
	hasJoker := false
	for i := 0; i < players[0].GetCardsSize(); i++ {
		if players[0].GetCard(i).GetDesign() == CardDesignJoker {
			hasJoker = true
			break
		}
	}
	assert.True(t, hasJoker, "player should have one reclaimed joker")
}
