package controller

import (
	"github.com/ant0ine/go-json-rest/rest"
)

// GameWebController is a generic web controller that eliminates boilerplate
// shared across all game-specific web controllers. Each game provides its own
// factory, default-output builder, and command dispatcher.
type GameWebController[I any, P WebInput, O any] struct {
	baseController
	factory    func() I
	store      *SessionStore[I]
	newDefault func(string) O
	dispatch   func(bc *baseController, w rest.ResponseWriter, interactor I, param P, newDefault func(string) O) bool
}

// NewGameWebController creates a GameWebController.
func NewGameWebController[I any, P WebInput, O any](
	factory func() I,
	newDefault func(string) O,
	dispatch func(bc *baseController, w rest.ResponseWriter, interactor I, param P, newDefault func(string) O) bool,
) *GameWebController[I, P, O] {
	return &GameWebController[I, P, O]{
		factory:    factory,
		store:      NewSessionStore[I](),
		newDefault: newDefault,
		dispatch:   dispatch,
	}
}

// Exec handles an incoming game request.
func (gwc *GameWebController[I, P, O]) Exec(w rest.ResponseWriter, r *rest.Request) {
	execWithSession(&gwc.baseController, w, r, gwc.store, gwc.factory,
		func(msg string) any { return gwc.newDefault(msg) },
		func(w rest.ResponseWriter, interactor I, param P) bool {
			return gwc.dispatch(&gwc.baseController, w, interactor, param, gwc.newDefault)
		})
}

// Stop stops the background cleanup goroutine of the session store.
func (gwc *GameWebController[I, P, O]) Stop() {
	gwc.store.Stop()
}
