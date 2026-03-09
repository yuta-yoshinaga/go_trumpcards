package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// GetJokerPlaced ジョーカー配置ビットマスク取得（テスト用）
func (s *Sevens) GetJokerPlaced() [5]uint16 { return s.jokerPlaced }

// GetJokerCardsCount ボード上のジョーカーカード数取得（テスト用）
func (s *Sevens) GetJokerCardsCount() int { return len(s.jokerCards) }

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
// These tests exercise edge cases that require internal struct manipulation.
// The main round-trip tests are in Sevens_test.go (public API).
// ---------------------------------------------------------------------------

func TestJokerReclaim_JokerCardsEmpty_DefensiveCheck_Internal(t *testing.T) {
	// Edge case: jokerPlaced bit is set but jokerCards slice is empty.
	// reclaimJokerIfNeeded should clear the bit but not crash.
	tc := NewTrumpCards(2)
	players := makeSevensPlayersInternal()
	cfg := SevensConfig{
		JokerCount:          2,
		JokerReclaimEnabled: true,
		MaxPasses:           SevensMaxPasses,
	}
	s := NewSevens(tc, players, cfg)

	for i := 1; i <= 3; i++ {
		for d := 0; d < 5; d++ {
			players[i].AddCard(NewCard(CardDesignDiamond, 2, false))
		}
	}

	// Manually set jokerPlaced bit WITHOUT adding to jokerCards (out-of-sync state)
	s.jokerPlaced[CardDesignSpade] |= 1 << 6
	// Also set tablePlaced so the position is "placed" on the board
	s.tablePlaced[CardDesignSpade] |= 1 << 6

	// Human has real Spade-6 + extra
	players[0].AddCard(NewCard(CardDesignSpade, 6, false))
	players[0].AddCard(NewCard(CardDesignHeart, 6, false))

	handSizeBefore := players[0].GetCardsSize()

	// Play Spade-6 on joker-occupied position
	err := s.PlayerPlay(0)
	assert.NoError(t, err)

	// Bit should be cleared but no joker returned (jokerCards was empty)
	assert.True(t, s.jokerPlaced[CardDesignSpade]&(1<<6) == 0,
		"jokerPlaced bit should be cleared")
	assert.Equal(t, handSizeBefore-1, players[0].GetCardsSize(),
		"hand size should decrease by 1 (no joker to return)")

	// No joker in hand
	for i := 0; i < players[0].GetCardsSize(); i++ {
		assert.NotEqual(t, CardDesignJoker, players[0].GetCard(i).GetDesign(),
			"player should not have any joker")
	}
}

func TestReclaimJokerIfNeeded_InvalidBounds_Internal(t *testing.T) {
	// reclaimJokerIfNeeded with invalid suit or value returns immediately (defensive guard).
	// This is unreachable via the public API since valid cards always have valid suit/value.
	tc := NewTrumpCards(2)
	players := makeSevensPlayersInternal()
	cfg := SevensConfig{
		JokerCount:          2,
		JokerReclaimEnabled: true,
		MaxPasses:           SevensMaxPasses,
	}
	s := NewSevens(tc, players, cfg)

	// Pre-set joker tracking so the bounds guard is the only thing stopping reclaim.
	s.jokerPlaced[CardDesignSpade] |= 1 << 6
	s.jokerCards = append(s.jokerCards, NewCard(CardDesignJoker, 1, false))

	initial := s.GetJokerCardsCount()

	// Invalid suit (0 = CardDesignJoker < CardDesignSpade)
	s.reclaimJokerIfNeeded(0, 0, 6)
	assert.Equal(t, initial, s.GetJokerCardsCount(), "invalid suit 0: no change")

	// Invalid suit (5 > CardDesignDiamond=4)
	s.reclaimJokerIfNeeded(0, 5, 6)
	assert.Equal(t, initial, s.GetJokerCardsCount(), "invalid suit 5: no change")

	// Invalid value (0 < 1)
	s.reclaimJokerIfNeeded(0, CardDesignSpade, 0)
	assert.Equal(t, initial, s.GetJokerCardsCount(), "invalid value 0: no change")

	// Invalid value (14 > 13)
	s.reclaimJokerIfNeeded(0, CardDesignSpade, 14)
	assert.Equal(t, initial, s.GetJokerCardsCount(), "invalid value 14: no change")

	// Sanity check: valid call with jokerPlaced bit set does perform reclaim
	s.reclaimJokerIfNeeded(0, CardDesignSpade, 6)
	assert.Equal(t, initial-1, s.GetJokerCardsCount(), "valid call: joker reclaimed")
}

func TestJokerReclaim_RecordJokerCard_InvalidBounds_Internal(t *testing.T) {
	// recordJokerCard with invalid suit/value should still append to jokerCards
	// but NOT set jokerPlaced bit.
	tc := NewTrumpCards(2)
	players := makeSevensPlayersInternal()
	cfg := SevensConfig{
		JokerCount:          2,
		JokerReclaimEnabled: true,
		MaxPasses:           SevensMaxPasses,
	}
	s := NewSevens(tc, players, cfg)

	joker := NewCard(CardDesignJoker, 1, false)

	// Call with out-of-range suit (0 = CardDesignJoker)
	s.recordJokerCard(joker, 0, 6)
	assert.Equal(t, 1, s.GetJokerCardsCount(), "joker card should be tracked")
	// No jokerPlaced bits should be set
	for suit := 1; suit <= 4; suit++ {
		assert.Equal(t, uint16(0), s.jokerPlaced[suit],
			"no jokerPlaced bits should be set for invalid suit")
	}

	// Call with out-of-range value
	joker2 := NewCard(CardDesignJoker, 2, false)
	s.recordJokerCard(joker2, CardDesignSpade, 0)
	assert.Equal(t, 2, s.GetJokerCardsCount())
	assert.Equal(t, uint16(0), s.jokerPlaced[CardDesignSpade],
		"no jokerPlaced bits set for invalid value")

	s.recordJokerCard(NewCard(CardDesignJoker, 3, false), CardDesignSpade, 14)
	assert.Equal(t, 3, s.GetJokerCardsCount())
	assert.Equal(t, uint16(0), s.jokerPlaced[CardDesignSpade],
		"no jokerPlaced bits set for value 14")
}

func TestWrapValue(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{1, 1},
		{13, 13},
		{14, 1},
		{15, 2},
		{0, 13},
		{-1, 12},
		{-2, 11},
		{26, 13},
		{27, 1},
		{-12, 1},
		{-13, 13},
		{-25, 1},
		{-26, 13},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.expected, wrapValue(tc.input), "wrapValue(%d)", tc.input)
	}
}

func TestSevens_countOpponentsHoldingCard(t *testing.T) {
	tc := NewTrumpCards(0)
	players := []*SevensPlayer{
		NewSevensPlayer(true),
		NewSevensPlayer(false),
		NewSevensPlayer(false),
		NewSevensPlayer(false),
	}
	cfg := DefaultSevensConfig()
	s := NewSevens(tc, players, cfg)
	// CPU1 has spade 3
	players[1].AddCard(NewCard(CardDesignSpade, 3, false))
	// CPU2 has spade 3 too
	players[2].AddCard(NewCard(CardDesignSpade, 3, false))
	// CPU3 is finished
	players[3].SetIsFinished(true)

	count := s.countOpponentsHoldingCard(players[0], CardDesignSpade, 3)
	assert.Greater(t, count, 0, "should count opponents holding the card")

	// No one has spade 10
	count2 := s.countOpponentsHoldingCard(players[0], CardDesignSpade, 10)
	assert.Equal(t, 0, count2, "no opponents hold spade 10")
}

func TestSevens_evaluatePlay_TunnelSkipWidth(t *testing.T) {
	t.Run("skip=3: positive score when player has card at skip distance", func(t *testing.T) {
		tc := NewTrumpCards(0)
		players := []*SevensPlayer{
			NewSevensPlayer(true),
			NewSevensPlayer(false),
			NewSevensPlayer(false),
			NewSevensPlayer(false),
		}
		cfg := SevensConfig{TunnelSkipWidth: 3, CpuStrategy: true, MaxPasses: 5}
		s := NewSevens(tc, players, cfg)
		for i := 0; i < 4; i++ {
			for d := 0; d < 10; d++ {
				players[i].AddCard(NewCard(CardDesignDiamond, 2, false))
			}
		}
		// Player has spade 1 (=4-3) AND spade 3 (=4-1) AND spade 10 (=4+3+3)
		// so all directions yield +2
		players[0].AddCard(NewCard(CardDesignSpade, 1, false))
		players[0].AddCard(NewCard(CardDesignSpade, 3, false))
		players[0].AddCard(NewCard(CardDesignSpade, 5, false))
		card := NewCard(CardDesignSpade, 4, false)
		// Without skip, evaluatePlay checks ±1: has 3 (+2), has 5 (+2) = 4
		scoreNoSkip := func() int {
			noSkipCfg := SevensConfig{CpuStrategy: true, MaxPasses: 5}
			s2 := NewSevens(NewTrumpCards(0), []*SevensPlayer{
				NewSevensPlayer(true), NewSevensPlayer(false),
				NewSevensPlayer(false), NewSevensPlayer(false),
			}, noSkipCfg)
			for i := 0; i < 4; i++ {
				for d := 0; d < 10; d++ {
					s2.players[i].AddCard(NewCard(CardDesignDiamond, 2, false))
				}
			}
			s2.players[0].AddCard(NewCard(CardDesignSpade, 1, false))
			s2.players[0].AddCard(NewCard(CardDesignSpade, 3, false))
			s2.players[0].AddCard(NewCard(CardDesignSpade, 5, false))
			return s2.evaluatePlay(s2.players[0], NewCard(CardDesignSpade, 4, false))
		}()
		score := s.evaluatePlay(players[0], card)
		// With skip=3, player has spade 1 at skip distance → extra +2
		assert.Greater(t, score, scoreNoSkip)
	})

	t.Run("skip=3: negative score when opponent has card at skip distance", func(t *testing.T) {
		tc := NewTrumpCards(0)
		players := []*SevensPlayer{
			NewSevensPlayer(true),
			NewSevensPlayer(false),
			NewSevensPlayer(false),
			NewSevensPlayer(false),
		}
		cfg := SevensConfig{TunnelSkipWidth: 3, CpuStrategy: true, MaxPasses: 5}
		s := NewSevens(tc, players, cfg)
		for i := 0; i < 4; i++ {
			for d := 0; d < 10; d++ {
				players[i].AddCard(NewCard(CardDesignDiamond, 2, false))
			}
		}
		// Player does NOT have spade 1 (=4-3), but CPU1 does → penalty
		players[1].AddCard(NewCard(CardDesignSpade, 1, false))
		// Player has spade 3 and 5 for ±1 adjacency
		players[0].AddCard(NewCard(CardDesignSpade, 3, false))
		players[0].AddCard(NewCard(CardDesignSpade, 5, false))
		card := NewCard(CardDesignSpade, 4, false)
		scoreWithSkip := s.evaluatePlay(players[0], card)
		// Without skip, no penalty for position 1
		noSkipCfg := SevensConfig{CpuStrategy: true, MaxPasses: 5}
		s2 := NewSevens(NewTrumpCards(0), []*SevensPlayer{
			NewSevensPlayer(true), NewSevensPlayer(false),
			NewSevensPlayer(false), NewSevensPlayer(false),
		}, noSkipCfg)
		for i := 0; i < 4; i++ {
			for d := 0; d < 10; d++ {
				s2.players[i].AddCard(NewCard(CardDesignDiamond, 2, false))
			}
		}
		s2.players[1].AddCard(NewCard(CardDesignSpade, 1, false))
		s2.players[0].AddCard(NewCard(CardDesignSpade, 3, false))
		s2.players[0].AddCard(NewCard(CardDesignSpade, 5, false))
		scoreNoSkip := s2.evaluatePlay(s2.players[0], NewCard(CardDesignSpade, 4, false))
		// Skip=3 adds penalty because opponent holds card at skip distance
		assert.Less(t, scoreWithSkip, scoreNoSkip)
	})

	t.Run("skip=3 with tunnel: evaluates wrap direction", func(t *testing.T) {
		tc := NewTrumpCards(0)
		players := []*SevensPlayer{
			NewSevensPlayer(true),
			NewSevensPlayer(false),
			NewSevensPlayer(false),
			NewSevensPlayer(false),
		}
		cfg := SevensConfig{TunnelEnabled: true, TunnelSkipWidth: 3, CpuStrategy: true, MaxPasses: 5}
		s := NewSevens(tc, players, cfg)
		for i := 0; i < 4; i++ {
			for d := 0; d < 10; d++ {
				players[i].AddCard(NewCard(CardDesignDiamond, 2, false))
			}
		}
		// Place 1-7 on spade board
		for v := 1; v <= 6; v++ {
			s.placePosition(CardDesignSpade, v)
		}
		// Player has spade 11 (wrapValue(1-3) = 11 when tunnel enabled)
		players[0].AddCard(NewCard(CardDesignSpade, 11, false))
		card := NewCard(CardDesignSpade, 1, false)
		score := s.evaluatePlay(players[0], card)
		// Should include consideration of skip distance (11 = wrap from 1-3)
		assert.GreaterOrEqual(t, score, 0)
	})
}
