package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- nextActivePlayer tests ---

func TestNextActivePlayer_ForwardNormal(t *testing.T) {
	players := []*OldMaidPlayer{
		NewOldMaidPlayer(true),
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
	}
	got := nextActivePlayer(players, 0, 1)
	assert.Equal(t, 1, got)
}

func TestNextActivePlayer_ForwardSkipFinished(t *testing.T) {
	players := []*OldMaidPlayer{
		NewOldMaidPlayer(true),
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
	}
	players[1].SetIsFinished(true)
	got := nextActivePlayer(players, 0, 1)
	assert.Equal(t, 2, got)
}

func TestNextActivePlayer_ForwardWrapAround(t *testing.T) {
	players := []*OldMaidPlayer{
		NewOldMaidPlayer(true),
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
	}
	players[1].SetIsFinished(true)
	got := nextActivePlayer(players, 2, 1)
	assert.Equal(t, 0, got)
}

func TestNextActivePlayer_ForwardAllFinished(t *testing.T) {
	players := []*OldMaidPlayer{
		NewOldMaidPlayer(true),
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
	}
	for _, p := range players {
		p.SetIsFinished(true)
	}
	got := nextActivePlayer(players, 0, 1)
	assert.Equal(t, -1, got)
}

func TestNextActivePlayer_ReverseNormal(t *testing.T) {
	players := []*SevensPlayer{
		NewSevensPlayer(true),
		NewSevensPlayer(false),
		NewSevensPlayer(false),
		NewSevensPlayer(false),
	}
	got := nextActivePlayer(players, 2, -1)
	assert.Equal(t, 1, got)
}

func TestNextActivePlayer_ReverseSkipFinished(t *testing.T) {
	players := []*SevensPlayer{
		NewSevensPlayer(true),
		NewSevensPlayer(false),
		NewSevensPlayer(false),
		NewSevensPlayer(false),
	}
	players[1].SetIsFinished(true)
	got := nextActivePlayer(players, 2, -1)
	assert.Equal(t, 0, got)
}

func TestNextActivePlayer_ReverseWrapAround(t *testing.T) {
	players := []*SevensPlayer{
		NewSevensPlayer(true),
		NewSevensPlayer(false),
		NewSevensPlayer(false),
		NewSevensPlayer(false),
	}
	players[3].SetIsFinished(true)
	got := nextActivePlayer(players, 0, -1)
	assert.Equal(t, 2, got)
}

func TestNextActivePlayer_ReverseAllFinished(t *testing.T) {
	players := []*SevensPlayer{
		NewSevensPlayer(true),
		NewSevensPlayer(false),
		NewSevensPlayer(false),
		NewSevensPlayer(false),
	}
	for _, p := range players {
		p.SetIsFinished(true)
	}
	got := nextActivePlayer(players, 1, -1)
	assert.Equal(t, -1, got)
}

// --- countPlayers tests ---

func TestCountPlayers_EmptySlice(t *testing.T) {
	var players []*OldMaidPlayer
	got := countPlayers(players, func(p *OldMaidPlayer) bool { return !p.GetIsFinished() })
	assert.Equal(t, 0, got)
}

func TestCountPlayers_AllMatch(t *testing.T) {
	players := []*OldMaidPlayer{
		NewOldMaidPlayer(true),
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
	}
	got := countPlayers(players, func(p *OldMaidPlayer) bool { return !p.GetIsFinished() })
	assert.Equal(t, 3, got)
}

func TestCountPlayers_NoneMatch(t *testing.T) {
	players := []*OldMaidPlayer{
		NewOldMaidPlayer(true),
		NewOldMaidPlayer(false),
	}
	for _, p := range players {
		p.SetIsFinished(true)
	}
	got := countPlayers(players, func(p *OldMaidPlayer) bool { return !p.GetIsFinished() })
	assert.Equal(t, 0, got)
}

func TestCountPlayers_SomeMatch(t *testing.T) {
	players := []*OldMaidPlayer{
		NewOldMaidPlayer(true),
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
	}
	players[1].SetIsFinished(true)
	got := countPlayers(players, func(p *OldMaidPlayer) bool { return !p.GetIsFinished() })
	assert.Equal(t, 2, got)
}

// --- sortPlayerHand tests ---

// sortTestHand is a minimal handHolder for exercising sortPlayerHand
// independently of any concrete game player type.
type sortTestHand struct{ cards []*Card }

func (h *sortTestHand) GetCardsSize() int   { return len(h.cards) }
func (h *sortTestHand) GetCard(i int) *Card { return h.cards[i] }
func (h *sortTestHand) Reset()              { h.cards = nil }
func (h *sortTestHand) AddCard(c *Card)     { h.cards = append(h.cards, c) }

func TestSortPlayerHand(t *testing.T) {
	h := &sortTestHand{cards: []*Card{
		NewCard(1, 5, false),
		NewCard(1, 2, false),
		NewCard(1, 9, false),
	}}
	sortPlayerHand(h, func(ci, cj *Card) bool { return ci.GetValue() < cj.GetValue() })

	// No cards lost in the Reset/re-add round-trip.
	assert.Equal(t, 3, h.GetCardsSize())
	assert.Equal(t, 2, h.GetCard(0).GetValue())
	assert.Equal(t, 5, h.GetCard(1).GetValue())
	assert.Equal(t, 9, h.GetCard(2).GetValue())
}

func TestSortPlayerHand_Empty(t *testing.T) {
	h := &sortTestHand{}
	sortPlayerHand(h, func(ci, cj *Card) bool { return ci.GetValue() < cj.GetValue() })
	assert.Equal(t, 0, h.GetCardsSize())
}

type fakeSeat struct{ human bool }

func (f fakeSeat) GetIsHuman() bool { return f.human }

func TestFindHumanIdx(t *testing.T) {
	assert.Equal(t, 0, findHumanIdx([]fakeSeat{{true}, {false}, {false}}))
	assert.Equal(t, 2, findHumanIdx([]fakeSeat{{false}, {false}, {true}}))
}

// -1 rather than 0 when nobody is human: callers compare against -1, and
// returning a valid-looking index would silently designate seat 0 as the human.
func TestFindHumanIdx_AllCPU(t *testing.T) {
	assert.Equal(t, -1, findHumanIdx([]fakeSeat{{false}, {false}}))
	assert.Equal(t, -1, findHumanIdx([]fakeSeat{}))
}

// The first human wins, matching what the 62 hand-written loops did.
func TestFindHumanIdx_FirstHumanWins(t *testing.T) {
	assert.Equal(t, 1, findHumanIdx([]fakeSeat{{false}, {true}, {true}}))
}

func TestPlayerName(t *testing.T) {
	seats := []fakeSeat{{true}, {false}, {false}}

	assert.Equal(t, "You", playerName(seats, 0))
	assert.Equal(t, "CPU 1", playerName(seats, 1))
	assert.Equal(t, "CPU 2", playerName(seats, 2))
}

// Out-of-range is a real code path, not padding: action-log entries use -1 for
// system events, and presenters format seats before the roster is filled.
// Callers want a string, not a panic.
func TestPlayerName_OutOfRange(t *testing.T) {
	seats := []fakeSeat{{true}, {false}}

	assert.Equal(t, "Player -1", playerName(seats, -1), "system events use -1")
	assert.Equal(t, "Player 2", playerName(seats, 2), "one past the end")
	assert.Equal(t, "Player 0", playerName([]fakeSeat{}, 0), "empty roster")
}

// Whichever seat is human gets "You" -- it is not assumed to be seat 0.
func TestPlayerName_HumanNeedNotBeFirst(t *testing.T) {
	seats := []fakeSeat{{false}, {false}, {true}}

	assert.Equal(t, "CPU 0", playerName(seats, 0))
	assert.Equal(t, "You", playerName(seats, 2))
}

func TestIsHumanTurn(t *testing.T) {
	seats := []fakeSeat{{false}, {true}, {false}}

	assert.True(t, isHumanTurn(seats, 1))
	assert.False(t, isHumanTurn(seats, 0))
	assert.False(t, isHumanTurn(seats, 2))
}

// Out-of-range returns false instead of panicking. 81 other games deliberately
// keep their own version -- 22 of them omit this check and would panic here --
// so the boundary is the reason those were not folded in.
func TestIsHumanTurn_OutOfRangeIsFalseNotPanic(t *testing.T) {
	seats := []fakeSeat{{true}}

	assert.NotPanics(t, func() { isHumanTurn(seats, -1) })
	assert.False(t, isHumanTurn(seats, -1))
	assert.False(t, isHumanTurn(seats, 1))
	assert.False(t, isHumanTurn([]fakeSeat{}, 0))
}

type snap struct{ stale bool }

func staleOf(s snap) bool { return s.stale }

func TestUndoToEscape_NotStalematedNeedsNoUndo(t *testing.T) {
	assert.Equal(t, 0, undoToEscape(false, []snap{{true}, {true}}, staleOf),
		"not stalemated: nothing to undo, regardless of history")
}

// The distance is counted from the end, so the most recent non-stalemate
// snapshot one step back is 1 undo away.
func TestUndoToEscape_DistanceToLastGoodSnapshot(t *testing.T) {
	assert.Equal(t, 1, undoToEscape(true, []snap{{false}}, staleOf))
	assert.Equal(t, 2, undoToEscape(true, []snap{{false}, {true}}, staleOf))
	assert.Equal(t, 3, undoToEscape(true, []snap{{false}, {true}, {true}}, staleOf))
}

// The most recent non-stalemate wins, not the earliest.
func TestUndoToEscape_PicksTheNearestEscape(t *testing.T) {
	assert.Equal(t, 2, undoToEscape(true, []snap{{false}, {false}, {true}}, staleOf))
}

func TestUndoToEscape_NoEscapeAnywhere(t *testing.T) {
	assert.Equal(t, -1, undoToEscape(true, []snap{{true}, {true}}, staleOf))
	assert.Equal(t, -1, undoToEscape(true, []snap{}, staleOf), "empty history cannot escape")
}

func TestGetPlayer(t *testing.T) {
	a, b := &fakeSeat{true}, &fakeSeat{false}
	seats := []*fakeSeat{a, b}

	assert.Same(t, a, getPlayer(seats, 0))
	assert.Same(t, b, getPlayer(seats, 1))
}

// Out of range yields the zero value -- nil for the pointer types every game
// uses -- rather than panicking. Callers nil-check, so this is the contract.
func TestGetPlayer_OutOfRangeIsZero(t *testing.T) {
	seats := []*fakeSeat{{true}}

	assert.Nil(t, getPlayer(seats, -1))
	assert.Nil(t, getPlayer(seats, 1))
	assert.Nil(t, getPlayer([]*fakeSeat{}, 0))
}

type fakeHand struct{ cards []*Card }

func (h *fakeHand) GetCardsSize() int   { return len(h.cards) }
func (h *fakeHand) GetCard(i int) *Card { return h.cards[i] }

func TestAllHandsEmpty(t *testing.T) {
	empty := &fakeHand{}
	held := &fakeHand{cards: []*Card{NewCard(1, 5, false)}}

	assert.True(t, allHandsEmpty([]*fakeHand{empty, empty}))
	assert.True(t, allHandsEmpty([]*fakeHand{}), "no seats means no cards outstanding")
	assert.False(t, allHandsEmpty([]*fakeHand{empty, held}), "one seat still holding is enough")
	assert.False(t, allHandsEmpty([]*fakeHand{held, empty}), "position must not matter")
}

func TestHandHasSuit(t *testing.T) {
	// design is the suit; value is deliberately varied so a body that compares
	// the wrong field fails rather than coincidentally passing.
	seats := []*fakeHand{
		{cards: []*Card{NewCard(1, 3, false), NewCard(2, 7, false)}},
		{cards: []*Card{NewCard(3, 1, false)}},
		{},
	}

	assert.True(t, handHasSuit(seats[0], 1))
	assert.True(t, handHasSuit(seats[0], 2), "not just the first card")
	assert.False(t, handHasSuit(seats[0], 3))
	assert.True(t, handHasSuit(seats[1], 3))
	assert.False(t, handHasSuit(seats[2], 1), "empty hand holds no suit")
}

func TestIndexOfPlayerInTrick(t *testing.T) {
	trick := []*TrickCard{
		{PlayerIdx: 2},
		{PlayerIdx: 0},
		{PlayerIdx: 3},
	}

	assert.Equal(t, 0, indexOfPlayerInTrick(trick, 2), "play order, not seat order")
	assert.Equal(t, 1, indexOfPlayerInTrick(trick, 0))
	assert.Equal(t, 2, indexOfPlayerInTrick(trick, 3))
}

// -1 rather than 0: callers compare against -1, and a valid-looking 0 would
// silently claim the seat led the trick.
func TestIndexOfPlayerInTrick_Absent(t *testing.T) {
	trick := []*TrickCard{{PlayerIdx: 1}}

	assert.Equal(t, -1, indexOfPlayerInTrick(trick, 9))
	assert.Equal(t, -1, indexOfPlayerInTrick(nil, 1), "no trick in progress")
}

// The first entry wins if a seat somehow appears twice, matching the 20
// hand-written loops.
func TestIndexOfPlayerInTrick_FirstMatchWins(t *testing.T) {
	trick := []*TrickCard{{PlayerIdx: 5}, {PlayerIdx: 5}}

	assert.Equal(t, 0, indexOfPlayerInTrick(trick, 5))
}

func TestDiscardTop(t *testing.T) {
	a, b := NewCard(1, 5, false), NewCard(2, 9, false)

	assert.Same(t, b, discardTop([]*Card{a, b}), "the top is the last card added")
	assert.Same(t, a, discardTop([]*Card{a}))
}

// nil rather than a panic on an empty pile: callers nil-check, and every one of
// the 22 hand-written bodies returned nil here.
func TestDiscardTop_EmptyPile(t *testing.T) {
	assert.Nil(t, discardTop([]*Card{}))
	assert.Nil(t, discardTop(nil))
}

func TestValidPlayIndices(t *testing.T) {
	hand := &fakeHand{cards: []*Card{
		NewCard(1, 2, false),
		NewCard(2, 7, false),
		NewCard(1, 9, false),
	}}

	// Only the cards of design 1 are playable.
	got := validPlayIndices(hand, func(c *Card) bool { return c.GetDesign() == 1 })
	assert.Equal(t, []int{0, 2}, got)
}

func TestValidPlayIndices_NoneAndAll(t *testing.T) {
	hand := &fakeHand{cards: []*Card{NewCard(1, 2, false), NewCard(1, 3, false)}}

	assert.Empty(t, validPlayIndices(hand, func(*Card) bool { return false }))
	assert.Equal(t, []int{0, 1}, validPlayIndices(hand, func(*Card) bool { return true }))
	assert.Empty(t, validPlayIndices(&fakeHand{}, func(*Card) bool { return true }), "empty hand")
}

// The predicate must see each card, not just the first -- a helper that passed
// the same card every time would still return a plausible-looking slice.
func TestValidPlayIndices_PredicateSeesEveryCard(t *testing.T) {
	hand := &fakeHand{cards: []*Card{NewCard(1, 4, false), NewCard(2, 5, false), NewCard(3, 6, false)}}

	var seen []int
	validPlayIndices(hand, func(c *Card) bool {
		seen = append(seen, c.GetValue())
		return true
	})
	assert.Equal(t, []int{4, 5, 6}, seen)
}

type fakeRoundPlayer struct {
	tricksReset bool
	handReset   bool
	finished    bool
	order       []string
}

func (p *fakeRoundPlayer) ResetTricks() { p.tricksReset = true; p.order = append(p.order, "tricks") }
func (p *fakeRoundPlayer) Reset()       { p.handReset = true; p.order = append(p.order, "hand") }
func (p *fakeRoundPlayer) SetIsFinished(v bool) {
	p.finished = v
	p.order = append(p.order, "finished")
}

func TestResetPlayerRound(t *testing.T) {
	p := &fakeRoundPlayer{finished: true}

	resetPlayerRound(p)

	assert.True(t, p.tricksReset, "tricks must be cleared")
	assert.True(t, p.handReset, "hand must be cleared")
	assert.False(t, p.finished, "the finished flag must be cleared, not merely set")
	// Order is pinned because the 32 hand-written bodies all did tricks, then
	// hand, then flag -- a player whose Reset depends on that order would break
	// silently otherwise.
	assert.Equal(t, []string{"tricks", "hand", "finished"}, p.order)
}

type fakeHandSorter struct{ sorted []int }

func (g *fakeHandSorter) sortHand(i int) { g.sorted = append(g.sorted, i) }

func TestSortHands(t *testing.T) {
	g := &fakeHandSorter{}

	sortHands(3, g)

	assert.Equal(t, []int{0, 1, 2}, g.sorted, "every seat, in order")
}

func TestSortHands_NoSeats(t *testing.T) {
	g := &fakeHandSorter{}

	sortHands(0, g)

	assert.Empty(t, g.sorted)
}

type fakeUndoer struct {
	calls  int
	failAt int // 1-based step that returns an error; 0 never fails
}

func (u *fakeUndoer) Undo() error {
	u.calls++
	if u.failAt != 0 && u.calls == u.failAt {
		return assert.AnError
	}
	return nil
}

func TestUndoN(t *testing.T) {
	u := &fakeUndoer{}

	require.NoError(t, undoN(u, 3))
	assert.Equal(t, 3, u.calls, "exactly n steps")
}

// Zero and negative are both no-ops. `for i := range n` iterates zero times for
// a negative n, matching the `for i := 0; i < n; i++` bodies this replaces --
// Gaps has a test pinning that a negative count undoes nothing.
func TestUndoN_ZeroOrNegativeIsNoop(t *testing.T) {
	u := &fakeUndoer{}

	require.NoError(t, undoN(u, 0))
	require.NoError(t, undoN(u, -3))
	assert.Equal(t, 0, u.calls)
}

// The failing step number appears in the message and the cause is wrapped, so
// callers can still errors.Is the original.
func TestUndoN_StopsAtTheFailingStep(t *testing.T) {
	u := &fakeUndoer{failAt: 2}

	err := undoN(u, 5)

	require.Error(t, err)
	assert.Equal(t, 2, u.calls, "must stop, not run all 5")
	assert.Contains(t, err.Error(), "undo step 2 failed")
	assert.ErrorIs(t, err, assert.AnError, "the cause must stay wrapped")
}

func TestUndoNChecked(t *testing.T) {
	u := &fakeUndoer{}

	require.NoError(t, undoNChecked(u, 2, 5))
	assert.Equal(t, 2, u.calls)
}

// n and the history length are both rejected before any Undo runs, so a bad
// request cannot leave the game half-rewound.
func TestUndoNChecked_RejectsBeforeUndoing(t *testing.T) {
	for _, tc := range []struct {
		name       string
		n, history int
		want       string
	}{
		{"zero", 0, 5, "n must be positive"},
		{"negative", -1, 5, "n must be positive"},
		{"beyond history", 6, 5, "not enough history"},
		{"empty history", 1, 0, "not enough history"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			u := &fakeUndoer{}
			err := undoNChecked(u, tc.n, tc.history)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.want)
			assert.Equal(t, 0, u.calls, "must not undo anything")
		})
	}
}

func TestUndoNChecked_StopsOnFailure(t *testing.T) {
	u := &fakeUndoer{failAt: 2}

	require.Error(t, undoNChecked(u, 4, 9))
	assert.Equal(t, 2, u.calls)
}

func TestCanPlaceOnFoundationPile(t *testing.T) {
	// An empty foundation takes only an ace.
	assert.True(t, canPlaceOnFoundationPile(nil, NewCard(1, 1, false)))
	assert.False(t, canPlaceOnFoundationPile(nil, NewCard(1, 2, false)))

	pile := []*Card{NewCard(1, 1, false)}
	assert.True(t, canPlaceOnFoundationPile(pile, NewCard(1, 2, false)), "same suit, next rank")
	assert.False(t, canPlaceOnFoundationPile(pile, NewCard(2, 2, false)), "wrong suit")
	assert.False(t, canPlaceOnFoundationPile(pile, NewCard(1, 3, false)), "skips a rank")
	assert.False(t, canPlaceOnFoundationPile(pile, NewCard(1, 1, false)), "same rank")
}

func TestBySuitThenValue(t *testing.T) {
	assert.True(t, bySuitThenValue(NewCard(1, 9, false), NewCard(2, 2, false)), "suit wins over value")
	assert.False(t, bySuitThenValue(NewCard(2, 2, false), NewCard(1, 9, false)))
	assert.True(t, bySuitThenValue(NewCard(1, 3, false), NewCard(1, 5, false)), "same suit falls back to value")
	assert.False(t, bySuitThenValue(NewCard(1, 5, false), NewCard(1, 3, false)))
}

type fakeSeatHand struct {
	human bool
	cards []*Card
}

func (h *fakeSeatHand) GetIsHuman() bool    { return h.human }
func (h *fakeSeatHand) GetCardsSize() int   { return len(h.cards) }
func (h *fakeSeatHand) GetCard(i int) *Card { return h.cards[i] }
func (h *fakeSeatHand) Reset()              { h.cards = nil }
func (h *fakeSeatHand) AddCard(c *Card)     { h.cards = append(h.cards, c) }

func TestSortHumanHands(t *testing.T) {
	human := &fakeSeatHand{human: true, cards: []*Card{
		NewCard(2, 5, false), NewCard(1, 9, false), NewCard(1, 3, false),
	}}
	cpu := &fakeSeatHand{cards: []*Card{NewCard(2, 5, false), NewCard(1, 9, false)}}

	sortHumanHands([]*fakeSeatHand{cpu, human})

	assert.Equal(t, []int{1, 1, 2}, []int{
		human.GetCard(0).GetDesign(), human.GetCard(1).GetDesign(), human.GetCard(2).GetDesign(),
	}, "human hand sorted by suit")
	assert.Equal(t, 3, human.GetCard(0).GetValue(), "then by value within a suit")

	// CPU hands are deliberately left alone -- the 5 games only sorted the human's.
	assert.Equal(t, 2, cpu.GetCard(0).GetDesign(), "cpu hand must not be reordered")
}

func TestLongestSuit(t *testing.T) {
	h := &fakeHand{cards: []*Card{
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignSpade, 7, false),
	}}
	assert.Equal(t, CardDesignHeart, longestSuit(h))
}

// Ties go to the first suit in spade/clover/heart/diamond order, and an empty
// hand yields spade -- both match the 7 hand-written bodies.
func TestLongestSuit_TiesAndEmpty(t *testing.T) {
	tie := &fakeHand{cards: []*Card{
		NewCard(CardDesignDiamond, 2, false),
		NewCard(CardDesignClover, 3, false),
	}}
	assert.Equal(t, CardDesignClover, longestSuit(tie), "clover precedes diamond")
	assert.Equal(t, CardDesignSpade, longestSuit(&fakeHand{}), "empty hand")
}

type fakeWritableHand struct{ cards []*Card }

func (h *fakeWritableHand) GetCardsSize() int { return len(h.cards) }
func (h *fakeWritableHand) AddCard(c *Card)   { h.cards = append(h.cards, c) }
func (h *fakeWritableHand) RemoveCard(i int) *Card {
	c := h.cards[i]
	h.cards = append(h.cards[:i], h.cards[i+1:]...)
	return c
}

func TestSetHandForTest(t *testing.T) {
	h := &fakeWritableHand{cards: []*Card{NewCard(1, 1, false), NewCard(2, 2, false)}}

	setHandForTest(h, []*Card{NewCard(3, 7, false)})

	require.Equal(t, 1, h.GetCardsSize(), "the old hand is replaced, not appended to")
	assert.Equal(t, 7, h.cards[0].GetValue())
}

func TestSetHandForTest_EmptyClearsTheHand(t *testing.T) {
	h := &fakeWritableHand{cards: []*Card{NewCard(1, 1, false)}}

	setHandForTest(h, nil)

	assert.Equal(t, 0, h.GetCardsSize())
}

// GetPlayer returns a typed nil for an out-of-range seat. A typed nil inside an
// interface is not nil, so this guard only works because the constraint is *T.
func TestSetHandForTest_NilPlayerIsIgnored(t *testing.T) {
	var missing *fakeWritableHand

	assert.NotPanics(t, func() { setHandForTest(missing, []*Card{NewCard(1, 1, false)}) })
}

type fakeMover struct {
	loggedTurn []int
	moveCount  *int
	cleared    int
	stalemated int
	lastCards  []*Card
}

func (m *fakeMover) appendLog(_, _ string, cards []*Card) {
	// Records the move count as seen from inside appendLog, which is what the
	// solitaires use as the entry's TurnNumber.
	m.loggedTurn = append(m.loggedTurn, *m.moveCount)
	m.lastCards = cards
}
func (m *fakeMover) checkGameClear() { m.cleared++ }
func (m *fakeMover) checkStalemate() { m.stalemated++ }

// The increment must happen before appendLog: these games pass moveCount as the
// log entry's TurnNumber, so reordering would silently number every entry one
// behind.
func TestAfterMove_IncrementsBeforeLogging(t *testing.T) {
	n := 4
	m := &fakeMover{moveCount: &n}

	afterMove(&n, m, "play", "detail", nil)

	assert.Equal(t, 5, n)
	assert.Equal(t, []int{5}, m.loggedTurn, "appendLog must see the incremented count")
	assert.Equal(t, 1, m.cleared)
	assert.Equal(t, 1, m.stalemated)
}

func TestAfterMove_WrapsTheCard(t *testing.T) {
	n := 0
	m := &fakeMover{moveCount: &n}
	c := NewCard(1, 5, false)

	afterMove(&n, m, "play", "d", c)
	require.Len(t, m.lastCards, 1)
	assert.Same(t, c, m.lastCards[0])

	afterMove(&n, m, "play", "d", nil)
	assert.Nil(t, m.lastCards, "a nil card logs no cards, not a one-element slice of nil")
}

func TestFindNextActiveHelper(t *testing.T) {
	seats := []*fakeBettor{{}, {folded: true}, {}, {allIn: true}}

	assert.Equal(t, 2, findNextActive(seats, 0), "skips folded and all-in")
	assert.Equal(t, 0, findNextActive(seats, 2), "wraps around")
}

// With nobody active it returns the next seat rather than -1, matching the 7
// hand-written bodies -- callers index with the result.
func TestFindNextActiveHelper_NoneActive(t *testing.T) {
	seats := []*fakeBettor{{folded: true}, {folded: true}}
	assert.Equal(t, 1, findNextActive(seats, 0))
}

func TestDrawFromDeck(t *testing.T) {
	deck := []*Card{NewCard(1, 1, false), NewCard(1, 2, false)}
	drawn := 0

	c := drawFromDeck(deck, &drawn)
	require.NotNil(t, c)
	assert.Equal(t, 1, c.GetValue())
	assert.True(t, c.GetDraw(), "the card is marked drawn")
	assert.Equal(t, 1, drawn, "the cursor advances")

	require.NotNil(t, drawFromDeck(deck, &drawn))
	assert.Nil(t, drawFromDeck(deck, &drawn), "exhausted deck yields nil")
	assert.Equal(t, 2, drawn, "the cursor does not run past the deck")
}

type fakeBettor struct {
	folded bool
	allIn  bool
}

func (b *fakeBettor) GetFolded() bool { return b.folded }
func (b *fakeBettor) GetAllIn() bool  { return b.allIn }

func TestCopyOf(t *testing.T) {
	orig := []int{1, 2, 3}
	got := copyOf(orig)

	assert.Equal(t, orig, got)
	got[0] = 99
	assert.Equal(t, 1, orig[0], "the copy must not alias the original")
}

func TestCopyOf_EmptyAndNil(t *testing.T) {
	assert.Empty(t, copyOf([]bool{}))
	assert.Empty(t, copyOf[[]*Card](nil))
}

func TestPercentOf(t *testing.T) {
	assert.Equal(t, 50, percentOf(1, 2))
	assert.Equal(t, 33, percentOf(1, 3), "integer division, matching the bodies replaced")
	assert.Equal(t, 100, percentOf(7, 7))
}

// A zero total yields 0 rather than dividing by zero.
func TestPercentOf_ZeroTotal(t *testing.T) {
	assert.Equal(t, 0, percentOf(5, 0))
	assert.Equal(t, 0, percentOf(0, 0))
}

func TestElemAt(t *testing.T) {
	s := []int{10, 20}
	assert.Equal(t, 10, elemAt(s, 0))
	assert.Equal(t, 20, elemAt(s, 1))
}

// Out of range yields the zero value, which is what the bodies returned.
func TestElemAt_OutOfRange(t *testing.T) {
	s := []int{10}
	assert.Equal(t, 0, elemAt(s, -1))
	assert.Equal(t, 0, elemAt(s, 1))
	assert.Equal(t, 0, elemAt([]int{}, 0))
}

func TestDropLast(t *testing.T) {
	assert.Equal(t, []int{1, 2}, dropLast([]int{1, 2, 3}))
	assert.Empty(t, dropLast([]int{1}))
	assert.Empty(t, dropLast([]int{}), "already empty stays empty rather than panicking")
}

func TestMaxIndexBy(t *testing.T) {
	assert.Equal(t, 1, maxIndexBy([]int{3, 9, 4}, func(v int) int { return v }))
	assert.Equal(t, 0, maxIndexBy([]int{5}, func(v int) int { return v }))
}

// Ties keep the earliest index, and an empty slice yields 0 -- both match the
// bodies replaced, which seeded best := 0 and used a strict >.
func TestMaxIndexBy_TiesAndEmpty(t *testing.T) {
	assert.Equal(t, 0, maxIndexBy([]int{7, 7}, func(v int) int { return v }))
	assert.Equal(t, 0, maxIndexBy([]int{}, func(v int) int { return v }))
}

func TestPokerHandName(t *testing.T) {
	assert.Equal(t, PokerHandNames[0], pokerHandName(0))
	assert.Equal(t, "Unknown", pokerHandName(-1))
	assert.Equal(t, "Unknown", pokerHandName(len(PokerHandNames)))
}

func TestSortHandInPlace(t *testing.T) {
	h := &fakeSeatHand{cards: []*Card{NewCard(2, 9, false), NewCard(1, 3, false)}}

	sortHandInPlace(h, func(cards []*Card) {
		if cards[0].GetValue() > cards[1].GetValue() {
			cards[0], cards[1] = cards[1], cards[0]
		}
	})

	require.Equal(t, 2, h.GetCardsSize(), "no cards lost in the Reset/re-add round trip")
	assert.Equal(t, 3, h.GetCard(0).GetValue())
}

func TestHandHasAny(t *testing.T) {
	h := &fakeHand{cards: []*Card{NewCard(1, 2, false), NewCard(2, 7, false)}}

	assert.True(t, handHasAny(h, func(c *Card) bool { return c.GetValue() == 7 }), "not just the first card")
	assert.False(t, handHasAny(h, func(c *Card) bool { return c.GetValue() == 9 }))
	assert.False(t, handHasAny(&fakeHand{}, func(*Card) bool { return true }), "empty hand")
}

func TestAFDisplay(t *testing.T) {
	assert.Equal(t, "-", afDisplay(0, 0), "no aggression and no calls is undefined, not zero")
	assert.Equal(t, "∞", afDisplay(3, 0), "aggression with no calls is infinite")
	assert.Equal(t, "1.5", afDisplay(3, 2))
	assert.Equal(t, "0.0", afDisplay(0, 4))
}

func TestChipsOfFirst(t *testing.T) {
	assert.Equal(t, 250, chipsOfFirst([]*fakeChipHolder{{chips: 250}, {chips: 10}}))
	assert.Equal(t, 0, chipsOfFirst([]*fakeChipHolder{}), "no seats yields zero, not a panic")
}

type fakeChipHolder struct{ chips int }

func (c *fakeChipHolder) GetChips() int { return c.chips }

type fakeRoundScorer struct {
	order []string
	score int
}

func (p *fakeRoundScorer) SetRoundScore(v int)  { p.score = v; p.order = append(p.order, "score") }
func (p *fakeRoundScorer) ResetTricks()         { p.order = append(p.order, "tricks") }
func (p *fakeRoundScorer) Reset()               { p.order = append(p.order, "hand") }
func (p *fakeRoundScorer) SetIsFinished(_ bool) { p.order = append(p.order, "finished") }

func TestResetRoundScored(t *testing.T) {
	p := &fakeRoundScorer{score: 7}
	resetRoundScored(p)
	assert.Equal(t, 0, p.score)
	assert.Equal(t, []string{"score", "hand", "finished"}, p.order)
}

func TestResetRoundWithTricks(t *testing.T) {
	p := &fakeRoundScorer{score: 7}
	resetRoundWithTricks(p)
	assert.Equal(t, []string{"score", "tricks", "hand", "finished"}, p.order,
		"tricks are cleared between the score and the hand")
}

type fakeRecycleLogger struct{ logged int }

func (g *fakeRecycleLogger) appendLog(_ int, _, _ string, _ []*Card) { g.logged++ }

func TestRecycleDiscardIntoStock(t *testing.T) {
	discard := []*Card{NewCard(1, 1, false), NewCard(1, 2, false), NewCard(1, 3, false)}
	draw := []*Card{NewCard(2, 9, false)}
	top := discard[len(discard)-1]
	g := &fakeRecycleLogger{}

	require.True(t, recycleDiscardIntoStock(&discard, &draw, g))

	require.Len(t, discard, 1)
	assert.Same(t, top, discard[0], "the visible top card stays on the discard pile")
	assert.Len(t, draw, 3, "the rest goes under the draw pile")
	assert.Equal(t, 1, g.logged)
}

// One card or none is not recyclable: taking the top would leave nothing to
// move, so the piles must be untouched and nothing logged.
func TestRecycleDiscardIntoStock_NothingToRecycle(t *testing.T) {
	for _, n := range []int{0, 1} {
		discard := make([]*Card, n)
		for i := range discard {
			discard[i] = NewCard(1, i+1, false)
		}
		draw := []*Card{}
		g := &fakeRecycleLogger{}

		assert.False(t, recycleDiscardIntoStock(&discard, &draw, g), "n=%d", n)
		assert.Len(t, discard, n, "n=%d: discard untouched", n)
		assert.Empty(t, draw, "n=%d: draw untouched", n)
		assert.Equal(t, 0, g.logged, "n=%d: nothing logged", n)
	}
}

func TestRecycleDiscardToDraw(t *testing.T) {
	discard := []*Card{NewCard(1, 1, false), NewCard(1, 2, false), NewCard(1, 3, false)}
	draw := []*Card{NewCard(2, 9, false)}
	top := discard[len(discard)-1]

	recycleDiscardToDraw(&discard, &draw)

	require.Len(t, discard, 1)
	assert.Same(t, top, discard[0], "the visible top card stays")
	// Unlike recycleDiscardIntoStock this replaces the draw pile rather than
	// appending to it -- the four bodies it folds in did exactly that.
	assert.Len(t, draw, 2, "the old draw pile is replaced, not extended")
}

func TestRecycleDiscardToDraw_NothingToRecycle(t *testing.T) {
	for _, n := range []int{0, 1} {
		discard := make([]*Card, n)
		for i := range discard {
			discard[i] = NewCard(1, i+1, false)
		}
		draw := []*Card{NewCard(3, 3, false)}

		recycleDiscardToDraw(&discard, &draw)

		assert.Len(t, discard, n, "n=%d", n)
		assert.Len(t, draw, 1, "n=%d: draw pile untouched", n)
	}
}

func TestSeatHoldingCard(t *testing.T) {
	seats := []*fakeHand{
		{cards: []*Card{NewCard(1, 5, false)}},
		{cards: []*Card{NewCard(CardDesignClover, 2, false)}},
	}
	isTwoOfClubs := func(c *Card) bool {
		return c.GetDesign() == CardDesignClover && c.GetValue() == 2
	}

	assert.Equal(t, 1, seatHoldingCard(seats, isTwoOfClubs))
	assert.Equal(t, -1, seatHoldingCard([]*fakeHand{{}}, isTwoOfClubs), "nobody holding yields -1, not 0")
}

func TestBettingRoundComplete(t *testing.T) {
	seats := []*fakeBettor{{}, {folded: true}, {allIn: true}}

	// Seat 0 is the only one that must act; folded and all-in seats are skipped
	// even though their acted flags are false.
	assert.True(t, bettingRoundComplete(seats, []bool{true, false, false}))
	assert.False(t, bettingRoundComplete(seats, []bool{false, true, true}))
}

func TestBettingRoundComplete_NoSeats(t *testing.T) {
	assert.True(t, bettingRoundComplete([]*fakeBettor{}, nil), "nothing outstanding is complete")
}

type fakeScorer struct{ total int }

func (p *fakeScorer) GetTotalScore() int { return p.total }

func TestTopScorers(t *testing.T) {
	assert.Equal(t, []int{1}, topScorers([]*fakeScorer{{3}, {9}, {4}}))
	assert.Equal(t, []int{0, 2}, topScorers([]*fakeScorer{{9}, {3}, {9}}), "every seat tied for the top")
	assert.Equal(t, []int{}, topScorers([]*fakeScorer{}), "empty roster yields an empty slice, not nil")
}

// Negative totals must not be beaten by the zero value -- seeding best from
// players[0] rather than 0 is what makes that work.
func TestTopScorers_AllNegative(t *testing.T) {
	assert.Equal(t, []int{1}, topScorers([]*fakeScorer{{-9}, {-2}, {-5}}))
}

type fakeTrickScorer struct{ lead int }

func (g *fakeTrickScorer) leadSuit() int { return g.lead }
func (g *fakeTrickScorer) trickScore(c *Card, lead int) int {
	if c.GetDesign() != lead {
		return 0
	}
	return c.GetValue()
}

func TestTrickWinnerByScore(t *testing.T) {
	g := &fakeTrickScorer{lead: 1}
	trick := []*TrickCard{
		{PlayerIdx: 2, Card: NewCard(1, 5, false)},
		{PlayerIdx: 0, Card: NewCard(2, 13, false)}, // off-suit, scores 0
		{PlayerIdx: 3, Card: NewCard(1, 9, false)},
	}
	assert.Equal(t, 3, trickWinnerByScore(trick, g))
	assert.Equal(t, 0, trickWinnerByScore(nil, g), "no cards played")
}

// A tie keeps the earliest play, which is what a strict > gives.
func TestTrickWinnerByScore_TieKeepsEarliest(t *testing.T) {
	g := &fakeTrickScorer{lead: 1}
	trick := []*TrickCard{
		{PlayerIdx: 2, Card: NewCard(1, 7, false)},
		{PlayerIdx: 1, Card: NewCard(1, 7, false)},
	}
	assert.Equal(t, 2, trickWinnerByScore(trick, g))
}

func TestDrawFromPile(t *testing.T) {
	pile := []*Card{NewCard(1, 1, false), NewCard(1, 2, false), NewCard(1, 3, false)}
	hand := &fakeSeatHand{}

	assert.Equal(t, 2, drawFromPile(&pile, hand, 2, func() {}))
	assert.Equal(t, 2, hand.GetCardsSize())
	assert.Len(t, pile, 1, "cards come off the end")
}

// The recycle hook is called when the pile empties, and the draw stops if it
// cannot refill -- otherwise this would loop forever on an exhausted deck.
func TestDrawFromPile_RecyclesThenGivesUp(t *testing.T) {
	pile := []*Card{NewCard(1, 1, false)}
	hand := &fakeSeatHand{}
	recycled := 0

	got := drawFromPile(&pile, hand, 5, func() { recycled++ })

	assert.Equal(t, 1, got, "only what was available")
	assert.Equal(t, 1, recycled, "recycle attempted once, then the loop stops")
}

func TestDealerQualifies(t *testing.T) {
	// A made hand qualifies regardless of the cards.
	assert.True(t, dealerQualifies(PokerHandOnePair, nil))

	ak := []*Card{NewCard(1, 1, false), NewCard(2, 13, false)}
	assert.True(t, dealerQualifies(0, ak), "ace-king qualifies below one pair")
	assert.False(t, dealerQualifies(0, []*Card{NewCard(1, 1, false)}), "ace alone does not")
	assert.False(t, dealerQualifies(0, []*Card{NewCard(1, 13, false)}), "king alone does not")
}

func TestBestSuitFrom(t *testing.T) {
	assert.Equal(t, CardDesignHeart, bestSuitFrom(map[int]int{CardDesignHeart: 3, CardDesignSpade: 1}))
	assert.Equal(t, CardDesignSpade, bestSuitFrom(map[int]int{}), "no cards defaults to spade")
	assert.Equal(t, CardDesignClover, bestSuitFrom(map[int]int{CardDesignClover: 2, CardDesignDiamond: 2}),
		"ties resolve in spade/clover/heart/diamond order")
}

func TestDealUpTo(t *testing.T) {
	tc := NewTrumpCards(0)
	tc.Shuffle()
	cards := []*Card{}

	dealUpTo(&cards, tc, 5)
	assert.Len(t, cards, 5)

	// Already at target: nothing more is drawn.
	dealUpTo(&cards, tc, 5)
	assert.Len(t, cards, 5)
}

func TestPotBet(t *testing.T) {
	assert.Equal(t, 50, potBet(100, 50, 10, 0), "half pot")
	assert.Equal(t, 10, potBet(4, 50, 10, 0), "floored by the big blind")
	assert.Equal(t, 30, potBet(100, 10, 10, 30), "floored by the minimum raise")
}

func TestNextIndexWhere(t *testing.T) {
	s := []bool{false, true, false, true}
	ok := func(v bool) bool { return v }

	assert.Equal(t, 1, nextIndexWhere(s, 0, ok))
	assert.Equal(t, 3, nextIndexWhere(s, 1, ok), "starts after from")
	assert.Equal(t, 1, nextIndexWhere(s, 3, ok), "wraps around")
}

// With nobody eligible it returns from rather than -1, because callers index
// with the result. An empty slice must not divide by zero.
func TestNextIndexWhere_NoneAndEmpty(t *testing.T) {
	assert.Equal(t, 2, nextIndexWhere([]bool{false, false, false}, 2, func(v bool) bool { return v }))
	assert.NotPanics(t, func() { nextIndexWhere([]bool{}, 0, func(bool) bool { return true }) })
}

func TestDrawOrTakeTrump(t *testing.T) {
	deck := NewTrumpCards(0)
	deck.Shuffle()
	var trump *Card

	assert.NotNil(t, drawOrTakeTrump(deck, &trump), "comes off the deck while it lasts")
}

func TestDrawOrTakeTrump_FallsBackToTrumpThenNil(t *testing.T) {
	deck := NewTrumpCards(0)
	deck.Shuffle()
	for deck.DrawCard() != nil { //nolint:revive // drain the deck
	}
	held := NewCard(1, 5, false)
	trump := held

	assert.Same(t, held, drawOrTakeTrump(deck, &trump), "the trump card is dealt once the deck runs out")
	assert.Nil(t, trump, "and is consumed")
	assert.Nil(t, drawOrTakeTrump(deck, &trump), "nothing left")
}

type fakeBeater struct{}

// Higher value wins, but only among cards of the led suit.
func (fakeBeater) cardBeats(a, b *Card, leadSuit int) bool {
	if a.GetDesign() != leadSuit {
		return false
	}
	if b.GetDesign() != leadSuit {
		return true
	}
	return a.GetValue() > b.GetValue()
}

func TestCurrentTrickWinnerCard(t *testing.T) {
	high := NewCard(1, 9, false)
	trick := []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(1, 4, false)},
		{PlayerIdx: 1, Card: NewCard(2, 13, false)}, // off-suit
		{PlayerIdx: 2, Card: high},
	}
	assert.Same(t, high, currentTrickWinnerCard(trick, fakeBeater{}))
	assert.Nil(t, currentTrickWinnerCard(nil, fakeBeater{}))
}

func TestValidateCardIsPlayable(t *testing.T) {
	a, b := NewCard(1, 2, false), NewCard(1, 3, false)
	h := &fakeHand{cards: []*Card{a, b}}

	require.NoError(t, validateCardIsPlayable([]int{0}, h, a))
	// Identity, not equality: an identical-looking card that is not in the hand
	// is rejected, which is what the bodies replaced did.
	assert.Error(t, validateCardIsPlayable([]int{0}, h, NewCard(1, 2, false)))
	assert.Error(t, validateCardIsPlayable([]int{0}, h, b), "b is in the hand but not among the valid indices")
	assert.Error(t, validateCardIsPlayable(nil, h, a), "nothing is playable")
}

type fakeEndgame struct {
	endgame   bool
	satisfies bool
}

func (g fakeEndgame) IsEndgame() bool                         { return g.endgame }
func (g fakeEndgame) cardSatisfiesFollow(_ int, _ *Card) bool { return g.satisfies }

func TestValidateEndgameFollow(t *testing.T) {
	trick := []*TrickCard{{PlayerIdx: 0, Card: NewCard(1, 5, false)}}
	card := NewCard(1, 2, false)

	assert.Error(t, validateEndgameFollow(trick, fakeEndgame{}, 0, nil), "a nil card is rejected first")
	require.NoError(t, validateEndgameFollow(trick, fakeEndgame{endgame: false}, 0, card), "no rule before the endgame")
	require.NoError(t, validateEndgameFollow(nil, fakeEndgame{endgame: true}, 0, card), "no rule when leading")
	assert.Error(t, validateEndgameFollow(trick, fakeEndgame{endgame: true, satisfies: false}, 0, card))
	require.NoError(t, validateEndgameFollow(trick, fakeEndgame{endgame: true, satisfies: true}, 0, card))
}

func TestRemoveIndices(t *testing.T) {
	assert.Equal(t, []int{1, 3}, removeIndices([]int{1, 2, 3, 4}, []int{1, 3}))
	assert.Equal(t, []int{1, 2}, removeIndices([]int{1, 2}, nil), "no indices leaves the slice alone")
	assert.Equal(t, []int{1, 2}, removeIndices([]int{1, 2}, []int{5, -1}), "out-of-range indices are ignored")
}

// Removal must run highest-first; ascending order would shift the later
// indices and delete the wrong elements.
func TestRemoveIndices_OrderIndependent(t *testing.T) {
	assert.Equal(t, []int{20}, removeIndices([]int{10, 20, 30}, []int{0, 2}))
	assert.Equal(t, []int{20}, removeIndices([]int{10, 20, 30}, []int{2, 0}), "input order must not matter")
}
