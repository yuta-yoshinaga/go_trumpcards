package domain

import (
	"fmt"
	"sort"
)

// handHolder is the minimal interface for a player whose hand can be sorted in
// place: enumerate the cards, clear the hand, and re-add them.
type handHolder interface {
	GetCardsSize() int
	GetCard(int) *Card
	Reset()
	AddCard(*Card)
}

// sortPlayerHand sorts p's hand by the given comparator, folding the
// "extract cards → sort.Slice → Reset → re-add" boilerplate that was duplicated
// across ~80 games (issue #4298) into one place. less reports whether card ci
// should sort before card cj.
func sortPlayerHand[T handHolder](p T, less func(ci, cj *Card) bool) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sort.Slice(cards, func(i, j int) bool { return less(cards[i], cards[j]) })
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// finishable is the minimal interface satisfied by OldMaidPlayer,
// SevensPlayer, DaifugoPlayer, etc.
type finishable interface {
	GetIsFinished() bool
}

// humanReporter is the minimal view needed to tell the human player apart from
// the CPUs.
type humanReporter interface {
	GetIsHuman() bool
}

// findHumanIdx returns the index of the human player, or -1 when every seat is
// a CPU. 62 games had written this loop out; the receiver was unused in all of
// them, which is what made it free-function-shaped. See issue #5185.
func findHumanIdx[T humanReporter](players []T) int {
	for i, p := range players {
		if p.GetIsHuman() {
			return i
		}
	}
	return -1
}

// isHumanTurn reports whether the seat at idx is the human, returning false
// rather than panicking when idx is outside the roster. 68 games had this exact
// body; another 81 differ for real reasons -- some omit the bounds check (and so
// panic where this returns false), and several gate on game state such as
// `phase == XPhasePlay` or `!gameEndFlag`. Those keep their own. See issue #5185.
func isHumanTurn[T humanReporter](players []T, idx int) bool {
	if idx < 0 || idx >= len(players) {
		return false
	}
	return players[idx].GetIsHuman()
}

// playerName renders a seat for display: "You" for the human, "CPU <idx>" for
// the rest, and "Player <idx>" when idx is outside the roster. 91 games spelled
// this out identically.
//
// The out-of-range branch is not defensive padding -- action-log entries carry
// a playerIdx of -1 for system events, and several presenters format a seat
// before the roster is populated. Callers rely on getting a string back rather
// than a panic. See issue #5185.
func playerName[T humanReporter](players []T, idx int) string {
	if idx < 0 || idx >= len(players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// nextActivePlayer performs a circular search for the next non-finished player.
// direction: 1 = forward, -1 = reverse (e.g. Daifugo 9-reverse).
// Returns -1 if no active player is found.
func nextActivePlayer[T finishable](players []T, from, direction int) int {
	n := len(players)
	for i := 1; i <= n; i++ {
		next := ((from+i*direction)%n + n) % n
		if !players[next].GetIsFinished() {
			return next
		}
	}
	return -1
}

// resettable is satisfied by player types whose hands can be cleared
// and whose finish flag can be toggled (OldMaidPlayer, DaifugoPlayer, etc.).
type resettable interface {
	Reset()
	SetIsFinished(bool)
}

// resetPlayers resets all players' hands and clears their finish flags.
// If extra is non-nil it is called per player for game-specific cleanup.
func resetPlayers[T resettable](players []T, extra func(T)) {
	for _, p := range players {
		p.Reset()
		p.SetIsFinished(false)
		if extra != nil {
			extra(p)
		}
	}
}

// countPlayers counts players matching a predicate.
func countPlayers[T any](players []T, pred func(T) bool) int {
	cnt := 0
	for _, p := range players {
		if pred(p) {
			cnt++
		}
	}
	return cnt
}
