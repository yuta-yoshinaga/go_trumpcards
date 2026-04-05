package controller

import (
	"net/http"
)

// GameWebController is a generic web controller that eliminates boilerplate
// shared across all game-specific web controllers. Each game provides its own
// factory, default-output builder, and command dispatcher.
type GameWebController[I any, P WebInput, O any] struct {
	baseController
	factory    func() I
	provider   SessionProvider[I]
	newDefault func(string) O
	dispatch   func(bc *baseController, w http.ResponseWriter, interactor I, param P, newDefault func(string) O) bool
}

// NewGameWebController creates a GameWebController with the default in-memory
// session provider.
func NewGameWebController[I any, P WebInput, O any](
	factory func() I,
	newDefault func(string) O,
	dispatch func(bc *baseController, w http.ResponseWriter, interactor I, param P, newDefault func(string) O) bool,
) *GameWebController[I, P, O] {
	return NewGameWebControllerWithProvider[I, P, O](NewMemorySessionProvider[I](), factory, newDefault, dispatch)
}

// NewGameWebControllerWithProvider creates a GameWebController with an
// explicit SessionProvider (e.g. KV-backed for Workers).
func NewGameWebControllerWithProvider[I any, P WebInput, O any](
	provider SessionProvider[I],
	factory func() I,
	newDefault func(string) O,
	dispatch func(bc *baseController, w http.ResponseWriter, interactor I, param P, newDefault func(string) O) bool,
) *GameWebController[I, P, O] {
	return &GameWebController[I, P, O]{
		factory:    factory,
		provider:   provider,
		newDefault: newDefault,
		dispatch:   dispatch,
	}
}

// Exec handles an incoming game request.
func (gwc *GameWebController[I, P, O]) Exec(w http.ResponseWriter, r *http.Request) {
	execWithSession(&gwc.baseController, w, r, gwc.provider, gwc.factory,
		func(msg string) any { return gwc.newDefault(msg) },
		func(w http.ResponseWriter, interactor I, param P) bool {
			return gwc.dispatch(&gwc.baseController, w, interactor, param, gwc.newDefault)
		})
}

// Stop stops the background cleanup goroutine of the session provider.
func (gwc *GameWebController[I, P, O]) Stop() {
	gwc.provider.Stop()
}

// WebControllerPair returns a (New, NewWithProvider) constructor pair for a
// game web controller, eliminating two boilerplate functions per game file.
func WebControllerPair[I any, P WebInput, O any](
	newDefault func(string) O,
	dispatch func(*baseController, http.ResponseWriter, I, P, func(string) O) bool,
) (
	func(func() I) *GameWebController[I, P, O],
	func(SessionProvider[I], func() I) *GameWebController[I, P, O],
) {
	return func(factory func() I) *GameWebController[I, P, O] {
			return NewGameWebController(factory, newDefault, dispatch)
		},
		func(provider SessionProvider[I], factory func() I) *GameWebController[I, P, O] {
			return NewGameWebControllerWithProvider(provider, factory, newDefault, dispatch)
		}
}
