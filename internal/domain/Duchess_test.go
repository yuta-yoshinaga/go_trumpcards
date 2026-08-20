//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestDuchess() *Duchess {
	d := NewDuchess(NewTrumpCards(0))
	d.Reset()
	return d
}

// setDuchessBoard installs an exact position. Tests must never lean on the
// shuffle: asserting what is or is not playable against a real deal is a flake
// waiting to happen (see #4467).
func setDuchessBoard(d *Duchess, baseRank int, foundation [DuchessFoundationCnt][]*Card, fans [][]*Card, cols [][]*Card, stock, waste []*Card) {
	d.baseRank = baseRank
	d.foundation = foundation
	for i := range DuchessReserveCnt {
		d.reserve[i] = nil
	}
	for i, f := range fans {
		d.reserve[i] = append([]*Card(nil), f...)
	}
	for i := range DuchessTableauCnt {
		d.tableau[i] = nil
	}
	for i, c := range cols {
		pile := make([]*DuchessTableauCard, 0, len(c))
		for _, card := range c {
			pile = append(pile, &DuchessTableauCard{Card: card, FaceUp: true})
		}
		d.tableau[i] = pile
	}
	d.stock = stock
	d.waste = waste
	d.phase = DuchessPhasePlaying
	d.isStalemate = false
	d.history = nil
	d.moveCount = 0
	d.actionLog = nil
}

func TestDuchess_Reset_DealsFourFansOfThree(t *testing.T) {
	d := newTestDuchess()

	reserved := 0
	for i, fan := range d.GetReserve() {
		// #4408 says four cards per fan; Duchess/Glenwood deals three.
		require.Len(t, fan, DuchessReserveFanSize, "fan %d", i)
		reserved += len(fan)
	}
	assert.Equal(t, 12, reserved)

	dealt := 0
	for col, pile := range d.GetTableau() {
		require.Len(t, pile, 1, "column %d starts with one card", col)
		assert.True(t, pile[0].FaceUp, "every card is face-up by rule")
		dealt += len(pile)
	}
	assert.Equal(t, DuchessTableauCnt, dealt)
	assert.Equal(t, CardCnt-12-DuchessTableauCnt, d.GetStockCount(), "36 left in the stock")
	assert.Equal(t, CardCnt, reserved+dealt+d.GetStockCount())
	assert.Empty(t, d.GetWaste())
	for _, f := range d.GetFoundation() {
		assert.Empty(t, f, "no foundation opens until the base rank is chosen")
	}
	assert.True(t, d.IsAwaitingBaseRank())
	assert.Equal(t, 0, d.GetBaseRank())
	assert.True(t, d.AllFaceUp())
}

// Nothing may happen until the player picks which reserve top sets the rank.
func TestDuchess_EverythingIsBlockedUntilTheBaseRankIsChosen(t *testing.T) {
	d := newTestDuchess()
	require.True(t, d.IsAwaitingBaseRank())

	assert.Error(t, d.Draw())
	assert.Error(t, d.MoveReserveToFoundation(0))
	assert.Error(t, d.MoveReserveToTableau(0, 0))
	assert.Error(t, d.MoveWasteToFoundation())
	assert.Error(t, d.MoveWasteToTableau(0))
	assert.Error(t, d.MoveTableauToFoundation(0))
	assert.Error(t, d.MoveTableauToTableau(0, -1, 1))
	assert.Error(t, d.AutoComplete())

	// The only hint is to choose.
	h := d.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "reserve", h.FromZone)
	assert.Equal(t, "foundation", h.ToZone)
	assert.False(t, d.IsStalemate(), "choosing is a move, so this is not a dead end")
}

func TestDuchess_ChooseBaseRankOpensThatSuitsFoundation(t *testing.T) {
	d := newTestDuchess()
	var f [DuchessFoundationCnt][]*Card
	setDuchessBoard(d, 0, f, [][]*Card{
		{NewCard(CardDesignHeart, 2, true), NewCard(CardDesignSpade, 9, true)},
	}, nil, nil, nil)

	require.NoError(t, d.ChooseBaseRank(0))
	assert.Equal(t, 9, d.GetBaseRank(), "the chosen card's rank sets it for all four")
	assert.Len(t, d.GetFoundation()[0], 1, "spade is foundation 0")
	assert.False(t, d.IsAwaitingBaseRank())
	assert.Len(t, d.GetReserve()[0], 1, "the card left the fan")

	assert.Error(t, d.ChooseBaseRank(0), "the base rank is already chosen")
}

func TestDuchess_ChooseBaseRankRejectsBadFans(t *testing.T) {
	d := newTestDuchess()
	var f [DuchessFoundationCnt][]*Card
	setDuchessBoard(d, 0, f, nil, nil, nil, nil)

	assert.Error(t, d.ChooseBaseRank(-1), "index out of range")
	assert.Error(t, d.ChooseBaseRank(DuchessReserveCnt), "index out of range")
	assert.Error(t, d.ChooseBaseRank(0), "fan is empty")
}

// Foundations run from the base rank and wrap King round to Ace.
func TestDuchess_FoundationWrapsAroundFromTheBaseRank(t *testing.T) {
	d := newTestDuchess()
	var f [DuchessFoundationCnt][]*Card
	f[0] = []*Card{NewCard(CardDesignSpade, CardValueMax, true)}
	setDuchessBoard(d, CardValueMax, f, nil, [][]*Card{
		{NewCard(CardDesignSpade, 1, true)},
		{NewCard(CardDesignSpade, 3, true)},
	}, nil, nil)

	assert.Error(t, d.MoveTableauToFoundation(1), "♠3 does not follow ♠K")
	require.NoError(t, d.MoveTableauToFoundation(0), "♠A follows ♠K")
	assert.Len(t, d.GetFoundation()[0], 2)
	assert.Equal(t, 1, duchessNextRank(CardValueMax))
	assert.Equal(t, CardValueMax, duchessPrevRank(1))
}

func TestDuchess_FoundationIsPinnedToItsSuitAndBaseRank(t *testing.T) {
	d := newTestDuchess()
	var f [DuchessFoundationCnt][]*Card
	setDuchessBoard(d, 5, f, nil, [][]*Card{
		{NewCard(CardDesignHeart, 5, true)},
		{NewCard(CardDesignHeart, 6, true)},
	}, nil, nil)

	assert.Error(t, d.MoveTableauToFoundation(1), "♥6 is not the base rank")
	require.NoError(t, d.MoveTableauToFoundation(0), "♥5 opens the heart pile")
	assert.Len(t, d.GetFoundation()[2], 1, "heart is foundation 2")
	for i, pile := range d.GetFoundation() {
		if i != 2 {
			assert.Empty(t, pile, "no other suit took it")
		}
	}
}

func TestDuchess_TableauBuildsDownInAlternatingColourAndWraps(t *testing.T) {
	d := newTestDuchess()
	var f [DuchessFoundationCnt][]*Card
	setDuchessBoard(d, 5, f, nil, [][]*Card{
		{NewCard(CardDesignSpade, 1, true)},
		{NewCard(CardDesignHeart, CardValueMax, true)},
		{NewCard(CardDesignClover, CardValueMax, true)},
	}, nil, nil)

	assert.Error(t, d.MoveTableauToTableau(2, -1, 0), "♣K on ♠A is the same colour")
	require.NoError(t, d.MoveTableauToTableau(1, -1, 0), "♥K wraps onto ♠A")
	assert.Len(t, d.GetTableau()[0], 2)
}

// The Canfield-family rule the issue omits: while the reserve holds cards, an
// empty column can only be filled from it.
func TestDuchess_EmptyColumnIsReservedForTheReserve(t *testing.T) {
	d := newTestDuchess()
	var f [DuchessFoundationCnt][]*Card
	setDuchessBoard(d, 5, f, [][]*Card{
		{NewCard(CardDesignHeart, 4, true)},
	}, [][]*Card{
		{NewCard(CardDesignSpade, 9, true)},
	}, nil, []*Card{NewCard(CardDesignClover, 7, true)})

	// Column 1 is empty and the reserve is not, so only the reserve may fill it.
	assert.Error(t, d.MoveWasteToTableau(1), "waste cannot fill it yet")
	assert.Error(t, d.MoveTableauToTableau(0, -1, 1), "nor can another column")
	require.NoError(t, d.MoveReserveToTableau(0, 1))
	assert.Len(t, d.GetTableau()[1], 1)

	// With the reserve now empty, anything may fill an empty column.
	require.NoError(t, d.MoveWasteToTableau(2))
	assert.Len(t, d.GetTableau()[2], 1)
}

func TestDuchess_MovesAnAlternatingRunAsOneUnit(t *testing.T) {
	d := newTestDuchess()
	var f [DuchessFoundationCnt][]*Card
	setDuchessBoard(d, 5, f, nil, [][]*Card{
		{
			NewCard(CardDesignHeart, 2, true), // buried, breaks the run
			NewCard(CardDesignSpade, 7, true),
			NewCard(CardDesignHeart, 6, true),
		},
		{NewCard(CardDesignHeart, 8, true)},
	}, nil, nil)

	assert.Error(t, d.MoveTableauToTableau(0, 0, 1), "♥2 is not part of the run")
	require.NoError(t, d.MoveTableauToTableau(0, 1, 1))
	assert.Len(t, d.GetTableau()[1], 3, "♥8 plus the two-card run")
	assert.Len(t, d.GetTableau()[0], 1)
}

func TestDuchess_RejectsInvalidArguments(t *testing.T) {
	d := newTestDuchess()
	var f [DuchessFoundationCnt][]*Card
	setDuchessBoard(d, 5, f, [][]*Card{{NewCard(CardDesignHeart, 4, true)}},
		[][]*Card{{NewCard(CardDesignSpade, 9, true)}}, nil, nil)

	assert.Error(t, d.MoveReserveToFoundation(-1), "reserve index out of range")
	assert.Error(t, d.MoveReserveToFoundation(1), "fan is empty")
	assert.Error(t, d.MoveReserveToTableau(0, -1), "column out of range")
	assert.Error(t, d.MoveWasteToFoundation(), "waste is empty")
	assert.Error(t, d.MoveWasteToTableau(0), "waste is empty")
	assert.Error(t, d.MoveTableauToFoundation(-1), "column out of range")
	assert.Error(t, d.MoveTableauToFoundation(1), "column is empty")
	assert.Error(t, d.MoveTableauToTableau(-1, 0, 1), "from out of range")
	assert.Error(t, d.MoveTableauToTableau(0, 0, DuchessTableauCnt), "to out of range")
	assert.Error(t, d.MoveTableauToTableau(0, 0, 0), "same column")
	assert.Error(t, d.MoveTableauToTableau(0, 9, 1), "card index past the pile")
	assert.Error(t, d.MoveTableauToTableau(0, -2, 1), "negative card index")
}

func TestDuchess_DrawMovesStockToWaste(t *testing.T) {
	d := newTestDuchess()
	var f [DuchessFoundationCnt][]*Card
	setDuchessBoard(d, 5, f, nil, nil, []*Card{NewCard(CardDesignSpade, 9, true)}, nil)

	require.NoError(t, d.Draw())
	assert.Equal(t, 0, d.GetStockCount())
	assert.Len(t, d.GetWaste(), 1)

	// One pass only: there is no redeal.
	assert.Error(t, d.Draw(), "stock is empty")
}

func TestDuchess_ReserveAndWasteFeedTheFoundation(t *testing.T) {
	d := newTestDuchess()
	var f [DuchessFoundationCnt][]*Card
	f[0] = []*Card{NewCard(CardDesignSpade, 5, true)}
	setDuchessBoard(d, 5, f, [][]*Card{
		{NewCard(CardDesignSpade, 6, true)},
	}, nil, nil, []*Card{NewCard(CardDesignSpade, 7, true)})

	require.NoError(t, d.MoveReserveToFoundation(0))
	assert.Len(t, d.GetFoundation()[0], 2)
	require.NoError(t, d.MoveWasteToFoundation())
	assert.Len(t, d.GetFoundation()[0], 3)
	assert.Empty(t, d.GetWaste())
}

func TestDuchess_GameClearWhenEveryFoundationHasThirteen(t *testing.T) {
	d := newTestDuchess()
	var f [DuchessFoundationCnt][]*Card
	for i, design := range duchessSuitOrder {
		pile := make([]*Card, 0, CardValueMax)
		v := 5
		for range CardValueMax {
			pile = append(pile, NewCard(design, v, true))
			v = duchessNextRank(v)
		}
		f[i] = pile
	}
	f[3] = f[3][:CardValueMax-1]
	last := f[3][len(f[3])-1].GetValue()
	setDuchessBoard(d, 5, f, nil, [][]*Card{
		{NewCard(CardDesignDiamond, duchessNextRank(last), true)},
	}, nil, nil)

	require.NoError(t, d.MoveTableauToFoundation(0))
	assert.Equal(t, DuchessPhaseGameClear, d.GetPhase())
	assert.True(t, d.GetGameEndFlag())
}

// A full pile stops accepting cards; without that the wraparound would lap.
func TestDuchess_FullFoundationRejectsMoreCards(t *testing.T) {
	d := newTestDuchess()
	var f [DuchessFoundationCnt][]*Card
	pile := make([]*Card, 0, CardValueMax)
	v := 5
	for range CardValueMax {
		pile = append(pile, NewCard(CardDesignSpade, v, true))
		v = duchessNextRank(v)
	}
	f[0] = pile
	setDuchessBoard(d, 5, f, nil, [][]*Card{{NewCard(CardDesignSpade, 5, true)}}, nil, nil)

	assert.False(t, d.canPlaceOnFoundation(NewCard(CardDesignSpade, 5, true), 0))
	assert.Error(t, d.MoveTableauToFoundation(0), "a completed pile would lap the suit")
}

func TestDuchess_GiveUpEndsTheGameOnce(t *testing.T) {
	d := newTestDuchess()
	d.GiveUp()
	assert.Equal(t, DuchessPhaseGameOver, d.GetPhase())
	logLen := len(d.GetActionLog())

	d.GiveUp()
	assert.Len(t, d.GetActionLog(), logLen, "a second give-up is a no-op")

	assert.Error(t, d.ChooseBaseRank(0))
	assert.Error(t, d.Draw())
	assert.Error(t, d.MoveReserveToFoundation(0))
	assert.Error(t, d.MoveTableauToTableau(0, -1, 1))
	assert.Error(t, d.AutoComplete())
	assert.Nil(t, d.GetHint())
	assert.False(t, d.IsAwaitingBaseRank(), "the game is over, not waiting")
}

func TestDuchess_HintPrefersTheFoundation(t *testing.T) {
	d := newTestDuchess()
	var f [DuchessFoundationCnt][]*Card
	f[0] = []*Card{NewCard(CardDesignSpade, 5, true)}
	setDuchessBoard(d, 5, f, nil, [][]*Card{
		{NewCard(CardDesignSpade, 9, true)},
		{NewCard(CardDesignHeart, 8, true)}, // a legal tableau move
		{NewCard(CardDesignSpade, 6, true)}, // but ♠6 goes up
	}, nil, nil)

	h := d.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "foundation", h.ToZone)
	assert.Equal(t, "tableau", h.FromZone)
	assert.Equal(t, 2, h.FromIdx)
}

// Emptying the reserve is the point of the game, so that move outranks a shuffle.
func TestDuchess_HintPrefersTheReserveOverATableauShuffle(t *testing.T) {
	d := newTestDuchess()
	var f [DuchessFoundationCnt][]*Card
	setDuchessBoard(d, 5, f, [][]*Card{
		{NewCard(CardDesignHeart, 8, true)},
	}, [][]*Card{
		{NewCard(CardDesignSpade, 9, true)},
		{NewCard(CardDesignClover, 9, true)},
		{NewCard(CardDesignDiamond, 8, true)}, // could also go onto ♠9 or ♣9
	}, nil, nil)

	h := d.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "reserve", h.FromZone)
	assert.Equal(t, "tableau", h.ToZone)
	require.NoError(t, d.MoveReserveToTableau(h.FromIdx, h.ToIdx))
}

func TestDuchess_HintFallsBackToTheStock(t *testing.T) {
	d := newTestDuchess()
	var f [DuchessFoundationCnt][]*Card
	setDuchessBoard(d, 5, f, nil, [][]*Card{
		{NewCard(CardDesignSpade, 9, true)},
		{NewCard(CardDesignClover, 3, true)},
		{NewCard(CardDesignHeart, 12, true)},
		{NewCard(CardDesignDiamond, 6, true)},
	}, []*Card{NewCard(CardDesignClover, 4, true)}, nil)

	h := d.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "stock", h.FromZone)
	d.checkStalemate()
	assert.False(t, d.IsStalemate(), "a card left to turn is not a dead end")
}

func TestDuchess_StalemateWhenNothingRemains(t *testing.T) {
	d := newTestDuchess()
	var f [DuchessFoundationCnt][]*Card
	// Four occupied columns whose tops pair with nothing, no reserve, no stock.
	setDuchessBoard(d, 5, f, nil, [][]*Card{
		{NewCard(CardDesignSpade, 9, true)},
		{NewCard(CardDesignClover, 3, true)},
		{NewCard(CardDesignHeart, 12, true)},
		{NewCard(CardDesignDiamond, 6, true)},
	}, nil, nil)

	d.checkStalemate()
	assert.True(t, d.IsStalemate())
	assert.Equal(t, -1, d.UndoToEscape(), "no history to rewind into")
}

// Turning the last stock card can strand the board, and that is what
// UndoToEscape has to count back out of.
func TestDuchess_UndoToEscapeCountsBackToAPlayablePosition(t *testing.T) {
	d := newTestDuchess()
	var f [DuchessFoundationCnt][]*Card
	// Four blocked columns, no reserve, and one card left to turn. The turn is
	// the only move, and ♣4 fits nothing once it is face up.
	setDuchessBoard(d, 5, f, nil, [][]*Card{
		{NewCard(CardDesignSpade, 9, true)},
		{NewCard(CardDesignClover, 3, true)},
		{NewCard(CardDesignHeart, 12, true)},
		{NewCard(CardDesignDiamond, 6, true)},
	}, []*Card{NewCard(CardDesignClover, 4, true)}, nil)

	assert.False(t, d.IsStalemate(), "the stock still has a card")
	assert.Equal(t, 0, d.UndoToEscape(), "not stalemated yet")

	require.NoError(t, d.Draw())
	assert.True(t, d.IsStalemate(), "and turning it strands the board")
	assert.Equal(t, 1, d.UndoToEscape())

	require.NoError(t, d.UndoN(d.UndoToEscape()))
	assert.False(t, d.IsStalemate())
	assert.Equal(t, 1, d.GetStockCount())
}
func TestDuchess_AutoCompleteOnlyFeedsFoundations(t *testing.T) {
	d := newTestDuchess()
	var f [DuchessFoundationCnt][]*Card
	f[0] = []*Card{NewCard(CardDesignSpade, 5, true)}
	setDuchessBoard(d, 5, f, [][]*Card{
		{NewCard(CardDesignSpade, 6, true)},
	}, [][]*Card{
		{NewCard(CardDesignSpade, 8, true), NewCard(CardDesignSpade, 7, true)},
	}, nil, nil)

	require.NoError(t, d.AutoComplete())
	assert.Len(t, d.GetFoundation()[0], 4, "6,7,8 all went up")
	assert.Empty(t, d.GetReserve()[0])
}

// AutoComplete must never make a tableau move, or it would shuffle the board
// instead of clearing it — it drives off foundationHint, not GetHint.
func TestDuchess_AutoCompleteIgnoresTableauMoves(t *testing.T) {
	d := newTestDuchess()
	var f [DuchessFoundationCnt][]*Card
	setDuchessBoard(d, 5, f, nil, [][]*Card{
		{NewCard(CardDesignSpade, 9, true)},
		{NewCard(CardDesignHeart, 8, true)},
	}, nil, nil)

	require.NotNil(t, d.GetHint(), "a tableau move exists")
	assert.Error(t, d.AutoComplete(), "but nothing reaches a foundation")
	assert.Equal(t, 0, d.GetMoveCount())
	assert.Len(t, d.GetTableau()[0], 1, "the board is untouched")
}

func TestDuchess_UndoRestoresEveryZoneIncludingTheBaseRank(t *testing.T) {
	d := newTestDuchess()
	var f [DuchessFoundationCnt][]*Card
	setDuchessBoard(d, 0, f, [][]*Card{
		{NewCard(CardDesignSpade, 5, true)},
	}, nil, []*Card{NewCard(CardDesignHeart, 4, true)}, nil)

	assert.False(t, d.CanUndo())
	assert.Error(t, d.Undo(), "nothing to undo")
	assert.Error(t, d.UndoN(0), "n must be positive")
	assert.Error(t, d.UndoN(1), "no history yet")

	require.NoError(t, d.ChooseBaseRank(0))
	require.NoError(t, d.Draw())
	assert.Equal(t, 5, d.GetBaseRank())

	assert.Error(t, d.UndoN(5), "more than the history holds")
	require.NoError(t, d.UndoN(2))
	assert.Equal(t, 0, d.GetBaseRank(), "the choice itself is undone")
	assert.True(t, d.IsAwaitingBaseRank())
	assert.Len(t, d.GetReserve()[0], 1)
	assert.Equal(t, 1, d.GetStockCount())
	assert.Empty(t, d.GetWaste())
}

func TestDuchess_ActionLogUsesZeroBasedIndices(t *testing.T) {
	d := newTestDuchess()
	var f [DuchessFoundationCnt][]*Card
	setDuchessBoard(d, 0, f, [][]*Card{
		{NewCard(CardDesignHeart, 5, true)},
		{NewCard(CardDesignHeart, 6, true)},
	}, [][]*Card{
		{NewCard(CardDesignSpade, 9, true)},
		{NewCard(CardDesignHeart, 8, true)},
	}, []*Card{NewCard(CardDesignClover, 4, true)}, nil)

	require.NoError(t, d.ChooseBaseRank(0))
	require.NoError(t, d.MoveReserveToFoundation(1))
	require.NoError(t, d.MoveTableauToTableau(1, -1, 0))
	require.NoError(t, d.Draw())

	log := d.GetActionLog()
	require.Len(t, log, 4)
	assert.Equal(t, "開始ランクを5に決定（リザーブ0）", log[0].Detail)
	assert.Equal(t, "リザーブ1→基礎札2", log[1].Detail, "heart is foundation 2, 0-indexed")
	assert.Equal(t, "タブロー列1→タブロー列0(1枚)", log[2].Detail)
	assert.Equal(t, "山札→ウェイスト", log[3].Detail)
}

func TestDuchess_JSONRoundTrip(t *testing.T) {
	d := newTestDuchess()
	require.NoError(t, d.ChooseBaseRank(0))
	data, err := json.Marshal(d)
	require.NoError(t, err)

	restored := NewDuchess(NewTrumpCards(0))
	require.NoError(t, json.Unmarshal(data, restored))
	assert.Equal(t, d.GetPhase(), restored.GetPhase())
	assert.Equal(t, d.GetMoveCount(), restored.GetMoveCount())
	assert.Equal(t, d.GetBaseRank(), restored.GetBaseRank())
	assert.Equal(t, d.GetStockCount(), restored.GetStockCount())
	assert.Len(t, restored.GetReserve()[0], DuchessReserveFanSize-1)
}

func TestDuchess_UnmarshalRejectsOutOfRangeState(t *testing.T) {
	d := NewDuchess(NewTrumpCards(0))
	assert.Error(t, json.Unmarshal([]byte(`nope`), d))
	assert.Error(t, json.Unmarshal([]byte(`{"ps":99}`), d))
	assert.Error(t, json.Unmarshal([]byte(`{"mc":-1}`), d))
	assert.Error(t, json.Unmarshal([]byte(`{"br":-1}`), d), "base rank below zero")
	assert.Error(t, json.Unmarshal([]byte(`{"br":14}`), d), "base rank above King")
	// 0 means "not yet chosen", so it has to stay legal.
	require.NoError(t, json.Unmarshal([]byte(`{"br":0}`), d))

	card := `{"d":0,"v":1,"u":true}`
	big := func(n int) string {
		s := `[` + card
		for range n {
			s += `,` + card
		}
		return s + `]`
	}
	assert.Error(t, json.Unmarshal([]byte(`{"st":`+big(CardCnt)+`}`), d), "stock too large")
	assert.Error(t, json.Unmarshal([]byte(`{"ws":`+big(CardCnt)+`}`), d), "waste too large")
	assert.Error(t, json.Unmarshal([]byte(`{"rs":[`+big(CardCnt)+`]}`), d), "reserve too large")
	assert.Error(t, json.Unmarshal([]byte(`{"fd":[`+big(CardValueMax)+`]}`), d), "foundation too large")
}

func TestDuchess_NewDefaultUsesAStandardDeck(t *testing.T) {
	d := NewDefaultDuchess()
	d.Reset()
	total := d.GetStockCount() + len(d.GetWaste())
	for _, r := range d.GetReserve() {
		total += len(r)
	}
	for _, f := range d.GetFoundation() {
		total += len(f)
	}
	for _, col := range d.GetTableau() {
		total += len(col)
	}
	assert.Equal(t, CardCnt, total)
}

func TestDuchess_SuitIndexAndRunHelpers(t *testing.T) {
	assert.Equal(t, 0, DuchessSuitIndex(CardDesignSpade))
	assert.Equal(t, 2, DuchessSuitIndex(CardDesignHeart))
	assert.Equal(t, -1, DuchessSuitIndex(CardDesignJoker))

	assert.True(t, duchessIsRun(nil), "an empty run is vacuously a run")
	assert.True(t, duchessIsRun([]*DuchessTableauCard{
		{Card: NewCard(CardDesignSpade, 5, true)},
	}), "a single card is always a run")
	assert.False(t, duchessIsRun([]*DuchessTableauCard{
		{Card: NewCard(CardDesignSpade, 5, true)}, {Card: nil},
	}), "a nil card is never part of a run")
	assert.True(t, duchessIsRun([]*DuchessTableauCard{
		{Card: NewCard(CardDesignSpade, 1, true)}, {Card: NewCard(CardDesignHeart, CardValueMax, true)},
	}), "A then K wraps around")
}

// #5557: ページは「組札に2枚以上乗っているか」で自動送りボタンを活性化していたが、
// それはドメインの条件ではない。Duchess は種札を置かないので 1 枚でも次を送れる
// ことがあり、逆に何枚乗っていても送れる札が無ければ AutoComplete は失敗する。
func TestDuchess_CanAutoComplete(t *testing.T) {
	d := NewDuchess(NewTrumpCards(0))
	d.Reset()

	// 基準ランク未選択のうちは何も送れない。
	assert.False(t, d.CanAutoComplete(), "基準ランク未選択")

	// 基準ランクを選ぶと、その1枚が組札に乗る。
	require.NoError(t, d.ChooseBaseRank(0))
	require.False(t, d.IsAwaitingBaseRank())

	// **CanAutoComplete と AutoComplete の可否が一致すること。**
	// ここが食い違うと「押せるのに拒否される」か「押せないのに送れる」になる。
	can := d.CanAutoComplete()
	err := d.AutoComplete()
	if can {
		assert.NoError(t, err)
	} else {
		assert.Error(t, err)
	}
}

// **どの局面でも一致していること。**1局面の一致は偶然でも起きる。
func TestDuchess_CanAutoCompleteMatchesAutoComplete(t *testing.T) {
	for seed := 0; seed < 20; seed++ {
		d := NewDuchess(NewTrumpCards(0))
		d.Reset()
		require.NoError(t, d.ChooseBaseRank(seed%DuchessReserveCnt))
		for step := 0; step < 5; step++ {
			can := d.CanAutoComplete()
			err := d.AutoComplete()
			if can {
				require.NoError(t, err, "seed %d step %d", seed, step)
			} else {
				require.Error(t, err, "seed %d step %d", seed, step)
				break
			}
		}
	}
}
