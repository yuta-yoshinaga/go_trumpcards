//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckKlondikeStalemate_NoHintEmptyStockAndWaste(t *testing.T) {
	k := NewKlondike(NewTrumpCards(0))
	k.phase = KlondikePhasePlaying
	// Empty stock and waste, no valid moves on tableau
	k.stock = nil
	k.waste = nil
	var tableau [KlondikeTableauCnt][]*KlondikeTableauCard
	// Cards that can't form any valid moves
	tableau[0] = []*KlondikeTableauCard{
		{Card: NewCard(CardDesignSpade, 5, false), FaceUp: true},
		{Card: NewCard(CardDesignSpade, 3, false), FaceUp: true},
	}
	k.tableau = tableau
	var foundation [KlondikeFoundationCnt][]*Card
	k.foundation = foundation

	k.checkKlondikeStalemate()
	assert.True(t, k.isStalemate)
}

func TestCheckKlondikeStalemate_NoHintStockExhaustedWithCycle(t *testing.T) {
	k := NewKlondike(NewTrumpCards(0))
	k.phase = KlondikePhasePlaying
	k.stock = nil
	k.waste = []*Card{NewCard(CardDesignSpade, 10, false)}
	k.noProgressCycles = 1
	var tableau [KlondikeTableauCnt][]*KlondikeTableauCard
	tableau[0] = []*KlondikeTableauCard{
		{Card: NewCard(CardDesignSpade, 5, false), FaceUp: true},
	}
	k.tableau = tableau
	var foundation [KlondikeFoundationCnt][]*Card
	k.foundation = foundation

	k.checkKlondikeStalemate()
	assert.True(t, k.isStalemate)
}

func TestCheckKlondikeStalemate_HintAvailable(t *testing.T) {
	k := NewKlondike(NewTrumpCards(0))
	k.phase = KlondikePhasePlaying
	k.stock = nil
	k.waste = nil
	var tableau [KlondikeTableauCnt][]*KlondikeTableauCard
	// Ace on tableau -> hint to foundation
	tableau[0] = []*KlondikeTableauCard{
		{Card: NewCard(CardDesignSpade, 1, false), FaceUp: true},
	}
	k.tableau = tableau
	var foundation [KlondikeFoundationCnt][]*Card
	k.foundation = foundation

	k.checkKlondikeStalemate()
	assert.False(t, k.isStalemate)
}

func TestCheckKlondikeStalemate_NotPlaying(t *testing.T) {
	k := NewKlondike(NewTrumpCards(0))
	k.phase = KlondikePhaseGameClear
	k.isStalemate = false

	k.checkKlondikeStalemate()
	assert.False(t, k.isStalemate) // unchanged
}

func TestCheckKlondikeStalemate_StockNotExhausted(t *testing.T) {
	k := NewKlondike(NewTrumpCards(0))
	k.phase = KlondikePhasePlaying
	k.stock = []*Card{NewCard(CardDesignSpade, 10, false)}
	k.waste = nil
	k.noProgressCycles = 0
	var tableau [KlondikeTableauCnt][]*KlondikeTableauCard
	tableau[0] = []*KlondikeTableauCard{
		{Card: NewCard(CardDesignSpade, 5, false), FaceUp: true},
	}
	k.tableau = tableau
	var foundation [KlondikeFoundationCnt][]*Card
	k.foundation = foundation

	k.checkKlondikeStalemate()
	assert.False(t, k.isStalemate) // Stock still has cards, can draw
}

func TestCheckKlondikeStalemate_NoCycleYet(t *testing.T) {
	k := NewKlondike(NewTrumpCards(0))
	k.phase = KlondikePhasePlaying
	k.stock = nil
	k.waste = []*Card{NewCard(CardDesignSpade, 10, false)}
	k.noProgressCycles = 0 // No cycle yet
	var tableau [KlondikeTableauCnt][]*KlondikeTableauCard
	tableau[0] = []*KlondikeTableauCard{
		{Card: NewCard(CardDesignSpade, 5, false), FaceUp: true},
	}
	k.tableau = tableau
	var foundation [KlondikeFoundationCnt][]*Card
	k.foundation = foundation

	k.checkKlondikeStalemate()
	assert.False(t, k.isStalemate) // No progress cycles yet
}

func TestKlondike_ProgressTracking(t *testing.T) {
	k := NewKlondike(NewTrumpCards(0))
	k.Reset()

	assert.False(t, k.isStalemate)
	assert.Equal(t, 0, k.noProgressCycles)
	assert.False(t, k.progressSinceRecycle)
}

func TestKlondike_RecycleIncrementsCycles(t *testing.T) {
	k := NewKlondike(NewTrumpCards(0))
	k.phase = KlondikePhasePlaying
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

func TestKlondike_RecycleDoesNotIncrementOnProgress(t *testing.T) {
	k := NewKlondike(NewTrumpCards(0))
	k.phase = KlondikePhasePlaying
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

func TestKlondike_StalemateRestoredByUndo(t *testing.T) {
	k := NewKlondike(NewTrumpCards(0))
	k.phase = KlondikePhasePlaying
	k.isStalemate = false
	k.noProgressCycles = 0

	// Set up a state where draw leads to stalemate
	k.stock = []*Card{NewCard(CardDesignSpade, 10, false)}
	k.waste = nil
	var tableau [KlondikeTableauCnt][]*KlondikeTableauCard
	// Only non-matching cards on tableau
	tableau[0] = []*KlondikeTableauCard{
		{Card: NewCard(CardDesignSpade, 5, false), FaceUp: true},
	}
	k.tableau = tableau
	var foundation [KlondikeFoundationCnt][]*Card
	k.foundation = foundation

	// Draw the card
	err := k.Draw()
	assert.NoError(t, err)

	// Undo should restore previous stalemate state
	err = k.Undo()
	assert.NoError(t, err)
	assert.False(t, k.isStalemate)
}
