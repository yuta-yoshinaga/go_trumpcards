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
