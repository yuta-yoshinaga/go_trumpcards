//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestDiplomat() *Diplomat {
	c := NewDefaultDiplomat()
	c.Reset()
	return c
}

// clearDiplomatBoard wipes the dealt layout so a test can state exactly the
// position it cares about. Never assert on a freshly Reset board -- the deal is
// shuffled, so any such assertion is a hidden flake.
func clearDiplomatBoard(c *Diplomat) {
	c.stock = nil
	c.waste = nil
	c.isStalemate = false
	c.history = nil
	c.moveCount = 0
	c.phase = DiplomatPhasePlaying
	for i := range DiplomatFoundationCnt {
		c.foundation[i] = nil
	}
	for i := range DiplomatTableauCnt {
		c.tableau[i] = nil
	}
}

// fillDiplomatColumns puts one dead card in every column so no gap exists. A
// gap changes which moves are legal, so a board full of holes would test a
// position the game never reaches by accident.
func fillDiplomatColumns(c *Diplomat) {
	for i := range DiplomatTableauCnt {
		c.tableau[i] = []*Card{NewCard(CardDesignSpade, 9, true)}
	}
}

func TestNewDiplomat(t *testing.T) {
	assert.NotNil(t, NewDiplomat(NewTrumpCardsWithDecks(2, 0)))
	assert.NotNil(t, NewDefaultDiplomat())
}

// The deal is 8 columns of FOUR with 72 left in the stock. #5276 adds a 4-card
// reserve, which would leave 68 -- every rulebook agrees on 32 dealt and 72
// held back.
func TestDiplomat_Reset(t *testing.T) {
	c := newTestDiplomat()

	for i, pile := range c.GetTableau() {
		assert.Len(t, pile, DiplomatDealPerColumn, "column %d", i)
	}
	assert.Equal(t, 4, DiplomatDealPerColumn)
	assert.Equal(t, 32, DiplomatTableauCnt*DiplomatDealPerColumn)
	assert.Equal(t, 72, c.GetStockCount())
	assert.Empty(t, c.GetWaste())

	// The foundations start EMPTY -- Aces go up as they turn up.
	for i, pile := range c.GetFoundation() {
		assert.Empty(t, pile, "foundation %d", i)
	}

	assert.Equal(t, DiplomatPhasePlaying, c.GetPhase())
	assert.Equal(t, 0, c.GetMoveCount())
	assert.True(t, c.AllFaceUp())
	assert.False(t, c.GetGameEndFlag())
}

func TestDiplomat_ResetDealsEveryCard(t *testing.T) {
	for range 20 {
		c := newTestDiplomat()
		total := c.GetStockCount() + len(c.GetWaste())
		for _, pile := range c.GetTableau() {
			total += len(pile)
		}
		for _, pile := range c.GetFoundation() {
			total += len(pile)
		}
		assert.Equal(t, DiplomatTotalCards, total)
	}
}

func TestDiplomat_ResetTwiceIsClean(t *testing.T) {
	c := newTestDiplomat()
	require.NoError(t, c.Draw())
	c.Reset()
	assert.Empty(t, c.GetWaste())
	assert.Equal(t, 0, c.GetMoveCount())
	assert.False(t, c.CanUndo())
	assert.Equal(t, 72, c.GetStockCount())
}

// --- Draw ---

// One card at a time, not three. #5276 asks for a Klondike-style three-card
// turn, which this family does not use.
func TestDiplomat_Draw_TurnsOneCard(t *testing.T) {
	c := newTestDiplomat()
	before := c.GetStockCount()
	require.NoError(t, c.Draw())
	assert.Equal(t, before-1, c.GetStockCount())
	assert.Len(t, c.GetWaste(), 1)
	assert.Equal(t, 1, c.GetMoveCount())
	assert.True(t, c.CanUndo())

	require.NoError(t, c.Draw())
	assert.Equal(t, before-2, c.GetStockCount())
	assert.Len(t, c.GetWaste(), 2)
}

func TestDiplomat_Draw_NoRedeal(t *testing.T) {
	c := newTestDiplomat()
	clearDiplomatBoard(c)
	fillDiplomatColumns(c)
	c.waste = []*Card{NewCard(CardDesignHeart, 5, true)}
	err := c.Draw()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no redeal")
	assert.Len(t, c.GetWaste(), 1, "waste must not be recycled into the stock")
}

func TestDiplomat_Draw_RejectedWhenNotPlaying(t *testing.T) {
	c := newTestDiplomat()
	c.GiveUp()
	assert.Error(t, c.Draw())
}

// --- Tableau rules ---

// Suit is ignored: only the rank sequence matters. This is what makes Diplomat
// the gentle member of the Forty Thieves family.
func TestDiplomat_CanPlaceOnTableau_IgnoresSuit(t *testing.T) {
	c := newTestDiplomat()
	clearDiplomatBoard(c)
	c.tableau[0] = []*Card{NewCard(CardDesignSpade, 9, true)}

	assert.True(t, c.canPlaceOnTableau(NewCard(CardDesignHeart, 8, true), 0), "a different suit is fine")
	assert.True(t, c.canPlaceOnTableau(NewCard(CardDesignSpade, 8, true), 0))
	assert.False(t, c.canPlaceOnTableau(NewCard(CardDesignHeart, 10, true), 0), "ascending is not allowed")
	assert.False(t, c.canPlaceOnTableau(NewCard(CardDesignHeart, 7, true), 0), "a gap is not allowed")
	assert.False(t, c.canPlaceOnTableau(nil, 0))
}

// No wraparound: an Ace ends its column.
func TestDiplomat_CanPlaceOnTableau_NoWraparound(t *testing.T) {
	c := newTestDiplomat()
	clearDiplomatBoard(c)
	c.tableau[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
	for v := 1; v <= CardValueMax; v++ {
		assert.False(t, c.canPlaceOnTableau(NewCard(CardDesignHeart, v, true), 0), "value %d under an Ace", v)
	}
}

func TestDiplomat_CanPlaceOnTableau_EmptyColumnTakesAnything(t *testing.T) {
	c := newTestDiplomat()
	clearDiplomatBoard(c)
	assert.True(t, c.canPlaceOnTableau(NewCard(CardDesignHeart, 1, true), 0))
	assert.True(t, c.canPlaceOnTableau(NewCard(CardDesignSpade, CardValueMax, true), 0))
}

// --- Tableau to tableau ---

func TestDiplomat_MoveTableauToTableau(t *testing.T) {
	c := newTestDiplomat()
	clearDiplomatBoard(c)
	fillDiplomatColumns(c)
	c.tableau[1] = []*Card{NewCard(CardDesignHeart, 8, true)}

	require.NoError(t, c.MoveTableauToTableau(1, 0))
	assert.Empty(t, c.GetTableau()[1])
	assert.Len(t, c.GetTableau()[0], 2)
	assert.Equal(t, 1, c.GetMoveCount())
}

// #5276 does not mention it, but an empty column takes any single card from
// another column -- that is the game's main release valve.
func TestDiplomat_MoveTableauToTableau_FillsAnEmptyColumn(t *testing.T) {
	c := newTestDiplomat()
	clearDiplomatBoard(c)
	fillDiplomatColumns(c)
	c.tableau[3] = nil
	c.tableau[0] = []*Card{NewCard(CardDesignSpade, 9, true), NewCard(CardDesignHeart, 2, true)}

	require.NoError(t, c.MoveTableauToTableau(0, 3))
	assert.Len(t, c.GetTableau()[3], 1)
	assert.Equal(t, 2, c.GetTableau()[3][0].GetValue())
	assert.Len(t, c.GetTableau()[0], 1)
}

func TestDiplomat_MoveTableauToTableau_Rejections(t *testing.T) {
	c := newTestDiplomat()
	clearDiplomatBoard(c)
	fillDiplomatColumns(c)

	assert.Error(t, c.MoveTableauToTableau(-1, 0))
	assert.Error(t, c.MoveTableauToTableau(0, DiplomatTableauCnt))

	err := c.MoveTableauToTableau(0, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "same pile")

	c.tableau[2] = nil
	err = c.MoveTableauToTableau(2, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")

	// Two 9s cannot stack: the rank has to drop by exactly one.
	err = c.MoveTableauToTableau(0, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be placed")

	c.GiveUp()
	assert.Error(t, c.MoveTableauToTableau(0, 1))
}

// Only the top card moves; a run never travels as a group.
func TestDiplomat_MoveTableauToTableau_MovesOneCardOnly(t *testing.T) {
	c := newTestDiplomat()
	clearDiplomatBoard(c)
	fillDiplomatColumns(c)
	c.tableau[0] = []*Card{
		NewCard(CardDesignSpade, 10, true),
		NewCard(CardDesignHeart, 9, true),
		NewCard(CardDesignClover, 8, true),
	}
	c.tableau[1] = []*Card{NewCard(CardDesignDiamond, 9, true)}

	require.NoError(t, c.MoveTableauToTableau(0, 1))
	assert.Len(t, c.GetTableau()[0], 2, "only the 8 left")
	assert.Len(t, c.GetTableau()[1], 2)
}

// --- Foundations ---

func TestDiplomat_MoveTableauToFoundation(t *testing.T) {
	c := newTestDiplomat()
	clearDiplomatBoard(c)
	fillDiplomatColumns(c)
	c.tableau[3] = []*Card{NewCard(CardDesignSpade, 1, true)}

	require.NoError(t, c.MoveTableauToFoundation(3))
	assert.Len(t, c.GetFoundation()[0], 1)
	assert.Empty(t, c.GetTableau()[3])
}

func TestDiplomat_CanPlaceOnFoundation(t *testing.T) {
	c := newTestDiplomat()
	clearDiplomatBoard(c)

	assert.True(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 1, true), 0))
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 2, true), 0))
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignHeart, 1, true), 0), "wrong suit")
	assert.False(t, c.canPlaceOnFoundation(nil, 0))

	c.foundation[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
	assert.True(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 2, true), 0))
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 3, true), 0))
}

// Two decks means two of each card, and each suit has two foundations, so the
// second copy always has a home.
func TestDiplomat_FindFoundation_SecondCopyGoesToTheOtherPile(t *testing.T) {
	c := newTestDiplomat()
	clearDiplomatBoard(c)
	ace := NewCard(CardDesignSpade, 1, true)
	assert.Equal(t, 0, c.findFoundation(ace))

	c.foundation[0] = []*Card{ace}
	assert.Equal(t, 4, c.findFoundation(NewCard(CardDesignSpade, 1, true)))
	assert.Equal(t, -1, c.findFoundation(NewCard(CardDesignSpade, 7, true)))
	assert.Equal(t, -1, c.findFoundation(nil))
}

func TestDiplomat_MoveTableauToFoundation_Rejections(t *testing.T) {
	c := newTestDiplomat()
	clearDiplomatBoard(c)
	fillDiplomatColumns(c)

	assert.Error(t, c.MoveTableauToFoundation(-1))
	assert.Error(t, c.MoveTableauToFoundation(DiplomatTableauCnt))

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

// --- Waste ---

func TestDiplomat_MoveWasteToFoundation(t *testing.T) {
	c := newTestDiplomat()
	clearDiplomatBoard(c)
	fillDiplomatColumns(c)
	c.waste = []*Card{NewCard(CardDesignClover, 1, true)}

	require.NoError(t, c.MoveWasteToFoundation())
	assert.Len(t, c.GetFoundation()[1], 1)
	assert.Empty(t, c.GetWaste())
}

func TestDiplomat_MoveWasteToFoundation_Rejections(t *testing.T) {
	c := newTestDiplomat()
	clearDiplomatBoard(c)
	fillDiplomatColumns(c)

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

func TestDiplomat_MoveWasteToTableau(t *testing.T) {
	c := newTestDiplomat()
	clearDiplomatBoard(c)
	fillDiplomatColumns(c)
	c.waste = []*Card{NewCard(CardDesignHeart, 8, true)}

	require.NoError(t, c.MoveWasteToTableau(0))
	assert.Len(t, c.GetTableau()[0], 2)
	assert.Empty(t, c.GetWaste())
}

func TestDiplomat_MoveWasteToTableau_FillsAnEmptyColumn(t *testing.T) {
	c := newTestDiplomat()
	clearDiplomatBoard(c)
	fillDiplomatColumns(c)
	c.tableau[5] = nil
	c.waste = []*Card{NewCard(CardDesignHeart, 2, true)}

	require.NoError(t, c.MoveWasteToTableau(5))
	assert.Len(t, c.GetTableau()[5], 1)
}

func TestDiplomat_MoveWasteToTableau_Rejections(t *testing.T) {
	c := newTestDiplomat()
	clearDiplomatBoard(c)
	fillDiplomatColumns(c)

	assert.Error(t, c.MoveWasteToTableau(-1))
	assert.Error(t, c.MoveWasteToTableau(DiplomatTableauCnt))

	err := c.MoveWasteToTableau(0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "waste is empty")

	// A 2 cannot sit on a 9.
	c.waste = []*Card{NewCard(CardDesignHeart, 2, true)}
	err = c.MoveWasteToTableau(0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be placed")

	c.GiveUp()
	assert.Error(t, c.MoveWasteToTableau(0))
}

// --- Hints and stalemate ---

func TestDiplomat_GetHint_PrefersFoundation(t *testing.T) {
	c := newTestDiplomat()
	clearDiplomatBoard(c)
	fillDiplomatColumns(c)
	c.tableau[6] = []*Card{NewCard(CardDesignClover, 1, true)}
	c.stock = []*Card{NewCard(CardDesignDiamond, 4, true)}

	h := c.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "tableau", h.FromZone)
	assert.Equal(t, 6, h.FromIdx)
	assert.Equal(t, "foundation", h.ToZone)
	assert.Equal(t, 1, h.ToIdx)
}

func TestDiplomat_GetHint_WasteToTableau(t *testing.T) {
	c := newTestDiplomat()
	clearDiplomatBoard(c)
	fillDiplomatColumns(c)
	c.waste = []*Card{NewCard(CardDesignHeart, 8, true)}

	h := c.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "waste", h.FromZone)
	assert.Equal(t, "tableau", h.ToZone)
}

func TestDiplomat_GetHint_DrawsWhenNothingElse(t *testing.T) {
	c := newTestDiplomat()
	clearDiplomatBoard(c)
	fillDiplomatColumns(c)
	c.stock = []*Card{NewCard(CardDesignDiamond, 4, true)}

	h := c.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "stock", h.FromZone)
	assert.Equal(t, "waste", h.ToZone)
}

// Moving a lone card into an empty column just swaps which column is empty, so
// the hint must not offer it -- it would repeat forever.
func TestDiplomat_GetHint_SkipsTheEmptyColumnShuffle(t *testing.T) {
	c := newTestDiplomat()
	clearDiplomatBoard(c)
	fillDiplomatColumns(c)
	c.tableau[4] = nil

	h := c.GetHint()
	assert.Nil(t, h, "every other column holds a single card; nothing progresses")

	// Negative control: a column with two cards CAN usefully shed its top.
	c.tableau[0] = []*Card{NewCard(CardDesignSpade, 9, true), NewCard(CardDesignHeart, 3, true)}
	h = c.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "tableau", h.FromZone)
	assert.Equal(t, 0, h.FromIdx)
	assert.Equal(t, 4, h.ToIdx)
}

func TestDiplomat_GetHint_NilWhenNotPlaying(t *testing.T) {
	c := newTestDiplomat()
	c.GiveUp()
	assert.Nil(t, c.GetHint())
	assert.Nil(t, c.foundationHint())
	assert.Nil(t, c.tableauHint())
}

func TestDiplomat_Stalemate(t *testing.T) {
	c := newTestDiplomat()
	clearDiplomatBoard(c)
	fillDiplomatColumns(c)
	c.checkStalemate()
	assert.True(t, c.IsStalemate(), "eight equal cards, no stock, no waste")

	c.stock = []*Card{NewCard(CardDesignDiamond, 4, true)}
	c.checkStalemate()
	assert.False(t, c.IsStalemate())
}

func TestDiplomat_Stalemate_NotSetOutsidePlaying(t *testing.T) {
	c := newTestDiplomat()
	clearDiplomatBoard(c)
	c.phase = DiplomatPhaseGameOver
	c.checkStalemate()
	assert.False(t, c.IsStalemate())
}

func TestDiplomat_UndoToEscape(t *testing.T) {
	c := newTestDiplomat()
	clearDiplomatBoard(c)
	fillDiplomatColumns(c)
	c.tableau[0] = []*Card{NewCard(CardDesignSpade, 9, true), NewCard(CardDesignSpade, 1, true)}
	c.checkStalemate()
	require.False(t, c.IsStalemate(), "the ace can still go up")

	// Sending the ace up leaves eight 9s and no 8 anywhere: nothing combines,
	// and the stock and waste are both gone.
	require.NoError(t, c.MoveTableauToFoundation(0))
	assert.True(t, c.IsStalemate())
	assert.Equal(t, 1, c.UndoToEscape())
}

// --- AutoComplete / clear ---

func TestDiplomat_AutoComplete(t *testing.T) {
	c := newTestDiplomat()
	clearDiplomatBoard(c)
	fillDiplomatColumns(c)
	c.tableau[0] = []*Card{NewCard(CardDesignClover, 2, true), NewCard(CardDesignClover, 1, true)}
	c.waste = []*Card{NewCard(CardDesignHeart, 1, true)}

	require.NoError(t, c.AutoComplete())
	assert.Len(t, c.GetFoundation()[1], 2, "the ace then the deuce under it")
	assert.Len(t, c.GetFoundation()[2], 1, "the waste ace too")
}

func TestDiplomat_AutoComplete_Rejections(t *testing.T) {
	c := newTestDiplomat()
	clearDiplomatBoard(c)
	fillDiplomatColumns(c)
	err := c.AutoComplete()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no card")

	c.GiveUp()
	assert.Error(t, c.AutoComplete())
}

func TestDiplomat_GameClear(t *testing.T) {
	c := newTestDiplomat()
	clearDiplomatBoard(c)
	for i := range DiplomatFoundationCnt {
		for v := 1; v <= CardValueMax; v++ {
			c.foundation[i] = append(c.foundation[i], NewCard(diplomatSuitOrder[i], v, true))
		}
	}
	c.foundation[7] = c.foundation[7][:CardValueMax-1]
	c.checkGameClear()
	assert.Equal(t, DiplomatPhasePlaying, c.GetPhase(), "one card still missing")

	c.tableau[0] = []*Card{NewCard(CardDesignDiamond, CardValueMax, true)}
	require.NoError(t, c.MoveTableauToFoundation(0))
	assert.Equal(t, DiplomatPhaseGameClear, c.GetPhase())
	assert.True(t, c.GetGameEndFlag())
}

// --- Undo / log ---

func TestDiplomat_Undo(t *testing.T) {
	c := newTestDiplomat()
	clearDiplomatBoard(c)
	fillDiplomatColumns(c)
	c.tableau[1] = []*Card{NewCard(CardDesignHeart, 8, true)}

	assert.False(t, c.CanUndo())
	require.Error(t, c.Undo())

	require.NoError(t, c.MoveTableauToTableau(1, 0))
	require.NoError(t, c.Undo())
	assert.Len(t, c.GetTableau()[1], 1)
	assert.Len(t, c.GetTableau()[0], 1)
	assert.Equal(t, 0, c.GetMoveCount())
}

func TestDiplomat_UndoN(t *testing.T) {
	c := newTestDiplomat()
	require.NoError(t, c.Draw())
	require.NoError(t, c.Draw())
	require.NoError(t, c.Draw())
	require.NoError(t, c.UndoN(2))
	assert.Equal(t, 1, c.GetMoveCount())
	assert.Len(t, c.GetWaste(), 1)

	assert.Error(t, c.UndoN(5))
	assert.Error(t, c.UndoN(0))
}

func TestDiplomat_GiveUp(t *testing.T) {
	c := newTestDiplomat()
	c.GiveUp()
	assert.Equal(t, DiplomatPhaseGameOver, c.GetPhase())
	assert.True(t, c.GetGameEndFlag())
	require.NotEmpty(t, c.GetActionLog())

	before := len(c.GetActionLog())
	c.GiveUp()
	assert.Len(t, c.GetActionLog(), before)
}

func TestDiplomat_ActionLog(t *testing.T) {
	c := newTestDiplomat()
	require.NoError(t, c.Draw())
	log := c.GetActionLog()
	require.NotEmpty(t, log)
	assert.Equal(t, "draw", log[len(log)-1].ActionType)
}

// --- JSON round-trip ---

func TestDiplomat_JSONRoundTrip(t *testing.T) {
	c := newTestDiplomat()
	require.NoError(t, c.Draw())
	require.NoError(t, c.Draw())

	data, err := json.Marshal(c)
	require.NoError(t, err)

	restored := NewDefaultDiplomat()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, c.GetStockCount(), restored.GetStockCount())
	assert.Len(t, restored.GetWaste(), len(c.GetWaste()))
	assert.Equal(t, c.GetMoveCount(), restored.GetMoveCount())
	assert.Equal(t, c.GetPhase(), restored.GetPhase())
	for i := range DiplomatTableauCnt {
		assert.Len(t, restored.GetTableau()[i], len(c.GetTableau()[i]), "column %d", i)
	}
}

// The undo stack has to survive the trip: the Worker rebuilds the game from KV
// on every request, so an unpersisted history means Undo silently never works.
func TestDiplomat_JSONRoundTripKeepsUndoHistory(t *testing.T) {
	c := newTestDiplomat()
	require.NoError(t, c.Draw())
	require.NoError(t, c.Draw())
	wasteBefore := len(c.GetWaste())

	data, err := json.Marshal(c)
	require.NoError(t, err)
	restored := NewDefaultDiplomat()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.True(t, restored.CanUndo())
	require.NoError(t, restored.Undo())
	assert.Len(t, restored.GetWaste(), wasteBefore-1)
	assert.NotEmpty(t, restored.GetTableau()[0], "the snapshot must carry the board, not a blank")
}

func TestDiplomat_UnmarshalJSON_Rejections(t *testing.T) {
	for _, tc := range []struct{ name, data string }{
		{"broken json", `{`},
		{"phase too low", `{"ps":-1}`},
		{"phase too high", `{"ps":99}`},
		{"negative move count", `{"mc":-1}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Error(t, json.Unmarshal([]byte(tc.data), NewDefaultDiplomat()))
		})
	}
}

func TestDiplomat_UnmarshalJSON_RejectsOversizedArrays(t *testing.T) {
	big := make([]*Card, DiplomatTotalCards+1)
	for i := range big {
		big[i] = NewCard(CardDesignSpade, 1, true)
	}
	overCap := make([]*Card, diplomatMaxSliceLen+1)
	for i := range overCap {
		overCap[i] = NewCard(CardDesignSpade, 1, true)
	}

	t.Run("stock", func(t *testing.T) {
		data, err := json.Marshal(&diplomatJSON{Stock: big})
		require.NoError(t, err)
		assert.Error(t, json.Unmarshal(data, NewDefaultDiplomat()))
	})
	t.Run("waste", func(t *testing.T) {
		data, err := json.Marshal(&diplomatJSON{Waste: big})
		require.NoError(t, err)
		assert.Error(t, json.Unmarshal(data, NewDefaultDiplomat()))
	})
	t.Run("tableau column", func(t *testing.T) {
		j := &diplomatJSON{}
		j.Tableau[0] = big
		data, err := json.Marshal(j)
		require.NoError(t, err)
		assert.Error(t, json.Unmarshal(data, NewDefaultDiplomat()))
	})
	t.Run("foundation pile", func(t *testing.T) {
		j := &diplomatJSON{}
		j.Foundation[0] = make([]*Card, DiplomatFoundationTarget+1)
		for i := range j.Foundation[0] {
			j.Foundation[0][i] = NewCard(CardDesignSpade, 1, true)
		}
		data, err := json.Marshal(j)
		require.NoError(t, err)
		assert.Error(t, json.Unmarshal(data, NewDefaultDiplomat()))
	})
	t.Run("action log", func(t *testing.T) {
		data, err := json.Marshal(&diplomatJSON{ActionLog: make([]*ActionLogEntry, diplomatMaxSliceLen+1)})
		require.NoError(t, err)
		assert.Error(t, json.Unmarshal(data, NewDefaultDiplomat()))
	})
	t.Run("history", func(t *testing.T) {
		j := &diplomatJSON{History: make([]*diplomatSnapshot, diplomatMaxSliceLen+1)}
		for i := range j.History {
			j.History[i] = &diplomatSnapshot{}
		}
		data, err := json.Marshal(j)
		require.NoError(t, err)
		assert.Error(t, json.Unmarshal(data, NewDefaultDiplomat()))
	})
	t.Run("snapshot stock", func(t *testing.T) {
		data, err := json.Marshal(diplomatSnapshotJSON{Stock: overCap})
		require.NoError(t, err)
		var s diplomatSnapshot
		assert.Error(t, s.UnmarshalJSON(data))
	})
	t.Run("snapshot waste", func(t *testing.T) {
		data, err := json.Marshal(diplomatSnapshotJSON{Waste: overCap})
		require.NoError(t, err)
		var s diplomatSnapshot
		assert.Error(t, s.UnmarshalJSON(data))
	})
	t.Run("snapshot tableau", func(t *testing.T) {
		j := diplomatSnapshotJSON{}
		j.Tableau[0] = overCap
		data, err := json.Marshal(j)
		require.NoError(t, err)
		var s diplomatSnapshot
		assert.Error(t, s.UnmarshalJSON(data))
	})
	t.Run("snapshot foundation", func(t *testing.T) {
		j := diplomatSnapshotJSON{}
		j.Foundation[0] = overCap
		data, err := json.Marshal(j)
		require.NoError(t, err)
		var s diplomatSnapshot
		assert.Error(t, s.UnmarshalJSON(data))
	})
	t.Run("snapshot broken json", func(t *testing.T) {
		var s diplomatSnapshot
		assert.Error(t, s.UnmarshalJSON([]byte(`{`)))
	})
}
