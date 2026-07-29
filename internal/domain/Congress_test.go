//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestCongress() *Congress {
	c := NewDefaultCongress()
	c.Reset()
	return c
}

// clearCongressBoard wipes the dealt layout so a test can state exactly the
// position it cares about. Never assert on a freshly Reset board -- the deal is
// shuffled, so any such assertion is a hidden flake.
func clearCongressBoard(c *Congress) {
	c.stock = nil
	c.waste = nil
	c.isStalemate = false
	for i := range CongressFoundationCnt {
		c.foundation[i] = nil
	}
	for i := range CongressTableauCnt {
		c.tableau[i] = nil
	}
}

// fillCongressPiles puts one dead card in every pile so no gap exists. A gap
// changes which moves are legal, so a board full of holes would test a position
// the game never reaches by accident.
func fillCongressPiles(c *Congress) {
	for i := range CongressTableauCnt {
		c.tableau[i] = []*Card{NewCard(CardDesignSpade, 9, true)}
	}
}

func TestNewCongress(t *testing.T) {
	assert.NotNil(t, NewCongress(NewTrumpCardsWithDecks(2, 0)))
	assert.NotNil(t, NewDefaultCongress())
}

// The deal is 8 piles of ONE card with the other 96 as stock. #4419 puts all 96
// into the tableau, which would leave the stock empty and contradict its own
// rule 5 ("turn cards from the stock").
func TestCongress_Reset(t *testing.T) {
	c := newTestCongress()

	for i, pile := range c.GetTableau() {
		assert.Len(t, pile, 1, "pile %d", i)
	}
	assert.Equal(t, CongressTotalCards-CongressTableauCnt, c.GetStockCount())
	assert.Empty(t, c.GetWaste())

	// The foundations start EMPTY -- Aces go up as they turn up. #4419 deals all
	// eight Aces onto them, which would need the deck searched for them.
	for i, pile := range c.GetFoundation() {
		assert.Empty(t, pile, "foundation %d", i)
	}

	assert.Equal(t, CongressPhasePlaying, c.GetPhase())
	assert.Equal(t, 0, c.GetMoveCount())
	assert.True(t, c.AllFaceUp())
	assert.False(t, c.GetGameEndFlag())
}

func TestCongress_ResetDealsEveryCard(t *testing.T) {
	for range 20 {
		c := newTestCongress()
		total := c.GetStockCount() + len(c.GetWaste())
		for _, pile := range c.GetTableau() {
			total += len(pile)
		}
		for _, pile := range c.GetFoundation() {
			total += len(pile)
		}
		assert.Equal(t, CongressTotalCards, total)
	}
}

func TestCongress_ResetIsRepeatable(t *testing.T) {
	c := newTestCongress()
	require.NoError(t, c.Draw())
	c.Reset()
	assert.Equal(t, 0, c.GetMoveCount())
	assert.Empty(t, c.GetWaste())
	assert.Empty(t, c.GetActionLog())
	assert.False(t, c.CanUndo())
}

// Eight foundations of 13 absorb all 104 cards. #4419's "up to 8 per suit" comes
// to 32, which cannot finish the game.
func TestCongress_FoundationRules(t *testing.T) {
	c := newTestCongress()
	clearCongressBoard(c)

	t.Run("an empty foundation takes only an Ace of its own suit", func(t *testing.T) {
		assert.True(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 1, true), 0))
		assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 2, true), 0))
		assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignHeart, 1, true), 0))
		assert.False(t, c.canPlaceOnFoundation(nil, 0))
	})

	t.Run("there are two foundations per suit", func(t *testing.T) {
		c.foundation[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
		assert.Equal(t, 4, c.findFoundation(NewCard(CardDesignSpade, 1, true)),
			"the second spade Ace opens the second spade foundation")
	})

	t.Run("builds up in suit to the King", func(t *testing.T) {
		c.foundation[1] = []*Card{NewCard(CardDesignClover, 1, true)}
		assert.True(t, c.canPlaceOnFoundation(NewCard(CardDesignClover, 2, true), 1))
		assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignClover, 3, true), 1))
	})

	t.Run("stops at the King", func(t *testing.T) {
		c.foundation[2] = nil
		for v := 1; v <= CardValueMax; v++ {
			c.foundation[2] = append(c.foundation[2], NewCard(CardDesignHeart, v, true))
		}
		require.Len(t, c.foundation[2], CongressFoundationTarget)
		for v := 1; v <= CardValueMax; v++ {
			assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignHeart, v, true), 2), "rank %d", v)
		}
	})
}

// The tableau builds down ignoring suit and colour, and does NOT wrap: an Ace
// ends its pile.
func TestCongress_TableauBuildRules(t *testing.T) {
	c := newTestCongress()
	clearCongressBoard(c)
	fillCongressPiles(c)
	c.tableau[0] = []*Card{NewCard(CardDesignSpade, 7, true)}

	assert.True(t, c.canPlaceOnTableau(NewCard(CardDesignSpade, 6, true), 0))
	assert.True(t, c.canPlaceOnTableau(NewCard(CardDesignHeart, 6, true), 0), "suit is ignored")
	assert.False(t, c.canPlaceOnTableau(NewCard(CardDesignHeart, 5, true), 0))
	assert.False(t, c.canPlaceOnTableau(nil, 0))

	c.tableau[1] = []*Card{NewCard(CardDesignClover, 1, true)}
	assert.False(t, c.canPlaceOnTableau(NewCard(CardDesignClover, CardValueMax, true), 1),
		"there is no Ace-to-King wraparound")
}

func TestCongress_Draw(t *testing.T) {
	t.Run("moves one card to the waste", func(t *testing.T) {
		c := newTestCongress()
		before := c.GetStockCount()
		require.NoError(t, c.Draw())
		assert.Equal(t, before-1, c.GetStockCount())
		assert.Len(t, c.GetWaste(), 1)
	})

	// One pass only -- there is no redeal, so an empty stock is final.
	t.Run("refuses once the stock is out", func(t *testing.T) {
		c := newTestCongress()
		clearCongressBoard(c)
		fillCongressPiles(c)
		c.waste = []*Card{NewCard(CardDesignSpade, 4, true)}
		assert.Error(t, c.Draw(), "the waste is never recycled")
	})

	t.Run("errors once the game has ended", func(t *testing.T) {
		c := newTestCongress()
		c.GiveUp()
		assert.Error(t, c.Draw())
	})
}

// Only the stock and the waste may fill a gap. #4419 does not mention this, and
// letting the tableau fill its own gaps would make the piles freely reorderable.
func TestCongress_EmptyPileRules(t *testing.T) {
	t.Run("a tableau card may not fill an empty pile", func(t *testing.T) {
		c := newTestCongress()
		clearCongressBoard(c)
		fillCongressPiles(c)
		c.tableau[0] = []*Card{NewCard(CardDesignSpade, 7, true)}
		c.tableau[1] = nil

		assert.Error(t, c.MoveTableauToTableau(0, 1))
		assert.Empty(t, c.GetTableau()[1])
	})

	t.Run("the waste may fill an empty pile", func(t *testing.T) {
		c := newTestCongress()
		clearCongressBoard(c)
		fillCongressPiles(c)
		c.tableau[2] = nil
		c.waste = []*Card{NewCard(CardDesignHeart, 8, true)}

		require.NoError(t, c.MoveWasteToTableau(2))
		assert.Len(t, c.GetTableau()[2], 1)
	})

	// Routing a gap-fill through the waste would burn a stock card early, so the
	// stock can also fill a gap directly.
	t.Run("the stock may fill an empty pile directly", func(t *testing.T) {
		c := newTestCongress()
		clearCongressBoard(c)
		fillCongressPiles(c)
		c.tableau[3] = nil
		top := NewCard(CardDesignDiamond, 6, true)
		c.stock = []*Card{top, NewCard(CardDesignSpade, 2, true)}

		require.NoError(t, c.MoveStockToTableau(3))
		assert.Equal(t, top, c.GetTableau()[3][0])
		assert.Equal(t, 1, c.GetStockCount())
		assert.Empty(t, c.GetWaste(), "the card never touches the waste")
	})

	t.Run("the stock may not be played onto an occupied pile", func(t *testing.T) {
		c := newTestCongress()
		clearCongressBoard(c)
		fillCongressPiles(c)
		c.stock = []*Card{NewCard(CardDesignSpade, 8, true)}

		assert.Error(t, c.MoveStockToTableau(0))
	})

	t.Run("rejects an empty stock and a bad index", func(t *testing.T) {
		c := newTestCongress()
		clearCongressBoard(c)
		c.tableau[0] = nil

		assert.Error(t, c.MoveStockToTableau(0))
		assert.Error(t, c.MoveStockToTableau(-1))
		assert.Error(t, c.MoveStockToTableau(CongressTableauCnt))
	})

	t.Run("errors once the game has ended", func(t *testing.T) {
		c := newTestCongress()
		c.GiveUp()
		assert.Error(t, c.MoveStockToTableau(0))
	})
}

func TestCongress_MoveTableauToFoundation(t *testing.T) {
	t.Run("moves the pile top", func(t *testing.T) {
		c := newTestCongress()
		clearCongressBoard(c)
		fillCongressPiles(c)
		c.tableau[4] = []*Card{NewCard(CardDesignHeart, 1, true)}

		require.NoError(t, c.MoveTableauToFoundation(4))
		assert.Len(t, c.GetFoundation()[2], 1)
		assert.Empty(t, c.GetTableau()[4])
	})

	t.Run("rejects an empty pile, an illegal card and a bad index", func(t *testing.T) {
		c := newTestCongress()
		clearCongressBoard(c)
		fillCongressPiles(c)
		c.tableau[0] = nil

		assert.Error(t, c.MoveTableauToFoundation(0))
		assert.Error(t, c.MoveTableauToFoundation(1))
		assert.Error(t, c.MoveTableauToFoundation(-1))
		assert.Error(t, c.MoveTableauToFoundation(CongressTableauCnt))
	})

	t.Run("errors once the game has ended", func(t *testing.T) {
		c := newTestCongress()
		c.GiveUp()
		assert.Error(t, c.MoveTableauToFoundation(0))
	})
}

// Cards move one at a time; there is no run-carrying in this game.
func TestCongress_MoveTableauToTableau(t *testing.T) {
	t.Run("moves a single card", func(t *testing.T) {
		c := newTestCongress()
		clearCongressBoard(c)
		fillCongressPiles(c)
		c.tableau[0] = []*Card{NewCard(CardDesignHeart, 2, true), NewCard(CardDesignSpade, 6, true)}
		c.tableau[1] = []*Card{NewCard(CardDesignClover, 7, true)}

		require.NoError(t, c.MoveTableauToTableau(0, 1))
		assert.Len(t, c.GetTableau()[0], 1, "only the top card left")
		assert.Len(t, c.GetTableau()[1], 2)
	})

	t.Run("rejects bad indices and illegal ranks", func(t *testing.T) {
		c := newTestCongress()
		clearCongressBoard(c)
		fillCongressPiles(c)
		c.tableau[0] = []*Card{NewCard(CardDesignSpade, 6, true)}
		c.tableau[1] = []*Card{NewCard(CardDesignSpade, 7, true)}
		c.tableau[2] = nil

		assert.Error(t, c.MoveTableauToTableau(-1, 1))
		assert.Error(t, c.MoveTableauToTableau(0, CongressTableauCnt))
		assert.Error(t, c.MoveTableauToTableau(0, 0), "same pile")
		assert.Error(t, c.MoveTableauToTableau(2, 1), "empty source")
		assert.Error(t, c.MoveTableauToTableau(1, 0), "rank does not fit")
	})

	t.Run("errors once the game has ended", func(t *testing.T) {
		c := newTestCongress()
		c.GiveUp()
		assert.Error(t, c.MoveTableauToTableau(0, 1))
	})
}

func TestCongress_MoveWaste(t *testing.T) {
	t.Run("to a foundation", func(t *testing.T) {
		c := newTestCongress()
		clearCongressBoard(c)
		fillCongressPiles(c)
		c.waste = []*Card{NewCard(CardDesignClover, 1, true)}

		require.NoError(t, c.MoveWasteToFoundation())
		assert.Len(t, c.GetFoundation()[1], 1)
		assert.Empty(t, c.GetWaste())
	})

	t.Run("rejects an empty waste, an illegal card and a bad index", func(t *testing.T) {
		c := newTestCongress()
		clearCongressBoard(c)
		fillCongressPiles(c)

		assert.Error(t, c.MoveWasteToFoundation())
		assert.Error(t, c.MoveWasteToTableau(0))

		c.waste = []*Card{NewCard(CardDesignHeart, 4, true)}
		assert.Error(t, c.MoveWasteToFoundation())
		assert.Error(t, c.MoveWasteToTableau(0))
		assert.Error(t, c.MoveWasteToTableau(-1))
		assert.Error(t, c.MoveWasteToTableau(CongressTableauCnt))
	})

	t.Run("errors once the game has ended", func(t *testing.T) {
		c := newTestCongress()
		c.GiveUp()
		assert.Error(t, c.MoveWasteToFoundation())
		assert.Error(t, c.MoveWasteToTableau(0))
	})
}

func TestCongress_GetHint(t *testing.T) {
	t.Run("prefers a foundation move", func(t *testing.T) {
		c := newTestCongress()
		clearCongressBoard(c)
		fillCongressPiles(c)
		c.tableau[3] = []*Card{NewCard(CardDesignSpade, 1, true)}

		h := c.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "tableau", h.FromZone)
		assert.Equal(t, 3, h.FromIdx)
		assert.Equal(t, "foundation", h.ToZone)
	})

	t.Run("offers the waste to a foundation", func(t *testing.T) {
		c := newTestCongress()
		clearCongressBoard(c)
		fillCongressPiles(c)
		c.waste = []*Card{NewCard(CardDesignSpade, 1, true)}

		h := c.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "waste", h.FromZone)
		assert.Equal(t, "foundation", h.ToZone)
	})

	t.Run("offers filling a gap from the stock", func(t *testing.T) {
		c := newTestCongress()
		clearCongressBoard(c)
		fillCongressPiles(c)
		c.tableau[5] = nil
		c.stock = []*Card{NewCard(CardDesignHeart, 4, true)}

		h := c.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "stock", h.FromZone)
		assert.Equal(t, "tableau", h.ToZone)
		assert.Equal(t, 5, h.ToIdx)
	})

	t.Run("offers a tableau move", func(t *testing.T) {
		c := newTestCongress()
		clearCongressBoard(c)
		fillCongressPiles(c)
		c.tableau[0] = []*Card{NewCard(CardDesignSpade, 6, true)}
		c.tableau[1] = []*Card{NewCard(CardDesignHeart, 7, true)}

		h := c.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "tableau", h.FromZone)
		assert.Equal(t, "tableau", h.ToZone)
	})

	// An empty pile is not a tableau destination, so it must not be offered.
	t.Run("never suggests moving a tableau card into a gap", func(t *testing.T) {
		c := newTestCongress()
		clearCongressBoard(c)
		fillCongressPiles(c)
		c.tableau[0] = []*Card{NewCard(CardDesignSpade, 6, true)}
		c.tableau[1] = nil

		assert.Nil(t, c.GetHint(), "no stock, no waste, and a gap is not a destination")
	})

	t.Run("falls back to drawing", func(t *testing.T) {
		c := newTestCongress()
		clearCongressBoard(c)
		fillCongressPiles(c)
		c.stock = []*Card{NewCard(CardDesignHeart, 4, true)}

		h := c.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "stock", h.FromZone)
		assert.Equal(t, "waste", h.ToZone)
	})

	t.Run("returns nil once the game has ended", func(t *testing.T) {
		c := newTestCongress()
		c.GiveUp()
		assert.Nil(t, c.GetHint())
	})
}

func TestCongress_Stalemate(t *testing.T) {
	t.Run("a dead board is a stalemate", func(t *testing.T) {
		c := newTestCongress()
		clearCongressBoard(c)
		fillCongressPiles(c)
		c.checkStalemate()
		assert.True(t, c.IsStalemate(), "every pile is a 9, no stock, no waste")
	})

	t.Run("a board with one legal move is not", func(t *testing.T) {
		c := newTestCongress()
		clearCongressBoard(c)
		fillCongressPiles(c)
		c.tableau[7] = []*Card{NewCard(CardDesignSpade, 1, true)}
		c.checkStalemate()
		assert.False(t, c.IsStalemate())
	})
}

func TestCongress_AutoComplete(t *testing.T) {
	t.Run("sends every reachable card to the foundations", func(t *testing.T) {
		c := newTestCongress()
		clearCongressBoard(c)
		fillCongressPiles(c)
		c.tableau[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
		c.tableau[1] = []*Card{NewCard(CardDesignClover, 1, true)}
		c.waste = []*Card{NewCard(CardDesignHeart, 1, true)}

		require.NoError(t, c.AutoComplete())
		assert.Len(t, c.GetFoundation()[0], 1)
		assert.Len(t, c.GetFoundation()[1], 1)
		assert.Len(t, c.GetFoundation()[2], 1)
	})

	// Auto-complete must never rearrange the tableau: that is a strategic
	// judgement, not a mechanical tidy-up.
	t.Run("never performs a tableau move", func(t *testing.T) {
		c := newTestCongress()
		clearCongressBoard(c)
		fillCongressPiles(c)
		c.tableau[0] = []*Card{NewCard(CardDesignSpade, 6, true)}
		c.tableau[1] = []*Card{NewCard(CardDesignHeart, 7, true)}

		assert.Error(t, c.AutoComplete())
		assert.Len(t, c.GetTableau()[0], 1)
		assert.Len(t, c.GetTableau()[1], 1)
	})

	t.Run("errors once the game has ended", func(t *testing.T) {
		c := newTestCongress()
		c.GiveUp()
		assert.Error(t, c.AutoComplete())
	})
}

func TestCongress_GameClear(t *testing.T) {
	c := newTestCongress()
	clearCongressBoard(c)
	fillCongressPiles(c)
	for i := range CongressFoundationCnt {
		c.foundation[i] = nil
		for v := 1; v <= CardValueMax; v++ {
			c.foundation[i] = append(c.foundation[i], NewCard(congressSuitOrder[i], v, true))
		}
	}
	// Strip one card back off so the final move completes the game.
	c.foundation[0] = c.foundation[0][:CardValueMax-1]
	c.tableau[0] = []*Card{NewCard(CardDesignSpade, CardValueMax, true)}

	require.NoError(t, c.MoveTableauToFoundation(0))
	assert.Equal(t, CongressPhaseGameClear, c.GetPhase())
	assert.True(t, c.GetGameEndFlag())
}

func TestCongress_GiveUp(t *testing.T) {
	c := newTestCongress()
	c.GiveUp()
	assert.Equal(t, CongressPhaseGameOver, c.GetPhase())
	require.NotEmpty(t, c.GetActionLog())

	before := len(c.GetActionLog())
	c.GiveUp()
	assert.Len(t, c.GetActionLog(), before)
}

func TestCongress_Undo(t *testing.T) {
	t.Run("restores the previous position", func(t *testing.T) {
		c := newTestCongress()
		clearCongressBoard(c)
		fillCongressPiles(c)
		c.tableau[0] = []*Card{NewCard(CardDesignSpade, 1, true)}

		require.NoError(t, c.MoveTableauToFoundation(0))
		require.True(t, c.CanUndo())
		require.NoError(t, c.Undo())

		assert.Len(t, c.GetTableau()[0], 1)
		assert.Empty(t, c.GetFoundation()[0])
		assert.Equal(t, 0, c.GetMoveCount())
	})

	t.Run("errors with no history", func(t *testing.T) {
		c := newTestCongress()
		c.history = nil
		assert.False(t, c.CanUndo())
		assert.Error(t, c.Undo())
	})
}

func TestCongress_UndoN(t *testing.T) {
	c := newTestCongress()
	clearCongressBoard(c)
	fillCongressPiles(c)
	c.tableau[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
	c.tableau[1] = []*Card{NewCard(CardDesignClover, 1, true)}

	require.NoError(t, c.MoveTableauToFoundation(0))
	require.NoError(t, c.MoveTableauToFoundation(1))
	require.NoError(t, c.UndoN(2))
	assert.Len(t, c.GetTableau()[0], 1)
	assert.Equal(t, 0, c.GetMoveCount())

	assert.Error(t, c.UndoN(0))
	assert.Error(t, c.UndoN(-1))
	assert.Error(t, c.UndoN(99))
}

func TestCongress_UndoToEscape(t *testing.T) {
	t.Run("zero when not stuck", func(t *testing.T) {
		c := newTestCongress()
		assert.Equal(t, 0, c.UndoToEscape())
	})

	t.Run("counts back to the last live position", func(t *testing.T) {
		c := newTestCongress()
		clearCongressBoard(c)
		fillCongressPiles(c)
		c.tableau[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
		c.checkStalemate()
		require.False(t, c.IsStalemate())

		require.NoError(t, c.MoveTableauToFoundation(0))
		require.True(t, c.IsStalemate(), "the gap cannot be filled without stock or waste")
		assert.Equal(t, 1, c.UndoToEscape())
	})

	t.Run("-1 when every recorded position was already stuck", func(t *testing.T) {
		c := newTestCongress()
		c.isStalemate = true
		c.history = []*congressSnapshot{{isStalemate: true}}
		assert.Equal(t, -1, c.UndoToEscape())
	})
}

func TestCongress_ActionLog(t *testing.T) {
	c := newTestCongress()
	clearCongressBoard(c)
	fillCongressPiles(c)
	c.tableau[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
	c.tableau[1] = []*Card{NewCard(CardDesignHeart, 7, true)}
	c.tableau[2] = []*Card{NewCard(CardDesignClover, 8, true)}
	c.waste = []*Card{NewCard(CardDesignClover, 1, true)}
	c.stock = []*Card{NewCard(CardDesignDiamond, 7, true), NewCard(CardDesignDiamond, 8, true)}

	require.NoError(t, c.MoveTableauToFoundation(0))
	require.NoError(t, c.MoveStockToTableau(0))
	require.NoError(t, c.MoveWasteToFoundation())
	require.NoError(t, c.MoveTableauToTableau(1, 2))
	require.NoError(t, c.Draw())

	// The board is 0-indexed everywhere, so the log must be too -- a 1-based log
	// silently disagrees with the hint and the CLI.
	details := make([]string, 0, len(c.GetActionLog()))
	for _, e := range c.GetActionLog() {
		details = append(details, e.Detail)
	}
	assert.Equal(t, []string{
		"タブロー山0→基礎札0",
		"山札→タブロー山0",
		"捨て札→基礎札1",
		"タブロー山1→タブロー山2",
		"山札から1枚めくった",
	}, details)
}

func TestCongress_JSONRoundTrip(t *testing.T) {
	c := newTestCongress()
	require.NoError(t, c.Draw())

	data, err := json.Marshal(c)
	require.NoError(t, err)

	restored := NewDefaultCongress()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, c.GetPhase(), restored.GetPhase())
	assert.Equal(t, c.GetMoveCount(), restored.GetMoveCount())
	assert.Equal(t, c.GetStockCount(), restored.GetStockCount())
	assert.Len(t, restored.GetWaste(), len(c.GetWaste()))
}

// The Worker rebuilds the game from KV on every request, so the undo stack has
// to round-trip or Undo is dead in production (#4478). Asserting on restored
// field counts would not catch a blank-snapshot regression -- this undoes.
func TestCongress_UndoSurvivesAKVRoundTrip(t *testing.T) {
	c := newTestCongress()
	clearCongressBoard(c)
	fillCongressPiles(c)
	c.tableau[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
	require.NoError(t, c.MoveTableauToFoundation(0))

	data, err := json.Marshal(c)
	require.NoError(t, err)
	restored := NewDefaultCongress()
	require.NoError(t, json.Unmarshal(data, restored))

	require.True(t, restored.CanUndo(), "the undo stack must survive KV")
	require.NoError(t, restored.Undo())
	assert.Len(t, restored.GetTableau()[0], 1, "the card came back")
	assert.Empty(t, restored.GetFoundation()[0])
}

// KV can hold anything an earlier version wrote, so a corrupt snapshot must be
// refused rather than started.
func TestCongress_UnmarshalJSONValidation(t *testing.T) {
	huge := make([]*Card, CongressTotalCards+1)
	for i := range huge {
		huge[i] = NewCard(CardDesignSpade, 1, true)
	}
	overFoundation := make([]*Card, CongressFoundationTarget+1)
	for i := range overFoundation {
		overFoundation[i] = NewCard(CardDesignSpade, 1, true)
	}
	overGuard := make([]*Card, congressMaxSliceLen+1)
	for i := range overGuard {
		overGuard[i] = NewCard(CardDesignSpade, 1, true)
	}
	bigLogEntries := make([]*ActionLogEntry, congressMaxSliceLen+1)
	for i := range bigLogEntries {
		bigLogEntries[i] = &ActionLogEntry{}
	}

	for _, tc := range []struct {
		name string
		j    congressJSON
	}{
		{"phase below range", congressJSON{Phase: -1}},
		{"phase above range", congressJSON{Phase: CongressPhaseGameOver + 1}},
		{"negative move count", congressJSON{MoveCount: -1}},
		{"stock overflows", congressJSON{Stock: huge}},
		{"waste overflows", congressJSON{Waste: huge}},
		{"foundation overflows", congressJSON{Foundation: [CongressFoundationCnt][]*Card{overFoundation}}},
		{"tableau overflows", congressJSON{Tableau: [CongressTableauCnt][]*Card{huge}}},
		{"action log overflows", congressJSON{ActionLog: bigLogEntries}},
		{"history overflows", congressJSON{History: make([]*congressSnapshot, congressMaxSliceLen+1)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(&tc.j)
			require.NoError(t, err)
			assert.Error(t, NewDefaultCongress().UnmarshalJSON(data))
		})
	}

	t.Run("malformed json", func(t *testing.T) {
		assert.Error(t, NewDefaultCongress().UnmarshalJSON([]byte("not json")))
	})

	t.Run("an oversized pile inside a snapshot", func(t *testing.T) {
		data, err := json.Marshal(&congressSnapshotJSON{Stock: overGuard})
		require.NoError(t, err)
		assert.Error(t, new(congressSnapshot).UnmarshalJSON(data))
	})

	t.Run("malformed snapshot json", func(t *testing.T) {
		assert.Error(t, new(congressSnapshot).UnmarshalJSON([]byte("not json")))
	})

	t.Run("a valid snapshot is accepted", func(t *testing.T) {
		data, err := json.Marshal(&congressJSON{Phase: CongressPhasePlaying, MoveCount: 3})
		require.NoError(t, err)
		c := NewDefaultCongress()
		require.NoError(t, c.UnmarshalJSON(data))
		assert.Equal(t, 3, c.GetMoveCount())
	})
}

// Drive a full game to make sure no path panics and the invariants hold
// throughout. The deal is random, so this is a fuzz-ish smoke test rather than
// an assertion about any particular position.
func TestCongress_FullGameDrive(t *testing.T) {
	for range 20 {
		c := newTestCongress()
		for range 600 {
			if c.GetGameEndFlag() {
				break
			}
			h := c.GetHint()
			if h == nil {
				break
			}
			var err error
			switch {
			case h.FromZone == "stock" && h.ToZone == "waste":
				err = c.Draw()
			case h.FromZone == "stock":
				err = c.MoveStockToTableau(h.ToIdx)
			case h.FromZone == "waste" && h.ToZone == "foundation":
				err = c.MoveWasteToFoundation()
			case h.FromZone == "waste":
				err = c.MoveWasteToTableau(h.ToIdx)
			case h.ToZone == "foundation":
				err = c.MoveTableauToFoundation(h.FromIdx)
			default:
				err = c.MoveTableauToTableau(h.FromIdx, h.ToIdx)
			}
			require.NoError(t, err)

			for i := range CongressFoundationCnt {
				assert.LessOrEqual(t, len(c.GetFoundation()[i]), CongressFoundationTarget)
			}
		}
	}
}
