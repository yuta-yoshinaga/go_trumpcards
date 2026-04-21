package games

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
)

// WebBinding captures the ingredients needed to wire a game into the HTTP server in a single declaration.
type WebBinding[I any, P controller.WebInput, O any] struct {
	// Name is the canonical short name — must match an entry in registry.
	Name string

	// NewInteractor produces a fresh interactor bound to a fresh presenter.
	NewInteractor func() I

	// NewController is the NewXxxWebController (in-memory session) variant.
	NewController func(func() I) *controller.GameWebController[I, P, O]
}

// Bind registers this binding as the server-side factory for the named
// game. Panics via BindWebController if Name is not a known game.
func (b WebBinding[I, P, O]) Bind() {
	BindWebController(b.Name, func() WebController {
		return b.NewController(b.NewInteractor)
	})
}
