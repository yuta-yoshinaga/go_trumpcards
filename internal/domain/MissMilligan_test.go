//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestMissMilligan() *MissMilligan {
	mm := NewMissMilligan(NewTrumpCardsWithDecks(2, 0))
	mm.Reset()
	return mm
}

// setMMBoard installs an exact position. Tests must never lean on the shuffle:
// asserting what is or is not playable against a real deal is a flake waiting
// to happen (see #4467).
func setMMBoard(mm *MissMilligan, foundation [MissMilliganFoundationCnt][]*Card, cols [][]*Card, stock, waived []*Card) {
	mm.foundation = foundation
	for i := range MissMilliganTableauCnt {
		mm.tableau[i] = nil
	}
	for i, c := range cols {
		pile := make([]*MissMilliganTableauCard, 0, len(c))
		for _, card := range c {
			pile = append(pile, &MissMilliganTableauCard{Card: card, FaceUp: true})
		}
		mm.tableau[i] = pile
	}
	mm.stock = stock
	mm.waived = waived
	mm.phase = MissMilliganPhasePlaying
	mm.isStalemate = false
	mm.history = nil
	mm.moveCount = 0
	mm.actionLog = nil
}

func TestMissMilligan_Reset_DealsEightAndKeepsTheRestAsStock(t *testing.T) {
	mm := newTestMissMilligan()

	dealt := 0
	for col, pile := range mm.GetTableau() {
		require.Len(t, pile, 1, "column %d starts with one card", col)
		assert.True(t, pile[0].FaceUp, "every card is face-up by rule")
		dealt += len(pile)
	}
	assert.Equal(t, MissMilliganTableauCnt, dealt)
	assert.Equal(t, missMilliganTotalCards-MissMilliganTableauCnt, mm.GetStockCount())
	for _, f := range mm.GetFoundation() {
		assert.Empty(t, f, "foundations start empty; only Aces open them")
	}
	assert.Empty(t, mm.GetWaived())
	assert.True(t, mm.AllFaceUp())
	assert.False(t, mm.GetGameEndFlag())
}

// The stock is dealt eight at a time, one per column. #4407's rule 5 describes
// a one-card-at-a-time deal into a waste; Miss Milligan has no waste at all.
func TestMissMilligan_DealPutsOneCardOnEveryColumn(t *testing.T) {
	mm := newTestMissMilligan()
	before := mm.GetStockCount()

	require.NoError(t, mm.Deal())
	assert.Equal(t, before-MissMilliganTableauCnt, mm.GetStockCount())
	for col, pile := range mm.GetTableau() {
		assert.Len(t, pile, 2, "column %d received one more card", col)
	}

	var empty [MissMilliganFoundationCnt][]*Card
	setMMBoard(mm, empty, nil, nil, nil)
	assert.Error(t, mm.Deal(), "stock is empty")
}

func TestMissMilligan_DealIsBlockedWhileCardsAreWaived(t *testing.T) {
	mm := newTestMissMilligan()
	var empty [MissMilliganFoundationCnt][]*Card
	setMMBoard(mm, empty, [][]*Card{{NewCard(CardDesignSpade, 5, true)}},
		[]*Card{NewCard(CardDesignHeart, 9, true)}, []*Card{NewCard(CardDesignClover, 4, true)})

	// Dealing while holding could bury the only square the held card fits.
	assert.Error(t, mm.Deal(), "return the waived cards first")
}

func TestMissMilligan_TableauBuildsDownInAlternatingColour(t *testing.T) {
	mm := newTestMissMilligan()
	var empty [MissMilliganFoundationCnt][]*Card
	setMMBoard(mm, empty, [][]*Card{
		{NewCard(CardDesignSpade, 9, true)},
		{NewCard(CardDesignHeart, 8, true)},
		{NewCard(CardDesignClover, 8, true)},
		{NewCard(CardDesignHeart, 10, true)},
	}, nil, nil)

	assert.Error(t, mm.MoveTableauToTableau(2, 0, 0), "♣8 on ♠9 is the same colour")
	assert.Error(t, mm.MoveTableauToTableau(3, 0, 0), "builds down, not up")
	require.NoError(t, mm.MoveTableauToTableau(1, 0, 0), "♥8 onto ♠9")
	assert.Len(t, mm.GetTableau()[0], 2)
	assert.Empty(t, mm.GetTableau()[1])
}

// #4407 never mentions this, but it is the real rule and without it an empty
// column becomes a universal parking space and the game falls apart.
func TestMissMilligan_EmptyColumnTakesOnlyAKing(t *testing.T) {
	mm := newTestMissMilligan()
	var empty [MissMilliganFoundationCnt][]*Card
	setMMBoard(mm, empty, [][]*Card{
		{NewCard(CardDesignSpade, 9, true)},
		{NewCard(CardDesignSpade, CardValueMax, true)},
	}, nil, nil)

	assert.Error(t, mm.MoveTableauToTableau(0, 0, 5), "♠9 cannot start an empty column")
	require.NoError(t, mm.MoveTableauToTableau(1, 0, 5), "a King can")
	assert.Len(t, mm.GetTableau()[5], 1)
}

func TestMissMilligan_MovesAnAlternatingRunAsOneUnit(t *testing.T) {
	mm := newTestMissMilligan()
	var empty [MissMilliganFoundationCnt][]*Card
	setMMBoard(mm, empty, [][]*Card{
		{
			NewCard(CardDesignHeart, 2, true), // buried, breaks the run
			NewCard(CardDesignSpade, 7, true),
			NewCard(CardDesignHeart, 6, true),
			NewCard(CardDesignClover, 5, true),
		},
		{NewCard(CardDesignHeart, 8, true)},
	}, nil, nil)

	assert.Error(t, mm.MoveTableauToTableau(0, 0, 1), "♥2 is not part of the run")
	require.NoError(t, mm.MoveTableauToTableau(0, 1, 1))
	assert.Len(t, mm.GetTableau()[1], 4, "♥8 plus the three-card run")
	require.Len(t, mm.GetTableau()[0], 1)

	got := mm.GetTableau()[1]
	assert.Equal(t, []int{8, 7, 6, 5}, []int{
		got[0].Card.GetValue(), got[1].Card.GetValue(),
		got[2].Card.GetValue(), got[3].Card.GetValue(),
	}, "order is preserved, not reversed")
}

func TestMissMilligan_RejectsASameColourRun(t *testing.T) {
	mm := newTestMissMilligan()
	var empty [MissMilliganFoundationCnt][]*Card
	setMMBoard(mm, empty, [][]*Card{
		{NewCard(CardDesignSpade, 7, true), NewCard(CardDesignClover, 6, true)},
		{NewCard(CardDesignHeart, 8, true)},
	}, nil, nil)

	assert.Error(t, mm.MoveTableauToTableau(0, 0, 1), "♠7,♣6 are both black")
}

func TestMissMilligan_CardIndexMinusOneMeansTheTopCard(t *testing.T) {
	mm := newTestMissMilligan()
	var empty [MissMilliganFoundationCnt][]*Card
	setMMBoard(mm, empty, [][]*Card{
		{NewCard(CardDesignHeart, 2, true), NewCard(CardDesignSpade, 7, true)},
		{NewCard(CardDesignHeart, 8, true)},
	}, nil, nil)

	require.NoError(t, mm.MoveTableauToTableau(0, -1, 1))
	assert.Len(t, mm.GetTableau()[1], 2, "only ♠7 moved")
	assert.Len(t, mm.GetTableau()[0], 1)
}

func TestMissMilligan_RejectsInvalidTableauArguments(t *testing.T) {
	mm := newTestMissMilligan()
	var empty [MissMilliganFoundationCnt][]*Card
	setMMBoard(mm, empty, [][]*Card{{NewCard(CardDesignSpade, 9, true)}}, nil, nil)

	assert.Error(t, mm.MoveTableauToTableau(-1, 0, 1), "from out of range")
	assert.Error(t, mm.MoveTableauToTableau(0, 0, MissMilliganTableauCnt), "to out of range")
	assert.Error(t, mm.MoveTableauToTableau(0, 0, 0), "same column")
	assert.Error(t, mm.MoveTableauToTableau(0, 5, 1), "card index past the pile")
	assert.Error(t, mm.MoveTableauToTableau(0, -2, 1), "negative card index")
	assert.Error(t, mm.MoveTableauToFoundation(-1), "column out of range")
	assert.Error(t, mm.MoveTableauToFoundation(1), "column is empty")
}

func TestMissMilligan_FoundationIsPinnedToItsSuit(t *testing.T) {
	mm := newTestMissMilligan()
	var f [MissMilliganFoundationCnt][]*Card
	setMMBoard(mm, f, [][]*Card{
		{NewCard(CardDesignSpade, 1, true)},
		{NewCard(CardDesignHeart, 5, true)},
	}, nil, nil)

	require.NoError(t, mm.MoveTableauToFoundation(0))
	assert.Len(t, mm.GetFoundation()[0], 1, "♠A opened the first spade pile")
	for i := 1; i < MissMilliganFoundationCnt; i++ {
		assert.Empty(t, mm.GetFoundation()[i], "no other pile took the spade Ace")
	}

	assert.Error(t, mm.MoveTableauToFoundation(1), "♥5 does not open a heart pile")
}

// With two decks each suit has two foundations; the second Ace opens the second.
func TestMissMilligan_SecondDeckUsesTheSecondFoundationOfThatSuit(t *testing.T) {
	mm := newTestMissMilligan()
	var f [MissMilliganFoundationCnt][]*Card
	f[0] = []*Card{NewCard(CardDesignSpade, 1, true), NewCard(CardDesignSpade, 2, true)}
	setMMBoard(mm, f, [][]*Card{{NewCard(CardDesignSpade, 1, true)}}, nil, nil)

	require.NoError(t, mm.MoveTableauToFoundation(0))
	assert.Len(t, mm.GetFoundation()[0], 2, "the first pile is untouched")
	assert.Len(t, mm.GetFoundation()[4], 1, "the Ace opened the second spade pile")
}

// Waiving is the game's rescue rule, and it is deliberately restricted.
func TestMissMilligan_WaiveOnlyOnceTheStockIsGone(t *testing.T) {
	mm := newTestMissMilligan()
	var empty [MissMilliganFoundationCnt][]*Card
	setMMBoard(mm, empty, [][]*Card{{NewCard(CardDesignSpade, 5, true)}},
		[]*Card{NewCard(CardDesignHeart, 9, true)}, nil)

	assert.False(t, mm.CanWaive())
	assert.Error(t, mm.Waive(0, -1), "the stock still has cards")

	setMMBoard(mm, empty, [][]*Card{{NewCard(CardDesignSpade, 5, true)}}, nil, nil)
	assert.True(t, mm.CanWaive())
	require.NoError(t, mm.Waive(0, -1))
	assert.Len(t, mm.GetWaived(), 1)
	assert.Empty(t, mm.GetTableau()[0])
	assert.False(t, mm.CanWaive(), "only one set may be held at a time")
	assert.Error(t, mm.Waive(1, -1), "cards are already waived")
}

func TestMissMilligan_WaiveTakesAWholeRun(t *testing.T) {
	mm := newTestMissMilligan()
	var empty [MissMilliganFoundationCnt][]*Card
	setMMBoard(mm, empty, [][]*Card{
		{
			NewCard(CardDesignHeart, 2, true),
			NewCard(CardDesignSpade, 7, true),
			NewCard(CardDesignHeart, 6, true),
		},
	}, nil, nil)

	assert.Error(t, mm.Waive(0, 0), "♥2,♠7 is not a run")
	assert.Error(t, mm.Waive(0, 9), "index past the pile")
	assert.Error(t, mm.Waive(-1, 0), "column out of range")
	require.NoError(t, mm.Waive(0, 1))
	assert.Len(t, mm.GetWaived(), 2)
	assert.Len(t, mm.GetTableau()[0], 1)
}

func TestMissMilligan_PlaceWaivedBackOntoTheTableau(t *testing.T) {
	mm := newTestMissMilligan()
	var empty [MissMilliganFoundationCnt][]*Card
	setMMBoard(mm, empty, [][]*Card{
		{NewCard(CardDesignSpade, 9, true)},
		{NewCard(CardDesignSpade, 3, true)},
	}, nil, []*Card{NewCard(CardDesignHeart, 8, true)})

	assert.Error(t, mm.PlaceWaived(1), "♥8 does not sit on ♠3")
	assert.Error(t, mm.PlaceWaived(-1), "column out of range")
	require.NoError(t, mm.PlaceWaived(0))
	assert.Empty(t, mm.GetWaived())
	assert.Len(t, mm.GetTableau()[0], 2)

	assert.Error(t, mm.PlaceWaived(0), "nothing is waived any more")
}

func TestMissMilligan_WaivedCardCanGoToAFoundation(t *testing.T) {
	mm := newTestMissMilligan()
	var f [MissMilliganFoundationCnt][]*Card
	f[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
	setMMBoard(mm, f, nil, nil, []*Card{NewCard(CardDesignSpade, 2, true)})

	require.NoError(t, mm.MoveWaivedToFoundation())
	assert.Empty(t, mm.GetWaived())
	assert.Len(t, mm.GetFoundation()[0], 2)

	assert.Error(t, mm.MoveWaivedToFoundation(), "nothing is waived")
}

// A held run cannot be split onto a foundation, so only a lone card may go up.
func TestMissMilligan_AWaivedRunCannotGoToAFoundation(t *testing.T) {
	mm := newTestMissMilligan()
	var f [MissMilliganFoundationCnt][]*Card
	f[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
	setMMBoard(mm, f, nil, nil, []*Card{
		NewCard(CardDesignSpade, 2, true), NewCard(CardDesignHeart, 1, true),
	})

	assert.Error(t, mm.MoveWaivedToFoundation(), "only a single waived card may go up")

	// A lone card that fits nowhere is refused too.
	setMMBoard(mm, f, nil, nil, []*Card{NewCard(CardDesignHeart, 9, true)})
	assert.Error(t, mm.MoveWaivedToFoundation(), "♥9 fits no foundation")
}

func TestMissMilligan_GameClearWhenAllEightFoundationsReachKing(t *testing.T) {
	mm := newTestMissMilligan()
	var f [MissMilliganFoundationCnt][]*Card
	for i, design := range missMilliganSuitOrder {
		pile := make([]*Card, 0, CardValueMax)
		for v := 1; v <= CardValueMax; v++ {
			pile = append(pile, NewCard(design, v, true))
		}
		f[i] = pile
	}
	f[7] = f[7][:CardValueMax-1]
	setMMBoard(mm, f, [][]*Card{{NewCard(CardDesignDiamond, CardValueMax, true)}}, nil, nil)

	require.NoError(t, mm.MoveTableauToFoundation(0))
	assert.Equal(t, MissMilliganPhaseGameClear, mm.GetPhase())
	assert.True(t, mm.GetGameEndFlag())
}

func TestMissMilligan_GiveUpEndsTheGameOnce(t *testing.T) {
	mm := newTestMissMilligan()
	mm.GiveUp()
	assert.Equal(t, MissMilliganPhaseGameOver, mm.GetPhase())
	logLen := len(mm.GetActionLog())

	mm.GiveUp()
	assert.Len(t, mm.GetActionLog(), logLen, "a second give-up is a no-op")

	assert.Error(t, mm.Deal())
	assert.Error(t, mm.MoveTableauToTableau(0, 0, 1))
	assert.Error(t, mm.MoveTableauToFoundation(0))
	assert.Error(t, mm.Waive(0, 0))
	assert.Error(t, mm.PlaceWaived(0))
	assert.Error(t, mm.MoveWaivedToFoundation())
	assert.Error(t, mm.AutoComplete())
	assert.Nil(t, mm.GetHint())
	assert.False(t, mm.CanWaive())
}

func TestMissMilligan_HintPrefersTheFoundation(t *testing.T) {
	mm := newTestMissMilligan()
	var f [MissMilliganFoundationCnt][]*Card
	setMMBoard(mm, f, [][]*Card{
		{NewCard(CardDesignSpade, 9, true)},
		{NewCard(CardDesignHeart, 8, true)}, // a legal tableau move
		{NewCard(CardDesignSpade, 1, true)}, // but the Ace goes up
	}, nil, nil)

	h := mm.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "foundation", h.ToZone)
	assert.Equal(t, "tableau", h.FromZone)
	assert.Equal(t, 2, h.FromCol)
}

func TestMissMilligan_HintFallsBackToTableauThenStock(t *testing.T) {
	mm := newTestMissMilligan()
	var f [MissMilliganFoundationCnt][]*Card

	setMMBoard(mm, f, [][]*Card{
		{NewCard(CardDesignSpade, 9, true)},
		{NewCard(CardDesignHeart, 8, true)},
	}, nil, nil)
	h := mm.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "tableau", h.ToZone)
	assert.Equal(t, 1, h.FromCol)
	assert.Equal(t, 0, h.ToIdx)
	require.NoError(t, mm.MoveTableauToTableau(h.FromCol, h.CardIndex, h.ToIdx))

	setMMBoard(mm, f, [][]*Card{
		{NewCard(CardDesignSpade, 5, true)},
		{NewCard(CardDesignHeart, 9, true)},
	}, []*Card{NewCard(CardDesignClover, 4, true)}, nil)
	h = mm.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "stock", h.FromZone)
	mm.checkStalemate()
	assert.False(t, mm.IsStalemate(), "a deal is still available")
}

// While something is held the only useful move is putting it back, so the hint
// must say so rather than suggesting a shuffle that cannot help.
func TestMissMilligan_HintInsistsOnReturningWaivedCards(t *testing.T) {
	mm := newTestMissMilligan()
	var f [MissMilliganFoundationCnt][]*Card
	setMMBoard(mm, f, [][]*Card{
		{NewCard(CardDesignSpade, 9, true)},
		{NewCard(CardDesignClover, 4, true)},
	}, nil, []*Card{NewCard(CardDesignHeart, 8, true)})

	h := mm.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "waived", h.FromZone)
	assert.Equal(t, "tableau", h.ToZone)
	assert.Equal(t, 0, h.ToIdx)
	require.NoError(t, mm.PlaceWaived(h.ToIdx))
}

func TestMissMilligan_HintIgnoresWholeColumnToEmptyColumnShuffles(t *testing.T) {
	mm := newTestMissMilligan()
	var f [MissMilliganFoundationCnt][]*Card
	// A lone King could legally move to an empty column, but that is no progress.
	setMMBoard(mm, f, [][]*Card{{NewCard(CardDesignSpade, CardValueMax, true)}}, nil, nil)

	assert.Nil(t, mm.GetHint(), "relocating the whole column achieves nothing")
}

// Waiving is a rescue, so a board with nothing else left is not a dead end
// while it is still on the table.
func TestMissMilligan_NotStalemateWhileWaivingIsStillAvailable(t *testing.T) {
	mm := newTestMissMilligan()
	var f [MissMilliganFoundationCnt][]*Card
	setMMBoard(mm, f, [][]*Card{{NewCard(CardDesignSpade, CardValueMax, true)}}, nil, nil)

	require.Nil(t, mm.GetHint())
	mm.checkStalemate()
	assert.False(t, mm.IsStalemate(), "the player can still waive")
	assert.True(t, mm.CanWaive())

	// Once the set is held and cannot be placed, there is genuinely nothing left.
	setMMBoard(mm, f, nil, nil, []*Card{NewCard(CardDesignHeart, 8, true)})
	mm.checkStalemate()
	assert.True(t, mm.IsStalemate(), "held card fits nowhere and no deal remains")
	assert.Equal(t, -1, mm.UndoToEscape(), "no history to rewind into")
}

func TestMissMilligan_AutoCompleteOnlyFeedsFoundations(t *testing.T) {
	mm := newTestMissMilligan()
	var f [MissMilliganFoundationCnt][]*Card
	f[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
	setMMBoard(mm, f, [][]*Card{
		{NewCard(CardDesignSpade, 3, true), NewCard(CardDesignSpade, 2, true)},
	}, nil, nil)

	require.NoError(t, mm.AutoComplete())
	assert.Len(t, mm.GetFoundation()[0], 3, "A,2,3 all went up")
	assert.Empty(t, mm.GetTableau()[0])
}

// AutoComplete must never make a tableau move, or it would shuffle the board
// instead of clearing it — it drives off foundationHint, not GetHint.
func TestMissMilligan_AutoCompleteIgnoresTableauMoves(t *testing.T) {
	mm := newTestMissMilligan()
	var f [MissMilliganFoundationCnt][]*Card
	setMMBoard(mm, f, [][]*Card{
		{NewCard(CardDesignSpade, 9, true)},
		{NewCard(CardDesignHeart, 8, true)},
	}, nil, nil)

	require.NotNil(t, mm.GetHint(), "a tableau move exists")
	assert.Error(t, mm.AutoComplete(), "but nothing reaches a foundation")
	assert.Equal(t, 0, mm.GetMoveCount())
	assert.Len(t, mm.GetTableau()[0], 1, "the board is untouched")
	assert.Len(t, mm.GetTableau()[1], 1)
}

func TestMissMilligan_AutoCompleteDrainsAWaivedCard(t *testing.T) {
	mm := newTestMissMilligan()
	var f [MissMilliganFoundationCnt][]*Card
	f[1] = []*Card{NewCard(CardDesignClover, 1, true)}
	setMMBoard(mm, f, nil, nil, []*Card{NewCard(CardDesignClover, 2, true)})

	require.NoError(t, mm.AutoComplete())
	assert.Empty(t, mm.GetWaived())
	assert.Len(t, mm.GetFoundation()[1], 2)
}

func TestMissMilligan_UndoRestoresEveryZone(t *testing.T) {
	mm := newTestMissMilligan()
	var f [MissMilliganFoundationCnt][]*Card
	f[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
	setMMBoard(mm, f, [][]*Card{
		{NewCard(CardDesignSpade, 2, true)},
		{NewCard(CardDesignHeart, 4, true)},
	}, nil, nil)

	assert.False(t, mm.CanUndo())
	assert.Error(t, mm.Undo(), "nothing to undo")
	assert.Error(t, mm.UndoN(0), "n must be positive")
	assert.Error(t, mm.UndoN(1), "no history yet")

	require.NoError(t, mm.MoveTableauToFoundation(0))
	require.NoError(t, mm.Waive(1, -1))
	assert.Len(t, mm.GetWaived(), 1)
	assert.Equal(t, 2, mm.GetMoveCount())

	assert.Error(t, mm.UndoN(5), "more than the history holds")
	require.NoError(t, mm.UndoN(2))
	assert.Equal(t, 0, mm.GetMoveCount())
	assert.Empty(t, mm.GetWaived(), "the waive was rolled back too")
	assert.Len(t, mm.GetTableau()[0], 1)
	assert.Len(t, mm.GetTableau()[1], 1)
	assert.Len(t, mm.GetFoundation()[0], 1)
}

// Waiving a card that fits nowhere is how a player strands themselves, and it
// is exactly what UndoToEscape has to count back out of.
func TestMissMilligan_UndoToEscapeCountsBackOutOfABadWaive(t *testing.T) {
	mm := newTestMissMilligan()
	var f [MissMilliganFoundationCnt][]*Card
	setMMBoard(mm, f, [][]*Card{{NewCard(CardDesignHeart, 8, true)}}, nil, nil)

	assert.False(t, mm.IsStalemate(), "waiving is still available")
	assert.Equal(t, 0, mm.UndoToEscape(), "not stalemated yet")

	require.NoError(t, mm.Waive(0, -1))
	// Everything is now off the table: ♥8 is not a King so no empty column takes
	// it, no foundation is open, the stock is gone and a second waive is barred.
	assert.True(t, mm.IsStalemate())
	assert.Equal(t, 1, mm.UndoToEscape())

	require.NoError(t, mm.UndoN(mm.UndoToEscape()))
	assert.False(t, mm.IsStalemate())
	assert.Empty(t, mm.GetWaived())
	assert.Len(t, mm.GetTableau()[0], 1)
}
func TestMissMilligan_ActionLogUsesZeroBasedIndices(t *testing.T) {
	mm := newTestMissMilligan()
	var f [MissMilliganFoundationCnt][]*Card
	f[2] = []*Card{NewCard(CardDesignHeart, 1, true)}
	setMMBoard(mm, f, [][]*Card{
		{NewCard(CardDesignSpade, 9, true)},
		{NewCard(CardDesignHeart, 8, true)},
		{NewCard(CardDesignHeart, 2, true)},
	}, nil, nil)

	require.NoError(t, mm.MoveTableauToTableau(1, 0, 0))
	require.NoError(t, mm.MoveTableauToFoundation(2))
	require.NoError(t, mm.Waive(0, -1))
	require.NoError(t, mm.PlaceWaived(0))

	log := mm.GetActionLog()
	require.Len(t, log, 4)
	assert.Equal(t, "タブロー列1→タブロー列0(1枚)", log[0].Detail)
	assert.Equal(t, "タブロー列2→基礎札2", log[1].Detail, "heart is foundation 2, 0-indexed")
	assert.Equal(t, "列0から1枚を保持", log[2].Detail)
	assert.Equal(t, "保持→タブロー列0(1枚)", log[3].Detail)
}

func TestMissMilligan_JSONRoundTrip(t *testing.T) {
	mm := newTestMissMilligan()
	require.NoError(t, mm.Deal())
	data, err := json.Marshal(mm)
	require.NoError(t, err)

	restored := NewMissMilligan(NewTrumpCardsWithDecks(2, 0))
	require.NoError(t, json.Unmarshal(data, restored))
	assert.Equal(t, mm.GetPhase(), restored.GetPhase())
	assert.Equal(t, mm.GetMoveCount(), restored.GetMoveCount())
	assert.Equal(t, mm.GetStockCount(), restored.GetStockCount())
	assert.Len(t, restored.GetTableau()[0], 2)
}

func TestMissMilligan_UnmarshalRejectsOutOfRangeState(t *testing.T) {
	mm := NewMissMilligan(NewTrumpCardsWithDecks(2, 0))
	assert.Error(t, json.Unmarshal([]byte(`nope`), mm))
	assert.Error(t, json.Unmarshal([]byte(`{"ps":99}`), mm))
	assert.Error(t, json.Unmarshal([]byte(`{"mc":-1}`), mm))

	card := `{"d":0,"v":1,"u":true}`
	big := func(n int) string {
		s := `[` + card
		for range n {
			s += `,` + card
		}
		return s + `]`
	}
	assert.Error(t, json.Unmarshal([]byte(`{"st":`+big(missMilliganTotalCards)+`}`), mm), "stock too large")
	assert.Error(t, json.Unmarshal([]byte(`{"wv":`+big(CardValueMax)+`}`), mm), "waived too large")
	assert.Error(t, json.Unmarshal([]byte(`{"fd":[`+big(CardValueMax)+`]}`), mm), "foundation too large")
}

func TestMissMilligan_NewDefaultUsesTwoDecks(t *testing.T) {
	mm := NewDefaultMissMilligan()
	mm.Reset()
	total := mm.GetStockCount() + len(mm.GetWaived())
	for _, f := range mm.GetFoundation() {
		total += len(f)
	}
	for _, col := range mm.GetTableau() {
		total += len(col)
	}
	assert.Equal(t, missMilliganTotalCards, total)
}

func TestMissMilligan_IsRunHandlesColourAndNil(t *testing.T) {
	assert.True(t, missMilliganIsRun(nil), "an empty run is vacuously a run")
	assert.True(t, missMilliganIsRun([]*MissMilliganTableauCard{
		{Card: NewCard(CardDesignSpade, 5, true)},
	}), "a single card is always a run")
	assert.False(t, missMilliganIsRun([]*MissMilliganTableauCard{
		{Card: NewCard(CardDesignSpade, 5, true)}, {Card: nil},
	}), "a nil card is never part of a run")
	assert.True(t, missMilliganIsRed(CardDesignHeart))
	assert.True(t, missMilliganIsRed(CardDesignDiamond))
	assert.False(t, missMilliganIsRed(CardDesignSpade))
	assert.False(t, missMilliganIsRed(CardDesignClover))
}
