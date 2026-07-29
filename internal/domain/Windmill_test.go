//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestWindmill() *Windmill {
	w := NewDefaultWindmill()
	w.Reset()
	return w
}

// clearWindmillBoard wipes the dealt layout so a test can state exactly the
// position it cares about. Never assert on a freshly Reset board -- the deal is
// shuffled, so any such assertion is a hidden flake.
func clearWindmillBoard(w *Windmill) {
	w.center = nil
	w.stock = nil
	w.waste = nil
	w.transferBlocked = false
	w.isStalemate = false
	for i := range WindmillCornerCnt {
		w.corners[i] = nil
	}
	for i := range WindmillSailCnt {
		w.sails[i] = nil
	}
}

// fillWindmillSails puts a dead card (rank 9, playable nowhere in these tests)
// in every sail. A real game never has more than one gap at a time, because the
// vacated slot is refilled immediately -- so a test board with seven holes would
// make refillSails look wrong when it is only doing its job.
func fillWindmillSails(w *Windmill) {
	for i := range WindmillSailCnt {
		w.sails[i] = NewCard(CardDesignSpade, 9, true)
	}
}

// windmillCenterUpTo fills the centre foundation with `n` cards so tests can
// place the board at any point of the four A-K circuits.
func windmillCenterUpTo(w *Windmill, n int) {
	w.center = nil
	for i := range n {
		w.center = append(w.center, NewCard(CardDesignSpade, windmillNextCenterRank(i), true))
	}
}

func TestNewWindmill(t *testing.T) {
	assert.NotNil(t, NewWindmill(NewTrumpCardsWithDecks(2, 0)))
	assert.NotNil(t, NewDefaultWindmill())
}

func TestWindmill_Reset(t *testing.T) {
	w := newTestWindmill()

	// The centre opens with a single Ace; the eight sails and the stock hold
	// everything else. 1 + 8 + 95 = 104.
	require.Len(t, w.GetCenter(), 1)
	assert.Equal(t, 1, w.GetCenter()[0].GetValue())
	sails := w.GetSails()
	for i := range WindmillSailCnt {
		assert.NotNil(t, sails[i], "sail %d", i)
	}
	assert.Equal(t, WindmillTotalCards-1-WindmillSailCnt, w.GetStockCount())
	assert.Empty(t, w.GetWaste())

	// The corners start EMPTY -- Kings arrive as they turn up. #4416 says they
	// are dealt at the start, which would hand the player four free foundations.
	for i, pile := range w.GetCorners() {
		assert.Empty(t, pile, "corner %d", i)
	}

	assert.Equal(t, WindmillPhasePlaying, w.GetPhase())
	assert.Equal(t, 0, w.GetMoveCount())
	assert.False(t, w.IsStalemate())
	assert.False(t, w.IsTransferBlocked())
	assert.True(t, w.AllFaceUp())
	assert.False(t, w.GetGameEndFlag())
}

// Every card must survive the deal exactly once -- an off-by-one in the
// "pull the first Ace out" loop would silently drop or duplicate one.
func TestWindmill_ResetDealsEveryCard(t *testing.T) {
	for range 20 {
		w := newTestWindmill()
		total := len(w.GetCenter()) + w.GetStockCount() + len(w.GetWaste())
		for _, c := range w.GetSails() {
			if c != nil {
				total++
			}
		}
		assert.Equal(t, WindmillTotalCards, total)
	}
}

func TestWindmill_ResetIsRepeatable(t *testing.T) {
	w := newTestWindmill()
	require.NoError(t, w.Draw())
	w.Reset()
	assert.Equal(t, 0, w.GetMoveCount())
	assert.Empty(t, w.GetWaste())
	assert.Empty(t, w.GetActionLog())
	assert.False(t, w.CanUndo())
}

func TestWindmill_Draw(t *testing.T) {
	t.Run("moves one card to the waste", func(t *testing.T) {
		w := newTestWindmill()
		before := w.GetStockCount()
		require.NoError(t, w.Draw())
		assert.Equal(t, before-1, w.GetStockCount())
		assert.Len(t, w.GetWaste(), 1)
		assert.Equal(t, 1, w.GetMoveCount())
	})

	t.Run("errors on an empty stock", func(t *testing.T) {
		w := newTestWindmill()
		clearWindmillBoard(w)
		assert.Error(t, w.Draw())
	})

	t.Run("errors once the game has ended", func(t *testing.T) {
		w := newTestWindmill()
		w.GiveUp()
		assert.Error(t, w.Draw())
	})
}

// The centre builds A->K four times through, wrapping King round to Ace, for 52
// cards. #4416 describes a single A->K run, which cannot absorb its half of the
// two decks.
func TestWindmill_CenterWrapsKingToAceAndStopsAtFiftyTwo(t *testing.T) {
	w := newTestWindmill()
	clearWindmillBoard(w)

	for _, tc := range []struct {
		name  string
		have  int
		card  int
		allow bool
	}{
		{"ace opens", 0, 1, true},
		{"two follows the ace", 1, 2, true},
		{"a gap is refused", 1, 3, false},
		{"king caps the first circuit", CardValueMax - 1, CardValueMax, true},
		{"ace restarts after the king", CardValueMax, 1, true},
		{"king is refused right after a king", CardValueMax, CardValueMax, false},
		{"the 52nd card is the fourth king", WindmillCenterTarget - 1, CardValueMax, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			windmillCenterUpTo(w, tc.have)
			assert.Equal(t, tc.allow, w.canPlaceOnCenter(NewCard(CardDesignHeart, tc.card, true)))
		})
	}

	t.Run("a complete centre refuses everything", func(t *testing.T) {
		windmillCenterUpTo(w, WindmillCenterTarget)
		for v := 1; v <= CardValueMax; v++ {
			assert.False(t, w.canPlaceOnCenter(NewCard(CardDesignHeart, v, true)), "rank %d", v)
		}
	})
}

func TestWindmill_CornerRules(t *testing.T) {
	w := newTestWindmill()
	clearWindmillBoard(w)

	t.Run("an empty corner takes only a King", func(t *testing.T) {
		for v := 1; v <= CardValueMax; v++ {
			assert.Equal(t, v == CardValueMax,
				w.canPlaceOnCorner(NewCard(CardDesignSpade, v, true), 0), "rank %d", v)
		}
	})

	t.Run("builds down ignoring suit", func(t *testing.T) {
		w.corners[1] = []*Card{NewCard(CardDesignSpade, CardValueMax, true)}
		assert.True(t, w.canPlaceOnCorner(NewCard(CardDesignHeart, CardValueMax-1, true), 1))
		assert.False(t, w.canPlaceOnCorner(NewCard(CardDesignHeart, CardValueMax-2, true), 1))
		assert.False(t, w.canPlaceOnCorner(NewCard(CardDesignHeart, CardValueMax, true), 1))
	})

	// K down to A is 13 cards; there is no wraparound on the corners, unlike the
	// centre. 52 + 13*4 = 104 only works if each corner takes exactly 13.
	t.Run("stops at the Ace", func(t *testing.T) {
		w.corners[2] = nil
		for v := CardValueMax; v >= 1; v-- {
			w.corners[2] = append(w.corners[2], NewCard(CardDesignClover, v, true))
		}
		require.Len(t, w.corners[2], WindmillCornerTarget)
		for v := 1; v <= CardValueMax; v++ {
			assert.False(t, w.canPlaceOnCorner(NewCard(CardDesignHeart, v, true), 2), "rank %d", v)
		}
	})
}

func TestWindmill_MoveSailToCenter(t *testing.T) {
	t.Run("moves and refills the gap from the waste", func(t *testing.T) {
		w := newTestWindmill()
		clearWindmillBoard(w)
		windmillCenterUpTo(w, 1)
		fillWindmillSails(w)
		w.sails[3] = NewCard(CardDesignHeart, 2, true)
		refill := NewCard(CardDesignClover, 9, true)
		w.waste = []*Card{NewCard(CardDesignSpade, 4, true), refill}

		require.NoError(t, w.MoveSailToCenter(3))
		assert.Len(t, w.GetCenter(), 2)
		// The gap is refilled immediately, taking the waste top first.
		assert.Equal(t, refill, w.GetSails()[3])
		assert.Len(t, w.GetWaste(), 1)
	})

	t.Run("falls back to the stock when the waste is empty", func(t *testing.T) {
		w := newTestWindmill()
		clearWindmillBoard(w)
		windmillCenterUpTo(w, 1)
		fillWindmillSails(w)
		w.sails[0] = NewCard(CardDesignHeart, 2, true)
		top := NewCard(CardDesignDiamond, 6, true)
		w.stock = []*Card{top, NewCard(CardDesignSpade, 7, true)}

		require.NoError(t, w.MoveSailToCenter(0))
		assert.Equal(t, top, w.GetSails()[0])
		assert.Equal(t, 1, w.GetStockCount())
	})

	t.Run("leaves the gap when nothing is left to refill it", func(t *testing.T) {
		w := newTestWindmill()
		clearWindmillBoard(w)
		windmillCenterUpTo(w, 1)
		fillWindmillSails(w)
		w.sails[5] = NewCard(CardDesignHeart, 2, true)

		require.NoError(t, w.MoveSailToCenter(5))
		assert.Nil(t, w.GetSails()[5])
	})

	t.Run("rejects an illegal rank, an empty sail and a bad index", func(t *testing.T) {
		w := newTestWindmill()
		clearWindmillBoard(w)
		windmillCenterUpTo(w, 1)
		w.sails[2] = NewCard(CardDesignHeart, 7, true)

		assert.Error(t, w.MoveSailToCenter(2))
		assert.Error(t, w.MoveSailToCenter(4))
		assert.Error(t, w.MoveSailToCenter(-1))
		assert.Error(t, w.MoveSailToCenter(WindmillSailCnt))
	})

	t.Run("errors once the game has ended", func(t *testing.T) {
		w := newTestWindmill()
		w.GiveUp()
		assert.Error(t, w.MoveSailToCenter(0))
	})
}

func TestWindmill_MoveSailToCorner(t *testing.T) {
	t.Run("opens a corner with a King and refills the gap", func(t *testing.T) {
		w := newTestWindmill()
		clearWindmillBoard(w)
		fillWindmillSails(w)
		w.sails[1] = NewCard(CardDesignSpade, CardValueMax, true)
		refill := NewCard(CardDesignHeart, 5, true)
		w.waste = []*Card{refill}

		require.NoError(t, w.MoveSailToCorner(1, 2))
		assert.Len(t, w.GetCorners()[2], 1)
		assert.Equal(t, refill, w.GetSails()[1])
	})

	t.Run("rejects an illegal rank, an empty sail and bad indices", func(t *testing.T) {
		w := newTestWindmill()
		clearWindmillBoard(w)
		w.sails[1] = NewCard(CardDesignSpade, 5, true)

		assert.Error(t, w.MoveSailToCorner(1, 0))
		assert.Error(t, w.MoveSailToCorner(0, 0))
		assert.Error(t, w.MoveSailToCorner(-1, 0))
		assert.Error(t, w.MoveSailToCorner(1, -1))
		assert.Error(t, w.MoveSailToCorner(1, WindmillCornerCnt))
	})

	t.Run("errors once the game has ended", func(t *testing.T) {
		w := newTestWindmill()
		w.GiveUp()
		assert.Error(t, w.MoveSailToCorner(0, 0))
	})
}

func TestWindmill_MoveWaste(t *testing.T) {
	t.Run("to the centre", func(t *testing.T) {
		w := newTestWindmill()
		clearWindmillBoard(w)
		windmillCenterUpTo(w, 1)
		w.waste = []*Card{NewCard(CardDesignHeart, 2, true)}

		require.NoError(t, w.MoveWasteToCenter())
		assert.Len(t, w.GetCenter(), 2)
		assert.Empty(t, w.GetWaste())
	})

	t.Run("to a corner", func(t *testing.T) {
		w := newTestWindmill()
		clearWindmillBoard(w)
		w.waste = []*Card{NewCard(CardDesignHeart, CardValueMax, true)}

		require.NoError(t, w.MoveWasteToCorner(3))
		assert.Len(t, w.GetCorners()[3], 1)
		assert.Empty(t, w.GetWaste())
	})

	t.Run("rejects an empty waste, an illegal rank and a bad index", func(t *testing.T) {
		w := newTestWindmill()
		clearWindmillBoard(w)
		windmillCenterUpTo(w, 1)

		assert.Error(t, w.MoveWasteToCenter())
		assert.Error(t, w.MoveWasteToCorner(0))

		w.waste = []*Card{NewCard(CardDesignHeart, 9, true)}
		assert.Error(t, w.MoveWasteToCenter())
		assert.Error(t, w.MoveWasteToCorner(0))
		assert.Error(t, w.MoveWasteToCorner(-1))
		assert.Error(t, w.MoveWasteToCorner(WindmillCornerCnt))
	})

	t.Run("errors once the game has ended", func(t *testing.T) {
		w := newTestWindmill()
		w.GiveUp()
		assert.Error(t, w.MoveWasteToCenter())
		assert.Error(t, w.MoveWasteToCorner(0))
	})
}

// The rescue transfer runs one way only -- corner to centre. #4416 describes it
// as bidirectional, which would let the player shuttle a card back and forth.
func TestWindmill_MoveCornerToCenter(t *testing.T) {
	t.Run("moves the corner top onto the centre", func(t *testing.T) {
		w := newTestWindmill()
		clearWindmillBoard(w)
		windmillCenterUpTo(w, 1)
		w.corners[0] = []*Card{
			NewCard(CardDesignSpade, CardValueMax, true),
			NewCard(CardDesignSpade, CardValueMax-1, true),
			NewCard(CardDesignHeart, 2, true),
		}

		require.NoError(t, w.MoveCornerToCenter(0))
		assert.Len(t, w.GetCenter(), 2)
		assert.Len(t, w.GetCorners()[0], 2)
		assert.True(t, w.IsTransferBlocked())
	})

	// Without this restriction a descending corner could be poured wholesale
	// into the ascending centre, which trivialises the game.
	t.Run("refuses two transfers in a row", func(t *testing.T) {
		w := newTestWindmill()
		clearWindmillBoard(w)
		windmillCenterUpTo(w, 1)
		w.corners[0] = []*Card{NewCard(CardDesignSpade, CardValueMax, true), NewCard(CardDesignHeart, 2, true)}
		w.corners[1] = []*Card{NewCard(CardDesignClover, CardValueMax, true), NewCard(CardDesignHeart, 3, true)}

		require.NoError(t, w.MoveCornerToCenter(0))
		// A different corner is still the same restricted move.
		assert.Error(t, w.MoveCornerToCenter(1))
	})

	t.Run("a sail card played to the centre lifts the block", func(t *testing.T) {
		w := newTestWindmill()
		clearWindmillBoard(w)
		windmillCenterUpTo(w, 1)
		w.corners[0] = []*Card{NewCard(CardDesignSpade, CardValueMax, true), NewCard(CardDesignHeart, 2, true)}
		w.corners[1] = []*Card{NewCard(CardDesignClover, CardValueMax, true), NewCard(CardDesignHeart, 4, true)}
		w.sails[0] = NewCard(CardDesignDiamond, 3, true)

		require.NoError(t, w.MoveCornerToCenter(0))
		require.True(t, w.IsTransferBlocked())
		require.NoError(t, w.MoveSailToCenter(0))
		assert.False(t, w.IsTransferBlocked())
		assert.NoError(t, w.MoveCornerToCenter(1))
	})

	t.Run("a waste card played to the centre lifts the block", func(t *testing.T) {
		w := newTestWindmill()
		clearWindmillBoard(w)
		windmillCenterUpTo(w, 1)
		w.corners[0] = []*Card{NewCard(CardDesignSpade, CardValueMax, true), NewCard(CardDesignHeart, 2, true)}
		w.waste = []*Card{NewCard(CardDesignDiamond, 3, true)}

		require.NoError(t, w.MoveCornerToCenter(0))
		require.NoError(t, w.MoveWasteToCenter())
		assert.False(t, w.IsTransferBlocked())
	})

	// A corner card sent to a corner is not a centre placement, so it must not
	// clear the block.
	t.Run("a move that misses the centre leaves the block up", func(t *testing.T) {
		w := newTestWindmill()
		clearWindmillBoard(w)
		windmillCenterUpTo(w, 1)
		w.corners[0] = []*Card{NewCard(CardDesignSpade, CardValueMax, true), NewCard(CardDesignHeart, 2, true)}
		w.sails[0] = NewCard(CardDesignDiamond, CardValueMax, true)

		require.NoError(t, w.MoveCornerToCenter(0))
		require.NoError(t, w.MoveSailToCorner(0, 1))
		assert.True(t, w.IsTransferBlocked())
	})

	t.Run("rejects an empty corner, an illegal rank and a bad index", func(t *testing.T) {
		w := newTestWindmill()
		clearWindmillBoard(w)
		windmillCenterUpTo(w, 1)

		assert.Error(t, w.MoveCornerToCenter(0))
		w.corners[0] = []*Card{NewCard(CardDesignSpade, CardValueMax, true)}
		assert.Error(t, w.MoveCornerToCenter(0))
		assert.Error(t, w.MoveCornerToCenter(-1))
		assert.Error(t, w.MoveCornerToCenter(WindmillCornerCnt))
	})

	t.Run("errors once the game has ended", func(t *testing.T) {
		w := newTestWindmill()
		w.GiveUp()
		assert.Error(t, w.MoveCornerToCenter(0))
	})
}

func TestWindmill_GetHint(t *testing.T) {
	t.Run("prefers a sail to the centre", func(t *testing.T) {
		w := newTestWindmill()
		clearWindmillBoard(w)
		windmillCenterUpTo(w, 1)
		w.sails[4] = NewCard(CardDesignHeart, 2, true)

		h := w.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "sail", h.FromZone)
		assert.Equal(t, 4, h.FromIdx)
		assert.Equal(t, "center", h.ToZone)
	})

	t.Run("offers a sail to a corner", func(t *testing.T) {
		w := newTestWindmill()
		clearWindmillBoard(w)
		windmillCenterUpTo(w, 1)
		w.sails[2] = NewCard(CardDesignHeart, CardValueMax, true)

		h := w.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "sail", h.FromZone)
		assert.Equal(t, "corner", h.ToZone)
		assert.Equal(t, 0, h.ToIdx)
	})

	t.Run("offers the waste when the sails are stuck", func(t *testing.T) {
		w := newTestWindmill()
		clearWindmillBoard(w)
		windmillCenterUpTo(w, 1)
		w.sails[0] = NewCard(CardDesignHeart, 9, true)
		w.waste = []*Card{NewCard(CardDesignSpade, 2, true)}

		h := w.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "waste", h.FromZone)
		assert.Equal(t, "center", h.ToZone)
	})

	// The transfer is a rescue that dismantles a finished corner, so it must not
	// be suggested while an ordinary move exists.
	t.Run("suggests the transfer only after ordinary moves run out", func(t *testing.T) {
		w := newTestWindmill()
		clearWindmillBoard(w)
		windmillCenterUpTo(w, 1)
		w.corners[0] = []*Card{NewCard(CardDesignSpade, CardValueMax, true), NewCard(CardDesignHeart, 2, true)}
		w.sails[0] = NewCard(CardDesignDiamond, 2, true)

		h := w.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "sail", h.FromZone)

		w.sails[0] = nil
		h = w.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "corner", h.FromZone)
		assert.Equal(t, "center", h.ToZone)
	})

	t.Run("does not suggest a blocked transfer", func(t *testing.T) {
		w := newTestWindmill()
		clearWindmillBoard(w)
		windmillCenterUpTo(w, 1)
		w.corners[0] = []*Card{NewCard(CardDesignSpade, CardValueMax, true), NewCard(CardDesignHeart, 2, true)}
		w.transferBlocked = true

		assert.Nil(t, w.GetHint())
	})

	t.Run("falls back to drawing", func(t *testing.T) {
		w := newTestWindmill()
		clearWindmillBoard(w)
		windmillCenterUpTo(w, 1)
		w.stock = []*Card{NewCard(CardDesignHeart, 9, true)}

		h := w.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "stock", h.FromZone)
		assert.Equal(t, "waste", h.ToZone)
	})

	t.Run("returns nil once the game has ended", func(t *testing.T) {
		w := newTestWindmill()
		w.GiveUp()
		assert.Nil(t, w.GetHint())
	})
}

func TestWindmill_Stalemate(t *testing.T) {
	t.Run("a dead board is a stalemate", func(t *testing.T) {
		w := newTestWindmill()
		clearWindmillBoard(w)
		windmillCenterUpTo(w, 1)
		// Every sail is a 9, the centre wants a 2, the corners want a King, and
		// the stock is gone.
		for i := range WindmillSailCnt {
			w.sails[i] = NewCard(CardDesignHeart, 9, true)
		}
		w.checkStalemate()
		assert.True(t, w.IsStalemate())
	})

	t.Run("a board with one legal move is not", func(t *testing.T) {
		w := newTestWindmill()
		clearWindmillBoard(w)
		windmillCenterUpTo(w, 1)
		for i := range WindmillSailCnt {
			w.sails[i] = NewCard(CardDesignHeart, 9, true)
		}
		w.sails[7] = NewCard(CardDesignHeart, CardValueMax, true)
		w.checkStalemate()
		assert.False(t, w.IsStalemate())
	})
}

func TestWindmill_AutoComplete(t *testing.T) {
	t.Run("sends every reachable card to the foundations", func(t *testing.T) {
		w := newTestWindmill()
		clearWindmillBoard(w)
		windmillCenterUpTo(w, 1)
		w.sails[0] = NewCard(CardDesignHeart, 2, true)
		w.sails[1] = NewCard(CardDesignSpade, 3, true)
		w.sails[2] = NewCard(CardDesignClover, 4, true)

		require.NoError(t, w.AutoComplete())
		assert.Len(t, w.GetCenter(), 4)
	})

	// Auto-complete must never dismantle a corner: pulling a card back is a
	// strategic rescue, not a mechanical tidy-up.
	t.Run("never performs the corner transfer", func(t *testing.T) {
		w := newTestWindmill()
		clearWindmillBoard(w)
		windmillCenterUpTo(w, 1)
		w.corners[0] = []*Card{NewCard(CardDesignSpade, CardValueMax, true), NewCard(CardDesignHeart, 2, true)}

		assert.Error(t, w.AutoComplete())
		assert.Len(t, w.GetCorners()[0], 2)
		assert.Len(t, w.GetCenter(), 1)
		assert.False(t, w.IsTransferBlocked())
	})

	t.Run("errors when nothing can move", func(t *testing.T) {
		w := newTestWindmill()
		clearWindmillBoard(w)
		windmillCenterUpTo(w, 1)
		w.sails[0] = NewCard(CardDesignHeart, 9, true)
		assert.Error(t, w.AutoComplete())
	})

	t.Run("errors once the game has ended", func(t *testing.T) {
		w := newTestWindmill()
		w.GiveUp()
		assert.Error(t, w.AutoComplete())
	})
	t.Run("drains the waste too", func(t *testing.T) {
		w := newTestWindmill()
		clearWindmillBoard(w)
		windmillCenterUpTo(w, 1)
		fillWindmillSails(w)
		// Sail 0 goes to the centre, its gap pulls the 3 up from the waste, and
		// that 3 then follows onto the centre -- so the waste drains as a cascade.
		w.sails[0] = NewCard(CardDesignHeart, 2, true)
		w.waste = []*Card{NewCard(CardDesignClover, 3, true)}

		require.NoError(t, w.AutoComplete())
		assert.Len(t, w.GetCenter(), 3)
		assert.Empty(t, w.GetWaste())
	})

	t.Run("sends a waste King to a corner", func(t *testing.T) {
		w := newTestWindmill()
		clearWindmillBoard(w)
		windmillCenterUpTo(w, 1)
		fillWindmillSails(w)
		w.waste = []*Card{NewCard(CardDesignClover, CardValueMax, true)}

		require.NoError(t, w.AutoComplete())
		assert.Len(t, w.GetCorners()[0], 1)
		assert.Empty(t, w.GetWaste())
	})
}

// A nil card is never placeable -- the sail slots are nilable, so every
// placement check has to survive being handed one.
func TestWindmill_NilCardIsNeverPlaceable(t *testing.T) {
	w := newTestWindmill()
	clearWindmillBoard(w)
	assert.False(t, w.canPlaceOnCenter(nil))
	assert.False(t, w.canPlaceOnCorner(nil, 0))
	assert.Equal(t, -1, w.findCorner(nil))
}

func TestWindmill_GameClear(t *testing.T) {
	w := newTestWindmill()
	clearWindmillBoard(w)
	windmillCenterUpTo(w, WindmillCenterTarget-1)
	for i := range WindmillCornerCnt {
		w.corners[i] = nil
		for v := CardValueMax; v >= 1; v-- {
			w.corners[i] = append(w.corners[i], NewCard(CardDesignSpade, v, true))
		}
	}
	w.sails[0] = NewCard(CardDesignHeart, CardValueMax, true)

	require.NoError(t, w.MoveSailToCenter(0))
	assert.Equal(t, WindmillPhaseGameClear, w.GetPhase())
	assert.True(t, w.GetGameEndFlag())
}

func TestWindmill_GiveUp(t *testing.T) {
	w := newTestWindmill()
	w.GiveUp()
	assert.Equal(t, WindmillPhaseGameOver, w.GetPhase())
	assert.True(t, w.GetGameEndFlag())
	require.NotEmpty(t, w.GetActionLog())

	// A second give-up must not append another entry.
	before := len(w.GetActionLog())
	w.GiveUp()
	assert.Len(t, w.GetActionLog(), before)
}

func TestWindmill_Undo(t *testing.T) {
	t.Run("restores the previous position", func(t *testing.T) {
		w := newTestWindmill()
		clearWindmillBoard(w)
		windmillCenterUpTo(w, 1)
		w.sails[0] = NewCard(CardDesignHeart, 2, true)

		require.NoError(t, w.MoveSailToCenter(0))
		require.True(t, w.CanUndo())
		require.NoError(t, w.Undo())

		assert.Len(t, w.GetCenter(), 1)
		assert.NotNil(t, w.GetSails()[0])
		assert.Equal(t, 0, w.GetMoveCount())
	})

	t.Run("restores the transfer block", func(t *testing.T) {
		w := newTestWindmill()
		clearWindmillBoard(w)
		windmillCenterUpTo(w, 1)
		w.corners[0] = []*Card{NewCard(CardDesignSpade, CardValueMax, true), NewCard(CardDesignHeart, 2, true)}
		w.sails[0] = NewCard(CardDesignDiamond, 3, true)

		require.NoError(t, w.MoveCornerToCenter(0))
		require.NoError(t, w.MoveSailToCenter(0))
		require.False(t, w.IsTransferBlocked())
		require.NoError(t, w.Undo())
		assert.True(t, w.IsTransferBlocked())
	})

	t.Run("errors with no history", func(t *testing.T) {
		w := newTestWindmill()
		clearWindmillBoard(w)
		w.history = nil
		assert.False(t, w.CanUndo())
		assert.Error(t, w.Undo())
	})
}

func TestWindmill_UndoN(t *testing.T) {
	w := newTestWindmill()
	clearWindmillBoard(w)
	windmillCenterUpTo(w, 1)
	w.sails[0] = NewCard(CardDesignHeart, 2, true)
	w.sails[1] = NewCard(CardDesignSpade, 3, true)

	require.NoError(t, w.MoveSailToCenter(0))
	require.NoError(t, w.MoveSailToCenter(1))
	require.NoError(t, w.UndoN(2))
	assert.Len(t, w.GetCenter(), 1)
	assert.Equal(t, 0, w.GetMoveCount())

	assert.Error(t, w.UndoN(0))
	assert.Error(t, w.UndoN(-1))
	assert.Error(t, w.UndoN(99))
}

func TestWindmill_UndoToEscape(t *testing.T) {
	t.Run("zero when not stuck", func(t *testing.T) {
		w := newTestWindmill()
		assert.Equal(t, 0, w.UndoToEscape())
	})

	t.Run("counts back to the last live position", func(t *testing.T) {
		w := newTestWindmill()
		clearWindmillBoard(w)
		windmillCenterUpTo(w, 1)
		w.sails[0] = NewCard(CardDesignHeart, 2, true)
		w.checkStalemate()
		require.False(t, w.IsStalemate())

		require.NoError(t, w.MoveSailToCenter(0))
		require.True(t, w.IsStalemate())
		assert.Equal(t, 1, w.UndoToEscape())
	})

	t.Run("-1 when every recorded position was already stuck", func(t *testing.T) {
		w := newTestWindmill()
		clearWindmillBoard(w)
		w.isStalemate = true
		w.history = []*windmillSnapshot{{isStalemate: true}}
		assert.Equal(t, -1, w.UndoToEscape())
	})
}

func TestWindmill_ActionLog(t *testing.T) {
	w := newTestWindmill()
	clearWindmillBoard(w)
	windmillCenterUpTo(w, 1)
	fillWindmillSails(w)
	w.sails[0] = NewCard(CardDesignHeart, 2, true)
	w.sails[1] = NewCard(CardDesignSpade, CardValueMax, true)
	w.corners[3] = []*Card{NewCard(CardDesignClover, CardValueMax, true), NewCard(CardDesignHeart, 3, true)}
	// Three: the two sail moves each pull one card in to refill their gap, and
	// the Draw at the end needs one of its own.
	w.stock = []*Card{
		NewCard(CardDesignDiamond, 7, true),
		NewCard(CardDesignDiamond, 8, true),
		NewCard(CardDesignDiamond, 9, true),
	}

	require.NoError(t, w.MoveSailToCenter(0))
	require.NoError(t, w.MoveSailToCorner(1, 0))
	require.NoError(t, w.MoveCornerToCenter(3))
	require.NoError(t, w.Draw())

	// The board is 0-indexed everywhere, so the log must be too -- a 1-based log
	// silently disagrees with the hint and the CLI.
	details := make([]string, 0, len(w.GetActionLog()))
	for _, e := range w.GetActionLog() {
		details = append(details, e.Detail)
	}
	assert.Equal(t, []string{
		"帆0→中央基礎",
		"帆1→四隅基礎0",
		"四隅基礎3→中央基礎",
		"山札から1枚めくった",
	}, details)
}

func TestWindmill_JSONRoundTrip(t *testing.T) {
	w := newTestWindmill()
	require.NoError(t, w.Draw())

	data, err := json.Marshal(w)
	require.NoError(t, err)

	restored := NewDefaultWindmill()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, w.GetPhase(), restored.GetPhase())
	assert.Equal(t, w.GetMoveCount(), restored.GetMoveCount())
	assert.Equal(t, w.GetStockCount(), restored.GetStockCount())
	assert.Len(t, restored.GetWaste(), len(w.GetWaste()))
	assert.Len(t, restored.GetCenter(), len(w.GetCenter()))
	assert.Equal(t, w.IsTransferBlocked(), restored.IsTransferBlocked())
}

// KV can hold anything an earlier version wrote, so a corrupt snapshot must be
// refused rather than started.
func TestWindmill_UnmarshalJSONValidation(t *testing.T) {
	overlong := make([]*Card, WindmillCenterTarget+1)
	for i := range overlong {
		overlong[i] = NewCard(CardDesignSpade, 1, true)
	}
	tooManyCorner := make([]*Card, WindmillCornerTarget+1)
	for i := range tooManyCorner {
		tooManyCorner[i] = NewCard(CardDesignSpade, 1, true)
	}
	hugeStock := make([]*Card, WindmillTotalCards+1)
	for i := range hugeStock {
		hugeStock[i] = NewCard(CardDesignSpade, 1, true)
	}

	for _, tc := range []struct {
		name string
		j    windmillJSON
	}{
		{"phase below range", windmillJSON{Phase: -1}},
		{"phase above range", windmillJSON{Phase: WindmillPhaseGameOver + 1}},
		{"negative move count", windmillJSON{MoveCount: -1}},
		{"center overflows", windmillJSON{Center: overlong}},
		{"stock overflows", windmillJSON{Stock: hugeStock}},
		{"waste overflows", windmillJSON{Waste: hugeStock}},
		{"corner overflows", windmillJSON{Corners: [WindmillCornerCnt][]*Card{tooManyCorner}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(&tc.j)
			require.NoError(t, err)
			assert.Error(t, NewDefaultWindmill().UnmarshalJSON(data))
		})
	}

	t.Run("malformed json", func(t *testing.T) {
		assert.Error(t, NewDefaultWindmill().UnmarshalJSON([]byte("not json")))
	})

	t.Run("a valid snapshot is accepted", func(t *testing.T) {
		data, err := json.Marshal(&windmillJSON{Phase: WindmillPhasePlaying, MoveCount: 3})
		require.NoError(t, err)
		w := NewDefaultWindmill()
		require.NoError(t, w.UnmarshalJSON(data))
		assert.Equal(t, 3, w.GetMoveCount())
	})
}

// Drive a full game to make sure no path panics and the invariants hold
// throughout. The deal is random, so this is a fuzz-ish smoke test rather than
// an assertion about any particular position.
func TestWindmill_FullGameDrive(t *testing.T) {
	for range 30 {
		w := newTestWindmill()
		for range 500 {
			if w.GetGameEndFlag() {
				break
			}
			h := w.GetHint()
			if h == nil {
				break
			}
			var err error
			switch {
			case h.FromZone == "stock":
				err = w.Draw()
			case h.FromZone == "corner":
				err = w.MoveCornerToCenter(h.FromIdx)
			case h.FromZone == "sail" && h.ToZone == "center":
				err = w.MoveSailToCenter(h.FromIdx)
			case h.FromZone == "sail":
				err = w.MoveSailToCorner(h.FromIdx, h.ToIdx)
			case h.ToZone == "center":
				err = w.MoveWasteToCenter()
			default:
				err = w.MoveWasteToCorner(h.ToIdx)
			}
			require.NoError(t, err)

			assert.LessOrEqual(t, len(w.GetCenter()), WindmillCenterTarget)
			for i := range WindmillCornerCnt {
				assert.LessOrEqual(t, len(w.GetCorners()[i]), WindmillCornerTarget)
			}
		}
	}
}
