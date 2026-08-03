//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestTerrace() *Terrace {
	t := NewDefaultTerrace()
	t.Reset()
	return t
}

// clearTerraceBoard wipes the dealt layout so a test can state exactly the
// position it cares about. Never assert on a freshly Reset board -- the deal is
// shuffled, so any such assertion is a hidden flake.
func clearTerraceBoard(tr *Terrace) {
	tr.reserve = nil
	tr.stock = nil
	tr.waste = nil
	tr.isStalemate = false
	tr.baseRank = 0
	for i := range TerraceFoundationCnt {
		tr.foundation[i] = nil
	}
	for i := range TerraceTableauCnt {
		tr.tableau[i] = nil
	}
}

// fillTerracePiles puts one dead card in every pile so no gap exists. Gaps are
// auto-refilled the instant they appear, so a board full of holes would test a
// position the game never actually reaches.
func fillTerracePiles(tr *Terrace) {
	for i := range TerraceTableauCnt {
		tr.tableau[i] = []*Card{NewCard(CardDesignSpade, 9, true)}
	}
}

func TestNewTerrace(t *testing.T) {
	assert.NotNil(t, NewTerrace(NewTrumpCardsWithDecks(2, 0)))
	assert.NotNil(t, NewDefaultTerrace())
}

func TestTerrace_Reset(t *testing.T) {
	tr := newTestTerrace()

	// 11 terrace + 9 piles of one + 84 stock = 104.
	assert.Len(t, tr.GetReserve(), TerraceReserveSize)
	for i, pile := range tr.GetTableau() {
		assert.Len(t, pile, 1, "pile %d", i)
	}
	assert.Equal(t, TerraceTotalCards-TerraceReserveSize-TerraceTableauCnt, tr.GetStockCount())
	assert.Empty(t, tr.GetWaste())

	for i, pile := range tr.GetFoundation() {
		assert.Empty(t, pile, "foundation %d", i)
	}
	assert.True(t, tr.IsAwaitingBaseRank())
	assert.Equal(t, 0, tr.GetBaseRank())

	assert.Equal(t, TerracePhasePlaying, tr.GetPhase())
	assert.Equal(t, 0, tr.GetMoveCount())
	assert.True(t, tr.AllFaceUp())
	assert.False(t, tr.GetGameEndFlag())
}

func TestTerrace_ResetDealsEveryCard(t *testing.T) {
	for range 20 {
		tr := newTestTerrace()
		total := len(tr.GetReserve()) + tr.GetStockCount() + len(tr.GetWaste())
		for _, pile := range tr.GetTableau() {
			total += len(pile)
		}
		for _, pile := range tr.GetFoundation() {
			total += len(pile)
		}
		assert.Equal(t, TerraceTotalCards, total)
	}
}

func TestTerrace_ResetIsRepeatable(t *testing.T) {
	tr := newTestTerrace()
	require.NoError(t, tr.Draw())
	tr.Reset()
	assert.Equal(t, 0, tr.GetMoveCount())
	assert.Empty(t, tr.GetWaste())
	assert.Empty(t, tr.GetActionLog())
	assert.False(t, tr.CanUndo())
}

// The signature rule: foundations build up in ALTERNATING COLOUR, not by suit.
// #4381 says "same suit ascending", which would be a different game.
func TestTerrace_FoundationBuildsInAlternatingColour(t *testing.T) {
	tr := newTestTerrace()
	clearTerraceBoard(tr)
	tr.baseRank = 5
	tr.foundation[0] = []*Card{NewCard(CardDesignSpade, 5, true)}

	assert.True(t, tr.canPlaceOnFoundation(NewCard(CardDesignHeart, 6, true), 0),
		"a red six goes on a black five")
	assert.True(t, tr.canPlaceOnFoundation(NewCard(CardDesignDiamond, 6, true), 0))
	assert.False(t, tr.canPlaceOnFoundation(NewCard(CardDesignClover, 6, true), 0),
		"same colour is refused even though the rank fits")
	assert.False(t, tr.canPlaceOnFoundation(NewCard(CardDesignSpade, 6, true), 0),
		"and the same suit is certainly refused")
	assert.False(t, tr.canPlaceOnFoundation(NewCard(CardDesignHeart, 7, true), 0))
	assert.False(t, tr.canPlaceOnFoundation(nil, 0))

	t.Run("wraps King round to Ace", func(t *testing.T) {
		tr.foundation[1] = []*Card{NewCard(CardDesignSpade, CardValueMax, true)}
		assert.True(t, tr.canPlaceOnFoundation(NewCard(CardDesignHeart, 1, true), 1))
		assert.False(t, tr.canPlaceOnFoundation(NewCard(CardDesignSpade, 1, true), 1))
	})

	t.Run("stops at thirteen cards", func(t *testing.T) {
		tr.foundation[2] = nil
		v := 5
		for i := range TerraceFoundationTarget {
			design := CardDesignSpade
			if i%2 == 1 {
				design = CardDesignHeart
			}
			tr.foundation[2] = append(tr.foundation[2], NewCard(design, v, true))
			v = terraceNextRank(v)
		}
		require.Len(t, tr.foundation[2], TerraceFoundationTarget)
		for r := 1; r <= CardValueMax; r++ {
			assert.False(t, tr.canPlaceOnFoundation(NewCard(CardDesignHeart, r, true), 2), "rank %d", r)
		}
	})
}

// The base rank is not dealt: the first card the player sends to a foundation
// sets it, and after that only that rank opens a new foundation.
func TestTerrace_BaseRankIsChosenByTheFirstFoundationCard(t *testing.T) {
	tr := newTestTerrace()
	clearTerraceBoard(tr)
	fillTerracePiles(tr)
	tr.tableau[0] = []*Card{NewCard(CardDesignHeart, 7, true)}

	require.True(t, tr.IsAwaitingBaseRank())
	// Any card may open the first foundation while the rank is unset.
	assert.True(t, tr.canPlaceOnFoundation(NewCard(CardDesignClover, 3, true), 0))

	require.NoError(t, tr.MoveTableauToFoundation(0))
	assert.Equal(t, 7, tr.GetBaseRank())
	assert.False(t, tr.IsAwaitingBaseRank())

	// Now only a 7 can open another foundation.
	assert.True(t, tr.canPlaceOnFoundation(NewCard(CardDesignSpade, 7, true), 1))
	assert.False(t, tr.canPlaceOnFoundation(NewCard(CardDesignSpade, 3, true), 1))
}

// The tableau builds down in alternating colour, one card at a time, wrapping
// Ace round to King.
func TestTerrace_TableauBuildRules(t *testing.T) {
	tr := newTestTerrace()
	clearTerraceBoard(tr)
	fillTerracePiles(tr)
	tr.tableau[0] = []*Card{NewCard(CardDesignSpade, 7, true)}

	assert.True(t, tr.canPlaceOnTableau(NewCard(CardDesignHeart, 6, true), 0))
	assert.False(t, tr.canPlaceOnTableau(NewCard(CardDesignClover, 6, true), 0), "same colour")
	assert.False(t, tr.canPlaceOnTableau(NewCard(CardDesignHeart, 5, true), 0))
	assert.False(t, tr.canPlaceOnTableau(nil, 0))

	tr.tableau[1] = []*Card{NewCard(CardDesignHeart, 1, true)}
	assert.True(t, tr.canPlaceOnTableau(NewCard(CardDesignSpade, CardValueMax, true), 1),
		"a black King goes under a red Ace")
}

// The terrace feeds the foundations and nothing else. #4381 implies its top card
// may also move to the tableau, which would defuse the whole game.
func TestTerrace_ReserveOnlyFeedsFoundations(t *testing.T) {
	tr := newTestTerrace()
	clearTerraceBoard(tr)
	fillTerracePiles(tr)
	tr.tableau[0] = []*Card{NewCard(CardDesignSpade, 7, true)}
	tr.reserve = []*Card{NewCard(CardDesignHeart, 6, true)}

	// A red six would fit the black seven, but there is no move that gets it there.
	assert.True(t, tr.canPlaceOnTableau(tr.reserveTop(), 0))
	h := tr.GetHint()
	if h != nil {
		assert.NotEqual(t, "tableau", h.ToZone, "the reserve never suggests a tableau move")
	}
}

func TestTerrace_MoveReserveToFoundation(t *testing.T) {
	t.Run("sends the terrace top up and sets the base rank", func(t *testing.T) {
		tr := newTestTerrace()
		clearTerraceBoard(tr)
		fillTerracePiles(tr)
		tr.reserve = []*Card{NewCard(CardDesignHeart, 4, true), NewCard(CardDesignSpade, 6, true)}

		require.NoError(t, tr.MoveReserveToFoundation())
		assert.Len(t, tr.GetFoundation()[0], 1)
		assert.Len(t, tr.GetReserve(), 1)
		assert.Equal(t, 6, tr.GetBaseRank())
	})

	t.Run("rejects an empty terrace and an illegal card", func(t *testing.T) {
		tr := newTestTerrace()
		clearTerraceBoard(tr)
		fillTerracePiles(tr)

		assert.Error(t, tr.MoveReserveToFoundation())

		tr.baseRank = 5
		tr.reserve = []*Card{NewCard(CardDesignHeart, 9, true)}
		assert.Error(t, tr.MoveReserveToFoundation())
	})

	t.Run("errors once the game has ended", func(t *testing.T) {
		tr := newTestTerrace()
		tr.GiveUp()
		assert.Error(t, tr.MoveReserveToFoundation())
	})
}

// The terrace is never refilled -- only the tableau is.
func TestTerrace_EmptyPileRefillsAutomatically(t *testing.T) {
	t.Run("from the waste first", func(t *testing.T) {
		tr := newTestTerrace()
		clearTerraceBoard(tr)
		fillTerracePiles(tr)
		tr.baseRank = 5
		tr.tableau[3] = []*Card{NewCard(CardDesignSpade, 5, true)}
		refill := NewCard(CardDesignClover, 2, true)
		tr.waste = []*Card{NewCard(CardDesignHeart, 8, true), refill}

		require.NoError(t, tr.MoveTableauToFoundation(3))
		assert.Len(t, tr.GetTableau()[3], 1, "the gap closed immediately")
		assert.Equal(t, refill, tr.GetTableau()[3][0])
		assert.Len(t, tr.GetWaste(), 1)
	})

	t.Run("then from the stock", func(t *testing.T) {
		tr := newTestTerrace()
		clearTerraceBoard(tr)
		fillTerracePiles(tr)
		tr.baseRank = 5
		tr.tableau[3] = []*Card{NewCard(CardDesignSpade, 5, true)}
		top := NewCard(CardDesignDiamond, 2, true)
		tr.stock = []*Card{top, NewCard(CardDesignSpade, 3, true)}

		require.NoError(t, tr.MoveTableauToFoundation(3))
		assert.Equal(t, top, tr.GetTableau()[3][0])
		assert.Equal(t, 1, tr.GetStockCount())
	})

	t.Run("the terrace is never refilled", func(t *testing.T) {
		tr := newTestTerrace()
		clearTerraceBoard(tr)
		fillTerracePiles(tr)
		tr.reserve = []*Card{NewCard(CardDesignSpade, 5, true)}
		tr.stock = []*Card{NewCard(CardDesignHeart, 2, true)}

		require.NoError(t, tr.MoveReserveToFoundation())
		assert.Empty(t, tr.GetReserve(), "no card takes its place")
	})

	t.Run("a gap persists when nothing is left to fill it", func(t *testing.T) {
		tr := newTestTerrace()
		clearTerraceBoard(tr)
		fillTerracePiles(tr)
		tr.baseRank = 5
		tr.tableau[3] = []*Card{NewCard(CardDesignSpade, 5, true)}

		require.NoError(t, tr.MoveTableauToFoundation(3))
		assert.Empty(t, tr.GetTableau()[3])
	})
}

func TestTerrace_Draw(t *testing.T) {
	t.Run("moves one card to the waste", func(t *testing.T) {
		tr := newTestTerrace()
		before := tr.GetStockCount()
		require.NoError(t, tr.Draw())
		assert.Equal(t, before-1, tr.GetStockCount())
		assert.Len(t, tr.GetWaste(), 1)
	})

	// One pass only -- there is no redeal.
	t.Run("refuses once the stock is out", func(t *testing.T) {
		tr := newTestTerrace()
		clearTerraceBoard(tr)
		fillTerracePiles(tr)
		tr.waste = []*Card{NewCard(CardDesignSpade, 4, true)}
		assert.Error(t, tr.Draw())
	})

	t.Run("errors once the game has ended", func(t *testing.T) {
		tr := newTestTerrace()
		tr.GiveUp()
		assert.Error(t, tr.Draw())
	})
}

func TestTerrace_MoveWaste(t *testing.T) {
	t.Run("to a foundation", func(t *testing.T) {
		tr := newTestTerrace()
		clearTerraceBoard(tr)
		fillTerracePiles(tr)
		tr.baseRank = 5
		tr.waste = []*Card{NewCard(CardDesignClover, 5, true)}

		require.NoError(t, tr.MoveWasteToFoundation())
		assert.Len(t, tr.GetFoundation()[0], 1)
		assert.Empty(t, tr.GetWaste())
	})

	t.Run("to a pile", func(t *testing.T) {
		tr := newTestTerrace()
		clearTerraceBoard(tr)
		fillTerracePiles(tr)
		tr.tableau[1] = []*Card{NewCard(CardDesignSpade, 7, true)}
		tr.waste = []*Card{NewCard(CardDesignHeart, 6, true)}

		require.NoError(t, tr.MoveWasteToTableau(1))
		assert.Len(t, tr.GetTableau()[1], 2)
	})

	t.Run("rejects an empty waste, an illegal card and a bad index", func(t *testing.T) {
		tr := newTestTerrace()
		clearTerraceBoard(tr)
		fillTerracePiles(tr)
		tr.baseRank = 5

		assert.Error(t, tr.MoveWasteToFoundation())
		assert.Error(t, tr.MoveWasteToTableau(0))

		tr.waste = []*Card{NewCard(CardDesignHeart, 4, true)}
		assert.Error(t, tr.MoveWasteToFoundation())
		assert.Error(t, tr.MoveWasteToTableau(0))
		assert.Error(t, tr.MoveWasteToTableau(-1))
		assert.Error(t, tr.MoveWasteToTableau(TerraceTableauCnt))
	})

	t.Run("errors once the game has ended", func(t *testing.T) {
		tr := newTestTerrace()
		tr.GiveUp()
		assert.Error(t, tr.MoveWasteToFoundation())
		assert.Error(t, tr.MoveWasteToTableau(0))
	})
}

func TestTerrace_MoveTableau(t *testing.T) {
	t.Run("to a foundation", func(t *testing.T) {
		tr := newTestTerrace()
		clearTerraceBoard(tr)
		fillTerracePiles(tr)
		tr.baseRank = 5
		tr.tableau[4] = []*Card{NewCard(CardDesignHeart, 5, true)}

		require.NoError(t, tr.MoveTableauToFoundation(4))
		assert.Len(t, tr.GetFoundation()[0], 1)
	})

	t.Run("one card between piles", func(t *testing.T) {
		tr := newTestTerrace()
		clearTerraceBoard(tr)
		fillTerracePiles(tr)
		tr.tableau[0] = []*Card{NewCard(CardDesignClover, 3, true), NewCard(CardDesignHeart, 6, true)}
		tr.tableau[1] = []*Card{NewCard(CardDesignSpade, 7, true)}

		require.NoError(t, tr.MoveTableauToTableau(0, 1))
		assert.Len(t, tr.GetTableau()[0], 1, "only the top card moved")
		assert.Len(t, tr.GetTableau()[1], 2)
	})

	// An empty pile refills itself, so it is not a manual destination.
	t.Run("refuses an empty destination", func(t *testing.T) {
		tr := newTestTerrace()
		clearTerraceBoard(tr)
		fillTerracePiles(tr)
		tr.tableau[0] = []*Card{NewCard(CardDesignSpade, 7, true)}
		tr.tableau[1] = nil

		assert.Error(t, tr.MoveTableauToTableau(0, 1))
	})

	t.Run("rejects bad indices and illegal ranks", func(t *testing.T) {
		tr := newTestTerrace()
		clearTerraceBoard(tr)
		fillTerracePiles(tr)
		tr.baseRank = 5
		tr.tableau[0] = []*Card{NewCard(CardDesignSpade, 6, true)}
		tr.tableau[1] = []*Card{NewCard(CardDesignSpade, 7, true)}
		tr.tableau[2] = nil

		assert.Error(t, tr.MoveTableauToFoundation(-1))
		assert.Error(t, tr.MoveTableauToFoundation(TerraceTableauCnt))
		assert.Error(t, tr.MoveTableauToFoundation(2), "empty pile")
		assert.Error(t, tr.MoveTableauToFoundation(0), "rank does not fit")
		assert.Error(t, tr.MoveTableauToTableau(-1, 1))
		assert.Error(t, tr.MoveTableauToTableau(0, TerraceTableauCnt))
		assert.Error(t, tr.MoveTableauToTableau(0, 0), "same pile")
		assert.Error(t, tr.MoveTableauToTableau(2, 1), "empty source")
		assert.Error(t, tr.MoveTableauToTableau(1, 0), "same colour")
	})

	t.Run("errors once the game has ended", func(t *testing.T) {
		tr := newTestTerrace()
		tr.GiveUp()
		assert.Error(t, tr.MoveTableauToFoundation(0))
		assert.Error(t, tr.MoveTableauToTableau(0, 1))
	})
}

func TestTerrace_GetHint(t *testing.T) {
	// The terrace can only ever reach a foundation, so it is offered first --
	// leaving it stuck is what loses the game.
	t.Run("prefers the terrace to a foundation", func(t *testing.T) {
		tr := newTestTerrace()
		clearTerraceBoard(tr)
		fillTerracePiles(tr)
		tr.baseRank = 5
		tr.reserve = []*Card{NewCard(CardDesignSpade, 5, true)}
		tr.waste = []*Card{NewCard(CardDesignHeart, 5, true)}

		h := tr.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "reserve", h.FromZone)
		assert.Equal(t, "foundation", h.ToZone)
	})

	t.Run("then the waste, then the tableau", func(t *testing.T) {
		tr := newTestTerrace()
		clearTerraceBoard(tr)
		fillTerracePiles(tr)
		tr.baseRank = 5
		tr.waste = []*Card{NewCard(CardDesignHeart, 5, true)}

		h := tr.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "waste", h.FromZone)
		assert.Equal(t, "foundation", h.ToZone)
	})

	t.Run("offers a tableau move", func(t *testing.T) {
		tr := newTestTerrace()
		clearTerraceBoard(tr)
		fillTerracePiles(tr)
		tr.baseRank = 5
		tr.tableau[0] = []*Card{NewCard(CardDesignHeart, 6, true)}
		tr.tableau[1] = []*Card{NewCard(CardDesignSpade, 7, true)}

		h := tr.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "tableau", h.FromZone)
		assert.Equal(t, "tableau", h.ToZone)
	})

	t.Run("falls back to drawing", func(t *testing.T) {
		tr := newTestTerrace()
		clearTerraceBoard(tr)
		fillTerracePiles(tr)
		tr.baseRank = 5
		tr.stock = []*Card{NewCard(CardDesignHeart, 9, true)}

		h := tr.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "stock", h.FromZone)
	})

	t.Run("returns nil once the game has ended", func(t *testing.T) {
		tr := newTestTerrace()
		tr.GiveUp()
		assert.Nil(t, tr.GetHint())
	})
}

func TestTerrace_Stalemate(t *testing.T) {
	t.Run("a dead board is a stalemate", func(t *testing.T) {
		tr := newTestTerrace()
		clearTerraceBoard(tr)
		fillTerracePiles(tr)
		tr.baseRank = 5
		tr.checkStalemate()
		assert.True(t, tr.IsStalemate())
	})

	t.Run("a board with one legal move is not", func(t *testing.T) {
		tr := newTestTerrace()
		clearTerraceBoard(tr)
		fillTerracePiles(tr)
		tr.baseRank = 9
		tr.checkStalemate()
		assert.False(t, tr.IsStalemate(), "every pile top is a 9")
	})
}

func TestTerrace_AutoComplete(t *testing.T) {
	t.Run("sends every reachable card to the foundations", func(t *testing.T) {
		tr := newTestTerrace()
		clearTerraceBoard(tr)
		fillTerracePiles(tr)
		tr.baseRank = 5
		tr.reserve = []*Card{NewCard(CardDesignSpade, 5, true)}
		tr.tableau[0] = []*Card{NewCard(CardDesignHeart, 5, true)}

		require.NoError(t, tr.AutoComplete())
		assert.Empty(t, tr.GetReserve())
		assert.Len(t, tr.GetFoundation()[0], 1)
		assert.Len(t, tr.GetFoundation()[1], 1)
	})

	// Auto-complete must never rearrange the tableau.
	t.Run("never performs a tableau move", func(t *testing.T) {
		tr := newTestTerrace()
		clearTerraceBoard(tr)
		fillTerracePiles(tr)
		tr.baseRank = 5
		tr.tableau[0] = []*Card{NewCard(CardDesignHeart, 6, true)}
		tr.tableau[1] = []*Card{NewCard(CardDesignSpade, 7, true)}

		assert.Error(t, tr.AutoComplete())
		assert.Len(t, tr.GetTableau()[0], 1)
		assert.Len(t, tr.GetTableau()[1], 1)
	})

	t.Run("errors once the game has ended", func(t *testing.T) {
		tr := newTestTerrace()
		tr.GiveUp()
		assert.Error(t, tr.AutoComplete())
	})
}

func TestTerrace_GameClear(t *testing.T) {
	tr := newTestTerrace()
	clearTerraceBoard(tr)
	fillTerracePiles(tr)
	tr.baseRank = 1
	for i := range TerraceFoundationCnt {
		tr.foundation[i] = nil
		for v := 1; v <= CardValueMax; v++ {
			design := CardDesignSpade
			if v%2 == 0 {
				design = CardDesignHeart
			}
			tr.foundation[i] = append(tr.foundation[i], NewCard(design, v, true))
		}
	}
	// Strip one card back off so the final move completes the game.
	tr.foundation[0] = tr.foundation[0][:CardValueMax-1]
	// The stripped pile now ends on a red queen (v=12), so the King that
	// completes it has to be black -- the foundations alternate colour.
	tr.tableau[0] = []*Card{NewCard(CardDesignSpade, CardValueMax, true)}

	require.NoError(t, tr.MoveTableauToFoundation(0))
	assert.Equal(t, TerracePhaseGameClear, tr.GetPhase())
	assert.True(t, tr.GetGameEndFlag())
}

func TestTerrace_GiveUp(t *testing.T) {
	tr := newTestTerrace()
	tr.GiveUp()
	assert.Equal(t, TerracePhaseGameOver, tr.GetPhase())
	require.NotEmpty(t, tr.GetActionLog())

	before := len(tr.GetActionLog())
	tr.GiveUp()
	assert.Len(t, tr.GetActionLog(), before)
}

func TestTerrace_Undo(t *testing.T) {
	t.Run("restores the previous position", func(t *testing.T) {
		tr := newTestTerrace()
		clearTerraceBoard(tr)
		fillTerracePiles(tr)
		tr.reserve = []*Card{NewCard(CardDesignSpade, 5, true)}

		require.NoError(t, tr.MoveReserveToFoundation())
		require.True(t, tr.CanUndo())
		require.NoError(t, tr.Undo())

		assert.Len(t, tr.GetReserve(), 1)
		assert.Empty(t, tr.GetFoundation()[0])
		assert.Equal(t, 0, tr.GetMoveCount())
	})

	// The base rank is part of the position: undoing the very first foundation
	// card has to give the player the choice back.
	t.Run("restores an unset base rank", func(t *testing.T) {
		tr := newTestTerrace()
		clearTerraceBoard(tr)
		fillTerracePiles(tr)
		tr.reserve = []*Card{NewCard(CardDesignSpade, 5, true)}

		require.NoError(t, tr.MoveReserveToFoundation())
		require.Equal(t, 5, tr.GetBaseRank())
		require.NoError(t, tr.Undo())
		assert.Equal(t, 0, tr.GetBaseRank())
		assert.True(t, tr.IsAwaitingBaseRank())
	})

	t.Run("errors with no history", func(t *testing.T) {
		tr := newTestTerrace()
		tr.history = nil
		assert.False(t, tr.CanUndo())
		assert.Error(t, tr.Undo())
	})
}

func TestTerrace_UndoN(t *testing.T) {
	tr := newTestTerrace()
	clearTerraceBoard(tr)
	fillTerracePiles(tr)
	tr.reserve = []*Card{NewCard(CardDesignHeart, 5, true), NewCard(CardDesignSpade, 5, true)}

	require.NoError(t, tr.MoveReserveToFoundation())
	require.NoError(t, tr.MoveReserveToFoundation())
	require.NoError(t, tr.UndoN(2))
	assert.Len(t, tr.GetReserve(), 2)
	assert.Equal(t, 0, tr.GetMoveCount())

	assert.Error(t, tr.UndoN(0))
	assert.Error(t, tr.UndoN(-1))
	assert.Error(t, tr.UndoN(99))
}

func TestTerrace_UndoToEscape(t *testing.T) {
	t.Run("zero when not stuck", func(t *testing.T) {
		tr := newTestTerrace()
		assert.Equal(t, 0, tr.UndoToEscape())
	})

	t.Run("counts back to the last live position", func(t *testing.T) {
		tr := newTestTerrace()
		clearTerraceBoard(tr)
		fillTerracePiles(tr)
		tr.baseRank = 5
		tr.reserve = []*Card{NewCard(CardDesignSpade, 5, true)}
		tr.checkStalemate()
		require.False(t, tr.IsStalemate())

		require.NoError(t, tr.MoveReserveToFoundation())
		require.True(t, tr.IsStalemate())
		assert.Equal(t, 1, tr.UndoToEscape())
	})

	t.Run("-1 when every recorded position was already stuck", func(t *testing.T) {
		tr := newTestTerrace()
		tr.isStalemate = true
		tr.history = []*terraceSnapshot{{isStalemate: true}}
		assert.Equal(t, -1, tr.UndoToEscape())
	})
}

func TestTerrace_ActionLog(t *testing.T) {
	tr := newTestTerrace()
	clearTerraceBoard(tr)
	fillTerracePiles(tr)
	tr.tableau[0] = []*Card{NewCard(CardDesignSpade, 7, true)}
	tr.tableau[1] = []*Card{NewCard(CardDesignHeart, 6, true)}
	tr.reserve = []*Card{NewCard(CardDesignSpade, 5, true)}
	tr.waste = []*Card{NewCard(CardDesignHeart, 5, true)}
	// Two: moving pile 1's only card empties it, and the gap immediately eats a
	// stock card, so the Draw at the end needs one of its own.
	tr.stock = []*Card{NewCard(CardDesignDiamond, 7, true), NewCard(CardDesignDiamond, 8, true)}

	require.NoError(t, tr.MoveReserveToFoundation())
	require.NoError(t, tr.MoveWasteToFoundation())
	require.NoError(t, tr.MoveTableauToTableau(1, 0))
	require.NoError(t, tr.Draw())

	// The board is 0-indexed everywhere, so the log must be too -- a 1-based log
	// silently disagrees with the hint and the CLI.
	details := make([]string, 0, len(tr.GetActionLog()))
	for _, e := range tr.GetActionLog() {
		details = append(details, e.Detail)
	}
	assert.Equal(t, []string{
		"テラス→基礎札0",
		"捨て札→基礎札1",
		"タブロー山1→タブロー山0",
		"山札から1枚めくった",
	}, details)
}

func TestTerrace_JSONRoundTrip(t *testing.T) {
	tr := newTestTerrace()
	require.NoError(t, tr.Draw())

	data, err := json.Marshal(tr)
	require.NoError(t, err)

	restored := NewDefaultTerrace()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, tr.GetPhase(), restored.GetPhase())
	assert.Equal(t, tr.GetMoveCount(), restored.GetMoveCount())
	assert.Equal(t, tr.GetStockCount(), restored.GetStockCount())
	assert.Equal(t, tr.GetBaseRank(), restored.GetBaseRank())
	assert.Len(t, restored.GetReserve(), len(tr.GetReserve()))
}

// The Worker rebuilds the game from KV on every request, so the undo stack has
// to round-trip or Undo is dead in production (#4478). This undoes rather than
// counting fields, which is what catches a blank-snapshot regression.
func TestTerrace_UndoSurvivesAKVRoundTrip(t *testing.T) {
	tr := newTestTerrace()
	clearTerraceBoard(tr)
	fillTerracePiles(tr)
	tr.reserve = []*Card{NewCard(CardDesignSpade, 5, true)}
	require.NoError(t, tr.MoveReserveToFoundation())

	data, err := json.Marshal(tr)
	require.NoError(t, err)
	restored := NewDefaultTerrace()
	require.NoError(t, json.Unmarshal(data, restored))

	require.True(t, restored.CanUndo(), "the undo stack must survive KV")
	require.NoError(t, restored.Undo())
	assert.Len(t, restored.GetReserve(), 1, "the card came back")
	assert.Equal(t, 0, restored.GetBaseRank(), "and so did the unset base rank")
}

// KV can hold anything an earlier version wrote, so a corrupt snapshot must be
// refused rather than started.
func TestTerrace_UnmarshalJSONValidation(t *testing.T) {
	huge := make([]*Card, TerraceTotalCards+1)
	for i := range huge {
		huge[i] = NewCard(CardDesignSpade, 1, true)
	}
	overFoundation := make([]*Card, TerraceFoundationTarget+1)
	for i := range overFoundation {
		overFoundation[i] = NewCard(CardDesignSpade, 1, true)
	}
	overGuard := make([]*Card, terraceMaxSliceLen+1)
	for i := range overGuard {
		overGuard[i] = NewCard(CardDesignSpade, 1, true)
	}
	bigLogEntries := make([]*ActionLogEntry, terraceMaxSliceLen+1)
	for i := range bigLogEntries {
		bigLogEntries[i] = &ActionLogEntry{}
	}

	for _, tc := range []struct {
		name string
		j    terraceJSON
	}{
		{"phase below range", terraceJSON{Phase: -1}},
		{"phase above range", terraceJSON{Phase: TerracePhaseGameOver + 1}},
		{"negative move count", terraceJSON{MoveCount: -1}},
		{"base rank out of range", terraceJSON{BaseRank: CardValueMax + 1}},
		{"reserve overflows", terraceJSON{Reserve: huge}},
		{"stock overflows", terraceJSON{Stock: huge}},
		{"waste overflows", terraceJSON{Waste: huge}},
		{"foundation overflows", terraceJSON{Foundation: [TerraceFoundationCnt][]*Card{overFoundation}}},
		{"tableau overflows", terraceJSON{Tableau: [TerraceTableauCnt][]*Card{huge}}},
		{"action log overflows", terraceJSON{ActionLog: bigLogEntries}},
		{"history overflows", terraceJSON{History: make([]*terraceSnapshot, terraceMaxSliceLen+1)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(&tc.j)
			require.NoError(t, err)
			assert.Error(t, NewDefaultTerrace().UnmarshalJSON(data))
		})
	}

	t.Run("malformed json", func(t *testing.T) {
		assert.Error(t, NewDefaultTerrace().UnmarshalJSON([]byte("not json")))
	})

	t.Run("an oversized pile inside a snapshot", func(t *testing.T) {
		data, err := json.Marshal(&terraceSnapshotJSON{Stock: overGuard})
		require.NoError(t, err)
		assert.Error(t, new(terraceSnapshot).UnmarshalJSON(data))
	})

	t.Run("malformed snapshot json", func(t *testing.T) {
		assert.Error(t, new(terraceSnapshot).UnmarshalJSON([]byte("not json")))
	})

	t.Run("a valid snapshot is accepted", func(t *testing.T) {
		data, err := json.Marshal(&terraceJSON{Phase: TerracePhasePlaying, MoveCount: 3, BaseRank: 5})
		require.NoError(t, err)
		tr := NewDefaultTerrace()
		require.NoError(t, tr.UnmarshalJSON(data))
		assert.Equal(t, 3, tr.GetMoveCount())
		assert.Equal(t, 5, tr.GetBaseRank())
	})
}

// Drive a full game to make sure no path panics and the invariants hold
// throughout. The deal is random, so this is a fuzz-ish smoke test rather than
// an assertion about any particular position.
func TestTerrace_FullGameDrive(t *testing.T) {
	for range 20 {
		tr := newTestTerrace()
		for range 600 {
			if tr.GetGameEndFlag() {
				break
			}
			h := tr.GetHint()
			if h == nil {
				break
			}
			var err error
			switch {
			case h.FromZone == "stock":
				err = tr.Draw()
			case h.FromZone == "reserve":
				err = tr.MoveReserveToFoundation()
			case h.FromZone == "waste" && h.ToZone == "foundation":
				err = tr.MoveWasteToFoundation()
			case h.FromZone == "waste":
				err = tr.MoveWasteToTableau(h.ToIdx)
			case h.ToZone == "foundation":
				err = tr.MoveTableauToFoundation(h.FromIdx)
			default:
				err = tr.MoveTableauToTableau(h.FromIdx, h.ToIdx)
			}
			require.NoError(t, err)

			for i := range TerraceFoundationCnt {
				assert.LessOrEqual(t, len(tr.GetFoundation()[i]), TerraceFoundationTarget)
			}
		}
	}
}
