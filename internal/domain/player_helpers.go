package domain

import (
	"errors"
	"fmt"
	"math/rand"
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
	return elemAt(players, idx)
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

// undoNChecked undoes n moves after validating the request against the history
// length, so a bad n is rejected before the game is half-rewound. 10 solitaires
// had this written out.
//
// The history is passed as a length rather than a slice because each game's
// snapshot type is its own struct. Interface parameter for the same reason as
// undoN above: it keeps one copy of the shared body. See issue #5185.
func undoNChecked(g undoer, n, historyLen int) error {
	if n <= 0 {
		return errors.New("n must be positive")
	}
	if n > historyLen {
		return errors.New("not enough history")
	}
	for range n {
		if err := g.Undo(); err != nil {
			return err
		}
	}
	return nil
}

// canPlaceOnFoundationPile reports whether card may go on a foundation pile:
// an ace onto an empty pile, otherwise the next rank up in the same suit.
// 9 solitaires had this written out. Needs no type parameter -- every
// foundation is a [][]*Card. See issue #5185.
func canPlaceOnFoundationPile(pile []*Card, card *Card) bool {
	if len(pile) == 0 {
		return card.GetValue() == 1
	}
	top := pile[len(pile)-1]
	return card.GetDesign() == top.GetDesign() && card.GetValue() == top.GetValue()+1
}

// bySuitThenValue orders cards by suit, then by rank within a suit. 13 games
// passed this exact comparator to sortPlayerHand.
func bySuitThenValue(ci, cj *Card) bool {
	if ci.GetDesign() != cj.GetDesign() {
		return ci.GetDesign() < cj.GetDesign()
	}
	return ci.GetValue() < cj.GetValue()
}

// humanHandHolder is a seat whose hand can be sorted and which knows whether a
// human is sitting in it.
type humanHandHolder interface {
	handHolder
	GetIsHuman() bool
}

// sortHumanHands sorts the human seat's hand by suit then rank, leaving CPU
// hands untouched. 5 games had this written out.
//
// Those bodies used sort.SliceStable where sortPlayerHand uses sort.Slice.
// That is not a behaviour change here: stability is only observable when two
// cards compare equal, i.e. share a (design, value), and no deck in this
// repository produces that -- including the hanafuda games among these five,
// whose cards are built as NewCard(month, index) and measured as 0 duplicate
// pairs. See issue #5185.
func sortHumanHands[P humanHandHolder](players []P) {
	for _, p := range players {
		if p.GetIsHuman() {
			sortPlayerHand(p, bySuitThenValue)
		}
	}
}

// longestSuit returns the suit p holds most of, breaking ties in
// spade/clover/heart/diamond order and returning spade for an empty hand.
// 7 games had this written out.
func longestSuit[P handReader](p P) int {
	counts := map[int]int{}
	for i := 0; i < p.GetCardsSize(); i++ {
		counts[p.GetCard(i).GetDesign()]++
	}
	bestSuit, bestCnt := CardDesignSpade, -1
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		if counts[suit] > bestCnt {
			bestCnt = counts[suit]
			bestSuit = suit
		}
	}
	return bestSuit
}

// handWriter is a seat whose hand can be emptied and refilled.
type handWriter interface {
	GetCardsSize() int
	RemoveCard(int) *Card
	AddCard(*Card)
}

// setHandForTest replaces a seat's hand outright. 9 games had this written out
// behind their exported SetHandForTest.
//
// The constraint is spelled `*T` rather than plain handWriter so that a nil
// player can be detected: GetPlayer returns a typed nil for an out-of-range
// index, and a typed nil stored in an interface is not itself nil, so an
// interface parameter would turn the guard into a panic. See issue #5185.
func setHandForTest[T any, P interface {
	*T
	handWriter
}](p P, cards []*Card) {
	if p == nil {
		return
	}
	for p.GetCardsSize() > 0 {
		p.RemoveCard(0)
	}
	for _, c := range cards {
		p.AddCard(c)
	}
}

// moveFinisher is a solitaire that records a move and re-evaluates the board.
type moveFinisher interface {
	appendLog(actionType, detail string, cards []*Card)
	checkGameClear()
	checkStalemate()
}

// afterMove completes one solitaire move: count it, log it, then re-check for a
// win and for a stalemate. 8 games had this written out.
//
// moveCount is passed as a pointer because these games use it as the log
// entry's TurnNumber, so the increment has to happen before appendLog runs --
// an order the tests pin rather than leave to reading.
func afterMove(moveCount *int, g moveFinisher, actionType, detail string, card *Card) {
	*moveCount++
	var cards []*Card
	if card != nil {
		cards = []*Card{card}
	}
	g.appendLog(actionType, detail, cards)
	g.checkGameClear()
	g.checkStalemate()
}

// bettor is a seat that can be out of the betting for this hand. Deliberately
// narrower than the existing BettingPlayer, which names twelve methods: these
// helpers only ask whether a seat is still in, and the narrow form is the same
// choice made for handReader above.
type bettor interface {
	GetFolded() bool
	GetAllIn() bool
}

// findNextActive returns the next seat after fromIdx that is still betting,
// falling back to the immediate next seat when nobody qualifies -- callers
// index with the result, so this never returns -1. 7 games had this written out.
func findNextActive[P bettor](players []P, fromIdx int) int {
	for i := 1; i <= len(players); i++ {
		next := (fromIdx + i) % len(players)
		if !players[next].GetFolded() && !players[next].GetAllIn() {
			return next
		}
	}
	return (fromIdx + 1) % len(players)
}

// drawFromDeck takes the next card from deck, advancing the cursor and marking
// the card drawn, or returns nil once the deck is exhausted. 8 games had this
// written out. Needs no type parameter -- every deck is a []*Card.
func drawFromDeck(deck []*Card, drawn *int) *Card {
	if *drawn >= len(deck) {
		return nil
	}
	card := deck[*drawn]
	card.SetDraw(true)
	*drawn++
	return card
}

// validateFollowSuit enforces following the led suit when the seat can. 8 games
// had this written out.
func validateFollowSuit[P handReader](trick []*TrickCard, players []P, playerIdx int, card *Card) error {
	if len(trick) == 0 {
		return nil
	}
	leadSuit := trick[0].Card.GetDesign()
	if card.GetDesign() != leadSuit && handHasSuit(players[playerIdx], leadSuit) {
		return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
	}
	return nil
}

// resetPlayer clears a player's hand and finished flag, for games whose reset
// does not also clear tricks. 9 player types had these two calls written out.
func resetPlayer[P resettable](p P) {
	p.Reset()
	p.SetIsFinished(false)
}

// sortHandInPlace reorders a seat's hand with a whole-slice sorter, for games
// whose ordering is expressed as a sort over []*Card rather than a comparator.
// 6 games had the extract/sort/Reset/re-add round trip written out.
func sortHandInPlace[P handHolder](p P, sortFn func([]*Card)) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sortFn(cards)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// handHasAny reports whether any card in p satisfies ok. 5 games had this loop
// written out.
func handHasAny[P handReader](p P, ok func(*Card) bool) bool {
	for i := 0; i < p.GetCardsSize(); i++ {
		if ok(p.GetCard(i)) {
			return true
		}
	}
	return false
}

// chipsOfFirst returns the first seat's chip count, or 0 when there are no
// seats. 5 games had this written out.
func chipsOfFirst[P interface{ GetChips() int }](players []P) int {
	if len(players) == 0 {
		return 0
	}
	return players[0].GetChips()
}

// roundScorable is a player whose per-round score can be cleared along with the
// rest of its round state.
type roundScorable interface {
	resettable
	SetRoundScore(int)
}

// resetRoundScored clears a player's round score, hand and finished flag.
func resetRoundScored[P roundScorable](p P) {
	p.SetRoundScore(0)
	p.Reset()
	p.SetIsFinished(false)
}

// trickRoundScorable additionally tracks tricks taken.
type trickRoundScorable interface {
	roundScorable
	ResetTricks()
}

// resetRoundWithTricks clears a player's round score, tricks, hand and finished
// flag, in that order -- the order the bodies it replaces used.
func resetRoundWithTricks[P trickRoundScorable](p P) {
	p.SetRoundScore(0)
	p.ResetTricks()
	p.Reset()
	p.SetIsFinished(false)
}

// stockRecycler is a game that can record recycling the discard pile.
type stockRecycler interface {
	appendLog(playerIdx int, actionType, detail string, cards []*Card)
}

// recycleDiscardIntoStock moves everything but the top discard back under the
// draw pile, shuffled, and logs it. Returns false when there is nothing to
// recycle. 5 games had this written out.
//
// The piles are passed by pointer because the helper rewrites both. Keeping the
// fmt.Sprintf here rather than at each call site means one copy of it instead
// of five.
func recycleDiscardIntoStock(discard, draw *[]*Card, g stockRecycler) bool {
	if len(*discard) <= 1 {
		return false
	}
	top := (*discard)[len(*discard)-1]
	rest := (*discard)[:len(*discard)-1]
	*discard = []*Card{top}
	rand.Shuffle(len(rest), func(i, j int) { rest[i], rest[j] = rest[j], rest[i] })
	*draw = append(*draw, rest...)
	g.appendLog(-1, "recycle", fmt.Sprintf("Discard pile recycled into stock (%d cards)", len(rest)), nil)
	return true
}

// seatHoldingCard returns the index of the first seat holding a card that
// satisfies want, or -1 when nobody does. 3 games looked for the two of clubs
// this way.
func seatHoldingCard[P handReader](players []P, want func(*Card) bool) int {
	for i, p := range players {
		if handHasAny(p, want) {
			return i
		}
	}
	return -1
}

// bettingRoundComplete reports whether every seat still in the hand has acted.
// 3 games had this written out. Reuses the bettor constraint rather than
// declaring a second interface with the same two methods.
func bettingRoundComplete[P bettor](players []P, acted []bool) bool {
	for i, p := range players {
		if p.GetFolded() || p.GetAllIn() {
			continue
		}
		if !acted[i] {
			return false
		}
	}
	return true
}

// totalScorer is a seat with a running game score.
type totalScorer interface {
	GetTotalScore() int
}

// topScorers returns the indices of every seat tied for the highest total
// score. 3 games had this two-pass scan written out.
func topScorers[P totalScorer](players []P) []int {
	if len(players) == 0 {
		return []int{}
	}
	best := players[0].GetTotalScore()
	for _, p := range players[1:] {
		if p.GetTotalScore() > best {
			best = p.GetTotalScore()
		}
	}
	winners := make([]int, 0)
	for i, p := range players {
		if p.GetTotalScore() == best {
			winners = append(winners, i)
		}
	}
	return winners
}

// trickScorer is a game that can rank the cards in the current trick.
type trickScorer interface {
	leadSuit() int
	trickScore(*Card, int) int
}

// trickWinnerByScore returns the seat that played the highest-scoring card in
// the trick, or 0 when no card has been played. Ties keep the earliest play,
// matching the bodies replaced, which compared with a strict >.
func trickWinnerByScore(trick []*TrickCard, g trickScorer) int {
	if len(trick) == 0 {
		return 0
	}
	ls := g.leadSuit()
	winnerIdx := trick[0].PlayerIdx
	winnerScore := g.trickScore(trick[0].Card, ls)
	for _, tc := range trick[1:] {
		if s := g.trickScore(tc.Card, ls); s > winnerScore {
			winnerScore = s
			winnerIdx = tc.PlayerIdx
		}
	}
	return winnerIdx
}

// drawFromPile deals up to n cards off the end of pile into to, recycling once
// the pile runs dry and stopping when recycling cannot refill it. Returns how
// many were actually dealt. 3 games had this written out.
func drawFromPile[P interface{ AddCard(*Card) }](pile *[]*Card, to P, n int, recycle func()) int {
	drawn := 0
	for i := 0; i < n; i++ {
		if len(*pile) == 0 {
			recycle()
		}
		if len(*pile) == 0 {
			break
		}
		card := (*pile)[len(*pile)-1]
		*pile = (*pile)[:len(*pile)-1]
		to.AddCard(card)
		drawn++
	}
	return drawn
}
