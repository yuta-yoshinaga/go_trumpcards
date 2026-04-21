package games

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
)

// WebBinding captures the ingredients needed to wire a game into the HTTP
// server in a single declaration. One WebBinding value replaces the four
// arguments previously passed to BindWebControllerFor, so a game can
// express its server-side registration once — near the game's own code if
// desired — instead of as a multi-line closure inside games_server.go.
// See issue #1458 for the motivation.
//
// Type parameters:
//   - I — the interactor interface (e.g. usecase.BlackJackInteractorIF)
//   - P — the web input struct (e.g. BlackJackWebInput)
//   - O — the web output struct pointer (e.g. *BlackJackWebOutput)
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
