//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestSlyFox() *SlyFox {
	c := NewDefaultSlyFox()
	c.Reset()
	return c
}

// clearSlyFoxBoard wipes the dealt layout so a test can state exactly the
// position it cares about. Never assert on a freshly Reset board -- the deal is
// shuffled, so any such assertion is a hidden flake.
func clearSlyFoxBoard(c *SlyFox) {
	c.stock = nil
	// **配り切った状態から始める。**0 にすると、テストが組みたい局面ではなく
	// 「まだ 20 枚配っていない盤」になり、リザーブが閉じたままになる。
	c.dealtThisCycle = SlyFoxDealCycle
	c.isStalemate = false
	c.history = nil
	c.moveCount = 0
	c.phase = SlyFoxPhasePlaying
	for i := range SlyFoxFoundationCnt {
		c.foundation[i] = nil
	}
	for i := range SlyFoxTableauCnt {
		c.tableau[i] = nil
	}
}

// startSlyFoxCycle は「この周をまだ 1 枚も配っていない」状態にする。
// clearSlyFoxBoard は配り切った状態から始めるので、閉じているところを見たい
// テストだけが明示的にこれを呼ぶ。
func startSlyFoxCycle(c *SlyFox) { c.dealtThisCycle = 0 }

// fillSlyFoxPiles puts one card in every tableau pile so no gap exists. A gap
// changes which hint the game offers, so a board full of holes would test a
// position the game never reaches by accident.
func fillSlyFoxPiles(c *SlyFox, card func() *Card) {
	for i := range SlyFoxTableauCnt {
		c.tableau[i] = []*Card{card()}
	}
}

func TestNewSlyFox(t *testing.T) {
	assert.NotNil(t, NewSlyFox(NewTrumpCardsWithDecks(2, 0)))
	assert.NotNil(t, NewDefaultSlyFox())
}

// The deal is 20 piles of ONE card with the other 84 as stock.
func TestSlyFox_Reset(t *testing.T) {
	c := newTestSlyFox()

	for i, pile := range c.GetTableau() {
		assert.Len(t, pile, 1, "pile %d", i)
	}
	assert.Equal(t, SlyFoxTotalCards-SlyFoxTableauCnt, c.GetStockCount())
	assert.Equal(t, 84, c.GetStockCount())
	// 開幕は 20 枚配り終えた形。0 から始めると、1 枚も収穫できないまま
	// 20 枚配らされる盤になる。
	assert.Equal(t, SlyFoxDealCycle, c.DealtThisCycle())
	assert.False(t, c.ReserveIsLocked())

	// Foundations start EMPTY. #5277 asks for 16 of them, which cannot be right:
	// 16*13 = 208 cards for a 104-card game. Eight (four up, four down) is the
	// only count that consumes the deck exactly.
	for i, pile := range c.GetFoundation() {
		assert.Empty(t, pile, "foundation %d", i)
	}
	assert.Equal(t, SlyFoxTotalCards, SlyFoxFoundationCnt*SlyFoxFoundationTarget)

	assert.Equal(t, SlyFoxPhasePlaying, c.GetPhase())
	assert.Equal(t, 0, c.GetMoveCount())
	assert.True(t, c.AllFaceUp())
	assert.False(t, c.GetGameEndFlag())
	assert.False(t, c.IsStalemate())
}

func TestSlyFox_ResetDealsEveryCard(t *testing.T) {
	for range 20 {
		c := newTestSlyFox()
		total := c.GetStockCount()
		for _, pile := range c.GetTableau() {
			total += len(pile)
		}
		for _, pile := range c.GetFoundation() {
			total += len(pile)
		}
		assert.Equal(t, SlyFoxTotalCards, total)
	}
}

func TestSlyFox_ResetTwiceIsClean(t *testing.T) {
	c := newTestSlyFox()
	require.NoError(t, c.DealToPile(0))
	c.Reset()
	assert.Equal(t, SlyFoxDealCycle, c.DealtThisCycle())
	assert.Equal(t, 0, c.GetMoveCount())
	assert.False(t, c.CanUndo())
}

// --- Foundation direction ---

func TestSlyFox_IsAscendingFoundation(t *testing.T) {
	c := newTestSlyFox()
	for i := range SlyFoxAscendingCnt {
		assert.True(t, c.IsAscendingFoundation(i), "foundation %d should build up", i)
	}
	for i := SlyFoxAscendingCnt; i < SlyFoxFoundationCnt; i++ {
		assert.False(t, c.IsAscendingFoundation(i), "foundation %d should build down", i)
	}
	assert.False(t, c.IsAscendingFoundation(-1))
	assert.False(t, c.IsAscendingFoundation(SlyFoxFoundationCnt))
}

// Each suit has one ascending and one descending foundation, and the two share
// the same suit order so the board layout is stable across deals.
func TestSlyFox_FoundationSuitPairing(t *testing.T) {
	for i := range SlyFoxAscendingCnt {
		assert.Equal(t, slyFoxSuitOrder[i], slyFoxSuitOrder[i+SlyFoxAscendingCnt],
			"foundation %d and %d must cover the same suit", i, i+SlyFoxAscendingCnt)
	}
}

func TestSlyFox_CanPlaceOnFoundation_Ascending(t *testing.T) {
	c := newTestSlyFox()
	clearSlyFoxBoard(c)
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

func TestSlyFox_CanPlaceOnFoundation_Descending(t *testing.T) {
	c := newTestSlyFox()
	clearSlyFoxBoard(c)
	// Foundation 4 is spades, descending: it starts at K.
	assert.True(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, CardValueMax, true), 4))
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 1, true), 4))
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, CardValueMax-1, true), 4))

	c.foundation[4] = []*Card{NewCard(CardDesignSpade, CardValueMax, true)}
	assert.True(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, CardValueMax-1, true), 4))
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, CardValueMax, true), 4))
}

func TestSlyFox_CanPlaceOnFoundation_RejectsFullPileAndBadIndex(t *testing.T) {
	c := newTestSlyFox()
	clearSlyFoxBoard(c)
	for v := 1; v <= CardValueMax; v++ {
		c.foundation[0] = append(c.foundation[0], NewCard(CardDesignSpade, v, true))
	}
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 1, true), 0))
	assert.False(t, c.canPlaceOnFoundation(nil, 0))
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 1, true), -1))
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 1, true), SlyFoxFoundationCnt))
}

// A card that fits both its ascending and its descending foundation may go to
// either -- two copies exist per suit+rank, one for each direction.
func TestSlyFox_FindFoundation_AcceptsEitherDirection(t *testing.T) {
	c := newTestSlyFox()
	clearSlyFoxBoard(c)
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

func TestSlyFox_FindFoundation_NoHome(t *testing.T) {
	c := newTestSlyFox()
	clearSlyFoxBoard(c)
	assert.Equal(t, -1, c.findFoundation(NewCard(CardDesignSpade, 7, true)))
	assert.Equal(t, -1, c.findFoundation(nil))
}

// --- Dealing ---

func TestSlyFox_MoveTableauToFoundation(t *testing.T) {
	c := newTestSlyFox()
	clearSlyFoxBoard(c)
	fillSlyFoxPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })
	c.tableau[3] = []*Card{NewCard(CardDesignSpade, 1, true)}

	require.NoError(t, c.MoveTableauToFoundation(3))
	assert.Len(t, c.GetFoundation()[0], 1)
	assert.Empty(t, c.GetTableau()[3])
	assert.Equal(t, 1, c.GetMoveCount())
}

func TestSlyFox_MoveTableauToFoundation_Descending(t *testing.T) {
	c := newTestSlyFox()
	clearSlyFoxBoard(c)
	fillSlyFoxPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })
	c.tableau[5] = []*Card{NewCard(CardDesignHeart, CardValueMax, true)}

	require.NoError(t, c.MoveTableauToFoundation(5))
	// Hearts descending is foundation index 6 (2 + SlyFoxAscendingCnt).
	assert.Len(t, c.GetFoundation()[6], 1)
	assert.Equal(t, CardValueMax, c.GetFoundation()[6][0].GetValue())
}

func TestSlyFox_MoveTableauToFoundation_Rejections(t *testing.T) {
	c := newTestSlyFox()
	clearSlyFoxBoard(c)
	fillSlyFoxPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })

	// **拒否はコードを名乗ること。**生の英文だと日本語ロケールの CUI が
	// 英語のまま表示する。
	for _, tc := range []struct {
		name, code string
		act        func() error
	}{
		{"index below range", "slyfox.errInvalidPile", func() error { return c.MoveTableauToFoundation(-1) }},
		{"index above range", "slyfox.errInvalidPile", func() error {
			return c.MoveTableauToFoundation(SlyFoxTableauCnt)
		}},
		{"empty pile", "slyfox.errPileEmpty", func() error {
			c.tableau[2] = nil
			return c.MoveTableauToFoundation(2)
		}},
		// A 7 with nothing under it has no home.
		{"no foundation wants it", "slyfox.errCannotPlaceFoundation", func() error {
			return c.MoveTableauToFoundation(0)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.act()
			require.Error(t, err)
			code, _ := ErrorMessageCode(err)
			assert.Equal(t, tc.code, code)
		})
	}

	c.GiveUp()
	assert.Error(t, c.MoveTableauToFoundation(0))
}

// --- Burying a card under another ---

func TestSlyFox_BuriedCardCannotReachFoundation(t *testing.T) {
	c := newTestSlyFox()
	clearSlyFoxBoard(c)
	fillSlyFoxPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })
	c.tableau[1] = []*Card{NewCard(CardDesignClover, 1, true)}
	c.stock = []*Card{NewCard(CardDesignHeart, 9, true)}

	// 配った札は下の札を埋める。埋めた枠の A はもう届かない。
	require.NoError(t, c.DealToPile(1))
	err := c.MoveTableauToFoundation(1)
	require.Error(t, err, "the ace is under the 9 now")
	code, _ := ErrorMessageCode(err)
	assert.Equal(t, "slyfox.errCannotPlaceFoundation", code)
}

// --- Stock -> tableau ---

func TestSlyFox_BuryCost(t *testing.T) {
	c := newTestSlyFox()
	clearSlyFoxBoard(c)

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
	assert.Equal(t, SlyFoxFoundationTarget+1, c.buryCost(NewCard(CardDesignSpade, 3, true)))
	assert.Equal(t, SlyFoxFoundationTarget+1, c.buryCost(nil))
}

func TestSlyFox_BestBuryPile_PrefersEmptyPile(t *testing.T) {
	c := newTestSlyFox()
	clearSlyFoxBoard(c)
	fillSlyFoxPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })
	c.tableau[13] = nil
	assert.Equal(t, 13, c.bestBuryPile())
}

func TestSlyFox_BestBuryPile_PicksTheLeastNeededCard(t *testing.T) {
	c := newTestSlyFox()
	clearSlyFoxBoard(c)
	// Spades ascending is at 3 (wants the 4) and spades descending is at 2 (wants
	// the Ace). Every pile holds an Ace, which the descending pile wants right
	// now, except pile 8 -- its 2 has been passed by both directions.
	fillSlyFoxPiles(c, func() *Card { return NewCard(CardDesignSpade, 1, true) })
	for v := 1; v <= 3; v++ {
		c.foundation[0] = append(c.foundation[0], NewCard(CardDesignSpade, v, true))
	}
	for v := CardValueMax; v >= 2; v-- {
		c.foundation[4] = append(c.foundation[4], NewCard(CardDesignSpade, v, true))
	}
	c.tableau[8] = []*Card{NewCard(CardDesignSpade, 2, true)}

	assert.Equal(t, 0, c.buryCost(NewCard(CardDesignSpade, 1, true)), "the aces are wanted now")
	assert.Equal(t, SlyFoxFoundationTarget+1, c.buryCost(NewCard(CardDesignSpade, 2, true)))
	assert.Equal(t, 8, c.bestBuryPile(), "bury the card no foundation can ever take")
}

func TestSlyFox_BestBuryPile_TieGoesToTheLowestIndex(t *testing.T) {
	c := newTestSlyFox()
	clearSlyFoxBoard(c)
	fillSlyFoxPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })
	assert.Equal(t, 0, c.bestBuryPile())
}

// --- Hints ---

func TestSlyFox_GetHint_PrefersFoundation(t *testing.T) {
	c := newTestSlyFox()
	clearSlyFoxBoard(c)
	fillSlyFoxPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })
	c.tableau[6] = []*Card{NewCard(CardDesignClover, 1, true)}
	c.stock = []*Card{NewCard(CardDesignDiamond, 4, true)}

	h := c.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "tableau", h.FromZone)
	assert.Equal(t, 6, h.FromIdx)
	assert.Equal(t, "foundation", h.ToZone)
	assert.Equal(t, 1, h.ToIdx)
}

func TestSlyFox_GetHint_NilWhenNotPlaying(t *testing.T) {
	c := newTestSlyFox()
	c.GiveUp()
	assert.Nil(t, c.GetHint())
	assert.Nil(t, c.foundationHint())
}

// --- Stalemate ---

func TestSlyFox_Stalemate_NotSetOutsidePlaying(t *testing.T) {
	c := newTestSlyFox()
	clearSlyFoxBoard(c)
	c.phase = SlyFoxPhaseGameOver
	c.checkStalemate()
	assert.False(t, c.IsStalemate())
}

func TestSlyFox_UndoToEscape(t *testing.T) {
	c := newTestSlyFox()
	clearSlyFoxBoard(c)
	fillSlyFoxPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })
	assert.Equal(t, 0, c.UndoToEscape(), "not stuck yet")

	c.stock = []*Card{NewCard(CardDesignHeart, 9, true)}
	require.NoError(t, c.DealToPile(0))
	assert.True(t, c.IsStalemate())
	assert.Equal(t, 1, c.UndoToEscape())
}

// --- Game clear ---

func TestSlyFox_GameClear(t *testing.T) {
	c := newTestSlyFox()
	clearSlyFoxBoard(c)
	// Fill every foundation but leave one card off the last one.
	for i := range SlyFoxFoundationCnt {
		for n := range SlyFoxFoundationTarget {
			v := n + 1
			if !slyFoxIsAscending(i) {
				v = CardValueMax - n
			}
			c.foundation[i] = append(c.foundation[i], NewCard(slyFoxSuitOrder[i], v, true))
		}
	}
	c.foundation[7] = c.foundation[7][:SlyFoxFoundationTarget-1]
	c.checkGameClear()
	assert.Equal(t, SlyFoxPhasePlaying, c.GetPhase(), "one card still missing")

	// Diamonds descending needs its Ace last.
	c.tableau[0] = []*Card{NewCard(CardDesignDiamond, 1, true)}
	require.NoError(t, c.MoveTableauToFoundation(0))
	assert.Equal(t, SlyFoxPhaseGameClear, c.GetPhase())
	assert.True(t, c.GetGameEndFlag())
}

// --- AutoComplete ---

func TestSlyFox_AutoComplete(t *testing.T) {
	c := newTestSlyFox()
	clearSlyFoxBoard(c)
	fillSlyFoxPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })
	c.tableau[0] = []*Card{NewCard(CardDesignClover, 2, true), NewCard(CardDesignClover, 1, true)}
	c.tableau[2] = []*Card{NewCard(CardDesignHeart, 1, true)}

	require.NoError(t, c.AutoComplete())
	assert.Len(t, c.GetFoundation()[1], 2, "the ace then the deuce under it")
	assert.Len(t, c.GetFoundation()[2], 1, "the heart ace too")
}

func TestSlyFox_AutoComplete_Rejections(t *testing.T) {
	c := newTestSlyFox()
	clearSlyFoxBoard(c)
	fillSlyFoxPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })
	err := c.AutoComplete()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no card")

	c.GiveUp()
	assert.Error(t, c.AutoComplete())
}

// --- Undo ---

func TestSlyFox_Undo(t *testing.T) {
	c := newTestSlyFox()
	clearSlyFoxBoard(c)
	fillSlyFoxPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })
	c.stock = []*Card{NewCard(CardDesignHeart, 9, true)}

	assert.False(t, c.CanUndo())
	require.Error(t, c.Undo())

	require.NoError(t, c.DealToPile(3))
	assert.True(t, c.CanUndo())
	require.NoError(t, c.Undo())
	assert.Equal(t, 1, c.GetStockCount(), "配った札は山札へ戻る")
	assert.Len(t, c.GetTableau()[3], 1)
	assert.Equal(t, 0, c.GetMoveCount())
}

func TestSlyFox_UndoN(t *testing.T) {
	c := newTestSlyFox()
	clearSlyFoxBoard(c)
	fillSlyFoxPiles(c, func() *Card { return NewCard(CardDesignSpade, 7, true) })
	c.stock = []*Card{
		NewCard(CardDesignHeart, 9, true),
		NewCard(CardDesignHeart, 8, true),
		NewCard(CardDesignHeart, 6, true),
	}
	require.NoError(t, c.DealToPile(0))
	require.NoError(t, c.DealToPile(1))
	require.NoError(t, c.DealToPile(2))
	assert.Equal(t, 3, c.GetMoveCount())

	require.NoError(t, c.UndoN(2))
	assert.Equal(t, 1, c.GetMoveCount())
	assert.Equal(t, 2, c.GetStockCount(), "戻した 2 枚は山札へ返る")

	assert.Error(t, c.UndoN(5))
	assert.Error(t, c.UndoN(0))
}

func TestSlyFox_GiveUp(t *testing.T) {
	c := newTestSlyFox()
	c.GiveUp()
	assert.Equal(t, SlyFoxPhaseGameOver, c.GetPhase())
	assert.True(t, c.GetGameEndFlag())
	assert.NotEmpty(t, c.GetActionLog())

	// A second give-up must not append another entry.
	before := len(c.GetActionLog())
	c.GiveUp()
	assert.Len(t, c.GetActionLog(), before)
}

func TestSlyFox_ActionLog(t *testing.T) {
	c := newTestSlyFox()
	require.NoError(t, c.DealToPile(0))
	log := c.GetActionLog()
	require.NotEmpty(t, log)
	assert.Equal(t, "deal", log[len(log)-1].ActionType)
}

// --- JSON round-trip ---

func TestSlyFox_JSONRoundTrip(t *testing.T) {
	c := newTestSlyFox()
	require.NoError(t, c.DealToPile(0))
	require.NoError(t, c.DealToPile(1))

	data, err := json.Marshal(c)
	require.NoError(t, err)

	restored := NewDefaultSlyFox()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, c.GetStockCount(), restored.GetStockCount())
	// **周のカウンタも往復すること。**KV から戻した盤で数え直しになると、
	// リザーブが開くタイミングがずれる。
	assert.Equal(t, c.DealtThisCycle(), restored.DealtThisCycle())
	assert.Equal(t, c.GetMoveCount(), restored.GetMoveCount())
	assert.Equal(t, c.GetPhase(), restored.GetPhase())
	assert.Equal(t, c.IsStalemate(), restored.IsStalemate())
	for i := range SlyFoxTableauCnt {
		assert.Len(t, restored.GetTableau()[i], len(c.GetTableau()[i]), "pile %d", i)
	}
}

// The undo stack has to survive the trip: the Worker rebuilds the game from KV
// on every request, so an unpersisted history means Undo silently never works.
func TestSlyFox_JSONRoundTripKeepsUndoHistory(t *testing.T) {
	c := newTestSlyFox()
	require.NoError(t, c.DealToPile(0))
	require.NoError(t, c.DealToPile(0))
	pileBefore := len(c.GetTableau()[0])

	data, err := json.Marshal(c)
	require.NoError(t, err)
	restored := NewDefaultSlyFox()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.True(t, restored.CanUndo())
	require.NoError(t, restored.Undo())
	assert.Len(t, restored.GetTableau()[0], pileBefore-1)
	assert.NotEmpty(t, restored.GetTableau()[0], "the snapshot must carry the board, not a blank")
}

func TestSlyFox_UnmarshalJSON_Rejections(t *testing.T) {
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
			assert.Error(t, json.Unmarshal([]byte(tt.data), NewDefaultSlyFox()))
		})
	}
}

func TestSlyFox_UnmarshalJSON_RejectsOversizedArrays(t *testing.T) {
	bigCards := make([]*Card, SlyFoxTotalCards+1)
	for i := range bigCards {
		bigCards[i] = NewCard(CardDesignSpade, 1, true)
	}
	overCap := make([]*Card, slyFoxMaxSliceLen+1)
	for i := range overCap {
		overCap[i] = NewCard(CardDesignSpade, 1, true)
	}

	t.Run("stock", func(t *testing.T) {
		data, err := json.Marshal(&slyFoxJSON{Stock: bigCards})
		require.NoError(t, err)
		assert.Error(t, json.Unmarshal(data, NewDefaultSlyFox()))
	})
	// **周のカウンタも値域を見る。**KV には前の版が書いた任意のバイト列が入る。
	t.Run("deal count out of range", func(t *testing.T) {
		for _, n := range []int{-1, SlyFoxDealCycle + 1} {
			data, err := json.Marshal(&slyFoxJSON{DealtThisCycle: n})
			require.NoError(t, err)
			assert.Error(t, json.Unmarshal(data, NewDefaultSlyFox()), "dealt=%d", n)
		}
	})
	t.Run("tableau pile", func(t *testing.T) {
		j := &slyFoxJSON{}
		j.Tableau[0] = bigCards
		data, err := json.Marshal(j)
		require.NoError(t, err)
		assert.Error(t, json.Unmarshal(data, NewDefaultSlyFox()))
	})
	t.Run("foundation pile", func(t *testing.T) {
		j := &slyFoxJSON{}
		j.Foundation[0] = make([]*Card, SlyFoxFoundationTarget+1)
		for i := range j.Foundation[0] {
			j.Foundation[0][i] = NewCard(CardDesignSpade, 1, true)
		}
		data, err := json.Marshal(j)
		require.NoError(t, err)
		assert.Error(t, json.Unmarshal(data, NewDefaultSlyFox()))
	})
	t.Run("action log", func(t *testing.T) {
		j := &slyFoxJSON{ActionLog: make([]*ActionLogEntry, slyFoxMaxSliceLen+1)}
		data, err := json.Marshal(j)
		require.NoError(t, err)
		assert.Error(t, json.Unmarshal(data, NewDefaultSlyFox()))
	})
	t.Run("history", func(t *testing.T) {
		j := &slyFoxJSON{History: make([]*slyFoxSnapshot, slyFoxMaxSliceLen+1)}
		for i := range j.History {
			j.History[i] = &slyFoxSnapshot{}
		}
		data, err := json.Marshal(j)
		require.NoError(t, err)
		assert.Error(t, json.Unmarshal(data, NewDefaultSlyFox()))
	})
	t.Run("snapshot stock", func(t *testing.T) {
		var s slyFoxSnapshot
		data, err := json.Marshal(slyFoxSnapshotJSON{Stock: overCap})
		require.NoError(t, err)
		assert.Error(t, s.UnmarshalJSON(data))
	})
	t.Run("snapshot tableau", func(t *testing.T) {
		var s slyFoxSnapshot
		j := slyFoxSnapshotJSON{}
		j.Tableau[0] = overCap
		data, err := json.Marshal(j)
		require.NoError(t, err)
		assert.Error(t, s.UnmarshalJSON(data))
	})
	t.Run("snapshot foundation", func(t *testing.T) {
		var s slyFoxSnapshot
		j := slyFoxSnapshotJSON{}
		j.Foundation[0] = overCap
		data, err := json.Marshal(j)
		require.NoError(t, err)
		assert.Error(t, s.UnmarshalJSON(data))
	})
	t.Run("snapshot broken json", func(t *testing.T) {
		var s slyFoxSnapshot
		assert.Error(t, s.UnmarshalJSON([]byte(`{`)))
	})
}
