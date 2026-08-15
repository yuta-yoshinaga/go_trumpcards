package controller

import "net/http"

// dispatchLog は "log"/"l" コマンドを処理する。処理した場合 true を返す。
func dispatchLog(cmd string, bc *baseController, w http.ResponseWriter, actionLogFn func() string) bool {
	switch cmd {
	case "log", "l":
		bc.writePresenterResponse(w, actionLogFn())
		return true
	}
	return false
}

// dispatchHintAndLog は "h"/"hint" と "log"/"l" コマンドを処理する。処理した場合 true を返す。
func dispatchHintAndLog(cmd string, bc *baseController, w http.ResponseWriter, hintFn, actionLogFn func() string) bool {
	switch cmd {
	case "h", "hint":
		bc.writePresenterResponse(w, hintFn())
		return true
	case "log", "l":
		bc.writePresenterResponse(w, actionLogFn())
		return true
	}
	return false
}

// dispatchResetAndLog handles "r"/"reset" and "log"/"l" for games whose reset
// takes no arguments — resetFn is typically interactor.Reset. Games that need
// configurable reset (ResetWithConfig, etc.) keep their own case and should
// fall through to dispatchLog or dispatchResetStepLog instead.
func dispatchResetAndLog(cmd string, bc *baseController, w http.ResponseWriter, resetFn, actionLogFn func() string) bool {
	switch cmd {
	case "r", "reset":
		bc.writePresenterResponse(w, resetFn())
		return true
	case "log", "l":
		bc.writePresenterResponse(w, actionLogFn())
		return true
	}
	return false
}

// dispatchResetHintAndLog handles "r"/"reset", "h"/"hint", and "log"/"l" for
// solitaire-style games that expose hints and a no-arg Reset(). Games without
// a Hint() method should use dispatchResetAndLog instead.
func dispatchResetHintAndLog(cmd string, bc *baseController, w http.ResponseWriter, resetFn, hintFn, actionLogFn func() string) bool {
	switch cmd {
	case "r", "reset":
		bc.writePresenterResponse(w, resetFn())
		return true
	case "h", "hint":
		bc.writePresenterResponse(w, hintFn())
		return true
	case "log", "l":
		bc.writePresenterResponse(w, actionLogFn())
		return true
	}
	return false
}

// resetStepLogger is the minimal interactor surface needed for the common
// reset/step/log command trio. Declared here — at the consumer — rather than
// in usecase, per the "accept interfaces, return structs" convention.
type resetStepLogger interface {
	Reset() string
	Step() string
	ActionLog() string
}

// dispatchResetStepLog handles the "r"/"reset", "s"/"step", and "log"/"l"
// commands that recur across most simple game dispatchers. Returns true if
// the command was handled. Games that need a custom reset (e.g. ResetWithConfig)
// should intercept "r"/"reset" themselves before falling through to this helper.
func dispatchResetStepLog(cmd string, bc *baseController, w http.ResponseWriter, i resetStepLogger) bool {
	switch cmd {
	case "r", "reset":
		bc.writePresenterResponse(w, i.Reset())
		return true
	case "s", "step":
		bc.writePresenterResponse(w, i.Step())
		return true
	case "log", "l":
		bc.writePresenterResponse(w, i.ActionLog())
		return true
	}
	return false
}

// trickPlayFns is the set of interactor actions the trick-taking command set
// needs, passed as function values rather than as an interface.
//
// An interface cannot express this: every game's ResetWithConfig takes its own
// config type (domain.AluetteConfig, domain.ManilleConfig, …), so there is no
// single method set they share. Binding the calls at the call site — where the
// concrete type is still known — is what lets one dispatcher serve all of them.
// Same reasoning as dispatchResetHintAndLog's resetFn/hintFn/actionLogFn.
type trickPlayFns struct {
	resetWithConfig func() string
	play            func(cardIndex int) string
	nextTrick       func() string
	nextRound       func() string
	hint            func() string
	actionLog       func() string
}

// dispatchTrickPlay handles the reset/play/next/nextround command set shared by
// the trick-taking games, falling through to dispatchHintAndLog for "h"/"log".
// Returns false for a command it does not own so the caller can handle it.
//
// Consolidates 8 byte-identical dispatchers that differed only in name and in
// the concrete interactor type: aluetteDispatch, ganjifaDispatch,
// klaverjasDispatch, manilleDispatch, mariasDispatch, sedmaDispatch,
// spoilFiveDispatch, suecaDispatch. Being spelled differently per game is why a
// name-based search never grouped them — see issue #5368.
//
// cardIndex is a pointer because it is optional on the wire; "p" with none set
// must answer 400 rather than dereference nil.
func dispatchTrickPlay[O any](cmd string, bc *baseController, w http.ResponseWriter, fns trickPlayFns, cardIndex *int, newDefault func(string) O) bool {
	switch cmd {
	case "r", "reset":
		bc.writePresenterResponse(w, fns.resetWithConfig())
	case "p", "play":
		if !requireParam(bc, w, newDefault, cardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, fns.play(*cardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, fns.nextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, fns.nextRound())
	default:
		return dispatchHintAndLog(cmd, bc, w, fns.hint, fns.actionLog)
	}
	return true
}

// bidTrickPlayFns is trickPlayFns plus the bid step used by the trick-taking
// games that open with an auction.
type bidTrickPlayFns struct {
	resetWithConfig func() string
	bid             func(bid int) string
	play            func(cardIndex int) string
	nextTrick       func() string
	nextRound       func() string
	hint            func() string
	actionLog       func() string
}

// dispatchBidTrickPlay handles reset/bid/play/next/nextround, falling through
// to dispatchHintAndLog. Consolidates 6 byte-identical dispatchers:
// fortyFivesDispatch, napDispatch, preferenceDispatch, soloWhistDispatch,
// twentyNineDispatch, viraDispatch. See issue #5368.
//
// A separate helper rather than an optional bid field on trickPlayFns: a
// nil-able function would let a game accept "b" and silently do nothing, and
// the compiler would not notice. Two structs make the command set explicit.
func dispatchBidTrickPlay[O any](cmd string, bc *baseController, w http.ResponseWriter, fns bidTrickPlayFns, bid, cardIndex *int, newDefault func(string) O) bool {
	switch cmd {
	case "r", "reset":
		bc.writePresenterResponse(w, fns.resetWithConfig())
	case "b", "bid":
		if !requireParam(bc, w, newDefault, bid == nil, "param error: bid is required.") {
			return true
		}
		bc.writePresenterResponse(w, fns.bid(*bid))
	case "p", "play":
		if !requireParam(bc, w, newDefault, cardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, fns.play(*cardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, fns.nextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, fns.nextRound())
	default:
		return dispatchHintAndLog(cmd, bc, w, fns.hint, fns.actionLog)
	}
	return true
}

// rummyMeldFns is the interactor surface of the canasta-family command set.
//
// drawFromDiscard and meld take their argument from the request rather than as
// a parameter here: the games differ in what they pass (natural-pair indices,
// meld groups), and closing over param at the call site keeps that difference
// out of this helper.
type rummyMeldFns struct {
	resetWithConfig func() string
	drawFromStock   func() string
	drawFromDiscard func() string
	meld            func() string
	skipMeld        func() string
	discard         func(cardIndex int) string
	goOut           func() string
	nextRound       func() string
	actionLog       func() string
}

// dispatchRummyMeld handles the draw/meld/discard/go-out command set shared by
// the canasta family, falling through to dispatchLog for "log"/"l".
//
// Consolidates 4 byte-identical dispatchers: burracoDispatch, canastaDispatch,
// handAndFootDispatch, sambaDispatch. See issue #5368.
//
// dispatchLog, not dispatchHintAndLog: these games have no hint command, so
// "h" must fall through to the caller's default rather than be answered here.
func dispatchRummyMeld[O any](cmd string, bc *baseController, w http.ResponseWriter, fns rummyMeldFns, cardIndex *int, newDefault func(string) O) bool {
	switch cmd {
	case "r", "reset":
		bc.writePresenterResponse(w, fns.resetWithConfig())
	case "ds", "drawstock":
		bc.writePresenterResponse(w, fns.drawFromStock())
	case "dd", "drawdiscard":
		bc.writePresenterResponse(w, fns.drawFromDiscard())
	case "m", "meld":
		bc.writePresenterResponse(w, fns.meld())
	case "sm", "skipmeld":
		bc.writePresenterResponse(w, fns.skipMeld())
	case "d", "discard":
		if !requireParam(bc, w, newDefault, cardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, fns.discard(*cardIndex))
	case "go", "goout":
		bc.writePresenterResponse(w, fns.goOut())
	case "nr", "nextround":
		bc.writePresenterResponse(w, fns.nextRound())
	default:
		return dispatchLog(cmd, bc, w, fns.actionLog)
	}
	return true
}

// tableauMove is one move request, unpacked from whichever per-game zone type
// carried it.
//
// Every solitaire declares its own zone struct (ScorpionWebZone,
// SpideretteWebZone, WaspWebZone, …) with identical fields, and Go has no
// structural typing, so there is nothing common to accept. The caller unpacks
// its own type into this; the helper owns the validation order.
type tableauMove struct {
	haveFrom, haveTo bool
	fromZone, toZone string
	fromCol          *int
	fromCardIndex    *int
	toCol            *int
}

// dispatchTableauOnlyMove validates and performs a tableau-to-tableau move for
// the solitaires that allow no other kind, answering 400 for anything else.
//
// Consolidates 3 byte-identical dispatchers: scorpionMoveDispatch,
// spideretteMoveDispatch, waspMoveDispatch. See issue #5368.
//
// Always returns true: unlike the command dispatchers, this one owns the whole
// request. Reporting false would leave the caller to answer a request that has
// already been written to.
func dispatchTableauOnlyMove[O any](bc *baseController, w http.ResponseWriter, mv tableauMove, move func(fromCol, cardIndex, toCol int) string, newDefault func(string) O) bool {
	if !requireParam(bc, w, newDefault, !mv.haveFrom || !mv.haveTo, "param error: from and to are required.") {
		return true
	}
	if !requireParam(bc, w, newDefault, mv.fromZone != "tableau" || mv.toZone != "tableau",
		"param error: invalid move zones. Only tableau to tableau is supported.") {
		return true
	}
	if !requireParam(bc, w, newDefault, mv.fromCol == nil || mv.fromCardIndex == nil || mv.toCol == nil,
		"param error: from.col, from.cardIndex, to.col are required.") {
		return true
	}
	bc.writePresenterResponse(w, move(*mv.fromCol, *mv.fromCardIndex, *mv.toCol))
	return true
}

// tarotDiscardPlayFns is the interactor surface of the tarot command set.
type tarotDiscardPlayFns struct {
	resetWithConfig func() string
	discard         func(cardIndices []int) string
	play            func(cardIndex int) string
	nextTrick       func() string
	nextRound       func() string
	hint            func() string
	actionLog       func() string
}

// dispatchTarotDiscardPlay handles reset/discard/play/next/nextround for the
// tarot games, falling through to dispatchHintAndLog.
//
// Consolidates 3 byte-identical dispatchers: minchiateDispatch, scartoDispatch,
// tarocchiniDispatch. See issue #5368.
//
// The discard step answers to two spellings ("s"/"scarto" and "d"/"discard")
// and takes a slice, so its missing-check is nil rather than len == 0: an empty
// slice is a legal "discard nothing", and rejecting it would 400 a valid
// request.
func dispatchTarotDiscardPlay[O any](cmd string, bc *baseController, w http.ResponseWriter, fns tarotDiscardPlayFns, cardIndices []int, cardIndex *int, newDefault func(string) O) bool {
	switch cmd {
	case "r", "reset":
		bc.writePresenterResponse(w, fns.resetWithConfig())
	case "s", "scarto", "d", "discard":
		if !requireParam(bc, w, newDefault, cardIndices == nil, "param error: cardIndices is required.") {
			return true
		}
		bc.writePresenterResponse(w, fns.discard(cardIndices))
	case "p", "play":
		if !requireParam(bc, w, newDefault, cardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, fns.play(*cardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, fns.nextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, fns.nextRound())
	default:
		return dispatchHintAndLog(cmd, bc, w, fns.hint, fns.actionLog)
	}
	return true
}

// topCardMove is one move request for the top-card-only solitaires, unpacked
// from whichever per-game zone type carried it. Same reasoning as tableauMove.
type topCardMove struct {
	haveFrom, haveTo bool
	fromZone, toZone string
	fromCol          *int
	toCol            *int
}

// topCardMoveFns is the interactor surface these moves need.
type topCardMoveFns struct {
	tableauToTableau    func(fromCol, cardIndex, toCol int) string
	tableauToFoundation func(col int) string
}

// dispatchTopCardMove validates and performs a move for the solitaires that
// only ever move the top card of a column.
//
// Consolidates 3 byte-identical dispatchers: bakersDozenMoveDispatch,
// beleagueredCastleMoveDispatch, streetsAndAlleysMoveDispatch. See #5368.
//
// Passes -1 as the card index rather than the client's value: these games move
// only the top card, so the domain resolves the index from its own state and
// the server has one less untrusted input. That contract is pinned by a test —
// forwarding the client value instead would let a request name a card that is
// not on top, and the move would still look legal.
func dispatchTopCardMove[O any](bc *baseController, w http.ResponseWriter, mv topCardMove, fns topCardMoveFns, newDefault func(string) O) bool {
	if !requireParam(bc, w, newDefault, !mv.haveFrom || !mv.haveTo, "param error: from and to are required.") {
		return true
	}
	switch {
	case mv.fromZone == "tableau" && mv.toZone == "tableau":
		if !requireParam(bc, w, newDefault, mv.fromCol == nil || mv.toCol == nil, "param error: from.col and to.col are required.") {
			return true
		}
		bc.writePresenterResponse(w, fns.tableauToTableau(*mv.fromCol, -1, *mv.toCol))
	case mv.fromZone == "tableau" && mv.toZone == "foundation":
		if !requireParam(bc, w, newDefault, mv.fromCol == nil, "param error: from.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, fns.tableauToFoundation(*mv.fromCol))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
