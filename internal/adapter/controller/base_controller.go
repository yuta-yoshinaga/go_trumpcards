package controller

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// WebInput is the interface that all web controller input structs must implement.
type WebInput interface {
	GetCommand() string
	GetSessionID() string
}

// BaseWebInput holds fields common to all game WebInput structs.
type BaseWebInput struct {
	Command   string `json:"command"`
	SessionID string `json:"sessionId"`
	N         *int   `json:"n,omitempty"`
}

// GetCommand returns the command string.
func (b BaseWebInput) GetCommand() string { return b.Command }

// GetSessionID returns the session ID string.
func (b BaseWebInput) GetSessionID() string { return b.SessionID }

// baseController 各WebController共通のレスポンス書き込みロジック
type baseController struct{}

func deref[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}

func derefDefault[T any](p *T, defaultVal T) T {
	if p == nil {
		return defaultVal
	}
	return *p
}

// writePresenterResponse プレゼンターの出力を再エンコードせず直接書き込む
func (bc *baseController) writePresenterResponse(w http.ResponseWriter, responseStr string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(responseStr)); err != nil {
		slog.Error("Write error", "error", err)
	}
}

// writeJsonResponse writes a JSON response with the given HTTP status code.
func (bc *baseController) writeJsonResponse(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("WriteJson error", "error", err)
	}
}

// requireParam guards a dispatch case against a missing required parameter.
// When missing is true it writes a 400 Bad Request carrying msg (built via the
// game-specific newDefault so the response keeps that game's output shape) and
// returns false; otherwise it returns true and the caller proceeds.
//
// This collapses the "if param.X == nil { writeJsonResponse(400, newDefault(
// msg)); return true }" block that was duplicated across ~240 dispatch cases in
// 64 *WebController.go files (issue #2102). The condition is passed verbatim as
// `missing` (no negation at the call site), so multi-field checks like
// `param.From == nil || param.To == nil` carry over unchanged. The generic O
// absorbs each game's distinct *WebOutput type. Canonical usage:
//
//	if !requireParam(bc, w, newDefault, param.Col == nil, msgColRequired) {
//		return true
//	}
//
// Returning true from the dispatch case (command was recognized but invalid)
// matches the previous behavior: the 400 has already been written here.
func requireParam[O any](bc *baseController, w http.ResponseWriter, newDefault func(string) O, missing bool, msg string) bool {
	if missing {
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault(msg))
		return false
	}
	return true
}

// execWithSession handles common web controller boilerplate.
// handler returns true if command was recognized, false for unsupported.
func execWithSession[P WebInput, T any](
	bc *baseController,
	w http.ResponseWriter,
	r *http.Request,
	provider SessionProvider[T],
	factory func() T,
	newDefault func(string) any,
	handler func(w http.ResponseWriter, interactor T, param P) bool,
) {
	var param P
	if err := json.NewDecoder(r.Body).Decode(&param); err != nil || param.GetCommand() == "" || param.GetSessionID() == "" {
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error."))
		return
	}
	if cmd := param.GetCommand(); cmd == "q" || cmd == "quit" {
		bc.writeJsonResponse(w, http.StatusOK, newDefault("bye."))
		return
	}
	interactor, release, ok := provider.Acquire(param.GetSessionID(), factory)
	if !ok {
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error."))
		return
	}
	defer release()
	if !handler(w, interactor, param) {
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("Unsupported command."))
	}
}
