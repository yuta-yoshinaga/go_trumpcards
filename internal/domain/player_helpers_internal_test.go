package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
