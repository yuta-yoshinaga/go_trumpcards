//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestAmericanToad() *AmericanToad {
	at := NewDefaultAmericanToad()
	at.Reset()
	return at
}

// clearAmericanToadBoard wipes the dealt layout so a test can state exactly the
// position it cares about. Never assert on a freshly Reset board -- the deal is
// shuffled, so any such assertion is a hidden flake.
func clearAmericanToadBoard(at *AmericanToad) {
	at.reserve = nil
	at.stock = nil
	at.waste = nil
	at.passesUsed = 0
	at.isStalemate = false
	at.baseRank = 1
	for i := range AmericanToadFoundationCnt {
		at.foundation[i] = nil
	}
	for i := range AmericanToadTableauCnt {
		at.tableau[i] = nil
	}
}

// atCol builds a tableau column from cards.
func atCol(cards ...*Card) []*AmericanToadTableauCard {
	col := make([]*AmericanToadTableauCard, len(cards))
	for i, c := range cards {
		col[i] = &AmericanToadTableauCard{Card: c, FaceUp: true}
	}
	return col
}

// fillAmericanToadColumns puts one dead card in every column so no gap exists.
// A real board only ever has a gap for the instant before it is auto-refilled,
// so a test board full of holes would make fillEmptyColumnsFromReserve look
// wrong when it is only doing its job.
func fillAmericanToadColumns(at *AmericanToad) {
	for i := range AmericanToadTableauCnt {
		at.tableau[i] = atCol(NewCard(CardDesignSpade, 9, true))
	}
}

func TestNewAmericanToad(t *testing.T) {
	assert.NotNil(t, NewAmericanToad(NewTrumpCardsWithDecks(2, 0)))
	assert.NotNil(t, NewDefaultAmericanToad())
}

func TestAmericanToad_Reset(t *testing.T) {
	at := newTestAmericanToad()

	// 20 reserve + 8 tableau + 1 starter + 75 stock = 104.
	assert.Len(t, at.GetReserve(), AmericanToadReserveSize)
	for i, col := range at.GetTableau() {
		assert.Len(t, col, 1, "column %d", i)
	}
	assert.Equal(t,
		AmericanToadTotalCards-AmericanToadReserveSize-AmericanToadTableauCnt-1,
		at.GetStockCount())
	assert.Empty(t, at.GetWaste())

	// Exactly one foundation is open, and its card sets the base rank.
	opened := 0
	for _, pile := range at.GetFoundation() {
		if len(pile) > 0 {
			opened++
			assert.Len(t, pile, 1)
			assert.Equal(t, at.GetBaseRank(), pile[0].GetValue())
		}
	}
	assert.Equal(t, 1, opened)
	assert.NotZero(t, at.GetBaseRank())

	assert.Equal(t, AmericanToadPhasePlaying, at.GetPhase())
	assert.Equal(t, 0, at.GetMoveCount())
	assert.Equal(t, 0, at.GetPassesUsed())
	assert.True(t, at.AllFaceUp())
	assert.False(t, at.GetGameEndFlag())
}

// An off-by-one in the deal would silently drop or duplicate a card.
func TestAmericanToad_ResetDealsEveryCard(t *testing.T) {
	for range 20 {
		at := newTestAmericanToad()
		total := len(at.GetReserve()) + at.GetStockCount() + len(at.GetWaste())
		for _, col := range at.GetTableau() {
			total += len(col)
		}
		for _, pile := range at.GetFoundation() {
			total += len(pile)
		}
		assert.Equal(t, AmericanToadTotalCards, total)
	}
}

func TestAmericanToad_ResetIsRepeatable(t *testing.T) {
	at := newTestAmericanToad()
	require.NoError(t, at.Draw())
	at.Reset()
	assert.Equal(t, 0, at.GetMoveCount())
	assert.Empty(t, at.GetWaste())
	assert.Empty(t, at.GetActionLog())
	assert.False(t, at.CanUndo())
}

// Eight foundations of 13 absorb all 104 cards. #4417's "up to 8 per suit"
// comes to 32, which cannot finish the game.
func TestAmericanToad_FoundationRules(t *testing.T) {
	at := newTestAmericanToad()
	clearAmericanToadBoard(at)
	at.baseRank = 5

	t.Run("an empty foundation takes only the base rank of its own suit", func(t *testing.T) {
		assert.True(t, at.canPlaceOnFoundation(NewCard(CardDesignSpade, 5, true), 0))
		assert.False(t, at.canPlaceOnFoundation(NewCard(CardDesignSpade, 6, true), 0))
		// Foundation 0 is pinned to spades, so a heart cannot open it.
		assert.False(t, at.canPlaceOnFoundation(NewCard(CardDesignHeart, 5, true), 0))
	})

	t.Run("there are two foundations per suit", func(t *testing.T) {
		at.foundation[0] = []*Card{NewCard(CardDesignSpade, 5, true)}
		assert.Equal(t, 4, at.findFoundation(NewCard(CardDesignSpade, 5, true)),
			"the second spade base card opens the second spade foundation")
	})

	t.Run("builds up in suit wrapping King to Ace", func(t *testing.T) {
		at.foundation[1] = []*Card{NewCard(CardDesignClover, CardValueMax, true)}
		assert.True(t, at.canPlaceOnFoundation(NewCard(CardDesignClover, 1, true), 1))
		assert.False(t, at.canPlaceOnFoundation(NewCard(CardDesignClover, 2, true), 1))
	})

	t.Run("stops at thirteen cards", func(t *testing.T) {
		at.foundation[2] = nil
		v := 5
		for range AmericanToadFoundationTarget {
			at.foundation[2] = append(at.foundation[2], NewCard(CardDesignHeart, v, true))
			v = americanToadNextRank(v)
		}
		require.Len(t, at.foundation[2], AmericanToadFoundationTarget)
		for r := 1; r <= CardValueMax; r++ {
			assert.False(t, at.canPlaceOnFoundation(NewCard(CardDesignHeart, r, true), 2), "rank %d", r)
		}
	})

	t.Run("nothing is placeable before the base rank is known", func(t *testing.T) {
		at.baseRank = 0
		assert.False(t, at.canPlaceOnFoundation(NewCard(CardDesignSpade, 5, true), 0))
		assert.False(t, at.canPlaceOnFoundation(nil, 0))
	})
}

// The tableau builds down in suit and wraps Ace round to King -- #4417 mentions
// only "descending in suit".
func TestAmericanToad_TableauBuildRules(t *testing.T) {
	at := newTestAmericanToad()
	clearAmericanToadBoard(at)
	fillAmericanToadColumns(at)
	at.tableau[0] = atCol(NewCard(CardDesignSpade, 7, true))

	assert.True(t, at.canPlaceOnTableau(NewCard(CardDesignSpade, 6, true), 0))
	assert.False(t, at.canPlaceOnTableau(NewCard(CardDesignHeart, 6, true), 0), "suit must match")
	assert.False(t, at.canPlaceOnTableau(NewCard(CardDesignSpade, 5, true), 0))
	assert.False(t, at.canPlaceOnTableau(nil, 0))

	at.tableau[1] = atCol(NewCard(CardDesignClover, 1, true))
	assert.True(t, at.canPlaceOnTableau(NewCard(CardDesignClover, CardValueMax, true), 1),
		"a King goes under an Ace")
}

func TestAmericanToad_Draw(t *testing.T) {
	t.Run("moves one card to the waste", func(t *testing.T) {
		at := newTestAmericanToad()
		before := at.GetStockCount()
		require.NoError(t, at.Draw())
		assert.Equal(t, before-1, at.GetStockCount())
		assert.Len(t, at.GetWaste(), 1)
	})

	// One redeal is allowed -- two passes in total. #4417 does not mention it.
	t.Run("recycles the waste once and then refuses", func(t *testing.T) {
		at := newTestAmericanToad()
		clearAmericanToadBoard(at)
		fillAmericanToadColumns(at)
		at.waste = []*Card{NewCard(CardDesignSpade, 4, true), NewCard(CardDesignHeart, 7, true)}

		require.NoError(t, at.Draw())
		assert.Equal(t, 2, at.GetStockCount())
		assert.Empty(t, at.GetWaste())
		assert.Equal(t, 1, at.GetPassesUsed())
		assert.False(t, at.CanRedeal())

		require.NoError(t, at.Draw())
		require.NoError(t, at.Draw())
		require.Empty(t, at.stock)
		assert.Error(t, at.Draw(), "a second redeal is refused")
	})

	t.Run("refuses a redeal with an empty waste", func(t *testing.T) {
		at := newTestAmericanToad()
		clearAmericanToadBoard(at)
		assert.False(t, at.CanRedeal())
		assert.Error(t, at.Draw())
	})

	t.Run("errors once the game has ended", func(t *testing.T) {
		at := newTestAmericanToad()
		at.GiveUp()
		assert.Error(t, at.Draw())
		assert.False(t, at.CanRedeal())
	})
}

// A gap is refilled from the reserve the instant it appears; the player never
// sees an empty column while the reserve holds cards.
func TestAmericanToad_EmptyColumnAutoFillsFromReserve(t *testing.T) {
	at := newTestAmericanToad()
	clearAmericanToadBoard(at)
	fillAmericanToadColumns(at)
	at.baseRank = 5
	at.foundation[0] = []*Card{NewCard(CardDesignSpade, 5, true)}
	at.tableau[3] = atCol(NewCard(CardDesignSpade, 6, true))
	refill := NewCard(CardDesignDiamond, 2, true)
	at.reserve = []*Card{NewCard(CardDesignClover, 3, true), refill}

	require.NoError(t, at.MoveTableauToFoundation(3))
	assert.Len(t, at.GetTableau()[3], 1, "the gap closed immediately")
	assert.Equal(t, refill, at.GetTableau()[3][0].Card)
	assert.Len(t, at.GetReserve(), 1)
}

// Once the reserve is gone the gap stays open, and only the waste may fill it.
// #4417 says empty columns are reserve-only, which would strand them forever.
func TestAmericanToad_EmptyColumnRules(t *testing.T) {
	t.Run("a tableau card may not fill an empty column", func(t *testing.T) {
		at := newTestAmericanToad()
		clearAmericanToadBoard(at)
		fillAmericanToadColumns(at)
		at.tableau[0] = atCol(NewCard(CardDesignSpade, 7, true))
		at.tableau[1] = nil

		assert.Error(t, at.MoveTableauToTableau(0, -1, 1))
		assert.Empty(t, at.GetTableau()[1])
	})

	t.Run("the waste may fill an empty column once the reserve is gone", func(t *testing.T) {
		at := newTestAmericanToad()
		clearAmericanToadBoard(at)
		fillAmericanToadColumns(at)
		at.tableau[2] = nil
		at.waste = []*Card{NewCard(CardDesignHeart, 8, true)}

		require.NoError(t, at.MoveWasteToTableau(2))
		assert.Len(t, at.GetTableau()[2], 1)
	})

	// While the reserve still holds cards the gap belongs to it, so a manual
	// placement into an empty column is refused.
	t.Run("the waste may not pre-empt the reserve", func(t *testing.T) {
		at := newTestAmericanToad()
		clearAmericanToadBoard(at)
		fillAmericanToadColumns(at)
		at.tableau[2] = nil
		at.reserve = []*Card{NewCard(CardDesignClover, 3, true)}
		at.waste = []*Card{NewCard(CardDesignHeart, 8, true)}

		assert.Error(t, at.MoveWasteToTableau(2))
	})
}

func TestAmericanToad_MoveReserve(t *testing.T) {
	t.Run("to a foundation", func(t *testing.T) {
		at := newTestAmericanToad()
		clearAmericanToadBoard(at)
		fillAmericanToadColumns(at)
		at.baseRank = 5
		at.reserve = []*Card{NewCard(CardDesignSpade, 5, true)}

		require.NoError(t, at.MoveReserveToFoundation())
		assert.Len(t, at.GetFoundation()[0], 1)
		assert.Empty(t, at.GetReserve())
	})

	t.Run("to a column", func(t *testing.T) {
		at := newTestAmericanToad()
		clearAmericanToadBoard(at)
		fillAmericanToadColumns(at)
		at.tableau[1] = atCol(NewCard(CardDesignSpade, 7, true))
		at.reserve = []*Card{NewCard(CardDesignSpade, 6, true)}

		require.NoError(t, at.MoveReserveToTableau(1))
		assert.Len(t, at.GetTableau()[1], 2)
		assert.Empty(t, at.GetReserve())
	})

	t.Run("rejects an empty reserve, an illegal card and a bad column", func(t *testing.T) {
		at := newTestAmericanToad()
		clearAmericanToadBoard(at)
		fillAmericanToadColumns(at)

		assert.Error(t, at.MoveReserveToFoundation())
		assert.Error(t, at.MoveReserveToTableau(0))

		at.reserve = []*Card{NewCard(CardDesignHeart, 4, true)}
		assert.Error(t, at.MoveReserveToFoundation())
		assert.Error(t, at.MoveReserveToTableau(0))
		assert.Error(t, at.MoveReserveToTableau(-1))
		assert.Error(t, at.MoveReserveToTableau(AmericanToadTableauCnt))
	})

	t.Run("errors once the game has ended", func(t *testing.T) {
		at := newTestAmericanToad()
		at.GiveUp()
		assert.Error(t, at.MoveReserveToFoundation())
		assert.Error(t, at.MoveReserveToTableau(0))
	})
}

func TestAmericanToad_MoveWaste(t *testing.T) {
	t.Run("to a foundation", func(t *testing.T) {
		at := newTestAmericanToad()
		clearAmericanToadBoard(at)
		fillAmericanToadColumns(at)
		at.baseRank = 5
		at.waste = []*Card{NewCard(CardDesignClover, 5, true)}

		require.NoError(t, at.MoveWasteToFoundation())
		assert.Len(t, at.GetFoundation()[1], 1)
		assert.Empty(t, at.GetWaste())
	})

	t.Run("rejects an empty waste, an illegal card and a bad column", func(t *testing.T) {
		at := newTestAmericanToad()
		clearAmericanToadBoard(at)
		fillAmericanToadColumns(at)

		assert.Error(t, at.MoveWasteToFoundation())
		assert.Error(t, at.MoveWasteToTableau(0))

		at.waste = []*Card{NewCard(CardDesignHeart, 4, true)}
		assert.Error(t, at.MoveWasteToFoundation())
		assert.Error(t, at.MoveWasteToTableau(0))
		assert.Error(t, at.MoveWasteToTableau(-1))
		assert.Error(t, at.MoveWasteToTableau(AmericanToadTableauCnt))
	})

	t.Run("errors once the game has ended", func(t *testing.T) {
		at := newTestAmericanToad()
		at.GiveUp()
		assert.Error(t, at.MoveWasteToFoundation())
		assert.Error(t, at.MoveWasteToTableau(0))
	})
}

func TestAmericanToad_MoveTableauToFoundation(t *testing.T) {
	t.Run("moves the column top", func(t *testing.T) {
		at := newTestAmericanToad()
		clearAmericanToadBoard(at)
		fillAmericanToadColumns(at)
		at.baseRank = 5
		at.tableau[4] = atCol(NewCard(CardDesignSpade, 3, true), NewCard(CardDesignHeart, 5, true))

		require.NoError(t, at.MoveTableauToFoundation(4))
		assert.Len(t, at.GetFoundation()[2], 1)
		assert.Len(t, at.GetTableau()[4], 1)
	})

	t.Run("rejects an empty column, an illegal card and a bad index", func(t *testing.T) {
		at := newTestAmericanToad()
		clearAmericanToadBoard(at)
		fillAmericanToadColumns(at)
		at.tableau[0] = nil

		assert.Error(t, at.MoveTableauToFoundation(0))
		assert.Error(t, at.MoveTableauToFoundation(1))
		assert.Error(t, at.MoveTableauToFoundation(-1))
		assert.Error(t, at.MoveTableauToFoundation(AmericanToadTableauCnt))
	})

	t.Run("errors once the game has ended", func(t *testing.T) {
		at := newTestAmericanToad()
		at.GiveUp()
		assert.Error(t, at.MoveTableauToFoundation(0))
	})
}

func TestAmericanToad_MoveTableauToTableau(t *testing.T) {
	t.Run("carries a same-suit run as a unit", func(t *testing.T) {
		at := newTestAmericanToad()
		clearAmericanToadBoard(at)
		fillAmericanToadColumns(at)
		at.tableau[0] = atCol(
			NewCard(CardDesignHeart, 2, true),
			NewCard(CardDesignSpade, 6, true),
			NewCard(CardDesignSpade, 5, true),
		)
		at.tableau[1] = atCol(NewCard(CardDesignSpade, 7, true))

		require.NoError(t, at.MoveTableauToTableau(0, 1, 1))
		assert.Len(t, at.GetTableau()[0], 1)
		assert.Len(t, at.GetTableau()[1], 3)
	})

	t.Run("defaults to the top card", func(t *testing.T) {
		at := newTestAmericanToad()
		clearAmericanToadBoard(at)
		fillAmericanToadColumns(at)
		at.tableau[0] = atCol(NewCard(CardDesignHeart, 2, true), NewCard(CardDesignSpade, 6, true))
		at.tableau[1] = atCol(NewCard(CardDesignSpade, 7, true))

		require.NoError(t, at.MoveTableauToTableau(0, -1, 1))
		assert.Len(t, at.GetTableau()[0], 1)
		assert.Len(t, at.GetTableau()[1], 2)
	})

	t.Run("refuses a group that is not a run", func(t *testing.T) {
		at := newTestAmericanToad()
		clearAmericanToadBoard(at)
		fillAmericanToadColumns(at)
		at.tableau[0] = atCol(NewCard(CardDesignSpade, 6, true), NewCard(CardDesignHeart, 2, true))
		at.tableau[1] = atCol(NewCard(CardDesignSpade, 7, true))

		assert.Error(t, at.MoveTableauToTableau(0, 0, 1))
	})

	t.Run("rejects bad columns and indices", func(t *testing.T) {
		at := newTestAmericanToad()
		clearAmericanToadBoard(at)
		fillAmericanToadColumns(at)
		at.tableau[0] = atCol(NewCard(CardDesignSpade, 6, true))
		at.tableau[1] = atCol(NewCard(CardDesignSpade, 7, true))
		at.tableau[2] = nil
		at.reserve = nil

		assert.Error(t, at.MoveTableauToTableau(-1, -1, 1))
		assert.Error(t, at.MoveTableauToTableau(0, -1, AmericanToadTableauCnt))
		assert.Error(t, at.MoveTableauToTableau(0, -1, 0), "same column")
		assert.Error(t, at.MoveTableauToTableau(2, -1, 1), "empty source")
		assert.Error(t, at.MoveTableauToTableau(0, 5, 1), "index past the end")
		assert.Error(t, at.MoveTableauToTableau(1, -1, 0), "rank does not fit")
	})

	t.Run("errors once the game has ended", func(t *testing.T) {
		at := newTestAmericanToad()
		at.GiveUp()
		assert.Error(t, at.MoveTableauToTableau(0, -1, 1))
	})
}

// isRun guards every card it walks, because a corrupt KV snapshot can put a nil
// where a card should be and a run check must not panic on it.
func TestAmericanToad_IsRunGuards(t *testing.T) {
	at := newTestAmericanToad()
	clearAmericanToadBoard(at)
	pile := atCol(NewCard(CardDesignSpade, 6, true), NewCard(CardDesignSpade, 5, true))

	assert.True(t, at.isRun(pile, 0))
	assert.True(t, at.isRun(pile, 1), "a single top card is trivially a run")
	assert.False(t, at.isRun(pile, -1))
	assert.False(t, at.isRun(pile, 2))
	assert.False(t, at.isRun([]*AmericanToadTableauCard{nil}, 0))
	assert.False(t, at.isRun([]*AmericanToadTableauCard{{Card: nil, FaceUp: true}}, 0))
	assert.False(t, at.isRun([]*AmericanToadTableauCard{
		{Card: NewCard(CardDesignSpade, 6, true), FaceUp: true},
		{Card: nil, FaceUp: true},
	}, 0))
	at.tableau[0] = []*AmericanToadTableauCard{{Card: nil, FaceUp: true}}
	assert.False(t, at.canPlaceOnTableau(NewCard(CardDesignSpade, 5, true), 0),
		"a column whose top card is nil accepts nothing")
}

// The suit lookup is a table scan, so it needs a defined answer for a design
// that is not in the table at all.
func TestAmericanToad_FoundationForSuit(t *testing.T) {
	for i, design := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		assert.Equal(t, i, americanToadFoundationForSuit(design))
	}
	assert.Equal(t, 0, americanToadFoundationForSuit(-1), "an unknown suit falls back to the first pile")
}

func TestAmericanToad_GetHint(t *testing.T) {
	t.Run("prefers a foundation move", func(t *testing.T) {
		at := newTestAmericanToad()
		clearAmericanToadBoard(at)
		fillAmericanToadColumns(at)
		at.baseRank = 5
		at.reserve = []*Card{NewCard(CardDesignSpade, 5, true)}

		h := at.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "reserve", h.FromZone)
		assert.Equal(t, "foundation", h.ToZone)
	})

	t.Run("falls back to a tableau move", func(t *testing.T) {
		at := newTestAmericanToad()
		clearAmericanToadBoard(at)
		fillAmericanToadColumns(at)
		at.baseRank = 5
		at.tableau[1] = atCol(NewCard(CardDesignSpade, 7, true))
		at.reserve = []*Card{NewCard(CardDesignSpade, 6, true)}

		h := at.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "reserve", h.FromZone)
		assert.Equal(t, "tableau", h.ToZone)
		assert.Equal(t, 1, h.ToIdx)
	})

	t.Run("offers a tableau run", func(t *testing.T) {
		at := newTestAmericanToad()
		clearAmericanToadBoard(at)
		fillAmericanToadColumns(at)
		at.baseRank = 5
		at.tableau[0] = atCol(NewCard(CardDesignHeart, 2, true), NewCard(CardDesignSpade, 6, true))
		at.tableau[1] = atCol(NewCard(CardDesignSpade, 7, true))

		h := at.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "tableau", h.FromZone)
		assert.Equal(t, 0, h.FromIdx)
		assert.Equal(t, 1, h.CardIndex)
		assert.Equal(t, 1, h.ToIdx)
	})

	// An empty column is not a tableau destination, so it must not be offered.
	t.Run("never suggests moving into an empty column", func(t *testing.T) {
		at := newTestAmericanToad()
		clearAmericanToadBoard(at)
		fillAmericanToadColumns(at)
		at.tableau[0] = atCol(NewCard(CardDesignSpade, 6, true))
		at.tableau[1] = nil

		assert.Nil(t, at.GetHint())
	})

	t.Run("falls back to drawing", func(t *testing.T) {
		at := newTestAmericanToad()
		clearAmericanToadBoard(at)
		fillAmericanToadColumns(at)
		at.stock = []*Card{NewCard(CardDesignHeart, 9, true)}

		h := at.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "stock", h.FromZone)
	})

	// The redeal is a move too, so a board with a recyclable waste is not stuck.
	t.Run("counts an available redeal as a move", func(t *testing.T) {
		at := newTestAmericanToad()
		clearAmericanToadBoard(at)
		fillAmericanToadColumns(at)
		at.waste = []*Card{NewCard(CardDesignHeart, 9, true)}

		h := at.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "stock", h.FromZone)
	})

	t.Run("returns nil once the game has ended", func(t *testing.T) {
		at := newTestAmericanToad()
		at.GiveUp()
		assert.Nil(t, at.GetHint())
	})
}

func TestAmericanToad_Stalemate(t *testing.T) {
	t.Run("a dead board is a stalemate", func(t *testing.T) {
		at := newTestAmericanToad()
		clearAmericanToadBoard(at)
		fillAmericanToadColumns(at)
		at.baseRank = 5
		at.passesUsed = AmericanToadMaxPasses - 1
		at.checkStalemate()
		assert.True(t, at.IsStalemate())
	})

	t.Run("a board with one legal move is not", func(t *testing.T) {
		at := newTestAmericanToad()
		clearAmericanToadBoard(at)
		fillAmericanToadColumns(at)
		at.baseRank = 9
		at.checkStalemate()
		assert.False(t, at.IsStalemate(), "every 9 can open its foundation")
	})
}

func TestAmericanToad_AutoComplete(t *testing.T) {
	t.Run("sends every reachable card to the foundations", func(t *testing.T) {
		at := newTestAmericanToad()
		clearAmericanToadBoard(at)
		fillAmericanToadColumns(at)
		at.baseRank = 5
		at.tableau[0] = atCol(NewCard(CardDesignSpade, 5, true))
		at.waste = []*Card{NewCard(CardDesignClover, 5, true)}
		at.reserve = []*Card{NewCard(CardDesignHeart, 5, true)}

		require.NoError(t, at.AutoComplete())
		assert.Len(t, at.GetFoundation()[0], 1)
		assert.Len(t, at.GetFoundation()[1], 1)
		assert.Len(t, at.GetFoundation()[2], 1)
	})

	// Auto-complete must never rearrange the tableau: that is a strategic
	// judgement, not a mechanical tidy-up.
	t.Run("never performs a tableau move", func(t *testing.T) {
		at := newTestAmericanToad()
		clearAmericanToadBoard(at)
		fillAmericanToadColumns(at)
		at.baseRank = 5
		at.tableau[0] = atCol(NewCard(CardDesignSpade, 6, true))
		at.tableau[1] = atCol(NewCard(CardDesignSpade, 7, true))

		assert.Error(t, at.AutoComplete())
		assert.Len(t, at.GetTableau()[0], 1)
		assert.Len(t, at.GetTableau()[1], 1)
	})

	t.Run("errors once the game has ended", func(t *testing.T) {
		at := newTestAmericanToad()
		at.GiveUp()
		assert.Error(t, at.AutoComplete())
	})
}

func TestAmericanToad_GameClear(t *testing.T) {
	at := newTestAmericanToad()
	clearAmericanToadBoard(at)
	fillAmericanToadColumns(at)
	at.baseRank = 1
	for i := range AmericanToadFoundationCnt {
		at.foundation[i] = nil
		for v := 1; v <= CardValueMax; v++ {
			at.foundation[i] = append(at.foundation[i], NewCard(americanToadSuitOrder[i], v, true))
		}
	}
	// Strip one card back off so the final move completes the game.
	at.foundation[0] = at.foundation[0][:CardValueMax-1]
	at.tableau[0] = atCol(NewCard(CardDesignSpade, CardValueMax, true))

	require.NoError(t, at.MoveTableauToFoundation(0))
	assert.Equal(t, AmericanToadPhaseGameClear, at.GetPhase())
	assert.True(t, at.GetGameEndFlag())
}

func TestAmericanToad_GiveUp(t *testing.T) {
	at := newTestAmericanToad()
	at.GiveUp()
	assert.Equal(t, AmericanToadPhaseGameOver, at.GetPhase())
	require.NotEmpty(t, at.GetActionLog())

	before := len(at.GetActionLog())
	at.GiveUp()
	assert.Len(t, at.GetActionLog(), before)
}

func TestAmericanToad_Undo(t *testing.T) {
	t.Run("restores the previous position", func(t *testing.T) {
		at := newTestAmericanToad()
		clearAmericanToadBoard(at)
		fillAmericanToadColumns(at)
		at.baseRank = 5
		at.reserve = []*Card{NewCard(CardDesignSpade, 5, true)}

		require.NoError(t, at.MoveReserveToFoundation())
		require.True(t, at.CanUndo())
		require.NoError(t, at.Undo())

		assert.Len(t, at.GetReserve(), 1)
		assert.Empty(t, at.GetFoundation()[0])
		assert.Equal(t, 0, at.GetMoveCount())
	})

	// The redeal counter is part of the position: undoing a recycle has to give
	// the pass back, or the player silently loses their only redeal.
	t.Run("restores the pass count", func(t *testing.T) {
		at := newTestAmericanToad()
		clearAmericanToadBoard(at)
		fillAmericanToadColumns(at)
		at.waste = []*Card{NewCard(CardDesignSpade, 4, true)}

		require.NoError(t, at.Draw())
		require.Equal(t, 1, at.GetPassesUsed())
		require.NoError(t, at.Undo())
		assert.Equal(t, 0, at.GetPassesUsed())
		assert.Len(t, at.GetWaste(), 1)
	})

	t.Run("errors with no history", func(t *testing.T) {
		at := newTestAmericanToad()
		at.history = nil
		assert.False(t, at.CanUndo())
		assert.Error(t, at.Undo())
	})
}

func TestAmericanToad_UndoN(t *testing.T) {
	at := newTestAmericanToad()
	clearAmericanToadBoard(at)
	fillAmericanToadColumns(at)
	at.baseRank = 5
	at.reserve = []*Card{NewCard(CardDesignHeart, 5, true), NewCard(CardDesignSpade, 5, true)}

	require.NoError(t, at.MoveReserveToFoundation())
	require.NoError(t, at.MoveReserveToFoundation())
	require.NoError(t, at.UndoN(2))
	assert.Len(t, at.GetReserve(), 2)
	assert.Equal(t, 0, at.GetMoveCount())

	assert.Error(t, at.UndoN(0))
	assert.Error(t, at.UndoN(-1))
	assert.Error(t, at.UndoN(99))
}

func TestAmericanToad_UndoToEscape(t *testing.T) {
	t.Run("zero when not stuck", func(t *testing.T) {
		at := newTestAmericanToad()
		assert.Equal(t, 0, at.UndoToEscape())
	})

	t.Run("counts back to the last live position", func(t *testing.T) {
		at := newTestAmericanToad()
		clearAmericanToadBoard(at)
		fillAmericanToadColumns(at)
		at.baseRank = 5
		at.passesUsed = AmericanToadMaxPasses - 1
		at.reserve = []*Card{NewCard(CardDesignSpade, 5, true)}
		at.checkStalemate()
		require.False(t, at.IsStalemate())

		require.NoError(t, at.MoveReserveToFoundation())
		require.True(t, at.IsStalemate())
		assert.Equal(t, 1, at.UndoToEscape())
	})

	t.Run("-1 when every recorded position was already stuck", func(t *testing.T) {
		at := newTestAmericanToad()
		at.isStalemate = true
		at.history = []*americanToadSnapshot{{isStalemate: true}}
		assert.Equal(t, -1, at.UndoToEscape())
	})
}

func TestAmericanToad_ActionLog(t *testing.T) {
	at := newTestAmericanToad()
	clearAmericanToadBoard(at)
	fillAmericanToadColumns(at)
	at.baseRank = 5
	at.tableau[0] = atCol(NewCard(CardDesignSpade, 7, true))
	at.tableau[1] = atCol(NewCard(CardDesignSpade, 6, true))
	at.reserve = []*Card{NewCard(CardDesignSpade, 5, true)}
	at.waste = []*Card{NewCard(CardDesignClover, 5, true)}
	at.stock = []*Card{NewCard(CardDesignDiamond, 7, true)}

	require.NoError(t, at.MoveReserveToFoundation())
	require.NoError(t, at.MoveWasteToFoundation())
	require.NoError(t, at.MoveTableauToTableau(1, -1, 0))
	require.NoError(t, at.Draw())

	// The board is 0-indexed everywhere, so the log must be too -- a 1-based log
	// silently disagrees with the hint and the CLI.
	details := make([]string, 0, len(at.GetActionLog()))
	for _, e := range at.GetActionLog() {
		details = append(details, e.Detail)
	}
	assert.Equal(t, []string{
		"リザーブ→基礎札0",
		"捨て札→基礎札1",
		"タブロー列1[0]→タブロー列0",
		"山札から1枚めくった",
	}, details)
}

// The Cloudflare Worker is stateless per request and rebuilds the game from KV
// on every call, so an undo stack that does not round-trip means Undo, UndoN and
// the whole stalemate-escape flow silently never work in production. Asserting
// only on the restored field counts would not catch it -- the test has to undo.
func TestAmericanToad_UndoSurvivesAKVRoundTrip(t *testing.T) {
	at := newTestAmericanToad()
	clearAmericanToadBoard(at)
	fillAmericanToadColumns(at)
	at.baseRank = 5
	at.reserve = []*Card{NewCard(CardDesignHeart, 5, true), NewCard(CardDesignSpade, 5, true)}

	require.NoError(t, at.MoveReserveToFoundation())
	require.NoError(t, at.MoveReserveToFoundation())
	require.True(t, at.CanUndo())

	data, err := json.Marshal(at)
	require.NoError(t, err)

	restored := NewDefaultAmericanToad()
	require.NoError(t, json.Unmarshal(data, restored))
	require.True(t, restored.CanUndo(), "the undo stack must survive KV")

	require.NoError(t, restored.UndoN(2))
	assert.Len(t, restored.GetReserve(), 2, "both cards came back")
	assert.Empty(t, restored.GetFoundation()[0])
	assert.Empty(t, restored.GetFoundation()[2])
	assert.Equal(t, 0, restored.GetMoveCount())
}

// A snapshot restored from KV must carry the whole position, not just its shape:
// a blank snapshot would let Undo wipe the board instead of rewinding it.
func TestAmericanToad_SnapshotRoundTripKeepsItsContents(t *testing.T) {
	at := newTestAmericanToad()
	require.NoError(t, at.Draw())
	require.NoError(t, at.Draw())
	before := at.history[0]

	data, err := json.Marshal(at)
	require.NoError(t, err)
	restored := NewDefaultAmericanToad()
	require.NoError(t, json.Unmarshal(data, restored))

	require.Len(t, restored.history, 2)
	after := restored.history[0]
	assert.Len(t, after.reserve, len(before.reserve))
	assert.Len(t, after.stock, len(before.stock))
	assert.Len(t, after.waste, len(before.waste))
	assert.Equal(t, before.baseRank, after.baseRank)
	assert.Equal(t, before.passesUsed, after.passesUsed)
	assert.Equal(t, before.moveCount, after.moveCount)
	for i := range AmericanToadTableauCnt {
		assert.Len(t, after.tableau[i], len(before.tableau[i]), "column %d", i)
	}
}

// A hostile or corrupt KV payload must not be able to make the game allocate
// without bound, and a snapshot is just as reachable as the top-level state.
func TestAmericanToad_HistoryRespectsMaxSliceLen(t *testing.T) {
	t.Run("too many snapshots", func(t *testing.T) {
		history := make([]*americanToadSnapshot, americanToadMaxSliceLen+1)
		for i := range history {
			history[i] = &americanToadSnapshot{}
		}
		data, err := json.Marshal(&americanToadJSON{BaseRank: 5, History: history})
		require.NoError(t, err)
		assert.Error(t, NewDefaultAmericanToad().UnmarshalJSON(data))
	})

	t.Run("too long an action log", func(t *testing.T) {
		log := make([]*ActionLogEntry, americanToadMaxSliceLen+1)
		for i := range log {
			log[i] = &ActionLogEntry{}
		}
		data, err := json.Marshal(&americanToadJSON{BaseRank: 5, ActionLog: log})
		require.NoError(t, err)
		assert.Error(t, NewDefaultAmericanToad().UnmarshalJSON(data))
	})

	t.Run("an oversized pile inside a snapshot", func(t *testing.T) {
		huge := make([]*Card, americanToadMaxSliceLen+1)
		for i := range huge {
			huge[i] = NewCard(CardDesignSpade, 1, true)
		}
		data, err := json.Marshal(&americanToadSnapshotJSON{Stock: huge})
		require.NoError(t, err)
		assert.Error(t, new(americanToadSnapshot).UnmarshalJSON(data))
	})

	t.Run("malformed snapshot json", func(t *testing.T) {
		assert.Error(t, new(americanToadSnapshot).UnmarshalJSON([]byte("not json")))
	})
}

func TestAmericanToad_JSONRoundTrip(t *testing.T) {
	at := newTestAmericanToad()
	require.NoError(t, at.Draw())

	data, err := json.Marshal(at)
	require.NoError(t, err)

	restored := NewDefaultAmericanToad()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, at.GetPhase(), restored.GetPhase())
	assert.Equal(t, at.GetMoveCount(), restored.GetMoveCount())
	assert.Equal(t, at.GetStockCount(), restored.GetStockCount())
	assert.Equal(t, at.GetBaseRank(), restored.GetBaseRank())
	assert.Equal(t, at.GetPassesUsed(), restored.GetPassesUsed())
	assert.Len(t, restored.GetReserve(), len(at.GetReserve()))
}

// KV can hold anything an earlier version wrote, so a corrupt snapshot must be
// refused rather than started.
func TestAmericanToad_UnmarshalJSONValidation(t *testing.T) {
	huge := make([]*Card, AmericanToadTotalCards+1)
	for i := range huge {
		huge[i] = NewCard(CardDesignSpade, 1, true)
	}
	tooManyFoundation := make([]*Card, AmericanToadFoundationTarget+1)
	for i := range tooManyFoundation {
		tooManyFoundation[i] = NewCard(CardDesignSpade, 1, true)
	}
	tooLongColumn := make([]*AmericanToadTableauCard, AmericanToadTotalCards+1)
	for i := range tooLongColumn {
		tooLongColumn[i] = &AmericanToadTableauCard{Card: NewCard(CardDesignSpade, 1, true), FaceUp: true}
	}

	for _, tc := range []struct {
		name string
		j    americanToadJSON
	}{
		{"phase below range", americanToadJSON{Phase: -1}},
		{"phase above range", americanToadJSON{Phase: AmericanToadPhaseGameOver + 1}},
		{"negative move count", americanToadJSON{MoveCount: -1}},
		{"base rank out of range", americanToadJSON{BaseRank: CardValueMax + 1}},
		{"pass count out of range", americanToadJSON{PassesUsed: AmericanToadMaxPasses}},
		{"negative pass count", americanToadJSON{PassesUsed: -1}},
		{"reserve overflows", americanToadJSON{Reserve: huge}},
		{"stock overflows", americanToadJSON{Stock: huge}},
		{"waste overflows", americanToadJSON{Waste: huge}},
		{"foundation overflows", americanToadJSON{Foundation: [AmericanToadFoundationCnt][]*Card{tooManyFoundation}}},
		{"tableau overflows", americanToadJSON{Tableau: [AmericanToadTableauCnt][]*AmericanToadTableauCard{tooLongColumn}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(&tc.j)
			require.NoError(t, err)
			assert.Error(t, NewDefaultAmericanToad().UnmarshalJSON(data))
		})
	}

	t.Run("malformed json", func(t *testing.T) {
		assert.Error(t, NewDefaultAmericanToad().UnmarshalJSON([]byte("not json")))
	})

	t.Run("a valid snapshot is accepted", func(t *testing.T) {
		data, err := json.Marshal(&americanToadJSON{Phase: AmericanToadPhasePlaying, MoveCount: 3, BaseRank: 5})
		require.NoError(t, err)
		at := NewDefaultAmericanToad()
		require.NoError(t, at.UnmarshalJSON(data))
		assert.Equal(t, 3, at.GetMoveCount())
		assert.Equal(t, 5, at.GetBaseRank())
	})
}

// Drive a full game to make sure no path panics and the invariants hold
// throughout. The deal is random, so this is a fuzz-ish smoke test rather than
// an assertion about any particular position.
func TestAmericanToad_FullGameDrive(t *testing.T) {
	for range 20 {
		at := newTestAmericanToad()
		for range 800 {
			if at.GetGameEndFlag() {
				break
			}
			h := at.GetHint()
			if h == nil {
				break
			}
			var err error
			switch {
			case h.FromZone == "stock":
				err = at.Draw()
			case h.FromZone == "reserve" && h.ToZone == "foundation":
				err = at.MoveReserveToFoundation()
			case h.FromZone == "reserve":
				err = at.MoveReserveToTableau(h.ToIdx)
			case h.FromZone == "waste" && h.ToZone == "foundation":
				err = at.MoveWasteToFoundation()
			case h.FromZone == "waste":
				err = at.MoveWasteToTableau(h.ToIdx)
			case h.ToZone == "foundation":
				err = at.MoveTableauToFoundation(h.FromIdx)
			default:
				err = at.MoveTableauToTableau(h.FromIdx, h.CardIndex, h.ToIdx)
			}
			require.NoError(t, err)

			for i := range AmericanToadFoundationCnt {
				assert.LessOrEqual(t, len(at.GetFoundation()[i]), AmericanToadFoundationTarget)
			}
			assert.LessOrEqual(t, at.GetPassesUsed(), AmericanToadMaxPasses-1)
		}
	}
}
