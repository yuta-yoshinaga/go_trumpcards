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
