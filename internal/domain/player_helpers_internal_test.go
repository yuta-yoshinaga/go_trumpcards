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
