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

// undoToEscape reports how many undos are needed to leave a stalemate: 0 when
// not stalemated, the distance back to the most recent non-stalemate snapshot
// otherwise, and -1 when every snapshot in history is also a stalemate.
//
// 40 solitaires had this loop written out. The predicate is passed in because
// each game's snapshot type is its own struct with an unexported isStalemate
// field, and Go constraints cannot require a field -- only a method. The part
// worth sharing is the walk and the `len(history) - i` distance, not the field
// access. See issue #5185.
func undoToEscape[T any](isStalemate bool, history []T, wasStalemate func(T) bool) int {
	if !isStalemate {
		return 0
	}
	for i := len(history) - 1; i >= 0; i-- {
		if !wasStalemate(history[i]) {
			return len(history) - i
		}
	}
	return -1
}

// finishable is the minimal interface satisfied by OldMaidPlayer,
// SevensPlayer, DaifugoPlayer, etc.
type finishable interface {
	GetIsFinished() bool
}

// getPlayer returns the seat at idx, or the zero value (nil for the pointer
// types every game uses) when idx is outside the roster. 152 games had this
// exact body.
//
// Constrained by `any` rather than an interface: the games' player types share
// no method this needs, and the helper only indexes. See issue #5185.
func getPlayer[T any](players []T, idx int) T {
	if idx < 0 || idx >= len(players) {
		var zero T
		return zero
	}
	return players[idx]
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

// handReader is the read-only view of a player's hand: how many cards it holds
// and what each one is. Deliberately narrower than handHolder, which also
// requires Reset/AddCard -- the predicates below never mutate a hand.
type handReader interface {
	GetCardsSize() int
	GetCard(int) *Card
}

// allHandsEmpty reports whether every seat has run out of cards. 18 games had
// this loop written out. An empty roster counts as empty: no seat is holding.
func allHandsEmpty[P handReader](players []P) bool {
	for _, p := range players {
		if p.GetCardsSize() > 0 {
			return false
		}
	}
	return true
}

// handHasSuit reports whether p holds a card of the given design. 30 games had
// this loop written out as a playerHasSuit method, and Schnapsen had a 31st as
// a free function taking the player directly -- which is why this takes the
// player rather than (players, idx): that shape serves both. See issue #5185.
func handHasSuit[P handReader](p P, design int) bool {
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == design {
			return true
		}
	}
	return false
}

// validPlayIndices returns the hand positions of p whose card satisfies ok.
//
// 41 games had this written out: fetch the seat, then feed
// collectValidIndices a closure that indexes back into the same hand. Only the
// predicate differs between them, so that is what is passed in. See issue #5185.
func validPlayIndices[P handReader](p P, ok func(*Card) bool) []int {
	return collectValidIndices(p.GetCardsSize(), func(i int) bool {
		return ok(p.GetCard(i))
	})
}

// roundResettable is a player that can be returned to its start-of-round state.
type roundResettable interface {
	ResetTricks()
	Reset()
	SetIsFinished(bool)
}

// resetPlayerRound clears a player's tricks, hand and finished flag. 32 player
// types had these three calls written out.
//
// Type parameter rather than interface parameter, which is measured rather than
// assumed: an interface parameter compiles to one function but makes TinyGo emit
// an itab and type descriptor per implementing type, and for a body this small
// that metadata costs more than the duplicated code. Measured at +2,038 bytes as
// interfaces versus this form. See issue #5185.
func resetPlayerRound[P roundResettable](p P) {
	p.ResetTricks()
	p.Reset()
	p.SetIsFinished(false)
}

// handSorter is a game that can sort one seat's hand by index.
type handSorter interface {
	sortHand(int)
}

// sortHands sorts every seat's hand, for the 22 games that looped over their
// players to do it. Takes the seat count separately because the rosters are
// each a different []*XPlayer, which no single interface can name.
//
// Type parameter for the same measured reason as resetPlayerRound above.
func sortHands[G handSorter](seats int, g G) {
	for i := range seats {
		g.sortHand(i)
	}
}

// undoer is a game that can step one move backwards.
type undoer interface {
	Undo() error
}

// undoN undoes n moves, stopping at the first failure and reporting which step
// failed with the cause wrapped. 31 solitaires had this loop written out.
//
// Interface parameter rather than a type parameter, which is the opposite
// choice from resetPlayerRound above and for a measured reason: this body
// contains a fmt.Errorf, and fmt is the one thing that is genuinely expensive
// to duplicate per type (#5202 measured +4,801 bytes for that shape). Those 31
// copies already exist, so collapsing them to one shared body is a saving,
// whereas a type parameter would keep them. See issue #5185.
func undoN(g undoer, n int) error {
	for i := range n {
		if err := g.Undo(); err != nil {
			return fmt.Errorf("undo step %d failed: %w", i+1, err)
		}
	}
	return nil
}
