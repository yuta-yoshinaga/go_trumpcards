//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestColorado() *Colorado {
	c := NewDefaultColorado()
	c.Reset()
	return c
}

// clearColoradoBoard wipes the dealt layout so a test can state exactly the
// position it cares about. Never assert on a freshly Reset board -- the deal is
// shuffled, so any such assertion is a hidden flake.
func clearColoradoBoard(c *Colorado) {
	c.stock = nil
	c.waste = nil
	c.isStalemate = false
	c.history = nil
	c.moveCount = 0
	c.phase = ColoradoPhasePlaying
	for i := range ColoradoFoundationCnt {
		c.foundation[i] = nil
	}
	for i := range ColoradoTableauCnt {
		c.tableau[i] = nil
	}
}

// fillColoradoPiles puts one card in every tableau pile so no gap exists. A gap
// changes which hint the game offers, so a board full of holes would test a
// position the game never reaches by accident.
func fillColoradoPiles(c *Colorado, card func() *Card) {
	for i := range ColoradoTableauCnt {
		c.tableau[i] = []*Card{card()}
	}
}

func TestNewColorado(t *testing.T) {
	assert.NotNil(t, NewColorado(NewTrumpCardsWithDecks(2, 0)))
	assert.NotNil(t, NewDefaultColorado())
}

// The deal is 20 piles of ONE card with the other 84 as stock.
func TestColorado_Reset(t *testing.T) {
	c := newTestColorado()

	for i, pile := range c.GetTableau() {
		assert.Len(t, pile, 1, "pile %d", i)
	}
	assert.Equal(t, ColoradoTotalCards-ColoradoTableauCnt, c.GetStockCount())
	assert.Equal(t, 84, c.GetStockCount())
	assert.Empty(t, c.GetWaste())

	// Foundations start EMPTY. #5277 asks for 16 of them, which cannot be right:
	// 16*13 = 208 cards for a 104-card game. Eight (four up, four down) is the
	// only count that consumes the deck exactly.
	for i, pile := range c.GetFoundation() {
		assert.Empty(t, pile, "foundation %d", i)
	}
	assert.Equal(t, ColoradoTotalCards, ColoradoFoundationCnt*ColoradoFoundationTarget)

	assert.Equal(t, ColoradoPhasePlaying, c.GetPhase())
	assert.Equal(t, 0, c.GetMoveCount())
	assert.True(t, c.AllFaceUp())
	assert.False(t, c.GetGameEndFlag())
	assert.False(t, c.IsStalemate())
}

func TestColorado_ResetDealsEveryCard(t *testing.T) {
	for range 20 {
		c := newTestColorado()
		total := c.GetStockCount() + len(c.GetWaste())
		for _, pile := range c.GetTableau() {
			total += len(pile)
		}
		for _, pile := range c.GetFoundation() {
			total += len(pile)
		}
		assert.Equal(t, ColoradoTotalCards, total)
	}
}

func TestColorado_ResetTwiceIsClean(t *testing.T) {
	c := newTestColorado()
	require.NoError(t, c.Draw())
	c.Reset()
	assert.Empty(t, c.GetWaste())
	assert.Equal(t, 0, c.GetMoveCount())
	assert.False(t, c.CanUndo())
}

// --- Foundation direction ---

func TestColorado_IsAscendingFoundation(t *testing.T) {
	c := newTestColorado()
	for i := range ColoradoAscendingCnt {
		assert.True(t, c.IsAscendingFoundation(i), "foundation %d should build up", i)
	}
	for i := ColoradoAscendingCnt; i < ColoradoFoundationCnt; i++ {
		assert.False(t, c.IsAscendingFoundation(i), "foundation %d should build down", i)
	}
	assert.False(t, c.IsAscendingFoundation(-1))
	assert.False(t, c.IsAscendingFoundation(ColoradoFoundationCnt))
}

// Each suit has one ascending and one descending foundation, and the two share
// the same suit order so the board layout is stable across deals.
func TestColorado_FoundationSuitPairing(t *testing.T) {
	for i := range ColoradoAscendingCnt {
		assert.Equal(t, coloradoSuitOrder[i], coloradoSuitOrder[i+ColoradoAscendingCnt],
			"foundation %d and %d must cover the same suit", i, i+ColoradoAscendingCnt)
	}
}

func TestColorado_CanPlaceOnFoundation_Ascending(t *testing.T) {
	c := newTestColorado()
	clearColoradoBoard(c)
	// Foundation 0 is spades, ascending.
	assert.True(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 1, true), 0))
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 2, true), 0))
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, CardValueMax, true), 0))
	// Wrong suit never fits.
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignHeart, 1, true), 0))

	c.foundation[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
	assert.True(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 2, true), 0))
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 1, true), 0))
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 3, true), 0))
}

func TestColorado_CanPlaceOnFoundation_Descending(t *testing.T) {
	c := newTestColorado()
	clearColoradoBoard(c)
	// Foundation 4 is spades, descending: it starts at K.
	assert.True(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, CardValueMax, true), 4))
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 1, true), 4))
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, CardValueMax-1, true), 4))

	c.foundation[4] = []*Card{NewCard(CardDesignSpade, CardValueMax, true)}
	assert.True(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, CardValueMax-1, true), 4))
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, CardValueMax, true), 4))
}

func TestColorado_CanPlaceOnFoundation_RejectsFullPileAndBadIndex(t *testing.T) {
	c := newTestColorado()
	clearColoradoBoard(c)
	for v := 1; v <= CardValueMax; v++ {
		c.foundation[0] = append(c.foundation[0], NewCard(CardDesignSpade, v, true))
	}
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 1, true), 0))
	assert.False(t, c.canPlaceOnFoundation(nil, 0))
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 1, true), -1))
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 1, true), ColoradoFoundationCnt))
}

// A card that fits both its ascending and its descending foundation may go to
// either -- two copies exist per suit+rank, one for each direction.
func TestColorado_FindFoundation_AcceptsEitherDirection(t *testing.T) {
	c := newTestColorado()
	clearColoradoBoard(c)
	// Spade ascending sits at Q, so it wants K. Spade descending is empty, so it
	// wants K too. Both are legal homes for a King of spades.
	for v := 1; v < CardValueMax; v++ {
		c.foundation[0] = append(c.foundation[0], NewCard(CardDesignSpade, v, true))
	}
	king := NewCard(CardDesignSpade, CardValueMax, true)
	assert.True(t, c.canPlaceOnFoundation(king, 0))
	assert.True(t, c.canPlaceOnFoundation(king, 4))
	assert.Equal(t, 0, c.findFoundation(king), "first fit wins")

	// After the first King lands, the second copy still has a home.
	c.foundation[0] = append(c.foundation[0], king)
	assert.Equal(t, 4, c.findFoundation(NewCard(CardDesignSpade, CardValueMax, true)))
}

func TestColorado_FindFoundation_NoHome(t *testing.T) {
	c := newTestColorado()
	clearColoradoBoard(c)
	assert.Equal(t, -1, c.findFoundation(NewCard(CardDesignSpade, 7, true)))
	assert.Equal(t, -1, c.findFoundation(nil))
}

// --- Draw ---

func TestColorado_Draw(t *testing.T) {
	c := newTestColorado()
	before := c.GetStockCount()
	require.NoError(t, c.Draw())
	assert.Equal(t, before-1, c.GetStockCount())
	assert.Len(t, c.GetWaste(), 1)
	assert.Equal(t, 1, c.GetMoveCount())
	assert.True(t, c.CanUndo())
}

// There is no redeal: once the stock runs out it stays out.
func TestColorado_Draw_NoRedeal(t *testing.T) {
	c := newTestColorado()
	clearColoradoBoard(c)
	fillColoradoPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })
	c.waste = []*Card{NewCard(CardDesignHeart, 5, true)}
	err := c.Draw()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no redeal")
	assert.Len(t, c.GetWaste(), 1, "waste must not be recycled into the stock")
}

func TestColorado_Draw_RejectedWhenNotPlaying(t *testing.T) {
	c := newTestColorado()
	c.GiveUp()
	assert.Error(t, c.Draw())
}

// --- Tableau -> foundation ---

func TestColorado_MoveTableauToFoundation(t *testing.T) {
	c := newTestColorado()
	clearColoradoBoard(c)
	fillColoradoPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })
	c.tableau[3] = []*Card{NewCard(CardDesignSpade, 1, true)}

	require.NoError(t, c.MoveTableauToFoundation(3))
	assert.Len(t, c.GetFoundation()[0], 1)
	assert.Empty(t, c.GetTableau()[3])
	assert.Equal(t, 1, c.GetMoveCount())
}

func TestColorado_MoveTableauToFoundation_Descending(t *testing.T) {
	c := newTestColorado()
	clearColoradoBoard(c)
	fillColoradoPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })
	c.tableau[5] = []*Card{NewCard(CardDesignHeart, CardValueMax, true)}

	require.NoError(t, c.MoveTableauToFoundation(5))
	// Hearts descending is foundation index 6 (2 + ColoradoAscendingCnt).
	assert.Len(t, c.GetFoundation()[6], 1)
	assert.Equal(t, CardValueMax, c.GetFoundation()[6][0].GetValue())
}

func TestColorado_MoveTableauToFoundation_Rejections(t *testing.T) {
	c := newTestColorado()
	clearColoradoBoard(c)
	fillColoradoPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })

	assert.Error(t, c.MoveTableauToFoundation(-1))
	assert.Error(t, c.MoveTableauToFoundation(ColoradoTableauCnt))

	c.tableau[2] = nil
	err := c.MoveTableauToFoundation(2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")

	// A 7 with nothing under it has no home.
	err = c.MoveTableauToFoundation(0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "foundation")

	c.GiveUp()
	assert.Error(t, c.MoveTableauToFoundation(0))
}

// --- Waste -> foundation ---

func TestColorado_MoveWasteToFoundation(t *testing.T) {
	c := newTestColorado()
	clearColoradoBoard(c)
	fillColoradoPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })
	c.waste = []*Card{NewCard(CardDesignClover, 1, true)}

	require.NoError(t, c.MoveWasteToFoundation())
	assert.Len(t, c.GetFoundation()[1], 1)
	assert.Empty(t, c.GetWaste())
}

func TestColorado_MoveWasteToFoundation_Rejections(t *testing.T) {
	c := newTestColorado()
	clearColoradoBoard(c)
	fillColoradoPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })

	err := c.MoveWasteToFoundation()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "waste is empty")

	c.waste = []*Card{NewCard(CardDesignClover, 7, true)}
	err = c.MoveWasteToFoundation()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "foundation")

	c.GiveUp()
	assert.Error(t, c.MoveWasteToFoundation())
}

// --- Waste -> tableau: the core rule ---

// The waste card goes on ANY pile, whatever the suit or rank. #5277 restricts
// this to empty spaces, which would delete the only decision the game has.
func TestColorado_MoveWasteToTableau_AcceptsAnyPile(t *testing.T) {
	c := newTestColorado()
	clearColoradoBoard(c)
	fillColoradoPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })
	c.waste = []*Card{NewCard(CardDesignHeart, 2, true)}

	// Pile 4 already holds a spade 7 -- a different suit and a distant rank.
	require.NoError(t, c.MoveWasteToTableau(4))
	assert.Len(t, c.GetTableau()[4], 2)
	assert.Equal(t, CardDesignHeart, c.GetTableau()[4][1].GetDesign())
	assert.Empty(t, c.GetWaste())
}

func TestColorado_MoveWasteToTableau_FillsEmptyPile(t *testing.T) {
	c := newTestColorado()
	clearColoradoBoard(c)
	fillColoradoPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })
	c.tableau[9] = nil
	c.waste = []*Card{NewCard(CardDesignHeart, 2, true)}

	require.NoError(t, c.MoveWasteToTableau(9))
	assert.Len(t, c.GetTableau()[9], 1)
}

func TestColorado_MoveWasteToTableau_Rejections(t *testing.T) {
	c := newTestColorado()
	clearColoradoBoard(c)
	fillColoradoPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })

	assert.Error(t, c.MoveWasteToTableau(-1))
	assert.Error(t, c.MoveWasteToTableau(ColoradoTableauCnt))

	err := c.MoveWasteToTableau(0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "waste is empty")

	c.waste = []*Card{NewCard(CardDesignHeart, 2, true)}
	c.GiveUp()
	assert.Error(t, c.MoveWasteToTableau(0))
}

// Burying a card really does put it out of reach -- only the top of a pile plays.
func TestColorado_BuriedCardCannotReachFoundation(t *testing.T) {
	c := newTestColorado()
	clearColoradoBoard(c)
	fillColoradoPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })
	c.tableau[1] = []*Card{NewCard(CardDesignClover, 1, true)}
	c.waste = []*Card{NewCard(CardDesignHeart, 9, true)}

	require.NoError(t, c.MoveWasteToTableau(1))
	err := c.MoveTableauToFoundation(1)
	require.Error(t, err, "the ace is under the 9 now")
	assert.Contains(t, err.Error(), "foundation")
}

// --- Stock -> tableau ---

func TestColorado_MoveStockToTableau_FillsGap(t *testing.T) {
	c := newTestColorado()
	clearColoradoBoard(c)
	fillColoradoPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })
	c.tableau[11] = nil
	c.stock = []*Card{NewCard(CardDesignDiamond, 4, true), NewCard(CardDesignDiamond, 5, true)}

	require.NoError(t, c.MoveStockToTableau(11))
	assert.Len(t, c.GetTableau()[11], 1)
	assert.Equal(t, 4, c.GetTableau()[11][0].GetValue(), "the stock is taken from the top")
	assert.Equal(t, 1, c.GetStockCount())
	assert.Empty(t, c.GetWaste(), "the card must not pass through the waste")
}

func TestColorado_MoveStockToTableau_Rejections(t *testing.T) {
	c := newTestColorado()
	clearColoradoBoard(c)
	fillColoradoPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })

	assert.Error(t, c.MoveStockToTableau(-1))
	assert.Error(t, c.MoveStockToTableau(ColoradoTableauCnt))

	err := c.MoveStockToTableau(0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "stock is empty")

	c.stock = []*Card{NewCard(CardDesignDiamond, 4, true)}
	err = c.MoveStockToTableau(0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty pile")

	c.tableau[0] = nil
	c.GiveUp()
	assert.Error(t, c.MoveStockToTableau(0))
}

// --- buryCost / bestBuryPile ---

func TestColorado_BuryCost(t *testing.T) {
	c := newTestColorado()
	clearColoradoBoard(c)

	// Spade ascending empty: an Ace is wanted right now.
	assert.Equal(t, 0, c.buryCost(NewCard(CardDesignSpade, 1, true)))
	// Spade descending empty wants a K right now.
	assert.Equal(t, 0, c.buryCost(NewCard(CardDesignSpade, CardValueMax, true)))
	// A spade 7 is 6 cards away going up and 6 going down.
	assert.Equal(t, 6, c.buryCost(NewCard(CardDesignSpade, 7, true)))

	// Fill spades ascending to 6 -- now the 7 is wanted immediately.
	for v := 1; v <= 6; v++ {
		c.foundation[0] = append(c.foundation[0], NewCard(CardDesignSpade, v, true))
	}
	assert.Equal(t, 0, c.buryCost(NewCard(CardDesignSpade, 7, true)))
	// The ascending pile has passed the 3, but the descending one still wants it
	// ten cards from now -- a card is only dead when BOTH directions passed it.
	assert.Equal(t, 10, c.buryCost(NewCard(CardDesignSpade, 3, true)))

	// Run the descending pile down past the 3 as well; now nothing can take it.
	for v := CardValueMax; v >= 2; v-- {
		c.foundation[4] = append(c.foundation[4], NewCard(CardDesignSpade, v, true))
	}
	assert.Equal(t, ColoradoFoundationTarget+1, c.buryCost(NewCard(CardDesignSpade, 3, true)))
	assert.Equal(t, ColoradoFoundationTarget+1, c.buryCost(nil))
}

func TestColorado_BestBuryPile_PrefersEmptyPile(t *testing.T) {
	c := newTestColorado()
	clearColoradoBoard(c)
	fillColoradoPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })
	c.tableau[13] = nil
	assert.Equal(t, 13, c.bestBuryPile())
}

func TestColorado_BestBuryPile_PicksTheLeastNeededCard(t *testing.T) {
	c := newTestColorado()
	clearColoradoBoard(c)
	// Spades ascending is at 3 (wants the 4) and spades descending is at 2 (wants
	// the Ace). Every pile holds an Ace, which the descending pile wants right
	// now, except pile 8 -- its 2 has been passed by both directions.
	fillColoradoPiles(c, func() *Card { return NewCard(CardDesignSpade, 1, true) })
	for v := 1; v <= 3; v++ {
		c.foundation[0] = append(c.foundation[0], NewCard(CardDesignSpade, v, true))
	}
	for v := CardValueMax; v >= 2; v-- {
		c.foundation[4] = append(c.foundation[4], NewCard(CardDesignSpade, v, true))
	}
	c.tableau[8] = []*Card{NewCard(CardDesignSpade, 2, true)}

	assert.Equal(t, 0, c.buryCost(NewCard(CardDesignSpade, 1, true)), "the aces are wanted now")
	assert.Equal(t, ColoradoFoundationTarget+1, c.buryCost(NewCard(CardDesignSpade, 2, true)))
	assert.Equal(t, 8, c.bestBuryPile(), "bury the card no foundation can ever take")
}

func TestColorado_BestBuryPile_TieGoesToTheLowestIndex(t *testing.T) {
	c := newTestColorado()
	clearColoradoBoard(c)
	fillColoradoPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })
	assert.Equal(t, 0, c.bestBuryPile())
}

// --- Hints ---

func TestColorado_GetHint_PrefersFoundation(t *testing.T) {
	c := newTestColorado()
	clearColoradoBoard(c)
	fillColoradoPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })
	c.tableau[6] = []*Card{NewCard(CardDesignClover, 1, true)}
	c.stock = []*Card{NewCard(CardDesignDiamond, 4, true)}

	h := c.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "tableau", h.FromZone)
	assert.Equal(t, 6, h.FromIdx)
	assert.Equal(t, "foundation", h.ToZone)
	assert.Equal(t, 1, h.ToIdx)
}

func TestColorado_GetHint_WasteToFoundation(t *testing.T) {
	c := newTestColorado()
	clearColoradoBoard(c)
	fillColoradoPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })
	c.waste = []*Card{NewCard(CardDesignHeart, 1, true)}

	h := c.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "waste", h.FromZone)
	assert.Equal(t, -1, h.FromIdx)
	assert.Equal(t, "foundation", h.ToZone)
	assert.Equal(t, 2, h.ToIdx)
}

func TestColorado_GetHint_FillsGapFromStock(t *testing.T) {
	c := newTestColorado()
	clearColoradoBoard(c)
	fillColoradoPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })
	c.tableau[2] = nil
	c.stock = []*Card{NewCard(CardDesignDiamond, 4, true)}

	h := c.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "stock", h.FromZone)
	assert.Equal(t, "tableau", h.ToZone)
	assert.Equal(t, 2, h.ToIdx)
}

func TestColorado_GetHint_DrawsWhenNothingElse(t *testing.T) {
	c := newTestColorado()
	clearColoradoBoard(c)
	fillColoradoPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })
	c.stock = []*Card{NewCard(CardDesignDiamond, 4, true)}

	h := c.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "stock", h.FromZone)
	assert.Equal(t, "waste", h.ToZone)
	assert.Equal(t, -1, h.ToIdx)
}

// With the stock gone, the only move left is to bury the waste card.
func TestColorado_GetHint_BuriesWasteAsLastResort(t *testing.T) {
	c := newTestColorado()
	clearColoradoBoard(c)
	fillColoradoPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })
	c.waste = []*Card{NewCard(CardDesignHeart, 9, true)}

	h := c.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "waste", h.FromZone)
	assert.Equal(t, "tableau", h.ToZone)
	assert.Equal(t, c.bestBuryPile(), h.ToIdx)
}

func TestColorado_GetHint_NilWhenNotPlaying(t *testing.T) {
	c := newTestColorado()
	c.GiveUp()
	assert.Nil(t, c.GetHint())
	assert.Nil(t, c.foundationHint())
}

// --- Stalemate ---

func TestColorado_Stalemate_OnlyWhenStockAndWasteAreGone(t *testing.T) {
	c := newTestColorado()
	clearColoradoBoard(c)
	fillColoradoPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })

	// Waste still holds a card, so a bury move exists -- not stalemate.
	c.waste = []*Card{NewCard(CardDesignHeart, 9, true)}
	c.checkStalemate()
	assert.False(t, c.IsStalemate())

	// Empty the waste and nothing is left.
	c.waste = nil
	c.checkStalemate()
	assert.True(t, c.IsStalemate())
}

func TestColorado_Stalemate_NotSetOutsidePlaying(t *testing.T) {
	c := newTestColorado()
	clearColoradoBoard(c)
	c.phase = ColoradoPhaseGameOver
	c.checkStalemate()
	assert.False(t, c.IsStalemate())
}

func TestColorado_UndoToEscape(t *testing.T) {
	c := newTestColorado()
	clearColoradoBoard(c)
	fillColoradoPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })
	assert.Equal(t, 0, c.UndoToEscape(), "not stuck yet")

	c.waste = []*Card{NewCard(CardDesignHeart, 9, true)}
	require.NoError(t, c.MoveWasteToTableau(0))
	assert.True(t, c.IsStalemate())
	assert.Equal(t, 1, c.UndoToEscape())
}

// --- Game clear ---

func TestColorado_GameClear(t *testing.T) {
	c := newTestColorado()
	clearColoradoBoard(c)
	// Fill every foundation but leave one card off the last one.
	for i := range ColoradoFoundationCnt {
		for n := range ColoradoFoundationTarget {
			v := n + 1
			if !coloradoIsAscending(i) {
				v = CardValueMax - n
			}
			c.foundation[i] = append(c.foundation[i], NewCard(coloradoSuitOrder[i], v, true))
		}
	}
	c.foundation[7] = c.foundation[7][:ColoradoFoundationTarget-1]
	c.checkGameClear()
	assert.Equal(t, ColoradoPhasePlaying, c.GetPhase(), "one card still missing")

	// Diamonds descending needs its Ace last.
	c.tableau[0] = []*Card{NewCard(CardDesignDiamond, 1, true)}
	require.NoError(t, c.MoveTableauToFoundation(0))
	assert.Equal(t, ColoradoPhaseGameClear, c.GetPhase())
	assert.True(t, c.GetGameEndFlag())
}

// --- AutoComplete ---

func TestColorado_AutoComplete(t *testing.T) {
	c := newTestColorado()
	clearColoradoBoard(c)
	fillColoradoPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })
	c.tableau[0] = []*Card{NewCard(CardDesignClover, 2, true), NewCard(CardDesignClover, 1, true)}
	c.waste = []*Card{NewCard(CardDesignHeart, 1, true)}

	require.NoError(t, c.AutoComplete())
	assert.Len(t, c.GetFoundation()[1], 2, "the ace then the deuce under it")
	assert.Len(t, c.GetFoundation()[2], 1, "the waste ace too")
	assert.Empty(t, c.GetWaste())
}

func TestColorado_AutoComplete_Rejections(t *testing.T) {
	c := newTestColorado()
	clearColoradoBoard(c)
	fillColoradoPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })
	err := c.AutoComplete()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no card")

	c.GiveUp()
	assert.Error(t, c.AutoComplete())
}

// --- Undo ---

func TestColorado_Undo(t *testing.T) {
	c := newTestColorado()
	clearColoradoBoard(c)
	fillColoradoPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })
	c.waste = []*Card{NewCard(CardDesignHeart, 9, true)}

	assert.False(t, c.CanUndo())
	require.Error(t, c.Undo())

	require.NoError(t, c.MoveWasteToTableau(3))
	assert.True(t, c.CanUndo())
	require.NoError(t, c.Undo())
	assert.Len(t, c.GetWaste(), 1)
	assert.Len(t, c.GetTableau()[3], 1)
	assert.Equal(t, 0, c.GetMoveCount())
}

func TestColorado_UndoN(t *testing.T) {
	c := newTestColorado()
	clearColoradoBoard(c)
	fillColoradoPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })
	c.stock = []*Card{
		NewCard(CardDesignHeart, 9, true),
		NewCard(CardDesignHeart, 8, true),
		NewCard(CardDesignHeart, 6, true),
	}
	require.NoError(t, c.Draw())
	require.NoError(t, c.Draw())
	require.NoError(t, c.Draw())
	assert.Equal(t, 3, c.GetMoveCount())

	require.NoError(t, c.UndoN(2))
	assert.Equal(t, 1, c.GetMoveCount())
	assert.Len(t, c.GetWaste(), 1)
	assert.Equal(t, 2, c.GetStockCount())

	assert.Error(t, c.UndoN(5))
	assert.Error(t, c.UndoN(0))
}

func TestColorado_GiveUp(t *testing.T) {
	c := newTestColorado()
	c.GiveUp()
	assert.Equal(t, ColoradoPhaseGameOver, c.GetPhase())
	assert.True(t, c.GetGameEndFlag())
	assert.NotEmpty(t, c.GetActionLog())

	// A second give-up must not append another entry.
	before := len(c.GetActionLog())
	c.GiveUp()
	assert.Len(t, c.GetActionLog(), before)
}

func TestColorado_ActionLog(t *testing.T) {
	c := newTestColorado()
	require.NoError(t, c.Draw())
	log := c.GetActionLog()
	require.NotEmpty(t, log)
	assert.Equal(t, "draw", log[len(log)-1].ActionType)
}

// --- JSON round-trip ---

func TestColorado_JSONRoundTrip(t *testing.T) {
	c := newTestColorado()
	require.NoError(t, c.Draw())
	require.NoError(t, c.Draw())

	data, err := json.Marshal(c)
	require.NoError(t, err)

	restored := NewDefaultColorado()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, c.GetStockCount(), restored.GetStockCount())
	assert.Equal(t, len(c.GetWaste()), len(restored.GetWaste()))
	assert.Equal(t, c.GetMoveCount(), restored.GetMoveCount())
	assert.Equal(t, c.GetPhase(), restored.GetPhase())
	assert.Equal(t, c.IsStalemate(), restored.IsStalemate())
	for i := range ColoradoTableauCnt {
		assert.Len(t, restored.GetTableau()[i], len(c.GetTableau()[i]), "pile %d", i)
	}
}

// The undo stack has to survive the trip: the Worker rebuilds the game from KV
// on every request, so an unpersisted history means Undo silently never works.
func TestColorado_JSONRoundTripKeepsUndoHistory(t *testing.T) {
	c := newTestColorado()
	require.NoError(t, c.Draw())
	require.NoError(t, c.Draw())
	wasteBefore := len(c.GetWaste())

	data, err := json.Marshal(c)
	require.NoError(t, err)
	restored := NewDefaultColorado()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.True(t, restored.CanUndo())
	require.NoError(t, restored.Undo())
	assert.Len(t, restored.GetWaste(), wasteBefore-1)
	assert.NotEmpty(t, restored.GetTableau()[0], "the snapshot must carry the board, not a blank")
}

func TestColorado_UnmarshalJSON_Rejections(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"broken json", `{`},
		{"phase too low", `{"ps":-1}`},
		{"phase too high", `{"ps":99}`},
		{"negative move count", `{"mc":-1}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Error(t, json.Unmarshal([]byte(tt.data), NewDefaultColorado()))
		})
	}
}

func TestColorado_UnmarshalJSON_RejectsOversizedArrays(t *testing.T) {
	bigCards := make([]*Card, ColoradoTotalCards+1)
	for i := range bigCards {
		bigCards[i] = NewCard(CardDesignSpade, 1, true)
	}
	overCap := make([]*Card, coloradoMaxSliceLen+1)
	for i := range overCap {
		overCap[i] = NewCard(CardDesignSpade, 1, true)
	}

	t.Run("stock", func(t *testing.T) {
		data, err := json.Marshal(&coloradoJSON{Stock: bigCards})
		require.NoError(t, err)
		assert.Error(t, json.Unmarshal(data, NewDefaultColorado()))
	})
	t.Run("waste", func(t *testing.T) {
		data, err := json.Marshal(&coloradoJSON{Waste: bigCards})
		require.NoError(t, err)
		assert.Error(t, json.Unmarshal(data, NewDefaultColorado()))
	})
	t.Run("tableau pile", func(t *testing.T) {
		j := &coloradoJSON{}
		j.Tableau[0] = bigCards
		data, err := json.Marshal(j)
		require.NoError(t, err)
		assert.Error(t, json.Unmarshal(data, NewDefaultColorado()))
	})
	t.Run("foundation pile", func(t *testing.T) {
		j := &coloradoJSON{}
		j.Foundation[0] = make([]*Card, ColoradoFoundationTarget+1)
		for i := range j.Foundation[0] {
			j.Foundation[0][i] = NewCard(CardDesignSpade, 1, true)
		}
		data, err := json.Marshal(j)
		require.NoError(t, err)
		assert.Error(t, json.Unmarshal(data, NewDefaultColorado()))
	})
	t.Run("action log", func(t *testing.T) {
		j := &coloradoJSON{ActionLog: make([]*ActionLogEntry, coloradoMaxSliceLen+1)}
		data, err := json.Marshal(j)
		require.NoError(t, err)
		assert.Error(t, json.Unmarshal(data, NewDefaultColorado()))
	})
	t.Run("history", func(t *testing.T) {
		j := &coloradoJSON{History: make([]*coloradoSnapshot, coloradoMaxSliceLen+1)}
		for i := range j.History {
			j.History[i] = &coloradoSnapshot{}
		}
		data, err := json.Marshal(j)
		require.NoError(t, err)
		assert.Error(t, json.Unmarshal(data, NewDefaultColorado()))
	})
	t.Run("snapshot stock", func(t *testing.T) {
		var s coloradoSnapshot
		data, err := json.Marshal(coloradoSnapshotJSON{Stock: overCap})
		require.NoError(t, err)
		assert.Error(t, s.UnmarshalJSON(data))
	})
	t.Run("snapshot waste", func(t *testing.T) {
		var s coloradoSnapshot
		data, err := json.Marshal(coloradoSnapshotJSON{Waste: overCap})
		require.NoError(t, err)
		assert.Error(t, s.UnmarshalJSON(data))
	})
	t.Run("snapshot tableau", func(t *testing.T) {
		var s coloradoSnapshot
		j := coloradoSnapshotJSON{}
		j.Tableau[0] = overCap
		data, err := json.Marshal(j)
		require.NoError(t, err)
		assert.Error(t, s.UnmarshalJSON(data))
	})
	t.Run("snapshot foundation", func(t *testing.T) {
		var s coloradoSnapshot
		j := coloradoSnapshotJSON{}
		j.Foundation[0] = overCap
		data, err := json.Marshal(j)
		require.NoError(t, err)
		assert.Error(t, s.UnmarshalJSON(data))
	})
	t.Run("snapshot broken json", func(t *testing.T) {
		var s coloradoSnapshot
		assert.Error(t, s.UnmarshalJSON([]byte(`{`)))
	})
}
