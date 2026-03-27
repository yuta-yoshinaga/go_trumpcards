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

// execWithSession handles common web controller boilerplate.
// handler returns true if command was recognized, false for unsupported.
func execWithSession[P WebInput, T any](
	bc *baseController,
	w http.ResponseWriter,
	r *http.Request,
	store *SessionStore[T],
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
	interactor, mu, ok := store.GetWithLock(param.GetSessionID(), factory)
	if !ok {
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error."))
		return
	}
	mu.Lock()
	defer mu.Unlock()
	if !handler(w, interactor, param) {
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("Unsupported command."))
	}
}
