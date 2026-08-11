//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestCrazyQuilt() *CrazyQuilt {
	c := NewDefaultCrazyQuilt()
	c.Reset()
	return c
}

// clearCrazyQuiltBoard wipes the dealt layout so a test can state exactly the
// position it cares about. Never assert on a freshly Reset board -- the deal is
// shuffled, so any such assertion is a hidden flake.
func clearCrazyQuiltBoard(c *CrazyQuilt) {
	c.stock = nil
	c.waste = nil
	c.isStalemate = false
	c.history = nil
	c.moveCount = 0
	c.phase = CrazyQuiltPhasePlaying
	c.redealsLeft = CrazyQuiltRedealCnt
	for i := range CrazyQuiltFoundationCnt {
		c.foundation[i] = nil
	}
	for i := range CrazyQuiltCells {
		c.quilt[i] = nil
	}
}

// fillCrazyQuilt puts a dead card in every cell so the quilt is intact and only
// the border has an exposed short side.
func fillCrazyQuilt(c *CrazyQuilt) {
	for i := range CrazyQuiltCells {
		c.quilt[i] = NewCard(CardDesignSpade, 7, true)
	}
}

func cellIdx(row, col int) int { return row*CrazyQuiltGridSize + col }

func TestNewCrazyQuilt(t *testing.T) {
	assert.NotNil(t, NewCrazyQuilt(NewTrumpCardsWithDecks(2, 0)))
	assert.NotNil(t, NewDefaultCrazyQuilt())
}

// The eight foundation seeds leave the deck before anything else, so the stock
// is 104 - 8 - 64 = 32. #5274 says 40, having counted only the quilt.
func TestCrazyQuilt_Reset(t *testing.T) {
	c := newTestCrazyQuilt()

	filled := 0
	for _, card := range c.GetQuilt() {
		if card != nil {
			filled++
		}
	}
	assert.Equal(t, CrazyQuiltCells, filled, "the quilt is full")
	assert.Equal(t, 64, CrazyQuiltCells)
	assert.Equal(t, 32, c.GetStockCount())
	assert.Empty(t, c.GetWaste())
	assert.Equal(t, CrazyQuiltRedealCnt, c.GetRedealsLeft())

	// Each foundation is SEEDED, not empty: an Ace or a King per suit.
	for i, pile := range c.GetFoundation() {
		require.Len(t, pile, 1, "foundation %d", i)
		if c.IsAscendingFoundation(i) {
			assert.Equal(t, 1, pile[0].GetValue(), "foundation %d starts on an Ace", i)
		} else {
			assert.Equal(t, CardValueMax, pile[0].GetValue(), "foundation %d starts on a King", i)
		}
		assert.Equal(t, crazyQuiltSuitOrder[i], pile[0].GetDesign())
	}

	assert.Equal(t, CrazyQuiltPhasePlaying, c.GetPhase())
	assert.Equal(t, 0, c.GetMoveCount())
	assert.True(t, c.AllFaceUp())
	assert.False(t, c.GetGameEndFlag())
}

func TestCrazyQuilt_ResetAccountsForEveryCard(t *testing.T) {
	for range 20 {
		c := newTestCrazyQuilt()
		total := c.GetStockCount() + len(c.GetWaste())
		for _, card := range c.GetQuilt() {
			if card != nil {
				total++
			}
		}
		for _, pile := range c.GetFoundation() {
			total += len(pile)
		}
		assert.Equal(t, CrazyQuiltTotalCards, total)
	}
}

// The seeded Aces and Kings must not also appear in the quilt or stock: one
// copy of each is consumed by the foundation.
func TestCrazyQuilt_ResetRemovesTheSeedCopies(t *testing.T) {
	for range 20 {
		c := newTestCrazyQuilt()
		counts := map[[2]int]int{}
		for _, card := range c.GetQuilt() {
			if card != nil {
				counts[[2]int{card.GetDesign(), card.GetValue()}]++
			}
		}
		for _, card := range c.stock {
			counts[[2]int{card.GetDesign(), card.GetValue()}]++
		}
		for _, suit := range crazyQuiltSuitOrder[:CrazyQuiltAscendingCnt] {
			assert.LessOrEqual(t, counts[[2]int{suit, 1}], 1, "only one Ace of a suit remains in play")
			assert.LessOrEqual(t, counts[[2]int{suit, CardValueMax}], 1, "only one King of a suit remains in play")
		}
	}
}

func TestCrazyQuilt_ResetTwiceIsClean(t *testing.T) {
	c := newTestCrazyQuilt()
	require.NoError(t, c.Draw())
	c.Reset()
	assert.Empty(t, c.GetWaste())
	assert.Equal(t, 0, c.GetMoveCount())
	assert.False(t, c.CanUndo())
	assert.Equal(t, 32, c.GetStockCount())
}

// --- Availability: the rule #5274 gets wrong ---

func TestCrazyQuilt_IsVertical_Alternates(t *testing.T) {
	assert.True(t, CrazyQuiltIsVertical(cellIdx(0, 0)))
	assert.False(t, CrazyQuiltIsVertical(cellIdx(0, 1)))
	assert.False(t, CrazyQuiltIsVertical(cellIdx(1, 0)))
	assert.True(t, CrazyQuiltIsVertical(cellIdx(1, 1)))
	assert.False(t, CrazyQuiltIsVertical(-1))
	assert.False(t, CrazyQuiltIsVertical(CrazyQuiltCells))
}

// A card is available when a SHORT side is exposed: up/down for a vertical
// card, left/right for a horizontal one. #5274 says "any orthogonal neighbour
// empty", which would free a vertical card whose left is open -- it does not.
func TestCrazyQuilt_IsAvailable_OnlyShortSidesCount(t *testing.T) {
	c := newTestCrazyQuilt()
	clearCrazyQuiltBoard(c)
	fillCrazyQuilt(c)

	// (2,2) is vertical: its short sides face (1,2) and (3,2).
	target := cellIdx(2, 2)
	require.True(t, CrazyQuiltIsVertical(target))
	assert.False(t, c.IsAvailable(target), "boxed in on all four sides")

	// Opening a LONG side changes nothing.
	c.quilt[cellIdx(2, 1)] = nil
	assert.False(t, c.IsAvailable(target), "a vertical card does not care about its left")
	c.quilt[cellIdx(2, 3)] = nil
	assert.False(t, c.IsAvailable(target), "nor about its right")

	// Opening a SHORT side frees it.
	c.quilt[cellIdx(1, 2)] = nil
	assert.True(t, c.IsAvailable(target), "the top is a short side")
}

func TestCrazyQuilt_IsAvailable_HorizontalCard(t *testing.T) {
	c := newTestCrazyQuilt()
	clearCrazyQuiltBoard(c)
	fillCrazyQuilt(c)

	// (2,3) is horizontal: its short sides face (2,2) and (2,4).
	target := cellIdx(2, 3)
	require.False(t, CrazyQuiltIsVertical(target))
	assert.False(t, c.IsAvailable(target))

	c.quilt[cellIdx(1, 3)] = nil
	assert.False(t, c.IsAvailable(target), "a horizontal card does not care about its top")
	c.quilt[cellIdx(2, 4)] = nil
	assert.True(t, c.IsAvailable(target), "the right is a short side")
}

// The board edge counts as exposed, which is what makes an intact quilt
// playable at all.
func TestCrazyQuilt_IsAvailable_BorderIsExposed(t *testing.T) {
	c := newTestCrazyQuilt()
	clearCrazyQuiltBoard(c)
	fillCrazyQuilt(c)

	assert.True(t, c.IsAvailable(cellIdx(0, 0)), "vertical on the top edge")
	assert.True(t, c.IsAvailable(cellIdx(7, 7)), "vertical on the bottom edge")
	assert.True(t, c.IsAvailable(cellIdx(0, 7)), "horizontal on the right edge")
	assert.True(t, c.IsAvailable(cellIdx(3, 0)), "horizontal on the left edge")

	// A vertical card on the LEFT edge is not freed by that edge.
	assert.False(t, c.IsAvailable(cellIdx(2, 0)) && CrazyQuiltIsVertical(cellIdx(2, 0)) == false,
		"sanity: (2,0) orientation")
	assert.True(t, CrazyQuiltIsVertical(cellIdx(2, 0)))
	assert.False(t, c.IsAvailable(cellIdx(2, 0)), "a vertical card on the left edge stays boxed in")
}

func TestCrazyQuilt_IsAvailable_CountOnAnIntactQuilt(t *testing.T) {
	c := newTestCrazyQuilt()
	clearCrazyQuiltBoard(c)
	fillCrazyQuilt(c)

	n := 0
	for i := range CrazyQuiltCells {
		if c.IsAvailable(i) {
			n++
		}
	}
	// Four vertical cards on each of the top and bottom rows, four horizontal
	// on each of the left and right columns.
	assert.Equal(t, 16, n)
}

func TestCrazyQuilt_IsAvailable_Rejections(t *testing.T) {
	c := newTestCrazyQuilt()
	clearCrazyQuiltBoard(c)
	assert.False(t, c.IsAvailable(-1))
	assert.False(t, c.IsAvailable(CrazyQuiltCells))
	assert.False(t, c.IsAvailable(0), "an emptied cell holds nothing to take")
}

// --- Foundations ---

func TestCrazyQuilt_CanPlaceOnFoundation(t *testing.T) {
	c := newTestCrazyQuilt()
	clearCrazyQuiltBoard(c)
	c.foundation[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
	c.foundation[4] = []*Card{NewCard(CardDesignSpade, CardValueMax, true)}

	assert.True(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 2, true), 0))
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 3, true), 0))
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignHeart, 2, true), 0), "wrong suit")

	assert.True(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, CardValueMax-1, true), 4))
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 1, true), 4))

	assert.False(t, c.canPlaceOnFoundation(nil, 0))
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 2, true), -1))
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 2, true), CrazyQuiltFoundationCnt))
}

// A pile that has not been seeded takes nothing: the seed is what opens it.
func TestCrazyQuilt_CanPlaceOnFoundation_RejectsUnseededAndFull(t *testing.T) {
	c := newTestCrazyQuilt()
	clearCrazyQuiltBoard(c)
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 2, true), 0), "not seeded yet")

	for v := 1; v <= CardValueMax; v++ {
		c.foundation[0] = append(c.foundation[0], NewCard(CardDesignSpade, v, true))
	}
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 1, true), 0), "already complete")
}

func TestCrazyQuilt_FindFoundation(t *testing.T) {
	c := newTestCrazyQuilt()
	clearCrazyQuiltBoard(c)
	c.foundation[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
	c.foundation[4] = []*Card{NewCard(CardDesignSpade, CardValueMax, true)}

	assert.Equal(t, 0, c.findFoundation(NewCard(CardDesignSpade, 2, true)))
	assert.Equal(t, 4, c.findFoundation(NewCard(CardDesignSpade, CardValueMax-1, true)))
	assert.Equal(t, -1, c.findFoundation(NewCard(CardDesignSpade, 7, true)))
	assert.Equal(t, -1, c.findFoundation(nil))
}

func TestCrazyQuilt_MoveQuiltToFoundation(t *testing.T) {
	c := newTestCrazyQuilt()
	clearCrazyQuiltBoard(c)
	fillCrazyQuilt(c)
	c.foundation[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
	c.quilt[cellIdx(0, 0)] = NewCard(CardDesignSpade, 2, true)

	require.NoError(t, c.MoveQuiltToFoundation(cellIdx(0, 0)))
	assert.Nil(t, c.GetQuilt()[cellIdx(0, 0)])
	assert.Len(t, c.GetFoundation()[0], 2)
	assert.Equal(t, 1, c.GetMoveCount())
}

func TestCrazyQuilt_MoveQuiltToFoundation_Rejections(t *testing.T) {
	c := newTestCrazyQuilt()
	clearCrazyQuiltBoard(c)
	fillCrazyQuilt(c)
	c.foundation[0] = []*Card{NewCard(CardDesignSpade, 1, true)}

	assert.Error(t, c.MoveQuiltToFoundation(-1))
	assert.Error(t, c.MoveQuiltToFoundation(CrazyQuiltCells))

	// Boxed in: the rank fits but the card cannot be reached.
	c.quilt[cellIdx(2, 2)] = NewCard(CardDesignSpade, 2, true)
	err := c.MoveQuiltToFoundation(cellIdx(2, 2))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "short side")

	// Reachable but the rank does not fit.
	err = c.MoveQuiltToFoundation(cellIdx(0, 0))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "foundation")

	c.quilt[cellIdx(0, 0)] = nil
	err = c.MoveQuiltToFoundation(cellIdx(0, 0))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already empty")

	c.GiveUp()
	assert.Error(t, c.MoveQuiltToFoundation(cellIdx(0, 2)))
}

func TestCrazyQuilt_MoveWasteToFoundation(t *testing.T) {
	c := newTestCrazyQuilt()
	clearCrazyQuiltBoard(c)
	c.foundation[1] = []*Card{NewCard(CardDesignClover, 1, true)}
	c.waste = []*Card{NewCard(CardDesignClover, 2, true)}

	require.NoError(t, c.MoveWasteToFoundation())
	assert.Len(t, c.GetFoundation()[1], 2)
	assert.Empty(t, c.GetWaste())
}

func TestCrazyQuilt_MoveWasteToFoundation_Rejections(t *testing.T) {
	c := newTestCrazyQuilt()
	clearCrazyQuiltBoard(c)

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

// --- Quilt to waste: the move #5274 omits entirely ---

func TestCrazyQuilt_MoveQuiltToWaste(t *testing.T) {
	c := newTestCrazyQuilt()
	clearCrazyQuiltBoard(c)
	fillCrazyQuilt(c)
	c.waste = []*Card{NewCard(CardDesignHeart, 5, true)}
	// Suit is irrelevant; only the rank has to be one away.
	c.quilt[cellIdx(0, 0)] = NewCard(CardDesignClover, 6, true)

	require.NoError(t, c.MoveQuiltToWaste(cellIdx(0, 0)))
	assert.Nil(t, c.GetQuilt()[cellIdx(0, 0)])
	assert.Len(t, c.GetWaste(), 2)
	assert.Equal(t, 6, c.GetWaste()[1].GetValue())
}

func TestCrazyQuilt_MoveQuiltToWaste_AcceptsEitherDirection(t *testing.T) {
	c := newTestCrazyQuilt()
	clearCrazyQuiltBoard(c)
	fillCrazyQuilt(c)
	c.waste = []*Card{NewCard(CardDesignHeart, 5, true)}
	c.quilt[cellIdx(0, 0)] = NewCard(CardDesignClover, 4, true)

	require.NoError(t, c.MoveQuiltToWaste(cellIdx(0, 0)), "one below is fine too")
	assert.Len(t, c.GetWaste(), 2)
}

// No wraparound: a King does not sit on an Ace.
func TestCrazyQuilt_AdjacentRank_DoesNotWrap(t *testing.T) {
	assert.True(t, crazyQuiltAdjacentRank(5, 4))
	assert.True(t, crazyQuiltAdjacentRank(4, 5))
	assert.False(t, crazyQuiltAdjacentRank(5, 5))
	assert.False(t, crazyQuiltAdjacentRank(5, 7))
	assert.False(t, crazyQuiltAdjacentRank(CardValueMax, 1), "K and A are not neighbours")
}

func TestCrazyQuilt_MoveQuiltToWaste_Rejections(t *testing.T) {
	c := newTestCrazyQuilt()
	clearCrazyQuiltBoard(c)
	fillCrazyQuilt(c)

	assert.Error(t, c.MoveQuiltToWaste(-1))
	assert.Error(t, c.MoveQuiltToWaste(CrazyQuiltCells))

	err := c.MoveQuiltToWaste(cellIdx(0, 0))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "waste is empty")

	c.waste = []*Card{NewCard(CardDesignHeart, 5, true)}

	// Boxed in, even though the rank is in sequence.
	c.quilt[cellIdx(2, 2)] = NewCard(CardDesignClover, 6, true)
	err = c.MoveQuiltToWaste(cellIdx(2, 2))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "short side")

	// Reachable, but the rank is two away.
	c.quilt[cellIdx(0, 0)] = NewCard(CardDesignClover, 7, true)
	err = c.MoveQuiltToWaste(cellIdx(0, 0))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sequence")

	c.quilt[cellIdx(0, 0)] = nil
	err = c.MoveQuiltToWaste(cellIdx(0, 0))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already empty")

	c.GiveUp()
	assert.Error(t, c.MoveQuiltToWaste(cellIdx(0, 2)))
}

// --- Draw and the redeal #5274 omits ---

func TestCrazyQuilt_Draw(t *testing.T) {
	c := newTestCrazyQuilt()
	before := c.GetStockCount()
	require.NoError(t, c.Draw())
	assert.Equal(t, before-1, c.GetStockCount())
	assert.Len(t, c.GetWaste(), 1)
	assert.True(t, c.CanUndo())
}

// The waste is turned over WITHOUT shuffling, so the order is preserved.
func TestCrazyQuilt_Draw_RedealsOnceWithoutShuffling(t *testing.T) {
	c := newTestCrazyQuilt()
	clearCrazyQuiltBoard(c)
	fillCrazyQuilt(c)
	c.waste = []*Card{
		NewCard(CardDesignHeart, 2, true),
		NewCard(CardDesignHeart, 3, true),
		NewCard(CardDesignHeart, 4, true),
	}
	require.Equal(t, CrazyQuiltRedealCnt, c.GetRedealsLeft())

	require.NoError(t, c.Draw())
	assert.Equal(t, 3, c.GetStockCount())
	assert.Empty(t, c.GetWaste())
	assert.Equal(t, 0, c.GetRedealsLeft())
	assert.Equal(t, 2, c.stock[0].GetValue(), "the order survives the turn-over")
	assert.Equal(t, 4, c.stock[2].GetValue())

	// A second redeal is refused.
	require.NoError(t, c.Draw())
	require.NoError(t, c.Draw())
	require.NoError(t, c.Draw())
	err := c.Draw()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no redeal is left")
}

func TestCrazyQuilt_Draw_NothingToRedeal(t *testing.T) {
	c := newTestCrazyQuilt()
	clearCrazyQuiltBoard(c)
	fillCrazyQuilt(c)
	err := c.Draw()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nothing to redeal")
}

func TestCrazyQuilt_Draw_RejectedWhenNotPlaying(t *testing.T) {
	c := newTestCrazyQuilt()
	c.GiveUp()
	assert.Error(t, c.Draw())
}

// --- Hints and stalemate ---

func TestCrazyQuilt_GetHint_PrefersFoundation(t *testing.T) {
	c := newTestCrazyQuilt()
	clearCrazyQuiltBoard(c)
	fillCrazyQuilt(c)
	c.foundation[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
	c.quilt[cellIdx(0, 0)] = NewCard(CardDesignSpade, 2, true)
	c.stock = []*Card{NewCard(CardDesignDiamond, 4, true)}

	h := c.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "quilt", h.FromZone)
	assert.Equal(t, cellIdx(0, 0), h.FromIdx)
	assert.Equal(t, "foundation", h.ToZone)
}

// Prising a card off the quilt opens its neighbour's short side, so it beats
// playing the waste even for the same single point.
func TestCrazyQuilt_FoundationHint_PrefersTheQuilt(t *testing.T) {
	c := newTestCrazyQuilt()
	clearCrazyQuiltBoard(c)
	fillCrazyQuilt(c)
	c.foundation[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
	c.foundation[1] = []*Card{NewCard(CardDesignClover, 1, true)}
	c.quilt[cellIdx(0, 0)] = NewCard(CardDesignSpade, 2, true)
	c.waste = []*Card{NewCard(CardDesignClover, 2, true)}

	h := c.foundationHint()
	require.NotNil(t, h)
	assert.Equal(t, "quilt", h.FromZone)

	// Negative control: with nothing playable on the quilt, the waste is used.
	c.quilt[cellIdx(0, 0)] = NewCard(CardDesignSpade, 9, true)
	h = c.foundationHint()
	require.NotNil(t, h)
	assert.Equal(t, "waste", h.FromZone)
}

func TestCrazyQuilt_GetHint_SuggestsBreakingTheQuiltOntoTheWaste(t *testing.T) {
	c := newTestCrazyQuilt()
	clearCrazyQuiltBoard(c)
	fillCrazyQuilt(c)
	c.waste = []*Card{NewCard(CardDesignHeart, 8, true)}
	c.quilt[cellIdx(0, 0)] = NewCard(CardDesignClover, 7, true)

	h := c.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "quilt", h.FromZone)
	assert.Equal(t, "waste", h.ToZone)
	assert.Equal(t, cellIdx(0, 0), h.FromIdx)
}

func TestCrazyQuilt_GetHint_DrawsWhenNothingElse(t *testing.T) {
	c := newTestCrazyQuilt()
	clearCrazyQuiltBoard(c)
	fillCrazyQuilt(c)
	c.stock = []*Card{NewCard(CardDesignDiamond, 4, true)}

	h := c.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "stock", h.FromZone)
	assert.Equal(t, "waste", h.ToZone)
}

// While a redeal is left, gathering the waste is still a move.
func TestCrazyQuilt_GetHint_OffersTheRedeal(t *testing.T) {
	c := newTestCrazyQuilt()
	clearCrazyQuiltBoard(c)
	fillCrazyQuilt(c)
	c.waste = []*Card{NewCard(CardDesignHeart, 2, true)}

	h := c.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "stock", h.FromZone)

	c.redealsLeft = 0
	assert.Nil(t, c.GetHint(), "with no redeal left there is nothing to do")
}

func TestCrazyQuilt_GetHint_NilWhenNotPlaying(t *testing.T) {
	c := newTestCrazyQuilt()
	c.GiveUp()
	assert.Nil(t, c.GetHint())
	assert.Nil(t, c.foundationHint())
}

func TestCrazyQuilt_Stalemate(t *testing.T) {
	c := newTestCrazyQuilt()
	clearCrazyQuiltBoard(c)
	fillCrazyQuilt(c)
	c.redealsLeft = 0
	c.checkStalemate()
	assert.True(t, c.IsStalemate())

	c.stock = []*Card{NewCard(CardDesignDiamond, 4, true)}
	c.checkStalemate()
	assert.False(t, c.IsStalemate())
}

func TestCrazyQuilt_Stalemate_NotSetOutsidePlaying(t *testing.T) {
	c := newTestCrazyQuilt()
	clearCrazyQuiltBoard(c)
	c.phase = CrazyQuiltPhaseGameOver
	c.checkStalemate()
	assert.False(t, c.IsStalemate())
}

func TestCrazyQuilt_UndoToEscape(t *testing.T) {
	c := newTestCrazyQuilt()
	clearCrazyQuiltBoard(c)
	fillCrazyQuilt(c)
	c.redealsLeft = 0
	c.foundation[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
	c.quilt[cellIdx(0, 0)] = NewCard(CardDesignSpade, 2, true)
	c.checkStalemate()
	require.False(t, c.IsStalemate())

	require.NoError(t, c.MoveQuiltToFoundation(cellIdx(0, 0)))
	assert.True(t, c.IsStalemate())
	assert.Equal(t, 1, c.UndoToEscape())
}

// --- AutoComplete / clear ---

func TestCrazyQuilt_AutoComplete(t *testing.T) {
	c := newTestCrazyQuilt()
	clearCrazyQuiltBoard(c)
	fillCrazyQuilt(c)
	c.foundation[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
	c.quilt[cellIdx(0, 0)] = NewCard(CardDesignSpade, 2, true)
	c.quilt[cellIdx(0, 2)] = NewCard(CardDesignSpade, 3, true)

	require.NoError(t, c.AutoComplete())
	assert.Len(t, c.GetFoundation()[0], 3)
	assert.Nil(t, c.GetQuilt()[cellIdx(0, 0)])
	assert.Nil(t, c.GetQuilt()[cellIdx(0, 2)])
}

func TestCrazyQuilt_AutoComplete_Rejections(t *testing.T) {
	c := newTestCrazyQuilt()
	clearCrazyQuiltBoard(c)
	fillCrazyQuilt(c)
	err := c.AutoComplete()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no card")

	c.GiveUp()
	assert.Error(t, c.AutoComplete())
}

func TestCrazyQuilt_GameClear(t *testing.T) {
	c := newTestCrazyQuilt()
	clearCrazyQuiltBoard(c)
	for i := range CrazyQuiltFoundationCnt {
		for n := range CrazyQuiltFoundationTarget {
			v := n + 1
			if !crazyQuiltIsAscending(i) {
				v = CardValueMax - n
			}
			c.foundation[i] = append(c.foundation[i], NewCard(crazyQuiltSuitOrder[i], v, true))
		}
	}
	c.foundation[7] = c.foundation[7][:CrazyQuiltFoundationTarget-1]
	c.checkGameClear()
	assert.Equal(t, CrazyQuiltPhasePlaying, c.GetPhase(), "one card still missing")

	// The descending diamond pile ends on the Ace.
	c.quilt[cellIdx(0, 0)] = NewCard(CardDesignDiamond, 1, true)
	require.NoError(t, c.MoveQuiltToFoundation(cellIdx(0, 0)))
	assert.Equal(t, CrazyQuiltPhaseGameClear, c.GetPhase())
	assert.True(t, c.GetGameEndFlag())
}

// --- Undo / log ---

func TestCrazyQuilt_Undo(t *testing.T) {
	c := newTestCrazyQuilt()
	clearCrazyQuiltBoard(c)
	fillCrazyQuilt(c)
	c.foundation[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
	c.quilt[cellIdx(0, 0)] = NewCard(CardDesignSpade, 2, true)

	assert.False(t, c.CanUndo())
	require.Error(t, c.Undo())

	require.NoError(t, c.MoveQuiltToFoundation(cellIdx(0, 0)))
	require.NoError(t, c.Undo())
	require.NotNil(t, c.GetQuilt()[cellIdx(0, 0)])
	assert.Equal(t, 2, c.GetQuilt()[cellIdx(0, 0)].GetValue())
	assert.Len(t, c.GetFoundation()[0], 1)
	assert.Equal(t, 0, c.GetMoveCount())
}

// Undoing a redeal has to give the redeal back, or the count silently drifts.
func TestCrazyQuilt_Undo_RestoresTheRedeal(t *testing.T) {
	c := newTestCrazyQuilt()
	clearCrazyQuiltBoard(c)
	fillCrazyQuilt(c)
	c.waste = []*Card{NewCard(CardDesignHeart, 2, true)}

	require.NoError(t, c.Draw())
	require.Equal(t, 0, c.GetRedealsLeft())

	require.NoError(t, c.Undo())
	assert.Equal(t, CrazyQuiltRedealCnt, c.GetRedealsLeft())
	assert.Len(t, c.GetWaste(), 1)
	assert.Equal(t, 0, c.GetStockCount())
}

func TestCrazyQuilt_UndoN(t *testing.T) {
	c := newTestCrazyQuilt()
	require.NoError(t, c.Draw())
	require.NoError(t, c.Draw())
	require.NoError(t, c.Draw())
	require.NoError(t, c.UndoN(2))
	assert.Equal(t, 1, c.GetMoveCount())
	assert.Len(t, c.GetWaste(), 1)

	assert.Error(t, c.UndoN(5))
	assert.Error(t, c.UndoN(0))
}

func TestCrazyQuilt_GiveUp(t *testing.T) {
	c := newTestCrazyQuilt()
	c.GiveUp()
	assert.Equal(t, CrazyQuiltPhaseGameOver, c.GetPhase())
	assert.True(t, c.GetGameEndFlag())
	require.NotEmpty(t, c.GetActionLog())

	before := len(c.GetActionLog())
	c.GiveUp()
	assert.Len(t, c.GetActionLog(), before)
}

func TestCrazyQuilt_ActionLog(t *testing.T) {
	c := newTestCrazyQuilt()
	require.NoError(t, c.Draw())
	log := c.GetActionLog()
	require.NotEmpty(t, log)
	assert.Equal(t, "draw", log[len(log)-1].ActionType)
}

// --- JSON round-trip ---

func TestCrazyQuilt_JSONRoundTrip(t *testing.T) {
	c := newTestCrazyQuilt()
	require.NoError(t, c.Draw())
	require.NoError(t, c.Draw())

	data, err := json.Marshal(c)
	require.NoError(t, err)

	restored := NewDefaultCrazyQuilt()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, c.GetStockCount(), restored.GetStockCount())
	assert.Len(t, restored.GetWaste(), len(c.GetWaste()))
	assert.Equal(t, c.GetMoveCount(), restored.GetMoveCount())
	assert.Equal(t, c.GetPhase(), restored.GetPhase())
	assert.Equal(t, c.GetRedealsLeft(), restored.GetRedealsLeft())
	for i := range CrazyQuiltCells {
		assert.Equal(t, c.GetQuilt()[i] == nil, restored.GetQuilt()[i] == nil, "cell %d", i)
	}
}

// The undo stack has to survive the trip: the Worker rebuilds the game from KV
// on every request, so an unpersisted history means Undo silently never works.
func TestCrazyQuilt_JSONRoundTripKeepsUndoHistory(t *testing.T) {
	c := newTestCrazyQuilt()
	require.NoError(t, c.Draw())
	require.NoError(t, c.Draw())
	wasteBefore := len(c.GetWaste())

	data, err := json.Marshal(c)
	require.NoError(t, err)
	restored := NewDefaultCrazyQuilt()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.True(t, restored.CanUndo())
	require.NoError(t, restored.Undo())
	assert.Len(t, restored.GetWaste(), wasteBefore-1)
	assert.NotNil(t, restored.GetQuilt()[0], "the snapshot must carry the board, not a blank")
}

func TestCrazyQuilt_UnmarshalJSON_Rejections(t *testing.T) {
	for _, tc := range []struct{ name, data string }{
		{"broken json", `{`},
		{"phase too low", `{"ps":-1}`},
		{"phase too high", `{"ps":99}`},
		{"negative move count", `{"mc":-1}`},
		{"negative redeals", `{"rd":-1}`},
		{"too many redeals", `{"rd":9}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Error(t, json.Unmarshal([]byte(tc.data), NewDefaultCrazyQuilt()))
		})
	}
}

func TestCrazyQuilt_UnmarshalJSON_RejectsOversizedArrays(t *testing.T) {
	big := make([]*Card, CrazyQuiltTotalCards+1)
	for i := range big {
		big[i] = NewCard(CardDesignSpade, 1, true)
	}
	overCap := make([]*Card, crazyQuiltMaxSliceLen+1)
	for i := range overCap {
		overCap[i] = NewCard(CardDesignSpade, 1, true)
	}

	t.Run("stock", func(t *testing.T) {
		data, err := json.Marshal(&crazyQuiltJSON{Stock: big})
		require.NoError(t, err)
		assert.Error(t, json.Unmarshal(data, NewDefaultCrazyQuilt()))
	})
	t.Run("waste", func(t *testing.T) {
		data, err := json.Marshal(&crazyQuiltJSON{Waste: big})
		require.NoError(t, err)
		assert.Error(t, json.Unmarshal(data, NewDefaultCrazyQuilt()))
	})
	t.Run("foundation pile", func(t *testing.T) {
		j := &crazyQuiltJSON{}
		j.Foundation[0] = make([]*Card, CrazyQuiltFoundationTarget+1)
		for i := range j.Foundation[0] {
			j.Foundation[0][i] = NewCard(CardDesignSpade, 1, true)
		}
		data, err := json.Marshal(j)
		require.NoError(t, err)
		assert.Error(t, json.Unmarshal(data, NewDefaultCrazyQuilt()))
	})
	t.Run("action log", func(t *testing.T) {
		data, err := json.Marshal(&crazyQuiltJSON{ActionLog: make([]*ActionLogEntry, crazyQuiltMaxSliceLen+1)})
		require.NoError(t, err)
		assert.Error(t, json.Unmarshal(data, NewDefaultCrazyQuilt()))
	})
	t.Run("history", func(t *testing.T) {
		j := &crazyQuiltJSON{History: make([]*crazyQuiltSnapshot, crazyQuiltMaxSliceLen+1)}
		for i := range j.History {
			j.History[i] = &crazyQuiltSnapshot{}
		}
		data, err := json.Marshal(j)
		require.NoError(t, err)
		assert.Error(t, json.Unmarshal(data, NewDefaultCrazyQuilt()))
	})
	t.Run("snapshot stock", func(t *testing.T) {
		data, err := json.Marshal(crazyQuiltSnapshotJSON{Stock: overCap})
		require.NoError(t, err)
		var s crazyQuiltSnapshot
		assert.Error(t, s.UnmarshalJSON(data))
	})
	t.Run("snapshot waste", func(t *testing.T) {
		data, err := json.Marshal(crazyQuiltSnapshotJSON{Waste: overCap})
		require.NoError(t, err)
		var s crazyQuiltSnapshot
		assert.Error(t, s.UnmarshalJSON(data))
	})
	t.Run("snapshot foundation", func(t *testing.T) {
		j := crazyQuiltSnapshotJSON{}
		j.Foundation[0] = overCap
		data, err := json.Marshal(j)
		require.NoError(t, err)
		var s crazyQuiltSnapshot
		assert.Error(t, s.UnmarshalJSON(data))
	})
	t.Run("snapshot redeals out of range", func(t *testing.T) {
		data, err := json.Marshal(crazyQuiltSnapshotJSON{RedealsLeft: 9})
		require.NoError(t, err)
		var s crazyQuiltSnapshot
		assert.Error(t, s.UnmarshalJSON(data))
	})
	t.Run("snapshot broken json", func(t *testing.T) {
		var s crazyQuiltSnapshot
		assert.Error(t, s.UnmarshalJSON([]byte(`{`)))
	})
}
