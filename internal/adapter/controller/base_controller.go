package controller

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
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

func derefBool(p *bool) bool {
	if p == nil {
		return false
	}
	return *p
}

func derefBoolDefault(p *bool, defaultVal bool) bool {
	if p == nil {
		return defaultVal
	}
	return *p
}

func derefInt(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

func derefIntDefault(p *int, defaultVal int) int {
	if p == nil {
		return defaultVal
	}
	return *p
}

// writePresenterResponse プレゼンターの出力を再エンコードせず直接書き込む
func (bc *baseController) writePresenterResponse(w rest.ResponseWriter, responseStr string) {
	w.WriteHeader(http.StatusOK)
	if err := w.WriteJson(json.RawMessage(responseStr)); err != nil {
		slog.Error("WriteJson error", "error", err)
	}
}

// writeJsonResponse writes a JSON response with the given HTTP status code.
func (bc *baseController) writeJsonResponse(w rest.ResponseWriter, status int, body any) {
	w.WriteHeader(status)
	if err := w.WriteJson(body); err != nil {
		slog.Error("WriteJson error", "error", err)
	}
}

// execWithSession handles common web controller boilerplate.
// handler returns true if command was recognized, false for unsupported.
func execWithSession[P WebInput, T any](
	bc *baseController,
	w rest.ResponseWriter,
	r *rest.Request,
	store *SessionStore[T],
	factory func() T,
	newDefault func(string) any,
	handler func(w rest.ResponseWriter, interactor T, param P) bool,
) {
	var param P
	if err := r.DecodeJsonPayload(&param); err != nil || param.GetCommand() == "" || param.GetSessionID() == "" {
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
