//go:build test

package domain

import (
	"math/rand"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHearts_GetValidPlayIndices(t *testing.T) {
	t.Run("first trick requires two clubs", func(t *testing.T) {
		h := newInternalTestHearts()
		h.phase, h.currentPlayerIdx, h.trickNumber = HeartsPhasePlay, 0, 1
		h.players[0].AddCard(NewCard(CardDesignHeart, 2, false))
		h.players[0].AddCard(NewCard(CardDesignClover, 2, false))
		h.players[0].AddCard(NewCard(CardDesignSpade, 3, false))
		assert.Equal(t, []int{1}, h.GetValidPlayIndices(0))
	})
	t.Run("unbroken hearts are excluded from lead when a non-heart exists", func(t *testing.T) {
		h := newInternalTestHearts()
		h.phase, h.currentPlayerIdx = HeartsPhasePlay, 0
		h.players[0].AddCard(NewCard(CardDesignHeart, 2, false))
		h.players[0].AddCard(NewCard(CardDesignSpade, 3, false))
		assert.Equal(t, []int{1}, h.GetValidPlayIndices(0))
	})
	t.Run("all hearts may lead when they are the whole hand", func(t *testing.T) {
		h := newInternalTestHearts()
		h.phase, h.currentPlayerIdx = HeartsPhasePlay, 0
		h.players[0].AddCard(NewCard(CardDesignHeart, 2, false))
		h.players[0].AddCard(NewCard(CardDesignHeart, 3, false))
		assert.Equal(t, []int{0, 1}, h.GetValidPlayIndices(0))
	})
	t.Run("must follow suit", func(t *testing.T) {
		h := newInternalTestHearts()
		h.phase, h.currentPlayerIdx = HeartsPhasePlay, 0
		h.players[0].AddCard(NewCard(CardDesignClover, 2, false))
		h.players[0].AddCard(NewCard(CardDesignHeart, 2, false))
		h.currentTrick = []*TrickCard{{PlayerIdx: 1, Card: NewCard(CardDesignClover, 3, false)}}
		assert.Equal(t, []int{0}, h.GetValidPlayIndices(0))
	})
	t.Run("first trick void excludes points except omnibus jack", func(t *testing.T) {
		h := newInternalTestHearts()
		h.phase, h.currentPlayerIdx, h.trickNumber = HeartsPhasePlay, 0, 1
		h.config.OmnibusJD = false
		h.currentTrick = []*TrickCard{{PlayerIdx: 1, Card: NewCard(CardDesignClover, 3, false)}}
		for _, card := range []*Card{NewCard(CardDesignHeart, 2, false), NewCard(CardDesignSpade, 12, false), NewCard(CardDesignDiamond, 11, false), NewCard(CardDesignSpade, 3, false)} {
			h.players[0].AddCard(card)
		}
		assert.Equal(t, []int{2, 3}, h.GetValidPlayIndices(0))
		h.config.OmnibusJD = true
		assert.Equal(t, []int{3}, h.GetValidPlayIndices(0))
	})
	t.Run("wrong turn, phase, and index return non-nil empty", func(t *testing.T) {
		h := newInternalTestHearts()
		h.phase, h.currentPlayerIdx = HeartsPhasePlay, 0
		h.players[0].AddCard(NewCard(CardDesignSpade, 3, false))
		for _, idx := range []int{1, -1, HeartsPlayerCnt} {
			got := h.GetValidPlayIndices(idx)
			assert.NotNil(t, got)
			assert.Empty(t, got)
		}
		h.phase = HeartsPhasePass
		got := h.GetValidPlayIndices(0)
		assert.NotNil(t, got)
		assert.Empty(t, got)
	})
	t.Run("matches validatePlay across randomized deals", func(t *testing.T) {
		for seed := int64(0); seed < 200; seed++ {
			h := newInternalTestHearts()
			h.phase, h.currentPlayerIdx = HeartsPhasePlay, int(seed%HeartsPlayerCnt)
			cards := make([]*Card, 0, 52)
			for suit := CardDesignSpade; suit <= CardDesignMax; suit++ {
				for value := 1; value <= 13; value++ {
					cards = append(cards, NewCard(suit, value, false))
				}
			}
			rand.New(rand.NewSource(seed)).Shuffle(len(cards), func(i, j int) { cards[i], cards[j] = cards[j], cards[i] })
			for i, card := range cards {
				h.players[i/13].AddCard(card)
			}
			h.trickNumber = int(seed%HeartsHandSize) + 1
			if seed%3 != 0 {
				h.currentTrick = []*TrickCard{{PlayerIdx: (h.currentPlayerIdx + 1) % HeartsPlayerCnt, Card: NewCard(CardDesignClover, 5, false)}}
			}
			if seed%2 == 0 {
				h.heartsBroken = true
			}
			valid := h.GetValidPlayIndices(h.currentPlayerIdx)
			player := h.players[h.currentPlayerIdx]
			for i := 0; i < player.GetCardsSize(); i++ {
				assert.Equal(t, h.validatePlay(h.currentPlayerIdx, player.GetCard(i)) == nil, slices.Contains(valid, i), "seed %d, index %d", seed, i)
			}
		}
	})
}

func newInternalTestHearts() *Hearts {
	players := []*HeartsPlayer{
		NewHeartsPlayer(true),
		NewHeartsPlayer(false),
		NewHeartsPlayer(false),
		NewHeartsPlayer(false),
	}
	cfg := DefaultHeartsConfig()
	tc := NewTrumpCards(0)
	return NewHearts(tc, players, cfg)
}

func TestHearts_trickWinner_EmptyTrick(t *testing.T) {
	h := newInternalTestHearts()
	h.currentTrick = nil
	assert.Equal(t, 0, h.trickWinner())
}

func TestHearts_cpuSelectPlayCard_EmptyValidIndices(t *testing.T) {
	h := newInternalTestHearts()
	h.phase = HeartsPhasePlay
	// Player with no cards → getValidPlayIndices returns empty
	h.players[1].Reset()
	h.currentPlayerIdx = 1
	h.currentTrick = nil
	result := h.cpuSelectPlayCard(1)
	assert.Equal(t, 0, result)
}

func TestHearts_passTarget_DefaultCase(t *testing.T) {
	h := newInternalTestHearts()
	// HeartsPassNone returns the same player
	result := h.passTarget(2, HeartsPassNone)
	assert.Equal(t, 2, result)
}

func TestHearts_playerName_OutOfBounds(t *testing.T) {
	h := newInternalTestHearts()
	assert.Equal(t, "Player -1", playerName(h.players, -1))
	assert.Equal(t, "Player 10", playerName(h.players, 10))
}

func TestHearts_startPlayPhase_NotFirstTrick(t *testing.T) {
	h := newInternalTestHearts()
	h.trickNumber = 5
	h.leadPlayerIdx = 2
	h.startPlayPhase()
	assert.Equal(t, 2, h.currentPlayerIdx)
}

// --- Omnibus J♦ internal tests ---

func TestCardPoints_OmnibusJD(t *testing.T) {
	jd := NewCard(CardDesignDiamond, 11, false)
	// Without omnibus: J♦ = 0 points
	assert.Equal(t, 0, cardPoints(jd, false))
	// With omnibus: J♦ = -10 points
	assert.Equal(t, -10, cardPoints(jd, true))

	// Hearts still 1, Q♠ still 13 regardless of omnibus
	h3 := NewCard(CardDesignHeart, 3, false)
	assert.Equal(t, 1, cardPoints(h3, false))
	assert.Equal(t, 1, cardPoints(h3, true))

	qs := NewCard(CardDesignSpade, 12, false)
	assert.Equal(t, 13, cardPoints(qs, false))
	assert.Equal(t, 13, cardPoints(qs, true))

	// Other cards still 0
	c5 := NewCard(CardDesignClover, 5, false)
	assert.Equal(t, 0, cardPoints(c5, false))
	assert.Equal(t, 0, cardPoints(c5, true))
}

func TestIsPointCard_OmnibusJD(t *testing.T) {
	jd := NewCard(CardDesignDiamond, 11, false)
	assert.False(t, isPointCard(jd, false))
	assert.True(t, isPointCard(jd, true))

	// Hearts always point cards
	h3 := NewCard(CardDesignHeart, 3, false)
	assert.True(t, isPointCard(h3, false))
	assert.True(t, isPointCard(h3, true))

	// Q♠ always point card
	qs := NewCard(CardDesignSpade, 12, false)
	assert.True(t, isPointCard(qs, false))
	assert.True(t, isPointCard(qs, true))
}

func TestIsPenaltyCard(t *testing.T) {
	// Hearts are penalty
	assert.True(t, isPenaltyCard(NewCard(CardDesignHeart, 5, false)))
	// Q♠ is penalty
	assert.True(t, isPenaltyCard(NewCard(CardDesignSpade, 12, false)))
	// J♦ is NOT penalty (even though it's a point card under omnibus)
	assert.False(t, isPenaltyCard(NewCard(CardDesignDiamond, 11, false)))
	// Other cards are not penalty
	assert.False(t, isPenaltyCard(NewCard(CardDesignClover, 5, false)))
}

func TestHearts_cpuPlayHard_FollowWithNonLeadSuitCards(t *testing.T) {
	// Exercise the continue branch in cpuPlayHard follow-suit when
	// validIndices contains non-lead-suit cards (shouldn't happen in practice
	// but the code handles it)
	h := newInternalTestHearts()
	h.phase = HeartsPhasePlay
	h.currentPlayerIdx = 1
	h.config.CpuDifficulty = HeartsCpuDifficultyHard

	// Set up trick with clover lead
	h.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignClover, 10, false)},
	}

	// Give CPU lead-suit card + non-lead-suit card
	player := h.players[1]
	player.Reset()
	player.AddCard(NewCard(CardDesignClover, 5, false)) // lead suit
	player.AddCard(NewCard(CardDesignHeart, 8, false))  // non-lead suit (should be skipped)

	result := h.cpuPlayHard(1, []int{0, 1})
	// Should pick clover 5 (lead suit), skipping heart 8
	assert.Equal(t, 0, result)
}

func TestHearts_playerTookCard(t *testing.T) {
	h := newInternalTestHearts()
	jd := NewCard(CardDesignDiamond, 11, false)
	h.players[0].AddTrick([]*Card{jd})

	// Found: J♦ is in player 0's tricksTaken
	assert.True(t, h.playerTookCard(0, CardDesignDiamond, 11))
	// Not found: wrong value
	assert.False(t, h.playerTookCard(0, CardDesignDiamond, 10))
	// Not found: wrong design
	assert.False(t, h.playerTookCard(0, CardDesignSpade, 11))
	// Not found: player 1 has no tricks
	assert.False(t, h.playerTookCard(1, CardDesignDiamond, 11))
}
