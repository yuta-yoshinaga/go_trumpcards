//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRoyalCotillion() *RoyalCotillion {
	c := NewDefaultRoyalCotillion()
	c.Reset()
	return c
}

// clearRoyalCotillionBoard wipes the dealt layout so a test can state exactly
// the position it cares about. Never assert on a freshly Reset board -- the
// deal is shuffled, so any such assertion is a hidden flake.
func clearRoyalCotillionBoard(c *RoyalCotillion) {
	c.stock = nil
	c.waste = nil
	c.isStalemate = false
	c.history = nil
	c.moveCount = 0
	c.phase = RoyalCotillionPhasePlaying
	for i := range RoyalCotillionFoundationCnt {
		c.foundation[i] = nil
	}
	for i := range RoyalCotillionTableauCnt {
		c.tableau[i] = nil
	}
	for i := range RoyalCotillionReserveCnt {
		c.reserve[i] = nil
	}
}

// fillRoyalCotillionSlots puts one dead card in every tableau slot so no gap
// exists. A gap changes which hint the game offers, so a board full of holes
// would test a position the game never reaches by accident.
func fillRoyalCotillionSlots(c *RoyalCotillion) {
	for i := range RoyalCotillionTableauCnt {
		c.tableau[i] = NewCard(CardDesignSpade, 7, true)
	}
}

func TestNewRoyalCotillion(t *testing.T) {
	assert.NotNil(t, NewRoyalCotillion(NewTrumpCardsWithDecks(2, 0)))
	assert.NotNil(t, NewDefaultRoyalCotillion())
}

// The deal is 16 single-card slots plus 4 reserve piles of 3, leaving 76.
// #5275 describes two 4x3 grids (24 cards), which is a different board.
func TestRoyalCotillion_Reset(t *testing.T) {
	c := newTestRoyalCotillion()

	for i, card := range c.GetTableau() {
		assert.NotNil(t, card, "slot %d", i)
	}
	for i, pile := range c.GetReserve() {
		assert.Len(t, pile, RoyalCotillionReserveDepth, "reserve %d", i)
	}
	assert.Equal(t, 76, c.GetStockCount())
	assert.Empty(t, c.GetWaste())

	// Eight foundations, not sixteen: each wraps and so takes all 13 ranks.
	assert.Equal(t, 8, RoyalCotillionFoundationCnt)
	assert.Equal(t, RoyalCotillionTotalCards, RoyalCotillionFoundationCnt*RoyalCotillionFoundationTarget)
	for i, pile := range c.GetFoundation() {
		assert.Empty(t, pile, "foundation %d", i)
	}

	assert.Equal(t, RoyalCotillionPhasePlaying, c.GetPhase())
	assert.Equal(t, 0, c.GetMoveCount())
	assert.True(t, c.AllFaceUp())
	assert.False(t, c.GetGameEndFlag())
}

func TestRoyalCotillion_ResetDealsEveryCard(t *testing.T) {
	for range 20 {
		c := newTestRoyalCotillion()
		total := c.GetStockCount() + len(c.GetWaste())
		for _, card := range c.GetTableau() {
			if card != nil {
				total++
			}
		}
		for _, pile := range c.GetReserve() {
			total += len(pile)
		}
		for _, pile := range c.GetFoundation() {
			total += len(pile)
		}
		assert.Equal(t, RoyalCotillionTotalCards, total)
	}
}

func TestRoyalCotillion_ResetTwiceIsClean(t *testing.T) {
	c := newTestRoyalCotillion()
	require.NoError(t, c.Draw())
	c.Reset()
	assert.Empty(t, c.GetWaste())
	assert.Equal(t, 0, c.GetMoveCount())
	assert.False(t, c.CanUndo())
	assert.Equal(t, 76, c.GetStockCount())
}

// --- The by-twos wraparound, which is the whole game ---

// A-start runs A,3,5,7,9,J,K,2,4,6,8,10,Q and 2-start runs the complement.
// #5275 stops the odd series at the King and the even at the Queen, giving
// 7- and 6-card piles; that needs 16 foundations and never wraps.
func TestRoyalCotillion_NthValue_WrapsThroughAllThirteen(t *testing.T) {
	odd := make([]int, RoyalCotillionFoundationTarget)
	even := make([]int, RoyalCotillionFoundationTarget)
	for n := range RoyalCotillionFoundationTarget {
		odd[n] = royalCotillionNthValue(1, n)
		even[n] = royalCotillionNthValue(2, n)
	}
	assert.Equal(t, []int{1, 3, 5, 7, 9, 11, 13, 2, 4, 6, 8, 10, 12}, odd)
	assert.Equal(t, []int{2, 4, 6, 8, 10, 12, 1, 3, 5, 7, 9, 11, 13}, even)

	// Each series covers every rank exactly once -- that is why eight piles
	// hold all 104 cards.
	assert.ElementsMatch(t, odd, even)
	seen := map[int]int{}
	for _, v := range odd {
		seen[v]++
	}
	assert.Len(t, seen, RoyalCotillionFoundationTarget)
}

func TestRoyalCotillion_IsOddFoundation(t *testing.T) {
	c := newTestRoyalCotillion()
	for i := range RoyalCotillionOddCnt {
		assert.True(t, c.IsOddFoundation(i), "foundation %d starts at the Ace", i)
	}
	for i := RoyalCotillionOddCnt; i < RoyalCotillionFoundationCnt; i++ {
		assert.False(t, c.IsOddFoundation(i), "foundation %d starts at the deuce", i)
	}
	assert.False(t, c.IsOddFoundation(-1))
	assert.False(t, c.IsOddFoundation(RoyalCotillionFoundationCnt))
}

func TestRoyalCotillion_CanPlaceOnFoundation_OddSeries(t *testing.T) {
	c := newTestRoyalCotillion()
	clearRoyalCotillionBoard(c)

	assert.True(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 1, true), 0))
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 2, true), 0))
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 3, true), 0))
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignHeart, 1, true), 0), "wrong suit")

	c.foundation[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
	assert.True(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 3, true), 0), "by twos, not by one")
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 2, true), 0))
}

// The King is the 7th card of an A-start pile, and the 8th is the deuce.
func TestRoyalCotillion_CanPlaceOnFoundation_WrapsAfterTheKing(t *testing.T) {
	c := newTestRoyalCotillion()
	clearRoyalCotillionBoard(c)
	for _, v := range []int{1, 3, 5, 7, 9, 11, 13} {
		c.foundation[0] = append(c.foundation[0], NewCard(CardDesignSpade, v, true))
	}
	assert.Len(t, c.foundation[0], 7)
	assert.True(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 2, true), 0), "K wraps to the deuce")
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 1, true), 0))
}

// The Queen is the 6th card of a 2-start pile, and the 7th is the Ace.
func TestRoyalCotillion_CanPlaceOnFoundation_WrapsAfterTheQueen(t *testing.T) {
	c := newTestRoyalCotillion()
	clearRoyalCotillionBoard(c)
	for _, v := range []int{2, 4, 6, 8, 10, 12} {
		c.foundation[4] = append(c.foundation[4], NewCard(CardDesignSpade, v, true))
	}
	assert.Len(t, c.foundation[4], 6)
	assert.True(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 1, true), 4), "Q wraps to the Ace")
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 13, true), 4))
}

func TestRoyalCotillion_CanPlaceOnFoundation_RejectsFullPileAndBadIndex(t *testing.T) {
	c := newTestRoyalCotillion()
	clearRoyalCotillionBoard(c)
	for n := range RoyalCotillionFoundationTarget {
		c.foundation[0] = append(c.foundation[0], NewCard(CardDesignSpade, royalCotillionNthValue(1, n), true))
	}
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 1, true), 0), "a complete pile takes nothing")
	assert.False(t, c.canPlaceOnFoundation(nil, 0))
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 1, true), -1))
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 1, true), RoyalCotillionFoundationCnt))
}

// Two decks give two of every card, and a suit's two piles each traverse all
// thirteen ranks, so the copies always land on different piles.
func TestRoyalCotillion_FindFoundation_SecondCopyGoesToTheOtherSeries(t *testing.T) {
	c := newTestRoyalCotillion()
	clearRoyalCotillionBoard(c)
	ace := NewCard(CardDesignSpade, 1, true)
	assert.Equal(t, 0, c.findFoundation(ace))

	c.foundation[0] = []*Card{ace}
	// The 2-start pile wants the Ace only as its 7th card, so a second Ace has
	// nowhere to go yet -- correct, and the reason the game can jam.
	assert.Equal(t, -1, c.findFoundation(NewCard(CardDesignSpade, 1, true)))

	for _, v := range []int{2, 4, 6, 8, 10, 12} {
		c.foundation[4] = append(c.foundation[4], NewCard(CardDesignSpade, v, true))
	}
	assert.Equal(t, 4, c.findFoundation(NewCard(CardDesignSpade, 1, true)))
	assert.Equal(t, -1, c.findFoundation(nil))
}

// --- Draw ---

func TestRoyalCotillion_Draw(t *testing.T) {
	c := newTestRoyalCotillion()
	before := c.GetStockCount()
	require.NoError(t, c.Draw())
	assert.Equal(t, before-1, c.GetStockCount())
	assert.Len(t, c.GetWaste(), 1)
	assert.Equal(t, 1, c.GetMoveCount())
	assert.True(t, c.CanUndo())
}

func TestRoyalCotillion_Draw_NoRedeal(t *testing.T) {
	c := newTestRoyalCotillion()
	clearRoyalCotillionBoard(c)
	fillRoyalCotillionSlots(c)
	c.waste = []*Card{NewCard(CardDesignHeart, 5, true)}
	err := c.Draw()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no redeal")
	assert.Len(t, c.GetWaste(), 1, "waste must not be recycled into the stock")
}

func TestRoyalCotillion_Draw_RejectedWhenNotPlaying(t *testing.T) {
	c := newTestRoyalCotillion()
	c.GiveUp()
	assert.Error(t, c.Draw())
}

// --- Moves to a foundation ---

func TestRoyalCotillion_MoveTableauToFoundation(t *testing.T) {
	c := newTestRoyalCotillion()
	clearRoyalCotillionBoard(c)
	fillRoyalCotillionSlots(c)
	c.tableau[3] = NewCard(CardDesignSpade, 1, true)

	require.NoError(t, c.MoveTableauToFoundation(3))
	assert.Len(t, c.GetFoundation()[0], 1)
	assert.Nil(t, c.GetTableau()[3], "the slot is now empty and refillable")
	assert.Equal(t, 1, c.GetMoveCount())
}

func TestRoyalCotillion_MoveTableauToFoundation_Rejections(t *testing.T) {
	c := newTestRoyalCotillion()
	clearRoyalCotillionBoard(c)
	fillRoyalCotillionSlots(c)

	assert.Error(t, c.MoveTableauToFoundation(-1))
	assert.Error(t, c.MoveTableauToFoundation(RoyalCotillionTableauCnt))

	c.tableau[2] = nil
	err := c.MoveTableauToFoundation(2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")

	err = c.MoveTableauToFoundation(0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "foundation")

	c.GiveUp()
	assert.Error(t, c.MoveTableauToFoundation(0))
}

func TestRoyalCotillion_MoveReserveToFoundation(t *testing.T) {
	c := newTestRoyalCotillion()
	clearRoyalCotillionBoard(c)
	fillRoyalCotillionSlots(c)
	c.reserve[1] = []*Card{
		NewCard(CardDesignHeart, 9, true),
		NewCard(CardDesignClover, 1, true),
	}

	require.NoError(t, c.MoveReserveToFoundation(1))
	assert.Len(t, c.GetFoundation()[1], 1)
	assert.Len(t, c.GetReserve()[1], 1, "only the top card leaves")
}

func TestRoyalCotillion_MoveReserveToFoundation_Rejections(t *testing.T) {
	c := newTestRoyalCotillion()
	clearRoyalCotillionBoard(c)
	fillRoyalCotillionSlots(c)

	assert.Error(t, c.MoveReserveToFoundation(-1))
	assert.Error(t, c.MoveReserveToFoundation(RoyalCotillionReserveCnt))

	err := c.MoveReserveToFoundation(0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")

	c.reserve[0] = []*Card{NewCard(CardDesignSpade, 7, true)}
	err = c.MoveReserveToFoundation(0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "foundation")

	c.GiveUp()
	assert.Error(t, c.MoveReserveToFoundation(0))
}

func TestRoyalCotillion_MoveWasteToFoundation(t *testing.T) {
	c := newTestRoyalCotillion()
	clearRoyalCotillionBoard(c)
	fillRoyalCotillionSlots(c)
	c.waste = []*Card{NewCard(CardDesignHeart, 1, true)}

	require.NoError(t, c.MoveWasteToFoundation())
	assert.Len(t, c.GetFoundation()[2], 1)
	assert.Empty(t, c.GetWaste())
}

func TestRoyalCotillion_MoveWasteToFoundation_Rejections(t *testing.T) {
	c := newTestRoyalCotillion()
	clearRoyalCotillionBoard(c)
	fillRoyalCotillionSlots(c)

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

// --- Refilling a slot ---

// A tableau slot holds exactly one card, so it is refillable only when empty.
func TestRoyalCotillion_MoveWasteToTableau(t *testing.T) {
	c := newTestRoyalCotillion()
	clearRoyalCotillionBoard(c)
	fillRoyalCotillionSlots(c)
	c.tableau[5] = nil
	c.waste = []*Card{NewCard(CardDesignHeart, 2, true)}

	require.NoError(t, c.MoveWasteToTableau(5))
	require.NotNil(t, c.GetTableau()[5])
	assert.Equal(t, 2, c.GetTableau()[5].GetValue())
	assert.Empty(t, c.GetWaste())
}

func TestRoyalCotillion_MoveWasteToTableau_Rejections(t *testing.T) {
	c := newTestRoyalCotillion()
	clearRoyalCotillionBoard(c)
	fillRoyalCotillionSlots(c)

	assert.Error(t, c.MoveWasteToTableau(-1))
	assert.Error(t, c.MoveWasteToTableau(RoyalCotillionTableauCnt))

	c.waste = []*Card{NewCard(CardDesignHeart, 2, true)}
	err := c.MoveWasteToTableau(0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already holds a card")

	c.tableau[0] = nil
	c.waste = nil
	err = c.MoveWasteToTableau(0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "waste is empty")

	c.waste = []*Card{NewCard(CardDesignHeart, 2, true)}
	c.GiveUp()
	assert.Error(t, c.MoveWasteToTableau(0))
}

func TestRoyalCotillion_MoveStockToTableau(t *testing.T) {
	c := newTestRoyalCotillion()
	clearRoyalCotillionBoard(c)
	fillRoyalCotillionSlots(c)
	c.tableau[9] = nil
	c.stock = []*Card{NewCard(CardDesignDiamond, 4, true), NewCard(CardDesignDiamond, 5, true)}

	require.NoError(t, c.MoveStockToTableau(9))
	require.NotNil(t, c.GetTableau()[9])
	assert.Equal(t, 4, c.GetTableau()[9].GetValue(), "the stock is taken from the top")
	assert.Equal(t, 1, c.GetStockCount())
	assert.Empty(t, c.GetWaste(), "the card must not pass through the waste")
}

func TestRoyalCotillion_MoveStockToTableau_Rejections(t *testing.T) {
	c := newTestRoyalCotillion()
	clearRoyalCotillionBoard(c)
	fillRoyalCotillionSlots(c)

	assert.Error(t, c.MoveStockToTableau(-1))
	assert.Error(t, c.MoveStockToTableau(RoyalCotillionTableauCnt))

	c.stock = []*Card{NewCard(CardDesignDiamond, 4, true)}
	err := c.MoveStockToTableau(0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already holds a card")

	c.tableau[0] = nil
	c.stock = nil
	err = c.MoveStockToTableau(0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "stock is empty")

	c.stock = []*Card{NewCard(CardDesignDiamond, 4, true)}
	c.GiveUp()
	assert.Error(t, c.MoveStockToTableau(0))
}

// An emptied reserve pile stays empty for good -- nothing refills it.
func TestRoyalCotillion_ReserveIsNeverRefilled(t *testing.T) {
	c := newTestRoyalCotillion()
	clearRoyalCotillionBoard(c)
	fillRoyalCotillionSlots(c)
	c.reserve[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
	c.stock = []*Card{NewCard(CardDesignDiamond, 4, true)}
	c.waste = []*Card{NewCard(CardDesignHeart, 6, true)}

	require.NoError(t, c.MoveReserveToFoundation(0))
	assert.Empty(t, c.GetReserve()[0])

	// Neither the stock nor the waste has any way to reach it: the only move
	// that targets a reserve index is the one that empties it.
	require.NoError(t, c.Draw())
	assert.Empty(t, c.GetReserve()[0], "the reserve stays empty for the rest of the game")
}

// --- Hints and stalemate ---

func TestRoyalCotillion_GetHint_PrefersFoundation(t *testing.T) {
	c := newTestRoyalCotillion()
	clearRoyalCotillionBoard(c)
	fillRoyalCotillionSlots(c)
	c.tableau[6] = NewCard(CardDesignClover, 1, true)
	c.stock = []*Card{NewCard(CardDesignDiamond, 4, true)}

	h := c.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "tableau", h.FromZone)
	assert.Equal(t, 6, h.FromIdx)
	assert.Equal(t, "foundation", h.ToZone)
	assert.Equal(t, 1, h.ToIdx)
}

// A tableau card outranks a reserve card of equal value: emptying a slot opens
// an outlet for the stock, while emptying a reserve pile closes one for good.
func TestRoyalCotillion_FoundationHint_PrefersTheTableau(t *testing.T) {
	c := newTestRoyalCotillion()
	clearRoyalCotillionBoard(c)
	fillRoyalCotillionSlots(c)
	c.reserve[0] = []*Card{NewCard(CardDesignClover, 1, true)}
	c.tableau[2] = NewCard(CardDesignSpade, 1, true)

	h := c.foundationHint()
	require.NotNil(t, h)
	assert.Equal(t, "tableau", h.FromZone)
	assert.Equal(t, 2, h.FromIdx)

	// Negative control: with no tableau candidate the reserve is offered.
	c.tableau[2] = NewCard(CardDesignSpade, 7, true)
	h = c.foundationHint()
	require.NotNil(t, h)
	assert.Equal(t, "reserve", h.FromZone)
	assert.Equal(t, 0, h.FromIdx)
}

func TestRoyalCotillion_GetHint_FillsASlotFromTheStock(t *testing.T) {
	c := newTestRoyalCotillion()
	clearRoyalCotillionBoard(c)
	fillRoyalCotillionSlots(c)
	c.tableau[4] = nil
	c.stock = []*Card{NewCard(CardDesignDiamond, 4, true)}

	h := c.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "stock", h.FromZone)
	assert.Equal(t, "tableau", h.ToZone)
	assert.Equal(t, 4, h.ToIdx)
}

func TestRoyalCotillion_GetHint_FillsASlotFromTheWasteWhenTheStockIsGone(t *testing.T) {
	c := newTestRoyalCotillion()
	clearRoyalCotillionBoard(c)
	fillRoyalCotillionSlots(c)
	c.tableau[4] = nil
	c.waste = []*Card{NewCard(CardDesignHeart, 6, true)}

	h := c.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "waste", h.FromZone)
	assert.Equal(t, "tableau", h.ToZone)
	assert.Equal(t, 4, h.ToIdx)
}

func TestRoyalCotillion_GetHint_DrawsWhenNothingElse(t *testing.T) {
	c := newTestRoyalCotillion()
	clearRoyalCotillionBoard(c)
	fillRoyalCotillionSlots(c)
	c.stock = []*Card{NewCard(CardDesignDiamond, 4, true)}

	h := c.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "stock", h.FromZone)
	assert.Equal(t, "waste", h.ToZone)
	assert.Equal(t, -1, h.ToIdx)
}

func TestRoyalCotillion_GetHint_NilWhenNotPlaying(t *testing.T) {
	c := newTestRoyalCotillion()
	c.GiveUp()
	assert.Nil(t, c.GetHint())
	assert.Nil(t, c.foundationHint())
}

func TestRoyalCotillion_Stalemate(t *testing.T) {
	c := newTestRoyalCotillion()
	clearRoyalCotillionBoard(c)
	fillRoyalCotillionSlots(c)
	c.checkStalemate()
	assert.True(t, c.IsStalemate(), "no stock, no waste, and nothing plays")

	c.stock = []*Card{NewCard(CardDesignDiamond, 4, true)}
	c.checkStalemate()
	assert.False(t, c.IsStalemate())
}

func TestRoyalCotillion_Stalemate_NotSetOutsidePlaying(t *testing.T) {
	c := newTestRoyalCotillion()
	clearRoyalCotillionBoard(c)
	c.phase = RoyalCotillionPhaseGameOver
	c.checkStalemate()
	assert.False(t, c.IsStalemate())
}

func TestRoyalCotillion_UndoToEscape(t *testing.T) {
	c := newTestRoyalCotillion()
	clearRoyalCotillionBoard(c)
	fillRoyalCotillionSlots(c)
	c.tableau[0] = NewCard(CardDesignSpade, 1, true)
	c.checkStalemate()
	require.False(t, c.IsStalemate())

	// Sending the ace up leaves an empty slot with no stock and no waste to
	// fill it, and nothing else can move.
	require.NoError(t, c.MoveTableauToFoundation(0))
	assert.True(t, c.IsStalemate())
	assert.Equal(t, 1, c.UndoToEscape())
}

// --- AutoComplete / clear ---

func TestRoyalCotillion_AutoComplete(t *testing.T) {
	c := newTestRoyalCotillion()
	clearRoyalCotillionBoard(c)
	fillRoyalCotillionSlots(c)
	c.tableau[0] = NewCard(CardDesignClover, 1, true)
	c.reserve[0] = []*Card{NewCard(CardDesignClover, 3, true)}
	c.waste = []*Card{NewCard(CardDesignHeart, 1, true)}

	require.NoError(t, c.AutoComplete())
	assert.Len(t, c.GetFoundation()[1], 2, "the ace then the three on top of it")
	assert.Len(t, c.GetFoundation()[2], 1, "the waste ace too")
	assert.Empty(t, c.GetReserve()[0])
}

func TestRoyalCotillion_AutoComplete_Rejections(t *testing.T) {
	c := newTestRoyalCotillion()
	clearRoyalCotillionBoard(c)
	fillRoyalCotillionSlots(c)
	err := c.AutoComplete()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no card")

	c.GiveUp()
	assert.Error(t, c.AutoComplete())
}

func TestRoyalCotillion_GameClear(t *testing.T) {
	c := newTestRoyalCotillion()
	clearRoyalCotillionBoard(c)
	for i := range RoyalCotillionFoundationCnt {
		base := 1
		if !royalCotillionIsOdd(i) {
			base = 2
		}
		for n := range RoyalCotillionFoundationTarget {
			c.foundation[i] = append(c.foundation[i],
				NewCard(royalCotillionSuitOrder[i], royalCotillionNthValue(base, n), true))
		}
	}
	c.foundation[7] = c.foundation[7][:RoyalCotillionFoundationTarget-1]
	c.checkGameClear()
	assert.Equal(t, RoyalCotillionPhasePlaying, c.GetPhase(), "one card still missing")

	// The 2-start diamond pile ends on the King.
	c.tableau[0] = NewCard(CardDesignDiamond, 13, true)
	require.NoError(t, c.MoveTableauToFoundation(0))
	assert.Equal(t, RoyalCotillionPhaseGameClear, c.GetPhase())
	assert.True(t, c.GetGameEndFlag())
}

// --- Undo / log ---

func TestRoyalCotillion_Undo(t *testing.T) {
	c := newTestRoyalCotillion()
	clearRoyalCotillionBoard(c)
	fillRoyalCotillionSlots(c)
	c.tableau[1] = NewCard(CardDesignSpade, 1, true)

	assert.False(t, c.CanUndo())
	require.Error(t, c.Undo())

	require.NoError(t, c.MoveTableauToFoundation(1))
	require.NoError(t, c.Undo())
	require.NotNil(t, c.GetTableau()[1])
	assert.Equal(t, 1, c.GetTableau()[1].GetValue())
	assert.Empty(t, c.GetFoundation()[0])
	assert.Equal(t, 0, c.GetMoveCount())
}

func TestRoyalCotillion_Undo_RestoresTheReserve(t *testing.T) {
	c := newTestRoyalCotillion()
	clearRoyalCotillionBoard(c)
	fillRoyalCotillionSlots(c)
	c.reserve[2] = []*Card{NewCard(CardDesignHeart, 9, true), NewCard(CardDesignSpade, 1, true)}

	require.NoError(t, c.MoveReserveToFoundation(2))
	require.NoError(t, c.Undo())
	assert.Len(t, c.GetReserve()[2], 2, "a reserve that cannot refill must come back on undo")
}

func TestRoyalCotillion_UndoN(t *testing.T) {
	c := newTestRoyalCotillion()
	require.NoError(t, c.Draw())
	require.NoError(t, c.Draw())
	require.NoError(t, c.Draw())
	require.NoError(t, c.UndoN(2))
	assert.Equal(t, 1, c.GetMoveCount())
	assert.Len(t, c.GetWaste(), 1)

	assert.Error(t, c.UndoN(5))
	assert.Error(t, c.UndoN(0))
}

func TestRoyalCotillion_GiveUp(t *testing.T) {
	c := newTestRoyalCotillion()
	c.GiveUp()
	assert.Equal(t, RoyalCotillionPhaseGameOver, c.GetPhase())
	assert.True(t, c.GetGameEndFlag())
	require.NotEmpty(t, c.GetActionLog())

	before := len(c.GetActionLog())
	c.GiveUp()
	assert.Len(t, c.GetActionLog(), before)
}

func TestRoyalCotillion_ActionLog(t *testing.T) {
	c := newTestRoyalCotillion()
	require.NoError(t, c.Draw())
	log := c.GetActionLog()
	require.NotEmpty(t, log)
	assert.Equal(t, "draw", log[len(log)-1].ActionType)
}

// --- JSON round-trip ---

func TestRoyalCotillion_JSONRoundTrip(t *testing.T) {
	c := newTestRoyalCotillion()
	require.NoError(t, c.Draw())
	require.NoError(t, c.Draw())

	data, err := json.Marshal(c)
	require.NoError(t, err)

	restored := NewDefaultRoyalCotillion()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, c.GetStockCount(), restored.GetStockCount())
	assert.Len(t, restored.GetWaste(), len(c.GetWaste()))
	assert.Equal(t, c.GetMoveCount(), restored.GetMoveCount())
	assert.Equal(t, c.GetPhase(), restored.GetPhase())
	for i := range RoyalCotillionReserveCnt {
		assert.Len(t, restored.GetReserve()[i], len(c.GetReserve()[i]), "reserve %d", i)
	}
	for i := range RoyalCotillionTableauCnt {
		assert.Equal(t, c.GetTableau()[i] == nil, restored.GetTableau()[i] == nil, "slot %d", i)
	}
}

// The undo stack has to survive the trip: the Worker rebuilds the game from KV
// on every request, so an unpersisted history means Undo silently never works.
func TestRoyalCotillion_JSONRoundTripKeepsUndoHistory(t *testing.T) {
	c := newTestRoyalCotillion()
	require.NoError(t, c.Draw())
	require.NoError(t, c.Draw())
	wasteBefore := len(c.GetWaste())

	data, err := json.Marshal(c)
	require.NoError(t, err)
	restored := NewDefaultRoyalCotillion()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.True(t, restored.CanUndo())
	require.NoError(t, restored.Undo())
	assert.Len(t, restored.GetWaste(), wasteBefore-1)
	assert.NotNil(t, restored.GetTableau()[0], "the snapshot must carry the board, not a blank")
	assert.Len(t, restored.GetReserve()[0], RoyalCotillionReserveDepth)
}

func TestRoyalCotillion_UnmarshalJSON_Rejections(t *testing.T) {
	for _, tc := range []struct{ name, data string }{
		{"broken json", `{`},
		{"phase too low", `{"ps":-1}`},
		{"phase too high", `{"ps":99}`},
		{"negative move count", `{"mc":-1}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Error(t, json.Unmarshal([]byte(tc.data), NewDefaultRoyalCotillion()))
		})
	}
}

func TestRoyalCotillion_UnmarshalJSON_RejectsOversizedArrays(t *testing.T) {
	big := make([]*Card, RoyalCotillionTotalCards+1)
	for i := range big {
		big[i] = NewCard(CardDesignSpade, 1, true)
	}
	overCap := make([]*Card, royalCotillionMaxSliceLen+1)
	for i := range overCap {
		overCap[i] = NewCard(CardDesignSpade, 1, true)
	}

	t.Run("stock", func(t *testing.T) {
		data, err := json.Marshal(&royalCotillionJSON{Stock: big})
		require.NoError(t, err)
		assert.Error(t, json.Unmarshal(data, NewDefaultRoyalCotillion()))
	})
	t.Run("waste", func(t *testing.T) {
		data, err := json.Marshal(&royalCotillionJSON{Waste: big})
		require.NoError(t, err)
		assert.Error(t, json.Unmarshal(data, NewDefaultRoyalCotillion()))
	})
	t.Run("reserve pile", func(t *testing.T) {
		j := &royalCotillionJSON{}
		j.Reserve[0] = big
		data, err := json.Marshal(j)
		require.NoError(t, err)
		assert.Error(t, json.Unmarshal(data, NewDefaultRoyalCotillion()))
	})
	t.Run("foundation pile", func(t *testing.T) {
		j := &royalCotillionJSON{}
		j.Foundation[0] = make([]*Card, RoyalCotillionFoundationTarget+1)
		for i := range j.Foundation[0] {
			j.Foundation[0][i] = NewCard(CardDesignSpade, 1, true)
		}
		data, err := json.Marshal(j)
		require.NoError(t, err)
		assert.Error(t, json.Unmarshal(data, NewDefaultRoyalCotillion()))
	})
	t.Run("action log", func(t *testing.T) {
		data, err := json.Marshal(&royalCotillionJSON{ActionLog: make([]*ActionLogEntry, royalCotillionMaxSliceLen+1)})
		require.NoError(t, err)
		assert.Error(t, json.Unmarshal(data, NewDefaultRoyalCotillion()))
	})
	t.Run("history", func(t *testing.T) {
		j := &royalCotillionJSON{History: make([]*royalCotillionSnapshot, royalCotillionMaxSliceLen+1)}
		for i := range j.History {
			j.History[i] = &royalCotillionSnapshot{}
		}
		data, err := json.Marshal(j)
		require.NoError(t, err)
		assert.Error(t, json.Unmarshal(data, NewDefaultRoyalCotillion()))
	})
	t.Run("snapshot stock", func(t *testing.T) {
		data, err := json.Marshal(royalCotillionSnapshotJSON{Stock: overCap})
		require.NoError(t, err)
		var s royalCotillionSnapshot
		assert.Error(t, s.UnmarshalJSON(data))
	})
	t.Run("snapshot waste", func(t *testing.T) {
		data, err := json.Marshal(royalCotillionSnapshotJSON{Waste: overCap})
		require.NoError(t, err)
		var s royalCotillionSnapshot
		assert.Error(t, s.UnmarshalJSON(data))
	})
	t.Run("snapshot reserve", func(t *testing.T) {
		j := royalCotillionSnapshotJSON{}
		j.Reserve[0] = overCap
		data, err := json.Marshal(j)
		require.NoError(t, err)
		var s royalCotillionSnapshot
		assert.Error(t, s.UnmarshalJSON(data))
	})
	t.Run("snapshot foundation", func(t *testing.T) {
		j := royalCotillionSnapshotJSON{}
		j.Foundation[0] = overCap
		data, err := json.Marshal(j)
		require.NoError(t, err)
		var s royalCotillionSnapshot
		assert.Error(t, s.UnmarshalJSON(data))
	})
	t.Run("snapshot broken json", func(t *testing.T) {
		var s royalCotillionSnapshot
		assert.Error(t, s.UnmarshalJSON([]byte(`{`)))
	})
}
