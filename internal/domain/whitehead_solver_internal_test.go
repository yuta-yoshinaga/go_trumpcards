//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// deadTableauWH fills ALL seven columns so no empty column is available.
//
// Whitehead lets any card take an empty column, so a stalemate fixture that
// leaves columns empty can never actually be stuck -- Klondike's version relied
// on the King-only rule to keep them shut.
func deadTableauWH() [WhiteheadTableauCnt][]*WhiteheadTableauCard {
	designs := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	var tab [WhiteheadTableauCnt][]*WhiteheadTableauCard
	for i := 0; i < WhiteheadTableauCnt; i++ {
		tab[i] = []*WhiteheadTableauCard{
			{Card: NewCard(designs[i%len(designs)], 5, false), FaceUp: true},
		}
	}
	return tab
}

func TestCheckWhiteheadStalemate_NoHintEmptyStockAndWaste(t *testing.T) {
	k := NewWhitehead(NewTrumpCards(0))
	k.phase = WhiteheadPhasePlaying
	// Empty stock and waste, no valid moves on tableau
	k.stock = nil
	k.waste = nil
	// **Every column must be occupied.** This fixture filled only column 0 and
	// asserted a stalemate, which was true only because GetHint's tableau
	// priority was dead -- moving a card into any of the six empty columns is
	// perfectly legal, so the board was never actually stuck.
	k.tableau = deadTableauWH()
	var foundation [WhiteheadFoundationCnt][]*Card
	k.foundation = foundation

	k.checkWhiteheadStalemate()
	assert.True(t, k.isStalemate)
}

func TestCheckWhiteheadStalemate_NoHintStockExhaustedWithCycle(t *testing.T) {
	k := NewWhitehead(NewTrumpCards(0))
	k.phase = WhiteheadPhasePlaying
	k.stock = nil
	k.waste = []*Card{NewCard(CardDesignSpade, 10, false)}
	k.noProgressCycles = 1
	k.tableau = deadTableauWH()
	var foundation [WhiteheadFoundationCnt][]*Card
	k.foundation = foundation

	k.checkWhiteheadStalemate()
	assert.True(t, k.isStalemate)
}

func TestCheckWhiteheadStalemate_HintAvailable(t *testing.T) {
	k := NewWhitehead(NewTrumpCards(0))
	k.phase = WhiteheadPhasePlaying
	k.stock = nil
	k.waste = nil
	var tableau [WhiteheadTableauCnt][]*WhiteheadTableauCard
	// Ace on tableau -> hint to foundation
	tableau[0] = []*WhiteheadTableauCard{
		{Card: NewCard(CardDesignSpade, 1, false), FaceUp: true},
	}
	k.tableau = tableau
	var foundation [WhiteheadFoundationCnt][]*Card
	k.foundation = foundation

	k.checkWhiteheadStalemate()
	assert.False(t, k.isStalemate)
}

func TestCheckWhiteheadStalemate_NotPlaying(t *testing.T) {
	k := NewWhitehead(NewTrumpCards(0))
	k.phase = WhiteheadPhaseGameClear
	k.isStalemate = false

	k.checkWhiteheadStalemate()
	assert.False(t, k.isStalemate) // unchanged
}

func TestCheckWhiteheadStalemate_StockNotExhausted(t *testing.T) {
	k := NewWhitehead(NewTrumpCards(0))
	k.phase = WhiteheadPhasePlaying
	k.stock = []*Card{NewCard(CardDesignSpade, 10, false)}
	k.waste = nil
	k.noProgressCycles = 0
	k.tableau = deadTableauWH()
	var foundation [WhiteheadFoundationCnt][]*Card
	k.foundation = foundation

	k.checkWhiteheadStalemate()
	assert.False(t, k.isStalemate) // Stock still has cards, can draw
}

func TestCheckWhiteheadStalemate_NoCycleYet(t *testing.T) {
	k := NewWhitehead(NewTrumpCards(0))
	k.phase = WhiteheadPhasePlaying
	k.stock = nil
	k.waste = []*Card{NewCard(CardDesignSpade, 10, false)}
	k.noProgressCycles = 0 // No cycle yet
	k.tableau = deadTableauWH()
	var foundation [WhiteheadFoundationCnt][]*Card
	k.foundation = foundation

	k.checkWhiteheadStalemate()
	assert.False(t, k.isStalemate) // No progress cycles yet
}

func TestWhitehead_ProgressTracking(t *testing.T) {
	k := NewWhitehead(NewTrumpCards(0))
	k.Reset()

	assert.False(t, k.isStalemate)
	assert.Equal(t, 0, k.noProgressCycles)
	assert.False(t, k.progressSinceRecycle)
}

func TestWhitehead_RecycleIncrementsCycles(t *testing.T) {
	k := NewWhitehead(NewTrumpCards(0))
	k.phase = WhiteheadPhasePlaying
	k.stock = nil
	k.waste = []*Card{
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignHeart, 5, false),
	}
	k.progressSinceRecycle = false
	k.noProgressCycles = 0

	// Recycle (Draw when stock empty)
	err := k.Draw()
	assert.NoError(t, err)
	assert.Equal(t, 1, k.noProgressCycles)
	assert.False(t, k.progressSinceRecycle)
}

func TestWhitehead_RecycleDoesNotIncrementOnProgress(t *testing.T) {
	k := NewWhitehead(NewTrumpCards(0))
	k.phase = WhiteheadPhasePlaying
	k.stock = nil
	k.waste = []*Card{
		NewCard(CardDesignSpade, 10, false),
	}
	k.progressSinceRecycle = true // Progress was made
	k.noProgressCycles = 0

	err := k.Draw()
	assert.NoError(t, err)
	assert.Equal(t, 0, k.noProgressCycles)
	assert.False(t, k.progressSinceRecycle) // Reset after recycle
}

func TestWhitehead_StalemateRestoredByUndo(t *testing.T) {
	k := NewWhitehead(NewTrumpCards(0))
	k.phase = WhiteheadPhasePlaying
	k.isStalemate = false
	k.noProgressCycles = 1 // Already cycled once without progress

	// Set up: stock empty, waste has one card, no valid moves on tableau
	k.stock = nil
	k.waste = []*Card{NewCard(CardDesignSpade, 10, false)}
	k.tableau = deadTableauWH()
	var foundation [WhiteheadFoundationCnt][]*Card
	k.foundation = foundation
	k.progressSinceRecycle = false

	// Recycle waste->stock (Draw when stock empty, no progress)
	err := k.Draw()
	assert.NoError(t, err)
	// noProgressCycles incremented to 2, stock now has 1 card
	// checkWhiteheadStalemate: stock not empty after recycle, so not stalemate yet

	// Draw the card from stock to waste
	err = k.Draw()
	assert.NoError(t, err)
	// Now stock is empty, noProgressCycles=2, no hint -> stalemate
	assert.True(t, k.isStalemate)

	// Undo should restore the non-stalemate state
	err = k.Undo()
	assert.NoError(t, err)
	assert.False(t, k.isStalemate)
}
