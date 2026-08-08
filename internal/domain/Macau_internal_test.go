//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newInternalTestMacau() *Macau {
	players := []*MacauPlayer{
		NewMacauPlayer(true),
		NewMacauPlayer(false),
		NewMacauPlayer(false),
		NewMacauPlayer(false),
	}
	return NewMacau(NewTrumpCards(0), players, DefaultMacauConfig())
}

func newInternalTestMacauWithDifficulty(d MacauCpuDifficulty) *Macau {
	players := []*MacauPlayer{
		NewMacauPlayer(true),
		NewMacauPlayer(false),
		NewMacauPlayer(false),
		NewMacauPlayer(false),
	}
	cfg := DefaultMacauConfig()
	cfg.CpuDifficulty = d
	return NewMacau(NewTrumpCards(0), players, cfg)
}

// --- isValidPlay ---

func TestMacau_isValidPlay(t *testing.T) {
	g := newInternalTestMacau()
	g.Reset()

	t.Run("wild 8 is always valid", func(t *testing.T) {
		g.discardPile = []*Card{NewCard(CardDesignSpade, 5, false)}
		g.chosenSuit = -1
		g.penaltyDrawCount = 0
		assert.True(t, g.isValidPlay(NewCard(CardDesignHeart, 8, false)))
	})

	t.Run("suit match", func(t *testing.T) {
		g.discardPile = []*Card{NewCard(CardDesignSpade, 5, false)}
		g.chosenSuit = -1
		g.penaltyDrawCount = 0
		assert.True(t, g.isValidPlay(NewCard(CardDesignSpade, 3, false)))
	})

	t.Run("rank match", func(t *testing.T) {
		g.discardPile = []*Card{NewCard(CardDesignSpade, 5, false)}
		g.chosenSuit = -1
		g.penaltyDrawCount = 0
		assert.True(t, g.isValidPlay(NewCard(CardDesignHeart, 5, false)))
	})

	t.Run("no match", func(t *testing.T) {
		g.discardPile = []*Card{NewCard(CardDesignSpade, 5, false)}
		g.chosenSuit = -1
		g.penaltyDrawCount = 0
		assert.False(t, g.isValidPlay(NewCard(CardDesignHeart, 7, false)))
	})

	t.Run("empty discard pile", func(t *testing.T) {
		g.discardPile = nil
		g.chosenSuit = -1
		g.penaltyDrawCount = 0
		assert.True(t, g.isValidPlay(NewCard(CardDesignHeart, 7, false)))
	})

	t.Run("chosenSuit match", func(t *testing.T) {
		g.discardPile = []*Card{NewCard(CardDesignSpade, 8, false)}
		g.chosenSuit = CardDesignHeart
		g.penaltyDrawCount = 0
		assert.True(t, g.isValidPlay(NewCard(CardDesignHeart, 7, false)))
	})

	t.Run("chosenSuit no match", func(t *testing.T) {
		g.discardPile = []*Card{NewCard(CardDesignSpade, 8, false)}
		g.chosenSuit = CardDesignHeart
		g.penaltyDrawCount = 0
		assert.False(t, g.isValidPlay(NewCard(CardDesignDiamond, 7, false)))
	})

	t.Run("during penalty only 2 is valid", func(t *testing.T) {
		g.discardPile = []*Card{NewCard(CardDesignSpade, 2, false)}
		g.chosenSuit = -1
		g.penaltyDrawCount = 2
		assert.True(t, g.isValidPlay(NewCard(CardDesignHeart, 2, false)))  // a 2 can stack
		assert.False(t, g.isValidPlay(NewCard(CardDesignSpade, 5, false))) // suit match but not a 2
		assert.False(t, g.isValidPlay(NewCard(CardDesignHeart, 8, false))) // even 8 cannot
	})
}

// --- playCard: magic cards ---

func TestMacau_playCard_MagicCards(t *testing.T) {
	t.Run("playing 2 accumulates draw stack", func(t *testing.T) {
		g := newInternalTestMacau()
		g.Reset()
		g.discardPile = []*Card{NewCard(CardDesignSpade, 5, false)}
		g.penaltyDrawCount = 0
		g.direction = 1
		g.currentPlayerIdx = 0
		g.players[0].Reset()
		g.players[0].AddCard(NewCard(CardDesignClover, 9, false)) // keep non-empty + non-declare
		g.players[0].AddCard(NewCard(CardDesignClover, 10, false))

		g.playCard(0, NewCard(CardDesignSpade, 2, false))
		assert.Equal(t, MacauDrawTwoAmount, g.penaltyDrawCount)
		assert.Equal(t, 1, g.currentPlayerIdx) // advanced normally
	})

	t.Run("stacking 2 on top of existing penalty", func(t *testing.T) {
		g := newInternalTestMacau()
		g.Reset()
		g.discardPile = []*Card{NewCard(CardDesignSpade, 2, false)}
		g.penaltyDrawCount = 2
		g.direction = 1
		g.currentPlayerIdx = 1
		g.players[1].Reset()
		g.players[1].AddCard(NewCard(CardDesignHeart, 9, false))
		g.players[1].AddCard(NewCard(CardDesignHeart, 10, false))

		g.playCard(1, NewCard(CardDesignHeart, 2, false))
		assert.Equal(t, 4, g.penaltyDrawCount)
	})

	t.Run("playing 7 skips next player", func(t *testing.T) {
		g := newInternalTestMacau()
		g.Reset()
		g.discardPile = []*Card{NewCard(CardDesignSpade, 5, false)}
		g.penaltyDrawCount = 0
		g.direction = 1
		g.currentPlayerIdx = 0
		g.players[0].Reset()
		g.players[0].AddCard(NewCard(CardDesignHeart, 9, false))
		g.players[0].AddCard(NewCard(CardDesignHeart, 10, false))

		g.playCard(0, NewCard(CardDesignSpade, 7, false))
		assert.Equal(t, 2, g.currentPlayerIdx) // skipped player 1
	})

	t.Run("playing J reverses direction", func(t *testing.T) {
		g := newInternalTestMacau()
		g.Reset()
		g.discardPile = []*Card{NewCard(CardDesignSpade, 5, false)}
		g.penaltyDrawCount = 0
		g.direction = 1
		g.currentPlayerIdx = 0
		g.players[0].Reset()
		g.players[0].AddCard(NewCard(CardDesignHeart, 9, false))
		g.players[0].AddCard(NewCard(CardDesignHeart, 10, false))

		g.playCard(0, NewCard(CardDesignSpade, 11, false))
		assert.Equal(t, -1, g.direction)
		assert.Equal(t, 3, g.currentPlayerIdx) // 0 - 1 wraps to 3
	})

	t.Run("normal card advances turn and resets chosenSuit", func(t *testing.T) {
		g := newInternalTestMacau()
		g.Reset()
		g.discardPile = []*Card{NewCard(CardDesignSpade, 5, false)}
		g.chosenSuit = CardDesignHeart
		g.penaltyDrawCount = 0
		g.direction = 1
		g.currentPlayerIdx = 0
		g.players[0].Reset()
		g.players[0].AddCard(NewCard(CardDesignClover, 9, false))
		g.players[0].AddCard(NewCard(CardDesignClover, 10, false))

		g.playCard(0, NewCard(CardDesignHeart, 9, false)) // 9 is not a magic card
		assert.Equal(t, -1, g.chosenSuit)
		assert.Equal(t, 1, g.currentPlayerIdx)
	})

	t.Run("playing 8 enters choose suit phase", func(t *testing.T) {
		g := newInternalTestMacau()
		g.Reset()
		g.discardPile = []*Card{NewCard(CardDesignSpade, 5, false)}
		g.penaltyDrawCount = 0
		g.currentPlayerIdx = 1
		g.players[1].Reset()
		g.players[1].AddCard(NewCard(CardDesignClover, 9, false))

		g.playCard(1, NewCard(CardDesignHeart, 8, false))
		assert.Equal(t, MacauPhaseChooseSuit, g.phase)
		assert.Equal(t, 1, g.currentPlayerIdx)
	})

	t.Run("playing last card ends round", func(t *testing.T) {
		g := newInternalTestMacau()
		g.Reset()
		g.discardPile = []*Card{NewCard(CardDesignSpade, 5, false)}
		g.currentPlayerIdx = 2
		g.players[2].Reset()

		g.playCard(2, NewCard(CardDesignSpade, 3, false))
		assert.Equal(t, MacauPhaseRoundEnd, g.phase)
		assert.True(t, g.players[2].GetIsFinished())
	})

	t.Run("reaching one card enters must declare", func(t *testing.T) {
		g := newInternalTestMacau()
		g.Reset()
		g.discardPile = []*Card{NewCard(CardDesignSpade, 5, false)}
		g.currentPlayerIdx = 0
		g.players[0].Reset()
		g.players[0].AddCard(NewCard(CardDesignHeart, 9, false)) // one card left after play
		g.players[0].SetHasDeclared(false)

		g.playCard(0, NewCard(CardDesignSpade, 3, false))
		assert.Equal(t, MacauPhaseMustDeclare, g.phase)
		assert.Equal(t, 0, g.currentPlayerIdx) // stays for declaration
	})
}

// --- advanceTurn with direction & skip ---

func TestMacau_advanceTurn(t *testing.T) {
	g := newInternalTestMacau()
	g.Reset()

	t.Run("forward", func(t *testing.T) {
		g.currentPlayerIdx = 0
		g.direction = 1
		g.pendingSkip = false
		g.advanceTurn()
		assert.Equal(t, 1, g.currentPlayerIdx)
		assert.Equal(t, MacauPhasePlay, g.phase)
	})

	t.Run("forward wraps", func(t *testing.T) {
		g.currentPlayerIdx = 3
		g.direction = 1
		g.pendingSkip = false
		g.advanceTurn()
		assert.Equal(t, 0, g.currentPlayerIdx)
	})

	t.Run("reverse direction", func(t *testing.T) {
		g.currentPlayerIdx = 0
		g.direction = -1
		g.pendingSkip = false
		g.advanceTurn()
		assert.Equal(t, 3, g.currentPlayerIdx) // wraps negative
	})

	t.Run("pending skip jumps two and clears flag", func(t *testing.T) {
		g.currentPlayerIdx = 0
		g.direction = 1
		g.pendingSkip = true
		g.advanceTurn()
		assert.Equal(t, 2, g.currentPlayerIdx)
		assert.False(t, g.pendingSkip)
	})
}

// --- wrapIdx ---

func TestMacau_wrapIdx(t *testing.T) {
	g := newInternalTestMacau()
	assert.Equal(t, 0, g.wrapIdx(0))
	assert.Equal(t, 3, g.wrapIdx(3))
	assert.Equal(t, 0, g.wrapIdx(4))
	assert.Equal(t, 3, g.wrapIdx(-1))
	assert.Equal(t, 2, g.wrapIdx(-2))
	assert.Equal(t, 1, g.wrapIdx(5))
}

// --- drawCard: penalty handling ---

func TestMacau_drawCard_Penalty(t *testing.T) {
	t.Run("taking penalty draws stack and advances", func(t *testing.T) {
		g := newInternalTestMacau()
		g.Reset()
		g.discardPile = []*Card{NewCard(CardDesignSpade, 2, false)}
		g.drawPile = []*Card{
			NewCard(CardDesignHeart, 3, false),
			NewCard(CardDesignHeart, 4, false),
			NewCard(CardDesignHeart, 5, false),
			NewCard(CardDesignHeart, 6, false),
		}
		g.penaltyDrawCount = 4
		g.direction = 1
		g.currentPlayerIdx = 0
		g.players[0].Reset()
		g.players[0].AddCard(NewCard(CardDesignClover, 9, false))

		err := g.drawCard(0)
		assert.NoError(t, err)
		assert.Equal(t, 0, g.penaltyDrawCount)
		assert.Equal(t, 5, g.players[0].GetCardsSize()) // 1 + 4 penalty
		assert.Equal(t, 1, g.currentPlayerIdx)          // turn advanced
	})

	t.Run("normal draw when no penalty", func(t *testing.T) {
		g := newInternalTestMacau()
		g.Reset()
		g.discardPile = []*Card{NewCard(CardDesignSpade, 5, false)}
		g.drawPile = []*Card{NewCard(CardDesignDiamond, 9, false)}
		g.penaltyDrawCount = 0
		g.chosenSuit = -1
		g.currentPlayerIdx = 0
		g.players[0].Reset()
		g.players[0].AddCard(NewCard(CardDesignHeart, 7, false))

		err := g.drawCard(0)
		assert.NoError(t, err)
		assert.Equal(t, 2, g.players[0].GetCardsSize())
		assert.Equal(t, 1, g.currentPlayerIdx) // not playable => advanced
	})
}

// --- drawCards helper ---

func TestMacau_drawCards(t *testing.T) {
	t.Run("draws requested count", func(t *testing.T) {
		g := newInternalTestMacau()
		g.drawPile = []*Card{
			NewCard(CardDesignHeart, 3, false),
			NewCard(CardDesignHeart, 4, false),
		}
		g.players[0].Reset()
		drawn := g.drawCards(0, 2)
		assert.Equal(t, 2, drawn)
		assert.Equal(t, 2, g.players[0].GetCardsSize())
	})

	t.Run("caps at available cards when piles empty", func(t *testing.T) {
		g := newInternalTestMacau()
		g.drawPile = []*Card{NewCard(CardDesignHeart, 3, false)}
		g.discardPile = []*Card{NewCard(CardDesignSpade, 5, false)} // cannot recycle (1 card)
		g.players[0].Reset()
		drawn := g.drawCards(0, 5)
		assert.Equal(t, 1, drawn) // only one card was available
	})
}

// --- finishTurn ---

func TestMacau_finishTurn(t *testing.T) {
	t.Run("enters declare when one card and not declared", func(t *testing.T) {
		g := newInternalTestMacau()
		g.Reset()
		g.currentPlayerIdx = 0
		g.players[0].Reset()
		g.players[0].AddCard(NewCard(CardDesignSpade, 3, false))
		g.players[0].SetHasDeclared(false)

		g.finishTurn(0)
		assert.Equal(t, MacauPhaseMustDeclare, g.phase)
	})

	t.Run("advances when already declared", func(t *testing.T) {
		g := newInternalTestMacau()
		g.Reset()
		g.currentPlayerIdx = 0
		g.direction = 1
		g.players[0].Reset()
		g.players[0].AddCard(NewCard(CardDesignSpade, 3, false))
		g.players[0].SetHasDeclared(true)

		g.finishTurn(0)
		assert.Equal(t, MacauPhasePlay, g.phase)
		assert.Equal(t, 1, g.currentPlayerIdx)
	})
}

// --- recycleDrawPile ---

func TestMacau_recycleDrawPile(t *testing.T) {
	g := newInternalTestMacau()

	t.Run("recycles keeping top", func(t *testing.T) {
		top := NewCard(CardDesignSpade, 5, false)
		g.discardPile = []*Card{NewCard(CardDesignHeart, 3, false), NewCard(CardDesignClover, 7, false), top}
		g.drawPile = nil
		g.recycleDrawPile()
		assert.Len(t, g.discardPile, 1)
		assert.Equal(t, top, g.discardPile[0])
		assert.Len(t, g.drawPile, 2)
	})

	t.Run("no-op with one card", func(t *testing.T) {
		g.discardPile = []*Card{NewCard(CardDesignSpade, 5, false)}
		g.drawPile = nil
		g.recycleDrawPile()
		assert.Nil(t, g.drawPile)
	})
}

// --- cpuSelectPlayCard delegation ---

func TestMacau_cpuSelectPlayCard(t *testing.T) {
	t.Run("no valid returns -1", func(t *testing.T) {
		g := newInternalTestMacau()
		g.Reset()
		g.discardPile = []*Card{NewCard(CardDesignSpade, 5, false)}
		g.chosenSuit = -1
		g.penaltyDrawCount = 0
		g.players[1].Reset()
		g.players[1].AddCard(NewCard(CardDesignHeart, 9, false))
		assert.Equal(t, -1, g.cpuSelectPlayCard(1))
	})

	t.Run("during penalty without a 2 returns -1", func(t *testing.T) {
		g := newInternalTestMacau()
		g.Reset()
		g.discardPile = []*Card{NewCard(CardDesignSpade, 2, false)}
		g.penaltyDrawCount = 2
		g.players[1].Reset()
		g.players[1].AddCard(NewCard(CardDesignSpade, 5, false)) // not a 2
		assert.Equal(t, -1, g.cpuSelectPlayCard(1))
	})

	t.Run("difficulty delegation", func(t *testing.T) {
		for _, d := range []MacauCpuDifficulty{MacauCpuDifficultyEasy, MacauCpuDifficultyNormal, MacauCpuDifficultyHard} {
			g := newInternalTestMacauWithDifficulty(d)
			g.Reset()
			g.discardPile = []*Card{NewCard(CardDesignSpade, 5, false)}
			g.chosenSuit = -1
			g.penaltyDrawCount = 0
			g.players[1].Reset()
			g.players[1].AddCard(NewCard(CardDesignSpade, 3, false))
			g.players[1].AddCard(NewCard(CardDesignSpade, 7, false))
			g.players[1].AddCard(NewCard(CardDesignHeart, 10, false))
			idx := g.cpuSelectPlayCard(1)
			assert.True(t, idx >= 0)
		}
	})
}

// --- cpuSelectSuit ---

func TestMacau_cpuSelectSuit(t *testing.T) {
	t.Run("smart picks most common suit", func(t *testing.T) {
		g := newInternalTestMacauWithDifficulty(MacauCpuDifficultyHard)
		g.players[1].Reset()
		g.players[1].AddCard(NewCard(CardDesignHeart, 3, false))
		g.players[1].AddCard(NewCard(CardDesignHeart, 5, false))
		assert.Equal(t, CardDesignHeart, g.cpuSelectSuit(1))
	})

	t.Run("easy is random in range", func(t *testing.T) {
		g := newInternalTestMacauWithDifficulty(MacauCpuDifficultyEasy)
		suit := g.cpuSelectSuit(1)
		assert.True(t, suit >= CardDesignSpade && suit <= CardDesignDiamond)
	})
}

// --- countSuits ---

func TestMacau_countSuits(t *testing.T) {
	g := newInternalTestMacau()
	g.players[0].Reset()
	g.players[0].AddCard(NewCard(CardDesignSpade, 3, false))
	g.players[0].AddCard(NewCard(CardDesignSpade, 5, false))
	g.players[0].AddCard(NewCard(CardDesignHeart, 7, false))
	g.players[0].AddCard(NewCard(CardDesignHeart, 8, false)) // 8 excluded
	counts := g.countSuits(0)
	assert.Equal(t, 2, counts[CardDesignSpade])
	assert.Equal(t, 1, counts[CardDesignHeart])
}

// --- playerName ---

func TestMacau_playerName(t *testing.T) {
	g := newInternalTestMacau()
	assert.Equal(t, "You", playerName(g.players, 0))
	assert.Equal(t, "CPU 1", playerName(g.players, 1))
	assert.Equal(t, "Player -1", playerName(g.players, -1))
	assert.Equal(t, "Player 4", playerName(g.players, 4))
}

// --- checkGameEnd ---

func TestMacau_checkGameEnd(t *testing.T) {
	t.Run("no winner below limit", func(t *testing.T) {
		g := newInternalTestMacau()
		g.Reset()
		for i := range g.players {
			g.players[i].cumulativeScore = 10
		}
		g.checkGameEnd()
		assert.False(t, g.gameEndFlag)
	})

	t.Run("highest cumulative wins", func(t *testing.T) {
		g := newInternalTestMacau()
		g.Reset()
		g.config.PointLimit = 50
		g.players[0].cumulativeScore = 55
		g.players[1].cumulativeScore = 70
		g.checkGameEnd()
		assert.True(t, g.gameEndFlag)
		assert.Equal(t, MacauPhaseGameEnd, g.phase)
		assert.Equal(t, 1, g.winnerIdx)
	})
}

// --- applyDeclarePenalty ---

func TestMacau_applyDeclarePenalty(t *testing.T) {
	g := newInternalTestMacau()
	g.Reset()
	g.drawPile = []*Card{NewCard(CardDesignHeart, 3, false), NewCard(CardDesignHeart, 4, false)}
	g.direction = 1
	g.currentPlayerIdx = 0
	g.players[0].Reset()
	g.players[0].AddCard(NewCard(CardDesignSpade, 9, false))

	g.applyDeclarePenalty(0)
	assert.Equal(t, 3, g.players[0].GetCardsSize()) // 1 + 2 penalty
	assert.Equal(t, 1, g.currentPlayerIdx)
}

// --- sortHand ---

func TestMacau_sortHand(t *testing.T) {
	g := newInternalTestMacau()
	g.players[0].Reset()
	g.players[0].AddCard(NewCard(CardDesignHeart, 7, false))
	g.players[0].AddCard(NewCard(CardDesignSpade, 3, false))
	g.sortHand(0)
	assert.Equal(t, CardDesignSpade, g.players[0].GetCard(0).GetDesign())
	assert.Equal(t, CardDesignHeart, g.players[0].GetCard(1).GetDesign())
}
