//go:build test

package domain

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestNapoleonsSquare() *NapoleonsSquare {
	ns := NewNapoleonsSquare(NewTrumpCardsWithDecks(2, 0))
	ns.Reset()
	return ns
}

// setNSBoard installs an exact position. Tests must never lean on the shuffle:
// asserting what is or is not playable against a real deal is a flake waiting
// to happen (see #4467).
func setNSBoard(ns *NapoleonsSquare, foundation [NapoleonsSquareFoundationCnt][]*Card, cols [][]*Card, stock, waste []*Card) {
	ns.foundation = foundation
	for i := range NapoleonsSquareTableauCnt {
		ns.tableau[i] = nil
	}
	for i, c := range cols {
		pile := make([]*NapoleonsSquareTableauCard, 0, len(c))
		for _, card := range c {
			pile = append(pile, &NapoleonsSquareTableauCard{Card: card, FaceUp: true})
		}
		ns.tableau[i] = pile
	}
	ns.stock = stock
	ns.waste = waste
	ns.phase = NapoleonsSquarePhasePlaying
	ns.isStalemate = false
	ns.history = nil
	ns.moveCount = 0
	ns.actionLog = nil
}

func TestNapoleonsSquare_Reset_SeedsEightAcesAndDealsTwelveByFour(t *testing.T) {
	ns := newTestNapoleonsSquare()

	seeded := 0
	for i, pile := range ns.GetFoundation() {
		require.Len(t, pile, 1, "foundation %d holds exactly its Ace", i)
		assert.Equal(t, 1, pile[0].GetValue())
		assert.Equal(t, napoleonsSquareSuitOrder[i], pile[0].GetDesign(),
			"foundation %d is pinned to its suit", i)
		seeded += len(pile)
	}
	assert.Equal(t, NapoleonsSquareFoundationCnt, seeded)

	dealt := 0
	for col, pile := range ns.GetTableau() {
		assert.Len(t, pile, NapoleonsSquareColumnLen, "column %d", col)
		for _, tc := range pile {
			assert.True(t, tc.FaceUp, "every card is face-up by rule")
		}
		dealt += len(pile)
	}
	// The issue says "the remaining 96 cards" go to a 12x4 tableau, but 12x4 is
	// 48. The layout wins: 48 to the tableau, the other 48 to the stock.
	assert.Equal(t, 48, dealt)
	assert.Equal(t, 48, ns.GetStockCount())
	assert.Equal(t, napoleonsSquareTotalCards, seeded+dealt+ns.GetStockCount())
	assert.Empty(t, ns.GetWaste())
	assert.True(t, ns.AllFaceUp())
	assert.False(t, ns.GetGameEndFlag())
}

func TestNapoleonsSquare_DrawMovesStockToWaste(t *testing.T) {
	ns := newTestNapoleonsSquare()
	before := ns.GetStockCount()

	require.NoError(t, ns.Draw())
	assert.Equal(t, before-1, ns.GetStockCount())
	assert.Len(t, ns.GetWaste(), 1)

	// One pass only: no recycle once the stock runs out.
	var empty [NapoleonsSquareFoundationCnt][]*Card
	setNSBoard(ns, empty, nil, nil, []*Card{NewCard(CardDesignSpade, 5, true)})
	assert.Error(t, ns.Draw(), "stock is empty")
}

func TestNapoleonsSquare_TableauBuildsDownInSuit(t *testing.T) {
	ns := newTestNapoleonsSquare()
	var empty [NapoleonsSquareFoundationCnt][]*Card
	setNSBoard(ns, empty, [][]*Card{
		{NewCard(CardDesignSpade, 9, true)},
		{NewCard(CardDesignSpade, 8, true)},
		{NewCard(CardDesignHeart, 8, true)},
	}, nil, nil)

	assert.Error(t, ns.MoveTableauToTableau(2, 0, 0), "wrong suit")
	assert.Error(t, ns.MoveTableauToTableau(0, 0, 1), "builds down, not up")
	require.NoError(t, ns.MoveTableauToTableau(1, 0, 0), "♠8 onto ♠9")
	assert.Len(t, ns.GetTableau()[0], 2)
	assert.Empty(t, ns.GetTableau()[1])
}

func TestNapoleonsSquare_RejectsInvalidTableauArguments(t *testing.T) {
	ns := newTestNapoleonsSquare()
	var empty [NapoleonsSquareFoundationCnt][]*Card
	setNSBoard(ns, empty, [][]*Card{{NewCard(CardDesignSpade, 9, true)}}, nil, nil)

	assert.Error(t, ns.MoveTableauToTableau(-1, 0, 1), "from out of range")
	assert.Error(t, ns.MoveTableauToTableau(0, 0, NapoleonsSquareTableauCnt), "to out of range")
	assert.Error(t, ns.MoveTableauToTableau(0, 0, 0), "same column")
	assert.Error(t, ns.MoveTableauToTableau(0, 5, 1), "card index past the pile")
	assert.Error(t, ns.MoveTableauToTableau(0, -2, 1), "negative card index")
	assert.Error(t, ns.MoveWasteToTableau(-1), "column out of range")
	assert.Error(t, ns.MoveWasteToTableau(0), "waste is empty")
	assert.Error(t, ns.MoveTableauToFoundation(1), "column is empty")
	assert.Error(t, ns.MoveTableauToFoundation(-1), "column out of range")
	assert.Error(t, ns.MoveWasteToFoundation(), "waste is empty")
}

// The group move is the rule that separates this game from Forty Thieves.
func TestNapoleonsSquare_MovesASameSuitRunAsOneUnit(t *testing.T) {
	ns := newTestNapoleonsSquare()
	var empty [NapoleonsSquareFoundationCnt][]*Card
	setNSBoard(ns, empty, [][]*Card{
		{
			NewCard(CardDesignHeart, 2, true), // buried, not part of the run
			NewCard(CardDesignSpade, 7, true),
			NewCard(CardDesignSpade, 6, true),
			NewCard(CardDesignSpade, 5, true),
		},
		{NewCard(CardDesignSpade, 8, true)},
	}, nil, nil)

	// The run starts at index 1; taking it from index 0 includes ♥2 and breaks it.
	assert.Error(t, ns.MoveTableauToTableau(0, 0, 1), "not a same-suit run")

	require.NoError(t, ns.MoveTableauToTableau(0, 1, 1))
	assert.Len(t, ns.GetTableau()[1], 4, "♠8 plus the three-card run")
	require.Len(t, ns.GetTableau()[0], 1)
	assert.Equal(t, CardDesignHeart, ns.GetTableau()[0][0].Card.GetDesign())

	// Order is preserved, not reversed.
	got := ns.GetTableau()[1]
	assert.Equal(t, []int{8, 7, 6, 5}, []int{
		got[0].Card.GetValue(), got[1].Card.GetValue(),
		got[2].Card.GetValue(), got[3].Card.GetValue(),
	})
}

// A partially-descending pile is only movable from where the run actually starts.
func TestNapoleonsSquare_RejectsABrokenRun(t *testing.T) {
	ns := newTestNapoleonsSquare()
	var empty [NapoleonsSquareFoundationCnt][]*Card
	setNSBoard(ns, empty, [][]*Card{
		{
			NewCard(CardDesignSpade, 7, true),
			NewCard(CardDesignSpade, 5, true), // gap: 7 then 5
		},
		{NewCard(CardDesignSpade, 8, true)},
	}, nil, nil)

	assert.Error(t, ns.MoveTableauToTableau(0, 0, 1), "7,5 is not consecutive")
	assert.Error(t, ns.MoveTableauToTableau(0, 1, 1), "♠5 does not sit on ♠8")
}

// -1 means "the top card" so a caller that does not know the pile length can
// still move it; the Web controller relies on this when cardIndex is omitted.
func TestNapoleonsSquare_CardIndexMinusOneMeansTheTopCard(t *testing.T) {
	ns := newTestNapoleonsSquare()
	var empty [NapoleonsSquareFoundationCnt][]*Card
	setNSBoard(ns, empty, [][]*Card{
		{NewCard(CardDesignHeart, 2, true), NewCard(CardDesignSpade, 6, true)},
		{NewCard(CardDesignSpade, 7, true)},
	}, nil, nil)

	require.NoError(t, ns.MoveTableauToTableau(0, -1, 1))
	assert.Len(t, ns.GetTableau()[1], 2, "only ♠6 moved, not the ♥2 under it")
	assert.Len(t, ns.GetTableau()[0], 1)

	// An empty column has no top card, so -1 stays out of range rather than
	// silently doing nothing.
	assert.Error(t, ns.MoveTableauToTableau(5, -1, 1), "no top card in an empty column")
}

func TestNapoleonsSquare_EmptyColumnAcceptsAnythingIncludingARun(t *testing.T) {
	ns := newTestNapoleonsSquare()
	var empty [NapoleonsSquareFoundationCnt][]*Card
	setNSBoard(ns, empty, [][]*Card{
		{NewCard(CardDesignSpade, 7, true), NewCard(CardDesignSpade, 6, true)},
	}, nil, nil)

	require.NoError(t, ns.MoveTableauToTableau(0, 0, 5), "any card may fill an empty column")
	assert.Len(t, ns.GetTableau()[5], 2)
	assert.Empty(t, ns.GetTableau()[0])
}

func TestNapoleonsSquare_FoundationBuildsUpInSuit(t *testing.T) {
	ns := newTestNapoleonsSquare()
	var foundation [NapoleonsSquareFoundationCnt][]*Card
	for i, design := range napoleonsSquareSuitOrder {
		foundation[i] = []*Card{NewCard(design, 1, true)}
	}
	setNSBoard(ns, foundation, [][]*Card{
		{NewCard(CardDesignSpade, 2, true)},
		{NewCard(CardDesignHeart, 5, true)},
	}, nil, []*Card{NewCard(CardDesignClover, 2, true)})

	require.NoError(t, ns.MoveTableauToFoundation(0))
	assert.Len(t, ns.GetFoundation()[0], 2)

	assert.Error(t, ns.MoveTableauToFoundation(1), "♥5 does not follow ♥A")

	require.NoError(t, ns.MoveWasteToFoundation())
	assert.Len(t, ns.GetFoundation()[1], 2)
	assert.Empty(t, ns.GetWaste())
}

// With two decks each suit has two foundations; the second Ace opens the second.
func TestNapoleonsSquare_SecondDeckUsesTheSecondFoundationOfThatSuit(t *testing.T) {
	ns := newTestNapoleonsSquare()
	var foundation [NapoleonsSquareFoundationCnt][]*Card
	// Only the first spade foundation is seeded and already at 2.
	foundation[0] = []*Card{NewCard(CardDesignSpade, 1, true), NewCard(CardDesignSpade, 2, true)}
	setNSBoard(ns, foundation, [][]*Card{
		{NewCard(CardDesignSpade, 1, true)}, // the other deck's spade Ace
	}, nil, nil)

	require.NoError(t, ns.MoveTableauToFoundation(0))
	assert.Len(t, ns.GetFoundation()[0], 2, "the first pile is untouched")
	assert.Len(t, ns.GetFoundation()[4], 1, "the Ace opened the second spade pile")
}

func TestNapoleonsSquare_MoveWasteToTableau(t *testing.T) {
	ns := newTestNapoleonsSquare()
	var empty [NapoleonsSquareFoundationCnt][]*Card
	setNSBoard(ns, empty, [][]*Card{
		{NewCard(CardDesignSpade, 9, true)},
		{NewCard(CardDesignHeart, 3, true)},
	}, nil, []*Card{NewCard(CardDesignSpade, 8, true)})

	assert.Error(t, ns.MoveWasteToTableau(1), "♠8 does not sit on ♥3")
	require.NoError(t, ns.MoveWasteToTableau(0))
	assert.Len(t, ns.GetTableau()[0], 2)
	assert.Empty(t, ns.GetWaste())
}

func TestNapoleonsSquare_GameClearWhenAllEightFoundationsReachKing(t *testing.T) {
	ns := newTestNapoleonsSquare()
	var foundation [NapoleonsSquareFoundationCnt][]*Card
	for i, design := range napoleonsSquareSuitOrder {
		pile := make([]*Card, 0, CardValueMax)
		for v := 1; v <= CardValueMax; v++ {
			pile = append(pile, NewCard(design, v, true))
		}
		foundation[i] = pile
	}
	// One pile is one short, with its King sitting on the tableau.
	foundation[7] = foundation[7][:CardValueMax-1]
	setNSBoard(ns, foundation, [][]*Card{
		{NewCard(CardDesignDiamond, CardValueMax, true)},
	}, nil, nil)

	require.NoError(t, ns.MoveTableauToFoundation(0))
	assert.Equal(t, NapoleonsSquarePhaseGameClear, ns.GetPhase())
	assert.True(t, ns.GetGameEndFlag())
}

func TestNapoleonsSquare_GiveUpEndsTheGameOnce(t *testing.T) {
	ns := newTestNapoleonsSquare()
	ns.GiveUp()
	assert.Equal(t, NapoleonsSquarePhaseGameOver, ns.GetPhase())
	logLen := len(ns.GetActionLog())

	ns.GiveUp()
	assert.Len(t, ns.GetActionLog(), logLen, "a second give-up is a no-op")

	// Every move is refused once the game is over.
	assert.Error(t, ns.Draw())
	assert.Error(t, ns.MoveWasteToTableau(0))
	assert.Error(t, ns.MoveWasteToFoundation())
	assert.Error(t, ns.MoveTableauToTableau(0, 0, 1))
	assert.Error(t, ns.MoveTableauToFoundation(0))
	assert.Error(t, ns.AutoComplete())
	assert.Nil(t, ns.GetHint())
}

func TestNapoleonsSquare_HintPrefersTheFoundation(t *testing.T) {
	ns := newTestNapoleonsSquare()
	var foundation [NapoleonsSquareFoundationCnt][]*Card
	foundation[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
	setNSBoard(ns, foundation, [][]*Card{
		{NewCard(CardDesignHeart, 9, true)},
		{NewCard(CardDesignHeart, 8, true)},
		{NewCard(CardDesignSpade, 2, true)},
	}, nil, nil)

	h := ns.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "foundation", h.ToZone, "♠2 goes up even though ♥9/♥8 pair")
	assert.Equal(t, "tableau", h.FromZone)
	assert.Equal(t, 2, h.FromCol)
}

func TestNapoleonsSquare_HintFallsBackToTableauThenStock(t *testing.T) {
	ns := newTestNapoleonsSquare()
	var empty [NapoleonsSquareFoundationCnt][]*Card

	// No foundation move, but ♥8 sits on ♥9.
	setNSBoard(ns, empty, [][]*Card{
		{NewCard(CardDesignHeart, 9, true)},
		{NewCard(CardDesignHeart, 8, true)},
	}, nil, nil)
	h := ns.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "tableau", h.ToZone)
	assert.Equal(t, 1, h.FromCol)
	assert.Equal(t, 0, h.ToCol)
	// The hint must be a move the domain actually accepts.
	require.NoError(t, ns.MoveTableauToTableau(h.FromCol, h.CardIndex, h.ToCol))

	// Nothing on the board, but the stock can still be turned.
	setNSBoard(ns, empty, [][]*Card{
		{NewCard(CardDesignSpade, 5, true)},
		{NewCard(CardDesignHeart, 9, true)},
	}, []*Card{NewCard(CardDesignClover, 4, true)}, nil)
	h = ns.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "stock", h.FromZone)
	ns.checkStalemate()
	assert.False(t, ns.IsStalemate(), "a card left to turn is not a dead end")
}

// A whole column shuffled into an empty one makes no progress, so it must not be
// offered as a hint — otherwise the hint loops forever on a dead board.
func TestNapoleonsSquare_HintIgnoresWholeColumnToEmptyColumnShuffles(t *testing.T) {
	ns := newTestNapoleonsSquare()
	var empty [NapoleonsSquareFoundationCnt][]*Card
	setNSBoard(ns, empty, [][]*Card{
		{NewCard(CardDesignSpade, 5, true)},
		{NewCard(CardDesignHeart, 9, true)},
	}, nil, nil)

	assert.Nil(t, ns.GetHint(), "moving either column to an empty one is not progress")
	ns.checkStalemate()
	assert.True(t, ns.IsStalemate())
	assert.Equal(t, -1, ns.UndoToEscape(), "no history to rewind into")
}

func TestNapoleonsSquare_AutoCompleteOnlyFeedsFoundations(t *testing.T) {
	ns := newTestNapoleonsSquare()
	var foundation [NapoleonsSquareFoundationCnt][]*Card
	foundation[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
	setNSBoard(ns, foundation, [][]*Card{
		{NewCard(CardDesignSpade, 3, true), NewCard(CardDesignSpade, 2, true)},
	}, nil, nil)

	require.NoError(t, ns.AutoComplete())
	assert.Len(t, ns.GetFoundation()[0], 3, "A,2,3 all went up")
	assert.Empty(t, ns.GetTableau()[0])
}

// AutoComplete must never make a tableau move, or it would shuffle the board
// instead of clearing it — it drives off foundationHint, not GetHint.
func TestNapoleonsSquare_AutoCompleteIgnoresTableauMoves(t *testing.T) {
	ns := newTestNapoleonsSquare()
	var empty [NapoleonsSquareFoundationCnt][]*Card
	setNSBoard(ns, empty, [][]*Card{
		{NewCard(CardDesignHeart, 9, true)},
		{NewCard(CardDesignHeart, 8, true)},
	}, nil, nil)

	require.NotNil(t, ns.GetHint(), "a tableau move exists")
	assert.Error(t, ns.AutoComplete(), "but nothing can reach a foundation")
	assert.Equal(t, 0, ns.GetMoveCount())
	assert.Len(t, ns.GetTableau()[0], 1, "the board is untouched")
	assert.Len(t, ns.GetTableau()[1], 1)
}

func TestNapoleonsSquare_AutoCompleteDrainsTheWaste(t *testing.T) {
	ns := newTestNapoleonsSquare()
	var foundation [NapoleonsSquareFoundationCnt][]*Card
	foundation[1] = []*Card{NewCard(CardDesignClover, 1, true)}
	setNSBoard(ns, foundation, nil, nil, []*Card{NewCard(CardDesignClover, 2, true)})

	require.NoError(t, ns.AutoComplete())
	assert.Empty(t, ns.GetWaste())
	assert.Len(t, ns.GetFoundation()[1], 2)
}

func TestNapoleonsSquare_UndoRestoresEveryZone(t *testing.T) {
	ns := newTestNapoleonsSquare()
	var foundation [NapoleonsSquareFoundationCnt][]*Card
	foundation[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
	setNSBoard(ns, foundation, [][]*Card{
		{NewCard(CardDesignSpade, 2, true)},
	}, []*Card{NewCard(CardDesignHeart, 4, true)}, nil)

	assert.False(t, ns.CanUndo())
	assert.Error(t, ns.Undo(), "nothing to undo")
	assert.Error(t, ns.UndoN(0), "n must be positive")
	assert.Error(t, ns.UndoN(1), "no history yet")

	require.NoError(t, ns.Draw())
	require.NoError(t, ns.MoveTableauToFoundation(0))
	assert.True(t, ns.CanUndo())
	assert.Equal(t, 2, ns.GetMoveCount())

	assert.Error(t, ns.UndoN(5), "more than the history holds")
	require.NoError(t, ns.UndoN(2))
	assert.Equal(t, 0, ns.GetMoveCount())
	assert.Len(t, ns.GetTableau()[0], 1)
	assert.Len(t, ns.GetFoundation()[0], 1)
	assert.Equal(t, 1, ns.GetStockCount())
	assert.Empty(t, ns.GetWaste())
}

func TestNapoleonsSquare_UndoToEscapeCountsBackToAPlayablePosition(t *testing.T) {
	ns := newTestNapoleonsSquare()
	var empty [NapoleonsSquareFoundationCnt][]*Card

	// Reaching a dead end takes some care in this game: an empty column accepts
	// anything, so any move that empties a column immediately creates a target.
	// All 12 columns are therefore filled, and the only move (♠5 onto ♠6) leaves
	// ♥3 behind rather than emptying its column. No Ace is in play, so the
	// foundations stay shut. The blockers are spaced two apart within each suit
	// so none of them is adjacent to another.
	cols := [][]*Card{
		{NewCard(CardDesignSpade, 6, true)},
		{NewCard(CardDesignHeart, 3, true), NewCard(CardDesignSpade, 5, true)},
		{NewCard(CardDesignHeart, 6, true)},
		{NewCard(CardDesignHeart, 8, true)},
		{NewCard(CardDesignHeart, 10, true)},
		{NewCard(CardDesignHeart, 12, true)},
		{NewCard(CardDesignDiamond, 2, true)},
		{NewCard(CardDesignDiamond, 4, true)},
		{NewCard(CardDesignDiamond, 6, true)},
		{NewCard(CardDesignDiamond, 8, true)},
		{NewCard(CardDesignClover, 2, true)},
		{NewCard(CardDesignClover, 4, true)},
	}
	require.Len(t, cols, NapoleonsSquareTableauCnt, "no column may be left empty")
	setNSBoard(ns, empty, cols, nil, nil)

	assert.False(t, ns.IsStalemate())
	assert.Equal(t, 0, ns.UndoToEscape(), "not stalemated yet")

	require.NoError(t, ns.MoveTableauToTableau(1, 1, 0), "♠5 onto ♠6 is the only move")
	assert.True(t, ns.IsStalemate(), "and it strands the board")
	assert.Equal(t, 1, ns.UndoToEscape())

	require.NoError(t, ns.UndoN(ns.UndoToEscape()))
	assert.False(t, ns.IsStalemate())
	assert.Len(t, ns.GetTableau()[1], 2, "♠5 is back on top of ♥3")
}

func TestNapoleonsSquare_ActionLogUsesZeroBasedIndices(t *testing.T) {
	ns := newTestNapoleonsSquare()
	var foundation [NapoleonsSquareFoundationCnt][]*Card
	foundation[2] = []*Card{NewCard(CardDesignHeart, 1, true)}
	setNSBoard(ns, foundation, [][]*Card{
		{NewCard(CardDesignSpade, 7, true), NewCard(CardDesignSpade, 6, true)},
		{NewCard(CardDesignSpade, 8, true)},
		{NewCard(CardDesignHeart, 2, true)},
	}, []*Card{NewCard(CardDesignClover, 9, true)}, nil)

	require.NoError(t, ns.MoveTableauToTableau(0, 0, 1))
	require.NoError(t, ns.MoveTableauToFoundation(2))
	require.NoError(t, ns.Draw())
	require.NoError(t, ns.MoveWasteToTableau(0))

	log := ns.GetActionLog()
	require.Len(t, log, 4)
	assert.Equal(t, "タブロー列0→タブロー列1(2枚)", log[0].Detail)
	assert.Equal(t, "タブロー列2→基礎札2", log[1].Detail, "heart is foundation 2, 0-indexed")
	assert.Equal(t, "山札→ウェイスト", log[2].Detail)
	assert.Equal(t, "ウェイスト→タブロー列0", log[3].Detail)
}

func TestNapoleonsSquare_JSONRoundTrip(t *testing.T) {
	ns := newTestNapoleonsSquare()
	require.NoError(t, ns.Draw())
	data, err := json.Marshal(ns)
	require.NoError(t, err)

	restored := NewNapoleonsSquare(NewTrumpCardsWithDecks(2, 0))
	require.NoError(t, json.Unmarshal(data, restored))
	assert.Equal(t, ns.GetPhase(), restored.GetPhase())
	assert.Equal(t, ns.GetMoveCount(), restored.GetMoveCount())
	assert.Equal(t, ns.GetStockCount(), restored.GetStockCount())
	assert.Len(t, restored.GetWaste(), 1)
	assert.Len(t, restored.GetFoundation()[0], 1)
	assert.Len(t, restored.GetTableau()[0], NapoleonsSquareColumnLen)
}

func TestNapoleonsSquare_UnmarshalRejectsOutOfRangeState(t *testing.T) {
	ns := NewNapoleonsSquare(NewTrumpCardsWithDecks(2, 0))
	assert.Error(t, json.Unmarshal([]byte(`nope`), ns))
	assert.Error(t, json.Unmarshal([]byte(`{"ps":99}`), ns))
	assert.Error(t, json.Unmarshal([]byte(`{"mc":-1}`), ns))

	big := make([]string, 0, napoleonsSquareTotalCards+1)
	for range napoleonsSquareTotalCards + 1 {
		big = append(big, `{"d":0,"v":1,"u":true}`)
	}
	oversize := `[` + big[0]
	for _, b := range big[1:] {
		oversize += `,` + b
	}
	oversize += `]`
	assert.Error(t, json.Unmarshal([]byte(`{"st":`+oversize+`}`), ns), "stock too large")
	assert.Error(t, json.Unmarshal([]byte(`{"ws":`+oversize+`}`), ns), "waste too large")
	assert.Error(t, json.Unmarshal([]byte(`{"fd":[`+oversize+`]}`), ns), "foundation too large")
}

func TestNapoleonsSquare_NewDefaultUsesTwoDecks(t *testing.T) {
	ns := NewDefaultNapoleonsSquare()
	ns.Reset()
	total := ns.GetStockCount() + len(ns.GetWaste())
	for _, f := range ns.GetFoundation() {
		total += len(f)
	}
	for _, col := range ns.GetTableau() {
		total += len(col)
	}
	assert.Equal(t, napoleonsSquareTotalCards, total)
}

func TestNapoleonsSquare_IsRunHandlesNilCards(t *testing.T) {
	assert.True(t, napoleonsSquareIsRun(nil), "an empty run is vacuously a run")
	assert.True(t, napoleonsSquareIsRun([]*NapoleonsSquareTableauCard{
		{Card: NewCard(CardDesignSpade, 5, true)},
	}), "a single card is always a run")
	assert.False(t, napoleonsSquareIsRun([]*NapoleonsSquareTableauCard{
		{Card: NewCard(CardDesignSpade, 5, true)}, {Card: nil},
	}), "a nil card is never part of a run")
}

func TestNapoleonsSquare_ErrorsCarryAnI18nCode(t *testing.T) {
	codeOf := func(t *testing.T, err error) string {
		t.Helper()
		if err == nil {
			return ""
		}
		code, _ := ErrorMessageCode(err)
		return code
	}

	t.Run("every refusal names a key instead of an English sentence", func(t *testing.T) {
		cases := []struct {
			name string
			run  func(ns *NapoleonsSquare) error
		}{
			{"draw empty stock", func(ns *NapoleonsSquare) error {
				for ns.GetStockCount() > 0 {
					_ = ns.Draw()
				}
				return ns.Draw()
			}},
			{"move from a column that does not exist", func(ns *NapoleonsSquare) error {
				return ns.MoveTableauToTableau(-1, 0, 0)
			}},
			{"move to a column that does not exist", func(ns *NapoleonsSquare) error {
				return ns.MoveTableauToTableau(0, NapoleonsSquareTableauCnt, 0)
			}},
			{"move a column onto itself", func(ns *NapoleonsSquare) error {
				return ns.MoveTableauToTableau(0, 0, 0)
			}},
			{"send a card up from a column that does not exist", func(ns *NapoleonsSquare) error {
				return ns.MoveTableauToFoundation(-1)
			}},
			{"undo with nothing to undo", func(ns *NapoleonsSquare) error {
				return ns.Undo()
			}},
			// 公開の入口を1つでも外すと、そこだけ素の英語が残っても誰も気付かない。
			// 実際、この2件を足す前は MoveWasteTo* の拒否を一度も踏んでいなかった。
			{"put the waste on a column that does not exist", func(ns *NapoleonsSquare) error {
				return ns.MoveWasteToTableau(-1)
			}},
			{"send the waste up with an empty waste", func(ns *NapoleonsSquare) error {
				return ns.MoveWasteToFoundation()
			}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				ns := newTestNapoleonsSquare()
				err := tc.run(ns)
				require.Error(t, err, "この操作は拒否されるはずで、拒否されないと何も測れない")
				code := codeOf(t, err)
				assert.NotEmpty(t, code, "コードが無いと CUI は英語をそのまま出す")
				assert.Truef(t, strings.HasPrefix(code, "napoleonssquare."),
					"キーは napoleonssquare 名前空間に置く (got %q)", code)
			})
		}
	})
}
