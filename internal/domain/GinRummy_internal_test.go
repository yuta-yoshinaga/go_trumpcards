//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newInternalTestGinRummy() *GinRummy {
	players := []*GinRummyPlayer{
		NewGinRummyPlayer(true),
		NewGinRummyPlayer(false),
	}
	return NewGinRummy(NewTrumpCards(0), players, DefaultGinRummyConfig())
}

// --- isSet ---

func TestGinRummy_isSet(t *testing.T) {
	t.Run("true for 3 same rank", func(t *testing.T) {
		meld := []*Card{
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignHeart, 5, false),
			NewCard(CardDesignDiamond, 5, false),
		}
		assert.True(t, isSet(meld))
	})

	t.Run("false for run", func(t *testing.T) {
		meld := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignSpade, 2, false),
			NewCard(CardDesignSpade, 3, false),
		}
		assert.False(t, isSet(meld))
	})

	t.Run("short less than 2", func(t *testing.T) {
		meld := []*Card{
			NewCard(CardDesignSpade, 5, false),
		}
		assert.False(t, isSet(meld))
	})

	t.Run("empty", func(t *testing.T) {
		assert.False(t, isSet(nil))
	})

	t.Run("2 same rank", func(t *testing.T) {
		meld := []*Card{
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignHeart, 5, false),
		}
		assert.True(t, isSet(meld))
	})
}

// --- canAddToMeld ---

func TestGinRummy_canAddToMeld(t *testing.T) {
	t.Run("add to set same rank different suit", func(t *testing.T) {
		meld := []*Card{
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignHeart, 5, false),
			NewCard(CardDesignDiamond, 5, false),
		}
		card := NewCard(CardDesignClover, 5, false)
		assert.True(t, canAddToMeld(meld, card))
	})

	t.Run("set full 4 cards", func(t *testing.T) {
		meld := []*Card{
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignHeart, 5, false),
			NewCard(CardDesignDiamond, 5, false),
			NewCard(CardDesignClover, 5, false),
		}
		card := NewCard(CardDesignSpade, 5, false)
		assert.False(t, canAddToMeld(meld, card))
	})

	t.Run("wrong rank for set", func(t *testing.T) {
		meld := []*Card{
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignHeart, 5, false),
			NewCard(CardDesignDiamond, 5, false),
		}
		card := NewCard(CardDesignClover, 6, false)
		assert.False(t, canAddToMeld(meld, card))
	})

	t.Run("same suit already in set", func(t *testing.T) {
		meld := []*Card{
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignHeart, 5, false),
			NewCard(CardDesignDiamond, 5, false),
		}
		card := NewCard(CardDesignSpade, 5, false)
		assert.False(t, canAddToMeld(meld, card))
	})

	t.Run("add to run front", func(t *testing.T) {
		meld := []*Card{
			NewCard(CardDesignSpade, 3, false),
			NewCard(CardDesignSpade, 4, false),
			NewCard(CardDesignSpade, 5, false),
		}
		card := NewCard(CardDesignSpade, 2, false)
		assert.True(t, canAddToMeld(meld, card))
	})

	t.Run("add to run back", func(t *testing.T) {
		meld := []*Card{
			NewCard(CardDesignSpade, 3, false),
			NewCard(CardDesignSpade, 4, false),
			NewCard(CardDesignSpade, 5, false),
		}
		card := NewCard(CardDesignSpade, 6, false)
		assert.True(t, canAddToMeld(meld, card))
	})

	t.Run("wrong suit for run", func(t *testing.T) {
		meld := []*Card{
			NewCard(CardDesignSpade, 3, false),
			NewCard(CardDesignSpade, 4, false),
			NewCard(CardDesignSpade, 5, false),
		}
		card := NewCard(CardDesignHeart, 6, false)
		assert.False(t, canAddToMeld(meld, card))
	})

	t.Run("not adjacent for run", func(t *testing.T) {
		meld := []*Card{
			NewCard(CardDesignSpade, 3, false),
			NewCard(CardDesignSpade, 4, false),
			NewCard(CardDesignSpade, 5, false),
		}
		card := NewCard(CardDesignSpade, 7, false)
		assert.False(t, canAddToMeld(meld, card))
	})

	t.Run("add to run with unsorted meld", func(t *testing.T) {
		// Meld cards not in ascending order to ensure both minVal and maxVal branches are hit
		meld := []*Card{
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignSpade, 3, false),
			NewCard(CardDesignSpade, 4, false),
		}
		card := NewCard(CardDesignSpade, 2, false)
		assert.True(t, canAddToMeld(meld, card))
		card2 := NewCard(CardDesignSpade, 6, false)
		assert.True(t, canAddToMeld(meld, card2))
	})

	t.Run("empty meld", func(t *testing.T) {
		card := NewCard(CardDesignSpade, 5, false)
		assert.False(t, canAddToMeld(nil, card))
	})
}

// --- excludeCards ---

func TestGinRummy_excludeCards(t *testing.T) {
	c1 := NewCard(CardDesignSpade, 1, false)
	c2 := NewCard(CardDesignHeart, 2, false)
	c3 := NewCard(CardDesignDiamond, 3, false)

	cards := []*Card{c1, c2, c3}
	meld := []*Card{c2}

	rest := excludeCards(cards, meld)
	assert.Len(t, rest, 2)
	assert.Contains(t, rest, c1)
	assert.Contains(t, rest, c3)

	t.Run("exclude none", func(t *testing.T) {
		rest := excludeCards(cards, nil)
		assert.Len(t, rest, 3)
	})

	t.Run("exclude all", func(t *testing.T) {
		rest := excludeCards(cards, cards)
		assert.Empty(t, rest)
	})
}

// --- copyMelds ---

func TestGinRummy_copyMelds(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		assert.Nil(t, copyMelds(nil))
	})

	t.Run("non-nil", func(t *testing.T) {
		melds := [][]*Card{
			{NewCard(CardDesignSpade, 1, false), NewCard(CardDesignSpade, 2, false)},
		}
		copied := copyMelds(melds)
		assert.Len(t, copied, 1)
		assert.Equal(t, melds[0], copied[0])
	})

	t.Run("empty", func(t *testing.T) {
		melds := [][]*Card{}
		copied := copyMelds(melds)
		assert.NotNil(t, copied)
		assert.Len(t, copied, 0)
	})
}

// --- findAllPossibleMelds ---

func TestGinRummy_findAllPossibleMelds(t *testing.T) {
	t.Run("finds sets", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignHeart, 5, false),
			NewCard(CardDesignDiamond, 5, false),
		}
		melds := findAllPossibleMelds(cards)
		assert.NotEmpty(t, melds)
	})

	t.Run("finds set of 4", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignHeart, 5, false),
			NewCard(CardDesignDiamond, 5, false),
			NewCard(CardDesignClover, 5, false),
		}
		melds := findAllPossibleMelds(cards)
		// Should contain both 3-card and 4-card sets
		assert.True(t, len(melds) >= 2)
	})

	t.Run("finds runs", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignSpade, 2, false),
			NewCard(CardDesignSpade, 3, false),
		}
		melds := findAllPossibleMelds(cards)
		assert.NotEmpty(t, melds)
	})

	t.Run("finds long runs", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignSpade, 2, false),
			NewCard(CardDesignSpade, 3, false),
			NewCard(CardDesignSpade, 4, false),
		}
		melds := findAllPossibleMelds(cards)
		// Should find run of 3 and run of 4
		assert.True(t, len(melds) >= 2)
	})

	t.Run("no melds", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 5, false),
		}
		melds := findAllPossibleMelds(cards)
		assert.Empty(t, melds)
	})

	t.Run("empty", func(t *testing.T) {
		melds := findAllPossibleMelds(nil)
		assert.Empty(t, melds)
	})

	t.Run("non-consecutive same suit under 3", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignSpade, 3, false),
		}
		melds := findAllPossibleMelds(cards)
		assert.Empty(t, melds)
	})
}

// --- playerName ---

func TestGinRummy_playerName(t *testing.T) {
	g := newInternalTestGinRummy()

	assert.Equal(t, "You", playerName(g.players, 0))
	assert.Equal(t, "CPU 1", playerName(g.players, 1))

	// Out of range
	assert.Equal(t, "Player -1", playerName(g.players, -1))
	assert.Equal(t, "Player 2", playerName(g.players, 2))
}

// --- appendLog ---

func TestGinRummy_appendLog(t *testing.T) {
	g := newInternalTestGinRummy()
	g.actionLog = nil

	g.appendLog(0, "draw_stock", "test detail", nil)
	assert.Len(t, g.actionLog, 1)
	assert.Equal(t, 1, g.actionLog[0].TurnNumber)
	assert.Equal(t, 0, g.actionLog[0].PlayerIdx)
	assert.Equal(t, "draw_stock", g.actionLog[0].ActionType)
	assert.Equal(t, "test detail", g.actionLog[0].Detail)

	g.appendLog(1, "discard", "discard detail", nil)
	assert.Len(t, g.actionLog, 2)
	assert.Equal(t, 2, g.actionLog[1].TurnNumber)
}

// --- sortHand ---

func TestGinRummy_sortHand(t *testing.T) {
	g := newInternalTestGinRummy()
	g.players[0].Reset()
	g.players[0].AddCard(NewCard(CardDesignHeart, 7, false))
	g.players[0].AddCard(NewCard(CardDesignSpade, 3, false))
	g.players[0].AddCard(NewCard(CardDesignHeart, 2, false))
	g.players[0].AddCard(NewCard(CardDesignSpade, 10, false))

	g.sortHand(0)

	p := g.players[0]
	assert.Equal(t, CardDesignSpade, p.GetCard(0).GetDesign())
	assert.Equal(t, 3, p.GetCard(0).GetValue())
	assert.Equal(t, CardDesignSpade, p.GetCard(1).GetDesign())
	assert.Equal(t, 10, p.GetCard(1).GetValue())
	assert.Equal(t, CardDesignHeart, p.GetCard(2).GetDesign())
	assert.Equal(t, 2, p.GetCard(2).GetValue())
	assert.Equal(t, CardDesignHeart, p.GetCard(3).GetDesign())
	assert.Equal(t, 7, p.GetCard(3).GetValue())
}

// --- advanceTurn ---

func TestGinRummy_advanceTurn(t *testing.T) {
	g := newInternalTestGinRummy()
	g.Reset()

	g.currentPlayerIdx = 0
	g.advanceTurn()
	assert.Equal(t, 1, g.currentPlayerIdx)
	assert.Equal(t, GinRummyPhaseDraw, g.phase)

	g.currentPlayerIdx = 1
	g.advanceTurn()
	assert.Equal(t, 0, g.currentPlayerIdx) // wraps around
}

// --- checkGameEnd ---

func TestGinRummy_checkGameEnd_Internal(t *testing.T) {
	t.Run("no winner when no one reached limit", func(t *testing.T) {
		g := newInternalTestGinRummy()
		g.Reset()
		for i := range g.players {
			g.players[i].cumulativeScore = 10
		}
		g.checkGameEnd()
		assert.False(t, g.gameEndFlag)
	})

	t.Run("game ends when player reaches limit", func(t *testing.T) {
		g := newInternalTestGinRummy()
		g.Reset()
		g.config.PointLimit = 50
		g.players[1].cumulativeScore = 60

		g.checkGameEnd()
		assert.True(t, g.gameEndFlag)
		assert.Equal(t, GinRummyPhaseGameEnd, g.phase)
		assert.Equal(t, 1, g.winnerIdx)
	})

	t.Run("highest score wins when multiple reach limit", func(t *testing.T) {
		g := newInternalTestGinRummy()
		g.Reset()
		g.config.PointLimit = 50
		g.players[0].cumulativeScore = 55
		g.players[1].cumulativeScore = 70

		g.checkGameEnd()
		assert.True(t, g.gameEndFlag)
		assert.Equal(t, 1, g.winnerIdx)
	})
}

// --- endRoundDraw ---

func TestGinRummy_endRoundDraw_Internal(t *testing.T) {
	g := newInternalTestGinRummy()
	g.Reset()
	g.endRoundDraw()
	assert.Equal(t, -1, g.knockerIdx)
	assert.Equal(t, GinRummyPhaseRoundEnd, g.phase)
}

// --- scoreRound ---

func TestGinRummy_scoreRound_Internal(t *testing.T) {
	t.Run("gin scoring", func(t *testing.T) {
		g := newInternalTestGinRummy()
		g.Reset()
		g.knockerIdx = 0
		g.isGin = true
		g.knockerDeadwood = nil // 0 deadwood

		g.players[1].Reset()
		g.players[1].AddCard(NewCard(CardDesignClover, 10, false)) // 10
		g.players[1].AddCard(NewCard(CardDesignClover, 5, false))  // 5

		g.scoreRound()
		// Knocker gets opponent deadwood (15) + GinBonus (25) = 40
		assert.Equal(t, 40, g.players[0].roundScore)
	})

	t.Run("undercut scoring", func(t *testing.T) {
		g := newInternalTestGinRummy()
		g.Reset()
		g.knockerIdx = 0
		g.isGin = false
		g.knockerDeadwood = []*Card{NewCard(CardDesignClover, 8, false)} // 8

		// Opponent has lower deadwood
		g.players[1].Reset()
		g.players[1].AddCard(NewCard(CardDesignClover, 3, false)) // 3

		g.scoreRound()
		// Opponent undercuts: gets knockerDeadwood(8) - opponentDeadwood(3) + UndercutBonus(25) = 30
		assert.Equal(t, 30, g.players[1].roundScore)
	})

	t.Run("normal knock scoring", func(t *testing.T) {
		g := newInternalTestGinRummy()
		g.Reset()
		g.knockerIdx = 0
		g.isGin = false
		g.knockerDeadwood = []*Card{NewCard(CardDesignClover, 3, false)} // 3

		// Opponent has higher deadwood
		g.players[1].Reset()
		g.players[1].AddCard(NewCard(CardDesignClover, 10, false)) // 10

		g.scoreRound()
		// Knocker gets opponentDeadwood(10) - knockerDeadwood(3) = 7
		assert.Equal(t, 7, g.players[0].roundScore)
	})
}

// --- canLayoff ---

func TestGinRummy_canLayoff(t *testing.T) {
	g := newInternalTestGinRummy()
	g.knockerMelds = [][]*Card{
		{
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignHeart, 5, false),
			NewCard(CardDesignDiamond, 5, false),
		},
	}

	t.Run("can lay off to set", func(t *testing.T) {
		card := NewCard(CardDesignClover, 5, false)
		assert.True(t, g.canLayoff(card))
	})

	t.Run("cannot lay off unrelated card", func(t *testing.T) {
		card := NewCard(CardDesignClover, 7, false)
		assert.False(t, g.canLayoff(card))
	})
}

// --- layoffCard ---

func TestGinRummy_layoffCard(t *testing.T) {
	g := newInternalTestGinRummy()
	g.knockerMelds = [][]*Card{
		{
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignHeart, 5, false),
			NewCard(CardDesignDiamond, 5, false),
		},
	}

	card := NewCard(CardDesignClover, 5, false)
	g.layoffCard(card)
	assert.Len(t, g.knockerMelds[0], 4)
}

// --- cpuLayoff ---

func TestGinRummy_cpuLayoff_NoLayoff(t *testing.T) {
	g := newInternalTestGinRummy()
	g.Reset()
	g.phase = GinRummyPhaseLayoff
	g.currentPlayerIdx = 1
	g.knockerIdx = 0
	g.knockerMelds = [][]*Card{
		{
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignHeart, 5, false),
			NewCard(CardDesignDiamond, 5, false),
		},
	}
	g.knockerDeadwood = nil

	g.players[1].Reset()
	g.players[1].AddCard(NewCard(CardDesignClover, 7, false)) // can't lay off

	g.cpuLayoff()
	assert.Equal(t, 1, g.players[1].GetCardsSize())
}

// --- dealInitialCards ---

func TestGinRummy_dealInitialCards(t *testing.T) {
	g := newInternalTestGinRummy()
	g.Reset()

	totalCards := 0
	for i := 0; i < GinRummyPlayerCnt; i++ {
		totalCards += g.players[i].GetCardsSize()
	}
	totalCards += len(g.discardPile) + len(g.drawPile)
	assert.Equal(t, 52, totalCards)
}
