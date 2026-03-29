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
