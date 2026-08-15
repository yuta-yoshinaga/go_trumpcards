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
