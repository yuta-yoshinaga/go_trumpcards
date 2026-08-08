//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newInternalTestWhist() *Whist {
	players := []*WhistPlayer{
		NewWhistPlayer(true, 0),
		NewWhistPlayer(false, 1),
		NewWhistPlayer(false, 0),
		NewWhistPlayer(false, 1),
	}
	return NewWhist(NewTrumpCards(0), players, DefaultWhistConfig())
}

func TestWhist_trickWinner_LeadSuitWins(t *testing.T) {
	w := newInternalTestWhist()
	w.trumpSuit = CardDesignSpade

	w.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 5, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 10, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 3, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignDiamond, 13, false)}, // off-suit, doesn't count
	}

	assert.Equal(t, 1, w.trickWinner())
}

func TestWhist_trickWinner_TrumpBeatsLead(t *testing.T) {
	w := newInternalTestWhist()
	w.trumpSuit = CardDesignSpade

	w.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 13, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 2, false)}, // lowest trump beats K of hearts
		{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 10, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignHeart, 8, false)},
	}

	assert.Equal(t, 1, w.trickWinner())
}

func TestWhist_trickWinner_HighestTrumpWins(t *testing.T) {
	w := newInternalTestWhist()
	w.trumpSuit = CardDesignSpade

	w.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 13, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 2, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignSpade, 10, false)}, // higher trump
		{PlayerIdx: 3, Card: NewCard(CardDesignHeart, 8, false)},
	}

	assert.Equal(t, 2, w.trickWinner())
}

func TestWhist_trickWinner_EmptyTrick(t *testing.T) {
	w := newInternalTestWhist()
	w.currentTrick = []*TrickCard{}

	assert.Equal(t, 0, w.trickWinner())
}

func TestWhist_validatePlay_LeadAnyCard(t *testing.T) {
	w := newInternalTestWhist()
	w.currentTrick = nil

	card := NewCard(CardDesignSpade, 5, false)
	err := w.validatePlay(0, card)
	assert.NoError(t, err)
}

func TestWhist_validatePlay_MustFollowSuit(t *testing.T) {
	w := newInternalTestWhist()
	w.currentTrick = []*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 5, false)},
	}

	// Player 0 has a heart
	w.players[0].Reset()
	w.players[0].AddCard(NewCard(CardDesignHeart, 10, false))
	w.players[0].AddCard(NewCard(CardDesignSpade, 5, false))

	// Playing spade when holding heart is invalid
	err := w.validatePlay(0, NewCard(CardDesignSpade, 5, false))
	assert.Error(t, err)

	// Playing heart is valid
	err = w.validatePlay(0, NewCard(CardDesignHeart, 10, false))
	assert.NoError(t, err)
}

func TestWhist_validatePlay_CanPlayAnythingWhenVoid(t *testing.T) {
	w := newInternalTestWhist()
	w.currentTrick = []*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 5, false)},
	}

	// Player 0 has no hearts
	w.players[0].Reset()
	w.players[0].AddCard(NewCard(CardDesignSpade, 10, false))
	w.players[0].AddCard(NewCard(CardDesignDiamond, 5, false))

	err := w.validatePlay(0, NewCard(CardDesignSpade, 10, false))
	assert.NoError(t, err)
}

func TestWhist_playerHasSuit(t *testing.T) {
	w := newInternalTestWhist()
	w.players[0].Reset()
	w.players[0].AddCard(NewCard(CardDesignHeart, 5, false))
	w.players[0].AddCard(NewCard(CardDesignSpade, 10, false))

	assert.True(t, w.playerHasSuit(0, CardDesignHeart))
	assert.True(t, w.playerHasSuit(0, CardDesignSpade))
	assert.False(t, w.playerHasSuit(0, CardDesignDiamond))
}

func TestWhist_checkGameEnd_NoWinner(t *testing.T) {
	w := newInternalTestWhist()
	w.config.PointLimit = 5
	w.teamScores = [WhistTeamCnt]int{3, 4}

	w.checkGameEnd()

	assert.False(t, w.gameEndFlag)
}

func TestWhist_checkGameEnd_Team1Wins(t *testing.T) {
	w := newInternalTestWhist()
	w.config.PointLimit = 5
	w.teamScores = [WhistTeamCnt]int{3, 5}

	w.checkGameEnd()

	assert.True(t, w.gameEndFlag)
	assert.Equal(t, 1, w.winnerTeam)
}

func TestWhist_checkGameEnd_TieTeam0Wins(t *testing.T) {
	w := newInternalTestWhist()
	w.config.PointLimit = 5
	w.teamScores = [WhistTeamCnt]int{5, 5}

	w.checkGameEnd()

	assert.True(t, w.gameEndFlag)
	assert.Equal(t, 0, w.winnerTeam) // tie goes to team 0
}

func TestWhist_isPartnerWinning(t *testing.T) {
	w := newInternalTestWhist()
	w.trumpSuit = CardDesignSpade

	// Player 0 (team 0), partner is player 2 (team 0)
	w.currentTrick = []*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 5, false)},  // team 1
		{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 13, false)}, // team 0 - winning
	}

	// Player 0 checks if partner is winning
	assert.True(t, w.isPartnerWinning(0))

	// Player 3 (team 1) checks - partner (player 1) is not winning
	assert.False(t, w.isPartnerWinning(3))
}

func TestWhist_findHumanIdx(t *testing.T) {
	w := newInternalTestWhist()
	assert.Equal(t, 0, findHumanIdx(w.players))
}

func TestWhist_playerName(t *testing.T) {
	w := newInternalTestWhist()
	assert.Equal(t, "You", w.playerName(0))
	assert.Equal(t, "CPU 1", w.playerName(1))
	assert.Contains(t, w.playerName(-1), "Player")
}

// TestWhist_cpuPlayHard_Follow_OverCardLowest pins the follow-to-win branch:
// with no trump in the trick and a beatable lead, the hard CPU plays its lowest
// card that still beats the current best (issue #4300 refactor guard).
func TestWhist_cpuPlayHard_Follow_OverCardLowest(t *testing.T) {
	w := newInternalTestWhist()
	w.trumpSuit = CardDesignSpade
	// Player 0 (team 0) leads Heart 9 and is currently winning; player 1's
	// partner (player 3, team 1) is not winning, so the CPU tries to win.
	w.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 9, false)},
	}
	p1 := w.players[1]
	p1.Reset()
	p1.AddCard(NewCard(CardDesignHeart, 3, false))  // idx 0 — under
	p1.AddCard(NewCard(CardDesignHeart, 11, false)) // idx 1 — the only over-card
	p1.AddCard(NewCard(CardDesignHeart, 5, false))  // idx 2 — under
	// Only Heart 11 beats the 9, so it is chosen.
	assert.Equal(t, 1, w.cpuPlayHard(1, []int{0, 1, 2}))
}

// TestWhist_cpuPlayHard_Follow_UnderCardHighest pins the can't-win branch:
// when no card beats the lead, the hard CPU sheds its highest losing card.
func TestWhist_cpuPlayHard_Follow_UnderCardHighest(t *testing.T) {
	w := newInternalTestWhist()
	w.trumpSuit = CardDesignSpade
	// Player 0 leads Heart 13 — unbeatable by player 1's hand.
	w.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 13, false)},
	}
	p1 := w.players[1]
	p1.Reset()
	p1.AddCard(NewCard(CardDesignHeart, 3, false))  // idx 0
	p1.AddCard(NewCard(CardDesignHeart, 11, false)) // idx 1 — highest under-card
	p1.AddCard(NewCard(CardDesignHeart, 5, false))  // idx 2
	assert.Equal(t, 1, w.cpuPlayHard(1, []int{0, 1, 2}))
}
