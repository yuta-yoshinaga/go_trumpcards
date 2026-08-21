//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// --- helpers ---

func newTestWhitehead() *domain.Whitehead {
	tc := domain.NewTrumpCards(0)
	k := domain.NewWhitehead(tc)
	return k
}

func setupPlayingWhitehead() *domain.Whitehead {
	k := newTestWhitehead()
	k.Reset()
	return k
}

func makeCardWH(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

func makeTableauCardWH(design, value int, faceUp bool) *domain.WhiteheadTableauCard {
	return &domain.WhiteheadTableauCard{Card: makeCardWH(design, value), FaceUp: faceUp}
}

// clearTableauWH clears all tableau columns
func clearTableauWH(k *domain.Whitehead) {
	var empty [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
	k.SetTableau(empty)
}

func TestNewWhitehead(t *testing.T) {
	k := newTestWhitehead()
	assert.NotNil(t, k)
	assert.Equal(t, domain.WhiteheadPhase(0), k.GetPhase())
}

func TestWhitehead_Reset(t *testing.T) {
	k := setupPlayingWhitehead()

	assert.Equal(t, domain.WhiteheadPhasePlaying, k.GetPhase())
	assert.Equal(t, 0, k.GetMoveCount())

	// Tableau: 7 columns, col i has i+1 cards
	tableau := k.GetTableau()
	totalTableauCards := 0
	for i := 0; i < domain.WhiteheadTableauCnt; i++ {
		assert.Equal(t, i+1, len(tableau[i]), "column %d should have %d cards", i, i+1)
		// **Every card is face up.** Klondike, which this was cloned from, leaves
		// all but the top card of each column face down; Whitehead is an open
		// game, so there is nothing hidden and nothing to flip.
		for j, tc := range tableau[i] {
			assert.True(t, tc.FaceUp, "column %d card %d should be face up", i, j)
		}
		totalTableauCards += len(tableau[i])
	}
	assert.Equal(t, 28, totalTableauCards)

	// Stock: 24 cards
	assert.Equal(t, 24, k.GetStockCount())

	// Waste: empty
	assert.Nil(t, k.GetWaste())

	// Foundation: empty
	foundation := k.GetFoundation()
	for i := 0; i < domain.WhiteheadFoundationCnt; i++ {
		assert.Nil(t, foundation[i])
	}
}

func TestWhitehead_Draw(t *testing.T) {
	t.Run("draw from stock", func(t *testing.T) {
		k := setupPlayingWhitehead()
		initialStock := k.GetStockCount()
		err := k.Draw()
		assert.NoError(t, err)
		assert.Equal(t, initialStock-1, k.GetStockCount())
		assert.Equal(t, 1, len(k.GetWaste()))
		assert.Equal(t, 1, k.GetMoveCount())
	})

	t.Run("recycle waste to stock", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		// Draw all stock cards
		for k.GetStockCount() > 0 {
			_ = k.Draw()
		}
		wasteLen := len(k.GetWaste())
		assert.Greater(t, wasteLen, 0)

		// Now drawing recycles waste to stock
		err := k.Draw()
		assert.NoError(t, err)
		assert.Equal(t, wasteLen, k.GetStockCount())
		assert.Nil(t, k.GetWaste())
	})

	t.Run("error when both stock and waste empty", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		k.SetStock(nil)
		k.SetWaste(nil)
		err := k.Draw()
		assert.Error(t, err)
		assert.Equal(t, "no cards in stock or waste", err.Error())
	})

	t.Run("error when not playing", func(t *testing.T) {
		k := setupPlayingWhitehead()
		k.SetPhase(domain.WhiteheadPhaseGameClear)
		err := k.Draw()
		assert.Error(t, err)
		assert.Equal(t, "game is not in playing phase", err.Error())
	})
}

func TestWhitehead_MoveWasteToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		// Whitehead builds by SAME colour: a black 6 onto a black 7.
		k.SetWaste([]*domain.Card{makeCardWH(domain.CardDesignClover, 6)})
		var tab [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
		tab[0] = []*domain.WhiteheadTableauCard{makeTableauCardWH(domain.CardDesignSpade, 7, true)}
		k.SetTableau(tab)
		err := k.MoveWasteToTableau(0)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(k.GetWaste()))
		assert.Equal(t, 2, len(k.GetTableau()[0]))
	})

	t.Run("K on empty column", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		k.SetWaste([]*domain.Card{makeCardWH(domain.CardDesignSpade, 13)})
		var tab [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
		k.SetTableau(tab)
		err := k.MoveWasteToTableau(0)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(k.GetTableau()[0]))
	})

	t.Run("any card may take an empty column", func(t *testing.T) {
		// Klondike restricted this to a King; Whitehead does not.
		k := newTestWhitehead()
		k.Reset()
		k.SetWaste([]*domain.Card{makeCardWH(domain.CardDesignSpade, 5)})
		var tab [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
		k.SetTableau(tab)
		assert.NoError(t, k.MoveWasteToTableau(0))
		assert.Equal(t, 1, len(k.GetTableau()[0]))
	})

	t.Run("error: alternating colour", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		k.SetWaste([]*domain.Card{makeCardWH(domain.CardDesignHeart, 6)})
		var tab [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
		tab[0] = []*domain.WhiteheadTableauCard{makeTableauCardWH(domain.CardDesignClover, 7, true)}
		k.SetTableau(tab)
		err := k.MoveWasteToTableau(0)
		assert.Error(t, err)
	})

	t.Run("error: wrong value", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		k.SetWaste([]*domain.Card{makeCardWH(domain.CardDesignHeart, 5)})
		var tab [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
		tab[0] = []*domain.WhiteheadTableauCard{makeTableauCardWH(domain.CardDesignSpade, 7, true)}
		k.SetTableau(tab)
		err := k.MoveWasteToTableau(0)
		assert.Error(t, err)
	})

	t.Run("error: invalid column", func(t *testing.T) {
		k := setupPlayingWhitehead()
		k.SetWaste([]*domain.Card{makeCardWH(domain.CardDesignSpade, 1)})
		assert.Error(t, k.MoveWasteToTableau(-1))
		assert.Error(t, k.MoveWasteToTableau(7))
	})

	t.Run("error: waste empty", func(t *testing.T) {
		k := setupPlayingWhitehead()
		k.SetWaste(nil)
		err := k.MoveWasteToTableau(0)
		assert.Error(t, err)
		assert.Equal(t, "waste is empty", err.Error())
	})

	t.Run("error: not playing", func(t *testing.T) {
		k := setupPlayingWhitehead()
		k.SetPhase(domain.WhiteheadPhaseGameOver)
		err := k.MoveWasteToTableau(0)
		assert.Error(t, err)
	})
}

func TestWhitehead_MoveWasteToFoundation(t *testing.T) {
	t.Run("success ace", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		k.SetWaste([]*domain.Card{makeCardWH(domain.CardDesignSpade, 1)})
		err := k.MoveWasteToFoundation()
		assert.NoError(t, err)
		assert.Equal(t, 0, len(k.GetWaste()))
		assert.Equal(t, 1, len(k.GetFoundation()[0])) // Spade = design 1, index 0
	})

	t.Run("success sequential", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		var f [domain.WhiteheadFoundationCnt][]*domain.Card
		f[0] = []*domain.Card{makeCardWH(domain.CardDesignSpade, 1)}
		k.SetFoundation(f)
		k.SetWaste([]*domain.Card{makeCardWH(domain.CardDesignSpade, 2)})
		err := k.MoveWasteToFoundation()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(k.GetFoundation()[0]))
	})

	t.Run("error: wrong sequence", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		var f [domain.WhiteheadFoundationCnt][]*domain.Card
		f[0] = []*domain.Card{makeCardWH(domain.CardDesignSpade, 1)}
		k.SetFoundation(f)
		k.SetWaste([]*domain.Card{makeCardWH(domain.CardDesignSpade, 3)})
		err := k.MoveWasteToFoundation()
		assert.Error(t, err)
	})

	t.Run("error: not ace on empty foundation", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		k.SetWaste([]*domain.Card{makeCardWH(domain.CardDesignSpade, 2)})
		err := k.MoveWasteToFoundation()
		assert.Error(t, err)
	})

	t.Run("error: waste empty", func(t *testing.T) {
		k := setupPlayingWhitehead()
		k.SetWaste(nil)
		err := k.MoveWasteToFoundation()
		assert.Error(t, err)
	})

	t.Run("error: not playing", func(t *testing.T) {
		k := setupPlayingWhitehead()
		k.SetPhase(domain.WhiteheadPhaseGameOver)
		err := k.MoveWasteToFoundation()
		assert.Error(t, err)
	})

	t.Run("error: joker card", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		k.SetWaste([]*domain.Card{makeCardWH(domain.CardDesignJoker, 0)})
		err := k.MoveWasteToFoundation()
		assert.Error(t, err)
		assert.Equal(t, "invalid card for foundation", err.Error())
	})

	t.Run("error: design out of range", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		k.SetWaste([]*domain.Card{makeCardWH(domain.CardDesignDiamond+1, 1)})
		err := k.MoveWasteToFoundation()
		assert.Error(t, err)
		assert.Equal(t, "invalid card for foundation", err.Error())
	})
}

// twoColWH puts one card in column 0 and one in column 1, everything else empty.
func twoColWH(t *testing.T, fromD, fromV, toD, toV int) *domain.Whitehead {
	t.Helper()
	k := newTestWhitehead()
	k.Reset()
	var tab [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
	tab[0] = []*domain.WhiteheadTableauCard{makeTableauCardWH(fromD, fromV, true)}
	tab[1] = []*domain.WhiteheadTableauCard{makeTableauCardWH(toD, toV, true)}
	k.SetTableau(tab)
	return k
}

// whiteheadDeadTableau fills ALL seven columns with same-rank cards.
//
// Whitehead lets any card take an empty column, so a fixture that leaves even
// one column empty always has a legal move -- Klondike's versions filled two of
// seven and leaned on the King-only rule to keep the rest shut. Same rank never
// connects in either direction, and no Ace is present, so the empty foundations
// stay closed too.
func whiteheadDeadTableau() [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard {
	designs := []int{
		domain.CardDesignSpade, domain.CardDesignClover,
		domain.CardDesignHeart, domain.CardDesignDiamond,
	}
	var tab [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
	for i := 0; i < domain.WhiteheadTableauCnt; i++ {
		tab[i] = []*domain.WhiteheadTableauCard{
			makeTableauCardWH(designs[i%len(designs)], 5, true),
		}
	}
	return tab
}

// Whitehead builds down by SAME COLOUR. Klondike -- the domain this was cloned
// from -- builds down by ALTERNATING colour, and only a King may take an empty
// column. Each half gets a test plus a negative control.
func TestWhitehead_TableauRule(t *testing.T) {
	t.Run("same colour descending is legal", func(t *testing.T) {
		k := twoColWH(t, domain.CardDesignSpade, 6, domain.CardDesignClover, 7)
		assert.NoError(t, k.MoveTableauToTableau(0, 0, 1))
		assert.Equal(t, 2, len(k.GetTableau()[1]))
	})

	t.Run("alternating colour is REJECTED", func(t *testing.T) {
		k := twoColWH(t, domain.CardDesignHeart, 6, domain.CardDesignSpade, 7)
		assert.Error(t, k.MoveTableauToTableau(0, 0, 1))
	})

	t.Run("same colour but not adjacent is rejected", func(t *testing.T) {
		k := twoColWH(t, domain.CardDesignSpade, 5, domain.CardDesignClover, 7)
		assert.Error(t, k.MoveTableauToTableau(0, 0, 1))
	})

	t.Run("ascending is rejected", func(t *testing.T) {
		k := twoColWH(t, domain.CardDesignSpade, 8, domain.CardDesignClover, 7)
		assert.Error(t, k.MoveTableauToTableau(0, 0, 1))
	})

	t.Run("any card may take an empty column", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		var tab [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
		tab[0] = []*domain.WhiteheadTableauCard{makeTableauCardWH(domain.CardDesignSpade, 5, true)}
		k.SetTableau(tab)
		assert.NoError(t, k.MoveTableauToTableau(0, 0, 1))
		assert.Equal(t, 1, len(k.GetTableau()[1]))
	})
}

// TestWhitehead_NothingToAutoFlip replaces Klondike's "auto-flip after move"
// cases. Whitehead deals all 28 tableau cards face up, so no move can expose a
// face-down card -- keeping those tests would have asserted a flip that can
// never happen.
func TestWhitehead_NothingToAutoFlip(t *testing.T) {
	k := setupPlayingWhitehead()
	for i := 0; i < domain.WhiteheadTableauCnt; i++ {
		for j, tc := range k.GetTableau()[i] {
			assert.True(t, tc.FaceUp, "column %d card %d must be face up", i, j)
		}
	}
	var tab [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
	tab[0] = []*domain.WhiteheadTableauCard{
		makeTableauCardWH(domain.CardDesignSpade, 9, true),
		makeTableauCardWH(domain.CardDesignClover, 6, true),
	}
	tab[1] = []*domain.WhiteheadTableauCard{makeTableauCardWH(domain.CardDesignSpade, 7, true)}
	k.SetTableau(tab)
	assert.NoError(t, k.MoveTableauToTableau(0, 1, 1))
	for j, tc := range k.GetTableau()[0] {
		assert.True(t, tc.FaceUp, "column 0 card %d stays face up after a move", j)
	}
}

func TestWhitehead_MoveTableauToTableau(t *testing.T) {
	t.Run("success single card", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		var tab [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
		// Same colour: Whitehead builds black-on-black / red-on-red.
		tab[0] = []*domain.WhiteheadTableauCard{makeTableauCardWH(domain.CardDesignClover, 6, true)}
		tab[1] = []*domain.WhiteheadTableauCard{makeTableauCardWH(domain.CardDesignSpade, 7, true)}
		k.SetTableau(tab)
		err := k.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(k.GetTableau()[0]))
		assert.Equal(t, 2, len(k.GetTableau()[1]))
	})

	t.Run("success multiple cards", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		var tab [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
		// A same-colour run moves as a group; the junction must be same-colour too.
		tab[0] = []*domain.WhiteheadTableauCard{
			makeTableauCardWH(domain.CardDesignHeart, 7, true),
			makeTableauCardWH(domain.CardDesignDiamond, 6, true),
		}
		tab[1] = []*domain.WhiteheadTableauCard{makeTableauCardWH(domain.CardDesignHeart, 8, true)}
		k.SetTableau(tab)
		err := k.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(k.GetTableau()[0]))
		assert.Equal(t, 3, len(k.GetTableau()[1]))
	})

	t.Run("success K to empty column", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		var tab [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
		tab[0] = []*domain.WhiteheadTableauCard{makeTableauCardWH(domain.CardDesignSpade, 13, true)}
		k.SetTableau(tab)
		err := k.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(k.GetTableau()[1]))
	})

	t.Run("error: face down card", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		var tab [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
		tab[0] = []*domain.WhiteheadTableauCard{makeTableauCardWH(domain.CardDesignHeart, 6, false)}
		tab[1] = []*domain.WhiteheadTableauCard{makeTableauCardWH(domain.CardDesignSpade, 7, true)}
		k.SetTableau(tab)
		err := k.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
		assert.Equal(t, "card is face down", err.Error())
	})

	t.Run("error: same column", func(t *testing.T) {
		k := setupPlayingWhitehead()
		err := k.MoveTableauToTableau(0, 0, 0)
		assert.Error(t, err)
	})

	t.Run("error: invalid from column", func(t *testing.T) {
		k := setupPlayingWhitehead()
		assert.Error(t, k.MoveTableauToTableau(-1, 0, 1))
		assert.Error(t, k.MoveTableauToTableau(7, 0, 1))
	})

	t.Run("error: invalid to column", func(t *testing.T) {
		k := setupPlayingWhitehead()
		assert.Error(t, k.MoveTableauToTableau(0, 0, -1))
		assert.Error(t, k.MoveTableauToTableau(0, 0, 7))
	})

	t.Run("error: invalid card index", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		var tab [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
		tab[0] = []*domain.WhiteheadTableauCard{makeTableauCardWH(domain.CardDesignHeart, 6, true)}
		k.SetTableau(tab)
		assert.Error(t, k.MoveTableauToTableau(0, -1, 1))
		assert.Error(t, k.MoveTableauToTableau(0, 1, 1))
	})

	t.Run("error: not playing", func(t *testing.T) {
		k := setupPlayingWhitehead()
		k.SetPhase(domain.WhiteheadPhaseGameOver)
		err := k.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})

	t.Run("error: cannot place", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		var tab [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
		tab[0] = []*domain.WhiteheadTableauCard{makeTableauCardWH(domain.CardDesignHeart, 6, true)}
		tab[1] = []*domain.WhiteheadTableauCard{makeTableauCardWH(domain.CardDesignSpade, 5, true)}
		k.SetTableau(tab)
		err := k.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})
}

func TestWhitehead_MoveTableauToFoundation(t *testing.T) {
	t.Run("success ace", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		var tab [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
		tab[0] = []*domain.WhiteheadTableauCard{makeTableauCardWH(domain.CardDesignSpade, 1, true)}
		k.SetTableau(tab)
		err := k.MoveTableauToFoundation(0)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(k.GetTableau()[0]))
		assert.Equal(t, 1, len(k.GetFoundation()[0]))
	})

	t.Run("error: empty column", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		clearTableauWH(k)
		err := k.MoveTableauToFoundation(0)
		assert.Error(t, err)
		assert.Equal(t, "tableau column is empty", err.Error())
	})

	t.Run("error: invalid column", func(t *testing.T) {
		k := setupPlayingWhitehead()
		assert.Error(t, k.MoveTableauToFoundation(-1))
		assert.Error(t, k.MoveTableauToFoundation(7))
	})

	t.Run("error: not playing", func(t *testing.T) {
		k := setupPlayingWhitehead()
		k.SetPhase(domain.WhiteheadPhaseGameClear)
		err := k.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("error: cannot place on foundation", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		var tab [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
		tab[0] = []*domain.WhiteheadTableauCard{makeTableauCardWH(domain.CardDesignSpade, 5, true)}
		k.SetTableau(tab)
		err := k.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("error: joker card", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		var tab [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
		tab[0] = []*domain.WhiteheadTableauCard{makeTableauCardWH(domain.CardDesignJoker, 0, true)}
		k.SetTableau(tab)
		err := k.MoveTableauToFoundation(0)
		assert.Error(t, err)
		assert.Equal(t, "invalid card for foundation", err.Error())
	})

	t.Run("error: design out of range", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		var tab [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
		tab[0] = []*domain.WhiteheadTableauCard{makeTableauCardWH(domain.CardDesignDiamond+1, 1, true)}
		k.SetTableau(tab)
		err := k.MoveTableauToFoundation(0)
		assert.Error(t, err)
		assert.Equal(t, "invalid card for foundation", err.Error())
	})

	t.Run("game clear", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		// Set up foundation with 12 cards each
		var f [domain.WhiteheadFoundationCnt][]*domain.Card
		for i := 0; i < domain.WhiteheadFoundationCnt; i++ {
			f[i] = make([]*domain.Card, 0)
			for v := 1; v <= 12; v++ {
				f[i] = append(f[i], makeCardWH(i+1, v))
			}
		}
		k.SetFoundation(f)
		// Set last cards on tableau
		var tab [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
		for i := 0; i < domain.WhiteheadFoundationCnt; i++ {
			tab[i] = []*domain.WhiteheadTableauCard{makeTableauCardWH(i+1, 13, true)}
		}
		k.SetTableau(tab)
		k.SetStock(nil)
		k.SetWaste(nil)

		for i := 0; i < domain.WhiteheadFoundationCnt; i++ {
			err := k.MoveTableauToFoundation(i)
			assert.NoError(t, err)
		}
		assert.Equal(t, domain.WhiteheadPhaseGameClear, k.GetPhase())
	})
}

func TestWhitehead_GiveUp(t *testing.T) {
	t.Run("give up during playing", func(t *testing.T) {
		k := setupPlayingWhitehead()
		k.GiveUp()
		assert.Equal(t, domain.WhiteheadPhaseGameOver, k.GetPhase())
	})

	t.Run("give up when already game over", func(t *testing.T) {
		k := setupPlayingWhitehead()
		k.SetPhase(domain.WhiteheadPhaseGameOver)
		k.GiveUp()
		assert.Equal(t, domain.WhiteheadPhaseGameOver, k.GetPhase())
	})

	t.Run("give up when game clear", func(t *testing.T) {
		k := setupPlayingWhitehead()
		k.SetPhase(domain.WhiteheadPhaseGameClear)
		k.GiveUp()
		assert.Equal(t, domain.WhiteheadPhaseGameClear, k.GetPhase())
	})
}

func TestWhitehead_GetHint(t *testing.T) {
	t.Run("no hint when not playing", func(t *testing.T) {
		k := setupPlayingWhitehead()
		k.SetPhase(domain.WhiteheadPhaseGameOver)
		assert.Nil(t, k.GetHint())
	})

	t.Run("hint: tableau to foundation", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		var tab [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
		tab[0] = []*domain.WhiteheadTableauCard{makeTableauCardWH(domain.CardDesignSpade, 1, true)}
		k.SetTableau(tab)
		k.SetWaste(nil)
		k.SetStock(nil)
		hint := k.GetHint()
		assert.NotNil(t, hint)
		assert.Equal(t, "tableau", hint.FromZone)
		assert.Equal(t, 0, hint.FromCol)
		assert.Equal(t, "foundation", hint.ToZone)
	})

	t.Run("hint: waste to foundation", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		clearTableauWH(k)
		k.SetWaste([]*domain.Card{makeCardWH(domain.CardDesignHeart, 1)})
		k.SetStock(nil)
		hint := k.GetHint()
		assert.NotNil(t, hint)
		assert.Equal(t, "waste", hint.FromZone)
		assert.Equal(t, "foundation", hint.ToZone)
	})

	// Klondike's "expose a face-down card" hint case is gone: Whitehead deals
	// every card face up, so there is never a hidden card to expose.

	t.Run("hint: waste to tableau", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		tab := whiteheadDeadTableau()
		tab[0] = []*domain.WhiteheadTableauCard{makeTableauCardWH(domain.CardDesignSpade, 7, true)}
		k.SetTableau(tab)
		// Same colour, and every other column is occupied, so column 0 is the
		// only target -- otherwise an empty column would win the hint.
		k.SetWaste([]*domain.Card{makeCardWH(domain.CardDesignClover, 6)})
		k.SetStock(nil)
		hint := k.GetHint()
		assert.NotNil(t, hint)
		assert.Equal(t, "waste", hint.FromZone)
		assert.Equal(t, "tableau", hint.ToZone)
		assert.Equal(t, 0, hint.ToCol)
	})

	t.Run("no hint available", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		k.SetTableau(whiteheadDeadTableau())
		// A 3 cannot join a column of 5s, and every column is occupied.
		k.SetWaste([]*domain.Card{makeCardWH(domain.CardDesignSpade, 3)})
		k.SetStock(nil)
		hint := k.GetHint()
		assert.Nil(t, hint)
	})

	t.Run("skip empty tableau columns in hint", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		clearTableauWH(k)
		k.SetWaste(nil)
		k.SetStock(nil)
		hint := k.GetHint()
		assert.Nil(t, hint)
	})

	t.Run("skip all face-up tableau in hint", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		var tab [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
		// All face up in col 0 - no face-down to expose, should skip
		// A same-colour run moves as a group; the junction must be same-colour too.
		tab[0] = []*domain.WhiteheadTableauCard{
			makeTableauCardWH(domain.CardDesignHeart, 7, true),
			makeTableauCardWH(domain.CardDesignDiamond, 6, true),
		}
		tab[1] = []*domain.WhiteheadTableauCard{makeTableauCardWH(domain.CardDesignHeart, 8, true)}
		k.SetTableau(tab)
		k.SetWaste(nil)
		k.SetStock(nil)
		hint := k.GetHint()
		// No hint for foundation, and tab[0] has no face-down cards to expose
		assert.Nil(t, hint)
	})

	t.Run("tableau to tableau: no valid target", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		// Every column occupied by a 5 and nothing connects. The Klondike version
		// left five columns empty and leaned on the King-only rule.
		k.SetTableau(whiteheadDeadTableau())
		k.SetWaste(nil)
		k.SetStock(nil)
		hint := k.GetHint()
		assert.Nil(t, hint)
	})
}

func TestWhitehead_AutoComplete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		// Set up: all cards face up in tableau, each column has cards from K down to A
		// so the top card (last in slice) is A and can be moved to foundation first
		var tab [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
		designs := []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond}
		for suit := 0; suit < 4; suit++ {
			tab[suit] = make([]*domain.WhiteheadTableauCard, 0)
			for v := 13; v >= 1; v-- {
				tab[suit] = append(tab[suit], makeTableauCardWH(designs[suit], v, true))
			}
		}
		k.SetTableau(tab)
		k.SetStock(nil)
		k.SetWaste(nil)
		err := k.AutoComplete()
		assert.NoError(t, err)
		assert.Equal(t, domain.WhiteheadPhaseGameClear, k.GetPhase())
		for i := 0; i < domain.WhiteheadFoundationCnt; i++ {
			assert.Equal(t, 13, len(k.GetFoundation()[i]))
		}
	})

	t.Run("success with waste", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		// Set up: all cards face up, some in waste
		k.SetWaste([]*domain.Card{makeCardWH(domain.CardDesignSpade, 1)})
		var tab [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
		tab[0] = []*domain.WhiteheadTableauCard{makeTableauCardWH(domain.CardDesignSpade, 2, true)}
		k.SetTableau(tab)
		k.SetStock(nil)

		// Build rest of foundations
		var f [domain.WhiteheadFoundationCnt][]*domain.Card
		for i := 1; i < domain.WhiteheadFoundationCnt; i++ {
			f[i] = make([]*domain.Card, 0)
			for v := 1; v <= 13; v++ {
				f[i] = append(f[i], makeCardWH(i+1, v))
			}
		}
		// Spade foundation has 0 cards, waste+tab has A,2
		f[0] = make([]*domain.Card, 0)
		for v := 3; v <= 13; v++ {
			f[0] = append(f[0], makeCardWH(domain.CardDesignSpade, v))
		}
		k.SetFoundation(f)
		err := k.AutoComplete()
		assert.NoError(t, err)
	})

	t.Run("error: not playing", func(t *testing.T) {
		k := setupPlayingWhitehead()
		k.SetPhase(domain.WhiteheadPhaseGameOver)
		err := k.AutoComplete()
		assert.Error(t, err)
	})

	t.Run("error: not all face up", func(t *testing.T) {
		k := setupPlayingWhitehead()
		err := k.AutoComplete()
		assert.Error(t, err)
		assert.Equal(t, "not all cards are face up", err.Error())
	})

	t.Run("waste card cannot go to foundation stops loop", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		k.SetWaste([]*domain.Card{makeCardWH(domain.CardDesignSpade, 5)})
		clearTableauWH(k)
		k.SetStock(nil)
		err := k.AutoComplete()
		assert.NoError(t, err)
		// Card 5 can't go on empty foundation, so it stays in waste
		assert.Equal(t, 1, len(k.GetWaste()))
	})
}

func TestWhitehead_AllFaceUp(t *testing.T) {
	t.Run("all face up", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		var tab [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
		tab[0] = []*domain.WhiteheadTableauCard{makeTableauCardWH(domain.CardDesignSpade, 1, true)}
		k.SetTableau(tab)
		k.SetStock(nil)
		assert.True(t, k.AllFaceUp())
	})

	t.Run("stock not empty", func(t *testing.T) {
		k := setupPlayingWhitehead()
		assert.False(t, k.AllFaceUp())
	})

	t.Run("face down card in tableau", func(t *testing.T) {
		k := newTestWhitehead()
		k.Reset()
		var tab [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
		tab[0] = []*domain.WhiteheadTableauCard{makeTableauCardWH(domain.CardDesignSpade, 1, false)}
		k.SetTableau(tab)
		k.SetStock(nil)
		assert.False(t, k.AllFaceUp())
	})
}

func TestWhitehead_ActionLog(t *testing.T) {
	k := setupPlayingWhitehead()
	assert.Nil(t, k.GetActionLog())
	_ = k.Draw()
	assert.NotNil(t, k.GetActionLog())
	assert.Equal(t, 1, len(k.GetActionLog()))
}

func TestWhitehead_MoveWasteToFoundation_DifferentSuit(t *testing.T) {
	k := newTestWhitehead()
	k.Reset()
	var f [domain.WhiteheadFoundationCnt][]*domain.Card
	f[0] = []*domain.Card{makeCardWH(domain.CardDesignSpade, 1)}
	k.SetFoundation(f)
	// Try to put heart 2 on spade foundation - different design
	k.SetWaste([]*domain.Card{makeCardWH(domain.CardDesignHeart, 2)})
	err := k.MoveWasteToFoundation()
	// This should fail because heart goes to foundation[2] (CardDesignHeart-1=2), not [0]
	// But foundation[2] is empty and card is 2, not ace
	assert.Error(t, err)
}

func TestWhitehead_IsSameColor(t *testing.T) {
	k := newTestWhitehead()
	k.Reset()
	// Whitehead is the mirror of Klondike here: same colour connects, alternating
	// colour does not. Both halves of the Klondike assertion were inverted rather
	// than deleted, so the flip stays visible.
	var tab [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
	tab[0] = []*domain.WhiteheadTableauCard{makeTableauCardWH(domain.CardDesignSpade, 7, true)}
	k.SetTableau(tab)

	k.SetWaste([]*domain.Card{makeCardWH(domain.CardDesignDiamond, 6)})
	assert.Error(t, k.MoveWasteToTableau(0))

	k.SetWaste([]*domain.Card{makeCardWH(domain.CardDesignClover, 6)})
	assert.NoError(t, k.MoveWasteToTableau(0))
}

func TestWhitehead_ResetClearsState(t *testing.T) {
	k := setupPlayingWhitehead()
	_ = k.Draw()
	k.GiveUp()
	assert.Equal(t, domain.WhiteheadPhaseGameOver, k.GetPhase())

	k.Reset()
	assert.Equal(t, domain.WhiteheadPhasePlaying, k.GetPhase())
	assert.Equal(t, 0, k.GetMoveCount())
	assert.Nil(t, k.GetActionLog())
}

func TestWhitehead_AutoFlipEmptyColumn(t *testing.T) {
	// Auto-flip should not panic on empty column
	k := newTestWhitehead()
	k.Reset()
	var tab [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
	tab[0] = []*domain.WhiteheadTableauCard{makeTableauCardWH(domain.CardDesignSpade, 1, true)}
	k.SetTableau(tab)
	err := k.MoveTableauToFoundation(0)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(k.GetTableau()[0]))
}

func TestWhitehead_CanPlaceOnFoundation_WrongSequence(t *testing.T) {
	k := newTestWhitehead()
	k.Reset()
	var f [domain.WhiteheadFoundationCnt][]*domain.Card
	f[0] = []*domain.Card{makeCardWH(domain.CardDesignSpade, 1), makeCardWH(domain.CardDesignSpade, 2)}
	k.SetFoundation(f)
	// Try to place Spade 4 (skipping 3)
	k.SetWaste([]*domain.Card{makeCardWH(domain.CardDesignSpade, 4)})
	err := k.MoveWasteToFoundation()
	assert.Error(t, err)
}

// --- Feature 1: 3-Card Draw ---

func TestWhitehead_DrawCount(t *testing.T) {
	t.Run("default drawCount is 1", func(t *testing.T) {
		k := newTestWhitehead()
		assert.Equal(t, 1, k.GetDrawCount())
	})

	t.Run("SetDrawCount(3) then Draw draws 3 cards", func(t *testing.T) {
		k := setupPlayingWhitehead()
		k.SetDrawCount(3)
		initialStock := k.GetStockCount()
		err := k.Draw()
		assert.NoError(t, err)
		assert.Equal(t, initialStock-3, k.GetStockCount())
		assert.Equal(t, 3, len(k.GetWaste()))
	})

	t.Run("Draw with stock < 3 draws remaining", func(t *testing.T) {
		k := setupPlayingWhitehead()
		k.SetDrawCount(3)
		// Leave only 2 cards in stock
		k.SetStock([]*domain.Card{
			makeCardWH(domain.CardDesignSpade, 1),
			makeCardWH(domain.CardDesignHeart, 2),
		})
		k.SetWaste(nil)
		err := k.Draw()
		assert.NoError(t, err)
		assert.Equal(t, 0, k.GetStockCount())
		assert.Equal(t, 2, len(k.GetWaste()))
	})

	t.Run("drawCount persists across Reset", func(t *testing.T) {
		k := newTestWhitehead()
		k.SetDrawCount(3)
		k.Reset()
		assert.Equal(t, 3, k.GetDrawCount())
	})

	t.Run("ResetWithConfig overrides drawCount", func(t *testing.T) {
		k := newTestWhitehead()
		k.SetDrawCount(3)
		k.ResetWithConfig(domain.WhiteheadConfig{DrawCount: 1})
		assert.Equal(t, 1, k.GetDrawCount())
		assert.Equal(t, domain.WhiteheadPhasePlaying, k.GetPhase())
	})

	t.Run("ResetWithConfig sets drawCount to 3", func(t *testing.T) {
		k := newTestWhitehead()
		k.ResetWithConfig(domain.WhiteheadConfig{DrawCount: 3})
		assert.Equal(t, 3, k.GetDrawCount())
	})

	t.Run("invalid drawCount clamps to 1", func(t *testing.T) {
		k := newTestWhitehead()
		k.ResetWithConfig(domain.WhiteheadConfig{DrawCount: 5})
		assert.Equal(t, 1, k.GetDrawCount())
	})

	t.Run("recycle in 3-card mode unchanged", func(t *testing.T) {
		k := setupPlayingWhitehead()
		k.SetDrawCount(3)
		// Draw all stock cards
		for k.GetStockCount() > 0 {
			_ = k.Draw()
		}
		wasteLen := len(k.GetWaste())
		assert.Greater(t, wasteLen, 0)

		// Recycle
		err := k.Draw()
		assert.NoError(t, err)
		assert.Equal(t, wasteLen, k.GetStockCount())
		assert.Nil(t, k.GetWaste())
	})
}

// --- Feature 2: Undo ---

func TestWhitehead_Undo(t *testing.T) {
	t.Run("Undo after draw restores stock/waste/moveCount", func(t *testing.T) {
		k := setupPlayingWhitehead()
		initialStock := k.GetStockCount()
		err := k.Draw()
		assert.NoError(t, err)
		assert.Equal(t, initialStock-1, k.GetStockCount())
		assert.Equal(t, 1, k.GetMoveCount())

		err = k.Undo()
		assert.NoError(t, err)
		assert.Equal(t, initialStock, k.GetStockCount())
		assert.Equal(t, 0, k.GetMoveCount())
		assert.Equal(t, 0, len(k.GetWaste()))
	})

	t.Run("Undo after MoveWasteToTableau restores state", func(t *testing.T) {
		k := setupPlayingWhitehead()
		k.SetWaste([]*domain.Card{makeCardWH(domain.CardDesignClover, 6)})
		var tab [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
		tab[0] = []*domain.WhiteheadTableauCard{makeTableauCardWH(domain.CardDesignSpade, 7, true)}
		k.SetTableau(tab)

		err := k.MoveWasteToTableau(0)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(k.GetTableau()[0]))
		assert.Equal(t, 0, len(k.GetWaste()))

		err = k.Undo()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(k.GetTableau()[0]))
		assert.Equal(t, 1, len(k.GetWaste()))
	})

	t.Run("Undo after MoveWasteToFoundation", func(t *testing.T) {
		k := setupPlayingWhitehead()
		k.SetWaste([]*domain.Card{makeCardWH(domain.CardDesignSpade, 1)})

		err := k.MoveWasteToFoundation()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(k.GetFoundation()[0]))

		err = k.Undo()
		assert.NoError(t, err)
		assert.Equal(t, 0, len(k.GetFoundation()[0]))
		assert.Equal(t, 1, len(k.GetWaste()))
	})

	t.Run("Undo after MoveTableauToTableau", func(t *testing.T) {
		k := setupPlayingWhitehead()
		var tab [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
		// Same colour: Whitehead builds black-on-black / red-on-red.
		tab[0] = []*domain.WhiteheadTableauCard{makeTableauCardWH(domain.CardDesignClover, 6, true)}
		tab[1] = []*domain.WhiteheadTableauCard{makeTableauCardWH(domain.CardDesignSpade, 7, true)}
		k.SetTableau(tab)

		err := k.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(k.GetTableau()[0]))
		assert.Equal(t, 2, len(k.GetTableau()[1]))

		err = k.Undo()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(k.GetTableau()[0]))
		assert.Equal(t, 1, len(k.GetTableau()[1]))
	})

	t.Run("Undo after MoveTableauToFoundation", func(t *testing.T) {
		k := setupPlayingWhitehead()
		var tab [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
		tab[0] = []*domain.WhiteheadTableauCard{makeTableauCardWH(domain.CardDesignSpade, 1, true)}
		k.SetTableau(tab)

		err := k.MoveTableauToFoundation(0)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(k.GetTableau()[0]))
		assert.Equal(t, 1, len(k.GetFoundation()[0]))

		err = k.Undo()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(k.GetTableau()[0]))
		assert.Equal(t, 0, len(k.GetFoundation()[0]))
	})

	t.Run("Multiple undos back to initial state", func(t *testing.T) {
		k := setupPlayingWhitehead()
		initialStock := k.GetStockCount()

		// Perform 3 draws
		for i := 0; i < 3; i++ {
			err := k.Draw()
			assert.NoError(t, err)
		}
		assert.Equal(t, 3, k.GetMoveCount())

		// Undo all 3
		for i := 0; i < 3; i++ {
			err := k.Undo()
			assert.NoError(t, err)
		}
		assert.Equal(t, 0, k.GetMoveCount())
		assert.Equal(t, initialStock, k.GetStockCount())
		assert.Equal(t, 0, len(k.GetWaste()))
	})

	t.Run("Undo when no history returns error", func(t *testing.T) {
		k := setupPlayingWhitehead()
		err := k.Undo()
		assert.Error(t, err)
		assert.Equal(t, "cannot undo: no history", err.Error())
	})

	t.Run("Undo when not playing returns error", func(t *testing.T) {
		k := setupPlayingWhitehead()
		_ = k.Draw()
		k.SetPhase(domain.WhiteheadPhaseGameOver)
		err := k.Undo()
		assert.Error(t, err)
		assert.Equal(t, "cannot undo: game is not in playing phase", err.Error())
	})

	t.Run("Reset clears history", func(t *testing.T) {
		k := setupPlayingWhitehead()
		_ = k.Draw()
		assert.True(t, k.CanUndo())

		k.Reset()
		assert.False(t, k.CanUndo())
	})

	t.Run("CanUndo true when history exists and playing", func(t *testing.T) {
		k := setupPlayingWhitehead()
		assert.False(t, k.CanUndo())

		_ = k.Draw()
		assert.True(t, k.CanUndo())
	})

	t.Run("CanUndo false when not playing", func(t *testing.T) {
		k := setupPlayingWhitehead()
		_ = k.Draw()
		k.SetPhase(domain.WhiteheadPhaseGameOver)
		assert.False(t, k.CanUndo())
	})

	t.Run("Deep copy: mutating original does not affect snapshot", func(t *testing.T) {
		k := setupPlayingWhitehead()
		var tab [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
		tab[0] = []*domain.WhiteheadTableauCard{
			makeTableauCardWH(domain.CardDesignSpade, 10, false),
			makeTableauCardWH(domain.CardDesignHeart, 6, true),
		}
		tab[1] = []*domain.WhiteheadTableauCard{makeTableauCardWH(domain.CardDesignSpade, 7, true)}
		k.SetTableau(tab)

		// Draw creates a snapshot
		_ = k.Draw()

		// Mutate tableau directly after snapshot
		k.GetTableau()[0][0].FaceUp = true

		// Undo should restore the original FaceUp=false from the snapshot
		err := k.Undo()
		assert.NoError(t, err)
		assert.False(t, k.GetTableau()[0][0].FaceUp, "snapshot should have preserved FaceUp=false")
	})

	t.Run("Undo after GiveUp fails because phase is GameOver", func(t *testing.T) {
		k := setupPlayingWhitehead()
		_ = k.Draw() // create some history
		k.GiveUp()
		assert.Equal(t, domain.WhiteheadPhaseGameOver, k.GetPhase())

		err := k.Undo()
		assert.Error(t, err)
		assert.Equal(t, "cannot undo: game is not in playing phase", err.Error())
	})

	t.Run("Undo after AutoComplete", func(t *testing.T) {
		k := setupPlayingWhitehead()
		// Set up all cards face up with just a few cards
		var tab [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
		tab[0] = []*domain.WhiteheadTableauCard{makeTableauCardWH(domain.CardDesignSpade, 1, true)}
		k.SetTableau(tab)
		k.SetStock(nil)
		k.SetWaste(nil)

		// Pre-fill foundation for other suits
		var f [domain.WhiteheadFoundationCnt][]*domain.Card
		for i := 1; i < domain.WhiteheadFoundationCnt; i++ {
			f[i] = make([]*domain.Card, 0)
			for v := 1; v <= 13; v++ {
				f[i] = append(f[i], makeCardWH(i+1, v))
			}
		}
		// Spade foundation has cards 1-12 already, tab has 13
		f[0] = make([]*domain.Card, 0)
		for v := 1; v <= 12; v++ {
			f[0] = append(f[0], makeCardWH(domain.CardDesignSpade, v))
		}
		tab[0] = []*domain.WhiteheadTableauCard{makeTableauCardWH(domain.CardDesignSpade, 13, true)}
		k.SetTableau(tab)
		k.SetFoundation(f)

		err := k.AutoComplete()
		assert.NoError(t, err)
		assert.Equal(t, domain.WhiteheadPhaseGameClear, k.GetPhase())

		// Undo should fail because phase is GameClear
		err = k.Undo()
		assert.Error(t, err)
	})

	t.Run("Undo after recycle", func(t *testing.T) {
		k := setupPlayingWhitehead()
		// Draw all stock cards
		for k.GetStockCount() > 0 {
			_ = k.Draw()
		}
		wasteLen := len(k.GetWaste())

		// Recycle waste to stock
		err := k.Draw()
		assert.NoError(t, err)
		assert.Equal(t, wasteLen, k.GetStockCount())
		assert.Nil(t, k.GetWaste())

		// Undo should restore waste and empty stock
		err = k.Undo()
		assert.NoError(t, err)
		assert.Equal(t, 0, k.GetStockCount())
		assert.Equal(t, wasteLen, len(k.GetWaste()))
	})
}

// --- Feature 3: Vegas Scoring ---

func TestWhitehead_Score(t *testing.T) {
	t.Run("GetScore with 0 foundation cards = -52", func(t *testing.T) {
		k := setupPlayingWhitehead()
		assert.Equal(t, -52, k.GetScore())
	})

	t.Run("GetScore with partial foundation", func(t *testing.T) {
		k := setupPlayingWhitehead()
		var f [domain.WhiteheadFoundationCnt][]*domain.Card
		f[0] = []*domain.Card{
			makeCardWH(domain.CardDesignSpade, 1),
			makeCardWH(domain.CardDesignSpade, 2),
		}
		f[1] = []*domain.Card{
			makeCardWH(domain.CardDesignClover, 1),
		}
		k.SetFoundation(f)
		// 3 cards in foundation: -52 + 5*3 = -37
		assert.Equal(t, -37, k.GetScore())
	})

	t.Run("GetScore with full foundation = 208", func(t *testing.T) {
		k := setupPlayingWhitehead()
		var f [domain.WhiteheadFoundationCnt][]*domain.Card
		designs := []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond}
		for i := 0; i < domain.WhiteheadFoundationCnt; i++ {
			f[i] = make([]*domain.Card, 0)
			for v := 1; v <= 13; v++ {
				f[i] = append(f[i], makeCardWH(designs[i], v))
			}
		}
		k.SetFoundation(f)
		// 52 cards: -52 + 5*52 = 208
		assert.Equal(t, 208, k.GetScore())
	})

	t.Run("scoring mode default is None", func(t *testing.T) {
		k := newTestWhitehead()
		assert.Equal(t, domain.WhiteheadScoringNone, k.GetScoringMode())
	})

	t.Run("SetScoringMode and GetScoringMode", func(t *testing.T) {
		k := newTestWhitehead()
		k.SetScoringMode(domain.WhiteheadScoringVegas)
		assert.Equal(t, domain.WhiteheadScoringVegas, k.GetScoringMode())
		k.SetScoringMode(domain.WhiteheadScoringNone)
		assert.Equal(t, domain.WhiteheadScoringNone, k.GetScoringMode())
	})

	t.Run("ResetWithConfig sets scoring mode to Vegas", func(t *testing.T) {
		k := newTestWhitehead()
		k.ResetWithConfig(domain.WhiteheadConfig{DrawCount: 1, ScoringMode: domain.WhiteheadScoringVegas})
		assert.Equal(t, domain.WhiteheadScoringVegas, k.GetScoringMode())
	})

	t.Run("ResetWithConfig sets scoring mode to None", func(t *testing.T) {
		k := newTestWhitehead()
		k.SetScoringMode(domain.WhiteheadScoringVegas)
		k.ResetWithConfig(domain.WhiteheadConfig{DrawCount: 1, ScoringMode: domain.WhiteheadScoringNone})
		assert.Equal(t, domain.WhiteheadScoringNone, k.GetScoringMode())
	})

	t.Run("scoring mode persists across Reset", func(t *testing.T) {
		k := newTestWhitehead()
		k.ResetWithConfig(domain.WhiteheadConfig{DrawCount: 1, ScoringMode: domain.WhiteheadScoringVegas})
		k.Reset()
		assert.Equal(t, domain.WhiteheadScoringVegas, k.GetScoringMode())
	})
}

// --- UndoToEscape / UndoN tests ---

func TestWhitehead_UndoToEscape_NotInStalemate(t *testing.T) {
	k := setupPlayingWhitehead()
	assert.Equal(t, 0, k.UndoToEscape())
}

func TestWhitehead_UndoToEscape_StalemateNoHistory(t *testing.T) {
	k := setupPlayingWhitehead()
	k.SetIsStalemate(true)
	assert.Equal(t, -1, k.UndoToEscape())
}

func TestWhitehead_UndoToEscape_StalemateWithEscape(t *testing.T) {
	k := setupPlayingWhitehead()
	// Make a move to create history
	err := k.Draw()
	assert.NoError(t, err)
	// Set stalemate after the move
	k.SetIsStalemate(true)
	n := k.UndoToEscape()
	assert.Equal(t, 1, n)
}

func TestWhitehead_UndoN_Zero(t *testing.T) {
	k := setupPlayingWhitehead()
	err := k.UndoN(0)
	assert.NoError(t, err)
}

func TestWhitehead_UndoN_Valid(t *testing.T) {
	k := setupPlayingWhitehead()
	_ = k.Draw()
	_ = k.Draw()
	err := k.UndoN(2)
	assert.NoError(t, err)
}

func TestWhitehead_UndoN_Excessive(t *testing.T) {
	k := setupPlayingWhitehead()
	_ = k.Draw()
	err := k.UndoN(5)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "undo step")
}

// **もう自動で揃えられることを CUI は知らせていなかった (#4776)。**Web は
// 条件が揃うとボタンを光らせバッジも出す。
func TestWhitehead_CanAutoComplete(t *testing.T) {
	card := func(v int) *domain.Card { return domain.NewCard(domain.CardDesignSpade, v, false) }
	board := func(stock []*domain.Card, faceDown bool) *domain.Whitehead {
		k := domain.NewWhitehead(domain.NewTrumpCards(0))
		k.Reset()
		k.SetPhase(domain.WhiteheadPhasePlaying)
		k.SetStock(stock)
		var tableau [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
		for i := range tableau {
			tableau[i] = []*domain.WhiteheadTableauCard{{Card: card(i + 2), FaceUp: true}}
		}
		if faceDown {
			tableau[0][0].FaceUp = false
		}
		k.SetTableau(tableau)
		return k
	}

	t.Run("ready once the stock is empty and everything is face up", func(t *testing.T) {
		assert.True(t, board(nil, false).CanAutoComplete())
	})

	// **山札が残っていれば駄目。**まだ引ける札があるうちは全部見えていない。
	t.Run("not ready while cards remain in the stock", func(t *testing.T) {
		assert.False(t, board([]*domain.Card{card(5)}, false).CanAutoComplete())
	})

	t.Run("not ready while a tableau card is face down", func(t *testing.T) {
		assert.False(t, board(nil, true).CanAutoComplete())
	})

	t.Run("not ready outside the playing phase", func(t *testing.T) {
		k := board(nil, false)
		k.SetPhase(domain.WhiteheadPhaseGameClear)
		assert.False(t, k.CanAutoComplete())
	})

	// **表示の条件と AutoComplete が通る条件は同じでなければならない。**
	t.Run("agrees with AutoComplete", func(t *testing.T) {
		ready := board(nil, false)
		require.True(t, ready.CanAutoComplete())
		assert.NoError(t, ready.AutoComplete())

		blocked := board([]*domain.Card{card(5)}, false)
		require.False(t, blocked.CanAutoComplete())
		assert.Error(t, blocked.AutoComplete())
	})
}
