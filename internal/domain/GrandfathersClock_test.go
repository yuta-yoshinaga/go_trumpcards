//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestGrandfathersClock() *GrandfathersClock {
	gc := NewGrandfathersClock(NewTrumpCards(0))
	gc.Reset()
	return gc
}

// setGCBoard installs an exact position. Tests must never lean on the shuffle:
// asserting what is or is not playable against a real deal is a flake waiting
// to happen (see #4467).
func setGCBoard(gc *GrandfathersClock, foundation [GrandfathersClockFoundationCnt][]*Card, cols [][]*Card) {
	gc.foundation = foundation
	for i := range GrandfathersClockTableauCnt {
		gc.tableau[i] = nil
	}
	for i, c := range cols {
		pile := make([]*GrandfathersClockTableauCard, 0, len(c))
		for _, card := range c {
			pile = append(pile, &GrandfathersClockTableauCard{Card: card, FaceUp: true})
		}
		gc.tableau[i] = pile
	}
	gc.phase = GrandfathersClockPhasePlaying
	gc.isStalemate = false
	gc.history = nil
	gc.moveCount = 0
	gc.actionLog = nil
}

// starterBoard seeds the twelve clock faces exactly as a real deal would.
func starterBoard() [GrandfathersClockFoundationCnt][]*Card {
	var f [GrandfathersClockFoundationCnt][]*Card
	for i, s := range grandfathersClockStarters {
		f[i] = []*Card{NewCard(s.design, s.value, true)}
	}
	return f
}

func TestGrandfathersClock_Reset_SeedsTheClockAndDealsEightByFive(t *testing.T) {
	gc := newTestGrandfathersClock()

	seeded := 0
	for i, pile := range gc.GetFoundation() {
		require.Len(t, pile, 1, "clock face %d holds exactly its starter", i)
		want := grandfathersClockStarters[i]
		assert.Equal(t, want.design, pile[0].GetDesign(), "face %d suit", i)
		assert.Equal(t, want.value, pile[0].GetValue(), "face %d rank", i)
		seeded += len(pile)
	}
	assert.Equal(t, GrandfathersClockFoundationCnt, seeded)

	dealt := 0
	for col, pile := range gc.GetTableau() {
		assert.Len(t, pile, GrandfathersClockColumnLen, "column %d", col)
		for _, tc := range pile {
			assert.True(t, tc.FaceUp, "every card is face-up by rule")
		}
		dealt += len(pile)
	}
	assert.Equal(t, 40, dealt)
	// 12 + 40 = 52, which is why this game has no stock at all, despite what
	// #4399's rule 5 says.
	assert.Equal(t, CardCnt, seeded+dealt)
	assert.True(t, gc.AllFaceUp())
	assert.False(t, gc.GetGameEndFlag())
}

// The fixed starters are what make the deal add up: each needs 3 or 4 more
// cards, and those totals come to exactly the 40 dealt to the tableau.
func TestGrandfathersClock_StartersNeedExactlyFortyCards(t *testing.T) {
	total := 0
	for i, s := range grandfathersClockStarters {
		target := GrandfathersClockTargetRank(i)
		steps := 0
		for v := s.value; v != target; v = grandfathersClockNextRank(v) {
			steps++
			require.Less(t, steps, CardValueMax, "face %d never reaches its target", i)
		}
		assert.Contains(t, []int{3, 4}, steps, "face %d needs 3 or 4 cards", i)
		total += steps
	}
	assert.Equal(t, GrandfathersClockTableauCnt*GrandfathersClockColumnLen, total)
}

func TestGrandfathersClock_TargetRankIsTheHour(t *testing.T) {
	assert.Equal(t, 1, GrandfathersClockTargetRank(0), "1 o'clock ends on an Ace")
	assert.Equal(t, 12, GrandfathersClockTargetRank(11), "12 o'clock ends on a Queen")
}

func TestGrandfathersClock_FoundationBuildsUpInSuit(t *testing.T) {
	gc := newTestGrandfathersClock()
	f := starterBoard()
	// Face 4 (5 o'clock) starts on 2♥ and wants 5♥.
	setGCBoard(gc, f, [][]*Card{
		{NewCard(CardDesignHeart, 3, true)},
		{NewCard(CardDesignSpade, 3, true)},
		{NewCard(CardDesignHeart, 5, true)},
	})

	assert.Error(t, gc.MoveTableauToFoundation(1, 4), "wrong suit")
	assert.Error(t, gc.MoveTableauToFoundation(2, 4), "not the next rank")
	require.NoError(t, gc.MoveTableauToFoundation(0, 4))
	assert.Len(t, gc.GetFoundation()[4], 2)
}

// K -> A is the wraparound that lets a face started on a high card reach a low
// target; without it the 1..4 o'clock faces could never be finished.
func TestGrandfathersClock_FoundationWrapsFromKingToAce(t *testing.T) {
	gc := newTestGrandfathersClock()
	f := starterBoard()
	// Face 0 (1 o'clock) wants an Ace; put it one step away.
	f[0] = []*Card{NewCard(CardDesignHeart, CardValueMax, true)}
	setGCBoard(gc, f, [][]*Card{{NewCard(CardDesignHeart, 1, true)}})

	require.NoError(t, gc.MoveTableauToFoundation(0, 0))
	assert.True(t, gc.IsFoundationComplete(0))
	assert.Equal(t, 1, grandfathersClockNextRank(CardValueMax))
	assert.Equal(t, 5, grandfathersClockNextRank(4))
}

// A completed face must stop accepting cards, or the wraparound would let it
// lap the whole suit.
func TestGrandfathersClock_CompletedFaceRejectsMoreCards(t *testing.T) {
	gc := newTestGrandfathersClock()
	f := starterBoard()
	// Face 4 (5 o'clock) already sits on its target, 5♥.
	f[4] = []*Card{NewCard(CardDesignHeart, 5, true)}
	setGCBoard(gc, f, [][]*Card{{NewCard(CardDesignHeart, 6, true)}})

	require.True(t, gc.IsFoundationComplete(4))
	assert.Error(t, gc.MoveTableauToFoundation(0, 4), "6♥ would lap a finished face")
}

func TestGrandfathersClock_RejectsInvalidFoundationArguments(t *testing.T) {
	gc := newTestGrandfathersClock()
	setGCBoard(gc, starterBoard(), [][]*Card{{NewCard(CardDesignHeart, 3, true)}})

	assert.Error(t, gc.MoveTableauToFoundation(0, -1), "index out of range")
	assert.Error(t, gc.MoveTableauToFoundation(0, GrandfathersClockFoundationCnt), "index out of range")
	assert.Error(t, gc.MoveTableauToFoundation(-1, 4), "column out of range")
	assert.Error(t, gc.MoveTableauToFoundation(1, 4), "column is empty")

	// An empty clock face can only happen in a restored snapshot; it must not
	// silently accept a card.
	var empty [GrandfathersClockFoundationCnt][]*Card
	setGCBoard(gc, empty, [][]*Card{{NewCard(CardDesignHeart, 3, true)}})
	assert.Error(t, gc.MoveTableauToFoundation(0, 4))
	assert.False(t, gc.IsFoundationComplete(4))
	assert.False(t, gc.IsFoundationComplete(-1), "out of range is never complete")
}

func TestGrandfathersClock_TableauBuildsDownIgnoringSuit(t *testing.T) {
	gc := newTestGrandfathersClock()
	setGCBoard(gc, starterBoard(), [][]*Card{
		{NewCard(CardDesignSpade, 9, true)},
		{NewCard(CardDesignHeart, 8, true)},
		{NewCard(CardDesignHeart, 10, true)},
	})

	assert.Error(t, gc.MoveTableauToTableau(2, 0), "builds down, not up")
	require.NoError(t, gc.MoveTableauToTableau(1, 0), "♥8 onto ♠9 — suit is ignored")
	assert.Len(t, gc.GetTableau()[0], 2)
	assert.Empty(t, gc.GetTableau()[1])
}

func TestGrandfathersClock_EmptyColumnAcceptsAnyCard(t *testing.T) {
	gc := newTestGrandfathersClock()
	setGCBoard(gc, starterBoard(), [][]*Card{
		{NewCard(CardDesignSpade, 9, true), NewCard(CardDesignDiamond, 2, true)},
	})

	require.NoError(t, gc.MoveTableauToTableau(0, 5))
	assert.Len(t, gc.GetTableau()[5], 1)
	assert.Len(t, gc.GetTableau()[0], 1)
}

func TestGrandfathersClock_RejectsInvalidTableauArguments(t *testing.T) {
	gc := newTestGrandfathersClock()
	setGCBoard(gc, starterBoard(), [][]*Card{{NewCard(CardDesignSpade, 9, true)}})

	assert.Error(t, gc.MoveTableauToTableau(-1, 1), "from out of range")
	assert.Error(t, gc.MoveTableauToTableau(0, GrandfathersClockTableauCnt), "to out of range")
	assert.Error(t, gc.MoveTableauToTableau(0, 0), "same column")
	assert.Error(t, gc.MoveTableauToTableau(1, 0), "column is empty")
}

func TestGrandfathersClock_GameClearWhenEveryFaceReachesItsHour(t *testing.T) {
	gc := newTestGrandfathersClock()
	var f [GrandfathersClockFoundationCnt][]*Card
	for i, s := range grandfathersClockStarters {
		f[i] = []*Card{NewCard(s.design, GrandfathersClockTargetRank(i), true)}
	}
	// Roll face 11 (12 o'clock, wants a Queen) back one step.
	last := grandfathersClockStarters[11]
	f[11] = []*Card{NewCard(last.design, 11, true)}
	setGCBoard(gc, f, [][]*Card{{NewCard(last.design, 12, true)}})

	require.NoError(t, gc.MoveTableauToFoundation(0, 11))
	assert.Equal(t, GrandfathersClockPhaseGameClear, gc.GetPhase())
	assert.True(t, gc.GetGameEndFlag())
}

func TestGrandfathersClock_GiveUpEndsTheGameOnce(t *testing.T) {
	gc := newTestGrandfathersClock()
	gc.GiveUp()
	assert.Equal(t, GrandfathersClockPhaseGameOver, gc.GetPhase())
	logLen := len(gc.GetActionLog())

	gc.GiveUp()
	assert.Len(t, gc.GetActionLog(), logLen, "a second give-up is a no-op")

	assert.Error(t, gc.MoveTableauToFoundation(0, 0))
	assert.Error(t, gc.MoveTableauToTableau(0, 1))
	assert.Error(t, gc.AutoComplete())
	assert.Nil(t, gc.GetHint())
}

func TestGrandfathersClock_HintPrefersTheClockFace(t *testing.T) {
	gc := newTestGrandfathersClock()
	setGCBoard(gc, starterBoard(), [][]*Card{
		{NewCard(CardDesignSpade, 9, true)},
		{NewCard(CardDesignHeart, 8, true)}, // a legal tableau move onto ♠9
		{NewCard(CardDesignHeart, 3, true)}, // but ♥3 goes onto face 4 (2♥)
	})

	h := gc.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "foundation", h.ToZone)
	assert.Equal(t, 2, h.FromCol)
	assert.Equal(t, 4, h.ToIdx)
}

func TestGrandfathersClock_HintFallsBackToTheTableau(t *testing.T) {
	gc := newTestGrandfathersClock()
	setGCBoard(gc, starterBoard(), [][]*Card{
		{NewCard(CardDesignSpade, 9, true)},
		{NewCard(CardDesignHeart, 8, true)},
	})

	h := gc.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "tableau", h.ToZone)
	assert.Equal(t, 1, h.FromCol)
	assert.Equal(t, 0, h.ToIdx)
	// The hint must be a move the domain actually accepts.
	require.NoError(t, gc.MoveTableauToTableau(h.FromCol, h.ToIdx))
}

// Shuffling a lone card into an empty column changes nothing, so it must not be
// offered — otherwise the hint loops on a dead board forever.
func TestGrandfathersClock_HintIgnoresLoneCardToEmptyColumnShuffles(t *testing.T) {
	gc := newTestGrandfathersClock()
	setGCBoard(gc, starterBoard(), [][]*Card{
		{NewCard(CardDesignSpade, 9, true)},
	})

	assert.Nil(t, gc.GetHint(), "the only move would be a pointless relocation")
	gc.checkStalemate()
	assert.True(t, gc.IsStalemate())
	assert.Equal(t, -1, gc.UndoToEscape(), "no history to rewind into")

	// With a card buried underneath, moving the top to an empty column does
	// expose something, so it is a real move.
	setGCBoard(gc, starterBoard(), [][]*Card{
		{NewCard(CardDesignDiamond, 2, true), NewCard(CardDesignSpade, 9, true)},
	})
	h := gc.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "tableau", h.ToZone)
}

func TestGrandfathersClock_AutoCompleteOnlyFeedsTheClock(t *testing.T) {
	gc := newTestGrandfathersClock()
	setGCBoard(gc, starterBoard(), [][]*Card{
		{NewCard(CardDesignHeart, 4, true), NewCard(CardDesignHeart, 3, true)},
	})

	require.NoError(t, gc.AutoComplete())
	assert.Len(t, gc.GetFoundation()[4], 3, "2♥ then 3♥ then 4♥")
	assert.Empty(t, gc.GetTableau()[0])
}

// AutoComplete must never make a tableau move, or it would shuffle the board
// instead of clearing it — it drives off foundationHint, not GetHint.
func TestGrandfathersClock_AutoCompleteIgnoresTableauMoves(t *testing.T) {
	gc := newTestGrandfathersClock()
	setGCBoard(gc, starterBoard(), [][]*Card{
		{NewCard(CardDesignSpade, 9, true)},
		{NewCard(CardDesignHeart, 8, true)},
	})

	require.NotNil(t, gc.GetHint(), "a tableau move exists")
	assert.Error(t, gc.AutoComplete(), "but nothing reaches a clock face")
	assert.Equal(t, 0, gc.GetMoveCount())
	assert.Len(t, gc.GetTableau()[0], 1, "the board is untouched")
	assert.Len(t, gc.GetTableau()[1], 1)
}

func TestGrandfathersClock_UndoRestoresBothZones(t *testing.T) {
	gc := newTestGrandfathersClock()
	setGCBoard(gc, starterBoard(), [][]*Card{
		{NewCard(CardDesignHeart, 3, true)},
		{NewCard(CardDesignSpade, 9, true), NewCard(CardDesignHeart, 8, true)},
	})

	assert.False(t, gc.CanUndo())
	assert.Error(t, gc.Undo(), "nothing to undo")
	assert.Error(t, gc.UndoN(0), "n must be positive")
	assert.Error(t, gc.UndoN(1), "no history yet")

	require.NoError(t, gc.MoveTableauToFoundation(0, 4))
	require.NoError(t, gc.MoveTableauToTableau(1, 0))
	assert.True(t, gc.CanUndo())
	assert.Equal(t, 2, gc.GetMoveCount())

	assert.Error(t, gc.UndoN(5), "more than the history holds")
	require.NoError(t, gc.UndoN(2))
	assert.Equal(t, 0, gc.GetMoveCount())
	assert.Len(t, gc.GetFoundation()[4], 1)
	assert.Len(t, gc.GetTableau()[0], 1)
	assert.Len(t, gc.GetTableau()[1], 2)
}

func TestGrandfathersClock_UndoToEscapeCountsBackToAPlayablePosition(t *testing.T) {
	gc := newTestGrandfathersClock()

	// Reaching a dead end takes care here: an empty column accepts anything, so
	// all eight columns must stay occupied. ♥8 onto ♠9 is the only move, and it
	// leaves ♦2 behind rather than emptying its column. The blockers are chosen
	// so no two tops differ by one, and none of them is the next card any clock
	// face wants.
	cols := [][]*Card{
		{NewCard(CardDesignSpade, 9, true)},
		{NewCard(CardDesignDiamond, 2, true), NewCard(CardDesignHeart, 8, true)},
		{NewCard(CardDesignSpade, 5, true)},
		{NewCard(CardDesignHeart, 5, true)},
		{NewCard(CardDesignDiamond, 11, true)},
		{NewCard(CardDesignClover, 11, true)},
		{NewCard(CardDesignSpade, CardValueMax, true)},
		{NewCard(CardDesignHeart, CardValueMax, true)},
	}
	require.Len(t, cols, GrandfathersClockTableauCnt, "no column may be left empty")
	setGCBoard(gc, starterBoard(), cols)

	assert.False(t, gc.IsStalemate())
	assert.Equal(t, 0, gc.UndoToEscape(), "not stalemated yet")

	require.NoError(t, gc.MoveTableauToTableau(1, 0), "♥8 onto ♠9 is the only move")
	assert.True(t, gc.IsStalemate(), "and it strands the board")
	assert.Equal(t, 1, gc.UndoToEscape())

	require.NoError(t, gc.UndoN(gc.UndoToEscape()))
	assert.False(t, gc.IsStalemate())
	assert.Len(t, gc.GetTableau()[1], 2, "♥8 is back on top of ♦2")
}

func TestGrandfathersClock_ActionLogUsesZeroBasedIndices(t *testing.T) {
	gc := newTestGrandfathersClock()
	setGCBoard(gc, starterBoard(), [][]*Card{
		{NewCard(CardDesignHeart, 3, true)},
		{NewCard(CardDesignSpade, 9, true), NewCard(CardDesignHeart, 8, true)},
	})

	require.NoError(t, gc.MoveTableauToFoundation(0, 4))
	require.NoError(t, gc.MoveTableauToTableau(1, 0))

	log := gc.GetActionLog()
	require.Len(t, log, 2)
	assert.Equal(t, "タブロー列0→文字盤4", log[0].Detail)
	assert.Equal(t, "タブロー列1→タブロー列0", log[1].Detail)
}

func TestGrandfathersClock_JSONRoundTrip(t *testing.T) {
	gc := newTestGrandfathersClock()
	data, err := json.Marshal(gc)
	require.NoError(t, err)

	restored := NewGrandfathersClock(NewTrumpCards(0))
	require.NoError(t, json.Unmarshal(data, restored))
	assert.Equal(t, gc.GetPhase(), restored.GetPhase())
	assert.Equal(t, gc.GetMoveCount(), restored.GetMoveCount())
	assert.Len(t, restored.GetFoundation()[0], 1)
	assert.Len(t, restored.GetTableau()[0], GrandfathersClockColumnLen)
}

func TestGrandfathersClock_UnmarshalRejectsOutOfRangeState(t *testing.T) {
	gc := NewGrandfathersClock(NewTrumpCards(0))
	assert.Error(t, json.Unmarshal([]byte(`nope`), gc))
	assert.Error(t, json.Unmarshal([]byte(`{"ps":99}`), gc))
	assert.Error(t, json.Unmarshal([]byte(`{"mc":-1}`), gc))

	card := `{"d":0,"v":1,"u":true}`
	oversizeFoundation := `[` + card
	for range CardValueMax {
		oversizeFoundation += `,` + card
	}
	oversizeFoundation += `]`
	assert.Error(t, json.Unmarshal([]byte(`{"fd":[`+oversizeFoundation+`]}`), gc), "clock face too large")

	tc := `{"c":` + card + `,"f":true}`
	oversizeTableau := `[` + tc
	for range CardCnt {
		oversizeTableau += `,` + tc
	}
	oversizeTableau += `]`
	assert.Error(t, json.Unmarshal([]byte(`{"tb":[`+oversizeTableau+`]}`), gc), "tableau too large")
}

func TestGrandfathersClock_NewDefaultUsesAStandardDeck(t *testing.T) {
	gc := NewDefaultGrandfathersClock()
	gc.Reset()
	total := 0
	for _, f := range gc.GetFoundation() {
		total += len(f)
	}
	for _, col := range gc.GetTableau() {
		total += len(col)
	}
	assert.Equal(t, CardCnt, total)
}

func TestGrandfathersClock_StarterIndexIgnoresNonStarters(t *testing.T) {
	assert.Equal(t, -1, grandfathersClockStarterIndex(nil))
	assert.Equal(t, -1, grandfathersClockStarterIndex(NewCard(CardDesignSpade, 2, true)),
		"2♠ is not one of the twelve; 2♥ is")
	assert.Equal(t, 4, grandfathersClockStarterIndex(NewCard(CardDesignHeart, 2, true)))
}
