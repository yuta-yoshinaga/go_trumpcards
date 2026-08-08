//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newInternalTestSpades() *Spades {
	players := []*SpadesPlayer{
		NewSpadesPlayer(true),
		NewSpadesPlayer(false),
		NewSpadesPlayer(false),
		NewSpadesPlayer(false),
	}
	return NewSpades(NewTrumpCards(0), players, DefaultSpadesConfig())
}

func TestSpades_findHumanIdx(t *testing.T) {
	s := newInternalTestSpades()
	assert.Equal(t, 0, findHumanIdx(s.players))

	// All CPU
	allCpu := []*SpadesPlayer{
		NewSpadesPlayer(false),
		NewSpadesPlayer(false),
		NewSpadesPlayer(false),
		NewSpadesPlayer(false),
	}
	s2 := NewSpades(NewTrumpCards(0), allCpu, DefaultSpadesConfig())
	assert.Equal(t, -1, findHumanIdx(s2.players))
}

func TestSpades_findTwoOfClubs(t *testing.T) {
	s := newInternalTestSpades()
	s.Reset()
	idx := s.findTwoOfClubs()
	assert.True(t, idx >= 0 && idx < 4)

	// Verify the player actually has 2♣
	p := s.players[idx]
	found := false
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c.GetDesign() == CardDesignClover && c.GetValue() == 2 {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestSpades_findTwoOfClubs_NotFound(t *testing.T) {
	s := newInternalTestSpades()
	// Don't deal cards - no one has 2♣
	for _, p := range s.players {
		p.Reset()
	}
	idx := s.findTwoOfClubs()
	assert.Equal(t, -1, idx)
}

func TestSpades_playerHasCard(t *testing.T) {
	s := newInternalTestSpades()
	s.players[0].Reset()
	s.players[0].AddCard(NewCard(CardDesignSpade, 5, false))
	assert.True(t, s.playerHasCard(0, CardDesignSpade, 5))
	assert.False(t, s.playerHasCard(0, CardDesignSpade, 6))
}

func TestSpades_playerHasSuit(t *testing.T) {
	s := newInternalTestSpades()
	s.players[0].Reset()
	s.players[0].AddCard(NewCard(CardDesignSpade, 5, false))
	assert.True(t, s.playerHasSuit(0, CardDesignSpade))
	assert.False(t, s.playerHasSuit(0, CardDesignHeart))
}

func TestSpades_playerHasNonSpade(t *testing.T) {
	s := newInternalTestSpades()
	s.players[0].Reset()

	// Only spades
	s.players[0].AddCard(NewCard(CardDesignSpade, 5, false))
	assert.False(t, s.playerHasNonSpade(0))

	// Add a heart
	s.players[0].AddCard(NewCard(CardDesignHeart, 3, false))
	assert.True(t, s.playerHasNonSpade(0))
}

func TestSpades_trickWinner_EmptyTrick(t *testing.T) {
	s := newInternalTestSpades()
	s.currentTrick = nil
	assert.Equal(t, 0, s.trickWinner())
}

func TestSpades_trickWinner_LeadSuitWins(t *testing.T) {
	s := newInternalTestSpades()
	s.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 5, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 10, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 3, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignDiamond, 13, false)},
	}
	assert.Equal(t, 1, s.trickWinner())
}

func TestSpades_trickWinner_SpadeBeatsLead(t *testing.T) {
	s := newInternalTestSpades()
	s.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 13, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 2, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 10, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignDiamond, 13, false)},
	}
	assert.Equal(t, 1, s.trickWinner())
}

func TestSpades_trickWinner_HighestSpadeWins(t *testing.T) {
	s := newInternalTestSpades()
	s.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 13, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 5, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignSpade, 10, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 3, false)},
	}
	assert.Equal(t, 2, s.trickWinner())
}

func TestSpades_trickWinner_OffSuitDoesNotWin(t *testing.T) {
	s := newInternalTestSpades()
	s.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignClover, 5, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 13, false)}, // off-suit K
		{PlayerIdx: 2, Card: NewCard(CardDesignClover, 3, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignDiamond, 13, false)}, // off-suit K
	}
	assert.Equal(t, 0, s.trickWinner()) // lead suit 5♣ wins since only it and 3♣ match
}

func TestSpades_validatePlay_FollowSuit(t *testing.T) {
	s := newInternalTestSpades()
	s.trickNumber = 2
	s.spadesBroken = true

	s.players[0].Reset()
	s.players[0].AddCard(NewCard(CardDesignHeart, 5, false))
	s.players[0].AddCard(NewCard(CardDesignClover, 3, false))

	s.currentTrick = []*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignClover, 7, false)},
	}

	// Must follow clover suit
	err := s.validatePlay(0, NewCard(CardDesignHeart, 5, false))
	assert.Error(t, err)

	err = s.validatePlay(0, NewCard(CardDesignClover, 3, false))
	assert.NoError(t, err)
}

func TestSpades_validatePlay_NoSuitAllowed(t *testing.T) {
	s := newInternalTestSpades()
	s.trickNumber = 2
	s.spadesBroken = true

	s.players[0].Reset()
	s.players[0].AddCard(NewCard(CardDesignHeart, 5, false))

	s.currentTrick = []*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignClover, 7, false)},
	}

	// No clover, can play anything
	err := s.validatePlay(0, NewCard(CardDesignHeart, 5, false))
	assert.NoError(t, err)
}

func TestSpades_validatePlay_SpadesNotBroken_Lead(t *testing.T) {
	s := newInternalTestSpades()
	s.trickNumber = 2
	s.spadesBroken = false
	s.currentTrick = nil

	s.players[0].Reset()
	s.players[0].AddCard(NewCard(CardDesignSpade, 5, false))
	s.players[0].AddCard(NewCard(CardDesignHeart, 3, false))

	// Can't lead spade when has non-spade
	err := s.validatePlay(0, NewCard(CardDesignSpade, 5, false))
	assert.Error(t, err)

	err = s.validatePlay(0, NewCard(CardDesignHeart, 3, false))
	assert.NoError(t, err)
}

func TestSpades_validatePlay_SpadesNotBroken_OnlySpades(t *testing.T) {
	s := newInternalTestSpades()
	s.trickNumber = 2
	s.spadesBroken = false
	s.currentTrick = nil

	s.players[0].Reset()
	s.players[0].AddCard(NewCard(CardDesignSpade, 5, false))

	// Only has spades, can lead with spade
	err := s.validatePlay(0, NewCard(CardDesignSpade, 5, false))
	assert.NoError(t, err)
}

func TestSpades_validatePlay_TwoOfClubs(t *testing.T) {
	s := newInternalTestSpades()
	s.trickNumber = 1
	s.currentTrick = nil

	s.players[0].Reset()
	s.players[0].AddCard(NewCard(CardDesignClover, 2, false))
	s.players[0].AddCard(NewCard(CardDesignHeart, 5, false))

	// Must play 2♣
	err := s.validatePlay(0, NewCard(CardDesignHeart, 5, false))
	assert.Error(t, err)

	err = s.validatePlay(0, NewCard(CardDesignClover, 2, false))
	assert.NoError(t, err)
}

func TestSpades_validatePlay_TwoOfClubs_NotInHand(t *testing.T) {
	s := newInternalTestSpades()
	s.trickNumber = 1
	s.currentTrick = nil

	s.players[0].Reset()
	s.players[0].AddCard(NewCard(CardDesignClover, 5, false))
	s.players[0].AddCard(NewCard(CardDesignHeart, 5, false))

	// No 2♣, can play anything
	err := s.validatePlay(0, NewCard(CardDesignClover, 5, false))
	assert.NoError(t, err)
}

func TestSpades_getValidPlayIndices(t *testing.T) {
	s := newInternalTestSpades()
	s.trickNumber = 2
	s.spadesBroken = true
	s.currentTrick = nil

	s.players[0].Reset()
	s.players[0].AddCard(NewCard(CardDesignSpade, 5, false))
	s.players[0].AddCard(NewCard(CardDesignHeart, 3, false))
	s.players[0].AddCard(NewCard(CardDesignClover, 7, false))

	valid := s.getValidPlayIndices(0)
	assert.Equal(t, 3, len(valid))
}

func TestSpades_getValidPlayIndices_FollowSuit(t *testing.T) {
	s := newInternalTestSpades()
	s.trickNumber = 2
	s.spadesBroken = true
	s.currentTrick = []*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 7, false)},
	}

	s.players[0].Reset()
	s.players[0].AddCard(NewCard(CardDesignSpade, 5, false))
	s.players[0].AddCard(NewCard(CardDesignHeart, 3, false))
	s.players[0].AddCard(NewCard(CardDesignClover, 7, false))

	valid := s.getValidPlayIndices(0)
	assert.Equal(t, 1, len(valid))
	assert.Equal(t, 1, valid[0]) // Only heart card
}

func TestSpades_cpuBidNormal(t *testing.T) {
	s := newInternalTestSpades()
	s.players[1].Reset()
	s.players[1].AddCard(NewCard(CardDesignSpade, 10, false))
	s.players[1].AddCard(NewCard(CardDesignSpade, 13, false))
	s.players[1].AddCard(NewCard(CardDesignHeart, 13, false))
	s.players[1].AddCard(NewCard(CardDesignClover, 2, false))

	bid := s.cpuBidNormal(1)
	assert.True(t, bid >= 1)
}

func TestSpades_cpuBidNormal_LowHand(t *testing.T) {
	s := newInternalTestSpades()
	s.players[1].Reset()
	s.players[1].AddCard(NewCard(CardDesignClover, 2, false))
	s.players[1].AddCard(NewCard(CardDesignHeart, 3, false))
	s.players[1].AddCard(NewCard(CardDesignDiamond, 4, false))

	bid := s.cpuBidNormal(1)
	assert.Equal(t, 1, bid) // minimum bid
}

func TestSpades_cpuBidHard(t *testing.T) {
	s := newInternalTestSpades()
	s.players[1].Reset()
	// Strong hand with high spades and aces
	s.players[1].AddCard(NewCard(CardDesignSpade, 1, false))  // A♠
	s.players[1].AddCard(NewCard(CardDesignSpade, 13, false)) // K♠
	s.players[1].AddCard(NewCard(CardDesignSpade, 12, false)) // Q♠
	s.players[1].AddCard(NewCard(CardDesignSpade, 11, false)) // J♠
	s.players[1].AddCard(NewCard(CardDesignSpade, 10, false)) // 10♠
	s.players[1].AddCard(NewCard(CardDesignHeart, 1, false))  // A♥
	s.players[1].AddCard(NewCard(CardDesignHeart, 13, false)) // K♥
	s.players[1].AddCard(NewCard(CardDesignClover, 2, false))

	bid := s.cpuBidHard(1)
	assert.True(t, bid >= 1 && bid <= 13)
}

func TestSpades_cpuBidHard_LowHand(t *testing.T) {
	s := newInternalTestSpades()
	s.players[1].Reset()
	s.players[1].AddCard(NewCard(CardDesignClover, 2, false))
	s.players[1].AddCard(NewCard(CardDesignHeart, 3, false))
	s.players[1].AddCard(NewCard(CardDesignDiamond, 4, false))

	bid := s.cpuBidHard(1)
	assert.Equal(t, 1, bid) // minimum bid
}

func TestSpades_cpuPlayNormal_Lead(t *testing.T) {
	s := newInternalTestSpades()
	s.currentTrick = nil

	s.players[1].Reset()
	s.players[1].AddCard(NewCard(CardDesignHeart, 5, false))
	s.players[1].AddCard(NewCard(CardDesignHeart, 10, false))
	s.players[1].SetBid(3)

	// Need more tricks: should pick highest
	idx := s.cpuPlayNormal(1, []int{0, 1})
	assert.Equal(t, 1, idx) // 10♥ is higher

	// Enough tricks: should pick lowest
	s.players[1].AddTrick([]*Card{NewCard(CardDesignHeart, 2, false)})
	s.players[1].AddTrick([]*Card{NewCard(CardDesignHeart, 3, false)})
	s.players[1].AddTrick([]*Card{NewCard(CardDesignHeart, 4, false)})
	idx = s.cpuPlayNormal(1, []int{0, 1})
	assert.Equal(t, 0, idx) // 5♥ is lower
}

func TestSpades_cpuPlayNormal_Follow_TryToWin(t *testing.T) {
	s := newInternalTestSpades()
	s.currentTrick = []*TrickCard{
		{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 7, false)},
	}

	s.players[1].Reset()
	s.players[1].AddCard(NewCard(CardDesignHeart, 5, false))  // idx 0 - under
	s.players[1].AddCard(NewCard(CardDesignHeart, 10, false)) // idx 1 - over
	s.players[1].SetBid(3)

	// Need tricks: should play minimum winning card
	idx := s.cpuPlayNormal(1, []int{0, 1})
	assert.Equal(t, 1, idx)
}

func TestSpades_cpuPlayNormal_Follow_UnderOnly(t *testing.T) {
	s := newInternalTestSpades()
	s.currentTrick = []*TrickCard{
		{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 13, false)}, // K♥
	}

	s.players[1].Reset()
	s.players[1].AddCard(NewCard(CardDesignHeart, 3, false))
	s.players[1].AddCard(NewCard(CardDesignHeart, 5, false))
	s.players[1].SetBid(3)

	// All under: play highest under
	idx := s.cpuPlayNormal(1, []int{0, 1})
	assert.Equal(t, 1, idx)
}

func TestSpades_cpuPlayNormal_Follow_NoLeadSuit(t *testing.T) {
	s := newInternalTestSpades()
	s.currentTrick = []*TrickCard{
		{PlayerIdx: 2, Card: NewCard(CardDesignClover, 7, false)},
	}

	s.players[1].Reset()
	s.players[1].AddCard(NewCard(CardDesignSpade, 5, false))
	s.players[1].AddCard(NewCard(CardDesignHeart, 10, false))
	s.players[1].SetBid(3)

	// No clover: needs tricks, should trump with spade
	idx := s.cpuPlayNormal(1, []int{0, 1})
	assert.Equal(t, 0, idx)
}

func TestSpades_cpuPlayNormal_Follow_NoLeadSuit_EnoughTricks(t *testing.T) {
	s := newInternalTestSpades()
	s.currentTrick = []*TrickCard{
		{PlayerIdx: 2, Card: NewCard(CardDesignClover, 7, false)},
	}

	s.players[1].Reset()
	s.players[1].AddCard(NewCard(CardDesignSpade, 5, false))
	s.players[1].AddCard(NewCard(CardDesignHeart, 10, false))
	s.players[1].AddCard(NewCard(CardDesignDiamond, 3, false))
	s.players[1].SetBid(2)
	s.players[1].AddTrick([]*Card{NewCard(CardDesignHeart, 2, false)})
	s.players[1].AddTrick([]*Card{NewCard(CardDesignHeart, 3, false)})

	// Enough tricks: play lowest non-spade
	idx := s.cpuPlayNormal(1, []int{0, 1, 2})
	assert.Equal(t, 2, idx) // 3♦
}

func TestSpades_cpuPlayNormal_Follow_MustFollowSuit_LowestLeadSuit(t *testing.T) {
	s := newInternalTestSpades()
	s.currentTrick = []*TrickCard{
		{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 13, false)},
	}

	s.players[1].Reset()
	s.players[1].AddCard(NewCard(CardDesignSpade, 5, false))
	s.players[1].AddCard(NewCard(CardDesignHeart, 10, false))
	s.players[1].SetBid(0) // nil bid, enough tricks already since needs 0

	// Has lead suit but enough tricks: play lowest lead suit (undercard)
	idx := s.cpuPlayNormal(1, []int{0, 1})
	// hasLeadSuit is true, tricks(0) >= bid(0), so underCards
	assert.Equal(t, 1, idx) // heart 10 is under K but is the only heart
}

func TestSpades_cpuPlayHard_Lead(t *testing.T) {
	s := newInternalTestSpades()
	s.currentTrick = nil

	s.players[1].Reset()
	s.players[1].AddCard(NewCard(CardDesignSpade, 13, false))
	s.players[1].AddCard(NewCard(CardDesignHeart, 5, false))
	s.players[1].SetBid(3)

	// Need tricks: spade high cards score higher
	idx := s.cpuPlayHard(1, []int{0, 1})
	assert.Equal(t, 0, idx) // K♠ has 13+100 = 113 score
}

func TestSpades_cpuPlayHard_Lead_EnoughTricks(t *testing.T) {
	s := newInternalTestSpades()
	s.currentTrick = nil

	s.players[1].Reset()
	s.players[1].AddCard(NewCard(CardDesignSpade, 10, false))
	s.players[1].AddCard(NewCard(CardDesignHeart, 3, false))
	s.players[1].SetBid(1)
	s.players[1].AddTrick([]*Card{NewCard(CardDesignHeart, 2, false)})

	// Enough tricks: play lowest non-spade
	idx := s.cpuPlayHard(1, []int{0, 1})
	assert.Equal(t, 1, idx) // 3♥
}

func TestSpades_cpuPlayHard_Follow_TryToWin(t *testing.T) {
	s := newInternalTestSpades()
	s.currentTrick = []*TrickCard{
		{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 7, false)},
	}

	s.players[1].Reset()
	s.players[1].AddCard(NewCard(CardDesignHeart, 5, false))
	s.players[1].AddCard(NewCard(CardDesignHeart, 10, false))
	s.players[1].SetBid(3)

	// Need tricks, no spade in trick: play minimum over card
	idx := s.cpuPlayHard(1, []int{0, 1})
	assert.Equal(t, 1, idx)
}

func TestSpades_cpuPlayHard_Follow_UnderCards(t *testing.T) {
	s := newInternalTestSpades()
	s.currentTrick = []*TrickCard{
		{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 13, false)},
	}

	s.players[1].Reset()
	s.players[1].AddCard(NewCard(CardDesignHeart, 3, false))
	s.players[1].AddCard(NewCard(CardDesignHeart, 5, false))
	s.players[1].SetBid(3)

	// All under: play highest under
	idx := s.cpuPlayHard(1, []int{0, 1})
	assert.Equal(t, 1, idx)
}

func TestSpades_cpuPlayHard_Follow_OverCardsOnly(t *testing.T) {
	s := newInternalTestSpades()
	s.currentTrick = []*TrickCard{
		{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 2, false)},
	}

	s.players[1].Reset()
	s.players[1].AddCard(NewCard(CardDesignHeart, 10, false))
	s.players[1].AddCard(NewCard(CardDesignHeart, 13, false))
	s.players[1].SetBid(0) // nil bid, don't need tricks

	// Enough tricks, over cards only: play lowest over
	idx := s.cpuPlayHard(1, []int{0, 1})
	assert.Equal(t, 0, idx) // 10♥ < K♥
}

func TestSpades_cpuPlayHard_Void_SpadesCut(t *testing.T) {
	s := newInternalTestSpades()
	s.currentTrick = []*TrickCard{
		{PlayerIdx: 2, Card: NewCard(CardDesignClover, 7, false)},
	}

	s.players[1].Reset()
	s.players[1].AddCard(NewCard(CardDesignSpade, 5, false))
	s.players[1].AddCard(NewCard(CardDesignHeart, 10, false))
	s.players[1].SetBid(3)

	// No clover, needs tricks: cut with spade
	idx := s.cpuPlayHard(1, []int{0, 1})
	assert.Equal(t, 0, idx)
}

func TestSpades_cpuPlayHard_Void_SpadesCut_WithExistingSpade(t *testing.T) {
	s := newInternalTestSpades()
	s.currentTrick = []*TrickCard{
		{PlayerIdx: 2, Card: NewCard(CardDesignClover, 7, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 5, false)},
	}

	s.players[1].Reset()
	s.players[1].AddCard(NewCard(CardDesignSpade, 3, false))  // can't beat 5♠
	s.players[1].AddCard(NewCard(CardDesignSpade, 10, false)) // can beat 5♠
	s.players[1].AddCard(NewCard(CardDesignHeart, 2, false))
	s.players[1].SetBid(3)

	// Spade already in trick: must play higher spade to win
	idx := s.cpuPlayHard(1, []int{0, 1, 2})
	assert.Equal(t, 1, idx) // 10♠ beats 5♠
}

func TestSpades_cpuPlayHard_Void_NoSpade(t *testing.T) {
	s := newInternalTestSpades()
	s.currentTrick = []*TrickCard{
		{PlayerIdx: 2, Card: NewCard(CardDesignClover, 7, false)},
	}

	s.players[1].Reset()
	s.players[1].AddCard(NewCard(CardDesignHeart, 10, false))
	s.players[1].AddCard(NewCard(CardDesignDiamond, 3, false))
	s.players[1].SetBid(1)
	s.players[1].AddTrick([]*Card{NewCard(CardDesignHeart, 2, false)})

	// Enough tricks, no spade to cut: discard highest
	idx := s.cpuPlayHard(1, []int{0, 1})
	assert.Equal(t, 0, idx) // 10♥ is higher
}

func TestSpades_cpuPlayHard_Follow_SpadeInTrick(t *testing.T) {
	s := newInternalTestSpades()
	s.currentTrick = []*TrickCard{
		{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 7, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 10, false)},
	}

	s.players[1].Reset()
	s.players[1].AddCard(NewCard(CardDesignHeart, 5, false))
	s.players[1].AddCard(NewCard(CardDesignHeart, 13, false))
	s.players[1].SetBid(3)

	// Has lead suit (heart), but spade in trick means can't win with heart
	// Should play under card
	idx := s.cpuPlayHard(1, []int{0, 1})
	assert.Equal(t, 0, idx) // 5♥ is the under card
}

func TestSpades_cpuPlayHard_Follow_SpadeLeadSuit(t *testing.T) {
	s := newInternalTestSpades()
	s.currentTrick = []*TrickCard{
		{PlayerIdx: 2, Card: NewCard(CardDesignSpade, 7, false)},
	}

	s.players[1].Reset()
	s.players[1].AddCard(NewCard(CardDesignSpade, 5, false))
	s.players[1].AddCard(NewCard(CardDesignSpade, 10, false))
	s.players[1].SetBid(3)

	// Lead suit is spade, need tricks: play minimum winning spade
	idx := s.cpuPlayHard(1, []int{0, 1})
	assert.Equal(t, 1, idx) // 10♠ beats 7♠
}

func TestSpades_playerName(t *testing.T) {
	s := newInternalTestSpades()
	assert.Equal(t, "You", playerName(s.players, 0))
	assert.Equal(t, "CPU 1", playerName(s.players, 1))
	assert.Equal(t, "Player -1", playerName(s.players, -1))
	assert.Equal(t, "Player 5", playerName(s.players, 5))
}

func TestSpades_checkGameEnd_NoEnd(t *testing.T) {
	s := newInternalTestSpades()
	for i := 0; i < 4; i++ {
		s.players[i].cumulativeScore = 100
	}
	s.checkGameEnd()
	assert.False(t, s.gameEndFlag)
}

func TestSpades_checkGameEnd_PointLimit(t *testing.T) {
	s := newInternalTestSpades()
	s.players[0].cumulativeScore = 500
	for i := 1; i < 4; i++ {
		s.players[i].cumulativeScore = 100
	}
	s.checkGameEnd()
	assert.True(t, s.gameEndFlag)
	assert.Equal(t, 0, s.winnerIdx)
}

func TestSpades_checkGameEnd_LoseThreshold(t *testing.T) {
	s := newInternalTestSpades()
	s.players[0].cumulativeScore = -200
	for i := 1; i < 4; i++ {
		s.players[i].cumulativeScore = 100
	}
	s.checkGameEnd()
	assert.True(t, s.gameEndFlag)
	// Winner is highest score
	assert.NotEqual(t, 0, s.winnerIdx)
}

func TestSpades_spadeSortHand(t *testing.T) {
	p := NewSpadesPlayer(false)
	p.AddCard(NewCard(CardDesignHeart, 10, false))
	p.AddCard(NewCard(CardDesignSpade, 5, false))
	p.AddCard(NewCard(CardDesignClover, 2, false))
	p.AddCard(NewCard(CardDesignSpade, 1, false))

	spadeSortHand(p)

	// Sorted by design then value
	assert.Equal(t, CardDesignSpade, p.GetCard(0).GetDesign())
	assert.Equal(t, 1, p.GetCard(0).GetValue())
	assert.Equal(t, CardDesignSpade, p.GetCard(1).GetDesign())
	assert.Equal(t, 5, p.GetCard(1).GetValue())
	assert.Equal(t, CardDesignClover, p.GetCard(2).GetDesign())
	assert.Equal(t, CardDesignHeart, p.GetCard(3).GetDesign())
}
